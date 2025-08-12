#!/usr/bin/env python3
"""
Test Environment Validation Script

This script validates that the test environment is properly configured
and consistent across different development environments.
"""

import os
import sys
import tempfile
import subprocess
from pathlib import Path

def check_python_version():
    """Check Python version compatibility."""
    version = sys.version_info
    if version.major != 3 or version.minor < 10:
        print(f"❌ Python version {version.major}.{version.minor} is not supported. Required: 3.10+")
        return False
    print(f"✅ Python version {version.major}.{version.minor}.{version.micro} is compatible")
    return True

def check_required_packages():
    """Check required packages are installed."""
    required_packages = [
        "pytest",
        "pytest-asyncio", 
        "pytest-cov",
        "aiohttp",
        "pyyaml",
        "cryptography"
    ]
    
    missing_packages = []
    for package in required_packages:
        try:
            __import__(package.replace("-", "_"))
            print(f"✅ {package} is installed")
        except ImportError:
            missing_packages.append(package)
            print(f"❌ {package} is missing")
    
    return len(missing_packages) == 0

def check_test_directories():
    """Check test directories are accessible."""
    test_dirs = [
        "/tmp/test_recordings",
        "/tmp/test_snapshots", 
        "/tmp/test_logs"
    ]
    
    for test_dir in test_dirs:
        path = Path(test_dir)
        try:
            path.mkdir(parents=True, exist_ok=True)
            # Test write access
            test_file = path / "test_write.tmp"
            test_file.write_text("test")
            test_file.unlink()
            print(f"✅ {test_dir} is accessible and writable")
        except Exception as e:
            print(f"❌ {test_dir} is not accessible: {e}")
            return False
    
    return True

def check_environment_variables():
    """Check required environment variables are set."""
    required_vars = [
        "CAMERA_SERVICE_JWT_SECRET",
        "CAMERA_SERVICE_RATE_RPM",
        "CAMERA_SERVICE_TEST_MODE",
        "CAMERA_SERVICE_DISABLE_HARDWARE"
    ]
    
    missing_vars = []
    for var in required_vars:
        if var not in os.environ:
            missing_vars.append(var)
            print(f"❌ {var} is not set")
        else:
            print(f"✅ {var} is set")
    
    return len(missing_vars) == 0

def check_network_access():
    """Check network access for test endpoints."""
    test_endpoints = [
        ("127.0.0.1", 8002),  # WebSocket
        ("127.0.0.1", 8003),  # Health
        ("127.0.0.1", 9997),  # MediaMTX API
    ]
    
    import socket
    for host, port in test_endpoints:
        try:
            sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            sock.settimeout(1)
            result = sock.connect_ex((host, port))
            sock.close()
            if result == 0:
                print(f"⚠️  {host}:{port} is already in use (may interfere with tests)")
            else:
                print(f"✅ {host}:{port} is available for testing")
        except Exception as e:
            print(f"❌ Cannot check {host}:{port}: {e}")
            return False
    
    return True

def check_pytest_configuration():
    """Check pytest configuration is valid."""
    try:
        result = subprocess.run(
            ["python3", "-m", "pytest", "--collect-only", "-q"],
            capture_output=True,
            text=True,
            timeout=30
        )
        if result.returncode == 0:
            print("✅ pytest configuration is valid")
            return True
        else:
            print(f"❌ pytest configuration error: {result.stderr}")
            return False
    except subprocess.TimeoutExpired:
        print("❌ pytest configuration check timed out")
        return False
    except Exception as e:
        print(f"❌ pytest configuration check failed: {e}")
        return False

def check_test_isolation():
    """Check test isolation by running a simple test."""
    with tempfile.NamedTemporaryFile(mode='w', suffix='.py', delete=False) as f:
        f.write("""
import pytest

def test_isolation():
    assert True
""")
        temp_test_file = f.name
    
    try:
        result = subprocess.run(
            ["python3", "-m", "pytest", temp_test_file, "-v"],
            capture_output=True,
            text=True,
            timeout=10
        )
        if result.returncode == 0:
            print("✅ Test isolation is working")
            return True
        else:
            print(f"❌ Test isolation failed: {result.stderr}")
            return False
    except Exception as e:
        print(f"❌ Test isolation check failed: {e}")
        return False
    finally:
        os.unlink(temp_test_file)

def main():
    """Main validation function."""
    print("🔍 Validating Test Environment...")
    print("=" * 50)
    
    checks = [
        ("Python Version", check_python_version),
        ("Required Packages", check_required_packages),
        ("Test Directories", check_test_directories),
        ("Environment Variables", check_environment_variables),
        ("Network Access", check_network_access),
        ("Pytest Configuration", check_pytest_configuration),
        ("Test Isolation", check_test_isolation),
    ]
    
    passed = 0
    total = len(checks)
    
    for name, check_func in checks:
        print(f"\n📋 {name}:")
        try:
            if check_func():
                passed += 1
            else:
                print(f"   ❌ {name} check failed")
        except Exception as e:
            print(f"   ❌ {name} check error: {e}")
    
    print("\n" + "=" * 50)
    print(f"📊 Validation Results: {passed}/{total} checks passed")
    
    if passed == total:
        print("🎉 Test environment is properly configured!")
        return 0
    else:
        print("⚠️  Some checks failed. Please review the issues above.")
        return 1

if __name__ == "__main__":
    sys.exit(main())
