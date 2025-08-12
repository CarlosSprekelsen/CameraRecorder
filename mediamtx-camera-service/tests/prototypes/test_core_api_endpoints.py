"""
Critical Prototype: Core API Endpoints with Real aiohttp

This prototype validates core API endpoints with real aiohttp and actual request processing.
It proves design implementability through actual system execution.

PRINCIPLE: NO MOCKING - Only real system validation
"""

import asyncio
import json
import os
import tempfile
import time
from pathlib import Path
from typing import Dict, Any, Optional

import pytest
import pytest_asyncio
import aiohttp
import websockets

# Import real components - NO MOCKING
from src.websocket_server.server import WebSocketJsonRpcServer
from src.camera_service.config import Config, ServerConfig, MediaMTXConfig, CameraConfig, RecordingConfig
from src.camera_service.service_manager import ServiceManager
from src.mediamtx_wrapper.controller import MediaMTXController


class RealCoreAPIEndpointsPrototype:
    """
    Critical prototype for real core API endpoints validation.
    
    This prototype systematically tests core API endpoints using real aiohttp
    to prove design implementability through actual system execution.
    """
    
    def __init__(self):
        self.test_results = {}
        self.system_issues = []
        self.websocket_server = None
        self.service_manager = None
        self.mediamtx_controller = None
        self.temp_dir = None
        self.server_url = "http://127.0.0.1:8000"
        self.websocket_url = "ws://127.0.0.1:8000/ws"
        
    async def setup_real_environment(self):
        """Set up real test environment with actual API endpoints."""
        self.temp_dir = tempfile.mkdtemp(prefix="pdr_api_")
        
        # Create real MediaMTX configuration
        mediamtx_config = MediaMTXConfig(
            host="127.0.0.1",
            api_port=9997,
            rtsp_port=8554,
            webrtc_port=8889,
            hls_port=8888,
            config_path=f"{self.temp_dir}/mediamtx.yml",
            recordings_path=f"{self.temp_dir}/recordings",
            snapshots_path=f"{self.temp_dir}/snapshots"
        )
        
        # Initialize real MediaMTX controller
        self.mediamtx_controller = MediaMTXController(
            host=mediamtx_config.host,
            api_port=mediamtx_config.api_port,
            rtsp_port=mediamtx_config.rtsp_port,
            webrtc_port=mediamtx_config.webrtc_port,
            hls_port=mediamtx_config.hls_port,
            config_path=mediamtx_config.config_path,
            recordings_path=mediamtx_config.recordings_path,
            snapshots_path=mediamtx_config.snapshots_path
        )
        
        # Initialize real service manager
        config = Config(
            server=ServerConfig(host="127.0.0.1", port=8000),
            mediamtx=mediamtx_config,
            camera=CameraConfig(device_range=[0, 1, 2]),
            recording=RecordingConfig(enabled=True)
        )
        
        self.service_manager = ServiceManager(config)
        
        # Initialize real WebSocket server
        self.websocket_server = WebSocketJsonRpcServer(
            host="127.0.0.1",
            port=8000,
            service_manager=self.service_manager
        )
        
    async def cleanup_real_environment(self):
        """Clean up real test environment."""
        if self.websocket_server:
            await self.websocket_server.stop()
        
        if self.mediamtx_controller:
            await self.mediamtx_controller.stop()
        
        if self.service_manager:
            await self.service_manager.shutdown()
            
        if self.temp_dir and os.path.exists(self.temp_dir):
            import shutil
            shutil.rmtree(self.temp_dir)
    
    async def validate_http_api_endpoints(self) -> Dict[str, Any]:
        """Validate real HTTP API endpoints responding to actual requests."""
        try:
            async with aiohttp.ClientSession() as session:
                # Test health endpoint
                async with session.get(f"{self.server_url}/health") as response:
                    health_status = response.status
                    health_data = await response.json()
                
                # Test status endpoint
                async with session.get(f"{self.server_url}/status") as response:
                    status_status = response.status
                    status_data = await response.json()
                
                # Test configuration endpoint
                async with session.get(f"{self.server_url}/config") as response:
                    config_status = response.status
                    config_data = await response.json()
                
                # Test cameras endpoint
                async with session.get(f"{self.server_url}/cameras") as response:
                    cameras_status = response.status
                    cameras_data = await response.json()
                
                # Test streams endpoint
                async with session.get(f"{self.server_url}/streams") as response:
                    streams_status = response.status
                    streams_data = await response.json()
            
            return {
                "health_endpoint": {"status": health_status, "data": health_data},
                "status_endpoint": {"status": status_status, "data": status_data},
                "config_endpoint": {"status": config_status, "data": config_data},
                "cameras_endpoint": {"status": cameras_status, "data": cameras_data},
                "streams_endpoint": {"status": streams_status, "data": streams_data}
            }
            
        except Exception as e:
            self.system_issues.append(f"HTTP API endpoints validation failed: {str(e)}")
            raise
    
    async def validate_websocket_json_rpc_endpoints(self) -> Dict[str, Any]:
        """Validate real WebSocket JSON-RPC endpoints."""
        try:
            async with websockets.connect(self.websocket_url) as websocket:
                # Test ping/pong
                ping_message = {
                    "jsonrpc": "2.0",
                    "method": "ping",
                    "params": {},
                    "id": 1
                }
                await websocket.send(json.dumps(ping_message))
                ping_response = await asyncio.wait_for(websocket.recv(), timeout=5.0)
                ping_data = json.loads(ping_response)
                
                # Test get_status
                status_message = {
                    "jsonrpc": "2.0",
                    "method": "get_status",
                    "params": {},
                    "id": 2
                }
                await websocket.send(json.dumps(status_message))
                status_response = await asyncio.wait_for(websocket.recv(), timeout=5.0)
                status_data = json.loads(status_response)
                
                # Test get_cameras
                cameras_message = {
                    "jsonrpc": "2.0",
                    "method": "get_cameras",
                    "params": {},
                    "id": 3
                }
                await websocket.send(json.dumps(cameras_message))
                cameras_response = await asyncio.wait_for(websocket.recv(), timeout=5.0)
                cameras_data = json.loads(cameras_response)
                
                # Test get_streams
                streams_message = {
                    "jsonrpc": "2.0",
                    "method": "get_streams",
                    "params": {},
                    "id": 4
                }
                await websocket.send(json.dumps(streams_message))
                streams_response = await asyncio.wait_for(websocket.recv(), timeout=5.0)
                streams_data = json.loads(streams_response)
            
            return {
                "ping_response": ping_data,
                "status_response": status_data,
                "cameras_response": cameras_data,
                "streams_response": streams_data
            }
            
        except Exception as e:
            self.system_issues.append(f"WebSocket JSON-RPC validation failed: {str(e)}")
            raise
    
    async def validate_api_request_processing(self) -> Dict[str, Any]:
        """Validate real API request processing with actual data."""
        try:
            async with aiohttp.ClientSession() as session:
                # Test POST request to create stream
                create_stream_data = {
                    "camera_id": "test_camera_1",
                    "stream_name": "test_stream",
                    "format": "rtsp"
                }
                
                async with session.post(
                    f"{self.server_url}/streams",
                    json=create_stream_data
                ) as response:
                    create_status = response.status
                    create_data = await response.json()
                
                # Test PUT request to update stream
                update_stream_data = {
                    "format": "webrtc",
                    "quality": "high"
                }
                
                async with session.put(
                    f"{self.server_url}/streams/test_stream",
                    json=update_stream_data
                ) as response:
                    update_status = response.status
                    update_data = await response.json()
                
                # Test DELETE request to remove stream
                async with session.delete(f"{self.server_url}/streams/test_stream") as response:
                    delete_status = response.status
                    delete_data = await response.json()
            
            return {
                "create_stream": {"status": create_status, "data": create_data},
                "update_stream": {"status": update_status, "data": update_data},
                "delete_stream": {"status": delete_status, "data": delete_data}
            }
            
        except Exception as e:
            self.system_issues.append(f"API request processing failed: {str(e)}")
            raise
    
    async def validate_api_error_handling(self) -> Dict[str, Any]:
        """Validate real API error handling with invalid requests."""
        try:
            async with aiohttp.ClientSession() as session:
                # Test invalid endpoint
                async with session.get(f"{self.server_url}/invalid_endpoint") as response:
                    invalid_status = response.status
                    invalid_data = await response.json()
                
                # Test invalid JSON
                async with session.post(
                    f"{self.server_url}/streams",
                    data="invalid json",
                    headers={"Content-Type": "application/json"}
                ) as response:
                    invalid_json_status = response.status
                    invalid_json_data = await response.json()
                
                # Test missing required fields
                incomplete_data = {"camera_id": "test_camera"}
                async with session.post(
                    f"{self.server_url}/streams",
                    json=incomplete_data
                ) as response:
                    incomplete_status = response.status
                    incomplete_data_response = await response.json()
            
            return {
                "invalid_endpoint": {"status": invalid_status, "data": invalid_data},
                "invalid_json": {"status": invalid_json_status, "data": invalid_json_data},
                "incomplete_data": {"status": incomplete_status, "data": incomplete_data_response}
            }
            
        except Exception as e:
            self.system_issues.append(f"API error handling validation failed: {str(e)}")
            raise
    
    async def validate_api_performance(self) -> Dict[str, Any]:
        """Validate real API performance under load."""
        try:
            async with aiohttp.ClientSession() as session:
                # Measure response times for multiple requests
                response_times = []
                
                for i in range(10):
                    start_time = time.time()
                    
                    async with session.get(f"{self.server_url}/health") as response:
                        await response.json()
                    
                    end_time = time.time()
                    response_times.append(end_time - start_time)
                
                # Calculate performance metrics
                avg_response_time = sum(response_times) / len(response_times)
                min_response_time = min(response_times)
                max_response_time = max(response_times)
                
                # Test concurrent requests
                concurrent_start = time.time()
                
                async def make_request():
                    async with session.get(f"{self.server_url}/status") as response:
                        return await response.json()
                
                concurrent_requests = [make_request() for _ in range(5)]
                concurrent_results = await asyncio.gather(*concurrent_requests)
                
                concurrent_end = time.time()
                concurrent_time = concurrent_end - concurrent_start
            
            return {
                "avg_response_time": avg_response_time,
                "min_response_time": min_response_time,
                "max_response_time": max_response_time,
                "concurrent_requests_time": concurrent_time,
                "concurrent_requests_count": len(concurrent_results),
                "response_times": response_times
            }
            
        except Exception as e:
            self.system_issues.append(f"API performance validation failed: {str(e)}")
            raise
    
    async def run_comprehensive_api_validation(self) -> Dict[str, Any]:
        """Run comprehensive API endpoints validation."""
        try:
            await self.setup_real_environment()
            
            # Start servers
            await self.websocket_server.start()
            await self.mediamtx_controller.start()
            await asyncio.sleep(2)
            
            # Execute all validation steps
            results = {
                "http_api_endpoints": await self.validate_http_api_endpoints(),
                "websocket_json_rpc": await self.validate_websocket_json_rpc_endpoints(),
                "api_request_processing": await self.validate_api_request_processing(),
                "api_error_handling": await self.validate_api_error_handling(),
                "api_performance": await self.validate_api_performance(),
                "system_issues": self.system_issues
            }
            
            self.test_results = results
            return results
            
        finally:
            await self.cleanup_real_environment()


@pytest.mark.pdr
@pytest.mark.asyncio
class TestRealCoreAPIEndpoints:
    """Critical prototype tests for real core API endpoints."""
    
    def setup_method(self):
        """Set up prototype for each test method."""
        self.prototype = RealCoreAPIEndpointsPrototype()
    
    async def teardown_method(self):
        """Clean up after each test method."""
        if hasattr(self, 'prototype'):
            await self.prototype.cleanup_real_environment()
    
    async def test_http_api_endpoints_real_responses(self):
        """Test real HTTP API endpoints responding to actual requests."""
        await self.prototype.setup_real_environment()
        
        try:
            # Start servers
            await self.prototype.websocket_server.start()
            await self.prototype.mediamtx_controller.start()
            await asyncio.sleep(2)
            
            result = await self.prototype.validate_http_api_endpoints()
            
            # Validate results
            assert result["health_endpoint"]["status"] == 200, "Health endpoint failed"
            assert result["status_endpoint"]["status"] == 200, "Status endpoint failed"
            assert result["config_endpoint"]["status"] == 200, "Config endpoint failed"
            assert result["cameras_endpoint"]["status"] == 200, "Cameras endpoint failed"
            assert result["streams_endpoint"]["status"] == 200, "Streams endpoint failed"
            
            print(f"✅ HTTP API endpoints validation: {result}")
            
        finally:
            await self.prototype.cleanup_real_environment()
    
    async def test_websocket_json_rpc_real_endpoints(self):
        """Test real WebSocket JSON-RPC endpoints."""
        await self.prototype.setup_real_environment()
        
        try:
            # Start servers
            await self.prototype.websocket_server.start()
            await self.prototype.mediamtx_controller.start()
            await asyncio.sleep(2)
            
            result = await self.prototype.validate_websocket_json_rpc_endpoints()
            
            # Validate results
            assert "result" in result["ping_response"], "Ping response invalid"
            assert "result" in result["status_response"], "Status response invalid"
            assert "result" in result["cameras_response"], "Cameras response invalid"
            assert "result" in result["streams_response"], "Streams response invalid"
            
            print(f"✅ WebSocket JSON-RPC validation: {result}")
            
        finally:
            await self.prototype.cleanup_real_environment()
    
    async def test_api_request_real_processing(self):
        """Test real API request processing with actual data."""
        await self.prototype.setup_real_environment()
        
        try:
            # Start servers
            await self.prototype.websocket_server.start()
            await self.prototype.mediamtx_controller.start()
            await asyncio.sleep(2)
            
            result = await self.prototype.validate_api_request_processing()
            
            # Validate results
            assert result["create_stream"]["status"] in [200, 201], "Create stream failed"
            assert result["update_stream"]["status"] in [200, 204], "Update stream failed"
            assert result["delete_stream"]["status"] in [200, 204], "Delete stream failed"
            
            print(f"✅ API request processing validation: {result}")
            
        finally:
            await self.prototype.cleanup_real_environment()
    
    async def test_api_error_real_handling(self):
        """Test real API error handling with invalid requests."""
        await self.prototype.setup_real_environment()
        
        try:
            # Start servers
            await self.prototype.websocket_server.start()
            await self.prototype.mediamtx_controller.start()
            await asyncio.sleep(2)
            
            result = await self.prototype.validate_api_error_handling()
            
            # Validate results
            assert result["invalid_endpoint"]["status"] == 404, "Invalid endpoint not handled"
            assert result["invalid_json"]["status"] == 400, "Invalid JSON not handled"
            assert result["incomplete_data"]["status"] == 400, "Incomplete data not handled"
            
            print(f"✅ API error handling validation: {result}")
            
        finally:
            await self.prototype.cleanup_real_environment()
    
    async def test_api_real_performance(self):
        """Test real API performance under load."""
        await self.prototype.setup_real_environment()
        
        try:
            # Start servers
            await self.prototype.websocket_server.start()
            await self.prototype.mediamtx_controller.start()
            await asyncio.sleep(2)
            
            result = await self.prototype.validate_api_performance()
            
            # Validate results
            assert result["avg_response_time"] < 1.0, "Average response time too high"
            assert result["concurrent_requests_time"] < 5.0, "Concurrent requests too slow"
            assert result["concurrent_requests_count"] == 5, "Not all concurrent requests completed"
            
            print(f"✅ API performance validation: {result}")
            
        finally:
            await self.prototype.cleanup_real_environment()
    
    async def test_comprehensive_api_validation(self):
        """Test comprehensive API endpoints validation."""
        result = await self.prototype.run_comprehensive_api_validation()
        
        # Validate comprehensive results
        assert len(result["system_issues"]) == 0, f"System issues found: {result['system_issues']}"
        assert result["http_api_endpoints"]["health_endpoint"]["status"] == 200, "Comprehensive HTTP failed"
        assert "result" in result["websocket_json_rpc"]["ping_response"], "Comprehensive WebSocket failed"
        assert result["api_request_processing"]["create_stream"]["status"] in [200, 201], "Comprehensive request processing failed"
        assert result["api_error_handling"]["invalid_endpoint"]["status"] == 404, "Comprehensive error handling failed"
        assert result["api_performance"]["avg_response_time"] < 1.0, "Comprehensive performance failed"
        
        print(f"✅ Comprehensive API validation: {result}")
        
        # Log results for evidence
        with open("/tmp/pdr_api_endpoints_results.json", "w") as f:
            json.dump(result, f, indent=2, default=str)
