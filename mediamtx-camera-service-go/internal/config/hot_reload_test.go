/*
Hot reload configuration management unit tests.

Requirements Coverage:
- REQ-CONFIG-005: Hot reload capability
- REQ-CONFIG-006: File change detection
- REQ-CONFIG-007: Debouncing of rapid changes
- REQ-CONFIG-008: Error handling during reload
- REQ-CONFIG-009: Thread-safe configuration updates

Test Categories: Unit
API Documentation Reference: docs/api/json_rpc_methods.md
*/

//go:build unit
// +build unit

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfigWatcher(t *testing.T) {
	tests := []struct {
		name          string
		configPath    string
		callback      func(*Config) error
		expectSuccess bool
	}{
		{
			name:          "valid config watcher creation",
			configPath:    "test-config.yaml",
			callback:      func(*Config) error { return nil },
			expectSuccess: true,
		},
		{
			name:          "empty config path",
			configPath:    "",
			callback:      func(*Config) error { return nil },
			expectSuccess: true, // fsnotify allows empty paths
		},
		{
			name:          "nil callback",
			configPath:    "test-config.yaml",
			callback:      nil,
			expectSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			watcher, err := NewConfigWatcher(tt.configPath, tt.callback)
			if tt.expectSuccess {
				require.NoError(t, err)
				require.NotNil(t, watcher)
				assert.Equal(t, tt.configPath, watcher.configPath)
				assert.False(t, watcher.isRunning)
			} else {
				require.Error(t, err)
				assert.Nil(t, watcher)
			}
		})
	}
}

func TestConfigWatcher_Start_Stop(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.yaml")
	
	// Write initial config
	err := os.WriteFile(configPath, []byte(`
server:
  host: "localhost"
  port: 8080
`), 0644)
	require.NoError(t, err)

	// Create watcher
	callbackCalled := false
	var callbackMutex sync.Mutex
	callback := func(*Config) error {
		callbackMutex.Lock()
		defer callbackMutex.Unlock()
		callbackCalled = true
		return nil
	}

	watcher, err := NewConfigWatcher(configPath, callback)
	require.NoError(t, err)
	require.NotNil(t, watcher)

	// Test Start
	err = watcher.Start()
	require.NoError(t, err)
	assert.True(t, watcher.IsRunning())

	// Test starting again (should fail)
	err = watcher.Start()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already running")

	// Test Stop
	err = watcher.Stop()
	require.NoError(t, err)
	assert.False(t, watcher.IsRunning())

	// Test stopping again (should succeed)
	err = watcher.Stop()
	require.NoError(t, err)
	
	// Use callbackCalled to avoid unused variable warning
	_ = callbackCalled
}

func TestConfigWatcher_Start_FileNotExists(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "non-existent-config.yaml")

	watcher, err := NewConfigWatcher(configPath, func(*Config) error { return nil })
	require.NoError(t, err)

	err = watcher.Start()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}

func TestConfigWatcher_FileChangeDetection(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.yaml")
	
	// Write initial config
	err := os.WriteFile(configPath, []byte(`
server:
  host: "localhost"
  port: 8080
`), 0644)
	require.NoError(t, err)

	// Create watcher with callback
	callbackCalled := false
	var callbackMutex sync.Mutex
	callback := func(*Config) error {
		callbackMutex.Lock()
		defer callbackMutex.Unlock()
		callbackCalled = true
		return nil
	}

	watcher, err := NewConfigWatcher(configPath, callback)
	require.NoError(t, err)

	// Start watcher
	err = watcher.Start()
	require.NoError(t, err)
	defer watcher.Stop()

	// Wait a bit for watcher to be ready
	time.Sleep(100 * time.Millisecond)

	// Modify the config file
	err = os.WriteFile(configPath, []byte(`
server:
  host: "127.0.0.1"
  port: 9090
`), 0644)
	require.NoError(t, err)

	// Wait for callback to be called
	time.Sleep(1 * time.Second)

	callbackMutex.Lock()
	assert.True(t, callbackCalled, "Callback should have been called")
	callbackMutex.Unlock()
}

func TestConfigWatcher_Debouncing(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.yaml")
	
	// Write initial config
	err := os.WriteFile(configPath, []byte(`
server:
  host: "localhost"
  port: 8080
`), 0644)
	require.NoError(t, err)

	// Create watcher with callback counter
	callbackCount := 0
	var callbackMutex sync.Mutex
	callback := func(*Config) error {
		callbackMutex.Lock()
		defer callbackMutex.Unlock()
		callbackCount++
		return nil
	}

	watcher, err := NewConfigWatcher(configPath, callback)
	require.NoError(t, err)

	// Start watcher
	err = watcher.Start()
	require.NoError(t, err)
	defer watcher.Stop()

	// Wait a bit for watcher to be ready
	time.Sleep(100 * time.Millisecond)

	// Rapidly modify the config file multiple times
	for i := 0; i < 5; i++ {
		err = os.WriteFile(configPath, []byte(fmt.Sprintf(`
server:
  host: "localhost"
  port: %d
`, 8080+i)), 0644)
		require.NoError(t, err)
		time.Sleep(50 * time.Millisecond) // Less than debounce interval
	}

	// Wait for debouncing to complete
	time.Sleep(1 * time.Second)

	callbackMutex.Lock()
	assert.LessOrEqual(t, callbackCount, 2, "Should have debounced rapid changes")
	callbackMutex.Unlock()
}

func TestConfigWatcher_ErrorHandling(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.yaml")
	
	// Write initial config
	err := os.WriteFile(configPath, []byte(`
server:
  host: "localhost"
  port: 8080
`), 0644)
	require.NoError(t, err)

	// Create watcher with error callback
	callback := func(*Config) error {
		return fmt.Errorf("simulated callback error")
	}

	watcher, err := NewConfigWatcher(configPath, callback)
	require.NoError(t, err)

	// Start watcher
	err = watcher.Start()
	require.NoError(t, err)
	defer watcher.Stop()

	// Wait a bit for watcher to be ready
	time.Sleep(100 * time.Millisecond)

	// Modify the config file
	err = os.WriteFile(configPath, []byte(`
server:
  host: "127.0.0.1"
  port: 9090
`), 0644)
	require.NoError(t, err)

	// Wait for error handling
	time.Sleep(1 * time.Second)

	// Watcher should still be running despite callback error
	assert.True(t, watcher.IsRunning())
}

func TestConfigWatcher_FileRemovalAndRecreation(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.yaml")
	
	// Write initial config
	err := os.WriteFile(configPath, []byte(`
server:
  host: "localhost"
  port: 8080
`), 0644)
	require.NoError(t, err)

	// Create watcher
	callbackCalled := false
	var callbackMutex sync.Mutex
	callback := func(*Config) error {
		callbackMutex.Lock()
		defer callbackMutex.Unlock()
		callbackCalled = true
		return nil
	}

	watcher, err := NewConfigWatcher(configPath, callback)
	require.NoError(t, err)

	// Start watcher
	err = watcher.Start()
	require.NoError(t, err)
	defer watcher.Stop()

	// Wait a bit for watcher to be ready
	time.Sleep(100 * time.Millisecond)

	// Remove the config file
	err = os.Remove(configPath)
	require.NoError(t, err)

	// Wait a bit
	time.Sleep(100 * time.Millisecond)

	// Recreate the config file
	err = os.WriteFile(configPath, []byte(`
server:
  host: "127.0.0.1"
  port: 9090
`), 0644)
	require.NoError(t, err)

	// Wait for callback to be called
	time.Sleep(1 * time.Second)

	callbackMutex.Lock()
	assert.True(t, callbackCalled, "Callback should have been called after file recreation")
	callbackMutex.Unlock()
}

func TestConfigWatcher_ContextCancellation(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.yaml")
	
	// Write initial config
	err := os.WriteFile(configPath, []byte(`
server:
  host: "localhost"
  port: 8080
`), 0644)
	require.NoError(t, err)

	watcher, err := NewConfigWatcher(configPath, func(*Config) error { return nil })
	require.NoError(t, err)

	// Start watcher
	err = watcher.Start()
	require.NoError(t, err)

	// Cancel the context
	watcher.cancel()

	// Wait for goroutine to stop
	time.Sleep(100 * time.Millisecond)

	// Watcher should be stopped
	assert.False(t, watcher.IsRunning())
}

func TestConfigWatcher_GetWatcher(t *testing.T) {
	watcher, err := NewConfigWatcher("test-config.yaml", func(*Config) error { return nil })
	require.NoError(t, err)

	// Test GetWatcher method
	fsWatcher := watcher.GetWatcher()
	assert.NotNil(t, fsWatcher)
}

func TestConfigWatcher_ConcurrentAccess(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.yaml")
	
	// Write initial config
	err := os.WriteFile(configPath, []byte(`
server:
  host: "localhost"
  port: 8080
`), 0644)
	require.NoError(t, err)

	watcher, err := NewConfigWatcher(configPath, func(*Config) error { return nil })
	require.NoError(t, err)

	// Test concurrent access to IsRunning
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = watcher.IsRunning()
		}()
	}
	wg.Wait()

	// Start watcher
	err = watcher.Start()
	require.NoError(t, err)

	// Test concurrent access while running
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = watcher.IsRunning()
		}()
	}
	wg.Wait()

	// Stop watcher
	err = watcher.Stop()
	require.NoError(t, err)
}

// REQ-CONFIG-005: Hot reload capability - Real file testing
func TestConfigWatcher_RealConfigFiles(t *testing.T) {
	// Test with REAL default.yaml file
	defaultConfigPath := "../../config/default.yaml"
	
	// Verify file exists
	_, err := os.Stat(defaultConfigPath)
	require.NoError(t, err, "Real default.yaml file must exist for testing")
	
	// Create watcher with real file
	callbackCalled := false
	var callbackMutex sync.Mutex
	callback := func(*Config) error {
		callbackMutex.Lock()
		defer callbackMutex.Unlock()
		callbackCalled = true
		return nil
	}

	watcher, err := NewConfigWatcher(defaultConfigPath, callback)
	require.NoError(t, err)
	require.NotNil(t, watcher)

	// Test Start with real file
	err = watcher.Start()
	require.NoError(t, err)
	assert.True(t, watcher.IsRunning())

	// Test Stop
	err = watcher.Stop()
	require.NoError(t, err)
	assert.False(t, watcher.IsRunning())
}

// REQ-CONFIG-008: Error handling during reload - Malformed YAML testing
func TestConfigWatcher_MalformedYAMLHandling(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.yaml")
	
	// Write initial valid config
	err := os.WriteFile(configPath, []byte(`
server:
  host: "localhost"
  port: 8080
`), 0644)
	require.NoError(t, err)

	// Create watcher with error tracking
	reloadErrors := make([]error, 0)
	var errorMutex sync.Mutex
	callback := func(*Config) error {
		errorMutex.Lock()
		defer errorMutex.Unlock()
		if len(reloadErrors) > 0 {
			return reloadErrors[len(reloadErrors)-1]
		}
		return nil
	}

	watcher, err := NewConfigWatcher(configPath, callback)
	require.NoError(t, err)

	// Start watcher
	err = watcher.Start()
	require.NoError(t, err)
	defer watcher.Stop()

	// Wait for watcher to be ready
	time.Sleep(100 * time.Millisecond)

	// Write malformed YAML
	err = os.WriteFile(configPath, []byte(`
server:
  host: "localhost"
  port: invalid_port
  - invalid: yaml: structure
`), 0644)
	require.NoError(t, err)

	// Wait for error handling
	time.Sleep(1 * time.Second)

	// Watcher should still be running despite YAML error
	assert.True(t, watcher.IsRunning())
}

// REQ-CONFIG-007: Debouncing of rapid changes - Performance validation
func TestConfigWatcher_PerformanceValidation(t *testing.T) {
	// Create a temporary config file
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.yaml")
	
	// Write initial config
	err := os.WriteFile(configPath, []byte(`
server:
  host: "localhost"
  port: 8080
`), 0644)
	require.NoError(t, err)

	// Create watcher with performance tracking
	reloadTimes := make([]time.Time, 0)
	var timeMutex sync.Mutex
	callback := func(*Config) error {
		timeMutex.Lock()
		defer timeMutex.Unlock()
		reloadTimes = append(reloadTimes, time.Now())
		return nil
	}

	watcher, err := NewConfigWatcher(configPath, callback)
	require.NoError(t, err)

	// Start watcher
	err = watcher.Start()
	require.NoError(t, err)
	defer watcher.Stop()

	// Wait for watcher to be ready
	time.Sleep(100 * time.Millisecond)

	// Rapidly modify the config file
	startTime := time.Now()
	for i := 0; i < 10; i++ {
		err = os.WriteFile(configPath, []byte(fmt.Sprintf(`
server:
  host: "localhost"
  port: %d
`, 8080+i)), 0644)
		require.NoError(t, err)
		time.Sleep(50 * time.Millisecond) // Less than debounce interval
	}

	// Wait for processing
	time.Sleep(2 * time.Second)

	// Validate performance: should not have more than 3 reloads (debounced)
	timeMutex.Lock()
	reloadCount := len(reloadTimes)
	timeMutex.Unlock()
	
	assert.LessOrEqual(t, reloadCount, 3, "Should have debounced rapid changes")
	
	// Validate timing: total time should be reasonable
	totalTime := time.Since(startTime)
	assert.Less(t, totalTime, 5*time.Second, "Total processing time should be reasonable")
}

func TestConfigWatcher_EdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		configPath    string
		callback      func(*Config) error
		expectSuccess bool
	}{
		{
			name:          "very long path",
			configPath:    string(make([]byte, 1000)), // Very long path
			callback:      func(*Config) error { return nil },
			expectSuccess: true, // fsnotify should handle this
		},
		{
			name:          "path with special characters",
			configPath:    "test-config-@#$%.yaml",
			callback:      func(*Config) error { return nil },
			expectSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			watcher, err := NewConfigWatcher(tt.configPath, tt.callback)
			if tt.expectSuccess {
				require.NoError(t, err)
				require.NotNil(t, watcher)
			} else {
				require.Error(t, err)
			}
		})
	}
}

// Benchmark tests for performance validation
func BenchmarkConfigWatcher_Start(b *testing.B) {
	// Create a temporary config file
	tempDir := b.TempDir()
	configPath := filepath.Join(tempDir, "bench-config.yaml")
	
	err := os.WriteFile(configPath, []byte(`
server:
  host: "localhost"
  port: 8080
`), 0644)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		watcher, err := NewConfigWatcher(configPath, func(*Config) error { return nil })
		require.NoError(b, err)
		
		err = watcher.Start()
		require.NoError(b, err)
		
		err = watcher.Stop()
		require.NoError(b, err)
	}
}

func BenchmarkConfigWatcher_FileChangeDetection(b *testing.B) {
	// Create a temporary config file
	tempDir := b.TempDir()
	configPath := filepath.Join(tempDir, "bench-config.yaml")
	
	err := os.WriteFile(configPath, []byte(`
server:
  host: "localhost"
  port: 8080
`), 0644)
	require.NoError(b, err)

	watcher, err := NewConfigWatcher(configPath, func(*Config) error { return nil })
	require.NoError(b, err)

	err = watcher.Start()
	require.NoError(b, err)
	defer watcher.Stop()

	// Wait for watcher to be ready
	time.Sleep(100 * time.Millisecond)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err = os.WriteFile(configPath, []byte(fmt.Sprintf(`
server:
  host: "localhost"
  port: %d
`, i)), 0644)
		require.NoError(b, err)
		
		// Small delay to allow processing
		time.Sleep(10 * time.Millisecond)
	}
}
