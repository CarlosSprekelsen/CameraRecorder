package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Ground truth: docs/api/json_rpc_methods.md §Recording File Management, §Snapshot File Management
// Reuse client methods: ListRecordings/ListRecordingsWithPagination, ListSnapshots/ListSnapshotsWithPagination,
// GetRecordingInfo/GetSnapshotInfo, DeleteRecording/DeleteSnapshot

func TestListRecordings_EmptySetContract(t *testing.T) {
	t.Parallel()
	fixture := NewE2EFixture(t)
	require.NoError(t, fixture.ConnectAndAuthenticate(RoleViewer))

	resp, err := fixture.client.ListRecordings()
	require.NoError(t, err)
	require.Nil(t, resp.Error)

	// Schema per docs: result { files: [], total, limit, offset }
	result := resp.Result.(map[string]interface{})
	assert.Contains(t, result, "files")
	assert.Contains(t, result, "total")
	assert.Contains(t, result, "limit")
	assert.Contains(t, result, "offset")
}

func TestListRecordings_PaginationContract(t *testing.T) {
	// Serial: listing is read-only, but keep serial to avoid flakiness from concurrent creators
	fixture := NewE2EFixture(t)
	require.NoError(t, fixture.ConnectAndAuthenticate(RoleViewer))

	resp, err := fixture.client.ListRecordingsWithPagination(10, 0)
	require.NoError(t, err)
	require.Nil(t, resp.Error)

	result := resp.Result.(map[string]interface{})
	assert.Equal(t, float64(10), result["limit"]) // JSON numbers decode as float64
	assert.Equal(t, float64(0), result["offset"])
	assert.Contains(t, result, "files")
}

func TestListSnapshots_EmptySetContract(t *testing.T) {
	t.Parallel()
	fixture := NewE2EFixture(t)
	require.NoError(t, fixture.ConnectAndAuthenticate(RoleViewer))

	resp, err := fixture.client.ListSnapshots()
	require.NoError(t, err)
	require.Nil(t, resp.Error)

	result := resp.Result.(map[string]interface{})
	assert.Contains(t, result, "files")
	assert.Contains(t, result, "total")
	assert.Contains(t, result, "limit")
	assert.Contains(t, result, "offset")
}

func TestGetRecordingInfo_Schema(t *testing.T) {
	// Assumes at least one recording may exist; when not present, server should return -32010 Not Found
	fixture := NewE2EFixture(t)
	require.NoError(t, fixture.ConnectAndAuthenticate(RoleViewer))

	resp, err := fixture.client.GetRecordingInfo("nonexistent_recording")
	require.NoError(t, err)
	if resp.Error != nil {
		// Validate correct error contract when not found
		ValidateJSONRPCError(t, resp.Error, -32010, "not found")
		return
	}

	// Otherwise validate schema
	result := resp.Result.(map[string]interface{})
	assert.Contains(t, result, "filename")
	assert.Contains(t, result, "file_size")
	assert.Contains(t, result, "created_time")
	assert.Contains(t, result, "download_url")
}

func TestGetSnapshotInfo_Schema(t *testing.T) {
	fixture := NewE2EFixture(t)
	require.NoError(t, fixture.ConnectAndAuthenticate(RoleViewer))

	resp, err := fixture.client.GetSnapshotInfo("nonexistent_snapshot.jpg")
	require.NoError(t, err)
	if resp.Error != nil {
		ValidateJSONRPCError(t, resp.Error, -32010, "not found")
		return
	}

	result := resp.Result.(map[string]interface{})
	assert.Contains(t, result, "filename")
	assert.Contains(t, result, "file_size")
	assert.Contains(t, result, "created_time")
	assert.Contains(t, result, "download_url")
}

func TestDeleteRecording_AdminOnly(t *testing.T) {
	// Validate permission matrix: viewer/operator denied, admin allowed
	fixture := NewE2EFixture(t)

	// Viewer denied
	require.NoError(t, fixture.ConnectAndAuthenticate(RoleViewer))
	resp, err := fixture.client.DeleteRecording("nonexistent_recording")
	require.NoError(t, err)
	require.NotNil(t, resp.Error)
	ValidateJSONRPCError(t, resp.Error, -32002, "Permission")
	fixture.client.Close()

	// Admin allowed (may still return not found, validate error code)
	require.NoError(t, fixture.ConnectAndAuthenticate(RoleAdmin))
	resp2, err := fixture.client.DeleteRecording("nonexistent_recording")
	require.NoError(t, err)
	if resp2.Error != nil {
		// Deleting nonexistent file should be Not Found per spec
		ValidateJSONRPCError(t, resp2.Error, -32010, "not found")
	} else {
		// If file existed and was deleted, validate shape
		result := resp2.Result.(map[string]interface{})
		assert.Contains(t, result, "filename")
		assert.Contains(t, result, "deleted")
	}
}
