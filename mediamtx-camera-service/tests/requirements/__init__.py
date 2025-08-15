"""
Requirements Test Package

This package contains comprehensive test coverage for all critical missing requirements:
- Performance Requirements (REQ-PERF-001 through REQ-PERF-004)
- Health Monitoring Requirements (REQ-HEALTH-001 through REQ-HEALTH-003)
- Configuration Management Requirements (REQ-CONFIG-002, REQ-CONFIG-003)
- Error Handling Requirements (REQ-ERROR-004 through REQ-ERROR-008)

Total: 14 critical missing requirements with comprehensive test coverage.
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
