# exit

## Issue Summary
**Priority:** HIGH  
**Type:** REQUIREMENT MISMATCH  
**Component:** FileManager  
**Test:** `test_file_manager_component.tsx` - "should format file sizes correctly"

## Problem Description
The FileManager component's `formatFileSize` function does not match the expected formatting requirements defined in the unit tests.

### Current Behavior
- Component renders: "1000 KB" and "1.95 MB"
- Test expects: "1 MB" and "2 MB"

### Test Data
```typescript
const mockRecordings = [
  {
    filename: 'recording-1.mp4',
    file_size: 1024000,  // 1 MB in bytes
    // ...
  },
  {
    filename: 'recording-2.mp4', 
    file_size: 2048000,  // 2 MB in bytes
    // ...
  }
];
```

## Requirements Analysis
**REQ-UNIT01-001:** File size display must be user-friendly and consistent
- File sizes should be displayed in appropriate units (KB, MB, GB)
- Rounding should be consistent and predictable
- Format should match user expectations

## Root Cause
The `formatFileSize` function in the FileManager component is not implementing the expected formatting logic:

### Server API Data Format
Server provides raw bytes in `file_size` field:
```json
{
  "filename": "recording-1.mp4",
  "file_size": 1024000,  // Raw bytes
  "modified_time": "2024-01-01T00:00:00Z"
}
```

### Current Component Behavior
Component formats sizes differently than test expectations:
- 1024000 bytes → "1000 KB" (test expects "1 MB")
- 2048000 bytes → "1.95 MB" (test expects "2 MB")

### Expected Formatting Logic
```typescript
const formatFileSize = (bytes: number): string => {
  if (bytes >= 1024 * 1024) {
    const mb = Math.round(bytes / (1024 * 1024));  // Round to whole number
    return `${mb} MB`;
  }
  // ... rest of function
};
```

## Expected Behavior
For the given test data:
- 1024000 bytes → "1 MB" (not "1000 KB")
- 2048000 bytes → "2 MB" (not "1.95 MB")

## Impact
- **Test Coverage:** 32 failed tests, 68 passed (75% pass rate)
- **User Experience:** Inconsistent file size display
- **Requirements Compliance:** FAILING

## Required Actions
1. **Review Requirements:** Confirm the expected formatting specification
2. **Fix Component:** Update `formatFileSize` function to match requirements
3. **Update Tests:** If requirements change, update test expectations
4. **Validate:** Ensure all file size displays are consistent

## Files Affected
- `src/components/FileManager/FileManager.tsx` - `formatFileSize` function
- `tests/unit/components/test_file_manager_component.tsx` - Test expectations

## Notes
- Do NOT force tests to pass if requirements are correct
- Component should be fixed to meet requirements
- If requirements are unclear, clarify with stakeholders first
