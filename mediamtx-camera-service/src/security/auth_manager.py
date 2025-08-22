"""
Authentication manager for MediaMTX Camera Service.

Coordinates JWT and API key authentication, providing unified authentication
interface for the WebSocket server as specified in Architecture Decision AD-7.
"""

import logging
from typing import Dict, Optional
from dataclasses import dataclass

from .jwt_handler import JWTHandler
from .api_key_handler import APIKeyHandler


@dataclass
class AuthResult:
    """Authentication result structure."""
    
    authenticated: bool
    user_id: Optional[str] = None
    role: Optional[str] = None
    auth_method: Optional[str] = None
    expires_at: Optional[int] = None
    error_message: Optional[str] = None


class AuthManager:
    """
    Authentication manager coordinating JWT and API key authentication.
    
    Provides unified authentication interface for the WebSocket server,
    supporting both JWT tokens and API keys as specified in Architecture Decision AD-7.
    """
    
    def __init__(self, jwt_handler: JWTHandler, api_key_handler: APIKeyHandler):
        """
        Initialize authentication manager.
        
        Args:
            jwt_handler: JWT token handler
            api_key_handler: API key handler
        """
        self.jwt_handler = jwt_handler
        self.api_key_handler = api_key_handler
        self.logger = logging.getLogger(f"{__name__}.AuthManager")
        
        self.logger.info("Authentication manager initialized")
    
    def authenticate(self, auth_token: str, auth_type: str = "auto") -> AuthResult:
        """
        Authenticate user with JWT token or API key.
        
        Args:
            auth_token: Authentication token (JWT or API key)
            auth_type: Authentication type ("jwt", "api_key", or "auto")
            
        Returns:
            AuthResult with authentication status and user information
        """
        if not auth_token:
            return AuthResult(
                authenticated=False,
                user_id=None,
                role=None,
                auth_method=auth_type if auth_type != "auto" else "jwt",
                error_message="No authentication token provided"
            )
        
        if auth_type == "jwt" or auth_type == "auto":
            # Try JWT authentication first
            jwt_result = self._authenticate_jwt(auth_token)
            if jwt_result.authenticated:
                return jwt_result
            # If JWT fails and auth_type is "jwt", return the JWT error
            elif auth_type == "jwt":
                return jwt_result
        
        if auth_type == "api_key" or auth_type == "auto":
            # Try API key authentication
            api_key_result = self._authenticate_api_key(auth_token)
            if api_key_result.authenticated:
                return api_key_result
            # If API key fails and auth_type is "api_key", return the API key error
            elif auth_type == "api_key":
                return api_key_result
        
        # For invalid auth types, try auto authentication
        if auth_type not in ["jwt", "api_key", "auto"]:
            # Try both JWT and API key
            jwt_result = self._authenticate_jwt(auth_token)
            if jwt_result.authenticated:
                return jwt_result
            
            api_key_result = self._authenticate_api_key(auth_token)
            if api_key_result.authenticated:
                return api_key_result
        
        # For auto authentication, if both JWT and API key fail, return the JWT error
        # since JWT was tried first and is more specific
        if auth_type == "auto":
            return jwt_result  # Return the JWT error message
        
        # Determine which auth method was attempted last
        if auth_type == "jwt":
            auth_method = "jwt"
        elif auth_type == "api_key":
            auth_method = "api_key"
        else:  # auto
            auth_method = "jwt"  # JWT was tried first
        
        return AuthResult(
            authenticated=False,
            auth_method=auth_method,
            error_message="Invalid authentication token"
        )
    
    def _authenticate_jwt(self, token: str) -> AuthResult:
        """
        Authenticate using JWT token.
        
        Args:
            token: JWT token string
            
        Returns:
            AuthResult with JWT authentication status
        """
        try:
            claims = self.jwt_handler.validate_token(token)
            if claims:
                return AuthResult(
                    authenticated=True,
                    user_id=claims.user_id,
                    role=claims.role,
                    auth_method="jwt",
                    expires_at=claims.exp
                )
            else:
                return AuthResult(
                    authenticated=False,
                    auth_method="jwt",
                    error_message="Invalid or expired JWT token"
                )
        except Exception as e:
            self.logger.error("JWT authentication error: %s", e)
            return AuthResult(
                authenticated=False,
                auth_method="jwt",
                error_message="JWT authentication failed"
            )
    
    def _authenticate_api_key(self, key: str) -> AuthResult:
        """
        Authenticate using API key.
        
        Args:
            key: API key string
            
        Returns:
            AuthResult with API key authentication status
        """
        try:
            api_key = self.api_key_handler.validate_api_key(key)
            if api_key:
                return AuthResult(
                    authenticated=True,
                    user_id=f"api_key_{api_key.key_id}",
                    role=api_key.role,
                    auth_method="api_key"
                )
            else:
                return AuthResult(
                    authenticated=False,
                    auth_method="api_key",
                    error_message="Invalid or expired API key"
                )
        except Exception as e:
            self.logger.error("API key authentication error: %s", e)
            return AuthResult(
                authenticated=False,
                auth_method="api_key",
                error_message="API key authentication failed"
            )
    
    def has_permission(self, auth_result: AuthResult, required_role: str) -> bool:
        """
        Check if authenticated user has required role permission.
        
        Args:
            auth_result: Authentication result
            required_role: Minimum required role
            
        Returns:
            True if user has permission, False otherwise
        """
        if not auth_result.authenticated or not auth_result.role:
            return False
        
        role_hierarchy = {
            "viewer": 1,
            "operator": 2,
            "admin": 3
        }
        
        user_level = role_hierarchy.get(auth_result.role, 0)
        required_level = role_hierarchy.get(required_role, 0)
        
        return user_level >= required_level
    
    def generate_jwt_token(self, user_id: str, role: str, expiry_hours: Optional[int] = None) -> str:
        """
        Generate JWT token for user authentication.
        
        Args:
            user_id: Unique user identifier
            role: User role (viewer, operator, admin)
            expiry_hours: Token expiry in hours (default: 24)
            
        Returns:
            JWT token string
        """
        return self.jwt_handler.generate_token(user_id, role, expiry_hours)
    
    def create_api_key(self, name: str, role: str, expires_in_days: Optional[int] = None) -> str:
        """
        Create new API key.
        
        Args:
            name: Human-readable name for the key
            role: Key role (viewer, operator, admin)
            expires_in_days: Key expiry in days (None for no expiry)
            
        Returns:
            Generated API key string
        """
        return self.api_key_handler.create_api_key(name, role, expires_in_days)
    
    def revoke_api_key(self, key_id: str) -> bool:
        """
        Revoke API key by ID.
        
        Args:
            key_id: Key ID to revoke
            
        Returns:
            True if key was revoked, False if not found
        """
        return self.api_key_handler.revoke_api_key(key_id)
    
    def list_api_keys(self) -> list:
        """
        List all API keys (without exposing actual keys).
        
        Returns:
            List of API key information dictionaries
        """
        return self.api_key_handler.list_api_keys()
    
    def cleanup_expired_keys(self) -> int:
        """
        Remove expired API keys from storage.
        
        Returns:
            Number of keys removed
        """
        return self.api_key_handler.cleanup_expired_keys()
    
    def get_auth_methods(self) -> Dict[str, str]:
        """
        Get available authentication methods.
        
        Returns:
            Dictionary of authentication method descriptions
        """
        return {
            "jwt": "JWT token authentication for user sessions",
            "api_key": "API key authentication for service access"
        } 