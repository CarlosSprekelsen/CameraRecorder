#!/bin/bash

# MediaMTX Camera Service (Go) Installation Verification Script
# Verifies that all components are properly installed and functioning

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

# Function to check systemd service status
check_service_status() {
    local service_name="$1"
    local service_display_name="$2"
    
    log_message "Checking $service_display_name service status..."
    
    # Check if service exists
    if ! systemctl list-unit-files | grep -q "$service_name"; then
        log_error "$service_display_name service not found in systemd"
        return 1
    fi
    
    # Check if service is enabled
    if systemctl is-enabled --quiet "$service_name" 2>/dev/null; then
        log_success "$service_display_name service is enabled"
    else
        log_error "$service_display_name service is not enabled"
        return 1
    fi
    
    # Check if service is running
    if systemctl is-active --quiet "$service_name" 2>/dev/null; then
        log_success "$service_display_name service is running"
    else
        log_error "$service_display_name service is not running"
        return 1
    fi
    
    # Check service health
    if systemctl is-failed --quiet "$service_name" 2>/dev/null; then
        log_error "$service_display_name service has failed"
        return 1
    else
        log_success "$service_display_name service is healthy"
    fi
}

# Function to check file and directory existence
check_file_exists() {
    local file_path="$1"
    local description="$2"
    
    if [[ -f "$file_path" ]]; then
        log_success "$description exists: $file_path"
    else
        log_error "$description not found: $file_path"
        return 1
    fi
}

check_directory_exists() {
    local dir_path="$1"
    local description="$2"
    
    if [[ -d "$dir_path" ]]; then
        log_success "$description exists: $dir_path"
    else
        log_error "$description not found: $dir_path"
        return 1
    fi
}

# Function to check user and group existence
check_user_exists() {
    local user_name="$1"
    local description="$2"
    
    if getent passwd "$user_name" >/dev/null 2>&1; then
        log_success "$description exists: $user_name"
    else
        log_error "$description not found: $user_name"
        return 1
    fi
}

check_group_exists() {
    local group_name="$1"
    local description="$2"
    
    if getent group "$group_name" >/dev/null 2>&1; then
        log_success "$description exists: $group_name"
    else
        log_error "$description not found: $group_name"
        return 1
    fi
}

# Function to check file permissions
check_file_permissions() {
    local file_path="$1"
    local expected_owner="$2"
    local expected_group="$3"
    local description="$4"
    
    if [[ ! -f "$file_path" ]]; then
        log_error "$description not found: $file_path"
        return 1
    fi
    
    local actual_owner=$(stat -c '%U' "$file_path")
    local actual_group=$(stat -c '%G' "$file_path")
    
    if [[ "$actual_owner" == "$expected_owner" && "$actual_group" == "$expected_group" ]]; then
        log_success "$description has correct ownership: $actual_owner:$actual_group"
    else
        log_error "$description has incorrect ownership: $actual_owner:$actual_group (expected: $expected_owner:$expected_group)"
        return 1
    fi
}

# Function to check directory permissions
check_directory_permissions() {
    local dir_path="$1"
    local expected_owner="$2"
    local expected_group="$3"
    local description="$4"
    
    if [[ ! -d "$dir_path" ]]; then
        log_error "$description not found: $dir_path"
        return 1
    fi
    
    local actual_owner=$(stat -c '%U' "$dir_path")
    local actual_group=$(stat -c '%G' "$dir_path")
    
    if [[ "$actual_owner" == "$expected_owner" && "$actual_group" == "$expected_group" ]]; then
        log_success "$description has correct ownership: $actual_owner:$actual_group"
    else
        log_error "$description has incorrect ownership: $actual_owner:$actual_group (expected: $expected_owner:$expected_group)"
        return 1
    fi
}

# Function to check MediaMTX API accessibility
check_mediamtx_api() {
    log_message "Checking MediaMTX API accessibility..."
    
    if command_exists curl; then
        # Test MediaMTX API endpoint
        if curl -s --max-time 10 http://localhost:9997/v3/paths/list >/dev/null 2>&1; then
            log_success "MediaMTX API is accessible at http://localhost:9997"
        else
            log_error "MediaMTX API is not accessible at http://localhost:9997"
            return 1
        fi
        
        # Test specific API endpoint
        if curl -s --max-time 10 http://localhost:9997/v3/paths/list | grep -q "pageCount" 2>/dev/null; then
            log_success "MediaMTX API is responding correctly"
        else
            log_warning "MediaMTX API response format may be unexpected"
        fi
    else
        log_warning "curl not available, skipping API accessibility test"
    fi
}

# Function to check camera service API accessibility
check_camera_service_api() {
    log_message "Checking Camera Service API accessibility..."
    
    if command_exists curl; then
        # Test camera service endpoint (assuming it has a health check endpoint)
        if curl -s --max-time 10 http://localhost:8080/health >/dev/null 2>&1; then
            log_success "Camera Service API is accessible at http://localhost:8080"
        else
            log_warning "Camera Service API may not be accessible at http://localhost:8080"
            log_message "This is expected if the service doesn't expose a health endpoint"
        fi
    else
        log_warning "curl not available, skipping API accessibility test"
    fi
}

# Function to check video device permissions
check_video_permissions() {
    log_message "Checking video device permissions..."
    
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
    
    # Check if service users are in video group
    if groups "$SERVICE_USER" 2>/dev/null | grep -q video; then
        log_success "$SERVICE_USER is in video group"
    else
        log_error "$SERVICE_USER is not in video group"
        return 1
    fi
    
    if groups mediamtx 2>/dev/null | grep -q video; then
        log_success "mediamtx user is in video group"
    else
        log_error "mediamtx user is not in video group"
        return 1
    fi
    
    # Check video device permissions
    local video_device="/dev/video0"
    if [[ -e "$video_device" ]]; then
        local perms=$(ls -l "$video_device" | awk '{print $1}')
        if [[ "$perms" == "crw-rw----+" ]]; then
            log_success "Video device permissions are correct: $perms"
        else
            log_warning "Video device permissions may need adjustment: $perms"
            log_message "Expected: crw-rw----+, Found: $perms"
        fi
    else
        log_warning "No video device found at $video_device"
    fi
    
    log_success "Video device permissions check completed"
}

# Function to check Go installation
check_go_installation() {
    log_message "Checking Go installation..."
    
    if command_exists go; then
        local go_version=$(go version 2>/dev/null | head -n1)
        log_success "Go is installed: $go_version"
        
        # Check if Go is in PATH
        if which go >/dev/null 2>&1; then
            log_success "Go is available in PATH"
        else
            log_error "Go is not available in PATH"
            return 1
        fi
    else
        log_error "Go is not installed"
        return 1
    fi
}

# Function to check system dependencies
check_system_dependencies() {
    log_message "Checking system dependencies..."
    
    local dependencies=(
        "curl"
        "wget"
        "git"
        "ffmpeg"
        "v4l2-ctl"
    )
    
    local missing_deps=()
    
    for dep in "${dependencies[@]}"; do
        if command_exists "$dep"; then
            log_success "$dep is installed"
        else
            log_error "$dep is not installed"
            missing_deps+=("$dep")
        fi
    done
    
    if [[ ${#missing_deps[@]} -gt 0 ]]; then
        log_warning "Missing dependencies: ${missing_deps[*]}"
        return 1
    fi
    
    log_success "All system dependencies are installed"
}

# Function to check configuration files
check_configuration_files() {
    log_message "Checking configuration files..."
    
    # Check MediaMTX configuration
    check_file_exists "$MEDIAMTX_DIR/config/mediamtx.yml" "MediaMTX configuration file"
    
    # Check camera service configuration
    check_file_exists "$INSTALL_DIR/config/default.yaml" "Camera service configuration file"
    
    # Check if configuration files are readable by service users
    if sudo -u "$SERVICE_USER" test -r "$INSTALL_DIR/config/default.yaml" 2>/dev/null; then
        log_success "Camera service configuration is readable by $SERVICE_USER"
    else
        log_error "Camera service configuration is not readable by $SERVICE_USER"
        return 1
    fi
    
    if sudo -u mediamtx test -r "$MEDIAMTX_DIR/config/mediamtx.yml" 2>/dev/null; then
        log_success "MediaMTX configuration is readable by mediamtx user"
    else
        log_error "MediaMTX configuration is not readable by mediamtx user"
        return 1
    fi
}

# Function to check log files
check_log_files() {
    log_message "Checking log files..."
    
    # Check if journalctl can access service logs
    if journalctl -u "$SERVICE_NAME" --no-pager -n 1 >/dev/null 2>&1; then
        log_success "Camera service logs are accessible via journalctl"
    else
        log_warning "Camera service logs may not be accessible via journalctl"
    fi
    
    if journalctl -u mediamtx --no-pager -n 1 >/dev/null 2>&1; then
        log_success "MediaMTX logs are accessible via journalctl"
    else
        log_warning "MediaMTX logs may not be accessible via journalctl"
    fi
}

# Function to check service dependencies
check_service_dependencies() {
    log_message "Checking service dependencies..."
    
    # Check if MediaMTX service depends on network
    if systemctl show mediamtx | grep -q "After=network.target"; then
        log_success "MediaMTX service has correct network dependency"
    else
        log_warning "MediaMTX service may not have correct network dependency"
    fi
    
    # Check if camera service depends on MediaMTX
    if systemctl show "$SERVICE_NAME" | grep -q "After=mediamtx.service"; then
        log_success "Camera service has correct MediaMTX dependency"
    else
        log_warning "Camera service may not have correct MediaMTX dependency"
    fi
}

# Function to generate verification report
generate_verification_report() {
    local report_file="/tmp/camera-service-verification-report-$(date +%Y%m%d_%H%M%S).txt"
    
    log_message "Generating verification report: $report_file"
    
    {
        echo "MediaMTX Camera Service (Go) Verification Report"
        echo "Generated: $(date)"
        echo "================================================"
        echo ""
        echo "Services Status:"
        echo "  Camera Service: $(systemctl is-active camera-service 2>/dev/null || echo 'not found')"
        echo "  MediaMTX Service: $(systemctl is-active mediamtx 2>/dev/null || echo 'not found')"
        echo ""
        echo "Directories Status:"
        echo "  Installation Directory: $([ -d "$INSTALL_DIR" ] && echo 'exists' || echo 'missing')"
        echo "  MediaMTX Directory: $([ -d "$MEDIAMTX_DIR" ] && echo 'exists' || echo 'missing')"
        echo ""
        echo "Service Files Status:"
        echo "  Camera Service File: $([ -f "/etc/systemd/system/$SERVICE_NAME.service" ] && echo 'exists' || echo 'missing')"
        echo "  MediaMTX Service File: $([ -f "/etc/systemd/system/mediamtx.service" ] && echo 'exists' || echo 'missing')"
        echo ""
        echo "User/Group Status:"
        echo "  Service User: $(getent passwd "$SERVICE_USER" >/dev/null 2>&1 && echo 'exists' || echo 'missing')"
        echo "  Service Group: $(getent group "$SERVICE_GROUP" >/dev/null 2>&1 && echo 'exists' || echo 'missing')"
        echo ""
        echo "Binary Status:"
        echo "  Camera Service Binary: $([ -f "$INSTALL_DIR/$BINARY_NAME" ] && echo 'exists' || echo 'missing')"
        echo ""
        echo "Configuration Status:"
        echo "  Camera Service Config: $([ -f "$INSTALL_DIR/config/default.yaml" ] && echo 'exists' || echo 'missing')"
        echo "  MediaMTX Config: $([ -f "$MEDIAMTX_DIR/config/mediamtx.yml" ] && echo 'exists' || echo 'missing')"
        echo ""
        echo "API Status:"
        echo "  MediaMTX API: $(curl -s --max-time 5 http://localhost:9997/v3/paths/list >/dev/null 2>&1 && echo 'accessible' || echo 'not accessible')"
        echo "  Camera Service API: $(curl -s --max-time 5 http://localhost:8080/health >/dev/null 2>&1 && echo 'accessible' || echo 'not accessible')"
        echo ""
        echo "Verification completed"
    } > "$report_file"
    
    log_success "Verification report generated: $report_file"
}

# Main verification function
main() {
    log_message "Starting MediaMTX Camera Service (Go) installation verification..."
    
    local verification_failed=false
    
    # Check system dependencies
    if ! check_system_dependencies; then
        verification_failed=true
    fi
    
    # Check Go installation
    if ! check_go_installation; then
        verification_failed=true
    fi
    
    # Check service status
    if ! check_service_status "mediamtx" "MediaMTX"; then
        verification_failed=true
    fi
    
    if ! check_service_status "$SERVICE_NAME" "Camera Service"; then
        verification_failed=true
    fi
    
    # Check file and directory existence
    check_directory_exists "$INSTALL_DIR" "Installation directory"
    check_directory_exists "$MEDIAMTX_DIR" "MediaMTX directory"
    check_file_exists "$INSTALL_DIR/$BINARY_NAME" "Camera service binary"
    check_file_exists "$INSTALL_DIR/config/default.yaml" "Camera service configuration"
    check_file_exists "$MEDIAMTX_DIR/config/mediamtx.yml" "MediaMTX configuration"
    
    # Check user and group existence
    check_user_exists "$SERVICE_USER" "Service user"
    check_group_exists "$SERVICE_GROUP" "Service group"
    check_user_exists "mediamtx" "MediaMTX user"
    check_group_exists "mediamtx" "MediaMTX group"
    
    # Check file permissions
    check_file_permissions "$INSTALL_DIR/$BINARY_NAME" "$SERVICE_USER" "$SERVICE_GROUP" "Camera service binary"
    check_directory_permissions "$INSTALL_DIR/recordings" "$SERVICE_USER" "$SERVICE_GROUP" "Recordings directory"
    check_directory_permissions "$INSTALL_DIR/snapshots" "$SERVICE_USER" "$SERVICE_GROUP" "Snapshots directory"
    
    # Check configuration files
    if ! check_configuration_files; then
        verification_failed=true
    fi
    
    # Check service dependencies
    check_service_dependencies
    
    # Check video permissions
    check_video_permissions
    
    # Check API accessibility
    if ! check_mediamtx_api; then
        verification_failed=true
    fi
    
    check_camera_service_api
    
    # Check log files
    check_log_files
    
    # Generate verification report
    generate_verification_report
    
    if [[ "$verification_failed" == "true" ]]; then
        log_error "Installation verification failed"
        log_message "Please check the verification report for details"
        exit 1
    else
        log_success "Installation verification completed successfully!"
        log_message "All components are properly installed and functioning."
        log_message "MediaMTX API: http://localhost:9997"
        log_message "Camera Service: http://localhost:8080"
    fi
}

# Run main function
main "$@"
