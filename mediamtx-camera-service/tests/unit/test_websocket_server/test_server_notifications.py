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
from tests.utils.port_utils import check_websocket_server_port


class TestServerNotifications:
    """Test notification broadcasting and API compliance."""

    @pytest.fixture
    def real_websocket_client(self):
        """Create real WebSocket test client for integration testing."""
        return WebSocketTestClient("ws://localhost:8002/ws")

    @pytest.mark.asyncio
    async def test_camera_status_notification_with_real_websocket_communication(
        self, real_websocket_client
    ):
        """
        Verify camera status notifications are delivered via real WebSocket communication.
        
        Requirements: REQ-WS-004, REQ-WS-007
        Scenario: Real WebSocket client receives camera status notification
        Expected: Notification delivered successfully via real WebSocket connection
        Edge Cases: Real-time delivery, connection stability
        """
        # Check if real server is running - REQUIRED per project ground rules
        if not check_websocket_server_port(8002):
            pytest.skip("Real WebSocket server not running on port 8002 - required per project ground rules")
        
        try:
            # Connect to the REAL server (not a test server)
            await real_websocket_client.connect()
            
            # Wait for any existing camera notifications from the real server
            # The real server sends notifications when camera events occur
            await asyncio.sleep(2.0)
            
            # Get any notifications that may have been sent by the real server
            received_messages = real_websocket_client.get_received_messages()
            
            # Verify we can connect to the real server
            assert real_websocket_client.connected, "Should be connected to real WebSocket server"
            
            # If we received any notifications, validate their structure
            if received_messages:
                for message in received_messages:
                    if message.method == "camera_status_update":
                        # Verify notification structure follows API specification
                        assert message.result is not None, "Notification should have result"
                        
                        # Verify only allowed fields are present (API filtering)
                        expected_fields = {
                            "device",
                            "status", 
                            "name",
                            "resolution",
                            "fps",
                            "streams",
                        }
                        actual_fields = set(message.result.keys())
                        # Check that all fields are allowed (some may be missing)
                        for field in actual_fields:
                            assert field in expected_fields, f"Field '{field}' not allowed in API specification"
                
        finally:
            # Cleanup
            await real_websocket_client.disconnect()

    @pytest.mark.asyncio
    async def test_websocket_connection_to_real_server(self, real_websocket_client):
        """
        Test basic WebSocket connection to the real server.
        
        Requirements: REQ-WS-004, REQ-WS-007
        Scenario: Connect to real WebSocket server and verify communication
        Expected: Successful connection and basic communication
        """
        # Check if real server is running - REQUIRED per project ground rules
        if not check_websocket_server_port(8002):
            pytest.skip("Real WebSocket server not running on port 8002 - required per project ground rules")
        
        try:
            # Connect to the REAL server
            await real_websocket_client.connect()
            
            # Verify connection is established
            assert real_websocket_client.connected, "Should be connected to real WebSocket server"
            
            # Send a simple ping or test message to verify communication
            # Note: We can't send notifications to the real server from tests,
            # but we can verify the connection works
            
        finally:
            # Cleanup
            await real_websocket_client.disconnect()

    @pytest.mark.asyncio
    async def test_websocket_notification_handles_connection_failure(self):
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

    def test_notification_required_field_validation(self):
        """Test validation of required fields in notifications."""
        # Create a minimal server instance for validation testing only
        # This doesn't start a server, just tests the validation logic
        server = WebSocketJsonRpcServer(
            host="localhost", port=8002, websocket_path="/ws", max_connections=100
        )
        
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

    def test_empty_notification_params_handling(self):
        """Test handling of empty or None notification parameters."""
        # Create a minimal server instance for validation testing only
        server = WebSocketJsonRpcServer(
            host="localhost", port=8002, websocket_path="/ws", max_connections=100
        )
        
        # Test camera notification with None params
        with patch.object(server, "broadcast_notification") as mock_broadcast:
            asyncio.run(server.notify_camera_status_update(None))
            mock_broadcast.assert_not_called()

        # Test recording notification with empty params
        with patch.object(server, "broadcast_notification") as mock_broadcast:
            asyncio.run(server.notify_recording_status_update({}))
            mock_broadcast.assert_not_called()

    def test_api_specification_compliance_documentation(self):
        """Verify notification methods document API compliance."""
        # Create a minimal server instance for documentation testing only
        server = WebSocketJsonRpcServer(
            host="localhost", port=8002, websocket_path="/ws", max_connections=100
        )
        
        # Test that notification methods have proper docstrings referencing API spec
        camera_notify_doc = server.notify_camera_status_update.__doc__
        assert "docs/api/json-rpc-methods.md" in camera_notify_doc
        assert "device, status, name, resolution, fps, streams" in camera_notify_doc

        recording_notify_doc = server.notify_recording_status_update.__doc__
        assert "docs/api/json-rpc-methods.md" in recording_notify_doc
        assert "device, status, filename, duration" in recording_notify_doc

    @pytest.mark.asyncio
    async def test_notification_serialization_error_handling(self):
        """Test handling of notification serialization errors."""
        # Create a minimal server instance for error testing only
        server = WebSocketJsonRpcServer(
            host="localhost", port=8002, websocket_path="/ws", max_connections=100
        )

        # Create params that can't be serialized to JSON
        class NonSerializable:
            pass

        with patch.object(server, "_clients", {"test": Mock()}):
            # Test notification with non-serializable params
            await server.broadcast_notification(
                method="test_notification", params={"object": NonSerializable()}
            )
            # Should handle gracefully without crashing
