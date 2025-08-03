# tests/mocks/common_types.py
"""
Mock implementations for common types used in testing.
"""

from dataclasses import dataclass
from typing import Optional


@dataclass
class CameraDevice:
    """Mock CameraDevice for testing."""
    device: str
    name: str
    status: str
    
    def __post_init__(self):
        # Validate status values
        valid_statuses = ["CONNECTED", "DISCONNECTED", "ERROR", "BUSY"]
        if self.status not in valid_statuses:
            raise ValueError(f"Invalid status: {self.status}, must be one of {valid_statuses}")