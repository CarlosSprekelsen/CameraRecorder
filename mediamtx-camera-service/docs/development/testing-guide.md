# Test Guidelines - MediaMTX Camera Service

**Version:** 1.0  
**Authors:** Project Team  
**Date:** 2025-08-15  
**Status:** Approved  
**Related Epic/Story:** All development stories  

**Purpose:**  
Define comprehensive test design principles, requirements traceability standards, and execution guidelines for the MediaMTX Camera Service project. This document ensures all tests validate real system behavior and maintain clear traceability to requirements.

---

## 1. Core Testing Principles

### Real System Testing Over Mocking
**Principle:** Test with real components whenever possible to discover actual integration issues.

#### MediaMTX Architecture Decision: Single Systemd-Managed Instance
**CRITICAL:** All tests MUST use the single systemd-managed MediaMTX service instance. Tests MUST NOT create multiple MediaMTX instances or start their own MediaMTX processes.

**Rationale:**
- **Port Conflicts:** Multiple MediaMTX instances cause port conflicts and resource exhaustion
- **Resource Management:** Single instance prevents orphaned processes and memory leaks
- **Production Reality:** Tests should validate against the actual production MediaMTX service
- **System Integration:** Real integration testing requires the actual systemd-managed service

**Implementation:**
```python
# ✅ CORRECT: Use systemd-managed MediaMTX service
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

# ❌ FORBIDDEN: Creating multiple MediaMTX instances
class WrongMediaMTXServer:
    async def start(self) -> None:
        # DON'T DO THIS - creates port conflicts and resource issues
        self.process = subprocess.Popen(["mediamtx", config_file])
```

**Configuration:**
- **API Port:** 9997 (fixed systemd service port)
- **RTSP Port:** 8554 (fixed systemd service port)
- **Health Check:** `/v3/config/global/get` endpoint
- **Service Management:** `systemctl start/stop/restart mediamtx`

```python
# ✅ PREFERRED: Test with real MediaMTX instance
async def test_stream_creation_real_mediamtx(real_mediamtx_config):
    controller = MediaMTXController(real_mediamtx_config)
    stream_id = await controller.create_stream("test_camera", "/dev/video0")
    assert stream_id is not None
    # Validate actual stream exists in MediaMTX
    streams = await controller.list_streams()
    assert "test_camera" in streams

# ❌ AVOID: Over-mocking that hides real issues
async def test_stream_creation_mocked(mock_mediamtx):
    mock_mediamtx.create_stream.return_value = "fake_id"
    # This test passes but tells us nothing about real behavior
```

**Real Components Available:**
- **MediaMTX Service:** Running on localhost with API access
- **USB Webcam:** Real `/dev/video*` devices for testing
- **FFmpeg:** Actual video processing and streaming
- **File System:** Real recordings and snapshot directories
- **Network Connections:** Actual WebSocket and HTTP communications

### Minimal Strategic Mocking
**When to Mock:**
- External services not under test control (remote APIs, cloud services)
- Expensive operations that don't affect core logic (large file operations)
- Hardware dependencies for unit tests (specific camera models)
- Time-dependent operations requiring precise control

**When NOT to Mock:**
- MediaMTX integration (real service available)
- File system operations (use temp directories)
- Network communication within system
- Configuration loading and validation
- Component interactions

```python
# ✅ GOOD: Mock external dependency only
@patch('requests.get')  # External API call
async def test_camera_metadata_fetch(mock_requests):
    mock_requests.return_value.json.return_value = {"model": "TestCam"}
    # Test real logic with mocked external dependency

# ✅ GOOD: Use real components with isolation
async def test_websocket_communication(isolated_websocket_server):
    client = WebSocketTestClient(server.url)
    await client.connect()
    response = await client.send_request("camera.list")
    # Real WebSocket communication, isolated test environment
```

---

## 2. Test Organization and Structure

### Directory Structure
```
tests/
├── unit/                           # Fast isolated tests
│   ├── test_camera_discovery/      # Camera discovery module
│   │   ├── test_hybrid_monitor.py          # Core monitor logic
│   │   ├── test_capability_detection.py    # V4L2 capability parsing
│   │   └── test_udev_integration.py        # Device event processing
│   ├── test_camera_service/        # Service management
│   │   ├── test_service_manager.py         # Lifecycle management
│   │   ├── test_config_manager.py          # Configuration handling
│   │   └── test_logging_config.py          # Logging setup
│   ├── test_mediamtx_wrapper/      # MediaMTX integration
│   │   ├── test_controller.py              # API wrapper logic
│   │   └── test_stream_manager.py          # Stream lifecycle
│   ├── test_websocket_server/      # WebSocket API
│   │   ├── test_server.py                  # Server lifecycle
│   │   ├── test_json_rpc.py               # RPC method handling
│   │   └── test_client_manager.py          # Client connection management
│   └── test_common/                # Shared utilities
│       ├── test_events.py                  # Event system
│       └── test_types.py                   # Common types
├── integration/                    # Component interaction tests
│   ├── test_camera_to_mediamtx.py          # Camera → MediaMTX flow
│   ├── test_websocket_notifications.py     # Event → WebSocket flow
│   └── test_end_to_end_workflows.py        # Complete user scenarios
└── fixtures/                       # Test data and utilities
    ├── conftest.py                         # Shared fixtures
    ├── test_configs/                       # Test configurations
    └── sample_data/                        # Sample outputs and responses
```

### Test File Naming Conventions
- **Test modules:** `test_<module_name>.py`
- **Test functions:** `test_<behavior>_<scenario>()` 
- **Test classes:** `TestModuleName` (when grouping related tests)

**Examples:**
```python
# File: tests/unit/test_camera_discovery/test_hybrid_monitor.py
def test_capability_detection_success_scenario()
def test_capability_detection_invalid_device()
def test_frame_rate_extraction_multiple_formats()

class TestHybridMonitor:
    def test_start_monitoring_with_valid_config()
    def test_stop_monitoring_cleanup_resources()
```

---

## 3. Requirements Traceability

### Test-to-Requirement Mapping
**Every test must clearly trace to specific requirements using standardized headers:**

```python
"""
Test camera discovery capability detection functionality.

Requirements Traceability:
- REQ-CAM-001: System shall detect USB camera capabilities automatically
- REQ-CAM-003: System shall extract supported resolutions and frame rates
- REQ-ERR-002: System shall handle invalid camera devices gracefully

Story Coverage: S3 - Camera Discovery & Monitoring
IV&V Control Point: Camera capability detection validation
"""

def test_capability_detection_standard_usb_camera():
    """
    Verify capability detection with real USB camera device.
    
    Requirements: REQ-CAM-001, REQ-CAM-003
    Scenario: Standard USB webcam with typical v4l2 capabilities
    Expected: Successful detection with resolution and frame rate metadata
    Edge Cases: Multiple resolution formats, unusual frame rate reporting
    """
```

### Requirement Coverage Matrix
**Maintain traceability in test docstrings and comments:**

```python
# Requirement coverage tracking
REQUIREMENT_COVERAGE = {
    "REQ-CAM-001": [
        "test_capability_detection_standard_usb_camera",
        "test_capability_detection_multiple_cameras", 
        "test_capability_detection_error_handling"
    ],
    "REQ-CAM-003": [
        "test_frame_rate_extraction_various_formats",
        "test_resolution_parsing_edge_cases"
    ]
}
```

### Acceptance Criteria Testing
**Each requirement's acceptance criteria must have corresponding test scenarios:**

```python
def test_req_cam_001_acceptance_criteria():
    """
    Test all acceptance criteria for REQ-CAM-001.
    
    Acceptance Criteria Coverage:
    1. ✓ Detect camera within 5 seconds of connection
    2. ✓ Extract resolution capabilities accurately  
    3. ✓ Handle device disconnection gracefully
    4. ✓ Report capability detection failures clearly
    """
    # Test implementation covering all criteria
```

---

## 4. Edge Case and Error Condition Testing

### Comprehensive Edge Case Coverage
**Test boundary conditions, error paths, and unusual scenarios:**

```python
class TestCameraDiscoveryEdgeCases:
    """Comprehensive edge case coverage for camera discovery."""
    
    def test_camera_disconnection_during_capability_detection(self):
        """Test device removal while detection in progress."""
        
    def test_multiple_cameras_same_model_identification(self):
        """Test disambiguation when identical cameras connected."""
        
    def test_camera_capability_detection_permission_denied(self):
        """Test behavior when lacking device access permissions."""
        
    def test_malformed_v4l2_output_parsing(self):
        """Test parsing with corrupted or unexpected v4l2-ctl output."""
        
    def test_camera_supports_zero_frame_rates(self):
        """Test edge case where camera reports no supported frame rates."""
        
    def test_extremely_long_device_path_handling(self):
        """Test system behavior with unusually long device paths."""
```

### Error Condition Validation
**Test all error paths and recovery mechanisms:**

```python
@pytest.mark.parametrize("error_scenario", [
    "device_not_found",
    "permission_denied", 
    "capability_detection_timeout",
    "invalid_device_format",
    "device_busy_by_other_process"
])
def test_camera_discovery_error_conditions(error_scenario):
    """Test camera discovery handles all error conditions appropriately."""
    # Implementation tests specific error scenario
    # Validates error reporting, logging, and recovery
```

### Performance and Load Testing
**Include performance validation in test suite:**

```python
@pytest.mark.performance
def test_camera_discovery_performance_multiple_devices():
    """Verify camera discovery performance with multiple devices."""
    start_time = time.time()
    
    # Simulate or use real multiple camera setup
    cameras = await discover_cameras(device_range=[0, 1, 2, 3, 4])
    
    detection_time = time.time() - start_time
    assert detection_time < 10.0  # REQ-PERF-001: <10s for 5 cameras
    assert len(cameras) >= 1  # At least one camera must be detected
```

---

## 5. Integration Testing Principles

### Real Component Integration
**Test actual component interactions without mocking internal interfaces:**

```python
@pytest.mark.integration
async def test_camera_to_websocket_notification_flow():
    """Test complete flow: camera event → MediaMTX → WebSocket notification."""
    
    # Use real components
    service_manager = ServiceManager(test_config)
    await service_manager.start()
    
    websocket_client = WebSocketTestClient("ws://localhost:8002/ws")
    await websocket_client.connect()
    
    # Trigger real camera event
    camera_simulator.connect_camera("/dev/video0")
    
    # Validate real notification delivery
    notification = await websocket_client.wait_for_notification("camera.connected")
    assert notification["params"]["device"] == "/dev/video0"
    
    await service_manager.stop()
```

### Cross-Component Data Validation
**Verify data consistency across component boundaries:**

```python
async def test_camera_metadata_consistency_across_components():
    """Verify camera metadata consistency between discovery and MediaMTX."""
    
    # Get metadata from camera discovery
    discovery_metadata = await camera_monitor.get_camera_capabilities("/dev/video0")
    
    # Create stream in MediaMTX with same camera
    stream_id = await mediamtx_controller.create_stream("test", "/dev/video0")
    
    # Get stream info from MediaMTX
    stream_info = await mediamtx_controller.get_stream_info(stream_id)
    
    # Validate metadata consistency
    assert discovery_metadata["resolution"] in stream_info["supported_resolutions"]
    assert discovery_metadata["frame_rate"] == stream_info["current_frame_rate"]
```

---

## 6. Test Execution Standards

### Test Environment Setup
**Ensure consistent test environment for reproducible results:**

```python
@pytest.fixture(scope="session")
def test_environment():
    """Setup consistent test environment."""
    return {
        "mediamtx_config": {
            "host": "localhost",
            "api_port": 9997,
            "rtsp_port": 8554
        },
        "temp_dirs": {
            "recordings": tempfile.mkdtemp(prefix="test_recordings_"),
            "snapshots": tempfile.mkdtemp(prefix="test_snapshots_")
        },
        "camera_devices": discover_available_test_cameras()
    }

@pytest.fixture(autouse=True)
def clean_test_state():
    """Ensure clean state between tests."""
    # Cleanup before test
    yield
    # Cleanup after test
    cleanup_test_resources()
```

### Test Data Management
**Use realistic test data that reflects actual system usage:**

```python
# fixtures/sample_data/v4l2_outputs.py
V4L2_SAMPLE_OUTPUTS = {
    "logitech_c920": """
        Driver name: uvcvideo
        Card type: HD Pro Webcam C920
        Capabilities: video capture, streaming
        Video input formats:
            [0]: 'YUYV' (YUYV 4:2:2)
            [1]: 'H264' (H.264)
        Frame rates: 30.000 fps, 25.000 fps, 20.000 fps
    """,
    "generic_usb_cam": """
        Driver name: uvcvideo
        Card type: USB 2.0 Camera
        Capabilities: video capture
        Video input formats:
            [0]: 'YUYV' (YUYV 4:2:2)
        Frame rates: 30.000 fps
    """
}
```

### Test Isolation and Cleanup
**Ensure tests don't interfere with each other:**

```python
async def test_stream_creation_with_cleanup():
    """Test stream creation with proper resource cleanup."""
    stream_id = None
    try:
        # Test implementation
        stream_id = await mediamtx_controller.create_stream("test", "/dev/video0")
        # Test assertions
        assert stream_id is not None
        
    finally:
        # Ensure cleanup even if test fails
        if stream_id:
            await mediamtx_controller.delete_stream(stream_id)
```

---

## 7. Quality Gates and Coverage

### Test Coverage Requirements
- **Unit Tests:** 85% minimum coverage for core modules
- **Integration Tests:** All critical user workflows covered
- **Edge Cases:** All error conditions and boundary cases tested
- **Requirements:** 100% requirement coverage through tests

### Test Quality Metrics
```python
# Quantitative quality gates
QUALITY_GATES = {
    "test_execution_time": {
        "unit_tests": "< 30 seconds",
        "integration_tests": "< 5 minutes",
        "full_suite": "< 10 minutes"
    },
    "test_reliability": {
        "flaky_test_rate": "< 1%",
        "false_positive_rate": "< 0.5%"
    },
    "requirement_coverage": {
        "functional_requirements": "100%",
        "non_functional_requirements": "90%"
    }
}
```

### Continuous Integration Requirements
**All tests must pass in CI environment:**

```bash
# CI test execution pipeline
pytest tests/unit/ --cov=src --cov-fail-under=85
pytest tests/integration/ --timeout=300
pytest tests/ -m "smoke" --timeout=60
```

---

## 8. Test Maintenance and Evolution

### Test Review Guidelines
- **Code Review:** All tests reviewed with same rigor as production code
- **Requirement Changes:** Update affected tests immediately when requirements change
- **Refactoring:** Keep tests synchronized with code refactoring
- **Documentation:** Update test documentation when test strategy evolves

### Test Debt Management
```python
# Acceptable test debt patterns
# TODO: HIGH: Add integration test for camera hot-swap [REQ-CAM-005]
# TODO: MEDIUM: Improve test execution speed [Performance]
# TODO: LOW: Add exhaustive edge case for unusual frame rates [REQ-CAM-003]
```

### Test Suite Evolution
**Evolve test suite as system matures:**
1. **Phase 1:** Focus on core functionality and happy paths
2. **Phase 2:** Add comprehensive edge case coverage
3. **Phase 3:** Performance and load testing
4. **Phase 4:** Security and reliability testing

---

## 9. Common Patterns and Examples

### Async Test Pattern
```python
@pytest.mark.asyncio
async def test_async_camera_discovery():
    """Standard async test with proper setup/teardown."""
    monitor = HybridCameraMonitor(test_config)
    try:
        await monitor.start()
        cameras = await monitor.discover_cameras()
        assert len(cameras) >= 0
    finally:
        await monitor.stop()
```

### Parametrized Test Pattern
```python
@pytest.mark.parametrize("device_path,expected_result", [
    ("/dev/video0", True),
    ("/dev/video999", False), 
    ("/invalid/path", False),
    ("", False)
])
def test_camera_device_validation(device_path, expected_result):
    """Test camera device path validation with various inputs."""
    result = validate_camera_device(device_path)
    assert result == expected_result
```

### Integration Test Pattern
```python
@pytest.mark.integration
@pytest.mark.timeout(60)
async def test_end_to_end_camera_streaming():
    """Test complete camera streaming workflow."""
    # Real system integration test
    service = await start_camera_service(real_config)
    client = WebSocketTestClient(service.websocket_url)
    
    try:
        await client.connect()
        
        # Start streaming
        response = await client.send_request("camera.start_stream", {
            "device": "/dev/video0",
            "resolution": "640x480",
            "frame_rate": 30
        })
        
        assert response["result"]["status"] == "streaming"
        
        # Verify stream accessible
        stream_url = response["result"]["stream_url"]
        assert await verify_stream_accessible(stream_url)
        
    finally:
        await service.stop()
```

---

## 10. Tools and Utilities

### Test Execution Commands
```bash
# Quick unit tests during development
pytest tests/unit/ -x -v

# Integration tests with real components
pytest tests/integration/ --tb=short

# Full test suite with coverage
pytest tests/ --cov=src --cov-report=html

# Performance tests
pytest tests/ -m performance --timeout=300

# Specific requirement coverage
pytest tests/ -k "req_cam_001" -v
```

### Test Utilities
```python
# tests/fixtures/test_utilities.py
async def wait_for_condition(condition_func, timeout=5.0, interval=0.1):
    """Wait for condition to become true."""
    start_time = time.time()
    while time.time() - start_time < timeout:
        if await condition_func():
            return True
        await asyncio.sleep(interval)
    return False

def assert_camera_metadata_valid(metadata):
    """Validate camera metadata structure."""
    required_fields = ["device", "capabilities", "formats"]
    for field in required_fields:
        assert field in metadata, f"Missing required field: {field}"
```

---

## Questions or Clarifications?

See related documentation:
- **Development Principles:** `docs/development/principles.md`
- **Coding Standards:** `docs/development/coding-standards.md`
- **Architecture Overview:** `docs/architecture/overview.md`
- **Project Roadmap:** `docs/roadmap.md`