"""
Integration tests focused on requirement-based behavior validation.

Requirements Traceability:
- REQ-INT-001: Integration system shall provide requirement-based behavior validation
- REQ-INT-004: Integration system shall validate real component orchestration
- REQ-INT-001: Integration system shall test error/edge-case behavior and business logic

Story Coverage: S4 - System Integration
IV&V Control Point: Real component integration validation

Replaces smoke/instantiation-only checks with tests that:
- Trace to requirements
- Exercise real component orchestration
- Validate error/edge-case behavior and business logic
"""

import pytest

from src.camera_service.config import ConfigManager
from src.camera_service.service_manager import ServiceManager


class TestConfigurationComponentIntegration:
    """Requirement-driven integration validations."""

    @pytest.mark.asyncio
    @pytest.mark.integration
    async def test_stream_creation_uses_configured_endpoints_on_connect(self, temp_test_dir):
        """
        Req: S5-STREAM-ADD-001
        On camera CONNECTED, service must create MediaMTX path using configured host/ports.
        Verifies real orchestration with external HTTP boundary patched.
        """
        import tempfile
        import os
        
        # Create temporary directories for recordings and snapshots
        recordings_dir = os.path.join(temp_test_dir, "recordings")
        snapshots_dir = os.path.join(temp_test_dir, "snapshots")
        os.makedirs(recordings_dir, exist_ok=True)
        os.makedirs(snapshots_dir, exist_ok=True)
        
        # Create configuration with temporary directories
        cfg = ConfigManager().load_config()
        cfg.mediamtx.recordings_path = recordings_dir
        cfg.mediamtx.snapshots_path = snapshots_dir
        svc = ServiceManager(cfg)

        # Patch only external HTTP client
        with pytest.MonkeyPatch.context() as mp:
            async def _fake_ctx_enter_ok():
                class Resp:
                    status = 200

                    async def json(self):
                        return {"serverVersion": "x", "serverUptime": 1}

                return Resp()

            class FakeCtx:
                async def __aenter__(self):
                    return await _fake_ctx_enter_ok()

                async def __aexit__(self, exc_type, exc, tb):
                    return False

            class FakeSession:
                def get(self, *_, **__):
                    return FakeCtx()

                def post(self, url, *_, **__):
                    FakeSession.post_urls.append(url)
                    return FakeCtx()

            FakeSession.post_urls = []

            from src.mediamtx_wrapper import controller as ctrl

            def _fake_client_session(*_args, **_kwargs):
                return FakeSession()

            mp.setattr(ctrl.aiohttp, "ClientSession", _fake_client_session)

            await svc.start()
            from camera_discovery.hybrid_monitor import CameraEventData, CameraEvent
            from src.common.types import CameraDevice

            event = CameraEventData(
                device_path="/dev/video0",
                event_type=CameraEvent.CONNECTED,
                device_info=CameraDevice(device="/dev/video0", name="Camera 0", status="CONNECTED"),
                timestamp=0.0,
            )

            await svc.handle_camera_event(event)
            assert svc.is_running is True

            # Business outcome: path add attempted for camera0
            assert any("/v3/config/paths/add/camera0" in u for u in FakeSession.post_urls)

            await svc.stop()

    @pytest.mark.asyncio
    async def test_resilience_on_stream_creation_failure(self, temp_test_dir):
        """
        Req: S5-RES-002
        If MediaMTX path creation fails, service remains operational and does not crash.
        """
        import tempfile
        import os
        
        # Create temporary directories for recordings and snapshots
        recordings_dir = os.path.join(temp_test_dir, "recordings")
        snapshots_dir = os.path.join(temp_test_dir, "snapshots")
        os.makedirs(recordings_dir, exist_ok=True)
        os.makedirs(snapshots_dir, exist_ok=True)
        
        # Create configuration with temporary directories
        cfg = ConfigManager().load_config()
        cfg.mediamtx.recordings_path = recordings_dir
        cfg.mediamtx.snapshots_path = snapshots_dir
        svc = ServiceManager(cfg)

        with pytest.MonkeyPatch.context() as mp:
            async def _fake_ctx_enter_fail():
                class Resp:
                    status = 500

                    async def text(self):
                        return "error"

                return Resp()

            async def _fake_ctx_enter_ok():
                class Resp:
                    status = 200

                    async def json(self):
                        return {"serverVersion": "x", "serverUptime": 1}

                return Resp()

            class FakeOkCtx:
                async def __aenter__(self):
                    return await _fake_ctx_enter_ok()

                async def __aexit__(self, exc_type, exc, tb):
                    return False

            class FakeFailCtx:
                async def __aenter__(self):
                    return await _fake_ctx_enter_fail()

                async def __aexit__(self, exc_type, exc, tb):
                    return False

            class FakeSession:
                def get(self, *_, **__):
                    return FakeOkCtx()

                def post(self, *_, **__):
                    return FakeFailCtx()

            from src.mediamtx_wrapper import controller as ctrl

            def _fake_client_session(*_args, **_kwargs):
                return FakeSession()

            mp.setattr(ctrl.aiohttp, "ClientSession", _fake_client_session)

            await svc.start()
            from camera_discovery.hybrid_monitor import CameraEventData, CameraEvent
            from src.common.types import CameraDevice

            event = CameraEventData(
                device_path="/dev/video1",
                event_type=CameraEvent.CONNECTED,
                device_info=CameraDevice(device="/dev/video1", name="Camera 1", status="CONNECTED"),
                timestamp=0.0,
            )
            await svc.handle_camera_event(event)

            # Requirement: service remains running despite MediaMTX error
            assert svc.is_running is True
            await svc.stop()


class TestConfigurationValidation:
    """Requirement-driven configuration validation."""

    def test_health_backoff_range_is_two_numeric_values(self):
        """
        Req: CONF-HEALTH-003
        Health backoff jitter range must be a 2-length numeric range used for backoff jittering.
        """
        cfg = ConfigManager().load_config()
        r = cfg.mediamtx.backoff_jitter_range
        assert isinstance(r, (list, tuple))
        assert len(r) == 2
        assert all(isinstance(x, (int, float)) for x in r)
