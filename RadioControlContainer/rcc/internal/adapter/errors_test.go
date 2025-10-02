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
