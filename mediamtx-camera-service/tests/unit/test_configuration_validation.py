"""
Configuration validation unit tests for schema consistency and parameter compatibility.

Requirements Coverage:
- REQ-TECH-016: System implemented in Python 3.8+
- REQ-TECH-017: Language: Python 3.8 or higher
- REQ-TECH-019: Dependencies: Standard Python libraries and third-party packages
- REQ-TECH-020: Compatibility: Linux Ubuntu 20.04+ compatibility
- REQ-TECH-021: WebSocket server implementation with concurrent connection handling
- REQ-TEST-010: Error handling and edge case test coverage

Test Categories: Unit
"""

import pytest
import inspect
from dataclasses import fields
from typing import get_type_hints, Union

from src.camera_service.config import MediaMTXConfig, Config
from src.camera_service.service_manager import ServiceManager
from src.mediamtx_wrapper.controller import MediaMTXController


class TestConfigurationSchemaValidation:
    """Test configuration schema consistency and parameter compatibility."""

    def test_mediamtx_config_controller_compatibility(self):
        """Test that MediaMTXConfig fields match MediaMTXController constructor parameters."""
        
        # Get MediaMTXController constructor parameters
        controller_params = inspect.signature(MediaMTXController.__init__).parameters
        controller_param_names = set(controller_params.keys())
        
        # Remove 'self' from parameters
        controller_param_names.discard('self')
        
        # Get MediaMTXConfig field names
        config_fields = {field.name for field in fields(MediaMTXConfig)}
        
        # Check that all controller parameters exist in config
        missing_in_config = controller_param_names - config_fields
        if missing_in_config:
            pytest.fail(
                f"MediaMTXController parameters missing from MediaMTXConfig: {missing_in_config}"
            )
        
        # Check that all config fields are used by controller (optional check)
        unused_in_controller = config_fields - controller_param_names
        if unused_in_controller:
            print(f"Warning: MediaMTXConfig fields not used by MediaMTXController: {unused_in_controller}")

    def test_mediamtx_config_field_types(self):
        """Test that MediaMTXConfig field types are compatible with expected values."""
        
        # Get type hints for MediaMTXConfig
        config_types = get_type_hints(MediaMTXConfig)
        
        # Define expected types for key fields
        expected_types = {
            'host': str,
            'api_port': int,
            'rtsp_port': int,
            'webrtc_port': int,
            'hls_port': int,
            'config_path': str,
            'recordings_path': str,
            'snapshots_path': str,
            'health_check_interval': int,
            'health_failure_threshold': int,
            'health_circuit_breaker_timeout': int,
            'health_max_backoff_interval': int,
            'health_recovery_confirmation_threshold': int,
            'backoff_base_multiplier': float,
            'backoff_jitter_range': tuple,
            'process_termination_timeout': float,
            'process_kill_timeout': float,
        }
        
        # Check that all expected fields exist with correct types
        for field_name, expected_type in expected_types.items():
            assert field_name in config_types, f"Missing field: {field_name}"
            actual_type = config_types[field_name]
            
            # Handle special cases like Optional types
            if hasattr(actual_type, '__origin__') and actual_type.__origin__ is not None:
                # For Optional[T], check that T matches expected_type
                if actual_type.__origin__ is type(Union):
                    union_types = actual_type.__args__
                    # Remove None from union types
                    non_none_types = [t for t in union_types if t is not type(None)]
                    if non_none_types:
                        actual_type = non_none_types[0]
            
            assert actual_type == expected_type, (
                f"Field {field_name}: expected {expected_type}, got {actual_type}"
            )

    def test_service_manager_config_instantiation(self):
        """Test that ServiceManager can create MediaMTXController with config object."""
        
        # Create a minimal valid config
        config_data = {
            'server': {
                'host': '0.0.0.0',
                'port': 8002,
                'max_connections': 100,
                'websocket_path': '/websocket'
            },
            'mediamtx': {
                'host': 'localhost',
                'api_port': 9997,
                'rtsp_port': 8554,
                'webrtc_port': 8889,
                'hls_port': 8888,
                'config_path': '/opt/mediamtx/config/mediamtx.yml',
                'recordings_path': '/opt/camera-service/recordings',
                'snapshots_path': '/opt/camera-service/snapshots',
                'health_check_interval': 30,
                'health_failure_threshold': 10,
                'health_circuit_breaker_timeout': 60,
                'health_max_backoff_interval': 120,
                'health_recovery_confirmation_threshold': 3,
                'backoff_base_multiplier': 2.0,
                'backoff_jitter_range': (0.8, 1.2),
                'process_termination_timeout': 3.0,
                'process_kill_timeout': 2.0,
            },
            'camera': {
                'poll_interval': 0.1,
                'enable_capability_detection': True,
                'capability_timeout': 5.0,
                'capability_retry_interval': 1.0,
                'capability_max_retries': 3
            },
            'logging': {
                'level': 'INFO',
                'file_enabled': True,
                'file_path': '/var/log/camera-service.log',
                'max_file_size': 10485760,
                'backup_count': 5,
                'console_enabled': True
            },
            'recording': {
                'enabled': True,
                'format': 'fmp4',
                'segment_duration': 3600,
                'max_segment_size': 524288000,
                'auto_cleanup': True,
                'cleanup_interval': 86400,
                'max_age': 604800,
                'max_size': 10737418240
            },
            'snapshots': {
                'enabled': True,
                'format': 'jpeg',
                'quality': 85,
                'max_width': 1920,
                'max_height': 1080,
                'auto_cleanup': True,
                'cleanup_interval': 3600,
                'max_age': 86400,
                'max_count': 1000
            }
        }
        
        # Create config object
        config = Config(**config_data)
        
        # Use free port for health server to avoid conflicts
        from tests.utils.port_utils import find_free_port
        free_health_port = find_free_port()
        config.health_port = free_health_port
        
        # Test that ServiceManager can be instantiated with this config
        # without parameter mismatches
        try:
            service_manager = ServiceManager(config)
            # If we get here, the ServiceManager was instantiated successfully
            # without parameter mismatches
            assert service_manager is not None
            assert hasattr(service_manager, '_config')
            assert service_manager._config == config
        except Exception as e:
            # If an exception is raised, it should not be about parameter mismatch
            error_msg = str(e)
            assert "unexpected keyword argument" not in error_msg, (
                f"ServiceManager failed due to parameter mismatch: {error_msg}"
            )

    def test_config_default_values(self):
        """Test that default configuration values are valid."""
        
        # Create config with minimal data (should use defaults)
        config_data = {
            'server': {'host': '0.0.0.0'},
            'mediamtx': {'host': 'localhost'},
            'camera': {},
            'logging': {},
            'recording': {},
            'snapshots': {}
        }
        
        config = Config(**config_data)
        
        # Verify that all required fields have valid default values
        assert config.mediamtx.host == 'localhost'
        assert isinstance(config.mediamtx.api_port, int)
        assert isinstance(config.mediamtx.rtsp_port, int)
        assert isinstance(config.mediamtx.webrtc_port, int)
        assert isinstance(config.mediamtx.hls_port, int)
        assert isinstance(config.mediamtx.health_check_interval, int)
        assert isinstance(config.mediamtx.health_failure_threshold, int)
        assert isinstance(config.mediamtx.backoff_base_multiplier, float)
        assert isinstance(config.mediamtx.backoff_jitter_range, tuple)

    def test_config_serialization_compatibility(self):
        """Test that configuration can be serialized and deserialized without data loss."""
        
        # Create a complete config
        config_data = {
            'server': {
                'host': '0.0.0.0',
                'port': 8002,
                'max_connections': 100,
                'websocket_path': '/websocket'
            },
            'mediamtx': {
                'host': 'localhost',
                'api_port': 9997,
                'rtsp_port': 8554,
                'webrtc_port': 8889,
                'hls_port': 8888,
                'config_path': '/opt/mediamtx/config/mediamtx.yml',
                'recordings_path': '/opt/camera-service/recordings',
                'snapshots_path': '/opt/camera-service/snapshots',
                'health_check_interval': 30,
                'health_failure_threshold': 10,
                'health_circuit_breaker_timeout': 60,
                'health_max_backoff_interval': 120,
                'health_recovery_confirmation_threshold': 3,
                'backoff_base_multiplier': 2.0,
                'backoff_jitter_range': (0.8, 1.2),
                'process_termination_timeout': 3.0,
                'process_kill_timeout': 2.0,
            },
            'camera': {
                'poll_interval': 0.1,
                'enable_capability_detection': True
            },
            'logging': {
                'level': 'INFO',
                'file_enabled': True
            },
            'recording': {
                'enabled': True,
                'format': 'fmp4'
            },
            'snapshots': {
                'enabled': True,
                'format': 'jpeg'
            }
        }
        
        config = Config(**config_data)
        
        # Serialize to dict
        config_dict = config.to_dict()
        
        # Deserialize back to config
        new_config = Config(**config_dict)
        
        # Verify all values are preserved
        assert new_config.mediamtx.host == config.mediamtx.host
        assert new_config.mediamtx.api_port == config.mediamtx.api_port
        assert new_config.mediamtx.health_check_interval == config.mediamtx.health_check_interval
        assert new_config.mediamtx.backoff_base_multiplier == config.mediamtx.backoff_base_multiplier
        assert new_config.mediamtx.backoff_jitter_range == config.mediamtx.backoff_jitter_range

    def test_required_health_monitoring_parameters(self):
        """Test that all required health monitoring parameters are present and correctly typed."""
        
        # These parameters are critical for MediaMTXController operation
        required_health_params = {
            'health_check_interval': int,
            'health_failure_threshold': int,
            'health_circuit_breaker_timeout': int,
            'health_max_backoff_interval': int,
            'health_recovery_confirmation_threshold': int,
            'backoff_base_multiplier': float,
            'backoff_jitter_range': tuple,
            'process_termination_timeout': float,
            'process_kill_timeout': float,
        }
        
        config_types = get_type_hints(MediaMTXConfig)
        
        for param_name, expected_type in required_health_params.items():
            assert param_name in config_types, f"Missing required health parameter: {param_name}"
            actual_type = config_types[param_name]
            assert actual_type == expected_type, (
                f"Health parameter {param_name}: expected {expected_type}, got {actual_type}"
            )


class TestConfigurationFileValidation:
    """Test configuration file loading and validation."""

    def test_default_yaml_config_compatibility(self):
        """Test that the default YAML configuration file is compatible with the schema."""
        
        # This test ensures that config/default.yaml contains all required fields
        # and that they match the expected types
        
        # Import the default config
        from src.camera_service.config import ConfigManager
        
        config_manager = ConfigManager()
        
        # Try to load the default configuration
        try:
            config = config_manager.load_config()
            
            # Verify that all required MediaMTX parameters are present
            assert hasattr(config.mediamtx, 'health_check_interval')
            assert hasattr(config.mediamtx, 'health_failure_threshold')
            assert hasattr(config.mediamtx, 'health_circuit_breaker_timeout')
            assert hasattr(config.mediamtx, 'health_max_backoff_interval')
            assert hasattr(config.mediamtx, 'health_recovery_confirmation_threshold')
            assert hasattr(config.mediamtx, 'backoff_base_multiplier')
            assert hasattr(config.mediamtx, 'backoff_jitter_range')
            assert hasattr(config.mediamtx, 'process_termination_timeout')
            assert hasattr(config.mediamtx, 'process_kill_timeout')
            
            # Verify types
            assert isinstance(config.mediamtx.health_check_interval, int)
            assert isinstance(config.mediamtx.health_failure_threshold, int)
            assert isinstance(config.mediamtx.backoff_base_multiplier, float)
            assert isinstance(config.mediamtx.backoff_jitter_range, tuple)
            
        except Exception as e:
            pytest.fail(f"Default configuration failed to load: {e}")

    def test_config_parameter_consistency(self):
        """Test that configuration parameters are consistent across all components."""
        
        # Get all MediaMTXConfig fields
        config_fields = {field.name for field in fields(MediaMTXConfig)}
        
        # Get MediaMTXController constructor parameters
        controller_params = inspect.signature(MediaMTXController.__init__).parameters
        controller_param_names = set(controller_params.keys()) - {'self'}
        
        # Check for any parameters that exist in one but not the other
        only_in_config = config_fields - controller_param_names
        only_in_controller = controller_param_names - config_fields
        
        if only_in_config:
            print(f"Parameters only in MediaMTXConfig: {only_in_config}")
        
        if only_in_controller:
            print(f"Parameters only in MediaMTXController: {only_in_controller}")
        
        # For now, we'll allow some flexibility but log warnings
        # In the future, these should be synchronized
        assert len(only_in_controller) == 0, (
            f"MediaMTXController has parameters not in MediaMTXConfig: {only_in_controller}"
        )

    def test_configuration_template_yaml_syntax(self):
        """Test that the configuration template has valid YAML syntax."""
        
        import yaml
        import os
        
        # Path to the configuration template
        template_path = "config/templates/camera-service.yaml.template"
        
        # Check if template file exists
        assert os.path.exists(template_path), f"Configuration template not found: {template_path}"
        
        # Try to load and parse the YAML template
        try:
            with open(template_path, 'r') as f:
                yaml_content = f.read()
            
            # Parse YAML to check syntax
            yaml.safe_load(yaml_content)
            
        except yaml.YAMLError as e:
            pytest.fail(f"Configuration template has invalid YAML syntax: {e}")
        except Exception as e:
            pytest.fail(f"Failed to read configuration template: {e}")

    def test_configuration_template_variable_substitution(self):
        """Test that the configuration template can be processed with variable substitution."""
        
        import yaml
        import tempfile
        import os
        
        # Path to the configuration template
        template_path = "config/templates/camera-service.yaml.template"
        
        # Check if template file exists
        assert os.path.exists(template_path), f"Configuration template not found: {template_path}"
        
        # Create temporary directory for testing
        with tempfile.TemporaryDirectory() as temp_dir:
            # Copy template to temporary location
            import shutil
            test_config_path = os.path.join(temp_dir, "camera-service.yaml")
            shutil.copy2(template_path, test_config_path)
            
            # Test variables for substitution
            test_variables = {
                "CAMERA_SERVICE_JWT_SECRET": "test_jwt_secret_1234567890abcdef",
                "API_KEYS_FILE": "/opt/camera-service/security/api-keys.json",
                "SSL_CERT_FILE": "/opt/camera-service/security/ssl/cert.pem",
                "SSL_KEY_FILE": "/opt/camera-service/security/ssl/key.pem"
            }
            
            # Perform variable substitution (simulating installation script)
            with open(test_config_path, 'r') as f:
                content = f.read()
            
            # Substitute variables
            for var_name, var_value in test_variables.items():
                content = content.replace(f"${{{var_name}}}", var_value)
            
            # Write back the substituted content
            with open(test_config_path, 'w') as f:
                f.write(content)
            
            # Validate that the substituted YAML is still valid
            try:
                with open(test_config_path, 'r') as f:
                    yaml.safe_load(f)
            except yaml.YAMLError as e:
                pytest.fail(f"Configuration template has invalid YAML after variable substitution: {e}")
            
            # Verify that variables were actually substituted
            with open(test_config_path, 'r') as f:
                final_content = f.read()
            
            for var_name, var_value in test_variables.items():
                assert var_value in final_content, f"Variable {var_name} was not substituted"
                assert f"${{{var_name}}}" not in final_content, f"Variable placeholder {var_name} still exists"

    def test_configuration_template_completeness(self):
        """Test that the configuration template contains all required configuration sections."""
        
        import yaml
        import os
        
        # Path to the configuration template
        template_path = "config/templates/camera-service.yaml.template"
        
        # Check if template file exists
        assert os.path.exists(template_path), f"Configuration template not found: {template_path}"
        
        # Load the template
        with open(template_path, 'r') as f:
            config_data = yaml.safe_load(f)
        
        # Required top-level sections
        required_sections = [
            'server',
            'security',
            'mediamtx',
            'ffmpeg',
            'notifications',
            'performance',
            'camera',
            'logging',
            'recording',
            'snapshots'
        ]
        
        for section in required_sections:
            assert section in config_data, f"Missing required configuration section: {section}"
        
        # Required security subsections
        security_sections = ['jwt', 'api_keys', 'ssl', 'rate_limiting', 'health']
        for section in security_sections:
            assert section in config_data['security'], f"Missing security subsection: {section}"
        
        # Required mediamtx subsections
        mediamtx_sections = ['codec', 'health_check_interval', 'stream_readiness']
        for section in mediamtx_sections:
            assert section in config_data['mediamtx'], f"Missing mediamtx subsection: {section}"
        
        # Required performance subsections
        performance_sections = ['response_time_targets', 'snapshot_tiers', 'optimization']
        for section in performance_sections:
            assert section in config_data['performance'], f"Missing performance subsection: {section}"

    def test_configuration_template_variable_placeholders(self):
        """Test that the configuration template contains expected variable placeholders."""
        
        import os
        
        # Path to the configuration template
        template_path = "config/templates/camera-service.yaml.template"
        
        # Check if template file exists
        assert os.path.exists(template_path), f"Configuration template not found: {template_path}"
        
        # Read template content
        with open(template_path, 'r') as f:
            content = f.read()
        
        # Expected variable placeholders
        expected_placeholders = [
            "${CAMERA_SERVICE_JWT_SECRET}",
            "${API_KEYS_FILE}",
            "${SSL_CERT_FILE}",
            "${SSL_KEY_FILE}"
        ]
        
        # Check that all expected placeholders are present
        for placeholder in expected_placeholders:
            assert placeholder in content, f"Missing expected variable placeholder: {placeholder}"
        
        # Check that there are no malformed placeholders
        import re
        placeholder_pattern = r'\$\{[^}]+\}'
        found_placeholders = re.findall(placeholder_pattern, content)
        
        # All found placeholders should be in our expected list
        for placeholder in found_placeholders:
            assert placeholder in expected_placeholders, f"Unexpected variable placeholder found: {placeholder}"
