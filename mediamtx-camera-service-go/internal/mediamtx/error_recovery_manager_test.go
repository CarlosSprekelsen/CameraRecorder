/*
MediaMTX Error Recovery Manager Tests

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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockRecoveryStrategy is a mock implementation of RecoveryStrategy for testing
type MockRecoveryStrategy struct {
	name           string
	canRecover     bool
	recoveryDelay  time.Duration
	recoveryResult error
}

func (m *MockRecoveryStrategy) CanRecover(ctx *ErrorContext, err error) bool {
	return m.canRecover
}

func (m *MockRecoveryStrategy) Recover(ctx context.Context, errorCtx *ErrorContext, err error) error {
	if m.recoveryDelay > 0 {
		time.Sleep(m.recoveryDelay)
	}
	return m.recoveryResult
}

func (m *MockRecoveryStrategy) GetRecoveryDelay() time.Duration {
	return m.recoveryDelay
}

func (m *MockRecoveryStrategy) GetStrategyName() string {
	return m.name
}

// TestNewErrorRecoveryManager_ReqMTX007 tests error recovery manager creation
func TestNewErrorRecoveryManager_ReqMTX007(t *testing.T) {
	// REQ-MTX-007: Error handling and recovery
	helper := SetupMediaMTXTestHelperOnly(t)

	logger := helper.GetLogger()
	erm := NewErrorRecoveryManager(logger)

	require.NotNil(t, erm)
	assert.NotNil(t, erm.strategies)
	assert.NotNil(t, erm.errorMetrics)
	assert.NotNil(t, erm.recoveryInProgress)
}

// TestErrorRecoveryManager_RegisterStrategy_ReqMTX007 tests strategy registration
func TestErrorRecoveryManager_RegisterStrategy_ReqMTX007(t *testing.T) {
	// REQ-MTX-007: Error handling and recovery
	helper := SetupMediaMTXTestHelperOnly(t)

	logger := helper.GetLogger()
	erm := NewErrorRecoveryManager(logger)

	strategy := &MockRecoveryStrategy{
		name: "test-strategy",
	}

	erm.RegisterStrategy(strategy)

	assert.Contains(t, erm.strategies, "test-strategy")
	assert.Equal(t, strategy, erm.strategies["test-strategy"])
}

// TestErrorRecoveryManager_HandleError_NoRecovery_ReqMTX007 tests error handling without recovery
func TestErrorRecoveryManager_HandleError_NoRecovery_ReqMTX007(t *testing.T) {
	// REQ-MTX-007: Error handling and recovery
	helper := SetupMediaMTXTestHelperOnly(t)

	logger := helper.GetLogger()
	erm := NewErrorRecoveryManager(logger)

	errorCtx := &ErrorContext{
		Component:   "TestComponent",
		Operation:   "TestOperation",
		Timestamp:   time.Now(),
		Severity:    SeverityError,
		Recoverable: false,
	}

	originalErr := errors.New("test error")
	err := erm.HandleError(context.Background(), errorCtx, originalErr)

	assert.Error(t, err)
	assert.Equal(t, originalErr, err)

	// Check metrics
	metrics := erm.GetMetrics()
	assert.Equal(t, int64(1), metrics.TotalErrors)
	assert.Equal(t, int64(1), metrics.ErrorsByComponent["TestComponent"])
	assert.Equal(t, int64(1), metrics.ErrorsBySeverity["error"])
}

// TestErrorRecoveryManager_HandleError_NoApplicableStrategy_ReqMTX007 tests error handling with no applicable strategies
func TestErrorRecoveryManager_HandleError_NoApplicableStrategy_ReqMTX007(t *testing.T) {
	// REQ-MTX-007: Error handling and recovery
	helper := SetupMediaMTXTestHelperOnly(t)

	logger := helper.GetLogger()
	erm := NewErrorRecoveryManager(logger)

	strategy := &MockRecoveryStrategy{
		name:       "test-strategy",
		canRecover: false, // Cannot recover
	}
	erm.RegisterStrategy(strategy)

	errorCtx := &ErrorContext{
		Component:   "TestComponent",
		Operation:   "TestOperation",
		Timestamp:   time.Now(),
		Severity:    SeverityError,
		Recoverable: true,
	}

	originalErr := errors.New("test error")
	err := erm.HandleError(context.Background(), errorCtx, originalErr)

	assert.Error(t, err)
	assert.Equal(t, originalErr, err)

	// Check metrics - no recovery attempts
	metrics := erm.GetMetrics()
	assert.Equal(t, int64(0), metrics.RecoveryAttempts)
}

// TestErrorRecoveryManager_HandleError_SuccessfulRecovery_ReqMTX007 tests successful error recovery
func TestErrorRecoveryManager_HandleError_SuccessfulRecovery_ReqMTX007(t *testing.T) {
	// REQ-MTX-007: Error handling and recovery
	helper := SetupMediaMTXTestHelperOnly(t)

	logger := helper.GetLogger()
	erm := NewErrorRecoveryManager(logger)

	strategy := &MockRecoveryStrategy{
		name:           "test-strategy",
		canRecover:     true,
		recoveryResult: nil, // Success
	}
	erm.RegisterStrategy(strategy)

	errorCtx := &ErrorContext{
		Component:   "TestComponent",
		Operation:   "TestOperation",
		Timestamp:   time.Now(),
		Severity:    SeverityError,
		Recoverable: true,
	}

	originalErr := errors.New("test error")
	err := erm.HandleError(context.Background(), errorCtx, originalErr)

	assert.NoError(t, err)

	// Check metrics
	metrics := erm.GetMetrics()
	assert.Equal(t, int64(1), metrics.TotalErrors)
	assert.Equal(t, int64(1), metrics.RecoveryAttempts)
	assert.Equal(t, int64(1), metrics.RecoverySuccesses)
	assert.Equal(t, int64(0), metrics.RecoveryFailures)
}

// TestErrorRecoveryManager_HandleError_FailedRecovery_ReqMTX007 tests failed error recovery
func TestErrorRecoveryManager_HandleError_FailedRecovery_ReqMTX007(t *testing.T) {
	// REQ-MTX-007: Error handling and recovery
	helper := SetupMediaMTXTestHelperOnly(t)

	logger := helper.GetLogger()
	erm := NewErrorRecoveryManager(logger)

	strategy := &MockRecoveryStrategy{
		name:           "test-strategy",
		canRecover:     true,
		recoveryResult: errors.New("recovery failed"),
	}
	erm.RegisterStrategy(strategy)

	errorCtx := &ErrorContext{
		Component:   "TestComponent",
		Operation:   "TestOperation",
		Timestamp:   time.Now(),
		Severity:    SeverityError,
		Recoverable: true,
	}

	originalErr := errors.New("test error")
	err := erm.HandleError(context.Background(), errorCtx, originalErr)

	assert.Error(t, err)
	assert.Equal(t, "recovery failed", err.Error())

	// Check metrics
	metrics := erm.GetMetrics()
	assert.Equal(t, int64(1), metrics.TotalErrors)
	assert.Equal(t, int64(1), metrics.RecoveryAttempts)
	assert.Equal(t, int64(0), metrics.RecoverySuccesses)
	assert.Equal(t, int64(1), metrics.RecoveryFailures)
}

// TestErrorRecoveryManager_HandleError_MultipleStrategies_ReqMTX007 tests multiple recovery strategies
func TestErrorRecoveryManager_HandleError_MultipleStrategies_ReqMTX007(t *testing.T) {
	// REQ-MTX-007: Error handling and recovery
	helper := SetupMediaMTXTestHelperOnly(t)

	logger := helper.GetLogger()
	erm := NewErrorRecoveryManager(logger)

	// First strategy fails
	strategy1 := &MockRecoveryStrategy{
		name:           "strategy-1",
		canRecover:     true,
		recoveryResult: errors.New("strategy 1 failed"),
	}
	erm.RegisterStrategy(strategy1)

	// Second strategy succeeds
	strategy2 := &MockRecoveryStrategy{
		name:           "strategy-2",
		canRecover:     true,
		recoveryResult: nil,
	}
	erm.RegisterStrategy(strategy2)

	errorCtx := &ErrorContext{
		Component:   "TestComponent",
		Operation:   "TestOperation",
		Timestamp:   time.Now(),
		Severity:    SeverityError,
		Recoverable: true,
	}

	originalErr := errors.New("test error")
	err := erm.HandleError(context.Background(), errorCtx, originalErr)

	assert.NoError(t, err)

	// Check metrics - should have 2 attempts, 1 success, 1 failure
	metrics := erm.GetMetrics()
	assert.Equal(t, int64(1), metrics.TotalErrors)
	assert.Equal(t, int64(2), metrics.RecoveryAttempts)
	assert.Equal(t, int64(1), metrics.RecoverySuccesses)
	assert.Equal(t, int64(1), metrics.RecoveryFailures)
}

// TestErrorRecoveryManager_HandleError_RecoveryInProgress_ReqMTX007 tests recovery in progress prevention
func TestErrorRecoveryManager_HandleError_RecoveryInProgress_ReqMTX007(t *testing.T) {
	// REQ-MTX-007: Error handling and recovery
	helper := SetupMediaMTXTestHelperOnly(t)

	logger := helper.GetLogger()
	erm := NewErrorRecoveryManager(logger)

	strategy := &MockRecoveryStrategy{
		name:          "test-strategy",
		canRecover:    true,
		recoveryDelay: 100 * time.Millisecond, // Slow recovery
	}
	erm.RegisterStrategy(strategy)

	errorCtx := &ErrorContext{
		Component:   "TestComponent",
		Operation:   "TestOperation",
		CameraID:    "camera0",
		Timestamp:   time.Now(),
		Severity:    SeverityError,
		Recoverable: true,
	}

	originalErr := errors.New("test error")

	// Start first recovery
	go func() {
		erm.HandleError(context.Background(), errorCtx, originalErr)
	}()

	// Wait a bit for first recovery to start
	time.Sleep(10 * time.Millisecond)

	// Second recovery should be skipped
	err := erm.HandleError(context.Background(), errorCtx, originalErr)
	assert.Error(t, err)
	assert.Equal(t, originalErr, err)
}

// TestErrorRecoveryManager_GetMetrics_ReqMTX008 tests metrics retrieval
func TestErrorRecoveryManager_GetMetrics_ReqMTX008(t *testing.T) {
	// REQ-MTX-008: Logging and monitoring
	helper := SetupMediaMTXTestHelperOnly(t)

	logger := helper.GetLogger()
	erm := NewErrorRecoveryManager(logger)

	// Record some errors
	errorCtx := &ErrorContext{
		Component:   "TestComponent",
		Operation:   "TestOperation",
		Timestamp:   time.Now(),
		Severity:    SeverityWarning,
		Recoverable: false,
	}

	erm.HandleError(context.Background(), errorCtx, errors.New("test error 1"))
	erm.HandleError(context.Background(), errorCtx, errors.New("test error 2"))

	metrics := erm.GetMetrics()

	assert.Equal(t, int64(2), metrics.TotalErrors)
	assert.Equal(t, int64(2), metrics.ErrorsByComponent["TestComponent"])
	assert.Equal(t, int64(2), metrics.ErrorsBySeverity["warning"])
	assert.False(t, metrics.LastErrorTime.IsZero())
}
