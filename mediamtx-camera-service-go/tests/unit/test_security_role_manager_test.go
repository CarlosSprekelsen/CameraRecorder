//go:build unit
// +build unit

/*
Role Manager Unit Tests

Requirements Coverage:
- REQ-SEC-003: Role-based access control for different user types

Test Categories: Unit/Security
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package security_test

import (
	"testing"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/security"
	"github.com/camerarecorder/mediamtx-camera-service-go/tests/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPermissionChecker_RoleHierarchy tests role hierarchy functionality
func TestPermissionChecker_RoleHierarchy(t *testing.T) {
	// REQ-SEC-003: Role-based access control for different user types

	// Use shared security test environment
	env := utils.SetupSecurityTestEnvironment(t)
	defer utils.TeardownSecurityTestEnvironment(t, env)

	// Test role hierarchy
	assert.True(t, security.RoleAdmin >= security.RoleOperator)
	assert.True(t, security.RoleOperator >= security.RoleViewer)
	assert.True(t, security.RoleAdmin >= security.RoleViewer)

	// Test role string conversion
	assert.Equal(t, "viewer", security.RoleViewer.String())
	assert.Equal(t, "operator", security.RoleOperator.String())
	assert.Equal(t, "admin", security.RoleAdmin.String())

	// Test role validation
	role, err := env.RoleManager.ValidateRole("admin")
	assert.NoError(t, err)
	assert.Equal(t, security.RoleAdmin, role)

	role, err = env.RoleManager.ValidateRole("invalid_role")
	assert.Error(t, err)
	assert.Equal(t, security.RoleViewer, role) // Default fallback
}

// TestPermissionChecker_MethodPermissions tests method permission checking
func TestPermissionChecker_MethodPermissions(t *testing.T) {
	// REQ-SEC-003: Role-based access control for different user types

	// Use shared security test environment
	env := utils.SetupSecurityTestEnvironment(t)
	defer utils.TeardownSecurityTestEnvironment(t, env)

	tests := []struct {
		name      string
		method    string
		userRole  security.Role
		hasAccess bool
	}{
		{
			name:      "viewer can access ping",
			method:    "ping",
			userRole:  security.RoleViewer,
			hasAccess: true,
		},
		{
			name:      "viewer cannot access take_snapshot",
			method:    "take_snapshot",
			userRole:  security.RoleViewer,
			hasAccess: false,
		},
		{
			name:      "operator can access take_snapshot",
			method:    "take_snapshot",
			userRole:  security.RoleOperator,
			hasAccess: true,
		},
		{
			name:      "admin can access get_metrics",
			method:    "get_metrics",
			userRole:  security.RoleAdmin,
			hasAccess: true,
		},
		{
			name:      "operator cannot access get_metrics",
			method:    "get_metrics",
			userRole:  security.RoleOperator,
			hasAccess: false,
		},
		{
			name:      "unknown method requires admin",
			method:    "unknown_method",
			userRole:  security.RoleAdmin,
			hasAccess: false, // Method not in permission matrix
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasPermission := env.RoleManager.HasPermission(tt.userRole, tt.method)
			assert.Equal(t, tt.hasAccess, hasPermission)
		})
	}
}

func TestPermissionChecker_EdgeCases(t *testing.T) {
	// Use shared security test environment
	env := utils.SetupSecurityTestEnvironment(t)
	defer utils.TeardownSecurityTestEnvironment(t, env)

	t.Run("get_required_role", func(t *testing.T) {
		role := env.RoleManager.GetRequiredRole("ping")
		assert.Equal(t, security.RoleViewer, role)

		role = env.RoleManager.GetRequiredRole("take_snapshot")
		assert.Equal(t, security.RoleOperator, role)

		role = env.RoleManager.GetRequiredRole("get_metrics")
		assert.Equal(t, security.RoleAdmin, role)

		role = env.RoleManager.GetRequiredRole("nonexistent_method")
		assert.Equal(t, security.RoleAdmin, role) // Default to admin for unknown methods
	})

	t.Run("get_role_hierarchy", func(t *testing.T) {
		hierarchy := env.RoleManager.GetRoleHierarchy()
		expected := map[string]int{
			"viewer":   1,
			"operator": 2,
			"admin":    3,
		}
		assert.Equal(t, expected, hierarchy)
	})

	t.Run("get_method_permissions", func(t *testing.T) {
		permissions := env.RoleManager.GetMethodPermissions()
		assert.NotEmpty(t, permissions)
		assert.Contains(t, permissions, "ping")
		assert.Contains(t, permissions, "take_snapshot")
		assert.Contains(t, permissions, "get_metrics")
	})

	t.Run("add_method_permission", func(t *testing.T) {
		err := env.RoleManager.AddMethodPermission("test_method", security.RoleViewer)
		require.NoError(t, err)
		assert.True(t, env.RoleManager.HasPermission(security.RoleViewer, "test_method"))
		assert.True(t, env.RoleManager.HasPermission(security.RoleOperator, "test_method"))
		assert.True(t, env.RoleManager.HasPermission(security.RoleAdmin, "test_method"))
	})

	t.Run("remove_method_permission", func(t *testing.T) {
		// First add a permission
		err := env.RoleManager.AddMethodPermission("temp_method", security.RoleOperator)
		require.NoError(t, err)
		assert.True(t, env.RoleManager.HasPermission(security.RoleOperator, "temp_method"))

		// Then remove it
		err = env.RoleManager.RemoveMethodPermission("temp_method")
		require.NoError(t, err)
		assert.False(t, env.RoleManager.HasPermission(security.RoleViewer, "temp_method"))
		assert.False(t, env.RoleManager.HasPermission(security.RoleOperator, "temp_method"))
		assert.False(t, env.RoleManager.HasPermission(security.RoleAdmin, "temp_method"))
	})

	t.Run("validate_role_edge_cases", func(t *testing.T) {
		role, err := env.RoleManager.ValidateRole("viewer")
		assert.NoError(t, err)
		assert.Equal(t, security.RoleViewer, role)

		role, err = env.RoleManager.ValidateRole("operator")
		assert.NoError(t, err)
		assert.Equal(t, security.RoleOperator, role)

		role, err = env.RoleManager.ValidateRole("admin")
		assert.NoError(t, err)
		assert.Equal(t, security.RoleAdmin, role)

		invalidRole, err := env.RoleManager.ValidateRole("invalid_role")
		assert.Error(t, err)
		assert.Equal(t, security.RoleViewer, invalidRole) // Returns default role

		emptyRole, err := env.RoleManager.ValidateRole("")
		assert.Error(t, err)
		assert.Equal(t, security.RoleViewer, emptyRole) // Returns default role

		caseRole, err := env.RoleManager.ValidateRole("VIEWER") // Should work due to ToLower conversion
		assert.NoError(t, err)
		assert.Equal(t, security.RoleViewer, caseRole) // Should return viewer role
	})
}

func TestPermissionChecker_AdditionalEdgeCases(t *testing.T) {
	t.Run("get_required_role_with_whitespace", func(t *testing.T) {
		checker := security.NewPermissionChecker()

		role := checker.GetRequiredRole("   ")
		assert.Equal(t, security.RoleAdmin, role) // Should default to admin
	})

	t.Run("add_method_permission_with_whitespace", func(t *testing.T) {
		checker := security.NewPermissionChecker()

		err := checker.AddMethodPermission("   ", security.RoleViewer)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "method name cannot be empty")
	})

	t.Run("remove_method_permission_with_whitespace", func(t *testing.T) {
		checker := security.NewPermissionChecker()

		err := checker.RemoveMethodPermission("   ")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "method name cannot be empty")
	})

	t.Run("add_method_permission_with_invalid_role", func(t *testing.T) {
		checker := security.NewPermissionChecker()

		err := checker.AddMethodPermission("test_method", security.Role(999))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid role")
	})
}

// Performance benchmarks for permission checker
func BenchmarkPermissionChecker_HasPermission(b *testing.B) {
	checker := security.NewPermissionChecker()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		checker.HasPermission(security.RoleViewer, "ping")
		checker.HasPermission(security.RoleOperator, "take_snapshot")
		checker.HasPermission(security.RoleAdmin, "get_metrics")
	}
}
