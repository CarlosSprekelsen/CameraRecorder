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

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
)

// SimpleHealthMonitor manages MediaMTX service health monitoring and connectivity.
//
// RESPONSIBILITIES:
// - MediaMTX service connectivity monitoring and circuit breaker functionality
// - Service-specific health status tracking (MediaMTX availability)
// - Success/failure rate tracking for MediaMTX operations
// - Health status API responses for service availability
//
// SCOPE:
// - Focused on MediaMTX service health only
// - Does NOT handle system-wide metrics (CPU, memory, disk)
// - Does NOT handle storage operations or file system monitoring
//
// API INTEGRATION:
// - Returns JSON-RPC API-ready responses for health queries
// - Provides real-time health status for MediaMTX connectivity
type SimpleHealthMonitor struct {
	client MediaMTXClient
	config *config.MediaMTXConfig
	logger *logging.Logger

	// Context for lifecycle management
	ctx    context.Context
	cancel context.CancelFunc

	// Threshold-crossing notifications
	systemNotifier SystemEventNotifier

	// Debounce state for notifications (all atomic to prevent race conditions)
	lastNotificationTime   int64 // Atomic timestamp (UnixNano)
	lastNotificationStatus int32 // Atomic status: 0=healthy, 1=unhealthy, 2=unknown
	debounceDuration       time.Duration

	// Atomic state: optimized for high-frequency reads
	isHealthy     int32 // 0 = false, 1 = true
	failureCount  int64 // Atomic counter
	lastCheckTime int64 // Atomic timestamp (UnixNano)

	// Keep mutex only for complex operations that need consistency
	mu sync.RWMutex

	// Control
	wg sync.WaitGroup

	// Event-driven health monitoring
	healthEventChan chan struct{}
	healthMutex     sync.RWMutex
}

// NewHealthMonitor creates a new simplified MediaMTX health monitor
func NewHealthMonitor(client MediaMTXClient, config *config.MediaMTXConfig, logger *logging.Logger) HealthMonitor {
	return &SimpleHealthMonitor{
		client:           client,
		config:           config,
		logger:           logger,
		isHealthy:        1, // Assume healthy initially (1 = true)
		failureCount:     0,
		lastCheckTime:    time.Now().UnixNano(),
		debounceDuration: 15 * time.Second,        // 15s debounce for health notifications
		healthEventChan:  make(chan struct{}, 10), // Buffered channel for health events
	}
}

// SetSystemNotifier sets the system event notifier for threshold-crossing notifications
func (h *SimpleHealthMonitor) SetSystemNotifier(notifier SystemEventNotifier) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.systemNotifier = notifier
}

// SubscribeToHealthChanges returns a channel that receives events when health status changes
func (h *SimpleHealthMonitor) SubscribeToHealthChanges() <-chan struct{} {
	h.healthMutex.RLock()
	defer h.healthMutex.RUnlock()

	subscriberChan := make(chan struct{}, 1)

	// If already healthy, send immediate event
	if h.IsHealthy() {
		select {
		case subscriberChan <- struct{}{}:
		default:
		}
	}

	return subscriberChan
}

// emitHealthEvent sends a health change event to subscribers
func (h *SimpleHealthMonitor) emitHealthEvent() {
	h.healthMutex.RLock()
	defer h.healthMutex.RUnlock()

	if h.healthEventChan != nil {
		select {
		case h.healthEventChan <- struct{}{}:
		default:
			// Channel is full, skip this event
		}
	}
}

// shouldNotifyWithDebounce checks if a notification should be sent based on debounce logic
// Uses atomic operations to prevent race conditions
func (h *SimpleHealthMonitor) shouldNotifyWithDebounce(status string) bool {
	now := time.Now().UnixNano()
	lastTime := atomic.LoadInt64(&h.lastNotificationTime)

	// Check if enough time has passed since last notification
	if now-lastTime < int64(h.debounceDuration) {
		return false
	}

	// Convert status string to atomic int32
	var statusValue int32
	switch status {
	case "healthy":
		statusValue = 0
	case "unhealthy":
		statusValue = 1
	default:
		statusValue = 2 // unknown
	}

	// Check if status actually changed using atomic compare-and-swap
	lastStatus := atomic.LoadInt32(&h.lastNotificationStatus)
	if lastStatus == statusValue {
		return false
	}

	// Update state atomically - use compare-and-swap to ensure atomicity
	if atomic.CompareAndSwapInt32(&h.lastNotificationStatus, lastStatus, statusValue) {
		atomic.StoreInt64(&h.lastNotificationTime, now)
		return true
	}

	// If CAS failed, another goroutine updated it, check again
	return atomic.LoadInt32(&h.lastNotificationStatus) != statusValue
}

// Start starts the health monitoring
func (h *SimpleHealthMonitor) Start(ctx context.Context) error {
	h.logger.Info("Starting simplified MediaMTX health monitor")

	// Create cancellable context
	h.ctx, h.cancel = context.WithCancel(ctx)

	h.wg.Add(1)
	go h.monitorLoop(h.ctx) // Pass the cancellable context
	return nil
}

// Stop stops the health monitoring
func (h *SimpleHealthMonitor) Stop(ctx context.Context) error {
	h.logger.Info("Stopping simplified MediaMTX health monitor")

	// Cancel context first - this interrupts checkHealth immediately!
	if h.cancel != nil {
		h.cancel()
	}

	// Wait with timeout
	done := make(chan struct{})
	go func() {
		h.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Clean shutdown
	case <-ctx.Done():
		// Force shutdown after timeout
		h.logger.Warn("Health monitor shutdown timeout, forcing stop")
	}

	h.logger.Info("simplified MediaMTX health monitor stopped")
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
		case <-ticker.C:
			h.checkHealth(ctx)
		}
	}
}

// checkHealth performs a health check
func (h *SimpleHealthMonitor) checkHealth(ctx context.Context) {
	// Create timeout context BUT inherit parent cancellation!
	timeout := h.config.HealthCheckTimeout
	checkCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// This will be cancelled immediately when Stop() is called!
	err := h.client.HealthCheck(checkCtx)

	// Check if cancelled
	if ctx.Err() != nil {
		return // Exit immediately on cancellation
	}

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
				// Emit health change event
				h.emitHealthEvent()
			}
		}
	} else {
		// Success - reset everything atomically
		currentHealthy := atomic.LoadInt32(&h.isHealthy)
		if currentHealthy == 0 {
			h.logger.Info("MediaMTX service recovered")
			// Emit health change event for recovery
			h.emitHealthEvent()
		}
		atomic.StoreInt32(&h.isHealthy, 1)
		atomic.StoreInt64(&h.failureCount, 0)
	}
}

// IsHealthy returns true if the service is healthy
func (h *SimpleHealthMonitor) IsHealthy() bool {
	return atomic.LoadInt32(&h.isHealthy) == 1
}

// WaitForHealthy waits for the health monitor to become healthy using event-driven approach
func (h *SimpleHealthMonitor) WaitForHealthy(ctx context.Context) error {
	// Return immediately if already healthy
	if h.IsHealthy() {
		return nil
	}

	// Create timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Use event-driven approach instead of polling
	healthChan := h.SubscribeToHealthChanges()

	// Wait for health event or timeout
	select {
	case <-healthChan:
		// Check if we're actually healthy now
		if h.IsHealthy() {
			return nil
		}
		// If not healthy, continue waiting with polling as fallback
	case <-timeoutCtx.Done():
		return fmt.Errorf("health monitor not healthy within timeout")
	}

	// Fallback to polling with short intervals if event-driven approach doesn't work
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
			return fmt.Errorf("health monitor not healthy within timeout")
		case <-ticker.C:
			if h.IsHealthy() {
				return nil
			}
		}
	}
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

		// Emit health change event for recovery
		h.emitHealthEvent()

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

			// Emit health change event
			h.emitHealthEvent()

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

// GetHealthAPI returns MediaMTX health status in API-ready format
func (h *SimpleHealthMonitor) GetHealthAPI(ctx context.Context, startTime time.Time) (*GetHealthResponse, error) {
	// Get current health status
	status := h.GetStatus()

	// Calculate uptime
	uptime := time.Since(startTime).String()

	// Build components map for MediaMTX-specific health
	components := map[string]interface{}{
		"mediamtx": map[string]interface{}{
			"status":        status.Status,
			"failure_count": status.ErrorCount,
			"last_check":    status.LastCheck.Format(time.RFC3339),
			"circuit_state": status.CircuitBreakerState,
		},
	}

	// Determine overall status
	overallStatus := "healthy"
	if !h.IsHealthy() {
		overallStatus = "unhealthy"
	}

	// Build API-ready response
	response := &GetHealthResponse{
		Status:       overallStatus,
		Uptime:       uptime,
		Version:      "1.0.0", // TODO-IMPL: Get from build info
		Components:   components,
		Checks:       []interface{}{}, // TODO-IMPL: Add health checks
		Timestamp:    time.Now().Format(time.RFC3339),
		ResponseTime: 0.0, // TODO-IMPL: Add response time calculation
	}

	h.logger.WithFields(logging.Fields{
		"overall_status":   overallStatus,
		"uptime":           uptime,
		"mediamtx_healthy": h.IsHealthy(),
	}).Debug("Health status collected successfully")

	return response, nil
}
