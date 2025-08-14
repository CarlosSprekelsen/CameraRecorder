"""
WebSocket Real Connection Test

Tests real WebSocket server startup and connection validation.
Validates actual JSON-RPC protocol compliance testing.

This test replaces complex unit test mocks with real system validation
to provide better confidence in WebSocket server reliability.
"""

import asyncio
import json
import logging
import pytest
import websockets
from typing import Dict, Any

# Import the actual WebSocket server implementation
import sys
import os
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', '..', 'src'))

from websocket_server.server import WebSocketJsonRpcServer


@pytest.fixture(scope="function")
async def websocket_server():
    """Create and manage WebSocket server for testing."""
    server = WebSocketJsonRpcServer(
        host="127.0.0.1", 
        port=8002, 
        websocket_path="/ws", 
        max_connections=10
    )
    
    # Start real server
    await server.start()
    
    # Wait for server to be ready
    await asyncio.sleep(0.2)
    
    yield server
    
    # Cleanup: stop server
    await server.stop()
    await asyncio.sleep(0.2)


class TestWebSocketRealConnection:
    """Test real WebSocket server startup and connection."""
    
    @pytest.mark.asyncio
    async def test_websocket_real_connection(self, websocket_server):
        """Test real WebSocket server startup and connection."""
        # Connect with real WebSocket client
        uri = "ws://127.0.0.1:8002/ws"
        
        try:
            async with websockets.connect(uri) as ws:
                # Test real JSON-RPC ping
                await ws.send(json.dumps({
                    "jsonrpc": "2.0", 
                    "id": 1, 
                    "method": "ping"
                }))
                
                response = json.loads(await ws.recv())
                assert response["result"] == "pong"
                assert response["jsonrpc"] == "2.0"
                assert "id" in response
                
        except Exception as e:
            pytest.fail(f"WebSocket connection test failed: {e}")
    
    @pytest.mark.asyncio
    async def test_websocket_server_lifecycle(self):
        """Test WebSocket server startup and shutdown lifecycle."""
        server = WebSocketJsonRpcServer(
            host="127.0.0.1", 
            port=8003, 
            websocket_path="/ws", 
            max_connections=5
        )
        
        # Test server startup
        await server.start()
        await asyncio.sleep(0.2)  # Wait for server to be ready
        assert server.is_running is True
        
        # Test server shutdown
        await server.stop()
        await asyncio.sleep(0.2)  # Wait for server to stop
        assert server.is_running is False
    
    @pytest.mark.asyncio
    async def test_websocket_json_rpc_compliance(self, websocket_server):
        """Test JSON-RPC 2.0 protocol compliance."""
        uri = "ws://127.0.0.1:8002/ws"
        
        try:
            async with websockets.connect(uri) as ws:
                # Test method not found
                await ws.send(json.dumps({
                    "jsonrpc": "2.0",
                    "id": 2,
                    "method": "nonexistent_method"
                }))
                
                response = json.loads(await ws.recv())
                assert response["jsonrpc"] == "2.0"
                assert "error" in response
                assert response["error"]["code"] == -32601  # Method not found
                assert response["id"] == 2
                
        except Exception as e:
            pytest.fail(f"JSON-RPC compliance test failed: {e}")
    
    @pytest.mark.asyncio
    async def test_websocket_server_stats(self, websocket_server):
        """Test WebSocket server statistics and status."""
        # Get server stats - websocket_server is the actual server instance
        stats = websocket_server.get_server_stats()
        
        assert "running" in stats
        assert "connected_clients" in stats
        assert "max_connections" in stats
        assert "registered_methods" in stats
        
        assert stats["running"] is True
        assert stats["max_connections"] == 10
        assert stats["connected_clients"] >= 0


if __name__ == "__main__":
    # Allow running as standalone script for manual testing
    async def run_tests():
        """Run smoke tests manually."""
        test_instance = TestWebSocketRealConnection()
        
        # Test server lifecycle
        await test_instance.test_websocket_server_lifecycle()
        print("✓ WebSocket server lifecycle test passed")
        
        # Test real connection with manual server management
        server = WebSocketJsonRpcServer(
            host="127.0.0.1", 
            port=8002, 
            websocket_path="/ws", 
            max_connections=10
        )
        
        try:
            await server.start()
            await asyncio.sleep(0.2)  # Wait for server to be ready
            
            await test_instance.test_websocket_real_connection(server)
            print("✓ WebSocket real connection test passed")
            
            await test_instance.test_websocket_json_rpc_compliance(server)
            print("✓ WebSocket JSON-RPC compliance test passed")
            
            await test_instance.test_websocket_server_stats(server)
            print("✓ WebSocket server stats test passed")
            
        finally:
            await server.stop()
            await asyncio.sleep(0.2)
        
        print("All WebSocket smoke tests passed!")
    
    asyncio.run(run_tests())
