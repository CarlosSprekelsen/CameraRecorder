/*
Test Environment Loader

Loads environment variables from .test_env file for Go tests.
This ensures JWT secrets and other test configuration are sourced from
the shell environment file rather than hardcoded in tests or fixtures.

Architecture Compliance:
- No hardcoded secrets in code or YAML
- Uses existing .test_env mechanism from tests/tools/setup_test_environment.sh
- Idempotent loading for parallel test safety
- Only loads CAMERA_SERVICE_* prefixed variables

Requirements Coverage:
- REQ-TEST-001: Test environment setup and management
- REQ-SEC-001: No hardcoded secrets in test code
- REQ-CONFIG-001: Environment-driven configuration loading
*/

package testutils

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

var (
	envLoadedOnce sync.Once
	envLoaded     bool
)

// LoadDotTestEnvOnce loads environment variables from .test_env file
// This function is idempotent and safe for parallel test execution.
// Only loads CAMERA_SERVICE_* prefixed variables to avoid polluting the environment.
func LoadDotTestEnvOnce(t *testing.T) {
	envLoadedOnce.Do(func() {
		loadDotTestEnv(t)
		envLoaded = true
	})
}

// loadDotTestEnv performs the actual environment loading
func loadDotTestEnv(t *testing.T) {
	// Find .test_env file relative to project root
	projectRoot := findProjectRoot(t)
	testEnvPath := filepath.Join(projectRoot, ".test_env")

	// Check if .test_env file exists
	if _, err := os.Stat(testEnvPath); os.IsNotExist(err) {
		t.Logf("Warning: .test_env file not found at %s - test environment variables not loaded", testEnvPath)
		return
	}

	// Read and parse .test_env file
	file, err := os.Open(testEnvPath)
	if err != nil {
		t.Logf("Warning: Failed to open .test_env file: %v", err)
		return
	}
	defer file.Close()

	loadedCount := 0
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse shell export format: export KEY="value"
		if strings.HasPrefix(line, "export ") {
			// Remove "export " prefix
			exportLine := strings.TrimPrefix(line, "export ")

			// Parse key=value pairs
			parts := strings.SplitN(exportLine, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])

				// Remove quotes if present
				value = strings.Trim(value, "\"'")

				// Only load CAMERA_SERVICE_* prefixed variables
				if strings.HasPrefix(key, "CAMERA_SERVICE_") {
					os.Setenv(key, value)
					loadedCount++
					t.Logf("Loaded test env: %s", key)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		t.Logf("Warning: Error reading .test_env file: %v", err)
		return
	}

	t.Logf("Loaded %d CAMERA_SERVICE_* environment variables from .test_env", loadedCount)
}

// findProjectRoot finds the project root directory by looking for go.mod
func findProjectRoot(t *testing.T) string {
	// Start from current directory and walk up
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	for {
		// Check if go.mod exists in current directory
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return dir
		}

		// Move to parent directory
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root
			break
		}
		dir = parent
	}

	// Fallback to current directory if go.mod not found
	t.Logf("Warning: go.mod not found, using current directory as project root")
	return "."
}
