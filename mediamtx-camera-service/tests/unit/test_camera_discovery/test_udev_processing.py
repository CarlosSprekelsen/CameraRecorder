import pytest
from unittest.mock import AsyncMock


@pytest.mark.asyncio
async def test_udev_event_filtering(monitor):
    class MockUdevDevice:
        def __init__(self, device_node, action):
            self.device_node = device_node
            self.action = action

    processed = []

    async def fake_handle(event_data):
        processed.append(event_data.device_path)

    monitor._handle_camera_event = fake_handle
    monitor._create_camera_device_info = AsyncMock()

    cases = [
        ("/dev/video0", "add", True),
        ("/dev/video5", "add", False),
    ]

    for node, action, should in cases:
        dev = MockUdevDevice(node, action)
        await monitor._process_udev_device_event(dev)

    assert "/dev/video0" in processed
    assert len(processed) == 1
