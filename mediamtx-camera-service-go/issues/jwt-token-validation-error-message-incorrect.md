# JWT Token Validation Error Message Incorrect

## Issue Description
The JWT handler is returning a generic "key is invalid" error message instead of the specific "token validation failed" message expected by the API specification when validating tokens with wrong signing methods.

## Current Behavior
- JWT validation returns "key is invalid" for wrong signing methods
- This violates the API specification which expects "token validation failed"

## Expected Behavior
- JWT validation should return "token validation failed" for wrong signing methods
- Error messages should match the API specification in `docs/api/json_rpc_methods.md`

## Evidence
From test output:
```
=== FAIL: TestJWTHandler_ValidateToken_EdgeCases/validate_token_wrong_signing_method (0.00s)
    test_security_jwt_handler_test.go:548: 
                Error:          Received unexpected error:
                                key is invalid
```

## Impact
- API compliance violation
- Inconsistent error messages
- Tests failing due to incorrect error handling

## Priority
High - This affects API compliance and user experience

## Category
Implementation Bug

## Files Affected
- `internal/security/jwt_handler.go` - JWT validation error handling

## Required Action
1. Update JWT validation to return "token validation failed" for wrong signing methods
2. Ensure all JWT error messages match API specification
3. Update tests to expect correct error messages

## Notes
- This was previously identified as a real bug
- Tests are correctly identifying the implementation issue
- API specification is the ground truth for expected behavior

