# Issue 065: Port Conflicts in Test Infrastructure

**Status**: Open  
**Priority**: Medium  
**Category**: Test Infrastructure  
**Date**: 2025-08-22  
**Reporter**: IV&V Assistant  

## Summary

Multiple integration tests are being skipped due to port conflicts. Tests are trying to bind to ports (8002, 8003) that are already in use by the running services.

## Error Details

```
[Errno 98] error while attempting to bind on address ('0.0.0.0', 8003): [errno 98] address already in use
```

## Affected Tests

**Skipped Tests**:
- `tests/unit/test_service_manager.py:92` - MediaMTX service not available due to port conflict
- `tests/integration/test_config_component_integration.py:39` - Service already running on port 8003
- `tests/integration/test_config_component_integration.py:128` - Service already running on port 8003

**Failed Tests**:
- `tests/integration/test_ffmpeg_integration.py::test_ffmpeg_integration` - Port 8002 already in use

## Root Cause Analysis

**Type**: Test Infrastructure Issue  
**Category**: Port Management

The tests are designed to start their own service instances on specific ports, but these ports are already occupied by the running MediaMTX and camera services. This violates AD-001 (Single Systemd-Managed MediaMTX Instance) which requires tests to use the existing systemd-managed services.

## Impact Assessment

- **Test Coverage**: 4 integration tests skipped, 1 failed
- **Test Isolation**: Tests cannot start independent service instances
- **Architecture Compliance**: Violates AD-001 decision

## Recommended Resolution

1. **Update Test Architecture**: Modify tests to use existing systemd-managed services instead of starting new instances
2. **Use Dynamic Port Allocation**: For tests that need independent instances, use dynamic port allocation
3. **Implement Service Coordination**: Ensure tests coordinate properly with running services
4. **Update Test Documentation**: Document the requirement to use existing services per AD-001

## Test Environment

- **MediaMTX Service**: ✅ Running on port 9997 (API), 8554 (RTSP), 8888 (HLS), 8889 (WebRTC)
- **Camera Service**: ✅ Running (likely on port 8002/8003)
- **Port Conflicts**: Ports 8002, 8003 already in use

## Architecture Decision Compliance

This issue highlights a violation of **AD-001: Single Systemd-Managed MediaMTX Instance**. Tests should be updated to:
- Use the existing systemd-managed MediaMTX service
- Coordinate on shared service with proper test isolation
- Not create multiple MediaMTX instances

## Related Issues

- Issue 063: Test infrastructure broken bug (overlapping infrastructure issues)
