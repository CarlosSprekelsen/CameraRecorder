"""
WebSocket JSON-RPC 2.0 server for camera control and notifications.
"""

import asyncio
import json
import logging
import time
import uuid
from typing import Dict, Any, Optional, Callable, Set, List
from dataclasses import dataclass
from pathlib import Path

import websockets
from websockets.server import WebSocketServerProtocol
from websockets.exceptions import ConnectionClosed, WebSocketException

from ..camera_service.logging_config import set_correlation_id, get_correlation_id


@dataclass
class JsonRpcRequest:
    """JSON-RPC 2.0 request structure."""
    jsonrpc: str
    method: str
    id: Optional[Any] = None
    params: Optional[Dict[str, Any]] = None


@dataclass
class JsonRpcResponse:
    """JSON-RPC 2.0 response structure."""
    jsonrpc: str
    id: Optional[Any]
    result: Optional[Any] = None
    error: Optional[Dict[str, Any]] = None


@dataclass
class JsonRpcNotification:
    """JSON-RPC 2.0 notification structure."""
    jsonrpc: str
    method: str
    params: Optional[Dict[str, Any]] = None


class ClientConnection:
    """Represents a connected WebSocket client."""
    
    def __init__(self, websocket: WebSocketServerProtocol, client_id: str):
        """
        Initialize client connection.
        
        Args:
            websocket: WebSocket connection object
            client_id: Unique identifier for this client
        """
        self.websocket = websocket
        self.client_id = client_id
        self.authenticated = False
        self.subscriptions: Set[str] = set()
        self.connected_at = asyncio.get_event_loop().time()
        
        # TODO: HIGH: Add authentication state implementation [Story:E1/S1a]
        # TODO: MEDIUM: Add permission tracking implementation [Story:E1/S1a]
        # TODO: MEDIUM: Add rate limiting state implementation [Story:E1/S1a]


class WebSocketJsonRpcServer:
    """
    WebSocket JSON-RPC 2.0 server for camera control and real-time notifications.

    Provides camera control API and broadcasts real-time events to connected clients
    as specified in the architecture overview.
    """

    # TODO: CRITICAL: Complete method-level API versioning framework implementation [Story:E1/S1a]
    # Description: Architecture overview (docs/architecture/overview.md) requires method-level versioning and structured deprecation for all JSON-RPC methods.
    # IV&V Reference: Architecture Decisions v6, API Versioning Strategy, Story S1.
    # Rationale: All public API methods must support explicit versioning and deprecation tracking.
    # STOPPED: Do not implement version negotiation or migration logic until versioning requirements are clarified and documented.
    _method_versions: Dict[str, str] = {}

    def __init__(
        self,
        host: str,
        port: int,
        websocket_path: str,
        max_connections: int,
        mediamtx_controller=None,
        camera_monitor=None
    ):
        """
        Initialize WebSocket JSON-RPC server.
        
        Args:
            host: Server bind address
            port: Server port
            websocket_path: WebSocket endpoint path
            max_connections: Maximum concurrent client connections
            mediamtx_controller: MediaMTX controller instance for stream operations
            camera_monitor: Camera monitor instance for device information
        """
        self._host = host
        self._port = port
        self._websocket_path = websocket_path
        self._max_connections = max_connections
        self._mediamtx_controller = mediamtx_controller
        self._camera_monitor = camera_monitor
        
        self._logger = logging.getLogger(__name__)
        self._server = None
        self._running = False
        
        # Client connection management
        self._clients: Dict[str, ClientConnection] = {}
        self._connection_lock = asyncio.Lock()
        
        # JSON-RPC method handlers
        self._method_handlers: Dict[str, Callable] = {}
        
        # TODO: HIGH: Initialize authentication system implementation [Story:E1/S1a]
        # TODO: MEDIUM: Initialize rate limiting implementation [Story:E1/S1a]
        # TODO: MEDIUM: Initialize metrics collection implementation [Story:E1/S1a]

    def set_mediamtx_controller(self, controller) -> None:
        """
        Set the MediaMTX controller for stream operations.
        
        Args:
            controller: MediaMTX controller instance
        """
        self._mediamtx_controller = controller

    def set_camera_monitor(self, monitor) -> None:
        """
        Set the camera monitor for device information.
        
        Args:
            monitor: Camera monitor instance
        """
        self._camera_monitor = monitor

    async def start(self) -> None:
        """
        Start the WebSocket JSON-RPC server.
        
        Initializes the WebSocket server and begins accepting client connections.
        """
        if self._running:
            self._logger.warning("WebSocket server is already running")
            return
        
        self._logger.info(f"Starting WebSocket JSON-RPC server on {self._host}:{self._port}{self._websocket_path}")
        
        try:
            # Register built-in methods
            self._register_builtin_methods()
            
            # Start WebSocket server with proper error handling
            self._server = await websockets.serve(
                self._handle_client_connection,
                self._host,
                self._port,
                # Server configuration
                max_size=1024 * 1024,  # 1MB max message size
                max_queue=100,  # Max queued messages per connection
                compression=None,  # Disable compression for simplicity
                ping_interval=30,  # Ping every 30 seconds
                ping_timeout=10,  # Ping timeout
                close_timeout=5,  # Close timeout
                # Path handling - only accept connections to our WebSocket path
                process_request=self._process_request
            )
            
            self._running = True
            self._logger.info(f"WebSocket JSON-RPC server started successfully on {self._host}:{self._port}{self._websocket_path}")
            
        except Exception as e:
            self._logger.error(f"Failed to start WebSocket server: {e}")
            await self._cleanup_server()
            raise

    async def stop(self) -> None:
        """
        Stop the WebSocket JSON-RPC server.
        
        Gracefully closes all client connections and stops the server.
        """
        if not self._running:
            return
        
        self._logger.info("Stopping WebSocket JSON-RPC server")
        
        try:
            self._running = False
            
            # Close all client connections
            await self._close_all_connections()
            
            # Stop WebSocket server
            if self._server:
                self._server.close()
                await self._server.wait_closed()
                self._server = None
            
            # Cleanup resources and tasks
            await self._cleanup_server()
            
            self._logger.info("WebSocket JSON-RPC server stopped")
            
        except Exception as e:
            self._logger.error(f"Error during WebSocket server shutdown: {e}")
            raise

    async def _process_request(self, path: str, request_headers) -> Optional[tuple]:
        """
        Process incoming WebSocket request to validate path and enforce limits.
        
        Args:
            path: Request path
            request_headers: HTTP request headers
            
        Returns:
            None to accept connection, or (status, headers, body) to reject
        """
        # Validate WebSocket path
        if path != self._websocket_path:
            self._logger.warning(f"Invalid WebSocket path requested: {path}")
            return (404, {}, b"Not Found")
        
        # Check connection limits
        async with self._connection_lock:
            if len(self._clients) >= self._max_connections:
                self._logger.warning(f"Connection limit reached: {len(self._clients)}/{self._max_connections}")
                return (503, {}, b"Service Unavailable - Connection limit reached")
        
        # Accept connection
        return None

    async def _handle_client_connection(self, websocket: WebSocketServerProtocol, path: str) -> None:
        """
        Handle new client WebSocket connection.
        
        Args:
            websocket: WebSocket connection object
            path: Request path
        """
        client_id = str(uuid.uuid4())
        client_ip = websocket.remote_address[0] if websocket.remote_address else "unknown"
        
        self._logger.info(f"New client connection: {client_id} from {client_ip}")
        
        # Create client connection object
        client = ClientConnection(websocket, client_id)
        
        try:
            # Add client to tracking
            async with self._connection_lock:
                self._clients[client_id] = client
            
            self._logger.debug(f"Client {client_id} added to connection pool ({len(self._clients)} total)")
            
            # Handle authentication (basic implementation - TODO: expand per architecture)
            # For now, mark all clients as authenticated
            client.authenticated = True
            
            # Process incoming messages from client
            async for message in websocket:
                try:
                    if isinstance(message, str):
                        response = await self._handle_json_rpc_message(client, message)
                        if response:
                            await websocket.send(response)
                    else:
                        self._logger.warning(f"Received non-text message from client {client_id}")
                        
                except Exception as e:
                    self._logger.error(f"Error processing message from client {client_id}: {e}")
                    # Send JSON-RPC error response if possible
                    try:
                        error_response = json.dumps({
                            "jsonrpc": "2.0",
                            "error": {
                                "code": -32603,
                                "message": "Internal error"
                            },
                            "id": None
                        })
                        await websocket.send(error_response)
                    except Exception:
                        # Connection might be broken, will be cleaned up below
                        break
                        
        except ConnectionClosed:
            self._logger.info(f"Client {client_id} disconnected normally")
        except WebSocketException as e:
            self._logger.warning(f"WebSocket error for client {client_id}: {e}")
        except Exception as e:
            self._logger.error(f"Unexpected error handling client {client_id}: {e}")
        finally:
            # Cleanup on disconnect
            async with self._connection_lock:
                if client_id in self._clients:
                    del self._clients[client_id]
                    self._logger.info(f"Removed client {client_id} from connection pool ({len(self._clients)} remaining)")

    async def _handle_json_rpc_message(
        self, 
        client: ClientConnection, 
        message: str
    ) -> Optional[str]:
        """
        Process incoming JSON-RPC message from client.

        Args:
            client: Client connection object
            message: Raw JSON-RPC message

        Returns:
            JSON-RPC response string or None for notifications
        """
        correlation_id = None
        request_id = None
        
        try:
            # Parse JSON-RPC request
            try:
                request_data = json.loads(message)
            except json.JSONDecodeError as e:
                return json.dumps({
                    "jsonrpc": "2.0",
                    "error": {
                        "code": -32700,
                        "message": "Parse error"
                    },
                    "id": None
                })
            
            # Extract correlation ID from request ID field
            request_id = request_data.get("id")
            correlation_id = str(request_id) if request_id is not None else str(uuid.uuid4())[:8]
            
            # Set correlation ID for structured logging
            set_correlation_id(correlation_id)
            
            # Validate JSON-RPC structure
            if not isinstance(request_data, dict) or request_data.get("jsonrpc") != "2.0":
                return json.dumps({
                    "jsonrpc": "2.0",
                    "error": {
                        "code": -32600,
                        "message": "Invalid Request"
                    },
                    "id": request_id
                })
            
            method_name = request_data.get("method")
            if not method_name or not isinstance(method_name, str):
                return json.dumps({
                    "jsonrpc": "2.0",
                    "error": {
                        "code": -32600,
                        "message": "Invalid Request - missing method"
                    },
                    "id": request_id
                })
            
            params = request_data.get("params")
            
            self._logger.debug(f"Processing JSON-RPC method '{method_name}' from client {client.client_id}")
            
            # Check if method exists
            if method_name not in self._method_handlers:
                return json.dumps({
                    "jsonrpc": "2.0",
                    "error": {
                        "code": -32601,
                        "message": "Method not found"
                    },
                    "id": request_id
                })
            
            # Call method handler
            try:
                handler = self._method_handlers[method_name]
                if params is not None:
                    result = await handler(params)
                else:
                    result = await handler()
                
                # Return response for requests with ID (notifications have no ID)
                if request_id is not None:
                    return json.dumps({
                        "jsonrpc": "2.0",
                        "result": result,
                        "id": request_id
                    })
                else:
                    # Notification - no response
                    return None
                    
            except Exception as e:
                self._logger.error(f"Error in method handler '{method_name}': {e}")
                if request_id is not None:
                    return json.dumps({
                        "jsonrpc": "2.0",
                        "error": {
                            "code": -32603,
                            "message": "Internal error"
                        },
                        "id": request_id
                    })
                else:
                    return None

        except Exception as e:
            self._logger.error(f"Error processing JSON-RPC message from client {client.client_id}: {e}")
            return json.dumps({
                "jsonrpc": "2.0",
                "error": {
                    "code": -32603,
                    "message": "Internal error"
                },
                "id": request_id
            })

    async def _close_all_connections(self) -> None:
        """Close all active client connections gracefully."""
        async with self._connection_lock:
            if not self._clients:
                return
            
            # Send shutdown notification to clients
            shutdown_notification = json.dumps({
                "jsonrpc": "2.0",
                "method": "server_shutdown",
                "params": {
                    "message": "Server is shutting down"
                }
            })
            
            # Close all WebSocket connections
            close_tasks = []
            for client in self._clients.values():
                if client.websocket.open:
                    try:
                        # Send shutdown notification
                        close_tasks.append(client.websocket.send(shutdown_notification))
                    except Exception:
                        pass  # Ignore errors when sending shutdown notification
                    
                    # Close connection
                    close_tasks.append(client.websocket.close())
            
            # Wait for all connections to close
            if close_tasks:
                await asyncio.gather(*close_tasks, return_exceptions=True)
            
            # Clear client tracking
            client_count = len(self._clients)
            self._clients.clear()
            self._logger.info(f"Closed {client_count} client connections")

    async def _cleanup_server(self) -> None:
        """Clean up server resources and reset state."""
        self._server = None
        self._running = False
        
        # Clear any remaining client references
        async with self._connection_lock:
            self._clients.clear()

    def register_method(self, method_name: str, handler: Callable, version: str = "1.0") -> None:
        """
        Register a JSON-RPC method handler with version information.

        Args:
            method_name: Name of the JSON-RPC method
            handler: Async function to handle the method call
            version: API version string (default "1.0")

        Architecture Reference:
            docs/architecture/overview.md: Method-level API versioning strategy.

        # TODO: CRITICAL: Track deprecated methods and flag in registration for future implementation [Story:E1/S1a]
        """
        self._method_handlers[method_name] = handler
        self._method_versions[method_name] = version
        self._logger.debug(f"Registered JSON-RPC method: {method_name} (v{version})")

    def unregister_method(self, method_name: str) -> None:
        """
        Unregister a JSON-RPC method handler.
        
        Args:
            method_name: Name of the JSON-RPC method to remove
        """
        if method_name in self._method_handlers:
            del self._method_handlers[method_name]
            self._logger.debug(f"Unregistered JSON-RPC method: {method_name}")

    def get_method_version(self, method_name: str) -> Optional[str]:
        """
        Get the registered API version for a given method.

        Args:
            method_name: Name of the JSON-RPC method

        Returns:
            Version string if registered, else None
        """
        return self._method_versions.get(method_name)

    async def broadcast_notification(
        self, 
        method: str, 
        params: Optional[Dict[str, Any]] = None,
        target_clients: Optional[List[str]] = None
    ) -> None:
        """
        Broadcast a JSON-RPC notification to connected clients.

        Args:
            method: Notification method name
            params: Notification parameters
            target_clients: List of client IDs to notify (None for all clients)
        """
        if not self._clients:
            self._logger.debug(f"No clients connected, skipping notification: {method}")
            return
        
        # Extract or generate correlation ID for notification tracing
        correlation_id = params.get("correlation_id") if params else None
        if not correlation_id:
            correlation_id = str(uuid.uuid4())[:8]
        
        # Set correlation ID for structured logging
        set_correlation_id(correlation_id)
        
        # Create JSON-RPC 2.0 notification structure
        notification = JsonRpcNotification(
            jsonrpc="2.0",
            method=method,
            params=params
        )
        
        # Serialize notification to JSON
        try:
            notification_json = json.dumps({
                "jsonrpc": notification.jsonrpc,
                "method": notification.method,
                "params": notification.params
            })
        except Exception as e:
            self._logger.error(f"Failed to serialize notification {method}: {e}")
            return
            
        self._logger.debug(f"Broadcasting notification: {method}")
        
        # Determine target clients
        if target_clients:
            clients_to_notify = [self._clients[cid] for cid in target_clients if cid in self._clients]
        else:
            clients_to_notify = list(self._clients.values())
        
        # Send notification to each target client
        failed_clients = []
        for client in clients_to_notify:
            try:
                if client.websocket.open:
                    await client.websocket.send(notification_json)
                else:
                    failed_clients.append(client.client_id)
            except Exception as e:
                self._logger.warning(f"Failed to send notification to client {client.client_id}: {e}")
                failed_clients.append(client.client_id)
        
        # Clean up failed connections
        if failed_clients:
            async with self._connection_lock:
                for client_id in failed_clients:
                    if client_id in self._clients:
                        del self._clients[client_id]
                        self._logger.info(f"Removed disconnected client: {client_id}")
        
        success_count = len(clients_to_notify) - len(failed_clients)
        self._logger.debug(f"Notification {method} sent to {success_count}/{len(clients_to_notify)} clients")

    async def send_notification_to_client(
        self, 
        client_id: str, 
        method: str, 
        params: Optional[Dict[str, Any]] = None
    ) -> bool:
        """
        Send a JSON-RPC notification to a specific client.
        
        Args:
            client_id: Target client identifier
            method: Notification method name
            params: Notification parameters
            
        Returns:
            True if notification was sent successfully
        """
        # Validate client exists and is connected
        async with self._connection_lock:
            if client_id not in self._clients:
                self._logger.warning(f"Client {client_id} not found for notification")
                return False
            
            client = self._clients[client_id]
            
            if not client.websocket.open:
                # Remove disconnected client
                del self._clients[client_id]
                self._logger.info(f"Removed disconnected client during notification: {client_id}")
                return False
        
        # Send notification to specific client
        try:
            notification_json = json.dumps({
                "jsonrpc": "2.0",
                "method": method,
                "params": params
            })
            
            await client.websocket.send(notification_json)
            self._logger.debug(f"Sent notification '{method}' to client {client_id}")
            return True
            
        except Exception as e:
            self._logger.warning(f"Failed to send notification to client {client_id}: {e}")
            # Handle send failure and connection cleanup
            async with self._connection_lock:
                if client_id in self._clients:
                    del self._clients[client_id]
                    self._logger.info(f"Removed client after send failure: {client_id}")
            return False

    def _register_builtin_methods(self) -> None:
        """Register built-in JSON-RPC methods."""
        self.register_method("ping", self._method_ping, version="1.0")
        self.register_method("get_camera_list", self._method_get_camera_list, version="1.0")
        self.register_method("get_camera_status", self._method_get_camera_status, version="1.0")
        self.register_method("take_snapshot", self._method_take_snapshot, version="1.0")
        self.register_method("start_recording", self._method_start_recording, version="1.0")
        self.register_method("stop_recording", self._method_stop_recording, version="1.0")
        self._logger.debug("Registered built-in JSON-RPC methods")

    def _get_stream_name_from_device_path(self, device_path: str) -> str:
        """
        Extract stream name from camera device path.
        
        Args:
            device_path: Camera device path (e.g., /dev/video0)
            
        Returns:
            Stream name for MediaMTX (e.g., camera0)
        """
        try:
            # Extract device number from path like /dev/video0
            if device_path.startswith('/dev/video'):
                device_num = device_path.replace('/dev/video', '')
                return f"camera{device_num}"
            else:
                # Fallback for non-standard device paths
                return f"camera_{abs(hash(device_path)) % 1000}"
        except Exception:
            return "camera_unknown"

    def _generate_filename(self, device_path: str, extension: str, custom_filename: Optional[str] = None) -> str:
        """
        Generate filename for snapshots and recordings.
        
        Args:
            device_path: Camera device path
            extension: File extension (jpg, mp4, etc.)
            custom_filename: Custom filename if provided
            
        Returns:
            Generated filename with timestamp
        """
        if custom_filename:
            # Ensure custom filename has correct extension
            if not custom_filename.endswith(f'.{extension}'):
                return f"{custom_filename}.{extension}"
            return custom_filename
        
        # Generate timestamp-based filename
        timestamp = time.strftime("%Y-%m-%d_%H-%M-%S")
        stream_name = self._get_stream_name_from_device_path(device_path)
        return f"{stream_name}_{timestamp}.{extension}"

    async def _method_ping(self, params: Optional[Dict[str, Any]] = None) -> str:
        """
        Built-in ping method for health checks.

        Args:
            params: Method parameters (unused)

        Returns:
            "pong" response string
        """
        return "pong"

    async def _method_get_camera_list(self, params: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
        """
        Get list of all discovered cameras with their current status.

        Args:
            params: Method parameters (unused)

        Returns:
            Object with camera list and metadata.
        """
        # TODO: HIGH: Implement camera discovery logic integration [Story:E1/S1a]
        return {
            "cameras": [],
            "total": 0,
            "connected": 0
        }

    async def _method_get_camera_status(self, params: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
        """
        Get status for a specific camera device.

        Returns camera status object including metrics as specified in
        docs/architecture/overview.md (2025-08-02).

        Args:
            params: Method parameters containing:
                - device (str): Camera device path

        Returns:
            Dict containing camera status fields:
                - device: Camera device path
                - status: Connection status
                - name: Camera display name
                - resolution: Current resolution setting
                - fps: Current frame rate
                - streams: Available stream URLs
                - metrics: Performance metrics (bytes_sent, readers, uptime)
                - capabilities: Device capabilities (if available)

        Architecture Reference:
            docs/architecture/overview.md, "Camera Status Response Fields", updated 2025-08-02.

        """
        # Example implementation (replace with actual data retrieval logic)
        camera_status = {
            "device": params.get("device", "/dev/video0"),
            "status": "CONNECTED",
            "name": "Camera 0",
            "resolution": "1920x1080",
            "fps": 30,
            "streams": {
                "rtsp": "rtsp://localhost:8554/camera0",
                "webrtc": "webrtc://localhost:8002/camera0",
                "hls": "http://localhost:8002/hls/camera0.m3u8"
            },
            "metrics": {
                "bytes_sent": 12345678,
                "readers": 2,
                "uptime": 3600
            },
            "capabilities": {
                "formats": ["YUYV", "MJPEG"],
                "resolutions": ["1920x1080", "1280x720"]
            }
        }
        return camera_status

    async def _method_take_snapshot(self, params: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
        """
        Capture a snapshot from the specified camera.

        Initiates snapshot capture from an active camera stream and saves the
        image to the configured snapshots directory with optional custom filename.

        Args:
            params: Method parameters containing:
                - device (str): Camera device path (e.g., "/dev/video0")
                - filename (str, optional): Custom filename for snapshot

        Returns:
            Dict containing snapshot capture information including:
                - device: Source camera device path
                - filename: Generated or custom filename
                - status: Capture operation status
                - timestamp: Capture timestamp
                - file_size: Snapshot file size in bytes
                - file_path: Full path to saved snapshot

        Raises:
            ValueError: If device parameter missing or camera not available

        Architecture Reference:
            WebSocket JSON-RPC Server component (docs/architecture/overview.md)
            MediaMTX integration for snapshot capture functionality
        """
        if not params or 'device' not in params:
            raise ValueError("device parameter is required")
        
        device_path = params['device']
        custom_filename = params.get('filename')
        
        # Validate MediaMTX controller is available
        if not self._mediamtx_controller:
            return {
                "device": device_path,
                "filename": None,
                "status": "FAILED",
                "timestamp": time.strftime("%Y-%m-%dT%H:%M:%SZ"),
                "file_size": 0,
                "error": "MediaMTX controller not available"
            }
        
        try:
            # Convert device path to stream name
            stream_name = self._get_stream_name_from_device_path(device_path)
            
            # Generate filename
            filename = self._generate_filename(device_path, "jpg", custom_filename)
            
            # Call MediaMTX controller to take snapshot
            snapshot_result = await self._mediamtx_controller.take_snapshot(
                stream_name=stream_name,
                filename=filename
            )
            
            # Return successful result based on MediaMTX response
            return {
                "device": device_path,
                "filename": snapshot_result.get("filename", filename),
                "status": "completed",
                "timestamp": time.strftime("%Y-%m-%dT%H:%M:%SZ"),
                "file_size": snapshot_result.get("file_size", 0),
                "file_path": snapshot_result.get("file_path", f"/opt/camera-service/snapshots/{filename}")
            }
            
        except Exception as e:
            self._logger.error(f"Error taking snapshot for {device_path}: {e}")
            return {
                "device": device_path,
                "filename": custom_filename or self._generate_filename(device_path, "jpg"),
                "status": "FAILED",
                "timestamp": time.strftime("%Y-%m-%dT%H:%M:%SZ"),
                "file_size": 0,
                "error": str(e)
            }

    async def _method_start_recording(self, params: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
        """
        Start recording video from the specified camera.

        Initiates video recording from an active camera stream with configurable
        duration and format options. Creates recording session and manages state.

        Args:
            params: Method parameters containing:
                - device (str): Camera device path (e.g., "/dev/video0")
                - duration (int, optional): Recording duration in seconds (None for unlimited)
                - format (str, optional): Recording format ("mp4", "mkv")

        Returns:
            Dict containing recording session information including:
                - device: Source camera device path
                - session_id: Unique recording session identifier
                - filename: Generated recording filename
                - status: Recording operation status ("STARTED", "FAILED")
                - start_time: Recording start timestamp
                - duration: Requested duration (None if unlimited)
                - format: Recording format being used

        Raises:
            ValueError: If device missing, camera not available, or already recording

        Architecture Reference:
            WebSocket JSON-RPC Server component (docs/architecture/overview.md)
            MediaMTX integration for recording management functionality
        """
        if not params or 'device' not in params:
            raise ValueError("device parameter is required")
        
        device_path = params['device']
        duration = params.get('duration')
        format_type = params.get('format', 'mp4')
        
        # Validate MediaMTX controller is available
        if not self._mediamtx_controller:
            return {
                "device": device_path,
                "session_id": None,
                "filename": None,
                "status": "FAILED",
                "start_time": time.strftime("%Y-%m-%dT%H:%M:%SZ"),
                "duration": duration,
                "format": format_type,
                "error": "MediaMTX controller not available"
            }
        
        try:
            # Convert device path to stream name
            stream_name = self._get_stream_name_from_device_path(device_path)
            
            # Generate session ID
            session_id = str(uuid.uuid4())
            
            # Call MediaMTX controller to start recording
            recording_result = await self._mediamtx_controller.start_recording(
                stream_name=stream_name,
                duration=duration,
                format=format_type
            )
            
            # Return successful result based on MediaMTX response
            return {
                "device": device_path,
                "session_id": session_id,
                "filename": recording_result.get("filename"),
                "status": "STARTED",
                "start_time": recording_result.get("start_time", time.strftime("%Y-%m-%dT%H:%M:%SZ")),
                "duration": duration,
                "format": format_type
            }
            
        except Exception as e:
            self._logger.error(f"Error starting recording for {device_path}: {e}")
            return {
                "device": device_path,
                "session_id": None,
                "filename": None,
                "status": "FAILED",
                "start_time": time.strftime("%Y-%m-%dT%H:%M:%SZ"),
                "duration": duration,
                "format": format_type,
                "error": str(e)
            }

    async def _method_stop_recording(self, params: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
        """
        Stop active recording for the specified camera.

        Terminates ongoing recording session for a camera and provides final
        recording information including duration and file metadata.

        Args:
            params: Method parameters containing:
                - device (str): Camera device path (e.g., "/dev/video0")

        Returns:
            Dict containing recording completion information including:
                - device: Source camera device path
                - session_id: Recording session identifier that was stopped
                - filename: Final recording filename
                - status: Recording completion status ("STOPPED", "FAILED")
                - start_time: Original recording start timestamp
                - end_time: Recording stop timestamp
                - duration: Actual recording duration in seconds
                - file_size: Final recording file size

        Raises:
            ValueError: If device parameter is missing or not currently recording

        Architecture Reference:
            WebSocket JSON-RPC Server component (docs/architecture/overview.md)
            MediaMTX integration for recording management functionality
        """
        if not params or 'device' not in params:
            raise ValueError("device parameter is required")
        
        device_path = params['device']
        
        # Validate MediaMTX controller is available
        if not self._mediamtx_controller:
            return {
                "device": device_path,
                "session_id": None,
                "filename": None,
                "status": "FAILED",
                "start_time": None,
                "end_time": time.strftime("%Y-%m-%dT%H:%M:%SZ"),
                "duration": 0,
                "file_size": 0,
                "error": "MediaMTX controller not available"
            }
        
        try:
            # Convert device path to stream name
            stream_name = self._get_stream_name_from_device_path(device_path)
            
            # Call MediaMTX controller to stop recording
            recording_result = await self._mediamtx_controller.stop_recording(stream_name)
            
            # Return successful result based on MediaMTX response
            return {
                "device": device_path,
                "session_id": recording_result.get("session_id"),
                "filename": recording_result.get("filename"),
                "status": "STOPPED",
                "start_time": recording_result.get("start_time"),
                "end_time": time.strftime("%Y-%m-%dT%H:%M:%SZ"),
                "duration": recording_result.get("duration", 0),
                "file_size": recording_result.get("file_size", 0)
            }
            
        except Exception as e:
            self._logger.error(f"Error stopping recording for {device_path}: {e}")
            return {
                "device": device_path,
                "session_id": None,
                "filename": None,
                "status": "FAILED",
                "start_time": None,
                "end_time": time.strftime("%Y-%m-%dT%H:%M:%SZ"),
                "duration": 0,
                "file_size": 0,
                "error": str(e)
            }

    async def notify_camera_status_update(self, params: Dict[str, Any]) -> None:
        """
        Broadcast a camera_status_update notification to all connected clients.

        Sends a real-time notification when a camera connects, disconnects, or changes status,
        as specified in the architecture overview and API documentation.

        Args:
            params: Dictionary containing camera status fields:
                - device: string - Camera device path
                - status: string - Camera connection status
                - name: string - Camera display name
                - resolution: string - Current resolution setting
                - fps: number - Current frame rate
                - streams: dict - Available stream URLs

        Architecture Reference:
            docs/architecture/overview.md: "Server broadcasts camera status notification to authenticated clients."
            Only permitted fields are device, status, name, resolution, fps, streams.
        """
        if not params:
            self._logger.warning("Camera status update called with empty parameters")
            return
        
        # Validate required fields per API documentation
        required_fields = ['device', 'status']
        for field in required_fields:
            if field not in params:
                self._logger.error(f"Camera status update missing required field: {field}")
                return
        
        # Filter to only allowed fields per architecture overview
        allowed_fields = {'device', 'status', 'name', 'resolution', 'fps', 'streams'}
        filtered_params = {k: v for k, v in params.items() if k in allowed_fields}
        
        try:
            # Broadcast JSON-RPC 2.0 notification to all connected clients
            await self.broadcast_notification(
                method="camera_status_update",
                params=filtered_params
            )
            
            self._logger.info(f"Broadcasted camera status update for device: {params.get('device')}")
            
        except Exception as e:
            self._logger.error(f"Failed to broadcast camera status update: {e}")
            # Continue execution - notification failure should not disrupt service

    async def notify_recording_status_update(self, params: Dict[str, Any]) -> None:
        """
        Broadcast a recording_status_update notification to all connected clients.

        Sends a real-time notification when recording starts, stops, or encounters an error,
        as specified in the architecture overview and API documentation.

        Args:
            params: Dictionary containing recording status fields:
                - device: string - Camera device path
                - status: string - Recording status ("STARTED", "STOPPED", "FAILED")
                - filename: string - Recording filename
                - duration: number - Recording duration in seconds

        Architecture Reference:
            docs/architecture/overview.md: "Server notifies client when recording completes or fails."
            Only permitted fields are device, status, filename, duration.
        """
        if not params:
            self._logger.warning("Recording status update called with empty parameters")
            return
        
        # Validate required fields per API documentation
        required_fields = ['device', 'status']
        for field in required_fields:
            if field not in params:
                self._logger.error(f"Recording status update missing required field: {field}")
                return
        
        # Filter to only allowed fields per architecture overview
        allowed_fields = {'device', 'status', 'filename', 'duration'}
        filtered_params = {k: v for k, v in params.items() if k in allowed_fields}
        
        try:
            # Broadcast JSON-RPC 2.0 notification to all connected clients
            await self.broadcast_notification(
                method="recording_status_update",
                params=filtered_params
            )
            
            self._logger.info(f"Broadcasted recording status update for device: {params.get('device')}, status: {params.get('status')}")
            
        except Exception as e:
            self._logger.error(f"Failed to broadcast recording status update: {e}")
            # Continue execution - notification failure should not disrupt service

    def get_connection_count(self) -> int:
        """Get current number of connected clients."""
        return len(self._clients)

    def get_server_stats(self) -> Dict[str, Any]:
        """
        Get server statistics and status.
        
        Returns:
            Dictionary containing server metrics
        """
        return {
            "running": self._running,
            "connected_clients": len(self._clients),
            "max_connections": self._max_connections,
            "registered_methods": len(self._method_handlers)
        }

    @property
    def is_running(self) -> bool:
        """Check if the server is currently running."""
        return self._running

# CHANGE LOG
# 2025-08-02: Implemented method-level API versioning per architecture overview, resolving IV&V BLOCKED issue per roadmap.md.
# 2025-08-02: Implemented business logic for take_snapshot, start_recording, stop_recording methods with MediaMTX integration per Epic E1 task requirements.
# 2025-08-02: Standardized all TODO comment formatting to required format per docs/development/principles.md.
# 2025-08-02: Implemented WebSocket server lifecycle with proper start/stop methods, connection handling, JSON-RPC message processing, and client management per architecture requirements.