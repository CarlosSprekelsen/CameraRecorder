package security

import (
	"fmt"
	"strings"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
)

// Role represents the user role hierarchy in the system.
// Higher values indicate higher permissions.
type Role int

const (
	// RoleViewer represents read-only access to camera status and basic information.
	RoleViewer Role = iota + 1
	// RoleOperator represents viewer permissions plus camera control operations.
	RoleOperator
	// RoleAdmin represents full access to all features including system management.
	RoleAdmin
)

// RoleNames maps role constants to their string representations.
var RoleNames = map[Role]string{
	RoleViewer:   "viewer",
	RoleOperator: "operator",
	RoleAdmin:    "admin",
}

// StringRoleNames maps string role names to their Role constants.
var StringRoleNames = map[string]Role{
	"viewer":   RoleViewer,
	"operator": RoleOperator,
	"admin":    RoleAdmin,
}

// String returns the string representation of the role.
func (r Role) String() string {
	if name, exists := RoleNames[r]; exists {
		return name
	}
	return "unknown"
}

// PermissionChecker manages role-based access control for API methods.
// Implements method-level permission checking with role hierarchy enforcement.
type PermissionChecker struct {
	methodPermissions map[string]Role
	logger            *logging.Logger
}

// NewPermissionChecker creates a new permission checker with default method permissions.
// Initializes the permission matrix based on the Python system's method permissions.
func NewPermissionChecker() *PermissionChecker {
	checker := &PermissionChecker{
		methodPermissions: make(map[string]Role),
		logger:            logging.NewLogger("permission-checker"),
	}

	// Initialize method permissions based on Python system
	// Viewer permissions (read-only operations)
	viewerMethods := []string{
		"ping",
		"get_camera_list",
		"get_camera_status",
		"get_camera_capabilities",
		"list_recordings",
		"list_snapshots",
		"get_recording_info",
		"get_snapshot_info",
		"get_streams",
		"get_stream_url",
		"get_stream_status",
	}

	// Operator permissions (camera control operations)
	operatorMethods := []string{
		"take_snapshot",
		"start_recording",
		"stop_recording",
		"delete_recording",
		"delete_snapshot",
		"start_streaming",
		"stop_streaming",
	}

	// Admin permissions (system management operations)
	adminMethods := []string{
		"get_metrics",
		"get_status",
		"get_server_info",
		"get_storage_info",
		"set_retention_policy",
		"cleanup_old_files",
	}

	// Set permissions for each method
	for _, method := range viewerMethods {
		checker.methodPermissions[method] = RoleViewer
	}

	for _, method := range operatorMethods {
		checker.methodPermissions[method] = RoleOperator
	}

	for _, method := range adminMethods {
		checker.methodPermissions[method] = RoleAdmin
	}

	checker.logger.WithField("method_count", fmt.Sprintf("%d", len(checker.methodPermissions))).Info("Permission checker initialized")
	return checker
}

// HasPermission checks if a user with the given role has permission to access the specified method.
// Returns true if the user has sufficient permissions, false otherwise.
func (p *PermissionChecker) HasPermission(userRole Role, method string) bool {
	if strings.TrimSpace(method) == "" {
		p.logger.Warn("Method name cannot be empty")
		return false
	}

	requiredRole, exists := p.methodPermissions[method]
	if !exists {
		p.logger.Warnf("Method '%s' not found in permission matrix", method)
		return false
	}

	hasPermission := userRole >= requiredRole

	p.logger.WithFields(logging.Fields{
		"method":         method,
		"user_role":      userRole.String(),
		"required_role":  requiredRole.String(),
		"has_permission": hasPermission,
	}).Debug("Permission check performed")

	return hasPermission
}

// GetRequiredRole returns the minimum role required to access the specified method.
// Returns RoleAdmin if the method is not found in the permission matrix.
func (p *PermissionChecker) GetRequiredRole(method string) Role {
	if strings.TrimSpace(method) == "" {
		return RoleAdmin
	}

	requiredRole, exists := p.methodPermissions[method]
	if !exists {
		p.logger.Warnf("Method '%s' not found in permission matrix", method)
		return RoleAdmin
	}

	return requiredRole
}

// ValidateRole validates a string role and converts it to a Role constant.
// Returns an error if the role string is invalid.
func (p *PermissionChecker) ValidateRole(roleString string) (Role, error) {
	if strings.TrimSpace(roleString) == "" {
		return RoleViewer, fmt.Errorf("role cannot be empty")
	}

	role, exists := StringRoleNames[strings.ToLower(roleString)]
	if !exists {
		return RoleViewer, fmt.Errorf("invalid role: %s", roleString)
	}

	return role, nil
}

// GetRoleHierarchy returns the role hierarchy information.
// Useful for debugging and validation purposes.
func (p *PermissionChecker) GetRoleHierarchy() map[string]int {
	hierarchy := make(map[string]int)
	for role, name := range RoleNames {
		hierarchy[name] = int(role)
	}
	return hierarchy
}

// GetMethodPermissions returns a copy of the method permissions map.
// Useful for debugging and validation purposes.
func (p *PermissionChecker) GetMethodPermissions() map[string]string {
	permissions := make(map[string]string)
	for method, role := range p.methodPermissions {
		permissions[method] = role.String()
	}
	return permissions
}

// AddMethodPermission adds a new method permission to the checker.
// This method allows dynamic permission configuration.
func (p *PermissionChecker) AddMethodPermission(method string, requiredRole Role) error {
	if strings.TrimSpace(method) == "" {
		return fmt.Errorf("method name cannot be empty")
	}

	if requiredRole < RoleViewer || requiredRole > RoleAdmin {
		return fmt.Errorf("invalid role: %d", requiredRole)
	}

	p.methodPermissions[method] = requiredRole

	p.logger.WithFields(logging.Fields{
		"method":        method,
		"required_role": requiredRole.String(),
	}).Info("Method permission added")

	return nil
}

// RemoveMethodPermission removes a method permission from the checker.
// Returns an error if the method doesn't exist.
func (p *PermissionChecker) RemoveMethodPermission(method string) error {
	if strings.TrimSpace(method) == "" {
		return fmt.Errorf("method name cannot be empty")
	}

	if _, exists := p.methodPermissions[method]; !exists {
		return fmt.Errorf("method '%s' not found in permission matrix", method)
	}

	delete(p.methodPermissions, method)

	p.logger.WithField("method", method).Info("Method permission removed")
	return nil
}
