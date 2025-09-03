//go:build unit
// +build unit

/*
Event System Tests

Tests the new subscription-based event system that replaces the inefficient
broadcast-to-all approach with topic-based filtering.

Requirements Coverage:
- REQ-API-001: Efficient event delivery
- REQ-API-002: Client subscription management
- REQ-API-003: Topic-based filtering

Test Categories: Unit
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package websocket

import (
	"fmt"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/websocket"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventManager(t *testing.T) {
	t.Run("event_manager_creation", func(t *testing.T) {
		logger := logrus.New()
		em := websocket.NewEventManager(logger)

		assert.NotNil(t, em, "Event manager should be created successfully")
		stats := em.GetSubscriptionStats()
		assert.Equal(t, 0, stats["total_clients"].(int), "Should start with no clients")
	})

	t.Run("event_topic_validation", func(t *testing.T) {
		logger := logrus.New()
		em := websocket.NewEventManager(logger)

		// Valid topics
		validTopics := []websocket.EventTopic{
			websocket.TopicCameraConnected,
			websocket.TopicRecordingStart,
			websocket.TopicSystemHealth,
		}

		for _, topic := range validTopics {
			err := em.Subscribe("client1", []websocket.EventTopic{topic}, nil)
			assert.NoError(t, err, "Valid topic %s should be accepted", topic)
		}

		// Invalid topic
		err := em.Subscribe("client1", []websocket.EventTopic{"invalid.topic"}, nil)
		assert.Error(t, err, "Invalid topic should be rejected")
	})

	t.Run("client_subscription_management", func(t *testing.T) {
		logger := logrus.New()
		em := websocket.NewEventManager(logger)

		clientID := "test_client"
		topics := []websocket.EventTopic{
			websocket.TopicCameraConnected,
			websocket.TopicRecordingStart,
		}

		// Subscribe client
		err := em.Subscribe(clientID, topics, nil)
		require.NoError(t, err, "Client subscription should succeed")

		// Verify subscription
		clientTopics := em.GetClientSubscriptions(clientID)
		assert.Equal(t, len(topics), len(clientTopics), "Client should be subscribed to correct number of topics")

		// Check subscribers for topic
		subscribers := em.GetSubscribersForTopic(websocket.TopicCameraConnected)
		assert.Contains(t, subscribers, clientID, "Client should be in subscribers list")

		// Unsubscribe from one topic
		err = em.Unsubscribe(clientID, []websocket.EventTopic{websocket.TopicCameraConnected})
		require.NoError(t, err, "Partial unsubscribe should succeed")

		// Verify partial unsubscribe
		clientTopics = em.GetClientSubscriptions(clientID)
		assert.Equal(t, 1, len(clientTopics), "Client should still be subscribed to one topic")
		assert.Equal(t, websocket.TopicRecordingStart, clientTopics[0], "Client should still be subscribed to recording topic")

		// Unsubscribe from all topics
		err = em.Unsubscribe(clientID, nil)
		require.NoError(t, err, "Full unsubscribe should succeed")

		// Verify full unsubscribe
		clientTopics = em.GetClientSubscriptions(clientID)
		assert.Equal(t, 0, len(clientTopics), "Client should have no subscriptions")
	})

	t.Run("event_filtering", func(t *testing.T) {
		logger := logrus.New()
		em := websocket.NewEventManager(logger)

		clientID := "filtered_client"
		topics := []websocket.EventTopic{websocket.TopicRecordingStart}
		filters := map[string]interface{}{
			"device": "/dev/video0",
		}

		// Subscribe with filters
		err := em.Subscribe(clientID, topics, filters)
		require.NoError(t, err, "Filtered subscription should succeed")

		// Event that matches filter
		matchingEvent := map[string]interface{}{
			"device": "/dev/video0",
			"status": "started",
		}

		// Event that doesn't match filter
		nonMatchingEvent := map[string]interface{}{
			"device": "/dev/video1",
			"status": "started",
		}

		// Publish matching event
		err = em.PublishEvent(websocket.TopicRecordingStart, matchingEvent)
		require.NoError(t, err, "Matching event should be published")

		// Publish non-matching event
		err = em.PublishEvent(websocket.TopicRecordingStart, nonMatchingEvent)
		require.NoError(t, err, "Non-matching event should be published")

		// Verify subscription stats
		stats := em.GetSubscriptionStats()
		assert.Equal(t, 1, stats["total_clients"].(int), "Should have one client")
		assert.Equal(t, 1, stats["active_subscriptions"].(int), "Should have one active subscription")
	})

	t.Run("event_handlers", func(t *testing.T) {
		logger := logrus.New()
		em := websocket.NewEventManager(logger)

		handlerCalled := false
		handlerData := make(map[string]interface{})

		// Add event handler
		em.AddEventHandler(websocket.TopicCameraConnected, func(event *websocket.EventMessage) error {
			handlerCalled = true
			handlerData = event.Data
			return nil
		})

		// Publish event
		eventData := map[string]interface{}{
			"device": "/dev/video0",
			"status": "connected",
		}

		err := em.PublishEvent(websocket.TopicCameraConnected, eventData)
		require.NoError(t, err, "Event should be published")

		// Verify handler was called
		assert.True(t, handlerCalled, "Event handler should have been called")
		assert.Equal(t, eventData["device"], handlerData["device"], "Handler should receive correct data")
	})

	t.Run("client_removal", func(t *testing.T) {
		logger := logrus.New()
		em := websocket.NewEventManager(logger)

		clientID := "removable_client"
		topics := []websocket.EventTopic{
			websocket.TopicCameraConnected,
			websocket.TopicRecordingStart,
		}

		// Subscribe client to multiple topics
		err := em.Subscribe(clientID, topics, nil)
		require.NoError(t, err, "Client subscription should succeed")

		// Verify initial state
		stats := em.GetSubscriptionStats()
		assert.Equal(t, 1, stats["total_clients"].(int), "Should have one client initially")

		// Remove client
		em.RemoveClient(clientID)

		// Verify client removal
		stats = em.GetSubscriptionStats()
		assert.Equal(t, 0, stats["total_clients"].(int), "Should have no clients after removal")

		// Verify no subscriptions remain
		clientTopics := em.GetClientSubscriptions(clientID)
		assert.Equal(t, 0, len(clientTopics), "Client should have no subscriptions after removal")
	})

	t.Run("subscription_statistics", func(t *testing.T) {
		logger := logrus.New()
		em := websocket.NewEventManager(logger)

		// Subscribe multiple clients to different topics
		clients := []string{"client1", "client2", "client3"}
		topics := []websocket.EventTopic{
			websocket.TopicCameraConnected,
			websocket.TopicRecordingStart,
			websocket.TopicSystemHealth,
		}

		for i, clientID := range clients {
			err := em.Subscribe(clientID, []websocket.EventTopic{topics[i]}, nil)
			require.NoError(t, err, "Client %s subscription should succeed", clientID)
		}

		// Get statistics
		stats := em.GetSubscriptionStats()

		// Verify statistics
		assert.Equal(t, 3, stats["total_clients"].(int), "Should have three clients")
		assert.Equal(t, 3, stats["active_subscriptions"].(int), "Should have three active subscriptions")
		assert.Equal(t, 3, stats["total_topics"].(int), "Should have three topics")

		// Verify topic distribution
		topicDistribution := stats["topic_distribution"].(map[string]int)
		assert.Equal(t, 1, topicDistribution["camera.connected"], "Should have one subscriber for camera.connected")
		assert.Equal(t, 1, topicDistribution["recording.start"], "Should have one subscriber for recording.start")
		assert.Equal(t, 1, topicDistribution["system.health"], "Should have one subscriber for system.health")
	})

	t.Run("concurrent_subscriptions", func(t *testing.T) {
		logger := logrus.New()
		em := websocket.NewEventManager(logger)

		// Test concurrent subscription operations
		done := make(chan bool, 10)

		for i := 0; i < 10; i++ {
			go func(id int) {
				clientID := fmt.Sprintf("concurrent_client_%d", id)
				topic := websocket.TopicCameraConnected

				// Subscribe
				err := em.Subscribe(clientID, []websocket.EventTopic{topic}, nil)
				assert.NoError(t, err, "Concurrent subscription should succeed")

				// Verify subscription
				topics := em.GetClientSubscriptions(clientID)
				assert.Equal(t, 1, len(topics), "Concurrent subscription should be valid")

				// Unsubscribe
				err = em.Unsubscribe(clientID, nil)
				assert.NoError(t, err, "Concurrent unsubscribe should succeed")

				done <- true
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}

		// Verify final state
		stats := em.GetSubscriptionStats()
		assert.Equal(t, 0, stats["total_clients"].(int), "Should have no clients after concurrent operations")
	})

	t.Run("event_message_structure", func(t *testing.T) {
		logger := logrus.New()
		em := websocket.NewEventManager(logger)

		// Subscribe to events
		clientID := "message_test_client"
		err := em.Subscribe(clientID, []websocket.EventTopic{websocket.TopicCameraConnected}, nil)
		require.NoError(t, err, "Subscription should succeed")

		// Publish event
		eventData := map[string]interface{}{
			"device": "/dev/video0",
			"status": "connected",
		}

		err = em.PublishEvent(websocket.TopicCameraConnected, eventData)
		require.NoError(t, err, "Event should be published")

		// Verify event structure through handler
		var capturedEvent *websocket.EventMessage
		em.AddEventHandler(websocket.TopicCameraConnected, func(event *websocket.EventMessage) error {
			capturedEvent = event
			return nil
		})

		// Publish another event to trigger handler
		err = em.PublishEvent(websocket.TopicCameraConnected, eventData)
		require.NoError(t, err, "Second event should be published")

		// Verify event structure
		require.NotNil(t, capturedEvent, "Event should be captured by handler")
		assert.Equal(t, websocket.TopicCameraConnected, capturedEvent.Topic, "Event should have correct topic")
		assert.Equal(t, eventData, capturedEvent.Data, "Event should have correct data")
		assert.NotEmpty(t, capturedEvent.EventID, "Event should have event ID")
		assert.WithinDuration(t, time.Now(), capturedEvent.Timestamp, 2*time.Second, "Event should have recent timestamp")
	})
}

func TestEventTopicConstants(t *testing.T) {
	t.Run("event_topic_values", func(t *testing.T) {
		// Verify all event topic constants are defined
		expectedTopics := map[websocket.EventTopic]string{
			websocket.TopicCameraConnected:    "camera.connected",
			websocket.TopicCameraDisconnected: "camera.disconnected",
			websocket.TopicCameraStatusChange: "camera.status_change",
			websocket.TopicRecordingStart:     "recording.start",
			websocket.TopicRecordingStop:      "recording.stop",
			websocket.TopicRecordingProgress:  "recording.progress",
			websocket.TopicRecordingError:     "recording.error",
			websocket.TopicSnapshotTaken:      "snapshot.taken",
			websocket.TopicSnapshotError:      "snapshot.error",
			websocket.TopicSystemHealth:       "system.health",
			websocket.TopicSystemError:        "system.error",
			websocket.TopicMediaMTXStream:     "mediamtx.stream",
			websocket.TopicMediaMTXPath:       "mediamtx.path",
			websocket.TopicMediaMTXError:      "mediamtx.error",
		}

		for topic, expectedValue := range expectedTopics {
			assert.Equal(t, expectedValue, string(topic), "Event topic %s should have correct value", topic)
		}
	})
}
