# Duration calculation, file handling

@pytest.mark.asyncio
async def test_recording_duration_calculation_precision():
    """Test accurate duration calculation using session timestamps."""
    # Start recording, wait known interval, stop recording
    # Verify duration matches expected value within tolerance

@pytest.mark.asyncio
async def test_recording_missing_file_handling():
    """Test stop_recording when file doesn't exist on disk."""
    # Mock session with non-existent file
    # Verify file_exists=False and appropriate logging