// Package main implements the MediaMTX Camera Service entry point.
//
// This service provides real-time video sensor management with streaming,
// recording, and snapshot capabilities. It operates as a containerized service
// that manages USB V4L2 cameras and external RTSP feeds within a coordinated
// sensor ecosystem.
//
// Architecture follows the layered approach:
//   - Foundation: Configuration and logging
//   - Core Services: MediaMTX client and camera monitoring
//   - Managers: Path, stream, and recording management
//   - Business Logic: Recording and snapshot orchestration
//   - Orchestration: MediaMTX controller coordination
//   - API: WebSocket JSON-RPC 2.0 server
//
// The startup sequence follows architectural compliance:
// 1. Load and validate configuration
// 2. Initialize logging with structured output
// 3. Create camera monitor with real hardware interfaces
// 4. Initialize MediaMTX controller (single source of truth)
// 5. Setup security framework with JWT authentication
// 6. Start WebSocket server (protocol layer only)
// 7. Connect event notification system
//
// Graceful shutdown reverses the startup order to ensure clean resource cleanup.
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
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/health"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/mediamtx"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/security"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/websocket"
)

// main implements the application entry point following the progressive readiness pattern.
// The system accepts connections immediately while features become available as components initialize.
func main() {
	// Layer 1: Foundation - Load and validate configuration
	configManager := config.CreateConfigManager()
	if err := configManager.LoadConfig("config/default.yaml"); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	cfg := configManager.GetConfig()
	if cfg == nil {
		log.Fatalf("Configuration not available")
	}

	// Validate path configuration at startup to prevent runtime errors
	if err := config.ValidatePathConfiguration(cfg); err != nil {
		log.Fatalf("Path configuration validation failed: %v", err)
	}

	// Initialize structured logging with JSON formatting for production
	_ = logging.SetupLogging(&logging.LoggingConfig{
		Level:          cfg.Logging.Level,
		Format:         cfg.Logging.Format,
		FileEnabled:    cfg.Logging.FileEnabled,
		FilePath:       cfg.Logging.FilePath,
		MaxFileSize:    int(cfg.Logging.MaxFileSize),
		BackupCount:    cfg.Logging.BackupCount,
		ConsoleEnabled: cfg.Logging.ConsoleEnabled,
	})

	// Enable dynamic logging configuration updates without restart
	configManager.RegisterLoggingConfigurationUpdates()

	logger := logging.GetLogger("camera-service")
	logger.Info("Starting MediaMTX Camera Service (Go)")

	// Initialize runtime path validation for MediaMTX configuration
	pathValidator := mediamtx.NewPathValidator(cfg, logger)

	// Start continuous path validation to detect configuration drift
	ctx := context.Background()
	pathValidator.StartPeriodicValidation(ctx)

	// Layer 2: Core Services - Initialize hardware abstraction layer
	// Use real implementations for production hardware access
	deviceChecker := &camera.RealDeviceChecker{}
	commandExecutor := &camera.RealV4L2CommandExecutor{}
	infoParser := &camera.RealDeviceInfoParser{}

	// Initialize camera monitor with event-driven device discovery
	// Uses udev/fsnotify for real-time device lifecycle events with polling fallback
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

	// Layer 5: Orchestration - Initialize MediaMTX controller as single source of truth
	// Controller coordinates all managers and provides API abstraction (camera0 ↔ /dev/video0)
	mediaMTXController, err := mediamtx.ControllerWithConfigManager(configManager, cameraMonitor, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to create MediaMTX controller")
	}

	// Security Framework - Initialize JWT authentication with role-based access control
	jwtHandler, err := security.NewJWTHandler(cfg.Security.JWTSecretKey, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to create JWT handler")
	}

	// Configure rate limiting for security protection
	if cfg.Security.RateLimitRequests > 0 {
		jwtHandler.SetRateLimit(int64(cfg.Security.RateLimitRequests), cfg.Security.RateLimitWindow)
	}

	// Layer 6: API - Initialize WebSocket server (protocol layer only, no business logic)
	// Server delegates all operations to MediaMTX controller following architectural constraints
	wsServer, err := websocket.NewWebSocketServer(
		configManager,
		logger,
		jwtHandler,
		mediaMTXController,
	)
	if err != nil {
		logger.WithError(err).Fatal("Failed to create WebSocket server")
	}

	// Initialize HTTP Health Server for container orchestration
	var httpHealthServer *health.HTTPHealthServer
	if cfg.HTTPHealth.Enabled {
		// Create health monitor for health API
		healthMonitor := health.NewHealthMonitor("1.0.0")
		
		// Create HTTP health server with thin delegation pattern
		httpHealthServer, err = health.NewHTTPHealthServer(&cfg.HTTPHealth, healthMonitor, logger)
		if err != nil {
			logger.WithError(err).Fatal("Failed to create HTTP health server")
		}
		logger.Info("HTTP Health Server initialized")
	}

	// Event System Integration - Connect MediaMTX controller to WebSocket event notifications
	// Controller implements DeviceToCameraIDMapper for API abstraction layer
	mediaMTXEventNotifier := websocket.NewMediaMTXEventNotifier(wsServer.GetEventManager(), mediaMTXController, logger)

	// Connect event notifier to controller for real-time client notifications
	if setterController, ok := mediaMTXController.(interface {
		SetEventNotifier(mediamtx.MediaMTXEventNotifier)
	}); ok {
		setterController.SetEventNotifier(mediaMTXEventNotifier)
	}

	// Initialize system-level event notifications for service lifecycle
	systemEventNotifier := websocket.NewSystemEventNotifier(wsServer.GetEventManager(), logger)
	systemEventNotifier.NotifySystemStartup("1.0.0", "Go implementation")

	// Connect system events to controller for unified health monitoring
	if controllerWithNotifier, ok := mediaMTXController.(interface {
		SetSystemEventNotifier(notifier mediamtx.SystemEventNotifier)
	}); ok {
		controllerWithNotifier.SetSystemEventNotifier(systemEventNotifier)
		logger.Info("Connected SystemEventNotifier to controller for unified health notifications")
	}

	// Service Startup - Follow architectural compliance with controller orchestration
	logger.Info("Starting MediaMTX Controller orchestration...")

	// Start controller first - it orchestrates all managed services (camera monitor, managers)
	if err := mediaMTXController.Start(ctx); err != nil {
		logger.WithError(err).Fatal("Failed to start MediaMTX controller")
	}
	logger.Info("MediaMTX Controller started successfully")

	// Progressive Readiness Pattern - Wait for controller readiness using event-driven approach
	logger.Info("Waiting for controller readiness...")
	readinessChan := mediaMTXController.SubscribeToReadiness()

	// Apply readiness timeout to prevent indefinite blocking
	readinessCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	select {
	case <-readinessChan:
		logger.Info("Controller readiness event received - all services ready")
	case <-readinessCtx.Done():
		logger.Warn("Controller readiness timeout - proceeding anyway")
	}

	// Verify actual readiness state from controller
	if mediaMTXController.IsReady() {
		logger.Info("Controller reports ready - all services operational")
	} else {
		logger.Warn("Controller not ready - some services may not be operational")
	}

	// Start WebSocket server after controller readiness (accepts connections immediately)
	logger.Info("Starting WebSocket server...")
	if err := wsServer.Start(); err != nil {
		logger.WithError(err).Fatal("Failed to start WebSocket server")
	}
	logger.Info("WebSocket server started successfully")

	// Start HTTP Health Server for container orchestration
	if httpHealthServer != nil {
		logger.Info("Starting HTTP Health Server...")
		if err := httpHealthServer.Start(ctx); err != nil {
			logger.WithError(err).Fatal("Failed to start HTTP Health Server")
		}
		logger.Info("HTTP Health Server started successfully")
	}

	logger.Info("Camera service started successfully - all components operational")

	// Graceful Shutdown - Wait for termination signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	logger.Info("Received shutdown signal, stopping services...")

	// Configure graceful shutdown timeout from configuration
	shutdownTimeout := 30 * time.Second // Default fallback
	if cfg.ServerDefaults.ShutdownTimeout > 0 {
		shutdownTimeout = time.Duration(cfg.ServerDefaults.ShutdownTimeout * float64(time.Second))
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	logger.Info("Starting graceful shutdown...")

	// Reverse Shutdown Order - Stop services in reverse order of startup for clean resource cleanup
	var wg sync.WaitGroup
	errorChan := make(chan error, 4)

	// Stop HTTP Health Server first - prevents new health check connections
	if httpHealthServer != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			logger.Info("Stopping HTTP Health Server...")
			if err := httpHealthServer.Stop(); err != nil {
				logger.WithError(err).Error("Error stopping HTTP Health Server")
				errorChan <- err
			} else {
				logger.Info("HTTP Health Server stopped successfully")
			}
		}()
	}

	// Stop WebSocket server - prevents new client connections
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

	// Stop MediaMTX controller - orchestrates shutdown of all managed services
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

	// Stop camera monitor - ensure hardware resources are properly released
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

	// Wait for all services to stop gracefully with timeout protection
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
		os.Exit(1) // Force exit on timeout to prevent hanging
	}

	// Collect and report shutdown errors for monitoring
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
