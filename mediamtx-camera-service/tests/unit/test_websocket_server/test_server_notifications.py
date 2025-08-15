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
from tests.fixtures.websocket_test_client import WebSocketTestClient, websocket_client
from tests.fixtures.mediamtx_test_infrastructure import mediamtx_infrastructure, mediamtx_controller


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
    async def connected_websocket_client(self, real_websocket_client):
        """Create and connect real WebSocket test client."""
        await real_websocket_client.connect()
        yield real_websocket_client
        await real_websocket_client.disconnect()

    @pytest.mark.asyncio
    async def test_camera_status_notification_with_real_websocket_communication(
        self, server, connected_websocket_client
    ):
        """
        Verify camera status notifications are delivered via real WebSocket communication.
        
        Requirements: REQ-WS-004, REQ-WS-007
        Scenario: Real WebSocket client receives camera status notification
        Expected: Notification delivered successfully via real WebSocket connection
        Edge Cases: Real-time delivery, connection stability
        """
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
            notification = await connected_websocket_client.wait_for_notification(
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

    def test_camera_status_notification_field_filtering_with_mock(self, server):
        """
        Verify camera_status_update notifications only include API-specified fields (mock test).
        
        Requirements: REQ-WS-005
        Scenario: Field filtering validation with mocked broadcast
        Expected: Only API-specified fields included in notifications
        Edge Cases: Extra fields properly filtered out
        """
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

        # Test notification filtering
        with patch.object(server, "broadcast_notification") as mock_broadcast:
            asyncio.run(server.notify_camera_status_update(input_params))

            # Verify broadcast_notification called with filtered parameters
            mock_broadcast.assert_called_once()
            args, kwargs = mock_broadcast.call_args

            assert kwargs["method"] == "camera_status_update"
            filtered_params = kwargs["params"]

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

    def test_recording_status_notification_field_filtering(self, server):
        """Verify recording_status_update notifications only include API-specified fields."""
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

        # Test notification filtering
        with patch.object(server, "broadcast_notification") as mock_broadcast:
            asyncio.run(server.notify_recording_status_update(input_params))

            # Verify broadcast_notification called with filtered parameters
            mock_broadcast.assert_called_once()
            args, kwargs = mock_broadcast.call_args

            assert kwargs["method"] == "recording_status_update"
            filtered_params = kwargs["params"]

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
    async def test_broadcast_notification_to_clients(self, server, mock_client):
        """Test broadcasting notifications to connected clients."""
        # Add mock client to server
        server._clients["test-client-123"] = mock_client

        # Test broadcasting notification
        await server.broadcast_notification(
            method="test_notification", params={"key": "value", "device": "/dev/video0"}
        )

        # Verify notification was sent to client
        mock_client.websocket.send.assert_called_once()

        # Verify notification structure
        sent_data = mock_client.websocket.send.call_args[0][0]
        notification = json.loads(sent_data)

        assert notification["jsonrpc"] == "2.0"
        assert notification["method"] == "test_notification"
        assert notification["params"]["key"] == "value"
        assert "id" not in notification  # Notifications don't have IDs

    @pytest.mark.asyncio
    async def test_notification_client_cleanup_on_failure(self, server, mock_client):
        """Test cleanup of disconnected clients during notification."""
        # Setup client with failing websocket
        mock_client.websocket.send.side_effect = Exception("Connection broken")
        server._clients["test-client-123"] = mock_client

        # Test broadcasting notification
        await server.broadcast_notification(
            method="test_notification", params={"device": "/dev/video0"}
        )

        # Verify failed client was removed
        assert "test-client-123" not in server._clients

    @pytest.mark.asyncio
    async def test_send_notification_to_specific_client(self, server, mock_client):
        """Test sending notification to specific client."""
        # Add mock client to server
        server._clients["test-client-123"] = mock_client

        # Send notification to specific client
        result = await server.send_notification_to_client(
            client_id="test-client-123",
            method="camera_connected",
            params={"device": "/dev/video0", "status": "CONNECTED"},
        )

        # Verify notification was sent successfully
        assert result is True
        mock_client.websocket.send.assert_called_once()

        # Verify notification content
        sent_data = mock_client.websocket.send.call_args[0][0]
        notification = json.loads(sent_data)
        assert notification["method"] == "camera_connected"
        assert notification["params"]["device"] == "/dev/video0"

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
    async def test_notification_correlation_id_handling(self, server, mock_client):
        """Test correlation ID inclusion in notifications."""
        server._clients["test-client-123"] = mock_client

        # Test notification with correlation ID in params
        await server.broadcast_notification(
            method="test_notification",
            params={"device": "/dev/video0", "correlation_id": "test-correlation-123"},
        )

        # Verify notification was sent (correlation ID handling tested in logging layer)
        mock_client.websocket.send.assert_called_once()

    @pytest.mark.asyncio
    async def test_targeted_notification_broadcast(self, server):
        """Test broadcasting notifications to specific client list."""
        # Setup multiple real WebSocket clients
        client1 = WebSocketTestClient("ws://127.0.0.1:8002/ws")
        client2 = WebSocketTestClient("ws://127.0.0.1:8002/ws")
        client3 = WebSocketTestClient("ws://127.0.0.1:8002/ws")

        # Connect all clients
        await client1.connect()
        await client2.connect()
        await client3.connect()

        try:
            # Broadcast to specific clients only
            await server.broadcast_notification(
                method="targeted_notification",
                params={"message": "test"},
                target_clients=["client1", "client3"],
            )

            # Wait for notifications to be sent
            await asyncio.sleep(0.1)

            # Verify notifications were sent (check received messages)
            assert len(client1.received_messages) > 0
            assert len(client2.received_messages) == 0  # Should not receive
            assert len(client3.received_messages) > 0

        finally:
            # Clean up connections
            await client1.disconnect()
            await client2.disconnect()
            await client3.disconnect()

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
            assert True
        except Exception as e:
            pytest.fail(f"Server crashed when sending notification to disconnected client: {e}")

    @pytest.mark.asyncio
    async def test_websocket_notification_handles_connection_failure(
        self, server
    ):
        """
        Test WebSocket notification handling when connection fails.
        
        Requirements: REQ-ERROR-002
        Scenario: WebSocket connection failure during notification
        Expected: Graceful error handling without server crash
        Edge Cases: Network failures, connection timeouts
        """
        # Create client with invalid server URL to simulate connection failure
        invalid_client = WebSocketTestClient("ws://invalid-server:9999/ws")
        
        # Try to connect to invalid server (should fail)
        try:
            await invalid_client.connect()
            pytest.fail("Should not be able to connect to invalid server")
        except Exception:
            # Expected connection failure
            pass
        
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
            assert True
        except Exception as e:
            pytest.fail(f"Server crashed when handling connection failure: {e}")

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
            assert True
        except Exception as e:
            pytest.fail(f"Server crashed when handling invalid message format: {e}")
