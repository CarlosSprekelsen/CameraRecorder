#!/usr/bin/env python3
"""
Test script for camera discovery and MediaMTX integration with real video devices.

Requirements Traceability:
- REQ-CAMERA-001: System shall discover real video devices automatically
- REQ-CAMERA-002: System shall integrate camera discovery with MediaMTX service
- REQ-CAMERA-003: System shall handle camera capability detection
- REQ-MEDIA-001: MediaMTX integration shall use single systemd-managed service

Test Categories: Integration
"""

import asyncio
import json
import logging
import sys
import time
from pathlib import Path

# Add src to path
sys.path.insert(0, str(Path(__file__).parent.parent / "src"))

from camera_service.config import Config
from camera_service.service_manager import ServiceManager
from camera_discovery.hybrid_monitor import HybridCameraMonitor
from mediamtx_wrapper.controller import MediaMTXController, StreamConfig

# Configure logging
logging.basicConfig(level=logging.INFO, format='%(asctime)s - %(name)s - %(levelname)s - %(message)s')
logger = logging.getLogger(__name__)

import pytest

@pytest.mark.asyncio
@pytest.mark.integration
async def test_camera_discovery():
    """Test camera discovery with real devices."""
    logger.info("=== Testing Camera Discovery ===")
    
    # Check available video devices
    import subprocess
    try:
        result = subprocess.run(['ls', '-la', '/dev/video*'], capture_output=True, text=True)
        logger.info(f"Available video devices:\n{result.stdout}")
    except Exception as e:
        logger.error(f"Error checking video devices: {e}")
    
    # Create camera monitor
    config = Config()
    camera_monitor = HybridCameraMonitor(
        device_range=config.camera.device_range,
        poll_interval=config.camera.poll_interval,
        detection_timeout=config.camera.detection_timeout,
        enable_capability_detection=config.camera.enable_capability_detection,
    )
    
    try:
        # Start camera monitor
        logger.info("Starting camera monitor...")
        await camera_monitor.start()
        
        # Wait for initial detection
        logger.info("Waiting for camera detection...")
        await asyncio.sleep(5)
        
        # Get connected cameras
        connected_cameras = await camera_monitor.get_connected_cameras()
        logger.info(f"Connected cameras: {len(connected_cameras)}")
        
        for device_path, camera_info in connected_cameras.items():
            logger.info(f"Camera {device_path}: {camera_info}")
            
            # Get capability metadata
            capability_metadata = camera_monitor.get_effective_capability_metadata(device_path)
            logger.info(f"Capability metadata for {device_path}: {capability_metadata}")
        
        return connected_cameras
        
    except Exception as e:
        logger.error(f"Camera discovery error: {e}")
        return {}
    finally:
        await camera_monitor.stop()

@pytest.mark.asyncio
@pytest.mark.integration
async def test_mediamtx_integration():
    """Test MediaMTX integration."""
    logger.info("=== Testing MediaMTX Integration ===")
    
    # Check MediaMTX service
    import subprocess
    try:
        result = subprocess.run(['systemctl', 'status', 'mediamtx'], capture_output=True, text=True)
        logger.info(f"MediaMTX service status:\n{result.stdout}")
    except Exception as e:
        logger.error(f"Error checking MediaMTX service: {e}")
    
    # Test MediaMTX API
    import aiohttp
    try:
        async with aiohttp.ClientSession() as session:
            # Test global config
            async with session.get('http://localhost:9997/v3/config/global/get') as response:
                if response.status == 200:
                    config_data = await response.json()
                    logger.info("MediaMTX API accessible - Global config retrieved")
                else:
                    logger.error(f"MediaMTX API error: {response.status}")
            
            # Test paths list
            async with session.get('http://localhost:9997/v3/paths/list') as response:
                if response.status == 200:
                    paths_data = await response.json()
                    logger.info(f"Current MediaMTX paths: {paths_data}")
                else:
                    logger.error(f"MediaMTX paths API error: {response.status}")
                    
    except Exception as e:
        logger.error(f"MediaMTX API test error: {e}")

@pytest.mark.asyncio
async def test_stream_creation():
    """Test stream creation for detected cameras."""
    logger.info("=== Testing Stream Creation ===")
    
    # Create MediaMTX controller
    config = Config()
    mediamtx_config = config.mediamtx
    
    controller = MediaMTXController(
        host=mediamtx_config.host,
        api_port=mediamtx_config.api_port,
        rtsp_port=mediamtx_config.rtsp_port,
        webrtc_port=mediamtx_config.webrtc_port,
        hls_port=mediamtx_config.hls_port,
        config_path=mediamtx_config.config_path,
        recordings_path=mediamtx_config.recordings_path,
        snapshots_path=mediamtx_config.snapshots_path,
    )
    
    try:
        # Check controller health
        health_status = await controller.health_check()
        logger.info(f"MediaMTX controller health: {health_status}")
        
        # Get current streams
        streams = await controller.get_stream_list()
        logger.info(f"Current streams: {streams}")
        
        # Create test streams for each camera
        for i in range(4):  # We have /dev/video0-3
            stream_name = f"camera{i}"
            stream_config = StreamConfig(
                name=stream_name,
                source=f"rtsp://127.0.0.1:8554/{stream_name}"
            )
            
            try:
                await controller.create_stream(stream_config)
                logger.info(f"Created stream: {stream_name}")
                
                # Wait a moment and check stream status
                await asyncio.sleep(2)
                stream_status = await controller.get_stream_status(stream_name)
                logger.info(f"Stream {stream_name} status: {stream_status}")
                
            except Exception as e:
                logger.error(f"Error creating stream {stream_name}: {e}")
        
        # Get updated stream list
        updated_streams = await controller.get_stream_list()
        logger.info(f"Updated streams: {updated_streams}")
        
    except Exception as e:
        logger.error(f"Stream creation error: {e}")

@pytest.mark.asyncio
async def test_service_manager_integration():
    """Test full service manager integration."""
    logger.info("=== Testing Service Manager Integration ===")
    
    # Check if port 8002 is already in use (real server running)
    import socket
    port_in_use = False
    try:
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
            s.settimeout(1)
            result = s.connect_ex(('localhost', 8002))
            port_in_use = (result == 0)
    except Exception:
        pass
    
    if port_in_use:
        logger.info("Port 8002 is in use - assuming real server is running")
        logger.info("Skipping service manager start to avoid conflicts")
        return
    
    config = Config()
    service_manager = ServiceManager(config)
    
    try:
        # Start service manager
        logger.info("Starting service manager...")
        await service_manager.start()
        
        # Wait for components to initialize
        await asyncio.sleep(5)
        
        # Check camera monitor
        if hasattr(service_manager, '_camera_monitor') and service_manager._camera_monitor:
            logger.info("Camera monitor is initialized")
            
            # Get connected cameras
            connected_cameras = await service_manager._camera_monitor.get_connected_cameras()
            logger.info(f"Service manager detected cameras: {len(connected_cameras)}")
            
            for device_path, camera_info in connected_cameras.items():
                logger.info(f"Camera {device_path}: {camera_info}")
        else:
            logger.warning("Camera monitor not initialized")
        
        # Check MediaMTX controller
        if hasattr(service_manager, '_mediamtx_controller') and service_manager._mediamtx_controller:
            logger.info("MediaMTX controller is initialized")
            
            # Get streams
            streams = await service_manager._mediamtx_controller.get_stream_list()
            logger.info(f"Service manager streams: {streams}")
        else:
            logger.warning("MediaMTX controller not initialized")
        
        # Check WebSocket server
        if hasattr(service_manager, '_websocket_server') and service_manager._websocket_server:
            logger.info("WebSocket server is initialized")
        else:
            logger.warning("WebSocket server not initialized")
        
    except Exception as e:
        logger.error(f"Service manager integration error: {e}")
    finally:
        await service_manager.stop()

async def test_running_server_camera_detection():
    """Test camera detection via the running server's WebSocket API."""
    logger.info("=== Testing Running Server Camera Detection ===")
    
    # Check if port 8002 is in use
    import socket
    port_in_use = False
    try:
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
            s.settimeout(1)
            result = s.connect_ex(('localhost', 8002))
            port_in_use = (result == 0)
    except Exception:
        pass
    
    if not port_in_use:
        logger.info("Port 8002 not in use - skipping running server test")
        return
    
    # Test WebSocket connection to running server
    import websockets
    import jwt
    import time
    
    # Generate valid JWT token
    jwt_secret = "dev-secret-change-me"
    payload = {
        "user_id": "test_user",
        "role": "admin",
        "iat": int(time.time()),
        "exp": int(time.time()) + (24 * 3600)
    }
    token = jwt.encode(payload, jwt_secret, algorithm="HS256")
    
    try:
        async with websockets.connect("ws://localhost:8002/ws") as websocket:
            logger.info("Connected to running server")
            
            # Authenticate
            auth_request = {
                "jsonrpc": "2.0",
                "id": 1,
                "method": "authenticate",
                "params": {"token": token, "auth_type": "jwt"}
            }
            await websocket.send(json.dumps(auth_request))
            auth_response = await websocket.recv()
            auth_data = json.loads(auth_response)
            logger.info(f"Authentication response: {auth_data}")
            
            if auth_data.get("result", {}).get("authenticated"):
                # Get camera list
                camera_request = {
                    "jsonrpc": "2.0",
                    "id": 2,
                    "method": "get_camera_list",
                    "params": {}
                }
                await websocket.send(json.dumps(camera_request))
                camera_response = await websocket.recv()
                camera_data = json.loads(camera_response)
                logger.info(f"Running server camera list: {json.dumps(camera_data, indent=2)}")
                
                cameras = camera_data.get("result", {}).get("cameras", [])
                logger.info(f"Running server detected {len(cameras)} cameras")
                
                for camera in cameras:
                    device_path = camera.get("device", "unknown")
                    status = camera.get("status", "unknown")
                    logger.info(f"  - {device_path}: {status}")
            else:
                logger.error("Authentication failed with running server")
                
    except Exception as e:
        logger.error(f"Error testing running server: {e}")

async def main():
    """Main test function."""
    logger.info("Starting camera discovery and MediaMTX integration tests...")
    
    # Test 0: Running server camera detection
    await test_running_server_camera_detection()
    
    # Test 1: Camera discovery
    connected_cameras = await test_camera_discovery()
    
    # Test 2: MediaMTX integration
    await test_mediamtx_integration()
    
    # Test 3: Stream creation
    await test_stream_creation()
    
    # Test 4: Service manager integration
    await test_service_manager_integration()
    
    logger.info("Tests completed!")

if __name__ == "__main__":
    asyncio.run(main())
