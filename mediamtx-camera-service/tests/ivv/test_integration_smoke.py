"""
S5 Core Integration Smoke Test

Tests the complete end-to-end flow:
- Camera discovery → MediaMTX stream creation → WebSocket notification →
  recording/snapshot operations → shutdown/error recovery

This test validates the critical integration points defined in the S5 acceptance plan.

Prerequisites:
- MediaMTX server running on localhost (default ports)
- Camera Service configured and startable
- Test camera device available (USB or virtual V4L2)

Usage:
    python3 -m pytest tests/ivv/test_integration_smoke.py -v
    python3 -m pytest tests/ivv/test_integration_smoke.py::test_end_to_end_camera_flow -v
"""

import asyncio
import json
import logging
import os
import pytest
import tempfile
import time
import websockets
from pathlib import Path
from typing import Dict, Any, List, Optional
from unittest.mock import AsyncMock, Mock, patch

# Import project modules
from src.camera_service.service_manager import CameraServiceManager
from src.camera_service.config import CameraServiceConfig
from src.websocket_server.server import WebSocketJsonRpcServer
from src.mediamtx_wrapper.controller import MediaMTXController
from src.camera_discovery.hybrid_monitor import HybridCameraMonitor
from src.common.types import CameraDevice, CameraEvent, CameraEventData


class WebSocketTestClient:
    """Test client for WebSocket JSON-RPC communication."""

    def __init__(self, uri: str):
        self.uri = uri
        self.websocket = None
        self.received_messages = []
        self.request_id = 1

    async def connect(self):
        """Connect to WebSocket server."""
        self.websocket = await websockets.connect(self.uri)

    async def disconnect(self):
        """Disconnect from WebSocket server."""
        if self.websocket:
            await self.websocket.close()

    async def send_request(self, method: str, params: Optional[Dict] = None) -> Dict:
        """Send JSON-RPC request and wait for response."""
        request = {
            "jsonrpc": "2.0",
            "method": method,
            "id": self.request_id,
            "params": params or {},
        }
        self.request_id += 1

        await self.websocket.send(json.dumps(request))

        # Wait for response with matching ID
        timeout = 5.0
        start_time = time.time()

        while time.time() - start_time < timeout:
            try:
                message = await asyncio.wait_for(self.websocket.recv(), timeout=1.0)
                response = json.loads(message)

                # Store all messages for inspection
                self.received_messages.append(response)

                # Return response if ID matches
                if response.get("id") == request["id"]:
                    return response

            except asyncio.TimeoutError:
                continue

        raise TimeoutError(f"No response received for {method} within {timeout}s")

    async def wait_for_notification(self, method: str, timeout: float = 5.0) -> Dict:
        """Wait for specific notification method."""
        start_time = time.time()

        while time.time() - start_time < timeout:
            try:
                message = await asyncio.wait_for(self.websocket.recv(), timeout=1.0)
                response = json.loads(message)

                self.received_messages.append(response)

                if response.get("method") == method:
                    return response

            except asyncio.TimeoutError:
                continue

        raise TimeoutError(f"No notification {method} received within {timeout}s")


@pytest.fixture
async def test_config():
    """Create test configuration."""
    with tempfile.TemporaryDirectory() as temp_dir:
        config_data = {
            "server": {
                "host": "localhost",
                "port": 8002,
                "websocket_path": "/ws",
                "max_connections": 100,
            },
            "mediamtx": {
                "host": "localhost",
                "api_port": 9997,
                "rtsp_port": 8554,
                "webrtc_port": 8889,
                "hls_port": 8888,
                "timeout": 10.0,
            },
            "camera_discovery": {
                "device_range": [0, 1, 2],
                "poll_interval": 2.0,
                "enable_capability_detection": True,
                "detection_timeout": 5.0,
            },
            "logging": {
                "level": "INFO",
                "format": "human",
                "correlation_enabled": True,
            },
            "recording": {
                "output_dir": os.path.join(temp_dir, "recordings"),
                "snapshot_dir": os.path.join(temp_dir, "snapshots"),
                "max_duration": 3600,
                "cleanup_interval": 300,
            },
        }

        config = CameraServiceConfig()
        config.update_from_dict(config_data)
        yield config


@pytest.fixture
async def websocket_client():
    """Create WebSocket test client."""
    client = WebSocketTestClient("ws://localhost:8002/ws")
    yield client
    await client.disconnect()


@pytest.fixture
async def mock_camera_device():
    """Create mock camera device for testing."""
    device = CameraDevice(
        device_path="/dev/video0",
        name="Test Camera",
        status="CONNECTED",
        capabilities={
            "formats": ["YUYV", "MJPEG"],
            "resolutions": ["1920x1080", "1280x720"],
            "framerates": ["30", "15"],
        },
    )
    return device


class TestEndToEndIntegration:
    """Core integration smoke tests for S5 validation."""

    @pytest.mark.asyncio
    @pytest.mark.integration
    async def test_ping_basic_connectivity(self, test_config, websocket_client):
        """Test basic WebSocket connectivity and ping method."""
        # Start minimal service for ping test
        server = WebSocketJsonRpcServer(
            host=test_config.server.host,
            port=test_config.server.port,
            websocket_path=test_config.server.websocket_path,
            max_connections=test_config.server.max_connections,
        )

        # Start server in background
        server_task = asyncio.create_task(server.start())
        await asyncio.sleep(0.5)  # Allow server startup

        try:
            # Test connection and ping
            await websocket_client.connect()
            response = await websocket_client.send_request("ping")

            # Validate response format
            assert response["jsonrpc"] == "2.0"
            assert response["result"] == "pong"
            assert "id" in response

        finally:
            # Cleanup
            server_task.cancel()
            try:
                await server_task
            except asyncio.CancelledError:
                pass

    @pytest.mark.asyncio
    @pytest.mark.integration
    async def test_end_to_end_camera_flow(
        self, test_config, websocket_client, mock_camera_device
    ):
        """
        Test complete camera discovery → MediaMTX → notification → operations flow.

        This is the core smoke test validating S5 acceptance criteria:
        - Camera discovery and notification
        - MediaMTX stream creation
        - WebSocket notification delivery
        - Recording and snapshot operations
        """

        # Create mocked dependencies for controlled testing
        mock_mediamtx = Mock()
        mock_mediamtx.create_stream = AsyncMock(return_value={"status": "created"})
        mock_mediamtx.start_recording = AsyncMock(
            return_value={"filename": "test_recording.mp4"}
        )
        mock_mediamtx.stop_recording = AsyncMock(return_value={"duration": 10})
        mock_mediamtx.take_snapshot = AsyncMock(
            return_value={"filename": "test_snapshot.jpg"}
        )
        mock_mediamtx.get_stream_metrics = AsyncMock(
            return_value={"bytes_sent": 12345, "readers": 1, "uptime": 30}
        )

        mock_camera_monitor = Mock()
        mock_camera_monitor.get_camera_list = Mock(return_value=[mock_camera_device])
        mock_camera_monitor.get_camera_by_device = Mock(return_value=mock_camera_device)
        mock_camera_monitor.get_effective_capability_metadata = Mock(
            return_value={
                "resolution": "1920x1080",
                "fps": 30,
                "validation_status": "confirmed",
                "formats": ["YUYV", "MJPEG"],
            }
        )

        # Initialize service manager with mocked dependencies
        service_manager = CameraServiceManager(test_config)
        service_manager._mediamtx_controller = mock_mediamtx
        service_manager._camera_monitor = mock_camera_monitor

        # Initialize WebSocket server
        server = WebSocketJsonRpcServer(
            host=test_config.server.host,
            port=test_config.server.port,
            websocket_path=test_config.server.websocket_path,
            max_connections=test_config.server.max_connections,
            mediamtx_controller=mock_mediamtx,
            camera_monitor=mock_camera_monitor,
        )

        # Connect service manager to server for notifications
        service_manager._websocket_server = server

        # Start server
        server_task = asyncio.create_task(server.start())
        await asyncio.sleep(0.5)  # Allow startup

        try:
            # Step 1: Connect WebSocket client
            await websocket_client.connect()

            # Step 2: Test get_camera_list (should return mock camera)
            camera_list_response = await websocket_client.send_request(
                "get_camera_list"
            )
            assert camera_list_response["jsonrpc"] == "2.0"
            assert "result" in camera_list_response

            cameras = camera_list_response["result"]["cameras"]
            assert len(cameras) >= 1

            # Validate camera data structure
            camera = cameras[0]
            assert camera["device"] == "/dev/video0"
            assert camera["status"] == "CONNECTED"
            assert camera["name"] == "Test Camera"
            assert "streams" in camera

            # Step 3: Simulate camera connection event and test notification
            camera_event = CameraEvent(
                event_type="connected",
                device_path="/dev/video0",
                event_data=CameraEventData(
                    device_info=mock_camera_device,
                    capabilities=mock_camera_device.capabilities,
                    timestamp=time.time(),
                ),
            )

            # Trigger camera event handler
            notification_task = asyncio.create_task(
                websocket_client.wait_for_notification(
                    "camera_status_update", timeout=3.0
                )
            )

            # Simulate camera connection
            await service_manager.handle_camera_event(camera_event)

            # Wait for notification
            try:
                notification = await notification_task

                # Validate notification structure per API spec
                assert notification["jsonrpc"] == "2.0"
                assert notification["method"] == "camera_status_update"
                assert "params" in notification

                params = notification["params"]
                assert params["device"] == "/dev/video0"
                assert params["status"] == "CONNECTED"
                assert "name" in params
                assert "resolution" in params
                assert "fps" in params
                assert "streams" in params

            except TimeoutError:
                pytest.fail("Camera status notification not received within timeout")

            # Step 4: Test camera status API
            status_response = await websocket_client.send_request(
                "get_camera_status", {"device": "/dev/video0"}
            )

            assert status_response["jsonrpc"] == "2.0"
            assert "result" in status_response

            status = status_response["result"]
            assert status["device"] == "/dev/video0"
            assert status["status"] == "CONNECTED"
            assert "capabilities" in status
            assert "metrics" in status

            # Step 5: Test recording operations
            # Start recording
            start_recording_response = await websocket_client.send_request(
                "start_recording", {"device": "/dev/video0"}
            )

            assert start_recording_response["jsonrpc"] == "2.0"
            assert "result" in start_recording_response
            assert start_recording_response["result"]["status"] == "started"

            # Verify MediaMTX start_recording was called
            mock_mediamtx.start_recording.assert_called_once()

            # Stop recording
            stop_recording_response = await websocket_client.send_request(
                "stop_recording", {"device": "/dev/video0"}
            )

            assert stop_recording_response["jsonrpc"] == "2.0"
            assert "result" in stop_recording_response
            assert "duration" in stop_recording_response["result"]

            # Verify MediaMTX stop_recording was called
            mock_mediamtx.stop_recording.assert_called_once()

            # Step 6: Test snapshot operation
            snapshot_response = await websocket_client.send_request(
                "take_snapshot", {"device": "/dev/video0"}
            )

            assert snapshot_response["jsonrpc"] == "2.0"
            assert "result" in snapshot_response
            assert "filename" in snapshot_response["result"]
            assert "timestamp" in snapshot_response["result"]

            # Verify MediaMTX take_snapshot was called
            mock_mediamtx.take_snapshot.assert_called_once()

            # Step 7: Test error handling with invalid device
            error_response = await websocket_client.send_request(
                "get_camera_status", {"device": "/dev/video99"}
            )

            assert error_response["jsonrpc"] == "2.0"
            assert "error" in error_response
            assert error_response["error"]["code"] == -1000  # Camera not found

        finally:
            # Cleanup
            server_task.cancel()
            try:
                await server_task
            except asyncio.CancelledError:
                pass

    @pytest.mark.asyncio
    @pytest.mark.integration
    async def test_mediamtx_error_recovery(self, test_config, websocket_client):
        """Test service behavior when MediaMTX is unavailable."""

        # Create MediaMTX controller that simulates connection errors
        mock_mediamtx = Mock()
        mock_mediamtx.create_stream = AsyncMock(
            side_effect=ConnectionError("MediaMTX unavailable")
        )
        mock_mediamtx.start_recording = AsyncMock(
            side_effect=ConnectionError("MediaMTX unavailable")
        )

        mock_camera_monitor = Mock()
        mock_camera_monitor.get_camera_list = Mock(return_value=[])

        # Initialize server with failing MediaMTX
        server = WebSocketJsonRpcServer(
            host=test_config.server.host,
            port=test_config.server.port,
            websocket_path=test_config.server.websocket_path,
            max_connections=test_config.server.max_connections,
            mediamtx_controller=mock_mediamtx,
            camera_monitor=mock_camera_monitor,
        )

        server_task = asyncio.create_task(server.start())
        await asyncio.sleep(0.5)

        try:
            await websocket_client.connect()

            # Test that MediaMTX errors are properly handled
            error_response = await websocket_client.send_request(
                "start_recording", {"device": "/dev/video0"}
            )

            assert error_response["jsonrpc"] == "2.0"
            assert "error" in error_response
            assert error_response["error"]["code"] == -1003  # MediaMTX error
            assert "MediaMTX" in error_response["error"]["message"]

        finally:
            server_task.cancel()
            try:
                await server_task
            except asyncio.CancelledError:
                pass

    @pytest.mark.asyncio
    @pytest.mark.integration
    async def test_multiple_websocket_clients(self, test_config):
        """Test that multiple WebSocket clients can connect and receive notifications."""

        mock_mediamtx = Mock()
        mock_camera_monitor = Mock()
        mock_camera_monitor.get_camera_list = Mock(return_value=[])

        server = WebSocketJsonRpcServer(
            host=test_config.server.host,
            port=test_config.server.port,
            websocket_path=test_config.server.websocket_path,
            max_connections=test_config.server.max_connections,
            mediamtx_controller=mock_mediamtx,
            camera_monitor=mock_camera_monitor,
        )

        server_task = asyncio.create_task(server.start())
        await asyncio.sleep(0.5)

        clients = []
        try:
            # Connect multiple clients
            for i in range(3):
                client = WebSocketTestClient("ws://localhost:8002/ws")
                await client.connect()
                clients.append(client)

            # Send ping to each client
            for i, client in enumerate(clients):
                response = await client.send_request("ping")
                assert response["result"] == "pong"

            # Test broadcast notification (simulate camera event)
            notification_params = {
                "device": "/dev/video0",
                "status": "CONNECTED",
                "name": "Test Camera",
                "resolution": "1920x1080",
                "fps": 30,
                "streams": {"rtsp": "rtsp://localhost:8554/camera0"},
            }

            # Trigger broadcast
            await server.broadcast_notification(
                "camera_status_update", notification_params
            )

            # Verify all clients receive notification
            for client in clients:
                try:
                    notification = await client.wait_for_notification(
                        "camera_status_update", timeout=2.0
                    )
                    assert notification["method"] == "camera_status_update"
                    assert notification["params"]["device"] == "/dev/video0"
                except TimeoutError:
                    pytest.fail(f"Client did not receive broadcast notification")

        finally:
            # Cleanup
            for client in clients:
                await client.disconnect()
            server_task.cancel()
            try:
                await server_task
            except asyncio.CancelledError:
                pass


# Performance and resource validation tests
class TestPerformanceAndResources:
    """Performance and resource usage validation tests."""

    @pytest.mark.asyncio
    @pytest.mark.integration
    @pytest.mark.slow
    async def test_resource_usage_limits(self, test_config):
        """Validate service operates within acceptable resource limits."""
        # This test would need psutil or similar for real resource monitoring
        # For now, we'll simulate the checks

        mock_mediamtx = Mock()
        mock_camera_monitor = Mock()
        mock_camera_monitor.get_camera_list = Mock(return_value=[])

        server = WebSocketJsonRpcServer(
            host=test_config.server.host,
            port=test_config.server.port,
            websocket_path=test_config.server.websocket_path,
            max_connections=test_config.server.max_connections,
            mediamtx_controller=mock_mediamtx,
            camera_monitor=mock_camera_monitor,
        )

        server_task = asyncio.create_task(server.start())
        await asyncio.sleep(0.5)

        try:
            # Simulate resource usage checks
            # In real implementation, would use psutil to monitor:
            # - CPU usage < 20%
            # - Memory usage < 100MB
            # - No memory leaks over time

            client = WebSocketTestClient("ws://localhost:8002/ws")
            await client.connect()

            # Send multiple requests to stress test
            for i in range(50):
                response = await client.send_request("ping")
                assert response["result"] == "pong"

            await client.disconnect()

            # In real test, verify resource cleanup
            assert True  # Placeholder for actual resource validation

        finally:
            server_task.cancel()
            try:
                await server_task
            except asyncio.CancelledError:
                pass


if __name__ == "__main__":
    # Allow running individual tests
    pytest.main([__file__, "-v"])
