# Bug Report: Missing get_test_auth_manager function in auth_utils

**Issue ID:** 027  
**Severity:** HIGH  
**Status:** OPEN  
**Date:** 2025-01-06  
**Reporter:** IV&V Team  

## Summary
The `tests/fixtures/auth_utils.py` file is missing the `get_test_auth_manager` function that is imported by multiple integration tests, causing import errors during test collection.

## Description
When running the test suite, the following import errors occur:

```
ERROR collecting tests/integration/test_critical_interfaces.py
ImportError while importing test module '/home/carlossprekelsen/CameraRecorder/mediamtx-camera-service/tests/integration/test_critical_interfaces.py'.
Hint: make sure your test modules/packages have valid Python names.
Traceback:
tests/integration/test_critical_interfaces.py:60: in <module>
    from tests.fixtures.auth_utils import get_test_auth_manager, TestUserFactory, WebSocketAuthTestClient, cleanup_test_auth_manager
E   ImportError: cannot import name 'get_test_auth_manager' from 'tests.fixtures.auth_utils'
```

Similar errors occur in:
- `tests/integration/test_security_authentication.py`

## Root Cause
The `tests/fixtures/auth_utils.py` file does not define the `get_test_auth_manager` function that is expected by integration tests.

## Impact
- Test collection fails for critical integration tests
- Prevents execution of authentication and security integration tests
- Blocks comprehensive test coverage validation
- Violates testing guide requirements for real authentication testing

## Steps to Reproduce
1. Run the test suite: `python3 -m pytest tests/ -v`
2. Observe import errors for integration test files

## Expected Behavior
Integration tests should be able to import and use the `get_test_auth_manager` function.

## Actual Behavior
Import errors prevent test collection for integration tests.

## Affected Files
- `tests/fixtures/auth_utils.py`
- `tests/integration/test_critical_interfaces.py`
- `tests/integration/test_security_authentication.py`

## Requirements Impact
- **REQ-SEC-001**: JWT token-based authentication for all API access
- **REQ-SEC-002**: Token Format: JSON Web Token (JWT) with standard claims
- **REQ-SEC-003**: Token Expiration: Configurable expiration time (default: 24 hours)
- **REQ-SEC-004**: Token Refresh: Support for token refresh mechanism
- **REQ-SEC-005**: Token Validation: Proper signature validation and claim verification

## Fix Required
Add the missing `get_test_auth_manager` function to `tests/fixtures/auth_utils.py`:

```python
def get_test_auth_manager():
    """Get a test authentication manager instance."""
    # Implementation needed
    pass
```

## Testing Guide Compliance
- **Violation**: Missing required test utility function
- **Impact**: Prevents execution of integration requirement validation tests
- **Priority**: HIGH - Integration testing is critical for system validation
