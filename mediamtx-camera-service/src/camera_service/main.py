"""
Main entry point for the camera service application.
"""

import asyncio
import logging
import signal
import sys
from pathlib import Path

from .config import load_config
from .logging_config import setup_logging
from .service_manager import ServiceManager

try:
    from importlib.metadata import version, PackageNotFoundError
except ImportError:
    # For Python <3.8 compatibility
    from importlib_metadata import version, PackageNotFoundError

def get_version():
    try:
        return version("mediamtx-camera-service")
    except PackageNotFoundError:
        return "unknown"

async def main():
    """Main application entry point."""
    try:
        # Load configuration
        config = load_config()
        
        # Setup logging
        setup_logging(config.logging)
        logger = logging.getLogger(__name__)

        service_version = get_version()
        logger.info(f"Starting MediaMTX Camera Service v{service_version}")
        
        # Create and start service manager
        service_manager = ServiceManager(config)
        
        # Setup signal handlers for graceful shutdown
        def signal_handler():
            logger.info("Received shutdown signal")
            asyncio.create_task(service_manager.stop())
        
        if sys.platform != 'win32':
            loop = asyncio.get_event_loop()
            for sig in (signal.SIGTERM, signal.SIGINT):
                loop.add_signal_handler(sig, signal_handler)
        
        # Start services
        await service_manager.start()
        
        logger.info("Camera service started successfully")
        
        # Wait for shutdown
        await service_manager.wait_for_shutdown()
        
    except KeyboardInterrupt:
        logger.info("Received keyboard interrupt")
    except Exception as e:
        logger.error(f"Fatal error: {e}", exc_info=True)
        sys.exit(1)
    finally:
        logger.info("Camera service stopped")


if __name__ == "__main__":
    asyncio.run(main())
