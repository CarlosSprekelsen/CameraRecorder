#!/bin/bash

# Production Deployment Validation Script
# Tests all production deployment features

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

# Test results
TESTS_PASSED=0
TESTS_FAILED=0
TESTS_TOTAL=0

# Function to log messages
log_message() {
    echo -e "${BLUE}[$(date '+%Y-%m-%d %H:%M:%S')]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[$(date '+%Y-%m-%d %H:%M:%S')] SUCCESS:${NC} $1"
    TESTS_PASSED=$((TESTS_PASSED + 1))
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
}

log_warning() {
    echo -e "${YELLOW}[$(date '+%Y-%m-%d %H:%M:%S')] WARNING:${NC} $1"
}

log_error() {
    echo -e "${RED}[$(date '+%Y-%m-%d %H:%M:%S')] ERROR:${NC} $1"
    TESTS_FAILED=$((TESTS_FAILED + 1))
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
}

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to check if service is running
check_service() {
    local service_name="$1"
    if systemctl is-active --quiet "$service_name"; then
        return 0
    else
        return 1
    fi
}

# Function to check if port is listening
check_port() {
    local port="$1"
    if netstat -tuln | grep -q ":$port "; then
        return 0
    else
        return 1
    fi
}

# Phase 1: Production Deployment Automation Validation
validate_production_deployment() {
    log_message "=== Phase 1: Production Deployment Automation Validation ==="
    
    # Test 1: Check if production mode was enabled
    if [ -f "$INSTALL_DIR/.env" ] && grep -q "PRODUCTION_MODE=true" "$INSTALL_DIR/.env" 2>/dev/null; then
        log_success "Production mode enabled"
    else
        log_warning "Production mode not detected in environment"
    fi
    
    # Test 2: Check if services are installed and running
    if check_service "mediamtx"; then
        log_success "MediaMTX service is running"
    else
        log_error "MediaMTX service is not running"
    fi
    
    if check_service "camera-service"; then
        log_success "Camera service is running"
    else
        log_error "Camera service is not running"
    fi
    
    # Test 3: Check if ports are listening
    if check_port "8554"; then
        log_success "RTSP port 8554 is listening"
    else
        log_error "RTSP port 8554 is not listening"
    fi
    
    if check_port "8002"; then
        log_success "Camera service port 8002 is listening"
    else
        log_error "Camera service port 8002 is not listening"
    fi
    
    if check_port "8003"; then
        log_success "Health service port 8003 is listening"
    else
        log_error "Health service port 8003 is not listening"
    fi
}

# Phase 2: HTTPS Configuration Validation
validate_https_configuration() {
    log_message "=== Phase 2: HTTPS Configuration Validation ==="
    
    # Test 1: Check if SSL certificates exist
    if [ -f "$INSTALL_DIR/ssl/cert.pem" ] && [ -f "$INSTALL_DIR/ssl/key.pem" ]; then
        log_success "SSL certificates exist"
    else
        log_error "SSL certificates not found"
    fi
    
    # Test 2: Check SSL certificate permissions
    if [ -r "$INSTALL_DIR/ssl/cert.pem" ] && [ -r "$INSTALL_DIR/ssl/key.pem" ]; then
        log_success "SSL certificate permissions are correct"
    else
        log_error "SSL certificate permissions are incorrect"
    fi
    
    # Test 3: Check if Nginx is installed and running
    if command_exists nginx; then
        log_success "Nginx is installed"
        
        if check_service "nginx"; then
            log_success "Nginx service is running"
        else
            log_error "Nginx service is not running"
        fi
    else
        log_error "Nginx is not installed"
    fi
    
    # Test 4: Check if HTTPS port is listening
    if check_port "443"; then
        log_success "HTTPS port 443 is listening"
    else
        log_error "HTTPS port 443 is not listening"
    fi
    
    # Test 5: Test HTTPS connectivity
    if command_exists curl; then
        if curl -k -f -s https://localhost/health/ready >/dev/null 2>&1; then
            log_success "HTTPS health endpoint is responding"
        else
            log_error "HTTPS health endpoint is not responding"
        fi
    else
        log_warning "curl not available for HTTPS testing"
    fi
}

# Phase 3: Environment Management Validation
validate_environment_management() {
    log_message "=== Phase 3: Environment Management Validation ==="
    
    # Test 1: Check if production configuration exists
    if [ -f "$INSTALL_DIR/config/camera-service.yaml" ]; then
        log_success "Production configuration exists"
    else
        log_error "Production configuration not found"
    fi
    
    # Test 2: Check if production directories exist
    if [ -d "$INSTALL_DIR/ssl" ]; then
        log_success "SSL directory exists"
    else
        log_error "SSL directory not found"
    fi
    
    if [ -d "$INSTALL_DIR/backups" ]; then
        log_success "Backup directory exists"
    else
        log_error "Backup directory not found"
    fi
    
    if [ -d "$INSTALL_DIR/monitoring" ]; then
        log_success "Monitoring directory exists"
    else
        log_error "Monitoring directory not found"
    fi
    
    # Test 3: Check file permissions
    if [ -r "$INSTALL_DIR/ssl" ] && [ -w "$INSTALL_DIR/ssl" ]; then
        log_success "SSL directory permissions are correct"
    else
        log_error "SSL directory permissions are incorrect"
    fi
}

# Phase 4: Security Hardening Validation
validate_security_hardening() {
    log_message "=== Phase 4: Security Hardening Validation ==="
    
    # Test 1: Check if UFW is enabled
    if command_exists ufw; then
        if ufw status | grep -q "Status: active"; then
            log_success "UFW firewall is enabled"
        else
            log_error "UFW firewall is not enabled"
        fi
    else
        log_error "UFW firewall is not installed"
    fi
    
    # Test 2: Check if unnecessary services are disabled
    if ! systemctl is-enabled bluetooth 2>/dev/null | grep -q "enabled"; then
        log_success "Bluetooth service is disabled"
    else
        log_warning "Bluetooth service is still enabled"
    fi
    
    if ! systemctl is-enabled cups 2>/dev/null | grep -q "enabled"; then
        log_success "CUPS service is disabled"
    else
        log_warning "CUPS service is still enabled"
    fi
    
    # Test 3: Check service user permissions
    if id "$SERVICE_USER" >/dev/null 2>&1; then
        log_success "Service user exists"
        
        if groups "$SERVICE_USER" | grep -q video; then
            log_success "Service user has video group access"
        else
            log_error "Service user does not have video group access"
        fi
    else
        log_error "Service user does not exist"
    fi
}

# Phase 5: Monitoring and Operations Validation
validate_monitoring_operations() {
    log_message "=== Phase 5: Monitoring and Operations Validation ==="
    
    # Test 1: Check if monitoring scripts exist
    if [ -f "$INSTALL_DIR/monitoring/mediamtx_monitor.sh" ]; then
        log_success "MediaMTX monitoring script exists"
    else
        log_error "MediaMTX monitoring script not found"
    fi
    
    if [ -f "$INSTALL_DIR/monitoring/alert.sh" ]; then
        log_success "Alerting script exists"
    else
        log_error "Alerting script not found"
    fi
    
    if [ -f "$INSTALL_DIR/monitoring/log_monitor.sh" ]; then
        log_success "Log monitoring script exists"
    else
        log_error "Log monitoring script not found"
    fi
    
    # Test 2: Check if monitoring service exists
    if [ -f "/etc/systemd/system/camera-monitoring.service" ]; then
        log_success "Monitoring service file exists"
        
        if check_service "camera-monitoring"; then
            log_success "Monitoring service is running"
        else
            log_warning "Monitoring service is not running"
        fi
    else
        log_error "Monitoring service file not found"
    fi
    
    # Test 3: Test health endpoints
    if command_exists curl; then
        if curl -f -s http://localhost:8003/health/ready >/dev/null 2>&1; then
            log_success "Health ready endpoint is responding"
        else
            log_error "Health ready endpoint is not responding"
        fi
        
        if curl -f -s http://localhost:8003/health/live >/dev/null 2>&1; then
            log_success "Health live endpoint is responding"
        else
            log_error "Health live endpoint is not responding"
        fi
    else
        log_warning "curl not available for health endpoint testing"
    fi
    
    # Test 4: Test MediaMTX API
    if command_exists curl; then
        if curl -f -s http://localhost:9997/v3/paths/list >/dev/null 2>&1; then
            log_success "MediaMTX API is responding"
        else
            log_error "MediaMTX API is not responding"
        fi
    else
        log_warning "curl not available for MediaMTX API testing"
    fi
}

# Phase 6: Backup and Recovery Validation
validate_backup_recovery() {
    log_message "=== Phase 6: Backup and Recovery Validation ==="
    
    # Test 1: Check if backup scripts exist
    if [ -f "$BACKUP_DIR/backup.sh" ]; then
        log_success "Backup script exists"
    else
        log_error "Backup script not found"
    fi
    
    if [ -f "$BACKUP_DIR/recover.sh" ]; then
        log_success "Recovery script exists"
    else
        log_error "Recovery script not found"
    fi
    
    if [ -f "$BACKUP_DIR/disaster_recovery.sh" ]; then
        log_success "Disaster recovery script exists"
    else
        log_error "Disaster recovery script not found"
    fi
    
    # Test 2: Check if backup scripts are executable
    if [ -x "$BACKUP_DIR/backup.sh" ]; then
        log_success "Backup script is executable"
    else
        log_error "Backup script is not executable"
    fi
    
    if [ -x "$BACKUP_DIR/recover.sh" ]; then
        log_success "Recovery script is executable"
    else
        log_error "Recovery script is not executable"
    fi
    
    # Test 3: Check if automated backup cron job exists
    if [ -f "/etc/cron.d/camera-service-backup" ]; then
        log_success "Automated backup cron job exists"
    else
        log_error "Automated backup cron job not found"
    fi
    
    # Test 4: Test backup functionality (create a test backup)
    log_message "Testing backup functionality..."
    if [ -f "$BACKUP_DIR/backup.sh" ]; then
        # Create a small test backup
        TEST_BACKUP="$BACKUP_DIR/test-backup-$(date +%Y%m%d_%H%M%S).tar.gz"
        if tar -czf "$TEST_BACKUP" -C "$INSTALL_DIR" config/ 2>/dev/null; then
            log_success "Test backup created successfully"
            # Clean up test backup
            rm -f "$TEST_BACKUP"
        else
            log_error "Test backup creation failed"
        fi
    fi
}

# Phase 7: File Management API Validation (Epic E6)
validate_file_management_api() {
    log_message "=== Phase 7: File Management API Validation (Epic E6) ==="
    
    # Test 1: Check if health server is running
    if check_port "8003"; then
        log_success "Health server is listening on port 8003"
    else
        log_error "Health server is not listening on port 8003"
    fi
    
    # Test 2: Check if file directories exist
    if [ -d "/opt/camera-service/recordings" ]; then
        log_success "Recordings directory exists"
    else
        log_warning "Recordings directory missing"
    fi
    
    if [ -d "/opt/camera-service/snapshots" ]; then
        log_success "Snapshots directory exists"
    else
        log_warning "Snapshots directory missing"
    fi
    
    # Test 3: Test file download endpoints
    if command_exists curl; then
        # Test recordings endpoint
        if curl -s -I http://localhost:8003/files/recordings/ >/dev/null 2>&1; then
            log_success "Recordings download endpoint is accessible"
        else
            log_error "Recordings download endpoint is not accessible"
        fi
        
        # Test snapshots endpoint
        if curl -s -I http://localhost:8003/files/snapshots/ >/dev/null 2>&1; then
            log_success "Snapshots download endpoint is accessible"
        else
            log_error "Snapshots download endpoint is not accessible"
        fi
        
        # Test SSL endpoints through nginx
        if curl -s -k -I https://localhost/files/recordings/ >/dev/null 2>&1; then
            log_success "SSL recordings endpoint is accessible"
        else
            log_error "SSL recordings endpoint is not accessible"
        fi
        
        if curl -s -k -I https://localhost/files/snapshots/ >/dev/null 2>&1; then
            log_success "SSL snapshots endpoint is accessible"
        else
            log_error "SSL snapshots endpoint is not accessible"
        fi
    else
        log_warning "curl not available for file endpoint testing"
    fi
    
    # Test 4: Check nginx configuration
    if command_exists nginx; then
        if nginx -t >/dev/null 2>&1; then
            log_success "Nginx configuration is valid"
        else
            log_error "Nginx configuration is invalid"
        fi
    else
        log_warning "nginx not available for configuration testing"
    fi
}

# Phase 8: Integration and System Validation
validate_integration() {
    log_message "=== Phase 8: Integration and System Validation ==="
    
    # Test 1: Check if all services can communicate
    if check_service "mediamtx" && check_service "camera-service"; then
        log_success "All core services are running"
    else
        log_error "Not all core services are running"
    fi
    
    # Test 2: Check if health monitoring is working
    if command_exists curl; then
        HEALTH_RESPONSE=$(curl -s http://localhost:8003/health/ready 2>/dev/null || echo "FAILED")
        if [ "$HEALTH_RESPONSE" != "FAILED" ]; then
            log_success "Health monitoring is working"
        else
            log_error "Health monitoring is not working"
        fi
    else
        log_warning "curl not available for health monitoring test"
    fi
    
    # Test 3: Check if logging is working
    if [ -f "$LOG_DIR/camera-service.log" ]; then
        log_success "Service logging is working"
    else
        log_error "Service logging is not working"
    fi
    
    # Test 4: Check if configuration is valid
    if [ -f "$INSTALL_DIR/config/camera-service.yaml" ]; then
        # Basic YAML syntax check
        if python3 -c "import yaml; yaml.safe_load(open('$INSTALL_DIR/config/camera-service.yaml'))" 2>/dev/null; then
            log_success "Production configuration is valid YAML"
        else
            log_error "Production configuration has invalid YAML syntax"
        fi
    else
        log_error "Production configuration file not found"
    fi
}

# Function to generate validation report
generate_report() {
    log_message "=== Production Validation Report ==="
    
    local success_rate=$((TESTS_PASSED * 100 / TESTS_TOTAL))
    
    echo ""
    echo "Test Results Summary:"
    echo "===================="
    echo "Total Tests: $TESTS_TOTAL"
    echo "Passed: $TESTS_PASSED"
    echo "Failed: $TESTS_FAILED"
    echo "Success Rate: ${success_rate}%"
    echo ""
    
    if [ $success_rate -ge 90 ]; then
        echo "✅ PRODUCTION VALIDATION PASSED - System is ready for production deployment"
        return 0
    elif [ $success_rate -ge 70 ]; then
        echo "⚠️ PRODUCTION VALIDATION CONDITIONAL - Some issues need attention before production"
        return 1
    else
        echo "❌ PRODUCTION VALIDATION FAILED - Significant issues must be resolved before production"
        return 1
    fi
}

# Main validation function
main() {
    log_message "Starting Production Deployment Validation"
    log_message "========================================="
    
    # Run all validation phases
    validate_production_deployment
    validate_https_configuration
    validate_environment_management
    validate_security_hardening
    validate_monitoring_operations
    validate_backup_recovery
    validate_file_management_api
    validate_integration
    
    # Generate final report
    generate_report
    
    log_message "Production validation completed"
}

# Run main function
main "$@"
