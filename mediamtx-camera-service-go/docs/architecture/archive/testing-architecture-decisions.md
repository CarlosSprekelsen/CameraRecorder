# Testing Architecture Decisions

**⚠️ ARCHIVED DOCUMENT**  
This document has been superseded by the consolidated **[Go Architecture Guide](../go-architecture-guide.md)**.  
The Go implementation patterns and code examples for testing architecture are now available in the main guide.

---

**Version:** 1.0  
**Date:** 2025-08-15  
**Status:** Approved  
**Related:** All test development

## AD-001: Single Systemd-Managed MediaMTX Instance

### Decision
**All tests MUST use the single systemd-managed MediaMTX service instance. Tests MUST NOT create multiple MediaMTX instances or start their own MediaMTX processes.**

### Context
The MediaMTX Camera Service requires MediaMTX as its media server backend. During development and testing, we need to decide how to handle MediaMTX instances for testing purposes.

### Options Considered

#### Option A: Multiple Test-Specific MediaMTX Instances ❌ REJECTED
- **Approach:** Each test creates its own MediaMTX instance with unique ports
- **Pros:** Test isolation, no interference between tests
- **Cons:** 
  - Port conflicts and resource exhaustion
  - Orphaned processes and memory leaks
  - Complex port management and cleanup
  - Doesn't validate against production environment
  - Multiple instances competing for system resources

#### Option B: Single Systemd-Managed MediaMTX Instance ✅ APPROVED
- **Approach:** All tests use the production systemd-managed MediaMTX service
- **Pros:**
  - No port conflicts or resource issues
  - Validates against actual production environment
  - Simple service management via systemd
  - Real integration testing
  - No orphaned processes
- **Cons:**
  - Tests must coordinate on shared service
  - Requires systemd service to be running

### Implementation

#### Service Configuration
```bash
# MediaMTX service configuration
sudo systemctl start mediamtx
sudo systemctl enable mediamtx
sudo systemctl status mediamtx
```

#### Test Integration
```python
class RealMediaMTXServer:
    """Real MediaMTX server integration testing using systemd-managed service."""
    
    async def start(self) -> None:
        """Verify systemd-managed MediaMTX server is running."""
        # Check if MediaMTX service is running via systemd
        result = subprocess.run(["systemctl", "is-active", "mediamtx"])
        if result.returncode != 0:
            raise RuntimeError("MediaMTX systemd service is not running")
        
        # Wait for MediaMTX API to be ready
        await self._wait_for_mediamtx_ready()
```

#### Port Configuration
- **API Port:** 9997 (fixed systemd service port)
- **RTSP Port:** 8554 (fixed systemd service port)
- **WebRTC Port:** 8889 (fixed systemd service port)
- **HLS Port:** 8888 (fixed systemd service port)

#### Health Check Endpoint
```python
# Correct MediaMTX v1.13.1 API endpoint
health_url = "http://127.0.0.1:9997/v3/config/global/get"
```

### Consequences

#### Positive Consequences
- ✅ **No Port Conflicts:** Single instance prevents multiple MediaMTX processes
- ✅ **Resource Efficiency:** No orphaned processes or memory leaks
- ✅ **Production Validation:** Tests against actual production MediaMTX service
- ✅ **Real Integration:** Validates actual systemd service management
- ✅ **Simplified Testing:** No complex port management or cleanup required

#### Negative Consequences
- ⚠️ **Service Dependency:** Tests require MediaMTX systemd service to be running
- ⚠️ **Shared State:** Tests must coordinate on shared service (mitigated by proper test isolation)

### Validation

#### Success Criteria
- [x] All integration tests pass using single MediaMTX instance
- [x] No port conflicts during test execution
- [x] No orphaned MediaMTX processes after test completion
- [x] Tests validate against actual production MediaMTX service
- [x] Systemd service management works correctly

#### Test Results
- ✅ **Issue T003 Resolved:** Comprehensive edge case coverage implemented
- ✅ **Real System Integration:** Tests use actual systemd-managed MediaMTX service
- ✅ **No Resource Issues:** Single instance prevents port conflicts and resource exhaustion

### Related Decisions
- **AD-002:** Real Component Testing Over Mocking
- **AD-003:** WebSocket JSON-RPC Protocol for API Communication

### References
- [Testing Guide](../development/testing-guide.md)
- [Integration Test README](../../tests/integration/README_REAL_SYSTEM.md)
- [Test Status](../../test_status.md)
