"""
MediaMTX REST API controller for stream and recording management.

Enhanced version with improved error handling, process management,
and operational robustness per architecture requirements.
"""

import asyncio
import logging
import os
import signal
import time
import uuid
from typing import Dict, Any, Optional, List
from dataclasses import dataclass
import aiohttp
import json

from ..camera_service.logging_config import set_correlation_id, get_correlation_id


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
        snapshots_path: str
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
        """
        self._host = host
        self._api_port = api_port
        self._rtsp_port = rtsp_port
        self._webrtc_port = webrtc_port
        self._hls_port = hls_port
        self._config_path = config_path
        self._recordings_path = recordings_path
        self._snapshots_path = snapshots_path
        
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
            'consecutive_failures': 0,
            'last_success_time': 0.0,
            'last_failure_time': 0.0,
            'total_checks': 0,
            'recovery_count': 0
        }

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
        
        self._logger.info("Starting MediaMTX controller", extra={'correlation_id': correlation_id})
        
        # Create aiohttp ClientSession with timeout configuration
        timeout = aiohttp.ClientTimeout(total=15, connect=5)
        connector = aiohttp.TCPConnector(limit=10, limit_per_host=5)
        self._session = aiohttp.ClientSession(
            timeout=timeout,
            connector=connector,
            headers={"Content-Type": "application/json"}
        )
        
        # Start health monitoring task
        self._health_check_task = asyncio.create_task(self._health_monitor_loop())
        self._running = True
        
        self._logger.info("MediaMTX controller started successfully", 
                         extra={'correlation_id': correlation_id})

    async def stop(self) -> None:
        """
        Stop the MediaMTX controller.
        
        Closes HTTP client session and stops monitoring tasks.
        """
        if not self._running:
            return
            
        correlation_id = get_correlation_id() or str(uuid.uuid4())[:8]
        set_correlation_id(correlation_id)
        
        self._logger.info("Stopping MediaMTX controller", extra={'correlation_id': correlation_id})
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
            
        self._logger.info("MediaMTX controller stopped", extra={'correlation_id': correlation_id})

    async def health_check(self) -> Dict[str, Any]:
        """
        Perform health check on MediaMTX server with enhanced error context.
        
        Returns:
            Dict containing health status and metrics
            
        Raises:
            ConnectionError: If MediaMTX is unreachable
        """
        if not self._session:
            raise ConnectionError("MediaMTX controller not started")
            
        correlation_id = get_correlation_id() or str(uuid.uuid4())[:8]
        start_time = time.time()
        
        try:
            # Call MediaMTX API config endpoint to verify connectivity
            async with self._session.get(f"{self._base_url}/v3/config/global/get") as response:
                response_time = int((time.time() - start_time) * 1000)
                
                if response.status == 200:
                    config_data = await response.json()
                    self._health_state['last_success_time'] = time.time()
                    self._health_state['total_checks'] += 1
                    
                    return {
                        "status": "healthy",
                        "version": config_data.get("serverVersion", "unknown"),
                        "uptime": config_data.get("serverUptime", 0),
                        "api_port": self._api_port,
                        "response_time_ms": response_time,
                        "correlation_id": correlation_id,
                        "consecutive_failures": self._health_state['consecutive_failures']
                    }
                else:
                    error_text = await response.text()
                    self._health_state['last_failure_time'] = time.time()
                    self._health_state['total_checks'] += 1
                    
                    return {
                        "status": "unhealthy",
                        "error": f"HTTP {response.status}: {error_text}",
                        "api_port": self._api_port,
                        "response_time_ms": response_time,
                        "correlation_id": correlation_id,
                        "consecutive_failures": self._health_state['consecutive_failures']
                    }
                    
        except aiohttp.ClientError as e:
            self._health_state['last_failure_time'] = time.time()
            self._health_state['total_checks'] += 1
            raise ConnectionError(f"MediaMTX unreachable at {self._base_url}: {e}")
        except Exception as e:
            self._logger.error(f"Health check failed with unexpected error: {e}", 
                             extra={'correlation_id': correlation_id})
            self._health_state['last_failure_time'] = time.time()
            self._health_state['total_checks'] += 1
            
            return {
                "status": "error",
                "error": f"Unexpected error: {e}",
                "api_port": self._api_port,
                "correlation_id": correlation_id,
                "consecutive_failures": self._health_state['consecutive_failures']
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
        
        self._logger.info(f"Creating stream path: {stream_config.name} from {stream_config.source}",
                         extra={'correlation_id': correlation_id})
        
        try:
            # Check if stream already exists (enhanced idempotent behavior)
            try:
                existing_stream = await self.get_stream_status(stream_config.name)
                if existing_stream.get("name") == stream_config.name:
                    self._logger.info(f"Stream path already exists: {stream_config.name}",
                                    extra={'correlation_id': correlation_id})
                    return self._generate_stream_urls(stream_config.name)
            except ValueError:
                # Stream doesn't exist, proceed with creation
                pass
            
            # Create MediaMTX path configuration
            path_config = {
                "source": stream_config.source,
                "sourceProtocol": "automatic",
                "record": stream_config.record
            }
            
            if stream_config.record and stream_config.record_path:
                path_config["recordPath"] = stream_config.record_path
            
            # Add stream path via MediaMTX API
            async with self._session.post(
                f"{self._base_url}/v3/config/paths/add/{stream_config.name}",
                json=path_config
            ) as response:
                if response.status in [200, 201]:
                    self._logger.info(f"Successfully created stream path: {stream_config.name}",
                                    extra={'correlation_id': correlation_id})
                    return self._generate_stream_urls(stream_config.name)
                elif response.status == 409:
                    # Stream already exists, return URLs (additional idempotency)
                    self._logger.info(f"Stream path already exists (409): {stream_config.name}",
                                    extra={'correlation_id': correlation_id})
                    return self._generate_stream_urls(stream_config.name)
                else:
                    error_text = await response.text()
                    error_msg = f"Failed to create stream {stream_config.name}: HTTP {response.status} - {error_text}"
                    self._logger.error(error_msg, extra={'correlation_id': correlation_id})
                    raise ConnectionError(error_msg)
                    
        except aiohttp.ClientError as e:
            error_msg = f"MediaMTX unreachable during stream creation for {stream_config.name}: {e}"
            self._logger.error(error_msg, extra={'correlation_id': correlation_id})
            raise ConnectionError(error_msg)

    def _generate_stream_urls(self, stream_name: str) -> Dict[str, str]:
        """Generate stream URLs for different protocols."""
        return {
            "rtsp": f"rtsp://{self._host}:{self._rtsp_port}/{stream_name}",
            "webrtc": f"http://{self._host}:{self._webrtc_port}/{stream_name}",
            "hls": f"http://{self._host}:{self._hls_port}/{stream_name}"
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
        
        self._logger.info(f"Deleting stream path: {stream_name}",
                         extra={'correlation_id': correlation_id})
        
        try:
            # Delete stream path via MediaMTX API
            async with self._session.post(f"{self._base_url}/v3/config/paths/delete/{stream_name}") as response:
                if response.status in [200, 204]:
                    self._logger.info(f"Successfully deleted stream path: {stream_name}",
                                    extra={'correlation_id': correlation_id})
                    return True
                elif response.status == 404:
                    # Stream already doesn't exist (idempotent)
                    self._logger.info(f"Stream path already deleted or never existed: {stream_name}",
                                    extra={'correlation_id': correlation_id})
                    return True
                else:
                    error_text = await response.text()
                    error_msg = f"Failed to delete stream {stream_name}: HTTP {response.status} - {error_text}"
                    self._logger.error(error_msg, extra={'correlation_id': correlation_id})
                    return False
                    
        except aiohttp.ClientError as e:
            error_msg = f"MediaMTX unreachable during stream deletion for {stream_name}: {e}"
            self._logger.error(error_msg, extra={'correlation_id': correlation_id})
            raise ConnectionError(error_msg)

    async def start_recording(
        self, 
        stream_name: str, 
        duration: Optional[int] = None,
        format: str = "mp4"
    ) -> Dict[str, Any]:
        """
        Start recording for the specified stream with enhanced session management.
        
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
            
        if stream_name in self._recording_sessions:
            raise ValueError(f"Recording already active for stream: {stream_name}")
            
        correlation_id = get_correlation_id() or str(uuid.uuid4())[:8]
        set_correlation_id(correlation_id)
        
        # Validate format
        valid_formats = ["mp4", "mkv", "avi"]
        if format not in valid_formats:
            raise ValueError(f"Invalid format: {format}. Must be one of: {valid_formats}")
        
        # Ensure recordings directory exists
        os.makedirs(self._recordings_path, exist_ok=True)
        
        try:
            # Generate recording filename with timestamp
            timestamp = time.strftime("%Y-%m-%d_%H-%M-%S")
            filename = f"{stream_name}_{timestamp}.{format}"
            record_path = os.path.join(self._recordings_path, filename)
            
            # Record start time for duration calculation
            start_time = time.time()
            start_time_iso = time.strftime("%Y-%m-%dT%H:%M:%SZ")
            
            # Update stream configuration to enable recording
            path_config = {
                "record": True,
                "recordPath": record_path
            }
            
            if duration:
                path_config["recordDuration"] = duration
                
            async with self._session.post(
                f"{self._base_url}/v3/config/paths/edit/{stream_name}",
                json=path_config
            ) as response:
                if response.status == 200:
                    # Store recording session for duration tracking
                    self._recording_sessions[stream_name] = {
                        "filename": filename,
                        "start_time": start_time,
                        "start_time_iso": start_time_iso,
                        "record_path": record_path,
                        "format": format,
                        "duration": duration,
                        "correlation_id": correlation_id
                    }
                    
                    self._logger.info(f"Started recording for stream {stream_name}: {filename}",
                                    extra={'correlation_id': correlation_id, 'stream_name': stream_name})
                    return {
                        "stream_name": stream_name,
                        "filename": filename,
                        "status": "started",
                        "start_time": start_time_iso,
                        "record_path": record_path,
                        "format": format,
                        "duration": duration
                    }
                else:
                    error_text = await response.text()
                    error_msg = f"Failed to start recording for {stream_name}: HTTP {response.status} - {error_text}"
                    self._logger.error(error_msg, extra={'correlation_id': correlation_id})
                    raise ValueError(error_msg)
                    
        except aiohttp.ClientError as e:
            error_msg = f"MediaMTX unreachable during recording start for {stream_name}: {e}"
            self._logger.error(error_msg, extra={'correlation_id': correlation_id})
            raise ConnectionError(error_msg)
        
    async def stop_recording(self, stream_name: str) -> Dict[str, Any]:
        """
        Stop recording for the specified stream with enhanced error handling and accurate duration calculation.
        
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
                raise ValueError(f"No active recording session found for stream: {stream_name}")
            
            # Calculate recording duration with high precision
            end_time = time.time()
            end_time_iso = time.strftime("%Y-%m-%dT%H:%M:%SZ")
            actual_duration = int(end_time - session["start_time"])
            
            # Update stream configuration to disable recording
            path_config = {
                "record": False
            }
            
            async with self._session.post(
                f"{self._base_url}/v3/config/paths/edit/{stream_name}",
                json=path_config
            ) as response:
                if response.status == 200:
                    # Calculate file size with robust error handling
                    file_path = session["record_path"]
                    file_size = 0
                    file_exists = False
                    
                    if os.path.exists(file_path):
                        try:
                            file_size = os.path.getsize(file_path)
                            file_exists = True
                        except OSError as e:
                            self._logger.warning(f"Could not get file size for {file_path}: {e}",
                                               extra={'correlation_id': correlation_id})
                    else:
                        self._logger.warning(f"Recording file not found: {file_path}",
                                           extra={'correlation_id': correlation_id})
                    
                    # Clean up recording session
                    del self._recording_sessions[stream_name]
                    
                    self._logger.info(
                        f"Stopped recording for stream {stream_name}: duration={actual_duration}s, "
                        f"file_size={file_size}, file_exists={file_exists}",
                        extra={'correlation_id': correlation_id, 'stream_name': stream_name}
                    )
                    
                    return {
                        "stream_name": stream_name,
                        "filename": session["filename"],
                        "status": "completed",
                        "start_time": session["start_time_iso"],
                        "end_time": end_time_iso,
                        "duration": actual_duration,
                        "file_size": file_size,
                        "file_exists": file_exists
                    }
                else:
                    error_text = await response.text()
                    # Keep session for retry - don't delete on API failure
                    error_msg = f"Failed to stop recording for {stream_name}: HTTP {response.status} - {error_text}"
                    self._logger.error(error_msg, extra={'correlation_id': correlation_id})
                    raise ValueError(error_msg)
                    
        except aiohttp.ClientError as e:
            error_msg = f"MediaMTX unreachable during recording stop for {stream_name}: {e}"
            self._logger.error(error_msg, extra={'correlation_id': correlation_id})
            raise ConnectionError(error_msg)

    async def take_snapshot(
        self, 
        stream_name: str, 
        filename: Optional[str] = None
    ) -> Dict[str, Any]:
        """
        Capture a snapshot from the specified stream using FFmpeg with enhanced process management.
        
        Args:
            stream_name: Name of the stream to capture
            filename: Custom filename (None for auto-generated)
            
        Returns:
            Dict containing snapshot information
            
        Raises:
            ValueError: If stream does not exist
            ConnectionError: If MediaMTX is unreachable
        """
        if not self._session:
            raise ConnectionError("MediaMTX controller not started")
            
        if not stream_name:
            raise ValueError("Stream name is required")
            
        correlation_id = get_correlation_id() or str(uuid.uuid4())[:8]
        set_correlation_id(correlation_id)
        
        ffmpeg_process = None
        process_killed = False
        
        try:
            # Generate filename if not provided
            if not filename:
                timestamp = time.strftime("%Y-%m-%d_%H-%M-%S")
                filename = f"{stream_name}_snapshot_{timestamp}.jpg"
            
            # Ensure snapshots directory exists
            os.makedirs(self._snapshots_path, exist_ok=True)
            snapshot_path = os.path.join(self._snapshots_path, filename)
            
            # Use FFmpeg to capture snapshot from RTSP stream
            rtsp_url = f"rtsp://{self._host}:{self._rtsp_port}/{stream_name}"
            
            self._logger.info(f"Capturing snapshot from {rtsp_url} to {snapshot_path}",
                            extra={'correlation_id': correlation_id, 'stream_name': stream_name})
            
            # FFmpeg command to capture single frame with enhanced options
            ffmpeg_cmd = [
                'ffmpeg',
                '-y',  # Overwrite output file
                '-i', rtsp_url,  # Input RTSP stream
                '-vframes', '1',  # Capture only 1 frame
                '-q:v', '2',  # High quality
                '-timeout', '5000000',  # 5 second timeout in microseconds
                '-rtsp_transport', 'tcp',  # Use TCP for reliability
                '-loglevel', 'warning',  # Reduce FFmpeg output
                snapshot_path
            ]
            
            # Execute FFmpeg with enhanced timeout and process management
            ffmpeg_process = await asyncio.wait_for(
                asyncio.create_subprocess_exec(
                    *ffmpeg_cmd,
                    stdout=asyncio.subprocess.PIPE,
                    stderr=asyncio.subprocess.PIPE
                ),
                timeout=10.0  # 10 second timeout for process creation
            )
            
            stdout, stderr = await asyncio.wait_for(
                ffmpeg_process.communicate(),
                timeout=15.0  # 15 second timeout for execution
            )
            
            if ffmpeg_process.returncode == 0 and os.path.exists(snapshot_path):
                try:
                    file_size = os.path.getsize(snapshot_path)
                    self._logger.info(
                        f"Successfully captured snapshot for stream {stream_name}: {filename} ({file_size} bytes)",
                        extra={'correlation_id': correlation_id, 'stream_name': stream_name}
                    )
                    return {
                        "stream_name": stream_name,
                        "filename": filename,
                        "status": "completed",
                        "timestamp": time.strftime("%Y-%m-%dT%H:%M:%SZ"),
                        "file_size": file_size,
                        "file_path": snapshot_path
                    }
                except OSError as e:
                    self._logger.error(f"Could not get file size for snapshot {snapshot_path}: {e}",
                                     extra={'correlation_id': correlation_id})
                    return {
                        "stream_name": stream_name,
                        "filename": filename,
                        "status": "completed",
                        "timestamp": time.strftime("%Y-%m-%dT%H:%M:%SZ"),
                        "file_size": 0,
                        "file_path": snapshot_path,
                        "warning": "Could not determine file size"
                    }
            else:
                # FFmpeg failed - log error and return failure
                error_msg = stderr.decode() if stderr else f"FFmpeg exit code: {ffmpeg_process.returncode}"
                self._logger.error(f"FFmpeg snapshot failed for {stream_name}: {error_msg}",
                                 extra={'correlation_id': correlation_id})
                return {
                    "stream_name": stream_name,
                    "filename": filename,
                    "status": "failed",
                    "timestamp": time.strftime("%Y-%m-%dT%H:%M:%SZ"),
                    "file_size": 0,
                    "error": f"FFmpeg capture failed: {error_msg}"
                }
                
        except asyncio.TimeoutError:
            # Enhanced timeout handling with process cleanup
            if ffmpeg_process and ffmpeg_process.returncode is None:
                try:
                    # Try graceful termination first
                    ffmpeg_process.terminate()
                    await asyncio.wait_for(ffmpeg_process.wait(), timeout=3.0)
                except asyncio.TimeoutError:
                    # Force kill if graceful termination fails
                    try:
                        ffmpeg_process.kill()
                        process_killed = True
                        await ffmpeg_process.wait()
                    except Exception as e:
                        self._logger.error(f"Failed to kill FFmpeg process: {e}",
                                         extra={'correlation_id': correlation_id})
                        
            timeout_context = "killed" if process_killed else "terminated"
            self._logger.error(f"Snapshot capture timeout for {stream_name} (process {timeout_context})",
                             extra={'correlation_id': correlation_id})
            return {
                "stream_name": stream_name,
                "filename": filename or f"{stream_name}_snapshot_timeout.jpg",
                "status": "failed",
                "timestamp": time.strftime("%Y-%m-%dT%H:%M:%SZ"),
                "file_size": 0,
                "error": f"Snapshot capture timeout (process {timeout_context})"
            }
        except Exception as e:
            # Enhanced error handling with process cleanup
            if ffmpeg_process and ffmpeg_process.returncode is None:
                try:
                    ffmpeg_process.terminate()
                    await asyncio.wait_for(ffmpeg_process.wait(), timeout=3.0)
                except:
                    try:
                        ffmpeg_process.kill()
                        process_killed = True
                    except:
                        pass
                        
            error_context = f"process_killed={process_killed}" if process_killed else "normal_cleanup"
            self._logger.error(f"Failed to capture snapshot for {stream_name} ({error_context}): {e}",
                             extra={'correlation_id': correlation_id})
            return {
                "stream_name": stream_name,
                "filename": filename or f"{stream_name}_snapshot_failed.jpg",
                "status": "failed",
                "timestamp": time.strftime("%Y-%m-%dT%H:%M:%SZ"),
                "file_size": 0,
                "error": f"Snapshot capture error: {e}"
            }
        
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
                        for path_name, path_info in data["items"].items():
                            streams.append({
                                "name": path_name,
                                "source": path_info.get("source", ""),
                                "ready": path_info.get("ready", False),
                                "readers": path_info.get("readers", 0),
                                "bytes_sent": path_info.get("bytesSent", 0)
                            })
                    
                    self._logger.debug(f"Retrieved {len(streams)} streams from MediaMTX",
                                     extra={'correlation_id': correlation_id})
                    return streams
                else:
                    error_text = await response.text()
                    error_msg = f"Failed to get stream list: HTTP {response.status} - {error_text}"
                    self._logger.error(error_msg, extra={'correlation_id': correlation_id})
                    raise ConnectionError(error_msg)
                    
        except aiohttp.ClientError as e:
            error_msg = f"MediaMTX unreachable during stream list: {e}"
            self._logger.error(error_msg, extra={'correlation_id': correlation_id})
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
            async with self._session.get(f"{self._base_url}/v3/paths/get/{stream_name}") as response:
                if response.status == 200:
                    data = await response.json()
                    return {
                        "name": stream_name,
                        "status": "active" if data.get("ready", False) else "inactive",
                        "source": data.get("source", ""),
                        "readers": data.get("readers", 0),
                        "bytes_sent": data.get("bytesSent", 0),
                        "recording": data.get("record", False),
                        "correlation_id": correlation_id
                    }
                elif response.status == 404:
                    raise ValueError(f"Stream not found: {stream_name}")
                else:
                    error_text = await response.text()
                    error_msg = f"Failed to get stream status for {stream_name}: HTTP {response.status} - {error_text}"
                    self._logger.error(error_msg, extra={'correlation_id': correlation_id})
                    raise ConnectionError(error_msg)
                    
        except aiohttp.ClientError as e:
            error_msg = f"MediaMTX unreachable during stream status for {stream_name}: {e}"
            self._logger.error(error_msg, extra={'correlation_id': correlation_id})
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
            'logLevel': {'type': str, 'values': ['error', 'warn', 'info', 'debug']},
            'logDestinations': {'type': list},
            'readTimeout': {'type': (str, int), 'min': 1, 'max': 300},
            'writeTimeout': {'type': (str, int), 'min': 1, 'max': 300},
            'readBufferCount': {'type': int, 'min': 1, 'max': 4096},
            'udpMaxPayloadSize': {'type': int, 'min': 1024, 'max': 65507},
            'runOnConnect': {'type': str},
            'runOnConnectRestart': {'type': bool},
            'api': {'type': bool},
            'apiAddress': {'type': str, 'pattern': r'^[0-9.:]+$'},
            'metrics': {'type': bool},
            'metricsAddress': {'type': str, 'pattern': r'^[0-9.:]+$'},
            'pprof': {'type': bool},
            'pprofAddress': {'type': str, 'pattern': r'^[0-9.:]+$'},
            'rtsp': {'type': bool},
            'rtspAddress': {'type': str, 'pattern': r'^[0-9.:]+$'},
            'rtspsAddress': {'type': str, 'pattern': r'^[0-9.:]+$'},
            'rtmp': {'type': bool},
            'rtmpAddress': {'type': str, 'pattern': r'^[0-9.:]+$'},
            'rtmps': {'type': bool},
            'rtmpsAddress': {'type': str, 'pattern': r'^[0-9.:]+$'},
            'hls': {'type': bool},
            'hlsAddress': {'type': str, 'pattern': r'^[0-9.:]+$'},
            'hlsAllowOrigin': {'type': str},
            'webrtc': {'type': bool},
            'webrtcAddress': {'type': str, 'pattern': r'^[0-9.:]+$'}
        }
        
        # Validate configuration keys
        unknown_keys = set(config_updates.keys()) - set(valid_config_schema.keys())
        if unknown_keys:
            raise ValueError(f"Unknown configuration keys: {list(unknown_keys)}")
            
        # Enhanced configuration value validation
        validation_errors = []
        for key, value in config_updates.items():
            schema = valid_config_schema[key]
            
            # Type validation
            if not isinstance(value, schema['type']):
                validation_errors.append(f"Invalid type for {key}: expected {schema['type']}, got {type(value)}")
                continue
            
            # Value constraints validation
            if 'values' in schema and value not in schema['values']:
                validation_errors.append(f"Invalid value for {key}: {value}, allowed values: {schema['values']}")
                
            if 'min' in schema and isinstance(value, (int, float)) and value < schema['min']:
                validation_errors.append(f"Value for {key} too small: {value}, minimum: {schema['min']}")
                
            if 'max' in schema and isinstance(value, (int, float)) and value > schema['max']:
                validation_errors.append(f"Value for {key} too large: {value}, maximum: {schema['max']}")
            
            # Pattern validation for string types
            if 'pattern' in schema and isinstance(value, str):
                import re
                if not re.match(schema['pattern'], value):
                    validation_errors.append(f"Invalid format for {key}: {value}")
        
        if validation_errors:
            error_msg = f"Configuration validation failed: {'; '.join(validation_errors)}"
            self._logger.error(error_msg, extra={'correlation_id': correlation_id})
            raise ValueError(error_msg)
        
        try:
            self._logger.info(f"Updating MediaMTX configuration: {list(config_updates.keys())}",
                            extra={'correlation_id': correlation_id})
            
            async with self._session.post(
                f"{self._base_url}/v3/config/global/patch",
                json=config_updates
            ) as response:
                if response.status == 200:
                    self._logger.info(f"MediaMTX configuration updated successfully: {list(config_updates.keys())}",
                                    extra={'correlation_id': correlation_id})
                    return True
                else:
                    error_text = await response.text()
                    error_msg = f"Failed to update configuration: HTTP {response.status} - {error_text}"
                    self._logger.error(error_msg, extra={'correlation_id': correlation_id})
                    raise ValueError(error_msg)
                    
        except aiohttp.ClientError as e:
            error_msg = f"MediaMTX unreachable during configuration update: {e}"
            self._logger.error(error_msg, extra={'correlation_id': correlation_id})
            raise ConnectionError(error_msg)

    async def _health_monitor_loop(self) -> None:
        """
        Background task for continuous health monitoring with enhanced retry, backoff, and state tracking.
        
        Monitors MediaMTX health and logs status changes with automatic recovery.
        Implements exponential backoff with jitter and detailed state transition logging.
        """
        correlation_id = str(uuid.uuid4())[:8]
        set_correlation_id(correlation_id)
        self._logger.info("Starting MediaMTX health monitoring loop",
                         extra={'correlation_id': correlation_id})
        
        consecutive_failures = 0
        max_failures_before_circuit_break = 10
        circuit_breaker_timeout = 60  # seconds
        circuit_breaker_active = False
        circuit_breaker_start_time = 0
        
        try:
            while self._running:
                try:
                    # Set unique correlation ID for each health check
                    check_correlation_id = str(uuid.uuid4())[:8]
                    set_correlation_id(check_correlation_id)
                    
                    # Check circuit breaker state
                    if circuit_breaker_active:
                        if time.time() - circuit_breaker_start_time > circuit_breaker_timeout:
                            circuit_breaker_active = False
                            consecutive_failures = 0
                            self._health_state['recovery_count'] += 1
                            self._logger.info("Circuit breaker reset - attempting health check recovery",
                                            extra={'correlation_id': check_correlation_id})
                        else:
                            # Skip health check during circuit breaker
                            await asyncio.sleep(10)
                            continue
                    
                    health_status = await self.health_check()
                    current_status = health_status.get("status")
                    
                    # Enhanced status change logging with state transitions
                    if current_status != self._last_health_status:
                        if current_status == "healthy":
                            if self._last_health_status in ["unhealthy", "error"]:
                                self._health_state['recovery_count'] += 1
                                self._logger.info(
                                    f"MediaMTX health RECOVERED: {self._last_health_status} -> {current_status} "
                                    f"(failures reset from {consecutive_failures}, recovery #{self._health_state['recovery_count']})",
                                    extra={'correlation_id': check_correlation_id, 'health_transition': True}
                                )
                            consecutive_failures = 0
                            self._health_state['consecutive_failures'] = 0
                            circuit_breaker_active = False
                        else:
                            consecutive_failures += 1
                            self._health_state['consecutive_failures'] = consecutive_failures
                            self._logger.warning(
                                f"MediaMTX health DEGRADED: {self._last_health_status or 'unknown'} -> {current_status} "
                                f"(consecutive_failures: {consecutive_failures}/{max_failures_before_circuit_break})",
                                extra={'correlation_id': check_correlation_id, 'health_transition': True}
                            )
                            
                            # Activate circuit breaker if too many failures
                            if consecutive_failures >= max_failures_before_circuit_break:
                                circuit_breaker_active = True
                                circuit_breaker_start_time = time.time()
                                self._logger.error(
                                    f"MediaMTX health circuit breaker ACTIVATED after {consecutive_failures} consecutive failures",
                                    extra={'correlation_id': check_correlation_id, 'circuit_breaker': True}
                                )
                        
                        self._last_health_status = current_status
                    
                    # Determine sleep interval based on health status with enhanced backoff
                    if current_status == "healthy":
                        sleep_interval = 30  # Normal interval
                        consecutive_failures = 0
                    else:
                        # Enhanced exponential backoff with jitter: 5s, 10s, 20s, 40s, max 120s
                        import random
                        base_interval = min(5 * (2 ** consecutive_failures), 120)
                        jitter = random.uniform(0.8, 1.2)  # ±20% jitter to avoid thundering herd
                        sleep_interval = base_interval * jitter
                        consecutive_failures += 1
                        
                        self._logger.debug(f"Health monitoring backoff: {sleep_interval:.1f}s (failure #{consecutive_failures})",
                                         extra={'correlation_id': check_correlation_id})
                    
                    await asyncio.sleep(sleep_interval)
                    
                except asyncio.CancelledError:
                    self._logger.info("Health monitoring loop cancelled",
                                    extra={'correlation_id': correlation_id})
                    break
                except Exception as e:
                    consecutive_failures += 1
                    self._health_state['consecutive_failures'] = consecutive_failures
                    self._logger.error(
                        f"Health monitoring error (failure #{consecutive_failures}): {e}",
                        extra={'correlation_id': check_correlation_id},
                        exc_info=True
                    )
                    
                    # Enhanced exponential backoff on errors with circuit breaker
                    if consecutive_failures >= max_failures_before_circuit_break and not circuit_breaker_active:
                        circuit_breaker_active = True
                        circuit_breaker_start_time = time.time()
                        self._logger.error(
                            f"Health monitoring circuit breaker ACTIVATED due to repeated errors",
                            extra={'correlation_id': check_correlation_id, 'circuit_breaker': True}
                        )
                    
                    import random
                    error_sleep = min(2 * (2 ** consecutive_failures), 60) * random.uniform(0.8, 1.2)
                    await asyncio.sleep(error_sleep)
                    
        except Exception as e:
            self._logger.error(f"Critical error in health monitoring loop: {e}",
                             extra={'correlation_id': correlation_id},
                             exc_info=True)
        finally:
            self._logger.info(
                f"Health monitoring loop ended - Final stats: checks={self._health_state['total_checks']}, "
                f"recoveries={self._health_state['recovery_count']}, failures={self._health_state['consecutive_failures']}",
                extra={'correlation_id': correlation_id}
            )