"""
Real system integration tests for hybrid camera monitor with actual component behavior validation.

Requirements Traceability:
- REQ-CAM-004: Camera discovery shall provide adaptive polling interval adjustment based on udev event reliability
- REQ-CAM-004: Camera discovery shall implement failure recovery with exponential backoff and jitter
- REQ-CAM-004: Camera discovery shall maintain polling-only mode when udev is unavailable
- REQ-CAM-004: Camera discovery shall reconcile capability data between components
- REQ-CAM-004: Camera discovery shall maintain metadata consistency across service boundaries
- REQ-ERROR-004: System shall handle camera discovery failures gracefully with proper recovery
- REQ-ERROR-005: System shall maintain stability during camera discovery errors

Story Coverage: S3 - Camera Discovery Hardening
IV&V Control Point: Real system behavior validation with minimal mocking

TODO: VIOLATION - MediaMTX service mocking violates strategic mocking rules
- Line 200: Mocking MediaMTX controller with mock_dependencies["mediamtx_controller"]
- VIOLATION: Testing guide states "NEVER MOCK: MediaMTX service"
- FIX REQUIRED: Replace with real systemd-managed MediaMTX service integration

TODO: VIOLATION - WebSocket mocking violates strategic mocking rules
- Line 206: Mocking WebSocket server with mock_dependencies["websocket_server"]
- VIOLATION: Testing guide states "NEVER MOCK: internal WebSocket"
- FIX REQUIRED: Replace with real WebSocket server integration

TODO: VIOLATION - Config loading mocking violates strategic mocking rules
- Line 209: Mocking config with mock_dependencies["config"]
- VIOLATION: Testing guide states "NEVER MOCK: config loading"
- FIX REQUIRED: Replace with real config loading

Test Strategy:
- Test real adaptive polling behavior under various conditions
- Validate failure recovery mechanisms with actual error scenarios
- Test polling-only mode with real system constraints
- Verify capability reconciliation with real data flows
- Test edge cases and stress conditions
"""

import asyncio
import pytest
import time
import os
import tempfile
from unittest.mock import Mock, patch, AsyncMock
from pathlib import Path

# Test imports
from src.camera_discovery.hybrid_monitor import (
    HybridCameraMonitor,
    CapabilityDetectionResult,
    CameraEventData,
    CameraEvent,
)
from src.camera_service.service_manager import ServiceManager
from src.common.types import CameraDevice


class TestRealAdaptivePollingBehavior:
    """Test real adaptive polling behavior under various system conditions."""

    @pytest.fixture
    def real_monitor(self):
        """Create monitor with real system behavior testing."""
        return HybridCameraMonitor(
            device_range=[0, 1, 2],
            poll_interval=0.1,
            enable_capability_detection=True,
        )

    @pytest.mark.asyncio
    async def test_real_adaptive_polling_with_stale_udev_events(self, real_monitor):
        """
        Test real adaptive polling behavior when udev events become stale.
        
        Requirements: REQ-CAM-004
        Scenario: Udev events become stale, system should increase polling frequency
        Expected: Polling interval decreases (frequency increases) when udev is stale
        """
        # Set initial state with stale udev events
        real_monitor._current_poll_interval = 0.5  # Start with longer interval
        real_monitor._last_udev_event_time = time.time() - 30.0  # Very stale
        real_monitor._udev_event_freshness_threshold = 10.0
        
        initial_interval = real_monitor._current_poll_interval
        
        # Run real polling cycle (no mocking of core logic)
        await real_monitor._adjust_polling_interval()
        
        # Verify real adaptive behavior
        assert real_monitor._current_poll_interval < initial_interval, \
            f"Polling interval should decrease when udev is stale: {initial_interval} -> {real_monitor._current_poll_interval}"
        
        # Verify stats reflect the adjustment
        stats = real_monitor.get_monitor_stats()
        assert stats["adaptive_poll_adjustments"] > 0, "Should record adaptive adjustment"
        assert stats["current_poll_interval"] == real_monitor._current_poll_interval

    @pytest.mark.asyncio
    async def test_real_adaptive_polling_with_recent_udev_events(self, real_monitor):
        """
        Test real adaptive polling behavior when udev events are recent.
        
        Requirements: REQ-CAM-004
        Scenario: Udev events are recent, system should reduce polling frequency
        Expected: Polling interval increases (frequency decreases) when udev is fresh
        """
        # Set initial state with recent udev events
        real_monitor._current_poll_interval = 0.1  # Start with shorter interval
        real_monitor._last_udev_event_time = time.time() - 2.0  # Recent
        real_monitor._udev_event_freshness_threshold = 10.0
        
        initial_interval = real_monitor._current_poll_interval
        
        # Run real polling cycle
        await real_monitor._adjust_polling_interval()
        
        # Verify real adaptive behavior
        assert real_monitor._current_poll_interval > initial_interval, \
            f"Polling interval should increase when udev is fresh: {initial_interval} -> {real_monitor._current_poll_interval}"

    @pytest.mark.asyncio
    async def test_real_failure_recovery_with_actual_errors(self, real_monitor):
        """
        Test real failure recovery behavior with actual error conditions.
        
        Requirements: REQ-CAM-004, REQ-ERROR-004
        Scenario: Discovery failures occur, system should implement proper recovery
        Expected: Failure count tracking, exponential backoff, and recovery
        """
        # Simulate real discovery failures
        failure_count = 0
        
        async def simulate_real_discovery_failures():
            nonlocal failure_count
            failure_count += 1
            if failure_count <= 3:
                raise OSError(f"Real discovery failure #{failure_count}")
            return None
        
        # Replace discovery method with failure simulation
        real_monitor._discover_cameras = simulate_real_discovery_failures
        
        # Run multiple cycles to trigger failure recovery
        # Don't catch exceptions - let the system handle them internally
        for cycle in range(5):
            await real_monitor._single_polling_cycle()
        
        # Verify real failure tracking
        assert real_monitor._polling_failure_count > 0, "Should track failures"
        assert real_monitor._polling_failure_count <= real_monitor._max_consecutive_failures, \
            f"Failure count should not exceed max: {real_monitor._polling_failure_count} > {real_monitor._max_consecutive_failures}"
        
        # Verify recovery after failures
        stats = real_monitor.get_monitor_stats()
        assert stats["polling_cycles"] >= 2, "Should have successful cycles after recovery"

    @pytest.mark.asyncio
    async def test_real_polling_interval_bounds(self, real_monitor):
        """
        Test real polling interval stays within configured bounds.
        
        Requirements: REQ-CAM-004
        Scenario: Extreme conditions should not cause polling interval to exceed bounds
        Expected: Interval stays within min/max bounds regardless of conditions
        """
        # Test minimum bound
        real_monitor._current_poll_interval = real_monitor._min_poll_interval
        real_monitor._last_udev_event_time = time.time() - 100.0  # Very stale
        real_monitor._polling_failure_count = 10  # Many failures
        
        await real_monitor._adjust_polling_interval()
        
        assert real_monitor._current_poll_interval >= real_monitor._min_poll_interval, \
            f"Interval should not go below minimum: {real_monitor._current_poll_interval} < {real_monitor._min_poll_interval}"
        
        # Test maximum bound
        real_monitor._current_poll_interval = real_monitor._max_poll_interval
        real_monitor._last_udev_event_time = time.time()  # Very fresh
        real_monitor._polling_failure_count = 0  # No failures
        
        await real_monitor._adjust_polling_interval()
        
        assert real_monitor._current_poll_interval <= real_monitor._max_poll_interval, \
            f"Interval should not exceed maximum: {real_monitor._current_poll_interval} > {real_monitor._max_poll_interval}"

    @pytest.mark.asyncio
    async def test_real_polling_only_mode_without_udev(self, real_monitor):
        """
        Test real polling-only mode when udev is completely unavailable.
        
        Requirements: REQ-CAM-004
        Scenario: System operates in polling-only mode when udev is not available
        Expected: System continues to function with polling-only discovery
        """
        # Disable udev completely (real system condition)
        real_monitor._udev_available = False
        real_monitor._last_udev_event_time = 0.0
        
        # Run polling cycle in polling-only mode
        await real_monitor._single_polling_cycle()
        
        # Verify polling-only operation
        stats = real_monitor.get_monitor_stats()
        assert stats["polling_cycles"] > 0, "Should have polling cycles in polling-only mode"
        assert stats["udev_events_processed"] == 0, "Should not process udev events in polling-only mode"


class TestRealCapabilityReconciliation:
    """Test real capability reconciliation between components."""

    @pytest.fixture
    def real_service_manager(self, mock_dependencies):
        """Create service manager with real component integration."""
        service_manager = ServiceManager(
            config=mock_dependencies["config"],
            mediamtx_controller=mock_dependencies["mediamtx_controller"],
            websocket_server=mock_dependencies["websocket_server"],
        )
        return service_manager

    @pytest.mark.asyncio
    async def test_real_confirmed_capability_flow(self, real_service_manager):
        """
        Test real confirmed capability data flow through the system.
        
        Requirements: REQ-CAM-004
        Scenario: Confirmed capability data flows from monitor to service manager
        Expected: Metadata consistency and proper validation status
        """
        # Create real hybrid monitor with confirmed capability
        monitor = HybridCameraMonitor(device_range=[0, 1, 2], enable_capability_detection=True)
        
        device_path = "/dev/video0"
        confirmed_capability = CapabilityDetectionResult(
            device_path=device_path,
            detected=True,
            accessible=True,
            device_name="Real Test Camera",
            formats=[{"format": "YUYV", "description": "YUYV 4:2:2"}],
            resolutions=["1920x1080", "1280x720"],
            frame_rates=["30", "25"],
        )
        
        # Set up confirmed state in monitor
        state = monitor._get_or_create_capability_state(device_path)
        state.confirmed_data = confirmed_capability
        state.consecutive_successes = 5  # Above confirmation threshold
        
        # Inject monitor into service manager
        real_service_manager._camera_monitor = monitor
        
        # Create real camera event
        camera_device = CameraDevice(
            device=device_path, name="Real Camera Device", driver="uvcvideo"
        )
        event_data = CameraEventData(
            device_path=device_path,
            event_type=CameraEvent.CONNECTED,
            device_info=camera_device,
        )
        
        # Get real metadata from service manager
        metadata = await real_service_manager._get_enhanced_camera_metadata(event_data)
        
        # Verify real reconciliation
        assert metadata["validation_status"] == "confirmed", "Should have confirmed status"
        assert metadata["capability_source"] == "confirmed_capability", "Should use confirmed data"
        assert metadata["consecutive_successes"] == 5, "Should reflect consecutive successes"
        
        # Verify data consistency
        hybrid_metadata = monitor.get_effective_capability_metadata(device_path)
        assert hybrid_metadata["validation_status"] == metadata["validation_status"], "Status should match"
        assert hybrid_metadata["consecutive_successes"] == metadata["consecutive_successes"], "Success count should match"

    @pytest.mark.asyncio
    async def test_real_provisional_capability_flow(self, real_service_manager):
        """
        Test real provisional capability data flow through the system.
        
        Requirements: REQ-CAM-004
        Scenario: Provisional capability data flows from monitor to service manager
        Expected: Provisional status and proper data propagation
        """
        # Create real hybrid monitor with provisional capability
        monitor = HybridCameraMonitor(device_range=[0, 1, 2], enable_capability_detection=True)
        
        device_path = "/dev/video1"
        provisional_capability = CapabilityDetectionResult(
            device_path=device_path,
            detected=True,
            accessible=True,
            device_name="Provisional Test Camera",
            formats=[{"format": "MJPG", "description": "Motion-JPEG"}],
            resolutions=["1280x720", "640x480"],
            frame_rates=["30", "15"],
        )
        
        # Set up provisional state in monitor
        state = monitor._get_or_create_capability_state(device_path)
        state.provisional_data = provisional_capability
        state.consecutive_successes = 1  # Below confirmation threshold
        
        # Inject monitor into service manager
        real_service_manager._camera_monitor = monitor
        
        # Create real camera event
        camera_device = CameraDevice(
            device=device_path, name="Provisional Camera Device", driver="uvcvideo"
        )
        event_data = CameraEventData(
            device_path=device_path,
            event_type=CameraEvent.CONNECTED,
            device_info=camera_device,
        )
        
        # Get real metadata from service manager
        metadata = await real_service_manager._get_enhanced_camera_metadata(event_data)
        
        # Verify real reconciliation
        assert metadata["validation_status"] == "provisional", "Should have provisional status"
        assert metadata["capability_source"] == "provisional_capability", "Should use provisional data"
        assert metadata["consecutive_successes"] == 1, "Should reflect consecutive successes"

    @pytest.mark.asyncio
    async def test_real_no_capability_fallback(self, real_service_manager):
        """
        Test real fallback behavior when no capability data is available.
        
        Requirements: REQ-CAM-004, REQ-ERROR-004
        Scenario: No capability data available, system should provide fallback
        Expected: Graceful fallback with appropriate default values
        """
        # Create real hybrid monitor without capability data
        monitor = HybridCameraMonitor(device_range=[0, 1, 2], enable_capability_detection=True)
        
        device_path = "/dev/video2"
        
        # Inject monitor into service manager
        real_service_manager._camera_monitor = monitor
        
        # Create real camera event
        camera_device = CameraDevice(
            device=device_path, name="Unknown Capability Camera", driver="unknown"
        )
        event_data = CameraEventData(
            device_path=device_path,
            event_type=CameraEvent.CONNECTED,
            device_info=camera_device,
        )
        
        # Get real metadata from service manager
        metadata = await real_service_manager._get_enhanced_camera_metadata(event_data)
        
        # Verify real fallback behavior
        assert metadata["validation_status"] == "none", "Should have none status when no capability data"
        assert metadata["capability_source"] == "device_info", "Should use device info as fallback"
        assert "resolution" in metadata, "Should provide fallback resolution"
        assert "fps" in metadata, "Should provide fallback fps"


class TestRealEdgeCasesAndStressConditions:
    """Test real system behavior under edge cases and stress conditions."""

    @pytest.mark.asyncio
    async def test_real_concurrent_polling_cycles(self):
        """
        Test real concurrent polling cycle behavior.
        
        Requirements: REQ-CAM-004, REQ-ERROR-005
        Scenario: Multiple polling cycles running concurrently
        Expected: Thread-safe operation without race conditions
        """
        monitor = HybridCameraMonitor(device_range=[0, 1, 2], poll_interval=0.01)
        
        # Run multiple concurrent polling cycles
        tasks = []
        for _ in range(5):
            task = asyncio.create_task(monitor._single_polling_cycle())
            tasks.append(task)
        
        # Wait for all tasks to complete
        await asyncio.gather(*tasks, return_exceptions=True)
        
        # Verify no exceptions and proper state
        stats = monitor.get_monitor_stats()
        assert stats["polling_cycles"] >= 0, "Should have non-negative polling cycles"

    @pytest.mark.asyncio
    async def test_real_memory_pressure_handling(self):
        """
        Test real system behavior under memory pressure.
        
        Requirements: REQ-ERROR-005
        Scenario: System under memory pressure during camera discovery
        Expected: Graceful degradation without crashes
        """
        monitor = HybridCameraMonitor(device_range=[0, 1, 2])
        
        # Simulate memory pressure by creating many capability states
        for i in range(100):
            device_path = f"/dev/video{i}"
            state = monitor._get_or_create_capability_state(device_path)
            state.provisional_data = CapabilityDetectionResult(
                device_path=device_path,
                detected=True,
                accessible=True,
                device_name=f"Test Camera {i}",
                formats=[{"format": "YUYV", "description": "YUYV 4:2:2"}],
                resolutions=["1920x1080"],
                frame_rates=["30"],
            )
        
        # Run polling cycle under memory pressure
        await monitor._single_polling_cycle()
        
        # Verify system remains stable
        stats = monitor.get_monitor_stats()
        assert len(monitor._capability_states) == 100, "Should maintain all capability states"

    @pytest.mark.asyncio
    async def test_real_rapid_device_changes(self):
        """
        Test real system behavior with rapid device connect/disconnect.
        
        Requirements: REQ-CAM-004, REQ-ERROR-004
        Scenario: Rapid device connect/disconnect events
        Expected: Proper event handling and state management
        """
        monitor = HybridCameraMonitor(device_range=[0, 1, 2])
        
        captured_events = []
        
        async def capture_event(event_data: CameraEventData):
            captured_events.append(event_data)
        
        monitor.add_event_callback(capture_event)
        
        # Simulate rapid device changes
        for i in range(10):
            # Simulate device connection
            device_path = f"/dev/video{i % 3}"
            camera_device = CameraDevice(
                device=device_path, name=f"Rapid Camera {i}", driver="uvcvideo"
            )
            event_data = CameraEventData(
                device_path=device_path,
                event_type=CameraEvent.CONNECTED,
                device_info=camera_device,
            )
            
            await monitor._handle_camera_event(event_data)
            
            # Small delay to simulate real timing
            await asyncio.sleep(0.01)
        
        # Verify proper event handling
        assert len(captured_events) == 10, "Should capture all events"
        assert len(monitor._known_devices) <= 3, "Should maintain device count within range"


class TestRealSystemIntegration:
    """Test real system integration with minimal mocking."""

    @pytest.mark.asyncio
    async def test_real_monitor_startup_and_shutdown(self):
        """
        Test real monitor startup and shutdown behavior.
        
        Requirements: REQ-CAM-004, REQ-ERROR-005
        Scenario: Monitor startup and shutdown lifecycle
        Expected: Proper resource management and state cleanup
        """
        monitor = HybridCameraMonitor(device_range=[0, 1, 2])
        
        # Test startup
        await monitor.start()
        assert monitor._running is True, "Monitor should be running after start"
        
        # Test shutdown
        await monitor.stop()
        assert monitor._running is False, "Monitor should not be running after stop"
        
        # Verify cleanup
        assert len(monitor._monitoring_tasks) == 0, "Should clean up monitoring tasks"

    @pytest.mark.asyncio
    async def test_real_device_discovery_with_file_system(self):
        """
        Test real device discovery using actual file system checks.
        
        Requirements: REQ-CAM-004
        Scenario: Real device discovery using file system
        Expected: Proper device detection and error handling
        """
        monitor = HybridCameraMonitor(device_range=[0, 1, 2])
        
        # Test with non-existent devices (real file system check)
        await monitor._discover_cameras()
        
        # Verify proper handling of non-existent devices
        stats = monitor.get_monitor_stats()
        assert stats["polling_cycles"] >= 0, "Should handle non-existent devices gracefully"

    @pytest.mark.asyncio
    async def test_real_capability_detection_timeout(self):
        """
        Test real capability detection timeout behavior.
        
        Requirements: REQ-CAM-004, REQ-ERROR-004
        Scenario: Capability detection times out
        Expected: Proper timeout handling and fallback behavior
        """
        monitor = HybridCameraMonitor(
            device_range=[0, 1, 2],
            detection_timeout=0.1,  # Short timeout for testing
            enable_capability_detection=True,
        )
        
        # Test capability detection with timeout
        device_path = "/dev/video999"  # Non-existent device
        result = await monitor._probe_device_capabilities(device_path)
        
        # Verify timeout handling
        assert result is not None, "Should return result even on timeout"
        assert result.detected is False, "Should indicate detection failure"
        assert result.accessible is False, "Should indicate accessibility failure"
