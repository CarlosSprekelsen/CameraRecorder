"""
Real System Test Infrastructure

Provides real system integration testing infrastructure without mocking.
Follows testing guide principles: use real MediaMTX service, WebSocket connections, file operations.

Requirements Traceability:
- REQ-INT-001: Integration system shall provide real end-to-end system behavior validation
- REQ-INT-002: Integration system shall validate real MediaMTX server integration
- REQ-INT-003: Integration system shall test real WebSocket connections and camera control
- REQ-INT-004: Integration system shall test real file system operations

Test Categories: Integration
"""

import asyncio
import os
import subprocess
import tempfile
import time
from pathlib import Path
from typing import Optional, Dict, Any

import pytest
import pytest_asyncio
import aiohttp

from src.mediamtx_wrapper.controller import MediaMTXController
from src.websocket_server.server import WebSocketJsonRpcServer
from src.camera_service.config import Config, MediaMTXConfig, ServerConfig


class RealSystemTestBase:
    """Base class for all real system integration tests.
    
    Provides real MediaMTX service, WebSocket server, and file system operations.
    Follows testing guide: NEVER mock MediaMTX service, WebSocket connections, file operations.
    """
    
    @pytest_asyncio.fixture(autouse=True)
    async def real_mediamtx_service(self):
        """Ensure systemd-managed MediaMTX service is running and provide real controller.
        
        Requirements: REQ-INT-002
        """
        # Verify MediaMTX service is active via systemd
        try:
            result = subprocess.run(
                ['systemctl', 'is-active', 'mediamtx'],
                capture_output=True,
                text=True,
                timeout=10
            )
            
            if result.returncode != 0 or result.stdout.strip() != 'active':
                raise RuntimeError(
                    "MediaMTX systemd service is not running. "
                    "Please start it with: sudo systemctl start mediamtx"
                )
            
            # Wait for MediaMTX API to be ready
            await self._wait_for_mediamtx_ready()
            
            # Create real MediaMTX controller
            controller = MediaMTXController(
                host="127.0.0.1",
                api_port=9997,
                rtsp_port=8554,
                webrtc_port=8889,
                hls_port=8888,
                config_path="/tmp/test_mediamtx.yml",
                recordings_path="/tmp/test_recordings",
                snapshots_path="/tmp/test_snapshots"
            )
            
            # Start controller
            await controller.start()
            
            yield controller
            
            # Cleanup
            await controller.stop()
            
        except Exception as e:
            pytest.fail(f"Failed to setup real MediaMTX service: {e}")
    
    @pytest_asyncio.fixture(autouse=True)
    async def real_websocket_server(self):
        """Connect to real deployed WebSocket server for testing.
        
        Requirements: REQ-INT-003
        FIXED: Now connects to real deployed server instead of creating mock server
        """
        # Connect to real deployed WebSocket server
        real_server_url = "ws://localhost:8002/ws"
        
        # Verify real server is available
        try:
            # Test connection to real server
            async with aiohttp.ClientSession() as session:
                async with session.ws_connect(real_server_url) as ws:
                    # Send ping to verify server is responsive
                    await ws.ping()
                    
        except Exception as e:
            pytest.fail(
                f"Real WebSocket server not available at {real_server_url}. "
                f"Please ensure the deployed server is running: {e}"
            )
        
        yield {
            "server": None,  # No server instance - using real deployed server
            "port": 8002,    # Real server port
            "url": real_server_url
        }
        
        # No cleanup needed - we didn't create a server instance
    
    @pytest_asyncio.fixture(autouse=True)
    async def real_file_system(self):
        """Use real file system operations with temporary directories.
        
        Requirements: REQ-INT-004
        """
        # Create temporary directories for real file operations
        with tempfile.TemporaryDirectory(prefix="real_test_") as temp_dir:
            recordings_dir = os.path.join(temp_dir, "recordings")
            snapshots_dir = os.path.join(temp_dir, "snapshots")
            config_dir = os.path.join(temp_dir, "config")
            
            # Create directories
            os.makedirs(recordings_dir, exist_ok=True)
            os.makedirs(snapshots_dir, exist_ok=True)
            os.makedirs(config_dir, exist_ok=True)
            
            yield {
                "temp_dir": temp_dir,
                "recordings_dir": recordings_dir,
                "snapshots_dir": snapshots_dir,
                "config_dir": config_dir
            }
    
    @pytest_asyncio.fixture(autouse=True)
    async def real_test_config(self, real_file_system):
        """Create real test configuration with actual component paths."""
        return Config(
            server=ServerConfig(
                host="127.0.0.1",
                port=8002,
                websocket_path="/ws",
                max_connections=10
            ),
            mediamtx=MediaMTXConfig(
                host="127.0.0.1",
                api_port=9997,
                rtsp_port=8554,
                webrtc_port=8889,
                hls_port=8888,
                config_path=os.path.join(real_file_system["config_dir"], "mediamtx.yml"),
                recordings_path=real_file_system["recordings_dir"],
                snapshots_path=real_file_system["snapshots_dir"],
                health_check_interval=5,
                health_failure_threshold=3,
                health_circuit_breaker_timeout=10,
                health_max_backoff_interval=20,
                health_recovery_confirmation_threshold=2,
                backoff_base_multiplier=1.5,
                backoff_jitter_range=(0.8, 1.2),
                process_termination_timeout=3.0,
                process_kill_timeout=2.0
            ),
            camera=None,  # Will be set by specific tests
            logging=None,  # Will be set by specific tests
            recording=None,  # Will be set by specific tests
            snapshots=None  # Will be set by specific tests
        )
    
    async def _wait_for_mediamtx_ready(self, timeout: float = 10.0) -> None:
        """Wait for MediaMTX service to be ready."""
        health_check_url = "http://127.0.0.1:9997/v3/config/global/get"
        
        start_time = time.time()
        while time.time() - start_time < timeout:
            try:
                async with aiohttp.ClientSession() as session:
                    async with session.get(health_check_url) as response:
                        if response.status == 200:
                            return
            except Exception:
                pass
            
            await asyncio.sleep(0.5)
        
        raise TimeoutError(f"MediaMTX service failed to be ready within {timeout} seconds")
    
    def _find_free_port(self) -> int:
        """Find a free port for WebSocket server."""
        import socket
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
            s.bind(('', 0))
            s.listen(1)
            port = s.getsockname()[1]
        return port
