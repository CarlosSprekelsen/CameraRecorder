# tests/unit/test_websocket_server/test_server_real_connections_simple.py
"""
Simplified real connection test for debugging.
"""

import asyncio
import json
import pytest
import socket

import websockets

from src.websocket_server.server import WebSocketJsonRpcServer


@pytest.mark.asyncio
async def test_simple_real_connection():
    """Test simple real WebSocket connection."""
    # Get random port
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
        s.bind(('', 0))
        s.listen(1)
        port = s.getsockname()[1]
    
    # Create and start server
    server = WebSocketJsonRpcServer(
        host="127.0.0.1",
        port=port,
        websocket_path="/ws",
        max_connections=10
    )
    
    try:
        await server.start()
        
        # Connect to server
        uri = f"ws://127.0.0.1:{port}/ws"
        async with websockets.connect(uri) as websocket:
            # Test ping
            await websocket.send(json.dumps({
                "jsonrpc": "2.0",
                "method": "ping",
                "id": 1
            }))
            
            response = await websocket.recv()
            result = json.loads(response)
            
            assert result["jsonrpc"] == "2.0"
            assert result["result"] == "pong"
            assert result["id"] == 1
            
    finally:
        await server.stop()


@pytest.mark.asyncio
async def test_real_connection_with_security():
    """Test real connection with security middleware."""
    # Get random port
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
        s.bind(('', 0))
        s.listen(1)
        port = s.getsockname()[1]
    
    # Create security components
    from src.security.jwt_handler import JWTHandler
    from src.security.api_key_handler import APIKeyHandler
    from src.security.auth_manager import AuthManager
    from src.security.middleware import SecurityMiddleware
    
    jwt_handler = JWTHandler(secret_key="test-secret")
    api_key_handler = APIKeyHandler(storage_file="/tmp/test_keys.json")
    auth_manager = AuthManager(jwt_handler=jwt_handler, api_key_handler=api_key_handler)
    security_middleware = SecurityMiddleware(
        auth_manager=auth_manager,
        max_connections=10,
        requests_per_minute=60
    )
    
    # Create and start server
    server = WebSocketJsonRpcServer(
        host="127.0.0.1",
        port=port,
        websocket_path="/ws",
        max_connections=10
    )
    
    server.set_security_middleware(security_middleware)
    
    try:
        await server.start()
        
        # Connect to server
        uri = f"ws://127.0.0.1:{port}/ws"
        async with websockets.connect(uri) as websocket:
            # Test ping (should work without auth)
            await websocket.send(json.dumps({
                "jsonrpc": "2.0",
                "method": "ping",
                "id": 1
            }))
            
            response = await websocket.recv()
            result = json.loads(response)
            
            assert result["jsonrpc"] == "2.0"
            assert result["result"] == "pong"
            
            # Test protected method (should fail without auth)
            await websocket.send(json.dumps({
                "jsonrpc": "2.0",
                "method": "take_snapshot",
                "params": {"device": "/dev/video0"},
                "id": 2
            }))
            
            response = await websocket.recv()
            result = json.loads(response)
            
            assert result["jsonrpc"] == "2.0"
            assert "error" in result
            assert result["error"]["code"] == -32001  # Authentication error
            
    finally:
        await server.stop()
