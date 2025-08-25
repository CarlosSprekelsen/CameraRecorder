#!/bin/bash

# Test Runner for MediaMTX Camera Service Go
# Purpose: Basic test runner with Go test integration
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
        exit 1
    fi
    log_info "Test environment validated"
}

# Run tests by category
run_unit_tests() {
    log_info "Running unit tests..."
    go test -tags=unit -v ./...
}

run_integration_tests() {
    log_info "Running integration tests..."
    go test -tags=integration -v ./...
}

run_security_tests() {
    log_info "Running security tests..."
    go test -tags=security -v ./...
}

run_performance_tests() {
    log_info "Running performance tests..."
    go test -tags=performance -v ./...
}

run_health_tests() {
    log_info "Running health tests..."
    go test -tags=health -v ./...
}

run_all_tests() {
    log_info "Running all tests..."
    go test -v ./...
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
            exit 1
            ;;
    esac
    
    log_success "Test run completed"
}

# Run main function
main "$@"
