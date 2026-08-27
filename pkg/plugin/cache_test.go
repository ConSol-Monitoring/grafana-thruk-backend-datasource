package plugin

import "testing"

func TestRewriteAliasedEndpoints(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		table       string
		wantTable   string
		wantChanged bool
	}{
		{name: "index", table: "/index", wantTable: "/", wantChanged: true},
		{name: "stats", table: "/thruk/stats", wantTable: "/thruk/metrics", wantChanged: true},
		{name: "node control", table: "/thruk/node-control/nodes", wantTable: "/thruk/nc/nodes", wantChanged: true},
		{name: "not an alias", table: "/hosts", wantTable: "/hosts", wantChanged: false},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			//nolint:exhaustruct_v5 // only the table participates in alias rewriting
			queryModel := QueryModel{Table: testCase.table}

			changed := rewriteAliasedEndpoints(&queryModel)
			if changed != testCase.wantChanged {
				t.Fatalf("rewriteAliasedEndpoints() changed = %t, want %t", changed, testCase.wantChanged)
			}

			if queryModel.Table != testCase.wantTable {
				t.Fatalf("rewriteAliasedEndpoints() table = %q, want %q", queryModel.Table, testCase.wantTable)
			}
		})
	}
}
