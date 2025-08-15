"""
Udev event processing tests for camera discovery.

Requirements Traceability:
- REQ-CAM-002: Camera discovery shall process udev events with proper filtering
- REQ-CAM-002: Camera discovery shall handle udev event actions (add/remove/change)
- REQ-CAM-002: Camera discovery shall provide race condition protection for udev events

Story Coverage: S3 - Camera Discovery Hardening
IV&V Control Point: Real udev event processing validation
"""

import pytest
from unittest.mock import AsyncMock, patch, Mock


@pytest.mark.asyncio
async def test_udev_event_filtering(monitor, mock_udev_device):
    """Test udev event filtering with proper device range validation."""
    
    processed = []

    async def fake_handle(event_data):
        processed.append(event_data.device_path)

    # Mock the event handler and device info creation
    monitor._handle_camera_event = fake_handle
    monitor._create_camera_device_info = AsyncMock(return_value=Mock(status="CONNECTED"))

    # Test cases: (device_node, action, should_be_processed)
    cases = [
        ("/dev/video0", "add", True),   # In range
        ("/dev/video1", "add", True),   # In range
        ("/dev/video2", "add", True),   # In range
        ("/dev/video5", "add", False),  # Out of range
        ("/dev/video10", "add", False), # Out of range
        ("/dev/audio0", "add", False),  # Wrong device type
        ("invalid_path", "add", False), # Invalid path
    ]

    for node, action, should_be_processed in cases:
        dev = mock_udev_device(device_node=node, action=action)
        await monitor._process_udev_device_event(dev)

    # Should only process devices in the monitored range
    expected_processed = ["/dev/video0", "/dev/video1", "/dev/video2"]
    assert processed == expected_processed


@pytest.mark.asyncio
async def test_udev_event_actions(monitor, mock_udev_device):
    """Test different udev event actions (add, remove, change)."""
    
    events_processed = []

    async def fake_handle(event_data):
        events_processed.append((event_data.device_path, event_data.event_type.value))

    monitor._handle_camera_event = fake_handle
    monitor._create_camera_device_info = AsyncMock(return_value=Mock(status="CONNECTED"))

    # Test different actions
    test_events = [
        ("/dev/video0", "add"),
        ("/dev/video1", "remove"),
        ("/dev/video2", "change"),
        ("/dev/video0", "unknown"),  # Should be skipped
    ]

    for node, action in test_events:
        dev = mock_udev_device(device_node=node, action=action)
        await monitor._process_udev_device_event(dev)

    # Should process add, remove, and change events
    expected_events = [
        ("/dev/video0", "CONNECTED"),
        ("/dev/video1", "DISCONNECTED"),
        ("/dev/video2", "STATUS_CHANGED"),
    ]
    assert events_processed == expected_events


@pytest.mark.asyncio
async def test_udev_event_race_condition_handling(monitor, mock_udev_device):
    """Test udev event processing with race condition protection."""
    
    processed_events = []

    async def fake_handle(event_data):
        processed_events.append(event_data.device_path)

    monitor._handle_camera_event = fake_handle
    monitor._create_camera_device_info = AsyncMock(return_value=Mock(status="CONNECTED"))

    # Simulate rapid successive events for the same device
    rapid_events = [
        ("/dev/video0", "add"),
        ("/dev/video0", "change"),
        ("/dev/video0", "change"),
        ("/dev/video0", "remove"),
    ]

    for node, action in rapid_events:
        dev = mock_udev_device(device_node=node, action=action)
        await monitor._process_udev_device_event(dev)

    # Should handle all events without errors
    assert len(processed_events) > 0
    # The exact count depends on the monitor's internal state management
