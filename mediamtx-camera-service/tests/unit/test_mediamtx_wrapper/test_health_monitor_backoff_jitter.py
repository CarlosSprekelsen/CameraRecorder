# tests/unit/test_mediamtx_wrapper/test_health_monitor_backoff_jitter.py
"""
Test health monitoring exponential backoff and jitter behavior.

Test policy: Verify configurable exponential backoff calculations, jitter application,
and maximum backoff limits are properly enforced during error conditions.
"""

import pytest
import asyncio
import time
import random
from unittest.mock import Mock, AsyncMock, patch
import aiohttp

from src.mediamtx_wrapper.controller import MediaMTXController


class TestHealthMonitorBackoffJitter:
    """Test health monitoring exponential backoff and jitter behavior."""

    @pytest.fixture
    def controller_backoff_test(self):
        """Create controller configured for backoff testing."""
        return MediaMTXController(
            host="localhost",
            api_port=9997,
            rtsp_port=8554,
            webrtc_port=8889,
            hls_port=8888,
            config_path="/tmp/test_config.yml",
            recordings_path="/tmp/recordings",
            snapshots_path="/tmp/snapshots",
            # Backoff-specific configuration
            health_check_interval=0.1,  # Base interval
            health_failure_threshold=5,  # Higher threshold to avoid CB during backoff test
            health_max_backoff_interval=2.0,  # Max backoff cap
            backoff_base_multiplier=2.0,  # Double each time
            backoff_jitter_range=(1.0, 1.0),  # No jitter for predictable testing
            health_circuit_breaker_timeout=10.0,  # Long timeout to focus on backoff
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
    async def test_exponential_backoff_calculation(
        self, controller_backoff_test, mock_session
    ):
        """Test exponential backoff interval calculation."""
        controller = controller_backoff_test
        controller._session = mock_session

        failure_response = self._mock_response(500, text_data="Service Error")

        # All failures to trigger exponential backoff
        responses = [failure_response] * 10
        mock_session.get.side_effect = responses

        # Mock asyncio.sleep to capture sleep intervals
        sleep_intervals = []

        async def mock_sleep(interval):
            sleep_intervals.append(interval)
            await asyncio.sleep(0.01)  # Short actual sleep for test speed

        with patch("asyncio.sleep", side_effect=mock_sleep):
            await controller.start()
            await asyncio.sleep(0.3)  # Let several failures occur
            await controller.stop()

        # Verify exponential backoff pattern
        # Expected intervals: 0.1, 0.2, 0.4, 0.8, 1.6, 2.0 (capped), 2.0, ...
        # Note: The sleep intervals include both health check intervals and error backoffs
        error_backoffs = [interval for interval in sleep_intervals if interval > 0.1]

        assert len(error_backoffs) >= 3, "Should have multiple error backoff intervals"

        # Verify exponential growth up to the cap
        if len(error_backoffs) >= 3:
            # First few should show exponential growth
            assert (
                error_backoffs[1] > error_backoffs[0]
            ), "Second backoff should be longer than first"
            # Later intervals should be capped
            max_intervals = [interval for interval in error_backoffs if interval >= 2.0]
            assert len(max_intervals) > 0, "Should hit maximum backoff interval cap"

    @pytest.mark.asyncio
    async def test_backoff_with_jitter(self, mock_session):
        """Test backoff calculation with jitter applied."""
        controller = MediaMTXController(
            host="localhost",
            api_port=9997,
            rtsp_port=8554,
            webrtc_port=8889,
            hls_port=8888,
            config_path="/tmp/config.yml",
            recordings_path="/tmp/recordings",
            snapshots_path="/tmp/snapshots",
            health_check_interval=0.1,
            health_failure_threshold=5,
            health_max_backoff_interval=2.0,
            backoff_base_multiplier=2.0,
            backoff_jitter_range=(0.8, 1.2),  # ±20% jitter
            health_circuit_breaker_timeout=10.0,
        )
        controller._session = mock_session

        failure_response = self._mock_response(500, text_data="Error")
        responses = [failure_response] * 8
        mock_session.get.side_effect = responses

        sleep_intervals = []

        async def mock_sleep(interval):
            sleep_intervals.append(interval)
            await asyncio.sleep(0.01)

        with patch("asyncio.sleep", side_effect=mock_sleep):
            await controller.start()
            await asyncio.sleep(0.25)
            await controller.stop()

        # Verify jitter is applied (intervals should vary within expected ranges)
        error_backoffs = [interval for interval in sleep_intervals if interval > 0.1]

        if len(error_backoffs) >= 4:
            # With jitter, consecutive intervals should not be identical
            # (except when hitting the maximum cap)
            unique_intervals = set(error_backoffs[:4])  # First 4 backoffs
            assert (
                len(unique_intervals) >= 2
            ), "Jitter should create variation in backoff intervals"

    @pytest.mark.asyncio
    async def test_backoff_maximum_cap_enforcement(
        self, controller_backoff_test, mock_session
    ):
        """Test that backoff intervals are capped at maximum value."""
        controller = controller_backoff_test
        controller._session = mock_session
        controller._health_max_backoff_interval = 1.0  # Lower cap for testing

        failure_response = self._mock_response(500, text_data="Error")
        responses = [failure_response] * 15  # Many failures to exceed cap
        mock_session.get.side_effect = responses

        sleep_intervals = []

        async def mock_sleep(interval):
            sleep_intervals.append(interval)
            await asyncio.sleep(0.01)

        with patch("asyncio.sleep", side_effect=mock_sleep):
            await controller.start()
            await asyncio.sleep(0.4)
            await controller.stop()

        # Verify no interval exceeds the maximum cap
        max_interval = max(sleep_intervals)
        assert (
            max_interval <= 1.1
        ), f"No interval should exceed cap (max: {max_interval})"

        # Verify cap is actually reached
        capped_intervals = [interval for interval in sleep_intervals if interval >= 1.0]
        assert len(capped_intervals) > 0, "Should reach maximum backoff cap"

    @pytest.mark.asyncio
    async def test_backoff_reset_on_success(
        self, controller_backoff_test, mock_session
    ):
        """Test that backoff resets when health check succeeds."""
        controller = controller_backoff_test
        controller._session = mock_session

        failure_response = self._mock_response(500, text_data="Error")
        success_response = self._mock_response(200, {"serverVersion": "1.0.0"})

        # Pattern: failures (build up backoff) → success (reset) → failures again
        responses = [
            failure_response,
            failure_response,
            failure_response,  # Build backoff
            success_response,  # Reset backoff
            failure_response,
            failure_response,  # Fresh backoff sequence
        ]
        mock_session.get.side_effect = responses

        sleep_intervals = []

        async def mock_sleep(interval):
            sleep_intervals.append(interval)
            await asyncio.sleep(0.01)

        with patch("asyncio.sleep", side_effect=mock_sleep):
            await controller.start()
            await asyncio.sleep(0.4)
            await controller.stop()

        # Verify backoff was reset after success
        # Should see: increasing intervals → success (normal interval) → increasing again
        error_backoffs = [interval for interval in sleep_intervals if interval > 0.15]
        normal_intervals = [
            interval for interval in sleep_intervals if 0.08 <= interval <= 0.12
        ]

        assert len(error_backoffs) >= 2, "Should have error backoff intervals"
        assert len(normal_intervals) >= 1, "Should have normal intervals after success"

    @pytest.mark.asyncio
    async def test_configurable_backoff_multiplier(self, mock_session):
        """Test different backoff multiplier configurations."""
        # Controller with aggressive backoff (multiplier = 3.0)
        controller_aggressive = MediaMTXController(
            host="localhost",
            api_port=9997,
            rtsp_port=8554,
            webrtc_port=8889,
            hls_port=8888,
            config_path="/tmp/config.yml",
            recordings_path="/tmp/recordings",
            snapshots_path="/tmp/snapshots",
            health_check_interval=0.1,
            health_failure_threshold=5,
            health_max_backoff_interval=5.0,
            backoff_base_multiplier=3.0,
            backoff_jitter_range=(1.0, 1.0),
            health_circuit_breaker_timeout=10.0,
        )

        # Controller with conservative backoff (multiplier = 1.5)
        controller_conservative = MediaMTXController(
            host="localhost",
            api_port=9997,
            rtsp_port=8554,
            webrtc_port=8889,
            hls_port=8888,
            config_path="/tmp/config.yml",
            recordings_path="/tmp/recordings",
            snapshots_path="/tmp/snapshots",
            health_check_interval=0.1,
            health_failure_threshold=5,
            health_max_backoff_interval=5.0,
            backoff_base_multiplier=1.5,
            backoff_jitter_range=(1.0, 1.0),
            health_circuit_breaker_timeout=10.0,
        )

        # Verify multipliers are configured correctly
        assert controller_aggressive._backoff_base_multiplier == 3.0
        assert controller_conservative._backoff_base_multiplier == 1.5

    @pytest.mark.asyncio
    async def test_jitter_range_configuration(self, mock_session):
        """Test different jitter range configurations."""
        # No jitter
        controller_no_jitter = MediaMTXController(
            host="localhost",
            api_port=9997,
            rtsp_port=8554,
            webrtc_port=8889,
            hls_port=8888,
            config_path="/tmp/config.yml",
            recordings_path="/tmp/recordings",
            snapshots_path="/tmp/snapshots",
            backoff_jitter_range=(1.0, 1.0),  # No variation
        )

        # Wide jitter range
        controller_wide_jitter = MediaMTXController(
            host="localhost",
            api_port=9997,
            rtsp_port=8554,
            webrtc_port=8889,
            hls_port=8888,
            config_path="/tmp/config.yml",
            recordings_path="/tmp/recordings",
            snapshots_path="/tmp/snapshots",
            backoff_jitter_range=(0.5, 2.0),  # Wide variation
        )

        # Verify jitter configurations
        assert controller_no_jitter._backoff_jitter_range == (1.0, 1.0)
        assert controller_wide_jitter._backoff_jitter_range == (0.5, 2.0)

    @pytest.mark.asyncio
    async def test_circuit_breaker_backoff_interaction(
        self, controller_backoff_test, mock_session
    ):
        """Test backoff behavior when circuit breaker is active."""
        controller = controller_backoff_test
        controller._session = mock_session
        controller._health_failure_threshold = 2  # Low threshold to trigger CB quickly

        failure_response = self._mock_response(500, text_data="Error")
        responses = [failure_response] * 8  # Enough to trigger CB and continue
        mock_session.get.side_effect = responses

        sleep_intervals = []

        async def mock_sleep(interval):
            sleep_intervals.append(interval)
            await asyncio.sleep(0.01)

        with patch("asyncio.sleep", side_effect=mock_sleep):
            await controller.start()
            await asyncio.sleep(0.3)
            await controller.stop()

        # Verify circuit breaker was activated
        assert controller._health_state["circuit_breaker_activations"] > 0

        # Verify there are different types of sleep intervals:
        # - Normal health check intervals
        # - Error backoff intervals
        # - Circuit breaker wait intervals
        unique_intervals = set(sleep_intervals)
        assert (
            len(unique_intervals) >= 2
        ), "Should have varied sleep intervals during CB and backoff"

    @pytest.mark.asyncio
    async def test_deterministic_backoff_with_no_jitter(
        self, controller_backoff_test, mock_session
    ):
        """Test that backoff is deterministic when jitter is disabled."""
        controller = controller_backoff_test
        controller._session = mock_session
        controller._backoff_jitter_range = (1.0, 1.0)  # No jitter

        failure_response = self._mock_response(500, text_data="Error")
        responses = [failure_response] * 6
        mock_session.get.side_effect = responses

        sleep_intervals = []

        async def mock_sleep(interval):
            sleep_intervals.append(interval)
            await asyncio.sleep(0.01)

        with patch("asyncio.sleep", side_effect=mock_sleep):
            await controller.start()
            await asyncio.sleep(0.25)
            await controller.stop()

        # With no jitter, consecutive error backoffs at same failure count should be identical
        error_backoffs = [interval for interval in sleep_intervals if interval > 0.1]

        # Verify predictable exponential progression
        if len(error_backoffs) >= 3:
            # Expected: 0.2, 0.4, 0.8, ... (base_interval * multiplier^failures)
            expected_first = 0.1 * 2.0  # First error backoff
            expected_second = 0.1 * 4.0  # Second error backoff

            # Allow small floating point tolerance
            assert abs(error_backoffs[0] - expected_first) < 0.01
            assert abs(error_backoffs[1] - expected_second) < 0.01
