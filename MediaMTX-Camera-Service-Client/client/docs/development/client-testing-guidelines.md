# Test Architecture Guide - Web Client

**Version:** 1.0  
**Date:** 2025-09-25  
**Status:** Active - Follow architecture principles from go-architecture-guide.md  

## Core Principles

### Ground Truth Validation
- **API Documentation is FROZEN** - Only documented API as reference
- **Architecture is AUTHORITATIVE** - Only architecture docs as reference  
- **Tests validate against ground truth** - Not existing code
- **Test failures indicate real problems** - Not test problems
- **No "accommodation" of broken implementations** - Tests do not fix implementation
- **STOP and ask for authorization** before making any test changes

### Critical Rules
1. **NEVER look at server implementation code** - Only use API documentation
2. **NEVER look at client implementation code** - Only use architecture documentation  
3. **If test fails, check ground truth first** - Don't adapt test to broken implementation
4. **Test failures are real bugs** - Not test bugs to be fixed
5. **Purpose of testing is validation** - Not to pass by any means

### DRY & Single Responsibility  
- One test utility per concern - no duplicate mock implementations
- Shared validation patterns - centralized response validation
- Consistent mocking strategy - single approach across test types
- Component reuse - leverage existing infrastructure

## Environment Setup - CRITICAL ⚠️

### ALWAYS Required Before Testing
```bash
# REQUIRED: Load environment before any test execution
cd MediaMTX-Camera-Service-Client/client
./set-test-env.sh
source .test_env
npm test
```

### Authentication Setup
- Tests accessing protected methods require valid authentication tokens
- JWT secret and server URL must be loaded before test execution
- `set-test-env.sh` ensures current valid keys for your server instance
- **This is the #1 cause of false test failures**

### Server Port Configuration
**MediaMTX Camera Service has TWO endpoints - use correct ports:**
- **WebSocket Server (JSON-RPC)**: Port 8002 - All camera operations, file management
- **Health Server (REST)**: Port 8003 - Health checks, system status
- **MANDATORY**: Do not mix WebSocket methods with health endpoints

```bash
# WebSocket operations (camera control, file management)
ws://localhost:8002/ws

# Health operations (system status, monitoring)  
http://localhost:8003/health/*
```

### IV&V Testing Protocol
**ALWAYS run tests from `client/` directory**: `cd client && npm test`  
**NEVER run from root** - conflicting dependencies cause failures

## Test Organization

### Directory Structure
```
client/tests/
├── unit/              # Isolated component/logic tests
├── integration/       # Real server communication tests  
├── e2e/              # Complete workflow tests
├── fixtures/         # Shared test utilities
├── utils/            # Centralized test utilities
└── config/           # Test configurations
```

### Naming Convention - STANDARDIZED
- **Files**: `*.test.{ts,js}` using standard Jest convention
- **Functions**: `test_<behavior>_<scenario>()` or `describe('<feature>', () => { it('should <behavior>', () => {}) })`
- **Examples**: `camera_operations.test.ts`, `websocket_integration.test.ts`, `auth_flow_e2e.test.ts`
- **No variations**: No _real, _v2, _mock suffixes
- **Jest Standard**: Follows Jest's `*.test.ts` pattern for automatic discovery

## Mocking Strategy

### Single Mock Pattern
- One mock implementation per API concern
- Centralized in `utils/mocks.ts`
- Based on documented API responses
- No duplicate mock patterns across tests

### WebSocket Mocking
- Environment-driven: real connections for integration, mocks for unit
- Single WebSocket abstraction in `utils/api-client.ts`
- Toggle real/mock via configuration, not separate implementations

## Shared Utilities

### API Response Validation
- Centralized validator in `utils/validators.ts`
- Validates against documented schemas
- Single validation pattern across all test types
- No per-test validation logic

### Authentication
- Dynamic token generation in `utils/auth-helper.ts`
- No hardcoded credentials
- Token caching for performance
- Single auth pattern for all test categories

## Test Categories

### Unit Tests (≥80% coverage)
- Component behavior in isolation  
- Business logic validation
- Edge case handling
- Use shared mock utilities
- Mocks permitted only for external APIs beyond project control

### Integration Tests (≥70% coverage)
- Real API communication via shared client
- JSON-RPC method validation against documented API
- Authentication flows with dynamic token generation (following documented flow)
- **Authentication tokens are always generated dynamically; no hardcoded credentials allowed**
- Error handling for network/server failures (using documented error codes)

### End-to-End Tests (Critical paths)
- Complete user workflows in real browser
- Camera operations with actual hardware/simulation  
- File management end-to-end
- Performance validation under load

### API Compliance Testing - MANDATORY
**Every test that calls server APIs MUST validate against API documentation, not implementation**

#### Mandatory API Compliance Rules
1. **Test against documented API format** - Use exact request/response formats from API documentation
2. **Validate documented error codes** - Use error codes and messages from API documentation  
3. **Test documented authentication flow** - Follow authentication flow exactly as documented
4. **Verify documented response fields** - Check all required fields are present and correct
5. **No implementation-specific testing** - Don't test server internals, only documented behavior

## Quality Gates

### Coverage Enforcement
- Automated threshold validation
- Build failure on regression
- Focus on business logic paths
- Component-specific thresholds

### Performance Targets
- Status methods: <50ms (p95)
- Control methods: <100ms (p95)
- WebSocket connection: <1s (p95)
- Client load: <3s (p95)

## Test Configuration - STANDARDIZED

### Jest Configuration Architecture
```javascript
// jest.config.base.cjs - Base configuration (shared settings)
const baseConfig = {
  transform: { /* shared transform config */ },
  moduleNameMapper: { /* shared module mapping */ },
  collectCoverageFrom: [ /* shared coverage patterns */ ]
};

// jest-unit.config.cjs - Unit tests (jsdom + mocks)
module.exports = {
  ...baseConfig,
  testEnvironment: 'jsdom',
  testMatch: ['**/tests/unit/**/*.test.{ts,js}'],
  setupFilesAfterEnv: ['<rootDir>/tests/setup.ts']
};

// jest.config.cjs - Integration tests (Node.js + real WebSocket)
module.exports = {
  ...baseConfig,
  testEnvironment: 'node', 
  testMatch: ['**/tests/integration/**/*.test.{ts,js}'],
  setupFilesAfterEnv: ['<rootDir>/tests/integration/setup.ts']
};

// jest-e2e.config.cjs - E2E tests (Node.js + real hardware)
module.exports = {
  ...baseConfig,
  testEnvironment: 'node',
  testMatch: ['**/tests/e2e/**/*.test.{ts,js}'],
  setupFilesAfterEnv: ['<rootDir>/tests/setup.integration.ts']
};
```

### Test Execution Requirements - UPDATED
- Unit tests: `npm run test:unit` (jsdom + mocks)
- Integration tests: `npm run test:integration` (Node.js + real WebSocket)
- E2E tests: `npm run test:e2e` (Node.js + real hardware)
- All tests: `npm run test:all` (unit + integration + e2e)

### WebSocket Testing
- Unit tests: Mock WebSocket services completely
- Integration tests: Use `require('ws')` for Node.js WebSocket
- Browser tests: Use native WebSocket API

## Anti-Patterns

### Avoid Multiple Mock Implementations
- No duplicate mock patterns
- No test-specific mock creation
- No parallel implementations

### Avoid Implementation Dependencies
- No testing private methods
- No testing internal state
- No hardcoded test data
- No mocking what you don't own

### Avoid Configuration Duplication
- Single source for test configuration
- Shared utilities across test types
- Consistent patterns across components

## Requirements Traceability - MANDATORY

### Mandatory Test Headers
```typescript
/**
 * Module description.
 * 
 * Ground Truth References:
 * - Server API: ../mediamtx-camera-service/docs/api/json-rpc-methods.md
 * - Client Architecture: ../docs/architecture/client-architecture.md
 * - Client Requirements: ../docs/requirements/client-requirements.md
 * 
 * Requirements Coverage:
 * - REQ-XXX-001: Requirement description
 * - REQ-XXX-002: Additional requirement
 * 
 * Test Categories: Unit/Integration/Security/Performance/Health
 * API Documentation Reference: docs/api/json-rpc-methods.md
 */
describe('Feature Tests', () => {
  test('REQ-XXX-001: Specific requirement validation', () => {
    // Test that would FAIL if requirement violated
    // Test against ground truth, not existing implementation
  });
});
```

### Requirements Coverage Rules
- Every test file MUST include REQ-* header
- REQ-* IDs MUST map to actual project requirements
- No test without requirements traceability
- Link tests to frozen requirements baseline
- Document in `requirements-coverage.md`

## Implementation Rules

### Before Writing Tests
- Check existing utilities for reusable patterns
- Validate against API documentation
- Plan shared mock strategy
- Identify architecture integration points

### During Implementation
- Use established patterns from utils/
- Follow naming conventions
- Add requirements traceability
- Validate against documented APIs

### After Implementation
- Verify no duplicate patterns
- Check coverage thresholds
- Test against real endpoints
- Document new shared utilities

## Multi-Developer Coordination

### Development Workflow - ENHANCED
1. **Environment Setup**: Always run `./set-test-env.sh` before any testing
2. **Test Execution**: Use standardized npm scripts (`test:unit`, `test:integration`, `test:e2e`)
3. **File Naming**: Follow `*.test.ts` convention for automatic Jest discovery
4. **Configuration**: Use centralized base config to prevent drift
5. **Code Reviews**: Verify test coverage and naming compliance
6. **CI/CD Integration**: Ensure all test types pass in pipeline

### Environment Synchronization - COMPREHENSIVE
- **Server Configuration**: Use `set-test-env.sh` to load current server settings
- **Test Environment**: All tests use `.test_env` file for consistent configuration
- **Port Configuration**: WebSocket (8002) vs Health (8003) endpoints clearly documented
- **Authentication**: Dynamic token generation prevents hardcoded credentials
- **Cross-Platform**: Guidelines work on Windows, macOS, and Linux
- **IDE Integration**: VS Code Jest extension configuration provided

### Team Coordination Rules - ENFORCED
- **No Custom Configs**: Use standardized Jest configurations only
- **Shared Utilities**: Leverage `tests/utils/` for common patterns
- **Consistent Patterns**: Follow established mocking and validation patterns
- **Documentation**: Update guidelines when adding new test patterns
- **Branch Protection**: Require test passes before merge
- **Code Coverage**: Maintain minimum thresholds per test type

### Developer Onboarding Checklist
- [ ] Clone repository and run `npm install`
- [ ] Execute `./set-test-env.sh` to configure environment
- [ ] Run `npm run test:unit` to verify setup
- [ ] Run `npm run test:integration` to test server connectivity
- [ ] Review this document and architecture guidelines
- [ ] Set up IDE Jest extension for test discovery
- [ ] Configure pre-commit hooks for test validation

### IDE Configuration (VS Code)
```json
// .vscode/settings.json
{
  "jest.jestCommandLine": "npm run test:unit",
  "jest.autoRun": "watch",
  "jest.showCoverageOnLoad": true,
  "jest.testExplorer": {
    "enabled": true
  },
  "typescript.preferences.includePackageJsonAutoImports": "on"
}
```

### Pre-commit Hooks Setup
```bash
# Install husky for git hooks
npm install --save-dev husky lint-staged

# Configure pre-commit hooks
npx husky add .husky/pre-commit "npm run test:unit -- --passWithNoTests"
```

### Troubleshooting Common Issues - EXPANDED
- **Test Failures**: Check server status with `curl localhost:8002/ws`
- **Authentication Errors**: Run `./set-test-env.sh` to refresh tokens
- **Port Conflicts**: Verify WebSocket (8002) vs Health (8003) usage
- **Configuration Drift**: Use base config inheritance, don't create custom configs
- **Memory Issues**: Increase Node.js memory with `NODE_OPTIONS="--max-old-space-size=4096"`
- **Timeout Issues**: Check server performance and network connectivity
- **Coverage Failures**: Run `npm run test:unit:coverage` to identify gaps
- **Import Errors**: Verify TypeScript path mapping in base config

### Performance Optimization
- **Parallel Testing**: Use `--maxWorkers=4` for faster execution
- **Test Caching**: Enable Jest cache with `--cache` flag
- **Selective Testing**: Use `--testPathPattern` for focused testing
- **Coverage Optimization**: Use `--coverageThreshold` for build validation

### Quality Gates - ENFORCED
```bash
# Pre-merge validation
npm run lint && \
npm run test:unit && \
npm run test:integration && \
npm run build
```

### Cross-Platform Compatibility
- **Windows**: Use Git Bash or WSL for shell scripts
- **macOS**: Native support for all commands
- **Linux**: Full compatibility with all features
- **Docker**: Containerized testing environment available

## Testing Workflows - COMPREHENSIVE

### Daily Development Workflow
```bash
# 1. Start development session
./set-test-env.sh
npm run test:unit -- --watch

# 2. Before committing changes
npm run test:unit
npm run test:integration
npm run lint

# 3. Before pushing to remote
npm run test:all
npm run build
```

### Feature Development Workflow
```bash
# 1. Create feature branch
git checkout -b feature/new-camera-control

# 2. Write tests first (TDD approach)
npm run test:unit -- --testPathPattern="camera"

# 3. Implement feature
npm run test:unit -- --watch

# 4. Integration testing
npm run test:integration -- --testPathPattern="camera"

# 5. End-to-end validation
npm run test:e2e -- --testPathPattern="camera"

# 6. Final validation
npm run test:all && npm run build
```

### Bug Fix Workflow
```bash
# 1. Reproduce bug with test
npm run test:unit -- --testNamePattern="bug_description"

# 2. Fix implementation
npm run test:unit -- --watch

# 3. Verify fix
npm run test:integration
npm run test:e2e

# 4. Regression testing
npm run test:all
```

## Best Practices - ENHANCED

### Test Writing Guidelines
- **Test Naming**: Use descriptive names that explain the scenario
- **Test Structure**: Follow AAA pattern (Arrange, Act, Assert)
- **Test Isolation**: Each test should be independent and repeatable
- **Mock Strategy**: Mock external dependencies, not internal logic
- **Assertions**: Use specific assertions, avoid generic `toBeTruthy()`

### Code Coverage Strategy
- **Unit Tests**: Target 80%+ coverage for business logic
- **Integration Tests**: Target 60%+ coverage for API interactions
- **E2E Tests**: Focus on critical user journeys
- **Coverage Reports**: Review coverage reports before merging

### Performance Testing
- **Load Testing**: Use integration tests with sustained load
- **Memory Testing**: Monitor memory usage in long-running tests
- **Timeout Management**: Set appropriate timeouts per test type
- **Resource Cleanup**: Ensure proper cleanup in afterEach/afterAll

### Security Testing
- **Authentication**: Test all authentication scenarios
- **Authorization**: Verify role-based access control
- **Input Validation**: Test with malicious inputs
- **Token Management**: Test token expiration and refresh

## Continuous Integration - ENHANCED

### GitHub Actions Workflow
```yaml
# .github/workflows/test.yml
name: Test Suite
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
        with:
          node-version: '18'
      - run: npm ci
      - run: npm run test:unit
      - run: npm run test:integration
      - run: npm run test:e2e
      - run: npm run build
```

### Quality Metrics Dashboard
- **Test Coverage**: Track coverage trends over time
- **Test Duration**: Monitor test execution time
- **Flaky Tests**: Identify and fix unreliable tests
- **Failure Analysis**: Categorize and track test failures

## Advanced Configuration

### Environment-Specific Testing
```bash
# Development environment
NODE_ENV=development npm run test:unit

# Staging environment  
NODE_ENV=staging npm run test:integration

# Production environment
NODE_ENV=production npm run test:e2e
```

### Custom Jest Matchers
```typescript
// tests/utils/custom-matchers.ts
expect.extend({
  toBeValidJWT(received: string) {
    // Custom matcher for JWT validation
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

---

**Architecture Integration**: Follows go-architecture-guide.md principles adapted for client testing  
**Maintenance Focus**: Centralized patterns prevent server-side bloat and duplication issues  
**Team Coordination**: Standardized workflow ensures consistent testing across developers  
**Quality Assurance**: Comprehensive testing strategy ensures reliable software delivery