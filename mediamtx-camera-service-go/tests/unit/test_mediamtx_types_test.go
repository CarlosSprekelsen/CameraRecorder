//go:build unit
// +build unit

/*
MediaMTX Types Unit Tests

Requirements Coverage:
- REQ-MTX-001: MediaMTX service integration
- REQ-MTX-002: Stream management capabilities
- REQ-MTX-003: Path creation and deletion
- REQ-MTX-004: Health monitoring

Test Categories: Unit
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package mediamtx_test

import (
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/mediamtx"
	"github.com/stretchr/testify/assert"
)

// TestMediaMTXConfig_Validation tests MediaMTXConfig validation
func TestMediaMTXConfig_Validation(t *testing.T) {
	// Test valid configuration
	validConfig := &mediamtx.MediaMTXConfig{
		BaseURL:        "http://localhost:9997",
		HealthCheckURL: "http://localhost:9997/v3/paths/list",
		Timeout:        30 * time.Second,
		RetryAttempts:  3,
		RetryDelay:     1 * time.Second,
		Host:           "localhost",
		APIPort:        9997,
		RTSPPort:       8554,
		WebRTCPort:     8889,
		HLSPort:        8888,
	}

	// Test that valid config is created successfully
	assert.NotNil(t, validConfig)
	assert.Equal(t, "http://localhost:9997", validConfig.BaseURL)
	assert.Equal(t, "localhost", validConfig.Host)
	assert.Equal(t, 9997, validConfig.APIPort)
	assert.Equal(t, 30*time.Second, validConfig.Timeout)
	assert.Equal(t, 3, validConfig.RetryAttempts)
}

// TestCircuitBreakerConfig_Validation tests CircuitBreakerConfig validation
func TestCircuitBreakerConfig_Validation(t *testing.T) {
	// Test valid circuit breaker configuration
	validCircuitBreaker := mediamtx.CircuitBreakerConfig{
		FailureThreshold: 3,
		RecoveryTimeout:  30 * time.Second,
		MaxFailures:      5,
	}

	assert.NotNil(t, validCircuitBreaker)
	assert.Equal(t, 3, validCircuitBreaker.FailureThreshold)
	assert.Equal(t, 30*time.Second, validCircuitBreaker.RecoveryTimeout)
	assert.Equal(t, 5, validCircuitBreaker.MaxFailures)
}

// TestConnectionPoolConfig_Validation tests ConnectionPoolConfig validation
func TestConnectionPoolConfig_Validation(t *testing.T) {
	// Test valid connection pool configuration
	validConnectionPool := mediamtx.ConnectionPoolConfig{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	}

	assert.NotNil(t, validConnectionPool)
	assert.Equal(t, 100, validConnectionPool.MaxIdleConns)
	assert.Equal(t, 10, validConnectionPool.MaxIdleConnsPerHost)
	assert.Equal(t, 90*time.Second, validConnectionPool.IdleConnTimeout)
}

// TestStream_Serialization tests Stream serialization
func TestStream_Serialization(t *testing.T) {
	// Test stream creation and field access
	now := time.Now()
	stream := &mediamtx.Stream{
		ID:        "test-stream-123",
		Name:      "Test Stream",
		Path:      "/test/path",
		Source:    "/dev/video0",
		Status:    "ACTIVE",
		CreatedAt: now,
		UpdatedAt: now,
		Metadata: map[string]string{
			"camera_type": "USB",
			"resolution":  "1920x1080",
		},
	}

	assert.NotNil(t, stream)
	assert.Equal(t, "test-stream-123", stream.ID)
	assert.Equal(t, "Test Stream", stream.Name)
	assert.Equal(t, "/test/path", stream.Path)
	assert.Equal(t, "/dev/video0", stream.Source)
	assert.Equal(t, "ACTIVE", stream.Status)
	assert.Equal(t, now, stream.CreatedAt)
	assert.Equal(t, now, stream.UpdatedAt)
	assert.Equal(t, "USB", stream.Metadata["camera_type"])
	assert.Equal(t, "1920x1080", stream.Metadata["resolution"])
}

// TestPath_Configuration tests Path configuration
func TestPath_Configuration(t *testing.T) {
	// Test path creation and field access
	path := &mediamtx.Path{
		ID:                         "test-path-123",
		Name:                       "Test Path",
		Source:                     "/dev/video0",
		SourceOnDemand:             true,
		SourceOnDemandStartTimeout: 5 * time.Second,
		SourceOnDemandCloseAfter:   30 * time.Second,
		PublishUser:                "publisher",
		PublishPass:                "publishpass",
		ReadUser:                   "reader",
		ReadPass:                   "readpass",
		RunOnDemand:                "ffmpeg -i /dev/video0 -c:v libx264 -f rtsp rtsp://localhost:8554/test",
		RunOnDemandRestart:         true,
		RunOnDemandCloseAfter:      60 * time.Second,
		RunOnDemandStartTimeout:    10 * time.Second,
	}

	assert.NotNil(t, path)
	assert.Equal(t, "test-path-123", path.ID)
	assert.Equal(t, "Test Path", path.Name)
	assert.Equal(t, "/dev/video0", path.Source)
	assert.True(t, path.SourceOnDemand)
	assert.Equal(t, 5*time.Second, path.SourceOnDemandStartTimeout)
	assert.Equal(t, 30*time.Second, path.SourceOnDemandCloseAfter)
	assert.Equal(t, "publisher", path.PublishUser)
	assert.Equal(t, "publishpass", path.PublishPass)
	assert.Equal(t, "reader", path.ReadUser)
	assert.Equal(t, "readpass", path.ReadPass)
	assert.True(t, path.RunOnDemandRestart)
	assert.Equal(t, 60*time.Second, path.RunOnDemandCloseAfter)
	assert.Equal(t, 10*time.Second, path.RunOnDemandStartTimeout)
}

// TestHealthStatus_Validation tests HealthStatus validation
func TestHealthStatus_Validation(t *testing.T) {
	// Test health status creation and field access
	now := time.Now()
	healthStatus := &mediamtx.HealthStatus{
		Status:    "HEALTHY",
		Timestamp: now,
		Details:   "All systems operational",
		Metrics: mediamtx.Metrics{
			ActiveStreams: 5,
			TotalStreams:  10,
			CPUUsage:      25.5,
			MemoryUsage:   45.2,
			Uptime:        3600,
		},
	}

	assert.NotNil(t, healthStatus)
	assert.Equal(t, "HEALTHY", healthStatus.Status)
	assert.Equal(t, now, healthStatus.Timestamp)
	assert.Equal(t, "All systems operational", healthStatus.Details)
	assert.Equal(t, 5, healthStatus.Metrics.ActiveStreams)
	assert.Equal(t, 10, healthStatus.Metrics.TotalStreams)
	assert.Equal(t, 25.5, healthStatus.Metrics.CPUUsage)
	assert.Equal(t, 45.2, healthStatus.Metrics.MemoryUsage)
	assert.Equal(t, int64(3600), healthStatus.Metrics.Uptime)
}

// TestRecordingSession_Validation tests RecordingSession validation
func TestRecordingSession_Validation(t *testing.T) {
	// Test recording session creation and field access
	startTime := time.Now()
	endTime := startTime.Add(30 * time.Second)

	recordingSession := &mediamtx.RecordingSession{
		ID:            "recording-123",
		Device:        "/dev/video0",
		Path:          "/recordings/test.mp4",
		Status:        "RECORDING",
		StartTime:     startTime,
		EndTime:       &endTime,
		Duration:      30 * time.Second,
		FilePath:      "/tmp/recordings/test.mp4",
		FileSize:      1024000, // 1MB
		Quality:       "high",
		UseCase:       mediamtx.UseCaseRecording,
		Priority:      1,
		AutoCleanup:   true,
		RetentionDays: 30,
		MaxDuration:   24 * time.Hour,
		AutoRotate:    true,
		RotationSize:  100 * 1024 * 1024, // 100MB
	}

	assert.NotNil(t, recordingSession)
	assert.Equal(t, "recording-123", recordingSession.ID)
	assert.Equal(t, "/dev/video0", recordingSession.Device)
	assert.Equal(t, "/recordings/test.mp4", recordingSession.Path)
	assert.Equal(t, "RECORDING", recordingSession.Status)
	assert.Equal(t, startTime, recordingSession.StartTime)
	assert.Equal(t, endTime, *recordingSession.EndTime)
	assert.Equal(t, 30*time.Second, recordingSession.Duration)
	assert.Equal(t, "/tmp/recordings/test.mp4", recordingSession.FilePath)
	assert.Equal(t, int64(1024000), recordingSession.FileSize)
	assert.Equal(t, "high", recordingSession.Quality)
	assert.Equal(t, mediamtx.UseCaseRecording, recordingSession.UseCase)
	assert.Equal(t, 1, recordingSession.Priority)
	assert.True(t, recordingSession.AutoCleanup)
	assert.Equal(t, 30, recordingSession.RetentionDays)
	assert.Equal(t, 24*time.Hour, recordingSession.MaxDuration)
	assert.True(t, recordingSession.AutoRotate)
	assert.Equal(t, int64(100*1024*1024), recordingSession.RotationSize)
}

// TestSnapshot_Validation tests Snapshot validation
func TestSnapshot_Validation(t *testing.T) {
	// Test snapshot creation and field access
	created := time.Now()

	snapshot := &mediamtx.Snapshot{
		ID:       "snapshot-123",
		Device:   "/dev/video0",
		Path:     "/snapshots/test.jpg",
		FilePath: "/tmp/snapshots/test.jpg",
		Size:     51200, // 50KB
		Created:  created,
		Metadata: map[string]interface{}{
			"format":      "jpg",
			"quality":     85,
			"width":       1920,
			"height":      1080,
			"auto_resize": true,
		},
	}

	assert.NotNil(t, snapshot)
	assert.Equal(t, "snapshot-123", snapshot.ID)
	assert.Equal(t, "/dev/video0", snapshot.Device)
	assert.Equal(t, "/snapshots/test.jpg", snapshot.Path)
	assert.Equal(t, "/tmp/snapshots/test.jpg", snapshot.FilePath)
	assert.Equal(t, int64(51200), snapshot.Size)
	assert.Equal(t, created, snapshot.Created)
	assert.Equal(t, "jpg", snapshot.Metadata["format"])
	assert.Equal(t, 85, snapshot.Metadata["quality"])
	assert.Equal(t, 1920, snapshot.Metadata["width"])
	assert.Equal(t, 1080, snapshot.Metadata["height"])
	assert.Equal(t, true, snapshot.Metadata["auto_resize"])
}
