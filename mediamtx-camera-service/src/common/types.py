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

    def __post_init__(self):
        # Validate status values
        valid_statuses = ["CONNECTED", "DISCONNECTED", "ERROR", "BUSY"]
        if self.status not in valid_statuses:
            raise ValueError(
                f"Invalid status: {self.status}, must be one of {valid_statuses}"
            )
