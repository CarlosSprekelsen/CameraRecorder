#!/bin/bash

# Debug script to test install script step by step

set -e

echo "=== Debug Install Script ==="
echo "Current directory: $(pwd)"
echo "Script directory: $(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
echo ""

# Test 1: Check if we can find the source files
echo "1. Checking source files..."
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")"

echo "Script directory: $SCRIPT_DIR"
echo "Project root: $PROJECT_ROOT"

if [ -d "$PROJECT_ROOT/src" ]; then
    echo "✅ Source directory found: $PROJECT_ROOT/src"
    ls -la "$PROJECT_ROOT/src"
else
    echo "❌ Source directory not found: $PROJECT_ROOT/src"
fi

if [ -f "$PROJECT_ROOT/requirements.txt" ]; then
    echo "✅ Requirements file found: $PROJECT_ROOT/requirements.txt"
else
    echo "❌ Requirements file not found: $PROJECT_ROOT/requirements.txt"
fi

echo ""

# Test 2: Check MediaMTX installation
echo "2. Testing MediaMTX installation..."
MEDIAMTX_DIR="/opt/mediamtx"

if [ -d "$MEDIAMTX_DIR" ]; then
    echo "✅ MediaMTX directory exists: $MEDIAMTX_DIR"
    ls -la "$MEDIAMTX_DIR"
else
    echo "❌ MediaMTX directory not found: $MEDIAMTX_DIR"
fi

if [ -d "$MEDIAMTX_DIR/config" ]; then
    echo "✅ MediaMTX config directory exists"
    ls -la "$MEDIAMTX_DIR/config"
else
    echo "❌ MediaMTX config directory not found"
fi

echo ""

# Test 3: Check service status
echo "3. Checking service status..."
if systemctl list-unit-files | grep -q camera-service; then
    echo "✅ Camera service found in systemd"
    systemctl status camera-service || echo "Service not running"
else
    echo "❌ Camera service not found in systemd"
fi

if systemctl list-unit-files | grep -q mediamtx; then
    echo "✅ MediaMTX service found in systemd"
    systemctl status mediamtx || echo "Service not running"
else
    echo "❌ MediaMTX service not found in systemd"
fi

echo ""

# Test 4: Check ports
echo "4. Checking ports..."
if netstat -tlnp 2>/dev/null | grep -q ":8002"; then
    echo "✅ Port 8002 is in use"
    netstat -tlnp | grep ":8002"
else
    echo "❌ Port 8002 is not in use"
fi

if netstat -tlnp 2>/dev/null | grep -q ":9997"; then
    echo "✅ Port 9997 is in use"
    netstat -tlnp | grep ":9997"
else
    echo "❌ Port 9997 is not in use"
fi

echo ""
echo "=== Debug complete ===" 