#!/usr/bin/env python3
"""
Test script to validate Python SDK functionality.
"""

import asyncio
import sys
import os

# Add SDK to path
sys.path.append(os.path.join(os.path.dirname(__file__), 'sdk', 'python'))

from mediamtx_camera_sdk import CameraClient

async def test_sdk():
    """Test the Python SDK functionality."""
    print("🧪 Testing Python SDK Functionality")
    print("=" * 50)
    
    # Generate JWT token
    import jwt
    import time
    
    JWT_SECRET = "dev-secret-change-me"
    USER_ID = "test_user"
    ROLE = "admin"
    
    payload = {
        "user_id": USER_ID,
        "role": ROLE,
        "iat": int(time.time()),
        "exp": int(time.time()) + (24 * 3600)
    }
    
    token = jwt.encode(payload, JWT_SECRET, algorithm="HS256")
    print(f"Generated JWT token: {token[:50]}...")
    
    # Create client
    client = CameraClient(
        host="localhost",
        port=8002,
        auth_type="jwt",
        auth_token=token
    )
    
    try:
        # Connect
        await client.connect()
        print("✅ Connected to camera service")
        
        # Test ping
        pong = await client.ping()
        print(f"✅ Ping response: {pong}")
        
        # Get camera list
        cameras = await client.get_camera_list()
        print(f"✅ Found {len(cameras)} cameras:")
        for camera in cameras:
            print(f"  - {camera.name} ({camera.device_path}) - {camera.status}")
        
        if cameras:
            # Test get camera status (this should work with SDK)
            camera = cameras[0]
            status = await client.get_camera_status(camera.device_path)
            print(f"✅ Camera status: {status.status}")
            
            # Test snapshot
            snapshot = await client.take_snapshot(camera.device_path)
            print(f"✅ Snapshot taken: {snapshot.filename}")
            
    except Exception as e:
        print(f"❌ SDK test error: {e}")
    finally:
        await client.disconnect()
        print("✅ Disconnected")

if __name__ == "__main__":
    asyncio.run(test_sdk())
