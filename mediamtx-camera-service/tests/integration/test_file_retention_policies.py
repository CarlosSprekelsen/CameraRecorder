"""
Integration tests for file retention policies and cleanup functionality.

Requirements Coverage:
- REQ-API-034: Configurable file retention policies and cleanup

Story Coverage: E6 - File Management Infrastructure
IV&V Control Point: File lifecycle management validation

Tests file retention policy configuration, automatic cleanup based on age/size,
and manual cleanup operations for recordings and snapshots.
"""

import pytest
import pytest_asyncio
import asyncio
import tempfile
import os
import time
import shutil
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


class FileRetentionTestSetup:
    """Test setup for file retention policy testing."""
    
    def __init__(self):
        self.config = self._build_test_config()
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
    
    def _build_test_config(self) -> Config:
        """Build test configuration for file retention testing."""
        # Use free ports to avoid conflicts
        free_websocket_port = find_free_port()
        free_health_port = find_free_port()
        
        # Create temporary directories
        self.temp_dir = tempfile.mkdtemp(prefix="retention_test_")
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
        """Set up test environment for file retention testing."""
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
        
        # Create test user and client
        test_user = self.user_factory.create_admin_user("retention_test_user")
        self.websocket_client = WebSocketAuthTestClient(
            f"ws://{server_config.host}:{server_config.port}{server_config.websocket_path}",
            test_user
        )
        await self.websocket_client.connect()
    
    async def cleanup(self):
        """Clean up test environment."""
        if self.websocket_client:
            await self.websocket_client.disconnect()
        
        if self.service_manager:
            await self.service_manager.stop()
        
        # Clean up temporary directories
        if self.temp_dir and os.path.exists(self.temp_dir):
            shutil.rmtree(self.temp_dir)
    
    def create_test_files(self, age_hours: int = 24, count: int = 5):
        """Create test files with specific ages for retention testing."""
        base_time = datetime.now() - timedelta(hours=age_hours)
        
        # Create test recordings
        for i in range(count):
            filename = f"test_recording_{i}_{int(base_time.timestamp())}.mp4"
            filepath = os.path.join(self.recordings_dir, filename)
            with open(filepath, 'w') as f:
                f.write(f"Test recording content {i}")
            # Set file modification time
            os.utime(filepath, (base_time.timestamp(), base_time.timestamp()))
        
        # Create test snapshots
        for i in range(count):
            filename = f"test_snapshot_{i}_{int(base_time.timestamp())}.jpg"
            filepath = os.path.join(self.snapshots_dir, filename)
            with open(filepath, 'w') as f:
                f.write(f"Test snapshot content {i}")
            # Set file modification time
            os.utime(filepath, (base_time.timestamp(), base_time.timestamp()))


@pytest.mark.asyncio
@pytest.mark.integration
class TestFileRetentionPolicies:
    """Integration tests for file retention policies and cleanup."""
    
    @pytest_asyncio.fixture
    async def retention_setup(self):
        """Set up file retention test environment."""
        setup = FileRetentionTestSetup()
        await setup.setup()
        yield setup
        await setup.cleanup()
    
    async def test_set_retention_policy_success(self, retention_setup):
        """Test successful setting of file retention policies.
        
        REQ-API-034: Configurable file retention policies and cleanup
        """
        # Set retention policy for recordings
        response = await retention_setup.websocket_client.send_request(
            "set_retention_policy",
            {
                "policy_type": "age",
                "max_age_days": 1,
                "enabled": True
            }
        )
        
        assert response.get("result") is not None
        assert response["result"]["policy_type"] == "age"
        assert response["result"]["max_age_days"] == 1
        assert response["result"]["enabled"] is True
        assert "message" in response["result"]
        
        # Set retention policy for snapshots
        response = await retention_setup.websocket_client.send_request(
            "set_retention_policy",
            {
                "policy_type": "size",
                "max_size_gb": 1,
                "enabled": True
            }
        )
        
        assert response.get("result") is not None
        assert response["result"]["policy_type"] == "size"
        assert response["result"]["max_size_gb"] == 1
        assert response["result"]["enabled"] is True
        assert "message" in response["result"]
    
    async def test_set_retention_policy_invalid_parameters(self, retention_setup):
        """Test setting retention policy with invalid parameters.
        
        REQ-API-034: Configurable file retention policies and cleanup
        """
        # Test invalid policy type
        response = await retention_setup.websocket_client.send_request(
            "set_retention_policy",
            {
                "policy_type": "invalid_type",
                "max_age_days": 30,
                "enabled": True
            }
        )
        
        assert response.get("error") is not None
        assert response["error"]["code"] == -32602  # Invalid params
        
        # Test negative age
        response = await retention_setup.websocket_client.send_request(
            "set_retention_policy",
            {
                "policy_type": "age",
                "max_age_days": -1,
                "enabled": True
            }
        )
        
        assert response.get("error") is not None
        assert response["error"]["code"] == -32602  # Invalid params
    
    async def test_cleanup_old_files_success(self, retention_setup):
        """Test successful cleanup of old files based on retention policies.
        
        REQ-API-034: Configurable file retention policies and cleanup
        """
        # Create test files that are 48 hours old
        retention_setup.create_test_files(age_hours=48, count=3)
        
        # Set retention policy to 1 day
        await retention_setup.websocket_client.send_request(
            "set_retention_policy",
            {
                "policy_type": "age",
                "max_age_days": 1,
                "enabled": True
            }
        )
        
        # Verify files exist before cleanup
        recordings_before = len(os.listdir(retention_setup.recordings_dir))
        snapshots_before = len(os.listdir(retention_setup.snapshots_dir))
        assert recordings_before >= 3
        assert snapshots_before >= 3
        
        # Run cleanup
        response = await retention_setup.websocket_client.send_request(
            "cleanup_old_files",
            {}
        )
        
        assert response.get("result") is not None
        assert response["result"]["cleanup_executed"] is True
        assert response["result"]["files_deleted"] >= 3
        assert response["result"]["space_freed"] > 0
        assert "message" in response["result"]
        
        # Verify files were deleted
        recordings_after = len(os.listdir(retention_setup.recordings_dir))
        assert recordings_after < recordings_before
    

    
    async def test_cleanup_old_files_no_policy(self, retention_setup):
        """Test cleanup when no retention policy is set.
        
        REQ-API-034: Configurable file retention policies and cleanup
        """
        # Create test files
        retention_setup.create_test_files(age_hours=48, count=2)
        
        # Try cleanup without setting policy
        response = await retention_setup.websocket_client.send_request(
            "cleanup_old_files",
            {}
        )
        
        assert response.get("error") is not None
        assert response["error"]["code"] == -32603  # Internal error or specific error code
    
    async def test_cleanup_old_files_disabled_policy(self, retention_setup):
        """Test cleanup when retention policy is disabled.
        
        REQ-API-034: Configurable file retention policies and cleanup
        """
        # Create test files
        retention_setup.create_test_files(age_hours=48, count=2)
        
        # Set disabled retention policy
        await retention_setup.websocket_client.send_request(
            "set_retention_policy",
            {
                "policy_type": "age",
                "max_age_days": 1,
                "enabled": False
            }
        )
        
        # Try cleanup with disabled policy
        response = await retention_setup.websocket_client.send_request(
            "cleanup_old_files",
            {}
        )
        
        assert response.get("error") is not None
        assert response["error"]["code"] == -32603  # Internal error or specific error code
    
    async def test_cleanup_old_files_authentication_required(self, retention_setup):
        """Test that cleanup requires proper authentication.
        
        REQ-API-034: Configurable file retention policies and cleanup
        """
        # Create unauthenticated client
        from tests.fixtures.auth_utils import WebSocketAuthTestClient
        unauthenticated_client = WebSocketAuthTestClient(
            retention_setup.websocket_client.websocket_url,
            None  # No user = unauthenticated
        )
        await unauthenticated_client.connect()
        
        # Try cleanup without authentication
        response = await unauthenticated_client.send_request(
            "cleanup_old_files",
            {
                "file_type": "recordings",
                "dry_run": True
            }
        )
        
        assert response.get("error") is not None
        assert response["error"]["code"] == -32001  # Authentication error
        
        await unauthenticated_client.disconnect()
    
    async def test_cleanup_old_files_permission_required(self, retention_setup):
        """Test that cleanup requires admin permissions.
        
        REQ-API-034: Configurable file retention policies and cleanup
        """
        # Create viewer user (non-admin)
        viewer_user = retention_setup.user_factory.create_viewer_user("viewer_user")
        viewer_client = WebSocketAuthTestClient(
            retention_setup.websocket_client.websocket_url,
            viewer_user
        )
        await viewer_client.connect()
        
        # Try cleanup with viewer permissions
        response = await viewer_client.send_request(
            "cleanup_old_files",
            {
                "file_type": "recordings",
                "dry_run": True
            }
        )
        
        assert response.get("error") is not None
        assert response["error"]["code"] == -32003  # Permission error
        
        await viewer_client.disconnect()
