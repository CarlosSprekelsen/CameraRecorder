# BUG-018: API Validation Inconsistencies

## Summary
Multiple API validation failures indicating inconsistencies between test expectations, validation logic, and actual server responses across different test suites.

## Impact
- **Contract validation failures** in URL format validation
- **Storage info validation** failing unexpectedly  
- **JSON-RPC error handling** inconsistencies
- **API response format** mismatches

## Root Cause Analysis

### Primary Issues Identified

#### 1. URL Format Validation Failure
```
REQ-CONTRACT-004: Data Structure Contracts › should validate URL format contract
Expected: false
Received: true
```
- URL validator incorrectly passing invalid URLs
- Test expects validation to fail but validation passes

#### 2. Storage Info Validation Failure  
```
REQ-INT-011: Get storage information
Expected: true
Received: false
```
- `APIResponseValidator.validateStorageInfo(result)` failing
- Server returns valid data but validator rejects it

#### 3. JSON-RPC Error Handling Inconsistencies
```
REQ-E2E-005: should handle invalid JSON-RPC requests
Expected: false  
Received: true
```
- Tests expect certain invalid requests to fail gracefully
- Server accepts requests that should be rejected

#### 4. Unauthorized Access Test Failure
```
REQ-INT-013: Error handling for unauthorized access
Expected: Promise rejection
Received: Promise resolved with recording data
```
- Test expects unauthorized access to be blocked
- Server allows operation that should be restricted

### Architecture Compliance Check
**Ground Truth Reference**: `docs/api/json_rpc_methods.md`

1. **Validation Logic Issues**: Validators not matching actual server behavior
2. **Permission Boundaries**: Authorization not working as expected
3. **Error Response Format**: Inconsistent error handling patterns
4. **API Contract Mismatch**: Tests expect different behavior than server provides

## Evidence from Test Failures

### Contract Validation Issues
```typescript
// URL validation test failing
validator.validateUrl(url, 'test_url');
const result = validator.getResult();
expect(result.passed).toBe(false); // ❌ Fails - validation passes when should fail
```

### Storage Info Validation
```typescript
// Storage info validation failing  
const result = await apiClient.call('get_storage_info');
expect(APIResponseValidator.validateStorageInfo(result)).toBe(true); // ❌ Fails
```

### Permission Testing
```typescript
// Unauthorized access should be blocked but isn't
await expect(apiClient.call('start_recording', { device: 'camera0' }))
  .rejects.toThrow(); // ❌ Fails - operation succeeds
```

## Specific Validation Issues

### 1. URL Validation Logic
- **Problem**: Validator accepts URLs that should be rejected
- **Impact**: Contract validation tests failing
- **Location**: `APIResponseValidator.validateUrl()` method

### 2. Storage Info Schema Validation  
- **Problem**: `validateStorageInfo()` rejecting valid server responses
- **Impact**: Storage operations appearing broken
- **Location**: `APIResponseValidator.validateStorageInfo()` method

### 3. Permission Enforcement
- **Problem**: Server not enforcing role-based access control
- **Impact**: Security tests failing, unauthorized operations succeeding
- **Location**: Server-side permission validation

### 4. JSON-RPC Error Handling
- **Problem**: Invalid requests not being rejected properly
- **Impact**: Error handling tests failing
- **Location**: Server-side request validation

## Recommended Solution

### 1. Fix URL Validation Logic
```typescript
// Update URL validation to properly reject invalid formats
static validateUrl(url: string, fieldName: string): boolean {
  try {
    const urlObj = new URL(url);
    // Add proper validation rules
    return ['http:', 'https:', 'rtsp:'].includes(urlObj.protocol);
  } catch {
    return false;
  }
}
```

### 2. Fix Storage Info Validation
```typescript
// Debug and fix storage info validation
static validateStorageInfo(storageInfo: unknown): boolean {
  if (typeof storageInfo !== 'object' || storageInfo === null) return false;
  const obj = storageInfo as Record<string, unknown>;
  
  // Check actual server response format
  return typeof obj.total_space === 'number' && 
         typeof obj.used_space === 'number' &&
         typeof obj.available_space === 'number';
}
```

### 3. Fix Permission Testing
- **Investigate server-side permission enforcement**
- **Update test expectations** to match actual server behavior
- **Validate role-based access control** implementation

### 4. Fix JSON-RPC Error Handling
- **Review server-side request validation**
- **Update error handling tests** to match actual behavior
- **Ensure proper error codes** are returned

## Testing Requirements
- Fix URL validation logic to properly reject invalid URLs
- Fix storage info validation to accept valid server responses  
- Investigate and fix permission enforcement on server
- Update JSON-RPC error handling tests to match actual behavior
- Validate all API contracts against real server responses

## Classification
**Implementation Bug - API Validation Inconsistencies** - Multiple validation logic issues causing test failures across different API contracts.

## Priority Justification
**High priority** because:
- ❌ API contract validation broken
- ❌ Security/permission testing failing  
- ❌ Multiple validation logic issues
- ⚠️ Affects confidence in API reliability

## Effort Estimate
- **Medium complexity** - Multiple validation fixes required
- **Multiple test suites affected** - Contract and integration tests
- **Estimated time**: 2-3 hours to fix validation logic and update tests

## Dependencies
- May require server-side permission enforcement fixes
- Depends on understanding actual server response formats
- Requires coordination between client validation and server behavior
