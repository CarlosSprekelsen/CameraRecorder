"""
Performance framework tests.

Tests performance monitoring and metrics collection framework
for real-time performance analysis and optimization.

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
import pytest

from src.websocket_server.server import WebSocketJsonRpcServer


@pytest.mark.asyncio
async def test_get_metrics_and_method_timings():
    host = "127.0.0.1"
    port = 8770
    server = WebSocketJsonRpcServer(host=host, port=port, websocket_path="/ws", max_connections=10)

    # Register a method with artificial delay to produce measurable timing
    async def delayed_method(params=None):
        await asyncio.sleep(0.05)
        return {"ok": True}

    server.register_method("delayed", delayed_method)
    await server.start()

    import websockets
    uri = f"ws://{host}:{port}/ws"
    async with websockets.connect(uri) as ws:
        # Trigger method multiple times
        for i in range(5):
            await ws.send(json.dumps({"jsonrpc": "2.0", "id": i+1, "method": "delayed"}))
            _ = await ws.recv()

        # Fetch metrics
        await ws.send(json.dumps({"jsonrpc": "2.0", "id": 99, "method": "get_metrics"}))
        metrics_resp = json.loads(await ws.recv())
        assert "methods" in metrics_resp["result"]
        m = metrics_resp["result"]["methods"]["delayed"]
        assert m["count"] >= 5
        assert m["avg_ms"] >= 50.0
        assert m["max_ms"] >= m["avg_ms"]

    await server.stop()


