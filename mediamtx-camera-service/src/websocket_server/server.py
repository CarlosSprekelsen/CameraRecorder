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
from collections import defaultdict

import websockets
from websockets.server import ServerProtocol as WebSocketServerProtocol
from websockets.exceptions import ConnectionClosed, WebSocketException

from camera_service.logging_config import set_correlation_id


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
    
    def __init__(self):
        self.request_count = 0
        self.response_times = defaultdict(list)
        self.error_count = 0
        self.active_connections = 0
        self.start_time = time.time()
    
    def record_request(self, method: str, response_time: float):
        """Record a request with its response time."""
        self.request_count += 1
        self.response_times[method].append(response_time)
    
    def record_error(self):
        """Record an error occurrence."""
        self.error_count += 1
    
    def get_metrics(self) -> Dict[str, Any]:
        """Get current performance metrics."""
        uptime = time.time() - self.start_time
        avg_response_times = {}
        methods = {}
        
        for method, times in self.response_times.items():
            if times:
                avg_time = sum(times) / len(times)
                max_time = max(times)
                avg_response_times[method] = avg_time
                
                # Add detailed method metrics for performance framework
                methods[method] = {
                    "count": len(times),
                    "avg_ms": avg_time * 1000,  # Convert to milliseconds
                    "max_ms": max_time * 1000   # Convert to milliseconds
                }
        
        return {
            "uptime": uptime,
            "request_count": self.request_count,
            "error_count": self.error_count,
            "active_connections": self.active_connections,
            "avg_response_times": avg_response_times,
            "requests_per_second": self.request_count / uptime if uptime > 0 else 0,
            "methods": methods  # Add the missing methods field
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


class WebSocketJsonRpcServer:
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
        mediamtx_controller=None,
        camera_monitor=None,
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

    def set_service_manager(self, service_manager) -> None:
        """
        Set the service manager for lifecycle coordination.

        Args:
            service_manager: Service manager instance
        """
        self._service_manager = service_manager
    
    def set_security_middleware(self, security_middleware) -> None:
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
            bind_host = self._host
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
            if self._security_middleware:
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
                                "error": {"code": -32603, "message": "Internal error"},
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
            if self._security_middleware:
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
                        "error": {"code": -32601, "message": "Method not found"},
                        "id": request_id,
                    }
                )

            # Allow authenticate method without prior enforcement and handle here to bind to correct client
            if method_name == "authenticate":
                if not self._security_middleware:
                    return json.dumps({
                        "jsonrpc": "2.0",
                        "result": {"authenticated": False, "error": "Security not configured"},
                        "id": request_id,
                    })
                token = None
                auth_type = "auto"
                if params and isinstance(params, dict):
                    token = params.get("token") or params.get("auth_token")
                    auth_type = params.get("auth_type", "auto")
                if not token:
                    return json.dumps({
                        "jsonrpc": "2.0",
                        "result": {"authenticated": False, "error": "Missing token"},
                        "id": request_id,
                    })
                auth_result = await self._security_middleware.authenticate_connection(
                    client.client_id, token, auth_type
                )
                if auth_result.authenticated:
                    client.authenticated = True
                    client.auth_result = auth_result
                    client.user_id = auth_result.user_id
                    client.role = auth_result.role
                    client.auth_method = auth_result.auth_method
                    return json.dumps({
                        "jsonrpc": "2.0",
                        "result": {"authenticated": True, "role": auth_result.role, "auth_method": auth_result.auth_method},
                        "id": request_id,
                    })
                return json.dumps({
                    "jsonrpc": "2.0",
                    "result": {"authenticated": False, "error": auth_result.error_message},
                    "id": request_id,
                })

            # Security enforcement and rate limiting (F3.2.5/F3.2.6/I1.7/N3.x)
            if self._security_middleware:
                protected_methods = {"take_snapshot", "start_recording", "stop_recording"}

                # Rate limiting applies to all methods
                if not self._security_middleware.check_rate_limit(client.client_id):
                    return json.dumps(
                        {
                            "jsonrpc": "2.0",
                            "error": {"code": -32002, "message": "Rate limit exceeded"},
                            "id": request_id,
                        }
                    )

                if method_name in protected_methods:
                    # Require authentication for protected methods
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
                                            "code": -32001,
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
                                        "code": -32001,
                                        "message": "Authentication required - call authenticate or provide auth_token",
                                    },
                                    "id": request_id,
                                }
                            )

                    # Authorization: operator role required
                    if not self._security_middleware.has_permission(client.client_id, "operator"):
                        return json.dumps(
                            {
                                "jsonrpc": "2.0",
                                "error": {
                                    "code": -32003,
                                    "message": "Insufficient permissions - operator role required",
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
                                                "code": -32001,
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
                        error_code = -1003
                        error_message = "MediaMTX operation failed"
                    elif isinstance(e, AuthenticationError):
                        error_code = -1001
                        error_message = "Authentication failed"
                    elif isinstance(e, PermissionError):
                        error_code = -1002
                        error_message = "Permission denied"
                    elif isinstance(e, StreamError):
                        error_code = -1004
                        error_message = "Stream operation failed"
                    elif isinstance(e, ValueError):
                        error_code = -32602
                        error_message = "Invalid params"
                    elif isinstance(e, KeyError):
                        error_code = -32602
                        error_message = "Missing required parameter"
                    elif isinstance(e, TypeError):
                        error_code = -32602
                        error_message = "Invalid parameter type"
                    else:
                        error_code = -32603
                        error_message = "Internal error"
                    
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
                    "error": {"code": -32603, "message": "Internal error"},
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
        else:
            clients_to_notify = list(self._clients.values())

        # Send notification to each target client (robust open/closed checks)
        failed_clients = []
        for client in clients_to_notify:
            try:
                is_open = getattr(client.websocket, "open", None)
                is_closed = getattr(client.websocket, "closed", None)
                if (is_open is True) or (is_closed is False) or (is_open is None and is_closed is None):
                    await client.websocket.send(notification_json)
                else:
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
                    "status": "COMPLETED",
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
            if getattr(client.websocket, "closed", False) or not getattr(client.websocket, "open", True):
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

            is_open = getattr(client.websocket, "open", None)
            is_closed = getattr(client.websocket, "closed", None)
            if (is_open is True) or (is_closed is False):
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
        # Alias for compatibility
        self.register_method(
            "get_cameras", self._method_get_camera_list, version="1.0"
        )
        self.register_method(
            "get_camera_status", self._method_get_camera_status, version="1.0"
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
        Params: { token: string, auth_type: "jwt"|"api_key" (optional) }
        """
        if not self._security_middleware:
            return {"authenticated": False, "error": "Security not configured"}
        token = None
        auth_type = "auto"
        if params and isinstance(params, dict):
            token = params.get("token") or params.get("auth_token")
            auth_type = params.get("auth_type", "auto")
        if not token:
            return {"authenticated": False, "error": "Missing token"}

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
            client.authenticated = True
            client.auth_result = auth_result
            client.user_id = auth_result.user_id
            client.role = auth_result.role
            client.auth_method = auth_result.auth_method
            return {
                "authenticated": True,
                "role": auth_result.role,
                "auth_method": auth_result.auth_method,
            }
        return {"authenticated": False, "error": auth_result.error_message}

    async def _method_get_metrics(
        self, params: Optional[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        """
        Get performance metrics for the WebSocket server.

        Args:
            params: Method parameters (unused)

        Returns:
            Dict containing performance metrics
        """
        return self.get_performance_metrics()

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
                streams = {}

                # Get stream URLs from concurrent request results
                if camera_device.status == "CONNECTED" and device_path in stream_status_tasks:
                    task = stream_status_tasks[device_path]
                    if task.done() and not task.exception():
                        try:
                            stream_status = task.result()
                            if stream_status and stream_status.get("status") == "active":
                                streams = {
                                    "rtsp": f"rtsp://localhost:8554/{stream_name}",
                                    "webrtc": f"http://localhost:8889/{stream_name}/webrtc",
                                    "hls": f"http://localhost:8888/{stream_name}",
                                }
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

    async def _get_stream_status_safe(self, mediamtx_controller, stream_name: str) -> Optional[Dict[str, Any]]:
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
        Get detailed status for a specific camera device with aggregated real data.

        Combines data from camera discovery monitor (with provisional/confirmed capability logic),
        MediaMTX controller (stream status and metrics), and provides graceful fallbacks.

        Args:
            params: Method parameters containing:
                - device (str): Camera device path

        Returns:
            Dict containing comprehensive camera status per API specification:
                - device, status, name, resolution, fps, streams, metrics, capabilities
        """
        if not params or "device" not in params:
            raise ValueError("device parameter is required")

        device_path = params["device"]

        # Initialize response with architecture defaults
        camera_status = {
            "device": device_path,
            "status": "DISCONNECTED",
            "name": f"Camera {device_path.split('video')[-1] if 'video' in device_path else 'unknown'}",
            "resolution": "1920x1080",  # Architecture default
            "fps": 30,  # Architecture default
            "streams": {
                "rtsp": f"rtsp://127.0.0.1:8554/{device_path.split('/')[-1]}",
                "webrtc": f"http://127.0.0.1:8889/{device_path.split('/')[-1]}",
                "hls": f"http://127.0.0.1:8888/{device_path.split('/')[-1]}"
            },
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
                                    camera_status["capabilities"]["formats"] = (
                                        capability_metadata["formats"]
                                    )
                                if capability_metadata.get("all_resolutions"):
                                    camera_status["capabilities"]["resolutions"] = (
                                        capability_metadata["all_resolutions"]
                                    )

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
                else:
                    # Camera not found - return error
                    raise CameraNotFoundError(f"Camera device {device_path} not found")
            else:
                # Without a camera monitor, we cannot validate the device. Fail closed for correctness.
                raise CameraNotFoundError(f"Camera device {device_path} not found")

            # Get stream info and metrics from MediaMTX controller
            if mediamtx_controller and camera_status["status"] == "CONNECTED":
                try:
                    stream_name = self._get_stream_name_from_device_path(device_path)
                    stream_status = await mediamtx_controller.get_stream_status(
                        stream_name
                    )

                    if stream_status.get("status") == "active":
                        # Update stream URLs
                        camera_status["streams"] = {
                            "rtsp": f"rtsp://localhost:8554/{stream_name}",
                            "webrtc": f"webrtc://localhost:8002/{stream_name}",
                            "hls": f"http://localhost:8002/hls/{stream_name}.m3u8",
                        }

                        # Update metrics from MediaMTX
                        camera_status["metrics"] = {
                            "bytes_sent": stream_status.get("bytes_sent", 0),
                            "readers": stream_status.get("readers", 0),
                            "uptime": int(time.time()),  # Current uptime proxy
                        }

                except Exception as e:
                    self._logger.debug(
                        f"Could not get MediaMTX status for {device_path}: {e}"
                    )

            return camera_status

        except CameraNotFoundError:
            self._logger.error(f"Camera device {device_path} not found")
            raise CameraNotFoundError(f"Camera device {device_path} not found")
        except Exception as e:
            self._logger.error(f"Error getting camera status for {device_path}: {e}")
            # Return JSON-RPC error response
            raise ValueError(f"Camera device {device_path} not found") from e

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
                "error": "MediaMTX controller not available",
            }

        try:
            stream_name = self._get_stream_name_from_device_path(device_path)

            snapshot_result = await mediamtx_controller.take_snapshot(
                stream_name=stream_name,
                format=format_type,
                quality=quality,
                filename=custom_filename,
            )

            return {
                "device": device_path,
                "filename": snapshot_result.get("filename"),
                "status": "SUCCESS",
                "timestamp": snapshot_result.get(
                    "timestamp", time.strftime("%Y-%m-%dT%H:%M:%SZ")
                ),
                "file_size": snapshot_result.get("file_size", 0),
                "format": format_type,
                "quality": quality,
            }

        except Exception as e:
            self._logger.error(f"Error taking snapshot for {device_path}: {e}")
            raise MediaMTXError(f"MediaMTX operation failed: {e}") from e

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
                "start_time": time.strftime("%Y-%m-%dT%H:%M:%SZ"),
                "duration": duration,
                "format": format_type,
                "error": "MediaMTX controller not available",
            }

        try:
            stream_name = self._get_stream_name_from_device_path(device_path)
            session_id = str(uuid.uuid4())

            recording_result = await mediamtx_controller.start_recording(
                stream_name=stream_name, duration=effective_duration, format=format_type
            )

            response = {
                "device": device_path,
                "session_id": session_id,
                "filename": recording_result.get("filename"),
                "status": "STARTED",
                "start_time": recording_result.get(
                    "start_time", time.strftime("%Y-%m-%dT%H:%M:%SZ")
                ),
                "duration": effective_duration,
                "format": format_type,
            }

            # Schedule auto-stop if timed recording requested
            if effective_duration and effective_duration > 0:
                async def _auto_stop():
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

            return response

        except Exception as e:
            self._logger.error(f"Error starting recording for {device_path}: {e}")
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
                "stop_time": time.strftime("%Y-%m-%dT%H:%M:%SZ"),
                "duration": None,
                "file_size": 0,
                "error": "MediaMTX controller not available",
            }

        try:
            stream_name = self._get_stream_name_from_device_path(device_path)

            recording_result = await mediamtx_controller.stop_recording(stream_name=stream_name)

            return {
                "device": device_path,
                "session_id": recording_result.get("session_id"),
                "filename": recording_result.get("filename"),
                "status": "STOPPED",
                "stop_time": recording_result.get(
                    "stop_time", time.strftime("%Y-%m-%dT%H:%M:%SZ")
                ),
                "duration": recording_result.get("duration"),
                "file_size": recording_result.get("file_size", 0),
            }

        except Exception as e:
            self._logger.error(f"Error stopping recording for {device_path}: {e}")
            raise MediaMTXError(f"MediaMTX operation failed: {e}") from e

    async def _method_get_status(
        self, params: Optional[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        """
        Get general server status information.

        Returns:
            Dict containing server status information
        """
        import time
        
        status = {
            "server": {
                "status": "running",
                "uptime": time.time() - getattr(self, '_start_time', time.time()),
                "version": "1.0.0",
                "connections": len(self._clients)
            },
            "mediamtx": {
                "status": "unknown",
                "connected": False
            }
        }
        
        # Check MediaMTX status if available
        if self._service_manager and hasattr(self._service_manager, '_mediamtx_controller'):
            mediamtx_controller = self._service_manager._mediamtx_controller
            if mediamtx_controller:
                try:
                    health = await mediamtx_controller.health_check()
                    status["mediamtx"]["status"] = health.get("status", "unknown")
                    status["mediamtx"]["connected"] = health.get("status") == "healthy"
                except:
                    pass
        
        return status

    async def _method_get_server_info(
        self, params: Optional[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        """
        Get server information and capabilities.

        Returns:
            Dict containing server information
        """
        return {
            "name": "MediaMTX Camera Service",
            "version": "1.0.0",
            "api_version": "1.0",
            "supported_methods": list(self._methods.keys()),
            "capabilities": [
                "camera_monitoring",
                "video_recording",
                "snapshot_capture",
                "rtsp_streaming",
                "websocket_notifications"
            ]
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
        Broadcast camera_status_update notification with strict API compliance.

        Filters notification parameters to only include API-specified fields:
        device, status, name, resolution, fps, streams (per docs/api/json-rpc-methods.md)

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
            await self.broadcast_notification(
                method="camera_status_update", params=filtered_params
            )

            self._logger.info(
                f"Broadcasted camera status update for device: {params.get('device')}"
            )

        except Exception as e:
            self._logger.error(f"Failed to broadcast camera status update: {e}")

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
