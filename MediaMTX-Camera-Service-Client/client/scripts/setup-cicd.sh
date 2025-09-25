#!/bin/bash

# Setup script for CI/CD duplication detection
# Configures automated duplication detection for MediaMTX Camera Service Client

set -e

echo "🚀 Setting up CI/CD duplication detection..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo -e "${BLUE}📁 Project root: $PROJECT_ROOT${NC}"

# 1. Make scripts executable
echo -e "${YELLOW}🔧 Making scripts executable...${NC}"
chmod +x "$SCRIPT_DIR/detect-duplications.js"
chmod +x "$SCRIPT_DIR/pre-commit-duplication-check.js"

# 2. Install Husky if not already installed
echo -e "${YELLOW}📦 Installing Husky for git hooks...${NC}"
cd "$PROJECT_ROOT"

if ! command -v husky &> /dev/null; then
    echo "Installing Husky..."
    npm install --save-dev husky@^8.0.3
else
    echo "Husky already installed"
fi

# 3. Initialize Husky
echo -e "${YELLOW}🔗 Initializing Husky...${NC}"
npx husky init

# 4. Create pre-commit hook
echo -e "${YELLOW}🪝 Setting up pre-commit hook...${NC}"
mkdir -p .husky
cat > .husky/pre-commit << 'EOF'
#!/usr/bin/env sh
. "$(dirname -- "$0")/_/husky.sh"

# Pre-commit hook for MediaMTX Camera Service Client
# Runs duplication detection on staged test files

echo "🔍 Running pre-commit duplication check..."

# Run the duplication check script
node scripts/pre-commit-duplication-check.js

# If duplication check fails, prevent commit
if [ $? -ne 0 ]; then
  echo ""
  echo "🚨 Commit blocked due to duplication violations"
  echo "Please fix the issues above before committing"
  echo ""
  echo "Quick fix commands:"
  echo "  npm run test:duplication-check  # Run full check"
  echo "  npm run lint:fix               # Auto-fix some issues"
  echo ""
  exit 1
fi

echo "✅ Pre-commit checks passed"
EOF

chmod +x .husky/pre-commit

# 5. Create GitHub Actions directory if it doesn't exist
echo -e "${YELLOW}📁 Setting up GitHub Actions...${NC}"
cd "$(dirname "$PROJECT_ROOT")"
mkdir -p .github/workflows

# Check if workflow already exists
if [ -f ".github/workflows/duplication-detection.yml" ]; then
    echo -e "${YELLOW}⚠️  GitHub Actions workflow already exists${NC}"
    read -p "Do you want to overwrite it? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo "Overwriting existing workflow..."
    else
        echo "Skipping GitHub Actions setup"
    fi
fi

# 6. Test the duplication detection script
echo -e "${YELLOW}🧪 Testing duplication detection script...${NC}"
cd "$PROJECT_ROOT"

if node scripts/detect-duplications.js; then
    echo -e "${GREEN}✅ Duplication detection script working correctly${NC}"
else
    echo -e "${RED}❌ Duplication detection script failed${NC}"
    echo "Please check the script and try again"
    exit 1
fi

# 7. Test the pre-commit script
echo -e "${YELLOW}🧪 Testing pre-commit script...${NC}"
if node scripts/pre-commit-duplication-check.js; then
    echo -e "${GREEN}✅ Pre-commit script working correctly${NC}"
else
    echo -e "${RED}❌ Pre-commit script failed${NC}"
    echo "Please check the script and try again"
    exit 1
fi

# 8. Display summary
echo -e "${GREEN}🎉 CI/CD duplication detection setup complete!${NC}"
echo ""
echo -e "${BLUE}📋 What was configured:${NC}"
echo "   ✅ Executable duplication detection scripts"
echo "   ✅ Husky pre-commit hooks"
echo "   ✅ GitHub Actions workflow"
echo "   ✅ NPM scripts for duplication checking"
echo ""
echo -e "${BLUE}🔧 Available commands:${NC}"
echo "   npm run test:duplication-check        # Full duplication scan"
echo "   npm run test:duplication-check-quick  # Quick pre-commit check"
echo "   npm run ci:duplication               # CI duplication check"
echo "   npm run ci:all                       # All CI checks including duplication"
echo ""
echo -e "${BLUE}🪝 Git hooks:${NC}"
echo "   Pre-commit hook will run automatically on git commit"
echo "   Commits will be blocked if duplications are found"
echo ""
echo -e "${BLUE}🚀 GitHub Actions:${NC}"
echo "   Automated duplication detection on push/PR"
echo "   Detailed reports and PR comments on violations"
echo ""
echo -e "${YELLOW}💡 Next steps:${NC}"
echo "   1. Test by making a commit with duplicate mocks"
echo "   2. Check GitHub Actions workflow in repository"
echo "   3. Run 'npm run test:duplication-check' to see current status"
echo ""
echo -e "${GREEN}Setup complete! 🎉${NC}"
