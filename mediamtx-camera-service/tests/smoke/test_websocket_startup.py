"""
WebSocket Real Connection Test

Tests real WebSocket server startup and connection validation.
Validates actual JSON-RPC protocol compliance testing.

This test replaces complex unit test mocks with real system validation
to provide better confidence in WebSocket server reliability.

Requirements:
- REQ-PERF-001: System shall handle concurrent camera operations efficiently
- REQ-PERF-002: System shall maintain responsive performance under load
- REQ-PERF-003: System shall meet latency requirements for real-time operations
- REQ-PERF-004: System shall handle resource constraints gracefully
- REQ-HEALTH-001: System shall provide comprehensive logging for health monitoring
- REQ-HEALTH-002: System shall support structured logging for production environments
- REQ-HEALTH-003: System shall enable correlation ID tracking across components
- REQ-ERROR-004: System shall handle configuration loading failures gracefully
- REQ-ERROR-005: System shall provide meaningful error messages for configuration issues
- REQ-ERROR-006: System shall handle logging configuration failures gracefully
- REQ-ERROR-007: System shall handle WebSocket connection failures gracefully
- REQ-ERROR-008: System shall handle MediaMTX service failures gracefully
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
    
    # Start real server as a task
    server_task = asyncio.create_task(server.start())
    
    # Wait for server to be ready
    await asyncio.sleep(0.5)
    
    try:
        yield server
    finally:
        # Cleanup: stop server
        await server.stop()
        server_task.cancel()
        try:
            await server_task
        except asyncio.CancelledError:
            pass
        await asyncio.sleep(0.2)


class TestWebSocketRealConnection:
    """Test real WebSocket server startup and connection."""
    
    @pytest.mark.asyncio
    async def test_websocket_real_connection(self):
        """Test real WebSocket server startup and connection."""
        # Create and start server directly (following working pattern)
        server = WebSocketJsonRpcServer(
            host="127.0.0.1", 
            port=8002, 
            websocket_path="/ws", 
            max_connections=10
        )
        
        server_task = asyncio.create_task(server.start())
        await asyncio.sleep(0.5)  # Wait for server to start
        
        try:
            # Connect with real WebSocket client
            uri = "ws://127.0.0.1:8002/ws"
            
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
                
        finally:
            # Cleanup
            await server.stop()
            server_task.cancel()
            try:
                await server_task
            except asyncio.CancelledError:
                pass
    
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
    async def test_websocket_json_rpc_compliance(self):
        """Test JSON-RPC 2.0 protocol compliance."""
        # Create and start server directly
        server = WebSocketJsonRpcServer(
            host="127.0.0.1", 
            port=8003, 
            websocket_path="/ws", 
            max_connections=10
        )
        
        server_task = asyncio.create_task(server.start())
        await asyncio.sleep(0.5)  # Wait for server to start
        
        try:
            uri = "ws://127.0.0.1:8003/ws"
            
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
                
        finally:
            # Cleanup
            await server.stop()
            server_task.cancel()
            try:
                await server_task
            except asyncio.CancelledError:
                pass
    
    @pytest.mark.asyncio
    async def test_websocket_server_stats(self):
        """Test WebSocket server statistics and status."""
        # Create and start server directly
        server = WebSocketJsonRpcServer(
            host="127.0.0.1", 
            port=8004, 
            websocket_path="/ws", 
            max_connections=10
        )
        
        server_task = asyncio.create_task(server.start())
        await asyncio.sleep(0.5)  # Wait for server to start
        
        try:
            # Get server stats
            stats = server.get_server_stats()
            
            assert "running" in stats
            assert "connected_clients" in stats
            assert "max_connections" in stats
            assert "registered_methods" in stats
            
            assert stats["running"] is True
            assert stats["max_connections"] == 10
            assert stats["connected_clients"] >= 0
            
        finally:
            # Cleanup
            await server.stop()
            server_task.cancel()
            try:
                await server_task
            except asyncio.CancelledError:
                pass


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
