#!/bin/bash

# Simple Unit Test Runner - Follows testing guidelines exactly
# Purpose: Run unit tests with coverage measurement
# Usage: ./tests/tools/run_unit_tests.sh

set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Simple logging
log() { echo -e "${GREEN}[INFO]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1"; }
warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }

# Test timeout (unit tests should be fast!)
TIMEOUT_SECONDS=60

# Check test environment
if [ -z "$CAMERA_SERVICE_JWT_SECRET" ]; then
    if [ -f ".test_env" ]; then
        source .test_env
        log "Test environment sourced from .test_env"
    else
        error "Test environment not sourced. Run: source .test_env"
        exit 1
    fi
fi

# Create coverage directory
mkdir -p coverage/unit
log "Coverage directory: coverage/unit"

# Function to run tests with timeout
run_package_tests() {
    local package_name=$1
    local test_pattern=$2
    local coverage_file=$3
    
    log "Testing $package_name package (timeout: ${TIMEOUT_SECONDS}s)..."
    
    # Run tests with timeout and moderate verbosity
    # Show test results but filter out excessive INFO messages
    if timeout $TIMEOUT_SECONDS go test -tags="unit" -coverpkg="./internal/$package_name" ./tests/unit/$test_pattern -coverprofile="$coverage_file" -v 2>&1 | grep -E "(PASS|FAIL|RUN|ERROR|WARNING|panic|fatal|coverage|level=warning|level=error)" | grep -v "level=info"; then
        log "✅ $package_name tests passed"
    else
        local exit_code=$?
        if [ $exit_code -eq 124 ]; then
            error "❌ $package_name tests TIMED OUT after ${TIMEOUT_SECONDS}s!"
            error "Unit tests should be fast. Check for hanging tests or integration tests mislabeled as unit tests."
            exit 1
        else
            error "❌ $package_name tests failed"
            exit 1
        fi
    fi
}

# Run unit tests for each package
log "Starting unit tests (max ${TIMEOUT_SECONDS}s per package)..."

# WebSocket tests
run_package_tests "websocket" "test_websocket_*_test.go" "coverage/unit/websocket_coverage.out"

# MediaMTX tests
run_package_tests "mediamtx" "test_mediamtx_*_test.go" "coverage/unit/mediamtx_coverage.out"

# Config tests
run_package_tests "config" "test_config_*_test.go" "coverage/unit/config_coverage.out"

# Camera tests
run_package_tests "camera" "test_camera_*_test.go" "coverage/unit/camera_coverage.out"

# Security tests
run_package_tests "security" "test_security_*_test.go" "coverage/unit/security_coverage.out"

# Logging tests
run_package_tests "logging" "test_logging_*_test.go" "coverage/unit/logging_coverage.out"

# Main tests
run_package_tests "main" "test_main_*_test.go" "coverage/unit/main_coverage.out"

# Generate overall coverage
log "Generating coverage report..."
echo "mode: set" > coverage/unit/combined_coverage.out
for f in coverage/unit/*_coverage.out; do
    if [ -f "$f" ]; then
        tail -n +2 "$f" >> coverage/unit/combined_coverage.out 2>/dev/null || true
    fi
done

# Show overall coverage
overall_coverage=$(go tool cover -func=coverage/unit/combined_coverage.out | grep total | awk '{print $3}' | sed 's/%//')
log "Overall coverage: ${overall_coverage}%"

log "✅ All unit tests completed successfully!"
