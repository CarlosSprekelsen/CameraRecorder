#!/usr/bin/env python3
"""
Integration tests for storage space monitoring and alerts.

Requirements Coverage:
- REQ-API-041: Storage space monitoring and alerts
- REQ-API-042: Storage space threshold configuration
- REQ-API-043: Storage space alert notifications
- REQ-API-044: Storage space cleanup recommendations

Test Categories: Integration
"""

import pytest
import pytest_asyncio
import asyncio
import tempfile
import os
import shutil
import sys
from typing import Dict, Any

# Add src to path for imports
sys.path.append('src')

from tests.utils.port_utils import find_free_port
from tests.fixtures.auth_utils import get_test_auth_manager, TestUserFactory, WebSocketAuthTestClient
from src.camera_service.config import Config, ServerConfig, MediaMTXConfig, CameraConfig, LoggingConfig, RecordingConfig, SnapshotConfig
from src.camera_service.service_manager import ServiceManager


def build_test_config() -> Config:
    """Build test configuration for storage space monitoring testing."""
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


class StorageSpaceTestSetup:
    """Test setup for storage space monitoring testing."""
    
    def __init__(self):
        self.config = build_test_config()
        self.service_manager = None
        self.server = None
        self.auth_manager = get_test_auth_manager()
        self.user_factory = TestUserFactory(self.auth_manager)
        self.websocket_client = None
        self.recordings_dir = self.config.mediamtx.recordings_path
        self.snapshots_dir = self.config.mediamtx.snapshots_path
    
    async def setup(self):
        """Set up test environment for storage space monitoring testing."""
        # Initialize service manager (this handles all component initialization)
        self.service_manager = ServiceManager(config=self.config)
        
        # Start service manager (this starts the WebSocket server with proper initialization)
        await self.service_manager.start()
        
        # Use the service manager's properly initialized WebSocket server
        self.server = self.service_manager._websocket_server
        
        # Create WebSocket client for testing
        websocket_url = f"ws://{self.config.server.host}:{self.config.server.port}{self.config.server.websocket_path}"
        # Create a test user for the WebSocket client (use operator like working tests)
        test_user = self.user_factory.create_operator_user("storage_test_user")
        self.websocket_client = WebSocketAuthTestClient(websocket_url, test_user)
        await self.websocket_client.connect()
        
        # Ensure directories exist
        os.makedirs(self.recordings_dir, exist_ok=True)
        os.makedirs(self.snapshots_dir, exist_ok=True)
    
    async def cleanup(self):
        """Clean up test environment."""
        if self.websocket_client:
            await self.websocket_client.disconnect()
        
        if self.service_manager:
            await self.service_manager.stop()
        
        # Clean up temporary directories if they exist
        if os.path.exists(self.recordings_dir):
            shutil.rmtree(self.recordings_dir, ignore_errors=True)
        if os.path.exists(self.snapshots_dir):
            shutil.rmtree(self.snapshots_dir, ignore_errors=True)
    
    def create_test_files(self, size_mb: int = 1, count: int = 5):
        """Create test files with specific sizes for storage testing."""
        # Create test recordings
        for i in range(count):
            filename = f"test_recording_{i}.mp4"
            filepath = os.path.join(self.recordings_dir, filename)
            with open(filepath, 'wb') as f:
                f.write(b"0" * (size_mb * 1024 * 1024))  # Create file of specified size
        
        # Create test snapshots
        for i in range(count):
            filename = f"test_snapshot_{i}.jpg"
            filepath = os.path.join(self.snapshots_dir, filename)
            with open(filepath, 'wb') as f:
                f.write(b"0" * (size_mb * 1024 * 1024))  # Create file of specified size


@pytest.mark.asyncio
@pytest.mark.integration
class TestStorageSpaceMonitoring:
    """Integration tests for storage space monitoring and alerts."""
    
    @pytest_asyncio.fixture
    async def storage_setup(self):
        """Set up storage space monitoring test environment."""
        setup = StorageSpaceTestSetup()
        await setup.setup()
        yield setup
        await setup.cleanup()
    
    async def test_get_storage_info_success(self, storage_setup):
        """Test successful retrieval of storage information.
        
        REQ-API-041: Storage space monitoring and alerts
        """
        # Create some test files
        storage_setup.create_test_files(size_mb=1, count=3)
        
        # Get storage information
        response = await storage_setup.websocket_client.send_request(
            "get_storage_info",
            {}
        )
        
        assert response.get("result") is not None
        assert "total_space" in response["result"]
        assert "used_space" in response["result"]
        assert "free_space" in response["result"]
        assert "usage_percentage" in response["result"]
        
        # Validate storage information
        result = response["result"]
        assert result["total_space"] > 0
        assert result["used_space"] >= 0
        assert result["free_space"] >= 0
        assert 0 <= result["usage_percentage"] <= 100
        
        # Verify calculations
        assert result["used_space"] + result["free_space"] == result["total_space"]
        calculated_percentage = (result["used_space"] / result["total_space"]) * 100
        assert abs(result["usage_percentage"] - calculated_percentage) < 1  # Allow small rounding differences
    
    async def test_storage_threshold_configuration(self, storage_setup):
        """Test storage threshold configuration.
        
        REQ-API-042: Storage space threshold configuration
        """
        # Set storage threshold
        response = await storage_setup.websocket_client.send_request(
            "set_storage_threshold",
            {
                "warning_threshold": 80,
                "critical_threshold": 95
            }
        )
        
        assert response.get("result") is not None
        assert "warning_threshold" in response["result"]
        assert "critical_threshold" in response["result"]
        assert response["result"]["warning_threshold"] == 80
        assert response["result"]["critical_threshold"] == 95
        
        # Get storage threshold
        response = await storage_setup.websocket_client.send_request(
            "get_storage_threshold",
            {}
        )
        
        assert response.get("result") is not None
        assert "warning_threshold" in response["result"]
        assert "critical_threshold" in response["result"]
        assert response["result"]["warning_threshold"] == 80
        assert response["result"]["critical_threshold"] == 95
    
    async def test_storage_threshold_validation(self, storage_setup):
        """Test storage threshold validation.
        
        REQ-API-042: Storage space threshold configuration
        """
        # Test invalid thresholds
        response = await storage_setup.websocket_client.send_request(
            "set_storage_threshold",
            {
                "warning_threshold": 101,  # Invalid: > 100
                "critical_threshold": 95
            }
        )
        
        assert response.get("error") is not None
        assert response["error"]["code"] == -32602  # Invalid params
        
        # Test negative thresholds
        response = await storage_setup.websocket_client.send_request(
            "set_storage_threshold",
            {
                "warning_threshold": -1,  # Invalid: negative
                "critical_threshold": 95
            }
        )
        
        assert response.get("error") is not None
        assert response["error"]["code"] == -32602  # Invalid params
        
        # Test warning > critical
        response = await storage_setup.websocket_client.send_request(
            "set_storage_threshold",
            {
                "warning_threshold": 90,  # Invalid: warning > critical
                "critical_threshold": 80
            }
        )
        
        assert response.get("error") is not None
        assert response["error"]["code"] == -32602  # Invalid params
    
    async def test_storage_alert_notifications(self, storage_setup):
        """Test storage alert notifications.
        
        REQ-API-043: Storage space alert notifications
        """
        # Set low thresholds for testing
        response = await storage_setup.websocket_client.send_request(
            "set_storage_threshold",
            {
                "warning_threshold": 10,  # Very low for testing
                "critical_threshold": 20
            }
        )
        
        assert response.get("result") is not None
        
        # Create files to trigger alerts
        storage_setup.create_test_files(size_mb=10, count=10)  # Create large files
        
        # Check for storage alerts
        response = await storage_setup.websocket_client.send_request(
            "get_storage_alerts",
            {}
        )
        
        assert response.get("result") is not None
        assert "alerts" in response["result"]
        
        # Should have alerts due to high usage
        alerts = response["result"]["alerts"]
        assert len(alerts) > 0
        
        # Check alert types
        alert_types = [alert["type"] for alert in alerts]
        assert "warning" in alert_types or "critical" in alert_types
    
    async def test_storage_cleanup_recommendations(self, storage_setup):
        """Test storage cleanup recommendations.
        
        REQ-API-044: Storage space cleanup recommendations
        """
        # Create test files
        storage_setup.create_test_files(size_mb=1, count=5)
        
        # Get cleanup recommendations
        response = await storage_setup.websocket_client.send_request(
            "get_storage_cleanup_recommendations",
            {}
        )
        
        assert response.get("result") is not None
        assert "recommendations" in response["result"]
        assert "potential_space_saved" in response["result"]
        
        recommendations = response["result"]["recommendations"]
        assert isinstance(recommendations, list)
        
        # Should have recommendations for old files
        if len(recommendations) > 0:
            for rec in recommendations:
                assert "file_path" in rec
                assert "file_size" in rec
                assert "last_modified" in rec
                assert "recommendation_type" in rec
    
    async def test_storage_automatic_cleanup(self, storage_setup):
        """Test automatic storage cleanup.
        
        REQ-API-044: Storage space cleanup recommendations
        """
        # Create test files
        storage_setup.create_test_files(size_mb=1, count=3)
        
        # Get initial storage info
        initial_response = await storage_setup.websocket_client.send_request(
            "get_storage_info",
            {}
        )
        
        assert initial_response.get("result") is not None
        initial_used_space = initial_response["result"]["used_space"]
        
        # Trigger automatic cleanup
        response = await storage_setup.websocket_client.send_request(
            "trigger_storage_cleanup",
            {
                "cleanup_type": "automatic",
                "dry_run": False
            }
        )
        
        assert response.get("result") is not None
        assert "files_removed" in response["result"]
        assert "space_freed" in response["result"]
        
        # Get storage info after cleanup
        final_response = await storage_setup.websocket_client.send_request(
            "get_storage_info",
            {}
        )
        
        assert final_response.get("result") is not None
        final_used_space = final_response["result"]["used_space"]
        
        # Space should be freed (unless no files were old enough to clean)
        assert final_used_space <= initial_used_space
    
    async def test_storage_monitoring_authentication_required(self, storage_setup):
        """Test that storage monitoring requires authentication.
        
        REQ-API-041: Storage space monitoring and alerts
        """
        # Create unauthenticated client
        websocket_url = f"ws://{storage_setup.config.server.host}:{storage_setup.config.server.port}{storage_setup.config.server.websocket_path}"
        unauthenticated_client = WebSocketAuthTestClient(websocket_url, None)
        await unauthenticated_client.connect()
        
        try:
            # Try to get storage info without authentication
            response = await unauthenticated_client.send_unauthenticated_request(
                "get_storage_info",
                {}
            )
            
            # Should fail with authentication error
            assert response.get("error") is not None
            assert response["error"]["code"] == -32001  # Authentication required
        finally:
            await unauthenticated_client.disconnect()
    
    async def test_storage_monitoring_role_validation(self, storage_setup):
        """Test storage monitoring role validation.
        
        REQ-API-041: Storage space monitoring and alerts
        """
        # Create viewer user
        viewer_user = storage_setup.user_factory.create_viewer_user("storage_viewer_user")
        viewer_client = WebSocketAuthTestClient(
            f"ws://{storage_setup.config.server.host}:{storage_setup.config.server.port}{storage_setup.config.server.websocket_path}",
            viewer_user
        )
        await viewer_client.connect()
        
        try:
            # Viewer should be able to get storage info
            response = await viewer_client.send_request(
                "get_storage_info",
                {}
            )
            
            assert response.get("result") is not None
            assert "total_space" in response["result"]
            
            # Viewer should NOT be able to set thresholds
            response = await viewer_client.send_request(
                "set_storage_threshold",
                {
                    "warning_threshold": 80,
                    "critical_threshold": 95
                }
            )
            
            # Should fail with permission error
            assert response.get("error") is not None
            assert response["error"]["code"] == -32003  # Permission denied
        finally:
            await viewer_client.disconnect()
