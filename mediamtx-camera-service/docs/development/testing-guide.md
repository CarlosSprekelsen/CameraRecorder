# Server Test Guide - MediaMTX Camera Service

**Version:** 2.0  
**Date:** 2025-08-20  
**Status:** Lean guidelines for server team  

## 1. Core Principles

### Real System Testing Over Mocking
- **MediaMTX:** Use systemd-managed service, never mock
- **File System:** Use `tempfile`, never mock
- **WebSocket:** Use real connections within system
- **Authentication:** Use real JWT tokens with test secrets

### Strategic Mocking Rules
**MOCK:** External APIs, time operations, expensive hardware simulation  
**NEVER MOCK:** MediaMTX service, filesystem, internal WebSocket, JWT auth, config loading

## 2. Authentication Standards

### Mandatory Shared Utilities
```python
# tests/fixtures/auth_utils.py - REQUIRED for all auth testing
def generate_valid_test_token(username="test_user", role="operator"):
    secret = os.getenv('CAMERA_SERVICE_JWT_SECRET', 'test-secret-dev-only')
    # Implementation returns valid JWT

def generate_expired_test_token():
    # Implementation returns expired JWT

def generate_invalid_test_token():
    return "invalid.jwt.token"
```

**FORBIDDEN:** Hard-coded credentials, expired tokens, auth mocking

## 3. Test Organization

### Directory Structure
```
tests/
├── unit/                    # <30 seconds total
├── integration/             # <5 minutes total  
├── fixtures/                # Shared utilities
└── performance/             # Load tests
```

### File Rules
- **One file per feature** - no variants (_real, _v2)
- **REQ-* references required** in every file docstring
- **Shared utilities over duplication**

## 4. Requirements Traceability

### Mandatory Format
```python
"""
Module description.

Requirements Coverage:
- REQ-XXX-001: Requirement description
- REQ-XXX-002: Additional requirement

Test Categories: Unit/Integration
"""

def test_feature_behavior_req_xxx_001(self):
    """REQ-XXX-001: Specific requirement validation."""
    # Test that would FAIL if requirement violated
```

## 5. Performance Targets

- **Unit tests:** <30 seconds total
- **Integration tests:** <5 minutes total  
- **Full suite:** <10 minutes total
- **Flaky rate:** <1%

### Test Markers
```python
@pytest.mark.unit          # Fast isolated tests
@pytest.mark.integration   # Real component integration
@pytest.mark.real_mediamtx # Requires systemd MediaMTX
@pytest.mark.performance   # Load/performance tests
```

## 6. Standard Patterns

### MediaMTX Integration
```python
@pytest.mark.real_mediamtx
async def test_stream_creation():
    controller = MediaMTXController("http://localhost:9997")
    stream_id = await controller.create_stream("test", "/dev/video0")
    assert stream_id is not None
```

### Authentication Testing
```python
async def test_valid_auth():
    token = generate_valid_test_token(role="operator")
    response = await client.authenticate(token)
    assert response["result"]["authenticated"] is True

async def test_expired_auth():
    token = generate_expired_test_token()
    with pytest.raises(AuthenticationError):
        await client.authenticate(token)
```

### File Operations
```python
async def test_recording():
    with tempfile.TemporaryDirectory() as temp_dir:
        manager = RecordingManager(recordings_dir=temp_dir)
        path = await manager.start_recording("/dev/video0", "test.mp4")
        assert Path(path).exists()
```

## 7. CI/CD Commands

```bash
make test-quick      # <30s unit tests only
make test-integration # <5min integration tests  
make test-full       # <10min complete suite
```

## 8. Quality Gates

**PASS Criteria:**
- Execution time <10 minutes
- Flaky rate <1%
- Requirements coverage >95%
- Zero hard-coded credentials
- Real MediaMTX integration working

**FAIL Criteria:**
- Hard-coded credentials found
- MediaMTX integration mocked
- Execution time >10 minutes
- Requirements coverage <95%

---

**Migration Required:** All existing tests must adopt these patterns