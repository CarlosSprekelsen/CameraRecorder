"""
Hybrid camera discovery monitor implementation.

Provides real-time USB camera detection using udev events with polling
fallback for reliability, as specified in the architecture design.
"""

import asyncio
import logging
import re
import subprocess
import json
from abc import ABC, abstractmethod
from dataclasses import dataclass
from enum import Enum
from pathlib import Path
from typing import Callable, Dict, List, Optional, Set, Tuple, Any

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


@dataclass
class CapabilityDetectionResult:
    """Structured result from capability detection."""
    device_path: str
    detected: bool
    accessible: bool
    device_name: Optional[str] = None
    driver: Optional[str] = None
    formats: List[Dict[str, Any]] = None
    resolutions: List[str] = None
    frame_rates: List[str] = None
    error: Optional[str] = None
    timeout_context: Optional[str] = None
    
    def __post_init__(self):
        if self.formats is None:
            self.formats = []
        if self.resolutions is None:
            self.resolutions = []
        if self.frame_rates is None:
            self.frame_rates = []


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
    
    Architecture Decision: Hybrid udev + polling approach provides real-time
    events when available while ensuring discovery completeness through polling.
    Priority order: udev events (real-time) > polling (fallback/validation).
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
        
        # Internal state tracking with thread safety considerations
        self._known_devices: Dict[str, CameraDevice] = {}
        self._monitoring_tasks: List[asyncio.Task] = []
        self._state_lock = asyncio.Lock()  # Protect against race conditions
        
        # Udev monitoring objects
        self._udev_context: Optional[pyudev.Context] = None
        self._udev_monitor: Optional[pyudev.Monitor] = None
        self._udev_available = HAS_PYUDEV
        
        # Diagnostic counters for observability
        self._stats = {
            'udev_events_processed': 0,
            'udev_events_filtered': 0,
            'polling_cycles': 0,
            'capability_probes_attempted': 0,
            'capability_probes_successful': 0,
            'capability_timeouts': 0,
            'device_state_changes': 0
        }
        
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
                self._logger.info("Udev event monitoring enabled")
            else:
                self._logger.info("Udev monitoring not available - using polling-only mode")
            
            # Start polling fallback (always enabled for validation and fallback)
            polling_task = asyncio.create_task(self._polling_loop())
            self._monitoring_tasks.append(polling_task)
            self._logger.info("Polling fallback monitoring enabled")
            
            # Perform initial camera discovery
            await self._initial_discovery()
            
            self._logger.info("Hybrid camera monitor started successfully")
            
        except Exception as e:
            self._logger.error(f"Failed to start camera monitor: {e}", exc_info=True)
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
        
        # Wait for tasks to complete with timeout
        if self._monitoring_tasks:
            try:
                await asyncio.wait_for(
                    asyncio.gather(*self._monitoring_tasks, return_exceptions=True),
                    timeout=5.0
                )
            except asyncio.TimeoutError:
                self._logger.warning("Some monitoring tasks did not complete within timeout")
        
        self._monitoring_tasks.clear()
        
        # Clean up udev resources
        if self._udev_available:
            await self._cleanup_udev_monitoring()
        
        # Clear known devices under lock
        async with self._state_lock:
            self._known_devices.clear()
        
        # Log final statistics for diagnostics
        self._logger.info(f"Camera monitor stopped. Final stats: {self._stats}")
    
    async def get_connected_cameras(self) -> Dict[str, CameraDevice]:
        """
        Get currently connected cameras.
        
        Returns:
            Dictionary mapping device paths to camera device information
        """
        async with self._state_lock:
            return self._known_devices.copy()
    
    async def refresh_camera_list(self) -> None:
        """
        Force a refresh of the camera list.
        
        Triggers immediate discovery of all cameras in the configured range.
        """
        self._logger.debug("Refreshing camera list")
        await self._discover_cameras()
    
    def get_monitor_stats(self) -> Dict[str, Any]:
        """
        Get monitoring statistics for diagnostics.
        
        Returns:
            Dictionary containing monitoring metrics and counters
        """
        return {
            **self._stats,
            'running': self._running,
            'udev_available': self._udev_available,
            'device_range': self._device_range,
            'poll_interval': self._poll_interval,
            'known_devices_count': len(self._known_devices),
            'active_tasks': len([t for t in self._monitoring_tasks if not t.done()])
        }
    
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
            self._logger.error(f"Failed to setup udev monitoring: {e}", exc_info=True)
            self._udev_available = False
            self._udev_context = None
            self._udev_monitor = None
            # Don't raise - continue with polling-only mode
    
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
        processes them in real-time with robust error handling.
        """
        if not self._udev_available or not self._udev_monitor:
            return
            
        self._logger.debug("Starting udev event loop")
        
        consecutive_poll_errors = 0
        max_consecutive_errors = 10
        
        try:
            while self._running:
                try:
                    # Poll udev monitor socket with timeout
                    device = self._udev_monitor.poll(timeout=0.1)
                    
                    if device is not None:
                        # Reset error counter on successful poll
                        consecutive_poll_errors = 0
                        
                        # Process udev device events
                        await self._process_udev_device_event(device)
                        self._stats['udev_events_processed'] += 1
                    else:
                        # No event, yield control briefly
                        await asyncio.sleep(0.01)
                        
                except Exception as poll_error:
                    consecutive_poll_errors += 1
                    self._logger.warning(
                        f"Udev poll error (#{consecutive_poll_errors}): {poll_error}"
                    )
                    
                    # Backoff strategy for repeated errors
                    if consecutive_poll_errors >= max_consecutive_errors:
                        self._logger.error(
                            f"Too many consecutive udev poll errors ({consecutive_poll_errors}), "
                            "disabling udev monitoring"
                        )
                        self._udev_available = False
                        break
                    
                    # Exponential backoff: 0.1s, 0.2s, 0.4s, max 2s
                    backoff_delay = min(0.1 * (2 ** consecutive_poll_errors), 2.0)
                    await asyncio.sleep(backoff_delay)
                    
        except asyncio.CancelledError:
            self._logger.debug("Udev event loop cancelled")
        except Exception as e:
            self._logger.error(f"Critical error in udev event loop: {e}", exc_info=True)
    
    async def _process_udev_device_event(self, device) -> None:
        """
        Process a single udev device event with comprehensive validation and error handling.
        
        Args:
            device: pyudev.Device object representing the event
        """
        device_path = None
        try:
            # Extract device information and event type
            device_path = device.device_node
            action = device.action
            
            # Validate device path exists and is a video device
            if not device_path:
                self._logger.debug("Skipping udev event with no device_node")
                self._stats['udev_events_filtered'] += 1
                return
                
            if not device_path.startswith('/dev/video'):
                self._logger.debug(f"Skipping non-video device: {device_path}")
                self._stats['udev_events_filtered'] += 1
                return
                
            # Extract device number and validate range
            device_match = re.search(r'/dev/video(\d+)', device_path)
            if not device_match:
                self._logger.debug(f"Could not extract device number from: {device_path}")
                self._stats['udev_events_filtered'] += 1
                return
                
            device_num = int(device_match.group(1))
            if device_num not in self._device_range:
                self._logger.debug(
                    f"Device {device_path} (num={device_num}) not in monitored range {self._device_range}"
                )
                self._stats['udev_events_filtered'] += 1
                return
            
            self._logger.debug(f"Processing udev event: {action} for {device_path}")
            
            # Process event based on action type with state protection
            async with self._state_lock:
                if action == 'add':
                    await self._handle_udev_device_added(device_path, device_num)
                elif action == 'remove':
                    await self._handle_udev_device_removed(device_path)
                elif action == 'change':
                    await self._handle_udev_device_changed(device_path, device_num)
                else:
                    self._logger.debug(f"Ignoring udev action '{action}' for {device_path}")
                    self._stats['udev_events_filtered'] += 1
                
        except Exception as e:
            device_context = device_path or "unknown device"
            self._logger.error(
                f"Error processing udev event for {device_context}: {e}",
                exc_info=True
            )
            # Continue processing - don't let one event break the loop
    
    async def _handle_udev_device_added(self, device_path: str, device_num: int) -> None:
        """Handle udev 'add' event for device connection."""
        # Verify device is actually accessible before creating event
        device_info = await self._create_camera_device_info(device_path, device_num)
        
        if device_info and device_info.status == "CONNECTED":
            event_data = CameraEventData(
                device_path=device_path,
                event_type=CameraEvent.CONNECTED,
                device_info=device_info,
                timestamp=asyncio.get_event_loop().time()
            )
            
            # Update known devices and handle event
            self._known_devices[device_path] = device_info
            self._stats['device_state_changes'] += 1
            await self._handle_camera_event(event_data)
        else:
            self._logger.warning(
                f"Device {device_path} detected via udev 'add' but not accessible"
            )
    
    async def _handle_udev_device_removed(self, device_path: str) -> None:
        """Handle udev 'remove' event for device disconnection."""
        device_info = self._known_devices.get(device_path)
        event_data = CameraEventData(
            device_path=device_path,
            event_type=CameraEvent.DISCONNECTED,
            device_info=device_info,
            timestamp=asyncio.get_event_loop().time()
        )
        
        # Remove from known devices and handle event
        if device_path in self._known_devices:
            del self._known_devices[device_path]
            self._stats['device_state_changes'] += 1
        
        await self._handle_camera_event(event_data)
    
    async def _handle_udev_device_changed(self, device_path: str, device_num: int) -> None:
        """Handle udev 'change' event for device property changes."""
        device_info = await self._create_camera_device_info(device_path, device_num)
        old_device_info = self._known_devices.get(device_path)
        
        # Only generate event if status actually changed
        if device_info and old_device_info and device_info.status != old_device_info.status:
            event_data = CameraEventData(
                device_path=device_path,
                event_type=CameraEvent.STATUS_CHANGED,
                device_info=device_info,
                timestamp=asyncio.get_event_loop().time()
            )
            
            # Update known devices and handle event
            self._known_devices[device_path] = device_info
            self._stats['device_state_changes'] += 1
            await self._handle_camera_event(event_data)
        else:
            self._logger.debug(f"No significant status change for {device_path}")
    
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
        udev events might be missed or unavailable. Also serves as
        validation for udev-based discovery.
        """
        self._logger.debug("Starting polling fallback loop")
        
        polling_error_count = 0
        max_polling_errors = 5
        
        try:
            while self._running:
                try:
                    await self._discover_cameras()
                    self._stats['polling_cycles'] += 1
                    polling_error_count = 0  # Reset on success
                    
                    await asyncio.sleep(self._poll_interval)
                    
                except Exception as e:
                    polling_error_count += 1
                    self._logger.error(
                        f"Polling discovery error (#{polling_error_count}): {e}",
                        exc_info=True
                    )
                    
                    if polling_error_count >= max_polling_errors:
                        self._logger.critical(
                            f"Too many consecutive polling errors ({polling_error_count}), "
                            "stopping polling loop"
                        )
                        break
                    
                    # Backoff on errors
                    await asyncio.sleep(min(self._poll_interval * 2, 5.0))
                    
        except asyncio.CancelledError:
            self._logger.debug("Polling loop cancelled")
        except Exception as e:
            self._logger.error(f"Critical error in polling loop: {e}", exc_info=True)
    
    async def _initial_discovery(self) -> None:
        """
        Perform initial camera discovery on startup.
        
        Scans all configured device paths to establish baseline state.
        """
        self._logger.info("Performing initial camera discovery")
        try:
            await self._discover_cameras()
            device_count = len(self._known_devices)
            self._logger.info(f"Initial discovery completed - found {device_count} cameras")
        except Exception as e:
            self._logger.error(f"Initial discovery failed: {e}", exc_info=True)
            # Don't raise - allow service to start even if initial discovery fails
    
    async def _discover_cameras(self) -> None:
        """
        Discover cameras by scanning configured device paths.
        
        Checks each device path in the configured range and detects
        changes compared to known device state.
        """
        current_devices: Dict[str, CameraDevice] = {}
        discovery_errors = []
        
        for device_num in self._device_range:
            device_path = f"/dev/video{device_num}"
            
            # Check if device path exists
            if Path(device_path).exists():
                try:
                    # Extract device number and create proper camera name
                    device_name = f"Camera {device_num}"
                    
                    # Determine device status based on accessibility
                    device_status = await self._determine_device_status(device_path)
                    
                    # Create CameraDevice with detected information
                    device_info = CameraDevice(
                        device=device_path,
                        name=device_name,
                        status=device_status
                    )
                    
                    # STOP: MEDIUM: Capability metadata integration deferred pending test validation [IV&V:S3]
                    # Rationale: Capability detection is implemented but integration with CameraDevice
                    # is deferred until comprehensive testing validates all v4l2 output variations.
                    # Current implementation uses defaults with capability probing for validation.
                    # Owner: Solo engineer | Date: 2025-08-03
                    
                    current_devices[device_path] = device_info
                    
                except Exception as e:
                    discovery_errors.append(f"{device_path}: {e}")
                    self._logger.warning(f"Error probing device {device_path}: {e}")
                    
                    # Create device with error status for tracking
                    device_info = CameraDevice(
                        device=device_path,
                        name=f"Camera {device_num}",
                        status="ERROR"
                    )
                    current_devices[device_path] = device_info
        
        # Log discovery errors for diagnostics
        if discovery_errors:
            self._logger.debug(f"Discovery errors encountered: {discovery_errors}")
        
        # Compare with known devices and generate events (only when udev unavailable)
        # When udev is available, it handles real-time updates and polling serves as validation
        async with self._state_lock:
            if not self._udev_available:
                await self._process_device_changes(current_devices)
            else:
                # Update known devices for initial discovery or validation
                if not self._known_devices:
                    self._known_devices = current_devices.copy()
                    self._logger.debug("Updated baseline device state from discovery")
    
    async def _determine_device_status(self, device_path: str) -> str:
        """
        Determine the status of a camera device with enhanced error context.
        
        Args:
            device_path: Path to video device (e.g., /dev/video0)
            
        Returns:
            Device status string ("CONNECTED", "DISCONNECTED", "ERROR", "BUSY")
        """
        try:
            device_file = Path(device_path)
            
            if not device_file.exists():
                return "DISCONNECTED"
            
            # Check if device is a character device (proper V4L2 device)
            if not device_file.is_char_device():
                self._logger.debug(f"Device {device_path} is not a character device")
                return "ERROR"
            
            # Attempt to query device capabilities to verify accessibility
            try:
                process = await asyncio.wait_for(
                    asyncio.create_subprocess_exec(
                        'v4l2-ctl', '--device', device_path, '--list-formats-ext',
                        stdout=asyncio.subprocess.PIPE,
                        stderr=asyncio.subprocess.PIPE
                    ),
                    timeout=self._detection_timeout
                )
                
                stdout, stderr = await process.communicate()
                
                if process.returncode == 0:
                    return "CONNECTED"
                else:
                    stderr_text = stderr.decode().lower()
                    if any(keyword in stderr_text for keyword in ['busy', 'device or resource busy']):
                        return "BUSY"
                    else:
                        self._logger.debug(f"v4l2-ctl failed for {device_path}: {stderr_text}")
                        return "ERROR"
                        
            except asyncio.TimeoutError:
                self._logger.debug(f"Timeout determining status for {device_path}")
                return "ERROR"
            except Exception as subprocess_error:
                self._logger.debug(f"Subprocess error for {device_path}: {subprocess_error}")
                # Fallback - device exists as character device but can't query
                return "CONNECTED"
                
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
                    device_info=device_info,
                    timestamp=asyncio.get_event_loop().time()
                ))
                self._stats['device_state_changes'] += 1
        
        # Detect removed devices
        for device_path in list(self._known_devices.keys()):
            if device_path not in current_devices:
                await self._handle_camera_event(CameraEventData(
                    device_path=device_path,
                    event_type=CameraEvent.DISCONNECTED,
                    device_info=self._known_devices[device_path],
                    timestamp=asyncio.get_event_loop().time()
                ))
                self._stats['device_state_changes'] += 1
        
        # Detect status changes for existing devices
        for device_path, device_info in current_devices.items():
            if device_path in self._known_devices:
                if self._known_devices[device_path].status != device_info.status:
                    await self._handle_camera_event(CameraEventData(
                        device_path=device_path,
                        event_type=CameraEvent.STATUS_CHANGED,
                        device_info=device_info,
                        timestamp=asyncio.get_event_loop().time()
                    ))
                    self._stats['device_state_changes'] += 1
        
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
        
        # Notify event handlers with individual error handling
        handler_errors = []
        for handler in self._event_handlers:
            try:
                await handler.handle_camera_event(event_data)
            except Exception as e:
                handler_error = f"{handler.__class__.__name__}: {e}"
                handler_errors.append(handler_error)
                self._logger.error(f"Error in event handler {handler_error}", exc_info=True)
        
        # Notify callback functions with individual error handling
        callback_errors = []
        for callback in self._event_callbacks:
            try:
                # Call callbacks in thread pool to avoid blocking async loop
                callback(event_data)
            except Exception as e:
                callback_error = f"callback: {e}"
                callback_errors.append(callback_error)
                self._logger.error(f"Error in event {callback_error}", exc_info=True)
        
        # Log summary if there were errors
        if handler_errors or callback_errors:
            all_errors = handler_errors + callback_errors
            self._logger.warning(
                f"Event notification errors for {event_data.device_path}: {all_errors}"
            )
    
    async def _probe_device_capabilities(self, device_path: str) -> Optional[CapabilityDetectionResult]:
        """
        Probe camera device capabilities using v4l2-ctl with robust error handling.
        
        Uses v4l2-ctl command to probe device capabilities including supported
        formats, resolutions, frame rates, and device information.
        Provides graceful fallback if v4l2-ctl is unavailable or probing fails.
        
        Args:
            device_path: Path to video device (e.g., /dev/video0)
            
        Returns:
            CapabilityDetectionResult with detection results or error information
        """
        if not self._enable_capability_detection:
            return None
        
        self._stats['capability_probes_attempted'] += 1
        
        try:
            self._logger.debug(f"Probing capabilities for {device_path}")
            
            result = CapabilityDetectionResult(
                device_path=device_path,
                detected=False,
                accessible=Path(device_path).exists()
            )
            
            # Probe device information and capabilities
            device_info = await self._probe_device_info_robust(device_path)
            if device_info:
                result.device_name = device_info.get("device_name")
                result.driver = device_info.get("driver")
            
            # Probe supported formats and resolutions
            formats_info = await self._probe_device_formats_robust(device_path)
            if formats_info:
                result.formats = formats_info["formats"]
                result.resolutions = formats_info["resolutions"]
            
            # Probe frame rate capabilities
            framerates_info = await self._probe_device_framerates_robust(device_path)
            if framerates_info:
                result.frame_rates = framerates_info
            
            # Mark as successful if we got any useful data
            if device_info or formats_info or framerates_info:
                result.detected = True
                self._stats['capability_probes_successful'] += 1
                self._logger.debug(
                    f"Capability detection completed for {device_path}: "
                    f"{len(result.formats)} formats, {len(result.resolutions)} resolutions"
                )
            else:
                result.error = "No capability data retrieved"
                self._logger.debug(f"No capability data retrieved for {device_path}")
            
            return result
            
        except asyncio.TimeoutError:
            self._stats['capability_timeouts'] += 1
            return CapabilityDetectionResult(
                device_path=device_path,
                detected=False,
                accessible=Path(device_path).exists(),
                error="Capability detection timeout",
                timeout_context=f"Total timeout after {self._detection_timeout}s"
            )
        except Exception as e:
            self._logger.warning(f"Failed to probe capabilities for {device_path}: {e}")
            return CapabilityDetectionResult(
                device_path=device_path,
                detected=False,
                accessible=Path(device_path).exists(),
                error=str(e)
            )

    async def _probe_device_info_robust(self, device_path: str) -> Optional[Dict[str, str]]:
        """
        Probe device information using v4l2-ctl --info with robust parsing.
        
        Args:
            device_path: Path to video device
            
        Returns:
            Dictionary with device info or None if failed
        """
        try:
            process = await asyncio.wait_for(
                asyncio.create_subprocess_exec(
                    'v4l2-ctl', '--device', device_path, '--info',
                    stdout=asyncio.subprocess.PIPE,
                    stderr=asyncio.subprocess.PIPE
                ),
                timeout=self._detection_timeout
            )
            
            stdout, stderr = await process.communicate()
            
            if process.returncode != 0:
                self._logger.debug(f"v4l2-ctl --info failed for {device_path}: {stderr.decode()}")
                return None
            
            info_output = stdout.decode()
            device_info = {}
            
            # Robust parsing with multiple patterns for each field
            info_patterns = {
                "device_name": [
                    r"Device name\s*[:\s]\s*(.+)",
                    r"Name\s*[:\s]\s*(.+)",
                    r"Device\s*[:\s]\s*(.+)"
                ],
                "driver": [
                    r"Driver name\s*[:\s]\s*(.+)",
                    r"Driver\s*[:\s]\s*(.+)"
                ],
                "card_type": [
                    r"Card type\s*[:\s]\s*(.+)",
                    r"Card\s*[:\s]\s*(.+)"
                ]
            }
            
            for field, patterns in info_patterns.items():
                for pattern in patterns:
                    match = re.search(pattern, info_output, re.IGNORECASE)
                    if match:
                        device_info[field] = match.group(1).strip()
                        break
            
            return device_info if device_info else None
            
        except asyncio.TimeoutError:
            self._logger.debug(f"Timeout probing device info for {device_path}")
            return None
        except Exception as e:
            self._logger.debug(f"Error probing device info for {device_path}: {e}")
            return None

    async def _probe_device_formats_robust(self, device_path: str) -> Optional[Dict[str, Any]]:
        """
        Probe supported formats and resolutions using v4l2-ctl --list-formats-ext with robust parsing.
        
        Args:
            device_path: Path to video device
            
        Returns:
            Dictionary with formats and resolutions or None if failed
        """
        try:
            process = await asyncio.wait_for(
                asyncio.create_subprocess_exec(
                    'v4l2-ctl', '--device', device_path, '--list-formats-ext',
                    stdout=asyncio.subprocess.PIPE,
                    stderr=asyncio.subprocess.PIPE
                ),
                timeout=self._detection_timeout
            )
            
            stdout, stderr = await process.communicate()
            
            if process.returncode != 0:
                self._logger.debug(f"v4l2-ctl --list-formats-ext failed for {device_path}: {stderr.decode()}")
                return None
            
            formats_output = stdout.decode()
            formats = []
            resolutions = set()
            
            current_format = None
            
            # Robust parsing with multiple format and resolution patterns
            for line in formats_output.splitlines():
                line = line.strip()
                
                # Format line patterns - handle various v4l2-ctl output formats
                format_patterns = [
                    r'\[\d+\]:\s*[\'"]([^\'\"]+)[\'\"]\s*\(([^)]+)\)',  # Standard: [0]: 'YUYV' (YUYV 4:2:2)
                    r'\[\d+\]:\s*([A-Z0-9]+)\s*\(([^)]+)\)',           # Alternative: [0]: YUYV (YUYV 4:2:2)
                    r'([A-Z0-9]{3,4})\s*[:\s]\s*([^:\n]+)'             # Fallback: YUYV : YUYV 4:2:2
                ]
                
                format_matched = False
                for pattern in format_patterns:
                    format_match = re.search(pattern, line)
                    if format_match:
                        format_code = format_match.group(1).strip()
                        format_desc = format_match.group(2).strip() if len(format_match.groups()) > 1 else format_code
                        current_format = {
                            "code": format_code,
                            "description": format_desc,
                            "resolutions": []
                        }
                        formats.append(current_format)
                        format_matched = True
                        break
                
                if format_matched:
                    continue
                
                # Resolution line patterns - handle various output formats
                if current_format:
                    resolution_patterns = [
                        r'Size:\s*Discrete\s+(\d+)x(\d+)',              # Standard: Size: Discrete 640x480
                        r'(\d+)x(\d+)',                                  # Simple: 640x480
                        r'Size:\s*(\d+)\s*x\s*(\d+)',                   # Alternative: Size: 640 x 480
                        r'Resolution:\s*(\d+)x(\d+)'                     # Alternative: Resolution: 640x480
                    ]
                    
                    for pattern in resolution_patterns:
                        resolution_match = re.search(pattern, line)
                        if resolution_match:
                            width = int(resolution_match.group(1))
                            height = int(resolution_match.group(2))
                            resolution = f"{width}x{height}"
                            current_format["resolutions"].append(resolution)
                            resolutions.add(resolution)
                            break
            
            # Fallback resolution extraction if no format-specific parsing worked
            if not resolutions:
                # Look for any resolution patterns in the entire output
                fallback_resolutions = re.findall(r'(\d{3,4})x(\d{3,4})', formats_output)
                for width_str, height_str in fallback_resolutions:
                    width, height = int(width_str), int(height_str)
                    # Filter reasonable camera resolutions
                    if 160 <= width <= 4096 and 120 <= height <= 3072:
                        resolution = f"{width}x{height}"
                        resolutions.add(resolution)
            
            return {
                "formats": formats,
                "resolutions": sorted(list(resolutions))
            }
            
        except asyncio.TimeoutError:
            self._logger.debug(f"Timeout probing device formats for {device_path}")
            return None
        except Exception as e:
            self._logger.debug(f"Error probing device formats for {device_path}: {e}")
            return None

    async def _probe_device_framerates_robust(self, device_path: str) -> Optional[List[str]]:
        """
        Probe supported frame rates using v4l2-ctl with robust parsing and multiple fallbacks.
        
        Args:
            device_path: Path to video device
            
        Returns:
            List of supported frame rates or None if failed
        """
        frame_rates = set()
        
        # Try multiple v4l2-ctl commands and formats to get frame rate information
        commands_to_try = [
            (['v4l2-ctl', '--device', device_path, '--list-framesizes', 'YUYV'], "YUYV framesizes"),
            (['v4l2-ctl', '--device', device_path, '--list-framesizes', 'MJPG'], "MJPG framesizes"), 
            (['v4l2-ctl', '--device', device_path, '--list-framerates'], "general framerates"),
            (['v4l2-ctl', '--device', device_path, '--all'], "all device info")
        ]
        
        for cmd, description in commands_to_try:
            try:
                process = await asyncio.wait_for(
                    asyncio.create_subprocess_exec(
                        *cmd,
                        stdout=asyncio.subprocess.PIPE,
                        stderr=asyncio.subprocess.PIPE
                    ),
                    timeout=self._detection_timeout
                )
                
                stdout, stderr = await process.communicate()
                
                if process.returncode == 0:
                    output = stdout.decode()
                    cmd_frame_rates = self._extract_frame_rates_from_output(output)
                    frame_rates.update(cmd_frame_rates)
                    
                    if frame_rates:
                        self._logger.debug(f"Found frame rates from {description}: {cmd_frame_rates}")
                        
            except asyncio.TimeoutError:
                self._logger.debug(f"Timeout getting {description} for {device_path}")
                continue
            except Exception as e:
                self._logger.debug(f"Error getting {description} for {device_path}: {e}")
                continue
        
        # Return sorted frame rates or reasonable defaults
        if frame_rates:
            return sorted(list(frame_rates), key=lambda x: float(x), reverse=True)
        else:
            # Return common default frame rates if detection fails
            default_rates = ["30", "25", "24", "15", "10", "5"]
            self._logger.debug(f"No frame rates detected for {device_path}, using defaults: {default_rates}")
            return default_rates

    def _extract_frame_rates_from_output(self, output: str) -> Set[str]:
        """
        Extract frame rates from v4l2-ctl output using multiple parsing strategies.
        
        Args:
            output: Raw v4l2-ctl command output
            
        Returns:
            Set of frame rate strings
        """
        frame_rates = set()
        
        # Multiple frame rate patterns to handle different v4l2-ctl output formats
        fps_patterns = [
            r'(\d+(?:\.\d+)?)\s*fps',                          # Standard: 30.000 fps
            r'(\d+(?:\.\d+)?)\s*FPS',                          # Alternative case: 30.000 FPS
            r'Frame\s*rate[:\s]+(\d+(?:\.\d+)?)',              # Frame rate: 30.0
            r'(\d+(?:\.\d+)?)\s*frames?\s*per\s*second',       # frames per second
            r'Interval:\s*\[\d+/(\d+)\]',                      # Interval format: [1/30]
            r'(\d{1,3})\s*Hz',                                 # Frequency: 30 Hz
            r'@\s*(\d+(?:\.\d+)?)',                            # Resolution@framerate: 1920x1080@30
        ]
        
        for pattern in fps_patterns:
            matches = re.findall(pattern, output, re.IGNORECASE)
            for match in matches:
                try:
                    fps_value = float(match)
                    # Filter reasonable frame rates (0.1 to 240 FPS)
                    if 0.1 <= fps_value <= 240:
                        # Convert to integer if it's a whole number
                        if fps_value == int(fps_value):
                            frame_rates.add(str(int(fps_value)))
                        else:
                            frame_rates.add(f"{fps_value:.1f}")
                except (ValueError, TypeError):
                    continue
        
        # Special handling for interval notation [numerator/denominator]
        interval_matches = re.findall(r'\[(\d+)/(\d+)\]', output)
        for num_str, den_str in interval_matches:
            try:
                numerator, denominator = int(num_str), int(den_str)
                if denominator > 0:
                    fps_value = denominator / numerator
                    if 0.1 <= fps_value <= 240:
                        if fps_value == int(fps_value):
                            frame_rates.add(str(int(fps_value)))
                        else:
                            frame_rates.add(f"{fps_value:.1f}")
            except (ValueError, ZeroDivisionError):
                continue
        
        return frame_rates

    def get_stream_name_from_device_path(self, device_path: str) -> str:
        """
        Extract stream name from camera device path with robust fallback handling.
        
        Args:
            device_path: Camera device path (e.g., /dev/video0)
            
        Returns:
            Stream name for MediaMTX (e.g., camera0)
        """
        try:
            # Primary pattern: /dev/video<number>
            match = re.search(r'/dev/video(\d+)', device_path)
            if match:
                device_num = match.group(1)
                return f"camera{device_num}"
            
            # Secondary pattern: any path with video and number
            match = re.search(r'video(\d+)', device_path, re.IGNORECASE)
            if match:
                device_num = match.group(1)
                return f"camera{device_num}"
            
            # Tertiary fallback: extract any digits from the path
            digits = re.findall(r'\d+', device_path)
            if digits:
                return f"camera{digits[-1]}"
            
            # Final fallback: hash-based deterministic name
            hash_val = abs(hash(device_path)) % 1000
            self._logger.debug(f"Using hash-based stream name for {device_path}: camera_{hash_val}")
            return f"camera_{hash_val}"
            
        except Exception as e:
            self._logger.warning(f"Error extracting stream name from {device_path}: {e}")
            return "camera_unknown"

    # TODO: MEDIUM: Add test injection hooks for capability detection validation [Story:E1/S3]
    # Purpose: Enable comprehensive unit testing of v4l2 output parsing variations
    # Implementation: Add optional mock interfaces for subprocess execution and parsing
    def _get_capability_probe_interface(self):
        """
        Test hook: Return interface for capability probing.
        
        This method enables test injection of mock v4l2-ctl behavior
        for comprehensive testing of parsing robustness.
        
        Returns:
            Dict with capability probing methods for test override
        """
        return {
            'probe_device_info': self._probe_device_info_robust,
            'probe_device_formats': self._probe_device_formats_robust,
            'probe_device_framerates': self._probe_device_framerates_robust,
            'extract_frame_rates': self._extract_frame_rates_from_output
        }

    # TODO: MEDIUM: Add udev event injection interface for testing [Story:E1/S3]
    # Purpose: Enable testing of udev event processing without real hardware
    # Implementation: Add method to inject synthetic udev events for testing
    async def _inject_test_udev_event(self, device_path: str, action: str) -> None:
        """
        Test hook: Inject synthetic udev event for testing.
        
        Args:
            device_path: Device path for synthetic event
            action: Udev action ('add', 'remove', 'change')
        """
        if not hasattr(self, '_test_mode'):
            self._logger.warning("Test event injection called outside test mode")
            return
            
        # Create mock udev device for testing
        class MockUdevDevice:
            def __init__(self, device_node: str, action: str):
                self.device_node = device_node
                self.action = action
        
        mock_device = MockUdevDevice(device_path, action)
        await self._process_udev_device_event(mock_device)