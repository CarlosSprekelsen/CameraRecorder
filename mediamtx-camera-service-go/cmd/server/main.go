package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"sync"
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

	cfg := configManager.GetConfig()
	if cfg == nil {
		log.Fatalf("Configuration not available")
	}

	// CRITICAL: Validate paths at startup
	if err := config.ValidatePathConfiguration(cfg); err != nil {
		log.Fatalf("Path configuration validation failed: %v", err)
	}

	// Setup logging from validated config
	_ = logging.SetupLogging(&logging.LoggingConfig{
		Level:          cfg.Logging.Level,
		Format:         cfg.Logging.Format,
		FileEnabled:    cfg.Logging.FileEnabled,
		FilePath:       cfg.Logging.FilePath,
		MaxFileSize:    int(cfg.Logging.MaxFileSize),
		BackupCount:    cfg.Logging.BackupCount,
		ConsoleEnabled: cfg.Logging.ConsoleEnabled,
	})

	// Register automatic logging configuration updates
	configManager.RegisterLoggingConfigurationUpdates()

	logger := logging.GetLogger("camera-service")
	logger.Info("Starting MediaMTX Camera Service (Go)")

	// Initialize path validator for runtime checks
	pathValidator := mediamtx.NewPathValidator(cfg, logger)

	// Start periodic validation
	ctx := context.Background()
	pathValidator.StartPeriodicValidation(ctx)

	// Initialize real implementations for camera monitor dependencies
	deviceChecker := &camera.RealDeviceChecker{}
	commandExecutor := &camera.RealV4L2CommandExecutor{}
	infoParser := &camera.RealDeviceInfoParser{}

	// Initialize camera monitor with real implementations
	// Monitor will acquire its own device event source reference
	cameraMonitor, err := camera.NewHybridCameraMonitor(
		configManager,
		logger,
		deviceChecker,
		commandExecutor,
		infoParser,
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

	// ARCHITECTURAL COMPLIANCE: Start MediaMTX Controller first (orchestrates all services)
	logger.Info("Starting MediaMTX Controller orchestration...")

	// Start the MediaMTX controller (this orchestrates all other services)
	if err := mediaMTXController.Start(ctx); err != nil {
		logger.WithError(err).Fatal("Failed to start MediaMTX controller")
	}
	logger.Info("MediaMTX Controller started successfully")

	// Wait for controller readiness using event-driven approach
	logger.Info("Waiting for controller readiness...")
	readinessChan := mediaMTXController.SubscribeToReadiness()

	// Wait for readiness event with timeout
	readinessCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	select {
	case <-readinessChan:
		logger.Info("Controller readiness event received - all services ready")
	case <-readinessCtx.Done():
		logger.Warn("Controller readiness timeout - proceeding anyway")
	}

	// Verify controller is actually ready
	if mediaMTXController.IsReady() {
		logger.Info("Controller reports ready - all services operational")
	} else {
		logger.Warn("Controller not ready - some services may not be operational")
	}

	// Start WebSocket server (after controller is ready)
	logger.Info("Starting WebSocket server...")
	if err := wsServer.Start(); err != nil {
		logger.WithError(err).Fatal("Failed to start WebSocket server")
	}
	logger.Info("WebSocket server started successfully")

	logger.Info("Camera service started successfully - all components operational")

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
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	logger.Info("Starting graceful shutdown...")

	// ARCHITECTURAL COMPLIANCE: Stop services in reverse order of startup
	var wg sync.WaitGroup
	errorChan := make(chan error, 3)

	// Stop WebSocket server first (stops accepting new connections)
	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Info("Stopping WebSocket server...")
		if err := wsServer.Stop(shutdownCtx); err != nil {
			logger.WithError(err).Error("Error stopping WebSocket server")
			errorChan <- err
		} else {
			logger.Info("WebSocket server stopped successfully")
		}
	}()

	// Stop MediaMTX controller (orchestrates shutdown of all managed services)
	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Info("Stopping MediaMTX controller...")
		if err := mediaMTXController.Stop(shutdownCtx); err != nil {
			logger.WithError(err).Error("Error stopping MediaMTX controller")
			errorChan <- err
		} else {
			logger.Info("MediaMTX controller stopped successfully")
		}
	}()

	// Stop camera monitor (managed by controller, but ensure cleanup)
	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Info("Stopping camera monitor...")
		if err := cameraMonitor.Stop(shutdownCtx); err != nil {
			logger.WithError(err).Error("Error stopping camera monitor")
			errorChan <- err
		} else {
			logger.Info("Camera monitor stopped successfully")
		}
	}()

	// Wait for all services to stop with overall timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logger.Info("All services stopped cleanly")
	case <-shutdownCtx.Done():
		logger.Error("Shutdown timeout - forcing exit")
		os.Exit(1) // Force exit on timeout
	}

	// Check for errors
	close(errorChan)
	var errors []error
	for err := range errorChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		logger.WithField("error_count", strconv.Itoa(len(errors))).Error("Some services failed to stop cleanly")
	}

	logger.Info("Camera service stopped")
}
