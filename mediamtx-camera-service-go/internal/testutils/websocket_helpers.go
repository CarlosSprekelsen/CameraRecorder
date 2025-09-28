/*
WebSocket Test Utilities - Minimal Extension to testutils

Extends existing testutils with WebSocket-specific helpers to eliminate
the 149 instances of duplicate WebSocket test setup without rewriting
the entire test suite.

Design Principles:
- Minimal extension of existing testutils
- Leverages existing UniversalTestSetup patterns
- Reuses existing WebSocketTestHelper logic
- No breaking changes to existing tests
*/

package testutils

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/camera"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/mediamtx"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/security"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/websocket"
)

// WebSocketTestInfrastructure provides shared WebSocket test setup
// This is a minimal wrapper around existing patterns
type WebSocketTestInfrastructure struct {
	setup              *UniversalTestSetup
	configManager      *config.ConfigManager
	logger             *logging.Logger
	jwtHandler         *security.JWTHandler
	cameraMonitor      camera.CameraMonitor
	mediaMTXController mediamtx.MediaMTXController
}

// SetupWebSocketTest creates WebSocket test infrastructure using existing patterns
// This reuses the exact logic from test_helper_test.go but makes it reusable
func SetupWebSocketTest(t *testing.T) *WebSocketTestInfrastructure {
	// Use existing testutils pattern (already working)
	setup := SetupTest(t, "config_clean_minimal.yaml")
	configManager := setup.GetConfigManager()
	logger := setup.GetLogger()

	// Create JWT handler (reuse existing pattern)
	jwtHandler, err := security.NewJWTHandler("test_secret_key_for_integration_testing", logger)
	if err != nil {
		t.Fatalf("Failed to create JWT handler: %v", err)
	}

	// Create camera monitor (reuse existing pattern)
	deviceChecker := &camera.RealDeviceChecker{}
	commandExecutor := &camera.RealV4L2CommandExecutor{}
	infoParser := &camera.RealDeviceInfoParser{}

	cameraMonitor, err := camera.NewHybridCameraMonitor(
		configManager,
		logger,
		deviceChecker,
		commandExecutor,
		infoParser,
	)
	if err != nil {
		t.Fatalf("Failed to create camera monitor: %v", err)
	}

	// Create MediaMTX controller (reuse existing pattern)
	mediaMTXController, err := mediamtx.ControllerWithConfigManager(configManager, cameraMonitor, logger)
	if err != nil {
		t.Fatalf("Failed to create MediaMTX controller: %v", err)
	}

	// Start controller with event-driven readiness (reuse existing pattern)
	ctx := context.Background()
	err = mediaMTXController.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start MediaMTX controller: %v", err)
	}

	// Event-driven readiness check (reuse existing pattern)
	if !mediaMTXController.IsReady() {
		readinessChan := mediaMTXController.SubscribeToReadiness()
		readinessCtx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)

		select {
		case <-readinessChan:
			logger.Info("Controller readiness event received - all services ready")
		case <-readinessCtx.Done():
			if !mediaMTXController.IsReady() {
				cancel()
				t.Fatalf("Controller not ready after timeout")
			}
			logger.Info("Controller ready via fallback check")
		}
		cancel()
	}

	infra := &WebSocketTestInfrastructure{
		setup:              setup,
		configManager:      configManager,
		logger:             logger,
		jwtHandler:         jwtHandler,
		cameraMonitor:      cameraMonitor,
		mediaMTXController: mediaMTXController,
	}

	// Register cleanup (reuse existing pattern)
	t.Cleanup(func() {
		infra.Cleanup()
	})

	return infra
}

// CreateWebSocketServer creates and starts a WebSocket server
// This reuses the exact logic from test_helper_test.go
func (w *WebSocketTestInfrastructure) CreateWebSocketServer(t *testing.T) (*websocket.WebSocketServer, net.Listener, string) {
	// Create WebSocket server (reuse existing pattern)
	server, err := websocket.NewWebSocketServer(
		w.configManager,
		w.logger,
		w.jwtHandler,
		w.mediaMTXController,
	)
	if err != nil {
		t.Fatalf("Failed to create WebSocket server: %v", err)
	}

	// Start server with listener (reuse existing pattern)
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}

	baseURL := fmt.Sprintf("ws://%s%s", listener.Addr().String(), w.configManager.GetConfig().Server.WebSocketPath)

	err = server.StartWithListener(listener)
	if err != nil {
		t.Fatalf("Failed to start WebSocket server: %v", err)
	}

	w.logger.Infof("WebSocket server started on %s", baseURL)
	return server, listener, baseURL
}

// GetJWTToken creates a JWT token for testing
func (w *WebSocketTestInfrastructure) GetJWTToken(role string) (string, error) {
	return w.jwtHandler.GenerateToken("test_user", role, 24)
}

// GetTestCameraID returns a test camera ID
func (w *WebSocketTestInfrastructure) GetTestCameraID() string {
	return GetTestCameraID()
}

// GetCameraMonitor returns the camera monitor
func (w *WebSocketTestInfrastructure) GetCameraMonitor() camera.CameraMonitor {
	return w.cameraMonitor
}

// Cleanup performs cleanup (reuse existing pattern)
func (w *WebSocketTestInfrastructure) Cleanup() {
	// Stop MediaMTX controller (reuse existing pattern)
	if w.mediaMTXController != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if recordingManager := w.mediaMTXController.GetRecordingManager(); recordingManager != nil {
			recordingManager.Cleanup(ctx)
		}

		if fullController, ok := w.mediaMTXController.(interface{ Stop(context.Context) error }); ok {
			fullController.Stop(ctx)
		}
	}

	// Cleanup test setup (reuse existing pattern)
	if w.setup != nil {
		w.setup.Cleanup()
	}
}

// CreateWebSocketTestClient creates a WebSocket test client
// This provides a simple helper to create clients
func CreateWebSocketTestClient(t *testing.T, serverURL string) *websocket.WebSocketTestClient {
	return websocket.NewWebSocketTestClient(t, serverURL)
}

// AuthenticateClient provides a simple authentication helper
func AuthenticateClient(t *testing.T, client *websocket.WebSocketTestClient, infra *WebSocketTestInfrastructure, role string) {
	token, err := infra.GetJWTToken(role)
	if err != nil {
		t.Fatalf("Failed to create JWT token: %v", err)
	}

	err = client.Authenticate(token)
	if err != nil {
		t.Fatalf("Authentication failed: %v", err)
	}
}
