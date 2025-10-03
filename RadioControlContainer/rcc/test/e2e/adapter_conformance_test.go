// Package e2e provides adapter conformance tests for SilvusMock.
package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/radio-control/rcc/internal/adapter"
	"github.com/radio-control/rcc/internal/adapter/silvusmock"
	"github.com/radio-control/rcc/internal/radio"
)

func TestE2E_SilvusMockConformance(t *testing.T) {
	// Test SilvusMock adapter directly
	silvus := silvusmock.NewSilvusMock("silvus-001", []adapter.Channel{
		{Index: 1, FrequencyMhz: 2412.0},
		{Index: 6, FrequencyMhz: 2437.0},
		{Index: 11, FrequencyMhz: 2462.0},
	})

	ctx := context.Background()

	// Test GetState
	state, err := silvus.GetState(ctx)
	if err != nil {
		t.Fatalf("GetState failed: %v", err)
	}

	// RadioState doesn't have RadioID field, check power instead
	if state.PowerDbm < 0 || state.PowerDbm > 100 {
		t.Errorf("Expected reasonable power value, got %d", state.PowerDbm)
	}

	// Test SetPower
	err = silvus.SetPower(ctx, 10)
	if err != nil {
		t.Fatalf("SetPower failed: %v", err)
	}

	// Test ReadPowerActual
	power, err := silvus.ReadPowerActual(ctx)
	if err != nil {
		t.Fatalf("ReadPowerActual failed: %v", err)
	}

	if power != 10 {
		t.Errorf("Expected power 10, got %d", power)
	}

	// Test SetFrequency
	err = silvus.SetFrequency(ctx, 2437.0)
	if err != nil {
		t.Fatalf("SetFrequency failed: %v", err)
	}

	// Test GetState after changes
	state, err = silvus.GetState(ctx)
	if err != nil {
		t.Fatalf("GetState after changes failed: %v", err)
	}

	if state.PowerDbm != 10 {
		t.Errorf("Expected power 10, got %d", state.PowerDbm)
	}

	if state.FrequencyMhz != 2437.0 {
		t.Errorf("Expected frequency 2437.0, got %f", state.FrequencyMhz)
	}

	t.Log("✅ SilvusMock adapter conformance working correctly")
}

func TestE2E_SilvusMockErrorHandling(t *testing.T) {
	silvus := silvusmock.NewSilvusMock("silvus-001", []adapter.Channel{
		{Index: 1, FrequencyMhz: 2412.0},
		{Index: 6, FrequencyMhz: 2437.0},
		{Index: 11, FrequencyMhz: 2462.0},
	})

	ctx := context.Background()

	// Test power out of range
	err := silvus.SetPower(ctx, 100)
	if err == nil {
		t.Error("Expected error for power out of range")
	}

	// Test frequency out of range
	err = silvus.SetFrequency(ctx, 10000.0)
	if err == nil {
		t.Error("Expected error for frequency out of range")
	}

	t.Log("✅ SilvusMock error handling working correctly")
}

func TestE2E_SilvusMockWithRadioManager(t *testing.T) {
	// Test SilvusMock integration with RadioManager
	rm := radio.NewManager()
	silvus := silvusmock.NewSilvusMock("silvus-001", []adapter.Channel{
		{Index: 1, FrequencyMhz: 2412.0},
		{Index: 6, FrequencyMhz: 2437.0},
		{Index: 11, FrequencyMhz: 2462.0},
	})

	// Load capabilities
	err := rm.LoadCapabilities("silvus-001", silvus, 5*time.Second)
	if err != nil {
		t.Fatalf("LoadCapabilities failed: %v", err)
	}

	// Set as active
	err = rm.SetActive("silvus-001")
	if err != nil {
		t.Fatalf("SetActive failed: %v", err)
	}

	// Test operations through RadioManager
	activeRadioID := rm.GetActive()
	if activeRadioID != "silvus-001" {
		t.Errorf("Expected active radio 'silvus-001', got '%s'", activeRadioID)
	}

	// Test getting active radio
	activeRadio := rm.GetActiveRadio()
	if activeRadio == nil {
		t.Fatal("Expected active radio, got nil")
	}

	if activeRadio.ID != "silvus-001" {
		t.Errorf("Expected active radio ID 'silvus-001', got '%s'", activeRadio.ID)
	}

	t.Log("✅ SilvusMock integration with RadioManager working correctly")
}

func TestE2E_SilvusMockConcurrentOperations(t *testing.T) {
	silvus := silvusmock.NewSilvusMock("silvus-001", []adapter.Channel{
		{Index: 1, FrequencyMhz: 2412.0},
		{Index: 6, FrequencyMhz: 2437.0},
		{Index: 11, FrequencyMhz: 2462.0},
	})

	ctx := context.Background()

	// Test concurrent power operations
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(power int) {
			defer func() { done <- true }()
			err := silvus.SetPower(ctx, power)
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
	power, err := silvus.ReadPowerActual(ctx)
	if err != nil {
		t.Fatalf("ReadPowerActual after concurrent operations failed: %v", err)
	}

	// Power should be the last value set (10)
	if power != 10 {
		t.Errorf("Expected final power 10, got %d", power)
	}

	t.Log("✅ SilvusMock concurrent operations working correctly")
}
