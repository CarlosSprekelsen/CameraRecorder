---
title: "Unified Testing Strategy - Client & Server Integration"
description: "Real Integration First testing approach with comprehensive client-server integration"
date: "2025-08-05"
---

# Unified Testing Strategy - Client & Server Integration

## Testing Philosophy
**"Real Integration First"** - Test against actual services whenever possible, use mocks only when external systems are truly unavailable.

## Test Environment Setup Requirements

### Development Testing Environment
**Objective**: Configure development environment for real server integration testing

**Prerequisites**: 
- MediaMTX Camera Service running via systemd (production deployment method)
- Client development server capability
- Network connectivity between client and server components

**Procedure**:
1. Verify server service status and accessibility
2. Configure client to connect to running server instance
3. Validate WebSocket connectivity between components
4. Establish test data state management procedures

## Test Categories and Coverage Requirements

### 1. Unit Tests (Component Level)
**Objective**: Test individual components in isolation
**Server Requirements**: Mock external dependencies only (MediaMTX, file system)
**Client Requirements**: Mock WebSocket service, test component logic
**Success Criteria**: 80%+ coverage for critical business logic

### 2. Integration Tests (Service Level) 
**Objective**: Test component interaction with real services
**Server Requirements**: Use real WebSocket, real MediaMTX, real file operations
**Client Requirements**: Connect to real WebSocket server instance
**Test Data Strategy**: Configure test cameras or video device simulation
**Performance Validation**: Verify API response times meet documented targets

### 3. End-to-End Tests (User Workflow)
**Objective**: Validate complete user scenarios
**Environment Requirements**: Real server, real client, real browser automation
**Scope**: Complete camera operations workflows
**Target Environment**: Production-like staging environment

## Client WebSocket Integration Testing

**Objective**: Create integration tests that validate WebSocket communication against running server

**Requirements**:
- Server availability validation before test execution
- WebSocket connection establishment verification
- JSON-RPC method call validation with expected response structures
- Error handling validation for connection failures

**Success Criteria**:
- All documented JSON-RPC methods return expected data structures
- Connection resilience tested (disconnect/reconnect scenarios)
- Performance targets validated (connection time, response times)

## Mock Service Fallback Strategy

**Objective**: Implement fallback testing for environments without server access

**Requirements**:
- Mock service configuration for CI/offline scenarios
- Environment variable controls for mock activation
- Mock response accuracy validation against real server responses

**Implementation Scope**:
- Enable mocking only when real server unavailable
- Use environment flags to control mock activation
- Ensure mock responses match real server behavior

## Performance Testing Integration

**Objective**: Validate performance requirements across client-server integration

### Server Performance Baseline
- Status methods: <50ms (documented server guarantee)
- Control methods: <100ms (documented server guarantee)
- WebSocket notifications: <20ms (documented server guarantee)

### Client Performance Requirements
- Initial load: <3s
- WebSocket connection: <1s  
- Bundle size: <2MB
- Memory usage: <50MB sustained

**Task**: Create performance validation procedures that measure end-to-end performance from client action to server response completion.

## Authentication Testing and Troubleshooting

**Objective**: Ensure proper authentication setup for all client-server integration tests

### Authentication Requirements
**Critical Issue**: All protected methods (take_snapshot, start_recording, stop_recording) require valid JWT authentication.

### JWT Token Generation
**Server Environment**: The MediaMTX Camera Service uses a JWT secret stored in the server environment:
- Environment file: `/opt/camera-service/.env`
- JWT Secret: `CAMERA_SERVICE_JWT_SECRET=d0adf90f433d25a0f1d8b9e384f77976fff12f3ecf57ab39364dcc83731aa6f7`

### ⚠️ CRITICAL: Authentication After Server Reinstall
**Important**: The JWT secret changes on every server reinstall, requiring environment variable setup for testing.

**Problem**: After reinstalling the MediaMTX Camera Service, authentication fails because:
1. **New JWT Secret**: Each reinstall generates a new `CAMERA_SERVICE_JWT_SECRET`
2. **Environment Mismatch**: The test environment doesn't have access to the new secret
3. **Authentication Failure**: Tests fail with "Invalid authentication token" errors

**Solution**: Use the provided environment setup script after each reinstall:

```bash
# Step 1: Set up environment variables after reinstall
./MediaMTX-Camera-Service-Client/client/set-test-env.sh

# Step 2: Run tests with correct environment
source .test_env && node MediaMTX-Camera-Service-Client/client/test-sprint-3-day-9-integration.js
```

**What the script does**:
- Reads the new JWT secret from `/opt/camera-service/.env` (requires sudo)
- Exports `CAMERA_SERVICE_JWT_SECRET` environment variable
- Creates `.test_env` file for future test runs
- Provides clear instructions for running tests

**Manual Alternative**:
```bash
# Extract JWT secret manually
JWT_SECRET=$(sudo grep "^CAMERA_SERVICE_JWT_SECRET=" /opt/camera-service/.env | cut -d'=' -f2)

# Run tests with environment variable
CAMERA_SERVICE_JWT_SECRET=$JWT_SECRET node MediaMTX-Camera-Service-Client/client/test-sprint-3-day-9-integration.js
```

### Authentication Process
1. **Generate Valid Token**: Use the correct JWT secret to generate tokens
2. **Authenticate**: Call the `authenticate` method with the token
3. **Use Protected Methods**: After successful authentication, protected methods become available

### Common Authentication Issues
**Problem**: "Authentication required - call authenticate or provide auth_token"
**Root Cause**: 
- Invalid JWT token (wrong secret, expired, malformed)
- Missing authentication call before using protected methods
- Token not properly passed in request parameters

**Solution**:
1. Use the correct JWT secret from server environment
2. Generate token with proper payload structure:
   ```javascript
   const payload = {
     user_id: 'test_user',
     role: 'operator', // or 'viewer', 'admin'
     iat: Math.floor(Date.now() / 1000),
     exp: Math.floor(Date.now() / 1000) + (24 * 60 * 60) // 24 hours
   };
   ```
3. Call authenticate method before using protected methods
4. Ensure token is passed as `token` parameter in authenticate request

### Testing Authentication
**Required Test Sequence**:
1. Connect to WebSocket server
2. Generate valid JWT token with correct secret
3. Call `authenticate` method with token
4. Verify authentication success
5. Test protected methods (take_snapshot, start_recording, stop_recording)

### Role-Based Access Control
- **viewer**: Can view camera status and take snapshots
- **operator**: Can take snapshots, start/stop recordings
- **admin**: Full access to all operations

### Authentication Test Example
```javascript
// Generate valid token
const token = jwt.sign(payload, JWT_SECRET, { algorithm: 'HS256' });

// Authenticate
const authResult = await sendRequest(ws, 'authenticate', { token });

// Use protected methods after successful authentication
const snapshotResult = await sendRequest(ws, 'take_snapshot', { device: '/dev/video0' });
```

## CI/CD Integration Testing

**Objective**: Establish automated testing pipeline with real server integration

### Pipeline Requirements
- Server service startup verification
- Service availability validation (health check endpoints)
- Client integration test execution against running server
- End-to-end workflow validation

**Implementation Steps**:
1. Define service startup and readiness verification procedures
2. Configure network connectivity validation between components
3. Establish test execution sequencing (server first, then client tests)
4. Define cleanup and teardown procedures

## Quality Gates for Integration Testing

**Objective**: Define acceptance criteria for integration test success

### Integration Test Requirements
- All JSON-RPC methods functional and returning correct data structures
- WebSocket connection resilience validated (disconnect/reconnect scenarios)
- Real-time notifications working with correct timing
- Error handling validated across all failure scenarios
- Performance targets met for all operations

### Mock Usage Guidelines
**Permitted Mocking**:
- External APIs beyond project control
- File system operations in unit tests only
- Time/date functions for deterministic testing

**Prohibited Mocking**:
- WebSocket communication between client and server
- JSON-RPC protocol implementation
- Client-server data flow integration

## Test Data Management Procedures

**Objective**: Establish consistent test data setup and cleanup

### Server Test Data Requirements
- Reference existing server test infrastructure and fixtures
- Utilize established camera simulation capabilities
- Maintain test file organization standards

### Client Test Data Requirements  
- Connect to established server test environment
- Implement state initialization procedures
- Define test isolation and cleanup requirements

## Documentation Testing Integration

**Objective**: Ensure all API documentation examples are validated against running implementation

### API Documentation Validation Requirements
- Server API documentation verified against actual implementation behavior
- Client API examples tested with real server responses  
- All code samples in documentation must be executable and tested

### Integration Documentation Validation
- Client-server integration guides tested end-to-end
- Deployment procedures validated in staging environment
- Troubleshooting guides verified with real failure scenarios

**Implementation Procedure**:
1. Identify all documentation containing API examples
2. Create validation tests that execute documented examples
3. Establish documentation update procedures when API changes
4. Define review process for documentation accuracy

## Folder Structure
```
tests/
  unit/         # isolated component and utility tests
  integration/  # API interaction tests (real server)
  e2e/          # end-to-end flows (Cypress)
  fixtures/     # test data and utilities
```

## ⚠️ CRITICAL TEST EXECUTION CONTEXT

### **REQUIRED EXECUTION DIRECTORY**
- **Integration Tests**: MUST be executed from `client/tests/integration/` directory
- **Component Paths**: Tests expect `client/src/components/` not `src/components/`
- **File References**: Use relative paths from test execution directory

### **COMMON EXECUTION ERRORS**
- ❌ Running integration tests from project root
- ❌ Incorrect component path references (`src/` vs `client/src/`)
- ❌ Creating test artifacts outside proper test directory structure

### **CORRECT TEST EXECUTION**
```bash
# Execute from correct directory
cd client/tests/integration && node test-with-valid-token.js

# Component paths in tests should reference from client root
const componentPath = 'src/components/Dashboard/Dashboard.tsx';  # ✅ Correct
const componentPath = 'client/src/components/Dashboard/Dashboard.tsx';  # ❌ Wrong
```


## Coverage Targets & Thresholds
- **Unit**: ≥ 80%
- **Integration**: ≥ 70%
- **E2E**: smoke tests covering critical flows
- **Performance**: All targets validated

## Environment Variables
```bash
# Real server integration
TEST_WEBSOCKET_URL=ws://localhost:8002/ws
TEST_API_URL=http://localhost:8002
TEST_HTTP_FALLBACK_URL=http://localhost:8003

# Authentication (set after reinstall)
CAMERA_SERVICE_JWT_SECRET=<extracted_from_server_env>

# Mock fallback
USE_MOCK_SERVER=true  # Only when real server unavailable
```

---

**Status**: Approved for Implementation  
**Alignment**: Fully aligned with server testing principles
