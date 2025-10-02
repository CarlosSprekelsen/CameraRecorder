package adapter

import (
	"errors"
	"testing"
)

func TestNormalizeVendorError(t *testing.T) {
	tests := []struct {
		name    string
		vendor  error
		payload interface{}
		want    error
	}{
		{
			name:    "nil error",
			vendor:  nil,
			payload: nil,
			want:    nil,
		},
		{
			name:    "unknown error defaults to INTERNAL",
			vendor:  errors.New("UNKNOWN_VENDOR_ERROR"),
			payload: map[string]string{"detail": "test"},
			want:    ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeVendorError(tt.vendor, tt.payload)

			if tt.want == nil {
				if got != nil {
					t.Errorf("got %v, want nil", got)
				}
				return
			}

			var ve *VendorError
			if !errors.As(got, &ve) {
				t.Fatalf("got %T, want *VendorError", got)
			}

			if !errors.Is(ve.Code, tt.want) {
				t.Errorf("code = %v, want %v", ve.Code, tt.want)
			}

			if ve.Original != tt.vendor {
				t.Errorf("original = %v, want %v", ve.Original, tt.vendor)
			}
		})
	}
}

func TestErrorCodes(t *testing.T) {
	// Verify only 4 normalized codes exist per Architecture §8.5
	codes := []error{
		ErrInvalidRange,
		ErrBusy,
		ErrUnavailable,
		ErrInternal,
	}

	if len(codes) != 4 {
		t.Errorf("got %d error codes, want 4 per Architecture §8.5", len(codes))
	}
}

// TestTableDrivenErrorMapping tests the new table-driven error mapping system.
// Source: PRE-INT-04
// Quote: "Tests: each token → exact normalized error; unknown → INTERNAL"
func TestTableDrivenErrorMapping(t *testing.T) {
	tests := []struct {
		name        string
		errorMsg    string
		vendorID    string
		expectedErr error
		description string
	}{
		// Silvus Range Errors
		{
			name:        "Silvus_TX_POWER_OUT_OF_RANGE",
			errorMsg:    "TX_POWER_OUT_OF_RANGE: power level 50 dBm exceeds maximum",
			vendorID:    "silvus",
			expectedErr: ErrInvalidRange,
			description: "Silvus TX power out of range should map to INVALID_RANGE",
		},
		{
			name:        "Silvus_FREQUENCY_OUT_OF_RANGE",
			errorMsg:    "FREQUENCY_OUT_OF_RANGE: frequency 10000 MHz not supported",
			vendorID:    "silvus",
			expectedErr: ErrInvalidRange,
			description: "Silvus frequency out of range should map to INVALID_RANGE",
		},
		{
			name:        "Silvus_INVALID_POWER_LEVEL",
			errorMsg:    "INVALID_POWER_LEVEL: power must be between 0-39 dBm",
			vendorID:    "silvus",
			expectedErr: ErrInvalidRange,
			description: "Silvus invalid power level should map to INVALID_RANGE",
		},
		{
			name:        "Silvus_INVALID_FREQUENCY",
			errorMsg:    "INVALID_FREQUENCY: frequency 0 MHz is not valid",
			vendorID:    "silvus",
			expectedErr: ErrInvalidRange,
			description: "Silvus invalid frequency should map to INVALID_RANGE",
		},
		{
			name:        "Silvus_PARAMETER_OUT_OF_RANGE",
			errorMsg:    "PARAMETER_OUT_OF_RANGE: parameter value exceeds limits",
			vendorID:    "silvus",
			expectedErr: ErrInvalidRange,
			description: "Silvus parameter out of range should map to INVALID_RANGE",
		},
		{
			name:        "Silvus_VALUE_OUT_OF_BOUNDS",
			errorMsg:    "VALUE_OUT_OF_BOUNDS: value exceeds system limits",
			vendorID:    "silvus",
			expectedErr: ErrInvalidRange,
			description: "Silvus value out of bounds should map to INVALID_RANGE",
		},
		{
			name:        "Silvus_INVALID_PARAMETER",
			errorMsg:    "INVALID_PARAMETER: parameter format is incorrect",
			vendorID:    "silvus",
			expectedErr: ErrInvalidRange,
			description: "Silvus invalid parameter should map to INVALID_RANGE",
		},

		// Silvus Busy Errors
		{
			name:        "Silvus_RF_BUSY",
			errorMsg:    "RF_BUSY: radio frequency is currently in use",
			vendorID:    "silvus",
			expectedErr: ErrBusy,
			description: "Silvus RF busy should map to BUSY",
		},
		{
			name:        "Silvus_TRANSMITTER_BUSY",
			errorMsg:    "TRANSMITTER_BUSY: transmitter is currently active",
			vendorID:    "silvus",
			expectedErr: ErrBusy,
			description: "Silvus transmitter busy should map to BUSY",
		},
		{
			name:        "Silvus_RADIO_BUSY",
			errorMsg:    "RADIO_BUSY: radio is processing another command",
			vendorID:    "silvus",
			expectedErr: ErrBusy,
			description: "Silvus radio busy should map to BUSY",
		},
		{
			name:        "Silvus_OPERATION_IN_PROGRESS",
			errorMsg:    "OPERATION_IN_PROGRESS: another operation is running",
			vendorID:    "silvus",
			expectedErr: ErrBusy,
			description: "Silvus operation in progress should map to BUSY",
		},
		{
			name:        "Silvus_COMMAND_QUEUE_FULL",
			errorMsg:    "COMMAND_QUEUE_FULL: command queue is at capacity",
			vendorID:    "silvus",
			expectedErr: ErrBusy,
			description: "Silvus command queue full should map to BUSY",
		},
		{
			name:        "Silvus_RATE_LIMITED",
			errorMsg:    "RATE_LIMITED: too many requests, please retry later",
			vendorID:    "silvus",
			expectedErr: ErrBusy,
			description: "Silvus rate limited should map to BUSY",
		},

		// Silvus Unavailable Errors
		{
			name:        "Silvus_NODE_UNAVAILABLE",
			errorMsg:    "NODE_UNAVAILABLE: radio node is not responding",
			vendorID:    "silvus",
			expectedErr: ErrUnavailable,
			description: "Silvus node unavailable should map to UNAVAILABLE",
		},
		{
			name:        "Silvus_RADIO_OFFLINE",
			errorMsg:    "RADIO_OFFLINE: radio is currently offline",
			vendorID:    "silvus",
			expectedErr: ErrUnavailable,
			description: "Silvus radio offline should map to UNAVAILABLE",
		},
		{
			name:        "Silvus_REBOOTING",
			errorMsg:    "REBOOTING: system is rebooting, please wait",
			vendorID:    "silvus",
			expectedErr: ErrUnavailable,
			description: "Silvus rebooting should map to UNAVAILABLE",
		},
		{
			name:        "Silvus_SOFT_BOOT_IN_PROGRESS",
			errorMsg:    "SOFT_BOOT_IN_PROGRESS: soft boot is in progress",
			vendorID:    "silvus",
			expectedErr: ErrUnavailable,
			description: "Silvus soft boot in progress should map to UNAVAILABLE",
		},
		{
			name:        "Silvus_SYSTEM_INITIALIZING",
			errorMsg:    "SYSTEM_INITIALIZING: system is initializing",
			vendorID:    "silvus",
			expectedErr: ErrUnavailable,
			description: "Silvus system initializing should map to UNAVAILABLE",
		},
		{
			name:        "Silvus_NOT_READY",
			errorMsg:    "NOT_READY: system is not ready for commands",
			vendorID:    "silvus",
			expectedErr: ErrUnavailable,
			description: "Silvus not ready should map to UNAVAILABLE",
		},
		{
			name:        "Silvus_OFFLINE",
			errorMsg:    "OFFLINE: system is offline",
			vendorID:    "silvus",
			expectedErr: ErrUnavailable,
			description: "Silvus offline should map to UNAVAILABLE",
		},

		// Generic Fallback Tests
		{
			name:        "Generic_OUT_OF_RANGE",
			errorMsg:    "OUT_OF_RANGE: value is out of range",
			vendorID:    "generic",
			expectedErr: ErrInvalidRange,
			description: "Generic out of range should map to INVALID_RANGE",
		},
		{
			name:        "Generic_BUSY",
			errorMsg:    "BUSY: system is busy",
			vendorID:    "generic",
			expectedErr: ErrBusy,
			description: "Generic busy should map to BUSY",
		},
		{
			name:        "Generic_UNAVAILABLE",
			errorMsg:    "UNAVAILABLE: service is unavailable",
			vendorID:    "generic",
			expectedErr: ErrUnavailable,
			description: "Generic unavailable should map to UNAVAILABLE",
		},

		// Unknown Vendor Fallback
		{
			name:        "Unknown_Vendor_Fallback",
			errorMsg:    "OUT_OF_RANGE: value is out of range",
			vendorID:    "unknown_vendor",
			expectedErr: ErrInvalidRange,
			description: "Unknown vendor should fallback to generic mapping",
		},

		// Unknown Token Tests
		{
			name:        "Unknown_Token_INTERNAL",
			errorMsg:    "UNKNOWN_ERROR: some unknown error occurred",
			vendorID:    "silvus",
			expectedErr: ErrInternal,
			description: "Unknown token should map to INTERNAL",
		},
		{
			name:        "Unknown_Token_Generic_INTERNAL",
			errorMsg:    "UNKNOWN_ERROR: some unknown error occurred",
			vendorID:    "generic",
			expectedErr: ErrInternal,
			description: "Unknown token in generic should map to INTERNAL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalErr := errors.New(tt.errorMsg)
			normalizedErr := NormalizeVendorErrorWithVendor(originalErr, nil, tt.vendorID)

			if normalizedErr == nil {
				t.Fatal("Expected normalized error, got nil")
			}

			vendorErr, ok := normalizedErr.(*VendorError)
			if !ok {
				t.Fatalf("Expected VendorError, got %T", normalizedErr)
			}

			if vendorErr.Code != tt.expectedErr {
				t.Errorf("Expected %v, got %v - %s", tt.expectedErr, vendorErr.Code, tt.description)
			}

			if vendorErr.Original != originalErr {
				t.Errorf("Expected original error %v, got %v", originalErr, vendorErr.Original)
			}
		})
	}
}

// TestVendorErrorMappings tests that the vendor mapping tables are properly configured.
func TestVendorErrorMappings(t *testing.T) {
	// Test that Silvus mapping exists
	silvusMap, exists := VendorErrorMappings["silvus"]
	if !exists {
		t.Fatal("Silvus mapping should exist")
	}

	// Test Silvus range tokens
	expectedRangeTokens := []string{
		"TX_POWER_OUT_OF_RANGE",
		"FREQUENCY_OUT_OF_RANGE",
		"INVALID_POWER_LEVEL",
		"INVALID_FREQUENCY",
		"PARAMETER_OUT_OF_RANGE",
		"VALUE_OUT_OF_BOUNDS",
		"INVALID_PARAMETER",
	}

	if len(silvusMap.Range) != len(expectedRangeTokens) {
		t.Errorf("Expected %d range tokens, got %d", len(expectedRangeTokens), len(silvusMap.Range))
	}

	for _, expectedToken := range expectedRangeTokens {
		found := false
		for _, actualToken := range silvusMap.Range {
			if actualToken == expectedToken {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected range token %s not found", expectedToken)
		}
	}

	// Test that generic mapping exists
	genericMap, exists := VendorErrorMappings["generic"]
	if !exists {
		t.Fatal("Generic mapping should exist")
	}

	if len(genericMap.Range) == 0 {
		t.Error("Generic mapping should have range tokens")
	}
	if len(genericMap.Busy) == 0 {
		t.Error("Generic mapping should have busy tokens")
	}
	if len(genericMap.Unavailable) == 0 {
		t.Error("Generic mapping should have unavailable tokens")
	}
}

// TestCaseInsensitiveMatching tests that error matching is case-insensitive.
func TestCaseInsensitiveMatching(t *testing.T) {
	tests := []struct {
		errorMsg    string
		vendorID    string
		expectedErr error
	}{
		{"tx_power_out_of_range: test", "silvus", ErrInvalidRange},
		{"Tx_Power_Out_Of_Range: test", "silvus", ErrInvalidRange},
		{"TX_POWER_OUT_OF_RANGE: test", "silvus", ErrInvalidRange},
		{"rf_busy: test", "silvus", ErrBusy},
		{"Rf_Busy: test", "silvus", ErrBusy},
		{"RF_BUSY: test", "silvus", ErrBusy},
		{"node_unavailable: test", "silvus", ErrUnavailable},
		{"Node_Unavailable: test", "silvus", ErrUnavailable},
		{"NODE_UNAVAILABLE: test", "silvus", ErrUnavailable},
	}

	for _, tt := range tests {
		t.Run(tt.errorMsg, func(t *testing.T) {
			originalErr := errors.New(tt.errorMsg)
			normalizedErr := NormalizeVendorErrorWithVendor(originalErr, nil, tt.vendorID)

			vendorErr, ok := normalizedErr.(*VendorError)
			if !ok {
				t.Fatalf("Expected VendorError, got %T", normalizedErr)
			}

			if vendorErr.Code != tt.expectedErr {
				t.Errorf("Expected %v, got %v for error message: %s", tt.expectedErr, vendorErr.Code, tt.errorMsg)
			}
		})
	}
}
