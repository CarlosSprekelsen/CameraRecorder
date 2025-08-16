#!/usr/bin/env python3
"""
Critical Error Handling Test Validation Script

This script validates that all critical error handling tests are properly implemented
and ready for execution.

Requirements Traceability:
- REQ-ERROR-002: WebSocket server shall handle client disconnection gracefully
- REQ-ERROR-003: System shall handle MediaMTX service unavailability gracefully
- REQ-ERROR-007: System shall handle service failure scenarios with graceful degradation
- REQ-ERROR-008: System shall handle network timeout scenarios with retry mechanisms
- REQ-ERROR-009: System shall handle resource exhaustion scenarios with graceful degradation
- REQ-ERROR-010: System shall provide comprehensive edge case coverage for production reliability
"""

import os
import sys
import importlib.util
from pathlib import Path

def validate_test_file(file_path, expected_tests):
    """Validate that a test file exists and contains expected test methods."""
    print(f"Validating {file_path}...")
    
    if not os.path.exists(file_path):
        print(f"❌ ERROR: Test file not found: {file_path}")
        return False
    
    try:
        # Import the test module
        spec = importlib.util.spec_from_file_location("test_module", file_path)
        test_module = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(test_module)
        
        # Find test class
        test_class = None
        for attr_name in dir(test_module):
            attr = getattr(test_module, attr_name)
            if hasattr(attr, '__name__') and 'Test' in attr.__name__:
                test_class = attr
                break
        
        if test_class is None:
            print(f"❌ ERROR: No test class found in {file_path}")
            return False
        
        # Check for expected test methods
        missing_tests = []
        found_tests = []
        
        for test_name in expected_tests:
            if hasattr(test_class, test_name):
                found_tests.append(test_name)
            else:
                missing_tests.append(test_name)
        
        if missing_tests:
            print(f"❌ ERROR: Missing test methods in {file_path}: {missing_tests}")
            return False
        
        print(f"✅ SUCCESS: {file_path} contains {len(found_tests)} expected test methods")
        for test in found_tests:
            print(f"   - {test}")
        
        return True
        
    except Exception as e:
        print(f"❌ ERROR: Failed to validate {file_path}: {e}")
        return False

def validate_test_runner():
    """Validate the critical error test runner."""
    print("Validating test runner...")
    
    runner_path = "run_critical_error_tests.py"
    if not os.path.exists(runner_path):
        print(f"❌ ERROR: Test runner not found: {runner_path}")
        return False
    
    try:
        # Import the test runner
        spec = importlib.util.spec_from_file_location("test_runner", runner_path)
        test_runner = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(test_runner)
        
        # Check for expected functions
        expected_functions = ["main", "run_critical_error_tests"]
        missing_functions = []
        
        for func_name in expected_functions:
            if not hasattr(test_runner, func_name):
                missing_functions.append(func_name)
        
        if missing_functions:
            print(f"❌ ERROR: Missing functions in test runner: {missing_functions}")
            return False
        
        print(f"✅ SUCCESS: Test runner contains expected functions")
        return True
        
    except Exception as e:
        print(f"❌ ERROR: Failed to validate test runner: {e}")
        return False

def validate_documentation():
    """Validate that documentation files exist."""
    print("Validating documentation...")
    
    docs = [
        "tests/integration/README_CRITICAL_ERROR_HANDLING.md",
        "CRITICAL_ERROR_HANDLING_IMPLEMENTATION_SUMMARY.md"
    ]
    
    all_exist = True
    for doc_path in docs:
        if os.path.exists(doc_path):
            print(f"✅ SUCCESS: Documentation exists: {doc_path}")
        else:
            print(f"❌ ERROR: Documentation missing: {doc_path}")
            all_exist = False
    
    return all_exist

def validate_requirements_coverage():
    """Validate that all critical error handling requirements are covered."""
    print("Validating requirements coverage...")
    
    # Define expected test methods for each requirement
    requirements_coverage = {
        "REQ-ERROR-002": [
            "test_websocket_client_disconnection_scenarios",
            "test_websocket_client_disconnection_graceful_handling"
        ],
        "REQ-ERROR-003": [
            "test_mediamtx_service_unavailability_scenarios",
            "test_mediamtx_service_unavailability_graceful_handling"
        ],
        "REQ-ERROR-007": [
            "test_real_error_scenarios_and_recovery"
        ],
        "REQ-ERROR-008": [
            "test_network_failure_and_timeout_scenarios",
            "test_network_timeout_and_retry_mechanisms"
        ],
        "REQ-ERROR-009": [
            "test_resource_constraint_and_exhaustion_scenarios"
        ],
        "REQ-ERROR-010": [
            "test_error_logging_and_monitoring_validation"
        ]
    }
    
    all_covered = True
    
    for req_id, expected_tests in requirements_coverage.items():
        print(f"Checking {req_id}...")
        
        # Check if tests exist in either file
        tests_found = []
        for test_name in expected_tests:
            # Check in critical error handling file
            if os.path.exists("tests/integration/test_critical_error_handling.py"):
                spec = importlib.util.spec_from_file_location("test_module", "tests/integration/test_critical_error_handling.py")
                test_module = importlib.util.module_from_spec(spec)
                spec.loader.exec_module(test_module)
                
                for attr_name in dir(test_module):
                    attr = getattr(test_module, attr_name)
                    if hasattr(attr, '__name__') and 'Test' in attr.__name__:
                        if hasattr(attr, test_name):
                            tests_found.append(f"test_critical_error_handling.py::{test_name}")
                            break
            
            # Check in real system integration file
            if os.path.exists("tests/integration/test_real_system_integration.py"):
                spec = importlib.util.spec_from_file_location("test_module", "tests/integration/test_real_system_integration.py")
                test_module = importlib.util.module_from_spec(spec)
                spec.loader.exec_module(test_module)
                
                for attr_name in dir(test_module):
                    attr = getattr(test_module, attr_name)
                    if hasattr(attr, '__name__') and 'Test' in attr.__name__:
                        if hasattr(attr, test_name):
                            tests_found.append(f"test_real_system_integration.py::{test_name}")
                            break
        
        if tests_found:
            print(f"   ✅ {req_id} covered by: {tests_found}")
        else:
            print(f"   ❌ {req_id} not covered by any tests")
            all_covered = False
    
    return all_covered

def main():
    """Main validation function."""
    print("Critical Error Handling Test Validation")
    print("=" * 50)
    
    # Change to the correct directory
    script_dir = Path(__file__).parent
    os.chdir(script_dir)
    
    validation_results = []
    
    # Validate test files
    print("\n1. Validating Test Files")
    print("-" * 30)
    
    # Critical error handling test file
    critical_error_tests = [
        "test_network_failure_and_timeout_scenarios",
        "test_mediamtx_service_unavailability_scenarios",
        "test_websocket_client_disconnection_scenarios",
        "test_resource_constraint_and_exhaustion_scenarios",
        "test_error_logging_and_monitoring_validation"
    ]
    
    result1 = validate_test_file(
        "tests/integration/test_critical_error_handling.py",
        critical_error_tests
    )
    validation_results.append(("Critical Error Handling Test File", result1))
    
    # Enhanced integration test file
    enhanced_tests = [
        "test_websocket_client_disconnection_graceful_handling",
        "test_mediamtx_service_unavailability_graceful_handling",
        "test_network_timeout_and_retry_mechanisms"
    ]
    
    result2 = validate_test_file(
        "tests/integration/test_real_system_integration.py",
        enhanced_tests
    )
    validation_results.append(("Enhanced Integration Test File", result2))
    
    # Validate test runner
    print("\n2. Validating Test Runner")
    print("-" * 30)
    result3 = validate_test_runner()
    validation_results.append(("Test Runner", result3))
    
    # Validate documentation
    print("\n3. Validating Documentation")
    print("-" * 30)
    result4 = validate_documentation()
    validation_results.append(("Documentation", result4))
    
    # Validate requirements coverage
    print("\n4. Validating Requirements Coverage")
    print("-" * 30)
    result5 = validate_requirements_coverage()
    validation_results.append(("Requirements Coverage", result5))
    
    # Print summary
    print("\n" + "=" * 50)
    print("VALIDATION SUMMARY")
    print("=" * 50)
    
    passed = 0
    total = len(validation_results)
    
    for component, result in validation_results:
        status = "✅ PASS" if result else "❌ FAIL"
        print(f"{status} - {component}")
        if result:
            passed += 1
    
    print(f"\nOverall Result: {passed}/{total} components validated successfully")
    
    if passed == total:
        print("🎉 ALL VALIDATIONS PASSED - Critical error handling tests are ready!")
        print("\nNext Steps:")
        print("1. Run tests: python run_critical_error_tests.py")
        print("2. Review documentation: tests/integration/README_CRITICAL_ERROR_HANDLING.md")
        print("3. Execute individual tests as needed")
        return True
    else:
        print("⚠️  SOME VALIDATIONS FAILED - Please review and fix issues")
        return False

if __name__ == "__main__":
    success = main()
    sys.exit(0 if success else 1)
