"""
Hybrid camera discovery monitor implementation.

Provides real-time USB camera detection using udev events with polling
fallback for reliability, as specified in the architecture design.
"""

import asyncio
import logging
import re
import time
import random
import hashlib
from abc import ABC, abstractmethod
from dataclasses import dataclass
from enum import Enum
from pathlib import Path
from typing import Callable, Dict, List, Optional, Set, Tuple, Any

from src.common.types import CameraDevice

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
    formats: Optional[List[Dict[str, Any]]] = None
    resolutions: Optional[List[str]] = None
    frame_rates: Optional[List[str]] = None
    error: Optional[str] = None
    timeout_context: Optional[str] = None
    probe_timestamp: float = 0.0
    structured_diagnostics: Optional[Dict[str, Any]] = None

    def __post_init__(self):
        if self.formats is None:
            self.formats = []
        if self.resolutions is None:
            self.resolutions = []
        if self.frame_rates is None:
            self.frame_rates = []
        if self.structured_diagnostics is None:
            self.structured_diagnostics = {}
        if self.probe_timestamp == 0.0:
            self.probe_timestamp = time.time()


@dataclass
class DeviceCapabilityState:
    """Tracks capability validation state for a device."""

    device_path: str
    provisional_data: Optional[CapabilityDetectionResult] = None
    confirmed_data: Optional[CapabilityDetectionResult] = None
    consecutive_successes: int = 0
    consecutive_failures: int = 0
    last_probe_time: float = 0.0
    confirmation_threshold: int = 2  # Require N consistent probes for confirmation
    validation_history: Optional[List[Dict[str, Any]]] = None

    # Enhanced frequency-based merge tracking
    format_frequency: Optional[Dict[str, int]] = None
    resolution_frequency: Optional[Dict[str, int]] = None
    frame_rate_frequency: Optional[Dict[str, int]] = None
    stability_threshold: int = 3  # Require N detections for stable capability

    def __post_init__(self):
        if self.validation_history is None:
            self.validation_history = []
        if self.format_frequency is None:
            self.format_frequency = {}
        if self.resolution_frequency is None:
            self.resolution_frequency = {}
        if self.frame_rate_frequency is None:
            self.frame_rate_frequency = {}

    def get_effective_capability(self) -> Optional[CapabilityDetectionResult]:
        """Get the capability data to use (confirmed or provisional)."""
        return self.confirmed_data if self.confirmed_data else self.provisional_data

    def is_confirmed(self) -> bool:
        """Check if current capability data is confirmed."""
        return self.confirmed_data is not None


class CameraEventHandler(ABC):
    """Abstract interface for camera event handling."""

    @abstractmethod
    async def handle_camera_event(self, event_data: CameraEventData) -> None:
        """
        Handle camera connection/disconnection events.

        Args:
            event_data: Event information including device path and type
        """


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
        device_range: Optional[List[int]] = None,
        poll_interval: float = 0.1,
        detection_timeout: float = 2.0,
        enable_capability_detection: bool = True,
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
        self._capability_states: Dict[str, DeviceCapabilityState] = {}
        self._monitoring_tasks: List[asyncio.Task] = []
        self._state_lock = asyncio.Lock()  # Protect against race conditions

        # Enhanced adaptive polling configuration
        self._base_poll_interval = poll_interval
        self._current_poll_interval = poll_interval
        self._min_poll_interval = max(0.05, poll_interval * 0.1)
        self._max_poll_interval = min(60.0, poll_interval * 50)
        self._last_udev_event_time = 0.0
        self._udev_event_freshness_threshold = 15.0  # seconds
        self._polling_failure_count = 0
        self._max_consecutive_failures = 5

        # Enhanced frame rate patterns for robust parsing
        self._frame_rate_patterns = [
            # Standard patterns (exclude negative numbers)
            r"(?<!-)(\d+(?:\.\d+)?)\s*fps\b",  # 30.000 fps (not -30 fps)
            r"(?<!-)(\d+(?:\.\d+)?)\s*FPS\b",  # 30.000 FPS (not -30 FPS)
            r"Frame\s*rate[:\s]+(?<!-)(\d+(?:\.\d+)?)",  # Frame rate: 30.0 (not -30.0)
            r"(?<!-)(\d+(?:\.\d+)?)\s*Hz\b",  # 30 Hz (not -30 Hz)
            r"@(?<!-)(\d+(?:\.\d+)?)\b",  # 1920x1080@60 (not @-60)
            # Interval patterns
            r"Interval:\s*\[1/(?<!-)(\d+(?:\.\d+)?)\]",  # Interval: [1/30] (not [1/-30])
            r"\[1/(?<!-)(\d+(?:\.\d+)?)\]",  # [1/30] (not [1/-30])
            r"1/(?<!-)(\d+(?:\.\d+)?)\s*s",  # 1/30 s (not 1/-30 s)
            # More complex patterns
            r"(?<!-)(\d+(?:\.\d+)?)\s*frame[s]?\s*per\s*second",  # 30 frames per second
            r"rate:\s*(?<!-)(\d+(?:\.\d+)?)",  # rate: 30 (not rate: -30)
            r"fps:\s*(?<!-)(\d+(?:\.\d+)?)",  # fps: 30 (not fps: -30)
        ]

        # Udev monitoring objects
        self._udev_context: Optional[pyudev.Context] = None
        self._udev_monitor: Optional[pyudev.Monitor] = None
        self._udev_available = HAS_PYUDEV

        # Enhanced diagnostic counters for observability
        self._stats = {
            "udev_events_processed": 0,
            "udev_events_filtered": 0,
            "udev_events_skipped": 0,
            "polling_cycles": 0,
            "capability_probes_attempted": 0,
            "capability_probes_successful": 0,
            "capability_probes_confirmed": 0,
            "capability_timeouts": 0,
            "capability_parse_errors": 0,
            "device_state_changes": 0,
            "adaptive_poll_adjustments": 0,
            "provisional_confirmations": 0,
            "confirmation_failures": 0,
            "current_poll_interval": poll_interval,
            "running": False,
            "active_tasks": 0,
        }

        # Initialize deterministic random for jitter
        self._rng = random.Random()
        self._rng.seed(hashlib.md5(str(id(self)).encode()).hexdigest())

        if not self._udev_available:
            self._logger.warning(
                "pyudev not available - falling back to polling-only monitoring",
                extra={"component": "hybrid_monitor", "mode": "polling_only"},
            )

        self._logger.debug(
            f"Initialized HybridCameraMonitor with device_range={self._device_range}, "
            f"poll_interval={self._poll_interval}s, udev_available={self._udev_available}",
            extra={
                "component": "hybrid_monitor",
                "device_range": self._device_range,
                "poll_interval": self._poll_interval,
                "udev_available": self._udev_available,
            },
        )

    def add_event_handler(self, handler: CameraEventHandler) -> None:
        """Add a camera event handler."""
        self._event_handlers.append(handler)
        self._logger.debug(f"Added event handler: {handler.__class__.__name__}")

    def add_event_callback(self, callback: Callable[[CameraEventData], None]) -> None:
        """Add a camera event callback function."""
        self._event_callbacks.append(callback)
        self._logger.debug(f"Added event callback: {callback.__name__}")

    @property
    def is_running(self) -> bool:
        """Check if the monitor is currently running."""
        return self._running

    def get_monitor_stats(self) -> Dict[str, Any]:
        """Get monitoring statistics and diagnostic information."""
        stats = self._stats.copy()
        stats["known_devices_count"] = len(self._known_devices)
        stats["capability_states_count"] = len(self._capability_states)
        stats["active_tasks"] = len([t for t in self._monitoring_tasks if not t.done()])
        stats["running"] = self._running
        return stats

    async def start(self) -> None:
        """Start the camera monitoring system."""
        if self._running:
            self._logger.warning("Monitor already running")
            return

        self._running = True
        self._stats["running"] = True
        self._logger.info(
            "Starting hybrid camera monitor", extra={"component": "hybrid_monitor"}
        )

        try:
            # Initialize udev monitoring if available
            if self._udev_available:
                await self._initialize_udev_monitoring()

            # Start polling fallback task
            polling_task = asyncio.create_task(self._adaptive_polling_loop())
            self._monitoring_tasks.append(polling_task)

            if self._udev_available and self._udev_monitor:
                udev_task = asyncio.create_task(self._udev_monitoring_loop())
                self._monitoring_tasks.append(udev_task)

            self._stats["active_tasks"] = len(self._monitoring_tasks)
            self._logger.info(
                f"Monitor started with {len(self._monitoring_tasks)} active tasks",
                extra={
                    "component": "hybrid_monitor",
                    "task_count": len(self._monitoring_tasks),
                },
            )

        except Exception as e:
            self._logger.error(f"Failed to start monitor: {e}", exc_info=True)
            self._running = False
            self._stats["running"] = False
            raise

    async def stop(self) -> None:
        """Stop the camera monitoring system."""
        if not self._running:
            return

        self._logger.info(
            "Stopping hybrid camera monitor", extra={"component": "hybrid_monitor"}
        )
        self._running = False
        self._stats["running"] = False

        # Cancel all monitoring tasks
        for task in self._monitoring_tasks:
            if not task.done():
                task.cancel()

        # Wait for tasks to complete
        if self._monitoring_tasks:
            await asyncio.gather(*self._monitoring_tasks, return_exceptions=True)

        self._monitoring_tasks.clear()

        # Cleanup udev resources
        if self._udev_monitor:
            self._udev_monitor = None
        if self._udev_context:
            self._udev_context = None

        self._stats["active_tasks"] = 0
        self._logger.info("Monitor stopped", extra={"component": "hybrid_monitor"})

    async def _initialize_udev_monitoring(self) -> None:
        """Initialize udev monitoring for real-time device events."""
        try:
            self._udev_context = pyudev.Context()
            self._udev_monitor = pyudev.Monitor.from_netlink(self._udev_context)
            self._udev_monitor.filter_by(subsystem="video4linux")
            self._udev_monitor.start()

            self._logger.info(
                "Udev monitoring initialized",
                extra={"component": "hybrid_monitor", "subsystem": "video4linux"},
            )
        except Exception as e:
            self._logger.error(
                f"Failed to initialize udev monitoring: {e}", exc_info=True
            )
            self._udev_available = False
            self._udev_monitor = None
            self._udev_context = None

    async def _udev_monitoring_loop(self) -> None:
        """Main udev event monitoring loop."""
        self._logger.debug("Starting udev monitoring loop")

        try:
            while self._running and self._udev_monitor:
                try:
                    # Poll for udev events with timeout
                    device = self._udev_monitor.poll(timeout=1.0)
                    if device:
                        await self._process_udev_device_event(device)

                except Exception as e:
                    self._logger.error(
                        f"Error in udev monitoring loop: {e}", exc_info=True
                    )
                    await asyncio.sleep(1.0)

        except asyncio.CancelledError:
            self._logger.debug("Udev monitoring loop cancelled")
        except Exception as e:
            self._logger.error(
                f"Critical error in udev monitoring loop: {e}", exc_info=True
            )

    async def _process_udev_device_event(self, device) -> None:
        """Process individual udev device events with enhanced filtering and diagnostics."""
        device_node = getattr(device, "device_node", None)
        action = getattr(device, "action", None)

        structured_event = {
            "device_node": device_node,
            "action": action,
            "timestamp": time.time(),
            "component": "hybrid_monitor",
            "event_type": "udev",
        }

        # Enhanced filtering with detailed logging
        if not device_node or not device_node.startswith("/dev/video"):
            self._stats["udev_events_filtered"] += 1
            self._logger.debug(
                f"Filtered udev event - invalid device node: {device_node}",
                extra=structured_event,
            )
            return

        # Extract device number for range validation
        device_match = re.search(r"/dev/video(\d+)", device_node)
        if not device_match:
            self._stats["udev_events_filtered"] += 1
            self._logger.debug(
                f"Filtered udev event - malformed device path: {device_node}",
                extra=structured_event,
            )
            return

        device_num = int(device_match.group(1))
        if device_num not in self._device_range:
            self._stats["udev_events_filtered"] += 1
            self._logger.debug(
                f"Filtered udev event - device {device_num} not in monitored range {self._device_range}",
                extra=structured_event,
            )
            return

        self._stats["udev_events_processed"] += 1
        self._last_udev_event_time = time.time()

        structured_event.update({"device_num": device_num, "processed": True})

        self._logger.info(
            f"Processing udev {action} event for {device_node}", extra=structured_event
        )

        try:
            # Process based on action with race condition protection
            async with self._state_lock:
                if action == "add":
                    await self._handle_udev_device_added(device_node, device_num)
                elif action == "remove":
                    await self._handle_udev_device_removed(device_node)
                elif action == "change":
                    await self._handle_udev_device_changed(device_node, device_num)
                else:
                    self._stats["udev_events_skipped"] += 1
                    self._logger.debug(
                        f"Skipped udev event with unknown action: {action}",
                        extra=structured_event,
                    )

        except Exception as e:
            self._logger.error(
                f"Error processing udev {action} event for {device_node}: {e}",
                extra=structured_event,
                exc_info=True,
            )

    async def _handle_udev_device_added(
        self, device_path: str, device_num: int
    ) -> None:
        """Handle udev 'add' event for device connection."""
        # Verify device is actually accessible before creating event
        device_info = await self._create_camera_device_info(device_path, device_num)

        if device_info and device_info.status == "CONNECTED":
            event_data = CameraEventData(
                device_path=device_path,
                event_type=CameraEvent.CONNECTED,
                device_info=device_info,
                timestamp=time.time(),
            )

            # Update known devices and handle event
            self._known_devices[device_path] = device_info
            self._stats["device_state_changes"] += 1
            await self._handle_camera_event(event_data)

            self._logger.info(
                f"Device {device_path} connected via udev event",
                extra={
                    "device_path": device_path,
                    "device_num": device_num,
                    "event_source": "udev_add",
                },
            )
        else:
            self._logger.warning(
                f"Device {device_path} detected via udev 'add' but not accessible",
                extra={
                    "device_path": device_path,
                    "device_num": device_num,
                    "accessibility_check": "failed",
                },
            )

    async def _handle_udev_device_removed(self, device_path: str) -> None:
        """Handle udev 'remove' event for device disconnection."""
        device_info = self._known_devices.get(device_path)
        event_data = CameraEventData(
            device_path=device_path,
            event_type=CameraEvent.DISCONNECTED,
            device_info=device_info,
            timestamp=time.time(),
        )

        # Remove from known devices and capability states
        had_device = device_path in self._known_devices
        if had_device:
            del self._known_devices[device_path]
            self._stats["device_state_changes"] += 1

        if device_path in self._capability_states:
            del self._capability_states[device_path]

        await self._handle_camera_event(event_data)

        self._logger.info(
            f"Device {device_path} disconnected via udev event",
            extra={
                "device_path": device_path,
                "was_known": had_device,
                "event_source": "udev_remove",
            },
        )

    async def _handle_udev_device_changed(
        self, device_path: str, device_num: int
    ) -> None:
        """Handle udev 'change' event for device property changes."""
        device_info = await self._create_camera_device_info(device_path, device_num)
        old_device_info = self._known_devices.get(device_path)

        # Only generate event if status actually changed
        if (
            device_info
            and old_device_info
            and device_info.status != old_device_info.status
        ):
            event_data = CameraEventData(
                device_path=device_path,
                event_type=CameraEvent.STATUS_CHANGED,
                device_info=device_info,
                timestamp=time.time(),
            )

            # Update known devices and handle event
            self._known_devices[device_path] = device_info
            self._stats["device_state_changes"] += 1
            await self._handle_camera_event(event_data)

            self._logger.info(
                f"Device {device_path} status changed: {old_device_info.status} → {device_info.status}",
                extra={
                    "device_path": device_path,
                    "old_status": old_device_info.status,
                    "new_status": device_info.status,
                    "event_source": "udev_change",
                },
            )
        else:
            self._logger.debug(
                f"No significant status change for {device_path}",
                extra={
                    "device_path": device_path,
                    "event_source": "udev_change",
                    "status_check": "no_change",
                },
            )

    async def _create_camera_device_info(
        self, device_path: str, device_num: int
    ) -> CameraDevice:
        """Create CameraDevice info for a detected device."""
        # Basic accessibility check
        try:
            path_obj = Path(device_path)
            if not path_obj.exists():
                return CameraDevice(
                    device=device_path,
                    name=f"Camera {device_num}",
                    status="DISCONNECTED",
                )
        except Exception as e:
            self._logger.debug(f"Path check failed for {device_path}: {e}")
            return CameraDevice(
                device=device_path, name=f"Camera {device_num}", status="ERROR"
            )

        # Determine status based on basic accessibility
        try:
            # Try to open device for basic validation
            with open(device_path, "rb"):
                status = "CONNECTED"
        except (OSError, PermissionError):
            status = "DISCONNECTED"
        except Exception:
            status = "ERROR"

        try:
            device_info = CameraDevice(
                device=device_path, name=f"Camera {device_num}", status=status
            )
            return device_info
        except Exception as e:
            self._logger.error(
                f"Error creating camera device info for {device_path}: {e}",
                extra={"device_path": device_path, "device_num": device_num},
            )
            return CameraDevice(
                device=device_path, name=f"Camera {device_num}", status="ERROR"
            )

        # Trigger capability detection if enabled and device is accessible
        if self._enable_capability_detection and status == "CONNECTED":
            try:
                await self._probe_device_capabilities(device_path)
            except Exception as e:
                self._logger.debug(f"Capability probing failed for {device_path}: {e}")

        return device_info

    async def _adaptive_polling_loop(self) -> None:
        """
        Adaptive polling fallback loop with enhanced reliability and diagnostics.

        Adapts polling frequency based on udev event reliability:
        - Reduces frequency when udev events are working reliably
        - Increases frequency when events are missed or stale
        - Exponential backoff on consecutive failures with jitter
        """
        self._logger.debug("Starting adaptive polling fallback loop")

        polling_error_count = 0

        try:
            while self._running:
                loop_start = time.time()

                try:
                    # Perform discovery
                    await self._discover_cameras()
                    self._stats["polling_cycles"] += 1
                    polling_error_count = 0  # Reset on success
                    self._polling_failure_count = 0

                    # Adaptive polling interval adjustment
                    await self._adjust_polling_interval()

                    # Sleep with consideration for loop execution time
                    loop_duration = time.time() - loop_start
                    sleep_time = max(0, self._current_poll_interval - loop_duration)

                    if sleep_time > 0:
                        await asyncio.sleep(sleep_time)

                except Exception as e:
                    polling_error_count += 1
                    self._polling_failure_count += 1

                    structured_error = {
                        "component": "hybrid_monitor",
                        "error_type": "polling_discovery",
                        "error_count": polling_error_count,
                        "consecutive_failures": self._polling_failure_count,
                    }

                    self._logger.error(
                        f"Polling discovery error (#{polling_error_count}): {e}",
                        extra=structured_error,
                        exc_info=True,
                    )

                    if polling_error_count >= self._max_consecutive_failures:
                        self._logger.critical(
                            f"Too many consecutive polling errors ({polling_error_count}), "
                            "stopping polling loop",
                            extra=structured_error,
                        )
                        break

                    # Enhanced exponential backoff with jitter
                    base_backoff = min(
                        self._base_poll_interval * (2**self._polling_failure_count),
                        self._max_poll_interval,
                    )
                    jitter = self._rng.uniform(0.8, 1.2)  # ±20% jitter
                    backoff_interval = base_backoff * jitter

                    self._logger.debug(
                        f"Polling backoff: {backoff_interval:.2f}s (base: {base_backoff:.2f}s, jitter: {jitter:.2f})",
                        extra={
                            "component": "hybrid_monitor",
                            "backoff_interval": backoff_interval,
                            "base_backoff": base_backoff,
                            "jitter_factor": jitter,
                        },
                    )
                    await asyncio.sleep(backoff_interval)

        except asyncio.CancelledError:
            self._logger.debug("Polling loop cancelled")
        except Exception as e:
            self._logger.error(f"Critical error in polling loop: {e}", exc_info=True)

    async def _adjust_polling_interval(self) -> None:
        """
        Adjust polling interval based on udev event reliability with enhanced logic.

        Factors considered:
        - Time since last udev event
        - Recent polling failures
        - Overall system responsiveness
        """
        current_time = time.time()
        time_since_udev = current_time - self._last_udev_event_time

        old_interval = self._current_poll_interval

        # Determine new interval based on udev event freshness
        if time_since_udev > self._udev_event_freshness_threshold:
            # No recent udev events - increase polling frequency
            self._current_poll_interval = max(
                self._min_poll_interval, self._current_poll_interval * 0.8
            )
        elif time_since_udev < self._udev_event_freshness_threshold / 2:
            # Recent udev events - can reduce polling frequency
            self._current_poll_interval = min(
                self._max_poll_interval, self._current_poll_interval * 1.2
            )

        # Factor in recent failures
        if self._polling_failure_count > 0:
            failure_penalty = 1.0 + (self._polling_failure_count * 0.1)
            self._current_poll_interval = min(
                self._max_poll_interval, self._current_poll_interval * failure_penalty
            )

        # Update stats if interval changed significantly
        if abs(self._current_poll_interval - old_interval) > 0.01:
            self._stats["adaptive_poll_adjustments"] += 1
            self._stats["current_poll_interval"] = self._current_poll_interval

            self._logger.debug(
                f"Adjusted polling interval: {old_interval:.2f}s → {self._current_poll_interval:.2f}s "
                f"(udev_age: {time_since_udev:.1f}s, failures: {self._polling_failure_count})",
                extra={
                    "component": "hybrid_monitor",
                    "old_interval": old_interval,
                    "new_interval": self._current_poll_interval,
                    "udev_age": time_since_udev,
                    "failure_count": self._polling_failure_count,
                    "adjustment_reason": "adaptive_tuning",
                },
            )

    async def _discover_cameras(self) -> None:
        """Discover currently connected cameras via polling."""
        current_devices = {}

        for device_num in self._device_range:
            device_path = f"/dev/video{device_num}"

            try:
                device_info = await self._create_camera_device_info(
                    device_path, device_num
                )
                if device_info and device_info.status in ["CONNECTED", "ERROR"]:
                    current_devices[device_path] = device_info
            except Exception as e:
                self._logger.debug(f"Error checking device {device_path}: {e}")
                continue

        # Compare with known devices and generate events
        await self._process_device_state_changes(current_devices)

    async def _process_device_state_changes(
        self, current_devices: Dict[str, CameraDevice]
    ) -> None:
        """Process changes in device state between polling cycles."""
        async with self._state_lock:
            # Detect new devices
            for device_path, device_info in current_devices.items():
                if device_path not in self._known_devices:
                    await self._handle_camera_event(
                        CameraEventData(
                            device_path=device_path,
                            event_type=CameraEvent.CONNECTED,
                            device_info=device_info,
                            timestamp=time.time(),
                        )
                    )
                    self._stats["device_state_changes"] += 1

            # Detect removed devices
            for device_path in list(self._known_devices.keys()):
                if device_path not in current_devices:
                    await self._handle_camera_event(
                        CameraEventData(
                            device_path=device_path,
                            event_type=CameraEvent.DISCONNECTED,
                            device_info=self._known_devices[device_path],
                            timestamp=time.time(),
                        )
                    )
                    self._stats["device_state_changes"] += 1

            # Detect status changes for existing devices
            for device_path, device_info in current_devices.items():
                if device_path in self._known_devices:
                    if self._known_devices[device_path].status != device_info.status:
                        await self._handle_camera_event(
                            CameraEventData(
                                device_path=device_path,
                                event_type=CameraEvent.STATUS_CHANGED,
                                device_info=device_info,
                                timestamp=time.time(),
                            )
                        )
                        self._stats["device_state_changes"] += 1

            # Update known devices
            self._known_devices = current_devices.copy()

    async def _handle_camera_event(self, event_data: CameraEventData) -> None:
        """Handle camera events by notifying all registered handlers and callbacks."""
        try:
            # Call all event handlers
            for handler in self._event_handlers:
                try:
                    await handler.handle_camera_event(event_data)
                except Exception as e:
                    self._logger.error(
                        f"Error in event handler {handler.__class__.__name__}: {e}",
                        exc_info=True,
                    )

            # Call all event callbacks
            for callback in self._event_callbacks:
                try:
                    callback(event_data)
                except Exception as e:
                    self._logger.error(
                        f"Error in event callback {callback.__name__}: {e}",
                        exc_info=True,
                    )

        except Exception as e:
            self._logger.error(
                f"Critical error handling camera event: {e}", exc_info=True
            )

    async def _probe_device_capabilities(
        self, device_path: str
    ) -> CapabilityDetectionResult:
        """
        Probe device capabilities with enhanced error handling and structured diagnostics.

        Args:
            device_path: Path to video device

        Returns:
            Structured capability detection result with enhanced diagnostics
        """
        self._stats["capability_probes_attempted"] += 1
        probe_start = time.time()

        result = CapabilityDetectionResult(
            device_path=device_path,
            detected=False,
            accessible=False,
            probe_timestamp=probe_start,
            structured_diagnostics={
                "probe_start": probe_start,
                "timeout_threshold": self._detection_timeout,
                "parsing_stages": [],
            },
        )

        try:
            # Basic device info probe
            device_info = await self._probe_device_info_robust(device_path)
            if device_info:
                result.device_name = device_info.get("name")
                result.driver = device_info.get("driver")
                result.accessible = True
                result.structured_diagnostics["device_info_success"] = True
            else:
                result.error = "Failed to probe basic device information (timeout or device unavailable)"
                result.timeout_context = "device_info_probe_timeout"
                result.structured_diagnostics["device_info_success"] = False
                return result

            # Format and resolution probe
            formats_data = await self._probe_device_formats_robust(device_path)
            if formats_data:
                result.formats = formats_data.get("formats", [])
                result.resolutions = formats_data.get("resolutions", [])
                result.structured_diagnostics["formats_found"] = len(result.formats)
                result.structured_diagnostics["resolutions_found"] = len(
                    result.resolutions
                )

            # Frame rate probe with hierarchical selection
            frame_rates = await self._probe_device_framerates_robust(device_path)
            if frame_rates:
                result.frame_rates = frame_rates
                result.structured_diagnostics["frame_rates_found"] = len(frame_rates)

            # Consider detection successful if we got basic info
            if result.accessible and (
                result.formats or result.resolutions or result.frame_rates
            ):
                result.detected = True
                self._stats["capability_probes_successful"] += 1

                # Update capability validation state
                await self._update_capability_validation_state(device_path, result)
            else:
                result.error = "Insufficient capability data detected"

        except asyncio.TimeoutError:
            self._stats["capability_timeouts"] += 1
            result.error = (
                f"Capability detection timeout after {self._detection_timeout}s"
            )
            result.timeout_context = "overall_probe_timeout"
            self._logger.warning(
                f"Capability detection timeout for {device_path}",
                extra={
                    "device_path": device_path,
                    "timeout_duration": self._detection_timeout,
                    "component": "capability_detection",
                },
            )
        except Exception as e:
            self._stats["capability_parse_errors"] += 1
            result.error = f"Capability detection error: {str(e)}"
            result.structured_diagnostics["exception_type"] = type(e).__name__
            self._logger.error(
                f"Error probing capabilities for {device_path}: {e}",
                extra={"device_path": device_path, "component": "capability_detection"},
                exc_info=True,
            )

        probe_duration = time.time() - probe_start
        result.structured_diagnostics["probe_duration"] = probe_duration

        return result

    async def _probe_device_info_robust(
        self, device_path: str
    ) -> Optional[Dict[str, str]]:
        """Probe basic device information with enhanced error handling."""
        try:
            process = await asyncio.wait_for(
                asyncio.create_subprocess_exec(
                    "v4l2-ctl",
                    "--device",
                    device_path,
                    "--info",
                    stdout=asyncio.subprocess.PIPE,
                    stderr=asyncio.subprocess.PIPE,
                ),
                timeout=self._detection_timeout,
            )

            stdout, stderr = await process.communicate()

            if process.returncode != 0:
                return None

            output = stdout.decode()
            info = {}

            # Enhanced parsing with multiple patterns
            name_patterns = [
                r"Card type\s*:\s*(.+)",
                r"Device name\s*:\s*(.+)",
                r"Card\s*:\s*(.+)",
            ]

            driver_patterns = [r"Driver name\s*:\s*(.+)", r"Driver\s*:\s*(.+)"]

            for pattern in name_patterns:
                match = re.search(pattern, output, re.IGNORECASE)
                if match:
                    info["name"] = match.group(1).strip()
                    break

            for pattern in driver_patterns:
                match = re.search(pattern, output, re.IGNORECASE)
                if match:
                    info["driver"] = match.group(1).strip()
                    break

            return info if info else None

        except asyncio.TimeoutError:
            self._logger.debug(f"Timeout probing device info for {device_path}")
            return None
        except Exception as e:
            self._logger.debug(f"Error probing device info for {device_path}: {e}")
            return None

    def _extract_frame_rates_from_output(self, output: str) -> Set[str]:
        """
        Extract frame rates from v4l2-ctl output using enhanced parsing strategies.

        Args:
            output: Raw v4l2-ctl command output

        Returns:
            Set of frame rate strings
        """
        if not output or not output.strip():
            return set()

        frame_rates = set()

        # Apply all frame rate patterns
        for pattern in self._frame_rate_patterns:
            try:
                matches = re.findall(pattern, output, re.IGNORECASE)
                for match in matches:
                    # Convert to float and back to string to normalize
                    try:
                        rate_val = float(match)
                        if (
                            1 <= rate_val <= 300
                        ):  # Extended frame rate range for high-end cameras
                            # Store as integer if it's a whole number
                            if rate_val == int(rate_val):
                                frame_rates.add(str(int(rate_val)))
                            else:
                                frame_rates.add(f"{rate_val:.1f}")
                    except (ValueError, TypeError):
                        continue
            except re.error:
                # Skip malformed patterns
                continue

        return frame_rates

    def _extract_resolutions_from_output(self, output: str) -> List[str]:
        """
        Extract resolutions from v4l2-ctl output.

        Args:
            output: Raw v4l2-ctl command output

        Returns:
            List of resolution strings
        """
        if not output or not output.strip():
            return []

        resolutions = set()
        # Pattern to match "Size: Discrete 1920x1080" format and variations
        patterns = [
            r"Size:\s*Discrete\s+(\d+x\d+)",  # Standard format
            r"Resolution:\s*(\d+[×x]\d+)",  # Resolution format with Unicode × or x
            r"(\d+\s*[×x]\s*\d+)",  # General format with spaces
            r"(\d+\*\d+)",  # Asterisk separator
            r"(\d+\s*x\s*\d+)",  # Space-separated format
        ]

        for pattern in patterns:
            matches = re.findall(pattern, output, re.IGNORECASE)
            for match in matches:
                # Normalize to standard format (replace × and * with x, remove spaces around x)
                normalized = re.sub(r"[×*]", "x", match.strip())
                normalized = re.sub(r"\s*x\s*", "x", normalized)
                # Validate resolution format
                if re.match(r"\d+x\d+", normalized):
                    resolutions.add(normalized)

        return list(resolutions)

    def _extract_formats_from_output(self, output: str) -> List[Dict[str, str]]:
        """
        Extract pixel formats from v4l2-ctl output.

        Args:
            output: Raw v4l2-ctl command output

        Returns:
            List of format dictionaries
        """
        if not output or not output.strip():
            return []

        formats = []
        # Pattern to match "Pixel Format: 'YUYV' (YUYV 4:2:2)" format
        pattern = r"Pixel Format:\s*'([^']+)'\s*(?:\(([^)]+)\))?"
        matches = re.findall(pattern, output, re.IGNORECASE)

        # Format descriptions mapping
        format_descriptions = {
            "YUYV": "YUYV 4:2:2",
            "MJPG": "Motion-JPEG",
            "RGB24": "RGB 24-bit",
            "BGR24": "BGR 24-bit",
            "NV12": "NV12 YUV",
            "NV21": "NV21 YUV",
            "YV12": "YV12 YUV",
            "YU12": "YU12 YUV",
        }

        for match in matches:
            format_code = match[0]
            description = (
                match[1]
                if match[1]
                else format_descriptions.get(format_code, f"{format_code} format")
            )
            formats.append({"format": format_code, "description": description})

        return formats

    def _get_or_create_capability_state(
        self, device_path: str
    ) -> DeviceCapabilityState:
        """
        Get or create capability state for a device.

        Args:
            device_path: Device path

        Returns:
            DeviceCapabilityState instance
        """
        if device_path not in self._capability_states:
            self._capability_states[device_path] = DeviceCapabilityState(
                device_path=device_path
            )
        return self._capability_states[device_path]

    async def _update_capability_state(
        self, device_path: str, result: CapabilityDetectionResult
    ) -> None:
        """
        Update capability state for a device.

        Args:
            device_path: Device path
            result: Capability detection result
        """
        await self._update_capability_validation_state(device_path, result)

    async def _probe_device_framerates_robust(
        self, device_path: str
    ) -> Optional[List[str]]:
        """
        Probe supported frame rates with hierarchical selection and robust parsing.

        Enhanced frame rate detection with multiple strategies and preference ordering.
        """
        all_frame_rates = set()
        detection_sources = []

        # Enhanced command list with more comprehensive probing
        commands_to_try = [
            (
                ["v4l2-ctl", "--device", device_path, "--list-framesizes", "YUYV"],
                "YUYV framesizes",
            ),
            (
                ["v4l2-ctl", "--device", device_path, "--list-framesizes", "MJPG"],
                "MJPG framesizes",
            ),
            (
                ["v4l2-ctl", "--device", device_path, "--list-framesizes", "RGB24"],
                "RGB24 framesizes",
            ),
            (
                ["v4l2-ctl", "--device", device_path, "--list-framerates"],
                "general framerates",
            ),
            (
                ["v4l2-ctl", "--device", device_path, "--list-formats-ext"],
                "extended formats",
            ),
            (["v4l2-ctl", "--device", device_path, "--all"], "all device info"),
        ]

        for cmd, description in commands_to_try:
            try:
                process = await asyncio.wait_for(
                    asyncio.create_subprocess_exec(
                        *cmd,
                        stdout=asyncio.subprocess.PIPE,
                        stderr=asyncio.subprocess.PIPE,
                    ),
                    timeout=self._detection_timeout,
                )

                stdout, stderr = await process.communicate()

                if process.returncode == 0:
                    output = stdout.decode()
                    cmd_frame_rates = self._extract_frame_rates_from_output(output)

                    if cmd_frame_rates:
                        all_frame_rates.update(cmd_frame_rates)
                        detection_sources.append((description, cmd_frame_rates))
                        self._logger.debug(
                            f"Found frame rates from {description}: {sorted(cmd_frame_rates)}",
                            extra={
                                "device_path": device_path,
                                "detection_source": description,
                                "frame_rates": sorted(cmd_frame_rates),
                            },
                        )

            except asyncio.TimeoutError:
                self._logger.debug(f"Timeout getting {description} for {device_path}")
                continue
            except Exception as e:
                self._logger.debug(
                    f"Error getting {description} for {device_path}: {e}"
                )
                continue

        # Apply hierarchical frame rate selection
        if all_frame_rates:
            return self._select_preferred_frame_rates(
                all_frame_rates, detection_sources, device_path
            )
        else:
            # Return common default frame rates if detection fails
            default_rates = ["30", "25", "24", "15", "10", "5"]
            self._logger.debug(
                f"No frame rates detected for {device_path}, using defaults: {default_rates}",
                extra={
                    "device_path": device_path,
                    "fallback_used": True,
                    "default_rates": default_rates,
                },
            )
            return default_rates

    def _select_preferred_frame_rates(
        self, all_rates: Set[str], sources: List[Tuple[str, Set[str]]], device_path: str
    ) -> List[str]:
        """
        Select preferred frame rates using enhanced hierarchical policy.

        Enhanced policy:
        1. Highest stable frame rate preferred for given resolution
        2. Common frame rates (30, 25, 24, 15) prioritized
        3. Rates detected by multiple sources weighted higher
        4. Consistent ordering for deterministic behavior
        """

        # Define preference tiers
        high_priority_rates = {"30", "25", "24"}
        medium_priority_rates = {"15", "60", "10"}

        # Count detection frequency (reliability indicator)
        rate_frequency: Dict[str, int] = {}
        for source_name, source_rates in sources:
            for rate in source_rates:
                rate_frequency[rate] = rate_frequency.get(rate, 0) + 1

        # Sort by multiple criteria
        def rate_sort_key(rate: str) -> Tuple[int, int, float]:
            try:
                rate_val = float(rate)

                # Priority tier (lower is better)
                if rate in high_priority_rates:
                    priority = 0
                elif rate in medium_priority_rates:
                    priority = 1
                else:
                    priority = 2

                # Detection frequency (higher is better, so negate)
                frequency = -rate_frequency.get(rate, 1)

                # Rate value for final ordering (higher is better, so negate)
                rate_value = -rate_val

                return (priority, frequency, rate_value)
            except (ValueError, TypeError):
                # Invalid rates go to the end
                return (999, 0, 0)

        sorted_rates = sorted(all_rates, key=rate_sort_key)

        self._logger.debug(
            f"Selected frame rate order for {device_path}: {sorted_rates}",
            extra={
                "device_path": device_path,
                "rate_frequency": rate_frequency,
                "final_order": sorted_rates,
                "selection_criteria": "hierarchical_policy",
            },
        )

        return sorted_rates

    async def _update_capability_validation_state(
        self, device_path: str, new_result: CapabilityDetectionResult
    ) -> None:
        """
        Update capability validation state with enhanced frequency-based merge logic.

        Uses weighted merge based on detection frequency with stability thresholds:
        - Tracks frequency of each capability element across probes
        - Promotes capabilities that meet stability threshold to stable set
        - Filters out one-off detections that may be transient
        - Prevents oscillation through frequency-based consistency validation

        Args:
            device_path: Device path
            new_result: New capability detection result
        """
        if device_path not in self._capability_states:
            self._capability_states[device_path] = DeviceCapabilityState(
                device_path=device_path
            )

        state = self._capability_states[device_path]
        state.last_probe_time = time.time()

        # Add to validation history
        history_entry = {
            "timestamp": state.last_probe_time,
            "detected": new_result.detected,
            "error": new_result.error,
            "formats_count": len(new_result.formats),
            "resolutions_count": len(new_result.resolutions),
            "frame_rates_count": len(new_result.frame_rates),
        }
        state.validation_history.append(history_entry)

        # Keep history manageable
        if len(state.validation_history) > 10:
            state.validation_history = state.validation_history[-10:]

        if new_result.detected:
            # Update frequency tracking for all capability elements
            self._update_capability_frequencies(state, new_result)

            # Generate frequency-based merged capability data
            merged_result = self._create_frequency_merged_capability(state, new_result)

            # Check stability-aware consistency with existing data
            if state.provisional_data and self._is_frequency_based_consistent(
                state, merged_result
            ):
                state.consecutive_successes += 1
                state.consecutive_failures = 0

                # Promote to confirmed if threshold met
                if state.consecutive_successes >= state.confirmation_threshold:
                    if not state.confirmed_data:
                        self._stats["provisional_confirmations"] += 1
                        stable_formats = len(
                            [
                                f
                                for f, freq in state.format_frequency.items()
                                if freq >= state.stability_threshold
                            ]
                        )
                        stable_resolutions = len(
                            [
                                r
                                for r, freq in state.resolution_frequency.items()
                                if freq >= state.stability_threshold
                            ]
                        )
                        stable_rates = len(
                            [
                                r
                                for r, freq in state.frame_rate_frequency.items()
                                if freq >= state.stability_threshold
                            ]
                        )

                        self._logger.info(
                            f"Capability data confirmed for {device_path} after {state.consecutive_successes} consistent probes "
                            f"(stable: {stable_formats} formats, {stable_resolutions} resolutions, {stable_rates} rates)",
                            extra={
                                "device_path": device_path,
                                "consecutive_successes": state.consecutive_successes,
                                "stable_capabilities": {
                                    "formats": stable_formats,
                                    "resolutions": stable_resolutions,
                                    "frame_rates": stable_rates,
                                },
                                "validation_transition": "provisional_to_confirmed",
                                "merge_strategy": "frequency_based",
                            },
                        )

                    state.confirmed_data = merged_result
                    self._stats["capability_probes_confirmed"] += 1
                else:
                    self._logger.debug(
                        f"Capability data frequency-consistency maintained for {device_path} "
                        f"({state.consecutive_successes}/{state.confirmation_threshold})",
                        extra={
                            "device_path": device_path,
                            "progress": f"{state.consecutive_successes}/{state.confirmation_threshold}",
                            "validation_status": "provisional_consistent",
                            "merge_strategy": "frequency_based",
                        },
                    )
            else:
                # Inconsistent data - check if it's minor variance or major change
                variance_score = self._calculate_capability_variance(
                    state, merged_result
                )

                if (
                    variance_score < 0.2
                ):  # Minor variance - continue with frequency merge
                    state.consecutive_successes += 1
                    state.consecutive_failures = 0
                    self._logger.debug(
                        f"Minor capability variance for {device_path} (score: {variance_score:.2f}), "
                        f"continuing frequency-based merge",
                        extra={
                            "device_path": device_path,
                            "variance_score": variance_score,
                            "validation_action": "continue_with_variance",
                        },
                    )
                else:
                    # Major variance - reset validation but preserve frequency data
                    if state.provisional_data:
                        self._logger.warning(
                            f"Major capability variance detected for {device_path} (score: {variance_score:.2f}), "
                            f"resetting validation but preserving frequency data",
                            extra={
                                "device_path": device_path,
                                "variance_score": variance_score,
                                "validation_action": "reset_with_frequency_preservation",
                            },
                        )

                    state.consecutive_successes = 1
                    state.consecutive_failures = 0
                    state.confirmed_data = None

            # For the first detection, preserve the original data
            if not state.provisional_data:
                state.provisional_data = new_result
            else:
                state.provisional_data = merged_result
        else:
            # Detection failed
            state.consecutive_failures += 1
            state.consecutive_successes = 0

            if state.consecutive_failures >= 3:
                self._stats["confirmation_failures"] += 1
                self._logger.warning(
                    f"Capability detection failing consistently for {device_path} "
                    f"({state.consecutive_failures} failures)",
                    extra={
                        "device_path": device_path,
                        "consecutive_failures": state.consecutive_failures,
                        "validation_status": "persistent_failure",
                    },
                )

    def _update_capability_frequencies(
        self, state: DeviceCapabilityState, result: CapabilityDetectionResult
    ) -> None:
        """Update frequency counters for capability elements."""

        # Update format frequencies
        for fmt in result.formats:
            fmt_code = fmt.get("code", "") if isinstance(fmt, dict) else str(fmt)
            if fmt_code:
                state.format_frequency[fmt_code] = (
                    state.format_frequency.get(fmt_code, 0) + 1
                )

        # Update resolution frequencies
        for resolution in result.resolutions:
            if resolution:
                state.resolution_frequency[resolution] = (
                    state.resolution_frequency.get(resolution, 0) + 1
                )

        # Update frame rate frequencies
        for rate in result.frame_rates:
            if rate:
                state.frame_rate_frequency[rate] = (
                    state.frame_rate_frequency.get(rate, 0) + 1
                )

    def _create_frequency_merged_capability(
        self, state: DeviceCapabilityState, latest_result: CapabilityDetectionResult
    ) -> CapabilityDetectionResult:
        """
        Create merged capability result based on frequency analysis.

        Includes capabilities that meet stability threshold and weights by frequency.
        """

        # Filter capabilities by stability threshold and sort by frequency
        stable_formats = [
            {"code": fmt, "description": f"Format {fmt}"}
            for fmt, freq in sorted(
                state.format_frequency.items(), key=lambda x: x[1], reverse=True
            )
            if freq >= state.stability_threshold
        ]

        stable_resolutions = [
            resolution
            for resolution, freq in sorted(
                state.resolution_frequency.items(), key=lambda x: x[1], reverse=True
            )
            if freq >= state.stability_threshold
        ]

        stable_frame_rates = [
            rate
            for rate, freq in sorted(
                state.frame_rate_frequency.items(), key=lambda x: x[1], reverse=True
            )
            if freq >= state.stability_threshold
        ]

        # Include recent detections if they don't conflict with stable set
        recent_formats = []
        for fmt in latest_result.formats:
            fmt_code = fmt.get("code", "") if isinstance(fmt, dict) else str(fmt)
            if fmt_code and fmt_code not in [sf["code"] for sf in stable_formats]:
                # Add if it's been seen at least once before or has high confidence
                if state.format_frequency.get(fmt_code, 0) > 0:
                    recent_formats.append(fmt)

        recent_resolutions = [
            res
            for res in latest_result.resolutions
            if res
            and res not in stable_resolutions
            and state.resolution_frequency.get(res, 0) > 0
        ]

        recent_frame_rates = [
            rate
            for rate in latest_result.frame_rates
            if rate
            and rate not in stable_frame_rates
            and state.frame_rate_frequency.get(rate, 0) > 0
        ]

        # Create merged result
        merged_result = CapabilityDetectionResult(
            device_path=latest_result.device_path,
            detected=True,
            accessible=latest_result.accessible,
            device_name=latest_result.device_name,
            driver=latest_result.driver,
            formats=stable_formats + recent_formats,
            resolutions=stable_resolutions + recent_resolutions,
            frame_rates=stable_frame_rates + recent_frame_rates,
            probe_timestamp=latest_result.probe_timestamp,
            structured_diagnostics={
                **latest_result.structured_diagnostics,
                "merge_strategy": "frequency_based",
                "stable_elements": {
                    "formats": len(stable_formats),
                    "resolutions": len(stable_resolutions),
                    "frame_rates": len(stable_frame_rates),
                },
                "recent_elements": {
                    "formats": len(recent_formats),
                    "resolutions": len(recent_resolutions),
                    "frame_rates": len(recent_frame_rates),
                },
                "frequency_data": {
                    "format_frequencies": dict(state.format_frequency),
                    "resolution_frequencies": dict(state.resolution_frequency),
                    "frame_rate_frequencies": dict(state.frame_rate_frequency),
                },
            },
        )

        return merged_result

    def _is_frequency_based_consistent(
        self, state: DeviceCapabilityState, new_result: CapabilityDetectionResult
    ) -> bool:
        """
        Check consistency using frequency-based stability analysis.

        More lenient than intersection-based consistency - allows for capability
        variance as long as core stable capabilities remain consistent.
        """
        if not state.provisional_data or not new_result.detected:
            return False

        # Check that stable capabilities (high frequency) remain present
        stable_formats = {
            fmt
            for fmt, freq in state.format_frequency.items()
            if freq >= state.stability_threshold
        }
        stable_resolutions = {
            res
            for res, freq in state.resolution_frequency.items()
            if freq >= state.stability_threshold
        }
        stable_frame_rates = {
            rate
            for rate, freq in state.frame_rate_frequency.items()
            if freq >= state.stability_threshold
        }

        # Get current result capability sets
        current_formats = {
            fmt.get("code", "") for fmt in new_result.formats if isinstance(fmt, dict)
        }
        current_resolutions = set(new_result.resolutions)
        current_frame_rates = set(new_result.frame_rates)

        # Calculate consistency scores for each capability type
        format_consistency = self._calculate_set_consistency(
            stable_formats, current_formats
        )
        resolution_consistency = self._calculate_set_consistency(
            stable_resolutions, current_resolutions
        )
        rate_consistency = self._calculate_set_consistency(
            stable_frame_rates, current_frame_rates
        )

        # Require high consistency for stable capabilities
        min_consistency = 0.7  # 70% of stable capabilities should be present

        overall_consistent = (
            format_consistency >= min_consistency
            and resolution_consistency >= min_consistency
            and rate_consistency >= min_consistency
        )

        return overall_consistent

    def _calculate_set_consistency(
        self, stable_set: Set[str], current_set: Set[str]
    ) -> float:
        """Calculate consistency score between stable and current capability sets."""
        if not stable_set:
            return 1.0  # No stable capabilities to validate against

        intersection = stable_set.intersection(current_set)
        return len(intersection) / len(stable_set)

    def _calculate_capability_variance(
        self, state: DeviceCapabilityState, new_result: CapabilityDetectionResult
    ) -> float:
        """
        Calculate variance score between frequency-merged capabilities and new result.

        Returns:
            float: Variance score (0.0 = identical, 1.0 = completely different)
        """
        if not state.provisional_data:
            return 0.0

        prev_result = state.provisional_data

        # Calculate variance for each capability type
        format_variance = self._calculate_list_variance(
            [f.get("code", "") for f in prev_result.formats],
            [f.get("code", "") for f in new_result.formats],
        )

        resolution_variance = self._calculate_list_variance(
            prev_result.resolutions, new_result.resolutions
        )

        rate_variance = self._calculate_list_variance(
            prev_result.frame_rates, new_result.frame_rates
        )

        # Weight variance by importance (resolutions and rates more critical than formats)
        weighted_variance = (
            format_variance * 0.2 + resolution_variance * 0.4 + rate_variance * 0.4
        )

        return weighted_variance

    def _calculate_list_variance(self, list1: List[str], list2: List[str]) -> float:
        """Calculate variance between two lists (Jaccard distance)."""
        set1, set2 = set(list1), set(list2)

        if not set1 and not set2:
            return 0.0

        if not set1 or not set2:
            return 1.0

        intersection = len(set1.intersection(set2))
        union = len(set1.union(set2))

        # Jaccard similarity = intersection / union
        # Jaccard distance = 1 - Jaccard similarity
        return 1.0 - (intersection / union if union > 0 else 0.0)

    def _is_capability_data_consistent(
        self, data1: CapabilityDetectionResult, data2: CapabilityDetectionResult
    ) -> bool:
        """
        Check if two capability detection results are consistent.

        Enhanced consistency checking with tolerance for minor variations.
        """
        if not (data1.detected and data2.detected):
            return False

        # Check format consistency (allow subset relationships)
        formats1 = set(f.get("code", "") for f in data1.formats)
        formats2 = set(f.get("code", "") for f in data2.formats)

        if formats1 and formats2:
            # Allow up to 50% difference in formats
            intersection = formats1.intersection(formats2)
            min_formats = min(len(formats1), len(formats2))
            if min_formats > 0 and len(intersection) / min_formats < 0.5:
                return False

        # Check resolution consistency (allow subset relationships)
        resolutions1 = set(data1.resolutions)
        resolutions2 = set(data2.resolutions)

        if resolutions1 and resolutions2:
            intersection = resolutions1.intersection(resolutions2)
            min_resolutions = min(len(resolutions1), len(resolutions2))
            if min_resolutions > 0 and len(intersection) / min_resolutions < 0.5:
                return False

        # Check frame rate consistency (more lenient)
        frame_rates1 = set(data1.frame_rates)
        frame_rates2 = set(data2.frame_rates)

        if frame_rates1 and frame_rates2:
            intersection = frame_rates1.intersection(frame_rates2)
            min_rates = min(len(frame_rates1), len(frame_rates2))
            if min_rates > 0 and len(intersection) / min_rates < 0.3:
                return False

        return True

    def get_effective_capability_metadata(
        self, device_path: str
    ) -> Optional[Dict[str, Any]]:
        """
        Get effective capability metadata for a device.

        Args:
            device_path: Device path

        Returns:
            Capability metadata or None if no data available
        """
        if device_path not in self._capability_states:
            return None

        capability_state = self._capability_states[device_path]
        effective_capability = capability_state.get_effective_capability()

        if effective_capability is None:
            return None

        # Build metadata from effective capability
        metadata = {
            "device_path": device_path,
            "detected": effective_capability.detected,
            "accessible": effective_capability.accessible,
            "device_name": effective_capability.device_name,
            "driver": effective_capability.driver,
            "formats": effective_capability.formats,
            "resolutions": effective_capability.resolutions,
            "frame_rates": effective_capability.frame_rates,
            "is_confirmed": capability_state.is_confirmed(),
            "validation_status": (
                "confirmed" if capability_state.is_confirmed() else "provisional"
            ),
            "consecutive_successes": capability_state.consecutive_successes,
            "consecutive_failures": capability_state.consecutive_failures,
            "last_probe_time": capability_state.last_probe_time,
            "diagnostics": {
                "has_been_probed": True,
                "merge_strategy": "frequency_weighted",
                "reason": "capability_state_exists",
            },
        }

        # Add frequency data if available
        if capability_state.format_frequency:
            metadata["format_frequency"] = capability_state.format_frequency
        if capability_state.resolution_frequency:
            metadata["resolution_frequency"] = capability_state.resolution_frequency
            metadata["all_resolutions"] = list(
                capability_state.resolution_frequency.keys()
            )
        else:
            # If no frequency data, use the resolutions from the capability
            metadata["all_resolutions"] = effective_capability.resolutions
        if capability_state.frame_rate_frequency:
            metadata["frame_rate_frequency"] = capability_state.frame_rate_frequency

        return metadata

    def get_stream_name_from_device_path(self, device_path: str) -> str:
        """
        Extract deterministic stream name from camera device path.

        Args:
            device_path: Camera device path (e.g., /dev/video0)

        Returns:
            Stream name for MediaMTX (e.g., camera0)
        """
        try:
            # Primary pattern: /dev/video<number>
            match = re.search(r"/dev/video(\d+)", device_path)
            if match:
                device_num = match.group(1)
                return f"camera{device_num}"

            # Secondary pattern: any path with video and number
            match = re.search(r"video(\d+)", device_path, re.IGNORECASE)
            if match:
                device_num = match.group(1)
                return f"camera{device_num}"

            # Tertiary fallback: extract any digits from the path
            digits = re.findall(r"\d+", device_path)
            if digits:
                return f"camera{digits[-1]}"

            # Final fallback: hash-based deterministic name
            hash_val = abs(hash(device_path)) % 1000
            self._logger.debug(
                f"Using hash-based stream name for {device_path}: camera_{hash_val}",
                extra={
                    "device_path": device_path,
                    "hash_value": hash_val,
                    "fallback_used": True,
                },
            )
            return f"camera_{hash_val}"

        except Exception as e:
            self._logger.warning(
                f"Error extracting stream name from {device_path}: {e}"
            )
            return "camera_unknown"

    async def _probe_device_formats_robust(
        self, device_path: str
    ) -> Optional[Dict[str, Any]]:
        """Probe device formats and resolutions with enhanced parsing and error handling."""
        try:
            process = await asyncio.wait_for(
                asyncio.create_subprocess_exec(
                    "v4l2-ctl",
                    "--device",
                    device_path,
                    "--list-formats-ext",
                    stdout=asyncio.subprocess.PIPE,
                    stderr=asyncio.subprocess.PIPE,
                ),
                timeout=self._detection_timeout,
            )

            stdout, stderr = await process.communicate()

            if process.returncode != 0:
                return None

            formats_output = stdout.decode()
            formats: List[Dict[str, Any]] = []
            resolutions: Set[str] = set()

            # Enhanced format parsing with multiple patterns
            current_format = None

            for line in formats_output.split("\n"):
                line = line.strip()
                if not line:
                    continue

                # Format detection with enhanced patterns
                format_patterns = [
                    r"\[(\d+)\]:\s*\'(\w+)\'\s*\(([^)]+)\)",  # [0]: 'YUYV' (YUYV 4:2:2)
                    r"Index\s*:\s*(\d+)\s*Type\s*:\s*Video\s*Capture\s*Pixel\s*Format\s*:\s*\'(\w+)\'",
                    r"Pixel\s*Format\s*:\s*\'(\w+)\'",
                ]

                for pattern in format_patterns:
                    match = re.search(pattern, line, re.IGNORECASE)
                    if match:
                        if len(match.groups()) >= 2:
                            format_code = (
                                match.group(2)
                                if len(match.groups()) >= 2
                                else match.group(1)
                            )
                            format_desc = (
                                match.group(3)
                                if len(match.groups()) >= 3
                                else format_code
                            )
                        else:
                            format_code = match.group(1)
                            format_desc = format_code

                        current_format = {
                            "code": format_code,
                            "description": format_desc,
                            "resolutions": [],
                        }
                        formats.append(current_format)
                        break

                # Resolution detection with enhanced patterns
                if current_format:
                    resolution_patterns = [
                        r"Size:\s*Discrete\s+(\d+)x(\d+)",  # Size: Discrete 640x480
                        r"(\d{3,4})\s*x\s*(\d{3,4})",  # 1920x1080
                        r"Width\s*(\d+)\s*Height\s*(\d+)",  # Width 1920 Height 1080
                        r"Resolution:\s*(\d+)\s*x\s*(\d+)",  # Resolution: 1920 x 1080
                    ]

                    for pattern in resolution_patterns:
                        match = re.search(pattern, line, re.IGNORECASE)
                        if match:
                            width = int(match.group(1))
                            height = int(match.group(2))

                            # Validate reasonable resolution ranges
                            if 160 <= width <= 4096 and 120 <= height <= 3072:
                                resolution = f"{width}x{height}"
                                current_format["resolutions"].append(resolution)
                                resolutions.add(resolution)
                            break

            # Fallback resolution extraction if no format-specific parsing worked
            if not resolutions:
                fallback_resolutions = re.findall(
                    r"(\d{3,4})x(\d{3,4})", formats_output
                )
                for width_str, height_str in fallback_resolutions:
                    width, height = int(width_str), int(height_str)
                    if 160 <= width <= 4096 and 120 <= height <= 3072:
                        resolution = f"{width}x{height}"
                        resolutions.add(resolution)

            return {
                "formats": formats,
                "resolutions": sorted(
                    list(resolutions),
                    key=lambda r: (int(r.split("x")[0]), int(r.split("x")[1])),
                    reverse=True,
                ),  # Sort by resolution, highest first
            }

        except asyncio.TimeoutError:
            self._logger.debug(f"Timeout probing device formats for {device_path}")
            return None
        except Exception as e:
            self._logger.debug(f"Error probing device formats for {device_path}: {e}")
            return None

    # Test interface methods (enhanced)
    def _get_capability_probe_interface(self):
        """Test hook: Enhanced capability probing interface for comprehensive testing."""
        return {
            "probe_device_info": self._probe_device_info_robust,
            "probe_device_formats": self._probe_device_formats_robust,
            "probe_device_framerates": self._probe_device_framerates_robust,
            "extract_frame_rates": self._extract_frame_rates_from_output,
            "select_preferred_rates": self._select_preferred_frame_rates,
            "update_capability_state": self._update_capability_validation_state,
            "check_consistency": self._is_capability_data_consistent,
            "get_effective_metadata": self.get_effective_capability_metadata,
            "stats": self.get_monitor_stats,
        }

    def _get_capability_state_for_testing(
        self, device_path: str
    ) -> Optional[DeviceCapabilityState]:
        """Test hook: Get capability state for validation testing."""
        return self._capability_states.get(device_path)

    def _set_capability_state_for_testing(
        self, device_path: str, state: DeviceCapabilityState
    ) -> None:
        """Test hook: Set capability state for validation testing."""
        self._capability_states[device_path] = state

    async def _inject_test_udev_event(self, device_path: str, action: str) -> None:
        """Test hook: Inject synthetic udev event for testing."""
        if not hasattr(self, "_test_mode"):
            self._logger.warning("Test event injection called outside test mode")
            return

        # Create mock udev device for testing
        class MockUdevDevice:
            def __init__(self, device_node: str, action: str):
                self.device_node = device_node
                self.action = action

        mock_device = MockUdevDevice(device_path, action)
        await self._process_udev_device_event(mock_device)

    def _set_test_mode(self, enabled: bool = True) -> None:
        """Test hook: Enable/disable test mode for injection methods."""
        if enabled:
            self._test_mode = True
        elif hasattr(self, "_test_mode"):
            delattr(self, "_test_mode")

    def _get_adaptive_polling_state_for_testing(self) -> Dict[str, Any]:
        """Test hook: Get adaptive polling state for testing."""
        return {
            "current_interval": self._current_poll_interval,
            "base_interval": self._base_poll_interval,
            "min_interval": self._min_poll_interval,
            "max_interval": self._max_poll_interval,
            "last_udev_time": self._last_udev_event_time,
            "failure_count": self._polling_failure_count,
            "adjustment_count": self._stats.get("adaptive_poll_adjustments", 0),
            "freshness_threshold": self._udev_event_freshness_threshold,
        }

    def _update_frequency_tracking(
        self, device_path: str, result: CapabilityDetectionResult
    ) -> None:
        """
        Update frequency tracking for capability data.

        Args:
            device_path: Device path
            result: Capability detection result
        """
        state = self._get_or_create_capability_state(device_path)
        self._update_capability_frequencies(state, result)
