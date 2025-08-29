//go:build performance
// +build performance

/*
API Performance Benchmark Test

Requirements Coverage:
- REQ-PERF-001: API response time performance
- REQ-PERF-002: Camera discovery performance
- REQ-PERF-003: Health check performance
- REQ-PERF-004: JWT token performance
- REQ-PERF-005: Active recording tracking performance
- REQ-PERF-006: Configuration access performance
- REQ-PERF-007: Concurrent operation performance
- REQ-PERF-008: Memory usage performance

Test Categories: Performance/Benchmark/System
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package benchmarks_test

import (
	"context"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/camera"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/mediamtx"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/security"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/websocket"
	"github.com/stretchr/testify/require"
)

// BenchmarkSuite provides common setup for all benchmarks
type BenchmarkSuite struct {
	configManager      *config.ConfigManager
	logger             *logging.Logger
	cameraMonitor      *camera.HybridCameraMonitor
	mediaMTXController mediamtx.MediaMTXController
	jwtHandler         *security.JWTHandler
	wsServer           *websocket.WebSocketServer
	ctx                context.Context
}

// NewBenchmarkSuite creates a new benchmark suite
func NewBenchmarkSuite() *BenchmarkSuite {
	return &BenchmarkSuite{}
}

// Setup initializes the benchmark suite
func (suite *BenchmarkSuite) Setup(b *testing.B) {
	// Create context
	suite.ctx = context.Background()

	// Load configuration
	suite.configManager = config.NewConfigManager()
	err := suite.configManager.LoadConfig("config/default.yaml")
	require.NoError(b, err, "Failed to load configuration")

	// Setup logging
	suite.logger = logging.NewLogger("benchmark-suite")

	// Initialize real implementations
	deviceChecker := &camera.RealDeviceChecker{}
	commandExecutor := &camera.RealV4L2CommandExecutor{}
	infoParser := &camera.RealDeviceInfoParser{}

	// Initialize camera monitor
	suite.cameraMonitor, err = camera.NewHybridCameraMonitor(
		suite.configManager,
		suite.logger,
		deviceChecker,
		commandExecutor,
		infoParser,
	)
	require.NoError(b, err, "Failed to create camera monitor")

	// Initialize MediaMTX controller
	suite.mediaMTXController, err = mediamtx.NewControllerWithConfigManager(suite.configManager, suite.logger.Logger)
	require.NoError(b, err, "Failed to create MediaMTX controller")

	// Initialize JWT handler
	cfg := suite.configManager.GetConfig()
	require.NotNil(b, cfg, "Configuration not available")

	suite.jwtHandler, err = security.NewJWTHandler(cfg.Security.JWTSecretKey)
	require.NoError(b, err, "Failed to create JWT handler")

	// Initialize WebSocket server
	suite.wsServer, err = websocket.NewWebSocketServer(
		suite.configManager,
		suite.logger,
		suite.cameraMonitor,
		suite.jwtHandler,
		suite.mediaMTXController,
	)
	require.NoError(b, err, "Failed to create WebSocket server")
}

// Teardown cleans up the benchmark suite
func (suite *BenchmarkSuite) Teardown(b *testing.B) {
	if suite.wsServer != nil {
		err := suite.wsServer.Stop()
		require.NoError(b, err, "Failed to stop WebSocket server")
	}
}

// BenchmarkCameraDiscovery benchmarks camera discovery performance
func BenchmarkCameraDiscovery(b *testing.B) {
	suite := NewBenchmarkSuite()
	suite.Setup(b)
	defer suite.Teardown(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cameras := suite.cameraMonitor.GetConnectedCameras()
		_ = len(cameras)
	}
}

// BenchmarkHealthCheck benchmarks health check performance
func BenchmarkHealthCheck(b *testing.B) {
	suite := NewBenchmarkSuite()
	suite.Setup(b)
	defer suite.Teardown(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		health, err := suite.mediaMTXController.GetHealth(suite.ctx)
		if err != nil {
			b.Fatalf("Health check failed: %v", err)
		}
		_ = health
	}
}

// BenchmarkSystemMetrics benchmarks system metrics retrieval performance
func BenchmarkSystemMetrics(b *testing.B) {
	suite := NewBenchmarkSuite()
	suite.Setup(b)
	defer suite.Teardown(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		metrics, err := suite.mediaMTXController.GetSystemMetrics(suite.ctx)
		if err != nil {
			b.Fatalf("System metrics failed: %v", err)
		}
		_ = metrics
	}
}

// BenchmarkListRecordings benchmarks recording list performance
func BenchmarkListRecordings(b *testing.B) {
	suite := NewBenchmarkSuite()
	suite.Setup(b)
	defer suite.Teardown(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		recordings, err := suite.mediaMTXController.ListRecordings(suite.ctx, 10, 0)
		if err != nil {
			b.Fatalf("List recordings failed: %v", err)
		}
		_ = recordings
	}
}

// BenchmarkListSnapshots benchmarks snapshot list performance
func BenchmarkListSnapshots(b *testing.B) {
	suite := NewBenchmarkSuite()
	suite.Setup(b)
	defer suite.Teardown(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		snapshots, err := suite.mediaMTXController.ListSnapshots(suite.ctx, 10, 0)
		if err != nil {
			b.Fatalf("List snapshots failed: %v", err)
		}
		_ = snapshots
	}
}

// BenchmarkJWTTokenGeneration benchmarks JWT token generation performance
func BenchmarkJWTTokenGeneration(b *testing.B) {
	suite := NewBenchmarkSuite()
	suite.Setup(b)
	defer suite.Teardown(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		token, err := suite.jwtHandler.GenerateToken("test-user", "admin", 1)
		if err != nil {
			b.Fatalf("JWT token generation failed: %v", err)
		}
		_ = token
	}
}

// BenchmarkJWTTokenValidation benchmarks JWT token validation performance
func BenchmarkJWTTokenValidation(b *testing.B) {
	suite := NewBenchmarkSuite()
	suite.Setup(b)
	defer suite.Teardown(b)

	// Generate a token once
	token, err := suite.jwtHandler.GenerateToken("test-user", "admin", 1)
	require.NoError(b, err, "Failed to generate test token")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		claims, err := suite.jwtHandler.ValidateToken(token)
		if err != nil {
			b.Fatalf("JWT token validation failed: %v", err)
		}
		_ = claims
	}
}

// BenchmarkActiveRecordingTracking benchmarks active recording tracking performance
func BenchmarkActiveRecordingTracking(b *testing.B) {
	suite := NewBenchmarkSuite()
	suite.Setup(b)
	defer suite.Teardown(b)

	devicePath := "/dev/video0"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Test multiple active recording operations
		isRecording := suite.mediaMTXController.IsDeviceRecording(devicePath)
		activeRecordings := suite.mediaMTXController.GetActiveRecordings()
		activeRecording := suite.mediaMTXController.GetActiveRecording(devicePath)

		_ = isRecording
		_ = len(activeRecordings)
		_ = activeRecording
	}
}

// BenchmarkConfigurationAccess benchmarks configuration access performance
func BenchmarkConfigurationAccess(b *testing.B) {
	suite := NewBenchmarkSuite()
	suite.Setup(b)
	defer suite.Teardown(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cfg := suite.configManager.GetConfig()
		_ = cfg
	}
}

// BenchmarkConcurrentHealthChecks benchmarks concurrent health check performance
func BenchmarkConcurrentHealthChecks(b *testing.B) {
	suite := NewBenchmarkSuite()
	suite.Setup(b)
	defer suite.Teardown(b)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			health, err := suite.mediaMTXController.GetHealth(suite.ctx)
			if err != nil {
				b.Fatalf("Concurrent health check failed: %v", err)
			}
			_ = health
		}
	})
}

// BenchmarkConcurrentCameraDiscovery benchmarks concurrent camera discovery performance
func BenchmarkConcurrentCameraDiscovery(b *testing.B) {
	suite := NewBenchmarkSuite()
	suite.Setup(b)
	defer suite.Teardown(b)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			cameras := suite.cameraMonitor.GetConnectedCameras()
			_ = len(cameras)
		}
	})
}

// BenchmarkMemoryUsage benchmarks memory usage patterns
func BenchmarkMemoryUsage(b *testing.B) {
	suite := NewBenchmarkSuite()
	suite.Setup(b)
	defer suite.Teardown(b)

	// Pre-allocate some data structures
	largeSlice := make([]string, 1000)
	for i := range largeSlice {
		largeSlice[i] = "test-data"
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate memory-intensive operations
		health, err := suite.mediaMTXController.GetHealth(suite.ctx)
		if err != nil {
			b.Fatalf("Health check failed: %v", err)
		}

		metrics, err := suite.mediaMTXController.GetSystemMetrics(suite.ctx)
		if err != nil {
			b.Fatalf("System metrics failed: %v", err)
		}

		recordings, err := suite.mediaMTXController.ListRecordings(suite.ctx, 100, 0)
		if err != nil {
			b.Fatalf("List recordings failed: %v", err)
		}

		_ = health
		_ = metrics
		_ = recordings
		_ = len(largeSlice)
	}
}

// BenchmarkResponseTimeTargets benchmarks response time targets
func BenchmarkResponseTimeTargets(b *testing.B) {
	suite := NewBenchmarkSuite()
	suite.Setup(b)
	defer suite.Teardown(b)

	// Test different response time scenarios
	testCases := []struct {
		name string
		fn   func() error
	}{
		{
			name: "SnapshotCapture",
			fn: func() error {
				_, err := suite.mediaMTXController.TakeAdvancedSnapshot(suite.ctx, "/dev/video0", "", map[string]interface{}{
					"quality":    85,
					"format":     "jpeg",
					"resolution": "1920x1080",
				})
				return err
			},
		},
		{
			name: "RecordingStart",
			fn: func() error {
				_, err := suite.mediaMTXController.StartAdvancedRecording(suite.ctx, "/dev/video0", "", map[string]interface{}{
					"use_case":       "recording",
					"priority":       1,
					"auto_cleanup":   true,
					"retention_days": 1,
					"quality":        "medium",
					"max_duration":   5 * time.Second,
				})
				return err
			},
		},
		{
			name: "FileListing",
			fn: func() error {
				_, err := suite.mediaMTXController.ListRecordings(suite.ctx, 10, 0)
				return err
			},
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				err := tc.fn()
				if err != nil {
					// Skip if operation is not available (e.g., no camera)
					b.Skipf("Operation not available: %v", err)
				}
			}
		})
	}
}

// BenchmarkRateLimiting benchmarks rate limiting performance
func BenchmarkRateLimiting(b *testing.B) {
	suite := NewBenchmarkSuite()
	suite.Setup(b)
	defer suite.Teardown(b)

	clientID := "benchmark-client"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Test rate limiting operations
		suite.jwtHandler.RecordRequest(clientID)
		rateInfo := suite.jwtHandler.GetClientRateInfo(clientID)
		_ = rateInfo
	}
}

// BenchmarkErrorHandling benchmarks error handling performance
func BenchmarkErrorHandling(b *testing.B) {
	suite := NewBenchmarkSuite()
	suite.Setup(b)
	defer suite.Teardown(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Test operations that might fail
		_, err := suite.mediaMTXController.GetRecordingStatus(suite.ctx, "non-existent-session")
		_ = err // We expect this to fail, but we're benchmarking the error handling

		_, err = suite.mediaMTXController.TakeAdvancedSnapshot(suite.ctx, "/dev/nonexistent", "", map[string]interface{}{})
		_ = err // We expect this to fail, but we're benchmarking the error handling
	}
}

// BenchmarkConfigurationValidation benchmarks configuration validation performance
func BenchmarkConfigurationValidation(b *testing.B) {
	suite := NewBenchmarkSuite()
	suite.Setup(b)
	defer suite.Teardown(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Test configuration validation operations
		cfg := suite.configManager.GetConfig()
		if cfg == nil {
			b.Fatalf("Configuration is nil")
		}

		// Validate key configuration fields
		_ = cfg.MediaMTX.Host
		_ = cfg.MediaMTX.APIPort
		_ = cfg.Security.JWTSecretKey
		_ = cfg.Storage.WarnPercent
	}
}

// BenchmarkLoggingPerformance benchmarks logging performance
func BenchmarkLoggingPerformance(b *testing.B) {
	suite := NewBenchmarkSuite()
	suite.Setup(b)
	defer suite.Teardown(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Test logging operations
		suite.logger.Info("Benchmark log message")
		suite.logger.WithField("benchmark", "test").Debug("Debug log message")
		suite.logger.WithFields(map[string]interface{}{
			"iteration": i,
			"timestamp": time.Now(),
		}).Info("Structured log message")
	}
}

// BenchmarkConcurrentOperations benchmarks concurrent operations performance
func BenchmarkConcurrentOperations(b *testing.B) {
	suite := NewBenchmarkSuite()
	suite.Setup(b)
	defer suite.Teardown(b)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Test multiple concurrent operations
			health, _ := suite.mediaMTXController.GetHealth(suite.ctx)
			metrics, _ := suite.mediaMTXController.GetSystemMetrics(suite.ctx)
			cameras := suite.cameraMonitor.GetConnectedCameras()
			recordings, _ := suite.mediaMTXController.ListRecordings(suite.ctx, 5, 0)

			_ = health
			_ = metrics
			_ = len(cameras)
			_ = recordings
		}
	})
}
