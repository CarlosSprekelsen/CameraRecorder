//go:build integration

// Package harness provides a minimal in-process integration test harness.
// Boundary: command+radio+adapter+telemetry+audit; no HTTP; deterministic.
package harness

import (
	"testing"
	"time"

	"github.com/radio-control/rcc/internal/adapter"
	"github.com/radio-control/rcc/internal/command"
	"github.com/radio-control/rcc/internal/telemetry"
	"github.com/radio-control/rcc/test/integration/fakes"
	"github.com/radio-control/rcc/test/integration/fixtures"
	"github.com/radio-control/rcc/test/integration/mocks"
)

// BuildCommandStack wires real implementations via public constructors only.
// Returns components via their public interfaces/ports with automatic cleanup.
func BuildCommandStack(t *testing.T, seed Radios) (orch command.OrchestratorPort, rm *mocks.MockRadioManager, tele *telemetry.Hub, mockAudit *mocks.MockAuditLogger, adapter adapter.IRadioAdapter) {
	// Create test config
	cfg := fixtures.TestTimingConfig()

	// Create telemetry hub
	tele = telemetry.NewHub(cfg)

	// Register cleanup for telemetry hub
	t.Cleanup(func() {
		tele.Stop()
	})

	// Create mock audit logger (no filesystem access)
	mockAudit = mocks.NewMockAuditLogger()

	// Create mock radio manager for test isolation
	rm = mocks.NewMockRadioManager()

	// Create orchestrator with real telemetry hub and mock radio manager
	orchestrator := command.NewOrchestratorWithRadioManager(tele, cfg, rm)
	orchestrator.SetAuditLogger(mockAudit)

	// Seed radios if provided and return the first adapter
	if seed != nil {
		for id, fakeAdapter := range seed {
			if err := rm.LoadCapabilities(id, fakeAdapter, 5*time.Second); err != nil {
				t.Fatalf("Failed to seed radio %s: %v", id, err)
			}
			orchestrator.SetActiveAdapter(fakeAdapter)
			adapter = fakeAdapter // Return the adapter for test verification
		}
	}

	return orchestrator, rm, tele, mockAudit, adapter
}

// Radios represents a collection of radios to seed for testing.
type Radios map[string]adapter.IRadioAdapter

// BuildTestStack creates a complete test stack with fake adapters and fixtures.
func BuildTestStack(t *testing.T) (orch command.OrchestratorPort, rm *mocks.MockRadioManager, tele *telemetry.Hub, mockAudit *mocks.MockAuditLogger, adapter adapter.IRadioAdapter) {
	// Create fake adapters
	fakeAdapter := fakes.NewFakeAdapter("fake-001").
		WithInitial(20.0, 2412.0, nil) // No channels needed for basic tests

	seedRadios := Radios{
		"fake-001": fakeAdapter,
	}

	// Build the command stack
	return BuildCommandStack(t, seedRadios)
}
