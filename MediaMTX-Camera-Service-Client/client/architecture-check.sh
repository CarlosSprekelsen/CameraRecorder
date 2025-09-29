#!/bin/bash

# Architecture Validation Script
# Run this before every commit to prevent architecture drift

echo "🏗️  Architecture Validation Check"
echo "================================"

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

ERRORS=0

# RULE 1: No services imported in components
echo -n "Checking components don't import services... "
if grep -r "from ['\"].*services/" src/components/ src/pages/ --include="*.tsx" --include="*.ts" | grep -v "LoggerService"; then
    echo -e "${RED}❌ FAILED${NC}"
    echo "Components/Pages are importing services directly!"
    echo "Fix: Use stores instead of services in components"
    ERRORS=$((ERRORS + 1))
else
    echo -e "${GREEN}✓ PASSED${NC}"
fi

# RULE 2: Services use APIClient, not WebSocketService
echo -n "Checking services use APIClient... "
if grep -r "WebSocketService" src/services/ --include="*.ts" | grep -v "APIClient.ts" | grep -v "test" | grep -v "mock" | grep -v "websocket/WebSocketService.ts"; then
    echo -e "${RED}❌ FAILED${NC}"
    echo "Services are using WebSocketService directly!"
    echo "Fix: Use APIClient instead of WebSocketService"
    ERRORS=$((ERRORS + 1))
else
    echo -e "${GREEN}✓ PASSED${NC}"
fi

# RULE 3: Check for consistent service constructors
echo -n "Checking service constructor patterns... "
if grep -r "constructor(" src/services/ --include="*Service.ts" | grep -v "apiClient: APIClient" | grep -v "APIClient.ts" | grep -v "test"; then
    echo -e "${YELLOW}⚠ WARNING${NC}"
    echo "Some services don't follow the standard constructor pattern"
    echo "Expected: constructor(private apiClient: APIClient, private logger: LoggerService)"
fi

# RULE 4: No stores importing components
echo -n "Checking stores don't import components... "
if grep -r "from ['\"].*components/" src/stores/ --include="*.ts"; then
    echo -e "${RED}❌ FAILED${NC}"
    echo "Stores are importing components (circular dependency)!"
    ERRORS=$((ERRORS + 1))
else
    echo -e "${GREEN}✓ PASSED${NC}"
fi

# RULE 5: Check for direct WebSocket.sendRPC calls
echo -n "Checking for direct WebSocket RPC calls... "
if grep -r "\.sendRPC(" src/ --include="*.ts" --include="*.tsx" | grep -v "APIClient.ts" | grep -v "WebSocketService.ts" | grep -v "test"; then
    echo -e "${RED}❌ FAILED${NC}"
    echo "Found direct WebSocket.sendRPC calls outside APIClient!"
    echo "Fix: Use APIClient.call() instead"
    ERRORS=$((ERRORS + 1))
else
    echo -e "${GREEN}✓ PASSED${NC}"
fi

# RULE 6: Check type definitions match server
echo -n "Checking API types alignment... "
# Check for common type mismatches
if grep -r "status: ['\"]completed['\"]" src/ --include="*.ts"; then
    echo -e "${YELLOW}⚠ WARNING${NC}"
    echo "Found 'completed' status - server uses 'SUCCESS'/'FAILED'"
fi

# RULE 7: No legacy patterns
echo -n "Checking for legacy patterns... "
LEGACY_FOUND=0
if grep -r "extends.*SessionInfo" src/ --include="*.ts"; then
    echo -e "${YELLOW}⚠ WARNING${NC}"
    echo "Found legacy interface extensions"
    LEGACY_FOUND=1
fi

if [ $LEGACY_FOUND -eq 0 ]; then
    echo -e "${GREEN}✓ PASSED${NC}"
fi

echo "================================"

if [ $ERRORS -gt 0 ]; then
    echo -e "${RED}Architecture validation FAILED with $ERRORS errors${NC}"
    echo "Fix the violations before committing!"
    exit 1
else
    echo -e "${GREEN}Architecture validation PASSED!${NC}"
fi

# Run TypeScript strict check
echo ""
echo "Running strict TypeScript check..."
npx tsc --noEmit --project tsconfig.strict.json
if [ $? -ne 0 ]; then
    echo -e "${RED}TypeScript strict mode failed${NC}"
    exit 1
fi

echo -e "${GREEN}All architecture checks passed!${NC}"
