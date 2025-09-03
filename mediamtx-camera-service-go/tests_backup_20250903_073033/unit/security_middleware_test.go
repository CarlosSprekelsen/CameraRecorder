//go:build unit
// +build unit

package security

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockClientConnection implements ClientConnection interface for testing
type MockClientConnection struct {
	clientID      string
	userID        string
	role          string
	authenticated bool
}

func (m *MockClientConnection) GetClientID() string   { return m.clientID }
func (m *MockClientConnection) GetUserID() string     { return m.userID }
func (m *MockClientConnection) GetRole() string       { return m.role }
func (m *MockClientConnection) IsAuthenticated() bool { return m.authenticated }

// MockJsonRpcResponse implements JsonRpcResponse interface for testing
type MockJsonRpcResponse struct {
	jsonrpc string
	result  interface{}
	error   security.JsonRpcError
	id      interface{}
}

func (m *MockJsonRpcResponse) GetJSONRPC() string              { return m.jsonrpc }
func (m *MockJsonRpcResponse) GetResult() interface{}          { return m.result }
func (m *MockJsonRpcResponse) GetError() security.JsonRpcError { return m.error }
func (m *MockJsonRpcResponse) GetID() interface{}              { return m.id }

// MockJsonRpcError implements JsonRpcError interface for testing
type MockJsonRpcError struct {
	code    int
	message string
	data    interface{}
}

func (m *MockJsonRpcError) GetCode() int         { return m.code }
func (m *MockJsonRpcError) GetMessage() string   { return m.message }
func (m *MockJsonRpcError) GetData() interface{} { return m.data }

// MockSecurityConfig implements SecurityConfig interface for testing
type MockSecurityConfig struct {
	rateLimitRequests int
	rateLimitWindow   interface{}
	jwtSecretKey      string
	jwtExpiryHours    int
}

func (m *MockSecurityConfig) GetRateLimitRequests() int       { return m.rateLimitRequests }
func (m *MockSecurityConfig) GetRateLimitWindow() interface{} { return m.rateLimitWindow }
func (m *MockSecurityConfig) GetJWTSecretKey() string         { return m.jwtSecretKey }
func (m *MockSecurityConfig) GetJWTExpiryHours() int          { return m.jwtExpiryHours }

func TestNewAuthMiddleware(t *testing.T) {
	// COMMON PATTERN: Use shared test environment instead of individual components
	env := testtestutils.SetupTestEnvironment(t)
	defer testtestutils.TeardownTestEnvironment(t, env)

	config := &MockSecurityConfig{}

	middleware := NewAuthMiddleware(env.Logger, config)

	assert.NotNil(t, middleware)
	// Note: Fields are unexported, so we can't test them directly
	// This is intentional for encapsulation
}

func TestAuthMiddleware_RequireAuth_Authenticated(t *testing.T) {
	// COMMON PATTERN: Use shared test environment instead of individual components
	env := testtestutils.SetupTestEnvironment(t)
	defer testtestutils.TeardownTestEnvironment(t, env)

	config := &MockSecurityConfig{}
	middleware := NewAuthMiddleware(env.Logger, config)

	// Mock authenticated client
	client := &MockClientConnection{
		clientID:      "test_client",
		userID:        "test_user",
		role:          "viewer",
		authenticated: true,
	}

	// Mock handler that should be called
	handlerCalled := false
	handler := func(params map[string]interface{}, client security.ClientConnection) (security.JsonRpcResponse, error) {
		handlerCalled = true
		return &MockJsonRpcResponse{result: "success"}, nil
	}

	// Apply authentication middleware
	securedHandler := middleware.RequireAuth(handler)

	// Call the secured handler
	response, err := securedHandler(map[string]interface{}{}, client)

	// Verify handler was called
	assert.True(t, handlerCalled, "Handler should be called for authenticated client")
	assert.NoError(t, err)
	assert.NotNil(t, response)
}

func TestAuthMiddleware_RequireAuth_NotAuthenticated(t *testing.T) {
	// COMMON PATTERN: Use shared test environment instead of individual components
	env := testtestutils.SetupTestEnvironment(t)
	defer testtestutils.TeardownTestEnvironment(t, env)

	config := &MockSecurityConfig{}
	middleware := NewAuthMiddleware(env.Logger, config)

	// Mock unauthenticated client
	client := &MockClientConnection{
		clientID:      "test_client",
		userID:        "",
		role:          "",
		authenticated: false,
	}

	// Mock handler that should NOT be called
	handlerCalled := false
	handler := func(params map[string]interface{}, client security.ClientConnection) (security.JsonRpcResponse, error) {
		handlerCalled = true
		return &MockJsonRpcResponse{result: "success"}, nil
	}

	// Apply authentication middleware
	securedHandler := middleware.RequireAuth(handler)

	// Call the secured handler
	response, err := securedHandler(map[string]interface{}{}, client)

	// Verify handler was NOT called
	assert.False(t, handlerCalled, "Handler should not be called for unauthenticated client")
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "authentication required")
}

func TestNewRBACMiddleware(t *testing.T) {
	// COMMON PATTERN: Use shared test environment instead of individual components
	env := testtestutils.SetupTestEnvironment(t)
	defer testtestutils.TeardownTestEnvironment(t, env)

	config := &MockSecurityConfig{}
	permissionChecker := NewPermissionChecker()

	middleware := NewRBACMiddleware(permissionChecker, env.Logger, config)

	assert.NotNil(t, middleware)
	// Note: Fields are unexported, so we can't test them directly
	// This is intentional for encapsulation
}

func TestRBACMiddleware_RequireRole_SufficientRole(t *testing.T) {
	// COMMON PATTERN: Use shared test environment instead of individual components
	env := testtestutils.SetupTestEnvironment(t)
	defer testtestutils.TeardownTestEnvironment(t, env)

	config := &MockSecurityConfig{}
	permissionChecker := NewPermissionChecker()
	middleware := NewRBACMiddleware(permissionChecker, env.Logger, config)

	// Mock client with operator role
	client := &MockClientConnection{
		clientID:      "test_client",
		userID:        "test_user",
		role:          "operator",
		authenticated: true,
	}

	// Mock handler that should be called
	handlerCalled := false
	handler := func(params map[string]interface{}, client security.ClientConnection) (security.JsonRpcResponse, error) {
		handlerCalled = true
		return &MockJsonRpcResponse{result: "success"}, nil
	}

	// Apply RBAC middleware requiring operator role
	securedHandler := middleware.RequireRole(security.RoleOperator, handler)

	// Call the secured handler
	response, err := securedHandler(map[string]interface{}{}, client)

	// Verify handler was called
	assert.True(t, handlerCalled, "Handler should be called for client with sufficient role")
	assert.NoError(t, err)
	assert.NotNil(t, response)
}

func TestRBACMiddleware_RequireRole_InsufficientRole(t *testing.T) {
	// COMMON PATTERN: Use shared test environment instead of individual components
	env := testtestutils.SetupTestEnvironment(t)
	defer testtestutils.TeardownTestEnvironment(t, env)

	config := &MockSecurityConfig{}
	permissionChecker := NewPermissionChecker()
	middleware := NewRBACMiddleware(permissionChecker, env.Logger, config)

	// Mock client with viewer role
	client := &MockClientConnection{
		clientID:      "test_client",
		userID:        "test_user",
		role:          "viewer",
		authenticated: true,
	}

	// Mock handler that should NOT be called
	handlerCalled := false
	handler := func(params map[string]interface{}, client security.ClientConnection) (security.JsonRpcResponse, error) {
		handlerCalled = true
		return &MockJsonRpcResponse{result: "success"}, nil
	}

	// Apply RBAC middleware requiring admin role
	securedHandler := middleware.RequireRole(security.RoleAdmin, handler)

	// Call the secured handler
	response, err := securedHandler(map[string]interface{}{}, client)

	// Verify handler was NOT called
	assert.False(t, handlerCalled, "Handler should not be called for client with insufficient role")
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "insufficient permissions")
}

func TestRBACMiddleware_RequireRole_InvalidRole(t *testing.T) {
	// COMMON PATTERN: Use shared test environment instead of individual components
	env := testtestutils.SetupTestEnvironment(t)
	defer testtestutils.TeardownTestEnvironment(t, env)

	config := &MockSecurityConfig{}
	permissionChecker := NewPermissionChecker()
	middleware := NewRBACMiddleware(permissionChecker, env.Logger, config)

	// Mock client with invalid role
	client := &MockClientConnection{
		clientID:      "test_client",
		userID:        "test_user",
		role:          "invalid_role",
		authenticated: true,
	}

	// Mock handler that should NOT be called
	handlerCalled := false
	handler := func(params map[string]interface{}, client security.ClientConnection) (security.JsonRpcResponse, error) {
		handlerCalled = true
		return &MockJsonRpcResponse{result: "success"}, nil
	}

	// Apply RBAC middleware requiring viewer role
	securedHandler := middleware.RequireRole(security.RoleViewer, handler)

	// Call the secured handler
	response, err := securedHandler(map[string]interface{}{}, client)

	// Verify handler was NOT called
	assert.False(t, handlerCalled, "Handler should not be called for client with invalid role")
	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "insufficient permissions")
}

func TestNewSecureMethodRegistry(t *testing.T) {
	// COMMON PATTERN: Use shared test environment instead of individual components
	env := testtestutils.SetupTestEnvironment(t)
	defer testtestutils.TeardownTestEnvironment(t, env)

	config := &MockSecurityConfig{}
	permissionChecker := NewPermissionChecker()
	authMiddleware := NewAuthMiddleware(env.Logger, config)
	rbacMiddleware := NewRBACMiddleware(permissionChecker, env.Logger, config)

	registry := NewSecureMethodRegistry(authMiddleware, rbacMiddleware, env.Logger, config)

	assert.NotNil(t, registry)
	// Note: Fields are unexported, so we can't test them directly
	// This is intentional for encapsulation
}

func TestSecureMethodRegistry_RegisterMethod(t *testing.T) {
	// COMMON PATTERN: Use shared test environment instead of individual components
	env := testtestutils.SetupTestEnvironment(t)
	defer testtestutils.TeardownTestEnvironment(t, env)

	config := &MockSecurityConfig{}
	permissionChecker := NewPermissionChecker()
	authMiddleware := NewAuthMiddleware(env.Logger, config)
	rbacMiddleware := NewRBACMiddleware(permissionChecker, env.Logger, config)

	registry := NewSecureMethodRegistry(authMiddleware, rbacMiddleware, env.Logger, config)

	// Mock handler
	handlerCalled := false
	handler := func(params map[string]interface{}, client security.ClientConnection) (security.JsonRpcResponse, error) {
		handlerCalled = true
		return &MockJsonRpcResponse{result: "success"}, nil
	}

	// Register method with viewer role
	registry.RegisterMethod("test_method", handler, security.RoleViewer)

	// Verify method was registered
	registeredHandler, exists := registry.GetMethod("test_method")
	assert.True(t, exists, "Method should be registered")
	assert.NotNil(t, registeredHandler)

	// Test that the method works with authenticated client
	client := &MockClientConnection{
		clientID:      "test_client",
		userID:        "test_user",
		role:          "viewer",
		authenticated: true,
	}

	response, err := registeredHandler(map[string]interface{}{}, client)
	assert.NoError(t, err)
	assert.True(t, handlerCalled, "Handler should be called")
	assert.NotNil(t, response)
}

func TestSecureMethodRegistry_RegisterMethod_AdminRole(t *testing.T) {
	// COMMON PATTERN: Use shared test environment instead of individual components
	env := testtestutils.SetupTestEnvironment(t)
	defer testtestutils.TeardownTestEnvironment(t, env)

	config := &MockSecurityConfig{}
	permissionChecker := NewPermissionChecker()
	authMiddleware := NewAuthMiddleware(env.Logger, config)
	rbacMiddleware := NewRBACMiddleware(permissionChecker, env.Logger, config)

	registry := NewSecureMethodRegistry(authMiddleware, rbacMiddleware, env.Logger, config)

	// Mock handler
	handlerCalled := false
	handler := func(params map[string]interface{}, client security.ClientConnection) (security.JsonRpcResponse, error) {
		handlerCalled = true
		return &MockJsonRpcResponse{result: "success"}, nil
	}

	// Register method with admin role
	registry.RegisterMethod("admin_method", handler, security.RoleAdmin)

	// Verify method was registered
	registeredHandler, exists := registry.GetMethod("admin_method")
	assert.True(t, exists, "Method should be registered")
	assert.NotNil(t, registeredHandler)

	// Test with insufficient role (should fail)
	client := &MockClientConnection{
		clientID:      "test_client",
		userID:        "test_user",
		role:          "viewer",
		authenticated: true,
	}

	response, err := registeredHandler(map[string]interface{}{}, client)
	assert.Error(t, err, "Should fail with insufficient role")
	assert.False(t, handlerCalled, "Handler should not be called")
	assert.Nil(t, response)

	// Test with sufficient role (should succeed)
	adminClient := &MockClientConnection{
		clientID:      "admin_client",
		userID:        "admin_user",
		role:          "admin",
		authenticated: true,
	}

	response, err = registeredHandler(map[string]interface{}{}, adminClient)
	assert.NoError(t, err, "Should succeed with admin role")
	assert.True(t, handlerCalled, "Handler should be called")
	assert.NotNil(t, response)
}

func TestSecureMethodRegistry_GetAllMethods(t *testing.T) {
	// COMMON PATTERN: Use shared test environment instead of individual components
	env := testtestutils.SetupTestEnvironment(t)
	defer testtestutils.TeardownTestEnvironment(t, env)

	config := &MockSecurityConfig{}
	permissionChecker := NewPermissionChecker()
	authMiddleware := NewAuthMiddleware(env.Logger, config)
	rbacMiddleware := NewRBACMiddleware(permissionChecker, env.Logger, config)

	registry := NewSecureMethodRegistry(authMiddleware, rbacMiddleware, env.Logger, config)

	// Register multiple methods
	handler := func(params map[string]interface{}, client security.ClientConnection) (security.JsonRpcResponse, error) {
		return &MockJsonRpcResponse{result: "success"}, nil
	}

	registry.RegisterMethod("method1", handler, security.RoleViewer)
	registry.RegisterMethod("method2", handler, security.RoleOperator)
	registry.RegisterMethod("method3", handler, security.RoleAdmin)

	// Get all methods
	methods := registry.GetAllMethods()

	// Verify all methods are returned
	assert.Len(t, methods, 3, "Should return all registered methods")
	assert.Contains(t, methods, "method1")
	assert.Contains(t, methods, "method2")
	assert.Contains(t, methods, "method3")
}

func TestSecureMethodRegistry_GetMethodSecurityInfo(t *testing.T) {
	// COMMON PATTERN: Use shared test environment instead of individual components
	env := testtestutils.SetupTestEnvironment(t)
	defer testtestutils.TeardownTestEnvironment(t, env)

	config := &MockSecurityConfig{}
	permissionChecker := NewPermissionChecker()
	authMiddleware := NewAuthMiddleware(env.Logger, config)
	rbacMiddleware := NewRBACMiddleware(permissionChecker, env.Logger, config)

	registry := NewSecureMethodRegistry(authMiddleware, rbacMiddleware, env.Logger, config)

	// Get security info for non-existent method
	info := registry.GetMethodSecurityInfo("non_existent")
	assert.Equal(t, "non_existent", info["method"])
	assert.True(t, info["secured"].(bool))
	assert.Equal(t, "required", info["authentication"])
	assert.True(t, info["rbac_enabled"].(bool))
}

func TestSecurityMiddleware_Integration(t *testing.T) {
	// COMMON PATTERN: Use shared test environment instead of individual components
	env := testtestutils.SetupTestEnvironment(t)
	defer testtestutils.TeardownTestEnvironment(t, env)

	config := &MockSecurityConfig{}
	permissionChecker := NewPermissionChecker()
	authMiddleware := NewAuthMiddleware(env.Logger, config)
	rbacMiddleware := NewRBACMiddleware(permissionChecker, env.Logger, config)
	registry := NewSecureMethodRegistry(authMiddleware, rbacMiddleware, env.Logger, config)

	// Mock handler
	handlerCalled := false
	handler := func(params map[string]interface{}, client security.ClientConnection) (security.JsonRpcResponse, error) {
		handlerCalled = true
		return &MockJsonRpcResponse{result: "success"}, nil
	}

	// Register method with operator role
	registry.RegisterMethod("test_integration", handler, security.RoleOperator)

	// Test unauthenticated access (should fail at auth layer)
	unauthenticatedClient := &MockClientConnection{
		clientID:      "test_client",
		userID:        "",
		role:          "",
		authenticated: false,
	}

	registeredHandler, _ := registry.GetMethod("test_integration")
	response, err := registeredHandler(map[string]interface{}{}, unauthenticatedClient)
	assert.Error(t, err, "Should fail authentication check")
	assert.False(t, handlerCalled, "Handler should not be called")
	assert.Nil(t, response)

	// Test authenticated but insufficient role (should fail at RBAC layer)
	viewerClient := &MockClientConnection{
		clientID:      "viewer_client",
		userID:        "viewer_user",
		role:          "viewer",
		authenticated: true,
	}

	handlerCalled = false // Reset flag
	response, err = registeredHandler(map[string]interface{}{}, viewerClient)
	assert.Error(t, err, "Should fail RBAC check")
	assert.False(t, handlerCalled, "Handler should not be called")
	assert.Nil(t, response)

	// Test authenticated with sufficient role (should succeed)
	operatorClient := &MockClientConnection{
		clientID:      "operator_client",
		userID:        "operator_user",
		role:          "operator",
		authenticated: true,
	}

	handlerCalled = false // Reset flag
	response, err = registeredHandler(map[string]interface{}{}, operatorClient)
	assert.NoError(t, err, "Should succeed with proper authentication and role")
	assert.True(t, handlerCalled, "Handler should be called")
	assert.NotNil(t, response)
}
