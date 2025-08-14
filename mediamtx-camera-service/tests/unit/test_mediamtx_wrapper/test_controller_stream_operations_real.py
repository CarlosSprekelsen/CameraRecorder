# tests/unit/test_mediamtx_wrapper/test_controller_stream_operations_real.py
"""
Real MediaMTX server integration tests for stream operations.

Test policy: Validate actual MediaMTX server integration, real stream creation/deletion,
and robust error handling with real HTTP API calls and process management.
"""

import pytest
import pytest_asyncio
import asyncio
import os
import tempfile
import subprocess
import socket
import time
from pathlib import Path
from typing import Dict, Any

from src.mediamtx_wrapper.controller import MediaMTXController, StreamConfig


class RealMediaMTXServer:
    """Real MediaMTX server for testing."""
    
    def __init__(self, api_port: int, rtsp_port: int, webrtc_port: int, hls_port: int):
        self.api_port = api_port
        self.rtsp_port = rtsp_port
        self.webrtc_port = webrtc_port
        self.hls_port = hls_port
        self.process: subprocess.Popen = None
        self.temp_dir: str = None
        self.config_file: str = None
        
    async def start(self) -> None:
        """Start real MediaMTX server process."""
        # Create temporary directory
        self.temp_dir = tempfile.mkdtemp(prefix="mediamtx_test_")
        
        # Create MediaMTX configuration
        self.config_file = os.path.join(self.temp_dir, "mediamtx.yml")
        self._create_config()
        
        # Start MediaMTX process
        cmd = ["mediamtx", self.config_file]
        print(f"Starting MediaMTX with command: {' '.join(cmd)}")
        print(f"Config file content:\n{open(self.config_file).read()}")
        
        self.process = subprocess.Popen(
            cmd,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            cwd=self.temp_dir
        )
        
        # Wait for server to be ready
        await self._wait_for_ready()
        
    async def stop(self) -> None:
        """Stop MediaMTX server process."""
        if self.process:
            self.process.terminate()
            try:
                self.process.wait(timeout=5)
            except subprocess.TimeoutExpired:
                self.process.kill()
                self.process.wait()
        
        # Clean up
        if self.temp_dir and os.path.exists(self.temp_dir):
            import shutil
            shutil.rmtree(self.temp_dir)
    
    def _create_config(self) -> None:
        """Create MediaMTX configuration file."""
        config = f"""api: yes
apiAddress: ":{self.api_port}"

rtspAddress: ":{self.rtsp_port}"
rtspTransports: [tcp, udp]

webrtcAddress: ":{self.webrtc_port}"

hlsAddress: ":{self.hls_port}"
hlsVariant: lowLatency

logLevel: info
logDestinations: [stdout]

paths:
  all:
    recordFormat: fmp4
    recordSegmentDuration: "3600s"
"""
        
        with open(self.config_file, 'w') as f:
            f.write(config)
    
    async def _wait_for_ready(self, timeout: float = 30.0) -> None:
        """Wait for MediaMTX server to be ready."""
        start_time = time.time()
        while time.time() - start_time < timeout:
            try:
                # Check if API port is listening
                sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
                sock.settimeout(1)
                result = sock.connect_ex(('127.0.0.1', self.api_port))
                sock.close()
                
                if result == 0:
                    # Test API health endpoint
                    import aiohttp
                    async with aiohttp.ClientSession() as session:
                        async with session.get(f"http://127.0.0.1:{self.api_port}/v3/health") as resp:
                            if resp.status == 200:
                                return
                
                await asyncio.sleep(1)
            except Exception:
                await asyncio.sleep(1)
        
        raise TimeoutError("MediaMTX server failed to start within timeout")


def get_free_port() -> int:
    """Get a free port for testing."""
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
        s.bind(('127.0.0.1', 0))
        return s.getsockname()[1]


class TestStreamOperationsReal:
    """Test stream operations with real MediaMTX server."""

    @pytest_asyncio.fixture
    async def mediamtx_server(self):
        """Create and manage real MediaMTX server."""
        api_port = get_free_port()
        rtsp_port = get_free_port()
        webrtc_port = get_free_port()
        hls_port = get_free_port()
        
        server = RealMediaMTXServer(api_port, rtsp_port, webrtc_port, hls_port)
        await server.start()
        try:
            yield server
        finally:
            await server.stop()

    @pytest_asyncio.fixture
    async def controller(self, mediamtx_server):
        """Create MediaMTX controller connected to real server."""
        controller = MediaMTXController(
            host="127.0.0.1",
            api_port=mediamtx_server.api_port,
            rtsp_port=mediamtx_server.rtsp_port,
            webrtc_port=mediamtx_server.webrtc_port,
            hls_port=mediamtx_server.hls_port,
            config_path="/tmp/test_config.yml",
            recordings_path="/tmp/recordings",
            snapshots_path="/tmp/snapshots",
        )
        
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
    async def test_create_stream_success(self, controller, sample_stream_config):
        """Test successful stream creation with real MediaMTX server."""
        # Create stream
        result = await controller.create_stream(sample_stream_config)
        
        # Verify result
        assert result is not None
        assert "rtsp" in result
        assert "webrtc" in result
        assert "hls" in result
        
        # Verify URLs are accessible
        assert f"rtsp://127.0.0.1:{controller._rtsp_port}/test_stream" in result["rtsp"]
        assert f"http://127.0.0.1:{controller._webrtc_port}/test_stream" in result["webrtc"]
        assert f"http://127.0.0.1:{controller._hls_port}/test_stream" in result["hls"]

    @pytest.mark.asyncio
    async def test_delete_stream_success(self, controller, sample_stream_config):
        """Test successful stream deletion with real MediaMTX server."""
        # Create stream first
        await controller.create_stream(sample_stream_config)
        
        # Delete stream
        result = await controller.delete_stream("test_stream")
        
        # Verify deletion was successful
        assert result is True

    @pytest.mark.asyncio
    async def test_create_stream_idempotent(self, controller, sample_stream_config):
        """Test that creating the same stream twice is idempotent."""
        # Create stream first time
        result1 = await controller.create_stream(sample_stream_config)
        
        # Create same stream second time
        result2 = await controller.create_stream(sample_stream_config)
        
        # Results should be identical
        assert result1 == result2

    @pytest.mark.asyncio
    async def test_delete_nonexistent_stream(self, controller):
        """Test deleting a stream that doesn't exist."""
        # Delete non-existent stream
        result = await controller.delete_stream("nonexistent_stream")
        
        # Should handle gracefully (return True or False, not raise exception)
        assert isinstance(result, bool)

    @pytest.mark.asyncio
    async def test_stream_config_with_recording(self, controller):
        """Test stream creation with recording enabled."""
        config = StreamConfig(
            name="recording_stream",
            source="/dev/video0",
            record=True
        )
        
        result = await controller.create_stream(config)
        
        # Verify result
        assert result is not None
        assert "rtsp" in result
        assert "webrtc" in result
        assert "hls" in result

    @pytest.mark.asyncio
    async def test_controller_creation(self):
        """Test that MediaMTX controller can be created without starting MediaMTX server."""
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
        
        # Test that controller can be created
        assert controller is not None
        assert controller._host == "127.0.0.1"
        assert controller._api_port == 9997

    @pytest.mark.asyncio
    async def test_stream_config_creation(self):
        """Test that StreamConfig can be created properly."""
        config = StreamConfig(
            name="test_stream",
            source="/dev/video0",
            record=True
        )
        
        assert config.name == "test_stream"
        assert config.source == "/dev/video0"
        assert config.record is True

    @pytest.mark.asyncio
    async def test_generate_stream_urls_format(self):
        """Test stream URL generation format without requiring MediaMTX server."""
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
        
        urls = controller._generate_stream_urls("test_stream")
        
        assert "rtsp" in urls
        assert "webrtc" in urls
        assert "hls" in urls
        
        # Verify URL formats
        assert urls["rtsp"].startswith("rtsp://")
        assert urls["webrtc"].startswith("http://")
        assert urls["hls"].startswith("http://")

    @pytest.mark.asyncio
    async def test_real_async_context_manager_flow(self, mediamtx_server):
        """Test real async context manager flow with MediaMTX controller."""
        controller = MediaMTXController(
            host="127.0.0.1",
            api_port=mediamtx_server.api_port,
            rtsp_port=mediamtx_server.rtsp_port,
            webrtc_port=mediamtx_server.webrtc_port,
            hls_port=mediamtx_server.hls_port,
            config_path="/tmp/test_config.yml",
            recordings_path="/tmp/recordings",
            snapshots_path="/tmp/snapshots",
        )
        
        async with controller:
            # Test that controller is working
            config = StreamConfig(name="context_test", source="/dev/video0", record=False)
            result = await controller.create_stream(config)
            assert result is not None
            assert "rtsp" in result
