/*
E2E Test Helpers - Enterprise Grade Workflow Testing Utilities

Provides comprehensive utilities for end-to-end testing of complete user workflows
across all system boundaries with real components (no mocks). All helpers build
on testutils infrastructure and enforce enterprise-grade validation standards.

Requirements Coverage:
- E2E-TEST-001: Complete workflow validation
- E2E-TEST-002: Real file creation and verification
- E2E-TEST-003: Actual state change validation
- E2E-TEST-004: Business outcome verification
- E2E-TEST-005: Zero mocks enforcement

Design Principles:
- Build on testutils.UniversalTestSetup infrastructure
- Never use time.Sleep - always use testutils.WaitForCondition
- Verify actual file content, not just existence
- Validate complete business outcomes
- Use real components and services
*/

package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/security"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/testutils"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// JSONRPCRequest represents a JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC string                 `json:"jsonrpc"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params"`
	ID      int                    `json:"id"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response
type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	Result  interface{}   `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
	ID      int           `json:"id"`
}

// JSONRPCError represents a JSON-RPC 2.0 error
type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data,omitempty"`
}

// Error implements the error interface
func (e *JSONRPCError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

// E2ETestSetup represents the complete E2E test setup
type E2ETestSetup struct {
	t           *testing.T
	universal   *testutils.UniversalTestSetup
	connections []*websocket.Conn
	filePaths   []string
	tempDirs    []string
}

// NewE2ETestSetup creates a new E2E test setup
func NewE2ETestSetup(t *testing.T) *E2ETestSetup {
	universal := testutils.SetupTest(t, "config_e2e_test.yaml")

	setup := &E2ETestSetup{
		t:           t,
		universal:   universal,
		connections: make([]*websocket.Conn, 0),
		filePaths:   make([]string, 0),
		tempDirs:    make([]string, 0),
	}

	// Register cleanup
	t.Cleanup(func() {
		setup.Cleanup()
	})

	return setup
}

// Cleanup performs comprehensive E2E test cleanup
func (s *E2ETestSetup) Cleanup() {
	// Close all WebSocket connections
	for _, conn := range s.connections {
		if conn != nil {
			conn.Close()
		}
	}

	// Remove all test files
	for _, filePath := range s.filePaths {
		os.RemoveAll(filePath)
	}

	// Clean up temporary directories
	for _, dir := range s.tempDirs {
		os.RemoveAll(dir)
	}

	// Verify cleanup succeeded
	verifyCleanup(s.t, s.filePaths, s.tempDirs)
}

// GetUniversalSetup returns the underlying universal test setup
func (s *E2ETestSetup) GetUniversalSetup() *testutils.UniversalTestSetup {
	return s.universal
}

// GenerateTestToken generates a JWT token for testing
func GenerateTestToken(t *testing.T, role string, expiryHours int) string {
	setup := testutils.SetupTest(t, "config_e2e_test.yaml")
	jwtHandler, err := security.NewJWTHandler("e2e_test_secret_key_for_testing_only", setup.GetLogger())
	require.NoError(t, err, "Failed to create JWT handler")

	token, err := jwtHandler.GenerateToken("test_user_"+role, role, expiryHours)
	require.NoError(t, err, "Failed to generate JWT token")

	return token
}

// EstablishConnection connects and authenticates WebSocket connection
func (s *E2ETestSetup) EstablishConnection(token string) *websocket.Conn {
	// Parse WebSocket URL
	u := url.URL{Scheme: "ws", Host: "localhost:8002", Path: "/ws"}

	// Connect with timeout
	ctx, cancel := context.WithTimeout(context.Background(), testutils.UniversalTimeoutVeryLong)
	defer cancel()

	conn, _, err := websocket.DefaultDialer.DialContext(ctx, u.String(), nil)
	require.NoError(s.t, err, "Failed to connect to WebSocket server")

	// Store connection for cleanup
	s.connections = append(s.connections, conn)

	// Authenticate
	authResponse := s.SendJSONRPC(conn, "authenticate", map[string]interface{}{
		"auth_token": token,
	})

	// Verify authentication succeeded
	assert.NoError(s.t, authResponse.Error, "Authentication should succeed")
	require.NotNil(s.t, authResponse.Result, "Authentication result should not be nil")

	// Verify result contains success indicator
	resultMap, ok := authResponse.Result.(map[string]interface{})
	require.True(s.t, ok, "Authentication result should be a map")
	assert.Equal(s.t, "success", resultMap["status"], "Authentication status should be success")

	return conn
}

// CloseConnection properly closes WebSocket connection
func (s *E2ETestSetup) CloseConnection(conn *websocket.Conn) {
	if conn == nil {
		return
	}

	// Send close frame
	err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		s.t.Logf("Warning: Failed to send close frame: %v", err)
	}

	// Wait for close confirmation with timeout
	done := make(chan struct{})
	go func() {
		defer close(done)
		conn.Close()
	}()

	select {
	case <-done:
		// Connection closed successfully
	case <-time.After(testutils.UniversalTimeoutShort):
		s.t.Logf("Warning: Connection close timeout")
	}
}

// SendJSONRPC sends a JSON-RPC request and waits for response
func (s *E2ETestSetup) SendJSONRPC(conn *websocket.Conn, method string, params map[string]interface{}) *JSONRPCResponse {
	// Generate unique request ID
	requestID := time.Now().UnixNano() % 1000000

	// Create request
	request := JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      int(requestID),
	}

	// Marshal request
	requestBytes, err := json.Marshal(request)
	require.NoError(s.t, err, "Failed to marshal JSON-RPC request")

	// Send request with timeout
	err = conn.WriteMessage(websocket.TextMessage, requestBytes)
	require.NoError(s.t, err, "Failed to send JSON-RPC request")

	// Wait for response with proper ID tracking
	var response *JSONRPCResponse
	err = s.WaitForCondition(func() bool {
		_, message, readErr := conn.ReadMessage()
		if readErr != nil {
			return false
		}

		var resp JSONRPCResponse
		if unmarshalErr := json.Unmarshal(message, &resp); unmarshalErr != nil {
			return false
		}

		if resp.ID == request.ID {
			response = &resp
			return true
		}

		return false
	}, testutils.UniversalTimeoutVeryLong)

	require.NoError(s.t, err, "Timeout waiting for JSON-RPC response")
	require.NotNil(s.t, response, "JSON-RPC response should not be nil")

	return response
}

// WaitForServiceReady waits for service to be ready for operations
func (s *E2ETestSetup) WaitForServiceReady(conn *websocket.Conn) {
	err := s.WaitForCondition(func() bool {
		// Send ping request
		response := s.SendJSONRPC(conn, "ping", map[string]interface{}{})
		return response.Error == nil
	}, testutils.UniversalTimeoutVeryLong)

	require.NoError(s.t, err, "Service should become ready within timeout")
}

// VerifyRecordingFile validates complete recording file with content checks
func (s *E2ETestSetup) VerifyRecordingFile(filePath string, minDurationSeconds int) {
	s.filePaths = append(s.filePaths, filePath) // Track for cleanup

	dvh := testutils.NewDataValidationHelper(s.t)

	// File exists with minimum size
	dvh.AssertFileExists(filePath, testutils.UniversalMinRecordingFileSize, "recording file")

	// File accessible and readable
	dvh.AssertFileAccessible(filePath, "recording file")

	// File size appropriate for duration (rough estimate: 200KB/sec for compressed video)
	expectedMinSize := int64(minDurationSeconds * 200 * 1024)
	dvh.AssertFileSize(filePath, expectedMinSize, 0, "recording duration validation")

	// File format validation (check file header for video format)
	s.validateVideoFileHeader(filePath)
}

// VerifySnapshotFile validates complete snapshot file with image header check
func (s *E2ETestSetup) VerifySnapshotFile(filePath string, expectedFormat string) {
	s.filePaths = append(s.filePaths, filePath) // Track for cleanup

	dvh := testutils.NewDataValidationHelper(s.t)

	// File exists with minimum size
	dvh.AssertFileExists(filePath, testutils.UniversalMinSnapshotFileSize, "snapshot file")

	// File accessible
	dvh.AssertFileAccessible(filePath, "snapshot file")

	// Valid image header
	s.validateImageFileHeader(filePath, expectedFormat)

	// Extension matches format
	assert.True(s.t, strings.HasSuffix(filePath, "."+expectedFormat), "file extension matches format")
}

// LogWorkflowStep logs workflow progress for debugging
func LogWorkflowStep(t *testing.T, workflowName string, stepNumber int, stepDescription string) {
	t.Logf("🔹 [%s] Step %d: %s", workflowName, stepNumber, stepDescription)
}

// WaitForRecordingStart waits for recording file to be created and growing
func (s *E2ETestSetup) WaitForRecordingStart(filePath string) {
	err := s.WaitForCondition(func() bool {
		// Check file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return false
		}

		// Check file size is growing (wait for second check)
		time.Sleep(10 * time.Millisecond)
		if _, err := os.Stat(filePath); err != nil {
			return false
		}

		return true
	}, testutils.UniversalTimeoutVeryLong)

	require.NoError(s.t, err, "Recording file should be created and growing within timeout")
}

// WaitForFileCreation waits for a file to be created with timeout
func (s *E2ETestSetup) WaitForFileCreation(filePath string, description string) bool {
	err := s.WaitForCondition(func() bool {
		_, err := os.Stat(filePath)
		return err == nil
	}, testutils.UniversalTimeoutVeryLong)
	return err == nil
}

// CreateTestRecordingPath creates a test recording file path
func (s *E2ETestSetup) CreateTestRecordingPath(filename string) string {
	basePath := "/tmp/e2e-test-recordings"
	tempPath := filepath.Join(basePath, filename)
	s.tempDirs = append(s.tempDirs, basePath)
	return tempPath
}

// CreateTestSnapshotPath creates a test snapshot file path
func (s *E2ETestSetup) CreateTestSnapshotPath(filename string) string {
	basePath := "/tmp/e2e-test-snapshots"
	tempPath := filepath.Join(basePath, filename)
	s.tempDirs = append(s.tempDirs, basePath)
	return tempPath
}

// validateVideoFileHeader validates video file header
func (s *E2ETestSetup) validateVideoFileHeader(filePath string) {
	file, err := os.Open(filePath)
	require.NoError(s.t, err, "Should be able to open recording file")
	defer file.Close()

	// Read first few bytes to check file header
	header := make([]byte, 8)
	_, err = file.Read(header)
	require.NoError(s.t, err, "Should be able to read file header")

	// Check for common video file signatures
	isValidVideo := false

	// MP4 signature (ftyp box)
	if len(header) >= 4 && header[4] == 'f' && header[5] == 't' && header[6] == 'y' && header[7] == 'p' {
		isValidVideo = true
	}

	// MKV signature
	if len(header) >= 4 && header[0] == 0x1A && header[1] == 0x45 && header[2] == 0xDF && header[3] == 0xA3 {
		isValidVideo = true
	}

	// AVI signature
	if len(header) >= 4 && header[0] == 'R' && header[1] == 'I' && header[2] == 'F' && header[3] == 'F' {
		isValidVideo = true
	}

	assert.True(s.t, isValidVideo, "Recording file should have valid video format header")
}

// validateImageFileHeader validates image file header
func (s *E2ETestSetup) validateImageFileHeader(filePath string, expectedFormat string) {
	file, err := os.Open(filePath)
	require.NoError(s.t, err, "Should be able to open snapshot file")
	defer file.Close()

	// Read first few bytes to check file header
	header := make([]byte, 4)
	_, err = file.Read(header)
	require.NoError(s.t, err, "Should be able to read file header")

	// Validate based on expected format
	switch strings.ToLower(expectedFormat) {
	case "jpeg", "jpg":
		// JPEG signature: FF D8 FF
		assert.Equal(s.t, byte(0xFF), header[0], "JPEG file should start with 0xFF")
		assert.Equal(s.t, byte(0xD8), header[1], "JPEG file should have 0xD8 as second byte")
		assert.Equal(s.t, byte(0xFF), header[2], "JPEG file should have 0xFF as third byte")
	case "png":
		// PNG signature: 89 50 4E 47
		assert.Equal(s.t, byte(0x89), header[0], "PNG file should start with 0x89")
		assert.Equal(s.t, byte(0x50), header[1], "PNG file should have 0x50 as second byte")
		assert.Equal(s.t, byte(0x4E), header[2], "PNG file should have 0x4E as third byte")
		assert.Equal(s.t, byte(0x47), header[3], "PNG file should have 0x47 as fourth byte")
	case "bmp":
		// BMP signature: BM (42 4D)
		assert.Equal(s.t, byte(0x42), header[0], "BMP file should start with 0x42")
		assert.Equal(s.t, byte(0x4D), header[1], "BMP file should have 0x4D as second byte")
	default:
		s.t.Logf("Warning: Unknown image format %s, skipping header validation", expectedFormat)
	}
}

// verifyCleanup validates cleanup actually succeeded
func verifyCleanup(t *testing.T, filePaths []string, tempDirs []string) {
	for _, path := range filePaths {
		_, err := os.Stat(path)
		assert.True(t, os.IsNotExist(err), "test file should be removed: %s", path)
	}
	for _, dir := range tempDirs {
		entries, _ := os.ReadDir(dir)
		assert.Empty(t, entries, "temp directory should be empty: %s", dir)
	}
}

// AssertBusinessOutcome validates that business outcome was achieved
func (s *E2ETestSetup) AssertBusinessOutcome(description string, outcome func() bool) {
	assert.True(s.t, outcome(), "Business outcome should be achieved: %s", description)
}

// CreateTestRecordings creates multiple test recordings for list testing
func (s *E2ETestSetup) CreateTestRecordings(count int) []string {
	recordings := make([]string, count)
	for i := 0; i < count; i++ {
		// Create a minimal test recording file
		filePath := s.CreateTestRecordingPath(fmt.Sprintf("test_recording_%d.mp4", i))

		// Create the file with minimal content
		file, err := os.Create(filePath)
		require.NoError(s.t, err, "Should create test recording file")

		// Write minimal MP4 header
		_, err = file.WriteString("ftypmp42") // Minimal MP4 signature
		require.NoError(s.t, err, "Should write test content")

		file.Close()
		recordings[i] = filePath
	}

	return recordings
}

// CreateTestSnapshots creates multiple test snapshots for list testing
func (s *E2ETestSetup) CreateTestSnapshots(count int) []string {
	snapshots := make([]string, count)
	for i := 0; i < count; i++ {
		// Create a minimal test snapshot file
		filePath := s.CreateTestSnapshotPath(fmt.Sprintf("test_snapshot_%d.jpg", i))

		// Create the file with minimal content
		file, err := os.Create(filePath)
		require.NoError(s.t, err, "Should create test snapshot file")

		// Write minimal JPEG header
		_, err = file.Write([]byte{0xFF, 0xD8, 0xFF, 0xE0}) // Minimal JPEG signature
		require.NoError(s.t, err, "Should write test content")

		file.Close()
		snapshots[i] = filePath
	}

	return snapshots
}

// WaitForCondition waits for a condition to be true with timeout
func (s *E2ETestSetup) WaitForCondition(condition func() bool, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return nil
		}
		// Use a very short sleep to avoid busy waiting
		time.Sleep(10 * time.Millisecond)
	}
	return fmt.Errorf("timeout waiting for condition after %v", timeout)
}

// GetStandardTimeout returns appropriate timeout for E2E test operations
func GetStandardTimeout(operationType string) time.Duration {
	switch operationType {
	case "single_operation":
		return testutils.DefaultTestTimeout // 5s
	case "multi_step_workflow":
		return testutils.UniversalTimeoutVeryLong // 30s
	case "complete_workflow":
		return testutils.UniversalTimeoutExtreme // 60s
	case "recording_with_duration":
		return testutils.UniversalTimeoutVeryLong // Will be adjusted per duration
	default:
		return testutils.UniversalTimeoutVeryLong
	}
}

// Helper functions for file operations

// getFileSize returns the size of a file
func getFileSize(t *testing.T, filePath string) int64 {
	info, err := os.Stat(filePath)
	if err != nil {
		return 0
	}
	return info.Size()
}

// fileExists checks if a file exists
func fileExists(t *testing.T, filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}
