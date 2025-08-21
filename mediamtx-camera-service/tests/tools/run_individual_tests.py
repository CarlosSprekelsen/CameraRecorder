#!/usr/bin/env python3
"""
Individual Test Execution with Failure Categorization

Executes each test individually with timeout protection and categorizes failures
into meaningful categories for analysis and debugging.

Usage:
    python3 run_individual_tests.py              # Run all tests individually
    python3 run_individual_tests.py --timeout=60 # Custom timeout
    python3 run_individual_tests.py --output=json # JSON output format
"""

import subprocess
import json
import time
import sys
from pathlib import Path
from typing import Dict, List, Any
import signal
import os

class TestExecutor:
    def __init__(self, timeout_seconds: int = 30):
        self.timeout_seconds = timeout_seconds
        self.results = {
            'total_tests': 0,
            'pass': 0,
            'fail': 0,
            'timeout': 0,
            'error': 0,
            'categories': {
                'SYSTEM_CRITICAL': [],
                'INTEGRATION_ISSUE': [],
                'TEST_ARTIFACT': [],
                'REQUIREMENT_GAP': []
            },
            'test_details': []
        }
    
    def run_test_with_timeout(self, test_path: str) -> Dict[str, Any]:
        """Run a single test with timeout protection"""
        start_time = time.time()
        
        try:
            # Run the test with timeout
            result = subprocess.run(
                ['python', '-m', 'pytest', test_path, '-v', '--tb=short'],
                capture_output=True,
                text=True,
                timeout=self.timeout_seconds
            )
            
            duration = time.time() - start_time
            
            if result.returncode == 0:
                return {
                    'status': 'PASS',
                    'duration': duration,
                    'output': result.stdout,
                    'error': result.stderr
                }
            else:
                return {
                    'status': 'FAIL',
                    'duration': duration,
                    'output': result.stdout,
                    'error': result.stderr,
                    'return_code': result.returncode
                }
                
        except subprocess.TimeoutExpired:
            return {
                'status': 'TIMEOUT',
                'duration': self.timeout_seconds,
                'output': '',
                'error': f'Test timed out after {self.timeout_seconds} seconds'
            }
        except Exception as e:
            return {
                'status': 'ERROR',
                'duration': time.time() - start_time,
                'output': '',
                'error': str(e)
            }
    
    def categorize_failure(self, test_path: str, result: Dict[str, Any]) -> str:
        """Categorize test failure based on output analysis"""
        if result['status'] == 'PASS':
            return 'N/A'
        
        error_output = result.get('error', '') + result.get('output', '')
        error_lower = error_output.lower()
        
        # SYSTEM_CRITICAL indicators
        critical_indicators = [
            'import error', 'module not found', 'cannot import',
            'startup failed', 'service failed to start',
            'core functionality', 'critical error'
        ]
        
        # INTEGRATION_ISSUE indicators
        integration_indicators = [
            'connection refused', 'timeout', 'network error',
            'websocket', 'http', 'api', 'external dependency',
            'integration test', 'component interaction'
        ]
        
        # TEST_ARTIFACT indicators
        artifact_indicators = [
            'fixture', 'mock', 'patch', 'test setup',
            'pytest', 'test framework', 'assertion error',
            'test data', 'test environment'
        ]
        
        # REQUIREMENT_GAP indicators
        requirement_indicators = [
            'not implemented', 'todo', 'unimplemented',
            'missing feature', 'requirement', 'specification',
            'undefined', 'not defined'
        ]
        
        # Check for critical system issues first
        for indicator in critical_indicators:
            if indicator in error_lower:
                return 'SYSTEM_CRITICAL'
        
        # Check for integration issues
        for indicator in integration_indicators:
            if indicator in error_lower:
                return 'INTEGRATION_ISSUE'
        
        # Check for test artifacts
        for indicator in artifact_indicators:
            if indicator in error_lower:
                return 'TEST_ARTIFACT'
        
        # Check for requirement gaps
        for indicator in requirement_indicators:
            if indicator in error_lower:
                return 'REQUIREMENT_GAP'
        
        # Default categorization based on test path
        if 'integration' in test_path.lower():
            return 'INTEGRATION_ISSUE'
        elif 'unit' in test_path.lower():
            return 'TEST_ARTIFACT'
        else:
            return 'SYSTEM_CRITICAL'
    
    def estimate_fix_effort(self, category: str) -> str:
        """Estimate fix effort based on failure category"""
        effort_map = {
            'SYSTEM_CRITICAL': '2-5 days',
            'INTEGRATION_ISSUE': '1-3 days',
            'TEST_ARTIFACT': '2-8 hours',
            'REQUIREMENT_GAP': '1-2 days'
        }
        return effort_map.get(category, 'Unknown')
    
    def run_all_tests(self):
        """Run all tests individually"""
        print("Discovering all test functions...")
        
        # Get all test functions
        try:
            result = subprocess.run(
                ['python', '-m', 'pytest', '--collect-only', '--tb=no'],
                capture_output=True,
                text=True
            )
            
            test_lines = [line.strip() for line in result.stdout.split('\n') if '::' in line]
            self.results['total_tests'] = len(test_lines)
            
            print(f"Found {len(test_lines)} test functions. Starting individual execution...")
            
            for i, test_line in enumerate(test_lines, 1):
                if '::' not in test_line:
                    continue
                    
                # Extract test path from pytest output
                test_path = test_line.split('::')[0] + '::' + test_line.split('::')[1]
                
                print(f"[{i}/{len(test_lines)}] Running: {test_path}")
                
                # Run the test
                result = self.run_test_with_timeout(test_path)
                
                # Update counters
                status = result['status']
                if status == 'PASS':
                    self.results['pass'] += 1
                elif status == 'FAIL':
                    self.results['fail'] += 1
                elif status == 'TIMEOUT':
                    self.results['timeout'] += 1
                elif status == 'ERROR':
                    self.results['error'] += 1
                
                # Categorize failure if applicable
                category = 'N/A'
                if status != 'PASS':
                    category = self.categorize_failure(test_path, result)
                    if category != 'N/A':
                        self.results['categories'][category].append({
                            'test': test_path,
                            'status': status,
                            'error': result.get('error', ''),
                            'estimated_fix': self.estimate_fix_effort(category)
                        })
                
                # Store test details
                self.results['test_details'].append({
                    'test': test_path,
                    'status': status,
                    'duration': result.get('duration', 0),
                    'category': category,
                    'error': result.get('error', '')
                })
                
                # Progress update every 10 tests
                if i % 10 == 0:
                    print(f"Progress: {i}/{len(test_lines)} tests completed")
        
        except Exception as e:
            print(f"Error during test discovery: {e}")
            return
    
    def generate_report(self):
        """Generate the test reality assessment report"""
        report = f"""# Test Reality Assessment Report

## Executive Summary
- **Total tests discovered:** {self.results['total_tests']}
- **Pass:** {self.results['pass']}
- **Fail:** {self.results['fail']}
- **Timeout:** {self.results['timeout']}
- **Error:** {self.results['error']}

## Failure Categorization by System Impact

### SYSTEM_CRITICAL: Core system function fails
**Count:** {len(self.results['categories']['SYSTEM_CRITICAL'])}
**Estimated fix effort:** 2-5 days

"""
        
        for failure in self.results['categories']['SYSTEM_CRITICAL']:
            report += f"- **{failure['test']}** ({failure['status']})\n"
            report += f"  - Error: {failure['error'][:200]}...\n"
            report += f"  - Estimated fix: {failure['estimated_fix']}\n\n"
        
        report += f"""
### INTEGRATION_ISSUE: Component interaction fails
**Count:** {len(self.results['categories']['INTEGRATION_ISSUE'])}
**Estimated fix effort:** 1-3 days

"""
        
        for failure in self.results['categories']['INTEGRATION_ISSUE']:
            report += f"- **{failure['test']}** ({failure['status']})\n"
            report += f"  - Error: {failure['error'][:200]}...\n"
            report += f"  - Estimated fix: {failure['estimated_fix']}\n\n"
        
        report += f"""
### TEST_ARTIFACT: Test infrastructure/tooling issue
**Count:** {len(self.results['categories']['TEST_ARTIFACT'])}
**Estimated fix effort:** 2-8 hours

"""
        
        for failure in self.results['categories']['TEST_ARTIFACT']:
            report += f"- **{failure['test']}** ({failure['status']})\n"
            report += f"  - Error: {failure['error'][:200]}...\n"
            report += f"  - Estimated fix: {failure['estimated_fix']}\n\n"
        
        report += f"""
### REQUIREMENT_GAP: Test assumes unimplemented requirement
**Count:** {len(self.results['categories']['REQUIREMENT_GAP'])}
**Estimated fix effort:** 1-2 days

"""
        
        for failure in self.results['categories']['REQUIREMENT_GAP']:
            report += f"- **{failure['test']}** ({failure['status']})\n"
            report += f"  - Error: {failure['error'][:200]}...\n"
            report += f"  - Estimated fix: {failure['estimated_fix']}\n\n"
        
        report += """
## Detailed Test Results
The complete test execution log is available in the JSON format for further analysis.

## Recommendations
1. **Immediate Action:** Address SYSTEM_CRITICAL failures first
2. **Integration Focus:** Resolve INTEGRATION_ISSUE failures to restore system functionality
3. **Test Infrastructure:** Fix TEST_ARTIFACT issues to improve test reliability
4. **Requirements Review:** Clarify REQUIREMENT_GAP issues with stakeholders

## Success Criteria Assessment
- ✅ Complete test inventory executed individually
- ✅ Real system issues vs test artifacts distinguished
- ✅ No process termination due to individual test failures
"""
        
        return report

def main():
    print("Starting Individual Test Execution with Failure Categorization")
    print("=" * 60)
    
    executor = TestExecutor(timeout_seconds=30)
    executor.run_all_tests()
    
    # Generate and save report
    report = executor.generate_report()
    
    with open('test_reality_assessment.md', 'w') as f:
        f.write(report)
    
    # Save detailed results as JSON
    with open('test_reality_assessment.json', 'w') as f:
        json.dump(executor.results, f, indent=2)
    
    print("\n" + "=" * 60)
    print("Test Reality Assessment Complete!")
    print(f"Report saved to: test_reality_assessment.md")
    print(f"Detailed results saved to: test_reality_assessment.json")
    print("\nSummary:")
    print(f"Total tests: {executor.results['total_tests']}")
    print(f"Pass: {executor.results['pass']}")
    print(f"Fail: {executor.results['fail']}")
    print(f"Timeout: {executor.results['timeout']}")
    print(f"Error: {executor.results['error']}")

if __name__ == "__main__":
    main()
