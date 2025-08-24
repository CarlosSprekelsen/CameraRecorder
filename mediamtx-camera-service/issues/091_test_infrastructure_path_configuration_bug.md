# Issue 091: Test Infrastructure PATH Configuration Bug

**Status:** 🚨 CRITICAL  
**Priority:** HIGH  
**Type:** Test Infrastructure Bug  
**Date:** 2025-08-24  
**Assigned:** Test Infrastructure Team  

## **Summary**

The test runner (`run_all_tests.py`) uses `shutil.which(tool)` to find development tools in the system PATH, but does not properly configure the PATH to include the virtual environment's bin directory. This causes all tool lookups to fail even when tools are installed in the virtual environment.

## **Impact**

- **All tool lookups fail** - `shutil.which()` cannot find tools in virtual environment
- **Test runner aborts early** - Missing tools prevent test execution
- **Manual PATH manipulation required** - Need to manually set PATH for each test run
- **Inconsistent behavior** - Different results depending on system PATH configuration
- **CI/CD failures** - Automated testing fails due to tool discovery issues

## **Root Cause Analysis**

### **PATH Configuration Issue**
The test runner validates required tools using `shutil.which(tool)` but does not ensure the virtual environment's bin directory is in the PATH:

```python
# Current (BROKEN):
required_tools = ["black", "flake8", "mypy", "pytest"]
missing_tools = []

for tool in required_tools:
    if not shutil.which(tool):  # Uses system PATH only
        missing_tools.append(tool)
```

### **Virtual Environment PATH Not Set**
The test runner detects virtual environment but doesn't configure PATH accordingly:

```python
def _detect_virtual_environment(self) -> bool:
    """Detect if running in virtual environment."""
    return (
        hasattr(sys, 'real_prefix') or
        (hasattr(sys, 'base_prefix') and sys.base_prefix != sys.prefix) or
        os.environ.get('VIRTUAL_ENV') is not None
    )
    # FIXME: Detects virtual environment but doesn't configure PATH
```

## **Evidence**

### **Tool Discovery Failure**
```bash
# Tools installed in virtual environment:
$ ls -la venv/bin/ | grep -E "(flake8|black|mypy|pytest)"
-rwxrwxr-x 1 carlossprekelsen carlossprekelsen  282 ago 24 14:18 black
-rwxrwxr-x 1 carlossprekelsen carlossprekelsen  276 ago 24 14:18 flake8
-rwxrwxr-x 1 carlossprekelsen carlossprekelsen  292 ago 24 14:18 mypy
-rwxrwxr-x 1 carlossprekelsen carlossprekelsen  283 ago 24 14:18 pytest

# But shutil.which() cannot find them:
$ python3 -c "import shutil; print(shutil.which('flake8'))"
None
```

### **Test Runner Error**
```
ERROR: Missing required tools: flake8
Install with: python -m pip install -r requirements-dev.txt
```

### **Manual PATH Fix Works**
```bash
# Manual PATH manipulation works:
$ PATH="venv/bin:$PATH" python3 -c "import shutil; print(shutil.which('flake8'))"
/home/carlossprekelsen/CameraRecorder/mediamtx-camera-service/venv/bin/flake8
```

## **Affected Components**

- **Test Runner**: `tests/tools/run_all_tests.py`
- **Tool Validation**: `shutil.which()` calls for tool discovery
- **Test Environment Setup**: PATH configuration for virtual environment
- **CI/CD Pipeline**: Automated testing and tool discovery

## **Proposed Solution**

### **Fix PATH Configuration in Test Runner**
Modify the test runner to properly configure PATH when virtual environment is detected:

```python
def _setup_environment_path(self):
    """Setup environment PATH to include virtual environment tools."""
    if self.venv_active:
        venv_bin = self.project_root / "venv" / "bin"
        if venv_bin.exists():
            # Add virtual environment bin to PATH
            current_path = os.environ.get('PATH', '')
            if str(venv_bin) not in current_path:
                os.environ['PATH'] = f"{venv_bin}:{current_path}"
                print(f"Added {venv_bin} to PATH")
```

### **Update Tool Validation**
Modify tool validation to use configured PATH:

```python
def _validate_required_tools(self) -> bool:
    """Validate required tools are available."""
    # Setup environment PATH first
    self._setup_environment_path()
    
    required_tools = ["black", "flake8", "mypy", "pytest"]
    missing_tools = []
    
    for tool in required_tools:
        tool_path = shutil.which(tool)
        if not tool_path:
            missing_tools.append(tool)
        else:
            print(f"Found {tool} at: {tool_path}")
    
    if missing_tools:
        print(f"ERROR: Missing required tools: {', '.join(missing_tools)}")
        print("Install with: python -m pip install -r requirements-dev.txt")
        return False
    
    return True
```

### **Add Environment Validation**
Add validation to ensure environment is properly configured:

```python
def _validate_environment(self) -> bool:
    """Validate test environment is properly configured."""
    # Check virtual environment
    if not self.venv_active:
        print("WARNING: Not running in virtual environment")
    
    # Check PATH configuration
    venv_bin = self.project_root / "venv" / "bin"
    if venv_bin.exists() and str(venv_bin) not in os.environ.get('PATH', ''):
        print("WARNING: Virtual environment not in PATH")
    
    return True
```

## **Testing Plan**

1. **Test PATH Configuration**: Verify virtual environment bin is added to PATH
2. **Test Tool Discovery**: Verify all required tools are found
3. **Test Command Execution**: Verify tools can be executed from test runner
4. **Test Environment Validation**: Verify environment validation works correctly
5. **Run Complete Test Suite**: Execute full test suite to confirm functionality

## **Acceptance Criteria**

- [ ] Virtual environment bin directory is automatically added to PATH
- [ ] All required tools (flake8, black, mypy, pytest) are discovered correctly
- [ ] Test runner can execute tools without manual PATH manipulation
- [ ] Environment validation provides clear feedback about configuration
- [ ] No "Missing required tools" errors occur

## **Related Issues**

- **Issue 089**: Test Infrastructure Working Directory Bug
- **Issue 090**: Virtual Environment Incomplete Setup Bug

## **Notes**

This issue is critical for test infrastructure reliability. The fix ensures consistent tool discovery across different environments and eliminates the need for manual PATH manipulation.

**Priority:** Must be fixed to enable reliable automated testing.
