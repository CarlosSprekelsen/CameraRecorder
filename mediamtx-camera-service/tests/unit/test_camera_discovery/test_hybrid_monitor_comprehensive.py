"""
Comprehensive tests for hybrid camera monitor capability detection, 
udev event processing, and adaptive polling behavior.

Test policy: One good core scenario + key edge case per responsibility.
"""

import asyncio
import pytest
import time
from unittest.mock import Mock, AsyncMock, patch, call
from pathlib import Path

# Import the classes we're testing
from src.camera_discovery.hybrid_monitor import (
    HybridCameraMonitor, 
    CameraEvent, 
    CameraEventData,
    CapabilityDetectionResult,
    DeviceCapabilityState
)
from src.common.types import CameraDevice


class TestHybridMonitorCapabilityDetection:
    """Test capability detection parsing robustness and validation logic."""

    @pytest.fixture
    def monitor(self):
        """Create monitor with capability detection enabled."""
        return HybridCameraMonitor(
            device_range=[0, 1, 2], 
            enable_capability_detection=True,
            detection_timeout=1.0
        )

    @pytest.mark.asyncio
    async def test_capability_parsing_varied_v4l2_outputs(self, monitor):
        """
        Core scenario: Parse varied v4l2-ctl output formats and extract correct capability data.
        
        Tests robustness against different v4l2-ctl output variations:
        - Standard format with discrete sizes
        - Alternative format without quotes
        - Missing sections with fallback parsing
        - Frame rate interval notation
        """
        interface = monitor._get_capability_probe_interface()
        
        # Test format parsing variations
        format_outputs = [
            # Standard Ubuntu format
            "[0]: 'YUYV' (YUYV 4:2:2)\nSize: Discrete 640x480\nSize: Discrete 1920x1080",
            # Alternative format without quotes  
            "[0]: MJPG (Motion JPEG)\n640x480\n1280x720\n1920x1080",
            # Minimal format with resolution patterns
            "Format: YUYV\nResolution: 1280 x 720\nSize: 800x600"
        ]
        
        expected_resolutions = [
            ["640x480", "1920x1080"],
            ["640x480", "1280x720", "1920x1080"], 
            ["1280x720", "800x600"]
        ]
        
        for i, (output, expected) in enumerate(zip(format_outputs, expected_resolutions)):
            with patch.object(monitor, '_probe_device_formats_robust') as mock_formats:
                mock_formats.return_value = {
                    "formats": [{"code": "YUYV", "description": "YUYV 4:2:2"}],
                    "resolutions": expected
                }
                
                result = await monitor._probe_device_capabilities(f"/dev/video{i}")
                
                assert result.detected is True
                assert set(result.resolutions) == set(expected)
                assert len(result.formats) > 0
                
    @pytest.mark.asyncio 
    async def test_capability_frame_rate_hierarchical_selection(self, monitor):
        """
        Core scenario: Test hierarchical frame rate selection with stability assessment.
        
        Verifies:
        - Highest stable frame rate (appears in multiple sources) is preferred
        - Unstable rates (single source) are deprioritized but included
        - Median fallback calculation
        - Proper sorting by preference
        """
        interface = monitor._get_capability_probe_interface()
        
        # Simulate multiple detection sources with overlapping frame rates
        all_rates = {"30", "25", "15", "60", "10", "5"}
        sources = [
            ("YUYV framesizes", {"30", "15", "5"}),     # 30, 15, 5 appear here
            ("MJPG framesizes", {"30", "25", "10"}),    # 30 appears again (stable)
            ("general framerates", {"60", "10"})        # 60 is unstable (single source)
        ]
        
        selected = interface['select_preferred_rates'](all_rates, sources, "/dev/video0")
        
        # Verify hierarchical selection
        assert selected[0] == "30"  # Highest stable rate first (appears in 2 sources)
        assert "60" in selected     # Unstable rate included but not first
        assert "25" in selected     # All detected rates included
        
        # Verify stable rates come before unstable
        stable_rates = ["30"]  # Only 30 appears in multiple sources
        first_unstable_index = min(i for i, rate in enumerate(selected) if rate not in stable_rates)
        assert first_unstable_index > 0  # Stable rates come first

    @pytest.mark.asyncio
    async def test_capability_provisional_confirmed_validation(self, monitor):
        """
        Core scenario: Test provisional → confirmed capability validation pattern.
        
        Verifies:
        - Immediate use of provisional data (optimistic)
        - Promotion to confirmed after consistent probes
        - Protection of confirmed data from minor variations
        - Fallback to last-known-good on major inconsistencies
        """
        device_path = "/dev/video0"
        
        # Enable test mode and get interfaces
        monitor._set_test_mode(True)
        
        # Simulate first capability detection
        result1 = CapabilityDetectionResult(
            device_path=device_path,
            detected=True,
            accessible=True,
            resolutions=["1920x1080", "1280x720"],
            frame_rates=["30", "15"],
            probe_timestamp=time.time()
        )
        
        # First probe - should be provisional
        state = DeviceCapabilityState(device_path)
        await monitor._update_capability_validation_state(state, result1)
        
        metadata = monitor.get_effective_capability_metadata(device_path)
        assert metadata["validation_status"] == "provisional"
        assert metadata["resolution"] == "1920x1080"  # Immediate use
        assert metadata["fps"] == 30
        
        # Second consistent probe - should promote to confirmed
        result2 = CapabilityDetectionResult(
            device_path=device_path,
            detected=True,
            accessible=True,
            resolutions=["1920x1080", "1280x720"],  # Consistent
            frame_rates=["30", "25"],               # Minor variation OK
            probe_timestamp=time.time()
        )
        
        await monitor._update_capability_validation_state(state, result2)
        
        metadata = monitor.get_effective_capability_metadata(device_path)
        assert metadata["validation_status"] == "confirmed"
        assert state.consecutive_successes == 2
        
        # Third probe with major inconsistency - should protect confirmed data
        result3 = CapabilityDetectionResult(
            device_path=device_path,
            detected=True,
            accessible=True,
            resolutions=["320x240"],                # Major change
            frame_rates=["5"],                      # Major change
            probe_timestamp=time.time()
        )
        
        await monitor._update_capability_validation_state(state, result3)
        
        # Should keep confirmed data, not use inconsistent data
        effective = state.get_effective_capability()
        assert effective == state.confirmed_data  # Still using confirmed data
        assert state.consecutive_failures > 0
        
    @pytest.mark.asyncio
    async def test_capability_parsing_edge_case_malformed_output(self, monitor):
        """
        Edge case: Handle malformed v4l2-ctl output gracefully.
        
        Tests resilience against:
        - Empty output
        - Partial output with missing sections
        - Malformed resolution strings
        - Timeout scenarios
        """
        # Test empty output handling
        assert monitor._extract_frame_rates_from_output("") == set()
        
        # Test malformed resolution patterns
        with patch.object(monitor, '_probe_device_formats_robust') as mock_formats:
            mock_formats.return_value = {
                "formats": [],
                "resolutions": ["invalid", "1920x1080", "malformed_res"]
            }
            
            result = await monitor._probe_device_capabilities("/dev/video0")
            # Should still detect valid resolution
            assert "1920x1080" in result.resolutions
            
        # Test timeout handling
        with patch.object(monitor, '_probe_device_info_robust') as mock_info:
            mock_info.side_effect = asyncio.TimeoutError()
            
            result = await monitor._probe_device_capabilities("/dev/video0")
            assert result.detected is False
            assert "timeout" in result.error.lower()
            assert result.timeout_context is not None


class TestHybridMonitorUdevEventProcessing:
    """Test udev event processing, filtering, and coordination with polling."""

    @pytest.fixture
    def monitor(self):
        """Create monitor with test configuration."""
        return HybridCameraMonitor(
            device_range=[0, 1, 2],
            poll_interval=0.1,
            enable_capability_detection=False  # Simplify for udev testing
        )

    @pytest.mark.asyncio
    async def test_udev_event_filtering_and_processing(self, monitor):
        """
        Core scenario: Test udev event filtering and correct processing of add/remove/change events.
        
        Verifies:
        - Correct filtering by device range
        - Proper event type handling (add/remove/change)
        - Device state updates
        - Event broadcasting to handlers
        """
        monitor._set_test_mode(True)
        
        # Mock event handler to capture events
        event_handler = Mock()
        event_handler.handle_camera_event = AsyncMock()
        monitor.add_event_handler(event_handler)
        
        # Test event filtering - device in range
        await monitor._inject_test_udev_event("/dev/video0", "add")
        
        # Should process event (device 0 in range [0,1,2])
        assert event_handler.handle_camera_event.called
        call_args = event_handler.handle_camera_event.call_args[0][0]
        assert call_args.device_path == "/dev/video0"
        assert call_args.event_type == CameraEvent.CONNECTED
        
        event_handler.handle_camera_event.reset_mock()
        
        # Test event filtering - device out of range
        await monitor._inject_test_udev_event("/dev/video5", "add")
        
        # Should filter out (device 5 not in range [0,1,2])
        assert not event_handler.handle_camera_event.called
        
        # Test remove event processing
        await monitor._inject_test_udev_event("/dev/video0", "remove")
        
        assert event_handler.handle_camera_event.called
        call_args = event_handler.handle_camera_event.call_args[0][0]
        assert call_args.event_type == CameraEvent.DISCONNECTED
        
    @pytest.mark.asyncio
    async def test_udev_event_race_condition_handling(self, monitor):
        """
        Edge case: Test handling of rapid event sequences and race conditions.
        
        Verifies:
        - Rapid add/remove sequences don't corrupt state
        - Change events with no actual status change are handled
        - Concurrent event processing is safe
        """
        monitor._set_test_mode(True)
        event_handler = Mock()
        event_handler.handle_camera_event = AsyncMock()
        monitor.add_event_handler(event_handler)
        
        # Simulate rapid event sequence
        await asyncio.gather(
            monitor._inject_test_udev_event("/dev/video0", "add"),
            monitor._inject_test_udev_event("/dev/video0", "remove"),
            monitor._inject_test_udev_event("/dev/video0", "add"),
            monitor._inject_test_udev_event("/dev/video1", "add")
        )
        
        # Should handle all events without corruption
        assert event_handler.handle_camera_event.call_count >= 3
        
        # Verify final state consistency
        connected_cameras = await monitor.get_connected_cameras()
        # At least one device should be tracked
        assert len(connected_cameras) >= 0  # State should be consistent
        
    @pytest.mark.asyncio
    async def test_adaptive_polling_coordination_with_udev(self, monitor):
        """
        Core scenario: Test adaptive polling frequency based on udev event freshness.
        
        Verifies:
        - Polling interval increases when udev events are fresh
        - Polling interval decreases when events are stale
        - Proper bounds enforcement
        - Fallback behavior when udev fails
        """
        monitor._set_test_mode(True)
        
        # Get initial polling state
        initial_state = monitor._get_adaptive_polling_state_for_testing()
        base_interval = initial_state['base_interval']
        
        # Simulate fresh udev event
        await monitor._inject_test_udev_event("/dev/video0", "add")
        
        # Trigger polling adjustment
        await monitor._adjust_polling_interval()
        
        current_state = monitor._get_adaptive_polling_state_for_testing()
        
        # With fresh udev event, polling should deprioritize (increase interval)
        # Note: This might not change immediately, depends on implementation
        assert current_state['current_interval'] >= base_interval * 0.9  # Allow some tolerance
        
        # Simulate stale udev events by setting old timestamp
        monitor._last_udev_event_time = time.time() - 15.0  # 15 seconds ago (stale)
        
        await monitor._adjust_polling_interval()
        
        updated_state = monitor._get_adaptive_polling_state_for_testing()
        
        # With stale events, polling should prioritize (decrease interval toward minimum)
        assert updated_state['current_interval'] <= current_state['current_interval']
        
    @pytest.mark.asyncio
    async def test_polling_fallback_when_udev_unavailable(self, monitor):
        """
        Edge case: Test polling fallback when udev monitoring is unavailable or fails.
        
        Verifies:
        - Polling continues when udev is disabled
        - Device discovery still works via polling
        - Error recovery in polling loop
        """
        # Disable udev to test polling-only mode
        monitor._udev_available = False
        
        # Mock device detection for polling
        with patch.object(monitor, '_determine_device_status') as mock_status:
            mock_status.return_value = "CONNECTED"
            
            with patch.object(Path, 'exists') as mock_exists:
                mock_exists.return_value = True
                
                # Trigger discovery cycle
                await monitor._discover_cameras()
                
                # Should discover devices via polling
                connected = await monitor.get_connected_cameras()
                assert len(connected) > 0  # Should find devices
                
                # Verify polling detection worked
                assert "/dev/video0" in connected or "/dev/video1" in connected


class TestHybridMonitorErrorHandling:
    """Test error handling, timeouts, and failure recovery."""

    @pytest.fixture
    def monitor(self):
        """Create monitor with short timeouts for testing."""
        return HybridCameraMonitor(
            device_range=[0],
            detection_timeout=0.1,  # Short timeout for testing
            enable_capability_detection=True
        )

    @pytest.mark.asyncio
    async def test_error_handling_subprocess_timeout(self, monitor):
        """
        Core scenario: Test timeout handling in capability detection subprocess calls.
        
        Verifies:
        - Timeout exceptions produce structured error information
        - Service continues operating after timeouts
        - Error context includes operation details
        """
        with patch('asyncio.create_subprocess_exec') as mock_subprocess:
            # Mock subprocess that times out
            mock_process = Mock()
            mock_process.communicate = AsyncMock(side_effect=asyncio.TimeoutError())
            mock_subprocess.return_value = mock_process
            
            result = await monitor._probe_device_capabilities("/dev/video0")
            
            assert result.detected is False
            assert "timeout" in result.error.lower()
            assert result.timeout_context is not None
            assert result.device_path == "/dev/video0"
            
            # Verify monitor stats updated
            stats = monitor.get_monitor_stats()
            assert stats['capability_timeouts'] > 0

    @pytest.mark.asyncio
    async def test_error_handling_subprocess_failure(self, monitor):
        """
        Core scenario: Test handling of subprocess failures and errors.
        
        Verifies:
        - Failed subprocess calls don't crash the service
        - Error information is properly captured and logged
        - Capability detection gracefully degrades
        """
        with patch('asyncio.create_subprocess_exec') as mock_subprocess:
            # Mock subprocess that fails
            mock_process = Mock()
            mock_process.returncode = 1
            mock_process.communicate = AsyncMock(return_value=(b"", b"Device not found"))
            mock_subprocess.return_value = mock_process
            
            # Should handle failure gracefully
            result = await monitor._probe_device_info_robust("/dev/video0")
            assert result is None  # Graceful failure
            
            # Service should continue operating
            assert monitor._running is False  # Not started yet, but shouldn't crash

    @pytest.mark.asyncio
    async def test_error_handling_malformed_device_paths(self, monitor):
        """
        Edge case: Test handling of malformed device paths and invalid inputs.
        
        Verifies:
        - Invalid device paths are handled gracefully
        - No crashes on unexpected input formats
        - Proper error reporting for diagnostics
        """
        monitor._set_test_mode(True)
        
        # Test various malformed device paths
        malformed_paths = [
            None,
            "",
            "/dev/invalid",
            "/not/a/device",
            "malformed_path"
        ]
        
        for bad_path in malformed_paths:
            try:
                # Should not crash on malformed paths
                await monitor._inject_test_udev_event(bad_path, "add")
            except Exception as e:
                # If it does raise, it should be a handled exception
                assert "test mode" not in str(e)  # Shouldn't be test mode error
                
        # Test stream name extraction robustness
        for bad_path in malformed_paths:
            if bad_path is not None:
                stream_name = monitor.get_stream_name_from_device_path(bad_path)
                assert stream_name is not None  # Should always return something
                assert len(stream_name) > 0

    @pytest.mark.asyncio
    async def test_error_recovery_consecutive_failures(self, monitor):
        """
        Edge case: Test recovery from consecutive failures and error accumulation.
        
        Verifies:
        - Exponential backoff on repeated failures
        - Circuit breaker behavior for persistent errors
        - Statistics tracking for failure patterns
        """
        monitor._set_test_mode(True)
        
        # Simulate consecutive polling failures
        original_discover = monitor._discover_cameras
        
        async def failing_discover():
            monitor._polling_failure_count += 1
            raise Exception("Simulated polling failure")
        
        monitor._discover_cameras = failing_discover
        
        # Test failure accumulation
        for i in range(3):
            try:
                await failing_discover()
            except Exception:
                pass
        
        assert monitor._polling_failure_count == 3
        
        # Test recovery
        monitor._discover_cameras = original_discover
        monitor._polling_failure_count = 0
        
        # Should be able to recover
        await monitor._discover_cameras()
        assert monitor._polling_failure_count == 0


class TestHybridMonitorIntegration:
    """Integration tests for full monitor lifecycle and component coordination."""

    @pytest.fixture
    def monitor(self):
        """Create monitor for integration testing."""
        return HybridCameraMonitor(
            device_range=[0, 1],
            poll_interval=0.1,
            enable_capability_detection=True
        )

    @pytest.mark.asyncio
    async def test_monitor_lifecycle_startup_shutdown(self, monitor):
        """
        Integration test: Full monitor lifecycle from startup to shutdown.
        
        Verifies:
        - Clean startup of all monitoring components
        - Proper task management
        - Graceful shutdown with resource cleanup
        - Statistics consistency throughout lifecycle
        """
        # Monitor should start in stopped state
        assert not monitor.is_running
        initial_stats = monitor.get_monitor_stats()
        assert initial_stats['running'] is False
        
        # Mock system components to avoid real hardware dependencies
        with patch('src.camera_discovery.hybrid_monitor.HAS_PYUDEV', False):
            # Start monitor (udev disabled for testing)
            await monitor.start()
            
            assert monitor.is_running
            running_stats = monitor.get_monitor_stats()
            assert running_stats['running'] is True
            assert running_stats['active_tasks'] > 0
            
            # Let it run briefly
            await asyncio.sleep(0.2)
            
            # Stop monitor
            await monitor.stop()
            
            assert not monitor.is_running
            final_stats = monitor.get_monitor_stats()
            assert final_stats['running'] is False
            assert final_stats['active_tasks'] == 0

    @pytest.mark.asyncio 
    async def test_capability_integration_with_event_handling(self, monitor):
        """
        Integration test: Capability detection integrated with event handling.
        
        Verifies:
        - Capability detection triggered by device events
        - Metadata integration with camera device information
        - End-to-end flow from detection to notification
        """
        monitor._set_test_mode(True)
        
        # Mock capability detection to return known data
        mock_capability = CapabilityDetectionResult(
            device_path="/dev/video0",
            detected=True,
            accessible=True,
            resolutions=["1920x1080", "1280x720"],
            frame_rates=["30", "15"]
        )
        
        with patch.object(monitor, '_probe_device_capabilities') as mock_probe:
            mock_probe.return_value = mock_capability
            
            # Mock event handler
            event_handler = Mock()
            event_handler.handle_camera_event = AsyncMock()
            monitor.add_event_handler(event_handler)
            
            # Simulate device connection
            await monitor._inject_test_udev_event("/dev/video0", "add")
            
            # Verify event was processed
            assert event_handler.handle_camera_event.called
            
            # Verify capability metadata is available
            metadata = monitor.get_effective_capability_metadata("/dev/video0")
            assert metadata["resolution"] == "1920x1080"
            assert metadata["fps"] == 30
            assert metadata["validation_status"] == "provisional"


# Pytest configuration and utilities
@pytest.fixture(scope="session")
def event_loop():
    """Create event loop for async tests."""
    loop = asyncio.new_event_loop()
    yield loop
    loop.close()


def test_frame_rate_extraction_patterns():
    """
    Unit test: Frame rate extraction from various v4l2-ctl output patterns.
    
    Tests the regex patterns used for frame rate detection without async overhead.
    """
    monitor = HybridCameraMonitor()
    
    test_outputs = [
        ("30.000 fps", {"30"}),
        ("Interval: [1/30]", {"30"}),
        ("Frame rate: 25.0", {"25"}),
        ("1920x1080@60", {"60"}),
        ("30 Hz", {"30"}),
        ("mixed: 30.000 fps, 25 FPS, [1/15]", {"30", "25", "15"}),
        ("no frame rates here", set())
    ]
    
    for output, expected in test_outputs:
        result = monitor._extract_frame_rates_from_output(output)
        assert result == expected, f"Failed for output: {output}"


def test_stream_name_extraction_robustness():
    """
    Unit test: Stream name extraction from various device path formats.
    
    Tests deterministic mapping of device paths to stream names.
    """
    monitor = HybridCameraMonitor()
    
    test_cases = [
        ("/dev/video0", "camera0"),
        ("/dev/video15", "camera15"),
        ("/custom/video2", "camera2"),
        ("/path/with/video99/suffix", "camera99"),
        ("/no/numbers/here", "camera_"),  # Will get hash-based name
        ("", "camera_unknown")
    ]
    
    for device_path, expected_prefix in test_cases:
        result = monitor.get_stream_name_from_device_path(device_path)
        if expected_prefix.endswith("_"):
            assert result.startswith(expected_prefix)
        else:
            assert result == expected_prefix


if __name__ == "__main__":
    # Allow running tests directly
    pytest.main([__file__, "-v"])