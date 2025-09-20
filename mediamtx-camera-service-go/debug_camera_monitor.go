package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/camera"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
)

func main() {
	logger := logging.CreateTestLogger(nil, nil)
	cameraMonitor := camera.NewHybridMonitor(logger)
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	// Start monitor
	err := cameraMonitor.Start(ctx)
	if err != nil {
		log.Fatalf("Failed to start camera monitor: %v", err)
	}
	
	// Wait for readiness
	for i := 0; i < 10 && !cameraMonitor.IsReady(); i++ {
		fmt.Printf("Waiting for camera monitor readiness... (%d/10)\n", i+1)
		time.Sleep(500 * time.Millisecond)
	}
	
	if !cameraMonitor.IsReady() {
		log.Fatalf("Camera monitor not ready after 5 seconds")
	}
	
	fmt.Println("Camera monitor is ready!")
	
	// Test GetDevice for both video devices
	devices := []string{"/dev/video0", "/dev/video1"}
	for _, devicePath := range devices {
		device, exists := cameraMonitor.GetDevice(devicePath)
		fmt.Printf("GetDevice('%s'): exists=%v, device=%+v\n", devicePath, exists, device)
	}
	
	// Also check GetConnectedCameras
	connectedCameras := cameraMonitor.GetConnectedCameras()
	fmt.Printf("GetConnectedCameras(): %d cameras found\n", len(connectedCameras))
	for path, camera := range connectedCameras {
		fmt.Printf("  - %s: %+v\n", path, camera)
	}
	
	cameraMonitor.Stop()
}
