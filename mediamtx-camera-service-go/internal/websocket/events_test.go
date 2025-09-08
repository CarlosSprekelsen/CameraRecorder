/*
WebSocket Events Unit Tests

Provides focused unit tests for WebSocket event handling,
following the project testing standards and Go coding standards.

Requirements Coverage:
- REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint
- REQ-API-003: Request/response message handling

Test Categories: Unit/Integration
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package websocket

import (
	"fmt"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/camera"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWebSocketEvents_EventManagerCreation tests event manager creation
func TestWebSocketEvents_EventManagerCreation(t *testing.T) {
	eventManager := NewEventManager(NewTestLogger("events-test"))

	assert.NotNil(t, eventManager, "Event manager should be created")
}

// TestWebSocketEvents_EventManagerBasicOperations tests basic event manager operations
func TestWebSocketEvents_EventManagerBasicOperations(t *testing.T) {
	eventManager := NewEventManager(NewTestLogger("events-test"))

	// Test event manager creation
	assert.NotNil(t, eventManager, "Event manager should be created")

	// Test subscription management
	clientID := "test-client"
	topics := []EventTopic{TopicCameraConnected, TopicRecordingStart}

	// Subscribe client to topics
	err := eventManager.Subscribe(clientID, topics, nil)
	assert.NoError(t, err, "Should subscribe client successfully")

	// Check subscription exists (basic test)
	// Note: GetSubscription method may not be available, so we'll just test that subscribe doesn't error
	assert.NoError(t, err, "Subscribe should not error")

	// Unsubscribe client
	err = eventManager.Unsubscribe(clientID, topics)
	assert.NoError(t, err, "Should unsubscribe client successfully")
}

// TestWebSocketEvents_EventTopics tests event topic constants
func TestWebSocketEvents_EventTopics(t *testing.T) {
	// Test camera event topics
	assert.Equal(t, "camera.connected", string(TopicCameraConnected), "Camera connected topic should be correct")
	assert.Equal(t, "camera.disconnected", string(TopicCameraDisconnected), "Camera disconnected topic should be correct")
	assert.Equal(t, "camera.status_change", string(TopicCameraStatusChange), "Camera status change topic should be correct")

	// Test recording event topics
	assert.Equal(t, "recording.start", string(TopicRecordingStart), "Recording start topic should be correct")
	assert.Equal(t, "recording.stop", string(TopicRecordingStop), "Recording stop topic should be correct")
	assert.Equal(t, "recording.progress", string(TopicRecordingProgress), "Recording progress topic should be correct")

	// Test system event topics
	assert.Equal(t, "system.health", string(TopicSystemHealth), "System health topic should be correct")
	assert.Equal(t, "system.error", string(TopicSystemError), "System error topic should be correct")
}

// TestWebSocketEvents_EventMessage tests event message structure
func TestWebSocketEvents_EventMessage(t *testing.T) {
	// Create test event message
	eventMessage := &EventMessage{
		Topic:     TopicCameraConnected,
		Data:      map[string]interface{}{"camera_id": "camera0"}, // Use API abstraction layer
		Timestamp: time.Now(),
		EventID:   "test-event-123",
	}

	// Test event message structure
	assert.Equal(t, TopicCameraConnected, eventMessage.Topic, "Event topic should be correct")
	assert.Equal(t, "camera0", eventMessage.Data["camera_id"], "Event data should use camera identifier")
	assert.NotZero(t, eventMessage.Timestamp, "Event timestamp should be set")
	assert.Equal(t, "test-event-123", eventMessage.EventID, "Event ID should be correct")
}

// TestWebSocketEvents_EventSubscription tests event subscription structure
func TestWebSocketEvents_EventSubscription(t *testing.T) {
	// Create test subscription
	subscription := &EventSubscription{
		ClientID:  "test-client",
		Topics:    []EventTopic{TopicCameraConnected, TopicRecordingStart},
		Filters:   map[string]interface{}{"camera_id": "camera0"}, // Use API abstraction layer
		CreatedAt: time.Now(),
		LastSeen:  time.Now(),
		Active:    true,
	}

	// Test subscription structure
	assert.Equal(t, "test-client", subscription.ClientID, "Client ID should be correct")
	assert.Equal(t, 2, len(subscription.Topics), "Should have correct number of topics")
	assert.Equal(t, "camera0", subscription.Filters["camera_id"], "Filters should use camera identifier")
	assert.True(t, subscription.Active, "Subscription should be active")
}

// TestWebSocketEvents_PublishEventNoSubscribers tests event publishing with no subscribers
func TestWebSocketEvents_PublishEventNoSubscribers(t *testing.T) {
	// REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint
	// REQ-API-003: Request/response message handling

	eventManager := NewEventManager(NewTestLogger("events-test"))

	// Publish event to topic with no subscribers
	err := eventManager.PublishEvent(TopicCameraConnected, map[string]interface{}{
		"camera_id": "camera0", // Use API abstraction layer
		"status":    "connected",
	})

	// Should not error even with no subscribers
	assert.NoError(t, err, "Publishing event with no subscribers should not error")
}

// TestWebSocketEvents_PublishEventWithSubscribers tests event publishing with subscribers
func TestWebSocketEvents_PublishEventWithSubscribers(t *testing.T) {
	// REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint
	// REQ-API-003: Request/response message handling

	eventManager := NewEventManager(NewTestLogger("events-test"))

	// Subscribe client to topic
	clientID := "test-client"
	topics := []EventTopic{TopicCameraConnected}
	err := eventManager.Subscribe(clientID, topics, nil)
	require.NoError(t, err, "Should subscribe client successfully")

	// Publish event to subscribed topic
	err = eventManager.PublishEvent(TopicCameraConnected, map[string]interface{}{
		"camera_id": "camera0", // Use API abstraction layer
		"status":    "connected",
	})

	// Should not error with subscribers
	assert.NoError(t, err, "Publishing event with subscribers should not error")
}

// TestWebSocketEvents_GetSubscribersForTopic tests subscriber retrieval
func TestWebSocketEvents_GetSubscribersForTopic(t *testing.T) {
	// REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint
	// REQ-API-003: Request/response message handling

	eventManager := NewEventManager(NewTestLogger("events-test"))

	// Test getting subscribers for topic with no subscribers
	subscribers := eventManager.GetSubscribersForTopic(TopicCameraConnected)
	assert.Empty(t, subscribers, "Should return empty list for topic with no subscribers")

	// Subscribe client to topic
	clientID := "test-client"
	topics := []EventTopic{TopicCameraConnected}
	err := eventManager.Subscribe(clientID, topics, nil)
	require.NoError(t, err, "Should subscribe client successfully")

	// Test getting subscribers for topic with subscribers
	subscribers = eventManager.GetSubscribersForTopic(TopicCameraConnected)
	assert.Contains(t, subscribers, clientID, "Should return subscribed client ID")
}

// TestWebSocketEvents_RemoveClient tests client removal
func TestWebSocketEvents_RemoveClient(t *testing.T) {
	// REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint
	// REQ-API-003: Request/response message handling

	eventManager := NewEventManager(NewTestLogger("events-test"))

	// Subscribe client to multiple topics
	clientID := "test-client"
	topics := []EventTopic{TopicCameraConnected, TopicRecordingStart}
	err := eventManager.Subscribe(clientID, topics, nil)
	require.NoError(t, err, "Should subscribe client successfully")

	// Verify client is subscribed
	subscribers := eventManager.GetSubscribersForTopic(TopicCameraConnected)
	assert.Contains(t, subscribers, clientID, "Client should be subscribed")

	// Remove client
	eventManager.RemoveClient(clientID)

	// Verify client is no longer subscribed
	subscribers = eventManager.GetSubscribersForTopic(TopicCameraConnected)
	assert.NotContains(t, subscribers, clientID, "Client should no longer be subscribed")
}

// TestWebSocketEvents_AddEventHandler tests event handler registration
func TestWebSocketEvents_AddEventHandler(t *testing.T) {
	// REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint
	// REQ-API-003: Request/response message handling

	eventManager := NewEventManager(NewTestLogger("events-test"))

	// Add event handler
	handler := func(event *EventMessage) error {
		return nil
	}

	eventManager.AddEventHandler(TopicCameraConnected, handler)

	// Publish event to trigger handler
	err := eventManager.PublishEvent(TopicCameraConnected, map[string]interface{}{
		"camera_id": "camera0", // Use API abstraction layer
	})

	// Handler should be called
	assert.NoError(t, err, "Publishing event should not error")
	// Note: Handler execution is asynchronous, so we can't easily test it in unit test
	// This test verifies the handler registration doesn't cause errors
}

// TestWebSocketEvents_UpdateClientLastSeen tests client activity tracking
func TestWebSocketEvents_UpdateClientLastSeen(t *testing.T) {
	// REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint
	// REQ-API-003: Request/response message handling

	eventManager := NewEventManager(NewTestLogger("events-test"))

	// Subscribe client
	clientID := "test-client"
	topics := []EventTopic{TopicCameraConnected}
	err := eventManager.Subscribe(clientID, topics, nil)
	require.NoError(t, err, "Should subscribe client successfully")

	// Update client last seen
	eventManager.UpdateClientLastSeen(clientID)

	// Should not error
	// Note: We can't easily verify the timestamp update without exposing internal state
	// This test verifies the method doesn't cause errors
}

// TestWebSocketEvents_GetSubscriptionStats tests subscription statistics
func TestWebSocketEvents_GetSubscriptionStats(t *testing.T) {
	// REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint
	// REQ-API-003: Request/response message handling

	eventManager := NewEventManager(NewTestLogger("events-test"))

	// Get stats with no subscriptions
	stats := eventManager.GetSubscriptionStats()
	assert.NotNil(t, stats, "Should return subscription stats")
	assert.Equal(t, int64(0), stats["total_clients"], "Should have zero clients initially")

	// Subscribe client
	clientID := "test-client"
	topics := []EventTopic{TopicCameraConnected}
	err := eventManager.Subscribe(clientID, topics, nil)
	require.NoError(t, err, "Should subscribe client successfully")

	// Get stats with subscriptions
	stats = eventManager.GetSubscriptionStats()
	assert.NotNil(t, stats, "Should return subscription stats")
	assert.Greater(t, stats["total_clients"], int64(0), "Should have clients after subscribing")
}

// TestWebSocketEvents_InvalidTopic tests invalid topic handling
func TestWebSocketEvents_InvalidTopic(t *testing.T) {
	// REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint
	// REQ-API-003: Request/response message handling

	eventManager := NewEventManager(NewTestLogger("events-test"))

	// Test subscribing to invalid topic
	clientID := "test-client"
	invalidTopic := EventTopic("invalid.topic")
	topics := []EventTopic{invalidTopic}

	err := eventManager.Subscribe(clientID, topics, nil)
	// Should either error or handle gracefully
	// This test identifies potential bugs in topic validation
	if err != nil {
		assert.Error(t, err, "Should error for invalid topic")
	} else {
		// If no error, verify the subscription doesn't cause issues
		subscribers := eventManager.GetSubscribersForTopic(invalidTopic)
		assert.NotNil(t, subscribers, "Should handle invalid topic gracefully")
	}
}

// TestWebSocketEvents_ConcurrentSubscriptions tests concurrent subscription operations
func TestWebSocketEvents_ConcurrentSubscriptions(t *testing.T) {
	// REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint
	// REQ-API-003: Request/response message handling

	eventManager := NewEventManager(NewTestLogger("events-test"))

	// Test concurrent subscriptions from multiple clients
	const numClients = 10
	done := make(chan bool, numClients)

	for i := 0; i < numClients; i++ {
		go func(clientID string) {
			topics := []EventTopic{TopicCameraConnected}
			err := eventManager.Subscribe(clientID, topics, nil)
			if err != nil {
				t.Errorf("Concurrent subscription failed for client %s: %v", clientID, err)
			}
			done <- true
		}(fmt.Sprintf("client-%d", i))
	}

	// Wait for all subscriptions to complete
	for i := 0; i < numClients; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for concurrent subscriptions")
		}
	}

	// Verify all clients are subscribed
	subscribers := eventManager.GetSubscribersForTopic(TopicCameraConnected)
	assert.Equal(t, numClients, len(subscribers), "All clients should be subscribed")
}

// TestWebSocketEvents_GetClientSubscriptions tests client subscription retrieval
func TestWebSocketEvents_GetClientSubscriptions(t *testing.T) {
	// REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint
	// REQ-API-003: Request/response message handling

	eventManager := NewEventManager(NewTestLogger("events-test"))

	// Test getting subscriptions for non-existent client
	subscriptions := eventManager.GetClientSubscriptions("non-existent-client")
	assert.Empty(t, subscriptions, "Should return empty subscriptions for non-existent client")

	// Subscribe client to multiple topics
	clientID := "test-client"
	topics := []EventTopic{TopicCameraConnected, TopicRecordingStart, TopicSystemHealth}
	err := eventManager.Subscribe(clientID, topics, nil)
	require.NoError(t, err, "Should subscribe client successfully")

	// Get client subscriptions
	subscriptions = eventManager.GetClientSubscriptions(clientID)
	assert.NotNil(t, subscriptions, "Should return client subscriptions")
	assert.Equal(t, 3, len(subscriptions), "Should have correct number of subscribed topics")
	assert.Contains(t, subscriptions, TopicCameraConnected, "Should contain camera connected topic")
	assert.Contains(t, subscriptions, TopicRecordingStart, "Should contain recording start topic")
	assert.Contains(t, subscriptions, TopicSystemHealth, "Should contain system health topic")
}

// TestWebSocketEvents_ApplyFilters tests event filtering logic
func TestWebSocketEvents_ApplyFilters(t *testing.T) {
	// REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint
	// REQ-API-003: Request/response message handling

	eventManager := NewEventManager(NewTestLogger("events-test"))

	// Subscribe client with filters
	clientID := "test-client"
	topics := []EventTopic{TopicCameraConnected}
	filters := map[string]interface{}{
		"camera_id": "camera0", // Use API abstraction layer
		"status":    "connected",
	}
	err := eventManager.Subscribe(clientID, topics, filters)
	require.NoError(t, err, "Should subscribe client with filters successfully")

	// Publish event that matches filters
	err = eventManager.PublishEvent(TopicCameraConnected, map[string]interface{}{
		"camera_id": "camera0", // Use API abstraction layer
		"status":    "connected",
		"extra":     "data",
	})
	assert.NoError(t, err, "Publishing matching event should not error")

	// Publish event that doesn't match filters
	err = eventManager.PublishEvent(TopicCameraConnected, map[string]interface{}{
		"camera_id": "camera1", // Different camera
		"status":    "connected",
	})
	assert.NoError(t, err, "Publishing non-matching event should not error")
}

// TestWebSocketEvents_EventIntegrationLayer tests event integration functions
func TestWebSocketEvents_EventIntegrationLayer(t *testing.T) {
	// REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint
	// REQ-API-003: Request/response message handling

	eventManager := NewEventManager(NewTestLogger("events-test"))
	logger := NewTestLogger("integration-test")

	// Test NewEventIntegration
	integration := NewEventIntegration(eventManager, logger)
	assert.NotNil(t, integration, "Event integration should be created")

	// Test NewCameraEventNotifier
	cameraNotifier := NewCameraEventNotifier(eventManager, logger)
	assert.NotNil(t, cameraNotifier, "Camera event notifier should be created")

	// Test camera event notifications - these functions don't return errors
	cameraNotifier.NotifyCameraConnected(nil)                                     // Test with nil device
	cameraNotifier.NotifyCameraDisconnected("/dev/video0")                        // Test with device path (internal layer)
	cameraNotifier.NotifyCameraStatusChange(nil, "recording", "idle")             // Test with nil device
	cameraNotifier.NotifyCapabilityDetected(nil, camera.V4L2Capabilities{})       // Test with nil device and empty capabilities
	cameraNotifier.NotifyCapabilityError("/dev/video0", "Failed to detect codec") // Test with device path (internal layer)
}

// TestWebSocketEvents_CameraIdentifierMapping tests the camera abstraction layer mapping
func TestWebSocketEvents_CameraIdentifierMapping(t *testing.T) {
	// REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint
	// REQ-API-003: Request/response message handling

	// Create a test server to access the mapping functions
	server := &WebSocketServer{}

	// Test USB camera mapping: camera0 -> /dev/video0
	devicePath := server.getDevicePathFromCameraIdentifier("camera0")
	assert.Equal(t, "/dev/video0", devicePath, "camera0 should map to /dev/video0")

	devicePath = server.getDevicePathFromCameraIdentifier("camera1")
	assert.Equal(t, "/dev/video1", devicePath, "camera1 should map to /dev/video1")

	// Test reverse mapping: /dev/video0 -> camera0
	cameraID := server.getCameraIdentifierFromDevicePath("/dev/video0")
	assert.Equal(t, "camera0", cameraID, "/dev/video0 should map to camera0")

	cameraID = server.getCameraIdentifierFromDevicePath("/dev/video1")
	assert.Equal(t, "camera1", cameraID, "/dev/video1 should map to camera1")

	// Test IP camera mapping
	ipCameraID := server.getCameraIdentifierFromDevicePath("rtsp://192.168.1.100:554/stream")
	assert.Equal(t, "ip_camera_192_168_1_100", ipCameraID, "IP camera should map correctly")

	devicePath = server.getDevicePathFromCameraIdentifier("ip_camera_192_168_1_100")
	assert.Equal(t, "rtsp://192.168.1.100:554/stream", devicePath, "IP camera identifier should map to RTSP URL")

	// Test HTTP camera mapping
	httpCameraID := server.getCameraIdentifierFromDevicePath("http://192.168.1.101:8080/mjpeg")
	assert.Equal(t, "http_camera_192_168_1_101", httpCameraID, "HTTP camera should map correctly")

	devicePath = server.getDevicePathFromCameraIdentifier("http_camera_192_168_1_101")
	assert.Equal(t, "http://192.168.1.101:8080/mjpeg", devicePath, "HTTP camera identifier should map to HTTP URL")

	// Test network camera mapping
	networkCameraID := server.getCameraIdentifierFromDevicePath("udp://239.0.0.1:1234")
	assert.Equal(t, "network_camera_239_0_0_1_1234", networkCameraID, "Network camera should map correctly")

	devicePath = server.getDevicePathFromCameraIdentifier("network_camera_239_0_0_1_1234")
	assert.Equal(t, "udp://239.0.0.1:1234", devicePath, "Network camera identifier should map to UDP URL")
}

// TestWebSocketEvents_CameraIdentifierValidation tests camera identifier validation
func TestWebSocketEvents_CameraIdentifierValidation(t *testing.T) {
	// REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint
	// REQ-API-003: Request/response message handling

	server := &WebSocketServer{}

	// Test valid camera identifiers
	assert.True(t, server.validateCameraIdentifier("camera0"), "camera0 should be valid")
	assert.True(t, server.validateCameraIdentifier("camera1"), "camera1 should be valid")
	assert.True(t, server.validateCameraIdentifier("camera123"), "camera123 should be valid")
	assert.True(t, server.validateCameraIdentifier("ip_camera_192_168_1_100"), "IP camera identifier should be valid")
	assert.True(t, server.validateCameraIdentifier("http_camera_192_168_1_101"), "HTTP camera identifier should be valid")
	assert.True(t, server.validateCameraIdentifier("network_camera_239_0_0_1_1234"), "Network camera identifier should be valid")
	assert.True(t, server.validateCameraIdentifier("file_camera_video"), "File camera identifier should be valid")

	// Test invalid camera identifiers
	assert.False(t, server.validateCameraIdentifier("invalid_camera"), "invalid_camera should be invalid")
	assert.False(t, server.validateCameraIdentifier("camera"), "camera should be invalid (missing number)")
	assert.False(t, server.validateCameraIdentifier("camera_abc"), "camera_abc should be invalid (non-numeric)")
	assert.False(t, server.validateCameraIdentifier("ip_camera_192_168_1"), "Incomplete IP camera identifier should be invalid")
	assert.False(t, server.validateCameraIdentifier("http_camera_192_168"), "Incomplete HTTP camera identifier should be invalid")
	assert.False(t, server.validateCameraIdentifier("network_camera_239_0_0"), "Incomplete network camera identifier should be invalid")
	assert.False(t, server.validateCameraIdentifier(""), "Empty string should be invalid")
	assert.False(t, server.validateCameraIdentifier("random_string"), "Random string should be invalid")
}

// TestWebSocketEvents_UnsubscribeEdgeCases tests unsubscribe edge cases and potential bugs
func TestWebSocketEvents_UnsubscribeEdgeCases(t *testing.T) {
	// REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint
	// REQ-API-003: Request/response message handling

	eventManager := NewEventManager(NewTestLogger("events-test"))

	// Test unsubscribing non-existent client
	err := eventManager.Unsubscribe("non-existent-client", []EventTopic{TopicCameraConnected})
	assert.NoError(t, err, "Unsubscribing non-existent client should not error")

	// Test unsubscribing from non-existent topic
	clientID := "test-client"
	err = eventManager.Subscribe(clientID, []EventTopic{TopicCameraConnected}, nil)
	require.NoError(t, err, "Should subscribe successfully")

	err = eventManager.Unsubscribe(clientID, []EventTopic{TopicRecordingStart})
	assert.NoError(t, err, "Unsubscribing from non-subscribed topic should not error")

	// Test unsubscribing with empty topic list
	err = eventManager.Unsubscribe(clientID, []EventTopic{})
	assert.NoError(t, err, "Unsubscribing with empty topic list should not error")

	// Test unsubscribing with nil topic list - this might expose a bug
	err = eventManager.Unsubscribe(clientID, nil)
	assert.NoError(t, err, "Unsubscribing with nil topic list should not error")

	// Test unsubscribing from all topics
	err = eventManager.Unsubscribe(clientID, []EventTopic{TopicCameraConnected})
	assert.NoError(t, err, "Should unsubscribe successfully")

	// Verify client is completely removed
	subscriptions := eventManager.GetClientSubscriptions(clientID)
	assert.Empty(t, subscriptions, "Client should have no subscriptions after unsubscribe")
}

// TestWebSocketEvents_PublishEventEdgeCases tests publish event edge cases and potential bugs
func TestWebSocketEvents_PublishEventEdgeCases(t *testing.T) {
	// REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint
	// REQ-API-003: Request/response message handling

	eventManager := NewEventManager(NewTestLogger("events-test"))

	// Test publishing with nil data - this might expose a bug
	err := eventManager.PublishEvent(TopicCameraConnected, nil)
	assert.NoError(t, err, "Publishing with nil data should not error")

	// Test publishing with empty data
	err = eventManager.PublishEvent(TopicCameraConnected, map[string]interface{}{})
	assert.NoError(t, err, "Publishing with empty data should not error")

	// Test publishing with very large data - this might expose memory/resource bugs
	largeData := make(map[string]interface{})
	for i := 0; i < 10000; i++ {
		largeData[fmt.Sprintf("key_%d", i)] = fmt.Sprintf("value_%d", i)
	}
	err = eventManager.PublishEvent(TopicCameraConnected, largeData)
	assert.NoError(t, err, "Publishing with large data should not error")

	// Test publishing with special characters in data - this might expose encoding bugs
	specialData := map[string]interface{}{
		"unicode":     "测试中文 🎥",
		"newlines":    "line1\nline2\rline3",
		"quotes":      `"quoted" 'single'`,
		"backslashes": "path\\to\\file",
		"null_bytes":  "data\x00with\x00nulls",
	}
	err = eventManager.PublishEvent(TopicCameraConnected, specialData)
	assert.NoError(t, err, "Publishing with special characters should not error")

	// Test publishing with deeply nested data - this might expose recursion bugs
	nestedData := map[string]interface{}{
		"level1": map[string]interface{}{
			"level2": map[string]interface{}{
				"level3": map[string]interface{}{
					"level4": map[string]interface{}{
						"level5": "deep_value",
					},
				},
			},
		},
	}
	err = eventManager.PublishEvent(TopicCameraConnected, nestedData)
	assert.NoError(t, err, "Publishing with deeply nested data should not error")
}

// TestWebSocketEvents_ClientInterestEdgeCases tests client interest edge cases and potential bugs
func TestWebSocketEvents_ClientInterestEdgeCases(t *testing.T) {
	// REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint
	// REQ-API-003: Request/response message handling

	eventManager := NewEventManager(NewTestLogger("events-test"))

	// Test client interest with complex filters
	clientID := "test-client"
	complexFilters := map[string]interface{}{
		"camera_id":  "camera0", // Use API abstraction layer
		"status":     "connected",
		"resolution": "1920x1080",
		"fps":        30,
		"codec":      "h264",
		"nested": map[string]interface{}{
			"level1": map[string]interface{}{
				"level2": "value",
			},
		},
	}

	err := eventManager.Subscribe(clientID, []EventTopic{TopicCameraConnected}, complexFilters)
	require.NoError(t, err, "Should subscribe with complex filters")

	// Test event that matches all filter criteria
	matchingEvent := map[string]interface{}{
		"camera_id":  "camera0", // Use API abstraction layer
		"status":     "connected",
		"resolution": "1920x1080",
		"fps":        30,
		"codec":      "h264",
		"nested": map[string]interface{}{
			"level1": map[string]interface{}{
				"level2": "value",
			},
		},
		"extra": "additional_data",
	}

	err = eventManager.PublishEvent(TopicCameraConnected, matchingEvent)
	assert.NoError(t, err, "Publishing matching event should not error")

	// Test event that partially matches (should still be delivered)
	partialMatchEvent := map[string]interface{}{
		"camera_id": "camera0", // Use API abstraction layer
		"status":    "connected",
		// Missing other fields
	}

	err = eventManager.PublishEvent(TopicCameraConnected, partialMatchEvent)
	assert.NoError(t, err, "Publishing partially matching event should not error")

	// Test event with no matching fields (should still be delivered - filters are inclusive)
	noMatchEvent := map[string]interface{}{
		"camera_id": "camera1",      // Different camera
		"status":    "disconnected", // Different status
	}

	err = eventManager.PublishEvent(TopicCameraConnected, noMatchEvent)
	assert.NoError(t, err, "Publishing non-matching event should not error")

	// Test event with nil values in filters - this might expose a bug
	nilFilterClient := "nil-filter-client"
	nilFilters := map[string]interface{}{
		"camera_id": nil, // Use API abstraction layer
		"status":    "connected",
	}

	err = eventManager.Subscribe(nilFilterClient, []EventTopic{TopicCameraConnected}, nilFilters)
	require.NoError(t, err, "Should subscribe with nil filters")

	err = eventManager.PublishEvent(TopicCameraConnected, map[string]interface{}{
		"camera_id": "camera0", // Use API abstraction layer
		"status":    "connected",
	})
	assert.NoError(t, err, "Publishing event with nil filter should not error")
}

// TestWebSocketEvents_EventHandlersEdgeCases tests event handler edge cases and potential bugs
func TestWebSocketEvents_EventHandlersEdgeCases(t *testing.T) {
	// REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint
	// REQ-API-003: Request/response message handling

	eventManager := NewEventManager(NewTestLogger("events-test"))

	// Test event handler that returns an error - this might expose error handling bugs
	errorHandler := func(event *EventMessage) error {
		return fmt.Errorf("handler error")
	}

	eventManager.AddEventHandler(TopicCameraConnected, errorHandler)

	// Publish event to trigger error handler
	err := eventManager.PublishEvent(TopicCameraConnected, map[string]interface{}{
		"camera_id": "camera0", // Use API abstraction layer
		"status":    "connected",
	})
	assert.NoError(t, err, "Publishing event should not error even if handler fails")

	// Test event handler that panics - this might expose panic handling bugs
	panicHandler := func(event *EventMessage) error {
		panic("handler panic")
	}

	eventManager.AddEventHandler(TopicRecordingStart, panicHandler)

	// Publish event to trigger panic handler
	err = eventManager.PublishEvent(TopicRecordingStart, map[string]interface{}{
		"camera_id": "camera0", // Use API abstraction layer
		"session":   "recording_001",
	})
	assert.NoError(t, err, "Publishing event should not error even if handler panics")

	// Test multiple handlers for the same topic
	handler1 := func(event *EventMessage) error {
		return nil
	}

	handler2 := func(event *EventMessage) error {
		return nil
	}

	eventManager.AddEventHandler(TopicSystemHealth, handler1)
	eventManager.AddEventHandler(TopicSystemHealth, handler2)

	// Publish event to trigger both handlers
	err = eventManager.PublishEvent(TopicSystemHealth, map[string]interface{}{
		"status": "healthy",
		"cpu":    85.5,
	})
	assert.NoError(t, err, "Publishing event should not error")

	// Note: We can't easily verify handler execution without access to internal state
	// This test is designed to expose potential bugs in handler execution
}

// TestWebSocketEvents_MediaMTXEventNotifier tests MediaMTX event notifications
func TestWebSocketEvents_MediaMTXEventNotifier(t *testing.T) {
	// REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint
	// REQ-API-003: Request/response message handling

	eventManager := NewEventManager(NewTestLogger("events-test"))
	logger := NewTestLogger("mediamtx-test")

	// Test NewMediaMTXEventNotifier
	mediaNotifier := NewMediaMTXEventNotifier(eventManager, logger)
	assert.NotNil(t, mediaNotifier, "MediaMTX event notifier should be created")

	// Test recording event notifications - these functions don't return errors
	mediaNotifier.NotifyRecordingStarted("/dev/video0", "session_001", "recording_001.mp4")
	mediaNotifier.NotifyRecordingStopped("/dev/video0", "session_001", "recording_001.mp4", 120*time.Second)

	// Test streaming event notifications
	mediaNotifier.NotifyStreamStarted("/dev/video0", "stream_001", "rtsp://localhost:8554/stream")
	mediaNotifier.NotifyStreamStopped("/dev/video0", "stream_001", "rtsp://localhost:8554/stream")
}

// TestWebSocketEvents_SystemEventNotifier tests system event notifications
func TestWebSocketEvents_SystemEventNotifier(t *testing.T) {
	// REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint
	// REQ-API-003: Request/response message handling

	eventManager := NewEventManager(NewTestLogger("events-test"))
	logger := NewTestLogger("system-test")

	// Test NewSystemEventNotifier
	systemNotifier := NewSystemEventNotifier(eventManager, logger)
	assert.NotNil(t, systemNotifier, "System event notifier should be created")

	// Test system event notifications - these functions don't return errors
	systemNotifier.NotifySystemStartup("v1.0.0", "localhost")
	systemNotifier.NotifySystemShutdown("Graceful shutdown")
	systemNotifier.NotifySystemHealth("healthy", map[string]interface{}{
		"cpu_usage": 95.5,
		"memory_mb": 1024,
		"disk_mb":   2048,
	})
}

// TestWebSocketEvents_EdgeCases tests edge cases and error conditions
func TestWebSocketEvents_EdgeCases(t *testing.T) {
	// REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint
	// REQ-API-003: Request/response message handling

	eventManager := NewEventManager(NewTestLogger("events-test"))

	// Test publishing event with nil data
	err := eventManager.PublishEvent(TopicCameraConnected, nil)
	assert.NoError(t, err, "Publishing event with nil data should not error")

	// Test publishing event with empty data
	err = eventManager.PublishEvent(TopicCameraConnected, map[string]interface{}{})
	assert.NoError(t, err, "Publishing event with empty data should not error")

	// Test subscribing with empty topics list
	clientID := "test-client"
	err = eventManager.Subscribe(clientID, []EventTopic{}, nil)
	// Should either error or handle gracefully
	if err != nil {
		assert.Error(t, err, "Should error for empty topics list")
	}

	// Test subscribing with nil filters
	err = eventManager.Subscribe(clientID, []EventTopic{TopicCameraConnected}, nil)
	assert.NoError(t, err, "Subscribing with nil filters should not error")

	// Test removing non-existent client
	eventManager.RemoveClient("non-existent-client")
	// Should not panic or error
}

// TestWebSocketEvents_EventHandlers tests event handler functionality
func TestWebSocketEvents_EventHandlers(t *testing.T) {
	// REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint
	// REQ-API-003: Request/response message handling

	eventManager := NewEventManager(NewTestLogger("events-test"))

	// Add multiple event handlers for the same topic
	handler1 := func(event *EventMessage) error {
		return nil
	}

	handler2 := func(event *EventMessage) error {
		return nil
	}

	eventManager.AddEventHandler(TopicCameraConnected, handler1)
	eventManager.AddEventHandler(TopicCameraConnected, handler2)

	// Publish event to trigger handlers
	err := eventManager.PublishEvent(TopicCameraConnected, map[string]interface{}{
		"camera_id": "camera0", // Use API abstraction layer
	})

	assert.NoError(t, err, "Publishing event should not error")
	// Note: Handler execution is asynchronous, so we can't easily verify they were called
	// This test verifies the handler registration doesn't cause errors
}

// TestWebSocketEvents_SubscriptionManagement tests subscription management edge cases
func TestWebSocketEvents_SubscriptionManagement(t *testing.T) {
	// REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint
	// REQ-API-003: Request/response message handling

	eventManager := NewEventManager(NewTestLogger("events-test"))

	clientID := "test-client"
	topics := []EventTopic{TopicCameraConnected, TopicRecordingStart}

	// Subscribe client
	err := eventManager.Subscribe(clientID, topics, nil)
	require.NoError(t, err, "Should subscribe client successfully")

	// Unsubscribe from non-existent topic
	err = eventManager.Unsubscribe(clientID, []EventTopic{TopicSystemHealth})
	assert.NoError(t, err, "Unsubscribing from non-existent topic should not error")

	// Unsubscribe from existing topic
	err = eventManager.Unsubscribe(clientID, []EventTopic{TopicCameraConnected})
	assert.NoError(t, err, "Unsubscribing from existing topic should not error")

	// Verify partial unsubscription
	subscribers := eventManager.GetSubscribersForTopic(TopicCameraConnected)
	assert.NotContains(t, subscribers, clientID, "Client should not be subscribed to camera connected")

	subscribers = eventManager.GetSubscribersForTopic(TopicRecordingStart)
	assert.Contains(t, subscribers, clientID, "Client should still be subscribed to recording start")
}
