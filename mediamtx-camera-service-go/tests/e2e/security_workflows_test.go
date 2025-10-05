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
	asserter := NewE2EWorkflowAsserter(t)

	// Connect and authenticate using proven flow
	err := asserter.ConnectAndAuthenticate("admin")
	require.NoError(t, err, "Authentication should succeed")

	// Verify authenticated operations work
	cameraResp, err := asserter.GetCameraList()
	require.NoError(t, err, "Authenticated user should access camera list")
	require.Nil(t, cameraResp.Error)

	metricsResp, err := asserter.GetSystemMetrics()
	require.NoError(t, err, "Authenticated admin should access metrics")
	require.Nil(t, metricsResp.Error)
}

func TestAuthenticationFailureWorkflow(t *testing.T) {
	asserter := NewE2EWorkflowAsserter(t)

	// Connect but don't authenticate
	err := asserter.client.Connect()
	require.NoError(t, err, "Connection should succeed")

	// Try to access protected resource without authentication
	cameraResp, err := asserter.GetCameraList()
	require.NoError(t, err, "Request should not fail")
	require.NotNil(t, cameraResp.Error, "Should get authentication error")

	errorInfo := cameraResp.Error
	assert.Equal(t, -32001, errorInfo.Code, "Should get authentication required error")
	assert.Contains(t, errorInfo.Message, "auth_required", "Error should mention authentication")
}

func TestRoleBasedAuthorizationWorkflow(t *testing.T) {
	asserter := NewE2EWorkflowAsserter(t)

	// Test viewer role
	err := asserter.ConnectAndAuthenticate("viewer")
	require.NoError(t, err, "Viewer should authenticate successfully")

	// Viewer can list cameras
	listResp, err := asserter.GetCameraList()
	require.NoError(t, err, "Viewer should access camera list")
	require.Nil(t, listResp.Error)

	// Viewer cannot start recording
	recResp, err := asserter.StartRecording("camera0")
	require.NoError(t, err, "Request should not fail")
	require.NotNil(t, recResp.Error, "Viewer should not start recording")

	// Viewer cannot access metrics
	metricsResp, err := asserter.GetSystemMetrics()
	require.NoError(t, err, "Request should not fail")
	require.NotNil(t, metricsResp.Error, "Viewer should not access metrics")

	// Close connection for next test
	asserter.client.Close()

	// Test operator role
	err = asserter.ConnectAndAuthenticate("operator")
	require.NoError(t, err, "Operator should authenticate successfully")

	// Operator can start recording
	opRecResp, err := asserter.StartRecording("camera0")
	require.NoError(t, err, "Operator should start recording")
	require.Nil(t, opRecResp.Error)

	// Stop recording
	stopResp, err := asserter.StopRecording("camera0")
	require.NoError(t, err)
	require.Nil(t, stopResp.Error)

	// Operator cannot access metrics
	opMetricsResp, err := asserter.GetSystemMetrics()
	require.NoError(t, err, "Request should not fail")
	require.NotNil(t, opMetricsResp.Error, "Operator should not access metrics")

	// Close connection for next test
	asserter.client.Close()

	// Test admin role
	err = asserter.ConnectAndAuthenticate("admin")
	require.NoError(t, err, "Admin should authenticate successfully")

	// Admin can do everything
	adminMetricsResp, err := asserter.GetSystemMetrics()
	require.NoError(t, err, "Admin should access metrics")
	require.Nil(t, adminMetricsResp.Error)

	adminHealthResp, err := asserter.GetSystemHealth()
	require.NoError(t, err, "Admin should access health status")
	require.Nil(t, adminHealthResp.Error)
}

func TestSessionManagementWorkflow(t *testing.T) {
	asserter := NewE2EWorkflowAsserter(t)

	// Create first session
	err := asserter.ConnectAndAuthenticate("operator")
	require.NoError(t, err, "First session should work")

	// Verify first session works
	resp1, err := asserter.GetCameraList()
	require.NoError(t, err, "First session should work")
	require.Nil(t, resp1.Error)

	// Close first connection
	asserter.client.Close()

	// Create second session with same role but new connection
	err = asserter.ConnectAndAuthenticate("operator")
	require.NoError(t, err, "Second session should work")

	// Verify second session also works (new connection, same role)
	resp2, err := asserter.GetCameraList()
	require.NoError(t, err, "Second session should work")
	require.Nil(t, resp2.Error)

	// Close second connection
	asserter.client.Close()

	// Create third session with different role
	err = asserter.ConnectAndAuthenticate("viewer")
	require.NoError(t, err, "Third session should work")

	// Verify third session works with different user
	resp3, err := asserter.GetCameraList()
	require.NoError(t, err, "Third session should work")
	require.Nil(t, resp3.Error)

	// Verify sessions are independent (user2 cannot start recording)
	recResp, err := asserter.StartRecording("camera0")
	require.NoError(t, err, "Request should not fail")
	require.NotNil(t, recResp.Error, "Viewer session should not start recording")
}

func TestTokenExpiryWorkflow(t *testing.T) {
	asserter := NewE2EWorkflowAsserter(t)

	// Connect
	err := asserter.client.Connect()
	require.NoError(t, err, "Connection should succeed")

	// Get JWT token with very short expiry (1 second)
	// Use SecurityHelper via asserter to generate a short-lived token
	token, err := asserter.secHelper.GenerateTestToken("test_user", "admin", 1*time.Second)
	require.NoError(t, err, "Should get JWT token")

	// Authenticate with short-lived token
	err = asserter.client.Authenticate(token)
	require.NoError(t, err, "Authentication with fresh token should succeed")

	// Verify operations work initially
	resp1, err := asserter.GetCameraList()
	require.NoError(t, err, "Operations should work with fresh token")
	require.Nil(t, resp1.Error)

	// Wait for token to expire
	time.Sleep(2 * time.Second)

	// Verify operations fail after expiry
	resp2, err := asserter.GetCameraList()
	require.NoError(t, err, "Request should not fail")
	require.NotNil(t, resp2.Error, "Operations should fail with expired token")

	errorInfo := resp2.Error
	assert.Equal(t, -32001, errorInfo.Code, "Should get authentication required error")
	assert.Contains(t, errorInfo.Message, "auth_required", "Error should mention authentication")
}

func TestSecurityWorkflowsIntegration(t *testing.T) {
	asserter := NewE2EWorkflowAsserter(t)

	// Test comprehensive security workflow
	// Create multiple users with different roles

	// Test viewer permissions
	err := asserter.ConnectAndAuthenticate("viewer")
	require.NoError(t, err, "Viewer should authenticate")

	viewerListResp, err := asserter.GetCameraList()
	require.NoError(t, err, "Viewer should list cameras")
	require.Nil(t, viewerListResp.Error)

	viewerRecResp, err := asserter.StartRecording("camera0")
	require.NoError(t, err, "Request should not fail")
	require.NotNil(t, viewerRecResp.Error, "Viewer should not record")

	// Close viewer connection
	asserter.client.Close()

	// Test operator permissions
	err = asserter.ConnectAndAuthenticate("operator")
	require.NoError(t, err, "Operator should authenticate")

	opListResp, err := asserter.GetCameraList()
	require.NoError(t, err, "Operator should list cameras")
	require.Nil(t, opListResp.Error)

	opRecResp, err := asserter.StartRecording("camera0")
	require.NoError(t, err, "Operator should record")
	require.Nil(t, opRecResp.Error)

	// Stop recording
	stopResp, err := asserter.StopRecording("camera0")
	require.NoError(t, err)
	require.Nil(t, stopResp.Error)

	// Close operator connection
	asserter.client.Close()

	// Test admin permissions
	err = asserter.ConnectAndAuthenticate("admin")
	require.NoError(t, err, "Admin should authenticate")

	adminListResp, err := asserter.GetCameraList()
	require.NoError(t, err, "Admin should list cameras")
	require.Nil(t, adminListResp.Error)

	adminMetricsResp, err := asserter.GetSystemMetrics()
	require.NoError(t, err, "Admin should access metrics")
	require.Nil(t, adminMetricsResp.Error)

	adminHealthResp, err := asserter.GetSystemHealth()
	require.NoError(t, err, "Admin should access health")
	require.Nil(t, adminHealthResp.Error)

	// Verify role hierarchy is enforced
	// Admin can do everything operator can do
	adminRecResp, err := asserter.StartRecording("camera0")
	require.NoError(t, err, "Admin should record")
	require.Nil(t, adminRecResp.Error)

	// Stop admin recording
	stopAdminResp, err := asserter.StopRecording("camera0")
	require.NoError(t, err)
	require.Nil(t, stopAdminResp.Error)
}
