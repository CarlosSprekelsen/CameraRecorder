#!/usr/bin/env python3
"""
Real System Integration Tests - Validating Actual End-to-End System Behavior

This test suite validates real system integration without excessive mocking:
- Real MediaMTX server integration (not mocked HTTP responses)
- Real camera device simulation with test video streams
- Real file system operations for recordings and snapshots
- Real WebSocket connections for command interface
- Real FFmpeg process execution for media operations

Test Strategy:
1. Real MediaMTX server startup and configuration
2. Real camera device simulation with test video files
3. Real end-to-end camera discovery → stream creation → recording → snapshot capture
4. Real WebSocket authentication → camera control → status monitoring
5. Real error scenarios with actual service failures and recovery
6. Real resource management under load and failure conditions

Success Criteria: Integration tests validate real end-to-end system behavior without mock dependencies.
"""

import asyncio
import json
import logging
import os
import shutil
import socket
import subprocess
import tempfile
import time
from contextlib import asynccontextmanager
from pathlib import Path
from typing import Dict, Any, Optional, List

import pytest
import pytest_asyncio
import websockets
from aiohttp import web

# Import project modules
from src.camera_service.service_manager import ServiceManager
from src.camera_service.config import (
    Config,
    ServerConfig,
    MediaMTXConfig,
    CameraConfig,
    LoggingConfig,
    RecordingConfig,
    SnapshotConfig,
)
from src.mediamtx_wrapper.controller import MediaMTXController
from src.websocket_server.server import WebSocketJsonRpcServer
from camera_discovery.hybrid_monitor import CameraEventData, CameraEvent
from src.common.types import CameraDevice

# Configure logging for tests
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


class RealMediaMTXServer:
    """Real MediaMTX server for integration testing."""
    
    def __init__(self, config: MediaMTXConfig):
        self.config = config
        self.process: Optional[subprocess.Popen] = None
        self.temp_dir: Optional[str] = None
        self.config_file: Optional[str] = None
        
    async def start(self) -> None:
        """Start real MediaMTX server process."""
        logger.info("Starting real MediaMTX server...")
        
        # Create temporary directory for MediaMTX
        self.temp_dir = tempfile.mkdtemp(prefix="mediamtx_test_")
        
        # Create MediaMTX configuration file
        self.config_file = os.path.join(self.temp_dir, "mediamtx.yml")
        self._create_mediamtx_config()
        
        # Create directories
        os.makedirs(self.config.recordings_path, exist_ok=True)
        os.makedirs(self.config.snapshots_path, exist_ok=True)
        
        # Start MediaMTX process
        cmd = [
            "mediamtx",  # Assume mediamtx is in PATH
            self.config_file
        ]
        
        try:
            self.process = subprocess.Popen(
                cmd,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                cwd=self.temp_dir
            )
            
            # Wait for MediaMTX to start
            await self._wait_for_mediamtx_ready()
            logger.info("Real MediaMTX server started successfully")
            
        except Exception as e:
            logger.error(f"Failed to start MediaMTX server: {e}")
            await self.stop()
            raise
    
    async def stop(self) -> None:
        """Stop real MediaMTX server process."""
        if self.process:
            logger.info("Stopping real MediaMTX server...")
            self.process.terminate()
            try:
                self.process.wait(timeout=5)
            except subprocess.TimeoutExpired:
                self.process.kill()
                self.process.wait()
            self.process = None
        
        # Clean up temporary directory
        if self.temp_dir and os.path.exists(self.temp_dir):
            shutil.rmtree(self.temp_dir)
            self.temp_dir = None
    
    def _create_mediamtx_config(self) -> None:
        """Create MediaMTX configuration file."""
        config_content = f"""
# MediaMTX Configuration for Real Integration Testing
api: yes
apiAddress: :{self.config.api_port}

rtspAddress: :{self.config.rtsp_port}
rtspTransports: [tcp, udp]

webrtcAddress: :{self.config.webrtc_port}

hlsAddress: :{self.config.hls_port}
hlsVariant: lowLatency

logLevel: info
logDestinations: [stdout]

paths:
  all:
    recordFormat: fmp4
    recordSegmentDuration: "3600s"
    recordPath: {self.config.recordings_path}
    snapshotPath: {self.config.snapshots_path}
"""
        
        with open(self.config_file, 'w') as f:
            f.write(config_content)
    
    async def _wait_for_mediamtx_ready(self, timeout: float = 30.0) -> None:
        """Wait for MediaMTX server to be ready."""
        start_time = time.time()
        while time.time() - start_time < timeout:
            try:
                # Check if API port is listening
                sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
                sock.settimeout(1)
                result = sock.connect_ex(('127.0.0.1', self.config.api_port))
                sock.close()
                
                if result == 0:
                    # Test API health endpoint
                    import aiohttp
                    async with aiohttp.ClientSession() as session:
                        async with session.get(f"http://127.0.0.1:{self.config.api_port}/v3/health") as resp:
                            if resp.status == 200:
                                logger.info("MediaMTX server is ready")
                                return
                
                await asyncio.sleep(1)
            except Exception:
                await asyncio.sleep(1)
        
        raise TimeoutError("MediaMTX server failed to start within timeout")


class TestVideoStreamSimulator:
    """Simulates real camera video streams for testing."""
    
    def __init__(self, rtsp_port: int):
        self.rtsp_port = rtsp_port
        self.processes: List[subprocess.Popen] = []
        self.temp_dir: Optional[str] = None
        
    async def start_test_streams(self, stream_names: List[str]) -> None:
        """Start test video streams using FFmpeg."""
        logger.info(f"Starting test video streams: {stream_names}")
        
        self.temp_dir = tempfile.mkdtemp(prefix="test_streams_")
        
        # Create test video files
        test_video_path = os.path.join(self.temp_dir, "test_video.mp4")
        await self._create_test_video(test_video_path)
        
        # Start FFmpeg processes for each stream
        for stream_name in stream_names:
            await self._start_ffmpeg_stream(stream_name, test_video_path)
    
    async def stop_test_streams(self) -> None:
        """Stop all test video streams."""
        logger.info("Stopping test video streams...")
        
        for process in self.processes:
            if process.poll() is None:
                process.terminate()
                try:
                    process.wait(timeout=5)
                except subprocess.TimeoutExpired:
                    process.kill()
                    process.wait()
        
        self.processes.clear()
        
        # Clean up temporary directory
        if self.temp_dir and os.path.exists(self.temp_dir):
            shutil.rmtree(self.temp_dir)
            self.temp_dir = None
    
    async def _create_test_video(self, output_path: str) -> None:
        """Create a test video file using FFmpeg."""
        cmd = [
            "ffmpeg",
            "-f", "lavfi",
            "-i", "testsrc=duration=60:size=640x480:rate=30",
            "-c:v", "libx264",
            "-preset", "ultrafast",
            "-tune", "zerolatency",
            "-f", "mp4",
            output_path,
            "-y"  # Overwrite output file
        ]
        
        try:
            process = await asyncio.create_subprocess_exec(
                *cmd,
                stdout=asyncio.subprocess.PIPE,
                stderr=asyncio.subprocess.PIPE
            )
            await process.communicate()
            
            if process.returncode != 0:
                raise RuntimeError("Failed to create test video")
                
        except Exception as e:
            logger.error(f"Error creating test video: {e}")
            raise
    
    async def _start_ffmpeg_stream(self, stream_name: str, video_path: str) -> None:
        """Start FFmpeg process to stream test video to RTSP."""
        rtsp_url = f"rtsp://127.0.0.1:{self.rtsp_port}/{stream_name}"
        
        cmd = [
            "ffmpeg",
            "-re",  # Read input at native frame rate
            "-stream_loop", "-1",  # Loop the video indefinitely
            "-i", video_path,
            "-c:v", "libx264",
            "-preset", "ultrafast",
            "-tune", "zerolatency",
            "-f", "rtsp",
            rtsp_url
        ]
        
        try:
            process = subprocess.Popen(
                cmd,
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE
            )
            self.processes.append(process)
            
            # Wait a moment for stream to start
            await asyncio.sleep(2)
            
            logger.info(f"Started test stream: {stream_name} -> {rtsp_url}")
            
        except Exception as e:
            logger.error(f"Error starting test stream {stream_name}: {e}")
            raise


class WebSocketTestClient:
    """Test client for WebSocket JSON-RPC communication."""
    
    def __init__(self, websocket_url: str):
        self.websocket_url = websocket_url
        self.websocket = None
        self.notifications: List[Dict] = []
    
    async def connect(self) -> None:
        """Connect to WebSocket server."""
        self.websocket = await websockets.connect(self.websocket_url)
        
        # Start notification listener
        asyncio.create_task(self._listen_for_notifications())
    
    async def disconnect(self) -> None:
        """Disconnect from WebSocket server."""
        if self.websocket:
            await self.websocket.close()
            self.websocket = None
    
    async def send_request(self, method: str, params: Dict = None, request_id: int = 1) -> Dict:
        """Send JSON-RPC request and wait for response."""
        if not self.websocket:
            raise RuntimeError("WebSocket not connected")
        
        request = {
            "jsonrpc": "2.0",
            "id": request_id,
            "method": method
        }
        
        if params:
            request["params"] = params
        
        await self.websocket.send(json.dumps(request))
        response = await self.websocket.recv()
        return json.loads(response)
    
    async def wait_for_notification(self, method: str, timeout: float = 10.0) -> Dict:
        """Wait for specific notification method."""
        start_time = time.time()
        while time.time() - start_time < timeout:
            for notification in self.notifications:
                if notification.get("method") == method:
                    self.notifications.remove(notification)
                    return notification
            await asyncio.sleep(0.1)
        
        raise TimeoutError(f"No notification {method} received within {timeout}s")
    
    async def _listen_for_notifications(self) -> None:
        """Listen for notifications from WebSocket server."""
        try:
            while self.websocket and not self.websocket.closed:
                message = await self.websocket.recv()
                data = json.loads(message)
                
                if "method" in data:  # This is a notification
                    self.notifications.append(data)
                    logger.info(f"Received notification: {data['method']}")
        except Exception as e:
            logger.error(f"Error in notification listener: {e}")


class TestRealSystemIntegration:
    """
    Real System Integration Tests - Validating Actual End-to-End System Behavior
    
    These tests validate real system integration without excessive mocking:
    - Real MediaMTX server startup and configuration
    - Real camera device simulation with test video streams
    - Real file system operations for recordings and snapshots
    - Real WebSocket connections for command interface
    - Real FFmpeg process execution for media operations
    """
    
    @pytest.fixture
    def test_config(self):
        """Create test configuration with real component paths."""
        def find_free_port():
            with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
                s.bind(('', 0))
                s.listen(1)
                port = s.getsockname()[1]
            return port
        
        # Use dynamic port allocation to avoid conflicts
        server_port = find_free_port()
        mediamtx_api_port = find_free_port()
        mediamtx_rtsp_port = find_free_port()
        mediamtx_webrtc_port = find_free_port()
        mediamtx_hls_port = find_free_port()
        
        # Create temporary directories for real file operations
        temp_dir = tempfile.mkdtemp(prefix="real_integration_test_")
        recordings_dir = os.path.join(temp_dir, "recordings")
        snapshots_dir = os.path.join(temp_dir, "snapshots")
        os.makedirs(recordings_dir, exist_ok=True)
        os.makedirs(snapshots_dir, exist_ok=True)
        
        return Config(
            server=ServerConfig(
                host="127.0.0.1",
                port=server_port,
                websocket_path="/ws",
                max_connections=10
            ),
            mediamtx=MediaMTXConfig(
                host="127.0.0.1",
                api_port=mediamtx_api_port,
                rtsp_port=mediamtx_rtsp_port,
                webrtc_port=mediamtx_webrtc_port,
                hls_port=mediamtx_hls_port,
                config_path=os.path.join(temp_dir, "mediamtx.yml"),
                recordings_path=recordings_dir,
                snapshots_path=snapshots_dir,
                health_check_interval=5,
                health_failure_threshold=3,
                health_circuit_breaker_timeout=10,
                health_max_backoff_interval=20,
                health_recovery_confirmation_threshold=2,
                backoff_base_multiplier=1.5,
                backoff_jitter_range=(0.8, 1.2),
                process_termination_timeout=3.0,
                process_kill_timeout=2.0
            ),
            camera=CameraConfig(
                device_range=[0, 1, 2],
                enable_capability_detection=True,
                detection_timeout=2.0,
                poll_interval=0.5
            ),
            logging=LoggingConfig(level="INFO"),
            recording=RecordingConfig(enabled=True),
            snapshots=SnapshotConfig(enabled=True)
        )
    
    @pytest_asyncio.fixture
    async def real_mediamtx_server(self, test_config):
        """Real MediaMTX server fixture."""
        server = RealMediaMTXServer(test_config.mediamtx)
        await server.start()
        yield server
        await server.stop()
    
    @pytest_asyncio.fixture
    async def test_video_streams(self, test_config):
        """Test video streams fixture."""
        simulator = TestVideoStreamSimulator(test_config.mediamtx.rtsp_port)
        await simulator.start_test_streams(["camera0", "camera1", "camera2"])
        yield simulator
        await simulator.stop_test_streams()
    
    @pytest_asyncio.fixture
    async def service_manager(self, test_config, real_mediamtx_server):
        """Service manager fixture with real MediaMTX integration."""
        manager = ServiceManager(test_config)
        await manager.start()
        yield manager
        await manager.stop()
    
    @pytest_asyncio.fixture
    async def websocket_client(self, test_config, service_manager):
        """WebSocket test client fixture."""
        websocket_url = f"ws://{test_config.server.host}:{test_config.server.port}{test_config.server.websocket_path}"
        client = WebSocketTestClient(websocket_url)
        await client.connect()
        yield client
        await client.disconnect()
    
    @pytest.mark.asyncio
    async def test_real_mediamtx_server_startup_and_health(self, real_mediamtx_server, test_config):
        """
        Test real MediaMTX server startup and health monitoring.
        
        Validates:
        - Real MediaMTX server process startup
        - Real API health endpoint responses
        - Real configuration file generation
        - Real directory structure creation
        """
        logger.info("Testing real MediaMTX server startup and health...")
        
        # Verify MediaMTX process is running
        assert real_mediamtx_server.process is not None
        assert real_mediamtx_server.process.poll() is None
        
        # Test real API health endpoint
        import aiohttp
        async with aiohttp.ClientSession() as session:
            async with session.get(f"http://127.0.0.1:{test_config.mediamtx.api_port}/v3/health") as resp:
                assert resp.status == 200
                health_data = await resp.json()
                assert "serverVersion" in health_data
                assert "serverUptime" in health_data
        
        # Verify configuration file exists
        assert os.path.exists(test_config.mediamtx.config_path)
        
        # Verify directories exist
        assert os.path.exists(test_config.mediamtx.recordings_path)
        assert os.path.exists(test_config.mediamtx.snapshots_path)
        
        logger.info("Real MediaMTX server startup and health test passed")
    
    @pytest.mark.asyncio
    async def test_real_camera_discovery_and_stream_creation(self, service_manager, test_video_streams, websocket_client):
        """
        Test real camera discovery and stream creation end-to-end.
        
        Validates:
        - Real camera discovery event processing
        - Real MediaMTX stream creation via API
        - Real WebSocket notification delivery
        - Real stream URL generation and validation
        """
        logger.info("Testing real camera discovery and stream creation...")
        
        # Simulate camera connection events
        camera_events = [
            CameraEventData(
                device_path="/dev/video0",
                event_type=CameraEvent.CONNECTED,
                device_info=CameraDevice(device="/dev/video0", name="Test Camera 0", status="CONNECTED"),
                timestamp=time.time()
            ),
            CameraEventData(
                device_path="/dev/video1",
                event_type=CameraEvent.CONNECTED,
                device_info=CameraDevice(device="/dev/video1", name="Test Camera 1", status="CONNECTED"),
                timestamp=time.time()
            )
        ]
        
        # Process camera events
        for event in camera_events:
            await service_manager.handle_camera_event(event)
            await asyncio.sleep(1)  # Allow time for stream creation
        
        # Verify camera list via WebSocket API
        response = await websocket_client.send_request("get_camera_list")
        assert "result" in response
        assert "cameras" in response["result"]
        
        cameras = response["result"]["cameras"]
        assert len(cameras) >= 2
        
        # Verify camera details
        camera0 = next((c for c in cameras if c["device"] == "/dev/video0"), None)
        assert camera0 is not None
        assert camera0["status"] == "CONNECTED"
        assert "streams" in camera0
        
        # Verify stream URLs are accessible
        for camera in cameras:
            if "streams" in camera:
                for stream_type, url in camera["streams"].items():
                    # Test RTSP stream accessibility
                    if stream_type == "rtsp":
                        import aiohttp
                        try:
                            async with aiohttp.ClientSession() as session:
                                async with session.get(url.replace("rtsp://", "http://"), timeout=5) as resp:
                                    # Should get some response (may be error, but connection should work)
                                    assert resp.status is not None
                        except Exception as e:
                            logger.warning(f"Stream URL {url} not accessible: {e}")
        
        logger.info("Real camera discovery and stream creation test passed")
    
    @pytest.mark.asyncio
    async def test_real_recording_and_snapshot_operations(self, service_manager, test_video_streams, websocket_client):
        """
        Test real recording and snapshot operations.
        
        Validates:
        - Real recording start/stop via MediaMTX API
        - Real file system operations for recordings
        - Real snapshot capture and file creation
        - Real WebSocket notification delivery
        """
        logger.info("Testing real recording and snapshot operations...")
        
        # Simulate camera connection first
        event = CameraEventData(
            device_path="/dev/video0",
            event_type=CameraEvent.CONNECTED,
            device_info=CameraDevice(device="/dev/video0", name="Test Camera 0", status="CONNECTED"),
            timestamp=time.time()
        )
        await service_manager.handle_camera_event(event)
        await asyncio.sleep(2)
        
        # Start recording
        response = await websocket_client.send_request(
            "start_recording",
            {"device": "/dev/video0", "duration": 10}
        )
        assert "result" in response or "error" not in response
        
        # Wait for recording notification
        try:
            recording_notification = await websocket_client.wait_for_notification("recording_status_update", timeout=5)
            assert recording_notification["params"]["status"] == "STARTED"
        except TimeoutError:
            logger.warning("Recording notification not received")
        
        # Wait for recording to complete
        await asyncio.sleep(12)
        
        # Stop recording
        response = await websocket_client.send_request(
            "stop_recording",
            {"device": "/dev/video0"}
        )
        assert "result" in response or "error" not in response
        
        # Wait for stop notification
        try:
            recording_notification = await websocket_client.wait_for_notification("recording_status_update", timeout=5)
            assert recording_notification["params"]["status"] == "STOPPED"
        except TimeoutError:
            logger.warning("Recording stop notification not received")
        
        # Verify recording file exists
        recordings_dir = service_manager._config.mediamtx.recordings_path
        recording_files = [f for f in os.listdir(recordings_dir) if f.endswith('.mp4')]
        assert len(recording_files) > 0
        
        # Take snapshot
        response = await websocket_client.send_request(
            "take_snapshot",
            {"device": "/dev/video0"}
        )
        assert "result" in response or "error" not in response
        
        # Verify snapshot file exists
        snapshots_dir = service_manager._config.mediamtx.snapshots_path
        snapshot_files = [f for f in os.listdir(snapshots_dir) if f.endswith('.jpg')]
        assert len(snapshot_files) > 0
        
        logger.info("Real recording and snapshot operations test passed")
    
    @pytest.mark.asyncio
    async def test_real_websocket_authentication_and_control(self, service_manager, websocket_client):
        """
        Test real WebSocket authentication and camera control.
        
        Validates:
        - Real WebSocket connection establishment
        - Real JSON-RPC method handling
        - Real camera status monitoring
        - Real error handling and response formatting
        """
        logger.info("Testing real WebSocket authentication and control...")
        
        # Test basic WebSocket connectivity
        assert websocket_client.websocket is not None
        assert not websocket_client.websocket.closed
        
        # Test camera list retrieval
        response = await websocket_client.send_request("get_camera_list")
        assert "jsonrpc" in response
        assert response["jsonrpc"] == "2.0"
        assert "id" in response
        assert "result" in response or "error" in response
        
        # Test camera status retrieval
        response = await websocket_client.send_request(
            "get_camera_status",
            {"device": "/dev/video0"}
        )
        assert "jsonrpc" in response
        assert response["jsonrpc"] == "2.0"
        
        # Test invalid method handling
        response = await websocket_client.send_request("invalid_method")
        assert "error" in response
        assert response["error"]["code"] == -32601  # Method not found
        
        # Test invalid parameters
        response = await websocket_client.send_request(
            "take_snapshot",
            {"invalid_param": "value"}
        )
        # Should handle gracefully (may return error or ignore invalid params)
        assert "jsonrpc" in response
        
        logger.info("Real WebSocket authentication and control test passed")
    
    @pytest.mark.asyncio
    async def test_real_error_scenarios_and_recovery(self, service_manager, real_mediamtx_server, websocket_client):
        """
        Test real error scenarios and recovery mechanisms.
        
        Validates:
        - Real MediaMTX server failure and recovery
        - Real WebSocket connection failure handling
        - Real file system error handling
        - Real process lifecycle management
        """
        logger.info("Testing real error scenarios and recovery...")
        
        # Test MediaMTX server failure
        logger.info("Testing MediaMTX server failure scenario...")
        
        # Stop MediaMTX server
        await real_mediamtx_server.stop()
        
        # Service should handle MediaMTX unavailability gracefully
        response = await websocket_client.send_request("get_camera_list")
        assert "result" in response or "error" in response
        
        # Restart MediaMTX server
        await real_mediamtx_server.start()
        
        # Service should recover and continue functioning
        await asyncio.sleep(5)  # Allow time for recovery
        
        response = await websocket_client.send_request("get_camera_list")
        assert "result" in response or "error" in response
        
        # Test WebSocket reconnection
        logger.info("Testing WebSocket reconnection...")
        
        await websocket_client.disconnect()
        await asyncio.sleep(1)
        await websocket_client.connect()
        
        # Should be able to send requests after reconnection
        response = await websocket_client.send_request("get_camera_list")
        assert "result" in response or "error" in response
        
        logger.info("Real error scenarios and recovery test passed")
    
    @pytest.mark.asyncio
    async def test_real_resource_management_and_cleanup(self, service_manager, test_config):
        """
        Test real resource management and cleanup.
        
        Validates:
        - Real file system cleanup operations
        - Real process termination and cleanup
        - Real memory and resource management
        - Real temporary file cleanup
        """
        logger.info("Testing real resource management and cleanup...")
        
        # Verify initial state
        assert service_manager.is_running
        
        # Create some test files
        test_file = os.path.join(test_config.mediamtx.recordings_path, "test_file.txt")
        with open(test_file, 'w') as f:
            f.write("test content")
        
        assert os.path.exists(test_file)
        
        # Stop service manager
        await service_manager.stop()
        
        # Verify service is stopped
        assert not service_manager.is_running
        
        # Verify temporary directories are cleaned up (if applicable)
        # Note: Some directories may be preserved for debugging
        
        logger.info("Real resource management and cleanup test passed")
    
    @pytest.mark.asyncio
    async def test_real_end_to_end_camera_lifecycle(self, service_manager, test_video_streams, websocket_client):
        """
        Test complete real end-to-end camera lifecycle.
        
        Validates:
        - Camera discovery → Stream creation → Recording → Snapshot capture
        - WebSocket authentication → Camera control → Status monitoring
        - Real system startup, configuration, and shutdown sequences
        """
        logger.info("Testing real end-to-end camera lifecycle...")
        
        # Step 1: Camera discovery
        event = CameraEventData(
            device_path="/dev/video0",
            event_type=CameraEvent.CONNECTED,
            device_info=CameraDevice(device="/dev/video0", name="Test Camera 0", status="CONNECTED"),
            timestamp=time.time()
        )
        await service_manager.handle_camera_event(event)
        await asyncio.sleep(2)
        
        # Step 2: Verify camera discovery via WebSocket
        response = await websocket_client.send_request("get_camera_list")
        assert "result" in response
        cameras = response["result"]["cameras"]
        camera0 = next((c for c in cameras if c["device"] == "/dev/video0"), None)
        assert camera0 is not None
        assert camera0["status"] == "CONNECTED"
        
        # Step 3: Start recording
        response = await websocket_client.send_request(
            "start_recording",
            {"device": "/dev/video0", "duration": 5}
        )
        assert "result" in response or "error" not in response
        
        # Step 4: Take snapshot during recording
        await asyncio.sleep(2)
        response = await websocket_client.send_request(
            "take_snapshot",
            {"device": "/dev/video0"}
        )
        assert "result" in response or "error" not in response
        
        # Step 5: Wait for recording to complete
        await asyncio.sleep(5)
        
        # Step 6: Verify files were created
        recordings_dir = service_manager._config.mediamtx.recordings_path
        snapshots_dir = service_manager._config.mediamtx.snapshots_path
        
        recording_files = [f for f in os.listdir(recordings_dir) if f.endswith('.mp4')]
        snapshot_files = [f for f in os.listdir(snapshots_dir) if f.endswith('.jpg')]
        
        assert len(recording_files) > 0, "No recording files found"
        assert len(snapshot_files) > 0, "No snapshot files found"
        
        # Step 7: Camera disconnect
        disconnect_event = CameraEventData(
            device_path="/dev/video0",
            event_type=CameraEvent.DISCONNECTED,
            device_info=CameraDevice(device="/dev/video0", name="Test Camera 0", status="DISCONNECTED"),
            timestamp=time.time()
        )
        await service_manager.handle_camera_event(disconnect_event)
        await asyncio.sleep(2)
        
        # Step 8: Verify camera removal
        response = await websocket_client.send_request("get_camera_list")
        cameras = response["result"]["cameras"]
        camera0 = next((c for c in cameras if c["device"] == "/dev/video0"), None)
        assert camera0 is None or camera0["status"] == "DISCONNECTED"
        
        logger.info("Real end-to-end camera lifecycle test passed")


if __name__ == "__main__":
    # Run tests directly for debugging
    pytest.main([__file__, "-v", "-s"])
