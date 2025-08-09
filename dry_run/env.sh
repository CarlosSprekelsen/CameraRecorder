#!/usr/bin/env bash
# Source this file to set the dry-run environment
export CDR_BASELINE_TAG=vX.Y.Z-cdr\ \(MISSING\)
export CDR_SHA=4f3ba0e07b3e
export OS_NAME=Ubuntu
export OS_VERSION=22.04
export KERNEL_VERSION=5.15.0-151-generic
export PYTHON_VERSION=Python\ 3.10.12
export MEDIAMTX_VERSION=v1.13.1
export CAMERA_IDS=0\,1
export CAMERA_COUNT=2

# Application-required settings (minimal)
export JWT_SECRET_KEY=dryrun-secret
export API_KEYS_FILE=/home/dts/CameraRecorder/dry_run/api-keys.json
export MEDIAMTX_HOST=localhost
export MEDIAMTX_API_PORT=9997
# Optional overrides:
# export SERVER_HOST="0.0.0.0"
# export SERVER_PORT="8002"
# export SSL_ENABLED="false"
