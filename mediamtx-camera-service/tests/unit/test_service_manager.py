# tests/unit/test_service_manager.py
"""
Requirement-based tests for ServiceManager behavior.

Each test traces to a customer requirement and validates real component
behavior via the public WebSocket API and MediaMTX HTTP integration.
"""

import asyncio
import json
import socket
from contextlib import asynccontextmanager

import pytest
import websockets

from src.camera_service.config import (
    Config,
    ServerConfig,
    MediaMTXConfig,
    CameraConfig,
    LoggingConfig,
    RecordingConfig,
    SnapshotConfig,
)
from src.camera_service.service_manager import ServiceManager
from aiohttp import web


def _free_port() -> int:
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
        s.bind(("127.0.0.1", 0))
        return s.getsockname()[1]


def _build_config(api_port: int, ws_port: int) -> Config:
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


@asynccontextmanager
async def _mediamtx_health_ok(host: str, port: int):
    async def health(_req: web.Request):
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


@pytest.mark.asyncio
async def test_req_svc_lifecycle_001_start_and_ping_ws():
    """
    Req: SVC-LIFECYCLE-001
    Service must start all components and respond to WebSocket ping.
    """
    api_port = _free_port()
    ws_port = _free_port()
    async with _mediamtx_health_ok("127.0.0.1", api_port):
        cfg = _build_config(api_port, ws_port)
        svc = ServiceManager(cfg)
        await svc.start()
        try:
            uri = f"ws://{cfg.server.host}:{cfg.server.port}{cfg.server.websocket_path}"
            async with websockets.connect(uri) as ws:
                await ws.send(json.dumps({"jsonrpc": "2.0", "id": 1, "method": "ping", "params": {}}))
                resp = json.loads(await ws.recv())
                assert "result" in resp
        finally:
            await svc.stop()


@pytest.mark.asyncio
async def test_req_svc_api_cam_list_001_get_camera_list_structure():
    """
    Req: SVC-API-CAM-LIST-001
    WebSocket API must return a list for get_camera_list (may be empty).
    """
    api_port = _free_port()
    ws_port = _free_port()
    async with _mediamtx_health_ok("127.0.0.1", api_port):
        cfg = _build_config(api_port, ws_port)
        svc = ServiceManager(cfg)
        await svc.start()
        try:
            uri = f"ws://{cfg.server.host}:{cfg.server.port}{cfg.server.websocket_path}"
            async with websockets.connect(uri) as ws:
                await ws.send(json.dumps({"jsonrpc": "2.0", "id": 2, "method": "get_camera_list"}))
                resp = json.loads(await ws.recv())
                assert isinstance(resp.get("result"), list)
        finally:
            await svc.stop()


@pytest.mark.asyncio
async def test_req_svc_error_001_invalid_method_returns_error():
    """
    Req: SVC-ERROR-001
    Invalid JSON-RPC method must return error without crashing service.
    """
    api_port = _free_port()
    ws_port = _free_port()
    async with _mediamtx_health_ok("127.0.0.1", api_port):
        cfg = _build_config(api_port, ws_port)
        svc = ServiceManager(cfg)
        await svc.start()
        try:
            uri = f"ws://{cfg.server.host}:{cfg.server.port}{cfg.server.websocket_path}"
            async with websockets.connect(uri) as ws:
                await ws.send(json.dumps({"jsonrpc": "2.0", "id": 3, "method": "non_existent"}))
                resp = json.loads(await ws.recv())
                assert "error" in resp and resp["error"]["code"] < 0
        finally:
            await svc.stop()


@pytest.mark.asyncio
async def test_req_svc_shutdown_001_clean_shutdown_and_ws_unavailable():
    """
    Req: SVC-SHUTDOWN-001
    Service must shut down cleanly; WebSocket endpoint must no longer accept connections.
    """
    api_port = _free_port()
    ws_port = _free_port()
    async with _mediamtx_health_ok("127.0.0.1", api_port):
        cfg = _build_config(api_port, ws_port)
        svc = ServiceManager(cfg)
        await svc.start()
        await svc.stop()
        uri = f"ws://{cfg.server.host}:{cfg.server.port}{cfg.server.websocket_path}"
        with pytest.raises(Exception):
            await websockets.connect(uri)
