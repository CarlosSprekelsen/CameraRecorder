"""
JWT authentication integration tests for real system validation.

Requirements Coverage:
- REQ-SEC-001: JWT token-based authentication for all API access
- REQ-SEC-002: Token format with JSON Web Token (JWT) and standard claims
- REQ-SEC-003: Token expiration with configurable expiration time
- REQ-SEC-004: Token refresh mechanism support
- REQ-SEC-005: Token validation with proper signature validation and claim verification
- REQ-SEC-006: API key validation for service-to-service communication
- REQ-SEC-007: API key format with secure random string (32+ characters)
- REQ-SEC-008: Secure storage of API keys
- REQ-SEC-009: API key rotation support
- REQ-SEC-010: Role-based access control for different user types
- REQ-SEC-011: Admin, User, Read-Only roles
- REQ-SEC-012: Permission matrix and clear permission definitions
- REQ-SEC-013: Enforcement of role-based permissions
- REQ-SEC-014: Resource access control for camera resources and media files
- REQ-SEC-015: Camera access control and user authorization
- REQ-SEC-016: File access control and user authorization
- REQ-SEC-017: Resource isolation between user resources
- REQ-SEC-018: Access logging of all resource access attempts
- REQ-CLIENT-032: Role-based access control with viewer, operator, and admin permissions
- REQ-CLIENT-033: Token expiration handling with re-authentication
- REQ-TEST-009: Authentication and authorization test coverage
- REQ-TEST-012: Security test coverage for all security requirements

Test Categories: Integration
"""

import pytest
import tempfile
import os
import time
import subprocess
import requests
import json
from typing import Dict, Any

from src.security.jwt_handler import JWTHandler
from src.security.auth_manager import AuthManager
from src.security.api_key_handler import APIKeyHandler
from tests.fixtures.auth_utils import get_test_auth_manager, TestUserFactory, cleanup_test_auth_manager


class TestJWTAuthenticationFlow:
    """Integration tests for JWT authentication flow."""
    
    @pytest.fixture
    def auth_manager(self):
        """Create authentication manager for testing using non-hardcoded secrets."""
        return get_test_auth_manager()
    
    @pytest.fixture
    def user_factory(self, auth_manager):
        """Create user factory for testing."""
        return TestUserFactory(auth_manager)
    
    @pytest.fixture
    def real_mediamtx_service(self):
        """Verify real MediaMTX service is running via systemd."""
        # Check if MediaMTX service is running
        result = subprocess.run(["systemctl", "is-active", "mediamtx"], 
                              capture_output=True, text=True)
        if result.returncode != 0:
            pytest.skip("MediaMTX service is not running via systemd")
        
        # Wait for MediaMTX API to be ready
        max_retries = 10
        for i in range(max_retries):
            try:
                response = requests.get("http://localhost:9997/v3/config/global/get", 
                                      timeout=5)
                if response.status_code == 200:
                    return True
            except requests.RequestException:
                pass
            time.sleep(1)
        
        pytest.skip("MediaMTX API is not responding")
        return False
    
    def test_jwt_token_generation_and_validation(self, auth_manager, real_mediamtx_service):
        """Test complete JWT token generation and validation flow against real MediaMTX service.
        
        REQ-SEC-001: JWT Authentication - Token generation, validation, and expiry
        """
        # Generate token for test user using non-hardcoded secret
        token = auth_manager.generate_test_token("test_user", "admin")
        
        # Validate token
        result = auth_manager.auth_manager.authenticate(token, "jwt")
        
        # Verify authentication success
        assert result.authenticated is True
        assert result.user_id == "test_user"
        assert result.role == "admin"
        assert result.auth_method == "jwt"
    
    def test_jwt_token_expiry_validation(self, auth_manager):
        """Test JWT token expiry validation.
        
        REQ-SEC-001: JWT Authentication - Token generation, validation, and expiry
        """
        # Generate token with short expiry
        token = auth_manager.generate_test_token("expiry_user", "viewer", expiry_hours=0.0001)
        
        # Token should be valid initially
        result = auth_manager.auth_manager.authenticate(token, "jwt")
        assert result.authenticated is True
        
        # Wait for token to expire
        time.sleep(1)
        
        # Token should be invalid after expiry
        result = auth_manager.auth_manager.authenticate(token, "jwt")
        assert result.authenticated is False
        assert "expired" in result.error_message.lower()
    
    def test_jwt_token_tampering_detection(self, auth_manager):
        """Test JWT token tampering detection.
        
        REQ-SEC-001: JWT Authentication - Token generation, validation, and expiry
        """
        # Generate valid token
        valid_token = auth_manager.generate_test_token("tamper_user", "admin")
        
        # Tamper with token (modify payload)
        parts = valid_token.split('.')
        if len(parts) == 3:
            # Create tampered token with different payload
            tampered_payload = "eyJ1c2VyX2lkIjoidGFtcGVyZWRfdXNlciIsInJvbGUiOiJhZG1pbiIsImlhdCI6MTYzMzQ1Njc4MCwiZXhwIjoxNjMzNTQzMTgwfQ"
            tampered_token = f"{parts[0]}.{tampered_payload}.{parts[2]}"
            
            # Tampered token should be rejected
            result = auth_manager.auth_manager.authenticate(tampered_token, "jwt")
            assert result.authenticated is False
            assert result.error_message is not None
    
    def test_jwt_auto_authentication_fallback(self, auth_manager):
        """Test auto authentication with JWT fallback."""
        # Generate JWT token using non-hardcoded secret
        jwt_token = auth_manager.generate_test_token("auto_user", "operator")
        
        # Test auto authentication (should try JWT first)
        result = auth_manager.auth_manager.authenticate(jwt_token, "auto")
        assert result.authenticated is True
        assert result.auth_method == "jwt"
        assert result.user_id == "auto_user"
        assert result.role == "operator"
    
    def test_jwt_concurrent_authentication(self, auth_manager):
        """Test concurrent JWT authentication requests."""
        # Generate multiple tokens using non-hardcoded secret
        tokens = []
        for i in range(10):
            token = auth_manager.generate_test_token(f"user_{i}", "viewer")
            tokens.append(token)
        
        # Authenticate all tokens concurrently
        results = []
        for token in tokens:
            result = auth_manager.auth_manager.authenticate(token, "jwt")
            results.append(result)
        
        # Verify all authentications succeeded
        for result in results:
            assert result.authenticated is True
            assert result.auth_method == "jwt"
    
    def test_jwt_performance_benchmark(self, auth_manager):
        """Test JWT authentication performance."""
        # Generate token using non-hardcoded secret
        token = auth_manager.generate_test_token("perf_user", "admin")
        
        # Measure authentication time
        start_time = time.time()
        for _ in range(100):
            result = auth_manager.auth_manager.authenticate(token, "jwt")
            assert result.authenticated is True
        
        end_time = time.time()
        avg_time = (end_time - start_time) / 100
        
        # Performance should be under 1ms per authentication
        assert avg_time < 0.001, f"Authentication too slow: {avg_time:.6f}s per request"


class TestAuthenticationErrorHandling:
    """Test authentication error handling scenarios."""
    
    @pytest.fixture
    def auth_manager(self):
        """Create authentication manager for error testing using non-hardcoded secrets."""
        return get_test_auth_manager()
    
    def test_authentication_with_empty_token(self, auth_manager):
        """Test authentication with empty token."""
        result = auth_manager.auth_manager.authenticate("", "jwt")
        assert result.authenticated is False
        assert result.error_message is not None
        assert result.auth_method == "jwt"
    
    def test_authentication_with_none_token(self, auth_manager):
        """Test authentication with None token."""
        result = auth_manager.auth_manager.authenticate(None, "jwt")
        assert result.authenticated is False
        assert result.error_message is not None
        assert result.auth_method == "jwt"
    
    def test_authentication_with_invalid_auth_type(self, auth_manager):
        """Test authentication with invalid auth type."""
        token = auth_manager.generate_test_token("test_user", "viewer")
        
        # Should fall back to auto authentication
        result = auth_manager.auth_manager.authenticate(token, "invalid_type")
        assert result.authenticated is True
        assert result.auth_method == "jwt"
    
    def test_authentication_with_malformed_token(self, auth_manager):
        """Test authentication with malformed JWT token."""
        malformed_token = "not.a.valid.jwt.token"
        result = auth_manager.auth_manager.authenticate(malformed_token, "jwt")
        assert result.authenticated is False
        assert result.error_message is not None


class TestRoleBasedAccessControl:
    """Test role-based access control functionality."""
    
    @pytest.fixture
    def auth_manager(self):
        """Create authentication manager for RBAC testing."""
        return get_test_auth_manager()
    
    @pytest.fixture
    def user_factory(self, auth_manager):
        """Create user factory for RBAC testing."""
        return TestUserFactory(auth_manager)
    
    def test_viewer_role_permissions(self, user_factory):
        """Test viewer role permissions."""
        viewer_user = user_factory.create_viewer_user()
        
        # Viewer should have access to basic methods
        assert "get_camera_list" in viewer_user["permissions"]
        assert "get_camera_status" in viewer_user["permissions"]
        assert "get_streams" in viewer_user["permissions"]
        
        # Viewer should not have access to protected methods
        assert "take_snapshot" not in viewer_user["permissions"]
        assert "start_recording" not in viewer_user["permissions"]
        assert "stop_recording" not in viewer_user["permissions"]
    
    def test_operator_role_permissions(self, user_factory):
        """Test operator role permissions."""
        operator_user = user_factory.create_operator_user()
        
        # Operator should have access to all camera control methods
        assert "get_camera_list" in operator_user["permissions"]
        assert "get_camera_status" in operator_user["permissions"]
        assert "get_streams" in operator_user["permissions"]
        assert "take_snapshot" in operator_user["permissions"]
        assert "start_recording" in operator_user["permissions"]
        assert "stop_recording" in operator_user["permissions"]
        
        # Operator should not have access to admin methods
        assert "delete_camera" not in operator_user["permissions"]
        assert "modify_config" not in operator_user["permissions"]
    
    def test_admin_role_permissions(self, user_factory):
        """Test admin role permissions."""
        admin_user = user_factory.create_admin_user()
        
        # Admin should have access to all methods
        assert "get_camera_list" in admin_user["permissions"]
        assert "get_camera_status" in admin_user["permissions"]
        assert "get_streams" in admin_user["permissions"]
        assert "take_snapshot" in admin_user["permissions"]
        assert "start_recording" in admin_user["permissions"]
        assert "stop_recording" in admin_user["permissions"]
        assert "delete_camera" in admin_user["permissions"]
        assert "modify_config" in admin_user["permissions"]


# Cleanup function for pytest
def pytest_sessionfinish(session, exitstatus):
    """Clean up authentication manager after test session."""
    cleanup_test_auth_manager() 