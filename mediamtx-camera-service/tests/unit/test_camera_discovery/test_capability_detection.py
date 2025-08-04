import pytest
from unittest.mock import Mock, patch


@pytest.mark.asyncio
async def test_probe_device_capabilities_with_mock(monitor):
    # prepare mocked v4l2 outputs
    mock_info_output = b"Driver name   : uvcvideo\nCard type     : USB Camera\n"
    mock_formats_output = b"[0]: 'YUYV' (YUYV 4:2:2)\nSize: Discrete 640x480\n"

    call_sequence = {"count": 0}

    async def mock_communicate():
        call_sequence["count"] += 1
        if call_sequence["count"] == 1:
            return (mock_info_output, b"")
        elif call_sequence["count"] == 2:
            return (mock_formats_output, b"")
        else:
            return (b"30.000 fps\n", b"")

    mock_proc = Mock()
    mock_proc.returncode = 0
    mock_proc.communicate = mock_communicate

    with patch("asyncio.create_subprocess_exec", return_value=mock_proc):
        caps = await monitor._probe_device_capabilities("/dev/video0")
        assert caps is not None
        assert caps.detected is True
        assert "YUYV" in [f["code"] for f in caps.formats]
