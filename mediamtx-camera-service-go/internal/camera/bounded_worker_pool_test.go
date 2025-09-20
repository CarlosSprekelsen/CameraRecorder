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
