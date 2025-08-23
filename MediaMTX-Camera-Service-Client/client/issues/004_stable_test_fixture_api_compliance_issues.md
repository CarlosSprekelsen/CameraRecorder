# Issue 004: Stable Test Fixture API Compliance Issues

**Status:** OPEN  
**Priority:** High  
**Type:** Test Infrastructure Issue  
**Created:** 2025-01-23  
**Discovered By:** API Compliance Validation  
**Category:** Test Suite Non-Compliance  

## Description

The stable test fixture in `tests/fixtures/stable-test-fixture.ts` contains multiple API compliance issues that violate the ground truth validation rules and do not properly validate against the frozen API documentation.

## Ground Truth Reference

**Source:** `mediamtx-camera-service/docs/api/json-rpc-methods.md` (FROZEN)  
**Requirement:** All test fixtures must validate against documented API formats

## Current Implementation Issues

**File:** `tests/fixtures/stable-test-fixture.ts`  

### 1. Authentication Parameter Format Error
**Lines:** 128  
**Current (Incorrect):**
```typescript
this.sendRequest(this.ws, 'authenticate', id, { token });
```

**API Documentation Ground Truth:**
```json
{
  "jsonrpc": "2.0",
  "method": "authenticate",
  "params": {
    "auth_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  },
  "id": 0
}
```

**Issue:** Uses `token` instead of `auth_token` parameter name

### 2. Missing API Compliance Validation
**Lines:** 140-160  
**Current Issues:**
- No validation of response format against API documentation
- No ground truth references in test documentation
- No API compliance validation rules
- Tests may adapt to implementation flaws instead of validating against documentation

### 3. Response Handling Not Validating Against API Documentation
**Lines:** 150-170  
**Current Issues:**
- Returns `data.result` directly without validating response structure
- No validation of required fields per API documentation
- No error response format validation
- No JSON-RPC response format validation

### 4. Multiple Method Calls Without API Compliance
**Lines:** 347-1051  
**Current Issues:**
- `get_camera_list` called without authentication validation
- `take_snapshot` calls don't validate against documented parameter format
- `start_recording` calls don't validate against documented parameter format
- `list_recordings` calls don't validate against documented parameter format
- No validation of response formats against API documentation

## Impact Assessment

**Severity:** HIGH
- **Test Failures:** Authentication will fail due to wrong parameter format
- **API Compliance:** Tests do not validate against ground truth
- **False Positives:** Tests may pass with wrong implementation
- **Ground Truth Violation:** Tests adapt to implementation instead of documentation

## Required Changes

### 1. Fix Authentication Parameter Format
**Current (Incorrect):**
```typescript
this.sendRequest(this.ws, 'authenticate', id, { token });
```

**Required (Correct):**
```typescript
this.sendRequest(this.ws, 'authenticate', id, { auth_token: token });
```

### 2. Add API Compliance Validation to Response Handling
**Current (Missing Validation):**
```typescript
if (data.error) {
  reject(new Error(data.error.message));
} else {
  resolve(data.result);
}
```

**Required (With Validation):**
```typescript
if (data.error) {
  // Validate error format against API documentation
  validateErrorResponse(data.error, method);
  reject(new Error(data.error.message));
} else {
  // Validate response format against API documentation
  validateResponseFormat(data.result, method);
  resolve(data.result);
}
```

### 3. Add Ground Truth References and Documentation
**Required Header:**
```typescript
/**
 * Stable Test Fixture for API Compliance Validation
 * 
 * Ground Truth References:
 * - Server API: ../mediamtx-camera-service/docs/api/json-rpc-methods.md
 * - Client Architecture: ../docs/architecture/client-architecture.md
 * - Client Requirements: ../docs/requirements/client-requirements.md
 * 
 * API Compliance Rules:
 * - All requests must match documented format
 * - All responses must be validated against API documentation
 * - No adaptation to implementation flaws
 * - Ground truth validation only
 */
```

### 4. Add API Compliance Validation Methods
```typescript
/**
 * Validate response format against API documentation
 */
private validateResponseFormat(result: any, method: string): void {
  // Validate based on method-specific requirements from API documentation
  switch (method) {
    case 'authenticate':
      this.validateAuthenticateResponse(result);
      break;
    case 'get_camera_list':
      this.validateCameraListResponse(result);
      break;
    // ... other methods
  }
}

/**
 * Validate error response format against API documentation
 */
private validateErrorResponse(error: any, method: string): void {
  // Validate error format per API documentation
  expect(error).toHaveProperty('code');
  expect(error).toHaveProperty('message');
  // ... method-specific error validation
}
```

### 5. Update All Method Calls to Include Validation
- Add authentication validation before protected method calls
- Validate all request parameters against API documentation
- Validate all response formats against API documentation
- Add proper error handling per API documentation

## Files Affected

### Primary Files:
- `tests/fixtures/stable-test-fixture.ts` (entire file)

### Related Files:
- All tests that use the stable test fixture
- Integration tests that depend on fixture functionality

## Acceptance Criteria

- [ ] Authentication parameter format matches API documentation
- [ ] All response handling validates against API documentation
- [ ] Ground truth references added to documentation
- [ ] API compliance validation methods implemented
- [ ] All method calls include proper validation
- [ ] No adaptation to implementation flaws
- [ ] Error response validation implemented

## Testing Rules Compliance

**✅ Ground Truth Validation:** Fixture must validate against frozen API documentation  
**❌ Current Status:** Fixture does not validate against API documentation  
**✅ No Code Peeking:** Fixture should not reference implementation code  
**❌ Current Status:** Fixture may be adapting to implementation details  

## Resolution Priority

**HIGH** - This affects all tests that use the stable test fixture and will cause authentication failures. The fixture must be updated to properly validate against the frozen API documentation.

## Related Issues

- Issue 001: Authentication Test Parameter Format Non-Compliance
- Issue 002: Camera List Test API Compliance Validation Missing
- Issue 003: Mock Server Response Format Non-Compliance
- Other fixture-related compliance issues
