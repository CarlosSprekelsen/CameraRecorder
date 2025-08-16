"""
MediaMTX Real Integration Test

Tests real MediaMTX API endpoint testing.
Validates actual health monitoring and stream management validation.

This test replaces complex unit test mocks with real system validation
to provide better confidence in MediaMTX integration reliability.

Requirements Traceability:
- REQ-ERROR-008: System shall handle MediaMTX service failures gracefully
- REQ-HEALTH-001: System shall provide comprehensive logging for health monitoring
- REQ-HEALTH-002: System shall support structured logging for production environments
- REQ-HEALTH-003: System shall enable correlation ID tracking across components

Story Coverage: S1 - MediaMTX Integration
IV&V Control Point: Real MediaMTX service integration validation
"""

import asyncio
import aiohttp
import os
import pytest
import subprocess
import sys
import tempfile
import time

# Import the actual MediaMTX controller implementation
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', '..', 'src'))

from mediamtx_wrapper.controller import MediaMTXController, StreamConfig  # noqa: E402


@pytest.fixture(scope="function")
async def mediamtx_controller():
    """
    Create and manage MediaMTX controller for testing.

    Requirements Traceability:
    - REQ-ERROR-008: System shall handle MediaMTX service failures gracefully

    Story Coverage: S1 - MediaMTX Integration
    IV&V Control Point: MediaMTX controller lifecycle validation
    """
    # Create temporary directories for testing
    with tempfile.TemporaryDirectory() as temp_dir:
        recordings_path = os.path.join(temp_dir, "recordings")
        snapshots_path = os.path.join(temp_dir, "snapshots")
        config_path = os.path.join(temp_dir, "mediamtx.yml")

        os.makedirs(recordings_path, exist_ok=True)
        os.makedirs(snapshots_path, exist_ok=True)

        controller = MediaMTXController(
            host="localhost",
            api_port=9997,
            rtsp_port=8554,
            webrtc_port=8889,
            hls_port=8888,
            config_path=config_path,
            recordings_path=recordings_path,
            snapshots_path=snapshots_path
        )

        # Start controller
        await controller.start()

        try:
            yield controller
        finally:
            # Cleanup: stop controller
            await controller.stop()


@pytest.fixture(scope="function")
async def mediamtx_server():
    """
    Verify systemd-managed MediaMTX service is running.

    This fixture follows the architectural decision AD-001: Single Systemd-Managed MediaMTX Instance.
    Tests MUST use the single systemd-managed MediaMTX service instance.
    Tests MUST NOT create multiple MediaMTX instances or start their own MediaMTX processes.

    Requirements Traceability:
    - REQ-ERROR-008: System shall handle MediaMTX service failures gracefully
    - REQ-HEALTH-001: System shall provide comprehensive logging for health monitoring

    Story Coverage: S1 - MediaMTX Integration
    IV&V Control Point: MediaMTX service availability validation
    """
    try:
        # Check if MediaMTX service is running via systemd
        result = subprocess.run(
            ['systemctl', 'is-active', 'mediamtx'],
            capture_output=True,
            text=True,
            timeout=10
        )

        if result.returncode != 0 or result.stdout.strip() != 'active':
            raise RuntimeError(
                "MediaMTX systemd service is not running. "
                "Please start it with: sudo systemctl start mediamtx"
            )

        # Wait for MediaMTX API to be ready
        await _wait_for_mediamtx_ready()

        print("✓ Using systemd-managed MediaMTX service")
        yield True

    except Exception as e:
        print(f"❌ MediaMTX service check failed: {e}")
        yield False


async def _wait_for_mediamtx_ready(timeout: float = 10.0) -> None:
    """Wait for MediaMTX service to be ready."""
    health_check_url = "http://localhost:9997/v3/config/global/get"

    start_time = time.time()
    while time.time() - start_time < timeout:
        try:
            async with aiohttp.ClientSession() as session:
                async with session.get(health_check_url) as response:
                    if response.status == 200:
                        print("✓ MediaMTX health check passed")
                        return
        except Exception:
            pass

        await asyncio.sleep(0.5)

    raise TimeoutError(f"MediaMTX service failed to be ready within {timeout} seconds")


class TestMediaMTXRealIntegration:
    """Test real MediaMTX integration."""

    @pytest.mark.asyncio
    async def test_mediamtx_real_integration(self):
        """
        Test real MediaMTX integration without fixtures.

        Requirements Traceability:
        - REQ-HEALTH-001: System shall provide comprehensive logging for health monitoring
        - REQ-HEALTH-003: System shall enable correlation ID tracking across components

        Story Coverage: S1 - MediaMTX Integration
        IV&V Control Point: Real MediaMTX health monitoring validation
        """
        # Create controller directly (following working pattern)
        with tempfile.TemporaryDirectory() as temp_dir:
            recordings_path = os.path.join(temp_dir, "recordings")
            snapshots_path = os.path.join(temp_dir, "snapshots")
            config_path = os.path.join(temp_dir, "mediamtx.yml")

            os.makedirs(recordings_path, exist_ok=True)
            os.makedirs(snapshots_path, exist_ok=True)

            controller = MediaMTXController(
                host="localhost",
                api_port=9997,
                rtsp_port=8554,
                webrtc_port=8889,
                hls_port=8888,
                config_path=config_path,
                recordings_path=recordings_path,
                snapshots_path=snapshots_path
            )

            # Start controller
            await controller.start()

            try:
                # Test real health check
                health_status = await controller.health_check()
                assert "status" in health_status
                assert health_status["status"] in ["healthy", "degraded", "unhealthy"]
                assert "api_port" in health_status
                assert health_status["api_port"] == 9997

                print("✓ MediaMTX real integration test passed")

            finally:
                # Cleanup
                await controller.stop()

    @pytest.mark.asyncio
    async def test_mediamtx_api_endpoints(self):
        """
        Test real MediaMTX API endpoints.

        Requirements Traceability:
        - REQ-ERROR-008: System shall handle MediaMTX service failures gracefully
        - REQ-HEALTH-002: System shall support structured logging for production environments

        Story Coverage: S1 - MediaMTX Integration
        IV&V Control Point: MediaMTX API endpoint validation
        """
        # Test against actual MediaMTX API
        api_port = 9997  # Use default port for existing server

        async with aiohttp.ClientSession() as session:
            # Test global config endpoint
            async with session.get(f'http://localhost:{api_port}/v3/config/global/get') as response:
                assert response.status == 200
                config_data = await response.json()
                assert "api" in config_data
                assert config_data["api"] is True
                print(f"✓ MediaMTX API accessible, API enabled: {config_data.get('api')}")

            # Test paths list endpoint
            async with session.get(f'http://localhost:{api_port}/v3/paths/list') as response:
                assert response.status == 200
                paths_data = await response.json()
                assert isinstance(paths_data, dict)
                print("✓ MediaMTX paths endpoint accessible")

    @pytest.mark.asyncio
    async def test_mediamtx_controller_lifecycle(self):
        """
        Test MediaMTX controller startup and shutdown lifecycle.

        Requirements Traceability:
        - REQ-ERROR-008: System shall handle MediaMTX service failures gracefully

        Story Coverage: S1 - MediaMTX Integration
        IV&V Control Point: MediaMTX controller lifecycle validation
        """
        with tempfile.TemporaryDirectory() as temp_dir:
            recordings_path = os.path.join(temp_dir, "recordings")
            snapshots_path = os.path.join(temp_dir, "snapshots")
            config_path = os.path.join(temp_dir, "mediamtx.yml")

            os.makedirs(recordings_path, exist_ok=True)
            os.makedirs(snapshots_path, exist_ok=True)

            controller = MediaMTXController(
                host="localhost",
                api_port=9997,
                rtsp_port=8554,
                webrtc_port=8889,
                hls_port=8888,
                config_path=config_path,
                recordings_path=recordings_path,
                snapshots_path=snapshots_path
            )

            # Test controller startup
            await controller.start()
            assert hasattr(controller, '_session')
            assert controller._session is not None

            # Test controller shutdown
            await controller.stop()
            assert controller._session is None

    @pytest.mark.asyncio
    async def test_mediamtx_stream_management(self):
        """
        Test MediaMTX stream management capabilities.

        Requirements Traceability:
        - REQ-ERROR-008: System shall handle MediaMTX service failures gracefully
        - REQ-HEALTH-001: System shall provide comprehensive logging for health monitoring

        Story Coverage: S1 - MediaMTX Integration
        IV&V Control Point: MediaMTX stream management validation
        """
        # Create controller directly
        with tempfile.TemporaryDirectory() as temp_dir:
            recordings_path = os.path.join(temp_dir, "recordings")
            snapshots_path = os.path.join(temp_dir, "snapshots")
            config_path = os.path.join(temp_dir, "mediamtx.yml")

            os.makedirs(recordings_path, exist_ok=True)
            os.makedirs(snapshots_path, exist_ok=True)

            controller = MediaMTXController(
                host="localhost",
                api_port=9997,
                rtsp_port=8554,
                webrtc_port=8889,
                hls_port=8888,
                config_path=config_path,
                recordings_path=recordings_path,
                snapshots_path=snapshots_path
            )

            await controller.start()

            try:
                # Test stream list retrieval
                streams = await controller.get_stream_list()
                assert isinstance(streams, list)

                # Test stream creation
                stream_config = StreamConfig(
                    name="test_stream",
                    source="rtsp://localhost:8554/test",
                    record=False
                )

                result = await controller.create_stream(stream_config)
                assert isinstance(result, dict)
                print(f"✓ Stream creation successful: {result}")

            finally:
                await controller.stop()

    @pytest.mark.asyncio
    async def test_mediamtx_health_monitoring(self):
        """
        Test MediaMTX health monitoring behavior.

        Requirements Traceability:
        - REQ-HEALTH-001: System shall provide comprehensive logging for health monitoring
        - REQ-HEALTH-003: System shall enable correlation ID tracking across components

        Story Coverage: S1 - MediaMTX Integration
        IV&V Control Point: MediaMTX health monitoring validation
        """
        # Create controller directly
        with tempfile.TemporaryDirectory() as temp_dir:
            recordings_path = os.path.join(temp_dir, "recordings")
            snapshots_path = os.path.join(temp_dir, "snapshots")
            config_path = os.path.join(temp_dir, "mediamtx.yml")

            os.makedirs(recordings_path, exist_ok=True)
            os.makedirs(snapshots_path, exist_ok=True)

            controller = MediaMTXController(
                host="localhost",
                api_port=9997,
                rtsp_port=8554,
                webrtc_port=8889,
                hls_port=8888,
                config_path=config_path,
                recordings_path=recordings_path,
                snapshots_path=snapshots_path
            )

            await controller.start()

            try:
                # Test health check with detailed response
                health_status = await controller.health_check()

                # Validate health response structure
                required_fields = ["status", "api_port", "correlation_id"]
                for field in required_fields:
                    assert field in health_status

                # Test health state tracking
                if hasattr(controller, '_health_state'):
                    health_state = controller._health_state
                    assert "total_checks" in health_state
                    assert "consecutive_failures" in health_state

            finally:
                await controller.stop()
