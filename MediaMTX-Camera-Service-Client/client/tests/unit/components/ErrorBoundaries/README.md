# Error Boundary Test Suite

## Overview

This directory contains comprehensive tests for the Error Boundary components in the MediaMTX Camera Service Client. The test suite provides enterprise-grade coverage for error handling, recovery, and user experience scenarios.

## Test Files

### 1. FeatureErrorBoundary.test.tsx
**Purpose**: Tests the FeatureErrorBoundary component for feature-specific error handling
**Coverage**: 95%+ target coverage
**Test Scenarios**:
- Error catching and state management
- Retry mechanism with attempt tracking (1-3 attempts)
- Max retry limit enforcement
- Custom fallback component rendering
- Error logging service integration
- Development vs production error display
- User interaction handling (retry, reload, details toggle)
- Props validation and edge cases
- Error details toggle functionality
- Component lifecycle error handling

### 2. ServiceErrorBoundary.test.tsx
**Purpose**: Tests the ServiceErrorBoundary component for service-specific error handling
**Coverage**: 95%+ target coverage
**Test Scenarios**:
- Service-specific error handling
- Retryable vs non-retryable configuration
- Async retry operations with delays
- Fallback mode activation
- Error severity classification
- Service degradation scenarios
- Custom max retries configuration
- Error reporting integration
- Service timeout handling
- Network error simulation

### 3. ErrorBoundary.test.tsx
**Purpose**: Tests the basic ErrorBoundary component for application-level error handling
**Coverage**: 90%+ target coverage
**Test Scenarios**:
- Basic error catching functionality
- Fallback component rendering
- Development error details display
- Page reload functionality
- Error state management
- Props handling
- Error info logging

### 4. ErrorBoundaryIntegration.test.tsx
**Purpose**: Tests the integration of Error Boundaries in the application
**Coverage**: 90%+ target coverage
**Test Scenarios**:
- App.tsx error boundary hierarchy
- Feature error boundary isolation
- Service error boundary recovery
- Cross-boundary error propagation
- Error recovery flow validation
- Real-world error scenarios

## Test Utilities

### test-utils.tsx
**Purpose**: Provides helper functions and components for testing Error Boundaries
**Components**:
- `ErrorThrowingComponent`: Throws errors when shouldThrow is true
- `AsyncErrorComponent`: Throws errors after a delay for async testing
- `ErrorTypeComponent`: Throws different types of errors for comprehensive testing
- `renderWithTheme`: Custom render function with Material-UI theme provider
- `userInteractions`: Helper functions for simulating user interactions
- `errorBoundaryHelpers`: Helper functions for checking error boundary state
- `mockWindowReload`: Mocks window.location.reload for testing
- `waitForAsync`: Helper for waiting for async operations

## Test Configuration

### jest.error-boundary.config.cjs
**Purpose**: Jest configuration specifically for Error Boundary tests
**Features**:
- Isolated test environment for Error Boundary components
- Mock configuration for logger service
- Coverage thresholds: 95% for Error Boundary components
- Test timeout: 30 seconds for async operations
- Module mapping for clean imports

## Running Tests

### Individual Test Files
```bash
# Run specific test file
npx jest tests/unit/components/ErrorBoundaries/FeatureErrorBoundary.test.tsx --config jest.error-boundary.config.cjs

# Run with coverage
npx jest tests/unit/components/ErrorBoundaries/FeatureErrorBoundary.test.tsx --config jest.error-boundary.config.cjs --coverage
```

### All Error Boundary Tests
```bash
# Run all Error Boundary tests
npx jest tests/unit/components/ErrorBoundaries/ --config jest.error-boundary.config.cjs

# Run with coverage
npx jest tests/unit/components/ErrorBoundaries/ --config jest.error-boundary.config.cjs --coverage
```

### Using Test Runner Script
```bash
# Run comprehensive Error Boundary test suite
./run-error-boundary-tests.sh
```

## Coverage Requirements

### Global Coverage Targets
- **Branches**: 90% minimum
- **Functions**: 90% minimum
- **Lines**: 90% minimum
- **Statements**: 90% minimum

### Error Boundary Specific Targets
- **Branches**: 95% minimum
- **Functions**: 95% minimum
- **Lines**: 95% minimum
- **Statements**: 95% minimum

## Test Scenarios Covered

### Error Types
- JavaScript runtime errors
- Promise rejection errors
- Component lifecycle errors
- Service communication errors
- State management errors
- Async operation failures
- Network interruption errors
- Timeout errors

### User Interactions
- Retry button clicks
- Reload button clicks
- Fallback button clicks
- Error details toggle
- Multiple retry attempts
- Max retry limit handling

### Error Recovery
- Successful retry scenarios
- Partial recovery scenarios
- Complete error recovery
- Error state persistence
- Cross-boundary error isolation

### Real-World Scenarios
- WebSocket connection failures
- Component rendering errors
- Service timeout scenarios
- Network interruption handling
- Memory leak error simulation
- Concurrent error handling

## Quality Gates

### Test Execution
- All tests must pass with zero failures
- No flaky or intermittent test failures
- Tests must complete within reasonable time limits
- No memory leaks in test execution

### Coverage Validation
- Coverage thresholds must be met
- No uncovered critical paths
- All error scenarios must be tested
- Edge cases must be covered

### Code Quality
- No linting errors in test files
- TypeScript compilation success
- Proper test naming conventions
- Clear test documentation

## Troubleshooting

### Common Issues

1. **Test Timeout Errors**
   - Increase timeout in jest configuration
   - Check for infinite loops in test code
   - Verify async operations complete properly

2. **Mock Issues**
   - Ensure mocks are properly reset between tests
   - Check mock implementation matches real service
   - Verify mock timing for async operations

3. **Coverage Issues**
   - Check for uncovered branches in error handling
   - Verify all error scenarios are tested
   - Ensure edge cases are covered

### Debug Mode
```bash
# Run tests in debug mode
DEBUG=* npx jest tests/unit/components/ErrorBoundaries/ --config jest.error-boundary.config.cjs

# Run specific test with debugging
npx jest tests/unit/components/ErrorBoundaries/FeatureErrorBoundary.test.tsx --config jest.error-boundary.config.cjs --verbose --detectOpenHandles
```

## Contributing

### Adding New Tests
1. Follow existing test patterns and naming conventions
2. Use descriptive test names that explain the scenario
3. Include both positive and negative test cases
4. Add appropriate mocks and test utilities
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
- Reset mocks between tests

## Performance Testing

### Load Testing
- Multiple concurrent error boundaries
- Rapid error state changes
- Memory usage during error handling
- Response time for error recovery

### Benchmarking
- Error boundary activation time
- Retry operation performance
- Error recovery time
- Memory cleanup after errors

## Security Testing

### Input Validation
- Test all error message inputs
- Validate error boundary props
- Check for XSS in error messages
- Test error boundary security

### Data Sanitization
- Error message sanitization
- Component stack sanitization
- User input validation
- Error boundary isolation

---

**Last Updated**: 2024-09-22  
**Test Suite Version**: 1.0  
**Coverage Target**: 95%  
**Performance Target**: <50ms error boundary activation
