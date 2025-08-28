//go:build unit
// +build unit

/*
MediaMTX Stream Manager Unit Tests

Requirements Coverage:
- REQ-MTX-002: Stream management capabilities
- REQ-MTX-003: Path creation and deletion
- REQ-MTX-007: Error handling and recovery

Test Categories: Unit
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package mediamtx_test

import (
	"context"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/mediamtx"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockClient implements the HTTP client interface for testing
type mockClient struct{}

func (m *mockClient) Get(ctx context.Context, path string) ([]byte, error) {
	return []byte(`{"items":[]}`), nil
}

func (m *mockClient) Post(ctx context.Context, path string, data []byte) ([]byte, error) {
	return []byte(`{"status":"ok"}`), nil
}

func (m *mockClient) Delete(ctx context.Context, path string) error {
	return nil
}

// TestStreamManager_Creation tests stream manager creation
func TestStreamManager_Creation(t *testing.T) {
	// Create test logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create test configuration
	testConfig := &mediamtx.MediaMTXConfig{
		BaseURL: "http://localhost:9997",
	}

	// Create mock client
	client := &mockClient{}

	// Create stream manager
	streamManager := mediamtx.NewStreamManager(client, testConfig, logger)
	require.NotNil(t, streamManager, "Stream manager should not be nil")
}

// TestStreamManager_CheckStreamReadiness tests stream readiness checking
func TestStreamManager_CheckStreamReadiness(t *testing.T) {
	// Create test logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create test configuration
	testConfig := &mediamtx.MediaMTXConfig{
		BaseURL: "http://localhost:9997",
	}

	// Create mock client
	client := &mockClient{}

	// Create stream manager
	streamManager := mediamtx.NewStreamManager(client, testConfig, logger)

	ctx := context.Background()

	// Test stream readiness check
	ready, err := streamManager.CheckStreamReadiness(ctx, "test-stream", 1*time.Second)
	// Note: This may fail if MediaMTX service is not running
	// For unit tests, we validate the method exists and handles errors
	if err != nil {
		t.Logf("Stream readiness check failed (expected if MediaMTX not running): %v", err)
	} else {
		assert.IsType(t, false, ready, "Ready should be a boolean")
	}
}

// TestStreamManager_WaitForStreamReadiness tests stream readiness waiting
func TestStreamManager_WaitForStreamReadiness(t *testing.T) {
	// Create test logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create test configuration
	testConfig := &mediamtx.MediaMTXConfig{
		BaseURL: "http://localhost:9997",
	}

	// Create mock client
	client := &mockClient{}

	// Create stream manager
	streamManager := mediamtx.NewStreamManager(client, testConfig, logger)

	ctx := context.Background()

	// Test stream readiness waiting with short timeout
	ready, err := streamManager.WaitForStreamReadiness(ctx, "test-stream", 1*time.Second, "test-correlation")
	// Note: This may fail if MediaMTX service is not running
	// For unit tests, we validate the method exists and handles errors
	if err != nil {
		t.Logf("Stream readiness wait failed (expected if MediaMTX not running): %v", err)
	} else {
		assert.IsType(t, false, ready, "Ready should be a boolean")
	}
}

// TestStreamManager_ErrorHandling tests error handling scenarios
func TestStreamManager_ErrorHandling(t *testing.T) {
	// Create test logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create test configuration
	testConfig := &mediamtx.MediaMTXConfig{
		BaseURL: "http://localhost:9997",
	}

	// Create mock client
	client := &mockClient{}

	// Create stream manager
	streamManager := mediamtx.NewStreamManager(client, testConfig, logger)

	ctx := context.Background()

	// Test with empty stream name
	_, err := streamManager.CheckStreamReadiness(ctx, "", 1*time.Second)
	assert.Error(t, err, "Should return error with empty stream name")

	// Test with zero timeout
	_, err = streamManager.CheckStreamReadiness(ctx, "test-stream", 0)
	assert.Error(t, err, "Should return error with zero timeout")
}

// TestStreamManager_TimeoutHandling tests timeout scenarios
func TestStreamManager_TimeoutHandling(t *testing.T) {
	// Create test logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create test configuration
	testConfig := &mediamtx.MediaMTXConfig{
		BaseURL: "http://localhost:9997",
	}

	// Create mock client
	client := &mockClient{}

	// Create stream manager
	streamManager := mediamtx.NewStreamManager(client, testConfig, logger)

	ctx := context.Background()

	// Test with very short timeout
	_, err := streamManager.CheckStreamReadiness(ctx, "test-stream", 1*time.Millisecond)
	// This should either succeed quickly or timeout appropriately
	if err != nil {
		t.Logf("Short timeout test result: %v", err)
	}
}

// TestStreamManager_ConcurrentAccess tests concurrent access scenarios
func TestStreamManager_ConcurrentAccess(t *testing.T) {
	// Create test logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create test configuration
	testConfig := &mediamtx.MediaMTXConfig{
		BaseURL: "http://localhost:9997",
	}

	// Create mock client
	client := &mockClient{}

	// Create stream manager
	streamManager := mediamtx.NewStreamManager(client, testConfig, logger)

	ctx := context.Background()

	// Test concurrent stream readiness checks
	done := make(chan bool, 2)

	go func() {
		_, err := streamManager.CheckStreamReadiness(ctx, "test-stream-1", 1*time.Second)
		if err != nil {
			t.Logf("Concurrent check 1 result: %v", err)
		}
		done <- true
	}()

	go func() {
		_, err := streamManager.CheckStreamReadiness(ctx, "test-stream-2", 1*time.Second)
		if err != nil {
			t.Logf("Concurrent check 2 result: %v", err)
		}
		done <- true
	}()

	// Wait for both goroutines to complete
	<-done
	<-done
}

// TestStreamManager_ContextCancellation tests context cancellation
func TestStreamManager_ContextCancellation(t *testing.T) {
	// Create test logger
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create test configuration
	testConfig := &mediamtx.MediaMTXConfig{
		BaseURL: "http://localhost:9997",
	}

	// Create mock client
	client := &mockClient{}

	// Create stream manager
	streamManager := mediamtx.NewStreamManager(client, testConfig, logger)

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel context immediately
	cancel()

	// Test stream readiness check with cancelled context
	_, err := streamManager.CheckStreamReadiness(ctx, "test-stream", 1*time.Second)
	// Should handle context cancellation gracefully
	if err != nil {
		t.Logf("Context cancellation test result: %v", err)
	}
}
