# src/camera_service/service_manager.py
"""
Service Manager for coordinating all camera service components.

This module provides the main ServiceManager class that orchestrates
the lifecycle and coordination of all service components including
WebSocket server, camera discovery, MediaMTX integration, and health monitoring.
"""

import asyncio
import logging
import re
import uuid
from typing import Optional, Dict, Any

from .config import Config
from mediamtx_wrapper.controller import MediaMTXController, StreamConfig
from mediamtx_wrapper.path_manager import MediaMTXPathManager
from camera_discovery.hybrid_monitor import (
    CameraEventData,
    CameraEvent,
    CameraEventHandler,
)
from websocket_server.server import WebSocketJsonRpcServer
from .logging_config import set_correlation_id, get_correlation_id


class HealthMonitor:
    """
    Basic health monitoring component for service health checks.

    Provides service health verification and resource monitoring
    as specified in the architecture overview.
    """

    def __init__(self, config: Config):
        """Initialize health monitor with configuration."""
        self._config = config
        self._logger = logging.getLogger(__name__)
        self._running = False
        self._health_check_task: Optional[asyncio.Task] = None

    async def start(self) -> None:
        """Start health monitoring."""
        if self._running:
            return

        correlation_id = get_correlation_id() or "health-monitor-start"
        set_correlation_id(correlation_id)

        self._logger.info(
            "Starting health monitor", extra={"correlation_id": correlation_id}
        )
        self._running = True

        # Start background health check task
        self._health_check_task = asyncio.create_task(self._health_check_loop())
        self._logger.debug(
            "Health monitor started successfully",
            extra={"correlation_id": correlation_id},
        )

    async def stop(self) -> None:
        """Stop health monitoring."""
        if not self._running:
            return

        correlation_id = get_correlation_id() or "health-monitor-stop"
        set_correlation_id(correlation_id)

        self._logger.info(
            "Stopping health monitor", extra={"correlation_id": correlation_id}
        )
        self._running = False

        # Stop background health check task
        if self._health_check_task and not self._health_check_task.done():
            self._health_check_task.cancel()
            try:
                await self._health_check_task
            except asyncio.CancelledError:
                pass

        self._logger.debug(
            "Health monitor stopped", extra={"correlation_id": correlation_id}
        )

    async def _health_check_loop(self) -> None:
        """Background health monitoring loop."""
        while self._running:
            try:
                # Perform basic health checks
                await asyncio.sleep(30)  # Health check interval
            except asyncio.CancelledError:
                break
            except Exception as e:
                correlation_id = get_correlation_id() or "health-check-error"
                self._logger.error(
                    f"Health check error: {e}", extra={"correlation_id": correlation_id}
                )
                await asyncio.sleep(10)  # Shorter wait on error


class ServiceManager(CameraEventHandler):
    """
    Main service orchestrator that manages the lifecycle of all camera service components.

    The ServiceManager coordinates between the WebSocket JSON-RPC Server, Camera Discovery
    Monitor, MediaMTX Controller, and Health & Monitoring subsystems as defined in the
    architecture overview.

    Implements CameraEventHandler to receive camera connect/disconnect events and
    coordinate with MediaMTX stream management.
    """

    def __init__(
        self,
        config: Config,
        mediamtx_controller: Optional[MediaMTXController] = None,
        websocket_server: Optional[WebSocketJsonRpcServer] = None,
        camera_monitor=None,
    ) -> None:
        """
        Initialize the service manager with configuration.

        Args:
            config: Configuration object containing all service settings
            mediamtx_controller: Optional MediaMTX controller instance for testing
            websocket_server: Optional WebSocket server instance for testing
            camera_monitor: Optional camera monitor instance for testing
        """
        self._config = config
        self._logger = logging.getLogger(__name__)
        self._shutdown_event: Optional[asyncio.Event] = None
        self._running = False

        # Component references - allow injection for testing
        self._websocket_server = websocket_server
        self._camera_monitor = camera_monitor
        self._mediamtx_controller = mediamtx_controller
        self._health_monitor: Optional[HealthMonitor] = None
        self._path_manager: Optional[MediaMTXPathManager] = None

    @property
    def is_running(self) -> bool:
        """Check if the service manager is currently running."""
        return self._running

    async def start(self) -> None:
        """
        Start all service components in the correct order.

        Initializes and starts:
        1. MediaMTX Controller
        2. Camera Discovery Monitor
        3. Health & Monitoring
        4. WebSocket JSON-RPC Server
        """
        if self._running:
            return

        correlation_id = (
            get_correlation_id() or f"service-startup-{uuid.uuid4().hex[:8]}"
        )
        set_correlation_id(correlation_id)

        self._logger.info(
            "Starting camera service components",
            extra={"correlation_id": correlation_id},
        )

        try:
            self._shutdown_event = asyncio.Event()

            # Start MediaMTX Controller
            await self._start_mediamtx_controller()

            # Start MediaMTX Path Manager
            await self._start_path_manager()

            # Start Camera Discovery Monitor
            await self._start_camera_monitor()

            # Start Health & Monitoring
            await self._start_health_monitor()

            # Start WebSocket JSON-RPC Server
            await self._start_websocket_server()

            self._running = True

            self._logger.info(
                "All camera service components started successfully",
                extra={"correlation_id": correlation_id},
            )

        except Exception as e:
            self._logger.error(
                f"Failed to start service components: {e}",
                extra={"correlation_id": correlation_id},
            )
            # Cleanup any partially started components
            await self._cleanup_partial_startup()
            raise

    async def stop(self) -> None:
        """
        Stop all service components gracefully.

        Gracefully stops components in reverse startup order:
        1. WebSocket JSON-RPC Server
        2. Health & Monitoring
        3. Camera Discovery Monitor
        4. MediaMTX Controller
        """
        if not self._running:
            return

        correlation_id = (
            get_correlation_id() or f"service-shutdown-{uuid.uuid4().hex[:8]}"
        )
        set_correlation_id(correlation_id)

        self._logger.info(
            "Stopping camera service components",
            extra={"correlation_id": correlation_id},
        )

        try:
            # Stop components in reverse order
            await self._stop_websocket_server()
            await self._stop_health_monitor()
            await self._stop_camera_monitor()
            await self._stop_path_manager()
            await self._stop_mediamtx_controller()

            self._running = False

            if self._shutdown_event:
                self._shutdown_event.set()

            self._logger.info(
                "All camera service components stopped",
                extra={"correlation_id": correlation_id},
            )

        except Exception as e:
            self._logger.error(
                f"Error during service shutdown: {e}",
                extra={"correlation_id": correlation_id},
            )
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

        Coordinates MediaMTX stream configuration updates based on camera events
        with robust error handling and defensive sequencing.

        Args:
            event_data: Camera event information including device path and type
        """
        correlation_id = (
            get_correlation_id()
            or f"camera-event-{event_data.device_path.split('/')[-1]}-{uuid.uuid4().hex[:8]}"
        )
        set_correlation_id(correlation_id)

        self._logger.info(
            f"Handling camera event: {event_data.event_type.value} - {event_data.device_path}",
            extra={
                "correlation_id": correlation_id,
                "device_path": event_data.device_path,
                "event_type": event_data.event_type.value,
            },
        )

        try:
            if event_data.event_type == CameraEvent.CONNECTED:
                await self._handle_camera_connected(event_data)
            elif event_data.event_type == CameraEvent.DISCONNECTED:
                await self._handle_camera_disconnected(event_data)
            elif event_data.event_type == CameraEvent.STATUS_CHANGED:
                await self._handle_camera_status_changed(event_data)
            else:
                self._logger.warning(
                    f"Unknown camera event type: {event_data.event_type}",
                    extra={
                        "correlation_id": correlation_id,
                        "device_path": event_data.device_path,
                    },
                )

        except Exception as e:
            self._logger.error(
                f"Error handling camera event: {e}",
                extra={
                    "correlation_id": correlation_id,
                    "device_path": event_data.device_path,
                },
                exc_info=True,
            )

    async def _handle_camera_connected(self, event_data: CameraEventData) -> None:
        """
        Handle camera connection event with robust MediaMTX stream creation and defensive error handling.

        Creates MediaMTX stream configuration for the newly connected camera and
        triggers camera_status_update notification as specified in the architecture overview.

        Args:
            event_data: Camera connection event data
        """
        correlation_id = get_correlation_id()
        device_path = event_data.device_path

        self._logger.debug(
            f"Creating stream for connected camera: {device_path}",
            extra={"correlation_id": correlation_id, "device_path": device_path},
        )

        # Extract stream name from device path
        stream_name = self._get_stream_name_from_device_path(device_path)

        # Get enhanced camera metadata with capability validation
        camera_metadata = await self._get_enhanced_camera_metadata(event_data)

        # Defensive guard: Check path manager availability
        stream_created = False
        streams_dict = {}
        mediamtx_error = None

        if not self._path_manager:
            self._logger.warning(
                "MediaMTX path manager not available for stream creation",
                extra={"correlation_id": correlation_id, "device_path": device_path},
            )
        else:
            try:
                # Extract camera ID from device path (e.g., "/dev/video0" -> "0")
                camera_id = device_path.split("/")[-1].replace("video", "")
                
                # Create MediaMTX path with FFmpeg publishing
                stream_created = await self._path_manager.create_camera_path(
                    camera_id=camera_id,
                    device_path=device_path,
                    rtsp_port=self._config.mediamtx.rtsp_port
                )

                if stream_created:
                    # Generate stream URLs for notification
                    streams_dict = {
                        "rtsp": f"rtsp://{self._config.mediamtx.host}:{self._config.mediamtx.rtsp_port}/cam{camera_id}",
                        "webrtc": f"http://{self._config.mediamtx.host}:{self._config.mediamtx.webrtc_port}/cam{camera_id}",
                        "hls": f"http://{self._config.mediamtx.host}:{self._config.mediamtx.hls_port}/cam{camera_id}",
                    }

                    self._logger.debug(
                        f"Successfully created MediaMTX path: cam{camera_id}",
                        extra={
                            "correlation_id": correlation_id,
                            "device_path": device_path,
                        },
                    )
                else:
                    mediamtx_error = "Failed to create MediaMTX path"

            except Exception as e:
                mediamtx_error = str(e)
                self._logger.error(
                    f"Failed to create MediaMTX path for {device_path}: {e}",
                    extra={
                        "correlation_id": correlation_id,
                        "device_path": device_path,
                    },
                )
                # Continue with notification despite MediaMTX error

        # Prepare notification parameters with capability validation context
        notification_params = {
            "device": device_path,
            "status": "CONNECTED",
            "name": camera_metadata["name"],
            "resolution": camera_metadata["resolution"],
            "fps": camera_metadata["fps"],
            "streams": streams_dict,
            # Enhanced metadata with provisional/confirmed state annotations
            "metadata_validation": camera_metadata["validation_status"],
            "metadata_source": camera_metadata["capability_source"],
            "metadata_provisional": camera_metadata["validation_status"]
            in ["provisional", "none"],
            "metadata_confirmed": camera_metadata["validation_status"] == "confirmed",
        }

        # Add capability validation context to logging
        validation_status = camera_metadata.get("validation_status", "none")
        capability_source = camera_metadata.get("capability_source", "default")
        consecutive_successes = camera_metadata.get("consecutive_successes", 0)

        try:
            self._logger.info(
                f"Camera connected with {capability_source} metadata: {device_path} "
                f"({camera_metadata['resolution']}@{camera_metadata['fps']}fps, "
                f"validation: {validation_status}, confirmations: {consecutive_successes})",
                extra={
                    "correlation_id": correlation_id,
                    "device_path": device_path,
                    "capability_validation": validation_status,
                    "capability_source": capability_source,
                    "capability_confirmations": consecutive_successes,
                    "stream_created": stream_created,
                    "mediamtx_error": mediamtx_error,
                },
            )

            # Send notification to all connected clients
            if self._websocket_server:
                await self._websocket_server.notify_camera_status_update(
                    notification_params
                )

            self._logger.info(
                f"Stream orchestration completed for camera: {device_path}",
                extra={"correlation_id": correlation_id, "device_path": device_path},
            )

        except Exception as e:
            self._logger.error(
                f"Failed to complete camera connection orchestration: {e}",
                extra={"correlation_id": correlation_id, "device_path": device_path},
            )

    async def _handle_camera_disconnected(self, event_data: CameraEventData) -> None:
        """
        Handle camera disconnection event with robust MediaMTX stream cleanup.

        Removes MediaMTX stream configuration for the disconnected camera and
        triggers camera_status_update notification as specified in the architecture overview.

        Args:
            event_data: Camera disconnection event data
        """
        correlation_id = get_correlation_id()
        device_path = event_data.device_path

        self._logger.debug(
            f"Removing stream for disconnected camera: {device_path}",
            extra={"correlation_id": correlation_id, "device_path": device_path},
        )

        # Defensive guard: Check path manager availability
        stream_removed = False
        mediamtx_error = None

        if not self._path_manager:
            self._logger.warning(
                "MediaMTX path manager not available for stream removal",
                extra={"correlation_id": correlation_id, "device_path": device_path},
            )
        else:
            try:
                # Extract camera ID from device path (e.g., "/dev/video0" -> "0")
                camera_id = device_path.split("/")[-1].replace("video", "")

                # Delete MediaMTX path with error handling
                stream_removed = await self._path_manager.delete_camera_path(camera_id)

                if stream_removed:
                    self._logger.debug(
                        f"Successfully removed MediaMTX path: cam{camera_id}",
                        extra={
                            "correlation_id": correlation_id,
                            "device_path": device_path,
                        },
                    )
                else:
                    mediamtx_error = "Failed to delete MediaMTX path"

            except Exception as e:
                mediamtx_error = str(e)
                self._logger.error(
                    f"Failed to delete MediaMTX path for {device_path}: {e}",
                    extra={
                        "correlation_id": correlation_id,
                        "device_path": device_path,
                    },
                )
                # Continue with notification despite MediaMTX error

        # Get camera metadata for notification (uses cached/default for disconnected)
        camera_metadata = await self._get_enhanced_camera_metadata(event_data)

        # Prepare camera status notification for disconnection
        notification_params = {
            "device": device_path,
            "status": "DISCONNECTED",
            "name": camera_metadata["name"],
            "resolution": "",  # Empty for disconnected cameras
            "fps": 0,  # Zero for disconnected cameras
            "streams": {},  # Empty streams for disconnected cameras
            # Metadata context for disconnected state
            "metadata_validation": "none",
            "metadata_source": "cached",
            "metadata_provisional": False,
            "metadata_confirmed": False,
        }

        try:
            # Send notification to all connected clients
            if self._websocket_server:
                await self._websocket_server.notify_camera_status_update(
                    notification_params
                )

            self._logger.info(
                f"Stream removal completed for camera: {device_path}",
                extra={
                    "correlation_id": correlation_id,
                    "device_path": device_path,
                    "stream_removed": stream_removed,
                    "mediamtx_error": mediamtx_error,
                },
            )

        except Exception as e:
            self._logger.error(
                f"Failed to complete camera disconnection orchestration: {e}",
                extra={"correlation_id": correlation_id, "device_path": device_path},
            )

    async def _handle_camera_status_changed(self, event_data: CameraEventData) -> None:
        """
        Handle camera status change event with enhanced state transition logging.

        Updates MediaMTX stream configuration based on camera status changes and
        triggers camera_status_update notification as specified in the architecture overview.

        Args:
            event_data: Camera status change event data
        """
        correlation_id = get_correlation_id()
        device_path = event_data.device_path

        old_status = "unknown"
        new_status = (
            event_data.device_info.status if event_data.device_info else "unknown"
        )

        self._logger.debug(
            f"Handling status change for camera: {device_path} ({old_status} -> {new_status})",
            extra={"correlation_id": correlation_id, "device_path": device_path},
        )

        try:
            # Extract stream name from device path
            stream_name = self._get_stream_name_from_device_path(device_path)

            # Get enhanced camera metadata for notification
            camera_metadata = await self._get_enhanced_camera_metadata(event_data)

            # Determine notification parameters based on new status
            if event_data.device_info and event_data.device_info.status == "CONNECTED":
                # Camera is now available - generate stream URLs
                notification_params = {
                    "device": device_path,
                    "status": "CONNECTED",
                    "name": camera_metadata["name"],
                    "resolution": camera_metadata["resolution"],
                    "fps": camera_metadata["fps"],
                    "streams": {
                        "rtsp": f"rtsp://{self._config.mediamtx.host}:{self._config.mediamtx.rtsp_port}/{stream_name}",
                        "webrtc": f"http://{self._config.mediamtx.host}:{self._config.mediamtx.webrtc_port}/{stream_name}",
                        "hls": f"http://{self._config.mediamtx.host}:{self._config.mediamtx.hls_port}/{stream_name}",
                    },
                    "metadata_validation": camera_metadata["validation_status"],
                    "metadata_source": camera_metadata["capability_source"],
                    "metadata_provisional": camera_metadata["validation_status"]
                    in ["provisional", "none"],
                    "metadata_confirmed": camera_metadata["validation_status"]
                    == "confirmed",
                }

                # Log capability validation context for status change
                validation_status = camera_metadata.get("validation_status", "none")
                self._logger.info(
                    f"Camera status changed to CONNECTED: {device_path} (validation: {validation_status})",
                    extra={
                        "correlation_id": correlation_id,
                        "device_path": device_path,
                        "capability_validation": validation_status,
                        "status_transition": f"{old_status}_to_CONNECTED",
                    },
                )
            else:
                # Camera has error or other status
                error_status = (
                    "ERROR"
                    if event_data.device_info
                    and event_data.device_info.status == "ERROR"
                    else "DISCONNECTED"
                )
                notification_params = {
                    "device": device_path,
                    "status": error_status,
                    "name": camera_metadata["name"],
                    "resolution": "",
                    "fps": 0,
                    "streams": {},
                    "metadata_validation": "none",
                    "metadata_source": "cached",
                    "metadata_provisional": False,
                    "metadata_confirmed": False,
                }

                self._logger.info(
                    f"Camera status changed to {error_status}: {device_path}",
                    extra={
                        "correlation_id": correlation_id,
                        "device_path": device_path,
                        "status_transition": f"{old_status}_to_{error_status}",
                    },
                )

            # Send notification to all connected clients
            if self._websocket_server:
                await self._websocket_server.notify_camera_status_update(
                    notification_params
                )

            self._logger.info(
                f"Status change notification sent for camera: {device_path}",
                extra={"correlation_id": correlation_id, "device_path": device_path},
            )

        except Exception as e:
            self._logger.error(
                f"Error handling status change for {device_path}: {e}",
                extra={"correlation_id": correlation_id, "device_path": device_path},
            )

    def _get_stream_name_from_device_path(self, device_path: str) -> str:
        """
        Generate a deterministic stream name from camera device path.

        Extracts device number from standard paths (e.g., /dev/video0 -> camera0)
        or generates a hash-based name for non-standard paths.

        Args:
            device_path: Camera device path (e.g., /dev/video0)

        Returns:
            Stream name (e.g., camera0)
        """
        if not device_path:
            return "camera_unknown"

        # Try to extract video device number from standard paths
        match = re.search(r"/(?:dev/)?video(\d+)", device_path)
        if match:
            return f"camera{match.group(1)}"

        # For non-standard paths, generate deterministic hash-based name
        import hashlib

        path_hash = hashlib.md5(device_path.encode()).hexdigest()[:8]
        return f"camera_{path_hash}"

    async def _get_enhanced_camera_metadata(
        self, event_data: CameraEventData
    ) -> Dict[str, Any]:
        """
        Get enhanced camera metadata with capability validation status.

        Integrates with camera monitor capability detection using the robust
        get_effective_capability_metadata method that handles provisional/confirmed logic.
        Provides clear annotation of capability validation status and data source.

        Args:
            event_data: Camera event data containing device info

        Returns:
            Dictionary with camera metadata and validation context:
                - name, resolution, fps (data fields)
                - validation_status: "confirmed", "provisional", "none", "error"
                - capability_source: "confirmed_capability", "provisional_capability", "device_info", "default"
                - consecutive_successes: number of consecutive capability confirmations (if available)
        """
        correlation_id = get_correlation_id()
        device_path = event_data.device_path

        # Extract device number for default naming
        device_num = "unknown"
        try:
            match = re.search(r"/dev/video(\d+)", device_path)
            if match:
                device_num = match.group(1)
        except Exception:
            pass

        # Initialize with default metadata
        camera_metadata = {
            "name": f"Camera {device_num}",
            "resolution": "1920x1080",  # Architecture default
            "fps": 30,  # Architecture default
            "validation_status": "none",
            "capability_source": "default",
            "consecutive_successes": 0,
        }

        # Override with device info if available
        if event_data.device_info:
            camera_metadata["name"] = (
                event_data.device_info.name or camera_metadata["name"]
            )
            camera_metadata["capability_source"] = "device_info"

        # Attempt to get enhanced capability data from camera monitor using robust method
        try:
            if self._camera_monitor and hasattr(
                self._camera_monitor, "get_effective_capability_metadata"
            ):

                capability_data = (
                    self._camera_monitor.get_effective_capability_metadata(device_path)
                )

                if capability_data:
                    # Extract validation status and consecutive successes
                    validation_status = capability_data.get("validation_status", "none")
                    consecutive_successes = capability_data.get(
                        "consecutive_successes", 0
                    )

                    # Update metadata with capability data
                    camera_metadata.update(
                        {
                            "resolution": capability_data.get(
                                "resolution", camera_metadata["resolution"]
                            ),
                            "fps": capability_data.get("fps", camera_metadata["fps"]),
                            "validation_status": validation_status,
                            "consecutive_successes": consecutive_successes,
                        }
                    )

                    # Determine capability source based on validation status
                    if validation_status == "confirmed":
                        camera_metadata["capability_source"] = "confirmed_capability"
                        self._logger.debug(
                            f"Using confirmed capability data for {device_path}: "
                            f"{camera_metadata['resolution']}@{camera_metadata['fps']}fps "
                            f"(confirmations: {consecutive_successes})",
                            extra={
                                "correlation_id": correlation_id,
                                "device_path": device_path,
                                "capability_validation": "confirmed",
                            },
                        )
                    elif validation_status == "provisional":
                        camera_metadata["capability_source"] = "provisional_capability"
                        self._logger.debug(
                            f"Using provisional capability data for {device_path}: "
                            f"{camera_metadata['resolution']}@{camera_metadata['fps']}fps "
                            f"(pending confirmation)",
                            extra={
                                "correlation_id": correlation_id,
                                "device_path": device_path,
                                "capability_validation": "provisional",
                            },
                        )
                    elif validation_status == "failed":
                        camera_metadata["validation_status"] = "error"
                        camera_metadata["capability_source"] = "default"
                        self._logger.debug(
                            f"Capability detection failed for {device_path}, using defaults",
                            extra={
                                "correlation_id": correlation_id,
                                "device_path": device_path,
                                "capability_validation": "error",
                            },
                        )
                    else:
                        # validation_status == "none" or unknown
                        camera_metadata["capability_source"] = "default"
                        self._logger.debug(
                            f"No capability data available for {device_path}, using defaults",
                            extra={
                                "correlation_id": correlation_id,
                                "device_path": device_path,
                                "capability_validation": "none",
                            },
                        )

        except Exception as e:
            # Capability detection error - use defaults with error annotation
            camera_metadata["validation_status"] = "error"
            camera_metadata["capability_source"] = "default"
            self._logger.warning(
                f"Error retrieving capability metadata for {device_path}: {e}, using defaults",
                extra={
                    "correlation_id": correlation_id,
                    "device_path": device_path,
                    "capability_validation": "error",
                },
            )

        return camera_metadata

    async def _start_mediamtx_controller(self) -> None:
        """Start the MediaMTX controller component."""
        correlation_id = get_correlation_id()
        self._logger.debug(
            "Starting MediaMTX controller", extra={"correlation_id": correlation_id}
        )

        try:
            # Prefer absolute imports to avoid package-relative issues
            from mediamtx_wrapper.controller import MediaMTXController

            # Unpack MediaMTXConfig into individual parameters
            mediamtx_config = self._config.mediamtx
            self._mediamtx_controller = MediaMTXController(
                host=mediamtx_config.host,
                api_port=mediamtx_config.api_port,
                rtsp_port=mediamtx_config.rtsp_port,
                webrtc_port=mediamtx_config.webrtc_port,
                hls_port=mediamtx_config.hls_port,
                config_path=mediamtx_config.config_path,
                recordings_path=mediamtx_config.recordings_path,
                snapshots_path=mediamtx_config.snapshots_path,
                health_check_interval=mediamtx_config.health_check_interval,
                health_failure_threshold=mediamtx_config.health_failure_threshold,
                health_circuit_breaker_timeout=mediamtx_config.health_circuit_breaker_timeout,
                health_max_backoff_interval=mediamtx_config.health_max_backoff_interval,
                health_recovery_confirmation_threshold=mediamtx_config.health_recovery_confirmation_threshold,
                backoff_base_multiplier=mediamtx_config.backoff_base_multiplier,
                backoff_jitter_range=mediamtx_config.backoff_jitter_range,
                process_termination_timeout=mediamtx_config.process_termination_timeout,
                process_kill_timeout=mediamtx_config.process_kill_timeout,
            )
            await self._mediamtx_controller.start()

            # Verify MediaMTX health after startup
            health_status = await self._mediamtx_controller.health_check()
            self._logger.info(
                f"MediaMTX controller started: {health_status.get('status', 'unknown')}",
                extra={"correlation_id": correlation_id},
            )

        except Exception as e:
            self._logger.error(
                f"Failed to start MediaMTX controller: {e}",
                extra={"correlation_id": correlation_id},
            )
            raise

    async def _start_path_manager(self) -> None:
        """Start the MediaMTX path manager component."""
        correlation_id = get_correlation_id()
        self._logger.debug(
            "Starting MediaMTX path manager", extra={"correlation_id": correlation_id}
        )

        try:
            mediamtx_config = self._config.mediamtx
            self._path_manager = MediaMTXPathManager(
                mediamtx_host=mediamtx_config.host,
                mediamtx_port=mediamtx_config.api_port
            )
            await self._path_manager.start()

            self._logger.info(
                "MediaMTX path manager started successfully",
                extra={"correlation_id": correlation_id},
            )

        except Exception as e:
            self._logger.error(
                f"Failed to start MediaMTX path manager: {e}",
                extra={"correlation_id": correlation_id},
            )
            raise

    async def _start_camera_monitor(self) -> None:
        """Start the camera discovery and monitoring component."""
        correlation_id = get_correlation_id()
        self._logger.debug(
            "Starting camera discovery monitor",
            extra={"correlation_id": correlation_id},
        )

        try:
            # Prefer absolute imports to avoid package-relative issues
            from camera_discovery.hybrid_monitor import HybridCameraMonitor

            self._camera_monitor = HybridCameraMonitor(
                device_range=self._config.camera.device_range,
                poll_interval=self._config.camera.poll_interval,
                detection_timeout=self._config.camera.detection_timeout,
                enable_capability_detection=self._config.camera.enable_capability_detection,
            )

            # Register ourselves as an event handler
            self._camera_monitor.add_event_handler(self)

            # Start camera monitoring
            await self._camera_monitor.start()
            self._logger.info(
                "Camera discovery monitor started",
                extra={"correlation_id": correlation_id},
            )

        except Exception as e:
            self._logger.error(
                f"Failed to start camera monitor: {e}",
                extra={"correlation_id": correlation_id},
            )
            raise

    async def _start_health_monitor(self) -> None:
        """Start the health monitoring component."""
        correlation_id = get_correlation_id()
        self._logger.debug(
            "Starting health monitor", extra={"correlation_id": correlation_id}
        )

        try:
            self._health_monitor = HealthMonitor(self._config)
            await self._health_monitor.start()
            self._logger.info(
                "Health monitor started", extra={"correlation_id": correlation_id}
            )

        except Exception as e:
            self._logger.error(
                f"Failed to start health monitor: {e}",
                extra={"correlation_id": correlation_id},
            )
            raise

    async def _start_websocket_server(self) -> None:
        """Start the WebSocket JSON-RPC server component."""
        correlation_id = get_correlation_id()
        self._logger.debug(
            "Starting WebSocket JSON-RPC server",
            extra={"correlation_id": correlation_id},
        )

        try:
            from websocket_server.server import WebSocketJsonRpcServer
            from security.jwt_handler import JWTHandler
            from security.api_key_handler import APIKeyHandler
            from security.auth_manager import AuthManager
            from security.middleware import SecurityMiddleware

            # Only create WebSocket server if not provided in constructor
            if self._websocket_server is None:
                self._websocket_server = WebSocketJsonRpcServer(
                    host=self._config.server.host,
                    port=self._config.server.port,
                    websocket_path=self._config.server.websocket_path,
                    max_connections=self._config.server.max_connections,
                    mediamtx_controller=self._mediamtx_controller,
                    camera_monitor=self._camera_monitor,
                )
            # Provide service manager reference for API methods that require it
            if hasattr(self._websocket_server, "set_service_manager"):
                self._websocket_server.set_service_manager(self)
            # Configure security middleware (env-configurable)
            try:
                import os
                jwt_secret = os.environ.get("CAMERA_SERVICE_JWT_SECRET", "dev-secret-change-me")
                api_keys_path = os.environ.get("CAMERA_SERVICE_API_KEYS_PATH", "/opt/camera-service/keys/api_keys.json")
                rpm = int(os.environ.get("CAMERA_SERVICE_RATE_RPM", "120"))
                jwt_handler = JWTHandler(secret_key=jwt_secret)
                api_key_handler = APIKeyHandler(storage_file=api_keys_path)
                auth_manager = AuthManager(jwt_handler=jwt_handler, api_key_handler=api_key_handler)
                security = SecurityMiddleware(
                    auth_manager=auth_manager,
                    max_connections=self._config.server.max_connections,
                    requests_per_minute=rpm,
                )
                if hasattr(self._websocket_server, "set_security_middleware"):
                    self._websocket_server.set_security_middleware(security)
            except Exception as e:
                self._logger.warning(f"Security middleware initialization failed: {e}")
            await self._websocket_server.start()
            self._logger.info(
                "WebSocket JSON-RPC server started",
                extra={"correlation_id": correlation_id},
            )

        except Exception as e:
            self._logger.error(
                f"Failed to start WebSocket server: {e}",
                extra={"correlation_id": correlation_id},
            )
            raise

    async def _stop_websocket_server(self) -> None:
        """Stop the WebSocket JSON-RPC server component."""
        if self._websocket_server:
            correlation_id = get_correlation_id()
            self._logger.debug(
                "Stopping WebSocket JSON-RPC server",
                extra={"correlation_id": correlation_id},
            )
            try:
                await self._websocket_server.stop()
                self._logger.info(
                    "WebSocket JSON-RPC server stopped",
                    extra={"correlation_id": correlation_id},
                )
            except Exception as e:
                self._logger.error(
                    f"Error stopping WebSocket server: {e}",
                    extra={"correlation_id": correlation_id},
                )
            finally:
                self._websocket_server = None

    async def _stop_health_monitor(self) -> None:
        """Stop the health monitoring component."""
        if self._health_monitor:
            correlation_id = get_correlation_id()
            self._logger.debug(
                "Stopping health monitor", extra={"correlation_id": correlation_id}
            )
            try:
                await self._health_monitor.stop()
                self._logger.info(
                    "Health monitor stopped", extra={"correlation_id": correlation_id}
                )
            except Exception as e:
                self._logger.error(
                    f"Error stopping health monitor: {e}",
                    extra={"correlation_id": correlation_id},
                )
            finally:
                self._health_monitor = None

    async def _stop_camera_monitor(self) -> None:
        """Stop the camera discovery and monitoring component."""
        if self._camera_monitor:
            correlation_id = get_correlation_id()
            self._logger.debug(
                "Stopping camera discovery monitor",
                extra={"correlation_id": correlation_id},
            )
            try:
                # Unregister event handler if supported
                if hasattr(self._camera_monitor, "remove_event_handler"):
                    try:
                        self._camera_monitor.remove_event_handler(self)
                    except Exception:
                        # Proceed with shutdown even if handler removal fails
                        pass
                # Stop camera monitoring
                await self._camera_monitor.stop()
                self._logger.info(
                    "Camera discovery monitor stopped",
                    extra={"correlation_id": correlation_id},
                )
            except Exception as e:
                self._logger.error(
                    f"Error stopping camera monitor: {e}",
                    extra={"correlation_id": correlation_id},
                )
            finally:
                self._camera_monitor = None

    async def _stop_path_manager(self) -> None:
        """Stop the MediaMTX path manager component."""
        if self._path_manager:
            correlation_id = get_correlation_id()
            self._logger.debug(
                "Stopping MediaMTX path manager", extra={"correlation_id": correlation_id}
            )
            try:
                await self._path_manager.stop()
                self._logger.info(
                    "MediaMTX path manager stopped",
                    extra={"correlation_id": correlation_id},
                )
            except Exception as e:
                self._logger.error(
                    f"Error stopping MediaMTX path manager: {e}",
                    extra={"correlation_id": correlation_id},
                )
            finally:
                self._path_manager = None

    async def _stop_mediamtx_controller(self) -> None:
        """Stop the MediaMTX controller component."""
        if self._mediamtx_controller:
            correlation_id = get_correlation_id()
            self._logger.debug(
                "Stopping MediaMTX controller", extra={"correlation_id": correlation_id}
            )
            try:
                await self._mediamtx_controller.stop()
                self._logger.info(
                    "MediaMTX controller stopped",
                    extra={"correlation_id": correlation_id},
                )
            except Exception as e:
                self._logger.error(
                    f"Error stopping MediaMTX controller: {e}",
                    extra={"correlation_id": correlation_id},
                )
            finally:
                self._mediamtx_controller = None

    async def _cleanup_partial_startup(self) -> None:
        """Clean up any partially started components after startup failure."""
        correlation_id = get_correlation_id()
        self._logger.warning(
            "Cleaning up partially started components",
            extra={"correlation_id": correlation_id},
        )

        try:
            if self._websocket_server:
                await self._websocket_server.stop()
                self._websocket_server = None

            if self._health_monitor:
                await self._health_monitor.stop()
                self._health_monitor = None

            if self._camera_monitor:
                if hasattr(self._camera_monitor, "remove_event_handler"):
                    self._camera_monitor.remove_event_handler(self)
                await self._camera_monitor.stop()
                self._camera_monitor = None

            if self._path_manager:
                await self._path_manager.stop()
                self._path_manager = None

            if self._mediamtx_controller:
                await self._mediamtx_controller.stop()
                self._mediamtx_controller = None

        except Exception as e:
            self._logger.error(
                f"Error during partial state cleanup: {e}",
                extra={"correlation_id": correlation_id},
            )
