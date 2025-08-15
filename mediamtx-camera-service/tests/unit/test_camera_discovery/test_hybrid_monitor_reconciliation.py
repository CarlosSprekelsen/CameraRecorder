"""
Reconciliation tests between hybrid_monitor capability output and service_manager
consumption.

Requirements Traceability:
- REQ-CAM-002: Camera discovery shall reconcile capability data between components
- REQ-CAM-004: Camera discovery shall maintain metadata consistency across service boundaries
- REQ-CAM-002: Camera discovery shall propagate confirmed vs provisional capability states

Story Coverage: S3 - Camera Discovery Hardening
IV&V Control Point: Real reconciliation validation

Test coverage: - End-to-end capability flow validation - Provisional vs
confirmed state propagation - Metadata consistency and drift detection - Integration
validation between components  Created: 2025-08-04 Related: S3 Camera Discovery
hardening, docs/roadmap.md Evidence: src/camera_discovery/hybrid_monitor.py lines
750-800 (get_effective_capability_metadata) Evidence:
src/camera_service/service_manager.py lines 270-380 (_get_enhanced_camera_metadata)
"""

import pytest
from unittest.mock import Mock

# Test imports
from src.camera_discovery.hybrid_monitor import (
    HybridCameraMonitor,
    CapabilityDetectionResult,
    CameraEventData,
    CameraEvent,
)
from src.camera_service.service_manager import ServiceManager
from src.common.types import CameraDevice


class TestCapabilityReconciliation:
    """Test reconciliation between hybrid_monitor output and service_manager consumption."""

    @pytest.fixture
    def hybrid_monitor(self):
        """Create hybrid monitor for reconciliation testing."""
        return HybridCameraMonitor(
            device_range=[0, 1, 2], enable_capability_detection=True
        )

    @pytest.fixture
    def service_manager(self, mock_dependencies):
        """Create service manager with mocked dependencies."""
        service_manager = ServiceManager(
            config=mock_dependencies["config"],
            mediamtx_controller=mock_dependencies["mediamtx_controller"],
            websocket_server=mock_dependencies["websocket_server"],
        )
        # Will inject camera_monitor per test
        return service_manager

    @pytest.mark.asyncio
    async def test_confirmed_capability_reconciliation(
        self, hybrid_monitor, service_manager, mock_dependencies
    ):
        """Test confirmed capability data flows correctly to service manager."""

        device_path = "/dev/video0"

        # Setup confirmed capability in hybrid_monitor
        confirmed_capability = CapabilityDetectionResult(
            device=device_path,
            detected=True,
            accessible=True,
            device_name="Confirmed Test Camera",
            formats=[{"format": "YUYV", "description": "YUYV 4:2:2"}],
            resolutions=["1920x1080", "1280x720", "640x480"],
            frame_rates=["30", "25", "15"],
        )

        # Create confirmed state in hybrid_monitor
        state = hybrid_monitor._get_or_create_capability_state(device_path)
        state.confirmed_data = confirmed_capability
        state.consecutive_successes = 5  # Above confirmation threshold

        # Inject hybrid_monitor into service_manager
        service_manager._camera_monitor = hybrid_monitor

        # Create camera event data
        camera_device = CameraDevice(
            device=device_path, name="Test Camera Device", driver="uvcvideo"
        )
        event_data = CameraEventData(
            device=device_path,
            event_type=CameraEvent.CONNECTED,
            device_info=camera_device,
        )

        # Get metadata from service_manager (this calls hybrid_monitor internally)
        metadata = await service_manager._get_enhanced_camera_metadata(event_data)

        # Verify reconciliation - confirmed data should propagate
        assert metadata["validation_status"] == "confirmed"
        assert metadata["capability_source"] == "confirmed_capability"
        assert metadata["consecutive_successes"] == 5


# ===== UDEV EVENT PROCESSING TESTS =====

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
            patch.object(
                monitor_with_udev, "_create_camera_device_info", 
                return_value=CameraDevice(
                    device="/dev/video0", 
                    name="Test Camera", 
                    status="CONNECTED"
                )
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
            device="/dev/video0", name="Test Camera", driver="uvcvideo"
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
            device="/dev/video0", name="Test Camera", driver="uvcvideo"
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

        # Verify capability data consistency
        hybrid_metadata = hybrid_monitor.get_effective_capability_metadata(device_path)
        assert hybrid_metadata["validation_status"] == metadata["validation_status"]
        assert (
            hybrid_metadata["consecutive_successes"]
            == metadata["consecutive_successes"]
        )

        # Verify resolution selection consistency (should pick highest available)
        assert metadata["resolution"] == hybrid_metadata["resolution"]
        assert metadata["fps"] == hybrid_metadata["fps"]

        print("✅ Confirmed capability reconciliation verified:")
        print(f"   - Validation status: {metadata['validation_status']}")
        print(f"   - Resolution: {metadata['resolution']}")
        print(f"   - FPS: {metadata['fps']}")
        print(f"   - Consecutive successes: {metadata['consecutive_successes']}")

    @pytest.mark.asyncio
    async def test_provisional_capability_reconciliation(
        self, hybrid_monitor, service_manager, mock_dependencies
    ):
        """Test provisional capability data flows correctly to service manager."""

        device_path = "/dev/video1"

        # Setup provisional capability in hybrid_monitor
        provisional_capability = CapabilityDetectionResult(
            device=device_path,
            detected=True,
            accessible=True,
            device_name="Provisional Test Camera",
            formats=[{"format": "MJPG", "description": "Motion-JPEG"}],
            resolutions=["1280x720", "640x480"],
            frame_rates=["30", "15"],
        )

        # Create provisional state in hybrid_monitor
        state = hybrid_monitor._get_or_create_capability_state(device_path)
        state.provisional_data = provisional_capability
        state.consecutive_successes = 1  # Below confirmation threshold

        # Inject hybrid_monitor into service_manager
        service_manager._camera_monitor = hybrid_monitor

        # Create camera event data
        camera_device = CameraDevice(
            device=device_path, name="Provisional Camera Device", driver="uvcvideo"
        )
        event_data = CameraEventData(
            device=device_path,
            event_type=CameraEvent.CONNECTED,
            device_info=camera_device,
        )

        # Get metadata from service_manager
        metadata = await service_manager._get_enhanced_camera_metadata(event_data)

        # Verify reconciliation - provisional data should propagate
        assert metadata["validation_status"] == "provisional"
        assert metadata["capability_source"] == "provisional_capability"
        assert metadata["consecutive_successes"] == 1

        # Verify capability data consistency
        hybrid_metadata = hybrid_monitor.get_effective_capability_metadata(device_path)
        assert hybrid_metadata["validation_status"] == metadata["validation_status"]
        assert (
            hybrid_metadata["consecutive_successes"]
            == metadata["consecutive_successes"]
        )

        # Verify data consistency
        assert metadata["resolution"] == hybrid_metadata["resolution"]
        assert metadata["fps"] == hybrid_metadata["fps"]

        print("✅ Provisional capability reconciliation verified:")
        print(f"   - Validation status: {metadata['validation_status']}")
        print(f"   - Resolution: {metadata['resolution']}")
        print(f"   - FPS: {metadata['fps']}")
        print(f"   - Consecutive successes: {metadata['consecutive_successes']}")

    @pytest.mark.asyncio
    async def test_no_capability_data_reconciliation(
        self, hybrid_monitor, service_manager, mock_dependencies
    ):
        """Test fallback behavior when no capability data is available."""

        device_path = "/dev/video2"

        # Don't setup any capability data - simulate device with no capability info

        # Inject hybrid_monitor into service_manager
        service_manager._camera_monitor = hybrid_monitor

        # Create camera event data
        camera_device = CameraDevice(
            device=device_path, name="Unknown Capability Camera", driver="unknown"
        )
        event_data = CameraEventData(
            device=device_path,
            event_type=CameraEvent.CONNECTED,
            device_info=camera_device,
        )

        # Get metadata from service_manager
        metadata = await service_manager._get_enhanced_camera_metadata(event_data)

        # Verify fallback behavior
        assert metadata["validation_status"] == "none"
        assert (
            metadata["capability_source"] == "device_info"
        )  # Falls back to device info
        assert metadata["consecutive_successes"] == 0

        # Should use architecture defaults
        assert metadata["resolution"] == "1920x1080"  # Architecture default
        assert metadata["fps"] == 30  # Architecture default

        # Verify hybrid_monitor also returns None for no capability data
        hybrid_metadata = hybrid_monitor.get_effective_capability_metadata(device_path)
        assert hybrid_metadata is None

        print("✅ No capability data reconciliation verified:")
        print(f"   - Validation status: {metadata['validation_status']}")
        print(f"   - Capability source: {metadata['capability_source']}")
        print(f"   - Fallback resolution: {metadata['resolution']}")
        print(f"   - Fallback FPS: {metadata['fps']}")

    @pytest.mark.asyncio
    async def test_capability_state_transition_reconciliation(
        self, hybrid_monitor, service_manager, mock_dependencies
    ):
        """Test reconciliation during capability state transitions (provisional → confirmed)."""

        device_path = "/dev/video0"

        # Setup initial provisional capability
        provisional_capability = CapabilityDetectionResult(
            device=device_path,
            detected=True,
            accessible=True,
            device_name="Transitioning Camera",
            resolutions=["1920x1080"],
            frame_rates=["30"],
        )

        state = hybrid_monitor._get_or_create_capability_state(device_path)
        state.provisional_data = provisional_capability
        state.consecutive_successes = 1

        # Inject hybrid_monitor into service_manager
        service_manager._camera_monitor = hybrid_monitor

        camera_device = CameraDevice(
            device=device_path, name="Test Camera", driver="uvcvideo"
        )
        event_data = CameraEventData(
            device=device_path,
            event_type=CameraEvent.CONNECTED,
            device_info=camera_device,
        )

        # Get initial metadata (should be provisional)
        initial_metadata = await service_manager._get_enhanced_camera_metadata(event_data)
        assert initial_metadata["validation_status"] == "provisional"

        # Simulate additional consistent detections to trigger confirmation
        for _ in range(3):  # Reach confirmation threshold
            await hybrid_monitor._update_capability_state(device_path, provisional_capability)

        # Get metadata after confirmation
        confirmed_metadata = await service_manager._get_enhanced_camera_metadata(event_data)

        # Verify transition to confirmed state
        assert confirmed_metadata["validation_status"] == "confirmed"
        assert confirmed_metadata["capability_source"] == "confirmed_capability"
        assert confirmed_metadata["consecutive_successes"] >= 3

        # Verify data consistency maintained during transition
        assert confirmed_metadata["resolution"] == initial_metadata["resolution"]
        assert confirmed_metadata["fps"] == initial_metadata["fps"]

        print("✅ State transition reconciliation verified:")
        print(
            f"   - Initial: {initial_metadata['validation_status']} → Final: {confirmed_metadata['validation_status']}"
        )
        print(
            f"   - Data consistency maintained: {confirmed_metadata['resolution']}@{confirmed_metadata['fps']}fps"
        )

    @pytest.mark.asyncio
    async def test_metadata_drift_detection(
        self, hybrid_monitor, service_manager, mock_dependencies
    ):
        """Test detection of metadata drift or inconsistencies between components."""

        device_path = "/dev/video0"

        # Setup capability in hybrid_monitor
        capability = CapabilityDetectionResult(
            device=device_path,
            detected=True,
            accessible=True,
            resolutions=["1920x1080", "1280x720"],
            frame_rates=["30", "25"],
        )

        state = hybrid_monitor._get_or_create_capability_state(device_path)
        state.confirmed_data = capability
        state.consecutive_successes = 5

        # Inject hybrid_monitor into service_manager
        service_manager._camera_monitor = hybrid_monitor

        camera_device = CameraDevice(
            device=device_path, name="Test Camera", driver="uvcvideo"
        )
        event_data = CameraEventData(
            device=device_path,
            event_type=CameraEvent.CONNECTED,
            device_info=camera_device,
        )

        # Get metadata from both sources
        hybrid_metadata = hybrid_monitor.get_effective_capability_metadata(device_path)
        service_metadata = await service_manager._get_enhanced_camera_metadata(event_data)

        # Comprehensive consistency check
        inconsistencies = []

        # Check validation status consistency
        if (
            hybrid_metadata["validation_status"]
            != service_metadata["validation_status"]
        ):
            inconsistencies.append(
                f"Validation status mismatch: hybrid={hybrid_metadata['validation_status']}, service={service_metadata['validation_status']}"
            )

        # Check consecutive successes consistency
        if (
            hybrid_metadata["consecutive_successes"]
            != service_metadata["consecutive_successes"]
        ):
            inconsistencies.append(
                f"Consecutive successes mismatch: hybrid={hybrid_metadata['consecutive_successes']}, service={service_metadata['consecutive_successes']}"
            )

        # Check resolution consistency
        if hybrid_metadata["resolution"] != service_metadata["resolution"]:
            inconsistencies.append(
                f"Resolution mismatch: hybrid={hybrid_metadata['resolution']}, service={service_metadata['resolution']}"
            )

        # Check FPS consistency
        if hybrid_metadata["fps"] != service_metadata["fps"]:
            inconsistencies.append(
                f"FPS mismatch: hybrid={hybrid_metadata['fps']}, service={service_metadata['fps']}"
            )

        # Check format availability consistency
        hybrid_formats = set(
            fmt["format"] for fmt in hybrid_metadata.get("formats", [])
        )
        service_formats = set(
            fmt["format"] for fmt in service_metadata.get("formats", [])
        )
        if hybrid_formats != service_formats:
            inconsistencies.append(
                f"Format mismatch: hybrid={hybrid_formats}, service={service_formats}"
            )

        # Report any inconsistencies
        if inconsistencies:
            print("❌ Metadata inconsistencies detected:")
            for inconsistency in inconsistencies:
                print(f"   - {inconsistency}")
            pytest.fail(
                f"Metadata drift detected: {len(inconsistencies)} inconsistencies found"
            )
        else:
            print("✅ No metadata drift detected - components are consistent")
            print(f"   - Validation status: {hybrid_metadata['validation_status']}")
            print(f"   - Resolution: {hybrid_metadata['resolution']}")
            print(f"   - FPS: {hybrid_metadata['fps']}")
            print(
                f"   - Consecutive successes: {hybrid_metadata['consecutive_successes']}"
            )

    @pytest.mark.asyncio
    async def test_frequency_weighted_reconciliation(
        self, hybrid_monitor, service_manager, mock_dependencies
    ):
        """Test frequency-weighted capability selection reconciliation."""

        device_path = "/dev/video0"

        # Simulate multiple detections with different frame rates to build frequency
        # data
        frame_rate_detections = [
            "30",
            "30",
            "25",
            "30",
            "60",
            "30",
        ]  # 30 appears most frequently
        resolution_detections = [
            "1920x1080",
            "1920x1080",
            "1280x720",
            "1920x1080",
        ]  # 1920x1080 most frequent

        state = hybrid_monitor._get_or_create_capability_state(device_path)

        for i, (fps, res) in enumerate(
            zip(frame_rate_detections, resolution_detections)
        ):
            capability = CapabilityDetectionResult(
                device=device_path,
                detected=True,
                accessible=True,
                resolutions=[res],
                frame_rates=[fps],
            )

            # Update frequency tracking
            hybrid_monitor._update_frequency_tracking(state, capability)
            await hybrid_monitor._update_capability_state(device_path, capability)

        # Inject hybrid_monitor into service_manager
        service_manager._camera_monitor = hybrid_monitor

        camera_device = CameraDevice(
            device=device_path, name="Frequency Test Camera", driver="uvcvideo"
        )
        event_data = CameraEventData(
            device=device_path,
            event_type=CameraEvent.CONNECTED,
            device_info=camera_device,
        )

        # Get metadata from service_manager
        metadata = await service_manager._get_enhanced_camera_metadata(event_data)

        # Verify frequency-weighted selections are consistent
        hybrid_metadata = hybrid_monitor.get_effective_capability_metadata(device_path)

        # Should select most frequent values
        assert metadata["fps"] == 30  # Most frequent frame rate
        assert "1920x1080" in metadata["resolution"]  # Most frequent resolution

        # Verify consistency between components
        assert metadata["fps"] == hybrid_metadata["fps"]
        assert metadata["resolution"] == hybrid_metadata["resolution"]

        print("✅ Frequency-weighted reconciliation verified:")
        print(
            f"   - Selected FPS: {metadata['fps']} (most frequent from {frame_rate_detections})"
        )
        print(
            f"   - Selected resolution: {metadata['resolution']} (most frequent from {resolution_detections})"
        )


class TestReconciliationErrorCases:
    """Test reconciliation behavior under error conditions."""

    @pytest.fixture
    def service_manager_with_broken_monitor(self, mock_dependencies):
        """Create service manager with broken camera monitor for error testing."""
        service_manager = ServiceManager(
            config=mock_dependencies["config"],
            mediamtx_controller=mock_dependencies["mediamtx_controller"],
            websocket_server=mock_dependencies["websocket_server"],
        )

        # Create mock monitor that raises errors
        broken_monitor = Mock()
        broken_monitor.get_effective_capability_metadata.side_effect = Exception(
            "Monitor error"
        )
        service_manager._camera_monitor = broken_monitor

        return service_manager

    @pytest.mark.asyncio
    async def test_reconciliation_with_monitor_error(
        self, service_manager_with_broken_monitor
    ):
        """Test reconciliation fallback when camera monitor raises errors."""

        device_path = "/dev/video0"

        camera_device = CameraDevice(
            device=device_path, name="Error Test Camera", driver="uvcvideo"
        )
        event_data = CameraEventData(
            device=device_path,
            event_type=CameraEvent.CONNECTED,
            device_info=camera_device,
        )

        # Should handle monitor error gracefully
        metadata = await service_manager_with_broken_monitor._get_enhanced_camera_metadata(
            event_data
        )

        # Should fall back to device info and defaults
        assert metadata["validation_status"] == "error"
        assert metadata["capability_source"] == "default"
        assert metadata["resolution"] == "1920x1080"  # Architecture default
        assert metadata["fps"] == 30  # Architecture default

        print("✅ Error reconciliation verified:")
        print(f"   - Validation status: {metadata['validation_status']}")
        print(
            f"   - Fallback to defaults: {metadata['resolution']}@{metadata['fps']}fps"
        )

    @pytest.mark.asyncio
    async def test_reconciliation_with_missing_monitor(self, mock_dependencies):
        """Test reconciliation when camera monitor is None."""

        service_manager = ServiceManager(
            config=mock_dependencies["config"],
            mediamtx_controller=mock_dependencies["mediamtx_controller"],
            websocket_server=mock_dependencies["websocket_server"],
        )
        # Don't set camera_monitor (remains None)

        device_path = "/dev/video0"
        camera_device = CameraDevice(
            device=device_path, name="No Monitor Camera", driver="uvcvideo"
        )
        event_data = CameraEventData(
            device=device_path,
            event_type=CameraEvent.CONNECTED,
            device_info=camera_device,
        )

        # Should handle missing monitor gracefully
        metadata = await service_manager._get_enhanced_camera_metadata(event_data)

        # Should fall back to device info
        assert metadata["validation_status"] == "none"
        assert metadata["capability_source"] == "device_info"
        assert metadata["name"] == "No Monitor Camera"  # From device info

        print("✅ Missing monitor reconciliation verified:")
        print(f"   - Validation status: {metadata['validation_status']}")
        print(f"   - Capability source: {metadata['capability_source']}")
        print(f"   - Name from device info: {metadata['name']}")
