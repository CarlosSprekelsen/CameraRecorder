"""
Security test package for MediaMTX Camera Service.

Requirements Coverage:
- REQ-SEC-001: JWT token-based authentication for all API access
- REQ-SEC-002: Token Format: JSON Web Token (JWT) with standard claims
- REQ-SEC-003: Token Expiration: Configurable expiration time (default: 24 hours)
- REQ-SEC-004: Token Refresh: Support for token refresh mechanism
- REQ-SEC-005: Token Validation: Proper signature validation and claim verification
- REQ-SEC-006: API key validation for service-to-service communication
- REQ-SEC-007: API Key Format: Secure random string (32+ characters)
- REQ-SEC-008: Key Storage: Secure storage of API keys
- REQ-SEC-009: Key Rotation: Support for API key rotation
- REQ-SEC-010: Role-based access control for different user types
- REQ-SEC-011: User Roles: Admin, User, Read-Only roles
- REQ-SEC-012: Permission Matrix: Clear permission definitions for each role
- REQ-SEC-013: Access Control: Enforcement of role-based permissions
- REQ-SEC-019: Sanitize and validate all input data
- REQ-SEC-020: Input Validation: Comprehensive validation of all input parameters
- REQ-SEC-021: Sanitization: Proper sanitization of user input
- REQ-SEC-022: Injection Prevention: Prevention of SQL injection, XSS, and command injection
- REQ-SEC-023: Parameter Validation: Validation of parameter types and ranges

Test Categories: Unit
"""
