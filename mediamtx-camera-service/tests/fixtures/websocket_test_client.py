"""
WebSocket test client fixture for MediaMTX Camera Service.

Test Categories: Test Infrastructure
"""

import asyncio
import json
import logging
from typing import Dict, Any, Optional, List
from dataclasses import dataclass

import pytest
import pytest_asyncio
import websockets
from websockets.client import WebSocketClientProtocol
from websockets.exceptions import ConnectionClosed
from contextlib import asynccontextmanager

logger = logging.getLogger(__name__)


@dataclass
class WebSocketMessage:
    """WebSocket message structure."""
    method: str
    params: Dict[str, Any]
    id: Optional[int] = None
    jsonrpc: str = "2.0"


@dataclass
class WebSocketResponse:
    """WebSocket response structure."""
    result: Optional[Dict[str, Any]] = None
    error: Optional[Dict[str, Any]] = None
    id: Optional[int] = None
    jsonrpc: str = "2.0"


class WebSocketTestClient:
    """Real WebSocket client for testing."""
    
    def __init__(self, server_url: str = "ws://localhost:8002/ws"):
        self.server_url = server_url
        self.websocket: Optional[WebSocketClientProtocol] = None
        self.connected = False
        self.message_id_counter = 1
        self._received_messages: List[WebSocketResponse] = []
        self._notification_queue: asyncio.Queue = asyncio.Queue()
        self._listener_task: Optional[asyncio.Task] = None
    
    async def connect(self) -> None:
        """Connect to WebSocket server."""
        try:
            self.websocket = await websockets.connect(self.server_url)
            self.connected = True
            logger.info(f"Connected to WebSocket server: {self.server_url}")
            
            # Start listening for notifications
            self._listener_task = asyncio.create_task(self._listen_for_notifications())
            # Give the task a moment to start
            await asyncio.sleep(0.01)
            
        except Exception as e:
            logger.error(f"Failed to connect to WebSocket server: {e}")
            raise
    
    async def disconnect(self) -> None:
        """Disconnect from server."""
        self.connected = False
        
        # Cancel listener task if running
        if self._listener_task and not self._listener_task.done():
            self._listener_task.cancel()
            try:
                await self._listener_task
            except asyncio.CancelledError:
                pass
        
        if self.websocket:
            await self.websocket.close()
            logger.info("Disconnected from WebSocket server")
    
    async def send_request(self, method: str, params: Dict[str, Any] = None) -> WebSocketResponse:
        """Send JSON-RPC request."""
        if not self.connected:
            raise RuntimeError("WebSocket client not connected")
        
        message = WebSocketMessage(
            method=method,
            params=params or {},
            id=self.message_id_counter,
            jsonrpc="2.0"
        )
        
        # Send request
        await self.websocket.send(json.dumps({
            "jsonrpc": message.jsonrpc,
            "id": message.id,
            "method": message.method,
            "params": message.params
        }))
        
        # Wait for response
        response_data = await self.websocket.recv()
        response = json.loads(response_data)
        
        # Store response
        ws_response = WebSocketResponse(
            result=response.get("result"),
            error=response.get("error"),
            id=response.get("id"),
            jsonrpc=response.get("jsonrpc", "2.0")
        )
        self._received_messages.append(ws_response)
        
        self.message_id_counter += 1
        return ws_response
    
    async def send_notification(self, method: str, params: Dict[str, Any] = None) -> None:
        """Send JSON-RPC notification (no response expected)."""
        if not self.connected:
            raise RuntimeError("WebSocket client not connected")
        
        message = WebSocketMessage(
            method=method,
            params=params or {},
            jsonrpc="2.0"
        )
        
        # Send notification
        await self.websocket.send(json.dumps({
            "jsonrpc": message.jsonrpc,
            "method": message.method,
            "params": message.params
        }))
    
    async def wait_for_notification(self, notification_type: str, timeout: float = 5.0) -> WebSocketResponse:
        """Wait for specific notification."""
        start_time = asyncio.get_event_loop().time()
        
        while asyncio.get_event_loop().time() - start_time < timeout:
            try:
                # Check if notification is already in queue
                notification = await asyncio.wait_for(
                    self._notification_queue.get(), 
                    timeout=0.1
                )
                
                if notification.get("method") == notification_type:
                    return WebSocketResponse(
                        result=notification.get("params"),
                        jsonrpc=notification.get("jsonrpc", "2.0")
                    )
                else:
                    # Put back in queue if not the expected type
                    await self._notification_queue.put(notification)
                    
            except asyncio.TimeoutError:
                continue
        
        raise TimeoutError(f"Notification {notification_type} not received within {timeout} seconds")
    
    async def _listen_for_notifications(self) -> None:
        """Listen for incoming notifications."""
        try:
            print(f"DEBUG CLIENT: Starting notification listener for {self.server_url}")
            while self.connected:
                try:
                    message_data = await self.websocket.recv()
                    print(f"DEBUG CLIENT: Received message: {message_data[:100]}...")
                    message = json.loads(message_data)
                    
                    # Check if it's a notification (no id field)
                    if "id" not in message:
                        print(f"DEBUG CLIENT: Processing notification: {message.get('method', 'unknown')}")
                        await self._notification_queue.put(message)
                    else:
                        print(f"DEBUG CLIENT: Received request/response with id: {message.get('id')}")
                        
                except websockets.exceptions.ConnectionClosed:
                    print(f"DEBUG CLIENT: Connection closed")
                    break
                except Exception as e:
                    print(f"DEBUG CLIENT: Error processing notification: {e}")
                    logger.warning(f"Error processing notification: {e}")
                    
        except Exception as e:
            print(f"DEBUG CLIENT: Error in notification listener: {e}")
            logger.error(f"Error in notification listener: {e}")
    
    def get_received_messages(self) -> List[WebSocketResponse]:
        """Get all received messages including notifications."""
        messages = self._received_messages.copy()
        
        # Also include notifications from the queue
        while not self._notification_queue.empty():
            try:
                notification = self._notification_queue.get_nowait()
                messages.append(WebSocketResponse(
                    result=notification.get("params"),
                    jsonrpc=notification.get("jsonrpc", "2.0")
                ))
            except asyncio.QueueEmpty:
                break
                
        return messages
    
    def clear_received_messages(self) -> None:
        """Clear received messages."""
        self._received_messages.clear()
    
    async def ping(self) -> WebSocketResponse:
        """Send ping request."""
        return await self.send_request("ping", {})
    
    async def get_camera_list(self) -> WebSocketResponse:
        """Get camera list."""
        return await self.send_request("get_camera_list", {})
    
    async def get_camera_status(self, device: str) -> WebSocketResponse:
        """Get camera status."""
        return await self.send_request("get_camera_status", {"device": device})
    
    async def start_stream(self, device: str, resolution: str = "1920x1080", frame_rate: int = 30) -> WebSocketResponse:
        """Start camera stream."""
        return await self.send_request("start_stream", {
            "device": device,
            "resolution": resolution,
            "frame_rate": frame_rate
        })
    
    async def stop_stream(self, device: str) -> WebSocketResponse:
        """Stop camera stream."""
        return await self.send_request("stop_stream", {"device": device})
    
    async def take_snapshot(self, device: str, format: str = "jpeg", quality: int = 85) -> WebSocketResponse:
        """Take camera snapshot."""
        return await self.send_request("take_snapshot", {
            "device": device,
            "format": format,
            "quality": quality
        })


# Pytest fixtures for easy integration
@pytest_asyncio.fixture
async def websocket_client():
    """Real WebSocket test client."""
    client = WebSocketTestClient()
    await client.connect()
    yield client
    await client.disconnect()


@pytest_asyncio.fixture
async def websocket_client_connected(websocket_client):
    """WebSocket client that's already connected."""
    # Wait a moment for connection to stabilize
    await asyncio.sleep(0.1)
    return websocket_client


# Test utilities
async def create_websocket_client(server_url: str = "ws://localhost:8002/ws") -> WebSocketTestClient:
    """Create and connect a WebSocket test client."""
    client = WebSocketTestClient(server_url)
    await client.connect()
    return client


async def send_websocket_request_and_validate(
    client: WebSocketTestClient,
    method: str,
    params: Dict[str, Any] = None,
    expected_result: bool = True
) -> WebSocketResponse:
    """Send WebSocket request and validate response."""
    response = await client.send_request(method, params or {})
    
    if expected_result:
        assert response.result is not None, f"Expected result for {method}, got error: {response.error}"
    else:
        assert response.error is not None, f"Expected error for {method}, got result: {response.result}"
    
    return response


async def wait_for_camera_notification(
    client: WebSocketTestClient,
    expected_device: str,
    expected_status: str,
    timeout: float = 5.0
) -> WebSocketResponse:
    """Wait for camera status notification."""
    notification = await client.wait_for_notification("camera_status_update", timeout)
    
    assert notification.result is not None, "Expected notification result"
    assert notification.result.get("device") == expected_device, f"Expected device {expected_device}"
    assert notification.result.get("status") == expected_status, f"Expected status {expected_status}"
    
    return notification


async def verify_websocket_connection(server_url: str = "ws://localhost:8002/ws", timeout: float = 5.0) -> bool:
    """Verify that WebSocket server is accessible."""
    try:
        async with websockets.connect(server_url) as websocket:
            # Send ping to verify connection
            await websocket.send(json.dumps({
                "jsonrpc": "2.0",
                "id": 1,
                "method": "ping",
                "params": {}
            }))
            
            response = await websocket.recv()
            response_data = json.loads(response)
            
            return "result" in response_data
            
    except Exception:
        return False


# Context manager for WebSocket testing
@asynccontextmanager
async def websocket_test_context(server_url: str = "ws://localhost:8002/ws"):
    """Context manager for WebSocket testing."""
    client = WebSocketTestClient(server_url)
    try:
        await client.connect()
        yield client
    finally:
        await client.disconnect()
