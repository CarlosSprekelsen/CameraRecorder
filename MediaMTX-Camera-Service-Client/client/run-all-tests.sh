#!/bin/bash

# Run all integration tests with proper authentication
# Updated for new test organization structure (2025-08-19)
echo "🧪 Running all integration tests with authentication..."

# Set up environment
echo "🔧 Setting up test environment..."
./set-test-env.sh

# Source the environment
source .test_env

echo ""
echo "📋 Test Results Summary:"
echo "========================"

# Test 1: WebSocket Integration (no auth needed)
echo "1️⃣ Testing WebSocket Integration..."
cd tests/integration/websocket
if node test-websocket-integration.js > /dev/null 2>&1; then
    echo "   ✅ PASSED"
else
    echo "   ❌ FAILED"
fi
cd ../../..

# Test 2: Recording Operations (auth required)
echo "2️⃣ Testing Recording Operations..."
cd tests/integration/camera_ops
if node test-recording-operations.js > /dev/null 2>&1; then
    echo "   ✅ PASSED"
else
    echo "   ❌ FAILED"
fi
cd ../../..

# Test 3: Take Snapshot (auth required)
echo "3️⃣ Testing Take Snapshot..."
cd tests/integration/camera_ops
if node test-take-snapshot.js > /dev/null 2>&1; then
    echo "   ✅ PASSED"
else
    echo "   ❌ FAILED"
fi
cd ../../..

# Test 4: Notification Timing (auth required)
echo "4️⃣ Testing Notification Timing..."
cd tests/performance
if node test-notification-timing.js > /dev/null 2>&1; then
    echo "   ✅ PASSED"
else
    echo "   ❌ FAILED"
fi
cd ../..

# Test 5: Recording Simple (auth required)
echo "5️⃣ Testing Recording Simple..."
cd tests/integration/camera_ops
if node test-recording-simple.js > /dev/null 2>&1; then
    echo "   ✅ PASSED"
else
    echo "   ❌ FAILED"
fi
cd ../../..

# Test 6: Auth Working (auth required)
echo "6️⃣ Testing Auth Working..."
cd tests/integration/authentication
if node test-auth-working.js > /dev/null 2>&1; then
    echo "   ✅ PASSED"
else
    echo "   ❌ FAILED"
fi
cd ../../..

# Test 7: Recording Client (auth required)
echo "7️⃣ Testing Recording Client..."
cd tests/integration/camera_ops
if node test-recording-client.js > /dev/null 2>&1; then
    echo "   ✅ PASSED"
else
    echo "   ❌ FAILED"
fi
cd ../../..

# Test 8: File Download (auth required)
echo "8️⃣ Testing File Download..."
cd tests/integration
if node test-file-download.js > /dev/null 2>&1; then
    echo "   ✅ PASSED"
else
    echo "   ❌ FAILED"
fi
cd ..

# Test 9: Real-time Implementation (auth required)
echo "9️⃣ Testing Real-time Implementation..."
cd tests/integration
if node test-realtime-implementation.js > /dev/null 2>&1; then
    echo "   ✅ PASSED"
else
    echo "   ❌ FAILED"
fi
cd ..

# Test 10: Sprint Integration Tests (auth required)
echo "🔟 Testing Sprint Integration..."
cd tests/integration
if node test-sprint-3-day-9-integration.js > /dev/null 2>&1; then
    echo "   ✅ PASSED"
else
    echo "   ❌ FAILED"
fi
cd ..

echo ""
echo "🎯 All tests completed!"
echo ""
echo "💡 To see detailed output, run individual tests from their directories:"
echo "   cd tests/integration && source ../.test_env && node test-with-valid-token.js"
echo "   cd tests/integration/camera_ops && source ../../.test_env && node test-take-snapshot.js"
echo "   cd tests/integration/websocket && source ../../.test_env && node test-websocket-integration.js"
echo ""
echo "⚠️  Remember: Always run tests from their respective directories for proper execution context!"
