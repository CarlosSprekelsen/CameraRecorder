/*
MediaMTX Stream Manager Unit Tests

Requirements Coverage:
- REQ-MTX-001: MediaMTX service integration
- REQ-MTX-002: Stream management capabilities
- REQ-MTX-003: Path creation and deletion

Test Categories: Unit (using real MediaMTX server as per guidelines)
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package mediamtx

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewStreamManager_ReqMTX001 tests stream manager creation
func TestNewStreamManager_ReqMTX001(t *testing.T) {
	// REQ-MTX-001: MediaMTX service integration
	EnsureSequentialExecution(t)
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Server is ready via shared test helper

	// Use shared stream manager from test helper
	streamManager := helper.GetStreamManager()
	require.NotNil(t, streamManager, "Stream manager should not be nil")
}

// TestStreamManager_CreateStream_ReqMTX002 tests stream creation
func TestStreamManager_CreateStream_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	EnsureSequentialExecution(t)
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Server is ready via shared test helper

	// Use shared stream manager from test helper
	streamManager := helper.GetStreamManager()
	require.NotNil(t, streamManager)

	ctx := context.Background()
	testStreamName := "test_stream_" + time.Now().Format("20060102_150405")
	testSource := "publisher"

	// Create a stream
	stream, err := streamManager.CreateStream(ctx, testStreamName, testSource)
	require.NoError(t, err, "Stream creation should succeed")
	require.NotNil(t, stream, "Created stream should not be nil")
	assert.Equal(t, testStreamName, stream.Name, "Stream name should match")

	// Clean up
	err = streamManager.DeleteStream(ctx, testStreamName)
	require.NoError(t, err, "Stream deletion should succeed")
}

// TestStreamManager_DeleteStream_ReqMTX002 tests stream deletion
func TestStreamManager_DeleteStream_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	EnsureSequentialExecution(t)
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Server is ready via shared test helper

	// Use shared stream manager from test helper
	streamManager := helper.GetStreamManager()
	require.NotNil(t, streamManager)

	ctx := context.Background()
	testStreamName := "test_delete_stream_" + time.Now().Format("20060102_150405")

	// Create a stream first
	_, err := streamManager.CreateStream(ctx, testStreamName, "publisher")
	require.NoError(t, err, "Stream creation should succeed")

	// Delete the stream
	err = streamManager.DeleteStream(ctx, testStreamName)
	require.NoError(t, err, "Stream deletion should succeed")
}

// TestStreamManager_GetStream_ReqMTX002 tests stream retrieval
func TestStreamManager_GetStream_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	EnsureSequentialExecution(t)
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Server is ready via shared test helper

	// Use shared stream manager from test helper
	streamManager := helper.GetStreamManager()
	require.NotNil(t, streamManager)

	ctx := context.Background()
	testStreamName := "test_get_stream_" + time.Now().Format("20060102_150405")

	// Create a stream first
	_, err := streamManager.CreateStream(ctx, testStreamName, "publisher")
	require.NoError(t, err, "Stream creation should succeed")

	// Get the stream
	stream, err := streamManager.GetStream(ctx, testStreamName)
	require.NoError(t, err, "Stream retrieval should succeed")
	require.NotNil(t, stream, "Retrieved stream should not be nil")
	assert.Equal(t, testStreamName, stream.Name, "Stream name should match")

	// Clean up
	err = streamManager.DeleteStream(ctx, testStreamName)
	require.NoError(t, err, "Stream deletion should succeed")
}

// TestStreamManager_ListStreams_ReqMTX002 tests stream listing
func TestStreamManager_ListStreams_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	EnsureSequentialExecution(t)
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Server is ready via shared test helper

	// Use shared stream manager from test helper
	streamManager := helper.GetStreamManager()
	require.NotNil(t, streamManager)

	ctx := context.Background()

	// List all streams
	streams, err := streamManager.ListStreams(ctx)
	require.NoError(t, err, "Stream listing should succeed")
	require.NotNil(t, streams, "Streams list should not be nil")
	assert.GreaterOrEqual(t, len(streams), 0, "Should return at least 0 streams")
}

// TestStreamManager_StartRecordingStream_ReqMTX002 tests recording stream creation
func TestStreamManager_StartRecordingStream_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	EnsureSequentialExecution(t)
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Server is ready via shared test helper

	// Use shared stream manager from test helper
	streamManager := helper.GetStreamManager()
	require.NotNil(t, streamManager)

	ctx := context.Background()
	devicePath := "/dev/video0"

	// Start recording stream
	stream, err := streamManager.StartStream(ctx, devicePath)
	require.NoError(t, err, "Recording stream creation should succeed")
	require.NotNil(t, stream, "Created recording stream should not be nil")
	assert.Equal(t, "camera0", stream.Name, "Recording stream should have base camera name without suffix")

	// Clean up
	err = streamManager.DeleteStream(ctx, stream.Name)
	require.NoError(t, err, "Stream deletion should succeed")
}

// TestStreamManager_StartStream_Viewing_ReqMTX002 tests stream creation for viewing
func TestStreamManager_StartStream_Viewing_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	EnsureSequentialExecution(t)
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Server is ready via shared test helper

	// Use shared stream manager from test helper
	streamManager := helper.GetStreamManager()
	require.NotNil(t, streamManager)

	ctx := context.Background()
	devicePath := "/dev/video0"

	// Start stream using single path approach (no separate viewing stream)
	stream, err := streamManager.StartStream(ctx, devicePath)
	require.NoError(t, err, "Stream creation should succeed")
	require.NotNil(t, stream, "Created stream should not be nil")
	assert.NotEmpty(t, stream.Name, "Stream name should not be empty")

	// Clean up
	err = streamManager.DeleteStream(ctx, stream.Name)
	require.NoError(t, err, "Stream deletion should succeed")
}

// TestStreamManager_StartStream_Snapshot_ReqMTX002 tests stream creation for snapshots
func TestStreamManager_StartStream_Snapshot_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	EnsureSequentialExecution(t)
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Server is ready via shared test helper

	// Use shared stream manager from test helper
	streamManager := helper.GetStreamManager()
	require.NotNil(t, streamManager)

	ctx := context.Background()
	// Use test fixture for external RTSP source (Tier 3 scenario)
	devicePath := helper.GetTestCameraDevice("network_failure")

	// Start stream using single path approach (no separate snapshot stream)
	stream, err := streamManager.StartStream(ctx, devicePath)
	require.NoError(t, err, "Stream creation should succeed")
	require.NotNil(t, stream, "Created stream should not be nil")
	assert.NotEmpty(t, stream.Name, "Stream name should not be empty")

	// Clean up
	err = streamManager.DeleteStream(ctx, stream.Name)
	require.NoError(t, err, "Stream deletion should succeed")
}

// TestStreamManager_ErrorHandling_ReqMTX001 tests error handling
func TestStreamManager_ErrorHandling_ReqMTX001(t *testing.T) {
	// REQ-MTX-001: MediaMTX service integration
	EnsureSequentialExecution(t)
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Server is ready via shared test helper

	// Use shared stream manager from test helper
	streamManager := helper.GetStreamManager()
	require.NotNil(t, streamManager)

	ctx := context.Background()

	// Test invalid stream name
	_, err := streamManager.CreateStream(ctx, "", "publisher")
	require.Error(t, err, "Empty stream name should cause error")

	// Test getting non-existent stream
	_, err = streamManager.GetStream(ctx, "test_non_existent_stream")
	require.Error(t, err, "Getting non-existent stream should cause error")

	// Test deleting non-existent stream
	err = streamManager.DeleteStream(ctx, "test_non_existent_stream")
	require.Error(t, err, "Deleting non-existent stream should cause error")
}
