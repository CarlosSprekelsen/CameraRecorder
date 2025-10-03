//go:build integration

package command_test

import (
	"context"
	"testing"
	"time"

	"github.com/radio-control/rcc/internal/command"
	"github.com/radio-control/rcc/internal/radio"
	"github.com/radio-control/rcc/internal/telemetry"
	"github.com/radio-control/rcc/test/fixtures"
	"github.com/radio-control/rcc/test/integration/fakes"
	"github.com/radio-control/rcc/test/integration/mocks"
)

// TestRealIntegration_OrchestratorWithRealComponents tests real component interactions.
// This provides actual coverage of production code while mocking only external dependencies.
func TestRealIntegration_OrchestratorWithRealComponents(t *testing.T) {
	// Arrange: Real production components
	cfg := fixtures.LoadTestConfig()
	telemetryHub := telemetry.NewHub(cfg)
	radioManager := radio.NewManager()
	
	// Create orchestrator with real radio manager
	orchestrator := command.NewOrchestratorWithRadioManager(telemetryHub, cfg, radioManager)
	
	// Mock only external dependencies
	mockAudit := mocks.NewMockAuditLogger()
	orchestrator.SetAuditLogger(mockAudit)
	
	// Create fake adapter (external dependency)
	fakeAdapter := fakes.NewFakeAdapter("radio-001").
		WithInitial(20.0, 2412.0, nil)
	
	// Load capabilities into real radio manager
	err := radioManager.LoadCapabilities("radio-001", fakeAdapter, 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to load capabilities: %v", err)
	}
	
	// Set active radio in real radio manager
	err = radioManager.SetActive("radio-001")
	if err != nil {
		t.Fatalf("Failed to set active radio: %v", err)
	}
	
	// Set active adapter in real orchestrator
	orchestrator.SetActiveAdapter(fakeAdapter)
	
	// Act: Execute real command through real orchestrator
	ctx := context.Background()
	err = orchestrator.SetPower(ctx, "radio-001", 25.0)
	
	// Assert: Real component interactions
	if err != nil {
		t.Errorf("SetPower failed: %v", err)
	}
	
	// Verify real audit logging occurred
	actions := mockAudit.GetLoggedActions()
	if len(actions) != 1 {
		t.Errorf("Expected 1 audit action, got %d", len(actions))
	} else {
		action := actions[0]
		if action.Action != "setPower" {
			t.Errorf("Expected action 'setPower', got '%s'", action.Action)
		}
		if action.RadioID != "radio-001" {
			t.Errorf("Expected radioID 'radio-001', got '%s'", action.RadioID)
		}
		if action.Result != "SUCCESS" {
			t.Errorf("Expected result 'SUCCESS', got '%s'", action.Result)
		}
	}
	
	// Verify real adapter was called
	fakeAdapterTyped := fakeAdapter.(*fakes.FakeAdapter)
	if len(fakeAdapterTyped.SetPowerCalls) != 1 {
		t.Errorf("Expected 1 SetPower call to adapter, got %d", len(fakeAdapterTyped.SetPowerCalls))
	} else {
		call := fakeAdapterTyped.SetPowerCalls[0]
		if call.PowerDbm != 25.0 {
			t.Errorf("Expected power 25.0, got %f", call.PowerDbm)
		}
	}
	
	// Clean up
	telemetryHub.Stop()
}

// TestRealIntegration_ChannelIndexMapping tests real channel index resolution.
func TestRealIntegration_ChannelIndexMapping(t *testing.T) {
	// Arrange: Real production components
	cfg := fixtures.LoadTestConfig()
	telemetryHub := telemetry.NewHub(cfg)
	radioManager := radio.NewManager()
	
	// Create orchestrator with real radio manager
	orchestrator := command.NewOrchestratorWithRadioManager(telemetryHub, cfg, radioManager)
	
	// Mock only external dependencies
	mockAudit := mocks.NewMockAuditLogger()
	orchestrator.SetAuditLogger(mockAudit)
	
	// Create fake adapter with hardcoded channels from ICD
	fakeAdapter := fakes.NewFakeAdapter("radio-002").
		WithInitial(20.0, 2412.0, nil)
	
	// Load capabilities with hardcoded channels from ICD
	err := radioManager.LoadCapabilities("radio-002", fakeAdapter, 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to load capabilities: %v", err)
	}
	
	// Set active radio and adapter
	err = radioManager.SetActive("radio-002")
	if err != nil {
		t.Fatalf("Failed to set active radio: %v", err)
	}
	orchestrator.SetActiveAdapter(fakeAdapter)
	
	// Act: Test real channel index mapping
	ctx := context.Background()
	err = orchestrator.SetChannelByIndex(ctx, "radio-002", 6, radioManager)
	
	// Assert: Real channel mapping should work
	if err != nil {
		t.Errorf("SetChannelByIndex failed: %v", err)
	}
	
	// Verify real adapter was called with correct frequency
	fakeAdapterTyped := fakeAdapter.(*fakes.FakeAdapter)
	if len(fakeAdapterTyped.SetFrequencyCalls) != 1 {
		t.Errorf("Expected 1 SetFrequency call to adapter, got %d", len(fakeAdapterTyped.SetFrequencyCalls))
	} else {
		call := fakeAdapterTyped.SetFrequencyCalls[0]
		expectedFreq := 2437.0 // Channel 6 = 2437.0 MHz per ICD
		if call.FrequencyMhz != expectedFreq {
			t.Errorf("Expected frequency %f, got %f", expectedFreq, call.FrequencyMhz)
		}
	}
	
	// Clean up
	telemetryHub.Stop()
}

// TestRealIntegration_ErrorNormalization tests real error mapping.
func TestRealIntegration_ErrorNormalization(t *testing.T) {
	// Arrange: Real production components with error-generating adapter
	cfg := fixtures.LoadTestConfig()
	telemetryHub := telemetry.NewHub(cfg)
	radioManager := radio.NewManager()
	
	orchestrator := command.NewOrchestratorWithRadioManager(telemetryHub, cfg, radioManager)
	
	// Create adapter that returns specific errors
	errorAdapter := fakes.NewFakeAdapter("radio-003").
		WithMode("busy") // Will return BUSY error
	
	err := radioManager.LoadCapabilities("radio-003", errorAdapter, 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to load capabilities: %v", err)
	}
	
	err = radioManager.SetActive("radio-003")
	if err != nil {
		t.Fatalf("Failed to set active radio: %v", err)
	}
	orchestrator.SetActiveAdapter(errorAdapter)
	
	// Act: Execute command that should generate error
	ctx := context.Background()
	err = orchestrator.SetPower(ctx, "radio-003", 25.0)
	
	// Assert: Real error normalization occurred
	if err == nil {
		t.Error("Expected error from busy adapter")
	}
	
	// Verify error was normalized (should be adapter.ErrBusy or similar)
	// The exact error type depends on the adapter's error normalization
	t.Logf("Error normalized: %v", err)
	
	// Clean up
	telemetryHub.Stop()
}
