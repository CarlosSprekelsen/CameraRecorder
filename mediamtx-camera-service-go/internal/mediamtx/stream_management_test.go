/*
MediaMTX Stream Management Tests - Real Server Integration

Requirements Coverage:
- REQ-MTX-001: MediaMTX service integration
- REQ-MTX-002: Stream management capabilities
- REQ-MTX-004: Health monitoring

Test Categories: Unit (using real MediaMTX server)
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package mediamtx

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestController_GetStreams_Management_ReqMTX002 tests getting all streams
func TestController_GetStreams_Management_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create controller
	controller, err := helper.GetController(t)
	require.NoError(t, err, "Controller creation should succeed")
	require.NotNil(t, controller, "Controller should not be nil")

	// Start the controller
	ctx, cancel := context.WithTimeout(context.Background(), TestTimeoutExtreme)
	defer cancel()
	err = controller.Start(ctx)
	require.NoError(t, err, "Controller start should succeed")

	// Ensure controller is stopped after test
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		controller.Stop(stopCtx)
	}()

	// Get all streams
	streams, err := controller.GetStreams(ctx)
	require.NoError(t, err, "Getting streams should succeed")
	require.NotNil(t, streams, "Streams should not be nil")
	assert.IsType(t, []*Path{}, streams, "Should return slice of Path")
}

// TestController_GetStream_Management_ReqMTX002 tests getting a specific stream
func TestController_GetStream_Management_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create controller
	controller, err := helper.GetController(t)
	require.NoError(t, err, "Controller creation should succeed")
	require.NotNil(t, controller, "Controller should not be nil")

	// Start the controller
	ctx, cancel := context.WithTimeout(context.Background(), TestTimeoutExtreme)
	defer cancel()
	err = controller.Start(ctx)
	require.NoError(t, err, "Controller start should succeed")

	// Ensure controller is stopped after test
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		controller.Stop(stopCtx)
	}()

	// First create a stream - get available device using optimized helper method
	device, err := helper.GetAvailableCameraDevice(ctx)
	require.NoError(t, err, "Should be able to get available camera device")

	streamName := "test_stream"
	source := "rtsp://localhost:8554/" + device // Use available camera device

	createdStream, err := controller.CreateStream(ctx, streamName, source)
	require.NoError(t, err, "Creating stream should succeed")
	require.NotNil(t, createdStream, "Created stream should not be nil")

	// Get the specific stream
	stream, err := controller.GetStream(ctx, createdStream.Name)
	require.NoError(t, err, "Getting stream should succeed")
	require.NotNil(t, stream, "Stream should not be nil")

	// Verify stream properties
	assert.Equal(t, createdStream.Name, stream.Name, "Stream name should match")
	assert.Equal(t, streamName, stream.Name, "Stream name should match")
	// Note: Path struct doesn't have URL field - source is in Path.Source

	// Clean up - delete the stream
	err = controller.DeleteStream(ctx, createdStream.Name)
	require.NoError(t, err, "Deleting stream should succeed")
}

// TestController_CreateStream_Management_ReqMTX002 tests creating a new stream
func TestController_CreateStream_Management_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create controller
	controller, err := helper.GetController(t)
	require.NoError(t, err, "Controller creation should succeed")
	require.NotNil(t, controller, "Controller should not be nil")

	// Start the controller
	ctx, cancel := context.WithTimeout(context.Background(), TestTimeoutExtreme)
	defer cancel()
	err = controller.Start(ctx)
	require.NoError(t, err, "Controller start should succeed")

	// Ensure controller is stopped after test
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		controller.Stop(stopCtx)
	}()

	// Create a new stream - get available device using optimized helper method
	device, err := helper.GetAvailableCameraDevice(ctx)
	require.NoError(t, err, "Should be able to get available camera device")

	streamName := "test_create_stream"
	source := "rtsp://localhost:8554/" + device // Use available camera device

	stream, err := controller.CreateStream(ctx, streamName, source)
	require.NoError(t, err, "Creating stream should succeed")
	require.NotNil(t, stream, "Created stream should not be nil")

	// Verify stream properties
	assert.NotEmpty(t, stream.Name, "Stream should have a name")
	assert.Equal(t, streamName, stream.Name, "Stream name should match")
	// Note: Path struct doesn't have URL field - source is in Path.Source

	// Clean up - delete the stream
	err = controller.DeleteStream(ctx, stream.Name)
	require.NoError(t, err, "Deleting stream should succeed")
}

// TestController_DeleteStream_Management_ReqMTX002 tests deleting a stream
func TestController_DeleteStream_Management_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create controller
	controller, err := helper.GetController(t)
	require.NoError(t, err, "Controller creation should succeed")
	require.NotNil(t, controller, "Controller should not be nil")

	// Start the controller
	ctx, cancel := context.WithTimeout(context.Background(), TestTimeoutExtreme)
	defer cancel()
	err = controller.Start(ctx)
	require.NoError(t, err, "Controller start should succeed")

	// Ensure controller is stopped after test
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		controller.Stop(stopCtx)
	}()

	// First create a stream - get available device using optimized helper method
	device, err := helper.GetAvailableCameraDevice(ctx)
	require.NoError(t, err, "Should be able to get available camera device")

	streamName := "test_delete_stream"
	source := "rtsp://localhost:8554/" + device // Use available camera device

	stream, err := controller.CreateStream(ctx, streamName, source)
	require.NoError(t, err, "Creating stream should succeed")
	require.NotNil(t, stream, "Created stream should not be nil")

	// Verify stream exists
	_, err = controller.GetStream(ctx, stream.Name)
	require.NoError(t, err, "Stream should exist before deletion")

	// Delete the stream
	err = controller.DeleteStream(ctx, stream.Name)
	require.NoError(t, err, "Deleting stream should succeed")

	// Verify stream no longer exists
	_, err = controller.GetStream(ctx, stream.Name)
	assert.Error(t, err, "Stream should not exist after deletion")
}

// TestController_StreamManagement_ErrorHandling_ReqMTX004 tests error handling for stream management
func TestController_StreamManagement_ErrorHandling_ReqMTX004(t *testing.T) {
	// REQ-MTX-004: Health monitoring and error handling
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create controller
	controller, err := helper.GetController(t)
	require.NoError(t, err, "Controller creation should succeed")
	require.NotNil(t, controller, "Controller should not be nil")

	// Start the controller
	ctx, cancel := context.WithTimeout(context.Background(), TestTimeoutExtreme)
	defer cancel()
	err = controller.Start(ctx)
	require.NoError(t, err, "Controller start should succeed")

	// Ensure controller is stopped after test
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		controller.Stop(stopCtx)
	}()

	// Test getting non-existent stream
	_, err = controller.GetStream(ctx, "non-existent-stream-id")
	assert.Error(t, err, "Getting non-existent stream should fail")

	// Test creating stream with empty name
	_, err = controller.CreateStream(ctx, "", "rtsp://localhost:8554/test_device")
	assert.Error(t, err, "Creating stream with empty name should fail")

	// Test creating stream with empty source
	_, err = controller.CreateStream(ctx, "test_stream", "")
	assert.Error(t, err, "Creating stream with empty source should fail")

	// Test deleting non-existent stream
	err = controller.DeleteStream(ctx, "non-existent-stream-id")
	assert.Error(t, err, "Deleting non-existent stream should fail")
}

// TestController_StreamManagement_NotRunning_ReqMTX004 tests stream management when controller is not running
func TestController_StreamManagement_NotRunning_ReqMTX004(t *testing.T) {
	// REQ-MTX-004: Health monitoring and error handling
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create controller but don't start it
	controller, err := helper.GetController(t)
	require.NoError(t, err, "Controller creation should succeed")
	require.NotNil(t, controller, "Controller should not be nil")

	ctx, cancel := context.WithTimeout(context.Background(), TestTimeoutExtreme)
	defer cancel()

	// Test getting streams when controller is not running
	_, err = controller.GetStreams(ctx)
	assert.Error(t, err, "Getting streams when controller is not running should fail")

	// Test getting specific stream when controller is not running
	_, err = controller.GetStream(ctx, "test-id")
	assert.Error(t, err, "Getting stream when controller is not running should fail")

	// Test creating stream when controller is not running
	_, err = controller.CreateStream(ctx, "test_stream", "rtsp://localhost:8554/test_device")
	assert.Error(t, err, "Creating stream when controller is not running should fail")

	// Test deleting stream when controller is not running
	err = controller.DeleteStream(ctx, "test-id")
	assert.Error(t, err, "Deleting stream when controller is not running should fail")
}

// TestController_StreamManagement_Concurrent_ReqMTX002 tests concurrent stream operations
func TestController_StreamManagement_Concurrent_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create controller
	controller, err := helper.GetController(t)
	require.NoError(t, err, "Controller creation should succeed")
	require.NotNil(t, controller, "Controller should not be nil")

	// Start the controller
	ctx, cancel := context.WithTimeout(context.Background(), TestTimeoutExtreme)
	defer cancel()
	err = controller.Start(ctx)
	require.NoError(t, err, "Controller start should succeed")

	// Ensure controller is stopped after test
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		controller.Stop(stopCtx)
	}()

	// Test concurrent stream creation
	const numStreams = 5
	streamIDs := make([]string, numStreams)
	errors := make([]error, numStreams)

	// Create streams concurrently
	for i := 0; i < numStreams; i++ {
		go func(index int) {
			streamName := fmt.Sprintf("concurrent_stream_%d", index)
			source := fmt.Sprintf("rtsp://localhost:8554/camera%d", index)

			stream, err := controller.CreateStream(ctx, streamName, source)
			if err != nil {
				errors[index] = err
			} else {
				streamIDs[index] = stream.Name
			}
		}(i)
	}

	// Check results
	successCount := 0
	for i := 0; i < numStreams; i++ {
		if errors[i] == nil && streamIDs[i] != "" {
			successCount++
			// Clean up successful streams
			controller.DeleteStream(ctx, streamIDs[i])
		}
	}

	// At least some streams should be created successfully
	assert.Greater(t, successCount, 0, "At least some concurrent stream creations should succeed")
}
