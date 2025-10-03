#!/bin/bash

# Anti-peek enforcement script for E2E tests
# Ensures E2E tests are truly black-box and don't access internal packages

set -e

echo "🔍 Running anti-peek enforcement checks..."

# Check for internal package imports
echo "Checking for internal package imports..."
INTERNAL_IMPORTS=$(grep -r "internal/" test/e2e/ || true)
if [ -n "$INTERNAL_IMPORTS" ]; then
    echo "❌ FAIL: Found internal package imports in E2E tests:"
    echo "$INTERNAL_IMPORTS"
    exit 1
fi

# Check for non-HTTP server symbols
echo "Checking for non-HTTP server symbols..."
SERVER_SYMBOLS=$(grep -r "\.Mux()\|\.Handler()\|server\.\*" test/e2e/ || true)
if [ -n "$SERVER_SYMBOLS" ]; then
    echo "❌ FAIL: Found non-HTTP server symbols in E2E tests:"
    echo "$SERVER_SYMBOLS"
    exit 1
fi

# Check for direct access to internal components
echo "Checking for direct access to internal components..."
INTERNAL_COMPONENTS=$(grep -r "radio\.Manager\|telemetry\.Hub\|command\.Orchestrator" test/e2e/ || true)
if [ -n "$INTERNAL_COMPONENTS" ]; then
    echo "❌ FAIL: Found direct access to internal components in E2E tests:"
    echo "$INTERNAL_COMPONENTS"
    exit 1
fi

# Check for concrete adapter types
echo "Checking for concrete adapter types..."
ADAPTER_TYPES=$(grep -r "silvusmock\.\|adapter\." test/e2e/ || true)
if [ -n "$ADAPTER_TYPES" ]; then
    echo "❌ FAIL: Found concrete adapter types in E2E tests:"
    echo "$ADAPTER_TYPES"
    exit 1
fi

# Check for allowed imports only
echo "Checking for allowed imports only..."
ALLOWED_IMPORTS="net/http|net/http/httptest|encoding/json|testing|time|context|strings|os|path/filepath|github.com/radio-control/rcc/test/harness"
FORBIDDEN_IMPORTS=$(grep -r "github.com" test/e2e/ | grep -v -E "($ALLOWED_IMPORTS)" || true)
if [ -n "$FORBIDDEN_IMPORTS" ]; then
    echo "❌ FAIL: Found forbidden imports in E2E tests:"
    echo "$FORBIDDEN_IMPORTS"
    exit 1
fi

echo "✅ PASS: All anti-peek checks passed"
echo "E2E tests are properly black-box and spec-driven"
