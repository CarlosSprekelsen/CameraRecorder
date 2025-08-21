# Bug Report: AuthManager auth_method not set for failed authentication

**Issue ID:** 030  
**Severity:** MEDIUM  
**Status:** OPEN  
**Date:** 2025-01-06  
**Reporter:** IV&V Team  

## Summary
The `AuthManager` class does not set the `auth_method` field in `AuthResult` when authentication fails, causing test assertions to fail. Tests expect the auth_method to be set even for failed authentication attempts.

## Description
When running authentication tests, the following failures occur:

```
FAILED tests/unit/test_security/test_auth_manager.py::TestAuthManager::test_authenticate_jwt_invalid_token
AssertionError: assert None == 'jwt'
 +  where None = AuthResult(authenticated=False, user_id=None, role=None, auth_method=None, expires_at=None, error_message='Invalid or expired JWT token').auth_method

FAILED tests/unit/test_security/test_auth_manager.py::TestAuthManager::test_authenticate_api_key_invalid
AssertionError: assert None == 'api_key'
 +  where None = AuthResult(authenticated=False, user_id=None, role=None, auth_method=None, expires_at=None, error_message='Invalid or expired API key').auth_method

FAILED tests/unit/test_security/test_auth_manager.py::TestAuthManager::test_authenticate_auto_both_fail
AssertionError: assert None == 'jwt'
 +  where None = AuthResult(authenticated=False, user_id=None, role=None, auth_method=None, expires_at=None, error_message='Invalid or expired JWT token').auth_method
```

## Root Cause
The `AuthManager.authenticate()` method does not set the `auth_method` field in the `AuthResult` when authentication fails, even though it attempted to authenticate using a specific method.

## Impact
- Authentication test failures
- Inconsistent behavior between successful and failed authentication
- Difficulty in debugging authentication issues
- Violates testing guide requirements for authentication validation

## Steps to Reproduce
1. Run authentication tests: `python3 -m pytest tests/unit/test_security/test_auth_manager.py -v`
2. Observe auth_method assertion failures

## Expected Behavior
The `auth_method` field should be set to the method that was attempted (e.g., 'jwt', 'api_key') even when authentication fails.

## Actual Behavior
The `auth_method` field is `None` for failed authentication attempts.

## Affected Files
- `tests/unit/test_security/test_auth_manager.py`
- `src/security/auth_manager.py`

## Requirements Impact
- **REQ-SEC-001**: JWT token-based authentication for all API access
- **REQ-SEC-002**: Token Format: JSON Web Token (JWT) with standard claims
- **REQ-SEC-003**: Token Expiration: Configurable expiration time (default: 24 hours)
- **REQ-SEC-004**: Token Refresh: Support for token refresh mechanism
- **REQ-SEC-005**: Token Validation: Proper signature validation and claim verification

## Fix Required
Update the `AuthManager.authenticate()` method to set the `auth_method` field in `AuthResult` even when authentication fails:

```python
# For JWT authentication failure
return AuthResult(
    authenticated=False,
    user_id=None,
    role=None,
    auth_method="jwt",  # Set the attempted method
    expires_at=None,
    error_message="Invalid or expired JWT token"
)

# For API key authentication failure
return AuthResult(
    authenticated=False,
    user_id=None,
    role=None,
    auth_method="api_key",  # Set the attempted method
    expires_at=None,
    error_message="Invalid or expired API key"
)
```

## Testing Guide Compliance
- **Violation**: Authentication behavior is inconsistent between success and failure cases
- **Impact**: Prevents proper authentication validation testing
- **Priority**: MEDIUM - Authentication consistency is important for security validation
