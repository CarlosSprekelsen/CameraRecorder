"""
Unit tests for camera service configuration management.

Tests the ConfigManager class for loading, validation, environment overrides,
hot reload, and error handling as specified in the architecture.
"""

import os
import tempfile
import pytest
from pathlib import Path
from unittest.mock import Mock, patch, mock_open

from camera_service.config import (
    ConfigManager, load_config, get_config_manager,
    Config, ServerConfig, MediaMTXConfig, CameraConfig,
    LoggingConfig, RecordingConfig, SnapshotConfig
)


class TestConfigDataClasses:
    """Test configuration data class defaults and validation."""
    
    def test_server_config_defaults(self):
        """Test ServerConfig default values."""
        config = ServerConfig()
        assert config.host == "0.0.0.0"
        assert config.port == 8002
        assert config.websocket_path == "/ws"
        assert config.max_connections == 100
    
    def test_camera_config_device_range_default(self):
        """Test CameraConfig device_range default initialization."""
        config = CameraConfig()
        assert config.device_range == list(range(10))
        assert config.poll_interval == 0.1
        assert config.enable_capability_detection is True


class TestConfigManager:
    """Test ConfigManager configuration loading and management."""
    
    @pytest.fixture
    def config_manager(self):
        """Create a fresh ConfigManager instance for testing."""
        return ConfigManager()
    
    @pytest.fixture
    def sample_yaml_config(self):
        """Sample YAML configuration for testing."""
        return """
server:
  host: "127.0.0.1"
  port: 8003
  websocket_path: "/websocket"

mediamtx:
  host: "localhost"
  api_port: 9998

camera:
  poll_interval: 0.2
  enable_capability_detection: false

logging:
  level: "DEBUG"
  file_enabled: true
"""
    
    def test_config_manager_initialization(self, config_manager):
        """Test ConfigManager initializes with proper defaults."""
        assert config_manager._config is None
        assert config_manager._config_path is None
        assert config_manager._update_callbacks == []
        assert config_manager._observer is None
    
    def test_load_config_from_file(self, config_manager, sample_yaml_config):
        """Test loading configuration from YAML file."""
        # TODO: HIGH: Test YAML config file loading [Story:S14]
        # TODO: HIGH: Mock file system and YAML parsing [Story:S14]
        # TODO: HIGH: Verify Config object creation with proper values [Story:S14]
        with patch('builtins.open', mock_open(read_data=sample_yaml_config)), \
             patch('os.path.exists', return_value=True):
            
            config = config_manager.load_config('test_config.yaml')
            
            assert isinstance(config, Config)
            assert config.server.host == "127.0.0.1"
            assert config.server.port == 8003
            assert config.mediamtx.api_port == 9998
            assert config.camera.poll_interval == 0.2
            assert config.logging.level == "DEBUG"
    
    def test_load_config_file_not_found(self, config_manager):
        """Test configuration loading when no file found."""
        # TODO: HIGH: Test FileNotFoundError when no config file exists [Story:S14]
        # TODO: HIGH: Mock _find_config_file to raise FileNotFoundError [Story:S14]
        with patch.object(config_manager, '_find_config_file', side_effect=FileNotFoundError("No config found")):
            
            with pytest.raises(FileNotFoundError):
                config_manager.load_config()
    
    def test_load_config_malformed_yaml(self, config_manager):
        """Test configuration loading with malformed YAML."""
        # TODO: HIGH: Test malformed YAML handling [Story:S14]
        # TODO: HIGH: Mock file with invalid YAML content [Story:S14]
        malformed_yaml = "server:\n  host: [\n  invalid yaml"
        
        with patch('builtins.open', mock_open(read_data=malformed_yaml)), \
             patch('os.path.exists', return_value=True):
            
            with pytest.raises(ValueError) as exc_info:
                config_manager.load_config('malformed.yaml')
            
            assert "Failed to load YAML configuration" in str(exc_info.value)
    
    def test_environment_variable_overrides(self, config_manager):
        """Test environment variable overrides for configuration."""
        # TODO: HIGH: Test environment variable override functionality [Story:S14]
        # TODO: HIGH: Mock environment variables and verify override [Story:S14]
        env_vars = {
            'CAMERA_SERVICE_SERVER_HOST': '192.168.1.100',
            'CAMERA_SERVICE_SERVER_PORT': '8004',
            'CAMERA_SERVICE_CAMERA_POLL_INTERVAL': '0.5'
        }
        
        with patch('builtins.open', mock_open(read_data="{}")), \
             patch('os.path.exists', return_value=True), \
             patch.dict(os.environ, env_vars):
            
            config = config_manager.load_config('test_config.yaml')
            
            assert config.server.host == '192.168.1.100'
            assert config.server.port == 8004
            assert config.camera.poll_interval == 0.5
    
    def test_invalid_environment_variable_override(self, config_manager):
        """Test handling of invalid environment variable values."""
        # TODO: HIGH: Test invalid environment variable handling [Story:S14]
        # TODO: HIGH: Verify ValueError is raised for invalid port [Story:S14]
        env_vars = {
            'CAMERA_SERVICE_SERVER_PORT': 'invalid_port'
        }
        
        with patch('builtins.open', mock_open(read_data="{}")), \
             patch('os.path.exists', return_value=True), \
             patch.dict(os.environ, env_vars):
            
            with pytest.raises(ValueError) as exc_info:
                config_manager.load_config('test_config.yaml')
            
            assert "Invalid integer value" in str(exc_info.value)
    
    def test_configuration_validation(self, config_manager):
        """Test configuration validation with invalid values."""
        # TODO: HIGH: Test configuration validation [Story:S14]
        # TODO: HIGH: Test port range validation [Story:S14]
        # TODO: HIGH: Test logging level validation [Story:S14]
        invalid_config = """
server:
  port: 70000  # Invalid port
logging:
  level: "INVALID_LEVEL"  # Invalid log level
"""
        
        with patch('builtins.open', mock_open(read_data=invalid_config)), \
             patch('os.path.exists', return_value=True):
            
            with pytest.raises(ValueError):
                config_manager.load_config('invalid_config.yaml')
    
    def test_config_update_runtime(self, config_manager, sample_yaml_config):
        """Test runtime configuration updates."""
        # TODO: MEDIUM: Test runtime config updates [Story:S14]
        # TODO: MEDIUM: Mock initial config load then update [Story:S14]
        # TODO: MEDIUM: Verify update callbacks are triggered [Story:S14]
        with patch('builtins.open', mock_open(read_data=sample_yaml_config)), \
             patch('os.path.exists', return_value=True):
            
            config = config_manager.load_config('test_config.yaml')
            
            updates = {
                'server': {'port': 8005},
                'camera': {'poll_interval': 0.3}
            }
            
            updated_config = config_manager.update_config(updates)
            
            assert updated_config.server.port == 8005
            assert updated_config.camera.poll_interval == 0.3
    
    def test_config_update_validation_failure(self, config_manager, sample_yaml_config):
        """Test runtime config update with invalid values."""
        # TODO: MEDIUM: Test update validation failure [Story:S14]
        # TODO: MEDIUM: Verify rollback to previous config [Story:S14]
        with patch('builtins.open', mock_open(read_data=sample_yaml_config)), \
             patch('os.path.exists', return_value=True):
            
            original_config = config_manager.load_config('test_config.yaml')
            original_port = original_config.server.port
            
            invalid_updates = {
                'server': {'port': 100000}  # Invalid port
            }
            
            with pytest.raises(ValueError):
                config_manager.update_config(invalid_updates)
            
            # Should rollback to original config
            current_config = config_manager.get_config()
            assert current_config.server.port == original_port
    
    def test_hot_reload_start_stop(self, config_manager):
        """Test hot reload functionality start and stop."""
        # TODO: MEDIUM: Test hot reload start/stop [Story:S14]
        # TODO: MEDIUM: Mock watchdog Observer [Story:S14]
        # Only test if watchdog is available
        pytest.importorskip("watchdog")
        
        with patch('camera_service.config.Observer') as mock_observer:
            mock_observer_instance = Mock()
            mock_observer.return_value = mock_observer_instance
            
            config_manager.start_hot_reload()
            mock_observer_instance.start.assert_called_once()
            
            config_manager.stop_hot_reload()
            mock_observer_instance.stop.assert_called_once()
    
    def test_hot_reload_without_watchdog(self, config_manager):
        """Test hot reload when watchdog is not available."""
        # TODO: LOW: Test hot reload graceful degradation [Story:S14]
        with patch('camera_service.config.HAS_WATCHDOG', False):
            # Should not raise exception, just log warning
            config_manager.start_hot_reload()
    
    def test_config_update_callbacks(self, config_manager, sample_yaml_config):
        """Test configuration update callback functionality."""
        # TODO: MEDIUM: Test update callback registration and triggering [Story:S14]
        callback_mock = Mock()
        config_manager.add_update_callback(callback_mock)
        
        with patch('builtins.open', mock_open(read_data=sample_yaml_config)), \
             patch('os.path.exists', return_value=True):
            
            config = config_manager.load_config('test_config.yaml')
            
            updates = {'server': {'port': 8006}}
            config_manager.update_config(updates)
            
            # Callback should be called with new config
            callback_mock.assert_called_once()
    
    def test_callback_error_handling(self, config_manager, sample_yaml_config):
        """Test error handling in update callbacks."""
        # TODO: MEDIUM: Test callback error handling [Story:S14]
        # TODO: MEDIUM: Verify errors are logged but don't break update [Story:S14]
        failing_callback = Mock(side_effect=Exception("Callback error"))
        config_manager.add_update_callback(failing_callback)
        
        with patch('builtins.open', mock_open(read_data=sample_yaml_config)), \
             patch('os.path.exists', return_value=True):
            
            config_manager.load_config('test_config.yaml')
            
            # Should not raise exception despite callback failure
            updates = {'server': {'port': 8007}}
            config_manager.update_config(updates)


class TestGlobalConfigFunctions:
    """Test global configuration utility functions."""
    
    def test_load_config_function(self):
        """Test global load_config function."""
        # TODO: MEDIUM: Test global load_config function [Story:S14]
        # TODO: MEDIUM: Mock ConfigManager instance [Story:S14]
        with patch('camera_service.config._config_manager') as mock_manager:
            mock_config = Mock()
            mock_manager.load_config.return_value = mock_config
            
            result = load_config('test_path.yaml')
            
            assert result == mock_config
            mock_manager.load_config.assert_called_once_with('test_path.yaml')
    
    def test_get_config_manager_function(self):
        """Test get_config_manager function returns singleton."""
        # TODO: LOW: Test config manager singleton [Story:S14]
        manager1 = get_config_manager()
        manager2 = get_config_manager()
        
        assert manager1 is manager2
        assert isinstance(manager1, ConfigManager)


class TestConfigFindLogic:
    """Test configuration file discovery logic."""
    
    def test_find_config_file_standard_locations(self):
        """Test _find_config_file searches standard locations."""
        # TODO: MEDIUM: Test config file discovery [Story:S14]
        # TODO: MEDIUM: Mock file existence for different paths [Story:S14]
        config_manager = ConfigManager()
        
        with patch('os.path.exists') as mock_exists:
            mock_exists.side_effect = lambda path: path == "/etc/camera-service/config.yaml"
            
            result = config_manager._find_config_file()
            
            assert result == "/etc/camera-service/config.yaml"
    
    def test_find_config_file_not_found(self):
        """Test _find_config_file raises when no file found."""
        # TODO: MEDIUM: Test config file not found scenario [Story:S14]
        config_manager = ConfigManager()
        
        with patch('os.path.exists', return_value=False):
            with pytest.raises(FileNotFoundError) as exc_info:
                config_manager._find_config_file()
            
            assert "No configuration file found" in str(exc_info.value)