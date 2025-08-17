#!/usr/bin/env python3
"""
Health Server Diagnostic Script
Tests and debugs health server implementation issues.
"""

import sys
import json
from pathlib import Path

# Add src to path
src_path = Path(__file__).parent / 'src'
sys.path.insert(0, str(src_path))

def test_health_component_serialization():
    """Test HealthComponent JSON serialization."""
    print("=== Testing HealthComponent Serialization ===")
    
    try:
        from health_server import HealthComponent
        
        component = HealthComponent(
            status='healthy',
            details='Test component',
            timestamp='2025-08-17T21:00:00Z'
        )
        
        print(f"HealthComponent created: {component}")
        print(f"Component.__dict__: {component.__dict__}")
        
        # Test JSON serialization
        json_str = json.dumps(component.__dict__)
        print(f"✅ HealthComponent JSON serialization successful: {json_str}")
        return True
        
    except Exception as e:
        print(f"❌ HealthComponent serialization failed: {e}")
        return False

def test_health_response_serialization():
    """Test HealthResponse JSON serialization."""
    print("\n=== Testing HealthResponse Serialization ===")
    
    try:
        from health_server import HealthComponent, HealthResponse
        
        component = HealthComponent(
            status='healthy',
            details='Test component',
            timestamp='2025-08-17T21:00:00Z'
        )
        
        response = HealthResponse(
            status='healthy',
            timestamp='2025-08-17T21:00:00Z',
            components={'test': component}
        )
        
        print(f"HealthResponse created: {response}")
        print(f"Response.__dict__: {response.__dict__}")
        
        # Test JSON serialization
        json_str = json.dumps(response.__dict__)
        print(f"✅ HealthResponse JSON serialization successful: {json_str}")
        return True
        
    except Exception as e:
        print(f"❌ HealthResponse serialization failed: {e}")
        return False

def test_camera_monitor_interface():
    """Test camera monitor interface."""
    print("\n=== Testing Camera Monitor Interface ===")
    
    try:
        # Import service manager to get camera monitor
        from camera_service.service_manager import ServiceManager
        from camera_service.config import Config
        
        # Create config and service manager
        config = Config()
        service_manager = ServiceManager(config)
        
        # Check camera monitor attributes
        if hasattr(service_manager, '_camera_monitor'):
            camera_monitor = service_manager._camera_monitor
            print(f"Camera monitor: {camera_monitor}")
            
            # Check is_running attribute
            if hasattr(camera_monitor, 'is_running'):
                is_running_attr = getattr(camera_monitor, 'is_running')
                print(f"is_running attribute: {is_running_attr}")
                print(f"is_running type: {type(is_running_attr)}")
                
                if callable(is_running_attr):
                    print("✅ is_running is callable")
                else:
                    print("❌ is_running is not callable (should be a method)")
                    
            else:
                print("❌ Camera monitor has no is_running attribute")
        else:
            print("❌ Service manager has no camera monitor")
            
        return True
        
    except Exception as e:
        print(f"❌ Camera monitor interface test failed: {e}")
        return False

def test_service_manager_interface():
    """Test service manager interface."""
    print("\n=== Testing Service Manager Interface ===")
    
    try:
        from camera_service.service_manager import ServiceManager
        from camera_service.config import Config
        
        config = Config()
        service_manager = ServiceManager(config)
        
        # Check is_running attribute
        if hasattr(service_manager, 'is_running'):
            is_running_attr = getattr(service_manager, 'is_running')
            print(f"is_running attribute: {is_running_attr}")
            print(f"is_running type: {type(is_running_attr)}")
            
            if callable(is_running_attr):
                print("✅ is_running is callable")
            else:
                print("❌ is_running is not callable (should be a method)")
        else:
            print("❌ Service manager has no is_running attribute")
            
        return True
        
    except Exception as e:
        print(f"❌ Service manager interface test failed: {e}")
        return False

def test_health_endpoints():
    """Test health endpoints."""
    print("\n=== Testing Health Endpoints ===")
    
    import requests
    
    endpoints = [
        "http://localhost:8003/health/ready",
        "http://localhost:8003/health/system", 
        "http://localhost:8003/health/mediamtx",
        "http://localhost:8003/health/cameras"
    ]
    
    for endpoint in endpoints:
        try:
            response = requests.get(endpoint, timeout=5)
            print(f"{endpoint}: {response.status_code}")
            if response.status_code == 200:
                print(f"  ✅ Response: {response.text[:100]}...")
            else:
                print(f"  ❌ Error: {response.text}")
        except Exception as e:
            print(f"{endpoint}: ❌ Failed - {e}")

def main():
    """Run all diagnostic tests."""
    print("Health Server Diagnostic Tests")
    print("=" * 50)
    
    results = []
    
    # Run tests
    results.append(test_health_component_serialization())
    results.append(test_health_response_serialization())
    results.append(test_camera_monitor_interface())
    results.append(test_service_manager_interface())
    
    # Test endpoints if service is running
    try:
        test_health_endpoints()
    except Exception as e:
        print(f"❌ Endpoint tests failed: {e}")
    
    # Summary
    print("\n" + "=" * 50)
    print("DIAGNOSTIC SUMMARY")
    print("=" * 50)
    
    passed = sum(results)
    total = len(results)
    
    print(f"Tests passed: {passed}/{total}")
    
    if passed == total:
        print("✅ All diagnostic tests passed")
    else:
        print("❌ Some diagnostic tests failed")
        print("\nRECOMMENDED FIXES:")
        print("1. Fix HealthComponent JSON serialization")
        print("2. Fix camera monitor is_running method call")
        print("3. Fix service manager is_running method call")

if __name__ == "__main__":
    main()
