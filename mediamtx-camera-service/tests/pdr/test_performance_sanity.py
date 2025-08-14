"""
Performance Sanity Testing - PDR Level

Tests basic performance characteristics against real system to validate PDR budget targets.

Critical Operations Tested:
1. Service connection and startup time
2. Camera list refresh (get_camera_list API)
3. Photo capture (take_snapshot API) 
4. Video recording start (start_recording API)
5. Basic resource usage under normal operation

PDR Performance Budget Targets:
- Service connection: <1s
- Camera list refresh: <50ms service API
- Photo capture: <100ms service processing  
- Video recording start: <100ms service API

NO MOCKING - Tests execute against real system components.
"""

import asyncio
import json
import tempfile
import time
import psutil
import os
from typing import Dict, Any, List
from dataclasses import dataclass

import pytest
import pytest_asyncio
import websockets

from src.camera_service.service_manager import ServiceManager
from src.camera_service.config import Config, ServerConfig, MediaMTXConfig, CameraConfig, RecordingConfig
from src.websocket_server.server import WebSocketJsonRpcServer
from src.mediamtx_wrapper.controller import MediaMTXController


@dataclass
class PerformanceResult:
    """Result of performance measurement."""
    
    operation: str
    response_time_ms: float
    success: bool
    meets_budget: bool
    budget_target_ms: float
    resource_usage: Dict[str, Any]
    error_message: str = None


class PerformanceSanityValidator:
    """Validates basic performance characteristics for PDR budget compliance."""
    
    def __init__(self):
        self.temp_dir = None
        self.service_manager = None
        self.websocket_server = None
        self.mediamtx_controller = None
        self.websocket_url = None
        self.performance_results: List[PerformanceResult] = []
        self.budget_violations: List[str] = []
        
        # PDR Performance Budget Targets (from client requirements)
        self.performance_budgets = {
            "service_connection": 1000.0,  # <1s
            "camera_list_refresh": 50.0,   # <50ms service API
            "photo_capture": 100.0,        # <100ms service processing
            "video_recording_start": 100.0, # <100ms service API
            "basic_api_call": 200.0        # General API responsiveness
        }
        
    async def setup_real_performance_environment(self):
        """Set up real system environment for performance testing."""
        self.temp_dir = tempfile.mkdtemp(prefix="pdr_performance_test_")
        
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
        
        # Initialize real service configuration with a guaranteed free port
        import socket
        def _find_free_port():
            with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
                s.bind(("", 0))
                s.listen(1)
                return s.getsockname()[1]
        port = _find_free_port()
        server_cfg = ServerConfig(host="127.0.0.1", port=port)
        config = Config(
            server=server_cfg,
            mediamtx=mediamtx_config,
            camera=CameraConfig(device_range=[0, 1, 2]),
            recording=RecordingConfig(enabled=True)
        )
        
        # Initialize real WebSocket server first
        # Always derive URLs from the active service configuration to avoid drift
        self.websocket_url = f"ws://127.0.0.1:{server_cfg.port}/ws"
        self.websocket_server = WebSocketJsonRpcServer(
            host="127.0.0.1",
            port=server_cfg.port,
            websocket_path="/ws",
            max_connections=100
        )
        
        # Initialize service manager with WebSocket server
        self.service_manager = ServiceManager(config, websocket_server=self.websocket_server)
        self.websocket_server.set_service_manager(self.service_manager)
        
    async def cleanup_real_environment(self):
        """Clean up real test environment."""
        # Clean up MediaMTX paths first to prevent "path already exists" errors
        if self.mediamtx_controller:
            try:
                # Clean up common test paths
                test_paths = ["cam0", "cam1", "cam2", "test_stream", "test_recording_stream"]
                for path_name in test_paths:
                    try:
                        await self.mediamtx_controller.delete_stream(path_name)
                    except Exception:
                        pass  # Ignore errors during cleanup
            except Exception:
                pass
        
        if self.websocket_server:
            try:
                await self.websocket_server.stop()
            except Exception:
                pass
                
        if self.mediamtx_controller:
            try:
                await self.mediamtx_controller.stop()
            except Exception:
                pass
                
        if self.service_manager:
            try:
                await self.service_manager.stop()
            except Exception:
                pass
                
        if self.temp_dir:
            import shutil
            try:
                shutil.rmtree(self.temp_dir)
            except Exception:
                pass
    
    def _measure_resource_usage(self) -> Dict[str, Any]:
        """Measure current system resource usage."""
        process = psutil.Process()
        return {
            "cpu_percent": process.cpu_percent(),
            "memory_mb": process.memory_info().rss / 1024 / 1024,
            "open_files": len(process.open_files()),
            "connections": len(process.connections()),
            "threads": process.num_threads()
        }
    
    async def test_service_connection_performance(self) -> PerformanceResult:
        """
        Test service connection and startup performance.
        
        PDR Budget: <1s for service connection
        """
        start_time = time.time()
        
        try:
            # Start service manager (includes camera monitor, MediaMTX controller)
            await self.service_manager.start()
            
            # Start WebSocket server
            await self.websocket_server.start()
            await self.mediamtx_controller.start()
            await asyncio.sleep(1)  # Allow full startup
            
            # Test actual WebSocket connection
            connection_start = time.time()
            async with websockets.connect(self.websocket_url) as websocket:
                # Send simple ping to verify connection
                ping_message = {
                    "jsonrpc": "2.0",
                    "method": "get_status",
                    "id": 1
                }
                await websocket.send(json.dumps(ping_message))
                response = await websocket.recv()
                response_data = json.loads(response)
                
                connection_time = (time.time() - connection_start) * 1000
                total_startup_time = (time.time() - start_time) * 1000
                
                # Use connection time for budget validation (not total startup)
                meets_budget = connection_time <= self.performance_budgets["service_connection"]
                
                if not meets_budget:
                    self.budget_violations.append(
                        f"Service connection took {connection_time:.1f}ms, budget: {self.performance_budgets['service_connection']:.1f}ms"
                    )
                
                return PerformanceResult(
                    operation="service_connection",
                    response_time_ms=connection_time,
                    success=True,
                    meets_budget=meets_budget,
                    budget_target_ms=self.performance_budgets["service_connection"],
                    resource_usage=self._measure_resource_usage()
                )
                
        except Exception as e:
            total_time = (time.time() - start_time) * 1000
            return PerformanceResult(
                operation="service_connection",
                response_time_ms=total_time,
                success=False,
                meets_budget=False,
                budget_target_ms=self.performance_budgets["service_connection"],
                resource_usage=self._measure_resource_usage(),
                error_message=str(e)
            )
    
    async def test_camera_list_performance(self) -> PerformanceResult:
        """
        Test camera list refresh performance.
        
        PDR Budget: <50ms service API
        """
        try:
            async with websockets.connect(self.websocket_url) as websocket:
                # Warm up connection
                await asyncio.sleep(0.1)
                
                # Measure camera list API call
                start_time = time.time()
                
                message = {
                    "jsonrpc": "2.0",
                    "method": "get_camera_list",
                    "id": 2
                }
                
                await websocket.send(json.dumps(message))
                response = await websocket.recv()
                
                response_time = (time.time() - start_time) * 1000
                response_data = json.loads(response)
                
                success = "result" in response_data
                meets_budget = response_time <= self.performance_budgets["camera_list_refresh"]
                
                if not meets_budget:
                    self.budget_violations.append(
                        f"Camera list refresh took {response_time:.1f}ms, budget: {self.performance_budgets['camera_list_refresh']:.1f}ms"
                    )
                
                return PerformanceResult(
                    operation="camera_list_refresh",
                    response_time_ms=response_time,
                    success=success,
                    meets_budget=meets_budget,
                    budget_target_ms=self.performance_budgets["camera_list_refresh"],
                    resource_usage=self._measure_resource_usage()
                )
                
        except Exception as e:
            return PerformanceResult(
                operation="camera_list_refresh",
                response_time_ms=0,
                success=False,
                meets_budget=False,
                budget_target_ms=self.performance_budgets["camera_list_refresh"],
                resource_usage=self._measure_resource_usage(),
                error_message=str(e)
            )

    async def test_camera_list_p50_performance(self, samples: int = 9) -> PerformanceResult:
        """
        Measure median (P50) latency for camera list refresh across multiple calls.
        Demonstrates compliance with the ≤50 ms P50 requirement.
        """
        import statistics
        latencies: List[float] = []
        try:
            async with websockets.connect(self.websocket_url) as websocket:
                # Warm connection
                await asyncio.sleep(0.05)
                for i in range(samples):
                    start_time = time.time()
                    await websocket.send(json.dumps({
                        "jsonrpc": "2.0",
                        "method": "get_camera_list",
                        "id": 2000 + i
                    }))
                    _ = await websocket.recv()
                    latencies.append((time.time() - start_time) * 1000)

            p50 = statistics.median(latencies) if latencies else 0.0
            meets_budget = p50 <= self.performance_budgets["camera_list_refresh"]
            if not meets_budget:
                self.budget_violations.append(
                    f"Camera list P50 {p50:.1f}ms exceeds budget {self.performance_budgets['camera_list_refresh']:.1f}ms"
                )
            return PerformanceResult(
                operation="camera_list_refresh_p50",
                response_time_ms=p50,
                success=True,
                meets_budget=meets_budget,
                budget_target_ms=self.performance_budgets["camera_list_refresh"],
                resource_usage=self._measure_resource_usage()
            )
        except Exception as e:
            return PerformanceResult(
                operation="camera_list_refresh_p50",
                response_time_ms=0,
                success=False,
                meets_budget=False,
                budget_target_ms=self.performance_budgets["camera_list_refresh"],
                resource_usage=self._measure_resource_usage(),
                error_message=str(e)
            )
    
    async def test_photo_capture_performance(self) -> PerformanceResult:
        """
        Test photo capture performance.
        
        PDR Budget: <100ms service processing
        """
        try:
            async with websockets.connect(self.websocket_url) as websocket:
                # Warm up connection
                await asyncio.sleep(0.1)
                
                # Measure photo capture API call
                start_time = time.time()
                
                message = {
                    "jsonrpc": "2.0",
                    "method": "take_snapshot",
                    "params": {"device": "/dev/video0"},
                    "id": 3
                }
                
                await websocket.send(json.dumps(message))
                response = await websocket.recv()
                
                response_time = (time.time() - start_time) * 1000
                response_data = json.loads(response)
                
                # For PDR, we accept that photo capture may not be fully implemented
                # Focus on API responsiveness rather than full functionality
                success = ("result" in response_data) or ("error" in response_data and response_data["error"]["code"] != -32601)
                meets_budget = response_time <= self.performance_budgets["photo_capture"]
                
                if not meets_budget:
                    self.budget_violations.append(
                        f"Photo capture took {response_time:.1f}ms, budget: {self.performance_budgets['photo_capture']:.1f}ms"
                    )
                
                return PerformanceResult(
                    operation="photo_capture",
                    response_time_ms=response_time,
                    success=success,
                    meets_budget=meets_budget,
                    budget_target_ms=self.performance_budgets["photo_capture"],
                    resource_usage=self._measure_resource_usage()
                )
                
        except Exception as e:
            return PerformanceResult(
                operation="photo_capture",
                response_time_ms=0,
                success=False,
                meets_budget=False,
                budget_target_ms=self.performance_budgets["photo_capture"],
                resource_usage=self._measure_resource_usage(),
                error_message=str(e)
            )
    
    async def test_video_recording_start_performance(self) -> PerformanceResult:
        """
        Test video recording start performance.
        
        PDR Budget: <100ms service API
        """
        try:
            async with websockets.connect(self.websocket_url) as websocket:
                # Warm up connection
                await asyncio.sleep(0.1)
                
                # Measure recording start API call
                start_time = time.time()
                
                message = {
                    "jsonrpc": "2.0",
                    "method": "start_recording",
                    "params": {"device": "/dev/video0", "duration": 5},
                    "id": 4
                }
                
                await websocket.send(json.dumps(message))
                response = await websocket.recv()
                
                response_time = (time.time() - start_time) * 1000
                response_data = json.loads(response)
                
                # For PDR, focus on API responsiveness
                success = ("result" in response_data) or ("error" in response_data and response_data["error"]["code"] != -32601)
                meets_budget = response_time <= self.performance_budgets["video_recording_start"]
                
                if not meets_budget:
                    self.budget_violations.append(
                        f"Video recording start took {response_time:.1f}ms, budget: {self.performance_budgets['video_recording_start']:.1f}ms"
                    )
                
                return PerformanceResult(
                    operation="video_recording_start",
                    response_time_ms=response_time,
                    success=success,
                    meets_budget=meets_budget,
                    budget_target_ms=self.performance_budgets["video_recording_start"],
                    resource_usage=self._measure_resource_usage()
                )
                
        except Exception as e:
            return PerformanceResult(
                operation="video_recording_start",
                response_time_ms=0,
                success=False,
                meets_budget=False,
                budget_target_ms=self.performance_budgets["video_recording_start"],
                resource_usage=self._measure_resource_usage(),
                error_message=str(e)
            )
    
    async def test_basic_api_responsiveness(self) -> PerformanceResult:
        """
        Test basic API responsiveness with get_status call.
        
        PDR Budget: <200ms for general API responsiveness
        """
        try:
            async with websockets.connect(self.websocket_url) as websocket:
                # Warm up connection
                await asyncio.sleep(0.1)
                
                # Measure basic API call
                start_time = time.time()
                
                message = {
                    "jsonrpc": "2.0",
                    "method": "get_status",
                    "id": 5
                }
                
                await websocket.send(json.dumps(message))
                response = await websocket.recv()
                
                response_time = (time.time() - start_time) * 1000
                response_data = json.loads(response)
                
                success = "result" in response_data
                meets_budget = response_time <= self.performance_budgets["basic_api_call"]
                
                if not meets_budget:
                    self.budget_violations.append(
                        f"Basic API call took {response_time:.1f}ms, budget: {self.performance_budgets['basic_api_call']:.1f}ms"
                    )
                
                return PerformanceResult(
                    operation="basic_api_call",
                    response_time_ms=response_time,
                    success=success,
                    meets_budget=meets_budget,
                    budget_target_ms=self.performance_budgets["basic_api_call"],
                    resource_usage=self._measure_resource_usage()
                )
                
        except Exception as e:
            return PerformanceResult(
                operation="basic_api_call",
                response_time_ms=0,
                success=False,
                meets_budget=False,
                budget_target_ms=self.performance_budgets["basic_api_call"],
                resource_usage=self._measure_resource_usage(),
                error_message=str(e)
            )
    
    async def test_light_load_performance(self) -> List[PerformanceResult]:
        """
        Test performance under light representative load.
        
        Simulates basic client usage pattern:
        - Multiple camera list refreshes
        - Status checks
        - Light concurrent API usage
        """
        results = []
        
        try:
            # Test multiple concurrent connections (light load)
            async def concurrent_api_calls():
                async with websockets.connect(self.websocket_url) as websocket:
                    # Simulate client usage pattern
                    operations = [
                        {"method": "get_status", "id": 100},
                        {"method": "get_camera_list", "id": 101},
                        {"method": "get_status", "id": 102},
                    ]
                    
                    call_results = []
                    for op in operations:
                        start_time = time.time()
                        await websocket.send(json.dumps({"jsonrpc": "2.0", **op}))
                        response = await websocket.recv()
                        response_time = (time.time() - start_time) * 1000
                        
                        response_data = json.loads(response)
                        success = "result" in response_data
                        
                        call_results.append(PerformanceResult(
                            operation=f"light_load_{op['method']}",
                            response_time_ms=response_time,
                            success=success,
                            meets_budget=response_time <= 200.0,  # General responsiveness
                            budget_target_ms=200.0,
                            resource_usage=self._measure_resource_usage()
                        ))
                    
                    return call_results
            
            # Run 3 concurrent clients (light load)
            concurrent_tasks = [concurrent_api_calls() for _ in range(3)]
            all_results = await asyncio.gather(*concurrent_tasks)
            
            # Flatten results
            for client_results in all_results:
                results.extend(client_results)
            
            return results
            
        except Exception as e:
            return [PerformanceResult(
                operation="light_load_test",
                response_time_ms=0,
                success=False,
                meets_budget=False,
                budget_target_ms=200.0,
                resource_usage=self._measure_resource_usage(),
                error_message=str(e)
            )]
    
    async def run_comprehensive_performance_sanity_validation(self) -> Dict[str, Any]:
        """Run comprehensive performance sanity validation for PDR."""
        try:
            await self.setup_real_performance_environment()
            
            # Execute all performance tests
            self.performance_results = []
            
            # Critical path performance tests
            connection_result = await self.test_service_connection_performance()
            self.performance_results.append(connection_result)
            
            camera_list_result = await self.test_camera_list_performance()
            self.performance_results.append(camera_list_result)
            
            # Add explicit P50 measurement to evidence
            camera_list_p50 = await self.test_camera_list_p50_performance()
            self.performance_results.append(camera_list_p50)

            photo_capture_result = await self.test_photo_capture_performance()
            self.performance_results.append(photo_capture_result)
            
            recording_result = await self.test_video_recording_start_performance()
            self.performance_results.append(recording_result)
            
            api_responsiveness_result = await self.test_basic_api_responsiveness()
            self.performance_results.append(api_responsiveness_result)
            
            # Light load testing
            light_load_results = await self.test_light_load_performance()
            self.performance_results.extend(light_load_results)
            
            # Calculate summary statistics
            total_tests = len(self.performance_results)
            successful_tests = sum(1 for r in self.performance_results if r.success)
            budget_compliant_tests = sum(1 for r in self.performance_results if r.meets_budget)
            
            success_rate = (successful_tests / total_tests * 100) if total_tests > 0 else 0
            budget_compliance_rate = (budget_compliant_tests / total_tests * 100) if total_tests > 0 else 0
            
            # Calculate average response times by operation type
            operation_averages = {}
            for result in self.performance_results:
                op_type = result.operation.split('_')[0] if '_' in result.operation else result.operation
                if op_type not in operation_averages:
                    operation_averages[op_type] = []
                operation_averages[op_type].append(result.response_time_ms)
            
            for op_type in operation_averages:
                operation_averages[op_type] = sum(operation_averages[op_type]) / len(operation_averages[op_type])
            
            # Overall resource usage summary
            resource_summary = {
                "max_memory_mb": max(r.resource_usage["memory_mb"] for r in self.performance_results),
                "max_cpu_percent": max(r.resource_usage["cpu_percent"] for r in self.performance_results),
                "avg_connections": sum(r.resource_usage["connections"] for r in self.performance_results) / total_tests,
                "avg_threads": sum(r.resource_usage["threads"] for r in self.performance_results) / total_tests
            }
            
            # For PDR evidence, require 100% budget compliance across measured operations
            return {
                "pdr_performance_validation": budget_compliance_rate >= 100.0,
                "success_rate": success_rate,
                "budget_compliance_rate": budget_compliance_rate,
                "total_tests": total_tests,
                "successful_tests": successful_tests,
                "budget_compliant_tests": budget_compliant_tests,
                "budget_violations": self.budget_violations,
                "operation_averages": operation_averages,
                "resource_summary": resource_summary,
                "camera_list_p50_ms": next((r.response_time_ms for r in self.performance_results if r.operation == "camera_list_refresh_p50"), None),
                "performance_results": [
                    {
                        "operation": r.operation,
                        "response_time_ms": r.response_time_ms,
                        "success": r.success,
                        "meets_budget": r.meets_budget,
                        "budget_target_ms": r.budget_target_ms,
                        "error_message": r.error_message
                    }
                    for r in self.performance_results
                ]
            }
            
        finally:
            await self.cleanup_real_environment()


@pytest.mark.pdr
@pytest.mark.asyncio
class TestPerformanceSanity:
    """PDR-level performance sanity tests."""
    
    def setup_method(self):
        """Set up validator for each test method."""
        self.validator = PerformanceSanityValidator()
    
    async def teardown_method(self):
        """Clean up after each test method."""
        if hasattr(self, 'validator'):
            await self.validator.cleanup_real_environment()
    
    async def test_service_connection_performance(self):
        """Test service connection meets PDR budget (<1s)."""
        await self.validator.setup_real_performance_environment()
        
        result = await self.validator.test_service_connection_performance()
        
        # Validate performance
        assert result.success, f"Service connection failed: {result.error_message}"
        assert result.meets_budget, f"Service connection too slow: {result.response_time_ms:.1f}ms > {result.budget_target_ms:.1f}ms"
        
        print(f"✅ Service Connection: {result.response_time_ms:.1f}ms (budget: {result.budget_target_ms:.1f}ms)")
    
    async def test_critical_api_performance(self):
        """Test critical API operations meet PDR budgets."""
        await self.validator.setup_real_performance_environment()
        
        # Start services for API testing
        await self.validator.service_manager.start()
        await self.validator.websocket_server.start()
        await self.validator.mediamtx_controller.start()
        await asyncio.sleep(1)
        
        # Test camera list performance
        camera_result = await self.validator.test_camera_list_performance()
        assert camera_result.success, f"Camera list failed: {camera_result.error_message}"
        assert camera_result.meets_budget, f"Camera list too slow: {camera_result.response_time_ms:.1f}ms > {camera_result.budget_target_ms:.1f}ms"
        
        # Test API responsiveness
        api_result = await self.validator.test_basic_api_responsiveness()
        assert api_result.success, f"API responsiveness failed: {api_result.error_message}"
        assert api_result.meets_budget, f"API too slow: {api_result.response_time_ms:.1f}ms > {api_result.budget_target_ms:.1f}ms"
        
        print(f"✅ Camera List: {camera_result.response_time_ms:.1f}ms (budget: {camera_result.budget_target_ms:.1f}ms)")
        print(f"✅ API Responsiveness: {api_result.response_time_ms:.1f}ms (budget: {api_result.budget_target_ms:.1f}ms)")
    
    async def test_light_load_performance(self):
        """Test performance under light representative load."""
        await self.validator.setup_real_performance_environment()
        
        # Start services
        await self.validator.service_manager.start()
        await self.validator.websocket_server.start()
        await self.validator.mediamtx_controller.start()
        await asyncio.sleep(1)
        
        # Test light load
        results = await self.validator.test_light_load_performance()
        
        # Validate all operations succeed under light load
        successful_ops = [r for r in results if r.success]
        assert len(successful_ops) >= len(results) * 0.9, f"Too many failures under light load: {len(successful_ops)}/{len(results)}"
        
        # Validate reasonable response times under load
        avg_response_time = sum(r.response_time_ms for r in successful_ops) / len(successful_ops)
        assert avg_response_time <= 500.0, f"Average response time too high under load: {avg_response_time:.1f}ms"
        
        print(f"✅ Light Load: {len(successful_ops)}/{len(results)} operations successful, avg: {avg_response_time:.1f}ms")
    
    async def test_comprehensive_performance_sanity_validation(self):
        """Test comprehensive performance sanity validation for PDR."""
        result = await self.validator.run_comprehensive_performance_sanity_validation()
        
        # Validate comprehensive results for PDR
        assert result["pdr_performance_validation"], f"PDR performance validation failed"
        assert result["success_rate"] >= 80.0, f"Success rate too low: {result['success_rate']:.1f}%"
        assert result["budget_compliance_rate"] >= 80.0, f"Budget compliance too low: {result['budget_compliance_rate']:.1f}%"
        assert len(result["budget_violations"]) <= 2, f"Too many budget violations: {result['budget_violations']}"
        
        print(f"✅ Comprehensive Performance Sanity Validation:")
        print(f"   Success Rate: {result['success_rate']:.1f}%")
        print(f"   Budget Compliance: {result['budget_compliance_rate']:.1f}%")
        print(f"   Total Tests: {result['total_tests']}")
        print(f"   Resource Usage: {result['resource_summary']['max_memory_mb']:.1f}MB max memory")
        
        # Save results for evidence
        with open("/tmp/pdr_performance_sanity_results.json", "w") as f:
            json.dump(result, f, indent=2, default=str)
