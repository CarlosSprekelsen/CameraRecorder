# tests/unit/test_camera_service/test_service_manager_lifecycle.py
"""
Real component lifecycle tests for ServiceManager.

Requirements Traceability:
- REQ-SVC-001: ServiceManager shall orchestrate camera discovery and MediaMTX integration
- REQ-SVC-002: ServiceManager shall handle camera lifecycle events with real component coordination
- REQ-ERROR-003: ServiceManager shall maintain operation during MediaMTX failures

Story Coverage: S1 - Service Manager Integration
IV&V Control Point: Real component orchestration validation

Replaces over-mocked tests by orchestrating actual internal components and mocking
only external boundaries (HTTP to MediaMTX via aiohttp and filesystem isolation).
"""

import asyncio
import tempfile
import shutil
import aiohttp
import aiohttp.test_utils
from pathlib import Path
from typing import Dict

import pytest
# Real HTTP integration - no mocks needed

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


@pytest.fixture
def temp_dirs() -> Dict[str, str]:
    base = tempfile.mkdtemp()
    rec = Path(base) / "recordings"
    snap = Path(base) / "snapshots"
    rec.mkdir()
    snap.mkdir()
    try:
        yield {"base": base, "recordings": str(rec), "snapshots": str(snap)}
    finally:
        shutil.rmtree(base, ignore_errors=True)


@pytest.fixture
def real_config(temp_dirs: Dict[str, str]) -> Config:
    return Config(
        server=ServerConfig(host="localhost", port=8002, websocket_path="/ws", max_connections=10),
        mediamtx=MediaMTXConfig(
            host="localhost",
            api_port=9997,
            rtsp_port=8554,
            webrtc_port=8889,
            hls_port=8888,
            recordings_path=temp_dirs["recordings"],
            snapshots_path=temp_dirs["snapshots"],
        ),
        camera=CameraConfig(device_range=[0, 1, 2], enable_capability_detection=True, detection_timeout=1.0),
        logging=LoggingConfig(),
        recording=RecordingConfig(),
        snapshots=SnapshotConfig(),
    )


@pytest.fixture
def service_manager(real_config: Config) -> ServiceManager:
    return ServiceManager(real_config)


@pytest.fixture
async def mock_mediamtx_server():
    """Create a real HTTP test server that simulates MediaMTX API responses."""
    
    async def handle_health_check(request):
        """Handle MediaMTX health check endpoint."""
        return aiohttp.web.json_response({
            "serverVersion": "v1.0.0",
            "serverUptime": 3600,
            "apiVersion": "v3"
        })
    
    async def handle_path_config(request):
        """Handle MediaMTX path configuration endpoint."""
        return aiohttp.web.json_response({"status": "ok"})
    
    async def handle_stream_status(request):
        """Handle MediaMTX stream status endpoint."""
        return aiohttp.web.json_response({
            "items": [
                {
                    "name": "camera0",
                    "status": "active",
                    "source": "rtsp://localhost:8554/camera0"
                }
            ]
        })
    
    app = aiohttp.web.Application()
    app.router.add_get('/v3/config/global/get', handle_health_check)
    app.router.add_post('/v3/config/paths/edit/{path_name}', handle_path_config)
    app.router.add_get('/v3/paths/list', handle_stream_status)
    
    runner = aiohttp.test_utils.TestServer(app, port=9997)
    await runner.start_server()
    
    try:
        yield runner
    finally:
        await runner.close()


@pytest.fixture
async def mock_mediamtx_server_failure():
    """Create a real HTTP test server that simulates MediaMTX failures."""
    
    async def handle_health_check(request):
        """Handle MediaMTX health check endpoint."""
        return aiohttp.web.json_response({
            "serverVersion": "v1.0.0",
            "serverUptime": 3600,
            "apiVersion": "v3"
        })
    
    async def handle_path_config_failure(request):
        """Handle MediaMTX path configuration endpoint with failure."""
        return aiohttp.web.json_response(
            {"error": "Internal server error"}, 
            status=500
        )
    
    async def handle_stream_status(request):
        """Handle MediaMTX stream status endpoint."""
        return aiohttp.web.json_response({
            "items": []
        })
    
    app = aiohttp.web.Application()
    app.router.add_get('/v3/config/global/get', handle_health_check)
    app.router.add_post('/v3/config/paths/edit/{path_name}', handle_path_config_failure)
    app.router.add_get('/v3/paths/list', handle_stream_status)
    
    runner = aiohttp.test_utils.TestServer(app, port=9998)
    await runner.start_server()
    
    try:
        yield runner
    finally:
        await runner.close()


def _connected_event(device_path: str = "/dev/video0") -> CameraEventData:
    dev = CameraDevice(device=device_path, name=f"Camera {device_path}", status="CONNECTED")
    return CameraEventData(device_path=device_path, event_type=CameraEvent.CONNECTED, device_info=dev, timestamp=1234567890.0)


def _disconnected_event(device_path: str = "/dev/video0") -> CameraEventData:
    dev = CameraDevice(device=device_path, name=f"Camera {device_path}", status="DISCONNECTED")
    return CameraEventData(device_path=device_path, event_type=CameraEvent.DISCONNECTED, device_info=dev, timestamp=1234567891.0)


@pytest.mark.asyncio
async def test_real_connect_flow(service_manager: ServiceManager, mock_mediamtx_server):
    """Test real camera connect flow with actual HTTP integration."""
    # Use real HTTP server instead of mocked session
    await asyncio.wait_for(service_manager.start(), timeout=10.0)
    await service_manager.handle_camera_event(_connected_event("/dev/video0"))

    # Ensure service remained running
    assert service_manager.is_running is True

    await service_manager.stop()


@pytest.mark.asyncio
async def test_real_disconnect_flow(service_manager: ServiceManager, mock_mediamtx_server):
    """Test real camera disconnect flow with actual HTTP integration."""
    # Use real HTTP server instead of mocked session
    await asyncio.wait_for(service_manager.start(), timeout=10.0)
    await service_manager.handle_camera_event(_connected_event("/dev/video0"))

    await service_manager.handle_camera_event(_disconnected_event("/dev/video0"))
    assert service_manager.is_running is True

    await service_manager.stop()


@pytest.mark.asyncio
async def test_real_mediamtx_failure_keeps_service_running(service_manager: ServiceManager, mock_mediamtx_server_failure):
    """Test that service remains running when MediaMTX fails with real HTTP integration."""
    # Use real HTTP server that returns 500 errors instead of mocked session
    await asyncio.wait_for(service_manager.start(), timeout=10.0)
    await service_manager.handle_camera_event(_connected_event("/dev/video0"))
    assert service_manager.is_running is True
    await service_manager.stop()


@pytest.mark.asyncio
async def test_real_capability_metadata(service_manager: ServiceManager, mock_mediamtx_server):
    """Test real camera capability metadata with actual HTTP integration."""
    # Use real HTTP server instead of mocked session
    await asyncio.wait_for(service_manager.start(), timeout=10.0)
    try:
        md = await service_manager._get_enhanced_camera_metadata(_connected_event("/dev/video0"))
        assert isinstance(md, dict)
        assert "resolution" in md and "fps" in md and "validation_status" in md
    finally:
        await service_manager.stop()


# TODO: HIGH: Add integration tests with real camera monitor instance [Story:E1/S5]
# TODO: MEDIUM: Add performance tests for camera event processing latency [Story:E1/S5]
# TODO: MEDIUM: Add stress tests for rapid connect/disconnect sequences [Story:E1/S5]
# TODO: LOW: Add tests for unknown camera event types [Story:E1/S5]
