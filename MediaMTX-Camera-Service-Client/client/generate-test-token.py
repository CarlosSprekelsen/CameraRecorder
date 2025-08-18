#!/usr/bin/env python3

import sys
import os
sys.path.append('/opt/camera-service/src')

from security.jwt_handler import JWTHandler

def generate_test_token():
    """Generate a valid JWT token for testing"""
    try:
        # Read JWT secret from the same source as the service (.env file)
        env_file_path = "/opt/camera-service/.env"
        jwt_secret = None
        
        # Try to read from .env file first (same as service)
        try:
            with open(env_file_path, 'r') as f:
                for line in f:
                    if line.startswith('CAMERA_SERVICE_JWT_SECRET='):
                        jwt_secret = line.split('=', 1)[1].strip()
                        break
        except (FileNotFoundError, PermissionError):
            pass
        
        # Fallback to environment variable
        if not jwt_secret:
            jwt_secret = os.getenv("CAMERA_SERVICE_JWT_SECRET")
        
        # Final fallback to default
        if not jwt_secret:
            jwt_secret = "dev-secret-change-me"
        
        if not jwt_secret:
            print("Could not find JWT secret in .env file or environment")
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
