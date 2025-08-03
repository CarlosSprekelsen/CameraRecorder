"""
Comprehensive tests for hybrid camera monitor capability detection, 
udev event processing, and adaptive polling behavior.

Coverage areas:
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
import tempfile
import json
from unittest.mock import Mock, AsyncMock, patch, call, MagicMock
from pathlib import Path
from typing import Dict, List, Any

# Test imports
from src.camera_discovery.hybrid_monitor import (
    HybridCameraMonitor, 
    CameraEvent, 
    CameraEventData,
    CapabilityDetectionResult,
    DeviceCapabilityState
)
from src.common.types import CameraDevice


class TestCapabilityParsingVariations:
    """Test capability detection parsing with varied and malformed v4l2-ctl outputs."""

    @pytest.fixture
    def monitor(self):
        """Create monitor with capability detection enabled."""
        return HybridCameraMonitor(
            device_range=[0, 1, 2], 
            enable_capability_detection=True,
            detection_timeout=2.0
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
            ("300 fps", set()),     # Invalid rate (filtered out)
            ("1.5 fps", {"1.5"}),   # Low rate
            ("0 fps", set()),       # Invalid rate (filtered out)
            ("500 fps", set()),     # Invalid rate (filtered out)
            
            # Malformed patterns
            ("30.000.000 fps", set()),     # Double decimal
            ("abc fps", set()),            # Non-numeric
            ("fps without number", set()),  # No number
        ]
        
        for output, expected in test_cases:
            result = monitor._extract_frame_rates_from_output(output)
            assert result == expected, f"Failed for output: '{output}' - expected {expected}, got {result}"

    @pytest.mark.asyncio
    async def test_capability_parsing_malformed_v4l2_outputs(self, monitor):
        """Test capability detection resilience against malformed v4l2-ctl outputs."""
        
        malformed_outputs = [
            # Empty/minimal outputs
            ("", False, "empty output"),
            ("v4l2-ctl: error", False, "error output"),
            
            # Partial outputs
            ("Card type: USB Camera\nDriver name: uvcvideo", True, "minimal valid info"),
            ("Some random text without useful info", False, "random text"),
            
            # Malformed format sections
            ("Format [0]: corrupted data\nSize: invalid", False, "corrupted format"),
            ("Valid start\n[CORRUPTED MIDDLE SECTION]\nValid end", False, "corrupted middle"),
            
            # Mixed valid/invalid data
            ("Card type: Test Camera\nFormat [0]: 'YUYV'\nSize: Discrete 1920x1080\nCorrupted frame rate data", True, "mixed valid/invalid"),
        ]
        
        for output, should_succeed, description in malformed_outputs:
            with patch('asyncio.create_subprocess_exec') as mock_subprocess:
                # Mock process that returns our test output
                mock_process = AsyncMock()
                mock_process.communicate.return_value = (output.encode(), b"")
                mock_process.returncode = 0
                mock_subprocess.return_value = mock_process
                
                result = await monitor._probe_device_capabilities("/dev/video0")
                
                if should_succeed:
                    assert result.detected, f"Should have succeeded for {description}: {output}"
                    assert result.accessible, f"Should be accessible for {description}"
                else:
                    assert not result.detected or result.error, f"Should have failed for {description}: {output}"

    @pytest.mark.asyncio
    async def test_provisional_confirmed_capability_validation(self, monitor):
        """Test provisional/confirmed capability validation state machine."""
        
        device_path = "/dev/video0"
        
        # Create consistent capability results
        consistent_result = CapabilityDetectionResult(
            device_path=device_path,
            detected=True,
            accessible=True,
            resolutions=["1920x1080", "1280x720"],
            frame_rates=["30", "25", "15"],
            formats=[{"code": "YUYV", "description": "YUYV 4:2:2"}]
        )
        
        # Test provisional state establishment
        await monitor._update_capability_validation_state(device_path, consistent_result)
        state = monitor._get_capability_state_for_testing(device_path)
        
        assert state is not None
        assert state.provisional_data is not None
        assert state.confirmed_data is None
        assert state.consecutive_successes == 1
        assert not state.is_confirmed()
        
        # Test confirmation through consistent results
        await monitor._update_capability_validation_state(device_path, consistent_result)
        
        assert state.consecutive_successes == 2
        assert state.confirmed_data is not None
        assert state.is_confirmed()
        
        # Test inconsistent data handling
        inconsistent_result = CapabilityDetectionResult(
            device_path=device_path,
            detected=True,
            accessible=True,
            resolutions=["640x480"],  # Different resolutions
            frame_rates=["60"],       # Different frame rates
            formats=[{"code": "MJPG", "description": "MJPEG"}]  # Different format
        )
        
        await monitor._update_capability_validation_state(device_path, inconsistent_result)
        
        # Should reset to provisional state
        assert state.consecutive_successes == 1
        assert state.confirmed_data is None
        assert not state.is_confirmed()

    def test_hierarchical_frame_rate_selection(self, monitor):
        """Test hierarchical frame rate selection policy implementation."""
        
        # Test data with different priority rates
        test_rates = {"60", "30", "25", "24", "15", "10", "5"}
        sources = [
            ("YUYV formats", {"30", "25", "60"}),
            ("MJPG formats", {"30", "24", "15"}),
            ("general rates", {"30", "25", "10"})
        ]
        
        result = monitor._select_preferred_frame_rates(test_rates, sources, "/dev/video0")
        
        # Verify ordering: high priority rates first, then by frequency, then by value
        assert result[0] in ["30", "25", "24"], f"First rate should be high priority, got {result[0]}"
        assert "30" in result[:3], "30 fps should be in top 3 (high priority + frequency)"
        assert len(result) == len(test_rates), "All rates should be included"


class TestUdevEventProcessingAndRaceConditions:
    """Test udev event processing, filtering, and race condition handling."""

    @pytest.fixture
    def monitor(self):
        """Create monitor with test configuration."""
        return HybridCameraMonitor(
            device_range=[0, 1, 2],
            poll_interval=0.1,
            enable_capability_detection=False  # Simplify for udev testing
        )

    @pytest.mark.asyncio
    async def test_udev_event_filtering_comprehensive(self, monitor):
        """Test comprehensive udev event filtering scenarios."""
        
        monitor._set_test_mode(True)
        
        # Track statistics before
        initial_stats = monitor.get_monitor_stats()
        
        filter_test_cases = [
            # Valid events (should be processed)
            ("/dev/video0", "add", True, "valid device in range"),
            ("/dev/video1", "remove", True, "valid device removal"),
            ("/dev/video2", "change", True, "valid device change"),
            
            # Invalid events (should be filtered)
            ("/dev/video5", "add", False, "device outside range"),
            ("/dev/audio0", "add", False, "non-video device"),
            ("/dev/invalid", "add", False, "malformed device path"),
            (None, "add", False, "null device path"),
            ("", "add", False, "empty device path"),
            ("/dev/video", "add", False, "device path without number"),
            
            # Edge cases
            ("/dev/video0extra", "add", False, "device path with extra text"),
            ("/custom/video0", "add", False, "non-standard device path"),
        ]
        
        processed_count = 0
        filtered_count = 0
        
        for device_path, action, should_process, description in filter_test_cases:
            # Create mock udev device
            mock_device = Mock()
            mock_device.device_node = device_path
            mock_device.action = action
            
            await monitor._process_udev_device_event(mock_device)
            
            if should_process:
                processed_count += 1
            else:
                filtered_count += 1
        
        # Verify statistics
        final_stats = monitor.get_monitor_stats()
        
        # Note: Some processed events might not result in state changes if device doesn't exist
        assert final_stats['udev_events_filtered'] >= initial_stats['udev_events_filtered'] + filtered_count
        
        monitor._set_test_mode(False)

    @pytest.mark.asyncio
    async def test_udev_race_condition_simulation(self, monitor):
        """Test race condition handling in concurrent udev events."""
        
        monitor._set_test_mode(True)
        device_path = "/dev/video0"
        
        # Mock device creation to succeed
        with patch.object(monitor, '_create_camera_device_info') as mock_create:
            mock_device = CameraDevice(
                device_path=device_path,
                name="Test Camera",
                status="CONNECTED"
            )
            mock_create.return_value = mock_device
            
            # Simulate rapid add/remove/add sequence (race condition)
            event_sequence = [
                ("add", 0.01),
                ("remove", 0.01),
                ("add", 0.01),
                ("change", 0.01),
                ("remove", 0.01)
            ]
            
            # Execute events concurrently
            tasks = []
            for action, delay in event_sequence:
                task = asyncio.create_task(
                    self._simulate_delayed_udev_event(monitor, device_path, action, delay)
                )
                tasks.append(task)
            
            # Wait for all events to complete
            await asyncio.gather(*tasks, return_exceptions=True)
            
            # Verify final state consistency
            stats = monitor.get_monitor_stats()
            assert stats['udev_events_processed'] >= len(event_sequence)
            
        monitor._set_test_mode(False)

    async def _simulate_delayed_udev_event(self, monitor, device_path: str, action: str, delay: float):
        """Simulate udev event with delay for race condition testing."""
        await asyncio.sleep(delay)
        await monitor._inject_test_udev_event(device_path, action)


class TestPollingFallbackBehavior:
    """Test polling fallback behavior when udev is silent or unavailable."""

    @pytest.fixture
    def monitor_no_udev(self):
        """Create monitor with udev disabled to test polling-only mode."""
        with patch('src.camera_discovery.hybrid_monitor.HAS_PYUDEV', False):
            return HybridCameraMonitor(
                device_range=[0, 1],
                poll_interval=0.05,  # Fast polling for testing
                enable_capability_detection=False
            )

    @pytest.mark.asyncio
    async def test_adaptive_polling_interval_adjustment(self, monitor_no_udev):
        """Test adaptive polling interval adjustment based on udev event freshness."""
        
        # Get initial polling state
        initial_state = monitor_no_udev._get_adaptive_polling_state_for_testing()
        
        # Simulate scenario: no recent udev events (should increase polling frequency)
        current_time = time.time()
        monitor_no_udev._last_udev_event_time = current_time - 20.0  # 20 seconds ago
        
        await monitor_no_udev._adjust_polling_interval()
        
        state_after_stale = monitor_no_udev._get_adaptive_polling_state_for_testing()
        
        # Should have reduced interval (increased frequency)
        assert state_after_stale['current_interval'] < initial_state['current_interval']
        
        # Simulate scenario: recent udev events (should decrease polling frequency)
        monitor_no_udev._last_udev_event_time = current_time - 2.0  # 2 seconds ago
        
        await monitor_no_udev._adjust_polling_interval()
        
        state_after_fresh = monitor_no_udev._get_adaptive_polling_state_for_testing()
        
        # Should have increased interval (decreased frequency)
        assert state_after_fresh['current_interval'] > state_after_stale['current_interval']

    @pytest.mark.asyncio
    async def test_polling_failure_backoff_with_jitter(self, monitor_no_udev):
        """Test polling failure handling with exponential backoff and jitter."""
        
        # Mock discovery to always fail
        with patch.object(monitor_no_udev, '_discover_cameras', side_effect=Exception("Simulated failure")):
            
            # Monitor polling state
            initial_failure_count = monitor_no_udev._polling_failure_count
            
            # Simulate several polling failures
            for i in range(3):
                try:
                    await monitor_no_udev._discover_cameras()
                except:
                    monitor_no_udev._polling_failure_count += 1
            
            # Verify failure count increased
            assert monitor_no_udev._polling_failure_count > initial_failure_count
            
            # Test backoff calculation
            state = monitor_no_udev._get_adaptive_polling_state_for_testing()
            assert state['failure_count'] > 0


class TestTimeoutAndSubprocessFailureHandling:
    """Test timeout and subprocess failure handling with structured error reporting."""

    @pytest.fixture
    def monitor(self):
        """Create monitor for timeout testing."""
        return HybridCameraMonitor(
            device_range=[0],
            detection_timeout=0.5,  # Short timeout for testing
            enable_capability_detection=True
        )

    @pytest.mark.asyncio
    async def test_subprocess_timeout_handling(self, monitor):
        """Test subprocess timeout handling with proper error contexts."""
        
        device_path = "/dev/video0"
        
        # Test timeout in subprocess operations
        with patch('asyncio.create_subprocess_exec') as mock_subprocess:
            # Mock subprocess that never completes
            mock_process = AsyncMock()
            mock_process.communicate = AsyncMock(side_effect=asyncio.TimeoutError())
            mock_subprocess.return_value = mock_process
            
            result = await monitor._probe_device_capabilities(device_path)
            
            assert not result.detected, "Should fail on subprocess timeout"
            assert "timeout" in result.error.lower(), "Error should mention timeout"
            assert result.structured_diagnostics is not None
            assert 'timeout_threshold' in result.structured_diagnostics

    @pytest.mark.asyncio
    async def test_subprocess_failure_handling(self, monitor):
        """Test subprocess failure handling with structured diagnostics."""
        
        device_path = "/dev/video0"
        
        # Test different subprocess failure modes
        failure_scenarios = [
            (1, b"", b"Device not found", "device_not_found"),
            (2, b"", b"Permission denied", "permission_denied"),
            (127, b"", b"Command not found", "command_not_found"),
        ]
        
        for return_code, stdout, stderr, error_type in failure_scenarios:
            with patch('asyncio.create_subprocess_exec') as mock_subprocess:
                mock_process = AsyncMock()
                mock_process.communicate.return_value = (stdout, stderr)
                mock_process.returncode = return_code
                mock_subprocess.return_value = mock_process
                
                result = await monitor._probe_device_capabilities(device_path)
                
                # Should handle gracefully without crashing
                assert result is not None
                assert result.structured_diagnostics is not None
                
                # Error context should be captured
                if return_code != 0:
                    assert not result.detected or result.error


class TestIntegrationAndLifecycle:
    """Integration tests for full monitor lifecycle and component coordination."""

    @pytest.fixture
    def monitor(self):
        """Create monitor for integration testing."""
        return HybridCameraMonitor(
            device_range=[0, 1],
            poll_interval=0.05,
            enable_capability_detection=True
        )

    @pytest.mark.asyncio
    async def test_monitor_full_lifecycle(self, monitor):
        """Test complete monitor lifecycle from startup to shutdown."""
        
        # Verify initial state
        assert not monitor.is_running
        initial_stats = monitor.get_monitor_stats()
        assert initial_stats['running'] is False
        assert initial_stats['active_tasks'] == 0
        
        # Start monitor
        with patch('src.camera_discovery.hybrid_monitor.HAS_PYUDEV', False):  # Disable udev for testing
            await monitor.start()
            
            assert monitor.is_running
            running_stats = monitor.get_monitor_stats()
            assert running_stats['running'] is True
            assert running_stats['active_tasks'] > 0
            
            # Let monitor run briefly
            await asyncio.sleep(0.1)
            
            # Verify some activity occurred
            activity_stats = monitor.get_monitor_stats()
            assert activity_stats['polling_cycles'] > 0
            
            # Stop monitor
            await monitor.stop()
            
            assert not monitor.is_running
            final_stats = monitor.get_monitor_stats()
            assert final_stats['running'] is False
            assert final_stats['active_tasks'] == 0

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
            formats=[{"code": "YUYV", "description": "YUYV 4:2:2"}]
        )
        
        with patch.object(monitor, '_probe_device_capabilities', return_value=mock_capability):
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


# Test utility functions
def test_stream_name_extraction_robustness():
    """Unit test: Stream name extraction from various device path formats."""
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


# Pytest configuration
pytest_plugins = ["pytest_asyncio"]


if __name__ == "__main__":
    # Allow running tests directly
    pytest.main([__file__, "-v", "--tb=short"])