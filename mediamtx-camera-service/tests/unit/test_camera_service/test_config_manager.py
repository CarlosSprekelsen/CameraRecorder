"""
Unit tests for camera service configuration management.

Tests the ConfigManager class for loading, validation, environment overrides,
hot reload, and error handling as specified in the architecture.
"""

import os
import pytest
from unittest.mock import Mock, patch, mock_open

from camera_service.config import (
    ConfigManager,
    load_config,
    get_config_manager,
    get_current_config,
    Config,
    ServerConfig,
    CameraConfig,
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
        with (
            patch("builtins.open", mock_open(read_data=sample_yaml_config)),
            patch("os.path.exists", return_value=True),
        ):

            config = config_manager.load_config("test_config.yaml")

            assert isinstance(config, Config)
            assert config.server.host == "127.0.0.1"
            assert config.server.port == 8003
            assert config.mediamtx.api_port == 9998
            assert config.camera.poll_interval == 0.2
            assert config.logging.level == "DEBUG"

    def test_load_config_file_not_found(self, config_manager):
        """Test configuration loading when no file found uses defaults."""
        # TODO: HIGH: Test fallback to defaults when no config file exists [Story:S14]
        # TODO: HIGH: Mock _find_config_file to raise FileNotFoundError [Story:S14]
        with patch.object(
            config_manager,
            "_find_config_file",
            side_effect=FileNotFoundError("No config found"),
        ):

            # Should not raise exception, should use defaults
            config = config_manager.load_config()
            assert isinstance(config, Config)
            assert config.server.port == 8002  # Default value

    def test_load_config_malformed_yaml(self, config_manager):
        """Test configuration loading with malformed YAML uses defaults."""
        # TODO: HIGH: Test malformed YAML handling with fallback [Story:S14]
        # TODO: HIGH: Mock file with invalid YAML content [Story:S14]
        malformed_yaml = "server:\n  host: [\n  invalid yaml"

        with (
            patch("builtins.open", mock_open(read_data=malformed_yaml)),
            patch("os.path.exists", return_value=True),
        ):

            # Should not raise exception, should use defaults and log error
            config = config_manager.load_config("malformed.yaml")
            assert isinstance(config, Config)
            assert config.server.port == 8002  # Default value

    def test_environment_variable_overrides(self, config_manager):
        """Test environment variable overrides for configuration."""
        # TODO: HIGH: Test environment variable override functionality [Story:S14]
        # TODO: HIGH: Mock environment variables and verify override [Story:S14]
        env_vars = {
            "CAMERA_SERVICE_SERVER_HOST": "192.168.1.100",
            "CAMERA_SERVICE_SERVER_PORT": "8004",
            "CAMERA_SERVICE_CAMERA_POLL_INTERVAL": "0.5",
        }

        with (
            patch("builtins.open", mock_open(read_data="{}")),
            patch("os.path.exists", return_value=True),
            patch.dict(os.environ, env_vars),
        ):

            config = config_manager.load_config("test_config.yaml")

            assert config.server.host == "192.168.1.100"
            assert config.server.port == 8004
            assert config.camera.poll_interval == 0.5

    def test_invalid_environment_variable_override(self, config_manager):
        """Test handling of invalid environment variable values."""
        # TODO: HIGH: Test invalid environment variable handling with fallback [Story:S14]
        # TODO: HIGH: Verify invalid values are logged but service continues [Story:S14]
        env_vars = {"CAMERA_SERVICE_SERVER_PORT": "invalid_port"}

        with (
            patch("builtins.open", mock_open(read_data="{}")),
            patch("os.path.exists", return_value=True),
            patch.dict(os.environ, env_vars),
        ):

            # Should not crash, should use default and log error
            config = config_manager.load_config("test_config.yaml")
            assert isinstance(config, Config)
            assert config.server.port == 8002  # Default value

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

        with (
            patch("builtins.open", mock_open(read_data=invalid_config)),
            patch("os.path.exists", return_value=True),
        ):

            with pytest.raises(ValueError):
                config_manager.load_config("invalid_config.yaml")

    def test_config_update_runtime(self, config_manager, sample_yaml_config):
        """Test runtime configuration updates."""
        # TODO: MEDIUM: Test runtime config updates [Story:S14]
        # TODO: MEDIUM: Mock initial config load then update [Story:S14]
        # TODO: MEDIUM: Verify update callbacks are triggered [Story:S14]
        with (
            patch("builtins.open", mock_open(read_data=sample_yaml_config)),
            patch("os.path.exists", return_value=True),
        ):

            config_manager.load_config("test_config.yaml")

            updates = {"server": {"port": 8005}, "camera": {"poll_interval": 0.3}}

            updated_config = config_manager.update_config(updates)

            assert updated_config.server.port == 8005
            assert updated_config.camera.poll_interval == 0.3

    def test_config_update_validation_failure(self, config_manager, sample_yaml_config):
        """Test runtime config update with invalid values."""
        # TODO: MEDIUM: Test update validation failure with rollback [Story:S14]
        # TODO: MEDIUM: Verify rollback to previous config [Story:S14]
        with (
            patch("builtins.open", mock_open(read_data=sample_yaml_config)),
            patch("os.path.exists", return_value=True),
        ):

            original_config = config_manager.load_config("test_config.yaml")
            original_port = original_config.server.port

            invalid_updates = {"server": {"port": 100000}}  # Invalid port

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

        with patch("camera_service.config.Observer") as mock_observer:
            mock_observer_instance = Mock()
            mock_observer.return_value = mock_observer_instance

            # Set a config path so hot reload can start
            config_manager._config_path = "/tmp/test_config.yaml"
            
            config_manager.start_hot_reload()
            mock_observer_instance.start.assert_called_once()

            config_manager.stop_hot_reload()
            mock_observer_instance.stop.assert_called_once()

    def test_hot_reload_without_watchdog(self, config_manager):
        """Test hot reload when watchdog is not available."""
        # TODO: MEDIUM: Test hot reload graceful degradation [Story:S14]
        # TODO: MEDIUM: Mock HAS_WATCHDOG to False [Story:S14]
        with patch("camera_service.config.HAS_WATCHDOG", False):
            # Should log warning but not crash
            config_manager.start_hot_reload()
            # No observer should be created
            assert config_manager._observer is None

    def test_hot_reload_file_change_simulation(
        self, config_manager, sample_yaml_config
    ):
        """Test hot reload file change detection and reload."""
        # TODO: MEDIUM: Test hot reload file change simulation [Story:S14]
        # TODO: MEDIUM: Mock file system events and verify reload [Story:S14]
        pytest.importorskip("watchdog")

        with (
            patch("builtins.open", mock_open(read_data=sample_yaml_config)),
            patch("os.path.exists", return_value=True),
            patch("camera_service.config.Observer") as mock_observer,
        ):

            # Load initial config
            config_manager.load_config("test_config.yaml")

            # Start hot reload
            config_manager.start_hot_reload()

            # Verify observer was configured
            mock_observer.assert_called_once()

    def test_default_config_fallback(self, config_manager):
        """Test fallback to default configuration when all else fails."""
        # TODO: HIGH: Test complete fallback to defaults [Story:S14]
        # TODO: HIGH: Mock all config sources to fail [Story:S14]
        with patch.object(
            config_manager, "_find_config_file", side_effect=FileNotFoundError()
        ):

            config = config_manager.load_config()

            # Should get default configuration
            assert isinstance(config, Config)
            assert config.server.host == "0.0.0.0"
            assert config.server.port == 8002
            assert config.mediamtx.api_port == 9997
            assert config.camera.poll_interval == 0.1

    def test_comprehensive_validation_error_accumulation(self, config_manager):
        """Test that validation accumulates multiple errors."""
        # TODO: HIGH: Test validation error accumulation [Story:S14]
        # TODO: HIGH: Verify multiple validation errors are collected [Story:S14]
        invalid_config = """
server:
  port: 999999  # Invalid port
  max_connections: -1  # Invalid negative value
mediamtx:
  api_port: 0  # Invalid port
logging:
  level: "INVALID"  # Invalid level
snapshots:
  quality: 150  # Invalid quality
"""

        with (
            patch("builtins.open", mock_open(read_data=invalid_config)),
            patch("os.path.exists", return_value=True),
        ):

            with pytest.raises(ValueError) as exc_info:
                config_manager.load_config("invalid_config.yaml")

            # Should contain multiple validation errors
            error_message = str(exc_info.value)
            assert "Configuration validation failed" in error_message
            assert "999999" in error_message  # The invalid port value
            assert "must be integer 1-65535" in error_message  # The actual error message


class TestConfigurationIntegration:
    """Integration tests for configuration loading and management."""

    def test_load_config_function(self):
        """Test module-level load_config function."""
        # TODO: MEDIUM: Test module-level config loading [Story:S14]
        # TODO: MEDIUM: Mock file system for function test [Story:S14]
        with patch("camera_service.config._config_manager") as mock_manager:
            mock_config = Mock()
            mock_manager.load_config.return_value = mock_config

            result = load_config("test.yaml")

            mock_manager.load_config.assert_called_once_with("test.yaml")
            assert result == mock_config

    def test_get_config_manager_function(self):
        """Test module-level get_config_manager function."""
        # TODO: LOW: Test config manager accessor [Story:S14]
        manager = get_config_manager()
        assert isinstance(manager, ConfigManager)

    def test_get_current_config_function(self):
        """Test module-level get_current_config function."""
        # TODO: LOW: Test current config accessor [Story:S14]
        with patch("camera_service.config._config_manager") as mock_manager:
            mock_config = Mock()
            mock_manager.get_config.return_value = mock_config

            result = get_current_config()

            mock_manager.get_config.assert_called_once()
            assert result == mock_config
