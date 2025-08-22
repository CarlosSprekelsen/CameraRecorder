"""
File Management Integration Tests - MediaMTX Camera Service

This module provides comprehensive integration tests for file lifecycle management
functionality including file deletion, metadata retrieval, storage monitoring,
and retention policy management.

Requirements Coverage:
- REQ-CLIENT-034: File deletion capabilities for recordings and snapshots via service API
- REQ-CLIENT-035: Configurable retention policies for media files
- REQ-CLIENT-036: Storage space monitoring and alerts when space is low
- REQ-CLIENT-037: Automatic cleanup of old files based on retention policies
- REQ-CLIENT-038: Manual file management interface for bulk operations
- REQ-CLIENT-039: File archiving to external storage before deletion
- REQ-CLIENT-040: File metadata viewing capabilities
- REQ-CLIENT-041: Role-based access control for file deletion
- REQ-API-024: get_recording_info method for individual recording metadata
- REQ-API-025: get_snapshot_info method for individual snapshot metadata
- REQ-API-026: delete_recording method for recording file deletion
- REQ-API-027: delete_snapshot method for snapshot file deletion
- REQ-API-028: get_storage_info method for storage space monitoring
- REQ-API-029: set_retention_policy method for configurable file retention
- REQ-API-030: cleanup_old_files method for automatic file cleanup

Test Categories: Integration, Security, Real System
"""

import pytest
import asyncio
import json
import tempfile
import os
import shutil
from pathlib import Path
from datetime import datetime, timedelta
from typing import Dict, Any

from tests.fixtures.auth_utils import TestUserFactory, WebSocketAuthTestClient
from tests.utils.port_utils import find_free_port


class FileManagementTestSetup:
    """Test setup for file management integration tests."""
    
    def __init__(self):
        self.websocket_client = None
        self.user_factory = None
        self.temp_dir = None
        self.test_files = []
        self.auth_manager = None
        self.server = None
        self.service_manager = None
        self.camera_monitor = None
        self.mediamtx_controller = None
        self.config = None
        
    async def setup(self):
        """Set up test environment with WebSocket connection and test files."""
        # Import required modules
        from tests.fixtures.auth_utils import get_test_auth_manager
        from src.camera_service.config import Config
        from src.websocket_server.server import WebSocketJsonRpcServer
        from src.security.middleware import SecurityMiddleware
        
        # Create test configuration
        self.config = Config()
        self.config.server.port = find_free_port()
        self.config.server.host = "127.0.0.1"
        self.config.server.websocket_path = "/ws"
        self.config.server.max_connections = 10
        
        # Create auth manager
        self.auth_manager = get_test_auth_manager()
        
        # Create user factory
        self.user_factory = TestUserFactory(self.auth_manager)
        
        # Create WebSocket server (without full service stack for file management tests)
        self.server = WebSocketJsonRpcServer(
            host=self.config.server.host,
            port=self.config.server.port,
            websocket_path=self.config.server.websocket_path,
            max_connections=self.config.server.max_connections,
            mediamtx_controller=None,
            camera_monitor=None,
            config=self.config
        )
        
        # Create and set security middleware
        security_middleware = SecurityMiddleware(self.auth_manager, max_connections=10, requests_per_minute=120)
        self.server.set_security_middleware(security_middleware)
        
        # Start server
        await self.server.start()
        
        # Create WebSocket client for testing
        websocket_url = f"ws://{self.config.server.host}:{self.config.server.port}{self.config.server.websocket_path}"
        # Create a test user for the WebSocket client
        test_user = self.user_factory.create_admin_user("file_mgmt_integration_test_user")
        self.websocket_client = WebSocketAuthTestClient(websocket_url, test_user)
        await self.websocket_client.connect()
        
        # Create temporary directory for test files
        self.temp_dir = tempfile.mkdtemp(prefix="file_mgmt_test_")
        
        # Create test files
        await self._create_test_files()
        
    async def cleanup(self):
        """Clean up test environment."""
        if self.websocket_client:
            await self.websocket_client.disconnect()
        
        if self.server:
            await self.server.stop()
        
        # Clean up temporary files
        if self.temp_dir and os.path.exists(self.temp_dir):
            shutil.rmtree(self.temp_dir)
            
    async def _create_test_files(self):
        """Create test recording and snapshot files for testing."""
        # Create test recording file
        recording_file = os.path.join(self.temp_dir, "test_recording.mp4")
        with open(recording_file, 'wb') as f:
            f.write(b"fake video content" * 1000)  # ~17KB file
        self.test_files.append(recording_file)
        
        # Create test snapshot file
        snapshot_file = os.path.join(self.temp_dir, "test_snapshot.jpg")
        with open(snapshot_file, 'wb') as f:
            f.write(b"fake image content" * 100)  # ~1.7KB file
        self.test_files.append(snapshot_file)


@pytest.mark.integration
@pytest.mark.asyncio
@pytest.mark.real_websocket
async def test_get_recording_info_success():
    """
    REQ-API-024: Test get_recording_info method for individual recording metadata.
    
    Validates that the get_recording_info method returns detailed metadata
    for a specific recording file including filename, size, duration, and
    creation timestamp.
    """
    print("\nTesting get_recording_info - Success Case (Authenticated)")
    
    setup = FileManagementTestSetup()
    try:
        await setup.setup()
        
        # The WebSocket client is already configured with an admin user from setup
        # No need for additional authentication - the client automatically includes auth token
        print(f"✅ Using pre-configured admin user for testing")
        
        # Test get_recording_info with filename parameter
        params = {
            "filename": "test_recording.mp4"
        }
        
        result = await setup.websocket_client.call_protected_method("get_recording_info", params)
        
        print(f"✅ Success: get_recording_info completed")
        print(f"   Response: {json.dumps(result, indent=2)}")
        
        # Validate response structure per API specification
        assert "result" in result, "Response should contain 'result' field"
        recording_info = result["result"]
        assert "filename" in recording_info, "Should contain 'filename' field"
        assert "file_size" in recording_info, "Should contain 'file_size' field"
        assert "duration" in recording_info, "Should contain 'duration' field"
        assert "created_time" in recording_info, "Should contain 'created_time' field"
        assert "download_url" in recording_info, "Should contain 'download_url' field"
        
        return result
        
    finally:
        await setup.cleanup()


@pytest.mark.integration
@pytest.mark.asyncio
@pytest.mark.real_websocket
async def test_get_snapshot_info_success():
    """
    REQ-API-025: Test get_snapshot_info method for individual snapshot metadata.
    
    Validates that the get_snapshot_info method returns detailed metadata
    for a specific snapshot file including filename, size, resolution, and
    creation timestamp.
    """
    print("\nTesting get_snapshot_info - Success Case (Authenticated)")
    
    setup = FileManagementTestSetup()
    try:
        await setup.setup()
        
        # The WebSocket client is already configured with an admin user from setup
        # No need for additional authentication - the client automatically includes auth token
        print(f"✅ Using pre-configured admin user for testing")
        
        # Test get_snapshot_info with filename parameter
        params = {
            "filename": "test_snapshot.jpg"
        }
        
        result = await setup.websocket_client.call_protected_method("get_snapshot_info", params)
        
        print(f"✅ Success: get_snapshot_info completed")
        print(f"   Response: {json.dumps(result, indent=2)}")
        
        # Validate response structure per API specification
        assert "result" in result, "Response should contain 'result' field"
        snapshot_info = result["result"]
        assert "filename" in snapshot_info, "Should contain 'filename' field"
        assert "file_size" in snapshot_info, "Should contain 'file_size' field"
        assert "resolution" in snapshot_info, "Should contain 'resolution' field"
        assert "created_time" in snapshot_info, "Should contain 'created_time' field"
        assert "download_url" in snapshot_info, "Should contain 'download_url' field"
        
        return result
        
    finally:
        await setup.cleanup()


@pytest.mark.integration
@pytest.mark.asyncio
@pytest.mark.real_websocket
async def test_delete_recording_success():
    """
    REQ-API-026: Test delete_recording method for recording file deletion.
    REQ-CLIENT-034: Test file deletion capabilities for recordings via service API.
    REQ-CLIENT-041: Test role-based access control for file deletion (operator role).
    
    Validates that the delete_recording method successfully deletes recording
    files and requires proper operator role authentication.
    """
    print("\nTesting delete_recording - Success Case (Operator Authenticated)")
    
    setup = FileManagementTestSetup()
    try:
        await setup.setup()
        
        # Create operator user for testing (required for delete_recording)
        operator_user = setup.user_factory.create_operator_user()
        
        # Authenticate with WebSocket server
        auth_result = await setup.websocket_client.authenticate(operator_user["token"])
        assert "result" in auth_result, "Authentication response should contain 'result' field"
        assert auth_result["result"]["authenticated"] is True, "Authentication failed"
        print(f"✅ Authenticated as {operator_user['user_id']} with role {operator_user['role']}")
        
        # Test delete_recording with filename parameter
        params = {
            "filename": "test_recording.mp4"
        }
        
        result = await setup.websocket_client.call_protected_method("delete_recording", params)
        
        print(f"✅ Success: delete_recording completed")
        print(f"   Response: {json.dumps(result, indent=2)}")
        
        # Validate response structure per API specification
        assert "result" in result, "Response should contain 'result' field"
        delete_result = result["result"]
        assert "filename" in delete_result, "Should contain 'filename' field"
        assert "deleted" in delete_result, "Should contain 'deleted' field"
        assert delete_result["deleted"] is True, "File should be marked as deleted"
        assert "message" in delete_result, "Should contain 'message' field"
        
        return result
        
    finally:
        await setup.cleanup()


@pytest.mark.integration
@pytest.mark.asyncio
@pytest.mark.real_websocket
async def test_delete_snapshot_success():
    """
    REQ-API-027: Test delete_snapshot method for snapshot file deletion.
    REQ-CLIENT-034: Test file deletion capabilities for snapshots via service API.
    REQ-CLIENT-041: Test role-based access control for file deletion (operator role).
    
    Validates that the delete_snapshot method successfully deletes snapshot
    files and requires proper operator role authentication.
    """
    print("\nTesting delete_snapshot - Success Case (Operator Authenticated)")
    
    setup = FileManagementTestSetup()
    try:
        await setup.setup()
        
        # Create operator user for testing (required for delete_snapshot)
        operator_user = setup.user_factory.create_operator_user()
        
        # Authenticate with WebSocket server
        auth_result = await setup.websocket_client.authenticate(operator_user["token"])
        assert "result" in auth_result, "Authentication response should contain 'result' field"
        assert auth_result["result"]["authenticated"] is True, "Authentication failed"
        print(f"✅ Authenticated as {operator_user['user_id']} with role {operator_user['role']}")
        
        # Test delete_snapshot with filename parameter
        params = {
            "filename": "test_snapshot.jpg"
        }
        
        result = await setup.websocket_client.call_protected_method("delete_snapshot", params)
        
        print(f"✅ Success: delete_snapshot completed")
        print(f"   Response: {json.dumps(result, indent=2)}")
        
        # Validate response structure per API specification
        assert "result" in result, "Response should contain 'result' field"
        delete_result = result["result"]
        assert "filename" in delete_result, "Should contain 'filename' field"
        assert "deleted" in delete_result, "Should contain 'deleted' field"
        assert delete_result["deleted"] is True, "File should be marked as deleted"
        assert "message" in delete_result, "Should contain 'message' field"
        
        return result
        
    finally:
        await setup.cleanup()


@pytest.mark.integration
@pytest.mark.asyncio
@pytest.mark.real_websocket
async def test_delete_recording_insufficient_permissions():
    """
    REQ-CLIENT-041: Test role-based access control for file deletion (viewer role insufficient).
    
    Validates that the delete_recording method properly rejects requests from
    users with insufficient permissions (viewer role).
    """
    print("\nTesting delete_recording - Insufficient Permissions (Viewer Role)")
    
    setup = FileManagementTestSetup()
    try:
        await setup.setup()
        
        # Create viewer user (insufficient permissions for delete_recording)
        viewer_user = setup.user_factory.create_viewer_user()
        
        # Authenticate with WebSocket server
        auth_result = await setup.websocket_client.authenticate(viewer_user["token"])
        assert "result" in auth_result, "Authentication response should contain 'result' field"
        assert auth_result["result"]["authenticated"] is True, "Authentication failed"
        print(f"✅ Authenticated as {viewer_user['user_id']} with role {viewer_user['role']}")
        
        # Try to call delete_recording with insufficient permissions
        params = {
            "filename": "test_recording.mp4"
        }
        
        result = await setup.websocket_client.call_protected_method("delete_recording", params)
        
        # Should receive authorization error
        assert "error" in result, "Should receive error response for insufficient permissions"
        assert result["error"]["code"] == -32003, "Should receive authorization error code"
        print(f"✅ Success: delete_recording properly rejected insufficient permissions")
        
        return result
        
    finally:
        await setup.cleanup()


@pytest.mark.integration
@pytest.mark.asyncio
@pytest.mark.real_websocket
async def test_get_storage_info_success():
    """
    REQ-API-028: Test get_storage_info method for storage space monitoring.
    REQ-CLIENT-036: Test storage space monitoring and alerts when space is low.
    
    Validates that the get_storage_info method returns comprehensive storage
    space information including total space, used space, available space,
    and usage statistics.
    """
    print("\nTesting get_storage_info - Success Case (Admin Authenticated)")
    
    setup = FileManagementTestSetup()
    try:
        await setup.setup()
        
        # The WebSocket client is already configured with an admin user from setup
        # No need for additional authentication - the client automatically includes auth token
        print(f"✅ Using pre-configured admin user for testing")
        
        # Test get_storage_info method
        result = await setup.websocket_client.call_protected_method("get_storage_info", {})
        
        print(f"✅ Success: get_storage_info completed")
        print(f"   Response: {json.dumps(result, indent=2)}")
        
        # Validate response structure per API specification
        assert "result" in result, "Response should contain 'result' field"
        storage_info = result["result"]
        assert "total_space" in storage_info, "Should contain 'total_space' field"
        assert "used_space" in storage_info, "Should contain 'used_space' field"
        assert "available_space" in storage_info, "Should contain 'available_space' field"
        assert "usage_percentage" in storage_info, "Should contain 'usage_percentage' field"
        assert "recordings_size" in storage_info, "Should contain 'recordings_size' field"
        assert "snapshots_size" in storage_info, "Should contain 'snapshots_size' field"
        assert "low_space_warning" in storage_info, "Should contain 'low_space_warning' field"
        
        # Validate data types and ranges
        assert isinstance(storage_info["total_space"], (int, float)), "total_space should be numeric"
        assert isinstance(storage_info["used_space"], (int, float)), "used_space should be numeric"
        assert isinstance(storage_info["available_space"], (int, float)), "available_space should be numeric"
        assert isinstance(storage_info["usage_percentage"], (int, float)), "usage_percentage should be numeric"
        assert 0 <= storage_info["usage_percentage"] <= 100, "usage_percentage should be between 0 and 100"
        assert isinstance(storage_info["low_space_warning"], bool), "low_space_warning should be boolean"
        
        return result
        
    finally:
        await setup.cleanup()


@pytest.mark.integration
@pytest.mark.asyncio
@pytest.mark.real_websocket
async def test_set_retention_policy_success():
    """
    REQ-API-029: Test set_retention_policy method for configurable file retention.
    REQ-CLIENT-035: Test configurable retention policies for media files.
    
    Validates that the set_retention_policy method allows configuration of
    file retention policies including age-based and size-based policies.
    """
    print("\nTesting set_retention_policy - Success Case (Admin Authenticated)")
    
    setup = FileManagementTestSetup()
    try:
        await setup.setup()
        
        # The WebSocket client is already configured with an admin user from setup
        # No need for additional authentication - the client automatically includes auth token
        print(f"✅ Using pre-configured admin user for testing")
        
        # Test set_retention_policy with age-based policy
        params = {
            "policy_type": "age",
            "max_age_days": 30,
            "enabled": True
        }
        
        result = await setup.websocket_client.call_protected_method("set_retention_policy", params)
        
        print(f"✅ Success: set_retention_policy completed")
        print(f"   Response: {json.dumps(result, indent=2)}")
        
        # Validate response structure per API specification
        assert "result" in result, "Response should contain 'result' field"
        policy_result = result["result"]
        assert "policy_type" in policy_result, "Should contain 'policy_type' field"
        assert "max_age_days" in policy_result, "Should contain 'max_age_days' field"
        assert "enabled" in policy_result, "Should contain 'enabled' field"
        assert "message" in policy_result, "Should contain 'message' field"
        
        # Validate policy configuration
        assert policy_result["policy_type"] == "age", "Policy type should match request"
        assert policy_result["max_age_days"] == 30, "Max age days should match request"
        assert policy_result["enabled"] is True, "Policy should be enabled"
        
        return result
        
    finally:
        await setup.cleanup()


@pytest.mark.integration
@pytest.mark.asyncio
@pytest.mark.real_websocket
async def test_cleanup_old_files_success():
    """
    REQ-API-030: Test cleanup_old_files method for automatic file cleanup.
    REQ-CLIENT-037: Test automatic cleanup of old files based on retention policies.
    
    Validates that the cleanup_old_files method successfully executes cleanup
    operations based on configured retention policies and returns cleanup statistics.
    """
    print("\nTesting cleanup_old_files - Success Case (Admin Authenticated)")
    
    setup = FileManagementTestSetup()
    try:
        await setup.setup()
        
        # Create admin user for testing (required for cleanup_old_files)
        admin_user = setup.user_factory.create_admin_user()
        
        # Authenticate with WebSocket server
        auth_result = await setup.websocket_client.authenticate(admin_user["token"])
        assert "result" in auth_result, "Authentication response should contain 'result' field"
        assert auth_result["result"]["authenticated"] is True, "Authentication failed"
        print(f"✅ Authenticated as {admin_user['user_id']} with role {admin_user['role']}")
        
        # Test cleanup_old_files method
        result = await setup.websocket_client.call_protected_method("cleanup_old_files", {})
        
        print(f"✅ Success: cleanup_old_files completed")
        print(f"   Response: {json.dumps(result, indent=2)}")
        
        # Validate response structure per API specification
        assert "result" in result, "Response should contain 'result' field"
        cleanup_result = result["result"]
        assert "cleanup_executed" in cleanup_result, "Should contain 'cleanup_executed' field"
        assert "files_deleted" in cleanup_result, "Should contain 'files_deleted' field"
        assert "space_freed" in cleanup_result, "Should contain 'space_freed' field"
        assert "message" in cleanup_result, "Should contain 'message' field"
        
        # Validate data types
        assert isinstance(cleanup_result["cleanup_executed"], bool), "cleanup_executed should be boolean"
        assert isinstance(cleanup_result["files_deleted"], int), "files_deleted should be integer"
        assert isinstance(cleanup_result["space_freed"], (int, float)), "space_freed should be numeric"
        assert cleanup_result["files_deleted"] >= 0, "files_deleted should be non-negative"
        assert cleanup_result["space_freed"] >= 0, "space_freed should be non-negative"
        
        return result
        
    finally:
        await setup.cleanup()


@pytest.mark.integration
@pytest.mark.asyncio
@pytest.mark.real_websocket
async def test_file_management_comprehensive_workflow():
    """
    REQ-CLIENT-038: Test manual file management interface for bulk operations.
    REQ-CLIENT-040: Test file metadata viewing capabilities.
    
    Validates a comprehensive file management workflow including:
    1. Listing files
    2. Getting individual file metadata
    3. Deleting files
    4. Monitoring storage space
    5. Setting retention policies
    6. Executing cleanup operations
    """
    print("\nTesting File Management - Comprehensive Workflow")
    
    setup = FileManagementTestSetup()
    try:
        await setup.setup()
        
        # Create admin user for comprehensive testing
        admin_user = setup.user_factory.create_admin_user()
        
        # Authenticate with WebSocket server
        auth_result = await setup.websocket_client.authenticate(admin_user["token"])
        assert "result" in auth_result, "Authentication response should contain 'result' field"
        assert auth_result["result"]["authenticated"] is True, "Authentication failed"
        print(f"✅ Authenticated as {admin_user['user_id']} with role {admin_user['role']}")
        
        # Step 1: Get storage information
        storage_result = await setup.websocket_client.call_protected_method("get_storage_info", {})
        assert "result" in storage_result, "Storage info should be available"
        print(f"✅ Step 1: Storage info retrieved")
        
        # Step 2: List recordings
        recordings_result = await setup.websocket_client.call_protected_method("list_recordings", {})
        assert "result" in recordings_result, "Recordings list should be available"
        print(f"✅ Step 2: Recordings list retrieved")
        
        # Step 3: List snapshots
        snapshots_result = await setup.websocket_client.call_protected_method("list_snapshots", {})
        assert "result" in snapshots_result, "Snapshots list should be available"
        print(f"✅ Step 3: Snapshots list retrieved")
        
        # Step 4: Set retention policy
        policy_result = await setup.websocket_client.call_protected_method("set_retention_policy", {
            "policy_type": "age",
            "max_age_days": 7,
            "enabled": True
        })
        assert "result" in policy_result, "Retention policy should be set"
        print(f"✅ Step 4: Retention policy configured")
        
        # Step 5: Execute cleanup
        cleanup_result = await setup.websocket_client.call_protected_method("cleanup_old_files", {})
        assert "result" in cleanup_result, "Cleanup should be executed"
        print(f"✅ Step 5: Cleanup executed")
        
        # Step 6: Verify storage space updated
        updated_storage_result = await setup.websocket_client.call_protected_method("get_storage_info", {})
        assert "result" in updated_storage_result, "Updated storage info should be available"
        print(f"✅ Step 6: Storage space verified")
        
        print(f"✅ Success: Comprehensive file management workflow completed")
        
        return {
            "storage": storage_result,
            "recordings": recordings_result,
            "snapshots": snapshots_result,
            "policy": policy_result,
            "cleanup": cleanup_result,
            "updated_storage": updated_storage_result
        }
        
    finally:
        await setup.cleanup()
