#!/bin/bash

# Script to set up test environment variables from camera service
# This reads the JWT secret from the service's .env file and exports it

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

echo "🎯 Test environment ready! You can now run:"
echo "   source .test_env && node test-sprint-3-day-9-integration.js"
echo ""
echo "Or run the test directly:"
echo "   CAMERA_SERVICE_JWT_SECRET=$JWT_SECRET node test-sprint-3-day-9-integration.js"
