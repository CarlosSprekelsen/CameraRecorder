# Issue 003: Mock Server Response Format Non-Compliance

**Status:** OPEN  
**Priority:** Medium  
**Type:** Test Infrastructure Issue  
**Created:** 2025-01-23  
**Discovered By:** API Compliance Validation  
**Category:** Test Suite Non-Compliance  

## Description

The mock server in `tests/fixtures/mock-server.ts` provides response formats that do not match the frozen API documentation ground truth. This creates false test passes when the actual server implementation differs from the documented API.

## Ground Truth Reference

**Source:** `mediamtx-camera-service/docs/api/json-rpc-methods.md` (FROZEN)  
**Methods:** All JSON-RPC methods  
**Requirement:** Mock responses must match documented response formats exactly

## Current Mock Implementation Issues

**File:** `tests/fixtures/mock-server.ts`  
**Lines:** 33-130  

### 1. GET_CAMERA_LIST Response
**Current Mock:**
```typescript
[RPC_METHODS.GET_CAMERA_LIST]: {
  cameras: [...],
  total: 2,
  connected: 2
}
```

**API Documentation Ground Truth:**
```json
{
  "jsonrpc": "2.0",
  "result": {
    "cameras": [...],
    "total": 1,
    "connected": 1
  },
  "id": 2
}
```

**Issue:** Mock returns raw result object instead of full JSON-RPC response format

### 2. TAKE_SNAPSHOT Response
**Current Mock:**
```typescript
[RPC_METHODS.TAKE_SNAPSHOT]: {
  device: '/dev/video0',
  filename: 'snapshot_2025-01-15_14-30-00.jpg',
  status: 'completed',
  timestamp: '2025-01-15T14:30:00Z',
  file_size: 204800,
  file_path: '/opt/camera-service/snapshots/snapshot_2025-01-15_14-30-00.jpg'
}
```

**API Documentation Ground Truth:**
```json
{
  "jsonrpc": "2.0",
  "result": {
    "status": "completed",
    "filename": "snapshot_2025-01-15_14-30-00.jpg",
    "file_size": 204800,
    "format": "jpeg",
    "quality": 85
  },
  "id": 4
}
```

**Issue:** Mock includes undocumented fields (`device`, `timestamp`, `file_path`) and missing documented fields (`format`, `quality`)

### 3. START_RECORDING Response
**Current Mock:**
```typescript
[RPC_METHODS.START_RECORDING]: {
  device: '/dev/video0',
  session_id: '550e8400-e29b-41d4-a716-446655440000',
  filename: 'camera0_2025-01-15_14-30-00.mp4',
  status: 'STARTED',
  start_time: '2025-01-15T14:30:00Z',
  duration: 3600,
  format: 'mp4'
}
```

**API Documentation Ground Truth:**
```json
{
  "jsonrpc": "2.0",
  "result": {
    "session_id": "550e8400-e29b-41d4-a716-446655440000",
    "status": "STARTED",
    "start_time": "2025-01-15T14:30:00Z"
  },
  "id": 5
}
```

**Issue:** Mock includes undocumented fields (`device`, `filename`, `duration`, `format`)

## Impact Assessment

**Severity:** MEDIUM
- **False Test Passes:** Tests pass with mock but fail with real server
- **API Compliance:** Mock does not validate against ground truth
- **Development Confusion:** Developers expect documented format but get different format
- **Ground Truth Violation:** Mock adapts to implementation instead of documentation

## Required Changes

### 1. Update Mock Response Format
**Current (Incorrect):**
```typescript
[RPC_METHODS.GET_CAMERA_LIST]: {
  cameras: [...],
  total: 2,
  connected: 2
}
```

**Required (Correct):**
```typescript
[RPC_METHODS.GET_CAMERA_LIST]: {
  jsonrpc: "2.0",
  result: {
    cameras: [...],
    total: 2,
    connected: 2
  },
  id: 1
}
```

### 2. Align All Mock Responses with API Documentation
- Remove undocumented fields from all mock responses
- Add missing documented fields
- Ensure response format matches API documentation exactly
- Include proper JSON-RPC wrapper (`jsonrpc`, `result`, `id`)

### 3. Add API Compliance Validation to Mock Server
- Validate mock responses against API documentation
- Add warnings when mock responses don't match ground truth
- Ensure mock server follows same validation rules as real server

### 4. Update Mock Server Documentation
- Add ground truth references
- Document API compliance requirements
- Reference frozen API documentation

## Files Affected

### Primary Files:
- `tests/fixtures/mock-server.ts` (lines 33-130)

### Related Files:
- Tests that use mock server responses
- Integration tests that may be affected by mock format changes

## Acceptance Criteria

- [ ] All mock responses match API documentation exactly
- [ ] Mock responses include proper JSON-RPC wrapper
- [ ] No undocumented fields in mock responses
- [ ] All documented fields present in mock responses
- [ ] Mock server includes API compliance validation
- [ ] Mock server documentation references ground truth

## Testing Rules Compliance

**✅ Ground Truth Validation:** Mock should validate against frozen API documentation  
**❌ Current Status:** Mock responses don't match API documentation  
**✅ No Code Peeking:** Mock should not adapt to implementation  
**❌ Current Status:** Mock may be adapting to implementation details  

## Resolution Priority

**MEDIUM** - This affects test reliability but doesn't block core functionality. Mock server should be updated to match API documentation to prevent false test passes and ensure proper validation.

## Related Issues

- Issue 001: Authentication Test Parameter Format Non-Compliance
- Issue 002: Camera List Test API Compliance Validation Missing
- Other mock-related compliance issues
