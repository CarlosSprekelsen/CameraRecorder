#!/bin/bash

# Backup and Recovery Script
# Uses existing backup mechanisms from the camera service

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
BACKUP_DIR="/opt/camera-service/backups"
LOG_DIR="/var/log/camera-service"

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

# Function to setup backup procedures
setup_backup_procedures() {
    log_message "Setting up backup procedures..."
    
    # Create backup directory
    mkdir -p "$BACKUP_DIR"
    chown "$SERVICE_USER:$SERVICE_GROUP" "$BACKUP_DIR"
    
    # Create comprehensive backup script
    cat > "$BACKUP_DIR/backup.sh" << 'EOF'
#!/bin/bash

# Comprehensive backup script for camera service

BACKUP_DIR="/opt/camera-service/backups"
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="camera-service-$DATE.tar.gz"
LOG_FILE="$BACKUP_DIR/backup.log"

# Log function
log() {
    echo "$(date): $1" >> "$LOG_FILE"
    echo "$1"
}

log "Starting backup: $BACKUP_FILE"

# Stop services for consistent backup
log "Stopping services..."
systemctl stop camera-service
systemctl stop mediamtx

# Create backup
log "Creating backup archive..."
tar -czf "$BACKUP_DIR/$BACKUP_FILE" \
    --exclude="$BACKUP_DIR" \
    --exclude="venv" \
    --exclude="*.log" \
    /opt/camera-service \
    /etc/systemd/system/camera-service.service \
    /etc/systemd/system/mediamtx.service \
    /etc/systemd/system/camera-monitoring.service

# Restart services
log "Restarting services..."
systemctl start mediamtx
systemctl start camera-service

# Verify backup
if [ -f "$BACKUP_DIR/$BACKUP_FILE" ]; then
    BACKUP_SIZE=$(du -h "$BACKUP_DIR/$BACKUP_FILE" | cut -f1)
    log "Backup completed successfully: $BACKUP_FILE ($BACKUP_SIZE)"
else
    log "ERROR: Backup file not created"
    exit 1
fi

# Cleanup old backups (keep last 7 days)
log "Cleaning up old backups..."
find "$BACKUP_DIR" -name "camera-service-*.tar.gz" -mtime +7 -delete

# List remaining backups
log "Remaining backups:"
ls -la "$BACKUP_DIR"/camera-service-*.tar.gz 2>/dev/null || log "No backup files found"

log "Backup process completed"
EOF
    
    chmod +x "$BACKUP_DIR/backup.sh"
    chown "$SERVICE_USER:$SERVICE_GROUP" "$BACKUP_DIR/backup.sh"
    
    log_success "Backup procedures configured"
}

# Function to setup recovery procedures
setup_recovery_procedures() {
    log_message "Setting up recovery procedures..."
    
    # Create recovery script
    cat > "$BACKUP_DIR/recover.sh" << 'EOF'
#!/bin/bash

# Recovery script for camera service

BACKUP_DIR="/opt/camera-service/backups"
LOG_FILE="$BACKUP_DIR/recovery.log"

# Log function
log() {
    echo "$(date): $1" >> "$LOG_FILE"
    echo "$1"
}

# Check if backup file provided
if [ -z "$1" ]; then
    echo "Usage: $0 <backup-file>"
    echo "Available backups:"
    ls -la "$BACKUP_DIR"/camera-service-*.tar.gz 2>/dev/null || echo "No backup files found"
    exit 1
fi

BACKUP_FILE="$1"

if [ ! -f "$BACKUP_FILE" ]; then
    log "ERROR: Backup file not found: $BACKUP_FILE"
    exit 1
fi

log "Starting recovery from: $BACKUP_FILE"

# Stop services
log "Stopping services..."
systemctl stop camera-service
systemctl stop mediamtx
systemctl stop camera-monitoring

# Create recovery backup
RECOVERY_BACKUP="$BACKUP_DIR/recovery-backup-$(date +%Y%m%d_%H%M%S).tar.gz"
log "Creating recovery backup: $RECOVERY_BACKUP"
tar -czf "$RECOVERY_BACKUP" /opt/camera-service /etc/systemd/system/camera-*.service

# Extract backup
log "Extracting backup..."
tar -xzf "$BACKUP_FILE" -C /

# Restore permissions
log "Restoring permissions..."
chown -R camera-service:camera-service /opt/camera-service
chmod +x /opt/camera-service/backups/*.sh

# Reload systemd
log "Reloading systemd..."
systemctl daemon-reload

# Start services
log "Starting services..."
systemctl start mediamtx
systemctl start camera-service
systemctl start camera-monitoring

# Verify recovery
log "Verifying recovery..."
sleep 10

if systemctl is-active --quiet camera-service && systemctl is-active --quiet mediamtx; then
    log "Recovery completed successfully"
else
    log "ERROR: Recovery failed - services not running"
    exit 1
fi

log "Recovery process completed"
EOF
    
    chmod +x "$BACKUP_DIR/recover.sh"
    chown "$SERVICE_USER:$SERVICE_GROUP" "$BACKUP_DIR/recover.sh"
    
    log_success "Recovery procedures configured"
}

# Function to setup automated backups
setup_automated_backups() {
    log_message "Setting up automated backups..."
    
    # Create cron job for daily backups
    cat > /etc/cron.d/camera-service-backup << EOF
# Daily backup for camera service
0 2 * * * camera-service /opt/camera-service/backups/backup.sh >> /var/log/camera-service/backup.log 2>&1
EOF
    
    chmod 644 /etc/cron.d/camera-service-backup
    
    log_success "Automated backups configured"
}

# Function to setup disaster recovery
setup_disaster_recovery() {
    log_message "Setting up disaster recovery..."
    
    # Create disaster recovery script
    cat > "$BACKUP_DIR/disaster_recovery.sh" << 'EOF'
#!/bin/bash

# Disaster recovery script for camera service

BACKUP_DIR="/opt/camera-service/backups"
LOG_FILE="$BACKUP_DIR/disaster_recovery.log"

log() {
    echo "$(date): $1" >> "$LOG_FILE"
    echo "$1"
}

log "Starting disaster recovery process"

# Check system status
log "Checking system status..."

# Check if services are running
if ! systemctl is-active --quiet camera-service; then
    log "WARNING: Camera service not running"
fi

if ! systemctl is-active --quiet mediamtx; then
    log "WARNING: MediaMTX service not running"
fi

# Check disk space
DISK_USAGE=$(df /opt/camera-service | tail -1 | awk '{print $5}' | sed 's/%//')
if [ "$DISK_USAGE" -gt 90 ]; then
    log "WARNING: Disk usage is ${DISK_USAGE}%"
fi

# Check log files for errors
ERROR_COUNT=$(grep -c "ERROR\|CRITICAL\|FATAL" /var/log/camera-service/camera-service.log 2>/dev/null || echo "0")
if [ "$ERROR_COUNT" -gt 0 ]; then
    log "WARNING: Found $ERROR_COUNT errors in logs"
fi

# Attempt service recovery
log "Attempting service recovery..."

# Restart services
systemctl restart mediamtx
systemctl restart camera-service
systemctl restart camera-monitoring

# Wait for services to start
sleep 15

# Verify recovery
if systemctl is-active --quiet camera-service && systemctl is-active --quiet mediamtx; then
    log "Disaster recovery completed successfully"
else
    log "ERROR: Disaster recovery failed"
    log "Consider manual recovery using: $BACKUP_DIR/recover.sh <backup-file>"
    exit 1
fi

log "Disaster recovery process completed"
EOF
    
    chmod +x "$BACKUP_DIR/disaster_recovery.sh"
    chown "$SERVICE_USER:$SERVICE_GROUP" "$BACKUP_DIR/disaster_recovery.sh"
    
    log_success "Disaster recovery configured"
}

# Function to test backup procedures
test_backup_procedures() {
    log_message "Testing backup procedures..."
    
    # Test backup script
    if [ -f "$BACKUP_DIR/backup.sh" ]; then
        log_message "Testing backup script..."
        sudo -u "$SERVICE_USER" "$BACKUP_DIR/backup.sh"
        
        # Check if backup was created
        if ls "$BACKUP_DIR"/camera-service-*.tar.gz >/dev/null 2>&1; then
            log_success "Backup test completed successfully"
        else
            log_error "Backup test failed"
            return 1
        fi
    else
        log_error "Backup script not found"
        return 1
    fi
}

# Main function
main() {
    log_message "Setting up backup and recovery procedures..."
    
    setup_backup_procedures
    setup_recovery_procedures
    setup_automated_backups
    setup_disaster_recovery
    test_backup_procedures
    
    log_success "Backup and recovery setup completed"
    log_message "Backup procedures:"
    log_message "- Manual backup: $BACKUP_DIR/backup.sh"
    log_message "- Recovery: $BACKUP_DIR/recover.sh <backup-file>"
    log_message "- Disaster recovery: $BACKUP_DIR/disaster_recovery.sh"
    log_message "- Automated backups: Daily at 2:00 AM"
}

# Run main function
main "$@"
