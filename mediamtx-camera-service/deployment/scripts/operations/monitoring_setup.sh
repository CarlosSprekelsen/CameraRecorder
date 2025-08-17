#!/bin/bash

# Operations Monitoring Setup Script
# Uses existing monitoring infrastructure from the camera service

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SERVICE_USER="camera-service"
SERVICE_GROUP="camera-service"
INSTALL_DIR="/opt/camera-service"

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

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to setup existing health monitoring
setup_health_monitoring() {
    log_message "Setting up existing health monitoring..."
    
    # The existing health server already provides:
    # - Health endpoints at /health/ready, /health/live
    # - REST API for monitoring
    # - Kubernetes readiness probes
    
    # Create monitoring directory
    mkdir -p "$INSTALL_DIR/monitoring"
    chown "$SERVICE_USER:$SERVICE_GROUP" "$INSTALL_DIR/monitoring"
    
    # Create monitoring configuration
    cat > "$INSTALL_DIR/monitoring/monitoring.conf" << EOF
# Production monitoring configuration
# Uses existing health monitoring from src/health_server.py

# Health check endpoints
HEALTH_READY_URL="http://localhost:8003/health/ready"
HEALTH_LIVE_URL="http://localhost:8003/health/live"

# MediaMTX monitoring endpoints
MEDIAMTX_API_URL="http://localhost:9997/v3/paths/list"
MEDIAMTX_HEALTH_URL="http://localhost:9997/v3/paths/list"

# Monitoring intervals
HEALTH_CHECK_INTERVAL=30
MEDIAMTX_CHECK_INTERVAL=60

# Alert thresholds
MAX_FAILURES=3
ALERT_COOLDOWN=300
EOF
    
    chown "$SERVICE_USER:$SERVICE_GROUP" "$INSTALL_DIR/monitoring/monitoring.conf"
    
    log_success "Health monitoring configured"
}

# Function to setup MediaMTX monitoring
setup_mediamtx_monitoring() {
    log_message "Setting up MediaMTX monitoring..."
    
    # Use existing MediaMTX monitoring from src/mediamtx_wrapper/controller.py
    # This already includes:
    # - Health monitoring with circuit breaker
    # - Adaptive backoff and recovery
    # - Comprehensive logging
    
    # Create MediaMTX monitoring script
    cat > "$INSTALL_DIR/monitoring/mediamtx_monitor.sh" << 'EOF'
#!/bin/bash

# MediaMTX monitoring script using existing monitoring infrastructure

MEDIAMTX_API="http://localhost:9997/v3/paths/list"
LOG_FILE="/var/log/camera-service/mediamtx-monitor.log"

# Check MediaMTX health
check_mediamtx_health() {
    if curl -f -s "$MEDIAMTX_API" >/dev/null; then
        echo "$(date): MediaMTX health check PASSED" >> "$LOG_FILE"
        return 0
    else
        echo "$(date): MediaMTX health check FAILED" >> "$LOG_FILE"
        return 1
    fi
}

# Main monitoring loop
while true; do
    check_mediamtx_health
    sleep 60
done
EOF
    
    chmod +x "$INSTALL_DIR/monitoring/mediamtx_monitor.sh"
    chown "$SERVICE_USER:$SERVICE_GROUP" "$INSTALL_DIR/monitoring/mediamtx_monitor.sh"
    
    log_success "MediaMTX monitoring configured"
}

# Function to setup alerting
setup_alerting() {
    log_message "Setting up alerting system..."
    
    # Create alerting script
    cat > "$INSTALL_DIR/monitoring/alert.sh" << 'EOF'
#!/bin/bash

# Simple alerting script for production monitoring

ALERT_LOG="/var/log/camera-service/alerts.log"
SERVICE_NAME="$1"
STATUS="$2"
MESSAGE="$3"

echo "$(date): ALERT - $SERVICE_NAME: $STATUS - $MESSAGE" >> "$ALERT_LOG"

# In production, this would send to your alerting system
# Examples: email, Slack, PagerDuty, etc.
echo "ALERT: $SERVICE_NAME is $STATUS - $MESSAGE"
EOF
    
    chmod +x "$INSTALL_DIR/monitoring/alert.sh"
    chown "$SERVICE_USER:$SERVICE_GROUP" "$INSTALL_DIR/monitoring/alert.sh"
    
    log_success "Alerting system configured"
}

# Function to setup log monitoring
setup_log_monitoring() {
    log_message "Setting up log monitoring..."
    
    # Create log monitoring script
    cat > "$INSTALL_DIR/monitoring/log_monitor.sh" << 'EOF'
#!/bin/bash

# Log monitoring script for production

LOG_FILE="/var/log/camera-service/camera-service.log"
ALERT_SCRIPT="/opt/camera-service/monitoring/alert.sh"

# Monitor for critical errors
monitor_logs() {
    tail -f "$LOG_FILE" | while read line; do
        # Check for critical errors
        if echo "$line" | grep -q "CRITICAL\|ERROR\|FATAL"; then
            "$ALERT_SCRIPT" "CameraService" "ERROR" "Critical error detected in logs"
        fi
        
        # Check for service failures
        if echo "$line" | grep -q "service.*failed\|connection.*failed"; then
            "$ALERT_SCRIPT" "CameraService" "WARNING" "Service failure detected"
        fi
    done
}

monitor_logs
EOF
    
    chmod +x "$INSTALL_DIR/monitoring/log_monitor.sh"
    chown "$SERVICE_USER:$SERVICE_GROUP" "$INSTALL_DIR/monitoring/log_monitor.sh"
    
    log_success "Log monitoring configured"
}

# Function to create monitoring service
create_monitoring_service() {
    log_message "Creating monitoring service..."
    
    cat > /etc/systemd/system/camera-monitoring.service << EOF
[Unit]
Description=Camera Service Monitoring
After=camera-service.service
Wants=camera-service.service

[Service]
Type=simple
User=$SERVICE_USER
Group=$SERVICE_GROUP
WorkingDirectory=$INSTALL_DIR/monitoring
ExecStart=$INSTALL_DIR/monitoring/mediamtx_monitor.sh
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal
SyslogIdentifier=camera-monitoring

[Install]
WantedBy=multi-user.target
EOF
    
    systemctl daemon-reload
    systemctl enable camera-monitoring
    
    log_success "Monitoring service created"
}

# Main function
main() {
    log_message "Setting up operations monitoring..."
    
    setup_health_monitoring
    setup_mediamtx_monitoring
    setup_alerting
    setup_log_monitoring
    create_monitoring_service
    
    log_success "Operations monitoring setup completed"
    log_message "Monitoring services:"
    log_message "- Health monitoring: http://localhost:8003/health/"
    log_message "- MediaMTX monitoring: http://localhost:9997/v3/paths/list"
    log_message "- Log monitoring: /var/log/camera-service/"
    log_message "- Alerting: /opt/camera-service/monitoring/alert.sh"
}

# Run main function
main "$@"
