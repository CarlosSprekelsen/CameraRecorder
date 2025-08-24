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
- REQ-REC-001.2: Error handling for recording conflicts
- REQ-REC-003.3: Storage error handling

Test Categories: Integration
API Documentation Reference: docs/api/json-rpc-methods.md
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
    Updated for new API structure with enhanced role-based access control.
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
                "get_recording_info", "get_snapshot_info", "get_storage_info"
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
                "list_recordings", "list_snapshots", "get_storage_info"
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
                "list_recordings", "list_snapshots", "get_storage_info"
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
    Updated for new API structure with enhanced error codes.
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
    """Real WebSocket client for authentication testing.
    
    Updated for new API structure with enhanced error codes and response validation.
    """
    
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
            "params": {"auth_token": token},
            "id": self.request_id
        }
        
        self.request_id += 1
        
        import json
        await self.websocket.send(json.dumps(request))
        response = await self.websocket.recv()
        return json.loads(response)
    
    async def authenticate_and_validate_api_compliance(self, token: str = None) -> Dict[str, Any]:
        """
        Authenticate and validate response against API documentation.
        
        This method would have caught Issue 081: Authentication Method Documentation vs Implementation Mismatch.
        Validates the authenticate method response format matches docs/api/json-rpc-methods.md exactly.
        
        API Documentation Reference: docs/api/json-rpc-methods.md - authenticate method
        Updated for new API structure with enhanced response validation.
        """
        response = await self.authenticate(token)
        
        # Validate JSON-RPC 2.0 structure
        assert "jsonrpc" in response, "Response must contain 'jsonrpc' field per JSON-RPC 2.0"
        assert response["jsonrpc"] == "2.0", "JSON-RPC version must be 2.0"
        assert "id" in response, "Response must contain 'id' field per JSON-RPC 2.0"
        
        # Must have either result or error (not both)
        assert ("result" in response) != ("error" in response), "Response must have either 'result' or 'error', not both"
        
        if "result" in response:
            # Validate documented response format from API documentation
            result = response["result"]
            
            # All fields required by API documentation
            assert "authenticated" in result, "Missing 'authenticated' field per API documentation"
            assert "role" in result, "Missing 'role' field per API documentation" 
            assert "permissions" in result, "Missing 'permissions' field per API documentation"
            assert "expires_at" in result, "Missing 'expires_at' field per API documentation"
            assert "session_id" in result, "Missing 'session_id' field per API documentation"
            
            # Validate field types per API documentation
            assert isinstance(result["authenticated"], bool), "authenticated must be boolean"
            assert isinstance(result["role"], str), "role must be string"
            assert isinstance(result["permissions"], list), "permissions must be list per API documentation"
            assert isinstance(result["expires_at"], str), "expires_at must be string (ISO format)"
            assert isinstance(result["session_id"], str), "session_id must be string"
            
            # Validate role values per API documentation
            valid_roles = ["viewer", "operator", "admin"]
            assert result["role"] in valid_roles, f"Role must be one of {valid_roles} per API documentation"
            
        elif "error" in response:
            # Validate documented error format
            error = response["error"]
            assert "code" in error, "Error must contain 'code' field"
            assert "message" in error, "Error must contain 'message' field"
            assert isinstance(error["code"], int), "Error code must be integer"
            assert isinstance(error["message"], str), "Error message must be string"
            
            # Validate error codes per API documentation
            valid_error_codes = [-32700, -32600, -32601, -32602, -32603, -32001, -32002, -32003, -32004, -32005, -32006, -32007, -32008, -1006, -1008, -1010]
            assert error["code"] in valid_error_codes, f"Error code {error['code']} not in valid codes: {valid_error_codes}"
        
        return response
    
    async def send_request(self, method: str, params: Optional[Dict] = None) -> Dict[str, Any]:
        """Send JSON-RPC request with authentication."""
        if not self.websocket:
            raise ConnectionError("WebSocket not connected")
        
        # Include authentication token in params
        if params is None:
            params = {}
        
        # Add auth token to params (new API requirement)
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
    
    def validate_error_response(self, response: Dict[str, Any], expected_code: int, expected_message: str = None):
        """
        Validate error response against new API error codes.
        
        Updated for new recording management error codes:
        - -1006: "Camera is currently recording" (recording conflict)
        - -1008: "Storage space is low" (below 10% available)
        - -1010: "Storage space is critical" (below 5% available)
        """
        assert "error" in response, "Response must contain 'error' field"
        error = response["error"]
        
        assert "code" in error, "Error must contain 'code' field"
        assert "message" in error, "Error must contain 'message' field"
        assert error["code"] == expected_code, f"Expected error code {expected_code}, got {error['code']}"
        
        if expected_message:
            assert expected_message in error["message"], f"Expected message containing '{expected_message}', got '{error['message']}'"
        
        # Validate error code is in documented range
        valid_error_codes = [-32700, -32600, -32601, -32602, -32603, -32001, -32002, -32003, -32004, -32005, -32006, -32007, -32008, -1006, -1008, -1010]
        assert error["code"] in valid_error_codes, f"Error code {error['code']} not in valid codes: {valid_error_codes}"
    
    def validate_recording_conflict_error(self, response: Dict[str, Any]):
        """Validate recording conflict error (-1006)."""
        self.validate_error_response(response, -1006, "Camera is currently recording")
    
    def validate_storage_error(self, response: Dict[str, Any], expected_code: int):
        """Validate storage error codes (-1008, -1010)."""
        if expected_code == -1008:
            self.validate_error_response(response, -1008, "Storage space is low")
        elif expected_code == -1010:
            self.validate_error_response(response, -1010, "Storage space is critical")
        else:
            raise ValueError(f"Invalid storage error code: {expected_code}")
    
    def validate_camera_status_response(self, response: Dict[str, Any]):
        """
        Validate enhanced camera status response format.
        
        Updated for new API structure with recording information.
        """
        assert "result" in response, "Response must contain 'result' field"
        result = response["result"]
        
        # Required fields per API documentation
        required_fields = ["camera_id", "device", "status", "recording"]
        for field in required_fields:
            assert field in result, f"Missing required field '{field}' in camera status response"
        
        # Validate recording-related fields
        if result.get("recording", False):
            assert "recording_session" in result, "Recording session must be present when recording is true"
            assert "current_file" in result, "Current file must be present when recording is true"
            assert "elapsed_time" in result, "Elapsed time must be present when recording is true"
    
    def validate_recording_response(self, response: Dict[str, Any]):
        """
        Validate recording response with new fields.
        
        Updated for new API structure with enhanced metadata.
        """
        assert "result" in response, "Response must contain 'result' field"
        result = response["result"]
        
        # Required fields per API documentation
        required_fields = ["device", "session_id", "filename", "status", "start_time", "duration", "format"]
        for field in required_fields:
            assert field in result, f"Missing required field '{field}' in recording response"
        
        # Validate field types
        assert isinstance(result["device"], str), "device must be string"
        assert isinstance(result["session_id"], str), "session_id must be string"
        assert isinstance(result["filename"], str), "filename must be string"
        assert isinstance(result["status"], str), "status must be string"
        assert isinstance(result["start_time"], str), "start_time must be string (ISO format)"
        assert isinstance(result["duration"], (int, float)), "duration must be number"
        assert isinstance(result["format"], str), "format must be string"
    
    def validate_storage_info_response(self, response: Dict[str, Any]):
        """
        Validate storage information response.
        
        Updated for new API structure with threshold information.
        """
        assert "result" in response, "Response must contain 'result' field"
        result = response["result"]
        
        # Required fields per API documentation
        required_fields = ["total_space", "used_space", "available_space", "usage_percent"]
        for field in required_fields:
            assert field in result, f"Missing required field '{field}' in storage info response"
        
        # Validate field types
        assert isinstance(result["total_space"], (int, float)), "total_space must be number"
        assert isinstance(result["used_space"], (int, float)), "used_space must be number"
        assert isinstance(result["available_space"], (int, float)), "available_space must be number"
        assert isinstance(result["usage_percent"], (int, float)), "usage_percent must be number"
        
        # Validate logical constraints
        assert result["total_space"] >= 0, "total_space must be non-negative"
        assert result["used_space"] >= 0, "used_space must be non-negative"
        assert result["available_space"] >= 0, "available_space must be non-negative"
        assert 0 <= result["usage_percent"] <= 100, "usage_percent must be between 0 and 100"
