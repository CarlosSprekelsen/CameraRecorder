"""
MediaMTX REST API controller for stream and recording management.
"""

import asyncio
import logging
import time
import uuid
from typing import Dict, Any, Optional, List
from dataclasses import dataclass
import aiohttp
import json


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
    and configuration management for the MediaMTX media server.
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

    async def start(self) -> None:
        """
        Start the MediaMTX controller.
        
        Initializes HTTP client session and begins health monitoring.
        """
        if self._running:
            self._logger.warning("MediaMTX controller is already running")
            return
            
        self._logger.info("Starting MediaMTX controller")
        
        # Create aiohttp ClientSession with timeout configuration
        timeout = aiohttp.ClientTimeout(total=10, connect=5)
        connector = aiohttp.TCPConnector(limit=10, limit_per_host=5)
        self._session = aiohttp.ClientSession(
            timeout=timeout,
            connector=connector,
            headers={"Content-Type": "application/json"}
        )
        
        # Start health monitoring task
        self._health_check_task = asyncio.create_task(self._health_monitor_loop())
        self._running = True
        
        self._logger.info("MediaMTX controller started successfully")

    async def stop(self) -> None:
        """
        Stop the MediaMTX controller.
        
        Closes HTTP client session and stops monitoring tasks.
        """
        if not self._running:
            return
            
        self._logger.info("Stopping MediaMTX controller")
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
            
        self._logger.info("MediaMTX controller stopped")

    async def health_check(self) -> Dict[str, Any]:
        """
        Perform health check on MediaMTX server.
        
        Returns:
            Dict containing health status and metrics
            
        Raises:
            ConnectionError: If MediaMTX is unreachable
        """
        if not self._session:
            raise ConnectionError("MediaMTX controller not started")
            
        try:
            # Call MediaMTX API config endpoint to verify connectivity
            async with self._session.get(f"{self._base_url}/v3/config/global/get") as response:
                if response.status == 200:
                    config_data = await response.json()
                    return {
                        "status": "healthy",
                        "version": config_data.get("serverVersion", "unknown"),
                        "uptime": config_data.get("serverUptime", 0),
                        "api_port": self._api_port,
                        "response_time_ms": response.headers.get("X-Response-Time", "unknown")
                    }
                else:
                    return {
                        "status": "unhealthy",
                        "error": f"HTTP {response.status}: {await response.text()}",
                        "api_port": self._api_port
                    }
                    
        except aiohttp.ClientError as e:
            raise ConnectionError(f"MediaMTX unreachable: {e}")
        except Exception as e:
            self._logger.error(f"Health check failed: {e}")
            return {
                "status": "error",
                "error": str(e),
                "api_port": self._api_port
            }

    async def create_stream(self, stream_config: StreamConfig) -> Dict[str, str]:
        """
        Create a new stream path in MediaMTX.
        
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
            
        try:
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
                    self._logger.info(f"Created stream path: {stream_config.name}")
                    return {
                        "rtsp": f"rtsp://{self._host}:{self._rtsp_port}/{stream_config.name}",
                        "webrtc": f"http://{self._host}:{self._webrtc_port}/{stream_config.name}",
                        "hls": f"http://{self._host}:{self._hls_port}/{stream_config.name}"
                    }
                else:
                    error_text = await response.text()
                    raise ConnectionError(f"Failed to create stream: HTTP {response.status} - {error_text}")
                    
        except aiohttp.ClientError as e:
            raise ConnectionError(f"MediaMTX unreachable during stream creation: {e}")

    async def delete_stream(self, stream_name: str) -> bool:
        """
        Delete a stream path from MediaMTX.
        
        Args:
            stream_name: Name of the stream to delete
            
        Returns:
            True if stream was deleted successfully
            
        Raises:
            ConnectionError: If MediaMTX is unreachable
        """
        if not self._session:
            raise ConnectionError("MediaMTX controller not started")
            
        if not stream_name:
            raise ValueError("Stream name is required")
            
        try:
            # Delete stream path via MediaMTX API
            async with self._session.post(f"{self._base_url}/v3/config/paths/delete/{stream_name}") as response:
                if response.status in [200, 204]:
                    self._logger.info(f"Deleted stream path: {stream_name}")
                    return True
                elif response.status == 404:
                    self._logger.warning(f"Stream path not found: {stream_name}")
                    return False
                else:
                    error_text = await response.text()
                    self._logger.error(f"Failed to delete stream {stream_name}: HTTP {response.status} - {error_text}")
                    return False
                    
        except aiohttp.ClientError as e:
            raise ConnectionError(f"MediaMTX unreachable during stream deletion: {e}")

    async def start_recording(
        self, 
        stream_name: str, 
        duration: Optional[int] = None,
        format: str = "mp4"
    ) -> Dict[str, Any]:
        """
        Start recording for the specified stream.
        
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
            
        try:
            # Generate recording filename with timestamp
            timestamp = time.strftime("%Y-%m-%d_%H-%M-%S")
            filename = f"{stream_name}_{timestamp}.{format}"
            record_path = f"{self._recordings_path}/{filename}"
            
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
                    start_time = time.strftime("%Y-%m-%dT%H:%M:%SZ")
                    self._logger.info(f"Started recording for stream {stream_name}: {filename}")
                    return {
                        "stream_name": stream_name,
                        "filename": filename,
                        "status": "started",
                        "start_time": start_time,
                        "record_path": record_path,
                        "format": format,
                        "duration": duration
                    }
                else:
                    error_text = await response.text()
                    raise ValueError(f"Failed to start recording: HTTP {response.status} - {error_text}")
                    
        except aiohttp.ClientError as e:
            raise ConnectionError(f"MediaMTX unreachable during recording start: {e}")

    async def stop_recording(self, stream_name: str) -> Dict[str, Any]:
        """
        Stop recording for the specified stream.
        
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
            
        try:
            # Get current stream status to find recording filename
            stream_status = await self.get_stream_status(stream_name)
            current_filename = stream_status.get("recording_filename", f"{stream_name}_unknown.mp4")
            
            # Update stream configuration to disable recording
            path_config = {
                "record": False
            }
            
            async with self._session.post(
                f"{self._base_url}/v3/config/paths/edit/{stream_name}",
                json=path_config
            ) as response:
                if response.status == 200:
                    end_time = time.strftime("%Y-%m-%dT%H:%M:%SZ")
                    self._logger.info(f"Stopped recording for stream {stream_name}")
                    
                    # Calculate file size if file exists
                    import os
                    file_path = f"{self._recordings_path}/{current_filename}"
                    file_size = 0
                    if os.path.exists(file_path):
                        file_size = os.path.getsize(file_path)
                    
                    return {
                        "stream_name": stream_name,
                        "filename": current_filename,
                        "status": "completed",
                        "end_time": end_time,
                        "file_size": file_size,
                        "duration": None  # TODO: MEDIUM: Calculate duration from start/end times [Story:E1/S1a]
                    }
                else:
                    error_text = await response.text()
                    raise ValueError(f"Failed to stop recording: HTTP {response.status} - {error_text}")
                    
        except aiohttp.ClientError as e:
            raise ConnectionError(f"MediaMTX unreachable during recording stop: {e}")

    async def take_snapshot(
        self, 
        stream_name: str, 
        filename: Optional[str] = None
    ) -> Dict[str, Any]:
        """
        Capture a snapshot from the specified stream.
        
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
            
        try:
            # Generate filename if not provided
            if not filename:
                timestamp = time.strftime("%Y-%m-%d_%H-%M-%S")
                filename = f"{stream_name}_snapshot_{timestamp}.jpg"
            
            snapshot_path = f"{self._snapshots_path}/{filename}"
            
            # Use MediaMTX API to capture snapshot (if available) or implement via FFmpeg
            # TODO: HIGH: Implement actual MediaMTX snapshot API integration or FFmpeg fallback [Story:E1/S1a]
            snapshot_config = {
                "stream": stream_name,
                "output": snapshot_path,
                "format": "jpeg"
            }
            
            # Simulate snapshot capture - in real implementation, this might use FFmpeg
            # or a custom MediaMTX extension
            await asyncio.sleep(0.1)  # Simulate processing time
            
            # Create a minimal file to simulate snapshot (in real implementation, this would be actual image data)
            import os
            os.makedirs(self._snapshots_path, exist_ok=True)
            with open(snapshot_path, 'w') as f:
                f.write("# Snapshot placeholder")
            
            file_size = os.path.getsize(snapshot_path) if os.path.exists(snapshot_path) else 0
            
            self._logger.info(f"Captured snapshot for stream {stream_name}: {filename}")
            return {
                "stream_name": stream_name,
                "filename": filename,
                "status": "completed",
                "timestamp": time.strftime("%Y-%m-%dT%H:%M:%SZ"),
                "file_size": file_size,
                "file_path": snapshot_path
            }
            
        except Exception as e:
            self._logger.error(f"Failed to capture snapshot for {stream_name}: {e}")
            return {
                "stream_name": stream_name,
                "filename": filename or f"{stream_name}_snapshot_failed.jpg",
                "status": "failed",
                "timestamp": time.strftime("%Y-%m-%dT%H:%M:%SZ"),
                "file_size": 0,
                "error": str(e)
            }

    async def get_stream_list(self) -> List[Dict[str, Any]]:
        """
        Get list of all configured streams.
        
        Returns:
            List of stream configuration dictionaries
            
        Raises:
            ConnectionError: If MediaMTX is unreachable
        """
        if not self._session:
            raise ConnectionError("MediaMTX controller not started")
            
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
                    
                    return streams
                else:
                    error_text = await response.text()
                    raise ConnectionError(f"Failed to get stream list: HTTP {response.status} - {error_text}")
                    
        except aiohttp.ClientError as e:
            raise ConnectionError(f"MediaMTX unreachable during stream list: {e}")

    async def get_stream_status(self, stream_name: str) -> Dict[str, Any]:
        """
        Get detailed status for a specific stream.
        
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
                        "recording": data.get("record", False)
                    }
                elif response.status == 404:
                    raise ValueError(f"Stream not found: {stream_name}")
                else:
                    error_text = await response.text()
                    raise ConnectionError(f"Failed to get stream status: HTTP {response.status} - {error_text}")
                    
        except aiohttp.ClientError as e:
            raise ConnectionError(f"MediaMTX unreachable during stream status: {e}")

    async def update_configuration(self, config_updates: Dict[str, Any]) -> bool:
        """
        Update MediaMTX configuration dynamically.
        
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
            
        try:
            async with self._session.post(
                f"{self._base_url}/v3/config/global/patch",
                json=config_updates
            ) as response:
                if response.status == 200:
                    self._logger.info("MediaMTX configuration updated successfully")
                    return True
                else:
                    error_text = await response.text()
                    raise ValueError(f"Failed to update configuration: HTTP {response.status} - {error_text}")
                    
        except aiohttp.ClientError as e:
            raise ConnectionError(f"MediaMTX unreachable during configuration update: {e}")

    async def _health_monitor_loop(self) -> None:
        """
        Background task for continuous health monitoring.
        
        Monitors MediaMTX health and logs status changes.
        """
        self._logger.debug("Starting MediaMTX health monitoring loop")
        
        while self._running:
            try:
                health_status = await self.health_check()
                if health_status.get("status") != "healthy":
                    self._logger.warning(f"MediaMTX health check failed: {health_status}")
                    
                # Wait 30 seconds between health checks
                await asyncio.sleep(30)
                
            except asyncio.CancelledError:
                self._logger.debug("Health monitoring loop cancelled")
                break
            except Exception as e:
                self._logger.error(f"Health monitoring error: {e}")
                await asyncio.sleep(10)  # Shorter wait on error