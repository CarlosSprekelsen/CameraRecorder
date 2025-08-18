"""
Configuration management for MediaMTX Camera Service.

Provides configuration loading, validation, environment variable overrides,
and hot reload functionality with comprehensive error handling and fallback behavior.

Key Features:
- YAML file loading with Config.from_file() method
- Dictionary-based configuration with Config.from_dict() method
- Environment variable overrides
- Hot reload capability
- Comprehensive validation and error handling
- Graceful fallback to default values

Usage Examples:
    # Load from YAML file
    config = Config.from_file("config.yml")
    
    # Create from dictionary
    config = Config.from_dict({"server": {"port": 8002}})
    
    # Standard instantiation
    config = Config()
"""

import os
import yaml
import logging
import threading
import time
from dataclasses import dataclass, asdict, field
from typing import Dict, Any, Optional, List, Callable
from pathlib import Path

# Optional dependencies
try:
    import jsonschema

    HAS_JSONSCHEMA = True
except ImportError:
    HAS_JSONSCHEMA = False

try:
    from watchdog.observers import Observer
    from watchdog.events import FileSystemEventHandler

    HAS_WATCHDOG = True
except ImportError:
    HAS_WATCHDOG = False


@dataclass
class ServerConfig:
    """Server configuration settings."""

    host: str = "0.0.0.0"
    port: int = 8002
    websocket_path: str = "/ws"
    max_connections: int = 100


@dataclass
class MediaMTXConfig:
    """MediaMTX integration configuration."""

    host: str = "localhost"
    api_port: int = 9997
    rtsp_port: int = 8554
    webrtc_port: int = 8889
    hls_port: int = 8888
    config_path: str = "/etc/mediamtx/mediamtx.yml"
    recordings_path: str = "/opt/camera-service/recordings"
    snapshots_path: str = "/opt/camera-service/snapshots"
    
    # Health monitoring configuration
    health_check_interval: int = 30
    health_failure_threshold: int = 10
    health_circuit_breaker_timeout: int = 60
    health_max_backoff_interval: int = 120
    health_recovery_confirmation_threshold: int = 3
    backoff_base_multiplier: float = 2.0
    backoff_jitter_range: tuple = (0.8, 1.2)
    process_termination_timeout: float = 3.0
    process_kill_timeout: float = 2.0


@dataclass
class CameraConfig:
    """Camera detection and monitoring configuration."""

    poll_interval: float = 0.1
    detection_timeout: float = 1.0
    device_range: List[int] = field(default_factory=lambda: list(range(10)))
    enable_capability_detection: bool = True
    auto_start_streams: bool = False
    capability_timeout: float = 5.0
    capability_retry_interval: float = 1.0
    capability_max_retries: int = 3


@dataclass
class LoggingConfig:
    """Logging configuration settings."""

    level: str = "INFO"
    format: str = "%(asctime)s - %(name)s - %(levelname)s - %(message)s"
    file_enabled: bool = False
    file_path: str = "/var/log/camera-service/camera-service.log"
    max_file_size: int = 10485760
    backup_count: int = 5
    console_enabled: bool = True


@dataclass
class RecordingConfig:
    """Recording configuration settings."""

    enabled: bool = False
    auto_record: bool = False
    format: str = "fmp4"
    quality: str = "medium"
    segment_duration: int = 3600
    max_segment_size: int = 524288000
    auto_cleanup: bool = True
    cleanup_interval: int = 86400
    max_age: int = 604800
    max_size: int = 10737418240
    max_duration: int = 3600
    cleanup_after_days: int = 30


@dataclass
class SnapshotConfig:
    """Snapshot configuration settings."""

    enabled: bool = True
    format: str = "jpeg"
    quality: int = 85
    max_width: int = 1920
    max_height: int = 1080
    auto_cleanup: bool = True
    cleanup_interval: int = 3600
    max_age: int = 86400
    max_count: int = 1000
    cleanup_after_days: int = 7


@dataclass
class Config:
    """Complete service configuration."""

    server: ServerConfig = field(default_factory=ServerConfig)
    mediamtx: MediaMTXConfig = field(default_factory=MediaMTXConfig)
    camera: CameraConfig = field(default_factory=CameraConfig)
    logging: LoggingConfig = field(default_factory=LoggingConfig)
    recording: RecordingConfig = field(default_factory=RecordingConfig)
    snapshots: SnapshotConfig = field(default_factory=SnapshotConfig)

    def __init__(self, **kwargs):
        """Initialize configuration with proper dataclass conversion."""
        # Create default instances
        self.server = ServerConfig()
        self.mediamtx = MediaMTXConfig()
        self.camera = CameraConfig()
        self.logging = LoggingConfig()
        self.recording = RecordingConfig()
        self.snapshots = SnapshotConfig()
        
        # Update with provided data
        if kwargs:
            self.update_from_dict(kwargs)

    def update_from_dict(self, config_data: Dict[str, Any]) -> None:
        """Update configuration from dictionary data."""
        if "server" in config_data:
            server_data = config_data["server"]
            if isinstance(server_data, dict):
                for key, value in server_data.items():
                    if hasattr(self.server, key):
                        setattr(self.server, key, value)
            elif isinstance(server_data, ServerConfig):
                self.server = server_data

        if "mediamtx" in config_data:
            mediamtx_data = config_data["mediamtx"]
            if isinstance(mediamtx_data, dict):
                for key, value in mediamtx_data.items():
                    if hasattr(self.mediamtx, key):
                        setattr(self.mediamtx, key, value)
            elif isinstance(mediamtx_data, MediaMTXConfig):
                self.mediamtx = mediamtx_data

        if "camera" in config_data:
            camera_data = config_data["camera"]
            if isinstance(camera_data, dict):
                for key, value in camera_data.items():
                    if hasattr(self.camera, key):
                        setattr(self.camera, key, value)
            elif isinstance(camera_data, CameraConfig):
                self.camera = camera_data

        if "logging" in config_data:
            logging_data = config_data["logging"]
            if isinstance(logging_data, dict):
                for key, value in logging_data.items():
                    if hasattr(self.logging, key):
                        setattr(self.logging, key, value)
            elif isinstance(logging_data, LoggingConfig):
                self.logging = logging_data

        if "recording" in config_data:
            recording_data = config_data["recording"]
            if isinstance(recording_data, dict):
                for key, value in recording_data.items():
                    if hasattr(self.recording, key):
                        setattr(self.recording, key, value)
            elif isinstance(recording_data, RecordingConfig):
                self.recording = recording_data

        if "snapshots" in config_data:
            snapshots_data = config_data["snapshots"]
            if isinstance(snapshots_data, dict):
                for key, value in snapshots_data.items():
                    if hasattr(self.snapshots, key):
                        setattr(self.snapshots, key, value)
            elif isinstance(snapshots_data, SnapshotConfig):
                self.snapshots = snapshots_data

    def to_dict(self) -> Dict[str, Any]:
        """Convert configuration to dictionary for serialization."""
        try:
            return {
                "server": asdict(self.server),
                "mediamtx": asdict(self.mediamtx),
                "camera": asdict(self.camera),
                "logging": asdict(self.logging),
                "recording": asdict(self.recording),
                "snapshots": asdict(self.snapshots),
            }
        except Exception as e:
            # Fallback to manual conversion if asdict fails
            def _to_dict(obj):
                if hasattr(obj, '__dataclass_fields__'):
                    return {k: getattr(obj, k) for k in obj.__dataclass_fields__}
                elif isinstance(obj, dict):
                    return obj
                else:
                    return obj.__dict__ if hasattr(obj, '__dict__') else str(obj)
            
            return {
                "server": _to_dict(self.server),
                "mediamtx": _to_dict(self.mediamtx),
                "camera": _to_dict(self.camera),
                "logging": _to_dict(self.logging),
                "recording": _to_dict(self.recording),
                "snapshots": _to_dict(self.snapshots),
            }

    @classmethod
    def from_file(cls, file_path: str) -> 'Config':
        """
        Load configuration from YAML file with comprehensive error handling.
        
        Args:
            file_path: Path to YAML configuration file
            
        Returns:
            Config object loaded from file
            
        Raises:
            FileNotFoundError: If configuration file does not exist
            yaml.YAMLError: If YAML file is malformed
            ValueError: If configuration validation fails
        """
        if not os.path.exists(file_path):
            raise FileNotFoundError(f"Configuration file not found: {file_path}")
        
        try:
            with open(file_path, 'r', encoding='utf-8') as f:
                config_data = yaml.safe_load(f)
            
            if config_data is None:
                # Empty or invalid YAML file
                raise ValueError(f"Configuration file is empty or invalid: {file_path}")
            
            # Create config instance and update from loaded data
            config = cls()
            config.update_from_dict(config_data)
            
            return config
            
        except yaml.YAMLError as e:
            raise ValueError(f"Invalid YAML configuration in {file_path}: {e}")
        except Exception as e:
            if "configuration" in str(e).lower() or "invalid" in str(e).lower():
                raise
            else:
                raise ValueError(f"Error loading configuration from {file_path}: {e}")

    @classmethod
    def from_dict(cls, config_data: Dict[str, Any]) -> 'Config':
        """
        Create configuration from dictionary data.
        
        Args:
            config_data: Dictionary containing configuration data
            
        Returns:
            Config object created from dictionary
        """
        config = cls()
        config.update_from_dict(config_data)
        return config


class ConfigManager:
    """
    Configuration manager with environment overrides, validation, and hot reload.

    Provides robust configuration loading with graceful fallback behavior,
    comprehensive validation, and safe hot reload functionality.
    """

    def __init__(self):
        self._logger = logging.getLogger(__name__)
        self._config: Optional[Config] = None
        self._config_path: Optional[str] = None
        self._update_callbacks: List[Callable[[Config], None]] = []
        self._observer: Optional[Observer] = None
        self._lock = threading.Lock()
        self._default_config = Config()  # Fallback configuration

    def load_config(self, config_path: Optional[str] = None) -> Config:
        """
        Load configuration with environment variable overrides and validation.

        Handles missing or malformed configuration files gracefully by falling back
        to default configuration values. Invalid environment overrides are logged
        but do not crash the service.

        Args:
            config_path: Path to YAML configuration file

        Returns:
            Validated configuration object

        Raises:
            ValueError: If configuration validation fails after all fallbacks
        """
        with self._lock:
            config_data = {}

            # Try to find and load configuration file
            if config_path is None:
                try:
                    config_path = self._find_config_file()
                    self._config_path = config_path
                except FileNotFoundError:
                    self._logger.warning(
                        "No configuration file found in standard locations, using defaults"
                    )
                    config_path = None
            else:
                self._config_path = config_path

            # Load YAML configuration with fallback
            if config_path:
                config_data = self._load_yaml_config_safe(config_path)
            else:
                self._logger.info("Using default configuration")

            # Apply environment variable overrides (with error tolerance)
            config_data = self._apply_environment_overrides_safe(config_data)

            # Ensure all required sections exist with defaults
            config_data = self._ensure_complete_config(config_data)

            # Validate configuration with comprehensive error reporting
            validation_errors = self._validate_config_comprehensive(config_data)
            if validation_errors:
                error_msg = "Configuration validation failed:\n" + "\n".join(
                    validation_errors
                )
                self._logger.error(error_msg)
                raise ValueError(error_msg)

            # Create configuration object
            self._config = self._create_config_object(config_data)

            self._logger.info(
                f"Configuration loaded successfully from {config_path or 'defaults'}"
            )
            return self._config

    def update_config(self, updates: Dict[str, Any]) -> Config:
        """
        Update configuration at runtime with validation and safe rollback.

        Args:
            updates: Dictionary of configuration updates

        Returns:
            Updated configuration object

        Raises:
            ValueError: If configuration validation fails
            RuntimeError: If no configuration is currently loaded
        """
        with self._lock:
            if not self._config:
                raise RuntimeError("Configuration not loaded")

            # Create backup of current config for rollback
            backup_data = asdict(self._config)

            try:
                # Apply updates to current config data
                current_data = asdict(self._config)
                updated_data = self._merge_config_updates(current_data, updates)

                # Validate updated configuration
                validation_errors = self._validate_config_comprehensive(updated_data)
                if validation_errors:
                    error_msg = "Configuration update validation failed:\n" + "\n".join(
                        validation_errors
                    )
                    raise ValueError(error_msg)

                # Create new configuration object
                new_config = self._create_config_object(updated_data)

                # Update current configuration
                old_config = self._config
                self._config = new_config

                # Notify callbacks of configuration change
                self._notify_config_updated(old_config, new_config)

                self._logger.info("Configuration updated successfully")
                return self._config

            except Exception as e:
                # Rollback to backup configuration on failure
                self._config = self._create_config_object(backup_data)
                self._logger.error(f"Configuration update failed, rolled back: {e}")
                raise ValueError(f"Configuration update failed: {e}")

    def start_hot_reload(self) -> None:
        """
        Start hot reload monitoring for configuration file changes.

        Uses file system monitoring to detect changes and reload configuration
        with proper validation and rollback on failure.
        """
        if not HAS_WATCHDOG:
            self._logger.warning(
                "Hot reload not available - watchdog dependency missing"
            )
            return

        if not self._config_path:
            self._logger.warning("Hot reload not started - no configuration file path")
            return

        if self._observer:
            self._logger.warning("Hot reload already started")
            return

        config_dir = Path(self._config_path).parent

        class ConfigFileHandler(FileSystemEventHandler):
            def __init__(self, manager: ConfigManager):
                self.manager = manager
                self._last_reload_time = 0

            def on_modified(self, event):
                if not event.is_directory and Path(event.src_path) == Path(
                    self.manager._config_path
                ):
                    # Debounce rapid file changes
                    current_time = time.time()
                    if current_time - self._last_reload_time < 1.0:
                        return
                    self._last_reload_time = current_time

                    self.manager._logger.info(
                        "Configuration file changed, reloading..."
                    )
                    try:
                        # Wait for file write completion
                        self._wait_for_file_stable()
                        self.manager.reload_config()
                    except Exception as e:
                        self.manager._logger.error(f"Hot reload failed: {e}")

            def _wait_for_file_stable(self):
                """Wait for file to be stable (no size changes)."""
                config_path = Path(self.manager._config_path)
                if not config_path.exists():
                    return

                last_size = -1
                stable_checks = 0
                max_wait = 10  # Maximum 1 second wait

                while stable_checks < 5 and max_wait > 0:
                    try:
                        current_size = config_path.stat().st_size
                        if current_size == last_size:
                            stable_checks += 1
                        else:
                            stable_checks = 0
                            last_size = current_size

                        time.sleep(0.1)
                        max_wait -= 1
                    except OSError:
                        # File might be temporarily unavailable
                        time.sleep(0.1)
                        max_wait -= 1

        self._observer = Observer()
        self._observer.schedule(
            ConfigFileHandler(self), str(config_dir), recursive=False
        )
        self._observer.start()

        self._logger.info(f"Hot reload monitoring started for {self._config_path}")

    def stop_hot_reload(self) -> None:
        """Stop hot reload monitoring."""
        if self._observer:
            self._observer.stop()
            self._observer.join()
            self._observer = None
            self._logger.info("Hot reload monitoring stopped")

    def reload_config(self) -> Config:
        """
        Reload configuration from file with validation and rollback.

        Returns:
            Reloaded configuration object

        Raises:
            RuntimeError: If no configuration file path is available
        """
        if not self._config_path:
            raise RuntimeError("No configuration file path available for reload")

        old_config = self._config
        try:
            new_config = self.load_config(self._config_path)
            if old_config:
                self._notify_config_updated(old_config, new_config)
            return new_config
        except Exception as e:
            self._logger.error(f"Configuration reload failed: {e}")
            raise

    def add_update_callback(self, callback: Callable[[Config], None]) -> None:
        """
        Add callback to be notified of configuration updates.

        Args:
            callback: Function to call when configuration changes
        """
        self._update_callbacks.append(callback)

    def remove_update_callback(self, callback: Callable[[Config], None]) -> None:
        """Remove configuration update callback."""
        if callback in self._update_callbacks:
            self._update_callbacks.remove(callback)

    def get_config(self) -> Optional[Config]:
        """Get current configuration object."""
        return self._config

    def _find_config_file(self) -> str:
        """Find configuration file in standard locations."""
        locations = [
            "config/default.yaml",
            "/etc/camera-service/config.yaml",
            "/opt/camera-service/config/camera-service.yaml",
        ]

        for location in locations:
            if os.path.exists(location):
                return location

        raise FileNotFoundError("No configuration file found in standard locations")

    def _load_yaml_config_safe(self, config_path: str) -> Dict[str, Any]:
        """
        Load YAML configuration file with error handling and fallback.

        Args:
            config_path: Path to configuration file

        Returns:
            Configuration dictionary (may be empty on errors)
        """
        if not os.path.exists(config_path):
            self._logger.warning(
                f"Configuration file not found: {config_path}, using defaults"
            )
            return {}

        try:
            with open(config_path, "r") as f:
                content = f.read().strip()
                if not content:
                    self._logger.warning(
                        f"Configuration file is empty: {config_path}, using defaults"
                    )
                    return {}

                data = yaml.safe_load(content)
                if data is None:
                    self._logger.warning(
                        f"Configuration file contains no data: {config_path}, using defaults"
                    )
                    return {}

                return data if isinstance(data, dict) else {}
        except yaml.YAMLError as e:
            self._logger.error(f"Malformed YAML in {config_path}: {e}, using defaults")
            return {}
        except Exception as e:
            self._logger.error(
                f"Failed to load configuration from {config_path}: {e}, using defaults"
            )
            return {}

    def _apply_environment_overrides_safe(
        self, config_data: Dict[str, Any]
    ) -> Dict[str, Any]:
        """
        Apply environment variable overrides with error tolerance.

        Invalid environment variables are logged but do not crash the service.

        Args:
            config_data: Base configuration data

        Returns:
            Configuration data with valid environment overrides applied
        """
        # Map of environment variable patterns to config paths
        env_mappings = {
            "CAMERA_SERVICE_SERVER_HOST": ("server", "host"),
            "CAMERA_SERVICE_SERVER_PORT": ("server", "port"),
            "CAMERA_SERVICE_SERVER_WEBSOCKET_PATH": ("server", "websocket_path"),
            "CAMERA_SERVICE_SERVER_MAX_CONNECTIONS": ("server", "max_connections"),
            "CAMERA_SERVICE_MEDIAMTX_HOST": ("mediamtx", "host"),
            "CAMERA_SERVICE_MEDIAMTX_API_PORT": ("mediamtx", "api_port"),
            "CAMERA_SERVICE_MEDIAMTX_RTSP_PORT": ("mediamtx", "rtsp_port"),
            "CAMERA_SERVICE_MEDIAMTX_WEBRTC_PORT": ("mediamtx", "webrtc_port"),
            "CAMERA_SERVICE_MEDIAMTX_HLS_PORT": ("mediamtx", "hls_port"),
            "CAMERA_SERVICE_MEDIAMTX_CONFIG_PATH": ("mediamtx", "config_path"),
            "CAMERA_SERVICE_MEDIAMTX_RECORDINGS_PATH": ("mediamtx", "recordings_path"),
            "CAMERA_SERVICE_MEDIAMTX_SNAPSHOTS_PATH": ("mediamtx", "snapshots_path"),
            "CAMERA_SERVICE_CAMERA_POLL_INTERVAL": ("camera", "poll_interval"),
            "CAMERA_SERVICE_CAMERA_DETECTION_TIMEOUT": ("camera", "detection_timeout"),
            "CAMERA_SERVICE_CAMERA_ENABLE_CAPABILITY_DETECTION": (
                "camera",
                "enable_capability_detection",
            ),
            "CAMERA_SERVICE_CAMERA_AUTO_START_STREAMS": (
                "camera",
                "auto_start_streams",
            ),
            "CAMERA_SERVICE_LOGGING_LEVEL": ("logging", "level"),
            "CAMERA_SERVICE_LOGGING_FORMAT": ("logging", "format"),
            "CAMERA_SERVICE_LOGGING_FILE_ENABLED": ("logging", "file_enabled"),
            "CAMERA_SERVICE_LOGGING_FILE_PATH": ("logging", "file_path"),
            "CAMERA_SERVICE_LOGGING_MAX_FILE_SIZE": ("logging", "max_file_size"),
            "CAMERA_SERVICE_LOGGING_BACKUP_COUNT": ("logging", "backup_count"),
            "CAMERA_SERVICE_RECORDING_AUTO_RECORD": ("recording", "auto_record"),
            "CAMERA_SERVICE_RECORDING_FORMAT": ("recording", "format"),
            "CAMERA_SERVICE_RECORDING_QUALITY": ("recording", "quality"),
            "CAMERA_SERVICE_RECORDING_MAX_DURATION": ("recording", "max_duration"),
            "CAMERA_SERVICE_RECORDING_CLEANUP_AFTER_DAYS": (
                "recording",
                "cleanup_after_days",
            ),
            "CAMERA_SERVICE_SNAPSHOTS_FORMAT": ("snapshots", "format"),
            "CAMERA_SERVICE_SNAPSHOTS_QUALITY": ("snapshots", "quality"),
            "CAMERA_SERVICE_SNAPSHOTS_CLEANUP_AFTER_DAYS": (
                "snapshots",
                "cleanup_after_days",
            ),
        }

        overridden_count = 0
        failed_overrides = []

        for env_var, (section, setting) in env_mappings.items():
            if env_var in os.environ:
                env_value = os.environ[env_var]
                try:
                    converted_value = self._convert_env_value_safe(
                        env_value, section, setting
                    )

                    # Ensure section exists
                    if section not in config_data:
                        config_data[section] = {}

                    config_data[section][setting] = converted_value
                    overridden_count += 1

                    self._logger.debug(
                        f"Applied environment override: {section}.{setting} = {converted_value}"
                    )

                except ValueError as e:
                    failed_overrides.append(f"{env_var}: {e}")
                    self._logger.error(
                        f"Invalid environment variable {env_var}: {e}, using default"
                    )

        if overridden_count > 0:
            self._logger.info(
                f"Applied {overridden_count} environment variable overrides"
            )

        if failed_overrides:
            self._logger.warning(
                f"Ignored {len(failed_overrides)} invalid environment overrides"
            )

        return config_data

    def _convert_env_value_safe(self, value: str, section: str, setting: str) -> Any:
        """
        Convert environment variable string to appropriate type with error handling.

        Args:
            value: Environment variable value
            section: Configuration section name
            setting: Configuration setting name

        Returns:
            Converted value

        Raises:
            ValueError: If conversion fails
        """
        # Boolean values
        if setting in [
            "file_enabled",
            "enable_capability_detection",
            "auto_start_streams",
            "auto_record",
        ]:
            return value.lower() in ("true", "1", "yes", "on")

        # Integer values with validation
        if setting in [
            "port",
            "max_connections",
            "api_port",
            "rtsp_port",
            "webrtc_port",
            "hls_port",
            "max_duration",
            "cleanup_after_days",
            "quality",
            "backup_count",
        ]:
            try:
                int_value = int(value)
                # Additional validation for specific fields
                if setting.endswith("_port") and (int_value < 1 or int_value > 65535):
                    raise ValueError(
                        f"Port must be between 1 and 65535, got {int_value}"
                    )
                if (
                    setting
                    in [
                        "max_connections",
                        "max_duration",
                        "cleanup_after_days",
                        "backup_count",
                    ]
                    and int_value < 0
                ):
                    raise ValueError(f"Value must be non-negative, got {int_value}")
                if setting == "quality" and (int_value < 1 or int_value > 100):
                    raise ValueError(
                        f"Quality must be between 1 and 100, got {int_value}"
                    )
                return int_value
            except (ValueError, TypeError) as e:
                raise ValueError(
                    f"Invalid integer value for {section}.{setting}: {value} ({e})"
                )

        # Float values with validation
        if setting in ["poll_interval", "detection_timeout"]:
            try:
                float_value = float(value)
                if setting == "poll_interval" and float_value < 0.01:
                    raise ValueError(
                        f"Poll interval must be at least 0.01 seconds, got {float_value}"
                    )
                if setting == "detection_timeout" and float_value < 0.1:
                    raise ValueError(
                        f"Detection timeout must be at least 0.1 seconds, got {float_value}"
                    )
                return float_value
            except (ValueError, TypeError) as e:
                raise ValueError(
                    f"Invalid float value for {section}.{setting}: {value} ({e})"
                )

        # String values with validation
        if setting == "level":
            valid_levels = ["DEBUG", "INFO", "WARNING", "ERROR", "CRITICAL"]
            if value not in valid_levels:
                raise ValueError(
                    f"Invalid logging level: {value}, must be one of {valid_levels}"
                )

        return value

    def _ensure_complete_config(self, config_data: Dict[str, Any]) -> Dict[str, Any]:
        """
        Ensure all required configuration sections exist with default values.

        Args:
            config_data: Partial configuration data

        Returns:
            Complete configuration data with defaults filled in
        """
        default_data = asdict(self._default_config)

        # Merge with defaults, preserving existing values
        for section_name, section_defaults in default_data.items():
            if section_name not in config_data:
                config_data[section_name] = section_defaults.copy()
            else:
                # Fill in missing keys within sections
                for key, default_value in section_defaults.items():
                    if key not in config_data[section_name]:
                        config_data[section_name][key] = default_value

        return config_data

    def _validate_config_comprehensive(self, config_data: Dict[str, Any]) -> List[str]:
        """
        Comprehensive configuration validation with error accumulation.

        Args:
            config_data: Configuration dictionary to validate

        Returns:
            List of validation error messages (empty if valid)
        """
        validation_errors = []

        if HAS_JSONSCHEMA:
            try:
                self._validate_with_jsonschema(config_data)
            except ValueError as e:
                validation_errors.append(str(e))
        else:
            validation_errors.extend(
                self._validate_basic_schema_comprehensive(config_data)
            )

        return validation_errors

    def _validate_with_jsonschema(self, config_data: Dict[str, Any]) -> None:
        """Validate configuration using JSON Schema."""
        schema = {
            "type": "object",
            "properties": {
                "server": {
                    "type": "object",
                    "properties": {
                        "host": {"type": "string", "minLength": 1},
                        "port": {"type": "integer", "minimum": 1, "maximum": 65535},
                        "websocket_path": {"type": "string", "minLength": 1},
                        "max_connections": {"type": "integer", "minimum": 1},
                    },
                    "required": ["host", "port"],
                },
                "mediamtx": {
                    "type": "object",
                    "properties": {
                        "host": {"type": "string", "minLength": 1},
                        "api_port": {"type": "integer", "minimum": 1, "maximum": 65535},
                        "rtsp_port": {
                            "type": "integer",
                            "minimum": 1,
                            "maximum": 65535,
                        },
                        "webrtc_port": {
                            "type": "integer",
                            "minimum": 1,
                            "maximum": 65535,
                        },
                        "hls_port": {"type": "integer", "minimum": 1, "maximum": 65535},
                        "config_path": {"type": "string", "minLength": 1},
                        "recordings_path": {"type": "string", "minLength": 1},
                        "snapshots_path": {"type": "string", "minLength": 1},
                    },
                },
                "camera": {
                    "type": "object",
                    "properties": {
                        "poll_interval": {"type": "number", "minimum": 0.01},
                        "detection_timeout": {"type": "number", "minimum": 0.1},
                        "device_range": {
                            "type": "array",
                            "items": {"type": "integer", "minimum": 0},
                            "maxItems": 100,
                        },
                        "enable_capability_detection": {"type": "boolean"},
                        "auto_start_streams": {"type": "boolean"},
                    },
                },
                "logging": {
                    "type": "object",
                    "properties": {
                        "level": {
                            "type": "string",
                            "enum": ["DEBUG", "INFO", "WARNING", "ERROR", "CRITICAL"],
                        },
                        "format": {"type": "string", "minLength": 1},
                        "file_enabled": {"type": "boolean"},
                        "file_path": {"type": "string", "minLength": 1},
                        "max_file_size": {"type": "integer", "minimum": 1},
                        "backup_count": {"type": "integer", "minimum": 0},
                    },
                },
                "recording": {
                    "type": "object",
                    "properties": {
                        "enabled": {"type": "boolean"},
                        "auto_record": {"type": "boolean"},
                        "format": {"type": "string", "enum": ["mp4", "fmp4", "mkv", "avi"]},
                        "quality": {
                            "type": "string",
                            "enum": ["low", "medium", "high"],
                        },
                        "segment_duration": {"type": "integer", "minimum": 1},
                        "max_segment_size": {"type": "integer", "minimum": 1},
                        "auto_cleanup": {"type": "boolean"},
                        "cleanup_interval": {"type": "integer", "minimum": 1},
                        "max_age": {"type": "integer", "minimum": 1},
                        "max_size": {"type": "integer", "minimum": 1},
                        "max_duration": {"type": "integer", "minimum": 1},
                        "cleanup_after_days": {"type": "integer", "minimum": 0},
                    },
                },
                "snapshots": {
                    "type": "object",
                    "properties": {
                        "format": {"type": "string", "enum": ["jpg", "jpeg", "png", "bmp"]},
                        "quality": {"type": "integer", "minimum": 1, "maximum": 100},
                        "cleanup_after_days": {"type": "integer", "minimum": 0},
                    },
                },
            },
        }

        try:
            jsonschema.validate(config_data, schema)
        except jsonschema.ValidationError as e:
            raise ValueError(f"Configuration validation failed: {e.message}")

    def _validate_basic_schema_comprehensive(
        self, config_data: Dict[str, Any]
    ) -> List[str]:
        """
        Basic configuration validation without jsonschema dependency.

        Args:
            config_data: Configuration dictionary to validate

        Returns:
            List of validation error messages
        """
        errors = []

        # Validate server section
        server = config_data.get("server", {})
        if "port" in server:
            port = server["port"]
            if not isinstance(port, int) or port < 1 or port > 65535:
                errors.append(f"Invalid server port: {port} (must be integer 1-65535)")

        if "max_connections" in server:
            max_conn = server["max_connections"]
            if not isinstance(max_conn, int) or max_conn < 1:
                errors.append(
                    f"Invalid max_connections: {max_conn} (must be positive integer)"
                )

        # Validate MediaMTX ports
        mediamtx = config_data.get("mediamtx", {})
        port_fields = ["api_port", "rtsp_port", "webrtc_port", "hls_port"]
        for port_field in port_fields:  # <-- FIXED: renamed to avoid shadowing
            if port_field in mediamtx:
                port = mediamtx[port_field]
                if not isinstance(port, int) or port < 1 or port > 65535:
                    errors.append(
                        f"Invalid MediaMTX {port_field}: {port} (must be integer 1-65535)"
                    )

        # Validate camera settings
        camera = config_data.get("camera", {})
        if "poll_interval" in camera:
            interval = camera["poll_interval"]
            if not isinstance(interval, (int, float)) or interval < 0.01:
                errors.append(
                    f"Invalid camera poll_interval: {interval} (must be >= 0.01)"
                )

        if "detection_timeout" in camera:
            timeout = camera["detection_timeout"]
            if not isinstance(timeout, (int, float)) or timeout < 0.1:
                errors.append(
                    f"Invalid camera detection_timeout: {timeout} (must be >= 0.1)"
                )

        # Validate logging level
        logging_config = config_data.get("logging", {})
        if "level" in logging_config:
            level = logging_config["level"]
            valid_levels = ["DEBUG", "INFO", "WARNING", "ERROR", "CRITICAL"]
            if level not in valid_levels:
                errors.append(
                    f"Invalid logging level: {level} (must be one of {valid_levels})"
                )

        # Validate recording settings
        recording = config_data.get("recording", {})
        if "format" in recording:
            format_val = recording["format"]
            valid_formats = ["mp4", "fmp4", "mkv", "avi"]
            if format_val not in valid_formats:
                errors.append(
                    f"Invalid recording format: {format_val} (must be one of {valid_formats})"
                )

        if "quality" in recording:
            quality = recording["quality"]
            valid_qualities = ["low", "medium", "high"]
            if quality not in valid_qualities:
                errors.append(
                    f"Invalid recording quality: {quality} (must be one of {valid_qualities})"
                )

        # Validate snapshot settings
        snapshots = config_data.get("snapshots", {})
        if "format" in snapshots:
            format_val = snapshots["format"]
            valid_formats = ["jpg", "jpeg", "png", "bmp"]
            if format_val not in valid_formats:
                errors.append(
                    f"Invalid snapshot format: {format_val} (must be one of {valid_formats})"
                )

        if "quality" in snapshots:
            quality = snapshots["quality"]
            if not isinstance(quality, int) or quality < 1 or quality > 100:
                errors.append(
                    f"Invalid snapshot quality: {quality} (must be integer 1-100)"
                )

        return errors

    def _create_config_object(self, config_data: Dict[str, Any]) -> Config:
        """Create Config object from validated configuration data."""
        # Convert backoff_jitter_range from list to tuple if needed
        mediamtx_data = config_data.get("mediamtx", {}).copy()
        if "backoff_jitter_range" in mediamtx_data and isinstance(mediamtx_data["backoff_jitter_range"], list):
            mediamtx_data["backoff_jitter_range"] = tuple(mediamtx_data["backoff_jitter_range"])
        
        return Config(
            server=ServerConfig(**config_data.get("server", {})),
            mediamtx=MediaMTXConfig(**mediamtx_data),
            camera=CameraConfig(**config_data.get("camera", {})),
            logging=LoggingConfig(**config_data.get("logging", {})),
            recording=RecordingConfig(**config_data.get("recording", {})),
            snapshots=SnapshotConfig(**config_data.get("snapshots", {})),
        )

    def _merge_config_updates(
        self, current_data: Dict[str, Any], updates: Dict[str, Any]
    ) -> Dict[str, Any]:
        """Merge configuration updates into current configuration data."""
        merged = current_data.copy()

        for section, section_updates in updates.items():
            if section not in merged:
                merged[section] = {}

            if isinstance(section_updates, dict):
                merged[section].update(section_updates)
            else:
                merged[section] = section_updates

        return merged

    def _notify_config_updated(
        self, old_config: Optional[Config], new_config: Config
    ) -> None:
        """Notify all callbacks of configuration update."""
        for callback in self._update_callbacks:
            try:
                callback(new_config)
            except Exception as e:
                self._logger.error(f"Error in configuration update callback: {e}")


# Global configuration manager instance
_config_manager = ConfigManager()


def load_config(config_path: Optional[str] = None) -> Config:
    """
    Load configuration from YAML file with environment overrides.

    Args:
        config_path: Path to YAML configuration file

    Returns:
        Configuration object with all settings
    """
    return _config_manager.load_config(config_path)


def get_config_manager() -> ConfigManager:
    """Get the global configuration manager instance."""
    return _config_manager


def get_current_config() -> Optional[Config]:
    """Get the current configuration object."""
    return _config_manager.get_config()
