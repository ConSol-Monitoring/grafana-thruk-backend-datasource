package plugin

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

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
			sdk:        log.NewNullLogger(),
			fileLogger: nil,
			fileClose:  nil,
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

func TestDisposeClosesFileLogger(t *testing.T) {
	t.Parallel()

	closed := false

	datasource := Datasource{
		url:        "",
		httpClient: nil,
		uid:        "",
		loggers: &Loggers{
			sdk:        log.NewNullLogger(),
			fileLogger: nil,
			fileClose: func() {
				closed = true
			},
		},
	}

	datasource.Dispose()

	if !closed {
		t.Fatal("expected Dispose to flush and close the file logger")
	}
}
