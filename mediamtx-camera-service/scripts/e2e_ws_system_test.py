import asyncio
import json
import socket
from contextlib import asynccontextmanager

from aiohttp import web
import websockets

from src.camera_service.config import ConfigManager
from src.camera_service.service_manager import ServiceManager


def get_free_port() -> int:
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
        s.bind(("127.0.0.1", 0))
        return s.getsockname()[1]


@asynccontextmanager
async def start_fake_mediamtx(host: str, port: int):
    async def health(_request: web.Request):
        return web.json_response({"serverVersion": "test", "serverUptime": 1})

    app = web.Application()
    app.router.add_get("/v3/health", health)

    runner = web.AppRunner(app)
    await runner.setup()
    site = web.TCPSite(runner, host, port)
    await site.start()
    try:
        yield
    finally:
        await runner.cleanup()


async def main() -> int:
    # Prepare config
    cfg = ConfigManager().load_config()
    cfg.server.host = "127.0.0.1"
    cfg.server.port = get_free_port()
    # Point MediaMTX controller to local fake endpoint for health
    cfg.mediamtx.host = "127.0.0.1"
    cfg.mediamtx.api_port = get_free_port()

    async with start_fake_mediamtx(cfg.mediamtx.host, cfg.mediamtx.api_port):
        svc = ServiceManager(cfg)
        await svc.start()
        try:
            uri = f"ws://{cfg.server.host}:{cfg.server.port}{cfg.server.websocket_path}"
            async with websockets.connect(uri) as ws:
                # Ping
                await ws.send(json.dumps({"jsonrpc": "2.0", "id": 1, "method": "ping", "params": {}}))
                ping_resp = await ws.recv()
                print("PING:", ping_resp)

                # Camera list
                await ws.send(json.dumps({"jsonrpc": "2.0", "id": 2, "method": "get_camera_list"}))
                cams_resp = await ws.recv()
                print("CAM_LIST:", cams_resp)

                # Basic validation: responses are JSON-RPC
                pr = json.loads(ping_resp)
                cr = json.loads(cams_resp)
                assert pr.get("result") in ("pong", {"status": "ok"}) or "result" in pr
                assert "result" in cr and isinstance(cr["result"], list)
                print("STATUS: OK")
                return 0
        finally:
            await svc.stop()


if __name__ == "__main__":
    raise SystemExit(asyncio.run(main()))


