package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Ground truth: docs/api/health-endpoints.md

func TestHealthHTTP_Liveness(t *testing.T) {
	t.Parallel()
	fixture := NewE2EFixture(t)

	// Health port comes from config; reuse setup config
	cfg := fixture.setup.GetConfigManager().GetConfig()
	port := cfg.HTTPHealth.Port
	path := cfg.HTTPHealth.LiveEndpoint
	if path == "" {
		path = "/health"
	}
	url := fmt.Sprintf("http://127.0.0.1:%d%s", port, path)

	resp, err := http.Get(url)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)

	var body map[string]interface{}
	_ = json.NewDecoder(resp.Body).Decode(&body)
	// Minimal: allow either status presence or empty body per docs tolerance
}

func TestHealthHTTP_Readiness(t *testing.T) {
	t.Parallel()
	fixture := NewE2EFixture(t)
	cfg := fixture.setup.GetConfigManager().GetConfig()
	port := cfg.HTTPHealth.Port
	path := cfg.HTTPHealth.ReadyEndpoint
	if path == "" {
		path = "/health/ready"
	}
	url := fmt.Sprintf("http://127.0.0.1:%d%s", port, path)

	resp, err := http.Get(url)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
}

func TestHealthHTTP_Detailed(t *testing.T) {
	t.Parallel()
	fixture := NewE2EFixture(t)
	cfg := fixture.setup.GetConfigManager().GetConfig()
	port := cfg.HTTPHealth.Port
	path := cfg.HTTPHealth.DetailedEndpoint
	if path == "" {
		path = "/health/detailed"
	}
	url := fmt.Sprintf("http://127.0.0.1:%d%s", port, path)

	resp, err := http.Get(url)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
}
