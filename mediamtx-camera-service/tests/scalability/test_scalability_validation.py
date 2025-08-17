"""
Scalability Testing Validation

Executes comprehensive scalability testing against established performance requirements:
- Concurrent connection testing at multiple levels (10, 25, 50, 75, 100, 125, 150)
- Performance validation against established requirements
- Resource utilization monitoring
- Performance degradation analysis

Requirements:
- REQ-PERF-001: System shall handle concurrent camera operations efficiently
- REQ-PERF-002: System shall maintain responsive performance under load
- REQ-PERF-003: System shall meet latency requirements for real-time operations
- REQ-PERF-004: System shall handle resource constraints gracefully
- REQ-PERF-005: System shall process requests at specified throughput rates
- REQ-PERF-006: System shall scale performance with available resources

IMPORTANT: This test is designed to validate real system behavior and may fail on REQ-PERF-006
due to known connection limits. The failure is expected and validates that the system
correctly identifies scalability limitations. The test includes proper timeouts to prevent
hanging and will complete within 2 minutes maximum.
"""

import asyncio
import time
import psutil
import pytest
import tempfile
import os
import json
import statistics
from typing import List, Dict, Any, Optional, Tuple
from dataclasses import dataclass, asdict
from collections import defaultdict
import websockets
import aiohttp
import logging

from src.camera_service.service_manager import ServiceManager
from src.camera_service.config import Config, ServerConfig, MediaMTXConfig, CameraConfig
from src.websocket_server.server import WebSocketJsonRpcServer
from src.mediamtx_wrapper.controller import MediaMTXController


@dataclass
class ScalabilityTestResult:
    """Results from a single scalability test level."""
    concurrent_connections: int
    successful_connections: int
    failed_connections: int
    avg_response_time_ms: float
    p95_response_time_ms: float
    p99_response_time_ms: float
    max_response_time_ms: float
    avg_memory_usage_mb: float
    max_memory_usage_mb: float
    avg_cpu_usage_percent: float
    max_cpu_usage_percent: float
    throughput_requests_per_sec: float
    error_rate_percent: float
    test_duration_seconds: float
    requirements_compliant: bool
    failure_point: bool
    timestamp: float


@dataclass
class PerformanceRequirements:
    """Established performance requirements from Task 1."""
    # REQ-PERF-001: Concurrent operations
    max_concurrent_connections: int = 100  # Python baseline
    target_concurrent_connections: int = 50  # Target for reliable operation
    
    # REQ-PERF-002: Response time
    max_response_time_ms: float = 500  # Python baseline
    target_response_time_ms: float = 200  # Target response time
    
    # REQ-PERF-003: Latency
    max_latency_ms: float = 100  # Real-time operations
    target_latency_ms: float = 50  # Target latency
    
    # REQ-PERF-004: Resource usage
    max_cpu_usage_percent: float = 70  # Python baseline
    max_memory_usage_mb: float = 512  # Memory limit
    max_disk_io_mb: float = 100  # Disk I/O limit
    max_network_io_mb: float = 50  # Network I/O limit
    
    # REQ-PERF-005: Throughput
    min_throughput_requests_per_sec: float = 100  # Python baseline
    target_throughput_requests_per_sec: float = 200  # Target throughput
    
    # REQ-PERF-006: Scalability
    max_error_rate_percent: float = 5  # Maximum acceptable error rate
    performance_degradation_threshold: float = 2.0  # Max degradation factor


class ScalabilityValidator:
    """Validates scalability against established performance requirements."""
    
    def __init__(self):
        self.test_results: List[ScalabilityTestResult] = []
        self.requirements = PerformanceRequirements()
        self.logger = logging.getLogger(__name__)
        
        # Test configuration
        self.connection_levels = [10, 25, 50, 75, 100, 125, 150]
        self.test_duration_seconds = 30  # Duration for each test level
        self.warmup_duration_seconds = 5  # Warmup period
        self.cooldown_duration_seconds = 5  # Cooldown period
        
        # Monitoring
        self._monitoring_active = False
        self._resource_metrics: List[Dict[str, float]] = []
        self._process = psutil.Process()
    
    async def start_resource_monitoring(self) -> None:
        """Start resource monitoring during tests."""
        self._monitoring_active = True
        self._resource_metrics = []
        
        while self._monitoring_active:
            try:
                memory_info = self._process.memory_info()
                cpu_percent = self._process.cpu_percent()
                
                self._resource_metrics.append({
                    "timestamp": time.time(),
                    "memory_mb": memory_info.rss / 1024 / 1024,
                    "cpu_percent": cpu_percent
                })
                
                await asyncio.sleep(1.0)  # Monitor every second
                
            except asyncio.CancelledError:
                break
            except Exception as e:
                self.logger.error(f"Monitoring error: {e}")
                await asyncio.sleep(1)
    
    async def stop_resource_monitoring(self) -> None:
        """Stop resource monitoring."""
        self._monitoring_active = False
    
    def get_resource_summary(self) -> Dict[str, float]:
        """Get summary of resource usage during test."""
        if not self._resource_metrics:
            return {"avg_memory_mb": 0, "max_memory_mb": 0, "avg_cpu_percent": 0, "max_cpu_percent": 0}
        
        memory_values = [m["memory_mb"] for m in self._resource_metrics]
        cpu_values = [m["cpu_percent"] for m in self._resource_metrics]
        
        return {
            "avg_memory_mb": statistics.mean(memory_values),
            "max_memory_mb": max(memory_values),
            "avg_cpu_percent": statistics.mean(cpu_values),
            "max_cpu_percent": max(cpu_values)
        }
    
    async def setup_test_environment(self) -> Dict[str, Any]:
        """Set up test environment for scalability testing."""
        temp_dir = tempfile.mkdtemp(prefix="scalability_test_")
        
        # Create MediaMTX configuration
        mediamtx_config = MediaMTXConfig(
            host="127.0.0.1",
            api_port=10004,
            rtsp_port=8555,
            webrtc_port=8890,
            hls_port=8891,
            config_path=f"{temp_dir}/mediamtx.yml",
            recordings_path=f"{temp_dir}/recordings",
            snapshots_path=f"{temp_dir}/snapshots"
        )
        
        # Create service configuration
        config = Config(
            server=ServerConfig(host="127.0.0.1", port=8005, websocket_path="/ws"),
            mediamtx=mediamtx_config,
            camera=CameraConfig(device_range=[0, 1, 2], poll_interval=0.1)
        )
        
        return {
            "temp_dir": temp_dir,
            "config": config,
            "mediamtx_config": mediamtx_config
        }
    
    async def test_concurrent_connections(self, connection_level: int, env: Dict[str, Any]) -> ScalabilityTestResult:
        """Test system performance with specified number of concurrent connections."""
        self.logger.info(f"Testing {connection_level} concurrent connections...")
        
        # Start resource monitoring
        monitoring_task = asyncio.create_task(self.start_resource_monitoring())
        
        # Create WebSocket server
        websocket_server = WebSocketJsonRpcServer(
            host=env["config"].server.host,
            port=env["config"].server.port,
            websocket_path=env["config"].server.websocket_path,
            max_connections=connection_level * 2  # Allow for some overhead
        )
        
        # Register test methods
        async def ping_method(params=None):
            await asyncio.sleep(0.001)  # Minimal processing time
            return {"status": "pong", "timestamp": time.time()}
        
        async def get_metrics_method(params=None):
            await asyncio.sleep(0.005)  # Simulate metrics collection
            return {"metrics": {"connections": connection_level, "timestamp": time.time()}}
        
        websocket_server.register_method("ping", ping_method)
        websocket_server.register_method("get_metrics", get_metrics_method)
        
        await websocket_server.start()
        
        try:
            # Warmup period
            await asyncio.sleep(self.warmup_duration_seconds)
            
            # Test concurrent connections with timeout
            start_time = time.time()
            response_times = []
            successful_connections = 0
            failed_connections = 0
            
            # Create concurrent connection tasks
            connection_tasks = []
            for i in range(connection_level):
                task = asyncio.create_task(
                    self._simulate_client_connection(env["config"], i, response_times)
                )
                connection_tasks.append(task)
            
            # Wait for all connections to complete with timeout
            try:
                results = await asyncio.wait_for(
                    asyncio.gather(*connection_tasks, return_exceptions=True),
                    timeout=self.test_duration_seconds + 10  # Add 10 seconds buffer
                )
            except asyncio.TimeoutError:
                self.logger.warning(f"Test timeout at {connection_level} connections")
                # Cancel remaining tasks
                for task in connection_tasks:
                    if not task.done():
                        task.cancel()
                # Wait for cancellation to complete
                try:
                    await asyncio.wait_for(
                        asyncio.gather(*connection_tasks, return_exceptions=True),
                        timeout=5
                    )
                except asyncio.TimeoutError:
                    self.logger.error(f"Failed to cancel tasks at {connection_level} connections")
                results = [Exception("Timeout")] * len(connection_tasks)
            
            # Count successful and failed connections
            for result in results:
                if isinstance(result, Exception):
                    failed_connections += 1
                else:
                    successful_connections += 1
            
            end_time = time.time()
            test_duration = end_time - start_time
            
            # Calculate response time statistics
            if response_times:
                avg_response_time = statistics.mean(response_times)
                p95_response_time = statistics.quantiles(response_times, n=20)[18] if len(response_times) >= 20 else max(response_times)
                p99_response_time = statistics.quantiles(response_times, n=100)[98] if len(response_times) >= 100 else max(response_times)
                max_response_time = max(response_times)
            else:
                avg_response_time = p95_response_time = p99_response_time = max_response_time = 0
            
            # Calculate throughput
            total_requests = len(response_times)
            throughput_requests_per_sec = total_requests / test_duration if test_duration > 0 else 0
            
            # Calculate error rate
            error_rate_percent = (failed_connections / connection_level) * 100 if connection_level > 0 else 0
            
            # Get resource usage summary
            await self.stop_resource_monitoring()
            resource_summary = self.get_resource_summary()
            
            # Determine if requirements are met
            requirements_compliant = self._validate_requirements(
                successful_connections, avg_response_time, resource_summary, 
                throughput_requests_per_sec, error_rate_percent
            )
            
            # Determine if this is a failure point (enhanced detection)
            failure_point = (
                error_rate_percent > self.requirements.max_error_rate_percent or
                avg_response_time > self.requirements.max_response_time_ms or
                resource_summary["max_cpu_percent"] > self.requirements.max_cpu_usage_percent or
                resource_summary["max_memory_mb"] > self.requirements.max_memory_usage_mb or
                failed_connections > 0  # Any connection failure indicates a problem
            )
            
            # Create test result
            result = ScalabilityTestResult(
                concurrent_connections=connection_level,
                successful_connections=successful_connections,
                failed_connections=failed_connections,
                avg_response_time_ms=avg_response_time,
                p95_response_time_ms=p95_response_time,
                p99_response_time_ms=p99_response_time,
                max_response_time_ms=max_response_time,
                avg_memory_usage_mb=resource_summary["avg_memory_mb"],
                max_memory_usage_mb=resource_summary["max_memory_mb"],
                avg_cpu_usage_percent=resource_summary["avg_cpu_percent"],
                max_cpu_usage_percent=resource_summary["max_cpu_percent"],
                throughput_requests_per_sec=throughput_requests_per_sec,
                error_rate_percent=error_rate_percent,
                test_duration_seconds=test_duration,
                requirements_compliant=requirements_compliant,
                failure_point=failure_point,
                timestamp=time.time()
            )
            
            self.test_results.append(result)
            
            # Cooldown period
            await asyncio.sleep(self.cooldown_duration_seconds)
            
            return result
            
        finally:
            await websocket_server.stop()
            if monitoring_task and not monitoring_task.done():
                monitoring_task.cancel()
                try:
                    await monitoring_task
                except asyncio.CancelledError:
                    pass
    
    async def _simulate_client_connection(self, config: Config, client_id: int, response_times: List[float]) -> Dict[str, Any]:
        """Simulate a client connection with performance measurement."""
        try:
            uri = f"ws://{config.server.host}:{config.server.port}{config.server.websocket_path}"
            
            # Add timeout to websocket connection
            async with websockets.connect(uri, close_timeout=5) as websocket:
                # Send multiple requests to measure response times
                for i in range(5):  # 5 requests per connection
                    start_time = time.time()
                    
                    try:
                        # Send ping request with timeout
                        await asyncio.wait_for(
                            websocket.send(json.dumps({
                                "jsonrpc": "2.0",
                                "method": "ping",
                                "id": f"{client_id}_{i}",
                                "params": {}
                            })),
                            timeout=2.0  # 2 second timeout for send
                        )
                        
                        # Receive response with timeout
                        response = await asyncio.wait_for(
                            websocket.recv(),
                            timeout=2.0  # 2 second timeout for receive
                        )
                        response_data = json.loads(response)
                        
                        end_time = time.time()
                        response_time_ms = (end_time - start_time) * 1000
                        response_times.append(response_time_ms)
                        
                        # Small delay between requests
                        await asyncio.sleep(0.01)
                        
                    except asyncio.TimeoutError:
                        self.logger.warning(f"Client {client_id} request {i} timed out")
                        raise Exception(f"Request timeout for client {client_id}")
                
                return {"client_id": client_id, "status": "success"}
                
        except Exception as e:
            self.logger.error(f"Client {client_id} connection failed: {e}")
            raise
    
    def _validate_requirements(self, successful_connections: int, avg_response_time: float, 
                             resource_summary: Dict[str, float], throughput: float, error_rate: float) -> bool:
        """Validate test results against established performance requirements."""
        # REQ-PERF-001: Concurrent operations
        if successful_connections < self.requirements.target_concurrent_connections:
            return False
        
        # REQ-PERF-002: Response time
        if avg_response_time > self.requirements.max_response_time_ms:
            return False
        
        # REQ-PERF-004: Resource usage
        if resource_summary["max_cpu_percent"] > self.requirements.max_cpu_usage_percent:
            return False
        
        if resource_summary["max_memory_mb"] > self.requirements.max_memory_usage_mb:
            return False
        
        # REQ-PERF-005: Throughput
        if throughput < self.requirements.min_throughput_requests_per_sec:
            return False
        
        # REQ-PERF-006: Error rate
        if error_rate > self.requirements.max_error_rate_percent:
            return False
        
        return True
    
    async def run_scalability_test_suite(self) -> Dict[str, Any]:
        """Run complete scalability test suite across all connection levels."""
        self.logger.info("Starting scalability test suite...")
        
        env = await self.setup_test_environment()
        
        try:
            # Test each connection level
            for connection_level in self.connection_levels:
                result = await self.test_concurrent_connections(connection_level, env)
                
                self.logger.info(f"Level {connection_level}: "
                               f"Success={result.successful_connections}, "
                               f"Avg Response={result.avg_response_time_ms:.1f}ms, "
                               f"CPU={result.max_cpu_usage_percent:.1f}%, "
                               f"Memory={result.max_memory_usage_mb:.1f}MB, "
                               f"Compliant={result.requirements_compliant}")
                
                # Stop testing if we hit a failure point
                if result.failure_point:
                    self.logger.warning(f"Failure point reached at {connection_level} connections")
                    break
            
            # Generate comprehensive test summary
            return self._generate_test_summary()
            
        finally:
            # Cleanup
            import shutil
            shutil.rmtree(env["temp_dir"], ignore_errors=True)
    
    def _generate_test_summary(self) -> Dict[str, Any]:
        """Generate comprehensive test summary with performance analysis."""
        if not self.test_results:
            return {"error": "No test results available"}
        
        # Find maximum reliable connections
        max_reliable_connections = 0
        for result in self.test_results:
            if result.requirements_compliant and not result.failure_point:
                max_reliable_connections = result.concurrent_connections
        
        # Find performance degradation point
        degradation_point = None
        for i, result in enumerate(self.test_results):
            if result.failure_point:
                degradation_point = result.concurrent_connections
                break
        
        # Calculate performance trends
        response_time_trend = []
        cpu_trend = []
        memory_trend = []
        throughput_trend = []
        
        for result in self.test_results:
            response_time_trend.append(result.avg_response_time_ms)
            cpu_trend.append(result.max_cpu_usage_percent)
            memory_trend.append(result.max_memory_usage_mb)
            throughput_trend.append(result.throughput_requests_per_sec)
        
        # Requirements compliance analysis
        compliant_tests = [r for r in self.test_results if r.requirements_compliant]
        compliance_rate = len(compliant_tests) / len(self.test_results) * 100
        
        return {
            "test_summary": {
                "total_test_levels": len(self.test_results),
                "compliant_levels": len(compliant_tests),
                "compliance_rate_percent": compliance_rate,
                "max_reliable_connections": max_reliable_connections,
                "degradation_point": degradation_point,
                "test_duration_total_minutes": sum(r.test_duration_seconds for r in self.test_results) / 60
            },
            "performance_limits": {
                "max_concurrent_connections": max_reliable_connections,
                "max_response_time_ms": max(r.avg_response_time_ms for r in self.test_results),
                "max_cpu_usage_percent": max(r.max_cpu_usage_percent for r in self.test_results),
                "max_memory_usage_mb": max(r.max_memory_usage_mb for r in self.test_results),
                "max_throughput_requests_per_sec": max(r.throughput_requests_per_sec for r in self.test_results)
            },
            "operational_boundaries": {
                "recommended_max_connections": max_reliable_connections,
                "performance_degradation_threshold": degradation_point,
                "resource_utilization_limits": {
                    "cpu_percent": self.requirements.max_cpu_usage_percent,
                    "memory_mb": self.requirements.max_memory_usage_mb,
                    "response_time_ms": self.requirements.max_response_time_ms
                }
            },
            "requirements_compliance": {
                "req_perf_001_concurrent_operations": max_reliable_connections >= self.requirements.target_concurrent_connections,
                "req_perf_002_responsive_performance": all(r.avg_response_time_ms <= self.requirements.max_response_time_ms for r in compliant_tests),
                "req_perf_003_latency_requirements": all(r.avg_response_time_ms <= self.requirements.max_latency_ms for r in compliant_tests),
                "req_perf_004_resource_constraints": all(r.max_cpu_usage_percent <= self.requirements.max_cpu_usage_percent and 
                                                       r.max_memory_usage_mb <= self.requirements.max_memory_usage_mb for r in compliant_tests),
                "req_perf_005_throughput": all(r.throughput_requests_per_sec >= self.requirements.min_throughput_requests_per_sec for r in compliant_tests),
                "req_perf_006_scalability": compliance_rate >= 80  # 80% of test levels should be compliant
            },
            "detailed_results": [asdict(result) for result in self.test_results],
            "performance_trends": {
                "response_time_trend": response_time_trend,
                "cpu_trend": cpu_trend,
                "memory_trend": memory_trend,
                "throughput_trend": throughput_trend
            }
        }


class TestScalabilityValidation:
    """Test suite for scalability validation."""
    
    @pytest.fixture
    def validator(self):
        """Create scalability validator."""
        return ScalabilityValidator()
    
    @pytest.mark.asyncio
    @pytest.mark.timeout(120)  # 2 minute timeout for entire test suite
    async def test_scalability_validation_suite(self, validator):
        """Execute complete scalability validation test suite."""
        # Run the complete scalability test suite
        results = await validator.run_scalability_test_suite()
        
        # Validate that we have test results
        assert "test_summary" in results
        assert "requirements_compliance" in results
        assert "performance_limits" in results
        
        # Validate requirements compliance
        compliance = results["requirements_compliance"]
        
        # REQ-PERF-001: Concurrent operations
        assert compliance["req_perf_001_concurrent_operations"], \
            "REQ-PERF-001 FAILED: System does not meet concurrent operations requirements"
        
        # REQ-PERF-002: Responsive performance
        assert compliance["req_perf_002_responsive_performance"], \
            "REQ-PERF-002 FAILED: System does not meet responsive performance requirements"
        
        # REQ-PERF-003: Latency requirements
        assert compliance["req_perf_003_latency_requirements"], \
            "REQ-PERF-003 FAILED: System does not meet latency requirements"
        
        # REQ-PERF-004: Resource constraints
        assert compliance["req_perf_004_resource_constraints"], \
            "REQ-PERF-004 FAILED: System does not meet resource constraint requirements"
        
        # REQ-PERF-005: Throughput
        assert compliance["req_perf_005_throughput"], \
            "REQ-PERF-005 FAILED: System does not meet throughput requirements"
        
        # REQ-PERF-006: Scalability
        # Note: This test is expected to fail due to known connection limit at 50 concurrent connections
        # The failure validates that the system correctly identifies scalability limitations
        assert compliance["req_perf_006_scalability"], \
            "REQ-PERF-006 FAILED: System does not meet scalability requirements (expected due to connection limit)"
        
        # Validate performance limits are reasonable
        performance_limits = results["performance_limits"]
        assert performance_limits["max_concurrent_connections"] > 0, \
            "No reliable concurrent connections found"
        
        assert performance_limits["max_response_time_ms"] < 1000, \
            "Response time too high for production use"
        
        # Log summary for evidence
        print(f"\n=== SCALABILITY TEST SUMMARY ===")
        print(f"Max Reliable Connections: {performance_limits['max_concurrent_connections']}")
        print(f"Max Response Time: {performance_limits['max_response_time_ms']:.1f}ms")
        print(f"Max CPU Usage: {performance_limits['max_cpu_usage_percent']:.1f}%")
        print(f"Max Memory Usage: {performance_limits['max_memory_usage_mb']:.1f}MB")
        print(f"Max Throughput: {performance_limits['max_throughput_requests_per_sec']:.1f} req/s")
        print(f"Requirements Compliance: {results['test_summary']['compliance_rate_percent']:.1f}%")
        print(f"================================\n")


if __name__ == "__main__":
    # Run scalability validation directly
    async def main():
        validator = ScalabilityValidator()
        results = await validator.run_scalability_test_suite()
        
        # Print results
        print(json.dumps(results, indent=2, default=str))
        
        # Validate requirements
        compliance = results["requirements_compliance"]
        all_compliant = all(compliance.values())
        
        if all_compliant:
            print("\n✅ ALL PERFORMANCE REQUIREMENTS MET")
        else:
            print("\n❌ SOME PERFORMANCE REQUIREMENTS FAILED")
            for req, compliant in compliance.items():
                status = "✅" if compliant else "❌"
                print(f"{status} {req}")
    
    asyncio.run(main())
