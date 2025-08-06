"""
JWT token generation and validation for MediaMTX Camera Service.

Implements JWT authentication with configurable expiry, secure secret key,
and role-based access control as specified in Architecture Decision AD-7.
"""

import jwt
import time
import logging
from datetime import datetime, timedelta
from typing import Dict, Any, Optional, Union
from dataclasses import dataclass

logger = logging.getLogger(__name__)


@dataclass
class JWTClaims:
    """JWT token claims structure."""
    
    user_id: str
    role: str
    exp: int
    iat: int
    
    @classmethod
    def create(cls, user_id: str, role: str, expiry_hours: int = 24) -> "JWTClaims":
        """Create JWT claims with current timestamp and specified expiry."""
        now = int(time.time())
        return cls(
            user_id=user_id,
            role=role,
            iat=now,
            exp=now + (expiry_hours * 3600)
        )


class JWTHandler:
    """
    JWT token generation and validation handler.
    
    Implements JWT authentication with HS256 algorithm, configurable expiry,
    and role-based access control as specified in Architecture Decision AD-7.
    """
    
    VALID_ROLES = {"viewer", "operator", "admin"}
    DEFAULT_ALGORITHM = "HS256"
    DEFAULT_EXPIRY_HOURS = 24
    
    def __init__(self, secret_key: str, algorithm: str = DEFAULT_ALGORITHM):
        """
        Initialize JWT handler.
        
        Args:
            secret_key: Secret key for JWT signing and validation
            algorithm: JWT algorithm (default: HS256)
        """
        if not secret_key:
            raise ValueError("Secret key must be provided")
        
        self.secret_key = secret_key
        self.algorithm = algorithm
        self.logger = logging.getLogger(f"{__name__}.JWTHandler")
        
        self.logger.info("JWT handler initialized with algorithm: %s", algorithm)
    
    def generate_token(self, user_id: str, role: str, expiry_hours: Optional[int] = None) -> str:
        """
        Generate JWT token for user authentication.
        
        Args:
            user_id: Unique user identifier
            role: User role (viewer, operator, admin)
            expiry_hours: Token expiry in hours (default: 24)
            
        Returns:
            JWT token string
            
        Raises:
            ValueError: If role is invalid or parameters are missing
        """
        if not user_id:
            raise ValueError("User ID must be provided")
        
        if role not in self.VALID_ROLES:
            raise ValueError(f"Invalid role '{role}'. Must be one of: {self.VALID_ROLES}")
        
        expiry_hours = expiry_hours or self.DEFAULT_EXPIRY_HOURS
        claims = JWTClaims.create(user_id, role, expiry_hours)
        
        payload = {
            "user_id": claims.user_id,
            "role": claims.role,
            "iat": claims.iat,
            "exp": claims.exp
        }
        
        try:
            token = jwt.encode(payload, self.secret_key, algorithm=self.algorithm)
            self.logger.info("Generated JWT token for user %s with role %s", user_id, role)
            return token
        except Exception as e:
            self.logger.error("Failed to generate JWT token: %s", e)
            raise
    
    def validate_token(self, token: str) -> Optional[JWTClaims]:
        """
        Validate JWT token and extract claims.
        
        Args:
            token: JWT token string
            
        Returns:
            JWTClaims object if valid, None if invalid
            
        Raises:
            jwt.InvalidTokenError: If token is malformed or invalid
        """
        if not token:
            return None
        
        try:
            payload = jwt.decode(token, self.secret_key, algorithms=[self.algorithm])
            
            # Validate required claims
            required_fields = ["user_id", "role", "iat", "exp"]
            for field in required_fields:
                if field not in payload:
                    self.logger.warning("JWT token missing required field: %s", field)
                    return None
            
            # Validate role
            if payload["role"] not in self.VALID_ROLES:
                self.logger.warning("JWT token has invalid role: %s", payload["role"])
                return None
            
            claims = JWTClaims(
                user_id=payload["user_id"],
                role=payload["role"],
                iat=payload["iat"],
                exp=payload["exp"]
            )
            
            self.logger.debug("JWT token validated for user %s with role %s", 
                            claims.user_id, claims.role)
            return claims
            
        except jwt.ExpiredSignatureError:
            self.logger.warning("JWT token expired")
            return None
        except jwt.InvalidTokenError as e:
            self.logger.warning("Invalid JWT token: %s", e)
            return None
        except Exception as e:
            self.logger.error("Unexpected error validating JWT token: %s", e)
            return None
    
    def is_token_expired(self, token: str) -> bool:
        """
        Check if JWT token is expired without full validation.
        
        Args:
            token: JWT token string
            
        Returns:
            True if token is expired, False otherwise
        """
        try:
            # Decode without verification to check expiry
            payload = jwt.decode(token, options={"verify_signature": False})
            exp = payload.get("exp")
            if exp is None:
                return True
            
            return time.time() > exp
        except Exception:
            return True
    
    def get_token_info(self, token: str) -> Optional[Dict[str, Any]]:
        """
        Get token information without validation.
        
        Args:
            token: JWT token string
            
        Returns:
            Dictionary with token information or None if invalid
        """
        try:
            payload = jwt.decode(token, options={"verify_signature": False})
            return {
                "user_id": payload.get("user_id"),
                "role": payload.get("role"),
                "issued_at": payload.get("iat"),
                "expires_at": payload.get("exp"),
                "expired": self.is_token_expired(token)
            }
        except Exception:
            return None
    
    def has_permission(self, claims: JWTClaims, required_role: str) -> bool:
        """
        Check if user has required role permission.
        
        Args:
            claims: JWT claims object
            required_role: Minimum required role
            
        Returns:
            True if user has permission, False otherwise
        """
        role_hierarchy = {
            "viewer": 1,
            "operator": 2, 
            "admin": 3
        }
        
        user_level = role_hierarchy.get(claims.role, 0)
        required_level = role_hierarchy.get(required_role, 0)
        
        return user_level >= required_level 