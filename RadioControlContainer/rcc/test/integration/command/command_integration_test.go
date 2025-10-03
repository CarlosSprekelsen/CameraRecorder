//go:build integration

package command

import (
	"context"
	"strings"
	"testing"

	"github.com/radio-control/rcc/internal/adapter"
	"github.com/radio-control/rcc/internal/command"
	"github.com/radio-control/rcc/test/integration/fakes"
	"github.com/radio-control/rcc/test/integration/harness"
)

// TestCommand_SetPower_PublishesAuditAndCallsAdapter tests the complete command flow through interfaces.
func TestCommand_SetPower_PublishesAuditAndCallsAdapter(t *testing.T) {
	// Arrange: Setup test stack with mocks
	orch, _, _, mockAudit, fakeAdapter := harness.BuildTestStack(t)
	
	// Cast to fake adapter for verification
	fakeAdapterTyped := fakeAdapter.(*fakes.FakeAdapter)

	// Act: Execute SetPower command
	ctx := context.Background()
	err := orch.SetPower(ctx, "fake-001", 25.0)
	
	// Assert: Command execution
	if err != nil {
		t.Errorf("SetPower failed: %v", err)
	}
	
	// Assert: Audit logging via mock (no filesystem access)
	if len(mockAudit.GetLoggedActions()) == 0 {
		t.Error("Expected audit log entry, but none was recorded")
	} else {
		actions := mockAudit.GetLoggedActions()
		if len(actions) != 1 {
			t.Errorf("Expected 1 audit action, got %d", len(actions))
		} else {
			action := actions[0]
			if action.Action != "setPower" {
				t.Errorf("Expected action 'setPower', got '%s'", action.Action)
			}
			if action.RadioID != "fake-001" {
				t.Errorf("Expected radioID 'fake-001', got '%s'", action.RadioID)
			}
			if action.Result != "SUCCESS" {
				t.Errorf("Expected result 'SUCCESS', got '%s'", action.Result)
			}
		}
	}
	
	// Assert: Adapter was called
	if fakeAdapterTyped.GetCallCount("SetPower") != 1 {
		t.Errorf("Expected SetPower to be called once, got %d calls", fakeAdapterTyped.GetCallCount("SetPower"))
	}
	if fakeAdapterTyped.GetLastSetPowerCall() != 25.0 {
		t.Errorf("Expected SetPower(25.0), got SetPower(%f)", fakeAdapterTyped.GetLastSetPowerCall())
	}
	
	t.Logf("✅ SetPower integration flow: Command → Audit → Adapter")
}

// TestCommand_SetChannelByIndex_ResolvesIndexToFrequency tests channel index resolution.
func TestCommand_SetChannelByIndex_ResolvesIndexToFrequency(t *testing.T) {
	// Arrange: Setup test stack with mocks
	orch, rm, _, mockAudit, fakeAdapter := harness.BuildTestStack(t)
	
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
	if fakeAdapterTyped.GetLastSetFrequencyCall() != 2437.0 {
		t.Errorf("Expected SetFrequency(2437.0), got SetFrequency(%f)", fakeAdapterTyped.GetLastSetFrequencyCall())
	}
	
	// Assert: Audit logging via mock
	if len(mockAudit.GetLoggedActions()) == 0 {
		t.Error("Expected audit log entry, but none was recorded")
	} else {
		actions := mockAudit.GetLoggedActions()
		if len(actions) != 1 {
			t.Errorf("Expected 1 audit action, got %d", len(actions))
		} else {
			action := actions[0]
			if action.Action != "setChannel" {
				t.Errorf("Expected action 'setChannel', got '%s'", action.Action)
			}
			if action.RadioID != "fake-001" {
				t.Errorf("Expected radioID 'fake-001', got '%s'", action.RadioID)
			}
			if action.Result != "SUCCESS" {
				t.Errorf("Expected result 'SUCCESS', got '%s'", action.Result)
			}
		}
	}
	
	t.Logf("✅ SetChannelByIndex integration flow: Index 6 → Frequency 2437.0 → Adapter")
}

// TestCommand_ErrorNormalization_Table tests error handling through interfaces.
func TestCommand_ErrorNormalization_Table(t *testing.T) {
	testCases := []struct {
		name        string
		mode        string
		operation   string
		expectedErr string
	}{
		{
			name:        "Happy mode - SetPower",
			mode:        "happy",
			operation:   "SetPower",
			expectedErr: "",
		},
		{
			name:        "Busy mode - SetPower",
			mode:        "busy",
			operation:   "SetPower",
			expectedErr: "BUSY",
		},
		{
			name:        "Unavailable mode - SetPower",
			mode:        "unavailable",
			operation:   "SetPower",
			expectedErr: "UNAVAILABLE",
		},
		{
			name:        "Invalid range mode - SetPower",
			mode:        "invalid-range",
			operation:   "SetPower",
			expectedErr: "INVALID_RANGE",
		},
		{
			name:        "Internal mode - SetPower",
			mode:        "internal",
			operation:   "SetPower",
			expectedErr: "INTERNAL",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange: Setup test stack with mocks
			orch, _, _, mockAudit, fakeAdapter := harness.BuildTestStack(t)
			
			// Cast to fake adapter and set mode
			fakeAdapterTyped := fakeAdapter.(*fakes.FakeAdapter)
			fakeAdapterTyped.WithMode(tc.mode)
			
			// Act: Execute operation
			ctx := context.Background()
			var operationErr error
			
			switch tc.operation {
			case "SetPower":
				operationErr = orch.SetPower(ctx, "fake-001", 25.0)
			default:
				t.Fatalf("Unknown operation: %s", tc.operation)
			}
			
			// Assert: Error normalization
			if tc.expectedErr == "" {
				if operationErr != nil {
					t.Errorf("Expected success, but got error: %v", operationErr)
				}
				// Success should emit audit
				if len(mockAudit.GetLoggedActions()) == 0 {
					t.Error("Expected audit log entry for successful operation")
				}
			} else {
				if operationErr == nil {
					t.Errorf("Expected %s error, but got none", tc.expectedErr)
				} else {
					if !strings.Contains(operationErr.Error(), tc.expectedErr) {
						t.Errorf("Expected error to contain '%s', got: %v", tc.expectedErr, operationErr)
					}
				}
			}
			
			t.Logf("✅ %s: %s", tc.operation, tc.name)
		})
	}
}