//go:build unit || integration || performance

/*
Test Cleanup Utilities - STANDALONE CLEANUP TOOLS

This file provides standalone utilities for cleaning up test artifacts and preventing
disk space issues. These can be run independently of tests to clean up accumulated files.

Usage:
1. Run cleanup before tests: go run tests/utils/test_cleanup.go
2. Run cleanup after test failures: go run tests/utils/test_cleanup.go --force
3. Run cleanup in CI/CD: go run tests/utils/test_cleanup.go --ci

Requirements Coverage:
- REQ-TEST-001: Test environment setup
- REQ-TEST-002: Test data preparation
- REQ-TEST-003: Test configuration management
- REQ-TEST-004: Test authentication setup
- REQ-TEST-005: Test evidence collection
- REQ-TEST-006: Real MediaMTX controller setup
- REQ-TEST-007: Test-specific MediaMTX configuration
- REQ-TEST-008: Test artifact cleanup
- REQ-TEST-009: Disk space management

Test Categories: Test Infrastructure
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package utils

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/sys/unix"
)

// TestCleanupConfig holds cleanup configuration
type TestCleanupConfig struct {
	ForceCleanup bool
	CICleanup    bool
	DryRun       bool
	Verbose      bool
	MaxAge       time.Duration
	Paths        []string
}

// TestArtifact represents a test artifact that can be cleaned up
type TestArtifact struct {
	Path         string
	Size         int64
	ModifiedTime time.Time
	Type         string
}

// RunTestCleanup runs the test cleanup process with the given configuration
func RunTestCleanup(config *TestCleanupConfig) (int64, int) {
	// Log cleanup start
	log.Printf("Starting test cleanup process...")
	log.Printf("Configuration: Force=%v, CI=%v, DryRun=%v, Verbose=%v",
		config.ForceCleanup, config.CICleanup, config.DryRun, config.Verbose)

	// Check disk space before cleanup
	checkDiskSpace("/tmp", "Before cleanup")

	// Perform cleanup
	cleanedBytes, cleanedCount := performCleanup(config)

	// Check disk space after cleanup
	checkDiskSpace("/tmp", "After cleanup")

	// Log results
	log.Printf("Cleanup completed: %d files removed, %.2f MB freed",
		cleanedCount, float64(cleanedBytes)/(1024*1024))

	return cleanedBytes, cleanedCount
}

// RunDefaultTestCleanup runs the test cleanup process with default configuration
func RunDefaultTestCleanup() (int64, int) {
	config := &TestCleanupConfig{
		ForceCleanup: false,
		CICleanup:    false,
		DryRun:       false,
		Verbose:      false,
		MaxAge:       1 * time.Hour,
		Paths:        []string{"/tmp", "/var/tmp", os.TempDir()},
	}
	return RunTestCleanup(config)
}

// parseFlags parses command line flags
func parseFlags() *TestCleanupConfig {
	force := flag.Bool("force", false, "Force aggressive cleanup")
	ci := flag.Bool("ci", false, "CI/CD mode cleanup")
	dryRun := flag.Bool("dry-run", false, "Show what would be cleaned without actually cleaning")
	verbose := flag.Bool("verbose", false, "Verbose output")
	maxAge := flag.Duration("max-age", 1*time.Hour, "Maximum age of files to keep")

	flag.Parse()

	config := &TestCleanupConfig{
		ForceCleanup: *force,
		CICleanup:    *ci,
		DryRun:       *dryRun,
		Verbose:      *verbose,
		MaxAge:       *maxAge,
		Paths:        []string{"/tmp", "/var/tmp", os.TempDir()},
	}

	return config
}

// performCleanup performs the actual cleanup operation
func performCleanup(config *TestCleanupConfig) (int64, int) {
	var totalBytes int64
	var totalCount int

	for _, path := range config.Paths {
		if path == "" {
			continue
		}

		bytes, count := cleanupPath(path, config)
		totalBytes += bytes
		totalCount += count
	}

	return totalBytes, totalCount
}

// cleanupPath cleans up a specific path
func cleanupPath(path string, config *TestCleanupConfig) (int64, int) {
	if config.Verbose {
		log.Printf("Cleaning up path: %s", path)
	}

	var totalBytes int64
	var totalCount int

	// Clean up test files
	bytes, count := cleanupTestFiles(path, config)
	totalBytes += bytes
	totalCount += count

	// Clean up test directories
	bytes, count = cleanupTestDirectories(path, config)
	totalBytes += bytes
	totalCount += count

	// Clean up old files based on age
	if config.ForceCleanup || config.CICleanup {
		bytes, count = cleanupOldFiles(path, config)
		totalBytes += bytes
		totalCount += count
	}

	if config.Verbose {
		log.Printf("Path %s: %d files removed, %.2f MB freed",
			path, totalCount, float64(totalBytes)/(1024*1024))
	}

	return totalBytes, totalCount
}

// cleanupTestFiles removes test files based on patterns
func cleanupTestFiles(path string, config *TestCleanupConfig) (int64, int) {
	testPatterns := []string{
		"*.mp4",       // Test recordings
		"*.avi",       // Test recordings
		"*.mkv",       // Test recordings
		"*.jpg",       // Test snapshots
		"*.jpeg",      // Test snapshots
		"*.png",       // Test snapshots
		"*.h264",      // Test video streams
		"*.h265",      // Test video streams
		"test_*",      // Test-specific files
		"*_test.*",    // Test-specific files
		"camera_*",    // Camera test files
		"recording_*", // Recording test files
		"snapshot_*",  // Snapshot test files
	}

	var totalBytes int64
	var totalCount int

	for _, pattern := range testPatterns {
		matches, err := filepath.Glob(filepath.Join(path, pattern))
		if err != nil {
			if config.Verbose {
				log.Printf("Warning: Failed to glob pattern %s: %v", pattern, err)
			}
			continue
		}

		for _, match := range matches {
			bytes, removed := removeFile(match, config)
			if removed {
				totalBytes += bytes
				totalCount++
			}
		}
	}

	return totalBytes, totalCount
}

// cleanupTestDirectories removes test directories
func cleanupTestDirectories(path string, config *TestCleanupConfig) (int64, int) {
	testDirs := []string{
		"recordings",
		"snapshots",
		"streams",
		"temp",
		"cache",
		"test_*",
		"*_test",
		"camera_test_*",
		"mediamtx_test_*",
	}

	var totalBytes int64
	var totalCount int

	for _, dir := range testDirs {
		dirPath := filepath.Join(path, dir)
		if _, err := os.Stat(dirPath); err == nil {
			bytes, removed := removeDirectory(dirPath, config)
			if removed {
				totalBytes += bytes
				totalCount++
			}
		}
	}

	return totalBytes, totalCount
}

// cleanupOldFiles removes files older than the specified age
func cleanupOldFiles(path string, config *TestCleanupConfig) (int64, int) {
	cutoffTime := time.Now().Add(-config.MaxAge)

	var totalBytes int64
	var totalCount int

	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files with errors
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if file is old enough to remove
		if info.ModTime().Before(cutoffTime) {
			// Only remove files that look like test artifacts
			if isTestArtifact(filePath) {
				bytes, removed := removeFile(filePath, config)
				if removed {
					totalBytes += bytes
					totalCount++
				}
			}
		}

		return nil
	})

	if err != nil && config.Verbose {
		log.Printf("Warning: Error walking path %s: %v", path, err)
	}

	return totalBytes, totalCount
}

// isTestArtifact checks if a file looks like a test artifact
func isTestArtifact(filePath string) bool {
	fileName := strings.ToLower(filepath.Base(filePath))

	// Check for test-related patterns
	testPatterns := []string{
		"test", "camera", "recording", "snapshot", "stream",
		"mediamtx", "websocket", "jwt", "auth",
	}

	for _, pattern := range testPatterns {
		if strings.Contains(fileName, pattern) {
			return true
		}
	}

	// Check for test file extensions
	testExtensions := []string{".mp4", ".avi", ".mkv", ".jpg", ".jpeg", ".png", ".h264", ".h265"}
	ext := strings.ToLower(filepath.Ext(fileName))
	for _, testExt := range testExtensions {
		if ext == testExt {
			return true
		}
	}

	return false
}

// removeFile removes a single file
func removeFile(filePath string, config *TestCleanupConfig) (int64, bool) {
	info, err := os.Stat(filePath)
	if err != nil {
		return 0, false
	}

	fileSize := info.Size()

	if config.DryRun {
		log.Printf("DRY RUN: Would remove file %s (%.2f MB)",
			filePath, float64(fileSize)/(1024*1024))
		return fileSize, false
	}

	if err := os.Remove(filePath); err != nil {
		if config.Verbose {
			log.Printf("Warning: Failed to remove file %s: %v", filePath, err)
		}
		return 0, false
	}

	if config.Verbose {
		log.Printf("Removed file: %s (%.2f MB)",
			filePath, float64(fileSize)/(1024*1024))
	}

	return fileSize, true
}

// removeDirectory removes a directory and its contents
func removeDirectory(dirPath string, config *TestCleanupConfig) (int64, bool) {
	if config.DryRun {
		log.Printf("DRY RUN: Would remove directory %s", dirPath)
		return 0, false
	}

	if err := os.RemoveAll(dirPath); err != nil {
		if config.Verbose {
			log.Printf("Warning: Failed to remove directory %s: %v", dirPath, err)
		}
		return 0, false
	}

	if config.Verbose {
		log.Printf("Removed directory: %s", dirPath)
	}

	return 0, true // Directory size calculation is complex, return 0
}

// checkDiskSpace checks available disk space
func checkDiskSpace(path, context string) {
	var stat unix.Statfs_t
	err := unix.Statfs(path, &stat)
	if err != nil {
		log.Printf("Warning: Could not check disk space for %s: %v", path, err)
		return
	}

	// Calculate available space in GB
	availableBytes := stat.Bavail * uint64(stat.Bsize)
	availableGB := float64(availableBytes) / (1024 * 1024 * 1024)

	log.Printf("%s: %.2f GB available at %s", context, availableGB, path)

	// Warn if available space is low
	if availableGB < 1.0 {
		log.Printf("WARNING: Low disk space detected! Available: %.2f GB", availableGB)
	}
}
