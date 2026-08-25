package plugin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/httpclient"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

// cookieNames extracts the cookie names from the raw values of a Cookie header.
// It never returns the cookie values themselves, so it is safe to log.
func cookieNames(cookieHeaderValues []string) []string {
	var names []string

	for _, value := range cookieHeaderValues {
		for _, part := range strings.Split(value, ";") {
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

// Ensures some http Client settings are the way that this datasource requires
// Intended to be run to modify http Client that backend.DatasourceInstanceSettings.HTTPClientOpts(ctx) generates
func HTTPClientOptionsSetDefaults(opts *httpclient.Options) {
	// Always forward the headers, this is how 'thruk_auth' cookies should be passed
	opts.ForwardHTTPHeaders = true

	// Modify some of the timeouts and other settings. golang-sdk still calls this struct TimeoutOpts
	// These end up in the golang http.Transport

	opts.Timeouts = &httpclient.TimeoutOptions{
		Timeout:               60 * time.Second,
		DialTimeout:           10 * time.Second,
		KeepAlive:             httpclient.DefaultTimeoutOptions.KeepAlive,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: httpclient.DefaultTimeoutOptions.ExpectContinueTimeout,
		// MaxIdleConns controls the maximum number of idle (keep-alive) connections across all hosts. Zero means no limit.
		MaxIdleConns: 0,
		// MaxConnsPerHost optionally limits the total number of connections per host, including connections in the dialing, active, and idle states. On limit violation, dials will block. Zero means no limit.
		MaxConnsPerHost: 0,
		// MaxIdleConnsPerHost, if non-zero, controls the maximum idle (keep-alive) connections to keep per-host. If zero, the stdlib DefaultMaxIdleConnsPerHost (2) is used.
		MaxIdleConnsPerHost: 0,
		IdleConnTimeout:     httpclient.DefaultTimeoutOptions.IdleConnTimeout,
	}
}

// String returns a string representation of the DataSourceInstanceSettings.
func DataSourceInstanceSettingsToString(settings *backend.DataSourceInstanceSettings) string {
	var jsonDataBytes []byte
	jsonDataStr := "nil"

	if settings.JSONData != nil {
		jsonDataBytes = []byte(settings.JSONData)
		// Pretty-print the JSON
		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, jsonDataBytes, "", "  "); err == nil {
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

func FieldTypeToString(ft data.FieldType) string {
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
	case data.FieldTypeNullableJSON:
		return "FieldTypeNullableJSON"
	case data.FieldTypeEnum:
		return "FieldTypeEnum"
	case data.FieldTypeNullableEnum:
		return "FieldTypeNullableEnum"
	default:
		return "FieldTypeUnknown"
	}
}

func HTTPClientOptionsToString(opts httpclient.Options) string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("HTTPClientOptions{ForwardHTTPHeaders:%t", opts.ForwardHTTPHeaders))
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
