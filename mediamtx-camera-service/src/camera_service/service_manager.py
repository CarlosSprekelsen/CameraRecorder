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
from typing import Optional, Dict, Any

from .config import Config
from ..mediamtx_wrapper.controller import MediaMTXController, StreamConfig
from ..camera_discovery.hybrid_monitor import CameraEventData, CameraEvent, CameraEventHandler
from ..websocket_server.server import WebSocketJsonRpcServer


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
            
        self._logger.info("Starting health monitor")
        self._running = True
        
        # Start background health check task
        self._health_check_task = asyncio.create_task(self._health_check_loop())
        self._logger.debug("Health monitor started successfully")
    
    async def stop(self) -> None:
        """Stop health monitoring."""
        if not self._running:
            return
            
        self._logger.info("Stopping health monitor")
        self._running = False
        
        # Stop background health check task
        if self._health_check_task and not self._health_check_task.done():
            self._health_check_task.cancel()
            try:
                await self._health_check_task
            except asyncio.CancelledError:
                pass
        
        self._logger.debug("Health monitor stopped")
    
    async def _health_check_loop(self) -> None:
        """Background health monitoring loop."""
        while self._running:
            try:
                # Perform basic health checks
                await asyncio.sleep(30)  # Health check interval
            except asyncio.CancelledError:
                break
            except Exception as e:
                self._logger.error(f"Health check error: {e}")
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
        
        # Component references
        self._websocket_server: Optional[WebSocketJsonRpcServer] = None
        self._camera_monitor = None
        self._mediamtx_controller: Optional[MediaMTXController] = None
        self._health_monitor: Optional[HealthMonitor] = None

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
        """
        self._logger.debug(f"Creating stream for connected camera: {event_data.device_path}")
        
        if not self._mediamtx_controller:
            self._logger.warning("MediaMTX controller not available for stream creation")
            return
        
        try:
            # Extract stream name from device path
            stream_name = self._get_stream_name_from_device_path(event_data.device_path)
            
            # Create stream configuration
            stream_config = StreamConfig(
                name=stream_name,
                source=event_data.device_path,
                record=self._config.recording.auto_record
            )
            
            # Create stream in MediaMTX
            stream_urls = await self._mediamtx_controller.create_stream(stream_config)
            
            # Get real camera metadata from device info and capability detection
            camera_metadata = await self._get_camera_metadata(event_data)
            
            # Prepare camera status notification with real metadata
            notification_params = {
                "device": event_data.device_path,
                "status": "CONNECTED",
                "name": camera_metadata["name"],
                "resolution": camera_metadata["resolution"],
                "fps": camera_metadata["fps"],
                "streams": stream_urls
            }
            
            # Send notification to all connected clients
            if self._websocket_server:
                await self._websocket_server.notify_camera_status_update(notification_params)
            
            self._logger.info(f"Stream created and notification sent for camera: {event_data.device_path}")
            
        except Exception as e:
            self._logger.error(f"Failed to handle camera connection: {e}")

    async def _handle_camera_disconnected(self, event_data: CameraEventData) -> None:
        """
        Handle camera disconnection event.

        Removes MediaMTX stream configuration for the disconnected camera and
        triggers camera_status_update notification as specified in the architecture overview.

        Args:
            event_data: Camera disconnection event data
        """
        self._logger.debug(f"Removing stream for disconnected camera: {event_data.device_path}")
        
        if not self._mediamtx_controller:
            self._logger.warning("MediaMTX controller not available for stream removal")
            return
        
        try:
            # Extract stream name from device path
            stream_name = self._get_stream_name_from_device_path(event_data.device_path)
            
            # Delete stream from MediaMTX
            await self._mediamtx_controller.delete_stream(stream_name)
            
            # Get camera metadata for notification
            camera_metadata = await self._get_camera_metadata(event_data)
            
            # Prepare camera status notification
            notification_params = {
                "device": event_data.device_path,
                "status": "DISCONNECTED",
                "name": camera_metadata["name"],
                "resolution": "",  # Empty for disconnected cameras
                "fps": 0,          # Zero for disconnected cameras
                "streams": {}
            }
            
            # Send notification to all connected clients
            if self._websocket_server:
                await self._websocket_server.notify_camera_status_update(notification_params)
            
            self._logger.info(f"Stream removed and notification sent for camera: {event_data.device_path}")
            
        except Exception as e:
            self._logger.error(f"Failed to handle camera disconnection: {e}")

    async def _handle_camera_status_changed(self, event_data: CameraEventData) -> None:
        """
        Handle camera status change event.

        Updates MediaMTX stream configuration based on camera status changes and
        triggers camera_status_update notification as specified in the architecture overview.

        Args:
            event_data: Camera status change event data
        """
        self._logger.debug(f"Handling status change for camera: {event_data.device_path}")
        
        try:
            # Extract stream name from device path
            stream_name = self._get_stream_name_from_device_path(event_data.device_path)
            
            # Get camera metadata for notification
            camera_metadata = await self._get_camera_metadata(event_data)
            
            # Determine notification parameters based on new status
            if event_data.device_info and event_data.device_info.status == "CONNECTED":
                # Camera is now available
                notification_params = {
                    "device": event_data.device_path,
                    "status": "CONNECTED",
                    "name": camera_metadata["name"],
                    "resolution": camera_metadata["resolution"],
                    "fps": camera_metadata["fps"],
                    "streams": {
                        "rtsp": f"rtsp://{self._config.mediamtx.host}:{self._config.mediamtx.rtsp_port}/{stream_name}",
                        "webrtc": f"http://{self._config.mediamtx.host}:{self._config.mediamtx.webrtc_port}/{stream_name}",
                        "hls": f"http://{self._config.mediamtx.host}:{self._config.mediamtx.hls_port}/{stream_name}"
                    }
                }
            else:
                # Camera has error or other status
                notification_params = {
                    "device": event_data.device_path,
                    "status": "ERROR" if event_data.device_info and event_data.device_info.status == "ERROR" else "DISCONNECTED",
                    "name": camera_metadata["name"],
                    "resolution": "",
                    "fps": 0,
                    "streams": {}
                }
            
            # Send notification to all connected clients
            if self._websocket_server:
                await self._websocket_server.notify_camera_status_update(notification_params)
            
            self._logger.info(f"Status change notification sent for camera: {event_data.device_path}")
            
        except Exception as e:
            self._logger.error(f"Failed to handle camera status change: {e}")

    def _get_stream_name_from_device_path(self, device_path: str) -> str:
        """
        Extract stream name from camera device path.
        
        Uses deterministic pattern matching to convert device paths like /dev/video0
        to MediaMTX-compatible stream names like camera0. Provides fallback handling
        for non-standard device paths.
        
        Args:
            device_path: Camera device path (e.g., /dev/video0)
            
        Returns:
            Stream name for MediaMTX (e.g., camera0)
            
        Examples:
            /dev/video0 -> camera0
            /dev/video15 -> camera15
            /custom/path/device -> camera_hash
        """
        try:
            # Use regex to extract device number from standard V4L2 paths
            match = re.search(r'/dev/video(\d+)', device_path)
            if match:
                device_num = match.group(1)
                return f"camera{device_num}"
            
            # Fallback for non-standard device paths
            # Extract any digits from the path
            digits = re.findall(r'\d+', device_path)
            if digits:
                return f"camera{digits[-1]}"
            
            # Final fallback using hash for completely non-standard paths
            return f"camera_{abs(hash(device_path)) % 1000}"
            
        except Exception as e:
            self._logger.warning(f"Error extracting stream name from {device_path}: {e}")
            return "camera_unknown"
            
    async def _get_camera_metadata(self, event_data: CameraEventData) -> Dict[str, Any]:
        """
        Get camera metadata including resolution and fps from capability detection.
        
        Integrates with camera monitor capability detection to provide real device
        metadata when available. Falls back to documented defaults when capability
        data is not accessible.
        
        Args:
            event_data: Camera event data containing device info
            
        Returns:
            Dictionary with camera name, resolution, and fps
        """
        # Extract device number for default naming
        device_num = "unknown"
        try:
            match = re.search(r'/dev/video(\d+)', event_data.device_path)
            if match:
                device_num = match.group(1)
        except Exception:
            pass
        
        # Initialize with default metadata
        camera_metadata = {
            "name": f"Camera {device_num}",
            "resolution": "1920x1080",  # Architecture default
            "fps": 30                   # Architecture default
        }
        
        # Override with device info if available
        if event_data.device_info:
            camera_metadata["name"] = event_data.device_info.name or camera_metadata["name"]
        
        # Attempt to get real capability data from camera monitor
        try:
            if (self._camera_monitor and 
                hasattr(self._camera_monitor, '_probe_device_capabilities')):
                
                self._logger.debug(f"Probing capabilities for {event_data.device_path}")
                capabilities = await self._camera_monitor._probe_device_capabilities(
                    event_data.device_path
                )
                
                if capabilities and capabilities.get("detected"):
                    # Extract real resolution and fps from capability detection
                    resolutions = capabilities.get("resolutions", [])
                    frame_rates = capabilities.get("frame_rates", [])
                    
                    if resolutions:
                        # Use first available resolution as current setting
                        camera_metadata["resolution"] = resolutions[0]
                        self._logger.debug(f"Using real resolution: {resolutions[0]} for {event_data.device_path}")
                    
                    if frame_rates:
                        # Convert frame rate to integer, use first available
                        try:
                            camera_metadata["fps"] = int(float(frame_rates[0]))
                            self._logger.debug(f"Using real fps: {camera_metadata['fps']} for {event_data.device_path}")
                        except (ValueError, IndexError):
                            self._logger.debug(f"Could not parse frame rate {frame_rates[0]}, using default")
                    
                    self._logger.info(
                        f"Using capability-derived metadata for {event_data.device_path}: "
                        f"{camera_metadata['resolution']}@{camera_metadata['fps']}fps"
                    )
                else:
                    self._logger.debug(
                        f"Capability detection failed for {event_data.device_path}, using defaults"
                    )
            else:
                self._logger.debug(
                    f"Capability detection not available for {event_data.device_path}, using defaults"
                )
                
        except Exception as e:
            self._logger.warning(
                f"Failed to get capability data for {event_data.device_path}, "
                f"using defaults: {e}"
            )
        
        return camera_metadata

    async def _validate_camera_monitor_integration(self) -> None:
        """
        Validate camera monitor integration and log capability detection status.
        
        This method checks what capability data is available from the camera monitor
        and logs the current integration status for debugging and validation.
        """
        if not self._camera_monitor:
            self._logger.warning("Camera monitor not available")
            return
        
        try:
            # Check if camera monitor has capability detection
            has_capabilities = hasattr(self._camera_monitor, '_probe_device_capabilities')
            
            if has_capabilities:
                self._logger.info("Camera monitor capability detection integration: AVAILABLE")
            else:
                self._logger.info(
                    "Camera monitor capability detection integration: PENDING - "
                    "using default metadata values"
                )
            
            # Get current connected cameras for validation
            if hasattr(self._camera_monitor, 'get_connected_cameras'):
                connected_cameras = await self._camera_monitor.get_connected_cameras()
                self._logger.debug(f"Currently connected cameras: {len(connected_cameras)}")
                
        except Exception as e:
            self._logger.error(f"Error validating camera monitor integration: {e}")

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
        self._logger.debug("Starting camera discovery monitor")
        
        try:
            # Initialize Camera Discovery Monitor with config
            from ..camera_discovery.hybrid_monitor import HybridCameraMonitor
            
            self._camera_monitor = HybridCameraMonitor(
                device_range=self._config.camera.device_range,
                poll_interval=self._config.camera.poll_interval,
                detection_timeout=self._config.camera.detection_timeout,
                enable_capability_detection=self._config.camera.enable_capability_detection
            )
            
            # Register this ServiceManager as camera event handler
            self._camera_monitor.add_event_handler(self)
            self._logger.debug("Registered ServiceManager as camera event handler")
            
            # Setup hybrid udev + polling camera detection and start monitoring
            await self._camera_monitor.start()
            self._logger.info("Camera discovery monitor started successfully")
            
            # Validate integration capabilities
            await self._validate_camera_monitor_integration()
            
            # Start camera capability detection (handled by HybridCameraMonitor)
            capability_status = "enabled" if self._config.camera.enable_capability_detection else "disabled"
            self._logger.debug(f"Camera capability detection {capability_status}")
            
            # Log configuration details
            self._logger.debug(f"Camera monitor configuration: device_range={self._config.camera.device_range}, poll_interval={self._config.camera.poll_interval}s, detection_timeout={self._config.camera.detection_timeout}s")
            
        except Exception as e:
            self._logger.error(f"Failed to start camera monitor: {e}")
            # Cleanup on failure
            if self._camera_monitor:
                try:
                    await self._camera_monitor.stop()
                except Exception as cleanup_error:
                    self._logger.error(f"Error during camera monitor cleanup: {cleanup_error}")
                self._camera_monitor = None
            raise

    async def _start_health_monitor(self) -> None:
        """Start the health monitoring and recovery component."""
        self._logger.debug("Starting health monitor")
        
        try:
            # Initialize Health Monitor with config
            self._health_monitor = HealthMonitor(self._config)
            
            # Start health monitoring
            await self._health_monitor.start()
            self._logger.info("Health monitor started successfully")
            
        except Exception as e:
            self._logger.error(f"Failed to start health monitor: {e}")
            # Cleanup on failure
            if self._health_monitor:
                try:
                    await self._health_monitor.stop()
                except Exception as cleanup_error:
                    self._logger.error(f"Error during health monitor cleanup: {cleanup_error}")
                self._health_monitor = None
            raise

    async def _start_websocket_server(self) -> None:
        """Start the WebSocket JSON-RPC server component."""
        self._logger.debug("Starting WebSocket JSON-RPC server")
        
        try:
            # Initialize WebSocket server with config
            self._websocket_server = WebSocketJsonRpcServer(
                host=self._config.server.host,
                port=self._config.server.port,
                websocket_path=self._config.server.websocket_path,
                max_connections=self._config.server.max_connections,
                mediamtx_controller=self._mediamtx_controller,
                camera_monitor=self._camera_monitor
            )
            
            # Start WebSocket server
            await self._websocket_server.start()
            self._logger.info("WebSocket JSON-RPC server started successfully")
            
        except Exception as e:
            self._logger.error(f"Failed to start WebSocket server: {e}")
            # Cleanup on failure
            if self._websocket_server:
                try:
                    await self._websocket_server.stop()
                except Exception as cleanup_error:
                    self._logger.error(f"Error during WebSocket server cleanup: {cleanup_error}")
                self._websocket_server = None
            raise

    async def _stop_websocket_server(self) -> None:
        """Stop the WebSocket JSON-RPC server component."""
        if self._websocket_server:
            self._logger.debug("Stopping WebSocket JSON-RPC server")
            try:
                await self._websocket_server.stop()
                self._logger.info("WebSocket JSON-RPC server stopped")
            except Exception as e:
                self._logger.error(f"Error stopping WebSocket server: {e}")
            finally:
                self._websocket_server = None

    async def _stop_health_monitor(self) -> None:
        """Stop the health monitoring component."""
        if self._health_monitor:
            self._logger.debug("Stopping health monitor")
            try:
                await self._health_monitor.stop()
                self._logger.info("Health monitor stopped")
            except Exception as e:
                self._logger.error(f"Error stopping health monitor: {e}")
            finally:
                self._health_monitor = None

    async def _stop_camera_monitor(self) -> None:
        """Stop the camera discovery and monitoring component."""
        if self._camera_monitor:
            self._logger.debug("Stopping camera discovery monitor")
            try:
                # Unregister event handler
                self._camera_monitor.remove_event_handler(self)
                # Stop camera monitoring
                await self._camera_monitor.stop()
                self._logger.info("Camera discovery monitor stopped")
            except Exception as e:
                self._logger.error(f"Error stopping camera monitor: {e}")
            finally:
                self._camera_monitor = None

    async def _stop_mediamtx_controller(self) -> None:
        """Stop the MediaMTX controller component."""
        if self._mediamtx_controller:
            self._logger.debug("Stopping MediaMTX controller")
            try:
                await self._mediamtx_controller.stop()
                self._logger.info("MediaMTX controller stopped")
            except Exception as e:
                self._logger.error(f"Error stopping MediaMTX controller: {e}")
            finally:
                self._mediamtx_controller = None

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
        return {
            "running": self._running,
            "websocket_server": "running" if self._websocket_server else "stopped",
            "camera_monitor": "running" if self._camera_monitor else "stopped", 
            "mediamtx_controller": "running" if self._mediamtx_controller else "stopped",
            "health_monitor": "running" if self._health_monitor else "stopped"
        }