"""
Udev event processing and polling fallback tests for hybrid camera monitor.

Test coverage:
- Udev add/remove/change events with race conditions
- Invalid device node handling
- Polling fallback when udev events are missed or stale
- Event filtering and device range validation
- Adaptive polling interval adjustment

Created: 2025-08-04
Related: S3 Camera Discovery hardening, docs/roadmap.md
Evidence: src/camera_discovery/hybrid_monitor.py lines 200-400 (udev event processing)
"""

import asyncio
import pytest
import time
from unittest.mock import Mock, AsyncMock, patch, MagicMock, call
from pathlib import Path
from typing import Dict, List, Any

# Test imports
from src.camera_discovery.hybrid_monitor import (
    HybridCameraMonitor,
    CameraEvent,
    CameraEventData,
    CapabilityDetectionResult,
    DeviceCapabilityState,
)
from src.common.types import CameraDevice


class TestUdevEventProcessing:
    """Test udev event handling including edge cases and race conditions."""

    @pytest.fixture
    def monitor_with_udev(self):
        """Create monitor with udev enabled for event testing."""
        with patch("src.camera_discovery.hybrid_monitor.HAS_PYUDEV", True):
            return HybridCameraMonitor(
                device_range=[0, 1, 2],
                poll_interval=0.1,
                enable_capability_detection=False,  # Focus on event processing
            )

    @pytest.fixture
    def mock_udev_device(self):
        """Create mock udev device for testing."""
        mock_device = Mock()
        mock_device.device_path = "/dev/video0"
        mock_device.device_node = "/dev/video0"
        mock_device.get.return_value = "camera_device"  # ID_V4L_PRODUCT
        mock_device.action = "add"
        mock_device.subsystem = "video4linux"
        return mock_device

    @pytest.mark.asyncio
    async def test_udev_add_event_processing(self, monitor_with_udev, mock_udev_device):
        """Test udev 'add' event processing and device registration."""

        # Mock device availability checks
        with (
            patch("pathlib.Path.exists", return_value=True),
            patch("builtins.open", return_value=Mock()),
            patch.object(
                monitor_with_udev, "_should_monitor_device", return_value=True
            ),
        ):

            # Setup event handler to capture events
            captured_events = []

            async def capture_event(event_data: CameraEventData):
                captured_events.append(event_data)

            monitor_with_udev.add_event_callback(capture_event)

            # Simulate udev add event
            mock_udev_device.action = "add"
            await monitor_with_udev._handle_udev_event(mock_udev_device)

            # Verify event processing
            assert len(captured_events) == 1
            event = captured_events[0]
            assert event.event_type == CameraEvent.CONNECTED
            assert event.device_path == "/dev/video0"

            # Verify device tracking
            assert "/dev/video0" in monitor_with_udev._known_devices

            # Verify stats update
            stats = monitor_with_udev.get_monitor_stats()
            assert stats["udev_events_processed"] == 1

    @pytest.mark.asyncio
    async def test_udev_remove_event_processing(
        self, monitor_with_udev, mock_udev_device
    ):
        """Test udev 'remove' event processing and device cleanup."""

        # Pre-populate device
        test_device = CameraDevice(
            device_path="/dev/video0", name="Test Camera", driver="uvcvideo"
        )
        monitor_with_udev._known_devices["/dev/video0"] = test_device

        captured_events = []

        async def capture_event(event_data: CameraEventData):
            captured_events.append(event_data)

        monitor_with_udev.add_event_callback(capture_event)

        # Simulate udev remove event
        mock_udev_device.action = "remove"
        await monitor_with_udev._handle_udev_event(mock_udev_device)

        # Verify event processing
        assert len(captured_events) == 1
        event = captured_events[0]
        assert event.event_type == CameraEvent.DISCONNECTED
        assert event.device_path == "/dev/video0"

        # Verify device removal
        assert "/dev/video0" not in monitor_with_udev._known_devices

        # Verify capability state cleanup if it exists
        assert "/dev/video0" not in monitor_with_udev._capability_states

    @pytest.mark.asyncio
    async def test_udev_change_event_processing(
        self, monitor_with_udev, mock_udev_device
    ):
        """Test udev 'change' event processing for device state updates."""

        # Pre-populate device
        test_device = CameraDevice(
            device_path="/dev/video0", name="Test Camera", driver="uvcvideo"
        )
        monitor_with_udev._known_devices["/dev/video0"] = test_device

        captured_events = []

        async def capture_event(event_data: CameraEventData):
            captured_events.append(event_data)

        monitor_with_udev.add_event_callback(capture_event)

        # Simulate udev change event
        mock_udev_device.action = "change"
        with (
            patch("pathlib.Path.exists", return_value=True),
            patch("builtins.open", return_value=Mock()),
        ):

            await monitor_with_udev._handle_udev_event(mock_udev_device)

        # Verify event processing
        assert len(captured_events) == 1
        event = captured_events[0]
        assert event.event_type == CameraEvent.STATUS_CHANGED
        assert event.device_path == "/dev/video0"

    @pytest.mark.asyncio
    async def test_udev_event_race_conditions(self, monitor_with_udev):
        """Test rapid sequential udev events to detect race conditions."""

        captured_events = []

        async def capture_event(event_data: CameraEventData):
            captured_events.append(event_data)

        monitor_with_udev.add_event_callback(capture_event)

        # Create multiple mock devices
        devices = []
        for i in range(3):
            mock_device = Mock()
            mock_device.device_path = f"/dev/video{i}"
            mock_device.device_node = f"/dev/video{i}"
            mock_device.get.return_value = f"camera_{i}"
            mock_device.action = "add"
            mock_device.subsystem = "video4linux"
            devices.append(mock_device)

        with (
            patch("pathlib.Path.exists", return_value=True),
            patch("builtins.open", return_value=Mock()),
            patch.object(
                monitor_with_udev, "_should_monitor_device", return_value=True
            ),
        ):

            # Fire rapid sequential events
            tasks = []
            for device in devices:
                task = asyncio.create_task(monitor_with_udev._handle_udev_event(device))
                tasks.append(task)

            # Wait for all events to process
            await asyncio.gather(*tasks)

        # Verify all events processed correctly
        assert len(captured_events) == 3
        assert len(monitor_with_udev._known_devices) == 3

        # Verify no race condition artifacts
        device_paths = [event.device_path for event in captured_events]
        assert "/dev/video0" in device_paths
        assert "/dev/video1" in device_paths
        assert "/dev/video2" in device_paths

    @pytest.mark.asyncio
    async def test_invalid_device_node_handling(self, monitor_with_udev):
        """Test handling of udev events with invalid or inaccessible device nodes."""

        captured_events = []

        async def capture_event(event_data: CameraEventData):
            captured_events.append(event_data)

        monitor_with_udev.add_event_callback(capture_event)

        # Create mock device with invalid path
        mock_device = Mock()
        mock_device.device_path = "/dev/video999"  # Out of range
        mock_device.device_node = "/dev/video999"
        mock_device.get.return_value = "invalid_camera"
        mock_device.action = "add"
        mock_device.subsystem = "video4linux"

        # Simulate device path doesn't exist
        with patch("pathlib.Path.exists", return_value=False):
            await monitor_with_udev._handle_udev_event(mock_device)

        # Should not generate events for invalid devices
        assert len(captured_events) == 0
        assert len(monitor_with_udev._known_devices) == 0

        # Verify stats show filtered event
        stats = monitor_with_udev.get_monitor_stats()
        assert stats["udev_events_filtered"] > 0

    @pytest.mark.asyncio
    async def test_device_range_filtering(self, monitor_with_udev):
        """Test udev event filtering based on configured device range."""

        captured_events = []

        async def capture_event(event_data: CameraEventData):
            captured_events.append(event_data)

        monitor_with_udev.add_event_callback(capture_event)

        # Test devices both in and out of range
        test_cases = [
            ("/dev/video0", True),  # In range
            ("/dev/video1", True),  # In range
            ("/dev/video2", True),  # In range
            ("/dev/video5", False),  # Out of range
            ("/dev/video10", False),  # Out of range
        ]

        for device_path, should_process in test_cases:
            mock_device = Mock()
            mock_device.device_path = device_path
            mock_device.device_node = device_path
            mock_device.get.return_value = "test_camera"
            mock_device.action = "add"
            mock_device.subsystem = "video4linux"

            with (
                patch("pathlib.Path.exists", return_value=True),
                patch("builtins.open", return_value=Mock()),
            ):

                await monitor_with_udev._handle_udev_event(mock_device)

        # Only devices in range [0,1,2] should generate events
        processed_devices = [event.device_path for event in captured_events]
        assert "/dev/video0" in processed_devices
        assert "/dev/video1" in processed_devices
        assert "/dev/video2" in processed_devices
        assert "/dev/video5" not in processed_devices
        assert "/dev/video10" not in processed_devices

        assert len(captured_events) == 3  # Only 3 in-range devices


class TestPollingFallback:
    """Test polling fallback behavior when udev events are missed or stale."""

    @pytest.fixture
    def monitor_polling_fallback(self):
        """Create monitor configured for polling fallback testing."""
        return HybridCameraMonitor(
            device_range=[0, 1],
            poll_interval=0.05,  # Fast polling for testing
            enable_capability_detection=False,
        )

    @pytest.mark.asyncio
    async def test_polling_fallback_when_udev_stale(self, monitor_polling_fallback):
        """Test polling fallback activation when udev events become stale."""

        # Mock initial state - udev events are fresh
        monitor_polling_fallback._last_udev_event_time = time.time()
        monitor_polling_fallback._udev_event_freshness_threshold = (
            1.0  # 1 second threshold
        )

        captured_events = []

        async def capture_event(event_data: CameraEventData):
            captured_events.append(event_data)

        monitor_polling_fallback.add_event_callback(capture_event)

        # Mock device discovery to find new device
        with patch.object(
            monitor_polling_fallback, "_discover_cameras"
        ) as mock_discover:
            mock_discover.return_value = None  # Discovery method doesn't return

            # Fast-forward time to make udev events stale
            with patch("time.time", return_value=time.time() + 2.0):
                # Run polling cycle
                await monitor_polling_fallback._polling_monitor()

            # Verify polling was triggered due to stale udev events
            mock_discover.assert_called_once()

    @pytest.mark.asyncio
    async def test_adaptive_polling_interval_adjustment(self, monitor_polling_fallback):
        """Test adaptive polling interval adjustment based on udev event freshness."""

        initial_interval = monitor_polling_fallback._current_poll_interval

        # Simulate stale udev events (should increase polling frequency)
        monitor_polling_fallback._last_udev_event_time = (
            time.time() - 30.0
        )  # Very stale
        monitor_polling_fallback._udev_event_freshness_threshold = 15.0

        # Mock polling cycle execution
        with patch.object(monitor_polling_fallback, "_discover_cameras"):
            await monitor_polling_fallback._polling_monitor()

        # Polling interval should have decreased (higher frequency)
        assert monitor_polling_fallback._current_poll_interval < initial_interval

        # Stats should reflect adjustment
        stats = monitor_polling_fallback.get_monitor_stats()
        assert stats["adaptive_poll_adjustments"] > 0
        assert (
            stats["current_poll_interval"]
            == monitor_polling_fallback._current_poll_interval
        )

    @pytest.mark.asyncio
    async def test_polling_failure_recovery(self, monitor_polling_fallback):
        """Test polling failure handling and recovery behavior."""

        # Mock discovery failures
        failure_count = 0

        async def mock_discover_with_failures():
            nonlocal failure_count
            failure_count += 1
            if failure_count <= 3:  # Fail first 3 attempts
                raise OSError("Mock discovery failure")
            # Succeed on 4th attempt
            return None

        with patch.object(
            monitor_polling_fallback,
            "_discover_cameras",
            side_effect=mock_discover_with_failures,
        ):

            # Run multiple polling cycles
            for _ in range(5):
                try:
                    await monitor_polling_fallback._polling_monitor()
                except Exception:
                    pass  # Expected for first few attempts

        # Verify failure tracking
        assert (
            monitor_polling_fallback._polling_failure_count
            <= monitor_polling_fallback._max_consecutive_failures
        )

        # Stats should reflect failures and recovery
        stats = monitor_polling_fallback.get_monitor_stats()
        assert stats["polling_cycles"] >= 5

    @pytest.mark.asyncio
    async def test_polling_discovers_missed_device(self, monitor_polling_fallback):
        """Test that polling detects devices missed by udev events."""

        captured_events = []

        async def capture_event(event_data: CameraEventData):
            captured_events.append(event_data)

        monitor_polling_fallback.add_event_callback(capture_event)

        # Mock device that exists but wasn't detected by udev
        test_devices = {"/dev/video0": ("CONNECTED", "Missed Camera")}

        def mock_path_exists(path_str):
            return str(path_str) in test_devices

        def mock_open_device(path, mode="rb"):
            if str(path) in test_devices and test_devices[str(path)][0] == "CONNECTED":
                return Mock()
            raise OSError("Device not accessible")

        with (
            patch("pathlib.Path.exists", side_effect=mock_path_exists),
            patch("builtins.open", side_effect=mock_open_device),
        ):

            # Run discovery cycle
            await monitor_polling_fallback._discover_cameras()

        # Should have discovered the missed device
        assert len(captured_events) > 0
        assert any(event.device_path == "/dev/video0" for event in captured_events)
        assert "/dev/video0" in monitor_polling_fallback._known_devices

    @pytest.mark.asyncio
    async def test_polling_only_mode_fallback(self, monitor_polling_fallback):
        """Test operation when udev is completely unavailable (polling-only mode)."""

        # Disable udev completely
        with patch("src.camera_discovery.hybrid_monitor.HAS_PYUDEV", False):
            monitor_no_udev = HybridCameraMonitor(
                device_range=[0, 1],
                poll_interval=0.05,
                enable_capability_detection=False,
            )

            captured_events = []

            async def capture_event(event_data: CameraEventData):
                captured_events.append(event_data)

            monitor_no_udev.add_event_callback(capture_event)

            # Mock device existence for polling detection
            with (
                patch("pathlib.Path.exists", return_value=True),
                patch("builtins.open", return_value=Mock()),
            ):

                await monitor_no_udev._discover_cameras()

            # Should still detect devices through polling
            assert len(captured_events) > 0

            # Verify polling-only stats
            stats = monitor_no_udev.get_monitor_stats()
            assert stats["polling_cycles"] > 0
            assert stats["udev_events_processed"] == 0  # No udev in polling-only mode
