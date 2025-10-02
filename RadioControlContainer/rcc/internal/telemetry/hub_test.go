package telemetry

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/radio-control/rcc/internal/config"
)

func TestNewHub(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	hub := NewHub(cfg)

	if hub == nil {
		t.Fatal("NewHub() returned nil")
	}

	if hub.clients == nil {
		t.Error("Hub clients map not initialized")
	}

	if hub.radioIDs == nil {
		t.Error("Hub radioIDs map not initialized")
	}

	if hub.buffers == nil {
		t.Error("Hub buffers map not initialized")
	}

	if hub.config != cfg {
		t.Error("Hub config not set correctly")
	}

	// Clean up
	hub.Stop()
}

func TestHubPublish(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	hub := NewHub(cfg)
	defer hub.Stop()

	// Publish an event without clients
	event := Event{
		Type: "test",
		Data: map[string]interface{}{
			"message": "test event",
		},
	}

	err := hub.Publish(event)
	if err != nil {
		t.Fatalf("Publish() failed: %v", err)
	}
}

func TestHubPublishRadio(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	hub := NewHub(cfg)
	defer hub.Stop()

	// Publish an event for a specific radio
	event := Event{
		Type: "state",
		Data: map[string]interface{}{
			"powerDbm":     30,
			"frequencyMhz": 2412,
		},
	}

	err := hub.PublishRadio("radio-01", event)
	if err != nil {
		t.Fatalf("PublishRadio() failed: %v", err)
	}

	// Check that event was buffered for the radio
	hub.mu.RLock()
	buffer, exists := hub.buffers["radio-01"]
	hub.mu.RUnlock()

	if !exists {
		t.Error("Event buffer not created for radio")
	}

	if buffer != nil && buffer.GetSize() != 1 {
		t.Errorf("Expected 1 event in buffer, got %d", buffer.GetSize())
	}
}

func TestEventBuffer(t *testing.T) {
	capacity := 5
	buffer := NewEventBuffer(capacity)

	if buffer.GetCapacity() != capacity {
		t.Errorf("Expected capacity %d, got %d", capacity, buffer.GetCapacity())
	}

	if buffer.GetSize() != 0 {
		t.Errorf("Expected initial size 0, got %d", buffer.GetSize())
	}

	// Add events
	for i := 0; i < 7; i++ { // More than capacity
		event := Event{
			Type: "test",
			Data: map[string]interface{}{
				"index": i,
			},
		}
		buffer.AddEvent(event)
	}

	// Should maintain capacity
	if buffer.GetSize() != capacity {
		t.Errorf("Expected size %d, got %d", capacity, buffer.GetSize())
	}

	// Test GetEventsAfter
	events := buffer.GetEventsAfter(2)
	if len(events) != 3 { // Events 3, 4, 5 (0-based indexing)
		t.Errorf("Expected 3 events after ID 2, got %d", len(events))
	}
}

func TestHubStop(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	hub := NewHub(cfg)

	// Stop the hub
	hub.Stop()

	// Check that clients are cleaned up
	hub.mu.RLock()
	clientCount := len(hub.clients)
	hub.mu.RUnlock()

	if clientCount != 0 {
		t.Errorf("Expected 0 clients after stop, got %d", clientCount)
	}
}

func TestEventTypes(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	hub := NewHub(cfg)
	defer hub.Stop()

	// Test different event types
	eventTypes := []string{"ready", "state", "channelChanged", "powerChanged", "fault", "heartbeat"}

	for _, eventType := range eventTypes {
		event := Event{
			Type: eventType,
			Data: map[string]interface{}{
				"test": "data",
			},
		}

		err := hub.Publish(event)
		if err != nil {
			t.Errorf("Publish() failed for event type %s: %v", eventType, err)
		}
	}
}

func TestEventIDGeneration(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	hub := NewHub(cfg)
	defer hub.Stop()

	// Test global event ID generation
	event1 := Event{Type: "test1", Data: map[string]interface{}{}}
	event2 := Event{Type: "test2", Data: map[string]interface{}{}}

	hub.Publish(event1)
	hub.Publish(event2)

	// Test radio-specific event ID generation
	radioEvent1 := Event{Type: "state", Data: map[string]interface{}{}, Radio: "radio-01"}
	radioEvent2 := Event{Type: "state", Data: map[string]interface{}{}, Radio: "radio-01"}

	hub.PublishRadio("radio-01", radioEvent1)
	hub.PublishRadio("radio-01", radioEvent2)

	// Check that radio buffer was created
	hub.mu.RLock()
	buffer, exists := hub.buffers["radio-01"]
	hub.mu.RUnlock()

	if !exists {
		t.Error("Radio buffer not created")
	}

	if buffer.GetSize() != 2 {
		t.Errorf("Expected 2 events in radio buffer, got %d", buffer.GetSize())
	}
}

func TestEventCreation(t *testing.T) {
	// Test event creation
	event := Event{
		ID:   42,
		Type: "test",
		Data: map[string]interface{}{
			"message": "test event",
		},
	}

	// Test that event has correct format
	if event.ID != 42 {
		t.Error("Event ID not set correctly")
	}

	if event.Type != "test" {
		t.Error("Event type not set correctly")
	}

	if event.Data["message"] != "test event" {
		t.Error("Event data not set correctly")
	}
}

func TestConcurrentPublish(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	hub := NewHub(cfg)
	defer hub.Stop()

	// Publish events concurrently without clients
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func(index int) {
			event := Event{
				Type: "concurrent",
				Data: map[string]interface{}{
					"index": index,
				},
			}
			hub.Publish(event)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestHubSubscribeBasic(t *testing.T) {
	cfg := config.LoadCBTimingBaseline()
	hub := NewHub(cfg)
	defer hub.Stop()

	// Create test request
	req := httptest.NewRequest("GET", "/telemetry", nil)
	req.Header.Set("Accept", "text/event-stream")

	// Create test response recorder
	w := httptest.NewRecorder()

	// Subscribe
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := hub.Subscribe(ctx, w, req)
	if err != nil {
		t.Fatalf("Subscribe() failed: %v", err)
	}

	// Check that client was registered
	hub.mu.RLock()
	clientCount := len(hub.clients)
	hub.mu.RUnlock()

	if clientCount != 1 {
		t.Errorf("Expected 1 client, got %d", clientCount)
	}

	// Check response headers
	if w.Header().Get("Content-Type") != "text/event-stream; charset=utf-8" {
		t.Error("Content-Type header not set correctly")
	}

	if w.Header().Get("Cache-Control") != "no-cache" {
		t.Error("Cache-Control header not set correctly")
	}

	// Wait for context to timeout and client to be cleaned up
	time.Sleep(150 * time.Millisecond)

	// Check that client was cleaned up
	hub.mu.RLock()
	clientCount = len(hub.clients)
	hub.mu.RUnlock()

	if clientCount != 0 {
		t.Errorf("Expected 0 clients after timeout, got %d", clientCount)
	}
}
