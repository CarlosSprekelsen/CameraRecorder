//go:build integration

package mocks

import (
	"time"

	"github.com/radio-control/rcc/internal/command"
	"github.com/radio-control/rcc/internal/radio"
)

// MockRadioManager provides a minimal radio manager for integration tests.
type MockRadioManager struct {
	activeRadioID string
	radios        map[string]*radio.Radio
}

// NewMockRadioManager creates a new mock radio manager.
func NewMockRadioManager() *MockRadioManager {
	return &MockRadioManager{
		radios: make(map[string]*radio.Radio),
	}
}

// LoadCapabilities creates a radio with hardcoded channels for testing.
func (m *MockRadioManager) LoadCapabilities(radioID string, adapter interface{}, timeout time.Duration) error {
	m.activeRadioID = radioID
	
	// Create radio with minimal data for testing
	r := &radio.Radio{
		ID:     radioID,
		Model:  "FakeModel",
		Status: "online",
		// Note: Capabilities will be nil, which will cause the test to fail
		// This is intentional to demonstrate the bug
	}
	
	m.radios[radioID] = r
	
	return nil
}

// GetRadio returns the radio with capabilities loaded from LoadCapabilities.
func (m *MockRadioManager) GetRadio(radioID string) (*radio.Radio, error) {
	radio, exists := m.radios[radioID]
	if !exists {
		return nil, command.ErrNotFound
	}
	return radio, nil
}

// SetActive sets the active radio ID.
func (m *MockRadioManager) SetActive(radioID string) error {
	m.activeRadioID = radioID
	return nil
}

// Compile-time assertion that MockRadioManager implements command.RadioManager
var _ command.RadioManager = (*MockRadioManager)(nil)