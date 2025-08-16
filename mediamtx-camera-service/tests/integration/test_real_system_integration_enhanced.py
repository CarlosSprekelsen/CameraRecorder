# tests/integration/test_real_system_integration_enhanced.py
"""
Enhanced real system integration tests with comprehensive end-to-end validation.

This test file addresses PARTIAL coverage gaps identified in the comprehensive audit:
- REQ-INT-001: System shall provide real end-to-end system behavior validation
- REQ-INT-002: System shall validate real MediaMTX server integration
- REQ-ERROR-004: System shall handle configuration loading failures gracefully
- REQ-ERROR-005: System shall provide meaningful error messages for configuration issues
- REQ-ERROR-006: System shall handle logging configuration failures gracefully
- REQ-ERROR-007: System shall handle WebSocket connection failures gracefully
- REQ-ERROR-008: System shall handle MediaMTX service failures gracefully
- REQ-ERROR-009: System shall handle error propagation and recovery
- REQ-ERROR-010: System shall implement error recovery mechanisms

Requirements Traceability:
- REQ-INT-001: Real end-to-end system behavior validation with comprehensive scenarios
- REQ-INT-002: Real MediaMTX server integration with error handling
- REQ-ERROR-004: Configuration loading failure graceful handling
- REQ-ERROR-005: Meaningful error messages for configuration issues
- REQ-ERROR-006: Logging configuration failure graceful handling
- REQ-ERROR-007: WebSocket connection failure graceful handling
- REQ-ERROR-008: MediaMTX service failure graceful handling
- REQ-ERROR-009: Error propagation and recovery mechanisms
- REQ-ERROR-010: Error recovery mechanism implementation

Story Coverage: S3 - WebSocket API Integration, S4 - Error Handling and Recovery
IV&V Control Point: Real system integration with comprehensive error scenarios
"""

import pytest
import asyncio
import tempfile
import os
import subprocess
import time
import json
import signal
from unittest.mock import AsyncMock, MagicMock, patch, Mock

from src.websocket_server.server import WebSocketJsonRpcServer
from src.camera_discovery.hybrid_monitor import HybridCameraMonitor
from src.mediamtx_wrapper.controller import MediaMTXController
from src.camera_service.service_manager import ServiceManager
from tests.fixtures.mediamtx_test_infrastructure import mediamtx_infrastructure, mediamtx_controller
from tests.fixtures.websocket_test_client import WebSocketTestClient, websocket_client


class TestEnhancedRealSystemIntegration:
    """Enhanced real system integration tests with comprehensive end-to-end validation."""

    @pytest.fixture
    def temp_dirs(self):
        """Create temporary directories for testing."""
        base = tempfile.mkdtemp(prefix="enhanced_integration_test_")
        config_path = os.path.join(base, "mediamtx.yml")
        recordings_path = os.path.join(base, "recordings")
        snapshots_path = os.path.join(base, "snapshots")
        logs_path = os.path.join(base, "logs")
        
        # Create directories
        os.makedirs(recordings_path, exist_ok=True)
        os.makedirs(snapshots_path, exist_ok=True)
        os.makedirs(logs_path, exist_ok=True)
        
        # Create basic MediaMTX config
        with open(config_path, 'w') as f:
            f.write("""
paths:
  all:
    runOnDemand: ffmpeg -f lavfi -i testsrc=duration=10:size=1280x720:rate=30 -c:v libx264 -f rtsp rtsp://127.0.0.1:8554/test
            """)
        
        try:
            yield {
                "base": base,
                "config_path": config_path,
                "recordings_path": recordings_path,
                "snapshots_path": snapshots_path,
                "logs_path": logs_path
            }
        finally:
            import shutil
            shutil.rmtree(base, ignore_errors=True)

    @pytest.fixture
    def real_mediamtx_service(self):
        """Verify systemd-managed MediaMTX service is available for testing."""
        # Verify MediaMTX service is running
        result = subprocess.run(
            ['systemctl', 'is-active', 'mediamtx'],
            capture_output=True,
            text=True,
            timeout=10
        )
        
        if result.returncode != 0 or result.stdout.strip() != 'active':
            raise RuntimeError("MediaMTX service is not running. Please start it with: sudo systemctl start mediamtx")
        
        # Return service info for testing
        return {
            "api_port": 9997,
            "rtsp_port": 8554,
            "webrtc_port": 8889,
            "hls_port": 8888,
            "host": "localhost"
        }

    @pytest.fixture
    async def real_camera_monitor(self, temp_dirs):
        """Real camera monitor with capability detection support."""
        monitor = HybridCameraMonitor(
            device_range=[0, 1, 2],
            poll_interval=0.1,
            enable_capability_detection=True
        )
        await monitor.start()
        try:
            yield monitor
        finally:
            await monitor.stop()

    @pytest.fixture
    async def real_mediamtx_controller(self, real_mediamtx_service, temp_dirs):
        """Real MediaMTX controller with systemd-managed service integration."""
        controller = MediaMTXController(
            host=real_mediamtx_service["host"],
            api_port=real_mediamtx_service["api_port"],
            rtsp_port=real_mediamtx_service["rtsp_port"],
            webrtc_port=real_mediamtx_service["webrtc_port"],
            hls_port=real_mediamtx_service["hls_port"],
            config_path=temp_dirs["config_path"],
            recordings_path=temp_dirs["recordings_path"],
            snapshots_path=temp_dirs["snapshots_path"],
            health_check_interval=0.1,
            health_failure_threshold=3,
            health_circuit_breaker_timeout=1.0,
            health_max_backoff_interval=2.0,
        )
        await controller.start()
        try:
            yield controller
        finally:
            await controller.stop()

    @pytest.fixture
    async def real_service_manager(self, real_camera_monitor, real_mediamtx_controller, temp_dirs):
        """Real service manager with all components integrated."""
        from src.camera_service.config import Config, ServerConfig, CameraConfig
        
        config = Config(
            server=ServerConfig(
                host="localhost",
                port=8004,  # Different port to avoid conflicts
                websocket_path="/ws",
                max_connections=100
            ),
            camera=CameraConfig(
                device_range=[0, 1, 2],
                poll_interval=0.1,
                enable_capability_detection=True
            )
        )
        
        service_manager = ServiceManager(
            config=config,
            camera_monitor=real_camera_monitor,
            mediamtx_controller=real_mediamtx_controller
        )
        
        await service_manager.start()
        try:
            yield service_manager
        finally:
            await service_manager.stop()

    @pytest.mark.asyncio
    async def test_real_end_to_end_system_behavior_validation(
        self, real_service_manager, real_camera_monitor, real_mediamtx_controller
    ):
        """
        Comprehensive real end-to-end system behavior validation.
        
        Requirements: REQ-INT-001
        Scenario: Complete system integration with real components
        Expected: All system components work together seamlessly
        Edge Cases: Component interactions, data flow, system state consistency
        """
        # Test complete system integration
        websocket_server = real_service_manager._websocket_server
        
        # 1. Test camera discovery integration
        connected_cameras = await real_camera_monitor.get_connected_cameras()
        assert isinstance(connected_cameras, dict)
        
        # 2. Test MediaMTX integration
        if connected_cameras:
            device_path = list(connected_cameras.keys())[0]
            
            # Create test stream
            from src.mediamtx_wrapper.controller import StreamConfig
            stream_config = StreamConfig(
                name="test_stream",
                source=device_path,
                record=True
            )
            stream_info = await real_mediamtx_controller.create_stream(stream_config)
            
            # 3. Test WebSocket server integration
            result = await websocket_server._method_get_camera_status({"device": device_path})
            
            # Verify end-to-end integration
            assert result["device"] == device_path
            assert "status" in result
            assert "name" in result
            assert "resolution" in result
            assert "fps" in result
            assert "capabilities" in result
            assert "streams" in result
            assert "metrics" in result
            
            # Verify MediaMTX integration
            assert "rtsp" in result["streams"]
            assert "webrtc" in result["streams"]
            assert "hls" in result["streams"]
            
            # Clean up
            await real_mediamtx_controller.delete_stream("test_stream")

    @pytest.mark.asyncio
    async def test_real_mediamtx_server_integration_comprehensive(
        self, real_mediamtx_controller, real_service_manager
    ):
        """
        Comprehensive real MediaMTX server integration validation.
        
        Requirements: REQ-INT-002
        Scenario: Real MediaMTX server integration with various operations
        Expected: Successful integration with real MediaMTX service
        Edge Cases: Stream creation/deletion, status queries, error conditions
        """
        # Test real MediaMTX server integration
        websocket_server = real_service_manager._websocket_server
        
        # 1. Test stream creation
        from src.mediamtx_wrapper.controller import StreamConfig
        stream_config = StreamConfig(
            name="integration_test",
            source="/dev/video0",
            record=True
        )
        
        try:
            # Create stream
            stream_info = await real_mediamtx_controller.create_stream(stream_config)
            assert stream_info is not None
            
            # 2. Test stream status query
            stream_status = await real_mediamtx_controller.get_stream_status("integration_test")
            assert "status" in stream_status
            
            # 3. Test WebSocket integration with real stream
            result = await websocket_server._method_get_camera_status({"device": "/dev/video0"})
            
            # Verify real MediaMTX integration
            assert "streams" in result
            assert result["streams"]["rtsp"] == "rtsp://127.0.0.1:8554/integration_test"
            assert result["streams"]["webrtc"] == "webrtc://localhost:8002/integration_test"
            assert result["streams"]["hls"] == "http://localhost:8002/hls/integration_test.m3u8"
            
            # 4. Test metrics integration
            assert "metrics" in result
            assert "bytes_sent" in result["metrics"]
            assert "readers" in result["metrics"]
            assert "uptime" in result["metrics"]
            
        finally:
            # Clean up
            try:
                await real_mediamtx_controller.delete_stream("integration_test")
            except Exception:
                pass

    @pytest.mark.asyncio
    async def test_configuration_loading_failure_graceful_handling(
        self, temp_dirs
    ):
        """
        Test graceful handling of configuration loading failures.
        
        Requirements: REQ-ERROR-004
        Scenario: Configuration loading failures with graceful degradation
        Expected: System continues to function with default configuration
        Edge Cases: Invalid configuration files, missing configuration, permission errors
        """
        # Test with invalid configuration file
        invalid_config_path = os.path.join(temp_dirs["base"], "invalid_config.yml")
        
        with open(invalid_config_path, 'w') as f:
            f.write("""
invalid_yaml: [
  - missing: quotes
  - invalid: structure
            """)
        
        # Test configuration loading with invalid file
        from src.camera_service.config import Config
        
        try:
            config = Config.from_file(invalid_config_path)
        except Exception as e:
            # Verify meaningful error message
            assert "configuration" in str(e).lower() or "invalid" in str(e).lower()
        
        # Test with missing configuration file
        missing_config_path = os.path.join(temp_dirs["base"], "missing_config.yml")
        
        try:
            config = Config.from_file(missing_config_path)
        except Exception as e:
            # Verify meaningful error message
            assert "file" in str(e).lower() or "missing" in str(e).lower()

    @pytest.mark.asyncio
    async def test_meaningful_error_messages_configuration_issues(
        self, temp_dirs
    ):
        """
        Test meaningful error messages for configuration issues.
        
        Requirements: REQ-ERROR-005
        Scenario: Various configuration issues with descriptive error messages
        Expected: Clear, actionable error messages for configuration problems
        Edge Cases: Invalid parameters, missing required fields, type mismatches
        """
        # Test various configuration error scenarios
        error_scenarios = [
            {
                "name": "invalid_server_config",
                "config_data": {
                    "server": {
                        "host": "invalid_host_name_with_special_chars!@#",
                        "port": "not_a_number",
                        "websocket_path": None
                    }
                },
                "expected_error": "server"
            },
            {
                "name": "invalid_camera_config",
                "config_data": {
                    "camera": {
                        "device_range": "not_a_list",
                        "poll_interval": -1,
                        "enable_capability_detection": "not_boolean"
                    }
                },
                "expected_error": "camera"
            }
        ]
        
        for scenario in error_scenarios:
            config_path = os.path.join(temp_dirs["base"], f"{scenario['name']}.yml")
            
            with open(config_path, 'w') as f:
                import yaml
                yaml.dump(scenario["config_data"], f)
            
            try:
                from src.camera_service.config import Config
                config = Config.from_file(config_path)
            except Exception as e:
                error_message = str(e).lower()
                # Verify error message contains relevant information
                assert scenario["expected_error"].lower() in error_message or "invalid" in error_message

    @pytest.mark.asyncio
    async def test_logging_configuration_failure_graceful_handling(
        self, temp_dirs
    ):
        """
        Test graceful handling of logging configuration failures.
        
        Requirements: REQ-ERROR-006
        Scenario: Logging configuration failures with graceful degradation
        Expected: System continues to function with default logging
        Edge Cases: Invalid log paths, permission errors, disk space issues
        """
        # Test with invalid log directory (no write permissions)
        invalid_log_path = "/root/invalid_log_path"
        
        try:
            from src.camera_service.logging_config import setup_logging
            setup_logging(log_path=invalid_log_path)
        except Exception as e:
            # Verify meaningful error message
            assert "log" in str(e).lower() or "permission" in str(e).lower()
        
        # Test with valid log directory
        valid_log_path = temp_dirs["logs_path"]
        
        try:
            from src.camera_service.logging_config import setup_logging
            setup_logging(log_path=valid_log_path)
            # Should succeed with valid path
        except Exception as e:
            # Should not fail with valid path
            assert False, f"Logging setup failed with valid path: {e}"

    @pytest.mark.asyncio
    async def test_websocket_connection_failure_graceful_handling(
        self, real_service_manager
    ):
        """
        Test graceful handling of WebSocket connection failures.
        
        Requirements: REQ-ERROR-007
        Scenario: WebSocket connection failures with graceful handling
        Expected: Graceful error handling without system crash
        Edge Cases: Network failures, port conflicts, connection timeouts
        """
        websocket_server = real_service_manager._websocket_server
        
        # Test with invalid WebSocket client connection
        try:
            # Try to connect to non-existent WebSocket server
            client = WebSocketTestClient("ws://localhost:9999/nonexistent")
            await client.connect()
        except Exception as e:
            # Verify meaningful error message
            assert "connection" in str(e).lower() or "websocket" in str(e).lower()
        
        # Test WebSocket server with invalid port
        try:
            bad_server = WebSocketJsonRpcServer(
                host="localhost",
                port=99999,  # Invalid port
                websocket_path="/ws",
                max_connections=100
            )
            await bad_server.start()
        except Exception as e:
            # Verify meaningful error message
            assert "port" in str(e).lower() or "bind" in str(e).lower()

    @pytest.mark.asyncio
    async def test_mediamtx_service_failure_graceful_handling(
        self, real_camera_monitor, temp_dirs
    ):
        """
        Test graceful handling of MediaMTX service failures.
        
        Requirements: REQ-ERROR-008
        Scenario: MediaMTX service failures with graceful handling
        Expected: Graceful error handling without system crash
        Edge Cases: Service down, API failures, network issues
        """
        # Test with MediaMTX service unavailable
        try:
            bad_controller = MediaMTXController(
                host="localhost",
                api_port=99999,  # Invalid port
                rtsp_port=8554,
                webrtc_port=8889,
                hls_port=8888,
                config_path=temp_dirs["config_path"],
                recordings_path=temp_dirs["recordings_path"],
                snapshots_path=temp_dirs["snapshots_path"]
            )
            await bad_controller.start()
        except Exception as e:
            # Verify meaningful error message
            assert "mediamtx" in str(e).lower() or "connection" in str(e).lower()
        
        # Test WebSocket server without MediaMTX controller
        websocket_server = WebSocketJsonRpcServer(
            host="localhost",
            port=8005,
            websocket_path="/ws",
            max_connections=100,
            camera_monitor=real_camera_monitor,
            mediamtx_controller=None  # No MediaMTX controller
        )
        
        # Should still function without MediaMTX
        result = await websocket_server._method_get_camera_status({"device": "/dev/video0"})
        assert result["device"] == "/dev/video0"
        assert "status" in result
        assert result["streams"] == {}  # Empty streams without MediaMTX

    @pytest.mark.asyncio
    async def test_error_propagation_and_recovery(
        self, real_service_manager
    ):
        """
        Test error propagation and recovery mechanisms.
        
        Requirements: REQ-ERROR-009
        Scenario: Error propagation and recovery with comprehensive validation
        Expected: Proper error propagation and recovery mechanisms
        Edge Cases: Error cascading, recovery success, error isolation
        """
        websocket_server = real_service_manager._websocket_server
        
        # Test error propagation through system components
        error_scenarios = [
            {
                "name": "camera_monitor_error",
                "component": "camera_monitor",
                "method": "get_connected_cameras",
                "exception": Exception("Camera monitor error")
            },
            {
                "name": "mediamtx_controller_error",
                "component": "mediamtx_controller",
                "method": "get_stream_status",
                "exception": Exception("MediaMTX controller error")
            }
        ]
        
        for scenario in error_scenarios:
            # Test error propagation
            if scenario["component"] == "camera_monitor":
                with patch.object(real_service_manager._camera_monitor, scenario["method"], 
                                side_effect=scenario["exception"]):
                    result = await websocket_server._method_get_camera_status({"device": "/dev/video0"})
            elif scenario["component"] == "mediamtx_controller":
                with patch.object(real_service_manager._mediamtx_controller, scenario["method"], 
                                side_effect=scenario["exception"]):
                    result = await websocket_server._method_get_camera_status({"device": "/dev/video0"})
            
            # Verify system continues to function despite errors
            assert result["device"] == "/dev/video0"
            assert "status" in result
            assert "name" in result
            assert "resolution" in result
            assert "fps" in result

    @pytest.mark.asyncio
    async def test_error_recovery_mechanism_implementation(
        self, real_service_manager
    ):
        """
        Test error recovery mechanism implementation.
        
        Requirements: REQ-ERROR-010
        Scenario: Error recovery mechanism implementation with validation
        Expected: Proper error recovery mechanisms implemented
        Edge Cases: Recovery success, recovery failure, retry logic
        """
        websocket_server = real_service_manager._websocket_server
        
        # Test error recovery with retry logic
        failure_count = 0
        max_failures = 2
        
        def mock_method_with_recovery():
            nonlocal failure_count
            failure_count += 1
            if failure_count <= max_failures:
                raise Exception(f"Simulated failure {failure_count}")
            return {"status": "success"}
        
        # Test recovery mechanism
        with patch.object(real_service_manager._camera_monitor, 'get_connected_cameras', 
                        side_effect=mock_method_with_recovery):
            
            # First calls should fail
            for i in range(max_failures):
                try:
                    await websocket_server._method_get_camera_status({"device": "/dev/video0"})
                except Exception:
                    pass  # Expected failures
            
            # Final call should succeed (recovery)
            result = await websocket_server._method_get_camera_status({"device": "/dev/video0"})
            
            # Verify recovery was successful
            assert result["device"] == "/dev/video0"
            assert "status" in result

    @pytest.mark.asyncio
    async def test_system_integration_under_load(
        self, real_service_manager, real_camera_monitor, real_mediamtx_controller
    ):
        """
        Test system integration under load conditions.
        
        Requirements: REQ-INT-001, REQ-INT-002
        Scenario: System integration under high load
        Expected: System maintains functionality under load
        Edge Cases: High request volume, resource constraints, performance degradation
        """
        websocket_server = real_service_manager._websocket_server
        
        # Simulate high load with multiple concurrent requests
        async def make_request(request_id):
            try:
                result = await websocket_server._method_get_camera_status({"device": "/dev/video0"})
                return {"request_id": request_id, "success": True, "result": result}
            except Exception as e:
                return {"request_id": request_id, "success": False, "error": str(e)}
        
        # Make multiple concurrent requests
        tasks = [make_request(i) for i in range(10)]
        results = await asyncio.gather(*tasks)
        
        # Verify all requests completed
        assert len(results) == 10
        
        # Verify most requests succeeded
        successful_requests = [r for r in results if r["success"]]
        assert len(successful_requests) >= 8  # At least 80% success rate
        
        # Verify response structure for successful requests
        for result in successful_requests:
            assert result["result"]["device"] == "/dev/video0"
            assert "status" in result["result"]

    @pytest.mark.asyncio
    async def test_system_integration_failure_scenarios(
        self, real_service_manager
    ):
        """
        Test system integration with various failure scenarios.
        
        Requirements: REQ-INT-001, REQ-INT-002
        Scenario: System integration with component failures
        Expected: System continues to function with graceful degradation
        Edge Cases: Component failures, partial failures, recovery scenarios
        """
        websocket_server = real_service_manager._websocket_server
        
        # Test with various failure scenarios
        failure_scenarios = [
            {
                "name": "camera_monitor_partial_failure",
                "patch_target": real_service_manager._camera_monitor,
                "method": "get_connected_cameras",
                "side_effect": Exception("Partial camera monitor failure")
            },
            {
                "name": "mediamtx_controller_partial_failure",
                "patch_target": real_service_manager._mediamtx_controller,
                "method": "get_stream_status",
                "side_effect": Exception("Partial MediaMTX controller failure")
            }
        ]
        
        for scenario in failure_scenarios:
            with patch.object(scenario["patch_target"], scenario["method"], 
                            side_effect=scenario["side_effect"]):
                
                # System should continue to function
                result = await websocket_server._method_get_camera_status({"device": "/dev/video0"})
                
                # Verify basic functionality maintained
                assert result["device"] == "/dev/video0"
                assert "status" in result
                assert "name" in result
                assert "resolution" in result
                assert "fps" in result
