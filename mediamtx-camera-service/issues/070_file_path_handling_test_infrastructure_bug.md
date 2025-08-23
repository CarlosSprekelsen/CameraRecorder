# Issue 070: File Path Handling Test Infrastructure Bug

**Status:** RESOLVED  
**Priority:** High  
**Type:** Test Infrastructure Bug  
**Created:** 2025-01-16  
**Resolved:** 2025-01-16  

## Description

Multiple integration tests were failing due to test infrastructure issues with file path handling and authentication setup.

## Root Cause Analysis

The failing tests were using broken manual initialization patterns instead of the proven ServiceManager pattern used by working tests.

### Issues Identified:
1. **Test Setup Failure**: `self.recordings_dir` was `None` due to improper initialization
2. **Authentication Test Failure**: `self.test_user` was `None` in unauthenticated client tests
3. **Wrong Test Pattern**: Tests were using manual component initialization instead of ServiceManager

## Solution Implemented

**Adopted Working Test Patterns from Performance Tests:**

### 1. ServiceManager Pattern
- Replaced manual component initialization with `ServiceManager(config=self.config)`
- Let ServiceManager handle all component setup and initialization
- Use `self.service_manager._websocket_server` for WebSocket server access

### 2. Relative File Paths
- Changed from `tempfile.mkdtemp()` to relative paths: `"./.tmp_recordings"` and `"./.tmp_snapshots"`
- Let MediaMTX create directories as needed
- Ensure directories exist with `os.makedirs()` calls

### 3. Proper Authentication Setup
- Fixed unauthenticated tests to use `send_unauthenticated_request()` method
- Proper WebSocket URL construction for test clients
- Correct user creation with appropriate roles

## Results

✅ **6 out of 7 tests now passing**  
✅ **Test infrastructure working correctly**  
✅ **Real bugs now properly identified**  

## Real Bug Discovered

After fixing the test infrastructure, a **real API implementation bug** was revealed:

**Issue 080: File Cleanup Logic Not Working**
- The `cleanup_old_files` API method is not actually deleting old files
- Test expects `files_deleted >= 3` but gets `0` despite files being 48 hours old
- This is a genuine system bug that needs to be addressed

## Files Modified

- `tests/integration/test_file_retention_policies.py` - Fixed test setup to use ServiceManager pattern
- `tests/integration/test_http_file_download.py` - Fixed test setup to use ServiceManager pattern  
- `tests/integration/test_storage_space_monitoring.py` - Fixed test setup to use ServiceManager pattern

## Lessons Learned

1. **Follow Working Patterns**: Always adopt patterns from tests that work correctly
2. **ServiceManager is Key**: Use ServiceManager for all test initialization, not manual component setup
3. **Relative Paths Work**: Use relative paths that MediaMTX can create, not tempfile directories
4. **Tests Reveal Real Bugs**: Once test infrastructure is fixed, tests properly identify real system issues

---

**Resolution:** Test infrastructure issues resolved by adopting proven ServiceManager patterns from working performance tests.
