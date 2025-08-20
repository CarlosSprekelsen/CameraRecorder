"""
Unit tests for health server file download endpoints.

Tests the HTTP file download functionality added in Epic E6:
- /files/recordings/{filename} endpoint
- /files/snapshots/{filename} endpoint

Requirements: REQ-FUNC-010, REQ-FUNC-011
Epic E6: Server Recording and Snapshot File Management Infrastructure

TODO: VIOLATION - Filesystem mocking violates strategic mocking rules
- Lines 86-90: Mocking os.path.join, os.path.exists, os.path.isfile, os.access, os.path.getsize
- Lines 119-123: Mocking os.path.join, os.path.exists, os.path.isfile, os.access, os.path.getsize
- Lines 185-186: Mocking os.path.join, os.path.exists
- Lines 208-210: Mocking os.path.join, os.path.exists, os.path.isfile
- Lines 233-236: Mocking os.path.join, os.path.exists, os.path.isfile, os.access
- Lines 260-264: Mocking os.path.join, os.path.exists, os.path.isfile, os.access, os.path.getsize
- Lines 302-306: Mocking os.path.join, os.path.exists, os.path.isfile, os.access, os.path.getsize
- Lines 345-349: Mocking os.path.join, os.path.exists, os.path.isfile, os.access, os.path.getsize
- Lines 371-375: Mocking os.path.join, os.path.exists, os.path.isfile, os.access, os.path.getsize
- VIOLATION: Testing guide states "NEVER MOCK: filesystem"
- FIX REQUIRED: Replace with tempfile.TemporaryDirectory() for real filesystem testing
"""

import pytest
import os
import tempfile
import shutil
from unittest.mock import patch, MagicMock
from aiohttp import web, ClientSession
from aiohttp.test_utils import make_mocked_request

from src.health_server import HealthServer


class TestFileDownloadEndpoints:
    """Test HTTP file download endpoints."""

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
        
        # Create sample recording file
        recording_file = "camera0_2025-01-15_14-30-00.mp4"
        recording_path = os.path.join(recordings_dir, recording_file)
        with open(recording_path, 'w') as f:
            f.write("Sample recording content")
        
        # Create sample snapshot file
        snapshot_file = "camera0_2025-01-15_14-30-00.jpg"
        snapshot_path = os.path.join(snapshots_dir, snapshot_file)
        with open(snapshot_path, 'w') as f:
            f.write("Sample snapshot content")
        
        return {
            "recording_file": recording_file,
            "recording_path": recording_path,
            "snapshot_file": snapshot_file,
            "snapshot_path": snapshot_path
        }

    @pytest.fixture
    def health_server(self):
        """Create health server instance for testing."""
        return HealthServer(host="127.0.0.1", port=8003)

    @pytest.fixture
    async def app(self, health_server):
        """Create aiohttp application for testing."""
        app = web.Application()
        
        # Add routes
        app.router.add_get("/files/recordings/{filename:.*}", health_server._handle_recording_download)
        app.router.add_get("/files/snapshots/{filename:.*}", health_server._handle_snapshot_download)
        
        return app

    @pytest.mark.asyncio
    @patch('src.health_server.os.path.join')
    @patch('src.health_server.os.path.exists')
    @patch('src.health_server.os.path.isfile')
    @patch('src.health_server.os.access')
    @patch('src.health_server.os.path.getsize')
    async def test_recording_download_success(self, mock_getsize, mock_access, mock_isfile, 
                                            mock_exists, mock_join, health_server, sample_files):
        """Test successful recording file download."""
        # Setup mocks
        mock_join.return_value = sample_files["recording_path"]
        mock_exists.return_value = True
        mock_isfile.return_value = True
        mock_access.return_value = True
        mock_getsize.return_value = 1024
        
        # Create mock request
        request = make_mocked_request(
            'GET', 
            '/files/recordings/camera0_2025-01-15_14-30-00.mp4',
            match_info={'filename': 'camera0_2025-01-15_14-30-00.mp4'}
        )
        
        # Test the handler
        response = await health_server._handle_recording_download(request)
        
        # Verify response
        assert response.status == 200
        assert response.headers['Content-Type'] == 'video/mp4'
        assert response.headers['Content-Disposition'] == 'attachment; filename="camera0_2025-01-15_14-30-00.mp4"'
        assert response.headers['Content-Length'] == '1024'
        assert response.headers['Accept-Ranges'] == 'bytes'

    @pytest.mark.asyncio
    @patch('src.health_server.os.path.join')
    @patch('src.health_server.os.path.exists')
    @patch('src.health_server.os.path.isfile')
    @patch('src.health_server.os.access')
    @patch('src.health_server.os.path.getsize')
    async def test_snapshot_download_success(self, mock_getsize, mock_access, mock_isfile, 
                                           mock_exists, mock_join, health_server, sample_files):
        """Test successful snapshot file download."""
        # Setup mocks
        mock_join.return_value = sample_files["snapshot_path"]
        mock_exists.return_value = True
        mock_isfile.return_value = True
        mock_access.return_value = True
        mock_getsize.return_value = 512
        
        # Create mock request
        request = make_mocked_request(
            'GET', 
            '/files/snapshots/camera0_2025-01-15_14-30-00.jpg',
            match_info={'filename': 'camera0_2025-01-15_14-30-00.jpg'}
        )
        
        # Test the handler
        response = await health_server._handle_snapshot_download(request)
        
        # Verify response
        assert response.status == 200
        assert response.headers['Content-Type'] == 'image/jpeg'
        assert response.headers['Content-Disposition'] == 'attachment; filename="camera0_2025-01-15_14-30-00.jpg"'
        assert response.headers['Content-Length'] == '512'

    @pytest.mark.asyncio
    async def test_recording_download_directory_traversal_attempt(self, health_server):
        """Test recording download with directory traversal attempt."""
        # Create mock request with directory traversal
        request = make_mocked_request(
            'GET', 
            '/files/recordings/../../../etc/passwd',
            match_info={'filename': '../../../etc/passwd'}
        )
        
        # Test the handler
        response = await health_server._handle_recording_download(request)
        
        # Verify response
        assert response.status == 400
        assert "Invalid filename" in response.text

    @pytest.mark.asyncio
    async def test_snapshot_download_directory_traversal_attempt(self, health_server):
        """Test snapshot download with directory traversal attempt."""
        # Create mock request with directory traversal
        request = make_mocked_request(
            'GET', 
            '/files/snapshots/../../../etc/passwd',
            match_info={'filename': '../../../etc/passwd'}
        )
        
        # Test the handler
        response = await health_server._handle_snapshot_download(request)
        
        # Verify response
        assert response.status == 400
        assert "Invalid filename" in response.text

    @pytest.mark.asyncio
    @patch('src.health_server.os.path.join')
    @patch('src.health_server.os.path.exists')
    async def test_recording_download_file_not_found(self, mock_exists, mock_join, health_server):
        """Test recording download when file doesn't exist."""
        # Setup mocks
        mock_join.return_value = "/opt/camera-service/recordings/nonexistent.mp4"
        mock_exists.return_value = False
        
        # Create mock request
        request = make_mocked_request(
            'GET', 
            '/files/recordings/nonexistent.mp4',
            match_info={'filename': 'nonexistent.mp4'}
        )
        
        # Test the handler
        response = await health_server._handle_recording_download(request)
        
        # Verify response
        assert response.status == 404
        assert "File not found" in response.text

    @pytest.mark.asyncio
    @patch('src.health_server.os.path.join')
    @patch('src.health_server.os.path.exists')
    @patch('src.health_server.os.path.isfile')
    async def test_recording_download_not_a_file(self, mock_isfile, mock_exists, mock_join, health_server):
        """Test recording download when path is not a file."""
        # Setup mocks
        mock_join.return_value = "/opt/camera-service/recordings/directory"
        mock_exists.return_value = True
        mock_isfile.return_value = False
        
        # Create mock request
        request = make_mocked_request(
            'GET', 
            '/files/recordings/directory',
            match_info={'filename': 'directory'}
        )
        
        # Test the handler
        response = await health_server._handle_recording_download(request)
        
        # Verify response
        assert response.status == 404
        assert "File not found" in response.text

    @pytest.mark.asyncio
    @patch('src.health_server.os.path.join')
    @patch('src.health_server.os.path.exists')
    @patch('src.health_server.os.path.isfile')
    @patch('src.health_server.os.access')
    async def test_recording_download_permission_denied(self, mock_access, mock_isfile, mock_exists, mock_join, health_server):
        """Test recording download when permission is denied."""
        # Setup mocks
        mock_join.return_value = "/opt/camera-service/recordings/restricted.mp4"
        mock_exists.return_value = True
        mock_isfile.return_value = True
        mock_access.return_value = False
        
        # Create mock request
        request = make_mocked_request(
            'GET', 
            '/files/recordings/restricted.mp4',
            match_info={'filename': 'restricted.mp4'}
        )
        
        # Test the handler
        response = await health_server._handle_recording_download(request)
        
        # Verify response
        assert response.status == 403
        assert "Permission denied" in response.text

    @pytest.mark.asyncio
    @patch('src.health_server.os.path.join')
    @patch('src.health_server.os.path.exists')
    @patch('src.health_server.os.path.isfile')
    @patch('src.health_server.os.access')
    @patch('src.health_server.os.path.getsize')
    async def test_recording_download_different_video_formats(self, mock_getsize, mock_access, mock_isfile, 
                                                            mock_exists, mock_join, health_server):
        """Test recording download with different video formats."""
        # Test different video formats
        test_cases = [
            ("video.avi", "video/x-msvideo"),
            ("video.mov", "video/quicktime"),
            ("video.mkv", "video/x-matroska"),
            ("video.wmv", "video/x-ms-wmv"),
            ("video.flv", "video/x-flv"),
            ("video.webm", "video/webm"),
            ("video.unknown", "video/mp4")  # Default fallback
        ]
        
        for filename, expected_mime in test_cases:
            # Setup mocks
            mock_join.return_value = f"/opt/camera-service/recordings/{filename}"
            mock_exists.return_value = True
            mock_isfile.return_value = True
            mock_access.return_value = True
            mock_getsize.return_value = 1024
            
            # Create mock request
            request = make_mocked_request(
                'GET', 
                f'/files/recordings/{filename}',
                match_info={'filename': filename}
            )
            
            # Test the handler
            response = await health_server._handle_recording_download(request)
            
            # Verify response
            assert response.status == 200
            assert response.headers['Content-Type'] == expected_mime

    @pytest.mark.asyncio
    @patch('src.health_server.os.path.join')
    @patch('src.health_server.os.path.exists')
    @patch('src.health_server.os.path.isfile')
    @patch('src.health_server.os.access')
    @patch('src.health_server.os.path.getsize')
    async def test_snapshot_download_different_image_formats(self, mock_getsize, mock_access, mock_isfile, 
                                                           mock_exists, mock_join, health_server):
        """Test snapshot download with different image formats."""
        # Test different image formats
        test_cases = [
            ("image.jpg", "image/jpeg"),
            ("image.jpeg", "image/jpeg"),
            ("image.png", "image/png"),
            ("image.gif", "image/gif"),
            ("image.bmp", "image/bmp"),
            ("image.tiff", "image/tiff"),
            ("image.webp", "image/jpeg"),  # webp not supported by default mimetypes
            ("image.unknown", "image/jpeg")  # Default fallback
        ]
        
        for filename, expected_mime in test_cases:
            # Setup mocks
            mock_join.return_value = f"/opt/camera-service/snapshots/{filename}"
            mock_exists.return_value = True
            mock_isfile.return_value = True
            mock_access.return_value = True
            mock_getsize.return_value = 512
            
            # Create mock request
            request = make_mocked_request(
                'GET', 
                f'/files/snapshots/{filename}',
                match_info={'filename': filename}
            )
            
            # Test the handler
            response = await health_server._handle_snapshot_download(request)
            
            # Verify response
            assert response.status == 200
            assert response.headers['Content-Type'] == expected_mime

    @pytest.mark.asyncio
    @patch('src.health_server.os.path.join')
    @patch('src.health_server.os.path.exists')
    @patch('src.health_server.os.path.isfile')
    @patch('src.health_server.os.access')
    @patch('src.health_server.os.path.getsize')
    async def test_recording_download_exception_handling(self, mock_getsize, mock_access, mock_isfile, 
                                                       mock_exists, mock_join, health_server):
        """Test recording download exception handling."""
        # Setup mocks to raise exception
        mock_join.side_effect = Exception("Test exception")
        
        # Create mock request
        request = make_mocked_request(
            'GET', 
            '/files/recordings/test.mp4',
            match_info={'filename': 'test.mp4'}
        )
        
        # Test the handler
        response = await health_server._handle_recording_download(request)
        
        # Verify response
        assert response.status == 500
        assert "Internal server error" in response.text

    @pytest.mark.asyncio
    @patch('src.health_server.os.path.join')
    @patch('src.health_server.os.path.exists')
    @patch('src.health_server.os.path.isfile')
    @patch('src.health_server.os.access')
    @patch('src.health_server.os.path.getsize')
    async def test_snapshot_download_exception_handling(self, mock_getsize, mock_access, mock_isfile, 
                                                      mock_exists, mock_join, health_server):
        """Test snapshot download exception handling."""
        # Setup mocks to raise exception
        mock_join.side_effect = Exception("Test exception")
        
        # Create mock request
        request = make_mocked_request(
            'GET', 
            '/files/snapshots/test.jpg',
            match_info={'filename': 'test.jpg'}
        )
        
        # Test the handler
        response = await health_server._handle_snapshot_download(request)
        
        # Verify response
        assert response.status == 500
        assert "Internal server error" in response.text
