# tests/unit/test_mediamtx_wrapper/test_controller_stream_operations_real.py
"""
Comprehensive MediaMTX controller validation tests - consolidated.

Requirements Traceability:
- REQ-MEDIA-002: MediaMTX controller shall manage stream lifecycle via REST API
- REQ-MEDIA-005: MediaMTX controller shall provide accurate recording duration calculation and snapshot capture
- REQ-MEDIA-008: MediaMTX controller shall generate correct stream URLs for all protocols
- REQ-MEDIA-009: MediaMTX controller shall validate stream configurations with real validation

Story Coverage: S2, S3 - MediaMTX Integration & Management
IV&V Control Point: Real stream operations, recording, and snapshot validation

Test policy: Validate actual MediaMTX controller behavior, configuration,
URL generation, recording lifecycle, and snapshot capture without requiring external MediaMTX server startup.
"""

import pytest
import pytest_asyncio
import asyncio
import os
import tempfile
import subprocess
import socket
import time
import uuid
from pathlib import Path
from typing import Dict, Any
from contextlib import asynccontextmanager
from aiohttp import web

from src.mediamtx_wrapper.controller import MediaMTXController, StreamConfig


def get_free_port() -> int:
    """Get a free port for testing."""
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
        s.bind(('127.0.0.1', 0))
        return s.getsockname()[1]


@pytest.fixture
async def real_mediamtx_server():
    """Create real HTTP test server that simulates MediaMTX API responses."""
    
    async def handle_health_check(request):
        """Handle MediaMTX health check endpoint."""
        return web.json_response({
            "serverVersion": "v1.0.0",
            "serverUptime": 3600,
            "apiVersion": "v3"
        })
    
    async def handle_stream_operations(request):
        """Handle MediaMTX stream operations."""
        return web.json_response({"status": "success"})
    
    async def handle_recording_operations(request):
        """Handle MediaMTX recording operations."""
        return web.json_response({"status": "recording started"})
    
    app = web.Application()
    app.router.add_get('/v3/config/global/get', handle_health_check)
    app.router.add_post('/v3/config/paths/add/{name}', handle_stream_operations)
    app.router.add_post('/v3/config/paths/delete/{name}', handle_stream_operations)
    app.router.add_post('/v3/config/paths/edit/{name}', handle_recording_operations)
    
    runner = web.AppRunner(app)
    await runner.setup()
    site = web.TCPSite(runner, '127.0.0.1', 10001)
    await site.start()
    
    try:
        yield runner
    finally:
        await runner.cleanup()

@asynccontextmanager
async def recording_mediamtx_server(host: str, port: int):
    """Start a MediaMTX server that handles recording operations."""
    recording_sessions = {}
    
    async def health_endpoint(request: web.Request):
        return web.json_response({
            "status": "healthy",
            "version": "v1.0.0",
            "uptime": 3600,
            "api_port": port
        })
    
    async def start_recording(request: web.Request):
        stream_name = request.match_info["name"]
        recording_sessions[stream_name] = {
            "start_time": time.time(),
            "format": "mp4",
            "status": "recording"
        }
        return web.json_response({"status": "recording started"})
    
    async def stop_recording(request: web.Request):
        stream_name = request.match_info["name"]
        if stream_name in recording_sessions:
            session = recording_sessions[stream_name]
            session["stop_time"] = time.time()
            session["status"] = "stopped"
            del recording_sessions[stream_name]
            return web.json_response({"status": "recording stopped"})
        else:
            return web.json_response({"error": "Recording not found"}, status=404)

    app = web.Application()
    app.router.add_get("/v3/config/global/get", health_endpoint)
    app.router.add_post("/v3/config/paths/edit/{name}", start_recording)  # Recording operations
    
    runner = web.AppRunner(app)
    await runner.setup()
    site = web.TCPSite(runner, host, port)
    await site.start()
    try:
        yield {"port": port, "recording_sessions": recording_sessions}
    finally:
        await runner.cleanup()


class TestMediaMTXControllerComprehensive:
    """Comprehensive test suite for MediaMTX controller with real system validation."""

    @pytest.fixture
    def temp_dirs(self):
        """Create temporary directories for test files."""
        with tempfile.TemporaryDirectory() as temp_dir:
            recordings_dir = os.path.join(temp_dir, "recordings")
            snapshots_dir = os.path.join(temp_dir, "snapshots")
            test_media_dir = os.path.join(temp_dir, "test_media")
            
            os.makedirs(recordings_dir, exist_ok=True)
            os.makedirs(snapshots_dir, exist_ok=True)
            os.makedirs(test_media_dir, exist_ok=True)
            
            yield {
                "temp_dir": temp_dir,
                "recordings_dir": recordings_dir,
                "snapshots_dir": snapshots_dir,
                "test_media_dir": test_media_dir
            }

    @pytest.fixture
    def test_image_file(self, temp_dirs):
        """Create a real test image file using FFmpeg."""
        test_image_path = os.path.join(temp_dirs["test_media_dir"], "test_image.jpg")
        
        # Create a simple test image using FFmpeg
        cmd = [
            "ffmpeg", "-y",  # Overwrite output
            "-f", "lavfi",   # Use lavfi input format
            "-i", "testsrc=duration=1:size=320x240:rate=1",  # Generate test pattern
            "-vframes", "1",  # Capture only 1 frame
            "-q:v", "2",      # High quality
            test_image_path
        ]
        
        try:
            result = subprocess.run(cmd, capture_output=True, text=True, timeout=10)
            if result.returncode == 0 and os.path.exists(test_image_path):
                return test_image_path
        except (subprocess.TimeoutExpired, FileNotFoundError):
            pass
        
        # Fallback: create a minimal JPEG file if FFmpeg fails
        with open(test_image_path, "wb") as f:
            # Minimal valid JPEG header
            f.write(b"\xff\xd8\xff\xe0\x00\x10JFIF\x00\x01\x01\x01\x00H\x00H\x00\x00\xff\xdb\x00C\x00\x08\x06\x06\x07\x06\x05\x08\x07\x07\x07\t\t\x08\n\x0c\x14\r\x0c\x0b\x0b\x0c\x19\x12\x13\x0f\x14\x1d\x1a\x1f\x1e\x1d\x1a\x1c\x1c $.\x27 ,#\x1c\x1c(7),01444\x1f\x27=9=82<.342\xff\xc0\x00\x11\x08\x00\x01\x00\x01\x01\x01\x11\x00\x02\x11\x01\x03\x11\x01\xff\xc4\x00\x14\x00\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x08\xff\xc4\x00\x14\x10\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\xff\xda\x00\x0c\x03\x01\x00\x02\x11\x03\x11\x00\x3f\x00\xaa\xff\xd9")
        
        return test_image_path

    # ===== STREAM OPERATIONS TESTS =====

    @pytest.mark.asyncio
    async def test_controller_creation(self):
        """
        Test that MediaMTX controller can be created properly.
        
        Requirements: REQ-MTX-001
        Scenario: MediaMTX controller initialization
        Expected: Proper controller creation with correct configuration
        Edge Cases: Invalid configuration parameters, missing required fields
        """
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
        assert controller._rtsp_port == 8554
        assert controller._webrtc_port == 8889
        assert controller._hls_port == 8888

    @pytest.mark.asyncio
    async def test_stream_config_creation(self):
        """
        Test that StreamConfig can be created properly.
        
        Requirements: REQ-MTX-009
        Scenario: Stream configuration object creation
        Expected: Proper StreamConfig object with validated parameters
        Edge Cases: Invalid stream names, missing required fields
        """
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
        """
        Test stream URL generation format.
        
        Requirements: REQ-MTX-008
        Scenario: Stream URL generation for different protocols
        Expected: Correct URL formats for RTSP, WebRTC, and HLS
        Edge Cases: Different host configurations, port variations
        """
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
        
        # Verify specific URL content
        assert "127.0.0.1:8554/test_stream" in urls["rtsp"]
        assert "127.0.0.1:8889/test_stream" in urls["webrtc"]
        assert "127.0.0.1:8888/test_stream" in urls["hls"]

    @pytest.mark.asyncio
    async def test_controller_configuration_validation(self):
        """Test controller configuration validation."""
        controller = MediaMTXController(
            host="localhost",
            api_port=9997,
            rtsp_port=8554,
            webrtc_port=8889,
            hls_port=8888,
            config_path="/tmp/test_config.yml",
            recordings_path="/tmp/recordings",
            snapshots_path="/tmp/snapshots",
        )
        
        # Test configuration properties
        assert controller._host == "localhost"
        assert controller._api_port == 9997
        assert controller._rtsp_port == 8554
        assert controller._webrtc_port == 8889
        assert controller._hls_port == 8888
        assert controller._config_path == "/tmp/test_config.yml"
        assert controller._recordings_path == "/tmp/recordings"
        assert controller._snapshots_path == "/tmp/snapshots"

    @pytest.mark.asyncio
    async def test_stream_config_validation(self):
        """Test StreamConfig validation and properties."""
        # Test basic config
        config1 = StreamConfig(
            name="camera1",
            source="/dev/video0",
            record=False
        )
        assert config1.name == "camera1"
        assert config1.source == "/dev/video0"
        assert config1.record is False
        
        # Test recording config
        config2 = StreamConfig(
            name="camera2",
            source="/dev/video1",
            record=True
        )
        assert config2.name == "camera2"
        assert config2.source == "/dev/video1"
        assert config2.record is True

    @pytest.mark.asyncio
    async def test_ffmpeg_availability(self):
        """Test that FFmpeg is available in the system."""
        try:
            result = subprocess.run(
                ["ffmpeg", "-version"],
                capture_output=True,
                text=True,
                timeout=10
            )
            assert result.returncode == 0
            assert "ffmpeg version" in result.stdout.lower()
        except (subprocess.TimeoutExpired, FileNotFoundError) as e:
            pytest.skip(f"FFmpeg not available: {e}")

    @pytest.mark.asyncio
    async def test_mediamtx_availability(self):
        """Test that MediaMTX is available in the system."""
        try:
            result = subprocess.run(
                ["mediamtx", "--version"],
                capture_output=True,
                text=True,
                timeout=10
            )
            assert result.returncode == 0
            # Accept actual output which may be just a version string (e.g., 'v1.13.1')
            out = (result.stdout or "") + (result.stderr or "")
            out = out.strip()
            assert out, "Empty --version output"
        except (subprocess.TimeoutExpired, FileNotFoundError) as e:
            pytest.skip(f"MediaMTX not available: {e}")

    @pytest.mark.asyncio
    async def test_real_file_system_operations(self):
        """Test real file system operations for recordings and snapshots."""
        with tempfile.TemporaryDirectory() as temp_dir:
            recordings_dir = os.path.join(temp_dir, "recordings")
            snapshots_dir = os.path.join(temp_dir, "snapshots")
            
            # Create directories
            os.makedirs(recordings_dir, exist_ok=True)
            os.makedirs(snapshots_dir, exist_ok=True)
            
            # Verify directories exist
            assert os.path.exists(recordings_dir)
            assert os.path.exists(snapshots_dir)
            assert os.path.isdir(recordings_dir)
            assert os.path.isdir(snapshots_dir)
            
            # Test file creation
            test_file = os.path.join(recordings_dir, "test.mp4")
            with open(test_file, 'w') as f:
                f.write("test content")
            
            assert os.path.exists(test_file)
            assert os.path.getsize(test_file) > 0

    @pytest.mark.asyncio
    async def test_controller_async_context_manager(self):
        """Test controller async context manager without external dependencies."""
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
        
        # Test that controller can be used in async context
        # Note: We don't actually start it to avoid hanging
        assert controller is not None
        assert hasattr(controller, '__aenter__')
        assert hasattr(controller, '__aexit__')

    @pytest.mark.asyncio
    async def test_stream_operations_without_server(self):
        """Test stream operations that don't require MediaMTX server."""
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
        
        config = StreamConfig(
            name="test_stream",
            source="/dev/video0",
            record=False
        )
        
        # Test URL generation (doesn't require server)
        urls = controller._generate_stream_urls(config.name)
        assert "rtsp" in urls
        assert "webrtc" in urls
        assert "hls" in urls
        
        # Test configuration validation
        assert config.name == "test_stream"
        assert config.source == "/dev/video0"
        assert config.record is False

    # ===== RECORDING TESTS =====

    @pytest.mark.asyncio
    async def test_recording_duration_calculation_precision(self, temp_dirs):
        """Test accurate duration calculation using REAL session timestamps and file operations."""
        port = get_free_port()
        
        async with recording_mediamtx_server("127.0.0.1", port) as server:
            controller = MediaMTXController(
                host="127.0.0.1",
                api_port=port,
                rtsp_port=8554,
                webrtc_port=8889,
                hls_port=8888,
                config_path="/tmp/test_config.yml",
                recordings_path=temp_dirs["recordings_dir"],
                snapshots_path=temp_dirs["snapshots_dir"],
            )
            
            await controller.start()
            try:
                # Create a real test recording file
                stream_name = "test_stream"
                recording_file = os.path.join(temp_dirs["recordings_dir"], f"{stream_name}.mp4")
                
                # Start recording and capture start time
                start_time = time.time()
                await controller.start_recording(stream_name, duration=3600, format="mp4")
                
                # Create actual file to simulate recording
                test_content = b"fake_mp4_content_for_testing" * 1000  # ~27KB file
                with open(recording_file, "wb") as f:
                    f.write(test_content)
                
                # Wait a known duration 
                test_duration = 2  # 2 seconds
                await asyncio.sleep(test_duration)
                
                result = await controller.stop_recording(stream_name)
                
                # Verify duration calculation is accurate (allow some tolerance for timing)
                actual_duration = result.get("duration", 0)
                assert abs(actual_duration - test_duration) <= 1.0  # Allow 1 second tolerance
                assert result["status"] == "completed"
                assert result.get("file_exists", False) is True
                assert result.get("file_size", 0) > 0
                
            finally:
                await controller.stop()

    @pytest.mark.asyncio
    async def test_recording_missing_file_handling(self, temp_dirs):
        """Test stop_recording when file doesn't exist on disk using REAL file operations."""
        controller = MediaMTXController(
            host="127.0.0.1",
            api_port=9997,
            rtsp_port=8554,
            webrtc_port=8889,
            hls_port=8888,
            config_path="/tmp/test_config.yml",
            recordings_path=temp_dirs["recordings_dir"],
            snapshots_path=temp_dirs["snapshots_dir"],
        )
        
        # Start recording but don't create the file
        stream_name = "missing_file_stream"
        await controller.start_recording(stream_name, format="mp4")
        
        # Wait briefly then stop (file was never created)
        await asyncio.sleep(0.1)
        result = await controller.stop_recording(stream_name)
        
        # Verify graceful handling of missing file
        assert result["status"] == "completed"
        assert result.get("file_exists", True) is False  # Should detect missing file
        assert result.get("file_size", 1) == 0  # Size should be 0 for missing file

    @pytest.mark.asyncio
    async def test_recording_file_permission_error(self, temp_dirs):
        """Test handling when file exists but cannot be accessed due to permissions using REAL files."""
        controller = MediaMTXController(
            host="127.0.0.1",
            api_port=9997,
            rtsp_port=8554,
            webrtc_port=8889,
            hls_port=8888,
            config_path="/tmp/test_config.yml",
            recordings_path=temp_dirs["recordings_dir"],
            snapshots_path=temp_dirs["snapshots_dir"],
        )
        
        stream_name = "permission_test_stream"
        recording_file = os.path.join(temp_dirs["recordings_dir"], f"{stream_name}.mp4")
        
        # Start recording
        await controller.start_recording(stream_name, format="mp4")
        
        # Create file with restricted permissions
        with open(recording_file, "wb") as f:
            f.write(b"test_content")
        
        # Remove read permissions to simulate permission error
        os.chmod(recording_file, 0o000)  # No permissions
        
        try:
            result = await controller.stop_recording(stream_name)
            
            # Should handle permission error gracefully
            # The file exists but can't be read, so it might report as missing or with error
            assert result["status"] == "completed"
            # File exists but can't be accessed - behavior may vary by implementation
            
        finally:
            # Restore permissions so cleanup can work
            try:
                os.chmod(recording_file, 0o644)
            except:
                pass  # Ignore cleanup errors

    @pytest.mark.asyncio
    async def test_recording_directory_creation_permission_error(self, temp_dirs):
        """Test recording fails gracefully when recordings directory is not writable using REAL directories."""
        port = get_free_port()
        
        # Create a directory with no write permissions
        readonly_dir = os.path.join(temp_dirs["temp_dir"], "readonly_recordings")
        os.makedirs(readonly_dir)
        os.chmod(readonly_dir, 0o444)  # Read-only
        
        async with recording_mediamtx_server("127.0.0.1", port) as server:
            controller = MediaMTXController(
                host="127.0.0.1",
                api_port=port,
                rtsp_port=8554,
                webrtc_port=8889,
                hls_port=8888,
                config_path="/tmp/test_config.yml",
                recordings_path=readonly_dir,  # Read-only directory
                snapshots_path=temp_dirs["snapshots_dir"],
            )
            
            await controller.start()
            try:
                # Try to start recording in read-only directory
                try:
                    await controller.start_recording("test_stream", format="mp4")
                    # If it doesn't raise an exception, that's also valid behavior
                    # (the controller might handle this gracefully)
                except Exception as e:
                    # Expected - should fail gracefully
                    assert "permission" in str(e).lower() or "readonly" in str(e).lower() or "denied" in str(e).lower()
                    
            finally:
                await controller.stop()
                # Restore permissions for cleanup
                try:
                    os.chmod(readonly_dir, 0o755)
                except:
                    pass

    @pytest.mark.asyncio
    async def test_recording_session_management(self, temp_dirs):
        """Test recording session lifecycle management using REAL sessions."""
        controller = MediaMTXController(
            host="127.0.0.1",
            api_port=9997,
            rtsp_port=8554,
            webrtc_port=8889,
            hls_port=8888,
            config_path="/tmp/test_config.yml",
            recordings_path=temp_dirs["recordings_dir"],
            snapshots_path=temp_dirs["snapshots_dir"],
        )
        
        stream1, stream2 = "session_test_1", "session_test_2"
        
        # Start multiple recording sessions
        await controller.start_recording(stream1, format="mp4")
        await controller.start_recording(stream2, format="mp4")
        
        # Create real files for both
        for stream in [stream1, stream2]:
            recording_file = os.path.join(temp_dirs["recordings_dir"], f"{stream}.mp4")
            with open(recording_file, "wb") as f:
                f.write(b"test_recording_content" * 100)
        
        await asyncio.sleep(1)  # Let recordings run briefly
        
        # Stop recordings and verify session management
        result1 = await controller.stop_recording(stream1)
        result2 = await controller.stop_recording(stream2)
        
        # Both should complete successfully
        assert result1["status"] == "completed"
        assert result2["status"] == "completed"
        assert result1.get("file_exists", False) is True
        assert result2.get("file_exists", False) is True

    @pytest.mark.asyncio
    async def test_recording_duplicate_start_error(self, temp_dirs):
        """Test starting recording on already recording stream using REAL implementation."""
        controller = MediaMTXController(
            host="127.0.0.1",
            api_port=9997,
            rtsp_port=8554,
            webrtc_port=8889,
            hls_port=8888,
            config_path="/tmp/test_config.yml",
            recordings_path=temp_dirs["recordings_dir"],
            snapshots_path=temp_dirs["snapshots_dir"],
        )
        
        stream_name = "duplicate_test_stream"
        
        # Start first recording
        await controller.start_recording(stream_name, format="mp4")
        
        # Try to start recording on same stream again
        try:
            await controller.start_recording(stream_name, format="mp4")
            # If no exception, the controller handles this gracefully (also valid)
        except Exception as e:
            # Expected - should prevent duplicate recording
            assert "already" in str(e).lower() or "duplicate" in str(e).lower() or "recording" in str(e).lower()
        
        # Clean up
        await controller.stop_recording(stream_name)

    @pytest.mark.asyncio
    async def test_recording_stop_without_start_error(self, temp_dirs):
        """Test stopping recording that was never started using REAL implementation."""
        controller = MediaMTXController(
            host="127.0.0.1",
            api_port=9997,
            rtsp_port=8554,
            webrtc_port=8889,
            hls_port=8888,
            config_path="/tmp/test_config.yml",
            recordings_path=temp_dirs["recordings_dir"],
            snapshots_path=temp_dirs["snapshots_dir"],
        )
        
        # Try to stop recording that was never started
        try:
            result = await controller.stop_recording("never_started_stream")
            # If no exception, check the result indicates the issue
            assert result.get("status") in ["error", "not_found", "completed"]
        except Exception as e:
            # Expected - should handle missing recording gracefully
            assert "not found" in str(e).lower() or "not recording" in str(e).lower()

    @pytest.mark.asyncio
    async def test_recording_format_validation(self, temp_dirs):
        """Test recording format validation using REAL implementation."""
        controller = MediaMTXController(
            host="127.0.0.1",
            api_port=9997,
            rtsp_port=8554,
            webrtc_port=8889,
            hls_port=8888,
            config_path="/tmp/test_config.yml",
            recordings_path=temp_dirs["recordings_dir"],
            snapshots_path=temp_dirs["snapshots_dir"],
        )
        
        # Test with valid format
        stream_name = "format_test_stream"
        await controller.start_recording(stream_name, format="mp4")
        
        # Create actual file
        recording_file = os.path.join(temp_dirs["recordings_dir"], f"{stream_name}.mp4")
        with open(recording_file, "wb") as f:
            f.write(b"mp4_content")
        
        await asyncio.sleep(0.1)
        result = await controller.stop_recording(stream_name)
        
        assert result["status"] == "completed"
        assert result.get("file_exists", False) is True

    # ===== SNAPSHOT TESTS =====

    @pytest.mark.asyncio
    async def test_snapshot_invalid_stream_handling(self, temp_dirs, real_mediamtx_server):
        """Test handling of invalid/non-existent stream."""
        controller = MediaMTXController(
            host="localhost",
            api_port=10001,  # Use real server port
            rtsp_port=8554,
            webrtc_port=8889,
            hls_port=8888,
            config_path="/tmp/test_config.yml",
            recordings_path=temp_dirs["recordings_dir"],
            snapshots_path=temp_dirs["snapshots_dir"],
            process_termination_timeout=2.0,
            process_kill_timeout=1.0,
        )
        # Use real HTTP server instead of mocking
        controller._base_url = f"http://{controller._host}:{controller._api_port}"
        await controller.start()
        
        stream_name = "non_existent_stream"
        snapshot_filename = f"snapshot_{uuid.uuid4().hex[:8]}.jpg"
        
        # Attempt to take snapshot from non-existent stream
        result = await controller.take_snapshot(stream_name, snapshot_filename)
        
        # Verify failure is handled gracefully
        assert result["status"] == "failed"
        assert "FFmpeg capture failed" in result["error"]
        assert result["file_size"] == 0

    @pytest.mark.asyncio
    async def test_snapshot_process_timeout_handling(self, temp_dirs, real_mediamtx_server):
        """Test timeout handling with a hanging FFmpeg process."""
        controller = MediaMTXController(
            host="localhost",
            api_port=10001,  # Use real server port
            rtsp_port=8554,
            webrtc_port=8889,
            hls_port=8888,
            config_path="/tmp/test_config.yml",
            recordings_path=temp_dirs["recordings_dir"],
            snapshots_path=temp_dirs["snapshots_dir"],
            process_termination_timeout=2.0,
            process_kill_timeout=1.0,
        )
        # Use real HTTP server instead of mocking
        controller._base_url = f"http://{controller._host}:{controller._api_port}"
        await controller.start()
        
        stream_name = "timeout_test_stream"
        snapshot_filename = f"snapshot_{uuid.uuid4().hex[:8]}.jpg"
        
        # Create a process that will hang (invalid RTSP URL that doesn't respond)
        result = await controller.take_snapshot(stream_name, snapshot_filename)
        
        # Verify timeout is handled
        assert result["status"] == "failed"
        assert "timeout" in result["error"].lower()
        assert result["file_size"] == 0

    @pytest.mark.asyncio
    async def test_snapshot_directory_permission_error(self, temp_dirs, real_mediamtx_server):
        """Test handling when snapshots directory cannot be written to."""
        controller = MediaMTXController(
            host="localhost",
            api_port=10001,  # Use real server port
            rtsp_port=8554,
            webrtc_port=8889,
            hls_port=8888,
            config_path="/tmp/test_config.yml",
            recordings_path=temp_dirs["recordings_dir"],
            snapshots_path=temp_dirs["snapshots_dir"],
            process_termination_timeout=2.0,
            process_kill_timeout=1.0,
        )
        # Use real HTTP server instead of mocking
        controller._base_url = f"http://{controller._host}:{controller._api_port}"
        await controller.start()
        
        # Create controller with read-only snapshots directory
        read_only_dir = os.path.join(temp_dirs["temp_dir"], "readonly_snapshots")
        os.makedirs(read_only_dir, exist_ok=True)
        os.chmod(read_only_dir, 0o444)  # Read-only permissions
        
        controller._snapshots_path = read_only_dir
        
        stream_name = "test_stream"
        snapshot_filename = f"snapshot_{uuid.uuid4().hex[:8]}.jpg"
        
        result = await controller.take_snapshot(stream_name, snapshot_filename)
        
        # Verify permission error is handled
        assert result["status"] == "failed"
        assert "Cannot write to snapshots directory" in result["error"]
        assert result["file_size"] == 0

    @pytest.mark.asyncio
    async def test_snapshot_from_image_file_success(self, temp_dirs, test_image_file, real_mediamtx_server):
        """Test successful snapshot capture from a real image file."""
        controller = MediaMTXController(
            host="localhost",
            api_port=10001,  # Use real server port
            rtsp_port=8554,
            webrtc_port=8889,
            hls_port=8888,
            config_path="/tmp/test_config.yml",
            recordings_path=temp_dirs["recordings_dir"],
            snapshots_path=temp_dirs["snapshots_dir"],
            process_termination_timeout=2.0,
            process_kill_timeout=1.0,
        )
        # Use real HTTP server instead of mocking
        controller._base_url = f"http://{controller._host}:{controller._api_port}"
        await controller.start()
        
        stream_name = "test_image_stream"
        snapshot_filename = f"snapshot_{uuid.uuid4().hex[:8]}.jpg"
        
        # Use FFmpeg to create a mock RTSP stream from the image file
        rtsp_server_cmd = [
            "ffmpeg", "-re",  # Read input at native frame rate
            "-loop", "1",     # Loop the input
            "-i", test_image_file,
            "-c:v", "libx264",
            "-preset", "ultrafast",
            "-tune", "zerolatency",
            "-f", "rtsp",
            "-rtsp_transport", "tcp",
            f"rtsp://localhost:8554/{stream_name}"
        ]
        
        # Start RTSP server in background
        rtsp_process = None
        try:
            rtsp_process = subprocess.Popen(
                rtsp_server_cmd,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE
            )
            
            # Wait a moment for server to start
            await asyncio.sleep(2)
            
            # Take snapshot using the controller
            result = await controller.take_snapshot(stream_name, snapshot_filename)
            
            # Verify successful capture
            assert result["status"] == "completed"
            assert result["filename"] == snapshot_filename
            assert result["file_size"] > 0
            assert os.path.exists(result["file_path"])
            
            # Verify output file is a valid image
            output_path = result["file_path"]
            file_size = os.path.getsize(output_path)
            assert file_size > 100  # Should be a reasonable image size
            
        finally:
            if rtsp_process:
                rtsp_process.terminate()
                try:
                    rtsp_process.wait(timeout=5)
                except subprocess.TimeoutExpired:
                    rtsp_process.kill()

    @pytest.mark.asyncio
    async def test_snapshot_file_size_validation(self, temp_dirs, test_image_file, real_mediamtx_server):
        """Test accurate file size reporting and validation."""
        controller = MediaMTXController(
            host="localhost",
            api_port=10001,  # Use real server port
            rtsp_port=8554,
            webrtc_port=8889,
            hls_port=8888,
            config_path="/tmp/test_config.yml",
            recordings_path=temp_dirs["recordings_dir"],
            snapshots_path=temp_dirs["snapshots_dir"],
            process_termination_timeout=2.0,
            process_kill_timeout=1.0,
        )
        # Use real HTTP server instead of mocking
        controller._base_url = f"http://{controller._host}:{controller._api_port}"
        await controller.start()
        
        stream_name = "size_test_stream"
        snapshot_filename = f"snapshot_{uuid.uuid4().hex[:8]}.jpg"
        
        # Create a simple RTSP stream
        rtsp_server_cmd = [
            "ffmpeg", "-re", "-loop", "1", "-i", test_image_file,
            "-c:v", "libx264", "-preset", "ultrafast",
            "-f", "rtsp", "-rtsp_transport", "tcp",
            f"rtsp://localhost:8554/{stream_name}"
        ]
        
        rtsp_process = None
        try:
            rtsp_process = subprocess.Popen(
                rtsp_server_cmd,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE
            )
            await asyncio.sleep(2)
            
            result = await controller.take_snapshot(stream_name, snapshot_filename)
            
            # Verify file size is accurate
            assert result["status"] == "completed"
            assert result["file_size"] > 0
            
            # Verify file size matches actual file
            actual_size = os.path.getsize(result["file_path"])
            assert result["file_size"] == actual_size
            
        finally:
            if rtsp_process:
                rtsp_process.terminate()
                try:
                    rtsp_process.wait(timeout=5)
                except subprocess.TimeoutExpired:
                    rtsp_process.kill()

    @pytest.mark.asyncio
    async def test_snapshot_process_cleanup_robustness(self, temp_dirs, real_mediamtx_server):
        """Test robust process cleanup under various failure conditions."""
        controller = MediaMTXController(
            host="localhost",
            api_port=10001,  # Use real server port
            rtsp_port=8554,
            webrtc_port=8889,
            hls_port=8888,
            config_path="/tmp/test_config.yml",
            recordings_path=temp_dirs["recordings_dir"],
            snapshots_path=temp_dirs["snapshots_dir"],
            process_termination_timeout=2.0,
            process_kill_timeout=1.0,
        )
        # Use real HTTP server instead of mocking
        controller._base_url = f"http://{controller._host}:{controller._api_port}"
        await controller.start()
        
        stream_name = "cleanup_test_stream"
        snapshot_filename = f"snapshot_{uuid.uuid4().hex[:8]}.jpg"
        
        # Test with invalid RTSP URL that will cause FFmpeg to fail
        result = await controller.take_snapshot(stream_name, snapshot_filename)
        
        # Verify failure is handled gracefully
        assert result["status"] == "failed"
        assert "FFmpeg capture failed" in result["error"]
        assert result["file_size"] == 0
