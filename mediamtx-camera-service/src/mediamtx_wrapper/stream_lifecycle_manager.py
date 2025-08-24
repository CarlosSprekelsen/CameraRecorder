"""
Stream Lifecycle Manager for MediaMTX Camera Service.

This module provides centralized stream lifecycle management for different use cases:
- Recording streams (long-duration, file rotation compatible)
- Viewing streams (live monitoring, auto-close after inactivity)
- Snapshot streams (quick capture, immediate deactivation)

The manager handles MediaMTX path configuration, stream activation/deactivation,
and health monitoring to ensure reliable operation across different scenarios.
"""

import re
import time
import uuid
from dataclasses import dataclass
from enum import Enum
from pathlib import Path
from typing import Dict, Optional, Any

import aiohttp

from camera_service.logging_config import set_correlation_id, get_correlation_id


class StreamUseCase(Enum):
    """Stream use case types for lifecycle management."""

    RECORDING = "recording"
    VIEWING = "viewing"
    SNAPSHOT = "snapshot"


@dataclass
class StreamConfig:
    """Configuration for a specific stream use case."""

    use_case: StreamUseCase
    device_path: str
    stream_name: str
    run_on_demand_close_after: str
    run_on_demand_restart: bool
    run_on_demand_start_timeout: str
    ffmpeg_command: str


class StreamLifecycleManagerError(Exception):
    """Base exception for StreamLifecycleManager errors."""

    pass


class ValidationError(StreamLifecycleManagerError):
    """Exception raised for validation errors."""

    pass


class MediaMTXAPIError(StreamLifecycleManagerError):
    """Exception raised for MediaMTX API errors."""

    pass


class StreamLifecycleManager:
    """
    Manages stream lifecycle for different use cases.

    This class provides centralized management of MediaMTX stream activation,
    deactivation, and configuration based on the specific use case requirements.
    It ensures file rotation compatibility for recording streams while maintaining
    power efficiency for other use cases.
    """

    def __init__(
        self,
        mediamtx_api_url: str = "http://localhost:9997",
        mediamtx_config_path: str = "/opt/mediamtx/config/mediamtx.yml",
        logger: Optional[Any] = None,
    ):
        """
        Initialize the Stream Lifecycle Manager.

        Args:
            mediamtx_api_url: MediaMTX API endpoint URL
            mediamtx_config_path: Path to MediaMTX configuration file
            logger: Logger instance for this manager
        """
        self.mediamtx_api_url = mediamtx_api_url.rstrip("/")
        self.mediamtx_config_path = Path(mediamtx_config_path)
        self.logger = logger or self._get_logger()

        # Stream configurations for different use cases
        self._stream_configs: Dict[str, StreamConfig] = {}
        self._active_streams: Dict[str, Dict[str, Any]] = {}
        self._session: Optional[aiohttp.ClientSession] = None

        # Use case specific configurations
        self._use_case_configs = {
            StreamUseCase.RECORDING: {
                "run_on_demand_close_after": "0s",  # Never auto-close for recording
                "run_on_demand_restart": True,
                "run_on_demand_start_timeout": "10s",
                "suffix": "",  # No suffix for recording streams
            },
            StreamUseCase.VIEWING: {
                "run_on_demand_close_after": "300s",  # 5 minutes after last viewer
                "run_on_demand_restart": True,
                "run_on_demand_start_timeout": "10s",
                "suffix": "_viewing",
            },
            StreamUseCase.SNAPSHOT: {
                "run_on_demand_close_after": "60s",  # 1 minute after capture
                "run_on_demand_restart": False,
                "run_on_demand_start_timeout": "5s",
                "suffix": "_snapshot",
            },
        }

    def _get_logger(self) -> Any:
        """Get logger instance using existing logging infrastructure."""
        import logging
        return logging.getLogger(__name__)

    def _get_correlation_id(self) -> str:
        """Get correlation ID for logging, generating one if not available."""
        correlation_id = get_correlation_id()
        if not correlation_id:
            correlation_id = f"stream-lifecycle-{str(uuid.uuid4())[:8]}"
            set_correlation_id(correlation_id)
        return correlation_id

    def _validate_device_path(self, device_path: str) -> None:
        """
        Validate device path format and accessibility.

        Args:
            device_path: Device path to validate

        Raises:
            ValidationError: If device path is invalid
        """
        if not device_path:
            raise ValidationError("Device path cannot be empty")

        if not isinstance(device_path, str):
            raise ValidationError("Device path must be a string")

        # Validate device path format
        device_pattern = r"^/dev/(video\d+|custom\w*)$"
        if not re.match(device_pattern, device_path):
            raise ValidationError(
                f"Invalid device path format: {device_path}. "
                "Must be /dev/video<N> or /dev/custom<name>"
            )

    def _validate_stream_name(self, stream_name: str) -> None:
        """
        Validate stream name format.

        Args:
            stream_name: Stream name to validate

        Raises:
            ValidationError: If stream name is invalid
        """
        if not stream_name:
            raise ValidationError("Stream name cannot be empty")

        if not isinstance(stream_name, str):
            raise ValidationError("Stream name must be a string")

        # Validate stream name format (alphanumeric, underscore, hyphen)
        stream_pattern = r"^[a-zA-Z0-9_-]+$"
        if not re.match(stream_pattern, stream_name):
            raise ValidationError(
                f"Invalid stream name format: {stream_name}. "
                "Must contain only alphanumeric characters, underscores, and hyphens"
            )

    def _validate_use_case(self, use_case: StreamUseCase) -> None:
        """
        Validate use case enum value.

        Args:
            use_case: Use case to validate

        Raises:
            ValidationError: If use case is invalid
        """
        if not isinstance(use_case, StreamUseCase):
            raise ValidationError("Use case must be a StreamUseCase enum value")

        if use_case not in self._use_case_configs:
            raise ValidationError(f"Unsupported use case: {use_case}")

    async def _validate_mediamtx_api_response(self, response: aiohttp.ClientResponse) -> None:
        """
        Validate MediaMTX API response.

        Args:
            response: HTTP response to validate

        Raises:
            MediaMTXAPIError: If response indicates an error
        """
        if response.status >= 400:
            try:
                error_text = await response.text()
                raise MediaMTXAPIError(
                    f"MediaMTX API error: {response.status} - {error_text}"
                )
            except Exception:
                raise MediaMTXAPIError(
                    f"MediaMTX API error: {response.status} - Unable to read error details"
                )

    async def __aenter__(self) -> "StreamLifecycleManager":
        """Async context manager entry."""
        self._session = aiohttp.ClientSession()
        return self

    async def __aexit__(self, exc_type: Any, exc_val: Any, exc_tb: Any) -> None:
        """Async context manager exit."""
        if self._session:
            await self._session.close()

    def _get_stream_name(self, device_path: str, use_case: StreamUseCase) -> str:
        """
        Generate stream name for the given device and use case.

        Args:
            device_path: Camera device path (e.g., /dev/video0)
            use_case: Stream use case type

        Returns:
            Stream name for MediaMTX path
        """
        # Extract device number from path
        device_name = Path(device_path).name
        if device_name.startswith("video"):
            device_num = device_name[5:]  # Remove "video" prefix
            base_name = f"camera{device_num}"
        else:
            base_name = device_name

        # Add use case suffix
        suffix = self._use_case_configs[use_case]["suffix"]
        stream_name = f"{base_name}{suffix}"
        
        # Validate generated stream name
        self._validate_stream_name(stream_name)
        return stream_name

    def _build_ffmpeg_command(self, device_path: str, stream_name: str) -> str:
        """
        Build FFmpeg command for camera stream.

        Args:
            device_path: Camera device path
            stream_name: MediaMTX stream name

        Returns:
            FFmpeg command string
        """
        return (
            f"ffmpeg -f v4l2 -i {device_path} "
            f"-c:v libx264 -preset ultrafast -tune zerolatency "
            f"-f rtsp rtsp://localhost:8554/{stream_name}"
        )

    async def configure_mediamtx_path(
        self, device_path: str, use_case: StreamUseCase
    ) -> bool:
        """
        Configure MediaMTX path for the specified use case.

        Args:
            device_path: Camera device path
            use_case: Stream use case type

        Returns:
            True if configuration was successful, False otherwise

        Raises:
            ValidationError: If inputs are invalid
            MediaMTXAPIError: If MediaMTX API call fails
        """
        # Input validation
        self._validate_device_path(device_path)
        self._validate_use_case(use_case)

        correlation_id = self._get_correlation_id()
        set_correlation_id(correlation_id)

        try:
            stream_name = self._get_stream_name(device_path, use_case)
            config = self._use_case_configs[use_case]

            # Build FFmpeg command
            ffmpeg_command = self._build_ffmpeg_command(device_path, stream_name)

            # Create stream configuration
            stream_config = StreamConfig(
                use_case=use_case,
                device_path=device_path,
                stream_name=stream_name,
                run_on_demand_close_after=config["run_on_demand_close_after"],
                run_on_demand_restart=config["run_on_demand_restart"],
                run_on_demand_start_timeout=config["run_on_demand_start_timeout"],
                ffmpeg_command=ffmpeg_command,
            )

            # Store configuration
            config_key = f"{device_path}:{use_case.value}"
            self._stream_configs[config_key] = stream_config

            # Configure MediaMTX path via API
            await self._configure_mediamtx_path_api(stream_config)

            self.logger.info(
                "Configured MediaMTX path for %s (%s): %s",
                device_path,
                use_case.value,
                stream_name,
                extra={"correlation_id": correlation_id},
            )
            return True

        except (ValidationError, MediaMTXAPIError) as e:
            self.logger.error(
                "Failed to configure MediaMTX path for %s (%s): %s",
                device_path,
                use_case.value,
                e,
                extra={"correlation_id": correlation_id},
            )
            raise
        except Exception as e:
            self.logger.error(
                "Unexpected error configuring MediaMTX path for %s (%s): %s",
                device_path,
                use_case.value,
                e,
                extra={"correlation_id": correlation_id},
            )
            return False

    async def _configure_mediamtx_path_api(self, stream_config: StreamConfig) -> None:
        """
        Configure MediaMTX path via API.

        Args:
            stream_config: Stream configuration to apply

        Raises:
            MediaMTXAPIError: If MediaMTX API call fails
        """
        if not self._session:
            raise RuntimeError("Session not initialized")

        # Prepare path configuration
        path_config = {
            "runOnDemand": stream_config.ffmpeg_command,
            "runOnDemandRestart": stream_config.run_on_demand_restart,
            "runOnDemandStartTimeout": stream_config.run_on_demand_start_timeout,
            "runOnDemandCloseAfter": stream_config.run_on_demand_close_after,
            "runOnUnDemand": "",
        }

        # Configure path via MediaMTX API
        url = f"{self.mediamtx_api_url}/v3/config/paths/add/{stream_config.stream_name}"

        async with self._session.post(url, json=path_config) as response:
            await self._validate_mediamtx_api_response(response)

    async def start_recording_stream(self, device_path: str) -> bool:
        """
        Start stream optimized for recording with file rotation.

        Args:
            device_path: Camera device path

        Returns:
            True if stream started successfully, False otherwise

        Raises:
            ValidationError: If device path is invalid
        """
        return await self._start_stream(device_path, StreamUseCase.RECORDING)

    async def start_viewing_stream(self, device_path: str) -> bool:
        """
        Start stream optimized for live viewing.

        Args:
            device_path: Camera device path

        Returns:
            True if stream started successfully, False otherwise

        Raises:
            ValidationError: If device path is invalid
        """
        return await self._start_stream(device_path, StreamUseCase.VIEWING)

    async def start_snapshot_stream(self, device_path: str) -> bool:
        """
        Start stream optimized for quick snapshot capture.

        Args:
            device_path: Camera device path

        Returns:
            True if stream started successfully, False otherwise

        Raises:
            ValidationError: If device path is invalid
        """
        return await self._start_stream(device_path, StreamUseCase.SNAPSHOT)

    async def _start_stream(self, device_path: str, use_case: StreamUseCase) -> bool:
        """
        Start stream for the specified use case.

        Args:
            device_path: Camera device path
            use_case: Stream use case type

        Returns:
            True if stream started successfully, False otherwise

        Raises:
            ValidationError: If inputs are invalid
        """
        # Input validation
        self._validate_device_path(device_path)
        self._validate_use_case(use_case)

        correlation_id = self._get_correlation_id()
        set_correlation_id(correlation_id)

        try:
            # Configure path if not already configured
            config_key = f"{device_path}:{use_case.value}"
            if config_key not in self._stream_configs:
                if not await self.configure_mediamtx_path(device_path, use_case):
                    return False

            stream_config = self._stream_configs[config_key]

            # Trigger stream activation via MediaMTX API
            await self._trigger_stream_activation(stream_config.stream_name)

            # Track active stream
            self._active_streams[stream_config.stream_name] = {
                "device_path": device_path,
                "use_case": use_case,
                "start_time": time.time(),
                "config": stream_config,
            }

            self.logger.info(
                "Started %s stream for %s: %s",
                use_case.value,
                device_path,
                stream_config.stream_name,
                extra={"correlation_id": correlation_id},
            )
            return True

        except ValidationError as e:
            self.logger.error(
                "Validation error starting %s stream for %s: %s",
                use_case.value,
                device_path,
                e,
                extra={"correlation_id": correlation_id},
            )
            raise
        except Exception as e:
            self.logger.error(
                "Failed to start %s stream for %s: %s",
                use_case.value,
                device_path,
                e,
                extra={"correlation_id": correlation_id},
            )
            return False

    async def _trigger_stream_activation(self, stream_name: str) -> None:
        """
        Trigger stream activation via MediaMTX API.

        Args:
            stream_name: Name of the stream to activate

        Raises:
            MediaMTXAPIError: If MediaMTX API call fails
        """
        if not self._session:
            raise RuntimeError("Session not initialized")

        # Trigger stream activation by making a request to the stream
        # This will cause MediaMTX to execute the runOnDemand command
        url = f"{self.mediamtx_api_url}/v3/paths/{stream_name}"

        async with self._session.get(url) as response:
            # 404 is expected if stream not ready yet, don't treat as error
            if response.status not in [200, 404]:
                await self._validate_mediamtx_api_response(response)

    async def stop_stream(
        self, device_path: str, use_case: StreamUseCase, reason: str = "manual"
    ) -> bool:
        """
        Stop stream with proper cleanup and logging.

        Args:
            device_path: Camera device path
            use_case: Stream use case type
            reason: Reason for stopping the stream

        Returns:
            True if stream stopped successfully, False otherwise

        Raises:
            ValidationError: If inputs are invalid
        """
        # Input validation
        self._validate_device_path(device_path)
        self._validate_use_case(use_case)

        correlation_id = self._get_correlation_id()
        set_correlation_id(correlation_id)

        try:
            stream_name = self._get_stream_name(device_path, use_case)

            # Remove from active streams
            if stream_name in self._active_streams:
                stream_info = self._active_streams.pop(stream_name)
                duration = time.time() - stream_info["start_time"]

                self.logger.info(
                    "Stopped %s stream for %s (%s): %s (duration: %.1fs)",
                    use_case.value,
                    device_path,
                    reason,
                    stream_name,
                    duration,
                    extra={"correlation_id": correlation_id},
                )

            # For recording streams, we don't actually stop the MediaMTX stream
            # as it should remain active during file rotation
            if use_case == StreamUseCase.RECORDING:
                self.logger.info(
                    "Recording stream %s kept active for file rotation compatibility",
                    stream_name,
                    extra={"correlation_id": correlation_id},
                )
                return True

            # For other use cases, we can stop the stream via API
            await self._stop_stream_api(stream_name)
            return True

        except ValidationError as e:
            self.logger.error(
                "Validation error stopping %s stream for %s: %s",
                use_case.value,
                device_path,
                e,
                extra={"correlation_id": correlation_id},
            )
            raise
        except Exception as e:
            self.logger.error(
                "Failed to stop %s stream for %s: %s",
                use_case.value,
                device_path,
                e,
                extra={"correlation_id": correlation_id},
            )
            return False

    async def _stop_stream_api(self, stream_name: str) -> None:
        """
        Stop stream via MediaMTX API.

        Args:
            stream_name: Name of the stream to stop

        Raises:
            MediaMTXAPIError: If MediaMTX API call fails
        """
        if not self._session:
            raise RuntimeError("Session not initialized")

        # Stop stream via MediaMTX API
        url = f"{self.mediamtx_api_url}/v3/paths/{stream_name}/stop"

        async with self._session.post(url) as response:
            # 404 if stream not found, don't treat as error
            if response.status not in [200, 404]:
                await self._validate_mediamtx_api_response(response)

    async def monitor_stream_health(
        self, device_path: str, use_case: StreamUseCase
    ) -> bool:
        """
        Monitor stream health during long operations.

        Args:
            device_path: Camera device path
            use_case: Stream use case type

        Returns:
            True if stream is healthy, False otherwise

        Raises:
            ValidationError: If inputs are invalid
        """
        # Input validation
        self._validate_device_path(device_path)
        self._validate_use_case(use_case)

        correlation_id = self._get_correlation_id()
        set_correlation_id(correlation_id)

        try:
            stream_name = self._get_stream_name(device_path, use_case)

            if not self._session:
                return False

            # Check stream status via MediaMTX API
            url = f"{self.mediamtx_api_url}/v3/paths/{stream_name}"

            async with self._session.get(url) as response:
                if response.status == 200:
                    data = await response.json()
                    is_ready = data.get("ready", False)

                    if is_ready:
                        self.logger.debug(
                            "Stream %s is healthy",
                            stream_name,
                            extra={"correlation_id": correlation_id},
                        )
                        return True
                    else:
                        self.logger.warning(
                            "Stream %s is not ready",
                            stream_name,
                            extra={"correlation_id": correlation_id},
                        )
                        return False
                else:
                    self.logger.warning(
                        "Stream %s not found or error",
                        stream_name,
                        extra={"correlation_id": correlation_id},
                    )
                    return False

        except ValidationError as e:
            self.logger.error(
                "Validation error monitoring stream health for %s: %s",
                device_path,
                e,
                extra={"correlation_id": correlation_id},
            )
            raise
        except Exception as e:
            self.logger.error(
                "Failed to monitor stream health for %s: %s",
                device_path,
                e,
                extra={"correlation_id": correlation_id},
            )
            return False

    def get_active_streams(self) -> Dict[str, Dict[str, Any]]:
        """
        Get information about currently active streams.

        Returns:
            Dictionary of active stream information
        """
        return self._active_streams.copy()

    def get_stream_config(
        self, device_path: str, use_case: StreamUseCase
    ) -> Optional[StreamConfig]:
        """
        Get stream configuration for the specified device and use case.

        Args:
            device_path: Camera device path
            use_case: Stream use case type

        Returns:
            Stream configuration if found, None otherwise

        Raises:
            ValidationError: If inputs are invalid
        """
        # Input validation
        self._validate_device_path(device_path)
        self._validate_use_case(use_case)

        config_key = f"{device_path}:{use_case.value}"
        return self._stream_configs.get(config_key)

    async def cleanup(self) -> None:
        """Clean up resources and stop all active streams."""
        correlation_id = self._get_correlation_id()
        set_correlation_id(correlation_id)

        try:
            # Stop all active streams (except recording streams)
            for stream_name, stream_info in list(self._active_streams.items()):
                use_case = stream_info["use_case"]
                device_path = stream_info["device_path"]

                if use_case != StreamUseCase.RECORDING:
                    await self.stop_stream(device_path, use_case, "cleanup")

            self.logger.info(
                "Stream lifecycle manager cleanup completed",
                extra={"correlation_id": correlation_id},
            )

        except Exception as e:
            self.logger.error(
                "Error during cleanup: %s",
                e,
                extra={"correlation_id": correlation_id},
            )
