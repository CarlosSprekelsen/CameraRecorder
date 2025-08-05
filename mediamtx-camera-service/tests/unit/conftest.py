import pytest
from unittest.mock import Mock
from camera_discovery.hybrid_monitor import HybridCameraMonitor


@pytest.fixture
def monitor():
    return HybridCameraMonitor(device_range=[0, 1, 2], enable_capability_detection=True)


@pytest.fixture
def mock_dependencies():
    """Create mock dependencies for service manager testing."""
    return {
        "config": Mock(),
        "mediamtx_controller": Mock(),
        "websocket_server": Mock(),
        "camera_monitor": None,  # Will be set per test
    }
