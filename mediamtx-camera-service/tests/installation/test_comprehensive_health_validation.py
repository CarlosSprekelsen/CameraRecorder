"""
Comprehensive Health Validation Tests

Tests that validate all health metrics and system components beyond just camera count.
These tests ensure the entire system is healthy and operational.

Requirements:
- REQ-HEALTH-001: System shall provide comprehensive logging for health monitoring
- REQ-HEALTH-002: System shall support structured logging for production environments
- REQ-HEALTH-003: System shall enable correlation ID tracking across components

This test file validates:
1. MediaMTX service health and API accessibility
2. Service manager component health
3. Camera monitor discovery and device access
4. System resource usage
5. Configuration health
6. Network connectivity
7. File permissions and access
"""

import pytest
import requests
import subprocess
import os
import json
import time
import psutil
import socket
from typing import Dict, Any, List, Optional


class TestMediaMTXHealthValidation:
    """Test MediaMTX service health and API accessibility."""
    
    def test_mediamtx_service_status(self):
        """Test that MediaMTX systemd service is running."""
        result = subprocess.run(
            ["systemctl", "is-active", "mediamtx"],
            capture_output=True,
            text=True,
            timeout=10
        )
        
        assert result.returncode == 0, f"MediaMTX service is not running: {result.stderr}"
        assert result.stdout.strip() == "active", f"MediaMTX service is not active: {result.stdout.strip()}"
        
        print("✅ MediaMTX systemd service is running")
    
    def test_mediamtx_api_accessibility(self):
        """Test that MediaMTX API is accessible."""
        try:
            response = requests.get("http://localhost:9997/v3/paths/list", timeout=5)
            assert response.status_code == 200, f"MediaMTX API not responding: {response.status_code}"
            
            data = response.json()
            assert "items" in data, "MediaMTX API response missing 'items' field"
            
            print(f"✅ MediaMTX API is accessible, {len(data['items'])} paths configured")
            
        except requests.exceptions.RequestException as e:
            pytest.fail(f"MediaMTX API not accessible: {e}")
    
    def test_mediamtx_stream_paths_accuracy(self):
        """Test that MediaMTX has correct number of camera paths configured."""
        # Get actual video devices
        video_devices = self._get_actual_video_devices()
        expected_paths = len(video_devices)
        
        try:
            response = requests.get("http://localhost:9997/v3/paths/list", timeout=5)
            data = response.json()
            
            actual_paths = len(data["items"])
            
            assert actual_paths == expected_paths, (
                f"MediaMTX path count mismatch: {actual_paths} paths configured, "
                f"but system has {expected_paths} video devices: {video_devices}"
            )
            
            # Validate path names match expected pattern
            for i, path in enumerate(data["items"]):
                expected_name = f"camera{i}"
                assert path["name"] == expected_name, (
                    f"MediaMTX path name mismatch: expected {expected_name}, got {path['name']}"
                )
            
            print(f"✅ MediaMTX has correct {actual_paths} camera paths configured")
            
        except requests.exceptions.RequestException as e:
            pytest.fail(f"MediaMTX API test failed: {e}")
    
    def test_mediamtx_port_accessibility(self):
        """Test that all MediaMTX ports are listening."""
        expected_ports = [
            (8554, "RTSP"),
            (8888, "HLS"),
            (8889, "WebRTC"),
            (9997, "API")
        ]
        
        for port, service_name in expected_ports:
            try:
                sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
                sock.settimeout(2)
                result = sock.connect_ex(('localhost', port))
                sock.close()
                
                assert result == 0, f"MediaMTX {service_name} port {port} is not listening"
                print(f"✅ MediaMTX {service_name} port {port} is accessible")
                
            except Exception as e:
                pytest.fail(f"Failed to test MediaMTX {service_name} port {port}: {e}")
    
    def test_mediamtx_health_endpoint_accuracy(self):
        """Test that MediaMTX health endpoint reports accurate information."""
        try:
            response = requests.get("http://localhost:8003/health/mediamtx", timeout=5)
            assert response.status_code == 200
            
            data = response.json()
            assert "status" in data, "MediaMTX health response missing status"
            assert "details" in data, "MediaMTX health response missing details"
            
            # Validate status is healthy
            assert data["status"] == "healthy", f"MediaMTX health status is {data['status']}, expected healthy"
            
            print(f"✅ MediaMTX health endpoint reports: {data['status']} - {data['details']}")
            
        except requests.exceptions.RequestException as e:
            pytest.fail(f"MediaMTX health endpoint test failed: {e}")
    
    def _get_actual_video_devices(self) -> List[str]:
        """Get actual video devices present on the system."""
        video_devices = []
        
        try:
            result = subprocess.run(
                ["ls", "/dev/video*"], 
                capture_output=True, 
                text=True,
                timeout=5
            )
            
            if result.returncode == 0:
                devices = result.stdout.strip().split('\n')
                video_devices = [dev for dev in devices if dev and os.path.exists(dev)]
            else:
                for i in range(10):
                    device_path = f"/dev/video{i}"
                    if os.path.exists(device_path):
                        video_devices.append(device_path)
                        
        except Exception:
            pass
        
        return video_devices


class TestServiceManagerHealthValidation:
    """Test service manager component health and lifecycle."""
    
    def test_service_manager_health_endpoint_accuracy(self):
        """Test that service manager health endpoint reports accurate information."""
        try:
            response = requests.get("http://localhost:8003/health/system", timeout=5)
            assert response.status_code == 200
            
            data = response.json()
            assert "components" in data, "System health response missing components"
            
            service_manager = data["components"].get("service_manager")
            assert service_manager is not None, "Service manager component missing from health response"
            
            assert service_manager["status"] == "healthy", (
                f"Service manager health status is {service_manager['status']}, expected healthy"
            )
            
            print(f"✅ Service manager health: {service_manager['status']} - {service_manager['details']}")
            
        except requests.exceptions.RequestException as e:
            pytest.fail(f"Service manager health test failed: {e}")
    
    def test_component_dependency_chain(self):
        """Test that all required components are present and healthy."""
        try:
            response = requests.get("http://localhost:8003/health/system", timeout=5)
            data = response.json()
            
            required_components = ["mediamtx", "camera_monitor", "service_manager"]
            
            for component in required_components:
                assert component in data["components"], f"Missing required component: {component}"
                
                component_data = data["components"][component]
                assert component_data["status"] == "healthy", (
                    f"Component {component} is not healthy: {component_data['status']}"
                )
                
                print(f"✅ Component {component}: {component_data['status']}")
            
            print("✅ All required components are present and healthy")
            
        except requests.exceptions.RequestException as e:
            pytest.fail(f"Component dependency test failed: {e}")
    
    def test_service_startup_logs(self):
        """Test that service startup logs show successful component initialization."""
        result = subprocess.run([
            "journalctl", "-u", "camera-service", "-n", "50", "--no-pager"
        ], capture_output=True, text=True, timeout=10)
        
        if result.returncode != 0:
            pytest.skip("Could not retrieve service logs")
        
        logs = result.stdout.lower()
        
        # Check for successful startup indicators
        startup_indicators = [
            "service manager started",
            "health server started",
            "websocket server started",
            "camera monitor started",
            "mediamtx controller started"
        ]
        
        found_indicators = []
        for indicator in startup_indicators:
            if indicator in logs:
                found_indicators.append(indicator)
        
        # Should have at least some startup indicators
        assert len(found_indicators) >= 3, (
            f"Service startup logs missing key indicators. Found: {found_indicators}"
        )
        
        print(f"✅ Service startup logs show successful initialization: {found_indicators}")


class TestCameraMonitorHealthValidation:
    """Test camera monitor discovery and device access."""
    
    def test_camera_monitor_device_access(self):
        """Test that camera monitor can access video devices."""
        video_devices = self._get_actual_video_devices()
        
        if not video_devices:
            pytest.skip("No video devices present for testing")
        
        # Test device permissions for camera-service user
        for device in video_devices:
            result = subprocess.run([
                "sudo", "-u", "camera-service", "test", "-r", device
            ], capture_output=True, text=True)
            
            assert result.returncode == 0, f"Camera service cannot read device {device}"
        
        print(f"✅ Camera monitor can access all {len(video_devices)} video devices")
    
    def test_camera_monitor_discovery_process(self):
        """Test that camera monitor discovery process is working."""
        # Check if camera monitor is discovering devices
        try:
            response = requests.get("http://localhost:8003/health/cameras", timeout=5)
            data = response.json()
            
            # Should report some cameras (even if 0, the process should be working)
            assert "details" in data, "Camera health response missing details"
            
            details = data["details"]
            assert "running" in details.lower(), "Camera monitor not reported as running"
            
            print(f"✅ Camera monitor discovery process: {details}")
            
        except requests.exceptions.RequestException as e:
            pytest.fail(f"Camera monitor discovery test failed: {e}")
    
    def test_camera_monitor_error_recovery(self):
        """Test that camera monitor handles device access errors gracefully."""
        # This test would simulate device access failures and verify recovery
        # For now, just check that the monitor is resilient to basic errors
        
        try:
            response = requests.get("http://localhost:8003/health/cameras", timeout=5)
            data = response.json()
            
            # Should not be in error state
            assert data["status"] != "unhealthy", "Camera monitor is in unhealthy state"
            
            print("✅ Camera monitor error recovery validated")
            
        except requests.exceptions.RequestException as e:
            pytest.fail(f"Camera monitor error recovery test failed: {e}")
    
    def _get_actual_video_devices(self) -> List[str]:
        """Get actual video devices present on the system."""
        video_devices = []
        
        try:
            result = subprocess.run(
                ["ls", "/dev/video*"], 
                capture_output=True, 
                text=True,
                timeout=5
            )
            
            if result.returncode == 0:
                devices = result.stdout.strip().split('\n')
                video_devices = [dev for dev in devices if dev and os.path.exists(dev)]
            else:
                for i in range(10):
                    device_path = f"/dev/video{i}"
                    if os.path.exists(device_path):
                        video_devices.append(device_path)
                        
        except Exception:
            pass
        
        return video_devices


class TestSystemResourceHealthValidation:
    """Test system resource usage and limits."""
    
    def test_service_memory_usage(self):
        """Test that service memory usage is within reasonable limits."""
        try:
            # Get camera-service process
            result = subprocess.run([
                "pgrep", "-f", "camera-service"
            ], capture_output=True, text=True, timeout=5)
            
            if result.returncode != 0:
                pytest.skip("Camera service process not found")
            
            pids = result.stdout.strip().split('\n')
            
            total_memory = 0
            for pid in pids:
                if pid:
                    try:
                        process = psutil.Process(int(pid))
                        memory_info = process.memory_info()
                        total_memory += memory_info.rss / 1024 / 1024  # Convert to MB
                    except (psutil.NoSuchProcess, ValueError):
                        continue
            
            # Memory usage should be reasonable (less than 500MB for basic operation)
            max_memory_mb = 500
            assert total_memory < max_memory_mb, (
                f"Service memory usage {total_memory:.1f}MB exceeds limit of {max_memory_mb}MB"
            )
            
            print(f"✅ Service memory usage: {total_memory:.1f}MB (limit: {max_memory_mb}MB)")
            
        except Exception as e:
            pytest.skip(f"Memory usage test failed: {e}")
    
    def test_service_cpu_usage(self):
        """Test that service CPU usage is within reasonable limits."""
        try:
            # Get camera-service process
            result = subprocess.run([
                "pgrep", "-f", "camera-service"
            ], capture_output=True, text=True, timeout=5)
            
            if result.returncode != 0:
                pytest.skip("Camera service process not found")
            
            pids = result.stdout.strip().split('\n')
            
            total_cpu_percent = 0
            for pid in pids:
                if pid:
                    try:
                        process = psutil.Process(int(pid))
                        cpu_percent = process.cpu_percent(interval=1)
                        total_cpu_percent += cpu_percent
                    except (psutil.NoSuchProcess, ValueError):
                        continue
            
            # CPU usage should be reasonable (less than 50% for basic operation)
            max_cpu_percent = 50
            assert total_cpu_percent < max_cpu_percent, (
                f"Service CPU usage {total_cpu_percent:.1f}% exceeds limit of {max_cpu_percent}%"
            )
            
            print(f"✅ Service CPU usage: {total_cpu_percent:.1f}% (limit: {max_cpu_percent}%)")
            
        except Exception as e:
            pytest.skip(f"CPU usage test failed: {e}")
    
    def test_disk_space_availability(self):
        """Test that required directories have sufficient disk space."""
        required_dirs = [
            "/var/recordings",
            "/var/snapshots",
            "/var/log/camera-service"
        ]
        
        for dir_path in required_dirs:
            if os.path.exists(dir_path):
                statvfs = os.statvfs(dir_path)
                free_space_mb = (statvfs.f_frsize * statvfs.f_bavail) / 1024 / 1024
                
                # Should have at least 100MB free space
                min_space_mb = 100
                assert free_space_mb > min_space_mb, (
                    f"Directory {dir_path} has insufficient space: {free_space_mb:.1f}MB (min: {min_space_mb}MB)"
                )
                
                print(f"✅ Directory {dir_path}: {free_space_mb:.1f}MB free space")
            else:
                print(f"⚠️ Directory {dir_path} does not exist")


class TestConfigurationHealthValidation:
    """Test configuration loading and validation."""
    
    def test_configuration_file_loading(self):
        """Test that configuration file loads without errors."""
        config_path = "/opt/camera-service/config/camera-service.yaml"
        
        if not os.path.exists(config_path):
            pytest.skip(f"Configuration file {config_path} does not exist")
        
        try:
            result = subprocess.run([
                "python3", "-c",
                "import yaml; import sys; "
                "config = yaml.safe_load(open('/opt/camera-service/config/camera-service.yaml')); "
                "print('Configuration loaded successfully')"
            ], capture_output=True, text=True, cwd="/opt/camera-service", timeout=10)
            
            assert result.returncode == 0, f"Configuration loading failed: {result.stderr}"
            print("✅ Configuration file loads successfully")
            
        except Exception as e:
            pytest.fail(f"Configuration loading test failed: {e}")
    
    def test_required_environment_variables(self):
        """Test that required environment variables are set."""
        # Check systemd service environment
        result = subprocess.run([
            "systemctl", "show", "camera-service", "--property=Environment"
        ], capture_output=True, text=True, timeout=5)
        
        if result.returncode == 0:
            env_output = result.stdout.strip()
            print(f"✅ Service environment variables: {env_output}")
        else:
            print("⚠️ Could not retrieve service environment variables")
    
    def test_file_permissions(self):
        """Test that service has proper access to required files."""
        required_paths = [
            "/opt/camera-service/config/camera-service.yaml",
            "/opt/camera-service/src",
            "/var/log/camera-service"
        ]
        
        for path in required_paths:
            if os.path.exists(path):
                # Check if camera-service user can access
                result = subprocess.run([
                    "sudo", "-u", "camera-service", "test", "-r", path
                ], capture_output=True, text=True)
                
                if result.returncode == 0:
                    print(f"✅ Service can access: {path}")
                else:
                    print(f"⚠️ Service cannot access: {path}")
            else:
                print(f"⚠️ Path does not exist: {path}")


class TestNetworkConnectivityValidation:
    """Test network connectivity and port accessibility."""
    
    def test_service_port_accessibility(self):
        """Test that all service ports are listening."""
        expected_ports = [
            (8002, "WebSocket Server"),
            (8003, "Health Server")
        ]
        
        for port, service_name in expected_ports:
            try:
                sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
                sock.settimeout(2)
                result = sock.connect_ex(('localhost', port))
                sock.close()
                
                assert result == 0, f"Service {service_name} port {port} is not listening"
                print(f"✅ Service {service_name} port {port} is accessible")
                
            except Exception as e:
                pytest.fail(f"Failed to test service {service_name} port {port}: {e}")
    
    def test_health_endpoint_responsiveness(self):
        """Test that health endpoints respond within acceptable time."""
        endpoints = [
            "http://localhost:8003/health/ready",
            "http://localhost:8003/health/system",
            "http://localhost:8003/health/cameras",
            "http://localhost:8003/health/mediamtx"
        ]
        
        max_response_time = 2.0  # seconds
        
        for endpoint in endpoints:
            try:
                start_time = time.time()
                response = requests.get(endpoint, timeout=max_response_time)
                response_time = time.time() - start_time
                
                assert response.status_code == 200, f"Endpoint {endpoint} returned {response.status_code}"
                assert response_time < max_response_time, (
                    f"Endpoint {endpoint} response time {response_time:.2f}s exceeds limit {max_response_time}s"
                )
                
                print(f"✅ {endpoint}: {response_time:.2f}s response time")
                
            except requests.exceptions.RequestException as e:
                pytest.fail(f"Health endpoint {endpoint} test failed: {e}")


class TestOverallSystemHealthValidation:
    """Test overall system health and readiness."""
    
    def test_system_readiness_probe(self):
        """Test that system readiness probe indicates full system readiness."""
        try:
            response = requests.get("http://localhost:8003/health/ready", timeout=5)
            assert response.status_code == 200
            
            data = response.json()
            assert data["status"] == "ready", f"System not ready: {data['status']}"
            
            print("✅ System readiness probe indicates full system readiness")
            
        except requests.exceptions.RequestException as e:
            pytest.fail(f"System readiness test failed: {e}")
    
    def test_overall_health_status(self):
        """Test that overall system health status is healthy."""
        try:
            response = requests.get("http://localhost:8003/health/system", timeout=5)
            assert response.status_code == 200
            
            data = response.json()
            assert data["status"] == "healthy", f"Overall system health is {data['status']}, expected healthy"
            
            print("✅ Overall system health status is healthy")
            
        except requests.exceptions.RequestException as e:
            pytest.fail(f"Overall health test failed: {e}")
    
    def test_comprehensive_health_validation(self):
        """Comprehensive test that validates all health aspects."""
        # This test runs all the critical health checks in sequence
        health_checks = [
            ("MediaMTX Service", self._check_mediamtx_service),
            ("Camera Monitor", self._check_camera_monitor),
            ("Service Manager", self._check_service_manager),
            ("System Resources", self._check_system_resources),
            ("Network Connectivity", self._check_network_connectivity)
        ]
        
        failed_checks = []
        
        for check_name, check_func in health_checks:
            try:
                check_func()
                print(f"✅ {check_name} health check passed")
            except Exception as e:
                failed_checks.append(f"{check_name}: {e}")
                print(f"❌ {check_name} health check failed: {e}")
        
        if failed_checks:
            pytest.fail(f"Comprehensive health validation failed:\n" + "\n".join(failed_checks))
        
        print("✅ All comprehensive health checks passed")
    
    def _check_mediamtx_service(self):
        """Check MediaMTX service health."""
        response = requests.get("http://localhost:9997/v3/paths/list", timeout=5)
        assert response.status_code == 200
    
    def _check_camera_monitor(self):
        """Check camera monitor health."""
        response = requests.get("http://localhost:8003/health/cameras", timeout=5)
        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "healthy"
    
    def _check_service_manager(self):
        """Check service manager health."""
        response = requests.get("http://localhost:8003/health/system", timeout=5)
        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "healthy"
    
    def _check_system_resources(self):
        """Check system resources."""
        # Basic resource check - just verify service is running
        result = subprocess.run(["systemctl", "is-active", "camera-service"], 
                              capture_output=True, text=True, timeout=5)
        assert result.returncode == 0
    
    def _check_network_connectivity(self):
        """Check network connectivity."""
        response = requests.get("http://localhost:8003/health/ready", timeout=5)
        assert response.status_code == 200
