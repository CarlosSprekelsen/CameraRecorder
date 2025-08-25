package main

import (
	"fmt"
	"log"
	"os"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
)

func main() {
	// Create a new configuration loader
	loader := config.NewConfigLoader()
	
	// Get config file path from command line argument or use default
	configPath := "config/default.yaml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}
	
	// Load configuration
	cfg, err := loader.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	
	// Print configuration summary
	fmt.Println("=== MediaMTX Camera Service Configuration ===")
	fmt.Printf("Server: %s:%d\n", cfg.Server.Host, cfg.Server.Port)
	fmt.Printf("WebSocket Path: %s\n", cfg.Server.WebSocketPath)
	fmt.Printf("Max Connections: %d\n", cfg.Server.MaxConnections)
	
	fmt.Printf("\nMediaMTX:\n")
	fmt.Printf("  Host: %s\n", cfg.MediaMTX.Host)
	fmt.Printf("  API Port: %d\n", cfg.MediaMTX.APIPort)
	fmt.Printf("  RTSP Port: %d\n", cfg.MediaMTX.RTSPPort)
	fmt.Printf("  WebRTC Port: %d\n", cfg.MediaMTX.WebRTCPort)
	fmt.Printf("  HLS Port: %d\n", cfg.MediaMTX.HLSPort)
	fmt.Printf("  Recordings Path: %s\n", cfg.MediaMTX.RecordingsPath)
	fmt.Printf("  Snapshots Path: %s\n", cfg.MediaMTX.SnapshotsPath)
	
	fmt.Printf("\nCamera:\n")
	fmt.Printf("  Poll Interval: %.2f seconds\n", cfg.Camera.PollInterval)
	fmt.Printf("  Detection Timeout: %.2f seconds\n", cfg.Camera.DetectionTimeout)
	fmt.Printf("  Device Range: %v\n", cfg.Camera.DeviceRange)
	fmt.Printf("  Enable Capability Detection: %t\n", cfg.Camera.EnableCapabilityDetection)
	fmt.Printf("  Auto Start Streams: %t\n", cfg.Camera.AutoStartStreams)
	
	fmt.Printf("\nLogging:\n")
	fmt.Printf("  Level: %s\n", cfg.Logging.Level)
	fmt.Printf("  File Enabled: %t\n", cfg.Logging.FileEnabled)
	fmt.Printf("  Console Enabled: %t\n", cfg.Logging.ConsoleEnabled)
	if cfg.Logging.FileEnabled {
		fmt.Printf("  File Path: %s\n", cfg.Logging.FilePath)
	}
	
	fmt.Printf("\nRecording:\n")
	fmt.Printf("  Enabled: %t\n", cfg.Recording.Enabled)
	fmt.Printf("  Auto Record: %t\n", cfg.Recording.AutoRecord)
	fmt.Printf("  Format: %s\n", cfg.Recording.Format)
	fmt.Printf("  Quality: %s\n", cfg.Recording.Quality)
	
	fmt.Printf("\nSnapshots:\n")
	fmt.Printf("  Enabled: %t\n", cfg.Snapshots.Enabled)
	fmt.Printf("  Format: %s\n", cfg.Snapshots.Format)
	fmt.Printf("  Quality: %d\n", cfg.Snapshots.Quality)
	fmt.Printf("  Max Width: %d\n", cfg.Snapshots.MaxWidth)
	fmt.Printf("  Max Height: %d\n", cfg.Snapshots.MaxHeight)
	
	fmt.Printf("\nSTANAG 4406 Codec Settings:\n")
	fmt.Printf("  Codec: %s\n", cfg.MediaMTX.Codec)
	fmt.Printf("  Video Profile: %s\n", cfg.MediaMTX.VideoProfile)
	fmt.Printf("  Video Level: %s\n", cfg.MediaMTX.VideoLevel)
	fmt.Printf("  Pixel Format: %s\n", cfg.MediaMTX.PixelFormat)
	fmt.Printf("  Bitrate: %s\n", cfg.MediaMTX.Bitrate)
	fmt.Printf("  Preset: %s\n", cfg.MediaMTX.Preset)
	
	fmt.Printf("\nPerformance Targets:\n")
	fmt.Printf("  Snapshot Capture: %.2f seconds\n", cfg.Performance.ResponseTimeTargets.SnapshotCapture)
	fmt.Printf("  Recording Start: %.2f seconds\n", cfg.Performance.ResponseTimeTargets.RecordingStart)
	fmt.Printf("  Recording Stop: %.2f seconds\n", cfg.Performance.ResponseTimeTargets.RecordingStop)
	fmt.Printf("  File Listing: %.2f seconds\n", cfg.Performance.ResponseTimeTargets.FileListing)
	
	fmt.Println("\n=== Configuration loaded successfully ===")
}
