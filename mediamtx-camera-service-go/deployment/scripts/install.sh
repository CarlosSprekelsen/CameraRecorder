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
BINARY_NAME="camera-service"

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
    
    # Get the script directory to find dependencies BEFORE changing directory
    # Use absolute path resolution to handle different calling contexts
    SCRIPT_DIR="$(dirname "$(readlink -f "${BASH_SOURCE[0]}" 2>/dev/null || echo "${BASH_SOURCE[0]}")")"
    PROJECT_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")"
    DEPENDENCIES_DIR="$PROJECT_ROOT/dependencies"
    
    # Copy MediaMTX from local dependencies
    cd "$MEDIAMTX_DIR"
    if [ ! -f "mediamtx" ]; then
        
        if [ -f "$DEPENDENCIES_DIR/mediamtx_v1.15.1_linux_amd64.tar.gz" ]; then
            cp "$DEPENDENCIES_DIR/mediamtx_v1.15.1_linux_amd64.tar.gz" .
            tar -xzf mediamtx_v1.15.1_linux_amd64.tar.gz
            rm -f mediamtx_v1.15.1_linux_amd64.tar.gz
            chmod +x mediamtx
            log_success "MediaMTX copied from local dependencies"
        else
            log_error "MediaMTX dependency not found at $DEPENDENCIES_DIR/mediamtx_v1.15.1_linux_amd64.tar.gz"
            exit 1
        fi
    else
        log_message "MediaMTX already exists"
    fi
    
    # Create MediaMTX configuration directory and copy default config
    mkdir -p "$MEDIAMTX_DIR/config"
    
    # Extract and use the default MediaMTX configuration from the downloaded package
    if [ -f "$DEPENDENCIES_DIR/mediamtx_v1.15.1_linux_amd64.tar.gz" ]; then
        log_message "Extracting default MediaMTX configuration..."
        tar -xf "$DEPENDENCIES_DIR/mediamtx_v1.15.1_linux_amd64.tar.gz" -C "$MEDIAMTX_DIR/config" mediamtx.yml
        log_success "Default MediaMTX configuration extracted"
        
        # CRITICAL: Enable API for our application integration
        log_message "Enabling MediaMTX API for application integration..."
        sed -i 's/api: no/api: yes/' "$MEDIAMTX_DIR/config/mediamtx.yml"
        log_success "MediaMTX API enabled for application integration"
    else
        log_error "MediaMTX package not found for configuration extraction"
        exit 1
    fi
    
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
    set -x  # Enable debug mode
    
    # Get the script directory to find source files
    log_message "Getting script directory..."
    SCRIPT_DIR="$(dirname "${BASH_SOURCE[0]}")"
    log_message "Initial SCRIPT_DIR: $SCRIPT_DIR"
    
    # Convert to absolute path if relative
    if [[ "$SCRIPT_DIR" != /* ]]; then
        log_message "Converting relative path to absolute..."
        # Use readlink to get absolute path without cd
        ABSOLUTE_SCRIPT="$(readlink -f "${BASH_SOURCE[0]}")"
        log_message "ABSOLUTE_SCRIPT: $ABSOLUTE_SCRIPT"
        if [[ -n "$ABSOLUTE_SCRIPT" ]]; then
            SCRIPT_DIR="$(dirname "$ABSOLUTE_SCRIPT")"
            log_message "Using readlink result: $SCRIPT_DIR"
        else
            # Fallback: use pwd and relative path
            SCRIPT_DIR="$(pwd)/$SCRIPT_DIR"
            log_message "Using fallback: $SCRIPT_DIR"
        fi
    fi
    
    # Clean up the path and get project root
    SCRIPT_DIR="$(realpath "$SCRIPT_DIR" 2>/dev/null || echo "$SCRIPT_DIR")"
    PROJECT_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")"
    
    # If the calculated project root doesn't have the expected structure,
    # try to find the actual project root by looking for go.mod
    if [[ ! -d "$PROJECT_ROOT/cmd" ]]; then
        log_message "Calculated project root doesn't have expected structure, searching for go.mod..."
        
        # Search for go.mod file starting from script location
        SEARCH_DIR="$SCRIPT_DIR"
        while [[ "$SEARCH_DIR" != "/" ]]; do
            if [[ -f "$SEARCH_DIR/../go.mod" ]]; then
                PROJECT_ROOT="$(dirname "$SEARCH_DIR")"
                log_message "Found go.mod at: $PROJECT_ROOT"
                break
            fi
            SEARCH_DIR="$(dirname "$SEARCH_DIR")"
        done
        
        # If still not found, try to find it from common locations
        if [[ ! -d "$PROJECT_ROOT/cmd" ]]; then
            log_message "Searching for project in common locations..."
            for possible_root in "/home/carlossprekelsen/CameraRecorder/mediamtx-camera-service-go" "/opt/camera-service" "$(dirname "$(dirname "$(dirname "$(realpath "${BASH_SOURCE[0]}" 2>/dev/null || echo "${BASH_SOURCE[0]}")")")")"; do
                if [[ -d "$possible_root/cmd" && -f "$possible_root/go.mod" ]]; then
                    PROJECT_ROOT="$possible_root"
                    log_message "Found project at: $PROJECT_ROOT"
                    break
                fi
            done
        fi
    fi
    
    # Final verification
    if [[ ! -d "$PROJECT_ROOT/cmd" ]]; then
        log_error "Project root verification failed: $PROJECT_ROOT/cmd not found"
        log_message "Current working directory: $(pwd)"
        log_message "Script location: ${BASH_SOURCE[0]}"
        log_message "Resolved SCRIPT_DIR: $SCRIPT_DIR"
        log_message "Calculated PROJECT_ROOT: $PROJECT_ROOT"
        log_message "Please run the script from the project root directory"
        exit 1
    fi
    
    # Debug output
    log_message "Script directory: $SCRIPT_DIR"
    log_message "Project root: $PROJECT_ROOT"
    
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
    cd "$PROJECT_ROOT"
    
    # Set Go environment
    export PATH=$PATH:/usr/local/go/bin
    export GOPATH="$PROJECT_ROOT"
    export GOCACHE="$PROJECT_ROOT/.cache"
    
    # Download dependencies
    go mod download
    
    # Build the binary
    go build -o "$BINARY_NAME" cmd/server/main.go
    
    # Move binary to installation directory
    mv "$BINARY_NAME" "$INSTALL_DIR/"
    
    # Make binary executable
    chmod +x "$INSTALL_DIR/$BINARY_NAME"
    
    # Set ownership
    chown "$SERVICE_USER:$SERVICE_GROUP" "$INSTALL_DIR/$BINARY_NAME"
    
    # Copy UltraEfficient configuration for edge/IoT devices
    log_message "Installing UltraEfficient configuration for edge/IoT devices..."
    
    # Copy the UltraEfficient configuration
    if [ -f "$PROJECT_ROOT/config/ultra-efficient-edge.yaml" ]; then
        cp "$PROJECT_ROOT/config/ultra-efficient-edge.yaml" "$INSTALL_DIR/config/default.yaml"
        log_success "UltraEfficient configuration copied"
    else
        log_warning "UltraEfficient configuration not found, using fallback configuration"
        
        # Create fallback UltraEfficient configuration
        cat > "$INSTALL_DIR/config/default.yaml" << 'EOF'
# UltraEfficient Configuration for Edge/IoT Devices
# Optimized for minimal power consumption and resource usage

server:
  host: "0.0.0.0"
  port: 8002
  websocket_path: "/ws"
  max_connections: 10
  read_timeout: 10s
  write_timeout: 5s
  ping_interval: 60s
  pong_wait: 120s
  max_message_size: 524288
  read_buffer_size: 1024
  write_buffer_size: 1024
  shutdown_timeout: 30s
  client_cleanup_timeout: 10s
  auto_close_after: 0s

mediamtx:
  host: "localhost"
  api_port: 9997
  rtsp_port: 8554
  webrtc_port: 8889
  hls_port: 8888
  config_path: "/opt/mediamtx/config/mediamtx.yml"
  recordings_path: "/opt/camera-service/recordings"
  snapshots_path: "/opt/camera-service/snapshots"
  health_check_interval: 60
  health_failure_threshold: 5
  health_circuit_breaker_timeout: 120
  health_max_backoff_interval: 600
  health_recovery_confirmation_threshold: 3
  backoff_base_multiplier: 1.5
  backoff_jitter_range: [0.1, 0.2]
  process_termination_timeout: 15.0
  process_kill_timeout: 10.0
  health_check_timeout: 10s
  
  # CRITICAL: Stream readiness configuration (prevents panic)
  stream_readiness:
    timeout: 60.0
    retry_attempts: 2
    retry_delay: 2.0
    check_interval: 2.0
    enable_progress_notifications: false
    graceful_fallback: true
    max_check_interval: 2.0
    initial_check_interval: 0.2
    controller_ticker_interval: 0.1        # CRITICAL: Prevents panic
    stream_manager_ticker_interval: 0.1   # CRITICAL: Prevents panic
    path_manager_retry_intervals: [0.1, 0.2, 0.4, 0.8]

camera:
  poll_interval: 2.0
  detection_timeout: 15.0
  device_range: [0, 9]
  enable_capability_detection: false
  auto_start_streams: false
  capability_timeout: 10.0
  capability_retry_interval: 2.0
  capability_max_retries: 2
  discovery_mode: "event-first"
  fallback_poll_interval: 90.0
  max_event_handler_goroutines: 10
  event_handler_timeout: 5s

logging:
  level: "error"
  format: "json"
  file_enabled: true
  file_path: "/var/log/camera-service.log"
  max_file_size: 5242880
  backup_count: 3
  console_enabled: false

recording:
  enabled: false
  format: "mp4"
  quality: "low"
  segment_duration: 600
  max_segment_size: 52428800
  auto_cleanup: true
  cleanup_interval: 7200
  max_age: 2592000
  max_size: 536870912
  default_rotation_size: 52428800
  default_max_duration: 12h
  default_retention_days: 30
  max_restart_count: 3
  process_timeout: 5s

snapshots:
  enabled: true
  format: "jpeg"
  quality: 70
  max_width: 1280
  max_height: 720
  auto_cleanup: true
  cleanup_interval: 7200
  max_age: 2592000
  max_count: 500

performance:
  response_time_targets:
    snapshot_capture: 5.0
    recording_start: 10.0
    recording_stop: 5.0
    file_listing: 2.0
  snapshot_tiers:
    tier1_usb_direct_timeout: 2.0
    tier2_rtsp_ready_check_timeout: 5.0
    tier3_activation_timeout: 10.0
    tier3_activation_trigger_timeout: 20.0
    total_operation_timeout: 30.0
    immediate_response_threshold: 1.0
    acceptable_response_threshold: 5.0
    slow_response_threshold: 10.0
  optimization:
    enable_caching: false
    cache_ttl: 60
    max_concurrent_operations: 3
    connection_pool_size: 2
  monitoring_thresholds:
    memory_usage_percent: 90.0
    error_rate_percent: 5.0
    average_response_time_ms: 1000.0
    active_connections_limit: 900
    goroutines_limit: 1000
  debounce:
    health_monitor_seconds: 15
    storage_monitor_seconds: 30
    performance_monitor_seconds: 45

security:
  rate_limit_requests: 50
  rate_limit_window: 2m
  jwt_secret_key: "edge-device-secret-key-change-in-production"
  jwt_expiry_hours: 48

storage:
  warn_percent: 70
  block_percent: 85
  default_path: "/opt/camera-service/recordings"
  fallback_path: "/tmp/recordings"

retention_policy:
  enabled: true
  type: "age"
  max_age_days: 30
  max_size_gb: 0.5
  auto_cleanup: true

external_discovery:
  enabled: false
  scan_interval: 30
  timeout: 5.0
  max_concurrent_scans: 2
  generic_uav:
    enabled: false
    common_ports: [554, 8554]
    stream_paths: ["/stream", "/live", "/video"]
    known_ips: []

health_port: 8080
EOF
    fi
    
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
    log_message "Starting camera service installation..."
    if ! install_camera_service; then
        log_error "Camera service installation failed"
        exit 1
    fi
    log_success "Camera service installation completed"
    
    # Setup HTTPS if enabled
    setup_https_configuration
    
    # Validate video permissions
    validate_video_permissions
    
    # Verify installation
    verify_installation
    
    log_success "MediaMTX Camera Service (Go) installation completed successfully!"
}

# Run main function
main "$@"
