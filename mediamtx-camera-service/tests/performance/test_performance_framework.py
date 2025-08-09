import asyncio
import json
import time
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


