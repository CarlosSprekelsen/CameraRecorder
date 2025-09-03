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
	"time"

	"github.com/sirupsen/logrus"
)

// SimpleHealthMonitor represents the simplified MediaMTX health monitor
type SimpleHealthMonitor struct {
	client MediaMTXClient
	config *MediaMTXConfig
	logger *logrus.Logger

	// Simple state: just healthy or not
	isHealthy     bool
	failureCount  int
	lastCheckTime time.Time
	mu            sync.RWMutex

	// Control
	stopChan chan struct{}
	wg       sync.WaitGroup
}

// NewHealthMonitor creates a new simplified MediaMTX health monitor
func NewHealthMonitor(client MediaMTXClient, config *MediaMTXConfig, logger *logrus.Logger) HealthMonitor {
	return &SimpleHealthMonitor{
		client:       client,
		config:       config,
		logger:       logger,
		isHealthy:    true, // Assume healthy initially
		failureCount: 0,
		stopChan:     make(chan struct{}, 1),
	}
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
	
	close(h.stopChan)
	h.wg.Wait()
	return nil
}

// monitorLoop runs the health monitoring loop
func (h *SimpleHealthMonitor) monitorLoop(ctx context.Context) {
	defer h.wg.Done()
	
	// Use configured interval or default to 5 seconds
	checkInterval := 5 * time.Second
	if h.config.HealthCheckInterval > 0 {
		checkInterval = time.Duration(h.config.HealthCheckInterval) * time.Second
	}
	
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
	// Use configured timeout or default to 5 seconds
	timeout := 5 * time.Second
	if h.config.HealthCheckTimeout > 0 {
		timeout = h.config.HealthCheckTimeout
	}
	
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	
	err := h.client.HealthCheck(ctx)
	
	h.mu.Lock()
	defer h.mu.Unlock()
	
	h.lastCheckTime = time.Now()
	
	if err != nil {
		h.failureCount++
		h.logger.WithError(err).Debug("Health check failed")
		
		// Simple threshold: 3 failures = unhealthy (or use configured threshold)
		threshold := 3
		if h.config.HealthFailureThreshold > 0 {
			threshold = h.config.HealthFailureThreshold
		}
		
		if h.failureCount >= threshold {
			if h.isHealthy {
				h.logger.Warn("MediaMTX service marked as unhealthy")
				h.isHealthy = false
			}
		}
	} else {
		// Success - reset everything
		if !h.isHealthy {
			h.logger.Info("MediaMTX service recovered")
		}
		h.isHealthy = true
		h.failureCount = 0
	}
}

// IsHealthy returns true if the service is healthy
func (h *SimpleHealthMonitor) IsHealthy() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.isHealthy
}

// IsCircuitOpen returns true if the circuit breaker is open (unhealthy)
func (h *SimpleHealthMonitor) IsCircuitOpen() bool {
	return !h.IsHealthy()
}

// GetStatus returns the current health status
func (h *SimpleHealthMonitor) GetStatus() HealthStatus {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	status := "healthy"
	if !h.isHealthy {
		status = "unhealthy"
	}
	
	return HealthStatus{
		Status:              status,
		Timestamp:           h.lastCheckTime,
		Details:             fmt.Sprintf("Failure count: %d", h.failureCount),
		ErrorCount:          int64(h.failureCount),
		LastCheck:           h.lastCheckTime,
		CircuitBreakerState: status,
	}
}

// GetMetrics returns current health metrics
func (h *SimpleHealthMonitor) GetMetrics() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	return map[string]interface{}{
		"is_healthy":     h.isHealthy,
		"failure_count":  h.failureCount,
		"last_check":     h.lastCheckTime,
		"status":         h.GetStatus().Status,
	}
}

// RecordSuccess records a successful operation
func (h *SimpleHealthMonitor) RecordSuccess() {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	if !h.isHealthy {
		h.logger.Info("Service recovered through success recording")
		h.isHealthy = true
	}
	h.failureCount = 0
}

// RecordFailure records a failed operation
func (h *SimpleHealthMonitor) RecordFailure() {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	h.failureCount++
	
	// Simple threshold: 3 failures = unhealthy (or use configured threshold)
	threshold := 3
	if h.config.HealthFailureThreshold > 0 {
		threshold = h.config.HealthFailureThreshold
	}
	
	if h.failureCount >= threshold {
		if h.isHealthy {
			h.logger.Warn("Service marked as unhealthy due to failure threshold")
			h.isHealthy = false
		}
	}
}
