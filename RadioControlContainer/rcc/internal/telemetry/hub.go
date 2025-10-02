// Package telemetry implements TelemetryHub from Architecture §5.
//
// Requirements:
//   - Architecture §5: "Fan-out events to all SSE clients; buffer last N events per client for reconnection (Last-Event-ID)."
//
// Source: Telemetry SSE v1
// Quote: "Event IDs are monotonic per radio, starting from 1"
package telemetry

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/radio-control/rcc/internal/config"
)

// Event represents a telemetry event with SSE formatting.
// Source: Telemetry SSE v1 §2.1
type Event struct {
	ID    int64                  `json:"id,omitempty"`
	Type  string                 `json:"type"`
	Data  map[string]interface{} `json:"data"`
	Radio string                 `json:"radio,omitempty"`
}

// Client represents an SSE client connection.
type Client struct {
	ID      string
	Writer  http.ResponseWriter
	Request *http.Request
	Context context.Context
	Cancel  context.CancelFunc
	LastID  int64
	Radio   string
	Events  chan Event
}

// Hub manages SSE telemetry distribution with per-radio buffering.
// Source: Architecture §5
// Quote: "Fan-out events to all SSE clients; buffer last N events per client for reconnection"
type Hub struct {
	mu       sync.RWMutex
	clients  map[string]*Client
	radioIDs map[string]int64 // Monotonic event IDs per radio

	// Per-radio event buffers
	buffers map[string]*EventBuffer

	// Configuration
	config *config.TimingConfig

	// Heartbeat ticker
	heartbeatTicker *time.Ticker
	stopHeartbeat   chan bool
}

// EventBuffer maintains a circular buffer of events for a specific radio.
// Source: CB-TIMING v0.3 §6.1
type EventBuffer struct {
	mu       sync.RWMutex
	events   []Event
	capacity int
	nextID   int64
	created  time.Time
}

// NewHub creates a new telemetry hub with the specified configuration.
// Source: CB-TIMING v0.3
// Quote: "Buffer size per radio: 50 events, Buffer retention: 1 hour"
func NewHub(timingConfig *config.TimingConfig) *Hub {
	hub := &Hub{
		clients:  make(map[string]*Client),
		radioIDs: make(map[string]int64),
		buffers:  make(map[string]*EventBuffer),
		config:   timingConfig,
	}

	return hub
}

// Subscribe handles SSE client subscription with Last-Event-ID resume support.
// Source: Telemetry SSE v1 §1.3
// Quote: "Clients should send Last-Event-ID on reconnect to resume from the last processed event ID"
func (h *Hub) Subscribe(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Cache-Control")

	// Create client context
	clientCtx, cancel := context.WithCancel(ctx)

	// Generate client ID
	clientID := fmt.Sprintf("client_%d", time.Now().UnixNano())

	// Parse Last-Event-ID header for resume
	lastEventID := int64(0)
	if lastIDStr := r.Header.Get("Last-Event-ID"); lastIDStr != "" {
		if id, err := strconv.ParseInt(lastIDStr, 10, 64); err == nil {
			lastEventID = id
		}
	}

	// Create client
	client := &Client{
		ID:      clientID,
		Writer:  w,
		Request: r,
		Context: clientCtx,
		Cancel:  cancel,
		LastID:  lastEventID,
		Events:  make(chan Event, 100), // Buffer for client events
	}

	// Register client
	h.mu.Lock()
	h.clients[clientID] = client
	h.mu.Unlock()

	// Send initial ready event
	if err := h.sendReadyEvent(client); err != nil {
		h.unregisterClient(clientID)
		return fmt.Errorf("failed to send ready event: %w", err)
	}

	// Replay buffered events if Last-Event-ID provided
	if lastEventID > 0 {
		if err := h.replayEvents(client, lastEventID); err != nil {
			h.unregisterClient(clientID)
			return fmt.Errorf("failed to replay events: %w", err)
		}
	}

	// Start heartbeat if this is the first client
	h.mu.Lock()
	if len(h.clients) == 0 {
		h.startHeartbeat()
	}
	h.mu.Unlock()

	// Handle client events in a separate goroutine
	go h.handleClient(client)

	return nil
}

// Publish publishes an event to all connected clients.
// Source: Telemetry SSE v1 §2.2
func (h *Hub) Publish(event Event) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Assign event ID if not set
	if event.ID == 0 {
		event.ID = h.getNextEventID(event.Radio)
	}

	// Buffer the event
	if event.Radio != "" {
		h.bufferEvent(event)
	}

	// Send to all clients
	for _, client := range h.clients {
		select {
		case client.Events <- event:
		default:
			// Client buffer full, skip this event
		}
	}

	return nil
}

// PublishRadio publishes an event for a specific radio.
// Source: Telemetry SSE v1 §2.2
func (h *Hub) PublishRadio(radioID string, event Event) error {
	event.Radio = radioID
	return h.Publish(event)
}

// sendReadyEvent sends the initial ready event to a client.
// Source: Telemetry SSE v1 §2.2a
func (h *Hub) sendReadyEvent(client *Client) error {
	readyEvent := Event{
		Type: "ready",
		Data: map[string]interface{}{
			"snapshot": map[string]interface{}{
				"activeRadioId": "",              // TODO: Get from radio manager
				"radios":        []interface{}{}, // TODO: Get from radio manager
			},
		},
	}

	return h.sendEventToClient(client, readyEvent)
}

// replayEvents replays buffered events for a client based on Last-Event-ID.
// Source: Telemetry SSE v1 §1.3
func (h *Hub) replayEvents(client *Client, lastEventID int64) error {
	h.mu.RLock()
	buffer, exists := h.buffers[client.Radio]
	h.mu.RUnlock()

	if !exists {
		return nil // No buffer for this radio
	}

	// Get events after the last event ID
	events := buffer.GetEventsAfter(lastEventID)

	// Send replayed events
	for _, event := range events {
		if err := h.sendEventToClient(client, event); err != nil {
			return err
		}
	}

	return nil
}

// sendEventToClient sends a single event to a client via SSE.
func (h *Hub) sendEventToClient(client *Client, event Event) error {
	// Format as SSE
	if event.ID > 0 {
		fmt.Fprintf(client.Writer, "id: %d\n", event.ID)
	}
	fmt.Fprintf(client.Writer, "event: %s\n", event.Type)

	// Serialize data as JSON
	data, err := json.Marshal(event.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	fmt.Fprintf(client.Writer, "data: %s\n\n", string(data))

	// Flush the response
	if flusher, ok := client.Writer.(http.Flusher); ok {
		flusher.Flush()
	}

	return nil
}

// handleClient manages a client connection and event delivery.
func (h *Hub) handleClient(client *Client) {
	defer h.unregisterClient(client.ID)

	for {
		select {
		case <-client.Context.Done():
			return
		case event, ok := <-client.Events:
			if !ok {
				// Channel closed
				return
			}
			if err := h.sendEventToClient(client, event); err != nil {
				return
			}
		}
	}
}

// unregisterClient removes a client from the hub.
func (h *Hub) unregisterClient(clientID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if client, exists := h.clients[clientID]; exists {
		client.Cancel()
		close(client.Events)
		delete(h.clients, clientID)

		// Stop heartbeat if no clients remain
		if len(h.clients) == 0 && h.heartbeatTicker != nil {
			h.heartbeatTicker.Stop()
			h.heartbeatTicker = nil
			if h.stopHeartbeat != nil {
				close(h.stopHeartbeat)
				h.stopHeartbeat = nil
			}
		}
	}
}

// getNextEventID returns the next monotonic event ID for a radio.
// Source: Telemetry SSE v1 §1.3
func (h *Hub) getNextEventID(radioID string) int64 {
	h.mu.Lock()
	defer h.mu.Unlock()

	if radioID == "" {
		radioID = "global"
	}

	h.radioIDs[radioID]++
	return h.radioIDs[radioID]
}

// bufferEvent adds an event to the per-radio buffer.
// Source: CB-TIMING v0.3 §6.1
func (h *Hub) bufferEvent(event Event) {
	if event.Radio == "" {
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	buffer, exists := h.buffers[event.Radio]
	if !exists {
		buffer = NewEventBuffer(h.config.EventBufferSize)
		h.buffers[event.Radio] = buffer
	}

	buffer.AddEvent(event)
}

// startHeartbeat starts the heartbeat ticker.
// Source: CB-TIMING v0.3 §3.1
func (h *Hub) startHeartbeat() {
	// Don't start if already running
	if h.heartbeatTicker != nil {
		return
	}

	interval := h.config.HeartbeatInterval
	jitter := h.config.HeartbeatJitter

	// Add jitter to prevent thundering herd
	actualInterval := interval + time.Duration(float64(jitter)*0.5)

	h.heartbeatTicker = time.NewTicker(actualInterval)
	h.stopHeartbeat = make(chan bool)

	go func() {
		defer func() {
			if h.heartbeatTicker != nil {
				h.heartbeatTicker.Stop()
			}
		}()

		for {
			select {
			case <-h.heartbeatTicker.C:
				h.sendHeartbeat()
			case <-h.stopHeartbeat:
				return
			}
		}
	}()
}

// sendHeartbeat sends a heartbeat event to all clients.
// Source: Telemetry SSE v1 §2.2f
func (h *Hub) sendHeartbeat() {
	heartbeatEvent := Event{
		Type: "heartbeat",
		Data: map[string]interface{}{
			"ts": time.Now().UTC().Format(time.RFC3339),
		},
	}

	h.Publish(heartbeatEvent)
}

// Stop stops the telemetry hub and cleans up resources.
func (h *Hub) Stop() {
	if h.heartbeatTicker != nil {
		h.heartbeatTicker.Stop()
	}

	if h.stopHeartbeat != nil {
		close(h.stopHeartbeat)
	}

	// Close all client connections
	h.mu.Lock()
	for _, client := range h.clients {
		client.Cancel()
		close(client.Events)
	}
	h.clients = make(map[string]*Client)
	h.mu.Unlock()
}

// NewEventBuffer creates a new event buffer with the specified capacity.
func NewEventBuffer(capacity int) *EventBuffer {
	return &EventBuffer{
		events:   make([]Event, 0, capacity),
		capacity: capacity,
		nextID:   1,
		created:  time.Now(),
	}
}

// AddEvent adds an event to the buffer.
func (b *EventBuffer) AddEvent(event Event) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Assign ID if not set
	if event.ID == 0 {
		event.ID = b.nextID
		b.nextID++
	}

	// Add to buffer
	b.events = append(b.events, event)

	// Maintain capacity
	if len(b.events) > b.capacity {
		b.events = b.events[1:]
	}
}

// GetEventsAfter returns events after the specified ID.
func (b *EventBuffer) GetEventsAfter(lastID int64) []Event {
	b.mu.RLock()
	defer b.mu.RUnlock()

	var result []Event
	for _, event := range b.events {
		if event.ID > lastID {
			result = append(result, event)
		}
	}

	return result
}

// GetCapacity returns the buffer capacity.
func (b *EventBuffer) GetCapacity() int {
	return b.capacity
}

// GetSize returns the current buffer size.
func (b *EventBuffer) GetSize() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.events)
}
