"""
Unit tests for JWT handler authentication.

Tests JWT token generation, validation, and role-based access control
as specified in Architecture Decision AD-7.
"""

import pytest
import time
import jwt

from src.security.jwt_handler import JWTHandler, JWTClaims


class TestJWTClaims:
    """
    Validates N3.2: JWT validation and expiration handling (unit-level)
    """
    """Test JWT claims structure."""
    
    def test_create_claims(self):
        """Test creating JWT claims."""
        claims = JWTClaims.create("test_user", "viewer", 24)
        
        assert claims.user_id == "test_user"
        assert claims.role == "viewer"
        assert claims.iat > 0
        assert claims.exp > claims.iat
        assert claims.exp == claims.iat + (24 * 3600)
    
    def test_create_claims_default_expiry(self):
        """Test creating JWT claims with default expiry."""
        claims = JWTClaims.create("test_user", "admin")
        
        assert claims.user_id == "test_user"
        assert claims.role == "admin"
        assert claims.exp == claims.iat + (24 * 3600)  # Default 24 hours


class TestJWTHandler:
    """
    Validates N3.2: Token generation/validation; expiration semantics
    """
    """Test JWT handler functionality."""
    
    @pytest.fixture
    def jwt_handler(self):
        """Create JWT handler with test secret."""
        return JWTHandler("test_secret_key_12345")
    
    def test_init_with_secret(self, jwt_handler):
        """Test JWT handler initialization."""
        assert jwt_handler.secret_key == "test_secret_key_12345"
        assert jwt_handler.algorithm == "HS256"
    
    def test_init_without_secret(self):
        """Test JWT handler initialization without secret."""
        with pytest.raises(ValueError, match="Secret key must be provided"):
            JWTHandler("")
    
    def test_generate_token_success(self, jwt_handler):
        """Test successful token generation."""
        token = jwt_handler.generate_token("test_user", "viewer", 1)
        
        assert token is not None
        assert isinstance(token, str)
        assert len(token) > 0
        
        # Verify token can be decoded
        payload = jwt.decode(token, jwt_handler.secret_key, algorithms=[jwt_handler.algorithm])
        assert payload["user_id"] == "test_user"
        assert payload["role"] == "viewer"
        assert "iat" in payload
        assert "exp" in payload
    
    def test_generate_token_invalid_role(self, jwt_handler):
        """Test token generation with invalid role."""
        with pytest.raises(ValueError, match="Invalid role"):
            jwt_handler.generate_token("test_user", "invalid_role")
    
    def test_generate_token_empty_user_id(self, jwt_handler):
        """Test token generation with empty user ID."""
        with pytest.raises(ValueError, match="User ID must be provided"):
            jwt_handler.generate_token("", "viewer")
    
    def test_validate_token_success(self, jwt_handler):
        """Test successful token validation."""
        token = jwt_handler.generate_token("test_user", "operator", 1)
        claims = jwt_handler.validate_token(token)
        
        assert claims is not None
        assert claims.user_id == "test_user"
        assert claims.role == "operator"
        assert claims.iat > 0
        assert claims.exp > claims.iat
    
    def test_validate_token_invalid_signature(self, jwt_handler):
        """Test token validation with invalid signature."""
        # Create token with different secret
        other_handler = JWTHandler("different_secret")
        token = other_handler.generate_token("test_user", "viewer", 1)
        
        claims = jwt_handler.validate_token(token)
        assert claims is None
    
    def test_validate_token_expired(self, jwt_handler):
        """Test validation of expired token."""
        # Create an expired token by manually setting expiry in the past
        now = int(time.time())
        expired_payload = {
            "user_id": "test_user",
            "role": "viewer",
            "iat": now - 3600,  # 1 hour ago
            "exp": now - 1800    # 30 minutes ago (expired)
        }
        expired_token = jwt.encode(expired_payload, jwt_handler.secret_key, algorithm=jwt_handler.algorithm)
        
        claims = jwt_handler.validate_token(expired_token)
        assert claims is None
    
    def test_validate_token_missing_fields(self, jwt_handler):
        """Test validation of token with missing fields."""
        # Create payload with missing fields
        payload = {"user_id": "test_user", "role": "viewer"}
        token = jwt.encode(payload, jwt_handler.secret_key, algorithm=jwt_handler.algorithm)
        
        claims = jwt_handler.validate_token(token)
        assert claims is None
    
    def test_validate_token_invalid_role(self, jwt_handler):
        """Test validation of token with invalid role."""
        # Create payload with invalid role
        payload = {
            "user_id": "test_user",
            "role": "invalid_role",
            "iat": int(time.time()),
            "exp": int(time.time()) + 3600
        }
        token = jwt.encode(payload, jwt_handler.secret_key, algorithm=jwt_handler.algorithm)
        
        claims = jwt_handler.validate_token(token)
        assert claims is None
    
    def test_validate_token_none(self, jwt_handler):
        """Test validation of None token."""
        claims = jwt_handler.validate_token(None)
        assert claims is None
    
    def test_validate_token_empty(self, jwt_handler):
        """Test validation of empty token."""
        claims = jwt_handler.validate_token("")
        assert claims is None
    
    def test_is_token_expired_true(self, jwt_handler):
        """Test checking expired token."""
        # Create expired token
        payload = {
            "user_id": "test_user",
            "role": "viewer",
            "iat": int(time.time()) - 3600,
            "exp": int(time.time()) - 1800  # Expired 30 minutes ago
        }
        token = jwt.encode(payload, jwt_handler.secret_key, algorithm=jwt_handler.algorithm)
        
        assert jwt_handler.is_token_expired(token) is True
    
    def test_is_token_expired_false(self, jwt_handler):
        """Test checking non-expired token."""
        # Create valid token
        payload = {
            "user_id": "test_user",
            "role": "viewer",
            "iat": int(time.time()),
            "exp": int(time.time()) + 3600  # Valid for 1 hour
        }
        token = jwt.encode(payload, jwt_handler.secret_key, algorithm=jwt_handler.algorithm)
        
        assert jwt_handler.is_token_expired(token) is False
    
    def test_get_token_info_success(self, jwt_handler):
        """Test getting token information."""
        token = jwt_handler.generate_token("test_user", "admin", 2)
        info = jwt_handler.get_token_info(token)
        
        assert info is not None
        assert info["user_id"] == "test_user"
        assert info["role"] == "admin"
        assert info["issued_at"] > 0
        assert info["expires_at"] > info["issued_at"]
        assert info["expired"] is False
    
    def test_get_token_info_invalid_token(self, jwt_handler):
        """Test getting info from invalid token."""
        info = jwt_handler.get_token_info("invalid_token")
        assert info is None
    
    def test_has_permission_viewer(self, jwt_handler):
        """Test permission checking for viewer role."""
        claims = JWTClaims.create("test_user", "viewer")
        
        # Viewer should have viewer permission
        assert jwt_handler.has_permission(claims, "viewer") is True
        
        # Viewer should not have operator permission
        assert jwt_handler.has_permission(claims, "operator") is False
        
        # Viewer should not have admin permission
        assert jwt_handler.has_permission(claims, "admin") is False
    
    def test_has_permission_operator(self, jwt_handler):
        """Test permission checking for operator role."""
        claims = JWTClaims.create("test_user", "operator")
        
        # Operator should have viewer permission
        assert jwt_handler.has_permission(claims, "viewer") is True
        
        # Operator should have operator permission
        assert jwt_handler.has_permission(claims, "operator") is True
        
        # Operator should not have admin permission
        assert jwt_handler.has_permission(claims, "admin") is False
    
    def test_has_permission_admin(self, jwt_handler):
        """Test permission checking for admin role."""
        claims = JWTClaims.create("test_user", "admin")
        
        # Admin should have all permissions
        assert jwt_handler.has_permission(claims, "viewer") is True
        assert jwt_handler.has_permission(claims, "operator") is True
        assert jwt_handler.has_permission(claims, "admin") is True
    
    def test_has_permission_invalid_role(self, jwt_handler):
        """Test permission checking with invalid role."""
        claims = JWTClaims.create("test_user", "viewer")
        
        # Invalid required role should return False
        assert jwt_handler.has_permission(claims, "invalid_role") is False
    
    def test_valid_roles_constant(self, jwt_handler):
        """Test VALID_ROLES constant."""
        assert "viewer" in jwt_handler.VALID_ROLES
        assert "operator" in jwt_handler.VALID_ROLES
        assert "admin" in jwt_handler.VALID_ROLES
        assert len(jwt_handler.VALID_ROLES) == 3


class TestJWTHandlerIntegration:
    """
    Validates N3.2: Integration behavior with token verification
    """
    """Integration tests for JWT handler."""
    
    @pytest.fixture
    def jwt_handler(self):
        """Create JWT handler for integration tests."""
        return JWTHandler("integration_test_secret_key")
    
    def test_full_token_lifecycle(self, jwt_handler):
        """Test complete token lifecycle."""
        # Generate token
        token = jwt_handler.generate_token("integration_user", "operator", 1)
        assert token is not None
        
        # Validate token
        claims = jwt_handler.validate_token(token)
        assert claims is not None
        assert claims.user_id == "integration_user"
        assert claims.role == "operator"
        
        # Check permissions
        assert jwt_handler.has_permission(claims, "viewer") is True
        assert jwt_handler.has_permission(claims, "operator") is True
        assert jwt_handler.has_permission(claims, "admin") is False
        
        # Check token info
        info = jwt_handler.get_token_info(token)
        assert info["user_id"] == "integration_user"
        assert info["role"] == "operator"
        assert info["expired"] is False
    
    def test_token_expiry_handling(self, jwt_handler):
        """Test token expiry handling."""
        # Create a valid token first
        token = jwt_handler.generate_token("expiry_user", "viewer", 1)
        
        # Token should be valid initially
        claims = jwt_handler.validate_token(token)
        assert claims is not None
        
        # Create an expired token manually
        now = int(time.time())
        expired_payload = {
            "user_id": "expiry_user",
            "role": "viewer",
            "iat": now - 3600,  # 1 hour ago
            "exp": now - 1800    # 30 minutes ago (expired)
        }
        expired_token = jwt.encode(expired_payload, jwt_handler.secret_key, algorithm=jwt_handler.algorithm)
        
        # Token should be expired
        claims = jwt_handler.validate_token(expired_token)
        assert claims is None
        
        # Token info should show expired
        info = jwt_handler.get_token_info(expired_token)
        assert info["expired"] is True 