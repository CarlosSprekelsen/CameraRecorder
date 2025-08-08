# tests/unit/test_camera_service/test_service_manager_lifecycle.py
"""
Real component lifecycle tests for ServiceManager.

Replaces over-mocked tests by orchestrating actual internal components and mocking
only external boundaries (HTTP to MediaMTX via aiohttp and filesystem isolation).
"""

import asyncio
import tempfile
import shutil
from pathlib import Path
from typing import Dict

import pytest
from unittest.mock import AsyncMock, patch

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


def _connected_event(device_path: str = "/dev/video0") -> CameraEventData:
    dev = CameraDevice(device=device_path, name=f"Camera {device_path}", status="CONNECTED")
    return CameraEventData(device_path=device_path, event_type=CameraEvent.CONNECTED, device_info=dev, timestamp=1234567890.0)


def _disconnected_event(device_path: str = "/dev/video0") -> CameraEventData:
    dev = CameraDevice(device=device_path, name=f"Camera {device_path}", status="DISCONNECTED")
    return CameraEventData(device_path=device_path, event_type=CameraEvent.DISCONNECTED, device_info=dev, timestamp=1234567891.0)


@pytest.mark.asyncio
async def test_real_connect_flow(service_manager: ServiceManager):
    with patch("src.mediamtx_wrapper.controller.aiohttp.ClientSession") as mock_session:
        from unittest.mock import MagicMock
        sess = MagicMock()
        get_ctx = MagicMock()
        get_enter = MagicMock()
        get_enter.status = 200
        get_enter.json = AsyncMock(return_value={"serverVersion": "x", "serverUptime": 1})
        get_ctx.__aenter__ = AsyncMock(return_value=get_enter)
        post_ctx = MagicMock()
        post_enter = MagicMock()
        post_enter.status = 200
        post_ctx.__aenter__ = AsyncMock(return_value=post_enter)
        sess.get.return_value = get_ctx
        sess.post.return_value = post_ctx
        mock_session.return_value = sess

        await asyncio.wait_for(service_manager.start(), timeout=10.0)
        await service_manager.handle_camera_event(_connected_event("/dev/video0"))

        # Ensure service remained running
        assert service_manager.is_running is True

        await service_manager.stop()


@pytest.mark.asyncio
async def test_real_disconnect_flow(service_manager: ServiceManager):
    with patch("src.mediamtx_wrapper.controller.aiohttp.ClientSession") as mock_session:
        from unittest.mock import MagicMock
        sess = MagicMock()
        get_ctx = MagicMock()
        get_enter = MagicMock()
        get_enter.status = 200
        get_enter.json = AsyncMock(return_value={"serverVersion": "x", "serverUptime": 1})
        get_ctx.__aenter__ = AsyncMock(return_value=get_enter)
        post_ctx = MagicMock()
        post_enter = MagicMock()
        post_enter.status = 200
        post_ctx.__aenter__ = AsyncMock(return_value=post_enter)
        sess.get.return_value = get_ctx
        sess.post.return_value = post_ctx
        mock_session.return_value = sess

        await asyncio.wait_for(service_manager.start(), timeout=10.0)
        await service_manager.handle_camera_event(_connected_event("/dev/video0"))

        await service_manager.handle_camera_event(_disconnected_event("/dev/video0"))
        assert service_manager.is_running is True

        await service_manager.stop()


@pytest.mark.asyncio
async def test_real_mediamtx_failure_keeps_service_running(service_manager: ServiceManager):
    with patch("src.mediamtx_wrapper.controller.aiohttp.ClientSession") as mock_session:
        from unittest.mock import MagicMock
        sess = MagicMock()
        get_ctx = MagicMock()
        get_enter = MagicMock()
        get_enter.status = 200
        get_enter.json = AsyncMock(return_value={"serverVersion": "x", "serverUptime": 1})
        get_ctx.__aenter__ = AsyncMock(return_value=get_enter)
        post_ctx = MagicMock()
        post_enter = MagicMock()
        post_enter.status = 500
        post_enter.text = AsyncMock(return_value="error")
        post_ctx.__aenter__ = AsyncMock(return_value=post_enter)
        sess.get.return_value = get_ctx
        sess.post.return_value = post_ctx
        mock_session.return_value = sess

        await asyncio.wait_for(service_manager.start(), timeout=10.0)
        await service_manager.handle_camera_event(_connected_event("/dev/video0"))
        assert service_manager.is_running is True
        await service_manager.stop()


@pytest.mark.asyncio
async def test_real_capability_metadata(service_manager: ServiceManager):
    with patch("src.mediamtx_wrapper.controller.aiohttp.ClientSession") as mock_session:
        from unittest.mock import MagicMock
        sess = MagicMock()
        get_ctx = MagicMock()
        get_enter = MagicMock()
        get_enter.status = 200
        get_enter.json = AsyncMock(return_value={"serverVersion": "x", "serverUptime": 1})
        get_ctx.__aenter__ = AsyncMock(return_value=get_enter)
        post_ctx = MagicMock()
        post_enter = MagicMock()
        post_enter.status = 200
        post_ctx.__aenter__ = AsyncMock(return_value=post_enter)
        sess.get.return_value = get_ctx
        sess.post.return_value = post_ctx
        mock_session.return_value = sess

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
