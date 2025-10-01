# BUG-010: Store Error Handling Pattern - Uncaught Service Initialization Errors

## Summary
Device, Recording, and File stores throw "service not initialized" errors OUTSIDE try-catch blocks, causing Jest worker crashes. The service validation check needs to be inside the error handling block.

## Type
Defect - Error Handling Pattern

## Priority
**HIGH**

## Severity
Critical (crashes test suite, likely affects production error handling)

## Affected Components
- **Store**: `DeviceStore` - File: `src/stores/device/deviceStore.ts`
- **Store**: `RecordingStore` - File: `src/stores/recording/recordingStore.ts`
- **Store**: `FileStore` - File: `src/stores/file/fileStore.ts`
- **Test Crashes**: 3 complete test suites fail to run

## Environment
- **Version**: Current development branch
- **Test Framework**: Jest 29.7.0
- **Node Version**: v24.6.0

## Root Cause - IDENTIFIED

### The Pattern
All three failing stores use this BROKEN pattern:

**DeviceStore - Line 61-76:**
```typescript
getCameraList: async () => {
  if (!deviceService) throw new Error('Device service not initialized');  // ❌ THROWN OUTSIDE try-catch
  
  set({ loading: true, error: null });
  try {
    const cameras = await deviceService.getCameraList();
    set({ cameras, loading: false, lastUpdated: new Date().toISOString(), error: null });
  } catch (error) {  // ⚠️ Never catches the service initialization error
    set({ loading: false, error: error instanceof Error ? error.message : 'Failed to get camera list' });
  }
}
```

**RecordingStore - Similar pattern**
**FileStore - Similar pattern**

### Why It Crashes
1. Test calls `getCameraList()` with null service (line 172 in test)
2. Store throws error on line 62: `throw new Error('Device service not initialized')`
3. Error is thrown BEFORE the try-catch block
4. Error propagates to Jest worker
5. Jest worker crashes with "child process exceptions"

## Evidence

### Test Crashes
```
Error: Device service not initialized
    at getCameraList (/src/stores/device/deviceStore.ts:62:15)
    at Object.<anonymous> (/tests/unit/stores/device_store.test.ts:172:7)

Error: Recording service not initialized
    at startRecording (/src/stores/recording/recordingStore.ts:1285:15)
    at Object.<anonymous> (/tests/unit/stores/recording_store.test.ts:179:7)

Error: File service not initialized
    at loadRecordings (/src/stores/file/fileStore.ts:2797:15)
    at Object.<anonymous> (/tests/unit/stores/file_store.test.ts:208:7)
```

### Test Expectations
**device_store.test.ts lines 166-177:**
```typescript
test('should handle service not initialized error', () => {
  const { getCameraList } = useDeviceStore.getState();
  
  useDeviceStore.getState().setDeviceService(null as any);
  
  getCameraList();  // ❌ Expects this to NOT throw
  
  const state = useDeviceStore.getState();
  expect(state.error).toBe('Device service not initialized');  // ✅ Expects error in state
  expect(state.loading).toBe(false);
});
```

Tests expect: error caught and stored in state  
Actual behavior: error thrown and crashes worker

## Working Pattern - Auth Store

**AuthStore - CORRECT pattern (lines 37-52):**
```typescript
login: async (username: string, password: string) => {
  set({ loading: true, error: null });
  try {
    if (!authService) {  // ✅ Check INSIDE try-catch
      throw new Error('Auth service not initialized');
    }
    
    const result = await authService.login({ username, password });
    // ... success handling
  } catch (error) {  // ✅ Catches ALL errors including service check
    set({
      loading: false,
      error: error instanceof Error ? error.message : 'Login failed',
    });
  }
}
```

## Expected Behavior
1. Store method is called with null service
2. Error is thrown but caught by try-catch
3. Error message is set in store state
4. Loading is set to false
5. No exception propagates to caller

## Actual Behavior
1. Store method is called with null service
2. Error is thrown BEFORE try-catch
3. Error propagates to Jest worker
4. Jest worker crashes
5. Test suite never runs

## Impact
- **CRITICAL**: 3 complete test suites crash (device, recording, file)
- **Zero test coverage** for these stores
- **Production risk**: Same pattern likely fails in production when services aren't initialized
- **User impact**: Unhandled exceptions instead of graceful error messages

## Recommended Solution

### Fix Pattern - Move Service Check Inside Try-Catch

**BEFORE (BROKEN):**
```typescript
getCameraList: async () => {
  if (!deviceService) throw new Error('Device service not initialized');  // ❌
  
  set({ loading: true, error: null });
  try {
    const cameras = await deviceService.getCameraList();
    set({ cameras, loading: false, ... });
  } catch (error) {
    set({ loading: false, error: ... });
  }
}
```

**AFTER (FIXED):**
```typescript
getCameraList: async () => {
  set({ loading: true, error: null });
  try {
    if (!deviceService) {  // ✅ Check INSIDE try
      throw new Error('Device service not initialized');
    }
    const cameras = await deviceService.getCameraList();
    set({ cameras, loading: false, lastUpdated: new Date().toISOString(), error: null });
  } catch (error) {  // ✅ Catches service check AND API errors
    set({
      loading: false,
      error: error instanceof Error ? error.message : 'Failed to get camera list',
    });
  }
}
```

## Files Requiring Fix

### DeviceStore (3 methods estimated)
- `getCameraList()` - Line ~62
- `getStreamUrl()` - Line ~81
- `getStreams()` - Line ~104

### RecordingStore (3 methods estimated)  
- `startRecording()` - Line ~1285
- `stopRecording()` - Similar pattern
- `takeSnapshot()` - Similar pattern

### FileStore (2 methods estimated)
- `loadRecordings()` - Line ~2797
- `loadSnapshots()` - Similar pattern

**Total**: ~8 methods across 3 stores

## DRY Principle
This fix leverages the EXISTING working pattern from AuthStore. Apply the same pattern consistently across all stores.

## Testing Requirements
Once fixed:
- All 3 store test suites should load without crashing
- Test "should handle service not initialized error" should PASS
- `state.error` should contain error message
- `state.loading` should be false
- No exceptions should propagate to Jest worker

## Priority Justification
**HIGH priority** because:
- ❌ 3 complete test suites cannot run
- ❌ Zero test coverage for critical stores
- ❌ Same bug likely exists in production code
- ❌ Unhandled exceptions = poor user experience

## Related Issues
- Similar to error handling patterns in other stores
- Auth Store already has correct pattern - use as reference
- Connection Store and Server Store also working (likely have correct pattern)

