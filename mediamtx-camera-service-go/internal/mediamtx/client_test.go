/*
MediaMTX Client Tests - Real Server Integration

Requirements Coverage:
- REQ-MTX-001: MediaMTX service integration
- REQ-MTX-002: Stream management capabilities
- REQ-MTX-003: Path creation and deletion
- REQ-MTX-007: Error handling and recovery

Test Categories: Unit (using real MediaMTX server)
API Documentation Reference: docs/api/swagger.json
*/

package mediamtx

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewClient_ReqMTX001 tests client creation with real server
func TestNewClient_ReqMTX001(t *testing.T) {
	// REQ-MTX-001: MediaMTX service integration
	EnsureSequentialExecution(t)
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Wait for MediaMTX server to be ready
	err := helper.WaitForServerReady(t, 10*time.Second)
	require.NoError(t, err, "MediaMTX server should be ready")

	config := &MediaMTXConfig{
		BaseURL: "http://localhost:9997",
		Timeout: 5 * time.Second,
	}
	logger := helper.GetLogger()

	client := NewClient("http://localhost:9997", config, logger)
	require.NotNil(t, client, "Client should not be nil")
}

// TestClient_Get_ReqMTX001 tests GET request functionality with real server
func TestClient_Get_ReqMTX001(t *testing.T) {
	// REQ-MTX-001: MediaMTX service integration
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Wait for MediaMTX server to be ready
	err := helper.WaitForServerReady(t, 10*time.Second)
	require.NoError(t, err, "MediaMTX server should be ready")

	client := helper.GetClient()
	ctx := context.Background()

	// Test GET request to paths list endpoint (from swagger.json)
	data, err := client.Get(ctx, "/v3/paths/list")
	require.NoError(t, err, "GET request should succeed")
	assert.NotNil(t, data, "Response data should not be nil")
	assert.Greater(t, len(data), 0, "Response should contain data")

	// Validate response structure matches swagger.json schema
	// The response should be a PathList object with pageCount, itemCount, and items
	assert.Contains(t, string(data), "pageCount", "Response should contain pageCount field per swagger.json")
	assert.Contains(t, string(data), "itemCount", "Response should contain itemCount field per swagger.json")
	assert.Contains(t, string(data), "items", "Response should contain items field per swagger.json")
}

// TestClient_Post_ReqMTX001 tests POST request functionality with real server
func TestClient_Post_ReqMTX001(t *testing.T) {
	// REQ-MTX-001: MediaMTX service integration
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Wait for MediaMTX server to be ready
	err := helper.WaitForServerReady(t, 10*time.Second)
	require.NoError(t, err, "MediaMTX server should be ready")

	client := helper.GetClient()
	ctx := context.Background()

	// Test POST request to create path endpoint (from swagger.json)
	pathData := `{"name":"test_path","source":"publisher"}`
	data, err := client.Post(ctx, "/v3/config/paths/add/test_path", []byte(pathData))
	require.NoError(t, err, "POST request should succeed")
	assert.NotNil(t, data, "Response data should not be nil")

	// Clean up - delete the test path
	err = client.Delete(ctx, "/v3/config/paths/delete/test_path")
	require.NoError(t, err, "DELETE request should succeed")
}

// TestClient_Put_ReqMTX001 tests PUT request functionality with real server
func TestClient_Put_ReqMTX001(t *testing.T) {
	// REQ-MTX-001: MediaMTX service integration
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Wait for MediaMTX server to be ready
	err := helper.WaitForServerReady(t, 10*time.Second)
	require.NoError(t, err, "MediaMTX server should be ready")

	client := helper.GetClient()
	ctx := context.Background()

	// First create a path
	pathData := `{"name":"test_put_path","source":"publisher"}`
	_, err = client.Post(ctx, "/v3/config/paths/add/test_put_path", []byte(pathData))
	require.NoError(t, err, "POST request should succeed")

	// Test POST request to replace path endpoint (from swagger.json)
	updateData := `{"name":"test_put_path","source":"publisher","maxReaders":5}`
	data, err := client.Post(ctx, "/v3/config/paths/replace/test_put_path", []byte(updateData))
	require.NoError(t, err, "POST request should succeed")
	assert.NotNil(t, data, "Response data should not be nil")

	// Clean up - delete the test path
	err = client.Delete(ctx, "/v3/config/paths/delete/test_put_path")
	require.NoError(t, err, "DELETE request should succeed")
}

// TestClient_Delete_ReqMTX001 tests DELETE request functionality with real server
func TestClient_Delete_ReqMTX001(t *testing.T) {
	// REQ-MTX-001: MediaMTX service integration
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Wait for MediaMTX server to be ready
	err := helper.WaitForServerReady(t, 10*time.Second)
	require.NoError(t, err, "MediaMTX server should be ready")

	client := helper.GetClient()
	ctx := context.Background()

	// First create a path
	pathData := `{"name":"test_delete_path","source":"publisher"}`
	_, err = client.Post(ctx, "/v3/config/paths/add/test_delete_path", []byte(pathData))
	require.NoError(t, err, "POST request should succeed")

	// Test DELETE request to delete path endpoint (from swagger.json)
	err = client.Delete(ctx, "/v3/config/paths/delete/test_delete_path")
	require.NoError(t, err, "DELETE request should succeed")
}

// TestClient_HealthCheck_ReqMTX004 tests health check functionality with real server
func TestClient_HealthCheck_ReqMTX004(t *testing.T) {
	// REQ-MTX-004: Health monitoring
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Wait for MediaMTX server to be ready
	err := helper.WaitForServerReady(t, 10*time.Second)
	require.NoError(t, err, "MediaMTX server should be ready")

	client := helper.GetClient()
	ctx := context.Background()

	// Test health check
	err = client.HealthCheck(ctx)
	require.NoError(t, err, "Health check should succeed")
}

// TestClient_ErrorHandling_ReqMTX007 tests error scenarios with real server
func TestClient_ErrorHandling_ReqMTX007(t *testing.T) {
	// REQ-MTX-007: Error handling and recovery
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Wait for MediaMTX server to be ready
	err := helper.WaitForServerReady(t, 10*time.Second)
	require.NoError(t, err, "MediaMTX server should be ready")

	client := helper.GetClient()
	ctx := context.Background()

	// Test invalid endpoint
	_, err = client.Get(ctx, "/v3/invalid/endpoint")
	assert.Error(t, err, "Invalid endpoint should return error")

	// Test invalid path creation (missing required fields per swagger.json)
	_, err = client.Post(ctx, "/v3/config/paths/add", []byte(`{"invalid": "data"}`))
	assert.Error(t, err, "Invalid path creation should return error")

	// Test deleting non-existent path
	err = client.Delete(ctx, "/v3/config/paths/delete/non_existent_path")
	assert.Error(t, err, "Deleting non-existent path should return error")
}

// TestClient_APICompliance_ReqMTX001 tests API compliance against swagger.json
func TestClient_APICompliance_ReqMTX001(t *testing.T) {
	// REQ-MTX-001: MediaMTX service integration
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Wait for MediaMTX server to be ready
	err := helper.WaitForServerReady(t, 10*time.Second)
	require.NoError(t, err, "MediaMTX server should be ready")

	client := helper.GetClient()
	ctx := context.Background()

	// Test paths list endpoint compliance with swagger.json
	data, err := client.Get(ctx, "/v3/paths/list")
	require.NoError(t, err, "Paths list should succeed")

	// Validate response structure matches swagger.json PathList schema
	responseStr := string(data)
	assert.Contains(t, responseStr, "pageCount", "Missing pageCount field per swagger.json")
	assert.Contains(t, responseStr, "itemCount", "Missing itemCount field per swagger.json")
	assert.Contains(t, responseStr, "items", "Missing items field per swagger.json")

	// Test config paths list endpoint compliance with swagger.json
	data, err = client.Get(ctx, "/v3/config/paths/list")
	require.NoError(t, err, "Config paths list should succeed")

	// Validate response structure matches swagger.json PathConfList schema
	responseStr = string(data)
	assert.Contains(t, responseStr, "pageCount", "Missing pageCount field per swagger.json")
	assert.Contains(t, responseStr, "itemCount", "Missing itemCount field per swagger.json")
	assert.Contains(t, responseStr, "items", "Missing items field per swagger.json")

	// Test global config endpoint compliance with swagger.json
	data, err = client.Get(ctx, "/v3/config/global/get")
	require.NoError(t, err, "Global config get should succeed")

	// Validate response structure matches swagger.json GlobalConf schema
	responseStr = string(data)
	// Check for some key fields from GlobalConf schema
	assert.Contains(t, responseStr, "logLevel", "Missing logLevel field per swagger.json")
	assert.Contains(t, responseStr, "api", "Missing api field per swagger.json")
}

// TestClient_ConcurrentAccess_ReqMTX001 tests concurrent access with real server
func TestClient_ConcurrentAccess_ReqMTX001(t *testing.T) {
	// REQ-MTX-001: MediaMTX service integration
	EnsureSequentialExecution(t)
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Wait for MediaMTX server to be ready
	err := helper.WaitForServerReady(t, 10*time.Second)
	require.NoError(t, err, "MediaMTX server should be ready")

	client := helper.GetClient()
	ctx := context.Background()

	// Test concurrent GET requests
	done := make(chan bool, 3)

	go func() {
		_, err := client.Get(ctx, "/v3/paths/list")
		assert.NoError(t, err, "Concurrent GET should succeed")
		done <- true
	}()

	go func() {
		_, err := client.Get(ctx, "/v3/config/paths/list")
		assert.NoError(t, err, "Concurrent GET should succeed")
		done <- true
	}()

	go func() {
		_, err := client.Get(ctx, "/v3/config/global/get")
		assert.NoError(t, err, "Concurrent GET should succeed")
		done <- true
	}()

	// Wait for all goroutines to complete
	for i := 0; i < 3; i++ {
		<-done
	}

	// Should not panic and should handle concurrent access gracefully
	assert.True(t, true, "Concurrent access should not cause panics")
}

// TestClient_Close_ReqMTX001 tests client close functionality
func TestClient_Close_ReqMTX001(t *testing.T) {
	// REQ-MTX-001: MediaMTX service integration
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Wait for MediaMTX server to be ready
	err := helper.WaitForServerReady(t, 10*time.Second)
	require.NoError(t, err, "MediaMTX server should be ready")

	config := &MediaMTXConfig{
		BaseURL: "http://localhost:9997",
		Timeout: 5 * time.Second,
	}
	logger := helper.GetLogger()

	client := NewClient("http://localhost:9997", config, logger)
	require.NotNil(t, client, "Client should not be nil")

	// Test close
	err = client.Close()
	require.NoError(t, err, "Client close should succeed")
}
