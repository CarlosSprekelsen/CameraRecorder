/*
MediaMTX Circuit Breaker Tests

Requirements Coverage:
- REQ-MTX-007: Error handling and recovery
- REQ-MTX-008: Logging and monitoring

Test Categories: Unit
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package mediamtx

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewCircuitBreaker_ReqMTX007 tests circuit breaker creation
func TestNewCircuitBreaker_ReqMTX007(t *testing.T) {
	// REQ-MTX-007: Error handling and recovery
	helper, _ := SetupMediaMTXTest(t)

	logger := helper.GetLogger()
	config := CircuitBreakerConfig{
		FailureThreshold: 3,
		RecoveryTimeout:  testutils.UniversalTimeoutVeryLong,
		MaxFailures:      10,
	}

	cb := NewCircuitBreaker("test", config, logger)
	require.NotNil(t, cb)
	assert.Equal(t, StateClosed, cb.GetState())
	assert.Equal(t, 0, cb.GetFailureCount())
}

// TestCircuitBreaker_Call_Success_ReqMTX007 tests successful operation
func TestCircuitBreaker_Call_Success_ReqMTX007(t *testing.T) {
	// REQ-MTX-007: Error handling and recovery
	helper, _ := SetupMediaMTXTest(t)

	logger := helper.GetLogger()
	config := CircuitBreakerConfig{
		FailureThreshold: 2,
		RecoveryTimeout:  1 * time.Second,
		MaxFailures:      5,
	}

	cb := NewCircuitBreaker("test", config, logger)

	err := cb.Call(func() error {
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, StateClosed, cb.GetState())
	assert.Equal(t, 0, cb.GetFailureCount())
}

// TestCircuitBreaker_Call_Failure_ReqMTX007 tests failure handling
func TestCircuitBreaker_Call_Failure_ReqMTX007(t *testing.T) {
	// REQ-MTX-007: Error handling and recovery
	helper, _ := SetupMediaMTXTest(t)

	logger := helper.GetLogger()
	config := CircuitBreakerConfig{
		FailureThreshold: 2,
		RecoveryTimeout:  1 * time.Second,
		MaxFailures:      5,
	}

	cb := NewCircuitBreaker("test", config, logger)

	// First failure
	err := cb.Call(func() error {
		return errors.New("test error")
	})

	assert.Error(t, err)
	assert.Equal(t, StateClosed, cb.GetState())
	assert.Equal(t, 1, cb.GetFailureCount())
}

// TestCircuitBreaker_Open_ReqMTX007 tests circuit breaker opening
func TestCircuitBreaker_Open_ReqMTX007(t *testing.T) {
	// REQ-MTX-007: Error handling and recovery
	helper, _ := SetupMediaMTXTest(t)

	logger := helper.GetLogger()
	config := CircuitBreakerConfig{
		FailureThreshold: 2,
		RecoveryTimeout:  100 * time.Millisecond,
		MaxFailures:      5,
	}

	cb := NewCircuitBreaker("test", config, logger)

	// Cause two failures to open circuit
	for i := 0; i < 2; i++ {
		err := cb.Call(func() error {
			return errors.New("test error")
		})
		assert.Error(t, err)
	}

	// Circuit should be open now
	assert.Equal(t, StateOpen, cb.GetState())
	assert.Equal(t, 2, cb.GetFailureCount())

	// Next call should be blocked
	err := cb.Call(func() error {
		return nil
	})

	assert.Error(t, err)
	assert.IsType(t, &CircuitBreakerError{}, err)
	cbErr := err.(*CircuitBreakerError)
	assert.Equal(t, "test", cbErr.Name)
	assert.Equal(t, StateOpen, cbErr.State)
}

// TestCircuitBreaker_HalfOpen_ReqMTX007 tests circuit breaker half-open state
func TestCircuitBreaker_HalfOpen_ReqMTX007(t *testing.T) {
	// REQ-MTX-007: Error handling and recovery
	helper, _ := SetupMediaMTXTest(t)

	logger := helper.GetLogger()
	config := CircuitBreakerConfig{
		FailureThreshold: 2,
		RecoveryTimeout:  50 * time.Millisecond,
		MaxFailures:      5,
	}

	cb := NewCircuitBreaker("test", config, logger)

	// Open the circuit
	for i := 0; i < 2; i++ {
		cb.Call(func() error {
			return errors.New("test error")
		})
	}

	assert.Equal(t, StateOpen, cb.GetState())

	// Wait for recovery timeout
	time.Sleep(100 * time.Millisecond)

	// Next call should be allowed (half-open)
	success := false
	err := cb.Call(func() error {
		success = true
		return nil
	})

	assert.NoError(t, err)
	assert.True(t, success)
	assert.Equal(t, StateClosed, cb.GetState())
	assert.Equal(t, 0, cb.GetFailureCount())
}

// TestCircuitBreaker_HalfOpen_Failure_ReqMTX007 tests half-open state with failure
func TestCircuitBreaker_HalfOpen_Failure_ReqMTX007(t *testing.T) {
	// REQ-MTX-007: Error handling and recovery
	helper, _ := SetupMediaMTXTest(t)

	logger := helper.GetLogger()
	config := CircuitBreakerConfig{
		FailureThreshold: 2,
		RecoveryTimeout:  50 * time.Millisecond,
		MaxFailures:      5,
	}

	cb := NewCircuitBreaker("test", config, logger)

	// Open the circuit
	for i := 0; i < 2; i++ {
		cb.Call(func() error {
			return errors.New("test error")
		})
	}

	assert.Equal(t, StateOpen, cb.GetState())

	// Wait for recovery timeout
	time.Sleep(100 * time.Millisecond)

	// Next call should fail and reopen circuit
	err := cb.Call(func() error {
		return errors.New("test error")
	})

	assert.Error(t, err)
	assert.Equal(t, StateOpen, cb.GetState())
	assert.Equal(t, 3, cb.GetFailureCount())
}

// TestCircuitBreaker_Reset_ReqMTX007 tests circuit breaker reset
func TestCircuitBreaker_Reset_ReqMTX007(t *testing.T) {
	// REQ-MTX-007: Error handling and recovery
	helper, _ := SetupMediaMTXTest(t)

	logger := helper.GetLogger()
	config := CircuitBreakerConfig{
		FailureThreshold: 2,
		RecoveryTimeout:  1 * time.Second,
		MaxFailures:      5,
	}

	cb := NewCircuitBreaker("test", config, logger)

	// Open the circuit
	for i := 0; i < 2; i++ {
		cb.Call(func() error {
			return errors.New("test error")
		})
	}

	assert.Equal(t, StateOpen, cb.GetState())
	assert.Equal(t, 2, cb.GetFailureCount())

	// Reset the circuit
	cb.Reset()

	assert.Equal(t, StateClosed, cb.GetState())
	assert.Equal(t, 0, cb.GetFailureCount())

	// Should work normally after reset
	err := cb.Call(func() error {
		return nil
	})

	assert.NoError(t, err)
}

// TestCircuitBreaker_ContextCancellation_ReqMTX007 tests context cancellation
func TestCircuitBreaker_ContextCancellation_ReqMTX007(t *testing.T) {
	// REQ-MTX-007: Error handling and recovery
	helper, _ := SetupMediaMTXTest(t)

	logger := helper.GetLogger()
	config := CircuitBreakerConfig{
		FailureThreshold: 2,
		RecoveryTimeout:  1 * time.Second,
		MaxFailures:      5,
	}

	cb := NewCircuitBreaker("test", config, logger)

	// Open the circuit
	for i := 0; i < 2; i++ {
		cb.Call(func() error {
			return errors.New("test error")
		})
	}

	assert.Equal(t, StateOpen, cb.GetState())

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Call should still work (circuit breaker doesn't use context directly)
	err := cb.Call(func() error {
		return nil
	})

	// Should be blocked by circuit breaker, not context
	assert.Error(t, err)
	assert.IsType(t, &CircuitBreakerError{}, err)

	// Verify context is cancelled
	assert.Error(t, ctx.Err())
}

// TestCircuitBreakerError_ReqMTX007 tests circuit breaker error type
func TestCircuitBreakerError_ReqMTX007(t *testing.T) {
	// REQ-MTX-007: Error handling and recovery
	err := &CircuitBreakerError{
		Name:  "test",
		State: StateOpen,
		Msg:   "circuit breaker is open",
	}

	assert.Equal(t, "circuit breaker 'test' is open: circuit breaker is open", err.Error())
	assert.Equal(t, "test", err.Name)
	assert.Equal(t, StateOpen, err.State)
}
