# Testing Strategy

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

## Naming Convention
- **Files**: `test_<what>_<type>.{ts,js}` using snake_case
- **Functions**: `test_<behavior>_<scenario>()`
- **Examples**: `test_camera_detail_component.ts`, `test_websocket_integration.ts`, `test_auth_flow_e2e.js`

---

**Status**: Core strategy for test-driven development with real integration focus