/*
WebSocket Server Helper for E2E Tests

Provides WebSocket server lifecycle management for E2E and integration tests.
Creates server with dynamic port allocation for concurrent test safety.

Key Features:
- Dynamic port allocation (no port conflicts)
- Automatic cleanup via t.Cleanup()
- Uses UniversalTestSetup for configuration
- Real WebSocket server (not mocked)
- Follows main.go initialization patterns

Usage:
    setup := testutils.SetupTest(t, "config_valid_complete.yaml")
    server := NewWebSocketServerHelper(t, setup)
    client := testutils.NewWebSocketTestClient(t, server.GetServerURL())
*/

package testutils

import (
	"context"
	"fmt"
	"net"
	"testing"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/camera"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/mediamtx"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/security"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/testutils"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/websocket"
	"github.com/stretchr/testify/require"
)

// WebSocketServerHelper manages WebSocket server lifecycle for testing
type WebSocketServerHelper struct {
	setup  *testutils.UniversalTestSetup
	server *websocket.WebSocketServer
	url    string
}

// NewWebSocketServerHelper creates and starts a WebSocket server for testing
// Follows main.go initialization pattern with proper component lifecycle
func NewWebSocketServerHelper(t *testing.T, setup *testutils.UniversalTestSetup) *WebSocketServerHelper {
	// Get dependencies from setup
	configManager := setup.GetConfigManager()
	config := configManager.GetConfig()
	logger := setup.GetLogger()

	// Allocate dynamic port for concurrent test safety
	listener, err := net.Listen("tcp", ":0")
	require.NoError(t, err, "Failed to allocate free port")
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	// Update config with dynamic port
	config.Server.Port = port

	// Create JWT handler (following main.go pattern)
	jwtHandler, err := security.NewJWTHandler(config.Security.JWTSecretKey, logger)
	require.NoError(t, err, "Failed to create JWT handler")

	// Create camera monitor (following main.go pattern)
	cameraMonitor, err := camera.NewHybridCameraMonitor(
		configManager,
		logger,
		&camera.RealDeviceChecker{},
		&camera.RealV4L2CommandExecutor{},
		&camera.RealDeviceInfoParser{},
	)
	require.NoError(t, err, "Failed to create camera monitor")

	// Create MediaMTX controller with camera monitor (following main.go pattern)
	mediaMTXControllerIface, err := mediamtx.ControllerWithConfigManager(
		configManager,
		cameraMonitor,
		logger,
	)
	require.NoError(t, err, "Failed to create MediaMTX controller")

	// Start MediaMTX controller (will start camera monitor if not running)
	ctx := context.Background()
	err = mediaMTXControllerIface.Start(ctx)
	require.NoError(t, err, "Failed to start MediaMTX controller")

	// Wait for MediaMTX readiness (following main.go pattern)
	readinessChan := mediaMTXControllerIface.SubscribeToReadiness()
	readinessCtx, cancel := context.WithTimeout(ctx, testutils.UniversalTimeoutExtreme)
	defer cancel()

	select {
	case <-readinessChan:
		logger.Info("MediaMTX controller ready for E2E testing")
	case <-readinessCtx.Done():
		if !mediaMTXControllerIface.IsReady() {
			require.Fail(t, "MediaMTX controller not ready within timeout")
		}
		logger.Info("MediaMTX controller ready (fallback check)")
	}

	// Create WebSocket server (following main.go pattern)
	server, err := websocket.NewWebSocketServer(
		configManager,
		logger,
		jwtHandler,
		mediaMTXControllerIface,
	)
	require.NoError(t, err, "Failed to create WebSocket server")

	// Start WebSocket server
	err = server.Start()
	require.NoError(t, err, "Failed to start WebSocket server")

	// Build WebSocket URL
	url := fmt.Sprintf("ws://%s:%d%s",
		config.Server.Host,
		port,
		config.Server.WebSocketPath)

	helper := &WebSocketServerHelper{
		setup:  setup,
		server: server,
		url:    url,
	}

	// Register cleanup (following main.go shutdown pattern)
	t.Cleanup(func() {
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), testutils.UniversalTimeoutExtreme)
		defer shutdownCancel()

		// Stop server
		if server != nil {
			server.Stop(shutdownCtx)
		}

		// Stop camera monitor
		if cameraMonitor != nil {
			cameraMonitor.Stop(shutdownCtx)
		}

		// Stop MediaMTX controller
		if mediaMTXControllerIface != nil {
			mediaMTXControllerIface.Stop(shutdownCtx)
		}

		logger.Info("E2E test server cleanup completed")
	})

	return helper
}

// GetServerURL returns the WebSocket server URL
func (h *WebSocketServerHelper) GetServerURL() string {
	return h.url
}
