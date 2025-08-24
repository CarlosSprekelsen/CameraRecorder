"""
WebSocket server binding and connectivity unit tests.

Requirements Coverage:
- REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint at ws://localhost:8002/ws
- REQ-API-002: ping method for health checks

Test Categories: Unit
API Documentation Reference: docs/api/json-rpc-methods.md
"""

import json
import pytest
import os

from src.websocket_server.server import WebSocketJsonRpcServer
from tests.fixtures.auth_utils import get_test_auth_manager, UserFactory


@pytest.mark.asyncio
async def test_websocket_server_binds_and_ping(monkeypatch):
    """Test WebSocket server binding and ping with authentication."""
    # Set up test environment
    os.environ['CAMERA_SERVICE_JWT_SECRET'] = 'test-secret-dev-only'
    
    # Create auth manager and test user
    auth_manager = get_test_auth_manager()
    user_factory = UserFactory(auth_manager)
    test_user = user_factory.create_operator_user("websocket_bind_test_user")
    
    server = WebSocketJsonRpcServer(
        host="127.0.0.1", port=8022, websocket_path="/ws", max_connections=5
    )
    
    # Configure security middleware like ServiceManager does
    from src.security.jwt_handler import JWTHandler
    from src.security.api_key_handler import APIKeyHandler
    from src.security.auth_manager import AuthManager
    from src.security.middleware import SecurityMiddleware
    
    jwt_handler = JWTHandler(secret_key='test-secret-dev-only')
    api_key_handler = APIKeyHandler(storage_file="/tmp/test_api_keys.json")
    auth_manager = AuthManager(jwt_handler=jwt_handler, api_key_handler=api_key_handler)
    security = SecurityMiddleware(
        auth_manager=auth_manager,
        max_connections=5,
        requests_per_minute=120,
    )
    
    if hasattr(server, "set_security_middleware"):
        server.set_security_middleware(security)
    
    await server.start()
    try:
        import websockets

        uri = "ws://127.0.0.1:8022/ws"
        async with websockets.connect(uri) as ws:
            # Send ping with authentication
            await ws.send(json.dumps({
                "jsonrpc": "2.0", 
                "id": 1, 
                "method": "ping", 
                "params": {"auth_token": test_user["token"]}
            }))
            resp = json.loads(await ws.recv())
            assert resp["result"] == "pong"
    finally:
        await server.stop()


@pytest.mark.asyncio
async def test_websocket_server_ping_requires_authentication(monkeypatch):
    """Test that ping method requires authentication according to API documentation."""
    # Set up test environment
    os.environ['CAMERA_SERVICE_JWT_SECRET'] = 'test-secret-dev-only'
    
    server = WebSocketJsonRpcServer(
        host="127.0.0.1", port=8023, websocket_path="/ws", max_connections=5
    )
    
    # Configure security middleware like ServiceManager does
    from src.security.jwt_handler import JWTHandler
    from src.security.api_key_handler import APIKeyHandler
    from src.security.auth_manager import AuthManager
    from src.security.middleware import SecurityMiddleware
    
    jwt_handler = JWTHandler(secret_key='test-secret-dev-only')
    api_key_handler = APIKeyHandler(storage_file="/tmp/test_api_keys.json")
    auth_manager = AuthManager(jwt_handler=jwt_handler, api_key_handler=api_key_handler)
    security = SecurityMiddleware(
        auth_manager=auth_manager,
        max_connections=5,
        requests_per_minute=120,
    )
    
    if hasattr(server, "set_security_middleware"):
        server.set_security_middleware(security)
    
    await server.start()
    try:
        import websockets

        uri = "ws://127.0.0.1:8023/ws"
        async with websockets.connect(uri) as ws:
            # Send ping without authentication (should fail)
            await ws.send(json.dumps({
                "jsonrpc": "2.0", 
                "id": 1, 
                "method": "ping", 
                "params": {}
            }))
            resp = json.loads(await ws.recv())
            
            # According to API documentation, this should return an authentication error
            assert "error" in resp, "Should return error for unauthenticated request"
            assert resp["error"]["code"] == -32001, "Should return authentication error code"
            assert "Authentication required" in resp["error"]["message"], "Should indicate authentication is required"
    finally:
        await server.stop()

