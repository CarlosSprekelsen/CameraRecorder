"""
Integration tests for file metadata tracking and retrieval functionality.

Requirements Coverage:
- REQ-API-036: Comprehensive file metadata tracking and retrieval

Story Coverage: E6 - File Management Infrastructure
IV&V Control Point: File metadata management validation

Tests file metadata extraction, storage, retrieval, and comprehensive
metadata tracking for recordings and snapshots.

API Documentation Reference: docs/api/json-rpc-methods.md
"""

import pytest
import pytest_asyncio
import asyncio
import tempfile
import os
import shutil
import json
from datetime import datetime, timedelta
from pathlib import Path
from typing import Dict, Any

from tests.utils.port_utils import find_free_port
from tests.fixtures.auth_utils import get_test_auth_manager, TestUserFactory, WebSocketAuthTestClient
from src.camera_service.config import Config, ServerConfig, MediaMTXConfig, CameraConfig, LoggingConfig, RecordingConfig, SnapshotConfig
from src.camera_service.service_manager import ServiceManager
from src.websocket_server.server import WebSocketJsonRpcServer
from mediamtx_wrapper.controller import MediaMTXController
from camera_discovery.hybrid_monitor import HybridCameraMonitor


class FileMetadataTestSetup:
    """Test setup for file metadata tracking testing."""
    
    def __init__(self):
        self.service_manager = None
        self.mediamtx_controller = None
        self.camera_monitor = None
        self.server = None
        self.auth_manager = get_test_auth_manager()
        self.user_factory = TestUserFactory(self.auth_manager)
        self.websocket_client = None
        self.temp_dir = None
        self.recordings_dir = None
        self.snapshots_dir = None
        self.config = self._build_test_config()
    
    def _build_test_config(self) -> Config:
        """Build test configuration for file metadata testing."""
        # Use free ports to avoid conflicts
        free_websocket_port = find_free_port()
        free_health_port = find_free_port()
        
        # Create temporary directories
        self.temp_dir = tempfile.mkdtemp(prefix="metadata_test_")
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
        """Set up test environment for file metadata testing."""
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
        
        # Initialize service manager (this starts the WebSocket server with proper initialization)
        self.service_manager = ServiceManager(
            config=self.config,
            mediamtx_controller=self.mediamtx_controller,
            camera_monitor=self.camera_monitor
        )
        
        # Start service manager (this starts the WebSocket server with proper initialization)
        await self.service_manager.start()
        
        # Use the service manager's properly initialized WebSocket server
        self.server = self.service_manager._websocket_server
        
        # Create test user and client
        test_user = self.user_factory.create_admin_user("metadata_test_user")
        server_config = self.config.server
        self.websocket_client = WebSocketAuthTestClient(
            f"ws://{server_config.host}:{server_config.port}{server_config.websocket_path}",
            test_user
        )
        await self.websocket_client.connect()
        
        # Create test files with metadata
        self.create_test_files_with_metadata()
    
    async def cleanup(self):
        """Clean up test environment."""
        if self.websocket_client:
            await self.websocket_client.disconnect()
        
        # Don't stop the server - it's managed by the service manager
        # if self.server:
        #     await self.server.stop()
        
        if self.service_manager:
            await self.service_manager.stop()
        
        if self.camera_monitor:
            await self.camera_monitor.stop()
        
        if self.mediamtx_controller:
            await self.mediamtx_controller.stop()
        
        # Clean up temporary directories
        if self.temp_dir and os.path.exists(self.temp_dir):
            shutil.rmtree(self.temp_dir)
    
    def create_test_files_with_metadata(self):
        """Create test files with comprehensive metadata."""
        # Create test recordings with metadata
        recording_files = [
            {
                "filename": "test_recording_1.mp4",
                "size": 1024 * 1024,  # 1MB
                "duration": 60,  # 60 seconds
                "resolution": "1920x1080",
                "fps": 30,
                "bitrate": 5000,  # 5 Mbps
                "codec": "H.264"
            },
            {
                "filename": "test_recording_2.mp4",
                "size": 2048 * 1024,  # 2MB
                "duration": 120,  # 2 minutes
                "resolution": "1280x720",
                "fps": 25,
                "bitrate": 3000,  # 3 Mbps
                "codec": "H.264"
            }
        ]
        
        for file_info in recording_files:
            filepath = os.path.join(self.recordings_dir, file_info["filename"])
            with open(filepath, 'wb') as f:
                # Create file with specified size
                f.write(b"0" * file_info["size"])
            
            # Create metadata file
            metadata_file = filepath + ".json"
            metadata = {
                "filename": file_info["filename"],
                "size_bytes": file_info["size"],
                "duration_seconds": file_info["duration"],
                "resolution": file_info["resolution"],
                "fps": file_info["fps"],
                "bitrate_kbps": file_info["bitrate"],
                "codec": file_info["codec"],
                "created_at": datetime.now().isoformat(),
                "modified_at": datetime.now().isoformat(),
                "camera_device": "/dev/video0",
                "stream_url": "rtsp://localhost:8554/camera0"
            }
            
            with open(metadata_file, 'w') as f:
                json.dump(metadata, f, indent=2)
        
        # Create test snapshots with metadata
        snapshot_files = [
            {
                "filename": "test_snapshot_1.jpg",
                "size": 512 * 1024,  # 512KB
                "resolution": "1920x1080",
                "quality": 85,
                "format": "JPEG"
            },
            {
                "filename": "test_snapshot_2.jpg",
                "size": 256 * 1024,  # 256KB
                "resolution": "1280x720",
                "quality": 90,
                "format": "JPEG"
            }
        ]
        
        for file_info in snapshot_files:
            filepath = os.path.join(self.snapshots_dir, file_info["filename"])
            with open(filepath, 'wb') as f:
                # Create file with specified size
                f.write(b"0" * file_info["size"])
            
            # Create metadata file
            metadata_file = filepath + ".json"
            metadata = {
                "filename": file_info["filename"],
                "size_bytes": file_info["size"],
                "resolution": file_info["resolution"],
                "quality": file_info["quality"],
                "format": file_info["format"],
                "created_at": datetime.now().isoformat(),
                "modified_at": datetime.now().isoformat(),
                "camera_device": "/dev/video0",
                "exposure_time": "1/60",
                "iso": 100,
                "focal_length": "35mm"
            }
            
            with open(metadata_file, 'w') as f:
                json.dump(metadata, f, indent=2)


@pytest.mark.asyncio
@pytest.mark.integration
class TestFileMetadataTracking:
    """Integration tests for file metadata tracking and retrieval."""
    
    @pytest_asyncio.fixture
    async def metadata_setup(self):
        """Set up file metadata test environment."""
        setup = FileMetadataTestSetup()
        await setup.setup()
        yield setup
        await setup.cleanup()
    
    async def test_get_recording_info_success(self, metadata_setup):
        """Test successful retrieval of recording metadata.
        
        REQ-API-036: Comprehensive file metadata tracking and retrieval
        """
        response = await metadata_setup.websocket_client.send_request(
            "get_recording_info",
            {
                "filename": "test_recording_1.mp4"
            }
        )
        
        assert response.get("result") is not None
        recording_info = response["result"]
        
        # Verify basic metadata according to API documentation
        assert recording_info["filename"] == "test_recording_1.mp4"
        assert "file_size" in recording_info
        assert "duration" in recording_info
        assert "created_time" in recording_info
        assert "download_url" in recording_info
        
        # Verify data types
        assert isinstance(recording_info["file_size"], int)
        assert isinstance(recording_info["duration"], (int, float))
        assert recording_info["file_size"] > 0
        assert recording_info["duration"] > 0
        
        # Verify download URL format
        assert recording_info["download_url"].startswith("/files/recordings/")
    
    async def test_get_snapshot_info_success(self, metadata_setup):
        """Test successful retrieval of snapshot metadata.
        
        REQ-API-036: Comprehensive file metadata tracking and retrieval
        """
        response = await metadata_setup.websocket_client.send_request(
            "get_snapshot_info",
            {
                "filename": "test_snapshot_1.jpg"
            }
        )
        
        assert response.get("result") is not None
        snapshot_info = response["result"]
        
        # Verify basic metadata according to API documentation
        assert snapshot_info["filename"] == "test_snapshot_1.jpg"
        assert "file_size" in snapshot_info
        assert "resolution" in snapshot_info
        assert "created_time" in snapshot_info
        assert "download_url" in snapshot_info
        
        # Verify data types
        assert isinstance(snapshot_info["file_size"], int)
        assert isinstance(snapshot_info["resolution"], str)
        assert snapshot_info["file_size"] > 0
        
        # Verify download URL format
        assert snapshot_info["download_url"].startswith("/files/snapshots/")
    
    async def test_get_recording_info_file_not_found(self, metadata_setup):
        """Test retrieval of metadata for non-existent recording.
        
        REQ-API-036: Comprehensive file metadata tracking and retrieval
        """
        response = await metadata_setup.websocket_client.send_request(
            "get_recording_info",
            {
                "filename": "nonexistent_recording.mp4"
            }
        )
        
        assert response.get("error") is not None
        assert response["error"]["code"] == -32603  # Internal error or specific error code
    
    async def test_get_snapshot_info_file_not_found(self, metadata_setup):
        """Test retrieval of metadata for non-existent snapshot.
        
        REQ-API-036: Comprehensive file metadata tracking and retrieval
        """
        response = await metadata_setup.websocket_client.send_request(
            "get_snapshot_info",
            {
                "filename": "nonexistent_snapshot.jpg"
            }
        )
        
        assert response.get("error") is not None
        assert response["error"]["code"] == -32603  # Internal error or specific error code
    
    async def test_get_recording_info_invalid_parameters(self, metadata_setup):
        """Test retrieval of recording metadata with invalid parameters.
        
        REQ-API-036: Comprehensive file metadata tracking and retrieval
        """
        # Test missing filename
        response = await metadata_setup.websocket_client.send_request(
            "get_recording_info",
            {}
        )
        
        assert response.get("error") is not None
        assert response["error"]["code"] == -32602  # Invalid params
        
        # Test empty filename
        response = await metadata_setup.websocket_client.send_request(
            "get_recording_info",
            {
                "filename": ""
            }
        )
        
        assert response.get("error") is not None
        assert response["error"]["code"] == -32602  # Invalid params
    
    async def test_get_snapshot_info_invalid_parameters(self, metadata_setup):
        """Test retrieval of snapshot metadata with invalid parameters.
        
        REQ-API-036: Comprehensive file metadata tracking and retrieval
        """
        # Test missing filename
        response = await metadata_setup.websocket_client.send_request(
            "get_snapshot_info",
            {}
        )
        
        assert response.get("error") is not None
        assert response["error"]["code"] == -32602  # Invalid params
    
    async def test_get_recording_info_authentication_required(self, metadata_setup):
        """Test that recording metadata retrieval requires authentication.
        
        REQ-API-036: Comprehensive file metadata tracking and retrieval
        """
        # Create unauthenticated client by using send_unauthenticated_request
        response = await metadata_setup.websocket_client.send_unauthenticated_request(
            "get_recording_info",
            {
                "filename": "test_recording_1.mp4"
            }
        )
        
        assert response.get("error") is not None
        assert response["error"]["code"] == -32001  # Authentication error
        
        await unauthenticated_client.disconnect()
    
    async def test_get_snapshot_info_authentication_required(self, metadata_setup):
        """Test that snapshot metadata retrieval requires authentication.
        
        REQ-API-036: Comprehensive file metadata tracking and retrieval
        """
        # Create unauthenticated client by using send_unauthenticated_request
        response = await metadata_setup.websocket_client.send_unauthenticated_request(
            "get_snapshot_info",
            {
                "filename": "test_snapshot_1.jpg"
            }
        )
        
        assert response.get("error") is not None
        assert response["error"]["code"] == -32001  # Authentication error
    
    async def test_get_recording_info_viewer_permissions(self, metadata_setup):
        """Test recording metadata retrieval with viewer permissions.
        
        REQ-API-036: Comprehensive file metadata tracking and retrieval
        """
        # Create viewer user
        viewer_user = metadata_setup.user_factory.create_viewer_user("viewer_user")
        viewer_client = WebSocketAuthTestClient(
            metadata_setup.websocket_client.websocket_url,
            viewer_user
        )
        await viewer_client.connect()
        
        # Try to get metadata with viewer permissions
        response = await viewer_client.send_request(
            "get_recording_info",
            {
                "filename": "test_recording_1.mp4"
            }
        )
        
        # Viewers should be able to access metadata
        assert response.get("result") is not None
        assert response["result"]["metadata"]["filename"] == "test_recording_1.mp4"
        
        await viewer_client.disconnect()
    
    async def test_get_snapshot_info_operator_permissions(self, metadata_setup):
        """Test snapshot metadata retrieval with operator permissions.
        
        REQ-API-036: Comprehensive file metadata tracking and retrieval
        """
        # Create operator user
        operator_user = metadata_setup.user_factory.create_operator_user("operator_user")
        operator_client = WebSocketAuthTestClient(
            metadata_setup.websocket_client.websocket_url,
            operator_user
        )
        await operator_client.connect()
        
        # Try to get metadata with operator permissions
        response = await operator_client.send_request(
            "get_snapshot_info",
            {
                "filename": "test_snapshot_1.jpg"
            }
        )
        
        # Operators should be able to access metadata
        assert response.get("result") is not None
        assert response["result"]["metadata"]["filename"] == "test_snapshot_1.jpg"
        
        await operator_client.disconnect()
    
    async def test_metadata_consistency_across_files(self, metadata_setup):
        """Test metadata consistency across multiple files.
        
        REQ-API-036: Comprehensive file metadata tracking and retrieval
        """
        # Get metadata for both recordings
        response1 = await metadata_setup.websocket_client.send_request(
            "get_recording_info",
            {
                "filename": "test_recording_1.mp4"
            }
        )
        
        response2 = await metadata_setup.websocket_client.send_request(
            "get_recording_info",
            {
                "filename": "test_recording_2.mp4"
            }
        )
        
        assert response1.get("result") is not None
        assert response2.get("result") is not None
        
        metadata1 = response1["result"]["metadata"]
        metadata2 = response2["result"]["metadata"]
        
        # Verify different files have different metadata
        assert metadata1["filename"] != metadata2["filename"]
        assert metadata1["size_bytes"] != metadata2["size_bytes"]
        assert metadata1["duration_seconds"] != metadata2["duration_seconds"]
        
        # Verify consistent structure
        required_fields = ["filename", "size_bytes", "created_at", "modified_at", "camera_device"]
        for field in required_fields:
            assert field in metadata1
            assert field in metadata2
    
    async def test_metadata_timestamp_format(self, metadata_setup):
        """Test that metadata timestamps are in proper format.
        
        REQ-API-036: Comprehensive file metadata tracking and retrieval
        """
        response = await metadata_setup.websocket_client.send_request(
            "get_recording_info",
            {
                "filename": "test_recording_1.mp4"
            }
        )
        
        assert response.get("result") is not None
        metadata = response["result"]["metadata"]
        
        # Verify timestamp format (ISO 8601)
        from datetime import datetime
        try:
            datetime.fromisoformat(metadata["created_at"])
            datetime.fromisoformat(metadata["modified_at"])
        except ValueError:
            pytest.fail("Timestamps are not in ISO 8601 format")
    
    async def test_metadata_file_size_accuracy(self, metadata_setup):
        """Test that metadata file sizes are accurate.
        
        REQ-API-036: Comprehensive file metadata tracking and retrieval
        """
        response = await metadata_setup.websocket_client.send_request(
            "get_recording_info",
            {
                "filename": "test_recording_1.mp4"
            }
        )
        
        assert response.get("result") is not None
        metadata = response["result"]["metadata"]
        
        # Verify file size matches actual file
        actual_file_path = os.path.join(metadata_setup.recordings_dir, "test_recording_1.mp4")
        actual_size = os.path.getsize(actual_file_path)
        assert metadata["size_bytes"] == actual_size
