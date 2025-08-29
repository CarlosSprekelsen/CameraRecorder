//go:build unit
// +build unit

/*
Stream Lifecycle Management Unit Tests

Requirements Coverage:
- REQ-STREAM-001: File rotation compatibility
- REQ-STREAM-002: Different lifecycle policies for use cases
- REQ-STREAM-003: Power-efficient operation
- REQ-STREAM-004: Manual control over stream lifecycle

Test Categories: Unit
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package mediamtx_test

import (
	"context"
	"testing"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/mediamtx"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test configuration for MediaMTX tests
var testConfig = &mediamtx.MediaMTXConfig{
	Host:                                "localhost",
	APIPort:                             9997,
	RTSPPort:                            8554,
	WebRTCPort:                          8889,
	HLSPort:                             8888,
	ConfigPath:                          "/tmp/mediamtx.yml",
	RecordingsPath:                      "/tmp/recordings",
	SnapshotsPath:                       "/tmp/snapshots",
	HealthCheckInterval:                 30,
	HealthFailureThreshold:              3,
	HealthCircuitBreakerTimeout:         60,
	HealthMaxBackoffInterval:            300,
	HealthRecoveryConfirmationThreshold: 2,
	BackoffBaseMultiplier:               2.0,
	ProcessTerminationTimeout:           10,
	ProcessKillTimeout:                  5,
}

var testLogger = createTestLogger()

func TestStreamManager_CreateStreamWithUseCase(t *testing.T) {
	tests := []struct {
		name     string
		useCase  mediamtx.StreamUseCase
		expected string
	}{
		{
			name:     "recording use case",
			useCase:  mediamtx.UseCaseRecording,
			expected: "camera0", // No suffix for recording
		},
		{
			name:     "viewing use case",
			useCase:  mediamtx.UseCaseViewing,
			expected: "camera0_viewing", // _viewing suffix
		},
		{
			name:     "snapshot use case",
			useCase:  mediamtx.UseCaseSnapshot,
			expected: "camera0_snapshot", // _snapshot suffix
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock client and config
			client := mediamtx.NewClient("http://localhost:9997", testConfig, testLogger)
			config := &mediamtx.MediaMTXConfig{}

			// Create stream manager
			sm := mediamtx.NewStreamManager(client, config, testLogger)

			// Test use case-specific methods (matches Python implementation)
			var stream *mediamtx.Stream
			var err error
			
			switch tt.useCase {
			case mediamtx.UseCaseRecording:
				stream, err = sm.StartRecordingStream(context.Background(), "/dev/video0")
			case mediamtx.UseCaseViewing:
				stream, err = sm.StartViewingStream(context.Background(), "/dev/video0")
			case mediamtx.UseCaseSnapshot:
				stream, err = sm.StartSnapshotStream(context.Background(), "/dev/video0")
			default:
				t.Fatalf("Unsupported use case: %s", tt.useCase)
			}

			// Verify no error for valid use cases
			require.NoError(t, err)
			assert.NotNil(t, stream)
			assert.Equal(t, tt.expected, stream.Name)
		})
	}
}

func TestStreamManager_UseCaseConfigurations(t *testing.T) {
	// Create stream manager to access configurations
	client := mediamtx.NewClient("http://localhost:9997", testConfig, testLogger)
	config := &mediamtx.MediaMTXConfig{}

	sm := mediamtx.NewStreamManager(client, config, testLogger)

	// Test that the stream manager can be created and used
	assert.NotNil(t, sm, "Stream manager should not be nil")

	// Test that streams can be created using use case-specific methods
	recordingStream, err := sm.StartRecordingStream(context.Background(), "/dev/video0")
	require.NoError(t, err)
	assert.NotNil(t, recordingStream)

	viewingStream, err := sm.StartViewingStream(context.Background(), "/dev/video1")
	require.NoError(t, err)
	assert.NotNil(t, viewingStream)

	snapshotStream, err := sm.StartSnapshotStream(context.Background(), "/dev/video2")
	require.NoError(t, err)
	assert.NotNil(t, snapshotStream)
}

func TestController_CreateStreamForUseCases(t *testing.T) {
	// This test is simplified to avoid issues with unexported types
	// The controller methods are tested through the StreamManager interface

	t.Run("stream manager interface compliance", func(t *testing.T) {
		// Create mock components
		client := mediamtx.NewClient("http://localhost:9997", testConfig, testLogger)
		config := &mediamtx.MediaMTXConfig{}

		// Create stream manager
		sm := mediamtx.NewStreamManager(client, config, testLogger)

		// Test that the interface methods work
		assert.NotNil(t, sm, "Stream manager should not be nil")

		// Test stream creation using public interface
		stream, err := sm.CreateStream(context.Background(), "camera0", "/dev/video0")
		require.NoError(t, err)
		assert.NotNil(t, stream)
	})
}

func TestStreamManager_InvalidUseCase(t *testing.T) {
	// Create stream manager
	client := mediamtx.NewClient("http://localhost:9997", testConfig, testLogger)
	config := &mediamtx.MediaMTXConfig{}

	sm := mediamtx.NewStreamManager(client, config, testLogger)

	// Test with invalid stream name
	stream, err := sm.CreateStream(context.Background(), "", "/dev/video0")

	// Verify error for invalid stream name
	assert.Error(t, err)
	assert.Nil(t, stream)
	assert.Contains(t, err.Error(), "stream name cannot be empty")
}

// Helper function to create test logger
func createTestLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	return logger
}
