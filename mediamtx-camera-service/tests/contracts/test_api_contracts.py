"""
Contract Tests: API Interface Validation

This module validates API contracts against real endpoints to ensure
interfaces work as specified in the design specifications.

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


class APIContractValidator:
    """
    Validates API contracts against real endpoints.
    
    This validator ensures that the API interfaces work as specified
    in the design specifications through real system testing.
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
        self.temp_dir = tempfile.mkdtemp(prefix="ivv_contracts_")
        
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
            websocket_path="/ws",
            max_connections=100
        )
        self.websocket_server.set_service_manager(self.service_manager)
        
    async def cleanup_real_environment(self):
        """Clean up real test environment."""
        if self.websocket_server:
            await self.websocket_server.stop()
        
        if self.mediamtx_controller:
            await self.mediamtx_controller.stop()
        
        if self.service_manager:
            await self.service_manager.stop()
            
        if self.temp_dir and os.path.exists(self.temp_dir):
            import shutil
            shutil.rmtree(self.temp_dir)
    
    async def validate_json_rpc_contract(self) -> Dict[str, Any]:
        """Validate JSON-RPC 2.0 contract compliance."""
        try:
            async with websockets.connect(self.websocket_url) as websocket:
                # Test valid JSON-RPC 2.0 request
                valid_request = {
                    "jsonrpc": "2.0",
                    "method": "ping",
                    "params": {},
                    "id": 1
                }
                await websocket.send(json.dumps(valid_request))
                response = await asyncio.wait_for(websocket.recv(), timeout=5.0)
                response_data = json.loads(response)
                
                # Validate JSON-RPC 2.0 response structure
                jsonrpc_valid = response_data.get("jsonrpc") == "2.0"
                id_valid = response_data.get("id") == 1
                result_valid = "result" in response_data or "error" in response_data
                
                # Test invalid JSON-RPC request
                invalid_request = {
                    "jsonrpc": "1.0",  # Invalid version
                    "method": "ping",
                    "params": {},
                    "id": 2
                }
                await websocket.send(json.dumps(invalid_request))
                error_response = await asyncio.wait_for(websocket.recv(), timeout=5.0)
                error_data = json.loads(error_response)
                
                # Validate error response
                error_valid = "error" in error_data
                error_code_valid = error_data.get("error", {}).get("code") == -32600  # Invalid Request
                
                return {
                    "jsonrpc_version_valid": jsonrpc_valid,
                    "id_field_valid": id_valid,
                    "result_error_valid": result_valid,
                    "error_handling_valid": error_valid,
                    "error_code_valid": error_code_valid,
                    "valid_response": response_data,
                    "error_response": error_data
                }
                
        except Exception as e:
            self.system_issues.append(f"JSON-RPC contract validation failed: {str(e)}")
            raise
    
    async def validate_method_contracts(self) -> Dict[str, Any]:
        """Validate method contracts for core API methods."""
        try:
            async with websockets.connect(self.websocket_url) as websocket:
                method_results = {}
                
                # Test get_status method
                status_request = {
                    "jsonrpc": "2.0",
                    "method": "get_status",
                    "params": {},
                    "id": 1
                }
                await websocket.send(json.dumps(status_request))
                status_response = await asyncio.wait_for(websocket.recv(), timeout=5.0)
                status_data = json.loads(status_response)
                
                method_results["get_status"] = {
                    "valid": "result" in status_data,
                    "response": status_data
                }
                
                # Test get_cameras method
                cameras_request = {
                    "jsonrpc": "2.0",
                    "method": "get_cameras",
                    "params": {},
                    "id": 2
                }
                await websocket.send(json.dumps(cameras_request))
                cameras_response = await asyncio.wait_for(websocket.recv(), timeout=5.0)
                cameras_data = json.loads(cameras_response)
                
                method_results["get_cameras"] = {
                    "valid": "result" in cameras_data,
                    "response": cameras_data
                }
                
                # Test get_streams method
                streams_request = {
                    "jsonrpc": "2.0",
                    "method": "get_streams",
                    "params": {},
                    "id": 3
                }
                await websocket.send(json.dumps(streams_request))
                streams_response = await asyncio.wait_for(websocket.recv(), timeout=5.0)
                streams_data = json.loads(streams_response)
                
                method_results["get_streams"] = {
                    "valid": "result" in streams_data,
                    "response": streams_data
                }
                
                return method_results
                
        except Exception as e:
            self.system_issues.append(f"Method contracts validation failed: {str(e)}")
            raise
    
    async def validate_error_contracts(self) -> Dict[str, Any]:
        """Validate error handling contracts."""
        try:
            async with websockets.connect(self.websocket_url) as websocket:
                error_results = {}
                
                # Test method not found
                invalid_method_request = {
                    "jsonrpc": "2.0",
                    "method": "invalid_method",
                    "params": {},
                    "id": 1
                }
                await websocket.send(json.dumps(invalid_method_request))
                invalid_method_response = await asyncio.wait_for(websocket.recv(), timeout=5.0)
                invalid_method_data = json.loads(invalid_method_response)
                
                error_results["method_not_found"] = {
                    "valid": "error" in invalid_method_data,
                    "code": invalid_method_data.get("error", {}).get("code"),
                    "message": invalid_method_data.get("error", {}).get("message"),
                    "response": invalid_method_data
                }
                
                # Test invalid parameters
                invalid_params_request = {
                    "jsonrpc": "2.0",
                    "method": "get_camera_status",
                    "params": {"invalid_param": "value"},
                    "id": 2
                }
                await websocket.send(json.dumps(invalid_params_request))
                invalid_params_response = await asyncio.wait_for(websocket.recv(), timeout=5.0)
                invalid_params_data = json.loads(invalid_params_response)
                
                error_results["invalid_params"] = {
                    "valid": "error" in invalid_params_data,
                    "code": invalid_params_data.get("error", {}).get("code"),
                    "message": invalid_params_data.get("error", {}).get("message"),
                    "response": invalid_params_data
                }
                
                # Test invalid JSON
                await websocket.send("invalid json")
                invalid_json_response = await asyncio.wait_for(websocket.recv(), timeout=5.0)
                invalid_json_data = json.loads(invalid_json_response)
                
                error_results["invalid_json"] = {
                    "valid": "error" in invalid_json_data,
                    "code": invalid_json_data.get("error", {}).get("code"),
                    "message": invalid_json_data.get("error", {}).get("message"),
                    "response": invalid_json_data
                }
                
                return error_results
                
        except Exception as e:
            self.system_issues.append(f"Error contracts validation failed: {str(e)}")
            raise
    
    async def validate_data_structure_contracts(self) -> Dict[str, Any]:
        """Validate data structure contracts from design specifications."""
        try:
            async with websockets.connect(self.websocket_url) as websocket:
                # Test camera status response structure
                camera_status_request = {
                    "jsonrpc": "2.0",
                    "method": "get_camera_status",
                    "params": {"device": "/dev/video0"},
                    "id": 1
                }
                await websocket.send(json.dumps(camera_status_request))
                camera_status_response = await asyncio.wait_for(websocket.recv(), timeout=5.0)
                camera_status_data = json.loads(camera_status_response)
                
                # Validate camera status response structure
                if "result" in camera_status_data:
                    result = camera_status_data["result"]
                    structure_valid = all([
                        "device" in result,
                        "status" in result,
                        "name" in result,
                        "resolution" in result,
                        "fps" in result,
                        "streams" in result
                    ])
                    
                    # Validate status values
                    status_valid = result["status"] in ["CONNECTED", "DISCONNECTED", "ERROR"]
                    
                    # Validate streams structure
                    streams_valid = isinstance(result["streams"], dict) and all([
                        "rtsp" in result["streams"],
                        "webrtc" in result["streams"],
                        "hls" in result["streams"]
                    ])
                    
                    return {
                        "structure_valid": structure_valid,
                        "status_valid": status_valid,
                        "streams_valid": streams_valid,
                        "response": camera_status_data
                    }
                else:
                    return {
                        "structure_valid": False,
                        "status_valid": False,
                        "streams_valid": False,
                        "response": camera_status_data
                    }
                
        except Exception as e:
            self.system_issues.append(f"Data structure contracts validation failed: {str(e)}")
            raise
    
    async def run_comprehensive_contract_validation(self) -> Dict[str, Any]:
        """Run comprehensive API contract validation."""
        try:
            await self.setup_real_environment()
            
            # Start servers
            await self.websocket_server.start()
            await self.mediamtx_controller.start()
            await asyncio.sleep(2)
            
            # Execute all validation steps
            results = {
                "json_rpc_contract": await self.validate_json_rpc_contract(),
                "method_contracts": await self.validate_method_contracts(),
                "error_contracts": await self.validate_error_contracts(),
                "data_structure_contracts": await self.validate_data_structure_contracts(),
                "system_issues": self.system_issues
            }
            
            self.test_results = results
            return results
            
        finally:
            await self.cleanup_real_environment()


@pytest.mark.integration
@pytest.mark.asyncio
class TestAPIContracts:
    """Contract tests for API interface validation."""
    
    def setup_method(self):
        """Set up validator for each test method."""
        self.validator = APIContractValidator()
    
    async def teardown_method(self):
        """Clean up after each test method."""
        if hasattr(self, 'validator'):
            await self.validator.cleanup_real_environment()
    
    async def test_json_rpc_contract_compliance(self):
        """Test JSON-RPC 2.0 contract compliance."""
        await self.validator.setup_real_environment()
        
        try:
            # Start servers
            await self.validator.websocket_server.start()
            await self.validator.mediamtx_controller.start()
            await asyncio.sleep(2)
            
            result = await self.validator.validate_json_rpc_contract()
            
            # Validate results
            assert result["jsonrpc_version_valid"] is True, "JSON-RPC version invalid"
            assert result["id_field_valid"] is True, "ID field invalid"
            assert result["result_error_valid"] is True, "Result/error field invalid"
            assert result["error_handling_valid"] is True, "Error handling invalid"
            
            print(f"✅ JSON-RPC contract validation: {result}")
            
        finally:
            await self.validator.cleanup_real_environment()
    
    async def test_method_contracts_validation(self):
        """Test method contracts validation."""
        await self.validator.setup_real_environment()
        
        try:
            # Start servers
            await self.validator.websocket_server.start()
            await self.validator.mediamtx_controller.start()
            await asyncio.sleep(2)
            
            result = await self.validator.validate_method_contracts()
            
            # Validate results
            assert result["get_status"]["valid"] is True, "get_status method invalid"
            assert result["get_cameras"]["valid"] is True, "get_cameras method invalid"
            assert result["get_streams"]["valid"] is True, "get_streams method invalid"
            
            print(f"✅ Method contracts validation: {result}")
            
        finally:
            await self.validator.cleanup_real_environment()
    
    async def test_error_contracts_validation(self):
        """Test error handling contracts validation."""
        await self.validator.setup_real_environment()
        
        try:
            # Start servers
            await self.validator.websocket_server.start()
            await self.validator.mediamtx_controller.start()
            await asyncio.sleep(2)
            
            result = await self.validator.validate_error_contracts()
            
            # Validate results
            assert result["method_not_found"]["valid"] is True, "Method not found error invalid"
            assert result["invalid_json"]["valid"] is True, "Invalid JSON error invalid"
            
            print(f"✅ Error contracts validation: {result}")
            
        finally:
            await self.validator.cleanup_real_environment()
    
    async def test_data_structure_contracts_validation(self):
        """Test data structure contracts validation."""
        await self.validator.setup_real_environment()
        
        try:
            # Start servers
            await self.validator.websocket_server.start()
            await self.validator.mediamtx_controller.start()
            await asyncio.sleep(2)
            
            result = await self.validator.validate_data_structure_contracts()
            
            # Validate results
            assert result["structure_valid"] is True, "Data structure invalid"
            assert result["status_valid"] is True, "Status values invalid"
            assert result["streams_valid"] is True, "Streams structure invalid"
            
            print(f"✅ Data structure contracts validation: {result}")
            
        finally:
            await self.validator.cleanup_real_environment()
    
    async def test_comprehensive_contract_validation(self):
        """Test comprehensive API contract validation."""
        result = await self.validator.run_comprehensive_contract_validation()
        
        # Validate comprehensive results
        assert len(result["system_issues"]) == 0, f"System issues found: {result['system_issues']}"
        assert result["json_rpc_contract"]["jsonrpc_version_valid"] is True, "Comprehensive JSON-RPC failed"
        assert result["method_contracts"]["get_status"]["valid"] is True, "Comprehensive methods failed"
        assert result["error_contracts"]["method_not_found"]["valid"] is True, "Comprehensive errors failed"
        assert result["data_structure_contracts"]["structure_valid"] is True, "Comprehensive data structures failed"
        
        print(f"✅ Comprehensive contract validation: {result}")
        
        # Log results for evidence
        with open("/tmp/ivv_api_contracts_results.json", "w") as f:
            json.dump(result, f, indent=2, default=str)
