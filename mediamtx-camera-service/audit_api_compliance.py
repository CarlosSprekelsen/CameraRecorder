#!/usr/bin/env python3
"""
API Compliance Audit Script

This script audits all test files to ensure they follow the API documentation
as ground truth instead of adapting to server implementation.

Critical Rules:
1. Tests must use API documentation as source of truth
2. Tests must validate against documented API format
3. Tests must NOT adapt to broken implementations
4. Tests must fail if API documentation is violated
"""

import os
import re
import sys
from pathlib import Path
from typing import List, Dict, Any, Tuple

# API Documentation references
API_DOCS = "docs/api/json-rpc-methods.md"
HEALTH_DOCS = "docs/api/health-endpoints.md"

# Known API methods from documentation
DOCUMENTED_API_METHODS = [
    "authenticate",
    "ping",
    "get_camera_list",
    "get_camera_status",
    "take_snapshot",
    "start_recording",
    "stop_recording",
    "list_recordings",
    "list_snapshots",
    "get_recording_info",
    "get_snapshot_info",
    "delete_recording",
    "delete_snapshot",
    "get_storage_info",
    "set_retention_policy",
    "cleanup_old_files",
    "get_metrics",
    "get_status",
    "get_server_info"
]

# Known health endpoints from documentation
DOCUMENTED_HEALTH_ENDPOINTS = [
    "/health/system",
    "/health/cameras",
    "/health/mediamtx"
]

class APIAuditResult:
    """Result of API compliance audit."""
    
    def __init__(self, file_path: str):
        self.file_path = file_path
        self.violations = []
        self.warnings = []
        self.api_methods_used = []
        self.health_endpoints_used = []
        self.implementation_references = []
        
    def add_violation(self, line_num: int, message: str):
        """Add a compliance violation."""
        self.violations.append((line_num, message))
        
    def add_warning(self, line_num: int, message: str):
        """Add a compliance warning."""
        self.warnings.append((line_num, message))
        
    def add_api_method(self, method: str, line_num: int):
        """Add an API method usage."""
        self.api_methods_used.append((method, line_num))
        
    def add_health_endpoint(self, endpoint: str, line_num: int):
        """Add a health endpoint usage."""
        self.health_endpoints_used.append((endpoint, line_num))
        
    def add_implementation_reference(self, reference: str, line_num: int):
        """Add an implementation reference."""
        self.implementation_references.append((reference, line_num))
        
    @property
    def has_violations(self) -> bool:
        """Check if there are any violations."""
        return len(self.violations) > 0
        
    @property
    def has_warnings(self) -> bool:
        """Check if there are any warnings."""
        return len(self.warnings) > 0

def audit_file(file_path: str) -> APIAuditResult:
    """Audit a single test file for API compliance."""
    result = APIAuditResult(file_path)
    
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            content = f.read()
            lines = content.split('\n')
    except Exception as e:
        result.add_violation(0, f"Could not read file: {e}")
        return result
    
    # Check for API documentation reference
    if not re.search(r'json-rpc-methods\.md', content, re.IGNORECASE):
        if any(method in content for method in DOCUMENTED_API_METHODS):
            result.add_violation(0, "Missing API documentation reference in docstring")
    
    # Check for health endpoints documentation reference
    if not re.search(r'health-endpoints\.md', content, re.IGNORECASE):
        if any(endpoint in content for endpoint in DOCUMENTED_HEALTH_ENDPOINTS):
            result.add_violation(0, "Missing health endpoints documentation reference in docstring")
    
    # Check for implementation-specific testing
    implementation_patterns = [
        r'server\.py',
        r'websocket_server',
        r'_method_',
        r'register_method',
        r'security_middleware',
        r'auth_manager',
        r'ServiceManager',
        r'WebSocketJsonRpcServer'
    ]
    
    for i, line in enumerate(lines, 1):
        # Check for API method usage
        for method in DOCUMENTED_API_METHODS:
            if method in line and 'method' in line.lower():
                result.add_api_method(method, i)
                
        # Check for health endpoint usage
        for endpoint in DOCUMENTED_HEALTH_ENDPOINTS:
            if endpoint in line:
                result.add_health_endpoint(endpoint, i)
                
        # Check for implementation references
        for pattern in implementation_patterns:
            if re.search(pattern, line, re.IGNORECASE):
                result.add_implementation_reference(pattern, i)
                
        # Check for hardcoded response expectations that might not match API docs
        if re.search(r'assert.*"result"', line):
            # Check if this is testing a documented response format
            if not any(method in line for method in DOCUMENTED_API_METHODS):
                result.add_warning(i, "Response assertion without clear API method context")
                
        # Check for authentication flow violations
        if 'authenticate' in line.lower() and 'auth_token' in line:
            # Check if this follows documented authentication flow
            if not re.search(r'params.*auth_token', line):
                result.add_violation(i, "Authentication not following documented parameter format")
                
        # Check for error code hardcoding
        if re.search(r'-3200[1-9]', line):
            # Check if this error code is documented
            result.add_warning(i, "Hardcoded error code - verify against API documentation")
            
    return result

def audit_test_directory(directory: str = "tests") -> List[APIAuditResult]:
    """Audit all test files in the directory."""
    results = []
    
    for root, dirs, files in os.walk(directory):
        for file in files:
            if file.endswith('.py') and file.startswith('test_'):
                file_path = os.path.join(root, file)
                result = audit_file(file_path)
                results.append(result)
                
    return results

def print_audit_report(results: List[APIAuditResult]):
    """Print a comprehensive audit report."""
    total_files = len(results)
    files_with_violations = sum(1 for r in results if r.has_violations)
    files_with_warnings = sum(1 for r in results if r.has_warnings)
    total_violations = sum(len(r.violations) for r in results)
    total_warnings = sum(len(r.warnings) for r in results)
    
    print("=" * 80)
    print("🚨 API COMPLIANCE AUDIT REPORT")
    print("=" * 80)
    print(f"Total test files audited: {total_files}")
    print(f"Files with violations: {files_with_violations}")
    print(f"Files with warnings: {files_with_warnings}")
    print(f"Total violations: {total_violations}")
    print(f"Total warnings: {total_warnings}")
    print()
    
    if files_with_violations == 0:
        print("✅ All test files comply with API documentation requirements!")
    else:
        print("❌ CRITICAL VIOLATIONS FOUND:")
        print()
        
        for result in results:
            if result.has_violations:
                print(f"📁 {result.file_path}")
                for line_num, message in result.violations:
                    print(f"   Line {line_num}: {message}")
                print()
                
    if files_with_warnings > 0:
        print("⚠️  WARNINGS:")
        print()
        
        for result in results:
            if result.has_warnings:
                print(f"📁 {result.file_path}")
                for line_num, message in result.warnings:
                    print(f"   Line {line_num}: {message}")
                print()
    
    # Summary of API methods used
    all_api_methods = []
    for result in results:
        all_api_methods.extend(result.api_methods_used)
    
    if all_api_methods:
        print("📊 API METHODS USED IN TESTS:")
        method_counts = {}
        for method, _ in all_api_methods:
            method_counts[method] = method_counts.get(method, 0) + 1
        
        for method, count in sorted(method_counts.items()):
            print(f"   {method}: {count} usages")
        print()
    
    # Implementation references
    all_impl_refs = []
    for result in results:
        all_impl_refs.extend(result.implementation_references)
    
    if all_impl_refs:
        print("🔍 IMPLEMENTATION REFERENCES FOUND:")
        impl_counts = {}
        for ref, _ in all_impl_refs:
            impl_counts[ref] = impl_counts.get(ref, 0) + 1
        
        for ref, count in sorted(impl_counts.items()):
            print(f"   {ref}: {count} references")
        print()
    
    print("=" * 80)
    print("RECOMMENDATIONS:")
    print("1. Fix all violations to ensure tests use API documentation as ground truth")
    print("2. Review warnings to ensure proper API compliance")
    print("3. Remove implementation-specific testing")
    print("4. Add API documentation references to all test files")
    print("5. Validate all response formats against API documentation")
    print("=" * 80)

def main():
    """Main audit function."""
    print("🔍 Starting API Compliance Audit...")
    print("Checking all test files against API documentation...")
    print()
    
    results = audit_test_directory()
    print_audit_report(results)
    
    # Return exit code based on violations
    total_violations = sum(len(r.violations) for r in results)
    if total_violations > 0:
        print(f"\n❌ Audit failed with {total_violations} violations")
        return 1
    else:
        print("\n✅ Audit passed - all tests comply with API documentation")
        return 0

if __name__ == "__main__":
    sys.exit(main())
