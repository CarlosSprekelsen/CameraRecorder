"""
Unit tests for StreamLifecycleManager.

Requirements Coverage:
- REQ-STREAM-001: Streams must remain active during file rotation (30-minute intervals)
- REQ-STREAM-002: Different lifecycle policies for different use cases
- REQ-STREAM-003: Power-efficient operation with on-demand activation
- REQ-STREAM-004: Manual control over stream lifecycle for recording scenarios

Test Categories: Unit
API Documentation Reference: N/A (Internal component, no external API)
"""

import asyncio
import pytest
from unittest.mock import AsyncMock, MagicMock, patch
from pathlib import Path

from mediamtx_wrapper.stream_lifecycle_manager import (
    StreamLifecycleManager,
    StreamUseCase,
    StreamConfig,
    ValidationError,
    MediaMTXAPIError,
    StreamLifecycleManagerError,
)


class TestStreamLifecycleManager:
    """Test cases for StreamLifecycleManager."""

    @pytest.fixture
    def manager(self):
        """Create a StreamLifecycleManager instance for testing."""
        return StreamLifecycleManager(
            mediamtx_api_url="http://localhost:9997",
            mediamtx_config_path="/tmp/mediamtx.yml",
            logger=MagicMock(),
        )

    @pytest.fixture
    async def async_manager(self):
        """Create an async StreamLifecycleManager instance for testing."""
        async with StreamLifecycleManager(
            mediamtx_api_url="http://localhost:9997",
            mediamtx_config_path="/tmp/mediamtx.yml",
            logger=MagicMock(),
        ) as manager:
            yield manager

    @pytest.mark.unit
    def test_init(self, manager):
        """REQ-STREAM-002: Test StreamLifecycleManager initialization."""
        assert manager.mediamtx_api_url == "http://localhost:9997"
        assert manager.mediamtx_config_path == Path("/tmp/mediamtx.yml")
        assert manager._stream_configs == {}
        assert manager._active_streams == {}
        assert manager._session is None

    @pytest.mark.unit
    def test_validate_device_path_valid(self, manager):
        """REQ-STREAM-003: Test valid device path validation."""
        valid_paths = ["/dev/video0", "/dev/video1", "/dev/custom", "/dev/custom123"]
        for path in valid_paths:
            # Should not raise exception
            manager._validate_device_path(path)

    @pytest.mark.unit
    def test_validate_device_path_invalid(self, manager):
        """REQ-STREAM-003: Test invalid device path validation."""
        invalid_paths = [
            "",  # Empty
            None,  # None
            "/dev/invalid",  # Invalid format
            "/tmp/video0",  # Wrong directory
            "video0",  # No /dev prefix
        ]
        for path in invalid_paths:
            with pytest.raises(ValidationError):
                manager._validate_device_path(path)

    @pytest.mark.unit
    def test_validate_stream_name_valid(self, manager):
        """REQ-STREAM-003: Test valid stream name validation."""
        valid_names = ["camera0", "camera1_viewing", "custom_stream", "test-123"]
        for name in valid_names:
            # Should not raise exception
            manager._validate_stream_name(name)

    @pytest.mark.unit
    def test_validate_stream_name_invalid(self, manager):
        """REQ-STREAM-003: Test invalid stream name validation."""
        invalid_names = [
            "",  # Empty
            None,  # None
            "camera@0",  # Invalid character
            "camera 0",  # Space
            "camera.0",  # Dot
        ]
        for name in invalid_names:
            with pytest.raises(ValidationError):
                manager._validate_stream_name(name)

    @pytest.mark.unit
    def test_validate_use_case_valid(self, manager):
        """REQ-STREAM-002: Test valid use case validation."""
        valid_cases = [StreamUseCase.RECORDING, StreamUseCase.VIEWING, StreamUseCase.SNAPSHOT]
        for use_case in valid_cases:
            # Should not raise exception
            manager._validate_use_case(use_case)

    @pytest.mark.unit
    def test_validate_use_case_invalid(self, manager):
        """REQ-STREAM-002: Test invalid use case validation."""
        # Test with non-enum value
        with pytest.raises(ValidationError):
            manager._validate_use_case("invalid")

        # Test with None
        with pytest.raises(ValidationError):
            manager._validate_use_case(None)

    @pytest.mark.unit
    def test_get_stream_name(self, manager):
        """REQ-STREAM-002: Test stream name generation for different devices and use cases."""
        # Test camera device paths
        assert manager._get_stream_name("/dev/video0", StreamUseCase.RECORDING) == "camera0"
        assert manager._get_stream_name("/dev/video1", StreamUseCase.VIEWING) == "camera1_viewing"
        assert manager._get_stream_name("/dev/video2", StreamUseCase.SNAPSHOT) == "camera2_snapshot"

        # Test other device names
        assert manager._get_stream_name("/dev/custom", StreamUseCase.RECORDING) == "custom"
        assert manager._get_stream_name("/dev/custom", StreamUseCase.VIEWING) == "custom_viewing"

    @pytest.mark.unit
    def test_build_ffmpeg_command(self, manager):
        """REQ-STREAM-003: Test FFmpeg command generation."""
        command = manager._build_ffmpeg_command("/dev/video0", "camera0")
        expected = (
            "ffmpeg -f v4l2 -i /dev/video0 -c:v libx264 -preset ultrafast "
            "-tune zerolatency -f rtsp rtsp://localhost:8554/camera0"
        )
        assert command == expected

    @pytest.mark.unit
    @pytest.mark.asyncio
    async def test_configure_mediamtx_path_recording(self, async_manager):
        """REQ-STREAM-001: Test MediaMTX path configuration for recording use case."""
        with patch.object(async_manager, "_configure_mediamtx_path_api") as mock_configure:
            mock_configure.return_value = None

            result = await async_manager.configure_mediamtx_path(
                "/dev/video0", StreamUseCase.RECORDING
            )

            assert result is True
            mock_configure.assert_called_once()

            # Verify configuration was stored
            config_key = "/dev/video0:recording"
            assert config_key in async_manager._stream_configs

            config = async_manager._stream_configs[config_key]
            assert config.use_case == StreamUseCase.RECORDING
            assert config.device_path == "/dev/video0"
            assert config.stream_name == "camera0"
            assert config.run_on_demand_close_after == "0s"  # Never auto-close for recording

    @pytest.mark.unit
    @pytest.mark.asyncio
    async def test_configure_mediamtx_path_invalid_device(self, async_manager):
        """REQ-STREAM-003: Test MediaMTX path configuration with invalid device path."""
        with pytest.raises(ValidationError):
            await async_manager.configure_mediamtx_path("", StreamUseCase.RECORDING)

    @pytest.mark.unit
    @pytest.mark.asyncio
    async def test_configure_mediamtx_path_api_error(self, async_manager):
        """REQ-STREAM-003: Test MediaMTX path configuration with API error."""
        with patch.object(async_manager, "_configure_mediamtx_path_api") as mock_configure:
            mock_configure.side_effect = MediaMTXAPIError("API Error")

            with pytest.raises(MediaMTXAPIError):
                await async_manager.configure_mediamtx_path(
                    "/dev/video0", StreamUseCase.RECORDING
                )

    @pytest.mark.unit
    @pytest.mark.asyncio
    async def test_start_recording_stream(self, async_manager):
        """REQ-STREAM-001: Test starting a recording stream."""
        with patch.object(async_manager, "configure_mediamtx_path") as mock_configure, patch.object(
            async_manager, "_trigger_stream_activation"
        ) as mock_trigger:

            mock_configure.return_value = True
            mock_trigger.return_value = None

            result = await async_manager.start_recording_stream("/dev/video0")

            assert result is True
            mock_configure.assert_called_once_with("/dev/video0", StreamUseCase.RECORDING)
            mock_trigger.assert_called_once_with("camera0")

            # Verify active stream tracking
            assert "camera0" in async_manager._active_streams
            stream_info = async_manager._active_streams["camera0"]
            assert stream_info["device_path"] == "/dev/video0"
            assert stream_info["use_case"] == StreamUseCase.RECORDING

    @pytest.mark.unit
    @pytest.mark.asyncio
    async def test_start_stream_invalid_device(self, async_manager):
        """REQ-STREAM-003: Test starting stream with invalid device path."""
        with pytest.raises(ValidationError):
            await async_manager.start_recording_stream("")

    @pytest.mark.unit
    @pytest.mark.asyncio
    async def test_stop_stream_recording(self, async_manager):
        """REQ-STREAM-001: Test stopping a recording stream (should keep it active for file rotation)."""
        # Setup active recording stream
        async_manager._active_streams["camera0"] = {
            "device_path": "/dev/video0",
            "use_case": StreamUseCase.RECORDING,
            "start_time": 1000.0,
            "config": MagicMock(),
        }

        with patch.object(async_manager, "_stop_stream_api") as mock_stop:
            result = await async_manager.stop_stream(
                "/dev/video0", StreamUseCase.RECORDING, "test"
            )

            assert result is True
            # Recording streams should not be stopped via API (for file rotation compatibility)
            mock_stop.assert_not_called()

            # Should be removed from active streams
            assert "camera0" not in async_manager._active_streams

    @pytest.mark.unit
    @pytest.mark.asyncio
    async def test_stop_stream_invalid_device(self, async_manager):
        """REQ-STREAM-003: Test stopping stream with invalid device path."""
        with pytest.raises(ValidationError):
            await async_manager.stop_stream("", StreamUseCase.RECORDING)

    @pytest.mark.unit
    @pytest.mark.asyncio
    async def test_monitor_stream_health_invalid_device(self, async_manager):
        """REQ-STREAM-003: Test monitoring stream health with invalid device path."""
        with pytest.raises(ValidationError):
            await async_manager.monitor_stream_health("", StreamUseCase.RECORDING)

    @pytest.mark.unit
    def test_get_stream_config_invalid_device(self, manager):
        """REQ-STREAM-003: Test getting stream config with invalid device path."""
        with pytest.raises(ValidationError):
            manager.get_stream_config("", StreamUseCase.RECORDING)

    @pytest.mark.unit
    @pytest.mark.asyncio
    async def test_validate_mediamtx_api_response_success(self, async_manager):
        """REQ-STREAM-003: Test MediaMTX API response validation for success."""
        mock_response = MagicMock()
        mock_response.status = 200

        # Should not raise exception
        await async_manager._validate_mediamtx_api_response(mock_response)

    @pytest.mark.unit
    @pytest.mark.asyncio
    async def test_validate_mediamtx_api_response_error(self, async_manager):
        """REQ-STREAM-003: Test MediaMTX API response validation for error."""
        mock_response = MagicMock()
        mock_response.status = 500
        mock_response.text = AsyncMock(return_value="Internal Server Error")

        with pytest.raises(MediaMTXAPIError):
            await async_manager._validate_mediamtx_api_response(mock_response)

    @pytest.mark.unit
    @pytest.mark.asyncio
    async def test_validate_mediamtx_api_response_read_error(self, async_manager):
        """REQ-STREAM-003: Test MediaMTX API response validation when error reading response."""
        mock_response = MagicMock()
        mock_response.status = 500
        mock_response.text = AsyncMock(side_effect=Exception("Read error"))

        with pytest.raises(MediaMTXAPIError) as exc_info:
            await async_manager._validate_mediamtx_api_response(mock_response)
        
        assert "Unable to read error details" in str(exc_info.value)

    @pytest.mark.unit
    def test_get_correlation_id_new(self, manager):
        """REQ-STREAM-004: Test correlation ID generation when none exists."""
        with patch("mediamtx_wrapper.stream_lifecycle_manager.get_correlation_id") as mock_get:
            mock_get.return_value = None
            
            correlation_id = manager._get_correlation_id()
            
            assert correlation_id.startswith("stream-lifecycle-")
            assert len(correlation_id) == len("stream-lifecycle-") + 8

    @pytest.mark.unit
    def test_get_correlation_id_existing(self, manager):
        """REQ-STREAM-004: Test correlation ID retrieval when one exists."""
        with patch("mediamtx_wrapper.stream_lifecycle_manager.get_correlation_id") as mock_get:
            mock_get.return_value = "existing-id-123"
            
            correlation_id = manager._get_correlation_id()
            
            assert correlation_id == "existing-id-123"


class TestStreamConfig:
    """Test cases for StreamConfig dataclass."""

    @pytest.mark.unit
    def test_stream_config_creation(self):
        """REQ-STREAM-002: Test StreamConfig dataclass creation."""
        config = StreamConfig(
            use_case=StreamUseCase.RECORDING,
            device_path="/dev/video0",
            stream_name="camera0",
            run_on_demand_close_after="0s",
            run_on_demand_restart=True,
            run_on_demand_start_timeout="10s",
            ffmpeg_command="ffmpeg -f v4l2 -i /dev/video0 -c:v libx264 -f rtsp rtsp://localhost:8554/camera0",
        )

        assert config.use_case == StreamUseCase.RECORDING
        assert config.device_path == "/dev/video0"
        assert config.stream_name == "camera0"
        assert config.run_on_demand_close_after == "0s"
        assert config.run_on_demand_restart is True
        assert config.run_on_demand_start_timeout == "10s"
        assert "ffmpeg" in config.ffmpeg_command


class TestStreamUseCase:
    """Test cases for StreamUseCase enum."""

    @pytest.mark.unit
    def test_stream_use_case_values(self):
        """REQ-STREAM-002: Test StreamUseCase enum values."""
        assert StreamUseCase.RECORDING.value == "recording"
        assert StreamUseCase.VIEWING.value == "viewing"
        assert StreamUseCase.SNAPSHOT.value == "snapshot"

    @pytest.mark.unit
    def test_stream_use_case_membership(self):
        """REQ-STREAM-002: Test StreamUseCase enum membership."""
        assert StreamUseCase.RECORDING in StreamUseCase
        assert StreamUseCase.VIEWING in StreamUseCase
        assert StreamUseCase.SNAPSHOT in StreamUseCase


class TestExceptions:
    """Test cases for custom exceptions."""

    @pytest.mark.unit
    def test_exception_inheritance(self):
        """REQ-STREAM-003: Test exception inheritance hierarchy."""
        assert issubclass(ValidationError, StreamLifecycleManagerError)
        assert issubclass(MediaMTXAPIError, StreamLifecycleManagerError)
        assert issubclass(StreamLifecycleManagerError, Exception)

    @pytest.mark.unit
    def test_validation_error_message(self):
        """REQ-STREAM-003: Test ValidationError message."""
        error = ValidationError("Invalid device path")
        assert str(error) == "Invalid device path"

    @pytest.mark.unit
    def test_mediamtx_api_error_message(self):
        """REQ-STREAM-003: Test MediaMTXAPIError message."""
        error = MediaMTXAPIError("API call failed")
        assert str(error) == "API call failed"
