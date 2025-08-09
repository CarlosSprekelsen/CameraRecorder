"""
Unit tests for authentication manager.

Tests authentication coordination between JWT and API key handlers
as specified in Architecture Decision AD-7.
"""

import pytest
import tempfile
import os
from unittest.mock import Mock, patch
from datetime import datetime, timedelta, timezone

from src.security.auth_manager import AuthManager, AuthResult
from src.security.jwt_handler import JWTHandler
from src.security.api_key_handler import APIKeyHandler, APIKey


class TestAuthResult:
    """
    Validates N3.2: Auth result structure including role and expiration
    """
    """Test authentication result structure."""
    
    def test_create_auth_result(self):
        """Test creating authentication result."""
        result = AuthResult(
            authenticated=True,
            user_id="test_user",
            role="viewer",
            auth_method="jwt",
            error_message=None
        )
        
        assert result.authenticated is True
        assert result.user_id == "test_user"
        assert result.role == "viewer"
        assert result.auth_method == "jwt"
        assert result.error_message is None
    
    def test_create_failed_auth_result(self):
        """Test creating failed authentication result."""
        result = AuthResult(
            authenticated=False,
            user_id=None,
            role=None,
            auth_method="api_key",
            error_message="Invalid API key"
        )
        
        assert result.authenticated is False
        assert result.user_id is None
        assert result.role is None
        assert result.auth_method == "api_key"
        assert result.error_message == "Invalid API key"


class TestAuthManager:
    """
    Validates N3.2: Auth manager behavior (JWT/API key) and token lifecycle
    """
    """Test authentication manager functionality."""
    
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
    
    def test_init(self, auth_manager, jwt_handler, api_key_handler):
        """Test authentication manager initialization."""
        assert auth_manager.jwt_handler == jwt_handler
        assert auth_manager.api_key_handler == api_key_handler
    
    def test_authenticate_jwt_success(self, auth_manager):
        """Test successful JWT authentication."""
        # Generate a valid JWT token
        token = auth_manager.jwt_handler.generate_token("test_user", "viewer")
        
        result = auth_manager.authenticate(token, "jwt")
        
        assert result.authenticated is True
        assert result.user_id == "test_user"
        assert result.role == "viewer"
        assert result.auth_method == "jwt"
        assert result.error_message is None
    
    def test_authenticate_jwt_invalid_token(self, auth_manager):
        """Test JWT authentication with invalid token."""
        result = auth_manager.authenticate("invalid_token", "jwt")
        
        assert result.authenticated is False
        assert result.user_id is None
        assert result.role is None
        assert result.auth_method == "jwt"
        assert result.error_message is not None
    
    def test_authenticate_api_key_success(self, auth_manager):
        """Test successful API key authentication."""
        # Create a valid API key
        key = auth_manager.api_key_handler.create_api_key("Test Key", "operator", 1)
        
        result = auth_manager.authenticate(key, "api_key")
        
        assert result.authenticated is True
        assert result.user_id is not None  # API key ID
        assert result.role == "operator"
        assert result.auth_method == "api_key"
        assert result.error_message is None
    
    def test_authenticate_api_key_invalid(self, auth_manager):
        """Test API key authentication with invalid key."""
        result = auth_manager.authenticate("invalid_key", "api_key")
        
        assert result.authenticated is False
        assert result.user_id is None
        assert result.role is None
        assert result.auth_method == "api_key"
        assert result.error_message is not None
    
    def test_authenticate_auto_jwt_first(self, auth_manager):
        """Test auto authentication with JWT token."""
        # Generate a valid JWT token
        token = auth_manager.jwt_handler.generate_token("test_user", "admin")
        
        result = auth_manager.authenticate(token, "auto")
        
        assert result.authenticated is True
        assert result.user_id == "test_user"
        assert result.role == "admin"
        assert result.auth_method == "jwt"
    
    def test_authenticate_auto_api_key_fallback(self, auth_manager):
        """Test auto authentication with API key fallback."""
        # Create a valid API key
        key = auth_manager.api_key_handler.create_api_key("Test Key", "viewer", 1)
        
        result = auth_manager.authenticate(key, "auto")
        
        assert result.authenticated is True
        assert result.role == "viewer"
        assert result.auth_method == "api_key"
    
    def test_authenticate_auto_both_fail(self, auth_manager):
        """Test auto authentication when both methods fail."""
        result = auth_manager.authenticate("invalid_token", "auto")
        
        assert result.authenticated is False
        assert result.auth_method == "jwt"  # Should try JWT first
    
    def test_has_permission_success(self, auth_manager):
        """Test permission checking with valid authentication."""
        # Generate token with admin role
        token = auth_manager.jwt_handler.generate_token("test_user", "admin")
        auth_result = auth_manager.authenticate(token, "jwt")
        
        # Check various permissions
        assert auth_manager.has_permission(auth_result, "viewer") is True
        assert auth_manager.has_permission(auth_result, "operator") is True
        assert auth_manager.has_permission(auth_result, "admin") is True
    
    def test_has_permission_insufficient(self, auth_manager):
        """Test permission checking with insufficient role."""
        # Generate token with viewer role
        token = auth_manager.jwt_handler.generate_token("test_user", "viewer")
        auth_result = auth_manager.authenticate(token, "jwt")
        
        # Check permissions
        assert auth_manager.has_permission(auth_result, "viewer") is True
        assert auth_manager.has_permission(auth_result, "operator") is False
        assert auth_manager.has_permission(auth_result, "admin") is False
    
    def test_has_permission_not_authenticated(self, auth_manager):
        """Test permission checking with unauthenticated user."""
        auth_result = AuthResult(
            authenticated=False,
            user_id=None,
            role=None,
            auth_method="jwt",
            error_message="Invalid token"
        )
        
        assert auth_manager.has_permission(auth_result, "viewer") is False
    
    def test_generate_jwt_token(self, auth_manager):
        """Test JWT token generation."""
        token = auth_manager.generate_jwt_token("test_user", "operator")
        
        assert isinstance(token, str)
        assert len(token) > 0
        
        # Verify token is valid
        result = auth_manager.authenticate(token, "jwt")
        assert result.authenticated is True
        assert result.user_id == "test_user"
        assert result.role == "operator"
    
    def test_create_api_key(self, auth_manager):
        """Test API key creation."""
        key = auth_manager.create_api_key("Test Key", "admin", 7)
        
        assert isinstance(key, str)
        assert len(key) == 32  # Standard key length
        
        # Verify key is valid
        result = auth_manager.authenticate(key, "api_key")
        assert result.authenticated is True
        assert result.role == "admin"
    
    def test_create_api_key_invalid_role(self, auth_manager):
        """Test API key creation with invalid role."""
        with pytest.raises(ValueError, match="Invalid role"):
            auth_manager.create_api_key("Test Key", "invalid_role", 1)
    
    def test_revoke_api_key(self, auth_manager):
        """Test API key revocation."""
        # Create a key
        key = auth_manager.create_api_key("Test Key", "viewer", 1)
        
        # Get the key ID from the stored keys
        stored_keys = list(auth_manager.api_key_handler._keys.values())
        key_id = stored_keys[0].key_id
        
        # Revoke the key
        result = auth_manager.revoke_api_key(key_id)
        assert result is True
        
        # Verify key is no longer valid
        auth_result = auth_manager.authenticate(key, "api_key")
        assert auth_result.authenticated is False
    
    def test_revoke_api_key_not_found(self, auth_manager):
        """Test API key revocation with non-existent key."""
        result = auth_manager.revoke_api_key("nonexistent_key")
        assert result is False
    
    def test_list_api_keys(self, auth_manager):
        """Test listing API keys."""
        # Create some keys
        auth_manager.create_api_key("Key 1", "viewer", 1)
        auth_manager.create_api_key("Key 2", "operator", 1)
        
        keys = auth_manager.list_api_keys()
        
        assert isinstance(keys, list)
        assert len(keys) == 2
        assert any(key["name"] == "Key 1" for key in keys)
        assert any(key["name"] == "Key 2" for key in keys)
    
    def test_cleanup_expired_keys(self, auth_manager):
        """Test cleanup of expired API keys."""
        # Create a key that will be cleaned up
        auth_manager.create_api_key("Expired Key", "viewer", 1)
        
        # Manually set the key to expired
        stored_keys = list(auth_manager.api_key_handler._keys.values())
        past_time = datetime.now(timezone.utc) - timedelta(hours=1)
        stored_keys[0].expires_at = past_time.isoformat()
        
        # Cleanup expired keys
        removed_count = auth_manager.cleanup_expired_keys()
        assert removed_count == 1
        
        # Verify key is removed
        keys = auth_manager.list_api_keys()
        assert len(keys) == 0


class TestAuthManagerIntegration:
    """
    Validates N3.2: End-to-end validation of authentication and role resolution
    """
    """Integration tests for authentication manager."""
    
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
    def auth_manager(self, temp_storage_file):
        """Create authentication manager for integration tests."""
        jwt_handler = JWTHandler("integration_test_secret")
        api_key_handler = APIKeyHandler(temp_storage_file)
        return AuthManager(jwt_handler, api_key_handler)
    
    def test_full_authentication_flow(self, auth_manager):
        """Test complete authentication flow with both methods."""
        # Test JWT authentication
        jwt_token = auth_manager.generate_jwt_token("jwt_user", "admin")
        jwt_result = auth_manager.authenticate(jwt_token, "auto")
        
        assert jwt_result.authenticated is True
        assert jwt_result.user_id == "jwt_user"
        assert jwt_result.role == "admin"
        assert jwt_result.auth_method == "jwt"
        
        # Test API key authentication
        api_key = auth_manager.create_api_key("API Key", "operator", 1)
        api_result = auth_manager.authenticate(api_key, "auto")
        
        assert api_result.authenticated is True
        assert api_result.role == "operator"
        assert api_result.auth_method == "api_key"
        
        # Test permission checking
        assert auth_manager.has_permission(jwt_result, "admin") is True
        assert auth_manager.has_permission(api_result, "operator") is True
        assert auth_manager.has_permission(api_result, "admin") is False
    
    def test_authentication_persistence(self, temp_storage_file):
        """Test that authentication state persists across instances."""
        # Create first manager and add API key
        jwt_handler1 = JWTHandler("test_secret")
        api_handler1 = APIKeyHandler(temp_storage_file)
        manager1 = AuthManager(jwt_handler1, api_handler1)
        
        api_key = manager1.create_api_key("Persistent Key", "viewer", 1)
        
        # Create second manager and verify key exists
        jwt_handler2 = JWTHandler("test_secret")
        api_handler2 = APIKeyHandler(temp_storage_file)
        manager2 = AuthManager(jwt_handler2, api_handler2)
        
        result = manager2.authenticate(api_key, "api_key")
        assert result.authenticated is True
        assert result.role == "viewer"
    
    def test_mixed_authentication_methods(self, auth_manager):
        """Test mixing JWT and API key authentication."""
        # Create both types of authentication
        jwt_token = auth_manager.generate_jwt_token("jwt_user", "viewer")
        api_key = auth_manager.create_api_key("API Key", "operator", 1)
        
        # Test both work independently
        jwt_result = auth_manager.authenticate(jwt_token, "jwt")
        api_result = auth_manager.authenticate(api_key, "api_key")
        
        assert jwt_result.authenticated is True
        assert api_result.authenticated is True
        assert jwt_result.auth_method == "jwt"
        assert api_result.auth_method == "api_key"
        
        # Test auto authentication with both
        jwt_auto = auth_manager.authenticate(jwt_token, "auto")
        api_auto = auth_manager.authenticate(api_key, "auto")
        
        assert jwt_auto.auth_method == "jwt"
        assert api_auto.auth_method == "api_key" 