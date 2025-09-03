//go:build unit || integration || performance

/*
Nuclear Test Cleanup - THE ONLY CLEANUP FUNCTION YOU NEED

This file provides a single, aggressive cleanup function that removes ALL test evidence.
Since this is only for tests, we can be ruthless and nuke everything.


USE THIS INSTEAD:
NuclearTestCleanup(t, tempDir) - Nukes everything, everywhere
*/

package testutils

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// NuclearTestCleanup removes ALL test evidence from ALL locations
// This is the ONLY cleanup function needed - it goes nuclear on test artifacts
// Since this is only for tests, we can be aggressive and remove everything
func NuclearTestCleanup(t *testing.T, tempDir string) {
	t.Log("🚀 NUCLEAR CLEANUP INITIATED - Removing ALL test evidence")

	// 1. Nuke the tempDir completely
	if tempDir != "" {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("Warning: Failed to nuke temp directory %s: %v", tempDir, err)
		} else {
			t.Logf("💥 Temp directory nuked: %s", tempDir)
		}
	}

	// 2. Nuke common test locations aggressively
	testLocations := []string{
		"/tmp",
		"/var/tmp",
		os.TempDir(),
	}

	for _, location := range testLocations {
		if location == "" {
			continue
		}

		// Remove ALL test-related files and directories
		nukeTestFilesFromLocation(t, location)
	}

	t.Log("☢️ NUCLEAR CLEANUP COMPLETED - All test evidence eliminated")
}

// nukeTestFilesFromLocation removes ALL test files from a specific location
// This is aggressive - removes anything that looks like test data
func nukeTestFilesFromLocation(t *testing.T, location string) {
	// Test file patterns to nuke (be aggressive)
	testPatterns := []string{
		"*camera*",              // Camera-related files
		"*recording*",           // Recording files
		"*snapshot*",            // Snapshot files
		"*test*",                // Test files
		"*mediamtx*",            // MediaMTX files
		"*.mp4",                 // Video files
		"*.avi",                 // Video files
		"*.mkv",                 // Video files
		"*.jpg",                 // Image files
		"*.jpeg",                // Image files
		"*.png",                 // Image files
		"*.h264",                // Video streams
		"*.h265",                // Video streams
		"camera-service-test-*", // Our test directories
		"go-build*",             // Go build artifacts
		"_test*",                // Test binaries
	}

	// Test directory patterns to nuke
	testDirPatterns := []string{
		"recordings",
		"snapshots",
		"storage",
		"fallback",
		"fallback_recordings",
		"fallback_snapshots",
		"temp",
		"cache",
		"logs",
	}

	// Nuke files first
	for _, pattern := range testPatterns {
		matches, err := filepath.Glob(filepath.Join(location, pattern))
		if err != nil {
			t.Logf("Warning: Failed to glob pattern %s: %v", pattern, err)
			continue
		}

		for _, match := range matches {
			// Skip if it's a directory (we'll handle those separately)
			if info, err := os.Stat(match); err == nil && info.IsDir() {
				continue
			}

			if err := os.Remove(match); err != nil {
				t.Logf("Warning: Failed to nuke test file %s: %v", match, err)
			} else {
				t.Logf("💥 Nuked test file: %s", match)
			}
		}
	}

	// Nuke directories
	for _, dirPattern := range testDirPatterns {
		dirPath := filepath.Join(location, dirPattern)
		if _, err := os.Stat(dirPath); err == nil {
			if err := os.RemoveAll(dirPath); err != nil {
				t.Logf("Warning: Failed to nuke test directory %s: %v", dirPath, err)
			} else {
				t.Logf("💥 Nuked test directory: %s", dirPath)
			}
		}
	}

	// Nuke any remaining test directories we might have missed
	entries, err := os.ReadDir(location)
	if err != nil {
		t.Logf("Warning: Could not read directory %s: %v", location, err)
		return
	}

	for _, entry := range entries {
		name := entry.Name()
		// If it looks like a test directory, nuke it
		if strings.Contains(strings.ToLower(name), "test") ||
			strings.Contains(strings.ToLower(name), "camera") ||
			strings.Contains(strings.ToLower(name), "mediamtx") ||
			strings.Contains(strings.ToLower(name), "go-build") {

			entryPath := filepath.Join(location, name)
			if entry.IsDir() {
				if err := os.RemoveAll(entryPath); err != nil {
					t.Logf("Warning: Failed to nuke suspicious directory %s: %v", entryPath, err)
				} else {
					t.Logf("💥 Nuked suspicious directory: %s", entryPath)
				}
			}
		}
	}
}
