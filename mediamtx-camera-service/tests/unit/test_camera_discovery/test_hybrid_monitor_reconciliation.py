"""
Reconciliation tests between hybrid_monitor capability output and service_manager consumption.

Test coverage:
- End-to-end capability flow validation
- Provisional vs confirmed state propagation
- Metadata consistency and drift detection
- Integration validation between components

Created: 2025-08-04
Related: S3 Camera Discovery hardening, docs/roadmap.md
Evidence: src/camera_discovery/hybrid_monitor.py lines 750-800 (get_effective_capability_metadata)
Evidence: src/camera_service/service_manager.py lines 270-380 (_get_enhanced_camera_metadata)
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
    def mock_dependencies(self):
        """Create mock dependencies for service manager."""
        return {
            "config": Mock(),
            "mediamtx_controller": Mock(),
            "websocket_server": Mock(),
            "camera_monitor": None,  # Will be set per test
        }

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
            device_path=device_path,
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
            device_path=device_path, name="Test Camera Device", driver="uvcvideo"
        )
        event_data = CameraEventData(
            device_path=device_path,
            event_type=CameraEvent.CONNECTED,
            device_info=camera_device,
        )

        # Get metadata from service_manager (this calls hybrid_monitor internally)
        metadata = service_manager._get_enhanced_camera_metadata(event_data)

        # Verify reconciliation - confirmed data should propagate
        assert metadata["validation_status"] == "confirmed"
        assert metadata["capability_source"] == "confirmed_capability"
        assert metadata["consecutive_successes"] == 5

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
            device_path=device_path,
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
            device_path=device_path, name="Provisional Camera Device", driver="uvcvideo"
        )
        event_data = CameraEventData(
            device_path=device_path,
            event_type=CameraEvent.CONNECTED,
            device_info=camera_device,
        )

        # Get metadata from service_manager
        metadata = service_manager._get_enhanced_camera_metadata(event_data)

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
            device_path=device_path, name="Unknown Capability Camera", driver="unknown"
        )
        event_data = CameraEventData(
            device_path=device_path,
            event_type=CameraEvent.CONNECTED,
            device_info=camera_device,
        )

        # Get metadata from service_manager
        metadata = service_manager._get_enhanced_camera_metadata(event_data)

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
            device_path=device_path,
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
            device_path=device_path, name="Test Camera", driver="uvcvideo"
        )
        event_data = CameraEventData(
            device_path=device_path,
            event_type=CameraEvent.CONNECTED,
            device_info=camera_device,
        )

        # Get initial metadata (should be provisional)
        initial_metadata = service_manager._get_enhanced_camera_metadata(event_data)
        assert initial_metadata["validation_status"] == "provisional"

        # Simulate additional consistent detections to trigger confirmation
        for _ in range(3):  # Reach confirmation threshold
            hybrid_monitor._update_capability_state(device_path, provisional_capability)

        # Get metadata after confirmation
        confirmed_metadata = service_manager._get_enhanced_camera_metadata(event_data)

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
            device_path=device_path,
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
            device_path=device_path, name="Test Camera", driver="uvcvideo"
        )
        event_data = CameraEventData(
            device_path=device_path,
            event_type=CameraEvent.CONNECTED,
            device_info=camera_device,
        )

        # Get metadata from both sources
        hybrid_metadata = hybrid_monitor.get_effective_capability_metadata(device_path)
        service_metadata = service_manager._get_enhanced_camera_metadata(event_data)

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

        # Simulate multiple detections with different frame rates to build frequency data
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
                device_path=device_path,
                detected=True,
                accessible=True,
                resolutions=[res],
                frame_rates=[fps],
            )

            # Update frequency tracking
            hybrid_monitor._update_frequency_tracking(state, capability)
            hybrid_monitor._update_capability_state(device_path, capability)

        # Inject hybrid_monitor into service_manager
        service_manager._camera_monitor = hybrid_monitor

        camera_device = CameraDevice(
            device_path=device_path, name="Frequency Test Camera", driver="uvcvideo"
        )
        event_data = CameraEventData(
            device_path=device_path,
            event_type=CameraEvent.CONNECTED,
            device_info=camera_device,
        )

        # Get metadata from service_manager
        metadata = service_manager._get_enhanced_camera_metadata(event_data)

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
            device_path=device_path, name="Error Test Camera", driver="uvcvideo"
        )
        event_data = CameraEventData(
            device_path=device_path,
            event_type=CameraEvent.CONNECTED,
            device_info=camera_device,
        )

        # Should handle monitor error gracefully
        metadata = service_manager_with_broken_monitor._get_enhanced_camera_metadata(
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
            device_path=device_path, name="No Monitor Camera", driver="uvcvideo"
        )
        event_data = CameraEventData(
            device_path=device_path,
            event_type=CameraEvent.CONNECTED,
            device_info=camera_device,
        )

        # Should handle missing monitor gracefully
        metadata = service_manager._get_enhanced_camera_metadata(event_data)

        # Should fall back to device info
        assert metadata["validation_status"] == "none"
        assert metadata["capability_source"] == "device_info"
        assert metadata["name"] == "No Monitor Camera"  # From device info

        print("✅ Missing monitor reconciliation verified:")
        print(f"   - Validation status: {metadata['validation_status']}")
        print(f"   - Capability source: {metadata['capability_source']}")
        print(f"   - Name from device info: {metadata['name']}")
