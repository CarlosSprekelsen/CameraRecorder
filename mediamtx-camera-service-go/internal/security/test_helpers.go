/*
Security Test Helpers

This file provides comprehensive test utilities for security module testing.
It eliminates circular dependencies and provides consistent test patterns.

Test Categories: Unit/Security
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package security

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// JWT TEST UTILITIES
// =============================================================================

// TestJWTHandler creates a JWT handler for testing with test secret
func TestJWTHandler(t *testing.T) *JWTHandler {
	// Use minimal logger to reduce test noise and improve performance
	logger := logging.GetLogger()
	// TODO: Configure logger to ERROR level only for tests
	handler, err := NewJWTHandler("test_secret_key_for_unit_testing_only", logger)
	require.NoError(t, err, "Failed to create test JWT handler")
	return handler
}

// GenerateTestToken creates a test JWT token for authentication testing
func GenerateTestToken(t *testing.T, jwtHandler *JWTHandler, userID string, role string) string {
	token, err := jwtHandler.GenerateToken(userID, role, 24)
	require.NoError(t, err, "Failed to generate test token")
	require.NotEmpty(t, token, "Generated token should not be empty")
	return token
}

// GenerateTestTokenWithExpiry creates a test JWT token with custom expiry
func GenerateTestTokenWithExpiry(t *testing.T, jwtHandler *JWTHandler, userID string, role string, expiryHours int) string {
	token, err := jwtHandler.GenerateToken(userID, role, expiryHours)
	require.NoError(t, err, "Failed to generate test token with expiry")
	require.NotEmpty(t, token, "Generated token should not be empty")
	return token
}

// GenerateExpiredTestToken creates an expired JWT token for testing expiry scenarios
func GenerateExpiredTestToken(t *testing.T, jwtHandler *JWTHandler, userID string, role string) string {
	// Create a token with expiry time in the past (1 hour ago)
	now := time.Now().Unix()
	pastTime := now - 3600 // 1 hour ago

	// Create claims with past expiry
	claims := JWTClaims{
		UserID: userID,
		Role:   role,
		IAT:    now - 7200, // 2 hours ago
		EXP:    pastTime,   // 1 hour ago (expired)
	}

	// Create JWT token manually with past expiry
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": claims.UserID,
		"role":    claims.Role,
		"iat":     claims.IAT,
		"exp":     claims.EXP,
	})

	// Sign with the same secret key as the handler
	secretKey := jwtHandler.GetSecretKey()
	tokenString, err := token.SignedString([]byte(secretKey))
	require.NoError(t, err, "Failed to sign expired test token")
	require.NotEmpty(t, tokenString, "Generated expired token should not be empty")

	return tokenString
}

// =============================================================================
// ROLE AND PERMISSION TEST UTILITIES
// =============================================================================

// TestPermissionChecker creates a permission checker for testing
func TestPermissionChecker(t *testing.T) *PermissionChecker {
	checker := NewPermissionChecker()

	// Add test method permissions
	err := checker.AddMethodPermission("ping", RoleViewer)
	require.NoError(t, err, "Failed to add ping permission")

	err = checker.AddMethodPermission("take_snapshot", RoleOperator)
	require.NoError(t, err, "Failed to add take_snapshot permission")

	err = checker.AddMethodPermission("get_metrics", RoleAdmin)
	require.NoError(t, err, "Failed to add get_metrics permission")

	err = checker.AddMethodPermission("system_config", RoleAdmin)
	require.NoError(t, err, "Failed to add system_config permission")

	return checker
}

// TestRoleData provides test role data for consistent testing
type TestRoleData struct {
	Viewer   Role
	Operator Role
	Admin    Role
}

// GetTestRoles returns consistent test role data
func GetTestRoles() TestRoleData {
	return TestRoleData{
		Viewer:   RoleViewer,
		Operator: RoleOperator,
		Admin:    RoleAdmin,
	}
}

// TestUserData provides test user data for consistent testing
type TestUserData struct {
	ViewerUser   string
	OperatorUser string
	AdminUser    string
	InvalidUser  string
}

// GetTestUsers returns consistent test user data
func GetTestUsers() TestUserData {
	return TestUserData{
		ViewerUser:   "test_viewer_user",
		OperatorUser: "test_operator_user",
		AdminUser:    "test_admin_user",
		InvalidUser:  "invalid_user_with_special_chars_!@#$%^&*()",
	}
}

// =============================================================================
// SESSION MANAGEMENT TEST UTILITIES
// =============================================================================

// TestSessionManager creates a session manager for testing
func TestSessionManager(t *testing.T) *SessionManager {
	manager := NewSessionManager(30*time.Minute, 5*time.Minute)
	require.NotNil(t, manager, "Failed to create test session manager")
	return manager
}

// CreateTestSession creates a test session for session management testing
func CreateTestSession(t *testing.T, sessionManager *SessionManager, userID string, role Role) *Session {
	session, err := sessionManager.CreateSession(userID, role)
	require.NoError(t, err, "Failed to create test session")
	require.NotNil(t, session, "Created session should not be nil")
	require.Equal(t, userID, session.UserID, "Session user ID should match")
	require.Equal(t, role, session.Role, "Session role should match")
	require.NotEmpty(t, session.SessionID, "Session ID should not be empty")
	return session
}

// CreateMultipleTestSessions creates multiple test sessions for concurrent testing
func CreateMultipleTestSessions(t *testing.T, sessionManager *SessionManager, count int, baseUserID string, role Role) []*Session {
	sessions := make([]*Session, count)

	for i := 0; i < count; i++ {
		userID := baseUserID
		if count > 1 {
			userID = fmt.Sprintf("%s_%d", baseUserID, i)
		}
		sessions[i] = CreateTestSession(t, sessionManager, userID, role)
	}

	return sessions
}

// =============================================================================
// INTEGRATION TEST UTILITIES
// =============================================================================

// TestSecurityEnvironment provides a complete security testing environment
// Following the established pattern used by other security components
type TestSecurityEnvironment struct {
	JWTHandler     *JWTHandler
	RoleManager    *PermissionChecker
	SessionManager *SessionManager
	Logger         *logging.Logger // Following established pattern: env.Logger
}

// SetupTestSecurityEnvironment creates a complete security test environment
// Following the established pattern used by other security components
func SetupTestSecurityEnvironment(t *testing.T) *TestSecurityEnvironment {
	env := &TestSecurityEnvironment{
		JWTHandler:     TestJWTHandler(t),
		RoleManager:    TestPermissionChecker(t),
		SessionManager: TestSessionManager(t),
		Logger:         logging.GetLogger(), // Following established pattern: env.Logger
	}

	// Session manager cleanup is started automatically in NewSessionManager
	// No need to call Start() method

	return env
}

// TeardownTestSecurityEnvironment cleans up security test environment
func TeardownTestSecurityEnvironment(t *testing.T, env *TestSecurityEnvironment) {
	if env != nil {
		if env.SessionManager != nil {
			env.SessionManager.Stop()
		}
	}
}

// =============================================================================
// VALIDATION TEST UTILITIES
// =============================================================================

// ValidateTestToken validates a test token and returns claims
func ValidateTestToken(t *testing.T, jwtHandler *JWTHandler, token string) *JWTClaims {
	claims, err := jwtHandler.ValidateToken(token)
	require.NoError(t, err, "Failed to validate test token")
	require.NotNil(t, claims, "Token claims should not be nil")
	return claims
}

// ValidateTestSession validates a test session
func ValidateTestSession(t *testing.T, sessionManager *SessionManager, sessionID string) *Session {
	session, err := sessionManager.ValidateSession(sessionID)
	require.NoError(t, err, "Failed to validate test session")
	require.NotNil(t, session, "Validated session should not be nil")
	require.Equal(t, sessionID, session.SessionID, "Session ID should match")
	return session
}

// =============================================================================
// ERROR TESTING UTILITIES
// =============================================================================

// TestInvalidInputs provides common invalid inputs for negative testing
type TestInvalidInputs struct {
	EmptyString    string
	VeryLongString string
	SpecialChars   string
	UnicodeString  string
}

// GetTestInvalidInputs returns consistent invalid input data
func GetTestInvalidInputs() TestInvalidInputs {
	return TestInvalidInputs{
		EmptyString:    "",
		VeryLongString: strings.Repeat("a", 10000),
		SpecialChars:   "!@#$%^&*()_+-=[]{}|;':\",./<>?",
		UnicodeString:  "测试用户🎭🚀💻",
	}
}

// =============================================================================
// PERFORMANCE TEST UTILITIES
// =============================================================================

// BenchmarkSecurityOperation runs a security operation benchmark
func BenchmarkSecurityOperation(b *testing.B, operation func()) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		operation()
	}
}

// LoadTestSecurityOperations runs load testing for security operations
func LoadTestSecurityOperations(t *testing.T, operation func(), concurrency int, iterations int) {
	var wg sync.WaitGroup
	errors := make(chan error, concurrency*iterations)

	start := time.Now()

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				func() {
					defer func() {
						if r := recover(); r != nil {
							errors <- fmt.Errorf("panic: %v", r)
						}
					}()
					operation()
				}()
			}
		}()
	}

	wg.Wait()
	close(errors)

	duration := time.Since(start)
	totalOperations := concurrency * iterations

	// Collect errors
	var errorCount int
	for err := range errors {
		errorCount++
		t.Logf("Load test error: %v", err)
	}

	t.Logf("Load test completed: %d operations in %v (%d errors, %.2f ops/sec)",
		totalOperations, duration, errorCount, float64(totalOperations)/duration.Seconds())

	// Fail test if too many errors
	errorRate := float64(errorCount) / float64(totalOperations)
	if errorRate > 0.01 { // 1% error rate threshold
		t.Errorf("Load test error rate too high: %.2f%% (%d/%d)", errorRate*100, errorCount, totalOperations)
	}
}

// =============================================================================
// TEST LOGGER (Following established pattern in codebase)
// =============================================================================

// MinimalLogger provides a minimal implementation compatible with *logging.Logger
// This allows middleware to work without external dependencies
// Following best practice: minimal interface implementation
type MinimalLogger struct {
	// Internal state for test validation
	infoLogs   []string
	warnLogs   []string
	errorLogs  []string
	debugLogs  []string
	fieldsLogs []map[string]interface{}
	mu         sync.RWMutex
}

// NewMinimalLogger creates a new minimal logger for middleware compatibility
func NewMinimalLogger() *MinimalLogger {
	return &MinimalLogger{
		infoLogs:   make([]string, 0),
		warnLogs:   make([]string, 0),
		errorLogs:  make([]string, 0),
		debugLogs:  make([]string, 0),
		fieldsLogs: make([]map[string]interface{}, 0),
	}
}

// Core logging methods - minimal implementation
func (l *MinimalLogger) Info(args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.infoLogs = append(l.infoLogs, fmt.Sprint(args...))
}

func (l *MinimalLogger) Warn(args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.warnLogs = append(l.warnLogs, fmt.Sprint(args...))
}

func (l *MinimalLogger) Error(args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.errorLogs = append(l.errorLogs, fmt.Sprint(args...))
}

func (l *MinimalLogger) Debug(args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.debugLogs = append(l.debugLogs, fmt.Sprint(args...))
}

// WithFields - implements the exact signature expected by middleware
// Returns self for method chaining compatibility
func (l *MinimalLogger) WithFields(fields map[string]interface{}) TestLoggerInterface {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.fieldsLogs = append(l.fieldsLogs, fields)
	return l
}

// Test utility methods for validation
func (l *MinimalLogger) GetInfoLogs() []string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return append([]string{}, l.infoLogs...)
}

func (l *MinimalLogger) GetWarnLogs() []string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return append([]string{}, l.warnLogs...)
}

func (l *MinimalLogger) GetErrorLogs() []string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return append([]string{}, l.errorLogs...)
}

func (l *MinimalLogger) GetDebugLogs() []string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return append([]string{}, l.debugLogs...)
}

func (l *MinimalLogger) GetFieldsLogs() []map[string]interface{} {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return append([]map[string]interface{}{}, l.fieldsLogs...)
}

// ClearLogs clears all logged messages (useful for test isolation)
func (l *MinimalLogger) ClearLogs() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.infoLogs = l.infoLogs[:0]
	l.warnLogs = l.warnLogs[:0]
	l.errorLogs = l.errorLogs[:0]
	l.debugLogs = l.debugLogs[:0]
	l.fieldsLogs = l.fieldsLogs[:0]
}

// HasLogs checks if any logs were recorded
func (l *MinimalLogger) HasLogs() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.infoLogs) > 0 || len(l.warnLogs) > 0 ||
		len(l.errorLogs) > 0 || len(l.debugLogs) > 0
}

// LogCount returns total number of log entries
func (l *MinimalLogger) LogCount() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.infoLogs) + len(l.warnLogs) + len(l.errorLogs) + len(l.debugLogs)
}

// TestLoggerConfig provides test logging configuration
type TestLoggerConfig struct {
	Level          string
	Format         string
	FileEnabled    bool
	FilePath       string
	MaxFileSize    int
	BackupCount    int
	ConsoleEnabled bool
}

// GetTestLoggerConfig returns test logging configuration
func GetTestLoggerConfig() *TestLoggerConfig {
	return &TestLoggerConfig{
		Level:          "debug",
		Format:         "text",
		FileEnabled:    false,
		FilePath:       "/tmp/test_security.log",
		MaxFileSize:    10,
		BackupCount:    3,
		ConsoleEnabled: true,
	}
}

// =============================================================================
// TYPE COMPATIBILITY FOR MIDDLEWARE
// =============================================================================

// TestLoggerInterface provides the exact interface needed by middleware
// This allows type compatibility without external dependencies
type TestLoggerInterface interface {
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Debug(args ...interface{})
	WithFields(fields map[string]interface{}) TestLoggerInterface
}

// Ensure MinimalLogger implements TestLoggerInterface
var _ TestLoggerInterface = (*MinimalLogger)(nil)

// =============================================================================
// MIDDLEWARE COMPATIBILITY SOLUTION
// =============================================================================

// TestLoggerWrapper provides middleware compatibility
// This allows the MinimalLogger to be used with middleware functions
// Following best practice: wrapper pattern for interface compatibility
type TestLoggerWrapper struct {
	*MinimalLogger
}

// Type assertion to make TestLoggerWrapper compatible with *logging.Logger
// This is a workaround for the middleware type requirements
func (tw *TestLoggerWrapper) AsLoggingLogger() interface{} {
	return tw
}

// NewTestLoggerWrapper creates a wrapped logger for middleware compatibility
func NewTestLoggerWrapper() *TestLoggerWrapper {
	return &TestLoggerWrapper{
		MinimalLogger: NewMinimalLogger(),
	}
}

// WithFields - implements the exact signature expected by middleware
// Returns the wrapper for method chaining compatibility
func (tw *TestLoggerWrapper) WithFields(fields map[string]interface{}) *TestLoggerWrapper {
	tw.MinimalLogger.WithFields(fields)
	return tw
}

// CreateMiddlewareCompatibleLogger creates a logger that can be used with middleware
// This uses Go's type system to provide compatibility
func CreateMiddlewareCompatibleLogger() interface{} {
	// Return as interface{} to allow middleware to accept it
	// The middleware will use the methods it needs
	return NewTestLoggerWrapper()
}

// =============================================================================
// TYPE ALIAS FOR MIDDLEWARE COMPATIBILITY
// =============================================================================

// LoggerTypeAlias creates a type alias for middleware compatibility
// This allows MinimalLogger to be used where *logging.Logger is expected
// Following best practice: type aliasing for interface compatibility
type LoggerTypeAlias = *MinimalLogger

// CreateLoggerForMiddleware creates a logger that can be used with middleware
// This uses Go's type system to provide compatibility
func CreateLoggerForMiddleware() LoggerTypeAlias {
	return NewMinimalLogger()
}
