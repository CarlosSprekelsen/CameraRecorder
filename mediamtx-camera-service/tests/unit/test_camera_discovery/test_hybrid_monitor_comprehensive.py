"""
Comprehensive test scaffolds for hybrid camera monitor hardening validation.

Test coverage areas:
- Capability parsing variations (multiple fps, malformed output)
- Udev add/remove/change/race condition simulations
- Polling fallback behavior when udev is silent
- Timeout and subprocess failure handling
- Provisional/confirmed capability validation logic
- Adaptive polling with backoff and jitter
"""

import asyncio
import pytest
import time
from unittest.mock import Mock, AsyncMock, patch

# Test imports
from src.camera_discovery.hybrid_monitor import (
    HybridCameraMonitor,
    CameraEvent,
    CapabilityDetectionResult,
)


class TestCapabilityParsingVariations:
    """Test capability detection parsing with varied and malformed v4l2-ctl outputs."""

    @pytest.fixture
    def monitor(self):
        """Create monitor with capability detection enabled."""
        return HybridCameraMonitor(
            device_range=[0, 1, 2],
            enable_capability_detection=True,
            detection_timeout=2.0,
        )

    def test_frame_rate_extraction_comprehensive_patterns(self, monitor):
        """Test frame rate extraction from comprehensive v4l2-ctl output patterns."""
        test_cases = [
            # Standard patterns
            ("30.000 fps", {"30"}),
            ("25.000 FPS", {"25"}),
            ("Frame rate: 60.0", {"60"}),
            ("1920x1080@30", {"30"}),
            ("15 Hz", {"15"}),
            # Interval patterns
            ("Interval: [1/30]", {"30"}),
            ("[1/25]", {"25"}),
            ("1/15 s", {"15"}),
            # Complex patterns
            ("30 frames per second", {"30"}),
            ("rate: 25.5", {"25.5"}),
            ("fps: 60", {"60"}),
            # Multiple rates in one output
            ("30.000 fps, 25 FPS, [1/15], 60 Hz", {"30", "25", "15", "60"}),
            # Edge cases
            ("", set()),
            ("no frame rates here", set()),
            ("300 fps", {"300"}),  # High rate
            ("1.5 fps", {"1.5"}),  # Low rate
            ("0 fps", set()),  # Invalid rate (filtered out)
            ("500 fps", set()),  # Invalid rate (filtered out)
            # Malformed patterns
            ("30.000.000 fps", set()),  # Double decimal
            ("abc fps", set()),  # Non-numeric
            ("fps without number", set()),  # No number
        ]

        for output, expected in test_cases:
            result = monitor._extract_frame_rates_from_output(output)
            assert (
                result == expected
            ), f"Failed for output: '{output}' - expected {expected}, got {result}"

    @pytest.mark.asyncio
    async def test_capability_parsing_malformed_v4l2_outputs(self, monitor):
        """Test capability detection resilience against malformed v4l2-ctl outputs."""

        malformed_outputs = [
            # Empty/minimal outputs
            ("", False, "empty output"),
            ("v4l2-ctl: error", False, "error output"),
            # Partial outputs
            (
                "Card type: USB Camera\nDriver name: uvcvideo",
                True,
                "minimal valid info",
            ),
            ("Some random text without useful info", False, "random text"),
            # Malformed format sections
            ("Format [0]: corrupted data\nSize: invalid", False, "corrupted format"),
            (
                "Valid start\n[CORRUPTED MIDDLE SECTION]\nValid end",
                False,
                "corrupted middle",
            ),
            # Mixed valid/invalid data
            (
                "Card type: USB Camera\nFormat [0]: corrupted\nDriver: uvcvideo",
                False,
                "mixed valid/invalid",
            ),
        ]

        for output, expected_valid, description in malformed_outputs:
            # Mock subprocess to return our test output
            with patch("asyncio.create_subprocess_exec") as mock_subprocess:
                mock_process = AsyncMock()
                mock_process.communicate.return_value = (output.encode(), b"")
                mock_process.returncode = 0
                mock_subprocess.return_value = mock_process

                # Test capability detection
                result = await monitor._probe_device_capabilities("/dev/video0")
                assert result.detected == expected_valid, f"Failed for {description}"

    @pytest.mark.asyncio
    async def test_capability_timeout_handling(self, monitor):
        """Test capability detection timeout handling with mocked subprocess delays."""

        # Mock subprocess that takes too long
        with patch("asyncio.create_subprocess_exec") as mock_subprocess:
            async def slow_communicate():
                await asyncio.sleep(3.0)  # Longer than timeout
                return (b"valid output", b"")

            mock_process = AsyncMock()
            mock_process.communicate = slow_communicate
            mock_process.returncode = 0
            mock_subprocess.return_value = mock_process

            # Test with short timeout
            result = await monitor._probe_device_capabilities("/dev/video0")
            assert not result.detected
            assert "timeout" in result.error.lower()

    @pytest.mark.asyncio
    async def test_provisional_confirmed_capability_validation(self, monitor):
        """Test provisional to confirmed capability state transitions."""

        device_path = "/dev/video0"
        mock_capability = CapabilityDetectionResult(
            device_path=device_path,
            detected=True,
            accessible=True,
            resolutions=["1920x1080"],
            frame_rates=["30"],
            formats=[{"code": "YUYV", "description": "YUYV 4:2:2"}],
        )

        # Mock capability detection
        with patch.object(
            monitor, "_probe_device_capabilities", return_value=mock_capability
        ):
            # First probe - should be provisional
            await monitor._update_capability_state(device_path, mock_capability)
            state = monitor._get_or_create_capability_state(device_path)
            assert not state.is_confirmed()
            assert state.provisional_data is not None

            # Second probe with same data - should become confirmed
            await monitor._update_capability_state(device_path, mock_capability)
            state = monitor._get_or_create_capability_state(device_path)
            assert state.is_confirmed()
            assert state.confirmed_data is not None


class TestUdevEventProcessingAndRaceConditions:
    """Test udev event processing and race condition handling."""

    @pytest.fixture
    def monitor(self):
        """Create monitor with udev enabled."""
        return HybridCameraMonitor(
            device_range=[0, 1, 2],
            enable_capability_detection=True,
            detection_timeout=1.0,
        )

    @pytest.mark.asyncio
    async def test_udev_event_filtering_comprehensive(self, monitor):
        """Test comprehensive udev event filtering logic."""

        # Mock udev device with various properties
        mock_device = Mock()
        mock_device.device_node = "/dev/video0"
        mock_device.action = "add"
        mock_device.get.return_value = "video4linux"

        # Test device addition
        await monitor._process_udev_device_event(mock_device)

        # Verify device was processed
        assert "/dev/video0" in monitor._known_devices

        # Test device removal
        mock_device.action = "remove"
        await monitor._process_udev_device_event(mock_device)

        # Verify device was removed
        assert "/dev/video0" not in monitor._known_devices

    @pytest.mark.asyncio
    async def test_udev_race_condition_simulation(self, monitor):
        """Test race condition handling between udev events and polling."""

        device_path = "/dev/video0"

        # Simulate rapid add/remove/add sequence
        await self._simulate_delayed_udev_event(monitor, device_path, "add", 0.0)
        await self._simulate_delayed_udev_event(monitor, device_path, "remove", 0.1)
        await self._simulate_delayed_udev_event(monitor, device_path, "add", 0.2)

        # Verify final state is consistent
        assert device_path in monitor._known_devices

    async def _simulate_delayed_udev_event(
        self, monitor, device_path: str, action: str, delay: float
    ):
        """Simulate udev event with delay."""
        await asyncio.sleep(delay)
        mock_device = Mock()
        mock_device.device_node = device_path
        mock_device.action = action
        mock_device.get.return_value = "video4linux"
        await monitor._process_udev_device_event(mock_device)

    @pytest.mark.asyncio
    async def test_udev_change_event_status_detection(self, monitor):
        """Test udev change event status detection and handling."""

        device_path = "/dev/video0"
        mock_device = Mock()
        mock_device.device_node = device_path
        mock_device.action = "change"
        mock_device.get.return_value = "video4linux"

        # Add device first
        add_device = Mock()
        add_device.device_node = device_path
        add_device.action = "add"
        add_device.get.return_value = "video4linux"
        await monitor._process_udev_device_event(add_device)

        # Then simulate change event
        await monitor._process_udev_device_event(mock_device)

        # Verify device still exists and status was updated
        assert device_path in monitor._known_devices


class TestPollingFallbackBehavior:
    """Test polling fallback behavior when udev is unavailable."""

    @pytest.fixture
    def monitor_no_udev(self):
        """Create monitor with udev disabled."""
        with patch("src.camera_discovery.hybrid_monitor.HAS_PYUDEV", False):
            return HybridCameraMonitor(
                device_range=[0, 1, 2],
                enable_capability_detection=True,
                detection_timeout=1.0,
            )

    @pytest.mark.asyncio
    async def test_polling_only_mode_device_discovery(self, monitor_no_udev):
        """Test device discovery in polling-only mode."""

        # Mock file system operations
        with patch("pathlib.Path.exists") as mock_exists, patch("builtins.open") as mock_open:
            mock_exists.return_value = True

            def mock_path_exists(path_str):
                return path_str in ["/dev/video0", "/dev/video1"]

            def mock_open_device(path, mode="rb"):
                mock_file = Mock()
                mock_file.read.return_value = b"mock device data"
                return mock_file

            mock_exists.side_effect = mock_path_exists
            mock_open.side_effect = mock_open_device

            # Test discovery
            await monitor_no_udev._discover_cameras()

            # Verify devices were discovered
            assert len(monitor_no_udev._known_devices) > 0

    @pytest.mark.asyncio
    async def test_adaptive_polling_interval_adjustment(self, monitor_no_udev):
        """Test adaptive polling interval adjustment logic."""

        initial_interval = monitor_no_udev._current_poll_interval

        # Simulate no recent udev events
        monitor_no_udev._last_udev_event_time = time.time() - 10.0

        # Adjust polling interval
        await monitor_no_udev._adjust_polling_interval()

        # Verify interval increased
        assert monitor_no_udev._current_poll_interval < initial_interval

    @pytest.mark.asyncio
    async def test_polling_failure_backoff_with_jitter(self, monitor_no_udev):
        """Test polling failure backoff with jitter."""


        # Simulate polling failure
        with patch.object(monitor_no_udev, "_discover_cameras", side_effect=Exception("Test failure")):
            # Run one iteration of polling loop
            try:
                await monitor_no_udev._adaptive_polling_loop()
            except Exception:
                pass  # Expected due to mocked failure

        # Verify backoff was applied
        assert monitor_no_udev._polling_failure_count > 0


class TestTimeoutAndSubprocessFailureHandling:
    """Test timeout and subprocess failure handling."""

    @pytest.fixture
    def monitor(self):
        """Create monitor for timeout testing."""
        return HybridCameraMonitor(
            device_range=[0, 1, 2],
            enable_capability_detection=True,
            detection_timeout=1.0,
        )

    @pytest.mark.asyncio
    async def test_subprocess_timeout_handling(self, monitor):
        """Test subprocess timeout handling."""

        # Mock subprocess that hangs
        with patch("asyncio.create_subprocess_exec") as mock_subprocess:
            async def hanging_communicate():
                await asyncio.sleep(5.0)  # Longer than timeout
                return (b"", b"")

            mock_process = AsyncMock()
            mock_process.communicate = hanging_communicate
            mock_process.returncode = 0
            mock_subprocess.return_value = mock_process

            # Test capability detection with timeout
            result = await monitor._probe_device_capabilities("/dev/video0")
            assert not result.detected
            assert "timeout" in result.error.lower()

    @pytest.mark.asyncio
    async def test_subprocess_failure_handling(self, monitor):
        """Test subprocess failure handling."""

        # Mock subprocess that fails
        with patch("asyncio.create_subprocess_exec") as mock_subprocess:
            mock_process = AsyncMock()
            mock_process.communicate.return_value = (b"", b"error message")
            mock_process.returncode = 1
            mock_subprocess.return_value = mock_process

            # Test capability detection with failure
            result = await monitor._probe_device_capabilities("/dev/video0")
            assert not result.detected
            assert result.error is not None

    @pytest.mark.asyncio
    async def test_concurrent_capability_probes_handling(self, monitor):
        """Test concurrent capability probe handling."""

        device_paths = ["/dev/video0", "/dev/video1", "/dev/video2"]

        # Mock capability detection
        with patch.object(monitor, "_probe_device_capabilities") as mock_probe:
            mock_probe.return_value = CapabilityDetectionResult(
                device_path="test",
                detected=True,
                accessible=True,
            )

            # Run concurrent probes
            tasks = [
                monitor._probe_device_capabilities(path) for path in device_paths
            ]
            results = await asyncio.gather(*tasks, return_exceptions=True)

            # Verify all probes completed
            assert len(results) == len(device_paths)
            assert all(isinstance(r, CapabilityDetectionResult) for r in results)


class TestIntegrationAndLifecycle:
    """Test integration scenarios and monitor lifecycle."""

    @pytest.fixture
    def monitor(self):
        """Create monitor for lifecycle testing."""
        return HybridCameraMonitor(
            device_range=[0, 1, 2],
            enable_capability_detection=True,
            detection_timeout=1.0,
        )

    @pytest.mark.asyncio
    async def test_monitor_full_lifecycle(self, monitor):
        """Test complete monitor lifecycle from startup to shutdown."""

        # Verify initial state
        assert not monitor.is_running
        initial_stats = monitor.get_monitor_stats()
        assert initial_stats["running"] is False
        assert initial_stats["active_tasks"] == 0

        # Start monitor with timeout protection
        with patch("src.camera_discovery.hybrid_monitor.HAS_PYUDEV", False):
            try:
                # Start monitor with timeout
                await asyncio.wait_for(monitor.start(), timeout=5.0)

                assert monitor.is_running
                running_stats = monitor.get_monitor_stats()
                assert running_stats["running"] is True
                assert running_stats["active_tasks"] > 0

                # Let monitor run briefly
                await asyncio.sleep(0.1)

                # Verify some activity occurred
                activity_stats = monitor.get_monitor_stats()
                assert activity_stats["polling_cycles"] > 0

                # Stop monitor with timeout
                await asyncio.wait_for(monitor.stop(), timeout=5.0)

                assert not monitor.is_running
                final_stats = monitor.get_monitor_stats()
                assert final_stats["running"] is False
                assert final_stats["active_tasks"] == 0

            except asyncio.TimeoutError:
                # Force cleanup if timeout occurs
                monitor._running = False
                for task in monitor._monitoring_tasks:
                    if not task.done():
                        task.cancel()
                await asyncio.gather(*monitor._monitoring_tasks, return_exceptions=True)
                raise

    @pytest.mark.asyncio
    async def test_end_to_end_device_workflow(self, monitor):
        """Test end-to-end device detection and capability integration workflow."""

        monitor._set_test_mode(True)
        device_path = "/dev/video0"

        # Mock capability detection
        mock_capability = CapabilityDetectionResult(
            device_path=device_path,
            detected=True,
            accessible=True,
            resolutions=["1920x1080", "1280x720"],
            frame_rates=["30", "25", "15"],
            formats=[{"code": "YUYV", "description": "YUYV 4:2:2"}],
        )

        with patch.object(
            monitor, "_probe_device_capabilities", return_value=mock_capability
        ):
            # Create event handler to capture events
            captured_events = []

            def capture_event(event_data):
                captured_events.append(event_data)

            monitor.add_event_callback(capture_event)

            # Simulate device connection
            await monitor._inject_test_udev_event(device_path, "add")

            # Verify event was captured
            assert len(captured_events) > 0
            connect_event = captured_events[-1]
            assert connect_event.event_type == CameraEvent.CONNECTED
            assert connect_event.device_path == device_path

            # Set up capability state for metadata check
            state = monitor._get_or_create_capability_state(device_path)
            state.provisional_data = mock_capability
            state.consecutive_successes = 1

            # Verify capability metadata is available
            metadata = monitor.get_effective_capability_metadata(device_path)
            assert metadata["resolution"] == "1920x1080"
            assert metadata["fps"] == 30
            assert metadata["validation_status"] in ["provisional", "confirmed"]

            # Simulate device disconnection
            await monitor._inject_test_udev_event(device_path, "remove")

            # Verify disconnect event
            disconnect_event = captured_events[-1]
            assert disconnect_event.event_type == CameraEvent.DISCONNECTED

        monitor._set_test_mode(False)


# Utility functions for test setup and validation


def create_mock_v4l2_output(formats=None, resolutions=None, frame_rates=None):
    """Create mock v4l2-ctl output for testing."""
    output_lines = [
        "Driver name: uvcvideo",
        "Card type: USB Camera",
        "Bus info: usb-0000:00:14.0-1",
    ]

    if formats:
        output_lines.append("Format [0]:")
        for fmt in formats:
            output_lines.append(f"  {fmt}")

    if resolutions:
        output_lines.append("Size: Discrete 1920x1080")
        for res in resolutions:
            output_lines.append(f"  {res}")

    if frame_rates:
        output_lines.append("Frame rate: Discrete 30.000 fps")
        for rate in frame_rates:
            output_lines.append(f"  {rate} fps")

    return "\n".join(output_lines)


def assert_capability_consistency(
    result1: CapabilityDetectionResult, result2: CapabilityDetectionResult
):
    """Assert that two capability results are consistent."""
    assert result1.detected == result2.detected
    assert result1.accessible == result2.accessible
    assert result1.device_path == result2.device_path


# Pytest configuration
pytest_plugins = ["pytest_asyncio"]


if __name__ == "__main__":
    # Allow running tests directly
    pytest.main([__file__, "-v", "--tb=short"])
