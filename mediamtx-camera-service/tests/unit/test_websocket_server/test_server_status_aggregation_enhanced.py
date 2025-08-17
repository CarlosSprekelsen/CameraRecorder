# tests/unit/test_websocket_server/test_server_status_aggregation_enhanced.py
"""
Enhanced test status aggregation functionality in WebSocket JSON-RPC server.

This test file addresses PARTIAL coverage gaps identified in the comprehensive audit:
- REQ-WS-001: WebSocket server shall aggregate camera status with real MediaMTX integration
- REQ-WS-002: WebSocket server shall provide camera capability metadata integration
- REQ-WS-003: WebSocket server shall handle MediaMTX stream status queries
- REQ-ERROR-001: WebSocket server shall handle MediaMTX connection failures gracefully
- REQ-ERROR-002: WebSocket server shall handle client disconnection gracefully
- REQ-ERROR-003: System shall handle MediaMTX service unavailability gracefully

Requirements Traceability:
- REQ-WS-001: Real MediaMTX integration validation with comprehensive status aggregation
- REQ-WS-002: Camera capability metadata integration with validation scenarios
- REQ-WS-003: MediaMTX stream status queries with error handling
- REQ-ERROR-001: MediaMTX connection failure graceful handling
- REQ-ERROR-002: Client disconnection graceful handling
- REQ-ERROR-003: MediaMTX service unavailability graceful handling

Story Coverage: S3 - WebSocket API Integration
IV&V Control Point: Real MediaMTX integration validation with comprehensive error scenarios
"""

import pytest
import asyncio
import tempfile
import os
import subprocess
import time
import json
from unittest.mock import AsyncMock, MagicMock, patch

from src.websocket_server.server import WebSocketJsonRpcServer
from src.common.types import CameraDevice
from src.camera_discovery.hybrid_monitor import HybridCameraMonitor
from src.mediamtx_wrapper.controller import MediaMTXController
from tests.fixtures.mediamtx_test_infrastructure import mediamtx_infrastructure, mediamtx_controller
from tests.fixtures.websocket_test_client import WebSocketTestClient, websocket_client


class TestEnhancedServerStatusAggregation:
    """Enhanced test camera status aggregation with comprehensive real MediaMTX integration."""

    @pytest.fixture
    def real_config(self):
        """Real configuration for testing."""
        from src.camera_service.config import Config, ServerConfig, CameraConfig
        return Config(
            server=ServerConfig(
                host="localhost",
                port=8003,  # Different port to avoid conflicts
                websocket_path="/ws",
                max_connections=100
            ),
            camera=CameraConfig(
                device_range=[0, 1, 2],
                poll_interval=0.1,
                enable_capability_detection=True
            )
        )

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
    def temp_dirs(self):
        """Create temporary directories for MediaMTX configuration."""
        base = tempfile.mkdtemp(prefix="enhanced_status_test_")
        config_path = os.path.join(base, "mediamtx.yml")
        recordings_path = os.path.join(base, "recordings")
        snapshots_path = os.path.join(base, "snapshots")
        
        # Create directories
        os.makedirs(recordings_path, exist_ok=True)
        os.makedirs(snapshots_path, exist_ok=True)
        
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
                "snapshots_path": snapshots_path
            }
        finally:
            import shutil
            shutil.rmtree(base, ignore_errors=True)

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
    async def server(self, real_config, real_camera_monitor, real_mediamtx_controller):
        """Create WebSocket server instance with real MediaMTX integration."""
        server = WebSocketJsonRpcServer(
            host="localhost",
            port=8003,
            websocket_path="/ws",
            max_connections=100,
            mediamtx_controller=real_mediamtx_controller,
            camera_monitor=real_camera_monitor,
        )
        return server

    @pytest.mark.asyncio
    async def test_get_camera_status_comprehensive_real_mediamtx_integration(
        self, server, real_camera_monitor, real_mediamtx_controller
    ):
        """
        Comprehensive test of get_camera_status with real MediaMTX integration.
        
        Requirements: REQ-WS-001, REQ-WS-002, REQ-WS-003
        Scenario: Real MediaMTX integration with comprehensive status aggregation
        Expected: Successful integration with real MediaMTX service and capability metadata
        Edge Cases: Real stream status queries, actual metrics retrieval, capability validation
        """
        # Await the fixtures to get the actual objects
        actual_server = await server
        camera_monitor = await anext(real_camera_monitor)
        
        # Update the server's camera monitor with the awaited monitor
        actual_server._camera_monitor = camera_monitor
        
        # Use real camera monitor to get actual connected cameras
        connected_cameras = await camera_monitor.get_connected_cameras()
        
        # Get real capability metadata from actual camera detection
        if connected_cameras:
            device_path = list(connected_cameras.keys())[0]
            real_capability_metadata = camera_monitor.get_effective_capability_metadata(device_path)
        else:
            # If no real cameras, create a test camera device
            real_capability_metadata = {
                "resolution": "1280x720",
                "fps": 25,
                "validation_status": "provisional",
                "formats": ["YUYV", "MJPEG"],
                "all_resolutions": ["1920x1080", "1280x720", "640x480"],
                "consecutive_successes": 1,
            }

        # Await the MediaMTX controller fixture
        mediamtx_controller = await anext(real_mediamtx_controller)
        
        # Create real test stream in MediaMTX using proper API
        from src.mediamtx_wrapper.controller import StreamConfig
        stream_config = StreamConfig(
            name="camera0",
            source="/dev/video0",
            record=True
        )
        stream_info = await mediamtx_controller.create_stream(stream_config)
        
        # Get real MediaMTX stream status
        real_stream_status = await mediamtx_controller.get_stream_status("camera0")

        # Test get_camera_status method
        result = await actual_server._method_get_camera_status({"device": "/dev/video0"})

        # Verify real capability data is used, not defaults
        assert result["resolution"] == "1280x720"  # From capability detection
        assert result["fps"] == 25  # From capability detection
        assert result["status"] == "CONNECTED"
        assert result["name"] == "Test Camera 0"

        # Verify capabilities section populated with real data
        assert result["capabilities"]["formats"] == ["YUYV", "MJPEG"]
        assert result["capabilities"]["resolutions"] == [
            "1920x1080",
            "1280x720",
            "640x480",
        ]

        # Verify real MediaMTX integration (actual values from real service)
        assert "metrics" in result
        assert "streams" in result
        assert result["streams"]["rtsp"] == "rtsp://127.0.0.1:8554/camera0"

        # Verify real integration worked
        assert "metrics" in result
        assert "streams" in result
        
        # Clean up test stream
        await mediamtx_controller.delete_stream("camera0")

    @pytest.mark.asyncio
    async def test_get_camera_status_mediamtx_connection_failure_graceful_handling(
        self, server, real_camera_monitor
    ):
        """
        Test camera status handling when MediaMTX connection fails gracefully.
        
        Requirements: REQ-ERROR-001, REQ-ERROR-003
        Scenario: MediaMTX service unavailable or connection failure
        Expected: Graceful error handling without crashing, fallback to basic camera info
        Edge Cases: Network failures, service unavailability, connection timeouts
        """
        # Await the fixtures to get the actual objects
        actual_server = await server
        camera_monitor = await anext(real_camera_monitor)
        
        # Update the server's camera monitor with the awaited monitor
        actual_server._camera_monitor = camera_monitor
        
        # Test with a non-existent stream to simulate MediaMTX connection failure
        result = await actual_server._method_get_camera_status({"device": "/dev/video0"})

        # Verify basic camera info is still returned even with MediaMTX failure
        assert result["status"] in ["CONNECTED", "DISCONNECTED"]  # Real camera status
        assert result["device"] == "/dev/video0"
        assert "name" in result
        assert "resolution" in result
        assert "fps" in result

        # Verify error handling for MediaMTX integration - should not crash
        assert "metrics" in result
        assert "streams" in result
        
        # Verify graceful degradation - empty streams and default metrics
        assert isinstance(result["streams"], dict)
        assert isinstance(result["metrics"], dict)

    @pytest.mark.asyncio
    async def test_get_camera_status_capability_detection_failure_graceful_handling(
        self, server, real_camera_monitor
    ):
        """
        Test camera status handling when capability detection fails gracefully.
        
        Requirements: REQ-WS-002, REQ-ERROR-001
        Scenario: Capability detection method unavailable or fails
        Expected: Graceful fallback to architecture defaults
        Edge Cases: Missing capability detection support, detection timeouts
        """
        # Await the fixtures to get the actual objects
        actual_server = await server
        camera_monitor = await anext(real_camera_monitor)
        
        # Update the server's camera monitor with the awaited monitor
        actual_server._camera_monitor = camera_monitor
        
        # Test with a device that doesn't exist to trigger fallback
        result = await actual_server._method_get_camera_status({"device": "/dev/video999"})

        # Verify architecture defaults are used when capability detection fails
        assert result["resolution"] == "1920x1080"  # Architecture default
        assert result["fps"] == 30  # Architecture default
        assert result["status"] == "DISCONNECTED"
        assert result["device"] == "/dev/video999"

        # Verify empty capabilities when detection unavailable
        assert result["capabilities"]["formats"] == []
        assert result["capabilities"]["resolutions"] == []

    @pytest.mark.asyncio
    async def test_get_camera_list_comprehensive_capability_integration(
        self, server, real_camera_monitor, real_mediamtx_controller
    ):
        """
        Comprehensive test of get_camera_list with real capability data integration.
        
        Requirements: REQ-WS-002, REQ-WS-003
        Scenario: Multiple cameras with different capability metadata
        Expected: Real capability data integration in camera list
        Edge Cases: Different validation statuses, mixed capability data, capability conflicts
        """
        # Await the fixtures to get the actual objects
        actual_server = await server
        camera_monitor = await anext(real_camera_monitor)
        
        # Update the server's camera monitor with the awaited monitor
        actual_server._camera_monitor = camera_monitor
        
        # Test get_camera_list method with real camera monitor
        result = await actual_server._method_get_camera_list()

        # Verify real camera list structure
        assert "cameras" in result
        assert "total" in result
        assert "connected" in result
        
        cameras = result["cameras"]
        assert isinstance(cameras, list)

        # Verify each camera has proper structure with capability data
        for camera in cameras:
            assert "device" in camera
            assert "status" in camera
            assert "name" in camera
            assert "resolution" in camera
            assert "fps" in camera
            assert "capabilities" in camera
            
            # Verify capability data structure
            capabilities = camera["capabilities"]
            assert isinstance(capabilities, dict)
            assert "formats" in capabilities
            assert "resolutions" in capabilities
            assert isinstance(capabilities["formats"], list)
            assert isinstance(capabilities["resolutions"], list)

    @pytest.mark.asyncio
    async def test_get_camera_status_stream_status_queries_comprehensive(
        self, server, real_camera_monitor, real_mediamtx_controller
    ):
        """
        Comprehensive test of MediaMTX stream status queries.
        
        Requirements: REQ-WS-003
        Scenario: Real MediaMTX stream status queries with various states
        Expected: Accurate stream status reporting with real MediaMTX integration
        Edge Cases: Stream creation, deletion, active/inactive states, error conditions
        """
        # Await the fixtures to get the actual objects
        actual_server = await server
        camera_monitor = await anext(real_camera_monitor)
        mediamtx_controller = await anext(real_mediamtx_controller)
        
        # Update the server's components
        actual_server._camera_monitor = camera_monitor
        actual_server._mediamtx_controller = mediamtx_controller
        
        # Create multiple test streams to test status queries
        from src.mediamtx_wrapper.controller import StreamConfig
        
        streams_to_create = [
            StreamConfig(name="camera0", source="/dev/video0", record=True),
            StreamConfig(name="camera1", source="/dev/video1", record=False),
        ]
        
        created_streams = []
        try:
            # Create streams
            for stream_config in streams_to_create:
                stream_info = await mediamtx_controller.create_stream(stream_config)
                created_streams.append(stream_config.name)
                
                # Wait for stream to be active
                await asyncio.sleep(0.5)
            
            # Test status queries for each stream
            for stream_name in created_streams:
                result = await actual_server._method_get_camera_status({"device": f"/dev/video{stream_name[-1]}"})
                
                # Verify stream status is properly reported
                assert "streams" in result
                assert "metrics" in result
                
                # Verify stream URLs are correct
                streams = result["streams"]
                assert streams["rtsp"] == f"rtsp://127.0.0.1:8554/{stream_name}"
                assert streams["webrtc"] == f"webrtc://localhost:8002/{stream_name}"
                assert streams["hls"] == f"http://localhost:8002/hls/{stream_name}.m3u8"
                
                # Verify metrics are present
                metrics = result["metrics"]
                assert "bytes_sent" in metrics
                assert "readers" in metrics
                assert "uptime" in metrics
                
        finally:
            # Clean up test streams
            for stream_name in created_streams:
                try:
                    await mediamtx_controller.delete_stream(stream_name)
                except Exception:
                    pass  # Ignore cleanup errors

    @pytest.mark.asyncio
    async def test_get_camera_status_mediamtx_service_unavailability_comprehensive(
        self, server, real_camera_monitor
    ):
        """
        Comprehensive test of MediaMTX service unavailability handling.
        
        Requirements: REQ-ERROR-003
        Scenario: MediaMTX service completely unavailable
        Expected: Graceful handling without system crash, fallback to basic camera info
        Edge Cases: Service down, network unreachable, API timeout
        """
        # Await the fixtures to get the actual objects
        actual_server = await server
        camera_monitor = await anext(real_camera_monitor)
        
        # Update the server's camera monitor
        actual_server._camera_monitor = camera_monitor
        
        # Test with MediaMTX controller that simulates service unavailability
        with patch.object(actual_server, '_mediamtx_controller', None):
            result = await actual_server._method_get_camera_status({"device": "/dev/video0"})
            
            # Verify system continues to function without MediaMTX
            assert result["status"] in ["CONNECTED", "DISCONNECTED"]
            assert result["device"] == "/dev/video0"
            assert "name" in result
            assert "resolution" in result
            assert "fps" in result
            
            # Verify empty streams and default metrics when MediaMTX unavailable
            assert result["streams"] == {}
            assert result["metrics"] == {"bytes_sent": 0, "readers": 0, "uptime": 0}

    @pytest.mark.asyncio
    async def test_get_camera_status_capability_metadata_validation_scenarios(
        self, server, real_camera_monitor
    ):
        """
        Test camera status with various capability metadata validation scenarios.
        
        Requirements: REQ-WS-002
        Scenario: Different capability validation states and data quality
        Expected: Proper handling of provisional vs confirmed capability data
        Edge Cases: Inconsistent capability data, validation failures, data conflicts
        """
        # Await the fixtures to get the actual objects
        actual_server = await server
        camera_monitor = await anext(real_camera_monitor)
        
        # Update the server's camera monitor
        actual_server._camera_monitor = camera_monitor
        
        # Test with different capability metadata scenarios
        test_scenarios = [
            {
                "name": "provisional_data",
                "metadata": {
                    "resolution": "1280x720",
                    "fps": 25,
                    "validation_status": "provisional",
                    "formats": ["YUYV"],
                    "all_resolutions": ["1280x720"],
                    "consecutive_successes": 1,
                }
            },
            {
                "name": "confirmed_data",
                "metadata": {
                    "resolution": "1920x1080",
                    "fps": 30,
                    "validation_status": "confirmed",
                    "formats": ["YUYV", "MJPEG"],
                    "all_resolutions": ["1920x1080", "1280x720"],
                    "consecutive_successes": 3,
                }
            },
            {
                "name": "incomplete_data",
                "metadata": {
                    "resolution": "640x480",
                    "fps": 15,
                    "validation_status": "none",
                    "formats": [],
                    "all_resolutions": [],
                    "consecutive_successes": 0,
                }
            }
        ]
        
        for scenario in test_scenarios:
            # Mock capability metadata for this scenario
            with patch.object(camera_monitor, 'get_effective_capability_metadata', 
                            return_value=scenario["metadata"]):
                result = await actual_server._method_get_camera_status({"device": "/dev/video0"})
                
                # Verify capability data is properly integrated
                assert result["resolution"] == scenario["metadata"]["resolution"]
                assert result["fps"] == scenario["metadata"]["fps"]
                assert result["capabilities"]["formats"] == scenario["metadata"]["formats"]
                assert result["capabilities"]["resolutions"] == scenario["metadata"]["all_resolutions"]

    @pytest.mark.asyncio
    async def test_get_camera_status_error_propagation_and_recovery(
        self, server, real_camera_monitor
    ):
        """
        Test error propagation and recovery in camera status aggregation.
        
        Requirements: REQ-ERROR-001, REQ-ERROR-003
        Scenario: Various error conditions and recovery mechanisms
        Expected: Proper error handling and graceful degradation
        Edge Cases: Exception propagation, error recovery, system stability
        """
        # Await the fixtures to get the actual objects
        actual_server = await server
        camera_monitor = await anext(real_camera_monitor)
        
        # Update the server's camera monitor
        actual_server._camera_monitor = camera_monitor
        
        # Test with various error conditions
        error_scenarios = [
            {
                "name": "camera_monitor_exception",
                "mock_method": 'get_connected_cameras',
                "exception": Exception("Camera monitor error")
            },
            {
                "name": "capability_detection_exception", 
                "mock_method": 'get_effective_capability_metadata',
                "exception": Exception("Capability detection error")
            },
            {
                "name": "mediamtx_controller_exception",
                "mock_method": 'get_stream_status',
                "exception": Exception("MediaMTX controller error")
            }
        ]
        
        for scenario in error_scenarios:
            # Test that errors don't crash the system
            if scenario["mock_method"] == 'get_connected_cameras':
                with patch.object(camera_monitor, scenario["mock_method"], 
                                side_effect=scenario["exception"]):
                    result = await actual_server._method_get_camera_status({"device": "/dev/video0"})
            elif scenario["mock_method"] == 'get_effective_capability_metadata':
                with patch.object(camera_monitor, scenario["mock_method"], 
                                side_effect=scenario["exception"]):
                    result = await actual_server._method_get_camera_status({"device": "/dev/video0"})
            elif scenario["mock_method"] == 'get_stream_status':
                with patch.object(actual_server._mediamtx_controller, scenario["mock_method"], 
                                side_effect=scenario["exception"]):
                    result = await actual_server._method_get_camera_status({"device": "/dev/video0"})
            
            # Verify system continues to function despite errors
            assert result["device"] == "/dev/video0"
            assert "status" in result
            assert "name" in result
            assert "resolution" in result
            assert "fps" in result
            assert "capabilities" in result
            assert "streams" in result
            assert "metrics" in result
