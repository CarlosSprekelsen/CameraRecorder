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
	Filters   map[string]interface{} `json:"filters,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	LastSeen  time.Time              `json:"last_seen"`
	Active    bool                   `json:"active"`
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
	} else {
		// Remove specific topics
		for _, topic := range topics {
			em.removeTopicSubscription(clientID, topic)
		}

		// Update subscription topics
		em.updateSubscriptionTopics(clientID, topics, true)
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

	em.removeClientSubscriptions(clientID)

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
	em.mu.RLock()
	defer em.mu.RUnlock()

	stats := map[string]interface{}{
		"total_clients":        len(em.subscriptions),
		"total_topics":         len(em.topicSubscriptions),
		"active_subscriptions": 0,
		"topic_distribution":   make(map[string]int),
	}

	// Count active subscriptions and topic distribution
	for _, subscription := range em.subscriptions {
		if subscription.Active {
			stats["active_subscriptions"] = stats["active_subscriptions"].(int) + 1
		}
	}

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
		if actualValue, exists := eventData[key]; !exists || !em.valuesEqual(actualValue, expectedValue) {
			return false
		}
	}
	return true
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
		if err := handler(event); err != nil {
			em.logger.WithError(err).WithField("event_id", event.EventID).Error("Event handler failed")
			return fmt.Errorf("event handler failed for event %s: %w", event.EventID, err)
		}
	}

	return nil
}

func generateEventID() string {
	return fmt.Sprintf("evt_%d", time.Now().UnixNano())
}
