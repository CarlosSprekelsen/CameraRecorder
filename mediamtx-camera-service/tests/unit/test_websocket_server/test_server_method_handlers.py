# tests/unit/test_websocket_server/test_server_method_handlers.py
"""
Test JSON-RPC method handlers in WebSocket server.

Covers method registration, version tracking, parameter validation,
and integration with backend services.
"""

import pytest
import asyncio
from unittest.mock import Mock, AsyncMock

from src.websocket_server.server import WebSocketJsonRpcServer


class TestServerMethodHandlers:
    """Test JSON-RPC method handler functionality."""

    @pytest.fixture
    def server(self):
        """Create WebSocket server instance for testing."""
        return WebSocketJsonRpcServer(
            host="localhost", port=8002, websocket_path="/ws", max_connections=100
        )

    def test_method_registration_and_versioning(self, server):
        """Test method registration with version tracking."""
        # Test built-in method registration
        server._register_builtin_methods()

        # Verify core methods are registered
        expected_methods = [
            "ping",
            "get_camera_list",
            "get_camera_status",
            "take_snapshot",
            "start_recording",
            "stop_recording",
        ]

        for method in expected_methods:
            assert method in server._method_handlers
            assert method in server._method_versions
            assert server._method_versions[method] == "1.0"

    def test_custom_method_registration(self, server):
        """Test registration of custom methods with versions."""

        async def custom_handler(params=None):
            return {"result": "custom"}

        # Register custom method with version
        server.register_method("custom_method", custom_handler, version="2.1")

        # Verify method is registered
        assert "custom_method" in server._method_handlers
        assert server.get_method_version("custom_method") == "2.1"

        # Test method unregistration
        server.unregister_method("custom_method")
        assert "custom_method" not in server._method_handlers

    @pytest.mark.asyncio
    async def test_ping_method(self, server):
        """Test ping method for health checks."""
        result = await server._method_ping()
        assert result == "pong"

        # Test with parameters (should be ignored)
        result = await server._method_ping({"test": "value"})
        assert result == "pong"

    def test_parameter_validation(self, server):
        """Test parameter validation in method handlers."""
        # Test methods requiring device parameter
        with pytest.raises(ValueError, match="device parameter is required"):
            asyncio.run(server._method_get_camera_status({}))

        with pytest.raises(ValueError, match="device parameter is required"):
            asyncio.run(server._method_take_snapshot(None))

        with pytest.raises(ValueError, match="device parameter is required"):
            asyncio.run(server._method_start_recording({}))

        with pytest.raises(ValueError, match="device parameter is required"):
            asyncio.run(server._method_stop_recording({}))

    @pytest.mark.asyncio
    async def test_take_snapshot_parameter_handling(self, server):
        """Test snapshot method parameter processing."""
        # Mock MediaMTX controller
        mock_controller = Mock()
        mock_controller.take_snapshot = AsyncMock(
            return_value={
                "filename": "test_snapshot.jpg",
                "file_size": 12345,
                "file_path": "/opt/snapshots/test_snapshot.jpg",
            }
        )
        server._mediamtx_controller = mock_controller

        # Test with device parameter only
        result = await server._method_take_snapshot({"device": "/dev/video0"})
        assert result["device"] == "/dev/video0"
        assert result["status"] == "completed"
        assert "filename" in result

        # Test with custom filename
        result = await server._method_take_snapshot(
            {"device": "/dev/video0", "filename": "custom_snapshot.jpg"}
        )
        assert result["filename"] == "test_snapshot.jpg"  # From mock controller

    @pytest.mark.asyncio
    async def test_recording_methods_parameter_handling(self, server):
        """Test recording method parameter processing."""
        # Mock MediaMTX controller
        mock_controller = Mock()
        mock_controller.start_recording = AsyncMock(
            return_value={
                "filename": "test_recording.mp4",
                "start_time": "2025-08-03T12:00:00Z",
            }
        )
        mock_controller.stop_recording = AsyncMock(
            return_value={
                "filename": "test_recording.mp4",
                "start_time": "2025-08-03T12:00:00Z",
                "duration": 3600,
                "file_size": 1073741824,
            }
        )
        server._mediamtx_controller = mock_controller

        # Test start_recording with parameters
        result = await server._method_start_recording(
            {"device": "/dev/video0", "duration": 3600, "format": "mp4"}
        )
        assert result["device"] == "/dev/video0"
        assert result["status"] == "STARTED"
        assert result["duration"] == 3600
        assert result["format"] == "mp4"

        # Test stop_recording
        result = await server._method_stop_recording({"device": "/dev/video0"})
        assert result["device"] == "/dev/video0"
        assert result["status"] == "STOPPED"
        assert result["duration"] == 3600

    @pytest.mark.asyncio
    async def test_method_error_handling_no_mediamtx(self, server):
        """Test method error handling when MediaMTX controller unavailable."""
        # Ensure no MediaMTX controller
        server._mediamtx_controller = None

        # Test snapshot without MediaMTX
        result = await server._method_take_snapshot({"device": "/dev/video0"})
        assert result["status"] == "FAILED"
        assert "MediaMTX controller not available" in result["error"]

        # Test start_recording without MediaMTX
        result = await server._method_start_recording({"device": "/dev/video0"})
        assert result["status"] == "FAILED"
        assert "MediaMTX controller not available" in result["error"]

        # Test stop_recording without MediaMTX
        result = await server._method_stop_recording({"device": "/dev/video0"})
        assert result["status"] == "FAILED"
        assert "MediaMTX controller not available" in result["error"]

    def test_filename_generation(self, server):
        """Test filename generation for recordings and snapshots."""
        # Test default filename generation
        filename = server._generate_filename("/dev/video0", "jpg")
        assert filename.startswith("camera0_")
        assert filename.endswith(".jpg")

        # Test custom filename without extension
        filename = server._generate_filename("/dev/video0", "mp4", "custom_recording")
        assert filename == "custom_recording.mp4"

        # Test custom filename with extension
        filename = server._generate_filename(
            "/dev/video0", "jpg", "custom_snapshot.jpg"
        )
        assert filename == "custom_snapshot.jpg"

    def test_stream_name_extraction(self, server):
        """Test stream name extraction from device paths."""
        # Test standard video device paths
        assert server._get_stream_name_from_device_path("/dev/video0") == "camera0"
        assert server._get_stream_name_from_device_path("/dev/video15") == "camera15"

        # Test non-standard device paths
        stream_name = server._get_stream_name_from_device_path("/custom/device")
        assert stream_name.startswith("camera_")
        assert stream_name != "camera_unknown"  # Should generate deterministic hash

        # Test empty/invalid paths
        assert server._get_stream_name_from_device_path("") == "camera_unknown"

    @pytest.mark.asyncio
    async def test_method_exception_handling(self, server):
        """Test exception handling in method execution."""
        # Mock MediaMTX controller that raises exceptions
        mock_controller = Mock()
        mock_controller.take_snapshot = AsyncMock(
            side_effect=Exception("MediaMTX error")
        )
        server._mediamtx_controller = mock_controller

        # Test that exceptions are caught and return error responses
        result = await server._method_take_snapshot({"device": "/dev/video0"})
        assert result["status"] == "FAILED"
        assert "error" in result

    def test_method_version_tracking(self, server):
        """Test method version tracking functionality."""

        # Register methods with different versions
        async def handler_v1():
            return {"version": "1.0"}

        async def handler_v2():
            return {"version": "2.0"}

        server.register_method("test_method", handler_v1, "1.0")
        assert server.get_method_version("test_method") == "1.0"

        # Update to new version
        server.register_method("test_method", handler_v2, "2.0")
        assert server.get_method_version("test_method") == "2.0"

        # Test non-existent method
        assert server.get_method_version("nonexistent_method") is None

    def test_server_stats_and_status(self, server):
        """Test server statistics and status reporting."""
        # Test initial stats
        stats = server.get_server_stats()
        assert stats["running"] is False
        assert stats["connected_clients"] == 0
        assert stats["max_connections"] == 100

        # Test after method registration
        server._register_builtin_methods()
        stats = server.get_server_stats()
        assert stats["registered_methods"] >= 6  # At least the built-in methods

        # Test connection count
        assert server.get_connection_count() == 0

        # Test is_running property
        assert server.is_running is False

    @pytest.mark.asyncio
    async def test_method_handlers_with_mock_dependencies(self, server):
        """Test method handlers with properly mocked dependencies."""
        # Setup mock camera monitor
        mock_camera_monitor = Mock()
        mock_camera_monitor.get_connected_cameras = AsyncMock(return_value={})
        server._camera_monitor = mock_camera_monitor

        # Setup mock MediaMTX controller
        mock_mediamtx = Mock()
        mock_mediamtx.get_stream_status = AsyncMock(return_value={"status": "inactive"})
        server._mediamtx_controller = mock_mediamtx

        # Test get_camera_list with mocked dependencies
        result = await server._method_get_camera_list()
        assert result["cameras"] == []
        assert result["total"] == 0
        assert result["connected"] == 0

        # Verify dependencies were called
        mock_camera_monitor.get_connected_cameras.assert_called_once()
