"""
Security module for MediaMTX Camera Service.

Provides authentication, authorization, and security features including:
- JWT token generation and validation
- API key management
- WebSocket authentication middleware
- Rate limiting and connection control
"""

from .auth_manager import AuthManager
from .jwt_handler import JWTHandler
from .api_key_handler import APIKeyHandler
from .middleware import SecurityMiddleware

__all__ = [
    "AuthManager",
    "JWTHandler", 
    "APIKeyHandler",
    "SecurityMiddleware"
] 