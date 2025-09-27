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

### Development Workflow
1. **Before Testing**: Always run `./set-test-env.sh` to ensure current server configuration
2. **Test Execution**: Use standardized npm scripts (`test:unit`, `test:integration`, `test:e2e`)
3. **File Naming**: Follow `*.test.ts` convention for automatic Jest discovery
4. **Configuration**: Use centralized base config to prevent drift

### Environment Synchronization
- **Server Configuration**: Use `set-test-env.sh` to load current server settings
- **Test Environment**: All tests use `.test_env` file for consistent configuration
- **Port Configuration**: WebSocket (8002) vs Health (8003) endpoints clearly documented
- **Authentication**: Dynamic token generation prevents hardcoded credentials

### Team Coordination Rules
- **No Custom Configs**: Use standardized Jest configurations only
- **Shared Utilities**: Leverage `tests/utils/` for common patterns
- **Consistent Patterns**: Follow established mocking and validation patterns
- **Documentation**: Update guidelines when adding new test patterns

### Troubleshooting Common Issues
- **Test Failures**: Check server status with `curl localhost:8002/ws`
- **Authentication Errors**: Run `./set-test-env.sh` to refresh tokens
- **Port Conflicts**: Verify WebSocket (8002) vs Health (8003) usage
- **Configuration Drift**: Use base config inheritance, don't create custom configs

---

**Architecture Integration**: Follows go-architecture-guide.md principles adapted for client testing  
**Maintenance Focus**: Centralized patterns prevent server-side bloat and duplication issues  
**Team Coordination**: Standardized workflow ensures consistent testing across developers