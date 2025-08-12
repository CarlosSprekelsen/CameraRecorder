"""
Critical Prototype: Real RTSP Stream Handling

This prototype validates real RTSP stream handling with actual camera feeds or simulators.
It proves design implementability through actual system execution.

PRINCIPLE: NO MOCKING - Only real system validation
"""

import asyncio
import json
import os
import tempfile
import time
import subprocess
from pathlib import Path
from typing import Dict, Any, Optional

import pytest
import pytest_asyncio
import aiohttp
import cv2
import numpy as np

# Import real components - NO MOCKING
from src.mediamtx_wrapper.controller import MediaMTXController
from src.camera_service.config import Config, MediaMTXConfig, ServerConfig, CameraConfig, RecordingConfig
from src.camera_service.service_manager import ServiceManager


class RealRTSPStreamHandlingPrototype:
    """
    Critical prototype for real RTSP stream handling validation.
    
    This prototype systematically tests RTSP stream handling using real components
    to prove design implementability through actual system execution.
    """
    
    def __init__(self):
        self.test_results = {}
        self.system_issues = []
        self.mediamtx_controller = None
        self.service_manager = None
        self.temp_dir = None
        self.test_stream_process = None
        
    async def setup_real_environment(self):
        """Set up real test environment with actual RTSP stream handling."""
        self.temp_dir = tempfile.mkdtemp(prefix="pdr_rtsp_")
        
        # Create real MediaMTX configuration
        mediamtx_config = MediaMTXConfig(
            host="127.0.0.1",
            api_port=9997,
            rtsp_port=8554,
            webrtc_port=8889,
            hls_port=8888,
            config_path=f"{self.temp_dir}/mediamtx.yml",
            recordings_path=f"{self.temp_dir}/recordings",
            snapshots_path=f"{self.temp_dir}/snapshots"
        )
        
        # Initialize real MediaMTX controller
        self.mediamtx_controller = MediaMTXController(
            host=mediamtx_config.host,
            api_port=mediamtx_config.api_port,
            rtsp_port=mediamtx_config.rtsp_port,
            webrtc_port=mediamtx_config.webrtc_port,
            hls_port=mediamtx_config.hls_port,
            config_path=mediamtx_config.config_path,
            recordings_path=mediamtx_config.recordings_path,
            snapshots_path=mediamtx_config.snapshots_path
        )
        
        # Initialize real service manager
        config = Config(
            server=ServerConfig(host="127.0.0.1", port=8000),
            mediamtx=mediamtx_config,
            camera=CameraConfig(device_range=[0, 1, 2]),
            recording=RecordingConfig(enabled=True)
        )
        
        self.service_manager = ServiceManager(config)
        
    async def cleanup_real_environment(self):
        """Clean up real test environment."""
        if self.test_stream_process:
            self.test_stream_process.terminate()
            try:
                self.test_stream_process.wait(timeout=5)
            except subprocess.TimeoutExpired:
                self.test_stream_process.kill()
        
        if self.mediamtx_controller:
            await self.mediamtx_controller.stop()
        
        if self.service_manager:
            await self.service_manager.shutdown()
            
        if self.temp_dir and os.path.exists(self.temp_dir):
            import shutil
            shutil.rmtree(self.temp_dir)
    
    def create_test_video_stream(self, stream_name: str, duration: int = 10) -> str:
        """Create a test video stream using FFmpeg."""
        try:
            # Create test video file
            test_video_path = f"{self.temp_dir}/test_video.mp4"
            
            # Generate test video using FFmpeg
            cmd = [
                "ffmpeg", "-f", "lavfi", "-i", "testsrc=duration=10:size=640x480:rate=30",
                "-f", "lavfi", "-i", "sine=frequency=1000:duration=10",
                "-c:v", "libx264", "-c:a", "aac", "-shortest",
                "-y", test_video_path
            ]
            
            subprocess.run(cmd, check=True, capture_output=True)
            
            # Stream the video to RTSP using FFmpeg
            rtsp_url = f"rtsp://127.0.0.1:8554/{stream_name}"
            
            stream_cmd = [
                "ffmpeg", "-re", "-i", test_video_path,
                "-c:v", "libx264", "-preset", "ultrafast", "-tune", "zerolatency",
                "-c:a", "aac", "-f", "rtsp", "-rtsp_transport", "tcp",
                rtsp_url
            ]
            
            self.test_stream_process = subprocess.Popen(
                stream_cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE
            )
            
            return rtsp_url
            
        except Exception as e:
            self.system_issues.append(f"Test video stream creation failed: {str(e)}")
            raise
    
    async def validate_rtsp_stream_creation(self, stream_name: str) -> Dict[str, Any]:
        """Validate real RTSP stream creation and availability."""
        try:
            # Create test stream
            rtsp_url = self.create_test_video_stream(stream_name)
            
            # Wait for stream to be available
            await asyncio.sleep(3)
            
            # Check if stream is registered with MediaMTX
            streams = await self.mediamtx_controller.list_streams()
            stream_registered = stream_name in streams
            
            # Test stream URL accessibility
            stream_url_valid = await self.mediamtx_controller.validate_stream_url(rtsp_url)
            
            # Get stream info
            stream_info = await self.mediamtx_controller.get_stream_info(stream_name)
            
            return {
                "stream_created": stream_registered,
                "stream_url": rtsp_url,
                "streams_list": streams,
                "url_valid": stream_url_valid,
                "stream_info": stream_info
            }
            
        except Exception as e:
            self.system_issues.append(f"RTSP stream creation failed: {str(e)}")
            raise
    
    async def validate_rtsp_stream_playback(self, stream_name: str) -> Dict[str, Any]:
        """Validate real RTSP stream playback capabilities."""
        try:
            rtsp_url = f"rtsp://127.0.0.1:8554/{stream_name}"
            
            # Test stream playback using OpenCV
            cap = cv2.VideoCapture(rtsp_url)
            
            if not cap.isOpened():
                raise Exception("Failed to open RTSP stream for playback")
            
            # Read a few frames to validate stream
            frames_read = 0
            max_frames = 30  # Read up to 30 frames
            
            for _ in range(max_frames):
                ret, frame = cap.read()
                if ret:
                    frames_read += 1
                    # Validate frame properties
                    if frame is not None and frame.size > 0:
                        height, width, channels = frame.shape
                        if width > 0 and height > 0 and channels > 0:
                            continue
                else:
                    break
                
                await asyncio.sleep(0.1)  # Small delay between frames
            
            cap.release()
            
            # Validate playback results
            playback_successful = frames_read > 0
            
            return {
                "playback_successful": playback_successful,
                "frames_read": frames_read,
                "stream_url": rtsp_url,
                "max_frames_attempted": max_frames
            }
            
        except Exception as e:
            self.system_issues.append(f"RTSP stream playback failed: {str(e)}")
            raise
    
    async def validate_multiple_stream_handling(self) -> Dict[str, Any]:
        """Validate handling of multiple concurrent RTSP streams."""
        try:
            stream_names = ["test_stream_1", "test_stream_2", "test_stream_3"]
            stream_results = {}
            
            # Create multiple streams
            for stream_name in stream_names:
                result = await self.validate_rtsp_stream_creation(stream_name)
                stream_results[stream_name] = result
                
                # Small delay between stream creation
                await asyncio.sleep(1)
            
            # Validate all streams are available
            all_streams = await self.mediamtx_controller.list_streams()
            all_streams_available = all(name in all_streams for name in stream_names)
            
            # Test playback for each stream
            playback_results = {}
            for stream_name in stream_names:
                playback_result = await self.validate_rtsp_stream_playback(stream_name)
                playback_results[stream_name] = playback_result
            
            return {
                "stream_creation": stream_results,
                "all_streams_available": all_streams_available,
                "playback_results": playback_results,
                "total_streams": len(stream_names)
            }
            
        except Exception as e:
            self.system_issues.append(f"Multiple stream handling failed: {str(e)}")
            raise
    
    async def validate_stream_quality_metrics(self, stream_name: str) -> Dict[str, Any]:
        """Validate stream quality metrics and performance."""
        try:
            rtsp_url = f"rtsp://127.0.0.1:8554/{stream_name}"
            
            # Measure stream startup time
            start_time = time.time()
            
            cap = cv2.VideoCapture(rtsp_url)
            if not cap.isOpened():
                raise Exception("Failed to open stream for quality metrics")
            
            startup_time = time.time() - start_time
            
            # Measure frame rate and quality
            frame_times = []
            frame_sizes = []
            
            for _ in range(30):  # Measure 30 frames
                frame_start = time.time()
                ret, frame = cap.read()
                frame_end = time.time()
                
                if ret and frame is not None:
                    frame_times.append(frame_end - frame_start)
                    frame_sizes.append(frame.size)
                
                await asyncio.sleep(0.1)
            
            cap.release()
            
            # Calculate metrics
            avg_frame_time = sum(frame_times) / len(frame_times) if frame_times else 0
            avg_frame_rate = 1.0 / avg_frame_time if avg_frame_time > 0 else 0
            avg_frame_size = sum(frame_sizes) / len(frame_sizes) if frame_sizes else 0
            
            return {
                "startup_time": startup_time,
                "avg_frame_time": avg_frame_time,
                "avg_frame_rate": avg_frame_rate,
                "avg_frame_size": avg_frame_size,
                "frames_measured": len(frame_times),
                "stream_url": rtsp_url
            }
            
        except Exception as e:
            self.system_issues.append(f"Stream quality metrics failed: {str(e)}")
            raise
    
    async def run_comprehensive_rtsp_validation(self) -> Dict[str, Any]:
        """Run comprehensive RTSP stream handling validation."""
        try:
            await self.setup_real_environment()
            
            # Start MediaMTX
            await self.mediamtx_controller.start()
            await asyncio.sleep(2)
            
            # Execute all validation steps
            results = {
                "single_stream_creation": await self.validate_rtsp_stream_creation("test_stream"),
                "single_stream_playback": await self.validate_rtsp_stream_playback("test_stream"),
                "multiple_stream_handling": await self.validate_multiple_stream_handling(),
                "stream_quality_metrics": await self.validate_stream_quality_metrics("test_stream"),
                "system_issues": self.system_issues
            }
            
            self.test_results = results
            return results
            
        finally:
            await self.cleanup_real_environment()


@pytest.mark.pdr
@pytest.mark.asyncio
class TestRealRTSPStreamHandling:
    """Critical prototype tests for real RTSP stream handling."""
    
    def setup_method(self):
        """Set up prototype for each test method."""
        self.prototype = RealRTSPStreamHandlingPrototype()
    
    async def teardown_method(self):
        """Clean up after each test method."""
        if hasattr(self, 'prototype'):
            await self.prototype.cleanup_real_environment()
    
    async def test_rtsp_stream_real_creation(self):
        """Test real RTSP stream creation and registration."""
        await self.prototype.setup_real_environment()
        
        try:
            # Start MediaMTX first
            await self.prototype.mediamtx_controller.start()
            await asyncio.sleep(2)
            
            result = await self.prototype.validate_rtsp_stream_creation("test_stream")
            
            # Validate results
            assert result["stream_created"] is True, "RTSP stream creation failed"
            assert result["url_valid"] is True, "RTSP stream URL invalid"
            assert "test_stream" in result["streams_list"], "Stream not registered"
            
            print(f"✅ RTSP stream creation validation: {result}")
            
        finally:
            await self.prototype.cleanup_real_environment()
    
    async def test_rtsp_stream_real_playback(self):
        """Test real RTSP stream playback capabilities."""
        await self.prototype.setup_real_environment()
        
        try:
            # Start MediaMTX first
            await self.prototype.mediamtx_controller.start()
            await asyncio.sleep(2)
            
            # Create stream first
            await self.prototype.validate_rtsp_stream_creation("test_stream")
            
            # Test playback
            result = await self.prototype.validate_rtsp_stream_playback("test_stream")
            
            # Validate results
            assert result["playback_successful"] is True, "RTSP stream playback failed"
            assert result["frames_read"] > 0, "No frames read from stream"
            
            print(f"✅ RTSP stream playback validation: {result}")
            
        finally:
            await self.prototype.cleanup_real_environment()
    
    async def test_multiple_rtsp_streams_real_handling(self):
        """Test handling of multiple concurrent RTSP streams."""
        await self.prototype.setup_real_environment()
        
        try:
            # Start MediaMTX first
            await self.prototype.mediamtx_controller.start()
            await asyncio.sleep(2)
            
            result = await self.prototype.validate_multiple_stream_handling()
            
            # Validate results
            assert result["all_streams_available"] is True, "Not all streams available"
            assert result["total_streams"] == 3, "Incorrect number of streams"
            
            # Validate playback for each stream
            for stream_name, playback_result in result["playback_results"].items():
                assert playback_result["playback_successful"] is True, f"Playback failed for {stream_name}"
            
            print(f"✅ Multiple RTSP streams validation: {result}")
            
        finally:
            await self.prototype.cleanup_real_environment()
    
    async def test_rtsp_stream_quality_metrics(self):
        """Test RTSP stream quality metrics and performance."""
        await self.prototype.setup_real_environment()
        
        try:
            # Start MediaMTX first
            await self.prototype.mediamtx_controller.start()
            await asyncio.sleep(2)
            
            # Create stream first
            await self.prototype.validate_rtsp_stream_creation("test_stream")
            
            # Test quality metrics
            result = await self.prototype.validate_stream_quality_metrics("test_stream")
            
            # Validate results
            assert result["startup_time"] < 5.0, "Stream startup too slow"
            assert result["avg_frame_rate"] > 0, "Invalid frame rate"
            assert result["frames_measured"] > 0, "No frames measured"
            
            print(f"✅ RTSP stream quality metrics: {result}")
            
        finally:
            await self.prototype.cleanup_real_environment()
    
    async def test_comprehensive_rtsp_validation(self):
        """Test comprehensive RTSP stream handling validation."""
        result = await self.prototype.run_comprehensive_rtsp_validation()
        
        # Validate comprehensive results
        assert len(result["system_issues"]) == 0, f"System issues found: {result['system_issues']}"
        assert result["single_stream_creation"]["stream_created"] is True, "Comprehensive stream creation failed"
        assert result["single_stream_playback"]["playback_successful"] is True, "Comprehensive playback failed"
        assert result["multiple_stream_handling"]["all_streams_available"] is True, "Comprehensive multiple streams failed"
        assert result["stream_quality_metrics"]["avg_frame_rate"] > 0, "Comprehensive quality metrics failed"
        
        print(f"✅ Comprehensive RTSP validation: {result}")
        
        # Log results for evidence
        with open("/tmp/pdr_rtsp_handling_results.json", "w") as f:
            json.dump(result, f, indent=2, default=str)
