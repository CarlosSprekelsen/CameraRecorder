package logging

import (
	"sync"

	"github.com/sirupsen/logrus"
)

// LoggerFactory provides centralized logger creation with consistent configuration.
//
// The factory ensures all loggers created share the same global configuration
// for level, format, and output destinations. Thread-safe for concurrent access.
type LoggerFactory struct {
	config *LoggingConfig
	mu     sync.RWMutex
}

// Global logger factory instance
var (
	factory     *LoggerFactory
	factoryOnce sync.Once
)

// GetLoggerFactory returns the global logger factory instance
func GetLoggerFactory() *LoggerFactory {
	factoryOnce.Do(func() {
		factory = &LoggerFactory{
			config: &LoggingConfig{
				Level:          "info",
				Format:         "text",
				FileEnabled:    false,
				ConsoleEnabled: true,
			},
		}
	})
	return factory
}

// ConfigureFactory sets the global configuration for the logger factory
func ConfigureFactory(config *LoggingConfig) {
	factory := GetLoggerFactory()
	factory.mu.Lock()
	defer factory.mu.Unlock()

	if config != nil {
		factory.config = config
	}
}

// CreateLogger creates a new logger instance for the specified component.
func (f *LoggerFactory) CreateLogger(component string) *Logger {
	f.mu.RLock()
	config := f.config
	f.mu.RUnlock()

	logger := &Logger{
		Logger:    logrus.New(),
		component: component,
	}

	// Apply global configuration
	f.configureLogger(logger, config)

	return logger
}

// configureLogger applies the configuration to a logger instance
func (f *LoggerFactory) configureLogger(logger *Logger, config *LoggingConfig) {
	// Set log level
	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// Set formatter based on configuration
	if config.Format == "json" {
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05",
		})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
	}

	// Configure output based on settings
	if !config.ConsoleEnabled && !config.FileEnabled {
		// If both disabled, use a no-op output
		logger.SetOutput(&noOpWriter{})
	}
	// Note: File and console output configuration would be handled by SetupLogging
	// The factory focuses on consistent logger creation with proper configuration
}

// noOpWriter is a no-op writer for when logging is disabled
type noOpWriter struct{}

func (w *noOpWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

// GetLogger is a convenience function that uses the global factory
func GetLogger(component string) *Logger {
	return GetLoggerFactory().CreateLogger(component)
}

// ConfigureGlobalLogging configures the global logger factory and sets up logging
func ConfigureGlobalLogging(config *LoggingConfig) error {
	// Configure the factory
	ConfigureFactory(config)

	// Also configure the global logger
	return SetupLogging(config)
}
