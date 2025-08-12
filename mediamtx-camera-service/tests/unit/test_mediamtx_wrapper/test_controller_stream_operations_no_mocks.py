# tests/unit/test_mediamtx_wrapper/test_controller_stream_operations_no_mocks.py
"""
Test stream creation/deletion with real aiohttp sessions and fake MediaMTX server.

Test policy: Verify idempotent operations, clear error contexts, and
reliability under transient failures using real async context managers.
"""

import pytest
import asyncio
import aiohttp
from aiohttp import web
from contextlib import asynccontextmanager
import socket

from src.mediamtx_wrapper.controller import MediaMTXController, StreamConfig


def get_free_port() -> int:
    """Get a free port for the test server."""
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
        s.bind(("127.0.0.1", 0))
        return s.getsockname()[1]


@asynccontextmanager
async def start_fake_mediamtx_server(host: str, port: int):
    """Start a fake MediaMTX server for testing."""
    calls = {"added": [], "deleted": [], "health": 0}

    async def health(request: web.Request):
        calls["health"] += 1
        return web.json_response({"serverVersion": "test", "serverUptime": 1})

    async def add_path(request: web.Request):
        name = request.match_info["name"]
        calls["added"].append(name)
        return web.json_response({"status": "ok"})

    async def delete_path(request: web.Request):
        name = request.match_info["name"]
        calls["deleted"].append(name)
        return web.json_response({"status": "ok"})

    app = web.Application()
    app.router.add_get("/v3/health", health)
    app.router.add_post("/v3/config/paths/add/{name}", add_path)
    app.router.add_post("/v3/config/paths/delete/{name}", delete_path)

    runner = web.AppRunner(app)
    await runner.setup()
    site = web.TCPSite(runner, host, port)
    await site.start()
    try:
        yield calls
    finally:
        await runner.cleanup()


class TestStreamOperationsNoMocks:
    """Test stream creation and deletion operations with real aiohttp sessions."""

    @pytest.fixture
    async def fake_mediamtx_server(self):
        """Create a fake MediaMTX server for testing."""
        port = get_free_port()
        async with start_fake_mediamtx_server("127.0.0.1", port) as calls:
            yield {"port": port, "calls": calls}

    @pytest.fixture
    async def controller(self, fake_mediamtx_server):
        """Create MediaMTX controller with real aiohttp session."""
        controller = MediaMTXController(
            host="127.0.0.1",
            api_port=fake_mediamtx_server["port"],
            rtsp_port=8554,
            webrtc_port=8889,
            hls_port=8888,
            config_path="/tmp/test_config.yml",
            recordings_path="/tmp/recordings",
            snapshots_path="/tmp/snapshots",
        )
        
        # Start the controller to create real aiohttp session
        await controller.start()
        try:
            yield controller
        finally:
            await controller.stop()

    @pytest.fixture
    def sample_stream_config(self):
        """Create sample stream configuration."""
        return StreamConfig(name="test_stream", source="/dev/video0", record=False)

    @pytest.mark.asyncio
    async def test_create_stream_success(self, controller, sample_stream_config, fake_mediamtx_server):
        """Test successful stream creation returns correct URLs."""
        # Create stream with real aiohttp session
        result = await controller.create_stream(sample_stream_config)

        # Verify URLs are correctly generated
        expected_urls = {
            "rtsp": "rtsp://127.0.0.1:8554/test_stream",
            "webrtc": "http://127.0.0.1:8889/test_stream",
            "hls": "http://127.0.0.1:8888/test_stream",
        }
        assert result == expected_urls
        
        # Verify the fake server received the request
        assert "test_stream" in fake_mediamtx_server["calls"]["added"]

    @pytest.mark.asyncio
    async def test_delete_stream_success(self, controller, fake_mediamtx_server):
        """Test successful stream deletion."""
        # Delete stream with real aiohttp session
        result = await controller.delete_stream("test_stream")

        assert result is True
        
        # Verify the fake server received the request
        assert "test_stream" in fake_mediamtx_server["calls"]["deleted"]

    @pytest.mark.asyncio
    async def test_create_stream_validation_errors(self, controller):
        """Test stream configuration validation."""
        # Test missing name
        with pytest.raises(ValueError, match="Stream name and source are required"):
            await controller.create_stream(StreamConfig(name="", source="/dev/video0"))

        # Test missing source
        with pytest.raises(ValueError, match="Stream name and source are required"):
            await controller.create_stream(StreamConfig(name="test", source=""))

    @pytest.mark.asyncio
    async def test_delete_stream_validation_error(self, controller):
        """Test stream name validation for deletion."""
        with pytest.raises(ValueError, match="Stream name is required"):
            await controller.delete_stream("")

    @pytest.mark.asyncio
    async def test_stream_operations_without_session(self):
        """Test operations fail gracefully when controller not started."""
        # Create controller but don't start it
        controller = MediaMTXController(
            host="127.0.0.1",
            api_port=9997,
            rtsp_port=8554,
            webrtc_port=8889,
            hls_port=8888,
            config_path="/tmp/test_config.yml",
            recordings_path="/tmp/recordings",
            snapshots_path="/tmp/snapshots",
        )

        sample_stream_config = StreamConfig(name="test_stream", source="/dev/video0", record=False)

        with pytest.raises(ConnectionError, match="MediaMTX controller not started"):
            await controller.create_stream(sample_stream_config)

        with pytest.raises(ConnectionError, match="MediaMTX controller not started"):
            await controller.delete_stream("test_stream")

    @pytest.mark.asyncio
    async def test_stream_config_with_recording(self, controller, fake_mediamtx_server):
        """Test stream configuration with recording enabled."""
        recording_config = StreamConfig(
            name="recording_stream",
            source="/dev/video1",
            record=True,
            record_path="/tmp/recordings/test.mp4",
        )

        result = await controller.create_stream(recording_config)

        # Verify URLs are correctly generated
        expected_urls = {
            "rtsp": "rtsp://127.0.0.1:8554/recording_stream",
            "webrtc": "http://127.0.0.1:8889/recording_stream",
            "hls": "http://127.0.0.1:8888/recording_stream",
        }
        assert result == expected_urls
        
        # Verify the fake server received the request
        assert "recording_stream" in fake_mediamtx_server["calls"]["added"]

    def test_generate_stream_urls_format(self, controller):
        """Test stream URL generation format."""
        urls = controller._generate_stream_urls("test_stream")

        expected_urls = {
            "rtsp": "rtsp://127.0.0.1:8554/test_stream",
            "webrtc": "http://127.0.0.1:8889/test_stream",
            "hls": "http://127.0.0.1:8888/test_stream",
        }
        assert urls == expected_urls

    @pytest.mark.asyncio
    async def test_real_async_context_manager_flow(self, controller, fake_mediamtx_server):
        """Test that real async context manager flow works correctly."""
        # This test demonstrates the real async with aiohttp.ClientSession() flow
        stream_config = StreamConfig(name="context_test", source="/dev/video0", record=False)
        
        # This should use the real async context manager in the controller
        result = await controller.create_stream(stream_config)
        
        # Verify the operation completed successfully
        assert result is not None
        assert "rtsp" in result
        assert "context_test" in result["rtsp"]
        
        # Verify the fake server received the request
        assert "context_test" in fake_mediamtx_server["calls"]["added"]
        
        # Now test deletion with real async context manager
        delete_result = await controller.delete_stream("context_test")
        assert delete_result is True
        
        # Verify the fake server received the delete request
        assert "context_test" in fake_mediamtx_server["calls"]["deleted"]
