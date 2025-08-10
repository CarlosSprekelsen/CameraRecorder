"""
Security middleware for MediaMTX Camera Service.

Provides authentication middleware for WebSocket connections,
rate limiting, and connection control as specified in Architecture Decision AD-7.
"""

import logging
import time
from typing import Dict, Any, Optional, Set
from dataclasses import dataclass

from .auth_manager import AuthManager, AuthResult


@dataclass
class RateLimitInfo:
    """Rate limiting information for a client."""
    
    request_count: int = 0
    window_start: float = 0.0
    last_request: float = 0.0


class SecurityMiddleware:
    """
    Security middleware for WebSocket authentication and rate limiting.
    
    Provides authentication check before method execution, rate limiting,
    and connection control as specified in Architecture Decision AD-7.
    """
    
    def __init__(
        self,
        auth_manager: AuthManager,
        max_connections: int = 100,
        requests_per_minute: int = 60,
        window_size_seconds: int = 60
    ):
        """
        Initialize security middleware.
        
        Args:
            auth_manager: Authentication manager
            max_connections: Maximum concurrent connections
            requests_per_minute: Maximum requests per minute per client
            window_size_seconds: Rate limiting window size in seconds
        """
        self.auth_manager = auth_manager
        self.max_connections = max_connections
        self.requests_per_minute = requests_per_minute
        self.window_size_seconds = window_size_seconds
        
        # Connection tracking
        self.active_connections: Set[str] = set()
        self.connection_auth: Dict[str, AuthResult] = {}
        
        # Rate limiting tracking
        self.rate_limit_info: Dict[str, RateLimitInfo] = {}
        
        self.logger = logging.getLogger(f"{__name__}.SecurityMiddleware")
        
        self.logger.info("Security middleware initialized with max_connections=%d, requests_per_minute=%d",
                        max_connections, requests_per_minute)
    
    def can_accept_connection(self, client_id: str) -> bool:
        """
        Check if new connection can be accepted.
        
        Args:
            client_id: Client identifier
            
        Returns:
            True if connection can be accepted, False otherwise
        """
        if len(self.active_connections) >= self.max_connections:
            self.logger.warning("Connection limit reached (%d), rejecting connection from %s",
                              self.max_connections, client_id)
            return False
        
        if client_id in self.active_connections:
            self.logger.warning("Client %s already connected", client_id)
            return False
        
        return True
    
    def register_connection(self, client_id: str) -> None:
        """
        Register new client connection.
        
        Args:
            client_id: Client identifier
        """
        self.active_connections.add(client_id)
        self.logger.debug("Registered connection for client %s (total: %d)",
                         client_id, len(self.active_connections))
    
    def unregister_connection(self, client_id: str) -> None:
        """
        Unregister client connection.
        
        Args:
            client_id: Client identifier
        """
        self.active_connections.discard(client_id)
        self.connection_auth.pop(client_id, None)
        self.rate_limit_info.pop(client_id, None)
        
        self.logger.debug("Unregistered connection for client %s (total: %d)",
                         client_id, len(self.active_connections))
    
    async def authenticate_connection(
        self,
        client_id: str,
        auth_token: str,
        auth_type: str = "auto"
    ) -> AuthResult:
        """
        Authenticate client connection.
        
        Args:
            client_id: Client identifier
            auth_token: Authentication token
            auth_type: Authentication type ("jwt", "api_key", or "auto")
            
        Returns:
            AuthResult with authentication status
        """
        auth_result = self.auth_manager.authenticate(auth_token, auth_type)
        
        if auth_result.authenticated:
            self.connection_auth[client_id] = auth_result
            self.logger.info("Client %s authenticated with role %s using %s",
                           client_id, auth_result.role, auth_result.auth_method)
        else:
            self.logger.warning("Authentication failed for client %s: %s",
                              client_id, auth_result.error_message)
        
        return auth_result
    
    def is_authenticated(self, client_id: str) -> bool:
        """
        Check if client is authenticated.
        
        Args:
            client_id: Client identifier
            
        Returns:
            True if client is authenticated, False otherwise
        """
        return client_id in self.connection_auth
    
    def get_auth_result(self, client_id: str) -> Optional[AuthResult]:
        """
        Get authentication result for client.
        
        Args:
            client_id: Client identifier
            
        Returns:
            AuthResult if authenticated, None otherwise
        """
        return self.connection_auth.get(client_id)
    
    def has_permission(self, client_id: str, required_role: str) -> bool:
        """
        Check if client has required permission.
        
        Args:
            client_id: Client identifier
            required_role: Minimum required role
            
        Returns:
            True if client has permission, False otherwise
        """
        auth_result = self.get_auth_result(client_id)
        if not auth_result:
            return False
        
        return self.auth_manager.has_permission(auth_result, required_role)
    
    def check_rate_limit(self, client_id: str) -> bool:
        """
        Check if client is within rate limits.
        
        Args:
            client_id: Client identifier
            
        Returns:
            True if within rate limits, False otherwise
        """
        now = time.time()
        
        # Get or create rate limit info
        if client_id not in self.rate_limit_info:
            self.rate_limit_info[client_id] = RateLimitInfo(window_start=now)
        
        rate_info = self.rate_limit_info[client_id]
        
        # Check if window has expired
        if now - rate_info.window_start >= self.window_size_seconds:
            # Reset window
            rate_info.window_start = now
            rate_info.request_count = 0
        
        # Check if limit exceeded
        if rate_info.request_count >= self.requests_per_minute:
            self.logger.warning("Rate limit exceeded for client %s (%d requests in window)",
                              client_id, rate_info.request_count)
            return False
        
        # Update request count and timestamp
        rate_info.request_count += 1
        rate_info.last_request = now
        
        return True
    
    def get_connection_stats(self) -> Dict[str, Any]:
        """
        Get connection and rate limiting statistics.
        
        Returns:
            Dictionary with connection statistics
        """
        return {
            "active_connections": len(self.active_connections),
            "max_connections": self.max_connections,
            "authenticated_connections": len(self.connection_auth),
            "rate_limited_clients": len(self.rate_limit_info),
            "connection_limit_reached": len(self.active_connections) >= self.max_connections
        }
    
    def cleanup_expired_rate_limits(self) -> int:
        """
        Clean up expired rate limit entries.
        
        Returns:
            Number of entries cleaned up
        """
        now = time.time()
        expired_clients = []
        
        for client_id, rate_info in self.rate_limit_info.items():
            if now - rate_info.last_request > self.window_size_seconds * 2:
                expired_clients.append(client_id)
        
        for client_id in expired_clients:
            del self.rate_limit_info[client_id]
        
        if expired_clients:
            self.logger.debug("Cleaned up %d expired rate limit entries", len(expired_clients))
        
        return len(expired_clients)
    
    async def authenticate_and_check_permission(
        self,
        client_id: str,
        auth_token: str,
        required_role: str = "viewer"
    ) -> AuthResult:
        """
        Authenticate client and check permission in one operation.
        
        Args:
            client_id: Client identifier
            auth_token: Authentication token
            required_role: Minimum required role
            
        Returns:
            AuthResult with authentication and permission status
        """
        # Authenticate
        auth_result = await self.authenticate_connection(client_id, auth_token)
        
        if not auth_result.authenticated:
            return auth_result
        
        # Check permission
        if not self.has_permission(client_id, required_role):
            auth_result.authenticated = False
            auth_result.error_message = f"Insufficient permissions. Required role: {required_role}"
            self.logger.warning("Permission denied for client %s. Required: %s, Actual: %s",
                              client_id, required_role, auth_result.role)
        
        return auth_result 