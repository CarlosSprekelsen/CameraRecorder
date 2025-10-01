# BUG-008: StreamingService Missing Methods

## Summary
StreamingService is missing `getStreamUrl()` and `getStreams()` methods that tests expect, causing 4 test failures. Methods may have been moved to DeviceService during refactoring.

## Type
Defect - Missing Functionality / API Mismatch

## Priority
**HIGH**

## Severity
Major

## Affected Component
- **Service**: `StreamingService`
- **File**: `src/services/streaming/StreamingService.ts`
- **Missing Methods**: `getStreamUrl()`, `getStreams()`
- **Test File**: `tests/unit/services/streaming_service.test.ts`

## Environment
- **Version**: Current development branch  
- **Test Failures**: 4 tests

## Evidence

### Test Failures
```
● StreamingService Unit Tests › REQ-STREAM-003: get_stream_url RPC method
  TypeError: streamingService.getStreamUrl is not a function
  at Object.<anonymous> (tests/unit/services/streaming_service.test.ts:128:45)

● StreamingService Unit Tests › REQ-STREAM-005: get_streams RPC method
  TypeError: streamingService.getStreams is not a function
  at Object.<anonymous> (tests/unit/services/streaming_service.test.ts:222:45)
```

### Current StreamingService Methods
**File**: `src/services/streaming/StreamingService.ts`
- ✅ `startStreaming(device: string)`
- ✅ `stopStreaming(device: string)`
- ✅ `getStreamStatus(device: string)`
- ❌ `getStreamUrl()` - MISSING
- ❌ `getStreams()` - MISSING

### DeviceService Has Similar Methods
**File**: `src/services/device/DeviceService.ts` (lines 88-110)
```typescript
async getStreamUrl(device: string): Promise<string | null> {
  // Implementation exists in DeviceService
}

async getStreams(): Promise<StreamsListResult> {
  // Implementation exists in DeviceService  
}
```

## Root Cause Analysis
**Possible scenarios:**

1. **Methods moved during refactoring**: `getStreamUrl()` and `getStreams()` were moved from StreamingService to DeviceService but tests weren't updated

2. **Interface mismatch**: StreamingService should implement IStreaming interface, but the interface may be missing these method signatures

3. **Incomplete implementation**: Methods were planned but never implemented in StreamingService

## Expected Behavior  
**Option A**: StreamingService should have these methods:
```typescript
async getStreamUrl(device: string): Promise<StreamUrlResult>
async getStreams(): Promise<StreamsListResult>
```

**Option B**: Tests should be updated to use DeviceService for these methods instead of StreamingService

## Actual Behavior
- StreamingService only has 3 methods (start, stop, getStatus)
- Tests expect 5 methods (start, stop, getStatus, getStreamUrl, getStreams)
- DeviceService has the missing methods

## Impact
- 4 test failures in streaming_service.test.ts
- Unclear service boundaries - which service is responsible for stream queries?
- API inconsistency - consumers don't know where to find stream information

## Recommended Solution

**STOP: Clarify service responsibility boundaries**

Need to decide:

### Option A: Add methods to StreamingService (DRY violation - code duplication)
```typescript
// StreamingService.ts
async getStreamUrl(device: string): Promise<StreamUrlResult> {
  // Duplicate implementation from DeviceService
}

async getStreams(): Promise<StreamsListResult> {
  // Duplicate implementation from DeviceService
}
```
**Pros**: Tests pass without modification
**Cons**: Violates DRY - same code in two services

### Option B: Update tests to use DeviceService (RECOMMENDED)
```typescript
// streaming_service.test.ts - REMOVE these tests
// Move them to device_service.test.ts instead

// OR update tests to use deviceService instead of streamingService
const url = await deviceService.getStreamUrl(device);
const streams = await deviceService.getStreams();
```
**Pros**: Leverages existing code, no duplication, clear boundaries
**Cons**: Requires test refactoring

### Option C: Delegate from StreamingService to DeviceService
```typescript
// StreamingService.ts
constructor(
  private apiClient: IAPIClient,
  private logger: LoggerService,
  private deviceService: DeviceService  // NEW dependency
) {}

async getStreamUrl(device: string): Promise<StreamUrlResult> {
  return this.deviceService.getStreamUrl(device);
}
```
**Pros**: Maintains API compatibility
**Cons**: Adds unnecessary dependency layer

## Recommendation
**Option B** is recommended because:
1. Code already exists in DeviceService
2. Maintains clear service boundaries (Device = discovery/queries, Streaming = start/stop control)
3. No code duplication (DRY principle)
4. Tests should reflect actual architecture

## Testing Requirements
- streaming_service.test.ts should only test startStreaming, stopStreaming, getStreamStatus
- device_service.test.ts should test getStreamUrl, getStreams (already does)
- Integration tests should verify correct service is used for each operation

## Related Issues
- Similar to ServiceFactory API mismatch issues
- Part of broader service responsibility clarification needed

## Authorization Required
**STOP** - Need stakeholder decision on service boundaries before implementing fix.

