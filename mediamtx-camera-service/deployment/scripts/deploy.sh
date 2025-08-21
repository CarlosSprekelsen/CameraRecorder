#!/bin/bash

# MediaMTX Camera Service Deployment Automation Script
# Automates the repetitive uninstall/install cycle for development and testing

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$(dirname "$SCRIPT_DIR")")"
UNINSTALL_SCRIPT="$SCRIPT_DIR/uninstall.sh"
INSTALL_SCRIPT="$SCRIPT_DIR/install.sh"
TEST_ENV_FILE="$PROJECT_ROOT/.test_env"

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

# Function to check if scripts exist
check_scripts() {
    if [[ ! -f "$UNINSTALL_SCRIPT" ]]; then
        log_error "Uninstall script not found: $UNINSTALL_SCRIPT"
        exit 1
    fi
    
    if [[ ! -f "$INSTALL_SCRIPT" ]]; then
        log_error "Install script not found: $INSTALL_SCRIPT"
        exit 1
    fi
    
    if [[ ! -f "$TEST_ENV_FILE" ]]; then
        log_error "Test environment file not found: $TEST_ENV_FILE"
        exit 1
    fi
}

# Function to get JWT secret from deployed service
get_deployed_jwt_secret() {
    local env_file="/opt/camera-service/.env"
    if [[ -f "$env_file" ]]; then
        grep "CAMERA_SERVICE_JWT_SECRET=" "$env_file" | cut -d'=' -f2
    else
        log_warning "Deployed service .env file not found, cannot get JWT secret"
        echo ""
    fi
}

# Function to update test environment with new JWT secret
update_test_env() {
    local new_secret="$1"
    if [[ -n "$new_secret" ]]; then
        log_message "Updating test environment with new JWT secret..."
        
        # Create backup of current .test_env
        cp "$TEST_ENV_FILE" "${TEST_ENV_FILE}.backup.$(date +%Y%m%d_%H%M%S)"
        
        # Update the JWT secret (escape special characters for sed)
        local escaped_secret=$(echo "$new_secret" | sed 's/[[\.*^$()+?{|]/\\&/g')
        sed -i "s/export CAMERA_SERVICE_JWT_SECRET=.*/export CAMERA_SERVICE_JWT_SECRET=$escaped_secret/" "$TEST_ENV_FILE"
        
        # Verify the update was successful
        local updated_secret=$(grep "CAMERA_SERVICE_JWT_SECRET=" "$TEST_ENV_FILE" | cut -d'=' -f2)
        if [[ "$updated_secret" == "$new_secret" ]]; then
            log_success "Server test environment updated with new JWT secret"
        else
            log_error "Failed to update JWT secret in server test environment"
            log_error "Expected: $new_secret"
            log_error "Found: $updated_secret"
            return 1
        fi
        
        # Also update client test environment
        local client_test_env="$PROJECT_ROOT/../MediaMTX-Camera-Service-Client/client/.test_env"
        if [[ -f "$client_test_env" ]]; then
            log_message "Updating client test environment with new JWT secret..."
            
            # Create backup of client .test_env
            cp "$client_test_env" "${client_test_env}.backup.$(date +%Y%m%d_%H%M%S)"
            
            # Update the client JWT secret
            sed -i "s/export CAMERA_SERVICE_JWT_SECRET=.*/export CAMERA_SERVICE_JWT_SECRET=$escaped_secret/" "$client_test_env"
            
            # Verify the client update was successful
            local client_updated_secret=$(grep "CAMERA_SERVICE_JWT_SECRET=" "$client_test_env" | cut -d'=' -f2)
            if [[ "$client_updated_secret" == "$new_secret" ]]; then
                log_success "Client test environment updated with new JWT secret"
            else
                log_error "Failed to update JWT secret in client test environment"
                log_error "Expected: $new_secret"
                log_error "Found: $client_updated_secret"
                return 1
            fi
        else
            log_warning "Client test environment file not found: $client_test_env"
        fi
    else
        log_warning "No JWT secret available, test environments not updated"
    fi
}

# Function to force uninstall (bypass user confirmation)
force_uninstall() {
    log_message "Performing forced uninstall..."
    
    # Create a temporary uninstall script that bypasses confirmation
    local temp_uninstall="/tmp/force_uninstall_$$.sh"
    
    # Copy the original uninstall script
    cp "$UNINSTALL_SCRIPT" "$temp_uninstall"
    
    # Replace the confirmation section with auto-confirm
    sed -i 's/read -p "Are you sure you want to continue? (y\/N): " -n 1 -r/REPLY=y/' "$temp_uninstall"
    
    # Make it executable and run it
    chmod +x "$temp_uninstall"
    "$temp_uninstall"
    
    # Clean up
    rm -f "$temp_uninstall"
    
    log_success "Forced uninstall completed"
}

# Function to install service
install_service() {
    log_message "Installing service..."
    "$INSTALL_SCRIPT"
    log_success "Service installation completed"
}

# Function to sync JWT secret
sync_jwt_secret() {
    log_message "Syncing JWT secret from deployed service..."
    local deployed_secret=$(get_deployed_jwt_secret)
    update_test_env "$deployed_secret"
}

# Function to validate deployment
validate_deployment() {
    log_message "Validating deployment..."
    
    # Check if services are running
    if systemctl is-active --quiet camera-service; then
        log_success "Camera service is running"
    else
        log_error "Camera service is not running"
        return 1
    fi
    
    if systemctl is-active --quiet mediamtx; then
        log_success "MediaMTX service is running"
    else
        log_error "MediaMTX service is not running"
        return 1
    fi
    
    # Check if health endpoint is responding
    if command_exists curl; then
        if curl -s http://localhost:8003/health/ready >/dev/null 2>&1; then
            log_success "Health endpoint is responding"
        else
            log_error "Health endpoint is not responding"
            return 1
        fi
    fi
    
    log_success "Deployment validation passed"
    return 0
}

# Function to show usage
show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  --force-uninstall    Force uninstall without user confirmation"
    echo "  --skip-uninstall     Skip uninstall step (useful for first-time install)"
    echo "  --skip-validation    Skip deployment validation"
    echo "  --help               Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                    # Full deployment with confirmation"
    echo "  $0 --force-uninstall  # Automated deployment without confirmation"
    echo "  $0 --skip-uninstall   # Install only (skip uninstall)"
}

# Main deployment function
main() {
    local force_uninstall=false
    local skip_uninstall=false
    local skip_validation=false
    
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --force-uninstall)
                force_uninstall=true
                shift
                ;;
            --skip-uninstall)
                skip_uninstall=true
                shift
                ;;
            --skip-validation)
                skip_validation=true
                shift
                ;;
            --help)
                show_usage
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                show_usage
                exit 1
                ;;
        esac
    done
    
    log_message "Starting MediaMTX Camera Service deployment automation..."
    log_message "================================================"
    
    # Check prerequisites
    check_root
    check_scripts
    
    # Store current JWT secret for comparison
    local old_secret=$(get_deployed_jwt_secret)
    
    # Uninstall step
    if [[ "$skip_uninstall" == false ]]; then
        if [[ "$force_uninstall" == true ]]; then
            force_uninstall
        else
            log_message "Running uninstall script (user confirmation required)..."
            "$UNINSTALL_SCRIPT"
        fi
    else
        log_message "Skipping uninstall step"
    fi
    
    # Install step
    install_service
    
    # Sync JWT secret
    sync_jwt_secret
    
    # Validation step
    if [[ "$skip_validation" == false ]]; then
        if validate_deployment; then
            log_success "Deployment completed successfully!"
        else
            log_error "Deployment validation failed"
            exit 1
        fi
    else
        log_message "Skipping deployment validation"
    fi
    
    log_message "================================================"
    log_success "Deployment automation completed!"
    
    # Show summary
    local new_secret=$(get_deployed_jwt_secret)
    if [[ "$old_secret" != "$new_secret" ]]; then
        log_message "JWT secret was updated during deployment"
    fi
    
    log_message "Next steps:"
    log_message "- Run tests to verify functionality"
    log_message "- Check service logs if needed: journalctl -u camera-service -f"
}

# Run main function
main "$@"
