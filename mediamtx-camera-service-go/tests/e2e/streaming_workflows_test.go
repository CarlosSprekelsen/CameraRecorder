package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Ground truth: docs/api/json_rpc_methods.md §Streaming Methods
// Reuse client methods: StartStreaming/StopStreaming/GetStreamURL/GetStreamStatus/GetStreams

func TestGetStreamURL_ViewerContract(t *testing.T) {
	t.Parallel()
	fixture := NewE2EFixture(t)
	require.NoError(t, fixture.ConnectAndAuthenticate(RoleViewer))

	resp, err := fixture.client.GetStreamURL(DefaultCameraID)
	require.NoError(t, err)
	require.Nil(t, resp.Error)

	result := resp.Result.(map[string]interface{})
	assert.Contains(t, result, "device")
	assert.Contains(t, result, "stream_name")
	assert.Contains(t, result, "stream_url")
	assert.Contains(t, result, "available")
}

func TestStartStopStreaming_OperatorOnly(t *testing.T) {
	fixture := NewE2EFixture(t)

	// Viewer denied
	require.NoError(t, fixture.ConnectAndAuthenticate(RoleViewer))
	denied, err := fixture.client.StartStreaming(DefaultCameraID)
	require.NoError(t, err)
	require.NotNil(t, denied.Error)
	ValidateJSONRPCError(t, denied.Error, -32002, "Permission")
	fixture.client.Close()

	// Operator allowed
	require.NoError(t, fixture.ConnectAndAuthenticate(RoleOperator))
	startResp, err := fixture.client.StartStreaming(DefaultCameraID)
	require.NoError(t, err)
	require.Nil(t, startResp.Error)

	// Validate minimal schema
	start := startResp.Result.(map[string]interface{})
	assert.Contains(t, start, "device")
	assert.Contains(t, start, "stream_name")
	assert.Contains(t, start, "stream_url")
	assert.Contains(t, start, "status")

	stopResp, err := fixture.client.StopStreaming(DefaultCameraID)
	require.NoError(t, err)
	require.Nil(t, stopResp.Error)
}

func TestGetStreamStatus_ViewerContract(t *testing.T) {
	t.Parallel()
	fixture := NewE2EFixture(t)
	require.NoError(t, fixture.ConnectAndAuthenticate(RoleViewer))

	resp, err := fixture.client.GetStreamStatus(DefaultCameraID)
	require.NoError(t, err)
	// May be active or not; still validate envelope
	if resp.Error != nil {
		// If stream not found, error catalog allows Not Found (-32010)
		ValidateJSONRPCError(t, resp.Error, -32010, "not")
		return
	}

	result := resp.Result.(map[string]interface{})
	assert.Contains(t, result, "device")
	assert.Contains(t, result, "stream_name")
	assert.Contains(t, result, "status")
}

func TestGetStreams_List(t *testing.T) {
	t.Parallel()
	fixture := NewE2EFixture(t)
	require.NoError(t, fixture.ConnectAndAuthenticate(RoleViewer))

	resp, err := fixture.client.GetStreams()
	require.NoError(t, err)
	if resp.Error != nil {
		// MediaMTX unavailable yields Dependency Failed (-32050)
		ValidateJSONRPCError(t, resp.Error, -32050, "Dependency")
		return
	}
	// Expect array result
	_, ok := resp.Result.([]interface{})
	assert.True(t, ok, "result should be an array of streams")
}
