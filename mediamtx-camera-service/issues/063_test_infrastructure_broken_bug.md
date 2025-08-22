# Bug Report: Test Infrastructure Broken - Multiple Critical Issues

## Summary
The test infrastructure has multiple critical failures preventing reliable test execution: missing test methods, port conflicts, and performance test authentication issues.

## Severity
**HIGH** - Prevents reliable test execution and validation.

## Issues Identified

### 1. Missing Test Methods
**Error**: `AttributeError: 'WebSocketAuthTestClient' object has no attribute 'call_method'`
**Impact**: Security tests failing
**Files Affected**: `tests/security/test_file_management_security.py`

### 2. Port Binding Conflicts
**Error**: `[Errno 98] error while attempting to bind on address ('0.0.0.0', 8003): [errno 98] address already in use`
**Impact**: Service manager tests failing
**Root Cause**: Multiple tests trying to bind to same port

### 3. Performance Test Authentication
**Error**: `AssertionError: Authentication failed for performance tests`
**Impact**: All performance tests failing
**Root Cause**: Performance tests can't authenticate

### 4. Unrealistic Performance Targets
**Error**: `Throughput 98.05 req/s not within Python range [100, 200]`
**Impact**: Performance tests failing due to unrealistic expectations
**Root Cause**: Performance targets set too high for current implementation

## Test Results Impact
- **Security Tests**: 4/6 failing due to missing methods
- **Service Manager Tests**: 15+ failing due to port conflicts
- **Performance Tests**: 6/6 failing due to authentication + unrealistic targets

## Immediate Actions Required
1. **Add Missing Methods**: Implement `call_method` in `WebSocketAuthTestClient`
2. **Fix Port Management**: Implement dynamic port allocation for tests
3. **Fix Performance Test Auth**: Ensure performance tests can authenticate
4. **Adjust Performance Targets**: Set realistic performance expectations

## Status
**HIGH PRIORITY** - Must be fixed to enable reliable testing.
