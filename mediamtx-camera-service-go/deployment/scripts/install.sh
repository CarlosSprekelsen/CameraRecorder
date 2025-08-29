#!/bin/bash

# MediaMTX Camera Service (Go) Installation Script
# Installs MediaMTX server and Camera Service with security configuration
# Adapted from Python version for Go implementation

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
BINARY_NAME="mediamtx-camera-service-go"

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

# Production mode detection and configuration
PRODUCTION_MODE=${PRODUCTION_MODE:-false}
ENABLE_HTTPS=${ENABLE_HTTPS:-false}
ENABLE_MONITORING=${ENABLE_MONITORING:-false}

# Production configuration
if [ "$PRODUCTION_MODE" = "true" ]; then
    log_message "Running in PRODUCTION mode"
    ENABLE_HTTPS=true
    ENABLE_MONITORING=true
    SECURITY_LEVEL="high"
else
    log_message "Running in DEVELOPMENT mode"
    SECURITY_LEVEL="standard"
fi

# Check if running as root
if [[ $EUID -ne 0 ]]; then
    log_error "This script must be run as root (use sudo)"
    exit 1
fi

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to setup HTTPS configuration
setup_https_configuration() {
    if [ "$ENABLE_HTTPS" = "true" ]; then
        log_message "Setting up HTTPS configuration..."
        
        # Create SSL directory
        mkdir -p "$INSTALL_DIR/ssl"
        
        # Generate SSL certificates
        log_message "Generating SSL certificates..."
        openssl req -x509 -newkey rsa:4096 -keyout "$INSTALL_DIR/ssl/key.pem" \
            -out "$INSTALL_DIR/ssl/cert.pem" -days 365 -nodes \
            -subj "/C=US/ST=State/L=City/O=Organization/CN=localhost"
        
        # Set proper permissions
        chmod 600 "$INSTALL_DIR/ssl/key.pem"
        chmod 644 "$INSTALL_DIR/ssl/cert.pem"
        chown -R "$SERVICE_USER:$SERVICE_GROUP" "$INSTALL_DIR/ssl"
        
        # Install nginx if not present
        if ! command_exists nginx; then
            apt-get install -y nginx
        fi
        
        # Create nginx configuration
        cat > /etc/nginx/sites-available/camera-service << 'EOF'
server {
    listen 443 ssl;
    server_name camera-service.local;
    
    ssl_certificate /opt/camera-service/ssl/cert.pem;
    ssl_certificate_key /opt/camera-service/ssl/key.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
EOF
        
        # Enable site
        ln -sf /etc/nginx/sites-available/camera-service /etc/nginx/sites-enabled/
        systemctl reload nginx
        
        log_success "HTTPS configuration completed"
    fi
}

# Function to install system dependencies
install_system_dependencies() {
    log_message "Installing system dependencies..."
    
    # Update package list
    apt-get update
    
    # Install required packages
    apt-get install -y \
        curl \
        wget \
        git \
        build-essential \
        pkg-config \
        libv4l-dev \
        v4l-utils \
        ffmpeg \
        systemd \
        systemd-sysv
    
    # Install Go if not present
    if ! command_exists go; then
        log_message "Installing Go..."
        wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
        tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
        echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
        export PATH=$PATH:/usr/local/go/bin
        rm go1.21.0.linux-amd64.tar.gz
        log_success "Go installed"
    else
        log_message "Go already installed"
    fi
    
    log_success "System dependencies installed"
}

# Function to create service user and group
create_service_user() {
    log_message "Creating service user and group..."
    
    # Create group if it doesn't exist
    if ! getent group "$SERVICE_GROUP" >/dev/null 2>&1; then
        groupadd "$SERVICE_GROUP"
        log_success "Group created: $SERVICE_GROUP"
    else
        log_message "Group already exists: $SERVICE_GROUP"
    fi
    
    # Create user if it doesn't exist
    if ! getent passwd "$SERVICE_USER" >/dev/null 2>&1; then
        useradd -r -s /bin/false -g "$SERVICE_GROUP" -d "$INSTALL_DIR" "$SERVICE_USER"
        log_success "User created: $SERVICE_USER"
    else
        log_message "User already exists: $SERVICE_USER"
    fi
    
    # Add user to video group for camera access
    usermod -a -G video "$SERVICE_USER"
    
    log_success "Service user setup completed"
}

# Function to install MediaMTX server
install_mediamtx() {
    log_message "Installing MediaMTX server..."
    
    # Create MediaMTX directory
    mkdir -p "$MEDIAMTX_DIR"
    
    # Download MediaMTX
    cd "$MEDIAMTX_DIR"
    if [ ! -f "mediamtx" ]; then
        wget -O mediamtx https://github.com/bluenviron/mediamtx/releases/latest/download/mediamtx_linux_amd64
        chmod +x mediamtx
        log_success "MediaMTX downloaded"
    else
        log_message "MediaMTX already exists"
    fi
    
    # Create MediaMTX configuration
    cat > "$MEDIAMTX_DIR/config/mediamtx.yml" << 'EOF'
# MediaMTX Configuration
paths:
  all:
    source: publisher
    sourceOnDemand: yes
    sourceOnDemandStartTimeout: 10s
    sourceOnDemandCloseAfter: 10s

# API Configuration
api: yes
apiAddress: :9997

# RTSP Configuration
rtsp: yes
rtspAddress: :8554

# RTMP Configuration
rtmp: yes
rtmpAddress: :1935

# HLS Configuration
hls: yes
hlsAddress: :8888

# WebRTC Configuration
webrtc: yes
webrtcAddress: :8889

# Logging
logLevel: info
logDestinations: stdout
EOF
    
    # Set ownership
    chown -R mediamtx:mediamtx "$MEDIAMTX_DIR"
    
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
    
    # Enable and start MediaMTX service (idempotent)
    systemctl daemon-reload
    
    # Check if service is already enabled
    if ! systemctl is-enabled --quiet mediamtx 2>/dev/null; then
        systemctl enable mediamtx
        log_success "MediaMTX service enabled"
    else
        log_message "MediaMTX service already enabled"
    fi
    
    # Check if service is already running
    if ! systemctl is-active --quiet mediamtx 2>/dev/null; then
        systemctl start mediamtx
        log_success "MediaMTX service started"
    else
        log_message "MediaMTX service already running"
    fi
    
    log_success "MediaMTX server installed and started"
}

# Function to build and install Camera Service
install_camera_service() {
    log_message "Installing Camera Service..."
    
    # Get the script directory to find source files
    SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    PROJECT_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")"
    
    # Create installation directory
    mkdir -p "$INSTALL_DIR"
    
    # Copy source files
    if [ -d "$PROJECT_ROOT/cmd" ]; then
        cp -r "$PROJECT_ROOT/cmd" "$INSTALL_DIR/"
        log_success "Command files copied"
    else
        log_error "Command directory not found at $PROJECT_ROOT/cmd"
        exit 1
    fi
    
    if [ -d "$PROJECT_ROOT/internal" ]; then
        cp -r "$PROJECT_ROOT/internal" "$INSTALL_DIR/"
        log_success "Internal files copied"
    else
        log_error "Internal directory not found at $PROJECT_ROOT/internal"
        exit 1
    fi
    
    if [ -f "$PROJECT_ROOT/go.mod" ]; then
        cp "$PROJECT_ROOT/go.mod" "$INSTALL_DIR/"
        log_success "Go module file copied"
    else
        log_error "Go module file not found at $PROJECT_ROOT/go.mod"
        exit 1
    fi
    
    if [ -f "$PROJECT_ROOT/go.sum" ]; then
        cp "$PROJECT_ROOT/go.sum" "$INSTALL_DIR/"
        log_success "Go sum file copied"
    fi
    
    # Create configuration directory
    mkdir -p "$INSTALL_DIR/config"
    
    # Create required directories with proper permissions
    log_message "Creating required directories..."
    mkdir -p "$INSTALL_DIR/recordings" "$INSTALL_DIR/snapshots"
    chown "$SERVICE_USER:$SERVICE_GROUP" "$INSTALL_DIR/recordings" "$INSTALL_DIR/snapshots"
    
    # Build the Go application
    log_message "Building Go application..."
    cd "$INSTALL_DIR"
    
    # Set Go environment
    export PATH=$PATH:/usr/local/go/bin
    export GOPATH="$INSTALL_DIR"
    export GOCACHE="$INSTALL_DIR/.cache"
    
    # Download dependencies
    go mod download
    
    # Build the binary
    go build -o "$BINARY_NAME" cmd/server/main.go
    
    # Make binary executable
    chmod +x "$BINARY_NAME"
    
    # Set ownership
    chown "$SERVICE_USER:$SERVICE_GROUP" "$BINARY_NAME"
    
    # Create default configuration
    cat > "$INSTALL_DIR/config/default.yaml" << 'EOF'
server:
  host: "0.0.0.0"
  port: 8080

mediamtx:
  host: "localhost"
  api_port: 9997
  health_check_url: "http://localhost:9997/v3/paths/list"

camera:
  scan_interval: "30s"
  device_paths:
    - "/dev/video0"
    - "/dev/video1"

logging:
  level: "info"
  format: "json"

security:
  jwt_secret_key: "your-secret-key-change-in-production"
  rate_limit_requests: 100
  rate_limit_window: "1m"

recording:
  default_rotation_size: 100
  segment_duration: "1h"
  max_segment_size: 1000
  cleanup_interval: "24h"

snapshots:
  interval: "5s"
  cleanup_interval: "1h"

storage:
  warn_percent: 80
  block_percent: 90
  default_path: "/opt/camera-service/recordings"
  fallback_path: "/tmp/recordings"
EOF
    
    # Create systemd service
    cat > /etc/systemd/system/camera-service.service << EOF
[Unit]
Description=MediaMTX Camera Service (Go)
Documentation=https://github.com/your-repo/mediamtx-camera-service-go
After=network.target mediamtx.service
Wants=network.target

[Service]
Type=simple
User=$SERVICE_USER
Group=$SERVICE_GROUP
WorkingDirectory=$INSTALL_DIR
Environment=GOPATH=$INSTALL_DIR
Environment=GOCACHE=$INSTALL_DIR/.cache
ExecStart=$INSTALL_DIR/$BINARY_NAME
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal
SyslogIdentifier=camera-service

[Install]
WantedBy=multi-user.target
EOF
    
    # Enable and start camera service (idempotent)
    systemctl daemon-reload
    
    # Check if service is already enabled
    if ! systemctl is-enabled --quiet camera-service 2>/dev/null; then
        systemctl enable camera-service
        log_success "Camera service enabled"
    else
        log_message "Camera service already enabled"
    fi
    
    # Check if service is already running
    if ! systemctl is-active --quiet camera-service 2>/dev/null; then
        systemctl start camera-service
        log_success "Camera service started"
    else
        log_message "Camera service already running"
    fi
    
    log_success "Camera Service installed and started"
}

# Function to validate video device permissions
validate_video_permissions() {
    log_message "Validating video device permissions..."
    
    # Check if video devices exist
    if ! ls /dev/video* >/dev/null 2>&1; then
        log_warning "No video devices found at /dev/video*"
        return 0
    fi
    
    # Check video group exists
    if ! getent group video >/dev/null 2>&1; then
        log_error "Video group does not exist"
        return 1
    fi
    
    # Check mediamtx user can access video devices
    if ! sudo -u mediamtx test -r /dev/video0 2>/dev/null; then
        log_error "MediaMTX user cannot access video devices"
        log_message "Adding mediamtx user to video group..."
        usermod -a -G video mediamtx
    else
        log_success "MediaMTX user can access video devices"
    fi
    
    # Check camera-service user can access video devices
    if ! sudo -u camera-service test -r /dev/video0 2>/dev/null; then
        log_error "Camera service user cannot access video devices"
        log_message "Adding camera-service user to video group..."
        usermod -a -G video camera-service
    else
        log_success "Camera service user can access video devices"
    fi
    
    # Verify video device permissions
    VIDEO_PERMS=$(ls -l /dev/video0 | awk '{print $1}')
    if [[ "$VIDEO_PERMS" == "crw-rw----+" ]]; then
        log_success "Video device permissions are correct: $VIDEO_PERMS"
    else
        log_warning "Video device permissions may need adjustment: $VIDEO_PERMS"
        log_message "Expected: crw-rw----+, Found: $VIDEO_PERMS"
    fi
    
    log_success "Video device permissions validation completed"
}

# Function to verify installation
verify_installation() {
    log_message "Verifying installation..."
    
    # Check if services are running
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
    
    # Check if binary exists
    if [ -f "$INSTALL_DIR/$BINARY_NAME" ]; then
        log_success "Camera service binary exists"
    else
        log_error "Camera service binary not found"
        return 1
    fi
    
    # Check if configuration exists
    if [ -f "$INSTALL_DIR/config/default.yaml" ]; then
        log_success "Configuration file exists"
    else
        log_error "Configuration file not found"
        return 1
    fi
    
    # Test MediaMTX API
    if curl -s http://localhost:9997/v3/paths/list >/dev/null; then
        log_success "MediaMTX API is accessible"
    else
        log_warning "MediaMTX API is not accessible"
    fi
    
    log_success "Installation verification completed"
}

# Main installation function
main() {
    log_message "Starting MediaMTX Camera Service (Go) installation..."
    
    # Install system dependencies
    install_system_dependencies
    
    # Create service user
    create_service_user
    
    # Install MediaMTX
    install_mediamtx
    
    # Install Camera Service
    install_camera_service
    
    # Setup HTTPS if enabled
    setup_https_configuration
    
    # Validate video permissions
    validate_video_permissions
    
    # Verify installation
    verify_installation
    
    log_success "MediaMTX Camera Service (Go) installation completed successfully!"
    log_message "Services are running and ready to use."
    log_message "MediaMTX API: http://localhost:9997"
    log_message "Camera Service: http://localhost:8080"
}

# Run main function
main "$@"
