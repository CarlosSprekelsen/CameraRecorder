# tests/unit/test_mediamtx_wrapper/test_controller_recording_duration.py
"""
Test recording lifecycle duration calculation and file handling robustness.

Test policy: Verify accurate duration computation, graceful handling of 
missing files, permission errors, and proper session management.
"""

import pytest
import asyncio
import os
import time
import tempfile
from unittest.mock import Mock, AsyncMock, patch
from pathlib import Path

from src.mediamtx_wrapper.controller import MediaMTXController


class TestRecordingDuration:
    """Test recording duration calculation and file handling."""

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
                snapshots_path=os.path.join(temp_dir, "snapshots")
            )
            # Mock session for HTTP calls
            controller._session = Mock()
            yield controller

    @pytest.fixture
    def mock_http_success(self):
        """Mock successful HTTP responses."""
        def _mock_response(status=200, json_data=None):
            response = Mock()
            response.status = status
            response.json = AsyncMock(return_value=json_data or {})
            response.text = AsyncMock(return_value="")
            return response
        return _mock_response

    @pytest.mark.asyncio
    async def test_recording_duration_calculation_precision(self, controller, mock_http_success):
        """Test accurate duration calculation using session timestamps."""
        # Mock successful HTTP responses for start and stop
        controller._session.post = AsyncMock(return_value=mock_http_success())
        
        # Start recording and capture start time
        start_time = time.time()
        await controller.start_recording("test_stream", duration=3600, format="mp4")
        
        # Simulate passage of time
        test_duration = 123  # 123 seconds
        with patch('time.time', return_value=start_time + test_duration):
            # Mock file exists and has size
            with patch('os.path.exists', return_value=True), \
                 patch('os.path.getsize', return_value=1024000):
                
                result = await controller.stop_recording("test_stream")
        
        # Verify duration calculation is accurate
        assert result["duration"] == test_duration
        assert result["status"] == "completed"
        assert abs(result["duration"] - test_duration) <= 1  # Allow 1 second tolerance

    @pytest.mark.asyncio
    async def test_recording_missing_file_handling(self, controller, mock_http_success):
        """Test stop_recording when file doesn't exist on disk."""
        # Setup recording session
        controller._session.post = AsyncMock(return_value=mock_http_success())
        await controller.start_recording("test_stream", format="mp4")
        
        # Mock file doesn't exist
        with patch('os.path.exists', return_value=False):
            result = await controller.stop_recording("test_stream")
        
        # Verify graceful handling of missing file
        assert result["status"] == "completed"
        assert result["file_exists"] is False
        assert result["file_size"] == 0
        assert "file_warning" in result
        assert "not found" in result["file_warning"]

    @pytest.mark.asyncio
    async def test_recording_file_permission_error(self, controller, mock_http_success):
        """Test handling when file exists but cannot be accessed due to permissions."""
        # Setup recording session
        controller._session.post = AsyncMock(return_value=mock_http_success())
        await controller.start_recording("test_stream", format="mp4")
        
        # Mock file exists but permission error on getsize
        with patch('os.path.exists', return_value=True), \
             patch('os.path.getsize', side_effect=PermissionError("Access denied")):
            
            result = await controller.stop_recording("test_stream")
        
        # Verify graceful error handling
        assert result["status"] == "completed"
        assert result["file_exists"] is True
        assert result["file_size"] == 0
        assert "file_warning" in result
        assert "Permission denied" in result["file_warning"]

    @pytest.mark.asyncio
    async def test_recording_directory_creation_permission_error(self, controller, mock_http_success):
        """Test handling when recordings directory cannot be created."""
        # Mock permission error when creating directory
        with patch('os.makedirs', side_effect=PermissionError("Access denied")), \
             patch('tempfile.NamedTemporaryFile', side_effect=PermissionError("Access denied")):
            
            # Attempt to start recording
            with pytest.raises(ValueError, match="Cannot write to recordings directory"):
                await controller.start_recording("test_stream", format="mp4")

    @pytest.mark.asyncio
    async def test_recording_session_management(self, controller, mock_http_success):
        """Test recording session tracking and cleanup."""
        controller._session.post = AsyncMock(return_value=mock_http_success())
        
        # Start recording - should create session
        await controller.start_recording("test_stream", format="mp4")
        assert "test_stream" in controller._recording_sessions
        
        # Verify session contains required fields
        session = controller._recording_sessions["test_stream"]
        assert "start_time" in session
        assert "filename" in session
        assert "record_path" in session
        assert "correlation_id" in session
        
        # Stop recording - should clean up session
        with patch('os.path.exists', return_value=True), \
             patch('os.path.getsize', return_value=1024):
            await controller.stop_recording("test_stream")
        
        # Session should be cleaned up
        assert "test_stream" not in controller._recording_sessions

    @pytest.mark.asyncio
    async def test_recording_api_failure_preserves_session(self, controller):
        """Test that API failures during stop don't lose session data for retry."""
        # Start recording successfully
        success_response = Mock()
        success_response.status = 200
        controller._session.post = AsyncMock(return_value=success_response)
        await controller.start_recording("test_stream", format="mp4")
        
        # Mock API failure during stop
        failure_response = Mock()
        failure_response.status = 500
        failure_response.text = AsyncMock(return_value="Internal Server Error")
        controller._session.post = AsyncMock(return_value=failure_response)
        
        # Attempt to stop recording
        with pytest.raises(ValueError, match="Failed to stop recording"):
            await controller.stop_recording("test_stream")
        
        # Session should still exist for retry
        assert "test_stream" in controller._recording_sessions

    @pytest.mark.asyncio
    async def test_recording_duplicate_start_error(self, controller, mock_http_success):
        """Test error when trying to start recording on already recording stream."""
        controller._session.post = AsyncMock(return_value=mock_http_success())
        
        # Start first recording
        await controller.start_recording("test_stream", format="mp4")
        
        # Attempt to start second recording on same stream
        with pytest.raises(ValueError, match="Recording already active"):
            await controller.start_recording("test_stream", format="mp4")

    @pytest.mark.asyncio
    async def test_recording_stop_without_start_error(self, controller):
        """Test error when trying to stop recording that was never started."""
        # Mock session for stop request
        controller._session = Mock()
        
        # Attempt to stop recording without starting
        with pytest.raises(ValueError, match="No active recording session found"):
            await controller.stop_recording("test_stream")

    @pytest.mark.asyncio
    async def test_recording_format_validation(self, controller):
        """Test validation of recording format parameter."""
        # Test invalid format
        with pytest.raises(ValueError, match="Invalid format.*Must be one of"):
            await controller.start_recording("test_stream", format="invalid")
        
        # Test valid formats
        valid_formats = ["mp4", "mkv", "avi"]
        for format_type in valid_formats:
            # This would normally require mocking the HTTP call
            # but we're just testing the validation doesn't raise
            try:
                controller._session = Mock()
                success_response = Mock()
                success_response.status = 200
                controller._session.post = AsyncMock(return_value=success_response)
                
                await controller.start_recording(f"test_stream_{format_type}", format=format_type)
                # Clean up for next iteration
                await controller.stop_recording(f"test_stream_{format_type}")
            except ValueError as e:
                if "Invalid format" in str(e):
                    pytest.fail(f"Valid format {format_type} was rejected")


# Test configuration expectations:
# - Mock aiohttp session.post for MediaMTX API calls
# - Mock time.time() for duration calculation testing
# - Mock os.path.exists and os.path.getsize for file operations
# - Mock os.makedirs for directory creation testing
# - Use temporary directories for file system tests
# - Test session management and cleanup
# - Verify correlation IDs in session data
# - Test both success and failure scenarios