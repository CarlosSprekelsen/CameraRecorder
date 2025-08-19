# MediaMTX Camera Service Client

A React-based client application for interacting with the MediaMTX Camera Service.

## Prerequisites

### Node.js Version Requirements
- **Node.js**: >= 20.19.0 (LTS Iron release)
- **npm**: >= 10.8.0

The project uses React 19.1.1 and requires a modern Node.js environment for optimal compatibility.

### Installing Node.js
If you need to upgrade Node.js, we recommend using nvm (Node Version Manager):

```bash
# Install nvm (if not already installed)
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh | bash

# Restart your terminal or source bashrc
source ~/.bashrc

# Install and use Node.js 20.19.4
nvm install 20.19.4
nvm use 20.19.4
```

## Environment Validation

### Quick Environment Check
Before starting development, validate your environment:

```bash
chmod +x validate-environment.sh
./validate-environment.sh
```

This script performs comprehensive validation:
- ✅ Node.js version (>= 20.19.0)
- ✅ npm version (>= 10.8.0)
- ✅ nvm availability (recommended)
- ✅ Project structure
- ✅ Dependencies installation
- ✅ TypeScript compilation
- ✅ Test environment configuration
- ✅ React version (19.1.1)
- ✅ Build process
- ✅ Test execution

## Test Environment Setup

This project includes two complementary test environment scripts:

### 1. Test Infrastructure Setup (`setup-test-environment.sh`)
**Purpose**: Recreates the entire test infrastructure from scratch using proven configurations.

**When to use**:
- Initial development setup
- After dependency conflicts
- CI/CD environment setup
- When React Testing Library has issues

**Usage**:
```bash
./setup-test-environment.sh
```

**What it does**:
- Nuclear reset of test environment
- Installs proven compatible package versions
- Creates battle-tested Jest and TypeScript configurations
- Fixes React DOM compatibility issues
- Validates test environment

### 2. Integration Test Authentication (`set-test-env.sh`)
**Purpose**: Sets up JWT authentication environment variables for integration tests.

**When to use**:
- Before running integration tests
- Before running E2E tests
- Before running performance tests

**Usage**:
```bash
./set-test-env.sh
```

**What it does**:
- Reads JWT secret from camera service
- Exports authentication environment variables
- Creates `.test_env` file for test execution

## Development Workflow

1. **Initial Setup**:
   ```bash
   ./setup-test-environment.sh
   ```

2. **Before Running Integration Tests**:
   ```bash
   ./set-test-env.sh
   npm test
   ```

3. **Running Specific Test Types**:
   ```bash
   # Unit tests (no authentication needed)
   npm test -- --testPathPattern=unit
   
   # Integration tests (authentication required)
   ./set-test-env.sh
   npm test -- --testPathPattern=integration
   ```

## Available Scripts

- `npm test` - Run all tests
- `npm run test:watch` - Run tests in watch mode
- `npm run test:coverage` - Run tests with coverage
- `npm run lint` - Run linting
- `npm run build` - Build the project
- `npm run dev` - Start development server

## Test Organization

- **Unit Tests**: `tests/unit/` - Component and logic tests (no authentication needed)
- **Integration Tests**: `tests/integration/` - Real server communication tests (authentication required)
- **E2E Tests**: `tests/e2e/` - Complete workflow tests (authentication required)
- **Performance Tests**: `tests/performance/` - Load and timing tests (authentication required)

## Troubleshooting

### React Testing Library Issues
If you encounter React DOM compatibility issues:
```bash
./setup-test-environment.sh
```

### Authentication Issues
If integration tests fail with authentication errors:
```bash
./set-test-env.sh
```

### WebSocket Connection Issues
Integration tests may timeout due to WebSocket connection issues in Jest environment. This is a known limitation and doesn't affect unit test functionality.
