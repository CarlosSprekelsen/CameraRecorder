package logging

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger represents the main logging interface with correlation ID support.
// It wraps logrus.Logger and adds correlation ID tracking and component identification.
type Logger struct {
	*logrus.Logger
	correlationID string
	component     string
	mu            sync.RWMutex
}

// LoggingConfig represents logging configuration settings.
// It mirrors the configuration structure from the main config system.
type LoggingConfig struct {
	Level          string `mapstructure:"level"`           // Log level (debug, info, warn, error, fatal)
	Format         string `mapstructure:"format"`          // Output format (text, json)
	FileEnabled    bool   `mapstructure:"file_enabled"`    // Enable file logging
	FilePath       string `mapstructure:"file_path"`       // Log file path
	MaxFileSize    int    `mapstructure:"max_file_size"`   // Maximum file size in MB
	BackupCount    int    `mapstructure:"backup_count"`    // Number of backup files to keep
	ConsoleEnabled bool   `mapstructure:"console_enabled"` // Enable console logging
}

// NewLoggingConfigFromConfig creates a LoggingConfig from config.LoggingConfig.
// This function provides integration between the logging system and the main configuration system.
// Note: This function is moved to the config package to avoid import cycles.

// CorrelationIDKey is the context key for correlation IDs.
// Used for storing and retrieving correlation IDs from context.Context.
const CorrelationIDKey = "correlation_id"

// Global logger instance with thread-safe initialization
var (
	globalLogger *Logger
	once         sync.Once
)

// NewLogger creates a new logger instance for the specified component.
// The component name is used for identification in log messages.
func NewLogger(component string) *Logger {
	logger := &Logger{
		Logger:    logrus.New(),
		component: component,
	}

	// Set default formatter with timestamp
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	return logger
}

// GetLogger returns the global logger instance.
// Uses singleton pattern with thread-safe initialization.
func GetLogger() *Logger {
	once.Do(func() {
		globalLogger = NewLogger("camera-service")
	})
	return globalLogger
}

// SetupLogging initializes the logging system with the given configuration.
// Configures log level, formatters, and output handlers based on the provided config.
func SetupLogging(config *LoggingConfig) error {
	logger := GetLogger()

	// Parse and set log level
	level, err := logrus.ParseLevel(strings.ToLower(config.Level))
	if err != nil {
		level = logrus.InfoLevel // Fallback to info level
	}
	logger.SetLevel(level)

	// Clear existing hooks to avoid duplication
	logger.ReplaceHooks(logrus.LevelHooks{})

	// Setup console handler if enabled
	if config.ConsoleEnabled {
		consoleHandler := logrus.New()
		consoleHandler.SetOutput(os.Stdout)
		consoleHandler.SetFormatter(createConsoleFormatter(config.Format))
		consoleHandler.SetLevel(level)

		// Add console handler to logger
		logger.SetOutput(consoleHandler.Out)
		logger.SetFormatter(consoleHandler.Formatter)
	}

	// Setup file handler if enabled
	if config.FileEnabled && config.FilePath != "" {
		if err := setupFileHandler(logger, config); err != nil {
			return fmt.Errorf("failed to setup file handler: %w", err)
		}
	}

	return nil
}

// setupFileHandler configures file-based logging with rotation.
// Creates log directory if it doesn't exist and sets up lumberjack for log rotation.
func setupFileHandler(logger *Logger, config *LoggingConfig) error {
	// Ensure log directory exists
	logDir := filepath.Dir(config.FilePath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Create rotating file handler
	fileHandler := &lumberjack.Logger{
		Filename:   config.FilePath,
		MaxSize:    config.MaxFileSize / (1024 * 1024), // Convert to MB
		MaxBackups: config.BackupCount,
		MaxAge:     30, // Keep logs for 30 days
		Compress:   true,
	}

	// Create file formatter
	fileFormatter := createFileFormatter(config.Format)

	// Set file handler
	logger.SetOutput(fileHandler)
	logger.SetFormatter(fileFormatter)

	return nil
}

// createConsoleFormatter creates a console-friendly formatter.
// Uses text format with colors and timestamps for console output.
func createConsoleFormatter(format string) logrus.Formatter {
	return &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		DisableColors:   false,
		ForceColors:     true,
	}
}

// createFileFormatter creates a file formatter (JSON for production, text for development).
// Automatically selects JSON format for production environment or when explicitly requested.
func createFileFormatter(format string) logrus.Formatter {
	// Check if we should use JSON format
	if strings.Contains(strings.ToLower(format), "json") ||
		os.Getenv("CAMERA_SERVICE_ENV") == "production" {
		return &logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05",
		}
	}

	// Use text format for development
	return &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		DisableColors:   true,
	}
}

// WithCorrelationID creates a new logger with correlation ID.
// Thread-safe method that returns a new logger instance with the specified correlation ID.
func (l *Logger) WithCorrelationID(id string) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()

	newLogger := &Logger{
		Logger:        l.Logger,
		correlationID: id,
		component:     l.component,
	}

	return newLogger
}

// WithField adds a field to the logger.
// Returns a new logger instance with the specified key-value field added.
func (l *Logger) WithField(key, value string) *Logger {
	return &Logger{
		Logger:        l.Logger.WithField(key, value).Logger,
		correlationID: l.correlationID,
		component:     l.component,
	}
}

// WithError adds an error to the logger.
// Returns a new logger instance with the specified error added.
func (l *Logger) WithError(err error) *Logger {
	return &Logger{
		Logger:        l.Logger.WithError(err).Logger,
		correlationID: l.correlationID,
		component:     l.component,
	}
}

// Fields is a type alias for logrus.Fields to provide clean API
type Fields = logrus.Fields

// WithFields adds multiple fields to the logger.
// Returns a new logger instance with the specified fields added.
func (l *Logger) WithFields(fields Fields) *Logger {
	return &Logger{
		Logger:        l.Logger.WithFields(fields).Logger,
		correlationID: l.correlationID,
		component:     l.component,
	}
}

// LogWithContext logs a message with context information.
// Automatically adds correlation ID from context and component information to log entries.
func (l *Logger) LogWithContext(ctx context.Context, level logrus.Level, msg string) {
	entry := l.Logger.WithFields(Fields{
		"component": l.component,
	})

	// Add correlation ID if available
	if l.correlationID != "" {
		entry = entry.WithField("correlation_id", l.correlationID)
	}

	// Add correlation ID from context if not already set
	if correlationID := GetCorrelationIDFromContext(ctx); correlationID != "" {
		entry = entry.WithField("correlation_id", correlationID)
	}

	entry.Log(level, msg)
}

// GenerateCorrelationID generates a new correlation ID.
// Returns a UUID v4 string for request tracing and correlation.
func GenerateCorrelationID() string {
	return uuid.New().String()
}

// GetCorrelationIDFromContext extracts correlation ID from context.
// Returns empty string if no correlation ID is found in the context.
func GetCorrelationIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	if correlationID, ok := ctx.Value(CorrelationIDKey).(string); ok {
		return correlationID
	}

	return ""
}

// WithCorrelationID adds correlation ID to context.
// Creates a new context with the specified correlation ID value.
func WithCorrelationID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, CorrelationIDKey, id)
}

// LogWithCorrelationID logs a message with correlation ID from context.
// Convenience function that uses the global logger to log with context correlation ID.
func LogWithCorrelationID(ctx context.Context, level logrus.Level, msg string) {
	logger := GetLogger()
	logger.LogWithContext(ctx, level, msg)
}

// SetLevel sets the log level for the logger.
// Controls which log messages are output based on their severity level.
func (l *Logger) SetLevel(level logrus.Level) {
	l.Logger.SetLevel(level)
}

// SetComponentLevel sets the log level for a specific component.
// Currently uses a single logger instance; could be extended for component-specific loggers.
func (l *Logger) SetComponentLevel(component string, level logrus.Level) {
	// For now, we use a single logger instance
	// In a more complex implementation, we could maintain component-specific loggers
	l.SetLevel(level)
}

// GetEffectiveLevel returns the effective log level.
// Returns the current log level for the specified component.
func (l *Logger) GetEffectiveLevel(component string) logrus.Level {
	return l.Logger.GetLevel()
}

// IsLevelEnabled checks if a level is enabled.
// Returns true if the specified log level is enabled for output.
func (l *Logger) IsLevelEnabled(level logrus.Level) bool {
	return l.Logger.IsLevelEnabled(level)
}

// DebugWithContext logs a debug message with context information.
func (l *Logger) DebugWithContext(ctx context.Context, msg string) {
	l.LogWithContext(ctx, logrus.DebugLevel, msg)
}

func (l *Logger) InfoWithContext(ctx context.Context, msg string) {
	l.LogWithContext(ctx, logrus.InfoLevel, msg)
}

func (l *Logger) WarnWithContext(ctx context.Context, msg string) {
	l.LogWithContext(ctx, logrus.WarnLevel, msg)
}

func (l *Logger) ErrorWithContext(ctx context.Context, msg string) {
	l.LogWithContext(ctx, logrus.ErrorLevel, msg)
}

func (l *Logger) FatalWithContext(ctx context.Context, msg string) {
	l.LogWithContext(ctx, logrus.FatalLevel, msg)
	os.Exit(1)
}

// SetupLoggingSimple provides a simple logging setup.
// Creates a basic logging configuration with file and console output.
func SetupLoggingSimple(logPath string, level string) error {
	config := &LoggingConfig{
		Level:          level,
		FileEnabled:    logPath != "",
		FilePath:       logPath,
		ConsoleEnabled: true,
		MaxFileSize:    10485760, // 10MB
		BackupCount:    5,
	}

	return SetupLogging(config)
}
