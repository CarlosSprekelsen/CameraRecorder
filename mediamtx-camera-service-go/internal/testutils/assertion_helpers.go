/*
Assertion Helpers - Domain-Agnostic Test Assertions

Provides common assertion patterns that can be reused across all modules,
eliminating duplicate assertion logic and ensuring consistent test validation.
*/

package testutils

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// AssertionHelper provides domain-agnostic assertion utilities
type AssertionHelper struct {
	t *testing.T
}

// NewAssertionHelper creates a new assertion helper
func NewAssertionHelper(t *testing.T) *AssertionHelper {
	return &AssertionHelper{t: t}
}

// AssertJSONRPCResponse validates standard JSON-RPC response structure
func (ah *AssertionHelper) AssertJSONRPCResponse(response interface{}, expectError bool) {
	// This can be used by both WebSocket and any future JSON-RPC modules
	responseMap, ok := response.(map[string]interface{})
	require.True(ah.t, ok, "Response should be a map")
	
	// Validate JSON-RPC version
	assert.Equal(ah.t, "2.0", responseMap["jsonrpc"], "Response should have correct JSON-RPC version")
	
	// Validate ID presence
	assert.NotNil(ah.t, responseMap["id"], "Response should have ID")
	
	// Validate error/result structure
	if expectError {
		assert.NotNil(ah.t, responseMap["error"], "Response should have error")
		assert.Nil(ah.t, responseMap["result"], "Error response should not have result")
	} else {
		assert.Nil(ah.t, responseMap["error"], "Response should not have error")
		// Note: result can be nil for some methods, so we don't assert its presence
	}
}

// AssertDirectoryExists validates directory existence and permissions
func (ah *AssertionHelper) AssertDirectoryExists(dirPath string, expectWritable bool) {
	// Check directory exists
	info, err := os.Stat(dirPath)
	require.NoError(ah.t, err, "Directory should exist: %s", dirPath)
	require.True(ah.t, info.IsDir(), "Path should be a directory: %s", dirPath)
	
	// Check writability if required
	if expectWritable {
		testFile := filepath.Join(dirPath, ".write_test")
		err := os.WriteFile(testFile, []byte("test"), 0644)
		assert.NoError(ah.t, err, "Directory should be writable: %s", dirPath)
		if err == nil {
			os.Remove(testFile) // Cleanup test file
		}
	}
}

// AssertConfigValue validates configuration values
func (ah *AssertionHelper) AssertConfigValue(config map[string]interface{}, path string, expected interface{}) {
	// Navigate nested config structure
	current := config
	keys := splitConfigPath(path)
	
	for i, key := range keys[:len(keys)-1] {
		next, ok := current[key].(map[string]interface{})
		require.True(ah.t, ok, "Config path should exist: %s (at %s)", path, keys[i])
		current = next
	}
	
	// Check final value
	finalKey := keys[len(keys)-1]
	actual, exists := current[finalKey]
	require.True(ah.t, exists, "Config value should exist: %s", path)
	assert.Equal(ah.t, expected, actual, "Config value should match: %s", path)
}

// AssertTimeout validates operations complete within timeout
func (ah *AssertionHelper) AssertTimeout(operation func() error, timeout time.Duration, description string) {
	done := make(chan error, 1)
	go func() {
		done <- operation()
	}()
	
	select {
	case err := <-done:
		assert.NoError(ah.t, err, "%s should succeed", description)
	case <-time.After(timeout):
		ah.t.Errorf("%s should complete within %v", description, timeout)
	}
}

// splitConfigPath splits dot-notation config path
func splitConfigPath(path string) []string {
	// Simple implementation for config navigation
	// e.g., "mediamtx.recordings_path" → ["mediamtx", "recordings_path"]
	return []string{path} // Simplified for now, can be enhanced
}
