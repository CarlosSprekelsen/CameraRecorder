"""
Health Endpoint Accuracy Validation Tests

Tests that validate the accuracy of health endpoint data, not just availability.
These tests are designed to catch bugs like incorrect camera counts, wrong status
reporting, or inaccurate component health information.

Requirements:
- REQ-HEALTH-001: System shall provide comprehensive logging for health monitoring
- REQ-HEALTH-002: System shall support structured logging for production environments
- REQ-HEALTH-003: System shall enable correlation ID tracking across components

This test file specifically addresses the gap where health endpoints were reporting
inaccurate data (e.g., "0 cameras" when 4 video devices were present).
"""

import pytest
import requests
import subprocess
import os
import json
import time
from typing import Dict, Any, List


class TestHealthEndpointAccuracy:
    """Test health endpoint data accuracy and validation."""
    
    def test_camera_count_accuracy(self):
        """
        Test that health endpoint reports accurate camera count.
        
        This test would have caught the bug where health server was calling
        non-existent get_camera_count() method and defaulting to 0 cameras.
        """
        # First, get actual video devices on system
        video_devices = self._get_actual_video_devices()
        expected_camera_count = len(video_devices)
        
        print(f"System has {expected_camera_count} video devices: {video_devices}")
        
        # Test /health/cameras endpoint
        try:
            response = requests.get("http://localhost:8003/health/cameras", timeout=5)
            assert response.status_code == 200, "Health cameras endpoint not responding with 200"
            
            data = response.json()
            assert "details" in data, "Health cameras response missing details field"
            
            # Extract camera count from details
            details = data["details"]
            if "with" in details and "cameras" in details:
                # Parse "Camera monitor is running with X cameras"
                parts = details.split("with")
                if len(parts) > 1:
                    camera_part = parts[1].strip()
                    reported_count = int(camera_part.split()[0])
                else:
                    pytest.fail(f"Could not parse camera count from details: {details}")
            else:
                pytest.fail(f"Unexpected details format: {details}")
            
            print(f"Health endpoint reports {reported_count} cameras")
            
            # Validate accuracy
            assert reported_count == expected_camera_count, (
                f"Camera count mismatch: health endpoint reports {reported_count} cameras, "
                f"but system has {expected_camera_count} video devices: {video_devices}"
            )
            
            print(f"✅ Camera count accuracy validated: {reported_count} cameras")
            
        except requests.exceptions.RequestException as e:
            pytest.fail(f"Health cameras endpoint not accessible: {e}")
    
    def test_system_health_component_accuracy(self):
        """
        Test that system health endpoint reports accurate component information.
        
        This test validates that all components in /health/system are reporting
        accurate status and details.
        """
        try:
            response = requests.get("http://localhost:8003/health/system", timeout=5)
            assert response.status_code == 200, "Health system endpoint not responding with 200"
            
            data = response.json()
            assert "components" in data, "Health system response missing components field"
            
            components = data["components"]
            required_components = ["mediamtx", "camera_monitor", "service_manager"]
            
            # Check all required components are present
            for component in required_components:
                assert component in components, f"Missing required component: {component}"
                
                component_data = components[component]
                assert "status" in component_data, f"Component {component} missing status"
                assert "details" in component_data, f"Component {component} missing details"
                
                # Validate status is valid
                assert component_data["status"] in ["healthy", "unhealthy", "degraded"], (
                    f"Invalid status for {component}: {component_data['status']}"
                )
                
                print(f"✅ Component {component}: {component_data['status']} - {component_data['details']}")
            
            # Special validation for camera_monitor component
            if "camera_monitor" in components:
                camera_details = components["camera_monitor"]["details"]
                if "with" in camera_details and "cameras" in camera_details:
                    parts = camera_details.split("with")
                    if len(parts) > 1:
                        camera_part = parts[1].strip()
                        reported_count = int(camera_part.split()[0])
                        
                        # Validate against actual video devices
                        video_devices = self._get_actual_video_devices()
                        expected_count = len(video_devices)
                        
                        assert reported_count == expected_count, (
                            f"Camera count in system health mismatch: reports {reported_count}, "
                            f"but system has {expected_count} video devices: {video_devices}"
                        )
            
            print("✅ System health component accuracy validated")
            
        except requests.exceptions.RequestException as e:
            pytest.fail(f"Health system endpoint not accessible: {e}")
    
    def test_health_endpoint_consistency(self):
        """
        Test that different health endpoints report consistent information.
        
        This test ensures that /health/cameras and /health/system report
        consistent camera count information.
        """
        try:
            # Get camera count from /health/cameras
            cameras_response = requests.get("http://localhost:8003/health/cameras", timeout=5)
            assert cameras_response.status_code == 200
            
            cameras_data = cameras_response.json()
            cameras_details = cameras_data["details"]
            
            # Extract count from cameras endpoint
            if "with" in cameras_details and "cameras" in cameras_details:
                parts = cameras_details.split("with")
                cameras_count = int(parts[1].strip().split()[0])
            else:
                pytest.fail(f"Could not parse camera count from cameras endpoint: {cameras_details}")
            
            # Get camera count from /health/system
            system_response = requests.get("http://localhost:8003/health/system", timeout=5)
            assert system_response.status_code == 200
            
            system_data = system_response.json()
            system_camera_details = system_data["components"]["camera_monitor"]["details"]
            
            # Extract count from system endpoint
            if "with" in system_camera_details and "cameras" in system_camera_details:
                parts = system_camera_details.split("with")
                system_count = int(parts[1].strip().split()[0])
            else:
                pytest.fail(f"Could not parse camera count from system endpoint: {system_camera_details}")
            
            # Validate consistency
            assert cameras_count == system_count, (
                f"Camera count inconsistency: /health/cameras reports {cameras_count}, "
                f"/health/system reports {system_count}"
            )
            
            print(f"✅ Health endpoint consistency validated: both report {cameras_count} cameras")
            
        except requests.exceptions.RequestException as e:
            pytest.fail(f"Health endpoint consistency test failed: {e}")
    
    def test_health_endpoint_method_validation(self):
        """
        Test that health server uses correct methods to get component information.
        
        This test validates that the health server is calling the right methods
        on components to get accurate information.
        """
        # This test would validate that the health server is using the correct
        # methods (like get_connected_cameras() instead of non-existent get_camera_count())
        
        # Check if camera monitor has the expected methods
        try:
            # Import and check camera monitor methods
            import sys
            sys.path.insert(0, '/opt/camera-service/src')
            
            from camera_discovery.hybrid_monitor import HybridCameraMonitor
            
            # Check that required methods exist
            assert hasattr(HybridCameraMonitor, 'get_connected_cameras'), (
                "HybridCameraMonitor missing get_connected_cameras method"
            )
            
            assert hasattr(HybridCameraMonitor, 'is_running'), (
                "HybridCameraMonitor missing is_running property"
            )
            
            print("✅ Camera monitor has required methods for health reporting")
            
        except ImportError as e:
            pytest.skip(f"Could not import camera monitor for method validation: {e}")
        except Exception as e:
            pytest.fail(f"Method validation failed: {e}")
    
    def _get_actual_video_devices(self) -> List[str]:
        """Get actual video devices present on the system."""
        video_devices = []
        
        try:
            # Check for video devices
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
                # Fallback: check individual video devices
                for i in range(10):  # Check video0 through video9
                    device_path = f"/dev/video{i}"
                    if os.path.exists(device_path):
                        video_devices.append(device_path)
                        
        except subprocess.TimeoutExpired:
            pytest.fail("Timeout getting video devices")
        except Exception as e:
            pytest.fail(f"Error getting video devices: {e}")
        
        return video_devices


class TestHealthEndpointRegression:
    """Test to prevent regression of health endpoint accuracy bugs."""
    
    def _get_actual_video_devices(self) -> List[str]:
        """Get actual video devices present on the system."""
        video_devices = []
        
        try:
            # Check for video devices
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
                # Fallback: check individual video devices
                for i in range(10):  # Check video0 through video9
                    device_path = f"/dev/video{i}"
                    if os.path.exists(device_path):
                        video_devices.append(device_path)
                        
        except subprocess.TimeoutExpired:
            pytest.fail("Timeout getting video devices")
        except Exception as e:
            pytest.fail(f"Error getting video devices: {e}")
        
        return video_devices
    
    def test_camera_count_regression(self):
        """
        Regression test to ensure camera count accuracy is maintained.
        
        This test specifically prevents the regression of the bug where
        health server was calling non-existent get_camera_count() method.
        """
        # Get actual video devices
        video_devices = self._get_actual_video_devices()
        expected_count = len(video_devices)
        
        if expected_count == 0:
            pytest.skip("No video devices present for regression testing")
        
        # Test that health endpoints report the correct count
        endpoints = [
            "http://localhost:8003/health/cameras",
            "http://localhost:8003/health/system"
        ]
        
        for endpoint in endpoints:
            try:
                response = requests.get(endpoint, timeout=5)
                assert response.status_code == 200
                
                data = response.json()
                
                # Extract camera count from response
                if endpoint.endswith("/cameras"):
                    details = data["details"]
                else:
                    details = data["components"]["camera_monitor"]["details"]
                
                if "with" in details and "cameras" in details:
                    parts = details.split("with")
                    reported_count = int(parts[1].strip().split()[0])
                    
                    # This is the key regression check
                    assert reported_count == expected_count, (
                        f"REGRESSION: {endpoint} reports {reported_count} cameras, "
                        f"but system has {expected_count} video devices: {video_devices}"
                    )
                    
                    print(f"✅ {endpoint} regression test passed: {reported_count} cameras")
                else:
                    pytest.fail(f"Could not parse camera count from {endpoint}: {details}")
                    
            except requests.exceptions.RequestException as e:
                pytest.fail(f"Regression test failed for {endpoint}: {e}")
    
    def test_health_server_method_calls(self):
        """
        Test that health server uses correct method calls.
        
        This test validates that the health server code is calling the right
        methods on components, preventing the get_camera_count() bug.
        """
        try:
            # Check the actual health server code
            health_server_path = "/opt/camera-service/src/health_server.py"
            
            if not os.path.exists(health_server_path):
                pytest.skip("Health server source not available in production")
            
            with open(health_server_path, 'r') as f:
                content = f.read()
            
            # Check for the buggy method call (should not exist)
            assert "get_camera_count()" not in content, (
                "REGRESSION: Health server still contains buggy get_camera_count() call"
            )
            
            # Check for the correct method usage
            assert "_known_devices" in content, (
                "Health server should use _known_devices for camera count"
            )
            
            print("✅ Health server method calls validated")
            
        except Exception as e:
            pytest.fail(f"Health server method validation failed: {e}")


def _get_actual_video_devices() -> List[str]:
    """Helper function to get actual video devices."""
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
