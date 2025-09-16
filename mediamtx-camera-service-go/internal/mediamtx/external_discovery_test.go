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

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
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

	// Test behavior when external discovery is not configured (optional component)
	err = controller.AddExternalStream(ctx, nil)
	assert.Error(t, err, "Adding stream should fail when external discovery not configured")
	assert.Contains(t, err.Error(), "not configured", "Error should indicate external discovery not configured")

	// Test adding stream with empty URL (should also fail due to not configured)
	invalidStream := &ExternalStream{
		Name: "invalid_stream",
		URL:  "", // Empty URL should fail
		Type: "skydio",
	}
	err = controller.AddExternalStream(ctx, invalidStream)
	assert.Error(t, err, "Adding stream should fail when external discovery not configured")
	assert.Contains(t, err.Error(), "not configured", "Error should indicate external discovery not configured")

	// Test removing non-existent stream (should fail due to not configured)
	err = controller.RemoveExternalStream(ctx, "rtsp://nonexistent:6554/stream")
	assert.Error(t, err, "Removing stream should fail when external discovery not configured")
	assert.Contains(t, err.Error(), "not configured", "Error should indicate external discovery not configured")
}

// TestExternalStreamDiscovery_OptionalComponent_ReqMTX004 tests optional component behavior
func TestExternalStreamDiscovery_OptionalComponent_ReqMTX004(t *testing.T) {
	// REQ-MTX-004: Health monitoring and error handling
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create controller (external discovery will be nil by default)
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

	// Test GetExternalStreams returns empty slice when not configured
	streams, err := controller.GetExternalStreams(ctx)
	require.NoError(t, err, "GetExternalStreams should not error when not configured")
	assert.Empty(t, streams, "Should return empty slice when external discovery not configured")

	// Test DiscoverExternalStreams returns error when not configured
	options := DiscoveryOptions{
		SkydioEnabled:  true,
		GenericEnabled: false,
	}
	result, err := controller.DiscoverExternalStreams(ctx, options)
	assert.Error(t, err, "DiscoverExternalStreams should error when not configured")
	assert.Nil(t, result, "Result should be nil when external discovery not configured")
	assert.Contains(t, err.Error(), "not configured", "Error should indicate external discovery not configured")

	// Test AddExternalStream returns error when not configured
	stream := &ExternalStream{
		Name: "test_stream",
		URL:  "rtsp://192.168.42.10:6554/infrared",
		Type: "skydio",
	}
	err = controller.AddExternalStream(ctx, stream)
	assert.Error(t, err, "AddExternalStream should error when not configured")
	assert.Contains(t, err.Error(), "not configured", "Error should indicate external discovery not configured")

	// Test RemoveExternalStream returns error when not configured
	err = controller.RemoveExternalStream(ctx, "rtsp://192.168.42.10:6554/infrared")
	assert.Error(t, err, "RemoveExternalStream should error when not configured")
	assert.Contains(t, err.Error(), "not configured", "Error should indicate external discovery not configured")
}

// TestExternalStreamDiscovery_ContextAwareShutdown tests the context-aware shutdown functionality
func TestExternalStreamDiscovery_ContextAwareShutdown(t *testing.T) {
	t.Run("graceful_shutdown_with_context", func(t *testing.T) {
		helper := NewMediaMTXTestHelper(t, nil)
		defer helper.Cleanup(t)

		// Create external discovery directly
		config := &config.ExternalDiscoveryConfig{
			Enabled:      true,
			ScanInterval: 30,
		}
		logger := helper.GetLogger()
		discovery := NewExternalStreamDiscovery(config, logger)

		// Start discovery
		ctx := context.Background()
		err := discovery.Start(ctx)
		require.NoError(t, err, "Discovery should start successfully")

		// Test graceful shutdown with context
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		start := time.Now()
		err = discovery.Stop(shutdownCtx)
		elapsed := time.Since(start)

		require.NoError(t, err, "Discovery should stop gracefully")
		assert.Less(t, elapsed, 1*time.Second, "Shutdown should be fast")
	})

	t.Run("shutdown_with_cancelled_context", func(t *testing.T) {
		helper := NewMediaMTXTestHelper(t, nil)
		defer helper.Cleanup(t)

		// Create external discovery directly
		config := &config.ExternalDiscoveryConfig{
			Enabled:      true,
			ScanInterval: 30,
		}
		logger := helper.GetLogger()
		discovery := NewExternalStreamDiscovery(config, logger)

		// Start discovery
		ctx := context.Background()
		err := discovery.Start(ctx)
		require.NoError(t, err, "Discovery should start successfully")

		// Cancel context immediately
		shutdownCtx, cancel := context.WithCancel(context.Background())
		cancel()

		// Stop should complete quickly since context is already cancelled
		start := time.Now()
		err = discovery.Stop(shutdownCtx)
		elapsed := time.Since(start)

		require.NoError(t, err, "Discovery should stop even with cancelled context")
		assert.Less(t, elapsed, 100*time.Millisecond, "Shutdown should be very fast with cancelled context")
	})

	t.Run("shutdown_timeout_handling", func(t *testing.T) {
		helper := NewMediaMTXTestHelper(t, nil)
		defer helper.Cleanup(t)

		// Create external discovery directly
		config := &config.ExternalDiscoveryConfig{
			Enabled:      true,
			ScanInterval: 30,
		}
		logger := helper.GetLogger()
		discovery := NewExternalStreamDiscovery(config, logger)

		// Start discovery
		ctx := context.Background()
		err := discovery.Start(ctx)
		require.NoError(t, err, "Discovery should start successfully")

		// Use very short timeout to test timeout handling
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		// Give context time to expire using proper synchronization
		select {
		case <-time.After(2 * time.Millisecond):
			// Context should be expired now
		case <-ctx.Done():
			// Context already cancelled, continue
		}

		start := time.Now()
		err = discovery.Stop(shutdownCtx)
		elapsed := time.Since(start)

		// Should timeout but not hang
		require.Error(t, err, "Should timeout with very short timeout")
		assert.Contains(t, err.Error(), "context deadline exceeded", "Error should indicate timeout")
		assert.Less(t, elapsed, 1*time.Second, "Should not hang indefinitely")
	})

	t.Run("double_stop_handling", func(t *testing.T) {
		helper := NewMediaMTXTestHelper(t, nil)
		defer helper.Cleanup(t)

		// Create external discovery directly
		config := &config.ExternalDiscoveryConfig{
			Enabled:      true,
			ScanInterval: 30,
		}
		logger := helper.GetLogger()
		discovery := NewExternalStreamDiscovery(config, logger)

		// Start discovery
		ctx := context.Background()
		err := discovery.Start(ctx)
		require.NoError(t, err, "Discovery should start successfully")

		// Stop first time
		ctx1, cancel1 := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel1()
		err = discovery.Stop(ctx1)
		require.NoError(t, err, "First stop should succeed")

		// Stop second time should not error
		ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel2()
		err = discovery.Stop(ctx2)
		assert.NoError(t, err, "Second stop should not error")
	})

	t.Run("stop_without_start", func(t *testing.T) {
		helper := NewMediaMTXTestHelper(t, nil)
		defer helper.Cleanup(t)

		// Create external discovery directly
		config := &config.ExternalDiscoveryConfig{
			Enabled:      true,
			ScanInterval: 30,
		}
		logger := helper.GetLogger()
		discovery := NewExternalStreamDiscovery(config, logger)

		// Stop without starting should not error
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := discovery.Stop(ctx)
		assert.NoError(t, err, "Stop without start should not error")
	})
}
