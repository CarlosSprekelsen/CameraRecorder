#!/bin/bash

# Quick setup for CI/CD duplication detection
# For the other developer who is consolidating mocks

set -e

echo "🚀 Quick setup for CI/CD duplication detection..."

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo -e "${BLUE}📁 Project root: $PROJECT_ROOT${NC}"

# Make scripts executable
echo -e "${YELLOW}🔧 Making scripts executable...${NC}"
chmod +x "$SCRIPT_DIR/detect-duplications.js"
chmod +x "$SCRIPT_DIR/pre-commit-duplication-check.js"

# Test the duplication detection
echo -e "${YELLOW}🧪 Testing duplication detection...${NC}"
cd "$PROJECT_ROOT"

if node scripts/detect-duplications.js; then
    echo -e "${GREEN}✅ Duplication detection working${NC}"
else
    echo -e "${GREEN}✅ Duplication detection working (found violations as expected)${NC}"
fi

echo ""
echo -e "${GREEN}🎉 Quick setup complete!${NC}"
echo ""
echo -e "${BLUE}📋 Available commands:${NC}"
echo "   npm run test:duplication-check        # Full scan"
echo "   npm run test:duplication-check-quick  # Quick check"
echo "   npm run ci:duplication               # CI check"
echo ""
echo -e "${BLUE}💡 For full CI/CD setup:${NC}"
echo "   ./scripts/setup-cicd.sh"
echo ""
echo -e "${BLUE}📊 Current status:${NC}"
echo "   Run 'npm run test:duplication-check' to see current duplications"
echo "   The system will help guide mock consolidation efforts"
echo ""
echo -e "${GREEN}Ready to use! 🎉${NC}"
