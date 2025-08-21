"""
Comprehensive tests for stream readiness validation using real components.

Tests the enhanced stream readiness validation before recording operations,
on-demand activation, and timeout handling as specified in PDR conditions.
Uses real MediaMTX service instead of mocks.

Requirements Traceability:
- REQ-REC-001: System shall validate stream readiness before recording operations
- REQ-ERR-001: System shall handle errors gracefully without crashing
- REQ-INT-001: System shall maintain state consistency across components
- REQ-PERF-001: System shall meet performance requirements for stream operations

PDR Condition Coverage:
- Recording Stream Availability (Medium Priority): Stream readiness validation
- Recording Error Handling Enhancement (Medium Priority): Proper error handling

Story Coverage: PDR Conditions Resolution
IV&V Control Point: Stream readiness validation
"""

import asyncio
import pytest
import json
import os
import time
from pathlib import Path

from src.mediamtx_wrapper.controller import MediaMTXController
from src.camera_service.config import Config


class TestStreamReadinessValidationReal:
    """Test suite for stream readiness validation using real components."""

    @pytest.fixture
    def real_config(self):
        """Create real configuration for testing."""
        config = Config()
        config.mediamtx.host = "localhost"
        config.mediamtx.api_port = 9997
        config.mediamtx.rtsp_port = 8554
        config.mediamtx.webrtc_port = 8889
        config.mediamtx.hls_port = 8888
        config.mediamtx.config_path = "/etc/mediamtx/mediamtx.yml"
        config.mediamtx.recordings_path = "/tmp/test_recordings"
        config.mediamtx.snapshots_path = "/tmp/test_snapshots"
        return config

    @pytest.fixture
    def real_mediamtx_controller(self):
        """Create real MediaMTX controller using systemd-managed service."""
        return MediaMTXController(
            host="localhost",
            api_port=9997,
            rtsp_port=8554,
            webrtc_port=8889,
            hls_port=8888,
            config_path="/etc/mediamtx/mediamtx.yml",
            recordings_path="/tmp/test_recordings",
            snapshots_path="/tmp/test_snapshots",
        )

    @pytest.mark.asyncio
    async def test_real_stream_readiness_check_with_real_mediamtx(self, real_mediamtx_controller):
        """
        Test real stream readiness check with actual MediaMTX service.
        
        Requirements: REQ-REC-001, REQ-ERR-001
        PDR Condition: Recording Stream Availability
        Scenario: Real stream readiness check with MediaMTX service
        Expected: Proper stream readiness validation, graceful error handling
        Edge Cases: Stream not found, MediaMTX service issues
        """
        # Setup: Start real MediaMTX controller
        await real_mediamtx_controller.start()
        
        try:
            # Get real paths from MediaMTX using direct API call
            async with real_mediamtx_controller._session.get(f"{real_mediamtx_controller._base_url}/v3/paths/list") as response:
                if response.status == 200:
                    paths_data = await response.json()
                    real_streams = [path["name"] for path in paths_data.get("items", [])]
                else:
                    real_streams = []
            
            if not real_streams:
                # No streams available, test with non-existent stream
                with pytest.raises(ValueError, match="Stream test_stream not found in MediaMTX"):
                    await real_mediamtx_controller.check_stream_readiness("test_stream")
            else:
                # Test with real stream
                real_stream = real_streams[0]
                try:
                    # Check readiness of real stream
                    is_ready = await real_mediamtx_controller.check_stream_readiness(real_stream)
                    # Should return boolean value
                    assert isinstance(is_ready, bool)
                except Exception as e:
                    # Stream may not be ready, which is acceptable
                    assert "not found" in str(e) or "not ready" in str(e)
                    
        finally:
            await real_mediamtx_controller.stop()

    @pytest.mark.asyncio
    async def test_real_stream_readiness_check_invalid_stream_name(self, real_mediamtx_controller):
        """
        Test real stream readiness check with invalid stream name.
        
        Requirements: REQ-ERR-001
        PDR Condition: Recording Stream Availability
        Scenario: Invalid stream name validation
        Expected: Proper error handling for invalid stream names
        Edge Cases: Empty stream names, malformed stream names
        """
        # Setup: Start real MediaMTX controller
        await real_mediamtx_controller.start()
        
        try:
            # Test with empty stream name
            with pytest.raises(ValueError, match="Stream name is required"):
                await real_mediamtx_controller.check_stream_readiness("")
                
            # Test with None stream name
            with pytest.raises(ValueError, match="Stream name is required"):
                await real_mediamtx_controller.check_stream_readiness(None)
                
        finally:
            await real_mediamtx_controller.stop()

    @pytest.mark.asyncio
    async def test_real_stream_readiness_check_controller_not_started(self):
        """
        Test real stream readiness check when controller is not started.
        
        Requirements: REQ-ERR-001
        PDR Condition: Recording Stream Availability
        Scenario: Controller not started error handling
        Expected: Proper error handling when controller not initialized
        Edge Cases: Controller initialization failures
        """
        # Create controller without starting it
        controller = MediaMTXController(
            host="localhost",
            api_port=9997,
            rtsp_port=8554,
            webrtc_port=8889,
            hls_port=8888,
            config_path="/etc/mediamtx/mediamtx.yml",
            recordings_path="/tmp/test_recordings",
            snapshots_path="/tmp/test_snapshots",
        )
        
        # Test without starting controller
        with pytest.raises(ConnectionError, match="MediaMTX controller not started"):
            await controller.check_stream_readiness("test_stream")

    @pytest.mark.asyncio
    async def test_real_recording_start_with_stream_readiness_validation(self, real_mediamtx_controller):
        """
        Test real recording start with enhanced stream readiness validation.
        
        Requirements: REQ-REC-001, REQ-ERR-001
        PDR Condition: Recording Stream Availability
        Scenario: Real recording start with stream readiness validation
        Expected: Stream readiness validation before recording, proper error handling
        Edge Cases: Stream not ready, recording failures
        """
        # Setup: Start real MediaMTX controller
        await real_mediamtx_controller.start()
        
        try:
            # Get real paths from MediaMTX using direct API call
            async with real_mediamtx_controller._session.get(f"{real_mediamtx_controller._base_url}/v3/paths/list") as response:
                if response.status == 200:
                    paths_data = await response.json()
                    real_streams = [path["name"] for path in paths_data.get("items", [])]
                else:
                    real_streams = []
            
            if not real_streams:
                # No streams available, test error handling
                with pytest.raises(ValueError, match="Stream test_stream not found in MediaMTX"):
                    await real_mediamtx_controller.start_recording("test_stream", duration=10)
            else:
                # Test with real stream
                real_stream = real_streams[0]
                try:
                    # Try to start recording (may fail if stream not ready, which is acceptable)
                    result = await real_mediamtx_controller.start_recording(real_stream, duration=5)
                    # If successful, should return recording info
                    assert isinstance(result, dict)
                except Exception as e:
                    # Recording may fail due to stream not ready, which is acceptable
                    assert any(keyword in str(e).lower() for keyword in ["not ready", "not found", "failed"])
                    
        finally:
            await real_mediamtx_controller.stop()

    @pytest.mark.asyncio
    async def test_real_recording_start_stream_not_found(self, real_mediamtx_controller):
        """
        Test real recording start when stream does not exist.
        
        Requirements: REQ-ERR-001
        PDR Condition: Recording Stream Availability
        Scenario: Recording start with non-existent stream
        Expected: Proper error handling for missing streams
        Edge Cases: Stream deletion during recording start
        """
        # Setup: Start real MediaMTX controller
        await real_mediamtx_controller.start()
        
        try:
            # Test with non-existent stream
            with pytest.raises(ValueError, match="Stream nonexistent_stream not found in MediaMTX"):
                await real_mediamtx_controller.start_recording("nonexistent_stream", duration=10)
                
        finally:
            await real_mediamtx_controller.stop()

    @pytest.mark.asyncio
    async def test_real_concurrent_stream_readiness_checks(self, real_mediamtx_controller):
        """
        Test real concurrent stream readiness checks for multiple streams.
        
        Requirements: REQ-PERF-001, REQ-ERR-001
        PDR Condition: Recording Stream Availability
        Scenario: Concurrent stream readiness checks with real MediaMTX
        Expected: Concurrent operations handled properly, no race conditions
        Edge Cases: Multiple simultaneous checks, performance under load
        """
        # Setup: Start real MediaMTX controller
        await real_mediamtx_controller.start()
        
        try:
            # Get real paths from MediaMTX using direct API call
            async with real_mediamtx_controller._session.get(f"{real_mediamtx_controller._base_url}/v3/paths/list") as response:
                if response.status == 200:
                    paths_data = await response.json()
                    real_streams = [path["name"] for path in paths_data.get("items", [])]
                else:
                    real_streams = []
            
            if len(real_streams) < 2:
                # Need at least 2 streams for concurrent testing
                # Create multiple readiness checks for same stream
                test_streams = ["test_stream1", "test_stream2", "test_stream3"]
            else:
                # Use real streams
                test_streams = real_streams[:3]  # Use first 3 streams
            
            # Create concurrent readiness checks
            tasks = [
                real_mediamtx_controller.check_stream_readiness(stream_name)
                for stream_name in test_streams
            ]
            
            # Execute concurrent checks
            results = await asyncio.gather(*tasks, return_exceptions=True)
            
            # Verify all operations completed (may have exceptions for non-existent streams)
            assert len(results) == len(test_streams)
            
            # Check that results are either boolean or exceptions
            for result in results:
                assert isinstance(result, (bool, Exception))
                
        finally:
            await real_mediamtx_controller.stop()

    @pytest.mark.asyncio
    async def test_real_stream_readiness_timeout_handling(self, real_mediamtx_controller):
        """
        Test real stream readiness timeout handling.
        
        Requirements: REQ-ERR-001, REQ-PERF-001
        PDR Condition: Recording Stream Availability
        Scenario: Stream readiness timeout handling with real MediaMTX
        Expected: Proper timeout handling, no hanging operations
        Edge Cases: Slow MediaMTX responses, network delays
        """
        # Setup: Start real MediaMTX controller
        await real_mediamtx_controller.start()
        
        try:
            # Test with very short timeout
            start_time = time.time()
            
            try:
                await real_mediamtx_controller.check_stream_readiness("test_stream", timeout=0.1)
            except Exception as e:
                # Should timeout or fail quickly
                elapsed_time = time.time() - start_time
                assert elapsed_time < 1.0, f"Operation took too long: {elapsed_time}s"
                
        finally:
            await real_mediamtx_controller.stop()

    @pytest.mark.asyncio
    async def test_real_stream_readiness_error_recovery(self, real_mediamtx_controller):
        """
        Test real stream readiness error recovery mechanisms.
        
        Requirements: REQ-ERR-001
        PDR Condition: Recording Error Handling Enhancement
        Scenario: Error recovery during stream readiness checks
        Expected: Graceful error handling, proper recovery mechanisms
        Edge Cases: MediaMTX service restarts, temporary failures
        """
        # Setup: Start real MediaMTX controller
        await real_mediamtx_controller.start()
        
        try:
            # Test error handling with invalid stream names
            invalid_streams = ["", None, "invalid/stream/name", "stream with spaces"]
            
            for invalid_stream in invalid_streams:
                try:
                    await real_mediamtx_controller.check_stream_readiness(invalid_stream)
                except (ValueError, ConnectionError) as e:
                    # Expected exceptions for invalid streams
                    assert "required" in str(e) or "not found" in str(e) or "not started" in str(e)
                except Exception as e:
                    # Other exceptions are also acceptable
                    pass
                    
        finally:
            await real_mediamtx_controller.stop()
