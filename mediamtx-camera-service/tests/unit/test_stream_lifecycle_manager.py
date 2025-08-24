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
    def mock_session(self):
        """Create a mock aiohttp session for testing."""
        session = AsyncMock()
        session.__aenter__ = AsyncMock(return_value=session)
        session.__aexit__ = AsyncMock(return_value=None)
        return session

    @pytest.fixture
    async def async_manager(self, mock_session):
        """Create an async StreamLifecycleManager instance for testing."""
        manager = StreamLifecycleManager(
            mediamtx_api_url="http://localhost:9997",
            mediamtx_config_path="/tmp/mediamtx.yml",
            logger=MagicMock(),
        )
        
        # Mock the session creation to prevent hanging
        with patch('aiohttp.ClientSession', return_value=mock_session):
            manager._session = mock_session
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
    def test_use_case_configs(self, manager):
        """REQ-STREAM-002: Test use case specific configurations."""
        # Test recording configuration
        recording_config = manager._use_case_configs[StreamUseCase.RECORDING]
        assert recording_config["run_on_demand_close_after"] == "0s"  # Never auto-close
        assert recording_config["run_on_demand_restart"] is True
        assert recording_config["suffix"] == ""

        # Test viewing configuration
        viewing_config = manager._use_case_configs[StreamUseCase.VIEWING]
        assert viewing_config["run_on_demand_close_after"] == "300s"  # 5 minutes
        assert viewing_config["run_on_demand_restart"] is True
        assert viewing_config["suffix"] == "_viewing"

        # Test snapshot configuration
        snapshot_config = manager._use_case_configs[StreamUseCase.SNAPSHOT]
        assert snapshot_config["run_on_demand_close_after"] == "60s"  # 1 minute
        assert snapshot_config["run_on_demand_restart"] is False
        assert snapshot_config["suffix"] == "_snapshot"

    @pytest.mark.unit
    def test_validate_device_path_valid(self, manager):
        """REQ-STREAM-003: Test valid device path validation."""
        valid_paths = ["/dev/video0", "/dev/video1", "/dev/custom", "/dev/custom123"]
        for path in valid_paths:
            manager._validate_device_path(path)  # Should not raise exception

    @pytest.mark.unit
    def test_validate_device_path_invalid(self, manager):
        """REQ-STREAM-003: Test invalid device path validation."""
        invalid_paths = ["", None, "/dev/invalid", "/tmp/video0", "video0"]
        for path in invalid_paths:
            with pytest.raises(ValidationError):
                manager._validate_device_path(path)

    @pytest.mark.unit
    def test_validate_stream_name_valid(self, manager):
        """REQ-STREAM-003: Test valid stream name validation."""
        valid_names = ["camera0", "camera1_viewing", "custom_stream", "test-123"]
        for name in valid_names:
            manager._validate_stream_name(name)  # Should not raise exception

    @pytest.mark.unit
    def test_validate_stream_name_invalid(self, manager):
        """REQ-STREAM-003: Test invalid stream name validation."""
        invalid_names = ["", None, "camera@0", "camera 0", "camera.0"]
        for name in invalid_names:
            with pytest.raises(ValidationError):
                manager._validate_stream_name(name)

    @pytest.mark.unit
    def test_validate_use_case_valid(self, manager):
        """REQ-STREAM-002: Test valid use case validation."""
        valid_cases = [StreamUseCase.RECORDING, StreamUseCase.VIEWING, StreamUseCase.SNAPSHOT]
        for use_case in valid_cases:
            manager._validate_use_case(use_case)  # Should not raise exception

    @pytest.mark.unit
    def test_validate_use_case_invalid(self, manager):
        """REQ-STREAM-002: Test invalid use case validation."""
        with pytest.raises(ValidationError):
            manager._validate_use_case("invalid")
        with pytest.raises(ValidationError):
            manager._validate_use_case(None)

    @pytest.mark.unit
    def test_get_stream_name(self, manager):
        """REQ-STREAM-002: Test stream name generation for different devices and use cases."""
        assert manager._get_stream_name("/dev/video0", StreamUseCase.RECORDING) == "camera0"
        assert manager._get_stream_name("/dev/video1", StreamUseCase.VIEWING) == "camera1_viewing"
        assert manager._get_stream_name("/dev/video2", StreamUseCase.SNAPSHOT) == "camera2_snapshot"
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
    async def test_context_manager(self, mock_session):
        """REQ-STREAM-003: Test async context manager functionality."""
        with patch('aiohttp.ClientSession', return_value=mock_session):
            async with StreamLifecycleManager() as manager:
                assert manager._session == mock_session
                assert mock_session.__aenter__.called
            assert mock_session.__aexit__.called

    @pytest.mark.unit
    @pytest.mark.asyncio
    async def test_configure_mediamtx_path_recording(self, async_manager):
        """REQ-STREAM-001: Test MediaMTX path configuration for recording use case."""
        with patch.object(async_manager, "_configure_mediamtx_path_api") as mock_configure:
            mock_configure.return_value = None
            result = await async_manager.configure_mediamtx_path("/dev/video0", StreamUseCase.RECORDING)
            assert result is True
            mock_configure.assert_called_once()
            
            config_key = "/dev/video0:recording"
            assert config_key in async_manager._stream_configs
            config = async_manager._stream_configs[config_key]
            assert config.use_case == StreamUseCase.RECORDING
            assert config.run_on_demand_close_after == "0s"

    @pytest.mark.unit
    @pytest.mark.asyncio
    async def test_configure_mediamtx_path_viewing(self, async_manager):
        """REQ-STREAM-002: Test MediaMTX path configuration for viewing use case."""
        with patch.object(async_manager, "_configure_mediamtx_path_api") as mock_configure:
            mock_configure.return_value = None
            result = await async_manager.configure_mediamtx_path("/dev/video0", StreamUseCase.VIEWING)
            assert result is True
            
            config_key = "/dev/video0:viewing"
            config = async_manager._stream_configs[config_key]
            assert config.use_case == StreamUseCase.VIEWING
            assert config.run_on_demand_close_after == "300s"
            assert config.stream_name == "camera0_viewing"

    @pytest.mark.unit
    @pytest.mark.asyncio
    async def test_configure_mediamtx_path_snapshot(self, async_manager):
        """REQ-STREAM-002: Test MediaMTX path configuration for snapshot use case."""
        with patch.object(async_manager, "_configure_mediamtx_path_api") as mock_configure:
            mock_configure.return_value = None
            result = await async_manager.configure_mediamtx_path("/dev/video0", StreamUseCase.SNAPSHOT)
            assert result is True
            
            config_key = "/dev/video0:snapshot"
            config = async_manager._stream_configs[config_key]
            assert config.use_case == StreamUseCase.SNAPSHOT
            assert config.run_on_demand_close_after == "60s"
            assert config.stream_name == "camera0_snapshot"

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
                await async_manager.configure_mediamtx_path("/dev/video0", StreamUseCase.RECORDING)

    @pytest.mark.unit
    @pytest.mark.asyncio
    async def test_configure_mediamtx_path_api_success(self, async_manager):
        """REQ-STREAM-003: Test MediaMTX path configuration API call."""
        stream_config = StreamConfig(
            use_case=StreamUseCase.RECORDING,
            device_path="/dev/video0",
            stream_name="camera0",
            run_on_demand_close_after="0s",
            run_on_demand_restart=True,
            run_on_demand_start_timeout="10s",
            ffmpeg_command="ffmpeg -f v4l2 -i /dev/video0 -c:v libx264 -f rtsp rtsp://localhost:8554/camera0",
        )

        mock_response = AsyncMock()
        mock_response.status = 200
        async_manager._session.post.return_value.__aenter__.return_value = mock_response

        await async_manager._configure_mediamtx_path_api(stream_config)
        async_manager._session.post.assert_called_once()

    @pytest.mark.unit
    @pytest.mark.asyncio
    async def test_configure_mediamtx_path_api_failure(self, async_manager):
        """REQ-STREAM-003: Test MediaMTX path configuration API failure."""
        stream_config = StreamConfig(
            use_case=StreamUseCase.RECORDING,
            device_path="/dev/video0",
            stream_name="camera0",
            run_on_demand_close_after="0s",
            run_on_demand_restart=True,
            run_on_demand_start_timeout="10s",
            ffmpeg_command="ffmpeg -f v4l2 -i /dev/video0 -c:v libx264 -f rtsp rtsp://localhost:8554/camera0",
        )

        mock_response = AsyncMock()
        mock_response.status = 500
        mock_response.text = AsyncMock(return_value="Internal Server Error")
        async_manager._session.post.return_value.__aenter__.return_value = mock_response

        with pytest.raises(MediaMTXAPIError):
            await async_manager._configure_mediamtx_path_api(stream_config)

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
            assert "camera0" in async_manager._active_streams

    @pytest.mark.unit
    @pytest.mark.asyncio
    async def test_start_viewing_stream(self, async_manager):
        """REQ-STREAM-002: Test starting a viewing stream."""
        with patch.object(async_manager, "configure_mediamtx_path") as mock_configure, patch.object(
            async_manager, "_trigger_stream_activation"
        ) as mock_trigger:
            mock_configure.return_value = True
            mock_trigger.return_value = None

            result = await async_manager.start_viewing_stream("/dev/video0")

            assert result is True
            mock_configure.assert_called_once_with("/dev/video0", StreamUseCase.VIEWING)
            mock_trigger.assert_called_once_with("camera0_viewing")
            assert "camera0_viewing" in async_manager._active_streams

    @pytest.mark.unit
    @pytest.mark.asyncio
    async def test_start_snapshot_stream(self, async_manager):
        """REQ-STREAM-002: Test starting a snapshot stream."""
        with patch.object(async_manager, "configure_mediamtx_path") as mock_configure, patch.object(
            async_manager, "_trigger_stream_activation"
        ) as mock_trigger:
            mock_configure.return_value = True
            mock_trigger.return_value = None

            result = await async_manager.start_snapshot_stream("/dev/video0")

            assert result is True
            mock_configure.assert_called_once_with("/dev/video0", StreamUseCase.SNAPSHOT)
            mock_trigger.assert_called_once_with("camera0_snapshot")
            assert "camera0_snapshot" in async_manager._active_streams

    @pytest.mark.unit
    @pytest.mark.asyncio
    async def test_start_stream_invalid_device(self, async_manager):
        """REQ-STREAM-003: Test starting stream with invalid device path."""
        with pytest.raises(ValidationError):
            await async_manager.start_recording_stream("")

    @pytest.mark.unit
    @pytest.mark.asyncio
    async def test_start_stream_configuration_failure(self, async_manager):
        """REQ-STREAM-003: Test starting stream when configuration fails."""
        with patch.object(async_manager, "configure_mediamtx_path") as mock_configure:
            mock_configure.return_value = False
            result = await async_manager.start_recording_stream("/dev/video0")
            assert result is False

    @pytest.mark.unit
    @pytest.mark.asyncio
    async def test_trigger_stream_activation(self, async_manager):
        """REQ-STREAM-003: Test stream activation triggering."""
        mock_response = AsyncMock()
        mock_response.status = 404  # Expected for stream not ready yet
        async_manager._session.get.return_value.__aenter__.return_value = mock_response

        await async_manager._trigger_stream_activation("camera0")
        async_manager._session.get.assert_called_once()

    @pytest.mark.unit
    @pytest.mark.asyncio
    async def test_trigger_stream_activation_error(self, async_manager):
        """REQ-STREAM-003: Test stream activation triggering with error."""
        mock_response = AsyncMock()
        mock_response.status = 500
        mock_response.text = AsyncMock(return_value="Internal Server Error")
        async_manager._session.get.return_value.__aenter__.return_value = mock_response

        with pytest.raises(MediaMTXAPIError):
            await async_manager._trigger_stream_activation("camera0")

    @pytest.mark.unit
    @pytest.mark.asyncio
    async def test_stop_stream_recording(self, async_manager):
        """REQ-STREAM-001: Test stopping a recording stream (should keep it active for file rotation)."""
        async_manager._active_streams["camera0"] = {
            "device_path": "/dev/video0",
            "use_case": StreamUseCase.RECORDING,
            "start_time": 1000.0,
            "config": MagicMock(),
        }

        with patch.object(async_manager, "_stop_stream_api") as mock_stop:
            result = await async_manager.stop_stream("/dev/video0", StreamUseCase.RECORDING, "test")
            assert result is True
            mock_stop.assert_not_called()  # Recording streams should not be stopped via API
            assert "camera0" not in async_manager._active_streams

    @pytest.mark.unit
    @pytest.mark.asyncio
    async def test_stop_stream_viewing(self, async_manager):
        """REQ-STREAM-002: Test stopping a viewing stream (should stop via API)."""
        async_manager._active_streams["camera0_viewing"] = {
            "device_path": "/dev/video0",
            "use_case": StreamUseCase.VIEWING,
            "start_time": 1000.0,
            "config": MagicMock(),
        }

        with patch.object(async_manager, "_stop_stream_api") as mock_stop:
            result = await async_manager.stop_stream("/dev/video0", StreamUseCase.VIEWING, "test")
            assert result is True
            mock_stop.assert_called_once_with("camera0_viewing")
            assert "camera0_viewing" not in async_manager._active_streams

    @pytest.mark.unit
    @pytest.mark.asyncio
    async def test_stop_stream_snapshot(self, async_manager):
        """REQ-STREAM-002: Test stopping a snapshot stream (should stop via API)."""
        async_manager._active_streams["camera0_snapshot"] = {
            "device_path": "/dev/video0",
            "use_case": StreamUseCase.SNAPSHOT,
            "start_time": 1000.0,
            "config": MagicMock(),
        }

        with patch.object(async_manager, "_stop_stream_api") as mock_stop:
            result = await async_manager.stop_stream("/dev/video0", StreamUseCase.SNAPSHOT, "test")
            assert result is True
            mock_stop.assert_called_once_with("camera0_snapshot")
            assert "camera0_snapshot" not in async_manager._active_streams

    @pytest.mark.unit
    @pytest.mark.asyncio
    async def test_stop_stream_invalid_device(self, async_manager):
        """REQ-STREAM-003: Test stopping stream with invalid device path."""
        with pytest.raises(ValidationError):
            await async_manager.stop_stream("", StreamUseCase.RECORDING)

    @pytest.mark.unit
    @pytest.mark.asyncio
    async def test_stop_stream_api_success(self, async_manager):
        """REQ-STREAM-003: Test stopping stream via API."""
        mock_response = AsyncMock()
        mock_response.status = 200
        async_manager._session.post.return_value.__aenter__.return_value = mock_response

        await async_manager._stop_stream_api("camera0")
        async_manager._session.post.assert_called_once()

    @pytest.mark.unit
    @pytest.mark.asyncio
    async def test_stop_stream_api_not_found(self, async_manager):
        """REQ-STREAM-003: Test stopping stream via API when stream not found."""
        mock_response = AsyncMock()
        mock_response.status = 404  # Stream not found
        async_manager._session.post.return_value.__aenter__.return_value = mock_response

        await async_manager._stop_stream_api("camera0")  # Should not raise exception for 404

    @pytest.mark.unit
    @pytest.mark.asyncio
    async def test_stop_stream_api_error(self, async_manager):
        """REQ-STREAM-003: Test stopping stream via API with error."""
        mock_response = AsyncMock()
        mock_response.status = 500
        mock_response.text = AsyncMock(return_value="Internal Server Error")
        async_manager._session.post.return_value.__aenter__.return_value = mock_response

        with pytest.raises(MediaMTXAPIError):
            await async_manager._stop_stream_api("camera0")

    @pytest.mark.unit
    @pytest.mark.asyncio
    async def test_monitor_stream_health_ready(self, async_manager):
        """REQ-STREAM-003: Test monitoring stream health when stream is ready."""
        mock_response = AsyncMock()
        mock_response.status = 200
        mock_response.json = AsyncMock(return_value={"ready": True})
        async_manager._session.get.return_value.__aenter__.return_value = mock_response

        result = await async_manager.monitor_stream_health("/dev/video0", StreamUseCase.RECORDING)
        assert result is True

    @pytest.mark.unit
    @pytest.mark.asyncio
    async def test_monitor_stream_health_not_ready(self, async_manager):
        """REQ-STREAM-003: Test monitoring stream health when stream is not ready."""
        mock_response = AsyncMock()
        mock_response.status = 200
        mock_response.json = AsyncMock(return_value={"ready": False})
        async_manager._session.get.return_value.__aenter__.return_value = mock_response

        result = await async_manager.monitor_stream_health("/dev/video0", StreamUseCase.RECORDING)
        assert result is False

    @pytest.mark.unit
    @pytest.mark.asyncio
    async def test_monitor_stream_health_not_found(self, async_manager):
        """REQ-STREAM-003: Test monitoring stream health when stream not found."""
        mock_response = AsyncMock()
        mock_response.status = 404
        async_manager._session.get.return_value.__aenter__.return_value = mock_response

        result = await async_manager.monitor_stream_health("/dev/video0", StreamUseCase.RECORDING)
        assert result is False

    @pytest.mark.unit
    @pytest.mark.asyncio
    async def test_monitor_stream_health_invalid_device(self, async_manager):
        """REQ-STREAM-003: Test monitoring stream health with invalid device path."""
        with pytest.raises(ValidationError):
            await async_manager.monitor_stream_health("", StreamUseCase.RECORDING)

    @pytest.mark.unit
    @pytest.mark.asyncio
    async def test_monitor_stream_health_no_session(self, async_manager):
        """REQ-STREAM-003: Test monitoring stream health when session is not initialized."""
        async_manager._session = None
        result = await async_manager.monitor_stream_health("/dev/video0", StreamUseCase.RECORDING)
        assert result is False

    @pytest.mark.unit
    def test_get_active_streams(self, manager):
        """REQ-STREAM-004: Test getting active streams."""
        manager._active_streams = {
            "camera0": {"device_path": "/dev/video0", "use_case": StreamUseCase.RECORDING},
            "camera1_viewing": {"device_path": "/dev/video1", "use_case": StreamUseCase.VIEWING},
        }

        active_streams = manager.get_active_streams()
        assert len(active_streams) == 2
        assert "camera0" in active_streams
        assert "camera1_viewing" in active_streams

    @pytest.mark.unit
    def test_get_stream_config_existing(self, manager):
        """REQ-STREAM-002: Test getting existing stream configuration."""
        config = StreamConfig(
            use_case=StreamUseCase.RECORDING,
            device_path="/dev/video0",
            stream_name="camera0",
            run_on_demand_close_after="0s",
            run_on_demand_restart=True,
            run_on_demand_start_timeout="10s",
            ffmpeg_command="ffmpeg -f v4l2 -i /dev/video0 -c:v libx264 -f rtsp rtsp://localhost:8554/camera0",
        )
        manager._stream_configs["/dev/video0:recording"] = config

        retrieved_config = manager.get_stream_config("/dev/video0", StreamUseCase.RECORDING)
        assert retrieved_config is not None
        assert retrieved_config.use_case == StreamUseCase.RECORDING

    @pytest.mark.unit
    def test_get_stream_config_not_found(self, manager):
        """REQ-STREAM-002: Test getting non-existent stream configuration."""
        config = manager.get_stream_config("/dev/video0", StreamUseCase.RECORDING)
        assert config is None

    @pytest.mark.unit
    def test_get_stream_config_invalid_device(self, manager):
        """REQ-STREAM-003: Test getting stream config with invalid device path."""
        with pytest.raises(ValidationError):
            manager.get_stream_config("", StreamUseCase.RECORDING)

    @pytest.mark.unit
    @pytest.mark.asyncio
    async def test_cleanup(self, async_manager):
        """REQ-STREAM-004: Test cleanup of active streams."""
        async_manager._active_streams = {
            "camera0": {
                "device_path": "/dev/video0",
                "use_case": StreamUseCase.RECORDING,
                "start_time": 1000.0,
                "config": MagicMock(),
            },
            "camera1_viewing": {
                "device_path": "/dev/video1",
                "use_case": StreamUseCase.VIEWING,
                "start_time": 1000.0,
                "config": MagicMock(),
            },
            "camera2_snapshot": {
                "device_path": "/dev/video2",
                "use_case": StreamUseCase.SNAPSHOT,
                "start_time": 1000.0,
                "config": MagicMock(),
            },
        }

        with patch.object(async_manager, "stop_stream") as mock_stop:
            await async_manager.cleanup()
            assert mock_stop.call_count == 2  # Should stop viewing and snapshot, not recording

    @pytest.mark.unit
    @pytest.mark.asyncio
    async def test_cleanup_with_error(self, async_manager):
        """REQ-STREAM-004: Test cleanup when stopping streams fails."""
        async_manager._active_streams = {
            "camera1_viewing": {
                "device_path": "/dev/video1",
                "use_case": StreamUseCase.VIEWING,
                "start_time": 1000.0,
                "config": MagicMock(),
            },
        }

        with patch.object(async_manager, "stop_stream") as mock_stop:
            mock_stop.side_effect = Exception("Stop failed")
            await async_manager.cleanup()  # Should not raise exception

    @pytest.mark.unit
    def test_get_correlation_id_new(self, manager):
        """REQ-STREAM-004: Test correlation ID generation when none exists."""
        with patch("mediamtx_wrapper.stream_lifecycle_manager.get_correlation_id") as mock_get:
            mock_get.return_value = None
            correlation_id = manager._get_correlation_id()
            assert correlation_id.startswith("stream-lifecycle-")

    @pytest.mark.unit
    def test_get_correlation_id_existing(self, manager):
        """REQ-STREAM-004: Test correlation ID retrieval when one exists."""
        with patch("mediamtx_wrapper.stream_lifecycle_manager.get_correlation_id") as mock_get:
            mock_get.return_value = "existing-id-123"
            correlation_id = manager._get_correlation_id()
            assert correlation_id == "existing-id-123"

    @pytest.mark.unit
    @pytest.mark.asyncio
    async def test_validate_mediamtx_api_response_success(self, async_manager):
        """REQ-STREAM-003: Test MediaMTX API response validation for success."""
        mock_response = MagicMock()
        mock_response.status = 200
        await async_manager._validate_mediamtx_api_response(mock_response)  # Should not raise

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
