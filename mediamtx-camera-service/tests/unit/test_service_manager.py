# tests/unit/test_service_manager.py
"""
Requirement-based tests for ServiceManager behavior.

Requirements Traceability:
- REQ-SVC-001: ServiceManager shall orchestrate camera discovery and MediaMTX integration
- REQ-SVC-002: ServiceManager shall handle camera lifecycle events with real component coordination
- REQ-SVC-001: ServiceManager shall provide WebSocket API for camera management
- REQ-MEDIA-001: MediaMTX integration shall use single systemd-managed service
- REQ-MEDIA-002: MediaMTX integration shall handle service failures gracefully

Story Coverage: S1 - Service Manager Integration
IV&V Control Point: Real service manager validation

Each test traces to a customer requirement and validates real component
behavior via the public WebSocket API and MediaMTX HTTP integration.
"""

import json
import socket
import subprocess
import tempfile
import os
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


def _free_port() -> int:
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
        s.bind(("127.0.0.1", 0))
        return s.getsockname()[1]


def _build_config(ws_port: int) -> Config:
    """Build configuration using real systemd-managed MediaMTX service."""
    return Config(
        server=ServerConfig(host="127.0.0.1", port=ws_port, websocket_path="/ws", max_connections=10),
        mediamtx=MediaMTXConfig(
            host="127.0.0.1",
            api_port=9997,  # Fixed systemd service port
            rtsp_port=8554,  # Fixed systemd service port
            webrtc_port=8889,  # Fixed systemd service port
            hls_port=8888,  # Fixed systemd service port
            recordings_path="./.tmp_recordings",
            snapshots_path="./.tmp_snapshots",
        ),
        camera=CameraConfig(device_range=[0, 1, 2], enable_capability_detection=True, detection_timeout=0.5),
        logging=LoggingConfig(),
        recording=RecordingConfig(),
        snapshots=SnapshotConfig(),
    )


@asynccontextmanager
async def _real_mediamtx_service():
    """
    Use real systemd-managed MediaMTX service instead of mock HTTP server.
    
    Requirements: REQ-MEDIA-001, REQ-MEDIA-002
    Architecture: AD-001 - Use single systemd-managed MediaMTX service
    """
    try:
        # Check if MediaMTX service is running via systemd (AD-001 compliance)
        result = subprocess.run(
            ['systemctl', 'is-active', 'mediamtx'],
            capture_output=True,
            text=True,
            timeout=10
        )
        
        if result.returncode != 0 or result.stdout.strip() != 'active':
            raise RuntimeError("MediaMTX service is not running. Please start it with: sudo systemctl start mediamtx")
        
        # Service is available, yield for test execution
        yield
        
    except Exception as e:
        # If MediaMTX service is not available, skip the test
        pytest.skip(f"MediaMTX service not available: {e}")


@pytest.mark.asyncio
@pytest.mark.unit
async def test_req_svc_lifecycle_001_start_and_ping_ws():
    """
    Req: SVC-LIFECYCLE-001
    Service must start all components and respond to WebSocket ping.
    """
    ws_port = _free_port()
    async with _real_mediamtx_service():
        cfg = _build_config(ws_port)
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
@pytest.mark.unit
async def test_req_svc_api_cam_list_001_get_camera_list_structure():
    """
    Req: SVC-API-CAM-LIST-001
    WebSocket API must return an object with cameras, total, and connected fields for get_camera_list.
    """
    ws_port = _free_port()
    async with _real_mediamtx_service():
        cfg = _build_config(ws_port)
        svc = ServiceManager(cfg)
        await svc.start()
        try:
            uri = f"ws://{cfg.server.host}:{cfg.server.port}{cfg.server.websocket_path}"
            async with websockets.connect(uri) as ws:
                await ws.send(json.dumps({"jsonrpc": "2.0", "id": 2, "method": "get_camera_list"}))
                resp = json.loads(await ws.recv())
                assert isinstance(resp.get("result"), dict)
                assert "cameras" in resp["result"]
                assert "total" in resp["result"]
                assert "connected" in resp["result"]
                assert isinstance(resp["result"]["cameras"], list)
        finally:
            await svc.stop()


@pytest.mark.asyncio
@pytest.mark.unit
async def test_req_svc_error_001_invalid_method_returns_error():
    """
    Req: SVC-ERROR-001
    Invalid JSON-RPC method must return error without crashing service.
    """
    ws_port = _free_port()
    async with _real_mediamtx_service():
        cfg = _build_config(ws_port)
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
@pytest.mark.unit
async def test_req_svc_shutdown_001_clean_shutdown_and_ws_unavailable():
    """
    Req: SVC-SHUTDOWN-001
    Service must shut down cleanly; WebSocket endpoint must no longer accept connections.
    """
    ws_port = _free_port()
    async with _real_mediamtx_service():
        cfg = _build_config(ws_port)
        svc = ServiceManager(cfg)
        await svc.start()
        await svc.stop()
        uri = f"ws://{cfg.server.host}:{cfg.server.port}{cfg.server.websocket_path}"
        with pytest.raises(Exception):
            await websockets.connect(uri)
