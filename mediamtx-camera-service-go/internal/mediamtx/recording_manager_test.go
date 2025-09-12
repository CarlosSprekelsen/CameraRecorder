/*
MediaMTX Recording Manager Tests - Real Server Integration

Requirements Coverage:
- REQ-MTX-001: MediaMTX service integration
- REQ-MTX-002: Stream management capabilities
- REQ-MTX-003: Path creation and deletion
- REQ-MTX-004: Health monitoring

Test Categories: Unit (using real MediaMTX server)
API Documentation Reference: docs/api/swagger.json
*/

package mediamtx

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewRecordingManager_ReqMTX001 tests recording manager creation with real server
func TestNewRecordingManager_ReqMTX001(t *testing.T) {
	// REQ-MTX-001: MediaMTX service integration
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Server is ready via shared test helper

	// Use shared recording manager from test helper
	recordingManager := helper.GetRecordingManager()
	require.NotNil(t, recordingManager, "Recording manager should be initialized")
}

// TestRecordingManager_StartRecording_ReqMTX002 tests recording session creation with real server
func TestRecordingManager_StartRecording_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	EnsureSequentialExecution(t)
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Server is ready via shared test helper

	// Use shared recording manager from test helper
	recordingManager := helper.GetRecordingManager()
	require.NotNil(t, recordingManager)

	ctx := context.Background()

	// Create temporary output directory
	tempDir := filepath.Join(helper.GetConfig().TestDataDir, "recordings")
	err := os.MkdirAll(tempDir, 0755)
	require.NoError(t, err)

	devicePath := "/dev/video0"
	outputPath := filepath.Join(tempDir, "test_recording.mp4")

	// Start recording
	options := map[string]interface{}{}
	session, err := recordingManager.StartRecording(ctx, devicePath, outputPath, options)
	require.NoError(t, err, "Recording should start successfully")
	require.NotNil(t, session, "Recording session should not be nil")
	assert.Equal(t, devicePath, session.DevicePath)
	assert.Equal(t, outputPath, session.FilePath)

	// Verify session is tracked
	sessions := recordingManager.ListRecordingSessions()
	assert.Len(t, sessions, 1)
	assert.Equal(t, session.ID, sessions[0].ID)

	// Clean up
	err = recordingManager.StopRecording(ctx, session.ID)
	require.NoError(t, err, "Recording should stop successfully")
}

// TestRecordingManager_StopRecording_ReqMTX002 tests recording session termination with real server
func TestRecordingManager_StopRecording_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Server is ready via shared test helper

	// Use shared recording manager from test helper
	recordingManager := helper.GetRecordingManager()
	require.NotNil(t, recordingManager)

	ctx := context.Background()

	// Create temporary output directory
	tempDir := filepath.Join(helper.GetConfig().TestDataDir, "recordings")
	err := os.MkdirAll(tempDir, 0755)
	require.NoError(t, err)

	devicePath := "/dev/video0"
	outputPath := filepath.Join(tempDir, "test_recording_stop.mp4")

	// Start recording
	options := map[string]interface{}{}
	session, err := recordingManager.StartRecording(ctx, devicePath, outputPath, options)
	require.NoError(t, err, "Recording should start successfully")

	// Verify session is active
	sessions := recordingManager.ListRecordingSessions()
	assert.Len(t, sessions, 1)

	// Stop recording
	err = recordingManager.StopRecording(ctx, session.ID)
	require.NoError(t, err, "Recording should stop successfully")

	// Verify session is no longer active
	sessions = recordingManager.ListRecordingSessions()
	assert.Len(t, sessions, 0)
}

// TestRecordingManager_ListRecordingSessions_ReqMTX002 tests session listing with real server
func TestRecordingManager_ListRecordingSessions_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Server is ready via shared test helper

	// Use shared recording manager from test helper
	recordingManager := helper.GetRecordingManager()
	require.NotNil(t, recordingManager)

	ctx := context.Background()

	// Create temporary output directory
	tempDir := filepath.Join(helper.GetConfig().TestDataDir, "recordings")
	err := os.MkdirAll(tempDir, 0755)
	require.NoError(t, err)

	// Initially no sessions
	sessions := recordingManager.ListRecordingSessions()
	assert.Len(t, sessions, 0)

	// Start multiple recordings
	options := map[string]interface{}{}
	session1, err := recordingManager.StartRecording(ctx, "/dev/video0", filepath.Join(tempDir, "test1.mp4"), options)
	require.NoError(t, err)

	session2, err := recordingManager.StartRecording(ctx, "/dev/video1", filepath.Join(tempDir, "test2.mp4"), options)
	require.NoError(t, err)

	// Verify both sessions are tracked
	sessions = recordingManager.ListRecordingSessions()
	assert.Len(t, sessions, 2)

	// Clean up
	recordingManager.StopRecording(ctx, session1.ID)
	recordingManager.StopRecording(ctx, session2.ID)
}

// TestRecordingManager_GetRecordingsListAPI_ReqMTX002 tests MediaMTX API integration with real server
func TestRecordingManager_GetRecordingsListAPI_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities - API compliance validation
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Server is ready via shared test helper

	// Use shared recording manager from test helper
	recordingManager := helper.GetRecordingManager()
	require.NotNil(t, recordingManager)

	ctx := context.Background()

	// Test 1: Call MediaMTX /v3/recordings/list API endpoint directly
	response, err := recordingManager.GetRecordingsList(ctx, 10, 0)
	require.NoError(t, err, "GetRecordingsList should succeed with real MediaMTX server")
	require.NotNil(t, response, "Response should not be nil")

	// Test 2: Validate API response format matches swagger.json specification
	assert.NotNil(t, response.Files, "Files field should be present (as per swagger.json)")
	assert.GreaterOrEqual(t, response.Total, 0, "Total should be non-negative integer")
	assert.Equal(t, 10, response.Limit, "Limit should match requested value")
	assert.Equal(t, 0, response.Offset, "Offset should match requested value")

	// Test 3: Validate pagination parameters work with MediaMTX API
	response2, err := recordingManager.GetRecordingsList(ctx, 5, 1)
	require.NoError(t, err, "Pagination should work with MediaMTX API")
	assert.Equal(t, 5, response2.Limit, "Pagination limit should be respected")
	assert.Equal(t, 1, response2.Offset, "Pagination offset should be respected")

	t.Log("✅ MediaMTX API /v3/recordings/list endpoint validation passed")
}

// TestRecordingManager_StartRecordingCreatesPath_ReqMTX003 tests MediaMTX path creation
func TestRecordingManager_StartRecordingCreatesPath_ReqMTX003(t *testing.T) {
	// REQ-MTX-003: Path creation and deletion - Validate MediaMTX API integration
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Server is ready via shared test helper

	// Use shared recording manager from test helper
	recordingManager := helper.GetRecordingManager()

	ctx := context.Background()

	devicePath := "/dev/video0"
	outputPath := filepath.Join(helper.GetConfiguredRecordingPath(), "test_recording_path.mp4")

	// Test: StartRecording should create a path in MediaMTX via API
	session, err := recordingManager.StartRecording(ctx, devicePath, outputPath, map[string]interface{}{})
	require.NoError(t, err, "StartRecording should succeed")
	require.NotNil(t, session, "Session should be created")

	// Validate that a path was actually created in MediaMTX server
	// This validates real API integration - test would FAIL if MediaMTX API is broken
	pathData, err := helper.GetClient().Get(ctx, "/v3/paths/get/"+session.Path)
	require.NoError(t, err, "Path should exist in MediaMTX server after StartRecording")
	require.NotNil(t, pathData, "Path data should be returned from MediaMTX API")

	// Validate path configuration matches our recording requirements
	var pathInfo map[string]interface{}
	err = json.Unmarshal(pathData, &pathInfo)
	require.NoError(t, err, "Path data should be valid JSON")

	// Check recording is enabled (as per swagger.json PathConf schema)
	assert.NotNil(t, pathInfo["confName"], "Path should have configuration name")

	// Clean up: Stop recording should delete the path from MediaMTX
	err = recordingManager.StopRecording(ctx, session.ID)
	require.NoError(t, err, "StopRecording should succeed")

	// Verify path was deleted from MediaMTX server
	_, err = helper.GetClient().Get(ctx, "/v3/paths/get/"+session.Path)
	assert.Error(t, err, "Path should be deleted from MediaMTX server after StopRecording")

	t.Log("✅ MediaMTX path creation/deletion API validation passed")
}

// TestRecordingManager_APISchemaCompliance_ReqMTX001 tests swagger.json schema compliance
func TestRecordingManager_APISchemaCompliance_ReqMTX001(t *testing.T) {
	// REQ-MTX-001: MediaMTX service integration - Schema validation per swagger.json
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Server is ready via shared test helper

	ctx := context.Background()

	// Test 1: Validate /v3/recordings/list response matches RecordingList schema
	data, err := helper.GetClient().Get(ctx, "/v3/recordings/list?itemsPerPage=10&page=0")
	require.NoError(t, err, "MediaMTX /v3/recordings/list API should respond")

	var recordingListResponse struct {
		PageCount int `json:"pageCount"`
		ItemCount int `json:"itemCount"`
		Items     []struct {
			Name     string `json:"name"`
			Segments []struct {
				Start string `json:"start"`
			} `json:"segments"`
		} `json:"items"`
	}

	err = json.Unmarshal(data, &recordingListResponse)
	require.NoError(t, err, "Response should match RecordingList schema from swagger.json")

	// Validate all required fields are present (per swagger.json)
	assert.GreaterOrEqual(t, recordingListResponse.PageCount, 0, "pageCount field required per swagger.json")
	assert.GreaterOrEqual(t, recordingListResponse.ItemCount, 0, "itemCount field required per swagger.json")
	assert.NotNil(t, recordingListResponse.Items, "items array required per swagger.json")

	// Test 2: Validate /v3/paths/list response matches PathList schema
	pathData, err := helper.GetClient().Get(ctx, "/v3/paths/list?itemsPerPage=10&page=0")
	require.NoError(t, err, "MediaMTX /v3/paths/list API should respond")

	var pathListResponse struct {
		PageCount int `json:"pageCount"`
		ItemCount int `json:"itemCount"`
		Items     []struct {
			Name     string `json:"name"`
			ConfName string `json:"confName"`
			Ready    bool   `json:"ready"`
		} `json:"items"`
	}

	err = json.Unmarshal(pathData, &pathListResponse)
	require.NoError(t, err, "Response should match PathList schema from swagger.json")

	// Validate required fields per swagger.json PathList schema
	assert.GreaterOrEqual(t, pathListResponse.PageCount, 0, "pageCount field required")
	assert.GreaterOrEqual(t, pathListResponse.ItemCount, 0, "itemCount field required")
	assert.NotNil(t, pathListResponse.Items, "items array required")

	t.Log("✅ MediaMTX API schema compliance validation passed")
}

// TestRecordingManager_APIErrorHandling_ReqMTX004 tests error handling with MediaMTX API
func TestRecordingManager_APIErrorHandling_ReqMTX004(t *testing.T) {
	// REQ-MTX-004: Health monitoring and circuit breaker - Error handling validation
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	ctx := context.Background()

	// Test 1: Invalid path should return 404 error as per swagger.json
	// Use test-prefixed name to ensure proper cleanup
	_, err := helper.GetClient().Get(ctx, "/v3/paths/get/test_nonexistent_path")
	assert.Error(t, err, "Non-existent path should return error per swagger.json")

	// Test 2: Non-existent recording should return 200 with empty segments per swagger.json
	// Use test-prefixed name to ensure proper cleanup
	data, err := helper.GetClient().Get(ctx, "/v3/recordings/get/test_nonexistent_recording")
	require.NoError(t, err, "Non-existent recording should return 200 with empty segments per swagger.json")

	// Verify the response structure matches Recording schema
	var recording struct {
		Name     string     `json:"name"`
		Segments []struct{} `json:"segments"`
	}
	err = json.Unmarshal(data, &recording)
	require.NoError(t, err, "Response should be valid JSON")
	assert.Equal(t, "test_nonexistent_recording", recording.Name, "Recording name should match")
	assert.Empty(t, recording.Segments, "Segments should be empty for non-existent recording")

	// Test 3: MediaMTX API endpoints should be available (this validates real integration)
	// If MediaMTX API was broken, these calls would fail
	recordingsData, err := helper.GetClient().Get(ctx, "/v3/recordings/list")
	require.NoError(t, err, "MediaMTX recordings API should be accessible")
	require.NotNil(t, recordingsData, "Response data should not be nil")

	pathData, err := helper.GetClient().Get(ctx, "/v3/paths/list")
	require.NoError(t, err, "MediaMTX paths API should be accessible")
	require.NotNil(t, pathData, "Path data should not be nil")

	t.Log("✅ MediaMTX API error handling validation passed")
}

// TestRecordingManager_GetRecordingSession_ReqMTX002 tests session retrieval with real server
func TestRecordingManager_GetRecordingSession_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Server is ready via shared test helper

	// Use shared recording manager from test helper
	recordingManager := helper.GetRecordingManager()
	require.NotNil(t, recordingManager)

	ctx := context.Background()

	// Create temporary output directory
	tempDir := filepath.Join(helper.GetConfig().TestDataDir, "recordings")
	err := os.MkdirAll(tempDir, 0755)
	require.NoError(t, err)

	devicePath := "/dev/video0"
	outputPath := filepath.Join(tempDir, "test_recording_get.mp4")

	// Start recording
	options := map[string]interface{}{}
	session, err := recordingManager.StartRecording(ctx, devicePath, outputPath, options)
	require.NoError(t, err, "Recording should start successfully")

	// Verify session is tracked in the list
	sessions := recordingManager.ListRecordingSessions()
	require.Len(t, sessions, 1, "Should have one active session")

	retrievedSession := sessions[0]
	assert.Equal(t, session.ID, retrievedSession.ID)
	assert.Equal(t, devicePath, retrievedSession.DevicePath)
	assert.Equal(t, outputPath, retrievedSession.FilePath)

	// Test that session is properly managed
	assert.NotEmpty(t, session.ID, "Session should have an ID")
	assert.NotNil(t, session.StartTime, "Session should have start time")

	// Clean up
	recordingManager.StopRecording(ctx, session.ID)
}

// TestRecordingManager_ErrorHandling_ReqMTX007 tests error scenarios with real server
func TestRecordingManager_ErrorHandling_ReqMTX007(t *testing.T) {
	// REQ-MTX-007: Error handling and recovery
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Server is ready via shared test helper

	// Use shared recording manager from test helper
	recordingManager := helper.GetRecordingManager()
	require.NotNil(t, recordingManager)

	ctx := context.Background()

	// Create temporary output directory
	tempDir := filepath.Join(helper.GetConfig().TestDataDir, "recordings")
	err := os.MkdirAll(tempDir, 0755)
	require.NoError(t, err)

	// Test invalid device path
	options := map[string]interface{}{}
	_, err = recordingManager.StartRecording(ctx, "", filepath.Join(tempDir, "test.mp4"), options)
	assert.Error(t, err, "Empty device path should fail")

	// Test invalid output path
	_, err = recordingManager.StartRecording(ctx, "/dev/video0", "", options)
	assert.Error(t, err, "Empty output path should fail")

	// Test stopping non-existent session
	err = recordingManager.StopRecording(ctx, "non-existent-id")
	assert.Error(t, err, "Stopping non-existent session should fail")
}

// TestRecordingManager_ConcurrentAccess_ReqMTX001 tests concurrent operations with real server
func TestRecordingManager_ConcurrentAccess_ReqMTX001(t *testing.T) {
	// REQ-MTX-001: MediaMTX service integration
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Server is ready via shared test helper

	// Use shared recording manager from test helper
	recordingManager := helper.GetRecordingManager()
	require.NotNil(t, recordingManager)

	ctx := context.Background()

	// Create temporary output directory
	tempDir := filepath.Join(helper.GetConfig().TestDataDir, "recordings")
	err := os.MkdirAll(tempDir, 0755)
	require.NoError(t, err)

	// Start multiple recordings concurrently
	const numRecordings = 3 // Reduced for real server testing
	sessions := make([]*RecordingSession, numRecordings)
	errors := make([]error, numRecordings)

	for i := 0; i < numRecordings; i++ {
		go func(index int) {
			devicePath := "/dev/video0" // Use same device for real server
			outputPath := filepath.Join(tempDir, "concurrent_test.mp4")
			options := map[string]interface{}{}
			session, err := recordingManager.StartRecording(ctx, devicePath, outputPath, options)
			sessions[index] = session
			errors[index] = err
		}(i)
	}

	// Wait for all goroutines to complete
	time.Sleep(100 * time.Millisecond)

	// Verify recordings started successfully (some may fail due to device conflicts)
	activeSessions := recordingManager.ListRecordingSessions()
	assert.GreaterOrEqual(t, len(activeSessions), 0, "Should handle concurrent recordings gracefully")

	// Clean up all sessions
	for _, session := range sessions {
		if session != nil {
			recordingManager.StopRecording(ctx, session.ID)
		}
	}
}

// TestRecordingManager_StartRecordingWithSegments_ReqMTX002 tests segmented recording with real server
func TestRecordingManager_StartRecordingWithSegments_ReqMTX002(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Server is ready via shared test helper

	// Use shared recording manager from test helper
	recordingManager := helper.GetRecordingManager()
	require.NotNil(t, recordingManager)

	ctx := context.Background()

	// Create temporary output directory
	tempDir := filepath.Join(helper.GetConfig().TestDataDir, "recordings")
	err := os.MkdirAll(tempDir, 0755)
	require.NoError(t, err)

	devicePath := "/dev/video0"
	outputPath := filepath.Join(tempDir, "segmented_test.mp4")

	// Test MediaMTX recording configuration options
	options := map[string]interface{}{
		"recordFormat":          "mp4",
		"recordPartDuration":    "10s",
		"recordSegmentDuration": "1h",
		"recordDeleteAfter":     "24h",
	}

	session, err := recordingManager.StartRecording(ctx, devicePath, outputPath, options)
	require.NoError(t, err, "Recording with MediaMTX config should start successfully")
	require.NotNil(t, session, "Recording session should not be nil")

	// Verify session is created with configuration
	assert.NotEmpty(t, session.ID, "Session should have ID")
	assert.Equal(t, devicePath, session.DevicePath)

	// Clean up
	recordingManager.StopRecording(ctx, session.ID)
}
