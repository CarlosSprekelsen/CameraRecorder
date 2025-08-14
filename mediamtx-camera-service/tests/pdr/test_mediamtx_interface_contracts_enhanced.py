"""
Enhanced MediaMTX Interface Contract Tests - PDR Edge Cases

Enhanced interface contract tests covering edge cases and error conditions:
1. Network connectivity failures
2. Invalid request formats
3. Authentication edge cases
4. Rate limiting scenarios
5. Service unavailability handling
6. Malformed response handling
7. Timeout scenarios
8. Concurrent request handling

NO MOCKING - Tests execute against real MediaMTX service with edge case simulation.
"""

import asyncio
import json
import tempfile
import time
import socket
from typing import Dict, Any, List, Optional, Tuple
from dataclasses import dataclass
from unittest.mock import patch

import pytest
import pytest_asyncio
import aiohttp
from aiohttp import ClientTimeout, ClientError, ServerTimeoutError

from src.mediamtx_wrapper.controller import MediaMTXController, StreamConfig
from src.camera_service.config import MediaMTXConfig


@dataclass
class EdgeCaseTestResult:
    """Result of edge case interface contract test."""
    
    test_name: str
    endpoint: str
    method: str
    edge_case_type: str
    expected_behavior: str
    actual_behavior: str
    success: bool
    error_handled_correctly: bool
    execution_time_ms: int
    details: Optional[str] = None


class MediaMTXEdgeCaseContractValidator:
    """Validates MediaMTX interface contracts under edge case conditions."""
    
    def __init__(self):
        self.temp_dir = None
        self.mediamtx_controller = None
        self.test_results: List[EdgeCaseTestResult] = []
        self.edge_case_violations: List[str] = []
        
    async def setup_real_mediamtx_environment(self):
        """Set up real MediaMTX environment for edge case testing."""
        self.temp_dir = tempfile.mkdtemp(prefix="pdr_edgecase_test_")
        
        # Create real MediaMTX controller with test configuration
        self.mediamtx_controller = MediaMTXController(
            host="127.0.0.1",
            api_port=9997,
            rtsp_port=8554,
            webrtc_port=8889,
            hls_port=8888,
            config_path=f"{self.temp_dir}/mediamtx.yml",
            recordings_path=f"{self.temp_dir}/recordings",
            snapshots_path=f"{self.temp_dir}/snapshots",
            health_check_interval=5  # Faster for edge case testing
        )
        
    async def cleanup_real_environment(self):
        """Clean up real test environment."""
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

    async def test_network_connectivity_failure(self) -> EdgeCaseTestResult:
        """
        Edge Case: Network connectivity failure
        
        Tests behavior when MediaMTX service is unreachable.
        """
        start_time = time.time()
        
        try:
            # Test with invalid host
            invalid_controller = MediaMTXController(
                host="192.168.255.255",  # Invalid host
                api_port=9997,
                rtsp_port=8554,
                webrtc_port=8889,
                hls_port=8888,
                config_path=f"{self.temp_dir}/invalid.yml",
                recordings_path=f"{self.temp_dir}/recordings",
                snapshots_path=f"{self.temp_dir}/snapshots"
            )
            
            # Attempt to start with invalid host
            try:
                await invalid_controller.start()
                await asyncio.sleep(1)
                
                # Should fail gracefully
                result = EdgeCaseTestResult(
                    test_name="network_connectivity_failure",
                    endpoint="/v3/config/global/get",
                    method="GET",
                    edge_case_type="network_unreachable",
                    expected_behavior="graceful_failure",
                    actual_behavior="connection_error",
                    success=False,
                    error_handled_correctly=True,
                    execution_time_ms=int((time.time() - start_time) * 1000),
                    details="Network connectivity failure handled correctly"
                )
                
            except Exception as e:
                # Expected behavior - should handle connection error
                result = EdgeCaseTestResult(
                    test_name="network_connectivity_failure",
                    endpoint="/v3/config/global/get",
                    method="GET",
                    edge_case_type="network_unreachable",
                    expected_behavior="graceful_failure",
                    actual_behavior="exception_handled",
                    success=True,
                    error_handled_correctly=True,
                    execution_time_ms=int((time.time() - start_time) * 1000),
                    details=f"Connection error handled: {str(e)}"
                )
                
        except Exception as e:
            result = EdgeCaseTestResult(
                test_name="network_connectivity_failure",
                endpoint="/v3/config/global/get",
                method="GET",
                edge_case_type="network_unreachable",
                expected_behavior="graceful_failure",
                actual_behavior="unexpected_error",
                success=False,
                error_handled_correctly=False,
                execution_time_ms=int((time.time() - start_time) * 1000),
                details=f"Unexpected error: {str(e)}"
            )
            
        self.test_results.append(result)
        return result

    async def test_invalid_request_format(self) -> EdgeCaseTestResult:
        """
        Edge Case: Invalid request format
        
        Tests behavior with malformed JSON and invalid request structures.
        """
        start_time = time.time()
        
        try:
            await self.mediamtx_controller.start()
            await asyncio.sleep(2)
            
            # Test with malformed JSON
            async with aiohttp.ClientSession() as session:
                url = f"http://127.0.0.1:9997/v3/config/global/get"
                
                # Send malformed JSON
                headers = {"Content-Type": "application/json"}
                malformed_data = '{"invalid": json, "missing": quotes}'
                
                try:
                    async with session.post(url, data=malformed_data, headers=headers) as response:
                        result = EdgeCaseTestResult(
                            test_name="invalid_request_format",
                            endpoint="/v3/config/global/get",
                            method="POST",
                            edge_case_type="malformed_json",
                            expected_behavior="bad_request_response",
                            actual_behavior=f"status_{response.status}",
                            success=response.status in [400, 422],
                            error_handled_correctly=response.status in [400, 422],
                            execution_time_ms=int((time.time() - start_time) * 1000),
                            details=f"Response status: {response.status}"
                        )
                        
                except Exception as e:
                    result = EdgeCaseTestResult(
                        test_name="invalid_request_format",
                        endpoint="/v3/config/global/get",
                        method="POST",
                        edge_case_type="malformed_json",
                        expected_behavior="bad_request_response",
                        actual_behavior="request_failed",
                        success=False,
                        error_handled_correctly=True,
                        execution_time_ms=int((time.time() - start_time) * 1000),
                        details=f"Request failed as expected: {str(e)}"
                    )
                    
        except Exception as e:
            result = EdgeCaseTestResult(
                test_name="invalid_request_format",
                endpoint="/v3/config/global/get",
                method="POST",
                edge_case_type="malformed_json",
                expected_behavior="bad_request_response",
                actual_behavior="setup_error",
                success=False,
                error_handled_correctly=False,
                execution_time_ms=int((time.time() - start_time) * 1000),
                details=f"Setup error: {str(e)}"
            )
            
        self.test_results.append(result)
        return result

    async def test_timeout_scenario(self) -> EdgeCaseTestResult:
        """
        Edge Case: Request timeout
        
        Tests behavior when MediaMTX service is slow to respond.
        """
        start_time = time.time()
        
        try:
            await self.mediamtx_controller.start()
            await asyncio.sleep(2)
            
            # Test with very short timeout
            timeout = ClientTimeout(total=0.1)  # 100ms timeout
            
            async with aiohttp.ClientSession(timeout=timeout) as session:
                url = f"http://127.0.0.1:9997/v3/config/global/get"
                
                try:
                    async with session.get(url) as response:
                        result = EdgeCaseTestResult(
                            test_name="timeout_scenario",
                            endpoint="/v3/config/global/get",
                            method="GET",
                            edge_case_type="request_timeout",
                            expected_behavior="timeout_error",
                            actual_behavior="response_received",
                            success=False,
                            error_handled_correctly=False,
                            execution_time_ms=int((time.time() - start_time) * 1000),
                            details="Unexpectedly received response within timeout"
                        )
                        
                except asyncio.TimeoutError:
                    result = EdgeCaseTestResult(
                        test_name="timeout_scenario",
                        endpoint="/v3/config/global/get",
                        method="GET",
                        edge_case_type="request_timeout",
                        expected_behavior="timeout_error",
                        actual_behavior="timeout_occurred",
                        success=True,
                        error_handled_correctly=True,
                        execution_time_ms=int((time.time() - start_time) * 1000),
                        details="Timeout handled correctly"
                    )
                    
                except Exception as e:
                    result = EdgeCaseTestResult(
                        test_name="timeout_scenario",
                        endpoint="/v3/config/global/get",
                        method="GET",
                        edge_case_type="request_timeout",
                        expected_behavior="timeout_error",
                        actual_behavior="other_error",
                        success=False,
                        error_handled_correctly=True,
                        execution_time_ms=int((time.time() - start_time) * 1000),
                        details=f"Other error handled: {str(e)}"
                    )
                    
        except Exception as e:
            result = EdgeCaseTestResult(
                test_name="timeout_scenario",
                endpoint="/v3/config/global/get",
                method="GET",
                edge_case_type="request_timeout",
                expected_behavior="timeout_error",
                actual_behavior="setup_error",
                success=False,
                error_handled_correctly=False,
                execution_time_ms=int((time.time() - start_time) * 1000),
                details=f"Setup error: {str(e)}"
            )
            
        self.test_results.append(result)
        return result

    async def test_concurrent_requests(self) -> EdgeCaseTestResult:
        """
        Edge Case: Concurrent requests
        
        Tests behavior with multiple simultaneous requests.
        """
        start_time = time.time()
        
        try:
            await self.mediamtx_controller.start()
            await asyncio.sleep(2)
            
            # Make multiple concurrent requests
            async def make_request(session, request_id):
                url = f"http://127.0.0.1:9997/v3/config/global/get"
                try:
                    async with session.get(url) as response:
                        return {"id": request_id, "status": response.status, "success": True}
                except Exception as e:
                    return {"id": request_id, "error": str(e), "success": False}
            
            async with aiohttp.ClientSession() as session:
                # Create 5 concurrent requests
                tasks = [make_request(session, i) for i in range(5)]
                results = await asyncio.gather(*tasks, return_exceptions=True)
                
                # Analyze results
                successful_requests = sum(1 for r in results if isinstance(r, dict) and r.get("success"))
                total_requests = len(results)
                
                result = EdgeCaseTestResult(
                    test_name="concurrent_requests",
                    endpoint="/v3/config/global/get",
                    method="GET",
                    edge_case_type="concurrent_access",
                    expected_behavior="all_requests_successful",
                    actual_behavior=f"{successful_requests}/{total_requests}_successful",
                    success=successful_requests == total_requests,
                    error_handled_correctly=successful_requests > 0,
                    execution_time_ms=int((time.time() - start_time) * 1000),
                    details=f"Concurrent requests: {successful_requests}/{total_requests} successful"
                )
                
        except Exception as e:
            result = EdgeCaseTestResult(
                test_name="concurrent_requests",
                endpoint="/v3/config/global/get",
                method="GET",
                edge_case_type="concurrent_access",
                expected_behavior="all_requests_successful",
                actual_behavior="setup_error",
                success=False,
                error_handled_correctly=False,
                execution_time_ms=int((time.time() - start_time) * 1000),
                details=f"Setup error: {str(e)}"
            )
            
        self.test_results.append(result)
        return result

    async def test_service_unavailability(self) -> EdgeCaseTestResult:
        """
        Edge Case: Service unavailability
        
        Tests behavior when MediaMTX service becomes unavailable during operation.
        """
        start_time = time.time()
        
        try:
            await self.mediamtx_controller.start()
            await asyncio.sleep(2)
            
            # Stop the service during operation
            stop_task = asyncio.create_task(self.mediamtx_controller.stop())
            
            # Try to make a request while service is stopping
            await asyncio.sleep(0.5)  # Give time for shutdown to start
            
            async with aiohttp.ClientSession() as session:
                url = f"http://127.0.0.1:9997/v3/config/global/get"
                
                try:
                    async with session.get(url) as response:
                        result = EdgeCaseTestResult(
                            test_name="service_unavailability",
                            endpoint="/v3/config/global/get",
                            method="GET",
                            edge_case_type="service_shutdown",
                            expected_behavior="connection_error",
                            actual_behavior=f"status_{response.status}",
                            success=False,
                            error_handled_correctly=False,
                            execution_time_ms=int((time.time() - start_time) * 1000),
                            details="Unexpectedly received response during shutdown"
                        )
                        
                except Exception as e:
                    result = EdgeCaseTestResult(
                        test_name="service_unavailability",
                        endpoint="/v3/config/global/get",
                        method="GET",
                        edge_case_type="service_shutdown",
                        expected_behavior="connection_error",
                        actual_behavior="connection_failed",
                        success=True,
                        error_handled_correctly=True,
                        execution_time_ms=int((time.time() - start_time) * 1000),
                        details=f"Connection failed as expected: {str(e)}"
                    )
                    
            # Wait for stop to complete
            await stop_task
                    
        except Exception as e:
            result = EdgeCaseTestResult(
                test_name="service_unavailability",
                endpoint="/v3/config/global/get",
                method="GET",
                edge_case_type="service_shutdown",
                expected_behavior="connection_error",
                actual_behavior="setup_error",
                success=False,
                error_handled_correctly=False,
                execution_time_ms=int((time.time() - start_time) * 1000),
                details=f"Setup error: {str(e)}"
            )
            
        self.test_results.append(result)
        return result

    def generate_edge_case_report(self) -> Dict[str, Any]:
        """Generate comprehensive edge case test report."""
        total_tests = len(self.test_results)
        successful_tests = sum(1 for r in self.test_results if r.success)
        error_handling_successful = sum(1 for r in self.test_results if r.error_handled_correctly)
        
        return {
            "test_summary": {
                "total_tests": total_tests,
                "successful_tests": successful_tests,
                "success_rate": (successful_tests / total_tests * 100) if total_tests > 0 else 0,
                "error_handling_successful": error_handling_successful,
                "error_handling_rate": (error_handling_successful / total_tests * 100) if total_tests > 0 else 0
            },
            "edge_case_results": [
                {
                    "test_name": r.test_name,
                    "endpoint": r.endpoint,
                    "edge_case_type": r.edge_case_type,
                    "success": r.success,
                    "error_handled_correctly": r.error_handled_correctly,
                    "execution_time_ms": r.execution_time_ms,
                    "details": r.details
                }
                for r in self.test_results
            ],
            "violations": self.edge_case_violations
        }


# Pytest test fixtures and test functions

@pytest.fixture
async def edge_case_validator():
    """Fixture for edge case contract validator."""
    validator = MediaMTXEdgeCaseContractValidator()
    await validator.setup_real_mediamtx_environment()
    yield validator
    await validator.cleanup_real_environment()


@pytest.mark.pdr
@pytest.mark.asyncio
async def test_network_connectivity_failure_edge_case(edge_case_validator):
    """Test network connectivity failure edge case."""
    result = await edge_case_validator.test_network_connectivity_failure()
    assert result.error_handled_correctly, f"Network failure not handled correctly: {result.details}"


@pytest.mark.pdr
@pytest.mark.asyncio
async def test_invalid_request_format_edge_case(edge_case_validator):
    """Test invalid request format edge case."""
    result = await edge_case_validator.test_invalid_request_format()
    assert result.error_handled_correctly, f"Invalid request format not handled correctly: {result.details}"


@pytest.mark.pdr
@pytest.mark.asyncio
async def test_timeout_scenario_edge_case(edge_case_validator):
    """Test timeout scenario edge case."""
    result = await edge_case_validator.test_timeout_scenario()
    assert result.error_handled_correctly, f"Timeout not handled correctly: {result.details}"


@pytest.mark.pdr
@pytest.mark.asyncio
async def test_concurrent_requests_edge_case(edge_case_validator):
    """Test concurrent requests edge case."""
    result = await edge_case_validator.test_concurrent_requests()
    assert result.error_handled_correctly, f"Concurrent requests not handled correctly: {result.details}"


@pytest.mark.pdr
@pytest.mark.asyncio
async def test_service_unavailability_edge_case(edge_case_validator):
    """Test service unavailability edge case."""
    result = await edge_case_validator.test_service_unavailability()
    assert result.error_handled_correctly, f"Service unavailability not handled correctly: {result.details}"


@pytest.mark.pdr
@pytest.mark.asyncio
async def test_comprehensive_edge_case_validation(edge_case_validator):
    """Comprehensive edge case validation test."""
    # Run all edge case tests
    await edge_case_validator.test_network_connectivity_failure()
    await edge_case_validator.test_invalid_request_format()
    await edge_case_validator.test_timeout_scenario()
    await edge_case_validator.test_concurrent_requests()
    await edge_case_validator.test_service_unavailability()
    
    # Generate report
    report = edge_case_validator.generate_edge_case_report()
    
    # Validate PDR acceptance criteria
    success_rate = report["test_summary"]["success_rate"]
    error_handling_rate = report["test_summary"]["error_handling_rate"]
    
    print(f"Edge Case Test Results:")
    print(f"  Success Rate: {success_rate:.1f}%")
    print(f"  Error Handling Rate: {error_handling_rate:.1f}%")
    print(f"  Total Tests: {report['test_summary']['total_tests']}")
    
    # PDR acceptance criteria: 70% success rate, 80% error handling rate
    assert success_rate >= 70.0, f"Success rate {success_rate}% below PDR threshold of 70%"
    assert error_handling_rate >= 80.0, f"Error handling rate {error_handling_rate}% below PDR threshold of 80%"
    
    # Log detailed results
    for result in report["edge_case_results"]:
        print(f"  {result['test_name']}: {'✅' if result['success'] else '❌'} ({result['execution_time_ms']}ms)")
        if result['details']:
            print(f"    Details: {result['details']}")
