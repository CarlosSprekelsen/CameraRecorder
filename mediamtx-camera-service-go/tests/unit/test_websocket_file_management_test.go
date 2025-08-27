//go:build unit
// +build unit

/*
WebSocket JSON-RPC file management methods unit tests.

Tests validate file management method implementations against ground truth API documentation.
Tests are designed to FAIL if implementation doesn't match API documentation exactly.

Requirements Coverage:
- REQ-FILE-001: list_recordings method for recording enumeration
- REQ-FILE-002: list_snapshots method for snapshot enumeration
- REQ-FILE-003: delete_recording method for recording deletion
- REQ-FILE-004: delete_snapshot method for snapshot deletion
- REQ-FILE-005: get_storage_info method for storage monitoring
- REQ-FILE-006: set_retention_policy method for policy management
- REQ-FILE-007: cleanup_old_files method for file cleanup
- REQ-API-011: API methods respond within specified time limits

Test Categories: Unit
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package websocket_test

import (
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/camera"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/security"
	ws "github.com/camerarecorder/mediamtx-camera-service-go/internal/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestListRecordingsMethodImplementation tests list_recordings method implementation
// REQ-FILE-001: list_recordings method for recording enumeration
func TestListRecordingsMethodImplementation(t *testing.T) {
	/*
		Unit Test for list_recordings method implementation

		API Documentation Reference: docs/api/json_rpc_methods.md
		Method: list_recordings
		Expected Response: {"jsonrpc": "2.0", "result": {"recordings": [{"filename": "recording_20250115_120000.mp4", "size": 1024000, "created": "2025-01-15T12:00:00Z", "duration": 300, "camera": "/dev/video0"}], "total_count": 1, "total_size": 1024000}, "id": 1}
		Performance Target: <200ms response time
	*/

	// Setup real components with proper dependency injection
	configManager := config.NewConfigManager()
	logger := logging.NewLogger("test")

	// Create concrete implementations for camera interfaces
	deviceChecker := &camera.RealDeviceChecker{}
	commandExecutor := &camera.RealV4L2CommandExecutor{}
	infoParser := &camera.RealDeviceInfoParser{}

	cameraMonitor := camera.NewHybridCameraMonitor(configManager, logger, deviceChecker, commandExecutor, infoParser)
	jwtHandler, err := security.NewJWTHandler("test-secret-key-for-testing-only")
	require.NoError(t, err)

	server := ws.NewWebSocketServer(configManager, logger, cameraMonitor, jwtHandler)
	require.NotNil(t, server)

	// Create test client
	client := &ws.ClientConnection{
		ClientID:      "test-client",
		Authenticated: true,
		UserID:        "test-user",
		Role:          "admin",
		ConnectedAt:   time.Now(),
	}

	// Test list_recordings method
	params := map[string]interface{}{}

	// Measure response time
	startTime := time.Now()
	response, err := server.MethodListRecordings(params, client)
	responseTime := time.Since(startTime)

	require.NoError(t, err)
	require.NotNil(t, response)

	// Validate response format per API documentation
	assert.Equal(t, "2.0", response.JSONRPC, "JSON-RPC version should be 2.0")
	assert.Nil(t, response.Error, "Response should not have error")
	assert.NotNil(t, response.Result, "Response should have result")

	// Validate result structure per API documentation
	result, ok := response.Result.(map[string]interface{})
	require.True(t, ok, "Result should be a map")

	// Check required fields from API documentation
	requiredFields := []string{"recordings", "total_count", "total_size"}
	for _, field := range requiredFields {
		assert.Contains(t, result, field, "Missing required field '%s' per API documentation", field)
	}

	// Validate data types per API documentation
	assert.IsType(t, []interface{}{}, result["recordings"], "recordings should be array")
	assert.IsType(t, float64(0), result["total_count"], "total_count should be float64")
	assert.IsType(t, float64(0), result["total_size"], "total_size should be float64")

	// Validate performance target
	assert.Less(t, responseTime, 200*time.Millisecond, "list_recordings response should be <200ms per API documentation")
}

// TestListSnapshotsMethodImplementation tests list_snapshots method implementation
// REQ-FILE-002: list_snapshots method for snapshot enumeration
func TestListSnapshotsMethodImplementation(t *testing.T) {
	/*
		Unit Test for list_snapshots method implementation

		API Documentation Reference: docs/api/json_rpc_methods.md
		Method: list_snapshots
		Expected Response: {"jsonrpc": "2.0", "result": {"snapshots": [{"filename": "snapshot_20250115_120000.jpg", "size": 51200, "created": "2025-01-15T12:00:00Z", "camera": "/dev/video0"}], "total_count": 1, "total_size": 51200}, "id": 1}
		Performance Target: <200ms response time
	*/

	// Setup real components with proper dependency injection
	configManager := config.NewConfigManager()
	logger := logging.NewLogger("test")

	// Create concrete implementations for camera interfaces
	deviceChecker := &camera.RealDeviceChecker{}
	commandExecutor := &camera.RealV4L2CommandExecutor{}
	infoParser := &camera.RealDeviceInfoParser{}

	cameraMonitor := camera.NewHybridCameraMonitor(configManager, logger, deviceChecker, commandExecutor, infoParser)
	jwtHandler, err := security.NewJWTHandler("test-secret-key-for-testing-only")
	require.NoError(t, err)

	server := ws.NewWebSocketServer(configManager, logger, cameraMonitor, jwtHandler)
	require.NotNil(t, server)

	// Create test client
	client := &ws.ClientConnection{
		ClientID:      "test-client",
		Authenticated: true,
		UserID:        "test-user",
		Role:          "admin",
		ConnectedAt:   time.Now(),
	}

	// Test list_snapshots method
	params := map[string]interface{}{}

	// Measure response time
	startTime := time.Now()
	response, err := server.MethodListSnapshots(params, client)
	responseTime := time.Since(startTime)

	require.NoError(t, err)
	require.NotNil(t, response)

	// Validate response format per API documentation
	assert.Equal(t, "2.0", response.JSONRPC, "JSON-RPC version should be 2.0")
	assert.Nil(t, response.Error, "Response should not have error")
	assert.NotNil(t, response.Result, "Response should have result")

	// Validate result structure per API documentation
	result, ok := response.Result.(map[string]interface{})
	require.True(t, ok, "Result should be a map")

	// Check required fields from API documentation
	requiredFields := []string{"snapshots", "total_count", "total_size"}
	for _, field := range requiredFields {
		assert.Contains(t, result, field, "Missing required field '%s' per API documentation", field)
	}

	// Validate data types per API documentation
	assert.IsType(t, []interface{}{}, result["snapshots"], "snapshots should be array")
	assert.IsType(t, float64(0), result["total_count"], "total_count should be float64")
	assert.IsType(t, float64(0), result["total_size"], "total_size should be float64")

	// Validate performance target
	assert.Less(t, responseTime, 200*time.Millisecond, "list_snapshots response should be <200ms per API documentation")
}

// TestDeleteRecordingMethodImplementation tests delete_recording method implementation
// REQ-FILE-003: delete_recording method for recording deletion
func TestDeleteRecordingMethodImplementation(t *testing.T) {
	/*
		Unit Test for delete_recording method implementation

		API Documentation Reference: docs/api/json_rpc_methods.md
		Method: delete_recording
		Expected Response: {"jsonrpc": "2.0", "result": {"deleted": true, "filename": "recording_20250115_120000.mp4", "size_freed": 1024000}, "id": 1}
		Performance Target: <500ms response time
	*/

	// Setup real components with proper dependency injection
	configManager := config.NewConfigManager()
	logger := logging.NewLogger("test")

	// Create concrete implementations for camera interfaces
	deviceChecker := &camera.RealDeviceChecker{}
	commandExecutor := &camera.RealV4L2CommandExecutor{}
	infoParser := &camera.RealDeviceInfoParser{}

	cameraMonitor := camera.NewHybridCameraMonitor(configManager, logger, deviceChecker, commandExecutor, infoParser)
	jwtHandler, err := security.NewJWTHandler("test-secret-key-for-testing-only")
	require.NoError(t, err)

	server := ws.NewWebSocketServer(configManager, logger, cameraMonitor, jwtHandler)
	require.NotNil(t, server)

	// Create test client
	client := &ws.ClientConnection{
		ClientID:      "test-client",
		Authenticated: true,
		UserID:        "test-user",
		Role:          "admin",
		ConnectedAt:   time.Now(),
	}

	// Test delete_recording method with valid filename
	params := map[string]interface{}{
		"filename": "test_recording.mp4",
	}

	// Measure response time
	startTime := time.Now()
	response, err := server.MethodDeleteRecording(params, client)
	responseTime := time.Since(startTime)

	require.NoError(t, err)
	require.NotNil(t, response)

	// Validate response format per API documentation
	assert.Equal(t, "2.0", response.JSONRPC, "JSON-RPC version should be 2.0")
	assert.Nil(t, response.Error, "Response should not have error")
	assert.NotNil(t, response.Result, "Response should have result")

	// Validate result structure per API documentation
	result, ok := response.Result.(map[string]interface{})
	require.True(t, ok, "Result should be a map")

	// Check required fields from API documentation
	requiredFields := []string{"deleted", "filename", "size_freed"}
	for _, field := range requiredFields {
		assert.Contains(t, result, field, "Missing required field '%s' per API documentation", field)
	}

	// Validate data types per API documentation
	assert.IsType(t, true, result["deleted"], "deleted should be bool")
	assert.IsType(t, "", result["filename"], "filename should be string")
	assert.IsType(t, float64(0), result["size_freed"], "size_freed should be float64")

	// Validate performance target
	assert.Less(t, responseTime, 500*time.Millisecond, "delete_recording response should be <500ms per API documentation")

	// Test error case - missing filename parameter
	params = map[string]interface{}{}
	response, err = server.MethodDeleteRecording(params, client)

	require.NoError(t, err)
	require.NotNil(t, response)

	// Should return error for missing filename parameter
	assert.NotNil(t, response.Error, "Should return error for missing filename parameter")
	assert.Equal(t, -32602, response.Error.Code, "Should return Invalid Params error code")
}

// TestDeleteSnapshotMethodImplementation tests delete_snapshot method implementation
// REQ-FILE-004: delete_snapshot method for snapshot deletion
func TestDeleteSnapshotMethodImplementation(t *testing.T) {
	/*
		Unit Test for delete_snapshot method implementation

		API Documentation Reference: docs/api/json_rpc_methods.md
		Method: delete_snapshot
		Expected Response: {"jsonrpc": "2.0", "result": {"deleted": true, "filename": "snapshot_20250115_120000.jpg", "size_freed": 51200}, "id": 1}
		Performance Target: <500ms response time
	*/

	// Setup real components with proper dependency injection
	configManager := config.NewConfigManager()
	logger := logging.NewLogger("test")

	// Create concrete implementations for camera interfaces
	deviceChecker := &camera.RealDeviceChecker{}
	commandExecutor := &camera.RealV4L2CommandExecutor{}
	infoParser := &camera.RealDeviceInfoParser{}

	cameraMonitor := camera.NewHybridCameraMonitor(configManager, logger, deviceChecker, commandExecutor, infoParser)
	jwtHandler, err := security.NewJWTHandler("test-secret-key-for-testing-only")
	require.NoError(t, err)

	server := ws.NewWebSocketServer(configManager, logger, cameraMonitor, jwtHandler)
	require.NotNil(t, server)

	// Create test client
	client := &ws.ClientConnection{
		ClientID:      "test-client",
		Authenticated: true,
		UserID:        "test-user",
		Role:          "admin",
		ConnectedAt:   time.Now(),
	}

	// Test delete_snapshot method with valid filename
	params := map[string]interface{}{
		"filename": "test_snapshot.jpg",
	}

	// Measure response time
	startTime := time.Now()
	response, err := server.MethodDeleteSnapshot(params, client)
	responseTime := time.Since(startTime)

	require.NoError(t, err)
	require.NotNil(t, response)

	// Validate response format per API documentation
	assert.Equal(t, "2.0", response.JSONRPC, "JSON-RPC version should be 2.0")
	assert.Nil(t, response.Error, "Response should not have error")
	assert.NotNil(t, response.Result, "Response should have result")

	// Validate result structure per API documentation
	result, ok := response.Result.(map[string]interface{})
	require.True(t, ok, "Result should be a map")

	// Check required fields from API documentation
	requiredFields := []string{"deleted", "filename", "size_freed"}
	for _, field := range requiredFields {
		assert.Contains(t, result, field, "Missing required field '%s' per API documentation", field)
	}

	// Validate data types per API documentation
	assert.IsType(t, true, result["deleted"], "deleted should be bool")
	assert.IsType(t, "", result["filename"], "filename should be string")
	assert.IsType(t, float64(0), result["size_freed"], "size_freed should be float64")

	// Validate performance target
	assert.Less(t, responseTime, 500*time.Millisecond, "delete_snapshot response should be <500ms per API documentation")

	// Test error case - missing filename parameter
	params = map[string]interface{}{}
	response, err = server.MethodDeleteSnapshot(params, client)

	require.NoError(t, err)
	require.NotNil(t, response)

	// Should return error for missing filename parameter
	assert.NotNil(t, response.Error, "Should return error for missing filename parameter")
	assert.Equal(t, -32602, response.Error.Code, "Should return Invalid Params error code")
}

// TestGetStorageInfoMethodImplementation tests get_storage_info method implementation
// REQ-FILE-005: get_storage_info method for storage monitoring
func TestGetStorageInfoMethodImplementation(t *testing.T) {
	/*
		Unit Test for get_storage_info method implementation

		API Documentation Reference: docs/api/json_rpc_methods.md
		Method: get_storage_info
		Expected Response: {"jsonrpc": "2.0", "result": {"total_space": 107374182400, "used_space": 21474836480, "free_space": 85899345920, "recordings_count": 50, "snapshots_count": 100, "recordings_size": 10737418240, "snapshots_size": 524288000}, "id": 1}
		Performance Target: <100ms response time
	*/

	// Setup real components with proper dependency injection
	configManager := config.NewConfigManager()
	logger := logging.NewLogger("test")

	// Create concrete implementations for camera interfaces
	deviceChecker := &camera.RealDeviceChecker{}
	commandExecutor := &camera.RealV4L2CommandExecutor{}
	infoParser := &camera.RealDeviceInfoParser{}

	cameraMonitor := camera.NewHybridCameraMonitor(configManager, logger, deviceChecker, commandExecutor, infoParser)
	jwtHandler, err := security.NewJWTHandler("test-secret-key-for-testing-only")
	require.NoError(t, err)

	server := ws.NewWebSocketServer(configManager, logger, cameraMonitor, jwtHandler)
	require.NotNil(t, server)

	// Create test client
	client := &ws.ClientConnection{
		ClientID:      "test-client",
		Authenticated: true,
		UserID:        "test-user",
		Role:          "admin",
		ConnectedAt:   time.Now(),
	}

	// Test get_storage_info method
	params := map[string]interface{}{}

	// Measure response time
	startTime := time.Now()
	response, err := server.MethodGetStorageInfo(params, client)
	responseTime := time.Since(startTime)

	require.NoError(t, err)
	require.NotNil(t, response)

	// Validate response format per API documentation
	assert.Equal(t, "2.0", response.JSONRPC, "JSON-RPC version should be 2.0")
	assert.Nil(t, response.Error, "Response should not have error")
	assert.NotNil(t, response.Result, "Response should have result")

	// Validate result structure per API documentation
	result, ok := response.Result.(map[string]interface{})
	require.True(t, ok, "Result should be a map")

	// Check required fields from API documentation
	requiredFields := []string{"total_space", "used_space", "free_space", "recordings_count", "snapshots_count", "recordings_size", "snapshots_size"}
	for _, field := range requiredFields {
		assert.Contains(t, result, field, "Missing required field '%s' per API documentation", field)
	}

	// Validate data types per API documentation
	assert.IsType(t, float64(0), result["total_space"], "total_space should be float64")
	assert.IsType(t, float64(0), result["used_space"], "used_space should be float64")
	assert.IsType(t, float64(0), result["free_space"], "free_space should be float64")
	assert.IsType(t, float64(0), result["recordings_count"], "recordings_count should be float64")
	assert.IsType(t, float64(0), result["snapshots_count"], "snapshots_count should be float64")
	assert.IsType(t, float64(0), result["recordings_size"], "recordings_size should be float64")
	assert.IsType(t, float64(0), result["snapshots_size"], "snapshots_size should be float64")

	// Validate performance target
	assert.Less(t, responseTime, 100*time.Millisecond, "get_storage_info response should be <100ms per API documentation")
}

// TestSetRetentionPolicyMethodImplementation tests set_retention_policy method implementation
// REQ-FILE-006: set_retention_policy method for policy management
func TestSetRetentionPolicyMethodImplementation(t *testing.T) {
	/*
		Unit Test for set_retention_policy method implementation

		API Documentation Reference: docs/api/json_rpc_methods.md
		Method: set_retention_policy
		Expected Response: {"jsonrpc": "2.0", "result": {"updated": true, "policy": {"recordings_days": 30, "snapshots_days": 7, "max_storage_gb": 100}}, "id": 1}
		Performance Target: <200ms response time
	*/

	// Setup real components with proper dependency injection
	configManager := config.NewConfigManager()
	logger := logging.NewLogger("test")

	// Create concrete implementations for camera interfaces
	deviceChecker := &camera.RealDeviceChecker{}
	commandExecutor := &camera.RealV4L2CommandExecutor{}
	infoParser := &camera.RealDeviceInfoParser{}

	cameraMonitor := camera.NewHybridCameraMonitor(configManager, logger, deviceChecker, commandExecutor, infoParser)
	jwtHandler, err := security.NewJWTHandler("test-secret-key-for-testing-only")
	require.NoError(t, err)

	server := ws.NewWebSocketServer(configManager, logger, cameraMonitor, jwtHandler)
	require.NotNil(t, server)

	// Create test client
	client := &ws.ClientConnection{
		ClientID:      "test-client",
		Authenticated: true,
		UserID:        "test-user",
		Role:          "admin",
		ConnectedAt:   time.Now(),
	}

	// Test set_retention_policy method with valid policy
	params := map[string]interface{}{
		"recordings_days": 30,
		"snapshots_days":  7,
		"max_storage_gb":  100,
	}

	// Measure response time
	startTime := time.Now()
	response, err := server.MethodSetRetentionPolicy(params, client)
	responseTime := time.Since(startTime)

	require.NoError(t, err)
	require.NotNil(t, response)

	// Validate response format per API documentation
	assert.Equal(t, "2.0", response.JSONRPC, "JSON-RPC version should be 2.0")
	assert.Nil(t, response.Error, "Response should not have error")
	assert.NotNil(t, response.Result, "Response should have result")

	// Validate result structure per API documentation
	result, ok := response.Result.(map[string]interface{})
	require.True(t, ok, "Result should be a map")

	// Check required fields from API documentation
	requiredFields := []string{"updated", "policy"}
	for _, field := range requiredFields {
		assert.Contains(t, result, field, "Missing required field '%s' per API documentation", field)
	}

	// Validate data types per API documentation
	assert.IsType(t, true, result["updated"], "updated should be bool")
	assert.IsType(t, map[string]interface{}{}, result["policy"], "policy should be map")

	// Validate policy structure
	policy, ok := result["policy"].(map[string]interface{})
	require.True(t, ok, "policy should be a map")

	policyFields := []string{"recordings_days", "snapshots_days", "max_storage_gb"}
	for _, field := range policyFields {
		assert.Contains(t, policy, field, "Missing policy field '%s' per API documentation", field)
		assert.IsType(t, float64(0), policy[field], "policy field should be float64")
	}

	// Validate performance target
	assert.Less(t, responseTime, 200*time.Millisecond, "set_retention_policy response should be <200ms per API documentation")
}

// TestCleanupOldFilesMethodImplementation tests cleanup_old_files method implementation
// REQ-FILE-007: cleanup_old_files method for file cleanup
func TestCleanupOldFilesMethodImplementation(t *testing.T) {
	/*
		Unit Test for cleanup_old_files method implementation

		API Documentation Reference: docs/api/json_rpc_methods.md
		Method: cleanup_old_files
		Expected Response: {"jsonrpc": "2.0", "result": {"cleaned": true, "files_deleted": 5, "space_freed": 524288000, "recordings_deleted": 3, "snapshots_deleted": 2}, "id": 1}
		Performance Target: <1000ms response time
	*/

	// Setup real components with proper dependency injection
	configManager := config.NewConfigManager()
	logger := logging.NewLogger("test")

	// Create concrete implementations for camera interfaces
	deviceChecker := &camera.RealDeviceChecker{}
	commandExecutor := &camera.RealV4L2CommandExecutor{}
	infoParser := &camera.RealDeviceInfoParser{}

	cameraMonitor := camera.NewHybridCameraMonitor(configManager, logger, deviceChecker, commandExecutor, infoParser)
	jwtHandler, err := security.NewJWTHandler("test-secret-key-for-testing-only")
	require.NoError(t, err)

	server := ws.NewWebSocketServer(configManager, logger, cameraMonitor, jwtHandler)
	require.NotNil(t, server)

	// Create test client
	client := &ws.ClientConnection{
		ClientID:      "test-client",
		Authenticated: true,
		UserID:        "test-user",
		Role:          "admin",
		ConnectedAt:   time.Now(),
	}

	// Test cleanup_old_files method
	params := map[string]interface{}{}

	// Measure response time
	startTime := time.Now()
	response, err := server.MethodCleanupOldFiles(params, client)
	responseTime := time.Since(startTime)

	require.NoError(t, err)
	require.NotNil(t, response)

	// Validate response format per API documentation
	assert.Equal(t, "2.0", response.JSONRPC, "JSON-RPC version should be 2.0")
	assert.Nil(t, response.Error, "Response should not have error")
	assert.NotNil(t, response.Result, "Response should have result")

	// Validate result structure per API documentation
	result, ok := response.Result.(map[string]interface{})
	require.True(t, ok, "Result should be a map")

	// Check required fields from API documentation
	requiredFields := []string{"cleaned", "files_deleted", "space_freed", "recordings_deleted", "snapshots_deleted"}
	for _, field := range requiredFields {
		assert.Contains(t, result, field, "Missing required field '%s' per API documentation", field)
	}

	// Validate data types per API documentation
	assert.IsType(t, true, result["cleaned"], "cleaned should be bool")
	assert.IsType(t, float64(0), result["files_deleted"], "files_deleted should be float64")
	assert.IsType(t, float64(0), result["space_freed"], "space_freed should be float64")
	assert.IsType(t, float64(0), result["recordings_deleted"], "recordings_deleted should be float64")
	assert.IsType(t, float64(0), result["snapshots_deleted"], "snapshots_deleted should be float64")

	// Validate performance target
	assert.Less(t, responseTime, 1000*time.Millisecond, "cleanup_old_files response should be <1000ms per API documentation")
}
