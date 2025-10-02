// Package adapter defines IRadioAdapter interface from Architecture §5.
//
// Requirements:
//   - Architecture §8.5: "Normalized error codes: INVALID_RANGE, BUSY, UNAVAILABLE, INTERNAL"
//   - Architecture §8.5.1: "Deterministic mapping with diagnostic preservation"
package adapter

import (
	"errors"
	"fmt"
	"strings"
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
	// Basic deterministic mapping by keyword; refined per vendor ICD later
	m := strings.ToUpper(msg)
	return strings.Contains(m, "OUT_OF_RANGE") ||
		strings.Contains(m, "INVALID_PARAMETER") ||
		strings.Contains(m, "INVALID_RANGE") ||
		strings.Contains(m, "BAD_VALUE") ||
		strings.Contains(m, "RANGE")
}

func isBusyError(msg string) bool {
	m := strings.ToUpper(msg)
	return strings.Contains(m, "BUSY") ||
		strings.Contains(m, "RETRY") ||
		strings.Contains(m, "RATE_LIMIT") ||
		strings.Contains(m, "TOO_MANY_REQUESTS") ||
		strings.Contains(m, "BACKOFF")
}

func isUnavailableError(msg string) bool {
	m := strings.ToUpper(msg)
	return strings.Contains(m, "UNAVAILABLE") ||
		strings.Contains(m, "REBOOT") ||
		strings.Contains(m, "SOFT_BOOT") ||
		strings.Contains(m, "OFFLINE") ||
		strings.Contains(m, "NOT_READY")
}
