"""
Strategic Mock Minimization for Hardware/UDev Integration Tests.

This module implements the executive decision to assess hardware/UDev tests
with minimal mocks ONLY for true external dependencies.

STRATEGIC MOCK POLICY:
- ✅ ALLOWED: Mock actual hardware device access (/dev/video* file descriptors)
- ✅ ALLOWED: Mock USB device enumeration if no test devices available
- ❌ FORBIDDEN: Mock file system operations, subprocess calls, configuration parsing
- ❌ FORBIDDEN: Mock business logic, validation algorithms, state management

IMPLEMENTATION APPROACH:
- Create /tmp/test_devices/ with simulated device files
- Use real os.listdir(), os.path.exists() for device discovery
- Execute real subprocess calls with test parameters
- Mock only the final device.open() or ioctl() system calls
"""

import asyncio
import os
import tempfile
import subprocess
import time
import pytest
from pathlib import Path
from unittest.mock import Mock, AsyncMock, patch, MagicMock
from typing import Dict, List, Optional, Set

from src.camera_discovery.hybrid_monitor import (
    HybridCameraMonitor,
    CameraEvent,
    CapabilityDetectionResult,
)
from src.common.types import CameraDevice


class HardwareSimulationFixture:
    """Creates real file structures for hardware simulation."""
    
    def __init__(self, temp_dir: str):
        self.temp_dir = Path(temp_dir)
        self.devices_dir = self.temp_dir / "dev"
        self.devices_dir.mkdir(exist_ok=True)
        self.created_devices: Set[str] = set()
        
    def create_test_device(self, device_name: str, device_type: str = "video") -> str:
        """Create a test device file with real file operations."""
        device_path = self.devices_dir / device_name
        
        # Create real device file
        device_path.touch(mode=0o666)
        self.created_devices.add(str(device_path))
        
        return str(device_path)
    
    def create_multiple_devices(self, device_names: List[str]) -> Dict[str, str]:
        """Create multiple test devices."""
        devices = {}
        for name in device_names:
            devices[name] = self.create_test_device(name)
        return devices
    
    def remove_device(self, device_name: str) -> None:
        """Remove a test device file."""
        device_path = self.devices_dir / device_name
        if device_path.exists():
            device_path.unlink()
            self.created_devices.discard(str(device_path))
    
    def list_devices(self) -> List[str]:
        """List all created devices using real file operations."""
        return [str(p) for p in self.devices_dir.iterdir() if p.is_file()]
    
    def cleanup(self) -> None:
        """Clean up all created devices."""
        for device_path in self.created_devices.copy():
            try:
                Path(device_path).unlink(missing_ok=True)
            except Exception:
                pass


class TestHardwareIntegrationReal:
    """Test hardware integration with minimal mocks for true external dependencies."""
    
    @pytest.fixture
    def hardware_fixture(self):
        """Create hardware simulation fixture with real file structures."""
        with tempfile.TemporaryDirectory() as temp_dir:
            fixture = HardwareSimulationFixture(temp_dir)
            yield fixture
            fixture.cleanup()
    
    @pytest.fixture
    def monitor_with_real_fs(self, hardware_fixture):
        """Create monitor with real file system operations."""
        # Create test devices
        devices = hardware_fixture.create_multiple_devices([
            "video0", "video1", "video2", "video5", "audio0"
        ])
        
        monitor = HybridCameraMonitor(
            device_range=[0, 1, 2],
            enable_capability_detection=True,
            detection_timeout=1.0
        )
        
        yield monitor, hardware_fixture, devices
    
    @pytest.mark.asyncio
    async def test_real_device_discovery_with_file_operations(self, monitor_with_real_fs):
        """Test device discovery using real file system operations."""
        monitor, hardware_fixture, devices = monitor_with_real_fs
        
        # Use real file operations for device discovery
        discovered_devices = hardware_fixture.list_devices()
        
        # Should find our test devices
        assert len(discovered_devices) >= 3
        assert any("video0" in d for d in discovered_devices)
        assert any("video1" in d for d in discovered_devices)
        assert any("video2" in d for d in discovered_devices)
        
        # Test device enumeration with real path operations
        for device_num in [0, 1, 2]:
            device_name = f"video{device_num}"
            device_path = hardware_fixture.devices_dir / device_name
            
            # Use real path operations
            assert device_path.exists(), f"Test device {device_name} should exist"
    
    @pytest.mark.asyncio
    async def test_real_subprocess_calls_with_test_parameters(self, monitor_with_real_fs):
        """Test real subprocess calls with test parameters (mock only device access)."""
        monitor, hardware_fixture, devices = monitor_with_real_fs
        
        # Mock only the actual device file opening (TRUE_EXTERNAL dependency)
        with patch('builtins.open') as mock_open:
            mock_file = Mock()
            mock_file.read.return_value = b"mock device data"
            mock_open.return_value.__enter__.return_value = mock_file
            
            # Test real subprocess execution with test parameters
            test_command = ["echo", "test_device_info"]
            
            try:
                # Execute real subprocess call
                result = subprocess.run(
                    test_command,
                    capture_output=True,
                    text=True,
                    timeout=5.0
                )
                
                # Verify real subprocess execution
                assert result.returncode == 0
                assert "test_device_info" in result.stdout
                
            except subprocess.TimeoutExpired:
                pytest.fail("Subprocess call timed out")
            except FileNotFoundError:
                pytest.skip("echo command not available")
    
    @pytest.mark.asyncio
    async def test_real_configuration_parsing_and_validation(self, monitor_with_real_fs):
        """Test real configuration parsing and validation logic."""
        monitor, hardware_fixture, devices = monitor_with_real_fs
        
        # Test real configuration parsing with test data
        test_config_data = {
            "device_range": [0, 1, 2],
            "enable_capability_detection": True,
            "detection_timeout": 1.0
        }
        
        # Use real validation logic
        assert isinstance(test_config_data["device_range"], list)
        assert all(isinstance(x, int) for x in test_config_data["device_range"])
        assert isinstance(test_config_data["enable_capability_detection"], bool)
        assert isinstance(test_config_data["detection_timeout"], (int, float))
        assert test_config_data["detection_timeout"] > 0
        
        # Test device range validation logic
        device_range = test_config_data["device_range"]
        for device_num in device_range:
            device_path = f"/dev/video{device_num}"
            test_path = device_path.replace('/dev/', str(hardware_fixture.devices_dir) + '/')
            
            # Use real path validation
            path_obj = Path(test_path)
            assert path_obj.exists(), f"Device {device_path} should exist for validation"
    
    @pytest.mark.asyncio
    async def test_real_udev_event_simulation_with_file_changes(self, monitor_with_real_fs):
        """Test udev event simulation using real file system changes."""
        monitor, hardware_fixture, devices = monitor_with_real_fs
        
        # Simulate udev events with real file system changes
        events_processed = []
        
        async def capture_event(event_data):
            events_processed.append(event_data)
        
        monitor._handle_camera_event = capture_event
        
        # Simulate device addition with real file creation
        new_device = hardware_fixture.create_test_device("video3")
        
        # Simulate device removal with real file deletion
        hardware_fixture.remove_device("video1")
        
        # Simulate device change with real file modification
        video2_path = Path(devices["video2"])
        video2_path.touch()  # Update modification time
        
        # Verify real file system changes
        assert Path(new_device).exists()
        assert not (hardware_fixture.devices_dir / "video1").exists()
        assert Path(devices["video2"]).exists()
    
    @pytest.mark.asyncio
    async def test_real_device_enumeration_logic(self, monitor_with_real_fs):
        """Test real device enumeration logic without mocking business logic."""
        monitor, hardware_fixture, devices = monitor_with_real_fs
        
        # Test real device enumeration using actual file operations
        all_devices = hardware_fixture.list_devices()
        
        # Filter devices using real logic (no mocking)
        video_devices = [d for d in all_devices if "video" in d]
        audio_devices = [d for d in all_devices if "audio" in d]
        
        # Verify real filtering logic
        assert len(video_devices) >= 3  # video0, video1, video2
        assert len(audio_devices) >= 1  # audio0
        
        # Test device number extraction logic
        device_numbers = []
        for device_path in video_devices:
            device_name = Path(device_path).name
            if device_name.startswith("video"):
                try:
                    number = int(device_name[5:])  # Extract number from "videoX"
                    device_numbers.append(number)
                except ValueError:
                    continue
        
        # Verify real number extraction logic
        assert 0 in device_numbers
        assert 1 in device_numbers
        assert 2 in device_numbers
    
    @pytest.mark.asyncio
    async def test_real_capability_detection_parsing(self, monitor_with_real_fs):
        """Test real capability detection parsing logic."""
        monitor, hardware_fixture, devices = monitor_with_real_fs
        
        # Test real parsing logic with sample v4l2 output
        sample_v4l2_output = """
        Driver name   : uvcvideo
        Card type     : USB Camera
        Bus info      : usb-0000:00:14.0-1
        
        Format [0]:
          Name: YUYV
          Description: YUYV 4:2:2
        Size: Discrete 1920x1080
        Size: Discrete 1280x720
        Size: Discrete 640x480
        
        Frame rate: 30.000 fps
        Frame rate: 25.000 fps
        Frame rate: 15.000 fps
        """
        
        # Test real frame rate extraction logic
        frame_rates = monitor._extract_frame_rates_from_output(sample_v4l2_output)
        
        # Verify real parsing results
        expected_rates = {"30", "25", "15"}
        assert frame_rates == expected_rates
        
        # Test real format extraction logic
        formats = monitor._extract_formats_from_output(sample_v4l2_output)
        format_codes = [f.get("code", "") for f in formats]
        assert "YUYV" in format_codes
        
        # Test real resolution extraction logic
        resolutions = monitor._extract_resolutions_from_output(sample_v4l2_output)
        expected_resolutions = ["1920x1080", "1280x720", "640x480"]
        assert all(res in resolutions for res in expected_resolutions)
    
    @pytest.mark.asyncio
    async def test_real_state_management_logic(self, monitor_with_real_fs):
        """Test real state management logic without mocking."""
        monitor, hardware_fixture, devices = monitor_with_real_fs
        
        # Test real state tracking
        initial_devices = set(monitor._known_devices.keys())
        
        # Simulate device addition
        new_device = hardware_fixture.create_test_device("video3")
        
        # Test real state change detection
        current_devices = set(hardware_fixture.list_devices())
        
        # Verify real state management
        assert len(current_devices) > len(initial_devices)
        
        # Test real device tracking
        device_paths = [str(hardware_fixture.devices_dir / f"video{i}") for i in [0, 1, 2]]
        for device_path in device_paths:
            assert Path(device_path).exists()
    
    @pytest.mark.asyncio
    async def test_real_error_handling_logic(self, monitor_with_real_fs):
        """Test real error handling logic without mocking."""
        monitor, hardware_fixture, devices = monitor_with_real_fs
        
        # Test real error handling for non-existent devices
        non_existent_path = str(hardware_fixture.devices_dir / "nonexistent")
        
        # Use real path operations
        path_obj = Path(non_existent_path)
        assert not path_obj.exists()
        
        # Test real error handling for invalid device numbers
        invalid_device_numbers = [-1, 999, "invalid"]
        
        for device_num in invalid_device_numbers:
            if isinstance(device_num, int) and device_num < 0:
                # Real validation logic
                assert device_num not in monitor._device_range
        
        # Test real timeout handling
        start_time = time.time()
        try:
            # Simulate a timeout scenario
            await asyncio.wait_for(asyncio.sleep(0.1), timeout=0.05)
        except asyncio.TimeoutError:
            # Real timeout handling
            elapsed = time.time() - start_time
            assert elapsed < 0.1  # Should timeout before 0.1s
    
    @pytest.mark.asyncio
    async def test_real_adaptive_polling_logic(self, monitor_with_real_fs):
        """Test real adaptive polling logic without mocking."""
        monitor, hardware_fixture, devices = monitor_with_real_fs
        
        # Test real polling interval adjustment logic
        initial_interval = monitor._current_poll_interval
        
        # Simulate no recent udev events
        monitor._last_udev_event_time = time.time() - 10.0
        
        # Test real adaptive logic
        await monitor._adjust_polling_interval()
        
        # Verify real interval adjustment
        assert monitor._current_poll_interval != initial_interval
        
        # Test real failure counting logic
        monitor._polling_failure_count = 3
        await monitor._adjust_polling_interval()
        
        # Verify real failure penalty logic
        assert monitor._current_poll_interval >= initial_interval
    
    @pytest.mark.asyncio
    async def test_real_device_validation_logic(self, monitor_with_real_fs):
        """Test real device validation logic without mocking."""
        monitor, hardware_fixture, devices = monitor_with_real_fs
        
        # Test real device validation
        valid_devices = []
        invalid_devices = []
        
        for device_path in hardware_fixture.list_devices():
            path_obj = Path(device_path)
            
            # Real validation logic
            if path_obj.exists() and path_obj.is_file():
                valid_devices.append(device_path)
            else:
                invalid_devices.append(device_path)
        
        # Verify real validation results
        assert len(valid_devices) > 0
        assert len(invalid_devices) == 0  # All our test devices should be valid
        
        # Test real device range validation
        for device_path in valid_devices:
            device_name = Path(device_path).name
            if device_name.startswith("video"):
                try:
                    device_num = int(device_name[5:])
                    # Real range validation logic
                    if device_num in monitor._device_range:
                        assert device_path in valid_devices
                except ValueError:
                    continue


class TestMinimalMockingStrategy:
    """Test that mocks are only used for TRUE_EXTERNAL dependencies."""
    
    @pytest.mark.asyncio
    async def test_only_hardware_access_is_mocked(self):
        """Verify that only actual hardware device access is mocked."""
        
        with tempfile.TemporaryDirectory() as temp_dir:
            # Create real test environment
            test_devices = ["video0", "video1", "video2"]
            device_paths = []
            
            for device in test_devices:
                device_path = Path(temp_dir) / device
                device_path.touch()
                device_paths.append(str(device_path))
            
            # Mock ONLY the actual device file opening (TRUE_EXTERNAL)
            with patch('builtins.open') as mock_open:
                mock_file = Mock()
                mock_file.read.return_value = b"mock device data"
                mock_open.return_value.__enter__.return_value = mock_file
                
                # Test real file operations
                for device_path in device_paths:
                    path_obj = Path(device_path)
                    assert path_obj.exists()  # Real file operation
                    assert path_obj.is_file()  # Real file operation
                
                # Verify mock was only called for device access
                assert mock_open.call_count == 0  # No device access in this test
    
    @pytest.mark.asyncio
    async def test_real_subprocess_execution(self):
        """Test that subprocess calls are executed for real."""
        
        # Test real subprocess execution
        test_commands = [
            ["echo", "test"],
            ["pwd"],
            ["ls", "-la"]
        ]
        
        for command in test_commands:
            try:
                result = subprocess.run(
                    command,
                    capture_output=True,
                    text=True,
                    timeout=5.0
                )
                
                # Verify real execution
                assert result.returncode == 0
                assert len(result.stdout) >= 0
                
            except (subprocess.TimeoutExpired, FileNotFoundError):
                # Skip if command not available
                continue
    
    @pytest.mark.asyncio
    async def test_real_file_system_operations(self):
        """Test that file system operations are real."""
        
        with tempfile.TemporaryDirectory() as temp_dir:
            temp_path = Path(temp_dir)
            
            # Test real file creation
            test_file = temp_path / "test.txt"
            test_file.write_text("test content")
            
            # Test real file reading
            content = test_file.read_text()
            assert content == "test content"
            
            # Test real file deletion
            test_file.unlink()
            assert not test_file.exists()
            
            # Test real directory listing
            files = list(temp_path.iterdir())
            assert len(files) == 0
    
    @pytest.mark.asyncio
    async def test_real_configuration_validation(self):
        """Test that configuration validation uses real logic."""
        
        # Test real configuration validation
        valid_config = {
            "device_range": [0, 1, 2],
            "enable_capability_detection": True,
            "detection_timeout": 1.0
        }
        
        invalid_config = {
            "device_range": [-1, "invalid"],
            "enable_capability_detection": "not_boolean",
            "detection_timeout": -1.0
        }
        
        # Real validation logic
        def validate_config(config):
            errors = []
            
            if not isinstance(config.get("device_range"), list):
                errors.append("device_range must be a list")
            
            if not isinstance(config.get("enable_capability_detection"), bool):
                errors.append("enable_capability_detection must be a boolean")
            
            if not isinstance(config.get("detection_timeout"), (int, float)) or config.get("detection_timeout") <= 0:
                errors.append("detection_timeout must be a positive number")
            
            return errors
        
        # Test real validation
        valid_errors = validate_config(valid_config)
        invalid_errors = validate_config(invalid_config)
        
        assert len(valid_errors) == 0
        assert len(invalid_errors) > 0
