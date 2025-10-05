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
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/testutils"
	testutilsshared "github.com/camerarecorder/mediamtx-camera-service-go/tests/testutils"
)

// E2EFixture provides E2E testing utilities backed by real server and clients
// Uses tests/testutils/websocket_server.go for server creation
type E2EFixture struct {
	t            *testing.T
	setup        *testutils.UniversalTestSetup
	serverHelper *testutilsshared.WebSocketServerHelper
	client       *testutils.WebSocketTestClient
	secHelper    *testutils.SecurityHelper
}

const (
	DefaultCameraID = "camera0"
	RoleViewer      = "viewer"
	RoleOperator    = "operator"
	RoleAdmin       = "admin"
)

// NewE2EFixture creates the E2E fixture using shared infrastructure
func NewE2EFixture(t *testing.T) *E2EFixture {
	// Use shared test infrastructure from tests/testutils/
	setup := testutils.SetupTest(t, "config_valid_complete.yaml")
	serverHelper := testutilsshared.NewWebSocketServerHelper(t, setup)

	// Create WebSocket client using the shared server URL
	client := testutils.NewWebSocketTestClient(t, serverHelper.GetServerURL())
	secHelper := testutils.NewSecurityHelper(t, setup)

	asserter := &E2EFixture{
		t:            t,
		setup:        setup,
		serverHelper: serverHelper,
		client:       client,
		secHelper:    secHelper,
	}

	t.Cleanup(func() { asserter.Cleanup() })
	return asserter
}

// ConnectAndAuthenticate establishes connection and authenticates
func (a *E2EFixture) ConnectAndAuthenticate(role string) error {
	// Connect using proven client
	err := a.client.Connect()
	if err != nil {
		return err
	}

	// Get JWT token using proven helper
	token, err := a.secHelper.GenerateTestToken(testutils.UniversalTestUserID, role, 24*time.Hour)
	if err != nil {
		return err
	}

	// Authenticate using proven client method
	return a.client.Authenticate(token)
}

// RecordingPath builds the absolute expected recording path from config
func (a *E2EFixture) RecordingPath(cameraID, basename string) string {
	cfg := a.setup.GetConfigManager().GetConfig()
	base := cfg.MediaMTX.RecordingsPath
	if base == "" {
		base = cfg.Storage.DefaultPath
	}
	return testutils.BuildRecordingFilePath(
		base,
		cameraID,
		basename,
		cfg.Recording.UseDeviceSubdirs,
		cfg.Recording.RecordFormat,
	)
}

// SnapshotPath builds the absolute expected snapshot path from config
func (a *E2EFixture) SnapshotPath(cameraID, basename string) string {
	cfg := a.setup.GetConfigManager().GetConfig()
	base := cfg.MediaMTX.SnapshotsPath
	if base == "" {
		base = cfg.Storage.DefaultPath
	}
	return testutils.BuildSnapshotFilePath(
		base,
		cameraID,
		basename,
		cfg.Snapshots.UseDeviceSubdirs,
		cfg.Snapshots.Format,
	)
}

// Cleanup delegates to helper cleanup
func (a *E2EFixture) Cleanup() {
	if a.client != nil {
		a.client.Close()
	}
}
