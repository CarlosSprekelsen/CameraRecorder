// Package adapter defines IRadioAdapter interface from Architecture §5.
//
// Requirements:
//   - Architecture §5: "Radio Adapters (per vendor): speak native IP protocols"
//   - Architecture §5: "IRadioAdapter: Stable API contract all adapters must implement"
//   - Architecture §8.5: "Error normalization to INVALID_RANGE, BUSY, UNAVAILABLE, INTERNAL"
package adapter

import (
	"context"
)

// RadioState represents the current state of a radio.
// Source: OpenAPI v1 §4.1 Radio model
type RadioState struct {
	PowerDbm     float64 `json:"powerDbm"`
	FrequencyMhz float64 `json:"frequencyMhz"`
}

// RadioCapabilities represents the capabilities of a radio.
// Source: OpenAPI v1 §4.1 Radio model
type RadioCapabilities struct {
	MinPowerDbm int       `json:"minPowerDbm"`
	MaxPowerDbm int       `json:"maxPowerDbm"`
	Channels    []Channel `json:"channels"`
}

// Channel represents a single channel mapping.
// Source: OpenAPI v1 §4.1 Radio model
type Channel struct {
	Index        int     `json:"index"`
	FrequencyMhz float64 `json:"frequencyMhz"`
}

// FrequencyProfile represents a supported frequency profile.
// Source: ICD §6.1.2 supported_frequency_profiles
type FrequencyProfile struct {
	Frequencies []float64 `json:"frequencies"`
	Bandwidth   float64   `json:"bandwidth"`
	AntennaMask int       `json:"antenna_mask"`
}

// IRadioAdapter defines the stable southbound adapter contract.
// Source: Architecture §5 - Radio Adapters
// Quote: "Stable API contract all adapters must implement"
type IRadioAdapter interface {
	// GetState returns the current radio state.
	// Source: ICD §6.1.1 freq (read), §6.1.3 power_dBm (read)
	GetState(ctx context.Context) (*RadioState, error)

	// SetPower sets the transmit power in dBm.
	// Source: ICD §6.1.3 power_dBm (set)
	// Params: dBm (0-39, accuracy 10-39)
	SetPower(ctx context.Context, dBm float64) error

	// SetFrequency sets the transmit frequency in MHz.
	// Source: ICD §6.1.1 freq (set)
	// Params: frequencyMhz (0.1 MHz resolution)
	// Side-effect: soft boot - driver/services reboot
	SetFrequency(ctx context.Context, frequencyMhz float64) error

	// ReadPowerActual reads the current power setting.
	// Source: ICD §6.1.3 power_dBm (read)
	ReadPowerActual(ctx context.Context) (float64, error)

	// SupportedFrequencyProfiles returns allowed frequency/bandwidth/antenna combinations.
	// Source: ICD §6.1.2 supported_frequency_profiles
	SupportedFrequencyProfiles(ctx context.Context) ([]FrequencyProfile, error)
}

// AdapterBase provides common functionality for adapter implementations.
// Source: Architecture §5
// Quote: "Radio Adapters (per vendor): speak native IP protocols"
type AdapterBase struct {
	// RadioID identifies the radio this adapter controls
	RadioID string

	// Model identifies the radio model
	Model string

	// Status indicates the current radio status
	Status string
}

// GetRadioID returns the radio identifier.
func (a *AdapterBase) GetRadioID() string {
	return a.RadioID
}

// GetModel returns the radio model.
func (a *AdapterBase) GetModel() string {
	return a.Model
}

// GetStatus returns the radio status.
func (a *AdapterBase) GetStatus() string {
	return a.Status
}

// SetStatus updates the radio status.
func (a *AdapterBase) SetStatus(status string) {
	a.Status = status
}
