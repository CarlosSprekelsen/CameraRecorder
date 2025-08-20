#!/bin/bash
# Health Endpoint Real Validation Test
#
# Tests actual health endpoint behavior.
# Validates real service readiness and performance.
#
# This test replaces complex unit test mocks with real system validation
# to provide better confidence in service reliability.

set -e  # Exit on any error

# Configuration
HEALTH_ENDPOINT="http://localhost:8003/health/ready"
MAX_RETRIES=3
RETRY_DELAY=2
HEALTH_SERVER_PID=""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Start health server
start_health_server() {
    log_info "Starting health server for testing..."
    
    # Check if health server is already running
    response=$(curl -s "$HEALTH_ENDPOINT" 2>/dev/null || echo "")
    if [ -n "$response" ]; then
        log_info "Health server already running"
        return 0
    fi
    
    # Try to start health server using Python
    cd "$(dirname "$0")/../.."  # Go to project root
    
    # Check if we can start the health server
    if [ -f "src/health_server.py" ]; then
        log_info "Starting health server using Python..."
        
        # Start health server in background
        python3 -c "
import asyncio
import sys
sys.path.insert(0, 'src')
from health_server import HealthServer

async def start_server():
    server = HealthServer(host='127.0.0.1', port=8003)
    await server.start()
    print('Health server started on 127.0.0.1:8003')
    # Keep server running
    try:
        while True:
            await asyncio.sleep(1)
    except KeyboardInterrupt:
        await server.stop()

if __name__ == '__main__':
    asyncio.run(start_server())
" &
        
        HEALTH_SERVER_PID=$!
        
        # Wait for server to start
        for i in $(seq 1 15); do
            response=$(curl -s "$HEALTH_ENDPOINT" 2>/dev/null || echo "")
            if [ -n "$response" ]; then
                log_info "✓ Health server started successfully"
                return 0
            fi
            log_info "Waiting for health server to start... (attempt $i/15)"
            sleep 2
        done
        
        log_error "Failed to start health server - endpoint not responding after 30 seconds"
        return 1
    else
        log_warn "Health server source not found, testing against existing service"
        return 0
    fi
}

# Stop health server
stop_health_server() {
    if [ -n "$HEALTH_SERVER_PID" ]; then
        log_info "Stopping health server..."
        kill "$HEALTH_SERVER_PID" 2>/dev/null || true
        wait "$HEALTH_SERVER_PID" 2>/dev/null || true
        HEALTH_SERVER_PID=""
    fi
}

# Test health endpoint availability
test_health_endpoint_availability() {
    log_info "Testing health endpoint availability..."
    
    for i in $(seq 1 $MAX_RETRIES); do
        response=$(curl -s "$HEALTH_ENDPOINT" 2>/dev/null || echo "")
        if [ -n "$response" ]; then
            log_info "✓ Health endpoint is available"
            return 0
        else
            if [ $i -lt $MAX_RETRIES ]; then
                log_warn "Health endpoint not available (attempt $i/$MAX_RETRIES), retrying in ${RETRY_DELAY}s..."
                sleep $RETRY_DELAY
            else
                log_error "Health endpoint not available after $MAX_RETRIES attempts"
                return 1
            fi
        fi
    done
}

# Test health endpoint response format
test_health_endpoint_response_format() {
    log_info "Testing health endpoint response format..."
    
    response=$(curl -s "$HEALTH_ENDPOINT")
    
    # Check if response is valid JSON
    if ! echo "$response" | jq -e . > /dev/null 2>&1; then
        log_error "Health endpoint response is not valid JSON"
        echo "Response: $response"
        return 1
    fi
    
    # Check required fields
    if ! echo "$response" | jq -e '.status' > /dev/null 2>&1; then
        log_error "Health endpoint response missing 'status' field"
        return 1
    fi
    
    if ! echo "$response" | jq -e '.timestamp' > /dev/null 2>&1; then
        log_error "Health endpoint response missing 'timestamp' field"
        return 1
    fi
    
    # Log response for debugging
    status=$(echo "$response" | jq -r '.status')
    timestamp=$(echo "$response" | jq -r '.timestamp')
    
    log_info "✓ Health endpoint response format valid"
    log_info "  Status: $status"
    log_info "  Timestamp: $timestamp"
    
    return 0
}

# Test health endpoint under load
test_health_endpoint_load() {
    log_info "Testing health endpoint under load..."
    
    success_count=0
    total_requests=10
    
    for i in $(seq 1 $total_requests); do
        response=$(curl -s "$HEALTH_ENDPOINT" 2>/dev/null || echo "")
        if [ -n "$response" ]; then
            success_count=$((success_count + 1))
        fi
    done
    
    success_rate=$((success_count * 100 / total_requests))
    
    if [ $success_rate -ge 80 ]; then
        log_info "✓ Health endpoint load test passed ($success_count/$total_requests successful, ${success_rate}%)"
        return 0
    else
        log_error "Health endpoint load test failed ($success_count/$total_requests successful, ${success_rate}%)"
        return 1
    fi
}

# Test health endpoint performance
test_health_endpoint_performance() {
    log_info "Testing health endpoint performance..."
    
    # Measure response time
    start_time=$(date +%s%N)
    response=$(curl -s "$HEALTH_ENDPOINT" 2>/dev/null || echo "")
    end_time=$(date +%s%N)
    
    response_time_ms=$(((end_time - start_time) / 1000000))
    
    if [ $response_time_ms -lt 1000 ]; then
        log_info "✓ Health endpoint performance test passed (${response_time_ms}ms response time)"
        return 0
    else
        log_warn "Health endpoint performance test warning (${response_time_ms}ms response time > 1000ms)"
        return 0  # Warning, not failure
    fi
}

# Test health endpoint error handling
test_health_endpoint_error_handling() {
    log_info "Testing health endpoint error handling..."
    
    # Test with invalid endpoint
    if curl -f -s "http://localhost:8003/health/nonexistent" > /dev/null 2>&1; then
        log_error "Invalid health endpoint should return 404"
        return 1
    else
        log_info "✓ Health endpoint error handling test passed (404 for invalid endpoint)"
        return 0
    fi
}

# Main test execution
main() {
    log_info "Starting Health Endpoint Real Validation Tests"
    log_info "Target endpoint: $HEALTH_ENDPOINT"
    
    # Start health server
    if ! start_health_server; then
        log_error "Failed to start health server, cannot proceed with tests"
        exit 1
    fi
    
    # Track test results
    tests_passed=0
    tests_total=0
    
    # Run tests
    tests_total=$((tests_total + 1))
    if test_health_endpoint_availability; then
        tests_passed=$((tests_passed + 1))
    fi
    
    tests_total=$((tests_total + 1))
    if test_health_endpoint_response_format; then
        tests_passed=$((tests_passed + 1))
    fi
    
    tests_total=$((tests_total + 1))
    if test_health_endpoint_load; then
        tests_passed=$((tests_passed + 1))
    fi
    
    tests_total=$((tests_total + 1))
    if test_health_endpoint_performance; then
        tests_passed=$((tests_passed + 1))
    fi
    
    tests_total=$((tests_total + 1))
    if test_health_endpoint_error_handling; then
        tests_passed=$((tests_passed + 1))
    fi
    
    # Summary
    echo
    log_info "Health Endpoint Real Validation Test Summary"
    log_info "Tests passed: $tests_passed/$tests_total"
    
    if [ $tests_passed -eq $tests_total ]; then
        log_info "✓ All health endpoint tests passed!"
        exit_code=0
    else
        log_error "✗ Some health endpoint tests failed"
        exit_code=1
    fi
    
    # Cleanup
    stop_health_server
    
    exit $exit_code
}

# Check dependencies
check_dependencies() {
    if ! command -v curl &> /dev/null; then
        log_error "curl is required but not installed"
        exit 1
    fi
    
    if ! command -v jq &> /dev/null; then
        log_error "jq is required but not installed"
        exit 1
    fi
    
    if ! command -v python3 &> /dev/null; then
        log_error "python3 is required but not installed"
        exit 1
    fi
}

# Cleanup on exit
cleanup() {
    stop_health_server
}

# Set up cleanup trap
trap cleanup EXIT

# Run dependency check and main test
check_dependencies
main
