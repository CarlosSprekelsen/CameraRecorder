package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"mediamtx-camera-service-go/internal/config"
)

func main() {
	// Set up logging
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	// Create a temporary config file for demonstration
	tempDir := os.TempDir()
	configPath := filepath.Join(tempDir, "hot-reload-demo.yaml")

	// Write initial configuration
	initialConfig := `
server:
  host: "localhost"
  port: 8080
  websocket_path: "/ws"
  max_connections: 100

camera:
  port: 8081
  streams: ["main", "sub"]
  frame_rate: 30
  resolution:
    width: 1920
    height: 1080

logging:
  level: "info"
  output_path: "/tmp/logs"
  max_file_size: 10485760
  max_files: 5
`

	err := os.WriteFile(configPath, []byte(initialConfig), 0644)
	if err != nil {
		log.Fatalf("Failed to create initial config file: %v", err)
	}

	logrus.Infof("Created initial config file at: %s", configPath)

	// Load initial configuration
	loader := config.NewConfigLoader()
	cfg, err := loader.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load initial configuration: %v", err)
	}

	logrus.Infof("Initial configuration loaded - Server: %s:%d", cfg.Server.Host, cfg.Server.Port)

	// Create configuration watcher
	reloadCallback := func(newConfig *config.Config) error {
		logrus.Infof("Configuration reloaded - Server: %s:%d", newConfig.Server.Host, newConfig.Server.Port)
		logrus.Infof("Camera streams: %v", newConfig.Camera.Streams)
		logrus.Infof("Logging level: %s", newConfig.Logging.Level)
		return nil
	}

	watcher, err := config.NewConfigWatcher(configPath, reloadCallback)
	if err != nil {
		log.Fatalf("Failed to create config watcher: %v", err)
	}

	// Start watching for configuration changes
	err = watcher.Start()
	if err != nil {
		log.Fatalf("Failed to start config watcher: %v", err)
	}

	logrus.Info("Configuration hot reload started. Press Ctrl+C to exit.")

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Simulate configuration changes
	go func() {
		time.Sleep(3 * time.Second)

		// First change: Update server port and logging level
		updatedConfig := `
server:
  host: "localhost"
  port: 9090
  websocket_path: "/ws"
  max_connections: 100

camera:
  port: 8081
  streams: ["main", "sub"]
  frame_rate: 30
  resolution:
    width: 1920
    height: 1080

logging:
  level: "debug"
  output_path: "/tmp/logs"
  max_file_size: 10485760
  max_files: 5
`
		logrus.Info("Updating configuration (port: 8080 -> 9090, log level: info -> debug)")
		err := os.WriteFile(configPath, []byte(updatedConfig), 0644)
		if err != nil {
			logrus.Errorf("Failed to update config file: %v", err)
		}

		time.Sleep(3 * time.Second)

		// Second change: Update camera streams
		updatedConfig2 := `
server:
  host: "localhost"
  port: 9090
  websocket_path: "/ws"
  max_connections: 100

camera:
  port: 8081
  streams: ["main", "sub", "preview"]
  frame_rate: 30
  resolution:
    width: 1920
    height: 1080

logging:
  level: "debug"
  output_path: "/tmp/logs"
  max_file_size: 10485760
  max_files: 5
`
		logrus.Info("Updating configuration (adding 'preview' stream)")
		err = os.WriteFile(configPath, []byte(updatedConfig2), 0644)
		if err != nil {
			logrus.Errorf("Failed to update config file: %v", err)
		}

		time.Sleep(3 * time.Second)

		// Third change: Introduce invalid configuration
		invalidConfig := `
server:
  host: ""
  port: 99999
  websocket_path: "/ws"
  max_connections: 100

camera:
  port: 8081
  streams: []
  frame_rate: 30
  resolution:
    width: 1920
    height: 1080

logging:
  level: "invalid_level"
  output_path: "/tmp/logs"
  max_file_size: 10485760
  max_files: 5
`
		logrus.Info("Updating configuration with invalid values (should trigger validation errors)")
		err = os.WriteFile(configPath, []byte(invalidConfig), 0644)
		if err != nil {
			logrus.Errorf("Failed to update config file: %v", err)
		}

		time.Sleep(3 * time.Second)

		// Fourth change: Restore valid configuration
		restoredConfig := `
server:
  host: "localhost"
  port: 8080
  websocket_path: "/ws"
  max_connections: 100

camera:
  port: 8081
  streams: ["main", "sub"]
  frame_rate: 30
  resolution:
    width: 1920
    height: 1080

logging:
  level: "info"
  output_path: "/tmp/logs"
  max_file_size: 10485760
  max_files: 5
`
		logrus.Info("Restoring valid configuration")
		err = os.WriteFile(configPath, []byte(restoredConfig), 0644)
		if err != nil {
			logrus.Errorf("Failed to update config file: %v", err)
		}
	}()

	// Wait for shutdown signal
	<-sigChan
	logrus.Info("Shutdown signal received, stopping config watcher...")

	// Stop the watcher
	err = watcher.Stop()
	if err != nil {
		logrus.Errorf("Error stopping config watcher: %v", err)
	}

	// Clean up temporary file
	err = os.Remove(configPath)
	if err != nil {
		logrus.Errorf("Error removing temporary config file: %v", err)
	}

	logrus.Info("Hot reload example completed successfully")
}
