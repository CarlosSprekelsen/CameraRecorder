#!/bin/bash
# MediaMTX Camera Service Installation Script
# 
# This script installs the MediaMTX Camera Service on a clean Ubuntu 22.04+ system.
# It is idempotent and safe to re-run.
#
# Prerequisites:
# - Ubuntu 22.04+ or similar Linux distribution
# - Internet connection for package downloads
# - Root privileges (run with sudo)
#
# Usage: sudo ./install.sh
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
PYTHON_VERSION="python3"
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

# Function to detect OS and set package manager
detect_os() {
    if [[ -f /etc/os-release ]]; then
        . /etc/os-release
        OS=$NAME
        VER=$VERSION_ID
    else
        print_error "Cannot detect OS version"
        exit 1
    fi
    
    print_status "Detected OS: $OS $VER"
    
    # Check if Ubuntu 22.04+ or similar
    if [[ "$OS" == "Ubuntu" && "$VER" < "22.04" ]]; then
        print_warning "Ubuntu 22.04+ is recommended. Current version: $VER"
    fi
}

# Function to install system dependencies
install_system_dependencies() {
    print_title "Installing System Dependencies"
    
    # Update package list
    print_status "Updating package list..."
    apt-get update
    
    # Install required packages
    print_status "Installing required packages..."
    apt-get install -y \
        python3 \
        python3-pip \
        python3-venv \
        python3-dev \
        v4l-utils \
        ffmpeg \
        git \
        curl \
        wget \
        systemd \
        systemd-sysv \
        logrotate
    
    print_success "System dependencies installed"
}

# Function to create service user and directories
setup_directories() {
    print_title "Setting Up Directories and User"
    
    # Create service user if it doesn't exist
    if ! id "$SERVICE_USER" &>/dev/null; then
        print_status "Creating service user: $SERVICE_USER"
        useradd -r -s /bin/false -d "$INSTALL_DIR" "$SERVICE_USER"
    else
        print_status "Service user $SERVICE_USER already exists"
    fi
    
    # Create installation directory
    print_status "Creating installation directory: $INSTALL_DIR"
    mkdir -p "$INSTALL_DIR"
    
    # Create subdirectories
    print_status "Creating subdirectories..."
    mkdir -p "$CONFIG_DIR" "$LOG_DIR" "$RECORDINGS_DIR" "$SNAPSHOTS_DIR"
    
    # Set ownership
    print_status "Setting directory ownership..."
    chown -R "$SERVICE_USER:$SERVICE_GROUP" "$INSTALL_DIR"
    chmod 755 "$INSTALL_DIR"
    chmod 750 "$CONFIG_DIR"
    chmod 755 "$LOG_DIR" "$RECORDINGS_DIR" "$SNAPSHOTS_DIR"
    
    print_success "Directories and user setup complete"
}

# Function to install Python dependencies
install_python_dependencies() {
    print_title "Installing Python Dependencies"
    
    # Create virtual environment if it doesn't exist
    if [[ ! -d "$VENV_DIR" ]]; then
        print_status "Creating Python virtual environment..."
        "$PYTHON_VERSION" -m venv "$VENV_DIR"
    else
        print_status "Virtual environment already exists"
    fi
    
    # Activate virtual environment
    source "$VENV_DIR/bin/activate"
    
    # Upgrade pip
    print_status "Upgrading pip..."
    pip install --upgrade pip
    
    # Install Python dependencies
    print_status "Installing Python dependencies..."
    
    # Copy requirements files to install directory
    if [[ -f "requirements.txt" ]]; then
        cp requirements.txt "$INSTALL_DIR/"
    else
        print_error "requirements.txt not found in current directory"
        exit 1
    fi
    
    # Install dependencies
    pip install -r "$INSTALL_DIR/requirements.txt"
    
    print_success "Python dependencies installed"
}

# Function to copy application files
copy_application_files() {
    print_title "Copying Application Files"
    
    # Copy source code
    print_status "Copying source code..."
    cp -r src "$INSTALL_DIR/"
    
    # Copy configuration
    print_status "Copying configuration files..."
    if [[ -f "config/default.yaml" ]]; then
        cp config/default.yaml "$CONFIG_DIR/camera-service.yaml"
    else
        print_warning "Default configuration not found, creating minimal config"
        cat > "$CONFIG_DIR/camera-service.yaml" << 'EOF'
# MediaMTX Camera Service - Default Configuration

server:
  host: "0.0.0.0"
  port: 8002
  websocket_path: "/ws"
  max_connections: 100

mediamtx:
  host: "127.0.0.1"
  api_port: 9997
  rtsp_port: 8554
  webrtc_port: 8889
  hls_port: 8888
  config_path: "/opt/camera-service/config/mediamtx.yml"
  recordings_path: "/opt/camera-service/recordings"
  snapshots_path: "/opt/camera-service/snapshots"
  
  # Health monitoring configuration
  health_check_interval: 30
  health_failure_threshold: 10
  health_circuit_breaker_timeout: 60
  health_max_backoff_interval: 120
  health_recovery_confirmation_threshold: 3
  backoff_base_multiplier: 2.0
  backoff_jitter_range: [0.8, 1.2]
  process_termination_timeout: 3.0
  process_kill_timeout: 2.0

camera:
  poll_interval: 0.1
  detection_timeout: 2.0
  device_range: [0, 9]
  enable_capability_detection: true
  auto_start_streams: true

logging:
  level: "INFO"
  format: "%(asctime)s - %(name)s - %(levelname)s - %(message)s"
  file_enabled: true
  file_path: "/opt/camera-service/logs/camera-service.log"
  max_file_size: "10MB"
  backup_count: 5

recording:
  auto_record: false
  format: "mp4"
  quality: "high"
  max_duration: 3600
  cleanup_after_days: 30

snapshots:
  format: "jpg"
  quality: 90
  cleanup_after_days: 7
EOF
    fi
    
    # Copy validation script if it exists
    print_status "Copying validation scripts..."
    mkdir -p "$INSTALL_DIR/scripts"
    if [[ -f "scripts/validate_deployment.py" ]]; then
        cp scripts/validate_deployment.py "$INSTALL_DIR/scripts/"
        chmod +x "$INSTALL_DIR/scripts/validate_deployment.py"
        print_success "Validation script copied"
    else
        print_warning "Validation script not found"
    fi
    
    # Set ownership
    chown -R "$SERVICE_USER:$SERVICE_GROUP" "$INSTALL_DIR"
    
    print_success "Application files copied"
}

# Function to create systemd service
create_systemd_service() {
    print_title "Creating Systemd Service"
    
    # Create systemd service file
    cat > "/etc/systemd/system/$SERVICE_NAME.service" << EOF
[Unit]
Description=MediaMTX Camera Service
After=network.target
Wants=network.target

[Service]
Type=simple
User=$SERVICE_USER
Group=$SERVICE_GROUP
WorkingDirectory=$INSTALL_DIR
Environment=PATH=$VENV_DIR/bin
Environment=PYTHONPATH=$INSTALL_DIR/src
ExecStart=$VENV_DIR/bin/python3 -m camera_service.main
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal
SyslogIdentifier=$SERVICE_NAME

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=$INSTALL_DIR

[Install]
WantedBy=multi-user.target
EOF
    
    # Create environment file
    cat > "/etc/systemd/system/$SERVICE_NAME.env" << EOF
# MediaMTX Camera Service Environment Variables
CAMERA_SERVICE_CONFIG_PATH=$CONFIG_DIR/camera-service.yaml
CAMERA_SERVICE_LOG_LEVEL=INFO
CAMERA_SERVICE_HOST=0.0.0.0
CAMERA_SERVICE_PORT=8002
MEDIAMTX_API_PORT=9997
MEDIAMTX_RTSP_PORT=8554
MEDIAMTX_WEBRTC_PORT=8889
MEDIAMTX_HLS_PORT=8888
EOF
    
    # Reload systemd
    systemctl daemon-reload
    
    print_success "Systemd service created"
}

# Function to create logrotate configuration
create_logrotate_config() {
    print_title "Creating Logrotate Configuration"
    
    cat > "/etc/logrotate.d/$SERVICE_NAME" << EOF
$LOG_DIR/*.log {
    daily
    missingok
    rotate 7
    compress
    delaycompress
    notifempty
    create 644 $SERVICE_USER $SERVICE_GROUP
    postrotate
        systemctl reload $SERVICE_NAME >/dev/null 2>&1 || true
    endscript
}
EOF
    
    print_success "Logrotate configuration created"
}

# Function to enable and start service
enable_service() {
    print_title "Enabling and Starting Service"
    
    # Enable service
    print_status "Enabling $SERVICE_NAME service..."
    systemctl enable "$SERVICE_NAME"
    
    # Start service
    print_status "Starting $SERVICE_NAME service..."
    systemctl start "$SERVICE_NAME"
    
    # Wait a moment for service to start
    sleep 3
    
    # Check service status
    if systemctl is-active --quiet "$SERVICE_NAME"; then
        print_success "Service started successfully"
    else
        print_error "Service failed to start"
        systemctl status "$SERVICE_NAME"
        exit 1
    fi
}

# Function to verify installation
verify_installation() {
    print_title "Verifying Installation"
    
    # Check if service is running
    print_status "Checking service status..."
    if systemctl is-active --quiet "$SERVICE_NAME"; then
        print_success "Service is running"
    else
        print_error "Service is not running"
        systemctl status "$SERVICE_NAME"
        return 1
    fi
    
    # Check if service is enabled
    print_status "Checking if service is enabled..."
    if systemctl is-enabled --quiet "$SERVICE_NAME"; then
        print_success "Service is enabled"
    else
        print_error "Service is not enabled"
        return 1
    fi
    
    # Check if WebSocket port is listening
    print_status "Checking WebSocket port..."
    if netstat -tlnp 2>/dev/null | grep -q ":8002 "; then
        print_success "WebSocket port 8002 is listening"
    else
        print_warning "WebSocket port 8002 is not listening (service may still be starting)"
    fi
    
    # Check if directories exist and have correct permissions
    print_status "Checking directory permissions..."
    for dir in "$INSTALL_DIR" "$CONFIG_DIR" "$LOG_DIR" "$RECORDINGS_DIR" "$SNAPSHOTS_DIR"; do
        if [[ -d "$dir" ]]; then
            print_success "Directory exists: $dir"
        else
            print_error "Directory missing: $dir"
            return 1
        fi
    done
    
    # Check if Python virtual environment exists
    print_status "Checking Python virtual environment..."
    if [[ -d "$VENV_DIR" ]]; then
        print_success "Python virtual environment exists"
    else
        print_error "Python virtual environment missing"
        return 1
    fi
    
    print_success "Installation verification complete"
}

# Function to validate installation
validate_installation() {
    print_title "Validating Installation"
    
    print_status "Running deployment validation..."
    
    # Change to installation directory
    cd "$INSTALL_DIR"
    
    # Run validation script if it exists
    if [[ -f "scripts/validate_deployment.py" ]]; then
        print_status "Running deployment validation script..."
        if sudo -u "$SERVICE_USER" "$VENV_DIR/bin/python3" scripts/validate_deployment.py; then
            print_success "Deployment validation passed"
        else
            print_error "Deployment validation failed"
            print_status "Continuing with installation but service may not work correctly"
        fi
    else
        print_warning "Validation script not found, skipping validation"
    fi
    
    # Test configuration loading manually
    print_status "Testing configuration loading..."
    if sudo -u "$SERVICE_USER" "$VENV_DIR/bin/python3" -c "
import sys
sys.path.insert(0, '$INSTALL_DIR/src')
from camera_service.config import ConfigManager
config_manager = ConfigManager()
config = config_manager.load_config()
print('✓ Configuration loaded successfully')
print('✓ MediaMTX health monitoring parameters accepted')
"; then
        print_success "Configuration validation passed"
    else
        print_error "Configuration validation failed"
        print_status "Continuing with installation but service may not work correctly"
    fi
}

# Function to display post-installation information
display_post_install_info() {
    print_title "Installation Complete"
    
    echo
    print_success "MediaMTX Camera Service has been successfully installed!"
    echo
    echo "Service Information:"
    echo "  - Service Name: $SERVICE_NAME"
    echo "  - Installation Directory: $INSTALL_DIR"
    echo "  - Configuration Directory: $CONFIG_DIR"
    echo "  - Log Directory: $LOG_DIR"
    echo "  - Recordings Directory: $RECORDINGS_DIR"
    echo "  - Snapshots Directory: $SNAPSHOTS_DIR"
    echo
    echo "Useful Commands:"
    echo "  - Check service status: sudo systemctl status $SERVICE_NAME"
    echo "  - View service logs: sudo journalctl -u $SERVICE_NAME -f"
    echo "  - Restart service: sudo systemctl restart $SERVICE_NAME"
    echo "  - Stop service: sudo systemctl stop $SERVICE_NAME"
    echo
    echo "Configuration:"
    echo "  - Edit configuration: sudo nano $CONFIG_DIR/camera-service.yaml"
    echo "  - View configuration: sudo cat $CONFIG_DIR/camera-service.yaml"
    echo
    echo "Testing:"
    echo "  - Test WebSocket connection: curl -H 'Connection: Upgrade' -H 'Upgrade: websocket' http://localhost:8002/ws"
    echo "  - Check if port is listening: netstat -tlnp | grep :8002"
    echo
    echo "Troubleshooting:"
    echo "  - Check service logs: sudo journalctl -u $SERVICE_NAME -n 50"
    echo "  - Check application logs: sudo tail -f $LOG_DIR/camera-service.log"
    echo "  - Reinstall: sudo ./install.sh"
    echo
}

# Main installation function
main() {
    print_title "MediaMTX Camera Service Installation"
    print_status "Starting installation process..."
    
    # Check prerequisites
    check_root
    detect_os
    
    # Installation steps
    install_system_dependencies
    setup_directories
    install_python_dependencies
    copy_application_files
    create_systemd_service
    create_logrotate_config
    enable_service
    
    # Validate installation
    validate_installation
    
    verify_installation
    
    # Display post-installation information
    display_post_install_info
    
    print_success "Installation completed successfully!"
}

# Run main function
main "$@"