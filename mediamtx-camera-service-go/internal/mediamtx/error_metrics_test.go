/*
MediaMTX Error Metrics and Alerting Tests

Requirements Coverage:
- REQ-MTX-007: Error handling and recovery
- REQ-MTX-008: Logging and monitoring

Test Categories: Unit
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package mediamtx

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewErrorMetricsCollector_ReqMTX008 tests error metrics collector creation
func TestNewErrorMetricsCollector_ReqMTX008(t *testing.T) {
	// REQ-MTX-008: Logging and monitoring
	helper, _ := SetupMediaMTXTest(t)

	logger := helper.GetLogger()
	collector := NewErrorMetricsCollector(logger)

	require.NotNil(t, collector)
	assert.NotNil(t, collector.metrics)
	assert.NotNil(t, collector.alerter)
	assert.False(t, collector.startTime.IsZero())
}

// TestErrorMetricsCollector_RecordError_ReqMTX008 tests error recording
func TestErrorMetricsCollector_RecordError_ReqMTX008(t *testing.T) {
	// REQ-MTX-008: Logging and monitoring
	helper, _ := SetupMediaMTXTest(t)

	logger := helper.GetLogger()
	collector := NewErrorMetricsCollector(logger)

	// Record some errors
	collector.RecordError("TestComponent", "error", true)
	collector.RecordError("TestComponent", "warning", false)
	collector.RecordError("AnotherComponent", "critical", true)

	metrics := collector.GetMetrics()

	assert.Equal(t, int64(3), metrics.TotalErrors)
	assert.Equal(t, int64(2), metrics.ErrorsByComponent["TestComponent"])
	assert.Equal(t, int64(1), metrics.ErrorsByComponent["AnotherComponent"])
	assert.Equal(t, int64(1), metrics.ErrorsBySeverity["error"])
	assert.Equal(t, int64(1), metrics.ErrorsBySeverity["warning"])
	assert.Equal(t, int64(1), metrics.ErrorsBySeverity["critical"])
	assert.False(t, metrics.LastErrorTime.IsZero())
}

// TestErrorMetricsCollector_RecordRecoveryAttempt_ReqMTX008 tests recovery attempt recording
func TestErrorMetricsCollector_RecordRecoveryAttempt_ReqMTX008(t *testing.T) {
	// REQ-MTX-008: Logging and monitoring
	helper, _ := SetupMediaMTXTest(t)

	logger := helper.GetLogger()
	collector := NewErrorMetricsCollector(logger)

	// Record recovery attempts
	collector.RecordRecoveryAttempt(true)
	collector.RecordRecoveryAttempt(false)
	collector.RecordRecoveryAttempt(true)

	metrics := collector.GetMetrics()

	assert.Equal(t, int64(3), metrics.RecoveryAttempts)
	assert.Equal(t, int64(2), metrics.RecoverySuccesses)
	assert.Equal(t, int64(1), metrics.RecoveryFailures)
	assert.False(t, metrics.LastRecoveryTime.IsZero())
}

// TestErrorMetricsCollector_GetMetrics_ReqMTX008 tests metrics retrieval
func TestErrorMetricsCollector_GetMetrics_ReqMTX008(t *testing.T) {
	// REQ-MTX-008: Logging and monitoring
	helper, _ := SetupMediaMTXTest(t)

	logger := helper.GetLogger()
	collector := NewErrorMetricsCollector(logger)

	// Record some data
	collector.RecordError("TestComponent", "error", true)
	collector.RecordRecoveryAttempt(true)

	metrics := collector.GetMetrics()

	assert.Equal(t, int64(1), metrics.TotalErrors)
	assert.Equal(t, int64(1), metrics.ErrorsByComponent["TestComponent"])
	assert.Equal(t, int64(1), metrics.ErrorsBySeverity["error"])
	assert.Equal(t, int64(1), metrics.RecoveryAttempts)
	assert.Equal(t, int64(1), metrics.RecoverySuccesses)
	assert.Equal(t, int64(0), metrics.RecoveryFailures)
}

// TestErrorMetricsCollector_GetUptime_ReqMTX008 tests uptime calculation
func TestErrorMetricsCollector_GetUptime_ReqMTX008(t *testing.T) {
	// REQ-MTX-008: Logging and monitoring
	helper, _ := SetupMediaMTXTest(t)

	logger := helper.GetLogger()
	collector := NewErrorMetricsCollector(logger)

	// Wait a bit
	time.Sleep(10 * time.Millisecond)

	uptime := collector.GetUptime()
	assert.True(t, uptime >= 10*time.Millisecond)
	assert.True(t, uptime < 100*time.Millisecond)
}

// TestErrorMetricsCollector_Reset_ReqMTX008 tests metrics reset
func TestErrorMetricsCollector_Reset_ReqMTX008(t *testing.T) {
	// REQ-MTX-008: Logging and monitoring
	helper, _ := SetupMediaMTXTest(t)

	logger := helper.GetLogger()
	collector := NewErrorMetricsCollector(logger)

	// Record some data
	collector.RecordError("TestComponent", "error", true)
	collector.RecordRecoveryAttempt(false)

	// Verify data exists
	metrics := collector.GetMetrics()
	assert.Equal(t, int64(1), metrics.TotalErrors)
	assert.Equal(t, int64(1), metrics.RecoveryAttempts)

	// Reset
	collector.Reset()

	// Verify data is cleared
	metrics = collector.GetMetrics()
	assert.Equal(t, int64(0), metrics.TotalErrors)
	assert.Equal(t, int64(0), metrics.ErrorsByComponent["TestComponent"])
	assert.Equal(t, int64(0), metrics.ErrorsBySeverity["error"])
	assert.Equal(t, int64(0), metrics.RecoveryAttempts)
	assert.Equal(t, int64(0), metrics.RecoverySuccesses)
	assert.Equal(t, int64(0), metrics.RecoveryFailures)
	assert.True(t, metrics.LastErrorTime.IsZero())
	assert.True(t, metrics.LastRecoveryTime.IsZero())
}

// TestNewErrorAlerter_ReqMTX008 tests error alerter creation
func TestNewErrorAlerter_ReqMTX008(t *testing.T) {
	// REQ-MTX-008: Logging and monitoring
	helper, _ := SetupMediaMTXTest(t)

	logger := helper.GetLogger()
	thresholds := ErrorAlertThresholds{
		PanicThreshold:              5,
		RecordingFailureThreshold:   10,
		FFmpegFailureThreshold:      15,
		CircuitBreakerOpenThreshold: 3,
		RecoveryFailureThreshold:    5,
		TimeWindow:                  5 * time.Minute,
	}

	alerter := NewErrorAlerter(logger, thresholds)

	require.NotNil(t, alerter)
	assert.Equal(t, thresholds, alerter.thresholds)
	assert.NotNil(t, alerter.alerts)
}

// TestErrorAlerter_CheckThresholds_ReqMTX008 tests threshold checking
func TestErrorAlerter_CheckThresholds_ReqMTX008(t *testing.T) {
	// REQ-MTX-008: Logging and monitoring
	helper, _ := SetupMediaMTXTest(t)

	logger := helper.GetLogger()
	thresholds := ErrorAlertThresholds{
		PanicThreshold:            2, // Low threshold for testing
		RecordingFailureThreshold: 2,
		FFmpegFailureThreshold:    2,
		RecoveryFailureThreshold:  2,
		TimeWindow:                1 * time.Second,
	}

	alerter := NewErrorAlerter(logger, thresholds)

	// Create mock metrics that exceed thresholds
	metrics := &ErrorMetrics{
		TotalErrors:       3,
		ErrorsByComponent: map[string]int64{"RecordingManager": 3},
		ErrorsBySeverity:  map[string]int64{"error": 3},
		RecoveryFailures:  3,
	}

	alerter.SetMetrics(metrics)

	// Check thresholds - should trigger alerts
	alerter.CheckThresholds()

	// Verify alerts were triggered
	status := alerter.GetAlertStatus()
	assert.NotEmpty(t, status)
}

// TestErrorAlerter_GetAlertStatus_ReqMTX008 tests alert status retrieval
func TestErrorAlerter_GetAlertStatus_ReqMTX008(t *testing.T) {
	// REQ-MTX-008: Logging and monitoring
	helper, _ := SetupMediaMTXTest(t)

	logger := helper.GetLogger()
	thresholds := ErrorAlertThresholds{
		PanicThreshold: 2,
		TimeWindow:     1 * time.Second,
	}

	alerter := NewErrorAlerter(logger, thresholds)

	// Create mock metrics
	metrics := &ErrorMetrics{
		TotalErrors: 3,
	}
	alerter.SetMetrics(metrics)

	// Trigger alert
	alerter.CheckThresholds()

	status := alerter.GetAlertStatus()

	// Should have panic_threshold alert
	assert.Contains(t, status, "panic_threshold")

	alertInfo := status["panic_threshold"].(map[string]interface{})
	assert.Contains(t, alertInfo, "last_alert")
	assert.Contains(t, alertInfo, "time_since")
	assert.Contains(t, alertInfo, "active")
	assert.True(t, alertInfo["active"].(bool))
}

// TestErrorAlerter_RateLimit_ReqMTX008 tests alert rate limiting
func TestErrorAlerter_RateLimit_ReqMTX008(t *testing.T) {
	// REQ-MTX-008: Logging and monitoring
	helper, _ := SetupMediaMTXTest(t)

	logger := helper.GetLogger()
	thresholds := ErrorAlertThresholds{
		PanicThreshold: 1,                      // Very low threshold
		TimeWindow:     100 * time.Millisecond, // Short window
	}

	alerter := NewErrorAlerter(logger, thresholds)

	metrics := &ErrorMetrics{
		TotalErrors: 2,
	}
	alerter.SetMetrics(metrics)

	// First check - should trigger alert
	alerter.CheckThresholds()

	// Immediate second check - should be rate limited
	alerter.CheckThresholds()

	status := alerter.GetAlertStatus()
	assert.NotEmpty(t, status)

	// Wait for rate limit window to expire
	time.Sleep(150 * time.Millisecond)

	// Third check - should trigger alert again
	alerter.CheckThresholds()
}

// TestErrorAlertThresholds_ReqMTX008 tests alert threshold structure
func TestErrorAlertThresholds_ReqMTX008(t *testing.T) {
	// REQ-MTX-008: Logging and monitoring
	thresholds := ErrorAlertThresholds{
		PanicThreshold:              5,
		RecordingFailureThreshold:   10,
		FFmpegFailureThreshold:      15,
		CircuitBreakerOpenThreshold: 3,
		RecoveryFailureThreshold:    5,
		TimeWindow:                  5 * time.Minute,
	}

	assert.Equal(t, int64(5), thresholds.PanicThreshold)
	assert.Equal(t, int64(10), thresholds.RecordingFailureThreshold)
	assert.Equal(t, int64(15), thresholds.FFmpegFailureThreshold)
	assert.Equal(t, int64(3), thresholds.CircuitBreakerOpenThreshold)
	assert.Equal(t, int64(5), thresholds.RecoveryFailureThreshold)
	assert.Equal(t, 5*time.Minute, thresholds.TimeWindow)
}

// TestErrorMetrics_ReqMTX008 tests error metrics structure
func TestErrorMetrics_ReqMTX008(t *testing.T) {
	// REQ-MTX-008: Logging and monitoring
	metrics := &ErrorMetrics{
		TotalErrors:       10,
		ErrorsByComponent: map[string]int64{"TestComponent": 5},
		ErrorsBySeverity:  map[string]int64{"error": 8, "warning": 2},
		RecoveryAttempts:  5,
		RecoverySuccesses: 3,
		RecoveryFailures:  2,
		LastErrorTime:     time.Now(),
		LastRecoveryTime:  time.Now(),
	}

	assert.Equal(t, int64(10), metrics.TotalErrors)
	assert.Equal(t, int64(5), metrics.ErrorsByComponent["TestComponent"])
	assert.Equal(t, int64(8), metrics.ErrorsBySeverity["error"])
	assert.Equal(t, int64(2), metrics.ErrorsBySeverity["warning"])
	assert.Equal(t, int64(5), metrics.RecoveryAttempts)
	assert.Equal(t, int64(3), metrics.RecoverySuccesses)
	assert.Equal(t, int64(2), metrics.RecoveryFailures)
	assert.False(t, metrics.LastErrorTime.IsZero())
	assert.False(t, metrics.LastRecoveryTime.IsZero())
}
