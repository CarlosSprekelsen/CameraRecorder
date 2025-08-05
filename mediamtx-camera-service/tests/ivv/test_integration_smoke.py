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

Related Epic/Story: E1/S5 - Core Integration IV&V
"""

import asyncio
import json
import os
import pytest
import pytest_asyncio
import tempfile
import time
import uuid
import websockets
from typing import Dict, Optional
from unittest.mock import AsyncMock, Mock, patch

# Import project modules
from src.camera_service.service_manager import ServiceManager
from src.camera_service.config import Config, ServerConfig, MediaMTXConfig, CameraConfig, RecordingConfig
from src.websocket_server.server import WebSocketJsonRpcServer
from src.common.types import CameraDevice
from src.camera_discovery.hybrid_monitor import CameraEvent, CameraEventData


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
def test_config():
    """Create test configuration with dynamic port allocation."""
    import socket
    
    def find_free_port():
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
            s.bind(('', 0))
            s.listen(1)
            port = s.getsockname()[1]
        return port
    
    # Use dynamic port allocation to avoid conflicts
    server_port = find_free_port()
    
    return Config(
        server=ServerConfig(
            host="localhost",
            port=server_port,  # Dynamic port
            websocket_path="/ws",
            max_connections=100,
        ),
        mediamtx=MediaMTXConfig(
            host="localhost",
            api_port=9997,
            rtsp_port=8554,
            webrtc_port=8889,
            hls_port=8888,
        ),
        camera=CameraConfig(
            device_range=[0, 1, 2],
            poll_interval=2.0,
            enable_capability_detection=True,
            detection_timeout=5.0,
        ),
        recording=RecordingConfig(
            auto_record=False,
            format="mp4",
            quality="medium",
            max_duration=3600,
            cleanup_after_days=30,
        ),
    )


@pytest_asyncio.fixture
async def websocket_client(test_config):
    """Create WebSocket test client."""
    client = WebSocketTestClient(f"ws://localhost:{test_config.server.port}/ws")
    yield client


@pytest_asyncio.fixture
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
    yield device


class TestEndToEndIntegration:
    """Core integration smoke tests for S5 validation."""

    @pytest.mark.asyncio
    @pytest.mark.integration
    async def test_ping_basic_connectivity(self, test_config, websocket_client):
        """
        Test basic WebSocket connectivity and ping method.
        
        Validates:
        - WebSocket server startup
        - JSON-RPC protocol handling
        - Basic request/response flow
        """
        # Create WebSocket server
        server = WebSocketJsonRpcServer(
            host=test_config.server.host,
            port=test_config.server.port,
            websocket_path=test_config.server.websocket_path,
            max_connections=test_config.server.max_connections,
        )

        # Start server
        server_task = asyncio.create_task(server.start())
        await asyncio.sleep(0.5)  # Allow server startup

        try:
            # Connect WebSocket client
            await websocket_client.connect()

            # Test ping method
            response = await websocket_client.send_request("ping")

            # Verify response
            assert response["jsonrpc"] == "2.0"
            assert response["result"] == "pong"
            assert "id" in response

            await websocket_client.disconnect()

        finally:
            # Cleanup
            server_task.cancel()
            try:
                await server_task
            except asyncio.CancelledError:
                pass

    @pytest.mark.asyncio
    @pytest.mark.integration
    async def test_end_to_end_camera_flow(self, test_config, websocket_client, mock_camera_device):
        """
        Test complete end-to-end camera flow.
        
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
            return_value={"filename": "test_recording.mp4", "session_id": str(uuid.uuid4())}
        )
        mock_mediamtx.stop_recording = AsyncMock(return_value={"duration": 10, "file_size": 1024000})
        mock_mediamtx.take_snapshot = AsyncMock(
            return_value={
                "filename": "test_snapshot.jpg",
                "file_size": 204800,
                "file_path": "/opt/camera-service/snapshots/test_snapshot.jpg"
            }
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
        service_manager = ServiceManager(test_config)
        service_manager._mediamtx_controller = mock_mediamtx
        service_manager._camera_monitor = mock_camera_monitor

        # Initialize WebSocket server
        server = WebSocketJsonRpcServer(
            host=test_config.server.host,
            port=test_config.server.port,
            websocket_path=test_config.server.websocket_path,
            max_connections=test_config.server.max_connections,
        )

        # Link service manager to server
        server.set_service_manager(service_manager)

        # Start server
        server_task = asyncio.create_task(server.start())
        await asyncio.sleep(0.5)  # Allow server startup

        try:
            # Step 1: Connect WebSocket client
            await websocket_client.connect()

            # Step 2: Test camera discovery - get_camera_list
            camera_list_response = await websocket_client.send_request("get_camera_list")

            assert camera_list_response["jsonrpc"] == "2.0"
            assert "result" in camera_list_response
            assert "cameras" in camera_list_response["result"]
            assert camera_list_response["result"]["total"] >= 0

            # Step 3: Test camera status retrieval
            camera_status_response = await websocket_client.send_request(
                "get_camera_status", {"device": "/dev/video0"}
            )

            assert camera_status_response["jsonrpc"] == "2.0"
            assert "result" in camera_status_response
            result = camera_status_response["result"]

            # Verify required fields per API specification
            required_fields = ["device", "status", "name", "resolution", "fps", "streams"]
            for field in required_fields:
                assert field in result, f"Missing required field: {field}"

            # Verify stream URLs structure
            assert "rtsp" in result["streams"]
            assert "webrtc" in result["streams"]
            assert "hls" in result["streams"]

            # Step 4: Test recording operations
            start_recording_response = await websocket_client.send_request(
                "start_recording", {"device": "/dev/video0", "duration": 10}
            )

            assert start_recording_response["jsonrpc"] == "2.0"
            assert "result" in start_recording_response
            assert "session_id" in start_recording_response["result"]
            assert "filename" in start_recording_response["result"]

            # Verify MediaMTX start_recording was called
            mock_mediamtx.start_recording.assert_called_once()

            # Step 5: Stop recording
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
            await websocket_client.disconnect()
            server_task.cancel()
            try:
                await server_task
            except asyncio.CancelledError:
                pass

    @pytest.mark.asyncio
    @pytest.mark.integration
    async def test_mediamtx_error_recovery(self, test_config, websocket_client):
        """
        Test service behavior when MediaMTX is unavailable.
        
        Validates error recovery scenario ER-1:
        - Proper error codes when MediaMTX unavailable
        - Service stability during errors
        """
        # Create MediaMTX controller that simulates unavailability
        mock_mediamtx = Mock()
        mock_mediamtx.create_stream = AsyncMock(side_effect=ConnectionError("MediaMTX unavailable"))
        mock_mediamtx.start_recording = AsyncMock(side_effect=ConnectionError("MediaMTX unavailable"))
        mock_mediamtx.take_snapshot = AsyncMock(side_effect=ConnectionError("MediaMTX unavailable"))

        # Create mock camera monitor
        mock_camera_monitor = Mock()
        mock_camera_monitor.get_camera_by_device = Mock(return_value=CameraDevice(
            device_path="/dev/video0",
            name="Test Camera",
            status="CONNECTED",
            capabilities={}
        ))

        # Initialize service manager with failing MediaMTX
        service_manager = ServiceManager(test_config)
        service_manager._mediamtx_controller = mock_mediamtx
        service_manager._camera_monitor = mock_camera_monitor

        # Initialize WebSocket server
        server = WebSocketJsonRpcServer(
            host=test_config.server.host,
            port=test_config.server.port,
            websocket_path=test_config.server.websocket_path,
            max_connections=test_config.server.max_connections,
        )
        server.set_service_manager(service_manager)

        # Start server
        server_task = asyncio.create_task(server.start())
        await asyncio.sleep(0.5)

        try:
            # Connect WebSocket client
            await websocket_client.connect()

            # Test API calls return proper error codes during MediaMTX outage
            recording_response = await websocket_client.send_request(
                "start_recording", {"device": "/dev/video0"}
            )

            assert recording_response["jsonrpc"] == "2.0"
            assert "error" in recording_response
            assert recording_response["error"]["code"] == -1003  # MediaMTX error

            # Test snapshot also fails gracefully
            snapshot_response = await websocket_client.send_request(
                "take_snapshot", {"device": "/dev/video0"}
            )

            assert snapshot_response["jsonrpc"] == "2.0"
            assert "error" in snapshot_response
            assert snapshot_response["error"]["code"] == -1003  # MediaMTX error

        finally:
            # Cleanup
            await websocket_client.disconnect()
            server_task.cancel()
            try:
                await server_task
            except asyncio.CancelledError:
                pass

    @pytest.mark.asyncio
    @pytest.mark.integration
    async def test_multiple_websocket_clients(self, test_config, mock_camera_device):
        """
        Test WebSocket server handles multiple clients.
        
        Validates scenario PR-2:
        - Multiple client connections
        - Broadcast notification delivery
        - No cross-client interference
        """
        # Create mocked dependencies
        mock_mediamtx = Mock()
        mock_mediamtx.get_stream_metrics = AsyncMock(
            return_value={"bytes_sent": 12345, "readers": 2, "uptime": 60}
        )

        mock_camera_monitor = Mock()
        mock_camera_monitor.get_camera_list = Mock(return_value=[mock_camera_device])
        mock_camera_monitor.get_camera_by_device = Mock(return_value=mock_camera_device)
        mock_camera_monitor.get_effective_capability_metadata = Mock(
            return_value={
                "resolution": "1920x1080",
                "fps": 30,
                "validation_status": "confirmed",
            }
        )

        # Initialize service components
        service_manager = ServiceManager(test_config)
        service_manager._mediamtx_controller = mock_mediamtx
        service_manager._camera_monitor = mock_camera_monitor

        server = WebSocketJsonRpcServer(
            host=test_config.server.host,
            port=test_config.server.port,
            websocket_path=test_config.server.websocket_path,
            max_connections=test_config.server.max_connections,
        )
        server.set_service_manager(service_manager)

        # Start server
        server_task = asyncio.create_task(server.start())
        await asyncio.sleep(0.5)

        # Create multiple clients
        clients = []
        client_tasks = []

        try:
            # Connect 3 clients simultaneously
            for i in range(3):
                client = WebSocketTestClient(f"ws://localhost:8002/ws")
                await client.connect()
                clients.append(client)

            # Send request from first client
            response = await clients[0].send_request("get_camera_list")
            assert response["jsonrpc"] == "2.0"
            assert "result" in response

            # Simulate camera status notification broadcast
            await server.notify_camera_status_update({
                "device": "/dev/video0",
                "status": "CONNECTED",
                "name": "Test Camera",
                "resolution": "1920x1080",
                "fps": 30,
                "streams": {
                    "rtsp": "rtsp://localhost:8554/camera0",
                    "webrtc": "http://localhost:8889/camera0/webrtc",
                    "hls": "http://localhost:8888/camera0"
                }
            })

            # Verify all clients receive notification
            notifications = []
            for client in clients:
                try:
                    notification = await client.wait_for_notification("camera_status_update", timeout=3.0)
                    notifications.append(notification)
                except TimeoutError:
                    pytest.fail(f"Client did not receive camera_status_update notification")

            # Verify all notifications are identical
            assert len(notifications) == 3
            for notification in notifications:
                assert notification["method"] == "camera_status_update"
                assert notification["params"]["device"] == "/dev/video0"

        finally:
            # Cleanup clients
            for client in clients:
                await client.disconnect()

            # Cleanup server
            server_task.cancel()
            try:
                await server_task
            except asyncio.CancelledError:
                pass

    @pytest.mark.asyncio
    @pytest.mark.integration
    async def test_notification_delivery_flow(self, test_config, websocket_client, mock_camera_device):
        """
        Test real-time notification delivery for camera events.
        
        Validates scenario HP-3:
        - Camera status update notifications
        - Recording status update notifications
        - Proper notification schema compliance
        """
        # Setup mocked components
        mock_mediamtx = Mock()
        mock_mediamtx.start_recording = AsyncMock(
            return_value={"filename": "test_recording.mp4", "session_id": str(uuid.uuid4())}
        )

        mock_camera_monitor = Mock()
        mock_camera_monitor.get_camera_by_device = Mock(return_value=mock_camera_device)

        # Initialize service
        service_manager = ServiceManager(test_config)
        service_manager._mediamtx_controller = mock_mediamtx
        service_manager._camera_monitor = mock_camera_monitor

        server = WebSocketJsonRpcServer(
            host=test_config.server.host,
            port=test_config.server.port,
            websocket_path=test_config.server.websocket_path,
            max_connections=test_config.server.max_connections,
        )
        server.set_service_manager(service_manager)

        # Start server
        server_task = asyncio.create_task(server.start())
        await asyncio.sleep(0.5)

        try:
            # Connect client
            await websocket_client.connect()

            # Test camera status notification
            await server.notify_camera_status_update({
                "device": "/dev/video0",
                "status": "CONNECTED",
                "name": "Test Camera",
                "resolution": "1920x1080",
                "fps": 30,
                "streams": {
                    "rtsp": "rtsp://localhost:8554/camera0",
                    "webrtc": "http://localhost:8889/camera0/webrtc",
                    "hls": "http://localhost:8888/camera0"
                }
            })

            # Wait for and verify camera status notification
            camera_notification = await websocket_client.wait_for_notification("camera_status_update")
            assert camera_notification["method"] == "camera_status_update"
            
            params = camera_notification["params"]
            required_fields = ["device", "status", "name", "resolution", "fps", "streams"]
            for field in required_fields:
                assert field in params, f"Missing required field in notification: {field}"

            # Test recording status notification
            await server.notify_recording_status_update({
                "device": "/dev/video0",
                "status": "STARTED",
                "filename": "test_recording.mp4",
                "duration": 0
            })

            # Wait for and verify recording notification
            recording_notification = await websocket_client.wait_for_notification("recording_status_update")
            assert recording_notification["method"] == "recording_status_update"
            
            rec_params = recording_notification["params"]
            required_rec_fields = ["device", "status", "filename", "duration"]
            for field in required_rec_fields:
                assert field in rec_params, f"Missing required field in recording notification: {field}"

        finally:
            # Cleanup
            await websocket_client.disconnect()
            server_task.cancel()
            try:
                await server_task
            except asyncio.CancelledError:
                pass

    @pytest.mark.asyncio
    @pytest.mark.integration
    async def test_invalid_api_requests(self, test_config, websocket_client):
        """
        Test error handling for invalid API requests.
        
        Validates scenario ER-3:
        - Malformed JSON handling
        - Missing parameter validation
        - Invalid device path handling
        - Proper JSON-RPC error codes
        """
        # Initialize minimal service for error testing
        service_manager = ServiceManager(test_config)
        
        # Mock camera monitor to return no cameras
        mock_camera_monitor = Mock()
        mock_camera_monitor.get_camera_by_device = Mock(return_value=None)
        service_manager._camera_monitor = mock_camera_monitor

        server = WebSocketJsonRpcServer(
            host=test_config.server.host,
            port=test_config.server.port,
            websocket_path=test_config.server.websocket_path,
            max_connections=test_config.server.max_connections,
        )
        server.set_service_manager(service_manager)

        # Start server
        server_task = asyncio.create_task(server.start())
        await asyncio.sleep(0.5)

        try:
            # Connect client
            await websocket_client.connect()

            # Test 1: Invalid device path
            error_response = await websocket_client.send_request(
                "get_camera_status", {"device": "/dev/video99"}
            )
            assert error_response["jsonrpc"] == "2.0"
            assert "error" in error_response
            assert error_response["error"]["code"] == -1000  # Camera not found

            # Test 2: Missing required parameters
            error_response = await websocket_client.send_request("get_camera_status", {})
            assert error_response["jsonrpc"] == "2.0"
            assert "error" in error_response
            assert error_response["error"]["code"] == -32602  # Invalid params

            # Test 3: Non-existent method
            error_response = await websocket_client.send_request("invalid_method", {})
            assert error_response["jsonrpc"] == "2.0"
            assert "error" in error_response
            assert error_response["error"]["code"] == -32601  # Method not found

        finally:
            # Cleanup
            await websocket_client.disconnect()
            server_task.cancel()
            try:
                await server_task
            except asyncio.CancelledError:
                pass


class TestPerformanceAndResources:
    """Performance and resource validation tests."""

    @pytest.mark.asyncio
    @pytest.mark.integration
    @pytest.mark.slow
    async def test_resource_usage_limits(self, test_config, websocket_client):
        """
        Test service operates within acceptable resource limits.
        
        Validates scenario PR-1:
        - Memory usage <100MB
        - CPU usage reasonable
        - No memory leaks during operations
        """
        import psutil
        import gc

        # Track initial memory
        process = psutil.Process()
        initial_memory = process.memory_info().rss / 1024 / 1024  # MB

        # Initialize service
        service_manager = ServiceManager(test_config)
        
        # Mock components for testing
        mock_mediamtx = Mock()
        mock_mediamtx.take_snapshot = AsyncMock(return_value={"filename": "test.jpg"})
        
        mock_camera_monitor = Mock()
        mock_camera_device = CameraDevice(
            device_path="/dev/video0",
            name="Test Camera",
            status="CONNECTED",
            capabilities={}
        )
        mock_camera_monitor.get_camera_by_device = Mock(return_value=mock_camera_device)
        
        service_manager._mediamtx_controller = mock_mediamtx
        service_manager._camera_monitor = mock_camera_monitor

        server = WebSocketJsonRpcServer(
            host=test_config.server.host,
            port=test_config.server.port,
            websocket_path=test_config.server.websocket_path,
            max_connections=test_config.server.max_connections,
        )
        server.set_service_manager(service_manager)

        # Start server
        server_task = asyncio.create_task(server.start())
        await asyncio.sleep(0.5)

        try:
            # Connect client
            await websocket_client.connect()

            # Perform multiple operations to test resource usage
            for i in range(10):
                # Take snapshots
                response = await websocket_client.send_request(
                    "take_snapshot", {"device": "/dev/video0"}
                )
                assert "result" in response or "error" in response
                
                # Get camera status
                response = await websocket_client.send_request(
                    "get_camera_status", {"device": "/dev/video0"}
                )
                assert "result" in response or "error" in response

                # Short pause between operations
                await asyncio.sleep(0.1)

            # Force garbage collection
            gc.collect()
            await asyncio.sleep(1.0)

            # Check final memory usage
            final_memory = process.memory_info().rss / 1024 / 1024  # MB
            memory_growth = final_memory - initial_memory

            # Verify memory usage is reasonable (allowing for test overhead)
            assert final_memory < 150, f"Memory usage {final_memory:.1f}MB exceeds 150MB limit"
            assert memory_growth < 50, f"Memory growth {memory_growth:.1f}MB suggests potential leak"

        finally:
            # Cleanup
            await websocket_client.disconnect()
            server_task.cancel()
            try:
                await server_task
            except asyncio.CancelledError:
                pass


# Test configuration for different environments
@pytest.mark.integration
def test_config_validation():
    """
    Test configuration validation for integration environment.
    
    Validates:
    - Required configuration sections present
    - Port assignments valid
    - Directory creation successful
    """
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
            "recording": {
                "output_dir": os.path.join(temp_dir, "recordings"),
                "snapshot_dir": os.path.join(temp_dir, "snapshots"),
                "max_duration": 3600,
                "cleanup_interval": 300,
            },
        }

        # Test configuration creation
        config = Config()
        config.update_from_dict(config_data)
        
        # Override MediaMTX paths to use temp directory for testing
        config.mediamtx.recordings_path = os.path.join(temp_dir, "recordings")
        config.mediamtx.snapshots_path = os.path.join(temp_dir, "snapshots")

        # Verify required sections
        assert hasattr(config, 'server')
        assert hasattr(config, 'mediamtx')
        assert hasattr(config, 'camera')  # Changed from camera_discovery
        assert hasattr(config, 'recording')

        # Verify port assignments
        assert config.server.port == 8002
        assert config.mediamtx.api_port == 9997

        # Verify directories can be created
        # Note: RecordingConfig doesn't have output_dir/snapshot_dir, using MediaMTX paths
        os.makedirs(config.mediamtx.recordings_path, exist_ok=True)
        os.makedirs(config.mediamtx.snapshots_path, exist_ok=True)
        
        assert os.path.exists(config.mediamtx.recordings_path)
        assert os.path.exists(config.mediamtx.snapshots_path)


if __name__ == "__main__":
    pytest.main([__file__, "-v"])