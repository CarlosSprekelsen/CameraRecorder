#!/bin/bash

# PDR-2: Server Integration Validation Test Runner
# 
# Executes comprehensive PDR-2 validation tests following IV&V standards
# Validates server integration requirements against real MediaMTX server
# 
# PDR-2 Requirements:
# - PDR-2.1: WebSocket connection stability under network interruption
# - PDR-2.2: All JSON-RPC method calls against real MediaMTX server
# - PDR-2.3: Real-time notification handling and state synchronization
# - PDR-2.4: Polling fallback mechanism when WebSocket fails
# - PDR-2.5: API error handling and user feedback mechanisms

set -e

echo "=========================================="
echo "PDR-2: Server Integration Validation"
echo "=========================================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
NC='\033[0m' # No Color

# Test counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
SKIPPED_TESTS=0

# Function to print colored output
print_status() {
    local status=$1
    local message=$2
    
    case $status in
        "PASS")
            echo -e "${GREEN}✓ PASS${NC}: $message"
            ;;
        "FAIL")
            echo -e "${RED}✗ FAIL${NC}: $message"
            ;;
        "SKIP")
            echo -e "${YELLOW}⚠ SKIP${NC}: $message"
            ;;
        "INFO")
            echo -e "${BLUE}ℹ INFO${NC}: $message"
            ;;
        "PDR")
            echo -e "${PURPLE}🔍 PDR-2${NC}: $message"
            ;;
    esac
}

# Function to run PDR-2 test suite
run_pdr_2_test() {
    local test_name=$1
    local test_command=$2
    local description=$3
    
    echo ""
    echo "Running PDR-2 Test: $test_name"
    echo "Description: $description"
    echo "----------------------------------------"
    
    if eval "$test_command"; then
        print_status "PASS" "PDR-2 $test_name completed successfully"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        print_status "FAIL" "PDR-2 $test_name failed"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
}

# Check if we're in the right directory
if [ ! -f "package.json" ]; then
    echo "Error: Must run from client directory"
    exit 1
fi

# Check Node.js version
NODE_VERSION=$(node --version)
print_status "INFO" "Node.js version: $NODE_VERSION"

# Check if dependencies are installed
if [ ! -d "node_modules" ]; then
    print_status "INFO" "Installing dependencies..."
    npm install
fi

# Check MediaMTX Camera Service status
print_status "INFO" "Checking MediaMTX Camera Service status..."
if sudo systemctl is-active --quiet camera-service; then
    print_status "PASS" "MediaMTX Camera Service is running"
else
    print_status "FAIL" "MediaMTX Camera Service is not running"
    print_status "INFO" "Start service with: sudo systemctl start camera-service"
    exit 1
fi

# Check WebSocket endpoint availability
print_status "INFO" "Checking WebSocket endpoint availability..."
if curl -s http://localhost:8002/health > /dev/null 2>&1; then
    print_status "PASS" "WebSocket endpoint is accessible"
else
    print_status "FAIL" "WebSocket endpoint is not accessible"
    print_status "INFO" "Check server configuration and firewall settings"
    exit 1
fi

echo ""
echo "=========================================="
echo "Starting PDR-2 Validation Tests"
echo "=========================================="

# PDR-2.1: WebSocket Connection Stability Under Network Interruption
run_pdr_2_test \
    "PDR-2.1 Network Interruption" \
    "npm test -- tests/integration/test_pdr_2_server_integration_validation.ts --testNamePattern='PDR-2.1.*WebSocket Connection Stability Under Network Interruption' --verbose" \
    "Validates WebSocket connection stability under various network interruption scenarios"

# PDR-2.2: All JSON-RPC Method Calls Against Real MediaMTX Server
run_pdr_2_test \
    "PDR-2.2 JSON-RPC Methods" \
    "npm test -- tests/integration/test_pdr_2_server_integration_validation.ts --testNamePattern='PDR-2.2.*All JSON-RPC Method Calls Against Real MediaMTX Server' --verbose" \
    "Validates all JSON-RPC method calls work correctly against real server"

# PDR-2.3: Real-time Notification Handling and State Synchronization
run_pdr_2_test \
    "PDR-2.3 Real-time Notifications" \
    "npm test -- tests/integration/test_pdr_2_server_integration_validation.ts --testNamePattern='PDR-2.3.*Real-time Notification Handling and State Synchronization' --verbose" \
    "Validates real-time notification handling and state synchronization"

# PDR-2.4: Polling Fallback Mechanism When WebSocket Fails
run_pdr_2_test \
    "PDR-2.4 Polling Fallback" \
    "npm test -- tests/integration/test_pdr_2_server_integration_validation.ts --testNamePattern='PDR-2.4.*Polling Fallback Mechanism When WebSocket Fails' --verbose" \
    "Validates polling fallback mechanism when WebSocket connection fails"

# PDR-2.5: API Error Handling and User Feedback Mechanisms
run_pdr_2_test \
    "PDR-2.5 Error Handling" \
    "npm test -- tests/integration/test_pdr_2_server_integration_validation.ts --testNamePattern='PDR-2.5.*API Error Handling and User Feedback Mechanisms' --verbose" \
    "Validates API error handling and user feedback mechanisms"

# PDR-2: Performance Validation Under Load
run_pdr_2_test \
    "PDR-2 Performance Under Load" \
    "npm test -- tests/integration/test_pdr_2_server_integration_validation.ts --testNamePattern='PDR-2.*Performance Validation Under Load' --verbose" \
    "Validates performance under load conditions"

echo ""
echo "=========================================="
echo "PDR-2 Validation Results Summary"
echo "=========================================="

print_status "INFO" "Total PDR-2 Tests: $TOTAL_TESTS"
print_status "PASS" "Passed: $PASSED_TESTS"
print_status "FAIL" "Failed: $FAILED_TESTS"
print_status "SKIP" "Skipped: $SKIPPED_TESTS"

# Calculate success rate
if [ $TOTAL_TESTS -gt 0 ]; then
    SUCCESS_RATE=$((PASSED_TESTS * 100 / TOTAL_TESTS))
    print_status "INFO" "Success Rate: $SUCCESS_RATE%"
    
    if [ $SUCCESS_RATE -ge 80 ]; then
        print_status "PASS" "PDR-2 Validation: SUCCESS (≥80% pass rate)"
        echo ""
        print_status "PDR" "✅ PDR-2 Server Integration Validation COMPLETED"
        print_status "PDR" "✅ Ready for PDR-2 completion sign-off"
        exit 0
    else
        print_status "FAIL" "PDR-2 Validation: FAILED (<80% pass rate)"
        echo ""
        print_status "PDR" "❌ PDR-2 Server Integration Validation FAILED"
        print_status "PDR" "❌ Must address failed tests before PDR-2 completion"
        exit 1
    fi
else
    print_status "FAIL" "No PDR-2 tests were executed"
    exit 1
fi
