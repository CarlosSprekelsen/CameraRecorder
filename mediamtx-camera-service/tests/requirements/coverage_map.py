"""
Requirements Coverage Mapping

Maps requirements to test files and validates adequate coverage.
Ensures every requirement has proper test coverage with real system integration.

Requirements Traceability:
- REQ-COV-001: All requirements shall have adequate test coverage
- REQ-COV-002: Test coverage shall include real system integration
- REQ-COV-003: Test coverage shall include edge cases and error scenarios
- REQ-COV-004: Coverage gaps shall be identified and addressed

Test Categories: Requirements Validation
"""

from typing import Dict, List, Any, Optional
from enum import Enum


class CoverageType(Enum):
    """Coverage type enumeration."""
    NONE = "NONE"
    PARTIAL = "PARTIAL"
    ADEQUATE = "ADEQUATE"
    COMPREHENSIVE = "COMPREHENSIVE"


class RequirementStatus(Enum):
    """Requirement status enumeration."""
    ORPHANED = "ORPHANED"
    NEEDS_ENHANCEMENT = "NEEDS_ENHANCEMENT"
    COMPLETE = "COMPLETE"


# Requirements coverage mapping
REQUIREMENTS_COVERAGE = {
    # WebSocket Requirements
    "REQ-WS-001": {
        "description": "WebSocket server shall aggregate camera status with real MediaMTX integration",
        "test_files": ["test_websocket_real_integration.py"],
        "test_methods": [
            "test_real_websocket_connection_lifecycle",
            "test_real_notification_delivery",
            "test_real_client_disconnection_scenarios"
        ],
        "edge_cases": [
            "MediaMTX service unavailable",
            "Multiple camera connections",
            "Client disconnection during status update"
        ],
        "coverage_type": CoverageType.ADEQUATE,
        "status": RequirementStatus.COMPLETE
    },
    "REQ-WS-002": {
        "description": "WebSocket server shall provide camera capability metadata integration",
        "test_files": ["test_websocket_real_integration.py"],
        "test_methods": [
            "test_real_camera_capability_metadata"
        ],
        "edge_cases": [
            "Camera capability detection failure",
            "Metadata format validation"
        ],
        "coverage_type": CoverageType.ADEQUATE,
        "status": RequirementStatus.COMPLETE
    },
    
    # MediaMTX Requirements
    "REQ-MEDIA-002": {
        "description": "Stream management and recording control",
        "test_files": ["test_mediamtx_real_integration.py"],
        "test_methods": [
            "test_real_stream_creation_and_management",
            "test_real_recording_operations",
            "test_real_snapshot_capture"
        ],
        "edge_cases": [
            "Stream creation failure",
            "Recording timeout",
            "Disk space exhaustion",
            "FFmpeg process failure"
        ],
        "coverage_type": CoverageType.ADEQUATE,
        "status": RequirementStatus.COMPLETE
    },
    "REQ-MEDIA-005": {
        "description": "Stream lifecycle management",
        "test_files": ["test_mediamtx_real_integration.py"],
        "test_methods": [
            "test_real_stream_lifecycle_management"
        ],
        "edge_cases": [
            "Stream cleanup failure",
            "Resource cleanup validation"
        ],
        "coverage_type": CoverageType.ADEQUATE,
        "status": RequirementStatus.COMPLETE
    },
    
    # Integration Requirements
    "REQ-INT-001": {
        "description": "Integration system shall provide real end-to-end system behavior validation",
        "test_files": ["test_system_real_integration.py"],
        "test_methods": [
            "test_real_end_to_end_camera_workflow",
            "test_real_cross_component_data_flow"
        ],
        "edge_cases": [
            "Component communication failure",
            "Data flow interruption"
        ],
        "coverage_type": CoverageType.ADEQUATE,
        "status": RequirementStatus.COMPLETE
    },
    "REQ-INT-002": {
        "description": "Integration system shall validate real MediaMTX server integration",
        "test_files": ["test_mediamtx_real_integration.py"],
        "test_methods": [
            "test_real_mediamtx_server_integration"
        ],
        "edge_cases": [
            "MediaMTX service failure",
            "API endpoint unavailability"
        ],
        "coverage_type": CoverageType.ADEQUATE,
        "status": RequirementStatus.COMPLETE
    },
    
    # Error Handling Requirements
    "REQ-ERROR-001": {
        "description": "WebSocket server shall handle MediaMTX connection failures gracefully",
        "test_files": ["test_error_handling_real.py"],
        "test_methods": [
            "test_real_mediamtx_connection_failures"
        ],
        "edge_cases": [
            "Network timeout",
            "Service unavailability",
            "Connection recovery"
        ],
        "coverage_type": CoverageType.ADEQUATE,
        "status": RequirementStatus.COMPLETE
    },
    "REQ-ERROR-002": {
        "description": "WebSocket server shall handle client disconnection gracefully",
        "test_files": ["test_websocket_real_integration.py"],
        "test_methods": [
            "test_real_client_disconnection_scenarios"
        ],
        "edge_cases": [
            "Abrupt disconnection",
            "Network failure disconnection",
            "Multiple client disconnections"
        ],
        "coverage_type": CoverageType.ADEQUATE,
        "status": RequirementStatus.COMPLETE
    },
    
    # Authentication Requirements
    "REQ-AUTH-001": {
        "description": "System shall provide secure authentication using JWT tokens",
        "test_files": ["test_authentication_real.py"],
        "test_methods": [
            "test_real_jwt_authentication",
            "test_real_token_validation"
        ],
        "edge_cases": [
            "Invalid token format",
            "Expired tokens",
            "Tampered tokens"
        ],
        "coverage_type": CoverageType.ADEQUATE,
        "status": RequirementStatus.COMPLETE
    },
    "REQ-AUTH-004": {
        "description": "System shall support role-based access control",
        "test_files": ["test_authentication_real.py"],
        "test_methods": [
            "test_real_role_based_access_control"
        ],
        "edge_cases": [
            "Invalid role access",
            "Role escalation attempts"
        ],
        "coverage_type": CoverageType.ADEQUATE,
        "status": RequirementStatus.COMPLETE
    }
}


class RequirementsCoverageValidator:
    """Validates requirements coverage and identifies gaps."""
    
    def __init__(self):
        self.coverage_map = REQUIREMENTS_COVERAGE
    
    def get_coverage_summary(self) -> Dict[str, Any]:
        """Get overall coverage summary."""
        total_requirements = len(self.coverage_map)
        covered_requirements = 0
        adequate_coverage = 0
        comprehensive_coverage = 0
        
        for req_id, coverage in self.coverage_map.items():
            if coverage["coverage_type"] != CoverageType.NONE:
                covered_requirements += 1
            
            if coverage["coverage_type"] in [CoverageType.ADEQUATE, CoverageType.COMPREHENSIVE]:
                adequate_coverage += 1
            
            if coverage["coverage_type"] == CoverageType.COMPREHENSIVE:
                comprehensive_coverage += 1
        
        return {
            "total_requirements": total_requirements,
            "covered_requirements": covered_requirements,
            "adequate_coverage": adequate_coverage,
            "comprehensive_coverage": comprehensive_coverage,
            "coverage_percentage": (adequate_coverage / total_requirements) * 100 if total_requirements > 0 else 0
        }
    
    def identify_coverage_gaps(self) -> List[Dict[str, Any]]:
        """Identify requirements with inadequate coverage."""
        gaps = []
        
        for req_id, coverage in self.coverage_map.items():
            if coverage["coverage_type"] in [CoverageType.NONE, CoverageType.PARTIAL]:
                gaps.append({
                    "requirement_id": req_id,
                    "description": coverage["description"],
                    "current_coverage": coverage["coverage_type"].value,
                    "status": coverage["status"].value
                })
        
        return gaps
    
    def validate_requirement_coverage(self, req_id: str) -> Dict[str, Any]:
        """Validate coverage for a specific requirement."""
        if req_id not in self.coverage_map:
            return {
                "valid": False,
                "error": f"Requirement {req_id} not found in coverage map"
            }
        
        coverage = self.coverage_map[req_id]
        
        # Check if requirement has adequate coverage
        has_adequate_coverage = coverage["coverage_type"] in [CoverageType.ADEQUATE, CoverageType.COMPREHENSIVE]
        
        # Check if requirement has real system integration
        has_real_integration = any("real" in test_file for test_file in coverage["test_files"])
        
        # Check if requirement has edge case coverage
        has_edge_cases = len(coverage["edge_cases"]) > 0
        
        return {
            "valid": has_adequate_coverage and has_real_integration and has_edge_cases,
            "requirement_id": req_id,
            "description": coverage["description"],
            "coverage_type": coverage["coverage_type"].value,
            "status": coverage["status"].value,
            "has_adequate_coverage": has_adequate_coverage,
            "has_real_integration": has_real_integration,
            "has_edge_cases": has_edge_cases,
            "test_files": coverage["test_files"],
            "edge_cases": coverage["edge_cases"]
        }
    
    def get_requirements_by_status(self, status: RequirementStatus) -> List[str]:
        """Get requirements by status."""
        return [
            req_id for req_id, coverage in self.coverage_map.items()
            if coverage["status"] == status
        ]
    
    def get_requirements_by_coverage_type(self, coverage_type: CoverageType) -> List[str]:
        """Get requirements by coverage type."""
        return [
            req_id for req_id, coverage in self.coverage_map.items()
            if coverage["coverage_type"] == coverage_type
        ]
