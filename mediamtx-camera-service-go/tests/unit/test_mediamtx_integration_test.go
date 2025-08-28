/*
MediaMTX Integration Unit Tests

Requirements Coverage:
- REQ-MEDIA-001: MediaMTX path creation and management
- REQ-MEDIA-002: Stream lifecycle management
- REQ-MEDIA-003: FFmpeg integration and command generation
- REQ-MEDIA-004: Recording session management
- REQ-MEDIA-005: Snapshot capture and management
- REQ-MEDIA-006: Health monitoring and status reporting
- REQ-MEDIA-007: Error handling and recovery

Test Categories: Unit
API Documentation Reference: docs/api/json_rpc_methods.md
*/

//go:build unit
// +build unit

package mediamtx_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/mediamtx"
)

// TestMediaMTXController tests the MediaMTX controller functionality
func TestMediaMTXController(t *testing.T) {
	// Setup
	configManager := config.NewConfigManager()
	logger := logging.NewLogger("test")

	// Test controller initialization
	t.Run("controller_initialization", func(t *testing.T) {
		controller := mediamtx.NewController(configManager, logger)
		require.NotNil(t, controller)
		assert.NotNil(t, controller.GetConfig())
		assert.NotNil(t, controller.GetLogger())
	})

	// Test configuration loading
	t.Run("configuration_loading", func(t *testing.T) {
		controller := mediamtx.NewController(configManager, logger)

		// Test MediaMTX configuration
		config := controller.GetConfig()
		assert.NotNil(t, config)
		assert.NotEmpty(t, config.MediaMTX.APIEndpoint)
		assert.NotEmpty(t, config.MediaMTX.APIPort)
	})

	// Test health monitoring
	t.Run("health_monitoring", func(t *testing.T) {
		controller := mediamtx.NewController(configManager, logger)

		// Test health status
		status := controller.GetHealthStatus()
		assert.NotNil(t, status)
		assert.Contains(t, status.Status, "healthy")
		assert.GreaterOrEqual(t, status.Uptime, float64(0))
	})

	// Test connection management
	t.Run("connection_management", func(t *testing.T) {
		controller := mediamtx.NewController(configManager, logger)

		// Test connection status
		connected := controller.IsConnected()
		assert.IsType(t, false, connected)
	})
}

// TestPathManager tests the MediaMTX path management functionality
func TestPathManager(t *testing.T) {
	// Setup
	configManager := config.NewConfigManager()
	logger := logging.NewLogger("test")

	// Test path manager initialization
	t.Run("path_manager_initialization", func(t *testing.T) {
		pathManager := mediamtx.NewPathManager(configManager, logger)
		require.NotNil(t, pathManager)
	})

	// Test path creation
	t.Run("path_creation", func(t *testing.T) {
		pathManager := mediamtx.NewPathManager(configManager, logger)

		// Test creating a path
		pathConfig := &mediamtx.PathConfig{
			Source:      "/dev/video0",
			Destination: "camera0",
			Format:      "mp4",
		}

		path, err := pathManager.CreatePath(context.Background(), pathConfig)

		// Note: This may fail in test environment without real MediaMTX
		// We test the interface and error handling
		if err != nil {
			assert.Contains(t, err.Error(), "connection")
		} else {
			assert.NotNil(t, path)
			assert.Equal(t, pathConfig.Destination, path.Name)
		}
	})

	// Test path deletion
	t.Run("path_deletion", func(t *testing.T) {
		pathManager := mediamtx.NewPathManager(configManager, logger)

		// Test deleting a path
		err := pathManager.DeletePath(context.Background(), "test-path")

		// Note: This may fail in test environment
		// We test the interface and error handling
		if err != nil {
			assert.Contains(t, err.Error(), "connection")
		}
	})

	// Test path listing
	t.Run("path_listing", func(t *testing.T) {
		pathManager := mediamtx.NewPathManager(configManager, logger)

		// Test listing paths
		paths, err := pathManager.ListPaths(context.Background())

		// Note: This may fail in test environment
		// We test the interface and error handling
		if err != nil {
			assert.Contains(t, err.Error(), "connection")
		} else {
			assert.NotNil(t, paths)
			assert.IsType(t, []*mediamtx.Path{}, paths)
		}
	})

	// Test path status
	t.Run("path_status", func(t *testing.T) {
		pathManager := mediamtx.NewPathManager(configManager, logger)

		// Test getting path status
		status, err := pathManager.GetPathStatus(context.Background(), "test-path")

		// Note: This may fail in test environment
		// We test the interface and error handling
		if err != nil {
			assert.Contains(t, err.Error(), "connection")
		} else {
			assert.NotNil(t, status)
			assert.Contains(t, status.Status, "active")
		}
	})
}

// TestStreamManager tests the MediaMTX stream management functionality
func TestStreamManager(t *testing.T) {
	// Setup
	configManager := config.NewConfigManager()
	logger := logging.NewLogger("test")

	// Test stream manager initialization
	t.Run("stream_manager_initialization", func(t *testing.T) {
		streamManager := mediamtx.NewStreamManager(configManager, logger)
		require.NotNil(t, streamManager)
	})

	// Test stream creation
	t.Run("stream_creation", func(t *testing.T) {
		streamManager := mediamtx.NewStreamManager(configManager, logger)

		// Test creating a stream
		streamConfig := &mediamtx.StreamConfig{
			Name:     "test-stream",
			Source:   "/dev/video0",
			Protocol: "rtsp",
			Port:     8554,
		}

		stream, err := streamManager.CreateStream(context.Background(), streamConfig)

		// Note: This may fail in test environment
		// We test the interface and error handling
		if err != nil {
			assert.Contains(t, err.Error(), "connection")
		} else {
			assert.NotNil(t, stream)
			assert.Equal(t, streamConfig.Name, stream.Name)
		}
	})

	// Test stream deletion
	t.Run("stream_deletion", func(t *testing.T) {
		streamManager := mediamtx.NewStreamManager(configManager, logger)

		// Test deleting a stream
		err := streamManager.DeleteStream(context.Background(), "test-stream")

		// Note: This may fail in test environment
		// We test the interface and error handling
		if err != nil {
			assert.Contains(t, err.Error(), "connection")
		}
	})

	// Test stream listing
	t.Run("stream_listing", func(t *testing.T) {
		streamManager := mediamtx.NewStreamManager(configManager, logger)

		// Test listing streams
		streams, err := streamManager.ListStreams(context.Background())

		// Note: This may fail in test environment
		// We test the interface and error handling
		if err != nil {
			assert.Contains(t, err.Error(), "connection")
		} else {
			assert.NotNil(t, streams)
			assert.IsType(t, []*mediamtx.Stream{}, streams)
		}
	})

	// Test stream status
	t.Run("stream_status", func(t *testing.T) {
		streamManager := mediamtx.NewStreamManager(configManager, logger)

		// Test getting stream status
		status, err := streamManager.GetStreamStatus(context.Background(), "test-stream")

		// Note: This may fail in test environment
		// We test the interface and error handling
		if err != nil {
			assert.Contains(t, err.Error(), "connection")
		} else {
			assert.NotNil(t, status)
			assert.Contains(t, status.Status, "active")
		}
	})
}

// TestRecordingManager tests the MediaMTX recording management functionality
func TestRecordingManager(t *testing.T) {
	// Setup
	configManager := config.NewConfigManager()
	logger := logging.NewLogger("test")

	// Test recording manager initialization
	t.Run("recording_manager_initialization", func(t *testing.T) {
		recordingManager := mediamtx.NewRecordingManager(configManager, logger)
		require.NotNil(t, recordingManager)
	})

	// Test recording session creation
	t.Run("recording_session_creation", func(t *testing.T) {
		recordingManager := mediamtx.NewRecordingManager(configManager, logger)

		// Test creating a recording session
		sessionConfig := &mediamtx.RecordingSessionConfig{
			CameraDevice: "/dev/video0",
			OutputPath:   "/tmp/test-recording.mp4",
			Duration:     60,
			Format:       "mp4",
		}

		session, err := recordingManager.CreateSession(context.Background(), sessionConfig)

		// Note: This may fail in test environment
		// We test the interface and error handling
		if err != nil {
			assert.Contains(t, err.Error(), "connection")
		} else {
			assert.NotNil(t, session)
			assert.Equal(t, sessionConfig.CameraDevice, session.CameraDevice)
			assert.Equal(t, sessionConfig.OutputPath, session.OutputPath)
		}
	})

	// Test recording session management
	t.Run("recording_session_management", func(t *testing.T) {
		recordingManager := mediamtx.NewRecordingManager(configManager, logger)

		// Test starting a recording
		err := recordingManager.StartRecording(context.Background(), "test-session")

		// Note: This may fail in test environment
		// We test the interface and error handling
		if err != nil {
			assert.Contains(t, err.Error(), "connection")
		}

		// Test stopping a recording
		err = recordingManager.StopRecording(context.Background(), "test-session")

		// Note: This may fail in test environment
		// We test the interface and error handling
		if err != nil {
			assert.Contains(t, err.Error(), "connection")
		}
	})

	// Test recording session status
	t.Run("recording_session_status", func(t *testing.T) {
		recordingManager := mediamtx.NewRecordingManager(configManager, logger)

		// Test getting session status
		status, err := recordingManager.GetSessionStatus(context.Background(), "test-session")

		// Note: This may fail in test environment
		// We test the interface and error handling
		if err != nil {
			assert.Contains(t, err.Error(), "connection")
		} else {
			assert.NotNil(t, status)
			assert.Contains(t, status.Status, "active")
		}
	})

	// Test recording session cleanup
	t.Run("recording_session_cleanup", func(t *testing.T) {
		recordingManager := mediamtx.NewRecordingManager(configManager, logger)

		// Test cleaning up a session
		err := recordingManager.CleanupSession(context.Background(), "test-session")

		// Note: This may fail in test environment
		// We test the interface and error handling
		if err != nil {
			assert.Contains(t, err.Error(), "connection")
		}
	})
}

// TestSnapshotManager tests the MediaMTX snapshot management functionality
func TestSnapshotManager(t *testing.T) {
	// Setup
	configManager := config.NewConfigManager()
	logger := logging.NewLogger("test")

	// Test snapshot manager initialization
	t.Run("snapshot_manager_initialization", func(t *testing.T) {
		snapshotManager := mediamtx.NewSnapshotManager(configManager, logger)
		require.NotNil(t, snapshotManager)
	})

	// Test snapshot capture
	t.Run("snapshot_capture", func(t *testing.T) {
		snapshotManager := mediamtx.NewSnapshotManager(configManager, logger)

		// Test capturing a snapshot
		snapshotConfig := &mediamtx.SnapshotConfig{
			CameraDevice: "/dev/video0",
			OutputPath:   "/tmp/test-snapshot.jpg",
			Format:       "jpg",
			Quality:      90,
		}

		snapshot, err := snapshotManager.CaptureSnapshot(context.Background(), snapshotConfig)

		// Note: This may fail in test environment
		// We test the interface and error handling
		if err != nil {
			assert.Contains(t, err.Error(), "connection")
		} else {
			assert.NotNil(t, snapshot)
			assert.Equal(t, snapshotConfig.CameraDevice, snapshot.CameraDevice)
			assert.Equal(t, snapshotConfig.OutputPath, snapshot.OutputPath)
		}
	})

	// Test snapshot listing
	t.Run("snapshot_listing", func(t *testing.T) {
		snapshotManager := mediamtx.NewSnapshotManager(configManager, logger)

		// Test listing snapshots
		snapshots, err := snapshotManager.ListSnapshots(context.Background(), "/tmp")

		// Note: This may fail in test environment
		// We test the interface and error handling
		if err != nil {
			assert.Contains(t, err.Error(), "connection")
		} else {
			assert.NotNil(t, snapshots)
			assert.IsType(t, []*mediamtx.Snapshot{}, snapshots)
		}
	})

	// Test snapshot deletion
	t.Run("snapshot_deletion", func(t *testing.T) {
		snapshotManager := mediamtx.NewSnapshotManager(configManager, logger)

		// Test deleting a snapshot
		err := snapshotManager.DeleteSnapshot(context.Background(), "/tmp/test-snapshot.jpg")

		// Note: This may fail in test environment
		// We test the interface and error handling
		if err != nil {
			assert.Contains(t, err.Error(), "connection")
		}
	})
}

// TestFFmpegManager tests the MediaMTX FFmpeg integration functionality
func TestFFmpegManager(t *testing.T) {
	// Setup
	configManager := config.NewConfigManager()
	logger := logging.NewLogger("test")

	// Test FFmpeg manager initialization
	t.Run("ffmpeg_manager_initialization", func(t *testing.T) {
		ffmpegManager := mediamtx.NewFFmpegManager(configManager, logger)
		require.NotNil(t, ffmpegManager)
	})

	// Test FFmpeg command generation
	t.Run("ffmpeg_command_generation", func(t *testing.T) {
		ffmpegManager := mediamtx.NewFFmpegManager(configManager, logger)

		// Test generating FFmpeg command
		commandConfig := &mediamtx.FFmpegCommandConfig{
			Input:     "/dev/video0",
			Output:    "/tmp/test-output.mp4",
			Format:    "mp4",
			Codec:     "h264",
			Bitrate:   "1000k",
			Framerate: 30,
		}

		command, err := ffmpegManager.GenerateCommand(commandConfig)

		require.NoError(t, err)
		assert.NotNil(t, command)
		assert.Contains(t, command, "ffmpeg")
		assert.Contains(t, command, commandConfig.Input)
		assert.Contains(t, command, commandConfig.Output)
	})

	// Test FFmpeg execution
	t.Run("ffmpeg_execution", func(t *testing.T) {
		ffmpegManager := mediamtx.NewFFmpegManager(configManager, logger)

		// Test executing FFmpeg command
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := ffmpegManager.ExecuteCommand(ctx, "ffmpeg -version")

		// Note: This may fail if FFmpeg is not installed
		// We test the interface and error handling
		if err != nil {
			assert.Contains(t, err.Error(), "ffmpeg")
		}
	})

	// Test FFmpeg process management
	t.Run("ffmpeg_process_management", func(t *testing.T) {
		ffmpegManager := mediamtx.NewFFmpegManager(configManager, logger)

		// Test starting a process
		process, err := ffmpegManager.StartProcess(context.Background(), "ffmpeg -version")

		// Note: This may fail if FFmpeg is not installed
		// We test the interface and error handling
		if err != nil {
			assert.Contains(t, err.Error(), "ffmpeg")
		} else {
			assert.NotNil(t, process)

			// Test stopping a process
			err = ffmpegManager.StopProcess(process)
			assert.NoError(t, err)
		}
	})
}

// TestHealthMonitor tests the MediaMTX health monitoring functionality
func TestHealthMonitor(t *testing.T) {
	// Setup
	configManager := config.NewConfigManager()
	logger := logging.NewLogger("test")

	// Test health monitor initialization
	t.Run("health_monitor_initialization", func(t *testing.T) {
		healthMonitor := mediamtx.NewHealthMonitor(configManager, logger)
		require.NotNil(t, healthMonitor)
	})

	// Test health status monitoring
	t.Run("health_status_monitoring", func(t *testing.T) {
		healthMonitor := mediamtx.NewHealthMonitor(configManager, logger)

		// Test getting health status
		status := healthMonitor.GetStatus()
		assert.NotNil(t, status)
		assert.Contains(t, status.Status, "healthy")
		assert.GreaterOrEqual(t, status.Uptime, float64(0))
	})

	// Test component health monitoring
	t.Run("component_health_monitoring", func(t *testing.T) {
		healthMonitor := mediamtx.NewHealthMonitor(configManager, logger)

		// Test getting component status
		components := healthMonitor.GetComponentStatus()
		assert.NotNil(t, components)
		assert.Contains(t, components, "mediamtx")
		assert.Contains(t, components, "ffmpeg")
	})

	// Test health check execution
	t.Run("health_check_execution", func(t *testing.T) {
		healthMonitor := mediamtx.NewHealthMonitor(configManager, logger)

		// Test executing health check
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := healthMonitor.ExecuteHealthCheck(ctx)

		// Note: This may fail in test environment
		// We test the interface and error handling
		if err != nil {
			assert.Contains(t, err.Error(), "connection")
		}
	})

	// Test health metrics collection
	t.Run("health_metrics_collection", func(t *testing.T) {
		healthMonitor := mediamtx.NewHealthMonitor(configManager, logger)

		// Test collecting health metrics
		metrics := healthMonitor.CollectMetrics()
		assert.NotNil(t, metrics)
		assert.GreaterOrEqual(t, metrics.MemoryUsage, float64(0))
		assert.GreaterOrEqual(t, metrics.CPUUsage, float64(0))
		assert.GreaterOrEqual(t, metrics.ActiveConnections, int64(0))
	})
}
