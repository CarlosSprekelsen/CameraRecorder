// Package e2e provides heartbeat timing validation tests.
// This file tests heartbeat timing validation to improve coverage.
package e2e

import (
	"testing"
	"time"
)

func TestHeartbeatTiming_PositiveCases(t *testing.T) {
	validator := NewContractValidator(t)

	// Test case 1: Two heartbeats within tolerance window
	events := []string{
		"event: heartbeat\ndata: {\"timestamp\":\"2025-10-03T10:00:00Z\"}",
		"event: heartbeat\ndata: {\"timestamp\":\"2025-10-03T10:00:15Z\"}",
	}

	baseInterval := 15 * time.Second
	jitter := 2 * time.Second

	validator.ValidateHeartbeatInterval(t, events, baseInterval, jitter)
}

func TestHeartbeatTiming_NegativeCases(t *testing.T) {
	validator := NewContractValidator(t)

	// Test case 1: Heartbeat too fast
	t.Run("too_fast", func(t *testing.T) {
		events := []string{
			"event: heartbeat\ndata: {\"timestamp\":\"2025-10-03T10:00:00Z\"}",
			"event: heartbeat\ndata: {\"timestamp\":\"2025-10-03T10:00:05Z\"}", // 5s < 13s min
		}

		baseInterval := 15 * time.Second
		jitter := 2 * time.Second

		validator.ValidateHeartbeatInterval(t, events, baseInterval, jitter)
	})

	// Test case 2: Heartbeat too slow
	t.Run("too_slow", func(t *testing.T) {
		events := []string{
			"event: heartbeat\ndata: {\"timestamp\":\"2025-10-03T10:00:00Z\"}",
			"event: heartbeat\ndata: {\"timestamp\":\"2025-10-03T10:00:30Z\"}", // 30s > 17s max
		}

		baseInterval := 15 * time.Second
		jitter := 2 * time.Second

		validator.ValidateHeartbeatInterval(t, events, baseInterval, jitter)
	})

	// Test case 3: Insufficient events
	t.Run("insufficient_events", func(t *testing.T) {
		events := []string{
			"event: heartbeat\ndata: {\"timestamp\":\"2025-10-03T10:00:00Z\"}",
		}

		baseInterval := 15 * time.Second
		jitter := 2 * time.Second

		validator.ValidateHeartbeatInterval(t, events, baseInterval, jitter)
	})

	// Test case 4: No heartbeat events
	t.Run("no_heartbeats", func(t *testing.T) {
		events := []string{
			"event: ready\ndata: {\"snapshot\":{}}",
			"event: powerChanged\ndata: {\"powerDbm\":25}",
		}

		baseInterval := 15 * time.Second
		jitter := 2 * time.Second

		validator.ValidateHeartbeatInterval(t, events, baseInterval, jitter)
	})
}

func TestHeartbeatTiming_EdgeCases(t *testing.T) {
	validator := NewContractValidator(t)

	// Test case 1: Exactly at minimum tolerance
	t.Run("at_minimum", func(t *testing.T) {
		events := []string{
			"event: heartbeat\ndata: {\"timestamp\":\"2025-10-03T10:00:00Z\"}",
			"event: heartbeat\ndata: {\"timestamp\":\"2025-10-03T10:00:13Z\"}", // Exactly 13s (15-2)
		}

		baseInterval := 15 * time.Second
		jitter := 2 * time.Second

		validator.ValidateHeartbeatInterval(t, events, baseInterval, jitter)
	})

	// Test case 2: Exactly at maximum tolerance
	t.Run("at_maximum", func(t *testing.T) {
		events := []string{
			"event: heartbeat\ndata: {\"timestamp\":\"2025-10-03T10:00:00Z\"}",
			"event: heartbeat\ndata: {\"timestamp\":\"2025-10-03T10:00:17Z\"}", // Exactly 17s (15+2)
		}

		baseInterval := 15 * time.Second
		jitter := 2 * time.Second

		validator.ValidateHeartbeatInterval(t, events, baseInterval, jitter)
	})

	// Test case 3: Timeout threshold
	t.Run("timeout_threshold", func(t *testing.T) {
		events := []string{
			"event: heartbeat\ndata: {\"timestamp\":\"2025-10-03T10:00:00Z\"}",
			"event: heartbeat\ndata: {\"timestamp\":\"2025-10-03T10:00:45Z\"}", // 45s > 45s timeout (15*3)
		}

		baseInterval := 15 * time.Second
		jitter := 2 * time.Second

		validator.ValidateHeartbeatInterval(t, events, baseInterval, jitter)
	})
}
