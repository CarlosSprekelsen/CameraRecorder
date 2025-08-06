"""
Health check server for MediaMTX Camera Service.

Provides REST health endpoints for monitoring and Kubernetes readiness probes
as specified in Architecture Decision AD-6.
"""

import asyncio
import json
import logging
import time
from datetime import datetime
from typing import Dict, Any, Optional
from dataclasses import dataclass

import aiohttp
from aiohttp import web


@dataclass
class HealthComponent:
    """Health component status."""
    
    status: str  # "healthy", "degraded", "unhealthy"
    details: str
    timestamp: str


@dataclass
class HealthResponse:
    """Health response structure."""
    
    status: str  # "healthy", "degraded", "unhealthy"
    timestamp: str
    components: Dict[str, HealthComponent]


class HealthServer:
    """
    Health check server for MediaMTX Camera Service.
    
    Provides REST health endpoints for monitoring and Kubernetes readiness probes
    as specified in Architecture Decision AD-6.
    """
    
    def __init__(self, host: str = "0.0.0.0", port: int = 8003):
        """
        Initialize health server.
        
        Args:
            host: Host address to bind to
            port: Port to listen on
        """
        self.host = host
        self.port = port
        self.logger = logging.getLogger(f"{__name__}.HealthServer")
        
        # Service references (set by service manager)
        self.mediamtx_controller = None
        self.camera_monitor = None
        self.service_manager = None
        
        # Server state
        self.app = None
        self.runner = None
        self._started = False
        
        self.logger.info("Health server initialized on %s:%d", host, port)
    
    def set_mediamtx_controller(self, controller) -> None:
        """Set MediaMTX controller reference."""
        self.mediamtx_controller = controller
    
    def set_camera_monitor(self, monitor) -> None:
        """Set camera monitor reference."""
        self.camera_monitor = monitor
    
    def set_service_manager(self, service_manager) -> None:
        """Set service manager reference."""
        self.service_manager = service_manager
    
    async def start(self) -> None:
        """Start the health server."""
        if self._started:
            self.logger.warning("Health server already started")
            return
        
        # Create aiohttp application
        self.app = web.Application()
        
        # Add routes
        self.app.router.add_get("/health/system", self._handle_system_health)
        self.app.router.add_get("/health/cameras", self._handle_cameras_health)
        self.app.router.add_get("/health/mediamtx", self._handle_mediamtx_health)
        self.app.router.add_get("/health/ready", self._handle_readiness)
        
        # Start server
        self.runner = web.AppRunner(self.app)
        await self.runner.setup()
        
        site = web.TCPSite(self.runner, self.host, self.port)
        await site.start()
        
        self._started = True
        self.logger.info("Health server started on %s:%d", self.host, self.port)
    
    async def stop(self) -> None:
        """Stop the health server."""
        if not self._started:
            return
        
        if self.runner:
            await self.runner.cleanup()
        
        self._started = False
        self.logger.info("Health server stopped")
    
    async def _handle_system_health(self, request: web.Request) -> web.Response:
        """Handle system health endpoint."""
        try:
            components = {}
            
            # Check MediaMTX health
            mediamtx_health = await self._check_mediamtx_health()
            components["mediamtx"] = mediamtx_health
            
            # Check camera monitor health
            camera_health = await self._check_camera_health()
            components["camera_monitor"] = camera_health
            
            # Check service manager health
            service_health = await self._check_service_health()
            components["service_manager"] = service_health
            
            # Determine overall status
            status = self._determine_overall_status(components)
            
            response = HealthResponse(
                status=status,
                timestamp=datetime.utcnow().isoformat(),
                components=components
            )
            
            return web.json_response(response.__dict__, status=200)
            
        except Exception as e:
            self.logger.error("Error in system health check: %s", e)
            return web.json_response({
                "status": "unhealthy",
                "timestamp": datetime.utcnow().isoformat(),
                "error": str(e)
            }, status=500)
    
    async def _handle_cameras_health(self, request: web.Request) -> web.Response:
        """Handle cameras health endpoint."""
        try:
            health = await self._check_camera_health()
            
            return web.json_response({
                "status": health.status,
                "timestamp": datetime.utcnow().isoformat(),
                "details": health.details
            }, status=200)
            
        except Exception as e:
            self.logger.error("Error in cameras health check: %s", e)
            return web.json_response({
                "status": "unhealthy",
                "timestamp": datetime.utcnow().isoformat(),
                "error": str(e)
            }, status=500)
    
    async def _handle_mediamtx_health(self, request: web.Request) -> web.Response:
        """Handle MediaMTX health endpoint."""
        try:
            health = await self._check_mediamtx_health()
            
            return web.json_response({
                "status": health.status,
                "timestamp": datetime.utcnow().isoformat(),
                "details": health.details
            }, status=200)
            
        except Exception as e:
            self.logger.error("Error in MediaMTX health check: %s", e)
            return web.json_response({
                "status": "unhealthy",
                "timestamp": datetime.utcnow().isoformat(),
                "error": str(e)
            }, status=500)
    
    async def _handle_readiness(self, request: web.Request) -> web.Request:
        """Handle Kubernetes readiness probe."""
        try:
            # Check if all critical components are healthy
            mediamtx_health = await self._check_mediamtx_health()
            camera_health = await self._check_camera_health()
            service_health = await self._check_service_health()
            
            # Service is ready if MediaMTX is healthy and service manager is healthy
            is_ready = (mediamtx_health.status == "healthy" and 
                       service_health.status == "healthy")
            
            if is_ready:
                return web.json_response({
                    "status": "ready",
                    "timestamp": datetime.utcnow().isoformat()
                }, status=200)
            else:
                return web.json_response({
                    "status": "not_ready",
                    "timestamp": datetime.utcnow().isoformat(),
                    "details": {
                        "mediamtx": mediamtx_health.status,
                        "service_manager": service_health.status
                    }
                }, status=503)
                
        except Exception as e:
            self.logger.error("Error in readiness check: %s", e)
            return web.json_response({
                "status": "not_ready",
                "timestamp": datetime.utcnow().isoformat(),
                "error": str(e)
            }, status=503)
    
    async def _check_mediamtx_health(self) -> HealthComponent:
        """Check MediaMTX health."""
        if not self.mediamtx_controller:
            return HealthComponent(
                status="unhealthy",
                details="MediaMTX controller not available",
                timestamp=datetime.utcnow().isoformat()
            )
        
        try:
            # Check if MediaMTX controller is healthy
            if hasattr(self.mediamtx_controller, 'is_healthy'):
                is_healthy = self.mediamtx_controller.is_healthy()
            else:
                # Fallback: check if controller exists
                is_healthy = self.mediamtx_controller is not None
            
            if is_healthy:
                return HealthComponent(
                    status="healthy",
                    details="MediaMTX controller is healthy",
                    timestamp=datetime.utcnow().isoformat()
                )
            else:
                return HealthComponent(
                    status="unhealthy",
                    details="MediaMTX controller is unhealthy",
                    timestamp=datetime.utcnow().isoformat()
                )
                
        except Exception as e:
            return HealthComponent(
                status="unhealthy",
                details=f"MediaMTX health check failed: {str(e)}",
                timestamp=datetime.utcnow().isoformat()
            )
    
    async def _check_camera_health(self) -> HealthComponent:
        """Check camera monitor health."""
        if not self.camera_monitor:
            return HealthComponent(
                status="unhealthy",
                details="Camera monitor not available",
                timestamp=datetime.utcnow().isoformat()
            )
        
        try:
            # Check if camera monitor is running
            if hasattr(self.camera_monitor, 'is_running'):
                is_running = self.camera_monitor.is_running()
            else:
                # Fallback: check if monitor exists
                is_running = self.camera_monitor is not None
            
            if is_running:
                # Get camera count if available
                camera_count = 0
                if hasattr(self.camera_monitor, 'get_camera_count'):
                    camera_count = self.camera_monitor.get_camera_count()
                
                return HealthComponent(
                    status="healthy",
                    details=f"Camera monitor is running with {camera_count} cameras",
                    timestamp=datetime.utcnow().isoformat()
                )
            else:
                return HealthComponent(
                    status="unhealthy",
                    details="Camera monitor is not running",
                    timestamp=datetime.utcnow().isoformat()
                )
                
        except Exception as e:
            return HealthComponent(
                status="unhealthy",
                details=f"Camera health check failed: {str(e)}",
                timestamp=datetime.utcnow().isoformat()
            )
    
    async def _check_service_health(self) -> HealthComponent:
        """Check service manager health."""
        if not self.service_manager:
            return HealthComponent(
                status="unhealthy",
                details="Service manager not available",
                timestamp=datetime.utcnow().isoformat()
            )
        
        try:
            # Check if service manager is running
            if hasattr(self.service_manager, 'is_running'):
                is_running = self.service_manager.is_running()
            else:
                # Fallback: check if service manager exists
                is_running = self.service_manager is not None
            
            if is_running:
                return HealthComponent(
                    status="healthy",
                    details="Service manager is running",
                    timestamp=datetime.utcnow().isoformat()
                )
            else:
                return HealthComponent(
                    status="unhealthy",
                    details="Service manager is not running",
                    timestamp=datetime.utcnow().isoformat()
                )
                
        except Exception as e:
            return HealthComponent(
                status="unhealthy",
                details=f"Service health check failed: {str(e)}",
                timestamp=datetime.utcnow().isoformat()
            )
    
    def _determine_overall_status(self, components: Dict[str, HealthComponent]) -> str:
        """Determine overall health status from components."""
        if not components:
            return "unhealthy"
        
        # Count statuses
        status_counts = {"healthy": 0, "degraded": 0, "unhealthy": 0}
        for component in components.values():
            status_counts[component.status] += 1
        
        # Determine overall status
        if status_counts["unhealthy"] > 0:
            return "unhealthy"
        elif status_counts["degraded"] > 0:
            return "degraded"
        else:
            return "healthy"
    
    @property
    def is_running(self) -> bool:
        """Check if health server is running."""
        return self._started 