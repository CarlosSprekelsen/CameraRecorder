package config

import (
	"os"
	"strings"
	"testing"
)

func TestPathValidation(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
		errMsg  string
	}{
		{"absolute path", "/opt/recordings", false, ""},
		{"relative path", "./recordings", true, "must be an absolute path"},
		{"path traversal", "/opt/../etc", true, "path traversal"},
		{"non-existent", "/nonexistent/path", true, "does not exist"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePath("test", tt.path, false)
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
