# FFmpeg execution, process cleanup

@pytest.mark.asyncio
async def test_snapshot_process_cleanup_on_timeout():
    """Test robust process cleanup when FFmpeg times out."""
    # Mock FFmpeg process that hangs
    # Verify graceful termination then force kill
    # Assert proper error context in response

@pytest.mark.asyncio  
async def test_snapshot_file_size_error_handling():
    """Test handling when file exists but size cannot be determined."""
    # Mock OSError on getsize
    # Verify warning in response but successful completion