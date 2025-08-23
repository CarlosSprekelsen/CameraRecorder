"""
Real Authentication Test Infrastructure

Provides real authentication testing infrastructure without mocking.
Follows testing guide principles: use real JWT tokens, real authentication flows.

Requirements Coverage:
- REQ-SEC-001: JWT token-based authentication for all API access
- REQ-SEC-002: Token Format: JSON Web Token (JWT) with standard claims
- REQ-SEC-003: Token Expiration: Configurable expiration time (default: 24 hours)
- REQ-SEC-004: Token Refresh: Support for token refresh mechanism
- REQ-SEC-005: Token Validation: Proper signature validation and claim verification

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


def get_test_auth_manager() -> AuthManager:
    """Get a test authentication manager instance.
    
    Creates a real AuthManager with test JWT secret for integration testing.
    Follows testing guide: use real authentication, never mock.
    """
    from src.security.jwt_handler import JWTHandler
    from src.security.api_key_handler import APIKeyHandler
    from src.security.middleware import SecurityMiddleware
    
    # Create real JWT handler with test secret
    jwt_secret = get_test_jwt_secret()
    jwt_handler = JWTHandler(jwt_secret)
    
    # Create real API key handler with test storage file
    import tempfile
    import os
    test_api_keys_file = os.path.join(tempfile.gettempdir(), "test_api_keys.json")
    api_key_handler = APIKeyHandler(storage_file=test_api_keys_file)
    
    # Create and return real auth manager
    auth_manager = AuthManager(jwt_handler, api_key_handler)
    return auth_manager


def cleanup_test_auth_manager(auth_manager: AuthManager):
    """Cleanup test authentication manager.
    
    Cleans up any test data or resources used by the auth manager.
    """
    if auth_manager and hasattr(auth_manager, 'api_key_handler'):
        # Clean up any test API keys
        if hasattr(auth_manager.api_key_handler, 'cleanup_expired_keys'):
            auth_manager.api_key_handler.cleanup_expired_keys()


class TestUserFactory:
    """Factory for creating test users with different roles and permissions.
    
    Provides convenient methods to create test users for integration testing.
    """
    
    def __init__(self, auth_manager: AuthManager):
        self.auth_manager = auth_manager
        self.user_counter = 0
    
    def create_admin_user(self, username: str = None) -> Dict[str, Any]:
        """Create a test admin user with admin privileges."""
        if username is None:
            username = f"admin_user_{self.user_counter}"
            self.user_counter += 1
        
        token = generate_valid_test_token(username, "admin")
        return {
            "username": username,
            "user_id": username,
            "role": "admin",
            "token": token,
            "permissions": [
                "get_camera_list", "get_camera_status", "get_streams",
                "take_snapshot", "start_recording", "stop_recording",
                "list_recordings", "list_snapshots", "get_metrics",
                "get_status", "get_server_info", "delete_recording",
                "get_recording_info", "get_snapshot_info"
            ]
        }
    
    def create_operator_user(self, username: str = None) -> Dict[str, Any]:
        """Create a test operator user with operator privileges."""
        if username is None:
            username = f"operator_user_{self.user_counter}"
            self.user_counter += 1
        
        token = generate_valid_test_token(username, "operator")
        return {
            "username": username,
            "user_id": username,
            "role": "operator",
            "token": token,
            "permissions": [
                "get_camera_list", "get_camera_status", "get_streams",
                "take_snapshot", "start_recording", "stop_recording",
                "list_recordings", "list_snapshots"
            ]
        }
    
    def create_viewer_user(self, username: str = None) -> Dict[str, Any]:
        """Create a test viewer user with viewer privileges."""
        if username is None:
            username = f"viewer_user_{self.user_counter}"
            self.user_counter += 1
        
        token = generate_valid_test_token(username, "viewer")
        return {
            "username": username,
            "user_id": username,
            "role": "viewer",
            "token": token,
            "permissions": [
                "get_camera_list", "get_camera_status", "get_streams",
                "list_recordings", "list_snapshots"
            ]
        }
    
    def create_expired_user(self, username: str = None) -> Dict[str, Any]:
        """Create a test user with expired token."""
        if username is None:
            username = f"expired_user_{self.user_counter}"
            self.user_counter += 1
        
        token = generate_expired_test_token()
        return {
            "username": username,
            "role": "operator",
            "token": token,
            "permissions": []
        }
    
    def create_invalid_user(self, username: str = None) -> Dict[str, Any]:
        """Create a test user with invalid token."""
        if username is None:
            username = f"invalid_user_{self.user_counter}"
            self.user_counter += 1
        
        token = generate_invalid_test_token()
        return {
            "username": username,
            "role": "unknown",
            "token": token,
            "permissions": []
        }
    
    def create_user_with_role(self, role: str, username: str = None) -> Dict[str, Any]:
        """Create a test user with specified role."""
        if role == "admin":
            return self.create_admin_user(username)
        elif role == "operator":
            return self.create_operator_user(username)
        elif role == "viewer":
            return self.create_viewer_user(username)
        else:
            raise ValueError(f"Unknown role: {role}")


class RealAuthTestBase:
    """Base class for real authentication testing.
    
    Provides real JWT tokens and authentication testing without mocking.
    Follows testing guide: NEVER mock JWT authentication.
    """
    
    @pytest_asyncio.fixture(autouse=True)
    async def real_jwt_tokens(self):
        """Generate real JWT tokens for testing."""
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
        """Real authentication manager for testing."""
        # Create real authentication manager with handlers
        from src.security.jwt_handler import JWTHandler
        from src.security.api_key_handler import APIKeyHandler
        from src.security.middleware import SecurityMiddleware
        
        # Use real JWT secret
        secret = os.getenv('CAMERA_SERVICE_JWT_SECRET', 'test-secret-dev-only')
        
        # Create real handlers
        jwt_handler = JWTHandler(secret)
        api_key_handler = APIKeyHandler()
        security_middleware = SecurityMiddleware(jwt_handler, api_key_handler)
        
        # Create real auth manager
        auth_manager = AuthManager(jwt_handler, api_key_handler, security_middleware)
        
        yield auth_manager
        
        # Cleanup
        if hasattr(api_key_handler, 'cleanup_expired_keys'):
            api_key_handler.cleanup_expired_keys()
    
    def _generate_valid_token(self, username: str, role: str, secret: str) -> str:
        """Generate a valid JWT token."""
        payload = {
            "user_id": username,
            "role": role,
            "exp": time.time() + 3600,  # 1 hour expiration
            "iat": time.time()
        }
        return jwt.encode(payload, secret, algorithm="HS256")
    
    def _generate_expired_token(self, username: str, role: str, secret: str) -> str:
        """Generate an expired JWT token."""
        payload = {
            "user_id": username,
            "role": role,
            "exp": time.time() - 3600,  # Expired 1 hour ago
            "iat": time.time() - 7200
        }
        return jwt.encode(payload, secret, algorithm="HS256")
    
    def get_test_jwt_secret(self) -> str:
        """Get the test JWT secret."""
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
    
    def __init__(self, websocket_url: str, test_user: Dict[str, Any]):
        self.websocket_url = websocket_url
        self.test_user = test_user
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
    
    async def authenticate(self, token: str = None) -> Dict[str, Any]:
        """Authenticate with WebSocket server using real JWT token."""
        if not self.websocket:
            raise ConnectionError("WebSocket not connected")
        
        # Use test user's token if no token provided
        if token is None:
            token = self.test_user["token"]
        
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
    
    async def send_request(self, method: str, params: Optional[Dict] = None) -> Dict[str, Any]:
        """Send JSON-RPC request with authentication."""
        if not self.websocket:
            raise ConnectionError("WebSocket not connected")
        
        # Include authentication token in params
        if params is None:
            params = {}
        
        # Add auth token to params
        params["auth_token"] = self.test_user["token"]
        
        request = {
            "jsonrpc": "2.0",
            "method": method,
            "params": params,
            "id": self.request_id
        }
        
        self.request_id += 1
        
        import json
        await self.websocket.send(json.dumps(request))
        response = await self.websocket.recv()
        return json.loads(response)
    
    async def call_protected_method(self, method: str, params: Optional[Dict] = None) -> Dict[str, Any]:
        """Call a protected method on WebSocket server."""
        return await self.send_request(method, params)
    
    async def send_unauthenticated_request(self, method: str, params: Optional[Dict] = None) -> Dict[str, Any]:
        """Send JSON-RPC request without authentication for negative testing."""
        if not self.websocket:
            raise ConnectionError("WebSocket not connected")
        
        # Don't include authentication token in params
        if params is None:
            params = {}
        
        request = {
            "jsonrpc": "2.0",
            "method": method,
            "params": params,
            "id": self.request_id
        }
        
        self.request_id += 1
        
        import json
        await self.websocket.send(json.dumps(request))
        response = await self.websocket.recv()
        return json.loads(response)
