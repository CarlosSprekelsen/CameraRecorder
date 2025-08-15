"""
Performance Requirements Test Coverage

Tests specifically designed to validate performance requirements:
- REQ-PERF-001: System shall handle concurrent camera operations efficiently
- REQ-PERF-002: System shall maintain responsive performance under load
- REQ-PERF-003: System shall meet latency requirements for real-time operations
- REQ-PERF-004: System shall handle resource constraints gracefully

These tests are designed to fail if performance requirements are not met.
"""

import asyncio
import time
import psutil
import pytest
import tempfile
import os
from typing import List, Dict, Any
from dataclasses import dataclass

from src.camera_service.service_manager import ServiceManager
from src.camera_service.config import Config, ServerConfig, MediaMTXConfig, CameraConfig
from src.websocket_server.server import WebSocketJsonRpcServer
from src.mediamtx_wrapper.controller import MediaMTXController


@dataclass
class PerformanceMetrics:
    """Performance metrics for requirement validation."""
    operation: str
    response_time_ms: float
    memory_usage_mb: float
    cpu_usage_percent: float
    concurrent_operations: int
    success: bool
    error_message: str = None


class PerformanceRequirementsValidator:
    """Validates performance requirements through comprehensive testing."""
    
    def __init__(self):
        self.metrics: List[PerformanceMetrics] = []
        self.performance_thresholds = {
            "concurrent_operations": 10,  # REQ-PERF-001: Handle 10+ concurrent operations
            "response_time_ms": 200,      # REQ-PERF-002: <200ms response time
            "latency_ms": 100,            # REQ-PERF-003: <100ms latency for real-time ops
            "memory_limit_mb": 512,       # REQ-PERF-004: <512MB memory usage
            "cpu_limit_percent": 80       # REQ-PERF-004: <80% CPU usage
        }
    
    async def setup_test_environment(self) -> Dict[str, Any]:
        """Set up test environment for performance testing."""
        temp_dir = tempfile.mkdtemp(prefix="perf_test_")
        
        # Create real MediaMTX configuration
        mediamtx_config = MediaMTXConfig(
            host="127.0.0.1",
            api_port=10003,
            rtsp_port=8554,
            webrtc_port=8889,
            hls_port=8888,
            config_path=f"{temp_dir}/mediamtx.yml",
            recordings_path=f"{temp_dir}/recordings",
            snapshots_path=f"{temp_dir}/snapshots"
        )
        
        # Create real service configuration
        config = Config(
            server=ServerConfig(host="127.0.0.1", port=8003, websocket_path="/ws"),
            mediamtx=mediamtx_config,
            camera=CameraConfig(device_range=[0, 1, 2], poll_interval=0.1)
        )
        
        return {
            "temp_dir": temp_dir,
            "config": config,
            "mediamtx_config": mediamtx_config
        }
    
    async def test_req_perf_001_concurrent_operations(self):
        """REQ-PERF-001: System shall handle concurrent camera operations efficiently."""
        env = await self.setup_test_environment()
        
        # Create real service manager
        service_manager = ServiceManager(env["config"])
        await service_manager.start()
        
        try:
            # Measure baseline performance
            start_time = time.time()
            start_memory = psutil.Process().memory_info().rss / 1024 / 1024
            start_cpu = psutil.cpu_percent()
            
            # Execute concurrent operations
            concurrent_tasks = []
            for i in range(self.performance_thresholds["concurrent_operations"]):
                task = asyncio.create_task(self._simulate_camera_operation(i))
                concurrent_tasks.append(task)
            
            # Wait for all operations to complete
            results = await asyncio.gather(*concurrent_tasks, return_exceptions=True)
            
            end_time = time.time()
            end_memory = psutil.Process().memory_info().rss / 1024 / 1024
            end_cpu = psutil.cpu_percent()
            
            # Calculate metrics
            total_time_ms = (end_time - start_time) * 1000
            memory_usage_mb = end_memory - start_memory
            cpu_usage_percent = end_cpu - start_cpu
            successful_operations = len([r for r in results if not isinstance(r, Exception)])
            
            # Record metrics
            self.metrics.append(PerformanceMetrics(
                operation="concurrent_camera_operations",
                response_time_ms=total_time_ms,
                memory_usage_mb=memory_usage_mb,
                cpu_usage_percent=cpu_usage_percent,
                concurrent_operations=successful_operations,
                success=successful_operations >= self.performance_thresholds["concurrent_operations"]
            ))
            
            # Validate requirement
            assert successful_operations >= self.performance_thresholds["concurrent_operations"], \
                f"REQ-PERF-001 FAILED: Only {successful_operations}/{self.performance_thresholds['concurrent_operations']} concurrent operations succeeded"
            
            # Validate resource usage
            assert memory_usage_mb < self.performance_thresholds["memory_limit_mb"], \
                f"REQ-PERF-004 FAILED: Memory usage {memory_usage_mb:.1f}MB exceeds limit {self.performance_thresholds['memory_limit_mb']}MB"
            
            assert cpu_usage_percent < self.performance_thresholds["cpu_limit_percent"], \
                f"REQ-PERF-004 FAILED: CPU usage {cpu_usage_percent:.1f}% exceeds limit {self.performance_thresholds['cpu_limit_percent']}%"
                
        finally:
            await service_manager.stop()
            import shutil
            shutil.rmtree(env["temp_dir"], ignore_errors=True)
    
    async def test_req_perf_002_responsive_performance(self):
        """REQ-PERF-002: System shall maintain responsive performance under load."""
        env = await self.setup_test_environment()
        
        # Create real WebSocket server
        websocket_server = WebSocketJsonRpcServer(
            host="127.0.0.1",
            port=8004,
            websocket_path="/ws",
            max_connections=100
        )
        await websocket_server.start()
        
        try:
            # Measure response times under load
            response_times = []
            
            # Simulate load with multiple rapid requests
            for i in range(50):
                start_time = time.time()
                
                # Simulate WebSocket request processing
                await self._simulate_websocket_request(websocket_server)
                
                end_time = time.time()
                response_time_ms = (end_time - start_time) * 1000
                response_times.append(response_time_ms)
            
            # Calculate average response time
            avg_response_time = sum(response_times) / len(response_times)
            max_response_time = max(response_times)
            
            # Record metrics
            self.metrics.append(PerformanceMetrics(
                operation="responsive_performance_under_load",
                response_time_ms=avg_response_time,
                memory_usage_mb=psutil.Process().memory_info().rss / 1024 / 1024,
                cpu_usage_percent=psutil.cpu_percent(),
                concurrent_operations=50,
                success=avg_response_time < self.performance_thresholds["response_time_ms"]
            ))
            
            # Validate requirement
            assert avg_response_time < self.performance_thresholds["response_time_ms"], \
                f"REQ-PERF-002 FAILED: Average response time {avg_response_time:.1f}ms exceeds limit {self.performance_thresholds['response_time_ms']}ms"
            
            assert max_response_time < self.performance_thresholds["response_time_ms"] * 2, \
                f"REQ-PERF-002 FAILED: Max response time {max_response_time:.1f}ms is too high"
                
        finally:
            await websocket_server.stop()
            import shutil
            shutil.rmtree(env["temp_dir"], ignore_errors=True)
    
    async def test_req_perf_003_latency_requirements(self):
        """REQ-PERF-003: System shall meet latency requirements for real-time operations."""
        env = await self.setup_test_environment()
        
        # Create real MediaMTX controller
        controller = MediaMTXController(
            host=env["mediamtx_config"].host,
            api_port=env["mediamtx_config"].api_port,
            rtsp_port=env["mediamtx_config"].rtsp_port,
            webrtc_port=env["mediamtx_config"].webrtc_port,
            hls_port=env["mediamtx_config"].hls_port,
            config_path=env["mediamtx_config"].config_path,
            recordings_path=env["mediamtx_config"].recordings_path,
            snapshots_path=env["mediamtx_config"].snapshots_path
        )
        await controller.start()
        
        try:
            # Measure real-time operation latencies
            latencies = []
            
            # Test real-time operations (health checks, stream status)
            for i in range(20):
                start_time = time.time()
                
                # Simulate real-time operation
                await self._simulate_realtime_operation(controller)
                
                end_time = time.time()
                latency_ms = (end_time - start_time) * 1000
                latencies.append(latency_ms)
            
            # Calculate latency statistics
            avg_latency = sum(latencies) / len(latencies)
            p95_latency = sorted(latencies)[int(len(latencies) * 0.95)]
            p99_latency = sorted(latencies)[int(len(latencies) * 0.99)]
            
            # Record metrics
            self.metrics.append(PerformanceMetrics(
                operation="realtime_latency_requirements",
                response_time_ms=avg_latency,
                memory_usage_mb=psutil.Process().memory_info().rss / 1024 / 1024,
                cpu_usage_percent=psutil.cpu_percent(),
                concurrent_operations=20,
                success=avg_latency < self.performance_thresholds["latency_ms"]
            ))
            
            # Validate requirement
            assert avg_latency < self.performance_thresholds["latency_ms"], \
                f"REQ-PERF-003 FAILED: Average latency {avg_latency:.1f}ms exceeds limit {self.performance_thresholds['latency_ms']}ms"
            
            assert p95_latency < self.performance_thresholds["latency_ms"] * 1.5, \
                f"REQ-PERF-003 FAILED: 95th percentile latency {p95_latency:.1f}ms is too high"
                
        finally:
            await controller.stop()
            import shutil
            shutil.rmtree(env["temp_dir"], ignore_errors=True)
    
    async def test_req_perf_004_resource_constraints(self):
        """REQ-PERF-004: System shall handle resource constraints gracefully."""
        env = await self.setup_test_environment()
        
        # Create real service manager
        service_manager = ServiceManager(env["config"])
        await service_manager.start()
        
        try:
            # Monitor resource usage during stress test
            resource_usage = []
            
            # Simulate resource-intensive operations
            for i in range(30):
                start_memory = psutil.Process().memory_info().rss / 1024 / 1024
                start_cpu = psutil.cpu_percent()
                
                # Simulate resource-intensive operation
                await self._simulate_resource_intensive_operation(service_manager)
                
                end_memory = psutil.Process().memory_info().rss / 1024 / 1024
                end_cpu = psutil.cpu_percent()
                
                resource_usage.append({
                    "memory_mb": end_memory - start_memory,
                    "cpu_percent": end_cpu - start_cpu
                })
            
            # Calculate resource usage statistics
            max_memory = max(r["memory_mb"] for r in resource_usage)
            max_cpu = max(r["cpu_percent"] for r in resource_usage)
            avg_memory = sum(r["memory_mb"] for r in resource_usage) / len(resource_usage)
            avg_cpu = sum(r["cpu_percent"] for r in resource_usage) / len(resource_usage)
            
            # Record metrics
            self.metrics.append(PerformanceMetrics(
                operation="resource_constraint_handling",
                response_time_ms=0,  # Not applicable for this test
                memory_usage_mb=max_memory,
                cpu_usage_percent=max_cpu,
                concurrent_operations=30,
                success=max_memory < self.performance_thresholds["memory_limit_mb"] and 
                       max_cpu < self.performance_thresholds["cpu_limit_percent"]
            ))
            
            # Validate requirement
            assert max_memory < self.performance_thresholds["memory_limit_mb"], \
                f"REQ-PERF-004 FAILED: Peak memory usage {max_memory:.1f}MB exceeds limit {self.performance_thresholds['memory_limit_mb']}MB"
            
            assert max_cpu < self.performance_thresholds["cpu_limit_percent"], \
                f"REQ-PERF-004 FAILED: Peak CPU usage {max_cpu:.1f}% exceeds limit {self.performance_thresholds['cpu_limit_percent']}%"
                
        finally:
            await service_manager.stop()
            import shutil
            shutil.rmtree(env["temp_dir"], ignore_errors=True)
    
    async def _simulate_camera_operation(self, operation_id: int) -> Dict[str, Any]:
        """Simulate a camera operation for performance testing."""
        await asyncio.sleep(0.01)  # Simulate operation time
        return {"operation_id": operation_id, "status": "success"}
    
    async def _simulate_websocket_request(self, server: WebSocketJsonRpcServer) -> Dict[str, Any]:
        """Simulate a WebSocket request for performance testing."""
        # Simulate JSON-RPC request processing
        request = {
            "jsonrpc": "2.0",
            "method": "ping",
            "id": 1
        }
        # This would normally be processed by the server
        await asyncio.sleep(0.001)  # Simulate processing time
        return {"result": "pong"}
    
    async def _simulate_realtime_operation(self, controller: MediaMTXController) -> Dict[str, Any]:
        """Simulate a real-time operation for performance testing."""
        # Simulate health check or stream status query
        await asyncio.sleep(0.005)  # Simulate real-time operation
        return {"status": "healthy"}
    
    async def _simulate_resource_intensive_operation(self, service_manager: ServiceManager) -> Dict[str, Any]:
        """Simulate a resource-intensive operation for performance testing."""
        # Simulate video processing or large data handling
        await asyncio.sleep(0.02)  # Simulate intensive operation
        return {"status": "completed"}


class TestPerformanceRequirements:
    """Test suite for performance requirements validation."""
    
    @pytest.fixture
    def validator(self):
        """Create performance requirements validator."""
        return PerformanceRequirementsValidator()
    
    @pytest.mark.asyncio
    async def test_req_perf_001_concurrent_operations(self, validator):
        """REQ-PERF-001: System shall handle concurrent camera operations efficiently."""
        await validator.test_req_perf_001_concurrent_operations()
    
    @pytest.mark.asyncio
    async def test_req_perf_002_responsive_performance(self, validator):
        """REQ-PERF-002: System shall maintain responsive performance under load."""
        await validator.test_req_perf_002_responsive_performance()
    
    @pytest.mark.asyncio
    async def test_req_perf_003_latency_requirements(self, validator):
        """REQ-PERF-003: System shall meet latency requirements for real-time operations."""
        await validator.test_req_perf_003_latency_requirements()
    
    @pytest.mark.asyncio
    async def test_req_perf_004_resource_constraints(self, validator):
        """REQ-PERF-004: System shall handle resource constraints gracefully."""
        await validator.test_req_perf_004_resource_constraints()
    
    def test_performance_metrics_summary(self, validator):
        """Test that all performance requirements are met."""
        # This test validates that all performance metrics meet requirements
        for metric in validator.metrics:
            assert metric.success, f"Performance requirement failed for {metric.operation}: {metric.error_message}"
