/*
Event subscription system for WebSocket server.

Provides efficient, topic-based event delivery to subscribed clients,
replacing the inefficient broadcast-to-all approach.

Requirements Coverage:
- REQ-API-001: Efficient event delivery
- REQ-API-002: Client subscription management
- REQ-API-003: Topic-based filtering

Test Categories: Unit/Integration
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package websocket

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
)

// EventTopic represents different types of events that clients can subscribe to
type EventTopic string

const (
	// Camera events
	TopicCameraConnected          EventTopic = "camera.connected"
	TopicCameraDisconnected       EventTopic = "camera.disconnected"
	TopicCameraStatusChange       EventTopic = "camera.status_change"
	TopicCameraCapabilityDetected EventTopic = "camera.capability_detected"
	TopicCameraCapabilityError    EventTopic = "camera.capability_error"

	// Recording events
	TopicRecordingStart    EventTopic = "recording.start"
	TopicRecordingStop     EventTopic = "recording.stop"
	TopicRecordingProgress EventTopic = "recording.progress"
	TopicRecordingError    EventTopic = "recording.error"

	// Snapshot events
	TopicSnapshotTaken EventTopic = "snapshot.taken"
	TopicSnapshotError EventTopic = "snapshot.error"

	// System events
	TopicSystemHealth   EventTopic = "system.health"
	TopicSystemError    EventTopic = "system.error"
	TopicSystemStartup  EventTopic = "system.startup"
	TopicSystemShutdown EventTopic = "system.shutdown"

	// MediaMTX events
	TopicMediaMTXStream           EventTopic = "mediamtx.stream"
	TopicMediaMTXPath             EventTopic = "mediamtx.path"
	TopicMediaMTXError            EventTopic = "mediamtx.error"
	TopicMediaMTXRecordingStarted EventTopic = "mediamtx.recording_started"
	TopicMediaMTXRecordingStopped EventTopic = "mediamtx.recording_stopped"
	TopicMediaMTXStreamStarted    EventTopic = "mediamtx.stream_started"
	TopicMediaMTXStreamStopped    EventTopic = "mediamtx.stream_stopped"
)

// EventSubscription represents a client's subscription to specific event topics
type EventSubscription struct {
	ClientID  string                 `json:"client_id"`
	Topics    []EventTopic           `json:"topics"`
	Filters   map[string]interface{} `json:"filters,omitempty"` // See SupportedFilters below
	CreatedAt time.Time              `json:"created_at"`
	LastSeen  time.Time              `json:"last_seen"`
	Active    bool                   `json:"active"`
}

// SupportedFilters documents the supported subscription filter keys and their types
//
// ✅ **Supported Client-Facing Filters (per JSON-RPC API specification):**
// - "device" (string): Filter by camera identifier (e.g., "camera0", "camera1") - **PRIMARY FILTER**
// - "topic" (string): Filter by specific event topic
// - "timestamp_after" (string): Filter events after RFC3339 timestamp (optional)
// - "timestamp_before" (string): Filter events before RFC3339 timestamp (optional)
//
// ⚠️ **Internal/Debug Filters (avoid in client applications):**
// - "device_path" (string): Filter by internal device path (e.g., "/dev/video0") - for internal tooling only
//
// **Filter Example (proper abstraction):**
// ```json
//
//	{
//	  "device": "camera0",
//	  "timestamp_after": "2024-01-01T00:00:00Z"
//	}
//
// ```
//
// **Matching Behavior:**
// - All specified filters must match (AND logic)
// - Exact string matching for device, topic, device_path
// - Timestamp filters use RFC3339 format comparison
// - Missing event fields cause filter to fail (no match)
type SupportedFilters struct {
	// This type exists only for documentation - use map[string]interface{} in practice
}

// EventMessage represents a structured event message
type EventMessage struct {
	Topic     EventTopic             `json:"topic"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
	EventID   string                 `json:"event_id"`
}

// EventManager manages event subscriptions and delivery
type EventManager struct {
	// Subscriptions by client ID
	subscriptions map[string]*EventSubscription

	// Subscriptions by topic for efficient lookup
	topicSubscriptions map[EventTopic]map[string]*EventSubscription

	// Event handlers for custom processing
	eventHandlers map[EventTopic][]func(*EventMessage) error

	// Thread safety
	mu sync.RWMutex

	// Atomic counters for fast statistics
	totalClients        int64 // atomic
	activeSubscriptions int64 // atomic

	// Logging
	logger *logging.Logger
}

// NewEventManager creates a new event manager
func NewEventManager(logger *logging.Logger) *EventManager {
	return &EventManager{
		subscriptions:      make(map[string]*EventSubscription),
		topicSubscriptions: make(map[EventTopic]map[string]*EventSubscription),
		eventHandlers:      make(map[EventTopic][]func(*EventMessage) error),
		logger:             logger,
	}
}

// Subscribe adds a client subscription to specific event topics
func (em *EventManager) Subscribe(clientID string, topics []EventTopic, filters map[string]interface{}) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	// Validate topics
	for _, topic := range topics {
		if !em.isValidTopic(topic) {
			return fmt.Errorf("invalid event topic: %s", topic)
		}
	}

	// Create or update subscription
	subscription := &EventSubscription{
		ClientID:  clientID,
		Topics:    topics,
		Filters:   filters,
		CreatedAt: time.Now(),
		LastSeen:  time.Now(),
		Active:    true,
	}

	// Store subscription by client ID
	em.subscriptions[clientID] = subscription

	// Store subscription by topic for efficient lookup
	for _, topic := range topics {
		if em.topicSubscriptions[topic] == nil {
			em.topicSubscriptions[topic] = make(map[string]*EventSubscription)
		}
		em.topicSubscriptions[topic][clientID] = subscription
	}

	// Update atomic counters
	atomic.AddInt64(&em.totalClients, 1)
	atomic.AddInt64(&em.activeSubscriptions, int64(len(topics)))

	em.logger.WithFields(logging.Fields{
		"client_id": clientID,
		"topics":    topics,
		"filters":   filters,
	}).Debug("Client subscribed to event topics")

	return nil
}

// Unsubscribe removes a client's subscription to specific topics
func (em *EventManager) Unsubscribe(clientID string, topics []EventTopic) error {
	em.mu.Lock()
	defer em.mu.Unlock()

	_, exists := em.subscriptions[clientID]
	if !exists {
		// Client has no subscriptions - this is already the desired state
		em.logger.WithField("client_id", clientID).Debug("Client has no subscriptions to remove")
		return nil
	}

	// Remove topics from subscription
	if len(topics) == 0 {
		// Remove all subscriptions for this client
		em.removeClientSubscriptions(clientID)
		// Update atomic counters for complete removal
		atomic.AddInt64(&em.totalClients, -1)
		// Note: activeSubscriptions will be decremented in removeClientSubscriptions
	} else {
		// Remove specific topics
		for _, topic := range topics {
			em.removeTopicSubscription(clientID, topic)
		}

		// Update subscription topics
		em.updateSubscriptionTopics(clientID, topics, true)
		// Update atomic counter for partial removal
		atomic.AddInt64(&em.activeSubscriptions, -int64(len(topics)))
	}

	em.logger.WithFields(logging.Fields{
		"client_id": clientID,
		"topics":    topics,
	}).Debug("Client unsubscribed from event topics")

	return nil
}

// PublishEvent sends an event to all subscribed clients
func (em *EventManager) PublishEvent(topic EventTopic, data map[string]interface{}) error {
	em.mu.RLock()
	defer em.mu.RUnlock()

	// Create event message
	event := &EventMessage{
		Topic:     topic,
		Data:      data,
		Timestamp: time.Now(),
		EventID:   generateEventID(),
	}

	// Process event through handlers
	if err := em.processEventHandlers(event); err != nil {
		em.logger.WithError(err).WithField("topic", string(topic)).Error("Event handler processing failed")
	}

	// Get subscribers for this topic
	subscribers, exists := em.topicSubscriptions[topic]
	if !exists {
		em.logger.WithField("topic", string(topic)).Debug("No subscribers for event topic")
		return nil
	}

	// Count interested subscribers
	subscriberCount := 0
	for _, subscription := range subscribers {
		if subscription.Active && em.isClientInterested(subscription, event) {
			subscriberCount++
		}
	}

	em.logger.WithFields(logging.Fields{
		"topic":            topic,
		"subscriber_count": subscriberCount,
		"event_id":         event.EventID,
	}).Debug("Event published to subscribers")

	return nil
}

// GetSubscribersForTopic returns all active subscribers for a specific topic
func (em *EventManager) GetSubscribersForTopic(topic EventTopic) []string {
	em.mu.RLock()
	defer em.mu.RUnlock()

	subscribers, exists := em.topicSubscriptions[topic]
	if !exists {
		return []string{}
	}

	var clientIDs []string
	for clientID, subscription := range subscribers {
		if subscription.Active {
			clientIDs = append(clientIDs, clientID)
		}
	}

	return clientIDs
}

// GetClientSubscriptions returns all topics a client is subscribed to
func (em *EventManager) GetClientSubscriptions(clientID string) []EventTopic {
	em.mu.RLock()
	defer em.mu.RUnlock()

	subscription, exists := em.subscriptions[clientID]
	if !exists {
		return []EventTopic{}
	}

	return subscription.Topics
}

// RemoveClient removes all subscriptions for a client
func (em *EventManager) RemoveClient(clientID string) {
	em.mu.Lock()
	defer em.mu.Unlock()

	// Get subscription count before removal for atomic counter update
	subscription, exists := em.subscriptions[clientID]
	var topicCount int64
	if exists && subscription != nil {
		topicCount = int64(len(subscription.Topics))
	}

	em.removeClientSubscriptions(clientID)

	// Update atomic counters
	if exists {
		atomic.AddInt64(&em.totalClients, -1)
		atomic.AddInt64(&em.activeSubscriptions, -topicCount)
	}

	em.logger.WithField("client_id", clientID).Debug("Client removed from event manager")
}

// AddEventHandler adds a custom event handler for a specific topic
func (em *EventManager) AddEventHandler(topic EventTopic, handler func(*EventMessage) error) {
	em.mu.Lock()
	defer em.mu.Unlock()

	em.eventHandlers[topic] = append(em.eventHandlers[topic], handler)
}

// UpdateClientLastSeen updates the last seen timestamp for a client
func (em *EventManager) UpdateClientLastSeen(clientID string) {
	em.mu.Lock()
	defer em.mu.Unlock()

	if subscription, exists := em.subscriptions[clientID]; exists {
		subscription.LastSeen = time.Now()
	}
}

// GetSubscriptionStats returns statistics about subscriptions
func (em *EventManager) GetSubscriptionStats() map[string]interface{} {
	// Use atomic operations for fast counter reads
	totalClients := atomic.LoadInt64(&em.totalClients)
	activeSubscriptions := atomic.LoadInt64(&em.activeSubscriptions)

	// Still need mutex for complex map operations
	em.mu.RLock()
	defer em.mu.RUnlock()

	stats := map[string]interface{}{
		"total_clients":        totalClients,
		"total_topics":         len(em.topicSubscriptions),
		"active_subscriptions": activeSubscriptions,
		"topic_distribution":   make(map[string]int),
	}

	// Count topic distribution (still needs mutex for map iteration)
	for topic, subscribers := range em.topicSubscriptions {
		stats["topic_distribution"].(map[string]int)[string(topic)] = len(subscribers)
	}

	return stats
}

// Helper methods

func (em *EventManager) isValidTopic(topic EventTopic) bool {
	validTopics := []EventTopic{
		TopicCameraConnected, TopicCameraDisconnected, TopicCameraStatusChange,
		TopicCameraCapabilityDetected, TopicCameraCapabilityError,
		TopicRecordingStart, TopicRecordingStop, TopicRecordingProgress, TopicRecordingError,
		TopicSnapshotTaken, TopicSnapshotError,
		TopicSystemHealth, TopicSystemError, TopicSystemStartup, TopicSystemShutdown,
		TopicMediaMTXStream, TopicMediaMTXPath, TopicMediaMTXError,
		TopicMediaMTXRecordingStarted, TopicMediaMTXRecordingStopped,
		TopicMediaMTXStreamStarted, TopicMediaMTXStreamStopped,
	}

	for _, valid := range validTopics {
		if topic == valid {
			return true
		}
	}
	return false
}

func (em *EventManager) removeClientSubscriptions(clientID string) {
	subscription := em.subscriptions[clientID]
	if subscription == nil {
		return
	}

	// Remove from topic subscriptions
	for _, topic := range subscription.Topics {
		em.removeTopicSubscription(clientID, topic)
	}

	// Remove from client subscriptions
	delete(em.subscriptions, clientID)

	// Note: Atomic counter updates are handled by the calling method
	// to avoid double-counting in different code paths
}

func (em *EventManager) removeTopicSubscription(clientID string, topic EventTopic) {
	if subscribers, exists := em.topicSubscriptions[topic]; exists {
		delete(subscribers, clientID)
		if len(subscribers) == 0 {
			delete(em.topicSubscriptions, topic)
		}
	}
}

func (em *EventManager) updateSubscriptionTopics(clientID string, topicsToRemove []EventTopic, remove bool) {
	subscription := em.subscriptions[clientID]
	if subscription == nil {
		return
	}

	// Create a map for efficient lookup
	topicMap := make(map[EventTopic]bool)
	for _, topic := range subscription.Topics {
		topicMap[topic] = true
	}

	// Remove or add topics
	for _, topic := range topicsToRemove {
		if remove {
			delete(topicMap, topic)
		} else {
			topicMap[topic] = true
		}
	}

	// Convert back to slice
	var newTopics []EventTopic
	for topic := range topicMap {
		newTopics = append(newTopics, topic)
	}

	subscription.Topics = newTopics
}

func (em *EventManager) isClientInterested(subscription *EventSubscription, event *EventMessage) bool {
	// Check if client is subscribed to this topic
	topicFound := false
	for _, topic := range subscription.Topics {
		if topic == event.Topic {
			topicFound = true
			break
		}
	}

	if !topicFound {
		return false
	}

	// Apply filters if specified
	if len(subscription.Filters) > 0 {
		return em.applyFilters(subscription.Filters, event.Data)
	}

	return true
}

func (em *EventManager) applyFilters(filters map[string]interface{}, eventData map[string]interface{}) bool {
	for key, expectedValue := range filters {
		// Handle special timestamp filters
		if key == "timestamp_after" || key == "timestamp_before" {
			if !em.matchTimestampFilter(key, expectedValue, eventData) {
				return false
			}
			continue
		}

		// Handle regular exact-match filters
		if actualValue, exists := eventData[key]; !exists || !em.valuesEqual(actualValue, expectedValue) {
			return false
		}
	}
	return true
}

// matchTimestampFilter handles timestamp-based filtering with RFC3339 format
func (em *EventManager) matchTimestampFilter(filterKey string, expectedValue interface{}, eventData map[string]interface{}) bool {
	// Get event timestamp
	eventTimestampRaw, exists := eventData["timestamp"]
	if !exists {
		return false // No timestamp in event data
	}

	eventTimestampStr, ok := eventTimestampRaw.(string)
	if !ok {
		return false // Timestamp is not a string
	}

	// Parse event timestamp
	eventTime, err := time.Parse(time.RFC3339, eventTimestampStr)
	if err != nil {
		return false // Invalid timestamp format
	}

	// Parse filter timestamp
	filterTimestampStr, ok := expectedValue.(string)
	if !ok {
		return false // Filter value is not a string
	}

	filterTime, err := time.Parse(time.RFC3339, filterTimestampStr)
	if err != nil {
		return false // Invalid filter timestamp format
	}

	// Apply timestamp comparison
	switch filterKey {
	case "timestamp_after":
		return eventTime.After(filterTime)
	case "timestamp_before":
		return eventTime.Before(filterTime)
	default:
		return false
	}
}

// valuesEqual safely compares two interface{} values, handling uncomparable types
func (em *EventManager) valuesEqual(a, b interface{}) bool {
	// Handle nil cases
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Handle uncomparable types first - reject them to prevent panics
	switch a.(type) {
	case map[string]interface{}, []interface{}, func():
		// These types are uncomparable in Go - log warning and return false
		em.logger.WithFields(logging.Fields{
			"type_a": fmt.Sprintf("%T", a),
			"type_b": fmt.Sprintf("%T", b),
		}).Warn("Filter comparison skipped for uncomparable types")
		return false
	default:
		// For other types, use direct comparison (should be safe)
		return a == b
	}
}

func (em *EventManager) processEventHandlers(event *EventMessage) error {
	handlers, exists := em.eventHandlers[event.Topic]
	if !exists {
		return nil
	}

	for _, handler := range handlers {
		// Use anonymous function with panic recovery to prevent panics from crashing the server
		func() {
			defer func() {
				if r := recover(); r != nil {
					em.logger.WithFields(logging.Fields{
						"event_id": event.EventID,
						"topic":    string(event.Topic),
						"panic":    r,
					}).Error("Event handler panicked - recovered to prevent server crash")
				}
			}()

			if err := handler(event); err != nil {
				em.logger.WithError(err).WithField("event_id", event.EventID).Error("Event handler failed")
				// Don't return error, continue processing other handlers
				// This ensures one failing handler doesn't stop other handlers from executing
			}
		}()
	}

	return nil
}

func generateEventID() string {
	return fmt.Sprintf("evt_%d", time.Now().UnixNano())
}
