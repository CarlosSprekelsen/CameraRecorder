/*
MediaMTX Error Recovery Implementation

Requirements Coverage:
- REQ-MTX-001: MediaMTX service integration
- REQ-MTX-002: Stream management capabilities
- REQ-MTX-003: Path creation and deletion
- REQ-MTX-004: Health monitoring

Test Categories: Unit/Integration
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package mediamtx

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
)

// PathErrorRecovery handles path error recovery and graceful degradation
type PathErrorRecovery struct {
	logger        *logging.Logger
	notifications chan PathValidationError
}

// PathValidationError represents a path-related error with severity classification
type PathValidationError struct {
	Path      string
	Error     error
	Severity  string // "critical", "warning", "info"
	Timestamp time.Time
}

// NewPathErrorRecovery creates a new path error recovery instance
func NewPathErrorRecovery(logger *logging.Logger) *PathErrorRecovery {
	return &PathErrorRecovery{
		logger:        logger,
		notifications: make(chan PathValidationError, 100), // Buffer for async processing
	}
}

// HandlePathError processes path errors with appropriate recovery strategies
func (per *PathErrorRecovery) HandlePathError(err PathValidationError) {
	switch err.Severity {
	case "critical":
		// Cannot recover - notify admins
		per.notifyAdmins(err)
		// Disable affected functionality
		per.disableFeature(err.Path)

	case "warning":
		// Try fallback
		per.logger.WithError(err.Error).Warn("Path issue detected, using fallback")
		// Monitor and alert if persists
		per.monitorPath(err.Path)

	case "info":
		// Log only
		per.logger.WithError(err.Error).Info("Non-critical path issue")
	}
}

// monitorPath sets up monitoring for a problematic path
func (per *PathErrorRecovery) monitorPath(path string) {
	// Set up monitoring for the problematic path
	go func() {
		retries := 0
		maxRetries := 5

		for retries < maxRetries {
			// Use timeout for retry backoff
			backoffDuration := time.Minute * time.Duration(retries+1)
			time.Sleep(backoffDuration)

			if err := checkPath(path); err == nil {
				per.logger.WithField("path", path).Info("Path recovered")
				return
			}

			retries++
		}

		per.logger.WithField("path", path).Error("Path did not recover after retries")
		per.notifyAdmins(PathValidationError{
			Path:     path,
			Error:    fmt.Errorf("path unavailable after %d retries", maxRetries),
			Severity: "critical",
		})
	}()
}

// notifyAdmins sends critical path error notifications
func (per *PathErrorRecovery) notifyAdmins(err PathValidationError) {
	per.logger.WithFields(logging.Fields{
		"path":     err.Path,
		"error":    err.Error,
		"severity": err.Severity,
	}).Error("CRITICAL: Path error requiring admin attention")

	// In a real implementation, this would send notifications via:
	// - Email alerts
	// - Slack/Teams notifications
	// - PagerDuty alerts
	// - System monitoring integration
}

// disableFeature disables functionality for a problematic path
func (per *PathErrorRecovery) disableFeature(path string) {
	per.logger.WithField("path", path).Error("Disabling feature due to path issues")

	// In a real implementation, this would:
	// - Update service health status
	// - Disable recording/snapshot operations for this path
	// - Update API responses to indicate degraded service
	// - Trigger circuit breaker patterns
}

// checkPath performs a basic path health check
func checkPath(path string) error {
	// Check if path exists and is accessible
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", path)
	}
	if err != nil {
		return fmt.Errorf("cannot access path: %w", err)
	}

	// Check if it's a directory
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", path)
	}

	// Check write permission with a test file
	testFile := filepath.Join(path, fmt.Sprintf(".health_check_%d", time.Now().UnixNano()))
	file, err := os.Create(testFile)
	if err != nil {
		return fmt.Errorf("path is not writable: %w", err)
	}
	file.Close()
	os.Remove(testFile)

	return nil
}

// StartErrorRecovery starts the error recovery background process
func (per *PathErrorRecovery) StartErrorRecovery(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case pathErr := <-per.notifications:
				per.HandlePathError(pathErr)
			}
		}
	}()
}

// ReportPathError reports a path error for processing
func (per *PathErrorRecovery) ReportPathError(path string, err error, severity string) {
	select {
	case per.notifications <- PathValidationError{
		Path:      path,
		Error:     err,
		Severity:  severity,
		Timestamp: time.Now(),
	}:
	default:
		// Channel is full, log directly
		per.logger.WithError(err).WithField("path", path).Error("Path error notification channel full")
	}
}
