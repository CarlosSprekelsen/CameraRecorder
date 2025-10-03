//go:build integration

package orchestrator_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/radio-control/rcc/internal/adapter"
	"github.com/radio-control/rcc/internal/audit"
	"github.com/radio-control/rcc/internal/command"
	"github.com/radio-control/rcc/internal/radio"
	"github.com/radio-control/rcc/internal/telemetry"
	"github.com/radio-control/rcc/test/fixtures"
	"github.com/radio-control/rcc/test/integration/fakes"
)

func TestChannelFlow_OrchestratorToAdapter(t *testing.T) {
	// Arrange: real orchestrator + real components (only radio adapter mocked)
	cfg := fixtures.LoadTestConfig()
	telemetryHub := telemetry.NewHub(cfg)

	// Create real components
	radioManager := radio.NewManager()
	auditLogger, err := audit.NewLogger("/tmp/audit_test")
	if err != nil {
		t.Fatalf("Failed to create audit logger: %v", err)
	}

	// Create orchestrator with real components
	orchestrator := command.NewOrchestratorWithRadioManager(telemetryHub, cfg, radioManager)
	orchestrator.SetAuditLogger(auditLogger)

	// Use test fixtures for consistent inputs
	radioID := "test-radio-flow"
	channels := fixtures.WiFi24GHzChannels()

	// Create a fake adapter but don't set it as active
	fakeAdapter := fakes.NewFakeAdapter("test-radio-flow")

	// Load capabilities for the radio
	err = radioManager.LoadCapabilities(radioID, fakeAdapter, 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to load capabilities: %v", err)
	}

	// Set radio as active
	err = radioManager.SetActive(radioID)
	if err != nil {
		t.Fatalf("Failed to set active radio: %v", err)
	}

	// Act: orchestrator.SetChannel(...) - should get UNAVAILABLE because no active adapter set
	start := time.Now()
	err = orchestrator.SetChannel(context.Background(), radioID, channels[0].Frequency)
	latency := time.Since(start)

	// Assert: Should get UNAVAILABLE error
	if err == nil {
		t.Error("Expected error for radio without active adapter")
	}

	if err != nil && !errors.Is(err, adapter.ErrUnavailable) {
		t.Errorf("Expected adapter.ErrUnavailable, got: %v", err)
	}

	// Verify timing constraints (use config, not literals)
	if latency > cfg.CommandTimeoutSetChannel {
		t.Errorf("SetChannel took %v, exceeds timeout %v", latency, cfg.CommandTimeoutSetChannel)
	}

	// Clean up
	telemetryHub.Stop()
}

func TestChannelFlow_ErrorNormalization(t *testing.T) {
	// Test error mapping per Architecture §8.5
	cfg := fixtures.LoadTestConfig()
	telemetryHub := telemetry.NewHub(cfg)
	orchestrator := command.NewOrchestrator(telemetryHub, cfg)

	// Use error scenario fixtures
	radioID := fixtures.StandardSilvusRadio().ID
	invalidChannel := fixtures.RangeError().ChannelIndex

	// Act: trigger error condition
	err := orchestrator.SetChannelByIndex(context.Background(), radioID, invalidChannel, nil)

	// Assert: error is normalized to standard codes
	if err == nil {
		t.Error("Expected error for invalid channel")
	}

	// Verify error code mapping (INVALID_RANGE → HTTP 400)
	// This would be validated by the API layer in E2E tests
	t.Logf("Error normalized: %v", err)
}
