# Testing Implementation Plan - Web Client

**Version:** 1.0  
**Date:** 2025-01-25  
**Status:** MANDATORY - All AI agents must follow this plan exactly  
**Authority:** Ground Truth - Overrides any conflicting instructions

## 🚨 CRITICAL: AI Agent Directives

### MANDATORY AI Behavior Rules
1. **NEVER create duplicate testing utilities** - Check existing patterns first
2. **NEVER deviate from established patterns** - Use exact patterns from this plan
3. **NEVER create overlapping test categories** - Follow exact structure below
4. **ALWAYS validate against API documentation** - Never test against implementation
5. **STOP and ask for authorization** before creating new testing patterns

### Ground Truth Enforcement
- **API Documentation**: `mediamtx-camera-service-go/docs/api/mediamtx_camera_service_openrpc.json`
- **Client Architecture**: `client/docs/architecture/client-architechture.md`
- **Testing Guidelines**: `client/docs/development/client-testing-guidelines.md`
- **This Plan**: `client/docs/development/testing-implementation-plan.md`

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
└── config/               # Test configurations (FROZEN)
    ├── jest-unit.config.cjs
    ├── jest-integration.config.cjs
    └── jest-e2e.config.cjs
```

## Testing Utility Patterns (FROZEN)

### 1. API Client Abstraction (`tests/utils/api-client.ts`)

```typescript
/**
 * SINGLE WebSocket client abstraction for all tests
 * Environment-driven: real connections for integration, mocks for unit
 * 
 * Ground Truth References:
 * - API Documentation: ../mediamtx-camera-service-go/docs/api/mediamtx_camera_service_openrpc.json
 * - Client Architecture: ../docs/architecture/client-architechture.md
 * 
 * Requirements Coverage:
 * - REQ-API-001: WebSocket connection management
 * - REQ-API-002: JSON-RPC 2.0 protocol compliance
 * - REQ-API-003: Authentication token handling
 * 
 * Test Categories: Unit/Integration/Performance
 * API Documentation Reference: mediamtx_camera_service_openrpc.json
 */
export class TestAPIClient {
  private ws: WebSocket | null = null;
  private isMockMode: boolean;
  
  constructor(config: { mockMode?: boolean } = {}) {
    this.isMockMode = config.mockMode ?? process.env.NODE_ENV === 'test';
  }
  
  async connect(): Promise<void> {
    if (this.isMockMode) {
      // Mock WebSocket for unit tests
      return Promise.resolve();
    }
    // Real WebSocket for integration tests
    return this.connectReal();
  }
  
  async call(method: string, params: any[] = []): Promise<any> {
    // Single implementation for all JSON-RPC calls
    // Validates against documented API schema
  }
  
  async authenticate(token: string): Promise<AuthResult> {
    // Single authentication implementation
    // Uses documented auth flow exactly
  }
}
```

### 2. Authentication Helper (`tests/utils/auth-helper.ts`)

```typescript
/**
 * SINGLE authentication utility for all tests
 * Dynamic token generation - NO hardcoded credentials
 * 
 * Ground Truth References:
 * - API Documentation: ../mediamtx-camera-service-go/docs/api/mediamtx_camera_service_openrpc.json
 * 
 * Requirements Coverage:
 * - REQ-AUTH-001: JWT token generation
 * - REQ-AUTH-002: Role-based access control
 * - REQ-AUTH-003: Session management
 * 
 * Test Categories: Unit/Integration/Security
 * API Documentation Reference: mediamtx_camera_service_openrpc.json
 */
export class AuthHelper {
  static async generateTestToken(role: 'admin' | 'operator' | 'viewer' = 'admin'): Promise<string> {
    // Dynamic token generation using test secrets
    // NO hardcoded credentials allowed
  }
  
  static async validateAuthResult(result: any): Promise<boolean> {
    // Validates against documented AuthResult schema
  }
}
```

### 3. Response Validators (`tests/utils/validators.ts`)

```typescript
/**
 * SINGLE validation utility for all API responses
 * Validates against documented schemas only
 * 
 * Ground Truth References:
 * - API Documentation: ../mediamtx-camera-service-go/docs/api/mediamtx_camera_service_openrpc.json
 * 
 * Requirements Coverage:
 * - REQ-VAL-001: API response validation
 * - REQ-VAL-002: Error code validation
 * - REQ-VAL-003: Schema compliance
 * 
 * Test Categories: Unit/Integration/API-Compliance
 * API Documentation Reference: mediamtx_camera_service_openrpc.json
 */
export class APIResponseValidator {
  static validateCameraListResult(result: any): boolean {
    // Validates against documented CameraListResult schema
  }
  
  static validateRecordingStartResult(result: any): boolean {
    // Validates against documented RecordingStart schema
  }
  
  static validateErrorResponse(error: any, expectedCode: number): boolean {
    // Validates against documented error codes
  }
}
```

### 4. Centralized Mocks (`tests/utils/mocks.ts`)

```typescript
/**
 * SINGLE mock implementation per API concern
 * Based on documented API responses only
 * 
 * Ground Truth References:
 * - API Documentation: ../mediamtx-camera-service-go/docs/api/mediamtx_camera_service_openrpc.json
 * 
 * Requirements Coverage:
 * - REQ-MOCK-001: Consistent mock patterns
 * - REQ-MOCK-002: API-compliant responses
 * - REQ-MOCK-003: No duplicate implementations
 * 
 * Test Categories: Unit/Mock
 * API Documentation Reference: mediamtx_camera_service_openrpc.json
 */
export class APIMocks {
  static getCameraListResult(): CameraListResult {
    // Single mock implementation based on documented schema
  }
  
  static getRecordingStartResult(): RecordingStart {
    // Single mock implementation based on documented schema
  }
  
  static getErrorResponse(code: number): any {
    // Single error mock implementation
  }
}
```

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
  
  test('REQ-UNIT-002: Business logic validation', () => {
    // Test business logic
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

## AI Agent Compliance

### MANDATORY AI Behavior
1. **NEVER create new testing utilities** without checking existing patterns
2. **NEVER deviate from established patterns** in this plan
3. **ALWAYS validate against API documentation** - Never test implementation
4. **STOP and ask for authorization** before creating new patterns
5. **ALWAYS use centralized utilities** - No duplicate implementations

### Ground Truth Enforcement
- Tests must validate against documented API schemas
- Tests must follow exact patterns from this plan
- Tests must use centralized utilities only
- Tests must include requirements traceability

---

**Authority**: This plan is FROZEN and MANDATORY for all AI agents  
**Compliance**: All testing must follow this plan exactly  
**Enforcement**: Deviations require explicit authorization
