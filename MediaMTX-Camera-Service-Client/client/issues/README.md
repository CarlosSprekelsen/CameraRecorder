# Bug Reports - MediaMTX Camera Service Client

This directory contains detailed bug reports identified during test suite validation on 2025-10-01.

## Bug Report Index

### Active Issues

| ID | Title | Type | Priority | Severity | Component |
|----|-------|------|----------|----------|-----------|
| [BUG-001](./BUG-001-ExternalStreamService-Inconsistent-Logging.md) | ExternalStreamService Uses Non-Standard Logging Pattern | Code Quality | Medium | Minor | ExternalStreamService |
| [BUG-002](./BUG-002-ServerService-Missing-Operation-Logging.md) | ServerService Missing Operation Logging | Missing Functionality | High | Major | ServerService |
| [BUG-003](./BUG-003-WebSocketServiceTest-Mock-Synchronization.md) | Mock State Synchronization Issue in WebSocket Service Test | Test Implementation | Low | Minor | Unit Tests |

## Issue Categories

### Code Defects (2)
- **BUG-001**: Logging pattern inconsistency in ExternalStreamService
- **BUG-002**: Missing logging in ServerService methods

### Test Defects (1)
- **BUG-003**: Mock synchronization issue in websocket test

## Summary Statistics

- **Total Issues**: 3
- **Code Bugs**: 2
- **Test Bugs**: 1
- **High Priority**: 1
- **Medium Priority**: 1
- **Low Priority**: 1

## Test Impact

| Bug ID | Failing Tests | Test File |
|--------|---------------|-----------|
| BUG-001 | 12 tests | `tests/unit/services/external_stream_service.test.ts` |
| BUG-002 | 9 tests | `tests/unit/services/event_subscription_service.test.ts` |
| BUG-003 | 1 test | `tests/unit/services/websocket_service_simple.test.ts` |
| **Total** | **22 tests** | **3 test files** |

## Root Cause Analysis Summary

### BUG-001: ExternalStreamService Inconsistent Logging
- **Root Cause**: Service implemented independently without reference to established logging standards
- **Impact**: 12 test failures, inconsistent debugging experience, log parsing difficulties
- **Fix Complexity**: Low - Simple find-and-replace in log statements

### BUG-002: ServerService Missing Operation Logging
- **Root Cause**: Methods implemented as simple pass-through wrappers without logging
- **Impact**: 9 test failures, zero visibility into subscription operations, audit trail gaps
- **Fix Complexity**: Low - Add try-catch blocks with logging to 3 methods

### BUG-003: WebSocket Test Mock Synchronization
- **Root Cause**: Test updated one mock but forgot to synchronize related mock
- **Impact**: 1 test failure, false negative in test suite
- **Fix Complexity**: Very Low - Add one line to synchronize mock state

## Verification Status

All bugs have been documented with:
- ✅ Steps to reproduce
- ✅ Expected vs actual behavior
- ✅ Root cause analysis
- ✅ Code evidence with line numbers
- ✅ Test evidence with failure output
- ✅ Proposed fixes with code examples
- ✅ Acceptance criteria
- ✅ Verification steps

## Environment

- **Test Run Date**: 2025-10-01
- **Branch**: Development
- **Test Framework**: Jest
- **Total Test Suite**: 496 tests
- **Tests Passing**: 252
- **Tests Failing**: 244
- **Tests Affected by These Bugs**: 22

## Related Documentation

- Testing Guidelines: `docs/development/client-testing-guidelines.md`
- Architecture: `docs/architecture/client-architechture.md`
- API Documentation: `docs/api/mediamtx_camera_service_openrpc.json`

## Notes

These bug reports follow professional bug tracking standards (Jira-style):
- Factual, objective descriptions
- No opinions or biases
- Detailed root cause analysis
- Clear reproduction steps
- Specific fix proposals
- Measurable acceptance criteria

Each bug is documented in a separate file for independent tracking and resolution.

