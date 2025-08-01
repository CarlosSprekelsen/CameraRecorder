"""
Configuration management for the camera service.
"""

import os
import yaml
from dataclasses import dataclass
from pathlib import Path
from typing import Dict, Any, List


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


def load_config(config_path: str = None) -> Config:
    """Load configuration from YAML file."""
    if config_path is None:
        # Try different locations
        locations = [
            "config/default.yaml",
            "/etc/camera-service/config.yaml",
            "/opt/camera-service/config/camera-service.yaml"
        ]
        
        for location in locations:
            if os.path.exists(location):
                config_path = location
                break
        else:
            raise FileNotFoundError("No configuration file found")
    
    with open(config_path, 'r') as f:
        data = yaml.safe_load(f)
    
    return Config(
        server=ServerConfig(**data.get('server', {})),
        mediamtx=MediaMTXConfig(**data.get('mediamtx', {})),
        camera=CameraConfig(**data.get('camera', {})),
        logging=LoggingConfig(**data.get('logging', {})),  
        recording=RecordingConfig(**data.get('recording', {})),
        snapshots=SnapshotConfig(**data.get('snapshots', {}))
    )
