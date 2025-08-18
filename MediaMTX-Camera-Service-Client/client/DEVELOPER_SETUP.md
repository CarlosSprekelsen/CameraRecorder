# MediaMTX Camera Service Client - Developer Setup Guide

## Quick Start

### For New Developers
```bash
# Clone the repository and navigate to client directory
cd MediaMTX-Camera-Service-Client/client

# Run the automated setup script
./scripts/setup-environment.sh
```

### For Existing Developers
```bash
# Quick environment validation
./scripts/quick-validate.sh

# Or run the full setup again
./scripts/setup-environment.sh
```

## Environment Requirements

### Node.js
- **Required**: Node.js >= 20.0.0
- **Recommended**: Node.js 24.x LTS
- **Package Manager**: npm >= 10.0.0

### Operating System
- **Linux**: Ubuntu 20.04+, CentOS 8+, or equivalent
- **macOS**: 10.15+ (Catalina)
- **Windows**: Windows 10+ with WSL2 (recommended)

### Development Tools
- **Git**: Latest version
- **Code Editor**: VS Code (recommended) with TypeScript support
- **Terminal**: Bash or Zsh

## Manual Setup (Alternative)

### 1. Install Node.js

#### Using NVM (Recommended)
```bash
# Install NVM
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.0/install.sh | bash

# Reload shell configuration
source ~/.bashrc

# Install and use Node.js LTS
nvm install --lts
nvm use --lts
nvm alias default node
```

#### Using Package Manager
```bash
# Ubuntu/Debian
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt-get install -y nodejs

# macOS (using Homebrew)
brew install node@20

# Windows (using Chocolatey)
choco install nodejs
```

### 2. Verify Installation
```bash
node --version  # Should be >= 20.0.0
npm --version   # Should be >= 10.0.0
```

### 3. Install Dependencies
```bash
cd MediaMTX-Camera-Service-Client/client
npm install
```

### 4. Environment Configuration
```bash
# Create environment file
cat > .env << EOF
NODE_ENV=development
VITE_API_URL=ws://localhost:8002/ws
VITE_SERVER_URL=http://localhost:8002
VITE_DEBUG=true
USE_MOCK_SERVER=false
TEST_TIMEOUT=30000
EOF
```

## Development Workflow

### Starting Development
```bash
# Start development server
npm run dev

# The application will be available at http://localhost:5173
```

### Running Tests
```bash
# Run all tests
npm test

# Run tests in watch mode
npm run test:watch

# Run specific test suites
npm run test:unit
npm run test:integration

# Run tests with coverage
npm run test:coverage
```

### Code Quality
```bash
# Check code quality
npm run lint

# Fix auto-fixable issues
npm run lint -- --fix

# Format code
npx prettier --write .
```

### Building for Production
```bash
# Build the application
npm run build

# Preview production build
npm run preview
```

## Available Scripts

| Script | Description |
|--------|-------------|
| `npm run dev` | Start development server |
| `npm run build` | Build for production |
| `npm run preview` | Preview production build |
| `npm run lint` | Check code quality |
| `npm test` | Run all tests |
| `npm run test:watch` | Run tests in watch mode |
| `npm run test:coverage` | Run tests with coverage |
| `npm run validate` | Run full validation suite |
| `npm run setup` | Install dependencies and validate |

## Project Structure

```
client/
├── src/                    # Source code
│   ├── components/         # React components
│   ├── services/          # API and business logic
│   ├── stores/            # State management
│   ├── types/             # TypeScript definitions
│   └── utils/             # Utility functions
├── tests/                 # Test files
│   ├── unit/              # Unit tests
│   ├── integration/       # Integration tests
│   └── fixtures/          # Test fixtures
├── public/                # Static assets
├── scripts/               # Setup and utility scripts
├── package.json           # Dependencies and scripts
├── tsconfig.json          # TypeScript configuration
├── vite.config.ts         # Vite configuration
└── jest.config.js         # Jest configuration
```

## Technology Stack

### Core Technologies
- **React**: ^19.1.0 - UI framework
- **TypeScript**: ~5.8.3 - Type safety
- **Vite**: ^7.0.4 - Build tool and dev server
- **Material-UI**: ^7.3.0 - UI components
- **Zustand**: ^5.0.7 - State management
- **React Router**: ^7.7.1 - Routing

### Development Tools
- **ESLint**: ^9.32.0 - Code linting
- **Prettier**: ^3.6.2 - Code formatting
- **Jest**: ^30.0.5 - Testing framework
- **ts-jest**: ^29.4.1 - TypeScript testing

### PWA Support
- **Vite PWA Plugin**: ^1.0.2 - Progressive Web App features
- **Workbox**: ^7.3.0 - Service worker toolkit

## Troubleshooting

### Common Issues

#### Node.js Version Issues
```bash
# Check current version
node --version

# If version is too old, use NVM to switch
nvm use 24.6.0
```

#### Dependency Issues
```bash
# Clean install
rm -rf node_modules package-lock.json
npm install
```

#### Build Issues
```bash
# Check TypeScript errors
npx tsc --noEmit

# Check linting errors
npm run lint
```

#### Test Issues
```bash
# Run tests with verbose output
npm test -- --verbose

# Run specific test file
npm test -- CameraDetail.test.tsx
```

### Getting Help

1. **Check the logs**: Look for error messages in the terminal
2. **Validate environment**: Run `./scripts/quick-validate.sh`
3. **Reinstall dependencies**: Run `./scripts/setup-environment.sh`
4. **Check documentation**: Review this guide and project README
5. **Search issues**: Check existing GitHub issues for similar problems

## Development Guidelines

### Code Style
- Follow TypeScript strict mode
- Use ESLint and Prettier for code formatting
- Write meaningful commit messages
- Add tests for new features

### Git Workflow
```bash
# Create feature branch
git checkout -b feature/your-feature-name

# Make changes and commit
git add .
git commit -m "feat: add new feature"

# Push and create pull request
git push origin feature/your-feature-name
```

### Testing Strategy
- Write unit tests for components and services
- Write integration tests for API interactions
- Maintain >80% test coverage
- Use mock server for development

### Performance
- Monitor bundle size with `npm run build`
- Use React DevTools for performance analysis
- Optimize images and assets
- Implement code splitting where appropriate

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `NODE_ENV` | Environment mode | `development` |
| `VITE_API_URL` | WebSocket API URL | `ws://localhost:8002/ws` |
| `VITE_SERVER_URL` | HTTP API URL | `http://localhost:8002` |
| `VITE_DEBUG` | Enable debug mode | `true` |
| `USE_MOCK_SERVER` | Use mock server for tests | `false` |
| `TEST_TIMEOUT` | Test timeout in milliseconds | `30000` |

## Support

For additional support:
1. Check the project documentation
2. Review the test files for examples
3. Look at existing components for patterns
4. Ask questions in the project discussions

---

**Last Updated**: 2025-08-18
**Version**: 1.0
