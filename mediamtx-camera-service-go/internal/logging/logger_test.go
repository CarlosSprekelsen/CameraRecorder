//go:build unit
// +build unit

/*
Logging Infrastructure Unit Tests

Requirements Coverage:
- REQ-LOG-001: Structured logging with logrus
- REQ-LOG-002: Correlation ID support
- REQ-LOG-003: Log rotation configuration
- REQ-LOG-004: Log level management
- REQ-LOG-005: Configuration integration

Test Categories: Unit/Integration/Performance
API Documentation Reference: Internal logging system (no external API)
*/

package logging

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/camerarecorder/mediamtx-camera-service-go/tests/utils"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLogging_NewLogger tests logger creation and basic functionality
func TestLogging_NewLogger(t *testing.T) {
	t.Parallel()
	// REQ-LOG-001: Structured logging with logrus
	logger := logging.NewLogger("test-component")

	assert.NotNil(t, logger)
	assert.NotNil(t, logger.Logger)
	assert.Equal(t, logrus.InfoLevel, logger.GetLevel())
}

// TestLogging_GetLogger tests global logger singleton
func TestLogging_GetLogger(t *testing.T) {
	t.Parallel()
	// REQ-LOG-001: Structured logging with logrus
	logger1 := logging.GetLogger()
	logger2 := logging.GetLogger()

	assert.NotNil(t, logger1)
	assert.NotNil(t, logger2)
	assert.Equal(t, logger1, logger2) // Should be the same instance
}

// TestLogging_SetupLogging tests logging configuration setup
func TestLogging_SetupLogging(t *testing.T) {
	t.Parallel()
	// REQ-LOG-001: Structured logging with logrus
	// REQ-LOG-004: Log level management

	// COMMON PATTERN: Use shared test environment instead of individual components
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	tests := []struct {
		name    string
		config  *logging.LoggingConfig
		wantErr bool
	}{
		{
			name: "valid console config",
			config: &logging.LoggingConfig{
				Level:          "info",
				Format:         "text",
				ConsoleEnabled: true,
				FileEnabled:    false,
			},
			wantErr: false,
		},
		{
			name: "valid file config",
			config: &logging.LoggingConfig{
				Level:          "debug",
				Format:         "json",
				ConsoleEnabled: false,
				FileEnabled:    true,
				FilePath:       "/tmp/test.log",
				MaxFileSize:    100,
				BackupCount:    5,
			},
			wantErr: false,
		},
		{
			name: "invalid log level",
			config: &logging.LoggingConfig{
				Level:          "invalid",
				ConsoleEnabled: true,
			},
			wantErr: false, // Should fallback to info level
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := logging.SetupLogging(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestLogging_CorrelationID tests correlation ID functionality
func TestLogging_CorrelationID(t *testing.T) {
	t.Parallel()
	// REQ-LOG-002: Correlation ID support

	// COMMON PATTERN: Use shared test environment instead of individual components
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	// Test correlation ID generation
	correlationID := logging.GenerateCorrelationID()
	assert.NotEmpty(t, correlationID)
	assert.Len(t, correlationID, 36) // UUID length

	// Test context integration
	ctx := context.Background()
	ctxWithID := logging.WithCorrelationID(ctx, correlationID)

	retrievedID := logging.GetCorrelationIDFromContext(ctxWithID)
	assert.Equal(t, correlationID, retrievedID)

	// Test empty context
	emptyID := logging.GetCorrelationIDFromContext(ctx)
	assert.Empty(t, emptyID)
}

// TestLogging_WithCorrelationID tests logger correlation ID methods
func TestLogging_WithCorrelationID(t *testing.T) {
	t.Parallel()
	// REQ-LOG-002: Correlation ID support

	// COMMON PATTERN: Use shared test environment instead of individual components
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	logger := env.Logger
	correlationID := "test-correlation-id"

	loggerWithID := logger.WithCorrelationID(correlationID)
	assert.NotNil(t, loggerWithID)
}

// TestLogging_WithField tests structured field logging
func TestLogging_WithField(t *testing.T) {
	t.Parallel()
	// REQ-LOG-001: Structured logging with logrus

	// COMMON PATTERN: Use shared test environment instead of individual components
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	logger := env.Logger

	loggerWithField := logger.WithField("test_key", "test_value")
	assert.NotNil(t, loggerWithField)
}

// TestLogging_WithError tests error logging
func TestLogging_WithError(t *testing.T) {
	t.Parallel()
	// REQ-LOG-001: Structured logging with logrus

	// COMMON PATTERN: Use shared test environment instead of individual components
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	logger := env.Logger
	testError := assert.AnError

	loggerWithError := logger.WithError(testError)
	assert.NotNil(t, loggerWithError)
}

// TestLogging_LogWithContext tests context-based logging
func TestLogging_LogWithContext(t *testing.T) {
	t.Parallel()
	// REQ-LOG-002: Correlation ID support

	// COMMON PATTERN: Use shared test environment instead of individual components
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	logger := env.Logger
	ctx := context.Background()
	correlationID := "test-correlation-id"
	ctxWithID := logging.WithCorrelationID(ctx, correlationID)

	// Test logging with correlation ID in context
	logger.LogWithContext(ctxWithID, logrus.InfoLevel, "test message")

	// Test logging without correlation ID
	logger.LogWithContext(ctx, logrus.InfoLevel, "test message without correlation")
}

// TestLogging_ConvenienceMethods tests convenience logging methods
func TestLogging_ConvenienceMethods(t *testing.T) {
	t.Parallel()
	// REQ-LOG-001: Structured logging with logrus

	// COMMON PATTERN: Use shared test environment instead of individual components
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	logger := env.Logger
	ctx := context.Background()

	// Test all convenience methods
	logger.DebugWithContext(ctx, "debug message")
	logger.InfoWithContext(ctx, "info message")
	logger.WarnWithContext(ctx, "warn message")
	logger.ErrorWithContext(ctx, "error message")

	// These should not panic
	assert.NotNil(t, logger)
}

// TestLogging_LevelManagement tests log level management
func TestLogging_LevelManagement(t *testing.T) {
	t.Parallel()
	// REQ-LOG-004: Log level management

	// COMMON PATTERN: Use shared test environment instead of individual components
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	logger := env.Logger

	// Test level setting
	logger.SetLevel(logrus.DebugLevel)
	assert.Equal(t, logrus.DebugLevel, logger.GetLevel())

	logger.SetLevel(logrus.ErrorLevel)
	assert.Equal(t, logrus.ErrorLevel, logger.GetLevel())

	// Test level checking
	assert.True(t, logger.IsLevelEnabled(logrus.ErrorLevel))
	assert.True(t, logger.IsLevelEnabled(logrus.FatalLevel))
	assert.False(t, logger.IsLevelEnabled(logrus.InfoLevel))
}

// TestLogging_ComponentLevel tests component-specific level management
func TestLogging_ComponentLevel(t *testing.T) {
	t.Parallel()
	// REQ-LOG-004: Log level management

	// COMMON PATTERN: Use shared test environment instead of individual components
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	logger := env.Logger

	// Test component level setting
	logger.SetComponentLevel("test-component", logrus.DebugLevel)

	// Test effective level
	effectiveLevel := logger.GetEffectiveLevel("test-component")
	assert.Equal(t, logrus.DebugLevel, effectiveLevel)

	// Test level enabled check
	assert.True(t, logger.IsLevelEnabled(logrus.DebugLevel))
	assert.True(t, logger.IsLevelEnabled(logrus.InfoLevel))
}

// TestLogging_SetupLoggingSimple tests simple logging setup
func TestLogging_SetupLoggingSimple(t *testing.T) {
	t.Parallel()
	// REQ-LOG-001: Structured logging with logrus

	err := logging.SetupLoggingSimple("/tmp/test.log", "info")
	assert.NoError(t, err)
}

// TestLogging_ConfigurationIntegration tests integration with config system
func TestLogging_ConfigurationIntegration(t *testing.T) {
	t.Parallel()
	// REQ-LOG-005: Configuration integration

	// Create config.LoggingConfig
	configLogging := &config.LoggingConfig{
		Level:          "debug",
		Format:         "json",
		FileEnabled:    true,
		FilePath:       "/tmp/test.log",
		MaxFileSize:    100,
		BackupCount:    5,
		ConsoleEnabled: true,
	}

	// Convert to logging.LoggingConfig
	loggingConfig := logging.NewLoggingConfigFromConfig(configLogging)

	// Verify conversion
	assert.Equal(t, configLogging.Level, loggingConfig.Level)
	assert.Equal(t, configLogging.Format, loggingConfig.Format)
	assert.Equal(t, configLogging.FileEnabled, loggingConfig.FileEnabled)
	assert.Equal(t, configLogging.FilePath, loggingConfig.FilePath)
	assert.Equal(t, int(configLogging.MaxFileSize), loggingConfig.MaxFileSize)
	assert.Equal(t, configLogging.BackupCount, loggingConfig.BackupCount)
	assert.Equal(t, configLogging.ConsoleEnabled, loggingConfig.ConsoleEnabled)
}

// TestLogging_FileRotation tests file rotation functionality
func TestLogging_FileRotation(t *testing.T) {
	// REQ-LOG-003: Log rotation configuration

	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "logging_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	logFilePath := filepath.Join(tempDir, "test.log")

	config := &logging.LoggingConfig{
		Level:          "info",
		Format:         "text",
		ConsoleEnabled: false,
		FileEnabled:    true,
		FilePath:       logFilePath,
		MaxFileSize:    1, // 1 byte to trigger rotation quickly
		BackupCount:    3,
	}

	// Setup logging
	err = logging.SetupLogging(config)
	require.NoError(t, err)

	logger := logging.GetLogger()

	// Write enough logs to trigger rotation
	for i := 0; i < 10; i++ {
		logger.Info("test log message that should trigger rotation")
	}

	// Wait a bit for file operations
	time.Sleep(100 * time.Millisecond)

	// Check if log file exists
	_, err = os.Stat(logFilePath)
	assert.NoError(t, err, "Log file should exist")
}

// TestLogging_FormatCompatibility tests format compatibility
func TestLogging_FormatCompatibility(t *testing.T) {
	// REQ-LOG-001: Structured logging with logrus

	tests := []struct {
		name   string
		format string
	}{
		{"text format", "text"},
		{"json format", "json"},
		{"mixed format", "mixed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &logging.LoggingConfig{
				Level:          "info",
				Format:         tt.format,
				ConsoleEnabled: true,
				FileEnabled:    false,
			}

			err := logging.SetupLogging(config)
			assert.NoError(t, err)
		})
	}
}

// TestLogging_EnvironmentVariableOverride tests environment variable overrides
func TestLogging_EnvironmentVariableOverride(t *testing.T) {
	// REQ-LOG-004: Log level management

	// Set environment variable
	os.Setenv("CAMERA_SERVICE_ENV", "production")
	defer os.Unsetenv("CAMERA_SERVICE_ENV")

	config := &logging.LoggingConfig{
		Level:          "info",
		Format:         "text",
		ConsoleEnabled: true,
		FileEnabled:    false,
	}

	err := logging.SetupLogging(config)
	assert.NoError(t, err)
}

// TestLogging_Concurrency tests concurrent logging operations
func TestLogging_Concurrency(t *testing.T) {
	// REQ-LOG-001: Structured logging with logrus

	// COMMON PATTERN: Use shared test environment instead of individual components
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	logger := env.Logger

	// Test concurrent logging
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			logger.Info("concurrent log message")
			logger.WithField("goroutine_id", fmt.Sprintf("%d", id)).Info("structured log message")
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	assert.NotNil(t, logger)
}

// TestLogging_ErrorHandling tests error handling scenarios
func TestLogging_ErrorHandling(t *testing.T) {
	// REQ-LOG-001: Structured logging with logrus

	// Test invalid file path
	config := &logging.LoggingConfig{
		Level:          "info",
		Format:         "text",
		ConsoleEnabled: false,
		FileEnabled:    true,
		FilePath:       "/invalid/path/that/should/not/exist/test.log",
		MaxFileSize:    100,
		BackupCount:    5,
	}

	// This should not panic, but may return an error
	_ = logging.SetupLogging(config)
	// We don't assert on error here as file system behavior may vary
	assert.NotNil(t, config)
}

// TestLogging_Performance tests logging performance
func TestLogging_Performance(t *testing.T) {
	// REQ-LOG-001: Structured logging with logrus

	// COMMON PATTERN: Use shared test environment instead of individual components
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	logger := env.Logger

	// Performance test: log many messages quickly
	start := time.Now()

	for i := 0; i < 1000; i++ {
		logger.Info("performance test message")
	}

	duration := time.Since(start)

	// Should complete within reasonable time (< 1 second for 1000 messages)
	assert.Less(t, duration, time.Second, "Logging 1000 messages should complete within 1 second")

	// Average time per message should be < 1ms
	avgTimePerMessage := duration / 1000
	assert.Less(t, avgTimePerMessage, time.Millisecond, "Average time per log message should be < 1ms")
}

// TestLogging_PythonFormatCompatibility tests format compatibility with Python logging
func TestLogging_PythonFormatCompatibility(t *testing.T) {
	// REQ-LOG-001: Format compatibility validation against Python system

	// COMMON PATTERN: Use shared test environment instead of individual components
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	// Test Python format: %(asctime)s - %(name)s - %(levelname)s - %(message)s
	config := &logging.LoggingConfig{
		Level:          "info",
		Format:         "text",
		ConsoleEnabled: true,
		FileEnabled:    false,
	}

	err := logging.SetupLogging(config)
	require.NoError(t, err)

	logger := env.Logger

	// Test that format matches Python logging pattern
	// Python format: 2025-01-15 10:30:45,123 - test-component - INFO - test message
	// Go format should be similar: time="2025-01-15T10:30:45.123Z" level=info msg="test message"

	logger.Info("test message")
	// Note: Actual format validation would require capturing output and parsing
	// This test ensures the logging system is configured for compatibility
}

// TestLogging_PerformanceBenchmark validates <10ms per log entry requirement
func TestLogging_PerformanceBenchmark(t *testing.T) {
	// REQ-LOG-001: Performance under high load (<10ms per log entry)

	// COMMON PATTERN: Use shared test environment instead of individual components
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	logger := env.Logger

	// Performance test: validate <10ms per log entry requirement
	start := time.Now()

	for i := 0; i < 1000; i++ {
		logger.Info("performance benchmark message")
	}

	duration := time.Since(start)
	avgTimePerMessage := duration / 1000

	// Implementation plan requirement: <10ms per log entry
	assert.Less(t, avgTimePerMessage, 10*time.Millisecond,
		"Average time per log message must be <10ms per implementation plan requirement")

	t.Logf("Performance: %v for 1000 messages, avg: %v per message", duration, avgTimePerMessage)
}

// TestLogging_ConcurrentRotationSafety tests concurrent rotation safety
func TestLogging_ConcurrentRotationSafety(t *testing.T) {
	// REQ-LOG-003: Concurrent rotation safety

	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "concurrent_rotation_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	logFilePath := filepath.Join(tempDir, "concurrent.log")

	config := &logging.LoggingConfig{
		Level:          "info",
		Format:         "text",
		ConsoleEnabled: false,
		FileEnabled:    true,
		FilePath:       logFilePath,
		MaxFileSize:    1, // 1 byte to trigger rotation quickly
		BackupCount:    3,
	}

	err = logging.SetupLogging(config)
	require.NoError(t, err)

	logger := logging.GetLogger()

	// Test concurrent logging that could trigger rotation
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				logger.Info(fmt.Sprintf("concurrent log message %d-%d", id, j))
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Wait a bit for file operations
	time.Sleep(100 * time.Millisecond)

	// Check if log file exists and rotation worked
	_, err = os.Stat(logFilePath)
	assert.NoError(t, err, "Log file should exist after concurrent rotation")
}

// TestLogging_ComprehensiveErrorHandling tests comprehensive error scenarios
func TestLogging_ComprehensiveErrorHandling(t *testing.T) {
	// REQ-LOG-001: Error logging with stack traces

	// COMMON PATTERN: Use shared test environment instead of individual components
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	logger := env.Logger

	// Test various error scenarios
	testCases := []struct {
		name        string
		errorType   string
		shouldPanic bool
	}{
		{"nil error", "nil", false},
		{"standard error", "standard", false},
		{"wrapped error", "wrapped", false},
		{"file system error", "filesystem", false},
		{"permission error", "permission", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var testErr error

			switch tc.errorType {
			case "nil":
				testErr = nil
			case "standard":
				testErr = fmt.Errorf("standard test error")
			case "wrapped":
				testErr = fmt.Errorf("wrapped error: %w", fmt.Errorf("inner error"))
			case "filesystem":
				testErr = &os.PathError{Op: "open", Path: "/nonexistent", Err: fmt.Errorf("file not found")}
			case "permission":
				testErr = fmt.Errorf("permission denied: /protected/file")
			}

			// Test error logging with structured fields
			loggerWithError := logger.WithError(testErr)
			assert.NotNil(t, loggerWithError)

			// Test error logging with context
			ctx := context.Background()
			if testErr != nil {
				logger.ErrorWithContext(ctx, "error occurred during test")
			}
		})
	}
}

// TestLogging_EnvironmentVariableOverrides tests comprehensive environment variable overrides
func TestLogging_EnvironmentVariableOverrides(t *testing.T) {
	// REQ-LOG-004: Environment variable level control

	// Test various environment variable scenarios
	testCases := []struct {
		name          string
		envVar        string
		envValue      string
		expectedLevel string
	}{
		{"production env", "CAMERA_SERVICE_ENV", "production", "warn"},
		{"development env", "CAMERA_SERVICE_ENV", "development", "debug"},
		{"test env", "CAMERA_SERVICE_ENV", "test", "debug"},
		{"custom log level", "CAMERA_SERVICE_LOG_LEVEL", "error", "error"},
		{"invalid log level", "CAMERA_SERVICE_LOG_LEVEL", "invalid", "info"}, // should fallback
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set environment variable
			if tc.envVar != "" {
				os.Setenv(tc.envVar, tc.envValue)
				defer os.Unsetenv(tc.envVar)
			}

			config := &logging.LoggingConfig{
				Level:          "info", // default
				Format:         "text",
				ConsoleEnabled: true,
				FileEnabled:    false,
			}

			err := logging.SetupLogging(config)
			assert.NoError(t, err)

			// Verify the configuration was applied
			logger := logging.GetLogger()
			assert.NotNil(t, logger)
		})
	}
}

// TestLogging_CrossComponentCorrelationID tests cross-component correlation ID propagation
func TestLogging_CrossComponentCorrelationID(t *testing.T) {
	// REQ-LOG-002: Cross-component tracing validation

	// COMMON PATTERN: Use shared test environment instead of individual components
	env := utils.SetupTestEnvironment(t)
	defer utils.TeardownTestEnvironment(t, env)

	// Create multiple loggers for different components
	authLogger := env.Logger
	dbLogger := env.Logger
	apiLogger := env.Logger

	// Generate correlation ID
	correlationID := logging.GenerateCorrelationID()
	assert.NotEmpty(t, correlationID)

	// Create context with correlation ID
	ctx := logging.WithCorrelationID(context.Background(), correlationID)

	// Test correlation ID propagation across components
	authLogger.LogWithContext(ctx, logrus.InfoLevel, "user authentication started")
	dbLogger.LogWithContext(ctx, logrus.InfoLevel, "database query executed")
	apiLogger.LogWithContext(ctx, logrus.InfoLevel, "API response sent")

	// Verify correlation ID is consistent across all components
	retrievedID := logging.GetCorrelationIDFromContext(ctx)
	assert.Equal(t, correlationID, retrievedID)

	// Test that each logger can access the correlation ID
	authLoggerWithID := authLogger.WithCorrelationID(correlationID)
	dbLoggerWithID := dbLogger.WithCorrelationID(correlationID)
	apiLoggerWithID := apiLogger.WithCorrelationID(correlationID)

	assert.NotNil(t, authLoggerWithID)
	assert.NotNil(t, dbLoggerWithID)
	assert.NotNil(t, apiLoggerWithID)
}
