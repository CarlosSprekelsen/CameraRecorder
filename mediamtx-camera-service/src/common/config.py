"""
Configuration management for MediaMTX Camera Service.

Provides YAML configuration loading with environment variable overrides
and comprehensive schema validation as specified in Architecture Decision AD-3.
"""

import os
import yaml
import logging
from typing import Dict, Any, Optional
from dataclasses import dataclass, field


@dataclass
class SecurityConfig:
    """Security configuration settings."""
    
    jwt: Dict[str, Any] = field(default_factory=lambda: {
        "secret_key": "${CAMERA_SERVICE_JWT_SECRET}",
        "expiry_hours": 24,
        "algorithm": "HS256"
    })
    
    api_keys: Dict[str, Any] = field(default_factory=lambda: {
        "storage_file": "${API_KEYS_FILE:/etc/camera-service/api-keys.json}"
    })
    
    ssl: Dict[str, Any] = field(default_factory=lambda: {
        "enabled": False,
        "cert_file": "${SSL_CERT_FILE}",
        "key_file": "${SSL_KEY_FILE}"
    })
    
    rate_limiting: Dict[str, Any] = field(default_factory=lambda: {
        "max_connections": 100,
        "requests_per_minute": 60
    })
    
    health: Dict[str, Any] = field(default_factory=lambda: {
        "port": 8003,
        "bind_address": "0.0.0.0"
    })


@dataclass
class ServerConfig:
    """Server configuration settings."""
    
    host: str = "0.0.0.0"
    port: int = 8002
    websocket_path: str = "/ws"
    max_connections: int = 100


@dataclass
class MediaMTXConfig:
    """MediaMTX configuration settings."""
    
    host: str = "127.0.0.1"
    api_port: int = 9997
    rtsp_port: int = 8554
    webrtc_port: int = 8889
    hls_port: int = 8888
    config_path: str = "/opt/mediamtx/config/mediamtx.yml"
    recordings_path: str = "./.tmp_recordings"
    snapshots_path: str = "./.tmp_snapshots"
    
    # Health monitoring configuration
    health_check_interval: int = 30
    health_failure_threshold: int = 10
    health_circuit_breaker_timeout: int = 60
    health_max_backoff_interval: int = 120
    health_recovery_confirmation_threshold: int = 3
    backoff_base_multiplier: float = 2.0
    backoff_jitter_range: list = field(default_factory=lambda: [0.8, 1.2])
    process_termination_timeout: float = 3.0
    process_kill_timeout: float = 2.0
    
    # Stream readiness configuration for improved reliability
    stream_readiness: Dict[str, Any] = field(default_factory=lambda: {
        "timeout": 15.0,              # Increased from 5.0s to 15.0s for reliability
        "retry_attempts": 3,          # Number of retry attempts
        "retry_delay": 2.0,           # Delay between retries in seconds
        "check_interval": 0.5,        # Interval between readiness checks
        "enable_progress_notifications": True,  # Send progress notifications during validation
        "graceful_fallback": True     # Enable graceful fallback when streams unavailable
    })


@dataclass
class CameraConfig:
    """Camera configuration settings."""
    
    poll_interval: float = 0.1
    detection_timeout: float = 2.0
    device_range: list = field(default_factory=lambda: [0, 9])
    enable_capability_detection: bool = True
    auto_start_streams: bool = True


@dataclass
class LoggingConfig:
    """Logging configuration settings."""
    
    level: str = "INFO"
    format: str = "%(asctime)s - %(name)s - %(levelname)s - %(message)s"
    file_enabled: bool = True
    file_path: str = "./.tmp_logs/camera-service.log"
    max_file_size: str = "10MB"
    backup_count: int = 5


@dataclass
class RecordingConfig:
    """Recording configuration settings."""
    
    auto_record: bool = False
    format: str = "mp4"
    quality: str = "high"
    max_duration: int = 3600
    cleanup_after_days: int = 30


@dataclass
class SnapshotsConfig:
    """Snapshots configuration settings."""
    
    format: str = "jpg"
    quality: int = 90
    cleanup_after_days: int = 7


@dataclass
class FFmpegConfig:
    """FFmpeg configuration settings for performance tuning."""
    
    snapshot: Dict[str, Any] = field(default_factory=lambda: {
        "process_creation_timeout": 5.0,
        "execution_timeout": 8.0,
        "internal_timeout": 5000000,
        "retry_attempts": 2,
        "retry_delay": 1.0
    })
    
    recording: Dict[str, Any] = field(default_factory=lambda: {
        "process_creation_timeout": 10.0,
        "execution_timeout": 15.0,
        "internal_timeout": 10000000,
        "retry_attempts": 3,
        "retry_delay": 2.0
    })


@dataclass
class NotificationConfig:
    """Notification configuration settings for real-time updates."""
    
    websocket: Dict[str, Any] = field(default_factory=lambda: {
        "delivery_timeout": 5.0,
        "retry_attempts": 3,
        "retry_delay": 1.0,
        "max_queue_size": 1000,
        "cleanup_interval": 30
    })
    
    real_time: Dict[str, Any] = field(default_factory=lambda: {
        "camera_status_interval": 1.0,
        "recording_progress_interval": 0.5,
        "connection_health_check": 10.0
    })


@dataclass
class PerformanceConfig:
    """Performance tuning configuration settings."""
    
    response_time_targets: Dict[str, float] = field(default_factory=lambda: {
        "snapshot_capture": 2.0,
        "recording_start": 2.0,
        "recording_stop": 2.0,
        "file_listing": 1.0
    })
    
    # Multi-tier snapshot capture configuration for optimal user experience
    snapshot_tiers: Dict[str, float] = field(default_factory=lambda: {
        # Tier 1: Immediate RTSP capture (when stream is already ready)
        "tier1_rtsp_ready_check_timeout": 1.0,    # seconds - Quick check if RTSP stream is ready
        
        # Tier 2: Quick stream activation (when RTSP needs to be started)
        "tier2_activation_timeout": 3.0,          # seconds - Time to wait for stream activation
        "tier2_activation_trigger_timeout": 1.0,  # seconds - Timeout for triggering activation
        
        # Tier 3: Direct camera capture (fallback when RTSP activation fails)
        "tier3_direct_capture_timeout": 5.0,      # seconds - Timeout for direct camera capture
        
        # Overall snapshot operation timeout
        "total_operation_timeout": 10.0,          # seconds - Maximum total time for snapshot operation
        
        # User experience thresholds
        "immediate_response_threshold": 0.5,      # seconds - Consider response "immediate" if under this
        "acceptable_response_threshold": 2.0,     # seconds - Consider response "acceptable" if under this
        "slow_response_threshold": 5.0            # seconds - Consider response "slow" if over this
    })
    
    optimization: Dict[str, Any] = field(default_factory=lambda: {
        "enable_caching": True,
        "cache_ttl": 300,
        "max_concurrent_operations": 5,
        "connection_pool_size": 10
    })


@dataclass
class Config:
    """Main configuration class."""
    
    server: ServerConfig = field(default_factory=ServerConfig)
    mediamtx: MediaMTXConfig = field(default_factory=MediaMTXConfig)
    camera: CameraConfig = field(default_factory=CameraConfig)
    logging: LoggingConfig = field(default_factory=LoggingConfig)
    recording: RecordingConfig = field(default_factory=RecordingConfig)
    snapshots: SnapshotsConfig = field(default_factory=SnapshotsConfig)
    security: SecurityConfig = field(default_factory=SecurityConfig)
    ffmpeg: FFmpegConfig = field(default_factory=FFmpegConfig)
    notifications: NotificationConfig = field(default_factory=NotificationConfig)
    performance: PerformanceConfig = field(default_factory=PerformanceConfig)


class ConfigManager:
    """
    Configuration manager for MediaMTX Camera Service.
    
    Handles YAML configuration loading, environment variable overrides,
    and schema validation as specified in Architecture Decision AD-3.
    """
    
    def __init__(self, config_path: Optional[str] = None):
        """
        Initialize configuration manager.
        
        Args:
            config_path: Path to configuration file (optional)
        """
        self.config_path = config_path or "/opt/camera-service/config/camera-service.yaml"
        self.logger = logging.getLogger(f"{__name__}.ConfigManager")
        self._config: Optional[Config] = None
        
        self.logger.info("Configuration manager initialized with config path: %s", self.config_path)
    
    def load_config(self) -> Config:
        """
        Load configuration from file and environment variables.
        
        Returns:
            Config object with loaded settings
            
        Raises:
            FileNotFoundError: If configuration file not found
            yaml.YAMLError: If configuration file is invalid
            ValueError: If configuration validation fails
        """
        if self._config is not None:
            return self._config
        
        # Load YAML configuration
        config_data = self._load_yaml_config()
        
        # Apply environment variable overrides
        config_data = self._apply_env_overrides(config_data)
        
        # Validate configuration
        self._validate_config(config_data)
        
        # Create Config object
        self._config = self._create_config_object(config_data)
        
        self.logger.info("Configuration loaded successfully")
        return self._config
    
    def _load_yaml_config(self) -> Dict[str, Any]:
        """Load YAML configuration file."""
        if not os.path.exists(self.config_path):
            self.logger.warning("Configuration file not found: %s", self.config_path)
            return {}
        
        try:
            with open(self.config_path, 'r') as f:
                config_data = yaml.safe_load(f)
            
            if config_data is None:
                config_data = {}
            
            self.logger.debug("Loaded configuration from: %s", self.config_path)
            return config_data
            
        except yaml.YAMLError as e:
            self.logger.error("Invalid YAML configuration: %s", e)
            raise
        except Exception as e:
            self.logger.error("Failed to load configuration: %s", e)
            raise
    
    def _apply_env_overrides(self, config_data: Dict[str, Any]) -> Dict[str, Any]:
        """Apply environment variable overrides to configuration."""
        # Security overrides
        if "security" not in config_data:
            config_data["security"] = {}
        
        security = config_data["security"]
        
        # JWT overrides
        if "jwt" not in security:
            security["jwt"] = {}
        
        jwt_secret = os.getenv("CAMERA_SERVICE_JWT_SECRET")
        if jwt_secret:
            security["jwt"]["secret_key"] = jwt_secret
        
        jwt_expiry = os.getenv("JWT_EXPIRY_HOURS")
        if jwt_expiry:
            try:
                security["jwt"]["expiry_hours"] = int(jwt_expiry)
            except ValueError:
                self.logger.warning("Invalid JWT_EXPIRY_HOURS: %s", jwt_expiry)
        
        # API keys overrides
        if "api_keys" not in security:
            security["api_keys"] = {}
        
        api_keys_file = os.getenv("API_KEYS_FILE")
        if api_keys_file:
            security["api_keys"]["storage_file"] = api_keys_file
        
        # SSL overrides
        if "ssl" not in security:
            security["ssl"] = {}
        
        ssl_enabled = os.getenv("SSL_ENABLED")
        if ssl_enabled:
            security["ssl"]["enabled"] = ssl_enabled.lower() == "true"
        
        ssl_cert = os.getenv("SSL_CERT_FILE")
        if ssl_cert:
            security["ssl"]["cert_file"] = ssl_cert
        
        ssl_key = os.getenv("SSL_KEY_FILE")
        if ssl_key:
            security["ssl"]["key_file"] = ssl_key
        
        # Rate limiting overrides
        if "rate_limiting" not in security:
            security["rate_limiting"] = {}
        
        max_conn = os.getenv("MAX_CONNECTIONS")
        if max_conn:
            try:
                security["rate_limiting"]["max_connections"] = int(max_conn)
            except ValueError:
                self.logger.warning("Invalid MAX_CONNECTIONS: %s", max_conn)
        
        req_per_min = os.getenv("REQUESTS_PER_MINUTE")
        if req_per_min:
            try:
                security["rate_limiting"]["requests_per_minute"] = int(req_per_min)
            except ValueError:
                self.logger.warning("Invalid REQUESTS_PER_MINUTE: %s", req_per_min)
        
        # Health endpoint overrides
        if "health" not in security:
            security["health"] = {}
        
        health_port = os.getenv("HEALTH_PORT")
        if health_port:
            try:
                security["health"]["port"] = int(health_port)
            except ValueError:
                self.logger.warning("Invalid HEALTH_PORT: %s", health_port)
        
        health_bind = os.getenv("HEALTH_BIND_ADDRESS")
        if health_bind:
            security["health"]["bind_address"] = health_bind
        
        # Server overrides
        if "server" not in config_data:
            config_data["server"] = {}
        
        server_host = os.getenv("SERVER_HOST")
        if server_host:
            config_data["server"]["host"] = server_host
        
        server_port = os.getenv("SERVER_PORT")
        if server_port:
            try:
                config_data["server"]["port"] = int(server_port)
            except ValueError:
                self.logger.warning("Invalid SERVER_PORT: %s", server_port)
        
        # MediaMTX overrides
        if "mediamtx" not in config_data:
            config_data["mediamtx"] = {}
        
        mediamtx_host = os.getenv("MEDIAMTX_HOST")
        if mediamtx_host:
            config_data["mediamtx"]["host"] = mediamtx_host
        
        mediamtx_api_port = os.getenv("MEDIAMTX_API_PORT")
        if mediamtx_api_port:
            try:
                config_data["mediamtx"]["api_port"] = int(mediamtx_api_port)
            except ValueError:
                self.logger.warning("Invalid MEDIAMTX_API_PORT: %s", mediamtx_api_port)
        
        return config_data
    
    def _validate_config(self, config_data: Dict[str, Any]) -> None:
        """Validate configuration data."""
        # Validate required security settings
        if "security" in config_data:
            security = config_data["security"]
            
            # Validate JWT settings
            if "jwt" in security:
                jwt = security["jwt"]
                if "secret_key" not in jwt or not jwt["secret_key"]:
                    raise ValueError("JWT secret key must be provided")
                
                if "expiry_hours" in jwt and jwt["expiry_hours"] <= 0:
                    raise ValueError("JWT expiry hours must be positive")
            
            # Validate SSL settings
            if "ssl" in security:
                ssl = security["ssl"]
                if ssl.get("enabled", False):
                    if not ssl.get("cert_file") or not ssl.get("key_file"):
                        raise ValueError("SSL certificate and key files must be provided when SSL is enabled")
        
        # Validate server settings
        if "server" in config_data:
            server = config_data["server"]
            if "port" in server and (server["port"] < 1 or server["port"] > 65535):
                raise ValueError("Server port must be between 1 and 65535")
        
        # Validate MediaMTX settings
        if "mediamtx" in config_data:
            mediamtx = config_data["mediamtx"]
            for port_name in ["api_port", "rtsp_port", "webrtc_port", "hls_port"]:
                if port_name in mediamtx and (mediamtx[port_name] < 1 or mediamtx[port_name] > 65535):
                    raise ValueError(f"MediaMTX {port_name} must be between 1 and 65535")
    
    def _create_config_object(self, config_data: Dict[str, Any]) -> Config:
        """Create Config object from configuration data."""
        config = Config()
        
        # Apply server configuration
        if "server" in config_data:
            server_data = config_data["server"]
            config.server = ServerConfig(
                host=server_data.get("host", config.server.host),
                port=server_data.get("port", config.server.port),
                websocket_path=server_data.get("websocket_path", config.server.websocket_path),
                max_connections=server_data.get("max_connections", config.server.max_connections)
            )
        
        # Apply MediaMTX configuration
        if "mediamtx" in config_data:
            mediamtx_data = config_data["mediamtx"]
            config.mediamtx = MediaMTXConfig(
                host=mediamtx_data.get("host", config.mediamtx.host),
                api_port=mediamtx_data.get("api_port", config.mediamtx.api_port),
                rtsp_port=mediamtx_data.get("rtsp_port", config.mediamtx.rtsp_port),
                webrtc_port=mediamtx_data.get("webrtc_port", config.mediamtx.webrtc_port),
                hls_port=mediamtx_data.get("hls_port", config.mediamtx.hls_port),
                config_path=mediamtx_data.get("config_path", config.mediamtx.config_path),
                recordings_path=mediamtx_data.get("recordings_path", config.mediamtx.recordings_path),
                snapshots_path=mediamtx_data.get("snapshots_path", config.mediamtx.snapshots_path),
                health_check_interval=mediamtx_data.get("health_check_interval", config.mediamtx.health_check_interval),
                health_failure_threshold=mediamtx_data.get("health_failure_threshold", config.mediamtx.health_failure_threshold),
                health_circuit_breaker_timeout=mediamtx_data.get("health_circuit_breaker_timeout", config.mediamtx.health_circuit_breaker_timeout),
                health_max_backoff_interval=mediamtx_data.get("health_max_backoff_interval", config.mediamtx.health_max_backoff_interval),
                health_recovery_confirmation_threshold=mediamtx_data.get("health_recovery_confirmation_threshold", config.mediamtx.health_recovery_confirmation_threshold),
                backoff_base_multiplier=mediamtx_data.get("backoff_base_multiplier", config.mediamtx.backoff_base_multiplier),
                backoff_jitter_range=mediamtx_data.get("backoff_jitter_range", config.mediamtx.backoff_jitter_range),
                process_termination_timeout=mediamtx_data.get("process_termination_timeout", config.mediamtx.process_termination_timeout),
                process_kill_timeout=mediamtx_data.get("process_kill_timeout", config.mediamtx.process_kill_timeout)
            )
        
        # Apply camera configuration
        if "camera" in config_data:
            camera_data = config_data["camera"]
            config.camera = CameraConfig(
                poll_interval=camera_data.get("poll_interval", config.camera.poll_interval),
                detection_timeout=camera_data.get("detection_timeout", config.camera.detection_timeout),
                device_range=camera_data.get("device_range", config.camera.device_range),
                enable_capability_detection=camera_data.get("enable_capability_detection", config.camera.enable_capability_detection),
                auto_start_streams=camera_data.get("auto_start_streams", config.camera.auto_start_streams)
            )
        
        # Apply logging configuration
        if "logging" in config_data:
            logging_data = config_data["logging"]
            config.logging = LoggingConfig(
                level=logging_data.get("level", config.logging.level),
                format=logging_data.get("format", config.logging.format),
                file_enabled=logging_data.get("file_enabled", config.logging.file_enabled),
                file_path=logging_data.get("file_path", config.logging.file_path),
                max_file_size=logging_data.get("max_file_size", config.logging.max_file_size),
                backup_count=logging_data.get("backup_count", config.logging.backup_count)
            )
        
        # Apply recording configuration
        if "recording" in config_data:
            recording_data = config_data["recording"]
            config.recording = RecordingConfig(
                auto_record=recording_data.get("auto_record", config.recording.auto_record),
                format=recording_data.get("format", config.recording.format),
                quality=recording_data.get("quality", config.recording.quality),
                max_duration=recording_data.get("max_duration", config.recording.max_duration),
                cleanup_after_days=recording_data.get("cleanup_after_days", config.recording.cleanup_after_days)
            )
        
        # Apply snapshots configuration
        if "snapshots" in config_data:
            snapshots_data = config_data["snapshots"]
            config.snapshots = SnapshotsConfig(
                format=snapshots_data.get("format", config.snapshots.format),
                quality=snapshots_data.get("quality", config.snapshots.quality),
                cleanup_after_days=snapshots_data.get("cleanup_after_days", config.snapshots.cleanup_after_days)
            )
        
        # Apply security configuration
        if "security" in config_data:
            security_data = config_data["security"]
            config.security = SecurityConfig(
                jwt=security_data.get("jwt", config.security.jwt),
                api_keys=security_data.get("api_keys", config.security.api_keys),
                ssl=security_data.get("ssl", config.security.ssl),
                rate_limiting=security_data.get("rate_limiting", config.security.rate_limiting),
                health=security_data.get("health", config.security.health)
            )
        
        return config
    
    def get_config(self) -> Config:
        """Get current configuration, loading if necessary."""
        if self._config is None:
            return self.load_config()
        return self._config
    
    def reload_config(self) -> Config:
        """Reload configuration from file."""
        self._config = None
        return self.load_config() 