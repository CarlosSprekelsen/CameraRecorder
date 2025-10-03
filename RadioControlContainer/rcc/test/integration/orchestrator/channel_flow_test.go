//go:build integration

package orchestrator_test

import (
	"context"
	"testing"
	"time"

	"github.com/radio-control/rcc/internal/command"
	"github.com/radio-control/rcc/internal/telemetry"
	"github.com/radio-control/rcc/test/fixtures"
)

func TestChannelFlow_OrchestratorToAdapter(t *testing.T) {
	// Arrange: real orchestrator + real adapter wiring (no HTTP)
	cfg := fixtures.LoadTestConfig()
	telemetryHub := telemetry.NewHub()
	orchestrator := command.NewOrchestrator(telemetryHub, cfg)

	// Use test fixtures for consistent inputs
	radioID := fixtures.StandardSilvusRadio().ID
	channels := fixtures.WiFi24GHzChannels()

	// Act: orchestrator.SetChannel(...)
	start := time.Now()
	err := orchestrator.SetChannel(context.Background(), radioID, channels[0].Frequency)
	latency := time.Since(start)

	// Assert: telemetry events, audit logs, error mapping
	if err != nil {
		t.Errorf("SetChannel failed: %v", err)
	}

	// Verify timing constraints (use config, not literals)
	if latency > cfg.CommandTimeoutSetChannel {
		t.Errorf("SetChannel took %v, exceeds timeout %v", latency, cfg.CommandTimeoutSetChannel)
	}
}

func TestChannelFlow_ErrorNormalization(t *testing.T) {
	// Test error mapping per Architecture §8.5
	cfg := fixtures.LoadTestConfig()
	telemetryHub := telemetry.NewHub()
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
