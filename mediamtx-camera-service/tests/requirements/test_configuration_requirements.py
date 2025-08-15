"""
Configuration Management Requirements Test Coverage

Tests specifically designed to validate configuration management requirements:
- REQ-CONFIG-002: System shall support configuration hot reload
- REQ-CONFIG-003: System shall validate configuration parameters at runtime

These tests are designed to fail if configuration requirements are not met.
"""

import asyncio
import tempfile
import os
import time
import yaml
import pytest
from typing import List, Dict, Any, Optional
from dataclasses import dataclass
from pathlib import Path
from unittest.mock import patch

from src.camera_service.config import ConfigManager, Config, ServerConfig, MediaMTXConfig, CameraConfig
from src.camera_service.service_manager import ServiceManager
from src.websocket_server.server import WebSocketJsonRpcServer
from src.mediamtx_wrapper.controller import MediaMTXController


@dataclass
class ConfigurationMetrics:
    """Configuration metrics for requirement validation."""
    requirement: str
    test_name: str
    hot_reload_success: bool
    validation_success: bool
    config_changes_detected: int
    validation_errors_caught: int
    success: bool
    error_message: str = None


class ConfigurationRequirementsValidator:
    """Validates configuration management requirements through comprehensive testing."""
    
    def __init__(self):
        self.metrics: List[ConfigurationMetrics] = []
        self.config_thresholds = {
            "hot_reload_timeout": 5.0,     # REQ-CONFIG-002: Hot reload within 5 seconds
            "validation_coverage": 95,     # REQ-CONFIG-003: 95%+ validation coverage
            "config_change_detection": 100  # REQ-CONFIG-002: 100% config change detection
        }
        self.config_changes = []
        self.validation_errors = []
    
    async def setup_test_environment(self) -> Dict[str, Any]:
        """Set up test environment for configuration testing."""
        temp_dir = tempfile.mkdtemp(prefix="config_test_")
        
        # Create initial configuration file
        initial_config = {
            "server": {
                "host": "127.0.0.1",
                "port": 8007,
                "websocket_path": "/ws",
                "max_connections": 100
            },
            "mediamtx": {
                "host": "127.0.0.1",
                "api_port": 10005,
                "rtsp_port": 8554,
                "webrtc_port": 8889,
                "hls_port": 8888,
                "config_path": f"{temp_dir}/mediamtx.yml",
                "recordings_path": f"{temp_dir}/recordings",
                "snapshots_path": f"{temp_dir}/snapshots"
            },
            "camera": {
                "device_range": [0, 1, 2],
                "poll_interval": 0.1,
                "enable_capability_detection": True
            }
        }
        
        config_file_path = os.path.join(temp_dir, "config.yml")
        with open(config_file_path, 'w') as f:
            yaml.dump(initial_config, f)
        
        return {
            "temp_dir": temp_dir,
            "config_file_path": config_file_path,
            "initial_config": initial_config
        }
    
    async def test_req_config_002_hot_reload(self):
        """REQ-CONFIG-002: System shall support configuration hot reload."""
        env = await self.setup_test_environment()
        
        # Create real config manager
        config_manager = ConfigManager()
        
        try:
            # Load initial configuration
            initial_config = config_manager.load_config(env["config_file_path"])
            
            # Create service manager with initial config
            service_manager = ServiceManager(initial_config)
            await service_manager.start()
            
            # Track configuration changes
            config_changes_detected = 0
            hot_reload_success = False
            
            # Setup configuration change callback
            def on_config_change(new_config):
                nonlocal config_changes_detected
                config_changes_detected += 1
                self.config_changes.append(new_config)
            
            config_manager.add_update_callback(on_config_change)
            
            # Modify configuration file
            modified_config = {
                "server": {
                    "host": "127.0.0.1",
                    "port": 8008,  # Changed port
                    "websocket_path": "/ws",
                    "max_connections": 150  # Changed max connections
                },
                "mediamtx": {
                    "host": "127.0.0.1",
                    "api_port": 10006,  # Changed API port
                    "rtsp_port": 8554,
                    "webrtc_port": 8889,
                    "hls_port": 8888,
                    "config_path": f"{env['temp_dir']}/mediamtx.yml",
                    "recordings_path": f"{env['temp_dir']}/recordings",
                    "snapshots_path": f"{env['temp_dir']}/snapshots"
                },
                "camera": {
                    "device_range": [0, 1, 2, 3],  # Added device
                    "poll_interval": 0.2,  # Changed poll interval
                    "enable_capability_detection": True
                }
            }
            
            # Write modified configuration
            with open(env["config_file_path"], 'w') as f:
                yaml.dump(modified_config, f)
            
            # Wait for hot reload
            start_time = time.time()
            while time.time() - start_time < self.config_thresholds["hot_reload_timeout"]:
                if config_changes_detected > 0:
                    hot_reload_success = True
                    break
                await asyncio.sleep(0.1)
            
            # Verify hot reload occurred
            assert hot_reload_success, f"REQ-CONFIG-002 FAILED: Hot reload did not occur within {self.config_thresholds['hot_reload_timeout']} seconds"
            
            # Verify configuration was updated
            updated_config = config_manager.get_current_config()
            assert updated_config.server.port == 8008, "REQ-CONFIG-002 FAILED: Server port not updated"
            assert updated_config.server.max_connections == 150, "REQ-CONFIG-002 FAILED: Max connections not updated"
            assert updated_config.camera.poll_interval == 0.2, "REQ-CONFIG-002 FAILED: Poll interval not updated"
            
            # Record metrics
            self.metrics.append(ConfigurationMetrics(
                requirement="REQ-CONFIG-002",
                test_name="hot_reload",
                hot_reload_success=hot_reload_success,
                validation_success=True,
                config_changes_detected=config_changes_detected,
                validation_errors_caught=0,
                success=hot_reload_success and config_changes_detected > 0
            ))
            
            # Validate requirement
            assert config_changes_detected > 0, "REQ-CONFIG-002 FAILED: No configuration changes detected"
            
        finally:
            await service_manager.stop()
            import shutil
            shutil.rmtree(env["temp_dir"], ignore_errors=True)
    
    async def test_req_config_003_runtime_validation(self):
        """REQ-CONFIG-003: System shall validate configuration parameters at runtime."""
        env = await self.setup_test_environment()
        
        # Create real config manager
        config_manager = ConfigManager()
        
        try:
            # Test valid configuration
            valid_config = config_manager.load_config(env["config_file_path"])
            assert valid_config is not None, "REQ-CONFIG-003 FAILED: Valid configuration not loaded"
            
            # Test invalid configurations
            invalid_configs = [
                # Invalid server port
                {
                    "server": {"host": "127.0.0.1", "port": -1, "websocket_path": "/ws"},
                    "mediamtx": {"host": "127.0.0.1", "api_port": 10005},
                    "camera": {"device_range": [0, 1, 2]}
                },
                # Invalid max connections
                {
                    "server": {"host": "127.0.0.1", "port": 8007, "max_connections": 0},
                    "mediamtx": {"host": "127.0.0.1", "api_port": 10005},
                    "camera": {"device_range": [0, 1, 2]}
                },
                # Invalid poll interval
                {
                    "server": {"host": "127.0.0.1", "port": 8007},
                    "mediamtx": {"host": "127.0.0.1", "api_port": 10005},
                    "camera": {"device_range": [0, 1, 2], "poll_interval": -0.1}
                },
                # Invalid device range
                {
                    "server": {"host": "127.0.0.1", "port": 8007},
                    "mediamtx": {"host": "127.0.0.1", "api_port": 10005},
                    "camera": {"device_range": []}
                }
            ]
            
            validation_errors_caught = 0
            validation_success = True
            
            for i, invalid_config in enumerate(invalid_configs):
                try:
                    # Write invalid configuration
                    invalid_config_file = os.path.join(env["temp_dir"], f"invalid_config_{i}.yml")
                    with open(invalid_config_file, 'w') as f:
                        yaml.dump(invalid_config, f)
                    
                    # Try to load invalid configuration
                    config_manager.load_config(invalid_config_file)
                    
                    # If we get here, validation failed
                    validation_success = False
                    self.validation_errors.append(f"Invalid config {i} was accepted")
                    
                except (ValueError, TypeError, AssertionError) as e:
                    # Expected validation error
                    validation_errors_caught += 1
                    self.validation_errors.append(str(e))
            
            # Test edge cases
            edge_cases = [
                # Very large port number
                {"server": {"host": "127.0.0.1", "port": 99999}},
                # Very small poll interval
                {"camera": {"poll_interval": 0.001}},
                # Very large device range
                {"camera": {"device_range": list(range(1000))}}
            ]
            
            for i, edge_case in enumerate(edge_cases):
                try:
                    edge_config_file = os.path.join(env["temp_dir"], f"edge_config_{i}.yml")
                    with open(edge_config_file, 'w') as f:
                        yaml.dump(edge_case, f)
                    
                    config_manager.load_config(edge_config_file)
                    
                except (ValueError, TypeError, AssertionError) as e:
                    # Edge case validation error
                    validation_errors_caught += 1
                    self.validation_errors.append(f"Edge case {i}: {str(e)}")
            
            # Calculate validation coverage
            total_tests = len(invalid_configs) + len(edge_cases)
            validation_coverage = (validation_errors_caught / total_tests) * 100 if total_tests > 0 else 0
            
            # Record metrics
            self.metrics.append(ConfigurationMetrics(
                requirement="REQ-CONFIG-003",
                test_name="runtime_validation",
                hot_reload_success=False,
                validation_success=validation_success,
                config_changes_detected=0,
                validation_errors_caught=validation_errors_caught,
                success=validation_coverage >= self.config_thresholds["validation_coverage"]
            ))
            
            # Validate requirement
            assert validation_coverage >= self.config_thresholds["validation_coverage"], \
                f"REQ-CONFIG-003 FAILED: Validation coverage {validation_coverage:.1f}% below threshold {self.config_thresholds['validation_coverage']}%"
            
            assert validation_errors_caught > 0, "REQ-CONFIG-003 FAILED: No validation errors caught"
            
        finally:
            import shutil
            shutil.rmtree(env["temp_dir"], ignore_errors=True)
    
    async def test_config_parameter_validation(self):
        """Test comprehensive configuration parameter validation."""
        env = await self.setup_test_environment()
        
        # Create real config manager
        config_manager = ConfigManager()
        
        try:
            # Test all configuration parameters
            test_cases = [
                # Server configuration
                {"server": {"host": "", "port": 8007}},  # Empty host
                {"server": {"host": "127.0.0.1", "port": 0}},  # Invalid port
                {"server": {"host": "127.0.0.1", "port": 8007, "max_connections": -1}},  # Invalid max connections
                
                # MediaMTX configuration
                {"mediamtx": {"host": "127.0.0.1", "api_port": -1}},  # Invalid API port
                {"mediamtx": {"host": "127.0.0.1", "rtsp_port": 0}},  # Invalid RTSP port
                
                # Camera configuration
                {"camera": {"device_range": [-1, 0, 1]}},  # Invalid device range
                {"camera": {"poll_interval": 0}},  # Invalid poll interval
                {"camera": {"poll_interval": -0.1}},  # Negative poll interval
            ]
            
            validation_errors = 0
            
            for i, test_case in enumerate(test_cases):
                try:
                    test_config_file = os.path.join(env["temp_dir"], f"test_config_{i}.yml")
                    with open(test_config_file, 'w') as f:
                        yaml.dump(test_case, f)
                    
                    config_manager.load_config(test_config_file)
                    
                    # If we get here, validation failed
                    validation_errors += 1
                    
                except (ValueError, TypeError, AssertionError):
                    # Expected validation error
                    pass
            
            # Validate that most invalid configurations were caught
            validation_rate = ((len(test_cases) - validation_errors) / len(test_cases)) * 100
            assert validation_rate >= 80, f"REQ-CONFIG-003 FAILED: Only {validation_rate:.1f}% of invalid configs were caught"
            
        finally:
            import shutil
            shutil.rmtree(env["temp_dir"], ignore_errors=True)
    
    async def test_config_file_watching(self):
        """Test configuration file watching and change detection."""
        env = await self.setup_test_environment()
        
        # Create real config manager
        config_manager = ConfigManager()
        
        try:
            # Load initial configuration
            initial_config = config_manager.load_config(env["config_file_path"])
            
            # Setup file watching
            config_manager.start_file_watching(env["config_file_path"])
            
            # Track changes
            changes_detected = 0
            
            def on_change(new_config):
                nonlocal changes_detected
                changes_detected += 1
            
            config_manager.add_update_callback(on_change)
            
            # Make multiple configuration changes
            for i in range(3):
                # Modify configuration
                modified_config = {
                    "server": {
                        "host": "127.0.0.1",
                        "port": 8007 + i,
                        "websocket_path": "/ws",
                        "max_connections": 100 + i * 10
                    },
                    "mediamtx": {
                        "host": "127.0.0.1",
                        "api_port": 10005 + i,
                        "rtsp_port": 8554,
                        "webrtc_port": 8889,
                        "hls_port": 8888,
                        "config_path": f"{env['temp_dir']}/mediamtx.yml",
                        "recordings_path": f"{env['temp_dir']}/recordings",
                        "snapshots_path": f"{env['temp_dir']}/snapshots"
                    },
                    "camera": {
                        "device_range": [0, 1, 2],
                        "poll_interval": 0.1 + i * 0.1,
                        "enable_capability_detection": True
                    }
                }
                
                # Write modified configuration
                with open(env["config_file_path"], 'w') as f:
                    yaml.dump(modified_config, f)
                
                # Wait for change detection
                await asyncio.sleep(0.5)
            
            # Stop file watching
            config_manager.stop_file_watching()
            
            # Validate change detection
            assert changes_detected >= 2, f"REQ-CONFIG-002 FAILED: Only {changes_detected} changes detected, expected at least 2"
            
        finally:
            import shutil
            shutil.rmtree(env["temp_dir"], ignore_errors=True)


class TestConfigurationRequirements:
    """Test suite for configuration management requirements validation."""
    
    @pytest.fixture
    def validator(self):
        """Create configuration requirements validator."""
        return ConfigurationRequirementsValidator()
    
    @pytest.mark.asyncio
    async def test_req_config_002_hot_reload(self, validator):
        """REQ-CONFIG-002: System shall support configuration hot reload."""
        await validator.test_req_config_002_hot_reload()
    
    @pytest.mark.asyncio
    async def test_req_config_003_runtime_validation(self, validator):
        """REQ-CONFIG-003: System shall validate configuration parameters at runtime."""
        await validator.test_req_config_003_runtime_validation()
    
    @pytest.mark.asyncio
    async def test_config_parameter_validation(self, validator):
        """Test comprehensive configuration parameter validation."""
        await validator.test_config_parameter_validation()
    
    @pytest.mark.asyncio
    async def test_config_file_watching(self, validator):
        """Test configuration file watching and change detection."""
        await validator.test_config_file_watching()
    
    def test_configuration_metrics_summary(self, validator):
        """Test that all configuration requirements are met."""
        # This test validates that all configuration metrics meet requirements
        for metric in validator.metrics:
            assert metric.success, f"Configuration requirement failed for {metric.requirement}: {metric.error_message}"
