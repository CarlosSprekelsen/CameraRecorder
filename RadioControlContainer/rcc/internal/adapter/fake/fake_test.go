// Package fake provides a fake radio adapter implementation for testing.
package fake

import (
	"context"
	"testing"

	"github.com/radio-control/rcc/internal/adapter"
	"github.com/radio-control/rcc/internal/adaptertest"
)

// TestFakeAdapterConformance runs the complete conformance test suite on the fake adapter.
// Source: RE-INT-03
// Quote: "A single call in internal/adapter/fake/fake_test.go invoking adaptertest.RunConformance"
func TestFakeAdapterConformance(t *testing.T) {
	// Define capabilities for the fake adapter
	capabilities := adaptertest.Capabilities{
		MinPowerDbm:      0,
		MaxPowerDbm:      39,
		ValidFrequencies: []float64{2412.0, 2417.0, 2422.0, 2427.0, 2432.0},
		Channels: []adapter.Channel{
			{Index: 1, FrequencyMhz: 2412.0},
			{Index: 2, FrequencyMhz: 2417.0},
			{Index: 3, FrequencyMhz: 2422.0},
			{Index: 4, FrequencyMhz: 2427.0},
			{Index: 5, FrequencyMhz: 2432.0},
		},
		ExpectedErrors: adaptertest.ErrorExpectations{
			InvalidRangeKeywords: []string{"INVALID_RANGE", "OUT_OF_RANGE", "INVALID_PARAMETER"},
			BusyKeywords:         []string{"BUSY", "RETRY", "RATE_LIMIT"},
			UnavailableKeywords:  []string{"UNAVAILABLE", "OFFLINE", "NOT_READY"},
			InternalKeywords:     []string{"INTERNAL", "UNKNOWN", "ERROR"},
		},
	}

	// Run conformance tests
	adaptertest.RunConformance(t, func() adapter.IRadioAdapter {
		return NewFakeAdapter("fake-radio-01")
	}, capabilities)
}

// TestFakeAdapterBasicFunctionality tests basic functionality of the fake adapter.
func TestFakeAdapterBasicFunctionality(t *testing.T) {
	adapter := NewFakeAdapter("test-radio")
	ctx := context.Background()

	// Test basic state
	state, err := adapter.GetState(ctx)
	if err != nil {
		t.Fatalf("GetState failed: %v", err)
	}

	if state.PowerDbm != 20.0 {
		t.Errorf("Expected power 20.0, got %f", state.PowerDbm)
	}

	if state.FrequencyMhz != 2412.0 {
		t.Errorf("Expected frequency 2412.0, got %f", state.FrequencyMhz)
	}

	// Test power setting
	err = adapter.SetPower(ctx, 30.0)
	if err != nil {
		t.Fatalf("SetPower failed: %v", err)
	}

	state, err = adapter.GetState(ctx)
	if err != nil {
		t.Fatalf("GetState after SetPower failed: %v", err)
	}

	if state.PowerDbm != 30.0 {
		t.Errorf("Expected power 30.0 after SetPower, got %f", state.PowerDbm)
	}

	// Test frequency setting
	err = adapter.SetFrequency(ctx, 2417.0)
	if err != nil {
		t.Fatalf("SetFrequency failed: %v", err)
	}

	state, err = adapter.GetState(ctx)
	if err != nil {
		t.Fatalf("GetState after SetFrequency failed: %v", err)
	}

	if state.FrequencyMhz != 2417.0 {
		t.Errorf("Expected frequency 2417.0 after SetFrequency, got %f", state.FrequencyMhz)
	}
}

// TestFakeAdapterErrorSimulation tests error simulation functionality.
func TestFakeAdapterErrorSimulation(t *testing.T) {
	adapter := NewFakeAdapter("test-radio")
	ctx := context.Background()

	// Test INVALID_RANGE error simulation
	adapter.SetErrorSimulation("INVALID_RANGE")

	_, err := adapter.GetState(ctx)
	if err == nil {
		t.Error("Expected error when error simulation is enabled")
	}

	if err.Error() != "INVALID_RANGE: simulated range error" {
		t.Errorf("Expected INVALID_RANGE error, got: %v", err)
	}

	// Test BUSY error simulation
	adapter.SetErrorSimulation("BUSY")

	err = adapter.SetPower(ctx, 20)
	if err == nil {
		t.Error("Expected error when error simulation is enabled")
	}

	if err.Error() != "BUSY: simulated busy error" {
		t.Errorf("Expected BUSY error, got: %v", err)
	}

	// Disable error simulation
	adapter.DisableErrorSimulation()

	_, err = adapter.GetState(ctx)
	if err != nil {
		t.Errorf("Expected no error when error simulation is disabled, got: %v", err)
	}
}

// TestFakeAdapterValidation tests input validation.
func TestFakeAdapterValidation(t *testing.T) {
	adapter := NewFakeAdapter("test-radio")
	ctx := context.Background()

	// Test invalid power range
	err := adapter.SetPower(ctx, -1)
	if err == nil {
		t.Error("Expected error for invalid power (-1)")
	}

	err = adapter.SetPower(ctx, 100)
	if err == nil {
		t.Error("Expected error for invalid power (100)")
	}

	// Test invalid frequency
	err = adapter.SetFrequency(ctx, 0)
	if err == nil {
		t.Error("Expected error for invalid frequency (0)")
	}

	err = adapter.SetFrequency(ctx, -100)
	if err == nil {
		t.Error("Expected error for invalid frequency (-100)")
	}

	err = adapter.SetFrequency(ctx, 10000)
	if err == nil {
		t.Error("Expected error for invalid frequency (10000)")
	}
}
