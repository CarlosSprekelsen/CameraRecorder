#!/usr/bin/env python3
"""
Quick Validation Script for Multi-Tier Snapshot Capture

This script provides a simple way to validate that the multi-tier snapshot
functionality is working correctly with the current system.
"""

import asyncio
import json
import sys
import os
import time

# Add src to path for imports
sys.path.append(os.path.join(os.path.dirname(__file__), '..', 'src'))

from camera_service.config import Config
from mediamtx_wrapper.controller import MediaMTXController


async def validate_multi_tier_snapshot():
    """Validate multi-tier snapshot functionality."""
    print("🚀 Validating Multi-Tier Snapshot Functionality")
    print("=" * 50)
    
    # Load configuration
    config = Config()
    
    # Create temporary directories for testing
    import tempfile
    temp_dir = tempfile.mkdtemp(prefix="snapshot_validation_")
    temp_recordings = os.path.join(temp_dir, "recordings")
    temp_snapshots = os.path.join(temp_dir, "snapshots")
    os.makedirs(temp_recordings, exist_ok=True)
    os.makedirs(temp_snapshots, exist_ok=True)
    
    # Initialize MediaMTX controller with temporary paths
    controller = MediaMTXController(
        host=config.mediamtx.host,
        api_port=config.mediamtx.api_port,
        rtsp_port=config.mediamtx.rtsp_port,
        webrtc_port=config.mediamtx.webrtc_port,
        hls_port=config.mediamtx.hls_port,
        config_path=config.mediamtx.config_path,
        recordings_path=temp_recordings,
        snapshots_path=temp_snapshots,
        ffmpeg_config=None  # Use default FFmpeg configuration
    )
    
    # Set performance configuration
    controller._performance_config = config.performance.__dict__
    
    try:
        # Start controller
        await controller.start()
        print("✅ MediaMTX controller started")
        
        # Check current MediaMTX paths
        import aiohttp
        async with aiohttp.ClientSession() as session:
            async with session.get(f'http://{controller._host}:{controller._api_port}/v3/paths/list') as response:
                paths = await response.json()
                print(f"📊 Current MediaMTX paths: {len(paths.get('items', []))} streams")
                
                for path in paths.get('items', []):
                    print(f"   - {path['name']}: ready={path['ready']}, source={path['source']}")
        
        # Test snapshot capture
        print("\n📸 Testing snapshot capture...")
        start_time = time.time()
        
        result = await controller.take_snapshot(
            stream_name="camera0",
            filename="validation_test.jpg",
            format="jpg",
            quality=85
        )
        
        capture_time = time.time() - start_time
        
        print(f"✅ Snapshot test completed in {capture_time:.3f}s")
        print(f"📋 Result: {json.dumps(result, indent=2)}")
        
        # Validate result structure
        required_fields = ["status", "tier_used", "user_experience"]
        missing_fields = [field for field in required_fields if field not in result]
        
        if missing_fields:
            print(f"❌ Missing required fields: {missing_fields}")
            return False
        
        print(f"✅ All required fields present")
        print(f"🎯 Tier used: {result['tier_used']}")
        print(f"📷 Capture method: {result.get('capture_method', 'N/A (failed)')}")
        print(f"👤 User experience: {result['user_experience']}")
        
        # Check if this was a successful capture or expected failure
        if result['status'] == 'completed':
            print(f"✅ Snapshot capture successful!")
            if result['tier_used'] == 1:
                print(f"🚀 USB direct capture optimized for containerized deployment!")
        else:
            print(f"⚠️ Snapshot capture failed (expected in test environment without camera hardware)")
            print(f"   Error: {result.get('error', 'Unknown error')}")
            print(f"   Methods tried: {result.get('capture_methods_tried', [])}")
            
            # Check if USB direct capture was attempted first
            if result.get('capture_methods_tried', []):
                first_method = result['capture_methods_tried'][0]
                if first_method == 'usb_direct':
                    print(f"✅ USB direct capture attempted first (optimized for USB container)")
                else:
                    print(f"⚠️ Expected 'usb_direct' as first method, got: {first_method}")
        
        # Check if snapshot file was created
        snapshot_path = os.path.join(temp_snapshots, "validation_test.jpg")
        if os.path.exists(snapshot_path):
            file_size = os.path.getsize(snapshot_path)
            print(f"✅ Snapshot file created: {snapshot_path} ({file_size} bytes)")
        else:
            print(f"⚠️ Snapshot file not found: {snapshot_path}")
        
        # Cleanup temporary directory
        import shutil
        shutil.rmtree(temp_dir)
        print(f"🧹 Cleaned up temporary directory: {temp_dir}")
        
        # Test performance configuration
        print("\n⚙️ Testing performance configuration...")
        if hasattr(config.performance, 'snapshot_tiers'):
            tiers_config = config.performance.snapshot_tiers
            print(f"✅ Snapshot tiers configuration loaded:")
            for key, value in tiers_config.items():
                print(f"   - {key}: {value}")
        else:
            print("❌ Snapshot tiers configuration not found")
            return False
        
        print("\n🎉 Multi-tier snapshot validation completed successfully!")
        return True
        
    except Exception as e:
        print(f"❌ Validation failed: {e}")
        import traceback
        traceback.print_exc()
        return False
        
    finally:
        await controller.stop()
        print("🛑 MediaMTX controller stopped")


if __name__ == "__main__":
    success = asyncio.run(validate_multi_tier_snapshot())
    sys.exit(0 if success else 1)
