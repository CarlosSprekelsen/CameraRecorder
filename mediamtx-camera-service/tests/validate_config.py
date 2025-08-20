#!/usr/bin/env python3
"""
Configuration validation script.

Requirements Coverage:
- REQ-CONFIG-001: System shall validate configuration compatibility
- REQ-CONFIG-002: System shall ensure MediaMTXConfig matches MediaMTXController
- REQ-CONFIG-003: System shall validate configuration field types
- REQ-CONFIG-004: System shall catch configuration mismatches before runtime

Test Categories: Configuration

This script validates that:
1. MediaMTXConfig dataclass matches MediaMTXController constructor parameters
2. Configuration files can be loaded without parameter mismatches
3. All required parameters are present and correctly typed

Run this script to catch configuration mismatches before they cause runtime errors.
"""

import sys
import inspect
from dataclasses import fields
from typing import get_type_hints, Union

# Add src to path for imports
sys.path.insert(0, 'src')

try:
    from camera_service.config import MediaMTXConfig, Config
    from mediamtx_wrapper.controller import MediaMTXController
except ImportError as e:
    print(f"Import error: {e}")
    print("Make sure you're running this from the project root directory")
    sys.exit(1)


@pytest.mark.config
def test_mediamtx_config_controller_compatibility():
    """Test that MediaMTXConfig fields match MediaMTXController constructor parameters."""
    print("Testing MediaMTXConfig and MediaMTXController parameter compatibility...")
    
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
        print(f"❌ ERROR: MediaMTXController parameters missing from MediaMTXConfig: {missing_in_config}")
        return False
    
    # Check that all config fields are used by controller (optional check)
    unused_in_controller = config_fields - controller_param_names
    if unused_in_controller:
        print(f"⚠️  WARNING: MediaMTXConfig fields not used by MediaMTXController: {unused_in_controller}")
    
    print("✅ MediaMTXConfig and MediaMTXController parameters are compatible")
    return True


@pytest.mark.config
def test_mediamtx_config_field_types():
    """Test that MediaMTXConfig field types are compatible with expected values."""
    print("Testing MediaMTXConfig field types...")
    
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
    
    errors = []
    
    # Check that all expected fields exist with correct types
    for field_name, expected_type in expected_types.items():
        if field_name not in config_types:
            errors.append(f"Missing field: {field_name}")
            continue
            
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
        
        if actual_type != expected_type:
            errors.append(f"Field {field_name}: expected {expected_type}, got {actual_type}")
    
    if errors:
        print("❌ Type validation errors:")
        for error in errors:
            print(f"  - {error}")
        return False
    
    print("✅ MediaMTXConfig field types are correct")
    return True


@pytest.mark.config
def test_required_health_monitoring_parameters():
    """Test that all required health monitoring parameters are present and correctly typed."""
    print("Testing required health monitoring parameters...")
    
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
    errors = []
    
    for param_name, expected_type in required_health_params.items():
        if param_name not in config_types:
            errors.append(f"Missing required health parameter: {param_name}")
            continue
            
        actual_type = config_types[param_name]
        if actual_type != expected_type:
            errors.append(f"Health parameter {param_name}: expected {expected_type}, got {actual_type}")
    
    if errors:
        print("❌ Health monitoring parameter errors:")
        for error in errors:
            print(f"  - {error}")
        return False
    
    print("✅ All required health monitoring parameters are present and correctly typed")
    return True


@pytest.mark.config
def test_config_instantiation():
    """Test that configuration can be instantiated with all required parameters."""
    print("Testing configuration instantiation...")
    
    try:
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
        
        # Create config object
        config = Config(**config_data)
        
        # Verify that all required fields are present
        assert hasattr(config.mediamtx, 'health_check_interval')
        assert hasattr(config.mediamtx, 'health_failure_threshold')
        assert hasattr(config.mediamtx, 'backoff_base_multiplier')
        assert hasattr(config.mediamtx, 'backoff_jitter_range')
        
        print("✅ Configuration instantiation successful")
        return True
        
    except Exception as e:
        print(f"❌ Configuration instantiation failed: {e}")
        return False


def main():
    """Run all configuration validation tests."""
    print("🔍 Running configuration validation tests...")
    print("=" * 50)
    
    tests = [
        test_mediamtx_config_controller_compatibility,
        test_mediamtx_config_field_types,
        test_required_health_monitoring_parameters,
        test_config_instantiation,
    ]
    
    passed = 0
    total = len(tests)
    
    for test in tests:
        try:
            if test():
                passed += 1
        except Exception as e:
            print(f"❌ Test {test.__name__} failed with exception: {e}")
        print()
    
    print("=" * 50)
    print(f"📊 Results: {passed}/{total} tests passed")
    
    if passed == total:
        print("✅ All configuration validation tests passed!")
        return 0
    else:
        print("❌ Some configuration validation tests failed!")
        return 1


if __name__ == "__main__":
    sys.exit(main())
