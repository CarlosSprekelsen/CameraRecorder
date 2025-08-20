#!/bin/bash
# Run tests with proper environment setup.
#
# This script automatically sets up the test environment by reading the JWT secret
# from the service configuration and then runs the specified tests.
#
# Usage:
#     ./run_tests_with_env.sh [pytest_args...]
#
# Examples:
#     ./run_tests_with_env.sh tests/integration/test_authentication.py
#     ./run_tests_with_env.sh tests/integration/ -v
#     ./run_tests_with_env.sh tests/ -k "authentication" -v

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔧 Setting up test environment...${NC}"

# Run the setup script to ensure environment is current
python scripts/setup_test_environment.py

# Source the environment file
if [ -f .test_env ]; then
    source .test_env
    echo -e "${GREEN}✅ Test environment loaded${NC}"
else
    echo -e "${RED}❌ Test environment file not found${NC}"
    exit 1
fi

# Show the JWT secret (first 10 chars for verification)
if [ -n "$CAMERA_SERVICE_JWT_SECRET" ]; then
    echo -e "${GREEN}🔑 JWT Secret: ${CAMERA_SERVICE_JWT_SECRET:0:10}...${NC}"
else
    echo -e "${RED}❌ JWT Secret not set${NC}"
    exit 1
fi

echo -e "${BLUE}🚀 Running tests...${NC}"

# Run pytest with the provided arguments
if [ $# -eq 0 ]; then
    # Default: run all tests
    python -m pytest tests/
else
    # Run with provided arguments
    python -m pytest "$@"
fi

echo -e "${GREEN}✅ Tests completed!${NC}"
