"""
Common type definitions for MediaMTX Camera Service.
"""

from dataclasses import dataclass
from typing import Optional


@dataclass
class CameraDevice:
    """Camera device information structure."""

    device: str
    name: str = ""
    status: str = "CONNECTED"
    driver: Optional[str] = None
    capabilities: Optional[dict] = None

    def __post_init__(self):
        """Validate status values after initialization."""
        valid_statuses = ["CONNECTED", "DISCONNECTED", "ERROR", "BUSY"]
        if self.status not in valid_statuses:
            raise ValueError(f"Invalid status '{self.status}'. Must be one of: {valid_statuses}")

    @classmethod
    def from_device_path(cls, device_path: str, **kwargs):
        """Create CameraDevice from device path with support for legacy parameter names."""
        return cls(device=device_path, **kwargs)
