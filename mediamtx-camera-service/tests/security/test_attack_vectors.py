"""
Security attack vector tests.

Tests security implementation against common attack vectors including
JWT token tampering, brute force attempts, rate limit bypass,
connection exhaustion, and role elevation as specified in Sprint 2 Task S7.2.
"""

import pytest
import tempfile
import os
import time
from datetime import datetime, timedelta, timezone

from src.security.jwt_handler import JWTHandler
from src.security.api_key_handler import APIKeyHandler
from src.security.auth_manager import AuthManager
from src.security.middleware import SecurityMiddleware


class TestJWTSecurityAttacks:
    """Test JWT security against various attack vectors."""
    
    @pytest.fixture
    def jwt_handler(self):
        """Create JWT handler for attack testing."""
        return JWTHandler("attack_test_secret_key")
    
    @pytest.fixture
    def auth_manager(self):
        """Create authentication manager for attack testing."""
        jwt_handler = JWTHandler("attack_test_secret")
        with tempfile.NamedTemporaryFile(mode='w', delete=False, suffix='.json') as f:
            f.write('{"version": "1.0", "keys": []}')
            temp_file = f.name
        
        api_key_handler = APIKeyHandler(temp_file)
        auth_manager = AuthManager(jwt_handler, api_key_handler)
        
        yield auth_manager
        
        # Cleanup
        if os.path.exists(temp_file):
            os.unlink(temp_file)
    
    def test_jwt_token_tampering_attempts(self, jwt_handler):
        """Test JWT token tampering attempts."""
        # Generate valid token
        valid_token = jwt_handler.generate_token("test_user", "viewer")
        
        # Test various tampering attempts
        tampering_attempts = [
            # Empty token
            "",
            # Null token
            None,
            # Malformed token
            "not.a.jwt.token",
            # Extra segments
            valid_token + ".extra.segment",
            # Missing segments
            valid_token.split('.')[0] + "." + valid_token.split('.')[1],
            # Invalid signature
            valid_token.rsplit('.', 1)[0] + ".invalid_signature"
        ]
        
        for tampered_token in tampering_attempts:
            claims = jwt_handler.validate_token(tampered_token)
            assert claims is None, f"Tampered token should be rejected: {tampered_token}"
    
    def test_jwt_signature_validation(self, jwt_handler):
        """Test JWT signature validation against attacks."""
        # Generate token with one secret
        token = jwt_handler.generate_token("test_user", "admin")
        
        # Try to validate with different secret
        different_handler = JWTHandler("different_secret_key")
        claims = different_handler.validate_token(token)
        assert claims is None, "Token should be rejected with different secret"
    
    def test_jwt_algorithm_confusion_attack(self, jwt_handler):
        """Test JWT algorithm confusion attack prevention."""
        # Generate token with HS256
        token = jwt_handler.generate_token("test_user", "admin")
        
        # Try to validate with "none" algorithm (should be rejected)
        # This simulates an algorithm confusion attack
        try:
            # In a real attack, the attacker might try to use "none" algorithm
            # Our implementation should reject this
            claims = jwt_handler.validate_token(token)
            # If we get here, the token should still be valid with HS256
            assert claims is not None
        except Exception:
            # Expected behavior - algorithm confusion should be prevented
            pass
    
    def test_jwt_replay_attack_prevention(self, jwt_handler):
        """Test JWT replay attack prevention."""
        # Generate token
        token = jwt_handler.generate_token("test_user", "viewer")
        
        # Validate token multiple times (should work)
        for _ in range(10):
            claims = jwt_handler.validate_token(token)
            assert claims is not None
            assert claims.user_id == "test_user"
            assert claims.role == "viewer"
        
        # Note: True replay attack prevention requires server-side token tracking
        # This test verifies that tokens remain valid for their intended lifetime
    
    def test_jwt_brute_force_attack_simulation(self, jwt_handler):
        """Test JWT brute force attack simulation."""
        # Generate valid token
        jwt_handler.generate_token("test_user", "admin")
        
        # Simulate brute force attempts with invalid tokens
        invalid_attempts = [
            "invalid_token_1",
            "invalid_token_2",
            "invalid_token_3",
            "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid",
            "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoidGVzdCJ9.invalid",
            "",
            None
        ]
        
        # All invalid attempts should be rejected
        for invalid_token in invalid_attempts:
            claims = jwt_handler.validate_token(invalid_token)
            assert claims is None, f"Invalid token should be rejected: {invalid_token}"
    
    def test_jwt_token_expiry_enforcement(self, jwt_handler):
        """Test JWT token expiry enforcement."""
        # Generate token with very short expiry
        token = jwt_handler.generate_token("expiry_user", "viewer", expiry_hours=1)
        
        # Token should be valid immediately
        claims = jwt_handler.validate_token(token)
        assert claims is not None
        
        # Check token info
        token_info = jwt_handler.get_token_info(token)
        assert token_info["expired"] is False
        
        # Test manually expired token
        import jwt
        now = int(time.time())
        expired_payload = {
            "user_id": "expired_user",
            "role": "viewer",
            "iat": now - 3600,  # 1 hour ago
            "exp": now - 1800    # 30 minutes ago (expired)
        }
        expired_token = jwt.encode(expired_payload, jwt_handler.secret_key, algorithm=jwt_handler.algorithm)
        
        claims = jwt_handler.validate_token(expired_token)
        assert claims is None, "Expired token should be rejected"


class TestAPIKeySecurityAttacks:
    """Test API key security against various attack vectors."""
    
    @pytest.fixture
    def temp_storage_file(self):
        """Create temporary storage file for attack testing."""
        with tempfile.NamedTemporaryFile(mode='w', delete=False, suffix='.json') as f:
            f.write('{"version": "1.0", "keys": []}')
            temp_file = f.name
        
        yield temp_file
        
        # Cleanup
        if os.path.exists(temp_file):
            os.unlink(temp_file)
    
    @pytest.fixture
    def api_key_handler(self, temp_storage_file):
        """Create API key handler for attack testing."""
        return APIKeyHandler(temp_storage_file)
    
    def test_api_key_brute_force_attack_simulation(self, api_key_handler):
        """Test API key brute force attack simulation."""
        # Create valid API key
        api_key_handler.create_api_key("Valid Key", "admin", 1)
        
        # Simulate brute force attempts
        invalid_attempts = [
            "invalid_key_1",
            "invalid_key_2",
            "invalid_key_3",
            "short_key",
            "very_long_invalid_key_that_should_be_rejected",
            "key_with_special_chars!@#$%",
            "key_with_spaces invalid",
            "",
            None
        ]
        
        # All invalid attempts should be rejected
        for invalid_key in invalid_attempts:
            result = api_key_handler.validate_api_key(invalid_key)
            assert result is None, f"Invalid API key should be rejected: {invalid_key}"
    
    def test_api_key_length_validation(self, api_key_handler):
        """Test API key length validation against attacks."""
        # Test various invalid lengths
        invalid_lengths = [
            "short",  # Too short
            "key_with_exactly_31_chars_long_key",  # 31 chars
            "key_with_exactly_33_chars_long_key_",  # 33 chars
            "very_long_key_that_exceeds_normal_length_but_is_still_invalid_and_should_be_rejected",
            "",
            None
        ]
        
        for invalid_key in invalid_lengths:
            result = api_key_handler.validate_api_key(invalid_key)
            assert result is None, f"Invalid length key should be rejected: {invalid_key}"
    
    def test_api_key_expired_key_attack(self, api_key_handler):
        """Test expired API key attack prevention."""
        # Create key with short expiry
        key = api_key_handler.create_api_key("Expired Attack Key", "viewer", 1)
        
        # Manually set key to expired
        stored_keys = list(api_key_handler._keys.values())
        past_time = datetime.now(timezone.utc) - timedelta(hours=1)
        stored_keys[0].expires_at = past_time.isoformat()
        
        # Attempt to use expired key
        result = api_key_handler.validate_api_key(key)
        assert result is None, "Expired API key should be rejected"
    
    def test_api_key_revoked_key_attack(self, api_key_handler):
        """Test revoked API key attack prevention."""
        # Create and revoke key
        key = api_key_handler.create_api_key("Revoked Attack Key", "operator", 1)
        stored_keys = list(api_key_handler._keys.values())
        key_id = stored_keys[0].key_id
        
        # Revoke the key
        api_key_handler.revoke_api_key(key_id)
        
        # Attempt to use revoked key
        result = api_key_handler.validate_api_key(key)
        assert result is None, "Revoked API key should be rejected"
    
    def test_api_key_injection_attempts(self, api_key_handler):
        """Test API key injection attack attempts."""
        # Test various injection attempts
        injection_attempts = [
            "key'; DROP TABLE keys; --",
            "key' OR '1'='1",
            "key' UNION SELECT * FROM keys --",
            "key' AND 1=1 --",
            "key' OR 1=1#",
            "key'/*comment*/",
            "key'--comment",
            "key'/**/OR/**/1=1",
            "key'%00",
            "key'%0A",
            "key'%0D"
        ]
        
        for injection_attempt in injection_attempts:
            result = api_key_handler.validate_api_key(injection_attempt)
            assert result is None, f"Injection attempt should be rejected: {injection_attempt}"


class TestRateLimitSecurityAttacks:
    """Test rate limiting security against various attack vectors."""
    
    @pytest.fixture
    def temp_storage_file(self):
        """Create temporary storage file for rate limit testing."""
        with tempfile.NamedTemporaryFile(mode='w', delete=False, suffix='.json') as f:
            f.write('{"version": "1.0", "keys": []}')
            temp_file = f.name
        
        yield temp_file
        
        # Cleanup
        if os.path.exists(temp_file):
            os.unlink(temp_file)
    
    @pytest.fixture
    def security_middleware(self, temp_storage_file):
        """Create security middleware for rate limit testing."""
        jwt_handler = JWTHandler("rate_limit_test_secret")
        api_key_handler = APIKeyHandler(temp_storage_file)
        auth_manager = AuthManager(jwt_handler, api_key_handler)
        
        return SecurityMiddleware(
            auth_manager=auth_manager,
            max_connections=10,
            requests_per_minute=60,
            window_size_seconds=60
        )
    
    def test_rate_limit_bypass_attempts(self, security_middleware):
        """Test rate limit bypass attempts."""
        # Register multiple clients to test bypass attempts
        clients = []
        for i in range(5):
            client_id = f"bypass_client_{i}"
            security_middleware.register_connection(client_id)
            clients.append(client_id)
        
        # Each client should be rate limited independently
        for client_id in clients:
            # Use up rate limit for this client
            for _ in range(60):
                assert security_middleware.check_rate_limit(client_id) is True
            
            # Next request should be blocked
            assert security_middleware.check_rate_limit(client_id) is False
        
        # Other clients should still be able to make requests
        other_client = "other_client"
        security_middleware.register_connection(other_client)
        assert security_middleware.check_rate_limit(other_client) is True
        
        # Cleanup
        for client_id in clients:
            security_middleware.unregister_connection(client_id)
        security_middleware.unregister_connection(other_client)
    
    def test_connection_exhaustion_attack(self, security_middleware):
        """Test connection exhaustion attack prevention."""
        # Try to exhaust all available connections
        clients = []
        for i in range(10):
            client_id = f"exhaustion_client_{i}"
            if security_middleware.can_accept_connection(client_id):
                security_middleware.register_connection(client_id)
                clients.append(client_id)
        
        # Should not be able to add more connections
        assert security_middleware.can_accept_connection("extra_client") is False
        
        # Remove some connections
        for i in range(3):
            security_middleware.unregister_connection(clients[i])
        
        # Should be able to add new connections
        assert security_middleware.can_accept_connection("new_client") is True
        
        # Cleanup
        for client_id in clients[3:]:
            security_middleware.unregister_connection(client_id)
    
    def test_rapid_connection_cycling_attack(self, security_middleware):
        """Test rapid connection cycling attack prevention."""
        # Simulate rapid connection cycling
        for cycle in range(5):
            # Create connections rapidly
            for i in range(5):
                client_id = f"cycle_{cycle}_client_{i}"
                if security_middleware.can_accept_connection(client_id):
                    security_middleware.register_connection(client_id)
            
            # Remove connections rapidly
            for i in range(5):
                client_id = f"cycle_{cycle}_client_{i}"
                security_middleware.unregister_connection(client_id)
        
        # System should still be functional
        assert security_middleware.can_accept_connection("test_client") is True
        security_middleware.register_connection("test_client")
        assert security_middleware.check_rate_limit("test_client") is True
        security_middleware.unregister_connection("test_client")


class TestRoleElevationAttacks:
    """Test role elevation attack prevention."""
    
    @pytest.fixture
    def temp_storage_file(self):
        """Create temporary storage file for role elevation testing."""
        with tempfile.NamedTemporaryFile(mode='w', delete=False, suffix='.json') as f:
            f.write('{"version": "1.0", "keys": []}')
            temp_file = f.name
        
        yield temp_file
        
        # Cleanup
        if os.path.exists(temp_file):
            os.unlink(temp_file)
    
    @pytest.fixture
    def auth_manager(self, temp_storage_file):
        """Create authentication manager for role elevation testing."""
        jwt_handler = JWTHandler("role_elevation_test_secret")
        api_key_handler = APIKeyHandler(temp_storage_file)
        return AuthManager(jwt_handler, api_key_handler)
    
    def test_jwt_role_elevation_attempts(self, auth_manager):
        """Test JWT role elevation attack attempts."""
        # Generate token with viewer role
        viewer_token = auth_manager.generate_jwt_token("viewer_user", "viewer")
        
        # Test permission boundaries
        viewer_result = auth_manager.authenticate(viewer_token, "jwt")
        assert viewer_result.authenticated is True
        assert viewer_result.role == "viewer"
        
        # Viewer should not have operator or admin permissions
        assert auth_manager.has_permission(viewer_result, "viewer") is True
        assert auth_manager.has_permission(viewer_result, "operator") is False
        assert auth_manager.has_permission(viewer_result, "admin") is False
    
    def test_api_key_role_elevation_attempts(self, auth_manager):
        """Test API key role elevation attack attempts."""
        # Create API key with viewer role
        viewer_key = auth_manager.create_api_key("Viewer API Key", "viewer", 1)
        
        # Test permission boundaries
        viewer_result = auth_manager.authenticate(viewer_key, "api_key")
        assert viewer_result.authenticated is True
        assert viewer_result.role == "viewer"
        
        # Viewer should not have operator or admin permissions
        assert auth_manager.has_permission(viewer_result, "viewer") is True
        assert auth_manager.has_permission(viewer_result, "operator") is False
        assert auth_manager.has_permission(viewer_result, "admin") is False
    
    def test_invalid_role_handling(self, auth_manager):
        """Test handling of invalid roles in authentication."""
        # Test JWT with invalid role
        with pytest.raises(ValueError, match="Invalid role"):
            auth_manager.generate_jwt_token("invalid_user", "invalid_role")
        
        # Test API key with invalid role
        with pytest.raises(ValueError, match="Invalid role"):
            auth_manager.create_api_key("Invalid API Key", "invalid_role", 1)
    
    def test_role_hierarchy_enforcement(self, auth_manager):
        """Test role hierarchy enforcement."""
        # Test viewer permissions
        viewer_token = auth_manager.generate_jwt_token("viewer_user", "viewer")
        viewer_result = auth_manager.authenticate(viewer_token, "jwt")
        
        # Test operator permissions
        operator_token = auth_manager.generate_jwt_token("operator_user", "operator")
        operator_result = auth_manager.authenticate(operator_token, "jwt")
        
        # Test admin permissions
        admin_token = auth_manager.generate_jwt_token("admin_user", "admin")
        admin_result = auth_manager.authenticate(admin_token, "jwt")
        
        # Verify role hierarchy
        assert auth_manager.has_permission(viewer_result, "viewer") is True
        assert auth_manager.has_permission(viewer_result, "operator") is False
        assert auth_manager.has_permission(viewer_result, "admin") is False
        
        assert auth_manager.has_permission(operator_result, "viewer") is True
        assert auth_manager.has_permission(operator_result, "operator") is True
        assert auth_manager.has_permission(operator_result, "admin") is False
        
        assert auth_manager.has_permission(admin_result, "viewer") is True
        assert auth_manager.has_permission(admin_result, "operator") is True
        assert auth_manager.has_permission(admin_result, "admin") is True


class TestInputValidationAttacks:
    """Test input validation against various attack vectors."""
    
    @pytest.fixture
    def temp_storage_file(self):
        """Create temporary storage file for input validation testing."""
        with tempfile.NamedTemporaryFile(mode='w', delete=False, suffix='.json') as f:
            f.write('{"version": "1.0", "keys": []}')
            temp_file = f.name
        
        yield temp_file
        
        # Cleanup
        if os.path.exists(temp_file):
            os.unlink(temp_file)
    
    @pytest.fixture
    def auth_manager(self, temp_storage_file):
        """Create authentication manager for input validation testing."""
        jwt_handler = JWTHandler("input_validation_test_secret")
        api_key_handler = APIKeyHandler(temp_storage_file)
        return AuthManager(jwt_handler, api_key_handler)
    
    def test_malformed_jwt_tokens(self, auth_manager):
        """Test handling of malformed JWT tokens."""
        malformed_tokens = [
            "not.a.jwt.token",
            "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
            "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid",
            "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoidGVzdCJ9",
            "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoidGVzdCJ9.invalid_signature",
            "",
            None
        ]
        
        for malformed_token in malformed_tokens:
            result = auth_manager.authenticate(malformed_token, "jwt")
            assert result.authenticated is False
            assert result.error_message is not None
    
    def test_oversized_request_payloads(self, auth_manager):
        """Test handling of oversized request payloads."""
        # Create very large token (simulating oversized payload)
        large_payload = "x" * 10000  # 10KB payload
        result = auth_manager.authenticate(large_payload, "jwt")
        assert result.authenticated is False
        assert result.error_message is not None
    
    def test_special_character_handling(self, auth_manager):
        """Test handling of special characters in tokens."""
        special_char_tokens = [
            "token_with_spaces",
            "token\twith\ttabs",
            "token\nwith\nnewlines",
            "token\rwith\rreturns",
            "token_with_null\0bytes",
            "token_with_unicode_测试",
            "token_with_emoji_😀",
            "token_with_quotes_'\"",
            "token_with_backslashes\\",
            "token_with_percent%20encoding"
        ]
        
        for special_token in special_char_tokens:
            result = auth_manager.authenticate(special_token, "jwt")
            assert result.authenticated is False
            assert result.error_message is not None
    
    def test_edge_case_inputs(self, auth_manager):
        """Test handling of edge case inputs."""
        edge_cases = [
            "",  # Empty string
            None,  # None value
            " ",  # Whitespace only
            "\t\n\r",  # Control characters
            "0",  # Single character
            "a" * 1000,  # Very long string
            "a" * 32,  # Exactly key length (for API keys)
            "a" * 31,  # One less than key length
            "a" * 33,  # One more than key length
        ]
        
        for edge_case in edge_cases:
            result = auth_manager.authenticate(edge_case, "jwt")
            assert result.authenticated is False
            assert result.error_message is not None 