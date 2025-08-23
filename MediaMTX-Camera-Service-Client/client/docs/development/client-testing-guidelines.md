# Testing Strategy

## ⚠️ CRITICAL: Environment Setup Required
**ALWAYS run `source .test_env` before executing tests**
- Authentication failures occur without proper environment variables
- JWT secret and server URL must be loaded before test execution
- This is the #1 cause of false test failures

```bash
# REQUIRED: Load environment before any test execution
cd MediaMTX-Camera-Service-Client/client
source .test_env
npm test
```

## ⚠️ CRITICAL: Authentication Setup Required
**ALWAYS run `./set-test-env.sh` before executing tests**
- Tests accessing protected methods require valid authentication tokens
- Environment variables get synced with git and may be outdated
- `set-test-env.sh` ensures current valid keys for your server instance
- **MANDATORY**: All tests calling protected methods must authenticate first

```bash
# REQUIRED: Set up authentication before any test execution
cd MediaMTX-Camera-Service-Client/client
./set-test-env.sh
source .test_env
npm test
```

## ⚠️ CRITICAL: Server Port Configuration
**MediaMTX Camera Service has TWO endpoints - use correct ports:**
- **WebSocket Server (JSON-RPC)**: Port 8002 - All camera operations, file management
- **Health Server (REST)**: Port 8003 - Health checks, system status
- **MANDATORY**: Tests must use correct endpoint for each operation type
- **MANDATORY**: Do not mix WebSocket methods with health endpoints

```bash
# WebSocket operations (camera control, file management)
ws://localhost:8002/ws

# Health operations (system status, monitoring)  
http://localhost:8003/health/*
```

## Philosophy
**"Test Against Ground Truth, Never Against Implementation"**

Write specifications as failing tests before implementation. Test against frozen documentation (API docs, architecture docs), never against existing code. Validate behavior against ground truth, not implementation details.

## Ground Truth Enforcement
- **API Documentation is FROZEN** - Only documented API as reference
- **Client Architecture is AUTHORITATIVE** - Only architecture docs as reference
- **Tests validate against ground truth** - Not existing code
- **Test failures indicate real problems** - Not test problems
- **No "accommodation" of broken implementations** - Tests do not fix the implementation

## 🚨 CRITICAL RULES
1. **STOP and ask for authorization** before making any test changes
2. **NEVER look at server implementation code** - Only use API documentation
3. **NEVER look at client implementation code** - Only use architecture documentation
4. **Tests must validate against ground truth** - Not against existing code
5. **If test fails, check ground truth first** - Don't adapt test to broken implementation
6. **Test failures are real bugs** - Not test bugs to be fixed
7. **Purpose of testing is validation** - Not to pass by any means

## Ground Truth Sources (FROZEN)
- **Server WebSocket API**: `mediamtx-camera-service/docs/api/json-rpc-methods.md` (FROZEN)
- **Server Health API**: `mediamtx-camera-service/docs/api/health-endpoints.md` (FROZEN)
- **Client Architecture**: `client/docs/architecture/client-architecture.md` (AUTHORITATIVE)
- **Client Requirements**: `client/docs/requirements/client-requirements.md` (AUTHORITATIVE)
- **Naming Strategy**: `client/docs/development/naming-strategy.md` (MANDATORY)

## Test Development Workflow
1. **Ground Truth Validation**: Write failing test against frozen documentation
2. **Implementation**: Write minimal code to align with ground truth
3. **Integration**: Validate against real MediaMTX server using documented API
4. **Refinement**: Optimize while maintaining ground truth compliance

## Test Categories

### Unit Tests (≥80% coverage)
- Component behavior in isolation
- Business logic validation
- Edge case handling
- Mocks are permitted only for external APIs beyond project control

## API Compliance Testing - MANDATORY

### 🚨 CRITICAL: API Documentation Compliance
Every test that calls server APIs MUST validate against API documentation, not implementation.

### Mandatory API Compliance Rules
1. **Test against documented API format** - Use exact request/response formats from `json-rpc-methods.md`
2. **Validate documented error codes** - Use error codes and messages from API documentation
3. **Test documented authentication flow** - Follow authentication flow exactly as documented
4. **Verify documented response fields** - Check all required fields are present and correct
5. **No implementation-specific testing** - Don't test server internals, only documented behavior

### API Compliance Test Template
```typescript
/**
 * API Compliance Test for [Method Name]
 * 
 * Ground Truth References:
 * - Server API: ../mediamtx-camera-service/docs/api/json-rpc-methods.md
 * - Client Architecture: ../docs/architecture/client-architecture.md
 * - Client Requirements: ../docs/requirements/client-requirements.md
 * 
 * Method: [method_name]
 * Expected Request Format: [documented format]
 * Expected Response Format: [documented format]
 * Expected Error Codes: [documented codes]
 */

describe('API Compliance Tests', () => {
  test('[method_name] validates against API documentation', async () => {
    // 1. Use documented request format from API documentation
    const request = {
      jsonrpc: "2.0",
      method: "[method_name]",
      params: {
        // Use exact parameter names from API documentation
      },
      id: 1
    };
    
    // 2. Validate documented response format
    const response = await sendRequest(request);
    
    // 3. Check all documented fields are present
    expect(response).toHaveProperty('result');
    const result = response.result;
    
    // 4. Validate documented response structure
    const requiredFields = ["field1", "field2"]; // From API documentation
    requiredFields.forEach(field => {
      expect(result).toHaveProperty(field, `Missing required field '${field}' per API documentation`);
    });
    
    // 5. Validate documented error handling
    // Test error cases exactly as documented
  });
});
```

### Integration Tests (≥70% coverage)  
- Client-server communication via real WebSocket using documented API
- JSON-RPC method contracts against running server (validated against API documentation)
- Authentication flows with dynamic token generation (following documented flow)
- Authentication tokens are always generated dynamically; no hardcoded credentials allowed
- Error handling for network/server failures (using documented error codes)

### End-to-End Tests (Critical paths)
- Complete user workflows in real browser
- Camera operations with actual hardware/simulation
- File management end-to-end
- Performance validation under load

## Quality Gates

### Performance Targets
- Status methods: <50ms (p95 under load)
- Control methods: <100ms (p95 under load)
- WebSocket connection: <1s (p95 under load)
- Client load: <3s (p95 under load)

### Coverage Enforcement
- Automated threshold validation
- Fail build on coverage regression
- Focus on critical business logic paths

### Integration Requirements
- All tests pass against real server
- Authentication handled automatically via environment
- No hardcoded credentials or mocked server responses

## Test Structure
```
tests/
├── unit/           # Isolated component/logic tests
├── integration/    # Real server communication tests  
├── e2e/           # Complete workflow tests
└── fixtures/      # Shared test utilities
```

## Environment-Specific Configurations
When test environment limitations prevent real integration testing (e.g., jsdom WebSocket restrictions), separate Jest configurations are permitted to maintain Real Integration First principles while preserving proper test boundaries.

## ⚠️ IV&V Testing Protocol
**ALWAYS run tests from `client/` directory**: `cd client && npm test`  
**NEVER run from root** - conflicting dependencies cause failures

## Quality Gates

### Ground Truth Compliance Gates
- **Authorization Required**: All test changes must be explicitly authorized
- **Ground Truth Compliance**: 100% validation against frozen documentation
- **No Code Peeking**: Tests must not reference implementation code
- **API Compliance**: All tests must validate against API documentation
- **Architecture Compliance**: All tests must validate against client architecture
- **Test Failures are Real**: No accommodation of broken implementations

### Performance Targets
- Status methods: <50ms (p95 under load)
- Control methods: <100ms (p95 under load)
- WebSocket connection: <1s (p95 under load)
- Client load: <3s (p95 under load)

### Coverage Enforcement
- Automated threshold validation
- Fail build on coverage regression
- Focus on critical business logic paths

## Naming Convention
- **Files**: `test_<what>_<type>.{ts,js}` using snake_case
- **Functions**: `test_<behavior>_<scenario>()`
- **Examples**: `test_camera_detail_component.ts`, `test_websocket_integration.ts`, `test_auth_flow_e2e.js`

---

## Ground Truth Validation Rules

### Prevent Sneak Peeking into Code
1. **NEVER look at server implementation code** - Only use API documentation
2. **NEVER look at client implementation code** - Only use architecture documentation
3. **Tests must validate against ground truth** - Not against existing code
4. **If test fails, check ground truth first** - Don't adapt test to broken implementation
5. **Test failures are real bugs** - Not test bugs to be fixed

### Testing Rules Violation Prevention
1. **Tests validate against ground truth** - Not existing code
2. **Tests must fail if implementation doesn't match ground truth**
3. **No adapting tests to existing code flaws**
4. **Test failures indicate real problems** - Not test problems
5. **Purpose of testing is validation** - Not to pass by any means

---

**Status**: **UPDATED** - Core strategy for test-driven development with ground truth validation focus

## Test Execution Requirements
- Unit tests: Use `jest.config.cjs` (jsdom + mocks)
- Integration tests: Use `jest.integration.config.cjs` (Node.js + real WebSocket)
- Performance tests: Use `jest.integration.config.cjs` (Node.js + real WebSocket)
- E2E tests: Use `jest.integration.config.cjs` (Node.js + real WebSocket)

## WebSocket Testing
- Unit tests: Mock WebSocket services completely
- Integration tests: Use `require('ws')` for Node.js WebSocket
- Browser tests: Use native WebSocket API

## Requirements Traceability
- Every test file MUST include REQ-* header
- REQ-* IDs MUST map to actual project requirements
- No test without requirements traceability

### Mandatory Format for Test Files
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
