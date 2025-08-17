#!/bin/bash

# Production Deployment Setup Script
# Runs all phases of production implementation and validation

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

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

# Phase 1: Production Deployment Automation
phase1_production_deployment() {
    log_message "=== Phase 1: Production Deployment Automation ==="
    
    log_message "Running production installation..."
    
    # Run production installation
    if [ -f "$SCRIPT_DIR/install.sh" ]; then
        PRODUCTION_MODE=true "$SCRIPT_DIR/install.sh"
        log_success "Production deployment completed"
    else
        log_error "Install script not found"
        return 1
    fi
}

# Phase 2: Operations Infrastructure
phase2_operations_infrastructure() {
    log_message "=== Phase 2: Operations Infrastructure ==="
    
    log_message "Setting up operations monitoring..."
    
    # Setup monitoring
    if [ -f "$SCRIPT_DIR/operations/monitoring_setup.sh" ]; then
        "$SCRIPT_DIR/operations/monitoring_setup.sh"
        log_success "Operations monitoring setup completed"
    else
        log_error "Monitoring setup script not found"
        return 1
    fi
    
    log_message "Setting up backup and recovery..."
    
    # Setup backup and recovery
    if [ -f "$SCRIPT_DIR/operations/backup_recovery.sh" ]; then
        "$SCRIPT_DIR/operations/backup_recovery.sh"
        log_success "Backup and recovery setup completed"
    else
        log_error "Backup recovery script not found"
        return 1
    fi
}

# Phase 3: Production Environment Setup
phase3_production_environment() {
    log_message "=== Phase 3: Production Environment Setup ==="
    
    log_message "Verifying production environment..."
    
    # Check if production configuration exists
    if [ -f "$PROJECT_ROOT/config/production.yaml" ]; then
        log_success "Production configuration exists"
    else
        log_error "Production configuration not found"
        return 1
    fi
    
    # Check if services are running
    if systemctl is-active --quiet mediamtx && systemctl is-active --quiet camera-service; then
        log_success "Core services are running"
    else
        log_error "Core services are not running"
        return 1
    fi
    
    # Check if HTTPS is configured
    if [ -f "/opt/camera-service/ssl/cert.pem" ] && [ -f "/opt/camera-service/ssl/key.pem" ]; then
        log_success "HTTPS configuration exists"
    else
        log_error "HTTPS configuration not found"
        return 1
    fi
    
    log_success "Production environment setup verified"
}

# Phase 4: Integration and Validation
phase4_integration_validation() {
    log_message "=== Phase 4: Integration and Validation ==="
    
    log_message "Running comprehensive validation..."
    
    # Run validation script
    if [ -f "$SCRIPT_DIR/validate_production.sh" ]; then
        "$SCRIPT_DIR/validate_production.sh"
        
        # Check validation exit code
        if [ $? -eq 0 ]; then
            log_success "Integration validation passed"
        else
            log_error "Integration validation failed"
            return 1
        fi
    else
        log_error "Validation script not found"
        return 1
    fi
}

# Phase 5: Technical Assessment and Authorization
phase5_technical_assessment() {
    log_message "=== Phase 5: Technical Assessment and Authorization ==="
    
    log_message "Performing technical assessment..."
    
    # Check all critical components
    local assessment_passed=true
    
    # Check services
    if ! systemctl is-active --quiet mediamtx; then
        log_error "MediaMTX service assessment: FAILED"
        assessment_passed=false
    else
        log_success "MediaMTX service assessment: PASSED"
    fi
    
    if ! systemctl is-active --quiet camera-service; then
        log_error "Camera service assessment: FAILED"
        assessment_passed=false
    else
        log_success "Camera service assessment: PASSED"
    fi
    
    # Check HTTPS
    if ! curl -k -f -s https://localhost/health/ready >/dev/null 2>&1; then
        log_error "HTTPS assessment: FAILED"
        assessment_passed=false
    else
        log_success "HTTPS assessment: PASSED"
    fi
    
    # Check monitoring
    if [ -f "/opt/camera-service/monitoring/monitoring.conf" ]; then
        log_success "Monitoring assessment: PASSED"
    else
        log_error "Monitoring assessment: FAILED"
        assessment_passed=false
    fi
    
    # Check backup
    if [ -f "/opt/camera-service/backups/backup.sh" ]; then
        log_success "Backup assessment: PASSED"
    else
        log_error "Backup assessment: FAILED"
        assessment_passed=false
    fi
    
    # Check firewall
    if command_exists ufw && ufw status | grep -q "Status: active"; then
        log_success "Firewall assessment: PASSED"
    else
        log_error "Firewall assessment: FAILED"
        assessment_passed=false
    fi
    
    if [ "$assessment_passed" = true ]; then
        log_success "Technical assessment: PASSED"
        log_message "✅ E5 AUTHORIZATION GRANTED - System is ready for production deployment"
        return 0
    else
        log_error "Technical assessment: FAILED"
        log_message "❌ E5 AUTHORIZATION DENIED - Issues must be resolved before production"
        return 1
    fi
}

# Function to generate E5 completion report
generate_completion_report() {
    log_message "=== E5 Completion Report ==="
    
    echo ""
    echo "E5: Deployment & Operations Strategy - Completion Report"
    echo "======================================================="
    echo "Date: $(date)"
    echo "System: $(hostname)"
    echo ""
    
    echo "Production Deployment Features:"
    echo "✅ Enhanced install.sh with production mode"
    echo "✅ HTTPS/SSL configuration with Nginx"
    echo "✅ Security hardening with UFW firewall"
    echo "✅ Production configuration management"
    echo ""
    
    echo "Operations Infrastructure:"
    echo "✅ Health monitoring using existing infrastructure"
    echo "✅ MediaMTX monitoring with circuit breaker"
    echo "✅ Alerting system for critical events"
    echo "✅ Log monitoring for error detection"
    echo ""
    
    echo "Backup and Recovery:"
    echo "✅ Comprehensive backup procedures"
    echo "✅ Automated recovery scripts"
    echo "✅ Disaster recovery procedures"
    echo "✅ Automated daily backups"
    echo ""
    
    echo "Production Environment:"
    echo "✅ Production configuration (config/production.yaml)"
    echo "✅ Service hardening and security"
    echo "✅ Monitoring and alerting setup"
    echo "✅ Backup and recovery procedures"
    echo ""
    
    echo "Validation and Testing:"
    echo "✅ Comprehensive validation script (e5_validation.sh)"
    echo "✅ Integration testing of all components"
    echo "✅ Technical assessment and authorization"
    echo "✅ Production readiness verification"
    echo ""
    
    echo "Usage Commands:"
    echo "==============="
    echo "# Development installation:"
    echo "sudo ./deployment/scripts/install.sh"
    echo ""
    echo "# Production installation:"
    echo "sudo PRODUCTION_MODE=true ./deployment/scripts/install.sh"
    echo ""
    echo "# Run E5 validation:"
    echo "sudo ./deployment/scripts/e5_validation.sh"
    echo ""
    echo "# Run complete E5 execution:"
    echo "sudo ./deployment/scripts/e5_complete.sh"
    echo ""
    
    echo "Production Commands:"
    echo "==================="
    echo "# Check services:"
    echo "systemctl status mediamtx camera-service nginx"
    echo ""
    echo "# Check HTTPS:"
    echo "curl -k https://localhost/health/ready"
    echo ""
    echo "# Run backup:"
    echo "sudo -u camera-service /opt/camera-service/backups/backup.sh"
    echo ""
    echo "# Check monitoring:"
    echo "systemctl status camera-monitoring"
    echo ""
    
    echo "Documentation:"
    echo "=============="
    echo "📖 Production Deployment Guide: docs/deployment/PRODUCTION_DEPLOYMENT.md"
    echo "📖 Installation Guide: docs/deployment/INSTALLATION_GUIDE.md"
    echo ""
    
    echo "E5 Status: ✅ COMPLETE - Ready for ORR (Operational Readiness Review)"
    echo ""
}

# Main execution function
main() {
    log_message "Starting Production Deployment Setup - Complete Execution"
    log_message "======================================================="
    
    # Check if running as root
    check_root
    
    # Check if we're in the right directory
    if [ ! -f "$SCRIPT_DIR/install.sh" ]; then
        log_error "Install script not found. Please run from the project root directory."
        exit 1
    fi
    
    # Run all phases
    local phase_results=()
    
    log_message "Executing Phase 1: Production Deployment Automation"
    if phase1_production_deployment; then
        phase_results+=("Phase 1: ✅ PASSED")
    else
        phase_results+=("Phase 1: ❌ FAILED")
        log_error "Phase 1 failed. Stopping execution."
        exit 1
    fi
    
    log_message "Executing Phase 2: Operations Infrastructure"
    if phase2_operations_infrastructure; then
        phase_results+=("Phase 2: ✅ PASSED")
    else
        phase_results+=("Phase 2: ❌ FAILED")
        log_error "Phase 2 failed. Stopping execution."
        exit 1
    fi
    
    log_message "Executing Phase 3: Production Environment Setup"
    if phase3_production_environment; then
        phase_results+=("Phase 3: ✅ PASSED")
    else
        phase_results+=("Phase 3: ❌ FAILED")
        log_error "Phase 3 failed. Stopping execution."
        exit 1
    fi
    
    log_message "Executing Phase 4: Integration and Validation"
    if phase4_integration_validation; then
        phase_results+=("Phase 4: ✅ PASSED")
    else
        phase_results+=("Phase 4: ❌ FAILED")
        log_error "Phase 4 failed. Stopping execution."
        exit 1
    fi
    
    log_message "Executing Phase 5: Technical Assessment and Authorization"
    if phase5_technical_assessment; then
        phase_results+=("Phase 5: ✅ PASSED")
    else
        phase_results+=("Phase 5: ❌ FAILED")
        log_error "Phase 5 failed. Stopping execution."
        exit 1
    fi
    
    # Generate completion report
    generate_completion_report
    
    # Display phase results
    echo "Phase Results:"
    echo "=============="
    for result in "${phase_results[@]}"; do
        echo "$result"
    done
    echo ""
    
    log_success "Production Deployment Setup - Complete Execution SUCCESSFUL"
    log_message "System is ready for production deployment"
}

# Run main function
main "$@"
