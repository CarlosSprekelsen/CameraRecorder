"""
WebSocket server file management tests.

Requirements Traceability:
- REQ-WS-001: WebSocket server shall handle file listing operations
- REQ-WS-002: WebSocket server shall provide file download URLs
- REQ-WS-003: WebSocket server shall handle file access permissions

Test Categories: Unit

TODO: VIOLATION - Filesystem mocking violates strategic mocking rules
- Lines 89-102: Mocking os.path.exists, os.access, os.listdir, os.path.isfile, os.stat
- VIOLATION: Testing guide states "NEVER MOCK: filesystem"
- FIX REQUIRED: Replace with tempfile.TemporaryDirectory() for real filesystem testing
"""

import pytest
import os
import tempfile
import shutil
from unittest.mock import patch, MagicMock
from datetime import datetime

from src.websocket_server.server import WebSocketJsonRpcServer, PermissionError


class TestFileManagementMethods:
    """Test file management JSON-RPC methods."""

    @pytest.fixture
    def temp_directories(self):
        """Create temporary directories for testing."""
        temp_dir = tempfile.mkdtemp()
        recordings_dir = os.path.join(temp_dir, "recordings")
        snapshots_dir = os.path.join(temp_dir, "snapshots")
        
        os.makedirs(recordings_dir, exist_ok=True)
        os.makedirs(snapshots_dir, exist_ok=True)
        
        yield {
            "temp_dir": temp_dir,
            "recordings_dir": recordings_dir,
            "snapshots_dir": snapshots_dir
        }
        
        # Cleanup
        shutil.rmtree(temp_dir)

    @pytest.fixture
    def sample_files(self, temp_directories):
        """Create sample files for testing."""
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
    def server(self):
        """Create WebSocket server instance for testing."""
        return WebSocketJsonRpcServer(host="127.0.0.1", port=8002, websocket_path="/ws", max_connections=100)

    @pytest.mark.asyncio
    @pytest.mark.unit
    @patch('src.websocket_server.server.os.path.exists')
    @patch('src.websocket_server.server.os.access')
    @patch('src.websocket_server.server.os.listdir')
    @patch('src.websocket_server.server.os.path.isfile')
    @patch('src.websocket_server.server.os.stat')
    async def test_list_recordings_success(self, mock_stat, mock_isfile, mock_listdir, 
                                         mock_access, mock_exists, server, temp_directories):
        """Test successful list_recordings method call."""
        # Setup mocks
        mock_exists.return_value = True
        mock_access.return_value = True
        mock_listdir.return_value = ["camera0_2025-01-15_14-30-00.mp4", "camera1_2025-01-15_15-45-30.mp4"]
        mock_isfile.return_value = True
        
        # Mock file stats
        mock_stat_info = MagicMock()
        mock_stat_info.st_size = 1048576  # 1MB
        mock_stat_info.st_mtime = datetime(2025, 1, 15, 14, 30, 0).timestamp()
        mock_stat.return_value = mock_stat_info
        
        # Test the method
        result = await server._method_list_recordings()
        
        # Verify result structure
        assert "files" in result
        assert "total_count" in result
        assert "has_more" in result
        assert result["total_count"] == 2
        assert len(result["files"]) == 2
        
        # Verify file information
        file_info = result["files"][0]
        assert "filename" in file_info
        assert "size" in file_info
        assert "timestamp" in file_info
        assert "download_url" in file_info
        assert "duration" in file_info  # Should be None for now
        
        # Verify download URL format
        assert file_info["download_url"].startswith("/files/recordings/")

    @pytest.mark.asyncio
    @patch('src.websocket_server.server.os.path.exists')
    @patch('src.websocket_server.server.os.access')
    @patch('src.websocket_server.server.os.listdir')
    @patch('src.websocket_server.server.os.path.isfile')
    @patch('src.websocket_server.server.os.stat')
    async def test_list_snapshots_success(self, mock_stat, mock_isfile, mock_listdir, 
                                        mock_access, mock_exists, server, temp_directories):
        """Test successful list_snapshots method call."""
        # Setup mocks
        mock_exists.return_value = True
        mock_access.return_value = True
        mock_listdir.return_value = ["camera0_2025-01-15_14-30-00.jpg", "camera1_2025-01-15_15-45-30.jpg"]
        mock_isfile.return_value = True
        
        # Mock file stats
        mock_stat_info = MagicMock()
        mock_stat_info.st_size = 524288  # 512KB
        mock_stat_info.st_mtime = datetime(2025, 1, 15, 14, 30, 0).timestamp()
        mock_stat.return_value = mock_stat_info
        
        # Test the method
        result = await server._method_list_snapshots()
        
        # Verify result structure
        assert "files" in result
        assert "total_count" in result
        assert "has_more" in result
        assert result["total_count"] == 2
        assert len(result["files"]) == 2
        
        # Verify file information
        file_info = result["files"][0]
        assert "filename" in file_info
        assert "size" in file_info
        assert "timestamp" in file_info
        assert "download_url" in file_info
        assert "duration" not in file_info  # Snapshots don't have duration
        
        # Verify download URL format
        assert file_info["download_url"].startswith("/files/snapshots/")

    @pytest.mark.asyncio
    @patch('src.websocket_server.server.os.path.exists')
    async def test_list_recordings_directory_not_exists(self, mock_exists, server):
        """Test list_recordings when directory doesn't exist."""
        mock_exists.return_value = False
        
        result = await server._method_list_recordings()
        
        assert result["files"] == []
        assert result["total_count"] == 0
        assert result["has_more"] is False

    @pytest.mark.asyncio
    @patch('src.websocket_server.server.os.path.exists')
    @patch('src.websocket_server.server.os.access')
    async def test_list_recordings_permission_denied(self, mock_access, mock_exists, server):
        """Test list_recordings when permission is denied."""
        mock_exists.return_value = True
        mock_access.return_value = False
        
        with pytest.raises(PermissionError):
            await server._method_list_recordings()

    @pytest.mark.asyncio
    async def test_list_recordings_invalid_limit_parameter(self, server):
        """Test list_recordings with invalid limit parameter."""
        params = {"limit": -1}
        
        with pytest.raises(ValueError, match="Invalid limit parameter"):
            await server._method_list_recordings(params)

    @pytest.mark.asyncio
    async def test_list_recordings_invalid_offset_parameter(self, server):
        """Test list_recordings with invalid offset parameter."""
        params = {"offset": -1}
        
        with pytest.raises(ValueError, match="Invalid offset parameter"):
            await server._method_list_recordings(params)

    @pytest.mark.asyncio
    @patch('src.websocket_server.server.os.path.exists')
    @patch('src.websocket_server.server.os.access')
    @patch('src.websocket_server.server.os.listdir')
    @patch('src.websocket_server.server.os.path.isfile')
    @patch('src.websocket_server.server.os.stat')
    async def test_list_recordings_pagination(self, mock_stat, mock_isfile, mock_listdir, 
                                            mock_access, mock_exists, server):
        """Test list_recordings pagination functionality."""
        # Setup mocks
        mock_exists.return_value = True
        mock_access.return_value = True
        mock_listdir.return_value = [f"file_{i}.mp4" for i in range(10)]
        mock_isfile.return_value = True
        
        # Mock file stats
        mock_stat_info = MagicMock()
        mock_stat_info.st_size = 1048576
        mock_stat_info.st_mtime = datetime(2025, 1, 15, 14, 30, 0).timestamp()
        mock_stat.return_value = mock_stat_info
        
        # Test with limit and offset
        params = {"limit": 3, "offset": 2}
        result = await server._method_list_recordings(params)
        
        assert result["total_count"] == 10
        assert len(result["files"]) == 3
        assert result["has_more"] is True

    @pytest.mark.asyncio
    @patch('src.websocket_server.server.os.path.exists')
    @patch('src.websocket_server.server.os.access')
    @patch('src.websocket_server.server.os.listdir')
    @patch('src.websocket_server.server.os.path.isfile')
    @patch('src.websocket_server.server.os.stat')
    async def test_list_recordings_sorting(self, mock_stat, mock_isfile, mock_listdir, 
                                         mock_access, mock_exists, server):
        """Test that files are sorted by timestamp (newest first)."""
        # Setup mocks
        mock_exists.return_value = True
        mock_access.return_value = True
        mock_listdir.return_value = ["old.mp4", "new.mp4", "middle.mp4"]
        mock_isfile.return_value = True
        
        # Mock file stats with different timestamps
        def mock_stat_side_effect(path):
            mock_stat_info = MagicMock()
            mock_stat_info.st_size = 1048576
            
            if "old" in path:
                mock_stat_info.st_mtime = datetime(2025, 1, 15, 10, 0, 0).timestamp()
            elif "new" in path:
                mock_stat_info.st_mtime = datetime(2025, 1, 15, 16, 0, 0).timestamp()
            else:  # middle
                mock_stat_info.st_mtime = datetime(2025, 1, 15, 13, 0, 0).timestamp()
            
            return mock_stat_info
        
        mock_stat.side_effect = mock_stat_side_effect
        
        result = await server._method_list_recordings()
        
        # Verify files are sorted by timestamp (newest first)
        filenames = [f["filename"] for f in result["files"]]
        assert filenames == ["new.mp4", "middle.mp4", "old.mp4"]

    @pytest.mark.asyncio
    @patch('src.websocket_server.server.os.path.exists')
    @patch('src.websocket_server.server.os.access')
    @patch('src.websocket_server.server.os.listdir')
    @patch('src.websocket_server.server.os.path.isfile')
    @patch('src.websocket_server.server.os.stat')
    async def test_list_recordings_video_duration_placeholder(self, mock_stat, mock_isfile, mock_listdir, 
                                                            mock_access, mock_exists, server):
        """Test that video files have duration field (currently None placeholder)."""
        # Setup mocks
        mock_exists.return_value = True
        mock_access.return_value = True
        mock_listdir.return_value = ["video.mp4", "document.txt"]
        mock_isfile.return_value = True
        
        # Mock file stats
        mock_stat_info = MagicMock()
        mock_stat_info.st_size = 1048576
        mock_stat_info.st_mtime = datetime(2025, 1, 15, 14, 30, 0).timestamp()
        mock_stat.return_value = mock_stat_info
        
        result = await server._method_list_recordings()
        
        # Find video file
        video_file = next(f for f in result["files"] if f["filename"] == "video.mp4")
        assert "duration" in video_file
        assert video_file["duration"] is None  # Placeholder for now
        
        # Find non-video file
        text_file = next(f for f in result["files"] if f["filename"] == "document.txt")
        assert "duration" not in text_file

    @pytest.mark.asyncio
    async def test_list_recordings_default_parameters(self, server):
        """Test list_recordings with default parameters (None)."""
        # This test would require more complex mocking, but we can test the parameter handling
        with patch('src.websocket_server.server.os.path.exists', return_value=False):
            result = await server._method_list_recordings()
            
            # Should use default values when params is None
            assert result["files"] == []
            assert result["total_count"] == 0
            assert result["has_more"] is False

    @pytest.mark.asyncio
    async def test_list_snapshots_default_parameters(self, server):
        """Test list_snapshots with default parameters (None)."""
        with patch('src.websocket_server.server.os.path.exists', return_value=False):
            result = await server._method_list_snapshots()
            
            # Should use default values when params is None
            assert result["files"] == []
            assert result["total_count"] == 0
            assert result["has_more"] is False
