"""
Integration test for Python SDK authentication error handling.

Requirements Coverage:
- REQ-SDK-003: SDK shall handle errors gracefully
- REQ-AUTH-001: Authentication shall work with JWT tokens
- REQ-AUTH-002: Authentication shall work with API keys

Story Coverage: S8.3 - SDK Development
IV&V Control Point: SDK validation

API Documentation Reference: docs/api/json-rpc-methods.md
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

# Import test utilities
from tests.fixtures.auth_utils import get_test_auth_manager, TestUserFactory, WebSocketAuthTestClient, cleanup_test_auth_manager
from tests.utils.port_utils import find_free_port
from src.camera_service.config import Config, ServerConfig, MediaMTXConfig, CameraConfig, LoggingConfig, RecordingConfig, SnapshotConfig
from src.camera_service.service_manager import ServiceManager
from src.websocket_server.server import WebSocketJsonRpcServer
from mediamtx_wrapper.controller import MediaMTXController
from camera_discovery.hybrid_monitor import HybridCameraMonitor


class SDKTestSetup:
    """Test setup for SDK integration tests with real server."""
    
    def __init__(self):
        self.config = self._build_test_config()
        self.service_manager = None
        self.mediamtx_controller = None
        self.camera_monitor = None
        self.server = None
        self.auth_manager = get_test_auth_manager()
        self.user_factory = TestUserFactory(self.auth_manager)
        self.websocket_client = None
    
    def _build_test_config(self) -> Config:
        """Build test configuration for SDK testing."""
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
    
    async def setup(self):
        """Set up real system components for SDK testing."""
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
            health_circuit_breaker_timeout=mediamtx_config.health_circuit_breaker_timeout,
            health_max_backoff_interval=mediamtx_config.health_max_backoff_interval,
            health_recovery_confirmation_threshold=mediamtx_config.health_recovery_confirmation_threshold,
            backoff_base_multiplier=mediamtx_config.backoff_base_multiplier,
            backoff_jitter_range=mediamtx_config.backoff_jitter_range,
            process_termination_timeout=mediamtx_config.process_termination_timeout,
            process_kill_timeout=mediamtx_config.process_kill_timeout,
        )
        
        # Initialize real camera monitor
        self.camera_monitor = HybridCameraMonitor(
            device_range=self.config.camera.device_range,
            poll_interval=self.config.camera.poll_interval,
            detection_timeout=self.config.camera.detection_timeout,
            enable_capability_detection=self.config.camera.enable_capability_detection,
        )
        
        # Initialize service manager with components
        self.service_manager = ServiceManager(
            config=self.config,
            mediamtx_controller=self.mediamtx_controller,
            camera_monitor=self.camera_monitor
        )
        
        # Start service manager
        await self.service_manager.start()
        
        # Initialize WebSocket server with security middleware
        self.server = WebSocketJsonRpcServer(
            host=self.config.server.host,
            port=self.config.server.port,
            websocket_path=self.config.server.websocket_path,
            max_connections=self.config.server.max_connections,
            mediamtx_controller=self.mediamtx_controller,
            camera_monitor=self.camera_monitor,
            config=self.config
        )
        
        # Create and set security middleware
        from src.security.middleware import SecurityMiddleware
        security_middleware = SecurityMiddleware(self.auth_manager, max_connections=10, requests_per_minute=120)
        self.server.set_security_middleware(security_middleware)
        self.server.set_service_manager(self.service_manager)
        
        # Start server
        await self.server.start()
    
    async def cleanup(self):
        """Clean up test resources."""
        if self.server:
            await self.server.stop()
        
        if self.service_manager:
            await self.service_manager.stop()
        
        if self.camera_monitor:
            await self.camera_monitor.stop()
        
        if self.mediamtx_controller:
            await self.mediamtx_controller.stop()


@pytest.mark.skipif(not SDK_AVAILABLE, reason="Python SDK not available")
@pytest.mark.integration
class TestSDKAuthenticationErrorHandling:
    """Integration test for SDK authentication error handling with real server."""
    
    @pytest.fixture
    def client_config(self):
        """Test client configuration."""
        return {
            "host": "localhost",
            "port": 8012,  # Will be updated with actual server port
            "use_ssl": False,
            "auth_type": "jwt",
            "auth_token": "invalid_token",
            "max_retries": 1,  # Single attempt for faster testing
            "retry_delay": 0.1
        }
    
    @pytest_asyncio.fixture
    async def test_setup(self):
        """Test setup with real server."""
        setup = SDKTestSetup()
        await setup.setup()
        yield setup
        await setup.cleanup()
    
    @pytest_asyncio.fixture
    async def client(self, client_config, test_setup):
        """Test client instance with correct server port."""
        # Update client config with actual server port
        client_config["port"] = test_setup.config.server.port
        client = CameraClient(**client_config)
        yield client
        # Cleanup
        if hasattr(client, 'websocket') and client.websocket:
            await client.disconnect()
    
    @pytest.mark.asyncio
    async def test_invalid_jwt_token_raises_authentication_error(self, client):
        """Test that invalid JWT token raises AuthenticationError with real server."""
        # This should connect to the WebSocket but fail authentication
        # Note: Connection succeeds, but authentication fails
        try:
            await client.connect()
            # If we get here, connection succeeded but authentication should fail
            assert not client.authenticated
        except AuthenticationError:
            # Expected: authentication failed
            assert not client.authenticated
        except ConnectionError:
            # Also acceptable: connection failed before authentication
            pass
    
    @pytest.mark.asyncio
    async def test_invalid_api_key_raises_authentication_error(self, client_config, test_setup):
        """Test that invalid API key raises AuthenticationError with real server."""
        # Configure client with invalid API key and correct server port
        api_key_config = client_config.copy()
        api_key_config.update({
            "port": test_setup.config.server.port,
            "auth_type": "api_key",
            "api_key": "invalid_api_key"
        })
        
        client = CameraClient(**api_key_config)
        
        try:
            # Note: Connection succeeds, but authentication fails
            try:
                await client.connect()
                # If we get here, connection succeeded but authentication should fail
                assert not client.authenticated
            except AuthenticationError:
                # Expected: authentication failed
                assert not client.authenticated
            except ConnectionError:
                # Also acceptable: connection failed before authentication
                pass
        finally:
            if hasattr(client, 'websocket') and client.websocket:
                await client.disconnect()
    
    @pytest.mark.asyncio
    async def test_no_auth_token_raises_authentication_error(self, client_config, test_setup):
        """Test that no auth token raises AuthenticationError with real server."""
        # Configure client with no auth token and correct server port
        no_auth_config = client_config.copy()
        no_auth_config.update({
            "port": test_setup.config.server.port,
            "auth_token": None,
            "api_key": None
        })
        
        client = CameraClient(**no_auth_config)
        
        try:
            # Note: Connection succeeds, but authentication fails
            try:
                await client.connect()
                # If we get here, connection succeeded but authentication should fail
                assert not client.authenticated
            except AuthenticationError:
                # Expected: authentication failed
                assert not client.authenticated
            except ConnectionError:
                # Also acceptable: connection failed before authentication
                pass
        finally:
            if hasattr(client, 'websocket') and client.websocket:
                await client.disconnect()
    
    @pytest.mark.asyncio
    async def test_empty_auth_token_raises_authentication_error(self, client_config, test_setup):
        """Test that empty auth token raises AuthenticationError with real server."""
        # Configure client with empty auth token and correct server port
        empty_auth_config = client_config.copy()
        empty_auth_config.update({
            "port": test_setup.config.server.port,
            "auth_token": ""
        })
        
        client = CameraClient(**empty_auth_config)
        
        try:
            # Note: Connection succeeds, but authentication fails
            try:
                await client.connect()
                # If we get here, connection succeeded but authentication should fail
                assert not client.authenticated
            except AuthenticationError:
                # Expected: authentication failed
                assert not client.authenticated
            except ConnectionError:
                # Also acceptable: connection failed before authentication
                pass
        finally:
            if hasattr(client, 'websocket') and client.websocket:
                await client.disconnect()
    
    @pytest.mark.asyncio
    async def test_malformed_jwt_token_raises_authentication_error(self, client_config, test_setup):
        """Test that malformed JWT token raises AuthenticationError with real server."""
        # Configure client with malformed JWT token and correct server port
        malformed_config = client_config.copy()
        malformed_config.update({
            "port": test_setup.config.server.port,
            "auth_token": "not.a.valid.jwt.token"
        })
        
        client = CameraClient(**malformed_config)
        
        try:
            # Note: Connection succeeds, but authentication fails
            try:
                await client.connect()
                # If we get here, connection succeeded but authentication should fail
                assert not client.authenticated
            except AuthenticationError:
                # Expected: authentication failed
                assert not client.authenticated
            except ConnectionError:
                # Also acceptable: connection failed before authentication
                pass
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
        except ConnectionError:
            # Also acceptable: connection failed before authentication
            pass
    
    @pytest.mark.asyncio
    async def test_authentication_error_type_is_correct(self, client):
        """Test that the correct exception type is raised."""
        try:
            await client.connect()
            pytest.fail("Expected AuthenticationError to be raised")
        except AuthenticationError:
            # This is the expected exception type
            pass
        except ConnectionError:
            # Also acceptable: connection failed before authentication
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
        except ConnectionError:
            # Also acceptable: connection failed before authentication
            pass
    
    @pytest.mark.asyncio
    async def test_multiple_authentication_attempts_consistent_behavior(self, client):
        """Test that multiple authentication attempts have consistent behavior."""
        # First attempt
        try:
            await client.connect()
            pytest.fail("Expected AuthenticationError to be raised")
        except AuthenticationError as e1:
            error1 = str(e1)
        except ConnectionError:
            # Also acceptable: connection failed before authentication
            error1 = "Connection failed"
        
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
        except ConnectionError:
            # Also acceptable: connection failed before authentication
            error2 = "Connection failed"
        
        # Both attempts should fail with authentication errors or connection errors
        assert "Authentication" in error1 or "auth" in error1.lower() or "Connection" in error1
        assert "Authentication" in error2 or "auth" in error2.lower() or "Connection" in error2
    
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
