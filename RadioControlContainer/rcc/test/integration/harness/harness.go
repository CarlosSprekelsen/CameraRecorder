//go:build integration

// Package harness provides a minimal in-process integration test harness.
// Boundary: command+radio+adapter+telemetry+audit; no HTTP; deterministic.
package harness

import (
	"os"
	"time"

	"github.com/radio-control/rcc/internal/adapter"
	"github.com/radio-control/rcc/internal/audit"
	"github.com/radio-control/rcc/internal/command"
	"github.com/radio-control/rcc/internal/config"
	"github.com/radio-control/rcc/internal/radio"
	"github.com/radio-control/rcc/internal/telemetry"
	"github.com/radio-control/rcc/test/integration/fixtures"
	"github.com/radio-control/rcc/test/integration/fakes"
)

// BuildCommandStack wires real implementations via public constructors only.
// Returns components via their public interfaces/ports.
func BuildCommandStack(cfg *config.TimingConfig, seed Radios) (orch command.OrchestratorPort, rm *radio.Manager, tele *telemetry.Hub, auditSink *audit.Logger, cleanup func()) {
	// Create telemetry hub
	tele = telemetry.NewHub(cfg)

	// Create audit logger in temp dir
	tmpDir, err := os.MkdirTemp("", "integration-audit-*")
	if err != nil {
		panic("failed to create temp dir for audit logs: " + err.Error())
	}
	auditSink, err = audit.NewLogger(tmpDir)
	if err != nil {
		panic("failed to create audit logger: " + err.Error())
	}

	// Create radio manager
	rm = radio.NewManager()

	// Create orchestrator with radio manager
	orchestrator := command.NewOrchestratorWithRadioManager(tele, cfg, rm)
	orchestrator.SetAuditLogger(auditSink)

	// Seed radios if provided
	if seed != nil {
		for id, fakeAdapter := range seed {
			if err := SeedRadios(rm, id, fakeAdapter); err != nil {
				panic("failed to seed radio " + id + ": " + err.Error())
			}
			orchestrator.SetActiveAdapter(fakeAdapter)
		}
	}

	cleanup = func() {
		tele.Stop()
		auditSink.Close()
		os.RemoveAll(tmpDir)
	}

	return orchestrator, rm, tele, auditSink, cleanup
}

// Radios represents a collection of radios to seed for testing.
type Radios map[string]adapter.IRadioAdapter

// SeedRadios registers a fake adapter under the given radio ID via manager port.
func SeedRadios(rm *radio.Manager, id string, fakeAdapter adapter.IRadioAdapter) error {
	// Use the public LoadCapabilities method to register the adapter
	// This avoids peeking into radio manager internals
	return rm.LoadCapabilities(id, fakeAdapter, 5*time.Second)
}

// BuildTestStack creates a complete test stack with fake adapters and fixtures.
func BuildTestStack() (orch command.OrchestratorPort, rm *radio.Manager, tele *telemetry.Hub, auditSink *audit.Logger, clock *fixtures.ManualClock, correlationIDGen *fixtures.CorrelationIDGenerator, cleanup func()) {
	// Create test fixtures
	clock = fixtures.NewManualClock()
	correlationIDGen = fixtures.NewCorrelationIDGenerator()
	
	// Create config with test timing
	cfg := config.LoadCBTimingBaseline()
	
	// Create fake adapters
	fakeAdapter := fakes.NewFakeAdapter("fake-001").
		WithInitial(20.0, 2412.0, []adapter.Channel{
			{Index: 1, FrequencyMhz: 2412.0},
			{Index: 6, FrequencyMhz: 2437.0},
			{Index: 11, FrequencyMhz: 2462.0},
		})
	
	seedRadios := Radios{
		"fake-001": fakeAdapter,
	}
	
	// Build the command stack
	orch, rm, tele, auditSink, cleanup = BuildCommandStack(cfg, seedRadios)
	
	return
}
