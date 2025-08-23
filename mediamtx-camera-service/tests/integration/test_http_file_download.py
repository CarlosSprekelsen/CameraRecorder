#!/usr/bin/env python3
"""
Integration tests for HTTP file download endpoints and secure file access.

Requirements Coverage:
- REQ-API-035: HTTP file download endpoints for secure file access

Test Categories: Integration
API Documentation Reference: docs/api/json-rpc-methods.md
"""

import pytest
import pytest_asyncio
import asyncio
import tempfile
import os
import shutil
import aiohttp
import aiofiles
import sys
from pathlib import Path
from typing import Dict, Any

# Add src to path for imports
sys.path.append('src')

from tests.utils.port_utils import find_free_port
from tests.fixtures.auth_utils import get_test_auth_manager, TestUserFactory, WebSocketAuthTestClient
from src.camera_service.config import Config, ServerConfig, MediaMTXConfig, CameraConfig, LoggingConfig, RecordingConfig, SnapshotConfig
from src.camera_service.service_manager import ServiceManager


def build_test_config() -> Config:
    """Build test configuration for HTTP file download testing."""
    # Use free ports to avoid conflicts
    free_websocket_port = find_free_port()
    free_health_port = find_free_port()
    
    return Config(
        server=ServerConfig(host="127.0.0.1", port=free_websocket_port, websocket_path="/ws", max_connections=10),
        mediamtx=MediaMTXConfig(
            host="127.0.0.1",
            api_port=9997,
            rtsp_port=8554,
            webrtc_port=8889,
            hls_port=8888,
            recordings_path="./.tmp_recordings",
            snapshots_path="./.tmp_snapshots",
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


class HTTPFileDownloadTestSetup:
    """Test setup for HTTP file download testing."""
    
    def __init__(self):
        self.config = build_test_config()
        self.service_manager = None
        self.server = None
        self.auth_manager = get_test_auth_manager()
        self.user_factory = TestUserFactory(self.auth_manager)
        self.websocket_client = None
        self.recordings_dir = self.config.mediamtx.recordings_path
        self.snapshots_dir = self.config.mediamtx.snapshots_path
        self.http_session = None
    
    async def setup(self):
        """Set up test environment for HTTP file download testing."""
        # Initialize service manager (this handles all component initialization)
        self.service_manager = ServiceManager(config=self.config)
        
        # Start service manager (this starts the WebSocket server with proper initialization)
        await self.service_manager.start()
        
        # Use the service manager's properly initialized WebSocket server
        self.server = self.service_manager._websocket_server
        
        # Create WebSocket client for testing
        websocket_url = f"ws://{self.config.server.host}:{self.config.server.port}{self.config.server.websocket_path}"
        # Create a test user for the WebSocket client (use operator like working tests)
        test_user = self.user_factory.create_operator_user("http_download_test_user")
        self.websocket_client = WebSocketAuthTestClient(websocket_url, test_user)
        await self.websocket_client.connect()
        
        # Create HTTP session for file downloads
        self.http_session = aiohttp.ClientSession()
        
        # Ensure directories exist
        os.makedirs(self.recordings_dir, exist_ok=True)
        os.makedirs(self.snapshots_dir, exist_ok=True)
    
    async def cleanup(self):
        """Clean up test environment."""
        if self.http_session:
            await self.http_session.close()
        
        if self.websocket_client:
            await self.websocket_client.disconnect()
        
        if self.service_manager:
            await self.service_manager.stop()
        
        # Clean up temporary directories if they exist
        if os.path.exists(self.recordings_dir):
            shutil.rmtree(self.recordings_dir, ignore_errors=True)
        if os.path.exists(self.snapshots_dir):
            shutil.rmtree(self.snapshots_dir, ignore_errors=True)
    
    def create_test_files(self):
        """Create test files for download testing."""
        # Create test recording
        recording_file = os.path.join(self.recordings_dir, "test_recording.mp4")
        with open(recording_file, 'w') as f:
            f.write("Test recording content")
        
        # Create test snapshot
        snapshot_file = os.path.join(self.snapshots_dir, "test_snapshot.jpg")
        with open(snapshot_file, 'w') as f:
            f.write("Test snapshot content")
        
        return recording_file, snapshot_file


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
        """Test successful download of recording file.
        
        REQ-API-035: HTTP file download endpoints for secure file access
        """
        # Create test recording file
        recording_file, _ = download_setup.create_test_files()
        
        # Get download URL via WebSocket API
        response = await download_setup.websocket_client.send_request(
            "get_download_url",
            {
                "file_type": "recording",
                "filename": "test_recording.mp4"
            }
        )
        
        assert response.get("result") is not None
        assert "download_url" in response["result"]
        
        # Download file via HTTP
        download_url = response["result"]["download_url"]
        async with download_setup.http_session.get(download_url) as resp:
            assert resp.status == 200
            content = await resp.read()
            assert content == b"Test recording content"
    
    async def test_download_snapshot_success(self, download_setup):
        """Test successful download of snapshot file.
        
        REQ-API-035: HTTP file download endpoints for secure file access
        """
        # Create test snapshot file
        _, snapshot_file = download_setup.create_test_files()
        
        # Get download URL via WebSocket API
        response = await download_setup.websocket_client.send_request(
            "get_download_url",
            {
                "file_type": "snapshot",
                "filename": "test_snapshot.jpg"
            }
        )
        
        assert response.get("result") is not None
        assert "download_url" in response["result"]
        
        # Download file via HTTP
        download_url = response["result"]["download_url"]
        async with download_setup.http_session.get(download_url) as resp:
            assert resp.status == 200
            content = await resp.read()
            assert content == b"Test snapshot content"
    
    async def test_download_authentication_required(self, download_setup):
        """Test that download URLs require authentication.
        
        REQ-API-035: HTTP file download endpoints for secure file access
        """
        # Create unauthenticated client
        websocket_url = f"ws://{download_setup.config.server.host}:{download_setup.config.server.port}{download_setup.config.server.websocket_path}"
        unauthenticated_client = WebSocketAuthTestClient(websocket_url, None)
        await unauthenticated_client.connect()
        
        try:
            # Try to get download URL without authentication
            response = await unauthenticated_client.send_unauthenticated_request(
                "get_download_url",
                {
                    "file_type": "recording",
                    "filename": "test_recording.mp4"
                }
            )
            
            # Should fail with authentication error
            assert response.get("error") is not None
            assert response["error"]["code"] == -32001  # Authentication required
        finally:
            await unauthenticated_client.disconnect()
    
    async def test_download_file_not_found(self, download_setup):
        """Test download URL for non-existent file.
        
        REQ-API-035: HTTP file download endpoints for secure file access
        """
        # Try to get download URL for non-existent file
        response = await download_setup.websocket_client.send_request(
            "get_download_url",
            {
                "file_type": "recording",
                "filename": "non_existent_file.mp4"
            }
        )
        
        # Should fail with file not found error
        assert response.get("error") is not None
        assert response["error"]["code"] == -32603  # Internal error or file not found
    
    async def test_download_invalid_file_type(self, download_setup):
        """Test download URL with invalid file type.
        
        REQ-API-035: HTTP file download endpoints for secure file access
        """
        # Try to get download URL with invalid file type
        response = await download_setup.websocket_client.send_request(
            "get_download_url",
            {
                "file_type": "invalid_type",
                "filename": "test_file.txt"
            }
        )
        
        # Should fail with invalid parameters error
        assert response.get("error") is not None
        assert response["error"]["code"] == -32602  # Invalid params
    
    async def test_download_url_expiration(self, download_setup):
        """Test that download URLs expire after use.
        
        REQ-API-035: HTTP file download endpoints for secure file access
        """
        # Create test file
        download_setup.create_test_files()
        
        # Get download URL
        response = await download_setup.websocket_client.send_request(
            "get_download_url",
            {
                "file_type": "recording",
                "filename": "test_recording.mp4"
            }
        )
        
        assert response.get("result") is not None
        download_url = response["result"]["download_url"]
        
        # Download file once (should work)
        async with download_setup.http_session.get(download_url) as resp:
            assert resp.status == 200
        
        # Try to download again (should fail - URL expired)
        async with download_setup.http_session.get(download_url) as resp:
            assert resp.status == 404  # URL should be expired/invalid
