/*
Bounded Worker Pool Implementation

Requirements Coverage:
- REQ-CAM-001: Camera device discovery and enumeration
- REQ-CAM-002: Real-time device status monitoring
- REQ-CAM-003: Device capability probing and format detection

Test Categories: Unit/Integration
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package camera

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
)

// DefaultBoundedWorkerPool implements BoundedWorkerPool interface with resource limits
type DefaultBoundedWorkerPool struct {
	maxWorkers  int
	taskTimeout time.Duration
	semaphore   chan struct{}
	wg          sync.WaitGroup
	logger      *logging.Logger

	// Statistics (atomic)
	activeWorkers  int64
	queuedTasks    int64
	completedTasks int64
	failedTasks    int64
	timeoutTasks   int64

	// State management
	running  int32 // Atomic flag
	stopChan chan struct{}
	stopOnce sync.Once // Ensures single channel close operation
}

// NewBoundedWorkerPool creates a new bounded worker pool
func NewBoundedWorkerPool(maxWorkers int, taskTimeout time.Duration, logger *logging.Logger) BoundedWorkerPool {
	if maxWorkers <= 0 {
		maxWorkers = 10 // Default
	}
	if taskTimeout <= 0 {
		taskTimeout = 5 * time.Second // Default
	}
	if logger == nil {
		logger = logging.GetLogger("worker-pool")
	}

	return &DefaultBoundedWorkerPool{
		maxWorkers:  maxWorkers,
		taskTimeout: taskTimeout,
		semaphore:   make(chan struct{}, maxWorkers),
		logger:      logger,
		stopChan:    make(chan struct{}),
	}
}

// Submit submits a task to the worker pool (implements BoundedWorkerPool)
func (pool *DefaultBoundedWorkerPool) Submit(ctx context.Context, task func(context.Context)) error {
	// Check if pool is running
	if atomic.LoadInt32(&pool.running) == 0 {
		return fmt.Errorf("worker pool is not running")
	}

	atomic.AddInt64(&pool.queuedTasks, 1)
	defer atomic.AddInt64(&pool.queuedTasks, -1)

	// Try to acquire semaphore with context timeout
	select {
	case pool.semaphore <- struct{}{}:
		// Successfully acquired semaphore
		atomic.AddInt64(&pool.activeWorkers, 1)

		pool.wg.Add(1)
		go pool.executeTask(ctx, task)

		return nil
	case <-ctx.Done():
		atomic.AddInt64(&pool.failedTasks, 1)
		return fmt.Errorf("failed to submit task: %w", ctx.Err())
	case <-pool.stopChan:
		atomic.AddInt64(&pool.failedTasks, 1)
		return fmt.Errorf("worker pool is shutting down")
	}
}

// executeTask executes a single task with timeout and error handling
func (pool *DefaultBoundedWorkerPool) executeTask(ctx context.Context, task func(context.Context)) {
	defer func() {
		// Always release resources
		atomic.AddInt64(&pool.activeWorkers, -1)
		<-pool.semaphore // Release semaphore
		pool.wg.Done()

		// Recover from panics
		if r := recover(); r != nil {
			atomic.AddInt64(&pool.failedTasks, 1)
			pool.logger.WithFields(logging.Fields{
				"panic":  r,
				"action": "task_panic_recovered",
			}).Error("Task panicked in worker pool")
		}
	}()

	// Create timeout context for the task
	taskCtx, cancel := context.WithTimeout(ctx, pool.taskTimeout)
	defer cancel()

	// Execute task with proper result tracking (single source of truth)
	type taskResult struct {
		completed bool
		panicked  bool
		panic     interface{}
		timedOut  bool
	}

	resultChan := make(chan taskResult, 1)

	go func() {
		var result taskResult

		defer func() {
			if r := recover(); r != nil {
				result.panicked = true
				result.panic = r
			}
			// Always send result - single source of truth
			resultChan <- result
		}()

		// Execute task with timeout awareness
		taskDone := make(chan struct{})
		go func() {
			defer func() {
				// Recover from panics in the task execution goroutine
				if r := recover(); r != nil {
					result.panicked = true
					result.panic = r
				}
				// Always close the channel to signal completion
				close(taskDone)
			}()
			task(taskCtx)
		}()

		select {
		case <-taskDone:
			result.completed = true
		case <-taskCtx.Done():
			result.timedOut = true
		}
	}()

	// Single classification point - no race conditions
	var result taskResult
	select {
	case result = <-resultChan:
		// Task finished - classify the result
		if result.panicked {
			atomic.AddInt64(&pool.failedTasks, 1)
			pool.logger.WithFields(logging.Fields{
				"panic":  result.panic,
				"action": "task_panic",
			}).Error("Task panicked during execution")
		} else if result.timedOut {
			atomic.AddInt64(&pool.timeoutTasks, 1)
			pool.logger.WithFields(logging.Fields{
				"timeout": pool.taskTimeout,
				"action":  "task_timeout",
			}).Warn("Task timed out in worker pool")
		} else if result.completed {
			atomic.AddInt64(&pool.completedTasks, 1)
		} else {
			// Fallback - should not happen
			atomic.AddInt64(&pool.failedTasks, 1)
			pool.logger.Warn("Task finished with unknown result")
		}
	case <-pool.stopChan:
		// Pool shutdown - wait for result respecting context timeout
		select {
		case result = <-resultChan:
			// Task finished during shutdown
			if result.panicked {
				atomic.AddInt64(&pool.failedTasks, 1)
			} else if result.completed {
				atomic.AddInt64(&pool.completedTasks, 1)
			} else {
				atomic.AddInt64(&pool.failedTasks, 1)
			}
		case <-ctx.Done():
			// Context timeout - respect caller's timeout instead of hardcoded 100ms
			atomic.AddInt64(&pool.failedTasks, 1)
			pool.logger.Debug("Task cancelled due to context timeout during shutdown")
		}
	}
}

// Start starts the worker pool (implements ResourceManager)
func (pool *DefaultBoundedWorkerPool) Start(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&pool.running, 0, 1) {
		return fmt.Errorf("worker pool is already running")
	}

	pool.logger.WithFields(logging.Fields{
		"max_workers":  pool.maxWorkers,
		"task_timeout": pool.taskTimeout,
	}).Info("Bounded worker pool started")

	return nil
}

// Stop gracefully stops the worker pool (implements BoundedWorkerPool and ResourceManager)
func (pool *DefaultBoundedWorkerPool) Stop(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&pool.running, 1, 0) {
		pool.logger.Debug("Worker pool is already stopped")
		return nil // Idempotent
	}

	pool.logger.Info("Stopping bounded worker pool...")

	// Signal stop to all workers - use sync.Once to prevent double-close
	pool.stopOnce.Do(func() {
		close(pool.stopChan)
	})

	// Wait for all workers to complete with timeout
	done := make(chan struct{})
	go func() {
		pool.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		pool.logger.Info("All worker pool tasks completed gracefully")
	case <-ctx.Done():
		pool.logger.Warn("Worker pool shutdown timeout, some tasks may have been interrupted")
		return ctx.Err()
	}

	stats := pool.GetStats()
	pool.logger.WithFields(logging.Fields{
		"completed_tasks": stats.CompletedTasks,
		"failed_tasks":    stats.FailedTasks,
		"timeout_tasks":   stats.TimeoutTasks,
	}).Info("Bounded worker pool stopped")

	return nil
}

// IsRunning returns whether the worker pool is running (implements ResourceManager)
func (pool *DefaultBoundedWorkerPool) IsRunning() bool {
	return atomic.LoadInt32(&pool.running) == 1
}

// GetStats returns current worker pool statistics (implements BoundedWorkerPool)
func (pool *DefaultBoundedWorkerPool) GetStats() WorkerPoolStats {
	return WorkerPoolStats{
		ActiveWorkers:  int(atomic.LoadInt64(&pool.activeWorkers)),
		QueuedTasks:    int(atomic.LoadInt64(&pool.queuedTasks)),
		CompletedTasks: atomic.LoadInt64(&pool.completedTasks),
		FailedTasks:    atomic.LoadInt64(&pool.failedTasks),
		TimeoutTasks:   atomic.LoadInt64(&pool.timeoutTasks),
		MaxWorkers:     pool.maxWorkers,
	}
}
