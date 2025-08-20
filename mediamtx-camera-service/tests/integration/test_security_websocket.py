"""
Integration tests for WebSocket security validation.

Requirements Traceability:
- REQ-SEC-001: Security system shall provide WebSocket authentication validation
- REQ-SEC-004: Security system shall support permission checking for WebSocket operations
- REQ-SEC-001: Security system shall handle rate limiting and connection control

Story Coverage: S7 - Security Implementation
IV&V Control Point: Real WebSocket security validation



Tests WebSocket authentication, permission checking, rate limiting,
and connection control as specified in Sprint 2 Task S7.1.
"""

import pytest
import asyncio
import tempfile
import os
import time

from src.security.middleware import SecurityMiddleware
from src.security.auth_manager import AuthManager
from src.security.jwt_handler import JWTHandler
from src.security.api_key_handler import APIKeyHandler
from tests.fixtures.auth_utils import get_test_jwt_secret


class TestWebSocketAuthentication:
    """Integration tests for WebSocket authentication."""
    
    @pytest.fixture
    def temp_storage_file(self):
        """Create temporary storage file for API key handler."""
        with tempfile.NamedTemporaryFile(mode='w', delete=False, suffix='.json') as f:
            f.write('{"version": "1.0", "keys": []}')
            temp_file = f.name
        
        yield temp_file
        
        # Cleanup
        if os.path.exists(temp_file):
            os.unlink(temp_file)
    
    @pytest.fixture
    def security_middleware(self, temp_storage_file):
        """Create security middleware for WebSocket testing."""
        jwt_handler = JWTHandler(get_test_jwt_secret())
        api_key_handler = APIKeyHandler(temp_storage_file)
        auth_manager = AuthManager(jwt_handler, api_key_handler)
        
        return SecurityMiddleware(
            auth_manager=auth_manager,
            max_connections=10,
            requests_per_minute=60,
            window_size_seconds=60
        )
    
    @pytest.mark.asyncio
    async def test_websocket_authentication_before_method_execution(self, security_middleware):
        """Test authentication check before method execution."""
        # Register connection
        client_id = "test_client_1"
        security_middleware.register_connection(client_id)
        
        # Generate JWT token
        token = security_middleware.auth_manager.jwt_handler.generate_token("websocket_user", "operator")
        
        # Authenticate connection
        auth_result = await security_middleware.authenticate_connection(client_id, token, "jwt")
        
        # Verify authentication success
        assert auth_result.authenticated is True
        assert auth_result.user_id == "websocket_user"
        assert auth_result.role == "operator"
        assert auth_result.auth_method == "jwt"
        
        # Verify client is authenticated
        assert security_middleware.is_authenticated(client_id) is True
        
        # Cleanup
        security_middleware.unregister_connection(client_id)
    
    @pytest.mark.asyncio
    async def test_websocket_permission_checking_for_sensitive_operations(self, security_middleware):
        """Test permission checking for sensitive WebSocket operations."""
        # Register and authenticate client with viewer role
        client_id = "viewer_client"
        security_middleware.register_connection(client_id)
        
        viewer_token = security_middleware.auth_manager.jwt_handler.generate_token("viewer_user", "viewer")
        await security_middleware.authenticate_connection(client_id, viewer_token, "jwt")
        
        # Test permission checking for different operations
        assert security_middleware.has_permission(client_id, "viewer") is True
        assert security_middleware.has_permission(client_id, "operator") is False
        assert security_middleware.has_permission(client_id, "admin") is False
        
        # Register and authenticate client with admin role
        admin_client_id = "admin_client"
        security_middleware.register_connection(admin_client_id)
        
        admin_token = security_middleware.auth_manager.jwt_handler.generate_token("admin_user", "admin")
        await security_middleware.authenticate_connection(admin_client_id, admin_token, "jwt")
        
        # Admin should have all permissions
        assert security_middleware.has_permission(admin_client_id, "viewer") is True
        assert security_middleware.has_permission(admin_client_id, "operator") is True
        assert security_middleware.has_permission(admin_client_id, "admin") is True
        
        # Cleanup
        security_middleware.unregister_connection(client_id)
        security_middleware.unregister_connection(admin_client_id)
    
    @pytest.mark.asyncio
    async def test_websocket_rate_limiting_enforcement(self, security_middleware):
        """Test rate limiting enforcement for WebSocket connections."""
        # Register connection
        client_id = "rate_limit_client"
        security_middleware.register_connection(client_id)
        
        # Authenticate client
        token = security_middleware.auth_manager.jwt_handler.generate_token("rate_user", "viewer")
        await security_middleware.authenticate_connection(client_id, token, "jwt")
        
        # Test rate limiting
        for i in range(30):
            assert security_middleware.check_rate_limit(client_id) is True
        
        # Should still be within limit
        assert security_middleware.rate_limit_info[client_id].request_count == 30
        
        # Test rate limit exceeded
        for i in range(30):
            assert security_middleware.check_rate_limit(client_id) is True
        
        # Next request should be blocked
        assert security_middleware.check_rate_limit(client_id) is False
        
        # Cleanup
        security_middleware.unregister_connection(client_id)
    
    @pytest.mark.asyncio
    async def test_websocket_connection_limits_and_cleanup(self, security_middleware):
        """Test connection limits and cleanup for WebSocket connections."""
        # Test connection limit enforcement
        for i in range(10):
            client_id = f"client_{i}"
            assert security_middleware.can_accept_connection(client_id) is True
            security_middleware.register_connection(client_id)
        
        # Try to add one more connection (should be rejected)
        assert security_middleware.can_accept_connection("client_11") is False
        
        # Remove some connections
        security_middleware.unregister_connection("client_0")
        security_middleware.unregister_connection("client_1")
        
        # Should be able to add new connections
        assert security_middleware.can_accept_connection("client_12") is True
        security_middleware.register_connection("client_12")
        
        # Cleanup remaining connections
        for i in range(2, 10):
            security_middleware.unregister_connection(f"client_{i}")
        security_middleware.unregister_connection("client_12")
    
    @pytest.mark.asyncio
    async def test_websocket_error_response_validation(self, security_middleware):
        """Test error response validation for WebSocket security."""
        # Test authentication failure
        client_id = "error_client"
        security_middleware.register_connection(client_id)
        
        # Try to authenticate with invalid token
        auth_result = await security_middleware.authenticate_connection(client_id, "invalid_token", "jwt")
        
        # Verify authentication failure
        assert auth_result.authenticated is False
        assert auth_result.error_message is not None
        assert client_id not in security_middleware.connection_auth
        
        # Test permission failure
        valid_token = security_middleware.auth_manager.jwt_handler.generate_token("error_user", "viewer")
        await security_middleware.authenticate_connection(client_id, valid_token, "jwt")
        
        # Try to access admin-only operation
        assert security_middleware.has_permission(client_id, "admin") is False
        
        # Cleanup
        security_middleware.unregister_connection(client_id)


class TestWebSocketSecurityIntegration:
    """Integration tests for WebSocket security with other components."""
    
    @pytest.fixture
    def temp_storage_file(self):
        """Create temporary storage file for integration tests."""
        with tempfile.NamedTemporaryFile(mode='w', delete=False, suffix='.json') as f:
            f.write('{"version": "1.0", "keys": []}')
            temp_file = f.name
        
        yield temp_file
        
        # Cleanup
        if os.path.exists(temp_file):
            os.unlink(temp_file)
    
    @pytest.fixture
    def security_middleware(self, temp_storage_file):
        """Create security middleware for integration testing."""
        jwt_handler = JWTHandler(get_test_jwt_secret())
        api_key_handler = APIKeyHandler(temp_storage_file)
        auth_manager = AuthManager(jwt_handler, api_key_handler)
        
        return SecurityMiddleware(
            auth_manager=auth_manager,
            max_connections=5,
            requests_per_minute=30,
            window_size_seconds=60
        )
    
    @pytest.mark.asyncio
    async def test_websocket_authentication_with_api_keys(self, security_middleware):
        """Test WebSocket authentication using API keys."""
        # Register connection
        client_id = "api_key_client"
        security_middleware.register_connection(client_id)
        
        # Create API key
        api_key = security_middleware.auth_manager.api_key_handler.create_api_key("WebSocket API Key", "operator", 1)
        
        # Authenticate with API key
        auth_result = await security_middleware.authenticate_connection(client_id, api_key, "api_key")
        
        # Verify authentication success
        assert auth_result.authenticated is True
        assert auth_result.role == "operator"
        assert auth_result.auth_method == "api_key"
        
        # Test permission checking
        assert security_middleware.has_permission(client_id, "viewer") is True
        assert security_middleware.has_permission(client_id, "operator") is True
        assert security_middleware.has_permission(client_id, "admin") is False
        
        # Cleanup
        security_middleware.unregister_connection(client_id)
    
    @pytest.mark.asyncio
    async def test_websocket_auto_authentication_fallback(self, security_middleware):
        """Test WebSocket auto authentication with fallback."""
        # Register connection
        client_id = "auto_client"
        security_middleware.register_connection(client_id)
        
        # Test with JWT token
        jwt_token = security_middleware.auth_manager.jwt_handler.generate_token("auto_user", "viewer")
        jwt_result = await security_middleware.authenticate_connection(client_id, jwt_token, "auto")
        
        assert jwt_result.authenticated is True
        assert jwt_result.auth_method == "jwt"
        
        # Cleanup and test with API key
        security_middleware.unregister_connection(client_id)
        security_middleware.register_connection(client_id)
        
        api_key = security_middleware.auth_manager.api_key_handler.create_api_key("Auto API Key", "admin", 1)
        api_result = await security_middleware.authenticate_connection(client_id, api_key, "auto")
        
        assert api_result.authenticated is True
        assert api_result.auth_method == "api_key"
        
        # Cleanup
        security_middleware.unregister_connection(client_id)
    
    @pytest.mark.asyncio
    async def test_websocket_concurrent_authentication_attempts(self, security_middleware):
        """Test concurrent authentication attempts for WebSocket connections."""
        # Register multiple connections
        clients = []
        for i in range(5):
            client_id = f"concurrent_client_{i}"
            security_middleware.register_connection(client_id)
            clients.append(client_id)
        
        # Authenticate all clients concurrently
        auth_tasks = []
        for i, client_id in enumerate(clients):
            if i % 2 == 0:
                # Use JWT for even clients
                token = security_middleware.auth_manager.jwt_handler.generate_token(f"jwt_user_{i}", "viewer")
                task = security_middleware.authenticate_connection(client_id, token, "jwt")
            else:
                # Use API key for odd clients
                api_key = security_middleware.auth_manager.api_key_handler.create_api_key(f"API Key {i}", "operator", 1)
                task = security_middleware.authenticate_connection(client_id, api_key, "api_key")
            auth_tasks.append(task)
        
        # Wait for all authentications to complete
        results = await asyncio.gather(*auth_tasks)
        
        # Verify all authentications succeeded
        for result in results:
            assert result.authenticated is True
        
        # Verify all clients are authenticated
        for client_id in clients:
            assert security_middleware.is_authenticated(client_id) is True
        
        # Cleanup
        for client_id in clients:
            security_middleware.unregister_connection(client_id)
    
    @pytest.mark.asyncio
    async def test_websocket_authentication_and_permission_combined(self, security_middleware):
        """Test combined authentication and permission checking."""
        # Register connection
        client_id = "combined_client"
        security_middleware.register_connection(client_id)
        
        # Generate token with admin role
        token = security_middleware.auth_manager.jwt_handler.generate_token("combined_user", "admin")
        
        # Test combined authentication and permission check
        result = await security_middleware.authenticate_and_check_permission(client_id, token, "operator")
        
        # Verify authentication and permission success
        assert result.authenticated is True
        assert result.role == "admin"
        assert security_middleware.has_permission(client_id, "operator") is True
        
        # Test insufficient permissions
        viewer_token = security_middleware.auth_manager.jwt_handler.generate_token("viewer_user", "viewer")
        security_middleware.unregister_connection(client_id)
        security_middleware.register_connection(client_id)
        
        insufficient_result = await security_middleware.authenticate_and_check_permission(client_id, viewer_token, "admin")
        
        # Should fail due to insufficient permissions
        assert insufficient_result.authenticated is False
        assert "Insufficient permissions" in insufficient_result.error_message
        
        # Cleanup
        security_middleware.unregister_connection(client_id)


class TestWebSocketSecurityPerformance:
    """Performance tests for WebSocket security."""
    
    @pytest.fixture
    def temp_storage_file(self):
        """Create temporary storage file for performance tests."""
        with tempfile.NamedTemporaryFile(mode='w', delete=False, suffix='.json') as f:
            f.write('{"version": "1.0", "keys": []}')
            temp_file = f.name
        
        yield temp_file
        
        # Cleanup
        if os.path.exists(temp_file):
            os.unlink(temp_file)
    
    @pytest.fixture
    def security_middleware(self, temp_storage_file):
        """Create security middleware for performance testing."""
        jwt_handler = JWTHandler(get_test_jwt_secret())
        api_key_handler = APIKeyHandler(temp_storage_file)
        auth_manager = AuthManager(jwt_handler, api_key_handler)
        
        return SecurityMiddleware(
            auth_manager=auth_manager,
            max_connections=100,
            requests_per_minute=1000,
            window_size_seconds=60
        )
    
    @pytest.mark.asyncio
    async def test_websocket_authentication_performance(self, security_middleware):
        """Test WebSocket authentication performance."""
        # Register connection
        client_id = "perf_client"
        security_middleware.register_connection(client_id)
        
        # Generate token
        token = security_middleware.auth_manager.jwt_handler.generate_token("perf_user", "admin")
        
        # Measure authentication performance
        start_time = time.time()
        for _ in range(100):
            result = await security_middleware.authenticate_connection(client_id, token, "jwt")
            assert result.authenticated is True
        
        end_time = time.time()
        avg_time = (end_time - start_time) / 100
        
        # Performance should be under 1ms per authentication
        assert avg_time < 0.001, f"WebSocket authentication too slow: {avg_time:.6f}s per request"
        
        # Cleanup
        security_middleware.unregister_connection(client_id)
    
    @pytest.mark.asyncio
    async def test_websocket_rate_limiting_performance(self, security_middleware):
        """Test WebSocket rate limiting performance."""
        # Register connection
        client_id = "rate_perf_client"
        security_middleware.register_connection(client_id)
        
        # Measure rate limiting performance
        start_time = time.time()
        for _ in range(1000):
            security_middleware.check_rate_limit(client_id)
        
        end_time = time.time()
        avg_time = (end_time - start_time) / 1000
        
        # Performance should be under 0.1ms per rate limit check
        assert avg_time < 0.0001, f"Rate limiting too slow: {avg_time:.6f}s per check"
        
        # Cleanup
        security_middleware.unregister_connection(client_id)
    
    @pytest.mark.asyncio
    async def test_websocket_connection_management_performance(self, security_middleware):
        """Test WebSocket connection management performance."""
        # Measure connection registration performance
        start_time = time.time()
        for i in range(100):
            client_id = f"conn_perf_client_{i}"
            security_middleware.register_connection(client_id)
        
        end_time = time.time()
        avg_registration_time = (end_time - start_time) / 100
        
        # Registration should be under 0.1ms per connection
        assert avg_registration_time < 0.0001, f"Connection registration too slow: {avg_registration_time:.6f}s per connection"
        
        # Measure connection cleanup performance
        start_time = time.time()
        for i in range(100):
            client_id = f"conn_perf_client_{i}"
            security_middleware.unregister_connection(client_id)
        
        end_time = time.time()
        avg_cleanup_time = (end_time - start_time) / 100
        
        # Cleanup should be under 0.1ms per connection
        assert avg_cleanup_time < 0.0001, f"Connection cleanup too slow: {avg_cleanup_time:.6f}s per connection"
    
    @pytest.mark.asyncio
    async def test_websocket_memory_usage_under_load(self, security_middleware):
        """Test WebSocket memory usage under load."""
        # Create many connections and authenticate them
        clients = []
        for i in range(50):
            client_id = f"memory_client_{i}"
            security_middleware.register_connection(client_id)
            
            # Authenticate with JWT
            token = security_middleware.auth_manager.jwt_handler.generate_token(f"memory_user_{i}", "viewer")
            await security_middleware.authenticate_connection(client_id, token, "jwt")
            
            clients.append(client_id)
        
        # Verify all connections are active
        assert len(security_middleware.active_connections) == 50
        assert len(security_middleware.connection_auth) == 50
        
        # Test rate limiting for all clients
        for client_id in clients:
            for _ in range(10):
                security_middleware.check_rate_limit(client_id)
        
        # Verify rate limiting data is maintained
        assert len(security_middleware.rate_limit_info) == 50
        
        # Cleanup all connections
        for client_id in clients:
            security_middleware.unregister_connection(client_id)
        
        # Verify cleanup
        assert len(security_middleware.active_connections) == 0
        assert len(security_middleware.connection_auth) == 0


class TestWebSocketSecurityErrorHandling:
    """Test WebSocket security error handling."""
    
    @pytest.fixture
    def temp_storage_file(self):
        """Create temporary storage file for error handling tests."""
        with tempfile.NamedTemporaryFile(mode='w', delete=False, suffix='.json') as f:
            f.write('{"version": "1.0", "keys": []}')
            temp_file = f.name
        
        yield temp_file
        
        # Cleanup
        if os.path.exists(temp_file):
            os.unlink(temp_file)
    
    @pytest.fixture
    def security_middleware(self, temp_storage_file):
        """Create security middleware for error handling tests."""
        jwt_handler = JWTHandler(get_test_jwt_secret())
        api_key_handler = APIKeyHandler(temp_storage_file)
        auth_manager = AuthManager(jwt_handler, api_key_handler)
        
        return SecurityMiddleware(
            auth_manager=auth_manager,
            max_connections=10,
            requests_per_minute=60,
            window_size_seconds=60
        )
    
    @pytest.mark.asyncio
    async def test_websocket_authentication_with_invalid_tokens(self, security_middleware):
        """Test WebSocket authentication with invalid tokens."""
        # Register connection
        client_id = "error_client"
        security_middleware.register_connection(client_id)
        
        # Test various invalid authentication scenarios
        invalid_scenarios = [
            ("invalid_token", "jwt"),
            ("", "jwt"),
            (None, "jwt"),
            ("invalid_api_key", "api_key"),
            ("", "api_key"),
            (None, "api_key")
        ]
        
        for token, auth_type in invalid_scenarios:
            result = await security_middleware.authenticate_connection(client_id, token, auth_type)
            assert result.authenticated is False
            assert result.error_message is not None
        
        # Cleanup
        security_middleware.unregister_connection(client_id)
    
    @pytest.mark.asyncio
    async def test_websocket_connection_limits_error_handling(self, security_middleware):
        """Test WebSocket connection limits error handling."""
        # Fill up connections
        for i in range(10):
            client_id = f"limit_client_{i}"
            security_middleware.register_connection(client_id)
        
        # Try to add more connections (should be rejected)
        for i in range(5):
            client_id = f"rejected_client_{i}"
            assert security_middleware.can_accept_connection(client_id) is False
        
        # Remove some connections
        for i in range(5):
            client_id = f"limit_client_{i}"
            security_middleware.unregister_connection(client_id)
        
        # Should be able to add new connections
        for i in range(5):
            client_id = f"new_client_{i}"
            assert security_middleware.can_accept_connection(client_id) is True
            security_middleware.register_connection(client_id)
        
        # Cleanup
        for i in range(5, 10):
            security_middleware.unregister_connection(f"limit_client_{i}")
        for i in range(5):
            security_middleware.unregister_connection(f"new_client_{i}")
    
    @pytest.mark.asyncio
    async def test_websocket_rate_limiting_error_handling(self, security_middleware):
        """Test WebSocket rate limiting error handling."""
        # Register connection
        client_id = "rate_error_client"
        security_middleware.register_connection(client_id)
        
        # Exceed rate limit
        for _ in range(60):
            security_middleware.check_rate_limit(client_id)
        
        # Next requests should be blocked
        for _ in range(10):
            assert security_middleware.check_rate_limit(client_id) is False
        
        # Cleanup
        security_middleware.unregister_connection(client_id) 