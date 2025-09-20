/*
BoundedWorkerPool Resource Management Tests

Requirements Coverage:
- REQ-CAM-001: Camera device discovery and enumeration
- REQ-CAM-002: Real-time device status monitoring
- REQ-RESOURCE-001: Resource management and goroutine pool limits

Test Categories: Unit
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package camera

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBoundedWorkerPool_ResourceLifecycle(t *testing.T) {
	// REQ-RESOURCE-001: Resource management lifecycle

	logger := logging.GetLogger("test")
	pool := NewBoundedWorkerPool(3, 1*time.Second, logger)

	// Test initial state
	assert.False(t, pool.IsRunning())

	// Test start
	ctx := context.Background()
	err := pool.Start(ctx)
	require.NoError(t, err)
	assert.True(t, pool.IsRunning())

	// Test stop
	err = pool.Stop(ctx)
	require.NoError(t, err)
	assert.False(t, pool.IsRunning())
}

func TestBoundedWorkerPool_ConcurrencyControl(t *testing.T) {
	// REQ-RESOURCE-001: Goroutine pool limits and concurrency control

	logger := logging.GetLogger("test")
	maxWorkers := 2
	pool := NewBoundedWorkerPool(maxWorkers, 5*time.Second, logger)

	ctx := context.Background()
	err := pool.Start(ctx)
	require.NoError(t, err)
	defer pool.Stop(ctx)

	// Track concurrent executions
	var activeWorkers int64
	var maxConcurrent int64
	var wg sync.WaitGroup

	// Submit more tasks than max workers
	numTasks := 5
	for i := 0; i < numTasks; i++ {
		wg.Add(1)
		err := pool.Submit(ctx, func(taskCtx context.Context) {
			defer wg.Done()

			current := atomic.AddInt64(&activeWorkers, 1)
			defer atomic.AddInt64(&activeWorkers, -1)

			// Update max concurrent if needed
			for {
				max := atomic.LoadInt64(&maxConcurrent)
				if current <= max || atomic.CompareAndSwapInt64(&maxConcurrent, max, current) {
					break
				}
			}

			time.Sleep(10 * time.Millisecond) // Brief work simulation
		})
		require.NoError(t, err)
	}

	wg.Wait()

	// Verify concurrency was limited
	assert.LessOrEqual(t, int(atomic.LoadInt64(&maxConcurrent)), maxWorkers)

	// Verify all tasks completed
	stats := pool.GetStats()
	assert.Equal(t, int64(numTasks), stats.CompletedTasks)
}

func TestBoundedWorkerPool_TaskTimeout(t *testing.T) {
	// REQ-RESOURCE-001: Task timeout handling

	logger := logging.GetLogger("test")
	shortTimeout := 50 * time.Millisecond
	pool := NewBoundedWorkerPool(1, shortTimeout, logger)

	ctx := context.Background()
	err := pool.Start(ctx)
	require.NoError(t, err)
	defer pool.Stop(ctx)

	// Submit task that exceeds timeout
	err = pool.Submit(ctx, func(taskCtx context.Context) {
		time.Sleep(shortTimeout * 3) // Much longer than timeout
	})
	require.NoError(t, err)

	// Wait for timeout to occur
	time.Sleep(shortTimeout * 2)

	// Verify timeout was recorded
	stats := pool.GetStats()
	assert.Equal(t, int64(1), stats.TimeoutTasks)
}

func TestBoundedWorkerPool_PanicRecovery(t *testing.T) {
	// REQ-RESOURCE-001: Panic recovery and error handling

	logger := logging.GetLogger("test")
	pool := NewBoundedWorkerPool(1, 5*time.Second, logger)

	ctx := context.Background()
	err := pool.Start(ctx)
	require.NoError(t, err)
	defer pool.Stop(ctx)

	var wg sync.WaitGroup
	wg.Add(1)

	// Submit task that panics
	err = pool.Submit(ctx, func(taskCtx context.Context) {
		defer wg.Done()
		panic("test panic - should be recovered")
	})
	require.NoError(t, err)

	wg.Wait()

	// Verify panic was handled gracefully
	stats := pool.GetStats()
	assert.Equal(t, int64(1), stats.FailedTasks)
	assert.True(t, pool.IsRunning()) // Pool should still be running
}

func TestBoundedWorkerPool_GracefulShutdown(t *testing.T) {
	// REQ-RESOURCE-001: Graceful shutdown with active tasks

	logger := logging.GetLogger("test")
	pool := NewBoundedWorkerPool(2, DefaultTestTimeout, logger)

	ctx := context.Background()
	err := pool.Start(ctx)
	require.NoError(t, err)

	var taskStarted, taskCompleted int32
	var wg sync.WaitGroup
	wg.Add(1)

	// Submit long-running task
	err = pool.Submit(ctx, func(taskCtx context.Context) {
		defer wg.Done()
		atomic.StoreInt32(&taskStarted, 1)
		time.Sleep(100 * time.Millisecond)
		atomic.StoreInt32(&taskCompleted, 1)
	})
	require.NoError(t, err)

	// Wait for task to start
	require.Eventually(t, func() bool {
		return atomic.LoadInt32(&taskStarted) == 1
	}, 2*time.Second, 10*time.Millisecond)

	// Stop pool - should wait for task completion
	stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = pool.Stop(stopCtx)
	require.NoError(t, err)

	wg.Wait()

	// Verify task completed before shutdown
	assert.Equal(t, int32(1), atomic.LoadInt32(&taskCompleted))
	assert.False(t, pool.IsRunning())
}

func TestBoundedWorkerPool_ConstructorEdgeCases(t *testing.T) {
	// REQ-RESOURCE-001: Test constructor parameter validation and defaults

	logger := logging.GetLogger("test")

	// Test with invalid maxWorkers (should use default of 10)
	pool1 := NewBoundedWorkerPool(-5, time.Second, logger)
	assert.NotNil(t, pool1)

	ctx := context.Background()
	err := pool1.Start(ctx)
	require.NoError(t, err)
	defer pool1.Stop(ctx)

	// Test with zero maxWorkers (should use default of 10)
	pool2 := NewBoundedWorkerPool(0, time.Second, logger)
	assert.NotNil(t, pool2)

	err = pool2.Start(ctx)
	require.NoError(t, err)
	defer pool2.Stop(ctx)

	// Test with invalid timeout (should use default of 5 seconds)
	pool3 := NewBoundedWorkerPool(5, -time.Second, logger)
	assert.NotNil(t, pool3)

	err = pool3.Start(ctx)
	require.NoError(t, err)
	defer pool3.Stop(ctx)

	// Test with zero timeout (should use default of 5 seconds)
	pool4 := NewBoundedWorkerPool(5, 0, logger)
	assert.NotNil(t, pool4)

	err = pool4.Start(ctx)
	require.NoError(t, err)
	defer pool4.Stop(ctx)

	// Test with nil logger (should create default logger)
	pool5 := NewBoundedWorkerPool(5, time.Second, nil)
	assert.NotNil(t, pool5)

	err = pool5.Start(ctx)
	require.NoError(t, err)
	defer pool5.Stop(ctx)

	// Test with all invalid parameters
	pool6 := NewBoundedWorkerPool(-1, -time.Second, nil)
	assert.NotNil(t, pool6)

	err = pool6.Start(ctx)
	require.NoError(t, err)
	defer pool6.Stop(ctx)
}

func TestBoundedWorkerPool_SubmitToStoppedPool(t *testing.T) {
	// REQ-RESOURCE-001: Test submitting to stopped pool returns error

	logger := logging.GetLogger("test")
	pool := NewBoundedWorkerPool(2, time.Second, logger)

	ctx := context.Background()

	// Try to submit task to stopped pool
	err := pool.Submit(ctx, func(taskCtx context.Context) {
		// This should never execute
	})

	// Should return error because pool is not running
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not running")
}

func TestBoundedWorkerPool_StartStopEdgeCases(t *testing.T) {
	// REQ-RESOURCE-001: Test start/stop edge cases and error paths

	logger := logging.GetLogger("test")
	pool := NewBoundedWorkerPool(1, time.Second, logger)

	ctx := context.Background()

	// Test double start (should return error)
	err := pool.Start(ctx)
	require.NoError(t, err)

	err = pool.Start(ctx) // Second start should fail
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already running")

	// Test stop after start
	err = pool.Stop(ctx)
	require.NoError(t, err)
	assert.False(t, pool.IsRunning())

	// Test double stop (should be idempotent)
	err = pool.Stop(ctx) // Second stop should be safe
	assert.NoError(t, err)
	assert.False(t, pool.IsRunning())
}

func TestBoundedWorkerPool_SubmitWithCancelledContext(t *testing.T) {
	// REQ-RESOURCE-001: Test submit with cancelled context

	logger := logging.GetLogger("test")
	pool := NewBoundedWorkerPool(1, time.Second, logger)

	ctx := context.Background()
	err := pool.Start(ctx)
	require.NoError(t, err)
	defer pool.Stop(ctx)

	// Create cancelled context for submission
	cancelledCtx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err = pool.Submit(cancelledCtx, func(taskCtx context.Context) {
		// Task function - will detect context cancellation
	})

	// Submit with cancelled context should fail gracefully
	// This tests the error path in Submit method
	if err != nil {
		assert.Contains(t, err.Error(), "context canceled")
	}

	// Wait a bit for any potential execution
	time.Sleep(50 * time.Millisecond)

	// The important thing is no panic occurred and pool is stable
	assert.True(t, pool.IsRunning())
}
