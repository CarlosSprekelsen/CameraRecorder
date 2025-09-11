/*
MediaMTX External Stream Discovery Tests - Real Server Integration

Requirements Coverage:
- REQ-MTX-001: MediaMTX service integration
- REQ-MTX-002: Stream management capabilities
- REQ-MTX-003: Path creation and deletion
- REQ-MTX-004: Health monitoring

Test Categories: Unit (using real MediaMTX server)
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

// TestExternalStreamDiscovery_DiscoverExternalStreams_ReqMTX001 tests external stream discovery
func TestExternalStreamDiscovery_DiscoverExternalStreams_ReqMTX001(t *testing.T) {
	// REQ-MTX-001: MediaMTX service integration
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create controller
	controller, err := helper.GetController(t)
	require.NoError(t, err, "Controller creation should succeed")
	require.NotNil(t, controller, "Controller should not be nil")

	// Start the controller
	ctx := context.Background()
	err = controller.Start(ctx)
	require.NoError(t, err, "Controller start should succeed")

	// Ensure controller is stopped after test
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		controller.Stop(stopCtx)
	}()

	// Test external stream discovery with Skydio configuration
	options := DiscoveryOptions{
		SkydioEnabled:  true,
		GenericEnabled: false,
	}

	result, err := controller.DiscoverExternalStreams(ctx, options)
	require.NoError(t, err, "External stream discovery should succeed")
	require.NotNil(t, result, "Discovery result should not be nil")

	// Verify result structure
	assert.NotNil(t, result.DiscoveredStreams, "Discovered streams should not be nil")
	assert.GreaterOrEqual(t, result.TotalFound, 0, "Total found should be non-negative")
	assert.GreaterOrEqual(t, result.ScanDuration, time.Duration(0), "Scan duration should be non-negative")
}

// TestExternalStreamDiscovery_AddExternalStream_ReqMTX002 tests adding external streams
func TestExternalStreamDiscovery_AddExternalStream_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create controller
	controller, err := helper.GetController(t)
	require.NoError(t, err, "Controller creation should succeed")
	require.NotNil(t, controller, "Controller should not be nil")

	// Start the controller
	ctx := context.Background()
	err = controller.Start(ctx)
	require.NoError(t, err, "Controller start should succeed")

	// Ensure controller is stopped after test
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		controller.Stop(stopCtx)
	}()

	// Test adding external stream
	stream := &ExternalStream{
		Name: "test_external_stream",
		URL:  "rtsp://192.168.42.10:6554/infrared",
		Type: "skydio",
	}

	err = controller.AddExternalStream(ctx, stream)
	require.NoError(t, err, "Adding external stream should succeed")

	// Verify stream was added
	streams, err := controller.GetExternalStreams(ctx)
	require.NoError(t, err, "Getting external streams should succeed")
	require.NotNil(t, streams, "Streams should not be nil")

	// Find our added stream
	found := false
	for _, s := range streams {
		if s.Name == stream.Name {
			found = true
			assert.Equal(t, stream.URL, s.URL, "Stream URL should match")
			assert.Equal(t, stream.Type, s.Type, "Stream type should match")
			break
		}
	}
	assert.True(t, found, "Added stream should be found in list")
}

// TestExternalStreamDiscovery_RemoveExternalStream_ReqMTX002 tests removing external streams
func TestExternalStreamDiscovery_RemoveExternalStream_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create controller
	controller, err := helper.GetController(t)
	require.NoError(t, err, "Controller creation should succeed")
	require.NotNil(t, controller, "Controller should not be nil")

	// Start the controller
	ctx := context.Background()
	err = controller.Start(ctx)
	require.NoError(t, err, "Controller start should succeed")

	// Ensure controller is stopped after test
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		controller.Stop(stopCtx)
	}()

	// First add a stream
	stream := &ExternalStream{
		Name: "test_removal_stream",
		URL:  "rtsp://192.168.42.11:6554/infrared",
		Type: "skydio",
	}

	err = controller.AddExternalStream(ctx, stream)
	require.NoError(t, err, "Adding external stream should succeed")

	// Verify stream was added
	streams, err := controller.GetExternalStreams(ctx)
	require.NoError(t, err, "Getting external streams should succeed")
	require.NotNil(t, streams, "Streams should not be nil")

	// Find our added stream
	found := false
	for _, s := range streams {
		if s.Name == stream.Name {
			found = true
			break
		}
	}
	assert.True(t, found, "Added stream should be found before removal")

	// Remove the stream
	err = controller.RemoveExternalStream(ctx, stream.URL)
	require.NoError(t, err, "Removing external stream should succeed")

	// Verify stream was removed
	streams, err = controller.GetExternalStreams(ctx)
	require.NoError(t, err, "Getting external streams should succeed")
	require.NotNil(t, streams, "Streams should not be nil")

	// Verify stream is no longer in the list
	found = false
	for _, s := range streams {
		if s.Name == stream.Name {
			found = true
			break
		}
	}
	assert.False(t, found, "Removed stream should not be found in list")
}

// TestExternalStreamDiscovery_GetExternalStreams_ReqMTX002 tests getting external streams
func TestExternalStreamDiscovery_GetExternalStreams_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create controller
	controller, err := helper.GetController(t)
	require.NoError(t, err, "Controller creation should succeed")
	require.NotNil(t, controller, "Controller should not be nil")

	// Start the controller
	ctx := context.Background()
	err = controller.Start(ctx)
	require.NoError(t, err, "Controller start should succeed")

	// Ensure controller is stopped after test
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		controller.Stop(stopCtx)
	}()

	// Get external streams (should be empty initially)
	streams, err := controller.GetExternalStreams(ctx)
	require.NoError(t, err, "Getting external streams should succeed")
	require.NotNil(t, streams, "Streams should not be nil")
	assert.IsType(t, []*ExternalStream{}, streams, "Should return slice of ExternalStream")

	// Add a test stream
	stream := &ExternalStream{
		Name: "test_list_stream",
		URL:  "rtsp://192.168.42.12:6554/infrared",
		Type: "skydio",
	}

	err = controller.AddExternalStream(ctx, stream)
	require.NoError(t, err, "Adding external stream should succeed")

	// Get streams again and verify our stream is there
	streams, err = controller.GetExternalStreams(ctx)
	require.NoError(t, err, "Getting external streams should succeed")
	require.NotNil(t, streams, "Streams should not be nil")
	assert.GreaterOrEqual(t, len(streams), 1, "Should have at least one stream")

	// Verify stream properties
	found := false
	for _, s := range streams {
		if s.Name == stream.Name {
			found = true
			assert.Equal(t, stream.URL, s.URL, "Stream URL should match")
			assert.Equal(t, stream.Type, s.Type, "Stream type should match")
			break
		}
	}
	assert.True(t, found, "Added stream should be found in list")
}

// TestExternalStreamDiscovery_ErrorHandling_ReqMTX004 tests error handling scenarios
func TestExternalStreamDiscovery_ErrorHandling_ReqMTX004(t *testing.T) {
	// REQ-MTX-004: Health monitoring and error handling
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create controller
	controller, err := helper.GetController(t)
	require.NoError(t, err, "Controller creation should succeed")
	require.NotNil(t, controller, "Controller should not be nil")

	// Start the controller
	ctx := context.Background()
	err = controller.Start(ctx)
	require.NoError(t, err, "Controller start should succeed")

	// Ensure controller is stopped after test
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		controller.Stop(stopCtx)
	}()

	// Test adding invalid stream (nil stream)
	err = controller.AddExternalStream(ctx, nil)
	assert.Error(t, err, "Adding nil stream should fail")

	// Test adding stream with empty URL
	invalidStream := &ExternalStream{
		Name: "invalid_stream",
		URL:  "", // Empty URL should fail
		Type: "skydio",
	}
	err = controller.AddExternalStream(ctx, invalidStream)
	assert.Error(t, err, "Adding stream with empty URL should fail")

	// Test removing non-existent stream
	err = controller.RemoveExternalStream(ctx, "rtsp://nonexistent:6554/stream")
	assert.Error(t, err, "Removing non-existent stream should fail")
}
