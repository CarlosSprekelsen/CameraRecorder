/*
Centralized Health Notification Manager

Provides unified debounce mechanism for all health notifications across the system.
Moves threshold logic from WebSocket layer to controller layer for proper architecture.

Features:
- Configurable debounce durations per component
- Atomic operations for thread safety
- Unified notification patterns
- Configurable thresholds from performance config
*/

package mediamtx

import (
	"sync/atomic"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
)

// HealthNotificationManager manages debounced health notifications
type HealthNotificationManager struct {
	// Configuration
	config *config.Config
	logger *logging.Logger

	// Debounce state per component (all atomic for thread safety)
	lastNotificationTimes    map[string]*int64 // Component -> timestamp (UnixNano) pointer for atomic ops
	lastNotificationStatuses map[string]*int32 // Component -> status (0=normal, 1=warning, 2=critical) pointer for atomic ops

	// System event notifier
	systemNotifier SystemEventNotifier
}

// NewHealthNotificationManager creates a new health notification manager
func NewHealthNotificationManager(cfg *config.Config, logger *logging.Logger, notifier SystemEventNotifier) *HealthNotificationManager {
	return &HealthNotificationManager{
		config: cfg,
		logger: logger,
		lastNotificationTimes: map[string]*int64{
			"storage_monitor":     new(int64),
			"performance_monitor": new(int64),
			"health_monitor":      new(int64),
		},
		lastNotificationStatuses: map[string]*int32{
			"storage_monitor":     new(int32),
			"performance_monitor": new(int32),
			"health_monitor":      new(int32),
		},
		systemNotifier: notifier,
	}
}

// ShouldNotifyStorage checks if a storage notification should be sent with debounce
func (h *HealthNotificationManager) ShouldNotifyStorage(status string, usagePercent, threshold float64, storageInfo interface {
	GetAvailableSpace() int64
	GetTotalSpace() int64
}) bool {
	component := "storage_monitor"
	debounceDuration := time.Duration(h.config.Performance.Debounce.StorageMonitorSeconds) * time.Second

	if !h.shouldNotifyWithDebounce(component, status, debounceDuration) {
		return false
	}

	// Send notification
	h.sendStorageNotification(status, usagePercent, threshold, storageInfo)
	return true
}

// ShouldNotifyPerformance checks if a performance notification should be sent with debounce
func (h *HealthNotificationManager) ShouldNotifyPerformance(status, metricName string, value, threshold float64, severity string) bool {
	component := "performance_monitor"
	debounceDuration := time.Duration(h.config.Performance.Debounce.PerformanceMonitorSeconds) * time.Second

	if !h.shouldNotifyWithDebounce(component, status, debounceDuration) {
		return false
	}

	// Send notification
	h.sendPerformanceNotification(status, metricName, value, threshold, severity)
	return true
}

// ShouldNotifyHealth checks if a health notification should be sent with debounce
func (h *HealthNotificationManager) ShouldNotifyHealth(status string, metrics map[string]interface{}) bool {
	component := "health_monitor"
	debounceDuration := time.Duration(h.config.Performance.Debounce.HealthMonitorSeconds) * time.Second

	if !h.shouldNotifyWithDebounce(component, status, debounceDuration) {
		return false
	}

	// Send notification
	if h.systemNotifier != nil {
		h.systemNotifier.NotifySystemHealth(status, metrics)
	}
	return true
}

// shouldNotifyWithDebounce checks if a notification should be sent based on debounce logic
func (h *HealthNotificationManager) shouldNotifyWithDebounce(component, status string, debounceDuration time.Duration) bool {
	now := time.Now().UnixNano()
	lastTime := atomic.LoadInt64(h.lastNotificationTimes[component])

	// Check if enough time has passed since last notification
	if now-lastTime < int64(debounceDuration) {
		h.logger.WithFields(logging.Fields{
			"component":         component,
			"status":            status,
			"time_since_last":   time.Duration(now - lastTime).String(),
			"debounce_duration": debounceDuration.String(),
		}).Debug("Notification suppressed due to debounce period")
		return false
	}

	// Convert status string to atomic int32
	var statusValue int32
	switch status {
	case "normal", "healthy":
		statusValue = 0
	case "warning", "storage_warning", "high_error_rate", "slow_response_time", "connection_limit_warning", "goroutine_leak_warning":
		statusValue = 1
	case "critical", "storage_critical", "memory_pressure", "unhealthy":
		statusValue = 2
	default:
		statusValue = 0 // default to normal
	}

	// Check if status actually changed using atomic compare-and-swap
	lastStatus := atomic.LoadInt32(h.lastNotificationStatuses[component])
	if lastStatus == statusValue {
		h.logger.WithFields(logging.Fields{
			"component": component,
			"status":    status,
		}).Debug("Notification suppressed - no state change")
		return false
	}

	// Update state atomically - use compare-and-swap to ensure atomicity
	if atomic.CompareAndSwapInt32(h.lastNotificationStatuses[component], lastStatus, statusValue) {
		atomic.StoreInt64(h.lastNotificationTimes[component], now)

		h.logger.WithFields(logging.Fields{
			"component":       component,
			"status":          status,
			"previous_status": lastStatus,
		}).Info("Health notification approved - state change detected")

		return true
	}

	// If CAS failed, another goroutine updated it, check again
	return atomic.LoadInt32(h.lastNotificationStatuses[component]) != statusValue
}

// sendStorageNotification sends storage threshold-crossing notifications
func (h *HealthNotificationManager) sendStorageNotification(status string, usagePercent, threshold float64, storageInfo interface {
	GetAvailableSpace() int64
	GetTotalSpace() int64
}) {
	if h.systemNotifier == nil {
		return
	}

	// Determine severity
	severity := "warning"
	if status == "storage_critical" {
		severity = "critical"
	}

	// Build notification payload
	notificationData := map[string]interface{}{
		"usage_percentage": usagePercent,
		"threshold":        threshold,
		"available_space":  storageInfo.GetAvailableSpace(),
		"total_space":      storageInfo.GetTotalSpace(),
		"component":        "storage_monitor",
		"severity":         severity,
		"timestamp":        time.Now().Format(time.RFC3339),
		"reason":           "storage_threshold_exceeded",
	}

	// Send system health notification
	h.systemNotifier.NotifySystemHealth(status, notificationData)

	h.logger.WithFields(logging.Fields{
		"status":           status,
		"usage_percentage": usagePercent,
		"threshold":        threshold,
		"severity":         severity,
	}).Warn("Storage threshold exceeded")
}

// sendPerformanceNotification sends performance threshold-crossing notifications
func (h *HealthNotificationManager) sendPerformanceNotification(status, metricName string, value, threshold float64, severity string) {
	if h.systemNotifier == nil {
		return
	}

	notificationData := map[string]interface{}{
		"metric":    metricName,
		"value":     value,
		"threshold": threshold,
		"component": "performance_monitor",
		"severity":  severity,
		"timestamp": time.Now().Format(time.RFC3339),
		"reason":    "performance_threshold_exceeded",
	}

	// Send system health notification
	h.systemNotifier.NotifySystemHealth(status, notificationData)

	h.logger.WithFields(logging.Fields{
		"status":    status,
		"metric":    metricName,
		"value":     value,
		"threshold": threshold,
		"severity":  severity,
	}).Warn("Performance threshold exceeded")
}

// CheckStorageThresholds checks storage usage against configurable thresholds
func (h *HealthNotificationManager) CheckStorageThresholds(storageInfo interface {
	GetUsagePercentage() float64
	GetAvailableSpace() int64
	GetTotalSpace() int64
	IsLowSpaceWarning() bool
}) {
	usagePercent := storageInfo.GetUsagePercentage()
	warnThreshold := float64(h.config.Storage.WarnPercent)
	blockThreshold := float64(h.config.Storage.BlockPercent)

	// Check critical threshold (block_percent)
	if usagePercent >= blockThreshold {
		h.ShouldNotifyStorage("storage_critical", usagePercent, blockThreshold, storageInfo)
	} else if usagePercent >= warnThreshold {
		// Check warning threshold (warn_percent)
		h.ShouldNotifyStorage("storage_warning", usagePercent, warnThreshold, storageInfo)
	}
}

// CheckPerformanceThresholds checks performance metrics against configurable thresholds
func (h *HealthNotificationManager) CheckPerformanceThresholds(metrics map[string]interface{}) {
	thresholds := h.config.Performance.MonitoringThresholds

	// Memory usage threshold
	if memUsage, ok := metrics["memory_usage"].(float64); ok && memUsage > thresholds.MemoryUsagePercent {
		h.ShouldNotifyPerformance("memory_pressure", "memory_usage", memUsage, thresholds.MemoryUsagePercent, "critical")
	}

	// Error rate threshold
	if errorRate, ok := metrics["error_rate"].(float64); ok && errorRate > thresholds.ErrorRatePercent {
		h.ShouldNotifyPerformance("high_error_rate", "error_rate", errorRate, thresholds.ErrorRatePercent, "warning")
	}

	// Average response time threshold
	if avgResponseTime, ok := metrics["average_response_time"].(float64); ok && avgResponseTime > thresholds.AverageResponseTimeMs {
		h.ShouldNotifyPerformance("slow_response_time", "average_response_time", avgResponseTime, thresholds.AverageResponseTimeMs, "warning")
	}

	// Active connections threshold
	if activeConn, ok := metrics["active_connections"].(int); ok && activeConn > thresholds.ActiveConnectionsLimit {
		h.ShouldNotifyPerformance("connection_limit_warning", "active_connections", float64(activeConn), float64(thresholds.ActiveConnectionsLimit), "warning")
	}

	// Goroutines threshold
	if goroutines, ok := metrics["goroutines"].(int); ok && goroutines > thresholds.GoroutinesLimit {
		h.ShouldNotifyPerformance("goroutine_leak_warning", "goroutines", float64(goroutines), float64(thresholds.GoroutinesLimit), "warning")
	}
}
