#!/usr/bin/env python3
"""
Deployment validation script.
Runs during installation to catch configuration and component mismatches early.
"""

import sys
import subprocess
import importlib.util
import os
from pathlib import Path

def test_python_compatibility():
    """Test that python3 is available and working."""
    try:
        result = subprocess.run(['python3', '--version'], 
                              capture_output=True, text=True)
        print(f"OK Python3 available: {result.stdout.strip()}")
        return True
    except FileNotFoundError:
        print("ERROR Python3 not found")
        return False

def test_configuration_loading():
    """Test that configuration can be loaded without errors."""
    try:
        # Add src to path for imports
        src_path = Path(__file__).parent.parent / 'src'
        sys.path.insert(0, str(src_path))
        
        from camera_service.config import ConfigManager
        
        config_manager = ConfigManager()
        config = config_manager.load_config()
        
        # Test that all required fields are present
        assert hasattr(config.mediamtx, 'health_check_interval')
        assert hasattr(config.mediamtx, 'health_failure_threshold')
        assert hasattr(config.mediamtx, 'health_circuit_breaker_timeout')
        assert hasattr(config.mediamtx, 'health_max_backoff_interval')
        assert hasattr(config.mediamtx, 'health_recovery_confirmation_threshold')
        assert hasattr(config.mediamtx, 'backoff_base_multiplier')
        assert hasattr(config.mediamtx, 'backoff_jitter_range')
        assert hasattr(config.mediamtx, 'process_termination_timeout')
        assert hasattr(config.mediamtx, 'process_kill_timeout')
        
        print("OK Configuration loading successful")
        return True
    except Exception as e:
        print(f"ERROR Configuration loading failed: {e}")
        return False

def test_component_instantiation():
    """Test that components can be instantiated with config."""
    try:
        # Add src to path for imports
        src_path = Path(__file__).parent.parent / 'src'
        sys.path.insert(0, str(src_path))
        
        from camera_service.config import ConfigManager
        from camera_service.service_manager import ServiceManager
        
        config_manager = ConfigManager()
        config = config_manager.load_config()
        
        # Test service manager instantiation
        service_manager = ServiceManager(config)
        
        # Test camera monitor instantiation
        from camera_discovery.hybrid_monitor import HybridCameraMonitor
        camera_monitor = HybridCameraMonitor(
            device_range=config.camera.device_range,
            poll_interval=config.camera.poll_interval,
            detection_timeout=config.camera.detection_timeout,
            enable_capability_detection=config.camera.enable_capability_detection,
        )
        
        # Test WebSocket server instantiation
        from websocket_server.server import WebSocketJsonRpcServer
        websocket_server = WebSocketJsonRpcServer(
            host=config.server.host,
            port=config.server.port,
            websocket_path=config.server.websocket_path,
            max_connections=config.server.max_connections,
        )
        
        print("OK Component instantiation successful")
        return True
    except Exception as e:
        print(f"ERROR Component instantiation failed: {e}")
        return False

def test_mediamtx_controller_compatibility():
    """Test that MediaMTXConfig can instantiate MediaMTXController."""
    try:
        # Add src to path for imports
        src_path = Path(__file__).parent.parent / 'src'
        sys.path.insert(0, str(src_path))
        
        from camera_service.config import ConfigManager
        from mediamtx_wrapper.controller import MediaMTXController
        
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
        
        print("OK MediaMTXController instantiation successful")
        return True
    except Exception as e:
        print(f"ERROR MediaMTXController instantiation failed: {e}")
        return False

def test_required_dependencies():
    """Test that all required Python dependencies are available."""
    required_modules = [
        'yaml',
        'asyncio',
        'logging',
        'dataclasses',
        'typing',
        'pathlib',
        'threading',
        'time',
        'os',
        'sys'
    ]
    
    missing_modules = []
    for module in required_modules:
        try:
            importlib.import_module(module)
        except ImportError:
            missing_modules.append(module)
    
    if missing_modules:
        print(f"ERROR Missing required modules: {missing_modules}")
        return False
    
    print("OK All required dependencies available")
    return True

def main():
    """Run all deployment validation tests."""
    print("Running deployment validation...")
    print("=" * 50)
    
    tests = [
        test_python_compatibility,
        test_required_dependencies,
        test_configuration_loading,
        test_component_instantiation,
        test_mediamtx_controller_compatibility,
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
    print(f"Results: {passed}/{total} tests passed")
    
    if passed == total:
        print("OK Deployment validation passed!")
        return 0
    else:
        print("ERROR Deployment validation failed!")
        return 1

if __name__ == "__main__":
    sys.exit(main())
