#!/bin/bash

# MediaMTX Camera Service Client - Quick Validation Script
# Quick check for developers to validate their environment

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[✓]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[⚠]${NC} $1"
}

print_error() {
    echo -e "${RED}[✗]${NC} $1"
}

echo "=========================================="
echo "MediaMTX Camera Service Client - Quick Validation"
echo "=========================================="
echo

# Check Node.js version
print_status "Checking Node.js version..."
if command -v node >/dev/null 2>&1; then
    NODE_VERSION=$(node --version | sed 's/v//')
    REQUIRED_VERSION="20.0.0"
    
    if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$NODE_VERSION" | sort -V | head -n1)" = "$REQUIRED_VERSION" ]; then
        print_success "Node.js $NODE_VERSION (>= $REQUIRED_VERSION)"
    else
        print_error "Node.js $NODE_VERSION (required >= $REQUIRED_VERSION)"
        exit 1
    fi
else
    print_error "Node.js not found"
    exit 1
fi

# Check npm version
print_status "Checking npm version..."
if command -v npm >/dev/null 2>&1; then
    NPM_VERSION=$(npm --version)
    print_success "npm $NPM_VERSION"
else
    print_error "npm not found"
    exit 1
fi

# Check if we're in the right directory
print_status "Checking project structure..."
if [ -f "package.json" ]; then
    print_success "package.json found"
else
    print_error "package.json not found. Run from client directory."
    exit 1
fi

# Check dependencies
print_status "Checking dependencies..."
if [ -d "node_modules" ]; then
    print_success "node_modules found"
else
    print_warning "node_modules not found. Run: npm install"
    exit 1
fi

# Quick TypeScript check
print_status "Checking TypeScript configuration..."
if [ -f "tsconfig.json" ]; then
    print_success "TypeScript config found"
else
    print_error "TypeScript config missing"
    exit 1
fi

# Quick build test (without full compilation)
print_status "Testing build system..."
if npm run build --dry-run >/dev/null 2>&1 || npm run build --help >/dev/null 2>&1; then
    print_success "Build system accessible"
else
    print_warning "Build system may have issues"
fi

# Quick test framework check
print_status "Testing test framework..."
if npm test -- --help >/dev/null 2>&1; then
    print_success "Test framework accessible"
else
    print_warning "Test framework may have issues"
fi

# Check environment file
print_status "Checking environment configuration..."
if [ -f ".env" ]; then
    print_success ".env file found"
else
    print_warning ".env file not found (will be created on first run)"
fi

echo
print_success "Quick validation completed!"
echo
print_status "Environment appears ready for development."
print_status "Run 'npm run dev' to start development server."
echo
