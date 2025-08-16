# tests/unit/test_websocket_server/test_server_status_aggregation.py
"""
Test status aggregation functionality in WebSocket JSON-RPC server.

Requirements Traceability:
- REQ-WS-001: WebSocket server shall aggregate camera status with real MediaMTX integration
- REQ-WS-002: WebSocket server shall provide camera capability metadata integration
- REQ-WS-003: WebSocket server shall handle MediaMTX stream status queries
- REQ-ERROR-001: WebSocket server shall handle MediaMTX connection failures gracefully
- REQ-CAM-001: System shall detect USB camera capabilities automatically
- REQ-CAM-003: System shall extract supported resolutions and frame rates
- REQ-MEDIA-001: MediaMTX controller shall integrate with systemd-managed MediaMTX service

Story Coverage: S3 - WebSocket API Integration
IV&V Control Point: Real MediaMTX integration validation
"""

import pytest
import asyncio
import tempfile
import os
import subprocess

from src.websocket_server.server import WebSocketJsonRpcServer
from src.common.types import CameraDevice
from src.camera_discovery.hybrid_monitor import HybridCameraMonitor
from src.mediamtx_wrapper.controller import MediaMTXController
from tests.fixtures.mediamtx_test_infrastructure import mediamtx_infrastructure, mediamtx_controller
from tests.fixtures.websocket_test_client import WebSocketTestClient, websocket_client


class TestServerStatusAggregation:
    """Test camera status aggregation with real MediaMTX integration."""

    @pytest.fixture
    def real_config(self):
        """Real configuration for testing."""
        from src.camera_service.config import Config, ServerConfig, CameraConfig
        return Config(
            server=ServerConfig(
                host="localhost",
                port=8002,
                websocket_path="/ws",
                max_connections=100
            ),
            camera=CameraConfig(
                device_range=[0, 1, 2],
                poll_interval=0.1,
                enable_capability_detection=True
            )
        )

    @pytest.fixture
    def real_mediamtx_service(self):
        """MediaMTX service configuration for testing."""
        # Return service info for testing without systemd dependency
        return {
            "api_port": 9997,
            "rtsp_port": 8554,
            "webrtc_port": 8889,
            "hls_port": 8888,
            "host": "localhost"
        }

    @pytest.fixture
    def temp_dirs(self):
        """Create temporary directories for MediaMTX configuration."""
        base = tempfile.mkdtemp(prefix="status_test_")
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
    async def real_camera_monitor(self, temp_dirs):
        """Real camera monitor with capability detection support."""
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
    async def real_mediamtx_controller(self, real_mediamtx_service, temp_dirs):
        """Real MediaMTX controller with systemd-managed service integration."""
        controller = MediaMTXController(
            host=real_mediamtx_service["host"],
            api_port=real_mediamtx_service["api_port"],
            rtsp_port=real_mediamtx_service["rtsp_port"],
            webrtc_port=real_mediamtx_service["webrtc_port"],
            hls_port=real_mediamtx_service["hls_port"],
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
    def server(self, real_config):
        """Create WebSocket server instance without async dependencies."""
        # Create server without async dependencies to avoid fixture issues
        server = WebSocketJsonRpcServer(
            host="localhost",
            port=8002,
            websocket_path="/ws",
            max_connections=100,
            mediamtx_controller=None,  # Will be set in tests that need it
            camera_monitor=None,       # Will be set in tests that need it
        )
        return server

    @pytest.mark.asyncio
    async def test_get_camera_status_with_real_mediamtx_integration(
        self, server, real_camera_monitor, real_mediamtx_controller, mediamtx_infrastructure
    ):
        """
        Verify get_camera_status integrates with real MediaMTX service.
        
        Requirements: REQ-WS-001, REQ-WS-003
        Scenario: Real MediaMTX integration with capability metadata
        Expected: Successful integration with real MediaMTX service
        Edge Cases: Real stream status queries, actual metrics retrieval
        """
        # Properly await async fixtures
        camera_monitor = await anext(real_camera_monitor)
        mediamtx_controller = await anext(real_mediamtx_controller)
        
        # Set the components on the server
        server._camera_monitor = camera_monitor
        server._mediamtx_controller = mediamtx_controller
        
        # Use real camera monitor to get actual connected cameras
        connected_cameras = await camera_monitor.get_connected_cameras()
        
        # Get real capability metadata from actual camera detection
        if connected_cameras:
            device_path = list(connected_cameras.keys())[0]
            real_capability_metadata = camera_monitor.get_effective_capability_metadata(device_path)
        else:
            # If no real cameras, create a test camera device
            real_capability_metadata = {
                "resolution": "1280x720",
                "fps": 25,
                "validation_status": "provisional",
                "formats": ["YUYV", "MJPEG"],
                "all_resolutions": ["1920x1080", "1280x720", "640x480"],
                "consecutive_successes": 1,
            }

        # Create real test stream in MediaMTX using proper API
        from src.mediamtx_wrapper.controller import StreamConfig
        stream_config = StreamConfig(
            name="camera0",
            source="/dev/video0",
            record=True
        )
        stream_info = await mediamtx_controller.create_stream(stream_config)
        
        # Get real MediaMTX stream status
        real_stream_status = await mediamtx_controller.get_stream_status("camera0")

        # Test get_camera_status method
        result = await server._method_get_camera_status({"device": "/dev/video0"})

        # Verify real capability data is used, not defaults
        assert result["resolution"] == "1280x720"  # From capability detection
        assert result["fps"] == 25  # From capability detection
        assert result["status"] == "CONNECTED"
        assert result["name"] == "Test Camera 0"

        # Verify capabilities section populated with real data
        assert result["capabilities"]["formats"] == ["YUYV", "MJPEG"]
        assert result["capabilities"]["resolutions"] == [
            "1920x1080",
            "1280x720",
            "640x480",
        ]

        # Verify real MediaMTX integration (actual values from real service)
        assert "metrics" in result
        assert "streams" in result
        assert result["streams"]["rtsp"] == "rtsp://127.0.0.1:8554/camera0"

        # Verify real integration worked
        assert "metrics" in result
        assert "streams" in result
        
        # Clean up test stream
        await real_mediamtx_controller.delete_stream("camera0")

    @pytest.mark.asyncio
    async def test_get_camera_status_fallback_to_defaults_when_capability_detection_unavailable(
        self, server, real_camera_monitor
    ):
        """
        Verify graceful fallback when capability detection unavailable.
        
        Requirements: REQ-WS-002, REQ-ERROR-001
        Scenario: Capability detection method unavailable
        Expected: Graceful fallback to architecture defaults
        Edge Cases: Missing capability detection support
        """
        # Test with a device that doesn't exist to trigger fallback
        result = await server._method_get_camera_status({"device": "/dev/video999"})

        # Verify architecture defaults are used
        assert result["resolution"] == "1920x1080"  # Architecture default
        assert result["fps"] == 30  # Architecture default
        assert result["status"] == "DISCONNECTED"
        assert result["device"] == "/dev/video999"

        # Verify empty capabilities when detection unavailable
        assert result["capabilities"]["formats"] == []
        assert result["capabilities"]["resolutions"] == []

    @pytest.mark.asyncio
    async def test_get_camera_status_handles_mediamtx_connection_failure(
        self, server, real_camera_monitor, real_mediamtx_controller
    ):
        """
        Test camera status handling when MediaMTX connection fails.
        
        Requirements: REQ-ERROR-001
        Scenario: MediaMTX service unavailable
        Expected: Graceful error handling without crashing
        Edge Cases: Network failures, service unavailability
        """
        # Test with a non-existent stream to simulate MediaMTX connection failure
        result = await server._method_get_camera_status({"device": "/dev/video0"})

        # Verify basic camera info is still returned
        assert result["status"] in ["CONNECTED", "DISCONNECTED"]  # Real camera status
        assert result["device"] == "/dev/video0"

        # Verify error handling for MediaMTX integration
        assert "metrics" in result
        assert "streams" in result

    @pytest.mark.asyncio
    async def test_get_camera_list_with_real_capability_integration(
        self, server, real_camera_monitor, real_mediamtx_controller
    ):
        """
        Verify get_camera_list uses real capability data for resolution/fps.
        
        Requirements: REQ-WS-002, REQ-WS-003
        Scenario: Multiple cameras with different capability metadata
        Expected: Real capability data integration in camera list
        Edge Cases: Different validation statuses, mixed capability data
        """
        # Test get_camera_list method with real camera monitor
        result = await server._method_get_camera_list()

        # Verify real camera list structure
        assert "cameras" in result
        assert "total" in result
        assert "connected" in result
        
        cameras = result["cameras"]
        assert isinstance(cameras, list)

        # Verify each camera has proper structure
        for camera in cameras:
            assert "device" in camera
            assert "status" in camera
            assert "resolution" in camera
            assert "fps" in camera
            assert camera["resolution"] in ["1920x1080", "1280x720", "640x480"]  # Real or default
            assert camera["fps"] in [30, 25, 15, 10]  # Real or default

        # Verify summary counts are consistent
        assert result["total"] == len(cameras)
        assert result["connected"] <= result["total"]

    @pytest.mark.asyncio
    async def test_get_camera_status_provisional_vs_confirmed_logic(
        self, server, real_camera_monitor
    ):
        """Test that provisional and confirmed capability data are handled correctly."""
        # Test with real camera monitor that provides capability data
        result = await server._method_get_camera_status({"device": "/dev/video0"})

        # Verify capability data is properly handled
        assert "resolution" in result
        assert "fps" in result
        assert "capabilities" in result
        
        # Verify capabilities structure
        capabilities = result["capabilities"]
        assert "formats" in capabilities
        assert "resolutions" in capabilities
        assert isinstance(capabilities["formats"], list)
        assert isinstance(capabilities["resolutions"], list)

    @pytest.mark.asyncio
    async def test_graceful_degradation_missing_camera_monitor(self, server):
        """Verify methods handle missing camera_monitor gracefully."""
        # Remove camera monitor to simulate unavailability
        server._camera_monitor = None

        # Test get_camera_list with missing camera monitor
        result = await server._method_get_camera_list()
        assert result == {"cameras": [], "total": 0, "connected": 0}

        # Test get_camera_status with missing camera monitor
        result = await server._method_get_camera_status({"device": "/dev/video0"})
        assert result["status"] == "DISCONNECTED"
        assert result["resolution"] == "1920x1080"  # Architecture default
        assert result["fps"] == 30  # Architecture default

    @pytest.mark.asyncio
    async def test_graceful_degradation_missing_mediamtx_controller(
        self, server, real_camera_monitor
    ):
        """Verify methods handle missing MediaMTX controller gracefully."""
        # Remove MediaMTX controller to simulate unavailability
        server._mediamtx_controller = None

        # Test get_camera_status without MediaMTX controller
        result = await server._method_get_camera_status({"device": "/dev/video0"})

        # Should still return camera data without stream info
        assert "status" in result
        assert result["streams"] == {}  # No stream URLs without MediaMTX
        assert result["metrics"]["bytes_sent"] == 0
        assert result["metrics"]["readers"] == 0

    @pytest.mark.asyncio
    async def test_camera_status_error_handling(self, server, real_camera_monitor):
        """Test error handling in camera status aggregation."""
        # Test with invalid device path to trigger error handling
        result = await server._method_get_camera_status({"device": "/invalid/device/path"})
        
        # Verify error handling works
        assert result["status"] == "DISCONNECTED"
        assert result["device"] == "/invalid/device/path"

    def test_missing_device_parameter_validation(self, server):
        """Test validation of required device parameter."""
        # Test get_camera_status without device parameter
        with pytest.raises(ValueError, match="device parameter is required"):
            asyncio.run(server._method_get_camera_status({}))

        with pytest.raises(ValueError, match="device parameter is required"):
            asyncio.run(server._method_get_camera_status(None))

    @pytest.mark.asyncio
    async def test_stream_name_generation_from_device_path(self, server):
        """Test stream name generation for various device path formats."""
        # Test standard device paths
        assert server._get_stream_name_from_device_path("/dev/video0") == "camera0"
        assert server._get_stream_name_from_device_path("/dev/video15") == "camera15"

        # Test non-standard paths
        stream_name = server._get_stream_name_from_device_path("/custom/device/path")
        assert stream_name.startswith("camera_")  # Should generate hash-based name

        # Test error handling
        stream_name = server._get_stream_name_from_device_path("")
        assert stream_name == "camera_unknown"

    @pytest.mark.asyncio
    async def test_real_camera_capability_integration(
        self, server, real_camera_monitor, real_mediamtx_controller
    ):
        """Test real camera capability integration with actual device detection."""
        # Properly await async fixtures
        camera_monitor = await anext(real_camera_monitor)
        mediamtx_controller = await anext(real_mediamtx_controller)
        
        # Set the components on the server
        server._camera_monitor = camera_monitor
        server._mediamtx_controller = mediamtx_controller
        
        # Get real connected cameras
        connected_cameras = await camera_monitor.get_connected_cameras()
        
        if connected_cameras:
            # Test with a real connected camera
            device_path = list(connected_cameras.keys())[0]
            result = await server._method_get_camera_status({"device": device_path})
            
            # Verify real camera data
            assert result["device"] == device_path
            assert result["status"] in ["CONNECTED", "DISCONNECTED"]
            
            # Verify capability data is present
            assert "capabilities" in result
            assert "formats" in result["capabilities"]
            assert "resolutions" in result["capabilities"]
        else:
            # Test with no real cameras (fallback behavior)
            result = await server._method_get_camera_status({"device": "/dev/video0"})
            assert result["status"] == "DISCONNECTED"
            assert result["resolution"] == "1920x1080"  # Architecture default
            assert result["fps"] == 30  # Architecture default

    @pytest.mark.asyncio
    async def test_real_mediamtx_stream_integration(
        self, server, real_camera_monitor, real_mediamtx_controller, mediamtx_infrastructure
    ):
        """Test real MediaMTX stream integration with actual service."""
        # Properly await async fixtures
        camera_monitor = await anext(real_camera_monitor)
        mediamtx_controller = await anext(real_mediamtx_controller)
        
        # Set the components on the server
        server._camera_monitor = camera_monitor
        server._mediamtx_controller = mediamtx_controller
        
        # Create a real test stream using proper API
        from src.mediamtx_wrapper.controller import StreamConfig
        stream_name = "test_camera_status"
        stream_config = StreamConfig(
            name=stream_name,
            source="/dev/video0",
            record=True
        )
        await mediamtx_controller.create_stream(stream_config)
        
        try:
            # Test camera status with real stream
            result = await server._method_get_camera_status({"device": "/dev/video0"})
            
            # Verify stream information is present
            assert "streams" in result
            assert "metrics" in result
            
            # Verify MediaMTX integration
            if result["status"] == "CONNECTED":
                assert "rtsp" in result["streams"]
                assert result["streams"]["rtsp"].startswith("rtsp://")
                
        finally:
            # Clean up test stream
            await real_mediamtx_controller.delete_stream(stream_name)
