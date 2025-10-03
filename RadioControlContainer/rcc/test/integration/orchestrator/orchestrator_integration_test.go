//go:build integration

package orchestrator

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/radio-control/rcc/internal/adapter"
	"github.com/radio-control/rcc/internal/command"
	"github.com/radio-control/rcc/internal/radio"
	"github.com/radio-control/rcc/internal/telemetry"
	"github.com/radio-control/rcc/test/fixtures"
	"github.com/radio-control/rcc/test/integration/fakes"
)

func TestOrchestratorIntegration_CommandValidation(t *testing.T) {
	// Test orchestrator command validation with proper setup
	cfg := fixtures.LoadTestConfig()
	telemetryHub := telemetry.NewHub(cfg)

	// Create real radio manager
	radioManager := radio.NewManager()

	// Create orchestrator with radio manager
	orchestrator := command.NewOrchestratorWithRadioManager(telemetryHub, cfg, radioManager)

	// Test with invalid radio ID (should get NOT_FOUND)
	invalidRadioID := "nonexistent-radio"
	err := orchestrator.SetChannel(context.Background(), invalidRadioID, 2412.0)
	if err == nil {
		t.Error("Expected error for invalid radio ID")
	}

	// Should get NOT_FOUND, not UNAVAILABLE
	if err != nil && !errors.Is(err, command.ErrNotFound) {
		t.Errorf("Expected command.ErrNotFound, got: %v", err)
	}

	// Test with valid radio ID but no adapter (should get UNAVAILABLE)
	validRadioID := "test-radio-001"

	// Create a fake adapter but don't set it as active
	fakeAdapter := fakes.NewFakeAdapter("test-radio-001")

	// Load capabilities for the radio
	err = radioManager.LoadCapabilities(validRadioID, fakeAdapter, 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to load capabilities: %v", err)
	}

	// Set radio as active
	err = radioManager.SetActive(validRadioID)
	if err != nil {
		t.Fatalf("Failed to set active radio: %v", err)
	}

	// Now test - should get UNAVAILABLE because no active adapter set
	err = orchestrator.SetChannel(context.Background(), validRadioID, 2412.0)
	if err == nil {
		t.Error("Expected error for radio without active adapter")
	}

	// Should get UNAVAILABLE for missing active adapter
	if err != nil && !errors.Is(err, adapter.ErrUnavailable) {
		t.Errorf("Expected adapter.ErrUnavailable, got: %v", err)
	}

	// Clean up
	telemetryHub.Stop()
}

func TestOrchestratorIntegration_PowerCommand(t *testing.T) {
	// Test power command validation with proper setup
	cfg := fixtures.LoadTestConfig()
	telemetryHub := telemetry.NewHub(cfg)

	// Create real radio manager
	radioManager := radio.NewManager()

	// Create orchestrator with radio manager
	orchestrator := command.NewOrchestratorWithRadioManager(telemetryHub, cfg, radioManager)

	// Test power command with valid radio but no adapter
	radioID := "test-radio-002"

	// Create a fake adapter but don't set it as active
	fakeAdapter := fakes.NewFakeAdapter("test-radio-002")

	// Load capabilities for the radio
	err := radioManager.LoadCapabilities(radioID, fakeAdapter, 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to load capabilities: %v", err)
	}

	// Set radio as active
	err = radioManager.SetActive(radioID)
	if err != nil {
		t.Fatalf("Failed to set active radio: %v", err)
	}

	// Test power command - should get UNAVAILABLE because no active adapter set
	err = orchestrator.SetPower(context.Background(), radioID, 5.0)
	if err == nil {
		t.Error("Expected error for radio without active adapter")
	}

	// Verify error type
	if err != nil && !errors.Is(err, adapter.ErrUnavailable) {
		t.Errorf("Expected adapter.ErrUnavailable, got: %v", err)
	}

	// Clean up
	telemetryHub.Stop()
}

func TestOrchestratorIntegration_ChannelByIndex(t *testing.T) {
	// Test channel by index command with proper setup
	cfg := fixtures.LoadTestConfig()
	telemetryHub := telemetry.NewHub(cfg)

	// Create real radio manager
	radioManager := radio.NewManager()

	// Create orchestrator with radio manager
	orchestrator := command.NewOrchestratorWithRadioManager(telemetryHub, cfg, radioManager)

	// Test with valid radio but no adapter
	radioID := "test-radio-003"

	// Create a fake adapter but don't set it as active
	fakeAdapter := fakes.NewFakeAdapter("test-radio-003")

	// Load capabilities for the radio
	err := radioManager.LoadCapabilities(radioID, fakeAdapter, 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to load capabilities: %v", err)
	}

	// Set radio as active
	err = radioManager.SetActive(radioID)
	if err != nil {
		t.Fatalf("Failed to set active radio: %v", err)
	}

	// Test channel by index - should get UNAVAILABLE because no active adapter set
	err = orchestrator.SetChannelByIndex(context.Background(), radioID, 6, radioManager)
	if err == nil {
		t.Error("Expected error for radio without active adapter")
	}

	// Verify error type
	if err != nil && !errors.Is(err, adapter.ErrUnavailable) {
		t.Errorf("Expected adapter.ErrUnavailable, got: %v", err)
	}

	// Clean up
	telemetryHub.Stop()
}

func TestOrchestratorIntegration_GetState(t *testing.T) {
	// Test get state command with proper setup
	cfg := fixtures.LoadTestConfig()
	telemetryHub := telemetry.NewHub(cfg)

	// Create real radio manager
	radioManager := radio.NewManager()

	// Create orchestrator with radio manager
	orchestrator := command.NewOrchestratorWithRadioManager(telemetryHub, cfg, radioManager)

	// Test with valid radio but no adapter
	radioID := "test-radio-004"

	// Create a fake adapter but don't set it as active
	fakeAdapter := fakes.NewFakeAdapter("test-radio-004")

	// Load capabilities for the radio
	err := radioManager.LoadCapabilities(radioID, fakeAdapter, 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to load capabilities: %v", err)
	}

	// Set radio as active
	err = radioManager.SetActive(radioID)
	if err != nil {
		t.Fatalf("Failed to set active radio: %v", err)
	}

	// Test get state - should get UNAVAILABLE because no active adapter set
	_, err = orchestrator.GetState(context.Background(), radioID)
	if err == nil {
		t.Error("Expected error for radio without active adapter")
	}

	// Verify error type
	if err != nil && !errors.Is(err, adapter.ErrUnavailable) {
		t.Errorf("Expected adapter.ErrUnavailable, got: %v", err)
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
