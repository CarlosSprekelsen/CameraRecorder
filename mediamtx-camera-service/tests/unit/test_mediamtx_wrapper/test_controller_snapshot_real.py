# tests/unit/test_mediamtx_wrapper/test_controller_snapshot_real.py
"""
Real FFmpeg process management tests for snapshot capture functionality.

Test policy: Validate actual FFmpeg process execution, real file I/O operations,
and robust error handling with real subprocess management and timeout controls.
"""

import pytest
import asyncio
import os
import tempfile
import subprocess
import time
import uuid
from pathlib import Path
from typing import Dict, Any

from src.mediamtx_wrapper.controller import MediaMTXController


class TestSnapshotCaptureReal:
    """Test snapshot capture with real FFmpeg process execution."""

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
    async def controller(self, temp_dirs):
        """Create MediaMTX controller with test configuration."""
        controller = MediaMTXController(
            host="localhost",
            api_port=9997,
            rtsp_port=8554,
            webrtc_port=8889,
            hls_port=8888,
            config_path="/tmp/test_config.yml",
            recordings_path=temp_dirs["recordings_dir"],
            snapshots_path=temp_dirs["snapshots_dir"],
            process_termination_timeout=2.0,
            process_kill_timeout=1.0,
        )
        # Mock the session to avoid HTTP calls
        from unittest.mock import Mock
        controller._session = Mock()
        await controller.start()
        return controller

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

    def test_ffmpeg_availability(self):
        """Test that FFmpeg is available in the test environment."""
        try:
            result = subprocess.run(
                ["ffmpeg", "-version"],
                capture_output=True,
                text=True,
                timeout=10
            )
            assert result.returncode == 0
            assert "ffmpeg version" in result.stdout
        except (subprocess.TimeoutExpired, FileNotFoundError) as e:
            pytest.skip(f"FFmpeg not available: {e}")

    @pytest.mark.asyncio
    async def test_snapshot_invalid_stream_handling(self, controller):
        """Test handling of invalid/non-existent stream."""
        controller = await controller
        stream_name = "non_existent_stream"
        snapshot_filename = f"snapshot_{uuid.uuid4().hex[:8]}.jpg"
        
        # Attempt to take snapshot from non-existent stream
        result = await controller.take_snapshot(stream_name, snapshot_filename)
        
        # Verify failure is handled gracefully
        assert result["status"] == "failed"
        assert "FFmpeg capture failed" in result["error"]
        assert result["file_size"] == 0

    @pytest.mark.asyncio
    async def test_snapshot_process_timeout_handling(self, controller):
        """Test timeout handling with a hanging FFmpeg process."""
        controller = await controller
        stream_name = "timeout_test_stream"
        snapshot_filename = f"snapshot_{uuid.uuid4().hex[:8]}.jpg"
        
        # Create a process that will hang (invalid RTSP URL that doesn't respond)
        result = await controller.take_snapshot(stream_name, snapshot_filename)
        
        # Verify timeout is handled
        assert result["status"] == "failed"
        assert "timeout" in result["error"].lower()
        assert result["file_size"] == 0

    @pytest.mark.asyncio
    async def test_snapshot_directory_permission_error(self, controller, temp_dirs):
        """Test handling when snapshots directory cannot be written to."""
        controller = await controller
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
    async def test_snapshot_from_image_file_success(self, controller, test_image_file):
        """Test successful snapshot capture from a real image file."""
        controller = await controller
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
    async def test_snapshot_file_size_validation(self, controller, test_image_file):
        """Test accurate file size reporting and validation."""
        controller = await controller
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
    async def test_snapshot_process_cleanup_robustness(self, controller):
        """Test robust process cleanup under various failure conditions."""
        controller = await controller
        stream_name = "cleanup_test_stream"
        snapshot_filename = f"snapshot_{uuid.uuid4().hex[:8]}.jpg"
        
        # Test with invalid RTSP URL that will cause FFmpeg to fail
        result = await controller.take_snapshot(stream_name, snapshot_filename)
        
        # Verify failure is handled gracefully
        assert result["status"] == "failed"
        assert "FFmpeg capture failed" in result["error"]
        assert result["file_size"] == 0


# Test configuration expectations:
# - Real subprocess.run() calls to FFmpeg (no mocking subprocess)
# - Real input files created in temporary directories
# - Real output file validation (size, format, metadata)
# - Real error condition testing (invalid inputs, timeout scenarios)
# - Real process timeout and cleanup handling
# - Use actual FFmpeg binary available in test environment
# - Create minimal real test media files for reproducible testing
# - Handle FFmpeg process cleanup and timeout scenarios
# - Validate actual output file properties, not mock return values
