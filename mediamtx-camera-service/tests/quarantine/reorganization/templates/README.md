# Test Templates and Patterns

**Purpose:** Prevent future test guideline violations by providing compliant templates and patterns for all test types.

## Overview

This directory contains comprehensive templates and patterns that follow all test guidelines:
- Real component integration over mocking
- Requirements traceability with REQ-* references
- Comprehensive error condition testing
- Descriptive test naming
- Proper documentation and comments

## Template Files

### 1. `unit_test_template.py`
**Purpose:** Compliant unit test structure with real component integration

**Key Features:**
- Real MediaMTX integration patterns
- Real WebSocket communication patterns
- Comprehensive error condition testing
- Requirements traceability examples
- Performance testing patterns
- Resource management testing

**Usage:**
```python
# Copy the template and replace placeholders
from tests.templates.unit_test_template import TestYourComponent

class TestMyComponent(TestYourComponent):
    # Override fixtures with your specific component
    @pytest.fixture
    def component(self, test_config):
        return MyComponent(test_config)
    
    # Add your specific test methods
    # They inherit the structure and patterns from the template
```

### 2. `integration_test_template.py`
**Purpose:** Real component integration testing patterns

**Key Features:**
- End-to-end system integration testing
- Real MediaMTX service integration
- Real WebSocket communication
- Real file system operations
- Performance and load testing
- Data consistency validation

**Usage:**
```python
# Copy the template for integration testing
from tests.templates.integration_test_template import TestSystemIntegration

class TestMySystemIntegration(TestSystemIntegration):
    # Override configuration for your system
    @pytest.fixture
    def integration_config(self, temp_test_directory):
        return MySystemConfig(...)
    
    # Add your specific integration test methods
```

### 3. `edge_case_patterns.py`
**Purpose:** Common edge case testing approaches

**Key Features:**
- Input validation edge cases
- Resource constraint scenarios
- Network and service failure patterns
- Performance and load edge cases
- Data boundary conditions
- Recovery and resilience patterns

**Usage:**
```python
# Import edge case patterns
from tests.templates.edge_case_patterns import (
    EdgeCasePatterns,
    test_input_validation_edge_cases,
    test_resource_constraint_edge_cases
)

# Use in your tests
@pytest.mark.parametrize("invalid_input", EdgeCasePatterns.input_validation_edge_cases())
def test_my_component_input_validation(invalid_input, my_component):
    # Your specific validation logic
    pass
```

### 4. `error_testing_patterns.py`
**Purpose:** Error condition validation approaches

**Key Features:**
- Service failure patterns
- Network failure patterns
- Input validation patterns
- Resource constraint patterns
- Recovery and resilience patterns
- Circuit breaker testing

**Usage:**
```python
# Import error testing patterns
from tests.templates.error_testing_patterns import (
    ErrorTestingPatterns,
    test_service_failure_patterns,
    test_network_failure_patterns
)

# Use in your tests
@pytest.mark.parametrize("failure_type", ErrorTestingPatterns.service_failure_patterns())
async def test_my_component_service_failures(failure_type, my_component):
    # Your specific error handling logic
    pass
```

### 5. `requirements_mapping_guide.md`
**Purpose:** How to map tests to requirements

**Key Features:**
- Requirements reference format
- Module-specific requirement prefixes
- Test file header templates
- Individual test method templates
- Coverage validation tools
- Best practices

**Usage:**
```python
"""
Test [specific functionality] with real component integration.

Requirements Traceability:
- REQ-[MODULE]-001: [Primary requirement description]
- REQ-[MODULE]-002: [Secondary requirement description]
- REQ-ERROR-[NUMBER]: [Error condition requirement]

Story Coverage: [Story ID] - [Story description]
IV&V Control Point: [Validation point description]
"""
```

## Template Usage Guidelines

### 1. Copy and Customize
- Copy the appropriate template for your test type
- Replace placeholder content with your specific logic
- Update requirements traceability to match your module
- Add real component integration where applicable

### 2. Requirements Traceability
- Every test must have requirements listed in docstring
- Use correct REQ-[MODULE]-[NUMBER] format
- Include error requirements for error condition tests
- Map integration requirements for end-to-end tests

### 3. Real Component Integration
- Use MediaMTX test infrastructure for MediaMTX integration
- Use WebSocket test client for WebSocket communication
- Use real file system operations with temporary directories
- Use real hardware integration where applicable

### 4. Error Condition Testing
- Include comprehensive error condition testing
- Test service failures, network failures, and resource constraints
- Validate graceful error handling and recovery
- Test boundary conditions and edge cases

### 5. Test Naming
- Use descriptive names that describe specific behaviors
- Include scenario information in test names
- Follow the pattern: `test_[behavior]_[scenario]()`

## Module-Specific Templates

### WebSocket Server Tests
```python
# Use unit_test_template.py with WebSocket patterns
from tests.templates.unit_test_template import TestYourComponent
from tests.fixtures.websocket_test_client import WebSocketTestClient

class TestWebSocketServer(TestYourComponent):
    # Override with WebSocket-specific fixtures
    @pytest.fixture
    def websocket_client(self):
        return WebSocketTestClient("ws://localhost:8002/ws")
```

### MediaMTX Integration Tests
```python
# Use integration_test_template.py with MediaMTX patterns
from tests.templates.integration_test_template import TestSystemIntegration
from tests.fixtures.mediamtx_test_infrastructure import mediamtx_infrastructure

class TestMediaMTXIntegration(TestSystemIntegration):
    # Override with MediaMTX-specific configuration
    @pytest.fixture
    def integration_config(self, temp_test_directory):
        return MediaMTXConfig(...)
```

### Camera Discovery Tests
```python
# Use unit_test_template.py with hardware integration patterns
from tests.templates.unit_test_template import TestYourComponent

class TestCameraDiscovery(TestYourComponent):
    # Add hardware-specific fixtures and patterns
    @pytest.fixture
    def camera_device(self):
        return "/dev/video0"
```

## Validation Checklist

### Before Using Templates
- [ ] Understand the test guidelines and requirements
- [ ] Identify the appropriate template for your test type
- [ ] Review the requirements mapping guide
- [ ] Understand real component integration patterns

### After Using Templates
- [ ] All tests have requirements traceability
- [ ] Real component integration is implemented
- [ ] Error condition testing is comprehensive
- [ ] Test names are descriptive and specific
- [ ] Documentation is complete and accurate

### Template Compliance
- [ ] Templates follow all test guidelines
- [ ] Real component usage is demonstrated
- [ ] Requirements traceability examples are included
- [ ] Error condition testing patterns are comprehensive
- [ ] Documentation explains how to apply templates

## Best Practices

### 1. Template Selection
- Choose the most appropriate template for your test type
- Unit tests: Use `unit_test_template.py`
- Integration tests: Use `integration_test_template.py`
- Error testing: Use `error_testing_patterns.py`
- Edge cases: Use `edge_case_patterns.py`

### 2. Customization
- Replace placeholder content with your specific logic
- Update requirements traceability to match your module
- Add real component integration where applicable
- Include comprehensive error condition testing

### 3. Documentation
- Keep requirements descriptions clear and concise
- Update requirements when they change
- Maintain requirements traceability matrix
- Document any deviations from templates

### 4. Validation
- Validate that templates follow all guidelines
- Ensure real component integration is implemented
- Verify requirements traceability is complete
- Test error condition coverage is comprehensive

## Examples

### Example 1: WebSocket Server Unit Test
```python
from tests.templates.unit_test_template import TestYourComponent
from tests.fixtures.websocket_test_client import WebSocketTestClient

class TestWebSocketServer(TestYourComponent):
    """
    Test WebSocket server functionality with real component integration.
    
    Requirements Traceability:
    - REQ-WS-001: WebSocket server shall aggregate camera status
    - REQ-WS-002: WebSocket server shall provide capability metadata
    - REQ-ERROR-001: WebSocket server shall handle connection failures
    
    Story Coverage: S3 - WebSocket API Integration
    IV&V Control Point: Real WebSocket communication validation
    """
    
    @pytest.fixture
    def component(self, test_config):
        return WebSocketJsonRpcServer(test_config)
    
    @pytest.mark.asyncio
    async def test_camera_status_aggregation_with_real_mediamtx_integration(
        self, component, mediamtx_controller, mediamtx_infrastructure
    ):
        """
        Test camera status aggregation with real MediaMTX integration.
        
        Requirements: REQ-WS-001, REQ-WS-002
        Scenario: Real MediaMTX integration with capability metadata
        Expected: Successful integration with real MediaMTX service
        Edge Cases: Real stream status queries, actual metrics retrieval
        """
        # Your specific test logic here
        pass
```

### Example 2: Integration Test
```python
from tests.templates.integration_test_template import TestSystemIntegration

class TestCameraServiceIntegration(TestSystemIntegration):
    """
    Test camera service integration with real components.
    
    Requirements Traceability:
    - REQ-INT-001: System shall integrate all components seamlessly
    - REQ-INT-002: System shall handle real MediaMTX service integration
    - REQ-INT-003: System shall support real WebSocket communication
    
    Story Coverage: S4 - System Integration
    IV&V Control Point: End-to-end system validation
    """
    
    @pytest.mark.asyncio
    async def test_end_to_end_camera_streaming_workflow(
        self, service_manager, websocket_client, mediamtx_infrastructure
    ):
        """
        Test complete end-to-end camera streaming workflow.
        
        Requirements: REQ-INT-001, REQ-INT-002, REQ-INT-003
        Scenario: Complete camera discovery → stream creation → WebSocket communication
        Expected: Successful end-to-end workflow with real components
        Edge Cases: Real service interactions, actual data flow, resource management
        """
        # Your specific integration test logic here
        pass
```

## Conclusion

These templates and patterns provide a solid foundation for creating compliant tests that:
- Follow all test guidelines
- Include real component integration
- Have proper requirements traceability
- Cover comprehensive error conditions
- Use descriptive naming and documentation

Use these templates consistently to prevent future violations and maintain high-quality test coverage.
