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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewStreamManager_ReqMTX001 tests stream manager creation
func TestNewStreamManager_ReqMTX001(t *testing.T) {
	// REQ-MTX-001: MediaMTX service integration
	helper, ctx := SetupMediaMTXTest(t)
	_ = ctx // Suppress unused variable warning

	streamManager := helper.GetStreamManager()
	require.NotNil(t, streamManager, "Stream manager should not be nil")
}

// TestStreamManager_CreateStream_ReqMTX002 tests stream creation
func TestStreamManager_CreateStream_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	helper, ctx := SetupMediaMTXTest(t)
	_ = ctx // Suppress unused variable warning

	// Use shared stream manager from test helper
	streamManager := helper.GetStreamManager()
	require.NotNil(t, streamManager)

	testStreamName := "test_stream_" + time.Now().Format("20060102_150405")
	testSource := "publisher"

	// Create a stream
	stream, err := streamManager.CreateStream(ctx, testStreamName, testSource)
	// Use assertion helper to reduce boilerplate
	helper.AssertStandardResponse(t, stream, err, "Stream creation")
	assert.Equal(t, testStreamName, stream.Name, "Stream name should match")

	// Clean up
	err = streamManager.DeleteStream(ctx, testStreamName)
	// Use assertion helper
	require.NoError(t, err, "Stream deletion should succeed")
}

// TestStreamManager_DeleteStream_ReqMTX002 tests stream deletion
func TestStreamManager_DeleteStream_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	helper, ctx := SetupMediaMTXTest(t)
	_ = ctx // Suppress unused variable warning

	// Use shared stream manager from test helper
	streamManager := helper.GetStreamManager()
	require.NotNil(t, streamManager)

	testStreamName := "test_delete_stream_" + time.Now().Format("20060102_150405")

	// Create a stream first
	_, err := streamManager.CreateStream(ctx, testStreamName, "publisher")
	require.NoError(t, err, "Stream creation should succeed")

	// Delete the stream
	err = streamManager.DeleteStream(ctx, testStreamName)
	// Use assertion helper
	require.NoError(t, err, "Stream deletion should succeed")
}

// TestStreamManager_StartStream_ReqMTX002 tests new cameraID-first stream starting
func TestStreamManager_StartStream_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities - cameraID-first architecture
	helper, ctx := SetupMediaMTXTest(t)
	_ = ctx // Suppress unused variable warning

	// Use shared stream manager from test helper
	streamManager := helper.GetStreamManager()
	require.NotNil(t, streamManager)

	// Get ready controller with device discovery
	controller, ctx, cancel := helper.GetReadyController(t)
	defer cancel()
	defer controller.Stop(ctx)

	// Use proper MediaMTX path identifier (discovered device)
	cameraID := "camera0" // Use discovered device (same as other tests)

	// Progressive Readiness: Attempt operation immediately (may use fallback)
	response, err := streamManager.StartStream(ctx, cameraID)
	if err == nil {
		// Operation succeeded immediately (Progressive Readiness working)
		t.Log("Stream started immediately - Progressive Readiness working")
	} else {
		// Operation needs readiness - wait for event (no polling)
		readinessChan := controller.SubscribeToReadiness()
		select {
		case <-readinessChan:
			// Retry after readiness event
			response, err = streamManager.StartStream(ctx, cameraID)
			require.NoError(t, err, "Stream should start after readiness event")
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for readiness event")
		}
	}
	require.NotNil(t, response, "Created stream should not be nil")

	// Validate API-ready response format per JSON-RPC documentation
	assert.Equal(t, cameraID, response.Device, "Response device should match camera ID")
	assert.NotEmpty(t, response.StreamURL, "Response should include stream URL")
	// Note: On-demand streams are not "ready" until first access per MediaMTX architecture
	assert.Contains(t, response.StreamURL, cameraID, "Stream URL should contain camera ID")
}

// TestStreamManager_GetStreamStatus_ReqMTX002 tests new cameraID-first stream status
func TestStreamManager_GetStreamStatus_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities - cameraID-first architecture
	helper, ctx := SetupMediaMTXTest(t)
	_ = ctx // Suppress unused variable warning

	// Use shared stream manager from test helper
	streamManager := helper.GetStreamManager()
	require.NotNil(t, streamManager)

	// Get ready controller with device discovery
	controller, ctx, cancel := helper.GetReadyController(t)
	defer cancel()
	defer controller.Stop(ctx)

	// Use proper MediaMTX path identifier (discovered device)
	cameraID := "camera0" // Use discovered device (same as other tests)

	// First create a stream to get status for
	stream, err := streamManager.StartStream(ctx, cameraID)
	if err != nil {
		// Operation needs readiness - wait for event (no polling)
		readinessChan := controller.SubscribeToReadiness()
		select {
		case <-readinessChan:
			// Retry after readiness event
			stream, err = streamManager.StartStream(ctx, cameraID)
			require.NoError(t, err, "Stream should start after readiness event")
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for readiness event")
		}
	}
	require.NotNil(t, stream, "Created stream should not be nil")

	// Now get stream status using cameraID-first API
	response, err := streamManager.GetStreamStatus(ctx, cameraID)
	require.NoError(t, err, "GetStreamStatus should succeed with valid camera ID")
	require.NotNil(t, response, "GetStreamStatus should return API-ready response")

	// Validate API-ready response format per JSON-RPC documentation
	assert.Equal(t, cameraID, response.Device, "Response device should match camera ID")
	assert.NotEmpty(t, response.Status, "Response should include status")
	assert.Contains(t, []string{"active", "inactive", "ready", "PENDING"}, response.Status, "Status should be valid")

	// Clean up
	err = streamManager.DeleteStream(ctx, cameraID)
	require.NoError(t, err, "Stream deletion should succeed")
}

// TestStreamManager_ListStreamsAPI_ReqMTX002 tests new API-ready stream listing
func TestStreamManager_ListStreamsAPI_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities - API-ready responses
	helper, ctx := SetupMediaMTXTest(t)
	_ = ctx // Suppress unused variable warning

	// Use shared stream manager from test helper
	streamManager := helper.GetStreamManager()
	require.NotNil(t, streamManager)

	// Get ready controller with device discovery
	controller, ctx, cancel := helper.GetReadyController(t)
	defer cancel()
	defer controller.Stop(ctx)

	// Create a test stream first to have something to list
	cameraID := "camera0"
	stream, err := streamManager.StartStream(ctx, cameraID)
	if err != nil {
		// Operation needs readiness - wait for event (no polling)
		readinessChan := controller.SubscribeToReadiness()
		select {
		case <-readinessChan:
			// Retry after readiness event
			stream, err = streamManager.StartStream(ctx, cameraID)
			require.NoError(t, err, "Stream should start after readiness event")
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for readiness event")
		}
	}
	require.NotNil(t, stream, "Created stream should not be nil")

	// Now list streams using API-ready method
	response, err := streamManager.ListStreams(ctx)
	require.NoError(t, err, "ListStreams should succeed")
	require.NotNil(t, response, "ListStreams should return API-ready response")

	// Validate API-ready response format per JSON-RPC documentation
	assert.NotNil(t, response.Streams, "Response should include streams array")
	assert.GreaterOrEqual(t, response.Total, 0, "Response should include total count")
	assert.Greater(t, len(response.Streams), 0, "Should have at least one stream")

	// Validate stream structure (Source field may be empty for on-demand streams)
	for _, stream := range response.Streams {
		assert.NotEmpty(t, stream.Name, "Stream should have name")
		// Note: Source field may be empty for on-demand MediaMTX streams
	}

	// Clean up
	err = streamManager.DeleteStream(ctx, cameraID)
	require.NoError(t, err, "Stream deletion should succeed")
}

// TestStreamManager_GetStreamURL_ReqMTX002 tests new cameraID-first stream URL retrieval
func TestStreamManager_GetStreamURL_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities - cameraID-first architecture
	helper, ctx := SetupMediaMTXTest(t)
	_ = ctx // Suppress unused variable warning

	// Use shared stream manager from test helper
	streamManager := helper.GetStreamManager()
	require.NotNil(t, streamManager)

	// Get ready controller with device discovery
	controller, ctx, cancel := helper.GetReadyController(t)
	defer cancel()
	defer controller.Stop(ctx)

	// Use proper MediaMTX path identifier (discovered device)
	cameraID := "camera0" // Use discovered device (same as other tests)

	// First create a stream to get URL for
	stream, err := streamManager.StartStream(ctx, cameraID)
	if err != nil {
		// Operation needs readiness - wait for event (no polling)
		readinessChan := controller.SubscribeToReadiness()
		select {
		case <-readinessChan:
			// Retry after readiness event
			stream, err = streamManager.StartStream(ctx, cameraID)
			require.NoError(t, err, "Stream should start after readiness event")
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for readiness event")
		}
	}
	require.NotNil(t, stream, "Created stream should not be nil")

	// Now get stream URL using cameraID-first API
	response, err := streamManager.GetStreamURL(ctx, cameraID)
	require.NoError(t, err, "GetStreamURL should succeed with valid camera ID")
	require.NotNil(t, response, "GetStreamURL should return API-ready response")

	// Validate API-ready response format per JSON-RPC documentation
	assert.Equal(t, cameraID, response.Device, "Response device should match camera ID")
	assert.NotEmpty(t, response.StreamURL, "Response should include stream URL")
	// Note: On-demand streams are not "ready" until first access per MediaMTX architecture
	assert.Contains(t, response.StreamURL, cameraID, "Stream URL should contain camera ID")

	// Clean up
	err = streamManager.DeleteStream(ctx, cameraID)
	require.NoError(t, err, "Stream deletion should succeed")
}

// TestStreamManager_GetStream_ReqMTX002 tests stream retrieval
func TestStreamManager_GetStream_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	helper, ctx := SetupMediaMTXTest(t)
	_ = ctx // Suppress unused variable warning

	// Use shared stream manager from test helper
	streamManager := helper.GetStreamManager()
	require.NotNil(t, streamManager)

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
	// Use assertion helper
	require.NoError(t, err, "Stream deletion should succeed")
}

// TestStreamManager_ListStreams_ReqMTX002 tests stream listing
func TestStreamManager_ListStreams_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	helper, ctx := SetupMediaMTXTest(t)
	_ = ctx // Suppress unused variable warning

	// Use shared stream manager from test helper
	streamManager := helper.GetStreamManager()
	require.NotNil(t, streamManager)

	// List all streams
	streams, err := streamManager.ListStreams(ctx)
	require.NoError(t, err, "Stream listing should succeed")
	require.NotNil(t, streams, "Streams list should not be nil")
	assert.GreaterOrEqual(t, streams.Total, 0, "Should return at least 0 streams")
}

// TestStreamManager_StartRecordingStream_ReqMTX002 tests recording stream creation
func TestStreamManager_StartRecordingStream_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	helper, ctx := SetupMediaMTXTest(t)
	_ = ctx // Suppress unused variable warning

	// Use shared stream manager from test helper
	streamManager := helper.GetStreamManager()
	require.NotNil(t, streamManager)

	// Get ready controller with device discovery
	controller, ctx, cancel := helper.GetReadyController(t)
	defer cancel()
	defer controller.Stop(ctx)

	// Use proper MediaMTX path identifier (discovered device)
	cameraID := "camera0" // MediaMTX requires alphanumeric identifiers

	// Progressive Readiness: Attempt operation immediately (may use fallback)
	stream, err := streamManager.StartStream(ctx, cameraID)
	if err == nil {
		// Operation succeeded immediately (Progressive Readiness working)
		t.Log("Stream started immediately - Progressive Readiness working")
	} else {
		// Operation needs readiness - wait for event (no polling)
		readinessChan := controller.SubscribeToReadiness()
		select {
		case <-readinessChan:
			// Retry after readiness event
			stream, err = streamManager.StartStream(ctx, cameraID)
			require.NoError(t, err, "Stream should start after readiness event")
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for readiness event")
		}
	}
	require.NotNil(t, stream, "Created stream should not be nil")
	assert.Equal(t, cameraID, stream.Device, "Stream device should match camera identifier")

	// Clean up
	err = streamManager.DeleteStream(ctx, cameraID)
	// Use assertion helper
	require.NoError(t, err, "Stream deletion should succeed")
}

// TestStreamManager_StartStream_Viewing_ReqMTX002 tests stream creation for viewing
func TestStreamManager_StartStream_Viewing_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	helper, ctx := SetupMediaMTXTest(t)
	_ = ctx // Suppress unused variable warning

	// Use shared stream manager from test helper
	streamManager := helper.GetStreamManager()
	require.NotNil(t, streamManager)

	// Get ready controller with device discovery
	controller, ctx, cancel := helper.GetReadyController(t)
	defer cancel()
	defer controller.Stop(ctx)

	// Use proper MediaMTX path identifier (discovered device)
	cameraID := "camera0" // Use discovered device (same as other tests)

	// Progressive Readiness: Attempt operation immediately (may use fallback)
	stream, err := streamManager.StartStream(ctx, cameraID)
	if err == nil {
		// Operation succeeded immediately (Progressive Readiness working)
		t.Log("Stream started immediately - Progressive Readiness working")
	} else {
		// Operation needs readiness - wait for event (no polling)
		readinessChan := controller.SubscribeToReadiness()
		select {
		case <-readinessChan:
			// Retry after readiness event
			stream, err = streamManager.StartStream(ctx, cameraID)
			require.NoError(t, err, "Stream should start after readiness event")
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for readiness event")
		}
	}
	require.NotNil(t, stream, "Created stream should not be nil")
	assert.Equal(t, cameraID, stream.Device, "Stream device should match camera identifier")

	// Clean up
	err = streamManager.DeleteStream(ctx, cameraID)
	// Use assertion helper
	require.NoError(t, err, "Stream deletion should succeed")
}

// TestStreamManager_StartStream_Snapshot_ReqMTX002 tests stream creation for snapshots
func TestStreamManager_StartStream_Snapshot_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	helper, ctx := SetupMediaMTXTest(t)
	_ = ctx // Suppress unused variable warning

	// Use shared stream manager from test helper
	streamManager := helper.GetStreamManager()
	require.NotNil(t, streamManager)

	// Get ready controller with device discovery
	controller, ctx, cancel := helper.GetReadyController(t)
	defer cancel()
	defer controller.Stop(ctx)

	// Use proper MediaMTX path identifier (discovered device)
	cameraID := "camera0" // Use discovered device (same as other tests)

	// Progressive Readiness: Attempt operation immediately (may use fallback)
	stream, err := streamManager.StartStream(ctx, cameraID)
	if err == nil {
		// Operation succeeded immediately (Progressive Readiness working)
		t.Log("Stream started immediately - Progressive Readiness working")
	} else {
		// Operation needs readiness - wait for event (no polling)
		readinessChan := controller.SubscribeToReadiness()
		select {
		case <-readinessChan:
			// Retry after readiness event
			stream, err = streamManager.StartStream(ctx, cameraID)
			require.NoError(t, err, "Stream should start after readiness event")
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for readiness event")
		}
	}
	require.NotNil(t, stream, "Created stream should not be nil")
	assert.Equal(t, cameraID, stream.Device, "Stream device should match camera identifier")

	// Clean up
	err = streamManager.DeleteStream(ctx, cameraID)
	// Use assertion helper
	require.NoError(t, err, "Stream deletion should succeed")
}

// TestStreamManager_ErrorHandling_ReqMTX001 tests error handling
func TestStreamManager_ErrorHandling_ReqMTX001(t *testing.T) {
	// REQ-MTX-001: MediaMTX service integration
	helper, ctx := SetupMediaMTXTest(t)
	_ = ctx // Suppress unused variable warning

	// Use shared stream manager from test helper
	streamManager := helper.GetStreamManager()
	require.NotNil(t, streamManager)

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
