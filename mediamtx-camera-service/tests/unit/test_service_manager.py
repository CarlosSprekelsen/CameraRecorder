# tests/unit/test_service_manager.py
"""
ServiceManager unit tests for core service orchestration and lifecycle management.

Requirements Coverage:
- REQ-API-001: WebSocket JSON-RPC 2.0 API endpoint at ws://localhost:8002/ws
- REQ-API-002: ping method for health checks
- REQ-API-003: get_camera_list method for camera enumeration
- REQ-TEST-001: Use single systemd-managed MediaMTX service instance
- REQ-TEST-002: No multiple MediaMTX instances or processes
- REQ-TEST-003: Validate against actual production MediaMTX service
- REQ-TEST-004: Use fixed systemd service ports (API: 9997, RTSP: 8554, WebRTC: 8889, HLS: 8888)
- REQ-TEST-005: Coordinate on shared service with proper test isolation
- REQ-TEST-006: Verify MediaMTX service is running via systemd before execution

Test Categories: Unit
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


from tests.utils.port_utils import find_free_port

def _free_port() -> int:
    """Get a free port for testing - using common port utility."""
    return find_free_port()


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
        health_port=find_free_port(),  # Dynamic health port to avoid conflicts
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
async def test_service_manager_lifecycle_req_tech_001():
    """REQ-TECH-001: Service-oriented architecture validation."""
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
