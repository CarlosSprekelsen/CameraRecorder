# Issue 089: Test Infrastructure Working Directory Bug

**Status:** 🚨 CRITICAL  
**Priority:** HIGH  
**Type:** Test Infrastructure Bug  
**Date:** 2025-08-24  
**Assigned:** Test Infrastructure Team  

## **Summary**

The test runner (`run_all_tests.py`) is executing from the wrong working directory, causing all test stages to fail with "No such file or directory" errors. This is a critical test infrastructure issue that prevents the complete test suite from running properly.

## **Impact**

- **All test stages fail** with return code 127 (command not found)
- **Type checking fails** due to incorrect working directory
- **Unit tests fail** due to missing venv/bin/python3 path
- **Integration tests fail** due to missing venv/bin/python3 path
- **Formatting fails** due to missing venv/bin/python3 path
- **Linting fails** due to missing flake8 path

## **Root Cause Analysis**

### **Working Directory Issue**
The test runner is executing from `/home/carlossprekelsen/CameraRecorder/mediamtx-camera-service/tests/tools/` instead of the project root directory. This causes:

1. **Relative path resolution failures**: Commands like `venv/bin/python3` resolve to the wrong location
2. **Source directory not found**: `mypy src/` fails because `src/` is not found from the tools directory
3. **Virtual environment path issues**: The venv directory is not found relative to the working directory

### **Command Execution Context**
```bash
# Current (BROKEN):
Working Directory: /home/carlossprekelsen/CameraRecorder/mediamtx-camera-service/tests/tools/
Command: venv/bin/python3 -m pytest tests/unit/
Result: [Errno 2] No such file or directory: 'venv/bin/python3'

# Expected (WORKING):
Working Directory: /home/carlossprekelsen/CameraRecorder/mediamtx-camera-service/
Command: venv/bin/python3 -m pytest tests/unit/
Result: Tests execute properly
```

## **Evidence**

### **Test Report Data**
```json
{
  "environment": {
    "working_directory": "/home/carlossprekelsen/CameraRecorder/mediamtx-camera-service/tests/tools"
  },
  "stages": [
    {
      "name": "Type Checking",
      "status": "FAILED",
      "return_code": 2
    },
    {
      "name": "Unit Tests", 
      "status": "FAILED",
      "return_code": 127
    },
    {
      "name": "Integration Tests",
      "status": "FAILED", 
      "return_code": 127
    },
    {
      "name": "Formatting",
      "status": "FAILED",
      "return_code": 127
    },
    {
      "name": "Linting",
      "status": "FAILED",
      "return_code": 127
    }
  ]
}
```

### **Error Messages**
- `[Errno 2] No such file or directory: 'venv/bin/python3'`
- `mypy: can't read file 'src': No such file or directory`
- `[Errno 2] No such file or directory: 'flake8'`

## **Affected Components**

- **Test Runner**: `tests/tools/run_all_tests.py`
- **Type Checking**: mypy configuration and execution
- **Unit Tests**: pytest execution
- **Integration Tests**: pytest execution  
- **Code Quality**: black and flake8 execution

## **Proposed Solution**

### **Fix Working Directory in Test Runner**
Modify `tests/tools/run_all_tests.py` to ensure proper working directory while maintaining test guidelines compliance:

```python
def __init__(self, args: argparse.Namespace):
    self.args = args
    # FIX: Use project root, not tools directory
    # Maintain test guidelines: tools stay in tests/tools/, but execute from project root
    self.tools_dir = Path(__file__).parent  # tests/tools/
    self.project_root = self.tools_dir.parent.parent  # project root
    self.artifacts_dir = self._create_artifacts_dir()
    self.stages: List[TestStage] = []
    self.overall_start_time = time.time()
    
    # Detect virtual environment
    self.venv_active = self._detect_virtual_environment()
```

### **Fix Command Execution Context**
Ensure all commands are executed from the project root directory while maintaining tool organization:

```python
def _run_command(self, cmd: List[str], cwd: Optional[Path] = None, capture_output: bool = True):
    if cwd is None:
        cwd = self.project_root  # Always use project root as default
    # Tools remain in tests/tools/ but commands execute from project root
```

### **Fix Virtual Environment Path Resolution**
Use absolute paths for virtual environment tools while respecting test structure:

```python
def _get_venv_tool_path(self, tool_name: str) -> str:
    """Get absolute path to virtual environment tool."""
    return str(self.project_root / "venv" / "bin" / tool_name)
```

## **Testing Plan**

1. **Verify Working Directory**: Ensure test runner executes from project root
2. **Test Command Resolution**: Verify all tool paths resolve correctly
3. **Run Complete Test Suite**: Execute full test suite to confirm all stages pass
4. **Validate Artifacts**: Ensure test artifacts are created in correct location

## **Acceptance Criteria**

- [ ] Test runner executes from project root directory
- [ ] All test stages (type checking, unit tests, integration tests, formatting, linting) pass
- [ ] Virtual environment tools are found and executed correctly
- [ ] Test artifacts are created in the expected location
- [ ] No "No such file or directory" errors occur

## **Related Issues**

- **Issue 090**: Virtual Environment Incomplete Setup
- **Issue 091**: Test Infrastructure PATH Configuration

## **Notes**

This is a critical infrastructure issue that blocks all testing activities. The fix should be implemented immediately to restore test functionality.

**Priority:** Must be fixed before any other testing can proceed.
