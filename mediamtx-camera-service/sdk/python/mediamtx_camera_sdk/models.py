"""
Data models for MediaMTX Camera Service SDK.

This module contains the data classes used to represent camera and recording information.
"""

from dataclasses import dataclass
from typing import List, Optional


@dataclass
class CameraInfo:
    """Camera device information."""
    device_path: str
    name: str
    capabilities: List[str]
    status: str
    stream_url: Optional[str] = None


@dataclass
class RecordingInfo:
    """Recording session information."""
    device_path: str
    recording_id: str
    filename: str
    start_time: float
    duration: Optional[float] = None
    status: str = "active"


@dataclass
class SnapshotInfo:
    """Snapshot information."""
    device_path: str
    filename: str
    timestamp: float
    size_bytes: Optional[int] = None
