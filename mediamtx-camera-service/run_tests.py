#!/usr/bin/env python3
"""
Test runner for MediaMTX Camera Service tests.

Usage:
    python run_tests.py                    # Run all tests
    python run_tests.py --unit             # Run only unit tests
    python run_tests.py --integration      # Run only integration tests
    python run_tests.py --coverage         # Run with coverage report
    python run_tests.py --specific <name>  # Run specific test
"""

import sys
import subprocess
import argparse
from pathlib import Path


def setup_test_environment():
    """Setup test environment and dependencies."""
    # Add project root to Python path
    project_root = Path(__file__).parent
    sys.path.insert(0, str(project_root))
    
    # Create test directory structure if it doesn't exist
    test_dirs = [
        "tests",
        "tests/unit", 
        "tests/integration",
        "tests/mocks"
    ]
    
    for dir_path in test_dirs:
        Path(dir_path).mkdir(exist_ok=True)
        
    # Create __init__.py files for test packages
    for dir_path in test_dirs:
        init_file = Path(dir_path) / "__init__.py"
        if not init_file.exists():
            init_file.write_text("# Test package\n")


def run_tests(args):
    """Run tests based on provided arguments."""
    setup_test_environment()
    
    # Base pytest command
    cmd = ["python", "-m", "pytest"]
    
    if args.verbose:
        cmd.append("-v")
        
    if args.coverage:
        cmd.extend([
            "--cov=src/camera_discovery",
            "--cov-report=term-missing",
            "--cov-report=html:htmlcov",
            "--cov-fail-under=70"
        ])
    
    if args.unit:
        cmd.extend(["-m", "unit"])
    elif args.integration:
        cmd.extend(["-m", "integration"])
    elif args.specific:
        cmd.extend(["-k", args.specific])
    
    if args.test_file:
        cmd.append(args.test_file)
    else:
        cmd.append("tests/")
    
    # Add extra pytest args
    if args.pytest_args:
        cmd.extend(args.pytest_args)
    
    print(f"Running: {' '.join(cmd)}")
    
    try:
        result = subprocess.run(cmd, check=False)
        return result.returncode
    except FileNotFoundError:
        print("Error: pytest not found. Install with: pip install pytest pytest-asyncio pytest-cov")
        return 1


def create_test_files():
    """Create the test files in the proper directory structure."""
    
    # Copy our comprehensive test to the tests directory
    test_content = '''"""
Comprehensive tests for hybrid camera monitor - moved from artifacts.
Run with: python3 -m pytest tests/test_hybrid_monitor_comprehensive.py -v
"""

# Import the test content here - for now just a placeholder
import pytest

def test_placeholder():
    """Placeholder test to ensure test runner works."""
    assert True

@pytest.mark.asyncio
async def test_async_placeholder():
    """Placeholder async test."""
    import asyncio
    await asyncio.sleep(0.01)
    assert True
'''
    
    test_file = Path("tests/test_hybrid_monitor_comprehensive.py")
    test_file.write_text(test_content)
    
    print(f"Created test file: {test_file}")


def main():
    parser = argparse.ArgumentParser(description="Run MediaMTX Camera Service tests")
    parser.add_argument("-v", "--verbose", action="store_true", help="Verbose output")
    parser.add_argument("--unit", action="store_true", help="Run only unit tests")
    parser.add_argument("--integration", action="store_true", help="Run only integration tests")
    parser.add_argument("--coverage", action="store_true", help="Run with coverage report")
    parser.add_argument("--specific", help="Run specific test by name pattern")
    parser.add_argument("--test-file", help="Run specific test file")
    parser.add_argument("--create-files", action="store_true", help="Create test file structure")
    parser.add_argument("pytest_args", nargs="*", help="Additional pytest arguments")
    
    args = parser.parse_args()
    
    if args.create_files:
        create_test_files()
        return 0
    
    return run_tests(args)


if __name__ == "__main__":
    sys.exit(main())