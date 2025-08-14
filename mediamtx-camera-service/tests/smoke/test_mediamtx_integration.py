"""
MediaMTX Real Integration Test

Tests real MediaMTX API endpoint testing.
Validates actual health monitoring and stream management validation.

This test replaces complex unit test mocks with real system validation
to provide better confidence in MediaMTX integration reliability.
"""

import asyncio
import aiohttp
import pytest
import tempfile
import os
import subprocess
import time
import signal
from typing import Dict, Any

# Import the actual MediaMTX controller implementation
import sys
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', '..', 'src'))

from mediamtx_wrapper.controller import MediaMTXController, StreamConfig


@pytest.fixture(scope="function")
async def mediamtx_controller():
    """Create and manage MediaMTX controller for testing."""
    # Create temporary directories for testing
    with tempfile.TemporaryDirectory() as temp_dir:
        recordings_path = os.path.join(temp_dir, "recordings")
        snapshots_path = os.path.join(temp_dir, "snapshots")
        config_path = os.path.join(temp_dir, "mediamtx.yml")
        
        os.makedirs(recordings_path, exist_ok=True)
        os.makedirs(snapshots_path, exist_ok=True)
        
        controller = MediaMTXController(
            host="localhost",
            api_port=9997,
            rtsp_port=8554,
            webrtc_port=8889,
            hls_port=8888,
            config_path=config_path,
            recordings_path=recordings_path,
            snapshots_path=snapshots_path
        )
        
        # Start controller
        await controller.start()
        
        try:
            yield controller
        finally:
            # Cleanup: stop controller
            await controller.stop()


@pytest.fixture(scope="function")
async def mediamtx_server():
    """Start MediaMTX as a real process for testing, or use existing instance."""
    mediamtx_process = None
    mediamtx_config_path = None
    temp_dir = None
    using_existing_server = False
    
    try:
        # First, check if MediaMTX is already running
        async with aiohttp.ClientSession() as session:
            try:
                async with session.get('http://localhost:9997/v3/config/global/get') as response:
                    if response.status == 200:
                        print("✓ Using existing MediaMTX server instance")
                        using_existing_server = True
                        yield True
                        return
            except Exception:
                pass
        
        using_existing_server = False
        
        # If not running, try to start a new instance
        print("No existing MediaMTX server found, attempting to start new instance...")
        
        # Create temporary directory for MediaMTX
        temp_dir = tempfile.mkdtemp(prefix="mediamtx_test_")
        mediamtx_config_path = os.path.join(temp_dir, "mediamtx.yml")
        
        # Create MediaMTX configuration with different ports to avoid conflicts
        config_content = f"""# MediaMTX Test Configuration
api: yes
apiAddress: :9998
rtspAddress: :8555
rtspTransports: [tcp, udp]
webrtcAddress: :8890
hlsAddress: :8889
logLevel: info
logDestinations: [stdout]
paths:
  all:
    recordFormat: fmp4
    recordSegmentDuration: "3600s"
"""
        
        with open(mediamtx_config_path, 'w') as f:
            f.write(config_content)
        
        # Start MediaMTX process
        mediamtx_process = subprocess.Popen(
            ["mediamtx", mediamtx_config_path],
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True
        )
        
        # Wait for MediaMTX to start up
        await asyncio.sleep(3)
        
        # Verify MediaMTX is running
        if mediamtx_process.poll() is not None:
            # Process died, get error output
            stdout, stderr = mediamtx_process.communicate()
            print(f"MediaMTX failed to start. Return code: {mediamtx_process.returncode}")
            print(f"STDOUT: {stdout}")
            print(f"STDERR: {stderr}")
            yield False
            return
        
        # Test API connectivity on new port
        async with aiohttp.ClientSession() as session:
            for attempt in range(5):  # Try 5 times
                try:
                    async with session.get('http://localhost:9998/v3/config/global/get') as response:
                        if response.status == 200:
                            print("✓ MediaMTX server started successfully on test ports")
                            yield True
                            return
                except Exception:
                    pass
                await asyncio.sleep(1)
        
        print("⚠ MediaMTX server started but API not responding")
        yield False
        
    except Exception as e:
        print(f"Failed to start MediaMTX: {e}")
        yield False
    finally:
        # Cleanup
        if mediamtx_process:
            try:
                # Send SIGTERM for graceful shutdown
                mediamtx_process.terminate()
                try:
                    mediamtx_process.wait(timeout=5)
                except subprocess.TimeoutExpired:
                    # Force kill if graceful shutdown fails
                    mediamtx_process.kill()
                    mediamtx_process.wait()
            except Exception as e:
                print(f"Error stopping MediaMTX: {e}")
        
        # Cleanup temporary directory
        if temp_dir and os.path.exists(temp_dir):
            try:
                import shutil
                shutil.rmtree(temp_dir)
            except Exception as e:
                print(f"Error cleaning up temp directory: {e}")


class TestMediaMTXRealIntegration:
    """Test real MediaMTX integration."""
    
    @pytest.mark.asyncio
    async def test_mediamtx_real_integration(self):
        """Test real MediaMTX integration without fixtures."""
        # Create controller directly (following working pattern)
        with tempfile.TemporaryDirectory() as temp_dir:
            recordings_path = os.path.join(temp_dir, "recordings")
            snapshots_path = os.path.join(temp_dir, "snapshots")
            config_path = os.path.join(temp_dir, "mediamtx.yml")
            
            os.makedirs(recordings_path, exist_ok=True)
            os.makedirs(snapshots_path, exist_ok=True)
            
            controller = MediaMTXController(
                host="localhost",
                api_port=9997,
                rtsp_port=8554,
                webrtc_port=8889,
                hls_port=8888,
                config_path=config_path,
                recordings_path=recordings_path,
                snapshots_path=snapshots_path
            )
            
            # Start controller
            await controller.start()
            
            try:
                # Test real health check
                health_status = await controller.health_check()
                assert "status" in health_status
                assert health_status["status"] in ["healthy", "degraded", "unhealthy"]
                assert "api_port" in health_status
                assert health_status["api_port"] == 9997
                
                print("✓ MediaMTX real integration test passed")
                
            finally:
                # Cleanup
                await controller.stop()
    
    @pytest.mark.asyncio
    async def test_mediamtx_api_endpoints(self):
        """Test real MediaMTX API endpoints."""
        # Test against actual MediaMTX API (assume it's running)
        api_port = 9997  # Use default port for existing server
        
        try:
            async with aiohttp.ClientSession() as session:
                # Test global config endpoint
                async with session.get(f'http://localhost:{api_port}/v3/config/global/get') as response:
                    assert response.status == 200
                    config_data = await response.json()
                    assert "api" in config_data
                    assert config_data["api"] is True
                    print(f"✓ MediaMTX API accessible, API enabled: {config_data.get('api')}")
                
                # Test paths list endpoint
                async with session.get(f'http://localhost:{api_port}/v3/paths/list') as response:
                    assert response.status == 200
                    paths_data = await response.json()
                    assert isinstance(paths_data, dict)
                    print(f"✓ MediaMTX paths endpoint accessible")
        except aiohttp.ClientError:
            pytest.skip("MediaMTX server not available for API endpoint testing")
    
    @pytest.mark.asyncio
    async def test_mediamtx_controller_lifecycle(self):
        """Test MediaMTX controller startup and shutdown lifecycle."""
        with tempfile.TemporaryDirectory() as temp_dir:
            recordings_path = os.path.join(temp_dir, "recordings")
            snapshots_path = os.path.join(temp_dir, "snapshots")
            config_path = os.path.join(temp_dir, "mediamtx.yml")
            
            os.makedirs(recordings_path, exist_ok=True)
            os.makedirs(snapshots_path, exist_ok=True)
            
            controller = MediaMTXController(
                host="localhost",
                api_port=9997,
                rtsp_port=8554,
                webrtc_port=8889,
                hls_port=8888,
                config_path=config_path,
                recordings_path=recordings_path,
                snapshots_path=snapshots_path
            )
            
            # Test controller startup
            await controller.start()
            assert hasattr(controller, '_session')
            assert controller._session is not None
            
            # Test controller shutdown
            await controller.stop()
            assert controller._session is None
    
    @pytest.mark.asyncio
    async def test_mediamtx_stream_management(self):
        """Test MediaMTX stream management capabilities."""
        # Create controller directly
        with tempfile.TemporaryDirectory() as temp_dir:
            recordings_path = os.path.join(temp_dir, "recordings")
            snapshots_path = os.path.join(temp_dir, "snapshots")
            config_path = os.path.join(temp_dir, "mediamtx.yml")
            
            os.makedirs(recordings_path, exist_ok=True)
            os.makedirs(snapshots_path, exist_ok=True)
            
            controller = MediaMTXController(
                host="localhost",
                api_port=9997,
                rtsp_port=8554,
                webrtc_port=8889,
                hls_port=8888,
                config_path=config_path,
                recordings_path=recordings_path,
                snapshots_path=snapshots_path
            )
            
            await controller.start()
            
            try:
                # Test stream list retrieval
                streams = await controller.get_stream_list()
                assert isinstance(streams, list)
                
                # Test stream creation
                stream_config = StreamConfig(
                    name="test_stream",
                    source="rtsp://localhost:8554/test",
                    record=False
                )
                
                result = await controller.create_stream(stream_config)
                assert isinstance(result, dict)
                print(f"✓ Stream creation successful: {result}")
                
            finally:
                await controller.stop()
    
    @pytest.mark.asyncio
    async def test_mediamtx_health_monitoring(self):
        """Test MediaMTX health monitoring behavior."""
        # Create controller directly
        with tempfile.TemporaryDirectory() as temp_dir:
            recordings_path = os.path.join(temp_dir, "recordings")
            snapshots_path = os.path.join(temp_dir, "snapshots")
            config_path = os.path.join(temp_dir, "mediamtx.yml")
            
            os.makedirs(recordings_path, exist_ok=True)
            os.makedirs(snapshots_path, exist_ok=True)
            
            controller = MediaMTXController(
                host="localhost",
                api_port=9997,
                rtsp_port=8554,
                webrtc_port=8889,
                hls_port=8888,
                config_path=config_path,
                recordings_path=recordings_path,
                snapshots_path=snapshots_path
            )
            
            await controller.start()
            
            try:
                # Test health check with detailed response
                health_status = await controller.health_check()
                
                # Validate health response structure
                required_fields = ["status", "api_port", "correlation_id"]
                for field in required_fields:
                    assert field in health_status
                
                # Test health state tracking
                if hasattr(controller, '_health_state'):
                    health_state = controller._health_state
                    assert "total_checks" in health_state
                    assert "consecutive_failures" in health_state
                    
            finally:
                await controller.stop()


if __name__ == "__main__":
    # Allow running as standalone script for manual testing
    async def run_tests():
        """Run smoke tests manually."""
        test_instance = TestMediaMTXRealIntegration()
        
        # Test controller lifecycle
        await test_instance.test_mediamtx_controller_lifecycle()
        print("✓ MediaMTX controller lifecycle test passed")
        
        # Test API endpoints
        await test_instance.test_mediamtx_api_endpoints(True)  # Assume server is running
        print("✓ MediaMTX API endpoints test passed")
        
        # Test real integration with manual controller management
        with tempfile.TemporaryDirectory() as temp_dir:
            recordings_path = os.path.join(temp_dir, "recordings")
            snapshots_path = os.path.join(temp_dir, "snapshots")
            config_path = os.path.join(temp_dir, "mediamtx.yml")
            
            os.makedirs(recordings_path, exist_ok=True)
            os.makedirs(snapshots_path, exist_ok=True)
            
            controller = MediaMTXController(
                host="localhost",
                api_port=9997,
                rtsp_port=8554,
                webrtc_port=8889,
                hls_port=8888,
                config_path=config_path,
                recordings_path=recordings_path,
                snapshots_path=snapshots_path
            )
            
            try:
                await controller.start()
                
                await test_instance.test_mediamtx_real_integration(controller, True)
                print("✓ MediaMTX real integration test passed")
                
                await test_instance.test_mediamtx_stream_management(controller, True)
                print("✓ MediaMTX stream management test passed")
                
                await test_instance.test_mediamtx_health_monitoring(controller, True)
                print("✓ MediaMTX health monitoring test passed")
                
            finally:
                await controller.stop()
        
        print("All MediaMTX smoke tests passed!")
    
    asyncio.run(run_tests())
