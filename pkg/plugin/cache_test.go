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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			queryModel := QueryModel{Table: tt.table}
			changed := rewriteAliasedEndpoints(&queryModel)
			if changed != tt.wantChanged {
				t.Fatalf("rewriteAliasedEndpoints() changed = %t, want %t", changed, tt.wantChanged)
			}

			if queryModel.Table != tt.wantTable {
				t.Fatalf("rewriteAliasedEndpoints() table = %q, want %q", queryModel.Table, tt.wantTable)
			}
		})
	}
}
