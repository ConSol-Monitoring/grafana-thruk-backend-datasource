package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

func TestSendBadRequest(t *testing.T) {
	t.Parallel()

	var sentResponse *backend.CallResourceResponse

	sender := backend.CallResourceResponseSenderFunc(func(response *backend.CallResourceResponse) error {
		sentResponse = response

		return nil
	})

	err := sendBadRequest(sender, ErrInvalidAPIPath)
	if err != nil {
		t.Fatalf("sendBadRequest() returned unexpected error: %v", err)
	}

	if sentResponse == nil {
		t.Fatal("sendBadRequest() did not send a response")
	}

	if sentResponse.Status != http.StatusBadRequest {
		t.Fatalf("sendBadRequest() status = %d, want %d", sentResponse.Status, http.StatusBadRequest)
	}

	if string(sentResponse.Body) != ErrInvalidAPIPath.Error() {
		t.Fatalf("sendBadRequest() body = %q, want %q", sentResponse.Body, ErrInvalidAPIPath.Error())
	}
}

func TestSendBadRequestReturnsSenderError(t *testing.T) {
	t.Parallel()

	wantErr := fmt.Errorf("test sender failure: %w", ErrInvalidAPIPath)
	sender := backend.CallResourceResponseSenderFunc(func(_ *backend.CallResourceResponse) error {
		return wantErr
	})

	err := sendBadRequest(sender, ErrInvalidAPIPath)
	if !errors.Is(err, wantErr) {
		t.Fatalf("sendBadRequest() error = %v, want wrapped %v", err, wantErr)
	}
}

func TestQueryData(t *testing.T) {
	t.Parallel()

	datasource := Datasource{
		url:        "",
		httpClient: nil,
		uid:        "",
		loggers: &Loggers{
			sdk: log.NewNullLogger(),
		},
	}

	resp, err := datasource.QueryData(
		context.Background(),
		//nolint:exhaustruct
		&backend.QueryDataRequest{
			//nolint: exhaustruct_v5
			PluginContext: backend.PluginContext{},
			Headers:       map[string]string{},
			//nolint: exhaustruct_v5
			Queries: []backend.DataQuery{
				{RefID: "A"},
			},
			Format: backend.DataFrameFormat_JSON,
		},
	)
	if err != nil {
		t.Error(err)
	}

	if len(resp.Responses) != 1 {
		t.Fatal("QueryData must return a response")
	}
}

func TestQueryDataRunsQueriesConcurrently(t *testing.T) {
	t.Parallel()

	var mutex sync.Mutex

	inflight, maxInflight := 0, 0

	// The test server sleeps so that multiple in-flight requests overlap when
	// QueryData executes the queries concurrently.
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		mutex.Lock()
		inflight++
		maxInflight = max(maxInflight, inflight)
		mutex.Unlock()

		defer func() {
			mutex.Lock()
			inflight--
			mutex.Unlock()
		}()

		time.Sleep(100 * time.Millisecond)

		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"data": [], "meta": {"columns": []}}`))
	}))
	defer server.Close()

	datasource := Datasource{
		url:        server.URL,
		httpClient: server.Client(),
		uid:        "test-uid",
		loggers: &Loggers{
			sdk: log.NewNullLogger(),
		},
	}

	const numQueries = 5

	queries := make([]backend.DataQuery, numQueries)
	for i := range queries {
		//nolint: exhaustruct_v5
		queries[i] = backend.DataQuery{
			RefID: string(rune('A' + i)),
			JSON:  json.RawMessage(`{"table": "/hosts", "columns": ["name"], "limit": 10}`),
		}
	}

	resp, err := datasource.QueryData(
		context.Background(),
		//nolint:exhaustruct
		&backend.QueryDataRequest{
			//nolint: exhaustruct_v5
			PluginContext: backend.PluginContext{},
			Headers:       map[string]string{},
			Queries:       queries,
			Format:        backend.DataFrameFormat_JSON,
		},
	)
	if err != nil {
		t.Fatalf("QueryData returned unexpected error: %v", err)
	}

	if len(resp.Responses) != numQueries {
		t.Fatalf("QueryData returned %d responses, want %d", len(resp.Responses), numQueries)
	}

	for _, q := range queries {
		if _, ok := resp.Responses[q.RefID]; !ok {
			t.Fatalf("missing response for refId %s", q.RefID)
		}
	}

	mutex.Lock()
	defer mutex.Unlock()

	if maxInflight < 2 {
		t.Fatalf("queries were not executed concurrently: max in-flight requests = %d, want >= 2", maxInflight)
	}
}

func TestQueryDataRecoversPanic(t *testing.T) {
	t.Parallel()

	// A datasource without an http client makes query() panic inside the
	// concurrent query handler. The SDK's concurrent.QueryData must recover
	// that panic into an error response for the query instead of crashing
	// the plugin process.
	datasource := Datasource{
		url:        "http://thruk.invalid",
		httpClient: nil,
		uid:        "test-uid",
		loggers: &Loggers{
			sdk: log.NewNullLogger(),
		},
	}

	resp, err := datasource.QueryData(
		context.Background(),
		//nolint:exhaustruct
		&backend.QueryDataRequest{
			//nolint: exhaustruct_v5
			PluginContext: backend.PluginContext{},
			Headers:       map[string]string{},
			Queries: []backend.DataQuery{
				//nolint: exhaustruct_v5
				{
					RefID: "A",
					JSON:  json.RawMessage(`{"table": "/hosts", "limit": 10}`),
				},
			},
			Format: backend.DataFrameFormat_JSON,
		},
	)
	if err != nil {
		t.Fatalf("QueryData returned unexpected error: %v", err)
	}

	res := resp.Responses["A"]
	if res.Status != backend.StatusInternal {
		t.Fatalf("expected StatusInternal after panic, got %v (error: %v)", res.Status, res.Error)
	}
}
