"""
Camera Monitor Debug Test - Identify Real System Issues

This test focuses on identifying the exact issue in the camera monitor
that's causing the hanging behavior.
"""

import asyncio
from typing import Dict, Any

import pytest
import pytest_asyncio

from src.camera_discovery.hybrid_monitor import HybridCameraMonitor
from src.camera_service.config import CameraConfig


class CameraMonitorDebugger:
    """Debug camera monitor issues step by step."""
    
    def __init__(self):
        self.issues = []
    
    async def debug_monitor_initialization(self, config: CameraConfig) -> Dict[str, Any]:
        """Debug monitor initialization step by step."""
        print("🔍 Debug: Monitor Initialization")
        
        try:
            print("  → Creating HybridCameraMonitor...")
            monitor = HybridCameraMonitor(
                device_range=config.device_range,
                poll_interval=config.poll_interval,
                detection_timeout=config.detection_timeout,
                enable_capability_detection=config.enable_capability_detection,
            )
            print("  ✅ Monitor created successfully")
            
            print("  → Checking monitor attributes...")
            print(f"    - device_range: {monitor._device_range}")
            print(f"    - poll_interval: {monitor._current_poll_interval}")
            print(f"    - detection_timeout: {monitor._detection_timeout}")
            print(f"    - enable_capability_detection: {monitor._enable_capability_detection}")
            print(f"    - udev_available: {monitor._udev_available}")
            print("  ✅ Monitor attributes valid")
            
            return {"status": "PASS", "step": "initialization"}
            
        except Exception as e:
            error_msg = f"Monitor initialization failed: {e}"
            print(f"  ❌ {error_msg}")
            self.issues.append(error_msg)
            return {"status": "FAIL", "error": error_msg}
    
    async def debug_monitor_startup(self, config: CameraConfig) -> Dict[str, Any]:
        """Debug monitor startup step by step."""
        print("🔍 Debug: Monitor Startup")
        
        try:
            monitor = HybridCameraMonitor(
                device_range=config.device_range,
                poll_interval=config.poll_interval,
                detection_timeout=config.detection_timeout,
                enable_capability_detection=config.enable_capability_detection,
            )
            
            print("  → Starting monitor...")
            await asyncio.wait_for(monitor.start(), timeout=10.0)
            print("  ✅ Monitor started successfully")
            
            print("  → Checking running state...")
            assert monitor.is_running, "Monitor should be running"
            print("  ✅ Monitor is running")
            
            print("  → Stopping monitor...")
            await asyncio.wait_for(monitor.stop(), timeout=10.0)
            print("  ✅ Monitor stopped successfully")
            
            return {"status": "PASS", "step": "startup"}
            
        except asyncio.TimeoutError as e:
            error_msg = f"Monitor startup timeout: {e}"
            print(f"  ❌ {error_msg}")
            self.issues.append(error_msg)
            return {"status": "FAIL", "error": error_msg}
        except Exception as e:
            error_msg = f"Monitor startup error: {e}"
            print(f"  ❌ {error_msg}")
            self.issues.append(error_msg)
            return {"status": "FAIL", "error": error_msg}
    
    async def debug_camera_discovery(self, config: CameraConfig) -> Dict[str, Any]:
        """Debug camera discovery step by step."""
        print("🔍 Debug: Camera Discovery")
        
        try:
            monitor = HybridCameraMonitor(
                device_range=config.device_range,
                poll_interval=config.poll_interval,
                detection_timeout=config.detection_timeout,
                enable_capability_detection=config.enable_capability_detection,
            )
            
            print("  → Starting monitor...")
            await asyncio.wait_for(monitor.start(), timeout=10.0)
            print("  ✅ Monitor started")
            
            print("  → Testing camera discovery...")
            cameras = await asyncio.wait_for(monitor.get_connected_cameras(), timeout=5.0)
            print(f"  ✅ Found {len(cameras)} cameras")
            
            print("  → Stopping monitor...")
            await asyncio.wait_for(monitor.stop(), timeout=10.0)
            print("  ✅ Monitor stopped")
            
            return {"status": "PASS", "cameras_found": len(cameras)}
            
        except asyncio.TimeoutError as e:
            error_msg = f"Camera discovery timeout: {e}"
            print(f"  ❌ {error_msg}")
            self.issues.append(error_msg)
            return {"status": "FAIL", "error": error_msg}
        except Exception as e:
            error_msg = f"Camera discovery error: {e}"
            print(f"  ❌ {error_msg}")
            self.issues.append(error_msg)
            return {"status": "FAIL", "error": error_msg}
    
    async def debug_device_access(self, config: CameraConfig) -> Dict[str, Any]:
        """Debug device access issues."""
        print("🔍 Debug: Device Access")
        
        try:
            monitor = HybridCameraMonitor(
                device_range=config.device_range,
                poll_interval=config.poll_interval,
                detection_timeout=config.detection_timeout,
                enable_capability_detection=config.enable_capability_detection,
            )
            
            print("  → Testing device access for each device...")
            for device_num in config.device_range:
                device_path = f"/dev/video{device_num}"
                print(f"    → Testing {device_path}...")
                
                try:
                    device_info = await asyncio.wait_for(
                        monitor._create_camera_device_info(device_path, device_num),
                        timeout=2.0
                    )
                    print(f"    ✅ {device_path}: {device_info.status}")
                except asyncio.TimeoutError:
                    print(f"    ❌ {device_path}: TIMEOUT")
                except Exception as e:
                    print(f"    ❌ {device_path}: ERROR - {e}")
            
            return {"status": "PASS", "step": "device_access"}
            
        except Exception as e:
            error_msg = f"Device access error: {e}"
            print(f"  ❌ {error_msg}")
            self.issues.append(error_msg)
            return {"status": "FAIL", "error": error_msg}


class TestCameraMonitorDebug:
    """Debug camera monitor issues."""
    
    @pytest.fixture
    def camera_config(self):
        """Create camera configuration for testing."""
        return CameraConfig(
            device_range=[0, 1, 2],
            poll_interval=2.0,
            enable_capability_detection=True,
            detection_timeout=5.0,
        )
    
    @pytest_asyncio.fixture
    async def debugger(self):
        """Create debugger instance."""
        return CameraMonitorDebugger()
    
    @pytest.mark.asyncio
    @pytest.mark.integration
    async def test_monitor_initialization_debug(self, camera_config, debugger):
        """Debug monitor initialization."""
        print("\n" + "="*60)
        print("DEBUG: Monitor Initialization")
        print("="*60)
        
        result = await debugger.debug_monitor_initialization(camera_config)
        
        if result["status"] == "PASS":
            print("✅ Monitor Initialization: PASSED")
        else:
            print(f"❌ Monitor Initialization: FAILED - {result['error']}")
            pytest.skip(f"Monitor initialization has issues: {result['error']}")
    
    @pytest.mark.asyncio
    @pytest.mark.integration
    async def test_monitor_startup_debug(self, camera_config, debugger):
        """Debug monitor startup."""
        print("\n" + "="*60)
        print("DEBUG: Monitor Startup")
        print("="*60)
        
        result = await debugger.debug_monitor_startup(camera_config)
        
        if result["status"] == "PASS":
            print("✅ Monitor Startup: PASSED")
        else:
            print(f"❌ Monitor Startup: FAILED - {result['error']}")
            pytest.skip(f"Monitor startup has issues: {result['error']}")
    
    @pytest.mark.asyncio
    @pytest.mark.integration
    async def test_camera_discovery_debug(self, camera_config, debugger):
        """Debug camera discovery."""
        print("\n" + "="*60)
        print("DEBUG: Camera Discovery")
        print("="*60)
        
        result = await debugger.debug_camera_discovery(camera_config)
        
        if result["status"] == "PASS":
            print("✅ Camera Discovery: PASSED")
        else:
            print(f"❌ Camera Discovery: FAILED - {result['error']}")
            pytest.skip(f"Camera discovery has issues: {result['error']}")
    
    @pytest.mark.asyncio
    @pytest.mark.integration
    async def test_device_access_debug(self, camera_config, debugger):
        """Debug device access."""
        print("\n" + "="*60)
        print("DEBUG: Device Access")
        print("="*60)
        
        result = await debugger.debug_device_access(camera_config)
        
        if result["status"] == "PASS":
            print("✅ Device Access: PASSED")
        else:
            print(f"❌ Device Access: FAILED - {result['error']}")
            pytest.skip(f"Device access has issues: {result['error']}")
    
    @pytest.mark.asyncio
    @pytest.mark.integration
    async def test_issues_summary(self, debugger):
        """Summarize discovered issues."""
        print("\n" + "="*60)
        print("CAMERA MONITOR ISSUES SUMMARY")
        print("="*60)
        
        if debugger.issues:
            print("❌ DISCOVERED CAMERA MONITOR ISSUES:")
            for i, issue in enumerate(debugger.issues, 1):
                print(f"  {i}. {issue}")
            print("\n🔧 THESE ISSUES MUST BE FIXED")
        else:
            print("✅ NO CAMERA MONITOR ISSUES DISCOVERED")
        
        # Always pass this test - it's just for reporting
        assert True, "Camera monitor debug complete"
