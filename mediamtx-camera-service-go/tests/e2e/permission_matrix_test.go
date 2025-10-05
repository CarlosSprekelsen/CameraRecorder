package e2e

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

// Core subset per docs/api/json_rpc_methods.md §Permissions Matrix
// Validates allow/deny without adapting to implementation.

func TestPermissionMatrix_CoreSubset(t *testing.T) {
	t.Parallel()

	type tc struct {
		method  string
		role    string
		allowed bool
	}

	cases := []tc{
		{"get_metrics", RoleAdmin, true},
		{"get_metrics", RoleOperator, false},
		{"get_metrics", RoleViewer, false},
		{"delete_recording", RoleAdmin, true},
		{"delete_recording", RoleOperator, false},
		{"delete_recording", RoleViewer, false},
		{"get_status", RoleAdmin, true},
		{"get_status", RoleOperator, false},
		{"get_status", RoleViewer, false},
		{"get_system_status", RoleAdmin, true},
		{"get_system_status", RoleOperator, true},
		{"get_system_status", RoleViewer, true},
	}

	for _, c := range cases {
		c := c
		t.Run(fmt.Sprintf("%s_%s", c.method, c.role), func(t *testing.T) {
			fixture := NewE2EFixture(t)
			require.NoError(t, fixture.ConnectAndAuthenticate(c.role))

			var resp *JSONRPCResponseAlias
			var err error
			switch c.method {
			case "get_metrics":
				r, e := fixture.client.GetSystemMetrics()
				resp, err = &JSONRPCResponseAlias{Error: r.Error}, e
			case "get_status":
				r, e := fixture.client.GetStatus()
				resp, err = &JSONRPCResponseAlias{Error: r.Error}, e
			case "get_system_status":
				r, e := fixture.client.GetSystemStatus()
				resp, err = &JSONRPCResponseAlias{Error: r.Error}, e
			case "delete_recording":
				r, e := fixture.client.DeleteRecording("nonexistent_recording")
				resp, err = &JSONRPCResponseAlias{Error: r.Error}, e
			default:
				t.Fatalf("unhandled method in matrix: %s", c.method)
			}

			require.NoError(t, err)
			if c.allowed {
				require.Nil(t, resp.Error)
			} else {
				require.NotNil(t, resp.Error)
				ValidateJSONRPCError(t, resp.Error, -32002, "Permission")
			}
		})
	}
}

// JSONRPCResponseAlias is a minimal alias to avoid importing testutils here.
// We only need the Error field for permission checks.
type JSONRPCResponseAlias struct {
	Error interface{} `json:"error,omitempty"`
}
