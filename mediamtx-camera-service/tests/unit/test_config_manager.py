"""
Real integration tests for configuration management without mocks.

Tests the ConfigManager class with real file system operations,
real environment variables, and real component integration.

Requirements:
- REQ-CONFIG-002: System shall support configuration hot reload
- REQ-CONFIG-003: System shall validate configuration parameters at runtime
- REQ-ERROR-004: System shall handle configuration loading failures gracefully
- REQ-ERROR-005: System shall provide meaningful error messages for configuration issues

Story Coverage: S14 - Configuration Management
IV&V Control Point: Real system integration validation
"""

import os
import tempfile
import yaml
from pathlib import Path

import pytest

from camera_service.config import (
    ConfigManager,
    load_config,
    get_config_manager,
    get_current_config,
    Config,
    ServerConfig,
    CameraConfig,
)


class TestConfigManagerRealIntegration:
    """Real integration tests for ConfigManager without mocks."""

    @pytest.fixture
    def temp_config_dir(self):
        """Create a temporary directory for configuration files."""
        with tempfile.TemporaryDirectory() as temp_dir:
            yield Path(temp_dir)

    @pytest.fixture
    def config_manager(self):
        """Create a fresh ConfigManager instance for testing."""
        return ConfigManager()

    def test_real_config_file_loading(self, config_manager, temp_config_dir):
        """Test loading configuration from a real YAML file."""
        config_file = temp_config_dir / "test_config.yaml"
        
        # Create a real YAML configuration file
        config_data = {
            "server": {
                "host": "127.0.0.1",
                "port": 8003,
                "websocket_path": "/websocket"
            },
            "mediamtx": {
                "host": "localhost",
                "api_port": 9998
            },
            "camera": {
                "poll_interval": 0.2,
                "enable_capability_detection": False
            },
            "logging": {
                "level": "DEBUG",
                "file_enabled": True
            }
        }
        
        with open(config_file, 'w') as f:
            yaml.dump(config_data, f)

        # Load configuration from real file
        config = config_manager.load_config(str(config_file))

        assert isinstance(config, Config)
        assert config.server.host == "127.0.0.1"
        assert config.server.port == 8003
        assert config.mediamtx.api_port == 9998
        assert config.camera.poll_interval == 0.2
        assert config.logging.level == "DEBUG"

    def test_real_config_file_not_found(self, config_manager, temp_config_dir):
        """Test configuration loading when real file doesn't exist."""
        non_existent_file = temp_config_dir / "non_existent.yaml"
        
        # Should not raise exception, should use defaults
        config = config_manager.load_config(str(non_existent_file))
        assert isinstance(config, Config)
        assert config.server.port == 8002  # Default value

    def test_real_malformed_yaml_handling(self, config_manager, temp_config_dir):
        """Test handling of real malformed YAML file."""
        config_file = temp_config_dir / "malformed.yaml"
        
        # Create a malformed YAML file
        malformed_yaml = "server:\n  host: [\n  invalid yaml"
        with open(config_file, 'w') as f:
            f.write(malformed_yaml)

        # Should not raise exception, should use defaults and log error
        config = config_manager.load_config(str(config_file))
        assert isinstance(config, Config)
        assert config.server.port == 8002  # Default value

    def test_real_environment_variable_overrides(self, config_manager, temp_config_dir):
        """Test real environment variable overrides."""
        config_file = temp_config_dir / "test_config.yaml"
        
        # Create minimal config file
        config_data = {"server": {"host": "default.host"}}
        with open(config_file, 'w') as f:
            yaml.dump(config_data, f)

        # Set real environment variables
        original_env = os.environ.copy()
        try:
            os.environ["CAMERA_SERVICE_SERVER_HOST"] = "192.168.1.100"
            os.environ["CAMERA_SERVICE_SERVER_PORT"] = "8004"
            os.environ["CAMERA_SERVICE_CAMERA_POLL_INTERVAL"] = "0.5"

            config = config_manager.load_config(str(config_file))

            assert config.server.host == "192.168.1.100"
            assert config.server.port == 8004
            assert config.camera.poll_interval == 0.5
        finally:
            # Restore original environment
            os.environ.clear()
            os.environ.update(original_env)

    def test_real_invalid_environment_variable_handling(self, config_manager, temp_config_dir):
        """Test handling of real invalid environment variable values."""
        config_file = temp_config_dir / "test_config.yaml"
        
        # Create minimal config file
        config_data = {"server": {"port": 8002}}
        with open(config_file, 'w') as f:
            yaml.dump(config_data, f)

        # Set invalid environment variable
        original_env = os.environ.copy()
        try:
            os.environ["CAMERA_SERVICE_SERVER_PORT"] = "invalid_port"

            # Should not crash, should use default and log error
            config = config_manager.load_config(str(config_file))
            assert isinstance(config, Config)
            assert config.server.port == 8002  # Default value
        finally:
            # Restore original environment
            os.environ.clear()
            os.environ.update(original_env)

    def test_real_configuration_validation(self, config_manager, temp_config_dir):
        """Test real configuration validation with invalid values."""
        config_file = temp_config_dir / "invalid_config.yaml"
        
        # Create config file with invalid values
        invalid_config = {
            "server": {"port": 70000},  # Invalid port
            "logging": {"level": "INVALID_LEVEL"}  # Invalid log level
        }
        
        with open(config_file, 'w') as f:
            yaml.dump(invalid_config, f)

        # Should raise ValueError for invalid configuration
        with pytest.raises(ValueError):
            config_manager.load_config(str(config_file))

    def test_real_config_update_runtime(self, config_manager, temp_config_dir):
        """Test real runtime configuration updates."""
        config_file = temp_config_dir / "test_config.yaml"
        
        # Create initial config file
        config_data = {
            "server": {"port": 8002},
            "camera": {"poll_interval": 0.1}
        }
        with open(config_file, 'w') as f:
            yaml.dump(config_data, f)

        # Load initial config
        config_manager.load_config(str(config_file))

        # Update configuration
        updates = {
            "server": {"port": 8005},
            "camera": {"poll_interval": 0.3}
        }

        updated_config = config_manager.update_config(updates)

        assert updated_config.server.port == 8005
        assert updated_config.camera.poll_interval == 0.3

    def test_real_config_update_validation_failure(self, config_manager, temp_config_dir):
        """Test real runtime config update with invalid values."""
        config_file = temp_config_dir / "test_config.yaml"
        
        # Create initial config file
        config_data = {"server": {"port": 8002}}
        with open(config_file, 'w') as f:
            yaml.dump(config_data, f)

        original_config = config_manager.load_config(str(config_file))
        original_port = original_config.server.port

        invalid_updates = {"server": {"port": 100000}}  # Invalid port

        with pytest.raises(ValueError):
            config_manager.update_config(invalid_updates)

        # Should rollback to original config
        current_config = config_manager.get_config()
        assert current_config.server.port == original_port

    def test_real_hot_reload_functionality(self, config_manager, temp_config_dir):
        """Test real hot reload functionality if watchdog is available."""
        try:
            from watchdog.observers import Observer
            HAS_WATCHDOG = True
        except ImportError:
            HAS_WATCHDOG = False
            pytest.skip("watchdog not available")

        config_file = temp_config_dir / "test_config.yaml"
        
        # Create initial config file
        config_data = {"server": {"port": 8002}}
        with open(config_file, 'w') as f:
            yaml.dump(config_data, f)

        # Load config and start hot reload
        config_manager.load_config(str(config_file))
        config_manager.start_hot_reload()

        try:
            # Verify hot reload is running
            assert config_manager._observer is not None
            assert config_manager._observer.is_alive()
        finally:
            # Stop hot reload
            config_manager.stop_hot_reload()

    def test_real_default_config_fallback(self, config_manager):
        """Test real fallback to default configuration when all else fails."""
        # Test with non-existent directory
        non_existent_path = "/non/existent/path/config.yaml"
        
        config = config_manager.load_config(non_existent_path)

        # Should get default configuration
        assert isinstance(config, Config)
        assert config.server.host == "0.0.0.0"
        assert config.server.port == 8002
        assert config.mediamtx.api_port == 9997
        assert config.camera.poll_interval == 0.1


class TestConfigurationIntegrationReal:
    """Real integration tests for configuration module functions."""

    @pytest.fixture
    def temp_config_dir(self):
        """Create a temporary directory for configuration files."""
        with tempfile.TemporaryDirectory() as temp_dir:
            yield Path(temp_dir)

    def test_real_load_config_function(self, temp_config_dir):
        """Test module-level load_config function with real file."""
        config_file = temp_config_dir / "test_config.yaml"
        
        # Create real config file
        config_data = {"server": {"port": 8005}}
        with open(config_file, 'w') as f:
            yaml.dump(config_data, f)

        result = load_config(str(config_file))
        assert isinstance(result, Config)
        assert result.server.port == 8005

    def test_real_get_config_manager_function(self):
        """Test module-level get_config_manager function."""
        manager = get_config_manager()
        assert isinstance(manager, ConfigManager)

    def test_real_get_current_config_function(self, temp_config_dir):
        """Test module-level get_current_config function."""
        config_file = temp_config_dir / "test_config.yaml"
        
        # Create real config file
        config_data = {"server": {"port": 8006}}
        with open(config_file, 'w') as f:
            yaml.dump(config_data, f)

        # Load config first
        load_config(str(config_file))
        
        # Get current config
        result = get_current_config()
        assert isinstance(result, Config)
        assert result.server.port == 8006
