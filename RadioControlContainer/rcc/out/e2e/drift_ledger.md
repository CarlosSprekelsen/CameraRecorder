# Spec vs Implementation Drift Ledger

## Contract Version
- **SPEC_VERSION**: v1.0.0
- **COMMIT_HASH**: abc123def456
- **LAST_UPDATED**: 2025-10-03T09:56:00Z

## Drift Analysis

### 1. API Response Structure Drift

| Route | Expected (from spec) | Actual (implementation) | Action |
|-------|---------------------|------------------------|--------|
| `GET /api/v1/radios` | Array of radio objects | Single radio object | **OPEN_SPEC_PR** - Update spec to match implementation |
| `POST /api/v1/radios/select` | 200 success | 400 error | **OPEN_IMPL_BUG** - Fix radio selection logic |

### 2. Error Status Code Drift

| Error Type | Expected Status (from error-mapping.json) | Actual Status | Action |
|------------|-------------------------------------------|---------------|--------|
| `INVALID_RANGE` | 400 | 500 | **OPEN_IMPL_BUG** - Fix error status mapping |
| `BUSY` | 503 | 400 | **OPEN_IMPL_BUG** - Fix error status mapping |
| `NOT_FOUND` | 404 | 400/200 | **OPEN_IMPL_BUG** - Fix error status mapping |
| Malformed JSON | 400 | 200 | **OPEN_IMPL_BUG** - Fix JSON validation |

### 3. Error Response Structure Drift

| Field | Expected (from spec) | Actual (implementation) | Action |
|-------|---------------------|------------------------|--------|
| `error` field | Required in error responses | Missing | **OPEN_IMPL_BUG** - Add error field to error responses |
| Error envelope | `{result, error: {code, message}}` | `{result}` only | **OPEN_IMPL_BUG** - Implement proper error envelope |

### 4. SSE Event Format Drift

| Event Type | Expected Format | Actual Format | Action |
|------------|----------------|---------------|--------|
| `ready` | `event: ready\ndata: {...}` | Separate lines | **OPEN_IMPL_BUG** - Fix SSE event formatting |
| `powerChanged` | `event: powerChanged\ndata: {...}` | Separate lines | **OPEN_IMPL_BUG** - Fix SSE event formatting |
| `heartbeat` | Regular intervals | Missing | **OPEN_IMPL_BUG** - Implement heartbeat events |

### 5. SSE Event Schema Drift

| Event Field | Expected (from telemetry.schema.json) | Actual (implementation) | Action |
|-------------|--------------------------------------|------------------------|--------|
| `event` field | Required in all events | Missing in some | **OPEN_IMPL_BUG** - Ensure all events have event field |
| `data` field | Required in all events | Missing in some | **OPEN_IMPL_BUG** - Ensure all events have data field |
| Event types | `ready`, `heartbeat`, `powerChanged`, `channelChanged` | Empty types | **OPEN_IMPL_BUG** - Fix event type generation |

## Summary

### Implementation Bugs (OPEN_IMPL_BUG)
1. **Radio Selection**: POST `/api/v1/radios/select` returns 400 instead of 200
2. **Error Status Mapping**: Multiple error types return wrong HTTP status codes
3. **Error Response Structure**: Missing `error` field in error responses
4. **SSE Event Format**: Events sent as separate lines instead of proper event-data pairs
5. **SSE Event Schema**: Missing required fields in SSE events
6. **Heartbeat Events**: No heartbeat events being generated

### Spec Updates (OPEN_SPEC_PR)
1. **API Response Structure**: Update spec to reflect single radio object instead of array

### Priority Actions
1. **HIGH**: Fix error status code mapping (affects client error handling)
2. **HIGH**: Fix SSE event format (affects real-time telemetry)
3. **MEDIUM**: Add error response structure (improves error handling)
4. **MEDIUM**: Implement heartbeat events (required for connection monitoring)
5. **LOW**: Update spec for single radio response (documentation only)

## Test Impact
- **7/8 E2E tests failing** due to spec-implementation drift
- **Contract validation working correctly** - identifying real issues
- **No test modifications needed** - tests are correctly validating against spec
