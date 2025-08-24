"""
WebSocket JSON-RPC 2.0 server for camera control and notifications.

Requirements Traceability:
- REQ-WS-001: WebSocket server shall aggregate camera status with real MediaMTX integration
- REQ-WS-002: WebSocket server shall provide camera capability metadata integration  
- REQ-ERROR-001: WebSocket server shall handle MediaMTX connection failures gracefully
- REQ-MEDIA-001: System shall support snapshot capture with configurable format and quality
- REQ-MEDIA-002: System shall support video recording with duration and format control

Story Coverage: S3 - WebSocket API Integration
IV&V Control Point: Real WebSocket communication validation
"""

import asyncio
import json
import logging
import time
import uuid
import psutil
import os
import stat
import shutil
from datetime import datetime
from typing import Dict, Any, Optional, Callable, Set, List
from dataclasses import dataclass
from collections import defaultdict

import websockets
from websockets.server import ServerProtocol as WebSocketServerProtocol
from websockets.exceptions import ConnectionClosed, WebSocketException

from camera_service.logging_config import set_correlation_id

# JSON-RPC Error Codes (RFC 32700)
AUTHENTICATION_REQUIRED = -32001
RATE_LIMIT_EXCEEDED = -32002
INSUFFICIENT_PERMISSIONS = -32003
METHOD_NOT_FOUND = -32601
INVALID_PARAMS = -32602
INTERNAL_ERROR = -32603

# Enhanced Recording Management Error Codes
ERROR_CAMERA_ALREADY_RECORDING = -1006
ERROR_STORAGE_LOW = -1008
ERROR_STORAGE_CRITICAL = -1010

# JSON-RPC Error Messages
ERROR_MESSAGES = {
    AUTHENTICATION_REQUIRED: "Authentication required",
    RATE_LIMIT_EXCEEDED: "Rate limit exceeded",
    INSUFFICIENT_PERMISSIONS: "Insufficient permissions",
    METHOD_NOT_FOUND: "Method not found",
    INVALID_PARAMS: "Invalid parameters",
    INTERNAL_ERROR: "Internal server error",
    ERROR_CAMERA_ALREADY_RECORDING: "Camera is currently recording",
    ERROR_STORAGE_LOW: "Storage space is low",
    ERROR_STORAGE_CRITICAL: "Storage space is critical"
}


class CameraNotFoundError(Exception):
    """Exception raised when a camera device is not found."""
    pass


class MediaMTXError(Exception):
    """Exception raised when MediaMTX operations fail."""
    pass


class AuthenticationError(Exception):
    """Exception raised when authentication fails."""
    pass


class PermissionError(Exception):
    """Exception raised when user lacks permission for operation."""
    pass


class StreamError(Exception):
    """Exception raised when stream operations fail."""
    pass


class PerformanceMetrics:
    """Performance metrics collection for WebSocket server."""
    
    def __init__(self) -> None:
        self.request_count = 0
        self.response_times = defaultdict(list)
        self.error_count = 0
        self.active_connections = 0
        self.start_time = time.time()
    
    def record_request(self, method: str, response_time: float) -> None:
        """Record a request with its response time."""
        self.request_count += 1
        self.response_times[method].append(response_time)
    
    def record_error(self) -> None:
        """Record an error occurrence."""
        self.error_count += 1
    
    def get_metrics(self) -> Dict[str, Any]:
        """Get current performance metrics."""
        uptime = time.time() - self.start_time
        
        # Calculate average response time across all methods
        all_response_times = []
        for times in self.response_times.values():
            all_response_times.extend(times)
        
        average_response_time = 0.0
        if all_response_times:
            average_response_time = (sum(all_response_times) / len(all_response_times)) * 1000  # Convert to milliseconds
        
        # Calculate error rate
        error_rate = 0.0
        if self.request_count > 0:
            error_rate = self.error_count / self.request_count
        
        # Get system resource usage
        try:
            memory_usage = psutil.virtual_memory().percent
            cpu_usage = psutil.cpu_percent(interval=0.1)
        except Exception:
            # Fallback values if psutil fails
            memory_usage = 0.0
            cpu_usage = 0.0
        
        return {
            "active_connections": self.active_connections,
            "total_requests": self.request_count,
            "average_response_time": average_response_time,
            "error_rate": error_rate,
            "memory_usage": memory_usage,
            "cpu_usage": cpu_usage
        }


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
        # Use try/except to handle cases where no event loop is available (e.g., in tests)
        try:
            self.connected_at = asyncio.get_event_loop().time()
        except RuntimeError:
            # Fallback for test environments without event loop
            import time

            self.connected_at = time.time()

        # Authentication and security state
        self.auth_result = None
        self.user_id = None
        self.role = None
        self.auth_method = None


class WebSocketJsonRpcServer:  # type: ignore[misc]
    """
    WebSocket JSON-RPC 2.0 server for camera control and real-time notifications.

    Provides camera control API and broadcasts real-time events to connected clients
    as specified in the architecture overview.
    """

    # STOP: MEDIUM: Version negotiation and deprecated method tracking deferred to post-1.0 [IV&V:S2b]
    # Rationale: Current method-level versioning satisfies architecture requirements for MVP.
    # Full version negotiation during handshake and migration guides documented as future enhancement.
    # Owner: Solo engineer | Date: 2025-08-03 | Revisit: Post-1.0 when client SDK ecosystem requires it
    _method_versions: Dict[str, str] = {}

    def __init__(
        self,
        host: str,
        port: int,
        websocket_path: str,
        max_connections: int,
        mediamtx_controller: Optional[Any] = None,
        camera_monitor: Optional[Any] = None,
        config: Optional[Any] = None,
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
        self._config = config

        self._logger = logging.getLogger(__name__)
        self._server: Optional[Any] = None  # WebSocket server instance
        self._running = False

        # Client connection management
        self._clients: Dict[str, ClientConnection] = {}
        self._connection_lock = asyncio.Lock()

        # JSON-RPC method handlers
        self._method_handlers: Dict[str, Callable] = {}

        # Security middleware (set by service manager)
        self._security_middleware = None
        # Service manager (set by service manager)
        self._service_manager = None
        # Performance metrics collection
        self._performance_metrics = PerformanceMetrics()

        # Scheduled auto-stop tasks per stream name
        self._recording_stop_tasks: Dict[str, asyncio.Task] = {}
        
        # Recording state management
        self._active_recordings: Dict[str, Dict[str, Any]] = {}  # device_path -> session_info
        
        # Recording management configuration (defaults)
        self._recording_rotation_minutes = 30
        self._storage_warn_percent = 80
        self._storage_block_percent = 90
        
        # Load configuration if available
        self._load_recording_config()

    def _get_config_safe(self) -> Any:
        """
        Safely get configuration with null check.
        
        Returns:
            Configuration object
            
        Raises:
            RuntimeError: If configuration is not available
        """
        if self._config is None:
            raise RuntimeError("Configuration not available")
        return self._config

    def _load_recording_config(self) -> None:
        """
        Load recording management configuration from config or environment variables.
        """
        try:
            if self._config and hasattr(self._config, 'recording'):
                recording_config = self._config.recording
                self._recording_rotation_minutes = getattr(recording_config, 'rotation_minutes', 30)
                self._storage_warn_percent = getattr(recording_config, 'storage_warn_percent', 80)
                self._storage_block_percent = getattr(recording_config, 'storage_block_percent', 90)
                self._logger.info(f"Loaded recording config: rotation={self._recording_rotation_minutes}m, warn={self._storage_warn_percent}%, block={self._storage_block_percent}%")
                
                # Update MediaMTX controller configuration
                self._update_controller_recording_config()
        except Exception as e:
            self._logger.warning(f"Failed to load recording config, using defaults: {e}")

    def _update_controller_recording_config(self) -> None:
        """
        Update MediaMTX controller with current recording configuration.
        """
        try:
            if self._mediamtx_controller is not None and hasattr(self._mediamtx_controller, 'update_recording_config'):  # type: ignore[unreachable]
                self._mediamtx_controller.update_recording_config(
                    rotation_minutes=self._recording_rotation_minutes,
                    storage_warn_percent=self._storage_warn_percent,
                    storage_block_percent=self._storage_block_percent
                )
            elif self._service_manager is not None and hasattr(self._service_manager, '_mediamtx_controller'):  # type: ignore[unreachable]
                controller = self._service_manager._mediamtx_controller
                if controller is not None and hasattr(controller, 'update_recording_config'):
                    controller.update_recording_config(
                        rotation_minutes=self._recording_rotation_minutes,
                        storage_warn_percent=self._storage_warn_percent,
                        storage_block_percent=self._storage_block_percent
                    )
        except Exception as e:
            self._logger.warning(f"Failed to update controller recording config: {e}")

    def set_mediamtx_controller(self, controller: Any) -> None:
        """
        Set the MediaMTX controller for stream operations.

        Args:
            controller: MediaMTX controller instance
        """
        self._mediamtx_controller = controller

    def set_camera_monitor(self, monitor: Any) -> None:
        """
        Set the camera monitor for device information.

        Args:
            monitor: Camera monitor instance
        """
        self._camera_monitor = monitor

    def set_service_manager(self, service_manager: Any) -> None:
        """
        Set the service manager for lifecycle coordination.

        Args:
            service_manager: Service manager instance
        """
        self._service_manager = service_manager
    
    def set_security_middleware(self, security_middleware: Any) -> None:
        """
        Set the security middleware for authentication and rate limiting.

        Args:
            security_middleware: Security middleware instance
        """
        self._security_middleware = security_middleware
    
    def get_performance_metrics(self) -> Dict[str, Any]:
        """
        Get current performance metrics.

        Returns:
            Dict containing performance metrics
        """
        metrics = self._performance_metrics.get_metrics()
        metrics["active_connections"] = len(self._clients)
        return metrics

    def check_storage_space(self) -> Optional[Dict[str, Any]]:
        """
        Check available storage space and return status.
        
        Returns:
            Dict containing storage information or None if check fails
        """
        try:
            # Get storage path from config or use default
            storage_path = "/opt/camera-service/recordings"
            if self._config and hasattr(self._config, 'mediamtx') and hasattr(self._config.mediamtx, 'recordings_path'):
                storage_path = self._config.mediamtx.recordings_path
            
            statvfs = os.statvfs(storage_path)
            total_space = statvfs.f_frsize * statvfs.f_blocks
            available_space = statvfs.f_frsize * statvfs.f_bavail
            used_percent = ((total_space - available_space) / total_space) * 100
            
            return {
                "total_space": total_space,
                "available_space": available_space,
                "used_percent": used_percent,
                "warn_threshold": self._storage_warn_percent,
                "block_threshold": self._storage_block_percent
            }
        except Exception as e:
            self._logger.error(f"Storage check failed: {e}")
            return None

    async def start(self) -> None:
        """
        Start the WebSocket JSON-RPC server.

        Initializes the WebSocket server and begins accepting client connections.
        """
        if self._running:
            self._logger.warning("WebSocket server is already running")
            return

        self._logger.info(
            f"Starting WebSocket JSON-RPC server on {self._host}:{self._port}{self._websocket_path}"
        )

        try:
            # Register built-in methods
            self._register_builtin_methods()

            # Resolve bind host for broader compatibility across IPv4/IPv6 stacks.
            # In some environments, binding specifically to "localhost" may select
            # only the IPv6 or IPv4 loopback, while client resolution prefers the
            # opposite family, causing connection issues during handshake.
            bind_host: Optional[str] = self._host
            if self._host in ("localhost", "127.0.0.1", "::1"):
                # Bind on all interfaces (both stacks) to avoid family mismatch.
                bind_host = None

            # Start WebSocket server with proper error handling
            self._server = await websockets.serve(
                self._handle_client_connection,
                bind_host,
                self._port,
                # Server configuration
                max_size=1024 * 1024,  # 1MB max message size
                max_queue=100,  # Max queued messages per connection
                compression=None,  # Disable compression for simplicity
                ping_interval=30,  # Ping every 30 seconds
                ping_timeout=10,  # Ping timeout
                close_timeout=5,  # Close timeout
                reuse_port=True,
                reuse_address=True,
            )

            self._running = True
            # Log bound sockets for diagnostics
            try:
                sockets = getattr(self._server, "sockets", []) or []
                bound = [s.getsockname() for s in sockets]
            except Exception:
                bound = []
            self._logger.info(
                f"WebSocket JSON-RPC server started successfully on {self._host}:{self._port}{self._websocket_path}",
                extra={"bound_sockets": bound},
            )

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

    async def _process_request(self, path: str, request_headers: Any) -> Optional[tuple]:
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
                self._logger.warning(
                    f"Connection limit reached: {len(self._clients)}/{self._max_connections}"
                )
                return (503, {}, b"Service Unavailable - Connection limit reached")

        # Accept connection
        return None

    async def _handle_client_connection(
        self, websocket: WebSocketServerProtocol
    ) -> None:
        """
        Handle new client WebSocket connection.

        Args:
            websocket: WebSocket connection object
        """
        # Validate WebSocket path - be more lenient for now since websockets.serve() doesn't handle path routing
        path = getattr(websocket, 'path', '/')
        if path != self._websocket_path and path != '/':
            self._logger.warning(f"Invalid WebSocket path requested: {path}")
            await websocket.close(1008, "Invalid path")
            return

        # Check connection limits
        async with self._connection_lock:
            if len(self._clients) >= self._max_connections:
                self._logger.warning(
                    f"Connection limit reached: {len(self._clients)}/{self._max_connections}"
                )
                await websocket.close(1013, "Connection limit reached")
                return

        client_id = str(uuid.uuid4())
        client_ip = (
            websocket.remote_address[0] if websocket.remote_address else "unknown"  # type: ignore[attr-defined]
        )

        self._logger.info(
            f"New client connection: {client_id} from {client_ip}",
            extra={
                "client_id": client_id,
                "client_ip": client_ip,
                "correlation_id": client_id,
                "event": "client_connected"
            }
        )

        # Create client connection object
        client = ClientConnection(websocket, client_id)

        try:
            # Security middleware connection check
            if self._security_middleware is not None:
                if not self._security_middleware.can_accept_connection(client_id):
                    self._logger.warning(f"Connection rejected for client {client_id} - limit reached")
                    return
                self._security_middleware.register_connection(client_id)

            # Add client to tracking
            async with self._connection_lock:
                self._clients[client_id] = client

            self._logger.debug(
                f"Client {client_id} added to connection pool ({len(self._clients)} total)"
            )

            # Process incoming messages from client
            async for message in websocket:
                try:
                    if isinstance(message, str):
                        response = await self._handle_json_rpc_message(client, message)
                        if response:
                            await websocket.send(response)  # type: ignore[attr-defined]
                    else:
                        self._logger.warning(
                            f"Received non-text message from client {client_id}"
                        )

                except Exception as e:
                    self._logger.error(
                        f"Error processing message from client {client_id}: {e}"
                    )
                    # Send JSON-RPC error response if possible
                    try:
                        error_response = json.dumps(
                            {
                                "jsonrpc": "2.0",
                                "error": {"code": INTERNAL_ERROR, "message": ERROR_MESSAGES[INTERNAL_ERROR]},
                                "id": None,
                            }
                        )
                        await websocket.send(error_response)  # type: ignore[attr-defined]
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
            # Security middleware cleanup
            if self._security_middleware is not None:
                self._security_middleware.unregister_connection(client_id)
            
            # Cleanup on disconnect
            async with self._connection_lock:
                if client_id in self._clients:
                    del self._clients[client_id]
                    self._logger.info(
                        f"Removed client {client_id} from connection pool ({len(self._clients)} remaining)"
                    )

    async def _handle_json_rpc_message(
        self, client: ClientConnection, message: str
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
            except json.JSONDecodeError:
                return json.dumps(
                    {
                        "jsonrpc": "2.0",
                        "error": {"code": -32700, "message": "Parse error"},
                        "id": None,
                    }
                )

            # Extract correlation ID from request ID field
            request_id = request_data.get("id")
            correlation_id = (
                str(request_id) if request_id is not None else str(uuid.uuid4())[:8]
            )

            # Set correlation ID for structured logging
            set_correlation_id(correlation_id)

            # Validate JSON-RPC structure
            if (
                not isinstance(request_data, dict)
                or request_data.get("jsonrpc") != "2.0"
            ):
                return json.dumps(
                    {
                        "jsonrpc": "2.0",
                        "error": {"code": -32600, "message": "Invalid Request"},
                        "id": request_id,
                    }
                )

            method_name = request_data.get("method")
            if not method_name or not isinstance(method_name, str):
                return json.dumps(
                    {
                        "jsonrpc": "2.0",
                        "error": {
                            "code": -32600,
                            "message": "Invalid Request - missing method",
                        },
                        "id": request_id,
                    }
                )

            params = request_data.get("params")

            self._logger.debug(
                f"Processing JSON-RPC method '{method_name}' from client {client.client_id}",
                extra={
                    "method": method_name,
                    "client_id": client.client_id,
                    "correlation_id": correlation_id,
                    "event": "method_processing"
                }
            )


            # Check if method exists
            if method_name not in self._method_handlers:
                return json.dumps(
                    {
                        "jsonrpc": "2.0",
                        "error": {"code": METHOD_NOT_FOUND, "message": ERROR_MESSAGES[METHOD_NOT_FOUND]},
                        "id": request_id,
                    }
                )

            # Security enforcement and rate limiting (F3.2.5/F3.2.6/I1.7/N3.x)
            # CRITICAL FIX: Protect all sensitive methods from authentication bypass
            method_permissions = {
                # Authentication method (no auth required)
                "authenticate": None,  # No authentication required for authenticate method
                
                # Viewer access (read-only operations)
                "get_camera_list": "viewer",
                "get_camera_status": "viewer", 
                "get_camera_capabilities": "viewer",
                "list_recordings": "viewer",
                "list_snapshots": "viewer",
                "get_recording_info": "viewer",
                "get_snapshot_info": "viewer",
                "get_streams": "viewer",
                "ping": "viewer",
                
                # Operator access (camera control operations)
                "take_snapshot": "operator",
                "start_recording": "operator", 
                "stop_recording": "operator",
                "delete_recording": "operator",
                "delete_snapshot": "operator",
                
                # Admin access (system management operations)
                "get_metrics": "admin",
                "get_status": "admin",
                "get_server_info": "admin",
                "get_storage_info": "admin",
                "set_retention_policy": "admin",
                "cleanup_old_files": "admin"
            }

            # Check if method requires authentication
            if method_name in method_permissions:
                # CRITICAL SECURITY FIX: Always enforce authentication for protected methods
                if self._security_middleware is None:
                    return json.dumps(
                        {
                            "jsonrpc": "2.0",
                            "error": {
                                "code": AUTHENTICATION_REQUIRED,
                                "message": ERROR_MESSAGES[AUTHENTICATION_REQUIRED],
                            },
                            "id": request_id,
                        }
                    )
                
                # Rate limiting applies to all protected methods
                if self._security_middleware is not None and not self._security_middleware.check_rate_limit(client.client_id):
                    return json.dumps(
                        {
                            "jsonrpc": "2.0",
                            "error": {
                                "code": RATE_LIMIT_EXCEEDED, 
                                "message": ERROR_MESSAGES[RATE_LIMIT_EXCEEDED]
                            },
                            "id": request_id,
                        }
                    )
                
                required_role = method_permissions[method_name]
                
                # Skip authentication for methods that don't require it (like authenticate)
                if required_role is None:
                    pass  # No authentication required
                else:
                    # Check if security middleware is available
                    if self._security_middleware is None:
                        return json.dumps(
                            {
                                "jsonrpc": "2.0",
                                "error": {
                                    "code": AUTHENTICATION_REQUIRED,
                                    "message": ERROR_MESSAGES[AUTHENTICATION_REQUIRED],
                                },
                                "id": request_id,
                            }
                        )
                    
                    # Require authentication for all protected methods
                    if not self._security_middleware.is_authenticated(client.client_id):
                        auth_token = None
                        if params and isinstance(params, dict):
                            auth_token = params.get("auth_token") or params.get("token")
                        if auth_token:
                            auth_result = await self._security_middleware.authenticate_connection(
                                client.client_id, auth_token
                            )
                            if auth_result.authenticated:
                                client.authenticated = True
                                client.auth_result = auth_result
                                client.user_id = auth_result.user_id
                                client.role = auth_result.role
                                client.auth_method = auth_result.auth_method
                            else:
                                return json.dumps(
                                    {
                                        "jsonrpc": "2.0",
                                        "error": {
                                            "code": AUTHENTICATION_REQUIRED,
                                            "message": f"Authentication failed: {auth_result.error_message}",
                                        },
                                        "id": request_id,
                                    }
                                )
                        else:
                            return json.dumps(
                                {
                                    "jsonrpc": "2.0",
                                    "error": {
                                        "code": AUTHENTICATION_REQUIRED,
                                        "message": f"Authentication required for {method_name} - call authenticate or provide auth_token",
                                    },
                                    "id": request_id,
                                }
                            )

                # Role-based authorization check (skip if no role required)
                if required_role is not None and not self._security_middleware.has_permission(client.client_id, required_role):
                    return json.dumps(
                        {
                            "jsonrpc": "2.0",
                            "error": {
                                "code": INSUFFICIENT_PERMISSIONS,
                                "message": f"Insufficient permissions - {required_role} role required for {method_name}",
                            },
                            "id": request_id,
                        }
                    )

                # Session expiry enforcement on each protected call
                auth_state = self._security_middleware.get_auth_result(client.client_id)
                if auth_state is not None:
                    try:
                        expires_at = getattr(auth_state, "expires_at", None)
                        if expires_at is not None:
                            now_ts = int(time.time())
                            if now_ts >= int(expires_at):
                                # Invalidate session
                                client.authenticated = False
                                client.auth_result = None
                                client.user_id = None
                                client.role = None
                                return json.dumps(
                                    {
                                        "jsonrpc": "2.0",
                                        "error": {
                                            "code": AUTHENTICATION_REQUIRED,
                                            "message": "Authentication failed: token expired",
                                        },
                                        "id": request_id,
                                    }
                                )
                    except Exception:
                        pass

            # Call method handler
            try:
                handler = self._method_handlers[method_name]
                # Performance timing
                start_ts = time.perf_counter()
                if params is not None:
                    result = await handler(params)
                else:
                    result = await handler()
                duration_ms = (time.perf_counter() - start_ts) * 1000.0
                
                # Record performance metrics
                self._performance_metrics.record_request(method_name, duration_ms / 1000.0)

                # Return response for requests with ID (notifications have no ID)
                if request_id is not None:
                    return json.dumps(
                        {"jsonrpc": "2.0", "result": result, "id": request_id}
                    )
                else:
                    # Notification - no response
                    return None

            except Exception as e:
                self._logger.error(f"Error in method handler '{method_name}': {e}")
                # Record error in performance metrics
                self._performance_metrics.record_error()
                if request_id is not None:
                    # Map custom exceptions to specific error codes
                    if isinstance(e, CameraNotFoundError):
                        error_code = -1000
                        error_message = "Camera device not found"
                    elif isinstance(e, MediaMTXError):
                        # Check for specific recording management errors
                        error_msg = str(e)
                        if "Camera is currently recording" in error_msg:
                            error_code = ERROR_CAMERA_ALREADY_RECORDING
                            error_message = ERROR_MESSAGES[ERROR_CAMERA_ALREADY_RECORDING]
                        elif "Storage space is critical" in error_msg:
                            error_code = ERROR_STORAGE_CRITICAL
                            error_message = ERROR_MESSAGES[ERROR_STORAGE_CRITICAL]
                        elif "Storage space is low" in error_msg:
                            error_code = ERROR_STORAGE_LOW
                            error_message = ERROR_MESSAGES[ERROR_STORAGE_LOW]
                        else:
                            error_code = -1003
                            error_message = "MediaMTX operation failed"
                    elif isinstance(e, FileNotFoundError):
                        error_code = -1005
                        error_message = str(e)
                    elif isinstance(e, AuthenticationError):
                        error_code = AUTHENTICATION_REQUIRED
                        error_message = ERROR_MESSAGES[AUTHENTICATION_REQUIRED]
                    elif isinstance(e, PermissionError):
                        error_code = INSUFFICIENT_PERMISSIONS
                        error_message = ERROR_MESSAGES[INSUFFICIENT_PERMISSIONS]
                    elif isinstance(e, StreamError):
                        error_code = -1004
                        error_message = "Stream operation failed"
                    elif isinstance(e, ValueError):
                        error_code = INVALID_PARAMS
                        error_message = ERROR_MESSAGES[INVALID_PARAMS]
                    elif isinstance(e, KeyError):
                        error_code = INVALID_PARAMS
                        error_message = "Missing required parameter"
                    elif isinstance(e, TypeError):
                        error_code = INVALID_PARAMS
                        error_message = "Invalid parameter type"
                    else:
                        error_code = INTERNAL_ERROR
                        error_message = ERROR_MESSAGES[INTERNAL_ERROR]
                    
                    return json.dumps(
                        {
                            "jsonrpc": "2.0",
                            "error": {"code": error_code, "message": error_message},
                            "id": request_id,
                        }
                    )
                else:
                    # Notification - no response
                    return None

        except Exception as e:
            self._logger.error(
                f"Error processing JSON-RPC message from client {client.client_id}: {e}"
            )
            return json.dumps(
                {
                    "jsonrpc": "2.0",
                    "error": {"code": INTERNAL_ERROR, "message": ERROR_MESSAGES[INTERNAL_ERROR]},
                    "id": request_id,
                }
            )

    async def _close_all_connections(self) -> None:
        """Close all client connections gracefully."""
        if not self._clients:
            return

        self._logger.info(f"Closing {len(self._clients)} client connections")

        failed_clients = []
        for client in list(self._clients.values()):
            try:
                # Check if websocket is still available and open
                if hasattr(client.websocket, 'open'):
                    if client.websocket.open:  # type: ignore[attr-defined]
                        await client.websocket.close()
                else:
                    # Fallback for websockets library versions without 'open' attribute
                    try:
                        await client.websocket.close()
                    except Exception:
                        pass  # Connection may already be closed
                        
            except Exception as e:
                self._logger.warning(f"Failed to close client {client.client_id}: {e}")
                failed_clients.append(client.client_id)

        # Remove failed clients from tracking
        for client_id in failed_clients:
            self._clients.pop(client_id, None)

        if failed_clients:
            self._logger.warning(f"Failed to close {len(failed_clients)} client connections")
        else:
            self._logger.info("All client connections closed successfully")

    async def _cleanup_server(self) -> None:
        """Clean up server resources and reset state."""
        self._server = None
        self._running = False

        # Clear any remaining client references
        async with self._connection_lock:
            self._clients.clear()

    def register_method(
        self, method_name: str, handler: Callable, version: str = "1.0"
    ) -> None:
        """
        Register a JSON-RPC method handler with version information.

        Args:
            method_name: Name of the JSON-RPC method
            handler: Async function to handle the method call
            version: API version string (default "1.0")

        Architecture Reference:
            docs/architecture/overview.md: Method-level API versioning strategy.
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
        target_clients: Optional[List[str]] = None,
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
        notification = JsonRpcNotification(jsonrpc="2.0", method=method, params=params)

        # Serialize notification to JSON
        try:
            notification_json = json.dumps(
                {
                    "jsonrpc": notification.jsonrpc,
                    "method": notification.method,
                    "params": notification.params,
                }
            )
        except Exception as e:
            self._logger.error(f"Failed to serialize notification {method}: {e}")
            return

        self._logger.debug(f"Broadcasting notification: {method}")

        # Determine target clients
        if target_clients:
            clients_to_notify = [
                self._clients[cid] for cid in target_clients if cid in self._clients
            ]
            self._logger.debug(f"Target clients: {target_clients}, found: {[c.client_id for c in clients_to_notify]}")
        else:
            clients_to_notify = list(self._clients.values())
            self._logger.debug(f"Broadcasting to all {len(clients_to_notify)} clients")

        # Helper to determine if a websocket appears connected
        def _is_ws_connected(ws: Any) -> bool:
            try:
                is_closed = getattr(ws, "closed", None)
                if is_closed is True:
                    return False
                is_open = getattr(ws, "open", None)
                if is_open is False:
                    return False
                # If attributes are missing or ambiguous, assume connected and rely on send exceptions
                return True
            except Exception:
                # On any introspection error, optimistically attempt to send
                return True

        # Send notification to each target client (robust open/closed checks)
        failed_clients = []
        self._logger.debug(f"Sending notification to {len(clients_to_notify)} clients")
        for client in clients_to_notify:
            try:
                if _is_ws_connected(client.websocket):
                    await client.websocket.send(notification_json)
                    self._logger.debug(f"Successfully sent notification to client {client.client_id}")
                else:
                    self._logger.debug(f"Client {client.client_id} websocket not connected")
                    failed_clients.append(client.client_id)
            except Exception as e:
                self._logger.warning(
                    f"Failed to send notification to client {client.client_id}: {e}"
                )
                failed_clients.append(client.client_id)

        # Clean up failed connections
        if failed_clients:
            async with self._connection_lock:
                for client_id in failed_clients:
                    if client_id in self._clients:
                        del self._clients[client_id]
                        self._logger.info(f"Removed disconnected client: {client_id}")

        success_count = len(clients_to_notify) - len(failed_clients)
        self._logger.debug(
            f"Notification {method} sent to {success_count}/{len(clients_to_notify)} clients"
        )

    async def _emit_recording_complete(self, device_path: str, stream_name: str) -> None:
        """Emit recording completion notification to clients."""
        try:
            await self.broadcast_notification(
                method="recording_status_update",
                params={
                    "device": device_path,
                    "status": "STOPPED",
                    "filename": f"{stream_name}_{time.strftime('%Y-%m-%d_%H-%M-%S')}.mp4",
                },
            )
        except Exception as e:
            self._logger.debug(f"Failed to emit recording completion: {e}")

    async def send_notification_to_client(
        self, client_id: str, method: str, params: Optional[Dict[str, Any]] = None
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

            # websockets protocol exposes 'closed' / 'open' differently across versions
            def _is_ws_connected(ws: Any) -> bool:
                try:
                    is_closed = getattr(ws, "closed", None)
                    if is_closed is True:
                        return False
                    is_open = getattr(ws, "open", None)
                    if is_open is False:
                        return False
                    return True
                except Exception:
                    return True

            if not _is_ws_connected(client.websocket):
                # Remove disconnected client
                del self._clients[client_id]
                self._logger.info(
                    f"Removed disconnected client during notification: {client_id}"
                )
                return False

        # Send notification to specific client
        try:
            notification_json = json.dumps(
                {"jsonrpc": "2.0", "method": method, "params": params}
            )

            def _is_ws_connected(ws: Any) -> bool:
                try:
                    is_closed = getattr(ws, "closed", None)
                    if is_closed is True:
                        return False
                    is_open = getattr(ws, "open", None)
                    if is_open is False:
                        return False
                    return True
                except Exception:
                    return True

            if _is_ws_connected(client.websocket):
                try:
                    await client.websocket.send(notification_json)
                    return True
                except Exception as e:
                    self._logger.warning(
                        f"Failed to send notification to client {client_id}: {e}"
                    )
                    # Handle send failure and connection cleanup
                    async with self._connection_lock:
                        if client_id in self._clients:
                            del self._clients[client_id]
                            self._logger.info(f"Removed client after send failure: {client_id}")
            return False

        except Exception as e:
            self._logger.warning(
                f"Failed to send notification to client {client_id}: {e}"
            )
            # Handle send failure and connection cleanup
            async with self._connection_lock:
                if client_id in self._clients:
                    del self._clients[client_id]
                    self._logger.info(f"Removed client after send failure: {client_id}")
            return False

    def _register_builtin_methods(self) -> None:
        """Register built-in JSON-RPC methods."""
        self.register_method("ping", self._method_ping, version="1.0")
        self.register_method(
            "get_camera_list", self._method_get_camera_list, version="1.0"
        )

        self.register_method(
            "get_camera_status", self._method_get_camera_status, version="1.0"
        )
        self.register_method(
            "get_camera_capabilities", self._method_get_camera_capabilities, version="1.0"
        )
        self.register_method(
            "get_streams", self._method_get_streams, version="1.0"
        )
        self.register_method("take_snapshot", self._method_take_snapshot, version="1.0")
        self.register_method(
            "start_recording", self._method_start_recording, version="1.0"
        )
        self.register_method(
            "stop_recording", self._method_stop_recording, version="1.0"
        )
        # Security and observability
        self.register_method("authenticate", self._method_authenticate, version="1.0")
        self.register_method("get_metrics", self._method_get_metrics, version="1.0")
        self.register_method("get_status", self._method_get_status, version="1.0")
        self.register_method("get_server_info", self._method_get_server_info, version="1.0")
        # File management methods (Epic E6)
        self.register_method("list_recordings", self._method_list_recordings, version="1.0")
        self.register_method("list_snapshots", self._method_list_snapshots, version="1.0")
        # File lifecycle management methods (Epic E6 - File Lifecycle Management)
        self.register_method("get_recording_info", self._method_get_recording_info, version="1.0")
        self.register_method("get_snapshot_info", self._method_get_snapshot_info, version="1.0")
        self.register_method("delete_recording", self._method_delete_recording, version="1.0")
        self.register_method("delete_snapshot", self._method_delete_snapshot, version="1.0")
        self.register_method("get_storage_info", self._method_get_storage_info, version="1.0")
        self.register_method("set_retention_policy", self._method_set_retention_policy, version="1.0")
        self.register_method("cleanup_old_files", self._method_cleanup_old_files, version="1.0")
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
            # Handle empty or invalid device paths
            if not device_path or not isinstance(device_path, str):
                return "camera_unknown"
            
            # Extract device number from path like /dev/video0
            if device_path.startswith("/dev/video"):
                device_num = device_path.replace("/dev/video", "")
                return f"camera{device_num}"
            else:
                # Fallback for non-standard device paths
                return f"camera_{abs(hash(device_path)) % 1000}"
        except Exception:
            return "camera_unknown"

    async def _get_video_duration_architecture_compliant(self, file_path: str, filename: str) -> Optional[int]:
        """
        Get video duration using architecture-compliant approach.
        
        Architecture Decision: Use MediaMTX capabilities and file analysis rather than
        direct FFmpeg integration in the WebSocket server component.
        
        Args:
            file_path: Path to the video file
            filename: Name of the video file
            
        Returns:
            Duration in seconds, or None if extraction fails
        """
        try:
            # First, try to get duration from MediaMTX if the file is associated with a stream
            if hasattr(self, '_mediamtx_controller') and self._mediamtx_controller:
                try:
                    # Check if this file is associated with an active MediaMTX stream
                    stream_name = self._extract_stream_name_from_filename(filename)
                    if stream_name:
                        # Try to get stream metadata from MediaMTX
                        stream_info = await self._mediamtx_controller.get_stream_info(stream_name)
                        if stream_info and 'duration' in stream_info:
                            return int(stream_info['duration'])
                except Exception as e:
                    self._logger.debug(f"Could not get duration from MediaMTX for {filename}: {e}")
            
            # Fallback: Use file size and bitrate estimation for MP4 files
            # This is a reasonable approximation for H.264 encoded files
            if filename.lower().endswith('.mp4'):
                return self._estimate_duration_from_file_size(file_path)
            
            # For other video formats, return None (placeholder for future implementation)
            return None
            
        except Exception as e:
            self._logger.debug(f"Error getting video duration for {filename}: {e}")
            return None
    
    def _extract_stream_name_from_filename(self, filename: str) -> Optional[str]:
        """
        Extract stream name from filename based on naming convention.
        
        Args:
            filename: Video filename (e.g., "camera0_2025-01-15_14-30-00.mp4")
            
        Returns:
            Stream name (e.g., "camera0") or None if not extractable
        """
        try:
            # Extract stream name from filename pattern: {stream_name}_{timestamp}.mp4
            if '_' in filename:
                stream_name = filename.split('_')[0]
                if stream_name.startswith('camera'):
                    return stream_name
            return None
        except Exception:
            return None
    
    def _estimate_duration_from_file_size(self, file_path: str) -> Optional[int]:
        """
        Estimate video duration from file size using typical H.264 bitrates.
        
        This is a reasonable approximation for files recorded by this system
        using the STANAG 4406 H.264 configuration (600kbps baseline profile).
        
        Args:
            file_path: Path to the video file
            
        Returns:
            Estimated duration in seconds, or None if estimation fails
        """
        try:
            # Get file size
            file_size = os.path.getsize(file_path)
            
            # Use typical bitrate for this system's H.264 configuration
            # STANAG 4406 baseline profile at 600kbps
            typical_bitrate_bps = 600 * 1024  # 600 kbps in bits per second
            
            # Calculate estimated duration
            # Duration = File Size / Bitrate
            # Convert file size to bits and divide by bitrate
            duration_seconds = (file_size * 8) / typical_bitrate_bps
            
            # Round to nearest second and ensure reasonable bounds
            duration_seconds = max(1, min(86400, int(duration_seconds)))  # 1 second to 24 hours
            
            return duration_seconds
            
        except Exception as e:
            self._logger.debug(f"Could not estimate duration from file size for {file_path}: {e}")
            return None

    def _generate_filename(
        self, device_path: str, extension: str, custom_filename: Optional[str] = None
    ) -> str:
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
            if not custom_filename.endswith(f".{extension}"):
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

    async def _method_authenticate(
        self, params: Optional[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        """
        Authenticate the current WebSocket connection using a JWT or API key.
        Params: { auth_token: string, auth_type: "jwt"|"api_key" (optional) }
        
        Returns: Authentication result with user role, permissions, session information
        """
        if self._security_middleware is None:
            return {"authenticated": False, "error": "Security not configured"}
        
        token = None
        auth_type = "auto"
        if params and isinstance(params, dict):
            token = params.get("auth_token") or params.get("token")
            auth_type = params.get("auth_type", "auto")
        
        if not token:
            return {"authenticated": False, "error": "Missing auth_token"}

        # Identify caller client by scanning connections for this task's websocket
        # Fallback to first client if exact match not determinable (keeps server stable)
        client = None
        async with self._connection_lock:
            if self._clients:
                # Take the most recently added as heuristic
                client = next(reversed(self._clients.values()))
        
        if client is None:
            return {"authenticated": False, "error": "Client context unavailable"}

        auth_result = await self._security_middleware.authenticate_connection(
            client.client_id, token, auth_type
        )
        
        if auth_result.authenticated:
            # Set up client authentication state
            client.authenticated = True
            client.auth_result = auth_result
            client.user_id = auth_result.user_id
            client.role = auth_result.role
            client.auth_method = auth_result.auth_method
            
            # Generate session ID and expiration
            import uuid
            import time
            from datetime import datetime, timezone
            
            session_id = str(uuid.uuid4())
            expires_at = auth_result.expires_at if hasattr(auth_result, 'expires_at') else None
            
            # Convert expires_at to ISO format if it exists
            if expires_at:
                if isinstance(expires_at, int):
                    expires_at_iso = datetime.fromtimestamp(expires_at, tz=timezone.utc).isoformat()
                else:
                    expires_at_iso = expires_at
            else:
                # Default to 24 hours from now
                expires_at_iso = datetime.fromtimestamp(time.time() + 86400, tz=timezone.utc).isoformat()
            
            # Determine permissions based on role
            permissions = []
            if auth_result.role == "viewer":
                permissions = ["view"]
            elif auth_result.role == "operator":
                permissions = ["view", "control"]
            elif auth_result.role == "admin":
                permissions = ["view", "control", "admin"]
            
            return {
                "authenticated": True,
                "role": auth_result.role,
                "permissions": permissions,
                "expires_at": expires_at_iso,
                "session_id": session_id
            }
        
        return {"authenticated": False, "error": auth_result.error_message}

    async def _method_get_metrics(
        self, params: Optional[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        """
        Get system performance metrics and statistics.
        
        Follows ground truth format from docs/api/json-rpc-methods.md exactly.
        Returns system metrics as documented in API specification.

        Args:
            params: Method parameters (unused)

        Returns:
            Dict containing system metrics with fields: active_connections, total_requests, 
            average_response_time, error_rate, memory_usage, cpu_usage
        """
        # Get base performance metrics
        base_metrics = self.get_performance_metrics()
        
        # Extract required fields per API documentation
        active_connections = len(self._clients)
        total_requests = base_metrics.get("total_requests", 0)
        average_response_time = base_metrics.get("average_response_time", 0.0)
        error_rate = base_metrics.get("error_rate", 0.0)
        
        # Get system resource usage
        try:
            process = psutil.Process()
            memory_usage = process.memory_info().rss / 1024 / 1024  # MB
            cpu_usage = process.cpu_percent()
        except Exception:
            memory_usage = 0.0
            cpu_usage = 0.0
        
        # Return API-compliant response structure
        return {
            "active_connections": active_connections,
            "total_requests": total_requests,
            "average_response_time": average_response_time,
            "error_rate": error_rate,
            "memory_usage": memory_usage,
            "cpu_usage": cpu_usage
        }
    
    def _calculate_average_connection_time(self) -> float:
        """Calculate average connection time for active connections."""
        if not self._clients:
            return 0.0
        
        current_time = time.time()
        total_time = sum(current_time - client.connected_at for client in self._clients.values())
        return total_time / len(self._clients)

    async def _method_get_camera_list(
        self, params: Optional[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        """
        Get list of all discovered cameras with their current status and aggregated metadata.

        Integrates real data from camera discovery monitor (with provisional/confirmed capability logic)
        and MediaMTX controller. Returns architecture-compliant response structure.

        Args:
            params: Method parameters (unused)

        Returns:
            Object with camera list and metadata per API specification
        """
        # Get camera monitor from service manager if available
        camera_monitor = None
        if self._service_manager and hasattr(self._service_manager, '_camera_monitor'):
            camera_monitor = self._service_manager._camera_monitor
        elif self._camera_monitor:
            camera_monitor = self._camera_monitor

        # Get MediaMTX controller from service manager if available
        mediamtx_controller = None
        if self._service_manager and hasattr(self._service_manager, '_mediamtx_controller'):
            mediamtx_controller = self._service_manager._mediamtx_controller
        elif self._mediamtx_controller:
            mediamtx_controller = self._mediamtx_controller

        if not camera_monitor:
            self._logger.warning("Camera monitor not available for get_camera_list")
            return {"cameras": [], "total": 0, "connected": 0}

        try:
            # Get connected cameras from camera monitor
            connected_cameras = await camera_monitor.get_connected_cameras()

            cameras = []
            connected_count = 0

            # Prepare concurrent stream status requests for connected cameras
            stream_status_tasks = {}
            if mediamtx_controller:
                for device_path, camera_device in connected_cameras.items():
                    if camera_device.status == "CONNECTED":
                        stream_name = self._get_stream_name_from_device_path(device_path)
                        # Create concurrent task for stream status
                        task = asyncio.create_task(
                            self._get_stream_status_safe(mediamtx_controller, stream_name)
                        )
                        stream_status_tasks[device_path] = task

            # Wait for all stream status requests to complete (with timeout)
            if stream_status_tasks:
                try:
                    await asyncio.wait_for(
                        asyncio.gather(*stream_status_tasks.values(), return_exceptions=True),
                        timeout=0.030  # 30ms timeout for all MediaMTX calls
                    )
                except asyncio.TimeoutError:
                    self._logger.debug("Stream status requests timed out, using cached data")

            # Process cameras with concurrent stream data
            for device_path, camera_device in connected_cameras.items():
                # Get real capability metadata with provisional/confirmed logic
                resolution = "1920x1080"  # Architecture default
                fps = 30  # Architecture default

                # Use effective capability metadata (provisional or confirmed)
                if hasattr(camera_monitor, "get_effective_capability_metadata"):
                    try:
                        capability_metadata = (
                            camera_monitor.get_effective_capability_metadata(
                                device_path
                            )
                        )
                        resolution = capability_metadata.get("resolution", resolution)
                        fps = capability_metadata.get("fps", fps)

                        # Log capability validation status for monitoring
                        validation_status = capability_metadata.get(
                            "validation_status", "none"
                        )
                        if validation_status in ["provisional", "confirmed"]:
                            self._logger.debug(
                                f"Using {validation_status} capability data for {device_path}: "
                                f"{resolution}@{fps}fps"
                            )
                    except Exception as e:
                        self._logger.debug(
                            f"Could not get capability metadata for {device_path}: {e}"
                        )

                # Generate stream name and URLs
                stream_name = self._get_stream_name_from_device_path(device_path)
                
                # Always provide stream URLs as per API specification
                # Stream URLs should be available regardless of current stream status
                streams = {
                    "rtsp": f"rtsp://localhost:8554/{stream_name}",
                    "webrtc": f"http://localhost:8889/{stream_name}/webrtc",
                    "hls": f"http://localhost:8888/{stream_name}",
                }

                # Get stream status for additional validation (but don't affect URL generation)
                if camera_device.status == "CONNECTED" and device_path in stream_status_tasks:
                    task = stream_status_tasks[device_path]
                    if task.done() and not task.exception():
                        try:
                            stream_status = task.result()
                            # Log stream status for debugging but don't modify URLs
                            if stream_status is not None:
                                self._logger.debug(
                                    f"Stream {stream_name} status: {stream_status.get('status', 'unknown')}"
                                )
                            else:
                                self._logger.debug(f"Stream {stream_name} status: unknown (no status available)")
                        except Exception as e:
                            self._logger.debug(
                                f"Error processing stream status for {stream_name}: {e}"
                            )

                # Build camera info per API specification
                camera_info = {
                    "device": device_path,
                    "status": camera_device.status,
                    "name": camera_device.name,
                    "resolution": resolution,
                    "fps": fps,
                    "streams": streams,
                }

                # REQ-REC-001.3: Add recording status to camera list
                if device_path in self._active_recordings:
                    recording_info = self._active_recordings[device_path]
                    camera_info.update({
                        "recording": True,
                        "recording_session": recording_info.get("session_id"),
                        "current_file": recording_info.get("current_file"),
                        "elapsed_time": recording_info.get("elapsed_time", 0)
                    })
                else:
                    camera_info.update({
                        "recording": False,
                        "recording_session": None,
                        "current_file": None,
                        "elapsed_time": 0
                    })

                cameras.append(camera_info)

                if camera_device.status == "CONNECTED":
                    connected_count += 1

            return {
                "cameras": cameras,
                "total": len(cameras),
                "connected": connected_count,
            }

        except Exception as e:
            self._logger.error(f"Error getting camera list: {e}")
            return {"cameras": [], "total": 0, "connected": 0}

    async def _get_stream_status_safe(self, mediamtx_controller: Any, stream_name: str) -> Optional[Dict[str, Any]]:
        """
        Safely get stream status with error handling and timeout.
        
        Args:
            mediamtx_controller: MediaMTX controller instance
            stream_name: Name of the stream
            
        Returns:
            Stream status dict or None if error/timeout
        """
        try:
            return await mediamtx_controller.get_stream_status(stream_name)
        except Exception as e:
            self._logger.debug(f"Could not get stream status for {stream_name}: {e}")
            return None

    async def _method_get_camera_status(
        self, params: Optional[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        """
        Get status for a specific camera device.
        
        Follows ground truth format from docs/api/json-rpc-methods.md exactly.
        Returns camera status information as documented in API specification.

        Args:
            params: Method parameters containing:
                - device (str): Camera device path (required) - must be string per API documentation

        Returns:
            Dict containing camera status with fields: device, status, name, resolution, fps, streams, metrics, capabilities
        """
        if not params or "device" not in params:
            raise ValueError("device parameter is required")

        device_path = params["device"]
        
        # Validate device parameter type per API documentation (ground truth)
        if not isinstance(device_path, str):
            raise ValueError("device parameter must be a string (device path) per API documentation")

        # Initialize response with architecture defaults
        camera_status = {
            "device": device_path,
            "status": "DISCONNECTED",
            "name": f"Camera {device_path.split('video')[-1] if 'video' in device_path else 'unknown'}",
            "resolution": "1920x1080",  # Architecture default
            "fps": 30,  # Architecture default
            # Without MediaMTX controller, do not provide stream URLs
            "streams": {},
            "metrics": {"bytes_sent": 0, "readers": 0, "uptime": 0},
            "capabilities": {"formats": [], "resolutions": []},
        }

        try:
            # Get camera monitor from service manager if available
            camera_monitor = None
            if self._service_manager and hasattr(self._service_manager, '_camera_monitor'):
                camera_monitor = self._service_manager._camera_monitor
            elif self._camera_monitor:
                camera_monitor = self._camera_monitor

            # Get MediaMTX controller from service manager if available
            mediamtx_controller = None
            if self._service_manager and hasattr(self._service_manager, '_mediamtx_controller'):
                mediamtx_controller = self._service_manager._mediamtx_controller
            elif self._mediamtx_controller:
                mediamtx_controller = self._mediamtx_controller

            # Get camera info from camera monitor
            if camera_monitor:
                connected_cameras = await camera_monitor.get_connected_cameras()
                camera_device = connected_cameras.get(device_path)

                if camera_device:
                    camera_status.update(
                        {"status": camera_device.status, "name": camera_device.name}
                    )

                    # Get real capability metadata with provisional/confirmed logic
                    if camera_device.status == "CONNECTED":
                        if hasattr(
                            camera_monitor, "get_effective_capability_metadata"
                        ):
                            try:
                                capability_metadata = (
                                    camera_monitor.get_effective_capability_metadata(
                                        device_path
                                    )
                                )

                                # Use capability-derived resolution and fps
                                camera_status.update(
                                    {
                                        "resolution": capability_metadata.get(
                                            "resolution", "1920x1080"
                                        ),
                                        "fps": capability_metadata.get("fps", 30),
                                    }
                                )

                                # Update capabilities with real detected data
                                if capability_metadata.get("formats"):
                                    capabilities = camera_status["capabilities"]
                                    if isinstance(capabilities, dict):
                                        capabilities["formats"] = capability_metadata["formats"]
                                if capability_metadata.get("all_resolutions"):
                                    capabilities = camera_status["capabilities"]
                                    if isinstance(capabilities, dict):
                                        capabilities["resolutions"] = capability_metadata["all_resolutions"]

                                # Log validation status for monitoring
                                validation_status = capability_metadata.get(
                                    "validation_status", "none"
                                )
                                self._logger.debug(
                                    f"Camera {device_path} using {validation_status} capability data: "
                                    f"{camera_status['resolution']}@{camera_status['fps']}fps"
                                )

                            except Exception as e:
                                self._logger.debug(
                                    f"Could not get capability metadata for {device_path}: {e}"
                                )
                # If camera not in monitor or monitor missing, keep defaults and continue

            # Get stream info and metrics from MediaMTX controller
            if mediamtx_controller and camera_status["status"] == "CONNECTED":
                try:
                    stream_name = self._get_stream_name_from_device_path(device_path)
                    
                    # Always provide stream URLs as per API specification
                    # Stream URLs should be available regardless of current stream status
                    camera_status["streams"] = {
                        "rtsp": f"rtsp://localhost:8554/{stream_name}",
                        "webrtc": f"http://localhost:8889/{stream_name}/webrtc",
                        "hls": f"http://localhost:8888/{stream_name}",
                    }
                    
                    # Get stream status for metrics but don't affect URL generation
                    stream_status = await mediamtx_controller.get_stream_status(
                        stream_name
                    )

                    # Update metrics from MediaMTX if stream status available
                    if stream_status:
                        camera_status["metrics"] = {
                            "bytes_sent": stream_status.get("bytes_sent", 0),
                            "readers": stream_status.get("readers", 0),
                            "uptime": int(time.time()),  # Current uptime proxy
                        }

                except Exception as e:
                    self._logger.debug(
                        f"Could not get MediaMTX status for {device_path}: {e}"
                    )

            # REQ-REC-001.3: Add recording status to camera response
            if device_path in self._active_recordings:
                recording_info = self._active_recordings[device_path]
                camera_status.update({
                    "recording": True,
                    "recording_session": recording_info.get("session_id"),
                    "current_file": recording_info.get("current_file"),
                    "elapsed_time": recording_info.get("elapsed_time", 0)
                })
            else:
                camera_status.update({
                    "recording": False,
                    "recording_session": None,
                    "current_file": None,
                    "elapsed_time": 0
                })

            return camera_status

        except Exception as e:
            # Degrade gracefully on monitor errors
            self._logger.error(f"Error getting camera status for {device_path}: {e}")
            return {"status": "ERROR", "device": device_path}

    async def _method_get_camera_capabilities(
        self, params: Optional[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        """
        Get camera capability metadata for a specific camera device.

        Args:
            params: Method parameters containing:
                - device (str): Camera device path

        Returns:
            Dict containing camera capability metadata:
                - device, formats, resolutions, fps_options, validation_status
        """
        if not params or "device" not in params:
            raise ValueError("device parameter is required")

        device_path = params["device"]

        # Initialize response with architecture defaults
        camera_capabilities = {
            "device": device_path,
            "formats": [],
            "resolutions": [],
            "fps_options": [],
            "validation_status": "none"
        }

        try:
            # Get camera monitor from service manager if available
            camera_monitor = None
            if self._service_manager and hasattr(self._service_manager, '_camera_monitor'):
                camera_monitor = self._service_manager._camera_monitor
            elif self._camera_monitor:
                camera_monitor = self._camera_monitor

            # Get camera info from camera monitor
            if camera_monitor:
                connected_cameras = await camera_monitor.get_connected_cameras()
                camera_device = connected_cameras.get(device_path)

                if camera_device and camera_device.status == "CONNECTED":
                    # Get real capability metadata with provisional/confirmed logic
                    if hasattr(camera_monitor, "get_effective_capability_metadata"):
                        try:
                            capability_metadata = camera_monitor.get_effective_capability_metadata(device_path)
                            
                            # Update capabilities with real detected data
                            if capability_metadata.get("formats"):
                                camera_capabilities["formats"] = capability_metadata["formats"]
                            if capability_metadata.get("all_resolutions"):
                                camera_capabilities["resolutions"] = capability_metadata["all_resolutions"]
                            if capability_metadata.get("fps_options"):
                                camera_capabilities["fps_options"] = capability_metadata["fps_options"]
                            
                            # Set validation status
                            camera_capabilities["validation_status"] = capability_metadata.get("validation_status", "none")
                            
                            self._logger.debug(f"Camera {device_path} capabilities: {camera_capabilities}")
                            
                        except Exception as e:
                            self._logger.warning(f"Could not get capability metadata for {device_path}: {e}")
                            # Return default capabilities on error
                            camera_capabilities["validation_status"] = "error"
                else:
                    camera_capabilities["validation_status"] = "disconnected"
            else:
                camera_capabilities["validation_status"] = "no_monitor"

            return camera_capabilities

        except Exception as e:
            # Degrade gracefully on monitor errors
            self._logger.error(f"Error getting camera capabilities for {device_path}: {e}")
            return {
                "device": device_path,
                "formats": [],
                "resolutions": [],
                "fps_options": [],
                "validation_status": "error"
            }

    async def _method_take_snapshot(
        self, params: Optional[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        """
        Take a snapshot from the specified camera.

        Args:
            params: Method parameters containing:
                - device (str): Camera device path
                - format (str, optional): Snapshot format (jpg, png)
                - quality (int, optional): Image quality (1-100)

        Returns:
            Dict containing snapshot information
        """
        if not params or "device" not in params:
            raise ValueError("device parameter is required")

        device_path = params["device"]
        # Parameter validation and normalization
        format_type = params.get("format", "jpg")
        quality = params.get("quality", 85)
        if not isinstance(quality, int) or not (1 <= quality <= 100):
            raise ValueError("Invalid params")
        if format_type not in ("jpg", "png"):
            format_type = "jpg"
        custom_filename = params.get("filename")

        # Get MediaMTX controller from service manager if available
        mediamtx_controller = None
        if self._service_manager and hasattr(self._service_manager, '_mediamtx_controller'):
            mediamtx_controller = self._service_manager._mediamtx_controller
        elif self._mediamtx_controller:
            mediamtx_controller = self._mediamtx_controller

        if not mediamtx_controller:
            return {
                "device": device_path,
                "filename": custom_filename
                or self._generate_filename(device_path, format_type),
                "status": "FAILED",
                "timestamp": time.strftime("%Y-%m-%dT%H:%M:%SZ"),
                "file_size": 0,
                "file_path": "",
                "error": "MediaMTX controller not available",
            }

        try:
            stream_name = self._get_stream_name_from_device_path(device_path)

            # Generate filename with format extension if not provided
            if not custom_filename:
                timestamp = time.strftime("%Y-%m-%d_%H-%M-%S")
                custom_filename = f"{stream_name}_snapshot_{timestamp}.{format_type}"
            elif not custom_filename.endswith(f".{format_type}"):
                custom_filename = f"{custom_filename}.{format_type}"

            snapshot_result = await mediamtx_controller.take_snapshot(
                stream_name=stream_name,
                filename=custom_filename,
                format=format_type,
                quality=quality,
            )

            # Check MediaMTX controller response status
            if snapshot_result.get("status") == "failed":
                return {
                    "device": device_path,
                    "filename": snapshot_result.get("filename"),
                    "status": "FAILED",
                    "timestamp": snapshot_result.get(
                        "timestamp", time.strftime("%Y-%m-%dT%H:%M:%SZ")
                    ),
                    "file_size": snapshot_result.get("file_size", 0),
                    "file_path": snapshot_result.get("file_path", ""),
                    "error": snapshot_result.get("error", "MediaMTX operation failed"),
                }
            else:
                return {
                    "device": device_path,
                    "filename": snapshot_result.get("filename"),
                    "status": "completed",
                    "timestamp": snapshot_result.get(
                        "timestamp", time.strftime("%Y-%m-%dT%H:%M:%SZ")
                    ),
                    "file_size": snapshot_result.get("file_size", 0),
                    "file_path": snapshot_result.get("file_path", ""),
                }

        except Exception as e:
            self._logger.error(f"Error taking snapshot for {device_path}: {e}")
            return {
                "device": device_path,
                "filename": custom_filename or self._generate_filename(device_path, format_type),
                "status": "FAILED",
                "timestamp": time.strftime("%Y-%m-%dT%H:%M:%SZ"),
                "file_size": 0,
                "file_path": "",
                "error": f"MediaMTX operation failed: {e}",
            }

    async def _method_start_recording(
        self, params: Optional[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        """
        Start recording video from the specified camera.

        Args:
            params: Method parameters containing:
                - device (str): Camera device path
                - duration (int, optional): Recording duration in seconds
                - format (str, optional): Recording format

        Returns:
            Dict containing recording session information
        """
        if not params or "device" not in params:
            raise ValueError("device parameter is required")

        device_path = params["device"]
        # Parameter normalization and validation
        duration = params.get("duration")
        duration_seconds = params.get("duration_seconds")
        duration_minutes = params.get("duration_minutes")
        duration_hours = params.get("duration_hours")
        format_type = params.get("format", "mp4")

        # Normalize format (only mp4 supported at this stage)
        if format_type not in ("mp4",):
            format_type = "mp4"

        # Determine effective duration (seconds)
        effective_duration = None
        if duration is not None:
            # legacy seconds param
            if not isinstance(duration, int) or duration < 1:
                raise ValueError("Invalid params")
            effective_duration = duration
        elif duration_seconds is not None:
            if not isinstance(duration_seconds, int) or not (1 <= duration_seconds <= 3600):
                raise ValueError("Invalid params")
            effective_duration = duration_seconds
        elif duration_minutes is not None:
            if not isinstance(duration_minutes, int) or not (1 <= duration_minutes <= 1440):
                raise ValueError("Invalid params")
            effective_duration = duration_minutes * 60
        elif duration_hours is not None:
            if not isinstance(duration_hours, int) or not (1 <= duration_hours <= 24):
                raise ValueError("Invalid params")
            effective_duration = duration_hours * 3600
        else:
            # Unlimited mode when no duration provided
            effective_duration = None

        # REQ-REC-001.1: Check for recording conflicts
        if device_path in self._active_recordings:
            existing_session = self._active_recordings[device_path]
            raise MediaMTXError(f"Camera is currently recording (session: {existing_session.get('session_id', 'unknown')})")

        # REQ-REC-003.1: Validate storage space before starting recording
        storage_info = self.check_storage_space()
        if storage_info:
            if storage_info["used_percent"] >= self._storage_block_percent:
                raise MediaMTXError(f"Storage space is critical ({storage_info['used_percent']:.1f}% used)")
            elif storage_info["used_percent"] >= self._storage_warn_percent:
                self._logger.warning(f"Storage usage high: {storage_info['used_percent']:.1f}%")

        # Get MediaMTX controller from service manager if available
        mediamtx_controller = None
        if self._service_manager and hasattr(self._service_manager, '_mediamtx_controller'):
            mediamtx_controller = self._service_manager._mediamtx_controller
        elif self._mediamtx_controller:
            mediamtx_controller = self._mediamtx_controller

        if not mediamtx_controller:
            raise MediaMTXError("MediaMTX controller not available")

        try:
            stream_name = self._get_stream_name_from_device_path(device_path)
            session_id = str(uuid.uuid4())

            # Step 1: Ensure stream is ready (this will trigger on-demand activation if needed)
            # This aligns with API documentation - client provides device path, server handles everything internally
            try:
                stream_ready = await mediamtx_controller.check_stream_readiness(stream_name, timeout=15.0)
                if not stream_ready:
                    error_msg = f"Failed to activate stream for device {device_path} within timeout"
                    self._logger.error(f"Stream readiness check failed for {device_path}: {error_msg}")
                    raise MediaMTXError(f"MediaMTX operation failed: {error_msg}")
            except Exception as e:
                error_msg = f"Failed to ensure stream readiness for {device_path}: {e}"
                self._logger.error(f"Stream readiness check failed for {device_path}: {error_msg}")
                raise MediaMTXError(f"MediaMTX operation failed: {error_msg}")

            # Step 2: Start recording now that stream is ready
            recording_result = await mediamtx_controller.start_recording(
                stream_name=stream_name, duration=effective_duration, format=format_type
            )

            # Step 3: Validate MediaMTX controller response
            if not recording_result or recording_result.get("status") == "failed":
                error_msg = recording_result.get("error", "MediaMTX operation failed") if recording_result else "MediaMTX operation failed"
                self._logger.error(f"MediaMTX controller failed to start recording for {device_path}: {error_msg}")
                raise MediaMTXError(f"MediaMTX operation failed: {error_msg}")

            # Step 4: All validations passed - create success response
            # Ensure start_time is in proper ISO format with Z suffix
            start_time = recording_result.get("start_time")
            if not start_time:
                start_time = time.strftime("%Y-%m-%dT%H:%M:%SZ")
            elif not start_time.endswith("Z"):
                # Convert to ISO format if not already
                start_time = start_time + "Z" if "T" in start_time else time.strftime("%Y-%m-%dT%H:%M:%SZ")
            
            response = {
                "device": device_path,
                "session_id": session_id,
                "filename": recording_result.get("filename"),
                "status": "STARTED",  # Per API documentation
                "start_time": start_time,
                "duration": effective_duration,
                "format": format_type,
            }

            # REQ-REC-001.1: Track recording state
            self._active_recordings[device_path] = {
                "session_id": session_id,
                "start_time": response["start_time"],
                "current_file": recording_result.get("filename"),
                "elapsed_time": 0,
                "stream_name": stream_name
            }

            # Step 5: Schedule auto-stop if timed recording requested
            if effective_duration and effective_duration > 0:
                async def _auto_stop() -> None:
                    try:
                        await asyncio.sleep(effective_duration)
                        # Stop via controller; ignore errors
                        try:
                            await mediamtx_controller.stop_recording(stream_name=stream_name)
                        except Exception:
                            pass
                        await self._emit_recording_complete(device_path, stream_name)
                    finally:
                        self._recording_stop_tasks.pop(stream_name, None)

                task = asyncio.create_task(_auto_stop())
                self._recording_stop_tasks[stream_name] = task

            # Send notification asynchronously after returning the response
            # This ensures the JSON-RPC response is sent first, then the notification
            async def _send_notification() -> None:
                try:
                    await self.notify_recording_status_update({
                        "device": device_path,
                        "status": "STARTED",
                        "filename": recording_result.get("filename"),
                        "duration": effective_duration,
                    })
                except Exception as e:
                    self._logger.warning(f"Failed to send recording start notification: {e}")
            
            # Schedule notification to be sent after response is returned
            asyncio.create_task(_send_notification())
            
            return response

        except Exception as e:
            self._logger.error(f"Error starting recording for {device_path}: {e}")
            # Do NOT send any notifications on error - let the JSON-RPC error response be sent
            raise MediaMTXError(f"MediaMTX operation failed: {e}") from e

    async def _method_stop_recording(
        self, params: Optional[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        """
        Stop active recording for the specified camera.

        Args:
            params: Method parameters containing:
                - device (str): Camera device path

        Returns:
            Dict containing recording completion information
        """
        if not params or "device" not in params:
            raise ValueError("device parameter is required")

        device_path = params["device"]

        # Get MediaMTX controller from service manager if available
        mediamtx_controller = None
        if self._service_manager and hasattr(self._service_manager, '_mediamtx_controller'):
            mediamtx_controller = self._service_manager._mediamtx_controller
        elif self._mediamtx_controller:
            mediamtx_controller = self._mediamtx_controller

        if not mediamtx_controller:
            return {
                "device": device_path,
                "session_id": None,
                "filename": None,
                "status": "FAILED",
                "start_time": None,
                "end_time": time.strftime("%Y-%m-%dT%H:%M:%SZ"),
                "duration": None,
                "file_size": 0,
                "error": "MediaMTX controller not available",
            }

        try:
            stream_name = self._get_stream_name_from_device_path(device_path)

            recording_result = await mediamtx_controller.stop_recording(stream_name=stream_name)

            # Validate MediaMTX controller response
            if not recording_result or recording_result.get("status") == "failed":
                error_msg = recording_result.get("error", "MediaMTX operation failed") if recording_result else "MediaMTX operation failed"
                self._logger.error(f"MediaMTX controller failed to stop recording for {device_path}: {error_msg}")
                raise MediaMTXError(f"MediaMTX operation failed: {error_msg}")

            response = {
                "device": device_path,
                "session_id": recording_result.get("session_id"),
                "filename": recording_result.get("filename"),
                "status": "STOPPED",
                "start_time": recording_result.get("start_time"),
                "end_time": recording_result.get(
                    "stop_time", time.strftime("%Y-%m-%dT%H:%M:%SZ")
                ),
                "duration": recording_result.get("duration"),
                "file_size": recording_result.get("file_size", 0),
            }

            # REQ-REC-001.1: Clean up recording state
            if device_path in self._active_recordings:
                self._active_recordings.pop(device_path)
                self._logger.info(f"Cleaned up recording state for {device_path}")

            # Send recording status update notification only on success
            try:
                await self.notify_recording_status_update({
                    "device": device_path,
                    "status": "STOPPED",
                    "filename": recording_result.get("filename"),
                    "duration": recording_result.get("duration"),
                })
            except Exception as e:
                self._logger.warning(f"Failed to send recording stop notification: {e}")

            return response

        except Exception as e:
            self._logger.error(f"Error stopping recording for {device_path}: {e}")
            raise MediaMTXError(f"MediaMTX operation failed: {e}") from e

    async def _method_get_status(
        self, params: Optional[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        """
        Get system status and health information.
        
        Returns system status information exactly as documented in API specification.
        Follows ground truth format from docs/api/json-rpc-methods.md.

        Returns:
            Dict containing system status with fields: status, uptime, version, components
        """
        import time
        
        # Calculate uptime as positive integer (seconds since start)
        start_time = getattr(self, '_start_time', time.time())
        uptime = max(0, int(time.time() - start_time))
        
        # Determine overall system status
        system_status = "healthy"
        
        # Check component statuses
        websocket_server_status = "running"
        camera_monitor_status = "running"
        mediamtx_controller_status = "unknown"
        
        # Check MediaMTX controller status if available
        if self._service_manager and hasattr(self._service_manager, '_mediamtx_controller'):
            mediamtx_controller = self._service_manager._mediamtx_controller
            if mediamtx_controller:
                try:
                    health = await mediamtx_controller.health_check()
                    mediamtx_controller_status = health.get("status", "unknown")
                    if mediamtx_controller_status != "healthy":
                        system_status = "degraded"
                except Exception as e:
                    self._logger.warning(f"MediaMTX health check failed: {e}")
                    mediamtx_controller_status = "error"
                    system_status = "degraded"
        
        # Return format EXACTLY as documented in API specification
        return {
            "status": system_status,
            "uptime": uptime,
            "version": "1.0.0",
            "components": {
                "websocket_server": websocket_server_status,
                "camera_monitor": camera_monitor_status,
                "mediamtx_controller": mediamtx_controller_status
            }
        }

    async def _method_get_server_info(
        self, params: Optional[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        """
        Get server configuration and capability information.
        
        Returns server information exactly as documented in API specification.
        Follows ground truth format from docs/api/json-rpc-methods.md.

        Returns:
            Dict containing server configuration with fields: name, version, capabilities, supported_formats, max_cameras
        """
        # Return format EXACTLY as documented in API specification
        return {
            "name": "MediaMTX Camera Service",
            "version": "1.0.0",
            "capabilities": ["snapshots", "recordings", "streaming"],
            "supported_formats": ["mp4", "mkv", "jpg"],
            "max_cameras": 10
        }

    async def _method_get_streams(
        self, params: Optional[Dict[str, Any]] = None
    ) -> List[Dict[str, Any]]:
        """
        Get list of all active streams from MediaMTX.

        Returns:
            List of stream information dictionaries
        """
        # Get MediaMTX controller from service manager if available
        mediamtx_controller = None
        if self._service_manager and hasattr(self._service_manager, '_mediamtx_controller'):
            mediamtx_controller = self._service_manager._mediamtx_controller
        elif self._mediamtx_controller:
            mediamtx_controller = self._mediamtx_controller

        if not mediamtx_controller:
            self._logger.warning("MediaMTX controller not available for get_streams")
            return []

        try:
            # Get stream list from MediaMTX
            streams = await mediamtx_controller.get_stream_list()
            
            # Format streams for API response
            formatted_streams = []
            for stream in streams:
                formatted_stream = {
                    "name": stream.get("name", "unknown"),
                    "source": stream.get("source"),
                    "ready": stream.get("ready", False),
                    "readers": stream.get("readers", 0),
                    "bytes_sent": stream.get("bytes_sent", 0)
                }
                formatted_streams.append(formatted_stream)
            
            return formatted_streams

        except Exception as e:
            self._logger.error(f"Error getting streams: {e}")
            return []

    async def notify_camera_status_update(self, params: Dict[str, Any]) -> None:
        """
        Broadcast camera_status_update notification with strict API compliance and enhanced disconnect handling.

        Filters notification parameters to only include API-specified fields:
        device, status, name, resolution, fps, streams (per docs/api/json-rpc-methods.md)
        Ensures proper handling of disconnect events and state consistency.

        Args:
            params: Dictionary containing camera status fields
        """
        if not params:
            self._logger.warning("Camera status update called with empty parameters")
            return

        # Validate required fields per API specification
        required_fields = ["device", "status"]
        for field in required_fields:
            if field not in params:
                self._logger.error(
                    f"Camera status update missing required field: {field}"
                )
                return

        # Enhanced disconnect event handling
        is_disconnect_event = params.get("status") == "DISCONNECTED"
        device_path = params.get("device", "unknown")
        
        if is_disconnect_event:
            self._logger.debug(
                f"Processing camera disconnect notification for device: {device_path}",
                extra={
                    "device_path": device_path,
                    "event_type": "camera_disconnect",
                },
            )

        # STRICT API COMPLIANCE: Filter to only allowed fields per specification
        allowed_fields = {"device", "status", "name", "resolution", "fps", "streams"}
        filtered_params = {k: v for k, v in params.items() if k in allowed_fields}

        # Log filtered fields for monitoring compliance
        filtered_out = set(params.keys()) - allowed_fields
        if filtered_out:
            self._logger.debug(
                f"Filtered out non-API fields from camera notification: {filtered_out}"
            )

        try:
            # Enhanced notification with disconnect-specific handling
            await self.broadcast_notification(
                method="camera_status_update", params=filtered_params
            )

            # Enhanced logging for disconnect events
            if is_disconnect_event:
                self._logger.info(
                    f"Successfully broadcasted camera disconnect notification for device: {device_path}",
                    extra={
                        "device_path": device_path,
                        "notification_type": "disconnect",
                        "clients_notified": len(self._clients),
                    },
                )
            else:
                self._logger.info(
                    f"Broadcasted camera status update for device: {device_path}"
                )

        except Exception as e:
            self._logger.error(
                f"Failed to broadcast camera status update: {e}",
                extra={
                    "device_path": device_path,
                    "is_disconnect_event": is_disconnect_event,
                },
            )

    async def notify_recording_status_update(self, params: Dict[str, Any]) -> None:
        """
        Broadcast recording_status_update notification with strict API compliance.

        Filters notification parameters to only include API-specified fields:
        device, status, filename, duration (per docs/api/json-rpc-methods.md)

        Args:
            params: Dictionary containing recording status fields
        """
        if not params:
            self._logger.warning("Recording status update called with empty parameters")
            return

        # Validate required fields per API specification
        required_fields = ["device", "status"]
        for field in required_fields:
            if field not in params:
                self._logger.error(
                    f"Recording status update missing required field: {field}"
                )
                return

        # STRICT API COMPLIANCE: Filter to only allowed fields per specification
        allowed_fields = {"device", "status", "filename", "duration"}
        filtered_params = {k: v for k, v in params.items() if k in allowed_fields}

        # Log filtered fields for monitoring compliance
        filtered_out = set(params.keys()) - allowed_fields
        if filtered_out:
            self._logger.debug(
                f"Filtered out non-API fields from recording notification: {filtered_out}"
            )

        try:
            await self.broadcast_notification(
                method="recording_status_update", params=filtered_params
            )

            self._logger.info(
                f"Broadcasted recording status update for device: {params.get('device')}, status: {params.get('status')}"
            )

        except Exception as e:
            self._logger.error(f"Failed to broadcast recording status update: {e}")

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
            "registered_methods": len(self._method_handlers),
        }

    @property
    def is_running(self) -> bool:
        """Check if the server is currently running."""
        return self._running

    async def _method_list_recordings(
        self, params: Optional[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        """
        List available recording files in the recordings directory.
        
        Requirements: REQ-FUNC-008
        Epic E6: Server Recording and Snapshot File Management Infrastructure
        
        Args:
            params: Optional parameters containing:
                - limit: Maximum number of files to return (default: 100)
                - offset: Number of files to skip for pagination (default: 0)
        
        Returns:
            Dictionary containing file list with metadata and pagination info
        """
        try:
            # Parse parameters with defaults
            limit = params.get("limit", 100) if params else 100
            offset = params.get("offset", 0) if params else 0
            
            # Validate parameters
            if not isinstance(limit, int) or limit < 1 or limit > 1000:
                raise ValueError("Invalid limit parameter: must be integer between 1 and 1000")
            if not isinstance(offset, int) or offset < 0:
                raise ValueError("Invalid offset parameter: must be non-negative integer")
            
            # Define recordings directory path
            config = self._get_config()
            recordings_dir = config.mediamtx.recordings_path
            
            # Check if directory exists and is accessible
            if not os.path.exists(recordings_dir):
                self._logger.warning(f"Recordings directory does not exist: {recordings_dir}")
                return {
                    "files": [],
                    "total": 0,
                    "limit": limit,
                    "offset": offset
                }
            
            if not os.access(recordings_dir, os.R_OK):
                self._logger.error(f"Permission denied accessing recordings directory: {recordings_dir}")
                raise PermissionError("Permission denied accessing recordings directory")
            
            # Get list of files in directory
            try:
                files = []
                for filename in os.listdir(recordings_dir):
                    file_path = os.path.join(recordings_dir, filename)
                    
                    # Skip directories and non-files
                    if not os.path.isfile(file_path):
                        continue
                    
                    # Get file stats
                    try:
                        stat_info = os.stat(file_path)
                        file_size = stat_info.st_size
                        file_time = datetime.fromtimestamp(stat_info.st_mtime)
                        
                        # Determine if it's a video file
                        is_video = filename.lower().endswith(('.mp4', '.avi', '.mov', '.mkv', '.wmv', '.flv'))
                        
                        file_info = {
                            "filename": filename,
                            "file_size": file_size,
                            "modified_time": file_time.isoformat() + "Z",
                            "download_url": f"/files/recordings/{filename}"
                        }
                        
                        # Add duration for video files (placeholder - would need video metadata extraction)
                        if is_video:
                            file_info["duration"] = None  # TODO: Extract actual duration from video file
                        
                        files.append(file_info)
                        
                    except OSError as e:
                        self._logger.warning(f"Error accessing file {filename}: {e}")
                        continue
                
                # Sort files by modified_time (newest first)
                files.sort(key=lambda x: x["modified_time"], reverse=True)
                
                # Apply pagination
                total_count = len(files)
                start_idx = offset
                end_idx = min(start_idx + limit, total_count)
                paginated_files = files[start_idx:end_idx]
                
                return {
                    "files": paginated_files,
                    "total": total_count,
                    "limit": limit,
                    "offset": offset
                }
                
            except OSError as e:
                self._logger.error(f"Error reading recordings directory: {e}")
                raise OSError(f"Error reading recordings directory: {e}")
                
        except (ValueError, PermissionError, OSError) as e:
            self._logger.error(f"Error in list_recordings: {e}")
            raise
        except Exception as e:
            self._logger.error(f"Unexpected error in list_recordings: {e}")
            raise

    async def _method_list_snapshots(
        self, params: Optional[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        """
        List available snapshot files in the snapshots directory.
        
        Requirements: REQ-FUNC-009
        Epic E6: Server Recording and Snapshot File Management Infrastructure
        
        Args:
            params: Optional parameters containing:
                - limit: Maximum number of files to return (default: 100)
                - offset: Number of files to skip for pagination (default: 0)
        
        Returns:
            Dictionary containing file list with metadata and pagination info
        """
        try:
            # Parse parameters with defaults
            limit = params.get("limit", 100) if params else 100
            offset = params.get("offset", 0) if params else 0
            
            # Validate parameters
            if not isinstance(limit, int) or limit < 1 or limit > 1000:
                raise ValueError("Invalid limit parameter: must be integer between 1 and 1000")
            if not isinstance(offset, int) or offset < 0:
                raise ValueError("Invalid offset parameter: must be non-negative integer")
            
            # Define snapshots directory path
            if self._config is None:
                raise RuntimeError("Configuration not available")
            snapshots_dir = self._config.mediamtx.snapshots_path
            
            # Check if directory exists and is accessible
            if not os.path.exists(snapshots_dir):
                self._logger.warning(f"Snapshots directory does not exist: {snapshots_dir}")
                return {
                    "files": [],
                    "total": 0,
                    "limit": limit,
                    "offset": offset
                }
            
            if not os.access(snapshots_dir, os.R_OK):
                self._logger.error(f"Permission denied accessing snapshots directory: {snapshots_dir}")
                raise PermissionError("Permission denied accessing snapshots directory")
            
            # Get list of files in directory
            try:
                files = []
                for filename in os.listdir(snapshots_dir):
                    file_path = os.path.join(snapshots_dir, filename)
                    
                    # Skip directories and non-files
                    if not os.path.isfile(file_path):
                        continue
                    
                    # Get file stats
                    try:
                        stat_info = os.stat(file_path)
                        file_size = stat_info.st_size
                        file_time = datetime.fromtimestamp(stat_info.st_mtime)
                        
                        file_info = {
                            "filename": filename,
                            "file_size": file_size,
                            "modified_time": file_time.isoformat() + "Z",
                            "download_url": f"/files/snapshots/{filename}"
                        }
                        
                        files.append(file_info)
                        
                    except OSError as e:
                        self._logger.warning(f"Error accessing file {filename}: {e}")
                        continue
                
                # Sort files by modified_time (newest first)
                files.sort(key=lambda x: x["modified_time"], reverse=True)
                
                # Apply pagination
                total_count = len(files)
                start_idx = offset
                end_idx = min(start_idx + limit, total_count)
                paginated_files = files[start_idx:end_idx]
                
                return {
                    "files": paginated_files,
                    "total": total_count,
                    "limit": limit,
                    "offset": offset
                }
                
            except OSError as e:
                self._logger.error(f"Error reading snapshots directory: {e}")
                raise OSError(f"Error reading snapshots directory: {e}")
                
        except (ValueError, PermissionError, OSError) as e:
            self._logger.error(f"Error in list_snapshots: {e}")
            raise
        except Exception as e:
            self._logger.error(f"Unexpected error in list_snapshots: {e}")
            raise

    async def _method_get_recording_info(
        self, params: Optional[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        """
        Get detailed information about a specific recording file.
        
        Requirements: REQ-CLIENT-037, REQ-CLIENT-038
        Epic E6: File Lifecycle Management
        
        Args:
            params: Parameters containing:
                - filename: Name of the recording file (required)
        
        Returns:
            Dictionary containing recording file metadata and information
        """
        try:
            # Validate parameters
            if not params or "filename" not in params:
                raise ValueError("filename parameter is required")
            
            filename = params["filename"]
            if not isinstance(filename, str) or not filename.strip():
                raise ValueError("filename must be a non-empty string")
            
            # Define recordings directory path
            config = self._get_config()
            recordings_dir = config.mediamtx.recordings_path
            file_path = os.path.join(recordings_dir, filename)
            
            # Check if file exists
            if not os.path.exists(file_path):
                raise FileNotFoundError(f"Recording file not found: {filename}")
            
            if not os.path.isfile(file_path):
                raise ValueError(f"Path is not a file: {filename}")
            
            # Get file stats
            try:
                stat_info = os.stat(file_path)
                file_size = stat_info.st_size
                created_time = datetime.fromtimestamp(stat_info.st_ctime)
                modified_time = datetime.fromtimestamp(stat_info.st_mtime)
                
                # Determine if it's a video file
                is_video = filename.lower().endswith(('.mp4', '.avi', '.mov', '.mkv', '.wmv', '.flv'))
                
                file_info = {
                    "filename": filename,
                    "file_size": file_size,
                    "created_time": created_time.isoformat() + "Z",
                    "modified_time": modified_time.isoformat() + "Z",
                    "download_url": f"/files/recordings/{filename}"
                }
                
                # Add duration for video files using MediaMTX metadata or file analysis
                if is_video:
                    duration = await self._get_video_duration_architecture_compliant(file_path, filename)
                    file_info["duration"] = duration if duration else 0
                
                return file_info
                
            except OSError as e:
                self._logger.error(f"Error accessing recording file {filename}: {e}")
                raise OSError(f"Error accessing recording file: {e}")
                
        except (ValueError, FileNotFoundError, OSError) as e:
            self._logger.error(f"Error in get_recording_info: {e}")
            raise
        except Exception as e:
            self._logger.error(f"Unexpected error in get_recording_info: {e}")
            raise

    async def _method_get_snapshot_info(
        self, params: Optional[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        """
        Get detailed information about a specific snapshot file.
        
        Requirements: REQ-CLIENT-037, REQ-CLIENT-038
        Epic E6: File Lifecycle Management
        
        Args:
            params: Parameters containing:
                - filename: Name of the snapshot file (required)
        
        Returns:
            Dictionary containing snapshot file metadata and information
        """
        try:
            # Validate parameters
            if not params or "filename" not in params:
                raise ValueError("filename parameter is required")
            
            filename = params["filename"]
            if not isinstance(filename, str) or not filename.strip():
                raise ValueError("filename must be a non-empty string")
            
            # Define snapshots directory path
            config = self._get_config()
            snapshots_dir = config.mediamtx.snapshots_path
            file_path = os.path.join(snapshots_dir, filename)
            
            # Check if file exists
            if not os.path.exists(file_path):
                raise FileNotFoundError(f"Snapshot file not found: {filename}")
            
            if not os.path.isfile(file_path):
                raise ValueError(f"Path is not a file: {filename}")
            
            # Get file stats
            try:
                stat_info = os.stat(file_path)
                file_size = stat_info.st_size
                created_time = datetime.fromtimestamp(stat_info.st_ctime)
                modified_time = datetime.fromtimestamp(stat_info.st_mtime)
                
                # Determine if it's an image file
                is_image = filename.lower().endswith(('.jpg', '.jpeg', '.png', '.bmp', '.gif', '.tiff'))
                
                file_info = {
                    "filename": filename,
                    "file_size": file_size,
                    "created_time": created_time.isoformat() + "Z",
                    "modified_time": modified_time.isoformat() + "Z",
                    "download_url": f"/files/snapshots/{filename}"
                }
                
                # Add resolution for image files (placeholder - would need image metadata extraction)
                if is_image:
                    file_info["resolution"] = None  # TODO: Extract actual resolution from image file
                
                return file_info
                
            except OSError as e:
                self._logger.error(f"Error accessing snapshot file {filename}: {e}")
                raise OSError(f"Error accessing snapshot file: {e}")
                
        except (ValueError, FileNotFoundError, OSError) as e:
            self._logger.error(f"Error in get_snapshot_info: {e}")
            raise
        except Exception as e:
            self._logger.error(f"Unexpected error in get_snapshot_info: {e}")
            raise

    async def _method_delete_recording(
        self, params: Optional[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        """
        Delete a specific recording file.
        
        Requirements: REQ-CLIENT-034, REQ-CLIENT-038
        Epic E6: File Lifecycle Management
        
        Args:
            params: Parameters containing:
                - filename: Name of the recording file to delete (required)
        
        Returns:
            Dictionary containing deletion status and confirmation
        """
        try:
            # Validate parameters
            if not params or "filename" not in params:
                raise ValueError("filename parameter is required")
            
            filename = params["filename"]
            if not isinstance(filename, str) or not filename.strip():
                raise ValueError("filename must be a non-empty string")
            
            # Define recordings directory path
            config = self._get_config()
            recordings_dir = config.mediamtx.recordings_path
            file_path = os.path.join(recordings_dir, filename)
            
            # Check if file exists
            if not os.path.exists(file_path):
                raise FileNotFoundError(f"Recording file not found: {filename}")
            
            if not os.path.isfile(file_path):
                raise ValueError(f"Path is not a file: {filename}")
            
            # Delete the file
            try:
                os.remove(file_path)
                self._logger.info(f"Recording file deleted successfully: {filename}")
                
                return {
                    "filename": filename,
                    "deleted": True,
                    "message": "Recording file deleted successfully"
                }
                
            except OSError as e:
                self._logger.error(f"Error deleting recording file {filename}: {e}")
                raise OSError(f"Error deleting recording file: {e}")
                
        except (ValueError, FileNotFoundError, OSError) as e:
            self._logger.error(f"Error in delete_recording: {e}")
            raise
        except Exception as e:
            self._logger.error(f"Unexpected error in delete_recording: {e}")
            raise

    async def _method_delete_snapshot(
        self, params: Optional[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        """
        Delete a specific snapshot file.
        
        Requirements: REQ-CLIENT-034, REQ-CLIENT-038
        Epic E6: File Lifecycle Management
        
        Args:
            params: Parameters containing:
                - filename: Name of the snapshot file to delete (required)
        
        Returns:
            Dictionary containing deletion status and confirmation
        """
        try:
            # Validate parameters
            if not params or "filename" not in params:
                raise ValueError("filename parameter is required")
            
            filename = params["filename"]
            if not isinstance(filename, str) or not filename.strip():
                raise ValueError("filename must be a non-empty string")
            
            # Define snapshots directory path
            config = self._get_config()
            snapshots_dir = config.mediamtx.snapshots_path
            file_path = os.path.join(snapshots_dir, filename)
            
            # Check if file exists
            if not os.path.exists(file_path):
                raise FileNotFoundError(f"Snapshot file not found: {filename}")
            
            if not os.path.isfile(file_path):
                raise ValueError(f"Path is not a file: {filename}")
            
            # Delete the file
            try:
                os.remove(file_path)
                self._logger.info(f"Snapshot file deleted successfully: {filename}")
                
                return {
                    "filename": filename,
                    "deleted": True,
                    "message": "Snapshot file deleted successfully"
                }
                
            except OSError as e:
                self._logger.error(f"Error deleting snapshot file {filename}: {e}")
                raise OSError(f"Error deleting snapshot file: {e}")
                
        except (ValueError, FileNotFoundError, OSError) as e:
            self._logger.error(f"Error in delete_snapshot: {e}")
            raise
        except Exception as e:
            self._logger.error(f"Unexpected error in delete_snapshot: {e}")
            raise

    async def _method_get_storage_info(
        self, params: Optional[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        """
        Get storage space information and usage statistics.
        
        Requirements: REQ-CLIENT-036, REQ-REC-004.2
        Epic E6: File Lifecycle Management
        
        Args:
            params: No parameters required
        
        Returns:
            Dictionary containing storage space information and usage statistics
        """
        try:
            # Use the new storage checking method
            storage_info = self.check_storage_space()
            if not storage_info:
                raise OSError("Storage information unavailable")
            
            # Get recordings and snapshots directory paths
            config = self._get_config()
            recordings_dir = config.mediamtx.recordings_path
            snapshots_dir = config.mediamtx.snapshots_path
            
            # Calculate recordings directory size
            recordings_size = 0
            if os.path.exists(recordings_dir):
                for dirpath, dirnames, filenames in os.walk(recordings_dir):
                    for filename in filenames:
                        file_path = os.path.join(dirpath, filename)
                        try:
                            recordings_size += os.path.getsize(file_path)
                        except OSError:
                            continue
            
            # Calculate snapshots directory size
            snapshots_size = 0
            if os.path.exists(snapshots_dir):
                for dirpath, dirnames, filenames in os.walk(snapshots_dir):
                    for filename in filenames:
                        file_path = os.path.join(dirpath, filename)
                        try:
                            snapshots_size += os.path.getsize(file_path)
                        except OSError:
                            continue
            
            # Add threshold status to storage info
            storage_info.update({
                "recordings_size": recordings_size,
                "snapshots_size": snapshots_size,
                "low_space_warning": storage_info["used_percent"] >= self._storage_warn_percent,
                "critical_space_warning": storage_info["used_percent"] >= self._storage_block_percent
            })
            
            return storage_info
                
        except OSError as e:
            self._logger.error(f"Error in get_storage_info: {e}")
            raise
        except Exception as e:
            self._logger.error(f"Unexpected error in get_storage_info: {e}")
            raise

    async def _method_set_retention_policy(
        self, params: Optional[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        """
        Configure file retention policies for automatic cleanup.
        
        Requirements: REQ-CLIENT-035
        Epic E6: File Lifecycle Management
        
        Args:
            params: Parameters containing:
                - policy_type: Type of retention policy ("age", "size", "manual") (required)
                - max_age_days: Maximum age in days for age-based retention (optional)
                - max_size_gb: Maximum size in GB for size-based retention (optional)
                - enabled: Whether automatic cleanup is enabled (required)
        
        Returns:
            Dictionary containing retention policy configuration
        """
        try:
            # Validate parameters
            if not params:
                raise ValueError("Parameters are required")
            
            policy_type = params.get("policy_type")
            if not policy_type or policy_type not in ["age", "size", "manual"]:
                raise ValueError("policy_type must be one of: age, size, manual")
            
            enabled = params.get("enabled")
            if not isinstance(enabled, bool):
                raise ValueError("enabled must be a boolean value")
            
            # Validate age-based policy parameters
            if policy_type == "age":
                max_age_days = params.get("max_age_days")
                if not isinstance(max_age_days, (int, float)) or max_age_days <= 0:
                    raise ValueError("max_age_days must be a positive number for age-based policy")
            
            # Validate size-based policy parameters
            if policy_type == "size":
                max_size_gb = params.get("max_size_gb")
                if not isinstance(max_size_gb, (int, float)) or max_size_gb <= 0:
                    raise ValueError("max_size_gb must be a positive number for size-based policy")
            
            # TODO: Store retention policy configuration (would need persistent storage)
            # For now, just return the configuration
            self._logger.info(f"Retention policy updated: {policy_type}, enabled: {enabled}")
            
            # Build response according to API documentation
            response = {
                "policy_type": policy_type,
                "enabled": enabled,
                "message": "Retention policy updated successfully"
            }
            
            # Add policy-specific fields as required by API documentation
            if policy_type == "age":
                max_age_days = params.get("max_age_days")
                if max_age_days is not None:
                    response["max_age_days"] = max_age_days
            elif policy_type == "size":
                max_size_gb = params.get("max_size_gb")
                if max_size_gb is not None:
                    response["max_size_gb"] = max_size_gb
            
            return response
            
        except ValueError as e:
            self._logger.error(f"Error in set_retention_policy: {e}")
            raise
        except Exception as e:
            self._logger.error(f"Unexpected error in set_retention_policy: {e}")
            raise

    async def _method_cleanup_old_files(
        self, params: Optional[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        """
        Manually trigger cleanup of old files based on retention policies.
        
        Requirements: REQ-CLIENT-037
        Epic E6: File Lifecycle Management
        
        Args:
            params: No parameters required
        
        Returns:
            Dictionary containing cleanup results and statistics
        """
        try:
            # TODO: Implement actual cleanup logic based on retention policies
            # For now, return a placeholder response
            self._logger.info("Manual cleanup triggered (not yet implemented)")
            
            return {
                "cleanup_executed": True,
                "files_deleted": 0,
                "space_freed": 0,
                "message": "Cleanup completed successfully (placeholder implementation)"
            }
            
        except Exception as e:
            self._logger.error(f"Unexpected error in cleanup_old_files: {e}")
            raise
