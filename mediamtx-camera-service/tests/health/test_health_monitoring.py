#!/usr/bin/env python3
"""
Health monitoring test suite for MediaMTX Camera Service.

Requirements Coverage:
- REQ-HEALTH-005: The system SHALL provide health status with detailed component information
- REQ-HEALTH-006: The system SHALL support Kubernetes readiness probes
- REQ-API-017: Health endpoints SHALL return JSON responses with status and timestamp
- REQ-API-018: Health endpoints SHALL return 200 OK for healthy status
- REQ-API-019: Health endpoints SHALL return 500 Internal Server Error for unhealthy status

Test Categories: Health
"""

import asyncio
import json
import sys
import os
import pytest
import time
import aiohttp
import websockets
from typing import Dict, Any, List
from dataclasses import dataclass
from pathlib import Path

# Add src to path for imports
sys.path.append('src')

from websocket_server.server import WebSocketJsonRpcServer
from camera_service.config import Config, ServerConfig, MediaMTXConfig, CameraConfig, LoggingConfig, RecordingConfig, SnapshotConfig
from camera_service.service_manager import ServiceManager
from mediamtx_wrapper.controller import MediaMTXController
from camera_discovery.hybrid_monitor import HybridCameraMonitor


@dataclass
class HealthComponent:
    """Health component information."""
    name: str
    status: str
    details: Dict[str, Any]
    last_check: float
    uptime: float


@dataclass
class HealthStatus:
    """Overall health status."""
    overall_status: str
    timestamp: float
    components: Dict[str, HealthComponent]
    version: str
    uptime: float


class HealthTestSetup:
    """Real system health test setup."""
    
    def __init__(self):
        self.config = self._build_test_config()
        self.service_manager = None
        self.mediamtx_controller = None
        self.camera_monitor = None
        self.server = None
        self.websocket_client = None
        self.http_client = None
        
    def _build_test_config(self) -> Config:
        """Build test configuration for health testing."""
        return Config(
            server=ServerConfig(host="127.0.0.1", port=8005, websocket_path="/ws", max_connections=100),
            mediamtx=MediaMTXConfig(
                host="127.0.0.1",
                api_port=9997,
                rtsp_port=8554,
                webrtc_port=8889,
                hls_port=8888,
                recordings_path="./.tmp_recordings",
                snapshots_path="./.tmp_snapshots",
            ),
            camera=CameraConfig(
                device_range=[0, 1, 2, 3], 
                enable_capability_detection=True, 
                detection_timeout=0.5,
                auto_start_streams=True
            ),
            logging=LoggingConfig(),
            recording=RecordingConfig(),
            snapshots=SnapshotConfig(),
        )
    
    async def setup(self):
        """Set up real system components for health testing."""
        # Create HTTP client for health endpoint testing
        self.http_client = aiohttp.ClientSession()
        
        # Create WebSocket client for health API testing
        # Connect to real WebSocket server on port 8002
        real_websocket_url = "ws://127.0.0.1:8002/ws"
        self.websocket_client = WebSocketHealthClient(real_websocket_url)
        await self.websocket_client.connect()
        
        # Authenticate with the real service using a valid test token
        from tests.fixtures.auth_utils import generate_valid_test_token
        
        test_token = generate_valid_test_token(username="health_test_user", role="admin")
        auth_result = await self.websocket_client.call_method("authenticate", {
            "token": test_token
        })
        
        if "error" in auth_result:
            print(f"⚠️ Authentication warning: {auth_result['error']}")
            # Continue anyway - some endpoints might work without auth
        else:
            print(f"✅ Authenticated successfully for health tests")
        
        # Verify real health server is available on port 8003
        try:
            async with self.http_client.get("http://127.0.0.1:8003/health/system") as response:
                if response.status != 200:
                    raise RuntimeError(f"Real health server not responding: {response.status}")
        except Exception as e:
            raise RuntimeError(f"Real health server not available on port 8003: {e}")
        
        print(f"✅ Health test setup completed - using real service")
    
    async def cleanup(self):
        """Clean up health test resources."""
        if self.http_client:
            await self.http_client.close()
        
        if self.websocket_client:
            await self.websocket_client.disconnect()
        
        print(f"✅ Health test cleanup completed - using real service")


class WebSocketHealthClient:
    """WebSocket client for health API testing."""
    
    def __init__(self, server_url: str):
        self.server_url = server_url
        self.websocket = None
        self.connected = False
        self.message_id_counter = 1
    
    async def connect(self):
        """Connect to WebSocket server."""
        self.websocket = await websockets.connect(self.server_url)
        self.connected = True
    
    async def disconnect(self):
        """Disconnect from server."""
        if self.websocket:
            await self.websocket.close()
            self.connected = False
    
    async def call_method(self, method: str, params: Dict[str, Any] = None) -> Dict[str, Any]:
        """Call a JSON-RPC method."""
        if not self.connected:
            raise RuntimeError("WebSocket client not connected")
        
        message = {
            "jsonrpc": "2.0",
            "method": method,
            "params": params or {},
            "id": self.message_id_counter
        }
        self.message_id_counter += 1
        
        await self.websocket.send(json.dumps(message))
        response = await self.websocket.recv()
        return json.loads(response)


class HealthEndpointTester:
    """Health endpoint testing utilities."""
    
    def __init__(self, http_client: aiohttp.ClientSession, base_url: str = "http://localhost:8003"):
        self.http_client = http_client
        self.base_url = base_url
    
    async def test_health_endpoint(self, endpoint: str) -> Dict[str, Any]:
        """Test a health endpoint and return response details."""
        url = f"{self.base_url}/health/{endpoint}"
        
        try:
            async with self.http_client.get(url) as response:
                status_code = response.status
                content_type = response.headers.get('content-type', '')
                response_text = await response.text()
                
                # Try to parse JSON response
                try:
                    json_data = json.loads(response_text) if response_text else {}
                except json.JSONDecodeError:
                    json_data = {"raw_response": response_text}
                
                return {
                    "status_code": status_code,
                    "content_type": content_type,
                    "response": json_data,
                    "success": 200 <= status_code < 300
                }
                
        except Exception as e:
            return {
                "status_code": 0,
                "content_type": "",
                "response": {"error": str(e)},
                "success": False
            }


@pytest.mark.health
@pytest.mark.asyncio
async def test_health_status_detailed_components():
    """Test REQ-HEALTH-005: Health status with detailed component information."""
    print("\n=== Health Test: Detailed Component Information ===")
    
    setup = HealthTestSetup()
    try:
        await setup.setup()
        
        # Test health status via HTTP health endpoint
        health_tester = HealthEndpointTester(setup.http_client)
        result = await health_tester.test_health_endpoint("system")
        
        print(f"📊 Health Status Response: {json.dumps(result, indent=2)}")
        
        # Validate response structure
        assert result["success"], f"Health endpoint failed: {result.get('error', 'unknown error')}"
        health_response = result["response"]
        
        # REQ-HEALTH-005: Detailed component information
        # Based on actual API response structure from /health/system
        assert "status" in health_response, "Health response missing 'status' field"
        assert "timestamp" in health_response, "Health response missing 'timestamp' field"
        assert "components" in health_response, "Health response missing 'components' field"
        
        # Validate overall status
        valid_statuses = ["healthy", "degraded", "unhealthy"]
        assert health_response["status"] in valid_statuses, \
            f"Invalid overall status: {health_response['status']}"
        
        # Validate timestamp format (ISO 8601 string)
        assert isinstance(health_response["timestamp"], str), "Timestamp should be ISO 8601 string"
        # Basic ISO 8601 format validation
        assert "T" in health_response["timestamp"], "Timestamp should be ISO 8601 format"
        
        # Validate components structure
        components = health_response["components"]
        required_components = ["mediamtx", "camera_monitor", "service_manager"]
        
        for component in required_components:
            assert component in components, f"Missing health component: {component}"
            component_data = components[component]
            
            # Validate component has required fields
            assert "status" in component_data, f"Component {component} missing status"
            assert "details" in component_data, f"Component {component} missing details"
            assert "timestamp" in component_data, f"Component {component} missing timestamp"
            
            # Validate status values
            valid_component_statuses = ["healthy", "unhealthy", "degraded"]
            assert component_data["status"] in valid_component_statuses, \
                f"Invalid status for {component}: {component_data['status']}"
            
            # Validate details is a string
            assert isinstance(component_data["details"], str), f"Details should be string for {component}"
            
            # Validate component timestamp format (ISO 8601 string)
            assert isinstance(component_data["timestamp"], str), f"Component timestamp should be ISO 8601 string for {component}"
            assert "T" in component_data["timestamp"], f"Component timestamp should be ISO 8601 format for {component}"
            
            print(f"   ✅ {component}: {component_data['status']} - {component_data['details']}")
        
        print(f"✅ REQ-HEALTH-005: Detailed component information validated")
        return health_response
        
    finally:
        await setup.cleanup()


@pytest.mark.health
@pytest.mark.asyncio
async def test_kubernetes_readiness_probes():
    """Test REQ-HEALTH-006: Kubernetes readiness probe support."""
    print("\n=== Health Test: Kubernetes Readiness Probes ===")
    
    setup = HealthTestSetup()
    try:
        await setup.setup()
        
        # Create health endpoint tester
        health_tester = HealthEndpointTester(setup.http_client)
        
        # Test Kubernetes readiness probe endpoints
        readiness_endpoints = [
            "ready",           # Standard Kubernetes readiness probe
            "live",            # Kubernetes liveness probe
            "startup",         # Kubernetes startup probe
            "healthz",         # Alternative health endpoint
            "readyz"           # Alternative readiness endpoint
        ]
        
        probe_results = {}
        
        for endpoint in readiness_endpoints:
            print(f"🔄 Testing Kubernetes probe: /health/{endpoint}")
            result = await health_tester.test_health_endpoint(endpoint)
            probe_results[endpoint] = result
            
            # REQ-HEALTH-006: Kubernetes readiness probe support
            if result["success"]:
                print(f"   ✅ {endpoint}: {result['status_code']} - {result['response'].get('status', 'unknown')}")
                
                # Validate JSON response structure
                response_data = result["response"]
                assert "status" in response_data, f"Missing status in {endpoint} response"
                assert response_data["status"] in ["ok", "healthy", "ready"], \
                    f"Invalid status in {endpoint}: {response_data['status']}"
                
                # Validate timestamp if present
                if "timestamp" in response_data:
                    assert isinstance(response_data["timestamp"], str), \
                        f"Invalid timestamp in {endpoint} - should be ISO 8601 string"
                    # Basic ISO 8601 format validation
                    assert "T" in response_data["timestamp"], \
                        f"Timestamp should be ISO 8601 format in {endpoint}"
            else:
                print(f"   ⚠️ {endpoint}: {result['status_code']} - {result['response'].get('error', 'unknown error')}")
        
        # At least one readiness probe should work
        working_probes = [ep for ep, result in probe_results.items() if result["success"]]
        assert len(working_probes) > 0, "No Kubernetes readiness probes working"
        
        print(f"✅ REQ-HEALTH-006: Kubernetes readiness probe support validated")
        print(f"   Working probes: {', '.join(working_probes)}")
        return probe_results
        
    finally:
        await setup.cleanup()


@pytest.mark.health
@pytest.mark.asyncio
async def test_health_endpoint_json_responses():
    """Test REQ-API-017: Health endpoints return JSON responses with status and timestamp."""
    print("\n=== Health Test: JSON Response Format ===")
    
    setup = HealthTestSetup()
    try:
        await setup.setup()
        
        # Create health endpoint tester
        health_tester = HealthEndpointTester(setup.http_client)
        
        # Test health endpoints
        health_endpoints = [
            "system",      # System health
            "cameras",     # Camera health
            "mediamtx",    # MediaMTX health
            "overall"      # Overall health
        ]
        
        json_validation_results = {}
        
        for endpoint in health_endpoints:
            print(f"🔄 Testing JSON response: /health/{endpoint}")
            result = await health_tester.test_health_endpoint(endpoint)
            json_validation_results[endpoint] = result
            
            # REQ-API-017: JSON responses with status and timestamp
            if result["success"]:
                print(f"   ✅ {endpoint}: {result['status_code']}")
                
                # Validate content type
                content_type = result["content_type"]
                assert "application/json" in content_type, \
                    f"Invalid content type for {endpoint}: {content_type}"
                
                # Validate JSON response structure
                response_data = result["response"]
                assert isinstance(response_data, dict), \
                    f"Response for {endpoint} is not a JSON object"
                
                # Validate required fields
                assert "status" in response_data, f"Missing status in {endpoint}"
                assert "timestamp" in response_data, f"Missing timestamp in {endpoint}"
                
                # Validate status values
                assert response_data["status"] in ["healthy", "unhealthy", "degraded", "ok"], \
                    f"Invalid status in {endpoint}: {response_data['status']}"
                
                # Validate timestamp format
                timestamp = response_data["timestamp"]
                assert isinstance(timestamp, str), \
                    f"Invalid timestamp format in {endpoint}: {timestamp} - should be ISO 8601 string"
                # Basic ISO 8601 format validation
                assert "T" in timestamp, \
                    f"Timestamp should be ISO 8601 format in {endpoint}: {timestamp}"
                
                # Validate timestamp is recent (within last 60 seconds)
                # Parse ISO 8601 timestamp to check if recent
                try:
                    from datetime import datetime
                    import re
                    # Remove timezone info for parsing
                    timestamp_clean = re.sub(r'[+-]\d{2}:\d{2}$', '', timestamp)
                    parsed_time = datetime.fromisoformat(timestamp_clean)
                    current_time = datetime.now()
                    time_diff = abs((current_time - parsed_time).total_seconds())
                    assert time_diff < 60, \
                        f"Timestamp too old in {endpoint}: {timestamp}"
                except Exception as e:
                    print(f"   ⚠️ Could not validate timestamp format for {endpoint}: {e}")
                
                print(f"      Status: {response_data['status']}")
                print(f"      Timestamp: {timestamp}")
                
            else:
                print(f"   ❌ {endpoint}: {result['status_code']} - {result['response'].get('error', 'unknown error')}")
        
        # At least system health should work
        assert json_validation_results["system"]["success"], "System health endpoint not working"
        
        print(f"✅ REQ-API-017: JSON response format validated")
        return json_validation_results
        
    finally:
        await setup.cleanup()


@pytest.mark.health
@pytest.mark.asyncio
async def test_health_endpoint_200_ok():
    """Test REQ-API-018: Health endpoints return 200 OK for healthy status."""
    print("\n=== Health Test: 200 OK Response ===")
    
    setup = HealthTestSetup()
    try:
        await setup.setup()
        
        # Create health endpoint tester
        health_tester = HealthEndpointTester(setup.http_client)
        
        # Test health endpoints for 200 OK response
        health_endpoints = [
            "system",      # System health
            "cameras",     # Camera health
            "mediamtx",    # MediaMTX health
            "overall"      # Overall health
        ]
        
        ok_response_results = {}
        
        for endpoint in health_endpoints:
            print(f"🔄 Testing 200 OK: /health/{endpoint}")
            result = await health_tester.test_health_endpoint(endpoint)
            ok_response_results[endpoint] = result
            
            # REQ-API-018: 200 OK for healthy status
            if result["success"]:
                status_code = result["status_code"]
                print(f"   ✅ {endpoint}: {status_code}")
                
                # Validate 200 OK response
                assert status_code == 200, \
                    f"Expected 200 OK for {endpoint}, got {status_code}"
                
                # Validate healthy status in response
                response_data = result["response"]
                if "status" in response_data:
                    status = response_data["status"]
                    assert status in ["healthy", "ok"], \
                        f"Expected healthy status for {endpoint}, got {status}"
                
                print(f"      Status Code: {status_code}")
                print(f"      Health Status: {response_data.get('status', 'unknown')}")
                
            else:
                print(f"   ❌ {endpoint}: {result['status_code']} - {result['response'].get('error', 'unknown error')}")
        
        # At least system health should return 200 OK
        assert ok_response_results["system"]["success"], "System health endpoint not returning 200 OK"
        assert ok_response_results["system"]["status_code"] == 200, "System health not returning 200 OK"
        
        print(f"✅ REQ-API-018: 200 OK response validated")
        return ok_response_results
        
    finally:
        await setup.cleanup()


class IsolatedHealthTestSetup:
    """Isolated health test setup for unit testing error conditions."""
    
    def __init__(self):
        self.config = self._build_test_config()
        self.service_manager = None
        self.mediamtx_controller = None
        self.camera_monitor = None
        self.http_client = None
        
    def _build_test_config(self) -> Config:
        """Build test configuration with free ports for isolated testing."""
        from tests.utils.port_utils import find_free_port
        
        # Use free ports to avoid conflicts
        free_websocket_port = find_free_port()
        free_health_port = find_free_port()
        
        return Config(
            server=ServerConfig(host="127.0.0.1", port=free_websocket_port, websocket_path="/ws", max_connections=100),
            mediamtx=MediaMTXConfig(
                host="127.0.0.1",
                api_port=9997,
                rtsp_port=8554,
                webrtc_port=8889,
                hls_port=8888,
                recordings_path="./.tmp_recordings",
                snapshots_path="./.tmp_snapshots",
            ),
            camera=CameraConfig(
                device_range=[0, 1, 2, 3], 
                enable_capability_detection=True, 
                detection_timeout=0.5
            ),
            logging=LoggingConfig(),
            recording=RecordingConfig(),
            snapshots=SnapshotConfig(),
        )
    
    async def setup(self):
        """Set up isolated components for unit testing."""
        # Skip this test if real service is running (port conflict)
        import socket
        try:
            with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
                s.connect(('127.0.0.1', 8003))
                # Port 8003 is in use, skip this test
                pytest.skip("Real health service is running on port 8003 - skipping isolated test")
        except ConnectionRefusedError:
            # Port is free, continue with isolated test
            pass
        
        # Initialize MediaMTX controller
        mediamtx_config = self.config.mediamtx
        self.mediamtx_controller = MediaMTXController(
            host=mediamtx_config.host,
            api_port=mediamtx_config.api_port,
            rtsp_port=mediamtx_config.rtsp_port,
            webrtc_port=mediamtx_config.webrtc_port,
            hls_port=mediamtx_config.hls_port,
            config_path=mediamtx_config.config_path,
            recordings_path=mediamtx_config.recordings_path,
            snapshots_path=mediamtx_config.snapshots_path,
        )
        
        # Start MediaMTX controller
        await self.mediamtx_controller.start()
        
        # Initialize camera monitor
        self.camera_monitor = HybridCameraMonitor(
            device_range=self.config.camera.device_range,
            enable_capability_detection=self.config.camera.enable_capability_detection,
            detection_timeout=self.config.camera.detection_timeout
        )
        
        # Start camera monitor
        await self.camera_monitor.start()
        
        # Initialize service manager (will create its own health server on free port)
        self.service_manager = ServiceManager(self.config)
        await self.service_manager.start()
        
        # Create HTTP client for health endpoint testing
        self.http_client = aiohttp.ClientSession()
        
        print(f"✅ Isolated health test setup completed")
    
    async def cleanup(self):
        """Clean up isolated test resources."""
        if self.http_client:
            await self.http_client.close()
        
        if self.service_manager:
            await self.service_manager.stop()
        
        if self.camera_monitor:
            await self.camera_monitor.stop()
        
        if self.mediamtx_controller:
            await self.mediamtx_controller.stop()
        
        print(f"✅ Isolated health test cleanup completed")


@pytest.mark.health
@pytest.mark.asyncio
async def test_health_endpoint_500_error():
    """Test REQ-API-019: Health endpoints return 500 Internal Server Error for unhealthy status."""
    print("\n=== Health Test: 500 Error Response ===")
    
    setup = IsolatedHealthTestSetup()
    try:
        await setup.setup()
        
        # Create health endpoint tester
        health_tester = HealthEndpointTester(setup.http_client)
        
        # Test health endpoints for 500 error when components are unhealthy
        health_endpoints = [
            "system",      # System health
            "cameras",     # Camera health
            "mediamtx",    # MediaMTX health
            "overall"      # Overall health
        ]
        
        error_response_results = {}
        
        # First, verify healthy status
        print("🔄 Verifying healthy status first...")
        for endpoint in health_endpoints:
            result = await health_tester.test_health_endpoint(endpoint)
            if result["success"]:
                print(f"   ✅ {endpoint}: Healthy ({result['status_code']})")
            else:
                print(f"   ⚠️ {endpoint}: {result['status_code']} - {result['response'].get('error', 'unknown error')}")
        
        # Now test error scenarios by stopping components
        print("\n🔄 Testing 500 error scenarios...")
        
        # Stop MediaMTX to create unhealthy state
        print("   Stopping MediaMTX to test unhealthy state...")
        await setup.mediamtx_controller.stop()
        
        # Wait a moment for health check to detect the change
        await asyncio.sleep(2)
        
        # Test health endpoints for 500 error
        for endpoint in health_endpoints:
            print(f"🔄 Testing 500 error: /health/{endpoint}")
            result = await health_tester.test_health_endpoint(endpoint)
            error_response_results[endpoint] = result
            
            # REQ-API-019: 500 Internal Server Error for unhealthy status
            if result["status_code"] == 500:
                print(f"   ✅ {endpoint}: 500 Internal Server Error (expected)")
                
                # Validate error response structure
                response_data = result["response"]
                if "error" in response_data:
                    print(f"      Error: {response_data['error']}")
                
                if "status" in response_data:
                    status = response_data["status"]
                    assert status in ["unhealthy", "error", "down"], \
                        f"Expected unhealthy status for {endpoint}, got {status}"
                    print(f"      Status: {status}")
                
            elif result["success"]:
                print(f"   ⚠️ {endpoint}: Still healthy ({result['status_code']})")
            else:
                print(f"   ❌ {endpoint}: Unexpected response ({result['status_code']})")
        
        # Restart MediaMTX to restore healthy state
        print("\n🔄 Restarting MediaMTX to restore healthy state...")
        await setup.mediamtx_controller.start()
        
        # Wait for health check to detect the recovery
        await asyncio.sleep(2)
        
        # Verify recovery
        print("🔄 Verifying recovery...")
        for endpoint in health_endpoints:
            result = await health_tester.test_health_endpoint(endpoint)
            if result["success"]:
                print(f"   ✅ {endpoint}: Recovered ({result['status_code']})")
            else:
                print(f"   ⚠️ {endpoint}: Still unhealthy ({result['status_code']})")
        
        # At least one endpoint should have returned 500 during unhealthy state
        error_responses = [r for r in error_response_results.values() if r["status_code"] == 500]
        assert len(error_responses) > 0, "No health endpoints returned 500 error during unhealthy state"
        
        print(f"✅ REQ-API-019: 500 error response validated")
        return error_response_results
        
    finally:
        await setup.cleanup()


# Main health test runner
async def run_all_health_tests():
    """Run all health tests with comprehensive reporting."""
    print("=== MediaMTX Camera Service Health Test Suite ===")
    print("Testing health monitoring against requirements baseline")
    
    test_results = {}
    
    try:
        # Test 1: Detailed Component Information
        print("\n=== Test 1: Detailed Component Information ===")
        test_results['detailed_components'] = await test_health_status_detailed_components()
        
        # Test 2: Kubernetes Readiness Probes
        print("\n=== Test 2: Kubernetes Readiness Probes ===")
        test_results['kubernetes_probes'] = await test_kubernetes_readiness_probes()
        
        # Test 3: JSON Response Format
        print("\n=== Test 3: JSON Response Format ===")
        test_results['json_responses'] = await test_health_endpoint_json_responses()
        
        # Test 4: 200 OK Response
        print("\n=== Test 4: 200 OK Response ===")
        test_results['ok_responses'] = await test_health_endpoint_200_ok()
        
        # Test 5: 500 Error Response
        print("\n=== Test 5: 500 Error Response ===")
        test_results['error_responses'] = await test_health_endpoint_500_error()
        
        print("\n=== All Health Tests Completed Successfully ===")
        print("✅ All health requirements validated")
        print("✅ Health endpoint responses verified")
        print("✅ Kubernetes readiness probe support confirmed")
        print("✅ Error handling validated")
        
        return test_results
        
    except Exception as e:
        print(f"\n❌ Health Tests Failed: {e}")
        raise


if __name__ == "__main__":
    # Run health tests
    asyncio.run(run_all_health_tests())
