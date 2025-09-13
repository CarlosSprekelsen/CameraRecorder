/*
Common Stoppable Interface for Graceful Shutdown

This package provides a common interface for all services that need graceful shutdown
capabilities. This ensures consistent shutdown behavior across the entire system.

Architecture Compliance:
- AD-GO-007: Use context.Context for cancellation
- Quality Attributes: Graceful Degradation
- Concurrency Design: Context Cancellation for graceful operation termination
*/

package common

import (
	"context"
	"time"
)

// Stoppable defines the interface for services that can be gracefully stopped
// with context-aware cancellation and timeout enforcement.
type Stoppable interface {
	// Stop gracefully stops the service with context-aware cancellation.
	// The context should be used for:
	// - Timeout enforcement
	// - Cancellation propagation
	// - Resource cleanup coordination
	//
	// Returns an error if the service fails to stop within the context timeout.
	Stop(ctx context.Context) error
}

// StopWithTimeout is a helper function that creates a timeout context
// and calls Stop on a Stoppable service.
func StopWithTimeout(service Stoppable, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return service.Stop(ctx)
}
