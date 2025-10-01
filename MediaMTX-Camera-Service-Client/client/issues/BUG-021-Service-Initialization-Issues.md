# BUG-021: Service Initialization Issues

## Summary
Multiple components are failing due to services not being properly initialized in tests. Components are showing "Service not initialized" errors instead of expected functionality.

## Affected Services
- **FileService**: "File service not initialized" in FilesPage
- **ServerService**: "Server service not initialized" in AboutPage  
- **DeviceService**: Missing service injection in CameraPage
- **RecordingService**: Missing service injection in RecordingController

## Root Cause
The `component-test-helper.ts` is not properly injecting mock services into stores. Components are trying to use services that haven't been mocked or initialized.

## Test Failures
```
● FilesPage Component › REQ-FILESPAGE-001: FilesPage renders file management interface
  Unable to find an element with the text: File Management
  Shows: "File service not initialized"

● AboutPage Component › REQ-ABOUT-001: AboutPage renders information content  
  Unable to find an element with the text: About
  Shows: "Server service not initialized"
```

## Expected Behavior
- Services should be properly mocked and injected into stores
- Components should render their main content instead of service initialization errors
- Tests should provide working service mocks

## Priority
**HIGH** - Blocks multiple component tests from passing

## Status
**OPEN** - Initial fix attempt failed. Test helper creates mock objects but doesn't inject them into real Zustand stores. Components still use real stores without mocked services.

## Assignee
**Service Architecture Team**

## Files to Fix
- `tests/utils/component-test-helper.ts` - Fix service initialization to inject into real stores
- `tests/utils/mocks.ts` - Ensure all services are properly mocked
- `tests/unit/components/pages/FilesPage.test.tsx`
- `tests/unit/components/pages/AboutPage.test.tsx`
- `tests/unit/components/pages/CameraPage.test.tsx`
