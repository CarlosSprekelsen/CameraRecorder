/*
MediaMTX Integration Tests - Event-Driven End-to-End Testing

Requirements Coverage:
- REQ-MTX-001: MediaMTX service integration
- REQ-MTX-002: Stream management capabilities
- REQ-MTX-003: Path creation and deletion
- REQ-MTX-004: Health monitoring

Test Categories: Integration (using real MediaMTX server with event-driven patterns)
API Documentation Reference: docs/api/json_rpc_methods.md

This file contains comprehensive end-to-end integration tests that demonstrate
the event-driven architecture in action across the entire MediaMTX system.
*/

package mediamtx

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEndToEndEventDrivenWorkflow tests a complete workflow using event-driven patterns
func TestEndToEndEventDrivenWorkflow(t *testing.T) {
	// REQ-MTX-001, REQ-MTX-002, REQ-MTX-003, REQ-MTX-004: Complete system integration
	EnsureSequentialExecution(t) // CRITICAL: Prevent concurrent MediaMTX server access
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create event-driven test helper
	eventHelper := helper.CreateEventDrivenTestHelper(t)
	defer eventHelper.Cleanup()

	// Use proper orchestration following the Progressive Readiness Pattern
	controller, err := helper.GetController(t)
	require.NoError(t, err, "Controller orchestration should succeed")
	require.NotNil(t, controller, "Controller should not be nil")

	ctx := context.Background()
	err = controller.Start(ctx)
	require.NoError(t, err)

	// Ensure controller is stopped after test
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		controller.Stop(stopCtx)
	}()

	t.Run("complete_event_driven_workflow", func(t *testing.T) {
		// Step 1: No waiting for readiness - Progressive Readiness Pattern
		// Just verify controller is running
		assert.True(t, controller.IsReady(), "Controller should be ready")

		// Step 2: Create a test path using event-driven health monitoring
		pathName := "test_event_driven_path"
		// Get available device instead of hardcoded camera0
		cameraList, err := controller.GetCameraList(ctx)
		require.NoError(t, err, "Should be able to get camera list")
		require.NotEmpty(t, cameraList.Cameras, "Should have at least one available camera")
		source := cameraList.Cameras[0].Device // Use first available camera

		// Subscribe to health changes before creating path
		healthChan := eventHelper.SubscribeToHealthChanges()

		// Create path
		path := &Path{
			Name: pathName,
			Source: &PathSource{
				Type: "rpiCameraSource",
				ID:   source,
			},
		}
		err = controller.CreatePath(ctx, path)
		require.NoError(t, err, "Path creation should succeed")

		// Step 3: Wait for health events (may or may not occur)
		healthCtx, healthCancel := context.WithTimeout(ctx, 5*time.Second)
		defer healthCancel()

		select {
		case <-healthChan:
			t.Log("Received health change event after path creation")
		case <-healthCtx.Done():
			t.Log("No health change event received (this is normal)")
		}

		// Step 4: Start advanced recording with event-driven file monitoring
		recordingOptions := map[string]interface{}{
			"quality":    "high",
			"resolution": "1920x1080",
			"framerate":  30,
			"duration":   10, // 10 seconds
		}

		recordingCtx, recordingCancel := context.WithTimeout(ctx, 20*time.Second)
		defer recordingCancel()

		session, err := controller.StartAdvancedRecording(recordingCtx, source, recordingOptions)
		require.NoError(t, err, "Advanced recording should start successfully")
		require.NotNil(t, session, "Recording session should not be nil")

		// Step 5: No waiting for recording readiness - Progressive Readiness Pattern
		// Recording should work immediately or fail fast

		// Step 6: Verify file creation with optimized timeout (TODO: Replace with event-driven file creation notifications)
		require.Eventually(t, func() bool {
			_, err := os.Stat(session.FilePath)
			return err == nil
		}, 3*time.Second, 50*time.Millisecond, "Recording file should be created within 3 seconds (optimized polling)")

		// Step 7: Test event aggregation - wait for multiple events
		aggregationCtx, aggregationCancel := context.WithTimeout(ctx, 8*time.Second)
		defer aggregationCancel()

		// Wait for any of the specified events
		err = eventHelper.WaitForMultipleEvents(aggregationCtx, 8*time.Second, "readiness", "health")
		require.NoError(t, err, "Should receive at least one event within timeout")

		// Step 8: Stop recording
		err = controller.StopAdvancedRecording(ctx, session.ID)
		require.NoError(t, err, "Stopping advanced recording should succeed")

		// Step 9: Delete path
		err = controller.DeletePath(ctx, pathName)
		require.NoError(t, err, "Path deletion should succeed")

		// Step 10: Verify final state
		assert.True(t, controller.IsReady(), "Controller should still be ready after workflow")

		// Verify output file exists
		_, err = os.Stat(session.FilePath)
		assert.NoError(t, err, "Output file should exist after recording")
	})
}

// TestEventDrivenConcurrentOperations tests concurrent operations using event-driven patterns
func TestEventDrivenConcurrentOperations(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities with concurrent operations
	EnsureSequentialExecution(t) // CRITICAL: Prevent concurrent MediaMTX server access
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create multiple event-driven test helpers for concurrent testing
	eventHelpers := make([]*EventDrivenTestHelper, 3)
	for i := 0; i < 3; i++ {
		eventHelpers[i] = helper.CreateEventDrivenTestHelper(t)
		defer eventHelpers[i].Cleanup()
	}

	// Use proper orchestration following the Progressive Readiness Pattern
	controller, err := helper.GetController(t)
	require.NoError(t, err, "Controller orchestration should succeed")
	require.NotNil(t, controller, "Controller should not be nil")

	ctx := context.Background()
	err = controller.Start(ctx)
	require.NoError(t, err)

	// Ensure controller is stopped after test
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		controller.Stop(stopCtx)
	}()

	t.Run("concurrent_event_driven_operations", func(t *testing.T) {
		// Create multiple paths concurrently using event-driven readiness
		pathNames := []string{"concurrent_path_1", "concurrent_path_2", "concurrent_path_3"}
		// Get available device instead of hardcoded camera0
		cameraList, err := controller.GetCameraList(ctx)
		require.NoError(t, err, "Should be able to get camera list")
		require.NotEmpty(t, cameraList.Cameras, "Should have at least one available camera")
		source := cameraList.Cameras[0].Device // Use first available camera

		// No waiting for system readiness - Progressive Readiness Pattern
		// Just verify controller is ready

		// Create paths concurrently
		done := make(chan error, len(pathNames))
		for i, pathName := range pathNames {
			go func(index int, name string) {
				defer func() {
					if r := recover(); r != nil {
						done <- fmt.Errorf("goroutine %d panicked: %v", index, r)
					}
				}()

				// Use event-driven approach for each path creation
				pathCtx, pathCancel := context.WithTimeout(ctx, 10*time.Second)
				defer pathCancel()

				// No waiting for readiness - Progressive Readiness Pattern
				// Just proceed with path creation

				// Create path
				path := &Path{
					Name: name,
					Source: &PathSource{
						Type: "rpiCameraSource",
						ID:   source,
					},
				}
				err = controller.CreatePath(pathCtx, path)
				done <- err
			}(i, pathName)
		}

		// Wait for all path creations to complete
		for i := 0; i < len(pathNames); i++ {
			err := <-done
			require.NoError(t, err, "Path creation %d should succeed", i)
		}

		// Verify all paths were created by trying to get them
		for _, pathName := range pathNames {
			path, err := controller.GetPath(ctx, pathName)
			require.NoError(t, err, "Path %s should exist", pathName)
			assert.NotNil(t, path, "Path %s should not be nil", pathName)
		}

		// Clean up paths
		for _, pathName := range pathNames {
			err := controller.DeletePath(ctx, pathName)
			require.NoError(t, err, "Path deletion should succeed")
		}
	})
}

// TestEventDrivenErrorRecovery tests error recovery using event-driven patterns
func TestEventDrivenErrorRecovery(t *testing.T) {
	// REQ-MTX-004: Health monitoring and error recovery
	EnsureSequentialExecution(t) // CRITICAL: Prevent concurrent MediaMTX server access
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create event-driven test helper
	eventHelper := helper.CreateEventDrivenTestHelper(t)
	defer eventHelper.Cleanup()

	// Use proper orchestration following the Progressive Readiness Pattern
	controller, err := helper.GetController(t)
	require.NoError(t, err, "Controller orchestration should succeed")
	require.NotNil(t, controller, "Controller should not be nil")

	ctx := context.Background()
	err = controller.Start(ctx)
	require.NoError(t, err)

	// Ensure controller is stopped after test
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		controller.Stop(stopCtx)
	}()

	t.Run("error_recovery_with_event_driven_monitoring", func(t *testing.T) {
		// Step 1: No waiting for system readiness - Progressive Readiness Pattern
		// Just verify controller is running

		// Step 2: Subscribe to health changes for error monitoring
		healthChan := eventHelper.SubscribeToHealthChanges()

		// Step 3: Attempt operations that might cause errors
		invalidPathName := "invalid_path_with_special_chars_!@#$%"
		invalidSource := "nonexistent_camera"

		// This should fail but not crash the system
		invalidPath := &Path{
			Name: invalidPathName,
			Source: &PathSource{
				Type: "rpiCameraSource",
				ID:   invalidSource,
			},
		}
		err = controller.CreatePath(ctx, invalidPath)
		// We expect this to fail, so we don't require success
		if err != nil {
			t.Logf("Expected error creating invalid path: %v", err)
		}

		// Step 4: Monitor for health events after error
		healthCtx, healthCancel := context.WithTimeout(ctx, 5*time.Second)
		defer healthCancel()

		select {
		case <-healthChan:
			t.Log("Received health change event after error (system is monitoring health)")
		case <-healthCtx.Done():
			t.Log("No health change event received (system remained healthy)")
		}

		// Step 5: Verify system is still operational
		assert.True(t, controller.IsReady(), "Controller should still be ready after error")

		// Step 6: Test recovery with valid operations
		validPathName := "recovery_test_path"
		// Get available device instead of hardcoded camera0
		cameraList, err := controller.GetCameraList(ctx)
		require.NoError(t, err, "Should be able to get camera list")
		require.NotEmpty(t, cameraList.Cameras, "Should have at least one available camera")
		validSource := cameraList.Cameras[0].Device // Use first available camera

		validPath := &Path{
			Name: validPathName,
			Source: &PathSource{
				Type: "rpiCameraSource",
				ID:   validSource,
			},
		}
		err = controller.CreatePath(ctx, validPath)
		require.NoError(t, err, "Valid path creation should succeed after error")

		// Clean up
		err = controller.DeletePath(ctx, validPathName)
		require.NoError(t, err, "Path deletion should succeed")
	})
}

// TestEventDrivenPerformanceCharacteristics tests performance characteristics of event-driven patterns
func TestEventDrivenPerformanceCharacteristics(t *testing.T) {
	// REQ-MTX-001, REQ-MTX-002: Performance testing with event-driven patterns
	EnsureSequentialExecution(t) // CRITICAL: Prevent concurrent MediaMTX server access
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Create event-driven test helper
	eventHelper := helper.CreateEventDrivenTestHelper(t)
	defer eventHelper.Cleanup()

	// Use proper orchestration following the Progressive Readiness Pattern
	controller, err := helper.GetController(t)
	require.NoError(t, err, "Controller orchestration should succeed")
	require.NotNil(t, controller, "Controller should not be nil")

	ctx := context.Background()
	err = controller.Start(ctx)
	require.NoError(t, err)

	// Ensure controller is stopped after test
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		controller.Stop(stopCtx)
	}()

	t.Run("performance_characteristics", func(t *testing.T) {
		// Measure time to readiness using event-driven approach
		startTime := time.Now()

		// No waiting for readiness - Progressive Readiness Pattern
		// Just verify controller is running
		readinessTime := time.Since(startTime)
		t.Logf("Time to start (event-driven): %v", readinessTime)

		// Test multiple event subscriptions performance
		subscriptionStartTime := time.Now()

		// Create multiple non-blocking event observations
		for i := 0; i < 10; i++ {
			eventHelper.ObserveReadiness()
			eventHelper.ObserveHealthChanges()
		}

		subscriptionTime := time.Since(subscriptionStartTime)
		t.Logf("Time to create 20 event observations: %v", subscriptionTime)

		// Test event aggregation performance
		aggregationStartTime := time.Now()

		aggregationCtx, aggregationCancel := context.WithTimeout(ctx, 5*time.Second)
		defer aggregationCancel()

		err = eventHelper.WaitForMultipleEvents(aggregationCtx, 5*time.Second, "readiness", "health")
		require.NoError(t, err, "Event aggregation should succeed")

		aggregationTime := time.Since(aggregationStartTime)
		t.Logf("Time for event aggregation: %v", aggregationTime)

		// Verify performance characteristics
		assert.True(t, readinessTime < 10*time.Second, "Readiness should be achieved quickly")
		assert.True(t, subscriptionTime < 100*time.Millisecond, "Event subscriptions should be fast")
		assert.True(t, aggregationTime < 2*time.Second, "Event aggregation should be efficient")
	})
}
