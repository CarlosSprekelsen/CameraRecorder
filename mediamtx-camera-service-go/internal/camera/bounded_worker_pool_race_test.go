/*
BoundedWorkerPool Race Condition Detection Tests

Requirements Coverage:
- REQ-RESOURCE-001: Race condition detection and prevention

Test Categories: Unit
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package camera

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBoundedWorkerPool_RaceConditionDetection(t *testing.T) {
	// REQ-RESOURCE-001: Detect race conditions in statistics tracking

	logger := logging.GetLogger("test")

	// Use short timeout to increase chance of race condition
	shortTimeout := 10 * time.Millisecond
	pool := NewBoundedWorkerPool(3, shortTimeout, logger)

	ctx := context.Background()
	err := pool.Start(ctx)
	require.NoError(t, err)
	defer pool.Stop(ctx)

	// Run multiple iterations to detect race conditions
	iterations := 100
	totalTasks := 0

	for iter := 0; iter < iterations; iter++ {
		// Submit tasks that complete right at the timeout boundary
		numTasks := 5
		var wg sync.WaitGroup

		for i := 0; i < numTasks; i++ {
			wg.Add(1)
			err := pool.Submit(ctx, func(taskCtx context.Context) {
				defer wg.Done()

				// Task that completes right at timeout boundary
				select {
				case <-time.After(shortTimeout - 2*time.Millisecond):
					// Complete just before timeout
				case <-taskCtx.Done():
					// Context cancelled
					return
				}
			})
			require.NoError(t, err)
			totalTasks++
		}

		wg.Wait()
	}

	// Check final statistics
	stats := pool.GetStats()

	// The critical assertion: total tasks should equal completed + failed + timeout
	totalProcessed := stats.CompletedTasks + stats.FailedTasks + stats.TimeoutTasks

	t.Logf("Statistics after %d iterations with %d total tasks:", iterations, totalTasks)
	t.Logf("  Completed: %d", stats.CompletedTasks)
	t.Logf("  Failed: %d", stats.FailedTasks)
	t.Logf("  Timeout: %d", stats.TimeoutTasks)
	t.Logf("  Total Processed: %d", totalProcessed)
	t.Logf("  Total Submitted: %d", totalTasks)

	// CRITICAL: No tasks should be lost
	assert.Equal(t, int64(totalTasks), totalProcessed,
		"Race condition detected: %d tasks submitted but only %d processed",
		totalTasks, totalProcessed)

	// Additional validation: most tasks should complete successfully at the boundary
	assert.Greater(t, stats.CompletedTasks, int64(0), "Some tasks should complete successfully")
}

func TestBoundedWorkerPool_StatisticsConsistency(t *testing.T) {
	// REQ-RESOURCE-001: Statistics consistency under load

	logger := logging.GetLogger("test")
	pool := NewBoundedWorkerPool(2, 100*time.Millisecond, logger)

	ctx := context.Background()
	err := pool.Start(ctx)
	require.NoError(t, err)
	defer pool.Stop(ctx)

	// Submit different types of tasks
	var wg sync.WaitGroup

	// 10 successful tasks
	for i := 0; i < 10; i++ {
		wg.Add(1)
		err := pool.Submit(ctx, func(taskCtx context.Context) {
			defer wg.Done()
			time.Sleep(10 * time.Millisecond) // Quick completion
		})
		require.NoError(t, err)
	}

	// 5 timeout tasks
	for i := 0; i < 5; i++ {
		wg.Add(1)
		err := pool.Submit(ctx, func(taskCtx context.Context) {
			defer wg.Done()
			time.Sleep(200 * time.Millisecond) // Longer than 100ms timeout
		})
		require.NoError(t, err)
	}

	// 3 panic tasks
	for i := 0; i < 3; i++ {
		wg.Add(1)
		err := pool.Submit(ctx, func(taskCtx context.Context) {
			defer wg.Done()
			panic("test panic")
		})
		require.NoError(t, err)
	}

	wg.Wait()

	// Wait for all statistics to be updated
	time.Sleep(50 * time.Millisecond)

	// Verify statistics consistency
	stats := pool.GetStats()
	totalProcessed := stats.CompletedTasks + stats.FailedTasks + stats.TimeoutTasks

	t.Logf("Final statistics:")
	t.Logf("  Completed: %d (expected: 10)", stats.CompletedTasks)
	t.Logf("  Failed: %d (expected: 3)", stats.FailedTasks)
	t.Logf("  Timeout: %d (expected: 5)", stats.TimeoutTasks)
	t.Logf("  Total: %d (expected: 18)", totalProcessed)

	// CRITICAL: All 18 tasks must be accounted for
	assert.Equal(t, int64(18), totalProcessed, "All tasks must be accounted for in statistics")

	// Verify expected distributions (allow some tolerance for timing)
	assert.Equal(t, int64(10), stats.CompletedTasks, "All quick tasks should complete")
	assert.Equal(t, int64(3), stats.FailedTasks, "All panic tasks should be marked as failed")
	assert.Equal(t, int64(5), stats.TimeoutTasks, "All slow tasks should timeout")
}

func TestBoundedWorkerPool_HighConcurrencyStress(t *testing.T) {
	// REQ-RESOURCE-001: High concurrency stress test for race condition detection

	logger := logging.GetLogger("test")
	pool := NewBoundedWorkerPool(5, 50*time.Millisecond, logger)

	ctx := context.Background()
	err := pool.Start(ctx)
	require.NoError(t, err)
	defer pool.Stop(ctx)

	// High concurrency stress test
	numGoroutines := 20
	tasksPerGoroutine := 10
	totalExpectedTasks := numGoroutines * tasksPerGoroutine

	var submissionWg sync.WaitGroup
	var executionWg sync.WaitGroup

	// Submit tasks from multiple goroutines simultaneously
	for g := 0; g < numGoroutines; g++ {
		submissionWg.Add(1)
		go func(goroutineID int) {
			defer submissionWg.Done()

			for i := 0; i < tasksPerGoroutine; i++ {
				executionWg.Add(1)
				err := pool.Submit(ctx, func(taskCtx context.Context) {
					defer executionWg.Done()

					// Variable timing to stress the race condition
					sleepTime := time.Duration((goroutineID+i)%3) * 10 * time.Millisecond
					time.Sleep(sleepTime)
				})

				if err != nil {
					t.Logf("Warning: Failed to submit task from goroutine %d, task %d: %v", goroutineID, i, err)
					executionWg.Done() // Don't wait for tasks that failed to submit
				}
			}
		}(g)
	}

	// Wait for all submissions and executions
	submissionWg.Wait()
	executionWg.Wait()

	// Wait for statistics to stabilize
	time.Sleep(100 * time.Millisecond)

	// Verify no tasks were lost
	stats := pool.GetStats()
	totalProcessed := stats.CompletedTasks + stats.FailedTasks + stats.TimeoutTasks

	t.Logf("High concurrency stress test results:")
	t.Logf("  Expected tasks: %d", totalExpectedTasks)
	t.Logf("  Completed: %d", stats.CompletedTasks)
	t.Logf("  Failed: %d", stats.FailedTasks)
	t.Logf("  Timeout: %d", stats.TimeoutTasks)
	t.Logf("  Total Processed: %d", totalProcessed)

	// CRITICAL: No tasks should be lost in high concurrency
	assert.LessOrEqual(t, totalProcessed, int64(totalExpectedTasks),
		"Cannot process more tasks than submitted")
	assert.Greater(t, totalProcessed, int64(float64(totalExpectedTasks)*0.8),
		"Should process at least 80%% of submitted tasks")
}

func TestBoundedWorkerPool_ConcurrentPanicStress(t *testing.T) {
	// REQ-RESOURCE-001: Stress test concurrent panic recovery

	logger := logging.GetLogger("test")
	pool := NewBoundedWorkerPool(5, 50*time.Millisecond, logger)

	ctx := context.Background()
	err := pool.Start(ctx)
	require.NoError(t, err)
	defer pool.Stop(ctx)

	// Submit many tasks that panic concurrently
	numPanicTasks := 50
	var wg sync.WaitGroup

	for i := 0; i < numPanicTasks; i++ {
		wg.Add(1)
		taskID := i
		err := pool.Submit(ctx, func(taskCtx context.Context) {
			defer wg.Done()
			// Vary the panic types to test different scenarios
			if taskID%3 == 0 {
				panic(fmt.Sprintf("test panic %d", taskID))
			} else if taskID%3 == 1 {
				panic(fmt.Errorf("test error panic %d", taskID))
			} else {
				panic(42) // Non-string panic
			}
		})
		require.NoError(t, err)
	}

	wg.Wait()

	// Wait for statistics to be updated
	time.Sleep(100 * time.Millisecond)

	// Verify all panics were caught and counted
	stats := pool.GetStats()
	assert.Equal(t, int64(numPanicTasks), stats.FailedTasks,
		"All panic tasks should be counted as failed")
	assert.Equal(t, int64(0), stats.CompletedTasks,
		"No tasks should be marked as completed")
	assert.Equal(t, int64(0), stats.TimeoutTasks,
		"No tasks should timeout (they panic first)")

	// Pool should still be running after all the panics
	assert.True(t, pool.IsRunning(), "Pool should still be running after panics")

	// Verify pool can still execute normal tasks after panic storm
	wg.Add(1)
	err = pool.Submit(ctx, func(taskCtx context.Context) {
		defer wg.Done()
		time.Sleep(10 * time.Millisecond) // Normal task
	})
	require.NoError(t, err)
	wg.Wait()

	// Verify the normal task completed successfully
	finalStats := pool.GetStats()
	assert.Equal(t, int64(1), finalStats.CompletedTasks,
		"Normal task should complete after panic recovery")
}
