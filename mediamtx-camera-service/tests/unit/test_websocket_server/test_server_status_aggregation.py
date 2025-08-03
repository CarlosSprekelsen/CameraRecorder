# tests/unit/test_websocket_server/test_server_status_aggregation.py

@pytest.mark.asyncio
async def test_get_camera_status_uses_real_capability_data():
    """Verify get_camera_status integrates real capability metadata when available."""
    # Mock camera monitor with capability metadata support
    # Verify resolution/fps come from capability data, not defaults

@pytest.mark.asyncio 
async def test_get_camera_list_capability_integration():
    """Verify get_camera_list uses real capability data for resolution/fps."""
    # Mock connected cameras with capability metadata
    # Verify real data used over architecture defaults

def test_notification_field_filtering():
    """Verify notifications only include API-specified fields."""
    # Test camera_status_update filters to: device, status, name, resolution, fps, streams
    # Test recording_status_update filters to: device, status, filename, duration

def test_graceful_degradation_missing_dependencies():
    """Verify methods handle missing camera_monitor/mediamtx_controller gracefully."""
    # Test empty responses when dependencies unavailable
    # Verify no crashes, appropriate error handling