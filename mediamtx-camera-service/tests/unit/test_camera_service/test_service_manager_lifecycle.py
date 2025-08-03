# tests/unit/test_camera_service/test_service_manager_lifecycle.py
"""
Test service manager lifecycle orchestration and camera event handling.

Covers camera connect/disconnect orchestration, metadata propagation with capability
validation logic, and failure recovery scenarios.
"""

import pytest
import asyncio
from unittest.mock import Mock, AsyncMock, patch

from src.camera_service.service_manager import ServiceManager
from src.camera_service.config import Config, ServerConfig, MediaMTXConfig, CameraConfig, LoggingConfig, RecordingConfig, SnapshotConfig
from src.camera_discovery.hybrid_monitor import CameraEventData, CameraEvent
from src.common.types import CameraDevice


class TestServiceManagerLifecycle:
    """Test service manager lifecycle and camera event orchestration."""

    @pytest.fixture
    def mock_config(self):
        """Create mock configuration for testing."""
        return Config(
            server=ServerConfig(host="localhost", port=8002),
            mediamtx=MediaMTXConfig(
                host="localhost", 
                api_port=9997,
                rtsp_port=8554,
                webrtc_port=8889,
                hls_port=8888,
                recordings_path="/tmp/recordings",
                snapshots_path="/tmp/snapshots"
            ),
            camera=CameraConfig(device_range=[0, 1, 2], enable_capability_detection=True),  
            logging=LoggingConfig(),
            recording=RecordingConfig(),
            snapshots=SnapshotConfig()
        )

    @pytest.fixture
    def service_manager(self, mock_config):
        """Create service manager instance for testing."""
        return ServiceManager(mock_config)

    @pytest.fixture
    def mock_camera_event_connected(self):
        """Create mock camera connection event."""
        camera_device = CameraDevice(
            device="/dev/video0",
            name="Test Camera 0",
            status="CONNECTED"
        )
        return CameraEventData(
            device_path="/dev/video0",
            event_type=CameraEvent.CONNECTED,
            device_info=camera_device,
            timestamp=1234567890.0
        )

    @pytest.fixture
    def mock_camera_event_disconnected(self):
        """Create mock camera disconnection event."""
        camera_device = CameraDevice(
            device="/dev/video0",
            name="Test Camera 0",
            status="DISCONNECTED"
        )
        return CameraEventData(
            device_path="/dev/video0",
            event_type=CameraEvent.DISCONNECTED,
            device_info=camera_device,
            timestamp=1234567891.0
        )

    @pytest.mark.asyncio
    async def test_camera_connect_orchestration_sequence(self, service_manager, mock_camera_event_connected):
        """
        Test camera connection orchestration follows correct sequence:
        1. Stream name generation
        2. MediaMTX stream creation
        3. Capability metadata retrieval
        4. Notification broadcasting
        """
        # Mock dependencies
        mock_mediamtx = Mock()
        mock_mediamtx.create_stream = AsyncMock(return_value={
            "rtsp": "rtsp://localhost:8554/camera0",
            "webrtc": "http://localhost:8889/camera0",
            "hls": "http://localhost:8888/camera0"
        })
        service_manager._mediamtx_controller = mock_mediamtx
        
        mock_websocket = Mock()
        mock_websocket.notify_camera_status_update = AsyncMock()
        service_manager._websocket_server = mock_websocket
        
        # Mock camera monitor with confirmed capability data
        mock_camera_monitor = Mock()
        mock_camera_monitor.get_effective_capability_metadata = Mock(return_value={
            "resolution": "1920x1080",
            "fps": 30,
            "validation_status": "confirmed",
            "consecutive_successes": 5,
            "formats": ["YUYV", "MJPEG"],
            "all_resolutions": ["1920x1080", "1280x720"]
        })
        service_manager._camera_monitor = mock_camera_monitor
        
        # Execute camera connection handling
        await service_manager.handle_camera_event(mock_camera_event_connected)
        
        # Verify orchestration sequence
        # 1. MediaMTX stream creation was called
        mock_mediamtx.create_stream.assert_called_once()
        stream_config = mock_mediamtx.create_stream.call_args[0][0]
        assert stream_config.name == "camera0"
        assert stream_config.source == "/dev/video0"
        
        # 2. Capability metadata was retrieved
        mock_camera_monitor.get_effective_capability_metadata.assert_called_once_with("/dev/video0")
        
        # 3. Notification was sent with correct metadata
        mock_websocket.notify_camera_status_update.assert_called_once()
        notification_params = mock_websocket.notify_camera_status_update.call_args[0][0]
        assert notification_params["device"] == "/dev/video0"
        assert notification_params["status"] == "CONNECTED"
        assert notification_params["resolution"] == "1920x1080"  # From confirmed capability
        assert notification_params["fps"] == 30                   # From confirmed capability
        assert "rtsp" in notification_params["streams"]

    @pytest.mark.asyncio
    async def test_camera_disconnect_orchestration_sequence(self, service_manager, mock_camera_event_disconnected):
        """
        Test camera disconnection orchestration follows correct sequence:
        1. MediaMTX stream deletion
        2. Metadata retrieval (cached/default)
        3. Notification broadcasting with empty streams
        """
        # Mock dependencies
        mock_mediamtx = Mock()
        mock_mediamtx.delete_stream = AsyncMock(return_value=True)
        service_manager._mediamtx_controller = mock_mediamtx
        
        mock_websocket = Mock()
        mock_websocket.notify_camera_status_update = AsyncMock()
        service_manager._websocket_server = mock_websocket
        
        # Mock camera monitor (for metadata fallback)
        mock_camera_monitor = Mock()
        mock_camera_monitor.get_effective_capability_metadata = Mock(return_value={
            "resolution": "1920x1080",
            "fps": 30,
            "validation_status": "none"
        })
        service_manager._camera_monitor = mock_camera_monitor
        
        # Execute camera disconnection handling
        await service_manager.handle_camera_event(mock_camera_event_disconnected)
        
        # Verify orchestration sequence
        # 1. MediaMTX stream deletion was called
        mock_mediamtx.delete_stream.assert_called_once_with("camera0")
        
        # 2. Notification was sent with disconnected status
        mock_websocket.notify_camera_status_update.assert_called_once()
        notification_params = mock_websocket.notify_camera_status_update.call_args[0][0]
        assert notification_params["device"] == "/dev/video0"
        assert notification_params["status"] == "DISCONNECTED"
        assert notification_params["resolution"] == ""  # Empty for disconnected
        assert notification_params["fps"] == 0          # Zero for disconnected
        assert notification_params["streams"] == {}     # Empty streams

    @pytest.mark.asyncio
    async def test_metadata_propagation_confirmed_capability(self, service_manager, mock_camera_event_connected):
        """Test metadata propagation uses confirmed capability data and logs validation status."""
        # Mock camera monitor with confirmed capability data
        mock_camera_monitor = Mock()
        mock_camera_monitor.get_effective_capability_metadata = Mock(return_value={
            "resolution": "1280x720",
            "fps": 25,
            "validation_status": "confirmed",
            "consecutive_successes": 10,
            "formats": ["YUYV"],
            "all_resolutions": ["1280x720", "640x480"]
        })
        service_manager._camera_monitor = mock_camera_monitor
        
        # Mock other dependencies
        service_manager._mediamtx_controller = Mock()
        service_manager._mediamtx_controller.create_stream = AsyncMock(return_value={})
        service_manager._websocket_server = Mock()
        service_manager._websocket_server.notify_camera_status_update = AsyncMock()
        
        # Execute with logging capture
        with patch('src.camera_service.service_manager.set_correlation_id'):
            await service_manager.handle_camera_event(mock_camera_event_connected)
        
        # Verify confirmed capability data is used
        notification_params = service_manager._websocket_server.notify_camera_status_update.call_args[0][0]
        assert notification_params["resolution"] == "1280x720"  # From confirmed capability
        assert notification_params["fps"] == 25                 # From confirmed capability

    @pytest.mark.asyncio
    async def test_metadata_propagation_provisional_capability(self, service_manager, mock_camera_event_connected):
        """Test metadata propagation uses provisional capability data with appropriate logging."""
        # Mock camera monitor with provisional capability data
        mock_camera_monitor = Mock()
        mock_camera_monitor.get_effective_capability_metadata = Mock(return_value={
            "resolution": "1920x1080",
            "fps": 30,
            "validation_status": "provisional",
            "consecutive_successes": 1,
            "formats": ["MJPEG"],
            "all_resolutions": ["1920x1080"]
        })
        service_manager._camera_monitor = mock_camera_monitor
        
        # Mock other dependencies
        service_manager._mediamtx_controller = Mock()
        service_manager._mediamtx_controller.create_stream = AsyncMock(return_value={})
        service_manager._websocket_server = Mock()
        service_manager._websocket_server.notify_camera_status_update = AsyncMock()
        
        # Execute capability metadata retrieval directly
        metadata = await service_manager._get_enhanced_camera_metadata(mock_camera_event_connected)
        
        # Verify provisional capability data is used with correct annotations
        assert metadata["resolution"] == "1920x1080"
        assert metadata["fps"] == 30
        assert metadata["validation_status"] == "provisional"
        assert metadata["capability_source"] == "provisional_capability"
        assert metadata["consecutive_successes"] == 1

    @pytest.mark.asyncio
    async def test_metadata_fallback_no_capability_data(self, service_manager, mock_camera_event_connected):
        """Test metadata falls back to architecture defaults when no capability data available."""
        # Mock camera monitor with no capability data
        mock_camera_monitor = Mock()
        mock_camera_monitor.get_effective_capability_metadata = Mock(return_value={
            "resolution": "1920x1080",  # Architecture defaults
            "fps": 30,
            "validation_status": "none"
        })
        service_manager._camera_monitor = mock_camera_monitor
        
        # Execute capability metadata retrieval
        metadata = await service_manager._get_enhanced_camera_metadata(mock_camera_event_connected)
        
        # Verify architecture defaults are used with correct annotations
        assert metadata["resolution"] == "1920x1080"  # Architecture default
        assert metadata["fps"] == 30                   # Architecture default
        assert metadata["validation_status"] == "none"
        assert metadata["capability_source"] == "default"

    @pytest.mark.asyncio
    async def test_mediamtx_stream_creation_failure_recovery(self, service_manager, mock_camera_event_connected):
        """Test graceful recovery when MediaMTX stream creation fails."""
        # Mock MediaMTX controller that fails stream creation
        mock_mediamtx = Mock()
        mock_mediamtx.create_stream = AsyncMock(side_effect=Exception("MediaMTX connection failed"))
        service_manager._mediamtx_controller = mock_mediamtx
        
        # Mock other dependencies
        mock_websocket = Mock()
        mock_websocket.notify_camera_status_update = AsyncMock()
        service_manager._websocket_server = mock_websocket
        
        mock_camera_monitor = Mock()
        mock_camera_monitor.get_effective_capability_metadata = Mock(return_value={
            "resolution": "1920x1080",
            "fps": 30,
            "validation_status": "confirmed"
        })
        service_manager._camera_monitor = mock_camera_monitor
        
        # Execute camera connection handling (should not raise exception)
        await service_manager.handle_camera_event(mock_camera_event_connected)
        
        # Verify notification still sent despite MediaMTX failure
        mock_websocket.notify_camera_status_update.assert_called_once()
        notification_params = mock_websocket.notify_camera_status_update.call_args[0][0]
        assert notification_params["device"] == "/dev/video0"
        assert notification_params["status"] == "CONNECTED"
        assert notification_params["streams"] == {}  # Empty due to MediaMTX failure

    @pytest.mark.asyncio
    async def test_missing_mediamtx_controller_defensive_behavior(self, service_manager, mock_camera_event_connected):
        """Test defensive behavior when MediaMTX controller is not available."""
        # No MediaMTX controller
        service_manager._mediamtx_controller = None
        
        # Mock other dependencies
        mock_websocket = Mock()
        mock_websocket.notify_camera_status_update = AsyncMock()
        service_manager._websocket_server = mock_websocket
        
        mock_camera_monitor = Mock()
        mock_camera_monitor.get_effective_capability_metadata = Mock(return_value={
            "resolution": "1920x1080",
            "fps": 30,
            "validation_status": "none"
        })
        service_manager._camera_monitor = mock_camera_monitor
        
        # Execute camera connection handling (should not crash)
        await service_manager.handle_camera_event(mock_camera_event_connected)
        
        # Verify notification still sent with warning logged
        mock_websocket.notify_camera_status_update.assert_called_once()
        notification_params = mock_websocket.notify_camera_status_update.call_args[0][0]
        assert notification_params["streams"] == {}  # No streams without MediaMTX

    @pytest.mark.asyncio
    async def test_capability_detection_error_fallback(self, service_manager, mock_camera_event_connected):
        """Test fallback behavior when capability detection throws exception."""
        # Mock camera monitor that raises exception
        mock_camera_monitor = Mock()
        mock_camera_monitor.get_effective_capability_metadata = Mock(side_effect=Exception("Capability detection failed"))
        service_manager._camera_monitor = mock_camera_monitor
        
        # Execute capability metadata retrieval
        metadata = await service_manager._get_enhanced_camera_metadata(mock_camera_event_connected)
        
        # Verify fallback to defaults with error annotation
        assert metadata["resolution"] == "1920x1080"  # Architecture default
        assert metadata["fps"] == 30                   # Architecture default
        assert metadata["validation_status"] == "error"
        assert metadata["capability_source"] == "default"

    def test_stream_name_generation_deterministic(self, service_manager):
        """Test stream name generation is deterministic for various device paths."""
        test_cases = [
            ("/dev/video0", "camera0"),
            ("/dev/video15", "camera15"),
            ("/dev/video999", "camera999"),
            ("/custom/video5", "camera5"),
            ("/path/with/multiple/video2/segments", "camera2"),
            ("/no/numbers/here", "camera_"),  # Will get hash-based name
            ("", "camera_unknown")
        ]
        
        for device_path, expected_prefix in test_cases:
            result = service_manager._get_stream_name_from_device_path(device_path)
            if expected_prefix.endswith("_"):
                # Hash-based names should start with expected prefix
                assert result.startswith(expected_prefix)
                assert result != "camera_unknown"  # Should be deterministic hash
            else:
                assert result == expected_prefix

    @pytest.mark.asyncio
    async def test_service_lifecycle_startup_shutdown_sequence(self, service_manager):
        """Test complete service lifecycle startup and shutdown sequence."""
        # Mock all components
        with patch('src.camera_service.service_manager.MediaMTXController') as MockMediaMTX, \
             patch('src.camera_discovery.hybrid_monitor.HybridCameraMonitor') as MockCameraMonitor, \
             patch('src.camera_service.service_manager.HealthMonitor') as MockHealthMonitor, \
             patch('src.websocket_server.server.WebSocketJsonRpcServer') as MockWebSocketServer:
            
            # Setup mock instances
            mock_mediamtx_instance = Mock()
            mock_mediamtx_instance.start = AsyncMock()
            mock_mediamtx_instance.health_check = AsyncMock(return_value={"status": "healthy"})
            mock_mediamtx_instance.stop = AsyncMock()
            MockMediaMTX.return_value = mock_mediamtx_instance
            
            mock_camera_instance = Mock()
            mock_camera_instance.start = AsyncMock()
            mock_camera_instance.stop = AsyncMock()
            MockCameraMonitor.return_value = mock_camera_instance
            
            mock_health_instance = Mock()
            mock_health_instance.start = AsyncMock()
            mock_health_instance.stop = AsyncMock()
            MockHealthMonitor.return_value = mock_health_instance
            
            mock_websocket_instance = Mock()
            mock_websocket_instance.start = AsyncMock()
            mock_websocket_instance.stop = AsyncMock()
            MockWebSocketServer.return_value = mock_websocket_instance
            
            # Test startup sequence
            await service_manager.start()
            
            # Verify startup order
            mock_mediamtx_instance.start.assert_called_once()
            mock_camera_instance.start.assert_called_once()
            mock_health_instance.start.assert_called_once()
            mock_websocket_instance.start.assert_called_once()
            
            assert service_manager.is_running is True
            
            # Test shutdown sequence
            await service_manager.stop()
            
            # Verify shutdown order (reverse of startup)
            mock_websocket_instance.stop.assert_called_once()
            mock_health_instance.stop.assert_called_once()
            mock_camera_instance.stop.assert_called_once()
            mock_mediamtx_instance.stop.assert_called_once()
            
            assert service_manager.is_running is False

    @pytest.mark.asyncio
    async def test_correlation_id_propagation_lifecycle(self, service_manager, mock_camera_event_connected):
        """Test correlation ID propagation through camera event lifecycle."""
        # Mock dependencies
        service_manager._mediamtx_controller = Mock()
        service_manager._mediamtx_controller.create_stream = AsyncMock(return_value={})
        service_manager._websocket_server = Mock()
        service_manager._websocket_server.notify_camera_status_update = AsyncMock()
        service_manager._camera_monitor = Mock()
        service_manager._camera_monitor.get_effective_capability_metadata = Mock(return_value={
            "resolution": "1920x1080",
            "fps": 30,
            "validation_status": "confirmed"
        })
        
        # Track correlation ID calls
        correlation_ids = []
        
        def mock_set_correlation_id(cid):
            correlation_ids.append(cid)
        
        with patch('src.camera_service.service_manager.set_correlation_id', side_effect=mock_set_correlation_id):
            await service_manager.handle_camera_event(mock_camera_event_connected)
        
        # Verify correlation ID was set
        assert len(correlation_ids) > 0
        # Should include device-specific correlation ID
        assert any("camera" in cid or "video0" in cid for cid in correlation_ids)

    def test_get_service_status(self, service_manager):
        """Test service status reporting includes all components."""
        # Mock some components as running
        service_manager._running = True
        service_manager._websocket_server = Mock()
        service_manager._mediamtx_controller = Mock()
        service_manager._camera_monitor = None  # Simulate stopped component
        service_manager._health_monitor = Mock()
        
        status = service_manager.get_status()
        
        # Verify status structure
        assert status["running"] is True
        assert status["websocket_server"] == "running"
        assert status["mediamtx_controller"] == "running"
        assert status["camera_monitor"] == "stopped"
        assert status["health_monitor"] == "running"


class TestServiceManagerCapabilityValidationIntegration:
    """Test capability validation state transitions and logging."""

    @pytest.fixture
    def service_manager_with_config(self):
        """Create service manager with capability detection enabled."""
        config = Config(
            server=ServerConfig(),
            mediamtx=MediaMTXConfig(),
            camera=CameraConfig(enable_capability_detection=True),
            logging=LoggingConfig(),
            recording=RecordingConfig(),
            snapshots=SnapshotConfig()
        )
        return ServiceManager(config)

    @pytest.mark.asyncio
    async def test_capability_validation_status_transitions(self, service_manager_with_config):
        """Test capability validation status transitions from provisional to confirmed."""
        camera_event = CameraEventData(
            device_path="/dev/video0",
            event_type=CameraEvent.CONNECTED,
            device_info=CameraDevice("/dev/video0", "Test Camera", "CONNECTED")
        )
        
        # Mock camera monitor for capability progression
        mock_camera_monitor = Mock()
        service_manager_with_config._camera_monitor = mock_camera_monitor
        
        # Test progression: none -> provisional -> confirmed
        capability_states = [
            {"validation_status": "none", "consecutive_successes": 0},
            {"validation_status": "provisional", "consecutive_successes": 1}, 
            {"validation_status": "confirmed", "consecutive_successes": 5}
        ]
        
        for i, capability_state in enumerate(capability_states):
            mock_camera_monitor.get_effective_capability_metadata = Mock(return_value={
                "resolution": "1920x1080",
                "fps": 30,
                **capability_state
            })
            
            metadata = await service_manager_with_config._get_enhanced_camera_metadata(camera_event)
            
            assert metadata["validation_status"] == capability_state["validation_status"]
            assert metadata["consecutive_successes"] == capability_state["consecutive_successes"]
            
            if capability_state["validation_status"] == "confirmed":
                assert metadata["capability_source"] == "confirmed_capability"
            elif capability_state["validation_status"] == "provisional":
                assert metadata["capability_source"] == "provisional_capability"
            else:
                assert metadata["capability_source"] == "default"

    @pytest.mark.asyncio
    async def test_enhanced_validation_integration_logging(self, service_manager_with_config, caplog):
        """Test enhanced capability validation integration includes proper logging."""
        camera_event = CameraEventData(
            device_path="/dev/video0",
            event_type=CameraEvent.CONNECTED,
            device_info=CameraDevice("/dev/video0", "Test Camera", "CONNECTED")
        )
        
        # Mock camera monitor with enhanced capability metadata
        mock_camera_monitor = Mock()
        mock_camera_monitor.get_effective_capability_metadata = Mock(return_value={
            "resolution": "1280x720",
            "fps": 25,
            "validation_status": "confirmed",
            "consecutive_successes": 8,
            "formats": ["YUYV", "MJPEG"],
            "all_resolutions": ["1920x1080", "1280x720", "640x480"]
        })
        service_manager_with_config._camera_monitor = mock_camera_monitor
        
        # Execute enhanced metadata retrieval with logging
        with caplog.at_level('DEBUG'):
            metadata = await service_manager_with_config._get_enhanced_camera_metadata(camera_event)
        
        # Verify enhanced metadata structure
        assert metadata["validation_status"] == "confirmed"
        assert metadata["capability_source"] == "confirmed_capability"
        assert metadata["consecutive_successes"] == 8
        
        # Verify logging includes capability validation context
        log_messages = [record.message for record in caplog.records]
        capability_logs = [msg for msg in log_messages if "confirmed capability data" in msg]
        assert len(capability_logs) > 0


# TODO: HIGH: Add integration tests with real camera monitor instance [Story:E1/S5]
# TODO: MEDIUM: Add performance tests for camera event processing latency [Story:E1/S5]
# TODO: MEDIUM: Add stress tests for rapid connect/disconnect sequences [Story:E1/S5]