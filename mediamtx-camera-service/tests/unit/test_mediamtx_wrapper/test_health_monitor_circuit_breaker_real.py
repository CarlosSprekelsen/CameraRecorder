# tests/unit/test_mediamtx_wrapper/test_health_monitor_circuit_breaker_real.py
"""
Test health monitoring circuit breaker flapping scenarios and edge cases.

Requirements Traceability:
- REQ-MEDIA-004: MediaMTX controller shall implement circuit breaker pattern for fault tolerance
- REQ-ERROR-003: MediaMTX controller shall maintain operation during MediaMTX failures
- REQ-MEDIA-004: MediaMTX controller shall provide configurable health monitoring with exponential backoff

Story Coverage: S2 - MediaMTX Integration
IV&V Control Point: Real circuit breaker validation

Test policy: Verify circuit breaker stability under alternating success/failure
patterns using REAL MediaMTX controller implementation and real HTTP servers.
"""

import pytest
import asyncio
import socket
import time
from contextlib import asynccontextmanager
from aiohttp import web

from src.mediamtx_wrapper.controller import MediaMTXController


def get_free_port() -> int:
    """Get a free port for the test server."""
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
        s.bind(("127.0.0.1", 0))
        return s.getsockname()[1]


@asynccontextmanager
async def controlled_mediamtx_server(host: str, port: int, response_sequence: list):
    """Start a MediaMTX server with controlled response sequence."""
    request_count = 0
    
    async def health_endpoint(request: web.Request):
        nonlocal request_count
        
        if request_count < len(response_sequence):
            response_type, response_data = response_sequence[request_count]
            request_count += 1
            
            if response_type == "success":
                return web.json_response({
                    "status": "healthy",
                    "version": "v1.0.0",
                    "uptime": 3600,
                    "api_port": port,
                    **response_data
                })
            elif response_type == "error":
                return web.json_response(
                    {"error": response_data.get("message", "Service Error")}, 
                    status=response_data.get("status", 500)
                )
            elif response_type == "timeout":
                await asyncio.sleep(response_data.get("delay", 5))  # Cause timeout
                return web.json_response({"status": "healthy"})
        else:
            # Default to success after sequence
            return web.json_response({
                "status": "healthy",
                "version": "v1.0.0",
                "uptime": 3600,
                "api_port": port
            })

    app = web.Application()
    app.router.add_get("/v3/config/global/get", health_endpoint)

    runner = web.AppRunner(app)
    await runner.setup()
    site = web.TCPSite(runner, host, port)
    await site.start()
    try:
        yield {"port": port, "request_count": lambda: request_count}
    finally:
        await runner.cleanup()


class TestHealthMonitorFlappingReal:
    """Test health monitoring circuit breaker flapping resistance with REAL implementation."""

    @pytest.fixture
    def controller_config(self):
        """Create controller configuration with fast timers for testing."""
        return {
            "host": "127.0.0.1",
            "rtsp_port": 8554,
            "webrtc_port": 8889,
            "hls_port": 8888,
            "config_path": "/tmp/test_config.yml",
            "recordings_path": "/tmp/recordings",
            "snapshots_path": "/tmp/snapshots",
            # Fast timers for testing
            "health_check_interval": 0.05,
            "health_failure_threshold": 3,
            "health_circuit_breaker_timeout": 0.3,
            "health_max_backoff_interval": 0.2,  # Cap backoff for faster testing
            "health_recovery_confirmation_threshold": 3,
            "backoff_base_multiplier": 1.5,  # Smaller multiplier for faster testing
            "backoff_jitter_range": (1.0, 1.0),  # No jitter for predictable testing
        }

    @pytest.mark.asyncio
    async def test_circuit_breaker_activation_threshold(self, controller_config):
        """Test circuit breaker opens exactly at configured failure threshold using REAL implementation."""
        port = get_free_port()
        
        # Pattern: 3 consecutive failures to trigger CB
        response_sequence = [
            ("error", {"status": 500, "message": "Service Error"}),
            ("error", {"status": 500, "message": "Service Error"}),
            ("error", {"status": 500, "message": "Service Error"}),  # 3 failures - should trigger CB
            ("success", {"serverVersion": "1.0.0"}),  # After CB activation
            ("success", {"serverVersion": "1.0.0"}),
            ("success", {"serverVersion": "1.0.0"}),  # Recovery
        ]
        
        async with controlled_mediamtx_server("127.0.0.1", port, response_sequence) as server:
            controller = MediaMTXController(api_port=port, **controller_config)
            
            await controller.start()
            try:
                # Wait for circuit breaker to activate (accounting for backoff delays)
                # With 1.5^1=1.5*0.05=0.075s, 1.5^2=2.25*0.05=0.11s, total ~0.3s
                timeout = 3.0  # Give plenty of time
                for i in range(30):  # Check every 100ms for 3 seconds
                    await asyncio.sleep(0.1)
                    if controller._health_state["circuit_breaker_activations"] > 0:
                        break
                else:
                    pytest.fail(f"Circuit breaker not activated within {timeout}s")
                
                # Verify circuit breaker activated exactly once
                assert controller._health_state["circuit_breaker_activations"] == 1
                assert controller._health_state["consecutive_failures"] >= 3
                
            finally:
                await controller.stop()

    @pytest.mark.asyncio
    async def test_flapping_resistance_during_recovery(self, controller_config, caplog):
        """Test circuit breaker resists flapping during recovery phase using REAL implementation."""
        port = get_free_port()
        
        # Pattern: failures → CB → timeout → alternating success/failure (should not fully recover)
        response_sequence = [
            ("error", {"status": 500, "message": "Error"}),
            ("error", {"status": 500, "message": "Error"}),
            ("error", {"status": 500, "message": "Error"}),  # Trigger CB
            # Recovery attempts - alternating pattern should prevent full recovery
            ("success", {"serverVersion": "1.0.0"}),  # 1st success during recovery
            ("error", {"status": 500, "message": "Error"}),  # Reset confirmation counter
            ("success", {"serverVersion": "1.0.0"}),  # 1st success again
            ("error", {"status": 500, "message": "Error"}),  # Reset confirmation counter again
            ("success", {"serverVersion": "1.0.0"}),  # 1st success yet again
            ("success", {"serverVersion": "1.0.0"}),  # 2nd consecutive success
            ("success", {"serverVersion": "1.0.0"}),  # 3rd consecutive success - should fully recover
        ]
        
        async with controlled_mediamtx_server("127.0.0.1", port, response_sequence) as server:
            controller = MediaMTXController(api_port=port, **controller_config)
            controller._health_recovery_confirmation_threshold = 3
            
            with caplog.at_level("INFO"):
                await controller.start()
                try:
                    # Wait for recovery sequence to complete
                    await asyncio.sleep(1.0)  
                    
                    # Verify circuit breaker eventually recovered after stable successes
                    assert controller._health_state["circuit_breaker_activations"] == 1
                    assert controller._health_state["recovery_count"] == 1
                    
                    # Verify intermediate "IMPROVING" logs during partial recovery
                    log_messages = [record.message for record in caplog.records]
                    improving_logs = [msg for msg in log_messages if "IMPROVING" in msg]
                    assert len(improving_logs) >= 2, "Should log multiple partial recovery attempts"
                    
                finally:
                    await controller.stop()

    @pytest.mark.asyncio
    async def test_rapid_flapping_scenario(self, controller_config):
        """Test circuit breaker behavior under rapid success/failure alternation using REAL implementation."""
        port = get_free_port()
        
        # Pattern: failures → CB → rapid alternation → eventual stable recovery
        response_sequence = [
            ("error", {"status": 503, "message": "Unavailable"}),
            ("error", {"status": 503, "message": "Unavailable"}),
            ("error", {"status": 503, "message": "Unavailable"}),  # Trigger CB
        ]
        
        # Add rapid alternation during recovery (10 cycles)
        for _ in range(10):
            response_sequence.extend([
                ("success", {"serverVersion": "1.0.0"}),
                ("error", {"status": 503, "message": "Unavailable"}),  # Reset
            ])
        
        # Add stable recovery (3 consecutive successes)
        for _ in range(3):
            response_sequence.append(("success", {"serverVersion": "1.0.0"}))
        
        async with controlled_mediamtx_server("127.0.0.1", port, response_sequence) as server:
            controller = MediaMTXController(api_port=port, **controller_config)
            controller._health_recovery_confirmation_threshold = 2
            
            await controller.start()
            try:
                # Wait for flapping sequence to complete
                await asyncio.sleep(2.0)  
                
                # Should have activated CB once and eventually recovered
                assert controller._health_state["circuit_breaker_activations"] == 1
                # Recovery should happen despite the flapping
                assert controller._health_state["recovery_count"] >= 1
                
            finally:
                await controller.stop()

    @pytest.mark.asyncio 
    async def test_multiple_circuit_breaker_cycles(self, controller_config):
        """Test multiple circuit breaker activation/recovery cycles using REAL implementation."""
        port = get_free_port()
        
        # Pattern: CB cycle 1 → recovery → CB cycle 2 → recovery
        response_sequence = []
        
        # First CB cycle
        for _ in range(3):
            response_sequence.append(("error", {"status": 500, "message": "Error"}))
        
        # Recovery from first CB
        for _ in range(3):
            response_sequence.append(("success", {"serverVersion": "1.0.0"}))
        
        # Second CB cycle
        for _ in range(3):
            response_sequence.append(("error", {"status": 500, "message": "Error"}))
        
        # Recovery from second CB
        for _ in range(3):
            response_sequence.append(("success", {"serverVersion": "1.0.0"}))
        
        async with controlled_mediamtx_server("127.0.0.1", port, response_sequence) as server:
            controller = MediaMTXController(api_port=port, **controller_config)
            controller._health_recovery_confirmation_threshold = 2
            
            await controller.start()
            try:
                # Wait for both cycles to complete
                await asyncio.sleep(1.5)  
                
                # Should have multiple CB activations and recoveries
                assert controller._health_state["circuit_breaker_activations"] == 2
                assert controller._health_state["recovery_count"] == 2
                
            finally:
                await controller.stop()

    @pytest.mark.asyncio
    async def test_recovery_confirmation_boundary_conditions(self, controller_config):
        """Test exact boundary conditions for recovery confirmation using REAL implementation."""
        port = get_free_port()
        
        # Pattern: trigger CB → exactly N-1 successes → fail → exactly N successes → recover
        response_sequence = [
            ("error", {"status": 500, "message": "Error"}),
            ("error", {"status": 500, "message": "Error"}),
            ("error", {"status": 500, "message": "Error"}),  # Trigger CB
            # Exactly 2 successes (one less than threshold of 3)
            ("success", {"serverVersion": "1.0.0"}),
            ("success", {"serverVersion": "1.0.0"}),
            ("error", {"status": 500, "message": "Error"}),  # Reset confirmation
            # Exactly 3 successes (meets threshold)
            ("success", {"serverVersion": "1.0.0"}),
            ("success", {"serverVersion": "1.0.0"}),
            ("success", {"serverVersion": "1.0.0"}),  # Should trigger recovery
        ]
        
        async with controlled_mediamtx_server("127.0.0.1", port, response_sequence) as server:
            controller = MediaMTXController(api_port=port, **controller_config)
            controller._health_recovery_confirmation_threshold = 3
            
            await controller.start()
            try:
                # Wait for boundary condition sequence
                await asyncio.sleep(1.0)  
                
                # Should activate CB once and recover once
                assert controller._health_state["circuit_breaker_activations"] == 1
                assert controller._health_state["recovery_count"] == 1
                
            finally:
                await controller.stop()

    @pytest.mark.asyncio
    async def test_no_premature_circuit_breaker_reset(self, controller_config):
        """Test that circuit breaker doesn't reset prematurely using REAL implementation."""
        port = get_free_port()
        
        # Pattern: trigger CB → insufficient successes → should remain in CB
        response_sequence = [
            ("error", {"status": 500, "message": "Error"}),
            ("error", {"status": 500, "message": "Error"}),
            ("error", {"status": 500, "message": "Error"}),  # Trigger CB
            # Only 2 successes (less than threshold of 3)
            ("success", {"serverVersion": "1.0.0"}),
            ("success", {"serverVersion": "1.0.0"}),
            # Then some more errors
            ("error", {"status": 500, "message": "Error"}),
            ("error", {"status": 500, "message": "Error"}),
        ]
        
        async with controlled_mediamtx_server("127.0.0.1", port, response_sequence) as server:
            controller = MediaMTXController(api_port=port, **controller_config)
            controller._health_recovery_confirmation_threshold = 3
            
            await controller.start()
            try:
                # Wait for sequence
                await asyncio.sleep(0.8)  
                
                # Should activate CB but not recover
                assert controller._health_state["circuit_breaker_activations"] == 1
                assert controller._health_state["recovery_count"] == 0
                
            finally:
                await controller.stop()
