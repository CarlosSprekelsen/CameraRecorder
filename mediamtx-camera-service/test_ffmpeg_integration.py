#!/usr/bin/env python3
"""
Test script for MediaMTX FFmpeg integration implementation.
"""

import asyncio
import logging
import sys
import os
sys.path.append(os.path.join(os.path.dirname(__file__), 'src'))

from camera_service.config import Config
from camera_service.service_manager import ServiceManager

# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

async def test_ffmpeg_integration():
    """Test the MediaMTX FFmpeg integration implementation."""
    
    logger.info("Starting MediaMTX FFmpeg integration test")
    
    try:
        # Create configuration
        config = Config()
        
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

if __name__ == "__main__":
    asyncio.run(test_ffmpeg_integration())
