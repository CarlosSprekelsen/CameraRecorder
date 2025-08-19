#!/bin/bash

# Run all integration tests with proper authentication
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
if node evidence/client-sprint-3/test-websocket-integration.js > /dev/null 2>&1; then
    echo "   ✅ PASSED"
else
    echo "   ❌ FAILED"
fi

# Test 2: Recording Operations (auth required)
echo "2️⃣ Testing Recording Operations..."
if node client/test-recording-operations.js > /dev/null 2>&1; then
    echo "   ✅ PASSED"
else
    echo "   ❌ FAILED"
fi

# Test 3: Take Snapshot (auth required)
echo "3️⃣ Testing Take Snapshot..."
if node client/test-take-snapshot.js > /dev/null 2>&1; then
    echo "   ✅ PASSED"
else
    echo "   ❌ FAILED"
fi

# Test 4: Notification Timing (auth required)
echo "4️⃣ Testing Notification Timing..."
if node client/test-notification-timing.js > /dev/null 2>&1; then
    echo "   ✅ PASSED"
else
    echo "   ❌ FAILED"
fi

# Test 5: Recording Simple (auth required)
echo "5️⃣ Testing Recording Simple..."
if node client/test-recording-simple.js > /dev/null 2>&1; then
    echo "   ✅ PASSED"
else
    echo "   ❌ FAILED"
fi

# Test 6: Auth Working (auth required)
echo "6️⃣ Testing Auth Working..."
if node client/test-auth-working.js > /dev/null 2>&1; then
    echo "   ✅ PASSED"
else
    echo "   ❌ FAILED"
fi

# Test 7: Recording Client (auth required)
echo "7️⃣ Testing Recording Client..."
if node client/test-recording-client.js > /dev/null 2>&1; then
    echo "   ✅ PASSED"
else
    echo "   ❌ FAILED"
fi

echo ""
echo "🎯 All tests completed!"
echo "💡 To see detailed output, run individual tests:"
echo "   source .test_env && node client/test-[name].js"
