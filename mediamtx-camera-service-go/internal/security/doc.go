// Package security provides authentication, authorization, and security middleware for the MediaMTX Camera Service.
//
// This package implements JWT-based authentication, role-based access control (RBAC),
// rate limiting, and input validation to secure the WebSocket API endpoints.
//
// Architecture Compliance:
//   - JWT Authentication: HS256 algorithm with configurable expiry
//   - Role-Based Access Control: viewer, operator, admin roles with permission checking
//   - Rate Limiting: Request throttling with sliding window implementation
//   - Input Validation: Sanitization and validation of user inputs
//   - Security Middleware: Decorator pattern for method-level security enforcement
//
// Key Components:
//   - JWTHandler: Token generation, validation, and rate limiting
//   - PermissionChecker: Role-based permission validation
//   - AuthMiddleware: Authentication requirement enforcement
//   - RBACMiddleware: Role-based access control enforcement
//   - RateLimiter: Request throttling and abuse prevention
//   - InputValidator: Input sanitization and validation
//   - SessionManager: Client session lifecycle management
//
// Security Features:
//   - JWT token generation and validation with configurable expiry
//   - Role-based access control with granular permissions
//   - Rate limiting with sliding window algorithm
//   - Input sanitization and validation
//   - CORS configuration support
//   - Session management with automatic cleanup
//
// Usage Pattern:
//   - Create JWTHandler with NewJWTHandler()
//   - Generate tokens with GenerateToken()
//   - Validate tokens with ValidateToken()
//   - Apply middleware with RequireAuth() and RequireRole()
//   - Check permissions with CheckPermission()
//
// Requirements Coverage:
//   - REQ-SEC-001: JWT authentication with configurable expiry
//   - REQ-SEC-002: Role-based access control (viewer, operator, admin)
//   - REQ-SEC-003: Rate limiting with sliding window
//   - REQ-SEC-004: Input validation and sanitization
//   - REQ-SEC-005: CORS configuration support
//
// Test Categories: Unit/Integration
// API Documentation Reference: docs/security.md
package security
