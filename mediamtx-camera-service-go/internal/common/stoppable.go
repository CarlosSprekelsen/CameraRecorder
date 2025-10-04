package common

import (
	"context"
	"time"
)

// Stoppable defines the interface for services that can be gracefully stopped.
//
// Implementations must handle context cancellation and timeout enforcement.
// The context should be used for timeout enforcement, cancellation propagation,
// and resource cleanup coordination. Returns an error if shutdown fails.
type Stoppable interface {
	Stop(ctx context.Context) error
}

// StopWithTimeout is a helper function that creates a timeout context
// and calls Stop on a Stoppable service.
func StopWithTimeout(service Stoppable, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return service.Stop(ctx)
}
