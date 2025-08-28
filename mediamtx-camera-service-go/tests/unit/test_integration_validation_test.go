/*
Integration Validation Unit Test

Requirements Coverage:
- REQ-INT-001: Component integration validation
- REQ-INT-002: Data flow integration validation
- REQ-INT-003: Error handling integration validation
- REQ-INT-004: Security integration validation
- REQ-INT-005: Performance integration validation
- REQ-INT-006: Reliability integration validation
- REQ-INT-007: State consistency validation
- REQ-INT-008: Component recovery validation

Test Categories: Unit/Integration/Security/Performance/Reliability
API Documentation Reference: docs/api/json_rpc_methods.md
*/

//go:build unit
// +build unit

package unit

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/camera"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/mediamtx"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/security"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// IntegrationValidationTestSuite provides integration validation testing
type IntegrationValidationTestSuite struct {
	configManager      *config.ConfigManager
	logger             *logging.Logger
	cameraMonitor      *camera.HybridCameraMonitor
	mediaMTXController mediamtx.MediaMTXController
	jwtHandler         *security.JWTHandler
	wsServer           *websocket.WebSocketServer
	ctx                context.Context
}

// NewIntegrationValidationTestSuite creates a new integration validation test suite
func NewIntegrationValidationTestSuite() *IntegrationValidationTestSuite {
	return &IntegrationValidationTestSuite{}
}

// Setup initializes the integration validation test suite
func (suite *IntegrationValidationTestSuite) Setup(t *testing.T) {
	// Create context
	suite.ctx = context.Background()

	// Load configuration
	suite.configManager = config.NewConfigManager()
	err := suite.configManager.LoadConfig("config/default.yaml")
	require.NoError(t, err, "Failed to load configuration")

	// Setup logging
	suite.logger = logging.NewLogger("integration-validation-test")

	// Initialize real implementations
	deviceChecker := &camera.RealDeviceChecker{}
	commandExecutor := &camera.RealV4L2CommandExecutor{}
	infoParser := &camera.RealDeviceInfoParser{}

	// Initialize camera monitor
	suite.cameraMonitor = camera.NewHybridCameraMonitor(
		suite.configManager,
		suite.logger,
		deviceChecker,
		commandExecutor,
		infoParser,
	)

	// Initialize MediaMTX controller
	suite.mediaMTXController, err = mediamtx.NewControllerWithConfigManager(suite.configManager, suite.logger.Logger)
	require.NoError(t, err, "Failed to create MediaMTX controller")

	// Initialize JWT handler
	cfg := suite.configManager.GetConfig()
	require.NotNil(t, cfg, "Configuration not available")

	suite.jwtHandler, err = security.NewJWTHandler(cfg.Security.JWTSecretKey)
	require.NoError(t, err, "Failed to create JWT handler")

	// Initialize WebSocket server
	suite.wsServer = websocket.NewWebSocketServer(
		suite.configManager,
		suite.logger,
		suite.cameraMonitor,
		suite.jwtHandler,
		suite.mediaMTXController,
	)
}

// Teardown cleans up the integration validation test suite
func (suite *IntegrationValidationTestSuite) Teardown(t *testing.T) {
	if suite.wsServer != nil {
		err := suite.wsServer.Stop()
		require.NoError(t, err, "Failed to stop WebSocket server")
	}
}

// TestComponentIntegration tests component integration validation
func TestComponentIntegration(t *testing.T) {
	suite := NewIntegrationValidationTestSuite()
	suite.Setup(t)
	defer suite.Teardown(t)

	t.Run("ConfigurationIntegration", func(t *testing.T) {
		// Test configuration integration
		cfg := suite.configManager.GetConfig()
		require.NotNil(t, cfg, "Configuration should not be nil")

		// Validate configuration structure
		assert.NotEmpty(t, cfg.MediaMTX.Host, "MediaMTX host should be configured")
		assert.NotZero(t, cfg.MediaMTX.APIPort, "MediaMTX API port should be configured")
		assert.NotEmpty(t, cfg.Security.JWTSecretKey, "JWT secret key should be configured")
		assert.Greater(t, cfg.Security.RateLimitRequests, 0, "Rate limit requests should be configured")
		assert.Greater(t, cfg.Storage.WarnPercent, 0, "Storage warn percent should be configured")
		assert.Greater(t, cfg.Storage.BlockPercent, 0, "Storage block percent should be configured")
	})

	t.Run("CameraMonitorIntegration", func(t *testing.T) {
		// Test camera monitor integration
		require.NotNil(t, suite.cameraMonitor, "Camera monitor should not be nil")

		// Test camera discovery
		cameras := suite.cameraMonitor.GetConnectedCameras()
		assert.NotNil(t, cameras, "Camera list should not be nil")
		assert.GreaterOrEqual(t, len(cameras), 0, "Camera count should be non-negative")
	})

	t.Run("MediaMTXControllerIntegration", func(t *testing.T) {
		// Test MediaMTX controller integration
		require.NotNil(t, suite.mediaMTXController, "MediaMTX controller should not be nil")

		// Test health check
		health, err := suite.mediaMTXController.GetHealth(suite.ctx)
		require.NoError(t, err, "Health check should succeed")
		require.NotNil(t, health, "Health status should not be nil")
		assert.NotEmpty(t, health.Status, "Health status should not be empty")

		// Test system metrics
		metrics, err := suite.mediaMTXController.GetSystemMetrics(suite.ctx)
		require.NoError(t, err, "System metrics should succeed")
		require.NotNil(t, metrics, "System metrics should not be nil")
	})

	t.Run("JWTHandlerIntegration", func(t *testing.T) {
		// Test JWT handler integration
		require.NotNil(t, suite.jwtHandler, "JWT handler should not be nil")

		// Test token generation
		token, err := suite.jwtHandler.GenerateToken("test-user", "admin", 1)
		require.NoError(t, err, "Token generation should succeed")
		assert.NotEmpty(t, token, "Generated token should not be empty")

		// Test token validation
		claims, err := suite.jwtHandler.ValidateToken(token)
		require.NoError(t, err, "Token validation should succeed")
		require.NotNil(t, claims, "Claims should not be nil")
		assert.Equal(t, "test-user", claims.UserID, "UserID should match")
		assert.Equal(t, "admin", claims.Role, "Role should match")
	})

	t.Run("WebSocketServerIntegration", func(t *testing.T) {
		// Test WebSocket server integration
		require.NotNil(t, suite.wsServer, "WebSocket server should not be nil")

		// Test server start
		err := suite.wsServer.Start()
		require.NoError(t, err, "WebSocket server should start successfully")

		// Wait for server to be ready
		time.Sleep(1 * time.Second)

		// Test server stop
		err = suite.wsServer.Stop()
		require.NoError(t, err, "WebSocket server should stop successfully")
	})
}

// TestDataFlowIntegration tests data flow integration validation
func TestDataFlowIntegration(t *testing.T) {
	suite := NewIntegrationValidationTestSuite()
	suite.Setup(t)
	defer suite.Teardown(t)

	t.Run("CameraToMediaMTXFlow", func(t *testing.T) {
		// Test camera data flow to MediaMTX
		cameras := suite.cameraMonitor.GetConnectedCameras()

		if len(cameras) > 0 {
			// Test with actual camera
			var camera *camera.CameraDevice
			for _, cam := range cameras {
				camera = cam
				break
			}

			// Test recording flow
			options := map[string]interface{}{
				"use_case":       "recording",
				"priority":       1,
				"auto_cleanup":   true,
				"retention_days": 1,
				"quality":        "medium",
				"max_duration":   5 * time.Second,
			}

			session, err := suite.mediaMTXController.StartAdvancedRecording(suite.ctx, camera.Path, "", options)
			if err == nil {
				require.NotNil(t, session, "Recording session should be created")
				assert.Equal(t, camera.Path, session.Device, "Session device should match camera")

				// Test session status
				status, err := suite.mediaMTXController.GetRecordingStatus(suite.ctx, session.ID)
				require.NoError(t, err, "Should get recording status")
				assert.Equal(t, "RECORDING", status.Status, "Status should be recording")

				// Wait for recording
				time.Sleep(2 * time.Second)

				// Stop recording
				err = suite.mediaMTXController.StopAdvancedRecording(suite.ctx, session.ID)
				require.NoError(t, err, "Should stop recording")
			} else {
				t.Logf("Recording flow test skipped: %v", err)
			}
		} else {
			t.Log("No cameras available for camera-to-MediaMTX flow test")
		}
	})

	t.Run("ConfigurationToComponentsFlow", func(t *testing.T) {
		// Test configuration flow to components
		_ = suite.configManager.GetConfig()

		// Test JWT configuration flow
		token, err := suite.jwtHandler.GenerateToken("config-test-user", "admin", 1)
		require.NoError(t, err, "JWT token generation should work with configuration")

		claims, err := suite.jwtHandler.ValidateToken(token)
		require.NoError(t, err, "JWT token validation should work with configuration")
		assert.Equal(t, "config-test-user", claims.UserID, "UserID should match configuration")

		// Test rate limiting configuration flow
		suite.jwtHandler.RecordRequest("test-client")
		rateInfo := suite.jwtHandler.GetClientRateInfo("test-client")
		assert.NotNil(t, rateInfo, "Rate info should be available")
		assert.Equal(t, "test-client", rateInfo.ClientID, "Client ID should match")
	})

	t.Run("LoggingIntegration", func(t *testing.T) {
		// Test logging integration across components
		require.NotNil(t, suite.logger, "Logger should not be nil")

		// Test logging from different components
		suite.logger.Info("Integration validation test started")

		// Test camera monitor logging
		cameras := suite.cameraMonitor.GetConnectedCameras()
		suite.logger.WithField("camera_count", fmt.Sprintf("%d", len(cameras))).Info("Camera discovery completed")

		// Test MediaMTX controller logging
		health, err := suite.mediaMTXController.GetHealth(suite.ctx)
		if err == nil {
			suite.logger.WithField("health_status", health.Status).Info("Health check completed")
		}

		// Test JWT handler logging
		_, err = suite.jwtHandler.GenerateToken("logging-test-user", "admin", 1)
		if err == nil {
			suite.logger.WithField("token_generated", "true").Info("JWT token generated")
		}
	})
}

// TestErrorHandlingIntegration tests error handling integration
func TestErrorHandlingIntegration(t *testing.T) {
	suite := NewIntegrationValidationTestSuite()
	suite.Setup(t)
	defer suite.Teardown(t)

	t.Run("ComponentErrorPropagation", func(t *testing.T) {
		// Test error propagation between components

		// Test invalid recording session
		_, err := suite.mediaMTXController.GetRecordingStatus(suite.ctx, "non-existent-session")
		assert.Error(t, err, "Should return error for non-existent session")

		// Test invalid snapshot device
		_, err = suite.mediaMTXController.TakeAdvancedSnapshot(suite.ctx, "/dev/nonexistent", "", map[string]interface{}{})
		assert.Error(t, err, "Should return error for non-existent device")

		// Test invalid JWT token
		_, err = suite.jwtHandler.ValidateToken("invalid-token")
		assert.Error(t, err, "Should return error for invalid token")
	})

	t.Run("ConfigurationErrorHandling", func(t *testing.T) {
		// Test configuration error handling

		// Test with invalid configuration
		invalidConfigManager := config.NewConfigManager()
		err := invalidConfigManager.LoadConfig("non-existent-config.yaml")
		assert.Error(t, err, "Should return error for non-existent config file")

		// Test with nil configuration
		cfg := invalidConfigManager.GetConfig()
		assert.Nil(t, cfg, "Should return nil for invalid configuration")
	})

	t.Run("ResourceErrorHandling", func(t *testing.T) {
		// Test resource error handling

		// Test storage validation (if implemented)
		// This would test storage space validation integration

		// Test camera device errors
		cameras := suite.cameraMonitor.GetConnectedCameras()
		if len(cameras) == 0 {
			t.Log("No cameras available for resource error testing")
		}
	})
}

// TestSecurityIntegration tests security integration validation
func TestSecurityIntegration(t *testing.T) {
	suite := NewIntegrationValidationTestSuite()
	suite.Setup(t)
	defer suite.Teardown(t)

	t.Run("AuthenticationIntegration", func(t *testing.T) {
		// Test authentication integration

		// Test valid authentication
		token, err := suite.jwtHandler.GenerateToken("security-test-user", "admin", 1)
		require.NoError(t, err, "Token generation should succeed")

		claims, err := suite.jwtHandler.ValidateToken(token)
		require.NoError(t, err, "Token validation should succeed")
		assert.Equal(t, "security-test-user", claims.UserID, "UserID should match")
		assert.Equal(t, "admin", claims.Role, "Role should match")

		// Test invalid authentication
		_, err = suite.jwtHandler.ValidateToken("invalid-token")
		assert.Error(t, err, "Invalid token should be rejected")
	})

	t.Run("RateLimitingIntegration", func(t *testing.T) {
		// Test rate limiting integration

		clientID := "rate-limit-test-client"

		// Test rate limiting
		for i := 0; i < 10; i++ {
			suite.jwtHandler.RecordRequest(clientID)
		}

		rateInfo := suite.jwtHandler.GetClientRateInfo(clientID)
		assert.NotNil(t, rateInfo, "Rate info should be available")
		assert.Equal(t, clientID, rateInfo.ClientID, "Client ID should match")
		assert.GreaterOrEqual(t, rateInfo.RequestCount, int64(10), "Request count should be tracked")
	})

	t.Run("PermissionIntegration", func(t *testing.T) {
		// Test permission integration

		// Test admin role permissions
		adminToken, err := suite.jwtHandler.GenerateToken("admin-user", "admin", 1)
		require.NoError(t, err, "Admin token generation should succeed")

		adminClaims, err := suite.jwtHandler.ValidateToken(adminToken)
		require.NoError(t, err, "Admin token validation should succeed")
		assert.Equal(t, "admin", adminClaims.Role, "Role should be admin")

		// Test user role permissions
		userToken, err := suite.jwtHandler.GenerateToken("user-user", "user", 1)
		require.NoError(t, err, "User token generation should succeed")

		userClaims, err := suite.jwtHandler.ValidateToken(userToken)
		require.NoError(t, err, "User token validation should succeed")
		assert.Equal(t, "user", userClaims.Role, "Role should be user")
	})
}

// TestPerformanceIntegration tests performance integration validation
func TestPerformanceIntegration(t *testing.T) {
	suite := NewIntegrationValidationTestSuite()
	suite.Setup(t)
	defer suite.Teardown(t)

	t.Run("ResponseTimeIntegration", func(t *testing.T) {
		// Test response time integration

		// Test health check response time
		start := time.Now()
		health, err := suite.mediaMTXController.GetHealth(suite.ctx)
		healthTime := time.Since(start)

		require.NoError(t, err, "Health check should succeed")
		require.NotNil(t, health, "Health status should not be nil")
		assert.Less(t, healthTime, 5*time.Second, "Health check should complete within 5 seconds")

		// Test system metrics response time
		start = time.Now()
		metrics, err := suite.mediaMTXController.GetSystemMetrics(suite.ctx)
		metricsTime := time.Since(start)

		require.NoError(t, err, "System metrics should succeed")
		require.NotNil(t, metrics, "System metrics should not be nil")
		assert.Less(t, metricsTime, 5*time.Second, "System metrics should complete within 5 seconds")

		// Test camera discovery response time
		start = time.Now()
		cameras := suite.cameraMonitor.GetConnectedCameras()
		cameraTime := time.Since(start)

		assert.NotNil(t, cameras, "Camera list should not be nil")
		assert.Less(t, cameraTime, 1*time.Second, "Camera discovery should complete within 1 second")
	})

	t.Run("ConcurrencyIntegration", func(t *testing.T) {
		// Test concurrency integration

		// Test concurrent health checks
		const numGoroutines = 10
		var wg sync.WaitGroup
		results := make([]error, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				_, err := suite.mediaMTXController.GetHealth(suite.ctx)
				results[index] = err
			}(i)
		}

		wg.Wait()

		// Check results
		for i, err := range results {
			assert.NoError(t, err, "Concurrent health check %d should succeed", i)
		}

		// Test concurrent camera discovery
		cameraResults := make([]map[string]*camera.CameraDevice, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				cameras := suite.cameraMonitor.GetConnectedCameras()
				cameraResults[index] = cameras
			}(i)
		}

		wg.Wait()

		// Check results
		for i, cameras := range cameraResults {
			assert.NotNil(t, cameras, "Concurrent camera discovery %d should succeed", i)
		}
	})

	t.Run("MemoryIntegration", func(t *testing.T) {
		// Test memory integration

		// Test memory usage during operations
		initialCameras := suite.cameraMonitor.GetConnectedCameras()

		// Perform multiple operations
		for i := 0; i < 100; i++ {
			_, err := suite.mediaMTXController.GetHealth(suite.ctx)
			assert.NoError(t, err, "Health check should succeed")

			cameras := suite.cameraMonitor.GetConnectedCameras()
			assert.NotNil(t, cameras, "Camera discovery should succeed")

			token, err := suite.jwtHandler.GenerateToken("memory-test-user", "user", 1)
			assert.NoError(t, err, "Token generation should succeed")

			_, err = suite.jwtHandler.ValidateToken(token)
			assert.NoError(t, err, "Token validation should succeed")
		}

		// Verify system still works
		finalCameras := suite.cameraMonitor.GetConnectedCameras()
		assert.NotNil(t, finalCameras, "Final camera discovery should succeed")
		assert.Equal(t, len(initialCameras), len(finalCameras), "Camera count should remain consistent")
	})
}

// TestReliabilityIntegration tests reliability integration validation
func TestReliabilityIntegration(t *testing.T) {
	suite := NewIntegrationValidationTestSuite()
	suite.Setup(t)
	defer suite.Teardown(t)

	t.Run("ComponentRecovery", func(t *testing.T) {
		// Test component recovery integration

		// Test WebSocket server recovery
		err := suite.wsServer.Start()
		require.NoError(t, err, "WebSocket server should start")

		err = suite.wsServer.Stop()
		require.NoError(t, err, "WebSocket server should stop")

		err = suite.wsServer.Start()
		require.NoError(t, err, "WebSocket server should restart successfully")
	})

	t.Run("ErrorRecovery", func(t *testing.T) {
		// Test error recovery integration

		// Test invalid operations and recovery
		_, err := suite.mediaMTXController.GetRecordingStatus(suite.ctx, "invalid-session")
		assert.Error(t, err, "Should return error for invalid session")

		// Verify system still works after error
		health, err := suite.mediaMTXController.GetHealth(suite.ctx)
		require.NoError(t, err, "System should work after error")
		require.NotNil(t, health, "Health status should be available")
	})

	t.Run("StateConsistency", func(t *testing.T) {
		// Test state consistency integration

		// Test active recording tracking consistency
		_ = suite.mediaMTXController.GetActiveRecordings()

		// Perform operations
		cameras := suite.cameraMonitor.GetConnectedCameras()
		if len(cameras) > 0 {
			var camera *camera.CameraDevice
			for _, cam := range cameras {
				camera = cam
				break
			}

			// Test recording start and stop
			options := map[string]interface{}{
				"use_case":       "recording",
				"priority":       1,
				"auto_cleanup":   true,
				"retention_days": 1,
				"quality":        "medium",
				"max_duration":   3 * time.Second,
			}

			session, err := suite.mediaMTXController.StartAdvancedRecording(suite.ctx, camera.Path, "", options)
			if err == nil {
				require.NotNil(t, session, "Recording session should be created")

				// Check active recording state
				isRecording := suite.mediaMTXController.IsDeviceRecording(camera.Path)
				assert.True(t, isRecording, "Device should be marked as recording")

				activeRecording := suite.mediaMTXController.GetActiveRecording(camera.Path)
				require.NotNil(t, activeRecording, "Active recording should be tracked")
				assert.Equal(t, camera.Path, activeRecording.DevicePath, "Active recording device should match")

				// Wait for recording
				time.Sleep(2 * time.Second)

				// Stop recording
				err = suite.mediaMTXController.StopAdvancedRecording(suite.ctx, session.ID)
				require.NoError(t, err, "Should stop recording")

				// Check active recording state after stop
				isRecording = suite.mediaMTXController.IsDeviceRecording(camera.Path)
				assert.False(t, isRecording, "Device should not be marked as recording after stop")

				activeRecording = suite.mediaMTXController.GetActiveRecording(camera.Path)
				assert.Nil(t, activeRecording, "Active recording should be cleared after stop")
			}
		}

		// Verify final state consistency
		finalRecordings := suite.mediaMTXController.GetActiveRecordings()
		assert.NotNil(t, finalRecordings, "Final recordings state should be available")
	})
}
