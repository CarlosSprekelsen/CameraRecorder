"""
Requirements test package for MediaMTX Camera Service.

Requirements Traceability:
- REQ-UTIL-026: Requirements test package shall provide requirement validation testing
- REQ-REQ-001: Requirements tests shall validate requirement completeness
- REQ-REQ-002: Requirements tests shall validate requirement consistency
- REQ-REQ-003: Requirements tests shall validate requirement testability
- REQ-REQ-004: Requirements tests shall validate requirement traceability
- REQ-REQ-005: Requirements tests shall validate requirement coverage

Story Coverage: S16 - Requirements Validation
IV&V Control Point: Requirements test validation and requirement quality assurance

This package provides:
1. Requirement validation testing
2. Requirement completeness validation
3. Requirement consistency testing
4. Requirement testability validation
5. Requirement traceability testing
6. Requirement coverage validation
"""

from .test_performance_requirements import TestPerformanceRequirements, PerformanceRequirementsValidator
from .test_health_monitoring_requirements import TestHealthMonitoringRequirements, HealthMonitoringRequirementsValidator
from .test_configuration_requirements import TestConfigurationRequirements, ConfigurationRequirementsValidator
from .test_error_handling_requirements import TestErrorHandlingRequirements, ErrorHandlingRequirementsValidator
from .test_all_requirements import TestAllRequirements, ComprehensiveRequirementsValidator

__all__ = [
    # Performance Requirements
    "TestPerformanceRequirements",
    "PerformanceRequirementsValidator",
    
    # Health Monitoring Requirements
    "TestHealthMonitoringRequirements", 
    "HealthMonitoringRequirementsValidator",
    
    # Configuration Management Requirements
    "TestConfigurationRequirements",
    "ConfigurationRequirementsValidator",
    
    # Error Handling Requirements
    "TestErrorHandlingRequirements",
    "ErrorHandlingRequirementsValidator",
    
    # Comprehensive Requirements
    "TestAllRequirements",
    "ComprehensiveRequirementsValidator"
]

# Requirements coverage summary
REQUIREMENTS_COVERAGE = {
    "performance": ["REQ-PERF-001", "REQ-PERF-002", "REQ-PERF-003", "REQ-PERF-004"],
    "health_monitoring": ["REQ-HEALTH-001", "REQ-HEALTH-002", "REQ-HEALTH-003"],
    "configuration": ["REQ-CONFIG-002", "REQ-CONFIG-003"],
    "error_handling": ["REQ-ERROR-004", "REQ-ERROR-005", "REQ-ERROR-006", "REQ-ERROR-007", "REQ-ERROR-008"]
}

TOTAL_REQUIREMENTS = 14
