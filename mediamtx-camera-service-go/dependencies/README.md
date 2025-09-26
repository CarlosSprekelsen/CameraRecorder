# Dependencies

This directory contains local dependencies for the MediaMTX Camera Service to ensure offline installation.

## MediaMTX v1.15.1

- **File**: `mediamtx_v1.15.1_linux_amd64.tar.gz`
- **Source**: https://github.com/bluenviron/mediamtx/releases/download/v1.15.1/mediamtx_v1.15.1_linux_amd64.tar.gz
- **Size**: ~21MB
- **Contents**: 
  - `mediamtx` - MediaMTX binary
  - `mediamtx.yml` - Default configuration
  - `LICENSE` - License file

## Usage

The install script automatically uses these local dependencies instead of downloading from the internet, ensuring:
- ✅ **Offline installation** - No internet dependency
- ✅ **Version stability** - Fixed version that works
- ✅ **Faster installation** - No download time
- ✅ **Reliability** - No network failures

## Updating Dependencies

To update MediaMTX to a newer version:
1. Download the new tar.gz file from GitHub releases
2. Replace the existing file in this directory
3. Update the install script to reference the new filename
4. Test the installation

## Verification

To verify the dependency is valid:
```bash
tar -tzf mediamtx_v1.15.1_linux_amd64.tar.gz
```

Should show:
- mediamtx
- mediamtx.yml  
- LICENSE
