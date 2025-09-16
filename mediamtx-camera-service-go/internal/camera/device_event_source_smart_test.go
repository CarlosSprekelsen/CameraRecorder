/*
Smart Device Event Source Selection Tests

Tests the environment detection and automatic selection of optimal device event source
based on the deployment environment (container vs bare metal).

Test Categories: Unit
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package camera

import (
	"context"
	"testing"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSmartDeviceEventSourceSelection tests the smart selection logic
func TestSmartDeviceEventSourceSelection(t *testing.T) {
	t.Parallel()
	logger := logging.CreateTestLogger(t, nil)

	t.Run("factory_creates_instance", func(t *testing.T) {
		factory := GetDeviceEventSourceFactory()
		require.NotNil(t, factory, "Factory should be available")

		// Create an instance - should use smart selection
		instance := factory.Create()
		require.NotNil(t, instance, "Should create a device event source instance")

		// Verify it implements the interface
		assert.Implements(t, (*DeviceEventSource)(nil), instance, "Should implement DeviceEventSource interface")

		// Clean up
		instance.Close()
	})

	t.Run("environment_detection_functions", func(t *testing.T) {
		// Test environment detection functions exist and return boolean values
		isContainer := isContainerEnvironment()
		assert.IsType(t, false, isContainer, "isContainerEnvironment should return bool")

		isUdev := isUdevAvailable()
		assert.IsType(t, false, isUdev, "isUdevAvailable should return bool")

		// Test optimal selection function
		sourceType := getOptimalDeviceEventSourceType(logger)
		assert.Contains(t, []string{"fsnotify", "udev"}, sourceType, "Should return valid source type")

		t.Logf("Environment detection results: container=%v, udev=%v, selected=%s",
			isContainer, isUdev, sourceType)
	})

	t.Run("smart_selection_logging", func(t *testing.T) {
		// Create factory and instance to trigger smart selection logging
		factory := GetDeviceEventSourceFactory()
		instance := factory.Create()
		require.NotNil(t, instance)

		// Verify the instance can be started and stopped
		ctx := context.Background()
		err := instance.Start(ctx)
		require.NoError(t, err, "Should start successfully")

		err = instance.Close()
		require.NoError(t, err, "Should close successfully")
	})

	t.Run("multiple_instances_isolation", func(t *testing.T) {
		factory := GetDeviceEventSourceFactory()

		// Create multiple instances - each should be independent
		instance1 := factory.Create()
		instance2 := factory.Create()

		require.NotNil(t, instance1)
		require.NotNil(t, instance2)
		assert.NotEqual(t, instance1, instance2, "Should create different instances")

		// Both should be able to start independently
		ctx := context.Background()
		err1 := instance1.Start(ctx)
		err2 := instance2.Start(ctx)

		require.NoError(t, err1, "First instance should start")
		require.NoError(t, err2, "Second instance should start")

		// Clean up
		instance1.Close()
		instance2.Close()
	})
}

// TestEnvironmentDetectionAccuracy tests the accuracy of environment detection
func TestEnvironmentDetectionAccuracy(t *testing.T) {
	t.Parallel()
	t.Run("container_detection", func(t *testing.T) {
		isContainer := isContainerEnvironment()

		// In our test environment, we might be in a container or bare metal
		// Just verify the function works and returns a boolean
		assert.IsType(t, false, isContainer, "Should return boolean")

		t.Logf("Container detection result: %v", isContainer)
	})

	t.Run("udev_detection", func(t *testing.T) {
		isUdev := isUdevAvailable()

		// Verify the function works and returns a boolean
		assert.IsType(t, false, isUdev, "Should return boolean")

		t.Logf("Udev detection result: %v", isUdev)
	})

	t.Run("selection_logic", func(t *testing.T) {
		logger := logging.CreateTestLogger(t, nil)

		isContainer := isContainerEnvironment()
		isUdev := isUdevAvailable()
		selectedType := getOptimalDeviceEventSourceType(logger)

		// Verify selection logic
		if isContainer {
			assert.Equal(t, "fsnotify", selectedType, "Container should select fsnotify")
		} else if isUdev {
			assert.Equal(t, "udev", selectedType, "Bare metal with udev should select udev")
		} else {
			assert.Equal(t, "fsnotify", selectedType, "Bare metal without udev should fallback to fsnotify")
		}

		t.Logf("Selection logic: container=%v, udev=%v, selected=%s",
			isContainer, isUdev, selectedType)
	})
}
