#!/bin/bash

# Pre-Deployment Validation Script
# Validates production deployment readiness before deployment
# Catches issues that were missed in the WebSocket binding incident

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

# Function to validate directory permissions
validate_directory_permissions() {
    log_message "Validating directory permissions..."
    
    local dirs=("/var/recordings" "/var/snapshots" "/var/log/camera-service")
    local errors=0
    
    for dir in "${dirs[@]}"; do
        if [[ ! -d "$dir" ]]; then
            log_error "Directory $dir does not exist"
            ((errors++))
            continue
        fi
        
        # Check permissions
        local perms=$(stat -c "%a" "$dir")
        if [[ "$perms" != "755" ]]; then
            log_error "Directory $dir has incorrect permissions: $perms (expected: 755)"
            ((errors++))
        fi
        
        # Check ownership
        local owner=$(stat -c "%U" "$dir")
        if [[ "$owner" != "$SERVICE_USER" ]]; then
            log_error "Directory $dir not owned by $SERVICE_USER (owner: $owner)"
            ((errors++))
        fi
        
        log_success "Directory $dir permissions validated"
    done
    
    if [[ $errors -eq 0 ]]; then
        log_success "All directory permissions validated successfully"
    else
        log_error "Directory permission validation failed with $errors errors"
        return 1
    fi
}

# Function to validate service user
validate_service_user() {
    log_message "Validating service user..."
    
    if ! id "$SERVICE_USER" &>/dev/null; then
        log_error "Service user $SERVICE_USER does not exist"
        return 1
    fi
    
    # Check if user can access required directories
    local dirs=("/var/recordings" "/var/snapshots")
    for dir in "${dirs[@]}"; do
        if [[ -d "$dir" ]]; then
            if ! sudo -u "$SERVICE_USER" test -r "$dir" 2>/dev/null; then
                log_error "Service user $SERVICE_USER cannot read directory $dir"
                return 1
            fi
            if ! sudo -u "$SERVICE_USER" test -w "$dir" 2>/dev/null; then
                log_error "Service user $SERVICE_USER cannot write to directory $dir"
                return 1
            fi
        fi
    done
    
    log_success "Service user validation passed"
}

# Function to validate configuration file
validate_configuration() {
    log_message "Validating configuration file..."
    
    local config_file="$INSTALL_DIR/config/camera-service.yaml"
    
    if [[ ! -f "$config_file" ]]; then
        log_error "Configuration file $config_file does not exist"
        return 1
    fi
    
    # Check if configuration can be loaded
    if ! python3 -c "
import yaml
import sys
try:
    with open('$config_file', 'r') as f:
        config = yaml.safe_load(f)
    print('Configuration loaded successfully')
except Exception as e:
    print(f'Configuration loading failed: {e}', file=sys.stderr)
    sys.exit(1)
" 2>/dev/null; then
        log_error "Configuration file $config_file cannot be loaded"
        return 1
    fi
    
    # Validate specific configuration parameters
    local errors=0
    
    # Check for correct MediaMTX configuration
    if ! grep -q "api_port: 9997" "$config_file"; then
        log_error "Configuration missing correct api_port setting"
        ((errors++))
    fi
    
    # Check for correct logging configuration
    if ! grep -q "file_path:" "$config_file"; then
        log_error "Configuration missing file_path setting"
        ((errors++))
    fi
    
    if [[ $errors -eq 0 ]]; then
        log_success "Configuration file validation passed"
    else
        log_error "Configuration validation failed with $errors errors"
        return 1
    fi
}

# Function to validate Python imports
validate_python_imports() {
    log_message "Validating Python imports..."
    
    local import_tests=(
        "from src.camera_service.main import main"
        "from src.camera_service.config import Config"
        "from src.camera_service.service_manager import ServiceManager"
        "from src.mediamtx_wrapper.controller import MediaMTXController"
        "from src.websocket_server.server import WebSocketJsonRpcServer"
    )
    
    local errors=0
    
    for import_test in "${import_tests[@]}"; do
        if ! python3 -c "$import_test" 2>/dev/null; then
            log_error "Import failed: $import_test"
            ((errors++))
        fi
    done
    
    if [[ $errors -eq 0 ]]; then
        log_success "Python imports validation passed"
    else
        log_error "Python imports validation failed with $errors errors"
        return 1
    fi
}

# Function to validate service configuration
validate_service_configuration() {
    log_message "Validating service configuration..."
    
    local service_file="/etc/systemd/system/camera-service.service"
    
    if [[ ! -f "$service_file" ]]; then
        log_error "Service file $service_file does not exist"
        return 1
    fi
    
    # Check for correct ExecStart command
    if ! grep -q "src.camera_service.main" "$service_file"; then
        log_error "Service configuration has incorrect ExecStart command"
        return 1
    fi
    
    # Check for correct user
    if ! grep -q "User=$SERVICE_USER" "$service_file"; then
        log_error "Service configuration has incorrect user"
        return 1
    fi
    
    log_success "Service configuration validation passed"
}

# Function to validate service startup
validate_service_startup() {
    log_message "Validating service startup..."
    
    # Check if service is currently running
    if systemctl is-active --quiet "$SERVICE_NAME"; then
        log_success "Service is currently running"
    else
        log_warning "Service is not currently running, attempting to start..."
        
        # Try to start the service
        if systemctl start "$SERVICE_NAME"; then
            sleep 5  # Wait for service to start
            if systemctl is-active --quiet "$SERVICE_NAME"; then
                log_success "Service started successfully"
            else
                log_error "Service failed to start"
                return 1
            fi
        else
            log_error "Failed to start service"
            return 1
        fi
    fi
}

# Function to validate WebSocket binding
validate_websocket_binding() {
    log_message "Validating WebSocket binding..."
    
    # Check if port 8002 is listening
    if command_exists "netstat"; then
        if netstat -tlnp 2>/dev/null | grep -q ":8002"; then
            log_success "WebSocket server is binding to port 8002"
        else
            log_error "WebSocket server is not binding to port 8002"
            return 1
        fi
    elif command_exists "ss"; then
        if ss -tlnp 2>/dev/null | grep -q ":8002"; then
            log_success "WebSocket server is binding to port 8002"
        else
            log_error "WebSocket server is not binding to port 8002"
            return 1
        fi
    else
        log_warning "Cannot check port binding (netstat/ss not available)"
    fi
}

# Function to validate health endpoint
validate_health_endpoint() {
    log_message "Validating health endpoint..."
    
    if command_exists "curl"; then
        if curl -f -s http://localhost:8003/health/ready >/dev/null; then
            log_success "Health endpoint is responding"
        else
            log_error "Health endpoint is not responding"
            return 1
        fi
    else
        log_warning "Cannot check health endpoint (curl not available)"
    fi
}

# Function to validate MediaMTX integration
validate_mediamtx_integration() {
    log_message "Validating MediaMTX integration..."
    
    if command_exists "curl"; then
        if curl -f -s http://localhost:9997/v3/paths/list >/dev/null; then
            log_success "MediaMTX API is responding"
        else
            log_error "MediaMTX API is not responding"
            return 1
        fi
    else
        log_warning "Cannot check MediaMTX integration (curl not available)"
    fi
}

# Function to validate service logs
validate_service_logs() {
    log_message "Validating service logs..."
    
    # Check recent logs for errors
    local recent_logs=$(journalctl -u "$SERVICE_NAME" -n 10 --no-pager 2>/dev/null)
    
    if [[ -z "$recent_logs" ]]; then
        log_warning "No recent service logs found"
        return 0
    fi
    
    # Check for critical errors
    local error_count=$(echo "$recent_logs" | grep -i "error\|exception\|traceback\|fatal" | wc -l)
    
    if [[ $error_count -gt 0 ]]; then
        log_warning "Found $error_count potential errors in recent logs"
        echo "$recent_logs" | grep -i "error\|exception\|traceback\|fatal"
    else
        log_success "No critical errors found in recent logs"
    fi
}

# Function to validate dependencies
validate_dependencies() {
    log_message "Validating dependencies..."
    
    local dependencies=("python3" "systemctl" "journalctl")
    local missing_deps=()
    
    for dep in "${dependencies[@]}"; do
        if ! command_exists "$dep"; then
            missing_deps+=("$dep")
        fi
    done
    
    if [[ ${#missing_deps[@]} -eq 0 ]]; then
        log_success "All dependencies are available"
    else
        log_error "Missing dependencies: ${missing_deps[*]}"
        return 1
    fi
}

# Main validation function
main() {
    log_message "Starting pre-deployment validation..."
    log_message "================================================"
    
    local total_tests=0
    local passed_tests=0
    local failed_tests=0
    
    # Array of validation functions
    local validations=(
        "validate_dependencies"
        "validate_directory_permissions"
        "validate_service_user"
        "validate_configuration"
        "validate_python_imports"
        "validate_service_configuration"
        "validate_service_startup"
        "validate_websocket_binding"
        "validate_health_endpoint"
        "validate_mediamtx_integration"
        "validate_service_logs"
    )
    
    # Run all validations
    for validation in "${validations[@]}"; do
        ((total_tests++))
        log_message "Running: $validation"
        
        if $validation; then
            ((passed_tests++))
        else
            ((failed_tests++))
        fi
        
        echo
    done
    
    # Summary
    log_message "================================================"
    log_message "Validation Summary:"
    log_message "Total tests: $total_tests"
    log_message "Passed: $passed_tests"
    log_message "Failed: $failed_tests"
    
    if [[ $failed_tests -eq 0 ]]; then
        log_success "All validations passed! Deployment is ready."
        exit 0
    else
        log_error "Validation failed with $failed_tests errors. Please fix issues before deployment."
        exit 1
    fi
}

# Run main function
main "$@" 