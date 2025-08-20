"""
WebSocket server file management tests.

Requirements Traceability:
- REQ-WS-001: WebSocket server shall handle file listing operations
- REQ-WS-002: WebSocket server shall provide file download URLs
- REQ-WS-003: WebSocket server shall handle file access permissions

Test Categories: Unit

COMPLIANT: Uses tempfile.TemporaryDirectory() for real filesystem testing
- Follows testing guide: "File Operations: Use tempfile.TemporaryDirectory()"
- No filesystem mocking - uses real filesystem operations
- Proper configuration setup for real system testing
"""

import pytest
import os
import tempfile
import shutil
from datetime import datetime
from pathlib import Path

from src.websocket_server.server import WebSocketJsonRpcServer, PermissionError
from src.camera_service.config import Config, ServerConfig, MediaMTXConfig, CameraConfig, RecordingConfig


class TestFileManagementMethods:
    """Test file management JSON-RPC methods using real filesystem."""

    @pytest.fixture
    def temp_directories(self):
        """Create temporary directories for testing using tempfile."""
        with tempfile.TemporaryDirectory() as temp_dir:
            recordings_dir = os.path.join(temp_dir, "recordings")
            snapshots_dir = os.path.join(temp_dir, "snapshots")
            
            os.makedirs(recordings_dir, exist_ok=True)
            os.makedirs(snapshots_dir, exist_ok=True)
            
            yield {
                "temp_dir": temp_dir,
                "recordings_dir": recordings_dir,
                "snapshots_dir": snapshots_dir
            }

    @pytest.fixture
    def sample_files(self, temp_directories):
        """Create sample files for testing using real filesystem."""
        recordings_dir = temp_directories["recordings_dir"]
        snapshots_dir = temp_directories["snapshots_dir"]
        
        # Create sample recording files
        recording_files = [
            "camera0_2025-01-15_14-30-00.mp4",
            "camera1_2025-01-15_15-45-30.mp4",
            "camera0_2025-01-16_09-15-45.mp4"
        ]
        
        for filename in recording_files:
            file_path = os.path.join(recordings_dir, filename)
            with open(file_path, 'w') as f:
                f.write(f"Sample recording content for {filename}")
        
        # Create sample snapshot files
        snapshot_files = [
            "camera0_2025-01-15_14-30-00.jpg",
            "camera1_2025-01-15_15-45-30.jpg",
            "camera0_2025-01-16_09-15-45.png"
        ]
        
        for filename in snapshot_files:
            file_path = os.path.join(snapshots_dir, filename)
            with open(file_path, 'w') as f:
                f.write(f"Sample snapshot content for {filename}")
        
        return {
            "recordings": recording_files,
            "snapshots": snapshot_files
        }

    @pytest.fixture
    def server_config(self, temp_directories):
        """Create proper server configuration with real paths."""
        # Create MediaMTX config with real temp directories
        mediamtx_config = MediaMTXConfig(
            host="127.0.0.1",
            api_port=9997,
            rtsp_port=8554,
            webrtc_port=8889,
            hls_port=8888,
            config_path=os.path.join(temp_directories["temp_dir"], "mediamtx.yml"),
            recordings_path=temp_directories["recordings_dir"],
            snapshots_path=temp_directories["snapshots_dir"]
        )
        
        # Create server config
        server_config = ServerConfig(
            host="127.0.0.1",
            port=8002,
            websocket_path="/ws"
        )
        
        # Create camera config
        camera_config = CameraConfig(device_range=[0, 1, 2])
        
        # Create recording config
        recording_config = RecordingConfig(enabled=True)
        
        # Create full config
        config = Config(
            server=server_config,
            mediamtx=mediamtx_config,
            camera=camera_config,
            recording=recording_config
        )
        
        return config

    @pytest.fixture
    def server(self, server_config):
        """Create WebSocket server instance with proper configuration."""
        server = WebSocketJsonRpcServer(
            host=server_config.server.host,
            port=server_config.server.port,
            websocket_path=server_config.server.websocket_path,
            max_connections=100
        )
        # Set the configuration properly
        server._config = server_config
        return server

    @pytest.mark.asyncio
    @pytest.mark.unit
    async def test_list_recordings_success(self, server, temp_directories, sample_files):
        """Test successful list_recordings method call using real filesystem."""
        # Test the method with real files
        result = await server._method_list_recordings()
        
        # Verify result structure
        assert "files" in result
        assert "total_count" in result
        assert "has_more" in result
        
        # Verify files are listed
        files = result["files"]
        assert len(files) == 3  # We created 3 recording files
        assert result["total_count"] == 3
        
        # Verify file structure
        for file_info in files:
            assert "filename" in file_info
            assert "size" in file_info
            assert "timestamp" in file_info
            assert "download_url" in file_info

    @pytest.mark.asyncio
    @pytest.mark.unit
    async def test_list_snapshots_success(self, server, temp_directories, sample_files):
        """Test successful list_snapshots method call using real filesystem."""
        # Test the method with real files
        result = await server._method_list_snapshots()
        
        # Verify result structure
        assert "files" in result
        assert "total_count" in result
        assert "has_more" in result
        
        # Verify files are listed
        files = result["files"]
        assert len(files) == 3  # We created 3 snapshot files
        assert result["total_count"] == 3
        
        # Verify file structure
        for file_info in files:
            assert "filename" in file_info
            assert "size" in file_info
            assert "timestamp" in file_info
            assert "download_url" in file_info

    @pytest.mark.asyncio
    @pytest.mark.unit
    async def test_list_recordings_directory_not_exists(self, server):
        """Test list_recordings when directory doesn't exist."""
        # Temporarily change recordings path to non-existent directory
        original_path = server._config.mediamtx.recordings_path
        server._config.mediamtx.recordings_path = "/non/existent/path"
        
        try:
            result = await server._method_list_recordings()
            
            # Should return empty result, not crash
            assert "files" in result
            assert len(result["files"]) == 0
            assert result["total_count"] == 0
        finally:
            # Restore original path
            server._config.mediamtx.recordings_path = original_path

    @pytest.mark.asyncio
    @pytest.mark.unit
    async def test_list_recordings_permission_denied(self, server, temp_directories):
        """Test list_recordings when permission is denied."""
        # Make directory read-only
        os.chmod(temp_directories["recordings_dir"], 0o000)
        
        try:
            # Should handle permission error gracefully
            with pytest.raises(PermissionError):
                await server._method_list_recordings()
        finally:
            # Restore permissions
            os.chmod(temp_directories["recordings_dir"], 0o755)

    @pytest.mark.asyncio
    @pytest.mark.unit
    async def test_list_recordings_pagination(self, server, temp_directories, sample_files):
        """Test list_recordings pagination functionality."""
        # Create more files for pagination testing
        recordings_dir = temp_directories["recordings_dir"]
        for i in range(10):
            file_path = os.path.join(recordings_dir, f"file_{i}.mp4")
            with open(file_path, 'w') as f:
                f.write(f"Content for file {i}")
        
        # Test with limit and offset
        params = {"limit": 3, "offset": 2}
        result = await server._method_list_recordings(params)
        
        # Verify pagination
        assert "files" in result
        assert len(result["files"]) == 3
        assert result["total_count"] == 13  # 3 original + 10 new files

    @pytest.mark.asyncio
    @pytest.mark.unit
    async def test_list_recordings_sorting(self, server, temp_directories):
        """Test that files are sorted by timestamp (newest first)."""
        recordings_dir = temp_directories["recordings_dir"]
        
        # Create files with different timestamps
        files = ["old.mp4", "new.mp4", "middle.mp4"]
        timestamps = [
            datetime(2025, 1, 15, 10, 0, 0).timestamp(),  # old
            datetime(2025, 1, 15, 16, 0, 0).timestamp(),  # new
            datetime(2025, 1, 15, 13, 0, 0).timestamp()   # middle
        ]
        
        for filename, timestamp in zip(files, timestamps):
            file_path = os.path.join(recordings_dir, filename)
            with open(file_path, 'w') as f:
                f.write(f"Content for {filename}")
            
            # Set file modification time
            os.utime(file_path, (timestamp, timestamp))
        
        result = await server._method_list_recordings()
        
        # Verify files are sorted by modification time (newest first)
        files_result = result["files"]
        assert len(files_result) == 3
        
        # Check that files are sorted by timestamp descending
        timestamps_result = [f["timestamp"] for f in files_result]
        assert timestamps_result == sorted(timestamps_result, reverse=True)

    @pytest.mark.asyncio
    @pytest.mark.unit
    async def test_list_recordings_video_duration_placeholder(self, server, temp_directories):
        """Test that video files have duration field (currently None placeholder)."""
        recordings_dir = temp_directories["recordings_dir"]
        
        # Create video and non-video files
        video_file = os.path.join(recordings_dir, "video.mp4")
        text_file = os.path.join(recordings_dir, "document.txt")
        
        with open(video_file, 'w') as f:
            f.write("Video content")
        with open(text_file, 'w') as f:
            f.write("Text content")
        
        result = await server._method_list_recordings()
        
        # Verify video files have duration field
        files_result = result["files"]
        video_files = [f for f in files_result if f["filename"] == "video.mp4"]
        text_files = [f for f in files_result if f["filename"] == "document.txt"]
        
        assert len(video_files) == 1
        assert "duration" in video_files[0]
        # Duration is currently None placeholder
        assert video_files[0]["duration"] is None
        
        assert len(text_files) == 1
        assert "duration" not in text_files[0]

    @pytest.mark.asyncio
    @pytest.mark.unit
    async def test_list_recordings_default_parameters(self, server, temp_directories, sample_files):
        """Test list_recordings with default parameters (None)."""
        result = await server._method_list_recordings()
        
        # Should return all files with default parameters
        assert "files" in result
        assert len(result["files"]) == 3  # All files
        assert result["total_count"] == 3

    @pytest.mark.asyncio
    @pytest.mark.unit
    async def test_list_snapshots_default_parameters(self, server, temp_directories, sample_files):
        """Test list_snapshots with default parameters (None)."""
        result = await server._method_list_snapshots()
        
        # Should return all files with default parameters
        assert "files" in result
        assert len(result["files"]) == 3  # All files
        assert result["total_count"] == 3
