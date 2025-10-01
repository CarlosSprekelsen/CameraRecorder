# BUG-024: File Format Validation Issues

## Summary
FileService tests are failing due to missing format information in file info responses. Tests expect format fields but the service doesn't provide them.

## Test Failures
```
● FileService Unit Tests › File format validation › should handle different recording formats
  Expected: "mp4"
  Received: undefined

● FileService Unit Tests › File format validation › should handle different snapshot formats  
  Expected: "jpg"
  Received: undefined
```

## Root Cause
- FileService doesn't extract or return format information from file responses
- Tests expect format field in file info responses but service doesn't provide it
- Missing format extraction logic in service implementation

## Expected Behavior
- FileService should extract file format from filenames or metadata
- File info responses should include format information
- Format validation should work for both recordings and snapshots

## Priority
**LOW** - Affects file format validation functionality

## Assignee
**File Management Team**

## Files to Fix
- `src/services/file/FileService.ts` - Add format extraction logic
- `tests/unit/services/file_service.test.ts` - Update format validation tests
- `src/types/api.ts` - Ensure format field is included in file info types
