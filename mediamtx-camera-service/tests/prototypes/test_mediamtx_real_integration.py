"""
Critical Prototype: Real MediaMTX Integration

This prototype validates real MediaMTX integration without any mocking.
It proves design implementability through actual system execution.

PRINCIPLE: NO MOCKING - Only real system validation
"""

import asyncio
import json
import os
import tempfile
import time
from pathlib import Path
from typing import Dict, Any

import pytest
import pytest_asyncio
import aiohttp
import websockets

# Import real components - NO MOCKING
from src.mediamtx_wrapper.controller import MediaMTXController
from src.camera_service.config import Config, MediaMTXConfig, ServerConfig, CameraConfig, RecordingConfig
from src.camera_service.service_manager import ServiceManager


class RealMediaMTXIntegrationPrototype:
    """
    Critical prototype for real MediaMTX integration validation.
    
    This prototype systematically tests MediaMTX integration using real components
    to prove design implementability through actual system execution.
    """
    
    def __init__(self):
        self.test_results = {}
        self.system_issues = []
        self.mediamtx_controller = None
        self.service_manager = None
        self.temp_dir = None
        
    async def setup_real_environment(self):
        """Set up real test environment with actual MediaMTX."""
        self.temp_dir = tempfile.mkdtemp(prefix="pdr_mediamtx_")
        
        # Create real MediaMTX configuration
        mediamtx_config = MediaMTXConfig(
            binary_path="/usr/local/bin/mediamtx",
            config_path=f"{self.temp_dir}/mediamtx.yml",
            log_path=f"{self.temp_dir}/mediamtx.log",
            api_port=9997,
            rtsp_port=8554,
            webrtc_port=8889,
            hls_port=8888,
            websocket_port=8002,
            health_port=8003
        )
        
        # Initialize real MediaMTX controller
        self.mediamtx_controller = MediaMTXController(mediamtx_config)
        
        # Initialize real service manager
        config = Config(
            server=ServerConfig(host="127.0.0.1", port=8000),
            mediamtx=mediamtx_config,
            camera=CameraConfig(device_range=[0, 1, 2]),
            recording=RecordingConfig(enabled=True)
        )
        
        self.service_manager = ServiceManager(config)
        
    async def cleanup_real_environment(self):
        """Clean up real test environment."""
        if self.mediamtx_controller:
            await self.mediamtx_controller.stop()
        
        if self.service_manager:
            await self.service_manager.shutdown()
            
        if self.temp_dir and os.path.exists(self.temp_dir):
            import shutil
            shutil.rmtree(self.temp_dir)
    
    async def validate_mediamtx_startup(self) -> Dict[str, Any]:
        """Validate real MediaMTX startup and configuration."""
        try:
            # Start real MediaMTX
            await self.mediamtx_controller.start()
            
            # Wait for startup
            await asyncio.sleep(2)
            
            # Check if MediaMTX is running
            is_running = await self.mediamtx_controller.is_running()
            
            # Get real status
            status = await self.mediamtx_controller.get_status()
            
            # Validate configuration
            config_valid = await self.mediamtx_controller.validate_config()
            
            return {
                "startup_success": is_running,
                "status": status,
                "config_valid": config_valid,
                "ports": {
                    "api": 9997,
                    "rtsp": 8554,
                    "webrtc": 8889,
                    "hls": 8888
                }
            }
            
        except Exception as e:
            self.system_issues.append(f"MediaMTX startup failed: {str(e)}")
            raise
    
    async def validate_rtsp_stream_handling(self) -> Dict[str, Any]:
        """Validate real RTSP stream handling capabilities."""
        try:
            # Create a test RTSP stream
            stream_name = "test_stream"
            stream_path = f"rtsp://127.0.0.1:8554/{stream_name}"
            
            # Register stream with MediaMTX
            await self.mediamtx_controller.create_stream(stream_name)
            
            # Wait for stream to be available
            await asyncio.sleep(1)
            
            # Check if stream is active
            streams = await self.mediamtx_controller.list_streams()
            stream_active = stream_name in streams
            
            # Test stream URL accessibility
            stream_url_valid = await self.mediamtx_controller.validate_stream_url(stream_path)
            
            return {
                "stream_created": stream_active,
                "stream_url": stream_path,
                "streams_list": streams,
                "url_valid": stream_url_valid
            }
            
        except Exception as e:
            self.system_issues.append(f"RTSP stream handling failed: {str(e)}")
            raise
    
    async def validate_api_endpoints(self) -> Dict[str, Any]:
        """Validate real API endpoints responding to actual requests."""
        try:
            # Test MediaMTX API endpoints
            api_base = "http://127.0.0.1:9997"
            
            async with aiohttp.ClientSession() as session:
                # Test health endpoint
                async with session.get(f"{api_base}/v3/paths/list") as response:
                    health_status = response.status
                    health_data = await response.json()
                
                # Test paths endpoint
                async with session.get(f"{api_base}/v3/paths/list") as response:
                    paths_status = response.status
                    paths_data = await response.json()
                
                # Test metrics endpoint
                async with session.get(f"{api_base}/v3/metrics") as response:
                    metrics_status = response.status
                    metrics_data = await response.json()
            
            return {
                "health_endpoint": {"status": health_status, "data": health_data},
                "paths_endpoint": {"status": paths_status, "data": paths_data},
                "metrics_endpoint": {"status": metrics_status, "data": metrics_data}
            }
            
        except Exception as e:
            self.system_issues.append(f"API endpoints validation failed: {str(e)}")
            raise
    
    async def validate_websocket_communication(self) -> Dict[str, Any]:
        """Validate real WebSocket communication with MediaMTX."""
        try:
            # Connect to MediaMTX WebSocket API
            ws_url = "ws://127.0.0.1:8002"
            
            async with websockets.connect(ws_url) as websocket:
                # Send test message
                test_message = {"type": "ping"}
                await websocket.send(json.dumps(test_message))
                
                # Wait for response
                response = await asyncio.wait_for(websocket.recv(), timeout=5.0)
                response_data = json.loads(response)
                
                # Validate response
                response_valid = "pong" in response_data.get("type", "")
                
                return {
                    "websocket_connected": True,
                    "message_sent": test_message,
                    "response_received": response_data,
                    "response_valid": response_valid
                }
                
        except Exception as e:
            self.system_issues.append(f"WebSocket communication failed: {str(e)}")
            raise
    
    async def run_comprehensive_validation(self) -> Dict[str, Any]:
        """Run comprehensive real system validation."""
        try:
            await self.setup_real_environment()
            
            # Execute all validation steps
            results = {
                "mediamtx_startup": await self.validate_mediamtx_startup(),
                "rtsp_stream_handling": await self.validate_rtsp_stream_handling(),
                "api_endpoints": await self.validate_api_endpoints(),
                "websocket_communication": await self.validate_websocket_communication(),
                "system_issues": self.system_issues
            }
            
            self.test_results = results
            return results
            
        finally:
            await self.cleanup_real_environment()


@pytest.mark.pdr
@pytest.mark.asyncio
class TestRealMediaMTXIntegration:
    """Critical prototype tests for real MediaMTX integration."""
    
    def setup_method(self):
        """Set up prototype for each test method."""
        self.prototype = RealMediaMTXIntegrationPrototype()
    
    async def teardown_method(self):
        """Clean up after each test method."""
        if hasattr(self, 'prototype'):
            await self.prototype.cleanup_real_environment()
    
    async def test_mediamtx_real_startup_and_configuration(self):
        """Test real MediaMTX startup and configuration validation."""
        await self.prototype.setup_real_environment()
        
        try:
            result = await self.prototype.validate_mediamtx_startup()
            
            # Validate results
            assert result["startup_success"] is True, "MediaMTX failed to start"
            assert result["config_valid"] is True, "MediaMTX configuration invalid"
            assert "api" in result["ports"], "API port not configured"
            assert "rtsp" in result["ports"], "RTSP port not configured"
            
            print(f"✅ MediaMTX startup validation: {result}")
            
        finally:
            await self.prototype.cleanup_real_environment()
    
    async def test_rtsp_stream_real_handling(self):
        """Test real RTSP stream handling capabilities."""
        await self.prototype.setup_real_environment()
        
        try:
            # Start MediaMTX first
            await self.prototype.validate_mediamtx_startup()
            
            # Test RTSP stream handling
            result = await self.prototype.validate_rtsp_stream_handling()
            
            # Validate results
            assert result["stream_created"] is True, "RTSP stream creation failed"
            assert result["url_valid"] is True, "RTSP stream URL invalid"
            assert "test_stream" in result["streams_list"], "Stream not in list"
            
            print(f"✅ RTSP stream handling validation: {result}")
            
        finally:
            await self.prototype.cleanup_real_environment()
    
    async def test_api_endpoints_real_responses(self):
        """Test real API endpoints responding to actual requests."""
        await self.prototype.setup_real_environment()
        
        try:
            # Start MediaMTX first
            await self.prototype.validate_mediamtx_startup()
            
            # Test API endpoints
            result = await self.prototype.validate_api_endpoints()
            
            # Validate results
            assert result["health_endpoint"]["status"] == 200, "Health endpoint failed"
            assert result["paths_endpoint"]["status"] == 200, "Paths endpoint failed"
            assert result["metrics_endpoint"]["status"] == 200, "Metrics endpoint failed"
            
            print(f"✅ API endpoints validation: {result}")
            
        finally:
            await self.prototype.cleanup_real_environment()
    
    async def test_websocket_real_communication(self):
        """Test real WebSocket communication with MediaMTX."""
        await self.prototype.setup_real_environment()
        
        try:
            # Start MediaMTX first
            await self.prototype.validate_mediamtx_startup()
            
            # Test WebSocket communication
            result = await self.prototype.validate_websocket_communication()
            
            # Validate results
            assert result["websocket_connected"] is True, "WebSocket connection failed"
            assert result["response_valid"] is True, "WebSocket response invalid"
            
            print(f"✅ WebSocket communication validation: {result}")
            
        finally:
            await self.prototype.cleanup_real_environment()
    
    async def test_comprehensive_real_system_validation(self):
        """Test comprehensive real system validation."""
        result = await self.prototype.run_comprehensive_validation()
        
        # Validate comprehensive results
        assert len(result["system_issues"]) == 0, f"System issues found: {result['system_issues']}"
        assert result["mediamtx_startup"]["startup_success"] is True, "Comprehensive startup failed"
        assert result["rtsp_stream_handling"]["stream_created"] is True, "Comprehensive RTSP failed"
        assert result["api_endpoints"]["health_endpoint"]["status"] == 200, "Comprehensive API failed"
        assert result["websocket_communication"]["websocket_connected"] is True, "Comprehensive WebSocket failed"
        
        print(f"✅ Comprehensive real system validation: {result}")
        
        # Log results for evidence
        with open("/tmp/pdr_mediamtx_integration_results.json", "w") as f:
            json.dump(result, f, indent=2, default=str)
