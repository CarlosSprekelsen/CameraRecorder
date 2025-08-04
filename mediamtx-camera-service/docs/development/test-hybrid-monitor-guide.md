# Test Execution Guide - MediaMTX Camera Service

## Quick Start (Immediate Testing)

### 1. Run Core Logic Tests (No Dependencies)
```bash
# Run immediately - tests core parsing logic without async/mocking
python3 tests/test_hybrid_monitor_core.py
```

This validates:
- ✅ Frame rate extraction from v4l2-ctl output patterns
- ✅ Hierarchical frame rate selection logic  
- ✅ Stream name extraction robustness
- ✅ Capability validation state transitions

### 2. Setup Full Test Environment
```bash
# Install test dependencies
pip install pytest pytest-asyncio pytest-cov

# Create test directory structure
python3 run_tests.py --create-files

# Run quick smoke tests
make test-quick
```

### 3. Run Comprehensive Test Suite
```bash
# Run all tests with coverage
make test-coverage

# Run specific test categories
make test-unit          # Unit tests only
make test-integration   # Integration tests only

# Run specific test file
python3 -m pytest tests/test_hybrid_monitor_comprehensive.py -v
```

## Test Categories

### Core Logic Tests (✅ Ready)
**File**: `tests/test_hybrid_monitor_core.py`
- **Status**: Immediately runnable, no dependencies
- **Coverage**: Parsing logic, selection algorithms, state transitions
- **Runtime**: <1 second

### Comprehensive Tests (📋 Scaffold Created)  
**File**: `tests/test_hybrid_monitor_comprehensive.py`
- **Status**: Full test suite with async/mocking infrastructure
- **Coverage**: End-to-end workflows, error handling, integration
- **Runtime**: 10-30 seconds

### Integration Tests (🚧 Planned)
**Files**: `tests/integration/test_*`
- **Status**: Planned for real hardware validation
- **Coverage**: Actual v4l2 devices, MediaMTX integration
- **Runtime**: Variable (depends on hardware)

## Test Execution Methods

### Method 1: Make Targets (Recommended)
```bash
make test              # Full test suite
make test-unit         # Unit tests only  
make test-integration  # Integration tests only
make test-coverage     # With coverage report
make test-quick        # Fast smoke tests
```

### Method 2: Direct pytest
```bash
# Specific test file
pytest tests/test_hybrid_monitor_core.py -v

# Specific test function
pytest -k "test_frame_rate_extraction" -v

# With coverage
pytest --cov=src/camera_discovery --cov-report=html tests/
```

### Method 3: Custom Test Runner
```bash
# Using provided test runner
python3 run_tests.py --coverage --verbose
python3 run_tests.py --specific "frame_rate"
python3 run_tests.py --test-file tests/test_hybrid_monitor_core.py
```

## Continuous Testing During Development

### Watch Mode (Recommended for Development)
```bash
# Re-run tests on file changes
make test-watch

# Or manually
pytest --tb=short -q --disable-warnings --looponfail tests/
```

### Pre-commit Testing
```bash
# Quick validation before commits
make test-quick && make lint

# Full validation before push
make test-coverage && make lint
```

## Test Output Examples

### Core Tests Success
```
MediaMTX Camera Service - Hybrid Monitor Core Tests
============================================================
Testing frame rate extraction patterns...
  ✓ '30.000 fps' → {'30'}
  ✓ 'Interval: [1/30]' → {'30'}
  ✓ 'Multiple: 30.000 fps, 25 FPS' → {'30', '25'}
Frame rate extraction: 8 passed, 0 failed

Testing hierarchical frame rate selection...
  ✓ Highest stable rate (30) should be first
  ✓ Stable rates should come before unstable
Selection result: ['30', '25', '60', '15', '10']
Hierarchical selection: 4 passed, 0 failed

SUMMARY:
Total test groups: 4
Passed: 4
Failed: 0
🎉 ALL TESTS PASSED!
```

### Coverage Report
```
Name                                    Stmts   Miss  Cover   Missing
---------------------------------------------------------------------
src/camera_discovery/hybrid_monitor.py   450     45    90%    120-125, 340-345
src/camera_discovery/__init__.py            0      0   100%
---------------------------------------------------------------------
TOTAL                                     450     45    90%
```

## Adding New Tests

### 1. Core Logic Tests (Fast, No Dependencies)
Add to `tests/test_hybrid_monitor_core.py`:
```python
def test_new_parsing_logic():
    """Test new parsing functionality."""
    monitor = HybridCameraMonitor()
    result = monitor.new_parsing_method("test input")
    assert result == expected_output
```

### 2. Async/Integration Tests  
Add to `tests/test_hybrid_monitor_comprehensive.py`:
```python
@pytest.mark.asyncio
async def test_new_async_behavior(monitor):
    """Test new async functionality."""
    result = await monitor.new_async_method()
    assert result.success
```

### 3. Test Naming Convention
- **Format**: `test_<module>_<behavior>_<scenario>`
- **Examples**:
  - `test_capability_parsing_varied_formats`
  - `test_udev_event_filtering_out_of_range`
  - `test_adaptive_polling_frequency_adjustment`

## Troubleshooting

### Import Errors
```bash
# Add project root to PYTHONPATH
export PYTHONPATH="$PWD:$PYTHONPATH"

# Or run from project root
cd /path/to/mediamtx-camera-service
python3 tests/test_hybrid_monitor_core.py
```

### Missing Dependencies
```bash
# Install test dependencies
pip install -r requirements-dev.txt

# Or minimal test setup
pip install pytest pytest-asyncio
```

### Test Discovery Issues
```bash
# Verify test discovery
pytest --collect-only tests/

# Check pytest configuration
pytest --help | grep -A5 "config file"
```

## Validation Checklist

Before declaring S3 complete, ensure:

- [ ] Core logic tests pass (immediate validation)
- [ ] Comprehensive test suite runs successfully
- [ ] Coverage >80% for critical path functions
- [ ] All TODO/STOP comments resolved or properly annotated
- [ ] Error handling tests verify structured error output
- [ ] Capability detection handles varied v4l2-ctl outputs
- [ ] Udev event processing filters correctly
- [ ] Adaptive polling adjusts based on event freshness
- [ ] Frame rate selection follows hierarchical policy

## Next Steps

1. **Execute core tests** to validate current implementation
2. **Expand comprehensive tests** with real device scenarios
3. **Add integration tests** for hardware validation
4. **Setup CI pipeline** for automated testing
5. **Document edge cases** discovered during testing

---

**Test-Driven Validation**: These tests lock in the correctness of the enhanced hybrid monitor, enabling safe evolution and regression detection for the solo developer workflow.