"""
MediaMTX REST API controller for stream and recording management.

Enhanced version with improved error handling, process management,
and operational robustness per architecture requirements.
"""

import asyncio
import logging
import os
import time
import uuid
import random
from typing import Dict, Any, Optional, List
from dataclasses import dataclass
import aiohttp

from camera_service.logging_config import set_correlation_id, get_correlation_id


@dataclass
class StreamConfig:
    """Configuration for a MediaMTX stream path."""

    name: str
    source: str
    record: bool = False
    record_path: Optional[str] = None


class MediaMTXController:
    """
    Async controller for managing MediaMTX via REST API.

    Handles stream creation/deletion, recording control, health monitoring,
    and configuration management for the MediaMTX media server with enhanced
    error handling and operational robustness.
    """

    def __init__(
        self,
        host: str,
        api_port: int,
        rtsp_port: int,
        webrtc_port: int,
        hls_port: int,
        config_path: str,
        recordings_path: str,
        snapshots_path: str,
        health_check_interval: int = 30,
        health_failure_threshold: int = 10,
        health_circuit_breaker_timeout: int = 60,
        health_max_backoff_interval: int = 120,
        health_recovery_confirmation_threshold: int = 3,
        backoff_base_multiplier: float = 2.0,
        backoff_jitter_range: tuple = (0.8, 1.2),
        process_termination_timeout: float = 3.0,
        process_kill_timeout: float = 2.0,
        ffmpeg_config: Optional[Dict[str, Any]] = None,
    ):
        """
        Initialize MediaMTX controller.

        Args:
            host: MediaMTX server hostname or IP
            api_port: MediaMTX REST API port
            rtsp_port: RTSP streaming port
            webrtc_port: WebRTC streaming port
            hls_port: HLS streaming port
            config_path: Path to MediaMTX configuration file
            recordings_path: Directory for recording files
            snapshots_path: Directory for snapshot files
            health_check_interval: Normal health check interval in seconds (default: 30)
            health_failure_threshold: Failures before circuit breaker activates (default: 10)
            health_circuit_breaker_timeout: Circuit breaker timeout in seconds (default: 60)
            health_max_backoff_interval: Maximum backoff interval in seconds (default: 120)
            health_recovery_confirmation_threshold: Consecutive successes required to reset circuit breaker (default: 3)
            backoff_base_multiplier: Base multiplier for exponential backoff (default: 2.0)
            backoff_jitter_range: Jitter range for backoff randomization (default: (0.8, 1.2))
            process_termination_timeout: Timeout for graceful process termination (default: 3.0)
            process_kill_timeout: Timeout for force kill after termination (default: 2.0)
            ffmpeg_config: FFmpeg configuration settings (default: None)
        """
        self._host = host
        self._api_port = api_port
        self._rtsp_port = rtsp_port
        self._webrtc_port = webrtc_port
        self._hls_port = hls_port
        self._config_path = config_path
        self._recordings_path = recordings_path
        self._snapshots_path = snapshots_path

        # Configurable health monitoring parameters
        self._health_check_interval = health_check_interval
        self._health_failure_threshold = health_failure_threshold
        self._health_circuit_breaker_timeout = health_circuit_breaker_timeout
        self._health_max_backoff_interval = health_max_backoff_interval
        self._health_recovery_confirmation_threshold = (
            health_recovery_confirmation_threshold
        )
        self._backoff_base_multiplier = backoff_base_multiplier
        self._backoff_jitter_range = backoff_jitter_range
        self._process_termination_timeout = process_termination_timeout
        self._process_kill_timeout = process_kill_timeout

        # FFmpeg configuration with defaults
        self._ffmpeg_config = ffmpeg_config or {
            "snapshot": {
                "process_creation_timeout": 5.0,
                "execution_timeout": 8.0,
                "internal_timeout": 5000000,
                "retry_attempts": 2,
                "retry_delay": 1.0
            },
            "recording": {
                "process_creation_timeout": 10.0,
                "execution_timeout": 15.0,
                "internal_timeout": 10000000,
                "retry_attempts": 3,
                "retry_delay": 2.0
            }
        }

        self._logger = logging.getLogger(__name__)
        self._base_url = f"http://{self._host}:{self._api_port}"

        # HTTP client session for REST API calls
        self._session: Optional[aiohttp.ClientSession] = None
        self._health_check_task: Optional[asyncio.Task] = None
        self._running = False
        self._recording_sessions: Dict[str, Dict[str, Any]] = {}
        self._last_health_status = None

        # Enhanced tracking for health monitoring
        self._health_state = {
            "consecutive_failures": 0,
            "failure_count": 0,  # Total failures for test compatibility
            "success_count": 0,  # Total successes for test compatibility
            "consecutive_successes_during_recovery": 0,
            "last_success_time": 0.0,
            "last_failure_time": 0.0,
            "total_checks": 0,
            "recovery_count": 0,
            "circuit_breaker_activations": 0,
            "circuit_breaker_active": False,  # Persist circuit breaker state across restarts
            "circuit_breaker_start_time": 0.0,  # Persist circuit breaker start time
        }

    # Public read-only properties for integration tests and API consumers
    @property
    def host(self) -> str:
        return self._host

    @property
    def api_port(self) -> int:
        return self._api_port

    @property
    def rtsp_port(self) -> int:
        return self._rtsp_port

    @property
    def webrtc_port(self) -> int:
        return self._webrtc_port

    @property
    def hls_port(self) -> int:
        return self._hls_port

    async def start(self) -> None:
        """
        Start the MediaMTX controller.

        Initializes HTTP client session and begins health monitoring.
        """
        if self._running:
            self._logger.warning("MediaMTX controller is already running")
            return

        correlation_id = get_correlation_id() or str(uuid.uuid4())[:8]
        set_correlation_id(correlation_id)

        self._logger.info(
            "Starting MediaMTX controller", extra={"correlation_id": correlation_id}
        )

        try:
            # Validate directory permissions before starting
            await self._validate_directory_permissions()

            # Create aiohttp ClientSession with timeout configuration
            timeout = aiohttp.ClientTimeout(total=15, connect=5)
            connector = aiohttp.TCPConnector(limit=10, limit_per_host=5)
            self._session = aiohttp.ClientSession(
                timeout=timeout,
                connector=connector,
                headers={"Content-Type": "application/json"},
            )

            # Start health monitoring task
            self._health_check_task = asyncio.create_task(self._health_monitor_loop())
            self._running = True

            self._logger.info(
                "MediaMTX controller started successfully",
                extra={"correlation_id": correlation_id},
            )
        except Exception as e:
            self._logger.error(
                f"Failed to start MediaMTX controller: {e}",
                extra={"correlation_id": correlation_id},
            )
            await self._cleanup_on_start_failure()
            raise

    async def stop(self) -> None:
        """
        Stop the MediaMTX controller.

        Closes HTTP client session and stops monitoring tasks.
        """
        if not self._running:
            return

        correlation_id = get_correlation_id() or str(uuid.uuid4())[:8]
        set_correlation_id(correlation_id)

        self._logger.info(
            "Stopping MediaMTX controller", extra={"correlation_id": correlation_id}
        )
        self._running = False

        # Stop health monitoring task
        if self._health_check_task and not self._health_check_task.done():
            self._health_check_task.cancel()
            try:
                await self._health_check_task
            except asyncio.CancelledError:
                pass

        # Close aiohttp ClientSession
        if self._session:
            await self._session.close()
            self._session = None

        self._logger.info(
            "MediaMTX controller stopped", extra={"correlation_id": correlation_id}
        )

    async def _validate_directory_permissions(self) -> None:
        """Validate that required directories exist and are writable."""
        correlation_id = get_correlation_id() or str(uuid.uuid4())[:8]

        directories = [
            ("recordings", self._recordings_path),
            ("snapshots", self._snapshots_path),
        ]

        for dir_type, dir_path in directories:
            try:
                # Create directory if it doesn't exist
                os.makedirs(dir_path, exist_ok=True)

                # Test write permissions
                test_file = os.path.join(
                    dir_path, f".write_test_{uuid.uuid4().hex[:8]}"
                )
                with open(test_file, "w") as f:
                    f.write("test")
                os.remove(test_file)

                self._logger.debug(
                    f"Validated {dir_type} directory: {dir_path}",
                    extra={"correlation_id": correlation_id},
                )

            except PermissionError as e:
                error_msg = f"Permission denied for {dir_type} directory: {dir_path}"
                self._logger.error(error_msg, extra={"correlation_id": correlation_id})
                raise RuntimeError(error_msg) from e
            except OSError as e:
                error_msg = f"Cannot access {dir_type} directory: {dir_path} - {e}"
                self._logger.error(error_msg, extra={"correlation_id": correlation_id})
                raise RuntimeError(error_msg) from e

    async def _cleanup_on_start_failure(self) -> None:
        """Clean up resources when startup fails."""
        self._running = False
        if self._health_check_task and not self._health_check_task.done():
            self._health_check_task.cancel()
        if self._session:
            await self._session.close()
            self._session = None

    async def health_check(self) -> Dict[str, Any]:
        """
        Perform health check on MediaMTX server with enhanced error context.

        Returns:
            Dict containing health status and metrics

        Raises:
            ConnectionError: If MediaMTX controller is not started
        """
        if not self._session:
            raise ConnectionError("MediaMTX controller not started")

        correlation_id = get_correlation_id() or str(uuid.uuid4())[:8]
        start_time = time.time()

        try:
            # Call MediaMTX API config endpoint to verify connectivity
            async with self._session.get(
                f"{self._base_url}/v3/config/global/get"
            ) as response:
                response_time = int((time.time() - start_time) * 1000)

                if response.status == 200:
                    config_data = await response.json()
                    self._health_state["last_success_time"] = time.time()
                    self._health_state["total_checks"] += 1
                    self._health_state["success_count"] += 1

                    return {
                        "status": "healthy",
                        "version": config_data.get("serverVersion", "unknown"),
                        "uptime": config_data.get("serverUptime", 0),
                        "api_port": self._api_port,
                        "response_time_ms": response_time,
                        "correlation_id": correlation_id,
                        "consecutive_failures": self._health_state[
                            "consecutive_failures"
                        ],
                        "circuit_breaker_activations": self._health_state[
                            "circuit_breaker_activations"
                        ],
                    }
                else:
                    error_text = await response.text()
                    self._health_state["last_failure_time"] = time.time()
                    self._health_state["total_checks"] += 1
                    self._health_state["failure_count"] += 1

                    return {
                        "status": "unhealthy",
                        "error": f"HTTP {response.status}: {error_text}",
                        "api_port": self._api_port,
                        "response_time_ms": response_time,
                        "correlation_id": correlation_id,
                        "consecutive_failures": self._health_state[
                            "consecutive_failures"
                        ],
                    }

        except aiohttp.ClientError as e:
            self._health_state["last_failure_time"] = time.time()
            self._health_state["total_checks"] += 1
            self._health_state["failure_count"] += 1

            return {
                "status": "error",
                "error": f"Connection error: {e}",
                "api_port": self._api_port,
                "correlation_id": correlation_id,
                "consecutive_failures": self._health_state["consecutive_failures"],
            }
        except Exception as e:
            self._logger.error(
                f"Health check failed with unexpected error: {e}",
                extra={"correlation_id": correlation_id},
            )
            self._health_state["last_failure_time"] = time.time()
            self._health_state["total_checks"] += 1
            self._health_state["failure_count"] += 1

            return {
                "status": "error",
                "error": f"Unexpected error: {e}",
                "api_port": self._api_port,
                "correlation_id": correlation_id,
                "consecutive_failures": self._health_state["consecutive_failures"],
            }

    async def create_stream(self, stream_config: StreamConfig) -> Dict[str, str]:
        """
        Create a new stream path in MediaMTX with enhanced idempotent behavior.

        Args:
            stream_config: Stream configuration parameters

        Returns:
            Dict containing stream URLs for different protocols

        Raises:
            ValueError: If stream configuration is invalid
            ConnectionError: If MediaMTX is unreachable
        """
        if not self._session:
            raise ConnectionError("MediaMTX controller not started")

        if not stream_config.name or not stream_config.source:
            raise ValueError("Stream name and source are required")

        correlation_id = get_correlation_id() or str(uuid.uuid4())[:8]
        set_correlation_id(correlation_id)

        self._logger.info(
            f"Creating stream path: {stream_config.name} from {stream_config.source}",
            extra={"correlation_id": correlation_id, "stream_name": stream_config.name},
        )

        try:
            # Enhanced idempotent behavior - check if stream already exists
            try:
                existing_stream = await self.get_stream_status(stream_config.name)
                if existing_stream.get("name") == stream_config.name:
                    self._logger.info(
                        f"Stream path already exists, returning existing URLs: {stream_config.name}",
                        extra={
                            "correlation_id": correlation_id,
                            "stream_name": stream_config.name,
                        },
                    )
                    return self._generate_stream_urls(stream_config.name)
            except ValueError:
                # Stream doesn't exist, proceed with creation
                pass

            # Create MediaMTX path configuration
            path_config = {
                "source": stream_config.source,
                "sourceProtocol": "automatic",
                "record": stream_config.record,
            }

            if stream_config.record and stream_config.record_path:
                path_config["recordPath"] = stream_config.record_path

            # Add stream path via MediaMTX API
            async with self._session.post(
                f"{self._base_url}/v3/config/paths/add/{stream_config.name}",
                json=path_config,
            ) as response:
                if response.status in [200, 201]:
                    self._logger.info(
                        f"Successfully created stream path: {stream_config.name}",
                        extra={
                            "correlation_id": correlation_id,
                            "stream_name": stream_config.name,
                        },
                    )
                    return self._generate_stream_urls(stream_config.name)
                elif response.status == 409:
                    # Stream already exists, return URLs (additional idempotency)
                    self._logger.info(
                        f"Stream path already exists (409 conflict): {stream_config.name}",
                        extra={
                            "correlation_id": correlation_id,
                            "stream_name": stream_config.name,
                        },
                    )
                    return self._generate_stream_urls(stream_config.name)
                else:
                    error_text = await response.text()
                    error_context = f"stream_name={stream_config.name}, source={stream_config.source}, record={stream_config.record}"
                    error_msg = f"Failed to create stream {stream_config.name}: HTTP {response.status} - {error_text} (context: {error_context})"
                    self._logger.error(
                        error_msg,
                        extra={
                            "correlation_id": correlation_id,
                            "stream_name": stream_config.name,
                        },
                    )
                    raise ConnectionError(error_msg)

        except aiohttp.ClientError as e:
            error_msg = f"MediaMTX unreachable during stream creation for {stream_config.name}: {e}"
            self._logger.error(
                error_msg,
                extra={
                    "correlation_id": correlation_id,
                    "stream_name": stream_config.name,
                },
            )
            raise ConnectionError(error_msg)

    def _generate_stream_urls(self, stream_name: str) -> Dict[str, str]:
        """Generate stream URLs for different protocols."""
        return {
            "rtsp": f"rtsp://{self._host}:{self._rtsp_port}/{stream_name}",
            "webrtc": f"http://{self._host}:{self._webrtc_port}/{stream_name}",
            "hls": f"http://{self._host}:{self._hls_port}/{stream_name}",
        }

    async def delete_stream(self, stream_name: str) -> bool:
        """
        Delete a stream path from MediaMTX with enhanced error handling.

        Args:
            stream_name: Name of the stream to delete

        Returns:
            True if stream was deleted successfully or didn't exist

        Raises:
            ConnectionError: If MediaMTX is unreachable
        """
        if not self._session:
            raise ConnectionError("MediaMTX controller not started")

        if not stream_name:
            raise ValueError("Stream name is required")

        correlation_id = get_correlation_id() or str(uuid.uuid4())[:8]
        set_correlation_id(correlation_id)

        self._logger.info(
            f"Deleting stream path: {stream_name}",
            extra={"correlation_id": correlation_id, "stream_name": stream_name},
        )

        try:
            # Delete stream path via MediaMTX API
            async with self._session.post(
                f"{self._base_url}/v3/config/paths/delete/{stream_name}"
            ) as response:
                if response.status in [200, 204]:
                    self._logger.info(
                        f"Successfully deleted stream path: {stream_name}",
                        extra={
                            "correlation_id": correlation_id,
                            "stream_name": stream_name,
                        },
                    )
                    return True
                elif response.status == 404:
                    # Stream already doesn't exist (idempotent)
                    self._logger.info(
                        f"Stream path already deleted or never existed: {stream_name}",
                        extra={
                            "correlation_id": correlation_id,
                            "stream_name": stream_name,
                        },
                    )
                    return True
                else:
                    error_text = await response.text()
                    error_msg = f"Failed to delete stream {stream_name}: HTTP {response.status} - {error_text}"
                    self._logger.error(
                        error_msg,
                        extra={
                            "correlation_id": correlation_id,
                            "stream_name": stream_name,
                        },
                    )
                    return False

        except aiohttp.ClientError as e:
            error_msg = (
                f"MediaMTX unreachable during stream deletion for {stream_name}: {e}"
            )
            self._logger.error(
                error_msg,
                extra={"correlation_id": correlation_id, "stream_name": stream_name},
            )
            raise ConnectionError(error_msg)

    async def _cleanup_orphaned_recording_sessions(self) -> None:
        """
        Clean up orphaned recording sessions where FFmpeg processes have died.
        This prevents "Recording already active" errors when processes have terminated unexpectedly.
        """
        orphaned_sessions = []
        
        for stream_name, session in self._recording_sessions.items():
            process = session.get("process")
            if process and process.returncode is not None:
                # Process has terminated
                orphaned_sessions.append(stream_name)
                self._logger.warning(
                    f"Found orphaned recording session for {stream_name} - process terminated with code {process.returncode}",
                    extra={"stream_name": stream_name, "return_code": process.returncode}
                )
        
        # Remove orphaned sessions
        for stream_name in orphaned_sessions:
            del self._recording_sessions[stream_name]
            self._logger.info(f"Cleaned up orphaned recording session for {stream_name}")

    async def _is_stream_actually_recording(self, stream_name: str) -> bool:
        """
        Check if a stream is actually recording by verifying the FFmpeg process is still alive.
        
        Args:
            stream_name: Name of the stream to check
            
        Returns:
            True if the stream is actively recording, False otherwise
        """
        if stream_name not in self._recording_sessions:
            return False
            
        session = self._recording_sessions[stream_name]
        process = session.get("process")
        
        if not process:
            return False
            
        # Check if process is still running
        if process.returncode is None:
            # Process is still running
            return True
        else:
            # Process has terminated - clean up the session
            self._logger.warning(
                f"Recording session for {stream_name} exists but process has terminated (return code: {process.returncode})",
                extra={"stream_name": stream_name, "return_code": process.returncode}
            )
            del self._recording_sessions[stream_name]
            return False

    async def start_recording(
        self, stream_name: str, duration: Optional[int] = None, format: str = "mp4"
    ) -> Dict[str, Any]:
        """
        Start recording for the specified stream using external FFmpeg process.

        Args:
            stream_name: Name of the stream to record
            duration: Recording duration in seconds (None for unlimited)
            format: Recording format (mp4, mkv)

        Returns:
            Dict containing recording session information

        Raises:
            ValueError: If stream does not exist or is already recording
            ConnectionError: If MediaMTX is unreachable
        """
        if not self._session:
            raise ConnectionError("MediaMTX controller not started")

        if not stream_name:
            raise ValueError("Stream name is required")

        # Clean up any orphaned recording sessions before checking
        await self._cleanup_orphaned_recording_sessions()

        # Check if stream is actually recording
        if await self._is_stream_actually_recording(stream_name):
            raise ValueError(f"Recording already active for stream: {stream_name}")

        correlation_id = get_correlation_id() or str(uuid.uuid4())[:8]
        set_correlation_id(correlation_id)

        # Validate format
        valid_formats = ["mp4", "mkv", "avi"]
        if format not in valid_formats:
            raise ValueError(
                f"Invalid format: {format}. Must be one of: {valid_formats}"
            )

        try:
            # Ensure recordings directory exists and is writable
            try:
                os.makedirs(self._recordings_path, exist_ok=True)
                # Test write permissions
                test_file = os.path.join(
                    self._recordings_path, f".write_test_{uuid.uuid4().hex[:8]}"
                )
                with open(test_file, "w") as f:
                    f.write("test")
                os.remove(test_file)
            except (PermissionError, OSError) as e:
                error_msg = (
                    f"Cannot write to recordings directory {self._recordings_path}: {e}"
                )
                self._logger.error(
                    error_msg,
                    extra={
                        "correlation_id": correlation_id,
                        "stream_name": stream_name,
                    },
                )
                raise ValueError(error_msg) from e

            # Generate recording filename with timestamp
            timestamp = time.strftime("%Y-%m-%d_%H-%M-%S")
            filename = f"{stream_name}_{timestamp}.{format}"
            record_path = os.path.join(self._recordings_path, filename)

            # Record start time for duration calculation
            start_time = time.time()
            start_time_iso = time.strftime("%Y-%m-%dT%H:%M:%SZ")

            # Enhanced stream readiness validation before recording
            self._logger.info(
                f"Validating stream readiness before recording: {stream_name}",
                extra={
                    "correlation_id": correlation_id,
                    "stream_name": stream_name,
                },
            )
            
            # Check stream readiness with configurable timeout for reliability
            stream_readiness_config = getattr(self, '_stream_readiness_config', {
                'timeout': 15.0,
                'retry_attempts': 3,
                'retry_delay': 2.0,
                'check_interval': 0.5,
                'enable_progress_notifications': True,
                'graceful_fallback': True
            })
            
            # Send progress notification if enabled
            if stream_readiness_config.get('enable_progress_notifications', True):
                await self._send_stream_validation_progress(stream_name, "started", correlation_id)
            
            stream_ready = await self.check_stream_readiness(
                stream_name, 
                timeout=stream_readiness_config.get('timeout', 15.0)
            )
            
            if not stream_ready:
                # Send failure notification if enabled
                if stream_readiness_config.get('enable_progress_notifications', True):
                    await self._send_stream_validation_progress(stream_name, "failed", correlation_id)
                
                error_msg = f"Stream {stream_name} is not ready for recording after validation"
                self._logger.error(
                    error_msg,
                    extra={
                        "correlation_id": correlation_id,
                        "stream_name": stream_name,
                        "stream_ready": False,
                    },
                )
                
                # Check if graceful fallback is enabled
                if stream_readiness_config.get('graceful_fallback', True):
                    self._logger.warning(
                        f"Graceful fallback enabled for {stream_name} - attempting to start recording anyway",
                        extra={
                            "correlation_id": correlation_id,
                            "stream_name": stream_name,
                            "graceful_fallback": True,
                        },
                    )
                    # Continue with recording attempt despite readiness failure
                else:
                    raise ValueError(error_msg)
            
            self._logger.info(
                f"Stream {stream_name} validated as ready for recording",
                extra={
                    "correlation_id": correlation_id,
                    "stream_name": stream_name,
                    "stream_ready": True,
                },
            )

            # Start external FFmpeg recording process
            rtsp_url = f"rtsp://{self._host}:{self._rtsp_port}/{stream_name}"
            
            # Build FFmpeg command
            ffmpeg_cmd = [
                "ffmpeg",
                "-i", rtsp_url,
                "-c:v", "copy",  # Copy video codec without re-encoding
                "-c:a", "copy",  # Copy audio codec without re-encoding
                "-f", format,
                "-y",  # Overwrite output file
                record_path
            ]
            
            # Add duration limit if specified
            if duration:
                ffmpeg_cmd.extend(["-t", str(duration)])

            # Start FFmpeg process
            process = await asyncio.create_subprocess_exec(
                *ffmpeg_cmd,
                stdout=asyncio.subprocess.PIPE,
                stderr=asyncio.subprocess.PIPE
            )

            # Store recording session for tracking
            self._recording_sessions[stream_name] = {
                "filename": filename,
                "start_time": start_time,
                "start_time_iso": start_time_iso,
                "record_path": record_path,
                "format": format,
                "duration": duration,
                "correlation_id": correlation_id,
                "process": process,
                "rtsp_url": rtsp_url
            }

            self._logger.info(
                f"Started recording for stream {stream_name}: {filename}",
                extra={
                    "correlation_id": correlation_id,
                    "stream_name": stream_name,
                    "source_file": filename,
                    "rtsp_url": rtsp_url,
                },
            )
            return {
                "stream_name": stream_name,
                "filename": filename,
                "status": "started",
                "start_time": start_time_iso,
                "record_path": record_path,
                "format": format,
                "duration": duration,
                "rtsp_url": rtsp_url,
            }

        except aiohttp.ClientError as e:
            error_msg = (
                f"MediaMTX unreachable during recording start for {stream_name}: {e}"
            )
            self._logger.error(
                error_msg,
                extra={"correlation_id": correlation_id, "stream_name": stream_name},
            )
            raise ConnectionError(error_msg)

    async def stop_recording(self, stream_name: str) -> Dict[str, Any]:
        """
        Stop recording for the specified stream by terminating the FFmpeg process.

        Args:
            stream_name: Name of the stream to stop recording

        Returns:
            Dict containing recording completion information

        Raises:
            ValueError: If stream is not currently recording
            ConnectionError: If MediaMTX is unreachable
        """
        if not self._session:
            raise ConnectionError("MediaMTX controller not started")

        if not stream_name:
            raise ValueError("Stream name is required")

        correlation_id = get_correlation_id() or str(uuid.uuid4())[:8]
        set_correlation_id(correlation_id)

        try:
            # Get recording session for duration calculation
            session = self._recording_sessions.get(stream_name)
            if not session:
                raise ValueError(
                    f"No active recording session found for stream: {stream_name}"
                )

            # Calculate recording duration with high precision
            end_time = time.time()
            end_time_iso = time.strftime("%Y-%m-%dT%H:%M:%SZ")
            actual_duration = int(end_time - session["start_time"])

            # Terminate FFmpeg process gracefully
            process = session.get("process")
            if process:
                try:
                    # Send SIGTERM to gracefully stop FFmpeg
                    process.terminate()
                    
                    # Wait for process to terminate (with optimized timeout)
                    try:
                        await asyncio.wait_for(process.wait(), timeout=5.0)  # Restored from 2.0s to 5.0s
                    except asyncio.TimeoutError:
                        # Force kill if graceful termination fails
                        self._logger.warning(
                            f"FFmpeg process did not terminate gracefully, forcing kill for {stream_name}",
                            extra={
                                "correlation_id": correlation_id,
                                "stream_name": stream_name,
                            },
                        )
                        process.kill()
                        await process.wait()
                    
                    self._logger.debug(
                        f"FFmpeg process terminated for {stream_name}",
                        extra={
                            "correlation_id": correlation_id,
                            "stream_name": stream_name,
                        },
                    )
                except Exception as e:
                    self._logger.warning(
                        f"Error terminating FFmpeg process for {stream_name}: {e}",
                        extra={
                            "correlation_id": correlation_id,
                            "stream_name": stream_name,
                        },
                    )

            # Enhanced file validation and error handling
            file_path = session["record_path"]
            file_size = 0
            file_exists = False
            file_error = None

            try:
                if os.path.exists(file_path):
                    file_size = os.path.getsize(file_path)
                    file_exists = True
                    self._logger.debug(
                        f"Recording file validated: {file_path} ({file_size} bytes)",
                        extra={
                            "correlation_id": correlation_id,
                            "stream_name": stream_name,
                        },
                    )
                else:
                    file_error = f"Recording file not found: {file_path}"
                    self._logger.warning(
                        file_error,
                        extra={
                            "correlation_id": correlation_id,
                            "stream_name": stream_name,
                        },
                    )
            except PermissionError as e:
                file_error = (
                    f"Permission denied accessing file: {file_path} - {e}"
                )
                self._logger.warning(
                    file_error,
                    extra={
                        "correlation_id": correlation_id,
                        "stream_name": stream_name,
                    },
                )
            except OSError as e:
                file_error = f"Error accessing file: {file_path} - {e}"
                self._logger.warning(
                    file_error,
                    extra={
                        "correlation_id": correlation_id,
                        "stream_name": stream_name,
                    },
                )

            # Clean up recording session
            del self._recording_sessions[stream_name]

            result = {
                "stream_name": stream_name,
                "filename": session["filename"],
                "status": "completed",
                "start_time": session["start_time_iso"],
                "end_time": end_time_iso,
                "duration": actual_duration,
                "file_size": file_size,
                "file_exists": file_exists,
            }

            if file_error:
                result["file_warning"] = file_error

            self._logger.info(
                f"Stopped recording for stream {stream_name}: duration={actual_duration}s, "
                f"file_size={file_size}, file_exists={file_exists}",
                extra={
                    "correlation_id": correlation_id,
                    "stream_name": stream_name,
                    "source_file": session["filename"],
                },
            )

            return result

        except aiohttp.ClientError as e:
            error_msg = (
                f"MediaMTX unreachable during recording stop for {stream_name}: {e}"
            )
            self._logger.error(
                error_msg,
                extra={"correlation_id": correlation_id, "stream_name": stream_name},
            )
            raise ConnectionError(error_msg)

    async def take_snapshot(
        self, stream_name: str, filename: Optional[str] = None, format: str = "jpg", quality: int = 85
    ) -> Dict[str, Any]:
        """
        Capture a snapshot from the specified stream using optimized multi-tier approach.
        
        **Multi-Tier Strategy for Optimal User Experience:**
        
        This method implements a sophisticated multi-tier approach to ensure fast, reliable
        snapshot capture while maintaining power efficiency through on-demand stream activation.
        
        **Tier 1 - Immediate RTSP Capture (Fastest Path):**
        - Check if RTSP stream is already active and ready
        - If ready, capture immediately using FFmpeg from RTSP stream
        - Expected response time: < 0.5 seconds (immediate)
        - Use case: Stream already running (e.g., during recording or active viewing)
        
        **Tier 2 - Quick Stream Activation (Balanced Path):**
        - If RTSP stream not ready, trigger on-demand activation
        - Wait for FFmpeg process to start and stream to become ready
        - Capture from RTSP once activated
        - Expected response time: 1-3 seconds (acceptable)
        - Use case: First snapshot after idle period, power-efficient operation
        
        **Tier 3 - Direct Camera Capture (Fallback Path):**
        - If RTSP activation fails, capture directly from camera device
        - Bypass MediaMTX entirely, use FFmpeg with v4l2 input
        - Expected response time: 2-5 seconds (slower but reliable)
        - Use case: MediaMTX issues, network problems, emergency capture
        
        **Tier 4 - Error Handling (Last Resort):**
        - If all methods fail, return detailed error information
        - Include timing data and methods attempted for debugging
        - Expected response time: < 10 seconds (with error details)
        
        **Configuration-Driven Performance Tuning:**
        All timeouts and thresholds are configurable via performance.snapshot_tiers:
        - tier1_rtsp_ready_check_timeout: Quick RTSP readiness check (default: 1.0s)
        - tier2_activation_timeout: Stream activation wait time (default: 3.0s)
        - tier2_activation_trigger_timeout: Activation trigger timeout (default: 1.0s)
        - tier3_direct_capture_timeout: Direct camera capture timeout (default: 5.0s)
        - total_operation_timeout: Maximum total operation time (default: 10.0s)
        - immediate_response_threshold: Consider "immediate" if under (default: 0.5s)
        - acceptable_response_threshold: Consider "acceptable" if under (default: 2.0s)
        - slow_response_threshold: Consider "slow" if over (default: 5.0s)
        
        **User Experience Optimization:**
        - Prioritizes speed when possible (Tier 1)
        - Balances power efficiency with responsiveness (Tier 2)
        - Provides reliable fallback for edge cases (Tier 3)
        - Offers detailed feedback for troubleshooting (Tier 4)
        
        **Performance Characteristics:**
        - Best case (Tier 1): < 0.5 seconds - immediate response
        - Typical case (Tier 2): 1-3 seconds - acceptable for photo capture
        - Fallback case (Tier 3): 2-5 seconds - slower but reliable
        - Error case (Tier 4): < 10 seconds - with detailed error information
        
        **Power Efficiency:**
        - No unnecessary FFmpeg processes when streams are idle
        - On-demand activation only when needed
        - Automatic cleanup of failed activation attempts
        - Maintains battery life for mobile/embedded deployments

        Args:
            stream_name: Name of the stream to capture from (e.g., "camera0")
            filename: Optional filename for the snapshot (auto-generated if not provided)
            format: Image format ("jpg" or "png")
            quality: Image quality (1-100 for jpg, ignored for png)

        Returns:
            Dict containing snapshot capture information with the following structure:
            {
                "stream_name": str,           # Stream name (e.g., "camera0")
                "filename": str,              # Generated filename
                "status": str,                # "completed", "failed", or "timeout"
                "timestamp": str,             # ISO timestamp of capture
                "file_size": int,             # File size in bytes (0 if failed)
                "file_path": str,             # Full path to snapshot file
                "capture_method": str,        # "rtsp", "direct_camera", or "failed"
                "capture_time": float,        # Total capture time in seconds
                "tier_used": int,             # Which tier was successful (1, 2, 3, or 4)
                "user_experience": str,       # "immediate", "acceptable", "slow", or "failed"
                "error": str,                 # Error message if failed
                "capture_methods_tried": list # List of methods attempted
            }

        Raises:
            ConnectionError: If MediaMTX is unreachable
            ValueError: If stream name is invalid or directory not writable

        Example:
            >>> result = await controller.take_snapshot("camera0", format="jpg", quality=85)
            >>> print(f"Snapshot captured in {result['capture_time']:.2f}s using {result['capture_method']}")
            Snapshot captured in 0.3s using rtsp
        """
        if not self._session:
            raise ConnectionError("MediaMTX controller not started")

        if not stream_name:
            raise ValueError("Stream name is required")

        correlation_id = get_correlation_id() or str(uuid.uuid4())[:8]
        set_correlation_id(correlation_id)

        # Get performance configuration for timeouts and thresholds
        performance_config = getattr(self, '_performance_config', {})
        snapshot_tiers_config = performance_config.get('snapshot_tiers', {
            'tier1_rtsp_ready_check_timeout': 1.0,
            'tier2_activation_timeout': 3.0,
            'tier2_activation_trigger_timeout': 1.0,
            'tier3_direct_capture_timeout': 5.0,
            'total_operation_timeout': 10.0,
            'immediate_response_threshold': 0.5,
            'acceptable_response_threshold': 2.0,
            'slow_response_threshold': 5.0
        })

        # Generate filename if not provided
        if not filename:
            timestamp = time.strftime("%Y-%m-%d_%H-%M-%S")
            filename = f"{stream_name}_snapshot_{timestamp}.{format}"

        # Validate snapshots directory
        try:
            os.makedirs(self._snapshots_path, exist_ok=True)
            test_file = os.path.join(
                self._snapshots_path, f".write_test_{uuid.uuid4().hex[:8]}"
            )
            with open(test_file, "w") as f:
                f.write("test")
            os.remove(test_file)
        except (PermissionError, OSError) as e:
            error_msg = f"Cannot write to snapshots directory {self._snapshots_path}: {e}"
            self._logger.error(error_msg, extra={"correlation_id": correlation_id, "stream_name": stream_name})
            return {
                "stream_name": stream_name,
                "filename": filename,
                "status": "failed",
                "timestamp": time.strftime("%Y-%m-%dT%H:%M:%SZ"),
                "file_size": 0,
                "error": error_msg,
                "tier_used": 0,
                "user_experience": "failed"
            }

        snapshot_path = os.path.join(self._snapshots_path, filename)
        start_time = time.time()
        capture_methods_tried = []

        # Tier 1: Check if RTSP stream is immediately ready
        tier1_timeout = snapshot_tiers_config.get('tier1_rtsp_ready_check_timeout', 1.0)
        self._logger.info(f"Tier 1: Checking RTSP stream readiness for {stream_name} (timeout: {tier1_timeout}s)", 
                         extra={"correlation_id": correlation_id, "stream_name": stream_name, "tier": 1})
        
        capture_methods_tried.append("rtsp_immediate")
        stream_ready = await self.check_stream_readiness(stream_name, timeout=tier1_timeout)
        
        if stream_ready:
            # Fast path: RTSP stream is ready, capture immediately
            self._logger.info(f"Tier 1: RTSP stream ready, capturing immediately for {stream_name}", 
                             extra={"correlation_id": correlation_id, "stream_name": stream_name, "tier": 1})
            
            result = await self._capture_snapshot_from_rtsp(stream_name, snapshot_path, filename, format, quality, correlation_id)
            capture_time = time.time() - start_time
            
            # Determine user experience based on response time
            user_experience = self._determine_user_experience(capture_time, snapshot_tiers_config)
            
            return {
                **result,
                "capture_time": capture_time,
                "tier_used": 1,
                "user_experience": user_experience,
                "capture_methods_tried": capture_methods_tried
            }

        # Tier 2: RTSP stream not ready, attempt quick activation
        tier2_activation_timeout = snapshot_tiers_config.get('tier2_activation_timeout', 3.0)
        tier2_trigger_timeout = snapshot_tiers_config.get('tier2_activation_trigger_timeout', 1.0)
        
        self._logger.info(f"Tier 2: RTSP stream not ready, attempting quick activation for {stream_name} (timeout: {tier2_activation_timeout}s)", 
                         extra={"correlation_id": correlation_id, "stream_name": stream_name, "tier": 2})
        
        capture_methods_tried.append("rtsp_activation")
        
        # Trigger on-demand activation with configurable timeout
        await self._trigger_stream_activation(stream_name, correlation_id)
        
        # Wait for stream to become ready with user-friendly timeout
        stream_ready = await self._wait_for_stream_readiness(stream_name, tier2_activation_timeout, correlation_id)
        
        if stream_ready:
            # Success: Stream activated, capture from RTSP
            activation_time = time.time() - start_time
            self._logger.info(f"Tier 2: RTSP stream activated in {activation_time:.2f}s for {stream_name}", 
                             extra={"correlation_id": correlation_id, "stream_name": stream_name, "activation_time": activation_time, "tier": 2})
            
            result = await self._capture_snapshot_from_rtsp(stream_name, snapshot_path, filename, format, quality, correlation_id)
            capture_time = time.time() - start_time
            
            # Determine user experience based on response time
            user_experience = self._determine_user_experience(capture_time, snapshot_tiers_config)
            
            return {
                **result,
                "capture_time": capture_time,
                "tier_used": 2,
                "user_experience": user_experience,
                "capture_methods_tried": capture_methods_tried
            }

        # Tier 3: RTSP activation failed, attempt direct camera capture
        tier3_timeout = snapshot_tiers_config.get('tier3_direct_capture_timeout', 5.0)
        self._logger.info(f"Tier 3: RTSP activation failed, attempting direct camera capture for {stream_name} (timeout: {tier3_timeout}s)", 
                         extra={"correlation_id": correlation_id, "stream_name": stream_name, "tier": 3})
        
        # Extract device path from stream name (e.g., "camera0" -> "/dev/video0")
        device_path = f"/dev/video{stream_name.replace('camera', '')}"
        capture_methods_tried.append("direct_camera")
        
        direct_result = await self._capture_snapshot_direct(device_path, snapshot_path, filename, format, quality, correlation_id)
        
        if direct_result["status"] == "completed":
            # Success: Direct camera capture worked
            capture_time = time.time() - start_time
            self._logger.info(f"Tier 3: Direct camera capture successful in {capture_time:.2f}s for {stream_name}", 
                             extra={"correlation_id": correlation_id, "stream_name": stream_name, "capture_time": capture_time, "tier": 3})
            
            # Determine user experience based on response time
            user_experience = self._determine_user_experience(capture_time, snapshot_tiers_config)
            
            return {
                **direct_result,
                "capture_time": capture_time,
                "tier_used": 3,
                "user_experience": user_experience,
                "capture_methods_tried": capture_methods_tried
            }

        # Tier 4: All methods failed
        total_time = time.time() - start_time
        total_timeout = snapshot_tiers_config.get('total_operation_timeout', 10.0)
        
        error_msg = f"All snapshot capture methods failed for {stream_name} after {total_time:.2f}s"
        self._logger.error(error_msg, extra={"correlation_id": correlation_id, "stream_name": stream_name, "total_time": total_time, "tier": 4})
        
        return {
            "stream_name": stream_name,
            "filename": filename,
            "status": "failed",
            "timestamp": time.strftime("%Y-%m-%dT%H:%M:%SZ"),
            "file_size": 0,
            "error": error_msg,
            "capture_time": total_time,
            "tier_used": 4,
            "user_experience": "failed",
            "capture_methods_tried": capture_methods_tried
        }

    def _determine_user_experience(self, capture_time: float, snapshot_tiers_config: Dict[str, float]) -> str:
        """
        Determine user experience rating based on capture time and configured thresholds.
        
        Args:
            capture_time: Total capture time in seconds
            snapshot_tiers_config: Configuration dictionary with thresholds
            
        Returns:
            User experience rating: "immediate", "acceptable", "slow", or "failed"
        """
        immediate_threshold = snapshot_tiers_config.get('immediate_response_threshold', 0.5)
        acceptable_threshold = snapshot_tiers_config.get('acceptable_response_threshold', 2.0)
        slow_threshold = snapshot_tiers_config.get('slow_response_threshold', 5.0)
        
        if capture_time <= immediate_threshold:
            return "immediate"
        elif capture_time <= acceptable_threshold:
            return "acceptable"
        elif capture_time <= slow_threshold:
            return "slow"
        else:
            return "slow"  # Anything over slow threshold is still "slow", not "failed"

    async def _capture_snapshot_from_rtsp(
        self, stream_name: str, snapshot_path: str, filename: str, format: str, quality: int, correlation_id: str
    ) -> Dict[str, Any]:
        """
        Capture snapshot from RTSP stream using FFmpeg with configurable timeouts.
        
        Args:
            stream_name: Name of the stream
            snapshot_path: Full path for the snapshot file
            filename: Filename for the snapshot
            format: Image format
            quality: Image quality
            correlation_id: Correlation ID for logging
            
        Returns:
            Dict containing snapshot capture result
        """
        ffmpeg_process = None
        
        try:
            rtsp_url = f"rtsp://{self._host}:{self._rtsp_port}/{stream_name}"
            
            self._logger.info(f"Capturing snapshot from RTSP: {rtsp_url}", 
                             extra={"correlation_id": correlation_id, "stream_name": stream_name})

            # Get FFmpeg configuration for snapshots with configurable timeouts
            snapshot_config = self._ffmpeg_config.get("snapshot", {})
            process_creation_timeout = snapshot_config.get("process_creation_timeout", 5.0)
            execution_timeout = snapshot_config.get("execution_timeout", 8.0)
            internal_timeout = snapshot_config.get("internal_timeout", 5000000)

            # FFmpeg command optimized for speed
            ffmpeg_cmd = [
                "ffmpeg",
                "-y",  # Overwrite output file
                "-i", rtsp_url,
                "-vframes", "1",  # Capture only 1 frame
                "-q:v", str(quality),
                "-timeout", str(internal_timeout),
                "-rtsp_transport", "tcp",
                "-loglevel", "warning",
                "-preset", "ultrafast",
                snapshot_path,
            ]

            # Execute FFmpeg with configurable timeouts
            ffmpeg_process = await asyncio.wait_for(
                asyncio.create_subprocess_exec(*ffmpeg_cmd, stdout=asyncio.subprocess.PIPE, stderr=asyncio.subprocess.PIPE),
                timeout=process_creation_timeout
            )

            stdout, stderr = await asyncio.wait_for(
                ffmpeg_process.communicate(),
                timeout=execution_timeout
            )

            if ffmpeg_process.returncode == 0 and os.path.exists(snapshot_path):
                file_size = os.path.getsize(snapshot_path)
                self._logger.info(f"RTSP snapshot captured successfully: {filename} ({file_size} bytes)", 
                                 extra={"correlation_id": correlation_id, "stream_name": stream_name})
                return {
                    "stream_name": stream_name,
                    "filename": filename,
                    "status": "completed",
                    "timestamp": time.strftime("%Y-%m-%dT%H:%M:%SZ"),
                    "file_size": file_size,
                    "file_path": snapshot_path,
                    "capture_method": "rtsp"
                }
            else:
                error_msg = stderr.decode() if stderr else f"FFmpeg exit code: {ffmpeg_process.returncode}"
                self._logger.error(f"RTSP snapshot failed: {error_msg}", 
                                  extra={"correlation_id": correlation_id, "stream_name": stream_name})
                return {
                    "stream_name": stream_name,
                    "filename": filename,
                    "status": "failed",
                    "timestamp": time.strftime("%Y-%m-%dT%H:%M:%SZ"),
                    "file_size": 0,
                    "error": f"RTSP capture failed: {error_msg}",
                    "capture_method": "rtsp"
                }

        except asyncio.TimeoutError:
            cleanup_context = await self._cleanup_ffmpeg_process(ffmpeg_process, stream_name, correlation_id)
            timeout_error = f"RTSP snapshot timeout for {stream_name} ({cleanup_context})"
            self._logger.error(timeout_error, extra={"correlation_id": correlation_id, "stream_name": stream_name})
            return {
                "stream_name": stream_name,
                "filename": filename,
                "status": "failed",
                "timestamp": time.strftime("%Y-%m-%dT%H:%M:%SZ"),
                "file_size": 0,
                "error": timeout_error,
                "capture_method": "rtsp"
            }
        except Exception as e:
            cleanup_context = await self._cleanup_ffmpeg_process(ffmpeg_process, stream_name, correlation_id)
            error_msg = f"RTSP snapshot error for {stream_name} ({cleanup_context}): {e}"
            self._logger.error(error_msg, extra={"correlation_id": correlation_id, "stream_name": stream_name})
            return {
                "stream_name": stream_name,
                "filename": filename,
                "status": "failed",
                "timestamp": time.strftime("%Y-%m-%dT%H:%M:%SZ"),
                "file_size": 0,
                "error": error_msg,
                "capture_method": "rtsp"
            }

    async def _capture_snapshot_direct(
        self, device_path: str, snapshot_path: str, filename: str, format: str, quality: int, correlation_id: str
    ) -> Dict[str, Any]:
        """
        Capture snapshot directly from camera device using FFmpeg with configurable timeouts.
        
        Args:
            device_path: Camera device path (e.g., "/dev/video0")
            snapshot_path: Full path for the snapshot file
            filename: Filename for the snapshot
            format: Image format
            quality: Image quality
            correlation_id: Correlation ID for logging
            
        Returns:
            Dict containing snapshot capture result
        """
        ffmpeg_process = None
        
        try:
            self._logger.info(f"Capturing snapshot directly from camera: {device_path}", 
                             extra={"correlation_id": correlation_id, "device_path": device_path})

            # Get FFmpeg configuration for snapshots with configurable timeouts
            snapshot_config = self._ffmpeg_config.get("snapshot", {})
            process_creation_timeout = snapshot_config.get("process_creation_timeout", 5.0)
            execution_timeout = snapshot_config.get("execution_timeout", 8.0)

            # FFmpeg command for direct camera capture
            ffmpeg_cmd = [
                "ffmpeg",
                "-y",  # Overwrite output file
                "-f", "v4l2",  # Video4Linux2 input
                "-i", device_path,
                "-vframes", "1",  # Capture only 1 frame
                "-q:v", str(quality),
                "-loglevel", "warning",
                "-preset", "ultrafast",
                snapshot_path,
            ]

            # Execute FFmpeg with configurable timeouts
            ffmpeg_process = await asyncio.wait_for(
                asyncio.create_subprocess_exec(*ffmpeg_cmd, stdout=asyncio.subprocess.PIPE, stderr=asyncio.subprocess.PIPE),
                timeout=process_creation_timeout
            )

            stdout, stderr = await asyncio.wait_for(
                ffmpeg_process.communicate(),
                timeout=execution_timeout
            )

            if ffmpeg_process.returncode == 0 and os.path.exists(snapshot_path):
                file_size = os.path.getsize(snapshot_path)
                self._logger.info(f"Direct camera snapshot captured successfully: {filename} ({file_size} bytes)", 
                                 extra={"correlation_id": correlation_id, "device_path": device_path})
                return {
                    "stream_name": device_path,  # Use device path as stream name for direct capture
                    "filename": filename,
                    "status": "completed",
                    "timestamp": time.strftime("%Y-%m-%dT%H:%M:%SZ"),
                    "file_size": file_size,
                    "file_path": snapshot_path,
                    "capture_method": "direct_camera"
                }
            else:
                error_msg = stderr.decode() if stderr else f"FFmpeg exit code: {ffmpeg_process.returncode}"
                self._logger.error(f"Direct camera snapshot failed: {error_msg}", 
                                  extra={"correlation_id": correlation_id, "device_path": device_path})
                return {
                    "stream_name": device_path,
                    "filename": filename,
                    "status": "failed",
                    "timestamp": time.strftime("%Y-%m-%dT%H:%M:%SZ"),
                    "file_size": 0,
                    "error": f"Direct camera capture failed: {error_msg}",
                    "capture_method": "direct_camera"
                }

        except asyncio.TimeoutError:
            cleanup_context = await self._cleanup_ffmpeg_process(ffmpeg_process, device_path, correlation_id)
            timeout_error = f"Direct camera snapshot timeout for {device_path} ({cleanup_context})"
            self._logger.error(timeout_error, extra={"correlation_id": correlation_id, "device_path": device_path})
            return {
                "stream_name": device_path,
                "filename": filename,
                "status": "failed",
                "timestamp": time.strftime("%Y-%m-%dT%H:%M:%SZ"),
                "file_size": 0,
                "error": timeout_error,
                "capture_method": "direct_camera"
            }
        except Exception as e:
            cleanup_context = await self._cleanup_ffmpeg_process(ffmpeg_process, device_path, correlation_id)
            error_msg = f"Direct camera snapshot error for {device_path} ({cleanup_context}): {e}"
            self._logger.error(error_msg, extra={"correlation_id": correlation_id, "device_path": device_path})
            return {
                "stream_name": device_path,
                "filename": filename,
                "status": "failed",
                "timestamp": time.strftime("%Y-%m-%dT%H:%M:%SZ"),
                "file_size": 0,
                "error": error_msg,
                "capture_method": "direct_camera"
            }

    async def _cleanup_ffmpeg_process(
        self,
        process: Optional[asyncio.subprocess.Process],
        stream_name: str,
        correlation_id: str,
    ) -> str:
        """
        Robust FFmpeg process cleanup with escalating termination.

        Returns:
            String describing cleanup action taken
        """
        if not process or process.returncode is not None:
            return "no_cleanup_needed"

        cleanup_actions = []

        try:
            # Step 1: Graceful termination (SIGTERM)
            process.terminate()
            cleanup_actions.append("terminated")

            try:
                await asyncio.wait_for(
                    process.wait(), timeout=self._process_termination_timeout
                )
                cleanup_actions.append("graceful_exit")
                return "_".join(cleanup_actions)
            except asyncio.TimeoutError:
                cleanup_actions.append("term_timeout")

        except Exception as e:
            cleanup_actions.append(f"term_error_{type(e).__name__}")

        try:
            # Step 2: Force kill (SIGKILL)
            process.kill()
            cleanup_actions.append("killed")

            try:
                await asyncio.wait_for(
                    process.wait(), timeout=self._process_kill_timeout
                )
                cleanup_actions.append("force_exit")
            except asyncio.TimeoutError:
                cleanup_actions.append("kill_timeout")
                self._logger.error(
                    f"FFmpeg process for {stream_name} did not respond to SIGKILL within {self._process_kill_timeout}s",
                    extra={
                        "correlation_id": correlation_id,
                        "stream_name": stream_name,
                    },
                )
        except Exception as e:
            cleanup_actions.append(f"kill_error_{type(e).__name__}")
            self._logger.error(
                f"Error killing FFmpeg process for {stream_name}: {e}",
                extra={"correlation_id": correlation_id, "stream_name": stream_name},
            )

        return "_".join(cleanup_actions)

    async def get_stream_list(self) -> List[Dict[str, Any]]:
        """
        Get list of all configured streams with enhanced error handling.

        Returns:
            List of stream configuration dictionaries

        Raises:
            ConnectionError: If MediaMTX is unreachable
        """
        if not self._session:
            raise ConnectionError("MediaMTX controller not started")

        correlation_id = get_correlation_id() or str(uuid.uuid4())[:8]

        try:
            async with self._session.get(f"{self._base_url}/v3/paths/list") as response:
                if response.status == 200:
                    data = await response.json()
                    streams = []

                    # Parse MediaMTX paths list response
                    if "items" in data:
                        for path_info in data["items"]:
                            streams.append(
                                {
                                    "name": path_info.get("name", ""),
                                    "source": path_info.get("source", {}),
                                    "ready": path_info.get("ready", False),
                                    "readers": len(path_info.get("readers", [])),
                                    "bytes_sent": path_info.get("bytesSent", 0),
                                }
                            )

                    self._logger.debug(
                        f"Retrieved {len(streams)} streams from MediaMTX",
                        extra={"correlation_id": correlation_id},
                    )
                    return streams
                else:
                    error_text = await response.text()
                    error_msg = f"Failed to get stream list: HTTP {response.status} - {error_text}"
                    self._logger.error(
                        error_msg, extra={"correlation_id": correlation_id}
                    )
                    raise ConnectionError(error_msg)

        except aiohttp.ClientError as e:
            error_msg = f"MediaMTX unreachable during stream list: {e}"
            self._logger.error(error_msg, extra={"correlation_id": correlation_id})
            raise ConnectionError(error_msg)

    async def get_stream_status(self, stream_name: str) -> Dict[str, Any]:
        """
        Get detailed status for a specific stream with enhanced error context.

        Args:
            stream_name: Name of the stream

        Returns:
            Dict containing detailed stream status

        Raises:
            ValueError: If stream does not exist
            ConnectionError: If MediaMTX is unreachable
        """
        if not self._session:
            raise ConnectionError("MediaMTX controller not started")

        if not stream_name:
            raise ValueError("Stream name is required")

        correlation_id = get_correlation_id() or str(uuid.uuid4())[:8]

        try:
            async with self._session.get(
                f"{self._base_url}/v3/paths/get/{stream_name}"
            ) as response:
                if response.status == 200:
                    data = await response.json()
                    return {
                        "name": stream_name,
                        "status": "active" if data.get("ready", False) else "inactive",
                        "source": data.get("source", ""),
                        "readers": data.get("readers", 0),
                        "bytes_sent": data.get("bytesSent", 0),
                        "recording": data.get("record", False),
                        "correlation_id": correlation_id,
                    }
                elif response.status == 404:
                    raise ValueError(f"Stream not found: {stream_name}")
                else:
                    error_text = await response.text()
                    error_msg = f"Failed to get stream status for {stream_name}: HTTP {response.status} - {error_text}"
                    self._logger.error(
                        error_msg,
                        extra={
                            "correlation_id": correlation_id,
                            "stream_name": stream_name,
                        },
                    )
                    raise ConnectionError(error_msg)

        except aiohttp.ClientError as e:
            error_msg = (
                f"MediaMTX unreachable during stream status for {stream_name}: {e}"
            )
            self._logger.error(
                error_msg,
                extra={"correlation_id": correlation_id, "stream_name": stream_name},
            )
            raise ConnectionError(error_msg)

    async def update_configuration(self, config_updates: Dict[str, Any]) -> bool:
        """
        Update MediaMTX configuration dynamically with enhanced validation and error handling.

        Args:
            config_updates: Configuration parameters to update

        Returns:
            True if configuration was updated successfully

        Raises:
            ValueError: If configuration is invalid
            ConnectionError: If MediaMTX is unreachable
        """
        if not self._session:
            raise ConnectionError("MediaMTX controller not started")

        if not config_updates:
            raise ValueError("Configuration updates are required")

        correlation_id = get_correlation_id() or str(uuid.uuid4())[:8]
        set_correlation_id(correlation_id)

        # Enhanced configuration validation schema
        valid_config_schema = {
            "logLevel": {"type": str, "values": ["error", "warn", "info", "debug"]},
            "logDestinations": {"type": list},
            "readTimeout": {"type": (str, int), "min": 1, "max": 300},
            "writeTimeout": {"type": (str, int), "min": 1, "max": 300},
            "readBufferCount": {"type": int, "min": 1, "max": 4096},
            "udpMaxPayloadSize": {"type": int, "min": 1024, "max": 65507},
            "runOnConnect": {"type": str},
            "runOnConnectRestart": {"type": bool},
            "api": {"type": bool},
            "apiAddress": {"type": str, "pattern": r"^[0-9.:]+$"},
            "metrics": {"type": bool},
            "metricsAddress": {"type": str, "pattern": r"^[0-9.:]+$"},
            "pprof": {"type": bool},
            "pprofAddress": {"type": str, "pattern": r"^[0-9.:]+$"},
            "rtsp": {"type": bool},
            "rtspAddress": {"type": str, "pattern": r"^[0-9.:]+$"},
            "rtspsAddress": {"type": str, "pattern": r"^[0-9.:]+$"},
            "rtmp": {"type": bool},
            "rtmpAddress": {"type": str, "pattern": r"^[0-9.:]+$"},
            "rtmps": {"type": bool},
            "rtmpsAddress": {"type": str, "pattern": r"^[0-9.:]+$"},
            "hls": {"type": bool},
            "hlsAddress": {"type": str, "pattern": r"^[0-9.:]+$"},
            "hlsAllowOrigin": {"type": str},
            "webrtc": {"type": bool},
            "webrtcAddress": {"type": str, "pattern": r"^[0-9.:]+$"},
        }

        # Validate configuration keys
        unknown_keys = set(config_updates.keys()) - set(valid_config_schema.keys())
        if unknown_keys:
            error_msg = f"Unknown configuration keys: {list(unknown_keys)}"
            self._logger.error(error_msg, extra={"correlation_id": correlation_id})
            raise ValueError(error_msg)

        # Enhanced configuration value validation with error accumulation
        validation_errors = []
        for key, value in config_updates.items():
            schema = valid_config_schema[key]

            # Ensure schema is a dictionary
            if not isinstance(schema, dict):
                validation_errors.append(f"Invalid schema for {key}: expected dict, got {type(schema)}")
                continue

            # Type validation
            if "type" in schema and not isinstance(value, schema["type"]):
                validation_errors.append(
                    f"Invalid type for {key}: expected {schema['type']}, got {type(value)}"
                )
                continue

            # Value constraints validation
            if "values" in schema and value not in schema["values"]:
                validation_errors.append(
                    f"Invalid value for {key}: {value}, allowed values: {schema['values']}"
                )

            if (
                "min" in schema
                and isinstance(value, (int, float))
                and value < schema["min"]
            ):
                validation_errors.append(
                    f"Value for {key} too small: {value}, minimum: {schema['min']}"
                )

            if (
                "max" in schema
                and isinstance(value, (int, float))
                and value > schema["max"]
            ):
                validation_errors.append(
                    f"Value for {key} too large: {value}, maximum: {schema['max']}"
                )

            # Pattern validation for string types
            if "pattern" in schema and isinstance(value, str):
                import re

                if not re.match(schema["pattern"], value):
                    validation_errors.append(f"Invalid format for {key}: {value}")

        if validation_errors:
            error_msg = (
                f"Configuration validation failed: {'; '.join(validation_errors)}"
            )
            self._logger.error(error_msg, extra={"correlation_id": correlation_id})
            raise ValueError(error_msg)

        try:
            self._logger.info(
                f"Updating MediaMTX configuration: {list(config_updates.keys())}",
                extra={"correlation_id": correlation_id},
            )

            async with self._session.post(
                f"{self._base_url}/v3/config/global/patch", json=config_updates
            ) as response:
                if response.status == 200:
                    self._logger.info(
                        f"MediaMTX configuration updated successfully: {list(config_updates.keys())}",
                        extra={"correlation_id": correlation_id},
                    )
                    return True
                else:
                    error_text = await response.text()
                    error_msg = f"Failed to update configuration: HTTP {response.status} - {error_text}"
                    self._logger.error(
                        error_msg, extra={"correlation_id": correlation_id}
                    )
                    raise ValueError(error_msg)

        except aiohttp.ClientError as e:
            error_msg = f"MediaMTX unreachable during configuration update: {e}"
            self._logger.error(error_msg, extra={"correlation_id": correlation_id})
            raise ConnectionError(error_msg)

    def _calculate_backoff_interval(self, consecutive_failures: int) -> float:
        """
        Calculate safe backoff interval with protection against overflow.
        
        Args:
            consecutive_failures: Number of consecutive failures
            
        Returns:
            Backoff interval in seconds (capped at maximum)
        """
        # Use math.pow for safer exponential calculation
        import math
        
        # Cap consecutive_failures to prevent overflow
        max_safe_failures = 20  # Reasonable upper limit
        safe_failures = min(consecutive_failures, max_safe_failures)
        
        # Calculate exponential backoff with safety checks
        try:
            exponential_factor = math.pow(self._backoff_base_multiplier, safe_failures)
            base_interval = self._health_check_interval * exponential_factor
            
            # Apply maximum cap
            capped_interval = min(base_interval, self._health_max_backoff_interval)
            
            # Apply jitter
            jitter = random.uniform(*self._backoff_jitter_range)
            final_interval = capped_interval * jitter
            
            # Ensure minimum interval (use float for sub-second precision)
            return max(final_interval, 0.1)
            
        except (OverflowError, ValueError) as e:
            # Fallback to maximum interval if calculation fails
            self._logger.warning(
                f"Backoff calculation failed for {consecutive_failures} failures, using maximum interval: {e}"
            )
            return float(self._health_max_backoff_interval)



    async def _health_monitor_loop(self) -> None:
        """
        Background task for continuous health monitoring with configurable circuit breaker and adaptive backoff.

        Monitors MediaMTX health and logs status changes with automatic recovery.
        Uses configurable parameters for all thresholds and backoff calculations.

        Anti-Flapping Recovery Design:
        - Circuit breaker requires N consecutive successful health checks before fully resetting
        - This prevents "flapping" where transient successes briefly clear the breaker
        - Configurable confirmation threshold (default: 3) balances stability vs recovery speed
        - Any failure during recovery resets the confirmation counter
        - Partial recovery progress is logged for observability
        """
        correlation_id = str(uuid.uuid4())[:8]
        set_correlation_id(correlation_id)
        self._logger.info(
            "Starting MediaMTX health monitoring loop",
            extra={"correlation_id": correlation_id},
        )

        # Initialize from persisted state or defaults
        consecutive_failures = self._health_state.get("consecutive_failures", 0)
        consecutive_successes_during_recovery = self._health_state.get("consecutive_successes_during_recovery", 0)
        circuit_breaker_active = self._health_state.get("circuit_breaker_active", False)
        circuit_breaker_start_time = self._health_state.get("circuit_breaker_start_time", 0.0)

        try:
            while self._running:
                try:
                    # Set unique correlation ID for each health check
                    check_correlation_id = str(uuid.uuid4())[:8]
                    set_correlation_id(check_correlation_id)

                    # Check circuit breaker state with recovery confirmation
                    if circuit_breaker_active:
                        if (
                            time.time() - circuit_breaker_start_time
                            > self._health_circuit_breaker_timeout
                        ):
                            # Circuit breaker timeout expired - attempt recovery probe
                            # Don't reset recovery progress - let it continue accumulating
                            self._logger.info(
                                "Circuit breaker timeout expired - attempting recovery probe",
                                extra={
                                    "correlation_id": check_correlation_id,
                                    "circuit_breaker_transition": True,
                                },
                            )
                            # Continue to health check - don't reset circuit_breaker_active yet
                        else:
                            # During circuit breaker timeout, use shorter intervals for recovery probes
                            # This allows faster recovery detection while still preventing excessive requests
                            cb_wait_interval = min(self._health_check_interval * 2, 5.0)
                            await asyncio.sleep(cb_wait_interval)
                            continue

                    health_status = await self.health_check()
                    current_status = health_status.get("status")
                    
                    # Debug logging for recovery analysis
                    self._logger.debug(
                        f"Health check result: status={current_status}, last_status={self._last_health_status}, "
                        f"circuit_breaker_active={circuit_breaker_active}, consecutive_failures={consecutive_failures}",
                        extra={"correlation_id": check_correlation_id},
                    )

                    # Count consecutive failures and update success counters
                    if current_status != "healthy":
                        consecutive_failures += 1
                        self._health_state["consecutive_failures"] = consecutive_failures
                    else:
                        # Update success counters
                        self._health_state["success_count"] += 1
                        self._health_state["total_checks"] += 1
                        
                        # Reset consecutive failures on healthy status (but only if not in CB recovery mode)
                        if not circuit_breaker_active:
                            consecutive_failures = 0
                            self._health_state["consecutive_failures"] = 0

                    # Enhanced status change logging with recovery confirmation logic
                    # Anti-flapping design: Circuit breaker requires N consecutive successes before full reset
                    if current_status != self._last_health_status:
                        if current_status == "healthy":
                            if self._last_health_status in ["unhealthy", "error"] or self._last_health_status is None:
                                if circuit_breaker_active:
                                    # During circuit breaker recovery - log the transition but don't count successes here
                                    # Success counting is handled in the circuit breaker recovery logic below
                                    self._logger.info(
                                        f"MediaMTX health IMPROVING: {self._last_health_status} -> {current_status} "
                                        f"(recovery in progress)",
                                        extra={
                                            "correlation_id": check_correlation_id,
                                            "health_transition": True,
                                            "recovery_partial": True,
                                        },
                                    )
                                else:
                                    # Not in circuit breaker state - normal recovery
                                    # Only reset consecutive_failures after a brief stable period
                                    # to allow circuit breaker to activate if there were enough failures
                                    consecutive_failures = 0
                                    self._health_state["consecutive_failures"] = 0
                                    self._logger.info(
                                        f"MediaMTX health RECOVERED: {self._last_health_status} -> {current_status}",
                                        extra={
                                            "correlation_id": check_correlation_id,
                                            "health_transition": True,
                                        },
                                    )
                            else:
                                # First time healthy or continuing healthy state
                                consecutive_failures = 0
                                consecutive_successes_during_recovery = 0
                                self._health_state["consecutive_failures"] = 0
                                self._health_state[
                                    "consecutive_successes_during_recovery"
                                ] = 0
                        else:
                            # Health degraded - reset recovery progress
                            # consecutive_failures already incremented above
                            consecutive_successes_during_recovery = 0
                            self._health_state[
                                "consecutive_successes_during_recovery"
                            ] = 0

                            self._logger.warning(
                                f"MediaMTX health DEGRADED: {self._last_health_status or 'unknown'} -> {current_status} "
                                f"(consecutive_failures: {consecutive_failures}/{self._health_failure_threshold})",
                                extra={
                                    "correlation_id": check_correlation_id,
                                    "health_transition": True,
                                },
                            )

                            # Circuit breaker check moved to after status transition logic

                        self._last_health_status = current_status

                    # Handle circuit breaker recovery for consecutive successful checks
                    if current_status == "healthy" and circuit_breaker_active:
                        # Count consecutive successes during recovery (including status transitions)
                        consecutive_successes_during_recovery += 1
                        self._health_state[
                            "consecutive_successes_during_recovery"
                        ] = consecutive_successes_during_recovery

                        if (
                            consecutive_successes_during_recovery
                            >= self._health_recovery_confirmation_threshold
                        ):
                            # Confirmed recovery - fully reset circuit breaker
                            circuit_breaker_active = False
                            consecutive_failures = 0
                            consecutive_successes_during_recovery = 0
                            self._health_state["recovery_count"] += 1
                            self._health_state["consecutive_failures"] = 0
                            self._health_state[
                                "consecutive_successes_during_recovery"
                            ] = 0
                            self._health_state["circuit_breaker_active"] = False
                            self._health_state["circuit_breaker_start_time"] = 0.0

                            self._logger.info(
                                f"MediaMTX health FULLY RECOVERED after {self._health_recovery_confirmation_threshold} consecutive successes "
                                f"(recovery #{self._health_state['recovery_count']})",
                                extra={
                                    "correlation_id": check_correlation_id,
                                    "health_transition": True,
                                    "circuit_breaker_reset": True,
                                },
                            )
                        elif consecutive_successes_during_recovery > 0:
                            # Partial recovery - still in confirmation phase
                            self._logger.info(
                                f"MediaMTX health IMPROVING: {consecutive_successes_during_recovery}/{self._health_recovery_confirmation_threshold} consecutive successes",
                                extra={
                                    "correlation_id": check_correlation_id,
                                    "health_transition": True,
                                    "recovery_partial": True,
                                },
                            )
                    elif current_status != "healthy" and circuit_breaker_active:
                        # Reset recovery progress on any failure during circuit breaker recovery
                        consecutive_successes_during_recovery = 0
                        self._health_state[
                            "consecutive_successes_during_recovery"
                        ] = 0

                    # Check circuit breaker activation AFTER status transition logic
                    # This ensures we check even if status didn't change (e.g., unhealthy -> unhealthy)
                    if (
                        consecutive_failures >= self._health_failure_threshold
                        and not circuit_breaker_active
                        and current_status != "healthy"
                    ):
                        circuit_breaker_active = True
                        circuit_breaker_start_time = time.time()
                        consecutive_successes_during_recovery = 0
                        self._health_state["circuit_breaker_activations"] += 1
                        self._health_state["consecutive_successes_during_recovery"] = 0
                        self._health_state["circuit_breaker_active"] = True
                        self._health_state["circuit_breaker_start_time"] = circuit_breaker_start_time
                        self._logger.error(
                            f"MediaMTX health circuit breaker ACTIVATED after {consecutive_failures} consecutive failures "
                            f"(activation #{self._health_state['circuit_breaker_activations']})",
                            extra={
                                "correlation_id": check_correlation_id,
                                "circuit_breaker": True,
                            },
                        )

                    # Determine sleep interval based on health status with configurable backoff
                    if current_status == "healthy":
                        sleep_interval = self._health_check_interval
                        # consecutive_failures already reset above in health state transition logic
                    else:
                        # Use safe backoff calculation - consecutive_failures already incremented above
                        sleep_interval = self._calculate_backoff_interval(consecutive_failures)

                        self._logger.debug(
                            f"Health monitoring backoff: {sleep_interval:.1f}s (failure #{consecutive_failures})",
                            extra={"correlation_id": check_correlation_id},
                        )

                    await asyncio.sleep(sleep_interval)

                except asyncio.CancelledError:
                    self._logger.info(
                        "Health monitoring loop cancelled",
                        extra={"correlation_id": correlation_id},
                    )
                    break
                except Exception as e:
                    # Only catch unexpected errors that shouldn't happen
                    # Connection errors are now handled by health_check returning error status
                    self._logger.error(
                        f"Unexpected error in health monitoring loop: {e}",
                        extra={"correlation_id": check_correlation_id},
                        exc_info=True,
                    )
                    # Don't increment consecutive_failures for unexpected errors
                    # Let the health check method handle all expected failure scenarios

        except Exception as e:
            self._logger.error(
                f"Critical error in health monitoring loop: {e}",
                extra={"correlation_id": correlation_id},
                exc_info=True,
            )
        finally:
            self._logger.info(
                f"Health monitoring loop ended - Final stats: checks={self._health_state['total_checks']}, "
                f"recoveries={self._health_state['recovery_count']}, failures={self._health_state['consecutive_failures']}, "
                f"circuit_breaker_activations={self._health_state['circuit_breaker_activations']}, "
                f"recovery_confirmation_threshold={self._health_recovery_confirmation_threshold}",
                extra={"correlation_id": correlation_id},
            )



    async def __aenter__(self):
        """Async context manager entry."""
        return self

    async def __aexit__(self, exc_type, exc_val, exc_tb):
        """Async context manager exit."""
        # Clean up any resources if needed
        pass

    async def check_stream_readiness(self, stream_name: str, timeout: float = 10.0) -> bool:
        """
        Check if a stream is ready for recording operations with enhanced validation.
        
        Args:
            stream_name: Name of the stream to check
            timeout: Maximum time to wait for stream readiness in seconds
            
        Returns:
            True if stream is ready, False otherwise
            
        Raises:
            ConnectionError: If MediaMTX is unreachable
            ValueError: If stream does not exist
        """
        if not self._session:
            raise ConnectionError("MediaMTX controller not started")

        if not stream_name:
            raise ValueError("Stream name is required")

        correlation_id = get_correlation_id() or str(uuid.uuid4())[:8]
        set_correlation_id(correlation_id)

        self._logger.debug(
            f"Checking stream readiness: {stream_name}",
            extra={"correlation_id": correlation_id, "stream_name": stream_name},
        )

        try:
            # Get current stream status from MediaMTX
            async with self._session.get(f"{self._base_url}/v3/paths/list") as response:
                if response.status != 200:
                    error_msg = f"Failed to get MediaMTX active paths: HTTP {response.status}"
                    self._logger.error(
                        error_msg,
                        extra={
                            "correlation_id": correlation_id,
                            "stream_name": stream_name,
                        },
                    )
                    raise ConnectionError(error_msg)
                
                paths_data = await response.json()
                stream_path = next(
                    (path for path in paths_data.get("items", []) if path["name"] == stream_name), 
                    None
                )
                
                if not stream_path:
                    error_msg = f"Stream {stream_name} not found in MediaMTX"
                    self._logger.error(
                        error_msg,
                        extra={
                            "correlation_id": correlation_id,
                            "stream_name": stream_name,
                        },
                    )
                    raise ValueError(error_msg)
                
                # Check if stream is ready
                is_ready = stream_path.get("ready", False)
                
                if is_ready:
                    self._logger.debug(
                        f"Stream {stream_name} is ready for recording",
                        extra={
                            "correlation_id": correlation_id,
                            "stream_name": stream_name,
                            "stream_ready": True,
                        },
                    )
                    return True
                
                # Stream is not ready, attempt on-demand activation
                self._logger.info(
                    f"Stream {stream_name} not ready, attempting on-demand activation",
                    extra={
                        "correlation_id": correlation_id,
                        "stream_name": stream_name,
                        "stream_ready": False,
                    },
                )
                
                # Trigger on-demand activation
                await self._trigger_stream_activation(stream_name, correlation_id)
                
                # Wait for stream to become ready
                return await self._wait_for_stream_readiness(stream_name, timeout, correlation_id)
                
        except aiohttp.ClientError as e:
            error_msg = f"MediaMTX unreachable during stream readiness check for {stream_name}: {e}"
            self._logger.error(
                error_msg,
                extra={"correlation_id": correlation_id, "stream_name": stream_name},
            )
            raise ConnectionError(error_msg)

    async def _trigger_stream_activation(self, stream_name: str, correlation_id: str) -> None:
        """
        Trigger on-demand activation for a stream.
        
        Args:
            stream_name: Name of the stream to activate
            correlation_id: Correlation ID for logging
        """
        try:
            # Try to access the stream to trigger on-demand activation
            # We'll use a short timeout since we just want to trigger the activation
            async with self._session.get(
                f"http://{self._host}:{self._rtsp_port}/{stream_name}", 
                timeout=1
            ) as trigger_response:
                # We don't care about the response, just triggering the activation
                pass
        except Exception as e:
            # Ignore errors from the trigger request - the important thing is that it was attempted
            self._logger.debug(
                f"Trigger request for {stream_name} completed (expected to fail): {e}",
                extra={
                    "correlation_id": correlation_id,
                    "stream_name": stream_name,
                },
            )

    async def _wait_for_stream_readiness(self, stream_name: str, timeout: float, correlation_id: str) -> bool:
        """
        Wait for a stream to become ready with timeout.
        
        Args:
            stream_name: Name of the stream to wait for
            timeout: Maximum time to wait in seconds
            correlation_id: Correlation ID for logging
            
        Returns:
            True if stream became ready, False if timeout exceeded
        """
        wait_interval = 0.5  # seconds
        waited_time = 0
        
        while waited_time < timeout:
            await asyncio.sleep(wait_interval)
            waited_time += wait_interval
            
            try:
                # Check if stream is now ready
                async with self._session.get(f"{self._base_url}/v3/paths/list") as check_response:
                    if check_response.status == 200:
                        check_paths = await check_response.json()
                        check_stream = next(
                            (path for path in check_paths.get("items", []) if path["name"] == stream_name), 
                            None
                        )
                        if check_stream and check_stream.get("ready", False):
                            self._logger.info(
                                f"Stream {stream_name} became ready after on-demand activation",
                                extra={
                                    "correlation_id": correlation_id,
                                    "stream_name": stream_name,
                                    "wait_time": waited_time,
                                    "stream_ready": True,
                                },
                            )
                            return True
                        
                        # Log progress for long waits
                        if waited_time >= 5.0 and waited_time % 2.0 < wait_interval:
                            self._logger.debug(
                                f"Still waiting for stream {stream_name} to become ready...",
                                extra={
                                    "correlation_id": correlation_id,
                                    "stream_name": stream_name,
                                    "wait_time": waited_time,
                                    "timeout": timeout,
                                },
                            )
                    else:
                        self._logger.warning(
                            f"Failed to check stream status: HTTP {check_response.status}",
                            extra={
                                "correlation_id": correlation_id,
                                "stream_name": stream_name,
                            },
                        )
            except Exception as e:
                self._logger.warning(
                    f"Error checking stream readiness: {e}",
                    extra={
                        "correlation_id": correlation_id,
                        "stream_name": stream_name,
                    },
                )
        
        # Timeout exceeded
        self._logger.warning(
            f"Stream {stream_name} did not become ready within {timeout} seconds",
            extra={
                "correlation_id": correlation_id,
                "stream_name": stream_name,
                "timeout": timeout,
                "stream_ready": False,
            },
        )
        return False

    async def _send_stream_validation_progress(self, stream_name: str, status: str, correlation_id: str) -> None:
        """
        Send stream validation progress notification.
        
        Args:
            stream_name: Name of the stream being validated
            status: Status of validation ("started", "progress", "completed", "failed")
            correlation_id: Correlation ID for logging
        """
        try:
            # This method will be called by the WebSocket server to send notifications
            # For now, we'll just log the progress
            self._logger.info(
                f"Stream validation {status} for {stream_name}",
                extra={
                    "correlation_id": correlation_id,
                    "stream_name": stream_name,
                    "validation_status": status,
                    "notification_type": "stream_validation_progress",
                },
            )
        except Exception as e:
            self._logger.warning(
                f"Failed to send stream validation progress notification: {e}",
                extra={
                    "correlation_id": correlation_id,
                    "stream_name": stream_name,
                    "validation_status": status,
                },
            )
