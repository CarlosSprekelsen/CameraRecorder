/*
MediaMTX Circuit Breaker Implementation

This package provides circuit breaker functionality for recording operations
to prevent cascade failures and improve system resilience.

Requirements Coverage:
- REQ-MTX-007: Error handling and recovery
- REQ-MTX-008: Logging and monitoring

Test Categories: Unit/Integration
*/

package mediamtx

import (
	"fmt"
	"sync"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
)

// CircuitBreakerState represents the current state of a circuit breaker
type CircuitBreakerState string

const (
	StateClosed   CircuitBreakerState = "closed"
	StateOpen     CircuitBreakerState = "open"
	StateHalfOpen CircuitBreakerState = "half-open"
)

// CircuitBreakerConfig is defined in types.go

// CircuitBreaker provides circuit breaker functionality for operations
type CircuitBreaker struct {
	config          CircuitBreakerConfig
	logger          *logging.Logger
	name            string
	state           CircuitBreakerState
	failureCount    int
	lastFailureTime time.Time
	mutex           sync.RWMutex
}

// NewCircuitBreaker creates a new circuit breaker instance
func NewCircuitBreaker(name string, config CircuitBreakerConfig, logger *logging.Logger) *CircuitBreaker {
	return &CircuitBreaker{
		config: config,
		logger: logger,
		name:   name,
		state:  StateClosed,
	}
}

// Call executes an operation with circuit breaker protection
func (cb *CircuitBreaker) Call(operation func() error) error {
	state := cb.getState()

	// Check if circuit breaker should allow the operation
	if state == StateOpen {
		// Check if enough time has passed to try half-open
		if time.Since(cb.lastFailureTime) > cb.config.RecoveryTimeout {
			cb.setState(StateHalfOpen)
			cb.logger.WithFields(logging.Fields{
				"circuit_breaker": cb.name,
				"state":           StateHalfOpen,
			}).Info("Circuit breaker transitioning to half-open state")
		} else {
			return &CircuitBreakerError{
				Name:  cb.name,
				State: StateOpen,
				Msg:   "circuit breaker is open",
			}
		}
	}

	// Execute the operation
	err := operation()

	if err != nil {
		cb.recordFailure()
		return err
	}

	cb.recordSuccess()
	return nil
}

// getState returns the current state of the circuit breaker
func (cb *CircuitBreaker) getState() CircuitBreakerState {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.state
}

// setState sets the state of the circuit breaker
func (cb *CircuitBreaker) setState(state CircuitBreakerState) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()
	cb.state = state
}

// recordFailure records a failed operation
func (cb *CircuitBreaker) recordFailure() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.failureCount++
	cb.lastFailureTime = time.Now()

	cb.logger.WithFields(logging.Fields{
		"circuit_breaker": cb.name,
		"failure_count":   cb.failureCount,
		"state":           cb.state,
	}).Debug("Circuit breaker recorded failure")

	// Check if we should open the circuit breaker
	if cb.failureCount >= cb.config.FailureThreshold {
		cb.state = StateOpen
		cb.logger.WithFields(logging.Fields{
			"circuit_breaker":   cb.name,
			"failure_count":     cb.failureCount,
			"failure_threshold": cb.config.FailureThreshold,
		}).Warn("Circuit breaker opened due to failure threshold")
	}
}

// recordSuccess records a successful operation
func (cb *CircuitBreaker) recordSuccess() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	// Reset failure count on success
	if cb.failureCount > 0 {
		cb.logger.WithFields(logging.Fields{
			"circuit_breaker":   cb.name,
			"previous_failures": cb.failureCount,
		}).Info("Circuit breaker reset failure count on success")
		cb.failureCount = 0
	}

	// Close circuit breaker if it was half-open
	if cb.state == StateHalfOpen {
		cb.state = StateClosed
		cb.logger.WithFields(logging.Fields{
			"circuit_breaker": cb.name,
			"state":           StateClosed,
		}).Info("Circuit breaker closed after successful operation")
	}
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	return cb.getState()
}

// GetFailureCount returns the current failure count
func (cb *CircuitBreaker) GetFailureCount() int {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return cb.failureCount
}

// Reset resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	cb.state = StateClosed
	cb.failureCount = 0
	cb.lastFailureTime = time.Time{}

	cb.logger.WithFields(logging.Fields{
		"circuit_breaker": cb.name,
	}).Info("Circuit breaker manually reset")
}

// CircuitBreakerError represents an error from a circuit breaker
type CircuitBreakerError struct {
	Name  string
	State CircuitBreakerState
	Msg   string
}

func (e *CircuitBreakerError) Error() string {
	return fmt.Sprintf("circuit breaker '%s' is %s: %s", e.Name, e.State, e.Msg)
}
