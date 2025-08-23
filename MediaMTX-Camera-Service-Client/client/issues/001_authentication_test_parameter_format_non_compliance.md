# Issue 001: Authentication Test Parameter Format Non-Compliance

**Status:** OPEN  
**Priority:** High  
**Type:** Test Compliance Issue  
**Created:** 2025-01-23  
**Discovered By:** API Compliance Validation  
**Category:** Test Suite Non-Compliance  

## Description

The authentication test in `tests/integration/test_authentication_setup_integration.js` uses incorrect parameter format that does not match the API documentation ground truth.

## Ground Truth Reference

**Source:** `mediamtx-camera-service/docs/api/json-rpc-methods.md` (FROZEN)  
**Method:** `authenticate`  
**Documented Parameter Format:** `{ auth_token: string }`

## Current Test Implementation

**File:** `tests/integration/test_authentication_setup_integration.js`  
**Lines:** 99-103  
**Current Format:**
```javascript
const authResult = await sendRequest(ws, 'authenticate', {
  token: authToken  // ❌ INCORRECT: should be 'auth_token'
});
```

## API Documentation Ground Truth

**Correct Format:**
```json
{
  "jsonrpc": "2.0",
  "method": "authenticate",
  "params": {
    "auth_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."  // ✅ CORRECT
  },
  "id": 0
}
```

## Impact Assessment

**Severity:** HIGH
- **Test Failure:** Test will fail with parameter validation errors
- **API Compliance:** Test does not validate against ground truth
- **Development Confusion:** Developers following test will use wrong format
- **Ground Truth Violation:** Test adapts to wrong implementation instead of validating against documentation

## Required Changes

### 1. Fix Parameter Format
**Current (Incorrect):**
```javascript
await sendRequest(ws, 'authenticate', {
  token: authToken
});
```

**Required (Correct):**
```javascript
await sendRequest(ws, 'authenticate', {
  auth_token: authToken
});
```

### 2. Add API Compliance Validation
- Add validation against documented response format
- Check for all required fields: `authenticated`, `role`, `permissions`, `expires_at`, `session_id`
- Validate error response format for invalid tokens

### 3. Update Test Documentation
- Add ground truth references to test header
- Document API compliance validation approach
- Reference frozen API documentation

## Files Affected

### Primary Files:
- `tests/integration/test_authentication_setup_integration.js` (lines 99-103, 167-170)

### Related Files:
- `tests/integration/test_real_network_integration.ts` (line 75)
- `tests/integration/test_real_camera_operations_integration.ts` (line 65)
- `tests/fixtures/stable-test-fixture.ts` (line 128)

## Acceptance Criteria

- [ ] Parameter format matches API documentation exactly
- [ ] Test validates against documented response format
- [ ] Test includes proper ground truth references
- [ ] Test follows API compliance validation rules
- [ ] No adaptation to existing implementation flaws

## Testing Rules Compliance

**✅ Ground Truth Validation:** Test must validate against frozen API documentation  
**❌ Current Status:** Test uses incorrect parameter format  
**✅ No Code Peeking:** Test should not reference implementation code  
**❌ Current Status:** Test may be adapting to implementation flaws  

## Resolution Priority

**HIGH** - This blocks proper API compliance validation and creates confusion for developers. The test must be updated to match the frozen API documentation before proceeding with other compliance work.

## Related Issues

- Issue 081: Authenticate Method Documentation vs Implementation Mismatch Bug (RESOLVED)
- Other authentication-related test compliance issues
