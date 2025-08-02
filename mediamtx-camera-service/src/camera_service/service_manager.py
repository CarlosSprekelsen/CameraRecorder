# src/camera_service/service_manager.py
"""
Service Manager for coordinating all camera service components.

This module provides the main ServiceManager class that orchestrates
the lifecycle and coordination of all service components including
WebSocket server, camera discovery, MediaMTX integration, and health monitoring.
"""

import asyncio
import logging
from typing import Optional

from .config import Config
from ..mediamtx_wrapper.controller import MediaMTXController
from ..camera_discovery.hybrid_monitor import CameraEventData, CameraEvent, CameraEventHandler


class ServiceManager(CameraEventHandler):
    """
    Main service orchestrator that manages the lifecycle of all camera service components.
    
    The ServiceManager coordinates between the WebSocket JSON-RPC Server, Camera Discovery
    Monitor, MediaMTX Controller, and Health & Monitoring subsystems as defined in the
    architecture overview.
    
    Implements CameraEventHandler to receive camera connect/disconnect events and
    coordinate with MediaMTX stream management.
    """

    def __init__(self, config: Config) -> None:
        """
        Initialize the service manager with configuration.
        
        Args:
            config: Configuration object containing all service settings
        """
        self._config = config
        self._logger = logging.getLogger(__name__)
        self._shutdown_event: Optional[asyncio.Event] = None
        self._running = False
        
        # Component references - TODO: Initialize actual components
        self._websocket_server = None
        self._camera_monitor = None
        self._mediamtx_controller: Optional[MediaMTXController] = None
        self._health_monitor = None

    async def start(self) -> None:
        """
        Start all service components in the correct order.
        
        Initializes and starts:
        1. MediaMTX Controller
        2. Camera Discovery Monitor  
        3. Health & Monitoring
        4. WebSocket JSON-RPC Server
        
        Raises:
            RuntimeError: If service is already running or startup fails
        """
        if self._running:
            raise RuntimeError("Service manager is already running")
            
        self._logger.info("Starting camera service components")
        self._shutdown_event = asyncio.Event()
        
        try:
            # Initialize and start MediaMTX Controller
            await self._start_mediamtx_controller()
            
            # Initialize and start Camera Discovery Monitor
            await self._start_camera_monitor()
            
            # Initialize and start Health & Monitoring
            await self._start_health_monitor()
            
            # Initialize and start WebSocket JSON-RPC Server
            await self._start_websocket_server()
            
            self._running = True
            self._logger.info("All camera service components started successfully")
            
        except Exception as e:
            self._logger.error(f"Failed to start service components: {e}")
            await self.stop()
            raise

    async def stop(self) -> None:
        """
        Stop all service components in reverse order.
        
        Gracefully stops components in reverse startup order:
        1. WebSocket JSON-RPC Server
        2. Health & Monitoring
        3. Camera Discovery Monitor
        4. MediaMTX Controller
        """
        if not self._running:
            return
            
        self._logger.info("Stopping camera service components")
        
        try:
            # Stop components in reverse order
            # Stop WebSocket JSON-RPC Server
            await self._stop_websocket_server()
            
            # Stop Health & Monitoring
            await self._stop_health_monitor()
            
            # Stop Camera Discovery Monitor
            await self._stop_camera_monitor()
            
            # Stop MediaMTX Controller
            await self._stop_mediamtx_controller()
            
            self._running = False
            
            if self._shutdown_event:
                self._shutdown_event.set()
                
            self._logger.info("All camera service components stopped")
            
        except Exception as e:
            self._logger.error(f"Error during service shutdown: {e}")
            raise

    async def wait_for_shutdown(self) -> None:
        """
        Wait for shutdown signal.
        
        Blocks until the service receives a shutdown signal or stop() is called.
        """
        if not self._shutdown_event:
            raise RuntimeError("Service not started")
            
        await self._shutdown_event.wait()

    async def handle_camera_event(self, event_data: CameraEventData) -> None:
        """
        Handle camera connect/disconnect events from the camera monitor.
        
        Coordinates MediaMTX stream configuration updates based on camera events.
        
        Args:
            event_data: Camera event information including device path and type
        """
        self._logger.info(
            f"Handling camera event: {event_data.event_type.value} - {event_data.device_path}"
        )
        
        try:
            if event_data.event_type == CameraEvent.CONNECTED:
                await self._handle_camera_connected(event_data)
            elif event_data.event_type == CameraEvent.DISCONNECTED:
                await self._handle_camera_disconnected(event_data)
            elif event_data.event_type == CameraEvent.STATUS_CHANGED:
                await self._handle_camera_status_changed(event_data)
                
        except Exception as e:
            self._logger.error(f"Error handling camera event: {e}", exc_info=True)

    async def _handle_camera_connected(self, event_data: CameraEventData) -> None:
        """
        Handle camera connection event.

        Creates MediaMTX stream configuration for the newly connected camera and
        triggers camera_status_update notification as specified in the architecture overview.

        Args:
            event_data: Camera connection event data

        Architecture Reference:
            docs/architecture/overview.md: Camera Discovery Monitor, MediaMTX Controller, WebSocket JSON-RPC Server.
            - Must create stream config, update MediaMTX, and notify clients via camera_status_update.

        # TODO: [CRITICAL] Align _handle_camera_connected signature and responsibilities with architecture overview.
        # Description: This stub is required for API alignment. Reference: IV&V finding 1.1, Story S1.
        # Rationale: Must trigger camera_status_update notification after stream creation.
        # STOPPED: Do not implement notification logic until notification handler interface is finalized.
        """
        self._logger.debug(f"Creating stream for connected camera: {event_data.device_path}")
        # TODO: Create stream config and call MediaMTXController.create_stream()
        # TODO: Prepare camera_status_update notification payload
        # TODO: Call notification handler (STOPPED: Await finalized notification interface)

    async def _handle_camera_disconnected(self, event_data: CameraEventData) -> None:
        """
        Handle camera disconnection event.

        Removes MediaMTX stream configuration for the disconnected camera and
        triggers camera_status_update notification as specified in the architecture overview.

        Args:
            event_data: Camera disconnection event data

        Architecture Reference:
            docs/architecture/overview.md: Camera Discovery Monitor, MediaMTX Controller, WebSocket JSON-RPC Server.
            - Must remove stream config, update MediaMTX, and notify clients via camera_status_update.

        # TODO: [CRITICAL] Align _handle_camera_disconnected signature and responsibilities with architecture overview.
        # Description: This stub is required for API alignment. Reference: IV&V finding 1.1, Story S1.
        # Rationale: Must trigger camera_status_update notification after stream removal.
        # STOPPED: Do not implement notification logic until notification handler interface is finalized.
        """
        self._logger.debug(f"Removing stream for disconnected camera: {event_data.device_path}")
        # TODO: Remove stream config and call MediaMTXController.delete_stream()
        # TODO: Prepare camera_status_update notification payload
        # TODO: Call notification handler (STOPPED: Await finalized notification interface)

    async def _handle_camera_status_changed(self, event_data: CameraEventData) -> None:
        """
        Handle camera status change event.

        Updates MediaMTX stream configuration based on camera status changes and
        triggers camera_status_update notification as specified in the architecture overview.

        Args:
            event_data: Camera status change event data

        Architecture Reference:
            docs/architecture/overview.md: Camera Discovery Monitor, MediaMTX Controller, WebSocket JSON-RPC Server.
            - Must update stream config if needed and notify clients via camera_status_update.

        # TODO: [CRITICAL] Align _handle_camera_status_changed signature and responsibilities with architecture overview.
        # Description: This stub is required for API alignment. Reference: IV&V finding 1.1, Story S1.
        # Rationale: Must trigger camera_status_update notification after stream update.
        # STOPPED: Do not implement notification logic until notification handler interface is finalized.
        """
        self._logger.debug(f"Handling status change for camera: {event_data.device_path}")
        # TODO: Update stream config and call MediaMTXController.update_stream() if needed
        # TODO: Prepare camera_status_update notification payload
        # TODO: Call notification handler (STOPPED: Await finalized notification interface)

    async def _start_mediamtx_controller(self) -> None:
        """Start the MediaMTX REST API controller component."""
        self._logger.debug("Starting MediaMTX controller")
        
        try:
            # Initialize MediaMTX Controller with configuration
            self._mediamtx_controller = MediaMTXController(
                host=self._config.mediamtx.host,
                api_port=self._config.mediamtx.api_port,
                rtsp_port=self._config.mediamtx.rtsp_port,
                webrtc_port=self._config.mediamtx.webrtc_port,
                hls_port=self._config.mediamtx.hls_port,
                config_path=self._config.mediamtx.config_path,
                recordings_path=self._config.mediamtx.recordings_path,
                snapshots_path=self._config.mediamtx.snapshots_path
            )
            
            # Start MediaMTX controller (initializes HTTP client and monitoring)
            await self._mediamtx_controller.start()
            self._logger.info("MediaMTX controller started successfully")
            
            # Verify MediaMTX connectivity and health
            try:
                health_status = await self._mediamtx_controller.health_check()
                if health_status.get("status") == "unknown":
                    self._logger.warning("MediaMTX health check returned unknown status - continuing with startup")
                else:
                    self._logger.info(f"MediaMTX connectivity verified: {health_status}")
            except Exception as health_error:
                self._logger.warning(f"MediaMTX health check failed during startup: {health_error}")
                # Continue startup - health monitoring will handle recovery
            
            # Setup MediaMTX configuration management
            # Validate that required directories exist for recordings and snapshots
            import os
            os.makedirs(self._config.mediamtx.recordings_path, exist_ok=True)
            os.makedirs(self._config.mediamtx.snapshots_path, exist_ok=True)
            self._logger.debug(f"Verified MediaMTX directories: recordings={self._config.mediamtx.recordings_path}, snapshots={self._config.mediamtx.snapshots_path}")
            
            self._logger.info("MediaMTX controller initialization completed")
            
        except Exception as e:
            self._logger.error(f"Failed to start MediaMTX controller: {e}")
            # Cleanup on failure
            if self._mediamtx_controller:
                try:
                    await self._mediamtx_controller.stop()
                except Exception as cleanup_error:
                    self._logger.error(f"Error during MediaMTX controller cleanup: {cleanup_error}")
                self._mediamtx_controller = None
            raise

    async def _start_camera_monitor(self) -> None:
        """Start the camera discovery and monitoring component."""
        # TODO: Initialize Camera Discovery Monitor with config
        # TODO: Register this ServiceManager as camera event handler
        # TODO: Setup hybrid udev + polling camera detection
        # TODO: Start camera capability detection
        pass

    async def _start_health_monitor(self) -> None:
        """Start the health monitoring and recovery component."""
        # TODO: Initialize Health Monitor with config
        # TODO: Setup service health checks and circuit breaker
        # TODO: Start resource usage monitoring
        pass

    async def _start_websocket_server(self) -> None:
        """Start the WebSocket JSON-RPC server component."""
        # TODO: Initialize WebSocket server with config
        # TODO: Setup JSON-RPC method handlers
        # TODO: Start client connection management
        pass

    async def _stop_websocket_server(self) -> None:
        """Stop the WebSocket JSON-RPC server component."""
        # TODO: Gracefully close client connections
        # TODO: Stop WebSocket server
        pass

    async def _stop_health_monitor(self) -> None:
        """Stop the health monitoring component."""
        # TODO: Stop health checks and monitoring
        # TODO: Cleanup monitoring resources
        pass

    async def _stop_camera_monitor(self) -> None:
        """Stop the camera discovery and monitoring component."""
        # TODO: Unregister camera event handler
        # TODO: Stop camera monitoring
        # TODO: Cleanup camera resources and streams
        pass

    async def _stop_mediamtx_controller(self) -> None:
        """Stop the MediaMTX controller component."""
        if self._mediamtx_controller:
            # TODO: Cleanup MediaMTX streams and configuration
            await self._mediamtx_controller.stop()
            self._mediamtx_controller = None

    def _get_stream_name_from_device_path(self, device_path: str) -> str:
        """
        Extract stream name from camera device path.
        
        Args:
            device_path: Camera device path (e.g., /dev/video0)
            
        Returns:
            Stream name for MediaMTX (e.g., camera0)
        """
        # TODO: Parse device path and extract device number
        # TODO: Return consistent stream name format
        # Example: /dev/video0 -> camera0
        return "camera0"  # Placeholder

    @property
    def is_running(self) -> bool:
        """Check if the service manager is currently running."""
        return self._running

    def get_status(self) -> dict:
        """
        Get current status of all service components.
        
        Returns:
            Dictionary with status information for each component
        """
        # TODO: Collect status from all components
        # TODO: Return comprehensive service status
        return {
            "running": self._running,
            "websocket_server": "not_implemented",
            "camera_monitor": "not_implemented", 
            "mediamtx_controller": "started" if self._mediamtx_controller else "not_started",
            "health_monitor": "not_implemented"
        }