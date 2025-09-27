/*
MediaMTX Error Metrics and Alerting Implementation

This package provides comprehensive error metrics tracking and alerting
capabilities for the MediaMTX camera service.

Requirements Coverage:
- REQ-MTX-007: Error handling and recovery
- REQ-MTX-008: Logging and monitoring

Test Categories: Unit/Integration
*/

package mediamtx

import (
	"sync"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
)

// ErrorAlertThresholds defines thresholds for error alerting
type ErrorAlertThresholds struct {
	PanicThreshold              int64         `json:"panic_threshold"`
	RecordingFailureThreshold   int64         `json:"recording_failure_threshold"`
	FFmpegFailureThreshold      int64         `json:"ffmpeg_failure_threshold"`
	CircuitBreakerOpenThreshold int64         `json:"circuit_breaker_open_threshold"`
	RecoveryFailureThreshold    int64         `json:"recovery_failure_threshold"`
	TimeWindow                  time.Duration `json:"time_window"`
}

// ErrorAlerter manages error alerting based on thresholds
type ErrorAlerter struct {
	logger     *logging.Logger
	thresholds ErrorAlertThresholds
	metrics    *ErrorMetrics
	alerts     map[string]time.Time // Track when alerts were last sent
	mutex      sync.RWMutex
}

// NewErrorAlerter creates a new error alerter
func NewErrorAlerter(logger *logging.Logger, thresholds ErrorAlertThresholds) *ErrorAlerter {
	return &ErrorAlerter{
		logger:     logger,
		thresholds: thresholds,
		alerts:     make(map[string]time.Time),
	}
}

// SetMetrics sets the metrics source for the alerter
func (ea *ErrorAlerter) SetMetrics(metrics *ErrorMetrics) {
	ea.metrics = metrics
}

// CheckThresholds checks if any error thresholds have been exceeded
func (ea *ErrorAlerter) CheckThresholds() {
	if ea.metrics == nil {
		return
	}

	ea.metrics.mutex.RLock()
	defer ea.metrics.mutex.RUnlock()

	now := time.Now()

	// Check panic threshold
	if ea.metrics.TotalErrors >= ea.thresholds.PanicThreshold {
		ea.checkAlert("panic_threshold", "Panic threshold exceeded - immediate attention required", now)
	}

	// Check recording failure threshold
	if ea.metrics.ErrorsByComponent["RecordingManager"] >= ea.thresholds.RecordingFailureThreshold {
		ea.checkAlert("recording_failure", "Recording failure threshold exceeded - system degraded", now)
	}

	// Check FFmpeg failure threshold
	if ea.metrics.ErrorsBySeverity["error"] >= ea.thresholds.FFmpegFailureThreshold {
		ea.checkAlert("ffmpeg_failure", "FFmpeg failure threshold exceeded - video processing issues", now)
	}

	// Check recovery failure threshold
	if ea.metrics.RecoveryFailures >= ea.thresholds.RecoveryFailureThreshold {
		ea.checkAlert("recovery_failure", "Error recovery failure threshold exceeded - system resilience compromised", now)
	}
}

// checkAlert checks if an alert should be sent (with rate limiting)
func (ea *ErrorAlerter) checkAlert(alertType, message string, now time.Time) {
	ea.mutex.Lock()
	defer ea.mutex.Unlock()

	lastAlert, exists := ea.alerts[alertType]
	if !exists || now.Sub(lastAlert) > ea.thresholds.TimeWindow {
		// Send alert
		ea.logger.WithFields(logging.Fields{
			"alert_type":    alertType,
			"message":       message,
			"threshold":     ea.getThresholdValue(alertType),
			"current_value": ea.getCurrentValue(alertType),
		}).Error("ERROR ALERT: " + message)

		ea.alerts[alertType] = now
	}
}

// getThresholdValue returns the threshold value for an alert type
func (ea *ErrorAlerter) getThresholdValue(alertType string) int64 {
	switch alertType {
	case "panic_threshold":
		return ea.thresholds.PanicThreshold
	case "recording_failure":
		return ea.thresholds.RecordingFailureThreshold
	case "ffmpeg_failure":
		return ea.thresholds.FFmpegFailureThreshold
	case "recovery_failure":
		return ea.thresholds.RecoveryFailureThreshold
	default:
		return 0
	}
}

// getCurrentValue returns the current value for an alert type
func (ea *ErrorAlerter) getCurrentValue(alertType string) int64 {
	if ea.metrics == nil {
		return 0
	}

	switch alertType {
	case "panic_threshold":
		return ea.metrics.TotalErrors
	case "recording_failure":
		return ea.metrics.ErrorsByComponent["RecordingManager"]
	case "ffmpeg_failure":
		return ea.metrics.ErrorsBySeverity["error"]
	case "recovery_failure":
		return ea.metrics.RecoveryFailures
	default:
		return 0
	}
}

// GetAlertStatus returns the current alert status
func (ea *ErrorAlerter) GetAlertStatus() map[string]interface{} {
	ea.mutex.RLock()
	defer ea.mutex.RUnlock()

	status := make(map[string]interface{})
	for alertType, lastAlert := range ea.alerts {
		status[alertType] = map[string]interface{}{
			"last_alert": lastAlert,
			"time_since": time.Since(lastAlert),
			"active":     time.Since(lastAlert) < ea.thresholds.TimeWindow,
		}
	}

	return status
}

// ErrorMetricsCollector collects and aggregates error metrics
type ErrorMetricsCollector struct {
	logger    *logging.Logger
	metrics   *ErrorMetrics
	alerter   *ErrorAlerter
	startTime time.Time
	mutex     sync.RWMutex
	stopChan  chan struct{} // Channel to signal stop
}

// NewErrorMetricsCollector creates a new error metrics collector
func NewErrorMetricsCollector(logger *logging.Logger) *ErrorMetricsCollector {
	thresholds := ErrorAlertThresholds{
		PanicThreshold:              5,               // Alert after 5 panics
		RecordingFailureThreshold:   10,              // Alert after 10 recording failures
		FFmpegFailureThreshold:      15,              // Alert after 15 FFmpeg failures
		CircuitBreakerOpenThreshold: 3,               // Alert after 3 circuit breaker opens
		RecoveryFailureThreshold:    5,               // Alert after 5 recovery failures
		TimeWindow:                  5 * time.Minute, // Rate limit alerts to 5 minutes
	}

	alerter := NewErrorAlerter(logger, thresholds)

	return &ErrorMetricsCollector{
		logger: logger,
		metrics: &ErrorMetrics{
			ErrorsByComponent: make(map[string]int64),
			ErrorsBySeverity:  make(map[string]int64),
		},
		alerter:   alerter,
		startTime: time.Now(),
		stopChan:  make(chan struct{}),
	}
}

// Initialize sets up the metrics collector
func (emc *ErrorMetricsCollector) Initialize() {
	emc.alerter.SetMetrics(emc.metrics)

	// Start periodic threshold checking
	go emc.startPeriodicChecking()
}

// startPeriodicChecking runs periodic threshold checks
func (emc *ErrorMetricsCollector) startPeriodicChecking() {
	ticker := time.NewTicker(30 * time.Second) // Check every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			emc.alerter.CheckThresholds()
		case <-emc.stopChan:
			emc.logger.Debug("ErrorMetricsCollector periodic checking stopped")
			return
		}
	}
}

// Stop stops the error metrics collector
func (emc *ErrorMetricsCollector) Stop() {
	emc.mutex.Lock()
	defer emc.mutex.Unlock()

	select {
	case <-emc.stopChan:
		// Channel already closed, nothing to do
	default:
		close(emc.stopChan)
	}
}

// RecordError records an error in the metrics
func (emc *ErrorMetricsCollector) RecordError(component string, severity string, recoverable bool) {
	emc.mutex.Lock()
	defer emc.mutex.Unlock()

	emc.metrics.mutex.Lock()
	emc.metrics.TotalErrors++
	emc.metrics.ErrorsByComponent[component]++
	emc.metrics.ErrorsBySeverity[severity]++
	emc.metrics.LastErrorTime = time.Now()
	emc.metrics.mutex.Unlock()

	emc.logger.WithFields(logging.Fields{
		"component":    component,
		"severity":     severity,
		"recoverable":  recoverable,
		"total_errors": emc.metrics.TotalErrors,
	}).Debug("Error recorded in metrics")
}

// RecordRecoveryAttempt records a recovery attempt
func (emc *ErrorMetricsCollector) RecordRecoveryAttempt(success bool) {
	emc.mutex.Lock()
	defer emc.mutex.Unlock()

	emc.metrics.mutex.Lock()
	emc.metrics.RecoveryAttempts++
	if success {
		emc.metrics.RecoverySuccesses++
		emc.metrics.LastRecoveryTime = time.Now()
	} else {
		emc.metrics.RecoveryFailures++
	}
	emc.metrics.mutex.Unlock()

	emc.logger.WithFields(logging.Fields{
		"success":               success,
		"total_attempts":        emc.metrics.RecoveryAttempts,
		"successful_recoveries": emc.metrics.RecoverySuccesses,
		"failed_recoveries":     emc.metrics.RecoveryFailures,
	}).Debug("Recovery attempt recorded in metrics")
}

// GetMetrics returns a copy of current metrics
func (emc *ErrorMetricsCollector) GetMetrics() *ErrorMetrics {
	emc.metrics.mutex.RLock()
	defer emc.metrics.mutex.RUnlock()

	return &ErrorMetrics{
		TotalErrors:       emc.metrics.TotalErrors,
		ErrorsByComponent: copyStringInt64Map(emc.metrics.ErrorsByComponent),
		ErrorsBySeverity:  copyStringInt64Map(emc.metrics.ErrorsBySeverity),
		RecoveryAttempts:  emc.metrics.RecoveryAttempts,
		RecoverySuccesses: emc.metrics.RecoverySuccesses,
		RecoveryFailures:  emc.metrics.RecoveryFailures,
		LastErrorTime:     emc.metrics.LastErrorTime,
		LastRecoveryTime:  emc.metrics.LastRecoveryTime,
	}
}

// GetUptime returns the uptime since metrics collection started
func (emc *ErrorMetricsCollector) GetUptime() time.Duration {
	return time.Since(emc.startTime)
}

// GetAlertStatus returns the current alert status
func (emc *ErrorMetricsCollector) GetAlertStatus() map[string]interface{} {
	return emc.alerter.GetAlertStatus()
}

// Reset resets all metrics (useful for testing)
func (emc *ErrorMetricsCollector) Reset() {
	emc.mutex.Lock()
	defer emc.mutex.Unlock()

	emc.metrics.mutex.Lock()
	emc.metrics.TotalErrors = 0
	emc.metrics.ErrorsByComponent = make(map[string]int64)
	emc.metrics.ErrorsBySeverity = make(map[string]int64)
	emc.metrics.RecoveryAttempts = 0
	emc.metrics.RecoverySuccesses = 0
	emc.metrics.RecoveryFailures = 0
	emc.metrics.LastErrorTime = time.Time{}
	emc.metrics.LastRecoveryTime = time.Time{}
	emc.metrics.mutex.Unlock()

	emc.startTime = time.Now()

	emc.logger.Info("Error metrics reset")
}
