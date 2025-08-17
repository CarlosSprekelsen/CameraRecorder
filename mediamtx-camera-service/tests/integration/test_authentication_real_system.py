"""
Real system authentication integration test.

Requirements Traceability:
- REQ-AUTH-001: Authentication shall work with real JWT tokens
- REQ-AUTH-002: Authentication shall work with real API keys
- REQ-AUTH-003: Authentication shall reject invalid tokens

Story Coverage: S6 - Security Features Implementation
IV&V Control Point: Authentication validation
"""

import pytest
import pytest_asyncio
import asyncio
import json
import jwt
import websockets
import os
import tempfile
from pathlib import Path
from typing import Dict, Any
import sys

# Add src to path for imports
sys.path.insert(0, str(Path(__file__).parent.parent.parent / "src"))

from security.jwt_handler import JWTHandler
from security.api_key_handler import APIKeyHandler
from security.auth_manager import AuthManager


@pytest.mark.integration
class TestAuthenticationRealSystem:
    """Real system authentication integration tests."""
    
    @pytest.fixture
    def jwt_secret(self):
        """JWT secret key for testing."""
        return "test-secret-key-for-integration-testing"
    
    @pytest.fixture
    def jwt_handler(self, jwt_secret):
        """JWT handler for testing."""
        return JWTHandler(jwt_secret)
    
    @pytest.fixture
    def api_keys_file(self):
        """Temporary API keys file."""
        with tempfile.NamedTemporaryFile(mode='w', delete=False, suffix='.json') as f:
            f.write('{"version": "1.0", "keys": []}')
            temp_file = f.name
        
        yield temp_file
        
        # Cleanup
        if os.path.exists(temp_file):
            os.unlink(temp_file)
    
    @pytest.fixture
    def api_key_handler(self, api_keys_file):
        """API key handler for testing."""
        return APIKeyHandler(api_keys_file)
    
    @pytest.fixture
    def auth_manager(self, jwt_handler, api_key_handler):
        """Authentication manager for testing."""
        return AuthManager(jwt_handler, api_key_handler)
    
    @pytest_asyncio.fixture
    async def websocket_client(self):
        """WebSocket client for testing."""
        client = WebSocketTestClient("ws://localhost:8080/ws")
        yield client
        await client.disconnect()
    
    def test_jwt_token_generation_and_validation(self, jwt_handler):
        """Test JWT token generation and validation."""
        # Generate token
        user_id = "test_user_123"
        role = "admin"
        token = jwt_handler.generate_token(user_id, role)
        
        # Validate token
        claims = jwt_handler.validate_token(token)
        
        assert claims is not None
        assert claims.user_id == user_id
        assert claims.role == role
    
    def test_api_key_generation_and_validation(self, api_key_handler):
        """Test API key generation and validation."""
        # Generate API key
        key_name = "test_key"
        role = "operator"
        api_key = api_key_handler.create_api_key(key_name, role)
        
        # Validate API key
        key_info = api_key_handler.validate_api_key(api_key)
        
        assert key_info is not None
        assert key_info["name"] == key_name
        assert key_info["role"] == role
    
    def test_auth_manager_jwt_authentication(self, auth_manager, jwt_handler):
        """Test authentication manager with JWT."""
        # Generate token
        token = jwt_handler.generate_token("test_user", "admin")
        
        # Authenticate
        result = auth_manager.authenticate(token, "jwt")
        
        assert result.authenticated is True
        assert result.user_id == "test_user"
        assert result.role == "admin"
        assert result.auth_method == "jwt"
    
    def test_auth_manager_api_key_authentication(self, auth_manager, api_key_handler):
        """Test authentication manager with API key."""
        # Generate API key
        api_key = api_key_handler.create_api_key("test_key", "operator")
        
        # Authenticate
        result = auth_manager.authenticate(api_key, "api_key")
        
        assert result.authenticated is True
        assert result.role == "operator"
        assert result.auth_method == "api_key"
    
    def test_auth_manager_invalid_token(self, auth_manager):
        """Test authentication manager with invalid token."""
        result = auth_manager.authenticate("invalid_token", "jwt")
        
        assert result.authenticated is False
        assert result.error_message is not None
    
    @pytest.mark.asyncio
    async def test_websocket_authentication_flow(self, websocket_client, jwt_handler):
        """Test WebSocket authentication flow."""
        # Connect to WebSocket
        await websocket_client.connect()
        
        # Generate valid JWT token
        token = jwt_handler.generate_token("websocket_test_user", "admin")
        
        # Authenticate
        success = await websocket_client.authenticate(token)
        
        assert success is True
        assert websocket_client.authenticated is True
    
    @pytest.mark.asyncio
    async def test_websocket_authentication_failure(self, websocket_client):
        """Test WebSocket authentication failure."""
        # Connect to WebSocket
        await websocket_client.connect()
        
        # Try to authenticate with invalid token
        success = await websocket_client.authenticate("invalid_token")
        
        assert success is False
        assert websocket_client.authenticated is False


class WebSocketTestClient:
    """Test client for WebSocket JSON-RPC communication."""
    
    def __init__(self, websocket_url: str):
        self.websocket_url = websocket_url
        self.websocket = None
        self.authenticated = False
    
    async def connect(self) -> None:
        """Connect to WebSocket server."""
        self.websocket = await websockets.connect(self.websocket_url)
    
    async def disconnect(self) -> None:
        """Disconnect from WebSocket server."""
        if self.websocket:
            await self.websocket.close()
            self.websocket = None
    
    async def authenticate(self, auth_token: str) -> bool:
        """Authenticate with the WebSocket server."""
        if not self.websocket:
            raise RuntimeError("WebSocket not connected")
        
        response = await self.send_request("authenticate", {"token": auth_token})
        
        if "result" in response and response["result"].get("authenticated"):
            self.authenticated = True
            return True
        else:
            return False
    
    async def send_request(self, method: str, params: Dict = None, request_id: int = 1) -> Dict:
        """Send JSON-RPC request."""
        if not self.websocket:
            raise RuntimeError("WebSocket not connected")
        
        request = {
            "jsonrpc": "2.0",
            "method": method,
            "id": request_id
        }
        
        if params:
            request["params"] = params
        
        await self.websocket.send(json.dumps(request))
        response = await self.websocket.recv()
        return json.loads(response)
