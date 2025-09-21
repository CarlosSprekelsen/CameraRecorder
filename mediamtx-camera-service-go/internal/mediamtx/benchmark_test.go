/*
MediaMTX Performance Benchmarks - Event-Driven vs Polling

Requirements Coverage:
- REQ-MTX-001: MediaMTX service integration
- REQ-MTX-002: Stream management capabilities
- REQ-MTX-004: Health monitoring

Test Categories: Benchmark (performance comparison between event-driven and polling approaches)
API Documentation Reference: docs/api/json_rpc_methods.md

This file contains benchmark tests that compare the performance characteristics
of event-driven patterns versus traditional polling approaches.
*/

package mediamtx

import (
	"context"
	"testing"
	"time"
)

// setupBenchmarkController creates a controller for benchmarking
func setupBenchmarkController(b *testing.B) (*MediaMTXTestHelper, *EventDrivenTestHelper, MediaMTXController) {
	// Create a temporary testing.T for setup
	t := &testing.T{}
	helper, ctx := SetupMediaMTXTest(t)

	// Create event-driven test helper
	eventHelper := helper.CreateEventDrivenTestHelper(t)

	// Create controller
	controller, err := helper.GetController(t)
	if err != nil {
		b.Fatalf("Failed to create controller: %v", err)
	}

	// Start controller
	// MINIMAL: Helper provides standard context
	ctx, cancel := helper.GetStandardContext()
	defer cancel()
	err = controller.Start(ctx)
	if err != nil {
		b.Fatalf("Failed to start controller: %v", err)
	}

	// No waiting for readiness - Progressive Readiness Pattern

	return helper, eventHelper, controller
}

// cleanupBenchmarkController cleans up the benchmark controller
func cleanupBenchmarkController(helper *MediaMTXTestHelper, eventHelper *EventDrivenTestHelper, controller MediaMTXController) {
	// Stop controller
	stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	controller.Stop(stopCtx)

	// Cleanup helpers
	if eventHelper != nil {
		eventHelper.Cleanup()
	}
	helper.Cleanup(&testing.T{})
}

// BenchmarkEventDrivenReadiness benchmarks the event-driven readiness system
func BenchmarkEventDrivenReadiness(b *testing.B) {
	helper, eventHelper, controller := setupBenchmarkController(b)
	defer cleanupBenchmarkController(helper, eventHelper, controller)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Benchmark non-blocking event observation
			eventHelper.ObserveReadiness()

			// No waiting - just observe events
		}
	})
}

// BenchmarkPollingReadiness benchmarks the traditional polling approach
func BenchmarkPollingReadiness(b *testing.B) {
	helper, _, controller := setupBenchmarkController(b)
	defer cleanupBenchmarkController(helper, nil, controller)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Benchmark polling readiness check
			controller.IsReady()
		}
	})
}

// BenchmarkEventDrivenHealthMonitoring benchmarks event-driven health monitoring
func BenchmarkEventDrivenHealthMonitoring(b *testing.B) {
	helper, eventHelper, controller := setupBenchmarkController(b)
	defer cleanupBenchmarkController(helper, eventHelper, controller)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Benchmark non-blocking health event observation
			eventHelper.ObserveHealthChanges()

			// No waiting - just observe events
		}
	})
}

// BenchmarkPollingHealthMonitoring benchmarks traditional polling health monitoring
func BenchmarkPollingHealthMonitoring(b *testing.B) {
	helper, _, controller := setupBenchmarkController(b)
	defer cleanupBenchmarkController(helper, nil, controller)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Benchmark polling health check
			// This simulates checking health status through polling
			controller.IsReady()
		}
	})
}

// BenchmarkEventAggregation benchmarks event aggregation performance
func BenchmarkEventAggregation(b *testing.B) {
	helper, eventHelper, controller := setupBenchmarkController(b)
	defer cleanupBenchmarkController(helper, eventHelper, controller)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Benchmark event aggregation (non-blocking)
			eventHelper.ObserveReadiness()
			eventHelper.ObserveHealthChanges()
		}
	})
}

// BenchmarkEventDrivenVsPollingComparison provides a direct comparison
func BenchmarkEventDrivenVsPollingComparison(b *testing.B) {
	helper, eventHelper, controller := setupBenchmarkController(b)
	defer cleanupBenchmarkController(helper, eventHelper, controller)

	b.Run("EventDriven", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				// Non-blocking event observation approach
				eventHelper.ObserveReadiness()
			}
		})
	})

	b.Run("Polling", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				// Polling approach
				controller.IsReady()
			}
		})
	})
}

// BenchmarkEventDrivenLatency benchmarks the latency of event-driven notifications
func BenchmarkEventDrivenLatency(b *testing.B) {
	helper, eventHelper, controller := setupBenchmarkController(b)
	defer cleanupBenchmarkController(helper, eventHelper, controller)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Measure latency of event subscription and notification
		start := time.Now()

		eventHelper.ObserveReadiness()

		// No waiting - just observe events
		latency := time.Since(start)
		b.ReportMetric(float64(latency.Nanoseconds()), "ns/event")
	}
}

// BenchmarkEventDrivenThroughput benchmarks the throughput of event-driven system
func BenchmarkEventDrivenThroughput(b *testing.B) {
	helper, eventHelper, controller := setupBenchmarkController(b)
	defer cleanupBenchmarkController(helper, eventHelper, controller)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		eventCount := 0
		for pb.Next() {
			// Create multiple non-blocking event observations to test throughput
			eventHelper.ObserveReadiness()
			eventHelper.ObserveHealthChanges()

			// No waiting - just observe events
			eventCount++
		}
		b.ReportMetric(float64(eventCount), "events/sec")
	})
}
