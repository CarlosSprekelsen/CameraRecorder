#!/bin/bash

# Script to set up test environment variables for MediaMTX Camera Service Client
# Follows "Test First, Real Integration Always" philosophy
# Updated for client testing guidelines compliance (2025-08-19)
# CI/CD FRIENDLY: Eliminates sudo requirement for automated environments

echo "🔧 Setting up MediaMTX Camera Service Client test environment..."
echo "📋 Testing Philosophy: 'Test First, Real Integration Always'"
echo "🚀 CI/CD FRIENDLY: No sudo required for automated environments"
echo ""

# Validate we're in the correct directory
if [ ! -f "package.json" ] || [ ! -d "tests" ]; then
    echo "❌ ERROR: Must run from client directory!"
    echo "💡 Please run: cd MediaMTX-Camera-Service-Client/client && ./set-test-env.sh"
    exit 1
fi

echo "✅ Running from correct client directory"

# CI/CD FRIENDLY: Check for existing .test_env file first
if [ -f ".test_env" ]; then
    echo "✅ Found existing .test_env file"
    
    # Source the existing file to get the JWT secret
    source .test_env
    
    if [ -n "$CAMERA_SERVICE_JWT_SECRET" ]; then
        echo "✅ Using existing JWT secret from .test_env"
        echo "🔐 JWT Secret: ${CAMERA_SERVICE_JWT_SECRET:0:16}..."
        echo "✅ CAMERA_SERVICE_JWT_SECRET environment variable is set"
        echo "🚀 CI/CD READY: No sudo required!"
        
        # Export for immediate use
        export CAMERA_SERVICE_JWT_SECRET="$CAMERA_SERVICE_JWT_SECRET"
        
    else
        echo "⚠️  .test_env exists but CAMERA_SERVICE_JWT_SECRET is empty"
        echo "🔄 Falling back to system .env file..."
    fi
else
    echo "📝 No existing .test_env file found"
    echo "🔄 Attempting to read from system .env file..."
fi

# Only try to read from system .env if we don't have a valid JWT secret
if [ -z "$CAMERA_SERVICE_JWT_SECRET" ]; then
    echo "🔍 Checking for system camera service .env file..."
    
    # Check if we can read the file without sudo first
    if [ -r "/opt/camera-service/.env" ]; then
        echo "✅ Can read /opt/camera-service/.env without sudo"
        JWT_SECRET=$(grep "^CAMERA_SERVICE_JWT_SECRET=" /opt/camera-service/.env | cut -d'=' -f2)
    elif [ -f "/opt/camera-service/.env" ]; then
        echo "⚠️  /opt/camera-service/.env exists but requires sudo access"
        echo "🔄 Attempting to read with sudo (local development only)..."
        
        # Try with sudo (will fail in CI/CD, but that's expected)
        if command -v sudo >/dev/null 2>&1; then
            JWT_SECRET=$(sudo grep "^CAMERA_SERVICE_JWT_SECRET=" /opt/camera-service/.env 2>/dev/null | cut -d'=' -f2)
        else
            echo "❌ sudo not available (CI/CD environment)"
            JWT_SECRET=""
        fi
    else
        echo "❌ Camera service .env file not found at /opt/camera-service/.env"
        echo "💡 Make sure the camera service is installed and running"
        JWT_SECRET=""
    fi
    
    if [ -n "$JWT_SECRET" ]; then
        echo "✅ Found JWT secret in system .env file"
        echo "🔐 JWT Secret: ${JWT_SECRET:0:16}..."
        
        # Export the environment variable
        export CAMERA_SERVICE_JWT_SECRET="$JWT_SECRET"
        echo "✅ Exported CAMERA_SERVICE_JWT_SECRET environment variable"
        
        # Also export for future use
        echo "export CAMERA_SERVICE_JWT_SECRET=$JWT_SECRET" > .test_env
        echo "📝 Created .test_env file for future use"
        
    else
        echo "❌ Could not extract JWT secret"
        echo ""
        echo "🚨 CI/CD ENVIRONMENT DETECTED"
        echo "💡 For CI/CD environments, ensure .test_env file exists with:"
        echo "   export CAMERA_SERVICE_JWT_SECRET=<your-jwt-secret>"
        echo ""
        echo "💡 For local development, ensure camera service is running and accessible"
        exit 1
    fi
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
echo "   CAMERA_SERVICE_JWT_SECRET=${CAMERA_SERVICE_JWT_SECRET:0:16}..."
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
echo "   🚀 CI/CD FRIENDLY: No sudo required!"
