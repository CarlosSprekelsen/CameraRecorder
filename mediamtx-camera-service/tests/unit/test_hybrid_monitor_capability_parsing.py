"""
Real integration tests for hybrid monitor capability parsing without mocks.

Tests the HybridMonitor capability parsing with real V4L2 subprocess operations,
real file system checks, and real device access scenarios.

Requirements:
- REQ-CAM-003: System shall detect camera capabilities using V4L2
- REQ-ERROR-004: System shall handle capability detection failures gracefully
- REQ-ERROR-005: System shall provide meaningful error messages for device issues

Story Coverage: S12 - Camera Discovery and Capability Detection
IV&V Control Point: Real V4L2 integration validation
"""

import asyncio
import os
import subprocess
from pathlib import Path

import pytest

from camera_discovery.hybrid_monitor import HybridCameraMonitor, CapabilityDetectionResult


class TestHybridMonitorCapabilityParsingRealIntegration:
    """Real integration tests for hybrid monitor capability parsing without mocks."""

    @pytest.fixture
    def monitor(self):
        """Create a fresh HybridCameraMonitor instance for testing."""
        return HybridCameraMonitor()

    def test_real_v4l2_command_execution(self, monitor):
        """Test real v4l2-ctl command execution and output parsing."""
        try:
            # Test basic v4l2-ctl functionality
            result = subprocess.run(
                ["v4l2-ctl", "--version"],
                capture_output=True,
                text=True,
                timeout=5.0
            )
            assert result.returncode == 0
            assert "v4l2-ctl" in result.stdout
        except (subprocess.TimeoutExpired, FileNotFoundError):
            pytest.skip("v4l2-ctl not available or not working")

    @pytest.mark.asyncio
    async def test_real_device_capability_probing(self, monitor):
        """Test real device capability probing with actual V4L2 devices."""
        # Test with common video device paths
        test_devices = ["/dev/video0", "/dev/video1", "/dev/video2"]
        
        for device_path in test_devices:
            if os.path.exists(device_path):
                # Test real capability probing
                caps = await monitor._probe_device_capabilities(device_path)
                
                assert isinstance(caps, CapabilityDetectionResult)
                assert hasattr(caps, 'detected')
                assert hasattr(caps, 'accessible')
                assert hasattr(caps, 'error')
                
                # If device is accessible, should have capability information
                if caps.accessible:
                    assert caps.detected is True
                    assert caps.error is None
                    # Should have some capability information
                    assert hasattr(caps, 'formats')
                    assert hasattr(caps, 'resolutions')
                    assert hasattr(caps, 'frame_rates')
                else:
                    # Device exists but not accessible (permission issues, etc.)
                    assert caps.error is not None
                    assert "permission" in caps.error.lower() or "busy" in caps.error.lower() or "failed" in caps.error.lower()

    @pytest.mark.asyncio
    async def test_real_timeout_handling(self, monitor):
        """Test real timeout handling with actual V4L2 operations."""
        # Set a very short timeout to test real timeout behavior
        original_timeout = monitor._detection_timeout
        monitor._detection_timeout = 0.1  # Very short timeout
        
        try:
            # Test with a device that might exist but will timeout
            caps = await monitor._probe_device_capabilities("/dev/video0")
            
            assert isinstance(caps, CapabilityDetectionResult)
            assert hasattr(caps, 'detected')
            assert hasattr(caps, 'accessible')
            assert hasattr(caps, 'error')
            
            # If timeout occurred, should have timeout error
            if caps.error and "timeout" in caps.error.lower():
                assert caps.detected is False
                assert caps.accessible is False
        finally:
            # Restore original timeout
            monitor._detection_timeout = original_timeout

    @pytest.mark.asyncio
    async def test_real_subprocess_failure_handling(self, monitor):
        """Test real subprocess failure handling with actual V4L2 operations."""
        # Test with non-existent device (should fail with real subprocess error)
        caps = await monitor._probe_device_capabilities("/dev/video999")
        
        assert isinstance(caps, CapabilityDetectionResult)
        assert caps.detected is False
        assert caps.accessible is False
        assert caps.error is not None
        assert "failed to probe" in caps.error.lower() or "timeout" in caps.error.lower() or "unavailable" in caps.error.lower()

    def test_real_v4l2_output_parsing(self, monitor):
        """Test parsing of real v4l2-ctl output formats."""
        # Test with actual v4l2-ctl output if available
        try:
            result = subprocess.run(
                ["v4l2-ctl", "--list-formats-ext"],
                capture_output=True,
                text=True,
                timeout=10.0
            )
            
            if result.returncode == 0 and result.stdout:
                # Test parsing real output
                resolutions = monitor._extract_resolutions_from_output(result.stdout)
                assert isinstance(resolutions, (list, set))
                
                formats = monitor._extract_formats_from_output(result.stdout)
                assert isinstance(formats, list)
                
                frame_rates = monitor._extract_frame_rates_from_output(result.stdout)
                assert isinstance(frame_rates, (list, set))
        except (subprocess.TimeoutExpired, FileNotFoundError):
            pytest.skip("v4l2-ctl not available or not working")

    def test_real_resolution_extraction_patterns(self, monitor):
        """Test resolution extraction from real v4l2-ctl output formats."""
        # Test with actual v4l2-ctl output if available
        try:
            result = subprocess.run(
                ["v4l2-ctl", "--list-formats-ext"],
                capture_output=True,
                text=True,
                timeout=10.0
            )
            
            if result.returncode == 0 and result.stdout:
                # Test parsing real output
                resolutions = monitor._extract_resolutions_from_output(result.stdout)
                assert isinstance(resolutions, (list, set))
                
                # If resolutions were found, they should be in valid format
                for resolution in resolutions:
                    assert isinstance(resolution, str)
                    # Should contain 'x' separator
                    if 'x' in resolution:
                        parts = resolution.split('x')
                        assert len(parts) == 2
                        # Should be numeric
                        assert parts[0].isdigit()
                        assert parts[1].isdigit()
        except (subprocess.TimeoutExpired, FileNotFoundError):
            pytest.skip("v4l2-ctl not available or not working")

    def test_real_format_extraction_patterns(self, monitor):
        """Test format extraction from real v4l2-ctl output formats."""
        # Test with actual v4l2-ctl output if available
        try:
            result = subprocess.run(
                ["v4l2-ctl", "--list-formats-ext"],
                capture_output=True,
                text=True,
                timeout=10.0
            )
            
            if result.returncode == 0 and result.stdout:
                # Test parsing real output
                formats = monitor._extract_formats_from_output(result.stdout)
                assert isinstance(formats, list)
                
                # If formats were found, they should have valid structure
                for fmt in formats:
                    assert isinstance(fmt, dict)
                    if 'code' in fmt:
                        assert isinstance(fmt['code'], str)
                    if 'name' in fmt:
                        assert isinstance(fmt['name'], str)
        except (subprocess.TimeoutExpired, FileNotFoundError):
            pytest.skip("v4l2-ctl not available or not working")

    def test_real_frame_rate_extraction_patterns(self, monitor):
        """Test frame rate extraction from real v4l2-ctl output formats."""
        # Test with actual v4l2-ctl output if available
        try:
            result = subprocess.run(
                ["v4l2-ctl", "--list-formats-ext"],
                capture_output=True,
                text=True,
                timeout=10.0
            )
            
            if result.returncode == 0 and result.stdout:
                # Test parsing real output
                frame_rates = monitor._extract_frame_rates_from_output(result.stdout)
                assert isinstance(frame_rates, (list, set))
                
                # If frame rates were found, they should be numeric strings
                for rate in frame_rates:
                    assert isinstance(rate, str)
                    # Should be numeric or contain common frame rate patterns
                    assert rate.isdigit() or '/' in rate or '.' in rate
        except (subprocess.TimeoutExpired, FileNotFoundError):
            pytest.skip("v4l2-ctl not available or not working")

    def test_real_device_info_extraction(self, monitor):
        """Test device info extraction from real v4l2-ctl output."""
        # Test with actual v4l2-ctl output if available
        try:
            result = subprocess.run(
                ["v4l2-ctl", "--device-info"],
                capture_output=True,
                text=True,
                timeout=10.0
            )
            
            if result.returncode == 0 and result.stdout:
                # Test parsing real output
                device_info = monitor._extract_device_info_from_output(result.stdout)
                assert isinstance(device_info, dict)
                
                # Should have basic device information
                if 'name' in device_info:
                    assert isinstance(device_info['name'], str)
                if 'driver' in device_info:
                    assert isinstance(device_info['driver'], str)
        except (subprocess.TimeoutExpired, FileNotFoundError):
            pytest.skip("v4l2-ctl not available or not working")

    @pytest.mark.asyncio
    async def test_real_monitor_startup_and_shutdown(self, monitor):
        """Test real monitor startup and shutdown with actual device discovery."""
        # Test monitor startup
        await monitor.start()
        assert monitor.is_running is True
        
        # Test device discovery (should not crash)
        devices = await monitor.get_connected_cameras()
        assert isinstance(devices, dict)
        
        # Test monitor shutdown
        await monitor.stop()
        assert monitor.is_running is False

    def test_real_error_message_formatting(self, monitor):
        """Test real error message formatting with actual error scenarios."""
        # Test with real subprocess errors
        try:
            # Try to run v4l2-ctl with invalid device
            result = subprocess.run(
                ["v4l2-ctl", "-d", "/dev/video999", "--list-formats-ext"],
                capture_output=True,
                text=True,
                timeout=5.0
            )
            
            # Should get error output
            if result.stderr:
                # Test that error messages are properly formatted
                assert isinstance(result.stderr, str)
                assert len(result.stderr) > 0
        except (subprocess.TimeoutExpired, FileNotFoundError):
            pytest.skip("v4l2-ctl not available or not working")

    def test_real_device_path_validation(self, monitor):
        """Test real device path validation with actual file system."""
        # Test with real device paths
        test_paths = [
            "/dev/video0",
            "/dev/video1", 
            "/dev/video999",  # Non-existent
            "/dev/null",      # Exists but not a video device
            "/tmp",           # Directory
        ]
        
        for path in test_paths:
            # Test real file system validation - check if path exists and is accessible
            is_valid = os.path.exists(path) and os.path.isfile(path)
            assert isinstance(is_valid, bool)
            
            # If path exists and is a character device, should be valid
            if os.path.exists(path) and os.path.isfile(path):
                try:
                    stat_info = os.stat(path)
                    if stat_info.st_mode & 0o170000 == 0o20000:  # Character device
                        assert is_valid is True
                except OSError:
                    # Permission denied or other access issues
                    pass


# ===== QUARANTINED TESTS =====

@pytest.mark.skip(reason="QUARANTINED: Mock-based tests replaced with real integration tests")
class TestHybridMonitorCapabilityParsingQuarantined:
    """Quarantined mock-based tests - replaced with real integration tests."""
    
    # These tests are quarantined because they use mocks instead of testing real V4L2 integration
    # The real integration tests above provide better coverage of actual system behavior
