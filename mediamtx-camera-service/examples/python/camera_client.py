#!/usr/bin/env python3
"""
MediaMTX Camera Service Python Client Example

This example demonstrates how to connect to the MediaMTX Camera Service
using WebSocket JSON-RPC 2.0 protocol with authentication support.

Features:
- JWT and API Key authentication
- WebSocket connection management
- Camera discovery and control
- Snapshot and recording operations
- Real-time status notifications
- Comprehensive error handling
- Retry logic and connection recovery

Usage:
    python camera_client.py --host localhost --port 8080 --auth-type jwt --token your_jwt_token
    python camera_client.py --host localhost --port 8080 --auth-type api_key --key your_api_key
"""

import asyncio
import json
import logging
import uuid
from typing import Dict, Any, Optional, List
from dataclasses import dataclass
import argparse
import websockets
from websockets.exceptions import ConnectionClosed
import ssl


@dataclass
class CameraInfo:
    """Camera device information."""
    device_path: str
    name: str
    capabilities: List[str]
    status: str
    stream_url: Optional[str] = None


@dataclass
class RecordingInfo:
    """Recording session information."""
    device_path: str
    recording_id: str
    filename: str
    start_time: float
    duration: Optional[float] = None
    status: str = "active"


class CameraServiceError(Exception):
    """Base exception for camera service errors."""
    pass


class AuthenticationError(CameraServiceError):
    """Authentication failed."""
    pass


class ConnectionError(CameraServiceError):
    """Connection failed."""
    pass


class CameraNotFoundError(CameraServiceError):
    """Camera device not found."""
    pass


class MediaMTXError(CameraServiceError):
    """MediaMTX operation failed."""
    pass


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
        self.on_camera_status_update = None
        self.on_recording_status_update = None
        self.on_connection_lost = None
        
        # Setup logging
        self.logger = logging.getLogger(__name__)
        self._setup_logging()

    def _setup_logging(self):
        """Setup logging configuration."""
        logging.basicConfig(
            level=logging.INFO,
            format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
        )

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

    def _get_auth_headers(self) -> Dict[str, str]:
        """Get authentication headers."""
        headers = {}
        
        if self.auth_type == "jwt" and self.auth_token:
            headers["Authorization"] = f"Bearer {self.auth_token}"
        elif self.auth_type == "api_key" and self.api_key:
            headers["X-API-Key"] = self.api_key
        
        return headers

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
                    future.set_result(response.get("result"))

    async def _handle_notification(self, notification: Dict[str, Any]) -> None:
        """Handle JSON-RPC notification."""
        method = notification.get("method")
        params = notification.get("params", {})
        
        if method == "camera_status_update" and self.on_camera_status_update:
            await self.on_camera_status_update(params)
        elif method == "recording_status_update" and self.on_recording_status_update:
            await self.on_recording_status_update(params)
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
            CameraServiceError: If request fails
        """
        if not self.connected:
            raise ConnectionError("Not connected to camera service")
        
        self.request_id += 1
        request_id = self.request_id
        
        request = {
            "jsonrpc": "2.0",
            "method": method,
            "id": request_id,
            "params": params or {}
        }
        
        # Create future for response
        future = asyncio.Future()
        self.pending_requests[request_id] = future
        
        try:
            await self.websocket.send(json.dumps(request))
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
        Send ping request to test connection.
        
        Returns:
            Pong response
        """
        return await self._send_request("ping")

    async def get_camera_list(self) -> List[CameraInfo]:
        """
        Get list of available cameras.
        
        Returns:
            List of camera information
        """
        result = await self._send_request("get_camera_list")
        
        cameras = []
        for camera_data in result.get("cameras", []):
            camera = CameraInfo(
                device_path=camera_data["device"],
                name=camera_data["name"],
                capabilities=camera_data.get("capabilities", []),
                status=camera_data["status"],
                stream_url=camera_data.get("stream_url")
            )
            cameras.append(camera)
        
        return cameras

    async def get_camera_status(self, device_path: str) -> CameraInfo:
        """
        Get status of specific camera.
        
        Args:
            device_path: Camera device path
            
        Returns:
            Camera information
            
        Raises:
            CameraNotFoundError: If camera not found
        """
        result = await self._send_request("get_camera_status", {"device_path": device_path})
        
        if not result.get("found"):
            raise CameraNotFoundError(f"Camera not found: {device_path}")
        
        camera_data = result["camera"]
        return CameraInfo(
            device_path=camera_data["device"],
            name=camera_data["name"],
            capabilities=camera_data.get("capabilities", []),
            status=camera_data["status"],
            stream_url=camera_data.get("stream_url")
        )

    async def take_snapshot(
        self,
        device_path: str,
        custom_filename: Optional[str] = None
    ) -> Dict[str, Any]:
        """
        Take a snapshot from camera.
        
        Args:
            device_path: Camera device path
            custom_filename: Optional custom filename
            
        Returns:
            Snapshot information
            
        Raises:
            CameraNotFoundError: If camera not found
            MediaMTXError: If snapshot fails
        """
        params = {"device_path": device_path}
        if custom_filename:
            params["custom_filename"] = custom_filename
        
        result = await self._send_request("take_snapshot", params)
        
        if not result.get("success"):
            error = result.get("error", "Unknown error")
            if "not found" in error.lower():
                raise CameraNotFoundError(f"Camera not found: {device_path}")
            else:
                raise MediaMTXError(f"Snapshot failed: {error}")
        
        return result

    async def start_recording(
        self,
        device_path: str,
        duration: Optional[int] = None,
        custom_filename: Optional[str] = None
    ) -> RecordingInfo:
        """
        Start recording from camera.
        
        Args:
            device_path: Camera device path
            duration: Recording duration in seconds (optional)
            custom_filename: Optional custom filename
            
        Returns:
            Recording information
            
        Raises:
            CameraNotFoundError: If camera not found
            MediaMTXError: If recording fails
        """
        params = {"device_path": device_path}
        if duration:
            params["duration"] = duration
        if custom_filename:
            params["custom_filename"] = custom_filename
        
        result = await self._send_request("start_recording", params)
        
        if not result.get("success"):
            error = result.get("error", "Unknown error")
            if "not found" in error.lower():
                raise CameraNotFoundError(f"Camera not found: {device_path}")
            else:
                raise MediaMTXError(f"Recording failed: {error}")
        
        recording_data = result["recording"]
        return RecordingInfo(
            device_path=recording_data["device_path"],
            recording_id=recording_data["recording_id"],
            filename=recording_data["filename"],
            start_time=recording_data["start_time"],
            status=recording_data["status"]
        )

    async def stop_recording(self, device_path: str) -> Dict[str, Any]:
        """
        Stop recording from camera.
        
        Args:
            device_path: Camera device path
            
        Returns:
            Recording stop information
            
        Raises:
            CameraNotFoundError: If camera not found
            MediaMTXError: If stop recording fails
        """
        result = await self._send_request("stop_recording", {"device_path": device_path})
        
        if not result.get("success"):
            error = result.get("error", "Unknown error")
            if "not found" in error.lower():
                raise CameraNotFoundError(f"Camera not found: {device_path}")
            else:
                raise MediaMTXError(f"Stop recording failed: {error}")
        
        return result

    def set_camera_status_callback(self, callback):
        """Set callback for camera status updates."""
        self.on_camera_status_update = callback

    def set_recording_status_callback(self, callback):
        """Set callback for recording status updates."""
        self.on_recording_status_update = callback

    def set_connection_lost_callback(self, callback):
        """Set callback for connection lost events."""
        self.on_connection_lost = callback


async def main():
    """Example usage of the camera client."""
    parser = argparse.ArgumentParser(description="MediaMTX Camera Service Python Client")
    parser.add_argument("--host", default="localhost", help="Server hostname")
    parser.add_argument("--port", type=int, default=8080, help="Server port")
    parser.add_argument("--ssl", action="store_true", help="Use SSL/TLS")
    parser.add_argument("--auth-type", choices=["jwt", "api_key"], default="jwt", help="Authentication type")
    parser.add_argument("--token", help="JWT token")
    parser.add_argument("--key", help="API key")
    
    args = parser.parse_args()
    
    # Create client
    client = CameraClient(
        host=args.host,
        port=args.port,
        use_ssl=args.ssl,
        auth_type=args.auth_type,
        auth_token=args.token,
        api_key=args.key
    )
    
    try:
        # Connect to service
        await client.connect()
        print("✅ Connected to camera service")
        
        # Test ping
        pong = await client.ping()
        print(f"✅ Ping response: {pong}")
        
        # Get camera list
        cameras = await client.get_camera_list()
        print(f"✅ Found {len(cameras)} cameras:")
        for camera in cameras:
            print(f"  - {camera.name} ({camera.device_path}) - {camera.status}")
        
        if cameras:
            # Get status of first camera
            camera = cameras[0]
            status = await client.get_camera_status(camera.device_path)
            print(f"✅ Camera status: {status.status}")
            
            # Take snapshot
            snapshot = await client.take_snapshot(camera.device_path)
            print(f"✅ Snapshot taken: {snapshot['filename']}")
            
            # Start recording
            recording = await client.start_recording(camera.device_path, duration=10)
            print(f"✅ Recording started: {recording.filename}")
            
            # Wait a bit
            await asyncio.sleep(5)
            
            # Stop recording
            stop_result = await client.stop_recording(camera.device_path)
            print(f"✅ Recording stopped: {stop_result['filename']}")
        
    except CameraServiceError as e:
        print(f"❌ Camera service error: {e}")
    except Exception as e:
        print(f"❌ Error: {e}")
    finally:
        await client.disconnect()
        print("✅ Disconnected")


if __name__ == "__main__":
    asyncio.run(main()) 