package logging

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

// TestLoggerConfig represents a test logger configuration
type TestLoggerConfig struct {
	Component     string
	Level         logrus.Level
	Format        string
	ConsoleOutput bool
	FileOutput    bool
	FilePath      string
}

// TestFixture represents a test fixture with predefined values
type TestFixture struct {
	CorrelationID string
	Component     string
	Message       string
	Level         logrus.Level
	Fields        map[string]interface{}
}

// DefaultTestConfig returns a default test configuration
func DefaultTestConfig() *TestLoggerConfig {
	return &TestLoggerConfig{
		Component:     "test-component",
		Level:         logrus.InfoLevel,
		Format:        "text",
		ConsoleOutput: true,
		FileOutput:    false,
		FilePath:      "",
	}
}

// CreateTestLogger creates a logger for testing with the given configuration
func CreateTestLogger(t *testing.T, config *TestLoggerConfig) *Logger {
	t.Helper()

	if config == nil {
		config = DefaultTestConfig()
	}

	// Use the factory to get a logger with consistent configuration
	logger := GetLogger(config.Component)
	logger.SetLevel(config.Level)

	return logger
}

// CreateTestContext creates a test context with optional correlation ID
func CreateTestContext(correlationID string) context.Context {
	if correlationID == "" {
		return context.Background()
	}
	return WithCorrelationID(context.Background(), correlationID)
}

// CreateTestFixtures creates a set of test fixtures for different scenarios
func CreateTestFixtures() []TestFixture {
	return []TestFixture{
		{
			CorrelationID: TestCorrelationID1,
			Component:     "auth",
			Message:       "user authentication started",
			Level:         logrus.InfoLevel,
			Fields: map[string]interface{}{
				"user_id": "user-123",
				"action":  "login",
			},
		},
		{
			CorrelationID: TestCorrelationID2,
			Component:     "database",
			Message:       "database query executed",
			Level:         logrus.DebugLevel,
			Fields: map[string]interface{}{
				"query":    "SELECT * FROM users",
				"duration": "15ms",
			},
		},
		{
			CorrelationID: TestCorrelationID3,
			Component:     "api",
			Message:       "API request processed",
			Level:         logrus.InfoLevel,
			Fields: map[string]interface{}{
				"method":   "POST",
				"endpoint": "/api/v1/users",
				"status":   200,
			},
		},
		{
			CorrelationID: TestCorrelationID4,
			Component:     "camera",
			Message:       "camera stream started",
			Level:         logrus.InfoLevel,
			Fields: map[string]interface{}{
				"camera_id":  "cam-001",
				"resolution": "1920x1080",
				"fps":        30,
			},
		},
	}
}

// CreateTestLoggingConfig creates a test logging configuration
func CreateTestLoggingConfig(level, format string, consoleEnabled, fileEnabled bool, filePath string) *LoggingConfig {
	return &LoggingConfig{
		Level:          level,
		Format:         format,
		ConsoleEnabled: consoleEnabled,
		FileEnabled:    fileEnabled,
		FilePath:       filePath,
		MaxFileSize:    10,
		BackupCount:    3,
	}
}

// CreateTempLogFile creates a temporary log file for testing
func CreateTempLogFile(t *testing.T) string {
	t.Helper()

	tempDir, err := os.MkdirTemp("", "logging_test")
	require.NoError(t, err)

	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})

	logFilePath := filepath.Join(tempDir, "test.log")

	// Create the actual log file
	file, err := os.Create(logFilePath)
	require.NoError(t, err)
	file.Close()

	return logFilePath
}

// TestLogLevels returns all available log levels for testing
func TestLogLevels() []logrus.Level {
	return []logrus.Level{
		logrus.TraceLevel,
		logrus.DebugLevel,
		logrus.InfoLevel,
		logrus.WarnLevel,
		logrus.ErrorLevel,
		logrus.FatalLevel,
	}
}

// TestFormats returns all available log formats for testing
func TestFormats() []string {
	return []string{"text", "json", ""}
}

// TestComponents returns common component names for testing
func TestComponents() []string {
	return []string{"auth", "database", "api", "camera", "storage", "cache"}
}

// AssertLoggerBasicProperties asserts basic logger properties
func AssertLoggerBasicProperties(t *testing.T, logger *Logger, expectedComponent string) {
	t.Helper()

	require.NotNil(t, logger)
	require.NotNil(t, logger.Logger)
	require.Equal(t, expectedComponent, logger.component)
}

// AssertCorrelationIDInContext asserts that correlation ID is properly set in context
func AssertCorrelationIDInContext(t *testing.T, ctx context.Context, expectedID string) {
	t.Helper()

	if expectedID == "" {
		require.Empty(t, GetCorrelationIDFromContext(ctx))
	} else {
		require.Equal(t, expectedID, GetCorrelationIDFromContext(ctx))
	}
}
