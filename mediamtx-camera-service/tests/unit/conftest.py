import pytest
from camera_discovery.hybrid_monitor import HybridCameraMonitor

@pytest.fixture
def monitor():
    return HybridCameraMonitor(device_range=[0, 1, 2], enable_capability_detection=True)
