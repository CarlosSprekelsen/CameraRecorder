#!/bin/bash

# MediaMTX Server Startup Script
# Starts the MediaMTX server for integration testing

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SERVER_DIR="/home/carlossprekelsen/CameraRecorder/mediamtx-camera-service-go"
INSTALL_DIR="/opt/camera-service"
MEDIAMTX_DIR="/opt/mediamtx"
SERVICE_NAME="camera-service"

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if service is running
check_service_running() {
    local service_name="$1"
    if systemctl is-active --quiet "$service_name" 2>/dev/null; then
        return 0
    else
        return 1
    fi
}

# Function to start MediaMTX server
start_mediamtx() {
    print_status "Starting MediaMTX server..."
    
    if check_service_running "mediamtx"; then
        print_success "MediaMTX server is already running"
    else
        print_status "Starting MediaMTX service..."
        sudo systemctl start mediamtx
        
        # Wait for service to start
        sleep 3
        
        if check_service_running "mediamtx"; then
            print_success "MediaMTX server started successfully"
        else
            print_error "Failed to start MediaMTX server"
            return 1
        fi
    fi
}

# Function to start Camera Service
start_camera_service() {
    print_status "Starting Camera Service..."
    
    if check_service_running "$SERVICE_NAME"; then
        print_success "Camera Service is already running"
    else
        print_status "Starting Camera Service..."
        sudo systemctl start "$SERVICE_NAME"
        
        # Wait for service to start
        sleep 3
        
        if check_service_running "$SERVICE_NAME"; then
            print_success "Camera Service started successfully"
        else
            print_error "Failed to start Camera Service"
            return 1
        fi
    fi
}

# Function to check server connectivity
check_server_connectivity() {
    print_status "Checking server connectivity..."
    
    # Check MediaMTX API
    if curl -s --max-time 5 http://localhost:9997/v3/paths/list >/dev/null 2>&1; then
        print_success "MediaMTX API is accessible at http://localhost:9997"
    else
        print_warning "MediaMTX API is not accessible at http://localhost:9997"
    fi
    
    # Check Camera Service (if it has a health endpoint)
    if curl -s --max-time 5 http://localhost:8080/health >/dev/null 2>&1; then
        print_success "Camera Service API is accessible at http://localhost:8080"
    else
        print_warning "Camera Service API may not be accessible at http://localhost:8080"
    fi
    
    # Check WebSocket port (8002)
    if nc -z localhost 8002 2>/dev/null; then
        print_success "WebSocket server is accessible at ws://localhost:8002/ws"
    else
        print_warning "WebSocket server is not accessible at ws://localhost:8002/ws"
    fi
}

# Function to install server if not present
install_server() {
    print_status "Checking if server is installed..."
    
    if [ ! -f "$INSTALL_DIR/$SERVICE_NAME" ] && [ ! -f "/etc/systemd/system/$SERVICE_NAME.service" ]; then
        print_warning "Server not installed. Installing MediaMTX Camera Service..."
        
        # Check if we have the installation script
        if [ -f "$SERVER_DIR/deployment/scripts/install.sh" ]; then
            print_status "Running server installation..."
            sudo "$SERVER_DIR/deployment/scripts/install.sh"
            print_success "Server installation completed"
        else
            print_error "Installation script not found at $SERVER_DIR/deployment/scripts/install.sh"
            return 1
        fi
    else
        print_success "Server is already installed"
    fi
}

# Function to build server from source
build_server() {
    print_status "Building server from source..."
    
    if [ -f "$SERVER_DIR/Makefile" ]; then
        cd "$SERVER_DIR"
        make build
        print_success "Server built successfully"
    else
        print_error "Makefile not found at $SERVER_DIR"
        return 1
    fi
}

# Main function
main() {
    echo "🚀 Starting MediaMTX Server for Integration Testing"
    echo "=================================================="
    
    # Check if running as root for service operations
    if [[ $EUID -eq 0 ]]; then
        print_warning "Running as root - this is not recommended for development"
    fi
    
    # Install server if needed
    install_server
    
    # Start MediaMTX server
    start_mediamtx
    
    # Start Camera Service
    start_camera_service
    
    # Check connectivity
    check_server_connectivity
    
    print_success "🎉 MediaMTX Server is ready for integration testing!"
    echo ""
    echo "📡 Server Endpoints:"
    echo "  - MediaMTX API: http://localhost:9997"
    echo "  - Camera Service: http://localhost:8080"
    echo "  - WebSocket: ws://localhost:8002/ws"
    echo ""
    echo "🔧 Service Management:"
    echo "  - Check status: sudo systemctl status mediamtx camera-service"
    echo "  - View logs: sudo journalctl -u mediamtx -u camera-service -f"
    echo "  - Stop services: sudo systemctl stop mediamtx camera-service"
    echo ""
    echo "✅ Ready to run integration tests!"
}

# Run main function
main "$@"
