// Package adapter defines IRadioAdapter interface from Architecture §5.
//
// Requirements:
//   - Architecture §8.5: "Normalized error codes: INVALID_RANGE, BUSY, UNAVAILABLE, INTERNAL"
//   - Architecture §8.5.1: "Deterministic mapping with diagnostic preservation"
package adapter

import (
	"errors"
	"fmt"
)

// Normalized container errors per Architecture §8.5
var (
	ErrInvalidRange = errors.New("INVALID_RANGE")
	ErrBusy         = errors.New("BUSY")
	ErrUnavailable  = errors.New("UNAVAILABLE")
	ErrInternal     = errors.New("INTERNAL")
)

// VendorError wraps vendor error with diagnostic details per Architecture §8.5.1
type VendorError struct {
	Code     error       // Normalized container code
	Original error       // Vendor error
	Details  interface{} // Vendor payload (opaque)
}

func (e *VendorError) Error() string {
	return fmt.Sprintf("%v (vendor: %v)", e.Code, e.Original)
}

func (e *VendorError) Unwrap() error {
	return e.Code
}

// NormalizeVendorError maps vendor errors to Architecture §8.5 codes.
func NormalizeVendorError(vendorErr error, vendorPayload interface{}) error {
	if vendorErr == nil {
		return nil
	}

	msg := vendorErr.Error()
	var code error

	// Architecture §8.5 normalization table
	switch {
	case isRangeError(msg):
		code = ErrInvalidRange
	case isBusyError(msg):
		code = ErrBusy
	case isUnavailableError(msg):
		code = ErrUnavailable
	default:
		code = ErrInternal
	}

	return &VendorError{
		Code:     code,
		Original: vendorErr,
		Details:  vendorPayload,
	}
}

func isRangeError(msg string) bool {
	// Vendor "OUT_OF_RANGE", "INVALID_PARAMETER", etc.
	// Add vendor-specific patterns as discovered
	return false // TODO: Implement per vendor ICD
}

func isBusyError(msg string) bool {
	// Vendor "BUSY", "RETRY", etc.
	return false // TODO: Implement per vendor ICD
}

func isUnavailableError(msg string) bool {
	// Vendor "UNAVAILABLE", "REBOOTING", "SOFT_BOOT", etc.
	return false // TODO: Implement per vendor ICD
}
