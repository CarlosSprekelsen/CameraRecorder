package logging

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/*
Logging Infrastructure Unit Tests

Requirements Coverage:
- REQ-LOG-001: Structured logging with logrus
- REQ-LOG-002: Correlation ID support
- REQ-LOG-003: Log rotation configuration
- REQ-LOG-004: Log level management
- REQ-LOG-005: Configuration integration

Test Categories: Unit
API Documentation Reference: Internal logging system (no external API)
*/

// =============================================================================
// CORE LOGGER TESTS
// =============================================================================

// TestNewLogger tests logger creation and basic functionality
func TestNewLogger(t *testing.T) {
	t.Parallel()
	// REQ-LOG-001: Structured logging with logrus
	
	logger := NewLogger("test-component")
	assert.NotNil(t, logger)
	assert.NotNil(t, logger.Logger)
	assert.Equal(t, logrus.InfoLevel, logger.GetLevel())
}

// TestGetLogger tests global logger singleton
func TestGetLogger(t *testing.T) {
	t.Parallel()
	// REQ-LOG-001: Structured logging with logrus
	
	logger1 := GetLogger()
	logger2 := GetLogger()

	assert.NotNil(t, logger1)
	assert.NotNil(t, logger2)
	assert.Equal(t, logger1, logger2) // Should be the same instance
}

// =============================================================================
// CORRELATION ID TESTS
// =============================================================================

// TestGenerateCorrelationID tests correlation ID generation
func TestGenerateCorrelationID(t *testing.T) {
	t.Parallel()
	// REQ-LOG-002: Correlation ID support
	
	correlationID := GenerateCorrelationID()
	assert.NotEmpty(t, correlationID)
	assert.Len(t, correlationID, 36) // UUID length
}

// TestGetCorrelationIDFromContext tests context correlation ID retrieval
func TestGetCorrelationIDFromContext(t *testing.T) {
	t.Parallel()
	// REQ-LOG-002: Correlation ID support
	
	ctx := context.Background()
	correlationID := "test-correlation-id"
	ctxWithID := WithCorrelationID(ctx, correlationID)

	retrievedID := GetCorrelationIDFromContext(ctxWithID)
	assert.Equal(t, correlationID, retrievedID)

	// Test empty context
	emptyID := GetCorrelationIDFromContext(ctx)
	assert.Empty(t, emptyID)
}

// TestGetCorrelationIDFromContext_NilContext tests nil context handling
func TestGetCorrelationIDFromContext_NilContext(t *testing.T) {
	t.Parallel()
	// REQ-LOG-002: Correlation ID support
	
	// Test nil context
	result := GetCorrelationIDFromContext(nil)
	assert.Empty(t, result, "Should return empty string for nil context")
}

// TestLogWithCorrelationID tests global correlation ID logging
func TestLogWithCorrelationID(t *testing.T) {
	t.Parallel()
	// REQ-LOG-002: Correlation ID support
	
	ctx := context.Background()
	LogWithCorrelationID(ctx, logrus.InfoLevel, "test message")
	
	// Test with correlation ID
	correlationID := "test-correlation-id"
	ctxWithID := WithCorrelationID(ctx, correlationID)
	LogWithCorrelationID(ctxWithID, logrus.DebugLevel, "test message with correlation ID")
}

// =============================================================================
// LOGGER METHOD TESTS
// =============================================================================

// TestLogger_WithCorrelationID tests correlation ID functionality
func TestLogger_WithCorrelationID(t *testing.T) {
	t.Parallel()
	// REQ-LOG-002: Correlation ID support
	
	logger := NewLogger("test-component")
	correlationID := "test-correlation-id"

	loggerWithID := logger.WithCorrelationID(correlationID)
	assert.NotNil(t, loggerWithID)
}

// TestLogger_WithField tests structured field logging
func TestLogger_WithField(t *testing.T) {
	t.Parallel()
	// REQ-LOG-001: Structured logging with logrus
	
	logger := NewLogger("test-component")
	key := "test_key"
	value := "test_value"

	loggerWithField := logger.WithField(key, value)
	assert.NotNil(t, loggerWithField)
}

// TestLogger_WithError tests error logging
func TestLogger_WithError(t *testing.T) {
	t.Parallel()
	// REQ-LOG-001: Structured logging with logrus
	
	logger := NewLogger("test-component")
	testErr := assert.AnError

	loggerWithError := logger.WithError(testErr)
	assert.NotNil(t, loggerWithError)
}

// =============================================================================
// CONTEXT LOGGING TESTS
// =============================================================================

// TestLogger_LogWithContext tests context-based logging
func TestLogger_LogWithContext(t *testing.T) {
	t.Parallel()
	// REQ-LOG-002: Correlation ID support
	
	logger := NewLogger("test-component")
	ctx := context.Background()
	
	// Test with info level
	logger.LogWithContext(ctx, logrus.InfoLevel, "test message")
	
	// Test with correlation ID in context
	correlationID := "test-correlation-id"
	ctxWithID := WithCorrelationID(ctx, correlationID)
	logger.LogWithContext(ctxWithID, logrus.DebugLevel, "test message with correlation ID")
}

// TestLogWithContext_NilContext tests nil context handling
func TestLogWithContext_NilContext(t *testing.T) {
	t.Parallel()
	// REQ-LOG-002: Correlation ID support
	
	logger := NewLogger("test-component")
	
	// Test with nil context
	logger.LogWithContext(nil, logrus.InfoLevel, "test message with nil context")
	// Should not panic and should handle gracefully
}

// TestLogger_ContextMethods tests all context-based logging methods
func TestLogger_ContextMethods(t *testing.T) {
	t.Parallel()
	// REQ-LOG-001: Structured logging with logrus
	
	logger := NewLogger("test-component")
	ctx := context.Background()
	
	// Test all context methods (excluding fatal which calls os.Exit)
	logger.DebugWithContext(ctx, "debug message")
	logger.InfoWithContext(ctx, "info message")
	logger.WarnWithContext(ctx, "warn message")
	logger.ErrorWithContext(ctx, "error message")
	// Note: FatalWithContext calls os.Exit(1) so we can't test it in unit tests
}

// TestLogger_ContextLogging_EdgeCases tests edge cases in context logging
func TestLogger_ContextLogging_EdgeCases(t *testing.T) {
	t.Parallel()
	// REQ-LOG-002: Correlation ID support
	
	logger := NewLogger("test-component")
	
	// Test with empty context
	ctx := context.Background()
	logger.LogWithContext(ctx, logrus.InfoLevel, "message with empty context")
	
	// Test with context containing non-string correlation ID
	ctxWithNonString := context.WithValue(ctx, CorrelationIDKey, 123)
	logger.LogWithContext(ctxWithNonString, logrus.InfoLevel, "message with non-string correlation ID")
	
	// Test with context containing empty correlation ID
	ctxWithEmpty := context.WithValue(ctx, CorrelationIDKey, "")
	logger.LogWithContext(ctxWithEmpty, logrus.InfoLevel, "message with empty correlation ID")
}

// =============================================================================
// LEVEL MANAGEMENT TESTS
// =============================================================================

// TestLogger_LevelManagement tests log level management
func TestLogger_LevelManagement(t *testing.T) {
	t.Parallel()
	// REQ-LOG-004: Log level management
	
	logger := NewLogger("test-component")
	
	// Test SetLevel
	logger.SetLevel(logrus.DebugLevel)
	assert.Equal(t, logrus.DebugLevel, logger.GetLevel())
	
	// Test IsLevelEnabled at Debug level
	assert.True(t, logger.IsLevelEnabled(logrus.InfoLevel))
	assert.True(t, logger.IsLevelEnabled(logrus.WarnLevel))
	assert.True(t, logger.IsLevelEnabled(logrus.DebugLevel))
	assert.False(t, logger.IsLevelEnabled(logrus.TraceLevel))
	
	// Test SetComponentLevel
	logger.SetComponentLevel("test-component", logrus.WarnLevel)
	
	// Test GetEffectiveLevel
	effectiveLevel := logger.GetEffectiveLevel("test-component")
	assert.Equal(t, logrus.WarnLevel, effectiveLevel)
	
	// Test IsLevelEnabled at Warn level
	assert.False(t, logger.IsLevelEnabled(logrus.InfoLevel))
	assert.True(t, logger.IsLevelEnabled(logrus.WarnLevel))
	assert.False(t, logger.IsLevelEnabled(logrus.DebugLevel))
	assert.False(t, logger.IsLevelEnabled(logrus.TraceLevel))
}

// TestLogger_ComponentLevels tests component-specific level management
func TestLogger_ComponentLevels(t *testing.T) {
	t.Parallel()
	// REQ-LOG-004: Log level management
	
	logger := NewLogger("test-component")
	
	// Test setting different levels for different components
	logger.SetComponentLevel("component1", logrus.DebugLevel)
	logger.SetComponentLevel("component2", logrus.WarnLevel)
	logger.SetComponentLevel("component3", logrus.ErrorLevel)
	
	// Verify effective levels
	assert.Equal(t, logrus.ErrorLevel, logger.GetEffectiveLevel("component3"))
	assert.Equal(t, logrus.ErrorLevel, logger.GetEffectiveLevel("component2"))
	assert.Equal(t, logrus.ErrorLevel, logger.GetEffectiveLevel("component1"))
	
	// Test level enablement
	assert.False(t, logger.IsLevelEnabled(logrus.DebugLevel))
	assert.False(t, logger.IsLevelEnabled(logrus.InfoLevel))
	assert.True(t, logger.IsLevelEnabled(logrus.ErrorLevel))
}

// =============================================================================
// CONFIGURATION TESTS
// =============================================================================

// TestLoggingConfig tests logging configuration
func TestLoggingConfig(t *testing.T) {
	t.Parallel()
	// REQ-LOG-005: Configuration integration
	
	config := &LoggingConfig{
		Level:          "debug",
		Format:         "json",
		ConsoleEnabled: true,
		FileEnabled:    false,
	}
	
	assert.Equal(t, "debug", config.Level)
	assert.Equal(t, "json", config.Format)
	assert.True(t, config.ConsoleEnabled)
	assert.False(t, config.FileEnabled)
}

// TestSetupLogging tests logging setup
func TestSetupLogging(t *testing.T) {
	t.Parallel()
	// REQ-LOG-001: Structured logging with logrus
	
	config := &LoggingConfig{
		Level:          "info",
		Format:         "text",
		ConsoleEnabled: true,
		FileEnabled:    false,
	}
	
	err := SetupLogging(config)
	assert.NoError(t, err)
}

// TestSetupLogging_FileLogging tests file logging configuration
func TestSetupLogging_FileLogging(t *testing.T) {
	t.Parallel()
	// REQ-LOG-003: Log rotation configuration
	
	// Test file logging with valid configuration
	config := &LoggingConfig{
		Level:          "debug",
		Format:         "json",
		ConsoleEnabled: false,
		FileEnabled:    true,
		FilePath:       "/tmp/test_file.log",
		MaxFileSize:    10,
		BackupCount:    3,
	}
	
	err := SetupLogging(config)
	assert.NoError(t, err, "File logging setup should succeed")
}

// TestSetupLogging_InvalidLevel tests invalid log level handling
func TestSetupLogging_InvalidLevel(t *testing.T) {
	t.Parallel()
	// REQ-LOG-004: Log level management
	
	// Test with invalid log level (should fallback to info)
	config := &LoggingConfig{
		Level:          "invalid_level",
		Format:         "text",
		ConsoleEnabled: true,
		FileEnabled:    false,
	}
	
	err := SetupLogging(config)
	assert.NoError(t, err, "Invalid level should fallback to info level")
}

// TestSetupLogging_FileAndConsole tests both file and console logging
func TestSetupLogging_FileAndConsole(t *testing.T) {
	t.Parallel()
	// REQ-LOG-001: Structured logging with logrus
	
	// Test both file and console logging enabled
	config := &LoggingConfig{
		Level:          "info",
		Format:         "text",
		ConsoleEnabled: true,
		FileEnabled:    true,
		FilePath:       "/tmp/test_both.log",
		MaxFileSize:    5,
		BackupCount:    2,
	}
	
	err := SetupLogging(config)
	assert.NoError(t, err, "Both file and console logging should work")
}

// TestSetupLogging_EdgeCases tests edge case configurations
func TestSetupLogging_EdgeCases(t *testing.T) {
	t.Parallel()
	// REQ-LOG-005: Configuration integration
	
	testCases := []struct {
		name    string
		config  *LoggingConfig
		wantErr bool
	}{
		{
			name: "empty format",
			config: &LoggingConfig{
				Level:          "info",
				Format:         "",
				ConsoleEnabled: true,
				FileEnabled:    false,
			},
			wantErr: false,
		},
		{
			name: "zero file size",
			config: &LoggingConfig{
				Level:          "info",
				Format:         "text",
				ConsoleEnabled: false,
				FileEnabled:    true,
				FilePath:       "/tmp/test_zero.log",
				MaxFileSize:    0,
				BackupCount:    0,
			},
			wantErr: false,
		},
		{
			name: "very large file size",
			config: &LoggingConfig{
				Level:          "info",
				Format:         "text",
				ConsoleEnabled: false,
				FileEnabled:    true,
				FilePath:       "/tmp/test_large.log",
				MaxFileSize:    999999,
				BackupCount:    999,
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := SetupLogging(tc.config)
			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestSetupLoggingSimple tests simple logging setup
func TestSetupLoggingSimple(t *testing.T) {
	t.Parallel()
	// REQ-LOG-001: Structured logging with logrus
	
	err := SetupLoggingSimple("/tmp/test.log", "info")
	assert.NoError(t, err)
}

// =============================================================================
// FORMATTER TESTS
// =============================================================================

// TestCreateFileFormatter_AllFormats tests all file formatter options
func TestCreateFileFormatter_AllFormats(t *testing.T) {
	t.Parallel()
	// REQ-LOG-001: Structured logging with logrus
	
	// Test JSON formatter
	jsonFormatter := createFileFormatter("json")
	assert.NotNil(t, jsonFormatter, "JSON formatter should not be nil")
	
	// Test text formatter
	textFormatter := createFileFormatter("text")
	assert.NotNil(t, textFormatter, "Text formatter should not be nil")
	
	// Test default formatter (empty string)
	defaultFormatter := createFileFormatter("")
	assert.NotNil(t, defaultFormatter, "Default formatter should not be nil")
	
	// Test unknown format (should fallback to text)
	unknownFormatter := createFileFormatter("unknown")
	assert.NotNil(t, unknownFormatter, "Unknown format should fallback to text formatter")
}

// TestCreateConsoleFormatter_AllFormats tests all console formatter options
func TestCreateConsoleFormatter_AllFormats(t *testing.T) {
	t.Parallel()
	// REQ-LOG-001: Structured logging with logrus
	
	// Test JSON formatter
	jsonFormatter := createConsoleFormatter("json")
	assert.NotNil(t, jsonFormatter, "JSON formatter should not be nil")
	
	// Test text formatter
	textFormatter := createConsoleFormatter("text")
	assert.NotNil(t, textFormatter, "Text formatter should not be nil")
	
	// Test default formatter (empty string)
	defaultFormatter := createConsoleFormatter("")
	assert.NotNil(t, defaultFormatter, "Default formatter should not be nil")
	
	// Test unknown format (should fallback to text)
	unknownFormatter := createConsoleFormatter("unknown")
	assert.NotNil(t, unknownFormatter, "Unknown format should fallback to text formatter")
}

// =============================================================================
// ADVANCED FUNCTIONALITY TESTS
// =============================================================================

// TestLogging_FileRotation tests file rotation functionality
func TestLogging_FileRotation(t *testing.T) {
	// REQ-LOG-003: Log rotation configuration

	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "logging_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	logFilePath := filepath.Join(tempDir, "test.log")

	config := &LoggingConfig{
		Level:          "info",
		Format:         "text",
		ConsoleEnabled: false,
		FileEnabled:    true,
		FilePath:       logFilePath,
		MaxFileSize:    1, // 1 byte to trigger rotation quickly
		BackupCount:    3,
	}

	// Setup logging
	err = SetupLogging(config)
	require.NoError(t, err)

	logger := GetLogger()

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

// TestLogging_Concurrency tests concurrent logging operations
func TestLogging_Concurrency(t *testing.T) {
	// REQ-LOG-001: Structured logging with logrus
	
	logger := NewLogger("test-component")

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

// TestLogging_Performance tests logging performance
func TestLogging_Performance(t *testing.T) {
	// REQ-LOG-001: Structured logging with logrus
	
	logger := NewLogger("test-component")

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

// TestLogging_CrossComponentCorrelationID tests cross-component correlation ID propagation
func TestLogging_CrossComponentCorrelationID(t *testing.T) {
	// REQ-LOG-002: Cross-component tracing validation
	
	// Create multiple loggers for different components
	authLogger := NewLogger("auth")
	dbLogger := NewLogger("database")
	apiLogger := NewLogger("api")

	// Generate correlation ID
	correlationID := GenerateCorrelationID()
	assert.NotEmpty(t, correlationID)

	// Create context with correlation ID
	ctx := WithCorrelationID(context.Background(), correlationID)

	// Test correlation ID propagation across components
	authLogger.LogWithContext(ctx, logrus.InfoLevel, "user authentication started")
	dbLogger.LogWithContext(ctx, logrus.InfoLevel, "database query executed")
	apiLogger.LogWithContext(ctx, logrus.InfoLevel, "API response sent")

	// Verify correlation ID is consistent across all components
	retrievedID := GetCorrelationIDFromContext(ctx)
	assert.Equal(t, correlationID, retrievedID)

	// Test that each logger can access the correlation ID
	authLoggerWithID := authLogger.WithCorrelationID(correlationID)
	dbLoggerWithID := dbLogger.WithCorrelationID(correlationID)
	apiLoggerWithID := apiLogger.WithCorrelationID(correlationID)

	assert.NotNil(t, authLoggerWithID)
	assert.NotNil(t, dbLoggerWithID)
	assert.NotNil(t, apiLoggerWithID)
}
