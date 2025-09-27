#!/bin/bash

# Complete Reinstall with Token Generation
# Orchestrates: uninstall -> install -> generate API keys -> setup test environment
# Usage: ./reinstall-with-tokens.sh

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SERVER_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
CLIENT_DIR="$(cd "$SERVER_DIR/../MediaMTX-Camera-Service-Client/client" && pwd)"

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
check_root() {
    if [ "$EUID" -ne 0 ]; then
        log_error "This script must be run as root (use sudo)"
        exit 1
    fi
}

# Step 1: Uninstall existing service
uninstall_service() {
    log_message "Step 1: Uninstalling existing service..."
    
    if [ -f "$SCRIPT_DIR/uninstall.sh" ]; then
        cd "$SCRIPT_DIR"
        ./uninstall.sh
        log_success "Service uninstalled"
    else
        log_warning "Uninstall script not found, attempting manual cleanup..."
        
        # Stop service if running
        systemctl stop camera-service 2>/dev/null || true
        systemctl disable camera-service 2>/dev/null || true
        
        # Remove service files
        rm -f /etc/systemd/system/camera-service.service
        rm -rf /opt/camera-service 2>/dev/null || true
        
        log_success "Manual cleanup completed"
    fi
    
    # Reload systemd
    systemctl daemon-reload
}

# Step 2: Install service
install_service() {
    log_message "Step 2: Installing service..."
    
    if [ -f "$SCRIPT_DIR/install.sh" ]; then
        cd "$SCRIPT_DIR"
        ./install.sh
        log_success "Service installed"
    else
        log_error "Install script not found at $SCRIPT_DIR/install.sh"
        exit 1
    fi
}

# Step 3: Generate fresh API keys
generate_api_keys() {
    log_message "Step 3: Generating fresh API keys..."
    
    if [ -f "$SCRIPT_DIR/manage-api-keys.sh" ]; then
        cd "$SCRIPT_DIR"
        ./manage-api-keys.sh generate test
        log_success "API keys generated"
    else
        log_error "API key management script not found"
        exit 1
    fi
}

# Step 4: Install API keys to server
install_api_keys() {
    log_message "Step 4: Installing API keys to server..."
    
    cd "$SCRIPT_DIR"
    ./manage-api-keys.sh install test
    log_success "API keys installed to server"
}

# Step 5: Setup test environment
setup_test_environment() {
    log_message "Step 5: Setting up test environment..."
    
    # Copy environment file to client
    if [ -f "$SERVER_DIR/config/test/api-keys/test-keys.env" ]; then
        cp "$SERVER_DIR/config/test/api-keys/test-keys.env" "$CLIENT_DIR/.test_env"
        log_success "Client environment updated: $CLIENT_DIR/.test_env"
    else
        log_error "Environment file not found"
        exit 1
    fi
    
    # Copy keys to client fixtures
    if [ -f "$SERVER_DIR/config/test/api-keys/test-keys.json" ]; then
        mkdir -p "$CLIENT_DIR/tests/fixtures"
        cp "$SERVER_DIR/config/test/api-keys/test-keys.json" "$CLIENT_DIR/tests/fixtures/"
        log_success "API keys copied to client fixtures"
    else
        log_error "API keys file not found"
        exit 1
    fi
}

# Step 6: Start service
start_service() {
    log_message "Step 6: Starting service..."
    
    systemctl start camera-service
    systemctl enable camera-service
    
    # Wait for service to start
    sleep 5
    
    # Check service status
    if systemctl is-active --quiet camera-service; then
        log_success "Service started successfully"
    else
        log_error "Service failed to start"
        systemctl status camera-service
        exit 1
    fi
}

# Step 7: Verify installation
verify_installation() {
    log_message "Step 7: Verifying installation..."
    
    # Check health endpoint
    if curl -s http://localhost:8003/health >/dev/null; then
        log_success "Health endpoint responding"
    else
        log_warning "Health endpoint not responding"
    fi
    
    # Check WebSocket endpoint
    if nc -z localhost 8002; then
        log_success "WebSocket endpoint available"
    else
        log_warning "WebSocket endpoint not available"
    fi
    
    # Run verification script if available
    if [ -f "$SCRIPT_DIR/verify_installation.sh" ]; then
        cd "$SCRIPT_DIR"
        ./verify_installation.sh
    fi
    
    log_success "Installation verification completed"
}

# Main execution
main() {
    log_message "Starting complete reinstall with token generation..."
    log_message "This will:"
    log_message "  1. Uninstall existing service"
    log_message "  2. Install fresh service"
    log_message "  3. Generate fresh API keys"
    log_message "  4. Install API keys to server"
    log_message "  5. Setup test environment"
    log_message "  6. Start service"
    log_message "  7. Verify installation"
    echo
    
    check_root
    
    uninstall_service
    install_service
    generate_api_keys
    install_api_keys
    setup_test_environment
    start_service
    verify_installation
    
    echo
    log_success "🎉 Complete reinstall with fresh tokens completed!"
    log_message "📋 Next steps:"
    log_message "  - Test API: cd $CLIENT_DIR && source .test_env"
    log_message "  - Run tests: npm run test:integration"
    log_message "  - Check logs: journalctl -u camera-service -f"
    echo
    log_message "🔑 Token files generated:"
    log_message "  - API keys: $SERVER_DIR/config/test/api-keys/test-keys.json"
    log_message "  - Environment: $CLIENT_DIR/.test_env"
    log_message "  - Client fixtures: $CLIENT_DIR/tests/fixtures/test-keys.json"
}

# Show usage
show_usage() {
    echo "Complete Reinstall with Token Generation"
    echo ""
    echo "Usage: $0"
    echo ""
    echo "This script orchestrates:"
    echo "  1. Uninstall existing service"
    echo "  2. Install fresh service"
    echo "  3. Generate fresh API keys"
    echo "  4. Install API keys to server"
    echo "  5. Setup test environment"
    echo "  6. Start service"
    echo "  7. Verify installation"
    echo ""
    echo "Examples:"
    echo "  $0                    # Complete reinstall"
    echo "  sudo $0               # Run as root"
}

# Handle help
if [[ "$1" == "--help" || "$1" == "-h" ]]; then
    show_usage
    exit 0
fi

# Run main function
main "$@"
