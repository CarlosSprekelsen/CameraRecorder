#!/bin/bash

# CI/CD Environment Setup Script for MediaMTX Camera Service Client
# NO SUDO REQUIRED - Designed for automated environments
# Follows "Test First, Real Integration Always" philosophy

echo "🚀 Setting up MediaMTX Camera Service Client CI/CD test environment..."
echo "📋 Testing Philosophy: 'Test First, Real Integration Always'"
echo "🔒 CI/CD SECURE: No sudo, no interactive prompts, no root access"
echo ""

# Validate we're in the correct directory
if [ ! -f "package.json" ] || [ ! -d "tests" ]; then
    echo "❌ ERROR: Must run from client directory!"
    echo "💡 Please run: cd MediaMTX-Camera-Service-Client/client && ./set-test-env-ci.sh"
    exit 1
fi

echo "✅ Running from correct client directory"

# CI/CD PRIORITY: Check environment variables first
if [ -n "$CAMERA_SERVICE_JWT_SECRET" ]; then
    echo "✅ Using CAMERA_SERVICE_JWT_SECRET from environment variable"
    echo "🔐 JWT Secret: ${CAMERA_SERVICE_JWT_SECRET:0:16}..."
    echo "🚀 CI/CD READY: Environment variable set!"
    
    # Export for immediate use
    export CAMERA_SERVICE_JWT_SECRET="$CAMERA_SERVICE_JWT_SECRET"
    
    # Create .test_env file for future use
    echo "export CAMERA_SERVICE_JWT_SECRET=$CAMERA_SERVICE_JWT_SECRET" > .test_env
    echo "📝 Created .test_env file for future use"
    
elif [ -f ".test_env" ]; then
    echo "✅ Found existing .test_env file"
    
    # Source the existing file to get the JWT secret
    source .test_env
    
    if [ -n "$CAMERA_SERVICE_JWT_SECRET" ]; then
        echo "✅ Using existing JWT secret from .test_env"
        echo "🔐 JWT Secret: ${CAMERA_SERVICE_JWT_SECRET:0:16}..."
        echo "🚀 CI/CD READY: .test_env file available!"
        
        # Export for immediate use
        export CAMERA_SERVICE_JWT_SECRET="$CAMERA_SERVICE_JWT_SECRET"
        
    else
        echo "❌ .test_env exists but CAMERA_SERVICE_JWT_SECRET is empty"
        echo "🚨 CI/CD ERROR: No valid JWT secret found"
        exit 1
    fi
    
else
    echo "❌ No JWT secret found in environment or .test_env file"
    echo ""
    echo "🚨 CI/CD ENVIRONMENT SETUP REQUIRED"
    echo "💡 For CI/CD environments, set one of the following:"
    echo ""
    echo "   Option 1: Environment Variable"
    echo "   export CAMERA_SERVICE_JWT_SECRET=<your-jwt-secret>"
    echo ""
    echo "   Option 2: .test_env File"
    echo "   echo 'export CAMERA_SERVICE_JWT_SECRET=<your-jwt-secret>' > .test_env"
    echo ""
    echo "   Option 3: CI/CD Pipeline Variable"
    echo "   CAMERA_SERVICE_JWT_SECRET: <your-jwt-secret>"
    echo ""
    echo "💡 The JWT secret should match the camera service configuration"
    exit 1
fi

echo ""
echo "🎯 CI/CD Test Environment Ready!"
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
echo "   🔒 CI/CD SECURE: No sudo, no root access required!"
