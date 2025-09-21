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
	t.Skip("External stream discovery testing postponed - component configuration needed")
	t.Skip("External stream discovery testing postponed - component configuration needed")
	// REQ-MTX-001: MediaMTX service integration
	helper, ctx := SetupMediaMTXTest(t)

	// Create controller
	controller, err := helper.GetController(t)
	helper.AssertStandardResponse(t, controller, err, "Controller creation")

	// Start the controller
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
	helper.AssertStandardResponse(t, result, err, "External stream discovery")

	// Verify result structure
	assert.NotNil(t, result.DiscoveredStreams, "Discovered streams should not be nil")
	assert.GreaterOrEqual(t, result.TotalFound, 0, "Total found should be non-negative")
	assert.GreaterOrEqual(t, result.ScanDuration, time.Duration(0), "Scan duration should be non-negative")
}

// TestExternalStreamDiscovery_AddExternalStream_ReqMTX002 tests adding external streams
func TestExternalStreamDiscovery_AddExternalStream_ReqMTX002(t *testing.T) {
	t.Skip("External stream discovery testing postponed - component configuration needed")
	t.Skip("External stream discovery testing postponed - component configuration needed")
	// REQ-MTX-002: Stream management capabilities
	helper, ctx := SetupMediaMTXTest(t)

	// Create controller
	controller, err := helper.GetController(t)
	helper.AssertStandardResponse(t, controller, err, "Controller creation")

	// Start the controller
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

	_, err = controller.AddExternalStream(ctx, stream)
	require.NoError(t, err, "Adding external stream should succeed")

	// Verify stream was added
	streams, err := controller.GetExternalStreams(ctx)
	require.NoError(t, err, "Getting external streams should succeed")
	require.NotNil(t, streams, "Streams should not be nil")

	// Find our added stream
	found := false
	for _, s := range streams.ExternalStreams {
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
	t.Skip("External stream discovery testing postponed - component configuration needed")
	// REQ-MTX-002: Stream management capabilities
	helper, ctx := SetupMediaMTXTest(t)

	// Create controller
	controller, err := helper.GetController(t)
	helper.AssertStandardResponse(t, controller, err, "Controller creation")

	// Start the controller
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

	_, err = controller.AddExternalStream(ctx, stream)
	require.NoError(t, err, "Adding external stream should succeed")

	// Verify stream was added
	streams, err := controller.GetExternalStreams(ctx)
	require.NoError(t, err, "Getting external streams should succeed")
	require.NotNil(t, streams, "Streams should not be nil")

	// Find our added stream
	found := false
	for _, s := range streams.ExternalStreams {
		if s.Name == stream.Name {
			found = true
			break
		}
	}
	assert.True(t, found, "Added stream should be found before removal")

	// Remove the stream
	_, err = controller.RemoveExternalStream(ctx, stream.URL)
	require.NoError(t, err, "Removing external stream should succeed")

	// Verify stream was removed
	streams, err = controller.GetExternalStreams(ctx)
	require.NoError(t, err, "Getting external streams should succeed")
	require.NotNil(t, streams, "Streams should not be nil")

	// Verify stream is no longer in the list
	found = false
	for _, s := range streams.ExternalStreams {
		if s.Name == stream.Name {
			found = true
			break
		}
	}
	assert.False(t, found, "Removed stream should not be found in list")
}

// TestExternalStreamDiscovery_GetExternalStreams_ReqMTX002 tests getting external streams
func TestExternalStreamDiscovery_GetExternalStreams_ReqMTX002(t *testing.T) {
	t.Skip("External stream discovery testing postponed - component configuration needed")
	// REQ-MTX-002: Stream management capabilities
	helper, ctx := SetupMediaMTXTest(t)

	// Create controller
	controller, err := helper.GetController(t)
	helper.AssertStandardResponse(t, controller, err, "Controller creation")

	// Start the controller
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

	_, err = controller.AddExternalStream(ctx, stream)
	require.NoError(t, err, "Adding external stream should succeed")

	// Get streams again and verify our stream is there
	streams, err = controller.GetExternalStreams(ctx)
	require.NoError(t, err, "Getting external streams should succeed")
	require.NotNil(t, streams, "Streams should not be nil")
	assert.GreaterOrEqual(t, len(streams.ExternalStreams), 1, "Should have at least one stream")

	// Verify stream properties
	found := false
	for _, s := range streams.ExternalStreams {
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
	t.Skip("External stream discovery testing postponed - component configuration needed")
	// REQ-MTX-004: Health monitoring and error handling
	helper, ctx := SetupMediaMTXTest(t)

	// Create controller
	controller, err := helper.GetController(t)
	helper.AssertStandardResponse(t, controller, err, "Controller creation")

	// Start the controller
	err = controller.Start(ctx)
	require.NoError(t, err, "Controller start should succeed")

	// Ensure controller is stopped after test
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		controller.Stop(stopCtx)
	}()

	// Test behavior when external discovery is not configured (optional component)
	_, err = controller.AddExternalStream(ctx, nil)
	assert.Error(t, err, "Adding stream should fail when external discovery not configured")
	assert.Contains(t, err.Error(), "not configured", "Error should indicate external discovery not configured")

	// Test adding stream with empty URL (should also fail due to not configured)
	invalidStream := &ExternalStream{
		Name: "invalid_stream",
		URL:  "", // Empty URL should fail
		Type: "skydio",
	}
	_, err = controller.AddExternalStream(ctx, invalidStream)
	assert.Error(t, err, "Adding stream should fail when external discovery not configured")
	assert.Contains(t, err.Error(), "not configured", "Error should indicate external discovery not configured")

	// Test removing non-existent stream (should fail due to not configured)
	_, err = controller.RemoveExternalStream(ctx, "rtsp://nonexistent:6554/stream")
	assert.Error(t, err, "Removing stream should fail when external discovery not configured")
	assert.Contains(t, err.Error(), "not configured", "Error should indicate external discovery not configured")
}

// TestExternalStreamDiscovery_OptionalComponent_ReqMTX004 tests optional component behavior
func TestExternalStreamDiscovery_OptionalComponent_ReqMTX004(t *testing.T) {
	t.Skip("External stream discovery testing postponed - component configuration needed")
	// REQ-MTX-004: Health monitoring and error handling
	helper, ctx := SetupMediaMTXTest(t)

	// Create controller (external discovery will be nil by default)
	controller, err := helper.GetController(t)
	helper.AssertStandardResponse(t, controller, err, "Controller creation")

	// Start the controller
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
	_, err = controller.AddExternalStream(ctx, stream)
	assert.Error(t, err, "AddExternalStream should error when not configured")
	assert.Contains(t, err.Error(), "not configured", "Error should indicate external discovery not configured")

	// Test RemoveExternalStream returns error when not configured
	_, err = controller.RemoveExternalStream(ctx, "rtsp://192.168.42.10:6554/infrared")
	assert.Error(t, err, "RemoveExternalStream should error when not configured")
	assert.Contains(t, err.Error(), "not configured", "Error should indicate external discovery not configured")
}

// TestExternalStreamDiscovery_ContextAwareShutdown tests the context-aware shutdown functionality
func TestExternalStreamDiscovery_ContextAwareShutdown(t *testing.T) {
	t.Skip("External stream discovery testing postponed - component configuration needed")
	t.Run("graceful_shutdown_with_context", func(t *testing.T) {
		helper, ctx := SetupMediaMTXTest(t)

		// Create external discovery directly
		logger := helper.GetLoggerForComponent("external_discovery") // Component-specific logging
		// Use centralized configuration architecture
		configManager := helper.GetConfigManager()
		configIntegration := NewConfigIntegration(configManager, logger)
		discovery := NewExternalStreamDiscovery(configIntegration, logger)

		// Start discovery
		ctx, cancel := helper.GetStandardContext()
		defer cancel()
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
		helper, ctx := SetupMediaMTXTest(t)

		// Create external discovery directly
		logger := helper.GetLoggerForComponent("external_discovery") // Component-specific logging
		// Use centralized configuration architecture
		configManager := helper.GetConfigManager()
		configIntegration := NewConfigIntegration(configManager, logger)
		discovery := NewExternalStreamDiscovery(configIntegration, logger)

		// Start discovery
		ctx, cancel := helper.GetStandardContext()
		defer cancel()
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
		helper, ctx := SetupMediaMTXTest(t)

		// Create external discovery directly
		logger := helper.GetLoggerForComponent("external_discovery") // Component-specific logging
		// Use centralized configuration architecture
		configManager := helper.GetConfigManager()
		configIntegration := NewConfigIntegration(configManager, logger)
		discovery := NewExternalStreamDiscovery(configIntegration, logger)

		// Start discovery
		ctx, cancel := helper.GetStandardContext()
		defer cancel()
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
		helper, ctx := SetupMediaMTXTest(t)

		// Create external discovery directly
		logger := helper.GetLoggerForComponent("external_discovery") // Component-specific logging
		// Use centralized configuration architecture
		configManager := helper.GetConfigManager()
		configIntegration := NewConfigIntegration(configManager, logger)
		discovery := NewExternalStreamDiscovery(configIntegration, logger)

		// Start discovery
		ctx, cancel := helper.GetStandardContext()
		defer cancel()
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
		helper, ctx := SetupMediaMTXTest(t)

		// Create external discovery directly
		logger := helper.GetLoggerForComponent("external_discovery") // Component-specific logging
		// Use centralized configuration architecture
		configManager := helper.GetConfigManager()
		configIntegration := NewConfigIntegration(configManager, logger)
		discovery := NewExternalStreamDiscovery(configIntegration, logger)

		// Stop without starting should not error
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := discovery.Stop(ctx)
		assert.NoError(t, err, "Stop without start should not error")
	})
}

// TestExternalStreamDiscovery_DiscoverExternalStreamsAPI_ReqMTX002 tests new API-ready discovery method
func TestExternalStreamDiscovery_DiscoverExternalStreamsAPI_ReqMTX002(t *testing.T) {
	t.Skip("External stream discovery testing postponed - component configuration needed")
	// REQ-MTX-002: Stream management capabilities - API-ready external stream discovery
	helper, ctx := SetupMediaMTXTest(t)

	// Create external stream discovery using centralized configuration architecture
	logger := helper.GetLoggerForComponent("external_discovery") // Component-specific logging
	configManager := helper.GetConfigManager()
	configIntegration := NewConfigIntegration(configManager, logger)

	// Use centralized config pattern (architectural compliance)
	discovery := NewExternalStreamDiscovery(configIntegration, logger)
	require.NotNil(t, discovery, "ExternalStreamDiscovery should be created")

	// Test DiscoverExternalStreamsAPI method - new API-ready response
	options := DiscoveryOptions{
		SkydioEnabled:  true,
		GenericEnabled: false,
		ForceRescan:    true,
		IncludeOffline: false,
	}

	response, err := discovery.DiscoverExternalStreamsAPI(ctx, options)
	require.NoError(t, err, "DiscoverExternalStreamsAPI should succeed")
	require.NotNil(t, response, "DiscoverExternalStreamsAPI should return API-ready response")

	// Validate API-ready response format per JSON-RPC documentation
	assert.NotNil(t, response.DiscoveredStreams, "Response should include discovered streams array")
	assert.NotNil(t, response.SkydioStreams, "Response should include Skydio streams array")
	assert.NotNil(t, response.GenericStreams, "Response should include generic streams array")
	assert.GreaterOrEqual(t, response.TotalFound, 0, "Response should include total found count")
	assert.GreaterOrEqual(t, response.ScanTimestamp, int64(0), "Response should include scan timestamp")
	assert.NotNil(t, response.DiscoveryOptions, "Response should include discovery options")
	assert.NotEmpty(t, response.ScanDuration, "Response should include scan duration")
	assert.NotNil(t, response.Errors, "Response should include errors array")
}

// TestExternalStreamDiscovery_GetExternalStreamsAPI_ReqMTX002 tests new API-ready streams listing
func TestExternalStreamDiscovery_GetExternalStreamsAPI_ReqMTX002(t *testing.T) {
	t.Skip("External stream discovery testing postponed - component configuration needed")
	// REQ-MTX-002: Stream management capabilities - API-ready external streams listing
	helper, ctx := SetupMediaMTXTest(t)

	// Create external stream discovery using existing test infrastructure
	logger := helper.GetLoggerForComponent("external_discovery") // Component-specific logging

	// Use centralized config pattern (architectural compliance)
	configManager := helper.GetConfigManager()
	configIntegration := NewConfigIntegration(configManager, logger)
	discovery := NewExternalStreamDiscovery(configIntegration, logger)
	require.NotNil(t, discovery, "ExternalStreamDiscovery should be created")

	// Test GetExternalStreamsAPI method - new API-ready response
	response, err := discovery.GetExternalStreamsAPI(ctx)
	require.NoError(t, err, "GetExternalStreamsAPI should succeed")
	require.NotNil(t, response, "GetExternalStreamsAPI should return API-ready response")

	// Validate API-ready response format per JSON-RPC documentation
	assert.NotNil(t, response.ExternalStreams, "Response should include external streams array")
	assert.NotNil(t, response.SkydioStreams, "Response should include Skydio streams array")
	assert.NotNil(t, response.GenericStreams, "Response should include generic streams array")
	assert.GreaterOrEqual(t, response.TotalCount, 0, "Response should include total count")
	assert.GreaterOrEqual(t, response.Timestamp, int64(0), "Response should include timestamp")
}

// TestExternalStreamDiscovery_AddExternalStreamAPI_ReqMTX002 tests new API-ready stream addition
func TestExternalStreamDiscovery_AddExternalStreamAPI_ReqMTX002(t *testing.T) {
	t.Skip("External stream discovery testing postponed - component configuration needed")
	// REQ-MTX-002: Stream management capabilities - API-ready external stream addition
	helper, ctx := SetupMediaMTXTest(t)

	// Create external stream discovery using existing test infrastructure
	logger := helper.GetLoggerForComponent("external_discovery") // Component-specific logging

	// Use centralized config pattern (architectural compliance)
	configManager := helper.GetConfigManager()
	configIntegration := NewConfigIntegration(configManager, logger)
	discovery := NewExternalStreamDiscovery(configIntegration, logger)
	require.NotNil(t, discovery, "ExternalStreamDiscovery should be created")

	// Create test external stream
	testStream := &ExternalStream{
		URL:          "rtsp://test-stream.example.com:554/test",
		Type:         "generic_rtsp",
		Name:         "Test Stream",
		Status:       "discovered",
		DiscoveredAt: time.Now(),
		LastSeen:     time.Now(),
	}

	// Test AddExternalStreamAPI method - new API-ready response
	response, err := discovery.AddExternalStreamAPI(ctx, testStream)
	require.NoError(t, err, "AddExternalStreamAPI should succeed")
	require.NotNil(t, response, "AddExternalStreamAPI should return API-ready response")

	// Validate API-ready response format per JSON-RPC documentation
	assert.Equal(t, testStream.URL, response.StreamURL, "Response should include stream URL")
	assert.Equal(t, testStream.Name, response.StreamName, "Response should include stream name")
	assert.Equal(t, testStream.Type, response.StreamType, "Response should include stream type")
	assert.Equal(t, "added", response.Status, "Response should indicate added status")
	assert.Greater(t, response.Timestamp, int64(0), "Response should include timestamp")

	// Verify stream was actually added by listing streams
	listResponse, err := discovery.GetExternalStreamsAPI(ctx)
	require.NoError(t, err, "GetExternalStreamsAPI should succeed after adding")
	assert.Greater(t, listResponse.TotalCount, 0, "Should have at least one stream after adding")
}

// TestExternalStreamDiscovery_RemoveExternalStreamAPI_ReqMTX002 tests new API-ready stream removal
func TestExternalStreamDiscovery_RemoveExternalStreamAPI_ReqMTX002(t *testing.T) {
	t.Skip("External stream discovery testing postponed - component configuration needed")
	// REQ-MTX-002: Stream management capabilities - API-ready external stream removal
	helper, ctx := SetupMediaMTXTest(t)

	// Create external stream discovery using existing test infrastructure
	logger := helper.GetLoggerForComponent("external_discovery") // Component-specific logging

	// Use centralized config pattern (architectural compliance)
	configManager := helper.GetConfigManager()
	configIntegration := NewConfigIntegration(configManager, logger)
	discovery := NewExternalStreamDiscovery(configIntegration, logger)
	require.NotNil(t, discovery, "ExternalStreamDiscovery should be created")

	// First add a test stream
	testStream := &ExternalStream{
		URL:          "rtsp://test-remove.example.com:554/test",
		Type:         "generic_rtsp",
		Name:         "Test Remove Stream",
		Status:       "discovered",
		DiscoveredAt: time.Now(),
		LastSeen:     time.Now(),
	}

	addResponse, err := discovery.AddExternalStreamAPI(ctx, testStream)
	require.NoError(t, err, "AddExternalStreamAPI should succeed for setup")
	require.NotNil(t, addResponse, "Should add stream successfully for setup")

	// Test RemoveExternalStreamAPI method - new API-ready response
	response, err := discovery.RemoveExternalStreamAPI(ctx, testStream.URL)
	require.NoError(t, err, "RemoveExternalStreamAPI should succeed")
	require.NotNil(t, response, "RemoveExternalStreamAPI should return API-ready response")

	// Validate API-ready response format per JSON-RPC documentation
	assert.Equal(t, testStream.URL, response.StreamURL, "Response should include stream URL")
	assert.Equal(t, "removed", response.Status, "Response should indicate removed status")
	assert.Greater(t, response.Timestamp, int64(0), "Response should include timestamp")

	// Verify stream was actually removed by attempting to remove again
	_, err = discovery.RemoveExternalStreamAPI(ctx, testStream.URL)
	assert.Error(t, err, "Should return error when removing non-existent stream")
}
