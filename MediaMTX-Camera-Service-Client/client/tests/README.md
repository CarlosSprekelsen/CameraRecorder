# MediaMTX Camera Service Client - Test Suite

## Overview

This test suite provides comprehensive validation for the MediaMTX Camera Service Client, ensuring all functionality works correctly and meets quality standards. The tests follow the "Real Integration First" approach, testing against actual server implementations when possible.

## Test Structure

```
tests/
├── unit/                          # Unit tests for individual components
│   ├── components/                # React component tests
│   │   ├── CameraDetail.test.tsx  # Camera detail component tests
│   │   └── FileManager.test.tsx   # File manager component tests
│   ├── services/                  # Service layer tests
│   │   └── websocket.test.ts      # WebSocket service tests
│   └── stores/                    # State management tests
│       └── fileStore.test.ts      # File store tests
├── integration/                   # Integration tests
│   ├── camera-operations-integration.test.ts  # Camera operations
│   ├── websocket-integration.test.ts          # WebSocket integration
│   └── ci-cd-integration.test.ts              # CI/CD pipeline tests
├── fixtures/                      # Test data and mocks
├── setup.ts                       # Test environment setup
├── setup-integration.ts           # Integration test setup
├── run-validation-tests.sh        # Test runner script
└── README.md                      # This file
```

## Test Categories

### 1. Unit Tests

**Purpose**: Test individual components and functions in isolation
**Coverage**: 80%+ for critical business logic
**Mocking**: Extensive use of mocks for dependencies

#### Component Tests
- **CameraDetail.test.tsx**: Tests camera detail component functionality
  - Snapshot controls (format, quality, capture)
  - Recording controls (start, stop, duration, format)
  - Camera status display
  - Error handling
  - Loading states
  - User interactions

- **FileManager.test.tsx**: Tests file management component
  - File browsing (recordings vs snapshots)
  - File download functionality
  - Pagination controls
  - Tab navigation
  - File metadata display
  - Error handling

#### Store Tests
- **fileStore.test.ts**: Tests file management state store
  - WebSocket integration
  - File listing operations
  - Download functionality
  - State management
  - Error handling
  - Event handling

#### Service Tests
- **websocket.test.ts**: Tests WebSocket JSON-RPC client
  - Connection management
  - JSON-RPC protocol
  - Error handling
  - Reconnection logic
  - Message handling

### 2. Integration Tests

**Purpose**: Test component interactions and real server communication
**Coverage**: End-to-end functionality validation
**Real Server**: Uses actual MediaMTX server when available

#### Camera Operations Integration
- **camera-operations-integration.test.ts**: Real server integration
  - Camera discovery and status
  - Snapshot capture with real cameras
  - Recording start/stop with real cameras
  - File management integration
  - Performance validation
  - Error handling with real server

#### WebSocket Integration
- **websocket-integration.test.ts**: WebSocket communication
  - Real WebSocket connection
  - JSON-RPC message exchange
  - Connection stability
  - Error recovery

## Running Tests



### Quick Start

```bash
# Run all tests
npm test

# Run with coverage
npm run test:coverage

# Run specific test categories
npm run test:unit          # Unit tests only
npm run test:integration   # Integration tests only
```

### Comprehensive Validation

```bash
# Run the full validation suite
./tests/run-validation-tests.sh
```

This script runs:
1. All unit tests
2. All integration tests
3. Coverage analysis
4. TypeScript compilation check
5. Linting validation
6. Comprehensive reporting

### Individual Test Files

```bash
# Run specific test files
npm test -- tests/unit/components/CameraDetail.test.tsx
npm test -- tests/integration/camera-operations-integration.test.ts
npm test -- tests/unit/stores/fileStore.test.ts
```

## Test Configuration

### Environment Variables

```bash
# WebSocket server URL for integration tests
TEST_WS_URL=ws://localhost:8002/ws

# Mock server flag (for testing without real server)
USE_MOCK_SERVER=true
```

### Jest Configuration

Key settings in `jest.config.js`:
- **Test Environment**: `jsdom` for React component testing
- **Coverage Threshold**: 80% for branches, functions, lines, statements
- **Timeout**: 30 seconds for integration tests
- **Setup Files**: Automatic test environment setup
- **Module Mapping**: Path aliases for clean imports

## Test Data and Fixtures

### Mock Camera Data
```typescript
const mockCamera = {
  device: 'test-camera-1',
  status: 'CONNECTED',
  name: 'Test Camera',
  resolution: '1920x1080',
  fps: 30,
  streams: {
    rtsp: 'rtsp://localhost:8554/test-camera-1',
    webrtc: 'webrtc://localhost:8889/test-camera-1',
    hls: 'http://localhost:8888/test-camera-1/index.m3u8'
  },
  metrics: {
    bytes_sent: 1024000,
    readers: 2,
    uptime: 3600
  }
};
```

### Mock File Data
```typescript
const mockRecordings = [
  {
    filename: 'recording-1.mp4',
    file_size: 1024000,
    created_at: '2024-01-01T00:00:00Z',
    modified_time: '2024-01-01T00:01:00Z',
    download_url: '/files/recordings/recording-1.mp4',
    duration: 60,
    format: 'mp4'
  }
];
```

## Quality Gates

### Coverage Requirements
- **Global Coverage**: 80% minimum
- **Critical Paths**: 90% minimum
- **Business Logic**: 95% minimum

### Performance Targets
- **Camera Operations**: <100ms response time
- **File Operations**: <5s for large file lists
- **WebSocket Connection**: <1s establishment time

### Code Quality
- **TypeScript**: Strict mode compilation
- **ESLint**: Zero linting errors
- **Prettier**: Consistent code formatting

## Test Scenarios

### Camera Operations
1. **Snapshot Capture**
   - Default settings (JPG, 80% quality)
   - Custom format (PNG)
   - Custom quality settings
   - Error handling (invalid camera)
   - Loading states

2. **Video Recording**
   - Timed recording (10 seconds)
   - Unlimited recording
   - Start/stop operations
   - Format selection (MP4, MKV)
   - Error handling

3. **Camera Status**
   - Connection status monitoring
   - Real-time updates
   - Metrics display
   - Stream URL access

### File Management
1. **File Browsing**
   - Recordings list with pagination
   - Snapshots list with pagination
   - File metadata display
   - Tab navigation

2. **File Download**
   - Secure HTTPS download
   - Progress indication
   - Error handling
   - Filename preservation

3. **File Operations**
   - File size formatting
   - Duration formatting
   - Date formatting
   - File type detection

### WebSocket Integration
1. **Connection Management**
   - Connection establishment
   - Automatic reconnection
   - Error recovery
   - Connection monitoring

2. **JSON-RPC Protocol**
   - Request/response handling
   - Error code mapping
   - Notification processing
   - Message validation

## Continuous Integration

### CI/CD Pipeline
The test suite integrates with CI/CD pipelines:
- **Pre-commit**: Unit tests and linting
- **Pull Request**: Full test suite + coverage
- **Deployment**: Integration tests with real server

### Test Reports
- **Coverage Reports**: HTML and LCOV formats
- **Test Results**: JUnit XML format
- **Performance Metrics**: Response time tracking
- **Quality Gates**: Automated validation

## Troubleshooting

### Common Issues

1. **Node.js Version**
   ```bash
   # Check Node.js version
   node --version
   # Required: ^20.19.0 || >=22.12.0
   ```

2. **WebSocket Connection**
   ```bash
   # Test WebSocket server
   curl -I http://localhost:8002/health
   ```

3. **Test Environment**
   ```bash
   # Reset test environment
   npm run test:clean
   npm install
   ```

### Debug Mode
```bash
# Run tests in debug mode
DEBUG=* npm test

# Run specific test with debugging
npm test -- --verbose --detectOpenHandles
```

## Contributing

### Adding New Tests
1. Follow existing test patterns
2. Use descriptive test names
3. Include both positive and negative cases
4. Add appropriate mocks
5. Update coverage thresholds if needed

### Test Naming Convention
```typescript
describe('ComponentName', () => {
  describe('Feature', () => {
    it('should do something when condition', () => {
      // Test implementation
    });
  });
});
```

### Mock Guidelines
- Mock external dependencies
- Use realistic test data
- Avoid over-mocking
- Test error conditions

## Performance Testing

### Load Testing
```bash
# Run performance tests
npm run test:performance
```

### Benchmarking
- Component render times
- Store operation performance
- WebSocket message throughput
- File download speeds

## Security Testing

### Input Validation
- Test all user inputs
- Validate file operations
- Check authentication flows
- Test error handling

### Data Sanitization
- File path validation
- URL construction
- Error message sanitization
- XSS prevention

---

**Last Updated**: 2024-01-01  
**Test Suite Version**: 1.0  
**Coverage Target**: 80%  
**Performance Target**: <100ms camera operations
