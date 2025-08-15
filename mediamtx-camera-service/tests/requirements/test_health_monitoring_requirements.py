"""
Health Monitoring Requirements Test Coverage

Tests specifically designed to validate health monitoring requirements:
- REQ-HEALTH-001: System shall provide comprehensive logging for health monitoring
- REQ-HEALTH-002: System shall support structured logging for production environments
- REQ-HEALTH-003: System shall enable correlation ID tracking across components

These tests are designed to fail if health monitoring requirements are not met.
"""

import asyncio
import json
import logging
import tempfile
import os
import time
import pytest
from typing import List, Dict, Any, Optional
from dataclasses import dataclass
from pathlib import Path

from src.camera_service.service_manager import ServiceManager
from src.camera_service.config import Config, ServerConfig, MediaMTXConfig, CameraConfig, LoggingConfig
from src.websocket_server.server import WebSocketJsonRpcServer
from src.mediamtx_wrapper.controller import MediaMTXController
from src.logging_config import setup_logging, get_correlation_id, set_correlation_id


@dataclass
class HealthMonitoringMetrics:
    """Health monitoring metrics for requirement validation."""
    requirement: str
    test_name: str
    log_entries_count: int
    structured_logs_count: int
    correlation_ids_tracked: int
    health_checks_performed: int
    success: bool
    error_message: str = None


class HealthMonitoringRequirementsValidator:
    """Validates health monitoring requirements through comprehensive testing."""
    
    def __init__(self):
        self.metrics: List[HealthMonitoringMetrics] = []
        self.health_thresholds = {
            "min_log_entries": 10,        # REQ-HEALTH-001: Minimum log entries for health monitoring
            "structured_log_percentage": 80,  # REQ-HEALTH-002: 80%+ structured logs
            "correlation_id_coverage": 90,    # REQ-HEALTH-003: 90%+ correlation ID coverage
            "health_check_frequency": 5       # REQ-HEALTH-001: Health checks every 5 seconds
        }
        self.log_entries = []
        self.correlation_ids = set()
    
    async def setup_test_environment(self) -> Dict[str, Any]:
        """Set up test environment for health monitoring testing."""
        temp_dir = tempfile.mkdtemp(prefix="health_test_")
        
        # Create log directory
        log_dir = os.path.join(temp_dir, "logs")
        os.makedirs(log_dir, exist_ok=True)
        
        # Create real MediaMTX configuration
        mediamtx_config = MediaMTXConfig(
            host="127.0.0.1",
            api_port=10004,
            rtsp_port=8554,
            webrtc_port=8889,
            hls_port=8888,
            config_path=f"{temp_dir}/mediamtx.yml",
            recordings_path=f"{temp_dir}/recordings",
            snapshots_path=f"{temp_dir}/snapshots"
        )
        
        # Create real logging configuration
        logging_config = LoggingConfig(
            level="DEBUG",
            file_enabled=True,
            file_path=os.path.join(log_dir, "health_test.log"),
            json_format=True,
            correlation_id_enabled=True
        )
        
        # Create real service configuration
        config = Config(
            server=ServerConfig(host="127.0.0.1", port=8005, websocket_path="/ws"),
            mediamtx=mediamtx_config,
            camera=CameraConfig(device_range=[0, 1, 2], poll_interval=0.1),
            logging=logging_config
        )
        
        return {
            "temp_dir": temp_dir,
            "log_dir": log_dir,
            "config": config,
            "mediamtx_config": mediamtx_config,
            "logging_config": logging_config
        }
    
    async def test_req_health_001_comprehensive_logging(self):
        """REQ-HEALTH-001: System shall provide comprehensive logging for health monitoring."""
        env = await self.setup_test_environment()
        
        # Setup logging
        setup_logging(env["logging_config"])
        logger = logging.getLogger("health_test")
        
        # Create real service manager
        service_manager = ServiceManager(env["config"])
        
        try:
            # Start service and perform operations to generate logs
            await service_manager.start()
            
            # Perform various operations to generate health-related logs
            operations = [
                self._simulate_health_check(logger),
                self._simulate_camera_discovery(logger),
                self._simulate_stream_operation(logger),
                self._simulate_error_condition(logger),
                self._simulate_performance_metric(logger)
            ]
            
            # Execute operations
            await asyncio.gather(*operations)
            
            # Wait for logs to be written
            await asyncio.sleep(1)
            
            # Analyze log file
            log_file_path = env["logging_config"].file_path
            log_entries = self._analyze_log_file(log_file_path)
            
            # Count health-related log entries
            health_log_entries = [entry for entry in log_entries if self._is_health_related(entry)]
            
            # Record metrics
            self.metrics.append(HealthMonitoringMetrics(
                requirement="REQ-HEALTH-001",
                test_name="comprehensive_logging",
                log_entries_count=len(log_entries),
                structured_logs_count=len([e for e in log_entries if self._is_structured(e)]),
                correlation_ids_tracked=len(self.correlation_ids),
                health_checks_performed=len(health_log_entries),
                success=len(log_entries) >= self.health_thresholds["min_log_entries"]
            ))
            
            # Validate requirement
            assert len(log_entries) >= self.health_thresholds["min_log_entries"], \
                f"REQ-HEALTH-001 FAILED: Only {len(log_entries)} log entries, need {self.health_thresholds['min_log_entries']}"
            
            assert len(health_log_entries) > 0, \
                "REQ-HEALTH-001 FAILED: No health-related log entries found"
                
        finally:
            await service_manager.stop()
            import shutil
            shutil.rmtree(env["temp_dir"], ignore_errors=True)
    
    async def test_req_health_002_structured_logging(self):
        """REQ-HEALTH-002: System shall support structured logging for production environments."""
        env = await self.setup_test_environment()
        
        # Setup structured logging
        setup_logging(env["logging_config"])
        logger = logging.getLogger("structured_test")
        
        # Create real WebSocket server
        websocket_server = WebSocketJsonRpcServer(
            host="127.0.0.1",
            port=8006,
            websocket_path="/ws",
            max_connections=100
        )
        await websocket_server.start()
        
        try:
            # Generate various log entries
            await self._generate_structured_logs(logger)
            
            # Wait for logs to be written
            await asyncio.sleep(1)
            
            # Analyze log file
            log_file_path = env["logging_config"].file_path
            log_entries = self._analyze_log_file(log_file_path)
            
            # Count structured log entries
            structured_logs = [entry for entry in log_entries if self._is_structured(entry)]
            structured_percentage = (len(structured_logs) / len(log_entries)) * 100 if log_entries else 0
            
            # Record metrics
            self.metrics.append(HealthMonitoringMetrics(
                requirement="REQ-HEALTH-002",
                test_name="structured_logging",
                log_entries_count=len(log_entries),
                structured_logs_count=len(structured_logs),
                correlation_ids_tracked=len(self.correlation_ids),
                health_checks_performed=0,
                success=structured_percentage >= self.health_thresholds["structured_log_percentage"]
            ))
            
            # Validate requirement
            assert structured_percentage >= self.health_thresholds["structured_log_percentage"], \
                f"REQ-HEALTH-002 FAILED: Only {structured_percentage:.1f}% structured logs, need {self.health_thresholds['structured_log_percentage']}%"
            
            # Verify structured log format
            for entry in structured_logs:
                assert "timestamp" in entry, "REQ-HEALTH-002 FAILED: Structured log missing timestamp"
                assert "level" in entry, "REQ-HEALTH-002 FAILED: Structured log missing level"
                assert "message" in entry, "REQ-HEALTH-002 FAILED: Structured log missing message"
                
        finally:
            await websocket_server.stop()
            import shutil
            shutil.rmtree(env["temp_dir"], ignore_errors=True)
    
    async def test_req_health_003_correlation_id_tracking(self):
        """REQ-HEALTH-003: System shall enable correlation ID tracking across components."""
        env = await self.setup_test_environment()
        
        # Setup logging with correlation ID tracking
        setup_logging(env["logging_config"])
        logger = logging.getLogger("correlation_test")
        
        # Create real MediaMTX controller
        controller = MediaMTXController(
            host=env["mediamtx_config"].host,
            api_port=env["mediamtx_config"].api_port,
            rtsp_port=env["mediamtx_config"].rtsp_port,
            webrtc_port=env["mediamtx_config"].webrtc_port,
            hls_port=env["mediamtx_config"].hls_port,
            config_path=env["mediamtx_config"].config_path,
            recordings_path=env["mediamtx_config"].recordings_path,
            snapshots_path=env["mediamtx_config"].snapshots_path
        )
        await controller.start()
        
        try:
            # Generate logs with correlation IDs across multiple components
            await self._generate_correlation_id_logs(logger, controller)
            
            # Wait for logs to be written
            await asyncio.sleep(1)
            
            # Analyze log file
            log_file_path = env["logging_config"].file_path
            log_entries = self._analyze_log_file(log_file_path)
            
            # Count logs with correlation IDs
            correlation_id_logs = [entry for entry in log_entries if self._has_correlation_id(entry)]
            correlation_percentage = (len(correlation_id_logs) / len(log_entries)) * 100 if log_entries else 0
            
            # Record metrics
            self.metrics.append(HealthMonitoringMetrics(
                requirement="REQ-HEALTH-003",
                test_name="correlation_id_tracking",
                log_entries_count=len(log_entries),
                structured_logs_count=len([e for e in log_entries if self._is_structured(e)]),
                correlation_ids_tracked=len(self.correlation_ids),
                health_checks_performed=0,
                success=correlation_percentage >= self.health_thresholds["correlation_id_coverage"]
            ))
            
            # Validate requirement
            assert correlation_percentage >= self.health_thresholds["correlation_id_coverage"], \
                f"REQ-HEALTH-003 FAILED: Only {correlation_percentage:.1f}% logs have correlation IDs, need {self.health_thresholds['correlation_id_coverage']}%"
            
            # Verify correlation ID consistency across components
            correlation_ids = self._extract_correlation_ids(log_entries)
            assert len(correlation_ids) > 0, "REQ-HEALTH-003 FAILED: No correlation IDs found"
            
            # Verify correlation ID format
            for corr_id in correlation_ids:
                assert len(corr_id) > 0, "REQ-HEALTH-003 FAILED: Empty correlation ID found"
                
        finally:
            await controller.stop()
            import shutil
            shutil.rmtree(env["temp_dir"], ignore_errors=True)
    
    async def _simulate_health_check(self, logger: logging.Logger) -> None:
        """Simulate health check operation."""
        correlation_id = f"health_check_{int(time.time())}"
        set_correlation_id(correlation_id)
        self.correlation_ids.add(correlation_id)
        
        logger.info("Health check started", extra={
            "operation": "health_check",
            "component": "service_manager",
            "status": "started"
        })
        
        await asyncio.sleep(0.1)
        
        logger.info("Health check completed", extra={
            "operation": "health_check",
            "component": "service_manager",
            "status": "completed",
            "duration_ms": 100
        })
    
    async def _simulate_camera_discovery(self, logger: logging.Logger) -> None:
        """Simulate camera discovery operation."""
        correlation_id = f"camera_discovery_{int(time.time())}"
        set_correlation_id(correlation_id)
        self.correlation_ids.add(correlation_id)
        
        logger.info("Camera discovery started", extra={
            "operation": "camera_discovery",
            "component": "camera_monitor",
            "devices_found": 2
        })
        
        await asyncio.sleep(0.1)
        
        logger.info("Camera discovery completed", extra={
            "operation": "camera_discovery",
            "component": "camera_monitor",
            "status": "completed"
        })
    
    async def _simulate_stream_operation(self, logger: logging.Logger) -> None:
        """Simulate stream operation."""
        correlation_id = f"stream_op_{int(time.time())}"
        set_correlation_id(correlation_id)
        self.correlation_ids.add(correlation_id)
        
        logger.info("Stream operation started", extra={
            "operation": "stream_creation",
            "component": "mediamtx_controller",
            "stream_name": "test_stream"
        })
        
        await asyncio.sleep(0.1)
        
        logger.info("Stream operation completed", extra={
            "operation": "stream_creation",
            "component": "mediamtx_controller",
            "status": "completed"
        })
    
    async def _simulate_error_condition(self, logger: logging.Logger) -> None:
        """Simulate error condition."""
        correlation_id = f"error_{int(time.time())}"
        set_correlation_id(correlation_id)
        self.correlation_ids.add(correlation_id)
        
        logger.error("Error condition detected", extra={
            "operation": "error_handling",
            "component": "error_handler",
            "error_type": "connection_timeout",
            "retry_count": 3
        })
    
    async def _simulate_performance_metric(self, logger: logging.Logger) -> None:
        """Simulate performance metric logging."""
        correlation_id = f"perf_{int(time.time())}"
        set_correlation_id(correlation_id)
        self.correlation_ids.add(correlation_id)
        
        logger.info("Performance metric recorded", extra={
            "operation": "performance_monitoring",
            "component": "performance_monitor",
            "metric": "response_time",
            "value_ms": 150
        })
    
    async def _generate_structured_logs(self, logger: logging.Logger) -> None:
        """Generate structured log entries."""
        for i in range(20):
            correlation_id = f"structured_{int(time.time())}_{i}"
            set_correlation_id(correlation_id)
            self.correlation_ids.add(correlation_id)
            
            logger.info(f"Structured log entry {i}", extra={
                "operation": "structured_logging",
                "component": "log_generator",
                "entry_id": i,
                "timestamp": time.time()
            })
            
            await asyncio.sleep(0.01)
    
    async def _generate_correlation_id_logs(self, logger: logging.Logger, controller: MediaMTXController) -> None:
        """Generate logs with correlation IDs across components."""
        # Generate logs for different components
        components = ["service_manager", "camera_monitor", "mediamtx_controller", "websocket_server"]
        
        for i, component in enumerate(components):
            correlation_id = f"corr_{component}_{int(time.time())}"
            set_correlation_id(correlation_id)
            self.correlation_ids.add(correlation_id)
            
            logger.info(f"Component operation {i}", extra={
                "operation": "component_operation",
                "component": component,
                "operation_id": i,
                "correlation_id": correlation_id
            })
            
            await asyncio.sleep(0.01)
    
    def _analyze_log_file(self, log_file_path: str) -> List[Dict[str, Any]]:
        """Analyze log file and extract entries."""
        log_entries = []
        
        if os.path.exists(log_file_path):
            with open(log_file_path, 'r') as f:
                for line in f:
                    line = line.strip()
                    if line:
                        try:
                            # Try to parse as JSON (structured log)
                            entry = json.loads(line)
                            log_entries.append(entry)
                        except json.JSONDecodeError:
                            # Plain text log
                            log_entries.append({"message": line, "format": "plain"})
        
        return log_entries
    
    def _is_health_related(self, log_entry: Dict[str, Any]) -> bool:
        """Check if log entry is health-related."""
        message = log_entry.get("message", "").lower()
        operation = log_entry.get("operation", "").lower()
        
        health_keywords = ["health", "status", "monitor", "check", "alive", "dead"]
        return any(keyword in message or keyword in operation for keyword in health_keywords)
    
    def _is_structured(self, log_entry: Dict[str, Any]) -> bool:
        """Check if log entry is structured."""
        return isinstance(log_entry, dict) and "timestamp" in log_entry
    
    def _has_correlation_id(self, log_entry: Dict[str, Any]) -> bool:
        """Check if log entry has correlation ID."""
        return "correlation_id" in log_entry or "corr_id" in log_entry
    
    def _extract_correlation_ids(self, log_entries: List[Dict[str, Any]]) -> List[str]:
        """Extract correlation IDs from log entries."""
        correlation_ids = []
        for entry in log_entries:
            if "correlation_id" in entry:
                correlation_ids.append(entry["correlation_id"])
            elif "corr_id" in entry:
                correlation_ids.append(entry["corr_id"])
        return correlation_ids


class TestHealthMonitoringRequirements:
    """Test suite for health monitoring requirements validation."""
    
    @pytest.fixture
    def validator(self):
        """Create health monitoring requirements validator."""
        return HealthMonitoringRequirementsValidator()
    
    @pytest.mark.asyncio
    async def test_req_health_001_comprehensive_logging(self, validator):
        """REQ-HEALTH-001: System shall provide comprehensive logging for health monitoring."""
        await validator.test_req_health_001_comprehensive_logging()
    
    @pytest.mark.asyncio
    async def test_req_health_002_structured_logging(self, validator):
        """REQ-HEALTH-002: System shall support structured logging for production environments."""
        await validator.test_req_health_002_structured_logging()
    
    @pytest.mark.asyncio
    async def test_req_health_003_correlation_id_tracking(self, validator):
        """REQ-HEALTH-003: System shall enable correlation ID tracking across components."""
        await validator.test_req_health_003_correlation_id_tracking()
    
    def test_health_monitoring_metrics_summary(self, validator):
        """Test that all health monitoring requirements are met."""
        # This test validates that all health monitoring metrics meet requirements
        for metric in validator.metrics:
            assert metric.success, f"Health monitoring requirement failed for {metric.requirement}: {metric.error_message}"
