"""
Unit test configuration and fixtures for MediaMTX Camera Service.

This configuration provides:
1. Real component fixtures without excessive mocking
2. Consistent test environment setup
3. Proper test isolation and cleanup
4. Standardized test data and configurations
5. Enforcement of real component testing policies
"""

import pytest
import pytest_asyncio
import asyncio
import tempfile
import warnings
from unittest.mock import Mock, AsyncMock, patch
from pathlib import Path
from camera_discovery.hybrid_monitor import HybridCameraMonitor


@pytest.fixture(scope="function")
def event_loop():
    """Create an instance of the default event loop for each test case."""
    loop = asyncio.new_event_loop()
    asyncio.set_event_loop(loop)
    yield loop
    
    # Proper cleanup of pending tasks
    pending = asyncio.all_tasks(loop)
    if pending:
        loop.run_until_complete(asyncio.gather(*pending, return_exceptions=True))
    
    loop.close()


@pytest_asyncio.fixture(autouse=True)
async def cleanup_async_resources():
    """Clean up async resources after each test."""
    yield
    
    # Clean up any remaining tasks
    try:
        loop = asyncio.get_running_loop()
        pending = asyncio.all_tasks(loop)
        if pending:
            await asyncio.gather(*pending, return_exceptions=True)
    except RuntimeError:
        # No running loop
        pass


@pytest.fixture
def monitor():
    """Create a basic monitor with capability detection enabled."""
    return HybridCameraMonitor(device_range=[0, 1, 2], enable_capability_detection=True)

@pytest.fixture
def test_controller_config():
    """Provide consistent MediaMTX controller configuration for tests."""
    return {
        "host": "127.0.0.1",  # Use IP instead of localhost
        "api_port": 9997,
        "rtsp_port": 8554,
        "webrtc_port": 8889,
        "hls_port": 8888,
        "config_path": "/tmp/test_config.yml",
        "recordings_path": "/tmp/test_recordings",
        "snapshots_path": "/tmp/test_snapshots",
        "health_check_interval": 0.1,
        "health_failure_threshold": 5,
        "health_max_backoff_interval": 2.0,
        "backoff_base_multiplier": 2.0,
        "backoff_jitter_range": (1.0, 1.0),
        "health_circuit_breaker_timeout": 10.0,
    }

@pytest.fixture
def temp_test_files():
    """Create temporary test files that are cleaned up automatically."""
    with tempfile.TemporaryDirectory() as temp_dir:
        temp_path = Path(temp_dir)
        
        # Create test files
        config_file = temp_path / "test_config.yml"
        config_file.write_text("test: config")
        
        recordings_dir = temp_path / "recordings"
        recordings_dir.mkdir()
        
        snapshots_dir = temp_path / "snapshots"
        snapshots_dir.mkdir()
        
        yield {
            "temp_dir": temp_dir,
            "config_path": str(config_file),
            "recordings_path": str(recordings_dir),
            "snapshots_path": str(snapshots_dir),
        }


@pytest.fixture
def mock_dependencies():
    """Create mock dependencies for service manager testing."""
    return {
        "config": Mock(),
        "mediamtx_controller": Mock(),
        "websocket_server": Mock(),
        "camera_monitor": None,  # Will be set per test
    }


@pytest.fixture
def mock_udev_device():
    """Create a mock udev device for testing."""
    class MockUdevDevice:
        def __init__(self, device_node="/dev/video0", action="add"):
            self.device_node = device_node
            self.action = action
            self.device_type = "video4linux"
            self.subsystem = "video4linux"
    
    return MockUdevDevice


@pytest.fixture
def mock_camera_device_info():
    """Create mock camera device information."""
    from src.common.types import CameraDevice
    
    return CameraDevice(
        device="/dev/video0",
        name="USB Camera",
        status="CONNECTED",
        driver="uvcvideo",
        capabilities={
            "formats": ["YUYV", "MJPG"],
            "resolutions": ["1920x1080", "1280x720", "640x480"],
            "frame_rates": ["30", "25", "15"]
        }
    )


@pytest.fixture
def mock_capability_detection_result():
    """Create mock capability detection result."""
    from src.camera_discovery.hybrid_monitor import CapabilityDetectionResult
    
    return CapabilityDetectionResult(
        device_path="/dev/video0",
        detected=True,
        accessible=True,
        device_name="USB Camera",
        driver="uvcvideo",
        formats=[
            {"code": "YUYV", "description": "YUYV 4:2:2"},
            {"code": "MJPG", "description": "Motion JPEG"}
        ],
        resolutions=["1920x1080", "1280x720", "640x480"],
        frame_rates=["30", "25", "15"],
        error=None,
        timeout_context=None
    )


@pytest.fixture
def mock_v4l2_outputs():
    """Create mock v4l2-ctl command outputs."""
    return {
        "device_info": b"Driver name   : uvcvideo\nCard type     : USB Camera\nBus info      : usb-0000:00:14.0-1\n",
        "formats": b"Format [0]:\n  Name: YUYV\n  Description: YUYV 4:2:2\nSize: Discrete 1920x1080\nSize: Discrete 1280x720\nSize: Discrete 640x480\n",
        "frame_rates": b"Frame rate: 30.000 fps\nFrame rate: 25.000 fps\nFrame rate: 15.000 fps\n",
        "error": b"v4l2-ctl: failed to open /dev/video0: Device or resource busy\n",
        "timeout": b"",  # Empty output for timeout
        "malformed": b"Some random text without useful info\n"
    }


@pytest.fixture
def mock_subprocess_process():
    """Create a mock subprocess process for v4l2-ctl commands."""
    class MockSubprocessProcess:
        def __init__(self, stdout=b"", stderr=b"", returncode=0):
            self.stdout = stdout
            self.stderr = stderr
            self.returncode = returncode
            self._closed = False
            self._transport = None
        
        async def communicate(self):
            return (self.stdout, self.stderr)
        
        def close(self):
            self._closed = True
            if self._transport:
                try:
                    self._transport.close()
                except Exception:
                    pass
        
        def __del__(self):
            try:
                if not self._closed:
                    self.close()
            except Exception:
                pass
    
    return MockSubprocessProcess


@pytest.fixture(autouse=True)
def suppress_event_loop_warnings():
    """Suppress event loop closed warnings during test cleanup."""
    with warnings.catch_warnings():
        warnings.filterwarnings("ignore", message=".*Event loop is closed.*", category=RuntimeWarning)
        yield


@pytest.fixture
def camera_discovery_environment():
    """Set up a complete camera discovery test environment."""
    # Mock pyudev if not available
    with patch.dict('sys.modules', {'pyudev': Mock()}):
        # Mock subprocess calls
        with patch('asyncio.create_subprocess_exec') as mock_subprocess:
            # Mock file system operations
            with patch('os.path.exists', return_value=True):
                # Mock device file operations
                with patch('builtins.open', create=True):
                    yield {
                        'mock_subprocess': mock_subprocess,
                        'monitor': HybridCameraMonitor(
                            device_range=[0, 1, 2],
                            enable_capability_detection=True,
                            detection_timeout=1.0  # Short timeout for testing
                        )
                    }
