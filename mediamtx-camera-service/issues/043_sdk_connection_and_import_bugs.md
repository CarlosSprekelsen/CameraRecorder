# Issue 043: SDK Connection and Import Bugs

**Date:** 2025-01-27  
**Severity:** Medium  
**Category:** SDK/Client Development  
**Status:** Identified  

## Summary

Multiple SDK-related bugs are causing test failures, including connection errors, import issues, and missing JavaScript client files. These bugs are affecting SDK integration tests and client functionality.

## Bugs Identified

### 1. SDK Connection Errors (MEDIUM)

**Error Pattern:**
```
mediamtx_camera_sdk.exceptions.ConnectionError: Failed to connect after 1 attempt(s)
```

**Affected Tests:**
- `test_invalid_jwt_token_raises_authentication_error`
- `test_invalid_api_key_raises_authentication_error`
- `test_no_auth_token_raises_authentication_error`
- `test_empty_auth_token_raises_authentication_error`
- `test_malformed_jwt_token_raises_authentication_error`
- `test_authentication_error_message_contains_details`
- `test_authentication_error_type_is_correct`
- `test_authentication_error_does_not_leave_connection_open`
- `test_multiple_authentication_attempts_consistent_behavior`
- `test_authentication_error_handling_performance`
- `test_get_camera_list_response_format_handling`

**Root Cause:** The SDK is unable to establish connections to the camera service, likely due to service not running or port binding issues.

### 2. SDK Import Error (MEDIUM)

**Error Pattern:**
```
ImportError
```

**Affected Tests:**
- `test_get_version_without_package`

**Root Cause:** The version detection is failing when the package is not properly installed or available.

### 3. SDK Response Format Issues (MEDIUM)

**Error Pattern:**
```
NameError: name 'camera_data' is not defined
```

**Affected Tests:**
- `test_get_streams_stream_data_structure`

**Root Cause:** The test is referencing an undefined variable `camera_data` in the response format validation.

### 4. SDK Authentication Error Type Mismatch (MEDIUM)

**Error Pattern:**
```
Failed: Expected AuthenticationError but got ConnectionError: Failed to connect after 1 attempt(s)
```

**Affected Tests:**
- `test_authentication_error_type_is_correct`

**Root Cause:** The SDK is throwing `ConnectionError` instead of the expected `AuthenticationError` when authentication fails.

## Impact Assessment

### Medium Impact
- **SDK Functionality**: SDK clients cannot connect to the service
- **Client Development**: SDK integration tests are failing
- **User Experience**: SDK users cannot authenticate or connect
- **Test Suite**: 11+ SDK-related tests are failing

### Low Impact
- **Development Workflow**: SDK testing is blocked
- **Documentation**: SDK examples may not work as expected

## Files Affected

- `sdk/python/mediamtx_camera_sdk/` - Python SDK implementation
- `tests/integration/test_sdk_authentication_error_handling.py` - SDK authentication tests
- `tests/integration/test_sdk_response_format.py` - SDK response format tests
- `src/camera_service/main.py` - Version detection

## Recommended Actions

### Immediate (High Priority)
1. **Fix Connection Issues**: Ensure the camera service is running during SDK tests
2. **Add Connection Retry Logic**: Implement proper retry logic in SDK connection handling
3. **Fix Import Issues**: Resolve version detection import problems

### High Priority
4. **Standardize Error Types**: Ensure authentication errors throw correct exception types
5. **Fix Response Format**: Correct undefined variable references in tests
6. **Add Connection Validation**: Add proper connection validation in SDK

### Medium Priority
7. **Add SDK Logging**: Add detailed logging for SDK connection attempts
8. **Add SDK Documentation**: Document expected connection behavior
9. **Add SDK Testing**: Add comprehensive SDK connection testing

## Test Evidence

The SDK bugs are consistently appearing in test logs:

```bash
# Connection Errors
FAILED tests/integration/test_sdk_authentication_error_handling.py::TestSDKAuthenticationErrorHandling::test_invalid_jwt_token_raises_authentication_error - mediamtx_camera_sdk.exceptions.ConnectionError: Failed to connect after 1 attempt(s)

# Import Errors
FAILED tests/unit/test_main_startup.py::TestGetVersion::test_get_version_without_package - ImportError

# Response Format Issues
FAILED tests/integration/test_sdk_response_format.py::TestSDKResponseFormat::test_get_streams_stream_data_structure - NameError: name 'camera_data' is not defined

# Error Type Mismatch
FAILED tests/integration/test_sdk_authentication_error_handling.py::TestSDKAuthenticationErrorHandling::test_authentication_error_type_is_correct - Failed: Expected AuthenticationError but got ConnectionError
```

## Root Cause Analysis

### Connection Issues
The SDK connection errors suggest that:
1. The camera service is not running during SDK tests
2. There are port binding conflicts preventing service startup
3. The SDK connection timeout is too short
4. The SDK is not properly handling connection failures

### Import Issues
The import errors suggest that:
1. The version detection logic is not properly handling missing packages
2. The mock setup for version detection is incomplete
3. There are dependency issues with the version detection library

## Conclusion

This is a **medium-priority SDK bug** that affects client connectivity and SDK functionality. The issues range from connection problems to import errors, indicating problems in both the SDK implementation and the test environment setup. These need attention to ensure SDK clients can properly connect to and interact with the camera service.
