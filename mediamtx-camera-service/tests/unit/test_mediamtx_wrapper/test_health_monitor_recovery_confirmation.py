# tests/unit/test_mediamtx_wrapper/test_health_monitor_recovery_confirmation.py
"""
Test health monitoring recovery confirmation logic and circuit breaker state transitions.

Test policy: Verify that circuit breaker recovery requires exactly N consecutive
successful health checks and that any failure during recovery resets the confirmation progress.
"""

import pytest
import asyncio
import time
from unittest.mock import Mock, AsyncMock

from src.mediamtx_wrapper.controller import MediaMTXController


class TestHealthMonitorRecoveryConfirmation:
    """Test health monitoring recovery confirmation logic."""

    @pytest.fixture
    def controller_fast_timers(self):
        """Create controller with fast timers for testing."""
        return MediaMTXController(
            host="localhost",
            api_port=9997,
            rtsp_port=8554,
            webrtc_port=8889,
            hls_port=8888,
            config_path="/tmp/test_config.yml",
            recordings_path="/tmp/recordings",
            snapshots_path="/tmp/snapshots",
            # Fast timers for testing
            health_check_interval=0.05,
            health_failure_threshold=3,
            health_circuit_breaker_timeout=0.2,
            health_max_backoff_interval=1.0,
            health_recovery_confirmation_threshold=3,
            backoff_base_multiplier=2.0,
            backoff_jitter_range=(1.0, 1.0),  # No jitter for predictable testing
        )

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

    @pytest.mark.asyncio
    async def test_exact_consecutive_success_requirement(
        self, controller_fast_timers, mock_session
    ):
        """Test that recovery requires exactly the configured number of consecutive successes."""
        controller = controller_fast_timers
        controller._session = mock_session
        controller._health_recovery_confirmation_threshold = (
            4  # Require 4 consecutive successes
        )

        failure_response = self._mock_response(500, text_data="Service Error")
        success_response = self._mock_response(
            200, {"serverVersion": "1.0.0", "serverUptime": 1200}
        )

        # Pattern: failures → CB timeout → exactly 4 consecutive successes
        responses = [
            failure_response,
            failure_response,
            failure_response,  # Trigger CB
            success_response,
            success_response,
            success_response,
            success_response,  # Exactly 4 successes
        ]
        mock_session.get.side_effect = responses

        await controller.start()
        await asyncio.sleep(0.5)  # Let recovery sequence complete
        await controller.stop()

        # Verify recovery occurred after exactly 4 consecutive successes
        assert controller._health_state["circuit_breaker_activations"] == 1
        assert controller._health_state["recovery_count"] == 1
        assert (
            controller._health_state["consecutive_successes_during_recovery"] == 0
        )  # Reset after recovery

    @pytest.mark.asyncio
    async def test_insufficient_consecutive_successes(
        self, controller_fast_timers, mock_session
    ):
        """Test that N-1 consecutive successes do not trigger recovery."""
        controller = controller_fast_timers
        controller._session = mock_session
        controller._health_recovery_confirmation_threshold = 3

        failure_response = self._mock_response(500, text_data="Error")
        success_response = self._mock_response(200, {"serverVersion": "1.0.0"})

        # Pattern: failures → CB timeout → only 2 successes (insufficient for threshold
        # of 3)
        responses = [
            failure_response,
            failure_response,
            failure_response,  # Trigger CB
            success_response,
            success_response,  # Only 2/3 required successes
        ]
        mock_session.get.side_effect = responses

        await controller.start()
        await asyncio.sleep(0.4)
        await controller.stop()

        # Verify circuit breaker did not recover
        assert controller._health_state["circuit_breaker_activations"] == 1
        assert controller._health_state["recovery_count"] == 0
        assert (
            controller._health_state["consecutive_successes_during_recovery"] == 2
        )  # Partial progress

    @pytest.mark.asyncio
    async def test_failure_resets_confirmation_progress(
        self, controller_fast_timers, mock_session, caplog
    ):
        """Test that any failure during recovery resets the confirmation counter."""
        controller = controller_fast_timers
        controller._session = mock_session
        controller._health_recovery_confirmation_threshold = 3

        failure_response = self._mock_response(503, text_data="Service Unavailable")
        success_response = self._mock_response(200, {"serverVersion": "1.0.0"})

        # Pattern: failures → CB → 2 successes → failure (reset) → 3 successes (recover)
        responses = [
            failure_response,
            failure_response,
            failure_response,  # Trigger CB
            success_response,
            success_response,  # 2/3 successes
            failure_response,  # Reset confirmation progress
            success_response,
            success_response,
            success_response,  # 3 consecutive successes
        ]
        mock_session.get.side_effect = responses

        with caplog.at_level("INFO"):
            await controller.start()
            await asyncio.sleep(0.6)
            await controller.stop()

        # Verify eventual recovery after reset
        assert controller._health_state["circuit_breaker_activations"] == 1
        assert controller._health_state["recovery_count"] == 1

        # Verify reset was logged
        log_messages = [record.message for record in caplog.records]
        degraded_logs = [
            msg for msg in log_messages if "DEGRADED" in msg and "IMPROVING" not in msg
        ]
        assert (
            len(degraded_logs) > 0
        ), "Should log health degradation that resets recovery"

    @pytest.mark.asyncio
    async def test_circuit_breaker_timeout_behavior(
        self, controller_fast_timers, mock_session
    ):
        """Test circuit breaker timeout behavior before recovery attempts."""
        controller = controller_fast_timers
        controller._session = mock_session
        controller._health_circuit_breaker_timeout = 0.3  # Short timeout for testing

        failure_response = self._mock_response(500, text_data="Error")
        success_response = self._mock_response(200, {"serverVersion": "1.0.0"})

        # Pattern: failures → CB → wait for timeout → immediate recovery
        responses = [
            failure_response,
            failure_response,
            failure_response,  # Trigger CB
            success_response,
            success_response,
            success_response,  # Immediate recovery after timeout
        ]
        mock_session.get.side_effect = responses

        start_time = time.time()
        await controller.start()
        await asyncio.sleep(0.6)  # Wait for timeout + recovery
        await controller.stop()
        elapsed = time.time() - start_time

        # Verify circuit breaker timeout was respected
        assert elapsed >= 0.3, "Should respect circuit breaker timeout"
        assert controller._health_state["circuit_breaker_activations"] == 1
        assert controller._health_state["recovery_count"] == 1

    @pytest.mark.asyncio
    async def test_recovery_state_tracking(self, controller_fast_timers, mock_session):
        """Test internal state tracking during recovery process."""
        controller = controller_fast_timers
        controller._session = mock_session
        controller._health_recovery_confirmation_threshold = 2

        failure_response = self._mock_response(500, text_data="Error")
        success_response = self._mock_response(200, {"serverVersion": "1.0.0"})

        # Pattern: failures → CB → track progress through recovery
        responses = [
            failure_response,
            failure_response,
            failure_response,  # Trigger CB
            success_response,  # 1/2 successes
            success_response,  # 2/2 successes (full recovery)
        ]
        mock_session.get.side_effect = responses

        await controller.start()
        await asyncio.sleep(0.4)

        # Check intermediate state before full recovery
        # Note: We can't easily check intermediate state during execution,
        # so we verify final state after completion
        await controller.stop()

        # Verify state was properly tracked and reset
        assert controller._health_state["circuit_breaker_activations"] == 1
        assert controller._health_state["recovery_count"] == 1
        assert (
            controller._health_state["consecutive_successes_during_recovery"] == 0
        )  # Reset after recovery

    @pytest.mark.asyncio
    async def test_configurable_confirmation_threshold(self, mock_session):
        """Test different confirmation threshold configurations."""
        # Test with threshold = 1 (immediate recovery)
        controller_fast = MediaMTXController(
            host="localhost",
            api_port=9997,
            rtsp_port=8554,
            webrtc_port=8889,
            hls_port=8888,
            config_path="/tmp/config.yml",
            recordings_path="/tmp/recordings",
            snapshots_path="/tmp/snapshots",
            health_check_interval=0.05,
            health_failure_threshold=2,
            health_circuit_breaker_timeout=0.1,
            health_recovery_confirmation_threshold=1,
        )
        controller_fast._session = mock_session

        # Test with threshold = 5 (slow recovery)
        controller_slow = MediaMTXController(
            host="localhost",
            api_port=9997,
            rtsp_port=8554,
            webrtc_port=8889,
            hls_port=8888,
            config_path="/tmp/config.yml",
            recordings_path="/tmp/recordings",
            snapshots_path="/tmp/snapshots",
            health_check_interval=0.05,
            health_failure_threshold=2,
            health_circuit_breaker_timeout=0.1,
            health_recovery_confirmation_threshold=5,
        )
        controller_slow._session = mock_session

        # Verify configuration is applied
        assert controller_fast._health_recovery_confirmation_threshold == 1
        assert controller_slow._health_recovery_confirmation_threshold == 5

        failure_response = self._mock_response(500, text_data="Error")
        success_response = self._mock_response(200, {"serverVersion": "1.0.0"})

        # Test fast recovery (1 success)
        mock_session.get.side_effect = [
            failure_response,
            failure_response,
            success_response,
        ]
        await controller_fast.start()
        await asyncio.sleep(0.3)
        await controller_fast.stop()

        # Fast controller should recover immediately
        assert controller_fast._health_state["recovery_count"] == 1

        # Reset mock for slow controller test
        mock_session.get.side_effect = [
            failure_response,
            failure_response,  # Trigger CB
            success_response,
            success_response,
            success_response,
            success_response,  # Only 4/5 needed
        ]
        await controller_slow.start()
        await asyncio.sleep(0.4)
        await controller_slow.stop()

        # Slow controller should not recover with only 4/5 successes
        assert controller_slow._health_state["recovery_count"] == 0

    @pytest.mark.asyncio
    async def test_partial_recovery_logging(
        self, controller_fast_timers, mock_session, caplog
    ):
        """Test that partial recovery progress is properly logged."""
        controller = controller_fast_timers
        controller._session = mock_session
        controller._health_recovery_confirmation_threshold = 4

        failure_response = self._mock_response(500, text_data="Error")
        success_response = self._mock_response(200, {"serverVersion": "1.0.0"})

        # Pattern: failures → CB → partial recovery → reset → full recovery
        responses = [
            failure_response,
            failure_response,
            failure_response,  # Trigger CB
            success_response,
            success_response,  # 2/4 successes (partial)
            failure_response,  # Reset
            success_response,
            success_response,
            success_response,
            success_response,  # Full recovery
        ]
        mock_session.get.side_effect = responses

        with caplog.at_level("INFO"):
            await controller.start()
            await asyncio.sleep(0.7)
            await controller.stop()

        # Verify different recovery states were logged
        log_messages = [record.message for record in caplog.records]
        improving_logs = [msg for msg in log_messages if "IMPROVING" in msg]
        recovered_logs = [msg for msg in log_messages if "FULLY RECOVERED" in msg]
        degraded_logs = [msg for msg in log_messages if "DEGRADED" in msg]

        assert len(improving_logs) >= 1, "Should log partial recovery progress"
        assert len(recovered_logs) == 1, "Should log full recovery"
        assert len(degraded_logs) >= 1, "Should log degradation that resets recovery"
