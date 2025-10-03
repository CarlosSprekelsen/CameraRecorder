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
	radios        map[string]*radio.Radio
}

// NewMockRadioManager creates a new mock radio manager.
func NewMockRadioManager() *MockRadioManager {
	return &MockRadioManager{
		radios: make(map[string]*radio.Radio),
	}
}

// LoadCapabilities creates a radio with hardcoded channels from ICD for testing.
func (m *MockRadioManager) LoadCapabilities(radioID string, adapter interface{}, timeout time.Duration) error {
	m.activeRadioID = radioID

	// Create radio - capabilities will be populated in GetRadio method
	r := &radio.Radio{
		ID:     radioID,
		Model:  "FakeModel",
		Status: "online",
		// Capabilities will be nil initially, populated in GetRadio
	}

	m.radios[radioID] = r

	return nil
}

// GetRadio returns the radio with hardcoded capabilities from ICD.
func (m *MockRadioManager) GetRadio(radioID string) (*radio.Radio, error) {
	radio, exists := m.radios[radioID]
	if !exists {
		return nil, command.ErrNotFound
	}

	// Create capabilities with hardcoded channels from ICD
	// 2.4 GHz band, 5 MHz spacing, 1-based indexing
	// Channel 1 = 2412 MHz, Channel 6 = 2437 MHz, Channel 11 = 2462 MHz
	capabilities := &adapter.RadioCapabilities{
		Channels: []adapter.Channel{
			{Index: 1, FrequencyMhz: 2412.0},
			{Index: 6, FrequencyMhz: 2437.0},
			{Index: 11, FrequencyMhz: 2462.0},
		},
	}

	// Set capabilities on the radio
	radio.Capabilities = capabilities

	return radio, nil
}

// SetActive sets the active radio ID.
func (m *MockRadioManager) SetActive(radioID string) error {
	m.activeRadioID = radioID
	return nil
}

// Compile-time assertion that MockRadioManager implements command.RadioManager
var _ command.RadioManager = (*MockRadioManager)(nil)
