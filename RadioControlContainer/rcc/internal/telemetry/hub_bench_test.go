// Package telemetry provides performance benchmarks for the telemetry hub.
package telemetry

import (
	"context"
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/radio-control/rcc/internal/config"
)

func BenchmarkPublishWithSubscribers(b *testing.B) {
	cfg := config.LoadCBTimingBaseline()
	hub := NewHub(cfg)
	defer hub.Stop()

	// Test with different numbers of subscribers
	subscriberCounts := []int{1, 10, 100}

	for _, count := range subscriberCounts {
		b.Run(fmt.Sprintf("Subscribers_%d", count), func(b *testing.B) {
			// Create subscribers
			subscribers := make([]*Client, count)
			for i := 0; i < count; i++ {
				req := httptest.NewRequest("GET", "/telemetry", nil)
				req.Header.Set("Accept", "text/event-stream")
				w := httptest.NewRecorder()

				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				client := &Client{
					ID:      fmt.Sprintf("client-%d", i),
					Context: ctx,
					Events:  make(chan Event, 100),
				}

				subscribers[i] = client
				hub.Subscribe(ctx, w, req)
			}

			b.ResetTimer()

			// Run b.N iterations of Publish
			for i := 0; i < b.N; i++ {
				event := Event{
					ID:    int64(i),
					Radio: "silvus-001",
					Type:  "powerChanged",
					Data:  map[string]interface{}{"powerDbm": 10.0 + float64(i%10)},
				}

				err := hub.Publish(event)
				if err != nil {
					b.Fatalf("Publish failed: %v", err)
				}
			}
		})
	}
}

func BenchmarkPublishWithoutSubscribers(b *testing.B) {
	cfg := config.LoadCBTimingBaseline()
	hub := NewHub(cfg)
	defer hub.Stop()

	b.ResetTimer()

	// Run b.N iterations of Publish without subscribers
	for i := 0; i < b.N; i++ {
		event := Event{
			ID:    int64(i),
			Radio: "silvus-001",
			Type:  "powerChanged",
			Data:  map[string]interface{}{"powerDbm": 10.0 + float64(i%10)},
		}

		err := hub.Publish(event)
		if err != nil {
			b.Fatalf("Publish failed: %v", err)
		}
	}
}

func BenchmarkSubscribe(b *testing.B) {
	cfg := config.LoadCBTimingBaseline()
	hub := NewHub(cfg)
	defer hub.Stop()

	b.ResetTimer()

	// Run b.N iterations of Subscribe
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/telemetry", nil)
		req.Header.Set("Accept", "text/event-stream")

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		w := httptest.NewRecorder()
		hub.Subscribe(ctx, w, req)
	}
}

func BenchmarkEventIDGeneration(b *testing.B) {
	cfg := config.LoadCBTimingBaseline()
	hub := NewHub(cfg)
	defer hub.Stop()

	b.ResetTimer()

	// Run b.N iterations of getNextEventID
	for i := 0; i < b.N; i++ {
		radioID := fmt.Sprintf("radio-%d", i%10) // Cycle through 10 radios
		hub.getNextEventID(radioID)
	}
}

func BenchmarkBufferEvent(b *testing.B) {
	cfg := config.LoadCBTimingBaseline()
	hub := NewHub(cfg)
	defer hub.Stop()

	b.ResetTimer()

	// Run b.N iterations of bufferEvent
	for i := 0; i < b.N; i++ {
		event := Event{
			ID:    int64(i),
			Radio: fmt.Sprintf("radio-%d", i%10),
			Type:  "powerChanged",
			Data:  map[string]interface{}{"powerDbm": 10.0 + float64(i%10)},
		}

		hub.bufferEvent(event)
	}
}

func BenchmarkHubConcurrent(b *testing.B) {
	cfg := config.LoadCBTimingBaseline()
	hub := NewHub(cfg)
	defer hub.Stop()

	b.ResetTimer()

	// Run b.N iterations with concurrent operations
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Mix of operations
			switch b.N % 3 {
			case 0:
				// Publish event
				event := Event{
					ID:    int64(b.N),
					Radio: "silvus-001",
					Type:  "powerChanged",
					Data:  map[string]interface{}{"powerDbm": 10.0},
				}
				hub.Publish(event)
			case 1:
				// Generate event ID
				hub.getNextEventID("silvus-001")
			case 2:
				// Buffer event
				event := Event{
					ID:    int64(b.N),
					Radio: "silvus-001",
					Type:  "powerChanged",
					Data:  map[string]interface{}{"powerDbm": 10.0},
				}
				hub.bufferEvent(event)
			}
		}
	})
}

func BenchmarkHeartbeat(b *testing.B) {
	cfg := config.LoadCBTimingBaseline()
	// Use shorter heartbeat interval for testing
	cfg.HeartbeatInterval = 10 * time.Millisecond
	cfg.HeartbeatJitter = 1 * time.Millisecond

	hub := NewHub(cfg)
	defer hub.Stop()

	// Start heartbeat
	hub.startHeartbeat()

	b.ResetTimer()

	// Run b.N iterations of sendHeartbeat
	for i := 0; i < b.N; i++ {
		hub.sendHeartbeat()
	}
}
