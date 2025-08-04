# tests/unit/test_websocket_server/test_server_status_aggregation.py
"""
Test status aggregation functionality in WebSocket JSON-RPC server.

Covers get_camera_status and get_camera_list methods with real data integration,
capability detection, and graceful degradation.
"""

import pytest
import asyncio
from unittest.mock import Mock, AsyncMock

from src.websocket_server.server import WebSocketJsonRpcServer
from src.common.types import CameraDevice


class TestServerStatusAggregation:
    """Test camera status aggregation with real data sources."""

    @pytest.fixture
    def mock_config(self):
        """Mock configuration for testing."""
        config = Mock()
        config.server.host = "localhost"
        config.server.port = 8002
        config.server.websocket_path = "/ws"
        config.server.max_connections = 100
        return config

    @pytest.fixture
    def mock_camera_monitor(self):
        """Mock camera monitor with capability detection support."""
        monitor = Mock()
        monitor.get_connected_cameras = AsyncMock()
        monitor.get_effective_capability_metadata = Mock()
        return monitor

    @pytest.fixture
    def mock_mediamtx_controller(self):
        """Mock MediaMTX controller."""
        controller = Mock()
        controller.get_stream_status = AsyncMock()
        return controller

    @pytest.fixture
    def server(self, mock_config, mock_camera_monitor, mock_mediamtx_controller):
        """Create WebSocket server instance for testing."""
        return WebSocketJsonRpcServer(
            host="localhost",
            port=8002,
            websocket_path="/ws",
            max_connections=100,
            mediamtx_controller=mock_mediamtx_controller,
            camera_monitor=mock_camera_monitor,
        )

    @pytest.mark.asyncio
    async def test_get_camera_status_uses_real_capability_data(
        self, server, mock_camera_monitor, mock_mediamtx_controller
    ):
        """Verify get_camera_status integrates real capability metadata when available."""
        # Setup connected cameras with real capability data
        mock_camera_device = CameraDevice(
            device="/dev/video0", name="Test Camera 0", status="CONNECTED"
        )

        mock_camera_monitor.get_connected_cameras.return_value = {
            "/dev/video0": mock_camera_device
        }

        # Mock capability metadata with provisional/confirmed data
        mock_capability_metadata = {
            "resolution": "1280x720",
            "fps": 25,
            "validation_status": "confirmed",
            "formats": ["YUYV", "MJPEG"],
            "all_resolutions": ["1920x1080", "1280x720", "640x480"],
            "consecutive_successes": 3,
        }
        mock_camera_monitor.get_effective_capability_metadata.return_value = (
            mock_capability_metadata
        )

        # Mock MediaMTX stream status
        mock_mediamtx_controller.get_stream_status.return_value = {
            "status": "active",
            "bytes_sent": 12345,
            "readers": 2,
        }

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

        # Verify MediaMTX integration
        assert result["metrics"]["bytes_sent"] == 12345
        assert result["metrics"]["readers"] == 2
        assert result["streams"]["rtsp"] == "rtsp://localhost:8554/camera0"

        # Verify method calls
        mock_camera_monitor.get_connected_cameras.assert_called_once()
        mock_camera_monitor.get_effective_capability_metadata.assert_called_once_with(
            "/dev/video0"
        )
        mock_mediamtx_controller.get_stream_status.assert_called_once_with("camera0")

    @pytest.mark.asyncio
    async def test_get_camera_status_fallback_to_defaults(
        self, server, mock_camera_monitor
    ):
        """Verify graceful fallback when capability detection unavailable."""
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
    async def test_get_camera_list_capability_integration(
        self, server, mock_camera_monitor, mock_mediamtx_controller
    ):
        """Verify get_camera_list uses real capability data for resolution/fps."""
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

        # Mock MediaMTX stream status
        mock_mediamtx_controller.get_stream_status.return_value = {"status": "active"}

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
