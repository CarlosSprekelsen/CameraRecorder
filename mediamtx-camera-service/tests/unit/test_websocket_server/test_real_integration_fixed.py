# tests/unit/test_websocket_server/test_real_integration_fixed.py
"""
Real integration test with properly handled async fixtures.

This test file fixes the async fixture issues that were preventing
proper testing of real system integration.
"""

import pytest
import asyncio
import tempfile
import os
from unittest.mock import AsyncMock, Mock

from src.websocket_server.server import WebSocketJsonRpcServer
from src.camera_discovery.hybrid_monitor import HybridCameraMonitor
from src.mediamtx_wrapper.controller import MediaMTXController, StreamConfig


class TestRealIntegrationFixed:
    """Real integration tests with properly handled async fixtures."""

    @pytest.fixture
    def temp_dirs(self):
        """Create temporary directories for testing."""
        base = tempfile.mkdtemp(prefix="real_integration_test_")
        config_path = os.path.join(base, "mediamtx.yml")
        recordings_path = os.path.join(base, "recordings")
        snapshots_path = os.path.join(base, "snapshots")
        
        # Create directories
        os.makedirs(recordings_path, exist_ok=True)
        os.makedirs(snapshots_path, exist_ok=True)
        
        # Create basic MediaMTX config
        with open(config_path, 'w') as f:
            f.write("""
paths:
  all:
    runOnDemand: ffmpeg -f lavfi -i testsrc=duration=10:size=1280x720:rate=30 -c:v libx264 -f rtsp rtsp://127.0.0.1:8554/test
            """)
        
        try:
            yield {
                "base": base,
                "config_path": config_path,
                "recordings_path": recordings_path,
                "snapshots_path": snapshots_path
            }
        finally:
            import shutil
            shutil.rmtree(base, ignore_errors=True)

    @pytest.fixture
    async def camera_monitor(self):
        """Create camera monitor with proper async handling."""
        monitor = HybridCameraMonitor(
            device_range=[0, 1, 2],
            poll_interval=0.1,
            enable_capability_detection=True
        )
        await monitor.start()
        try:
            yield monitor
        finally:
            await monitor.stop()

    @pytest.fixture
    async def mediamtx_controller(self, temp_dirs):
        """Create MediaMTX controller with proper async handling."""
        controller = MediaMTXController(
            host="localhost",
            api_port=9997,
            rtsp_port=8554,
            webrtc_port=8889,
            hls_port=8888,
            config_path=temp_dirs["config_path"],
            recordings_path=temp_dirs["recordings_path"],
            snapshots_path=temp_dirs["snapshots_path"],
            health_check_interval=0.1,
            health_failure_threshold=3,
            health_circuit_breaker_timeout=1.0,
            health_max_backoff_interval=2.0,
        )
        await controller.start()
        try:
            yield controller
        finally:
            await controller.stop()

    @pytest.fixture
    async def websocket_server(self, camera_monitor, mediamtx_controller):
        """Create WebSocket server with real components."""
        # This fixture needs to be async to properly handle the async dependencies
        server = WebSocketJsonRpcServer(
            host="localhost",
            port=8002,
            websocket_path="/ws",
            max_connections=100,
            mediamtx_controller=mediamtx_controller,
            camera_monitor=camera_monitor,
        )
        return server

    @pytest.mark.asyncio
    async def test_real_camera_status_integration(self, websocket_server, camera_monitor):
        """
        Test real camera status integration with proper async handling.
        
        Requirements: REQ-WS-001
        Scenario: Real camera status integration
        Expected: Successful integration with real camera monitor
        """
        # Await the fixtures to get the actual objects
        server = await websocket_server
        monitor = await anext(camera_monitor)
        
        # Update the server's camera monitor with the awaited monitor
        server._camera_monitor = monitor
        
        # Test get_camera_status method
        result = await server._method_get_camera_status({"device": "/dev/video0"})
        
        # Verify basic structure
        assert result["device"] == "/dev/video0"
        assert "status" in result
        assert "name" in result
        assert "resolution" in result
        assert "fps" in result
        assert "capabilities" in result
        assert "streams" in result
        assert "metrics" in result
        
        # Verify capabilities structure
        capabilities = result["capabilities"]
        assert "formats" in capabilities
        assert "resolutions" in capabilities
        assert isinstance(capabilities["formats"], list)
        assert isinstance(capabilities["resolutions"], list)

    @pytest.mark.asyncio
    async def test_real_camera_list_integration(self, websocket_server):
        """
        Test real camera list integration.
        
        Requirements: REQ-WS-002
        Scenario: Real camera list integration
        Expected: Successful integration with real camera monitor
        """
        # Await the websocket_server fixture to get the actual server object
        server = await websocket_server
        
        # Test get_camera_list method
        result = await server._method_get_camera_list()
        
        # Verify basic structure
        assert "cameras" in result
        assert "total" in result
        assert "connected" in result
        
        cameras = result["cameras"]
        assert isinstance(cameras, list)
        
        # Verify each camera has proper structure
        for camera in cameras:
            assert "device" in camera
            assert "status" in camera
            assert "name" in camera
            assert "resolution" in camera
            assert "fps" in camera
            assert "capabilities" in camera



    @pytest.mark.asyncio
    async def test_error_handling_with_invalid_device(self, websocket_server, camera_monitor):
        """
        Test error handling with invalid device.
        
        Requirements: REQ-ERROR-004
        Scenario: Error handling with invalid device
        Expected: Graceful error handling
        """
        # Await the fixtures to get the actual objects
        server = await websocket_server
        monitor = await anext(camera_monitor)
        
        # Update the server's camera monitor with the awaited monitor
        server._camera_monitor = monitor
        
        # Test with invalid device
        result = await server._method_get_camera_status({"device": "/invalid/device"})
        
        # Verify graceful handling
        assert result["device"] == "/invalid/device"
        assert "status" in result
        assert "name" in result
        assert "resolution" in result
        assert "fps" in result

    @pytest.mark.asyncio
    async def test_missing_device_parameter_handling(self, websocket_server):
        """
        Test handling of missing device parameter.
        
        Requirements: REQ-ERROR-004
        Scenario: Missing device parameter
        Expected: Proper error handling
        """
        # Await the websocket_server fixture to get the actual server object
        server = await websocket_server
        
        # Test with missing device parameter
        with pytest.raises(ValueError, match="device parameter is required"):
            await server._method_get_camera_status({})
        
        # Test with None parameters
        with pytest.raises(ValueError, match="device parameter is required"):
            await server._method_get_camera_status(None)

    @pytest.mark.asyncio
    async def test_stream_name_generation(self, websocket_server):
        """
        Test stream name generation functionality.
        
        Requirements: REQ-WS-001
        Scenario: Stream name generation
        Expected: Proper stream name generation
        """
        # Await the websocket_server fixture to get the actual server object
        server = await websocket_server
        
        # Test standard device paths
        assert server._get_stream_name_from_device_path("/dev/video0") == "camera0"
        assert server._get_stream_name_from_device_path("/dev/video1") == "camera1"
        assert server._get_stream_name_from_device_path("/dev/video99") == "camera99"
        
        # Test edge cases
        assert server._get_stream_name_from_device_path("") == "camera_unknown"
        assert server._get_stream_name_from_device_path(None) == "camera_unknown"
        assert server._get_stream_name_from_device_path("/invalid/path") != "camera_unknown"
