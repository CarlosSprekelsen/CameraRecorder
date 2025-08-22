#!/usr/bin/env python3
"""
Test script for MediaMTX FFmpeg integration implementation.

Requirements Traceability:
- REQ-FFMPEG-001: System shall integrate with FFmpeg for video processing
- REQ-FFMPEG-002: System shall handle FFmpeg process lifecycle management
- REQ-FFMPEG-003: System shall validate FFmpeg integration functionality

Test Categories: Integration
"""

import asyncio
import logging
import sys
import os
import tempfile
import pytest
sys.path.append(os.path.join(os.path.dirname(__file__), '..', '..', 'src'))

from camera_service.config import Config, MediaMTXConfig
from camera_service.service_manager import ServiceManager

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

@pytest.mark.asyncio
@pytest.mark.integration
async def test_ffmpeg_integration():
    """Test the MediaMTX FFmpeg integration implementation."""
    
    logger.info("Starting MediaMTX FFmpeg integration test")
    
    try:
        # Create temporary directories for testing
        temp_dir = tempfile.mkdtemp(prefix="ffmpeg_test_")
        recordings_dir = os.path.join(temp_dir, "recordings")
        snapshots_dir = os.path.join(temp_dir, "snapshots")
        os.makedirs(recordings_dir, exist_ok=True)
        os.makedirs(snapshots_dir, exist_ok=True)
        
        # Create configuration with temporary directories
        mediamtx_config = MediaMTXConfig(
            recordings_path=recordings_dir,
            snapshots_path=snapshots_dir
        )
        
        # Use free port for health server to avoid conflicts
        from tests.utils.port_utils import find_free_port
        free_health_port = find_free_port()
        
        config = Config(mediamtx=mediamtx_config, health_port=free_health_port)
        
        # Create service manager
        service_manager = ServiceManager(config)
        
        logger.info("Starting service manager...")
        await service_manager.start()
        
        logger.info("Service manager started successfully")
        logger.info("Waiting for camera events...")
        
        # Wait for some time to allow camera discovery
        await asyncio.sleep(30)
        
        logger.info("Test completed successfully")
        
    except Exception as e:
        logger.error(f"Test failed: {e}")
        raise
    finally:
        # Cleanup
        if 'service_manager' in locals():
            logger.info("Stopping service manager...")
            await service_manager.stop()
            logger.info("Service manager stopped")
        
        # Clean up temporary directories
        if 'temp_dir' in locals() and os.path.exists(temp_dir):
            import shutil
            shutil.rmtree(temp_dir)
            logger.info(f"Cleaned up temporary directory: {temp_dir}")

if __name__ == "__main__":
    asyncio.run(test_ffmpeg_integration())
