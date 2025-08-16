"""
Real MediaMTX testing infrastructure for all tests.

This module provides real MediaMTX service integration for testing,
replacing mocked components with actual service interactions.
"""

import asyncio
import json
import logging
import os
import subprocess
import tempfile
import time
from contextlib import asynccontextmanager
from pathlib import Path
from typing import Dict, Any, Optional

import pytest
import pytest_asyncio
import aiohttp
from aiohttp import web

from src.camera_service.config import MediaMTXConfig
from src.mediamtx_wrapper.controller import MediaMTXController

logger = logging.getLogger(__name__)


class MediaMTXTestInfrastructure:
    """Real MediaMTX testing infrastructure for all tests."""
    
    def __init__(self, config: Optional[MediaMTXConfig] = None):
        self.config = config or self._create_test_config()
        self.process: Optional[subprocess.Popen] = None
        self.temp_dir: Optional[str] = None
        self.config_file: Optional[str] = None
        self.controller: Optional[MediaMTXController] = None
        self._health_check_url: Optional[str] = None
        
    def _create_test_config(self) -> MediaMTXConfig:
        """Create test MediaMTX configuration."""
        return MediaMTXConfig(
            host="127.0.0.1",
            api_port=9997,
            rtsp_port=8554,
            webrtc_port=8889,
            hls_port=8888,
            config_path="/tmp/test_mediamtx.yml",
            recordings_path="/tmp/test_recordings",
            snapshots_path="/tmp/test_snapshots",
            health_check_interval=1,
            health_failure_threshold=3,
            health_circuit_breaker_timeout=5,
            health_max_backoff_interval=10,
            health_recovery_confirmation_threshold=2,
            backoff_base_multiplier=2.0,
            backoff_jitter_range=(0.8, 1.2),
            process_termination_timeout=3.0,
            process_kill_timeout=2.0,
        )
    
    async def setup_mediamtx_service(self) -> None:
        """Start real MediaMTX service for testing."""
        logger.info("Setting up real MediaMTX service for testing...")
        
        # Create temporary directory for MediaMTX
        self.temp_dir = tempfile.mkdtemp(prefix="mediamtx_test_")
        
        # Create MediaMTX configuration file
        self.config_file = os.path.join(self.temp_dir, "mediamtx.yml")
        self._create_mediamtx_config()
        
        # Create directories
        os.makedirs(self.config.recordings_path, exist_ok=True)
        os.makedirs(self.config.snapshots_path, exist_ok=True)
        
        # Start MediaMTX process
        # Use existing systemd-managed MediaMTX service instead of creating new instance
        try:
            # Check if MediaMTX service is running
            result = subprocess.run(
                ['systemctl', 'is-active', 'mediamtx'],
                capture_output=True,
                text=True,
                timeout=10
            )
            
            if result.returncode != 0 or result.stdout.strip() != 'active':
                raise RuntimeError("MediaMTX service is not running. Please start it with: sudo systemctl start mediamtx")
            
            # Wait for MediaMTX to be ready
            await self._wait_for_mediamtx_startup()
            
            # Create MediaMTX controller using existing service
            self.controller = MediaMTXController(
                host=self.config.host,
                api_port=self.config.api_port,
                rtsp_port=self.config.rtsp_port,
                webrtc_port=self.config.webrtc_port,
                hls_port=self.config.hls_port,
                config_path=self.config.config_path,
                recordings_path=self.config.recordings_path,
                snapshots_path=self.config.snapshots_path,
            )
            
            logger.info("Using existing MediaMTX service successfully")
            
        except Exception as e:
            logger.error(f"Failed to start MediaMTX service: {e}")
            await self.cleanup_mediamtx_service()
            raise
    
    async def _wait_for_mediamtx_startup(self, timeout: float = 10.0) -> None:
        """Wait for MediaMTX service to be ready."""
        self._health_check_url = f"http://{self.config.host}:{self.config.api_port}/v3/config/global/get"
        
        start_time = time.time()
        while time.time() - start_time < timeout:
            try:
                async with aiohttp.ClientSession() as session:
                    async with session.get(self._health_check_url) as response:
                        if response.status == 200:
                            logger.info("MediaMTX health check passed")
                            return
            except Exception:
                pass
            
            await asyncio.sleep(0.5)
        
        raise TimeoutError(f"MediaMTX service failed to start within {timeout} seconds")
    
    def _create_mediamtx_config(self) -> None:
        """Create MediaMTX configuration file."""
        config = {
            "api": True,
            "apiAddress": f"{self.config.host}:{self.config.api_port}",
            "rtsp": True,
            "rtspAddress": f"{self.config.host}:{self.config.rtsp_port}",
            "webrtc": True,
            "webrtcAddress": f"{self.config.host}:{self.config.webrtc_port}",
            "hls": True,
            "hlsAddress": f"{self.config.host}:{self.config.hls_port}",
            "paths": {
                "all": {
                    "sourceOnDemand": True,
                    "record": True,
                    "recordPath": self.config.recordings_path,
                    "snapshot": True,
                    "snapshotPath": self.config.snapshots_path,
                }
            }
        }
        
        with open(self.config_file, 'w') as f:
            import yaml
            yaml.dump(config, f)
    
    async def create_test_stream(self, stream_name: str, source: str = "/dev/video0") -> Dict[str, Any]:
        """Create real test stream in MediaMTX."""
        if not self.controller:
            raise RuntimeError("MediaMTX controller not initialized")
        
        # Create stream configuration
        stream_config = {
            "name": stream_name,
            "source": source,
            "record": True,
            "snapshot": True
        }
        
        # Add stream to MediaMTX
        result = await self.controller.create_stream(stream_name, source)
        
        # Wait for stream to be ready
        await asyncio.sleep(1.0)
        
        return {
            "stream_id": result,
            "config": stream_config,
            "urls": self.controller._generate_stream_urls(stream_name)
        }
    
    async def get_stream_status(self, stream_name: str) -> Dict[str, Any]:
        """Get real stream status from MediaMTX."""
        if not self.controller:
            raise RuntimeError("MediaMTX controller not initialized")
        
        return await self.controller.get_stream_status(stream_name)
    
    async def delete_test_stream(self, stream_name: str) -> None:
        """Delete test stream from MediaMTX."""
        if not self.controller:
            raise RuntimeError("MediaMTX controller not initialized")
        
        await self.controller.delete_stream(stream_name)
    
    async def cleanup_mediamtx_service(self) -> None:
        """Clean up MediaMTX test environment."""
        logger.info("Cleaning up MediaMTX test environment...")
        
        # Note: We don't stop the MediaMTX service since it's systemd-managed
        # and shared across all tests. Only clean up test-specific resources.
        
        # Clean up temporary directory
        if self.temp_dir and os.path.exists(self.temp_dir):
            try:
                import shutil
                shutil.rmtree(self.temp_dir)
            except Exception as e:
                logger.warning(f"Error cleaning up temp directory: {e}")
        
        # Reset state
        self.process = None
        self.temp_dir = None
        self.config_file = None
        self.controller = None
        self._health_check_url = None


# Pytest fixtures for easy integration
@pytest_asyncio.fixture
async def mediamtx_infrastructure():
    """Real MediaMTX infrastructure for testing."""
    infra = MediaMTXTestInfrastructure()
    await infra.setup_mediamtx_service()
    yield infra
    await infra.cleanup_mediamtx_service()


@pytest_asyncio.fixture
async def mediamtx_controller(mediamtx_infrastructure):
    """Real MediaMTX controller for testing."""
    return mediamtx_infrastructure.controller


@asynccontextmanager
async def mediamtx_test_context(config: Optional[MediaMTXConfig] = None):
    """Context manager for MediaMTX testing."""
    infra = MediaMTXTestInfrastructure(config)
    try:
        await infra.setup_mediamtx_service()
        yield infra
    finally:
        await infra.cleanup_mediamtx_service()


# Test utilities
async def create_test_stream_with_mediamtx(
    mediamtx_infrastructure: MediaMTXTestInfrastructure,
    stream_name: str = "test_camera",
    source: str = "/dev/video0"
) -> Dict[str, Any]:
    """Create a test stream using real MediaMTX infrastructure."""
    return await mediamtx_infrastructure.create_test_stream(stream_name, source)


async def verify_stream_accessible(stream_url: str, timeout: float = 5.0) -> bool:
    """Verify that a stream URL is accessible."""
    try:
        async with aiohttp.ClientSession() as session:
            async with session.get(stream_url, timeout=timeout) as response:
                return response.status == 200
    except Exception:
        return False


async def wait_for_mediamtx_health(mediamtx_infrastructure: MediaMTXTestInfrastructure, timeout: float = 10.0) -> bool:
    """Wait for MediaMTX health check to pass."""
    start_time = time.time()
    while time.time() - start_time < timeout:
        try:
            if mediamtx_infrastructure._health_check_url:
                async with aiohttp.ClientSession() as session:
                    async with session.get(mediamtx_infrastructure._health_check_url) as response:
                        if response.status == 200:
                            return True
        except Exception:
            pass
        
        await asyncio.sleep(0.5)
    
    return False
