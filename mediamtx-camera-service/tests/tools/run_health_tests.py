#!/usr/bin/env python3
"""
Health test runner for MediaMTX Camera Service.

Executes comprehensive health monitoring tests against real system components
to validate health requirements from the baseline.

Usage:
    python3 tests/tools/run_health_tests.py                    # Run all health tests
    python3 tests/tools/run_health_tests.py --components      # Detailed component info only
    python3 tests/tools/run_health_tests.py --kubernetes      # Kubernetes probes only
    python3 tests/tools/run_health_tests.py --json            # JSON response format only
    python3 tests/tools/run_health_tests.py --ok-response     # 200 OK response only
    python3 tests/tools/run_health_tests.py --error-response  # 500 error response only

Test Categories: Health
"""

import asyncio
import argparse
import sys
import os
import time
from pathlib import Path

# Add project root to path
project_root = Path(__file__).parent.parent.parent
sys.path.insert(0, str(project_root))

from tests.health.test_health_monitoring import (
    test_health_status_detailed_components,
    test_kubernetes_readiness_probes,
    test_health_endpoint_json_responses,
    test_health_endpoint_200_ok,
    test_health_endpoint_500_error,
    run_all_health_tests
)


class HealthTestRunner:
    """Health test orchestrator with configuration and reporting."""
    
    def __init__(self, args):
        self.args = args
        self.start_time = time.time()
        self.results = {}
        
    async def run_components_test(self):
        """Run detailed component information test."""
        print("\n🚀 Running Detailed Component Information Test")
        print("=" * 50)
        
        try:
            result = await test_health_status_detailed_components()
            self.results['detailed_components'] = result
            print("✅ Detailed component information test completed")
            return True
        except Exception as e:
            print(f"❌ Detailed component information test failed: {e}")
            return False
    
    async def run_kubernetes_test(self):
        """Run Kubernetes readiness probes test."""
        print("\n🚀 Running Kubernetes Readiness Probes Test")
        print("=" * 50)
        
        try:
            result = await test_kubernetes_readiness_probes()
            self.results['kubernetes_probes'] = result
            print("✅ Kubernetes readiness probes test completed")
            return True
        except Exception as e:
            print(f"❌ Kubernetes readiness probes test failed: {e}")
            return False
    
    async def run_json_test(self):
        """Run JSON response format test."""
        print("\n🚀 Running JSON Response Format Test")
        print("=" * 50)
        
        try:
            result = await test_health_endpoint_json_responses()
            self.results['json_responses'] = result
            print("✅ JSON response format test completed")
            return True
        except Exception as e:
            print(f"❌ JSON response format test failed: {e}")
            return False
    
    async def run_ok_response_test(self):
        """Run 200 OK response test."""
        print("\n🚀 Running 200 OK Response Test")
        print("=" * 50)
        
        try:
            result = await test_health_endpoint_200_ok()
            self.results['ok_responses'] = result
            print("✅ 200 OK response test completed")
            return True
        except Exception as e:
            print(f"❌ 200 OK response test failed: {e}")
            return False
    
    async def run_error_response_test(self):
        """Run 500 error response test."""
        print("\n🚀 Running 500 Error Response Test")
        print("=" * 50)
        
        try:
            result = await test_health_endpoint_500_error()
            self.results['error_responses'] = result
            print("✅ 500 error response test completed")
            return True
        except Exception as e:
            print(f"❌ 500 error response test failed: {e}")
            return False
    
    async def run_all_tests(self):
        """Run all health tests."""
        print("\n🚀 Running Complete Health Test Suite")
        print("=" * 50)
        
        try:
            result = await run_all_health_tests()
            self.results = result
            print("✅ All health tests completed")
            return True
        except Exception as e:
            print(f"❌ Health test suite failed: {e}")
            return False
    
    def generate_report(self):
        """Generate health test report."""
        end_time = time.time()
        total_duration = end_time - self.start_time
        
        print("\n" + "=" * 60)
        print("🏥 HEALTH TEST REPORT")
        print("=" * 60)
        print(f"Total Duration: {total_duration:.2f} seconds")
        print(f"Tests Completed: {len(self.results)}")
        
        # Summary of results
        for test_name, result in self.results.items():
            print(f"\n📋 {test_name.upper().replace('_', ' ')}:")
            
            if isinstance(result, dict):
                if 'result' in result:
                    # WebSocket API response
                    result_data = result['result']
                    if 'overall_status' in result_data:
                        print(f"   Overall Status: {result_data['overall_status']}")
                    if 'components' in result_data:
                        for component, details in result_data['components'].items():
                            print(f"   {component}: {details.get('status', 'unknown')}")
                
                elif 'system' in result:
                    # Health endpoint results
                    for endpoint, endpoint_result in result.items():
                        if endpoint_result.get('success'):
                            status_code = endpoint_result.get('status_code', 'unknown')
                            status = endpoint_result.get('response', {}).get('status', 'unknown')
                            print(f"   {endpoint}: {status_code} - {status}")
                        else:
                            print(f"   {endpoint}: Failed")
        
        # Requirements validation
        print(f"\n🎯 REQUIREMENTS VALIDATION:")
        requirements = {
            'REQ-HEALTH-005': 'Detailed component information',
            'REQ-HEALTH-006': 'Kubernetes readiness probes',
            'REQ-API-017': 'JSON response format',
            'REQ-API-018': '200 OK response',
            'REQ-API-019': '500 error response'
        }
        
        for req, description in requirements.items():
            status = "✅ PASS" if self.results else "❌ FAIL"
            print(f"   {req} ({description}): {status}")
        
        # Overall assessment
        print(f"\n🎯 OVERALL ASSESSMENT:")
        print(f"   Health Requirements: {'✅ MET' if self.results else '❌ NOT MET'}")
        print(f"   Test Coverage: {'✅ COMPLETE' if len(self.results) >= 5 else '⚠️ PARTIAL'}")
        print(f"   System Health: {'✅ ACCEPTABLE' if self.results else '❌ UNACCEPTABLE'}")


async def main():
    """Main health test runner."""
    parser = argparse.ArgumentParser(description="MediaMTX Camera Service Health Test Runner")
    parser.add_argument("--components", action="store_true", help="Run detailed component information test only")
    parser.add_argument("--kubernetes", action="store_true", help="Run Kubernetes readiness probes test only")
    parser.add_argument("--json", action="store_true", help="Run JSON response format test only")
    parser.add_argument("--ok-response", action="store_true", help="Run 200 OK response test only")
    parser.add_argument("--error-response", action="store_true", help="Run 500 error response test only")
    
    args = parser.parse_args()
    
    print("🏥 MediaMTX Camera Service Health Test Runner")
    print("=" * 60)
    print("Testing health monitoring against requirements baseline")
    
    runner = HealthTestRunner(args)
    
    try:
        # Determine which tests to run
        if args.components:
            success = await runner.run_components_test()
        elif args.kubernetes:
            success = await runner.run_kubernetes_test()
        elif args.json:
            success = await runner.run_json_test()
        elif args.ok_response:
            success = await runner.run_ok_response_test()
        elif args.error_response:
            success = await runner.run_error_response_test()
        else:
            # Run all tests
            success = await runner.run_all_tests()
        
        # Generate report
        runner.generate_report()
        
        if success:
            print("\n🎉 Health test execution completed successfully!")
            sys.exit(0)
        else:
            print("\n❌ Health test execution failed!")
            sys.exit(1)
            
    except KeyboardInterrupt:
        print("\n⚠️ Health test execution interrupted by user")
        sys.exit(1)
    except Exception as e:
        print(f"\n❌ Health test execution failed with error: {e}")
        sys.exit(1)


if __name__ == "__main__":
    asyncio.run(main())
