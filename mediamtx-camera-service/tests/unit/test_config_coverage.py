"""
Unit tests for ConfigManager coverage gaps.

Requirements Coverage:
- REQ-CONF-001: Configuration management
- REQ-CONF-002: Hot reload functionality
- REQ-CONF-003: Configuration validation
- REQ-CONF-004: Configuration callbacks

Test Categories: Unit
API Documentation Reference: docs/api/json-rpc-methods.md
"""

import pytest
import tempfile
import time
from pathlib import Path
from unittest.mock import Mock, patch, MagicMock
from src.camera_service.config import ConfigManager, Config, ServerConfig, MediaMTXConfig


@pytest.mark.unit
class TestConfigManagerCoverage:
    """Test cases to cover missing lines in ConfigManager."""
    
    def setup_method(self):
        """Set up test fixtures."""
        self.temp_dir = tempfile.mkdtemp()
        self.config_file = Path(self.temp_dir) / "test_config.yml"
        self.config_manager = ConfigManager()
    
    def teardown_method(self):
        """Clean up test fixtures."""
        if hasattr(self, 'temp_dir'):
            import shutil
            shutil.rmtree(self.temp_dir, ignore_errors=True)
    
    def test_load_config_with_none_path_uses_defaults(self):
        """REQ-CONF-001: Test loading config with None path uses defaults."""
        # The implementation logs a warning and uses defaults
        config = self.config_manager.load_config(None)
        assert config is not None
        assert isinstance(config, Config)
    
    def test_load_config_with_empty_path_uses_defaults(self):
        """REQ-CONF-001: Test loading config with empty path uses defaults."""
        # The implementation logs a warning and uses defaults
        config = self.config_manager.load_config("")
        assert config is not None
        assert isinstance(config, Config)
    
    def test_load_config_file_not_found_uses_defaults(self):
        """REQ-CONF-001: Test loading config from non-existent file uses defaults."""
        non_existent_path = Path(self.temp_dir) / "nonexistent.yml"
        
        # The implementation logs a warning and uses defaults
        config = self.config_manager.load_config(str(non_existent_path))
        assert config is not None
        assert isinstance(config, Config)
    
    def test_start_hot_reload_without_watchdog_logs_warning(self):
        """REQ-CONF-002: Test hot reload without watchdog logs warning."""
        with patch('src.camera_service.config.HAS_WATCHDOG', False):
            # This should log a warning and return without error
            self.config_manager.start_hot_reload()
            # No exception should be raised
    
    def test_start_hot_reload_with_none_config_path_logs_warning(self):
        """REQ-CONF-002: Test hot reload with None config path logs warning."""
        self.config_manager._config_path = None
        
        with patch('src.camera_service.config.HAS_WATCHDOG', True):
            # This should log a warning and return without error
            self.config_manager.start_hot_reload()
            # No exception should be raised
    
    def test_stop_hot_reload_without_observer(self):
        """REQ-CONF-002: Test stop hot reload without observer."""
        self.config_manager._observer = None
        
        # This should not raise an error
        self.config_manager.stop_hot_reload()
    
    def test_reload_config_with_none_path(self):
        """REQ-CONF-001: Test reload config with None path."""
        self.config_manager._config_path = None
        
        with pytest.raises(RuntimeError, match="No configuration file path available for reload"):
            self.config_manager.reload_config()
    
    def test_reload_config_with_exception(self):
        """REQ-CONF-001: Test reload config with exception during load."""
        self.config_manager._config_path = str(self.config_file)
        self.config_file.touch()  # Create empty file
        
        with patch.object(self.config_manager, 'load_config', side_effect=Exception("Load failed")):
            with pytest.raises(Exception, match="Load failed"):
                self.config_manager.reload_config()
    
    def test_add_update_callback(self):
        """REQ-CONF-004: Test adding update callback."""
        callback = Mock()
        initial_count = len(self.config_manager._update_callbacks)
        
        self.config_manager.add_update_callback(callback)
        
        assert len(self.config_manager._update_callbacks) == initial_count + 1
        assert callback in self.config_manager._update_callbacks
    
    def test_remove_update_callback(self):
        """REQ-CONF-004: Test removing update callback."""
        callback = Mock()
        self.config_manager.add_update_callback(callback)
        initial_count = len(self.config_manager._update_callbacks)
        
        self.config_manager.remove_update_callback(callback)
        
        assert len(self.config_manager._update_callbacks) == initial_count - 1
        assert callback not in self.config_manager._update_callbacks
    
    def test_remove_update_callback_not_found(self):
        """REQ-CONF-004: Test removing non-existent callback."""
        callback = Mock()
        initial_count = len(self.config_manager._update_callbacks)
        
        # This should not raise an error
        self.config_manager.remove_update_callback(callback)
        
        assert len(self.config_manager._update_callbacks) == initial_count
    
    def test_notify_config_updated(self):
        """REQ-CONF-004: Test notifying config update callbacks."""
        callback1 = Mock()
        callback2 = Mock()
        self.config_manager.add_update_callback(callback1)
        self.config_manager.add_update_callback(callback2)
        
        old_config = Config()
        new_config = Config()
        
        self.config_manager._notify_config_updated(old_config, new_config)
        
        callback1.assert_called_once_with(new_config)
        callback2.assert_called_once_with(new_config)
    
    def test_notify_config_updated_with_callback_exception(self):
        """REQ-CONF-004: Test notifying config update with callback exception."""
        def failing_callback(config):
            raise Exception("Callback failed")
        
        self.config_manager.add_update_callback(failing_callback)
        
        old_config = Config()
        new_config = Config()
        
        # This should not raise an exception
        self.config_manager._notify_config_updated(old_config, new_config)
    
    def test_update_config_with_callbacks(self):
        """REQ-CONF-004: Test updating config with callbacks."""
        # First load a config
        self.config_manager._config = Config()
        
        callback = Mock()
        self.config_manager.add_update_callback(callback)
        
        old_config = self.config_manager._config
        updates = {"server": {"port": 9999}}  # Pass dictionary, not Config object
        
        self.config_manager.update_config(updates)
        
        callback.assert_called_once()
        assert self.config_manager._config.server.port == 9999
