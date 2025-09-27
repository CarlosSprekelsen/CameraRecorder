# MediaMTX Camera Client - Testing Guide

## Quick Start

```bash
# 1. Setup environment
./set-test-env.sh

# 2. Run all tests
npm run test:all

# 3. Run specific test types
npm run test:unit          # Unit tests (jsdom + mocks)
npm run test:integration   # Integration tests (Node.js + real WebSocket)
npm run test:e2e          # End-to-end tests (Node.js + real hardware)
```

## Test Organization

### Directory Structure
```
tests/
├── unit/              # Isolated component/logic tests
├── integration/       # Real server communication tests  
├── e2e/              # Complete workflow tests
├── fixtures/         # Shared test utilities
├── utils/            # Centralized test utilities
└── config/           # Test configurations
```

### File Naming Convention
- **Standard**: `*.test.ts` (Jest automatic discovery)
- **Examples**: `camera_operations.test.ts`, `websocket_integration.test.ts`
- **No variations**: No `_real`, `_v2`, `_mock` suffixes

## Test Types

### Unit Tests
- **Environment**: jsdom (browser simulation)
- **Mocking**: External dependencies mocked
- **Coverage**: 80%+ threshold
- **Execution**: `npm run test:unit`

### Integration Tests  
- **Environment**: Node.js
- **Communication**: Real WebSocket connections
- **Coverage**: 60%+ threshold
- **Execution**: `npm run test:integration`

### End-to-End Tests
- **Environment**: Node.js
- **Hardware**: Real camera devices
- **Coverage**: No coverage (focus on user journeys)
- **Execution**: `npm run test:e2e`

## Configuration

### Jest Configuration Architecture
- **Base Config**: `jest.config.base.cjs` (shared settings)
- **Unit Config**: `tests/config/jest-unit.config.cjs`
- **Integration Config**: `tests/integration/jest.config.cjs`
- **E2E Config**: `tests/config/jest-e2e.config.cjs`

### Environment Variables
```bash
# Required for integration and E2E tests
MEDIA_SERVER_URL=ws://localhost:8002/ws
HEALTH_SERVER_URL=http://localhost:8003
JWT_SECRET=your-secret-key
```

## Development Workflow

### Daily Development
```bash
# Start development session
./set-test-env.sh
npm run test:unit -- --watch

# Before committing
npm run test:unit
npm run test:integration
npm run lint
```

### Feature Development (TDD)
```bash
# 1. Write test first
npm run test:unit -- --testPathPattern="feature_name"

# 2. Implement feature
npm run test:unit -- --watch

# 3. Integration testing
npm run test:integration -- --testPathPattern="feature_name"

# 4. Final validation
npm run test:all && npm run build
```

## IDE Integration

### VS Code Setup
1. Install Jest extension
2. Configure settings in `.vscode/settings.json`
3. Enable test discovery and coverage
4. Set up auto-run in watch mode

### Test Discovery
- Jest extension automatically discovers `*.test.ts` files
- Test explorer shows all tests in sidebar
- Click to run individual tests or test suites

## Troubleshooting

### Common Issues

#### Test Failures
```bash
# Check server status
curl localhost:8002/ws

# Refresh environment
./set-test-env.sh
```

#### Authentication Errors
```bash
# Regenerate tokens
./set-test-env.sh

# Check server connectivity
npm run test:integration
```

#### Memory Issues
```bash
# Increase Node.js memory
NODE_OPTIONS="--max-old-space-size=4096" npm run test:unit
```

#### Coverage Failures
```bash
# Generate coverage report
npm run test:unit:coverage

# Check specific files
npm run test:unit -- --coverage --testPathPattern="specific_file"
```

### Performance Optimization
```bash
# Parallel testing
npm run test:unit -- --maxWorkers=4

# Test caching
npm run test:unit -- --cache

# Selective testing
npm run test:unit -- --testPathPattern="specific_pattern"
```

## Quality Gates

### Pre-merge Checklist
- [ ] Unit tests pass (80%+ coverage)
- [ ] Integration tests pass (60%+ coverage)
- [ ] Lint passes without warnings
- [ ] Build succeeds
- [ ] No test file naming violations

### CI/CD Pipeline
- **GitHub Actions**: Automated testing on push/PR
- **Quality Gates**: Block merge if tests fail
- **Coverage Reports**: Track coverage trends
- **Artifact Storage**: Build artifacts for deployment

## Best Practices

### Test Writing
- **AAA Pattern**: Arrange, Act, Assert
- **Descriptive Names**: Explain the scenario being tested
- **Test Isolation**: Each test should be independent
- **Mock Strategy**: Mock external dependencies only
- **Specific Assertions**: Avoid generic `toBeTruthy()`

### Code Coverage
- **Unit Tests**: Focus on business logic
- **Integration Tests**: Focus on API interactions
- **E2E Tests**: Focus on user journeys
- **Coverage Reports**: Review before merging

### Performance
- **Parallel Execution**: Use `--maxWorkers` for speed
- **Test Caching**: Enable Jest cache
- **Resource Cleanup**: Proper cleanup in teardown
- **Timeout Management**: Appropriate timeouts per test type

## Advanced Features

### Custom Matchers
```typescript
// tests/utils/custom-matchers.ts
expect.extend({
  toBeValidJWT(received: string) {
    // Custom JWT validation matcher
  }
});
```

### Test Data Management
```typescript
// tests/fixtures/test-data.ts
export const TEST_CAMERAS = {
  camera0: { device: 'camera0', status: 'CONNECTED' },
  camera1: { device: 'camera1', status: 'DISCONNECTED' }
};
```

### Environment-Specific Testing
```bash
# Development
NODE_ENV=development npm run test:unit

# Staging
NODE_ENV=staging npm run test:integration

# Production
NODE_ENV=production npm run test:e2e
```

## Support

### Documentation
- **Testing Guidelines**: `docs/development/client-testing-guidelines.md`
- **Architecture**: `docs/architecture/client-architechture.md`
- **API Reference**: `mediamtx-camera-service-go/docs/api/json_rpc_methods.md`

### Team Coordination
- **Standardized Workflow**: Consistent testing across developers
- **Quality Gates**: Automated quality enforcement
- **Cross-Platform**: Works on Windows, macOS, Linux
- **IDE Integration**: VS Code Jest extension support

---

**Last Updated**: January 2025  
**Version**: 1.0  
**Maintainer**: Development Team