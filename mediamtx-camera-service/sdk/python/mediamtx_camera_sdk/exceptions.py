"""
Custom exceptions for MediaMTX Camera Service SDK.

This module contains the exception classes used by the SDK for error handling.
"""


class CameraServiceError(Exception):
    """Base exception for camera service errors."""
    pass


class AuthenticationError(CameraServiceError):
    """Authentication failed."""
    pass


class ConnectionError(CameraServiceError):
    """Connection failed."""
    pass


class CameraNotFoundError(CameraServiceError):
    """Camera device not found."""
    pass


class MediaMTXError(CameraServiceError):
    """MediaMTX operation failed."""
    pass


class TimeoutError(CameraServiceError):
    """Operation timed out."""
    pass


class ValidationError(CameraServiceError):
    """Input validation failed."""
    pass
