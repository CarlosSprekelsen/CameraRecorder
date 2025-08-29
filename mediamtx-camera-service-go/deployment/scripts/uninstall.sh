#!/bin/bash

# MediaMTX Camera Service (Go) Uninstall Script
# Removes all components of the camera service installation
# Used for testing, maintenance, and clean reinstallation

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

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to check if running as root
check_root() {
    if [[ $EUID -ne 0 ]]; then
        log_error "This script must be run as root (use sudo)"
        exit 1
    fi
}

# Function to stop and disable camera service
stop_camera_service() {
    log_message "Stopping camera service..."
    
    if systemctl list-unit-files | grep -q "$SERVICE_NAME"; then
        # Stop the service if it's running
        if systemctl is-active --quiet "$SERVICE_NAME"; then
            systemctl stop "$SERVICE_NAME"
            log_success "Camera service stopped"
        else
            log_message "Camera service was not running"
        fi
        
        # Disable the service
        systemctl disable "$SERVICE_NAME"
        log_success "Camera service disabled"
    else
        log_message "Camera service not found in systemd"
    fi
}

# Function to remove systemd service file
remove_service_file() {
    log_message "Removing systemd service file..."
    
    local service_file="/etc/systemd/system/$SERVICE_NAME.service"
    
    if [[ -f "$service_file" ]]; then
        rm -f "$service_file"
        log_success "Service file removed: $service_file"
    else
        log_message "Service file not found: $service_file"
    fi
}

# Function to remove installation directory
remove_installation_directory() {
    log_message "Removing installation directory..."
    
    if [[ -d "$INSTALL_DIR" ]]; then
        rm -rf "$INSTALL_DIR"
        log_success "Installation directory removed: $INSTALL_DIR"
    else
        log_message "Installation directory not found: $INSTALL_DIR"
    fi
}

# Function to remove MediaMTX installation
remove_mediamtx() {
    log_message "Removing MediaMTX installation..."
    
    # Stop MediaMTX service
    if systemctl list-unit-files | grep -q "mediamtx"; then
        if systemctl is-active --quiet mediamtx; then
            systemctl stop mediamtx
            log_success "MediaMTX service stopped"
        else
            log_message "MediaMTX service was not running"
        fi
        
        systemctl disable mediamtx
        log_success "MediaMTX service disabled"
    else
        log_message "MediaMTX service not found in systemd"
    fi
    
    # Remove MediaMTX service file
    local mediamtx_service="/etc/systemd/system/mediamtx.service"
    if [[ -f "$mediamtx_service" ]]; then
        rm -f "$mediamtx_service"
        log_success "MediaMTX service file removed"
    else
        log_message "MediaMTX service file not found"
    fi
    
    # Remove MediaMTX directory
    if [[ -d "$MEDIAMTX_DIR" ]]; then
        rm -rf "$MEDIAMTX_DIR"
        log_success "MediaMTX directory removed: $MEDIAMTX_DIR"
    else
        log_message "MediaMTX directory not found: $MEDIAMTX_DIR"
    fi
}

# Function to remove service user and group
remove_service_user() {
    log_message "Removing service user and group..."
    
    # Remove user from video group
    if getent passwd "$SERVICE_USER" >/dev/null 2>&1; then
        usermod -G "$SERVICE_USER" "$SERVICE_USER" 2>/dev/null || true
        log_message "Removed $SERVICE_USER from video group"
    fi
    
    # Delete user if it exists
    if getent passwd "$SERVICE_USER" >/dev/null 2>&1; then
        userdel "$SERVICE_USER" 2>/dev/null || true
        log_success "User removed: $SERVICE_USER"
    else
        log_message "User not found: $SERVICE_USER"
    fi
    
    # Delete group if it exists
    if getent group "$SERVICE_GROUP" >/dev/null 2>&1; then
        groupdel "$SERVICE_GROUP" 2>/dev/null || true
        log_success "Group removed: $SERVICE_GROUP"
    else
        log_message "Group not found: $SERVICE_GROUP"
    fi
}

# Function to remove monitoring service
remove_monitoring_service() {
    log_message "Removing monitoring service..."
    
    # Stop monitoring service
    if systemctl list-unit-files | grep -q "camera-monitoring"; then
        if systemctl is-active --quiet camera-monitoring; then
            systemctl stop camera-monitoring
            log_success "Monitoring service stopped"
        else
            log_message "Monitoring service was not running"
        fi
        
        systemctl disable camera-monitoring
        log_success "Monitoring service disabled"
    else
        log_message "Monitoring service not found in systemd"
    fi
    
    # Remove monitoring service file
    local monitoring_service="/etc/systemd/system/camera-monitoring.service"
    if [[ -f "$monitoring_service" ]]; then
        rm -f "$monitoring_service"
        log_success "Monitoring service file removed"
    else
        log_message "Monitoring service file not found"
    fi
}

# Function to remove environment files
remove_environment_files() {
    log_message "Removing environment files..."
    
    local env_files=(
        "/etc/systemd/system/camera-service.env"
        "/etc/systemd/system/camera-monitoring.service"
    )
    
    for env_file in "${env_files[@]}"; do
        if [[ -f "$env_file" ]]; then
            rm -f "$env_file"
            log_success "Environment file removed: $env_file"
        else
            log_message "Environment file not found: $env_file"
        fi
    done
}

# Function to remove systemd symlinks
remove_systemd_symlinks() {
    log_message "Removing systemd service symlinks..."
    
    # Remove systemd service symlinks
    if [[ -L "/etc/systemd/system/multi-user.target.wants/camera-monitoring.service" ]]; then
        rm -f "/etc/systemd/system/multi-user.target.wants/camera-monitoring.service"
        log_success "Systemd service symlink removed: camera-monitoring.service"
    fi
    
    if [[ -L "/etc/systemd/system/multi-user.target.wants/camera-service.service" ]]; then
        rm -f "/etc/systemd/system/multi-user.target.wants/camera-service.service"
        log_success "Systemd service symlink removed: camera-service.service"
    fi
    
    if [[ -L "/etc/systemd/system/multi-user.target.wants/mediamtx.service" ]]; then
        rm -f "/etc/systemd/system/multi-user.target.wants/mediamtx.service"
        log_success "Systemd service symlink removed: mediamtx.service"
    fi
}

# Function to remove SSL certificates
remove_ssl_certificates() {
    log_message "Removing SSL certificates..."
    
    local ssl_dir="$INSTALL_DIR/ssl"
    if [[ -d "$ssl_dir" ]]; then
        rm -rf "$ssl_dir"
        log_success "SSL certificates removed: $ssl_dir"
    else
        log_message "SSL certificates not found: $ssl_dir"
    fi
}

# Function to remove nginx configuration
remove_nginx_configuration() {
    log_message "Removing nginx configuration..."
    
    # Remove nginx site configuration
    if [[ -f "/etc/nginx/sites-available/camera-service" ]]; then
        rm -f "/etc/nginx/sites-available/camera-service"
        log_success "Nginx site configuration removed"
    else
        log_message "Nginx site configuration not found"
    fi
    
    # Remove nginx site symlink
    if [[ -L "/etc/nginx/sites-enabled/camera-service" ]]; then
        rm -f "/etc/nginx/sites-enabled/camera-service"
        log_success "Nginx site symlink removed"
    else
        log_message "Nginx site symlink not found"
    fi
    
    # Reload nginx if it's running
    if systemctl is-active --quiet nginx; then
        systemctl reload nginx
        log_success "Nginx reloaded"
    fi
}

# Function to reload systemd
reload_systemd() {
    log_message "Reloading systemd..."
    
    systemctl daemon-reload
    log_success "Systemd daemon reloaded"
}

# Function to verify uninstallation
verify_uninstallation() {
    log_message "Verifying uninstallation..."
    
    local verification_failed=false
    
    # Check if services are still running
    if systemctl is-active --quiet "$SERVICE_NAME" 2>/dev/null; then
        log_error "Camera service is still running"
        verification_failed=true
    else
        log_success "Camera service is not running"
    fi
    
    if systemctl is-active --quiet mediamtx 2>/dev/null; then
        log_error "MediaMTX service is still running"
        verification_failed=true
    else
        log_success "MediaMTX service is not running"
    fi
    
    # Check if installation directory still exists
    if [[ -d "$INSTALL_DIR" ]]; then
        log_error "Installation directory still exists: $INSTALL_DIR"
        verification_failed=true
    else
        log_success "Installation directory removed"
    fi
    
    # Check if MediaMTX directory still exists
    if [[ -d "$MEDIAMTX_DIR" ]]; then
        log_error "MediaMTX directory still exists: $MEDIAMTX_DIR"
        verification_failed=true
    else
        log_success "MediaMTX directory removed"
    fi
    
    # Check if service files still exist
    if [[ -f "/etc/systemd/system/$SERVICE_NAME.service" ]]; then
        log_error "Service file residue found: /etc/systemd/system/$SERVICE_NAME.service"
        verification_failed=true
    else
        log_success "Service file removed"
    fi
    
    if [[ -f "/etc/systemd/system/mediamtx.service" ]]; then
        log_error "MediaMTX service file residue found: /etc/systemd/system/mediamtx.service"
        verification_failed=true
    else
        log_success "MediaMTX service file removed"
    fi
    
    # Check if user still exists
    if getent passwd "$SERVICE_USER" >/dev/null 2>&1; then
        log_error "Service user still exists: $SERVICE_USER"
        verification_failed=true
    else
        log_success "Service user removed"
    fi
    
    # Check if group still exists
    if getent group "$SERVICE_GROUP" >/dev/null 2>&1; then
        log_error "Service group still exists: $SERVICE_GROUP"
        verification_failed=true
    else
        log_success "Service group removed"
    fi
    
    if [[ "$verification_failed" == "true" ]]; then
        log_error "Uninstallation verification failed"
        return 1
    else
        log_success "Uninstallation verification completed successfully"
    fi
}

# Function to generate uninstall report
generate_uninstall_report() {
    local report_file="/tmp/camera-service-uninstall-report-$(date +%Y%m%d_%H%M%S).txt"
    
    log_message "Generating uninstall report: $report_file"
    
    {
        echo "MediaMTX Camera Service (Go) Uninstall Report"
        echo "Generated: $(date)"
        echo "=============================================="
        echo ""
        echo "Services Status:"
        echo "  Camera Service: $(systemctl is-active camera-service 2>/dev/null || echo 'not found')"
        echo "  MediaMTX Service: $(systemctl is-active mediamtx 2>/dev/null || echo 'not found')"
        echo ""
        echo "Directories Status:"
        echo "  Installation Directory: $([ -d "$INSTALL_DIR" ] && echo 'exists' || echo 'removed')"
        echo "  MediaMTX Directory: $([ -d "$MEDIAMTX_DIR" ] && echo 'exists' || echo 'removed')"
        echo ""
        echo "Service Files Status:"
        echo "  Camera Service File: $([ -f "/etc/systemd/system/$SERVICE_NAME.service" ] && echo 'exists' || echo 'removed')"
        echo "  MediaMTX Service File: $([ -f "/etc/systemd/system/mediamtx.service" ] && echo 'exists' || echo 'removed')"
        echo ""
        echo "User/Group Status:"
        echo "  Service User: $(getent passwd "$SERVICE_USER" >/dev/null 2>&1 && echo 'exists' || echo 'removed')"
        echo "  Service Group: $(getent group "$SERVICE_GROUP" >/dev/null 2>&1 && echo 'exists' || echo 'removed')"
        echo ""
        echo "Uninstall completed successfully"
    } > "$report_file"
    
    log_success "Uninstall report generated: $report_file"
}

# Main uninstall function
main() {
    log_message "Starting MediaMTX Camera Service (Go) uninstallation..."
    
    # Check if running as root
    check_root
    
    # Stop and disable services
    stop_camera_service
    remove_mediamtx
    
    # Remove monitoring service
    remove_monitoring_service
    
    # Remove systemd files and symlinks
    remove_service_file
    remove_systemd_symlinks
    remove_environment_files
    
    # Remove installation directories
    remove_installation_directory
    
    # Remove SSL certificates and nginx configuration
    remove_ssl_certificates
    remove_nginx_configuration
    
    # Remove service user and group
    remove_service_user
    
    # Reload systemd
    reload_systemd
    
    # Verify uninstallation
    verify_uninstallation
    
    # Generate uninstall report
    generate_uninstall_report
    
    log_success "MediaMTX Camera Service (Go) uninstallation completed successfully!"
    log_message "All components have been removed from the system."
}

# Run main function
main "$@"
