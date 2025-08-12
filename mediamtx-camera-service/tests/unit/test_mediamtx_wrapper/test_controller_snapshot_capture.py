# tests/unit/test_mediamtx_wrapper/test_controller_snapshot_capture.py
"""
Test snapshot capture robustness, process cleanup, and metadata accuracy.

Test policy: Verify FFmpeg process management, timeout handling, and accurate
metadata return even under failure conditions.
"""

import pytest
import asyncio
import os
import tempfile
from unittest.mock import Mock, AsyncMock, patch

from src.mediamtx_wrapper.controller import MediaMTXController
from .async_mock_helpers import (
    create_mock_session,
    create_async_mock_with_response,
    create_async_mock_with_side_effect
)


class TestSnapshotCapture:
    """Test snapshot capture with robust process management."""

    @pytest.fixture
    def controller(self):
        """Create MediaMTX controller with test configuration."""
        with tempfile.TemporaryDirectory() as temp_dir:
            controller = MediaMTXController(
                host="localhost",
                api_port=9997,
                rtsp_port=8554,
                webrtc_port=8889,
                hls_port=8888,
                config_path="/tmp/test_config.yml",
                recordings_path=os.path.join(temp_dir, "recordings"),
                snapshots_path=os.path.join(temp_dir, "snapshots"),
                process_termination_timeout=1.0,  # Short timeout for testing
                process_kill_timeout=0.5,
            )
            # Mock session to avoid HTTP calls with proper async context manager support
            controller._session = create_mock_session()
            yield controller

    @pytest.mark.asyncio
    async def test_snapshot_process_cleanup_on_timeout(self, controller):
        """Test robust process cleanup when FFmpeg times out."""
        # Mock FFmpeg process that hangs indefinitely
        mock_process = Mock()
        mock_process.returncode = None  # Process still running
        mock_process.communicate = AsyncMock(side_effect=asyncio.TimeoutError())
        mock_process.terminate = Mock()
        mock_process.kill = Mock()
        mock_process.wait = AsyncMock(
            side_effect=asyncio.TimeoutError()
        )  # Doesn't respond to signals

        with patch("asyncio.create_subprocess_exec", return_value=mock_process):
            result = await controller.take_snapshot("test_stream", "test_snapshot.jpg")

            # Verify graceful termination then force kill was attempted
            mock_process.terminate.assert_called_once()
            mock_process.kill.assert_called_once()

            # Verify error context includes process cleanup information
            assert result["status"] == "failed"
            assert "timeout" in result["error"].lower()
            assert "killed" in result["error"] or "terminated" in result["error"]

    @pytest.mark.asyncio
    async def test_snapshot_file_size_error_handling(self, controller):
        """Test handling when file exists but size cannot be determined."""
        # Mock successful FFmpeg execution
        mock_process = Mock()
        mock_process.returncode = 0
        mock_process.communicate = AsyncMock(return_value=(b"success", b""))

        # Mock file that exists but raises OSError on getsize
        with (
            patch("asyncio.create_subprocess_exec", return_value=mock_process),
            patch("os.path.exists", return_value=True),
            patch("os.path.getsize", side_effect=OSError("Permission denied")),
        ):

            result = await controller.take_snapshot("test_stream", "test_snapshot.jpg")

            # Verify successful completion with warning about file size
            assert result["status"] == "completed"
            assert result["file_size"] == 0
            assert "warning" in result
            assert "Could not determine file size" in result["warning"]

    @pytest.mark.asyncio
    async def test_snapshot_directory_permission_error(self, controller):
        """Test handling when snapshots directory cannot be created or written to."""
        # Mock permission error when creating directory
        with patch("os.makedirs", side_effect=PermissionError("Access denied")):
            result = await controller.take_snapshot("test_stream", "test_snapshot.jpg")

            # Verify graceful error handling
            assert result["status"] == "failed"
            assert "Cannot write to snapshots directory" in result["error"]
            assert (
                "Permission denied" in result["error"]
                or "Access denied" in result["error"]
            )

    @pytest.mark.asyncio
    async def test_snapshot_ffmpeg_nonzero_exit_code(self, controller):
        """Test handling when FFmpeg exits with error code."""
        # Mock FFmpeg process that fails
        mock_process = Mock()
        mock_process.returncode = 1  # Error exit code
        mock_process.communicate = AsyncMock(return_value=(b"", b"Input/output error"))

        with (
            patch("asyncio.create_subprocess_exec", return_value=mock_process),
            patch("os.path.exists", return_value=False),
        ):  # No output file created

            result = await controller.take_snapshot("test_stream", "test_snapshot.jpg")

            # Verify error is properly captured and reported
            assert result["status"] == "failed"
            assert "FFmpeg capture failed" in result["error"]
            assert "Input/output error" in result["error"]

    @pytest.mark.asyncio
    async def test_snapshot_success_with_accurate_metadata(self, controller):
        """Test successful snapshot capture returns accurate metadata."""
        # Mock successful FFmpeg execution
        mock_process = Mock()
        mock_process.returncode = 0
        mock_process.communicate = AsyncMock(return_value=(b"success", b""))

        test_file_size = 12345
        test_file_path = os.path.join(controller._snapshots_path, "test_snapshot.jpg")

        with (
            patch("asyncio.create_subprocess_exec", return_value=mock_process),
            patch("os.path.exists", return_value=True),
            patch("os.path.getsize", return_value=test_file_size),
            patch("os.makedirs"),
        ):

            result = await controller.take_snapshot("test_stream", "test_snapshot.jpg")

            # Verify accurate metadata
            assert result["status"] == "completed"
            assert result["filename"] == "test_snapshot.jpg"
            assert result["file_size"] == test_file_size
            assert result["file_path"] == test_file_path
            assert "timestamp" in result

    @pytest.mark.asyncio
    async def test_snapshot_process_creation_timeout(self, controller):
        """Test timeout during FFmpeg process creation."""
        # Mock process creation that times out
        with patch(
            "asyncio.create_subprocess_exec", side_effect=asyncio.TimeoutError()
        ):
            result = await controller.take_snapshot("test_stream", "test_snapshot.jpg")

            # Verify timeout is handled gracefully
            assert result["status"] == "failed"
            assert "timeout" in result["error"].lower()

    def test_cleanup_ffmpeg_process_escalation_path(self, controller):
        """Test FFmpeg process cleanup escalation: SIGTERM → SIGKILL."""
        # This would be a unit test of the _cleanup_ffmpeg_process method
        # Testing the escalation logic without actual process creation

        # Mock process that doesn't respond to SIGTERM but does to SIGKILL
        mock_process = Mock()
        mock_process.returncode = None
        mock_process.terminate = Mock()
        mock_process.kill = Mock()

        # First wait (after SIGTERM) times out, second wait (after SIGKILL) succeeds
        mock_process.wait = AsyncMock(side_effect=[asyncio.TimeoutError(), None])

        # Test the cleanup method directly
        cleanup_result = asyncio.run(
            controller._cleanup_ffmpeg_process(
                mock_process, "test_stream", "test_correlation"
            )
        )

        # Verify escalation path was followed
        mock_process.terminate.assert_called_once()
        mock_process.kill.assert_called_once()
        assert "terminated" in cleanup_result
        assert "killed" in cleanup_result
        assert "force_exit" in cleanup_result


# Test configuration expectations:
# - Mock asyncio.create_subprocess_exec for FFmpeg process control
# - Mock os.path.exists and os.path.getsize for file system operations
# - Mock os.makedirs for directory creation testing
# - Use temporary directories for file system tests
# - Test both success and failure paths for robustness
# - Verify correlation IDs are properly set in error logging
