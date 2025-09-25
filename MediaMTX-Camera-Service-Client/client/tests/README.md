# Testing Implementation - Web Client

**Version:** 1.0  
**Date:** 2025-01-25  
**Status:** MANDATORY - All AI agents must follow this implementation exactly  
**Authority:** Ground Truth - Overrides any conflicting instructions

## 🚨 CRITICAL: AI Agent Directives

### MANDATORY AI Behavior Rules
1. **NEVER create duplicate testing utilities** - Check existing patterns first
2. **NEVER deviate from established patterns** - Use exact patterns from this implementation
3. **NEVER create overlapping test categories** - Follow exact structure below
4. **ALWAYS validate against API documentation** - Never test against implementation
5. **STOP and ask for authorization** before creating new testing patterns

### Ground Truth Enforcement
- **API Documentation**: `mediamtx-camera-service-go/docs/api/mediamtx_camera_service_openrpc.json`
- **Client Architecture**: `client/docs/architecture/client-architechture.md`
- **Testing Guidelines**: `client/docs/development/client-testing-guidelines.md`
- **Testing Implementation Plan**: `client/docs/development/testing-implementation-plan.md`
- **This Implementation**: `client/tests/README.md`

## Directory Structure (FROZEN)

```
client/tests/
├── unit/                    # Component isolation tests
│   ├── components/         # React component tests
│   ├── hooks/             # Custom hook tests
│   ├── services/          # Service layer tests
│   └── utils/             # Utility function tests
├── integration/            # Real server communication tests
│   ├── api/              # JSON-RPC method tests
│   ├── auth/             # Authentication flow tests
│   └── websocket/        # WebSocket connection tests
├── e2e/                   # Complete workflow tests
│   ├── camera-operations/ # Camera control workflows
│   ├── file-management/  # File operations workflows
│   └── system-monitoring/ # Health and status workflows
├── fixtures/              # Shared test data (FROZEN)
│   ├── api-responses/    # Documented API response samples
│   ├── camera-data/      # Camera configuration samples
│   └── auth-tokens/      # Test authentication tokens
├── utils/                 # Centralized test utilities (FROZEN)
│   ├── api-client.ts     # Single WebSocket client abstraction
│   ├── auth-helper.ts    # Authentication utilities
│   ├── validators.ts     # Response validation utilities
│   ├── mocks.ts          # Centralized mock implementations
│   └── test-helpers.ts   # Common test utilities
├── config/               # Test configurations (FROZEN)
│   ├── jest-unit.config.cjs
│   ├── jest-integration.config.cjs
│   └── jest-e2e.config.cjs
├── setup.ts              # Unit test setup
├── setup.integration.ts  # Integration test setup
└── README.md             # This file
```

## Testing Utility Patterns (FROZEN)

### 1. API Client Abstraction (`utils/api-client.ts`)
- **SINGLE WebSocket client abstraction** for all tests
- Environment-driven: real connections for integration, mocks for unit
- Validates against documented API schema
- **MANDATORY**: Use this client for all API tests

### 2. Authentication Helper (`utils/auth-helper.ts`)
- **SINGLE authentication utility** for all tests
- Dynamic token generation - NO hardcoded credentials
- Role-based access control validation
- **MANDATORY**: Use this helper for all auth tests

### 3. Response Validators (`utils/validators.ts`)
- **SINGLE validation utility** for all API responses
- Validates against documented schemas only
- Error code validation
- **MANDATORY**: Use this validator for all response tests

### 4. Centralized Mocks (`utils/mocks.ts`)
- **SINGLE mock implementation** per API concern
- Based on documented API responses only
- No duplicate mock patterns
- **MANDATORY**: Use this mock for all unit tests

### 5. Test Helpers (`utils/test-helpers.ts`)
- Common test utilities for all test categories
- Environment setup and cleanup
- Test data generation
- **MANDATORY**: Use this helper for all test setup

## Test Category Patterns (FROZEN)

### Unit Tests Pattern
```typescript
/**
 * Unit test template - Component isolation
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - API Documentation: ../mediamtx-camera-service-go/docs/api/mediamtx_camera_service_openrpc.json
 * 
 * Requirements Coverage:
 * - REQ-UNIT-001: Component behavior validation
 * - REQ-UNIT-002: Business logic testing
 * - REQ-UNIT-003: Edge case handling
 * 
 * Test Categories: Unit
 * API Documentation Reference: mediamtx_camera_service_openrpc.json
 */
describe('ComponentName Tests', () => {
  beforeEach(() => {
    // Use centralized mocks
    jest.clearAllMocks();
  });
  
  test('REQ-UNIT-001: Component renders correctly', () => {
    // Test component rendering
  });
});
```

### Integration Tests Pattern
```typescript
/**
 * Integration test template - Real server communication
 * 
 * Ground Truth References:
 * - API Documentation: ../mediamtx-camera-service-go/docs/api/mediamtx_camera_service_openrpc.json
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * 
 * Requirements Coverage:
 * - REQ-INT-001: Real API communication
 * - REQ-INT-002: Authentication flow validation
 * - REQ-INT-003: Error handling validation
 * 
 * Test Categories: Integration/API-Compliance
 * API Documentation Reference: mediamtx_camera_service_openrpc.json
 */
describe('APIMethodName Integration Tests', () => {
  let apiClient: TestAPIClient;
  let authHelper: AuthHelper;
  
  beforeAll(async () => {
    // Load test environment
    await loadTestEnvironment();
    apiClient = new TestAPIClient({ mockMode: false });
    authHelper = new AuthHelper();
  });
  
  test('REQ-INT-001: Method call with valid parameters', async () => {
    const token = await authHelper.generateTestToken('admin');
    await apiClient.authenticate(token);
    
    const result = await apiClient.call('method_name', [param1, param2]);
    
    expect(APIResponseValidator.validateMethodResult(result)).toBe(true);
  });
});
```

### E2E Tests Pattern
```typescript
/**
 * E2E test template - Complete workflows
 * 
 * Ground Truth References:
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * - API Documentation: ../mediamtx-camera-service-go/docs/api/mediamtx_camera_service_openrpc.json
 * 
 * Requirements Coverage:
 * - REQ-E2E-001: Complete user workflows
 * - REQ-E2E-002: Real hardware interaction
 * - REQ-E2E-003: Performance validation
 * 
 * Test Categories: E2E/Performance
 * API Documentation Reference: mediamtx_camera_service_openrpc.json
 */
describe('WorkflowName E2E Tests', () => {
  let apiClient: TestAPIClient;
  
  beforeAll(async () => {
    await loadTestEnvironment();
    apiClient = new TestAPIClient({ mockMode: false });
  });
  
  test('REQ-E2E-001: Complete workflow execution', async () => {
    // Complete user workflow test
  });
});
```

## Test Execution Commands

### Unit Tests
```bash
cd MediaMTX-Camera-Service-Client/client
npm run test:unit
# or
jest --config tests/config/jest-unit.config.cjs
```

### Integration Tests
```bash
cd MediaMTX-Camera-Service-Client/client
./set-test-env.sh
source .test_env
npm run test:integration
# or
jest --config tests/config/jest-integration.config.cjs
```

### E2E Tests
```bash
cd MediaMTX-Camera-Service-Client/client
./set-test-env.sh
source .test_env
npm run test:e2e
# or
jest --config tests/config/jest-e2e.config.cjs
```

### All Tests
```bash
cd MediaMTX-Camera-Service-Client/client
./set-test-env.sh
source .test_env
npm run test:all
# or
npm run test:unit && npm run test:integration && npm run test:e2e
```

## Environment Setup (MANDATORY)

### ALWAYS Required Before Testing
```bash
cd MediaMTX-Camera-Service-Client/client
./set-test-env.sh
source .test_env
npm test
```

### Server Port Configuration
- **WebSocket Server (JSON-RPC)**: Port 8002
- **Health Server (REST)**: Port 8003
- **MANDATORY**: Do not mix WebSocket methods with health endpoints

## Quality Gates

### Coverage Enforcement
- Unit tests: ≥80% coverage
- Integration tests: ≥70% coverage
- E2E tests: Critical paths only
- API compliance: 100% of documented methods

### Performance Targets
- Status methods: <50ms (p95)
- Control methods: <100ms (p95)
- WebSocket connection: <1s (p95)
- Client load: <3s (p95)

## Anti-Patterns (FORBIDDEN)

### ❌ FORBIDDEN: Multiple Mock Implementations
```typescript
// ❌ NEVER DO THIS - Creates duplicate patterns
export class CameraServiceMock { }
export class CameraServiceMockV2 { }
export class CameraServiceTestMock { }
```

### ❌ FORBIDDEN: Implementation-Dependent Testing
```typescript
// ❌ NEVER DO THIS - Tests implementation, not API
test('should call internal method', () => {
  const service = new CameraService();
  service.internalMethod(); // Testing private implementation
});
```

### ❌ FORBIDDEN: Hardcoded Test Data
```typescript
// ❌ NEVER DO THIS - Hardcoded credentials
const testToken = 'hardcoded-jwt-token';
const testCamera = { device: 'camera0', status: 'CONNECTED' };
```

### ✅ REQUIRED: Ground Truth Validation
```typescript
// ✅ ALWAYS DO THIS - Validate against documentation
test('API response matches documented schema', () => {
  const result = await apiClient.call('get_camera_list');
  expect(APIResponseValidator.validateCameraListResult(result)).toBe(true);
});
```

## Implementation Checklist

### Before Writing Any Test
- [ ] Check existing utilities in `tests/utils/`
- [ ] Validate against API documentation
- [ ] Plan shared mock strategy
- [ ] Identify architecture integration points

### During Implementation
- [ ] Use established patterns from `tests/utils/`
- [ ] Follow exact naming conventions
- [ ] Add requirements traceability headers
- [ ] Validate against documented APIs

### After Implementation
- [ ] Verify no duplicate patterns created
- [ ] Check coverage thresholds
- [ ] Test against real endpoints
- [ ] Document any new shared utilities

## AI Agent Compliance

### MANDATORY AI Behavior
1. **NEVER create new testing utilities** without checking existing patterns
2. **NEVER deviate from established patterns** in this implementation
3. **ALWAYS validate against API documentation** - Never test implementation
4. **STOP and ask for authorization** before creating new patterns
5. **ALWAYS use centralized utilities** - No duplicate implementations

### Ground Truth Enforcement
- Tests must validate against documented API schemas
- Tests must follow exact patterns from this implementation
- Tests must use centralized utilities only
- Tests must include requirements traceability

---

**Authority**: This implementation is FROZEN and MANDATORY for all AI agents  
**Compliance**: All testing must follow this implementation exactly  
**Enforcement**: Deviations require explicit authorization
