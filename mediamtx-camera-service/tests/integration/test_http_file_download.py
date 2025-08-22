"""
Integration tests for HTTP file download endpoints and secure file access.

Requirements Coverage:
- REQ-API-035: HTTP file download endpoints for secure file access

Story Coverage: E6 - File Management Infrastructure
IV&V Control Point: File access security validation

Tests HTTP file download endpoints for recordings and snapshots,
authentication requirements, and secure file access controls.
"""

import pytest
import pytest_asyncio
import asyncio
import tempfile
import os
import shutil
import aiohttp
import aiofiles
from pathlib import Path
from typing import Dict, Any

from tests.utils.port_utils import find_free_port
from tests.fixtures.auth_utils import get_test_auth_manager, TestUserFactory
from src.camera_service.config import Config, ServerConfig, MediaMTXConfig, CameraConfig, LoggingConfig, RecordingConfig, SnapshotConfig
from src.camera_service.service_manager import ServiceManager
from src.websocket_server.server import WebSocketJsonRpcServer
from mediamtx_wrapper.controller import MediaMTXController
from camera_discovery.hybrid_monitor import HybridCameraMonitor


class HTTPFileDownloadTestSetup:
    """Test setup for HTTP file download testing."""
    
    def __init__(self):
        self.config = self._build_test_config()
        self.service_manager = None
        self.mediamtx_controller = None
        self.camera_monitor = None
        self.server = None
        self.auth_manager = get_test_auth_manager()
        self.user_factory = TestUserFactory(self.auth_manager)
        self.temp_dir = None
        self.recordings_dir = None
        self.snapshots_dir = None
        self.http_session = None
    
    def _build_test_config(self) -> Config:
        """Build test configuration for HTTP file download testing."""
        # Use free ports to avoid conflicts
        free_websocket_port = find_free_port()
        free_health_port = find_free_port()
        
        # Create temporary directories
        self.temp_dir = tempfile.mkdtemp(prefix="http_download_test_")
        self.recordings_dir = os.path.join(self.temp_dir, "recordings")
        self.snapshots_dir = os.path.join(self.temp_dir, "snapshots")
        os.makedirs(self.recordings_dir, exist_ok=True)
        os.makedirs(self.snapshots_dir, exist_ok=True)
        
        return Config(
            server=ServerConfig(host="127.0.0.1", port=free_websocket_port, websocket_path="/ws", max_connections=10),
            mediamtx=MediaMTXConfig(
                host="127.0.0.1",
                api_port=9997,
                rtsp_port=8554,
                webrtc_port=8889,
                hls_port=8888,
                recordings_path=self.recordings_dir,
                snapshots_path=self.snapshots_dir,
            ),
            camera=CameraConfig(
                device_range=[0, 1, 2, 3], 
                enable_capability_detection=True, 
                detection_timeout=0.5,
                auto_start_streams=True
            ),
            logging=LoggingConfig(),
            recording=RecordingConfig(),
            snapshots=SnapshotConfig(),
            health_port=free_health_port,
        )
    
    async def setup(self):
        """Set up test environment for HTTP file download testing."""
        # Initialize real MediaMTX controller
        mediamtx_config = self.config.mediamtx
        self.mediamtx_controller = MediaMTXController(
            host=mediamtx_config.host,
            api_port=mediamtx_config.api_port,
            rtsp_port=mediamtx_config.rtsp_port,
            webrtc_port=mediamtx_config.webrtc_port,
            hls_port=mediamtx_config.hls_port,
            config_path=mediamtx_config.config_path,
            recordings_path=mediamtx_config.recordings_path,
            snapshots_path=mediamtx_config.snapshots_path,
            health_check_interval=mediamtx_config.health_check_interval,
            health_failure_threshold=mediamtx_config.health_failure_threshold,
        )
        
        # Initialize camera monitor
        camera_config = self.config.camera
        self.camera_monitor = HybridCameraMonitor(
            device_range=camera_config.device_range,
            poll_interval=camera_config.poll_interval,
            detection_timeout=camera_config.detection_timeout,
            enable_capability_detection=camera_config.enable_capability_detection,
        )
        
        # Initialize WebSocket server
        server_config = self.config.server
        self.server = WebSocketJsonRpcServer(
            host=server_config.host,
            port=server_config.port,
            websocket_path=server_config.websocket_path,
            max_connections=server_config.max_connections,
            mediamtx_controller=self.mediamtx_controller,
            camera_monitor=self.camera_monitor,
        )
        
        # Initialize service manager
        self.service_manager = ServiceManager(self.config)
        
        # Start all components
        await self.service_manager.start()
        
        # Create HTTP session for testing
        self.http_session = aiohttp.ClientSession()
        
        # Create test files
        self.create_test_files()
    
    async def cleanup(self):
        """Clean up test environment."""
        if self.http_session:
            await self.http_session.close()
        
        if self.service_manager:
            await self.service_manager.stop()
        
        # Clean up temporary directories
        if self.temp_dir and os.path.exists(self.temp_dir):
            shutil.rmtree(self.temp_dir)
    
    def create_test_files(self):
        """Create test files for download testing."""
        # Create test recordings
        recording_files = [
            "test_recording_1.mp4",
            "test_recording_2.mp4",
            "test_recording_3.mp4"
        ]
        
        for filename in recording_files:
            filepath = os.path.join(self.recordings_dir, filename)
            with open(filepath, 'w') as f:
                f.write(f"Test recording content for {filename}")
        
        # Create test snapshots
        snapshot_files = [
            "test_snapshot_1.jpg",
            "test_snapshot_2.jpg",
            "test_snapshot_3.jpg"
        ]
        
        for filename in snapshot_files:
            filepath = os.path.join(self.snapshots_dir, filename)
            with open(filepath, 'w') as f:
                f.write(f"Test snapshot content for {filename}")
    
    def get_auth_headers(self, user_role: str = "admin") -> Dict[str, str]:
        """Get authentication headers for HTTP requests."""
        if user_role == "admin":
            user = self.user_factory.create_admin_user(f"http_test_user_{user_role}")
        elif user_role == "operator":
            user = self.user_factory.create_operator_user(f"http_test_user_{user_role}")
        elif user_role == "viewer":
            user = self.user_factory.create_viewer_user(f"http_test_user_{user_role}")
        else:
            user = self.user_factory.create_operator_user(f"http_test_user_{user_role}")
        token = self.auth_manager.jwt_handler.generate_token(user.user_id, user.role)
        return {"Authorization": f"Bearer {token}"}
    
    async def download_file(self, file_type: str, filename: str, auth_headers: Dict[str, str] = None) -> aiohttp.ClientResponse:
        """Download a file via HTTP endpoint."""
        base_url = f"http://127.0.0.1:{self.config.health_port}"
        url = f"{base_url}/files/{file_type}/{filename}"
        
        headers = auth_headers or {}
        async with self.http_session.get(url, headers=headers) as response:
            return response


@pytest.mark.asyncio
@pytest.mark.integration
class TestHTTPFileDownload:
    """Integration tests for HTTP file download endpoints."""
    
    @pytest_asyncio.fixture
    async def download_setup(self):
        """Set up HTTP file download test environment."""
        setup = HTTPFileDownloadTestSetup()
        await setup.setup()
        yield setup
        await setup.cleanup()
    
    async def test_download_recording_success(self, download_setup):
        """Test successful download of recording files.
        
        REQ-API-035: HTTP file download endpoints for secure file access
        """
        auth_headers = download_setup.get_auth_headers("admin")
        response = await download_setup.download_file("recordings", "test_recording_1.mp4", auth_headers)
        
        assert response.status == 200
        assert response.headers.get("Content-Type") == "video/mp4"
        assert response.headers.get("Content-Disposition") == 'attachment; filename="test_recording_1.mp4"'
        
        content = await response.read()
        assert b"Test recording content for test_recording_1.mp4" in content
    
    async def test_download_snapshot_success(self, download_setup):
        """Test successful download of snapshot files.
        
        REQ-API-035: HTTP file download endpoints for secure file access
        """
        auth_headers = download_setup.get_auth_headers("admin")
        response = await download_setup.download_file("snapshots", "test_snapshot_1.jpg", auth_headers)
        
        assert response.status == 200
        assert response.headers.get("Content-Type") == "image/jpeg"
        assert response.headers.get("Content-Disposition") == 'attachment; filename="test_snapshot_1.jpg"'
        
        content = await response.read()
        assert b"Test snapshot content for test_snapshot_1.jpg" in content
    
    async def test_download_file_not_found(self, download_setup):
        """Test download of non-existent files.
        
        REQ-API-035: HTTP file download endpoints for secure file access
        """
        auth_headers = download_setup.get_auth_headers("admin")
        response = await download_setup.download_file("recordings", "nonexistent_file.mp4", auth_headers)
        
        assert response.status == 404
    
    async def test_download_file_unauthorized(self, download_setup):
        """Test download without authentication.
        
        REQ-API-035: HTTP file download endpoints for secure file access
        """
        response = await download_setup.download_file("recordings", "test_recording_1.mp4")
        
        assert response.status == 401
    
    async def test_download_file_invalid_token(self, download_setup):
        """Test download with invalid authentication token.
        
        REQ-API-035: HTTP file download endpoints for secure file access
        """
        invalid_headers = {"Authorization": "Bearer invalid_token"}
        response = await download_setup.download_file("recordings", "test_recording_1.mp4", invalid_headers)
        
        assert response.status == 401
    
    async def test_download_file_viewer_permissions(self, download_setup):
        """Test download with viewer permissions.
        
        REQ-API-035: HTTP file download endpoints for secure file access
        """
        auth_headers = download_setup.get_auth_headers("viewer")
        response = await download_setup.download_file("recordings", "test_recording_1.mp4", auth_headers)
        
        # Viewers should be able to download files
        assert response.status == 200
    
    async def test_download_file_operator_permissions(self, download_setup):
        """Test download with operator permissions.
        
        REQ-API-035: HTTP file download endpoints for secure file access
        """
        auth_headers = download_setup.get_auth_headers("operator")
        response = await download_setup.download_file("snapshots", "test_snapshot_1.jpg", auth_headers)
        
        # Operators should be able to download files
        assert response.status == 200
    
    async def test_download_file_invalid_file_type(self, download_setup):
        """Test download with invalid file type.
        
        REQ-API-035: HTTP file download endpoints for secure file access
        """
        auth_headers = download_setup.get_auth_headers("admin")
        base_url = f"http://127.0.0.1:{download_setup.config.health_port}"
        url = f"{base_url}/files/invalid_type/test_file.txt"
        
        async with download_setup.http_session.get(url, headers=auth_headers) as response:
            assert response.status == 404
    
    async def test_download_file_path_traversal_protection(self, download_setup):
        """Test protection against path traversal attacks.
        
        REQ-API-035: HTTP file download endpoints for secure file access
        """
        auth_headers = download_setup.get_auth_headers("admin")
        
        # Test various path traversal attempts
        malicious_filenames = [
            "../../../etc/passwd",
            "..\\..\\..\\windows\\system32\\config\\sam",
            "....//....//....//etc/passwd",
            "%2e%2e%2f%2e%2e%2f%2e%2e%2fetc%2fpasswd"
        ]
        
        for filename in malicious_filenames:
            response = await download_setup.download_file("recordings", filename, auth_headers)
            assert response.status == 404, f"Path traversal attack succeeded with {filename}"
    
    async def test_download_file_content_type_detection(self, download_setup):
        """Test proper content type detection for different file types.
        
        REQ-API-035: HTTP file download endpoints for secure file access
        """
        auth_headers = download_setup.get_auth_headers("admin")
        
        # Test recordings
        response = await download_setup.download_file("recordings", "test_recording_1.mp4", auth_headers)
        assert response.status == 200
        assert response.headers.get("Content-Type") == "video/mp4"
        
        # Test snapshots
        response = await download_setup.download_file("snapshots", "test_snapshot_1.jpg", auth_headers)
        assert response.status == 200
        assert response.headers.get("Content-Type") == "image/jpeg"
    
    async def test_download_file_content_disposition(self, download_setup):
        """Test proper content disposition headers.
        
        REQ-API-035: HTTP file download endpoints for secure file access
        """
        auth_headers = download_setup.get_auth_headers("admin")
        response = await download_setup.download_file("recordings", "test_recording_1.mp4", auth_headers)
        
        assert response.status == 200
        content_disposition = response.headers.get("Content-Disposition")
        assert content_disposition is not None
        assert 'attachment' in content_disposition
        assert 'filename="test_recording_1.mp4"' in content_disposition
    
    async def test_download_file_size_headers(self, download_setup):
        """Test proper content length headers.
        
        REQ-API-035: HTTP file download endpoints for secure file access
        """
        auth_headers = download_setup.get_auth_headers("admin")
        response = await download_setup.download_file("recordings", "test_recording_1.mp4", auth_headers)
        
        assert response.status == 200
        content_length = response.headers.get("Content-Length")
        assert content_length is not None
        assert int(content_length) > 0
    
    async def test_download_file_concurrent_requests(self, download_setup):
        """Test concurrent file download requests.
        
        REQ-API-035: HTTP file download endpoints for secure file access
        """
        auth_headers = download_setup.get_auth_headers("admin")
        
        # Make concurrent requests
        tasks = []
        for i in range(3):
            task = download_setup.download_file("recordings", f"test_recording_{i+1}.mp4", auth_headers)
            tasks.append(task)
        
        responses = await asyncio.gather(*tasks)
        
        # All requests should succeed
        for response in responses:
            assert response.status == 200
    
    async def test_download_file_large_file_handling(self, download_setup):
        """Test handling of large files.
        
        REQ-API-035: HTTP file download endpoints for secure file access
        """
        # Create a large test file
        large_file_path = os.path.join(download_setup.recordings_dir, "large_test_file.mp4")
        with open(large_file_path, 'wb') as f:
            # Create a 1MB file
            f.write(b"0" * 1024 * 1024)
        
        auth_headers = download_setup.get_auth_headers("admin")
        response = await download_setup.download_file("recordings", "large_test_file.mp4", auth_headers)
        
        assert response.status == 200
        content_length = response.headers.get("Content-Length")
        assert content_length == "1048576"  # 1MB in bytes
