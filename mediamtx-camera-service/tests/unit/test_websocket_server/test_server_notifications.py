# tests/unit/test_websocket_server/test_server_notifications.py
"""
Test notification functionality and API compliance in WebSocket JSON-RPC server.

Covers notification field filtering, API specification compliance,
and broadcast functionality.
"""

import pytest
import asyncio
import json
from unittest.mock import Mock, AsyncMock, patch

from src.websocket_server.server import WebSocketJsonRpcServer, ClientConnection


class TestServerNotifications:
    """Test notification broadcasting and API compliance."""

    @pytest.fixture
    def server(self):
        """Create WebSocket server instance for testing."""
        return WebSocketJsonRpcServer(
            host="localhost", port=8002, websocket_path="/ws", max_connections=100
        )

    @pytest.fixture
    def mock_client(self):
        """Create mock WebSocket client."""
        mock_websocket = Mock()
        mock_websocket.open = True
        mock_websocket.send = AsyncMock()

        client = ClientConnection(mock_websocket, "test-client-123")
        client.authenticated = True
        return client

    def test_camera_status_notification_field_filtering(self, server):
        """Verify camera_status_update notifications only include API-specified fields."""
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
        # Setup multiple mock clients
        client1 = Mock()
        client1.websocket.open = True
        client1.websocket.send = AsyncMock()

        client2 = Mock()
        client2.websocket.open = True
        client2.websocket.send = AsyncMock()

        client3 = Mock()
        client3.websocket.open = True
        client3.websocket.send = AsyncMock()

        server._clients = {"client1": client1, "client2": client2, "client3": client3}

        # Broadcast to specific clients only
        await server.broadcast_notification(
            method="targeted_notification",
            params={"message": "test"},
            target_clients=["client1", "client3"],
        )

        # Verify only targeted clients received notification
        client1.websocket.send.assert_called_once()
        client2.websocket.send.assert_not_called()  # Should not receive
        client3.websocket.send.assert_called_once()

    def test_api_specification_compliance_documentation(self, server):
        """Verify notification methods document API compliance."""
        # Test that notification methods have proper docstrings referencing API spec
        camera_notify_doc = server.notify_camera_status_update.__doc__
        assert "docs/api/json-rpc-methods.md" in camera_notify_doc
        assert "device, status, name, resolution, fps, streams" in camera_notify_doc

        recording_notify_doc = server.notify_recording_status_update.__doc__
        assert "docs/api/json-rpc-methods.md" in recording_notify_doc
        assert "device, status, filename, duration" in recording_notify_doc
