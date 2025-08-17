"""
MediaMTX Camera Service Python SDK Client.

This module provides the main CameraClient class for interacting with the MediaMTX Camera Service.
"""

import asyncio
import json
import logging
import uuid
from typing import Dict, Any, Optional, List, Callable
import websockets
from websockets.exceptions import ConnectionClosed
import ssl

from .exceptions import (
    CameraServiceError,
    AuthenticationError,
    ConnectionError,
    CameraNotFoundError,
    MediaMTXError,
)
from .models import CameraInfo, RecordingInfo, SnapshotInfo


class CameraClient:
    """
    Python client for MediaMTX Camera Service.
    
    Provides a high-level interface for camera control and monitoring
    with support for JWT and API key authentication.
    """

    def __init__(
        self,
        host: str = "localhost",
        port: int = 8080,
        use_ssl: bool = False,
        auth_type: str = "jwt",
        auth_token: Optional[str] = None,
        api_key: Optional[str] = None,
        max_retries: int = 3,
        retry_delay: float = 1.0,
    ):
        """
        Initialize the camera client.

        Args:
            host: Server hostname
            port: Server port
            use_ssl: Whether to use SSL/TLS
            auth_type: Authentication type ('jwt' or 'api_key')
            auth_token: JWT token for authentication
            api_key: API key for authentication
            max_retries: Maximum number of connection retries
            retry_delay: Delay between retries in seconds
        """
        self.host = host
        self.port = port
        self.use_ssl = use_ssl
        self.auth_type = auth_type
        self.auth_token = auth_token
        self.api_key = api_key
        self.max_retries = max_retries
        self.retry_delay = retry_delay
        
        # Connection state
        self.websocket = None
        self.connected = False
        self.authenticated = False
        self.client_id = str(uuid.uuid4())
        
        # Request tracking
        self.request_id = 0
        self.pending_requests: Dict[int, asyncio.Future] = {}
        
        # Event handlers
        self.on_camera_status_update: Optional[Callable[[CameraInfo], None]] = None
        self.on_recording_status_update: Optional[Callable[[RecordingInfo], None]] = None
        self.on_connection_lost: Optional[Callable[[], None]] = None
        
        # Setup logging
        self.logger = logging.getLogger(__name__)

    def _get_ws_url(self) -> str:
        """Get WebSocket URL."""
        protocol = "wss" if self.use_ssl else "ws"
        return f"{protocol}://{self.host}:{self.port}/ws"

    async def _authenticate(self) -> None:
        """
        Authenticate with the camera service using JWT or API key.
        
        Raises:
            AuthenticationError: If authentication fails
        """
        if not self.connected:
            raise ConnectionError("Not connected to camera service")
        
        # Determine token to use
        token = None
        auth_type = "auto"
        
        if self.auth_type == "jwt" and self.auth_token:
            token = self.auth_token
            auth_type = "jwt"
        elif self.auth_type == "api_key" and self.api_key:
            token = self.api_key
            auth_type = "api_key"
        else:
            raise AuthenticationError("No authentication token provided")
        
        try:
            # Send authentication request
            response = await self._send_request("authenticate", {
                "token": token,
                "auth_type": auth_type
            })
            
            if response.get("authenticated"):
                self.authenticated = True
                self.logger.info(f"Authenticated successfully with role: {response.get('role', 'unknown')}")
            else:
                error_msg = response.get("error", "Authentication failed")
                raise AuthenticationError(f"Authentication failed: {error_msg}")
                
        except Exception as e:
            if isinstance(e, AuthenticationError):
                raise
            raise AuthenticationError(f"Authentication error: {e}")

    async def connect(self) -> None:
        """
        Connect to the camera service.
        
        Raises:
            ConnectionError: If connection fails
            AuthenticationError: If authentication fails
        """
        for attempt in range(self.max_retries):
            try:
                self.logger.info(f"Connecting to {self._get_ws_url()} (attempt {attempt + 1})")
                
                # Create SSL context if needed
                ssl_context = None
                if self.use_ssl:
                    ssl_context = ssl.create_default_context()
                    ssl_context.check_hostname = False
                    ssl_context.verify_mode = ssl.CERT_NONE
                
                # Connect to WebSocket
                self.websocket = await websockets.connect(
                    self._get_ws_url(),
                    ssl=ssl_context
                )
                
                self.connected = True
                self.logger.info("Connected to camera service")
                
                # Start message handler
                asyncio.create_task(self._message_handler())
                
                # Authenticate if token provided
                if self.auth_token or self.api_key:
                    await self._authenticate()
                
                # Test connection with ping
                await self.ping()
                self.logger.info("Connection test successful")
                
                return
                
            except Exception as e:
                self.logger.error(f"Connection attempt {attempt + 1} failed: {e}")
                if attempt < self.max_retries - 1:
                    await asyncio.sleep(self.retry_delay * (attempt + 1))
                else:
                    raise ConnectionError(f"Failed to connect after {self.max_retries} attempts: {e}")

    async def disconnect(self) -> None:
        """Disconnect from the camera service."""
        if self.websocket:
            await self.websocket.close()
            self.websocket = None
        self.connected = False
        self.authenticated = False
        self.logger.info("Disconnected from camera service")

    async def _message_handler(self) -> None:
        """Handle incoming WebSocket messages."""
        try:
            async for message in self.websocket:
                await self._process_message(message)
        except ConnectionClosed:
            self.logger.warning("WebSocket connection closed")
            self.connected = False
            if self.on_connection_lost:
                await self.on_connection_lost()
        except Exception as e:
            self.logger.error(f"Error in message handler: {e}")
            self.connected = False

    async def _process_message(self, message: str) -> None:
        """Process incoming message."""
        try:
            data = json.loads(message)
            
            # Handle JSON-RPC response
            if "id" in data and "result" in data:
                await self._handle_response(data)
            # Handle JSON-RPC notification
            elif "method" in data and "id" not in data:
                await self._handle_notification(data)
            else:
                self.logger.warning(f"Unknown message format: {data}")
                
        except json.JSONDecodeError as e:
            self.logger.error(f"Invalid JSON message: {e}")
        except Exception as e:
            self.logger.error(f"Error processing message: {e}")

    async def _handle_response(self, response: Dict[str, Any]) -> None:
        """Handle JSON-RPC response."""
        request_id = response.get("id")
        if request_id in self.pending_requests:
            future = self.pending_requests.pop(request_id)
            if not future.done():
                if "error" in response:
                    future.set_exception(CameraServiceError(response["error"].get("message", "Unknown error")))
                else:
                    future.set_result(response["result"])

    async def _handle_notification(self, notification: Dict[str, Any]) -> None:
        """Handle JSON-RPC notification."""
        method = notification.get("method")
        params = notification.get("params", {})
        
        if method == "camera_status_update" and self.on_camera_status_update:
            camera_info = CameraInfo(
                device_path=params.get("device_path", ""),
                name=params.get("name", ""),
                capabilities=params.get("capabilities", []),
                status=params.get("status", ""),
                stream_url=params.get("stream_url")
            )
            await self.on_camera_status_update(camera_info)
        elif method == "recording_status_update" and self.on_recording_status_update:
            recording_info = RecordingInfo(
                device_path=params.get("device_path", ""),
                recording_id=params.get("recording_id", ""),
                filename=params.get("filename", ""),
                start_time=params.get("start_time", 0),
                duration=params.get("duration"),
                status=params.get("status", "active")
            )
            await self.on_recording_status_update(recording_info)
        else:
            self.logger.info(f"Received notification: {method}")

    async def _send_request(self, method: str, params: Optional[Dict[str, Any]] = None) -> Any:
        """
        Send JSON-RPC request and wait for response.
        
        Args:
            method: RPC method name
            params: Method parameters
            
        Returns:
            Response result
            
        Raises:
            ConnectionError: If not connected
            CameraServiceError: If request fails
        """
        if not self.connected:
            raise ConnectionError("Not connected to camera service")
        
        self.request_id += 1
        request_id = self.request_id
        
        request = {
            "jsonrpc": "2.0",
            "id": request_id,
            "method": method,
            "params": params or {}
        }
        
        # Create future for response
        future = asyncio.Future()
        self.pending_requests[request_id] = future
        
        try:
            # Send request
            await self.websocket.send(json.dumps(request))
            
            # Wait for response with timeout
            result = await asyncio.wait_for(future, timeout=30.0)
            return result
            
        except asyncio.TimeoutError:
            self.pending_requests.pop(request_id, None)
            raise CameraServiceError(f"Request timeout: {method}")
        except Exception as e:
            self.pending_requests.pop(request_id, None)
            raise CameraServiceError(f"Request failed: {e}")

    async def ping(self) -> str:
        """
        Test connection with ping.
        
        Returns:
            "pong" response
            
        Raises:
            CameraServiceError: If ping fails
        """
        return await self._send_request("ping")

    async def get_camera_list(self) -> List[CameraInfo]:
        """
        Get list of available cameras.
        
        Returns:
            List of camera information
            
        Raises:
            CameraServiceError: If request fails
        """
        response = await self._send_request("get_camera_list")
        
        cameras = []
        for camera_data in response:
            camera = CameraInfo(
                device_path=camera_data.get("device_path", ""),
                name=camera_data.get("name", ""),
                capabilities=camera_data.get("capabilities", []),
                status=camera_data.get("status", ""),
                stream_url=camera_data.get("stream_url")
            )
            cameras.append(camera)
        
        return cameras

    async def get_camera_status(self, device_path: str) -> CameraInfo:
        """
        Get camera status.
        
        Args:
            device_path: Camera device path
            
        Returns:
            Camera information
            
        Raises:
            CameraServiceError: If request fails
            CameraNotFoundError: If camera not found
        """
        response = await self._send_request("get_camera_status", {"device": device_path})
        
        if not response:
            raise CameraNotFoundError(f"Camera not found: {device_path}")
        
        return CameraInfo(
            device_path=response.get("device_path", device_path),
            name=response.get("name", ""),
            capabilities=response.get("capabilities", []),
            status=response.get("status", ""),
            stream_url=response.get("stream_url")
        )

    async def take_snapshot(self, device_path: str, filename: Optional[str] = None) -> SnapshotInfo:
        """
        Take camera snapshot.
        
        Args:
            device_path: Camera device path
            filename: Optional custom filename
            
        Returns:
            Snapshot information
            
        Raises:
            CameraServiceError: If request fails
            CameraNotFoundError: If camera not found
        """
        params = {"device": device_path}
        if filename:
            params["filename"] = filename
        
        response = await self._send_request("take_snapshot", params)
        
        return SnapshotInfo(
            device_path=device_path,
            filename=response.get("filename", ""),
            timestamp=response.get("timestamp", 0),
            size_bytes=response.get("size_bytes")
        )

    async def start_recording(self, device_path: str, filename: Optional[str] = None) -> RecordingInfo:
        """
        Start camera recording.
        
        Args:
            device_path: Camera device path
            filename: Optional custom filename
            
        Returns:
            Recording information
            
        Raises:
            CameraServiceError: If request fails
            CameraNotFoundError: If camera not found
        """
        params = {"device": device_path}
        if filename:
            params["filename"] = filename
        
        response = await self._send_request("start_recording", params)
        
        return RecordingInfo(
            device_path=device_path,
            recording_id=response.get("recording_id", ""),
            filename=response.get("filename", ""),
            start_time=response.get("start_time", 0),
            status="active"
        )

    async def stop_recording(self, device_path: str) -> RecordingInfo:
        """
        Stop camera recording.
        
        Args:
            device_path: Camera device path
            
        Returns:
            Recording information
            
        Raises:
            CameraServiceError: If request fails
            CameraNotFoundError: If camera not found
        """
        response = await self._send_request("stop_recording", {"device": device_path})
        
        return RecordingInfo(
            device_path=device_path,
            recording_id=response.get("recording_id", ""),
            filename=response.get("filename", ""),
            start_time=response.get("start_time", 0),
            duration=response.get("duration"),
            status="stopped"
        )

    async def get_recording_status(self, device_path: str) -> Optional[RecordingInfo]:
        """
        Get recording status.
        
        Args:
            device_path: Camera device path
            
        Returns:
            Recording information or None if not recording
            
        Raises:
            CameraServiceError: If request fails
        """
        response = await self._send_request("get_recording_status", {"device": device_path})
        
        if not response:
            return None
        
        return RecordingInfo(
            device_path=device_path,
            recording_id=response.get("recording_id", ""),
            filename=response.get("filename", ""),
            start_time=response.get("start_time", 0),
            duration=response.get("duration"),
            status=response.get("status", "active")
        )
