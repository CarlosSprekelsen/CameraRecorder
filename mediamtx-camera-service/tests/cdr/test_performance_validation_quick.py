#!/usr/bin/env python3
"""
CDR Performance Validation - Quick Diagnostic Version

Quick performance validation to identify issues and estimate full test duration.
"""

import asyncio
import json
import time
import psutil
import os
import statistics
import sys

import websockets
import numpy as np

# Add src to path for imports
sys.path.append('src')


async def quick_performance_test():
    """Quick performance test to identify issues."""
    print("🔍 Quick Performance Diagnostic")
    print("=" * 40)
    
    # Test 1: Basic connectivity
    print("\n1. Testing basic connectivity...")
    try:
        uri = "ws://localhost:8002/ws"
        start_time = time.time()
        async with websockets.connect(uri) as websocket:
            await websocket.send(json.dumps({
                "jsonrpc": "2.0",
                "id": 1,
                "method": "ping"
            }))
            response = await websocket.recv()
            end_time = time.time()
            response_time = (end_time - start_time) * 1000
            print(f"✅ Connectivity test: {response_time:.2f}ms")
            print(f"   Response: {response[:100]}...")
    except Exception as e:
        print(f"❌ Connectivity failed: {e}")
        return
    
    # Test 2: Single operation timing
    print("\n2. Testing single operation timing...")
    operations = [
        ("ping", {"method": "ping"}),
        ("get_camera_list", {"method": "get_camera_list"}),
        ("get_status", {"method": "get_status"}),
    ]
    
    for op_name, payload in operations:
        try:
            start_time = time.time()
            async with websockets.connect(uri) as websocket:
                await websocket.send(json.dumps({
                    "jsonrpc": "2.0",
                    "id": 1,
                    **payload
                }))
                response = await websocket.recv()
                end_time = time.time()
                response_time = (end_time - start_time) * 1000
                print(f"✅ {op_name}: {response_time:.2f}ms")
        except Exception as e:
            print(f"❌ {op_name} failed: {e}")
    
    # Test 3: Resource monitoring overhead
    print("\n3. Testing resource monitoring overhead...")
    start_time = time.time()
    for i in range(10):
        cpu_percent = psutil.cpu_percent(interval=0.1)
        memory = psutil.virtual_memory()
        print(f"   Sample {i+1}: CPU={cpu_percent:.1f}%, Memory={memory.percent:.1f}%")
    end_time = time.time()
    monitoring_time = (end_time - start_time) * 1000
    print(f"✅ Resource monitoring: {monitoring_time:.2f}ms for 10 samples")
    
    # Test 4: Concurrent connections (small scale)
    print("\n4. Testing concurrent connections (5 connections)...")
    async def single_worker(worker_id):
        try:
            start_time = time.time()
            async with websockets.connect(uri) as websocket:
                await websocket.send(json.dumps({
                    "jsonrpc": "2.0",
                    "id": worker_id,
                    "method": "ping"
                }))
                response = await websocket.recv()
                end_time = time.time()
                return (end_time - start_time) * 1000
        except Exception as e:
            return f"Error: {e}"
    
    start_time = time.time()
    tasks = [single_worker(i) for i in range(5)]
    results = await asyncio.gather(*tasks, return_exceptions=True)
    end_time = time.time()
    
    successful_results = [r for r in results if isinstance(r, (int, float))]
    failed_results = [r for r in results if isinstance(r, str)]
    
    print(f"✅ Concurrent test: {len(successful_results)}/5 successful")
    if successful_results:
        avg_time = statistics.mean(successful_results)
        print(f"   Average response time: {avg_time:.2f}ms")
    if failed_results:
        print(f"   Failures: {failed_results}")
    
    total_time = (end_time - start_time) * 1000
    print(f"   Total concurrent test time: {total_time:.2f}ms")
    
    # Estimate full test duration
    print("\n📊 Estimated Full Test Duration:")
    print("Based on current performance:")
    print(f"- Baseline test (50 operations): ~{len(successful_results) * 10:.0f} seconds")
    print(f"- Load tests (3 scenarios): ~{len(successful_results) * 60:.0f} seconds") 
    print(f"- Stress test: ~{len(successful_results) * 120:.0f} seconds")
    print(f"- Endurance test: 300 seconds (5 minutes)")
    print(f"- Recovery test: ~{len(successful_results) * 30:.0f} seconds")
    
    total_estimate = len(successful_results) * 10 + len(successful_results) * 60 + len(successful_results) * 120 + 300 + len(successful_results) * 30
    print(f"\n🎯 Total estimated duration: ~{total_estimate/60:.1f} minutes")
    
    if len(successful_results) == 5:
        print("\n✅ System appears ready for full performance testing")
    else:
        print(f"\n⚠️  System has issues - only {len(successful_results)}/5 concurrent connections working")
        print("   Full test may fail or take much longer than estimated")


if __name__ == "__main__":
    asyncio.run(quick_performance_test())
