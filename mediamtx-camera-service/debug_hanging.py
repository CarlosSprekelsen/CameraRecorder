#!/usr/bin/env python3
"""
Debug script to isolate where HybridCameraMonitor hangs.
"""

import asyncio
import time
from src.camera_discovery.hybrid_monitor import HybridCameraMonitor

async def debug_monitor():
    print("=== DEBUGGING HYBRID CAMERA MONITOR ===")
    
    # Test 1: Basic initialization
    print("\n1. Testing basic initialization...")
    start_time = time.time()
    monitor = HybridCameraMonitor(
        device_range=[0], 
        poll_interval=1.0, 
        enable_capability_detection=False
    )
    print(f"   Monitor created in {time.time() - start_time:.2f}s")
    
    # Test 2: Start method - with more granular debugging
    print("\n2. Testing monitor.start()...")
    start_time = time.time()
    try:
        print("   About to call monitor.start()...")
        await asyncio.wait_for(monitor.start(), timeout=5.0)
        print(f"   Monitor started in {time.time() - start_time:.2f}s")
    except asyncio.TimeoutError:
        print(f"   Monitor.start() TIMED OUT after {time.time() - start_time:.2f}s")
        print("   This is where the hanging occurs!")
        return
    except Exception as e:
        print(f"   Monitor.start() FAILED: {e}")
        return
    
    # Test 3: Get connected cameras
    print("\n3. Testing get_connected_cameras()...")
    start_time = time.time()
    try:
        print("   About to call get_connected_cameras()...")
        cameras = await asyncio.wait_for(monitor.get_connected_cameras(), timeout=5.0)
        print(f"   get_connected_cameras() completed in {time.time() - start_time:.2f}s")
        print(f"   Result: {cameras}")
    except asyncio.TimeoutError:
        print(f"   get_connected_cameras() TIMED OUT after {time.time() - start_time:.2f}s")
        print("   This is where the hanging occurs!")
    except Exception as e:
        print(f"   get_connected_cameras() FAILED: {e}")
    
    # Test 4: Stop method
    print("\n4. Testing monitor.stop()...")
    start_time = time.time()
    try:
        print("   About to call monitor.stop()...")
        await asyncio.wait_for(monitor.stop(), timeout=5.0)
        print(f"   Monitor stopped in {time.time() - start_time:.2f}s")
    except asyncio.TimeoutError:
        print(f"   Monitor.stop() TIMED OUT after {time.time() - start_time:.2f}s")
        print("   This is where the hanging occurs!")
    except Exception as e:
        print(f"   Monitor.stop() FAILED: {e}")

if __name__ == "__main__":
    asyncio.run(debug_monitor())
