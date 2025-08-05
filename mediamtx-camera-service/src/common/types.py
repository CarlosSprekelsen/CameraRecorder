"""
Common type definitions for MediaMTX Camera Service.
"""

from dataclasses import dataclass
from typing import Optional


@dataclass
class CameraDevice:
    """Camera device information structure."""

    device: str
    name: str
    status: str
    driver: Optional[str] = None
    capabilities: Optional[dict] = None

    def __init__(self, device: str = None, device_path: str = None, **kwargs):
        """Initialize CameraDevice with support for device_path alias."""
        if device_path is not None and device is None:
            device = device_path
        elif device is None:
            raise ValueError("Either device or device_path must be provided")
        
        self.device = device
        self.name = kwargs.get('name', '')
        self.status = kwargs.get('status', 'CONNECTED')  # Default to CONNECTED
        self.driver = kwargs.get('driver')
        self.capabilities = kwargs.get('capabilities')
        
        # Validate status values
        valid_statuses = ["CONNECTED", "DISCONNECTED", "ERROR", "BUSY"]
        if self.status not in valid_statuses:
            raise ValueError(
                f"Invalid status: {self.status}, must be one of {valid_statuses}"
            )
