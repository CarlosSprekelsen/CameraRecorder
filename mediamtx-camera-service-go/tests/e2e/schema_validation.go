package e2e

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

// ValidateJSON ensures response/result can be marshaled into the provided schema type.
// Pass a pointer to a struct in targetSchema. The function will unmarshal into it.
func ValidateJSON(t *testing.T, response interface{}, targetSchema interface{}) {
	t.Helper()

	data, err := json.Marshal(response)
	require.NoError(t, err, "failed to marshal response to JSON for schema validation")

	err = json.Unmarshal(data, targetSchema)
	require.NoError(t, err, "response does not conform to expected schema")
}

// ValidateJSONRPCError checks JSON-RPC error code and that message contains a substring.
func ValidateJSONRPCError(t *testing.T, rpcErr interface{}, expectedCode int, containsMessage string) {
	t.Helper()

	// We rely on the test client's JSONRPCError shape
	// Convert via JSON to avoid brittle type assertions across packages
	var raw map[string]interface{}
	data, err := json.Marshal(rpcErr)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(data, &raw))

	code, ok := raw["code"].(float64)
	require.True(t, ok, "error.code must be a number")
	require.Equal(t, float64(expectedCode), code)

	if containsMessage != "" {
		msg, ok := raw["message"].(string)
		require.True(t, ok, "error.message must be a string")
		require.Contains(t, msg, containsMessage)
	}
}
