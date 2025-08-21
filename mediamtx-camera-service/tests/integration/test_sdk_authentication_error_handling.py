"""
Integration test for Python SDK authentication error handling.

Requirements Traceability:
- REQ-SDK-003: SDK shall handle errors gracefully
- REQ-AUTH-001: Authentication shall work with JWT tokens
- REQ-AUTH-002: Authentication shall work with API keys

Story Coverage: S8.3 - SDK Development
IV&V Control Point: SDK validation
"""

import pytest
import pytest_asyncio
import asyncio
import sys
import os
from pathlib import Path

# Add SDK directory to path for testing
sdk_path = Path(__file__).parent.parent.parent / "sdk" / "python"
sys.path.insert(0, str(sdk_path))

try:
    from mediamtx_camera_sdk import CameraClient
    from mediamtx_camera_sdk.exceptions import (
        CameraServiceError, AuthenticationError, ConnectionError, 
        CameraNotFoundError, MediaMTXError, TimeoutError, ValidationError
    )
    SDK_AVAILABLE = True
except ImportError:
    SDK_AVAILABLE = False


@pytest.mark.skipif(not SDK_AVAILABLE, reason="Python SDK not available")
@pytest.mark.integration
class TestSDKAuthenticationErrorHandling:
    """Integration test for SDK authentication error handling with real server."""
    
    @pytest.fixture
    def client_config(self):
        """Test client configuration."""
        return {
            "host": "localhost",
            "port": 8012,
            "use_ssl": False,
            "auth_type": "jwt",
            "auth_token": "invalid_token",
            "max_retries": 1,  # Single attempt for faster testing
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
    
    @pytest.mark.asyncio
    async def test_invalid_jwt_token_raises_authentication_error(self, client):
        """Test that invalid JWT token raises AuthenticationError with real server."""
        # This should connect to the WebSocket but fail authentication
        with pytest.raises(AuthenticationError):
            await client.connect()
        
        assert not client.authenticated
    
    @pytest.mark.asyncio
    async def test_invalid_api_key_raises_authentication_error(self, client_config):
        """Test that invalid API key raises AuthenticationError with real server."""
        # Configure client with invalid API key
        api_key_config = client_config.copy()
        api_key_config.update({
            "auth_type": "api_key",
            "api_key": "invalid_api_key"
        })
        
        client = CameraClient(**api_key_config)
        
        try:
            with pytest.raises(AuthenticationError):
                await client.connect()
            
            assert not client.authenticated
        finally:
            if hasattr(client, 'websocket') and client.websocket:
                await client.disconnect()
    
    @pytest.mark.asyncio
    async def test_no_auth_token_raises_authentication_error(self, client_config):
        """Test that no auth token raises AuthenticationError with real server."""
        # Configure client with no auth token
        no_auth_config = client_config.copy()
        no_auth_config.update({
            "auth_token": None,
            "api_key": None
        })
        
        client = CameraClient(**no_auth_config)
        
        try:
            with pytest.raises(AuthenticationError):
                await client.connect()
            
            assert not client.authenticated
        finally:
            if hasattr(client, 'websocket') and client.websocket:
                await client.disconnect()
    
    @pytest.mark.asyncio
    async def test_empty_auth_token_raises_authentication_error(self, client_config):
        """Test that empty auth token raises AuthenticationError with real server."""
        # Configure client with empty auth token
        empty_auth_config = client_config.copy()
        empty_auth_config.update({
            "auth_token": ""
        })
        
        client = CameraClient(**empty_auth_config)
        
        try:
            with pytest.raises(AuthenticationError):
                await client.connect()
            
            assert not client.authenticated
        finally:
            if hasattr(client, 'websocket') and client.websocket:
                await client.disconnect()
    
    @pytest.mark.asyncio
    async def test_malformed_jwt_token_raises_authentication_error(self, client_config):
        """Test that malformed JWT token raises AuthenticationError with real server."""
        # Configure client with malformed JWT token
        malformed_config = client_config.copy()
        malformed_config.update({
            "auth_token": "not.a.valid.jwt.token"
        })
        
        client = CameraClient(**malformed_config)
        
        try:
            with pytest.raises(AuthenticationError):
                await client.connect()
            
            assert not client.authenticated
        finally:
            if hasattr(client, 'websocket') and client.websocket:
                await client.disconnect()
    
    @pytest.mark.asyncio
    async def test_authentication_error_message_contains_details(self, client):
        """Test that authentication error message contains useful details."""
        try:
            await client.connect()
            pytest.fail("Expected AuthenticationError to be raised")
        except AuthenticationError as e:
            # Verify the error message contains useful information
            error_message = str(e)
            assert "Authentication" in error_message or "auth" in error_message.lower()
            assert len(error_message) > 10  # Should have meaningful content
    
    @pytest.mark.asyncio
    async def test_authentication_error_type_is_correct(self, client):
        """Test that the correct exception type is raised."""
        try:
            await client.connect()
            pytest.fail("Expected AuthenticationError to be raised")
        except AuthenticationError:
            # This is the expected exception type
            pass
        except Exception as e:
            # Any other exception type is wrong
            pytest.fail(f"Expected AuthenticationError but got {type(e).__name__}: {e}")
    
    @pytest.mark.asyncio
    async def test_connection_succeeds_with_valid_auth(self, client_config):
        """Test that connection succeeds with valid authentication (if available)."""
        # This test requires a valid token, so we'll skip if not available
        # In a real environment, you might have a test token
        pytest.skip("Requires valid authentication token for testing")
        
        # If we had a valid token, the test would look like this:
        # valid_config = client_config.copy()
        # valid_config.update({
        #     "auth_token": "valid_jwt_token_here"
        # })
        # 
        # client = CameraClient(**valid_config)
        # 
        # try:
        #     await client.connect()
        #     assert client.authenticated
        #     assert client.connected
        # finally:
        #     await client.disconnect()
    
    @pytest.mark.asyncio
    async def test_authentication_error_does_not_leave_connection_open(self, client):
        """Test that authentication failure doesn't leave connection open."""
        try:
            await client.connect()
            pytest.fail("Expected AuthenticationError to be raised")
        except AuthenticationError:
            # Verify connection is properly closed
            assert not client.connected
            assert not client.authenticated
            assert client.websocket is None
    
    @pytest.mark.asyncio
    async def test_multiple_authentication_attempts_consistent_behavior(self, client):
        """Test that multiple authentication attempts have consistent behavior."""
        # First attempt
        try:
            await client.connect()
            pytest.fail("Expected AuthenticationError to be raised")
        except AuthenticationError as e1:
            error1 = str(e1)
        
        # Reset client state
        client.connected = False
        client.authenticated = False
        client.websocket = None
        
        # Second attempt
        try:
            await client.connect()
            pytest.fail("Expected AuthenticationError to be raised")
        except AuthenticationError as e2:
            error2 = str(e2)
        
        # Both attempts should fail with authentication errors
        assert "Authentication" in error1 or "auth" in error1.lower()
        assert "Authentication" in error2 or "auth" in error2.lower()
    
    @pytest.mark.asyncio
    async def test_authentication_error_handling_performance(self, client):
        """Test that authentication error handling is reasonably fast."""
        import time
        
        start_time = time.time()
        
        try:
            await client.connect()
            pytest.fail("Expected AuthenticationError to be raised")
        except AuthenticationError:
            pass
        
        end_time = time.time()
        duration = end_time - start_time
        
        # Authentication failure should be reasonably fast (< 5 seconds)
        assert duration < 5.0, f"Authentication error handling took too long: {duration:.2f} seconds"
    
    @pytest.mark.asyncio
    async def test_authentication_error_with_different_ports(self, client_config):
        """Test authentication error handling with different server ports."""
        # Test with a non-existent port to ensure proper error handling
        invalid_port_config = client_config.copy()
        invalid_port_config.update({
            "port": 9999  # Non-existent port
        })
        
        client = CameraClient(**invalid_port_config)
        
        try:
            # Should raise ConnectionError, not AuthenticationError
            with pytest.raises(ConnectionError):
                await client.connect()
            
            assert not client.connected
            assert not client.authenticated
        finally:
            if hasattr(client, 'websocket') and client.websocket:
                await client.disconnect()
    
    @pytest.mark.asyncio
    async def test_authentication_error_with_different_hosts(self, client_config):
        """Test authentication error handling with different server hosts."""
        # Test with a non-existent host to ensure proper error handling
        invalid_host_config = client_config.copy()
        invalid_host_config.update({
            "host": "nonexistent.host.local"
        })
        
        client = CameraClient(**invalid_host_config)
        
        try:
            # Should raise ConnectionError, not AuthenticationError
            with pytest.raises(ConnectionError):
                await client.connect()
            
            assert not client.connected
            assert not client.authenticated
        finally:
            if hasattr(client, 'websocket') and client.websocket:
                await client.disconnect()
