# BUG-001: ExternalStreamService Uses Non-Standard Logging Pattern

## Summary
ExternalStreamService uses custom log messages instead of the standard RPC logging pattern, causing inconsistency across the codebase and making debugging more difficult.

## Type
Defect - Code Quality

## Priority
Medium

## Severity
Minor

## Affected Component
- **Service**: `ExternalStreamService`
- **File**: `src/services/external/ExternalStreamService.ts`
- **Methods**: All public methods

## Environment
- **Version**: Current development branch
- **Architecture**: Client-side service layer

## Steps to Reproduce
1. Initialize `ExternalStreamService` with APIClient and LoggerService
2. Call any method (e.g., `discoverExternalStreams()`)
3. Inspect logger output
4. Compare with other services (e.g., `AuthService`, `DeviceService`)

## Expected Behavior
All service methods should follow the standard RPC logging pattern:
- **Request log**: `'{method_name} request', params`
- **Success log**: `'{method_name} success', result` (optional)
- **Error log**: `'{method_name} failed', error`

Example from standard pattern:
```typescript
this.logger.info('discover_external_streams request', params);
// ... operation ...
this.logger.error('discover_external_streams failed', error);
```

## Actual Behavior
ExternalStreamService uses custom log messages:
- **Request log**: `'Discovering external streams', params`
- **Success log**: `'Discovered ${totalStreams} external streams'`
- **Error log**: `'Failed to discover external streams', error`

## Root Cause Analysis

### Code Location
File: `src/services/external/ExternalStreamService.ts`

### Affected Methods
1. `discoverExternalStreams()` (lines 56, 61, 64)
2. `addExternalStream()` (lines 79, 82, 85)
3. `removeExternalStream()` (lines 93, 96, 99)
4. `getExternalStreams()` (lines 107, 110, 113)
5. `setDiscoveryInterval()` (lines 121, 124, 127)

### Code Evidence
```typescript
// Line 56 - Non-standard request log
this.logger.info('Discovering external streams', params);

// Line 61 - Non-standard success log
this.logger.info(`Discovered ${totalStreams} external streams`);

// Line 64 - Non-standard error log
this.logger.error('Failed to discover external streams', error as Record<string, unknown>);
```

### Expected Standard Pattern
```typescript
// Standard request log
this.logger.info('discover_external_streams request', params);

// Standard error log
this.logger.error('discover_external_streams failed', error);
```

### Why This Occurred
The service was likely implemented independently without reference to the established logging standards used in other services (AuthService, DeviceService, RecordingService, etc.).

### Impact Assessment
- **Debugging**: Developers cannot use consistent log filtering patterns (e.g., grep for 'request' or 'failed')
- **Monitoring**: Log aggregation tools cannot reliably parse service operations
- **Testing**: Tests expect standard pattern, causing 12 test failures
- **Documentation**: Logging pattern is inconsistent with API documentation expectations

## Test Evidence

### Failing Tests
File: `tests/unit/services/external_stream_service.test.ts`

Failing assertions (12 total):
```typescript
// Line 70 - Expected standard pattern
expect(mockLoggerService.info).toHaveBeenCalledWith('discover_external_streams request', {});
// Actual: Called with 'Discovering external streams', {}

// Line 119 - Expected standard pattern  
expect(mockLoggerService.info).toHaveBeenCalledWith('discover_external_streams request', options);
// Actual: Called with 'Discovering external streams', options

// Line 131 - Expected standard error pattern
expect(mockLoggerService.error).toHaveBeenCalledWith('discover_external_streams failed', error);
// Actual: Called with 'Failed to discover external streams', error
```

### Test Output
```
FAIL tests/unit/services/external_stream_service.test.ts
  ● ExternalStreamService Unit Tests › REQ-EXT-001: discover_external_streams RPC method › Should call WebSocket service with default parameters
    
    expect(jest.fn()).toHaveBeenCalledWith(...expected)
    
    Expected: "discover_external_streams request", {}
    Received: "Discovering external streams", {}
```

## Comparison with Correct Implementation

### AuthService (Correct Standard Pattern)
File: `src/services/auth/AuthService.ts`
```typescript
// Uses standard pattern (method_name + 'request'/'failed')
this.logger.info('authenticate request', params);
this.logger.error('authenticate failed', error);
```

### DeviceService (Correct Standard Pattern)
File: `src/services/device/DeviceService.ts`
```typescript
// Uses standard pattern
this.logger.info('get_camera_list request');
this.logger.error('get_camera_list failed', error);
```

## Related Documentation
- **Testing Guidelines**: `docs/development/client-testing-guidelines.md`
- **Architecture**: `docs/architecture/client-architechture.md`
- **API Documentation**: `docs/api/mediamtx_camera_service_openrpc.json`

## Acceptance Criteria
1. All log messages in `ExternalStreamService` follow the pattern: `'{method_name} request'` for requests
2. All error logs follow the pattern: `'{method_name} failed'` for errors
3. All 12 failing tests in `external_stream_service.test.ts` pass
4. Log messages match the JSON-RPC method names from API documentation

## Proposed Fix

### Changes Required
Update all logging statements in `src/services/external/ExternalStreamService.ts`:

**Method: discoverExternalStreams()**
```typescript
// Line 56: Change from
this.logger.info('Discovering external streams', params);
// To
this.logger.info('discover_external_streams request', params);

// Line 64: Change from
this.logger.error('Failed to discover external streams', error as Record<string, unknown>);
// To
this.logger.error('discover_external_streams failed', error);
```

**Method: addExternalStream()**
```typescript
// Line 79: Change from
this.logger.info('Adding external stream', params);
// To
this.logger.info('add_external_stream request', params);

// Line 85: Change from
this.logger.error('Failed to add external stream', error as Record<string, unknown>);
// To
this.logger.error('add_external_stream failed', error);
```

**Method: removeExternalStream()**
```typescript
// Line 93: Change from
this.logger.info('Removing external stream: ' + params);
// To
this.logger.info('remove_external_stream request', { streamUrl: params });

// Line 99: Change from
this.logger.error('Failed to remove external stream: ' + params, error as Record<string, unknown>);
// To
this.logger.error('remove_external_stream failed', error);
```

**Method: getExternalStreams()**
```typescript
// Line 107: Change from
this.logger.info('Getting external streams');
// To
this.logger.info('get_external_streams request');

// Line 113: Change from
this.logger.error('Failed to get external streams', error as Record<string, unknown>);
// To
this.logger.error('get_external_streams failed', error);
```

**Method: setDiscoveryInterval()**
```typescript
// Line 121: Change from
this.logger.info('Setting discovery interval: ' + params);
// To
this.logger.info('set_discovery_interval request', { scanInterval: params });

// Line 127: Change from
this.logger.error('Failed to set discovery interval: ' + params, error as Record<string, unknown>);
// To
this.logger.error('set_discovery_interval failed', error);
```

## Verification Steps
1. Apply proposed changes to `src/services/external/ExternalStreamService.ts`
2. Run unit tests: `npm run test:unit -- --testPathPattern=external_stream_service.test.ts`
3. Verify all 12 previously failing tests now pass
4. Verify log output matches standard pattern during integration testing
5. Verify no regressions in other services

## Additional Notes
- Success logs (e.g., `'Discovered ${totalStreams} external streams'`) can be removed or kept as additional context, but should not replace the standard pattern
- The standard pattern ensures consistency across all services and enables automated log parsing
- This pattern aligns with JSON-RPC 2.0 method naming conventions

## Attachments
- Test failure output: See test run from 2025-10-01
- Service implementation: `src/services/external/ExternalStreamService.ts`
- Test expectations: `tests/unit/services/external_stream_service.test.ts`

