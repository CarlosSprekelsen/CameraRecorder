"""
Health check server for MediaMTX Camera Service.

Provides REST health endpoints for monitoring and Kubernetes readiness probes
as specified in Architecture Decision AD-6.
"""

import logging
import os
import mimetypes
from datetime import datetime, timezone
from typing import Dict
from dataclasses import dataclass

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
    
    def __init__(self, host: str = "0.0.0.0", port: int = 8003, recordings_path: str = None, snapshots_path: str = None):
        """
        Initialize health server.
        
        Args:
            host: Host address to bind to
            port: Port to listen on
            recordings_path: Path to recordings directory
            snapshots_path: Path to snapshots directory
        """
        self.host = host
        self.port = port
        self.recordings_path = recordings_path or "/opt/camera-service/recordings"
        self.snapshots_path = snapshots_path or "/opt/camera-service/snapshots"
        self.logger = logging.getLogger(f"{__name__}.HealthServer")
        
        # Service references (set by service manager)
        self.mediamtx_controller = None
        self.camera_monitor = None
        self.service_manager = None
        
        # Server state
        self.app = None
        self.runner = None
        self._started = False
        
        self.logger.info("Health server initialized on %s:%d with recordings_path=%s, snapshots_path=%s", 
                        host, port, self.recordings_path, self.snapshots_path)
    
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
        
        # HTTP polling fallback endpoint for WebSocket clients
        self.app.router.add_get("/api/cameras", self._handle_api_cameras)
        
        # File download endpoints (Epic E6)
        self.app.router.add_get("/files/recordings/{filename:.*}", self._handle_recording_download)
        self.app.router.add_get("/files/snapshots/{filename:.*}", self._handle_snapshot_download)
        
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
            
            # Convert HealthComponent objects to dictionaries for JSON serialization
            components_dict = {}
            for name, component in components.items():
                components_dict[name] = component.__dict__
            
            response = HealthResponse(
                status=status,
                timestamp=datetime.now(timezone.utc).isoformat(),
                components=components_dict
            )
            
            return web.json_response(response.__dict__, status=200)
            
        except Exception as e:
            self.logger.error("Error in system health check: %s", e)
            return web.json_response({
                "status": "unhealthy",
                "timestamp": datetime.now(timezone.utc).isoformat(),
                "error": str(e)
            }, status=500)
    
    async def _handle_cameras_health(self, request: web.Request) -> web.Response:
        """Handle cameras health endpoint."""
        try:
            health = await self._check_camera_health()
            
            return web.json_response({
                "status": health.status,
                "timestamp": datetime.now(timezone.utc).isoformat(),
                "details": health.details
            }, status=200)
            
        except Exception as e:
            self.logger.error("Error in cameras health check: %s", e)
            return web.json_response({
                "status": "unhealthy",
                "timestamp": datetime.now(timezone.utc).isoformat(),
                "error": str(e)
            }, status=500)
    
    async def _handle_mediamtx_health(self, request: web.Request) -> web.Response:
        """Handle MediaMTX health endpoint."""
        try:
            health = await self._check_mediamtx_health()
            
            return web.json_response({
                "status": health.status,
                "timestamp": datetime.now(timezone.utc).isoformat(),
                "details": health.details
            }, status=200)
            
        except Exception as e:
            self.logger.error("Error in MediaMTX health check: %s", e)
            return web.json_response({
                "status": "unhealthy",
                "timestamp": datetime.now(timezone.utc).isoformat(),
                "error": str(e)
            }, status=500)
    
    async def _handle_readiness(self, request: web.Request) -> web.Request:
        """Handle Kubernetes readiness probe."""
        try:
            # Check if all critical components are healthy
            mediamtx_health = await self._check_mediamtx_health()
            await self._check_camera_health()
            service_health = await self._check_service_health()
            
            # Service is ready if MediaMTX is healthy and service manager is healthy
            is_ready = (mediamtx_health.status == "healthy" and 
                       service_health.status == "healthy")
            
            if is_ready:
                return web.json_response({
                    "status": "ready",
                    "timestamp": datetime.now(timezone.utc).isoformat()
                }, status=200)
            else:
                return web.json_response({
                    "status": "not_ready",
                    "timestamp": datetime.now(timezone.utc).isoformat(),
                    "details": {
                        "mediamtx": mediamtx_health.status,
                        "service_manager": service_health.status
                    }
                }, status=503)
                
        except Exception as e:
            self.logger.error("Error in readiness check: %s", e)
            return web.json_response({
                "status": "not_ready",
                "timestamp": datetime.now(timezone.utc).isoformat(),
                "error": str(e)
            }, status=503)
    
    async def _check_mediamtx_health(self) -> HealthComponent:
        """Check MediaMTX health."""
        if not self.mediamtx_controller:
            return HealthComponent(
                status="unhealthy",
                details="MediaMTX controller not available",
                timestamp=datetime.now(timezone.utc).isoformat()
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
                    timestamp=datetime.now(timezone.utc).isoformat()
                )
            else:
                return HealthComponent(
                    status="unhealthy",
                    details="MediaMTX controller is unhealthy",
                    timestamp=datetime.now(timezone.utc).isoformat()
                )
                
        except Exception as e:
            return HealthComponent(
                status="unhealthy",
                details=f"MediaMTX health check failed: {str(e)}",
                timestamp=datetime.now(timezone.utc).isoformat()
            )
    
    async def _check_camera_health(self) -> HealthComponent:
        """Check camera monitor health."""
        if not self.camera_monitor:
            return HealthComponent(
                status="unhealthy",
                details="Camera monitor not available",
                timestamp=datetime.now(timezone.utc).isoformat()
            )
        
        try:
            # Check if camera monitor is running
            if hasattr(self.camera_monitor, 'is_running'):
                # is_running is a property, not a method
                is_running = self.camera_monitor.is_running
            else:
                # Fallback: check if monitor exists
                is_running = self.camera_monitor is not None
            
            if is_running:
                # Get camera count if available
                camera_count = 0
                if hasattr(self.camera_monitor, 'get_connected_cameras'):
                    try:
                        # get_connected_cameras is async, but we're in a sync context
                        # For now, use the known_devices directly
                        if hasattr(self.camera_monitor, '_known_devices'):
                            camera_count = len(self.camera_monitor._known_devices)
                        else:
                            camera_count = 0
                    except Exception as e:
                        self.logger.warning(f"Could not get camera count: {e}")
                        camera_count = 0
                
                return HealthComponent(
                    status="healthy",
                    details=f"Camera monitor is running with {camera_count} cameras",
                    timestamp=datetime.now(timezone.utc).isoformat()
                )
            else:
                return HealthComponent(
                    status="unhealthy",
                    details="Camera monitor is not running",
                    timestamp=datetime.now(timezone.utc).isoformat()
                )
                
        except Exception as e:
            return HealthComponent(
                status="unhealthy",
                details=f"Camera health check failed: {str(e)}",
                timestamp=datetime.now(timezone.utc).isoformat()
            )
    
    async def _check_service_health(self) -> HealthComponent:
        """Check service manager health."""
        if not self.service_manager:
            return HealthComponent(
                status="unhealthy",
                details="Service manager not available",
                timestamp=datetime.now(timezone.utc).isoformat()
            )
        
        try:
            # Check if service manager is running
            if hasattr(self.service_manager, 'is_running'):
                # is_running is a property, not a method
                is_running = self.service_manager.is_running
            else:
                # Fallback: check if service manager exists
                is_running = self.service_manager is not None
            
            if is_running:
                return HealthComponent(
                    status="healthy",
                    details="Service manager is running",
                    timestamp=datetime.now(timezone.utc).isoformat()
                )
            else:
                return HealthComponent(
                    status="unhealthy",
                    details="Service manager is not running",
                    timestamp=datetime.now(timezone.utc).isoformat()
                )
                
        except Exception as e:
            return HealthComponent(
                status="unhealthy",
                details=f"Service health check failed: {str(e)}",
                timestamp=datetime.now(timezone.utc).isoformat()
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

    async def _handle_api_cameras(self, request: web.Request) -> web.Response:
        """
        Handle API cameras endpoint for HTTP polling fallback.
        
        Requirements: REQ-FUNC-012
        Epic E6: Server Recording and Snapshot File Management Infrastructure
        
        Args:
            request: aiohttp request object
            
        Returns:
            JSON response with camera status
        """
        try:
            # Get camera status from camera monitor
            cameras = []
            if self.camera_monitor and hasattr(self.camera_monitor, '_known_devices'):
                try:
                    # Convert camera devices to simple format
                    for device_path, device_info in self.camera_monitor._known_devices.items():
                        if hasattr(device_info, 'status') and device_info.status == "CONNECTED":
                            cameras.append({
                                "device": device_path,
                                "status": "CONNECTED",
                                "name": getattr(device_info, 'name', f"Camera {device_path}"),
                                "resolution": getattr(device_info, 'resolution', "unknown"),
                                "fps": getattr(device_info, 'fps', 30)
                            })
                except Exception as e:
                    self.logger.warning(f"Could not get camera status from monitor: {e}")
            
            # Prepare response
            response_data = {
                "status": "healthy",
                "timestamp": datetime.now(timezone.utc).isoformat(),
                "cameras": cameras,
                "total": len(cameras),
                "connected": len(cameras)
            }
            
            return web.json_response(response_data, status=200)
            
        except Exception as e:
            self.logger.error(f"Error in API cameras endpoint: {e}")
            return web.json_response({
                "status": "unhealthy",
                "timestamp": datetime.now(timezone.utc).isoformat(),
                "error": str(e),
                "cameras": [],
                "total": 0,
                "connected": 0
            }, status=500)

    async def _handle_recording_download(self, request: web.Request) -> web.Response:
        """
        Handle recording file download requests.
        
        Requirements: REQ-FUNC-010
        Epic E6: Server Recording and Snapshot File Management Infrastructure
        
        Args:
            request: aiohttp request object
            
        Returns:
            File response or 404 error
        """
        try:
            filename = request.match_info['filename']
            
            # Security: Prevent directory traversal
            if '..' in filename or filename.startswith('/'):
                self.logger.warning(f"Directory traversal attempt detected: {filename}")
                return web.Response(status=400, text="Invalid filename")
            
            # Construct file path
            file_path = os.path.join(self.recordings_path, filename)
            
            # Check if file exists and is accessible
            if not os.path.exists(file_path):
                self.logger.warning(f"Recording file not found: {filename}")
                return web.Response(status=404, text="File not found")
            
            if not os.path.isfile(file_path):
                self.logger.warning(f"Path is not a file: {filename}")
                return web.Response(status=404, text="File not found")
            
            if not os.access(file_path, os.R_OK):
                self.logger.error(f"Permission denied accessing recording file: {filename}")
                return web.Response(status=403, text="Permission denied")
            
            # Get file info
            file_size = os.path.getsize(file_path)
            
            # Determine MIME type
            mime_type, _ = mimetypes.guess_type(filename)
            if not mime_type:
                # Default to video/mp4 for recordings
                mime_type = "video/mp4"
            
            # Log file access for security audit
            self.logger.info(f"Recording file download: {filename} (size: {file_size})")
            
            # Create response with proper headers
            response = web.FileResponse(
                path=file_path,
                headers={
                    'Content-Type': mime_type,
                    'Content-Disposition': f'attachment; filename="{filename}"',
                    'Content-Length': str(file_size),
                    'Accept-Ranges': 'bytes'
                }
            )
            
            return response
            
        except Exception as e:
            self.logger.error(f"Error serving recording file {filename}: {e}")
            return web.Response(status=500, text="Internal server error")

    async def _handle_snapshot_download(self, request: web.Request) -> web.Response:
        """
        Handle snapshot file download requests.
        
        Requirements: REQ-FUNC-011
        Epic E6: Server Recording and Snapshot File Management Infrastructure
        
        Args:
            request: aiohttp request object
            
        Returns:
            File response or 404 error
        """
        try:
            filename = request.match_info['filename']
            
            # Security: Prevent directory traversal
            if '..' in filename or filename.startswith('/'):
                self.logger.warning(f"Directory traversal attempt detected: {filename}")
                return web.Response(status=400, text="Invalid filename")
            
            # Construct file path
            file_path = os.path.join(self.snapshots_path, filename)
            
            # Check if file exists and is accessible
            if not os.path.exists(file_path):
                self.logger.warning(f"Snapshot file not found: {filename}")
                return web.Response(status=404, text="File not found")
            
            if not os.path.isfile(file_path):
                self.logger.warning(f"Path is not a file: {filename}")
                return web.Response(status=404, text="File not found")
            
            if not os.access(file_path, os.R_OK):
                self.logger.error(f"Permission denied accessing snapshot file: {filename}")
                return web.Response(status=403, text="Permission denied")
            
            # Get file info
            file_size = os.path.getsize(file_path)
            
            # Determine MIME type
            mime_type, _ = mimetypes.guess_type(filename)
            if not mime_type:
                # Default to image/jpeg for snapshots
                mime_type = "image/jpeg"
            
            # Log file access for security audit
            self.logger.info(f"Snapshot file download: {filename} (size: {file_size})")
            
            # Create response with proper headers
            response = web.FileResponse(
                path=file_path,
                headers={
                    'Content-Type': mime_type,
                    'Content-Disposition': f'attachment; filename="{filename}"',
                    'Content-Length': str(file_size)
                }
            )
            
            return response
            
        except Exception as e:
            self.logger.error(f"Error serving snapshot file {filename}: {e}")
            return web.Response(status=500, text="Internal server error") 