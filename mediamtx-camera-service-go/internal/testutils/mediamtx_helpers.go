package testutils

import (
	"context"
	"fmt"
	"os"
	"testing"
)

// MediaMTXClientInterface minimal interface for readiness checking
// Avoids import cycle with internal/mediamtx package
type MediaMTXClientInterface interface {
	Get(ctx context.Context, endpoint string) ([]byte, error)
}

// MediaMTXHelper provides MediaMTX integration utilities
type MediaMTXHelper struct {
	setup *UniversalTestSetup
}

// NewMediaMTXHelper creates helper
func NewMediaMTXHelper(setup *UniversalTestSetup) *MediaMTXHelper {
	return &MediaMTXHelper{setup: setup}
}

// GetMediaMTXBaseURL constructs URL from config fixture
func (m *MediaMTXHelper) GetMediaMTXBaseURL() string {
	config := m.setup.GetConfigManager().GetConfig()
	return fmt.Sprintf("http://%s:%d", config.MediaMTX.Host, config.MediaMTX.APIPort)
}

// WaitForMediaMTXReady waits for server using WaitForCondition
func (m *MediaMTXHelper) WaitForMediaMTXReady(
	ctx context.Context,
	client MediaMTXClientInterface,
	endpoint string,
) error {
	return WaitForCondition(ctx, func() bool {
		probeCtx, probeCancel := context.WithTimeout(context.Background(), UniversalTimeoutShort)
		defer probeCancel()
		
		_, err := client.Get(probeCtx, endpoint)
		return err == nil
	}, UniversalTimeoutLong, "MediaMTX server ready")
}

// SkipIfMediaMTXUnavailable handles skip/fail based on environment
func (m *MediaMTXHelper) SkipIfMediaMTXUnavailable(t *testing.T, err error) {
	if err != nil {
		if os.Getenv("CI") == "true" {
			t.Fatalf("MediaMTX required in CI: %v", err)
		}
		t.Skipf("MediaMTX not available: %v", err)
	}
}
