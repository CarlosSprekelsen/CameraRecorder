#!/bin/bash

# Script to set up test environment variables from camera service
# This reads the JWT secret from the service's .env file and exports it
# Updated for new test organization structure (2025-08-19)

echo "🔧 Setting up test environment variables..."

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
    echo "💡 Make sure the camera service is installed"
    exit 1
fi

echo ""
echo "🎯 Test environment ready! You can now run comprehensive tests:"
echo ""
echo "📁 Comprehensive Camera Operations Test:"
echo "   cd tests/integration/camera_ops && source ../../.test_env && node test-camera-operations-comprehensive.js"
echo ""
echo "📁 Comprehensive Authentication Test:"
echo "   cd tests/integration/authentication && source ../../.test_env && node test-authentication-comprehensive.js"
echo ""
echo "📁 Basic WebSocket Integration Test:"
echo "   cd tests/integration/websocket && source ../../.test_env && node test-websocket.js"
echo ""
echo "📁 Sprint Integration Tests:"
echo "   cd tests/integration && source ../.test_env && node test-sprint-3-day-9-integration.js"
echo "   cd tests/integration && source ../.test_env && node test-sprint-3-integration.js"
echo ""
echo "📁 Performance Tests:"
echo "   cd tests/performance && source ../.test_env && node test-notification-timing.js"
echo "   cd tests/performance && source ../.test_env && node test-realtime-updates.js"
echo ""
echo "📁 E2E Tests:"
echo "   cd tests/e2e && source ../.test_env && node test-take-snapshot-e2e.cjs"
echo ""
echo "🔧 Or run with direct environment variable:"
echo "   CAMERA_SERVICE_JWT_SECRET=$JWT_SECRET node tests/integration/camera_ops/test-camera-operations-comprehensive.js"
echo ""
echo "⚠️  IMPORTANT: Always run integration tests from their respective directories!"
echo "   This ensures proper component path resolution and test execution context."
echo ""
echo "📋 Test Organization Summary:"
echo "   ✅ Removed duplicate/obsolete tests"
echo "   ✅ Created comprehensive test suites following server API specification"
echo "   ✅ All tests validate against real server implementation"
echo "   ✅ Proper authentication and error handling coverage"
