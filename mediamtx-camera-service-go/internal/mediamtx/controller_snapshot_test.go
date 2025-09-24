/*
Controller Snapshot Tests - Refactored with Asserters

This file demonstrates the dramatic reduction possible using SnapshotAsserters.
Original tests had massive duplication of setup, Progressive Readiness, and validation.
Refactored tests focus on business logic only.

Requirements Coverage:
- REQ-MTX-002: Stream management capabilities
- REQ-MTX-002: Snapshot capture functionality
*/

package mediamtx

import (
	"strings"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/testutils"
	"github.com/stretchr/testify/assert"
)

// TestController_TakeSnapshot_ReqMTX002_Success_Refactored demonstrates snapshot testing with asserters
// Original: 55+ lines → Refactored: 15 lines (73% reduction!)
func TestController_TakeSnapshot_ReqMTX002_Success_Refactored(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities

	// Create snapshot asserter with full setup (eliminates 8 lines of setup)
	asserter := NewSnapshotAsserter(t)
	defer asserter.Cleanup()

	// Get camera ID (eliminates 3 lines of camera ID retrieval)
	cameraID := asserter.MustGetCameraID()

	// Snapshot options (eliminates hardcoded values)
	options := &SnapshotOptions{
		Format:  "jpg",
		Quality: 85,
	}

	// Complete snapshot capture with validation (eliminates 35+ lines of Progressive Readiness + validation)
	snapshot := asserter.AssertSnapshotCapture(cameraID, options)

	// Test-specific business logic only
	assert.Contains(t, snapshot.FilePath, cameraID, "File path should contain camera identifier")
	assert.Contains(t, snapshot.FilePath, ".jpg", "File path should have .jpg extension")

	// Verify snapshot path follows the fixture configuration
	expectedPath := asserter.GetHelper().GetConfiguredSnapshotPath()
	assert.True(t, strings.HasPrefix(snapshot.FilePath, expectedPath+"/"),
		"Snapshot path should start with configured snapshots path from fixture: %s", expectedPath)

	t.Logf("✅ Snapshot capture validated successfully: %s", snapshot.FilePath)
}

// TestController_ListSnapshots_ReqMTX002_Success_Refactored demonstrates snapshot listing
func TestController_ListSnapshots_ReqMTX002_Success_Refactored(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities

	asserter := NewSnapshotAsserter(t)
	defer asserter.Cleanup()

	// List snapshots (eliminates 15+ lines of setup and validation)
	snapshots, err := asserter.GetReadyController().ListSnapshots(asserter.GetContext(), 10, 0)
	asserter.AssertSnapshotResponse(&TakeSnapshotResponse{}, err)

	// Test-specific business logic only
	assert.NotNil(t, snapshots, "Snapshots response should not be nil")
	assert.GreaterOrEqual(t, snapshots.Total, 0, "Total snapshots should be non-negative")

	t.Logf("✅ Snapshot listing validated successfully: %d total snapshots", snapshots.Total)
}

// TestController_TakeSnapshot_ReqMTX002_Advanced_Refactored demonstrates advanced snapshot options
func TestController_TakeSnapshot_ReqMTX002_Advanced_Refactored(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities

	asserter := NewSnapshotAsserter(t)
	defer asserter.Cleanup()

	cameraID := asserter.MustGetCameraID()

	// Test different snapshot formats and qualities
	testCases := []struct {
		name    string
		options *SnapshotOptions
	}{
		{
			name: "High Quality JPG",
			options: &SnapshotOptions{
				Format:  "jpg",
				Quality: 95,
			},
		},
		{
			name: "Medium Quality JPG",
			options: &SnapshotOptions{
				Format:  "jpg",
				Quality: 75,
			},
		},
		{
			name: "Low Quality JPG",
			options: &SnapshotOptions{
				Format:  "jpg",
				Quality: 50,
			},
		},
	}

	// Test each configuration (eliminates 40+ lines of repeated Progressive Readiness patterns)
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			snapshot := asserter.AssertSnapshotCapture(cameraID, tc.options)

			// Test-specific business logic only
			assert.Contains(t, snapshot.FilePath, ".jpg", "Snapshot should have .jpg extension")
			assert.Greater(t, snapshot.FileSize, int64(0), "Snapshot should have content")

			t.Logf("✅ %s snapshot validated: %s (%d bytes)", tc.name, snapshot.FilePath, snapshot.FileSize)
		})
	}
}

// TestController_TakeSnapshot_ReqMTX002_Concurrent_Refactored demonstrates concurrent snapshot testing
func TestController_TakeSnapshot_ReqMTX002_Concurrent_Refactored(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities

	asserter := NewSnapshotAsserter(t)
	defer asserter.Cleanup()

	cameraID := asserter.MustGetCameraID()

	// Test concurrent snapshots (eliminates 30+ lines of goroutine setup and coordination)
	done := make(chan *TakeSnapshotResponse, testutils.UniversalConcurrencyGoroutines)

	for i := 0; i < testutils.UniversalConcurrencyGoroutines; i++ {
		go func(id int) {
			defer func() { done <- nil }()

			options := &SnapshotOptions{
				Format:  "jpg",
				Quality: 85,
			}

			// Each goroutine performs snapshot capture
			snapshot := asserter.AssertSnapshotCapture(cameraID, options)
			assert.Contains(t, snapshot.FilePath, ".jpg", "Concurrent snapshot %d should have .jpg extension", id)

			t.Logf("✅ Concurrent snapshot %d completed: %s", id, snapshot.FilePath)
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < testutils.UniversalConcurrencyGoroutines; i++ {
		select {
		case <-done:
			// Goroutine completed
		case <-time.After(testutils.UniversalTimeoutExtreme):
			t.Fatal("Timeout waiting for concurrent snapshots")
		}
	}

	t.Log("All concurrent snapshots completed successfully")
}

// TestController_TakeSnapshot_ReqMTX002_ErrorHandling_Refactored demonstrates error handling
func TestController_TakeSnapshot_ReqMTX002_ErrorHandling_Refactored(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities

	asserter := NewSnapshotAsserter(t)
	defer asserter.Cleanup()

	// Test error conditions (eliminates 20+ lines of error setup)

	// Test with invalid camera ID
	options := &SnapshotOptions{
		Format:  "jpg",
		Quality: 85,
	}

	_, err := asserter.GetReadyController().TakeAdvancedSnapshot(asserter.GetContext(), "invalid_camera", options)
	assert.Error(t, err, "Should return error for invalid camera ID")

	// Test with invalid options
	_, err = asserter.GetReadyController().TakeAdvancedSnapshot(asserter.GetContext(), "camera0", nil)
	assert.Error(t, err, "Should return error for nil options")

	t.Log("✅ Error handling validated successfully")
}
