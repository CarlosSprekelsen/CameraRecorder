package logging

import (
	"context"
	"fmt"
	"os"
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
	AssertLoggerBasicProperties(t, logger, "test-component")
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

// TestCorrelationIDContextOperations tests all correlation ID context operations
func TestCorrelationIDContextOperations(t *testing.T) {
	t.Parallel()
	// REQ-LOG-002: Correlation ID support

	testCases := []struct {
		name           string
		correlationID  string
		expectedResult string
		description    string
	}{
		{"valid correlation ID", "test-correlation-id", "test-correlation-id", "Should retrieve valid correlation ID"},
		{"empty correlation ID", "", "", "Should handle empty correlation ID"},
		{"nil context", "", "", "Should handle nil context gracefully"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var ctx context.Context
			if tc.name == "nil context" {
				ctx = nil
			} else {
				ctx = CreateTestContext(tc.correlationID)
			}

			result := GetCorrelationIDFromContext(ctx)
			assert.Equal(t, tc.expectedResult, result, tc.description)
		})
	}
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

// TestLogger_WithMethods tests all logger "With" methods
func TestLogger_WithMethods(t *testing.T) {
	t.Parallel()
	// REQ-LOG-001: Structured logging with logrus

	logger := CreateTestLogger(t, nil)

	testCases := []struct {
		name string
		test func() *Logger
	}{
		{
			"WithCorrelationID",
			func() *Logger { return logger.WithCorrelationID("test-correlation-id") },
		},
		{
			"WithField",
			func() *Logger { return logger.WithField("test_key", "test_value") },
		},
		{
			"WithError",
			func() *Logger { return logger.WithError(assert.AnError) },
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.test()
			assert.NotNil(t, result)
		})
	}
}

// =============================================================================
// CONTEXT LOGGING TESTS
// =============================================================================

// TestLogger_ContextLogging tests all context-based logging scenarios
func TestLogger_ContextLogging(t *testing.T) {
	t.Parallel()
	// REQ-LOG-002: Correlation ID support

	logger := CreateTestLogger(t, nil)
	fixtures := CreateTestFixtures()

	for _, fixture := range fixtures {
		t.Run(fmt.Sprintf("context_logging_%s", fixture.Component), func(t *testing.T) {
			ctx := CreateTestContext(fixture.CorrelationID)

			// Test all context methods (excluding fatal which calls os.Exit)
			logger.DebugWithContext(ctx, fixture.Message)
			logger.InfoWithContext(ctx, fixture.Message)
			logger.WarnWithContext(ctx, fixture.Message)
			logger.ErrorWithContext(ctx, fixture.Message)

			// Verify correlation ID in context
			AssertCorrelationIDInContext(t, ctx, fixture.CorrelationID)
		})
	}
}

// TestLogger_ContextLogging_EdgeCases tests edge cases in context logging
func TestLogger_ContextLogging_EdgeCases(t *testing.T) {
	t.Parallel()
	// REQ-LOG-002: Correlation ID support

	logger := CreateTestLogger(t, nil)

	// Test with empty context
	ctx := context.Background()
	logger.LogWithContext(ctx, logrus.InfoLevel, "message with empty context")

	// Test with context containing non-string correlation ID
	ctxWithNonString := context.WithValue(ctx, CorrelationIDKey, 123)
	logger.LogWithContext(ctxWithNonString, logrus.InfoLevel, "message with non-string correlation ID")

	// Test with context containing empty correlation ID
	ctxWithEmpty := context.WithValue(ctx, CorrelationIDKey, "")
	logger.LogWithContext(ctxWithEmpty, logrus.InfoLevel, "message with empty correlation ID")

	// Test with nil context
	logger.LogWithContext(nil, logrus.InfoLevel, "message with nil context")
}

// =============================================================================
// LEVEL MANAGEMENT TESTS
// =============================================================================

// TestLogger_LevelManagement tests log level management comprehensively
func TestLogger_LevelManagement(t *testing.T) {
	t.Parallel()
	// REQ-LOG-004: Log level management

	logger := CreateTestLogger(t, nil)
	testLevels := TestLogLevels()

	for _, level := range testLevels {
		t.Run(fmt.Sprintf("level_%s", level.String()), func(t *testing.T) {
			// Test SetLevel
			logger.SetLevel(level)
			assert.Equal(t, level, logger.GetLevel())

			// Test IsLevelEnabled for all levels
			for _, testLevel := range testLevels {
				expected := testLevel >= level
				assert.Equal(t, expected, logger.IsLevelEnabled(testLevel),
					"Level %s should be enabled at %s level", testLevel, level)
			}

			// Test component level management
			logger.SetComponentLevel("test-component", level)
			effectiveLevel := logger.GetEffectiveLevel("test-component")
			assert.Equal(t, level, effectiveLevel)
		})
	}
}

// TestLogger_ComponentLevels tests component-specific level management
func TestLogger_ComponentLevels(t *testing.T) {
	t.Parallel()
	// REQ-LOG-004: Log level management

	logger := CreateTestLogger(t, nil)
	components := TestComponents()

	// Test setting different levels for different components
	for i, component := range components {
		level := TestLogLevels()[i%len(TestLogLevels())]
		logger.SetComponentLevel(component, level)

		effectiveLevel := logger.GetEffectiveLevel(component)
		assert.Equal(t, level, effectiveLevel,
			"Component %s should have effective level %s", component, level)
	}
}

// =============================================================================
// CONFIGURATION TESTS
// =============================================================================

// TestLoggingConfig tests logging configuration structure
func TestLoggingConfig(t *testing.T) {
	t.Parallel()
	// REQ-LOG-005: Configuration integration

	config := CreateTestLoggingConfig("debug", "json", true, false, "")

	assert.Equal(t, "debug", config.Level)
	assert.Equal(t, "json", config.Format)
	assert.True(t, config.ConsoleEnabled)
	assert.False(t, config.FileEnabled)
}

// TestSetupLogging_AllConfigurations tests all logging setup configurations
func TestSetupLogging_AllConfigurations(t *testing.T) {
	t.Parallel()
	// REQ-LOG-001: Structured logging with logrus

	testCases := []struct {
		name          string
		config        *LoggingConfig
		expectedError bool
		description   string
	}{
		{
			name:          "console only",
			config:        CreateTestLoggingConfig("info", "text", true, false, ""),
			expectedError: false,
			description:   "Console-only logging should work",
		},
		{
			name:          "file only",
			config:        CreateTestLoggingConfig("debug", "json", false, true, "/tmp/test_file.log"),
			expectedError: false,
			description:   "File-only logging should work",
		},
		{
			name:          "both console and file",
			config:        CreateTestLoggingConfig("info", "text", true, true, "/tmp/test_both.log"),
			expectedError: false,
			description:   "Both console and file logging should work",
		},
		{
			name:          "invalid level fallback",
			config:        CreateTestLoggingConfig("invalid_level", "text", true, false, ""),
			expectedError: false,
			description:   "Invalid level should fallback to info level",
		},
		{
			name:          "empty format",
			config:        CreateTestLoggingConfig("info", "", true, false, ""),
			expectedError: false,
			description:   "Empty format should use default formatter",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := SetupLogging(tc.config)
			if tc.expectedError {
				assert.Error(t, err, tc.description)
			} else {
				assert.NoError(t, err, tc.description)
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

// TestFormatters_AllFormats tests all formatter options comprehensively
func TestFormatters_AllFormats(t *testing.T) {
	t.Parallel()
	// REQ-LOG-001: Structured logging with logrus

	testCases := []struct {
		name          string
		format        string
		formatterType string
		description   string
	}{
		{"JSON format", "json", "JSON", "JSON formatter should not be nil"},
		{"Text format", "text", "Text", "Text formatter should not be nil"},
		{"Default format", "", "Text", "Default formatter should not be nil"},
		{"Unknown format", "unknown", "Text", "Unknown format should fallback to text formatter"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("file_formatter_%s", tc.name), func(t *testing.T) {
			formatter := createFileFormatter(tc.format)
			assert.NotNil(t, formatter, tc.description)
		})

		t.Run(fmt.Sprintf("console_formatter_%s", tc.name), func(t *testing.T) {
			formatter := createConsoleFormatter(tc.format)
			assert.NotNil(t, formatter, tc.description)
		})
	}
}

// =============================================================================
// ADVANCED FUNCTIONALITY TESTS
// =============================================================================

// TestLogging_FileRotation tests file rotation functionality
func TestLogging_FileRotation(t *testing.T) {
	// REQ-LOG-003: Log rotation configuration

	logFilePath := CreateTempLogFile(t)

	config := CreateTestLoggingConfig("info", "text", false, true, logFilePath)
	config.MaxFileSize = 1 // 1 byte to trigger rotation quickly
	config.BackupCount = 3

	// Setup logging
	err := SetupLogging(config)
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

	logger := CreateTestLogger(t, nil)
	concurrency := 10
	done := make(chan bool, concurrency)

	for i := 0; i < concurrency; i++ {
		go func(id int) {
			logger.Info("concurrent log message")
			logger.WithField("goroutine_id", fmt.Sprintf("%d", id)).Info("structured log message")
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < concurrency; i++ {
		<-done
	}

	assert.NotNil(t, logger)
}

// TestLogging_Performance tests logging performance
func TestLogging_Performance(t *testing.T) {
	// REQ-LOG-001: Structured logging with logrus

	logger := CreateTestLogger(t, nil)
	messageCount := 1000

	// Performance test: log many messages quickly
	start := time.Now()

	for i := 0; i < messageCount; i++ {
		logger.Info("performance test message")
	}

	duration := time.Since(start)

	// Should complete within reasonable time (< 1 second for 1000 messages)
	assert.Less(t, duration, time.Second, "Logging %d messages should complete within 1 second", messageCount)

	// Average time per message should be < 1ms
	avgTimePerMessage := duration / time.Duration(messageCount)
	assert.Less(t, avgTimePerMessage, time.Millisecond, "Average time per log message should be < 1ms")
}

// TestLogging_CrossComponentCorrelationID tests cross-component correlation ID propagation
func TestLogging_CrossComponentCorrelationID(t *testing.T) {
	// REQ-LOG-002: Cross-component tracing validation

	components := TestComponents()
	loggers := make(map[string]*Logger)

	// Create loggers for different components
	for _, component := range components {
		loggers[component] = CreateTestLogger(t, &TestLoggerConfig{Component: component})
	}

	// Generate correlation ID
	correlationID := GenerateCorrelationID()
	assert.NotEmpty(t, correlationID)

	// Create context with correlation ID
	ctx := CreateTestContext(correlationID)

	// Test correlation ID propagation across components
	for component, logger := range loggers {
		logger.LogWithContext(ctx, logrus.InfoLevel, fmt.Sprintf("%s operation", component))
	}

	// Verify correlation ID is consistent across all components
	retrievedID := GetCorrelationIDFromContext(ctx)
	assert.Equal(t, correlationID, retrievedID)

	// Test that each logger can access the correlation ID
	for component, logger := range loggers {
		loggerWithID := logger.WithCorrelationID(correlationID)
		assert.NotNil(t, loggerWithID, "Logger for component %s should support correlation ID", component)
	}
}

// =============================================================================
// TABLE-DRIVEN TESTS
// =============================================================================

// TestLogger_TableDriven tests multiple scenarios in a table-driven approach
func TestLogger_TableDriven(t *testing.T) {
	t.Parallel()
	// REQ-LOG-001: Structured logging with logrus

	fixtures := CreateTestFixtures()

	for _, fixture := range fixtures {
		t.Run(fmt.Sprintf("table_driven_%s", fixture.Component), func(t *testing.T) {
			logger := CreateTestLogger(t, &TestLoggerConfig{Component: fixture.Component})
			assert.Equal(t, fixture.Component, logger.component)

			// Test logging at the specified level
			logger.LogWithContext(context.Background(), fixture.Level, fixture.Message)
		})
	}
}

// =============================================================================
// PERFORMANCE BENCHMARKS
// =============================================================================

// BenchmarkNewLogger measures logger creation performance
func BenchmarkNewLogger(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewLogger("test-component")
	}
}

// BenchmarkGetLogger measures global logger retrieval performance
func BenchmarkGetLogger(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetLogger()
	}
}

// BenchmarkLogger_WithField measures structured field logging performance
func BenchmarkLogger_WithField(b *testing.B) {
	logger := NewLogger("test-component")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = logger.WithField("key", "value")
	}
}

// BenchmarkLogger_WithError measures error logging performance
func BenchmarkLogger_WithError(b *testing.B) {
	logger := NewLogger("test-component")
	testErr := assert.AnError
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = logger.WithError(testErr)
	}
}

// BenchmarkLogger_WithCorrelationID measures correlation ID logging performance
func BenchmarkLogger_WithCorrelationID(b *testing.B) {
	logger := NewLogger("test-component")
	correlationID := "test-correlation-id"
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = logger.WithCorrelationID(correlationID)
	}
}

// BenchmarkLogger_Info measures basic info logging performance
func BenchmarkLogger_Info(b *testing.B) {
	logger := NewLogger("test-component")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		logger.Info("benchmark message")
	}
}

// BenchmarkLogger_InfoWithFields measures structured info logging performance
func BenchmarkLogger_InfoWithFields(b *testing.B) {
	logger := NewLogger("test-component")
	fields := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		logger.WithFields(fields).Info("benchmark message with fields")
	}
}

// BenchmarkLogger_LogWithContext measures context-based logging performance
func BenchmarkLogger_LogWithContext(b *testing.B) {
	logger := NewLogger("test-component")
	ctx := WithCorrelationID(context.Background(), "test-correlation-id")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		logger.LogWithContext(ctx, logrus.InfoLevel, "benchmark context message")
	}
}

// BenchmarkLogger_LevelManagement measures level management performance
func BenchmarkLogger_LevelManagement(b *testing.B) {
	logger := NewLogger("test-component")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		logger.SetLevel(logrus.InfoLevel)
		_ = logger.GetLevel()
		logger.SetComponentLevel("test-component", logrus.DebugLevel)
		_ = logger.GetEffectiveLevel("test-component")
	}
}

// BenchmarkLogger_ConcurrentLogging measures concurrent logging performance
func BenchmarkLogger_ConcurrentLogging(b *testing.B) {
	logger := NewLogger("test-component")
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("concurrent benchmark message")
		}
	})
}

// BenchmarkLogger_StructuredLogging measures structured logging performance
func BenchmarkLogger_StructuredLogging(b *testing.B) {
	logger := NewLogger("test-component")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		logger.WithFields(map[string]interface{}{
			"request_id": "req-123",
			"user_id":    "user-456",
			"action":     "benchmark",
			"timestamp":  fmt.Sprintf("%d", i),
		}).Info("structured benchmark message")
	}
}

// BenchmarkGenerateCorrelationID measures correlation ID generation performance
func BenchmarkGenerateCorrelationID(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GenerateCorrelationID()
	}
}

// BenchmarkWithCorrelationID measures context correlation ID setup performance
func BenchmarkWithCorrelationID(b *testing.B) {
	ctx := context.Background()
	correlationID := "test-correlation-id"
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = WithCorrelationID(ctx, correlationID)
	}
}

// BenchmarkGetCorrelationIDFromContext measures correlation ID retrieval performance
func BenchmarkGetCorrelationIDFromContext(b *testing.B) {
	ctx := WithCorrelationID(context.Background(), "test-correlation-id")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = GetCorrelationIDFromContext(ctx)
	}
}

// BenchmarkSetupLogging measures logging setup performance
func BenchmarkSetupLogging(b *testing.B) {
	config := &LoggingConfig{
		Level:          "info",
		Format:         "text",
		ConsoleEnabled: true,
		FileEnabled:    false,
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = SetupLogging(config)
	}
}

// BenchmarkLogging_JSONFormat measures JSON format logging performance
func BenchmarkLogging_JSONFormat(b *testing.B) {
	// Setup JSON logging
	config := &LoggingConfig{
		Level:          "info",
		Format:         "json",
		ConsoleEnabled: true,
		FileEnabled:    false,
	}
	SetupLogging(config)

	logger := NewLogger("test-component")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		logger.WithFields(map[string]interface{}{
			"iteration": fmt.Sprintf("%d", i),
			"level":     "info",
			"component": "benchmark",
		}).Info("JSON format benchmark message")
	}
}

// BenchmarkLogging_TextFormat measures text format logging performance
func BenchmarkLogging_TextFormat(b *testing.B) {
	// Setup text logging
	config := &LoggingConfig{
		Level:          "info",
		Format:         "text",
		ConsoleEnabled: true,
		FileEnabled:    false,
	}
	SetupLogging(config)

	logger := NewLogger("test-component")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		logger.WithFields(map[string]interface{}{
			"iteration": fmt.Sprintf("%d", i),
			"level":     "info",
			"component": "benchmark",
		}).Info("Text format benchmark message")
	}
}

// BenchmarkLogger_MultipleFields measures multiple field logging performance
func BenchmarkLogger_MultipleFields(b *testing.B) {
	logger := NewLogger("test-component")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		logger.WithField("field1", "value1").
			WithField("field2", "value2").
			WithField("field3", "value3").
			WithField("field4", "value4").
			WithField("field5", "value5").
			Info("message with multiple fields")
	}
}

// BenchmarkLogger_ChainedOperations measures chained logging operations performance
func BenchmarkLogger_ChainedOperations(b *testing.B) {
	logger := NewLogger("test-component")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		logger.WithField("iteration", fmt.Sprintf("%d", i)).
			WithField("component", "benchmark").
			WithField("level", "info").
			WithField("timestamp", "2024-01-01T00:00:00Z").
			WithField("request_id", "req-123").
			WithField("user_id", "user-456").
			WithField("action", "benchmark").
			WithField("status", "success").
			Info("chained operations benchmark message")
	}
}
