//go:build integration

package command

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/radio-control/rcc/internal/adapter"
	"github.com/radio-control/rcc/test/integration/fakes"
	"github.com/radio-control/rcc/test/integration/harness"
)

// TestErrorNormalization_Table tests error normalization at the command layer.
// Boundary: command; validates that adapter errors are properly normalized.
func TestErrorNormalization_Table(t *testing.T) {
	testCases := []struct {
		name        string
		mode        string
		operation   string
		expectedErr string
		description string
	}{
		{
			name:        "Happy mode - SetPower",
			mode:        "happy",
			operation:   "SetPower",
			expectedErr: "",
			description: "Happy mode should succeed",
		},
		{
			name:        "Busy mode - SetPower",
			mode:        "busy",
			operation:   "SetPower",
			expectedErr: "BUSY",
			description: "Busy mode should return BUSY error",
		},
		{
			name:        "Unavailable mode - SetPower",
			mode:        "unavailable",
			operation:   "SetPower",
			expectedErr: "UNAVAILABLE",
			description: "Unavailable mode should return UNAVAILABLE error",
		},
		{
			name:        "Invalid range mode - SetPower",
			mode:        "invalid-range",
			operation:   "SetPower",
			expectedErr: "INVALID_RANGE",
			description: "Invalid range mode should return INVALID_RANGE error",
		},
		{
			name:        "Internal mode - SetPower",
			mode:        "internal",
			operation:   "SetPower",
			expectedErr: "INTERNAL",
			description: "Internal mode should return INTERNAL error",
		},
		{
			name:        "Busy mode - SetFrequency",
			mode:        "busy",
			operation:   "SetFrequency",
			expectedErr: "BUSY",
			description: "Busy mode should return BUSY error for SetFrequency",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange: Build test stack with fake adapter in specific mode
			orch, rm, _, _, _, _, cleanup := harness.BuildTestStack()
			defer cleanup()

			// Get the fake adapter and set mode
			activeAdapter, _, err := rm.GetActiveAdapter()
			if err != nil {
				t.Fatalf("Failed to get active adapter: %v", err)
			}
			fakeAdapter := activeAdapter.(*fakes.FakeAdapter)
			fakeAdapter.WithMode(tc.mode)

			// Act: Perform the operation
			ctx := context.Background()
			var operationErr error

			switch tc.operation {
			case "SetPower":
				operationErr = orch.SetPower(ctx, "fake-001", 25.0)
			case "SetFrequency":
				operationErr = orch.SetChannel(ctx, "fake-001", 2437.0)
			default:
				t.Fatalf("Unknown operation: %s", tc.operation)
			}

			// Assert: Error normalization
			if tc.expectedErr == "" {
				// Should succeed
				if operationErr != nil {
					t.Errorf("Expected success, but got error: %v", operationErr)
				}
			} else {
				// Should fail with expected error
				if operationErr == nil {
					t.Errorf("Expected %s error, but got none", tc.expectedErr)
				} else {
					if !strings.Contains(operationErr.Error(), tc.expectedErr) {
						t.Errorf("Expected error to contain '%s', but got: %v", tc.expectedErr, operationErr)
					}
				}
			}

			t.Logf("✅ %s: %s", tc.operation, tc.description)
		})
	}
}

// TestErrorNormalization_AdapterErrors tests specific adapter error types.
func TestErrorNormalization_AdapterErrors(t *testing.T) {
	// Test that adapter errors are properly normalized
	testCases := []struct {
		name           string
		power          float64
		expectedErr    error
		expectedString string
	}{
		{
			name:           "Valid power",
			power:          25.0,
			expectedErr:    nil,
			expectedString: "",
		},
		{
			name:           "Power too high",
			power:          50.0,
			expectedErr:    adapter.ErrInvalidRange,
			expectedString: "INVALID_RANGE",
		},
		{
			name:           "Power negative",
			power:          -5.0,
			expectedErr:    adapter.ErrInvalidRange,
			expectedString: "INVALID_RANGE",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange: Build test stack
			orch, _, _, _, _, _, cleanup := harness.BuildTestStack()
			defer cleanup()

			// Act: Call SetPower
			ctx := context.Background()
			err := orch.SetPower(ctx, "fake-001", tc.power)

			// Assert: Error type and content
			if tc.expectedErr == nil {
				if err != nil {
					t.Errorf("Expected no error, but got: %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("Expected error, but got none")
				} else {
					// Check if error is properly wrapped/normalized
					if !errors.Is(err, tc.expectedErr) && !strings.Contains(err.Error(), tc.expectedString) {
						t.Errorf("Expected error to contain '%s' or be of type %v, but got: %v", tc.expectedString, tc.expectedErr, err)
					}
				}
			}
		})
	}
}

// TestErrorNormalization_UnknownRadio tests error handling for unknown radios.
func TestErrorNormalization_UnknownRadio(t *testing.T) {
	// Arrange: Build test stack
	orch, _, _, _, _, _, cleanup := harness.BuildTestStack()
	defer cleanup()

	// Act: Try to operate on unknown radio
	ctx := context.Background()
	err := orch.SetPower(ctx, "unknown-radio", 25.0)

	// Assert: Should get NOT_FOUND error
	if err == nil {
		t.Error("Expected error for unknown radio, but got none")
	} else {
		if !strings.Contains(err.Error(), "NOT_FOUND") && !strings.Contains(err.Error(), "not found") {
			t.Errorf("Expected NOT_FOUND error for unknown radio, got: %v", err)
		}
	}

	t.Logf("✅ Unknown radio handling: Returns NOT_FOUND error")
}
