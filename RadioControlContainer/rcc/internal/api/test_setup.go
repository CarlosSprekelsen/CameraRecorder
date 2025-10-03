package api

import (
	"testing"
	"time"

	"github.com/radio-control/rcc/internal/adapter"
	"github.com/radio-control/rcc/internal/adapter/silvusmock"
	"github.com/radio-control/rcc/internal/audit"
	"github.com/radio-control/rcc/internal/command"
	"github.com/radio-control/rcc/internal/config"
	"github.com/radio-control/rcc/internal/radio"
	"github.com/radio-control/rcc/internal/telemetry"
)

// radioManagerWrapper wraps radio.Manager to implement the RadioManager interface
type radioManagerWrapper struct {
	rm *radio.Manager
}

func (w *radioManagerWrapper) GetRadio(radioID string) (interface{}, error) {
	radio, err := w.rm.GetRadio(radioID)
	if err != nil {
		return nil, err
	}

	// Convert radio to map[string]interface{} as expected by orchestrator
	radioMap := map[string]interface{}{
		"id":           radio.ID,
		"model":        radio.Model,
		"status":       radio.Status,
		"capabilities": radio.Capabilities,
		"state":        radio.State,
		"lastSeen":     radio.LastSeen,
	}

	return radioMap, nil
}

// setupAPITest creates a fully wired API test environment with SilvusMock
func setupAPITest(t *testing.T) (*Server, *radio.Manager, *command.Orchestrator, *silvusmock.SilvusMock) {
	cfg := config.LoadCBTimingBaseline()
	hub := telemetry.NewHub(cfg)
	t.Cleanup(func() { hub.Stop() })

	// Create radio manager and register SilvusMock
	rm := radio.NewManager()
	adapter := silvusmock.NewSilvusMock("silvus-001", []adapter.Channel{
		{Index: 1, FrequencyMhz: 2412},
		{Index: 6, FrequencyMhz: 2437},
		{Index: 11, FrequencyMhz: 2462},
	})

	// Load capabilities to register adapter
	rm.LoadCapabilities("silvus-001", adapter, 5*time.Second)
	rm.SetActive("silvus-001")

	// Create orchestrator and set radio manager with wrapper
	orch := command.NewOrchestrator(hub, cfg)
	radioManagerWrapper := &radioManagerWrapper{rm: rm}
	orch.SetRadioManager(radioManagerWrapper)

	// Set up audit logger
	auditLogger, err := audit.NewLogger(t.TempDir())
	if err != nil {
		t.Fatalf("Failed to create audit logger: %v", err)
	}
	orch.SetAuditLogger(auditLogger)

	// Set the active adapter on the orchestrator
	orch.SetActiveAdapter(adapter)

	// Create API server
	server := NewServer(hub, orch, rm, 30*time.Second, 30*time.Second, 120*time.Second)

	return server, rm, orch, adapter
}

// setupAPITestWithFault creates API test environment with specific fault mode
func setupAPITestWithFault(t *testing.T, faultMode string) (*Server, *radio.Manager, *command.Orchestrator, *silvusmock.SilvusMock) {
	server, rm, orch, adapter := setupAPITest(t)

	// Set fault mode if specified
	if faultMode != "" {
		adapter.SetFaultMode(faultMode)
	}

	return server, rm, orch, adapter
}
