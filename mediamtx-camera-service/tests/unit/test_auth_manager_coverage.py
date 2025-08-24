"""
Unit tests for AuthManager coverage gaps.

Requirements Coverage:
- REQ-SEC-001: Authentication and authorization
- REQ-SEC-002: Role-based access control
- REQ-SEC-003: JWT token validation
- REQ-SEC-004: API key validation

Test Categories: Unit
API Documentation Reference: docs/api/json-rpc-methods.md
"""

import pytest
from unittest.mock import Mock, patch, MagicMock
from src.security.auth_manager import AuthManager, AuthResult
from src.security.jwt_handler import JWTHandler
from src.security.api_key_handler import APIKeyHandler


@pytest.mark.unit
class TestAuthManagerCoverage:
    """Test cases to cover missing lines in AuthManager."""
    
    def setup_method(self):
        """Set up test fixtures."""
        self.jwt_handler = Mock(spec=JWTHandler)
        self.api_key_handler = Mock(spec=APIKeyHandler)
        self.auth_manager = AuthManager(self.jwt_handler, self.api_key_handler)
    
    def test_authenticate_no_token_provided(self):
        """REQ-SEC-001: Test authentication with no token."""
        result = self.auth_manager.authenticate(None, "jwt")
        
        assert not result.authenticated
        assert result.auth_method == "jwt"
        assert result.error_message == "No authentication token provided"
    
    def test_authenticate_jwt_fails_api_key_succeeds(self):
        """REQ-SEC-001: Test auto authentication with JWT failure, API key success."""
        # Mock JWT failure
        self.jwt_handler.validate_token.return_value = None
        
        # Mock API key success
        mock_api_key = Mock()
        mock_api_key.key_id = "test_key_123"
        mock_api_key.role = "operator"
        self.api_key_handler.validate_api_key.return_value = mock_api_key
        
        result = self.auth_manager.authenticate("valid_api_key", "auto")
        
        assert result.authenticated
        assert result.user_id == "api_key_test_key_123"
        assert result.role == "operator"
        assert result.auth_method == "api_key"
    
    def test_authenticate_jwt_specific_failure(self):
        """REQ-SEC-001: Test JWT-specific authentication failure."""
        self.jwt_handler.validate_token.return_value = None
        
        result = self.auth_manager.authenticate("invalid_jwt", "jwt")
        
        assert not result.authenticated
        assert result.auth_method == "jwt"
        assert result.error_message == "Invalid or expired JWT token"
    
    def test_authenticate_api_key_specific_failure(self):
        """REQ-SEC-001: Test API key-specific authentication failure."""
        self.api_key_handler.validate_api_key.return_value = None
        
        result = self.auth_manager.authenticate("invalid_api_key", "api_key")
        
        assert not result.authenticated
        assert result.auth_method == "api_key"
        assert result.error_message == "Invalid or expired API key"
    
    def test_authenticate_invalid_auth_type_falls_back_to_auto(self):
        """REQ-SEC-001: Test authentication with invalid auth type falls back to auto."""
        # Mock JWT success
        mock_claims = Mock()
        mock_claims.user_id = "test_user"
        mock_claims.role = "admin"
        mock_claims.exp = 1234567890
        self.jwt_handler.validate_token.return_value = mock_claims
        
        result = self.auth_manager.authenticate("valid_token", "invalid_type")
        
        assert result.authenticated
        assert result.user_id == "test_user"
        assert result.role == "admin"
        assert result.auth_method == "jwt"
    
    def test_authenticate_auto_both_fail_returns_jwt_error(self):
        """REQ-SEC-001: Test auto authentication when both JWT and API key fail."""
        # Mock both failures
        self.jwt_handler.validate_token.return_value = None
        self.api_key_handler.validate_api_key.return_value = None
        
        result = self.auth_manager.authenticate("invalid_token", "auto")
        
        assert not result.authenticated
        assert result.auth_method == "jwt"
        assert result.error_message == "Invalid or expired JWT token"
    
    def test_authenticate_jwt_exception_handling(self):
        """REQ-SEC-003: Test JWT authentication exception handling."""
        self.jwt_handler.validate_token.side_effect = Exception("JWT validation error")
        
        result = self.auth_manager.authenticate("token", "jwt")
        
        assert not result.authenticated
        assert result.auth_method == "jwt"
        assert result.error_message == "JWT authentication failed"
    
    def test_authenticate_api_key_exception_handling(self):
        """REQ-SEC-004: Test API key authentication exception handling."""
        self.api_key_handler.validate_api_key.side_effect = Exception("API key validation error")
        
        result = self.auth_manager.authenticate("key", "api_key")
        
        assert not result.authenticated
        assert result.auth_method == "api_key"
        assert result.error_message == "API key authentication failed"
    
    def test_has_permission_no_authentication(self):
        """REQ-SEC-002: Test permission check with no authentication."""
        auth_result = AuthResult(authenticated=False, user_id=None, role=None)
        
        has_permission = self.auth_manager.has_permission(auth_result, "viewer")
        
        assert not has_permission
    
    def test_has_permission_no_role(self):
        """REQ-SEC-002: Test permission check with no role."""
        auth_result = AuthResult(authenticated=True, user_id="user", role=None)
        
        has_permission = self.auth_manager.has_permission(auth_result, "viewer")
        
        assert not has_permission
    
    def test_has_permission_unknown_role(self):
        """REQ-SEC-002: Test permission check with unknown role."""
        auth_result = AuthResult(authenticated=True, user_id="user", role="unknown")
        
        has_permission = self.auth_manager.has_permission(auth_result, "viewer")
        
        assert not has_permission
    
    def test_has_permission_unknown_required_role(self):
        """REQ-SEC-002: Test permission check with unknown required role."""
        auth_result = AuthResult(authenticated=True, user_id="user", role="admin")
        
        has_permission = self.auth_manager.has_permission(auth_result, "unknown")
        
        # Unknown required role gets level 0, admin has level 3, so permission granted
        assert has_permission
    
    def test_generate_jwt_token_with_custom_expiry(self):
        """REQ-SEC-003: Test JWT token generation with custom expiry."""
        expected_token = "custom.jwt.token"
        self.jwt_handler.generate_token.return_value = expected_token
        
        token = self.auth_manager.generate_jwt_token("user", "admin", 48)
        
        assert token == expected_token
        self.jwt_handler.generate_token.assert_called_once_with("user", "admin", 48)
    
    def test_create_api_key_with_expiry(self):
        """REQ-SEC-004: Test API key creation with expiry."""
        expected_key = "api_key_123"
        self.api_key_handler.create_api_key.return_value = expected_key
        
        key = self.auth_manager.create_api_key("test_key", "operator", 30)
        
        assert key == expected_key
        self.api_key_handler.create_api_key.assert_called_once_with("test_key", "operator", 30)
    
    def test_revoke_api_key_success(self):
        """REQ-SEC-004: Test API key revocation success."""
        self.api_key_handler.revoke_api_key.return_value = True
        
        result = self.auth_manager.revoke_api_key("key_123")
        
        assert result is True
        self.api_key_handler.revoke_api_key.assert_called_once_with("key_123")
    
    def test_revoke_api_key_failure(self):
        """REQ-SEC-004: Test API key revocation failure."""
        self.api_key_handler.revoke_api_key.return_value = False
        
        result = self.auth_manager.revoke_api_key("key_123")
        
        assert result is False
        self.api_key_handler.revoke_api_key.assert_called_once_with("key_123")
