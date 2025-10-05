package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Ground truth: docs/api/json_rpc_methods.md §System Status and Health, §System Metrics and Monitoring
// Reuse client methods: GetStatus (admin), GetSystemStatus (viewer), GetSystemMetrics (admin), GetStorageInfo (admin)

func TestGetStatus_AdminContract(t *testing.T) {
	t.Parallel()
	fixture := NewE2EFixture(t)
	require.NoError(t, fixture.ConnectAndAuthenticate(RoleAdmin))

	resp, err := fixture.client.GetStatus()
	require.NoError(t, err)
	require.Nil(t, resp.Error)

	result := resp.Result.(map[string]interface{})
	assert.Contains(t, result, "status")
	assert.Contains(t, result, "uptime")
	assert.Contains(t, result, "version")
	assert.Contains(t, result, "components")

	comps := result["components"].(map[string]interface{})
	assert.Contains(t, comps, "websocket_server")
	assert.Contains(t, comps, "camera_monitor")
	assert.Contains(t, comps, "mediamtx")
}

func TestGetSystemStatus_ViewerContract(t *testing.T) {
	t.Parallel()
	fixture := NewE2EFixture(t)
	require.NoError(t, fixture.ConnectAndAuthenticate(RoleViewer))

	resp, err := fixture.client.GetSystemStatus()
	require.NoError(t, err)
	require.Nil(t, resp.Error)

	result := resp.Result.(map[string]interface{})
	assert.Contains(t, result, "status")
	assert.Contains(t, result, "message")
	assert.Contains(t, result, "available_cameras")
	assert.Contains(t, result, "discovery_active")
}

func TestGetMetrics_AdminOnly(t *testing.T) {
	t.Parallel()
	fixture := NewE2EFixture(t)

	// Operator denied
	require.NoError(t, fixture.ConnectAndAuthenticate(RoleOperator))
	denied, err := fixture.client.GetSystemMetrics()
	require.NoError(t, err)
	require.NotNil(t, denied.Error)
	ValidateJSONRPCError(t, denied.Error, -32002, "Permission")
	fixture.client.Close()

	// Admin allowed
	require.NoError(t, fixture.ConnectAndAuthenticate(RoleAdmin))
	resp, err := fixture.client.GetSystemMetrics()
	require.NoError(t, err)
	require.Nil(t, resp.Error)

	result := resp.Result.(map[string]interface{})
	assert.Contains(t, result, "timestamp")
	assert.Contains(t, result, "system_metrics")
	assert.Contains(t, result, "camera_metrics")
	assert.Contains(t, result, "stream_metrics")
}

func TestGetStorageInfo_AdminOnly(t *testing.T) {
	fixture := NewE2EFixture(t)

	// Viewer denied
	require.NoError(t, fixture.ConnectAndAuthenticate(RoleViewer))
	denied, err := fixture.client.GetStorageInfo()
	require.NoError(t, err)
	require.NotNil(t, denied.Error)
	ValidateJSONRPCError(t, denied.Error, -32002, "Permission")
	fixture.client.Close()

	// Admin allowed
	require.NoError(t, fixture.ConnectAndAuthenticate(RoleAdmin))
	resp, err := fixture.client.GetStorageInfo()
	require.NoError(t, err)
	require.Nil(t, resp.Error)

	result := resp.Result.(map[string]interface{})
	assert.Contains(t, result, "total_space")
	assert.Contains(t, result, "used_space")
	assert.Contains(t, result, "available_space")
	assert.Contains(t, result, "usage_percentage")
}

func TestGetServerInfo_AdminContract(t *testing.T) {
	t.Parallel()
	fixture := NewE2EFixture(t)

	// Viewer denied
	require.NoError(t, fixture.ConnectAndAuthenticate(RoleViewer))
	denied, err := fixture.client.GetServerInfo()
	require.NoError(t, err)
	require.NotNil(t, denied.Error)
	ValidateJSONRPCError(t, denied.Error, -32002, "Permission")
	fixture.client.Close()

	// Admin allowed
	require.NoError(t, fixture.ConnectAndAuthenticate(RoleAdmin))
	resp, err := fixture.client.GetServerInfo()
	require.NoError(t, err)
	require.Nil(t, resp.Error)

	result := resp.Result.(map[string]interface{})
	assert.Contains(t, result, "name")
	assert.Contains(t, result, "version")
	assert.Contains(t, result, "capabilities")
}

func TestRetentionPolicyAndCleanup_AdminOnly(t *testing.T) {
	fixture := NewE2EFixture(t)

	// Operator denied
	require.NoError(t, fixture.ConnectAndAuthenticate(RoleOperator))
	deniedSet, err := fixture.client.SetRetentionPolicy("age", 30, 0, true)
	require.NoError(t, err)
	require.NotNil(t, deniedSet.Error)
	ValidateJSONRPCError(t, deniedSet.Error, -32002, "Permission")
	fixture.client.Close()

	// Admin allowed: set policy
	require.NoError(t, fixture.ConnectAndAuthenticate(RoleAdmin))
	setResp, err := fixture.client.SetRetentionPolicy("age", 30, 0, true)
	require.NoError(t, err)
	if setResp.Error != nil {
		// Validate documented error codes if unsupported
		ValidateJSONRPCError(t, setResp.Error, -32603, "")
	} else {
		result := setResp.Result.(map[string]interface{})
		assert.Contains(t, result, "policy_type")
		assert.Contains(t, result, "enabled")
	}

	// Admin allowed: cleanup
	cleanupResp, err := fixture.client.CleanupOldFiles()
	require.NoError(t, err)
	if cleanupResp.Error != nil {
		ValidateJSONRPCError(t, cleanupResp.Error, -32603, "")
	} else {
		result := cleanupResp.Result.(map[string]interface{})
		assert.Contains(t, result, "cleanup_executed")
		assert.Contains(t, result, "files_deleted")
	}
}
