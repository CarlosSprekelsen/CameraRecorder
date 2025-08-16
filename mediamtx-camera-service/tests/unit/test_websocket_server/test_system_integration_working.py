# tests/unit/test_websocket_server/test_system_integration_working.py
"""
Working system integration test that properly tests real components.

This test file demonstrates proper testing of real system integration
by avoiding the async fixture issues and using direct component testing.
"""

import pytest
import asyncio
import tempfile
import os
from unittest.mock import AsyncMock, Mock

from src.websocket_server.server import WebSocketJsonRpcServer
from src.camera_discovery.hybrid_monitor import HybridCameraMonitor
from src.mediamtx_wrapper.controller import MediaMTXController, StreamConfig


class TestSystemIntegrationWorking:
    """Working system integration tests that properly test real components."""

    @pytest.mark.asyncio
    async def test_real_websocket_server_with_real_components(self):
        """
        Test real WebSocket server with real components using direct instantiation.
        
        Requirements: REQ-WS-001, REQ-WS-002, REQ-WS-003
        Scenario: Real system integration with direct component instantiation
        Expected: Successful integration with real components
        """
        # Create temporary directories
        base = tempfile.mkdtemp(prefix="working_integration_test_")
        config_path = os.path.join(base, "mediamtx.yml")
        recordings_path = os.path.join(base, "recordings")
        snapshots_path = os.path.join(base, "snapshots")
        
        try:
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
            
            # Create real camera monitor
            camera_monitor = HybridCameraMonitor(
                device_range=[0, 1, 2],
                poll_interval=0.1,
                enable_capability_detection=True
            )
            await camera_monitor.start()
            
            # Create real MediaMTX controller
            mediamtx_controller = MediaMTXController(
                host="localhost",
                api_port=9997,
                rtsp_port=8554,
                webrtc_port=8889,
                hls_port=8888,
                config_path=config_path,
                recordings_path=recordings_path,
                snapshots_path=snapshots_path,
                health_check_interval=0.1,
                health_failure_threshold=3,
                health_circuit_breaker_timeout=1.0,
                health_max_backoff_interval=2.0,
            )
            await mediamtx_controller.start()
            
            # Create WebSocket server with real components
            websocket_server = WebSocketJsonRpcServer(
                host="localhost",
                port=8002,
                websocket_path="/ws",
                max_connections=100,
                mediamtx_controller=mediamtx_controller,
                camera_monitor=camera_monitor,
            )
            
            # Test get_camera_status method
            result = await websocket_server._method_get_camera_status({"device": "/dev/video0"})
            
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
            
            # Test get_camera_list method
            list_result = await websocket_server._method_get_camera_list()
            
            # Verify camera list structure
            assert "cameras" in list_result
            assert "total" in list_result
            assert "connected" in list_result
            
            cameras = list_result["cameras"]
            assert isinstance(cameras, list)
            
            # Clean up
            await mediamtx_controller.stop()
            await camera_monitor.stop()
            
        finally:
            # Clean up temporary directories
            import shutil
            shutil.rmtree(base, ignore_errors=True)

    @pytest.mark.asyncio
    async def test_websocket_server_error_handling(self):
        """
        Test WebSocket server error handling with real components.
        
        Requirements: REQ-ERROR-004
        Scenario: Error handling with real components
        Expected: Graceful error handling
        """
        # Create WebSocket server without components to test error handling
        websocket_server = WebSocketJsonRpcServer(
            host="localhost",
            port=8002,
            websocket_path="/ws",
            max_connections=100,
            mediamtx_controller=None,
            camera_monitor=None,
        )
        
        # Test with missing device parameter
        with pytest.raises(ValueError, match="device parameter is required"):
            await websocket_server._method_get_camera_status({})
        
        # Test with None parameters
        with pytest.raises(ValueError, match="device parameter is required"):
            await websocket_server._method_get_camera_status(None)
        
        # Test with invalid device (should handle gracefully)
        result = await websocket_server._method_get_camera_status({"device": "/invalid/device"})
        assert result["device"] == "/invalid/device"
        assert result["status"] == "DISCONNECTED"  # Default status when no camera monitor

    @pytest.mark.asyncio
    async def test_mediamtx_integration_with_mock(self):
        """
        Test MediaMTX integration with mocked controller.
        
        Requirements: REQ-WS-001, REQ-WS-003
        Scenario: MediaMTX integration with mocked controller
        Expected: Successful integration with mocked MediaMTX controller
        """
        # Create camera monitor
        camera_monitor = HybridCameraMonitor(
            device_range=[0, 1, 2],
            poll_interval=0.1,
            enable_capability_detection=True
        )
        await camera_monitor.start()
        
        # Create mock MediaMTX controller
        mock_controller = Mock()
        mock_controller.get_stream_status = AsyncMock(return_value={
            'status': 'active',
            'bytes_sent': 1024,
            'readers': 1
        })
        
        # Create WebSocket server with real camera monitor and mock MediaMTX controller
        websocket_server = WebSocketJsonRpcServer(
            host="localhost",
            port=8002,
            websocket_path="/ws",
            max_connections=100,
            mediamtx_controller=mock_controller,
            camera_monitor=camera_monitor,
        )
        
        # Test get_camera_status method
        result = await websocket_server._method_get_camera_status({"device": "/dev/video0"})
        
        # Verify MediaMTX integration worked
        assert result["device"] == "/dev/video0"
        assert "streams" in result
        assert "metrics" in result
        
        # Verify MediaMTX controller was called
        mock_controller.get_stream_status.assert_called_once()
        
        # Clean up
        await camera_monitor.stop()

    def test_stream_name_generation(self):
        """
        Test stream name generation functionality.
        
        Requirements: REQ-WS-001
        Scenario: Stream name generation
        Expected: Proper stream name generation
        """
        # Create WebSocket server
        websocket_server = WebSocketJsonRpcServer(
            host="localhost",
            port=8002,
            websocket_path="/ws",
            max_connections=100,
        )
        
        # Test standard device paths
        assert websocket_server._get_stream_name_from_device_path("/dev/video0") == "camera0"
        assert websocket_server._get_stream_name_from_device_path("/dev/video1") == "camera1"
        assert websocket_server._get_stream_name_from_device_path("/dev/video99") == "camera99"
        
        # Test edge cases
        assert websocket_server._get_stream_name_from_device_path("") == "camera_unknown"
        assert websocket_server._get_stream_name_from_device_path(None) == "camera_unknown"
        assert websocket_server._get_stream_name_from_device_path("/invalid/path") != "camera_unknown"

    @pytest.mark.asyncio
    async def test_concurrent_operations(self):
        """
        Test concurrent operations handling.
        
        Requirements: REQ-PERF-001
        Scenario: Concurrent operations
        Expected: Efficient handling of concurrent operations
        """
        # Create WebSocket server
        websocket_server = WebSocketJsonRpcServer(
            host="localhost",
            port=8002,
            websocket_path="/ws",
            max_connections=100,
        )
        
        # Test concurrent ping operations
        async def make_ping():
            return await websocket_server._method_ping()
        
        # Make concurrent requests
        start_time = asyncio.get_event_loop().time()
        
        tasks = [make_ping() for _ in range(10)]
        results = await asyncio.gather(*tasks)
        
        end_time = asyncio.get_event_loop().time()
        
        # Verify all requests completed successfully
        assert len(results) == 10
        assert all(result == "pong" for result in results)
        
        # Verify performance (should complete within reasonable time)
        total_time = end_time - start_time
        assert total_time < 1.0  # Should complete within 1 second
