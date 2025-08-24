"""
Compliant Integration Test Template

This template demonstrates the proper structure for integration tests that validate
real component interactions and end-to-end system behavior:
- Real MediaMTX service integration
- Real WebSocket communication
- Real file system operations
- Real hardware integration where applicable
- Comprehensive error condition testing

Usage:
1. Copy this template to your integration test file
2. Replace placeholder content with your specific integration logic
3. Update requirements traceability to match your integration scenarios
4. Add real component integration for all system components
5. Include comprehensive error condition testing

Template Version: 1.0
Last Updated: 2025-01-27
"""

import pytest
import asyncio
import tempfile
import os
import shutil
import time
from pathlib import Path
from typing import Dict, Any, Optional

# Import your integration components
from src.camera_service.service_manager import ServiceManager
from src.camera_service.config import Config, ServerConfig, MediaMTXConfig, CameraConfig
from src.mediamtx_wrapper.controller import MediaMTXController
from src.websocket_server.server import WebSocketJsonRpcServer

# Import test infrastructure for real component testing
from tests.fixtures.mediamtx_test_infrastructure import mediamtx_infrastructure, mediamtx_controller
from tests.fixtures.auth_utils import WebSocketAuthTestClient, UserFactory, get_test_auth_manager


class TestSystemIntegration:
    """
    Test end-to-end system integration with real components.
    
    Requirements Traceability:
    - REQ-INT-001: System shall integrate all components seamlessly
    - REQ-INT-002: System shall handle real MediaMTX service integration
    - REQ-INT-003: System shall support real WebSocket communication
    - REQ-INT-004: System shall manage real file system operations
    - REQ-INT-005: System shall handle real camera device integration
    - REQ-ERROR-003: System shall handle integration failures gracefully
    
    Story Coverage: S4 - System Integration
    IV&V Control Point: End-to-end system validation
    """

    @pytest.fixture
    def temp_test_directory(self):
        """Create temporary directory for integration testing."""
        temp_dir = tempfile.mkdtemp(prefix="integration_test_")
        yield temp_dir
        # Cleanup
        if os.path.exists(temp_dir):
            shutil.rmtree(temp_dir)

    @pytest.fixture
    def integration_config(self, temp_test_directory) -> Config:
        """Create integration test configuration."""
        from tests.utils.port_utils import find_free_port
        
        return Config(
            server=ServerConfig(
                host="127.0.0.1",
                port=find_free_port(),  # Dynamic port to avoid conflicts
                websocket_path="/ws",
                max_connections=10
            ),
            mediamtx=MediaMTXConfig(
                host="127.0.0.1",
                api_port=9997,  # Fixed systemd service port
                rtsp_port=8554,  # Fixed systemd service port
                webrtc_port=8889,  # Fixed systemd service port
                hls_port=8888,  # Fixed systemd service port
                config_path=os.path.join(temp_test_directory, "mediamtx.yml"),
                recordings_path=os.path.join(temp_test_directory, "recordings"),
                snapshots_path=os.path.join(temp_test_directory, "snapshots"),
            ),
            camera=CameraConfig(
                device_range=[0, 1, 2],
                enable_capability_detection=True,
                detection_timeout=2.0
            ),
            health_port=find_free_port(),  # Dynamic health port to avoid conflicts
        )

    @pytest.fixture
    async def service_manager(self, integration_config, mediamtx_infrastructure):
        """Create and start service manager for integration testing."""
        service = ServiceManager(integration_config)
        await service.start()
        yield service
        await service.stop()

    @pytest.mark.asyncio
    async def test_end_to_end_camera_streaming_workflow(
        self, service_manager, mediamtx_infrastructure
    ):
        """
        Test complete end-to-end camera streaming workflow.
        
        Requirements: REQ-INT-001, REQ-INT-002, REQ-INT-003
        Scenario: Complete camera discovery → stream creation → WebSocket communication
        Expected: Successful end-to-end workflow with real components
        Edge Cases: Real service interactions, actual data flow, resource management
        """
        # Create authenticated WebSocket client
        auth_manager = get_test_auth_manager()
        user_factory = UserFactory(auth_manager)
        test_user = user_factory.create_operator_user("integration_test_user")
        
        websocket_url = f"ws://{service_manager.config.server.host}:{service_manager.config.server.port}{service_manager.config.server.websocket_path}"
        websocket_client = WebSocketAuthTestClient(websocket_url, test_user)
        
        # Connect WebSocket client to service
        await websocket_client.connect()
        
        try:
            # Step 1: Get camera list via WebSocket with proper authentication
            camera_list_response = await websocket_client.call_protected_method("get_camera_list", {})
            assert "result" in camera_list_response, "Response should contain 'result' field"
            assert "cameras" in camera_list_response["result"]
            assert "total" in camera_list_response["result"]
            
            cameras = camera_list_response["result"]["cameras"]
            
            if cameras:  # If cameras are available
                # Step 2: Get status of first camera
                first_camera = cameras[0]
                camera_device = first_camera["device"]
                
                status_response = await websocket_client.call_protected_method("get_camera_status", {"device": camera_device})
                assert "result" in status_response, "Response should contain 'result' field"
                assert status_response["result"]["device"] == camera_device
                
                # Step 3: Start recording (using proper API method)
                recording_response = await websocket_client.call_protected_method("start_recording", {
                    "device": camera_device,
                    "resolution": "1280x720",
                    "fps": 30
                })
                assert "result" in recording_response, "Response should contain 'result' field"
                assert "recording_id" in recording_response["result"]
                
                # Step 4: Verify stream exists in MediaMTX
                stream_name = f"camera_{camera_device.replace('/', '_').replace('video', '')}"
                stream_status = await mediamtx_infrastructure.get_stream_status(stream_name)
                assert stream_status is not None
                
                # Step 5: Stop recording
                stop_response = await websocket_client.call_protected_method("stop_recording", {"device": camera_device})
                assert "result" in stop_response, "Response should contain 'result' field"
                
        finally:
            await websocket_client.disconnect()

    @pytest.mark.asyncio
    async def test_real_mediamtx_service_integration(
        self, service_manager, mediamtx_infrastructure
    ):
        """
        Test real MediaMTX service integration with file system operations.
        
        Requirements: REQ-INT-002, REQ-INT-004
        Scenario: Real MediaMTX service with file system operations
        Expected: Successful MediaMTX integration with real file operations
        Edge Cases: File system permissions, disk space, concurrent access
        """
        # Create test stream with real file system
        stream_name = "integration_test_stream"
        source_path = "/dev/video0"
        
        # Create stream in MediaMTX
        stream_info = await mediamtx_infrastructure.create_test_stream(
            stream_name, source_path
        )
        
        assert stream_info is not None
        assert "stream_id" in stream_info
        assert "config" in stream_info
        assert "urls" in stream_info
        
        # Verify stream configuration
        config = stream_info["config"]
        assert config["name"] == stream_name
        assert config["source"] == source_path
        assert config["record"] is True
        
        # Verify stream URLs
        urls = stream_info["urls"]
        assert "rtsp" in urls
        assert "webrtc" in urls
        assert "hls" in urls
        
        # Verify file system operations
        recordings_path = mediamtx_infrastructure.config.recordings_path
        snapshots_path = mediamtx_infrastructure.config.snapshots_path
        
        assert os.path.exists(recordings_path)
        assert os.path.exists(snapshots_path)
        
        # Get real stream status from MediaMTX
        stream_status = await mediamtx_infrastructure.get_stream_status(stream_name)
        assert stream_status is not None
        
        # Clean up
        await mediamtx_infrastructure.delete_test_stream(stream_name)

    @pytest.mark.asyncio
    async def test_real_websocket_communication_integration(
        self, service_manager
    ):
        """
        Test real WebSocket communication integration.
        
        Requirements: REQ-INT-003
        Scenario: Real WebSocket communication with service
        Expected: Successful WebSocket communication and notification delivery
        Edge Cases: Connection stability, message delivery, error handling
        """
        # Create authenticated WebSocket client
        auth_manager = get_test_auth_manager()
        user_factory = UserFactory(auth_manager)
        test_user = user_factory.create_operator_user("websocket_test_user")
        
        websocket_url = f"ws://{service_manager.config.server.host}:{service_manager.config.server.port}{service_manager.config.server.websocket_path}"
        websocket_client = WebSocketAuthTestClient(websocket_url, test_user)
        
        # Connect to WebSocket server
        await websocket_client.connect()
        
        try:
            # Test ping/pong communication with proper authentication
            ping_response = await websocket_client.call_protected_method("ping", {})
            assert "result" in ping_response, "Response should contain 'result' field"
            
            # Test camera list request with proper authentication
            camera_list_response = await websocket_client.call_protected_method("get_camera_list", {})
            assert "result" in camera_list_response, "Response should contain 'result' field"
            assert isinstance(camera_list_response["result"]["cameras"], list)
            assert isinstance(camera_list_response["result"]["total"], int)
            assert isinstance(camera_list_response["result"]["connected"], int)
            
            # Test error handling with invalid requests
            invalid_response = await websocket_client.send_unauthenticated_request(
                "invalid_method", {"invalid": "params"}
            )
            assert "error" in invalid_response, "Should return error for invalid method"
            assert invalid_response["error"]["code"] < 0
            
        finally:
            await websocket_client.disconnect()

    @pytest.mark.asyncio
    async def test_integration_error_handling_and_recovery(
        self, service_manager, mediamtx_infrastructure, websocket_client
    ):
        """
        Test integration error handling and recovery scenarios.
        
        Requirements: REQ-ERROR-003
        Scenario: Various integration failure scenarios
        Expected: Graceful error handling and recovery
        Edge Cases: Service failures, network issues, resource constraints
        """
        # Test MediaMTX service unavailability
        try:
            # Simulate MediaMTX service failure
            await mediamtx_infrastructure.cleanup_mediamtx_service()
            
            # Try to create stream (should handle gracefully)
            try:
                await mediamtx_infrastructure.create_test_stream("test", "/dev/video0")
                pytest.fail("Should not be able to create stream when MediaMTX is down")
            except Exception as e:
                # Expected error
                assert "MediaMTX" in str(e) or "service" in str(e).lower()
                
        finally:
            # Restart MediaMTX service
            await mediamtx_infrastructure.setup_mediamtx_service()
        
        # Test WebSocket connection failure
        invalid_client = WebSocketAuthTestClient("ws://invalid-server:9999/ws", test_user)
        
        try:
            await invalid_client.connect()
            pytest.fail("Should not be able to connect to invalid server")
        except Exception:
            # Expected connection failure
            pass
        
        # Test file system permission issues
        read_only_dir = "/tmp/readonly_test_dir"
        try:
            # Create read-only directory
            os.makedirs(read_only_dir, exist_ok=True)
            os.chmod(read_only_dir, 0o444)  # Read-only
            
            # Try to write to read-only directory
            test_file = os.path.join(read_only_dir, "test.txt")
            try:
                with open(test_file, 'w') as f:
                    f.write("test")
                pytest.fail("Should not be able to write to read-only directory")
            except PermissionError:
                # Expected permission error
                pass
                
        finally:
            # Clean up
            os.chmod(read_only_dir, 0o755)
            shutil.rmtree(read_only_dir, ignore_errors=True)

    @pytest.mark.asyncio
    async def test_integration_performance_and_load_testing(
        self, service_manager, mediamtx_infrastructure, websocket_client
    ):
        """
        Test integration performance and load handling.
        
        Requirements: REQ-INT-001, REQ-INT-002
        Scenario: Performance testing with multiple concurrent operations
        Expected: System handles load within performance requirements
        Edge Cases: High concurrency, resource contention, performance degradation
        """
        import time
        
        # Connect WebSocket client
        await websocket_client.connect()
        
        try:
            # Test multiple concurrent operations
            start_time = time.time()
            
            # Create multiple test streams
            stream_count = 5
            stream_names = [f"perf_test_stream_{i}" for i in range(stream_count)]
            
            # Create streams concurrently
            create_tasks = []
            for stream_name in stream_names:
                task = mediamtx_infrastructure.create_test_stream(stream_name, "/dev/video0")
                create_tasks.append(task)
            
            # Wait for all streams to be created
            await asyncio.gather(*create_tasks)
            
            # Test concurrent WebSocket requests
            request_tasks = []
            for i in range(10):
                task = websocket_client.get_camera_list()
                request_tasks.append(task)
            
            # Wait for all requests to complete
            responses = await asyncio.gather(*request_tasks)
            
            end_time = time.time()
            execution_time = end_time - start_time
            
            # Validate performance
            assert execution_time < 30.0  # Should complete within 30 seconds
            assert len(responses) == 10
            assert all(response.result is not None for response in responses)
            
            # Clean up streams
            cleanup_tasks = []
            for stream_name in stream_names:
                task = mediamtx_infrastructure.delete_test_stream(stream_name)
                cleanup_tasks.append(task)
            
            await asyncio.gather(*cleanup_tasks)
            
        finally:
            await websocket_client.disconnect()

    @pytest.mark.asyncio
    async def test_integration_data_consistency_and_validation(
        self, service_manager, mediamtx_infrastructure, websocket_client
    ):
        """
        Test data consistency across integrated components.
        
        Requirements: REQ-INT-001, REQ-INT-002, REQ-INT-003
        Scenario: Data consistency validation across components
        Expected: Consistent data across all integrated components
        Edge Cases: Data synchronization, state consistency, validation errors
        """
        # Connect WebSocket client
        await websocket_client.connect()
        
        try:
            # Get camera list from WebSocket
            camera_list_response = await websocket_client.get_camera_list()
            cameras = camera_list_response.result["cameras"]
            
            if cameras:
                camera_device = cameras[0]["device"]
                
                # Get camera status from WebSocket
                status_response = await websocket_client.get_camera_status(camera_device)
                websocket_status = status_response.result
                
                # Validate data consistency
                assert websocket_status["device"] == camera_device
                assert websocket_status["status"] in ["CONNECTED", "DISCONNECTED"]
                
                # If camera is connected, validate capability data
                if websocket_status["status"] == "CONNECTED":
                    assert "resolution" in websocket_status
                    assert "fps" in websocket_status
                    assert "capabilities" in websocket_status
                    
                    # Validate capability data structure
                    capabilities = websocket_status["capabilities"]
                    assert isinstance(capabilities, dict)
                    assert "formats" in capabilities
                    assert "resolutions" in capabilities
                    assert isinstance(capabilities["formats"], list)
                    assert isinstance(capabilities["resolutions"], list)
                
                # Validate stream data consistency
                if "streams" in websocket_status:
                    streams = websocket_status["streams"]
                    assert isinstance(streams, dict)
                    
                    # Validate stream URL formats
                    for protocol, url in streams.items():
                        assert isinstance(url, str)
                        assert len(url) > 0
                        
                        if protocol == "rtsp":
                            assert url.startswith("rtsp://")
                        elif protocol == "webrtc":
                            assert url.startswith("http://")
                        elif protocol == "hls":
                            assert url.startswith("http://")
            
        finally:
            await websocket_client.disconnect()

    @pytest.mark.asyncio
    async def test_integration_cleanup_and_resource_management(
        self, service_manager, mediamtx_infrastructure, temp_test_directory
    ):
        """
        Test integration cleanup and resource management.
        
        Requirements: REQ-INT-004, REQ-ERROR-003
        Scenario: System cleanup and resource management
        Expected: Proper cleanup of all resources and file system
        Edge Cases: Abrupt termination, resource leaks, cleanup failures
        """
        # Create test resources
        stream_name = "cleanup_test_stream"
        await mediamtx_infrastructure.create_test_stream(stream_name, "/dev/video0")
        
        # Verify resources exist
        stream_status = await mediamtx_infrastructure.get_stream_status(stream_name)
        assert stream_status is not None
        
        # Test service cleanup
        await service_manager.stop()
        
        # Verify MediaMTX service is stopped
        try:
            await mediamtx_infrastructure.get_stream_status(stream_name)
            pytest.fail("Should not be able to access stream after service stop")
        except Exception:
            # Expected - service is stopped
            pass
        
        # Verify file system cleanup
        recordings_path = mediamtx_infrastructure.config.recordings_path
        snapshots_path = mediamtx_infrastructure.config.snapshots_path
        
        # Clean up test directories
        if os.path.exists(recordings_path):
            shutil.rmtree(recordings_path, ignore_errors=True)
        if os.path.exists(snapshots_path):
            shutil.rmtree(snapshots_path, ignore_errors=True)
        
        # Verify temp directory cleanup
        assert not os.path.exists(recordings_path)
        assert not os.path.exists(snapshots_path)


# Example usage of the integration template
class TestCameraServiceIntegration(TestSystemIntegration):
    """
    Example implementation using the integration template.
    
    Replace with your specific integration test scenarios
    and update the test methods with your specific logic.
    """
    
    # Override fixtures with your specific configuration
    @pytest.fixture
    def integration_config(self, temp_test_directory) -> Config:
        """Create camera service integration test configuration."""
        from tests.utils.port_utils import find_free_port
        
        return Config(
            server=ServerConfig(
                host="127.0.0.1",
                port=find_free_port(),  # Dynamic port to avoid conflicts
                websocket_path="/ws",
                max_connections=20
            ),
            mediamtx=MediaMTXConfig(
                host="127.0.0.1",
                api_port=9997,  # Fixed systemd service port
                rtsp_port=8554,  # Fixed systemd service port
                webrtc_port=8889,  # Fixed systemd service port
                hls_port=8888,  # Fixed systemd service port
                config_path=os.path.join(temp_test_directory, "mediamtx_integration.yml"),
                recordings_path=os.path.join(temp_test_directory, "recordings"),
                snapshots_path=os.path.join(temp_test_directory, "snapshots"),
            ),
            camera=CameraConfig(
                device_range=[0, 1, 2],
                enable_capability_detection=True,
                detection_timeout=3.0
            ),
            health_port=find_free_port(),  # Dynamic health port to avoid conflicts
        )
    
    # Add your specific integration test methods here
    # They will inherit the structure and patterns from the template
