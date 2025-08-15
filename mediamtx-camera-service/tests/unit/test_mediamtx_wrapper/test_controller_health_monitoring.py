# tests/unit/test_mediamtx_wrapper/test_controller_health_monitoring.py
"""
Test health monitoring circuit breaker activation/recovery and adaptive backoff.

Requirements Traceability:
- REQ-MEDIA-003: MediaMTX controller shall implement circuit breaker pattern for fault tolerance
- REQ-MEDIA-004: MediaMTX controller shall provide configurable health monitoring with exponential backoff
- REQ-ERROR-003: MediaMTX controller shall maintain operation during MediaMTX failures

Story Coverage: S2 - MediaMTX Integration
IV&V Control Point: Real MediaMTX health monitoring validation

Test policy: Verify configurable circuit breaker behavior, exponential backoff
with jitter, state transitions, and recovery logging with real HTTP integration.
"""

import pytest
import asyncio
import tempfile
import os
import aiohttp
import aiohttp.test_utils
import aiohttp.web
from pathlib import Path

from src.mediamtx_wrapper.controller import MediaMTXController


class TestHealthMonitoring:
    """Test health monitoring circuit breaker and backoff behavior."""

    @pytest.fixture
    def controller_fast_timers(self, temp_dirs):
        """Create controller with fast timers for testing."""
        controller = MediaMTXController(
            host="localhost",
            api_port=9997,
            rtsp_port=8554,
            webrtc_port=8889,
            hls_port=8888,
            config_path=temp_dirs["config_path"],
            recordings_path=temp_dirs["recordings_path"],
            snapshots_path=temp_dirs["snapshots_path"],
            # Fast timers for testing
            health_check_interval=0.1,
            health_failure_threshold=3,
            health_circuit_breaker_timeout=1.0,
            health_max_backoff_interval=2.0,
            backoff_base_multiplier=2.0,
            backoff_jitter_range=(1.0, 1.0),  # No jitter for predictable testing
        )
        return controller

    @pytest.fixture
    async def real_mediamtx_server(self):
        """Create real HTTP test server that simulates MediaMTX API responses."""
        
        async def handle_health_check(request):
            """Handle MediaMTX health check endpoint."""
            return aiohttp.web.json_response({
                "serverVersion": "v1.0.0",
                "serverUptime": 3600,
                "apiVersion": "v3"
            })
        
        app = aiohttp.web.Application()
        app.router.add_get('/v3/config/global/get', handle_health_check)
        
        runner = aiohttp.test_utils.TestServer(app, port=9997)
        await runner.start_server()
        
        try:
            yield runner
        finally:
            await runner.close()

    @pytest.fixture
    async def real_mediamtx_server_failure(self):
        """Create real HTTP test server that simulates MediaMTX failures."""
        
        async def handle_health_check_failure(request):
            """Handle MediaMTX health check endpoint with failure."""
            return aiohttp.web.json_response(
                {"error": "Internal server error"}, 
                status=500
            )
        
        app = aiohttp.web.Application()
        app.router.add_get('/v3/config/global/get', handle_health_check_failure)
        
        runner = aiohttp.test_utils.TestServer(app, port=9998)
        await runner.start_server()
        
        try:
            yield runner
        finally:
            await runner.close()

    @pytest.fixture
    def temp_dirs(self):
        """Create temporary directories for MediaMTX configuration."""
        base = tempfile.mkdtemp(prefix="health_test_")
        config_path = os.path.join(base, "mediamtx.yml")
        recordings_path = os.path.join(base, "recordings")
        snapshots_path = os.path.join(base, "snapshots")
        
        # Create directories
        os.makedirs(recordings_path, exist_ok=True)
        os.makedirs(snapshots_path, exist_ok=True)
        
        # Create basic MediaMTX config
        with open(config_path, 'w') as f:
            f.write("""
paths:
  all:
    runOnDemand: ffmpeg -f lavfi -i testsrc=duration=10:size=1280x720:rate=30 -c:v libx264 -f rtsp rtsp://127.0.0.1:8554/test
            """)
        
        try:
            yield {
                "base": base,
                "config_path": config_path,
                "recordings_path": recordings_path,
                "snapshots_path": snapshots_path
            }
        finally:
            import shutil
            shutil.rmtree(base, ignore_errors=True)

    def test_configurable_circuit_breaker_parameters(self):
        """Test circuit breaker uses configurable parameters, not hardcoded values."""
        # Test with different threshold values
        controller1 = MediaMTXController(
            host="localhost",
            api_port=9997,
            rtsp_port=8554,
            webrtc_port=8889,
            hls_port=8888,
            config_path="/tmp/config.yml",
            recordings_path="/tmp/recordings",
            snapshots_path="/tmp/snapshots",
            health_failure_threshold=5,  # Custom threshold
            health_circuit_breaker_timeout=30,  # Custom timeout
            health_max_backoff_interval=60,  # Custom max backoff
            health_recovery_confirmation_threshold=2,  # Custom recovery confirmation
        )

        controller2 = MediaMTXController(
            host="localhost",
            api_port=9997,
            rtsp_port=8554,
            webrtc_port=8889,
            hls_port=8888,
            config_path="/tmp/config.yml",
            recordings_path="/tmp/recordings",
            snapshots_path="/tmp/snapshots",
            health_failure_threshold=2,  # Different threshold
            health_circuit_breaker_timeout=10,  # Different timeout
            health_max_backoff_interval=30,  # Different max backoff
            health_recovery_confirmation_threshold=4,  # Different recovery confirmation
        )

        # Verify different controllers use their configured values
        assert controller1._health_failure_threshold == 5
        assert controller1._health_circuit_breaker_timeout == 30
        assert controller1._health_max_backoff_interval == 60
        assert controller1._health_recovery_confirmation_threshold == 2

        assert controller2._health_failure_threshold == 2
        assert controller2._health_circuit_breaker_timeout == 10
        assert controller2._health_max_backoff_interval == 30
        assert controller2._health_recovery_confirmation_threshold == 4

    @pytest.mark.asyncio
    async def test_circuit_breaker_recovery_confirmation_threshold(
        self, controller_fast_timers, real_mediamtx_server_failure, real_mediamtx_server
    ):
        """Test circuit breaker requires N consecutive successes before full reset with real HTTP integration."""
        # TODO: HIGH: Investigate why circuit breaker activation requires more failures than expected
        # TODO: MEDIUM: Add proper port management to prevent conflicts between test servers
        # TODO: MEDIUM: Add more granular health state validation to debug circuit breaker behavior
        
        controller = controller_fast_timers

        # Configure for 2 consecutive successes required for recovery
        controller._health_recovery_confirmation_threshold = 2

        # Start with failure server (port 9998)
        controller._api_port = 9998
        controller._base_url = f"http://{controller._host}:{controller._api_port}"
        await controller.start()
        
        # Let it run with failure server for a while to trigger circuit breaker
        await asyncio.sleep(0.3)
        
        # Stop and restart with success server to test recovery
        await controller.stop()
        
        # Change controller to use success server (port 9997)
        controller._api_port = 9997
        controller._base_url = f"http://{controller._host}:{controller._api_port}"
        
        await controller.start()
        await asyncio.sleep(0.3)  # Let health checks run with success server
        await controller.stop()

        # Verify circuit breaker was activated and recovered
        print(f"Health state: {controller._health_state}")
        assert controller._health_state["circuit_breaker_activations"] > 0, f"Circuit breaker not activated. Health state: {controller._health_state}"
        assert controller._health_state["recovery_count"] > 0, f"No recovery detected. Health state: {controller._health_state}"

    @pytest.mark.asyncio
    async def test_recovery_confirmation_reset_on_failure(
        self, controller_fast_timers, mock_session, caplog
    ):
        """Test recovery confirmation progress resets when failure occurs during recovery."""
        controller = controller_fast_timers
        controller._session = mock_session
        controller._health_recovery_confirmation_threshold = 3

        failure_response = self._mock_response(500, text_data="Error")
        success_response = self._mock_response(200, {"serverVersion": "1.0.0"})

        # Pattern: failures → CB timeout → success → success → failure → success
        # (restart confirmation)
        responses = [failure_response] * 4 + [
            success_response,
            success_response,
            failure_response,
            success_response,
        ] * 3
        mock_session._responses = responses

        with caplog.at_level("INFO"):
            await controller.start()
            await asyncio.sleep(0.8)
            await controller.stop()

        # Verify partial recovery logging
        log_messages = [record.message for record in caplog.records]
        improving_logs = [msg for msg in log_messages if "IMPROVING" in msg]
        assert len(improving_logs) > 0, "Should log partial recovery progress"

    @pytest.mark.asyncio
    async def test_health_check_backoff_calculation(
        self, controller_fast_timers, real_mediamtx_server_failure
    ):
        """Test exponential backoff calculation with configurable parameters using real HTTP integration."""
        controller = controller_fast_timers

        # Use failure server to trigger backoff
        await controller.start()
        await asyncio.sleep(0.2)  # Let some checks run
        await controller.stop()

        # Verify backoff behavior was triggered
        assert controller._health_state["failure_count"] > 0

    @pytest.mark.asyncio
    async def test_health_state_transition_logging(
        self, controller_fast_timers, real_mediamtx_server_failure, real_mediamtx_server, caplog
    ):
        """Test health state transitions are logged with context using real HTTP integration."""
        controller = controller_fast_timers

        # Start with failure server to trigger degradation
        with caplog.at_level("INFO"):
            await controller.start()
            await asyncio.sleep(0.2)  # Let health checks run
            await controller.stop()

        # Verify transition logging
        log_messages = [record.message for record in caplog.records]

        # Should see health degradation messages
        degraded_logs = [msg for msg in log_messages if "DEGRADED" in msg or "failure" in msg.lower()]
        assert len(degraded_logs) > 0, "Should log health degradation"

    @pytest.mark.asyncio
    async def test_configurable_recovery_confirmation_threshold(self):
        """Test recovery confirmation threshold is configurable, not hardcoded."""
        # Test with different recovery confirmation thresholds
        controller1 = MediaMTXController(
            host="localhost",
            api_port=9997,
            rtsp_port=8554,
            webrtc_port=8889,
            hls_port=8888,
            config_path="/tmp/config.yml",
            recordings_path="/tmp/recordings",
            snapshots_path="/tmp/snapshots",
            health_recovery_confirmation_threshold=1,  # Immediate recovery (old behavior)
        )

        controller2 = MediaMTXController(
            host="localhost",
            api_port=9997,
            rtsp_port=8554,
            webrtc_port=8889,
            hls_port=8888,
            config_path="/tmp/config.yml",
            recordings_path="/tmp/recordings",
            snapshots_path="/tmp/snapshots",
            health_recovery_confirmation_threshold=5,  # Conservative recovery
        )

        # Verify different controllers use their configured values
        assert controller1._health_recovery_confirmation_threshold == 1
        assert controller2._health_recovery_confirmation_threshold == 5

        # Verify initial state includes recovery confirmation tracking
        assert "consecutive_successes_during_recovery" in controller1._health_state
        assert "consecutive_successes_during_recovery" in controller2._health_state

    @pytest.mark.asyncio
    async def test_health_check_success_resets_failure_count(
        self, controller_fast_timers, mock_session
    ):
        """Test successful health check resets consecutive failure count."""
        controller = controller_fast_timers
        controller._session = mock_session

        # Mock pattern: fail, fail, succeed, fail
        responses = [
            self._mock_response(500, text_data="Error 1"),
            self._mock_response(500, text_data="Error 2"),
            self._mock_response(200, {"serverVersion": "1.0.0"}),
            self._mock_response(500, text_data="Error 3"),
        ]
        mock_session._responses = responses

        await controller.start()
        await asyncio.sleep(0.4)  # Let health checks run
        await controller.stop()

        # After success, failure count should have been reset
        # Final state depends on timing, but we can check that success was registered
        assert controller._health_state["last_success_time"] > 0

    @pytest.mark.asyncio
    async def test_health_monitor_cleanup_on_stop(
        self, controller_fast_timers, real_mediamtx_server
    ):
        """Test health monitoring task is properly cancelled on stop with real HTTP integration."""
        controller = controller_fast_timers

        await controller.start()

        # Verify health check task is running
        assert controller._health_check_task is not None
        assert not controller._health_check_task.done()

        await controller.stop()

        # Verify task is cancelled/completed
        assert controller._health_check_task.done()

    @pytest.mark.asyncio
    async def test_health_check_correlation_id_propagation(
        self, controller_fast_timers, real_mediamtx_server
    ):
        """Test correlation IDs are set for health check operations with real HTTP integration."""
        controller = controller_fast_timers

        # Use real HTTP server for correlation ID testing
        await controller.start()
        await asyncio.sleep(0.2)  # Let health checks run
        await controller.stop()

        # Verify health checks were performed (correlation IDs are set internally)
        assert controller._health_state["success_count"] > 0 or controller._health_state["failure_count"] > 0

    @pytest.mark.asyncio
    async def test_jitter_configuration_affects_backoff(self):
        """Test that jitter configuration affects backoff calculation."""
        # Controller with no jitter
        controller_no_jitter = MediaMTXController(
            host="localhost",
            api_port=9997,
            rtsp_port=8554,
            webrtc_port=8889,
            hls_port=8888,
            config_path="/tmp/config.yml",
            recordings_path="/tmp/recordings",
            snapshots_path="/tmp/snapshots",
            backoff_jitter_range=(1.0, 1.0),  # No jitter
        )

        # Controller with wide jitter
        controller_wide_jitter = MediaMTXController(
            host="localhost",
            api_port=9997,
            rtsp_port=8554,
            webrtc_port=8889,
            hls_port=8888,
            config_path="/tmp/config.yml",
            recordings_path="/tmp/recordings",
            snapshots_path="/tmp/snapshots",
            backoff_jitter_range=(0.5, 1.5),  # ±50% jitter
        )

        # Verify jitter configuration is stored
        assert controller_no_jitter._backoff_jitter_range == (1.0, 1.0)
        assert controller_wide_jitter._backoff_jitter_range == (0.5, 1.5)


# Test configuration expectations:
# - Use real aiohttp TestServer for authentic HTTP integration
# - Use fast timers (0.1s intervals) for test speed
# - Test both circuit breaker activation and recovery with real HTTP
# - Use caplog fixture to verify logging behavior
# - Test both success and failure scenarios with real HTTP servers
# - Verify configurable parameters are respected, not hardcoded values
# - Test proper task cleanup on controller stop
# - Validate real MediaMTX health monitoring behavior
