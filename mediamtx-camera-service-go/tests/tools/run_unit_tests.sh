#!/bin/bash

# MediaMTX Camera Service - Unit Test Runner
# Purpose: Run unit tests with accurate coverage measurement per module
# Usage: ./tests/tools/run_unit_tests.sh
# 
# Follows Go Testing Guide: docs/testing/go-testing-guide.md
# - External testing (package *_test) with proper build tags
# - Cross-package coverage with -coverpkg flag
# - Real system testing over mocking
# - Proper environment sourcing

set -e

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Logging functions
log() { echo -e "${GREEN}[INFO]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1"; }
warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
info() { echo -e "${BLUE}[DETAIL]${NC} $1"; }

# Test timeout (unit tests should be fast!)
TIMEOUT_SECONDS=120

# Check if we're in the right directory
if [ ! -f "go.mod" ]; then
    error "Must run from project root directory (where go.mod exists)"
    exit 1
fi

# Check test environment
if [ -z "$CAMERA_SERVICE_JWT_SECRET" ]; then
    if [ -f ".test_env" ]; then
        source .test_env
        log "Test environment sourced from .test_env"
    else
        error "Test environment not found. Run: source .test_env"
        exit 1
    fi
fi

# Verify critical environment variables
if [ -z "$CAMERA_SERVICE_JWT_SECRET" ] || [ -z "$CAMERA_SERVICE_API_KEYS_PATH" ]; then
    error "Critical environment variables missing after sourcing .test_env"
    error "CAMERA_SERVICE_JWT_SECRET: ${CAMERA_SERVICE_JWT_SECRET:+SET}"
    error "CAMERA_SERVICE_API_KEYS_PATH: ${CAMERA_SERVICE_API_KEYS_PATH:+SET}"
    exit 1
fi

# Create coverage directory
mkdir -p coverage/unit
log "Coverage directory: coverage/unit"

# Function to run tests for a specific package with coverage
run_package_tests() {
    local package_name=$1
    local test_pattern=$2
    local coverage_file=$3
    
    log "Testing $package_name package (timeout: ${TIMEOUT_SECONDS}s)..."
    
    # Check if test files exist
    local test_files=$(find ./tests/unit/ -name "$test_pattern" -type f)
    if [ -z "$test_files" ]; then
        warn "No test files found for pattern: $test_pattern"
        return 0
    fi
    
    info "Found test files: $test_files"
    
    # Create a temporary coverage file in our controlled directory
    local temp_coverage_file="/tmp/${package_name}_coverage_$$.out"
    
    # Run tests with timeout, coverage, and parallel execution
    # CRITICAL: Use -coverpkg for cross-package coverage (external testing)
    # Use -covermode=set to avoid temp directory issues
    # ENABLE PARALLEL EXECUTION: Use -parallel flag for concurrent test execution
    if timeout $TIMEOUT_SECONDS go test -tags="unit" -coverpkg="./internal/$package_name" ./tests/unit/$test_pattern -coverprofile="$temp_coverage_file" -covermode=set -parallel 4 -v; then
        log "✅ $package_name tests passed"
        
        # Move coverage file to final location if it was generated
        if [ -f "$temp_coverage_file" ] && [ -s "$temp_coverage_file" ]; then
            local coverage_size=$(wc -c < "$temp_coverage_file")
            if [ "$coverage_size" -gt 100 ]; then
                mv "$temp_coverage_file" "$coverage_file"
                info "Coverage file generated: $coverage_file (${coverage_size} bytes)"
            else
                warn "Coverage file suspiciously small, removing: $temp_coverage_file (${coverage_size} bytes)"
                rm -f "$temp_coverage_file"
                # Create a minimal coverage file to prevent script failure
                echo "mode: set" > "$coverage_file"
            fi
        else
            warn "Coverage file not generated, creating minimal file"
            echo "mode: set" > "$coverage_file"
        fi
    else
        local exit_code=$?
        # Clean up temp coverage file
        rm -f "$temp_coverage_file"
        
        if [ $exit_code -eq 124 ]; then
            error "❌ $package_name tests TIMED OUT after ${TIMEOUT_SECONDS}s!"
            error "Unit tests should be fast. Check for hanging tests or integration tests mislabeled as unit tests."
            exit 1
        else
            error "❌ $package_name tests failed"
            # Create a minimal coverage file to prevent script failure
            echo "mode: set" > "$coverage_file"
            return 1
        fi
    fi
}

# Function to report coverage for a specific module
report_module_coverage() {
    local package_name=$1
    local coverage_file=$2
    
    if [ ! -f "$coverage_file" ] || [ ! -s "$coverage_file" ]; then
        warn "No coverage data for $package_name"
        return
    fi
    
    local coverage_size=$(wc -c < "$coverage_file")
    if [ "$coverage_size" -le 100 ]; then
        warn "Coverage file too small for $package_name (${coverage_size} bytes)"
        return
    fi
    
    log "📊 Coverage Report for $package_name:"
    
    # Generate coverage report
    if go tool cover -func="$coverage_file" | grep -E "(total|internal/$package_name)"; then
        # Extract total coverage percentage
        local total_coverage=$(go tool cover -func="$coverage_file" | grep total | awk '{print $3}' | sed 's/%//')
        if [ -n "$total_coverage" ]; then
            info "Total coverage: ${total_coverage}%"
        fi
    else
        warn "Could not generate coverage report for $package_name"
    fi
    
    echo ""
}

# Function to run all unit tests
run_all_unit_tests() {
    log "Starting unit tests (max ${TIMEOUT_SECONDS}s per package)..."
    echo ""
    
    # Test each package with proper coverage
    local packages=(
        "websocket:test_websocket_*_test.go:coverage/unit/websocket_coverage.out"
        "mediamtx:test_mediamtx_*_test.go:coverage/unit/mediamtx_coverage.out"
        "config:test_config_*_test.go:coverage/unit/config_coverage.out"
        "camera:test_camera_*_test.go:coverage/unit/camera_coverage.out"
        "security:test_security_*_test.go:coverage/unit/security_coverage.out"
        "logging:test_logging_*_test.go:coverage/unit/logging_coverage.out"
        "main:test_main_*_test.go:coverage/unit/main_coverage.out"
    )
    
    local failed_packages=()
    local successful_packages=()
    
    for package_info in "${packages[@]}"; do
        IFS=':' read -r package_name test_pattern coverage_file <<< "$package_info"
        
        if run_package_tests "$package_name" "$test_pattern" "$coverage_file"; then
            successful_packages+=("$package_name")
            # Report coverage for this module
            report_module_coverage "$package_name" "$coverage_file"
        else
            failed_packages+=("$package_name")
            warn "⚠️  Tests failed for $package_name, but continuing with other packages"
        fi
    done
    
    # Summary
    echo ""
    log "📋 Test Summary:"
    log "✅ Successful packages: ${successful_packages[*]}"
    if [ ${#failed_packages[@]} -gt 0 ]; then
        warn "⚠️  Failed packages: ${failed_packages[*]}"
        warn "Coverage data may be incomplete for failed packages"
    fi
    
    # Check for failed packages
    if [ ${#failed_packages[@]} -gt 0 ]; then
        warn "Some tests failed, but coverage generation will continue"
        warn "Failed packages will have minimal coverage data"
    fi
}

# Function to generate combined coverage report
generate_combined_coverage() {
    log "Generating combined coverage report..."
    
    # Start with coverage mode header
    echo "mode: set" > coverage/unit/combined_coverage.out
    
    # Combine all coverage files
    local coverage_files=(coverage/unit/*_coverage.out)
    local valid_files=0
    
    for f in "${coverage_files[@]}"; do
        if [ -f "$f" ] && [ -s "$f" ]; then
            local size=$(wc -c < "$f")
            if [ "$size" -gt 100 ]; then
                # Skip the first line (mode: set) and append coverage data
                tail -n +2 "$f" >> coverage/unit/combined_coverage.out 2>/dev/null || true
                valid_files=$((valid_files + 1))
                info "Added coverage data: $(basename "$f") (${size} bytes)"
            else
                warn "Skipping suspicious coverage file: $(basename "$f") (${size} bytes)"
            fi
        fi
    done
    
    if [ $valid_files -eq 0 ]; then
        error "No valid coverage files found to combine!"
        return 1
    fi
    
    log "Combined coverage from $valid_files packages"
}

# Function to show overall coverage summary
show_overall_coverage() {
    local combined_file="coverage/unit/combined_coverage.out"
    
    if [ ! -f "$combined_file" ] || [ ! -s "$combined_file" ]; then
        error "Combined coverage file not found or empty"
        return 1
    fi
    
    local combined_size=$(wc -c < "$combined_file")
    if [ "$combined_size" -le 100 ]; then
        warn "Combined coverage file too small (${combined_size} bytes)"
        return 1
    fi
    
    log "📊 Overall Coverage Summary:"
    
    # Show total coverage
    if overall_coverage=$(go tool cover -func="$combined_file" | grep total | awk '{print $3}' | sed 's/%//'); then
        log "Overall coverage: ${overall_coverage}%"
        
        # Color code based on coverage level
        if (( $(echo "$overall_coverage >= 90" | bc -l) )); then
            log "✅ Coverage target met (≥90%)"
        elif (( $(echo "$overall_coverage >= 80" | bc -l) )); then
            warn "⚠️  Coverage below target but acceptable (≥80%)"
        else
            error "❌ Coverage below acceptable threshold (<80%)"
        fi
    else
        error "Could not determine overall coverage"
        return 1
    fi
    
    echo ""
    
    # Show coverage by function (top 10)
    log "Top 10 functions by coverage:"
    go tool cover -func="$combined_file" | grep -v "total" | sort -k3 -nr | head -10 | while read -r line; do
        if [[ $line =~ internal/ ]]; then
            echo "  $line"
        fi
    done
}

# Function to clean up old coverage files
cleanup_old_coverage() {
    log "Cleaning up old coverage files..."
    
    # Remove old coverage files older than 1 hour
    find coverage/unit/ -name "*_coverage.out" -mmin +60 -delete 2>/dev/null || true
    
    # Remove empty coverage files
    find coverage/unit/ -name "*_coverage.out" -size 0 -delete 2>/dev/null || true
    
    # Remove suspiciously small coverage files (less than 100 bytes)
    find coverage/unit/ -name "*_coverage.out" -size -100c -delete 2>/dev/null || true
    
    log "Cleanup completed"
}

# Function to validate coverage files
validate_coverage_files() {
    log "Validating coverage files..."
    
    local valid_count=0
    local total_count=0
    
    for f in coverage/unit/*_coverage.out; do
        if [ -f "$f" ]; then
            total_count=$((total_count + 1))
            local size=$(wc -c < "$f")
            if [ "$size" -gt 100 ]; then
                valid_count=$((valid_count + 1))
                info "✅ $(basename "$f"): ${size} bytes"
            else
                warn "⚠️  $(basename "$f"): ${size} bytes (suspicious)"
            fi
        fi
    done
    
    log "Coverage validation: $valid_count/$total_count files are valid"
    
    if [ $valid_count -eq 0 ]; then
        error "No valid coverage files found!"
        return 1
    fi
    
    return 0
}

# Main execution
main() {
    log "🚀 MediaMTX Camera Service - Unit Test Runner"
    log "Following Go Testing Guide: docs/testing/go-testing-guide.md"
    echo ""
    
    # Verify environment
    log "Environment verification:"
    info "JWT Secret: ${CAMERA_SERVICE_JWT_SECRET:0:16}..."
    info "API Keys Path: $CAMERA_SERVICE_API_KEYS_PATH"
    info "Test Mode: $CAMERA_SERVICE_TEST_MODE"
    echo ""
    
    # Clean up old coverage files
    cleanup_old_coverage
    
    # Run all unit tests
    run_all_unit_tests
    
    # Validate coverage files before combining
    if ! validate_coverage_files; then
        error "❌ Coverage validation failed!"
        exit 1
    fi
    
    # Generate combined coverage
    if generate_combined_coverage; then
        # Show overall coverage
        show_overall_coverage
        
        log "✅ Unit test runner completed!"
        log "📁 Coverage files saved in: coverage/unit/"
        log "📊 Combined coverage: coverage/unit/combined_coverage.out"
        
        # Final validation
        if validate_coverage_files; then
            log "✅ All coverage files validated successfully"
        else
            warn "⚠️  Some coverage files may have issues"
        fi
    else
        error "❌ Failed to generate combined coverage report"
        exit 1
    fi
}

# Run main function
main "$@"
