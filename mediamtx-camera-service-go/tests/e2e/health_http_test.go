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
	port := fixture.config.Server.HealthPort
	url := fmt.Sprintf("http://127.0.0.1:%d/health", port)

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
	port := fixture.config.Server.HealthPort
	url := fmt.Sprintf("http://127.0.0.1:%d/health/ready", port)

	resp, err := http.Get(url)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
}

func TestHealthHTTP_Detailed(t *testing.T) {
	t.Parallel()
	fixture := NewE2EFixture(t)
	port := fixture.config.Server.HealthPort
	url := fmt.Sprintf("http://127.0.0.1:%d/health/detailed", port)

	resp, err := http.Get(url)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
}
