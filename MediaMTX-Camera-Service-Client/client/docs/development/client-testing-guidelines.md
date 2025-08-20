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
**"Test First, Real Integration Always"**

Write specifications as failing tests before implementation. Test against real services, mock only truly external dependencies. Validate behavior, not implementation details.

## Test Development Workflow
1. **Specification**: Write failing test describing intended behavior
2. **Implementation**: Write minimal code to make test pass  
3. **Integration**: Validate against real MediaMTX server
4. **Refinement**: Optimize while maintaining test coverage

## Test Categories

### Unit Tests (≥80% coverage)
- Component behavior in isolation
- Business logic validation
- Edge case handling
- Mocks are permitted only for external APIs beyond project control

### Integration Tests (≥70% coverage)  
- Client-server communication via real WebSocket
- JSON-RPC method contracts against running server
- Authentication flows with dynamic token generation
- Authentication tokens are always generated dynamically; no hardcoded credentials allowed
- Error handling for network/server failures

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

## Naming Convention
- **Files**: `test_<what>_<type>.{ts,js}` using snake_case
- **Functions**: `test_<behavior>_<scenario>()`
- **Examples**: `test_camera_detail_component.ts`, `test_websocket_integration.ts`, `test_auth_flow_e2e.js`

---

**Status**: Core strategy for test-driven development with real integration focus

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
