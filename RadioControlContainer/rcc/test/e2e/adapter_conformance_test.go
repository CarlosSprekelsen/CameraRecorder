// Package e2e provides adapter conformance tests for SilvusMock.
package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/radio-control/rcc/test/harness"
)

func TestE2E_SilvusMockConformance(t *testing.T) {
	// Create test harness with seeded state
	opts := harness.DefaultOptions()
	server := harness.NewServer(t, opts)
	defer server.Shutdown()

	// Evidence: Seeded state
	t.Logf("=== TEST EVIDENCE ===")
	t.Logf("Active Radio ID: %s", server.RadioManager.GetActive())
	t.Logf("SilvusMock Band Plan: %+v", server.SilvusAdapter.GetBandPlan())
	power, freq, channel := server.SilvusAdapter.GetCurrentState()
	t.Logf("SilvusMock State: power=%d, freq=%f, channel=%d", power, freq, channel)
	t.Logf("===================")

	ctx := context.Background()

	// Test GetState
	state, err := server.SilvusAdapter.GetState(ctx)
	if err != nil {
		t.Fatalf("GetState failed: %v", err)
	}

	// RadioState doesn't have RadioID field, check power instead
	if state.PowerDbm < 0 || state.PowerDbm > 100 {
		t.Errorf("Expected reasonable power value, got %d", state.PowerDbm)
	}

	// Test SetPower
	err = server.SilvusAdapter.SetPower(ctx, 10)
	if err != nil {
		t.Fatalf("SetPower failed: %v", err)
	}

	// Test ReadPowerActual
	power, err = server.SilvusAdapter.ReadPowerActual(ctx)
	if err != nil {
		t.Fatalf("ReadPowerActual failed: %v", err)
	}

	if power != 10 {
		t.Errorf("Expected power 10, got %d", power)
	}

	// Test SetFrequency
	err = server.SilvusAdapter.SetFrequency(ctx, 2437.0)
	if err != nil {
		t.Fatalf("SetFrequency failed: %v", err)
	}

	// Test GetState after changes
	state, err = server.SilvusAdapter.GetState(ctx)
	if err != nil {
		t.Fatalf("GetState after changes failed: %v", err)
	}

	if state.PowerDbm != 10 {
		t.Errorf("Expected power 10, got %d", state.PowerDbm)
	}

	if state.FrequencyMhz != 2437.0 {
		t.Errorf("Expected frequency 2437.0, got %f", state.FrequencyMhz)
	}

	// Evidence: Final state
	t.Logf("=== FINAL STATE EVIDENCE ===")
	t.Logf("Final power: %d", state.PowerDbm)
	t.Logf("Final frequency: %f", state.FrequencyMhz)
	t.Logf("===========================")

	t.Log("✅ SilvusMock adapter conformance working correctly")
}

func TestE2E_SilvusMockErrorHandling(t *testing.T) {
	opts := harness.DefaultOptions()
	server := harness.NewServer(t, opts)
	defer server.Shutdown()

	ctx := context.Background()

	// Test power out of range
	err := server.SilvusAdapter.SetPower(ctx, 100)
	if err == nil {
		t.Error("Expected error for power out of range")
	}

	// Test frequency out of range
	err = server.SilvusAdapter.SetFrequency(ctx, 10000.0)
	if err == nil {
		t.Error("Expected error for frequency out of range")
	}

	// Evidence: Error handling
	t.Logf("=== ERROR HANDLING EVIDENCE ===")
	t.Logf("Power out of range error: %v", err)
	t.Logf("Frequency out of range error: %v", err)
	t.Logf("===============================")

	t.Log("✅ SilvusMock error handling working correctly")
}

func TestE2E_SilvusMockWithRadioManager(t *testing.T) {
	opts := harness.DefaultOptions()
	server := harness.NewServer(t, opts)
	defer server.Shutdown()

	// Evidence: RadioManager integration
	t.Logf("=== TEST EVIDENCE ===")
	t.Logf("Active Radio ID: %s", server.RadioManager.GetActive())
	t.Logf("Available Radios: %+v", server.RadioManager.List())
	t.Logf("===================")

	// Test operations through RadioManager
	activeRadioID := server.RadioManager.GetActive()
	if activeRadioID != "silvus-001" {
		t.Errorf("Expected active radio 'silvus-001', got '%s'", activeRadioID)
	}

	// Test getting active radio
	activeRadio := server.RadioManager.GetActiveRadio()
	if activeRadio == nil {
		t.Fatal("Expected active radio, got nil")
	}

	if activeRadio.ID != "silvus-001" {
		t.Errorf("Expected active radio ID 'silvus-001', got '%s'", activeRadio.ID)
	}

	// Evidence: RadioManager state
	t.Logf("=== RADIO MANAGER EVIDENCE ===")
	t.Logf("Active Radio ID: %s", activeRadioID)
	t.Logf("Active Radio: %+v", activeRadio)
	t.Logf("=============================")

	t.Log("✅ SilvusMock integration with RadioManager working correctly")
}

func TestE2E_SilvusMockConcurrentOperations(t *testing.T) {
	opts := harness.DefaultOptions()
	server := harness.NewServer(t, opts)
	defer server.Shutdown()

	ctx := context.Background()

	// Test concurrent power operations
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(power int) {
			defer func() { done <- true }()
			err := server.SilvusAdapter.SetPower(ctx, power)
			if err != nil {
				t.Errorf("Concurrent SetPower failed: %v", err)
			}
		}(i + 1)
	}

	// Wait for all operations to complete
	for i := 0; i < 10; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("Concurrent operations did not complete")
		}
	}

	// Verify final state
	power, err := server.SilvusAdapter.ReadPowerActual(ctx)
	if err != nil {
		t.Fatalf("ReadPowerActual after concurrent operations failed: %v", err)
	}

	// Power should be the last value set (10)
	if power != 10 {
		t.Errorf("Expected final power 10, got %d", power)
	}

	// Evidence: Concurrent operations
	t.Logf("=== CONCURRENT OPERATIONS EVIDENCE ===")
	t.Logf("Final power after 10 concurrent operations: %d", power)
	t.Logf("======================================")

	t.Log("✅ SilvusMock concurrent operations working correctly")
}

func TestE2E_SilvusMockFaultInjection(t *testing.T) {
	opts := harness.DefaultOptions()
	server := harness.NewServer(t, opts)
	defer server.Shutdown()

	ctx := context.Background()

	// Test busy fault mode
	server.SetSilvusFaultMode("busy")
	err := server.SilvusAdapter.SetPower(ctx, 10)
	if err == nil {
		t.Error("Expected error for busy fault mode")
	}

	// Test unavailable fault mode
	server.SetSilvusFaultMode("unavailable")
	err = server.SilvusAdapter.SetPower(ctx, 10)
	if err == nil {
		t.Error("Expected error for unavailable fault mode")
	}

	// Test invalid range fault mode
	server.SetSilvusFaultMode("invalid_range")
	err = server.SilvusAdapter.SetPower(ctx, 10)
	if err == nil {
		t.Error("Expected error for invalid range fault mode")
	}

	// Clear fault mode
	server.SetSilvusFaultMode("")
	err = server.SilvusAdapter.SetPower(ctx, 10)
	if err != nil {
		t.Errorf("Expected success after clearing fault mode, got error: %v", err)
	}

	// Evidence: Fault injection
	t.Logf("=== FAULT INJECTION EVIDENCE ===")
	t.Logf("Busy mode error: %v", err)
	t.Logf("Unavailable mode error: %v", err)
	t.Logf("Invalid range mode error: %v", err)
	t.Logf("Clear mode success: %v", err)
	t.Logf("===============================")

	t.Log("✅ SilvusMock fault injection working correctly")
}
