#!/bin/bash

# MediaMTX Camera Service Client Test Runner
# Ensures proper authentication setup before running tests

set -e

echo "🚀 MediaMTX Camera Service Client Test Runner"
echo "📋 Following 'Test First, Real Integration Always' philosophy"
echo ""

# Validate we're in the correct directory
if [ ! -f "package.json" ] || [ ! -d "tests" ]; then
    echo "❌ ERROR: Must run from client directory!"
    echo "💡 Please run: cd MediaMTX-Camera-Service-Client/client && ./run-tests.sh"
    exit 1
fi

echo "✅ Running from correct client directory"

# Set up authentication environment
echo "🔧 Setting up authentication environment..."
./set-test-env.sh
source .test_env

echo "✅ Authentication environment ready"
echo ""

# Parse command line arguments
TEST_TYPE=${1:-"all"}
EXTRA_ARGS=${@:2}

case $TEST_TYPE in
    "unit")
        echo "🧪 Running Unit Tests..."
        npm run test:unit $EXTRA_ARGS
        ;;
    "integration")
        echo "🔗 Running Integration Tests..."
        npm run test:integration $EXTRA_ARGS
        ;;
    "e2e")
        echo "🌐 Running E2E Tests..."
        npm run test:e2e $EXTRA_ARGS
        ;;
    "performance")
        echo "⚡ Running Performance Tests..."
        npm run test:performance $EXTRA_ARGS
        ;;
    "all")
        echo "🎯 Running All Tests..."
        npm run test:all $EXTRA_ARGS
        ;;
    *)
        echo "❌ Unknown test type: $TEST_TYPE"
        echo "💡 Usage: ./run-tests.sh [unit|integration|e2e|performance|all] [jest-args...]"
        echo "💡 Examples:"
        echo "   ./run-tests.sh unit"
        echo "   ./run-tests.sh integration --testNamePattern='ping'"
        echo "   ./run-tests.sh all --verbose"
        exit 1
        ;;
esac

echo ""
echo "✅ Test execution completed!"
echo "📊 Check test results above"
