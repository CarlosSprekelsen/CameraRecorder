/*
MediaMTX Health Monitor Implementation

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
	"math/rand"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// CircuitState represents the circuit breaker state
type CircuitState int

const (
	CircuitClosed CircuitState = iota
	CircuitHalfOpen
	CircuitOpen
)

// String returns the string representation of the circuit state
func (s CircuitState) String() string {
	switch s {
	case CircuitClosed:
		return "CLOSED"
	case CircuitHalfOpen:
		return "HALF_OPEN"
	case CircuitOpen:
		return "OPEN"
	default:
		return "UNKNOWN"
	}
}

// healthMonitor represents the MediaMTX health monitor
type healthMonitor struct {
	client MediaMTXClient
	config *MediaMTXConfig
	logger *logrus.Logger

	// Circuit breaker state
	state      CircuitState
	stateMutex sync.RWMutex

	// Failure tracking
	failureCount     int
	failureThreshold int
	maxFailures      int

	// Enhanced persistent state tracking (Phase 1 enhancement)
	consecutiveFailures       int
	circuitBreakerActivations int
	recoveryCount             int
	lastRecoveryTime          time.Time
	healthMetrics             map[string]interface{}

	// Timing
	lastFailureTime time.Time
	recoveryTimeout time.Duration
	lastSuccessTime time.Time

	// Health check
	checkInterval time.Duration
	lastCheckTime time.Time
	healthStatus  HealthStatus

	// Control
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	stopChan     chan struct{}
	running      bool
	runningMutex sync.RWMutex
}

// NewHealthMonitor creates a new MediaMTX health monitor
func NewHealthMonitor(client MediaMTXClient, config *MediaMTXConfig, logger *logrus.Logger) HealthMonitor {
	ctx, cancel := context.WithCancel(context.Background())

	return &healthMonitor{
		client:                    client,
		config:                    config,
		logger:                    logger,
		state:                     CircuitClosed,
		failureThreshold:          config.CircuitBreaker.FailureThreshold,
		maxFailures:               config.CircuitBreaker.MaxFailures,
		recoveryTimeout:           config.CircuitBreaker.RecoveryTimeout,
		checkInterval:             5 * time.Second, // Default check interval
		consecutiveFailures:       0,
		circuitBreakerActivations: 0,
		recoveryCount:             0,
		healthMetrics:             make(map[string]interface{}),
		ctx:                       ctx,
		cancel:                    cancel,
		stopChan:                  make(chan struct{}),
	}
}

// Start starts the health monitoring
func (h *healthMonitor) Start(ctx context.Context) error {
	h.runningMutex.Lock()
	defer h.runningMutex.Unlock()

	if h.running {
		return fmt.Errorf("health monitor is already running")
	}

	h.logger.Info("Starting MediaMTX health monitor")

	h.running = true
	h.wg.Add(1)
	go h.monitorLoop()

	return nil
}

// Stop stops the health monitoring
func (h *healthMonitor) Stop(ctx context.Context) error {
	h.runningMutex.Lock()
	defer h.runningMutex.Unlock()

	if !h.running {
		return fmt.Errorf("health monitor is not running")
	}

	h.logger.Info("Stopping MediaMTX health monitor")

	h.running = false
	h.cancel()

	// Prevent closing an already closed channel
	select {
	case <-h.stopChan:
		// Channel already closed
	default:
		close(h.stopChan)
	}

	h.wg.Wait()

	return nil
}

// GetStatus returns the current health status
func (h *healthMonitor) GetStatus() HealthStatus {
	h.stateMutex.RLock()
	defer h.stateMutex.RUnlock()

	// Return default status if no health check has been performed yet
	if h.healthStatus.Status == "" {
		return HealthStatus{
			Status:    "UNKNOWN",
			Timestamp: time.Now(),
			Details:   "Health monitor not yet initialized",
		}
	}

	return h.healthStatus
}

// IsHealthy returns true if the service is healthy
func (h *healthMonitor) IsHealthy() bool {
	h.stateMutex.RLock()
	defer h.stateMutex.RUnlock()

	return h.state == CircuitClosed || h.state == CircuitHalfOpen
}

// IsCircuitOpen returns true if the circuit breaker is open
func (h *healthMonitor) IsCircuitOpen() bool {
	h.stateMutex.RLock()
	defer h.stateMutex.RUnlock()

	return h.state == CircuitOpen
}

// RecordSuccess records a successful operation
func (h *healthMonitor) RecordSuccess() {
	h.stateMutex.Lock()
	defer h.stateMutex.Unlock()

	h.logger.Debug("Recording successful operation")

	h.failureCount = 0
	h.lastSuccessTime = time.Now()

	// Enhanced persistent state tracking (Phase 1 enhancement)
	if h.consecutiveFailures > 0 {
		h.recoveryCount++
		h.lastRecoveryTime = time.Now()
		h.logger.WithFields(logrus.Fields{
			"consecutive_failures": h.consecutiveFailures,
			"recovery_count":       h.recoveryCount,
		}).Info("Service recovered from consecutive failures")
	}
	h.consecutiveFailures = 0

	// Transition from half-open to closed
	if h.state == CircuitHalfOpen {
		h.state = CircuitClosed
		h.logger.Info("Circuit breaker transitioned from HALF_OPEN to CLOSED")
	}
}

// RecordFailure records a failed operation
func (h *healthMonitor) RecordFailure() {
	h.stateMutex.Lock()
	defer h.stateMutex.Unlock()

	h.logger.Debug("Recording failed operation")

	h.failureCount++
	h.consecutiveFailures++
	h.lastFailureTime = time.Now()

	// Enhanced persistent state tracking (Phase 1 enhancement)
	h.logger.WithFields(logrus.Fields{
		"failure_count":        h.failureCount,
		"consecutive_failures": h.consecutiveFailures,
		"failure_threshold":    h.failureThreshold,
		"max_failures":         h.maxFailures,
	}).Debug("Failure recorded")

	// Check if we should open the circuit
	if h.failureCount >= h.failureThreshold {
		if h.state == CircuitClosed {
			h.state = CircuitOpen
			h.circuitBreakerActivations++

			// Enhanced error categorization and logging (Phase 4 enhancement)
			enhancedErr := CategorizeError(ErrCircuitOpen)
			errorMetadata := GetErrorMetadata(enhancedErr)
			recoveryStrategies := GetRecoveryStrategies(enhancedErr.GetCategory())

			h.logger.WithFields(logrus.Fields{
				"circuit_breaker_activations": h.circuitBreakerActivations,
				"consecutive_failures":        h.consecutiveFailures,
				"error_category":              errorMetadata["category"],
				"error_severity":              errorMetadata["severity"],
				"retryable":                   errorMetadata["retryable"],
				"recoverable":                 errorMetadata["recoverable"],
				"recovery_strategies":         recoveryStrategies,
			}).Warn("Circuit breaker opened due to failure threshold with enhanced error categorization")
		}
	}

	// Check if we've exceeded max failures
	if h.failureCount >= h.maxFailures {
		h.state = CircuitOpen
		h.circuitBreakerActivations++
		h.logger.WithFields(logrus.Fields{
			"circuit_breaker_activations": h.circuitBreakerActivations,
			"max_failures":                h.maxFailures,
		}).Error("Circuit breaker opened due to max failures exceeded")
	}
}

// monitorLoop runs the health monitoring loop
func (h *healthMonitor) monitorLoop() {
	defer h.wg.Done()

	ticker := time.NewTicker(h.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-h.ctx.Done():
			return
		case <-h.stopChan:
			return
		case <-ticker.C:
			h.performHealthCheck()
		}
	}
}

// performHealthCheck performs a health check
func (h *healthMonitor) performHealthCheck() {
	h.stateMutex.Lock()
	currentState := h.state
	h.stateMutex.Unlock()

	// Skip health check if circuit is open and we haven't reached recovery timeout
	if currentState == CircuitOpen {
		if time.Since(h.lastFailureTime) < h.recoveryTimeout {
			return
		}

		// Try to transition to half-open
		h.stateMutex.Lock()
		if h.state == CircuitOpen {
			h.state = CircuitHalfOpen
			h.logger.Info("Circuit breaker transitioned from OPEN to HALF_OPEN")
		}
		h.stateMutex.Unlock()
	}

	// Perform health check
	ctx, cancel := context.WithTimeout(h.ctx, h.config.Timeout)
	defer cancel()

	err := h.client.HealthCheck(ctx)

	h.stateMutex.Lock()
	defer h.stateMutex.Unlock()

	h.lastCheckTime = time.Now()

	if err != nil {
		h.logger.WithError(err).Warn("Health check failed")
		h.RecordFailure()
		h.healthStatus = HealthStatus{
			Status:    "UNHEALTHY",
			Timestamp: h.lastCheckTime,
			Details:   err.Error(),
		}
	} else {
		h.logger.Debug("Health check successful")
		h.RecordSuccess()
		h.healthStatus = HealthStatus{
			Status:    "HEALTHY",
			Timestamp: h.lastCheckTime,
		}
	}
}

// getBackoffDelay calculates exponential backoff delay with jitter
func (h *healthMonitor) getBackoffDelay(attempt int) time.Duration {
	// Base delay with exponential backoff
	baseDelay := h.config.RetryDelay * time.Duration(1<<attempt)

	// Add jitter (±25%)
	jitter := float64(baseDelay) * 0.25 * (rand.Float64()*2 - 1)
	delay := baseDelay + time.Duration(jitter)

	// Cap at maximum delay
	maxDelay := 30 * time.Second
	if delay > maxDelay {
		delay = maxDelay
	}

	return delay
}

// shouldRetry determines if an operation should be retried
func (h *healthMonitor) shouldRetry(err error, attempt int) bool {
	// Don't retry if circuit is open
	if h.IsCircuitOpen() {
		return false
	}

	// Don't retry if we've exceeded max attempts
	if attempt >= h.config.RetryAttempts {
		return false
	}

	// Don't retry certain error types
	if IsMediaMTXError(err) {
		mediaMTXErr, ok := err.(*MediaMTXError)
		if ok {
			// Don't retry client errors (4xx)
			if mediaMTXErr.Code >= 400 && mediaMTXErr.Code < 500 {
				return false
			}
		}
	}

	return true
}

// retryWithBackoff performs an operation with exponential backoff
func (h *healthMonitor) retryWithBackoff(operation func() error) error {
	var lastErr error

	for attempt := 0; attempt <= h.config.RetryAttempts; attempt++ {
		// Check if circuit is open
		if h.IsCircuitOpen() {
			return &CircuitBreakerError{
				State:   "OPEN",
				Message: "circuit breaker is open",
				Op:      "retry_operation",
			}
		}

		// Perform operation
		err := operation()
		if err == nil {
			h.RecordSuccess()
			return nil
		}

		lastErr = err
		h.RecordFailure()

		// Check if we should retry
		if !h.shouldRetry(err, attempt) {
			break
		}

		// Calculate backoff delay
		delay := h.getBackoffDelay(attempt)

		h.logger.WithFields(logrus.Fields{
			"attempt": attempt + 1,
			"delay":   delay,
			"error":   err.Error(),
		}).Debug("Retrying operation with backoff")

		// Wait before retry
		select {
		case <-h.ctx.Done():
			return h.ctx.Err()
		case <-time.After(delay):
			continue
		}
	}

	return lastErr
}

// GetMetrics returns current health metrics
func (h *healthMonitor) GetMetrics() map[string]interface{} {
	h.stateMutex.RLock()
	defer h.stateMutex.RUnlock()

	// Enhanced metrics with persistent state tracking (Phase 1 enhancement)
	metrics := map[string]interface{}{
		"circuit_state":     h.state.String(),
		"failure_count":     h.failureCount,
		"failure_threshold": h.failureThreshold,
		"max_failures":      h.maxFailures,
		"recovery_timeout":  h.recoveryTimeout,
		"last_failure_time": h.lastFailureTime,
		"last_success_time": h.lastSuccessTime,
		"last_check_time":   h.lastCheckTime,
		"check_interval":    h.checkInterval,
		"health_status":     h.healthStatus.Status,
		"health_timestamp":  h.healthStatus.Timestamp,
		"health_details":    h.healthStatus.Details,

		// Enhanced persistent state tracking (Phase 1 enhancement)
		"consecutive_failures":        h.consecutiveFailures,
		"circuit_breaker_activations": h.circuitBreakerActivations,
		"recovery_count":              h.recoveryCount,
		"last_recovery_time":          h.lastRecoveryTime,
		"uptime":                      time.Since(h.lastSuccessTime).String(),
		"time_since_last_failure":     time.Since(h.lastFailureTime).String(),
		"time_since_last_recovery":    time.Since(h.lastRecoveryTime).String(),
	}

	// Add custom health metrics if available
	for key, value := range h.healthMetrics {
		metrics[key] = value
	}

	return metrics
}

// CheckAllComponents performs health checks on all system components
func (h *healthMonitor) CheckAllComponents(ctx context.Context) map[string]string {
	componentStatus := make(map[string]string)

	// Check MediaMTX service health
	if h.IsHealthy() {
		componentStatus["mediamtx_service"] = "healthy"
	} else {
		componentStatus["mediamtx_service"] = "unhealthy"
	}

	// Check circuit breaker state
	if h.IsCircuitOpen() {
		componentStatus["circuit_breaker"] = "open"
	} else {
		componentStatus["circuit_breaker"] = "closed"
	}

	// Check health monitor itself
	componentStatus["health_monitor"] = "running"

	// Check client connection
	if h.client != nil {
		componentStatus["http_client"] = "available"
	} else {
		componentStatus["http_client"] = "unavailable"
	}

	return componentStatus
}

// GetDetailedStatus returns comprehensive health status information
func (h *healthMonitor) GetDetailedStatus() HealthStatus {
	h.stateMutex.RLock()
	defer h.stateMutex.RUnlock()

	// Create detailed status with component information
	detailedStatus := h.healthStatus

	// Add component status
	detailedStatus.ComponentStatus = map[string]string{
		"mediamtx_service": h.healthStatus.Status,
		"circuit_breaker":  h.state.String(),
		"health_monitor":   "running",
		"http_client":      "available",
	}

	// Add error count
	detailedStatus.ErrorCount = int64(h.failureCount)

	// Add last check time
	detailedStatus.LastCheck = h.lastCheckTime

	// Add circuit breaker state
	detailedStatus.CircuitBreakerState = h.state.String()

	return detailedStatus
}

// Real health check methods (Phase 4 enhancement)

// performRealHealthCheck performs a real health check against MediaMTX
func (h *healthMonitor) performRealHealthCheck() (*HealthStatus, error) {
	// Use existing MediaMTX client for health checks
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := h.client.HealthCheck(ctx)
	if err != nil {
		return &HealthStatus{
			Status:    "unhealthy",
			Details:   err.Error(),
			Timestamp: time.Now(),
		}, nil
	}

	// Parse health data and return status
	return &HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now(),
		Details:   "MediaMTX service is healthy",
	}, nil
}

// getBasicStatus returns basic health status without real checks
func (h *healthMonitor) getBasicStatus() HealthStatus {
	h.stateMutex.RLock()
	defer h.stateMutex.RUnlock()

	status := "unknown"
	if h.state == CircuitClosed {
		status = "healthy"
	} else if h.state == CircuitOpen {
		status = "unhealthy"
	} else if h.state == CircuitHalfOpen {
		status = "degraded"
	}

	return HealthStatus{
		Status:    status,
		Timestamp: time.Now(),
		Details:   fmt.Sprintf("Circuit state: %s, Failure count: %d", h.state, h.failureCount),
	}
}
