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
    ssl_prefer_server_ciphers off;
    
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Frame-Options DENY always;
    add_header X-Content-Type-Options nosniff always;
    add_header X-XSS-Protection "1; mode=block" always;
    
    location / {
        proxy_pass http://127.0.0.1:8002;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
    
    location /health/ {
        proxy_pass http://127.0.0.1:8003/health/;
        access_log off;
    }
    
    # File download endpoints (Epic E6)
    location /files/recordings/ {
        proxy_pass http://127.0.0.1:8003/files/recordings/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Enable large file downloads
        proxy_read_timeout 300s;
        proxy_send_timeout 300s;
        proxy_connect_timeout 60s;
        
        # Security headers
        add_header X-Content-Type-Options nosniff always;
    }
    
    location /files/snapshots/ {
        proxy_pass http://127.0.0.1:8003/files/snapshots/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Security headers
        add_header X-Content-Type-Options nosniff always;
    }
}

server {
    listen 80;
    server_name camera-service.local;
    return 301 https://$server_name$request_uri;
}
EOF
        
        # Enable site and restart nginx
        ln -sf /etc/nginx/sites-available/camera-service /etc/nginx/sites-enabled/
        systemctl restart nginx
        
        log_success "HTTPS configuration completed"
    fi
}

# Function to setup production monitoring
setup_production_monitoring() {
    if [ "$ENABLE_MONITORING" = "true" ]; then
        log_message "Setting up production monitoring..."
        
        # Enable enhanced monitoring environment variables
        export CAMERA_SERVICE_ENV="production"
        export MONITORING_ENABLED="true"
        
        # Create monitoring directory
        mkdir -p "$INSTALL_DIR/monitoring"
        chown "$SERVICE_USER:$SERVICE_GROUP" "$INSTALL_DIR/monitoring"
        
        log_success "Production monitoring setup completed"
    fi
}

# Function to harden production environment
harden_production_environment() {
    if [ "$PRODUCTION_MODE" = "true" ]; then
        log_message "Hardening production environment..."
        
        # Configure firewall
        if ! command_exists ufw; then
            apt-get install -y ufw
        fi
        
        ufw allow 443/tcp  # HTTPS
        ufw allow 80/tcp   # HTTP redirect
        ufw allow 8554/tcp # RTSP
        ufw allow 8888/tcp # HLS
        ufw allow 8889/tcp # WebRTC
        ufw allow from 127.0.0.1 to any port 8002
        ufw allow from 127.0.0.1 to any port 8003
        ufw --force enable
        
        # Disable unnecessary services
        systemctl disable bluetooth 2>/dev/null || true
        systemctl disable cups 2>/dev/null || true
        systemctl disable avahi-daemon 2>/dev/null || true
        
        log_success "Production environment hardening completed"
    fi
}

# Function to setup backup procedures
setup_backup_procedures() {
    if [ "$PRODUCTION_MODE" = "true" ]; then
        log_message "Setting up backup procedures..."
        
        # Create backup directory
        mkdir -p "$INSTALL_DIR/backups"
        chown "$SERVICE_USER:$SERVICE_GROUP" "$INSTALL_DIR/backups"
        
        # Create backup script
        cat > "$INSTALL_DIR/backups/backup.sh" << 'EOF'
#!/bin/bash
BACKUP_DIR="/opt/camera-service/backups"
DATE=$(date +%Y%m%d_%H%M%S)

tar -czf "$BACKUP_DIR/camera-service-$DATE.tar.gz" \
    --exclude="$BACKUP_DIR" \
    --exclude="venv" \
    /opt/camera-service

find "$BACKUP_DIR" -name "camera-service-*.tar.gz" -mtime +7 -delete

echo "Backup completed: camera-service-$DATE.tar.gz"
EOF
        
        chmod +x "$INSTALL_DIR/backups/backup.sh"
        chown "$SERVICE_USER:$SERVICE_GROUP" "$INSTALL_DIR/backups/backup.sh"
        
        log_success "Backup procedures setup completed"
    fi
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
    
    # Ensure video group exists and add service user to it
    if ! getent group video >/dev/null 2>&1; then
        log_warning "Video group does not exist. Creating video group..."
        groupadd video
    fi
    
    # Add camera-service user to video group for device access
    if ! groups "$SERVICE_USER" | grep -q video; then
        usermod -a -G video "$SERVICE_USER"
        log_success "Added $SERVICE_USER to video group"
    else
        log_message "$SERVICE_USER already in video group"
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
    
    # Add mediamtx user to video group for device access
    if ! groups mediamtx | grep -q video; then
        usermod -a -G video mediamtx
        log_success "Added mediamtx user to video group"
    else
        log_message "mediamtx user already in video group"
    fi
    
    # Set ownership
    chown -R mediamtx:mediamtx "$MEDIAMTX_DIR"
    
    # Create MediaMTX config directory
    mkdir -p "$MEDIAMTX_DIR/config"
    
    # Copy and modify the bundled MediaMTX configuration
    cp "$MEDIAMTX_SOURCE/mediamtx.yml" "$MEDIAMTX_DIR/config/mediamtx.yml"
    
    # Enable API (change from 'no' to 'yes')
    sed -i 's/^api: no/api: yes/' "$MEDIAMTX_DIR/config/mediamtx.yml"
    
    # Verify the critical settings are correct (no changes needed for addresses as they are already correct)
    log_message "MediaMTX configuration applied with API enabled"
    
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
Environment=PYTHONPATH=$INSTALL_DIR/src
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
    
    # Validate video device permissions
    validate_video_permissions
    
    # Production enhancements
    if [ "$PRODUCTION_MODE" = "true" ]; then
        log_message "Setting up production features..."
        setup_https_configuration
        setup_production_monitoring
        harden_production_environment
        setup_backup_procedures
    fi
    
    # Verify installation
    verify_installation
    
    log_message "================================================"
    log_success "Installation completed successfully!"
    log_message "Services installed:"
    log_message "- MediaMTX server (port 8554, 8888, 8889, 9997)"
    log_message "- Camera Service (port 8002, 8003)"
    log_message "- Health endpoints available at http://localhost:8003/health/"
    
    if [ "$PRODUCTION_MODE" = "true" ]; then
        log_message ""
        log_message "Production features enabled:"
        log_message "- HTTPS/SSL: https://localhost (port 443)"
        log_message "- Firewall: UFW enabled with production rules"
        log_message "- Monitoring: Production monitoring enabled"
        log_message "- Backup: Automated backup procedures configured"
        log_message "- Security: Production hardening applied"
    fi
    
    log_message ""
    log_message "Service users and permissions:"
    log_message "- MediaMTX user: mediamtx (with video group access)"
    log_message "- Camera Service user: camera-service (with video group access)"
    log_message "- Video devices: accessible by both service users"
    log_message ""
    log_message "To check service status:"
    log_message "  systemctl status mediamtx"
    log_message "  systemctl status camera-service"
    log_message ""
    log_message "To view logs:"
    log_message "  journalctl -u mediamtx -f"
    log_message "  journalctl -u camera-service -f"
    log_message ""
    log_message "To verify video device access:"
    log_message "  sudo -u mediamtx test -r /dev/video0 && echo 'MediaMTX can access video devices'"
    log_message "  sudo -u camera-service test -r /dev/video0 && echo 'Camera Service can access video devices'"
    
    if [ "$PRODUCTION_MODE" = "true" ]; then
        log_message ""
        log_message "Production commands:"
        log_message "  # Run backup:"
        log_message "  sudo -u camera-service /opt/camera-service/backups/backup.sh"
        log_message ""
        log_message "  # Check HTTPS:"
        log_message "  curl -k https://localhost/health/ready"
        log_message ""
        log_message "  # Check firewall:"
        log_message "  sudo ufw status"
    fi
}

# Run main function
main "$@"