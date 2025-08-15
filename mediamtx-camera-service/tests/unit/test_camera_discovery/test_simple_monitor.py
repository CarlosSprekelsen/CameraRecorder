"""
Simple test to isolate hanging issues in HybridCameraMonitor.

Requirements Traceability:
- REQ-CAM-001: Camera discovery shall provide simple monitor functionality without hanging
- REQ-CAM-001: Camera discovery shall handle monitor lifecycle with timeout controls
- REQ-CAM-001: Camera discovery shall isolate hanging issues in monitor operations

Story Coverage: S3 - Camera Discovery Hardening
IV&V Control Point: Real monitor lifecycle validation
"""

import asyncio
import pytest
from unittest.mock import patch

from src.camera_discovery.hybrid_monitor import HybridCameraMonitor


class TestSimpleMonitor:
    """Simple tests to isolate hanging issues."""

    @pytest.fixture
    def monitor(self):
        """Create a simple monitor instance."""
        return HybridCameraMonitor(
            device_range=[0, 1],
            enable_capability_detection=False,  # Disable for simple test
            detection_timeout=1.0,
        )

    @pytest.mark.asyncio
    async def test_monitor_creation(self, monitor):
        """Test that monitor can be created without hanging."""
        assert monitor is not None
        assert not monitor.is_running
        assert monitor._known_devices == {}

    @pytest.mark.asyncio
    async def test_monitor_start_stop_simple(self, monitor):
        """Test simple start/stop without complex operations."""
        
        # Mock udev to be unavailable
        with patch("src.camera_discovery.hybrid_monitor.HAS_PYUDEV", False):
            # Start monitor with timeout
            try:
                await asyncio.wait_for(monitor.start(), timeout=10.0)
                
                assert monitor.is_running
                
                # Stop monitor with timeout
                await asyncio.wait_for(monitor.stop(), timeout=10.0)
                
                assert not monitor.is_running
                
            except asyncio.TimeoutError:
                # Force cleanup
                monitor._running = False
                for task in monitor._monitoring_tasks:
                    if not task.done():
                        task.cancel()
                await asyncio.gather(*monitor._monitoring_tasks, return_exceptions=True)
                raise

    @pytest.mark.asyncio
    async def test_monitor_discovery_timeout(self, monitor):
        """Test that discovery operations timeout properly."""
        
        # Mock discovery to hang
        with patch.object(monitor, "_discover_cameras") as mock_discover:
            async def hanging_discovery():
                await asyncio.sleep(10.0)  # Hang for 10 seconds
                
            mock_discover.side_effect = hanging_discovery
            
            # Start monitor
            with patch("src.camera_discovery.hybrid_monitor.HAS_PYUDEV", False):
                await asyncio.wait_for(monitor.start(), timeout=15.0)
                
                # Let it run for a short time
                await asyncio.sleep(0.5)
                
                # Stop monitor
                await asyncio.wait_for(monitor.stop(), timeout=10.0)
                
                assert not monitor.is_running
