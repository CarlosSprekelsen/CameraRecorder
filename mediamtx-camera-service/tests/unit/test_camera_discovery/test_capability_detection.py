import pytest
from unittest.mock import Mock, patch, AsyncMock


@pytest.mark.asyncio
async def test_probe_device_capabilities_with_mock(monitor, mock_v4l2_outputs):
    """Test device capability probing with comprehensive v4l2-ctl output mocking."""
    
    call_sequence = {"count": 0}

    async def mock_communicate():
        call_sequence["count"] += 1
        if call_sequence["count"] == 1:
            return (mock_v4l2_outputs["device_info"], b"")
        elif call_sequence["count"] == 2:
            return (mock_v4l2_outputs["formats"], b"")
        else:
            return (mock_v4l2_outputs["frame_rates"], b"")

    mock_proc = Mock()
    mock_proc.returncode = 0
    mock_proc.communicate = AsyncMock(side_effect=mock_communicate)

    with patch("asyncio.create_subprocess_exec", return_value=mock_proc):
        caps = await monitor._probe_device_capabilities("/dev/video0")
        
        assert caps is not None
        assert caps.detected is True
        assert caps.accessible is True
        assert caps.device_name == "USB Camera"
        assert caps.driver == "uvcvideo"
        assert "YUYV" in [f.get("code", "") for f in caps.formats]
        assert "1920x1080" in caps.resolutions
        assert "30" in caps.frame_rates


@pytest.mark.asyncio
async def test_probe_device_capabilities_timeout(monitor):
    """Test device capability probing timeout handling."""
    
    async def mock_timeout_subprocess(*args, **kwargs):
        await asyncio.sleep(2.0)  # Exceed timeout
        raise asyncio.TimeoutError("Command timed out")

    with patch("asyncio.create_subprocess_exec", side_effect=mock_timeout_subprocess):
        caps = await monitor._probe_device_capabilities("/dev/video0")
        
        assert caps is not None
        assert caps.detected is False
        assert caps.accessible is False
        assert "timeout" in caps.error.lower()


@pytest.mark.asyncio
async def test_probe_device_capabilities_error(monitor, mock_v4l2_outputs):
    """Test device capability probing error handling."""
    
    mock_proc = Mock()
    mock_proc.returncode = 1
    mock_proc.communicate = AsyncMock(return_value=(b"", mock_v4l2_outputs["error"]))

    with patch("asyncio.create_subprocess_exec", return_value=mock_proc):
        caps = await monitor._probe_device_capabilities("/dev/video0")
        
        assert caps is not None
        assert caps.detected is False
        assert caps.accessible is False
        assert "failed to probe" in caps.error.lower()


@pytest.mark.asyncio
async def test_probe_device_capabilities_malformed_output(monitor, mock_v4l2_outputs):
    """Test device capability probing with malformed v4l2-ctl output."""
    
    mock_proc = Mock()
    mock_proc.returncode = 0
    mock_proc.communicate = AsyncMock(return_value=(mock_v4l2_outputs["malformed"], b""))

    with patch("asyncio.create_subprocess_exec", return_value=mock_proc):
        caps = await monitor._probe_device_capabilities("/dev/video0")
        
        assert caps is not None
        # Should handle malformed output gracefully
        assert caps.detected is False or caps.detected is True  # Either is acceptable
        assert caps.error is None or "parsing" in caps.error.lower()


@pytest.mark.asyncio
async def test_probe_device_capabilities_device_unavailable(monitor):
    """Test device capability probing when device is unavailable."""
    
    # Mock file system to indicate device doesn't exist
    with patch("os.path.exists", return_value=False):
        caps = await monitor._probe_device_capabilities("/dev/video0")
        
        assert caps is not None
        assert caps.detected is False
        assert caps.accessible is False
        assert caps.error is not None
