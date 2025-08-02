"""
Configuration management for the camera service.

Enhanced with environment variable overrides, schema validation,
runtime updates, and hot reload capability per architecture overview.
"""

import json
import logging
import os
import threading
import time
import yaml
from dataclasses import dataclass, asdict, fields
from pathlib import Path
from typing import Dict, Any, List, Optional, Callable, Union

# Optional dependencies for full feature support
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
    host: str = "0.0.0.0"
    port: int = 8002
    websocket_path: str = "/ws"
    max_connections: int = 100


@dataclass  
class MediaMTXConfig:
    host: str = "127.0.0.1"
    api_port: int = 9997
    rtsp_port: int = 8554
    webrtc_port: int = 8889
    hls_port: int = 8888
    config_path: str = "/opt/camera-service/config/mediamtx.yml"
    recordings_path: str = "/opt/camera-service/recordings"
    snapshots_path: str = "/opt/camera-service/snapshots"


@dataclass
class CameraConfig:
    poll_interval: float = 0.1
    detection_timeout: float = 2.0
    device_range: List[int] = None
    enable_capability_detection: bool = True
    auto_start_streams: bool = True
    
    def __post_init__(self):
        if self.device_range is None:
            self.device_range = list(range(10))


@dataclass
class LoggingConfig:
    level: str = "INFO"
    format: str = "%(asctime)s - %(name)s - %(levelname)s - %(message)s"
    file_enabled: bool = True
    file_path: str = "/opt/camera-service/logs/camera-service.log"
    max_file_size: str = "10MB"
    backup_count: int = 5


@dataclass
class RecordingConfig:
    auto_record: bool = False
    format: str = "mp4"
    quality: str = "high"
    max_duration: int = 3600
    cleanup_after_days: int = 30


@dataclass
class SnapshotConfig:
    format: str = "jpg"
    quality: int = 90
    cleanup_after_days: int = 7


@dataclass
class Config:
    server: ServerConfig
    mediamtx: MediaMTXConfig
    camera: CameraConfig
    logging: LoggingConfig
    recording: RecordingConfig
    snapshots: SnapshotConfig


class ConfigManager:
    """
    Advanced configuration manager with environment overrides,
    schema validation, runtime updates, and hot reload capability.
    """
    
    def __init__(self):
        self._logger = logging.getLogger(__name__)
        self._config: Optional[Config] = None
        self._config_path: Optional[str] = None
        self._update_callbacks: List[Callable[[Config], None]] = []
        self._observer: Optional[Observer] = None
        self._lock = threading.Lock()
        
    def load_config(self, config_path: str = None) -> Config:
        """
        Load configuration with environment variable overrides and validation.
        
        Args:
            config_path: Path to YAML configuration file
            
        Returns:
            Validated configuration object
            
        Raises:
            FileNotFoundError: If no configuration file found
            ValueError: If configuration validation fails
        """
        with self._lock:
            # Find configuration file
            if config_path is None:
                config_path = self._find_config_file()
            
            self._config_path = config_path
            
            # Load YAML configuration
            yaml_data = self._load_yaml_config(config_path)
            
            # Apply environment variable overrides
            config_data = self._apply_environment_overrides(yaml_data)
            
            # Validate configuration
            self._validate_config(config_data)
            
            # Create configuration object
            self._config = self._create_config_object(config_data)
            
            self._logger.info(f"Configuration loaded successfully from {config_path}")
            return self._config
    
    def update_config(self, updates: Dict[str, Any]) -> Config:
        """
        Update configuration at runtime with validation.
        
        Args:
            updates: Dictionary of configuration updates
            
        Returns:
            Updated configuration object
            
        Raises:
            ValueError: If configuration validation fails
        """
        with self._lock:
            if not self._config:
                raise RuntimeError("Configuration not loaded")
            
            # Create backup of current config
            backup_data = asdict(self._config)
            
            try:
                # Apply updates to current config data
                current_data = asdict(self._config)
                updated_data = self._merge_config_updates(current_data, updates)
                
                # Validate updated configuration
                self._validate_config(updated_data)
                
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
        """
        if not HAS_WATCHDOG:
            self._logger.warning("Hot reload not available - watchdog dependency missing")
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
                
            def on_modified(self, event):
                if not event.is_directory and Path(event.src_path) == Path(self.manager._config_path):
                    self.manager._logger.info("Configuration file changed, reloading...")
                    try:
                        # Delay to ensure file write is complete
                        time.sleep(0.1)
                        self.manager.reload_config()
                    except Exception as e:
                        self.manager._logger.error(f"Hot reload failed: {e}")
        
        self._observer = Observer()
        self._observer.schedule(ConfigFileHandler(self), str(config_dir), recursive=False)
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
        """
        if not self._config_path:
            raise RuntimeError("No configuration file path available for reload")
        
        old_config = self._config
        try:
            new_config = self.load_config(self._config_path)
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
            "/opt/camera-service/config/camera-service.yaml"
        ]
        
        for location in locations:
            if os.path.exists(location):
                return location
        
        raise FileNotFoundError("No configuration file found in standard locations")
    
    def _load_yaml_config(self, config_path: str) -> Dict[str, Any]:
        """Load YAML configuration file."""
        try:
            with open(config_path, 'r') as f:
                return yaml.safe_load(f) or {}
        except Exception as e:
            raise ValueError(f"Failed to load YAML configuration from {config_path}: {e}")
    
    def _apply_environment_overrides(self, config_data: Dict[str, Any]) -> Dict[str, Any]:
        """
        Apply environment variable overrides to configuration.
        
        Environment variables use format: CAMERA_SERVICE_<SECTION>_<SETTING>
        Example: CAMERA_SERVICE_SERVER_PORT=8003 overrides server.port
        """
        # Map of environment variable patterns to config paths
        env_mappings = {
            'CAMERA_SERVICE_SERVER_HOST': ('server', 'host'),
            'CAMERA_SERVICE_SERVER_PORT': ('server', 'port'),
            'CAMERA_SERVICE_SERVER_WEBSOCKET_PATH': ('server', 'websocket_path'),
            'CAMERA_SERVICE_SERVER_MAX_CONNECTIONS': ('server', 'max_connections'),
            
            'CAMERA_SERVICE_MEDIAMTX_HOST': ('mediamtx', 'host'),
            'CAMERA_SERVICE_MEDIAMTX_API_PORT': ('mediamtx', 'api_port'),
            'CAMERA_SERVICE_MEDIAMTX_RTSP_PORT': ('mediamtx', 'rtsp_port'),
            'CAMERA_SERVICE_MEDIAMTX_WEBRTC_PORT': ('mediamtx', 'webrtc_port'),
            'CAMERA_SERVICE_MEDIAMTX_HLS_PORT': ('mediamtx', 'hls_port'),
            'CAMERA_SERVICE_MEDIAMTX_CONFIG_PATH': ('mediamtx', 'config_path'),
            'CAMERA_SERVICE_MEDIAMTX_RECORDINGS_PATH': ('mediamtx', 'recordings_path'),
            'CAMERA_SERVICE_MEDIAMTX_SNAPSHOTS_PATH': ('mediamtx', 'snapshots_path'),
            
            'CAMERA_SERVICE_CAMERA_POLL_INTERVAL': ('camera', 'poll_interval'),
            'CAMERA_SERVICE_CAMERA_DETECTION_TIMEOUT': ('camera', 'detection_timeout'),
            'CAMERA_SERVICE_CAMERA_ENABLE_CAPABILITY_DETECTION': ('camera', 'enable_capability_detection'),
            'CAMERA_SERVICE_CAMERA_AUTO_START_STREAMS': ('camera', 'auto_start_streams'),
            
            'CAMERA_SERVICE_LOGGING_LEVEL': ('logging', 'level'),
            'CAMERA_SERVICE_LOGGING_FILE_ENABLED': ('logging', 'file_enabled'),
            'CAMERA_SERVICE_LOGGING_FILE_PATH': ('logging', 'file_path'),
            
            'CAMERA_SERVICE_RECORDING_AUTO_RECORD': ('recording', 'auto_record'),
            'CAMERA_SERVICE_RECORDING_FORMAT': ('recording', 'format'),
            'CAMERA_SERVICE_RECORDING_QUALITY': ('recording', 'quality'),
            'CAMERA_SERVICE_RECORDING_MAX_DURATION': ('recording', 'max_duration'),
            'CAMERA_SERVICE_RECORDING_CLEANUP_AFTER_DAYS': ('recording', 'cleanup_after_days'),
            
            'CAMERA_SERVICE_SNAPSHOTS_FORMAT': ('snapshots', 'format'),
            'CAMERA_SERVICE_SNAPSHOTS_QUALITY': ('snapshots', 'quality'),
            'CAMERA_SERVICE_SNAPSHOTS_CLEANUP_AFTER_DAYS': ('snapshots', 'cleanup_after_days'),
        }
        
        overridden_count = 0
        
        for env_var, (section, setting) in env_mappings.items():
            if env_var in os.environ:
                value = os.environ[env_var]
                
                # Ensure section exists in config
                if section not in config_data:
                    config_data[section] = {}
                
                # Convert value to appropriate type
                converted_value = self._convert_env_value(value, section, setting)
                config_data[section][setting] = converted_value
                
                overridden_count += 1
                self._logger.debug(f"Environment override: {env_var} -> {section}.{setting} = {converted_value}")
        
        if overridden_count > 0:
            self._logger.info(f"Applied {overridden_count} environment variable overrides")
        
        return config_data
    
    def _convert_env_value(self, value: str, section: str, setting: str) -> Any:
        """Convert environment variable string to appropriate type."""
        # Boolean values
        if setting in ['file_enabled', 'enable_capability_detection', 'auto_start_streams', 'auto_record']:
            return value.lower() in ('true', '1', 'yes', 'on')
        
        # Integer values
        if setting in ['port', 'max_connections', 'api_port', 'rtsp_port', 'webrtc_port', 'hls_port', 
                      'max_duration', 'cleanup_after_days', 'quality', 'backup_count']:
            try:
                return int(value)
            except ValueError:
                raise ValueError(f"Invalid integer value for {section}.{setting}: {value}")
        
        # Float values
        if setting in ['poll_interval', 'detection_timeout']:
            try:
                return float(value)
            except ValueError:
                raise ValueError(f"Invalid float value for {section}.{setting}: {value}")
        
        # String values (default)
        return value
    
    def _validate_config(self, config_data: Dict[str, Any]) -> None:
        """
        Validate configuration data against schema.
        
        Args:
            config_data: Configuration dictionary to validate
            
        Raises:
            ValueError: If configuration is invalid
        """
        if HAS_JSONSCHEMA:
            self._validate_with_jsonschema(config_data)
        else:
            self._validate_basic_schema(config_data)
    
    def _validate_with_jsonschema(self, config_data: Dict[str, Any]) -> None:
        """Validate configuration using JSON Schema."""
        schema = {
            "type": "object",
            "properties": {
                "server": {
                    "type": "object",
                    "properties": {
                        "host": {"type": "string"},
                        "port": {"type": "integer", "minimum": 1, "maximum": 65535},
                        "websocket_path": {"type": "string"},
                        "max_connections": {"type": "integer", "minimum": 1}
                    },
                    "required": ["host", "port"]
                },
                "mediamtx": {
                    "type": "object",
                    "properties": {
                        "host": {"type": "string"},
                        "api_port": {"type": "integer", "minimum": 1, "maximum": 65535},
                        "rtsp_port": {"type": "integer", "minimum": 1, "maximum": 65535},
                        "webrtc_port": {"type": "integer", "minimum": 1, "maximum": 65535},
                        "hls_port": {"type": "integer", "minimum": 1, "maximum": 65535},
                        "config_path": {"type": "string"},
                        "recordings_path": {"type": "string"},
                        "snapshots_path": {"type": "string"}
                    }
                },
                "camera": {
                    "type": "object",
                    "properties": {
                        "poll_interval": {"type": "number", "minimum": 0.01},
                        "detection_timeout": {"type": "number", "minimum": 0.1},
                        "device_range": {
                            "type": "array",
                            "items": {"type": "integer", "minimum": 0}
                        },
                        "enable_capability_detection": {"type": "boolean"},
                        "auto_start_streams": {"type": "boolean"}
                    }
                },
                "logging": {
                    "type": "object",
                    "properties": {
                        "level": {"type": "string", "enum": ["DEBUG", "INFO", "WARNING", "ERROR", "CRITICAL"]},
                        "format": {"type": "string"},
                        "file_enabled": {"type": "boolean"},
                        "file_path": {"type": "string"},
                        "max_file_size": {"type": "string"},
                        "backup_count": {"type": "integer", "minimum": 0}
                    }
                },
                "recording": {
                    "type": "object",
                    "properties": {
                        "auto_record": {"type": "boolean"},
                        "format": {"type": "string", "enum": ["mp4", "mkv", "avi"]},
                        "quality": {"type": "string", "enum": ["low", "medium", "high"]},
                        "max_duration": {"type": "integer", "minimum": 1},
                        "cleanup_after_days": {"type": "integer", "minimum": 0}
                    }
                },
                "snapshots": {
                    "type": "object",
                    "properties": {
                        "format": {"type": "string", "enum": ["jpg", "png", "bmp"]},
                        "quality": {"type": "integer", "minimum": 1, "maximum": 100},
                        "cleanup_after_days": {"type": "integer", "minimum": 0}
                    }
                }
            }
        }
        
        try:
            jsonschema.validate(config_data, schema)
        except jsonschema.ValidationError as e:
            raise ValueError(f"Configuration validation failed: {e.message}")
    
    def _validate_basic_schema(self, config_data: Dict[str, Any]) -> None:
        """Basic configuration validation without jsonschema dependency."""
        # Validate required sections
        required_sections = ['server', 'mediamtx', 'camera', 'logging', 'recording', 'snapshots']
        for section in required_sections:
            if section not in config_data:
                config_data[section] = {}
        
        # Validate server section
        server = config_data.get('server', {})
        if 'port' in server:
            port = server['port']
            if not isinstance(port, int) or port < 1 or port > 65535:
                raise ValueError(f"Invalid server port: {port}")
        
        # Validate MediaMTX ports
        mediamtx = config_data.get('mediamtx', {})
        port_fields = ['api_port', 'rtsp_port', 'webrtc_port', 'hls_port']
        for field in port_fields:
            if field in mediamtx:
                port = mediamtx[field]
                if not isinstance(port, int) or port < 1 or port > 65535:
                    raise ValueError(f"Invalid MediaMTX {field}: {port}")
        
        # Validate camera settings
        camera = config_data.get('camera', {})
        if 'poll_interval' in camera:
            interval = camera['poll_interval']
            if not isinstance(interval, (int, float)) or interval < 0.01:
                raise ValueError(f"Invalid camera poll_interval: {interval}")
        
        # Validate logging level
        logging_config = config_data.get('logging', {})
        if 'level' in logging_config:
            level = logging_config['level']
            valid_levels = ['DEBUG', 'INFO', 'WARNING', 'ERROR', 'CRITICAL']
            if level not in valid_levels:
                raise ValueError(f"Invalid logging level: {level}")
    
    def _create_config_object(self, config_data: Dict[str, Any]) -> Config:
        """Create Config object from validated configuration data."""
        return Config(
            server=ServerConfig(**config_data.get('server', {})),
            mediamtx=MediaMTXConfig(**config_data.get('mediamtx', {})),
            camera=CameraConfig(**config_data.get('camera', {})),
            logging=LoggingConfig(**config_data.get('logging', {})),
            recording=RecordingConfig(**config_data.get('recording', {})),
            snapshots=SnapshotConfig(**config_data.get('snapshots', {}))
        )
    
    def _merge_config_updates(self, current_data: Dict[str, Any], updates: Dict[str, Any]) -> Dict[str, Any]:
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
    
    def _notify_config_updated(self, old_config: Optional[Config], new_config: Config) -> None:
        """Notify all callbacks of configuration update."""
        for callback in self._update_callbacks:
            try:
                callback(new_config)
            except Exception as e:
                self._logger.error(f"Error in configuration update callback: {e}")


# Global configuration manager instance
_config_manager = ConfigManager()


def load_config(config_path: str = None) -> Config:
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