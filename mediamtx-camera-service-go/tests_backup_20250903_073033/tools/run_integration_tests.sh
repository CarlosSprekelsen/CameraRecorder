#!/bin/bash

# Integration Test Runner for MediaMTX Camera Service Go
# Purpose: Run integration tests with real system testing following testing guidelines
# Usage: ./tests/tools/run_integration_tests.sh
# 
# Following testing guidelines:
# - Uses real system testing over mocking
# - Tests end-to-end workflows
# - Validates API compliance against documentation
# - Tests real MediaMTX service, filesystem, WebSocket connections

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if test environment is sourced
check_test_environment() {
    if [ -z "$CAMERA_SERVICE_JWT_SECRET" ]; then
        log_warning "Test environment not sourced, attempting to source automatically..."
        
        # Try to source from current directory first (since .test_env is here)
        if [ -f ".test_env" ]; then
            log_info "Found .test_env in current directory, sourcing automatically..."
            source .test_env
            
            # Check if it worked
            if [ -n "$CAMERA_SERVICE_JWT_SECRET" ]; then
                log_success "Test environment sourced automatically from .test_env"
                return 0
            fi
        fi
        
        # Try to source from parent directory as fallback
        if [ -f "../.test_env" ]; then
            log_info "Found .test_env in parent directory, sourcing automatically..."
            source ../.test_env
            
            # Check if it worked
            if [ -n "$CAMERA_SERVICE_JWT_SECRET" ]; then
                log_success "Test environment sourced automatically from ../.test_env"
                return 0
            fi
        fi
        
        # If automatic sourcing failed, show error
        log_error "Test environment not sourced. Please run: source .test_env"
        log_info "Current directory: $(pwd)"
        log_info "Looking for .test_env in: $(pwd)/.test_env"
        exit 1
    fi
    log_info "Test environment validated"
}

# Check if MediaMTX service is running
check_mediamtx_service() {
    log_info "Checking MediaMTX service status..."
    
    if systemctl is-active --quiet mediamtx; then
        log_success "MediaMTX service is running"
    else
        log_warning "MediaMTX service is not running"
        log_info "Starting MediaMTX service..."
        if sudo systemctl start mediamtx; then
            log_success "MediaMTX service started successfully"
        else
            log_error "Failed to start MediaMTX service"
            exit 1
        fi
    fi
}

# Check if required services are available
check_required_services() {
    log_info "Checking required services..."
    
    # Check if WebSocket port is available
    if netstat -tuln | grep -q ":8002"; then
        log_success "WebSocket port 8002 is available"
    else
        log_warning "WebSocket port 8002 is not available"
    fi
    
    # Check if camera devices are available
    if ls /dev/video* 1> /dev/null 2>&1; then
        local camera_count=$(ls /dev/video* | wc -l)
        log_success "Found $camera_count camera device(s)"
    else
        log_warning "No camera devices found (/dev/video*)"
    fi
}

# Create coverage directory for integration tests
setup_coverage_dir() {
    COVERAGE_DIR="coverage/integration"
    mkdir -p "$COVERAGE_DIR"
    log_info "Integration coverage directory: $COVERAGE_DIR"
}

# Run integration tests with proper coverage measurement
run_integration_tests() {
    log_info "Starting integration test execution..."
    
    # Run integration tests with real system testing
    # Following guidelines: real MediaMTX service, filesystem, WebSocket connections
    
    local coverage_file="$COVERAGE_DIR/integration_coverage.out"
    
            # ENABLE PARALLEL EXECUTION: Use -parallel 2 for integration tests (conservative due to real services)
        if go test -tags="integration,real_system" -coverpkg=./internal/... ./tests/integration/... -coverprofile="$coverage_file" -parallel 2 -v; then
        log_success "Integration tests completed successfully"
        
        # Analyze coverage if file was created
        if [ -f "$coverage_file" ]; then
            local coverage_percent=$(go tool cover -func="$coverage_file" | grep total | awk '{print $3}' | sed 's/%//')
            log_info "Integration test coverage: ${coverage_percent}%"
            
            # Generate HTML coverage report
            local html_report="$COVERAGE_DIR/integration_coverage.html"
            go tool cover -html="$coverage_file" -o="$html_report"
            log_info "HTML coverage report: $html_report"
        fi
    else
        log_error "Integration tests failed"
        return 1
    fi
}

# Run quarantined integration tests (if any)
run_quarantined_tests() {
    if [ -d "./tests/quarantine/integration" ]; then
        log_info "Running quarantined integration tests..."
        
        local quarantine_coverage="$COVERAGE_DIR/quarantine_coverage.out"
        
        # ENABLE PARALLEL EXECUTION: Use -parallel 2 for quarantined tests (conservative due to real services)
        if go test -tags="integration,real_system" -coverpkg=./internal/... ./tests/quarantine/integration/... -coverprofile="$quarantine_coverage" -parallel 2 -v; then
            log_success "Quarantined integration tests completed"
            
            if [ -f "$quarantine_coverage" ]; then
                local coverage_percent=$(go tool cover -func="$quarantine_coverage" | grep total | awk '{print $3}' | sed 's/%//')
                log_info "Quarantined test coverage: ${coverage_percent}%"
            fi
        else
            log_warning "Quarantined integration tests failed (this may be expected)"
        fi
    else
        log_info "No quarantined integration tests found"
    fi
}

# Run end-to-end workflow tests
run_e2e_tests() {
    log_info "Running end-to-end workflow tests..."
    
    # Look for e2e test files
    local e2e_tests=$(find ./tests/integration/ -name "*e2e*" -o -name "*end_to_end*" 2>/dev/null || true)
    
    if [ -n "$e2e_tests" ]; then
        log_info "Found E2E tests: $e2e_tests"
        
        local e2e_coverage="$COVERAGE_DIR/e2e_coverage.out"
        
        # ENABLE PARALLEL EXECUTION: Use -parallel 2 for E2E tests (conservative due to real services)
        if go test -tags="integration,real_system" -coverpkg=./internal/... ./tests/integration/... -run ".*[Ee]2[Ee].*" -coverprofile="$e2e_coverage" -parallel 2 -v; then
            log_success "E2E tests completed successfully"
            
            if [ -f "$e2e_coverage" ]; then
                local coverage_percent=$(go tool cover -func="$e2e_coverage" | grep total | awk '{print $3}' | sed 's/%//')
                log_info "E2E test coverage: ${coverage_percent}%"
            fi
        else
            log_error "E2E tests failed"
            return 1
        fi
    else
        log_info "No specific E2E test files found"
    fi
}

# Validate API compliance
validate_api_compliance() {
    log_info "Validating API compliance..."
    
    # Check if API documentation exists
    if [ -f "./docs/api/json_rpc_methods.md" ]; then
        log_success "API documentation found"
        
        # Check if tests reference API documentation
        local api_refs=$(grep -r "json_rpc_methods.md" ./tests/integration/ || true)
        if [ -n "$api_refs" ]; then
            log_success "Tests reference API documentation"
        else
            log_warning "Tests may not reference API documentation"
        fi
    else
        log_warning "API documentation not found"
    fi
}

# Generate integration test summary
generate_test_summary() {
    log_info "Generating integration test summary..."
    
    echo "=========================================="
    echo "Integration Test Summary"
    echo "=========================================="
    echo "Coverage Directory: $COVERAGE_DIR"
    echo "Test Categories:"
    echo "  - Standard Integration Tests"
    echo "  - Quarantined Tests (if any)"
    echo "  - End-to-End Workflow Tests"
    echo "  - API Compliance Validation"
    echo ""
    echo "Real System Testing:"
    echo "  - MediaMTX Service: Active"
    echo "  - File System: Real (no mocking)"
    echo "  - WebSocket: Real connections"
    echo "  - Authentication: Real JWT tokens"
    echo "=========================================="
}

# Cleanup function
cleanup() {
    log_info "Cleaning up integration test environment..."
    # Keep coverage files for analysis
}

# Main function
main() {
    log_info "Starting integration test runner following testing guidelines..."
    
    # Check test environment
    check_test_environment
    
    # Check required services
    check_mediamtx_service
    check_required_services
    
    # Setup coverage directory
    setup_coverage_dir
    
    # Run integration tests
    if run_integration_tests; then
        log_success "Standard integration tests completed"
    else
        log_error "Standard integration tests failed"
        exit 1
    fi
    
    # Run quarantined tests
    run_quarantined_tests
    
    # Run E2E tests
    run_e2e_tests
    
    # Validate API compliance
    validate_api_compliance
    
    # Generate summary
    generate_test_summary
    
    log_success "Integration test execution completed"
    
    # Cleanup
    cleanup
}

# Trap cleanup on exit
trap cleanup EXIT

# Run main function
main "$@"
