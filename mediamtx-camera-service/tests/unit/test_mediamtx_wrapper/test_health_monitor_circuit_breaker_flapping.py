# tests/unit/test_mediamtx_wrapper/test_health_monitor_circuit_breaker_flapping.py
"""
Test health monitoring circuit breaker flapping scenarios and edge cases.

Test policy: Verify circuit breaker stability under alternating success/failure
patterns to prevent oscillation and ensure proper recovery confirmation logic.
"""

import pytest
import asyncio
from unittest.mock import Mock, AsyncMock

from src.mediamtx_wrapper.controller import MediaMTXController
from .async_mock_helpers import (
    create_mock_session,
    create_async_mock_with_response,
    create_async_mock_with_side_effect,
    MockResponse
)


class TestHealthMonitorFlapping:
    """Test health monitoring circuit breaker flapping resistance."""

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
            health_circuit_breaker_timeout=0.3,
            health_max_backoff_interval=1.0,
            health_recovery_confirmation_threshold=3,
            backoff_base_multiplier=2.0,
            backoff_jitter_range=(1.0, 1.0),  # No jitter for predictable testing
        )

    @pytest.fixture
    def mock_session(self):
        """Create mock aiohttp session with proper async context manager support."""
        return create_mock_session()

    def _mock_response(self, status, json_data=None, text_data=""):
        """Helper to create mock HTTP response."""
        return MockResponse(status, json_data, text_data)

    @pytest.mark.asyncio
    async def test_circuit_breaker_activation_threshold(
        self, controller_fast_timers, mock_session
    ):
        """Test circuit breaker opens exactly at configured failure threshold."""
        controller = controller_fast_timers
        controller._session = mock_session

        failure_response = self._mock_response(500, text_data="Service Error")
        success_response = self._mock_response(200, {"serverVersion": "1.0.0"})

        # Pattern: 2 failures (below threshold) → 1 success → 3 failures (trigger CB)
        responses = [
            failure_response,
            failure_response,  # 2 failures - should not trigger CB
            success_response,  # Reset consecutive failures
            failure_response,
            failure_response,
            failure_response,  # 3 failures - should trigger CB
        ]
        mock_session.get = create_async_mock_with_side_effect(
            lambda *args, **kwargs: responses.pop(0) if responses else MockResponse(200, {"status": "ok"})
        )

        await controller.start()
        await asyncio.sleep(0.4)  # Let sequence run
        await controller.stop()

        # Verify circuit breaker activated exactly once
        assert controller._health_state["circuit_breaker_activations"] == 1
        assert controller._health_state["consecutive_failures"] >= 3

    @pytest.mark.asyncio
    async def test_flapping_resistance_during_recovery(
        self, controller_fast_timers, mock_session, caplog
    ):
        """Test circuit breaker resists flapping during recovery phase."""
        controller = controller_fast_timers
        controller._session = mock_session
        controller._health_recovery_confirmation_threshold = 3

        failure_response = self._mock_response(500, text_data="Error")
        success_response = self._mock_response(200, {"serverVersion": "1.0.0"})

        # Pattern: failures → CB → timeout → alternating success/failure (should not
        # fully recover)
        responses = [
            failure_response,
            failure_response,
            failure_response,  # Trigger CB
            success_response,  # 1st success during recovery
            failure_response,  # Reset confirmation counter
            success_response,  # 1st success again
            failure_response,  # Reset confirmation counter again
            success_response,  # 1st success yet again
            success_response,  # 2nd consecutive success
            success_response,  # 3rd consecutive success - should fully recover
        ]
        mock_session.get = create_async_mock_with_side_effect(
            lambda *args, **kwargs: responses.pop(0) if responses else MockResponse(200, {"status": "ok"})
        )

        with caplog.at_level("INFO"):
            await controller.start()
            await asyncio.sleep(0.6)  # Let recovery sequence run
            await controller.stop()

        # Verify circuit breaker eventually recovered after stable successes
        assert controller._health_state["circuit_breaker_activations"] == 1
        assert controller._health_state["recovery_count"] == 1

        # Verify intermediate "IMPROVING" logs during partial recovery
        log_messages = [record.message for record in caplog.records]
        improving_logs = [msg for msg in log_messages if "IMPROVING" in msg]
        assert len(improving_logs) >= 2, "Should log multiple partial recovery attempts"

    @pytest.mark.asyncio
    async def test_rapid_flapping_scenario(self, controller_fast_timers, mock_session):
        """Test circuit breaker behavior under rapid success/failure alternation."""
        controller = controller_fast_timers
        controller._session = mock_session
        controller._health_recovery_confirmation_threshold = 2

        failure_response = self._mock_response(503, text_data="Unavailable")
        success_response = self._mock_response(200, {"serverVersion": "1.0.0"})

        # Pattern: failures → CB → rapid alternation → eventual stable recovery
        responses = [
            failure_response,
            failure_response,
            failure_response,  # Trigger CB
            # Rapid alternation during recovery (10 cycles)
            success_response,
            failure_response,  # Reset
            success_response,
            failure_response,  # Reset
            success_response,
            failure_response,  # Reset
            success_response,
            failure_response,  # Reset
            success_response,
            failure_response,  # Reset
            # Finally stable recovery
            success_response,
            success_response,  # Should fully recover
        ]
        mock_session.get = create_async_mock_with_side_effect(
            lambda *args, **kwargs: responses.pop(0) if responses else MockResponse(200, {"status": "ok"})
        )

        await controller.start()
        await asyncio.sleep(0.8)  # Extended time for rapid sequence
        await controller.stop()

        # Verify circuit breaker stayed stable during flapping
        assert controller._health_state["circuit_breaker_activations"] == 1
        assert controller._health_state["recovery_count"] == 1
        # Consecutive successes should be reset to 0 after recovery
        assert controller._health_state["consecutive_successes_during_recovery"] == 0

    @pytest.mark.asyncio
    async def test_multiple_circuit_breaker_cycles(
        self, controller_fast_timers, mock_session
    ):
        """Test multiple circuit breaker activation/recovery cycles."""
        controller = controller_fast_timers
        controller._session = mock_session
        controller._health_recovery_confirmation_threshold = 2

        failure_response = self._mock_response(500, text_data="Error")
        success_response = self._mock_response(200, {"serverVersion": "1.0.0"})

        # Pattern: CB cycle 1 → recovery → CB cycle 2 → recovery
        responses = [
            # First CB cycle
            failure_response,
            failure_response,
            failure_response,  # Trigger CB #1
            success_response,
            success_response,  # Recover from CB #1
            # Second CB cycle
            failure_response,
            failure_response,
            failure_response,  # Trigger CB #2
            success_response,
            success_response,  # Recover from CB #2
        ]
        mock_session.get = create_async_mock_with_side_effect(
            lambda *args, **kwargs: responses.pop(0) if responses else MockResponse(200, {"status": "ok"})
        )

        await controller.start()
        await asyncio.sleep(0.8)  # Extended time for two cycles
        await controller.stop()

        # Verify both circuit breaker cycles occurred
        assert controller._health_state["circuit_breaker_activations"] == 2
        assert controller._health_state["recovery_count"] == 2

    @pytest.mark.asyncio
    async def test_recovery_confirmation_boundary_conditions(
        self, controller_fast_timers, mock_session
    ):
        """Test recovery confirmation at exactly the threshold boundary."""
        controller = controller_fast_timers
        controller._session = mock_session
        controller._health_recovery_confirmation_threshold = 4  # Higher threshold

        failure_response = self._mock_response(500, text_data="Error")
        success_response = self._mock_response(200, {"serverVersion": "1.0.0"})

        # Pattern: failures → CB → exactly N-1 successes → failure → N successes
        responses = [
            failure_response,
            failure_response,
            failure_response,  # Trigger CB
            success_response,
            success_response,
            success_response,  # 3/4 successes (not enough)
            failure_response,  # Reset confirmation counter
            success_response,
            success_response,
            success_response,
            success_response,  # 4/4 successes (should recover)
        ]
        mock_session.get.side_effect = responses

        await controller.start()
        await asyncio.sleep(0.7)
        await controller.stop()

        # Verify recovery required exactly the configured threshold
        assert controller._health_state["circuit_breaker_activations"] == 1
        assert controller._health_state["recovery_count"] == 1

    @pytest.mark.asyncio
    async def test_no_premature_circuit_breaker_reset(
        self, controller_fast_timers, mock_session
    ):
        """Test that transient successes don't prematurely reset circuit breaker state."""
        controller = controller_fast_timers
        controller._session = mock_session
        controller._health_recovery_confirmation_threshold = 3

        failure_response = self._mock_response(500, text_data="Error")
        success_response = self._mock_response(200, {"serverVersion": "1.0.0"})

        # Pattern: failures → CB → single success → more failures (should not reset CB
        # state)
        responses = [
            failure_response,
            failure_response,
            failure_response,  # Trigger CB
            success_response,  # Single success (insufficient)
            failure_response,
            failure_response,  # More failures during recovery
        ]
        mock_session.get.side_effect = responses

        await controller.start()
        await asyncio.sleep(0.5)
        await controller.stop()

        # Verify circuit breaker activated but did not recover
        assert controller._health_state["circuit_breaker_activations"] == 1
        assert controller._health_state["recovery_count"] == 0
        # Consecutive successes should be reset to 0 after failure
        assert controller._health_state["consecutive_successes_during_recovery"] == 0
