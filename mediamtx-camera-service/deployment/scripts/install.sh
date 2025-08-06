#!/bin/bash

# MediaMTX Camera Service Installation Script
# Installs MediaMTX server and Camera Service with security configuration
# as specified in Sprint 2 Day 2 Task S7.3

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SERVICE_NAME="camera-service"
SERVICE_USER="camera-service"
SERVICE_GROUP="camera-service"
INSTALL_DIR="/opt/camera-service"
MEDIAMTX_DIR="/opt/mediamtx"

# Function to log messages
log_message() {
    echo -e "${BLUE}[$(date '+%Y-%m-%d %H:%M:%S')]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[$(date '+%Y-%m-%d %H:%M:%S')] SUCCESS:${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[$(date '+%Y-%m-%d %H:%M:%S')] WARNING:${NC} $1"
}

log_error() {
    echo -e "${RED}[$(date '+%Y-%m-%d %H:%M:%S')] ERROR:${NC} $1"
}

# Check if running as root
if [[ $EUID -ne 0 ]]; then
    log_error "This script must be run as root (use sudo)"
    exit 1
fi

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to install system dependencies
install_system_dependencies() {
    log_message "Installing system dependencies..."
    
    # Update package list
    apt-get update
    
    # Install required packages
    apt-get install -y \
        python3 \
        python3-pip \
        python3-venv \
        git \
        wget \
        curl \
        ffmpeg \
        v4l-utils \
        systemd \
        systemd-sysv
    
    log_success "System dependencies installed"
}

# Function to create service user
create_service_user() {
    log_message "Creating service user..."
    
    # Create user if it doesn't exist
    if ! id "$SERVICE_USER" &>/dev/null; then
        useradd -r -s /bin/false -d "$INSTALL_DIR" "$SERVICE_USER"
        log_success "Service user created: $SERVICE_USER"
    else
        log_message "Service user already exists: $SERVICE_USER"
    fi
}

# Function to install MediaMTX
install_mediamtx() {
    log_message "Installing MediaMTX server..."
    
    # Save original directory
    ORIGINAL_DIR="$(pwd)"
    
    # Create MediaMTX directory
    mkdir -p "$MEDIAMTX_DIR"
    
    # Copy bundled MediaMTX v1.13.1
    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    PROJECT_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")"
    MEDIAMTX_SOURCE="$PROJECT_ROOT/dependencies/mediamtx"
    
    if [ ! -f "$MEDIAMTX_SOURCE/mediamtx" ]; then
        log_error "Bundled MediaMTX not found at $MEDIAMTX_SOURCE/mediamtx"
        exit 1
    fi
    
    cp "$MEDIAMTX_SOURCE/mediamtx" "$MEDIAMTX_DIR/"
    chmod +x "$MEDIAMTX_DIR/mediamtx"
    
    # Create MediaMTX user
    if ! id "mediamtx" &>/dev/null; then
        useradd -r -s /bin/false -d "$MEDIAMTX_DIR" mediamtx
    fi
    
    # Set ownership
    chown -R mediamtx:mediamtx "$MEDIAMTX_DIR"
    
    # Create MediaMTX config directory
    mkdir -p "$MEDIAMTX_DIR/config"
    
    # Create MediaMTX configuration
    cat > "$MEDIAMTX_DIR/config/mediamtx.yml" << EOF
# MediaMTX v1.13.1 Configuration for Camera Service
logLevel: info
logDestinations: [stdout]

# Enable Control API
api: yes
apiAddress: :9997

# Enable RTSP server
rtsp: yes
rtspTransports: [udp, tcp]
rtspAddress: :8554

# Enable WebRTC server
webrtc: yes
webrtcAddress: :8889

# Enable HLS server
hls: yes
hlsAddress: :8888
hlsVariant: lowLatency

# Path defaults
pathDefaults:
  recordFormat: fmp4
  recordSegmentDuration: 3600s

# Paths configuration
paths:
  all:
EOF
    
    # Create MediaMTX systemd service
    cat > /etc/systemd/system/mediamtx.service << EOF
[Unit]
Description=MediaMTX Media Server
Documentation=https://github.com/bluenviron/mediamtx
After=network.target
Wants=network.target

[Service]
Type=simple
User=mediamtx
Group=mediamtx
WorkingDirectory=$MEDIAMTX_DIR
ExecStart=$MEDIAMTX_DIR/mediamtx $MEDIAMTX_DIR/config/mediamtx.yml
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal
SyslogIdentifier=mediamtx

[Install]
WantedBy=multi-user.target
EOF
    
    # Enable and start MediaMTX service
    systemctl daemon-reload
    systemctl enable mediamtx
    systemctl start mediamtx
    
    # Return to original directory
    cd "$ORIGINAL_DIR"
    
    log_success "MediaMTX server installed and started"
}

# Function to install Camera Service
install_camera_service() {
    log_message "Installing Camera Service..."
    
    # Get the script directory to find source files
    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    PROJECT_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")"
    
    # Create installation directory
    mkdir -p "$INSTALL_DIR"
    
    # Copy service files using absolute paths
    if [ -d "$PROJECT_ROOT/src" ]; then
        cp -r "$PROJECT_ROOT/src" "$INSTALL_DIR/"
        log_success "Source files copied"
    else
        log_error "Source directory not found at $PROJECT_ROOT/src"
        exit 1
    fi
    
    if [ -d "$PROJECT_ROOT/tests" ]; then
        cp -r "$PROJECT_ROOT/tests" "$INSTALL_DIR/"
        log_success "Test files copied"
    else
        log_warning "Test directory not found at $PROJECT_ROOT/tests"
    fi
    
    if [ -f "$PROJECT_ROOT/requirements.txt" ]; then
        cp "$PROJECT_ROOT/requirements.txt" "$INSTALL_DIR/"
        log_success "Requirements file copied"
    else
        log_warning "Requirements file not found at $PROJECT_ROOT/requirements.txt"
    fi
    
    if [ -f "$PROJECT_ROOT/run_all_tests.py" ]; then
        cp "$PROJECT_ROOT/run_all_tests.py" "$INSTALL_DIR/"
        log_success "Test runner copied"
    else
        log_warning "Test runner not found at $PROJECT_ROOT/run_all_tests.py"
    fi
    
    # Create configuration directory
    mkdir -p "$INSTALL_DIR/config"
    
    # Create required directories with proper permissions
    log_message "Creating required directories..."
    mkdir -p /var/recordings /var/snapshots
    chown "$SERVICE_USER:$SERVICE_GROUP" /var/recordings /var/snapshots
    chmod 755 /var/recordings /var/snapshots
    log_success "Required directories created with proper permissions"
    
    # Create camera service configuration
    cat > "$INSTALL_DIR/config/camera-service.yaml" << EOF
# Camera Service Configuration
server:
  host: "0.0.0.0"
  port: 8002
  websocket_path: "/ws"
  max_connections: 100

security:
  jwt:
    secret_key: "\${JWT_SECRET_KEY}"
    expiry_hours: 24
    algorithm: "HS256"
  
  api_keys:
    storage_file: "\${API_KEYS_FILE}"
  
  ssl:
    enabled: false
    cert_file: "\${SSL_CERT_FILE}"
    key_file: "\${SSL_KEY_FILE}"
  
  rate_limiting:
    max_connections: 100
    requests_per_minute: 60
  
  health:
    port: 8003
    bind_address: "0.0.0.0"

mediamtx:
  host: "localhost"
  api_port: 9997
  rtsp_port: 8554
  webrtc_port: 8889
  hls_port: 8888
  config_path: "/etc/mediamtx/mediamtx.yml"
  recordings_path: "/var/recordings"
  snapshots_path: "/var/snapshots"

cameras:
  discovery_enabled: true
  polling_interval: 30

logging:
  level: "INFO"
  format: "%(asctime)s - %(name)s - %(levelname)s - %(message)s"
  file_enabled: true
  file_path: "/var/log/camera-service/camera-service.log"
  max_file_size: "10MB"
  backup_count: 5

recording:
  auto_record: false
  format: "mp4"
  quality: "medium"
  max_duration: 3600
  cleanup_after_days: 30

snapshots:
  format: "jpg"
  quality: 85
  cleanup_after_days: 7
EOF
    
    # Create security directories
    mkdir -p "$INSTALL_DIR/security/api-keys"
    
    # Generate JWT secret
    JWT_SECRET=$(openssl rand -hex 32)
    echo "JWT_SECRET_KEY=$JWT_SECRET" > "$INSTALL_DIR/.env"
    
    # Create API keys file
    cat > "$INSTALL_DIR/security/api-keys.json" << EOF
{
  "version": "1.0",
  "updated_at": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "keys": []
}
EOF
    
    # Set secure permissions
    chown -R "$SERVICE_USER:$SERVICE_GROUP" "$INSTALL_DIR"
    chmod 700 "$INSTALL_DIR/security"
    chmod 600 "$INSTALL_DIR/security/api-keys.json"
    chmod 600 "$INSTALL_DIR/.env"
    
    # Create log directory
    mkdir -p /var/log/camera-service
    chown "$SERVICE_USER:$SERVICE_GROUP" /var/log/camera-service
    

    
    # Install Python dependencies
    cd "$INSTALL_DIR"
    python3 -m venv venv
    source venv/bin/activate
    pip install -r requirements.txt
    
    # Create systemd service
    cat > /etc/systemd/system/camera-service.service << EOF
[Unit]
Description=MediaMTX Camera Service
Documentation=https://github.com/your-repo/mediamtx-camera-service
After=network.target mediamtx.service
Wants=network.target

[Service]
Type=simple
User=$SERVICE_USER
Group=$SERVICE_GROUP
WorkingDirectory=$INSTALL_DIR
EnvironmentFile=$INSTALL_DIR/.env
ExecStart=$INSTALL_DIR/venv/bin/python -m src.camera_service.main
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal
SyslogIdentifier=camera-service

[Install]
WantedBy=multi-user.target
EOF
    
    # Enable and start camera service
    systemctl daemon-reload
    systemctl enable camera-service
    systemctl start camera-service
    
    log_success "Camera Service installed and started"
}

# Function to verify installation
verify_installation() {
    log_message "Verifying installation..."
    
    # Check services are running
    if systemctl is-active --quiet mediamtx; then
        log_success "MediaMTX service is running"
    else
        log_error "MediaMTX service is not running"
        return 1
    fi
    
    if systemctl is-active --quiet camera-service; then
        log_success "Camera service is running"
    else
        log_error "Camera service is not running"
        return 1
    fi
    
    # Check health endpoints
    sleep 5  # Wait for services to start
    
    if curl -f -s http://localhost:8003/health/ready >/dev/null; then
        log_success "Health endpoint is responding"
    else
        log_warning "Health endpoint not responding yet"
    fi
    
    if curl -f -s http://localhost:9997/v3/paths/list >/dev/null; then
        log_success "MediaMTX API is responding"
    else
        log_warning "MediaMTX API not responding yet"
    fi
    
    log_success "Installation verification completed"
}

# Main installation function
main() {
    log_message "Starting MediaMTX Camera Service installation..."
    log_message "================================================"
    
    # Install system dependencies
    install_system_dependencies
    
    # Create service user
    create_service_user
    
    # Install MediaMTX
    install_mediamtx
    
    # Install Camera Service
    install_camera_service
    
    # Verify installation
    verify_installation
    
    log_message "================================================"
    log_success "Installation completed successfully!"
    log_message "Services installed:"
    log_message "- MediaMTX server (port 8554, 8888, 8889, 9997)"
    log_message "- Camera Service (port 8002, 8003)"
    log_message "- Health endpoints available at http://localhost:8003/health/"
    log_message ""
    log_message "To check service status:"
    log_message "  systemctl status mediamtx"
    log_message "  systemctl status camera-service"
    log_message ""
    log_message "To view logs:"
    log_message "  journalctl -u mediamtx -f"
    log_message "  journalctl -u camera-service -f"
}

# Run main function
main "$@"