#!/usr/bin/env python3

import sys
import os
sys.path.append('/opt/camera-service/src')

from security.jwt_handler import JWTHandler

def generate_test_token():
    """Generate a valid JWT token for testing"""
    try:
        # Get the actual JWT secret from environment
        jwt_secret = os.getenv("JWT_SECRET_KEY")
        if not jwt_secret:
            print("JWT_SECRET_KEY environment variable not found")
            return None
        
        # Create JWT handler with the actual secret
        jwt_handler = JWTHandler(secret_key=jwt_secret)
        
        # Generate token for operator role (can take snapshots)
        token = jwt_handler.generate_token("test_user", "operator", expiry_hours=24)
        
        print(f"Generated JWT token: {token}")
        return token
        
    except Exception as e:
        print(f"Error generating token: {e}")
        return None

if __name__ == "__main__":
    token = generate_test_token()
    if token:
        print(f"\nUse this token in your test script:")
        print(f"token: '{token}'")
