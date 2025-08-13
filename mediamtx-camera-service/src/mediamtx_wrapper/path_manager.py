# src/mediamtx_wrapper/path_manager.py
"""
MediaMTX Path Manager for API-driven path creation with FFmpeg publishing.

This module provides the MediaMTXPathManager class that manages MediaMTX path creation
via the MediaMTX API using FFmpeg commands for device publishing.
"""

import asyncio
import logging
import aiohttp
from typing import Dict, Any, Optional


class MediaMTXPathManager:
    """Manages MediaMTX path creation via API with FFmpeg publishing."""

    def __init__(self, mediamtx_host: str = "localhost", mediamtx_port: int = 9997):
        """Initialize MediaMTX path manager."""
        self.api_base = f"http://{mediamtx_host}:{mediamtx_port}/v3"
        self._logger = logging.getLogger(__name__)
        self._session: Optional[aiohttp.ClientSession] = None

    async def start(self) -> None:
        """Start the path manager and create HTTP session."""
        if self._session is None:
            self._session = aiohttp.ClientSession()
            self._logger.info("MediaMTX Path Manager started")

    async def stop(self) -> None:
        """Stop the path manager and close HTTP session."""
        if self._session:
            await self._session.close()
            self._session = None
            self._logger.info("MediaMTX Path Manager stopped")

    async def create_camera_path(self, camera_id: str, device_path: str, rtsp_port: int = 8554) -> bool:
        """
        Create MediaMTX path for camera with FFmpeg publishing.
        
        Args:
            camera_id: Camera identifier (e.g., "0", "1", "2", "3")
            device_path: Device path (e.g., "/dev/video0")
            rtsp_port: RTSP port for MediaMTX
            
        Returns:
            True if path creation successful, False otherwise
        """
        if not self._session:
            self._logger.error("Path manager not started")
            return False

        path_name = f"cam{camera_id}"
        ffmpeg_command = (
            f"ffmpeg -f v4l2 -i {device_path} -c:v libx264 -pix_fmt yuv420p "
            f"-preset ultrafast -b:v 600k -f rtsp rtsp://127.0.0.1:{rtsp_port}/{path_name}"
        )
        
        payload = {
            "runOnDemand": ffmpeg_command,
            "runOnDemandRestart": True
        }
        
        try:
            async with self._session.post(
                f"{self.api_base}/config/paths/add/{path_name}",
                json=payload,
                headers={"Content-Type": "application/json"}
            ) as response:
                if response.status == 200:
                    self._logger.info(f"Successfully created MediaMTX path: {path_name}")
                    return True
                else:
                    error_text = await response.text()
                    self._logger.error(
                        f"Failed to create path {path_name}: HTTP {response.status} - {error_text}"
                    )
                    return False
        except Exception as e:
            self._logger.error(f"Failed to create path {path_name}: {e}")
            return False

    async def delete_camera_path(self, camera_id: str) -> bool:
        """
        Delete MediaMTX path for camera.
        
        Args:
            camera_id: Camera identifier (e.g., "0", "1", "2", "3")
            
        Returns:
            True if path deletion successful, False otherwise
        """
        if not self._session:
            self._logger.error("Path manager not started")
            return False

        path_name = f"cam{camera_id}"
        
        try:
            async with self._session.delete(f"{self.api_base}/config/paths/delete/{path_name}") as response:
                if response.status == 200:
                    self._logger.info(f"Successfully deleted MediaMTX path: {path_name}")
                    return True
                else:
                    error_text = await response.text()
                    self._logger.error(
                        f"Failed to delete path {path_name}: HTTP {response.status} - {error_text}"
                    )
                    return False
        except Exception as e:
            self._logger.error(f"Failed to delete path {path_name}: {e}")
            return False

    async def verify_path_exists(self, camera_id: str) -> bool:
        """
        Verify that MediaMTX path exists.
        
        Args:
            camera_id: Camera identifier (e.g., "0", "1", "2", "3")
            
        Returns:
            True if path exists, False otherwise
        """
        if not self._session:
            self._logger.error("Path manager not started")
            return False

        path_name = f"cam{camera_id}"
        
        try:
            async with self._session.get(f"{self.api_base}/paths/list") as response:
                if response.status == 200:
                    data = await response.json()
                    return any(item["name"] == path_name for item in data["items"])
                else:
                    self._logger.error(f"Failed to get paths list: HTTP {response.status}")
                    return False
        except Exception as e:
            self._logger.error(f"Failed to verify path {path_name}: {e}")
            return False

    async def get_path_status(self, camera_id: str) -> Optional[Dict[str, Any]]:
        """
        Get status of MediaMTX path.
        
        Args:
            camera_id: Camera identifier (e.g., "0", "1", "2", "3")
            
        Returns:
            Path status dictionary or None if not found
        """
        if not self._session:
            self._logger.error("Path manager not started")
            return None

        path_name = f"cam{camera_id}"
        
        try:
            async with self._session.get(f"{self.api_base}/paths/list") as response:
                if response.status == 200:
                    data = await response.json()
                    for item in data["items"]:
                        if item["name"] == path_name:
                            return item
                    return None
                else:
                    self._logger.error(f"Failed to get paths list: HTTP {response.status}")
                    return None
        except Exception as e:
            self._logger.error(f"Failed to get path status for {path_name}: {e}")
            return None

    async def list_all_paths(self) -> Dict[str, Any]:
        """
        Get list of all MediaMTX paths.
        
        Returns:
            Dictionary containing all paths information
        """
        if not self._session:
            self._logger.error("Path manager not started")
            return {"items": []}

        try:
            async with self._session.get(f"{self.api_base}/paths/list") as response:
                if response.status == 200:
                    return await response.json()
                else:
                    self._logger.error(f"Failed to get paths list: HTTP {response.status}")
                    return {"items": []}
        except Exception as e:
            self._logger.error(f"Failed to list paths: {e}")
            return {"items": []}
