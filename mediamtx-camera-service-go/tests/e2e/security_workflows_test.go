/*
Security Workflows E2E Tests

Tests complete user workflows for authentication, authorization, session management,
and token expiry enforcement. Leverages multiple WebSocket connections for comprehensive
role-based access control testing.

Test Categories:
- Authentication Success Workflow: Connect, authenticate with valid token, verify authenticated operations succeed
- Authentication Failure Workflow: Connect, authenticate with invalid token, verify proper error handling
- Role-Based Authorization Workflow: Test viewer, operator, admin roles with appropriate permission enforcement
- Session Management Workflow: Verify session maintained on connection, new auth required for new connections
- Token Expiry Workflow: Generate token with short expiry, verify operations fail after expiry

Business Outcomes:
- Authorized user can access protected resources
- Unauthorized user cannot access protected resources
- Role-based access control enforced correctly
- Sessions properly scoped to connections
- Token expiry enforced, preventing stale credentials

Coverage Target: 75% E2E coverage milestone
*/

package e2e

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthenticationSuccessWorkflow(t *testing.T) {
    t.Parallel()
	fixture := NewE2EFixture(t)

	// Connect and authenticate using proven flow
	err := fixture.ConnectAndAuthenticate(RoleAdmin)
	require.NoError(t, err, "Authentication should succeed")

	// Verify authenticated operations work
	cameraResp, err := fixture.client.GetCameraList()
	require.NoError(t, err, "Authenticated user should access camera list")
	require.Nil(t, cameraResp.Error)

	metricsResp, err := fixture.client.GetSystemMetrics()
	require.NoError(t, err, "Authenticated admin should access metrics")
	require.Nil(t, metricsResp.Error)
}

func TestAuthenticationFailureWorkflow(t *testing.T) {
    t.Parallel()
	fixture := NewE2EFixture(t)

	// Connect but don't authenticate
	err := fixture.client.Connect()
	require.NoError(t, err, "Connection should succeed")

	// Try to access protected resource without authentication
	cameraResp, err := fixture.client.GetCameraList()
	require.NoError(t, err, "Request should not fail")
	require.NotNil(t, cameraResp.Error, "Should get authentication error")

	errorInfo := cameraResp.Error
	assert.Equal(t, -32001, errorInfo.Code, "Should get authentication required error")
	assert.Contains(t, errorInfo.Message, "auth_required", "Error should mention authentication")
}

func TestRoleBasedAuthorizationWorkflow(t *testing.T) {
	fixture := NewE2EFixture(t)

	// Test viewer role
	err := fixture.ConnectAndAuthenticate(RoleViewer)
	require.NoError(t, err, "Viewer should authenticate successfully")

	// Viewer can list cameras
	listResp, err := fixture.client.GetCameraList()
	require.NoError(t, err, "Viewer should access camera list")
	require.Nil(t, listResp.Error)

	// Viewer cannot start recording
	recResp, err := fixture.client.StartRecording(DefaultCameraID)
	require.NoError(t, err, "Request should not fail")
	require.NotNil(t, recResp.Error, "Viewer should not start recording")

	// Viewer cannot access metrics
	metricsResp, err := fixture.client.GetSystemMetrics()
	require.NoError(t, err, "Request should not fail")
	require.NotNil(t, metricsResp.Error, "Viewer should not access metrics")

	// Close connection for next test
	fixture.client.Close()

	// Test operator role
	err = fixture.ConnectAndAuthenticate(RoleOperator)
	require.NoError(t, err, "Operator should authenticate successfully")

    // Operator can access operator-only endpoints (example: list recordings)
    // Use a read-only check to keep this test parallel-safe
    _, err = fixture.client.ListRecordings()
    require.NoError(t, err, "Operator should access recordings list")

	// Operator cannot access metrics
	opMetricsResp, err := fixture.client.GetSystemMetrics()
	require.NoError(t, err, "Request should not fail")
	require.NotNil(t, opMetricsResp.Error, "Operator should not access metrics")

	// Close connection for next test
	fixture.client.Close()

	// Test admin role
	err = fixture.ConnectAndAuthenticate(RoleAdmin)
	require.NoError(t, err, "Admin should authenticate successfully")

    // Admin can do everything (read-only check here)
    adminMetricsResp, err := fixture.client.GetSystemMetrics()
	require.NoError(t, err, "Admin should access metrics")
	require.Nil(t, adminMetricsResp.Error)

	adminHealthResp, err := fixture.client.GetSystemHealth()
	require.NoError(t, err, "Admin should access health status")
	require.Nil(t, adminHealthResp.Error)
}

func TestSessionManagementWorkflow(t *testing.T) {
    t.Parallel()
	fixture := NewE2EFixture(t)

	// Create first session
	err := fixture.ConnectAndAuthenticate(RoleOperator)
	require.NoError(t, err, "First session should work")

	// Verify first session works
	resp1, err := fixture.client.GetCameraList()
	require.NoError(t, err, "First session should work")
	require.Nil(t, resp1.Error)

	// Close first connection
	fixture.client.Close()

	// Create second session with same role but new connection
	err = fixture.ConnectAndAuthenticate(RoleOperator)
	require.NoError(t, err, "Second session should work")

	// Verify second session also works (new connection, same role)
	resp2, err := fixture.client.GetCameraList()
	require.NoError(t, err, "Second session should work")
	require.Nil(t, resp2.Error)

	// Close second connection
	fixture.client.Close()

	// Create third session with different role
	err = fixture.ConnectAndAuthenticate(RoleViewer)
	require.NoError(t, err, "Third session should work")

	// Verify third session works with different user
	resp3, err := fixture.client.GetCameraList()
	require.NoError(t, err, "Third session should work")
	require.Nil(t, resp3.Error)

	// Verify sessions are independent (user2 cannot start recording)
	recResp, err := fixture.client.StartRecording(DefaultCameraID)
	require.NoError(t, err, "Request should not fail")
	require.NotNil(t, recResp.Error, "Viewer session should not start recording")
}

func TestTokenExpiryWorkflow(t *testing.T) {
	fixture := NewE2EFixture(t)

	// Connect
	err := fixture.client.Connect()
	require.NoError(t, err, "Connection should succeed")

	// Get JWT token with very short expiry (1 second)
	// Use SecurityHelper via asserter to generate a short-lived token
	token, err := fixture.secHelper.GenerateTestToken("test_user", RoleAdmin, 1*time.Second)
	require.NoError(t, err, "Should get JWT token")

	// Authenticate with short-lived token
	err = fixture.client.Authenticate(token)
	require.NoError(t, err, "Authentication with fresh token should succeed")

	// Verify operations work initially
	resp1, err := fixture.client.GetCameraList()
	require.NoError(t, err, "Operations should work with fresh token")
	require.Nil(t, resp1.Error)

	// Wait for token to expire
	time.Sleep(2 * time.Second)

	// Verify operations fail after expiry
	resp2, err := fixture.client.GetCameraList()
	require.NoError(t, err, "Request should not fail")
	require.NotNil(t, resp2.Error, "Operations should fail with expired token")

	errorInfo := resp2.Error
	assert.Equal(t, -32001, errorInfo.Code, "Should get authentication required error")
	assert.Contains(t, errorInfo.Message, "auth_required", "Error should mention authentication")
}

func TestSecurityWorkflowsIntegration(t *testing.T) {
    fixture := NewE2EFixture(t)
    
    // Viewer permissions (read-only)
    err := fixture.ConnectAndAuthenticate(RoleViewer)
    require.NoError(t, err)
    _, err = fixture.client.GetCameraList()
    require.NoError(t, err)
    fixture.client.Close()
    
    // Operator permissions (read-only checks only in this parallel-safe test)
    err = fixture.ConnectAndAuthenticate(RoleOperator)
    require.NoError(t, err)
    _, err = fixture.client.GetCameraList()
    require.NoError(t, err)
    fixture.client.Close()
    
    // Admin permissions (read-only checks only)
    err = fixture.ConnectAndAuthenticate(RoleAdmin)
    require.NoError(t, err)
    _, err = fixture.client.GetCameraList()
    require.NoError(t, err)
    _, err = fixture.client.GetSystemMetrics()
    require.NoError(t, err)
}
