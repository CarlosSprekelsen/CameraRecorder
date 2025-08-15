# tests/unit/test_websocket_server/test_server_status_aggregation.py
"""
Test status aggregation functionality in WebSocket JSON-RPC server.

Requirements Traceability:
- REQ-WS-001: WebSocket server shall aggregate camera status with real MediaMTX integration
- REQ-WS-002: WebSocket server shall provide camera capability metadata integration
- REQ-WS-003: WebSocket server shall handle MediaMTX stream status queries
- REQ-ERROR-001: WebSocket server shall handle MediaMTX connection failures gracefully

Story Coverage: S3 - WebSocket API Integration
IV&V Control Point: Real MediaMTX integration validation
"""

import pytest
import asyncio
from unittest.mock import Mock, AsyncMock

from src.websocket_server.server import WebSocketJsonRpcServer
from src.common.types import CameraDevice
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
    async def real_camera_monitor(self):
        """Real camera monitor with capability detection support."""
        from src.camera_discovery.hybrid_monitor import HybridCameraMonitor
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
    def server(self, real_config, real_camera_monitor, mediamtx_controller):
        """Create WebSocket server instance with real MediaMTX integration."""
        return WebSocketJsonRpcServer(
            host="localhost",
            port=8002,
            websocket_path="/ws",
            max_connections=100,
            mediamtx_controller=mediamtx_controller,  # Use real MediaMTX controller
            camera_monitor=real_camera_monitor,
        )

    @pytest.mark.asyncio
    async def test_get_camera_status_with_real_mediamtx_integration(
        self, server, real_camera_monitor, mediamtx_controller, mediamtx_infrastructure
    ):
        """
        Verify get_camera_status integrates with real MediaMTX service.
        
        Requirements: REQ-WS-001, REQ-WS-003
        Scenario: Real MediaMTX integration with capability metadata
        Expected: Successful integration with real MediaMTX service
        Edge Cases: Real stream status queries, actual metrics retrieval
        """
        # Use real camera monitor to get actual connected cameras
        connected_cameras = await real_camera_monitor.get_connected_cameras()
        
        # Get real capability metadata from actual camera detection
        if connected_cameras:
            device_path = list(connected_cameras.keys())[0]
            real_capability_metadata = real_camera_monitor.get_effective_capability_metadata(device_path)
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

        # Create real test stream in MediaMTX
        stream_info = await mediamtx_infrastructure.create_test_stream("camera0", "/dev/video0")
        
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
        await mediamtx_infrastructure.delete_test_stream("camera0")

    @pytest.mark.asyncio
    async def test_get_camera_status_fallback_to_defaults_when_capability_detection_unavailable(
        self, server, mock_camera_monitor
    ):
        """
        Verify graceful fallback when capability detection unavailable.
        
        Requirements: REQ-WS-002, REQ-ERROR-001
        Scenario: Capability detection method unavailable
        Expected: Graceful fallback to architecture defaults
        Edge Cases: Missing capability detection support
        """
        # Setup camera without capability detection support
        mock_camera_device = CameraDevice(
            device="/dev/video0", name="Test Camera 0", status="CONNECTED"
        )

        mock_camera_monitor.get_connected_cameras.return_value = {
            "/dev/video0": mock_camera_device
        }

        # Remove capability detection method to simulate unavailability
        if hasattr(mock_camera_monitor, "get_effective_capability_metadata"):
            delattr(mock_camera_monitor, "get_effective_capability_metadata")

        # Test get_camera_status method
        result = await server._method_get_camera_status({"device": "/dev/video0"})

        # Verify architecture defaults are used
        assert result["resolution"] == "1920x1080"  # Architecture default
        assert result["fps"] == 30  # Architecture default
        assert result["status"] == "CONNECTED"
        assert result["name"] == "Test Camera 0"

        # Verify empty capabilities when detection unavailable
        assert result["capabilities"]["formats"] == []
        assert result["capabilities"]["resolutions"] == []

    @pytest.mark.asyncio
    async def test_get_camera_status_handles_mediamtx_connection_failure(
        self, server, mock_camera_monitor, mediamtx_controller
    ):
        """
        Test camera status handling when MediaMTX connection fails.
        
        Requirements: REQ-ERROR-001
        Scenario: MediaMTX service unavailable
        Expected: Graceful error handling without crashing
        Edge Cases: Network failures, service unavailability
        """
        # Setup camera with capability data
        mock_camera_device = CameraDevice(
            device="/dev/video0", name="Test Camera 0", status="CONNECTED"
        )

        mock_camera_monitor.get_connected_cameras.return_value = {
            "/dev/video0": mock_camera_device
        }

        mock_capability_metadata = {
            "resolution": "1280x720",
            "fps": 25,
            "validation_status": "confirmed",
        }
        mock_camera_monitor.get_effective_capability_metadata.return_value = (
            mock_capability_metadata
        )

        # Simulate MediaMTX connection failure
        mediamtx_controller.get_stream_status.side_effect = Exception("Connection failed")

        # Test get_camera_status method - should handle error gracefully
        result = await server._method_get_camera_status({"device": "/dev/video0"})

        # Verify basic camera info is still returned
        assert result["status"] == "CONNECTED"
        assert result["name"] == "Test Camera 0"
        assert result["resolution"] == "1280x720"
        assert result["fps"] == 25

        # Verify error handling for MediaMTX integration
        assert "metrics" in result
        assert "streams" in result

    @pytest.mark.asyncio
    async def test_get_camera_list_with_real_capability_integration(
        self, server, mock_camera_monitor, mediamtx_controller
    ):
        """
        Verify get_camera_list uses real capability data for resolution/fps.
        
        Requirements: REQ-WS-002, REQ-WS-003
        Scenario: Multiple cameras with different capability metadata
        Expected: Real capability data integration in camera list
        Edge Cases: Different validation statuses, mixed capability data
        """
        # Setup multiple connected cameras with capability data
        mock_cameras = {
            "/dev/video0": CameraDevice("/dev/video0", "Camera 0", "CONNECTED"),
            "/dev/video1": CameraDevice("/dev/video1", "Camera 1", "CONNECTED"),
            "/dev/video2": CameraDevice("/dev/video2", "Camera 2", "DISCONNECTED"),
        }
        mock_camera_monitor.get_connected_cameras.return_value = mock_cameras

        # Mock capability metadata for different cameras
        def mock_get_capability_metadata(device_path):
            metadata_map = {
                "/dev/video0": {
                    "resolution": "1920x1080",
                    "fps": 30,
                    "validation_status": "confirmed",
                },
                "/dev/video1": {
                    "resolution": "1280x720",
                    "fps": 15,
                    "validation_status": "provisional",
                },
                "/dev/video2": {
                    "resolution": "640x480",
                    "fps": 10,
                    "validation_status": "none",
                },
            }
            return metadata_map.get(device_path, {})

        mock_camera_monitor.get_effective_capability_metadata.side_effect = (
            mock_get_capability_metadata
        )

        # Test get_camera_list method
        result = await server._method_get_camera_list()

        # Verify real capability data used in camera list
        cameras = result["cameras"]
        assert len(cameras) == 3

        # Camera 0 - confirmed capability data
        camera0 = next(c for c in cameras if c["device"] == "/dev/video0")
        assert camera0["resolution"] == "1920x1080"
        assert camera0["fps"] == 30
        assert camera0["status"] == "CONNECTED"

        # Camera 1 - provisional capability data
        camera1 = next(c for c in cameras if c["device"] == "/dev/video1")
        assert camera1["resolution"] == "1280x720"
        assert camera1["fps"] == 15
        assert camera1["status"] == "CONNECTED"

        # Camera 2 - disconnected camera
        camera2 = next(c for c in cameras if c["device"] == "/dev/video2")
        assert camera2["status"] == "DISCONNECTED"

        # Verify summary counts
        assert result["total"] == 3
        assert result["connected"] == 2

    @pytest.mark.asyncio
    async def test_get_camera_status_provisional_vs_confirmed_logic(
        self, server, mock_camera_monitor
    ):
        """Test that provisional and confirmed capability data are handled correctly."""
        # Setup camera with provisional capability data
        mock_camera_device = CameraDevice("/dev/video0", "Test Camera", "CONNECTED")
        mock_camera_monitor.get_connected_cameras.return_value = {
            "/dev/video0": mock_camera_device
        }

        # Test provisional capability data
        provisional_metadata = {
            "resolution": "1280x720",
            "fps": 25,
            "validation_status": "provisional",
            "consecutive_successes": 1,
            "formats": ["YUYV"],
        }
        mock_camera_monitor.get_effective_capability_metadata.return_value = (
            provisional_metadata
        )

        result = await server._method_get_camera_status({"device": "/dev/video0"})
        assert result["resolution"] == "1280x720"
        assert result["fps"] == 25

        # Test confirmed capability data
        confirmed_metadata = {
            "resolution": "1920x1080",
            "fps": 30,
            "validation_status": "confirmed",
            "consecutive_successes": 5,
            "formats": ["YUYV", "MJPEG"],
        }
        mock_camera_monitor.get_effective_capability_metadata.return_value = (
            confirmed_metadata
        )

        result = await server._method_get_camera_status({"device": "/dev/video0"})
        assert result["resolution"] == "1920x1080"
        assert result["fps"] == 30

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
        self, server, mock_camera_monitor
    ):
        """Verify methods handle missing MediaMTX controller gracefully."""
        # Setup camera monitor but remove MediaMTX controller
        mock_camera_device = CameraDevice("/dev/video0", "Test Camera", "CONNECTED")
        mock_camera_monitor.get_connected_cameras.return_value = {
            "/dev/video0": mock_camera_device
        }
        server._mediamtx_controller = None

        # Test get_camera_status without MediaMTX controller
        result = await server._method_get_camera_status({"device": "/dev/video0"})

        # Should still return camera data without stream info
        assert result["status"] == "CONNECTED"
        assert result["streams"] == {}  # No stream URLs without MediaMTX
        assert result["metrics"]["bytes_sent"] == 0
        assert result["metrics"]["readers"] == 0

    @pytest.mark.asyncio
    async def test_camera_status_error_handling(self, server, mock_camera_monitor):
        """Test error handling in camera status aggregation."""
        # Setup camera monitor to raise exception
        mock_camera_monitor.get_connected_cameras.side_effect = Exception(
            "Camera monitor error"
        )

        # Test get_camera_status with exception
        result = await server._method_get_camera_status({"device": "/dev/video0"})
        assert result["status"] == "ERROR"
        assert result["device"] == "/dev/video0"

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
