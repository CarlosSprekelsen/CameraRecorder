package security

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// SecurityEventType defines the type of security event
type SecurityEventType string

const (
	EventAuthenticationSuccess SecurityEventType = "authentication_success"
	EventAuthenticationFailure SecurityEventType = "authentication_failure"
	EventAuthorizationSuccess  SecurityEventType = "authorization_success"
	EventAuthorizationFailure  SecurityEventType = "authorization_failure"
	EventRateLimitExceeded     SecurityEventType = "rate_limit_exceeded"
	EventInputValidationFailed SecurityEventType = "input_validation_failed"
	EventMethodAccess          SecurityEventType = "method_access"
	EventClientConnection      SecurityEventType = "client_connection"
	EventClientDisconnection   SecurityEventType = "client_disconnection"
	EventSecurityViolation     SecurityEventType = "security_violation"
	EventSystemAccess          SecurityEventType = "system_access"
)

// RiskLevel defines the risk level of a security event
type RiskLevel int

const (
	RiskLevelLow RiskLevel = iota
	RiskLevelMedium
	RiskLevelHigh
	RiskLevelCritical
)

func (rl RiskLevel) String() string {
	switch rl {
	case RiskLevelLow:
		return "low"
	case RiskLevelMedium:
		return "medium"
	case RiskLevelHigh:
		return "high"
	case RiskLevelCritical:
		return "critical"
	default:
		return "unknown"
	}
}

// SecurityEvent represents a security event for auditing
type SecurityEvent struct {
	Timestamp   time.Time         `json:"timestamp"`
	EventID     string            `json:"event_id"`
	EventType   SecurityEventType `json:"event_type"`
	ClientID    string            `json:"client_id"`
	IPAddress   string            `json:"ip_address,omitempty"`
	UserID      string            `json:"user_id,omitempty"`
	Role        string            `json:"role,omitempty"`
	Method      string            `json:"method,omitempty"`
	Success     bool              `json:"success"`
	ErrorCode   int               `json:"error_code,omitempty"`
	ErrorMessage string           `json:"error_message,omitempty"`
	RiskLevel   RiskLevel         `json:"risk_level"`
	RiskScore   int               `json:"risk_score"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CorrelationID string          `json:"correlation_id,omitempty"`
}

// SecurityAuditLogger provides comprehensive security event logging and monitoring
type SecurityAuditLogger struct {
	logger        *logrus.Logger
	events        chan *SecurityEvent
	eventHandlers []func(*SecurityEvent)
	mutex         sync.RWMutex
	config        *AuditLoggerConfig
	securityConfig interface{} // Will be typed based on existing config structure
	
	// File rotation
	currentLogFile *os.File
	logFilePath    string
	maxFileSize    int64
	maxFileAge     time.Duration
	
	// Statistics
	eventCount    int64
	errorCount    int64
	lastRotation  time.Time
}

// AuditLoggerConfig defines configuration for the audit logger
type AuditLoggerConfig struct {
	LogDirectory     string        `json:"log_directory"`
	MaxFileSize      int64         `json:"max_file_size"`      // in bytes
	MaxFileAge       time.Duration `json:"max_file_age"`       // how long to keep files
	RotationInterval time.Duration `json:"rotation_interval"`  // how often to check rotation
	BufferSize       int           `json:"buffer_size"`        // event channel buffer size
	EnableFileLogging bool         `json:"enable_file_logging"`
	EnableConsoleLogging bool      `json:"enable_console_logging"`
	LogLevel         string        `json:"log_level"`
}

// DefaultAuditLoggerConfig returns default configuration
func DefaultAuditLoggerConfig() *AuditLoggerConfig {
	return &AuditLoggerConfig{
		LogDirectory:      "/var/log/camera-service/security",
		MaxFileSize:       100 * 1024 * 1024, // 100 MB
		MaxFileAge:        30 * 24 * time.Hour, // 30 days
		RotationInterval:  1 * time.Hour,
		BufferSize:        1000,
		EnableFileLogging: true,
		EnableConsoleLogging: true,
		LogLevel:          "info",
	}
}

// NewSecurityAuditLogger creates a new security audit logger
func NewSecurityAuditLogger(config *AuditLoggerConfig, logger *logrus.Logger, securityConfig interface{}) (*SecurityAuditLogger, error) {
	if config == nil {
		config = DefaultAuditLoggerConfig()
	}

	auditLogger := &SecurityAuditLogger{
		logger:         logger,
		events:         make(chan *SecurityEvent, config.BufferSize),
		eventHandlers:  make([]func(*SecurityEvent), 0),
		config:         config,
		securityConfig: securityConfig,
		lastRotation:   time.Now(),
	}

	// Create log directory if it doesn't exist
	if config.EnableFileLogging {
		if err := os.MkdirAll(config.LogDirectory, 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}

		// Initialize log file
		if err := auditLogger.initializeLogFile(); err != nil {
			return nil, fmt.Errorf("failed to initialize log file: %w", err)
		}
	}

	// Start event processing
	go auditLogger.processEvents()

	// Start file rotation routine
	if config.EnableFileLogging {
		go auditLogger.startRotationRoutine()
	}

	return auditLogger, nil
}

// initializeLogFile creates and opens a new log file
func (sal *SecurityAuditLogger) initializeLogFile() error {
	if sal.currentLogFile != nil {
		sal.currentLogFile.Close()
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("security_audit_%s.log", timestamp)
	sal.logFilePath = filepath.Join(sal.config.LogDirectory, filename)

	file, err := os.OpenFile(sal.logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	sal.currentLogFile = file
	sal.lastRotation = time.Now()

	return nil
}

// processEvents processes security events from the channel
func (sal *SecurityAuditLogger) processEvents() {
	for event := range sal.events {
		sal.processEvent(event)
	}
}

// processEvent processes a single security event
func (sal *SecurityAuditLogger) processEvent(event *SecurityEvent) {
	// Increment counters
	sal.eventCount++
	if !event.Success {
		sal.errorCount++
	}

	// Log to console if enabled
	if sal.config.EnableConsoleLogging {
		sal.logToConsole(event)
	}

	// Write to file if enabled
	if sal.config.EnableFileLogging {
		sal.logToFile(event)
	}

	// Call event handlers
	sal.mutex.RLock()
	for _, handler := range sal.eventHandlers {
		handler(event)
	}
	sal.mutex.RUnlock()

	// Check if rotation is needed
	if sal.config.EnableFileLogging && sal.shouldRotate() {
		if err := sal.rotateLogFile(); err != nil {
			sal.logger.WithError(err).Error("Failed to rotate log file")
		}
	}
}

// logToConsole logs the event to the console using structured logging
func (sal *SecurityAuditLogger) logToConsole(event *SecurityEvent) {
	fields := logrus.Fields{
		"event_type":   event.EventType,
		"client_id":    event.ClientID,
		"success":      event.Success,
		"risk_level":   event.RiskLevel.String(),
		"risk_score":   event.RiskScore,
		"timestamp":    event.Timestamp,
	}

	if event.UserID != "" {
		fields["user_id"] = event.UserID
	}
	if event.Role != "" {
		fields["role"] = event.Role
	}
	if event.Method != "" {
		fields["method"] = event.Method
	}
	if event.ErrorCode != 0 {
		fields["error_code"] = event.ErrorCode
	}
	if event.ErrorMessage != "" {
		fields["error_message"] = event.ErrorMessage
	}
	if event.IPAddress != "" {
		fields["ip_address"] = event.IPAddress
	}

	level := logrus.InfoLevel
	if !event.Success {
		level = logrus.WarnLevel
	}
	if event.RiskLevel >= RiskLevelHigh {
		level = logrus.ErrorLevel
	}

	sal.logger.WithFields(fields).Log(level, "Security event logged")
}

// logToFile writes the event to the log file in JSON format
func (sal *SecurityAuditLogger) logToFile(event *SecurityEvent) {
	if sal.currentLogFile == nil {
		return
	}

	// Marshal event to JSON
	data, err := json.Marshal(event)
	if err != nil {
		sal.logger.WithError(err).Error("Failed to marshal security event")
		return
	}

	// Write to file with newline
	if _, err := sal.currentLogFile.Write(append(data, '\n')); err != nil {
		sal.logger.WithError(err).Error("Failed to write security event to file")
	}
}

// shouldRotate checks if log file rotation is needed
func (sal *SecurityAuditLogger) shouldRotate() bool {
	if sal.currentLogFile == nil {
		return false
	}

	// Check file size
	if info, err := sal.currentLogFile.Stat(); err == nil {
		if info.Size() > sal.config.MaxFileSize {
			return true
		}
	}

	// Check time since last rotation
	if time.Since(sal.lastRotation) > sal.config.RotationInterval {
		return true
	}

	return false
}

// rotateLogFile rotates the current log file
func (sal *SecurityAuditLogger) rotateLogFile() error {
	sal.logger.Info("Rotating security audit log file")

	// Close current file
	if sal.currentLogFile != nil {
		sal.currentLogFile.Close()
	}

	// Initialize new log file
	if err := sal.initializeLogFile(); err != nil {
		return err
	}

	// Clean up old log files
	return sal.cleanupOldLogFiles()
}

// cleanupOldLogFiles removes log files older than the configured max age
func (sal *SecurityAuditLogger) cleanupOldLogFiles() error {
	entries, err := os.ReadDir(sal.config.LogDirectory)
	if err != nil {
		return err
	}

	cutoff := time.Now().Add(-sal.config.MaxFileAge)
	removed := 0

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".log" {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoff) {
			filePath := filepath.Join(sal.config.LogDirectory, entry.Name())
			if err := os.Remove(filePath); err != nil {
				sal.logger.WithError(err).WithField("file", filePath).Warn("Failed to remove old log file")
			} else {
				removed++
			}
		}
	}

	if removed > 0 {
		sal.logger.WithField("removed_files", removed).Info("Old security audit log files cleaned up")
	}

	return nil
}

// startRotationRoutine starts the background log rotation routine
func (sal *SecurityAuditLogger) startRotationRoutine() {
	ticker := time.NewTicker(sal.config.RotationInterval)
	defer ticker.Stop()

	for range ticker.C {
		if sal.shouldRotate() {
			if err := sal.rotateLogFile(); err != nil {
				sal.logger.WithError(err).Error("Failed to rotate log file during routine check")
			}
		}
	}
}

// LogSecurityEvent logs a security event
func (sal *SecurityAuditLogger) LogSecurityEvent(event *SecurityEvent) {
	// Generate event ID if not provided
	if event.EventID == "" {
		event.EventID = sal.generateEventID()
	}

	// Set timestamp if not provided
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Calculate risk score if not provided
	if event.RiskScore == 0 {
		event.RiskScore = sal.calculateRiskScore(event)
	}

	// Send event to processing channel
	select {
	case sal.events <- event:
		// Event queued successfully
	default:
		// Channel is full, log warning
		sal.logger.Warn("Security audit event channel is full, event dropped")
	}
}

// generateEventID generates a unique event ID
func (sal *SecurityAuditLogger) generateEventID() string {
	return fmt.Sprintf("evt_%d_%d", time.Now().UnixNano(), sal.eventCount)
}

// calculateRiskScore calculates a risk score for an event
func (sal *SecurityAuditLogger) calculateRiskScore(event *SecurityEvent) int {
	score := 0

	// Base score based on event type
	switch event.EventType {
	case EventAuthenticationFailure:
		score += 20
	case EventAuthorizationFailure:
		score += 30
	case EventRateLimitExceeded:
		score += 15
	case EventInputValidationFailed:
		score += 10
	case EventSecurityViolation:
		score += 50
	}

	// Adjust score based on success/failure
	if !event.Success {
		score += 25
	}

	// Adjust score based on role
	if event.Role == "admin" {
		score += 20
	} else if event.Role == "operator" {
		score += 10
	}

	// Adjust score based on method sensitivity
	if event.Method == "authenticate" || event.Method == "start_recording" {
		score += 15
	}

	// Cap score at 100
	if score > 100 {
		score = 100
	}

	return score
}

// AddEventHandler adds a new event handler
func (sal *SecurityAuditLogger) AddEventHandler(handler func(*SecurityEvent)) {
	sal.mutex.Lock()
	defer sal.mutex.Unlock()

	sal.eventHandlers = append(sal.eventHandlers, handler)
}

// GetStats returns audit logger statistics
func (sal *SecurityAuditLogger) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"total_events":     sal.eventCount,
		"error_events":     sal.errorCount,
		"success_rate":     float64(sal.eventCount-sal.errorCount) / float64(sal.eventCount) * 100,
		"buffer_size":      cap(sal.events),
		"current_buffer":   len(sal.events),
		"event_handlers":   len(sal.eventHandlers),
		"file_logging":     sal.config.EnableFileLogging,
		"console_logging":  sal.config.EnableConsoleLogging,
		"log_directory":    sal.config.LogDirectory,
		"max_file_size":    sal.config.MaxFileSize,
		"max_file_age":     sal.config.MaxFileAge,
	}
}

// Close closes the audit logger and cleans up resources
func (sal *SecurityAuditLogger) Close() error {
	// Close event channel
	close(sal.events)

	// Close current log file
	if sal.currentLogFile != nil {
		return sal.currentLogFile.Close()
	}

	return nil
}

// Convenience methods for common security events

// LogAuthSuccess logs a successful authentication event
func (sal *SecurityAuditLogger) LogAuthSuccess(clientID, userID, role, ipAddress string) {
	sal.LogSecurityEvent(&SecurityEvent{
		EventType: EventAuthenticationSuccess,
		ClientID:  clientID,
		UserID:    userID,
		Role:      role,
		IPAddress: ipAddress,
		Success:   true,
		RiskLevel: RiskLevelLow,
		RiskScore: 5,
	})
}

// LogAuthFailure logs a failed authentication event
func (sal *SecurityAuditLogger) LogAuthFailure(clientID, ipAddress string, errorCode int, errorMessage string) {
	sal.LogSecurityEvent(&SecurityEvent{
		EventType:    EventAuthenticationFailure,
		ClientID:     clientID,
		IPAddress:    ipAddress,
		Success:      false,
		ErrorCode:    errorCode,
		ErrorMessage: errorMessage,
		RiskLevel:    RiskLevelMedium,
		RiskScore:    35,
	})
}



// LogMethodAccess logs a method access event
func (sal *SecurityAuditLogger) LogMethodAccess(clientID, userID, role, method string, success bool) {
	riskLevel := RiskLevelLow
	riskScore := 5

	if !success {
		riskLevel = RiskLevelMedium
		riskScore = 25
	}

	sal.LogSecurityEvent(&SecurityEvent{
		EventType: EventMethodAccess,
		ClientID:  clientID,
		UserID:    userID,
		Role:      role,
		Method:    method,
		Success:   success,
		RiskLevel: riskLevel,
		RiskScore: riskScore,
	})
}

// LogRateLimitExceeded logs a rate limit exceeded event
func (sal *SecurityAuditLogger) LogRateLimitExceeded(clientID, method, ipAddress string) {
	sal.LogSecurityEvent(&SecurityEvent{
		EventType:    EventRateLimitExceeded,
		ClientID:     clientID,
		Method:       method,
		IPAddress:    ipAddress,
		Success:      false,
		RiskLevel:    RiskLevelMedium,
		RiskScore:    30,
		Metadata: map[string]interface{}{
			"rate_limit_type": "method_specific",
		},
	})
}
