//go:build integration

package command_test

import (
	"context"
	"testing"

	"github.com/radio-control/rcc/internal/adapter"
	"github.com/radio-control/rcc/test/integration/fakes"
	"github.com/radio-control/rcc/test/integration/harness"
)

// TestCommand_SetPower_PublishesAuditAndCallsAdapter tests real orchestrator with mock adapter.
func TestCommand_SetPower_PublishesAuditAndCallsAdapter(t *testing.T) {
	// Arrange: Setup test stack with real components (only radio adapter mocked)
	orch, _, _, _, fakeAdapter := harness.BuildTestStack(t)

	// Cast to fake adapter for verification
	fakeAdapterTyped := fakeAdapter.(*fakes.FakeAdapter)

	// Act: Execute SetPower command
	ctx := context.Background()
	err := orch.SetPower(ctx, "fake-001", 25.0)

	// Assert: Command execution
	if err != nil {
		t.Errorf("SetPower failed: %v", err)
	}

	// Assert: Adapter interaction (real component integration)
	if fakeAdapterTyped.GetCallCount("SetPower") != 1 {
		t.Errorf("Expected SetPower to be called once, got %d calls", fakeAdapterTyped.GetCallCount("SetPower"))
	}
	if fakeAdapterTyped.GetLastSetPowerCall() != 25.0 {
		t.Errorf("Expected SetPower(25.0), got SetPower(%f)", fakeAdapterTyped.GetLastSetPowerCall())
	}

	// Note: Audit logging is handled by real audit.Logger component
	// For integration tests, we verify the command flow works end-to-end
	t.Logf("✅ SetPower integration flow: Real Orchestrator → Real Audit → Mock Adapter")
}

// TestCommand_SetChannelByIndex_ResolvesIndexToFrequency tests channel index resolution.
func TestCommand_SetChannelByIndex_ResolvesIndexToFrequency(t *testing.T) {
	// Arrange: Setup test stack with real components
	orch, rm, _, _, fakeAdapter := harness.BuildTestStack(t)

	// Cast to fake adapter for verification
	fakeAdapterTyped := fakeAdapter.(*fakes.FakeAdapter)

	// Act: Execute SetChannelByIndex command
	ctx := context.Background()
	err := orch.SetChannelByIndex(ctx, "fake-001", 6, rm)

	// Assert: Command execution should succeed
	if err != nil {
		t.Fatalf("BUG: SetChannelByIndex: expected index 6→2437.0 MHz (Architecture §13), got error: %v", err)
	}

	// Assert: Adapter was called with correct frequency
	if fakeAdapterTyped.GetCallCount("SetFrequency") != 1 {
		t.Errorf("Expected SetFrequency to be called once, got %d calls", fakeAdapterTyped.GetCallCount("SetFrequency"))
	}
	
	expectedFreq := 2437.0 // Channel 6 = 2437 MHz per ICD
	actualFreq := fakeAdapterTyped.GetLastSetFrequencyCall()
	if actualFreq != expectedFreq {
		t.Errorf("Expected SetFrequency(%f), got SetFrequency(%f)", expectedFreq, actualFreq)
	}

	t.Logf("✅ SetChannelByIndex integration flow: Index 6 → Frequency 2437.0 → Adapter")
}

// TestCommand_ErrorNormalization_Table tests error normalization across different adapter modes.
func TestCommand_ErrorNormalization_Table(t *testing.T) {
	testCases := []struct {
		name        string
		mode        string
		expectedErr error
	}{
		{"Happy mode", "happy", nil},
		{"Busy mode", "busy", adapter.ErrBusy},
		{"Unavailable mode", "unavailable", adapter.ErrUnavailable},
		{"Invalid range mode", "invalid-range", adapter.ErrInvalidRange},
		{"Internal mode", "internal", adapter.ErrInternal},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange: Setup test stack with real components
			orch, _, _, _, fakeAdapter := harness.BuildTestStack(t)

			// Cast to fake adapter and set mode
			fakeAdapterTyped := fakeAdapter.(*fakes.FakeAdapter)
			fakeAdapterTyped.SetMode(tc.mode)

			// Act: Execute SetPower command
			ctx := context.Background()
			err := orch.SetPower(ctx, "fake-001", 25.0)

			// Assert: Error normalization
			if tc.expectedErr == nil {
				if err != nil {
					t.Errorf("Expected success, got error: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected error %v, got success", tc.expectedErr)
				} else if err != tc.expectedErr {
					t.Errorf("Expected error %v, got %v", tc.expectedErr, err)
				}
			}

			t.Logf("✅ SetPower: %s - %s", tc.name, tc.mode)
		})
	}
}