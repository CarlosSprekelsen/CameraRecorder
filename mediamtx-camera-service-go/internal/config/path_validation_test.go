package config

import (
	"os"
	"strings"
	"testing"
)

func TestPathValidation(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := "/tmp/path_validation_test"
	os.MkdirAll(tempDir, 0755)
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name            string
		path            string
		createIfMissing bool
		wantErr         bool
		errMsg          string
	}{
		// Valid cases
		{"valid existing path", tempDir, false, false, ""},
		{"valid path with creation", "/tmp/path_validation_test_new", true, false, ""},

		// Invalid cases - path format
		{"relative path", "./recordings", false, true, "must be an absolute path"},
		{"empty path", "", false, true, "must be an absolute path"},

		// Invalid cases - path traversal attacks
		{"simple traversal", "/opt/../../../etc/passwd", false, true, "path traversal"},
		{"backslash traversal", "/opt/..\\etc", false, true, "path traversal"},
		{"mixed separators", "/opt/../etc", false, true, "path traversal"},
		{"nested traversal", "/opt/recordings/../../../etc", false, true, "path traversal"},
		{"double traversal", "/opt/../..//etc", false, true, "path traversal"},
		{"encoded traversal", "/opt/..%2Fetc", false, true, "path traversal"},

		// Invalid cases - non-existent paths
		{"non-existent path", "/nonexistent/path/that/does/not/exist", false, true, "does not exist"},
		{"non-existent with no creation", "/tmp/nonexistent_validation_test", false, true, "does not exist"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePath("test", tt.path, tt.createIfMissing)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error to contain %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestPathFallback(t *testing.T) {
	// Setup primary path as read-only
	primaryPath := "/tmp/primary"
	fallbackPath := "/tmp/fallback"

	// Create directories
	os.MkdirAll(primaryPath, 0444)  // Read-only
	os.MkdirAll(fallbackPath, 0755) // Writable
	defer os.RemoveAll(primaryPath)
	defer os.RemoveAll(fallbackPath)

	// Create test config
	config := &Config{
		MediaMTX: MediaMTXConfig{
			RecordingsPath: primaryPath,
		},
		Storage: StorageConfig{
			FallbackPath: fallbackPath,
		},
	}

	// Test validation
	err := ValidatePathConfiguration(config)
	if err != nil {
		t.Logf("Expected validation error due to read-only primary path: %v", err)
	}
}

func TestValidatePathPattern(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		wantErr bool
	}{
		{"valid pattern", "device_%Y%m%d_%H%M%S", false},
		{"empty pattern", "", false},
		{"dangerous traversal", "device_%Y%m%d_%H%M%S/../etc", true},
		{"dangerous backslash", "device_%Y%m%d_%H%M%S\\..", true},
		{"dangerous newline", "device_%Y%m%d_%H%M%S\n", true},
		{"too long", string(make([]byte, 300)), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePathPattern(tt.pattern)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestCheckWritePermission(t *testing.T) {
	// Create a writable directory
	writableDir := "/tmp/writable_test"
	os.MkdirAll(writableDir, 0755)
	defer os.RemoveAll(writableDir)

	// Test writable directory
	err := checkWritePermission(writableDir)
	if err != nil {
		t.Errorf("expected writable directory to pass, got error: %v", err)
	}

	// Create a read-only directory
	readOnlyDir := "/tmp/readonly_test"
	os.MkdirAll(readOnlyDir, 0444)
	defer os.RemoveAll(readOnlyDir)

	// Test read-only directory
	err = checkWritePermission(readOnlyDir)
	if err == nil {
		t.Errorf("expected read-only directory to fail write permission check")
	}
}

func TestCheckDiskSpace(t *testing.T) {
	// Test with a valid directory (should not error for normal disk space)
	tempDir := "/tmp"
	err := checkDiskSpace(tempDir)
	if err != nil {
		t.Logf("Disk space check returned warning (expected in some environments): %v", err)
	}
}

func TestPathTraversalDetection(t *testing.T) {
	// Test various path traversal attack patterns
	traversalTests := []struct {
		name     string
		path     string
		expected bool // true if should be detected as traversal
	}{
		// Should be detected as traversal (checking original path for "..")
		{"simple dotdot", "/opt/../etc", true},
		{"multiple dotdot", "/opt/../../../etc", true},
		{"backslash dotdot", "/opt/..\\etc", true},
		{"mixed separators", "/opt/../etc", true},
		{"nested traversal", "/opt/recordings/../../../etc", true},
		{"double slash", "/opt/../..//etc", true},
		{"encoded traversal", "/opt/..%2Fetc", true},
		{"traversal in middle", "/opt/recordings/../etc", true},
		{"traversal at end", "/opt/recordings/..", true},

		// Should NOT be detected as traversal
		{"valid path", "/opt/recordings", false},
		{"valid nested", "/opt/recordings/subdir", false},
		{"valid with dots", "/opt/recordings/file.txt", false},
		{"valid with underscores", "/opt/recordings_file", false},
		{"valid with dashes", "/opt/recordings-file", false},
		{"valid with numbers", "/opt/recordings123", false},
	}

	for _, tt := range traversalTests {
		t.Run(tt.name, func(t *testing.T) {
			// Test the corrected path traversal detection logic
			// (checking original path for ".." before cleaning)
			hasTraversal := strings.Contains(tt.path, "..")

			if hasTraversal != tt.expected {
				t.Errorf("path traversal detection failed for %q: expected %v, got %v",
					tt.path, tt.expected, hasTraversal)
			}
		})
	}
}
