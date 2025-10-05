package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Ground truth: docs/api/json_rpc_methods.md §External Stream Discovery Methods
// Reuse client by calling SendJSONRPC directly for methods without wrappers

func TestExternalStreams_DiscoveryOperatorOnly(t *testing.T) {
	fixture := NewE2EFixture(t)

	// Viewer denied
	require.NoError(t, fixture.ConnectAndAuthenticate(RoleViewer))
	denied, err := fixture.client.SendJSONRPC("discover_external_streams", map[string]interface{}{})
	require.NoError(t, err)
	require.NotNil(t, denied.Error)
	ValidateJSONRPCError(t, denied.Error, -32002, "Permission")
	fixture.client.Close()

	// Operator allowed
	require.NoError(t, fixture.ConnectAndAuthenticate(RoleOperator))
	resp, err := fixture.client.SendJSONRPC("discover_external_streams", map[string]interface{}{})
	require.NoError(t, err)
	// If feature disabled, expect Unsupported (-32030). Otherwise result schema.
	if resp.Error != nil {
		ValidateJSONRPCError(t, resp.Error, -32030, "Unsupported")
		return
	}
	result := resp.Result.(map[string]interface{})
	assert.Contains(t, result, "discovered_streams")
}

func TestExternalStreams_AddRemove(t *testing.T) {
	fixture := NewE2EFixture(t)
	require.NoError(t, fixture.ConnectAndAuthenticate(RoleOperator))

	addResp, err := fixture.client.AddExternalStream("rtsp://127.0.0.1:8554/demo", "Demo")
	require.NoError(t, err)
	if addResp.Error != nil {
		// If unsupported/disabled, assert error code
		ValidateJSONRPCError(t, addResp.Error, -32030, "Unsupported")
		return
	}

	remResp, err := fixture.client.RemoveExternalStream("rtsp://127.0.0.1:8554/demo")
	require.NoError(t, err)
	if remResp.Error != nil {
		ValidateJSONRPCError(t, remResp.Error, -32030, "Unsupported")
	}
}

func TestExternalStreams_GetAndInterval(t *testing.T) {
	fixture := NewE2EFixture(t)

	// Viewer can list
	require.NoError(t, fixture.ConnectAndAuthenticate(RoleViewer))
	getResp, err := fixture.client.SendJSONRPC("get_external_streams", nil)
	require.NoError(t, err)
	if getResp.Error == nil {
		_, ok := getResp.Result.(map[string]interface{})
		assert.True(t, ok)
	}
	fixture.client.Close()

	// Admin can set interval
	require.NoError(t, fixture.ConnectAndAuthenticate(RoleAdmin))
	updResp, err := fixture.client.SendJSONRPC("set_discovery_interval", map[string]interface{}{"scan_interval": 0})
	require.NoError(t, err)
	if updResp.Error != nil {
		ValidateJSONRPCError(t, updResp.Error, -32030, "Unsupported")
	}
}
