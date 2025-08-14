#!/usr/bin/env python3
"""
Real System Integration Test Runner

This script runs the real system integration tests with proper setup and teardown.
It ensures all dependencies are available and handles test environment preparation.
"""

import asyncio
import logging
import os
import subprocess
import sys
import tempfile
from pathlib import Path

# Add project root to path
project_root = Path(__file__).parent.parent.parent
sys.path.insert(0, str(project_root))

from tests.integration.test_real_system_integration import TestRealSystemIntegration

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)


class RealIntegrationTestRunner:
    """Runner for real system integration tests."""
    
    def __init__(self):
        self.temp_dir = None
        self.dependencies_checked = False
    
    async def check_dependencies(self) -> bool:
        """Check if all required dependencies are available."""
        logger.info("Checking dependencies for real integration tests...")
        
        dependencies = {
            "mediamtx": "MediaMTX server",
            "ffmpeg": "FFmpeg for video processing",
            "python3": "Python 3 interpreter"
        }
        
        missing_deps = []
        
        for cmd, description in dependencies.items():
            try:
                result = subprocess.run([cmd, "--version"], 
                                      capture_output=True, text=True, timeout=10)
                if result.returncode == 0:
                    logger.info(f"✓ {description} found")
                else:
                    missing_deps.append(description)
            except (subprocess.TimeoutExpired, FileNotFoundError):
                missing_deps.append(description)
        
        if missing_deps:
            logger.error(f"Missing dependencies: {', '.join(missing_deps)}")
            logger.error("Please install missing dependencies before running tests")
            return False
        
        self.dependencies_checked = True
        logger.info("All dependencies available")
        return True
    
    async def setup_test_environment(self) -> bool:
        """Set up test environment."""
        logger.info("Setting up test environment...")
        
        # Create temporary directory for tests
        self.temp_dir = tempfile.mkdtemp(prefix="real_integration_runner_")
        logger.info(f"Test environment created: {self.temp_dir}")
        
        # Set environment variables for tests
        os.environ["TEST_TEMP_DIR"] = self.temp_dir
        os.environ["PYTHONPATH"] = f"{project_root}:{os.environ.get('PYTHONPATH', '')}"
        
        return True
    
    async def cleanup_test_environment(self) -> None:
        """Clean up test environment."""
        if self.temp_dir and os.path.exists(self.temp_dir):
            import shutil
            shutil.rmtree(self.temp_dir)
            logger.info("Test environment cleaned up")
    
    async def run_single_test(self, test_name: str) -> bool:
        """Run a single integration test."""
        logger.info(f"Running test: {test_name}")
        
        try:
            # Create test instance
            test_instance = TestRealSystemIntegration()
            
            # Run the specific test
            if test_name == "test_real_mediamtx_server_startup_and_health":
                await test_instance.test_real_mediamtx_server_startup_and_health(
                    await test_instance.real_mediamtx_server(test_instance.test_config()),
                    test_instance.test_config()
                )
            elif test_name == "test_real_camera_discovery_and_stream_creation":
                await test_instance.test_real_camera_discovery_and_stream_creation(
                    await test_instance.service_manager(test_instance.test_config()),
                    await test_instance.test_video_streams(test_instance.test_config()),
                    await test_instance.websocket_client(test_instance.test_config())
                )
            elif test_name == "test_real_recording_and_snapshot_operations":
                await test_instance.test_real_recording_and_snapshot_operations(
                    await test_instance.service_manager(test_instance.test_config()),
                    await test_instance.test_video_streams(test_instance.test_config()),
                    await test_instance.websocket_client(test_instance.test_config())
                )
            elif test_name == "test_real_websocket_authentication_and_control":
                await test_instance.test_real_websocket_authentication_and_control(
                    await test_instance.service_manager(test_instance.test_config()),
                    await test_instance.websocket_client(test_instance.test_config())
                )
            elif test_name == "test_real_error_scenarios_and_recovery":
                await test_instance.test_real_error_scenarios_and_recovery(
                    await test_instance.service_manager(test_instance.test_config()),
                    await test_instance.real_mediamtx_server(test_instance.test_config()),
                    await test_instance.websocket_client(test_instance.test_config())
                )
            elif test_name == "test_real_resource_management_and_cleanup":
                await test_instance.test_real_resource_management_and_cleanup(
                    await test_instance.service_manager(test_instance.test_config()),
                    test_instance.test_config()
                )
            elif test_name == "test_real_end_to_end_camera_lifecycle":
                await test_instance.test_real_end_to_end_camera_lifecycle(
                    await test_instance.service_manager(test_instance.test_config()),
                    await test_instance.test_video_streams(test_instance.test_config()),
                    await test_instance.websocket_client(test_instance.test_config())
                )
            else:
                logger.error(f"Unknown test: {test_name}")
                return False
            
            logger.info(f"✓ Test {test_name} passed")
            return True
            
        except Exception as e:
            logger.error(f"✗ Test {test_name} failed: {e}")
            return False
    
    async def run_all_tests(self) -> bool:
        """Run all real integration tests."""
        logger.info("Running all real integration tests...")
        
        # Check dependencies first
        if not await self.check_dependencies():
            return False
        
        # Set up test environment
        if not await self.setup_test_environment():
            return False
        
        try:
            # List of all tests to run
            tests = [
                "test_real_mediamtx_server_startup_and_health",
                "test_real_camera_discovery_and_stream_creation",
                "test_real_recording_and_snapshot_operations",
                "test_real_websocket_authentication_and_control",
                "test_real_error_scenarios_and_recovery",
                "test_real_resource_management_and_cleanup",
                "test_real_end_to_end_camera_lifecycle"
            ]
            
            passed = 0
            failed = 0
            
            for test_name in tests:
                if await self.run_single_test(test_name):
                    passed += 1
                else:
                    failed += 1
            
            logger.info(f"Test results: {passed} passed, {failed} failed")
            return failed == 0
            
        finally:
            await self.cleanup_test_environment()


async def main():
    """Main function to run real integration tests."""
    import argparse
    
    parser = argparse.ArgumentParser(description="Run real system integration tests")
    parser.add_argument("--test", help="Run specific test")
    parser.add_argument("--all", action="store_true", help="Run all tests")
    parser.add_argument("--check-deps", action="store_true", help="Check dependencies only")
    
    args = parser.parse_args()
    
    runner = RealIntegrationTestRunner()
    
    if args.check_deps:
        success = await runner.check_dependencies()
        sys.exit(0 if success else 1)
    
    if args.test:
        # Check dependencies
        if not await runner.check_dependencies():
            sys.exit(1)
        
        # Set up environment
        if not await runner.setup_test_environment():
            sys.exit(1)
        
        try:
            success = await runner.run_single_test(args.test)
            sys.exit(0 if success else 1)
        finally:
            await runner.cleanup_test_environment()
    
    elif args.all:
        success = await runner.run_all_tests()
        sys.exit(0 if success else 1)
    
    else:
        parser.print_help()
        sys.exit(1)


if __name__ == "__main__":
    asyncio.run(main())
