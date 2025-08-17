#!/usr/bin/env python3
"""
Test script to generate valid JWT token and test camera detection.
"""

import asyncio
import json
import jwt
import time
import websockets
from typing import Dict, Any

# JWT Configuration (matches server config)
JWT_SECRET = "dev-secret-change-me"  # Default development secret
USER_ID = "test_user"
ROLE = "admin"

def generate_jwt_token() -> str:
    """Generate a valid JWT token for testing."""
    payload = {
        "user_id": USER_ID,
        "role": ROLE,
        "iat": int(time.time()),
        "exp": int(time.time()) + (24 * 3600)  # 24 hours
    }
    
    token = jwt.encode(payload, JWT_SECRET, algorithm="HS256")
    print(f"Generated JWT token: {token[:50]}...")
    return token

async def test_camera_detection():
    """Test camera detection with valid authentication."""
    token = generate_jwt_token()
    
    # Connect to WebSocket server
    uri = "ws://localhost:8002/ws"
    print(f"Connecting to {uri}...")
    
    async with websockets.connect(uri) as websocket:
        print("✅ Connected to WebSocket server")
        
        # Authenticate
        auth_request = {
            "jsonrpc": "2.0",
            "id": 1,
            "method": "authenticate",
            "params": {
                "token": token,
                "auth_type": "jwt"
            }
        }
        
        print("🔐 Sending authentication request...")
        await websocket.send(json.dumps(auth_request))
        
        # Get authentication response
        auth_response = await websocket.recv()
        auth_data = json.loads(auth_response)
        print(f"🔐 Authentication response: {auth_data}")
        
        if auth_data.get("result", {}).get("authenticated"):
            print("✅ Authentication successful!")
            
            # Test camera list
            camera_request = {
                "jsonrpc": "2.0",
                "id": 2,
                "method": "get_camera_list",
                "params": {}
            }
            
            print("📷 Requesting camera list...")
            await websocket.send(json.dumps(camera_request))
            
            # Get camera list response
            camera_response = await websocket.recv()
            camera_data = json.loads(camera_response)
            print(f"📷 Camera list response: {json.dumps(camera_data, indent=2)}")
            
            # Test camera status for each camera
            if "result" in camera_data and "cameras" in camera_data["result"]:
                cameras = camera_data["result"]["cameras"]
                print(f"📷 Found {len(cameras)} cameras:")
                
                for camera in cameras:
                    device_path = camera.get("device", "unknown")
                    print(f"  - {device_path}: {camera.get('status', 'unknown')}")
                    
                    # Test get_camera_status for this camera
                    status_request = {
                        "jsonrpc": "2.0",
                        "id": 3,
                        "method": "get_camera_status",
                        "params": {
                            "device": device_path
                        }
                    }
                    
                    print(f"  📊 Getting status for {device_path}...")
                    await websocket.send(json.dumps(status_request))
                    
                    status_response = await websocket.recv()
                    status_data = json.loads(status_response)
                    print(f"  📊 Status response: {json.dumps(status_data, indent=2)}")
            else:
                print("❌ No cameras found or invalid response format")
        else:
            print("❌ Authentication failed!")
            print(f"Error: {auth_data.get('error', 'Unknown error')}")

if __name__ == "__main__":
    print("🧪 Testing Camera Detection with Valid Authentication")
    print("=" * 60)
    
    # Check available video devices
    import subprocess
    try:
        result = subprocess.run(['ls', '-la', '/dev/video*'], capture_output=True, text=True)
        print("📹 Available video devices:")
        print(result.stdout)
    except Exception as e:
        print(f"❌ Error checking video devices: {e}")
    
    print("\n" + "=" * 60)
    
    # Run the test
    asyncio.run(test_camera_detection())
