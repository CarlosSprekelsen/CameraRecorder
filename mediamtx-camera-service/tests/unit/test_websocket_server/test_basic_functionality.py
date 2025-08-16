# tests/unit/test_websocket_server/test_basic_functionality.py
"""
Basic WebSocket server functionality test to isolate real system issues.

This test file focuses on basic functionality without complex async fixtures
to identify and fix real system issues that are preventing proper testing.
"""

import pytest
from src.websocket_server.server import WebSocketJsonRpcServer


class TestBasicWebSocketFunctionality:
    """Basic WebSocket server functionality tests."""

    def test_stream_name_generation(self):
        """Test basic stream name generation functionality."""
        # Create a simple server instance without complex setup
        server = WebSocketJsonRpcServer(
            host="localhost",
            port=8002,
            websocket_path="/ws",
            max_connections=100
        )
        
        # Test stream name generation
        assert server._get_stream_name_from_device_path("/dev/video0") == "camera0"
        assert server._get_stream_name_from_device_path("/dev/video1") == "camera1"
        assert server._get_stream_name_from_device_path("/dev/video99") == "camera99"
        
        # Test edge cases
        assert server._get_stream_name_from_device_path("") == "camera_unknown"
        assert server._get_stream_name_from_device_path(None) == "camera_unknown"
        assert server._get_stream_name_from_device_path("/invalid/path") != "camera_unknown"

    def test_filename_generation(self):
        """Test filename generation functionality."""
        server = WebSocketJsonRpcServer(
            host="localhost",
            port=8002,
            websocket_path="/ws",
            max_connections=100
        )
        
        # Test filename generation
        filename = server._generate_filename("/dev/video0", "jpg")
        assert filename.startswith("camera0_")
        assert filename.endswith(".jpg")
        
        # Test custom filename
        custom_filename = server._generate_filename("/dev/video0", "mp4", "test_video.mp4")
        assert custom_filename == "test_video.mp4"

    @pytest.mark.asyncio
    async def test_ping_method(self):
        """Test basic ping method functionality."""
        server = WebSocketJsonRpcServer(
            host="localhost",
            port=8002,
            websocket_path="/ws",
            max_connections=100
        )
        
        # Test ping method
        result = await server._method_ping()
        assert result == "pong"

    def test_server_initialization(self):
        """Test server initialization with various configurations."""
        # Test basic initialization
        server = WebSocketJsonRpcServer(
            host="localhost",
            port=8002,
            websocket_path="/ws",
            max_connections=100
        )
        
        assert server._host == "localhost"
        assert server._port == 8002
        assert server._websocket_path == "/ws"
        assert server._max_connections == 100
        
        # Test with different configuration
        server2 = WebSocketJsonRpcServer(
            host="127.0.0.1",
            port=9000,
            websocket_path="/api/ws",
            max_connections=50
        )
        
        assert server2._host == "127.0.0.1"
        assert server2._port == 9000
        assert server2._websocket_path == "/api/ws"
        assert server2._max_connections == 50
