# tests/unit/test_mediamtx_wrapper/test_controller_stream_operations.py
"""
Test stream creation/deletion idempotent behavior and error handling.

Test policy: Verify idempotent operations, clear error contexts, and
reliability under transient failures.
"""

import pytest
from unittest.mock import Mock, AsyncMock
import aiohttp

from src.mediamtx_wrapper.controller import MediaMTXController, StreamConfig
from .async_mock_helpers import (
    create_mock_session, 
    create_success_response, 
    create_failure_response,
    create_async_mock_with_response,
    create_async_mock_with_side_effect
)


class TestStreamOperations:
    """Test stream creation and deletion operations."""

    @pytest.fixture
    def controller(self):
        """Create MediaMTX controller with test configuration."""
        controller = MediaMTXController(
            host="localhost",
            api_port=9997,
            rtsp_port=8554,
            webrtc_port=8889,
            hls_port=8888,
            config_path="/tmp/test_config.yml",
            recordings_path="/tmp/recordings",
            snapshots_path="/tmp/snapshots",
        )
        # Mock session with proper async context manager support
        controller._session = create_mock_session()
        return controller

    @pytest.fixture
    def sample_stream_config(self):
        """Create sample stream configuration."""
        return StreamConfig(name="test_stream", source="/dev/video0", record=False)

    def _mock_response(self, status, json_data=None, text_data=""):
        """Helper to create mock HTTP response."""
        from .async_mock_helpers import MockResponse
        return MockResponse(status, json_data, text_data)

    @pytest.mark.asyncio
    async def test_create_stream_success(self, controller, sample_stream_config):
        """Test successful stream creation returns correct URLs."""
        # Mock successful response - session is already properly mocked

        # Mock get_stream_status to return stream doesn't exist (for idempotency check)
        controller.get_stream_status = AsyncMock(
            side_effect=ValueError("Stream not found")
        )

        result = await controller.create_stream(sample_stream_config)

        # Verify URLs are correctly generated
        expected_urls = {
            "rtsp": "rtsp://localhost:8554/test_stream",
            "webrtc": "http://localhost:8889/test_stream",
            "hls": "http://localhost:8888/test_stream",
        }
        assert result == expected_urls

    @pytest.mark.asyncio
    async def test_create_stream_idempotent_behavior(
        self, controller, sample_stream_config
    ):
        """Test that creating existing stream returns URLs without error."""
        # Mock get_stream_status to return existing stream
        controller.get_stream_status = AsyncMock(return_value={"name": "test_stream"})

        result = await controller.create_stream(sample_stream_config)

        # Should return URLs without making create API call
        expected_urls = {
            "rtsp": "rtsp://localhost:8554/test_stream",
            "webrtc": "http://localhost:8889/test_stream",
            "hls": "http://localhost:8888/test_stream",
        }
        assert result == expected_urls

    @pytest.mark.asyncio
    async def test_create_stream_conflict_409_idempotent(
        self, controller, sample_stream_config
    ):
        """Test 409 conflict response is handled idempotently."""
        # Mock get_stream_status to indicate stream doesn't exist initially
        controller.get_stream_status = AsyncMock(
            side_effect=ValueError("Stream not found")
        )

        # Mock 409 conflict response from create call
        conflict_response = self._mock_response(409, text_data="Path already exists")
        controller._session.post = create_async_mock_with_response(conflict_response)

        result = await controller.create_stream(sample_stream_config)

        # Should return URLs despite 409 conflict
        assert "rtsp" in result
        assert "test_stream" in result["rtsp"]

    @pytest.mark.asyncio
    async def test_create_stream_validation_errors(self, controller):
        """Test stream configuration validation."""
        # Test missing name
        with pytest.raises(ValueError, match="Stream name and source are required"):
            await controller.create_stream(StreamConfig(name="", source="/dev/video0"))

        # Test missing source
        with pytest.raises(ValueError, match="Stream name and source are required"):
            await controller.create_stream(StreamConfig(name="test", source=""))

    @pytest.mark.asyncio
    async def test_create_stream_api_error_with_context(
        self, controller, sample_stream_config
    ):
        """Test API error includes detailed context information."""
        # Mock get_stream_status to indicate stream doesn't exist
        controller.get_stream_status = AsyncMock(
            side_effect=ValueError("Stream not found")
        )

        # Mock API error response
        error_response = self._mock_response(500, text_data="Internal Server Error")
        controller._session.post = create_async_mock_with_response(error_response)

        with pytest.raises(ConnectionError) as exc_info:
            await controller.create_stream(sample_stream_config)

        # Verify error context includes stream details
        error_msg = str(exc_info.value)
        assert "test_stream" in error_msg
        assert "/dev/video0" in error_msg
        assert "record=False" in error_msg
        assert "HTTP 500" in error_msg

    @pytest.mark.asyncio
    async def test_create_stream_network_error(self, controller, sample_stream_config):
        """Test network connectivity error handling."""
        # Mock get_stream_status to indicate stream doesn't exist
        controller.get_stream_status = AsyncMock(
            side_effect=ValueError("Stream not found")
        )

        # Mock network error
        controller._session.post = create_async_mock_with_side_effect(
            lambda *args, **kwargs: aiohttp.ClientError("Connection refused")
        )

        with pytest.raises(ConnectionError, match="MediaMTX unreachable"):
            await controller.create_stream(sample_stream_config)

    @pytest.mark.asyncio
    async def test_delete_stream_success(self, controller):
        """Test successful stream deletion."""
        # Mock successful deletion response
        success_response = self._mock_response(200)
        controller._session.post = create_async_mock_with_response(success_response)

        result = await controller.delete_stream("test_stream")

        assert result is True

    @pytest.mark.asyncio
    async def test_delete_stream_idempotent_404(self, controller):
        """Test 404 response is handled idempotently (stream already deleted)."""
        # Mock 404 not found response
        not_found_response = self._mock_response(404, text_data="Path not found")
        controller._session.post = create_async_mock_with_response(not_found_response)

        result = await controller.delete_stream("nonexistent_stream")

        # Should return True (idempotent - stream already doesn't exist)
        assert result is True

    @pytest.mark.asyncio
    async def test_delete_stream_validation_error(self, controller):
        """Test stream name validation for deletion."""
        with pytest.raises(ValueError, match="Stream name is required"):
            await controller.delete_stream("")

    @pytest.mark.asyncio
    async def test_delete_stream_api_error(self, controller):
        """Test API error during deletion."""
        # Mock API error response
        error_response = self._mock_response(500, text_data="Internal Server Error")
        controller._session.post = create_async_mock_with_response(error_response)

        result = await controller.delete_stream("test_stream")

        # Should return False on API error (not idempotent case)
        assert result is False

    @pytest.mark.asyncio
    async def test_delete_stream_network_error(self, controller):
        """Test network error during deletion."""
        # Mock network error
        controller._session.post = create_async_mock_with_side_effect(
            lambda *args, **kwargs: aiohttp.ClientError("Connection refused")
        )

        with pytest.raises(ConnectionError, match="MediaMTX unreachable"):
            await controller.delete_stream("test_stream")

    @pytest.mark.asyncio
    async def test_stream_operations_without_session(
        self, controller, sample_stream_config
    ):
        """Test operations fail gracefully when controller not started."""
        # Remove session to simulate unstarted controller
        controller._session = None

        with pytest.raises(ConnectionError, match="MediaMTX controller not started"):
            await controller.create_stream(sample_stream_config)

        with pytest.raises(ConnectionError, match="MediaMTX controller not started"):
            await controller.delete_stream("test_stream")

    @pytest.mark.asyncio
    async def test_stream_config_with_recording(self, controller):
        """Test stream configuration with recording enabled."""
        # Mock get_stream_status to indicate stream doesn't exist
        controller.get_stream_status = AsyncMock(
            side_effect=ValueError("Stream not found")
        )

        # Mock successful response
        success_response = self._mock_response(200)
        controller._session.post = create_async_mock_with_response(success_response)

        recording_config = StreamConfig(
            name="recording_stream",
            source="/dev/video1",
            record=True,
            record_path="/tmp/recordings/test.mp4",
        )

        await controller.create_stream(recording_config)

        # Verify API call was made with correct recording configuration
        call_args = controller._session.post.call_args
        assert call_args is not None
        json_data = call_args.kwargs.get("json", {})
        assert json_data.get("record") is True
        assert json_data.get("recordPath") == "/tmp/recordings/test.mp4"

    def test_generate_stream_urls_format(self, controller):
        """Test stream URL generation format."""
        urls = controller._generate_stream_urls("test_stream")

        expected_urls = {
            "rtsp": "rtsp://localhost:8554/test_stream",
            "webrtc": "http://localhost:8889/test_stream",
            "hls": "http://localhost:8888/test_stream",
        }
        assert urls == expected_urls


# Test configuration expectations:
# - Mock aiohttp ClientSession for HTTP operations
# - Mock get_stream_status method for idempotency testing
# - Test both successful and error response scenarios
# - Verify error messages include relevant context
# - Test validation of input parameters
# - Test idempotent behavior for both operations
# - Verify correlation IDs are set for logging
