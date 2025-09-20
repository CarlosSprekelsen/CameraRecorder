/*
Stoppable Interface Unit Tests

Tests the Stoppable interface and helper functions for context-aware shutdown.

Test Categories: Unit
*/

package common

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockStoppable is a test implementation of the Stoppable interface
type mockStoppable struct {
	stopFunc func(ctx context.Context) error
	mu       sync.RWMutex
	running  bool
}

func newMockStoppable(stopFunc func(ctx context.Context) error) *mockStoppable {
	return &mockStoppable{
		stopFunc: stopFunc,
		running:  true,
	}
}

func (m *mockStoppable) Stop(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return nil
	}

	m.running = false
	if m.stopFunc != nil {
		return m.stopFunc(ctx)
	}
	return nil
}

func (m *mockStoppable) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.running
}

// TestStoppable_InterfaceCompliance tests that our mock implements the Stoppable interface
func TestStoppable_InterfaceCompliance(t *testing.T) {
	var _ Stoppable = (*mockStoppable)(nil)

	// Test that the interface can be used
	mock := newMockStoppable(nil)
	ctx := context.Background()
	err := mock.Stop(ctx)
	assert.NoError(t, err)
	assert.False(t, mock.IsRunning())
}

// TestStoppable_GracefulShutdown tests graceful shutdown scenarios
func TestStoppable_GracefulShutdown(t *testing.T) {
	t.Run("successful_shutdown", func(t *testing.T) {
		mock := newMockStoppable(nil)
		assert.True(t, mock.IsRunning())

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		start := time.Now()
		err := mock.Stop(ctx)
		elapsed := time.Since(start)

		require.NoError(t, err, "Stop should succeed")
		assert.False(t, mock.IsRunning(), "Should not be running after stop")
		assert.Less(t, elapsed, 100*time.Millisecond, "Shutdown should be fast")
	})

	t.Run("shutdown_with_error", func(t *testing.T) {
		expectedError := errors.New("stop failed")
		mock := newMockStoppable(func(ctx context.Context) error {
			return expectedError
		})

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := mock.Stop(ctx)
		require.Error(t, err, "Stop should return error")
		assert.Equal(t, expectedError, err, "Should return the expected error")
		assert.False(t, mock.IsRunning(), "Should not be running even after error")
	})

	t.Run("shutdown_with_context_cancellation", func(t *testing.T) {
		mock := newMockStoppable(func(ctx context.Context) error {
			// Simulate work that checks context
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(100 * time.Millisecond):
				return nil
			}
		})

		// Cancel context immediately
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		start := time.Now()
		err := mock.Stop(ctx)
		elapsed := time.Since(start)

		require.Error(t, err, "Stop should return context error")
		assert.Contains(t, err.Error(), "context canceled", "Should indicate context cancellation")
		assert.Less(t, elapsed, 50*time.Millisecond, "Should be fast with cancelled context")
	})

	t.Run("shutdown_with_timeout", func(t *testing.T) {
		mock := newMockStoppable(func(ctx context.Context) error {
			// Simulate work that takes longer than timeout
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(200 * time.Millisecond):
				return nil
			}
		})

		// Use very short timeout
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		start := time.Now()
		err := mock.Stop(ctx)
		elapsed := time.Since(start)

		require.Error(t, err, "Stop should timeout")
		assert.Contains(t, err.Error(), "context deadline exceeded", "Should indicate timeout")
		assert.Less(t, elapsed, 100*time.Millisecond, "Should timeout quickly")
	})

	t.Run("double_stop_handling", func(t *testing.T) {
		stopCount := 0
		mock := newMockStoppable(func(ctx context.Context) error {
			stopCount++
			return nil
		})

		// First stop
		ctx1, cancel1 := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel1()
		err := mock.Stop(ctx1)
		require.NoError(t, err, "First stop should succeed")
		assert.Equal(t, 1, stopCount, "Stop function should be called once")

		// Second stop
		ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel2()
		err = mock.Stop(ctx2)
		require.NoError(t, err, "Second stop should succeed")
		assert.Equal(t, 1, stopCount, "Stop function should not be called again")
	})
}

// TestStoppable_ConcurrentAccess tests concurrent access to Stoppable
func TestStoppable_ConcurrentAccess(t *testing.T) {
	t.Run("concurrent_stop_calls", func(t *testing.T) {
		stopCount := 0
		mock := newMockStoppable(func(ctx context.Context) error {
			stopCount++
			time.Sleep(10 * time.Millisecond) // Simulate some work
			return nil
		})

		// Start multiple goroutines calling Stop concurrently
		var wg sync.WaitGroup
		numGoroutines := 10

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				mock.Stop(ctx)
			}()
		}

		wg.Wait()

		// Only one stop should have been executed
		assert.Equal(t, 1, stopCount, "Stop function should be called only once")
		assert.False(t, mock.IsRunning(), "Should not be running after concurrent stops")
	})

	t.Run("concurrent_is_running_checks", func(t *testing.T) {
		mock := newMockStoppable(nil)

		// Start goroutine that calls IsRunning repeatedly
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				mock.IsRunning()
				time.Sleep(1 * time.Millisecond)
			}
		}()

		// Stop the mock while IsRunning is being called
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := mock.Stop(ctx)

		wg.Wait()

		require.NoError(t, err, "Stop should succeed")
		assert.False(t, mock.IsRunning(), "Should not be running after stop")
	})
}

// TestStoppable_RealWorldScenarios tests realistic shutdown scenarios
func TestStoppable_RealWorldScenarios(t *testing.T) {
	t.Run("service_with_cleanup_work", func(t *testing.T) {
		cleanupDone := false
		mock := newMockStoppable(func(ctx context.Context) error {
			// Simulate cleanup work that takes some time
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(50 * time.Millisecond):
				cleanupDone = true
				return nil
			}
		})

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		err := mock.Stop(ctx)
		require.NoError(t, err, "Stop should succeed")
		assert.True(t, cleanupDone, "Cleanup should be completed")
	})

	t.Run("service_that_ignores_context", func(t *testing.T) {
		mock := newMockStoppable(func(ctx context.Context) error {
			// Simulate service that doesn't respect context (bad practice)
			// This will block until the context times out
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(500 * time.Millisecond):
				// This should never be reached due to timeout
				return nil
			}
		})

		// Use very short timeout
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		start := time.Now()
		err := mock.Stop(ctx)
		elapsed := time.Since(start)

		// Should timeout but not hang indefinitely
		require.Error(t, err, "Should timeout")
		assert.Contains(t, err.Error(), "context deadline exceeded", "Should indicate timeout")
		assert.Less(t, elapsed, 100*time.Millisecond, "Should timeout quickly")
	})
}

// TestStopWithTimeout tests the helper function
func TestStopWithTimeout(t *testing.T) {
	t.Run("successful_stop_with_timeout", func(t *testing.T) {
		mock := newMockStoppable(nil)

		start := time.Now()
		err := StopWithTimeout(mock, 5*time.Second)
		elapsed := time.Since(start)

		require.NoError(t, err, "StopWithTimeout should succeed")
		assert.False(t, mock.IsRunning(), "Should not be running after stop")
		assert.Less(t, elapsed, 100*time.Millisecond, "Should complete quickly")
	})

	t.Run("stop_with_timeout_error", func(t *testing.T) {
		expectedError := errors.New("stop failed")
		mock := newMockStoppable(func(ctx context.Context) error {
			return expectedError
		})

		err := StopWithTimeout(mock, 5*time.Second)
		require.Error(t, err, "StopWithTimeout should return error")
		assert.Equal(t, expectedError, err, "Should return the expected error")
	})

	t.Run("stop_with_timeout_exceeds", func(t *testing.T) {
		mock := newMockStoppable(func(ctx context.Context) error {
			// Simulate work that takes longer than timeout
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(200 * time.Millisecond):
				return nil
			}
		})

		start := time.Now()
		err := StopWithTimeout(mock, 50*time.Millisecond)
		elapsed := time.Since(start)

		require.Error(t, err, "StopWithTimeout should timeout")
		assert.Contains(t, err.Error(), "context deadline exceeded", "Should indicate timeout")
		assert.Less(t, elapsed, 100*time.Millisecond, "Should timeout quickly")
	})
}
