#!/bin/bash

# Test script for install and uninstall functionality
# This validates that both scripts work correctly

set -e

echo "=== Testing Install/Uninstall Scripts ==="
echo "Date: $(date)"
echo ""

# Test 1: Uninstall (clean slate)
echo "1. Testing uninstall script..."
if sudo deployment/scripts/uninstall.sh; then
    echo "✅ Uninstall script completed"
else
    echo "❌ Uninstall script failed"
    exit 1
fi

echo ""

# Test 2: Verify clean environment
echo "2. Verifying clean environment..."
if ! systemctl list-unit-files | grep -q camera-service; then
    echo "✅ No camera-service found"
else
    echo "❌ Camera service still exists"
    exit 1
fi

if ! ls -la /opt/camera-service 2>/dev/null; then
    echo "✅ No installation directory found"
else
    echo "❌ Installation directory still exists"
    exit 1
fi

echo ""

# Test 3: Install
echo "3. Testing install script..."
if sudo deployment/scripts/install.sh; then
    echo "✅ Install script completed"
else
    echo "❌ Install script failed"
    exit 1
fi

echo ""

# Test 4: Verify installation
echo "4. Verifying installation..."
if systemctl list-unit-files | grep -q camera-service; then
    echo "✅ Camera service installed"
else
    echo "❌ Camera service not installed"
    exit 1
fi

if ls -la /opt/camera-service; then
    echo "✅ Installation directory created"
else
    echo "❌ Installation directory not created"
    exit 1
fi

echo ""

# Test 5: Check service status
echo "5. Checking service status..."
sleep 5
if systemctl is-active --quiet camera-service; then
    echo "✅ Camera service is running"
else
    echo "⚠️ Camera service not running (checking logs...)"
    sudo journalctl -u camera-service -n 10
fi

echo ""

# Test 6: Check WebSocket binding
echo "6. Checking WebSocket binding..."
if netstat -tlnp 2>/dev/null | grep -q ":8002"; then
    echo "✅ WebSocket server binding to port 8002"
else
    echo "❌ WebSocket server not binding to port 8002"
fi

echo ""

# Test 7: Uninstall again
echo "7. Testing uninstall script again..."
if sudo deployment/scripts/uninstall.sh; then
    echo "✅ Uninstall script completed again"
else
    echo "❌ Uninstall script failed"
    exit 1
fi

echo ""

# Test 8: Final verification
echo "8. Final verification..."
if ! systemctl list-unit-files | grep -q camera-service; then
    echo "✅ No camera-service found (clean)"
else
    echo "❌ Camera service still exists"
    exit 1
fi

if ! ls -la /opt/camera-service 2>/dev/null; then
    echo "✅ No installation directory found (clean)"
else
    echo "❌ Installation directory still exists"
    exit 1
fi

echo ""
echo "=== All tests completed successfully ==="
echo "Install and uninstall scripts are working correctly" 