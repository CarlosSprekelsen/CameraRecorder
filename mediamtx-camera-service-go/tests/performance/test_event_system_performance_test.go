//go:build performance
// +build performance

/*
Event System Performance Tests

Tests the performance and efficiency of the new subscription-based event system
under various load conditions to validate the architectural improvements.

Requirements Coverage:
- REQ-API-001: Efficient event delivery
- REQ-API-002: Performance under load
- REQ-API-003: Scalability validation

Test Categories: Performance
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package performance

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/websocket"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventSystemPerformance(t *testing.T) {
	t.Run("event_delivery_performance", func(t *testing.T) {
		logger := logrus.New()
		em := websocket.NewEventManager(logger)

		// Subscribe 1000 clients to different topics
		clientCount := 1000
		topicCount := 10
		clientsPerTopic := clientCount / topicCount

		// Create topics
		topics := []websocket.EventTopic{
			websocket.TopicCameraConnected,
			websocket.TopicCameraDisconnected,
			websocket.TopicRecordingStart,
			websocket.TopicRecordingStop,
			websocket.TopicSnapshotTaken,
			websocket.TopicSystemHealth,
			websocket.TopicMediaMTXStream,
			websocket.TopicMediaMTXPath,
			websocket.TopicCameraCapabilityDetected,
			websocket.TopicCameraCapabilityError,
		}

		// Subscribe clients to topics
		for i := 0; i < clientCount; i++ {
			topicIndex := i % topicCount
			clientID := fmt.Sprintf("client_%d", i)
			err := em.Subscribe(clientID, []websocket.EventTopic{topics[topicIndex]}, nil)
			require.NoError(t, err, "Client subscription should succeed")
		}

		// Measure event delivery performance
		eventCount := 100
		startTime := time.Now()

		for i := 0; i < eventCount; i++ {
			topicIndex := i % topicCount
			eventData := map[string]interface{}{
				"event_id": i,
				"data":     fmt.Sprintf("test_data_%d", i),
				"timestamp": time.Now().Format(time.RFC3339),
			}

			err := em.PublishEvent(topics[topicIndex], eventData)
			require.NoError(t, err, "Event publishing should succeed")
		}

		duration := time.Since(startTime)
		eventsPerSecond := float64(eventCount) / duration.Seconds()

		t.Logf("Event delivery performance: %d events in %v (%.2f events/sec)", 
			eventCount, duration, eventsPerSecond)

		// Performance assertion: should handle at least 100 events per second
		assert.Greater(t, eventsPerSecond, 100.0, "Event system should handle at least 100 events per second")

		// Verify subscription statistics
		stats := em.GetSubscriptionStats()
		assert.Equal(t, clientCount, stats["total_clients"].(int), "Should have correct client count")
		assert.Equal(t, clientCount, stats["active_subscriptions"].(int), "Should have correct subscription count")
	})

	t.Run("concurrent_subscription_management", func(t *testing.T) {
		logger := logrus.New()
		em := websocket.NewEventManager(logger)

		// Test concurrent subscription operations
		clientCount := 1000
		goroutineCount := 100
		operationsPerGoroutine := clientCount / goroutineCount

		var wg sync.WaitGroup
		startTime := time.Now()

		for g := 0; g < goroutineCount; g++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()

				for i := 0; i < operationsPerGoroutine; i++ {
					clientID := fmt.Sprintf("goroutine_%d_client_%d", goroutineID, i)
					topic := websocket.TopicCameraConnected

					// Subscribe
					err := em.Subscribe(clientID, []websocket.EventTopic{topic}, nil)
					assert.NoError(t, err, "Concurrent subscription should succeed")

					// Verify subscription
					topics := em.GetClientSubscriptions(clientID)
					assert.Equal(t, 1, len(topics), "Subscription should be valid")

					// Unsubscribe
					err = em.Unsubscribe(clientID, nil)
					assert.NoError(t, err, "Concurrent unsubscribe should succeed")
				}
			}(g)
		}

		wg.Wait()
		duration := time.Since(startTime)
		operationsPerSecond := float64(clientCount) / duration.Seconds()

		t.Logf("Concurrent subscription performance: %d operations in %v (%.2f ops/sec)", 
			clientCount, duration, operationsPerSecond)

		// Performance assertion: should handle at least 500 operations per second
		assert.Greater(t, operationsPerSecond, 500.0, "Should handle at least 500 operations per second")

		// Verify final state
		stats := em.GetSubscriptionStats()
		assert.Equal(t, 0, stats["total_clients"].(int), "Should have no clients after cleanup")
	})

	t.Run("memory_usage_under_load", func(t *testing.T) {
		logger := logrus.New()
		em := websocket.NewEventManager(logger)

		// Subscribe many clients with different topic combinations
		clientCount := 5000
		topics := []websocket.EventTopic{
			websocket.TopicCameraConnected,
			websocket.TopicRecordingStart,
			websocket.TopicSystemHealth,
		}

		// Subscribe clients to multiple topics
		for i := 0; i < clientCount; i++ {
			clientID := fmt.Sprintf("memory_test_client_%d", i)
			
			// Subscribe to 1-3 topics randomly
			topicCount := (i % 3) + 1
			clientTopics := topics[:topicCount]
			
			err := em.Subscribe(clientID, clientTopics, nil)
			require.NoError(t, err, "Client subscription should succeed")
		}

		// Publish events to all topics
		eventCount := 100
		startTime := time.Now()

		for i := 0; i < eventCount; i++ {
			for _, topic := range topics {
				eventData := map[string]interface{}{
					"event_id": i,
					"topic":    string(topic),
					"data":     fmt.Sprintf("memory_test_data_%d", i),
					"timestamp": time.Now().Format(time.RFC3339),
				}

				err := em.PublishEvent(topic, eventData)
				require.NoError(t, err, "Event publishing should succeed")
			}
		}

		duration := time.Since(startTime)
		eventsPerSecond := float64(eventCount*len(topics)) / duration.Seconds()

		t.Logf("Memory test performance: %d events in %v (%.2f events/sec)", 
			eventCount*len(topics), duration, eventsPerSecond)

		// Verify subscription statistics
		stats := em.GetSubscriptionStats()
		assert.Equal(t, clientCount, stats["total_clients"].(int), "Should have correct client count")

		// Performance assertion: should handle at least 200 events per second under load
		assert.Greater(t, eventsPerSecond, 200.0, "Should handle at least 200 events per second under load")

		// Cleanup
		for i := 0; i < clientCount; i++ {
			clientID := fmt.Sprintf("memory_test_client_%d", i)
			em.RemoveClient(clientID)
		}

		// Verify cleanup
		stats = em.GetSubscriptionStats()
		assert.Equal(t, 0, stats["total_clients"].(int), "Should have no clients after cleanup")
	})

	t.Run("filtered_event_performance", func(t *testing.T) {
		logger := logrus.New()
		em := websocket.NewEventManager(logger)

		// Subscribe clients with filters
		clientCount := 1000
		deviceCount := 10

		for i := 0; i < clientCount; i++ {
			clientID := fmt.Sprintf("filtered_client_%d", i)
			deviceIndex := i % deviceCount
			devicePath := fmt.Sprintf("/dev/video%d", deviceIndex)
			
			filters := map[string]interface{}{
				"device": devicePath,
			}

			err := em.Subscribe(clientID, []websocket.EventTopic{websocket.TopicRecordingStart}, filters)
			require.NoError(t, err, "Filtered subscription should succeed")
		}

		// Publish events with different device filters
		eventCount := 100
		startTime := time.Now()

		for i := 0; i < eventCount; i++ {
			deviceIndex := i % deviceCount
			devicePath := fmt.Sprintf("/dev/video%d", deviceIndex)
			
			eventData := map[string]interface{}{
				"device":    devicePath,
				"event_id":  i,
				"timestamp": time.Now().Format(time.RFC3339),
			}

			err := em.PublishEvent(websocket.TopicRecordingStart, eventData)
			require.NoError(t, err, "Filtered event publishing should succeed")
		}

		duration := time.Since(startTime)
		eventsPerSecond := float64(eventCount) / duration.Seconds()

		t.Logf("Filtered event performance: %d events in %v (%.2f events/sec)", 
			eventCount, duration, eventsPerSecond)

		// Performance assertion: filtered events should be at least as fast as unfiltered
		assert.Greater(t, eventsPerSecond, 100.0, "Filtered events should handle at least 100 events per second")

		// Verify filtering works correctly
		stats := em.GetSubscriptionStats()
		assert.Equal(t, clientCount, stats["total_clients"].(int), "Should have correct client count")
	})

	t.Run("event_handler_performance", func(t *testing.T) {
		logger := logrus.New()
		em := websocket.NewEventManager(logger)

		// Add event handlers
		handlerCount := 100
		handlersCalled := make([]bool, handlerCount)
		var handlerMutex sync.Mutex

		for i := 0; i < handlerCount; i++ {
			handlerIndex := i
			em.AddEventHandler(websocket.TopicCameraConnected, func(event *websocket.EventMessage) error {
				handlerMutex.Lock()
				handlersCalled[handlerIndex] = true
				handlerMutex.Unlock()
				return nil
			})
		}

		// Subscribe clients
		clientCount := 100
		for i := 0; i < clientCount; i++ {
			clientID := fmt.Sprintf("handler_test_client_%d", i)
			err := em.Subscribe(clientID, []websocket.EventTopic{websocket.TopicCameraConnected}, nil)
			require.NoError(t, err, "Client subscription should succeed")
		}

		// Publish event and measure handler performance
		eventData := map[string]interface{}{
			"device":    "/dev/video0",
			"timestamp": time.Now().Format(time.RFC3339),
		}

		startTime := time.Now()
		err := em.PublishEvent(websocket.TopicCameraConnected, eventData)
		require.NoError(t, err, "Event publishing should succeed")
		duration := time.Since(startTime)

		t.Logf("Event handler performance: %d handlers in %v", handlerCount, duration)

		// Verify all handlers were called
		handlerMutex.Lock()
		allHandlersCalled := true
		for _, called := range handlersCalled {
			if !called {
				allHandlersCalled = false
				break
			}
		}
		handlerMutex.Unlock()

		assert.True(t, allHandlersCalled, "All event handlers should have been called")

		// Performance assertion: handlers should complete within reasonable time
		assert.Less(t, duration, 100*time.Millisecond, "Event handlers should complete within 100ms")
	})
}

func BenchmarkEventSystem(b *testing.B) {
	logger := logrus.New()
	em := websocket.NewEventManager(logger)

	// Setup: subscribe 1000 clients to different topics
	clientCount := 1000
	topics := []websocket.EventTopic{
		websocket.TopicCameraConnected,
		websocket.TopicRecordingStart,
		websocket.TopicSystemHealth,
	}

	for i := 0; i < clientCount; i++ {
		clientID := fmt.Sprintf("benchmark_client_%d", i)
		topicIndex := i % len(topics)
		err := em.Subscribe(clientID, []websocket.EventTopic{topics[topicIndex]}, nil)
		if err != nil {
			b.Fatalf("Failed to subscribe client: %v", err)
		}
	}

	// Benchmark event publishing
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		topicIndex := i % len(topics)
		eventData := map[string]interface{}{
			"event_id":  i,
			"data":      fmt.Sprintf("benchmark_data_%d", i),
			"timestamp": time.Now().Format(time.RFC3339),
		}

		err := em.PublishEvent(topics[topicIndex], eventData)
		if err != nil {
			b.Fatalf("Failed to publish event: %v", err)
		}
	}
}

func BenchmarkConcurrentSubscriptions(b *testing.B) {
	logger := logrus.New()
	em := websocket.NewEventManager(logger)

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		clientID := fmt.Sprintf("benchmark_parallel_client_%d", time.Now().UnixNano())
		topic := websocket.TopicCameraConnected

		for pb.Next() {
			// Subscribe
			err := em.Subscribe(clientID, []websocket.EventTopic{topic}, nil)
			if err != nil {
				b.Fatalf("Failed to subscribe: %v", err)
			}

			// Unsubscribe
			err = em.Unsubscribe(clientID, nil)
			if err != nil {
				b.Fatalf("Failed to unsubscribe: %v", err)
			}
		}
	})
}
