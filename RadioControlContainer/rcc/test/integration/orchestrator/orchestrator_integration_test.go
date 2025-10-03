//go:build integration

package orchestrator

import (
	"context"
	"testing"
	"time"

	"github.com/radio-control/rcc/internal/command"
	"github.com/radio-control/rcc/internal/telemetry"
	"github.com/radio-control/rcc/test/fixtures"
)

func TestOrchestratorIntegration_CommandValidation(t *testing.T) {
	// Test orchestrator command validation
	cfg := fixtures.LoadTestConfig()
	telemetryHub := telemetry.NewHub(cfg)
	orchestrator := command.NewOrchestrator(telemetryHub, cfg)

	// Test with invalid radio ID
	invalidRadioID := "nonexistent-radio"
	err := orchestrator.SetChannel(context.Background(), invalidRadioID, 2412.0)
	if err == nil {
		t.Error("Expected error for invalid radio ID")
	}

	// Test with valid radio ID but no adapter (expected behavior)
	validRadioID := fixtures.StandardSilvusRadio().ID
	err = orchestrator.SetChannel(context.Background(), validRadioID, 2412.0)
	if err == nil {
		t.Error("Expected error for radio without adapter")
	}

	// Verify error message contains expected text
	if err != nil && err.Error() != "no active radio adapter" {
		t.Errorf("Unexpected error message: %v", err)
	}

	// Clean up
	telemetryHub.Stop()
}

func TestOrchestratorIntegration_PowerCommand(t *testing.T) {
	// Test power command validation
	cfg := fixtures.LoadTestConfig()
	telemetryHub := telemetry.NewHub(cfg)
	orchestrator := command.NewOrchestrator(telemetryHub, cfg)

	// Test power command with valid radio but no adapter
	radioID := fixtures.StandardSilvusRadio().ID
	err := orchestrator.SetPower(context.Background(), radioID, 5)
	if err == nil {
		t.Error("Expected error for radio without adapter")
	}

	// Verify error message
	if err != nil && err.Error() != "no active radio adapter" {
		t.Errorf("Unexpected error message: %v", err)
	}

	// Clean up
	telemetryHub.Stop()
}

func TestOrchestratorIntegration_ChannelByIndex(t *testing.T) {
	// Test channel by index command
	cfg := fixtures.LoadTestConfig()
	telemetryHub := telemetry.NewHub(cfg)
	orchestrator := command.NewOrchestrator(telemetryHub, cfg)

	// Test with valid radio but no adapter
	radioID := fixtures.StandardSilvusRadio().ID
	err := orchestrator.SetChannelByIndex(context.Background(), radioID, 6, nil)
	if err == nil {
		t.Error("Expected error for radio without adapter")
	}

	// Verify error message
	if err != nil && err.Error() != "no active radio adapter" {
		t.Errorf("Unexpected error message: %v", err)
	}

	// Clean up
	telemetryHub.Stop()
}

func TestOrchestratorIntegration_GetState(t *testing.T) {
	// Test get state command
	cfg := fixtures.LoadTestConfig()
	telemetryHub := telemetry.NewHub(cfg)
	orchestrator := command.NewOrchestrator(telemetryHub, cfg)

	// Test with valid radio but no adapter
	radioID := fixtures.StandardSilvusRadio().ID
	_, err := orchestrator.GetState(context.Background(), radioID)
	if err == nil {
		t.Error("Expected error for radio without adapter")
	}

	// Verify error message
	if err != nil && err.Error() != "no active radio adapter" {
		t.Errorf("Unexpected error message: %v", err)
	}

	// Clean up
	telemetryHub.Stop()
}

func TestOrchestratorIntegration_TimingConstraints(t *testing.T) {
	// Test timing constraints
	cfg := fixtures.LoadTestConfig()
	telemetryHub := telemetry.NewHub(cfg)
	orchestrator := command.NewOrchestrator(telemetryHub, cfg)

	// Test command timeout behavior
	radioID := fixtures.StandardSilvusRadio().ID
	start := time.Now()

	// Use context with timeout shorter than command timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := orchestrator.SetChannel(ctx, radioID, 2412.0)
	duration := time.Since(start)

	// Should fail due to context timeout or no adapter
	if err == nil {
		t.Error("Expected error due to timeout or no adapter")
	}

	// Should complete within reasonable time (not hang)
	if duration > 1*time.Second {
		t.Errorf("Command took too long: %v", duration)
	}

	// Clean up
	telemetryHub.Stop()
}
