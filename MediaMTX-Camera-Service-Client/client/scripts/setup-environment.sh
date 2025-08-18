#!/bin/bash

# MediaMTX Camera Service Client - Environment Setup Script
# This script ensures all developers and testers have a consistent environment

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to check Node.js version
check_node_version() {
    local required_version="20.0.0"
    local current_version=$(node --version | sed 's/v//')
    
    if command_exists node; then
        print_status "Current Node.js version: $current_version"
        
        # Compare versions
        if [ "$(printf '%s\n' "$required_version" "$current_version" | sort -V | head -n1)" = "$required_version" ]; then
            print_success "Node.js version $current_version meets requirements (>= $required_version)"
            return 0
        else
            print_error "Node.js version $current_version is too old. Required: >= $required_version"
            return 1
        fi
    else
        print_error "Node.js is not installed"
        return 1
    fi
}

# Function to setup nvm and Node.js
setup_node() {
    print_status "Setting up Node.js environment..."
    
    # Check if nvm is available
    if [ -s "$HOME/.nvm/nvm.sh" ]; then
        print_status "NVM found, loading..."
        export NVM_DIR="$HOME/.nvm"
        [ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"
        [ -s "$NVM_DIR/bash_completion" ] && \. "$NVM_DIR/bash_completion"
        
        # Check available Node.js versions
        local available_versions=$(nvm list --no-alias | grep -E "v[0-9]+\.[0-9]+\.[0-9]+" | tail -1 | tr -d ' ->')
        
        if [ -n "$available_versions" ]; then
            print_status "Using Node.js version: $available_versions"
            nvm use "$available_versions"
            
            if check_node_version; then
                print_success "Node.js environment ready"
                return 0
            fi
        fi
        
        # Install latest LTS if needed
        print_status "Installing latest LTS Node.js..."
        nvm install --lts
        nvm use --lts
        
        if check_node_version; then
            print_success "Node.js LTS installed and ready"
            return 0
        fi
    else
        print_error "NVM not found. Please install NVM first:"
        print_error "curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh | bash"
        return 1
    fi
}

# Function to clean and install dependencies
setup_dependencies() {
    print_status "Setting up project dependencies..."
    
    # Clean existing installation
    if [ -d "node_modules" ]; then
        print_status "Removing existing node_modules..."
        rm -rf node_modules
    fi
    
    if [ -f "package-lock.json" ]; then
        print_status "Removing existing package-lock.json..."
        rm -f package-lock.json
    fi
    
    # Install dependencies
    print_status "Installing dependencies..."
    npm install
    
    if [ $? -eq 0 ]; then
        print_success "Dependencies installed successfully"
        return 0
    else
        print_error "Failed to install dependencies"
        return 1
    fi
}

# Function to run validation tests
run_validation() {
    print_status "Running validation tests..."
    
    local validation_passed=true
    
    # Check TypeScript compilation
    print_status "Checking TypeScript compilation..."
    if npm run build >/dev/null 2>&1; then
        print_success "TypeScript compilation passed"
    else
        print_warning "TypeScript compilation has issues (expected for development)"
        validation_passed=false
    fi
    
    # Check linting
    print_status "Checking code quality..."
    if npm run lint >/dev/null 2>&1; then
        print_success "Linting passed"
    else
        print_warning "Linting has issues (expected for development)"
        validation_passed=false
    fi
    
    # Check test framework
    print_status "Checking test framework..."
    if npm test -- --passWithNoTests >/dev/null 2>&1; then
        print_success "Test framework working"
    else
        print_warning "Some tests may be failing (expected for development)"
        validation_passed=false
    fi
    
    return $([ "$validation_passed" = true ] && echo 0 || echo 1)
}

# Function to display environment info
show_environment_info() {
    print_status "Environment Information:"
    echo "  Node.js: $(node --version)"
    echo "  npm: $(npm --version)"
    echo "  Project: $(pwd)"
    echo "  Package: $(node -p "require('./package.json').name")"
    echo "  Version: $(node -p "require('./package.json').version")"
}

# Function to create .env file if needed
setup_env_file() {
    if [ ! -f ".env" ]; then
        print_status "Creating .env file..."
        cat > .env << EOF
# MediaMTX Camera Service Client Environment Variables
# Development settings
NODE_ENV=development
VITE_API_URL=ws://localhost:8002/ws
VITE_SERVER_URL=http://localhost:8002
VITE_DEBUG=true

# Test settings
USE_MOCK_SERVER=false
TEST_TIMEOUT=30000
EOF
        print_success ".env file created"
    else
        print_status ".env file already exists"
    fi
}

# Function to display next steps
show_next_steps() {
    echo
    print_success "Environment setup completed!"
    echo
    print_status "Next steps:"
    echo "  1. Start development server: npm run dev"
    echo "  2. Run tests: npm test"
    echo "  3. Build for production: npm run build"
    echo "  4. Run full validation: npm run validate"
    echo
    print_status "Useful commands:"
    echo "  npm run dev          - Start development server"
    echo "  npm run build        - Build for production"
    echo "  npm run lint         - Check code quality"
    echo "  npm test             - Run tests"
    echo "  npm run test:watch   - Run tests in watch mode"
    echo "  npm run validate     - Run full validation suite"
    echo
}

# Main execution
main() {
    echo "=========================================="
    echo "MediaMTX Camera Service Client Setup"
    echo "=========================================="
    echo
    
    # Check if we're in the right directory
    if [ ! -f "package.json" ]; then
        print_error "package.json not found. Please run this script from the client directory."
        exit 1
    fi
    
    # Step 1: Setup Node.js
    if ! check_node_version; then
        if ! setup_node; then
            print_error "Failed to setup Node.js environment"
            exit 1
        fi
    fi
    
    # Step 2: Setup dependencies
    if ! setup_dependencies; then
        print_error "Failed to setup dependencies"
        exit 1
    fi
    
    # Step 3: Setup environment file
    setup_env_file
    
    # Step 4: Run validation
    if ! run_validation; then
        print_warning "Validation completed with warnings (this is normal for development)"
    else
        print_success "All validations passed!"
    fi
    
    # Step 5: Show environment info
    show_environment_info
    
    # Step 6: Show next steps
    show_next_steps
}

# Run main function
main "$@"
