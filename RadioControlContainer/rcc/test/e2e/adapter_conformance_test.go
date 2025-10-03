// Package e2e provides adapter conformance tests for SilvusMock.
package e2e

import (
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

	// Test GetCapabilities
	caps, err := silvus.GetCapabilities()
	if err != nil {
		t.Fatalf("GetCapabilities failed: %v", err)
	}

	if caps.RadioID != "silvus-001" {
		t.Errorf("Expected radio ID 'silvus-001', got '%s'", caps.RadioID)
	}

	if len(caps.Channels) != 3 {
		t.Errorf("Expected 3 channels, got %d", len(caps.Channels))
	}

	// Test SetPower
	err = silvus.SetPower(10.0)
	if err != nil {
		t.Fatalf("SetPower failed: %v", err)
	}

	// Test GetPower
	power, err := silvus.GetPower()
	if err != nil {
		t.Fatalf("GetPower failed: %v", err)
	}

	if power != 10.0 {
		t.Errorf("Expected power 10.0, got %f", power)
	}

	// Test SetChannel
	err = silvus.SetChannel(6)
	if err != nil {
		t.Fatalf("SetChannel failed: %v", err)
	}

	// Test GetChannel
	channel, err := silvus.GetChannel()
	if err != nil {
		t.Fatalf("GetChannel failed: %v", err)
	}

	if channel != 6 {
		t.Errorf("Expected channel 6, got %d", channel)
	}

	// Test GetState
	state, err := silvus.GetState()
	if err != nil {
		t.Fatalf("GetState failed: %v", err)
	}

	if state.RadioID != "silvus-001" {
		t.Errorf("Expected radio ID 'silvus-001', got '%s'", state.RadioID)
	}

	if state.PowerDbm != 10.0 {
		t.Errorf("Expected power 10.0, got %f", state.PowerDbm)
	}

	if state.ChannelIndex != 6 {
		t.Errorf("Expected channel 6, got %d", state.ChannelIndex)
	}

	t.Log("✅ SilvusMock adapter conformance working correctly")
}

func TestE2E_SilvusMockErrorHandling(t *testing.T) {
	silvus := silvusmock.NewSilvusMock("silvus-001", []adapter.Channel{
		{Index: 1, FrequencyMhz: 2412.0},
		{Index: 6, FrequencyMhz: 2437.0},
		{Index: 11, FrequencyMhz: 2462.0},
	})

	// Test power out of range
	err := silvus.SetPower(100.0)
	if err == nil {
		t.Error("Expected error for power out of range")
	}

	// Test channel out of range
	err = silvus.SetChannel(99)
	if err == nil {
		t.Error("Expected error for channel out of range")
	}

	// Test invalid frequency
	err = silvus.SetChannelByFrequency(10000.0)
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
	activeRadio, err := rm.GetActive()
	if err != nil {
		t.Fatalf("GetActive failed: %v", err)
	}

	if activeRadio.ID != "silvus-001" {
		t.Errorf("Expected active radio 'silvus-001', got '%s'", activeRadio.ID)
	}

	// Test power operations
	err = rm.SetPower("silvus-001", 15.0)
	if err != nil {
		t.Fatalf("SetPower through RadioManager failed: %v", err)
	}

	power, err := rm.GetPower("silvus-001")
	if err != nil {
		t.Fatalf("GetPower through RadioManager failed: %v", err)
	}

	if power != 15.0 {
		t.Errorf("Expected power 15.0, got %f", power)
	}

	// Test channel operations
	err = rm.SetChannel("silvus-001", 11)
	if err != nil {
		t.Fatalf("SetChannel through RadioManager failed: %v", err)
	}

	channel, err := rm.GetChannel("silvus-001")
	if err != nil {
		t.Fatalf("GetChannel through RadioManager failed: %v", err)
	}

	if channel != 11 {
		t.Errorf("Expected channel 11, got %d", channel)
	}

	t.Log("✅ SilvusMock integration with RadioManager working correctly")
}

func TestE2E_SilvusMockConcurrentOperations(t *testing.T) {
	silvus := silvusmock.NewSilvusMock("silvus-001", []adapter.Channel{
		{Index: 1, FrequencyMhz: 2412.0},
		{Index: 6, FrequencyMhz: 2437.0},
		{Index: 11, FrequencyMhz: 2462.0},
	})

	// Test concurrent power operations
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(power float64) {
			defer func() { done <- true }()
			err := silvus.SetPower(power)
			if err != nil {
				t.Errorf("Concurrent SetPower failed: %v", err)
			}
		}(float64(i + 1))
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
	power, err := silvus.GetPower()
	if err != nil {
		t.Fatalf("GetPower after concurrent operations failed: %v", err)
	}

	// Power should be the last value set (10.0)
	if power != 10.0 {
		t.Errorf("Expected final power 10.0, got %f", power)
	}

	t.Log("✅ SilvusMock concurrent operations working correctly")
}
