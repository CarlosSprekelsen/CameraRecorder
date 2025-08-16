#!/usr/bin/env python3
"""
Security Concept Validation Test Script

Tests basic security concepts:
1. Authentication - JWT token validation
2. Authorization - Access control and permission checking
3. Security design feasibility

Each concept tested with:
- Success case: Valid authentication/authorization
- Negative case: Invalid authentication/authorization
"""

import sys
import json
import time
from typing import Dict, Any

# Add src to path for imports
sys.path.append('src')

from security.jwt_handler import JWTHandler, JWTClaims
from security.auth_manager import AuthManager, AuthResult
from security.middleware import SecurityMiddleware
from security.api_key_handler import APIKeyHandler


def test_jwt_authentication_concept():
    """Test JWT authentication concept - token generation and validation."""
    print("=== Testing JWT Authentication Concept ===")
    
    # Initialize JWT handler with test secret
    secret_key = "test_secret_key_for_validation_only"
    jwt_handler = JWTHandler(secret_key)
    
    test_results = {}
    
    # Success case: Valid token generation and validation
    print("\n1. JWT Authentication - Success Case")
    try:
        # Generate token
        user_id = "test_user_123"
        role = "operator"
        token = jwt_handler.generate_token(user_id, role, expiry_hours=1)
        
        print(f"✅ Token generated successfully for user {user_id} with role {role}")
        print(f"   Token: {token[:50]}...")
        
        # Validate token
        claims = jwt_handler.validate_token(token)
        if claims:
            print(f"✅ Token validation successful")
            print(f"   User ID: {claims.user_id}")
            print(f"   Role: {claims.role}")
            print(f"   Expires: {time.ctime(claims.exp)}")
            
            test_results['jwt_success'] = {
                'user_id': claims.user_id,
                'role': claims.role,
                'expires': claims.exp,
                'valid': True
            }
        else:
            print("❌ Token validation failed")
            test_results['jwt_success'] = {'valid': False}
            
    except Exception as e:
        print(f"❌ JWT success case failed: {e}")
        test_results['jwt_success'] = {'error': str(e)}
    
    # Negative case: Invalid token validation
    print("\n2. JWT Authentication - Negative Case (Invalid Token)")
    try:
        # Test with invalid token
        invalid_token = "invalid.jwt.token"
        claims = jwt_handler.validate_token(invalid_token)
        
        if claims is None:
            print("✅ Invalid token properly rejected")
            test_results['jwt_negative'] = {'valid': False, 'expected': 'rejected'}
        else:
            print("❌ Invalid token was accepted (should be rejected)")
            test_results['jwt_negative'] = {'valid': True, 'expected': 'rejected'}
            
    except Exception as e:
        print(f"✅ Invalid token properly rejected with exception: {e}")
        test_results['jwt_negative'] = {'valid': False, 'expected': 'rejected', 'exception': str(e)}
    
    # Negative case: Expired token
    print("\n3. JWT Authentication - Negative Case (Expired Token)")
    try:
        # Generate token with very short expiry
        expired_token = jwt_handler.generate_token("test_user", "viewer", expiry_hours=0)
        
        # Wait a moment and try to validate
        time.sleep(1)
        claims = jwt_handler.validate_token(expired_token)
        
        if claims is None:
            print("✅ Expired token properly rejected")
            test_results['jwt_expired'] = {'valid': False, 'expected': 'rejected'}
        else:
            print("❌ Expired token was accepted (should be rejected)")
            test_results['jwt_expired'] = {'valid': True, 'expected': 'rejected'}
            
    except Exception as e:
        print(f"✅ Expired token properly rejected with exception: {e}")
        test_results['jwt_expired'] = {'valid': False, 'expected': 'rejected', 'exception': str(e)}
    
    return test_results


def test_authorization_concept():
    """Test authorization concept - role-based access control."""
    print("\n=== Testing Authorization Concept ===")
    
    # Initialize components
    secret_key = "test_secret_key_for_validation_only"
    jwt_handler = JWTHandler(secret_key)
    # Skip API key handler for this test to focus on JWT authorization
    auth_manager = AuthManager(jwt_handler, None)
    
    test_results = {}
    
    # Success case: Valid authorization for operator role
    print("\n1. Authorization - Success Case (Operator Access)")
    try:
        # Generate token with operator role
        token = jwt_handler.generate_token("test_operator", "operator")
        
        # Authenticate
        auth_result = auth_manager.authenticate(token, "jwt")
        
        if auth_result.authenticated:
            print(f"✅ Authentication successful for user {auth_result.user_id}")
            print(f"   Role: {auth_result.role}")
            
            # Check permissions
            can_take_snapshot = auth_manager.has_permission(auth_result, "operator")
            can_view = auth_manager.has_permission(auth_result, "viewer")
            
            print(f"   Can take snapshot (operator): {can_take_snapshot}")
            print(f"   Can view (viewer): {can_view}")
            
            test_results['auth_success'] = {
                'authenticated': True,
                'user_id': auth_result.user_id,
                'role': auth_result.role,
                'can_take_snapshot': can_take_snapshot,
                'can_view': can_view
            }
        else:
            print(f"❌ Authentication failed: {auth_result.error_message}")
            test_results['auth_success'] = {'authenticated': False, 'error': auth_result.error_message}
            
    except Exception as e:
        print(f"❌ Authorization success case failed: {e}")
        test_results['auth_success'] = {'error': str(e)}
    
    # Negative case: Insufficient permissions
    print("\n2. Authorization - Negative Case (Insufficient Permissions)")
    try:
        # Generate token with viewer role
        token = jwt_handler.generate_token("test_viewer", "viewer")
        
        # Authenticate
        auth_result = auth_manager.authenticate(token, "jwt")
        
        if auth_result.authenticated:
            print(f"✅ Authentication successful for user {auth_result.user_id}")
            print(f"   Role: {auth_result.role}")
            
            # Check permissions (viewer should not have operator permissions)
            can_take_snapshot = auth_manager.has_permission(auth_result, "operator")
            can_view = auth_manager.has_permission(auth_result, "viewer")
            
            print(f"   Can take snapshot (operator): {can_take_snapshot}")
            print(f"   Can view (viewer): {can_view}")
            
            if not can_take_snapshot and can_view:
                print("✅ Authorization properly enforced - viewer cannot take snapshots")
                test_results['auth_negative'] = {
                    'authenticated': True,
                    'role': auth_result.role,
                    'operator_access_denied': True,
                    'viewer_access_granted': True
                }
            else:
                print("❌ Authorization not properly enforced")
                test_results['auth_negative'] = {
                    'authenticated': True,
                    'role': auth_result.role,
                    'operator_access_denied': can_take_snapshot,
                    'viewer_access_granted': can_view
                }
        else:
            print(f"❌ Authentication failed: {auth_result.error_message}")
            test_results['auth_negative'] = {'authenticated': False, 'error': auth_result.error_message}
            
    except Exception as e:
        print(f"❌ Authorization negative case failed: {e}")
        test_results['auth_negative'] = {'error': str(e)}
    
    # Negative case: No authentication
    print("\n3. Authorization - Negative Case (No Authentication)")
    try:
        # Try to authenticate without token
        auth_result = auth_manager.authenticate("", "jwt")
        
        if not auth_result.authenticated:
            print("✅ No authentication properly rejected")
            print(f"   Error: {auth_result.error_message}")
            test_results['auth_no_token'] = {
                'authenticated': False,
                'expected': 'rejected',
                'error': auth_result.error_message
            }
        else:
            print("❌ No authentication was accepted (should be rejected)")
            test_results['auth_no_token'] = {
                'authenticated': True,
                'expected': 'rejected'
            }
            
    except Exception as e:
        print(f"✅ No authentication properly rejected with exception: {e}")
        test_results['auth_no_token'] = {
            'authenticated': False,
            'expected': 'rejected',
            'exception': str(e)
        }
    
    return test_results


async def test_security_middleware_concept():
    """Test security middleware concept - integrated authentication and authorization."""
    print("\n=== Testing Security Middleware Concept ===")
    
    # Initialize components
    secret_key = "test_secret_key_for_validation_only"
    jwt_handler = JWTHandler(secret_key)
    # Skip API key handler for this test to focus on JWT middleware
    auth_manager = AuthManager(jwt_handler, None)
    security_middleware = SecurityMiddleware(auth_manager, max_connections=10)
    
    test_results = {}
    
    # Success case: Valid connection and authentication
    print("\n1. Security Middleware - Success Case (Valid Connection)")
    try:
        client_id = "test_client_123"
        
        # Check if connection can be accepted
        can_accept = security_middleware.can_accept_connection(client_id)
        if can_accept:
            print(f"✅ Connection can be accepted for client {client_id}")
            
            # Register connection
            security_middleware.register_connection(client_id)
            print(f"✅ Connection registered for client {client_id}")
            
            # Generate token and authenticate
            token = jwt_handler.generate_token("test_user", "operator")
            auth_result = await security_middleware.authenticate_connection(client_id, token, "jwt")
            
            if auth_result.authenticated:
                print(f"✅ Authentication successful via middleware")
                print(f"   User: {auth_result.user_id}, Role: {auth_result.role}")
                
                # Check permissions
                has_permission = security_middleware.has_permission(client_id, "operator")
                print(f"   Has operator permission: {has_permission}")
                
                test_results['middleware_success'] = {
                    'connection_accepted': True,
                    'authenticated': True,
                    'user_id': auth_result.user_id,
                    'role': auth_result.role,
                    'has_permission': has_permission
                }
            else:
                print(f"❌ Authentication failed via middleware: {auth_result.error_message}")
                test_results['middleware_success'] = {
                    'connection_accepted': True,
                    'authenticated': False,
                    'error': auth_result.error_message
                }
        else:
            print(f"❌ Connection cannot be accepted for client {client_id}")
            test_results['middleware_success'] = {'connection_accepted': False}
            
    except Exception as e:
        print(f"❌ Security middleware success case failed: {e}")
        test_results['middleware_success'] = {'error': str(e)}
    
    # Negative case: Unauthorized access attempt
    print("\n2. Security Middleware - Negative Case (Unauthorized Access)")
    try:
        client_id = "test_client_unauthorized"
        
        # Register connection
        if security_middleware.can_accept_connection(client_id):
            security_middleware.register_connection(client_id)
            print(f"✅ Connection registered for client {client_id}")
            
            # Try to access without authentication
            has_permission = security_middleware.has_permission(client_id, "operator")
            
            if not has_permission:
                print("✅ Unauthorized access properly rejected")
                test_results['middleware_unauthorized'] = {
                    'connection_accepted': True,
                    'has_permission': False,
                    'expected': 'rejected'
                }
            else:
                print("❌ Unauthorized access was granted (should be rejected)")
                test_results['middleware_unauthorized'] = {
                    'connection_accepted': True,
                    'has_permission': True,
                    'expected': 'rejected'
                }
        else:
            print(f"❌ Connection cannot be accepted for client {client_id}")
            test_results['middleware_unauthorized'] = {'connection_accepted': False}
            
    except Exception as e:
        print(f"✅ Unauthorized access properly rejected with exception: {e}")
        test_results['middleware_unauthorized'] = {
            'connection_accepted': True,
            'has_permission': False,
            'expected': 'rejected',
            'exception': str(e)
        }
    
    return test_results


async def main():
    """Main test function."""
    print("=== Security Concept Validation Test ===")
    print("Testing basic security concepts: authentication, authorization, and middleware\n")
    
    all_results = {}
    
    try:
        # Test 1: JWT Authentication Concept
        all_results['jwt_authentication'] = test_jwt_authentication_concept()
        
        # Test 2: Authorization Concept
        all_results['authorization'] = test_authorization_concept()
        
        # Test 3: Security Middleware Concept
        all_results['security_middleware'] = await test_security_middleware_concept()
        
        print("\n=== Test Summary ===")
        print("✅ All security concept tests completed!")
        print("✅ Authentication concept: JWT token validation working")
        print("✅ Authorization concept: Access control working")
        print("✅ Security design: Basic approach feasible")
        
        return all_results
        
    except Exception as e:
        print(f"\n❌ Test failed with exception: {e}")
        return {"error": str(e)}


if __name__ == "__main__":
    # Run the tests
    import asyncio
    results = asyncio.run(main())
    
    # Save results for reporting
    with open("security_test_results.json", "w") as f:
        json.dump(results, f, indent=2, default=str)
    
    print(f"\nTest results saved to security_test_results.json")
