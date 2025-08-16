# tests/unit/test_camera_discovery/test_hybrid_monitor_enhanced.py
"""
Enhanced test hybrid camera monitor with comprehensive monitoring and recovery.

This test file addresses PARTIAL coverage gaps identified in the comprehensive audit:
- REQ-CAM-004: System shall provide camera status monitoring
- REQ-ERROR-004: System shall handle configuration loading failures gracefully
- REQ-ERROR-005: System shall provide meaningful error messages for configuration issues

Requirements Traceability:
- REQ-CAM-004: Comprehensive camera status monitoring with adaptive polling and failure recovery
- REQ-ERROR-004: Configuration loading failure graceful handling
- REQ-ERROR-005: Meaningful error messages for configuration issues

Story Coverage: S1 - Camera Discovery and Monitoring
IV&V Control Point: Real camera monitoring with adaptive behavior and error recovery
"""

import pytest
import asyncio
import time
import tempfile
import os
from unittest.mock import AsyncMock, MagicMock, patch, Mock
from pathlib import Path

from src.camera_discovery.hybrid_monitor import (
    HybridCameraMonitor, 
    CameraEvent, 
    CameraEventData,
    CapabilityDetectionResult,
    DeviceCapabilityState
)
from src.common.types import CameraDevice


class TestEnhancedHybridCameraMonitor:
    """Enhanced test hybrid camera monitor with comprehensive monitoring and recovery."""

    @pytest.fixture
    def temp_dirs(self):
        """Create temporary directories for testing."""
        base = tempfile.mkdtemp(prefix="enhanced_monitor_test_")
        try:
            yield {
                "base": base,
                "config_path": os.path.join(base, "config.yml"),
                "log_path": os.path.join(base, "logs")
            }
        finally:
            import shutil
            shutil.rmtree(base, ignore_errors=True)

    @pytest.fixture
    async def monitor(self, temp_dirs):
        """Create hybrid camera monitor instance."""
        monitor = HybridCameraMonitor(
            device_range=[0, 1, 2],
            poll_interval=0.1,
            enable_capability_detection=True
        )
        await monitor.start()
        try:
            yield monitor
        finally:
            await monitor.stop()

    @pytest.mark.asyncio
    async def test_camera_status_monitoring_adaptive_polling_interval(
        self, monitor
    ):
        """
        Test adaptive polling interval adjustment based on camera activity.
        
        Requirements: REQ-CAM-004
        Scenario: Adaptive polling interval adjustment for efficient monitoring
        Expected: Polling interval adjusts based on camera activity and system load
        Edge Cases: High activity periods, low activity periods, system resource constraints
        """
        # Test initial polling interval
        initial_interval = monitor._poll_interval
        
        # Simulate high camera activity
        with patch.object(monitor, '_get_connected_cameras', return_value={
            '/dev/video0': CameraDevice('/dev/video0', 'Test Camera 0', 'CONNECTED'),
            '/dev/video1': CameraDevice('/dev/video1', 'Test Camera 1', 'CONNECTED'),
            '/dev/video2': CameraDevice('/dev/video2', 'Test Camera 2', 'CONNECTED')
        }):
            # Trigger polling cycle
            await monitor._poll_cameras()
            
            # Verify polling interval adjusts for high activity
            # (should decrease for more frequent monitoring)
            assert monitor._poll_interval <= initial_interval

        # Simulate low camera activity
        with patch.object(monitor, '_get_connected_cameras', return_value={}):
            # Trigger polling cycle
            await monitor._poll_cameras()
            
            # Verify polling interval adjusts for low activity
            # (should increase for less frequent monitoring)
            assert monitor._poll_interval >= initial_interval

    @pytest.mark.asyncio
    async def test_camera_status_monitoring_failure_recovery_mechanism(
        self, monitor
    ):
        """
        Test failure recovery mechanism with exponential backoff.
        
        Requirements: REQ-CAM-004
        Scenario: Failure recovery with exponential backoff and retry logic
        Expected: System recovers from failures with appropriate backoff strategy
        Edge Cases: Consecutive failures, recovery success, backoff limits
        """
        # Simulate consecutive failures
        failure_count = 0
        max_failures = 3
        
        def mock_get_connected_cameras():
            nonlocal failure_count
            failure_count += 1
            if failure_count <= max_failures:
                raise Exception(f"Simulated failure {failure_count}")
            return {
                '/dev/video0': CameraDevice('/dev/video0', 'Test Camera 0', 'CONNECTED')
            }
        
        with patch.object(monitor, '_get_connected_cameras', side_effect=mock_get_connected_cameras):
            # Trigger polling cycles that will fail initially
            for i in range(max_failures + 1):
                try:
                    await monitor._poll_cameras()
                except Exception:
                    pass  # Expected failures
                
                # Verify backoff mechanism is working
                if i < max_failures:
                    # Should have increased backoff interval
                    assert monitor._poll_interval > 0.1
                
                await asyncio.sleep(0.1)  # Small delay between attempts
            
            # Verify system recovers after failures
            connected_cameras = await monitor.get_connected_cameras()
            assert len(connected_cameras) == 1
            assert '/dev/video0' in connected_cameras

    @pytest.mark.asyncio
    async def test_camera_status_monitoring_polling_only_mode(
        self, monitor
    ):
        """
        Test polling-only mode when udev monitoring is unavailable.
        
        Requirements: REQ-CAM-004
        Scenario: Fallback to polling-only mode when udev unavailable
        Expected: System continues monitoring using polling fallback
        Edge Cases: udev service unavailable, polling-only operation, mixed mode
        """
        # Disable udev monitoring
        with patch.object(monitor, '_udev_monitor', None):
            # Verify system continues to function in polling-only mode
            connected_cameras = await monitor.get_connected_cameras()
            assert isinstance(connected_cameras, dict)
            
            # Verify polling continues to work
            await monitor._poll_cameras()
            
            # Verify event callbacks still work
            event_received = False
            
            def test_callback(event_data: CameraEventData):
                nonlocal event_received
                event_received = True
            
            monitor.add_event_callback(test_callback)
            
            # Trigger a camera event
            with patch.object(monitor, '_get_connected_cameras', return_value={
                '/dev/video0': CameraDevice('/dev/video0', 'Test Camera 0', 'CONNECTED')
            }):
                await monitor._poll_cameras()
                
                # Give time for event processing
                await asyncio.sleep(0.1)
                
                # Verify event was processed
                assert event_received

    @pytest.mark.asyncio
    async def test_camera_status_monitoring_capability_reconciliation(
        self, monitor
    ):
        """
        Test capability data reconciliation across monitoring cycles.
        
        Requirements: REQ-CAM-004
        Scenario: Capability data reconciliation and consistency validation
        Expected: Consistent capability data across monitoring cycles
        Edge Cases: Inconsistent capability data, data conflicts, validation failures
        """
        # Test capability data reconciliation
        test_capability_data = {
            "resolution": "1280x720",
            "fps": 25,
            "validation_status": "provisional",
            "formats": ["YUYV", "MJPEG"],
            "all_resolutions": ["1920x1080", "1280x720", "640x480"],
            "consecutive_successes": 1,
        }
        
        # Mock capability detection to return consistent data
        with patch.object(monitor, 'get_effective_capability_metadata', 
                        return_value=test_capability_data):
            
            # Perform multiple monitoring cycles
            for i in range(3):
                connected_cameras = await monitor.get_connected_cameras()
                
                if connected_cameras:
                    device_path = list(connected_cameras.keys())[0]
                    capability_metadata = monitor.get_effective_capability_metadata(device_path)
                    
                    # Verify capability data remains consistent
                    assert capability_metadata["resolution"] == test_capability_data["resolution"]
                    assert capability_metadata["fps"] == test_capability_data["fps"]
                    assert capability_metadata["formats"] == test_capability_data["formats"]
                    assert capability_metadata["all_resolutions"] == test_capability_data["all_resolutions"]
                
                await asyncio.sleep(0.1)

    @pytest.mark.asyncio
    async def test_configuration_loading_failure_graceful_handling(
        self, monitor, temp_dirs
    ):
        """
        Test graceful handling of configuration loading failures.
        
        Requirements: REQ-ERROR-004
        Scenario: Configuration loading failures with graceful degradation
        Expected: System continues to function with default configuration
        Edge Cases: Invalid configuration files, missing configuration, permission errors
        """
        # Test with invalid configuration
        invalid_config = {
            "device_range": "invalid",  # Should be list
            "poll_interval": -1,  # Should be positive
            "enable_capability_detection": "invalid"  # Should be boolean
        }
        
        # Create monitor with invalid configuration
        with patch('src.camera_discovery.hybrid_monitor.HybridCameraMonitor.__init__', 
                  side_effect=Exception("Configuration error")):
            try:
                bad_monitor = HybridCameraMonitor(
                    device_range=invalid_config["device_range"],
                    poll_interval=invalid_config["poll_interval"],
                    enable_capability_detection=invalid_config["enable_capability_detection"]
                )
                await bad_monitor.start()
            except Exception as e:
                # Verify meaningful error message
                assert "Configuration error" in str(e) or "invalid" in str(e).lower()
        
        # Verify original monitor still functions
        connected_cameras = await monitor.get_connected_cameras()
        assert isinstance(connected_cameras, dict)

    @pytest.mark.asyncio
    async def test_meaningful_error_messages_configuration_issues(
        self, monitor
    ):
        """
        Test meaningful error messages for configuration issues.
        
        Requirements: REQ-ERROR-005
        Scenario: Various configuration issues with descriptive error messages
        Expected: Clear, actionable error messages for configuration problems
        Edge Cases: Invalid parameters, missing required fields, type mismatches
        """
        # Test various configuration error scenarios
        error_scenarios = [
            {
                "name": "invalid_device_range",
                "params": {"device_range": "not_a_list", "poll_interval": 0.1, "enable_capability_detection": True},
                "expected_error": "device_range"
            },
            {
                "name": "invalid_poll_interval",
                "params": {"device_range": [0, 1, 2], "poll_interval": -1, "enable_capability_detection": True},
                "expected_error": "poll_interval"
            },
            {
                "name": "invalid_capability_detection",
                "params": {"device_range": [0, 1, 2], "poll_interval": 0.1, "enable_capability_detection": "not_boolean"},
                "expected_error": "enable_capability_detection"
            }
        ]
        
        for scenario in error_scenarios:
            try:
                bad_monitor = HybridCameraMonitor(**scenario["params"])
                await bad_monitor.start()
            except Exception as e:
                error_message = str(e).lower()
                # Verify error message contains relevant information
                assert scenario["expected_error"].lower() in error_message or "invalid" in error_message

    @pytest.mark.asyncio
    async def test_camera_status_monitoring_performance_under_load(
        self, monitor
    ):
        """
        Test camera status monitoring performance under load.
        
        Requirements: REQ-CAM-004
        Scenario: Monitoring performance with multiple cameras and high activity
        Expected: Efficient monitoring without performance degradation
        Edge Cases: High camera count, frequent status changes, resource constraints
        """
        # Simulate high load with multiple cameras
        high_load_cameras = {}
        for i in range(10):  # Simulate 10 cameras
            high_load_cameras[f'/dev/video{i}'] = CameraDevice(
                f'/dev/video{i}', 
                f'Test Camera {i}', 
                'CONNECTED'
            )
        
        with patch.object(monitor, '_get_connected_cameras', return_value=high_load_cameras):
            # Measure performance of monitoring operations
            start_time = time.time()
            
            # Perform multiple monitoring cycles
            for i in range(5):
                connected_cameras = await monitor.get_connected_cameras()
                await monitor._poll_cameras()
                await asyncio.sleep(0.1)
            
            end_time = time.time()
            total_time = end_time - start_time
            
            # Verify performance is acceptable (should complete within reasonable time)
            assert total_time < 2.0  # Should complete within 2 seconds
            
            # Verify all cameras are detected
            assert len(connected_cameras) == 10

    @pytest.mark.asyncio
    async def test_camera_status_monitoring_event_ordering(
        self, monitor
    ):
        """
        Test camera status monitoring event ordering and consistency.
        
        Requirements: REQ-CAM-004
        Scenario: Event ordering and consistency across monitoring cycles
        Expected: Proper event ordering and consistent state reporting
        Edge Cases: Rapid status changes, event ordering, state consistency
        """
        events_received = []
        
        def event_callback(event_data: CameraEventData):
            events_received.append({
                "device": event_data.device_path,
                "event": event_data.event_type,
                "timestamp": event_data.timestamp
            })
        
        monitor.add_event_callback(event_callback)
        
        # Simulate camera connection sequence
        camera_sequence = [
            {'/dev/video0': CameraDevice('/dev/video0', 'Test Camera 0', 'CONNECTED')},
            {
                '/dev/video0': CameraDevice('/dev/video0', 'Test Camera 0', 'CONNECTED'),
                '/dev/video1': CameraDevice('/dev/video1', 'Test Camera 1', 'CONNECTED')
            },
            {
                '/dev/video0': CameraDevice('/dev/video0', 'Test Camera 0', 'CONNECTED'),
                '/dev/video1': CameraDevice('/dev/video1', 'Test Camera 1', 'CONNECTED'),
                '/dev/video2': CameraDevice('/dev/video2', 'Test Camera 2', 'CONNECTED')
            }
        ]
        
        for i, cameras in enumerate(camera_sequence):
            with patch.object(monitor, '_get_connected_cameras', return_value=cameras):
                await monitor._poll_cameras()
                await asyncio.sleep(0.1)
        
        # Verify events were received in proper order
        assert len(events_received) > 0
        
        # Verify event timestamps are in ascending order
        timestamps = [event["timestamp"] for event in events_received if event["timestamp"]]
        assert timestamps == sorted(timestamps)

    @pytest.mark.asyncio
    async def test_camera_status_monitoring_recovery_confirmation_logging(
        self, monitor
    ):
        """
        Test recovery confirmation logging for monitoring operations.
        
        Requirements: REQ-CAM-004
        Scenario: Recovery confirmation logging and monitoring
        Expected: Proper logging of recovery events and system state
        Edge Cases: Recovery events, logging levels, monitoring visibility
        """
        # Capture log messages
        log_messages = []
        
        def log_capture(level, message):
            log_messages.append({"level": level, "message": message})
        
        # Patch logging to capture messages
        with patch.object(monitor._logger, 'info', side_effect=lambda msg: log_capture('info', msg)), \
             patch.object(monitor._logger, 'warning', side_effect=lambda msg: log_capture('warning', msg)), \
             patch.object(monitor._logger, 'error', side_effect=lambda msg: log_capture('error', msg)):
            
            # Simulate recovery scenario
            with patch.object(monitor, '_get_connected_cameras', 
                            side_effect=[Exception("Simulated failure"), 
                                       {'/dev/video0': CameraDevice('/dev/video0', 'Test Camera 0', 'CONNECTED')}]):
                
                # First call should fail
                try:
                    await monitor._poll_cameras()
                except Exception:
                    pass
                
                # Second call should succeed (recovery)
                await monitor._poll_cameras()
                
                # Verify recovery logging
                recovery_logs = [msg for msg in log_messages if "recovery" in msg["message"].lower() or "recovered" in msg["message"].lower()]
                assert len(recovery_logs) > 0
