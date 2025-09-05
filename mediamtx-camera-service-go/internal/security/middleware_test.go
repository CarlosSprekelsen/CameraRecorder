/*
Security Middleware Unit Tests

Requirements Coverage:
- REQ-SEC-002: Role-based access control for different user types
- REQ-SEC-003: Authentication enforcement for protected methods

Test Categories: Unit/Security
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package security

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
)

// =============================================================================
// MOCK IMPLEMENTATIONS (Following established pattern in codebase)
// =============================================================================

// MockClientConnection implements ClientConnection interface for testing
// Following the established pattern used by other security tests
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
	error   JsonRpcError
	id      interface{}
}

func (m *MockJsonRpcResponse) GetJSONRPC() string     { return m.jsonrpc }
func (m *MockJsonRpcResponse) GetResult() interface{} { return m.result }
func (m *MockJsonRpcResponse) GetError() JsonRpcError { return m.error }
func (m *MockJsonRpcResponse) GetID() interface{}     { return m.id }

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

// =============================================================================
// AUTHENTICATION MIDDLEWARE TESTS
// =============================================================================

// TestNewAuthMiddleware tests authentication middleware creation
// Following the established pattern: use test environment with logger
func TestNewAuthMiddleware(t *testing.T) {
	// Use security test environment following established pattern
	env := SetupTestSecurityEnvironment(t)
	defer TeardownTestSecurityEnvironment(t, env)

	config := &MockSecurityConfig{}

	// Following established pattern: use minimal logger for middleware compatibility
	middleware := NewAuthMiddleware(logging.NewLogger("test"), config)

	assert.NotNil(t, middleware, "Auth middleware should be created successfully")
	// Note: Fields are unexported, so we can't test them directly
	// This is intentional for encapsulation (following established pattern)
}

// TestAuthMiddleware_RequireAuth_Authenticated tests authenticated client access
func TestAuthMiddleware_RequireAuth_Authenticated(t *testing.T) {
	// Use security test environment following established pattern
	env := SetupTestSecurityEnvironment(t)
	defer TeardownTestSecurityEnvironment(t, env)

	config := &MockSecurityConfig{}

	// Following established pattern: use minimal logger for middleware compatibility
	middleware := NewAuthMiddleware(logging.NewLogger("test"), config)

	// Mock authenticated client
	client := &MockClientConnection{
		clientID:      "test_client",
		userID:        "test_user",
		role:          "viewer",
		authenticated: true,
	}

	// Mock handler that should be called
	handlerCalled := false
	handler := func(params map[string]interface{}, client ClientConnection) (JsonRpcResponse, error) {
		handlerCalled = true
		return &MockJsonRpcResponse{jsonrpc: "2.0", result: "success", id: 1}, nil
	}

	// Test that authenticated client can access protected method
	securedHandler := middleware.RequireAuth(handler)
	response, err := securedHandler(map[string]interface{}{"test": "data"}, client)

	assert.NoError(t, err, "Authenticated client should not get error")
	assert.NotNil(t, response, "Authenticated client should get response")
	assert.True(t, handlerCalled, "Handler should have been called for authenticated client")
}

// TestAuthMiddleware_RequireAuth_NotAuthenticated tests unauthenticated client access
func TestAuthMiddleware_RequireAuth_NotAuthenticated(t *testing.T) {
	// Use security test environment following established pattern
	env := SetupTestSecurityEnvironment(t)
	defer TeardownTestSecurityEnvironment(t, env)

	config := &MockSecurityConfig{}

	// Following established pattern: use minimal logger for middleware compatibility
	middleware := NewAuthMiddleware(logging.NewLogger("test"), config)

	// Mock unauthenticated client
	client := &MockClientConnection{
		clientID:      "test_client",
		userID:        "",
		role:          "",
		authenticated: false,
	}

	// Mock handler that should NOT be called
	handlerCalled := false
	handler := func(params map[string]interface{}, client ClientConnection) (JsonRpcResponse, error) {
		handlerCalled = true
		return &MockJsonRpcResponse{jsonrpc: "2.0", result: "success", id: 1}, nil
	}

	// Test that unauthenticated client cannot access protected method
	securedHandler := middleware.RequireAuth(handler)
	response, err := securedHandler(map[string]interface{}{"test": "data"}, client)

	assert.Error(t, err, "Unauthenticated client should get error")
	assert.Nil(t, response, "Unauthenticated client should not get response")
	assert.False(t, handlerCalled, "Handler should not have been called for unauthenticated client")
	assert.Contains(t, err.Error(), "authentication required", "Error should indicate authentication required")
}

// =============================================================================
// RBAC MIDDLEWARE TESTS
// =============================================================================

// TestNewRBACMiddleware tests RBAC middleware creation
func TestNewRBACMiddleware(t *testing.T) {
	// Use security test environment following established pattern
	env := SetupTestSecurityEnvironment(t)
	defer TeardownTestSecurityEnvironment(t, env)

	permissionChecker := NewPermissionChecker()
	config := &MockSecurityConfig{}

	// Following established pattern: use minimal logger for middleware compatibility
	middleware := NewRBACMiddleware(permissionChecker, logging.NewLogger("test"), config)

	assert.NotNil(t, middleware, "RBAC middleware should be created successfully")
	// Note: Fields are unexported, so we can't test them directly
	// This is intentional for encapsulation (following established pattern)
}

// TestRBACMiddleware_RequireRole_SufficientRole tests role-based access with sufficient permissions
func TestRBACMiddleware_RequireRole_SufficientRole(t *testing.T) {
	// Use security test environment following established pattern
	env := SetupTestSecurityEnvironment(t)
	defer TeardownTestSecurityEnvironment(t, env)

	permissionChecker := NewPermissionChecker()
	config := &MockSecurityConfig{}

	// Following established pattern: logging.NewLogger("test") (like other security components)
	middleware := NewRBACMiddleware(permissionChecker, logging.NewLogger("test"), config)

	// Mock authenticated client with sufficient role
	client := &MockClientConnection{
		clientID:      "test_client",
		userID:        "test_user",
		role:          "admin", // Admin role should have access to most methods
		authenticated: true,
	}

	// Mock handler that should be called
	handlerCalled := false
	handler := func(params map[string]interface{}, client ClientConnection) (JsonRpcResponse, error) {
		handlerCalled = true
		return &MockJsonRpcResponse{jsonrpc: "2.0", result: "success", id: 1}, nil
	}

	// Test that client with sufficient role can access protected method
	securedHandler := middleware.RequireRole(RoleAdmin, handler)
	response, err := securedHandler(map[string]interface{}{"test": "data"}, client)

	assert.NoError(t, err, "Client with sufficient role should not get error")
	assert.NotNil(t, response, "Client with sufficient role should get response")
	assert.True(t, handlerCalled, "Handler should have been called for client with sufficient role")
}

// TestRBACMiddleware_RequireRole_InsufficientRole tests role-based access with insufficient permissions
func TestRBACMiddleware_RequireRole_InsufficientRole(t *testing.T) {
	// Use security test environment following established pattern
	env := SetupTestSecurityEnvironment(t)
	defer TeardownTestSecurityEnvironment(t, env)

	permissionChecker := NewPermissionChecker()
	config := &MockSecurityConfig{}

	// Following established pattern: logging.NewLogger("test") (like other security components)
	middleware := NewRBACMiddleware(permissionChecker, logging.NewLogger("test"), config)

	// Mock authenticated client with insufficient role
	client := &MockClientConnection{
		clientID:      "test_client",
		userID:        "test_user",
		role:          "viewer", // Viewer role should not have admin access
		authenticated: true,
	}

	// Mock handler that should NOT be called
	handlerCalled := false
	handler := func(params map[string]interface{}, client ClientConnection) (JsonRpcResponse, error) {
		handlerCalled = true
		return &MockJsonRpcResponse{jsonrpc: "2.0", result: "success", id: 1}, nil
	}

	// Test that client with insufficient role cannot access protected method
	securedHandler := middleware.RequireRole(RoleAdmin, handler)
	response, err := securedHandler(map[string]interface{}{"test": "data"}, client)

	assert.Error(t, err, "Client with insufficient role should get error")
	assert.Nil(t, response, "Client with insufficient role should not get response")
	assert.False(t, handlerCalled, "Handler should not have been called for client with insufficient role")
}

// =============================================================================
// INTEGRATION TESTS
// =============================================================================

// TestSecurityMiddleware_Integration tests security middleware integration
func TestSecurityMiddleware_Integration(t *testing.T) {
	// Use security test environment following established pattern
	env := SetupTestSecurityEnvironment(t)
	defer TeardownTestSecurityEnvironment(t, env)

	// Test complete security flow: Auth + RBAC
	authMiddleware := NewAuthMiddleware(logging.NewLogger("test"), &MockSecurityConfig{})
	rbacMiddleware := NewRBACMiddleware(env.RoleManager, logging.NewLogger("test"), &MockSecurityConfig{})

	// Mock authenticated admin client
	client := &MockClientConnection{
		clientID:      "test_client",
		userID:        "test_user",
		role:          "admin",
		authenticated: true,
	}

	// Mock handler
	handlerCalled := false
	handler := func(params map[string]interface{}, client ClientConnection) (JsonRpcResponse, error) {
		handlerCalled = true
		return &MockJsonRpcResponse{jsonrpc: "2.0", result: "success", id: 1}, nil
	}

	// Apply both middlewares
	securedHandler := authMiddleware.RequireAuth(handler)
	roleSecuredHandler := rbacMiddleware.RequireRole(RoleAdmin, securedHandler)

	// Test complete flow
	response, err := roleSecuredHandler(map[string]interface{}{"test": "data"}, client)

	assert.NoError(t, err, "Complete security flow should succeed for authorized client")
	assert.NotNil(t, response, "Complete security flow should return response")
	assert.True(t, handlerCalled, "Handler should have been called in complete security flow")
}

// =============================================================================
// EDGE CASE TESTS
// =============================================================================

// TestSecurityMiddleware_EdgeCases tests edge cases in security middleware
func TestSecurityMiddleware_EdgeCases(t *testing.T) {
	// Use security test environment following established pattern
	env := SetupTestSecurityEnvironment(t)
	defer TeardownTestSecurityEnvironment(t, env)

	t.Run("nil_handler", func(t *testing.T) {
		authMiddleware := NewAuthMiddleware(logging.NewLogger("test"), &MockSecurityConfig{})

		// This should not panic
		securedHandler := authMiddleware.RequireAuth(nil)
		assert.NotNil(t, securedHandler, "Should handle nil handler gracefully")
	})

	t.Run("empty_client_data", func(t *testing.T) {
		authMiddleware := NewAuthMiddleware(logging.NewLogger("test"), &MockSecurityConfig{})

		client := &MockClientConnection{
			clientID:      "",
			userID:        "",
			role:          "",
			authenticated: false,
		}

		handler := func(params map[string]interface{}, client ClientConnection) (JsonRpcResponse, error) {
			return &MockJsonRpcResponse{jsonrpc: "2.0", result: "success", id: 1}, nil
		}

		securedHandler := authMiddleware.RequireAuth(handler)
		_, err := securedHandler(map[string]interface{}{}, client)

		assert.Error(t, err, "Should reject client with empty data")
		assert.Contains(t, err.Error(), "authentication required", "Should indicate authentication required")
	})
}

// =============================================================================
// COMPREHENSIVE MIDDLEWARE TESTS FOR 90%+ COVERAGE
// =============================================================================

func TestNewSecureMethodRegistry(t *testing.T) {
	t.Parallel()

	registry := NewSecureMethodRegistry()
	assert.NotNil(t, registry)
	assert.NotNil(t, registry.methods)
	assert.Empty(t, registry.methods)
}

func TestSecureMethodRegistry_RegisterMethod(t *testing.T) {
	t.Parallel()

	registry := NewSecureMethodRegistry()
	
	tests := []struct {
		name         string
		method       string
		requiredRole string
		description  string
		wantErr      bool
	}{
		{"Valid method", "test_method", "viewer", "Test method", false},
		{"Admin method", "admin_method", "admin", "Admin method", false},
		{"Empty method", "", "viewer", "Empty method", true},
		{"Empty role", "test_method", "", "Empty role", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := registry.RegisterMethod(tt.method, tt.requiredRole, tt.description)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Verify method was registered
				method, exists := registry.methods[tt.method]
				assert.True(t, exists)
				assert.Equal(t, tt.requiredRole, method.RequiredRole)
				assert.Equal(t, tt.description, method.Description)
			}
		})
	}
}

func TestSecureMethodRegistry_GetMethod(t *testing.T) {
	t.Parallel()

	registry := NewSecureMethodRegistry()
	
	// Register a method
	err := registry.RegisterMethod("test_method", "viewer", "Test method")
	require.NoError(t, err)
	
	tests := []struct {
		name     string
		method   string
		want     *MethodSecurityInfo
		wantErr  bool
	}{
		{"Existing method", "test_method", &MethodSecurityInfo{
			Method: "test_method",
			RequiredRole: "viewer",
			Description: "Test method",
		}, false},
		{"Non-existing method", "nonexistent", nil, true},
		{"Empty method", "", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			method, err := registry.GetMethod(tt.method)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, method)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, method)
			}
		})
	}
}

func TestSecureMethodRegistry_GetAllMethods(t *testing.T) {
	t.Parallel()

	registry := NewSecureMethodRegistry()
	
	// Initially empty
	methods := registry.GetAllMethods()
	assert.Empty(t, methods)
	
	// Register some methods
	err := registry.RegisterMethod("method1", "viewer", "Method 1")
	require.NoError(t, err)
	
	err = registry.RegisterMethod("method2", "admin", "Method 2")
	require.NoError(t, err)
	
	// Get all methods
	methods = registry.GetAllMethods()
	assert.Len(t, methods, 2)
	
	// Verify methods are present
	methodNames := make(map[string]bool)
	for _, method := range methods {
		methodNames[method.Method] = true
	}
	assert.True(t, methodNames["method1"])
	assert.True(t, methodNames["method2"])
}

func TestSecureMethodRegistry_GetMethodSecurityInfo(t *testing.T) {
	t.Parallel()

	registry := NewSecureMethodRegistry()
	
	// Register a method
	err := registry.RegisterMethod("test_method", "viewer", "Test method")
	require.NoError(t, err)
	
	tests := []struct {
		name     string
		method   string
		want     *MethodSecurityInfo
		wantErr  bool
	}{
		{"Existing method", "test_method", &MethodSecurityInfo{
			Method: "test_method",
			RequiredRole: "viewer",
			Description: "Test method",
		}, false},
		{"Non-existing method", "nonexistent", nil, true},
		{"Empty method", "", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := registry.GetMethodSecurityInfo(tt.method)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, info)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, info)
			}
		})
	}
}

func TestSecureMethodRegistry_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	registry := NewSecureMethodRegistry()
	
	// Test concurrent registration
	done := make(chan bool, 10)
	
	for i := 0; i < 10; i++ {
		go func(methodID int) {
			method := "method" + string(rune(methodID))
			err := registry.RegisterMethod(method, "viewer", "Concurrent method")
			assert.NoError(t, err)
			done <- true
		}(i)
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
	
	// Verify all methods were registered
	methods := registry.GetAllMethods()
	assert.Len(t, methods, 10)
}

func TestSecureMethodRegistry_DuplicateRegistration(t *testing.T) {
	t.Parallel()

	registry := NewSecureMethodRegistry()
	
	// Register method first time
	err := registry.RegisterMethod("test_method", "viewer", "First description")
	require.NoError(t, err)
	
	// Try to register same method again
	err = registry.RegisterMethod("test_method", "admin", "Second description")
	assert.Error(t, err, "Should error on duplicate registration")
	
	// Verify original method is unchanged
	method, err := registry.GetMethod("test_method")
	require.NoError(t, err)
	assert.Equal(t, "viewer", method.RequiredRole)
	assert.Equal(t, "First description", method.Description)
}
