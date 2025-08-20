"""
Shared authentication utilities for MediaMTX Camera Service tests.

Requirements Coverage:
- REQ-AUTH-001: Authentication shall work with real JWT tokens
- REQ-AUTH-002: Authentication shall work with real API keys
- REQ-AUTH-003: Authentication shall reject invalid tokens

Test Categories: Unit/Integration
"""

import os
import time
import jwt
from typing import Optional
from src.security.jwt_handler import JWTHandler


def get_test_jwt_secret() -> str:
    """Get JWT secret for testing from environment or use safe default."""
    return os.getenv('CAMERA_SERVICE_JWT_SECRET', 'test-secret-dev-only')


def generate_valid_test_token(username: str = "test_user", role: str = "operator", expiry_hours: int = 24) -> str:
    """
    Generate valid JWT token for testing.
    
    Args:
        username: User identifier
        role: User role (operator, viewer, admin)
        expiry_hours: Token expiry in hours
        
    Returns:
        Valid JWT token string
    """
    secret = get_test_jwt_secret()
    jwt_handler = JWTHandler(secret_key=secret)
    return jwt_handler.generate_token(username, role, expiry_hours)


def generate_expired_test_token(username: str = "expired_user", role: str = "operator") -> str:
    """
    Generate expired JWT token for testing.
    
    Args:
        username: User identifier
        role: User role
        
    Returns:
        Expired JWT token string
    """
    secret = get_test_jwt_secret()
    jwt_handler = JWTHandler(secret_key=secret)
    return jwt_handler.generate_token(username, role, expiry_hours=-1)


def generate_invalid_test_token() -> str:
    """
    Generate invalid JWT token for testing.
    
    Returns:
        Invalid JWT token string
    """
    return "invalid.jwt.token"


def generate_tampered_test_token(username: str = "tampered_user", role: str = "admin") -> str:
    """
    Generate tampered JWT token for testing.
    
    Args:
        username: User identifier
        role: User role
        
    Returns:
        Tampered JWT token string
    """
    secret = get_test_jwt_secret()
    
    # Create payload with tampered data
    payload = {
        "user_id": username,
        "role": role,
        "iat": int(time.time()),
        "exp": int(time.time()) + 3600  # 1 hour from now
    }
    
    # Use wrong secret to create tampered token
    wrong_secret = "wrong-secret-key"
    return jwt.encode(payload, wrong_secret, algorithm="HS256")


def generate_malformed_test_token() -> str:
    """
    Generate malformed JWT token for testing.
    
    Returns:
        Malformed JWT token string
    """
    return "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid"


def validate_test_token(token: str) -> Optional[dict]:
    """
    Validate test token and return claims.
    
    Args:
        token: JWT token to validate
        
    Returns:
        Token claims if valid, None if invalid
    """
    try:
        secret = get_test_jwt_secret()
        jwt_handler = JWTHandler(secret_key=secret)
        claims = jwt_handler.validate_token(token)
        return claims.__dict__ if claims else None
    except Exception:
        return None


def is_test_token_expired(token: str) -> bool:
    """
    Check if test token is expired.
    
    Args:
        token: JWT token to check
        
    Returns:
        True if expired, False otherwise
    """
    try:
        secret = get_test_jwt_secret()
        jwt_handler = JWTHandler(secret_key=secret)
        return jwt_handler.is_token_expired(token)
    except Exception:
        return True
