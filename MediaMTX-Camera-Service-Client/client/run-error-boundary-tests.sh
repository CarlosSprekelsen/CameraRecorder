#!/bin/bash

# Error Boundary Test Runner
# Runs comprehensive tests for Error Boundary components

set -e

echo "🧪 Error Boundary Test Suite Runner"
echo "=================================="

# Check if we're in the right directory
if [ ! -f "package.json" ]; then
    echo "❌ Error: Must run from client directory"
    echo "   cd client && ./run-error-boundary-tests.sh"
    exit 1
fi

# Check if Jest is available
if ! command -v jest &> /dev/null; then
    echo "❌ Error: Jest not found. Installing dependencies..."
    npm install
fi

echo "📋 Running Error Boundary Tests..."
echo ""

# Run Error Boundary tests with coverage
echo "🔍 Running FeatureErrorBoundary tests..."
npx jest tests/unit/components/ErrorBoundaries/FeatureErrorBoundary.test.tsx --config jest.error-boundary.config.cjs --verbose

echo ""
echo "🔍 Running ServiceErrorBoundary tests..."
npx jest tests/unit/components/ErrorBoundaries/ServiceErrorBoundary.test.tsx --config jest.error-boundary.config.cjs --verbose

echo ""
echo "🔍 Running ErrorBoundary tests..."
npx jest tests/unit/components/ErrorBoundaries/ErrorBoundary.test.tsx --config jest.error-boundary.config.cjs --verbose

echo ""
echo "🔍 Running Error Boundary Integration tests..."
npx jest tests/unit/components/ErrorBoundaries/ErrorBoundaryIntegration.test.tsx --config jest.error-boundary.config.cjs --verbose

echo ""
echo "📊 Running Error Boundary Coverage Analysis..."
npx jest tests/unit/components/ErrorBoundaries/ --config jest.error-boundary.config.cjs --coverage --coverageReporters=text --coverageReporters=html

echo ""
echo "✅ Error Boundary Test Suite Complete!"
echo ""
echo "📈 Coverage Report:"
echo "   - HTML Report: coverage/lcov-report/index.html"
echo "   - Text Report: See above output"
echo ""
echo "🎯 Test Results Summary:"
echo "   - FeatureErrorBoundary: Comprehensive unit tests"
echo "   - ServiceErrorBoundary: Service-specific error handling"
echo "   - ErrorBoundary: Basic error boundary functionality"
echo "   - Integration: Cross-boundary error propagation"
echo ""
echo "🚀 All Error Boundary tests completed successfully!"
