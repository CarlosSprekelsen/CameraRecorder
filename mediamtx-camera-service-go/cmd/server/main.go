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

	// Initialize camera monitor with default implementations
	cameraMonitor := camera.NewHybridCameraMonitor(
		configManager,
		logger,
		nil, // Will use default implementations
		nil, // Will use default implementations
		nil, // Will use default implementations
	)

	// Initialize MediaMTX controller
	mediaMTXController, err := mediamtx.NewControllerWithConfigManager(configManager, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to create MediaMTX controller")
	}

	// Initialize JWT handler
	jwtHandler, err := security.NewJWTHandler("default-secret-key-change-in-production")
	if err != nil {
		logger.WithError(err).Fatal("Failed to create JWT handler")
	}

	// Initialize WebSocket server
	wsServer := websocket.NewWebSocketServer(
		configManager,
		logger,
		cameraMonitor,
		jwtHandler,
		mediaMTXController,
	)

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
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := wsServer.Stop(); err != nil {
		logger.WithError(err).Error("Error stopping WebSocket server")
	}

	if err := cameraMonitor.Stop(); err != nil {
		logger.WithError(err).Error("Error stopping camera monitor")
	}

	logger.Info("Camera service stopped")
}
