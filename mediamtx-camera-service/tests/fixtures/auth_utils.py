"""
Real Authentication Test Infrastructure

Provides real authentication testing infrastructure without mocking.
Follows testing guide principles: use real JWT tokens, real authentication flows.

Requirements Traceability:
- REQ-AUTH-001: System shall provide secure authentication using JWT tokens
- REQ-AUTH-002: System shall validate token expiration and renewal
- REQ-AUTH-003: System shall handle authentication failures gracefully
- REQ-AUTH-004: System shall support role-based access control

Test Categories: Integration
"""

import os
import time
import jwt
from typing import Dict, Any, Optional

import pytest
import pytest_asyncio

from src.security.auth_manager import AuthManager


# Standalone functions for direct import
def get_test_jwt_secret() -> str:
    """Get the test JWT secret from environment or use default."""
    return os.getenv('CAMERA_SERVICE_JWT_SECRET', 'test-secret-dev-only')


def generate_valid_test_token(username: str = "test_user", role: str = "operator") -> str:
    """Generate a valid test JWT token."""
    secret = get_test_jwt_secret()
    payload = {
        "user_id": username,
        "role": role,
        "exp": time.time() + 3600,  # 1 hour expiration
        "iat": time.time()
    }
    return jwt.encode(payload, secret, algorithm="HS256")


def generate_expired_test_token() -> str:
    """Generate an expired test JWT token."""
    secret = get_test_jwt_secret()
    payload = {
        "user_id": "expired_user",
        "role": "operator",
        "exp": time.time() - 3600,  # Expired 1 hour ago
        "iat": time.time() - 7200
    }
    return jwt.encode(payload, secret, algorithm="HS256")


def generate_invalid_test_token() -> str:
    """Generate an invalid test JWT token."""
    return "invalid.jwt.token"


class RealAuthTestBase:
    """Base class for real authentication testing.
    
    Provides real JWT tokens and authentication testing without mocking.
    Follows testing guide: NEVER mock JWT authentication.
    """
    
    @pytest_asyncio.fixture(autouse=True)
    async def real_jwt_tokens(self):
        """Generate real JWT tokens for testing.
        
        Requirements: REQ-AUTH-001, REQ-AUTH-002
        """
        # Use real JWT secret from environment or test secret
        secret = os.getenv('CAMERA_SERVICE_JWT_SECRET', 'test-secret-dev-only')
        
        # Generate valid tokens with different roles
        valid_tokens = {
            "admin": self._generate_valid_token("admin_user", "admin", secret),
            "operator": self._generate_valid_token("operator_user", "operator", secret),
            "viewer": self._generate_valid_token("viewer_user", "viewer", secret)
        }
        
        # Generate expired token
        expired_token = self._generate_expired_token("expired_user", "operator", secret)
        
        # Generate invalid token
        invalid_token = "invalid.jwt.token"
        
        yield {
            "valid": valid_tokens,
            "expired": expired_token,
            "invalid": invalid_token,
            "secret": secret
        }
    
    @pytest_asyncio.fixture(autouse=True)
    async def real_auth_manager(self):
        """Real authentication manager for testing.
        
        Requirements: REQ-AUTH-003, REQ-AUTH-004
        """
        # Create real authentication manager with handlers
        from src.security.jwt_handler import JWTHandler
        from src.security.api_key_handler import APIKeyHandler
        
        jwt_handler = JWTHandler(secret_key="test-secret-dev-only")
        api_key_handler = APIKeyHandler(storage_file="/tmp/test_api_keys.json")
        auth_manager = AuthManager(jwt_handler, api_key_handler)
        
        yield auth_manager
    
    def _generate_valid_token(self, username: str, role: str, secret: str) -> str:
        """Generate a valid JWT token."""
        payload = {
            "user_id": username,  # Use user_id field as expected by JWT handler
            "role": role,
            "exp": time.time() + 3600,  # 1 hour expiration
            "iat": time.time()
        }
        return jwt.encode(payload, secret, algorithm="HS256")
    
    def _generate_expired_token(self, username: str, role: str, secret: str) -> str:
        """Generate an expired JWT token."""
        payload = {
            "user_id": username,  # Use user_id field as expected by JWT handler
            "role": role,
            "exp": time.time() - 3600,  # Expired 1 hour ago
            "iat": time.time() - 7200
        }
        return jwt.encode(payload, secret, algorithm="HS256")
    
    def get_test_jwt_secret(self) -> str:
        """Get the test JWT secret from environment or use default."""
        return os.getenv('CAMERA_SERVICE_JWT_SECRET', 'test-secret-dev-only')
    
    def generate_valid_test_token(self, username: str = "test_user", role: str = "operator") -> str:
        """Generate a valid test JWT token."""
        secret = self.get_test_jwt_secret()
        return self._generate_valid_token(username, role, secret)
    
    def generate_expired_test_token(self) -> str:
        """Generate an expired test JWT token."""
        secret = self.get_test_jwt_secret()
        return self._generate_expired_token("expired_user", "operator", secret)
    
    def generate_invalid_test_token(self) -> str:
        """Generate an invalid test JWT token."""
        return "invalid.jwt.token"
    
    async def validate_token_authentication(self, auth_manager: AuthManager, token: str, expected_valid: bool = True):
        """Validate token authentication with real auth manager."""
        try:
            result = await auth_manager.authenticate(token)
            if expected_valid:
                assert result["authenticated"] is True, f"Token should be valid: {token}"
                assert "user" in result, "Authentication result should contain user info"
            else:
                pytest.fail(f"Token should be invalid: {token}")
        except Exception as e:
            if expected_valid:
                pytest.fail(f"Valid token authentication failed: {e}")
            # Expected failure for invalid tokens
    
    async def validate_role_based_access(self, auth_manager: AuthManager, token: str, required_role: str, should_have_access: bool = True):
        """Validate role-based access control."""
        try:
            result = await auth_manager.authenticate(token)
            if should_have_access:
                assert result["authenticated"] is True, f"User should have access with role {required_role}"
                assert result["user"]["role"] == required_role, f"User should have role {required_role}"
            else:
                assert not result["authenticated"] or result["user"]["role"] != required_role, f"User should not have access with role {required_role}"
        except Exception as e:
            if should_have_access:
                pytest.fail(f"Role-based access validation failed: {e}")


class WebSocketAuthTestClient:
    """Real WebSocket client for authentication testing."""
    
    def __init__(self, websocket_url: str, auth_manager: AuthManager):
        self.websocket_url = websocket_url
        self.auth_manager = auth_manager
        self.websocket = None
        self.request_id = 1
    
    async def connect(self):
        """Connect to WebSocket server."""
        import websockets
        self.websocket = await websockets.connect(self.websocket_url)
    
    async def disconnect(self):
        """Disconnect from WebSocket server."""
        if self.websocket:
            await self.websocket.close()
            self.websocket = None
    
    async def authenticate(self, token: str) -> Dict[str, Any]:
        """Authenticate with WebSocket server using real JWT token."""
        if not self.websocket:
            raise ConnectionError("WebSocket not connected")
        
        request = {
            "jsonrpc": "2.0",
            "method": "authenticate",
            "params": {"token": token},
            "id": self.request_id
        }
        
        self.request_id += 1
        
        import json
        await self.websocket.send(json.dumps(request))
        response = await self.websocket.recv()
        return json.loads(response)
    
    async def call_protected_method(self, method: str, params: Optional[Dict] = None) -> Dict[str, Any]:
        """Call a protected method on WebSocket server."""
        if not self.websocket:
            raise ConnectionError("WebSocket not connected")
        
        request = {
            "jsonrpc": "2.0",
            "method": method,
            "id": self.request_id
        }
        
        if params:
            request["params"] = params
        
        self.request_id += 1
        
        import json
        await self.websocket.send(json.dumps(request))
        response = await self.websocket.recv()
        return json.loads(response)
