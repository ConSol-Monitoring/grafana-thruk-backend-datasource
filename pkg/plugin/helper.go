package plugin

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/httpclient"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

const maxResponseBodyBytes int64 = 100 * 1024 * 1024

var (
	ErrResponseBodyTooLarge = errors.New("response body exceeds configured limit")
	ErrInvalidAPIPath       = errors.New("invalid API path")
	ErrAPIPathEscapes       = errors.New("API path must not escape the Thruk API")
)

func readResponseBody(body io.Reader) ([]byte, error) {
	limited := io.LimitReader(body, maxResponseBodyBytes+1)

	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrResponseBodyRead, err)
	}

	if int64(len(data)) > maxResponseBodyBytes {
		return nil, ErrResponseBodyTooLarge
	}

	return data, nil
}

func validateAPIPath(value string) error {
	if value == "" || strings.ContainsAny(value, "?#\x00\r\n") {
		return ErrInvalidAPIPath
	}

	const maxDecodePasses = 8

	decoded := value
	for range maxDecodePasses {
		if strings.ContainsAny(decoded, "\\\x00\r\n") {
			return ErrInvalidAPIPath
		}

		for segment := range strings.SplitSeq(decoded, "/") {
			if segment == ".." {
				return ErrAPIPathEscapes
			}
		}

		next, err := url.PathUnescape(decoded)
		if err != nil {
			return ErrInvalidAPIPath
		}

		if next == decoded {
			return nil
		}

		decoded = next
	}

	return ErrInvalidAPIPath
}

// cookieNames extracts the cookie names from the raw values of a Cookie header.
// It never returns the cookie values themselves, so it is safe to log.
func cookieNames(cookieHeaderValues []string) []string {
	var names []string

	for _, value := range cookieHeaderValues {
		for part := range strings.SplitSeq(value, ";") {
			name := strings.TrimSpace(part)
			if eq := strings.IndexByte(name, '='); eq >= 0 {
				name = name[:eq]
			}

			if name != "" {
				names = append(names, name)
			}
		}
	}

	return names
}

// forwardAuthHeaders copies the explicitly allow-listed authentication headers from the incoming
// Grafana request onto the outgoing Thruk request. Headers that are not in the allow-list are never
// forwarded, so unrelated Grafana identity or tracing headers cannot reach the Thruk server.
func forwardAuthHeaders(outReq *http.Request, inHeaders http.Header) {
	allowedHeaders := []string{"Cookie", "Authorization", "X-Id-Token"}

	for _, name := range allowedHeaders {
		for _, value := range inHeaders.Values(name) {
			outReq.Header.Add(name, value)
		}
	}
}

// HTTPClientOptionsSetDefaults ensures some http Client settings are the way that this datasource requires
// Intended to be run to modify http Client that backend.DatasourceInstanceSettings.HTTPClientOpts(ctx) generates.
func HTTPClientOptionsSetDefaults(opts *httpclient.Options) {
	// Disable the SDK's blanket header forwarding. The plugin copies an explicit allow-list of
	// authentication headers (Cookie, Authorization, X-Id-Token) onto each upstream request itself,
	// so unrelated Grafana headers are never sent to Thruk.
	opts.ForwardHTTPHeaders = false

	// Modify some of the timeouts and other settings. golang-sdk still calls this struct TimeoutOpts
	// These end up in the golang http.Transport

	const timeoutSeconds = 60

	const dialTimeoutSeconds = 10

	const tlsHandshakeTimeoutSeconds = 10

	opts.Timeouts = &httpclient.TimeoutOptions{
		Timeout:               timeoutSeconds * time.Second,
		DialTimeout:           dialTimeoutSeconds * time.Second,
		KeepAlive:             httpclient.DefaultTimeoutOptions.KeepAlive,
		TLSHandshakeTimeout:   tlsHandshakeTimeoutSeconds * time.Second,
		ExpectContinueTimeout: httpclient.DefaultTimeoutOptions.ExpectContinueTimeout,
		// MaxIdleConns controls the maximum number of idle (keep-alive) connections across all hosts. Zero means no limit.
		MaxIdleConns: 0,
		// MaxConnsPerHost optionally limits the total number of connections per host, including connections in the dialing, active, and idle states.
		// On limit violation, dials will block. Zero means no limit.
		MaxConnsPerHost: 0,
		// MaxIdleConnsPerHost, if non-zero, controls the maximum idle (keep-alive) connections to keep per-host. If zero, the stdlib DefaultMaxIdleConnsPerHost (2) is used.
		MaxIdleConnsPerHost: 0,
		IdleConnTimeout:     httpclient.DefaultTimeoutOptions.IdleConnTimeout,
	}
}

// DataSourceInstanceSettingsToString returns a string representation of the DataSourceInstanceSettings.
func DataSourceInstanceSettingsToString(settings *backend.DataSourceInstanceSettings) string {
	var jsonDataBytes []byte

	jsonDataStr := "nil"

	if settings.JSONData != nil {
		jsonDataBytes = []byte(settings.JSONData)
		// Pretty-print the JSON
		var prettyJSON bytes.Buffer

		err := json.Indent(&prettyJSON, jsonDataBytes, "", "  ")
		if err == nil {
			jsonDataStr = "\n" + prettyJSON.String()
		} else {
			jsonDataStr = "\n" + string(jsonDataBytes)
		}
	}

	var decryptedDataStr string

	if settings.DecryptedSecureJSONData != nil {
		// log only the keys, never the secret values
		keys := make([]string, 0, len(settings.DecryptedSecureJSONData))
		for k := range settings.DecryptedSecureJSONData {
			keys = append(keys, k)
		}

		sort.Strings(keys)

		decryptedDataStr = "\n" + strings.Join(keys, "\n")
	} else {
		decryptedDataStr = "nil"
	}

	return fmt.Sprintf(`DataSourceInstanceSettings{
  ID: %d
  UID: %s
  Type: %s
  Name: %s
  URL: %s
  User: %s
  Database: %s
  BasicAuthEnabled: %t
  BasicAuthUser: %s
  JSONData: %s
  DecryptedSecureJSONData: %s
  Updated: %s
  APIVersion: %s
}`,
		settings.ID,
		settings.UID,
		settings.Type,
		settings.Name,
		settings.URL,
		settings.User,
		settings.Database,
		settings.BasicAuthEnabled,
		settings.BasicAuthUser,
		jsonDataStr,
		decryptedDataStr,
		settings.Updated.Format(time.RFC3339),
		settings.APIVersion,
	)
}

// FieldTypeToString returns the name of the data.FieldType
//
//nolint:gocyclo,cyclop,funlen
func FieldTypeToString(ft data.FieldType) string {
	//nolint:exhaustive // data.FieldTypeNullableJSON is deprecated, not adding it
	switch ft {
	case data.FieldTypeUnknown:
		return "FieldTypeUnknown"
	case data.FieldTypeInt8:
		return "FieldTypeInt8"
	case data.FieldTypeNullableInt8:
		return "FieldTypeNullableInt8"
	case data.FieldTypeInt16:
		return "FieldTypeInt16"
	case data.FieldTypeNullableInt16:
		return "FieldTypeNullableInt16"
	case data.FieldTypeInt32:
		return "FieldTypeInt32"
	case data.FieldTypeNullableInt32:
		return "FieldTypeNullableInt32"
	case data.FieldTypeInt64:
		return "FieldTypeInt64"
	case data.FieldTypeNullableInt64:
		return "FieldTypeNullableInt64"
	case data.FieldTypeUint8:
		return "FieldTypeUint8"
	case data.FieldTypeNullableUint8:
		return "FieldTypeNullableUint8"
	case data.FieldTypeUint16:
		return "FieldTypeUint16"
	case data.FieldTypeNullableUint16:
		return "FieldTypeNullableUint16"
	case data.FieldTypeUint32:
		return "FieldTypeUint32"
	case data.FieldTypeNullableUint32:
		return "FieldTypeNullableUint32"
	case data.FieldTypeUint64:
		return "FieldTypeUint64"
	case data.FieldTypeNullableUint64:
		return "FieldTypeNullableUint64"
	case data.FieldTypeFloat32:
		return "FieldTypeFloat32"
	case data.FieldTypeNullableFloat32:
		return "FieldTypeNullableFloat32"
	case data.FieldTypeFloat64:
		return "FieldTypeFloat64"
	case data.FieldTypeNullableFloat64:
		return "FieldTypeNullableFloat64"
	case data.FieldTypeString:
		return "FieldTypeString"
	case data.FieldTypeNullableString:
		return "FieldTypeNullableString"
	case data.FieldTypeBool:
		return "FieldTypeBool"
	case data.FieldTypeNullableBool:
		return "FieldTypeNullableBool"
	case data.FieldTypeTime:
		return "FieldTypeTime"
	case data.FieldTypeNullableTime:
		return "FieldTypeNullableTime"
	case data.FieldTypeJSON:
		return "FieldTypeJSON"
	case data.FieldTypeEnum:
		return "FieldTypeEnum"
	case data.FieldTypeNullableEnum:
		return "FieldTypeNullableEnum"
	default:
		return "FieldTypeUnknown"
	}
}

// HTTPClientOptionsToString returns a string repesentation of httpclient.Options .
func HTTPClientOptionsToString(opts httpclient.Options) string {
	var buf bytes.Buffer

	fmt.Fprintf(&buf, "HTTPClientOptions{ForwardHTTPHeaders:%t", opts.ForwardHTTPHeaders)

	if opts.Timeouts != nil {
		fmt.Fprintf(&buf, ", Timeout:%s", opts.Timeouts.Timeout)
	}

	if opts.TLS != nil {
		fmt.Fprintf(&buf, ", TLS.InsecureSkipVerify:%t", opts.TLS.InsecureSkipVerify)
	}

	if opts.BasicAuth != nil {
		fmt.Fprintf(&buf, ", BasicAuth.User:%s", opts.BasicAuth.User)
	}

	if len(opts.Header) > 0 {
		// log only the header names, never the values (they may contain secrets)
		keys := make([]string, 0, len(opts.Header))
		for key := range opts.Header {
			keys = append(keys, key)
		}

		sort.Strings(keys)

		fmt.Fprintf(&buf, ", Headers:[%s]", strings.Join(keys, ", "))
	}

	if len(opts.Labels) > 0 {
		buf.WriteString(", Labels:[")

		for key, values := range opts.Labels {
			fmt.Fprintf(&buf, "%s=%v, ", key, values)
		}

		buf.WriteString("]")
	}

	buf.WriteByte('}')

	return buf.String()
}
