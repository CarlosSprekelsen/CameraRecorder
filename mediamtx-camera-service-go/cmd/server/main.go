package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/camera"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/mediamtx"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/security"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/websocket"
)

func main() {
	// Load configuration
	configManager := config.CreateConfigManager()
	if err := configManager.LoadConfig("config/default.yaml"); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Setup logging from validated config
	cfg := configManager.GetConfig()
	if cfg == nil {
		log.Fatalf("Configuration not available")
	}
	_ = logging.SetupLogging(&logging.LoggingConfig{
		Level:          cfg.Logging.Level,
		Format:         cfg.Logging.Format,
		FileEnabled:    cfg.Logging.FileEnabled,
		FilePath:       cfg.Logging.FilePath,
		MaxFileSize:    int(cfg.Logging.MaxFileSize),
		BackupCount:    cfg.Logging.BackupCount,
		ConsoleEnabled: cfg.Logging.ConsoleEnabled,
	})
	logger := logging.GetLogger("camera-service")
	logger.Info("Starting MediaMTX Camera Service (Go)")

	// Initialize real implementations for camera monitor dependencies
	deviceChecker := &camera.RealDeviceChecker{}
	commandExecutor := &camera.RealV4L2CommandExecutor{}
	infoParser := &camera.RealDeviceInfoParser{}

	// Create device event source (default to fsnotify)
	deviceEventSource, err := camera.NewFsnotifyDeviceEventSource(logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to create device event source")
	}

	// Initialize camera monitor with real implementations
	cameraMonitor, err := camera.NewHybridCameraMonitor(
		configManager,
		logger,
		deviceChecker,
		commandExecutor,
		infoParser,
		deviceEventSource,
	)
	if err != nil {
		logger.WithError(err).Fatal("Failed to create camera monitor")
	}

	// Initialize MediaMTX controller first (without event notifier for now)
	mediaMTXController, err := mediamtx.ControllerWithConfigManager(configManager, cameraMonitor, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to create MediaMTX controller")
	}

	// Get configuration (already loaded above)

	// Initialize JWT handler with configuration
	jwtHandler, err := security.NewJWTHandler(cfg.Security.JWTSecretKey, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to create JWT handler")
	}

	// Update rate limiting configuration if specified
	if cfg.Security.RateLimitRequests > 0 {
		jwtHandler.SetRateLimit(int64(cfg.Security.RateLimitRequests), cfg.Security.RateLimitWindow)
	}

	// Initialize WebSocket server
	wsServer, err := websocket.NewWebSocketServer(
		configManager,
		logger,
		jwtHandler,
		mediaMTXController,
	)
	if err != nil {
		logger.WithError(err).Fatal("Failed to create WebSocket server")
	}

	// Connect MediaMTX controller to event system (MediaMTX manages camera events internally)
	// MediaMTX Controller implements DeviceToCameraIDMapper interface for abstraction
	mediaMTXEventNotifier := websocket.NewMediaMTXEventNotifier(wsServer.GetEventManager(), mediaMTXController, logger)

	// Connect the event notifier to MediaMTX controller
	// Note: SetEventNotifier method needs to be added to MediaMTXController interface
	if setterController, ok := mediaMTXController.(interface {
		SetEventNotifier(mediamtx.MediaMTXEventNotifier)
	}); ok {
		setterController.SetEventNotifier(mediaMTXEventNotifier)
	}

	// Connect system events to event system
	systemEventNotifier := websocket.NewSystemEventNotifier(wsServer.GetEventManager(), logger)
	systemEventNotifier.NotifySystemStartup("1.0.0", "Go implementation")

	// Connect SystemEventNotifier to controller for unified health notifications
	if controllerWithNotifier, ok := mediaMTXController.(interface {
		SetSystemEventNotifier(notifier mediamtx.SystemEventNotifier)
	}); ok {
		controllerWithNotifier.SetSystemEventNotifier(systemEventNotifier)
		logger.Info("Connected SystemEventNotifier to controller for unified health notifications")
	}

	// Start camera monitor
	ctx := context.Background()
	if err := cameraMonitor.Start(ctx); err != nil {
		logger.WithError(err).Fatal("Failed to start camera monitor")
	}

	// Start WebSocket server
	if err := wsServer.Start(); err != nil {
		logger.WithError(err).Fatal("Failed to start WebSocket server")
	}

	logger.Info("Camera service started successfully")

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	logger.Info("Received shutdown signal, stopping services...")

	// Graceful shutdown with configurable timeout
	shutdownTimeout := 30 * time.Second // Default fallback
	if cfg.ServerDefaults.ShutdownTimeout > 0 {
		shutdownTimeout = time.Duration(cfg.ServerDefaults.ShutdownTimeout * float64(time.Second))
	}
	_, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := wsServer.Stop(); err != nil {
		logger.WithError(err).Error("Error stopping WebSocket server")
	}

	if err := cameraMonitor.Stop(); err != nil {
		logger.WithError(err).Error("Error stopping camera monitor")
	}

	logger.Info("Camera service stopped")
}
