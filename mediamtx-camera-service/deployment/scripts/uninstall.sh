#!/bin/bash

# MediaMTX Camera Service Uninstall Script
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
    
    if [[ -d "$MEDIAMTX_DIR" ]]; then
        # Stop MediaMTX service if it exists
        if systemctl list-unit-files | grep -q "mediamtx"; then
            if systemctl is-active --quiet mediamtx; then
                systemctl stop mediamtx
                log_success "MediaMTX service stopped"
            fi
            systemctl disable mediamtx
            log_success "MediaMTX service disabled"
        fi
        
        # Remove MediaMTX service file
        local mediamtx_service="/etc/systemd/system/mediamtx.service"
        if [[ -f "$mediamtx_service" ]]; then
            rm -f "$mediamtx_service"
            log_success "MediaMTX service file removed"
        fi
        
        # Remove MediaMTX installation directory
        rm -rf "$MEDIAMTX_DIR"
        log_success "MediaMTX installation directory removed: $MEDIAMTX_DIR"
    else
        log_message "MediaMTX installation directory not found: $MEDIAMTX_DIR"
    fi
}

# Function to remove data directories
remove_data_directories() {
    log_message "Removing data directories..."
    
    local data_dirs=("/var/recordings" "/var/snapshots" "/var/log/camera-service")
    
    for dir in "${data_dirs[@]}"; do
        if [[ -d "$dir" ]]; then
            rm -rf "$dir"
            log_success "Data directory removed: $dir"
        else
            log_message "Data directory not found: $dir"
        fi
    done
}

# Function to remove service user (optional)
remove_service_user() {
    log_message "Checking service user..."
    
    if id "$SERVICE_USER" &>/dev/null; then
        log_warning "Service user $SERVICE_USER exists"
        log_warning "Consider whether to remove the user:"
        log_warning "  - If this is a test environment: userdel -r $SERVICE_USER"
        log_warning "  - If this is production: Keep the user for security"
        log_message "Service user $SERVICE_USER preserved (manual removal required if needed)"
    else
        log_message "Service user $SERVICE_USER not found"
    fi
}

# Function to remove configuration files
remove_configuration_files() {
    log_message "Removing configuration files..."
    
    local config_files=(
        "/etc/mediamtx/mediamtx.yml"
        "/etc/mediamtx/mediamtx.yml.backup"
    )
    
    for config_file in "${config_files[@]}"; do
        if [[ -f "$config_file" ]]; then
            rm -f "$config_file"
            log_success "Configuration file removed: $config_file"
        else
            log_message "Configuration file not found: $config_file"
        fi
    done
    
    # Remove MediaMTX config directory if empty
    if [[ -d "/etc/mediamtx" ]]; then
        if [[ -z "$(ls -A /etc/mediamtx)" ]]; then
            rmdir "/etc/mediamtx"
            log_success "Empty MediaMTX config directory removed"
        else
            log_warning "MediaMTX config directory not empty, preserving: /etc/mediamtx"
        fi
    fi
}

# Function to check for port residues
check_port_residues() {
    log_message "Checking for port residues..."
    
    local ports=(8002 8003 8554 8888 8889 9997)
    local residues_found=0
    
    for port in "${ports[@]}"; do
        if command_exists "netstat"; then
            if netstat -tlnp 2>/dev/null | grep -q ":$port "; then
                log_warning "Port $port is still in use"
                ((residues_found++))
            fi
        elif command_exists "ss"; then
            if ss -tlnp 2>/dev/null | grep -q ":$port "; then
                log_warning "Port $port is still in use"
                ((residues_found++))
            fi
        fi
    done
    
    if [[ $residues_found -eq 0 ]]; then
        log_success "No port residues found"
    else
        log_warning "Found $residues_found port residues (may be from other services)"
    fi
}

# Function to reload systemd
reload_systemd() {
    log_message "Reloading systemd..."
    
    systemctl daemon-reload
    log_success "Systemd daemon reloaded"
}

# Function to validate uninstallation
validate_uninstallation() {
    log_message "Validating uninstallation..."
    
    local residues_found=0
    
    # Check for service residues
    if systemctl list-unit-files | grep -q "$SERVICE_NAME"; then
        log_error "Service residue found: $SERVICE_NAME"
        ((residues_found++))
    fi
    
    if [[ -f "/etc/systemd/system/$SERVICE_NAME.service" ]]; then
        log_error "Service file residue found: /etc/systemd/system/$SERVICE_NAME.service"
        ((residues_found++))
    fi
    
    # Check for directory residues
    if [[ -d "$INSTALL_DIR" ]]; then
        log_error "Installation directory residue found: $INSTALL_DIR"
        ((residues_found++))
    fi
    
    if [[ -d "/var/recordings" ]]; then
        log_error "Recordings directory residue found: /var/recordings"
        ((residues_found++))
    fi
    
    if [[ -d "/var/snapshots" ]]; then
        log_error "Snapshots directory residue found: /var/snapshots"
        ((residues_found++))
    fi
    
    # Check for MediaMTX residues
    if [[ -d "$MEDIAMTX_DIR" ]]; then
        log_error "MediaMTX directory residue found: $MEDIAMTX_DIR"
        ((residues_found++))
    fi
    
    if systemctl list-unit-files | grep -q "mediamtx"; then
        log_error "MediaMTX service residue found"
        ((residues_found++))
    fi
    
    if [[ $residues_found -eq 0 ]]; then
        log_success "Uninstallation validation passed - no critical residues found"
        return 0
    else
        log_error "Uninstallation validation failed - found $residues_found residues"
        return 1
    fi
}

# Function to generate uninstall report
generate_uninstall_report() {
    log_message "Generating uninstall report..."
    
    local report_file="uninstall_report_$(date +%Y%m%d_%H%M%S).txt"
    
    {
        echo "MediaMTX Camera Service Uninstall Report"
        echo "========================================"
        echo "Date: $(date)"
        echo "System: $(uname -a)"
        echo ""
        echo "Uninstallation Summary:"
        echo "----------------------"
        echo "Service stopped and disabled: $SERVICE_NAME"
        echo "Service file removed: /etc/systemd/system/$SERVICE_NAME.service"
        echo "Installation directory removed: $INSTALL_DIR"
        echo "MediaMTX directory removed: $MEDIAMTX_DIR"
        echo "Data directories removed: /var/recordings, /var/snapshots"
        echo ""
        echo "Validation Results:"
        echo "------------------"
        
        # Check for residues
        local residues=0
        if systemctl list-unit-files | grep -q "$SERVICE_NAME"; then
            echo "❌ Service residue found: $SERVICE_NAME"
            ((residues++))
        else
            echo "✅ No service residue"
        fi
        
        if [[ -f "/etc/systemd/system/$SERVICE_NAME.service" ]]; then
            echo "❌ Service file residue found"
            ((residues++))
        else
            echo "✅ No service file residue"
        fi
        
        if [[ -d "$INSTALL_DIR" ]]; then
            echo "❌ Installation directory residue found"
            ((residues++))
        else
            echo "✅ No installation directory residue"
        fi
        
        if [[ -d "/var/recordings" ]]; then
            echo "❌ Recordings directory residue found"
            ((residues++))
        else
            echo "✅ No recordings directory residue"
        fi
        
        if [[ -d "/var/snapshots" ]]; then
            echo "❌ Snapshots directory residue found"
            ((residues++))
        else
            echo "✅ No snapshots directory residue"
        fi
        
        echo ""
        echo "Residues found: $residues"
        if [[ $residues -eq 0 ]]; then
            echo "Status: ✅ UNINSTALLATION COMPLETE"
        else
            echo "Status: ⚠️ UNINSTALLATION INCOMPLETE"
        fi
        
    } > "$report_file"
    
    log_success "Uninstall report generated: $report_file"
    echo "$report_file"
}

# Main uninstall function
main() {
    log_message "Starting MediaMTX Camera Service uninstallation..."
    log_message "================================================"
    
    # Check if running as root
    check_root
    
    # Confirm uninstallation
    echo -e "${YELLOW}WARNING: This will completely remove the MediaMTX Camera Service installation.${NC}"
    echo -e "${YELLOW}This includes:${NC}"
    echo -e "${YELLOW}- Camera service and all its data${NC}"
    echo -e "${YELLOW}- MediaMTX server and configuration${NC}"
    echo -e "${YELLOW}- All recordings and snapshots${NC}"
    echo -e "${YELLOW}- Service configuration files${NC}"
    echo ""
    read -p "Are you sure you want to continue? (y/N): " -n 1 -r
    echo
    
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_message "Uninstallation cancelled by user"
        exit 0
    fi
    
    # Perform uninstallation steps
    stop_camera_service
    remove_service_file
    remove_installation_directory
    remove_mediamtx
    remove_data_directories
    remove_service_user
    remove_configuration_files
    check_port_residues
    reload_systemd
    
    # Validate uninstallation
    if validate_uninstallation; then
        log_success "Uninstallation completed successfully!"
    else
        log_warning "Uninstallation completed with residues - manual cleanup may be required"
    fi
    
    # Generate report
    local report_file=$(generate_uninstall_report)
    
    log_message "================================================"
    log_success "Uninstallation completed!"
    log_message "Report generated: $report_file"
    log_message ""
    log_message "Next steps:"
    log_message "- Review the uninstall report"
    log_message "- Verify no critical residues remain"
    log_message "- If needed, manually remove service user: userdel -r $SERVICE_USER"
    log_message "- If needed, manually remove any remaining configuration files"
}

# Run main function
main "$@" 