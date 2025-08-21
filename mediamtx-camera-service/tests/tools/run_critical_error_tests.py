#!/usr/bin/env python3
"""
Critical Error Handling Test Runner

This script runs the critical error handling tests that focus on failure scenarios
that could break the system during PDR.

This test runner focuses on:
- Network failures and timeouts
- Service unavailability scenarios
- Resource constraints and exhaustion
- Graceful degradation and recovery mechanisms
- Error logging and monitoring validation

Usage:
    python3 run_critical_error_tests.py          # Run all critical error tests
    python3 run_critical_error_tests.py --timeout=180 # Custom timeout
    python3 run_critical_error_tests.py --retries=3   # Custom retry count
"""

import asyncio
import logging
import os
import sys
import time
from pathlib import Path

# Add src to path
sys.path.insert(0, str(Path(__file__).parent / "src"))

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


async def run_critical_error_tests():
    """Run critical error handling tests."""
    
    logger.info("Starting Critical Error Handling Test Suite")
    logger.info("=" * 60)
    
    # Test configuration
    test_config = {
        "timeout": 180,  # 3 minutes per test
        "retries": 3,    # Retry failed tests
        "parallel": False  # Run tests sequentially for stability
    }
    
    # Test results tracking
    test_results = {
        "passed": 0,
        "failed": 0,
        "skipped": 0,
        "total": 0
    }
    
    # Test suite definition
    test_suite = [
        {
            "name": "Network Failure and Timeout Scenarios",
            "file": "tests/integration/test_critical_error_handling.py",
            "test": "test_network_failure_and_timeout_scenarios",
            "description": "REQ-ERROR-008: Network timeout scenarios with retry mechanisms"
        },
        {
            "name": "MediaMTX Service Unavailability Scenarios",
            "file": "tests/integration/test_critical_error_handling.py",
            "test": "test_mediamtx_service_unavailability_scenarios",
            "description": "REQ-ERROR-003: MediaMTX service unavailability scenarios"
        },
        {
            "name": "WebSocket Client Disconnection Scenarios",
            "file": "tests/integration/test_critical_error_handling.py",
            "test": "test_websocket_client_disconnection_scenarios",
            "description": "REQ-ERROR-002: WebSocket client disconnection scenarios"
        },
        {
            "name": "Resource Constraint and Exhaustion Scenarios",
            "file": "tests/integration/test_critical_error_handling.py",
            "test": "test_resource_constraint_and_exhaustion_scenarios",
            "description": "REQ-ERROR-009: Resource exhaustion scenarios with graceful degradation"
        },
        {
            "name": "Error Logging and Monitoring Validation",
            "file": "tests/integration/test_critical_error_handling.py",
            "test": "test_error_logging_and_monitoring_validation",
            "description": "REQ-ERROR-010: Comprehensive edge case coverage for production reliability"
        },
        {
            "name": "WebSocket Client Disconnection Graceful Handling",
            "file": "tests/integration/test_real_system_integration.py",
            "test": "test_websocket_client_disconnection_graceful_handling",
            "description": "REQ-ERROR-002: WebSocket server shall handle client disconnection gracefully"
        },
        {
            "name": "MediaMTX Service Unavailability Graceful Handling",
            "file": "tests/integration/test_real_system_integration.py",
            "test": "test_mediamtx_service_unavailability_graceful_handling",
            "description": "REQ-ERROR-003: System shall handle MediaMTX service unavailability gracefully"
        },
        {
            "name": "Network Timeout and Retry Mechanisms",
            "file": "tests/integration/test_real_system_integration.py",
            "test": "test_network_timeout_and_retry_mechanisms",
            "description": "REQ-ERROR-008: System shall handle network timeout scenarios with retry mechanisms"
        }
    ]
    
    logger.info(f"Running {len(test_suite)} critical error handling tests")
    logger.info("")
    
    # Run each test
    for i, test_info in enumerate(test_suite, 1):
        logger.info(f"Test {i}/{len(test_suite)}: {test_info['name']}")
        logger.info(f"Description: {test_info['description']}")
        logger.info(f"File: {test_info['file']}")
        logger.info(f"Test: {test_info['test']}")
        logger.info("-" * 60)
        
        test_results["total"] += 1
        
        # Check if test file exists
        test_file_path = Path(__file__).parent / test_info["file"]
        if not test_file_path.exists():
            logger.error(f"Test file not found: {test_file_path}")
            test_results["failed"] += 1
            continue
        
        # Run test with pytest
        try:
            start_time = time.time()
            
            # Import and run test
            import pytest
            import importlib.util
            
            # Import test module
            spec = importlib.util.spec_from_file_location(
                "test_module", 
                test_file_path
            )
            test_module = importlib.util.module_from_spec(spec)
            spec.loader.exec_module(test_module)
            
            # Find test class and method
            test_class = None
            test_method = None
            
            for attr_name in dir(test_module):
                attr = getattr(test_module, attr_name)
                if hasattr(attr, '__name__') and 'Test' in attr.__name__:
                    if hasattr(attr, test_info["test"]):
                        test_class = attr
                        test_method = getattr(attr, test_info["test"])
                        break
            
            if test_class is None or test_method is None:
                logger.error(f"Test method {test_info['test']} not found in {test_info['file']}")
                test_results["failed"] += 1
                continue
            
            # Run test
            logger.info(f"Executing test: {test_info['test']}")
            
            # Create test instance and run
            test_instance = test_class()
            
            # Run test with timeout
            try:
                await asyncio.wait_for(
                    test_method(test_instance),
                    timeout=test_config["timeout"]
                )
                
                elapsed_time = time.time() - start_time
                logger.info(f"✅ PASSED - {test_info['name']} (took {elapsed_time:.2f}s)")
                test_results["passed"] += 1
                
            except asyncio.TimeoutError:
                logger.error(f"⏰ TIMEOUT - {test_info['name']} (exceeded {test_config['timeout']}s)")
                test_results["failed"] += 1
                
            except Exception as e:
                logger.error(f"❌ FAILED - {test_info['name']}: {e}")
                test_results["failed"] += 1
                
        except Exception as e:
            logger.error(f"❌ ERROR - {test_info['name']}: {e}")
            test_results["failed"] += 1
        
        logger.info("")
    
    # Print summary
    logger.info("=" * 60)
    logger.info("CRITICAL ERROR HANDLING TEST SUMMARY")
    logger.info("=" * 60)
    logger.info(f"Total Tests: {test_results['total']}")
    logger.info(f"Passed: {test_results['passed']}")
    logger.info(f"Failed: {test_results['failed']}")
    logger.info(f"Skipped: {test_results['skipped']}")
    
    if test_results['failed'] == 0:
        logger.info("🎉 ALL CRITICAL ERROR HANDLING TESTS PASSED")
        logger.info("✅ System handles critical failure modes gracefully")
        return True
    else:
        logger.error(f"❌ {test_results['failed']} CRITICAL ERROR HANDLING TESTS FAILED")
        logger.error("⚠️  System may not handle critical failure modes gracefully")
        return False


def main():
    """Main entry point."""
    logger.info("Critical Error Handling Test Runner")
    logger.info("Focus: Critical error conditions that could break system during PDR")
    logger.info("Target: Network failures, service unavailability, resource constraints")
    logger.info("")
    
    # Check prerequisites
    logger.info("Checking prerequisites...")
    
    # Check if MediaMTX service is available
    try:
        import subprocess
        result = subprocess.run(
            ["systemctl", "is-active", "mediamtx"],
            capture_output=True,
            text=True,
            check=False
        )
        if result.returncode == 0:
            logger.info("✅ MediaMTX service is active")
        else:
            logger.warning("⚠️  MediaMTX service is not active (some tests may fail)")
    except FileNotFoundError:
        logger.warning("⚠️  systemctl not available (some tests may fail)")
    
    # Check if required Python packages are available
    required_packages = ["pytest", "aiohttp", "websockets", "asyncio"]
    missing_packages = []
    
    for package in required_packages:
        try:
            __import__(package)
            logger.info(f"✅ {package} is available")
        except ImportError:
            missing_packages.append(package)
            logger.error(f"❌ {package} is not available")
    
    if missing_packages:
        logger.error(f"Missing required packages: {', '.join(missing_packages)}")
        logger.error("Please install missing packages before running tests")
        return False
    
    logger.info("")
    
    # Run tests
    try:
        success = asyncio.run(run_critical_error_tests())
        return success
    except KeyboardInterrupt:
        logger.info("Test execution interrupted by user")
        return False
    except Exception as e:
        logger.error(f"Test execution failed: {e}")
        return False


if __name__ == "__main__":
    success = main()
    sys.exit(0 if success else 1)
