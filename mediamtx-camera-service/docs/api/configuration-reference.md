# Configuration Reference

**Version:** 1.0  
**Last Updated:** 2025-01-15

## Overview

This document describes the configuration structure and parameters for the MediaMTX Camera Service, including the MediaMTXConfig dataclass and its STANAG 4406 H.264 codec parameters.

## MediaMTXConfig Dataclass

The `MediaMTXConfig` dataclass in `src/camera_service/config.py` defines the configuration for MediaMTX integration and STANAG 4406 H.264 codec compliance.

### Core MediaMTX Configuration

```python
@dataclass
class MediaMTXConfig:
    # MediaMTX server configuration
    host: str = "127.0.0.1"
    api_port: int = 9997
    rtsp_port: int = 8554
    webrtc_port: int = 8889
    hls_port: int = 8888
    
    # File system paths
    config_path: str = "/etc/mediamtx/mediamtx.yml"
    recordings_path: str = "/opt/camera-service/recordings"
    snapshots_path: str = "/opt/camera-service/snapshots"
```

### STANAG 4406 H.264 Codec Configuration

The following parameters ensure STANAG 4406 (MIL-STD-188-110B) H.264 compliance:

```python
@dataclass
class MediaMTXConfig:
    # ... core configuration ...
    
    # STANAG 4406 H.264 codec configuration
    codec: str = "libx264"  # H.264 codec for STANAG 4406 compliance
    video_profile: str = "baseline"  # Baseline profile for STANAG 4406
    video_level: str = "3.0"  # Level 3.0 for STANAG 4406
    pixel_format: str = "yuv420p"  # 4:2:0 pixel format for STANAG 4406
    bitrate: str = "600k"  # STANAG 4406 compatible bitrate
    preset: str = "ultrafast"  # Encoding preset
```

### Parameter Details

#### codec
- **Type:** `str`
- **Default:** `"libx264"`
- **Description:** H.264 video codec for STANAG 4406 compliance
- **Options:** `"libx264"` (recommended for STANAG 4406)

#### video_profile
- **Type:** `str`
- **Default:** `"baseline"`
- **Description:** H.264 profile for STANAG 4406 compliance
- **Options:** `"baseline"`, `"main"`, `"high"`
- **STANAG 4406:** Use `"baseline"` for maximum compatibility

#### video_level
- **Type:** `str`
- **Default:** `"3.0"`
- **Description:** H.264 level for STANAG 4406 compliance
- **Options:** `"1.0"`, `"1.1"`, `"1.2"`, `"1.3"`, `"2.0"`, `"2.1"`, `"2.2"`, `"3.0"`, `"3.1"`, `"3.2"`, `"4.0"`, `"4.1"`, `"4.2"`, `"5.0"`, `"5.1"`, `"5.2"`
- **STANAG 4406:** Use `"3.0"` for up to 720p resolution support

#### pixel_format
- **Type:** `str`
- **Default:** `"yuv420p"`
- **Description:** Pixel format for STANAG 4406 compliance
- **Options:** `"yuv420p"`, `"yuv422p"`, `"yuv444p"`
- **STANAG 4406:** Use `"yuv420p"` (4:2:0) for maximum compatibility

#### bitrate
- **Type:** `str`
- **Default:** `"600k"`
- **Description:** Video bitrate for STANAG 4406 compliance
- **Options:** Any valid FFmpeg bitrate string (e.g., `"600k"`, `"1M"`, `"2M"`)
- **STANAG 4406:** `"600k"` provides good quality/bandwidth balance

#### preset
- **Type:** `str`
- **Default:** `"ultrafast"`
- **Description:** FFmpeg encoding preset for performance optimization
- **Options:** `"ultrafast"`, `"superfast"`, `"veryfast"`, `"faster"`, `"fast"`, `"medium"`, `"slow"`, `"slower"`, `"veryslow"`
- **STANAG 4406:** Use `"ultrafast"` for minimal encoding latency

## Configuration File Format

### YAML Configuration

Configuration can be specified in YAML files (e.g., `config/default.yaml`):

```yaml
mediamtx:
  # Core MediaMTX configuration
  host: "127.0.0.1"
  api_port: 9997
  rtsp_port: 8554
  webrtc_port: 8889
  hls_port: 8888
  
  # File system paths
  config_path: "/etc/mediamtx/mediamtx.yml"
  recordings_path: "/opt/camera-service/recordings"
  snapshots_path: "/opt/camera-service/snapshots"
  
  # STANAG 4406 H.264 codec configuration
  codec: "libx264"
  video_profile: "baseline"
  video_level: "3.0"
  pixel_format: "yuv420p"
  bitrate: "600k"
  preset: "ultrafast"
```

### Environment Variables

Configuration can also be set via environment variables:

```bash
# Core MediaMTX configuration
export MEDIAMTX_HOST="127.0.0.1"
export MEDIAMTX_API_PORT="9997"
export MEDIAMTX_RTSP_PORT="8554"
export MEDIAMTX_WEBRTC_PORT="8889"
export MEDIAMTX_HLS_PORT="8888"

# File system paths
export MEDIAMTX_CONFIG_PATH="/etc/mediamtx/mediamtx.yml"
export MEDIAMTX_RECORDINGS_PATH="/opt/camera-service/recordings"
export MEDIAMTX_SNAPSHOTS_PATH="/opt/camera-service/snapshots"

# STANAG 4406 H.264 codec configuration
export MEDIAMTX_CODEC="libx264"
export MEDIAMTX_VIDEO_PROFILE="baseline"
export MEDIAMTX_VIDEO_LEVEL="3.0"
export MEDIAMTX_PIXEL_FORMAT="yuv420p"
export MEDIAMTX_BITRATE="600k"
export MEDIAMTX_PRESET="ultrafast"
```

## STANAG 4406 Compliance

### FFmpeg Command Generation

The MediaMTXConfig parameters are used to generate STANAG 4406 compliant FFmpeg commands:

```bash
ffmpeg -f v4l2 -i {device_path} \
  -c:v {codec} \
  -profile:v {video_profile} \
  -level {video_level} \
  -pix_fmt {pixel_format} \
  -preset {preset} \
  -b:v {bitrate} \
  -f rtsp rtsp://{host}:{rtsp_port}/{path_name}
```

### Example STANAG 4406 Command

With default parameters, the generated command is:

```bash
ffmpeg -f v4l2 -i /dev/video0 \
  -c:v libx264 \
  -profile:v baseline \
  -level 3.0 \
  -pix_fmt yuv420p \
  -preset ultrafast \
  -b:v 600k \
  -f rtsp rtsp://127.0.0.1:8554/camera0
```

### Compliance Verification

To verify STANAG 4406 compliance:

1. **Profile Check:** Ensure `video_profile` is set to `"baseline"`
2. **Level Check:** Ensure `video_level` is set to `"3.0"` or lower
3. **Pixel Format Check:** Ensure `pixel_format` is set to `"yuv420p"`
4. **Bitrate Check:** Ensure `bitrate` is appropriate for your bandwidth requirements

## Usage Examples

### Basic Configuration

```python
from camera_service.config import MediaMTXConfig

# Use default STANAG 4406 configuration
config = MediaMTXConfig()
```

### Custom STANAG 4406 Configuration

```python
from camera_service.config import MediaMTXConfig

# Custom STANAG 4406 configuration
config = MediaMTXConfig(
    video_profile="baseline",
    video_level="3.0",
    pixel_format="yuv420p",
    bitrate="1M",  # Higher bitrate for better quality
    preset="fast"  # Better compression efficiency
)
```

### Configuration Loading

```python
from camera_service.config import ConfigManager

# Load configuration from file
config_manager = ConfigManager()
config = config_manager.load_config()

# Access MediaMTX configuration
mediamtx_config = config.mediamtx
print(f"Codec: {mediamtx_config.codec}")
print(f"Profile: {mediamtx_config.video_profile}")
print(f"Level: {mediamtx_config.video_level}")
```

## Migration Guide

### From Previous Configuration

If upgrading from a previous version without STANAG 4406 parameters:

1. **Add Missing Parameters:** The new parameters have sensible defaults
2. **Verify Compatibility:** Ensure your configuration works with STANAG 4406 requirements
3. **Test Streams:** Validate that generated FFmpeg commands work correctly

### Default Parameter Changes

- **video_profile:** Now defaults to `"baseline"` (was previously auto-detected)
- **video_level:** Now defaults to `"3.0"` (was previously auto-detected)
- **pixel_format:** Remains `"yuv420p"` (no change)
- **bitrate:** Remains `"600k"` (no change)
- **preset:** Remains `"ultrafast"` (no change)

## Troubleshooting

### Common Issues

1. **Configuration Loading Errors**
   - Ensure YAML syntax is correct
   - Check environment variable names
   - Verify file permissions

2. **STANAG 4406 Compliance Issues**
   - Verify `video_profile` is set to `"baseline"`
   - Check `video_level` is `"3.0"` or lower
   - Ensure `pixel_format` is `"yuv420p"`

3. **FFmpeg Command Generation**
   - Validate all parameters are strings
   - Check for invalid bitrate values
   - Verify preset names are correct

### Debug Configuration

```python
from camera_service.config import MediaMTXConfig

# Print configuration for debugging
config = MediaMTXConfig()
print(f"MediaMTXConfig: {config}")

# Validate specific parameters
assert config.video_profile == "baseline", "Profile must be baseline for STANAG 4406"
assert config.video_level == "3.0", "Level must be 3.0 for STANAG 4406"
assert config.pixel_format == "yuv420p", "Pixel format must be yuv420p for STANAG 4406"
```

---

**Configuration Version:** 1.0  
**STANAG 4406 Compliance:** ✅ **FULLY SUPPORTED**  
**Last Updated:** 2025-01-15
