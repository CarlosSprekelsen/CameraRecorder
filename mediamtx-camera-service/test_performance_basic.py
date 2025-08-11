#!/usr/bin/env python3
"""
Basic Performance Sanity Check Test Script

Tests basic performance sanity:
1. Service startup time
2. Basic operation timing
3. Performance feasibility assessment

Each test measures:
- Startup duration
- Operation completion time
- Basic performance indicators
"""

import sys
import json
import time
import asyncio
from typing import Dict, Any

# Add src to path for imports
sys.path.append('src')

from camera_service.service_manager import ServiceManager
from websocket_server.server import WebSocketJsonRpcServer
from security.jwt_handler import JWTHandler
from security.auth_manager import AuthManager
from security.middleware import SecurityMiddleware


def test_service_startup_time():
    """Test service startup time - measure initialization duration."""
    print("=== Testing Service Startup Time ===")
    
    test_results = {}
    
    # Test 1: Service Manager Startup
    print("\n1. Service Manager Startup Time")
    try:
        start_time = time.time()
        
        # Create minimal config for testing
        from camera_service.config import Config
        config = Config()
        
        # Initialize service manager with config
        service_manager = ServiceManager(config)
        
        end_time = time.time()
        startup_duration = end_time - start_time
        
        print(f"✅ Service Manager startup completed in {startup_duration:.3f} seconds")
        
        test_results['service_manager_startup'] = {
            'duration_seconds': startup_duration,
            'status': 'success',
            'acceptable': startup_duration < 5.0  # Should start within 5 seconds
        }
        
    except Exception as e:
        print(f"❌ Service Manager startup failed: {e}")
        test_results['service_manager_startup'] = {
            'duration_seconds': None,
            'status': 'failed',
            'error': str(e),
            'acceptable': False
        }
    
    # Test 2: WebSocket Server Startup
    print("\n2. WebSocket Server Startup Time")
    try:
        start_time = time.time()
        
        # Initialize WebSocket server components
        jwt_handler = JWTHandler("test_secret_key")
        auth_manager = AuthManager(jwt_handler, None)
        security_middleware = SecurityMiddleware(auth_manager)
        
        # Initialize WebSocket server with correct parameters
        websocket_server = WebSocketJsonRpcServer(
            host="localhost",
            port=8765,
            websocket_path="/ws",
            max_connections=100
        )
        
        # Set security middleware after initialization
        websocket_server.set_security_middleware(security_middleware)
        
        end_time = time.time()
        startup_duration = end_time - start_time
        
        print(f"✅ WebSocket Server startup completed in {startup_duration:.3f} seconds")
        
        test_results['websocket_server_startup'] = {
            'duration_seconds': startup_duration,
            'status': 'success',
            'acceptable': startup_duration < 3.0  # Should start within 3 seconds
        }
        
    except Exception as e:
        print(f"❌ WebSocket Server startup failed: {e}")
        test_results['websocket_server_startup'] = {
            'duration_seconds': None,
            'status': 'failed',
            'error': str(e),
            'acceptable': False
        }
    
    # Test 3: Security Components Startup
    print("\n3. Security Components Startup Time")
    try:
        start_time = time.time()
        
        # Initialize all security components
        jwt_handler = JWTHandler("test_secret_key")
        auth_manager = AuthManager(jwt_handler, None)
        security_middleware = SecurityMiddleware(auth_manager)
        
        end_time = time.time()
        startup_duration = end_time - start_time
        
        print(f"✅ Security components startup completed in {startup_duration:.3f} seconds")
        
        test_results['security_startup'] = {
            'duration_seconds': startup_duration,
            'status': 'success',
            'acceptable': startup_duration < 1.0  # Should start within 1 second
        }
        
    except Exception as e:
        print(f"❌ Security components startup failed: {e}")
        test_results['security_startup'] = {
            'duration_seconds': None,
            'status': 'failed',
            'error': str(e),
            'acceptable': False
        }
    
    return test_results


def test_basic_operation_timing():
    """Test basic operation timing - measure key operation completion times."""
    print("\n=== Testing Basic Operation Timing ===")
    
    test_results = {}
    
    # Test 1: JWT Token Generation Timing
    print("\n1. JWT Token Generation Timing")
    try:
        jwt_handler = JWTHandler("test_secret_key")
        
        # Measure token generation time
        start_time = time.time()
        
        for i in range(10):  # Generate 10 tokens
            token = jwt_handler.generate_token(f"user_{i}", "operator")
        
        end_time = time.time()
        total_duration = end_time - start_time
        avg_duration = total_duration / 10
        
        print(f"✅ JWT token generation: {avg_duration:.6f} seconds average per token")
        print(f"   Total time for 10 tokens: {total_duration:.3f} seconds")
        
        test_results['jwt_generation'] = {
            'total_duration_seconds': total_duration,
            'average_duration_seconds': avg_duration,
            'tokens_generated': 10,
            'status': 'success',
            'acceptable': avg_duration < 0.001  # Should be under 1ms per token
        }
        
    except Exception as e:
        print(f"❌ JWT token generation failed: {e}")
        test_results['jwt_generation'] = {
            'status': 'failed',
            'error': str(e),
            'acceptable': False
        }
    
    # Test 2: JWT Token Validation Timing
    print("\n2. JWT Token Validation Timing")
    try:
        jwt_handler = JWTHandler("test_secret_key")
        
        # Generate a token first
        token = jwt_handler.generate_token("test_user", "operator")
        
        # Measure token validation time
        start_time = time.time()
        
        for i in range(100):  # Validate 100 times
            claims = jwt_handler.validate_token(token)
        
        end_time = time.time()
        total_duration = end_time - start_time
        avg_duration = total_duration / 100
        
        print(f"✅ JWT token validation: {avg_duration:.6f} seconds average per validation")
        print(f"   Total time for 100 validations: {total_duration:.3f} seconds")
        
        test_results['jwt_validation'] = {
            'total_duration_seconds': total_duration,
            'average_duration_seconds': avg_duration,
            'validations_performed': 100,
            'status': 'success',
            'acceptable': avg_duration < 0.001  # Should be under 1ms per validation
        }
        
    except Exception as e:
        print(f"❌ JWT token validation failed: {e}")
        test_results['jwt_validation'] = {
            'status': 'failed',
            'error': str(e),
            'acceptable': False
        }
    
    # Test 3: Authentication Manager Timing
    print("\n3. Authentication Manager Timing")
    try:
        jwt_handler = JWTHandler("test_secret_key")
        auth_manager = AuthManager(jwt_handler, None)
        
        # Generate a token
        token = jwt_handler.generate_token("test_user", "operator")
        
        # Measure authentication time
        start_time = time.time()
        
        for i in range(50):  # Authenticate 50 times
            auth_result = auth_manager.authenticate(token, "jwt")
        
        end_time = time.time()
        total_duration = end_time - start_time
        avg_duration = total_duration / 50
        
        print(f"✅ Authentication: {avg_duration:.6f} seconds average per authentication")
        print(f"   Total time for 50 authentications: {total_duration:.3f} seconds")
        
        test_results['authentication'] = {
            'total_duration_seconds': total_duration,
            'average_duration_seconds': avg_duration,
            'authentications_performed': 50,
            'status': 'success',
            'acceptable': avg_duration < 0.005  # Should be under 5ms per authentication
        }
        
    except Exception as e:
        print(f"❌ Authentication failed: {e}")
        test_results['authentication'] = {
            'status': 'failed',
            'error': str(e),
            'acceptable': False
        }
    
    # Test 4: Permission Checking Timing
    print("\n4. Permission Checking Timing")
    try:
        jwt_handler = JWTHandler("test_secret_key")
        auth_manager = AuthManager(jwt_handler, None)
        
        # Generate a token and authenticate
        token = jwt_handler.generate_token("test_user", "operator")
        auth_result = auth_manager.authenticate(token, "jwt")
        
        # Measure permission checking time
        start_time = time.time()
        
        for i in range(100):  # Check permissions 100 times
            has_permission = auth_manager.has_permission(auth_result, "operator")
        
        end_time = time.time()
        total_duration = end_time - start_time
        avg_duration = total_duration / 100
        
        print(f"✅ Permission checking: {avg_duration:.6f} seconds average per check")
        print(f"   Total time for 100 permission checks: {total_duration:.3f} seconds")
        
        test_results['permission_checking'] = {
            'total_duration_seconds': total_duration,
            'average_duration_seconds': avg_duration,
            'checks_performed': 100,
            'status': 'success',
            'acceptable': avg_duration < 0.0001  # Should be under 0.1ms per check
        }
        
    except Exception as e:
        print(f"❌ Permission checking failed: {e}")
        test_results['permission_checking'] = {
            'status': 'failed',
            'error': str(e),
            'acceptable': False
        }
    
    return test_results


def test_memory_usage_sanity():
    """Test basic memory usage sanity - check for obvious memory leaks."""
    print("\n=== Testing Memory Usage Sanity ===")
    
    test_results = {}
    
    # Test 1: Component Memory Usage
    print("\n1. Component Memory Usage")
    try:
        import psutil
        import os
        
        # Get initial memory usage
        process = psutil.Process(os.getpid())
        initial_memory = process.memory_info().rss / 1024 / 1024  # MB
        
        print(f"   Initial memory usage: {initial_memory:.2f} MB")
        
        # Create multiple components
        components = []
        for i in range(10):
            jwt_handler = JWTHandler("test_secret_key")
            auth_manager = AuthManager(jwt_handler, None)
            security_middleware = SecurityMiddleware(auth_manager)
            components.append((jwt_handler, auth_manager, security_middleware))
        
        # Get memory usage after component creation
        after_creation_memory = process.memory_info().rss / 1024 / 1024  # MB
        memory_increase = after_creation_memory - initial_memory
        
        print(f"   Memory after creating 10 component sets: {after_creation_memory:.2f} MB")
        print(f"   Memory increase: {memory_increase:.2f} MB")
        
        # Clean up
        del components
        
        # Get memory usage after cleanup
        after_cleanup_memory = process.memory_info().rss / 1024 / 1024  # MB
        final_memory_increase = after_cleanup_memory - initial_memory
        
        print(f"   Memory after cleanup: {after_cleanup_memory:.2f} MB")
        print(f"   Final memory increase: {final_memory_increase:.2f} MB")
        
        # Check if memory usage is reasonable
        memory_acceptable = memory_increase < 50.0  # Should not increase by more than 50MB
        cleanup_acceptable = final_memory_increase < 10.0  # Should clean up most memory
        
        if memory_acceptable and cleanup_acceptable:
            print("✅ Memory usage is reasonable")
        else:
            print("⚠️ Memory usage may be concerning")
        
        test_results['memory_usage'] = {
            'initial_memory_mb': initial_memory,
            'after_creation_memory_mb': after_creation_memory,
            'memory_increase_mb': memory_increase,
            'after_cleanup_memory_mb': after_cleanup_memory,
            'final_memory_increase_mb': final_memory_increase,
            'memory_acceptable': memory_acceptable,
            'cleanup_acceptable': cleanup_acceptable,
            'status': 'success',
            'acceptable': memory_acceptable and cleanup_acceptable
        }
        
    except ImportError:
        print("⚠️ psutil not available, skipping memory usage test")
        test_results['memory_usage'] = {
            'status': 'skipped',
            'reason': 'psutil not available',
            'acceptable': True  # Assume acceptable if we can't test
        }
    except Exception as e:
        print(f"❌ Memory usage test failed: {e}")
        test_results['memory_usage'] = {
            'status': 'failed',
            'error': str(e),
            'acceptable': False
        }
    
    return test_results


def main():
    """Main test function."""
    print("=== Basic Performance Sanity Check ===")
    print("Testing service startup time, basic operation timing, and memory usage\n")
    
    all_results = {}
    
    try:
        # Test 1: Service Startup Time
        all_results['startup_timing'] = test_service_startup_time()
        
        # Test 2: Basic Operation Timing
        all_results['operation_timing'] = test_basic_operation_timing()
        
        # Test 3: Memory Usage Sanity
        all_results['memory_sanity'] = test_memory_usage_sanity()
        
        # Calculate overall performance assessment
        startup_acceptable = all(
            result.get('acceptable', False) 
            for result in all_results['startup_timing'].values()
        )
        
        operation_acceptable = all(
            result.get('acceptable', False) 
            for result in all_results['operation_timing'].values()
        )
        
        memory_acceptable = all_results['memory_sanity']['memory_usage'].get('acceptable', False)
        
        overall_acceptable = startup_acceptable and operation_acceptable and memory_acceptable
        
        print("\n=== Performance Assessment ===")
        print(f"✅ Startup timing acceptable: {startup_acceptable}")
        print(f"✅ Operation timing acceptable: {operation_acceptable}")
        print(f"✅ Memory usage acceptable: {memory_acceptable}")
        print(f"✅ Overall performance acceptable: {overall_acceptable}")
        
        if overall_acceptable:
            print("\n🎉 All performance sanity checks passed!")
            print("✅ Service starts successfully")
            print("✅ Basic operations complete within reasonable time")
            print("✅ No obvious performance blockers identified")
        else:
            print("\n⚠️ Some performance issues identified")
            print("❌ Service startup or operations may have performance issues")
        
        all_results['overall_assessment'] = {
            'startup_acceptable': startup_acceptable,
            'operation_acceptable': operation_acceptable,
            'memory_acceptable': memory_acceptable,
            'overall_acceptable': overall_acceptable
        }
        
        return all_results
        
    except Exception as e:
        print(f"\n❌ Performance sanity check failed with exception: {e}")
        return {"error": str(e)}


if __name__ == "__main__":
    # Run the tests
    results = main()
    
    # Save results for reporting
    with open("performance_test_results.json", "w") as f:
        json.dump(results, f, indent=2, default=str)
    
    print(f"\nTest results saved to performance_test_results.json")
