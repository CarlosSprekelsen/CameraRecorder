#!/bin/bash

# Test Runner Script for MediaMTX Camera Service Client
# Validates all test suites and provides comprehensive reporting

set -e

echo "=========================================="
echo "MediaMTX Camera Service Client Test Suite"
echo "=========================================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
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
    esac
}

# Function to run test suite
run_test_suite() {
    local suite_name=$1
    local test_command=$2
    local description=$3
    
    echo ""
    echo "Running $suite_name..."
    echo "Description: $description"
    echo "----------------------------------------"
    
    if eval "$test_command"; then
        print_status "PASS" "$suite_name completed successfully"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        print_status "FAIL" "$suite_name failed"
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

echo ""
echo "=========================================="
echo "Starting Test Suite Execution"
echo "=========================================="

# 1. Unit Tests - Components
run_test_suite \
    "CameraDetail Component Tests" \
    "npm test -- tests/unit/components/CameraDetail.test.tsx --verbose" \
    "Tests for camera detail component including snapshot and recording controls"

run_test_suite \
    "FileManager Component Tests" \
    "npm test -- tests/unit/components/FileManager.test.tsx --verbose" \
    "Tests for file manager component including file browsing and download functionality"

# 2. Unit Tests - Stores
run_test_suite \
    "File Store Tests" \
    "npm test -- tests/unit/stores/fileStore.test.ts --verbose" \
    "Tests for file store including WebSocket integration and file operations"

# 3. Integration Tests
run_test_suite \
    "Camera Operations Integration Tests" \
    "npm test -- tests/integration/camera-operations-integration.test.ts --verbose" \
    "Integration tests for camera operations with real server validation"

# 4. Existing Tests (for completeness)
run_test_suite \
    "WebSocket Service Tests" \
    "npm test -- tests/unit/services/websocket.test.ts --verbose" \
    "Tests for WebSocket JSON-RPC client implementation"

run_test_suite \
    "WebSocket Integration Tests" \
    "npm test -- tests/integration/websocket-integration.test.ts --verbose" \
    "Integration tests for WebSocket communication"

# 5. Coverage Report
echo ""
echo "=========================================="
echo "Generating Coverage Report"
echo "=========================================="

if npm run test:coverage; then
    print_status "PASS" "Coverage report generated successfully"
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    print_status "FAIL" "Coverage report generation failed"
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi

TOTAL_TESTS=$((TOTAL_TESTS + 1))

# 6. TypeScript Compilation Check
echo ""
echo "=========================================="
echo "TypeScript Compilation Check"
echo "=========================================="

if npx tsc --noEmit; then
    print_status "PASS" "TypeScript compilation successful"
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    print_status "FAIL" "TypeScript compilation failed"
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi

TOTAL_TESTS=$((TOTAL_TESTS + 1))

# 7. Linting Check
echo ""
echo "=========================================="
echo "Code Quality Check"
echo "=========================================="

if npm run lint; then
    print_status "PASS" "Linting passed"
    PASSED_TESTS=$((PASSED_TESTS + 1))
else
    print_status "FAIL" "Linting failed"
    FAILED_TESTS=$((FAILED_TESTS + 1))
fi

TOTAL_TESTS=$((TOTAL_TESTS + 1))

# Summary Report
echo ""
echo "=========================================="
echo "Test Suite Summary"
echo "=========================================="

echo "Total Test Suites: $TOTAL_TESTS"
echo -e "Passed: ${GREEN}$PASSED_TESTS${NC}"
echo -e "Failed: ${RED}$FAILED_TESTS${NC}"
echo -e "Skipped: ${YELLOW}$SKIPPED_TESTS${NC}"

# Calculate success rate
if [ $TOTAL_TESTS -gt 0 ]; then
    SUCCESS_RATE=$((PASSED_TESTS * 100 / TOTAL_TESTS))
    echo "Success Rate: $SUCCESS_RATE%"
fi

echo ""
echo "=========================================="
echo "Test Coverage Summary"
echo "=========================================="

# Display coverage summary if available
if [ -f "coverage/lcov-report/index.html" ]; then
    echo "Coverage report available at: coverage/lcov-report/index.html"
    print_status "INFO" "Open coverage/lcov-report/index.html in your browser to view detailed coverage"
else
    print_status "SKIP" "Coverage report not available"
fi

echo ""
echo "=========================================="
echo "Validation Results"
echo "=========================================="

# Determine overall result
if [ $FAILED_TESTS -eq 0 ]; then
    print_status "PASS" "All test suites passed! ✅"
    echo ""
    echo "Validation Summary:"
    echo "✓ Component functionality validated"
    echo "✓ Store operations tested"
    echo "✓ Integration scenarios covered"
    echo "✓ Type safety verified"
    echo "✓ Code quality standards met"
    echo "✓ Coverage requirements satisfied"
    exit 0
else
    print_status "FAIL" "Some test suites failed! ❌"
    echo ""
    echo "Issues Found:"
    echo "✗ $FAILED_TESTS test suite(s) failed"
    echo "✗ Please review the failed tests above"
    echo "✗ Fix issues before proceeding with deployment"
    exit 1
fi
