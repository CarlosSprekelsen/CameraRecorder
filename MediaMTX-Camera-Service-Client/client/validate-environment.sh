#!/bin/bash

# =============================================================================
# MediaMTX Camera Service Client - Environment Validation Script
# =============================================================================
# 
# This script validates that the development environment meets all requirements
# for the React 19.1.1 + Node.js 20.19.x setup.
#
# Usage: ./validate-environment.sh
# =============================================================================

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}==========================================${NC}"
echo -e "${BLUE}MediaMTX Camera Service Client${NC}"
echo -e "${BLUE}Environment Validation Script${NC}"
echo -e "${BLUE}==========================================${NC}"
echo ""

# Function to print colored output
print_status() {
    local status=$1
    local message=$2
    
    case $status in
        "INFO")
            echo -e "${BLUE}ℹ INFO${NC}: $message"
            ;;
        "SUCCESS")
            echo -e "${GREEN}✅ SUCCESS${NC}: $message"
            ;;
        "WARNING")
            echo -e "${YELLOW}⚠ WARNING${NC}: $message"
            ;;
        "ERROR")
            echo -e "${RED}❌ ERROR${NC}: $message"
            ;;
    esac
}

# Function to check version requirements
check_version() {
    local current=$1
    local required=$2
    local name=$3
    
    if [[ "$current" < "$required" ]]; then
        print_status "ERROR" "$name version $current is below required $required"
        return 1
    else
        print_status "SUCCESS" "$name version $current meets requirements"
        return 0
    fi
}

# Validation counters
TOTAL_CHECKS=0
PASSED_CHECKS=0
FAILED_CHECKS=0

# Check 1: Node.js Version
print_status "INFO" "Checking Node.js version..."
NODE_VERSION=$(node --version)
TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

if check_version "$NODE_VERSION" "v20.19.0" "Node.js"; then
    PASSED_CHECKS=$((PASSED_CHECKS + 1))
else
    FAILED_CHECKS=$((FAILED_CHECKS + 1))
fi

# Check 2: npm Version
print_status "INFO" "Checking npm version..."
NPM_VERSION=$(npm --version)
TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

if check_version "$NPM_VERSION" "10.8.0" "npm"; then
    PASSED_CHECKS=$((PASSED_CHECKS + 1))
else
    FAILED_CHECKS=$((FAILED_CHECKS + 1))
fi

# Check 3: nvm Availability (optional but recommended)
print_status "INFO" "Checking nvm availability..."
TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

if command -v nvm &> /dev/null; then
    print_status "SUCCESS" "nvm is available for Node.js version management"
    PASSED_CHECKS=$((PASSED_CHECKS + 1))
else
    print_status "WARNING" "nvm not found - recommended for Node.js version management"
    print_status "INFO" "Install nvm: curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh | bash"
    PASSED_CHECKS=$((PASSED_CHECKS + 1))  # Not a failure, just a warning
fi

# Check 4: Project Structure
print_status "INFO" "Checking project structure..."
TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

if [ -f "package.json" ] && [ -f "tsconfig.json" ] && [ -d "src" ] && [ -d "tests" ]; then
    print_status "SUCCESS" "Project structure is valid"
    PASSED_CHECKS=$((PASSED_CHECKS + 1))
else
    print_status "ERROR" "Project structure is incomplete"
    FAILED_CHECKS=$((FAILED_CHECKS + 1))
fi

# Check 5: Dependencies Installation
print_status "INFO" "Checking dependencies..."
TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

if [ -d "node_modules" ]; then
    print_status "SUCCESS" "Dependencies are installed"
    PASSED_CHECKS=$((PASSED_CHECKS + 1))
else
    print_status "WARNING" "Dependencies not installed - run 'npm install'"
    FAILED_CHECKS=$((FAILED_CHECKS + 1))
fi

# Check 6: TypeScript Compilation
print_status "INFO" "Checking TypeScript compilation..."
TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

if npx tsc --noEmit > /dev/null 2>&1; then
    print_status "SUCCESS" "TypeScript compilation successful"
    PASSED_CHECKS=$((PASSED_CHECKS + 1))
else
    print_status "ERROR" "TypeScript compilation failed"
    FAILED_CHECKS=$((FAILED_CHECKS + 1))
fi

# Check 7: Test Environment
print_status "INFO" "Checking test environment..."
TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

if [ -f "jest.config.cjs" ] && [ -f "tests/setup.ts" ]; then
    print_status "SUCCESS" "Test environment is configured"
    PASSED_CHECKS=$((PASSED_CHECKS + 1))
else
    print_status "ERROR" "Test environment is not properly configured"
    FAILED_CHECKS=$((FAILED_CHECKS + 1))
fi

# Check 8: React Version
print_status "INFO" "Checking React version..."
TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

if [ -d "node_modules" ]; then
    REACT_VERSION=$(npm list react --depth=0 2>/dev/null | grep react@ | cut -d' ' -f2 || echo "not installed")
    if [[ "$REACT_VERSION" == "react@19.1.1" ]]; then
        print_status "SUCCESS" "React 19.1.1 is installed"
        PASSED_CHECKS=$((PASSED_CHECKS + 1))
    else
        print_status "ERROR" "React version mismatch - expected 19.1.1, found $REACT_VERSION"
        FAILED_CHECKS=$((FAILED_CHECKS + 1))
    fi
else
    print_status "WARNING" "Cannot check React version - dependencies not installed"
    FAILED_CHECKS=$((FAILED_CHECKS + 1))
fi

# Check 9: Build Process
print_status "INFO" "Checking build process..."
TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

if npm run build > /dev/null 2>&1; then
    print_status "SUCCESS" "Build process successful"
    PASSED_CHECKS=$((PASSED_CHECKS + 1))
else
    print_status "ERROR" "Build process failed"
    FAILED_CHECKS=$((FAILED_CHECKS + 1))
fi

# Check 10: Test Execution
print_status "INFO" "Checking test execution..."
TOTAL_CHECKS=$((TOTAL_CHECKS + 1))

if npm test > /dev/null 2>&1; then
    print_status "SUCCESS" "Test execution successful"
    PASSED_CHECKS=$((PASSED_CHECKS + 1))
else
    print_status "ERROR" "Test execution failed"
    FAILED_CHECKS=$((FAILED_CHECKS + 1))
fi

# Summary Report
echo ""
echo -e "${BLUE}==========================================${NC}"
echo -e "${BLUE}Validation Summary${NC}"
echo -e "${BLUE}==========================================${NC}"

echo "Total Checks: $TOTAL_CHECKS"
echo -e "Passed: ${GREEN}$PASSED_CHECKS${NC}"
echo -e "Failed: ${RED}$FAILED_CHECKS${NC}"

# Calculate success rate
if [ $TOTAL_CHECKS -gt 0 ]; then
    SUCCESS_RATE=$((PASSED_CHECKS * 100 / TOTAL_CHECKS))
    echo "Success Rate: $SUCCESS_RATE%"
fi

echo ""
echo -e "${BLUE}==========================================${NC}"
echo -e "${BLUE}Environment Status${NC}"
echo -e "${BLUE}==========================================${NC}"

if [ $FAILED_CHECKS -eq 0 ]; then
    print_status "SUCCESS" "Environment is fully validated and ready for development!"
    echo ""
    echo -e "${GREEN}🎉 Your development environment is ready!${NC}"
    echo -e "${GREEN}✅ All requirements are met${NC}"
    echo -e "${GREEN}✅ React 19.1.1 + Node.js 20.19.x setup is working${NC}"
    echo -e "${GREEN}✅ Build and test processes are functional${NC}"
    echo ""
    echo -e "${BLUE}Next Steps:${NC}"
    echo -e "  ${YELLOW}npm run dev${NC}     - Start development server"
    echo -e "  ${YELLOW}npm test${NC}        - Run test suite"
    echo -e "  ${YELLOW}npm run build${NC}   - Build for production"
    exit 0
else
    print_status "ERROR" "Environment validation failed!"
    echo ""
    echo -e "${RED}❌ Some requirements are not met${NC}"
    echo -e "${YELLOW}Please fix the issues above before proceeding${NC}"
    echo ""
    echo -e "${BLUE}Recommended Actions:${NC}"
    echo -e "  ${YELLOW}./setup-test-environment.sh${NC} - Recreate environment from scratch"
    echo -e "  ${YELLOW}npm install${NC}                - Install dependencies"
    echo -e "  ${YELLOW}Check Node.js version${NC}      - Ensure >= 20.19.0"
    exit 1
fi
