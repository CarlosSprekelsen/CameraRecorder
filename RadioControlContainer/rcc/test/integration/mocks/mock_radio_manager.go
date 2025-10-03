//go:build integration

package mocks

import (
	"time"

	"github.com/radio-control/rcc/internal/adapter"
	"github.com/radio-control/rcc/internal/command"
	"github.com/radio-control/rcc/internal/radio"
)

// MockRadioManager provides a minimal radio manager for integration tests.
type MockRadioManager struct {
	activeRadioID string
}

// NewMockRadioManager creates a new mock radio manager.
func NewMockRadioManager() *MockRadioManager {
	return &MockRadioManager{}
}

// LoadCapabilities is a stub implementation.
func (m *MockRadioManager) LoadCapabilities(radioID string, adapter adapter.IRadioAdapter, timeout time.Duration) error {
	m.activeRadioID = radioID
	return nil
}

// GetActiveRadioID returns the active radio ID.
func (m *MockRadioManager) GetActiveRadioID() string {
	return m.activeRadioID
}

// GetChannelByIndex returns a dummy channel.
func (m *MockRadioManager) GetChannelByIndex(radioID string, index int) (adapter.Channel, error) {
	return adapter.Channel{
		Index:        index,
		FrequencyMhz: 2412.0,
	}, nil
}

// GetRadio returns a dummy radio.
func (m *MockRadioManager) GetRadio(radioID string) (*radio.Radio, error) {
	return &radio.Radio{
		ID:     radioID,
		Model:  "FakeModel",
		Status: "online",
	}, nil
}

// Compile-time assertion that MockRadioManager implements command.RadioManager
var _ command.RadioManager = (*MockRadioManager)(nil)
