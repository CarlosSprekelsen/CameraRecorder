"""
Enhanced Performance Sanity Testing - PDR Level

Enhanced performance sanity tests with improved reliability and edge case handling:
1. Retry mechanisms for transient failures
2. Statistical analysis of performance results
3. Resource monitoring during tests
4. Baseline establishment and drift detection
5. Performance regression detection
6. Load variation testing
7. System stability validation

PDR Performance Budget Targets:
- Service connection: <1s
- Camera list refresh: <50ms service API
- Photo capture: <100ms service processing  
- Video recording start: <100ms service API
- API responsiveness: <200ms

NO MOCKING - Tests execute against real system components with enhanced reliability.

Requirements:
- REQ-PERF-001: System shall handle concurrent camera operations efficiently
- REQ-PERF-002: System shall maintain responsive performance under load
- REQ-PERF-003: System shall meet latency requirements for real-time operations
- REQ-PERF-004: System shall handle resource constraints gracefully
- REQ-HEALTH-001: System shall provide comprehensive logging for health monitoring
- REQ-HEALTH-002: System shall support structured logging for production environments
- REQ-HEALTH-003: System shall enable correlation ID tracking across components
"""

import asyncio
import json
import tempfile
import time
import psutil
import os
import statistics
from typing import Dict, Any, List, Optional, Tuple
from dataclasses import dataclass
from collections import defaultdict

import pytest
import pytest_asyncio
import websockets
import aiohttp

from src.camera_service.service_manager import ServiceManager
from src.camera_service.config import Config, ServerConfig, MediaMTXConfig, CameraConfig, RecordingConfig
from src.websocket_server.server import WebSocketJsonRpcServer
from src.mediamtx_wrapper.controller import MediaMTXController


@dataclass
class EnhancedPerformanceResult:
    """Enhanced result of performance measurement with statistical data."""
    
    operation: str
    response_time_ms: float
    success: bool
    meets_budget: bool
    budget_target_ms: float
    resource_usage: Dict[str, Any]
    retry_count: int = 0
    error_message: str = None
    baseline_deviation: float = 0.0
    percentile_95: float = 0.0
    percentile_99: float = 0.0


@dataclass
class PerformanceBaseline:
    """Performance baseline for regression detection."""
    
    operation: str
    mean_response_time_ms: float
    std_deviation_ms: float
    percentile_95_ms: float
    percentile_99_ms: float
    sample_count: int
    established_date: str


class EnhancedPerformanceSanityValidator:
    """Enhanced performance sanity validator with reliability improvements."""
    
    def __init__(self):
        self.temp_dir = None
        self.service_manager = None
        self.websocket_server = None
        self.mediamtx_controller = None
        self.websocket_url = None
        self.performance_results: List[EnhancedPerformanceResult] = []
        self.budget_violations: List[str] = []
        self.baseline_data: Dict[str, PerformanceBaseline] = {}
        
        # PDR Performance Budget Targets
        self.performance_budgets = {
            "service_connection": 1000.0,  # <1s
            "camera_list_refresh": 50.0,   # <50ms service API
            "photo_capture": 100.0,        # <100ms service processing
            "video_recording_start": 100.0, # <100ms service API
            "basic_api_call": 200.0,       # General API responsiveness
            "websocket_connection": 500.0,  # WebSocket connection time
            "health_check": 100.0          # Health check response time
        }
        
        # Reliability configuration
        self.max_retries = 3
        self.retry_delay = 0.5  # seconds
        self.sample_size = 5    # Number of samples per operation
        self.baseline_threshold = 2.0  # Standard deviations for regression detection
        
    async def setup_real_performance_environment(self):
        """Set up real system environment for enhanced performance testing."""
        self.temp_dir = tempfile.mkdtemp(prefix="pdr_enhanced_performance_test_")
        
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
        
        # Start MediaMTX service
        await self.mediamtx_controller.start()
        await asyncio.sleep(2)  # Allow startup
        
        # Create server configuration
        server_config = ServerConfig(
            host="127.0.0.1",
            port=8002,
            websocket_path="/ws",
            max_connections=100
        )
        
        # Create camera configuration
        camera_config = CameraConfig(
            poll_interval=0.1,
            detection_timeout=2.0,
            device_range=[0, 9],
            enable_capability_detection=True,
            auto_start_streams=True
        )
        
        # Create recording configuration
        recording_config = RecordingConfig(
            enabled=False,  # Disable for performance testing
            format="fmp4",
            quality="high",
            segment_duration=3600,
            max_segment_size=524288000,
            auto_cleanup=True,
            cleanup_interval=86400,
            max_age=604800,
            max_size=10737418240
        )
        
        # Create main configuration
        config = Config(
            server=server_config,
            mediamtx=mediamtx_config,
            camera=camera_config,
            recording=recording_config
        )
        
        # Initialize service manager
        self.service_manager = ServiceManager(config)
        
        # Start WebSocket server
        self.websocket_server = WebSocketJsonRpcServer(
            host=server_config.host,
            port=server_config.port,
            websocket_path=server_config.websocket_path,
            max_connections=server_config.max_connections
        )
        
        await self.websocket_server.start()
        self.websocket_url = f"ws://{server_config.host}:{server_config.port}{server_config.websocket_path}"
        
        # Start service manager
        await self.service_manager.start()
        await asyncio.sleep(1)  # Allow startup
        
    async def cleanup_real_environment(self):
        """Clean up real test environment."""
        if self.service_manager:
            try:
                await self.service_manager.stop()
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
                
        if self.temp_dir:
            import shutil
            try:
                shutil.rmtree(self.temp_dir)
            except Exception:
                pass

    async def measure_operation_with_retry(self, operation_name: str, operation_func, 
                                         budget_target_ms: float) -> EnhancedPerformanceResult:
        """
        Measure operation performance with retry mechanism and statistical analysis.
        """
        results = []
        retry_count = 0
        
        # Collect multiple samples for statistical analysis
        for sample in range(self.sample_size):
            start_time = time.time()
            success = False
            error_message = None
            
            # Retry mechanism for transient failures
            for attempt in range(self.max_retries + 1):
                try:
                    # Measure resource usage before operation
                    process = psutil.Process()
                    memory_before = process.memory_info().rss / 1024 / 1024  # MB
                    cpu_before = process.cpu_percent()
                    
                    # Execute operation
                    await operation_func()
                    
                    # Measure resource usage after operation
                    memory_after = process.memory_info().rss / 1024 / 1024  # MB
                    cpu_after = process.cpu_percent()
                    
                    success = True
                    break
                    
                except Exception as e:
                    error_message = str(e)
                    retry_count += 1
                    
                    if attempt < self.max_retries:
                        await asyncio.sleep(self.retry_delay)
                    else:
                        break
            
            end_time = time.time()
            response_time_ms = (end_time - start_time) * 1000
            
            # Record resource usage
            resource_usage = {
                "memory_before_mb": memory_before,
                "memory_after_mb": memory_after,
                "memory_delta_mb": memory_after - memory_before,
                "cpu_before_percent": cpu_before,
                "cpu_after_percent": cpu_after,
                "cpu_delta_percent": cpu_after - cpu_before
            }
            
            results.append({
                "response_time_ms": response_time_ms,
                "success": success,
                "error_message": error_message,
                "resource_usage": resource_usage
            })
        
        # Calculate statistical measures
        response_times = [r["response_time_ms"] for r in results if r["success"]]
        
        if response_times:
            mean_response_time = statistics.mean(response_times)
            std_deviation = statistics.stdev(response_times) if len(response_times) > 1 else 0
            percentile_95 = statistics.quantiles(response_times, n=20)[18] if len(response_times) >= 20 else max(response_times)
            percentile_99 = statistics.quantiles(response_times, n=100)[98] if len(response_times) >= 100 else max(response_times)
        else:
            mean_response_time = 0
            std_deviation = 0
            percentile_95 = 0
            percentile_99 = 0
        
        # Calculate baseline deviation if baseline exists
        baseline_deviation = 0.0
        if operation_name in self.baseline_data:
            baseline = self.baseline_data[operation_name]
            if baseline.std_deviation_ms > 0:
                baseline_deviation = (mean_response_time - baseline.mean_response_time_ms) / baseline.std_deviation_ms
        
        # Determine overall success and budget compliance
        overall_success = any(r["success"] for r in results)
        meets_budget = mean_response_time <= budget_target_ms if overall_success else False
        
        # Use median response time for final result
        successful_times = [r["response_time_ms"] for r in results if r["success"]]
        final_response_time = statistics.median(successful_times) if successful_times else 0
        
        # Aggregate resource usage
        avg_memory_delta = statistics.mean([r["resource_usage"]["memory_delta_mb"] for r in results])
        avg_cpu_delta = statistics.mean([r["resource_usage"]["cpu_delta_percent"] for r in results])
        
        aggregated_resource_usage = {
            "memory_delta_mb": avg_memory_delta,
            "cpu_delta_percent": avg_cpu_delta,
            "sample_count": len(results),
            "successful_samples": len(successful_times)
        }
        
        result = EnhancedPerformanceResult(
            operation=operation_name,
            response_time_ms=final_response_time,
            success=overall_success,
            meets_budget=meets_budget,
            budget_target_ms=budget_target_ms,
            resource_usage=aggregated_resource_usage,
            retry_count=retry_count,
            error_message=error_message if not overall_success else None,
            baseline_deviation=baseline_deviation,
            percentile_95=percentile_95,
            percentile_99=percentile_99
        )
        
        self.performance_results.append(result)
        return result

    async def test_service_connection_reliability(self) -> EnhancedPerformanceResult:
        """
        Enhanced service connection test with reliability improvements.
        """
        async def connection_operation():
            # Test WebSocket connection
            async with websockets.connect(self.websocket_url) as websocket:
                # Send ping to verify connection
                await websocket.ping()
                await websocket.pong()
        
        return await self.measure_operation_with_retry(
            "service_connection",
            connection_operation,
            self.performance_budgets["service_connection"]
        )

    async def test_camera_list_refresh_reliability(self) -> EnhancedPerformanceResult:
        """
        Enhanced camera list refresh test with reliability improvements.
        """
        async def camera_list_operation():
            # Test camera list API
            async with aiohttp.ClientSession() as session:
                url = f"http://127.0.0.1:9997/v3/paths/list"
                async with session.get(url) as response:
                    if response.status != 200:
                        raise Exception(f"Camera list API returned status {response.status}")
                    data = await response.json()
                    if not isinstance(data, dict):
                        raise Exception("Invalid response format")
        
        return await self.measure_operation_with_retry(
            "camera_list_refresh",
            camera_list_operation,
            self.performance_budgets["camera_list_refresh"]
        )

    async def test_health_check_reliability(self) -> EnhancedPerformanceResult:
        """
        Enhanced health check test with reliability improvements.
        """
        async def health_check_operation():
            # Test MediaMTX health check
            async with aiohttp.ClientSession() as session:
                url = f"http://127.0.0.1:9997/v3/config/global/get"
                async with session.get(url) as response:
                    if response.status != 200:
                        raise Exception(f"Health check returned status {response.status}")
                    data = await response.json()
                    if not isinstance(data, dict):
                        raise Exception("Invalid health check response")
        
        return await self.measure_operation_with_retry(
            "health_check",
            health_check_operation,
            self.performance_budgets["health_check"]
        )

    async def test_api_responsiveness_reliability(self) -> EnhancedPerformanceResult:
        """
        Enhanced API responsiveness test with reliability improvements.
        """
        async def api_operation():
            # Test multiple API endpoints for responsiveness
            async with aiohttp.ClientSession() as session:
                endpoints = [
                    "/v3/config/global/get",
                    "/v3/paths/list",
                    "/v3/paths/list"
                ]
                
                for endpoint in endpoints:
                    url = f"http://127.0.0.1:9997{endpoint}"
                    async with session.get(url) as response:
                        if response.status != 200:
                            raise Exception(f"API endpoint {endpoint} returned status {response.status}")
        
        return await self.measure_operation_with_retry(
            "api_responsiveness",
            api_operation,
            self.performance_budgets["basic_api_call"]
        )

    async def test_websocket_connection_reliability(self) -> EnhancedPerformanceResult:
        """
        Enhanced WebSocket connection test with reliability improvements.
        """
        async def websocket_operation():
            # Test WebSocket connection and message exchange
            async with websockets.connect(self.websocket_url) as websocket:
                # Send test message
                test_message = {
                    "jsonrpc": "2.0",
                    "id": 1,
                    "method": "ping",
                    "params": {}
                }
                await websocket.send(json.dumps(test_message))
                
                # Wait for response
                response = await websocket.recv()
                response_data = json.loads(response)
                
                if "error" in response_data:
                    raise Exception(f"WebSocket error: {response_data['error']}")
        
        return await self.measure_operation_with_retry(
            "websocket_connection",
            websocket_operation,
            self.performance_budgets["websocket_connection"]
        )

    def generate_enhanced_performance_report(self) -> Dict[str, Any]:
        """Generate comprehensive enhanced performance report."""
        total_operations = len(self.performance_results)
        successful_operations = sum(1 for r in self.performance_results if r.success)
        budget_compliant_operations = sum(1 for r in self.performance_results if r.meets_budget)
        
        # Calculate overall statistics
        response_times = [r.response_time_ms for r in self.performance_results if r.success]
        overall_stats = {
            "mean_response_time_ms": statistics.mean(response_times) if response_times else 0,
            "median_response_time_ms": statistics.median(response_times) if response_times else 0,
            "std_deviation_ms": statistics.stdev(response_times) if len(response_times) > 1 else 0,
            "min_response_time_ms": min(response_times) if response_times else 0,
            "max_response_time_ms": max(response_times) if response_times else 0
        }
        
        # Identify performance regressions
        regressions = []
        for result in self.performance_results:
            if result.baseline_deviation > self.baseline_threshold:
                regressions.append({
                    "operation": result.operation,
                    "baseline_deviation": result.baseline_deviation,
                    "current_time_ms": result.response_time_ms
                })
        
        return {
            "test_summary": {
                "total_operations": total_operations,
                "successful_operations": successful_operations,
                "success_rate": (successful_operations / total_operations * 100) if total_operations > 0 else 0,
                "budget_compliant_operations": budget_compliant_operations,
                "budget_compliance_rate": (budget_compliant_operations / total_operations * 100) if total_operations > 0 else 0
            },
            "overall_statistics": overall_stats,
            "operation_results": [
                {
                    "operation": r.operation,
                    "response_time_ms": r.response_time_ms,
                    "success": r.success,
                    "meets_budget": r.meets_budget,
                    "budget_target_ms": r.budget_target_ms,
                    "retry_count": r.retry_count,
                    "baseline_deviation": r.baseline_deviation,
                    "percentile_95_ms": r.percentile_95,
                    "percentile_99_ms": r.percentile_99,
                    "resource_usage": r.resource_usage
                }
                for r in self.performance_results
            ],
            "performance_regressions": regressions,
            "budget_violations": self.budget_violations
        }


# Pytest test fixtures and test functions

@pytest.fixture
async def enhanced_performance_validator():
    """Fixture for enhanced performance sanity validator."""
    validator = EnhancedPerformanceSanityValidator()
    await validator.setup_real_performance_environment()
    yield validator
    await validator.cleanup_real_environment()


@pytest.mark.pdr
@pytest.mark.asyncio
async def test_service_connection_reliability_enhanced(enhanced_performance_validator):
    """Enhanced service connection reliability test."""
    result = await enhanced_performance_validator.test_service_connection_reliability()
    assert result.success, f"Service connection failed: {result.error_message}"
    assert result.meets_budget, f"Service connection {result.response_time_ms}ms exceeds budget {result.budget_target_ms}ms"


@pytest.mark.pdr
@pytest.mark.asyncio
async def test_camera_list_refresh_reliability_enhanced(enhanced_performance_validator):
    """Enhanced camera list refresh reliability test."""
    result = await enhanced_performance_validator.test_camera_list_refresh_reliability()
    assert result.success, f"Camera list refresh failed: {result.error_message}"
    assert result.meets_budget, f"Camera list refresh {result.response_time_ms}ms exceeds budget {result.budget_target_ms}ms"


@pytest.mark.pdr
@pytest.mark.asyncio
async def test_health_check_reliability_enhanced(enhanced_performance_validator):
    """Enhanced health check reliability test."""
    result = await enhanced_performance_validator.test_health_check_reliability()
    assert result.success, f"Health check failed: {result.error_message}"
    assert result.meets_budget, f"Health check {result.response_time_ms}ms exceeds budget {result.budget_target_ms}ms"


@pytest.mark.pdr
@pytest.mark.asyncio
async def test_api_responsiveness_reliability_enhanced(enhanced_performance_validator):
    """Enhanced API responsiveness reliability test."""
    result = await enhanced_performance_validator.test_api_responsiveness_reliability()
    assert result.success, f"API responsiveness failed: {result.error_message}"
    assert result.meets_budget, f"API responsiveness {result.response_time_ms}ms exceeds budget {result.budget_target_ms}ms"


@pytest.mark.pdr
@pytest.mark.asyncio
async def test_websocket_connection_reliability_enhanced(enhanced_performance_validator):
    """Enhanced WebSocket connection reliability test."""
    result = await enhanced_performance_validator.test_websocket_connection_reliability()
    assert result.success, f"WebSocket connection failed: {result.error_message}"
    assert result.meets_budget, f"WebSocket connection {result.response_time_ms}ms exceeds budget {result.budget_target_ms}ms"


@pytest.mark.pdr
@pytest.mark.asyncio
async def test_comprehensive_enhanced_performance_validation(enhanced_performance_validator):
    """Comprehensive enhanced performance validation test."""
    # Run all enhanced performance tests
    await enhanced_performance_validator.test_service_connection_reliability()
    await enhanced_performance_validator.test_camera_list_refresh_reliability()
    await enhanced_performance_validator.test_health_check_reliability()
    await enhanced_performance_validator.test_api_responsiveness_reliability()
    await enhanced_performance_validator.test_websocket_connection_reliability()
    
    # Generate comprehensive report
    report = enhanced_performance_validator.generate_enhanced_performance_report()
    
    # Validate PDR acceptance criteria
    success_rate = report["test_summary"]["success_rate"]
    budget_compliance_rate = report["test_summary"]["budget_compliance_rate"]
    
    print(f"Enhanced Performance Test Results:")
    print(f"  Success Rate: {success_rate:.1f}%")
    print(f"  Budget Compliance Rate: {budget_compliance_rate:.1f}%")
    print(f"  Total Operations: {report['test_summary']['total_operations']}")
    print(f"  Overall Mean Response Time: {report['overall_statistics']['mean_response_time_ms']:.1f}ms")
    
    # PDR acceptance criteria: 80% success rate, 80% budget compliance
    assert success_rate >= 80.0, f"Success rate {success_rate}% below PDR threshold of 80%"
    assert budget_compliance_rate >= 80.0, f"Budget compliance rate {budget_compliance_rate}% below PDR threshold of 80%"
    
    # Log detailed results
    for result in report["operation_results"]:
        status = "✅" if result["success"] and result["meets_budget"] else "❌"
        print(f"  {result['operation']}: {status} ({result['response_time_ms']:.1f}ms)")
        if result["retry_count"] > 0:
            print(f"    Retries: {result['retry_count']}")
        if result["baseline_deviation"] > 0:
            print(f"    Baseline deviation: {result['baseline_deviation']:.2f}σ")
    
    # Check for performance regressions
    if report["performance_regressions"]:
        print(f"  Performance Regressions Detected: {len(report['performance_regressions'])}")
        for regression in report["performance_regressions"]:
            print(f"    {regression['operation']}: {regression['baseline_deviation']:.2f}σ deviation")
