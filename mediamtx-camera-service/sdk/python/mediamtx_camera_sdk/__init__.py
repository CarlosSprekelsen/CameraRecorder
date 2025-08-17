"""
MediaMTX Camera Service Python SDK

A Python SDK for interacting with the MediaMTX Camera Service via WebSocket JSON-RPC.
"""

__version__ = "1.0.0"
__author__ = "MediaMTX Camera Service Team"
__email__ = "team@mediamtx-camera-service.com"

from .client import CameraClient
from .exceptions import (
    CameraServiceError,
    AuthenticationError,
    ConnectionError,
    CameraNotFoundError,
    MediaMTXError,
)
from .models import CameraInfo, RecordingInfo

__all__ = [
    "CameraClient",
    "CameraServiceError",
    "AuthenticationError",
    "ConnectionError",
    "CameraNotFoundError",
    "MediaMTXError",
    "CameraInfo",
    "RecordingInfo",
]
