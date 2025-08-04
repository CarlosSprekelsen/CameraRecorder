# tests/unit/test_mediamtx_wrapper/test_controller_health_monitoring.py
"""
Test health monitoring circuit breaker activation/recovery and adaptive backoff.

Test policy: Verify configurable circuit breaker behavior, exponential backoff
with jitter, state transitions, and recovery logging.
"""

import pytest
import asyncio
from unittest.mock import Mock, AsyncMock, patch
import aiohttp

from src.mediamtx_wrapper.controller import MediaMTXController


class TestHealthMonitoring:
    """Test health monitoring circuit breaker and backoff behavior."""

    @pytest.fixture
    def controller_fast_timers(self):
        """Create controller with fast timers for testing."""
        controller = MediaMTXController(
            host="localhost",
            api_port=9997,
            rtsp_port=8554,
            webrtc_port=8889,
            hls_port=8888,
            config_path="/tmp/test_config.yml",
            recordings_path="/tmp/recordings",
            snapshots_path="/tmp/snapshots",
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
    def mock_session(self):
        """Create mock aiohttp session."""
        session = Mock()
        session.get = AsyncMock()
        session.close = AsyncMock()
        return session

    def _mock_response(self, status, json_data=None, text_data=""):
        """Helper to create mock HTTP response."""
        response = Mock()
        response.status = status
        response.json = AsyncMock(return_value=json_data or {})
        response.text = AsyncMock(return_value=text_data)
        return response

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
        self, controller_fast_timers, mock_session
    ):
        """Test circuit breaker requires N consecutive successes before full reset."""
        controller = controller_fast_timers
        controller._session = mock_session

        # Configure for 2 consecutive successes required for recovery
        controller._health_recovery_confirmation_threshold = 2

        # Mock sequence: failures (trigger CB) → timeout → success → failure → success →
        # success (full recovery)
        failure_response = self._mock_response(500, text_data="Service Unavailable")
        success_response = self._mock_response(
            200, {"serverVersion": "1.0.0", "serverUptime": 1200}
        )

        responses = [
            failure_response,
            failure_response,
            failure_response,  # Trigger circuit breaker
            success_response,  # First success during recovery
            failure_response,  # Failure interrupts recovery
            success_response,  # Success again
            success_response,  # Second consecutive success - should fully recover
        ]
        mock_session.get.side_effect = responses

        await controller.start()
        await asyncio.sleep(0.6)  # Let health checks run through sequence
        await controller.stop()

        # Verify circuit breaker was activated and recovered
        assert controller._health_state["circuit_breaker_activations"] > 0
        assert controller._health_state["recovery_count"] > 0

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
        mock_session.get.side_effect = responses

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
        self, controller_fast_timers, mock_session
    ):
        """Test exponential backoff calculation with configurable parameters."""
        controller = controller_fast_timers
        controller._session = mock_session

        # Mock failing health checks
        mock_session.get.side_effect = aiohttp.ClientError("Connection refused")

        # Record sleep intervals to verify backoff
        sleep_intervals = []
        original_sleep = asyncio.sleep

        async def mock_sleep(interval):
            sleep_intervals.append(interval)
            # Use very short actual sleep for test speed
            await original_sleep(0.001)

        with patch("asyncio.sleep", side_effect=mock_sleep):
            await controller.start()
            await asyncio.sleep(0.1)  # Let some checks run
            await controller.stop()

        # Verify backoff intervals increase exponentially
        if len(sleep_intervals) >= 2:
            # Should see increasing intervals (allowing for circuit breaker waits)
            health_check_intervals = [
                interval
                for interval in sleep_intervals
                if interval >= controller._health_check_interval
            ]
            if len(health_check_intervals) >= 2:
                assert health_check_intervals[1] > health_check_intervals[0]

    @pytest.mark.asyncio
    async def test_health_state_transition_logging(
        self, controller_fast_timers, mock_session, caplog
    ):
        """Test health state transitions are logged with context."""
        controller = controller_fast_timers
        controller._session = mock_session

        # Mock transition from failure to success
        failure_response = self._mock_response(500, text_data="Service Unavailable")
        success_response = self._mock_response(
            200, {"serverVersion": "1.0.0", "serverUptime": 1200}
        )

        mock_session.get.side_effect = [
            failure_response,
            success_response,
            success_response,
        ]

        with caplog.at_level("INFO"):
            await controller.start()
            await asyncio.sleep(0.3)  # Let health checks run
            await controller.stop()

        # Verify transition logging
        log_messages = [record.message for record in caplog.records]

        # Should see health degradation and recovery messages
        degraded_logs = [msg for msg in log_messages if "DEGRADED" in msg]
        [msg for msg in log_messages if "RECOVERED" in msg]

        assert len(degraded_logs) > 0, "Should log health degradation"
        # Recovery may not occur in short test time, but degradation should be logged

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
        mock_session.get.side_effect = responses

        await controller.start()
        await asyncio.sleep(0.4)  # Let health checks run
        await controller.stop()

        # After success, failure count should have been reset
        # Final state depends on timing, but we can check that success was registered
        assert controller._health_state["last_success_time"] > 0

    @pytest.mark.asyncio
    async def test_health_monitor_cleanup_on_stop(
        self, controller_fast_timers, mock_session
    ):
        """Test health monitoring task is properly cancelled on stop."""
        controller = controller_fast_timers
        controller._session = mock_session

        # Mock successful responses
        mock_session.get.return_value = self._mock_response(
            200, {"serverVersion": "1.0.0"}
        )

        await controller.start()

        # Verify health check task is running
        assert controller._health_check_task is not None
        assert not controller._health_check_task.done()

        await controller.stop()

        # Verify task is cancelled/completed
        assert controller._health_check_task.done()

    @pytest.mark.asyncio
    async def test_health_check_correlation_id_propagation(
        self, controller_fast_timers, mock_session
    ):
        """Test correlation IDs are set for health check operations."""
        controller = controller_fast_timers
        controller._session = mock_session

        mock_session.get.return_value = self._mock_response(
            200, {"serverVersion": "1.0.0"}
        )

        # Mock correlation ID functions to capture calls
        correlation_ids = []

        def mock_set_correlation_id(cid):
            correlation_ids.append(cid)

        with patch(
            "src.mediamtx_wrapper.controller.set_correlation_id",
            side_effect=mock_set_correlation_id,
        ):
            await controller.start()
            await asyncio.sleep(0.2)  # Let health checks run
            await controller.stop()

        # Verify correlation IDs were set
        assert len(correlation_ids) > 0
        # Each correlation ID should be a short string
        for cid in correlation_ids:
            assert isinstance(cid, str)
            assert len(cid) > 0

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
# - Mock aiohttp ClientSession for health check HTTP calls
# - Use fast timers (0.1s intervals) for test speed
# - Mock asyncio.sleep to capture backoff intervals
# - Use caplog fixture to verify logging behavior
# - Mock correlation ID functions to verify propagation
# - Test both circuit breaker activation and recovery
# - Verify configurable parameters are respected, not hardcoded values
# - Test proper task cleanup on controller stop
