"""
Unit tests for security middleware.

Tests WebSocket security middleware including authentication,
rate limiting, and connection control as specified in Architecture Decision AD-7.
"""

import pytest
import asyncio
import tempfile
import os
import time
from unittest.mock import Mock, patch, AsyncMock
from datetime import datetime, timedelta, timezone

from src.security.middleware import SecurityMiddleware, RateLimitInfo
from src.security.auth_manager import AuthManager, AuthResult
from src.security.jwt_handler import JWTHandler
from src.security.api_key_handler import APIKeyHandler


class TestRateLimitInfo:
    """Test rate limit information structure."""
    
    def test_create_rate_limit_info(self):
        """Test creating rate limit info."""
        now = time.time()
        info = RateLimitInfo(
            request_count=5,
            window_start=now,
            last_request=now
        )
        
        assert info.request_count == 5
        assert info.window_start == now
        assert info.last_request == now


class TestSecurityMiddleware:
    """Test security middleware functionality."""
    
    @pytest.fixture
    def jwt_handler(self):
        """Create JWT handler for testing."""
        return JWTHandler("test_secret_key")
    
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
    def api_key_handler(self, temp_storage_file):
        """Create API key handler for testing."""
        return APIKeyHandler(temp_storage_file)
    
    @pytest.fixture
    def auth_manager(self, jwt_handler, api_key_handler):
        """Create authentication manager for testing."""
        return AuthManager(jwt_handler, api_key_handler)
    
    @pytest.fixture
    def security_middleware(self, auth_manager):
        """Create security middleware for testing."""
        return SecurityMiddleware(
            auth_manager=auth_manager,
            max_connections=10,
            requests_per_minute=60,
            window_size_seconds=60
        )
    
    def test_init(self, security_middleware, auth_manager):
        """Test security middleware initialization."""
        assert security_middleware.auth_manager == auth_manager
        assert security_middleware.max_connections == 10
        assert security_middleware.requests_per_minute == 60
        assert security_middleware.window_size_seconds == 60
        assert len(security_middleware.active_connections) == 0
        assert len(security_middleware.connection_auth) == 0
        assert len(security_middleware.rate_limit_info) == 0
    
    def test_can_accept_connection_success(self, security_middleware):
        """Test successful connection acceptance."""
        result = security_middleware.can_accept_connection("client_1")
        assert result is True
    
    def test_can_accept_connection_limit_exceeded(self, security_middleware):
        """Test connection rejection when limit exceeded."""
        # Fill up connections
        for i in range(10):
            security_middleware.register_connection(f"client_{i}")
        
        # Try to add one more
        result = security_middleware.can_accept_connection("client_11")
        assert result is False
    
    def test_register_connection(self, security_middleware):
        """Test connection registration."""
        security_middleware.register_connection("client_1")
        
        assert "client_1" in security_middleware.active_connections
        assert len(security_middleware.active_connections) == 1
    
    def test_unregister_connection(self, security_middleware):
        """Test connection unregistration."""
        security_middleware.register_connection("client_1")
        security_middleware.register_connection("client_2")
        
        security_middleware.unregister_connection("client_1")
        
        assert "client_1" not in security_middleware.active_connections
        assert "client_2" in security_middleware.active_connections
        assert len(security_middleware.active_connections) == 1
    
    def test_unregister_connection_not_registered(self, security_middleware):
        """Test unregistering non-existent connection."""
        # Should not raise an error
        security_middleware.unregister_connection("nonexistent_client")
    
    @pytest.mark.asyncio
    async def test_authenticate_connection_jwt_success(self, security_middleware):
        """Test successful JWT authentication."""
        # Generate a valid JWT token
        token = security_middleware.auth_manager.jwt_handler.generate_token("test_user", "viewer")
        
        result = await security_middleware.authenticate_connection("client_1", token, "jwt")
        
        assert result.authenticated is True
        assert result.user_id == "test_user"
        assert result.role == "viewer"
        assert result.auth_method == "jwt"
        assert "client_1" in security_middleware.connection_auth
    
    @pytest.mark.asyncio
    async def test_authenticate_connection_api_key_success(self, security_middleware):
        """Test successful API key authentication."""
        # Create a valid API key
        key = security_middleware.auth_manager.api_key_handler.create_api_key("Test Key", "operator", 1)
        
        result = await security_middleware.authenticate_connection("client_2", key, "api_key")
        
        assert result.authenticated is True
        assert result.role == "operator"
        assert result.auth_method == "api_key"
        assert "client_2" in security_middleware.connection_auth
    
    @pytest.mark.asyncio
    async def test_authenticate_connection_failure(self, security_middleware):
        """Test failed authentication."""
        result = await security_middleware.authenticate_connection("client_3", "invalid_token", "jwt")
        
        assert result.authenticated is False
        assert result.error_message is not None
        assert "client_3" not in security_middleware.connection_auth
    
    def test_is_authenticated_true(self, security_middleware):
        """Test checking authenticated client."""
        security_middleware.connection_auth["client_1"] = Mock()
        security_middleware.connection_auth["client_2"] = Mock()
        
        assert security_middleware.is_authenticated("client_1") is True
        assert security_middleware.is_authenticated("client_2") is True
    
    def test_is_authenticated_false(self, security_middleware):
        """Test checking non-authenticated client."""
        assert security_middleware.is_authenticated("client_3") is False
    
    def test_get_auth_result(self, security_middleware):
        """Test getting authentication result."""
        mock_result = Mock()
        security_middleware.connection_auth["client_1"] = mock_result
        
        result = security_middleware.get_auth_result("client_1")
        assert result == mock_result
    
    def test_get_auth_result_none(self, security_middleware):
        """Test getting authentication result for non-authenticated client."""
        result = security_middleware.get_auth_result("client_1")
        assert result is None
    
    def test_has_permission_true(self, security_middleware):
        """Test permission checking for authenticated client."""
        # Create a mock auth result with admin role
        mock_result = Mock()
        mock_result.role = "admin"
        security_middleware.connection_auth["client_1"] = mock_result
        
        assert security_middleware.has_permission("client_1", "viewer") is True
        assert security_middleware.has_permission("client_1", "operator") is True
        assert security_middleware.has_permission("client_1", "admin") is True
    
    def test_has_permission_false(self, security_middleware):
        """Test permission checking for insufficient role."""
        # Create a mock auth result with viewer role
        mock_result = Mock()
        mock_result.role = "viewer"
        security_middleware.connection_auth["client_1"] = mock_result
        
        assert security_middleware.has_permission("client_1", "viewer") is True
        assert security_middleware.has_permission("client_1", "operator") is False
        assert security_middleware.has_permission("client_1", "admin") is False
    
    def test_has_permission_not_authenticated(self, security_middleware):
        """Test permission checking for non-authenticated client."""
        assert security_middleware.has_permission("client_1", "viewer") is False
    
    def test_check_rate_limit_new_client(self, security_middleware):
        """Test rate limiting for new client."""
        result = security_middleware.check_rate_limit("client_1")
        assert result is True
        
        # Should create rate limit entry
        assert "client_1" in security_middleware.rate_limit_info
        assert security_middleware.rate_limit_info["client_1"].request_count == 1
    
    def test_check_rate_limit_within_limit(self, security_middleware):
        """Test rate limiting within allowed limit."""
        # Make several requests
        for i in range(30):
            result = security_middleware.check_rate_limit("client_1")
            assert result is True
        
        # Should still be within limit (60 requests per minute)
        assert security_middleware.rate_limit_info["client_1"].request_count == 30
    
    def test_check_rate_limit_exceeded(self, security_middleware):
        """Test rate limiting when limit exceeded."""
        # Make more requests than allowed
        for i in range(60):
            result = security_middleware.check_rate_limit("client_1")
            assert result is True
        
        # Next request should be blocked
        result = security_middleware.check_rate_limit("client_1")
        assert result is False
    
    def test_check_rate_limit_window_reset(self, security_middleware):
        """Test rate limiting with window reset."""
        # Make some requests
        for i in range(10):
            security_middleware.check_rate_limit("client_1")
        
        # Manually advance the window start time
        rate_limit = security_middleware.rate_limit_info["client_1"]
        rate_limit.window_start = time.time() - 120  # 2 minutes ago
        
        # Should reset and allow new requests
        result = security_middleware.check_rate_limit("client_1")
        assert result is True
        assert rate_limit.request_count == 1  # Reset to 1
    
    def test_get_connection_stats(self, security_middleware):
        """Test getting connection statistics."""
        # Add some connections and authentication
        security_middleware.register_connection("client_1")
        security_middleware.register_connection("client_2")
        security_middleware.connection_auth["client_1"] = Mock()
        
        stats = security_middleware.get_connection_stats()
        
        assert stats["active_connections"] == 2
        assert stats["authenticated_connections"] == 1
        assert stats["rate_limited_clients"] == 0
        assert "max_connections" in stats
    
    def test_cleanup_expired_rate_limits(self, security_middleware):
        """Test cleanup of expired rate limits."""
        # Create some rate limits
        security_middleware.check_rate_limit("client_1")
        security_middleware.check_rate_limit("client_2")
        
        # Manually expire one rate limit
        rate_limit = security_middleware.rate_limit_info["client_1"]
        rate_limit.last_request = time.time() - 3600  # 1 hour ago
        
        # Cleanup expired rate limits
        removed_count = security_middleware.cleanup_expired_rate_limits()
        assert removed_count == 1
        
        # Verify expired rate limit is removed
        assert "client_1" not in security_middleware.rate_limit_info
        assert "client_2" in security_middleware.rate_limit_info
    
    @pytest.mark.asyncio
    async def test_authenticate_and_check_permission_success(self, security_middleware):
        """Test combined authentication and permission check."""
        # Generate a valid JWT token with admin role
        token = security_middleware.auth_manager.jwt_handler.generate_token("test_user", "admin")
        
        result = await security_middleware.authenticate_and_check_permission(
            "client_1", token, "operator"
        )
        
        assert result.authenticated is True
        assert result.role == "admin"
        assert security_middleware.has_permission("client_1", "operator") is True
    
    @pytest.mark.asyncio
    async def test_authenticate_and_check_permission_insufficient(self, security_middleware):
        """Test combined authentication and permission check with insufficient role."""
        # Generate a valid JWT token with viewer role
        token = security_middleware.auth_manager.jwt_handler.generate_token("test_user", "viewer")
        
        result = await security_middleware.authenticate_and_check_permission(
            "client_1", token, "admin"
        )
        
        assert result.authenticated is False  # Should be False due to insufficient permissions
        assert result.role == "viewer"
        assert result.error_message is not None
        assert "Insufficient permissions" in result.error_message
    
    @pytest.mark.asyncio
    async def test_authenticate_and_check_permission_auth_failure(self, security_middleware):
        """Test combined authentication and permission check with auth failure."""
        result = await security_middleware.authenticate_and_check_permission(
            "client_1", "invalid_token", "viewer"
        )
        
        assert result.authenticated is False
        assert result.error_message is not None


class TestSecurityMiddlewareIntegration:
    """Integration tests for security middleware."""
    
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
        """Create security middleware for integration tests."""
        jwt_handler = JWTHandler("integration_test_secret")
        api_key_handler = APIKeyHandler(temp_storage_file)
        auth_manager = AuthManager(jwt_handler, api_key_handler)
        
        return SecurityMiddleware(
            auth_manager=auth_manager,
            max_connections=5,
            requests_per_minute=30,
            window_size_seconds=60
        )
    
    @pytest.mark.asyncio
    async def test_full_connection_lifecycle(self, security_middleware):
        """Test complete connection lifecycle with authentication."""
        # Test connection acceptance
        assert security_middleware.can_accept_connection("client_1") is True
        
        # Register connection
        security_middleware.register_connection("client_1")
        assert "client_1" in security_middleware.active_connections
        
        # Authenticate connection
        token = security_middleware.auth_manager.jwt_handler.generate_token("test_user", "operator")
        auth_result = await security_middleware.authenticate_connection("client_1", token, "jwt")
        
        assert auth_result.authenticated is True
        assert security_middleware.is_authenticated("client_1") is True
        
        # Test rate limiting
        for i in range(10):
            assert security_middleware.check_rate_limit("client_1") is True
        
        # Test permission checking
        assert security_middleware.has_permission("client_1", "viewer") is True
        assert security_middleware.has_permission("client_1", "operator") is True
        assert security_middleware.has_permission("client_1", "admin") is False
        
        # Unregister connection
        security_middleware.unregister_connection("client_1")
        assert "client_1" not in security_middleware.active_connections
    
    @pytest.mark.asyncio
    async def test_multiple_clients(self, security_middleware):
        """Test handling multiple clients."""
        # Add multiple clients
        clients = ["client_1", "client_2", "client_3"]
        
        for client in clients:
            security_middleware.register_connection(client)
            token = security_middleware.auth_manager.jwt_handler.generate_token(f"user_{client}", "viewer")
            await security_middleware.authenticate_connection(client, token, "jwt")
        
        # Verify all clients are authenticated
        for client in clients:
            assert security_middleware.is_authenticated(client) is True
        
        # Test rate limiting for each client
        for client in clients:
            for i in range(5):
                assert security_middleware.check_rate_limit(client) is True
        
        # Verify stats
        stats = security_middleware.get_connection_stats()
        assert stats["active_connections"] == 3
        assert stats["authenticated_connections"] == 3
    
    @pytest.mark.asyncio
    async def test_connection_limit_enforcement(self, security_middleware):
        """Test connection limit enforcement."""
        # Fill up to the limit
        for i in range(5):
            client = f"client_{i}"
            security_middleware.register_connection(client)
            token = security_middleware.auth_manager.jwt_handler.generate_token(f"user_{i}", "viewer")
            await security_middleware.authenticate_connection(client, token, "jwt")
        
        # Try to add one more connection
        assert security_middleware.can_accept_connection("client_6") is False
        
        # Remove one connection
        security_middleware.unregister_connection("client_0")
        
        # Should be able to add new connection
        assert security_middleware.can_accept_connection("client_6") is True
    
    @pytest.mark.asyncio
    async def test_rate_limit_enforcement(self, security_middleware):
        """Test rate limit enforcement."""
        # Register and authenticate a client
        security_middleware.register_connection("client_1")
        token = security_middleware.auth_manager.jwt_handler.generate_token("test_user", "viewer")
        await security_middleware.authenticate_connection("client_1", token, "jwt")
        
        # Make requests up to the limit
        for i in range(30):
            assert security_middleware.check_rate_limit("client_1") is True
        
        # Next request should be blocked
        assert security_middleware.check_rate_limit("client_1") is False
        
        # Verify stats show rate limited client
        stats = security_middleware.get_connection_stats()
        assert stats["rate_limited_clients"] == 1 