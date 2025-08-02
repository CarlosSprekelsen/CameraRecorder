"""
Hybrid camera discovery monitor implementation.

Provides real-time USB camera detection using udev events with polling
fallback for reliability, as specified in the architecture design.
"""

import asyncio
import logging
import re
from abc import ABC, abstractmethod
from dataclasses import dataclass
from enum import Enum
from pathlib import Path
from typing import Callable, Dict, List, Optional, Set

from ..common.types import CameraDevice

# Optional dependency for udev monitoring
try:
    import pyudev
    HAS_PYUDEV = True
except ImportError:
    HAS_PYUDEV = False


class CameraEvent(Enum):
    """Camera connection events."""
    CONNECTED = "CONNECTED"
    DISCONNECTED = "DISCONNECTED"
    STATUS_CHANGED = "STATUS_CHANGED"


@dataclass
class CameraEventData:
    """Data structure for camera events."""
    device_path: str
    event_type: CameraEvent
    device_info: Optional[CameraDevice] = None
    timestamp: Optional[float] = None


class CameraEventHandler(ABC):
    """Abstract interface for camera event handling."""
    
    @abstractmethod
    async def handle_camera_event(self, event_data: CameraEventData) -> None:
        """
        Handle camera connection/disconnection events.
        
        Args:
            event_data: Event information including device path and type
        """
        pass


class HybridCameraMonitor:
    """
    Hybrid camera discovery monitor using udev events and polling fallback.
    
    Implements the Camera Discovery Monitor component from the architecture,
    providing real-time USB camera detection with reliability through dual
    monitoring approaches.
    """
    
    def __init__(
        self,
        device_range: List[int] = None,
        poll_interval: float = 0.1,
        detection_timeout: float = 2.0,
        enable_capability_detection: bool = True
    ):
        """
        Initialize the hybrid camera monitor.
        
        Args:
            device_range: List of video device numbers to monitor (e.g., [0, 1, 2])
            poll_interval: Polling interval in seconds for fallback monitoring
            detection_timeout: Timeout for camera capability detection
            enable_capability_detection: Whether to probe v4l2 capabilities
        """
        self._device_range = device_range or list(range(10))
        self._poll_interval = poll_interval
        self._detection_timeout = detection_timeout
        self._enable_capability_detection = enable_capability_detection
        
        self._logger = logging.getLogger(__name__)
        self._running = False
        self._event_handlers: List[CameraEventHandler] = []
        self._event_callbacks: List[Callable[[CameraEventData], None]] = []
        
        # Internal state tracking
        self._known_devices: Dict[str, CameraDevice] = {}
        self._monitoring_tasks: List[asyncio.Task] = []
        
        # Udev monitoring objects
        self._udev_context: Optional[pyudev.Context] = None
        self._udev_monitor: Optional[pyudev.Monitor] = None
        self._udev_available = HAS_PYUDEV
        
        if not self._udev_available:
            self._logger.warning("pyudev not available - falling back to polling-only monitoring")
        
        self._logger.debug(
            f"Initialized HybridCameraMonitor with device_range={self._device_range}, "
            f"poll_interval={self._poll_interval}s, udev_available={self._udev_available}"
        )
    
    def add_event_handler(self, handler: CameraEventHandler) -> None:
        """
        Add a camera event handler.
        
        Args:
            handler: Event handler implementing CameraEventHandler interface
        """
        if handler not in self._event_handlers:
            self._event_handlers.append(handler)
            self._logger.debug(f"Added event handler: {handler.__class__.__name__}")
    
    def remove_event_handler(self, handler: CameraEventHandler) -> None:
        """
        Remove a camera event handler.
        
        Args:
            handler: Event handler to remove
        """
        if handler in self._event_handlers:
            self._event_handlers.remove(handler)
            self._logger.debug(f"Removed event handler: {handler.__class__.__name__}")
    
    def add_event_callback(self, callback: Callable[[CameraEventData], None]) -> None:
        """
        Add a callback function for camera events.
        
        Args:
            callback: Function to call when camera events occur
        """
        if callback not in self._event_callbacks:
            self._event_callbacks.append(callback)
            self._logger.debug("Added event callback function")
    
    def remove_event_callback(self, callback: Callable[[CameraEventData], None]) -> None:
        """
        Remove a callback function for camera events.
        
        Args:
            callback: Function to remove from callbacks
        """
        if callback in self._event_callbacks:
            self._event_callbacks.remove(callback)
            self._logger.debug("Removed event callback function")
    
    async def start(self) -> None:
        """
        Start the hybrid camera monitoring system.
        
        Initializes both udev event monitoring and polling fallback systems.
        """
        if self._running:
            self._logger.warning("Camera monitor is already running")
            return
        
        self._logger.info("Starting hybrid camera monitor")
        self._running = True
        
        try:
            # Initialize udev monitoring if available
            if self._udev_available:
                await self._setup_udev_monitoring()
                
                # Start udev event monitoring task
                udev_task = asyncio.create_task(self._udev_event_loop())
                self._monitoring_tasks.append(udev_task)
            
            # Start polling fallback
            polling_task = asyncio.create_task(self._polling_loop())
            self._monitoring_tasks.append(polling_task)
            
            # Perform initial camera discovery
            await self._initial_discovery()
            
            self._logger.info("Hybrid camera monitor started successfully")
            
        except Exception as e:
            self._logger.error(f"Failed to start camera monitor: {e}")
            await self.stop()
            raise
    
    async def stop(self) -> None:
        """
        Stop the hybrid camera monitoring system.
        
        Cleanly shuts down all monitoring tasks and releases resources.
        """
        if not self._running:
            return
        
        self._logger.info("Stopping hybrid camera monitor")
        self._running = False
        
        # Cancel all monitoring tasks
        for task in self._monitoring_tasks:
            if not task.done():
                task.cancel()
        
        # Wait for tasks to complete
        if self._monitoring_tasks:
            await asyncio.gather(*self._monitoring_tasks, return_exceptions=True)
        
        self._monitoring_tasks.clear()
        
        # Clean up udev resources
        if self._udev_available:
            await self._cleanup_udev_monitoring()
        
        # Clear known devices
        self._known_devices.clear()
        
        self._logger.info("Hybrid camera monitor stopped")
    
    async def get_connected_cameras(self) -> Dict[str, CameraDevice]:
        """
        Get currently connected cameras.
        
        Returns:
            Dictionary mapping device paths to camera device information
        """
        return self._known_devices.copy()
    
    async def refresh_camera_list(self) -> None:
        """
        Force a refresh of the camera list.
        
        Triggers immediate discovery of all cameras in the configured range.
        """
        self._logger.debug("Refreshing camera list")
        await self._discover_cameras()
    
    async def _setup_udev_monitoring(self) -> None:
        """
        Initialize udev monitoring for real-time camera events.
        
        Sets up udev context and monitor for USB video device events.
        """
        if not self._udev_available:
            return
            
        try:
            # Initialize udev context
            self._udev_context = pyudev.Context()
            
            # Create udev monitor for 'video4linux' subsystem
            self._udev_monitor = pyudev.Monitor.from_netlink(self._udev_context)
            
            # Set up event filtering for USB video devices
            self._udev_monitor.filter_by(subsystem='video4linux')
            
            # Configure monitor socket for async operation
            # Set monitor to non-blocking mode for async polling
            self._udev_monitor.start()
            
            self._logger.info("Udev monitoring initialized successfully")
            
        except Exception as e:
            self._logger.error(f"Failed to setup udev monitoring: {e}")
            self._udev_available = False
            self._udev_context = None
            self._udev_monitor = None
            raise
    
    async def _cleanup_udev_monitoring(self) -> None:
        """
        Clean up udev monitoring resources.
        
        Properly releases udev context and monitor objects.
        """
        try:
            # Close udev monitor
            if self._udev_monitor:
                self._udev_monitor = None
                
            # Release udev context  
            if self._udev_context:
                self._udev_context = None
                
            self._logger.debug("Udev monitoring resources cleaned up")
            
        except Exception as e:
            self._logger.warning(f"Error during udev cleanup: {e}")
    
    async def _udev_event_loop(self) -> None:
        """
        Main loop for processing udev events.
        
        Monitors udev socket for camera connect/disconnect events and
        processes them in real-time.
        """
        if not self._udev_available or not self._udev_monitor:
            return
            
        self._logger.debug("Starting udev event loop")
        
        try:
            while self._running:
                # Poll udev monitor socket with timeout
                # Use non-blocking poll to check for events
                try:
                    # Poll for events with a short timeout to allow cancellation
                    device = self._udev_monitor.poll(timeout=0.1)
                    
                    if device is not None:
                        # Process udev device events
                        await self._process_udev_device_event(device)
                    else:
                        # No event, yield control briefly
                        await asyncio.sleep(0.01)
                        
                except Exception as poll_error:
                    self._logger.warning(f"Udev poll error: {poll_error}")
                    await asyncio.sleep(0.1)
                    
        except asyncio.CancelledError:
            self._logger.debug("Udev event loop cancelled")
        except Exception as e:
            self._logger.error(f"Error in udev event loop: {e}")
    
    async def _process_udev_device_event(self, device) -> None:
        """
        Process a single udev device event.
        
        Args:
            device: pyudev.Device object representing the event
        """
        try:
            # Extract device information and event type
            device_path = device.device_node
            action = device.action
            
            # Filter for video devices in our monitored range
            if not device_path or not device_path.startswith('/dev/video'):
                return
                
            # Extract device number and check if it's in our range
            match = re.search(r'/dev/video(\d+)', device_path)
            if not match:
                return
                
            device_num = int(match.group(1))
            if device_num not in self._device_range:
                return
            
            self._logger.debug(f"Processing udev event: {action} for {device_path}")
            
            # Map udev actions to camera events
            if action == 'add':
                # Device connected
                device_info = await self._create_camera_device_info(device_path, device_num)
                event_data = CameraEventData(
                    device_path=device_path,
                    event_type=CameraEvent.CONNECTED,
                    device_info=device_info
                )
                
                # Update known devices and handle event
                self._known_devices[device_path] = device_info
                await self._handle_camera_event(event_data)
                
            elif action == 'remove':
                # Device disconnected
                device_info = self._known_devices.get(device_path)
                event_data = CameraEventData(
                    device_path=device_path,
                    event_type=CameraEvent.DISCONNECTED,
                    device_info=device_info
                )
                
                # Remove from known devices and handle event
                if device_path in self._known_devices:
                    del self._known_devices[device_path]
                await self._handle_camera_event(event_data)
                
        except Exception as e:
            self._logger.error(f"Error processing udev event: {e}")
    
    async def _create_camera_device_info(self, device_path: str, device_num: int) -> CameraDevice:
        """
        Create CameraDevice info for a detected device.
        
        Args:
            device_path: Path to the video device
            device_num: Device number extracted from path
            
        Returns:
            CameraDevice object with device information
        """
        device_name = f"Camera {device_num}"
        device_status = await self._determine_device_status(device_path)
        
        return CameraDevice(
            device=device_path,
            name=device_name,
            status=device_status
        )
    
    async def _polling_loop(self) -> None:
        """
        Polling fallback loop for camera discovery.
        
        Periodically scans for cameras as a fallback mechanism when
        udev events might be missed or unavailable.
        """
        self._logger.debug("Starting polling fallback loop")
        
        try:
            while self._running:
                await self._discover_cameras()
                await asyncio.sleep(self._poll_interval)
                
        except asyncio.CancelledError:
            self._logger.debug("Polling loop cancelled")
        except Exception as e:
            self._logger.error(f"Error in polling loop: {e}")
    
    async def _initial_discovery(self) -> None:
        """
        Perform initial camera discovery on startup.
        
        Scans all configured device paths to establish baseline state.
        """
        self._logger.info("Performing initial camera discovery")
        await self._discover_cameras()
    
    async def _discover_cameras(self) -> None:
        """
        Discover cameras by scanning configured device paths.
        
        Checks each device path in the configured range and detects
        changes compared to known device state.
        """
        current_devices: Dict[str, CameraDevice] = {}
        
        for device_num in self._device_range:
            device_path = f"/dev/video{device_num}"
            
            # Check if device path exists
            if Path(device_path).exists():
                try:
                    # Extract device number and create proper camera name
                    device_name = f"Camera {device_num}"
                    
                    # Determine device status based on accessibility
                    device_status = await self._determine_device_status(device_path)
                    
                    # Probe device capabilities if enabled
                    capabilities = None
                    if self._enable_capability_detection:
                        capabilities = await self._probe_device_capabilities(device_path)
                    
                    # Create CameraDevice with detected information
                    device_info = CameraDevice(
                        device=device_path,
                        name=device_name,
                        status=device_status
                    )
                    
                    # Add capabilities if detected
                    if capabilities:
                        # Add capability information to device_info if the CameraDevice supports it
                        # This depends on the actual CameraDevice structure
                        pass
                    
                    current_devices[device_path] = device_info
                    
                except Exception as e:
                    self._logger.warning(f"Error probing device {device_path}: {e}")
                    # Create device with error status
                    device_info = CameraDevice(
                        device=device_path,
                        name=f"Camera {device_num}",
                        status="ERROR"
                    )
                    current_devices[device_path] = device_info
        
        # Compare with known devices and generate events only if not using udev
        # (udev events will handle real-time updates)
        if not self._udev_available:
            await self._process_device_changes(current_devices)
        else:
            # When udev is available, only update known devices for initial discovery
            if not self._known_devices:
                self._known_devices = current_devices.copy()
    
    async def _determine_device_status(self, device_path: str) -> str:
        """
        Determine the status of a camera device.
        
        Args:
            device_path: Path to video device (e.g., /dev/video0)
            
        Returns:
            Device status string ("CONNECTED", "DISCONNECTED", "ERROR", "BUSY")
        """
        try:
            # Check if device is accessible by attempting to open it
            # This is a basic check - more sophisticated probing could be added
            device_file = Path(device_path)
            
            if not device_file.exists():
                return "DISCONNECTED"
            
            # Check if device is readable (indicates it's accessible)
            if device_file.is_char_device():
                # Device exists and is a character device
                # Additional checks could be added here for device availability
                return "CONNECTED"
            else:
                return "ERROR"
                
        except Exception as e:
            self._logger.debug(f"Error determining status for {device_path}: {e}")
            return "ERROR"
    
    async def _process_device_changes(self, current_devices: Dict[str, CameraDevice]) -> None:
        """
        Process changes between current and known device states.
        
        Args:
            current_devices: Currently discovered devices
        """
        # Detect new devices
        for device_path, device_info in current_devices.items():
            if device_path not in self._known_devices:
                await self._handle_camera_event(CameraEventData(
                    device_path=device_path,
                    event_type=CameraEvent.CONNECTED,
                    device_info=device_info
                ))
        
        # Detect removed devices
        for device_path in list(self._known_devices.keys()):
            if device_path not in current_devices:
                await self._handle_camera_event(CameraEventData(
                    device_path=device_path,
                    event_type=CameraEvent.DISCONNECTED,
                    device_info=self._known_devices[device_path]
                ))
        
        # Detect status changes for existing devices
        for device_path, device_info in current_devices.items():
            if device_path in self._known_devices:
                if self._known_devices[device_path].status != device_info.status:
                    await self._handle_camera_event(CameraEventData(
                        device_path=device_path,
                        event_type=CameraEvent.STATUS_CHANGED,
                        device_info=device_info
                    ))
        
        # Update known devices
        self._known_devices = current_devices.copy()
    
    async def _handle_camera_event(self, event_data: CameraEventData) -> None:
        """
        Handle camera events by notifying all registered handlers and callbacks.
        
        Args:
            event_data: Camera event information
        """
        self._logger.info(
            f"Camera event: {event_data.event_type.value} - {event_data.device_path}"
        )
        
        # Notify event handlers
        for handler in self._event_handlers:
            try:
                await handler.handle_camera_event(event_data)
            except Exception as e:
                self._logger.error(f"Error in event handler {handler.__class__.__name__}: {e}")
        
        # Notify callback functions
        for callback in self._event_callbacks:
            try:
                # TODO: Consider making callbacks async or running in thread pool
                callback(event_data)
            except Exception as e:
                self._logger.error(f"Error in event callback: {e}")
    
    async def _probe_device_capabilities(self, device_path: str) -> Optional[Dict]:
        """
        Probe camera device capabilities using v4l2.
        
        Args:
            device_path: Path to video device (e.g., /dev/video0)
            
        Returns:
            Dictionary containing device capabilities or None if probe failed
        """
        if not self._enable_capability_detection:
            return None
        
        try:
            # Basic capability detection - this could be expanded with actual v4l2 probing
            capabilities = {
                "device_path": device_path,
                "detected": True,
                "accessible": Path(device_path).exists()
            }
            
            # TODO: Use v4l2-ctl or python v4l2 bindings to probe:
            # TODO: - Supported formats and resolutions
            # TODO: - Frame rates
            # TODO: - Device name and capabilities
            # TODO: - Driver information
            
            self._logger.debug(f"Basic capability detection for {device_path}: {capabilities}")
            return capabilities
            
        except Exception as e:
            self._logger.warning(f"Failed to probe capabilities for {device_path}: {e}")
            return None

    def get_stream_name_from_device_path(self, device_path: str) -> str:
        """
        Extract stream name from camera device path.
        
        Args:
            device_path: Camera device path (e.g., /dev/video0)
            
        Returns:
            Stream name for MediaMTX (e.g., camera0)
        """
        try:
            # Use regex to extract device number from path
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