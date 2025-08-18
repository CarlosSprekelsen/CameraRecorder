#!/bin/bash
# MediaMTX Camera Service Installation Verification Script
# 
# This script verifies that the MediaMTX Camera Service installation is complete and working.
# It checks services, ports, files, and basic functionality.
#
# Usage: sudo ./verify_installation.sh
#
# Author: MediaMTX Camera Service Team
# Date: 2025-01-27

set -euo pipefail

# Configuration
SERVICE_USER="camera-service"
SERVICE_GROUP="camera-service"
INSTALL_DIR="/opt/camera-service"
CONFIG_DIR="/opt/camera-service/config"
LOG_DIR="/opt/camera-service/logs"
RECORDINGS_DIR="/opt/camera-service/recordings"
SNAPSHOTS_DIR="/opt/camera-service/snapshots"
VENV_DIR="/opt/camera-service/venv"
SERVICE_NAME="camera-service"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
print_status() { echo -e "${BLUE}[INFO]${NC} $1"; }
print_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
print_error() { echo -e "${RED}[ERROR]${NC} $1"; }
print_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
print_title() { echo -e "\n${BLUE}=== $1 ===${NC}"; }

# Function to check if running as root
check_root() {
    if [[ $EUID -ne 0 ]]; then
        print_error "This script must be run as root (use sudo)"
        exit 1
    fi
}

# Function to verify MediaMTX installation
verify_mediamtx() {
    print_title "Verifying MediaMTX Installation"
    
    # Check if MediaMTX service is running
    if systemctl is-active --quiet mediamtx; then
        print_success "MediaMTX service is running"
    else
        print_error "MediaMTX service is not running"
        return 1
    fi
    
    # Check if MediaMTX service is enabled
    if systemctl is-enabled --quiet mediamtx; then
        print_success "MediaMTX service is enabled"
    else
        print_error "MediaMTX service is not enabled"
        return 1
    fi
    
    # Check MediaMTX ports
    print_status "Checking MediaMTX ports..."
    for port in 8554 8888 8889 9997; do
        if netstat -tlnp 2>/dev/null | grep -q ":$port "; then
            print_success "Port $port is listening"
        else
            print_warning "Port $port is not listening"
        fi
    done
    
    # Test MediaMTX API
    print_status "Testing MediaMTX API..."
    if curl -s http://localhost:9997/v3/config/global/get >/dev/null 2>&1; then
        print_success "MediaMTX API is responding"
    else
        print_error "MediaMTX API is not responding"
        return 1
    fi
    
    # Check MediaMTX files
    print_status "Checking MediaMTX files..."
    if [[ -f "/opt/mediamtx/mediamtx" ]]; then
        print_success "MediaMTX binary exists"
    else
        print_error "MediaMTX binary missing"
        return 1
    fi
    
    if [[ -f "/opt/mediamtx/config/mediamtx.yml" ]]; then
        print_success "MediaMTX configuration exists"
    else
        print_error "MediaMTX configuration missing"
        return 1
    fi
}

# Function to verify camera service installation
verify_camera_service() {
    print_title "Verifying Camera Service Installation"
    
    # Check if camera service is running
    if systemctl is-active --quiet "$SERVICE_NAME"; then
        print_success "Camera service is running"
    else
        print_error "Camera service is not running"
        return 1
    fi
    
    # Check if camera service is enabled
    if systemctl is-enabled --quiet "$SERVICE_NAME"; then
        print_success "Camera service is enabled"
    else
        print_error "Camera service is not enabled"
        return 1
    fi
    
    # Check camera service port
    print_status "Checking camera service port..."
    if netstat -tlnp 2>/dev/null | grep -q ":8002 "; then
        print_success "Port 8002 (WebSocket) is listening"
    else
        print_warning "Port 8002 (WebSocket) is not listening"
    fi
    
    # Check camera service files
    print_status "Checking camera service files..."
    for dir in "$INSTALL_DIR" "$CONFIG_DIR" "$LOG_DIR" "$RECORDINGS_DIR" "$SNAPSHOTS_DIR"; do
        if [[ -d "$dir" ]]; then
            print_success "Directory exists: $dir"
        else
            print_error "Directory missing: $dir"
            return 1
        fi
    done
    
    # Check Python virtual environment
    if [[ -d "$VENV_DIR" ]]; then
        print_success "Python virtual environment exists"
    else
        print_error "Python virtual environment missing"
        return 1
    fi
    
    # Check source files
    if [[ -f "$INSTALL_DIR/src/camera_service/main.py" ]]; then
        print_success "Camera service source files exist"
    else
        print_error "Camera service source files missing"
        return 1
    fi
    
    # Check configuration file
    if [[ -f "$CONFIG_DIR/camera-service.yaml" ]]; then
        print_success "Camera service configuration exists"
    else
        print_error "Camera service configuration missing"
        return 1
    fi
}

# Function to verify system dependencies
verify_dependencies() {
    print_title "Verifying System Dependencies"
    
    # Check Python version
    PYTHON_VERSION=$(python3 --version 2>&1 | cut -d' ' -f2 | cut -d'.' -f1,2)
    if [[ "$PYTHON_VERSION" == "3.10" ]] || [[ "$PYTHON_VERSION" > "3.10" ]]; then
        print_success "Python version: $PYTHON_VERSION"
    else
        print_warning "Python version: $PYTHON_VERSION (3.10+ recommended)"
    fi
    
    # Check required commands
    for cmd in v4l2-ctl ffmpeg curl wget; do
        if command -v "$cmd" >/dev/null 2>&1; then
            print_success "$cmd is available"
        else
            print_error "$cmd is not available"
        fi
    done
    
    # Check camera devices
    print_status "Checking camera devices..."
    CAMERA_COUNT=$(ls /dev/video* 2>/dev/null | wc -l)
    if [[ "$CAMERA_COUNT" -gt 0 ]]; then
        print_success "Found $CAMERA_COUNT camera device(s)"
    else
        print_warning "No camera devices found"
    fi
}

# Function to verify network connectivity
verify_network() {
    print_title "Verifying Network Connectivity"
    
    # Check if ports are accessible
    print_status "Testing port accessibility..."
    
    # Test WebSocket port
    if curl -s -H "Connection: Upgrade" -H "Upgrade: websocket" http://localhost:8002/ws >/dev/null 2>&1; then
        print_success "WebSocket endpoint is accessible"
    else
        print_warning "WebSocket endpoint is not accessible"
    fi
    
    # Test MediaMTX API
    if curl -s http://localhost:9997/v3/config/global/get >/dev/null 2>&1; then
        print_success "MediaMTX API is accessible"
    else
        print_warning "MediaMTX API is not accessible"
    fi
}

# Function to verify file management API (Epic E6)
verify_file_management_api() {
    print_title "Verifying File Management API (Epic E6)"
    
    # Check if health server is running on port 8003
    if netstat -tlnp 2>/dev/null | grep -q ":8003 "; then
        print_success "Health server is listening on port 8003"
    else
        print_error "Health server is not listening on port 8003"
        return 1
    fi
    
    # Test health endpoint
    print_status "Testing health endpoint..."
    if curl -s http://localhost:8003/health/system >/dev/null 2>&1; then
        print_success "Health endpoint is responding"
    else
        print_error "Health endpoint is not responding"
        return 1
    fi
    
    # Test file download endpoints
    print_status "Testing file download endpoints..."
    
    # Test recordings endpoint
    if curl -s -I http://localhost:8003/files/recordings/ >/dev/null 2>&1; then
        print_success "Recordings download endpoint is accessible"
    else
        print_warning "Recordings download endpoint is not accessible"
    fi
    
    # Test snapshots endpoint
    if curl -s -I http://localhost:8003/files/snapshots/ >/dev/null 2>&1; then
        print_success "Snapshots download endpoint is accessible"
    else
        print_warning "Snapshots download endpoint is not accessible"
    fi
    
    # Check if directories exist and are accessible
    print_status "Checking file directories..."
    if [[ -d "$RECORDINGS_DIR" ]]; then
        print_success "Recordings directory exists: $RECORDINGS_DIR"
    else
        print_warning "Recordings directory missing: $RECORDINGS_DIR"
    fi
    
    if [[ -d "$SNAPSHOTS_DIR" ]]; then
        print_success "Snapshots directory exists: $SNAPSHOTS_DIR"
    else
        print_warning "Snapshots directory missing: $SNAPSHOTS_DIR"
    fi
    
    # Test nginx configuration
    print_status "Testing nginx configuration..."
    if nginx -t >/dev/null 2>&1; then
        print_success "Nginx configuration is valid"
    else
        print_error "Nginx configuration is invalid"
        return 1
    fi
    
    # Test SSL file endpoints through nginx
    print_status "Testing SSL file endpoints..."
    if curl -s -k -I https://localhost/files/recordings/ >/dev/null 2>&1; then
        print_success "SSL recordings endpoint is accessible"
    else
        print_warning "SSL recordings endpoint is not accessible"
    fi
    
    if curl -s -k -I https://localhost/files/snapshots/ >/dev/null 2>&1; then
        print_success "SSL snapshots endpoint is accessible"
    else
        print_warning "SSL snapshots endpoint is not accessible"
    fi
}

# Function to verify logs
verify_logs() {
    print_title "Verifying Log Files"
    
    # Check if log directory exists and has files
    if [[ -d "$LOG_DIR" ]]; then
        LOG_FILES=$(ls "$LOG_DIR"/*.log 2>/dev/null | wc -l)
        if [[ "$LOG_FILES" -gt 0 ]]; then
            print_success "Log files exist: $LOG_FILES file(s)"
        else
            print_warning "No log files found"
        fi
    else
        print_error "Log directory missing"
    fi
    
    # Check recent service logs
    print_status "Checking recent service logs..."
    if journalctl -u "$SERVICE_NAME" -n 5 2>/dev/null | grep -q .; then
        print_success "Camera service logs are being generated"
    else
        print_warning "No recent camera service logs found"
    fi
    
    if journalctl -u mediamtx -n 5 2>/dev/null | grep -q .; then
        print_success "MediaMTX logs are being generated"
    else
        print_warning "No recent MediaMTX logs found"
    fi
}

# Function to display summary
display_summary() {
    print_title "Verification Summary"
    
    echo
    echo "Installation Status:"
    echo "  - MediaMTX Server: $(systemctl is-active mediamtx 2>/dev/null || echo 'FAILED')"
    echo "  - Camera Service: $(systemctl is-active $SERVICE_NAME 2>/dev/null || echo 'FAILED')"
    echo
    echo "Port Status:"
    echo "  - RTSP (8554): $(netstat -tlnp 2>/dev/null | grep -q ':8554 ' && echo 'LISTENING' || echo 'NOT LISTENING')"
    echo "  - WebSocket (8002): $(netstat -tlnp 2>/dev/null | grep -q ':8002 ' && echo 'LISTENING' || echo 'NOT LISTENING')"
    echo "  - Health Server (8003): $(netstat -tlnp 2>/dev/null | grep -q ':8003 ' && echo 'LISTENING' || echo 'NOT LISTENING')"
    echo "  - MediaMTX API (9997): $(netstat -tlnp 2>/dev/null | grep -q ':9997 ' && echo 'LISTENING' || echo 'NOT LISTENING')"
    echo
    echo "Camera Devices: $(ls /dev/video* 2>/dev/null | wc -l)"
    echo
    echo "Useful Commands:"
    echo "  - View camera service logs: sudo journalctl -u $SERVICE_NAME -f"
    echo "  - View MediaMTX logs: sudo journalctl -u mediamtx -f"
    echo "  - Check service status: sudo systemctl status $SERVICE_NAME mediamtx"
    echo "  - Test WebSocket: curl -H 'Connection: Upgrade' -H 'Upgrade: websocket' http://localhost:8002/ws"
    echo "  - Test MediaMTX API: curl http://localhost:9997/v3/config/global/get"
    echo "  - Test file download: curl -I https://localhost/files/recordings/"
    echo "  - Test file download: curl -I https://localhost/files/snapshots/"
    echo
}

# Main verification function
main() {
    print_title "MediaMTX Camera Service Installation Verification"
    print_status "Starting verification process..."
    
    # Check prerequisites
    check_root
    
    # Run verification checks
    verify_mediamtx
    verify_camera_service
    verify_dependencies
    verify_network
    verify_file_management_api
    verify_logs
    
    # Display summary
    display_summary
    
    print_success "Verification completed!"
}

# Run main function
main "$@" 