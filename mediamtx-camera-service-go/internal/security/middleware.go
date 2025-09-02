package security

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

// ClientConnection represents a client connection interface
type ClientConnection interface {
	GetClientID() string
	GetUserID() string
	GetRole() string
	IsAuthenticated() bool
}

// JsonRpcResponse represents a JSON-RPC response interface
type JsonRpcResponse interface {
	GetJSONRPC() string
	GetResult() interface{}
	GetError() JsonRpcError
	GetID() interface{}
}

// JsonRpcError represents a JSON-RPC error interface
type JsonRpcError interface {
	GetCode() int
	GetMessage() string
	GetData() interface{}
}

// SecurityConfig represents security configuration interface
type SecurityConfig interface {
	GetRateLimitRequests() int
	GetRateLimitWindow() interface{}
	GetJWTSecretKey() string
	GetJWTExpiryHours() int
}

// MethodHandler represents a method handler function
type MethodHandler func(params map[string]interface{}, client ClientConnection) (JsonRpcResponse, error)

// AuthMiddleware provides centralized authentication enforcement
type AuthMiddleware struct {
	logger *logrus.Logger
	config SecurityConfig
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(logger *logrus.Logger, securityConfig SecurityConfig) *AuthMiddleware {
	return &AuthMiddleware{
		logger: logger,
		config: securityConfig,
	}
}

// RequireAuth decorates a method handler to require authentication
func (am *AuthMiddleware) RequireAuth(handler MethodHandler) MethodHandler {
	return func(params map[string]interface{}, client ClientConnection) (JsonRpcResponse, error) {
		if !client.IsAuthenticated() {
			am.logger.WithFields(logrus.Fields{
				"client_id": client.GetClientID(),
				"method":    "authentication_required",
				"action":    "auth_bypass_attempt",
			}).Warn("Authentication bypass attempt blocked")

			// Return error response (implementation will need to create concrete type)
			return nil, fmt.Errorf("authentication required")
		}

		am.logger.WithFields(logrus.Fields{
			"client_id": client.GetClientID(),
			"user_id":   client.GetUserID(),
			"role":      client.GetRole(),
			"method":    "authentication_success",
			"action":    "auth_check_passed",
		}).Debug("Authentication check passed")

		return handler(params, client)
	}
}

// RBACMiddleware provides centralized role-based access control
type RBACMiddleware struct {
	permissionChecker *PermissionChecker
	logger            *logrus.Logger
	config            SecurityConfig
}

// NewRBACMiddleware creates a new RBAC middleware
func NewRBACMiddleware(permissionChecker *PermissionChecker, logger *logrus.Logger, securityConfig SecurityConfig) *RBACMiddleware {
	return &RBACMiddleware{
		permissionChecker: permissionChecker,
		logger:            logger,
		config:            securityConfig,
	}
}

// RequireRole decorates a method handler to require a specific role
func (rm *RBACMiddleware) RequireRole(requiredRole Role, handler MethodHandler) MethodHandler {
	return func(params map[string]interface{}, client ClientConnection) (JsonRpcResponse, error) {
		userRole, err := rm.permissionChecker.ValidateRole(client.GetRole())
		if err != nil {
			rm.logger.WithFields(logrus.Fields{
				"client_id": client.GetClientID(),
				"role":      client.GetRole(),
				"error":     err.Error(),
				"action":    "role_validation_failed",
			}).Error("Role validation failed")

			return nil, fmt.Errorf("insufficient permissions: invalid role %s", client.GetRole())
		}

		if userRole < requiredRole {
			rm.logger.WithFields(logrus.Fields{
				"client_id":     client.GetClientID(),
				"user_role":     userRole.String(),
				"required_role": requiredRole.String(),
				"action":        "permission_denied",
			}).Warn("Permission denied - insufficient role")

			return nil, fmt.Errorf("insufficient permissions: required role %s, user role %s", requiredRole.String(), userRole.String())
		}

		rm.logger.WithFields(logrus.Fields{
			"client_id":     client.GetClientID(),
			"user_role":     userRole.String(),
			"required_role": requiredRole.String(),
			"action":        "permission_granted",
		}).Debug("Permission check passed")

		return handler(params, client)
	}
}

// SecureMethodRegistry provides secure method registration with automatic security enforcement
type SecureMethodRegistry struct {
	methods map[string]MethodHandler
	auth    *AuthMiddleware
	rbac    *RBACMiddleware
	logger  *logrus.Logger
	config  SecurityConfig
}

// NewSecureMethodRegistry creates a new secure method registry
func NewSecureMethodRegistry(auth *AuthMiddleware, rbac *RBACMiddleware, logger *logrus.Logger, securityConfig SecurityConfig) *SecureMethodRegistry {
	return &SecureMethodRegistry{
		methods: make(map[string]MethodHandler),
		auth:    auth,
		rbac:    rbac,
		logger:  logger,
		config:  securityConfig,
	}
}

// RegisterMethod registers a method with automatic security enforcement
func (smr *SecureMethodRegistry) RegisterMethod(methodName string, handler MethodHandler, requiredRole Role) {
	// Apply security decorators in order: authentication first, then role-based access control
	securedHandler := smr.auth.RequireAuth(handler)

	// Only apply RBAC if role requirement is higher than viewer
	if requiredRole > RoleViewer {
		securedHandler = smr.rbac.RequireRole(requiredRole, securedHandler)
	}

	smr.methods[methodName] = securedHandler

	smr.logger.WithFields(logrus.Fields{
		"method":        methodName,
		"required_role": requiredRole.String(),
		"action":        "method_registered",
	}).Info("Method registered with security enforcement")
}

// GetMethod retrieves a secured method handler
func (smr *SecureMethodRegistry) GetMethod(methodName string) (MethodHandler, bool) {
	handler, exists := smr.methods[methodName]
	return handler, exists
}

// GetAllMethods returns all registered method names for auditing
func (smr *SecureMethodRegistry) GetAllMethods() []string {
	methods := make([]string, 0, len(smr.methods))
	for method := range smr.methods {
		methods = append(methods, method)
	}
	return methods
}

// GetMethodSecurityInfo returns security information for a method
func (smr *SecureMethodRegistry) GetMethodSecurityInfo(methodName string) map[string]interface{} {
	// This would return detailed security information for auditing
	// For now, return basic info
	return map[string]interface{}{
		"method":         methodName,
		"secured":        true,
		"authentication": "required",
		"rbac_enabled":   true,
	}
}
