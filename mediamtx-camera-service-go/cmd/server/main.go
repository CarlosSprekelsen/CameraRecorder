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
	configManager := config.NewConfigManager()
	if err := configManager.LoadConfig("config/default.yaml"); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Setup logging
	logger := logging.NewLogger("camera-service")
	logger.Info("Starting MediaMTX Camera Service (Go)")

	// Initialize real implementations for camera monitor dependencies
	deviceChecker := &camera.RealDeviceChecker{}
	commandExecutor := &camera.RealV4L2CommandExecutor{}
	infoParser := &camera.RealDeviceInfoParser{}

	// Initialize camera monitor with real implementations
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

	// Initialize MediaMTX controller with existing logger
	mediaMTXController, err := mediamtx.ControllerWithConfigManager(configManager, logger.Logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to create MediaMTX controller")
	}

	// Get configuration
	cfg := configManager.GetConfig()
	if cfg == nil {
		logger.Fatal("Configuration not available")
	}

	// Initialize JWT handler with configuration
	jwtHandler, err := security.JWTHandler(cfg.Security.JWTSecretKey)
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
		cameraMonitor,
		jwtHandler,
		mediaMTXController,
	)
	if err != nil {
		logger.WithError(err).Fatal("Failed to create WebSocket server")
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

	// Graceful shutdown
	_, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := wsServer.Stop(); err != nil {
		logger.WithError(err).Error("Error stopping WebSocket server")
	}

	if err := cameraMonitor.Stop(); err != nil {
		logger.WithError(err).Error("Error stopping camera monitor")
	}

	logger.Info("Camera service stopped")
}
