/*
E2E Test Helpers - Thin Adapter to Proven Infrastructure

Provides E2E testing utilities by leveraging the proven WebSocket test infrastructure
from internal/websocket and internal/testutils. This is a thin wrapper that delegates
all real work to the working infrastructure.

Key Principles:
- Thin adapter - all real work delegated to proven infrastructure
- Zero duplication - reuse WebSocketTestHelper and WebSocketTestClient
- Proven patterns - use working authentication and client methods
- Guidelines compliance - tests stay in tests/e2e/ per documentation
*/

package e2e

import (
	"testing"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/testutils"
)

// E2EWorkflowAsserter provides E2E workflow testing using shared infrastructure
// Uses tests/testutils/websocket_server.go for server creation
type E2EWorkflowAsserter struct {
	t            *testing.T
	setup        *testutils.UniversalTestSetup
	serverHelper *sharedutils.WebSocketServerHelper
	client       *testutils.WebSocketTestClient
}

// NewE2EWorkflowAsserter creates E2E workflow asserter using shared infrastructure
func NewE2EWorkflowAsserter(t *testing.T) *E2EWorkflowAsserter {
	// Use shared test infrastructure from tests/testutils/
	setup := testutils.SetupTest(t, "config_valid_complete.yaml")
	serverHelper := sharedutils.NewWebSocketServerHelper(t, setup)

	// Create WebSocket client using the shared server URL
	client := testutils.NewWebSocketTestClient(t, serverHelper.GetServerURL())

	asserter := &E2EWorkflowAsserter{
		t:            t,
		setup:        setup,
		serverHelper: serverHelper,
		client:       client,
	}

	t.Cleanup(func() { asserter.Cleanup() })
	return asserter
}

// ConnectAndAuthenticate establishes connection and authenticates
func (a *E2EWorkflowAsserter) ConnectAndAuthenticate(role string) error {
	// Connect using proven client
	err := a.client.Connect()
	if err != nil {
		return err
	}

	// Get JWT token using proven helper
	token, err := a.helper.GetJWTToken(role)
	if err != nil {
		return err
	}

	// Authenticate using proven client method
	return a.client.Authenticate(token)
}

// SendJSONRPC delegates to proven client implementation
func (a *E2EWorkflowAsserter) SendJSONRPC(method string, params interface{}) (*testutils.JSONRPCResponse, error) {
	return a.client.SendJSONRPC(method, params)
}

// GetCameraList workflow helper
func (a *E2EWorkflowAsserter) GetCameraList() (*testutils.JSONRPCResponse, error) {
	return a.client.GetCameraList()
}

// GetCameraStatus workflow helper
func (a *E2EWorkflowAsserter) GetCameraStatus(device string) (*testutils.JSONRPCResponse, error) {
	return a.client.GetCameraStatus(device)
}

// GetCameraCapabilities workflow helper
func (a *E2EWorkflowAsserter) GetCameraCapabilities(device string) (*testutils.JSONRPCResponse, error) {
	return a.client.GetCameraCapabilities(device)
}

// StartRecording workflow helper
func (a *E2EWorkflowAsserter) StartRecording(device string) (*testutils.JSONRPCResponse, error) {
	return a.client.StartRecording(device)
}

// StartRecordingWithDuration workflow helper
func (a *E2EWorkflowAsserter) StartRecordingWithDuration(device string, duration int) (*testutils.JSONRPCResponse, error) {
	return a.client.StartRecordingWithDuration(device, duration)
}

// StopRecording workflow helper
func (a *E2EWorkflowAsserter) StopRecording(device string) (*testutils.JSONRPCResponse, error) {
	return a.client.StopRecording(device)
}

// ListRecordings workflow helper
func (a *E2EWorkflowAsserter) ListRecordings() (*testutils.JSONRPCResponse, error) {
	return a.client.ListRecordings()
}

// TakeSnapshot workflow helper
func (a *E2EWorkflowAsserter) TakeSnapshot(device string) (*testutils.JSONRPCResponse, error) {
	return a.client.TakeSnapshot(device)
}

// TakeSnapshotWithFormat workflow helper
func (a *E2EWorkflowAsserter) TakeSnapshotWithFormat(device string, format string, quality int) (*testutils.JSONRPCResponse, error) {
	return a.client.TakeSnapshotWithFormat(device, format, quality)
}

// ListSnapshots workflow helper
func (a *E2EWorkflowAsserter) ListSnapshots() (*testutils.JSONRPCResponse, error) {
	return a.client.ListSnapshots()
}

// GetSystemHealth workflow helper
func (a *E2EWorkflowAsserter) GetSystemHealth() (*testutils.JSONRPCResponse, error) {
	return a.client.GetSystemHealth()
}

// GetSystemMetrics workflow helper
func (a *E2EWorkflowAsserter) GetSystemMetrics() (*testutils.JSONRPCResponse, error) {
	return a.client.GetSystemMetrics()
}

// Cleanup delegates to helper cleanup
func (a *E2EWorkflowAsserter) Cleanup() {
	if a.client != nil {
		a.client.Close()
	}
}
