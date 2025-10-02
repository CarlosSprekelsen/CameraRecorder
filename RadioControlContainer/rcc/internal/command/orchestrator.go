// Package command implements CommandOrchestrator from Architecture §5.
//
// Requirements:
//   - Architecture §5: "Route validated API intents to the active adapter; emit events; write audit records."
//
// Source: Architecture §5
// Quote: "CommandOrchestrator: Route validated API intents to the active adapter; emit events; write audit records."
package command

import (
	"context"
	"fmt"
	"time"

	"github.com/radio-control/rcc/internal/adapter"
	"github.com/radio-control/rcc/internal/config"
	"github.com/radio-control/rcc/internal/telemetry"
)

// Orchestrator routes validated API intents to the active adapter.
// Source: Architecture §5
// Quote: "Route validated API intents to the active adapter; emit events; write audit records."
type Orchestrator struct {
	// Active radio adapter
	activeAdapter adapter.IRadioAdapter

	// Telemetry hub for event publishing
	telemetryHub *telemetry.Hub

	// Configuration for validation
	config *config.TimingConfig

	// Audit logger (to be implemented)
	auditLogger AuditLogger
}

// AuditLogger interface for writing audit records.
type AuditLogger interface {
	LogAction(ctx context.Context, action string, radioID string, result string, latency time.Duration)
}

// NewOrchestrator creates a new command orchestrator.
func NewOrchestrator(telemetryHub *telemetry.Hub, timingConfig *config.TimingConfig) *Orchestrator {
	return &Orchestrator{
		telemetryHub: telemetryHub,
		config:       timingConfig,
	}
}

// SetActiveAdapter sets the active radio adapter.
func (o *Orchestrator) SetActiveAdapter(adapter adapter.IRadioAdapter) {
	o.activeAdapter = adapter
}

// SetPower sets the transmit power for the active radio.
// Source: OpenAPI v1 §3.6
// Quote: "Set TX power for a radio (dBm)"
func (o *Orchestrator) SetPower(ctx context.Context, radioID string, dBm int) error {
	start := time.Now()

	// Validate power range
	if err := o.validatePowerRange(dBm); err != nil {
		o.logAudit(ctx, "setPower", radioID, "INVALID_RANGE", time.Since(start))
		return err
	}

	// Check if adapter is available
	if o.activeAdapter == nil {
		o.logAudit(ctx, "setPower", radioID, "UNAVAILABLE", time.Since(start))
		return fmt.Errorf("no active radio adapter")
	}

	// Execute command with timeout
	timeout := o.config.CommandTimeoutSetPower
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	err := o.activeAdapter.SetPower(ctx, dBm)
	latency := time.Since(start)

	if err != nil {
		// Map adapter error to normalized code
		normalizedErr := adapter.NormalizeVendorError(err, nil)
		o.logAudit(ctx, "setPower", radioID, "ERROR", latency)

		// Publish fault event
		o.publishFaultEvent(radioID, normalizedErr, "Failed to set power")

		return normalizedErr
	}

	// Log successful action
	o.logAudit(ctx, "setPower", radioID, "SUCCESS", latency)

	// Publish power changed event
	o.publishPowerChangedEvent(radioID, dBm)

	return nil
}

// SetChannel sets the channel for the active radio by frequency or index.
// Source: OpenAPI v1 §3.8
// Quote: "Set radio channel by UI channel index or by frequency"
func (o *Orchestrator) SetChannel(ctx context.Context, radioID string, frequencyMhz float64) error {
	start := time.Now()

	// Validate frequency range
	if err := o.validateFrequencyRange(frequencyMhz); err != nil {
		o.logAudit(ctx, "setChannel", radioID, "INVALID_RANGE", time.Since(start))
		return err
	}

	// Check if adapter is available
	if o.activeAdapter == nil {
		o.logAudit(ctx, "setChannel", radioID, "UNAVAILABLE", time.Since(start))
		return fmt.Errorf("no active radio adapter")
	}

	// Execute command with timeout
	timeout := o.config.CommandTimeoutSetChannel
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	err := o.activeAdapter.SetFrequency(ctx, frequencyMhz)
	latency := time.Since(start)

	if err != nil {
		// Map adapter error to normalized code
		normalizedErr := adapter.NormalizeVendorError(err, nil)
		o.logAudit(ctx, "setChannel", radioID, "ERROR", latency)

		// Publish fault event
		o.publishFaultEvent(radioID, normalizedErr, "Failed to set channel")

		return normalizedErr
	}

	// Log successful action
	o.logAudit(ctx, "setChannel", radioID, "SUCCESS", latency)

	// Publish channel changed event
	o.publishChannelChangedEvent(radioID, frequencyMhz, 0) // channelIndex will be derived later

	return nil
}

// SelectRadio selects the active radio for subsequent operations.
// Source: OpenAPI v1 §3.3
// Quote: "Select the active radio for subsequent operations"
func (o *Orchestrator) SelectRadio(ctx context.Context, radioID string) error {
	start := time.Now()

	// Validate radio ID
	if radioID == "" {
		o.logAudit(ctx, "selectRadio", radioID, "INVALID_RANGE", time.Since(start))
		return fmt.Errorf("radio ID cannot be empty")
	}

	// Check if adapter is available
	if o.activeAdapter == nil {
		o.logAudit(ctx, "selectRadio", radioID, "UNAVAILABLE", time.Since(start))
		return fmt.Errorf("no active radio adapter")
	}

	// Execute command with timeout
	timeout := o.config.CommandTimeoutSelectRadio
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// For now, just validate the adapter is responsive
	_, err := o.activeAdapter.GetState(ctx)
	latency := time.Since(start)

	if err != nil {
		// Map adapter error to normalized code
		normalizedErr := adapter.NormalizeVendorError(err, nil)
		o.logAudit(ctx, "selectRadio", radioID, "ERROR", latency)

		// Publish fault event
		o.publishFaultEvent(radioID, normalizedErr, "Failed to select radio")

		return normalizedErr
	}

	// Log successful action
	o.logAudit(ctx, "selectRadio", radioID, "SUCCESS", latency)

	// Publish state event to confirm selection
	o.publishStateEvent(radioID)

	return nil
}

// GetState retrieves the current state of the active radio.
// Source: OpenAPI v1 §3.4
func (o *Orchestrator) GetState(ctx context.Context, radioID string) (*adapter.RadioState, error) {
	start := time.Now()

	// Check if adapter is available
	if o.activeAdapter == nil {
		o.logAudit(ctx, "getState", radioID, "UNAVAILABLE", time.Since(start))
		return nil, fmt.Errorf("no active radio adapter")
	}

	// Execute command with timeout
	timeout := o.config.CommandTimeoutGetState
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	state, err := o.activeAdapter.GetState(ctx)
	latency := time.Since(start)

	if err != nil {
		// Map adapter error to normalized code
		normalizedErr := adapter.NormalizeVendorError(err, nil)
		o.logAudit(ctx, "getState", radioID, "ERROR", latency)

		// Publish fault event
		o.publishFaultEvent(radioID, normalizedErr, "Failed to get state")

		return nil, normalizedErr
	}

	// Log successful action
	o.logAudit(ctx, "getState", radioID, "SUCCESS", latency)

	return state, nil
}

// validatePowerRange validates the power range.
// Source: OpenAPI v1 §3.6
// Quote: "Range: 0..39 (accuracy typically 10..39)"
func (o *Orchestrator) validatePowerRange(dBm int) error {
	if dBm < 0 || dBm > 39 {
		return fmt.Errorf("power must be between 0 and 39 dBm, got %d", dBm)
	}
	return nil
}

// validateFrequencyRange validates the frequency range.
// Source: OpenAPI v1 §3.8
// Quote: "Frequency must be within the radio's allowed ranges"
func (o *Orchestrator) validateFrequencyRange(frequencyMhz float64) error {
	// Basic frequency validation - more sophisticated validation will be added later
	// with derived channel maps
	if frequencyMhz <= 0 {
		return fmt.Errorf("frequency must be positive, got %f", frequencyMhz)
	}

	// Check against reasonable frequency ranges (will be enhanced with channel maps)
	if frequencyMhz < 100 || frequencyMhz > 6000 {
		return fmt.Errorf("frequency must be between 100 and 6000 MHz, got %f", frequencyMhz)
	}

	return nil
}

// publishPowerChangedEvent publishes a power changed event.
// Source: Telemetry SSE v1 §2.2d
func (o *Orchestrator) publishPowerChangedEvent(radioID string, powerDbm int) {
	if o.telemetryHub == nil {
		return // Skip if no telemetry hub
	}

	event := telemetry.Event{
		Type: "powerChanged",
		Data: map[string]interface{}{
			"radioId":  radioID,
			"powerDbm": powerDbm,
			"ts":       time.Now().UTC().Format(time.RFC3339),
		},
	}

	o.telemetryHub.PublishRadio(radioID, event)
}

// publishChannelChangedEvent publishes a channel changed event.
// Source: Telemetry SSE v1 §2.2c
func (o *Orchestrator) publishChannelChangedEvent(radioID string, frequencyMhz float64, channelIndex int) {
	if o.telemetryHub == nil {
		return // Skip if no telemetry hub
	}

	event := telemetry.Event{
		Type: "channelChanged",
		Data: map[string]interface{}{
			"radioId":      radioID,
			"frequencyMhz": frequencyMhz,
			"channelIndex": channelIndex,
			"ts":           time.Now().UTC().Format(time.RFC3339),
		},
	}

	o.telemetryHub.PublishRadio(radioID, event)
}

// publishStateEvent publishes a state event.
// Source: Telemetry SSE v1 §2.2b
func (o *Orchestrator) publishStateEvent(radioID string) {
	if o.telemetryHub == nil {
		return // Skip if no telemetry hub
	}

	event := telemetry.Event{
		Type: "state",
		Data: map[string]interface{}{
			"radioId": radioID,
			"status":  "online",
			"ts":      time.Now().UTC().Format(time.RFC3339),
		},
	}

	o.telemetryHub.PublishRadio(radioID, event)
}

// publishFaultEvent publishes a fault event.
// Source: Telemetry SSE v1 §2.2e
func (o *Orchestrator) publishFaultEvent(radioID string, err error, message string) {
	if o.telemetryHub == nil {
		return // Skip if no telemetry hub
	}

	event := telemetry.Event{
		Type: "fault",
		Data: map[string]interface{}{
			"radioId": radioID,
			"code":    err.Error(),
			"message": message,
			"ts":      time.Now().UTC().Format(time.RFC3339),
		},
	}

	o.telemetryHub.PublishRadio(radioID, event)
}

// logAudit logs an audit record for a command action.
// Source: Architecture §8.6
// Quote: "Structured audit logs per Architecture §8.6 schema"
func (o *Orchestrator) logAudit(ctx context.Context, action, radioID, result string, latency time.Duration) {
	if o.auditLogger != nil {
		o.auditLogger.LogAction(ctx, action, radioID, result, latency)
	}
}

// SetAuditLogger sets the audit logger.
func (o *Orchestrator) SetAuditLogger(logger AuditLogger) {
	o.auditLogger = logger
}
