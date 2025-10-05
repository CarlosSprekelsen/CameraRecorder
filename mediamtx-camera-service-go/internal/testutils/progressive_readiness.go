/*
Progressive Readiness Test Utilities - Universal Pattern Implementation

Provides shared Progressive Readiness testing patterns that eliminate code duplication
across all modules while enforcing architectural behavioral invariants.

Requirements Coverage:
- REQ-ARCH-001: Progressive Readiness behavioral invariants
- REQ-TEST-005: Consistent Progressive Readiness validation
- REQ-TEST-006: Cross-module pattern sharing

Design Principles:
- Try operation immediately (Progressive Readiness)
- Fall back to event-driven waiting (not polling)
- Use universal constants (no magic numbers)
- Generic pattern for all component types
*/

package testutils

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// ReadinessSubscriber defines components that can notify about readiness
type ReadinessSubscriber interface {
	SubscribeToReadiness() <-chan struct{}
}

// ProgressiveReadinessResult holds the result of a Progressive Readiness test
type ProgressiveReadinessResult[T any] struct {
	Result       T
	Error        error
	UsedFallback bool
}

// TestProgressiveReadiness tests an operation using Progressive Readiness pattern
// This eliminates 15-20 lines of duplicated code in every test
func TestProgressiveReadiness[T any](
	t *testing.T,
	operation func() (T, error),
	component ReadinessSubscriber,
	operationName string,
) ProgressiveReadinessResult[T] {
	// BEHAVIORAL INVARIANT #2: Request handling always returns response, never blocks
	// Try operation immediately - no waiting for component initialization
	result, err := operation()
	if err == nil {
		// SUCCESS: Progressive Readiness working perfectly
		t.Logf("✅ %s succeeded immediately - Progressive Readiness working", operationName)
		return ProgressiveReadinessResult[T]{
			Result:       result,
			Error:        nil,
			UsedFallback: false,
		}
	}

	// FALLBACK: Component not ready yet - wait for readiness event (not polling!)
	t.Logf("⏳ %s needs readiness - waiting for event (not polling)", operationName)

	readinessChan := component.SubscribeToReadiness()
	select {
	case <-readinessChan:
		// BEHAVIORAL INVARIANT #4: Component independence - retry after readiness event
		t.Logf("🔔 Readiness event received for %s - retrying operation", operationName)
		result, err := operation()
		return ProgressiveReadinessResult[T]{
			Result:       result,
			Error:        err,
			UsedFallback: true,
		}
	case <-time.After(UniversalTimeoutVeryLong): // Use universal constant - no magic numbers
		// BEHAVIORAL INVARIANT #3: Error semantics - operation-specific errors
		t.Fatalf("❌ Timeout waiting for readiness event in %s after %v", operationName, UniversalTimeoutVeryLong)
		var zero T
		return ProgressiveReadinessResult[T]{
			Result:       zero,
			Error:        fmt.Errorf("timeout waiting for readiness in %s", operationName),
			UsedFallback: true,
		}
	}
}

// TestProgressiveReadinessSimple is a simplified version for operations that don't need result tracking
func TestProgressiveReadinessSimple(
	t *testing.T,
	operation func() error,
	component ReadinessSubscriber,
	operationName string,
) error {
	// Wrap simple operation for generic function
	wrappedOp := func() (struct{}, error) {
		err := operation()
		return struct{}{}, err
	}

	result := TestProgressiveReadiness(t, wrappedOp, component, operationName)
	return result.Error
}

// AssertProgressiveReadiness validates that an operation follows Progressive Readiness
// This enforces the behavioral invariants in test validation
func AssertProgressiveReadiness[T any](
	t *testing.T,
	result ProgressiveReadinessResult[T],
	expectImmediate bool,
) {
	if expectImmediate && result.UsedFallback {
		t.Errorf("❌ PROGRESSIVE READINESS VIOLATION: %s should succeed immediately but used fallback",
			"operation")
	}

	if result.Error != nil {
		t.Errorf("❌ Operation failed even after Progressive Readiness handling: %v", result.Error)
	}

	if !result.UsedFallback {
		t.Logf("✅ PROGRESSIVE READINESS SUCCESS: Operation succeeded immediately")
	} else {
		t.Logf("⚠️  PROGRESSIVE READINESS FALLBACK: Operation needed readiness event (acceptable)")
	}
}

// WaitForCondition polls condition function until true or timeout
func WaitForCondition(
	ctx context.Context,
	condition func() bool,
	timeout time.Duration,
	description string,
) error {
	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	
	for {
		// Check timeout FIRST
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout waiting for %s after %v", description, timeout)
		}
		
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled while waiting for %s: %w", description, ctx.Err())
		case <-ticker.C:
			if condition() {
				return nil
			}
		}
	}
}
