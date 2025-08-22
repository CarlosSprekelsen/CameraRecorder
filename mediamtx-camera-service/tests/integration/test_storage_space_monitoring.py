"""
Integration tests for real-time storage space monitoring and alerts.

Requirements Coverage:
- REQ-API-037: Real-time storage space monitoring and alerts

Story Coverage: E6 - File Management Infrastructure
IV&V Control Point: Storage monitoring validation

Tests real-time storage space monitoring, threshold alerts, storage calculations,
and storage space management for recordings and snapshots.
"""

import pytest
import pytest_asyncio
import asyncio
import tempfile
import os
import shutil
import psutil
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


class StorageMonitoringTestSetup:
    """Test setup for storage space monitoring testing."""
    
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
        """Build test configuration for storage monitoring testing."""
        # Use free ports to avoid conflicts
        free_websocket_port = find_free_port()
        free_health_port = find_free_port()
        
        # Create temporary directories
        self.temp_dir = tempfile.mkdtemp(prefix="storage_monitoring_test_")
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
        """Set up test environment for storage monitoring testing."""
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
        test_user = self.user_factory.create_admin_user("storage_test_user")
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
    
    def create_test_files_with_sizes(self, recording_size_mb: int = 10, snapshot_size_mb: int = 5):
        """Create test files with specific sizes to simulate storage usage."""
        # Create test recordings
        recording_files = [
            f"test_recording_{i}.mp4" for i in range(3)
        ]
        
        for filename in recording_files:
            filepath = os.path.join(self.recordings_dir, filename)
            with open(filepath, 'wb') as f:
                # Create file with specified size
                f.write(b"0" * (recording_size_mb * 1024 * 1024))
        
        # Create test snapshots
        snapshot_files = [
            f"test_snapshot_{i}.jpg" for i in range(5)
        ]
        
        for filename in snapshot_files:
            filepath = os.path.join(self.snapshots_dir, filename)
            with open(filepath, 'wb') as f:
                # Create file with specified size
                f.write(b"0" * (snapshot_size_mb * 1024 * 1024))
    
    def get_disk_usage_info(self) -> Dict[str, Any]:
        """Get disk usage information for the test directory."""
        disk_usage = psutil.disk_usage(self.temp_dir)
        return {
            "total_bytes": disk_usage.total,
            "used_bytes": disk_usage.used,
            "free_bytes": disk_usage.free,
            "percent_used": disk_usage.percent
        }


@pytest.mark.asyncio
@pytest.mark.integration
class TestStorageSpaceMonitoring:
    """Integration tests for storage space monitoring and alerts."""
    
    @pytest_asyncio.fixture
    async def storage_setup(self):
        """Set up storage monitoring test environment."""
        setup = StorageMonitoringTestSetup()
        await setup.setup()
        yield setup
        await setup.cleanup()
    
    async def test_get_storage_info_success(self, storage_setup):
        """Test successful retrieval of storage information.
        
        REQ-API-037: Real-time storage space monitoring and alerts
        """
        response = await storage_setup.websocket_client.send_request(
            "get_storage_info",
            {}
        )
        
        assert response.get("result") is not None
        storage_info = response["result"]
        
        # Verify basic storage information according to API documentation
        assert "total_space" in storage_info
        assert "used_space" in storage_info
        assert "available_space" in storage_info
        assert "usage_percentage" in storage_info
        assert "recordings_size" in storage_info
        assert "snapshots_size" in storage_info
        assert "low_space_warning" in storage_info
        
        # Verify data types
        assert isinstance(storage_info["total_space"], int)
        assert isinstance(storage_info["used_space"], int)
        assert isinstance(storage_info["available_space"], int)
        assert isinstance(storage_info["usage_percentage"], (int, float))
        assert isinstance(storage_info["recordings_size"], int)
        assert isinstance(storage_info["snapshots_size"], int)
        assert isinstance(storage_info["low_space_warning"], bool)
        
        # Verify logical relationships
        assert storage_info["total_space"] > 0
        assert storage_info["used_space"] >= 0
        assert storage_info["available_space"] >= 0
        assert 0 <= storage_info["usage_percentage"] <= 100
        assert storage_info["used_space"] + storage_info["available_space"] == storage_info["total_space"]
    
    async def test_get_storage_info_with_files(self, storage_setup):
        """Test storage information with test files.
        
        REQ-API-037: Real-time storage space monitoring and alerts
        """
        # Get initial storage info
        initial_response = await storage_setup.websocket_client.send_request(
            "get_storage_info",
            {}
        )
        
        assert initial_response.get("result") is not None
        initial_info = initial_response["result"]
        initial_used = initial_info["used_space"]
        
        # Create test files
        storage_setup.create_test_files_with_sizes(recording_size_mb=5, snapshot_size_mb=2)
        
        # Get updated storage info
        updated_response = await storage_setup.websocket_client.send_request(
            "get_storage_info",
            {}
        )
        
        assert updated_response.get("result") is not None
        updated_info = updated_response["result"]
        updated_used = updated_info["used_space"]
        
        # Verify storage usage increased
        assert updated_used > initial_used
        
        # Calculate expected increase (3 recordings * 5MB + 5 snapshots * 2MB = 25MB)
        expected_increase = (3 * 5 + 5 * 2) * 1024 * 1024
        actual_increase = updated_used - initial_used
        
        # Allow for some variance due to filesystem overhead
        assert actual_increase >= expected_increase * 0.9  # At least 90% of expected
    
    async def test_storage_info_accuracy(self, storage_setup):
        """Test accuracy of storage information against system data.
        
        REQ-API-037: Real-time storage space monitoring and alerts
        """
        response = await storage_setup.websocket_client.send_request(
            "get_storage_info",
            {}
        )
        
        assert response.get("result") is not None
        api_storage_info = response["result"]
        
        # Get system storage info
        system_storage_info = storage_setup.get_disk_usage_info()
        
        # Verify API matches system (with small tolerance for timing differences)
        tolerance = 1024 * 1024  # 1MB tolerance
        
        assert abs(api_storage_info["total_space"] - system_storage_info["total_bytes"]) < tolerance
        assert abs(api_storage_info["used_space"] - system_storage_info["used_bytes"]) < tolerance
        assert abs(api_storage_info["available_space"] - system_storage_info["free_bytes"]) < tolerance
        assert abs(api_storage_info["usage_percentage"] - system_storage_info["percent_used"]) < 1  # 1% tolerance
    
    async def test_storage_alerts_thresholds(self, storage_setup):
        """Test storage alerts based on usage thresholds.
        
        REQ-API-037: Real-time storage space monitoring and alerts
        """
        # Set storage alert thresholds
        response = await storage_setup.websocket_client.send_request(
            "set_storage_alerts",
            {
                "warning_threshold_percent": 80,
                "critical_threshold_percent": 95,
                "enabled": True
            }
        )
        
        assert response.get("result") is not None
        assert response["result"]["success"] is True
        
        # Get storage info with alerts
        storage_response = await storage_setup.websocket_client.send_request(
            "get_storage_info",
            {}
        )
        
        assert storage_response.get("result") is not None
        storage_info = storage_response["result"]["storage_info"]
        
        # Verify alert information is included
        assert "alerts" in storage_info
        alerts = storage_info["alerts"]
        
        # Verify alert structure
        assert "warning_threshold_percent" in alerts
        assert "critical_threshold_percent" in alerts
        assert "current_status" in alerts
        assert "warnings" in alerts
        
        # Verify threshold values
        assert alerts["warning_threshold_percent"] == 80
        assert alerts["critical_threshold_percent"] == 95
    
    async def test_storage_info_authentication_required(self, storage_setup):
        """Test that storage info retrieval requires authentication.
        
        REQ-API-037: Real-time storage space monitoring and alerts
        """
        # Create unauthenticated client
        from tests.fixtures.auth_utils import WebSocketAuthTestClient
        unauthenticated_client = WebSocketAuthTestClient(
            storage_setup.websocket_client.websocket_url,
            None  # No user = unauthenticated
        )
        await unauthenticated_client.connect()
        
        # Try to get storage info without authentication
        response = await unauthenticated_client.send_request(
            "get_storage_info",
            {}
        )
        
        assert response.get("error") is not None
        assert response["error"]["code"] == -32001  # Authentication error
        
        await unauthenticated_client.disconnect()
    
    async def test_storage_info_viewer_permissions(self, storage_setup):
        """Test storage info retrieval with viewer permissions.
        
        REQ-API-037: Real-time storage space monitoring and alerts
        """
        # Create viewer user
        viewer_user = storage_setup.user_factory.create_viewer_user("viewer_user")
        viewer_client = WebSocketAuthTestClient(
            storage_setup.websocket_client.websocket_url,
            viewer_user
        )
        await viewer_client.connect()
        
        # Try to get storage info with viewer permissions
        response = await viewer_client.send_request(
            "get_storage_info",
            {}
        )
        
        # Viewers should be able to access storage info
        assert response.get("result") is not None
        assert "storage_info" in response["result"]
        
        await viewer_client.disconnect()
    
    async def test_storage_info_operator_permissions(self, storage_setup):
        """Test storage info retrieval with operator permissions.
        
        REQ-API-037: Real-time storage space monitoring and alerts
        """
        # Create operator user
        operator_user = storage_setup.user_factory.create_operator_user("operator_user")
        operator_client = WebSocketAuthTestClient(
            storage_setup.websocket_client.websocket_url,
            operator_user
        )
        await operator_client.connect()
        
        # Try to get storage info with operator permissions
        response = await operator_client.send_request(
            "get_storage_info",
            {}
        )
        
        # Operators should be able to access storage info
        assert response.get("result") is not None
        assert "storage_info" in response["result"]
        
        await operator_client.disconnect()
    
    async def test_storage_info_real_time_updates(self, storage_setup):
        """Test that storage info provides real-time updates.
        
        REQ-API-037: Real-time storage space monitoring and alerts
        """
        # Get initial storage info
        initial_response = await storage_setup.websocket_client.send_request(
            "get_storage_info",
            {}
        )
        
        assert initial_response.get("result") is not None
        initial_info = initial_response["result"]["storage_info"]
        
        # Create a large file to change storage usage
        large_file_path = os.path.join(storage_setup.recordings_dir, "large_test_file.mp4")
        with open(large_file_path, 'wb') as f:
            f.write(b"0" * (50 * 1024 * 1024))  # 50MB file
        
        # Get updated storage info
        updated_response = await storage_setup.websocket_client.send_request(
            "get_storage_info",
            {}
        )
        
        assert updated_response.get("result") is not None
        updated_info = updated_response["result"]["storage_info"]
        
        # Verify storage usage increased
        assert updated_info["used_bytes"] > initial_info["used_bytes"]
        assert updated_info["free_bytes"] < initial_info["free_bytes"]
        assert updated_info["percent_used"] > initial_info["percent_used"]
    
    async def test_storage_info_detailed_breakdown(self, storage_setup):
        """Test detailed storage breakdown by file type.
        
        REQ-API-037: Real-time storage space monitoring and alerts
        """
        # Create test files
        storage_setup.create_test_files_with_sizes(recording_size_mb=10, snapshot_size_mb=5)
        
        response = await storage_setup.websocket_client.send_request(
            "get_storage_info",
            {
                "include_breakdown": True
            }
        )
        
        assert response.get("result") is not None
        storage_info = response["result"]["storage_info"]
        
        # Verify detailed breakdown is included
        assert "breakdown" in storage_info
        breakdown = storage_info["breakdown"]
        
        # Verify breakdown structure
        assert "recordings" in breakdown
        assert "snapshots" in breakdown
        
        recordings_info = breakdown["recordings"]
        snapshots_info = breakdown["snapshots"]
        
        # Verify recordings breakdown
        assert "total_files" in recordings_info
        assert "total_size_bytes" in recordings_info
        assert "average_size_bytes" in recordings_info
        
        # Verify snapshots breakdown
        assert "total_files" in snapshots_info
        assert "total_size_bytes" in snapshots_info
        assert "average_size_bytes" in snapshots_info
        
        # Verify logical relationships
        assert recordings_info["total_files"] == 3
        assert snapshots_info["total_files"] == 5
        assert recordings_info["total_size_bytes"] > 0
        assert snapshots_info["total_size_bytes"] > 0
    
    async def test_storage_info_performance(self, storage_setup):
        """Test performance of storage info retrieval.
        
        REQ-API-037: Real-time storage space monitoring and alerts
        """
        import time
        
        # Measure response time
        start_time = time.time()
        response = await storage_setup.websocket_client.send_request(
            "get_storage_info",
            {}
        )
        end_time = time.time()
        
        response_time = end_time - start_time
        
        # Verify response is successful
        assert response.get("result") is not None
        
        # Verify response time is reasonable (should be fast for storage info)
        assert response_time < 1.0  # Should complete within 1 second
    
    async def test_storage_info_error_handling(self, storage_setup):
        """Test error handling for storage info retrieval.
        
        REQ-API-037: Real-time storage space monitoring and alerts
        """
        # Test with invalid parameters
        response = await storage_setup.websocket_client.send_request(
            "get_storage_info",
            {
                "invalid_param": "value"
            }
        )
        
        # Should still work (ignore invalid parameters)
        assert response.get("result") is not None
        assert "storage_info" in response["result"]
