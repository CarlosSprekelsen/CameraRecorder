//go:build integration

package mocks

import (
	"context"
	"sync"
	"time"
)

// MockAuditLogger captures audit calls for verification in integration tests.
// This replaces filesystem access with in-memory capture.
type MockAuditLogger struct {
	mu            sync.RWMutex
	LoggedActions []AuditCall
}

// AuditCall represents a captured audit log entry.
type AuditCall struct {
	Action  string
	RadioID string
	Result  string
	Latency time.Duration
}

// NewMockAuditLogger creates a new mock audit logger.
func NewMockAuditLogger() *MockAuditLogger {
	return &MockAuditLogger{
		LoggedActions: make([]AuditCall, 0),
	}
}

// LogAction captures the audit call in memory.
func (m *MockAuditLogger) LogAction(ctx context.Context, action, radioID, result string, latency time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.LoggedActions = append(m.LoggedActions, AuditCall{
		Action:  action,
		RadioID: radioID,
		Result:  result,
		Latency: latency,
	})
}

// GetLoggedActions returns a copy of all logged actions.
func (m *MockAuditLogger) GetLoggedActions() []AuditCall {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]AuditCall, len(m.LoggedActions))
	copy(result, m.LoggedActions)
	return result
}

// Clear removes all logged actions.
func (m *MockAuditLogger) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.LoggedActions = m.LoggedActions[:0]
}

// Close is a no-op for the mock.
func (m *MockAuditLogger) Close() error {
	return nil
}


