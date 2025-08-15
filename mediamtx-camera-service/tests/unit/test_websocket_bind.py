"""
WebSocket server binding and connectivity tests.

Requirements Traceability:
- REQ-WS-006: WebSocket server shall bind to specified host and port
- REQ-WS-006: WebSocket server shall accept connections and respond to ping requests
- REQ-WS-006: WebSocket server shall provide JSON-RPC 2.0 protocol support

Story Coverage: S3 - WebSocket API Integration
IV&V Control Point: Real WebSocket binding validation
"""

import json
import pytest

from src.websocket_server.server import WebSocketJsonRpcServer


@pytest.mark.asyncio
async def test_websocket_server_binds_and_ping(monkeypatch):
    server = WebSocketJsonRpcServer(
        host="127.0.0.1", port=8022, websocket_path="/ws", max_connections=5
    )
    await server.start()
    try:
        import websockets

        uri = "ws://127.0.0.1:8022/ws"
        async with websockets.connect(uri) as ws:
            await ws.send(json.dumps({"jsonrpc": "2.0", "id": 1, "method": "ping"}))
            resp = json.loads(await ws.recv())
            assert resp["result"] == "pong"
    finally:
        await server.stop()

