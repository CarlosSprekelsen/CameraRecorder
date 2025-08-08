import asyncio
import socket
from contextlib import asynccontextmanager
from dataclasses import replace

import pytest
from aiohttp import web

from src.camera_service.service_manager import ServiceManager
from src.camera_service.config import (
    Config,
    ServerConfig,
    MediaMTXConfig,
    CameraConfig,
    LoggingConfig,
    RecordingConfig,
    SnapshotConfig,
)
from camera_discovery.hybrid_monitor import CameraEventData, CameraEvent
from src.common.types import CameraDevice
import websockets
import json


def get_free_port() -> int:
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
        s.bind(("127.0.0.1", 0))
        return s.getsockname()[1]


@asynccontextmanager
async def start_fake_mediamtx_server(host: str, port: int):
    calls = {"added": [], "deleted": []}

    async def health(request: web.Request):
        return web.json_response({"serverVersion": "test", "serverUptime": 1})

    async def add_path(request: web.Request):
        name = request.match_info["name"]
        calls["added"].append(name)
        return web.json_response({"status": "ok"})

    async def delete_path(request: web.Request):
        name = request.match_info["name"]
        calls["deleted"].append(name)
        return web.json_response({"status": "ok"})

    app = web.Application()
    app.router.add_get("/v3/health", health)
    app.router.add_post("/v3/config/paths/add/{name}", add_path)
    app.router.add_post("/v3/config/paths/delete/{name}", delete_path)

    runner = web.AppRunner(app)
    await runner.setup()
    site = web.TCPSite(runner, host, port)
    await site.start()
    try:
        yield calls
    finally:
        await runner.cleanup()


def build_config(api_port: int, ws_port: int) -> Config:
    return Config(
        server=ServerConfig(host="127.0.0.1", port=ws_port, websocket_path="/ws", max_connections=10),
        mediamtx=MediaMTXConfig(
            host="127.0.0.1",
            api_port=api_port,
            rtsp_port=8554,
            webrtc_port=8889,
            hls_port=8888,
            recordings_path="./.tmp_recordings",
            snapshots_path="./.tmp_snapshots",
        ),
        camera=CameraConfig(device_range=[0, 1, 2], enable_capability_detection=True, detection_timeout=0.5),
        logging=LoggingConfig(),
        recording=RecordingConfig(),
        snapshots=SnapshotConfig(),
    )


@pytest.mark.asyncio
async def test_e2e_connect_disconnect_creates_and_deletes_paths():
    api_port = get_free_port()
    ws_port = get_free_port()
    async with start_fake_mediamtx_server("127.0.0.1", api_port) as calls:
        cfg = build_config(api_port, ws_port)
        svc = ServiceManager(cfg)
        await svc.start()
        try:
            # Simulate camera connect
            event_conn = CameraEventData(
                device_path="/dev/video0",
                event_type=CameraEvent.CONNECTED,
                device_info=CameraDevice(device="/dev/video0", name="Camera 0", status="CONNECTED"),
                timestamp=0.0,
            )
            await svc.handle_camera_event(event_conn)

            # Simulate camera disconnect
            event_disc = CameraEventData(
                device_path="/dev/video0",
                event_type=CameraEvent.DISCONNECTED,
                device_info=CameraDevice(device="/dev/video0", name="Camera 0", status="DISCONNECTED"),
                timestamp=1.0,
            )
            await svc.handle_camera_event(event_disc)

            # Validate MediaMTX interactions
            assert "camera0" in calls["added"]
            assert "camera0" in calls["deleted"]
            assert svc.is_running is True

            # WebSocket API: validate camera list availability (F3.1.1)
            uri = f"ws://{cfg.server.host}:{cfg.server.port}{cfg.server.websocket_path}"
            async with websockets.connect(uri) as ws:
                await ws.send(json.dumps({"jsonrpc":"2.0","id":1,"method":"get_camera_list"}))
                resp = json.loads(await ws.recv())
                assert "result" in resp and isinstance(resp["result"], list)
        finally:
            await svc.stop()


@pytest.mark.asyncio
async def test_e2e_resilience_on_mediamtx_failure():
    api_port = get_free_port()
    ws_port = get_free_port()

    # Start server that fails on add but succeeds on health
    calls = {"added": [], "deleted": []}

    async def health(request: web.Request):
        return web.json_response({"serverVersion": "test", "serverUptime": 1})

    async def add_path(request: web.Request):
        calls["added"].append(request.match_info["name"])
        return web.Response(status=500, text="error")

    app = web.Application()
    app.router.add_get("/v3/health", health)
    app.router.add_post("/v3/config/paths/add/{name}", add_path)
    runner = web.AppRunner(app)
    await runner.setup()
    site = web.TCPSite(runner, "127.0.0.1", api_port)
    await site.start()

    cfg = build_config(api_port, ws_port)
    svc = ServiceManager(cfg)
    await svc.start()
    try:
        event_conn = CameraEventData(
            device_path="/dev/video1",
            event_type=CameraEvent.CONNECTED,
            device_info=CameraDevice(device="/dev/video1", name="Camera 1", status="CONNECTED"),
            timestamp=0.0,
        )
        await svc.handle_camera_event(event_conn)
        # Service should remain running even if add fails
        assert svc.is_running is True
        assert "camera1" in calls["added"]
    finally:
        await svc.stop()
        await runner.cleanup()


