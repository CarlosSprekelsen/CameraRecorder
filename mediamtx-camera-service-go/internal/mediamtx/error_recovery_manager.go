/*
MediaMTX Error Recovery Manager Implementation

This package provides centralized error recovery management for the MediaMTX camera service.
It coordinates error handling, recovery strategies, and provides a unified interface
for error management across all components.

Requirements Coverage:
- REQ-MTX-007: Error handling and recovery
- REQ-MTX-008: Logging and monitoring

Test Categories: Unit/Integration
*/

package mediamtx

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
)

// ErrorSeverity represents the severity level of an error
type ErrorSeverity string

const (
	SeverityInfo     ErrorSeverity = "info"
	SeverityWarning  ErrorSeverity = "warning"
	SeverityError    ErrorSeverity = "error"
	SeverityCritical ErrorSeverity = "critical"
)

// ErrorContext represents contextual information about an error
type ErrorContext struct {
	Component   string            `json:"component"`
	Operation   string            `json:"operation"`
	CameraID    string            `json:"camera_id,omitempty"`
	PathName    string            `json:"path_name,omitempty"`
	Filename    string            `json:"filename,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Timestamp   time.Time         `json:"timestamp"`
	Severity    ErrorSeverity     `json:"severity"`
	Recoverable bool              `json:"recoverable"`
}

// RecoveryStrategy defines the interface for error recovery strategies
type RecoveryStrategy interface {
	CanRecover(ctx *ErrorContext, err error) bool
	Recover(ctx context.Context, errorCtx *ErrorContext, err error) error
	GetRecoveryDelay() time.Duration
	GetStrategyName() string
}

// ErrorRecoveryManager manages error recovery across all components
type ErrorRecoveryManager struct {
	logger             *logging.Logger
	strategies         map[string]RecoveryStrategy
	errorMetrics       *ErrorMetrics
	recoveryInProgress map[string]bool
	mutex              sync.RWMutex
}

// ErrorMetrics tracks error statistics
type ErrorMetrics struct {
	TotalErrors       int64            `json:"total_errors"`
	ErrorsByComponent map[string]int64 `json:"errors_by_component"`
	ErrorsBySeverity  map[string]int64 `json:"errors_by_severity"`
	RecoveryAttempts  int64            `json:"recovery_attempts"`
	RecoverySuccesses int64            `json:"recovery_successes"`
	RecoveryFailures  int64            `json:"recovery_failures"`
	LastErrorTime     time.Time        `json:"last_error_time"`
	LastRecoveryTime  time.Time        `json:"last_recovery_time"`
	mutex             sync.RWMutex
}

// NewErrorRecoveryManager creates a new error recovery manager
func NewErrorRecoveryManager(logger *logging.Logger) *ErrorRecoveryManager {
	return &ErrorRecoveryManager{
		logger:     logger,
		strategies: make(map[string]RecoveryStrategy),
		errorMetrics: &ErrorMetrics{
			ErrorsByComponent: make(map[string]int64),
			ErrorsBySeverity:  make(map[string]int64),
		},
		recoveryInProgress: make(map[string]bool),
	}
}

// RegisterStrategy registers a recovery strategy
func (erm *ErrorRecoveryManager) RegisterStrategy(strategy RecoveryStrategy) {
	erm.mutex.Lock()
	defer erm.mutex.Unlock()

	erm.strategies[strategy.GetStrategyName()] = strategy
	erm.logger.WithField("strategy", strategy.GetStrategyName()).Info("Recovery strategy registered")
}

// HandleError processes an error and attempts recovery
func (erm *ErrorRecoveryManager) HandleError(ctx context.Context, errorCtx *ErrorContext, err error) error {
	// Update metrics
	erm.updateErrorMetrics(errorCtx)

	// Log the error
	erm.logError(errorCtx, err)

	// Check if recovery is possible
	if !errorCtx.Recoverable {
		erm.logger.WithFields(logging.Fields{
			"component": errorCtx.Component,
			"operation": errorCtx.Operation,
			"severity":  errorCtx.Severity,
		}).Warn("Error marked as non-recoverable, skipping recovery attempts")
		return err
	}

	// Find applicable recovery strategies
	applicableStrategies := erm.findApplicableStrategies(errorCtx, err)
	if len(applicableStrategies) == 0 {
		erm.logger.WithFields(logging.Fields{
			"component": errorCtx.Component,
			"operation": errorCtx.Operation,
		}).Debug("No recovery strategies applicable for error")
		return err
	}

	// Attempt recovery with each strategy
	return erm.attemptRecovery(ctx, errorCtx, err, applicableStrategies)
}

// updateErrorMetrics updates error tracking metrics
func (erm *ErrorRecoveryManager) updateErrorMetrics(errorCtx *ErrorContext) {
	erm.errorMetrics.mutex.Lock()
	defer erm.errorMetrics.mutex.Unlock()

	erm.errorMetrics.TotalErrors++
	erm.errorMetrics.ErrorsByComponent[errorCtx.Component]++
	erm.errorMetrics.ErrorsBySeverity[string(errorCtx.Severity)]++
	erm.errorMetrics.LastErrorTime = time.Now()
}

// logError logs the error with appropriate level
func (erm *ErrorRecoveryManager) logError(errorCtx *ErrorContext, err error) {
	fields := logging.Fields{
		"component": errorCtx.Component,
		"operation": errorCtx.Operation,
		"severity":  errorCtx.Severity,
		"error":     err.Error(),
	}

	if errorCtx.CameraID != "" {
		fields["camera_id"] = errorCtx.CameraID
	}
	if errorCtx.PathName != "" {
		fields["path_name"] = errorCtx.PathName
	}
	if errorCtx.Filename != "" {
		fields["filename"] = errorCtx.Filename
	}

	switch errorCtx.Severity {
	case SeverityInfo:
		erm.logger.WithFields(fields).Info("Error occurred")
	case SeverityWarning:
		erm.logger.WithFields(fields).Warn("Error occurred")
	case SeverityError:
		erm.logger.WithFields(fields).Error("Error occurred")
	case SeverityCritical:
		erm.logger.WithFields(fields).Error("Critical error occurred")
	}
}

// findApplicableStrategies finds recovery strategies that can handle the error
func (erm *ErrorRecoveryManager) findApplicableStrategies(errorCtx *ErrorContext, err error) []RecoveryStrategy {
	erm.mutex.RLock()
	defer erm.mutex.RUnlock()

	var applicable []RecoveryStrategy
	for _, strategy := range erm.strategies {
		if strategy.CanRecover(errorCtx, err) {
			applicable = append(applicable, strategy)
		}
	}

	return applicable
}

// attemptRecovery attempts to recover from the error using applicable strategies
func (erm *ErrorRecoveryManager) attemptRecovery(ctx context.Context, errorCtx *ErrorContext, err error, strategies []RecoveryStrategy) error {
	recoveryKey := fmt.Sprintf("%s:%s:%s", errorCtx.Component, errorCtx.Operation, errorCtx.CameraID)

	// Check if recovery is already in progress
	erm.mutex.Lock()
	if erm.recoveryInProgress[recoveryKey] {
		erm.mutex.Unlock()
		erm.logger.WithFields(logging.Fields{
			"component": errorCtx.Component,
			"operation": errorCtx.Operation,
			"camera_id": errorCtx.CameraID,
		}).Debug("Recovery already in progress, skipping")
		return err
	}
	erm.recoveryInProgress[recoveryKey] = true
	erm.mutex.Unlock()

	// Clean up recovery state when done
	defer func() {
		erm.mutex.Lock()
		delete(erm.recoveryInProgress, recoveryKey)
		erm.mutex.Unlock()
	}()

	// Try each strategy in order - attempt ALL strategies for comprehensive recovery
	var lastErr error = err
	var hasSuccessfulRecovery bool = false

	for _, strategy := range strategies {
		// Update recovery metrics for each attempt
		erm.errorMetrics.mutex.Lock()
		erm.errorMetrics.RecoveryAttempts++
		erm.errorMetrics.mutex.Unlock()
		erm.logger.WithFields(logging.Fields{
			"component": errorCtx.Component,
			"operation": errorCtx.Operation,
			"strategy":  strategy.GetStrategyName(),
		}).Info("Attempting error recovery")

		// Apply recovery delay if specified
		if delay := strategy.GetRecoveryDelay(); delay > 0 {
			select {
			case <-time.After(delay):
				// Continue with recovery
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		// Attempt recovery
		recoveryErr := strategy.Recover(ctx, errorCtx, lastErr)
		if recoveryErr == nil {
			// Recovery successful - track success but continue with other strategies
			erm.errorMetrics.mutex.Lock()
			erm.errorMetrics.RecoverySuccesses++
			erm.errorMetrics.LastRecoveryTime = time.Now()
			erm.errorMetrics.mutex.Unlock()

			erm.logger.WithFields(logging.Fields{
				"component": errorCtx.Component,
				"operation": errorCtx.Operation,
				"strategy":  strategy.GetStrategyName(),
			}).Info("Error recovery successful")

			hasSuccessfulRecovery = true
			// Continue to try other strategies for comprehensive recovery
		} else {
			// Recovery failed - track failure
			erm.errorMetrics.mutex.Lock()
			erm.errorMetrics.RecoveryFailures++
			erm.errorMetrics.mutex.Unlock()

			erm.logger.WithFields(logging.Fields{
				"component": errorCtx.Component,
				"operation": errorCtx.Operation,
				"strategy":  strategy.GetStrategyName(),
				"error":     recoveryErr.Error(),
			}).Warn("Recovery strategy failed")

			lastErr = recoveryErr
		}
	}

	// Return result based on whether any recovery was successful
	if hasSuccessfulRecovery {
		erm.logger.WithFields(logging.Fields{
			"component":        errorCtx.Component,
			"operation":        errorCtx.Operation,
			"strategies_tried": len(strategies),
		}).Info("Error recovery completed with at least one successful strategy")
		return nil
	}

	erm.logger.WithFields(logging.Fields{
		"component":        errorCtx.Component,
		"operation":        errorCtx.Operation,
		"strategies_tried": len(strategies),
	}).Error("All recovery strategies failed")

	return lastErr
}

// GetMetrics returns current error metrics
func (erm *ErrorRecoveryManager) GetMetrics() *ErrorMetrics {
	erm.errorMetrics.mutex.RLock()
	defer erm.errorMetrics.mutex.RUnlock()

	// Return a copy to avoid race conditions
	return &ErrorMetrics{
		TotalErrors:       erm.errorMetrics.TotalErrors,
		ErrorsByComponent: copyStringInt64Map(erm.errorMetrics.ErrorsByComponent),
		ErrorsBySeverity:  copyStringInt64Map(erm.errorMetrics.ErrorsBySeverity),
		RecoveryAttempts:  erm.errorMetrics.RecoveryAttempts,
		RecoverySuccesses: erm.errorMetrics.RecoverySuccesses,
		RecoveryFailures:  erm.errorMetrics.RecoveryFailures,
		LastErrorTime:     erm.errorMetrics.LastErrorTime,
		LastRecoveryTime:  erm.errorMetrics.LastRecoveryTime,
	}
}

// copyStringInt64Map creates a copy of a map[string]int64
func copyStringInt64Map(src map[string]int64) map[string]int64 {
	dst := make(map[string]int64)
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
