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
	client           MediaMTXClient
	config           *MediaMTXConfig
	logger           *logrus.Logger
	
	// Circuit breaker state
	state            CircuitState
	stateMutex       sync.RWMutex
	
	// Failure tracking
	failureCount     int
	failureThreshold int
	maxFailures      int
	
	// Timing
	lastFailureTime  time.Time
	recoveryTimeout  time.Duration
	lastSuccessTime  time.Time
	
	// Health check
	checkInterval    time.Duration
	lastCheckTime    time.Time
	healthStatus     HealthStatus
	
	// Control
	ctx              context.Context
	cancel           context.CancelFunc
	wg               sync.WaitGroup
	stopChan         chan struct{}
}

// NewHealthMonitor creates a new MediaMTX health monitor
func NewHealthMonitor(client MediaMTXClient, config *MediaMTXConfig, logger *logrus.Logger) HealthMonitor {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &healthMonitor{
		client:           client,
		config:           config,
		logger:           logger,
		state:            CircuitClosed,
		failureThreshold: config.CircuitBreaker.FailureThreshold,
		maxFailures:      config.CircuitBreaker.MaxFailures,
		recoveryTimeout:  config.CircuitBreaker.RecoveryTimeout,
		checkInterval:    5 * time.Second, // Default check interval
		ctx:              ctx,
		cancel:           cancel,
		stopChan:         make(chan struct{}),
	}
}

// Start starts the health monitoring
func (h *healthMonitor) Start(ctx context.Context) error {
	h.logger.Info("Starting MediaMTX health monitor")
	
	h.wg.Add(1)
	go h.monitorLoop()
	
	return nil
}

// Stop stops the health monitoring
func (h *healthMonitor) Stop(ctx context.Context) error {
	h.logger.Info("Stopping MediaMTX health monitor")
	
	h.cancel()
	close(h.stopChan)
	h.wg.Wait()
	
	return nil
}

// GetStatus returns the current health status
func (h *healthMonitor) GetStatus() HealthStatus {
	h.stateMutex.RLock()
	defer h.stateMutex.RUnlock()
	
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
	h.lastFailureTime = time.Now()
	
	// Check if we should open the circuit
	if h.failureCount >= h.failureThreshold {
		if h.state == CircuitClosed {
			h.state = CircuitOpen
			h.logger.Warn("Circuit breaker opened due to failure threshold")
		}
	}
	
	// Check if we've exceeded max failures
	if h.failureCount >= h.maxFailures {
		h.state = CircuitOpen
		h.logger.Error("Circuit breaker opened due to max failures exceeded")
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
	
	return map[string]interface{}{
		"circuit_state":        h.state.String(),
		"failure_count":        h.failureCount,
		"failure_threshold":    h.failureThreshold,
		"max_failures":         h.maxFailures,
		"recovery_timeout":     h.recoveryTimeout,
		"last_failure_time":    h.lastFailureTime,
		"last_success_time":    h.lastSuccessTime,
		"last_check_time":      h.lastCheckTime,
		"check_interval":       h.checkInterval,
		"health_status":        h.healthStatus.Status,
		"health_timestamp":     h.healthStatus.Timestamp,
		"health_details":       h.healthStatus.Details,
	}
}
