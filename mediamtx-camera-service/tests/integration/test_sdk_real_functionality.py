"""
Real integration tests for SDK functionality.

Requirements Traceability:
- REQ-SDK-001: SDK shall be installable and functional
- REQ-SDK-002: SDK shall provide client classes with authentication
- REQ-SDK-003: SDK shall handle errors gracefully
- REQ-SDK-004: SDK shall be properly documented

Story Coverage: S8.3 - SDK Development
IV&V Control Point: SDK validation
"""

import pytest
import pytest_asyncio
import sys
import os
from pathlib import Path

# Add SDK directory to path for testing
sdk_path = Path(__file__).parent.parent.parent / "sdk" / "python"
sys.path.insert(0, str(sdk_path))

try:
    from mediamtx_camera_sdk import CameraClient, CameraInfo, RecordingInfo, SnapshotInfo
    from mediamtx_camera_sdk.exceptions import (
        CameraServiceError, AuthenticationError, ConnectionError, 
        CameraNotFoundError, MediaMTXError, TimeoutError, ValidationError
    )
    SDK_AVAILABLE = True
except ImportError:
    SDK_AVAILABLE = False


@pytest.mark.skipif(not SDK_AVAILABLE, reason="Python SDK not available")
@pytest.mark.integration
class TestSDKRealFunctionality:
    """Real integration tests for SDK functionality using actual server."""
    
    @pytest.fixture
    def client_config(self):
        """Test client configuration."""
        return {
            "host": "localhost",
            "port": 8002,
            "use_ssl": False,
            "auth_type": "jwt",
            "auth_token": "invalid_token",  # Will fail auth but we can test real behavior
            "max_retries": 1,
            "retry_delay": 0.1
        }
    
    @pytest_asyncio.fixture
    async def client(self, client_config):
        """Test client instance."""
        client = CameraClient(**client_config)
        yield client
        # Cleanup
        if hasattr(client, 'websocket') and client.websocket:
            await client.disconnect()
    
    def test_sdk_imports_correctly(self):
        """Test that SDK imports correctly."""
        # This test verifies that all SDK components can be imported
        assert CameraClient is not None
        assert CameraInfo is not None
        assert RecordingInfo is not None
        assert SnapshotInfo is not None
        
        # Test exception classes
        assert AuthenticationError is not None
        assert ConnectionError is not None
        assert CameraNotFoundError is not None
        assert CameraServiceError is not None
        assert MediaMTXError is not None
        assert TimeoutError is not None
        assert ValidationError is not None
    
    def test_sdk_client_initialization(self, client_config):
        """Test that SDK client initializes correctly."""
        client = CameraClient(**client_config)
        
        # Verify configuration is set correctly
        assert client.host == "localhost"
        assert client.port == 8002
        assert client.use_ssl == False
        assert client.auth_type == "jwt"
        assert client.auth_token == "invalid_token"
        assert client.max_retries == 1
        assert client.retry_delay == 0.1
        
        # Verify initial state
        assert not client.connected
        assert not client.authenticated
        assert client.websocket is None
    
    def test_sdk_client_api_key_initialization(self, client_config):
        """Test that SDK client initializes correctly with API key."""
        api_key_config = client_config.copy()
        api_key_config.update({
            "auth_type": "api_key",
            "api_key": "invalid_api_key"
        })
        
        client = CameraClient(**api_key_config)
        
        # Verify configuration is set correctly
        assert client.auth_type == "api_key"
        assert client.api_key == "invalid_api_key"
        # Note: auth_token might still be set from the original config, that's OK
    
    @pytest.mark.asyncio
    async def test_sdk_authentication_failure_real_server(self, client):
        """Test that SDK handles authentication failure with real server."""
        # This should connect to the WebSocket but fail authentication
        with pytest.raises(AuthenticationError):
            await client.connect()
        
        # Verify client state after failed authentication
        assert not client.authenticated
        # Note: connected might be True briefly during connection attempt, that's OK
    
    @pytest.mark.asyncio
    async def test_sdk_connection_failure_invalid_host(self, client_config):
        """Test that SDK handles connection failure with invalid host."""
        invalid_config = client_config.copy()
        invalid_config.update({
            "host": "nonexistent.host.local",
            "port": 9999
        })
        
        client = CameraClient(**invalid_config)
        
        with pytest.raises(ConnectionError):
            await client.connect()
        
        # Verify client state after failed connection
        assert not client.connected
        assert not client.authenticated
    
    @pytest.mark.asyncio
    async def test_sdk_connection_failure_invalid_port(self, client_config):
        """Test that SDK handles connection failure with invalid port."""
        invalid_config = client_config.copy()
        invalid_config.update({
            "port": 9999  # Non-existent port
        })
        
        client = CameraClient(**invalid_config)
        
        with pytest.raises(ConnectionError):
            await client.connect()
        
        # Verify client state after failed connection
        assert not client.connected
        assert not client.authenticated
    
    def test_sdk_client_methods_exist(self, client):
        """Test that SDK client has all required methods."""
        # Test that all required methods exist
        assert hasattr(client, 'connect')
        assert hasattr(client, 'disconnect')
        assert hasattr(client, 'ping')
        assert hasattr(client, 'get_camera_list')
        assert hasattr(client, 'get_camera_status')
        assert hasattr(client, 'take_snapshot')
        assert hasattr(client, 'start_recording')
        assert hasattr(client, 'stop_recording')
        assert hasattr(client, 'get_recording_status')
        
        # Test that methods are callable
        assert callable(client.connect)
        assert callable(client.disconnect)
        assert callable(client.ping)
        assert callable(client.get_camera_list)
        assert callable(client.get_camera_status)
        assert callable(client.take_snapshot)
        assert callable(client.start_recording)
        assert callable(client.stop_recording)
        assert callable(client.get_recording_status)
    
    def test_sdk_data_models(self):
        """Test that SDK data models work correctly."""
        # Test CameraInfo
        camera = CameraInfo(
            device_path="/dev/video0",
            name="Test Camera",
            capabilities=["snapshot", "recording"],
            status="CONNECTED",
            stream_url="rtsp://localhost:8554/camera0"
        )
        
        assert camera.device_path == "/dev/video0"
        assert camera.name == "Test Camera"
        assert camera.capabilities == ["snapshot", "recording"]
        assert camera.status == "CONNECTED"
        assert camera.stream_url == "rtsp://localhost:8554/camera0"
        
        # Test RecordingInfo
        recording = RecordingInfo(
            device_path="/dev/video0",
            recording_id="rec_123",
            filename="test.mp4",
            start_time=1234567890.0,
            duration=60.0,
            status="active"
        )
        
        assert recording.device_path == "/dev/video0"
        assert recording.recording_id == "rec_123"
        assert recording.filename == "test.mp4"
        assert recording.start_time == 1234567890.0
        assert recording.duration == 60.0
        assert recording.status == "active"
        
        # Test SnapshotInfo
        snapshot = SnapshotInfo(
            device_path="/dev/video0",
            filename="snapshot.jpg",
            timestamp=1234567890.0,
            size_bytes=1024
        )
        
        assert snapshot.device_path == "/dev/video0"
        assert snapshot.filename == "snapshot.jpg"
        assert snapshot.timestamp == 1234567890.0
        assert snapshot.size_bytes == 1024
    
    def test_sdk_exception_hierarchy(self):
        """Test that SDK exception hierarchy is correct."""
        # Test that exceptions inherit from the correct base class
        assert issubclass(AuthenticationError, CameraServiceError)
        assert issubclass(ConnectionError, CameraServiceError)
        assert issubclass(CameraNotFoundError, CameraServiceError)
        assert issubclass(MediaMTXError, CameraServiceError)
        assert issubclass(TimeoutError, CameraServiceError)
        assert issubclass(ValidationError, CameraServiceError)
        
        # Test that exceptions can be instantiated
        auth_error = AuthenticationError("Authentication failed")
        conn_error = ConnectionError("Connection failed")
        not_found_error = CameraNotFoundError("Camera not found")
        
        assert str(auth_error) == "Authentication failed"
        assert str(conn_error) == "Connection failed"
        assert str(not_found_error) == "Camera not found"
    
    def test_sdk_client_configuration_validation(self, client_config):
        """Test that SDK client validates configuration correctly."""
        # Test with valid configuration
        client = CameraClient(**client_config)
        assert client is not None
        
        # Test with minimal configuration
        minimal_config = {
            "host": "localhost",
            "port": 8002
        }
        client = CameraClient(**minimal_config)
        assert client.host == "localhost"
        assert client.port == 8002
        assert client.auth_type == "jwt"  # Default value
        assert client.auth_token is None
        assert client.api_key is None
    
    @pytest.mark.asyncio
    async def test_sdk_client_disconnect_safety(self, client):
        """Test that SDK client handles disconnect safely."""
        # Disconnect should not raise an exception even if not connected
        await client.disconnect()
        
        # Verify state after disconnect
        assert not client.connected
        assert not client.authenticated
        assert client.websocket is None
    
    def test_sdk_client_url_generation(self, client):
        """Test that SDK client generates correct URLs."""
        # Test WebSocket URL generation
        ws_url = client._get_ws_url()
        assert ws_url == "ws://localhost:8002/ws"
        
        # Test with SSL
        client.use_ssl = True
        ws_url_ssl = client._get_ws_url()
        assert ws_url_ssl == "wss://localhost:8002/ws"
        
        # Test with different host/port
        client.host = "example.com"
        client.port = 9000
        ws_url_custom = client._get_ws_url()
        assert ws_url_custom == "wss://example.com:9000/ws"
    
    def test_sdk_client_retry_configuration(self, client_config):
        """Test that SDK client handles retry configuration correctly."""
        # Test with different retry configurations
        config_no_retries = client_config.copy()
        config_no_retries.update({
            "max_retries": 0,
            "retry_delay": 0.1
        })
        
        client = CameraClient(**config_no_retries)
        assert client.max_retries == 0
        assert client.retry_delay == 0.1
        
        # Test with high retry count
        config_high_retries = client_config.copy()
        config_high_retries.update({
            "max_retries": 10,
            "retry_delay": 2.0
        })
        
        client = CameraClient(**config_high_retries)
        assert client.max_retries == 10
        assert client.retry_delay == 2.0
