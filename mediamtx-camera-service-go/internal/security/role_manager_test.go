/*
Role Manager Unit Tests

Requirements Coverage:
- REQ-SEC-003: Role-based access control for different user types

Test Categories: Unit/Security
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package security

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPermissionChecker_RoleHierarchy tests role hierarchy functionality
func TestPermissionChecker_RoleHierarchy(t *testing.T) {
	// REQ-SEC-003: Role-based access control for different user types

	// Use security test environment from test helpers
	env := SetupTestSecurityEnvironment(t)
	defer TeardownTestSecurityEnvironment(t, env)

	// Test role hierarchy
	assert.True(t, RoleAdmin >= RoleOperator)
	assert.True(t, RoleOperator >= RoleViewer)
	assert.True(t, RoleAdmin >= RoleViewer)

	// Test role string conversion
	assert.Equal(t, "viewer", RoleViewer.String())
	assert.Equal(t, "operator", RoleOperator.String())
	assert.Equal(t, "admin", RoleAdmin.String())

	// Test role validation
	role, err := env.RoleManager.ValidateRole("admin")
	assert.NoError(t, err)
	assert.Equal(t, RoleAdmin, role)

	role, err = env.RoleManager.ValidateRole("invalid_role")
	assert.Error(t, err)
	assert.Equal(t, RoleViewer, role) // Default fallback
}

// TestPermissionChecker_MethodPermissions tests method permission checking
func TestPermissionChecker_MethodPermissions(t *testing.T) {
	// REQ-SEC-003: Role-based access control for different user types

	// Use security test environment from test helpers
	env := SetupTestSecurityEnvironment(t)
	defer TeardownTestSecurityEnvironment(t, env)

	tests := []struct {
		name      string
		method    string
		userRole  Role
		hasAccess bool
	}{
		{
			name:      "viewer can access ping",
			method:    "ping",
			userRole:  RoleViewer,
			hasAccess: true,
		},
		{
			name:      "viewer cannot access take_snapshot",
			method:    "take_snapshot",
			userRole:  RoleViewer,
			hasAccess: false,
		},
		{
			name:      "operator can access take_snapshot",
			method:    "take_snapshot",
			userRole:  RoleOperator,
			hasAccess: true,
		},
		{
			name:      "admin can access get_metrics",
			method:    "get_metrics",
			userRole:  RoleAdmin,
			hasAccess: true,
		},
		{
			name:      "operator cannot access get_metrics",
			method:    "get_metrics",
			userRole:  RoleOperator,
			hasAccess: false,
		},
		{
			name:      "unknown method requires admin",
			method:    "unknown_method",
			userRole:  RoleAdmin,
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
	// Use security test environment from test helpers
	env := SetupTestSecurityEnvironment(t)
	defer TeardownTestSecurityEnvironment(t, env)

	t.Run("get_required_role", func(t *testing.T) {
		role := env.RoleManager.GetRequiredRole("ping")
		assert.Equal(t, RoleViewer, role)

		role = env.RoleManager.GetRequiredRole("take_snapshot")
		assert.Equal(t, RoleOperator, role)

		role = env.RoleManager.GetRequiredRole("get_metrics")
		assert.Equal(t, RoleAdmin, role)

		role = env.RoleManager.GetRequiredRole("nonexistent_method")
		assert.Equal(t, RoleAdmin, role) // Default to admin for unknown methods
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
		err := env.RoleManager.AddMethodPermission("test_method", RoleViewer)
		require.NoError(t, err)
		assert.True(t, env.RoleManager.HasPermission(RoleViewer, "test_method"))
		assert.True(t, env.RoleManager.HasPermission(RoleOperator, "test_method"))
		assert.True(t, env.RoleManager.HasPermission(RoleAdmin, "test_method"))
	})

	t.Run("remove_method_permission", func(t *testing.T) {
		// First add a permission
		err := env.RoleManager.AddMethodPermission("temp_method", RoleOperator)
		require.NoError(t, err)
		assert.True(t, env.RoleManager.HasPermission(RoleOperator, "temp_method"))

		// Then remove it
		err = env.RoleManager.RemoveMethodPermission("temp_method")
		require.NoError(t, err)
		assert.False(t, env.RoleManager.HasPermission(RoleViewer, "temp_method"))
		assert.False(t, env.RoleManager.HasPermission(RoleOperator, "temp_method"))
		assert.False(t, env.RoleManager.HasPermission(RoleAdmin, "temp_method"))
	})

	t.Run("validate_role_edge_cases", func(t *testing.T) {
		role, err := env.RoleManager.ValidateRole("viewer")
		assert.NoError(t, err)
		assert.Equal(t, RoleViewer, role)

		role, err = env.RoleManager.ValidateRole("operator")
		assert.NoError(t, err)
		assert.Equal(t, RoleOperator, role)

		role, err = env.RoleManager.ValidateRole("admin")
		assert.NoError(t, err)
		assert.Equal(t, RoleAdmin, role)

		invalidRole, err := env.RoleManager.ValidateRole("invalid_role")
		assert.Error(t, err)
		assert.Equal(t, RoleViewer, invalidRole) // Returns default role

		emptyRole, err := env.RoleManager.ValidateRole("")
		assert.Error(t, err)
		assert.Equal(t, RoleViewer, emptyRole) // Returns default role

		caseRole, err := env.RoleManager.ValidateRole("VIEWER") // Should work due to ToLower conversion
		assert.NoError(t, err)
		assert.Equal(t, RoleViewer, caseRole) // Should return viewer role
	})
}

func TestPermissionChecker_AdditionalEdgeCases(t *testing.T) {
	t.Run("get_required_role_with_whitespace", func(t *testing.T) {
		checker := NewPermissionChecker()

		role := checker.GetRequiredRole("   ")
		assert.Equal(t, RoleAdmin, role) // Should default to admin
	})

	t.Run("add_method_permission_with_whitespace", func(t *testing.T) {
		checker := NewPermissionChecker()

		err := checker.AddMethodPermission("   ", RoleViewer)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "method name cannot be empty")
	})

	t.Run("remove_method_permission_with_whitespace", func(t *testing.T) {
		checker := NewPermissionChecker()

		err := checker.RemoveMethodPermission("   ")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "method name cannot be empty")
	})

	t.Run("add_method_permission_with_invalid_role", func(t *testing.T) {
		checker := NewPermissionChecker()

		err := checker.AddMethodPermission("test_method", Role(999))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid role")
	})
}

// TestPermissionChecker_GetPermissionsForRole tests the GetPermissionsForRole function
// CRITICAL SECURITY TEST: This function determines what permissions a role has
func TestPermissionChecker_GetPermissionsForRole(t *testing.T) {
	t.Parallel()

	checker := NewPermissionChecker()

	testCases := []struct {
		name        string
		roleStr     string
		expected    []string
		description string
	}{
		{
			name:        "admin_permissions",
			roleStr:     "admin",
			expected:    []string{"view", "control", "admin"},
			description: "Admin should have all permission categories",
		},
		{
			name:        "operator_permissions",
			roleStr:     "operator",
			expected:    []string{"view", "control"},
			description: "Operator should have view and control permissions",
		},
		{
			name:        "viewer_permissions",
			roleStr:     "viewer",
			expected:    []string{"view"},
			description: "Viewer should have only view permissions",
		},
		{
			name:        "invalid_role",
			roleStr:     "invalid",
			expected:    []string{},
			description: "Invalid role should return empty permissions",
		},
		{
			name:        "empty_role",
			roleStr:     "",
			expected:    []string{},
			description: "Empty role should return empty permissions",
		},
		{
			name:        "uppercase_role",
			roleStr:     "ADMIN",
			expected:    []string{"view", "control", "admin"},
			description: "Uppercase role should work (case insensitive)",
		},
		{
			name:        "numeric_role",
			roleStr:     "123",
			expected:    []string{},
			description: "Numeric role should return empty permissions",
		},
		{
			name:        "special_chars_role",
			roleStr:     "admin@#$",
			expected:    []string{},
			description: "Role with special characters should return empty permissions",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := checker.GetPermissionsForRole(tc.roleStr)
			assert.Equal(t, tc.expected, result, tc.description)
		})
	}
}

// TestRole_String tests the String method for complete coverage
func TestRole_String(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		role     Role
		expected string
	}{
		{RoleViewer, "viewer"},
		{RoleOperator, "operator"},
		{RoleAdmin, "admin"},
		{Role(999), "unknown"}, // Test unknown role
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			result := tc.role.String()
			assert.Equal(t, tc.expected, result)
		})
	}
}

// Performance benchmarks for permission checker
func BenchmarkPermissionChecker_HasPermission(b *testing.B) {
	checker := NewPermissionChecker()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		checker.HasPermission(RoleViewer, "ping")
		checker.HasPermission(RoleOperator, "take_snapshot")
		checker.HasPermission(RoleAdmin, "get_metrics")
	}
}
