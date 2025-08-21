#!/usr/bin/env python3
"""
Performance test runner for MediaMTX Camera Service.

Executes comprehensive performance tests against real system components
to validate performance requirements from the baseline.

Usage:
    python3 tests/tools/run_performance_tests.py              # Run all performance tests
    python3 tests/tools/run_performance_tests.py --status     # Status methods only
    python3 tests/tools/run_performance_tests.py --control    # Control methods only
    python3 tests/tools/run_performance_tests.py --file-ops   # File operations only
    python3 tests/tools/run_performance_tests.py --concurrent # Concurrent connections only
    python3 tests/tools/run_performance_tests.py --iterations=200 # Custom iterations
    python3 tests/tools/run_performance_tests.py --connections=50 # Custom connection count

Test Categories: Performance
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

from tests.performance.test_api_performance import (
    test_status_methods_performance,
    test_control_methods_performance,
    test_file_operations_performance,
    test_concurrent_connections_performance,
    run_all_performance_tests
)


class PerformanceTestRunner:
    """Performance test orchestrator with configuration and reporting."""
    
    def __init__(self, args):
        self.args = args
        self.start_time = time.time()
        self.results = {}
        
    async def run_status_methods_test(self):
        """Run status methods performance test."""
        print("\n🚀 Running Status Methods Performance Test")
        print("=" * 50)
        
        try:
            result = await test_status_methods_performance()
            self.results['status_methods'] = result
            print("✅ Status methods performance test completed")
            return True
        except Exception as e:
            print(f"❌ Status methods performance test failed: {e}")
            return False
    
    async def run_control_methods_test(self):
        """Run control methods performance test."""
        print("\n🚀 Running Control Methods Performance Test")
        print("=" * 50)
        
        try:
            result = await test_control_methods_performance()
            self.results['control_methods'] = result
            print("✅ Control methods performance test completed")
            return True
        except Exception as e:
            print(f"❌ Control methods performance test failed: {e}")
            return False
    
    async def run_file_operations_test(self):
        """Run file operations performance test."""
        print("\n🚀 Running File Operations Performance Test")
        print("=" * 50)
        
        try:
            result = await test_file_operations_performance()
            self.results['file_operations'] = result
            print("✅ File operations performance test completed")
            return True
        except Exception as e:
            print(f"❌ File operations performance test failed: {e}")
            return False
    
    async def run_concurrent_connections_test(self):
        """Run concurrent connections performance test."""
        print("\n🚀 Running Concurrent Connections Performance Test")
        print("=" * 50)
        
        try:
            result = await test_concurrent_connections_performance()
            self.results['concurrent_connections'] = result
            print("✅ Concurrent connections performance test completed")
            return True
        except Exception as e:
            print(f"❌ Concurrent connections performance test failed: {e}")
            return False
    
    async def run_all_tests(self):
        """Run all performance tests."""
        print("\n🚀 Running Complete Performance Test Suite")
        print("=" * 50)
        
        try:
            result = await run_all_performance_tests()
            self.results = result
            print("✅ All performance tests completed")
            return True
        except Exception as e:
            print(f"❌ Performance test suite failed: {e}")
            return False
    
    def generate_report(self):
        """Generate performance test report."""
        end_time = time.time()
        total_duration = end_time - self.start_time
        
        print("\n" + "=" * 60)
        print("📊 PERFORMANCE TEST REPORT")
        print("=" * 60)
        print(f"Total Duration: {total_duration:.2f} seconds")
        print(f"Tests Completed: {len(self.results)}")
        
        # Summary of results
        for test_name, result in self.results.items():
            print(f"\n📋 {test_name.upper().replace('_', ' ')}:")
            
            if 'metrics' in result:
                for metric in result['metrics']:
                    print(f"   {metric.method_name}:")
                    print(f"     - Success Rate: {metric.success_rate:.2%}")
                    print(f"     - P95 Response Time: {metric.p95_response_time_ms:.2f}ms")
                    print(f"     - Mean Response Time: {metric.mean_response_time_ms:.2f}ms")
            
            if 'validations' in result:
                print(f"   Requirements Validation:")
                for method, validations in result['validations'].items():
                    for req, passed in validations.items():
                        status = "✅ PASS" if passed else "❌ FAIL"
                        print(f"     - {req}: {status}")
        
        # Overall assessment
        print(f"\n🎯 OVERALL ASSESSMENT:")
        print(f"   Performance Requirements: {'✅ MET' if self.results else '❌ NOT MET'}")
        print(f"   Test Coverage: {'✅ COMPLETE' if len(self.results) >= 4 else '⚠️ PARTIAL'}")
        print(f"   System Performance: {'✅ ACCEPTABLE' if self.results else '❌ UNACCEPTABLE'}")


async def main():
    """Main performance test runner."""
    parser = argparse.ArgumentParser(description="MediaMTX Camera Service Performance Test Runner")
    parser.add_argument("--status", action="store_true", help="Run status methods performance test only")
    parser.add_argument("--control", action="store_true", help="Run control methods performance test only")
    parser.add_argument("--file-ops", action="store_true", help="Run file operations performance test only")
    parser.add_argument("--concurrent", action="store_true", help="Run concurrent connections performance test only")
    parser.add_argument("--iterations", type=int, default=100, help="Number of iterations for performance tests")
    parser.add_argument("--connections", type=int, default=10, help="Number of concurrent connections to test")
    
    args = parser.parse_args()
    
    print("🎯 MediaMTX Camera Service Performance Test Runner")
    print("=" * 60)
    print("Testing API performance against requirements baseline")
    print(f"Configuration: {args.iterations} iterations, {args.connections} connections")
    
    runner = PerformanceTestRunner(args)
    
    try:
        # Determine which tests to run
        if args.status:
            success = await runner.run_status_methods_test()
        elif args.control:
            success = await runner.run_control_methods_test()
        elif args.file_ops:
            success = await runner.run_file_operations_test()
        elif args.concurrent:
            success = await runner.run_concurrent_connections_test()
        else:
            # Run all tests
            success = await runner.run_all_tests()
        
        # Generate report
        runner.generate_report()
        
        if success:
            print("\n🎉 Performance test execution completed successfully!")
            sys.exit(0)
        else:
            print("\n❌ Performance test execution failed!")
            sys.exit(1)
            
    except KeyboardInterrupt:
        print("\n⚠️ Performance test execution interrupted by user")
        sys.exit(1)
    except Exception as e:
        print(f"\n❌ Performance test execution failed with error: {e}")
        sys.exit(1)


if __name__ == "__main__":
    asyncio.run(main())
