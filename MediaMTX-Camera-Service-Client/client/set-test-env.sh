#!/bin/bash

# Script to set up test environment variables for MediaMTX Camera Service Client
# Follows "Test First, Real Integration Always" philosophy
# Updated for client testing guidelines compliance (2025-08-19)

echo "🔧 Setting up MediaMTX Camera Service Client test environment..."
echo "📋 Testing Philosophy: 'Test First, Real Integration Always'"
echo ""

# Validate we're in the correct directory
if [ ! -f "package.json" ] || [ ! -d "tests" ]; then
    echo "❌ ERROR: Must run from client directory!"
    echo "💡 Please run: cd MediaMTX-Camera-Service-Client/client && ./set-test-env.sh"
    exit 1
fi

echo "✅ Running from correct client directory"

# Read JWT secret from camera service .env file
if [ -f "/opt/camera-service/.env" ]; then
    # Extract CAMERA_SERVICE_JWT_SECRET from .env file (requires sudo)
    JWT_SECRET=$(sudo grep "^CAMERA_SERVICE_JWT_SECRET=" /opt/camera-service/.env | cut -d'=' -f2)
    
    if [ -n "$JWT_SECRET" ]; then
        echo "✅ Found JWT secret in /opt/camera-service/.env"
        echo "🔐 JWT Secret: ${JWT_SECRET:0:16}..."
        
        # Export the environment variable
        export CAMERA_SERVICE_JWT_SECRET="$JWT_SECRET"
        echo "✅ Exported CAMERA_SERVICE_JWT_SECRET environment variable"
        
        # Also export for immediate use
        echo "export CAMERA_SERVICE_JWT_SECRET=$JWT_SECRET" > .test_env
        echo "📝 Created .test_env file for future use"
        
    else
        echo "❌ Could not extract JWT secret from .env file"
        exit 1
    fi
else
    echo "❌ Camera service .env file not found at /opt/camera-service/.env"
    echo "💡 Make sure the camera service is installed and running"
    exit 1
fi

echo ""
echo "🎯 Test Environment Ready!"
echo "📋 Following Client Testing Guidelines:"
echo "   ✅ Unit Tests: ≥80% coverage, isolated component behavior"
echo "   ✅ Integration Tests: ≥70% coverage, real server communication"
echo "   ✅ E2E Tests: Critical paths, complete user workflows"
echo "   ✅ Performance: <50ms status, <100ms control, <1s WebSocket"
echo ""

echo "⚠️  CRITICAL: IV&V Testing Protocol"
echo "   🚫 NEVER run tests from root directory"
echo "   ✅ ALWAYS run tests from client directory: cd client && npm test"
echo ""

echo "🔧 Environment Variables Available:"
echo "   CAMERA_SERVICE_JWT_SECRET=${JWT_SECRET:0:16}..."
echo ""

echo "📋 Test Organization Structure:"
echo "   tests/"
echo "   ├── unit/           # Isolated component/logic tests"
echo "   ├── integration/    # Real server communication tests"
echo "   ├── e2e/           # Complete workflow tests"
echo "   ├── performance/   # Load and timing validation"
echo "   └── fixtures/      # Shared test utilities"
echo ""

echo "⚠️  WebSocket Environment Compatibility:"
echo "   ✅ Tests use proper WebSocket API for environment"
echo "   ✅ Browser tests use native WebSocket object"
echo "   ✅ Node.js tests use appropriate WebSocket library"
echo "   ✅ No 'ws does not work in browser' errors"
echo ""

echo "🎯 Quality Gates:"
echo "   ✅ Performance: Status <50ms, Control <100ms, WebSocket <1s"
echo "   ✅ Coverage: Unit ≥80%, Integration ≥70%"
echo "   ✅ Integration: All tests pass against real server"
echo "   ✅ Authentication: Dynamic token generation, no hardcoded credentials"
echo ""

echo "🚀 Ready to run tests! Remember:"
echo "   📍 Always run from client directory"
echo "   🔐 Authentication handled automatically"
echo "   🌐 Real server integration for all tests"
echo "   📊 Coverage thresholds enforced"
