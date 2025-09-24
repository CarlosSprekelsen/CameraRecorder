/*
Data Validation Helpers - Functionality Validation Testing Utilities

Provides utilities for testing actual data creation and validation instead of accommodation testing.
These helpers ensure tests verify that operations actually create files, modify state, and produce
expected results rather than just checking for no errors.

Requirements Coverage:
- REQ-TEST-001: Data-driven test validation
- REQ-TEST-002: File existence and content validation
- REQ-TEST-003: State transition verification
- REQ-TEST-004: Lifecycle validation

Design Principles:
- Verify actual data creation (not just no errors)
- Test complete lifecycles (start → active → stop → cleanup)
- Validate state transitions throughout operations
- Test error conditions properly (not accommodate them)
- Use progressive readiness patterns for event-driven readiness
*/

package testutils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// DataValidationHelper provides utilities for validating actual data creation
type DataValidationHelper struct {
	t       *testing.T
	tempDir string
}

// NewDataValidationHelper creates a new data validation helper
func NewDataValidationHelper(t *testing.T) *DataValidationHelper {
	tempDir := t.TempDir()
	return &DataValidationHelper{
		t:       t,
		tempDir: tempDir,
	}
}

// GetTempDir returns the temporary directory for test files
func (dvh *DataValidationHelper) GetTempDir() string {
	return dvh.tempDir
}

// AssertFileNotExists validates that a file does not exist
func (dvh *DataValidationHelper) AssertFileNotExists(filePath string, description string) {
	_, err := os.Stat(filePath)
	assert.True(dvh.t, os.IsNotExist(err), "%s: File should not exist initially: %s", description, filePath)
}

// AssertFileExists validates that a file exists and has content
func (dvh *DataValidationHelper) AssertFileExists(filePath string, minSize int64, description string) {
	info, err := os.Stat(filePath)
	require.NoError(dvh.t, err, "%s: File should exist: %s", description, filePath)

	// Validate file has meaningful content
	if minSize > 0 {
		assert.Greater(dvh.t, info.Size(), minSize,
			"%s: File should have meaningful size (min %d bytes): %s", description, minSize, filePath)
	}
}

// AssertFileSize validates that a file has a specific size range
func (dvh *DataValidationHelper) AssertFileSize(filePath string, minSize, maxSize int64, description string) {
	info, err := os.Stat(filePath)
	require.NoError(dvh.t, err, "%s: File should exist: %s", description, filePath)

	if minSize > 0 {
		assert.GreaterOrEqual(dvh.t, info.Size(), minSize,
			"%s: File should be at least %d bytes: %s", description, minSize, filePath)
	}
	if maxSize > 0 {
		assert.LessOrEqual(dvh.t, info.Size(), maxSize,
			"%s: File should be at most %d bytes: %s", description, maxSize, filePath)
	}
}

// AssertFileCreated validates that a file was created during an operation
func (dvh *DataValidationHelper) AssertFileCreated(operation func() error, filePath string, minSize int64, description string) error {
	// Step 1: Verify initial state - file should not exist
	dvh.AssertFileNotExists(filePath, description+" initial state")

	// Step 2: Execute operation
	err := operation()

	// Step 3: Verify data was actually created (if operation succeeded)
	if err == nil {
		dvh.AssertFileExists(filePath, minSize, description+" after operation")
	}

	return err
}

// AssertFileDeleted validates that a file was deleted during an operation
func (dvh *DataValidationHelper) AssertFileDeleted(operation func() error, filePath string, description string) error {
	// Step 1: Verify initial state - file should exist
	dvh.AssertFileExists(filePath, 0, description+" initial state")

	// Step 2: Execute operation
	err := operation()

	// Step 3: Verify file was actually deleted (if operation succeeded)
	if err == nil {
		dvh.AssertFileNotExists(filePath, description+" after operation")
	}

	return err
}

// AssertProgressiveFileCreation validates file creation with progressive readiness
func (dvh *DataValidationHelper) AssertProgressiveFileCreation(
	operation func() error,
	filePath string,
	minSize int64,
	component ReadinessSubscriber,
	operationName string,
	description string,
) error {
	// Step 1: Verify initial state
	dvh.AssertFileNotExists(filePath, description+" initial state")

	// Step 2: Execute operation with progressive readiness
	result := TestProgressiveReadiness(dvh.t, func() (struct{}, error) {
		err := operation()
		return struct{}{}, err
	}, component, operationName)

	// Step 3: Verify data was actually created (if operation succeeded)
	if result.Error == nil {
		dvh.AssertFileExists(filePath, minSize, description+" after operation")
	}

	return result.Error
}

// AssertSnapshotCreation validates complete snapshot creation lifecycle
func (dvh *DataValidationHelper) AssertSnapshotCreation(
	takeSnapshot func() (interface{}, error),
	getSnapshot func(id string) interface{},
	snapshotID string,
	filePath string,
	description string,
) (interface{}, error) {
	// Step 1: Verify initial state
	dvh.AssertFileNotExists(filePath, description+" initial state")

	// Step 2: Execute snapshot creation
	snapshot, err := takeSnapshot()
	if err != nil {
		return snapshot, err
	}

	// Step 3: Verify data was actually created
	dvh.AssertFileExists(filePath, 1024, description+" file creation") // Min 1KB for image

	// Step 4: Verify metadata
	if snapshotID != "" {
		retrievedSnapshot := getSnapshot(snapshotID)
		assert.NotNil(dvh.t, retrievedSnapshot, description+" retrieval")
	}

	return snapshot, nil
}

// AssertRecordingLifecycle validates complete recording lifecycle
func (dvh *DataValidationHelper) AssertRecordingLifecycle(
	startRecording func() (interface{}, error),
	stopRecording func(id string) error,
	recordingID string,
	filePath string,
	recordDuration time.Duration,
	description string,
) (interface{}, error) {
	// Step 1: Verify initial state
	dvh.AssertFileNotExists(filePath, description+" initial state")

	// Step 2: Start recording
	session, err := startRecording()
	if err != nil {
		return session, err
	}

	// Step 3: Verify recording is active (progressive validation)
	time.Sleep(2 * time.Second)                                        // Allow time for recording to start
	dvh.AssertFileExists(filePath, 0, description+" during recording") // File should exist during recording

	// Step 4: Record for specified duration
	time.Sleep(recordDuration)

	// Step 5: Stop recording
	err = stopRecording(recordingID)
	if err != nil {
		return session, err
	}

	// Step 6: Verify final state
	dvh.AssertFileExists(filePath, 10000, description+" final state") // Min 10KB for video content

	return session, nil
}

// AssertStateTransition validates state changes throughout an operation
func (dvh *DataValidationHelper) AssertStateTransition(
	operation func() error,
	getState func() string,
	expectedStates []string,
	description string,
) error {
	// Step 1: Record initial state
	initialState := getState()

	// Step 2: Execute operation
	err := operation()

	// Step 3: Verify state transitions occurred
	if err == nil {
		finalState := getState()

		// Verify final state is in expected states
		found := false
		for _, expected := range expectedStates {
			if finalState == expected {
				found = true
				break
			}
		}
		assert.True(dvh.t, found,
			"%s: Final state '%s' should be one of %v (initial: '%s')",
			description, finalState, expectedStates, initialState)
	}

	return err
}

// AssertErrorCondition validates that error conditions work correctly
func (dvh *DataValidationHelper) AssertErrorCondition(
	operation func() error,
	expectedErrorContains string,
	shouldCreateFile bool,
	filePath string,
	description string,
) {
	// Step 1: Execute operation that should fail
	err := operation()

	// Step 2: Verify error occurred and contains expected message
	require.Error(dvh.t, err, "%s: Operation should fail", description)
	if expectedErrorContains != "" {
		assert.Contains(dvh.t, err.Error(), expectedErrorContains,
			"%s: Error should contain expected message", description)
	}

	// Step 3: Verify file state matches expectation
	if shouldCreateFile {
		dvh.AssertFileExists(filePath, 0, description+" file should exist despite error")
	} else {
		dvh.AssertFileNotExists(filePath, description+" file should not exist on error")
	}
}

// AssertConcurrentOperations validates concurrent operations work correctly
func (dvh *DataValidationHelper) AssertConcurrentOperations(
	operations []func() error,
	expectedSuccessCount int,
	description string,
) {
	var wg sync.WaitGroup
	results := make(chan error, len(operations))

	// Start all operations concurrently
	for _, op := range operations {
		wg.Add(1)
		go func(operation func() error) {
			defer wg.Done()
			results <- operation()
		}(op)
	}

	// Wait for all operations to complete
	wg.Wait()
	close(results)

	// Count successful operations
	successCount := 0
	for err := range results {
		if err == nil {
			successCount++
		}
	}

	assert.Equal(dvh.t, expectedSuccessCount, successCount,
		"%s: Should have %d successful operations", description, expectedSuccessCount)
}

// CreateTestSnapshotPath creates a test snapshot file path
func (dvh *DataValidationHelper) CreateTestSnapshotPath(filename string) string {
	return filepath.Join(dvh.tempDir, fmt.Sprintf("test_snapshot_%s.jpg", filename))
}

// CreateTestRecordingPath creates a test recording file path
func (dvh *DataValidationHelper) CreateTestRecordingPath(filename string) string {
	return filepath.Join(dvh.tempDir, fmt.Sprintf("test_recording_%s.mp4", filename))
}

// BuildMediaMTXFilePath builds a complete file path for ANY MediaMTX operation (recording, snapshot, etc.)
// This eliminates hardcoded path construction across all modules and unifies the pattern
func BuildMediaMTXFilePath(basePath, cameraID, filename string, useSubdirs bool, format string) string {
	// Handle file extension - MediaMTX handles this automatically, but for testing we need to predict it
	ext := format
	if ext == "" {
		// No default - let MediaMTX handle it (it adds extensions automatically)
		ext = ""
	} else {
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		// Add extension to filename if not present
		if !strings.HasSuffix(filename, ext) {
			filename = filename + ext
		}
	}

	// Handle subdirectories based on configuration (SAME PATTERN for all MediaMTX operations)
	if useSubdirs {
		return filepath.Join(basePath, cameraID, filename)
	}
	return filepath.Join(basePath, filename)
}

// BuildRecordingFilePath is a convenience wrapper for recordings
func BuildRecordingFilePath(basePath, cameraID, filename string, useSubdirs bool, format string) string {
	return BuildMediaMTXFilePath(basePath, cameraID, filename, useSubdirs, format)
}

// BuildSnapshotFilePath is a convenience wrapper for snapshots
func BuildSnapshotFilePath(basePath, cameraID, filename string, useSubdirs bool, format string) string {
	return BuildMediaMTXFilePath(basePath, cameraID, filename, useSubdirs, format)
}

// WaitForFileCreation waits for a file to be created with timeout
func (dvh *DataValidationHelper) WaitForFileCreation(filePath string, timeout time.Duration, description string) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(filePath); err == nil {
			dvh.t.Logf("✅ %s: File created successfully: %s", description, filePath)
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}

	dvh.t.Errorf("❌ %s: Timeout waiting for file creation: %s", description, filePath)
	return false
}

// AssertFileAccessible validates that a file is accessible and readable
func (dvh *DataValidationHelper) AssertFileAccessible(filePath string, description string) {
	// Check file exists
	dvh.AssertFileExists(filePath, 0, description)

	// Check file is readable
	file, err := os.Open(filePath)
	if err != nil {
		dvh.t.Errorf("%s: File should be readable: %s (error: %v)", description, filePath, err)
		return
	}
	defer file.Close()

	// Try to read a small amount to verify accessibility
	buffer := make([]byte, 1024)
	_, err = file.Read(buffer)
	assert.NoError(dvh.t, err, "%s: File should be readable: %s", description, filePath)
}
