#!/bin/bash

# Test Runner for MediaMTX Camera Service Go
# Purpose: Main test runner that delegates to specialized runners
# Usage: ./tests/tools/run_tests.sh [category]

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
        log_error "Test environment not sourced. Please run: source .test_env"
        log_info "Current directory: $(pwd)"
        log_info "Looking for .test_env in: $(pwd)/.test_env"
        log_info "Make sure you're in the mediamtx-camera-service-go directory"
        exit 1
    fi
    log_info "Test environment validated"
}

# Run tests by category using specialized runners
run_unit_tests() {
    log_info "Running unit tests using specialized runner..."
    ./tests/tools/run_unit_tests.sh
}

run_integration_tests() {
    log_info "Running integration tests using specialized runner..."
    ./tests/tools/run_integration_tests.sh
}

run_security_tests() {
    log_info "Running security tests..."
    go test -tags="security" -v ./tests/unit/...
}

run_performance_tests() {
    log_info "Running performance tests..."
    go test -tags="performance" -v ./tests/unit/...
}

run_health_tests() {
    log_info "Running health tests..."
    go test -tags="health" -v ./tests/unit/...
}

run_all_tests() {
    log_info "Running all tests..."
    log_info "Starting with unit tests..."
    run_unit_tests
    
    log_info "Starting with integration tests..."
    run_integration_tests
    
    log_info "Starting with security tests..."
    run_security_tests
    
    log_info "Starting with performance tests..."
    run_performance_tests
    
    log_info "Starting with health tests..."
    run_health_tests
}

# Main function
main() {
    log_info "Starting test runner..."
    
    # Check test environment
    check_test_environment
    
    # Parse category argument
    category=${1:-all}
    
    case $category in
        "unit")
            run_unit_tests
            ;;
        "integration")
            run_integration_tests
            ;;
        "security")
            run_security_tests
            ;;
        "performance")
            run_performance_tests
            ;;
        "health")
            run_health_tests
            ;;
        "all")
            run_all_tests
            ;;
        *)
            log_error "Unknown test category: $category"
            echo "Usage: $0 [unit|integration|security|performance|health|all]"
            echo ""
            echo "Specialized runners:"
            echo "  unit        - Uses run_unit_tests.sh with proper -coverpkg flags"
            echo "  integration - Uses run_integration_tests.sh with real system testing"
            echo "  all         - Runs all test categories in sequence"
            exit 1
            ;;
    esac
    
    log_success "Test run completed"
}

# Run main function
main "$@"
