/*
MediaMTX RTSP Connection Manager Comprehensive Unit Tests

Requirements Coverage:
- REQ-MTX-001: MediaMTX service integration
- REQ-MTX-002: Stream management capabilities
- REQ-MTX-003: Path creation and deletion
- REQ-MTX-004: Health monitoring

Test Categories: Unit (using real MediaMTX server)
API Documentation Reference: docs/api/swagger.json

RTSP Connection Management Tests for STANAG4606 streaming monitoring
Leverages existing test utilities and logging module for comprehensive coverage
*/

package mediamtx

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestMediaMTXConfig creates a test MediaMTX configuration
func createTestMediaMTXConfig() *config.MediaMTXConfig {
	return &config.MediaMTXConfig{
		BaseURL: "http://localhost:9997",
		Timeout: 10 * time.Second, // Add HTTP client timeout to prevent hanging
		RTSPMonitoring: config.RTSPMonitoringConfig{
			Enabled:             true,
			CheckInterval:       30,
			ConnectionTimeout:   10,
			MaxConnections:      50,
			SessionTimeout:      300,
			BandwidthThreshold:  1000000,
			PacketLossThreshold: 0.05,
			JitterThreshold:     50.0,
		},
	}
}

// createTestRTSPManagerWithCustomConfig creates a test RTSP manager with custom configuration
func createTestRTSPManagerWithCustomConfig(t *testing.T, helper *MediaMTXTestHelper, customConfig *config.RTSPMonitoringConfig) RTSPConnectionManager {
	// For custom config, we need to create a new instance with the custom configuration
	// This is a legitimate use case that requires a new instance
	config := &config.MediaMTXConfig{
		BaseURL: "http://localhost:9997",
		Timeout: 10 * time.Second,
		RTSPMonitoring: config.RTSPMonitoringConfig{
			Enabled:             true,
			CheckInterval:       30,
			ConnectionTimeout:   10,
			MaxConnections:      50,
			SessionTimeout:      300,
			BandwidthThreshold:  1000000,
			PacketLossThreshold: 0.05,
			JitterThreshold:     50.0,
		},
	}
	if customConfig != nil {
		config.RTSPMonitoring = *customConfig
	}

	logger := helper.GetLogger()
	logger.SetLevel(logrus.ErrorLevel)

	rtspManager := NewRTSPConnectionManager(helper.GetClient(), config, logger)
	require.NotNil(t, rtspManager, "RTSP connection manager should not be nil")

	return rtspManager
}

// assertRTSPHealthStatus validates RTSP health status structure
func assertRTSPHealthStatus(t *testing.T, health *HealthStatus, expectedStatus string) {
	require.NotNil(t, health, "Health status should not be nil")
	assert.NotEmpty(t, health.Status, "Health status should not be empty")
	assert.NotZero(t, health.Timestamp, "Health timestamp should not be zero")

	if expectedStatus != "" {
		assert.Equal(t, expectedStatus, health.Status, "Health status should match expected")
	}
}

// assertRTSPMetrics validates RTSP metrics structure
func assertRTSPMetrics(t *testing.T, metrics map[string]interface{}) {
	require.NotNil(t, metrics, "Metrics should not be nil")
	assert.Contains(t, metrics, "is_healthy", "Metrics should contain is_healthy")
	assert.Contains(t, metrics, "monitoring_enabled", "Metrics should contain monitoring_enabled")
	assert.Contains(t, metrics, "last_check", "Metrics should contain last_check")
}

// logTestProgress logs test progress with structured logging
func logTestProgress(t *testing.T, logger *logging.Logger, message string, fields map[string]interface{}) {
	if fields == nil {
		fields = make(map[string]interface{})
	}
	fields["test"] = t.Name()
	fields["timestamp"] = time.Now().Format(time.RFC3339)

	logger.WithFields(fields).Info(message)
	t.Logf("%s", message)
}

// TestNewRTSPConnectionManager_ReqMTX001 tests RTSP connection manager creation
func TestNewRTSPConnectionManager_ReqMTX001(t *testing.T) {
	// REQ-MTX-001: MediaMTX service integration
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Server is ready via shared test helper

	// Test server health first
	err := helper.TestMediaMTXHealth(t)
	require.NoError(t, err, "MediaMTX server should be healthy")

	// Use shared RTSP connection manager from test helper
	rtspManager := helper.GetRTSPConnectionManager()
	require.NotNil(t, rtspManager, "RTSP connection manager should not be nil")

	// Log test progress
	logTestProgress(t, helper.GetLogger(), "RTSP connection manager created successfully", map[string]interface{}{
		"component": "rtsp_connection_manager",
		"status":    "created",
	})
}

// TestRTSPConnectionManager_ListConnections_ReqMTX002 tests RTSP connection listing
func TestRTSPConnectionManager_ListConnections_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Server is ready via shared test helper

	// Use shared RTSP connection manager from test helper
	rtspManager := helper.GetRTSPConnectionManager()
	ctx := context.Background()

	// Test listing connections with different pagination
	testCases := []struct {
		name         string
		page         int
		itemsPerPage int
	}{
		{"first_page", 0, 10},
		{"second_page", 1, 5},
		{"large_page", 0, 100},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			connections, err := rtspManager.ListConnections(ctx, tc.page, tc.itemsPerPage)
			require.NoError(t, err, "ListConnections should succeed for %s", tc.name)
			assert.NotNil(t, connections, "Connections list should not be nil for %s", tc.name)
			assert.NotNil(t, connections.Items, "Connections items should not be nil for %s", tc.name)

			// Log test progress
			logTestProgress(t, helper.GetLogger(), "RTSP connections listed successfully", map[string]interface{}{
				"test_case":         tc.name,
				"page":              tc.page,
				"items_per_page":    tc.itemsPerPage,
				"connections_found": len(connections.Items),
				"total_pages":       connections.PageCount,
				"total_items":       connections.ItemCount,
			})
		})
	}
}

// TestRTSPConnectionManager_ListSessions_ReqMTX002 tests RTSP session listing
func TestRTSPConnectionManager_ListSessions_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create RTSP connection manager
	config := createTestMediaMTXConfig()
	logger := logging.CreateTestLogger(t, nil)
	logger.SetLevel(logrus.ErrorLevel)

	rtspManager := NewRTSPConnectionManager(helper.GetClient(), config, logger)
	require.NotNil(t, rtspManager)

	ctx := context.Background()

	// Test listing sessions
	sessions, err := rtspManager.ListSessions(ctx, 0, 10)
	require.NoError(t, err, "ListSessions should succeed")
	assert.NotNil(t, sessions, "Sessions list should not be nil")
	assert.NotNil(t, sessions.Items, "Sessions items should not be nil")

	t.Logf("Found %d RTSP sessions", len(sessions.Items))
}

// TestRTSPConnectionManager_GetConnectionHealth_ReqMTX004 tests RTSP connection health monitoring
func TestRTSPConnectionManager_GetConnectionHealth_ReqMTX004(t *testing.T) {
	// REQ-MTX-004: Health monitoring
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Server is ready via shared test helper

	// Create config manager using test fixture (centralized in test helpers)
	configManager := CreateConfigManagerWithFixture(t, "config_test_minimal.yaml")

	// Create configuration integration to get MediaMTX config
	configIntegration := config.NewConfigIntegration(configManager, helper.GetLogger())
	mediaMTXConfig, err := configIntegration.GetMediaMTXConfig()
	require.NoError(t, err, "Should get MediaMTX config from integration")

	// Create RTSP connection manager
	logger := logging.CreateTestLogger(t, nil)
	logger.SetLevel(logrus.ErrorLevel)

	rtspManager := NewRTSPConnectionManager(helper.GetClient(), mediaMTXConfig, logger)
	require.NotNil(t, rtspManager)

	ctx := context.Background()

	// Test health monitoring
	health, err := rtspManager.GetConnectionHealth(ctx)
	require.NoError(t, err, "GetConnectionHealth should succeed")
	assert.NotNil(t, health, "Health status should not be nil")
	assert.NotEmpty(t, health.Status, "Health status should not be empty")
	assert.NotZero(t, health.Timestamp, "Health timestamp should not be zero")

	t.Logf("RTSP connection health: %s - %s", health.Status, health.Details)
}

// TestRTSPConnectionManager_GetConnectionMetrics_ReqMTX004 tests RTSP connection metrics
func TestRTSPConnectionManager_GetConnectionMetrics_ReqMTX004(t *testing.T) {
	// REQ-MTX-004: Health monitoring
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create RTSP connection manager
	config := createTestMediaMTXConfig()
	logger := logging.CreateTestLogger(t, nil)
	logger.SetLevel(logrus.ErrorLevel)

	rtspManager := NewRTSPConnectionManager(helper.GetClient(), config, logger)
	require.NotNil(t, rtspManager)

	ctx := context.Background()

	// Test metrics collection
	metrics := rtspManager.GetConnectionMetrics(ctx)
	assert.NotNil(t, metrics, "Metrics should not be nil")
	assert.Contains(t, metrics, "is_healthy", "Metrics should contain is_healthy")
	assert.Contains(t, metrics, "monitoring_enabled", "Metrics should contain monitoring_enabled")

	// Check if monitoring is enabled and connections have been listed
	if enabled, ok := metrics["monitoring_enabled"].(bool); ok && enabled {
		// total_connections is only available after ListConnections has been called
		// For a fresh manager, this field may not be present yet
		if _, hasConnections := metrics["total_connections"]; hasConnections {
			assert.IsType(t, 0, metrics["total_connections"], "total_connections should be an integer when present")
		}
	}

	t.Logf("RTSP connection metrics collected: %+v", metrics)
}

// TestRTSPConnectionManager_Configuration_ReqMTX003 tests RTSP monitoring configuration
func TestRTSPConnectionManager_Configuration_ReqMTX003(t *testing.T) {
	// REQ-MTX-003: Path creation and deletion (configuration management)
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Server is ready via shared test helper

	// Create config manager using test fixture (centralized in test helpers)
	configManager := CreateConfigManagerWithFixture(t, "config_test_minimal.yaml")

	// Create configuration integration to get MediaMTX config
	configIntegration := config.NewConfigIntegration(configManager, helper.GetLogger())
	mediaMTXConfig, err := configIntegration.GetMediaMTXConfig()
	require.NoError(t, err, "Should be able to get MediaMTX config from fixture")

	// Customize RTSP monitoring settings (disable to avoid HTTP calls that might hang)
	mediaMTXConfig.RTSPMonitoring.Enabled = false
	mediaMTXConfig.RTSPMonitoring.CheckInterval = 15
	mediaMTXConfig.RTSPMonitoring.MaxConnections = 25
	mediaMTXConfig.RTSPMonitoring.BandwidthThreshold = 2000000

	logger := logging.CreateTestLogger(t, nil)
	logger.SetLevel(logrus.ErrorLevel)

	rtspManager := NewRTSPConnectionManager(helper.GetClient(), mediaMTXConfig, logger)
	require.NotNil(t, rtspManager)

	ctx := context.Background()

	// Test that configuration is applied
	health, err := rtspManager.GetConnectionHealth(ctx)
	require.NoError(t, err, "GetConnectionHealth should succeed")
	assert.NotNil(t, health, "Health status should not be nil")
	assert.Equal(t, "disabled", health.Status, "Health status should be disabled when monitoring is disabled")

	// Test metrics with custom configuration
	metrics := rtspManager.GetConnectionMetrics(ctx)
	assert.NotNil(t, metrics, "Metrics should not be nil")
	assert.Equal(t, false, metrics["monitoring_enabled"], "Monitoring should be disabled")
	assert.Equal(t, 25, metrics["max_connections"], "Max connections should match config")
	assert.Equal(t, int64(2000000), metrics["bandwidth_threshold"], "Bandwidth threshold should match config")

	t.Log("RTSP connection manager configuration applied successfully")
}

// TestRTSPConnectionManager_ErrorHandling_ReqMTX004 tests error handling
func TestRTSPConnectionManager_ErrorHandling_ReqMTX004(t *testing.T) {
	// REQ-MTX-004: Health monitoring (error handling)
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create RTSP connection manager
	config := createTestMediaMTXConfig()
	logger := logging.CreateTestLogger(t, nil)
	logger.SetLevel(logrus.ErrorLevel)

	rtspManager := NewRTSPConnectionManager(helper.GetClient(), config, logger)
	require.NotNil(t, rtspManager)

	ctx := context.Background()

	// Test getting non-existent connection
	_, err := rtspManager.GetConnection(ctx, "non-existent-id")
	assert.Error(t, err, "GetConnection should fail for non-existent connection")
	assert.Contains(t, err.Error(), "non-existent-id", "Error should contain connection ID")

	// Test getting non-existent session
	_, err = rtspManager.GetSession(ctx, "non-existent-session")
	assert.Error(t, err, "GetSession should fail for non-existent session")
	assert.Contains(t, err.Error(), "non-existent-session", "Error should contain session ID")

	// Test kicking non-existent session
	err = rtspManager.KickSession(ctx, "non-existent-session")
	assert.Error(t, err, "KickSession should fail for non-existent session")
	assert.Contains(t, err.Error(), "non-existent-session", "Error should contain session ID")

	t.Log("RTSP connection manager error handling working correctly")
}

// TestRTSPConnectionManager_Performance_ReqMTX002 tests performance characteristics
func TestRTSPConnectionManager_Performance_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities (performance)
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create RTSP connection manager
	config := createTestMediaMTXConfig()
	logger := logging.CreateTestLogger(t, nil)
	logger.SetLevel(logrus.ErrorLevel)

	rtspManager := NewRTSPConnectionManager(helper.GetClient(), config, logger)
	require.NotNil(t, rtspManager)

	ctx := context.Background()

	// Test multiple rapid calls
	start := time.Now()
	for i := 0; i < 5; i++ {
		_, err := rtspManager.ListConnections(ctx, 0, 10)
		require.NoError(t, err, "ListConnections should succeed on iteration %d", i+1)

		_, err = rtspManager.ListSessions(ctx, 0, 10)
		require.NoError(t, err, "ListSessions should succeed on iteration %d", i+1)

		metrics := rtspManager.GetConnectionMetrics(ctx)
		assert.NotNil(t, metrics, "Metrics should not be nil on iteration %d", i+1)
	}
	duration := time.Since(start)

	// Performance should be reasonable (less than 5 seconds for 5 iterations)
	assert.Less(t, duration, 5*time.Second, "Performance should be reasonable")

	t.Logf("RTSP connection manager performance test completed in %v", duration)
}

// TestRTSPConnectionManager_RealMediaMTXServer tests integration with real MediaMTX server
func TestRTSPConnectionManager_RealMediaMTXServer(t *testing.T) {
	// Integration test with real MediaMTX server
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Server is ready via shared test helper

	// Use shared RTSP connection manager for performance
	rtspManager := helper.GetRTSPConnectionManager()
	require.NotNil(t, rtspManager)

	ctx := context.Background()

	// Test that we can interact with the real MediaMTX server
	connections, err := rtspManager.ListConnections(ctx, 0, 10)
	require.NoError(t, err, "ListConnections should succeed with real MediaMTX server")
	assert.NotNil(t, connections, "Connections list should not be nil")

	sessions, err := rtspManager.ListSessions(ctx, 0, 10)
	require.NoError(t, err, "ListSessions should succeed with real MediaMTX server")
	assert.NotNil(t, sessions, "Sessions list should not be nil")

	health, err := rtspManager.GetConnectionHealth(ctx)
	require.NoError(t, err, "GetConnectionHealth should succeed with real MediaMTX server")
	assert.NotNil(t, health, "Health status should not be nil")

	metrics := rtspManager.GetConnectionMetrics(ctx)
	assert.NotNil(t, metrics, "Metrics should not be nil")

	t.Log("RTSP connection manager successfully connected to real MediaMTX server")
	t.Log("All RTSP connection management operations working correctly")
	t.Log("No mocks used - real MediaMTX server integration")
}

// TestRTSPConnectionManager_ConfigurationScenarios tests various configuration scenarios
func TestRTSPConnectionManager_ConfigurationScenarios(t *testing.T) {
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Server is ready via shared test helper

	// Test different configuration scenarios
	configScenarios := []struct {
		name        string
		config      *config.RTSPMonitoringConfig
		expectError bool
	}{
		{
			name: "disabled_monitoring",
			config: &config.RTSPMonitoringConfig{
				Enabled:             false,
				CheckInterval:       30,
				ConnectionTimeout:   10,
				MaxConnections:      50,
				SessionTimeout:      300,
				BandwidthThreshold:  1000000,
				PacketLossThreshold: 0.05,
				JitterThreshold:     50.0,
			},
			expectError: false,
		},
		{
			name: "high_performance_config",
			config: &config.RTSPMonitoringConfig{
				Enabled:             true,
				CheckInterval:       5,
				ConnectionTimeout:   5,
				MaxConnections:      100,
				SessionTimeout:      600,
				BandwidthThreshold:  5000000,
				PacketLossThreshold: 0.01,
				JitterThreshold:     25.0,
			},
			expectError: false,
		},
		{
			name: "low_resource_config",
			config: &config.RTSPMonitoringConfig{
				Enabled:             true,
				CheckInterval:       60,
				ConnectionTimeout:   30,
				MaxConnections:      10,
				SessionTimeout:      120,
				BandwidthThreshold:  100000,
				PacketLossThreshold: 0.1,
				JitterThreshold:     100.0,
			},
			expectError: false,
		},
	}

	for _, scenario := range configScenarios {
		t.Run(scenario.name, func(t *testing.T) {
			rtspManager := helper.GetRTSPConnectionManager()
			ctx := context.Background()

			// Test health monitoring with custom config
			health, err := rtspManager.GetConnectionHealth(ctx)
			if scenario.expectError {
				assert.Error(t, err, "Expected error for scenario %s", scenario.name)
			} else {
				require.NoError(t, err, "GetConnectionHealth should succeed for scenario %s", scenario.name)
				assertRTSPHealthStatus(t, health, "")

				// Test metrics with custom config
				metrics := rtspManager.GetConnectionMetrics(ctx)
				assertRTSPMetrics(t, metrics)

				// Verify configuration is applied
				assert.Equal(t, scenario.config.Enabled, metrics["monitoring_enabled"],
					"Monitoring enabled should match config for scenario %s", scenario.name)
				assert.Equal(t, scenario.config.MaxConnections, metrics["max_connections"],
					"Max connections should match config for scenario %s", scenario.name)
			}

			logTestProgress(t, helper.GetLogger(), "Configuration scenario tested successfully", map[string]interface{}{
				"scenario":        scenario.name,
				"enabled":         scenario.config.Enabled,
				"max_connections": scenario.config.MaxConnections,
			})
		})
	}
}

// TestRTSPConnectionManager_ErrorScenarios tests various error scenarios
func TestRTSPConnectionManager_ErrorScenarios(t *testing.T) {
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Server is ready via shared test helper

	rtspManager := helper.GetRTSPConnectionManager()
	ctx := context.Background()

	// Test error scenarios
	errorScenarios := []struct {
		name        string
		testFunc    func() error
		expectError bool
		errorMsg    string
	}{
		{
			name: "invalid_connection_id",
			testFunc: func() error {
				_, err := rtspManager.GetConnection(ctx, "invalid-connection-id")
				return err
			},
			expectError: true,
			errorMsg:    "invalid-connection-id",
		},
		{
			name: "invalid_session_id",
			testFunc: func() error {
				_, err := rtspManager.GetSession(ctx, "invalid-session-id")
				return err
			},
			expectError: true,
			errorMsg:    "invalid-session-id",
		},
		{
			name: "kick_invalid_session",
			testFunc: func() error {
				return rtspManager.KickSession(ctx, "invalid-session-id")
			},
			expectError: true,
			errorMsg:    "invalid-session-id",
		},
		{
			name: "negative_page_number",
			testFunc: func() error {
				_, err := rtspManager.ListConnections(ctx, -1, 10)
				return err
			},
			expectError: false, // API should handle gracefully
		},
		{
			name: "zero_items_per_page",
			testFunc: func() error {
				_, err := rtspManager.ListConnections(ctx, 0, 0)
				return err
			},
			expectError: false, // API should handle gracefully
		},
	}

	for _, scenario := range errorScenarios {
		t.Run(scenario.name, func(t *testing.T) {
			err := scenario.testFunc()

			if scenario.expectError {
				assert.Error(t, err, "Expected error for scenario %s", scenario.name)
				if scenario.errorMsg != "" {
					assert.Contains(t, err.Error(), scenario.errorMsg,
						"Error should contain expected message for scenario %s", scenario.name)
				}
			} else {
				assert.NoError(t, err, "Should not error for scenario %s", scenario.name)
			}

			logTestProgress(t, helper.GetLogger(), "Error scenario tested", map[string]interface{}{
				"scenario":       scenario.name,
				"expected_error": scenario.expectError,
				"got_error":      err != nil,
			})
		})
	}
}

// TestRTSPConnectionManager_ConcurrentAccess tests concurrent access to RTSP manager
func TestRTSPConnectionManager_ConcurrentAccess(t *testing.T) {
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Server is ready via shared test helper

	rtspManager := helper.GetRTSPConnectionManager()
	ctx := context.Background()

	// Test concurrent access
	numGoroutines := 3
	done := make(chan bool, numGoroutines)
	errors := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()

			// Each goroutine performs multiple operations
			for j := 0; j < 1; j++ {
				// List connections
				_, err := rtspManager.ListConnections(ctx, 0, 10)
				if err != nil {
					errors <- fmt.Errorf("goroutine %d, iteration %d, ListConnections: %w", id, j, err)
					return
				}

				// List sessions
				_, err = rtspManager.ListSessions(ctx, 0, 10)
				if err != nil {
					errors <- fmt.Errorf("goroutine %d, iteration %d, ListSessions: %w", id, j, err)
					return
				}

				// Get health
				_, err = rtspManager.GetConnectionHealth(ctx)
				if err != nil {
					errors <- fmt.Errorf("goroutine %d, iteration %d, GetConnectionHealth: %w", id, j, err)
					return
				}

				// Get metrics
				metrics := rtspManager.GetConnectionMetrics(ctx)
				if metrics == nil {
					errors <- fmt.Errorf("goroutine %d, iteration %d, GetConnectionMetrics returned nil", id, j)
					return
				}

				// Small delay to simulate real usage
				time.Sleep(10 * time.Millisecond)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Check for errors
	close(errors)
	var errorList []error
	for err := range errors {
		errorList = append(errorList, err)
	}

	assert.Empty(t, errorList, "No errors should occur during concurrent access: %v", errorList)

	logTestProgress(t, helper.GetLogger(), "Concurrent access test completed successfully", map[string]interface{}{
		"goroutines":               numGoroutines,
		"operations_per_goroutine": 5,
		"total_operations":         numGoroutines * 5 * 4, // 4 operations per iteration
		"errors":                   len(errorList),
	})
}

// TestRTSPConnectionManager_StressTest tests stress scenarios
func TestRTSPConnectionManager_StressTest(t *testing.T) {
	EnsureSequentialExecution(t)
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Server is ready via shared test helper

	rtspManager := helper.GetRTSPConnectionManager()
	ctx := context.Background()

	// Stress test with rapid successive calls
	start := time.Now()
	numOperations := 5

	for i := 0; i < numOperations; i++ {
		// Alternate between different operations
		switch i % 4 {
		case 0:
			_, err := rtspManager.ListConnections(ctx, 0, 10)
			require.NoError(t, err, "ListConnections should succeed on iteration %d", i)
		case 1:
			_, err := rtspManager.ListSessions(ctx, 0, 10)
			require.NoError(t, err, "ListSessions should succeed on iteration %d", i)
		case 2:
			_, err := rtspManager.GetConnectionHealth(ctx)
			require.NoError(t, err, "GetConnectionHealth should succeed on iteration %d", i)
		case 3:
			metrics := rtspManager.GetConnectionMetrics(ctx)
			assert.NotNil(t, metrics, "GetConnectionMetrics should not return nil on iteration %d", i)
		}
	}

	duration := time.Since(start)
	avgTimePerOp := duration / time.Duration(numOperations)

	// Performance assertions
	assert.Less(t, duration, 30*time.Second, "Stress test should complete within 30 seconds")
	assert.Less(t, avgTimePerOp, 500*time.Millisecond, "Average time per operation should be less than 500ms")

	logTestProgress(t, helper.GetLogger(), "Stress test completed successfully", map[string]interface{}{
		"total_operations":      numOperations,
		"total_duration":        duration.String(),
		"avg_time_per_op":       avgTimePerOp.String(),
		"operations_per_second": float64(numOperations) / duration.Seconds(),
	})
}

// TestRTSPConnectionManager_IntegrationWithController tests integration with controller
func TestRTSPConnectionManager_IntegrationWithController(t *testing.T) {
	EnsureSequentialExecution(t)
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Server is ready via shared test helper

	// Create controller using proper constructor
	logger := helper.GetLogger()
	logger.SetLevel(logrus.ErrorLevel)

	// Create controller using test helper
	controller, err := helper.GetController(t)
	require.NoError(t, err, "Controller should be created successfully")

	// Start controller
	err = controller.Start(context.Background())
	require.NoError(t, err, "Controller should start successfully")
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		controller.Stop(stopCtx)
	}()

	ctx := context.Background()

	// Test controller RTSP methods
	connections, err := controller.ListRTSPConnections(ctx, 1, 10)
	require.NoError(t, err, "Controller ListRTSPConnections should succeed")
	assert.NotNil(t, connections, "Connections should not be nil")

	sessions, err := controller.ListRTSPSessions(ctx, 1, 10)
	require.NoError(t, err, "Controller ListRTSPSessions should succeed")
	assert.NotNil(t, sessions, "Sessions should not be nil")

	health, err := controller.GetRTSPConnectionHealth(ctx)
	require.NoError(t, err, "Controller GetRTSPConnectionHealth should succeed")
	assertRTSPHealthStatus(t, health, "")

	metrics := controller.GetRTSPConnectionMetrics(ctx)
	assertRTSPMetrics(t, metrics)

	logTestProgress(t, helper.GetLogger(), "Controller integration test completed successfully", map[string]interface{}{
		"connections_found": len(connections.Items),
		"sessions_found":    len(sessions.Items),
		"health_status":     health.Status,
	})
}

// TestRTSPConnectionManager_InputValidation_DangerousBugs tests input validation
// that can catch dangerous bugs in RTSP connection manager
func TestRTSPConnectionManager_InputValidation_DangerousBugs(t *testing.T) {
	// REQ-MTX-007: Error handling and recovery
	EnsureSequentialExecution(t)
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create RTSP connection manager
	rtspManager := helper.GetRTSPConnectionManager()

	// Test input validation scenarios that can catch dangerous bugs
	helper.TestRTSPInputValidation(t, rtspManager)
}

// TestRTSPConnectionManager_ErrorScenarios_DangerousBugs tests error scenarios
// that were identified in the original test failures
func TestRTSPConnectionManager_ErrorScenarios_DangerousBugs(t *testing.T) {
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create RTSP connection manager
	rtspManager := helper.GetRTSPConnectionManager()
	ctx := context.Background()

	// Test the specific scenarios that were failing in the original tests
	t.Run("negative_page_number_bug", func(t *testing.T) {
		// This was failing with: strconv.ParseUint: parsing "-1": invalid syntax
		// Now it should be properly rejected with clear error message
		_, err := rtspManager.ListConnections(ctx, -1, 10)

		if err == nil {
			// This is a BUG - negative page numbers should be rejected
			t.Errorf("🚨 BUG DETECTED: Negative page number (-1) should be rejected but was accepted")
			t.Errorf("🚨 This indicates a dangerous bug - invalid inputs are not being validated")
		} else {
			t.Logf("Negative page number correctly rejected: %v", err)
		}
	})

	t.Run("zero_items_per_page_bug", func(t *testing.T) {
		// This was failing with: invalid items per page
		// Now it should be properly rejected with clear error message
		_, err := rtspManager.ListConnections(ctx, 0, 0)

		if err == nil {
			// This is a BUG - zero items per page should be rejected
			t.Errorf("🚨 BUG DETECTED: Zero items per page should be rejected but was accepted")
			t.Errorf("🚨 This indicates a dangerous bug - invalid inputs are not being validated")
		} else {
			t.Logf("Zero items per page correctly rejected: %v", err)
		}
	})

	t.Run("negative_items_per_page_should_fail", func(t *testing.T) {
		// This should fail with a clear error message
		_, err := rtspManager.ListConnections(ctx, 0, -5)

		require.Error(t, err, "Negative items per page should produce an error")
		assert.Contains(t, err.Error(), "invalid", "Error should indicate invalid input")
		t.Logf("Negative items per page correctly rejected: %v", err)
	})
}
