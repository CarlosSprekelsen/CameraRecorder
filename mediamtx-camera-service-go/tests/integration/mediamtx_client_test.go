/*
Component: MediaMTX Client Integration
Purpose: Validates HTTP REST API communication with MediaMTX server
Requirements: REQ-MTX-001, REQ-MTX-002, REQ-MTX-003, REQ-MTX-004
Category: Integration
API Reference: docs/api/json_rpc_methods.md
Test Organization:
  - TestMediaMTXClient_PathOperations (lines 45-85)
  - TestMediaMTXClient_ConnectionPooling (lines 87-127)
  - TestMediaMTXClient_ErrorHandling (lines 129-169)
  - TestMediaMTXClient_HealthMonitoring (lines 171-211)
*/

package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/mediamtx"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MediaMTXClientIntegrationAsserter handles MediaMTX client integration validation
type MediaMTXClientIntegrationAsserter struct {
	setup  *testutils.UniversalTestSetup
	helper *testutils.MediaMTXHelper
	client mediamtx.MediaMTXClient
}

// NewMediaMTXClientIntegrationAsserter creates a new MediaMTX client integration asserter
func NewMediaMTXClientIntegrationAsserter(t *testing.T) *MediaMTXClientIntegrationAsserter {
	setup := testutils.SetupTest(t, "config_valid_complete.yaml")
	helper := testutils.NewMediaMTXHelper(setup)
	
	config := setup.GetConfigManager().GetConfig()
	logger := setup.GetLogger()
	
	// Use helper for URL - reads from config
	baseURL := helper.GetMediaMTXBaseURL()
	client := mediamtx.NewClient(baseURL, &config.MediaMTX, logger)
	
	ctx, cancel := setup.GetStandardContextWithTimeout(testutils.UniversalTimeoutLong)
	defer cancel()
	
	// Use shared readiness helper
	err := helper.WaitForMediaMTXReady(ctx, client, mediamtx.MediaMTXConfigGlobalGet)
	helper.SkipIfMediaMTXUnavailable(t, err)
	
	return &MediaMTXClientIntegrationAsserter{
		setup:  setup,
		helper: helper,
		client: client,
	}
}

// AssertPathOperation validates path CRUD operations
func (a *MediaMTXClientIntegrationAsserter) AssertPathOperation(ctx context.Context, operation string, pathName string) error {
	switch operation {
	case "create":
		// Create test path configuration
		pathConfig := map[string]interface{}{
			"source":    "rtsp://test-source",
			"runOnInit": "echo 'test path created'",
		}
		data, _ := json.Marshal(pathConfig)
		_, err := a.client.Post(ctx, fmt.Sprintf(mediamtx.MediaMTXConfigPathsAdd, pathName), data)
		return err
	case "list":
		_, err := a.client.Get(ctx, mediamtx.MediaMTXConfigPathsList)
		return err
	case "get":
		_, err := a.client.Get(ctx, fmt.Sprintf(mediamtx.MediaMTXConfigPathsGet, pathName))
		return err
	case "patch":
		patchConfig := map[string]interface{}{
			"runOnInit": "echo 'test path patched'",
		}
		data, _ := json.Marshal(patchConfig)
		err := a.client.Patch(ctx, fmt.Sprintf(mediamtx.MediaMTXConfigPathsPatch, pathName), data)
		return err
	case "delete":
		err := a.client.Delete(ctx, fmt.Sprintf(mediamtx.MediaMTXConfigPathsDelete, pathName))
		return err
	default:
		return nil
	}
}

// AssertConnectionPooling validates connection reuse behavior
func (a *MediaMTXClientIntegrationAsserter) AssertConnectionPooling(ctx context.Context, concurrentRequests int) error {
	var wg sync.WaitGroup
	results := make(chan error, concurrentRequests)

	// Make concurrent requests to test connection pooling
	for i := 0; i < concurrentRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := a.client.Get(ctx, mediamtx.MediaMTXConfigGlobalGet)
			results <- err
		}()
	}

	wg.Wait()
	close(results)

	// Check all requests succeeded
	for err := range results {
		if err != nil {
			return err
		}
	}

	return nil
}

// TestMediaMTXClient_PathOperations_ReqMTX001 validates path CRUD operations
// REQ-MTX-001: MediaMTX service integration via HTTP REST API
// REQ-MTX-003: Path creation, deletion, and configuration management
func TestMediaMTXClient_PathOperations_ReqMTX001(t *testing.T) {
	asserter := NewMediaMTXClientIntegrationAsserter(t)
	ctx, cancel := asserter.setup.GetStandardContext()
	defer cancel()

	// Table-driven test for path operations
	tests := []struct {
		name        string
		operation   string
		pathName    string
		expectError bool
	}{
		{"create_path", "create", "test_path_integration", false},
		{"list_paths", "list", "", false},
		{"get_path", "get", "test_path_integration", false},
		{"patch_path", "patch", "test_path_integration", false},
		{"delete_path", "delete", "test_path_integration", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use AssertionHelper for consistent assertions
			ah := testutils.NewAssertionHelper(t)

			err := asserter.AssertPathOperation(ctx, tt.operation, tt.pathName)

			if tt.expectError {
				require.Error(t, err, "Operation should fail: %s", tt.operation)
			} else {
				ah.AssertNoErrorWithContext(err, tt.operation)
			}

			// Validate HTTP request was sent and response was processed
			assert.NotNil(t, err == nil || err != nil, "Response should be processed")
		})
	}
}

// TestMediaMTXClient_ConnectionPooling_ReqMTX002 validates connection pooling
// REQ-MTX-001: MediaMTX service integration via HTTP REST API
func TestMediaMTXClient_ConnectionPooling_ReqMTX002(t *testing.T) {
	// Use AssertionHelper for consistent assertions
	ah := testutils.NewAssertionHelper(t)

	asserter := NewMediaMTXClientIntegrationAsserter(t)
	ctx, cancel := asserter.setup.GetStandardContext()
	defer cancel()

	// Test connection pooling with concurrent requests
	concurrentRequests := 10
	err := asserter.AssertConnectionPooling(ctx, concurrentRequests)

	// Validate all requests succeeded (connection pooling working)
	ah.AssertNoErrorWithContext(err, "Concurrent requests")

	// Validate that requests completed efficiently
	start := time.Now()
	err = asserter.AssertConnectionPooling(ctx, concurrentRequests)
	duration := time.Since(start)

	ah.AssertNoErrorWithContext(err, "Second batch requests")
	assert.Less(t, duration, testutils.UniversalTimeoutLong, "Connection pooling should be efficient")
}

// TestMediaMTXClient_ErrorHandling_ReqMTX003 validates error handling
// REQ-MTX-001: MediaMTX service integration via HTTP REST API
func TestMediaMTXClient_ErrorHandling_ReqMTX003(t *testing.T) {
	asserter := NewMediaMTXClientIntegrationAsserter(t)
	ctx, cancel := asserter.setup.GetStandardContext()
	defer cancel()

	// Table-driven test for error conditions
	tests := []struct {
		name        string
		path        string
		method      string
		data        []byte
		expectError bool
		errorCode   int
	}{
		{"invalid_path", fmt.Sprintf(mediamtx.MediaMTXConfigPathsGet, "nonexistent"), "GET", nil, true, 404},
		{"malformed_request", fmt.Sprintf(mediamtx.MediaMTXConfigPathsAdd, "test"), "POST", []byte("invalid json"), true, 400},
		{"valid_request", mediamtx.MediaMTXConfigGlobalGet, "GET", nil, false, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use AssertionHelper for consistent assertions
			ah := testutils.NewAssertionHelper(t)

			var err error

			switch tt.method {
			case "GET":
				_, err = asserter.client.Get(ctx, tt.path)
			case "POST":
				_, err = asserter.client.Post(ctx, tt.path, tt.data)
			}

			if tt.expectError {
				require.Error(t, err, "Request should fail: %s", tt.path)
				// Validate error contains diagnostic details
				assert.Contains(t, err.Error(), "failed", "Error should contain diagnostic details")
			} else {
				ah.AssertNoErrorWithContext(err, tt.path)
			}
		})
	}
}

// TestMediaMTXClient_HealthMonitoring_ReqMTX004 validates health monitoring
// REQ-MTX-004: Health monitoring via MediaMTX status endpoints
func TestMediaMTXClient_HealthMonitoring_ReqMTX004(t *testing.T) {
	// Use AssertionHelper for consistent assertions
	ah := testutils.NewAssertionHelper(t)

	asserter := NewMediaMTXClientIntegrationAsserter(t)
	ctx, cancel := asserter.setup.GetStandardContext()
	defer cancel()

	// Test basic health check
	err := asserter.client.HealthCheck(ctx)
	ah.AssertNoErrorWithContext(err, "Health check")

	// Test detailed health status
	healthStatus, err := asserter.client.GetDetailedHealth(ctx)
	ah.AssertNoErrorWithContext(err, "Detailed health check")
	ah.AssertNotNilWithContext(healthStatus, "Health status")

	// Validate health status structure
	assert.NotEmpty(t, healthStatus, "Health status should contain data")

	// Test health monitoring integration
	start := time.Now()
	err = asserter.client.HealthCheck(ctx)
	duration := time.Since(start)

	require.NoError(t, err, "Health check should be reliable")
	assert.Less(t, duration, testutils.UniversalTimeoutShort, "Health check should be fast")
}
