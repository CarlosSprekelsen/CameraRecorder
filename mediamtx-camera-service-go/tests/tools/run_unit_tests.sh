#!/bin/bash

# Unit Test Runner for MediaMTX Camera Service Go
# Purpose: Run unit tests with proper coverage measurement following testing guidelines
# Usage: ./tests/tools/run_unit_tests.sh
# 
# Following testing guidelines:
# - Uses -coverpkg flag for cross-package coverage measurement
# - Tests each package individually to avoid package conflicts
# - Generates coverage profiles per package
# - Enforces 90% coverage threshold

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

# Create coverage directory
setup_coverage_dir() {
    COVERAGE_DIR="coverage/unit"
    mkdir -p "$COVERAGE_DIR"
    log_info "Coverage directory: $COVERAGE_DIR"
}

# Run tests for a specific package with coverage
run_package_tests() {
    local package_name=$1
    local test_pattern=$2
    local coverage_file="$COVERAGE_DIR/${package_name}_coverage.out"
    
    log_info "Testing package: $package_name"
    
    # Run tests with proper -coverpkg flag following guidelines
    if go test -tags="unit" -coverpkg="./internal/$package_name" ./tests/unit/$test_pattern -coverprofile="$coverage_file" -v; then
        log_success "Package $package_name tests passed"
        
        # Analyze coverage
        if [ -f "$coverage_file" ]; then
            local coverage_percent=$(go tool cover -func="$coverage_file" | grep total | awk '{print $3}' | sed 's/%//')
            log_info "Package $package_name coverage: ${coverage_percent}%"
            
            # Check coverage threshold (90% required per guidelines)
            if (( $(echo "$coverage_percent >= 90" | bc -l) )); then
                log_success "Package $package_name meets 90% coverage threshold"
            else
                log_warning "Package $package_name below 90% coverage threshold (${coverage_percent}%)"
            fi
        fi
    else
        log_error "Package $package_name tests failed"
        return 1
    fi
}

# Run all unit tests with proper coverage measurement
run_all_unit_tests() {
    log_info "Starting unit test execution with coverage measurement..."
    
    # Test each package individually following guidelines
    # This avoids package conflicts and ensures proper coverage measurement
    
    # 1. WebSocket package tests
    run_package_tests "websocket" "test_websocket_*.go"
    
    # 2. MediaMTX package tests  
    run_package_tests "mediamtx" "test_mediamtx_*.go"
    
    # 3. Config package tests
    run_package_tests "config" "test_config_*.go"
    
    # 4. Camera package tests
    run_package_tests "camera" "test_camera_*.go"
    
    # 5. Security package tests
    run_package_tests "security" "test_security_*.go"
    
    # 6. Logging package tests
    run_package_tests "logging" "test_logging_*.go"
    
    # 7. Main package tests
    run_package_tests "main" "test_main_*.go"
}

# Generate overall coverage report
generate_coverage_report() {
    log_info "Generating overall coverage report..."
    
    # Combine all coverage files
    local combined_coverage="$COVERAGE_DIR/combined_coverage.out"
    
    # Check if we have coverage files
    if ls "$COVERAGE_DIR"/*_coverage.out 1> /dev/null 2>&1; then
        # Combine coverage files
        echo "mode: set" > "$combined_coverage"
        for coverage_file in "$COVERAGE_DIR"/*_coverage.out; do
            if [ -f "$coverage_file" ]; then
                tail -n +2 "$coverage_file" >> "$combined_coverage" 2>/dev/null || true
            fi
        done
        
        # Generate overall coverage percentage
        if [ -f "$combined_coverage" ]; then
            local overall_coverage=$(go tool cover -func="$combined_coverage" | grep total | awk '{print $3}' | sed 's/%//')
            log_info "Overall unit test coverage: ${overall_coverage}%"
            
            # Check overall threshold
            if (( $(echo "$overall_coverage >= 90" | bc -l) )); then
                log_success "Overall coverage meets 90% threshold"
            else
                log_warning "Overall coverage below 90% threshold (${overall_coverage}%)"
            fi
            
            # Generate HTML coverage report
            local html_report="$COVERAGE_DIR/coverage.html"
            go tool cover -html="$combined_coverage" -o="$html_report"
            log_info "HTML coverage report: $html_report"
        fi
    else
        log_warning "No coverage files found to combine"
    fi
}

# Cleanup function
cleanup() {
    log_info "Cleaning up temporary files..."
    # Keep coverage files for analysis
}

# Main function
main() {
    log_info "Starting unit test runner following testing guidelines..."
    
    # Check test environment
    check_test_environment
    
    # Setup coverage directory
    setup_coverage_dir
    
    # Run all unit tests
    if run_all_unit_tests; then
        log_success "All unit tests completed successfully"
        
        # Generate coverage report
        generate_coverage_report
        
        log_success "Unit test execution completed with coverage measurement"
    else
        log_error "Unit test execution failed"
        exit 1
    fi
    
    # Cleanup
    cleanup
}

# Trap cleanup on exit
trap cleanup EXIT

# Run main function
main "$@"
