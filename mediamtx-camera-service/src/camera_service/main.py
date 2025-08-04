"""
Main entry point for the camera service application.

Provides robust startup/shutdown with proper error handling, graceful signal
handling, and clear failure modes as required by architecture principles.
"""

import asyncio
import logging
import signal
import sys
from typing import Optional

from .config import load_config
from .logging_config import setup_logging
from .service_manager import ServiceManager

try:
    from importlib.metadata import version, PackageNotFoundError
except ImportError:
    # For Python <3.8 compatibility
    from importlib_metadata import version, PackageNotFoundError


def get_version() -> str:
    """Get the service version from package metadata."""
    try:
        return version("mediamtx-camera-service")
    except PackageNotFoundError:
        return "unknown"


class ServiceCoordinator:
    """
    Coordinates service lifecycle with proper error handling and cleanup.

    Manages startup sequence, graceful shutdown, and cleanup of partially
    initialized state on failure.
    """

    def __init__(self):
        self.service_manager: Optional[ServiceManager] = None
        self.logger: Optional[logging.Logger] = None
        self._shutdown_requested = asyncio.Event()

    async def startup(self) -> None:
        """
        Initialize all service components with proper error handling.

        Raises:
            SystemExit: On unrecoverable startup failure
        """
        try:
            # Step 1: Load configuration
            config = load_config()

            # Step 2: Setup logging
            setup_logging(config.logging)
            self.logger = logging.getLogger(__name__)

            service_version = get_version()
            self.logger.info(f"Starting MediaMTX Camera Service v{service_version}")

            # Step 3: Create service manager
            self.service_manager = ServiceManager(config)

            # Step 4: Setup signal handlers for graceful shutdown
            self._setup_signal_handlers()

            # Step 5: Start services
            await self.service_manager.start()

            self.logger.info("Camera service started successfully")

        except KeyboardInterrupt:
            self.logger.info("Received keyboard interrupt during startup")
            raise
        except Exception as e:
            if self.logger:
                self.logger.error(f"Fatal startup error: {e}", exc_info=True)
            else:
                # Fallback logging if setup_logging failed
                print(
                    f"Fatal startup error (logging not available): {e}", file=sys.stderr
                )

            # Cleanup any partially initialized state
            await self._cleanup_partial_state()
            raise SystemExit(1) from e

    async def shutdown(self) -> None:
        """Gracefully shutdown all service components."""
        if self.logger:
            self.logger.info("Initiating graceful shutdown")

        if self.service_manager:
            try:
                await self.service_manager.stop()
                if self.logger:
                    self.logger.info("Service manager stopped successfully")
            except Exception as e:
                if self.logger:
                    self.logger.error(
                        f"Error during service manager shutdown: {e}", exc_info=True
                    )

        if self.logger:
            self.logger.info("Camera service stopped")

    async def wait_for_shutdown(self) -> None:
        """Wait for shutdown signal or service completion."""
        if not self.service_manager:
            raise RuntimeError("Service not started")

        # Wait for either shutdown signal or service completion
        shutdown_task = asyncio.create_task(self._shutdown_requested.wait())
        service_task = asyncio.create_task(self.service_manager.wait_for_shutdown())

        try:
            done, pending = await asyncio.wait(
                [shutdown_task, service_task], return_when=asyncio.FIRST_COMPLETED
            )

            # Cancel any remaining tasks
            for task in pending:
                task.cancel()
                try:
                    await task
                except asyncio.CancelledError:
                    pass

        except Exception as e:
            if self.logger:
                self.logger.error(f"Error waiting for shutdown: {e}", exc_info=True)

    def _setup_signal_handlers(self) -> None:
        """Setup signal handlers for graceful shutdown."""
        if sys.platform == "win32":
            # Windows doesn't support signal handlers in asyncio
            if self.logger:
                self.logger.warning("Signal handlers not supported on Windows")
            return

        def signal_handler(signum: int) -> None:
            """Handle shutdown signals."""
            signal_name = signal.Signals(signum).name
            if self.logger:
                self.logger.info(f"Received {signal_name} signal")

            # Signal shutdown event (thread-safe)
            self._shutdown_requested.set()

        try:
            loop = asyncio.get_event_loop()
            for sig in (signal.SIGTERM, signal.SIGINT):
                loop.add_signal_handler(sig, signal_handler, sig)

            if self.logger:
                self.logger.debug("Signal handlers configured for SIGTERM and SIGINT")

        except Exception as e:
            if self.logger:
                self.logger.warning(f"Failed to setup signal handlers: {e}")

    async def _cleanup_partial_state(self) -> None:
        """Cleanup any partially initialized state on startup failure."""
        if self.service_manager:
            try:
                await self.service_manager.stop()
            except Exception as e:
                if self.logger:
                    self.logger.error(f"Error during partial state cleanup: {e}")


async def main() -> None:
    """Main application entry point with robust error handling."""
    coordinator = ServiceCoordinator()

    try:
        # Initialize and start all services
        await coordinator.startup()

        # Wait for shutdown signal or service completion
        await coordinator.wait_for_shutdown()

    except KeyboardInterrupt:
        # Handle Ctrl+C gracefully
        if coordinator.logger:
            coordinator.logger.info("Received keyboard interrupt")
    except SystemExit:
        # Re-raise SystemExit to preserve exit code
        raise
    except Exception as e:
        # Catch any unexpected errors
        if coordinator.logger:
            coordinator.logger.error(f"Unexpected error in main: {e}", exc_info=True)
        else:
            print(f"Unexpected error: {e}", file=sys.stderr)
        sys.exit(1)
    finally:
        # Always attempt graceful shutdown
        try:
            await coordinator.shutdown()
        except Exception as e:
            if coordinator.logger:
                coordinator.logger.error(f"Error during final shutdown: {e}")
            else:
                print(f"Shutdown error: {e}", file=sys.stderr)


if __name__ == "__main__":
    asyncio.run(main())
