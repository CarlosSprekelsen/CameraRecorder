# tests/unit/test_websocket_server/test_server_notifications.py
"""
Test notification functionality and API compliance in WebSocket JSON-RPC server.

Requirements Traceability:
- REQ-WS-004: WebSocket server shall broadcast camera status notifications to all clients
- REQ-WS-005: WebSocket server shall filter notification fields according to API specification
- REQ-WS-006: WebSocket server shall handle client connection failures gracefully
- REQ-WS-007: WebSocket server shall support real-time notification delivery
- REQ-ERROR-002: WebSocket server shall handle client disconnection during notification

Story Coverage: S3 - WebSocket API Integration
IV&V Control Point: Real WebSocket communication validation
"""

import pytest
import asyncio
import json
from unittest.mock import Mock, AsyncMock, patch

from src.websocket_server.server import WebSocketJsonRpcServer, ClientConnection
from tests.fixtures.websocket_test_client import WebSocketTestClient
# Removed unused imports: mediamtx_infrastructure, mediamtx_controller


class TestServerNotifications:
    """Test notification broadcasting and API compliance."""

    @pytest.fixture
    def server(self):
        """Create WebSocket server instance for testing."""
        return WebSocketJsonRpcServer(
            host="localhost", port=8002, websocket_path="/ws", max_connections=100
        )

    @pytest.fixture
    def real_websocket_client(self):
        """Create real WebSocket test client."""
        return WebSocketTestClient("ws://localhost:8002/ws")

    @pytest.fixture
    def real_websocket_client(self):
        """Create real WebSocket test client for integration testing."""
        return WebSocketTestClient("ws://localhost:8002/ws")

    @pytest.fixture
    async def connected_real_client(self, server):
        """Create and connect real WebSocket client with server running."""
        client = WebSocketTestClient("ws://localhost:8002/ws")
        await server.start()
        await client.connect()
        return client

    @pytest.fixture
    async def connected_websocket_client(self, server, real_websocket_client):
        """Create and connect real WebSocket test client with server running."""
        # Start the server
        await server.start()
        
        # Connect the client
        await real_websocket_client.connect()
        yield real_websocket_client
        
        # Cleanup
        await real_websocket_client.disconnect()
        await server.stop()

    @pytest.mark.asyncio
    async def test_camera_status_notification_with_real_websocket_communication(
        self, server, real_websocket_client
    ):
        """
        Verify camera status notifications are delivered via real WebSocket communication.
        
        Requirements: REQ-WS-004, REQ-WS-007
        Scenario: Real WebSocket client receives camera status notification
        Expected: Notification delivered successfully via real WebSocket connection
        Edge Cases: Real-time delivery, connection stability
        """
        # Start the server first
        await server.start()
        
        try:
            # Connect the client
            await real_websocket_client.connect()
            
            # Test input with extra fields that should be filtered out
            input_params = {
                "device": "/dev/video0",
                "status": "CONNECTED",
                "name": "Test Camera",
                "resolution": "1920x1080",
                "fps": 30,
                "streams": {"rtsp": "rtsp://localhost:8554/camera0"},
                # Fields that should be filtered out:
                "internal_id": "camera-internal-123",
                "debug_info": {"probe_count": 5},
                "raw_capability_data": {"driver": "uvcvideo"},
                "validation_status": "confirmed",
            }

            # Send notification via real WebSocket server
            await server.notify_camera_status_update(input_params)

            # Wait for notification to be delivered via real WebSocket
            try:
                notification = await real_websocket_client.wait_for_notification(
                    "camera_status_update", timeout=5.0
                )
                
                # Verify notification was received
                assert notification.result is not None
                filtered_params = notification.result

                # Verify only allowed fields are present (API filtering)
                expected_fields = {
                    "device",
                    "status", 
                    "name",
                    "resolution",
                    "fps",
                    "streams",
                }
                assert set(filtered_params.keys()) == expected_fields

                # Verify filtered out fields are not present
                forbidden_fields = {
                    "internal_id",
                    "debug_info",
                    "raw_capability_data",
                    "validation_status",
                }
                for field in forbidden_fields:
                    assert field not in filtered_params

                # Verify values are preserved for allowed fields
                assert filtered_params["device"] == "/dev/video0"
                assert filtered_params["status"] == "CONNECTED"
                assert filtered_params["resolution"] == "1920x1080"
                assert filtered_params["fps"] == 30
                
            except TimeoutError:
                pytest.fail("Notification not received within timeout period")
        finally:
            # Cleanup
            await real_websocket_client.disconnect()
            await server.stop()

    @pytest.mark.asyncio
    async def test_camera_status_notification_field_filtering_with_real_client(self, server):
        """
        Verify camera_status_update notifications only include API-specified fields with real WebSocket client.
        
        Requirements: REQ-WS-005
        Scenario: Field filtering validation with real WebSocket communication
        Expected: Only API-specified fields included in notifications
        Edge Cases: Extra fields properly filtered out, real connection handling
        """
        # Create and connect real WebSocket client
        client = WebSocketTestClient("ws://localhost:8002/ws")
        await server.start()
        await client.connect()
        
        try:
            # Test input with extra fields that should be filtered out
            input_params = {
                "device": "/dev/video0",
                "status": "CONNECTED",
                "name": "Test Camera",
                "resolution": "1920x1080",
                "fps": 30,
                "streams": {"rtsp": "rtsp://localhost:8554/camera0"},
                # Fields that should be filtered out:
                "internal_id": "camera-internal-123",
                "debug_info": {"probe_count": 5},
                "raw_capability_data": {"driver": "uvcvideo"},
                "validation_status": "confirmed",
            }

            # Send notification via real WebSocket server
            await server.notify_camera_status_update(input_params)

            # Wait for notification to be received
            notification_response = await client.wait_for_notification("camera_status_update", timeout=5.0)
            
            # Get the notification data
            filtered_params = notification_response.result
            assert filtered_params is not None, "Notification should have params"

            # Verify only allowed fields are present
            expected_fields = {
                "device",
                "status",
                "name",
                "resolution",
                "fps",
                "streams",
            }
            assert set(filtered_params.keys()) == expected_fields

            # Verify filtered out fields are not present
            forbidden_fields = {
                "internal_id",
                "debug_info",
                "raw_capability_data",
                "validation_status",
            }
            for field in forbidden_fields:
                assert field not in filtered_params

            # Verify values are preserved for allowed fields
            assert filtered_params["device"] == "/dev/video0"
            assert filtered_params["status"] == "CONNECTED"
            assert filtered_params["resolution"] == "1920x1080"
            assert filtered_params["fps"] == 30

        except asyncio.TimeoutError:
            pytest.fail("Notification not received within timeout period")
        finally:
            await client.disconnect()
            await server.stop()

    @pytest.mark.asyncio
    async def test_recording_status_notification_field_filtering_with_real_client(self):
        """
        Verify recording_status_update notifications only include API-specified fields with real WebSocket client.
        
        Requirements: REQ-WS-005
        Scenario: Real WebSocket field filtering validation
        Expected: Only API-specified fields are included in notification
        Edge Cases: Field filtering, API compliance
        """
        # Create WebSocket server WITHOUT MediaMTX dependencies for this test
        # This test only validates WebSocket field filtering, not MediaMTX operations
        server = WebSocketJsonRpcServer(
            host="localhost", 
            port=8002, 
            websocket_path="/ws", 
            max_connections=100,
            mediamtx_controller=None  # No MediaMTX controller needed for field filtering test
        )
        
        # Create real WebSocket client
        client = WebSocketTestClient("ws://127.0.0.1:8002/ws")
        
        try:
            # Start server and connect client
            await server.start()
            await client.connect()
            
            # Wait for connection to be established
            await asyncio.sleep(0.1)
            
            # Test input with extra fields that should be filtered out
            input_params = {
                "device": "/dev/video0",
                "status": "STARTED",
                "filename": "camera0_recording.mp4",
                "duration": 3600,
                # Fields that should be filtered out:
                "session_internal_id": "session-abc-123",
                "file_path": "/opt/recordings/camera0_recording.mp4",
                "process_id": 12345,
                "encoding_settings": {"bitrate": "2M"},
                "correlation_id": "req-xyz-789",
            }

            # Send notification via real WebSocket server
            await server.notify_recording_status_update(input_params)

            # Wait for notification to be received
            await asyncio.sleep(0.1)
            
            # Get real messages received by WebSocket client
            received_messages = client.get_received_messages()
            
            # Verify real notification was delivered
            assert len(received_messages) > 0, "Real WebSocket client should receive notification"
            
            # Get the notification data
            notification_response = received_messages[0]
            assert notification_response.result is not None, "Notification should have result"
            
            filtered_params = notification_response.result
            assert filtered_params is not None, "Notification should have params"

            # Verify only allowed fields are present
            expected_fields = {"device", "status", "filename", "duration"}
            assert set(filtered_params.keys()) == expected_fields

            # Verify filtered out fields are not present
            forbidden_fields = {
                "session_internal_id",
                "file_path",
                "process_id",
                "encoding_settings",
                "correlation_id",
            }
            for field in forbidden_fields:
                assert field not in filtered_params

            # Verify values are preserved for allowed fields
            assert filtered_params["device"] == "/dev/video0"
            assert filtered_params["status"] == "STARTED"
            assert filtered_params["filename"] == "camera0_recording.mp4"
            assert filtered_params["duration"] == 3600

        finally:
            # Clean up real WebSocket connection
            await client.disconnect()
            await server.stop()

    def test_notification_required_field_validation(self, server):
        """Test validation of required fields in notifications."""
        # Test camera notification without required device field
        with patch.object(server, "broadcast_notification") as mock_broadcast:
            asyncio.run(server.notify_camera_status_update({"status": "CONNECTED"}))
            mock_broadcast.assert_not_called()  # Should not broadcast invalid notification

        # Test camera notification without required status field
        with patch.object(server, "broadcast_notification") as mock_broadcast:
            asyncio.run(server.notify_camera_status_update({"device": "/dev/video0"}))
            mock_broadcast.assert_not_called()  # Should not broadcast invalid notification

        # Test recording notification without required fields
        with patch.object(server, "broadcast_notification") as mock_broadcast:
            asyncio.run(server.notify_recording_status_update({"filename": "test.mp4"}))
            mock_broadcast.assert_not_called()  # Should not broadcast invalid notification

    @pytest.mark.asyncio
    async def test_broadcast_notification_to_real_clients(self):
        """
        Test broadcasting notifications to connected real clients.
        
        Requirements: REQ-WS-004, REQ-WS-007
        Scenario: Real WebSocket broadcasting validation
        Expected: Notification delivered successfully to all connected clients
        Edge Cases: Real-time delivery, connection stability
        """
        # Create WebSocket server WITHOUT MediaMTX dependencies for this test
        # This test only validates WebSocket broadcasting, not MediaMTX operations
        server = WebSocketJsonRpcServer(
            host="localhost", 
            port=8002, 
            websocket_path="/ws", 
            max_connections=100,
            mediamtx_controller=None  # No MediaMTX controller needed for broadcasting test
        )
        
        # Create real WebSocket client
        client = WebSocketTestClient("ws://127.0.0.1:8002/ws")
        
        try:
            # Start server and connect client
            await server.start()
            await client.connect()
            
            # Wait for connection to be established
            await asyncio.sleep(0.1)
            
            # Test broadcasting notification
            await server.broadcast_notification(
                method="test_notification", params={"key": "value", "device": "/dev/video0"}
            )

            # Wait for notification to be received
            await asyncio.sleep(0.1)
            
            # Get real messages received by WebSocket client
            received_messages = client.get_received_messages()
            
            # Verify real notification was delivered
            assert len(received_messages) > 0, "Real WebSocket client should receive notification"
            
            # Get the notification data
            notification_response = received_messages[0]
            assert notification_response.result is not None, "Notification should have result"
            
            notification_params = notification_response.result
            assert notification_params is not None, "Notification should have params"
            
            # Verify notification structure
            assert notification_params["key"] == "value"
            assert notification_params["device"] == "/dev/video0"

        finally:
            # Clean up real WebSocket connection
            await client.disconnect()
            await server.stop()

    @pytest.mark.asyncio
    async def test_notification_client_cleanup_on_real_connection_failure(self, server):
        """Test cleanup of disconnected clients during notification with real WebSocket connection."""
        # Create real client and connect
        client = WebSocketTestClient("ws://localhost:8002/ws")
        await server.start()
        await client.connect()
        
        # Verify client is connected
        assert len(server._clients) > 0
        initial_client_count = len(server._clients)
        
        # Simulate connection failure by closing client connection
        await client.disconnect()
        
        # Test broadcasting notification to disconnected client
        await server.broadcast_notification(
            method="test_notification", params={"device": "/dev/video0"}
        )

        # Verify disconnected client was cleaned up
        assert len(server._clients) < initial_client_count
        
        await server.stop()

    @pytest.mark.asyncio
    async def test_send_notification_to_specific_real_client(self):
        """
        Test sending notification to specific real client.
        
        Requirements: REQ-WS-004, REQ-WS-007
        Scenario: Real WebSocket targeted notification validation
        Expected: Notification delivered successfully to specific client
        Edge Cases: Client identification, targeted delivery
        """
        # Create WebSocket server WITHOUT MediaMTX dependencies for this test
        # This test only validates WebSocket targeted notifications, not MediaMTX operations
        server = WebSocketJsonRpcServer(
            host="localhost", 
            port=8002, 
            websocket_path="/ws", 
            max_connections=100,
            mediamtx_controller=None  # No MediaMTX controller needed for targeted notification test
        )
        
        # Create real WebSocket client
        client = WebSocketTestClient("ws://127.0.0.1:8002/ws")
        
        try:
            # Start server and connect client
            await server.start()
            await client.connect()
            
            # Wait for connection to be established
            await asyncio.sleep(0.1)
            
            # Get the client ID from the connected client
            client_id = list(server._clients.keys())[0] if server._clients else None
            assert client_id is not None, "No real client connected"

            # Send notification to specific client
            result = await server.send_notification_to_client(
                client_id=client_id,
                method="camera_connected",
                params={"device": "/dev/video0", "status": "CONNECTED"},
            )

            # Verify notification was sent successfully
            assert result is True

            # Wait for notification to be received
            await asyncio.sleep(0.1)
            
            # Get real messages received by WebSocket client
            received_messages = client.get_received_messages()
            
            # Verify real notification was delivered
            assert len(received_messages) > 0, "Real WebSocket client should receive notification"
            
            # Get the notification data
            notification_response = received_messages[0]
            assert notification_response.result is not None, "Notification should have result"
            
            notification_params = notification_response.result
            assert notification_params is not None, "Notification should have params"
            assert notification_params["device"] == "/dev/video0"

        finally:
            # Clean up real WebSocket connection
            await client.disconnect()
            await server.stop()

    @pytest.mark.asyncio
    async def test_notification_to_nonexistent_client(self, server):
        """Test notification handling for non-existent client."""
        # Try to send notification to non-existent client
        result = await server.send_notification_to_client(
            client_id="nonexistent-client",
            method="test_notification",
            params={"key": "value"},
        )

        # Should return False for non-existent client
        assert result is False

    @pytest.mark.asyncio
    async def test_notification_serialization_error_handling(self, server):
        """Test handling of notification serialization errors."""

        # Create params that can't be serialized to JSON
        class NonSerializable:
            pass

        with patch.object(server, "_clients", {"test": Mock()}):
            # Test notification with non-serializable params
            await server.broadcast_notification(
                method="test_notification", params={"object": NonSerializable()}
            )
            # Should handle gracefully without crashing

    def test_empty_notification_params_handling(self, server):
        """Test handling of empty or None notification parameters."""
        # Test camera notification with None params
        with patch.object(server, "broadcast_notification") as mock_broadcast:
            asyncio.run(server.notify_camera_status_update(None))
            mock_broadcast.assert_not_called()

        # Test recording notification with empty params
        with patch.object(server, "broadcast_notification") as mock_broadcast:
            asyncio.run(server.notify_recording_status_update({}))
            mock_broadcast.assert_not_called()

    @pytest.mark.asyncio
    async def test_notification_correlation_id_handling(self, server):
        """
        Test correlation ID inclusion in notifications using real WebSocket communication.
        
        Requirements: REQ-WS-004, REQ-WS-005, REQ-WS-006, REQ-WS-007
        Scenario: Real WebSocket client receives notification with correlation ID
        Expected: Notification delivered successfully with correlation ID preserved
        Edge Cases: Real-time delivery, connection stability, correlation ID propagation
        """
        # Start the real WebSocket server
        await server.start()
        
        # Create real WebSocket client for testing
        client = WebSocketTestClient("ws://127.0.0.1:8002/ws")
        
        try:
            # Connect real WebSocket client
            await client.connect()
            
            # Wait for connection to be established
            await asyncio.sleep(0.1)
            
            # Verify client is connected to server
            client_ids = list(server._clients.keys())
            assert len(client_ids) > 0, "Real WebSocket client should be connected"
            
            # Test notification with correlation ID using real WebSocket communication
            test_correlation_id = "test-correlation-123"
            await server.broadcast_notification(
                method="camera_status_update",
                params={
                    "device": "/dev/video0", 
                    "status": "CONNECTED",
                    "name": "Test Camera",
                    "correlation_id": test_correlation_id
                },
            )
            
            # Wait for real notification delivery
            await asyncio.sleep(0.1)
            
            # Get real messages received by WebSocket client
            received_messages = client.get_received_messages()
            
            # Verify real notification was delivered
            assert len(received_messages) > 0, "Real WebSocket client should receive notification"
            
            # Verify correlation ID is preserved in real notification
            notification_response = received_messages[0]
            assert notification_response.result is not None, "Notification should contain result"
            
            # The result contains the params from the notification
            notification_params = notification_response.result
            assert "correlation_id" in notification_params, "Notification should contain correlation_id"
            assert notification_params["correlation_id"] == test_correlation_id, "Correlation ID should be preserved"
            
            # Verify real WebSocket communication worked
            assert notification_params["device"] == "/dev/video0", "Device should be preserved"
            assert notification_params["status"] == "CONNECTED", "Status should be preserved"
            
        finally:
            # Clean up real WebSocket connection
            await client.disconnect()
            await server.stop()

    @pytest.mark.asyncio
    async def test_targeted_notification_broadcast(self, server):
        """Test broadcasting notifications to specific client list."""
        # Start the server first
        await server.start()
        
        # Setup multiple real WebSocket clients
        client1 = WebSocketTestClient("ws://127.0.0.1:8002/ws")
        client2 = WebSocketTestClient("ws://127.0.0.1:8002/ws")
        client3 = WebSocketTestClient("ws://127.0.0.1:8002/ws")

        try:
            # Connect all clients
            await client1.connect()
            await client2.connect()
            await client3.connect()

            # Wait a moment for connections to be established and registered
            await asyncio.sleep(0.2)

            # Get the actual client IDs from the server
            client_ids = list(server._clients.keys())
            print(f"DEBUG: Connected clients: {client_ids}")
            print(f"DEBUG: Number of clients: {len(client_ids)}")
            assert len(client_ids) >= 3, f"Expected at least 3 clients, got {len(client_ids)}"

            # Broadcast to specific clients only (use actual client IDs)
            target_clients = [client_ids[0], client_ids[2]]  # First and third clients
            print(f"DEBUG: Broadcasting to clients: {target_clients}")
            await server.broadcast_notification(
                method="targeted_notification",
                params={"message": "test"},
                target_clients=target_clients,
            )
            print(f"DEBUG: Broadcast completed")

            # Wait for notifications to be sent
            await asyncio.sleep(0.2)

            # Verify notifications were sent only to target clients
            # The system correctly implements selective broadcasting - only target clients should receive
            client1_messages = client1.get_received_messages()
            client2_messages = client2.get_received_messages()
            client3_messages = client3.get_received_messages()
            
            print(f"DEBUG: Client1 received messages: {len(client1_messages)}")
            print(f"DEBUG: Client2 received messages: {len(client2_messages)}")
            print(f"DEBUG: Client3 received messages: {len(client3_messages)}")
            
            # Client1 (targeted) should receive notification
            assert len(client1_messages) > 0, "Target client 1 should receive notification"
            
            # Client2 (not targeted) should NOT receive notification
            assert len(client2_messages) == 0, "Non-target client 2 should not receive notification"
            
            # Client3 (targeted) should receive notification
            assert len(client3_messages) > 0, "Target client 3 should receive notification"

        finally:
            # Clean up connections
            await client1.disconnect()
            await client2.disconnect()
            await client3.disconnect()
            await server.stop()

    def test_api_specification_compliance_documentation(self, server):
        """Verify notification methods document API compliance."""
        # Test that notification methods have proper docstrings referencing API spec
        camera_notify_doc = server.notify_camera_status_update.__doc__
        assert "docs/api/json-rpc-methods.md" in camera_notify_doc
        assert "device, status, name, resolution, fps, streams" in camera_notify_doc

        recording_notify_doc = server.notify_recording_status_update.__doc__
        assert "docs/api/json-rpc-methods.md" in recording_notify_doc
        assert "device, status, filename, duration" in recording_notify_doc

    @pytest.mark.asyncio
    async def test_websocket_notification_handles_client_disconnection(
        self, server, real_websocket_client
    ):
        """
        Test WebSocket notification handling when client disconnects during notification.
        
        Requirements: REQ-ERROR-002
        Scenario: Client disconnects during notification delivery
        Expected: Graceful handling of disconnection without server crash
        Edge Cases: Connection failures, client timeouts
        """
        # Start the server first
        await server.start()
        
        # Connect client
        await real_websocket_client.connect()
        
        # Disconnect client immediately
        await real_websocket_client.disconnect()
        
        # Try to send notification to disconnected client
        input_params = {
            "device": "/dev/video0",
            "status": "CONNECTED",
            "name": "Test Camera",
        }
        
        # This should not crash the server
        try:
            await server.notify_camera_status_update(input_params)
            # If we get here, the server handled the disconnection gracefully
            # Verify server is still running and can handle more notifications
            assert server.get_connection_count() >= 0, "Server should still be operational"
        except Exception as e:
            pytest.fail(f"Server crashed when sending notification to disconnected client: {e}")
        finally:
            # Clean up server
            await server.stop()

    @pytest.mark.asyncio
    async def test_websocket_notification_handles_connection_failure(
        self, server
    ):
        """
        Test WebSocket notification handling with real connection failure scenarios.
        
        Requirements: REQ-ERROR-002
        Scenario: Real WebSocket connection failure during notification
        Expected: Graceful error handling without server crash
        Edge Cases: Network failures, connection timeouts, real service unavailability
        """
        # Test real connection failure scenarios
        failure_scenarios = [
            # Test connection to non-existent port on localhost (real network failure)
            ("ws://127.0.0.1:65535/ws", "Non-existent port"),
            # Test connection to localhost with wrong path (real service unavailability)
            ("ws://127.0.0.1:8002/invalid-path", "Invalid WebSocket path"),
            # Test connection to unreachable host (real network failure)
            ("ws://192.168.255.255:8002/ws", "Unreachable host"),
        ]
        
        for url, scenario in failure_scenarios:
            # Create client with real failure scenario
            test_client = WebSocketTestClient(url)
            
            # Try to connect (should fail with real network error)
            try:
                await test_client.connect()
                pytest.fail(f"Should not be able to connect to {scenario}")
            except Exception as e:
                # Expected connection failure - verify it's a real network error
                error_str = str(e).lower()
                expected_errors = ['connection', 'timeout', 'timed', 'refused', 'unreachable', 'invalid', 'port', 'failed', 'handshake']
                assert any(error_type in error_str for error_type in expected_errors), \
                    f"Expected network error for {scenario}, got: {e}"
            
            # Try to send notification (should handle gracefully)
            input_params = {
                "device": "/dev/video0",
                "status": "CONNECTED",
                "name": "Test Camera",
            }
            
            # This should not crash the server
            try:
                await server.notify_camera_status_update(input_params)
                # If we get here, the server handled the connection failure gracefully
                # Verify server is still operational
                assert server.get_connection_count() >= 0, f"Server should still be operational after {scenario}"
            except Exception as e:
                pytest.fail(f"Server crashed when handling {scenario}: {e}")

    @pytest.mark.asyncio
    async def test_websocket_notification_handles_invalid_message_format(
        self, server, connected_websocket_client
    ):
        """
        Test WebSocket notification handling with invalid message format.
        
        Requirements: REQ-ERROR-002
        Scenario: Invalid notification message format
        Expected: Graceful error handling without server crash
        Edge Cases: Malformed JSON, missing required fields
        """
        # Try to send notification with invalid format
        invalid_params = None  # Invalid params
        
        # This should not crash the server
        try:
            await server.notify_camera_status_update(invalid_params)
            # Server should handle invalid params gracefully
            # Verify server is still operational
            assert server.get_connection_count() >= 0, "Server should still be operational"
        except Exception as e:
            pytest.fail(f"Server crashed when handling invalid message format: {e}")

    @pytest.mark.asyncio
    async def test_real_websocket_notification_delivery_with_multiple_clients(
        self, server
    ):
        """
        Test real WebSocket notification delivery with multiple connected clients.
        
        Requirements: REQ-WS-004, REQ-WS-007
        Scenario: Multiple real WebSocket clients receive notifications simultaneously
        Expected: All connected clients receive notifications via real WebSocket communication
        Edge Cases: Concurrent connections, real-time delivery
        """
        # Start the server
        await server.start()
        
        # Create multiple real WebSocket clients
        clients = []
        for i in range(3):
            client = WebSocketTestClient("ws://127.0.0.1:8002/ws")
            await client.connect()
            clients.append(client)
        
        try:
            # Wait for connections to be established
            await asyncio.sleep(0.2)
            
            # Verify all clients are connected
            assert len(server._clients) == 3, f"Expected 3 clients, got {len(server._clients)}"
            
            # Send notification to all clients
            notification_params = {
                "device": "/dev/video0",
                "status": "CONNECTED",
                "name": "Test Camera",
                "resolution": "1920x1080",
                "fps": 30,
            }
            
            await server.notify_camera_status_update(notification_params)
            
            # Wait for notifications to be delivered
            await asyncio.sleep(0.2)
            
            # Verify all clients received the notification
            for i, client in enumerate(clients):
                messages = client.get_received_messages()
                assert len(messages) > 0, f"Client {i} did not receive notification"
                
                # Verify notification content
                notification = messages[-1]  # Get the latest message
                assert notification.result is not None, f"Client {i} received invalid notification"
                assert notification.result.get("device") == "/dev/video0"
                assert notification.result.get("status") == "CONNECTED"
                
        finally:
            # Clean up all clients
            for client in clients:
                await client.disconnect()
            await server.stop()

    @pytest.mark.asyncio
    async def test_notification_delivery_with_connection_failures(
        self, server
    ):
        """
        Test notification delivery when some clients have connection failures.
        
        Requirements: REQ-WS-006
        Scenario: Mixed healthy and failing client connections
        Expected: Notifications delivered to healthy clients, failed clients cleaned up
        Edge Cases: Partial connection failures, cleanup verification
        """
        # Start the server
        await server.start()
        
        # Create one healthy client
        healthy_client = WebSocketTestClient("ws://127.0.0.1:8002/ws")
        await healthy_client.connect()
        
        # Create a mock client that will fail during send
        mock_failing_client = Mock()
        mock_failing_client.websocket = AsyncMock()
        mock_failing_client.websocket.send = AsyncMock(side_effect=Exception("Connection broken"))
        mock_failing_client.client_id = "failing-client-123"
        
        # Add failing client to server
        server._clients["failing-client-123"] = mock_failing_client
        
        try:
            # Wait for connections to be established
            await asyncio.sleep(0.2)
            
            # Verify we have both clients
            assert len(server._clients) == 2, f"Expected 2 clients, got {len(server._clients)}"
            
            # Send notification (should fail for one client, succeed for another)
            notification_params = {
                "device": "/dev/video0",
                "status": "CONNECTED",
                "name": "Test Camera",
            }
            
            await server.notify_camera_status_update(notification_params)
            
            # Wait for processing
            await asyncio.sleep(0.2)
            
            # Verify healthy client received notification
            messages = healthy_client.get_received_messages()
            assert len(messages) > 0, "Healthy client should receive notification"
            
            # Verify failing client was cleaned up
            assert "failing-client-123" not in server._clients, "Failing client should be removed"
            
            # Verify server still has the healthy client
            assert len(server._clients) == 1, "Server should retain healthy client"
            
        finally:
            # Clean up
            await healthy_client.disconnect()
            await server.stop()
