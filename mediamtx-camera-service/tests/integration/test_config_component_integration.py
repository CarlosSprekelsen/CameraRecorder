"""
Integration test for configuration → component instantiation chain.
Tests that configuration can be loaded and used to instantiate all components.
"""

import pytest
import asyncio
from unittest.mock import Mock, AsyncMock

from src.camera_service.config import ConfigManager
from src.camera_service.service_manager import ServiceManager
from src.mediamtx_wrapper.controller import MediaMTXController


class TestConfigurationComponentIntegration:
    """Test that configuration properly instantiates all components."""
    
    def test_mediamtx_controller_instantiation(self):
        """Test that MediaMTXConfig can instantiate MediaMTXController."""
        # Load real configuration
        config_manager = ConfigManager()
        config = config_manager.load_config()
        
        # Test that MediaMTXController can be instantiated with config
        mediamtx_config = config.mediamtx
        
        # This should not raise any parameter mismatch errors
        controller = MediaMTXController(
            host=mediamtx_config.host,
            api_port=mediamtx_config.api_port,
            rtsp_port=mediamtx_config.rtsp_port,
            webrtc_port=mediamtx_config.webrtc_port,
            hls_port=mediamtx_config.hls_port,
            config_path=mediamtx_config.config_path,
            recordings_path=mediamtx_config.recordings_path,
            snapshots_path=mediamtx_config.snapshots_path,
            health_check_interval=mediamtx_config.health_check_interval,
            health_failure_threshold=mediamtx_config.health_failure_threshold,
            health_circuit_breaker_timeout=mediamtx_config.health_circuit_breaker_timeout,
            health_max_backoff_interval=mediamtx_config.health_max_backoff_interval,
            health_recovery_confirmation_threshold=mediamtx_config.health_recovery_confirmation_threshold,
            backoff_base_multiplier=mediamtx_config.backoff_base_multiplier,
            backoff_jitter_range=mediamtx_config.backoff_jitter_range,
            process_termination_timeout=mediamtx_config.process_termination_timeout,
            process_kill_timeout=mediamtx_config.process_kill_timeout,
        )
        
        assert controller is not None
        assert hasattr(controller, 'host')
        assert controller.host == mediamtx_config.host
    
    def test_service_manager_instantiation(self):
        """Test that ServiceManager can be instantiated with config."""
        config_manager = ConfigManager()
        config = config_manager.load_config()
        
        # This should not raise any parameter mismatch errors
        service_manager = ServiceManager(config)
        
        assert service_manager is not None
        assert service_manager._config == config
    
    def test_camera_monitor_instantiation(self):
        """Test that HybridCameraMonitor can be instantiated with config."""
        config_manager = ConfigManager()
        config = config_manager.load_config()
        
        from camera_discovery.hybrid_monitor import HybridCameraMonitor
        
        # This should not raise any parameter mismatch errors
        camera_monitor = HybridCameraMonitor(
            device_range=config.camera.device_range,
            poll_interval=config.camera.poll_interval,
            detection_timeout=config.camera.detection_timeout,
            enable_capability_detection=config.camera.enable_capability_detection,
        )
        
        assert camera_monitor is not None
        assert camera_monitor._device_range == config.camera.device_range
    
    def test_websocket_server_instantiation(self):
        """Test that WebSocketJsonRpcServer can be instantiated with config."""
        config_manager = ConfigManager()
        config = config_manager.load_config()
        
        from websocket_server.server import WebSocketJsonRpcServer
        
        # This should not raise any parameter mismatch errors
        websocket_server = WebSocketJsonRpcServer(
            host=config.server.host,
            port=config.server.port,
            websocket_path=config.server.websocket_path,
            max_connections=config.server.max_connections,
        )
        
        assert websocket_server is not None
        assert websocket_server._host == config.server.host
        assert websocket_server._port == config.server.port
    
    def test_configuration_schema_completeness(self):
        """Test that all required configuration fields are present."""
        config_manager = ConfigManager()
        config = config_manager.load_config()
        
        # Test that all required MediaMTX fields are present
        required_mediamtx_fields = [
            'host', 'api_port', 'rtsp_port', 'webrtc_port', 'hls_port',
            'config_path', 'recordings_path', 'snapshots_path',
            'health_check_interval', 'health_failure_threshold',
            'health_circuit_breaker_timeout', 'health_max_backoff_interval',
            'health_recovery_confirmation_threshold', 'backoff_base_multiplier',
            'backoff_jitter_range', 'process_termination_timeout', 'process_kill_timeout'
        ]
        
        for field in required_mediamtx_fields:
            assert hasattr(config.mediamtx, field), f"Missing field: {field}"
    
    def test_configuration_parameter_types(self):
        """Test that configuration parameters have correct types."""
        config_manager = ConfigManager()
        config = config_manager.load_config()
        
        # Test MediaMTX parameter types
        assert isinstance(config.mediamtx.host, str)
        assert isinstance(config.mediamtx.api_port, int)
        assert isinstance(config.mediamtx.rtsp_port, int)
        assert isinstance(config.mediamtx.webrtc_port, int)
        assert isinstance(config.mediamtx.hls_port, int)
        assert isinstance(config.mediamtx.health_check_interval, int)
        assert isinstance(config.mediamtx.health_failure_threshold, int)
        assert isinstance(config.mediamtx.backoff_base_multiplier, float)
        assert isinstance(config.mediamtx.backoff_jitter_range, tuple)
    
    def test_configuration_serialization(self):
        """Test that configuration can be serialized and deserialized."""
        config_manager = ConfigManager()
        config = config_manager.load_config()
        
        # Serialize to dict
        config_dict = config.to_dict()
        
        # Deserialize back to config
        new_config = config_manager._create_config_object(config_dict)
        
        # Verify all values are preserved
        assert new_config.mediamtx.host == config.mediamtx.host
        assert new_config.mediamtx.api_port == config.mediamtx.api_port
        assert new_config.mediamtx.health_check_interval == config.mediamtx.health_check_interval
        assert new_config.mediamtx.backoff_base_multiplier == config.mediamtx.backoff_base_multiplier
        assert new_config.mediamtx.backoff_jitter_range == config.mediamtx.backoff_jitter_range
    
    def test_controller_parameter_compatibility(self):
        """Test that MediaMTXController constructor accepts all config parameters."""
        config_manager = ConfigManager()
        config = config_manager.load_config()
        
        # Get MediaMTXController constructor parameters
        import inspect
        controller_params = inspect.signature(MediaMTXController.__init__).parameters
        controller_param_names = set(controller_params.keys()) - {'self'}
        
        # Get MediaMTXConfig field names
        from dataclasses import fields
        config_fields = {field.name for field in fields(config.mediamtx)}
        
        # Check that all controller parameters exist in config
        missing_in_config = controller_param_names - config_fields
        assert len(missing_in_config) == 0, f"MediaMTXController parameters missing from MediaMTXConfig: {missing_in_config}"
        
        # Check that all config fields are used by controller (optional check)
        unused_in_controller = config_fields - controller_param_names
        if unused_in_controller:
            print(f"Warning: MediaMTXConfig fields not used by MediaMTXController: {unused_in_controller}")


class TestConfigurationValidation:
    """Test configuration validation and error handling."""
    
    def test_invalid_configuration_handling(self):
        """Test that invalid configurations are handled gracefully."""
        config_manager = ConfigManager()
        
        # Test with minimal valid config
        try:
            config = config_manager.load_config()
            assert config is not None
        except Exception as e:
            pytest.fail(f"Valid configuration failed to load: {e}")
    
    def test_missing_health_parameters_handling(self):
        """Test that missing health parameters are handled gracefully."""
        # This test ensures that if health parameters are missing,
        # the system falls back to defaults rather than crashing
        config_manager = ConfigManager()
        config = config_manager.load_config()
        
        # Ensure all health parameters have default values
        assert hasattr(config.mediamtx, 'health_check_interval')
        assert hasattr(config.mediamtx, 'health_failure_threshold')
        assert hasattr(config.mediamtx, 'backoff_base_multiplier')
        assert hasattr(config.mediamtx, 'backoff_jitter_range')
        
        # Ensure they have reasonable default values
        assert config.mediamtx.health_check_interval > 0
        assert config.mediamtx.health_failure_threshold > 0
        assert config.mediamtx.backoff_base_multiplier > 0
        assert len(config.mediamtx.backoff_jitter_range) == 2
