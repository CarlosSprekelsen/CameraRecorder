"""
Real integration tests for client examples.

Requirements Traceability:
- REQ-CLIENT-001: Client examples shall be fully functional and tested
- REQ-CLIENT-002: Client examples shall demonstrate proper authentication
- REQ-CLIENT-003: Client examples shall handle errors gracefully

Story Coverage: S8.1 - Client Usage Examples
IV&V Control Point: Client examples validation
"""

import pytest
import pytest_asyncio
import sys
import os
from pathlib import Path

# Add examples directory to path for testing
examples_path = Path(__file__).parent.parent.parent / "examples" / "python"
sys.path.insert(0, str(examples_path))

try:
    from camera_client import CameraClient, CameraInfo, CameraNotFoundError, AuthenticationError
    CLIENT_AVAILABLE = True
except ImportError:
    CLIENT_AVAILABLE = False


@pytest.mark.skipif(not CLIENT_AVAILABLE, reason="Python client example not available")
@pytest.mark.integration
class TestClientExamplesReal:
    """Real integration tests for client examples using actual server."""
    
    @pytest.fixture
    def client_config(self):
        """Test client configuration."""
        return {
            "host": "localhost",
            "port": 8080,
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
    
    def test_client_imports_correctly(self):
        """Test that client example imports correctly."""
        # This test verifies that all client components can be imported
        assert CameraClient is not None
        assert CameraInfo is not None
        assert CameraNotFoundError is not None
        assert AuthenticationError is not None
    
    def test_client_initialization(self, client_config):
        """Test client initialization with various configurations."""
        # Test JWT authentication
        client = CameraClient(**client_config)
        assert client.host == "localhost"
        assert client.port == 8080
        assert client.auth_type == "jwt"
        assert client.auth_token == "invalid_token"
        assert not client.authenticated
        
        # Test API key authentication
        api_key_config = client_config.copy()
        api_key_config.update({
            "auth_type": "api_key",
            "api_key": "invalid_api_key"
        })
        client = CameraClient(**api_key_config)
        assert client.auth_type == "api_key"
        assert client.api_key == "invalid_api_key"
    
    @pytest.mark.asyncio
    async def test_client_authentication_failure_real_server(self, client):
        """Test that client handles authentication failure with real server."""
        # This should connect to the WebSocket but fail authentication
        # The client example wraps AuthenticationError in ConnectionError
        with pytest.raises(Exception):  # Either AuthenticationError or ConnectionError
            await client.connect()
        
        # Verify client state after failed authentication
        assert not client.authenticated
    
    @pytest.mark.asyncio
    async def test_client_connection_failure_invalid_host(self, client_config):
        """Test that client handles connection failure with invalid host."""
        invalid_config = client_config.copy()
        invalid_config.update({
            "host": "nonexistent.host.local",
            "port": 9999
        })
        
        client = CameraClient(**invalid_config)
        
        with pytest.raises(Exception):  # Should raise some connection error
            await client.connect()
        
        # Verify client state after failed connection
        assert not client.connected
        assert not client.authenticated
    
    @pytest.mark.asyncio
    async def test_client_connection_failure_invalid_port(self, client_config):
        """Test that client handles connection failure with invalid port."""
        invalid_config = client_config.copy()
        invalid_config.update({
            "port": 9999  # Non-existent port
        })
        
        client = CameraClient(**invalid_config)
        
        with pytest.raises(Exception):  # Should raise some connection error
            await client.connect()
        
        # Verify client state after failed connection
        assert not client.connected
        assert not client.authenticated
    
    def test_client_methods_exist(self, client):
        """Test that client has all required methods."""
        # Test that all required methods exist
        assert hasattr(client, 'connect')
        assert hasattr(client, 'disconnect')
        assert hasattr(client, 'ping')
        assert hasattr(client, 'get_camera_list')
        assert hasattr(client, 'get_camera_status')
        assert hasattr(client, 'take_snapshot')
        assert hasattr(client, 'start_recording')
        assert hasattr(client, 'stop_recording')
        
        # Test that methods are callable
        assert callable(client.connect)
        assert callable(client.disconnect)
        assert callable(client.ping)
        assert callable(client.get_camera_list)
        assert callable(client.get_camera_status)
        assert callable(client.take_snapshot)
        assert callable(client.start_recording)
        assert callable(client.stop_recording)
    
    def test_client_data_models(self):
        """Test that client data models work correctly."""
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
    
    def test_client_exception_classes(self):
        """Test that client exception classes work correctly."""
        # Test that exceptions can be instantiated
        auth_error = AuthenticationError("Authentication failed")
        not_found_error = CameraNotFoundError("Camera not found")
        
        assert str(auth_error) == "Authentication failed"
        assert str(not_found_error) == "Camera not found"
        
        # Test that exceptions are the correct type
        assert isinstance(auth_error, AuthenticationError)
        assert isinstance(not_found_error, CameraNotFoundError)
    
    def test_client_configuration_validation(self, client_config):
        """Test that client validates configuration correctly."""
        # Test with valid configuration
        client = CameraClient(**client_config)
        assert client is not None
        
        # Test with minimal configuration
        minimal_config = {
            "host": "localhost",
            "port": 8080
        }
        client = CameraClient(**minimal_config)
        assert client.host == "localhost"
        assert client.port == 8080
        assert client.auth_type == "jwt"  # Default value
        assert client.auth_token is None
        assert client.api_key is None
    
    @pytest.mark.asyncio
    async def test_client_disconnect_safety(self, client):
        """Test that client handles disconnect safely."""
        # Disconnect should not raise an exception even if not connected
        await client.disconnect()
        
        # Verify state after disconnect
        assert not client.connected
        assert not client.authenticated
        assert client.websocket is None
    
    def test_client_url_generation(self, client):
        """Test that client generates correct URLs."""
        # Test WebSocket URL generation
        ws_url = client._get_ws_url()
        assert ws_url == "ws://localhost:8080/ws"
        
        # Test with SSL
        client.use_ssl = True
        ws_url_ssl = client._get_ws_url()
        assert ws_url_ssl == "wss://localhost:8080/ws"
        
        # Test with different host/port
        client.host = "example.com"
        client.port = 9000
        ws_url_custom = client._get_ws_url()
        assert ws_url_custom == "wss://example.com:9000/ws"
    
    def test_client_retry_configuration(self, client_config):
        """Test that client handles retry configuration correctly."""
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
    
    def test_client_file_structure(self):
        """Test that client example file has proper structure."""
        client_file = Path(__file__).parent.parent.parent / "examples" / "python" / "camera_client.py"
        
        # Check that file exists and is readable
        assert client_file.exists(), f"Client file not found: {client_file}"
        assert client_file.is_file(), f"Client path is not a file: {client_file}"
        
        # Check file size (should be substantial)
        file_size = client_file.stat().st_size
        assert file_size > 1000, f"Client file seems too small: {file_size} bytes"
        
        # Check file content has expected patterns
        content = client_file.read_text()
        assert "class CameraClient" in content, "Missing CameraClient class"
        assert "class CameraInfo" in content, "Missing CameraInfo class"
        assert "async def connect" in content, "Missing connect method"
        assert "async def get_camera_list" in content, "Missing get_camera_list method"
