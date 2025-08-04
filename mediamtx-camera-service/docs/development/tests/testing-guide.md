# MediaMTX Camera Service - Testing Guide

**Version:** 1.0  
**Date:** 2025-08-04  
**Status:** Approved  

## Overview

This guide covers environment setup, testing procedures, and quality gates for the MediaMTX Camera Service project. All testing follows the IV&V (Independent Verification & Validation) control points defined in the project roadmap.

**Quality Gates:**
- Code formatting (black)
- Linting (flake8) 
- Type checking (mypy)
- Unit tests with 80% coverage requirement
- Integration/smoke tests
- Professional output standards (no emojis, structured logs)

## Environment Setup

### Prerequisites

- Python 3.10+ 
- Git
- Virtual environment capability

### Quick Setup

```bash
# 1. Clone and navigate to project
git clone <repository-url>
cd mediamtx-camera-service

# 2. Create and activate virtual environment
python3 -m venv venv

# Linux/macOS:
source venv/bin/activate

# Windows:
venv\Scripts\activate

# 3. Install dependencies
pip install -r requirements-dev.txt
pip install -e .

# 4. Verify setup
python3 run_all_tests.py --help
```

### Manual Setup (Alternative)

```bash
# Production dependencies
pip install -r requirements.txt

# Development and testing dependencies  
pip install pytest>=7.4.0 pytest-asyncio>=0.21.1 pytest-cov>=4.1.0
pip install black>=23.7.0 flake8>=6.0.0 mypy>=1.5.0

# Install project in development mode
pip install -e .
```

## Running All Tests

### One-Command Execution

```bash
# Run complete quality gate pipeline
python3 run_all_tests.py

# With custom coverage threshold
python3 run_all_tests.py --threshold=85

# Skip specific stages
python3 run_all_tests.py --no-lint --no-type-check

# Unit tests only
python3 run_all_tests.py --only-unit
```

### Using Makefile (Alternative)

```bash
make test-coverage    # Full test suite with coverage
make test-unit        # Unit tests only
make test-integration # Integration tests only
make lint            # Linting and type checking
make format          # Auto-format code
```

## Individual Quality Gates

### 1. Code Formatting Check

```bash
# Check formatting compliance
black --check src/ tests/

# Auto-format (if needed)
black src/ tests/
```

**Expected Output:**
```
would reformat 0 files
All done! ✨ 🍰 ✨
```

### 2. Linting

```bash
# Run flake8 linting
flake8 src/ tests/

# With configuration
flake8 --config=.flake8 src/ tests/
```

**Expected Output:**
```
(No output indicates success)
```

### 3. Type Checking

```bash
# Run mypy type checking
mypy src/

# With specific configuration
mypy --config-file=pyproject.toml src/
```

**Expected Output:**
```
Success: no issues found in X files
```

### 4. Unit Tests with Coverage

```bash
# Run unit tests with coverage measurement
pytest tests/unit/ --cov=src/camera_discovery --cov=src/camera_service \
  --cov-report=term-missing --cov-report=html:htmlcov \
  --cov-fail-under=80

# Quick unit test run
pytest tests/unit/ -v
```

**Expected Output:**
```
============================= test session starts ==============================
collected 42 items

tests/unit/test_camera_discovery/test_hybrid_monitor.py::test_placeholder PASSED
tests/unit/test_camera_service/test_service_manager.py::test_lifecycle PASSED
...

==================== 42 passed in 2.34s ====================

Name                                    Stmts   Miss  Cover   Missing
---------------------------------------------------------------------
src/camera_discovery/hybrid_monitor.py   450     45    90%    120-125
src/camera_service/service_manager.py    380     30    92%    245-250
---------------------------------------------------------------------
TOTAL                                     830     75    91%
```

### 5. Integration/Smoke Tests

```bash
# Run integration tests  
pytest tests/integration/ -v -m "not slow"

# Run smoke tests specifically
pytest tests/ -k "smoke" -v

# Run with custom markers
pytest tests/ -m "integration and not slow" -v
```

## Test Conventions and Structure

### Directory Organization

```
tests/
├── unit/                    # Fast, isolated unit tests
│   ├── test_camera_discovery/    # Camera discovery module tests
│   ├── test_camera_service/      # Service manager tests  
│   ├── test_mediamtx_wrapper/    # MediaMTX integration tests
│   ├── test_websocket_server/    # WebSocket API tests
│   └── test_common/              # Shared utilities tests
├── integration/             # Integration and end-to-end tests
│   ├── test_ivv/                 # IV&V acceptance scenarios
│   └── test_hardware/            # Hardware-dependent tests
└── mocks/                   # Test fixtures and mocks
```

### Naming Conventions

- **Test files:** `test_<module_name>.py` (snake_case)
- **Test functions:** `test_<behavior>_<scenario>()` 
- **Test classes:** `TestModuleName` (CamelCase)

**Examples:**
- `test_camera_discovery_udev_events.py`
- `test_capability_detection_varied_formats()`
- `test_service_lifecycle_error_recovery()`

### Test Categories and Markers

```python
import pytest

@pytest.mark.unit
def test_parsing_logic():
    """Fast unit test."""
    pass

@pytest.mark.integration  
def test_mediamtx_integration():
    """Integration test requiring MediaMTX."""
    pass

@pytest.mark.asyncio
async def test_async_behavior():
    """Async test."""
    pass

@pytest.mark.slow
def test_hardware_detection():
    """Slow test, skipped in quick runs."""
    pass
```

### Adding New Tests

1. **Create test file** in appropriate directory following naming convention
2. **Include docstrings** explaining test purpose and scenario
3. **Use appropriate markers** (`@pytest.mark.unit`, `@pytest.mark.integration`)
4. **Follow IV&V traceability** - link to roadmap stories in docstrings
5. **Update coverage** - ensure new code has corresponding tests

**Template:**
```python
"""
Test module for <component_name>.

Related Epic/Story: E1/S3 - Camera Discovery Implementation
IV&V Control Point: Architecture compliance verification
"""

import pytest
from src.camera_discovery.hybrid_monitor import HybridCameraMonitor

class TestHybridMonitor:
    """Test suite for hybrid camera monitor functionality."""
    
    def test_capability_detection_success(self):
        """
        Verify capability detection with valid v4l2-ctl output.
        
        Scenario: Valid camera device with standard capabilities
        Expected: Successful detection with proper metadata extraction
        """
        # Test implementation
        pass
```

## Troubleshooting

### Common Issues

**Import Errors:**
```bash
# Ensure project is installed in development mode
pip install -e .

# Add project root to Python path (if needed)
export PYTHONPATH="$PWD:$PYTHONPATH"
```

**Coverage Below Threshold:**
```bash
# Identify missing coverage
pytest --cov=src --cov-report=html:htmlcov tests/
# Open htmlcov/index.html in browser

# Run specific test file with coverage
pytest --cov=src.camera_discovery.hybrid_monitor tests/unit/test_camera_discovery/ -v
```

**Test Discovery Issues:**
```bash
# Verify test discovery
pytest --collect-only tests/

# Check for syntax errors
python3 -m py_compile tests/unit/test_*.py
```

**Virtual Environment Issues:**
```bash
# Recreate virtual environment
deactivate
rm -rf venv
python3 -m venv venv
source venv/bin/activate  # or venv\Scripts\activate on Windows
pip install -r requirements-dev.txt
```

### Expected Test Execution Times

- **Unit tests:** 5-15 seconds
- **Integration tests:** 30-60 seconds  
- **Full test suite:** 1-2 minutes
- **Coverage generation:** Additional 10-30 seconds

### Interpreting Test Results

**Success Indicators:**
- All test stages report "PASSED"
- Coverage ≥ 80% for critical modules
- No linting or type checking violations
- Clean artifacts directory with timestamped logs

**Failure Indicators:**
- Non-zero exit codes from any stage
- Coverage below threshold
- Flake8 violations in code
- MyPy type errors

## IV&V Sign-off Instructions

### Pre-Control Point Checklist

Before advancing any IV&V control point, verify:

- [ ] All quality gates pass with `python3 run_all_tests.py`
- [ ] Coverage meets or exceeds 80% threshold
- [ ] All TODO/STOP comments resolved or properly annotated
- [ ] Documentation updated for any API or behavioral changes
- [ ] No regression in existing test coverage
- [ ] Test artifacts archived with timestamp for audit trail

### Control Point Evidence Collection

```bash
# Generate comprehensive test evidence
python3 run_all_tests.py --verbose > test_evidence_$(date +%Y%m%d_%H%M%S).log

# Archive results
mkdir -p evidence/$(date +%Y%m%d)
cp -r artifacts/$(date +%Y%m%d)_* evidence/$(date +%Y%m%d)/
cp test_evidence_*.log evidence/$(date +%Y%m%d)/
```

### Story Completion Validation

For roadmap stories (S3, S4, S5), ensure:

1. **Behavior tests pass** - Core functionality validated
2. **Edge case coverage** - Error conditions and boundary cases
3. **Integration validation** - Component interaction verified  
4. **Documentation updated** - API docs, architecture compliance
5. **Reviewer sign-off** - Evidence reviewed and approved

## Advanced Usage

### Continuous Integration

```bash
# CI pipeline command
python3 run_all_tests.py --no-interactive --junit-xml=results.xml
```

### Development Workflow

```bash
# Watch mode for development (if using pytest-watch)
ptw tests/ -- --tb=short

# Quick pre-commit validation
python3 run_all_tests.py --only-unit --no-coverage

# Full pre-push validation  
python3 run_all_tests.py --threshold=80
```

### Custom Test Configurations

```bash
# Test specific components
pytest tests/unit/test_camera_discovery/ -v

# Run with custom markers
pytest -m "unit and not slow" tests/

# Debug mode with output
pytest -s -vv tests/unit/test_camera_discovery/test_hybrid_monitor.py::test_specific_function
```

---

**Questions?**  
See `docs/development/principles.md` for project guidelines and `docs/roadmap.md` for current development status.