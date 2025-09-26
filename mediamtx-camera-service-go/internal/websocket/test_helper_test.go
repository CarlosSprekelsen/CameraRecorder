/*
WebSocket Test Helper - Real Component Integration Testing

Provides WebSocket test infrastructure that creates real components and validates
against the OpenRPC API specification. Uses testutils for universal patterns.

API Documentation Reference: docs/api/mediamtx_camera_service_openrpc.json
Requirements Coverage:
- REQ-WS-001: WebSocket connection and authentication
- REQ-WS-002: Real-time camera operations
- REQ-WS-003: Error handling and recovery
- REQ-WS-004: Concurrent client support
- REQ-WS-005: Session management

Design Principles:
- Real components only (no mocks)
- Fixture-driven configuration
- Progressive Readiness pattern
- OpenRPC API compliance validation
*/

package websocket

import (
	"context"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/camera"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/mediamtx"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/security"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/testutils"
)

// Global mutex to prevent parallel test execution
// WebSocket tests must run sequentially because they share the same MediaMTX controller resources
var testMutex sync.Mutex

// WebSocketTestHelper provides real WebSocket server setup for integration testing
type WebSocketTestHelper struct {
	t                  *testing.T
	setup              *testutils.UniversalTestSetup
	server             *WebSocketServer
	listener           net.Listener
	baseURL            string
	jwtHandler         *security.JWTHandler
	configManager      *config.ConfigManager
	mediaMTXController mediamtx.MediaMTXController
	logger             *logging.Logger
}

// NewWebSocketTestHelper creates a new WebSocket test helper with real components
// Follows main.go orchestration pattern exactly
// CRITICAL: Uses global mutex to prevent parallel test execution due to shared MediaMTX controller resources
func NewWebSocketTestHelper(t *testing.T) *WebSocketTestHelper {
	// CRITICAL: Lock to prevent parallel test execution
	// WebSocket tests share the same MediaMTX controller, camera devices, and file paths
	testMutex.Lock()
	t.Cleanup(func() { testMutex.Unlock() })
	// Use testutils.UniversalTestSetup for fixture-based configuration
	setup := testutils.SetupTest(t, "config_clean_minimal.yaml")
	configManager := setup.GetConfigManager()
	logger := setup.GetLogger()

	// Create real JWT handler with test secret (following main.go pattern)
	jwtHandler, err := security.NewJWTHandler("test_secret_key_for_integration_testing", logger)
	if err != nil {
		t.Fatalf("Failed to create JWT handler: %v", err)
	}

	// Create real camera monitor (following main.go orchestration pattern)
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

	// Create real MediaMTX controller (following main.go orchestration pattern)
	mediaMTXController, err := mediamtx.ControllerWithConfigManager(configManager, cameraMonitor, logger)
	if err != nil {
		t.Fatalf("Failed to create MediaMTX controller: %v", err)
	}

	helper := &WebSocketTestHelper{
		t:                  t,
		setup:              setup,
		configManager:      configManager,
		jwtHandler:         jwtHandler,
		mediaMTXController: mediaMTXController,
		logger:             logger,
	}

	// Start the MediaMTX controller and wait for readiness (following main.go exactly)
	ctx := context.Background()
	err = mediaMTXController.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start MediaMTX controller: %v", err)
	}

	// DETERMINISTIC SETUP: Wait for controller to be truly ready
	// This ensures camera monitor completes discovery before tests run
	maxWait := 30 * time.Second
	start := time.Now()
	for time.Since(start) < maxWait {
		if mediaMTXController.IsReady() {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if !mediaMTXController.IsReady() {
		t.Fatalf("MediaMTX controller not ready within %v", maxWait)
	}

	// Controller is now truly ready - tests can run deterministically
	logger.Info("MediaMTX controller started and ready - WebSocket server ready for connections")

	// Register cleanup
	t.Cleanup(func() {
		helper.Cleanup()
	})

	return helper
}

// CreateRealServer creates and starts the WebSocket server
func (h *WebSocketTestHelper) CreateRealServer() error {
	// Create real WebSocket server using the production constructor
	server, err := NewWebSocketServer(
		h.configManager,
		h.logger,
		h.jwtHandler,
		h.mediaMTXController,
	)
	if err != nil {
		return fmt.Errorf("failed to create WebSocket server: %w", err)
	}
	h.server = server

	// Start server with listener for race-free testing
	listener, err := net.Listen("tcp", ":0") // Use dynamic port
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}
	h.listener = listener

	h.baseURL = fmt.Sprintf("ws://%s%s", listener.Addr().String(), h.configManager.GetConfig().Server.WebSocketPath)

	err = h.server.StartWithListener(listener)
	if err != nil {
		return fmt.Errorf("failed to start WebSocket server with listener: %w", err)
	}

	h.t.Logf("WebSocket server started on %s", h.baseURL)
	return nil
}

// Cleanup stops the server and cleans up resources
// Uses MediaMTX cleanup pattern for proper test isolation
func (h *WebSocketTestHelper) Cleanup() {
	if h.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		h.server.Stop(ctx)
		h.t.Log("WebSocket server stopped")
	}
	if h.listener != nil {
		h.listener.Close()
	}
	if h.mediaMTXController != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Use MediaMTX cleanup pattern: force stop all active recordings
		// This follows the existing MediaMTX test helper cleanup mechanism
		if recordingManager := h.mediaMTXController.GetRecordingManager(); recordingManager != nil {
			if err := recordingManager.Cleanup(ctx); err != nil {
				h.t.Logf("Warning: Failed to cleanup recording manager: %v", err)
			} else {
				h.t.Log("Recording manager cleanup completed")
			}
		}

		// Cast to full interface to access Stop method
		if fullController, ok := h.mediaMTXController.(interface{ Stop(context.Context) error }); ok {
			fullController.Stop(ctx)
			h.t.Log("MediaMTX controller stopped")
		}
	}
	h.setup.Cleanup()
	h.t.Log("WebSocketTestHelper cleanup completed")
}

// GetServerURL returns the WebSocket server URL
func (h *WebSocketTestHelper) GetServerURL() string {
	return h.baseURL
}

// GetJWTToken creates a valid JWT token for testing
func (h *WebSocketTestHelper) GetJWTToken(role string) (string, error) {
	// Create test JWT token with specified role using real JWT handler
	token, err := h.jwtHandler.GenerateToken("test_user", role, 24)
	if err != nil {
		return "", fmt.Errorf("failed to generate JWT token: %w", err)
	}
	return token, nil
}

// GetTestCameraID returns a valid camera ID for testing
func (h *WebSocketTestHelper) GetTestCameraID() string {
	// Return a valid camera ID according to OpenRPC DeviceId pattern: ^camera[0-9]+$
	return "camera0"
}
