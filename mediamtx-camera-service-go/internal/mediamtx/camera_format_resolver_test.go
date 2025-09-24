package mediamtx

import (
	"testing"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
)

func TestCameraFormatResolver_GetOptimalFormatForDevice(t *testing.T) {
	logger := logging.GetLogger("test")

	t.Run("no_camera_monitor", func(t *testing.T) {
		resolver := NewCameraFormatResolver(nil, logger)

		format, err := resolver.GetOptimalFormatForDevice("/dev/video0")

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if format != "YUYV" {
			t.Errorf("Expected YUYV fallback, got %s", format)
		}
	})

	t.Run("non_v4l2_device", func(t *testing.T) {
		resolver := NewCameraFormatResolver(nil, logger)

		format, err := resolver.GetOptimalFormatForDevice("rtsp://192.168.1.100/stream")

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if format != "YUYV" {
			t.Errorf("Expected YUYV fallback, got %s", format)
		}
	})
}

func TestCameraFormatResolver_GetOptimalFormatForDeviceWithFallback(t *testing.T) {
	logger := logging.GetLogger("test")
	resolver := NewCameraFormatResolver(nil, logger)

	format, err := resolver.GetOptimalFormatForDeviceWithFallback("/dev/video0", "yuv420p")

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if format != "yuv420p" {
		t.Errorf("Expected yuv420p fallback, got %s", format)
	}
}
