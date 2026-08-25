package plugin

import (
	"context"
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"go.uber.org/zap"
)

func TestQueryData(t *testing.T) {
	ds := Datasource{
		url:        "",
		httpClient: nil,
		uid:        "",
		logger:     zap.NewNop().Sugar(),
	}

	resp, err := ds.QueryData(
		context.Background(),
		&backend.QueryDataRequest{
			PluginContext: backend.PluginContext{},
			Headers:       map[string]string{},
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
