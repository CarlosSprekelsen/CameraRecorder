/*
MediaMTX Health Monitor Implementation - Simplified Version

Requirements Coverage:
- REQ-MTX-001: MediaMTX service integration
- REQ-MTX-002: Stream management capabilities
- REQ-MTX-003: Path creation and deletion
- REQ-MTX-004: Health monitoring

Test Categories: Unit/Integration
API Documentation Reference: docs/api/json_rpc_methods.md

Simplified for 20-user scale - removes over-engineering while maintaining functionality
*/

package mediamtx

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
)

// SimpleHealthMonitor represents the simplified MediaMTX health monitor
type SimpleHealthMonitor struct {
	client MediaMTXClient
	config *MediaMTXConfig
	logger *logging.Logger

	// Threshold-crossing notifications
	systemNotifier SystemEventNotifier

	// Debounce state for notifications
	lastNotificationTime   int64  // Atomic timestamp (UnixNano)
	lastNotificationStatus string // Last notification status sent
	debounceDuration       time.Duration

	// Atomic state: optimized for high-frequency reads
	isHealthy     int32 // 0 = false, 1 = true
	failureCount  int64 // Atomic counter
	lastCheckTime int64 // Atomic timestamp (UnixNano)

	// Keep mutex only for complex operations that need consistency
	mu sync.RWMutex

	// Control
	stopChan chan struct{}
	wg       sync.WaitGroup
}

// NewHealthMonitor creates a new simplified MediaMTX health monitor
func NewHealthMonitor(client MediaMTXClient, config *MediaMTXConfig, logger *logging.Logger) HealthMonitor {
	return &SimpleHealthMonitor{
		client:           client,
		config:           config,
		logger:           logger,
		isHealthy:        1, // Assume healthy initially (1 = true)
		failureCount:     0,
		lastCheckTime:    time.Now().UnixNano(),
		debounceDuration: 15 * time.Second, // 15s debounce for health notifications
		stopChan:         make(chan struct{}, 1),
	}
}

// SetSystemNotifier sets the system event notifier for threshold-crossing notifications
func (h *SimpleHealthMonitor) SetSystemNotifier(notifier SystemEventNotifier) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.systemNotifier = notifier
}

// shouldNotifyWithDebounce checks if a notification should be sent based on debounce logic
func (h *SimpleHealthMonitor) shouldNotifyWithDebounce(status string) bool {
	now := time.Now().UnixNano()
	lastTime := atomic.LoadInt64(&h.lastNotificationTime)

	// Check if enough time has passed since last notification
	if now-lastTime < int64(h.debounceDuration) {
		return false
	}

	// Check if status actually changed
	h.mu.RLock()
	lastStatus := h.lastNotificationStatus
	h.mu.RUnlock()

	if lastStatus == status {
		return false
	}

	// Update state atomically
	atomic.StoreInt64(&h.lastNotificationTime, now)
	h.mu.Lock()
	h.lastNotificationStatus = status
	h.mu.Unlock()

	return true
}

// Start starts the health monitoring
func (h *SimpleHealthMonitor) Start(ctx context.Context) error {
	h.logger.Info("Starting simplified MediaMTX health monitor")

	h.wg.Add(1)
	go h.monitorLoop(ctx)
	return nil
}

// Stop stops the health monitoring
func (h *SimpleHealthMonitor) Stop(ctx context.Context) error {
	h.logger.Info("Stopping simplified MediaMTX health monitor")

	// Signal stop - use atomic check to prevent double-close
	select {
	case h.stopChan <- struct{}{}:
		// Successfully sent stop signal
	default:
		// Channel might be full or already closed, but that's okay
	}

	// Wait for goroutine to finish
	h.wg.Wait()

	// Close the channel safely
	select {
	case <-h.stopChan:
		// Channel was still open, close it
		close(h.stopChan)
	default:
		// Channel was already closed, do nothing
	}

	h.logger.Info("Simplified MediaMTX health monitor stopped")
	return nil
}

// monitorLoop runs the health monitoring loop
func (h *SimpleHealthMonitor) monitorLoop(ctx context.Context) {
	defer h.wg.Done()

	// Use configured interval from centralized config
	checkInterval := time.Duration(h.config.HealthCheckInterval) * time.Second

	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-h.stopChan:
			return
		case <-ticker.C:
			h.checkHealth(ctx)
		}
	}
}

// checkHealth performs a health check
func (h *SimpleHealthMonitor) checkHealth(ctx context.Context) {
	// Use configured timeout from centralized config
	timeout := h.config.HealthCheckTimeout

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	err := h.client.HealthCheck(ctx)

	// Update timestamp atomically
	atomic.StoreInt64(&h.lastCheckTime, time.Now().UnixNano())

	if err != nil {
		// Increment failure count atomically
		failures := atomic.AddInt64(&h.failureCount, 1)
		h.logger.WithError(err).Debug("Health check failed")

		// Simple threshold: 3 failures = unhealthy (or use configured threshold)
		threshold := int64(3)
		if h.config.HealthFailureThreshold > 0 {
			threshold = int64(h.config.HealthFailureThreshold)
		}

		if failures >= threshold {
			// Use atomic compare-and-swap to set unhealthy
			if atomic.CompareAndSwapInt32(&h.isHealthy, 1, 0) {
				h.logger.Warn("MediaMTX service marked as unhealthy")
			}
		}
	} else {
		// Success - reset everything atomically
		currentHealthy := atomic.LoadInt32(&h.isHealthy)
		if currentHealthy == 0 {
			h.logger.Info("MediaMTX service recovered")
		}
		atomic.StoreInt32(&h.isHealthy, 1)
		atomic.StoreInt64(&h.failureCount, 0)
	}
}

// IsHealthy returns true if the service is healthy
func (h *SimpleHealthMonitor) IsHealthy() bool {
	return atomic.LoadInt32(&h.isHealthy) == 1
}

// IsCircuitOpen returns true if the circuit breaker is open (unhealthy)
func (h *SimpleHealthMonitor) IsCircuitOpen() bool {
	return !h.IsHealthy()
}

// GetStatus returns the current health status
func (h *SimpleHealthMonitor) GetStatus() HealthStatus {
	// Read all atomic values
	isHealthy := atomic.LoadInt32(&h.isHealthy) == 1
	failureCount := atomic.LoadInt64(&h.failureCount)
	lastCheckNano := atomic.LoadInt64(&h.lastCheckTime)
	lastCheckTime := time.Unix(0, lastCheckNano)

	status := "healthy"
	if !isHealthy {
		status = "unhealthy"
	}

	return HealthStatus{
		Status:              status,
		Timestamp:           lastCheckTime,
		Details:             fmt.Sprintf("Failure count: %d", failureCount),
		ErrorCount:          failureCount,
		LastCheck:           lastCheckTime,
		CircuitBreakerState: status,
	}
}

// GetMetrics returns current health metrics
func (h *SimpleHealthMonitor) GetMetrics() map[string]interface{} {
	// Read all atomic values
	isHealthy := atomic.LoadInt32(&h.isHealthy) == 1
	failureCount := atomic.LoadInt64(&h.failureCount)
	lastCheckNano := atomic.LoadInt64(&h.lastCheckTime)
	lastCheckTime := time.Unix(0, lastCheckNano)

	return map[string]interface{}{
		"is_healthy":    isHealthy,
		"failure_count": failureCount,
		"last_check":    lastCheckTime,
		"status":        h.GetStatus().Status,
	}
}

// RecordSuccess records a successful operation
func (h *SimpleHealthMonitor) RecordSuccess() {
	currentHealthy := atomic.LoadInt32(&h.isHealthy)
	if currentHealthy == 0 {
		h.logger.Info("Service recovered through success recording")
		atomic.StoreInt32(&h.isHealthy, 1)

		// Send recovery notification with debounce
		if h.systemNotifier != nil && h.shouldNotifyWithDebounce("healthy") {
			h.systemNotifier.NotifySystemHealth("healthy", map[string]interface{}{
				"component":       "mediamtx_health_monitor",
				"severity":        "info",
				"timestamp":       time.Now().Format(time.RFC3339),
				"reason":          "service_recovered",
				"previous_status": "unhealthy",
			})
		}
	}
	atomic.StoreInt64(&h.failureCount, 0)
}

// RecordFailure records a failed operation
func (h *SimpleHealthMonitor) RecordFailure() {
	// Increment failure count atomically
	failures := atomic.AddInt64(&h.failureCount, 1)

	// Simple threshold: 3 failures = unhealthy (or use configured threshold)
	threshold := int64(3)
	if h.config.HealthFailureThreshold > 0 {
		threshold = int64(h.config.HealthFailureThreshold)
	}

	if failures >= threshold {
		// Use atomic compare-and-swap to set unhealthy
		if atomic.CompareAndSwapInt32(&h.isHealthy, 1, 0) {
			h.logger.Warn("Service marked as unhealthy due to failure threshold")

			// Send threshold-crossing notification with debounce
			if h.systemNotifier != nil && h.shouldNotifyWithDebounce("unhealthy") {
				h.systemNotifier.NotifySystemHealth("unhealthy", map[string]interface{}{
					"failure_count": failures,
					"threshold":     threshold,
					"component":     "mediamtx_health_monitor",
					"severity":      "critical",
					"timestamp":     time.Now().Format(time.RFC3339),
					"reason":        "failure_threshold_exceeded",
				})
			}
		}
	}
}
