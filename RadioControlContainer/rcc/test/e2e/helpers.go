// Package e2e provides shared helper functions for end-to-end tests.
package e2e

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/radio-control/rcc/internal/adapter"
	"github.com/radio-control/rcc/internal/adapter/silvusmock"
	"github.com/radio-control/rcc/internal/api"
	"github.com/radio-control/rcc/internal/command"
	"github.com/radio-control/rcc/internal/config"
	"github.com/radio-control/rcc/internal/radio"
	"github.com/radio-control/rcc/internal/telemetry"
)

// newServerForE2E creates a test server with SilvusMock adapter
func newServerForE2E(t *testing.T) *httptest.Server {
	t.Helper()
	cfg := config.LoadCBTimingBaseline()

	hub := telemetry.NewHub(cfg)
	rm := radio.NewManager()

	// Register SilvusMock with a band plan
	silvus := silvusmock.NewSilvusMock("silvus-001", []adapter.Channel{
		{Index: 1, FrequencyMhz: 2412.0},
		{Index: 6, FrequencyMhz: 2437.0},
		{Index: 11, FrequencyMhz: 2462.0},
	})

	err := rm.LoadCapabilities("silvus-001", silvus, 5*time.Second)
	if err != nil {
		t.Fatalf("Failed to load capabilities: %v", err)
	}

	err = rm.SetActive("silvus-001")
	if err != nil {
		t.Fatalf("Failed to set active radio: %v", err)
	}

	orch := command.NewOrchestratorWithRadioManager(hub, cfg, rm)
	s := api.NewServer(hub, orch, rm, 30*time.Second, 30*time.Second, 60*time.Second)
	ts := httptest.NewServer(s.Handler())
	t.Cleanup(ts.Close)
	return ts
}

// HTTP helper functions
func httpGetJSON(t *testing.T, url string) map[string]interface{} {
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("GET %s failed: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET %s returned status %d", url, resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to decode JSON response: %v", err)
	}

	return result
}

func httpPostJSON200(t *testing.T, url string, payload map[string]any) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal JSON: %v", err)
	}

	resp, err := http.Post(url, "application/json", strings.NewReader(string(jsonData)))
	if err != nil {
		t.Fatalf("POST %s failed: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST %s returned status %d", url, resp.StatusCode)
	}
}

func httpGetWithStatus(t *testing.T, url string) *http.Response {
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("GET %s failed: %v", url, err)
	}
	return resp
}

func httpPostWithStatus(t *testing.T, url, payload string) *http.Response {
	resp, err := http.Post(url, "application/json", strings.NewReader(payload))
	if err != nil {
		t.Fatalf("POST %s failed: %v", url, err)
	}
	return resp
}

// JSON assertion helpers
func mustHave(t *testing.T, data map[string]interface{}, path string, expected interface{}) {
	actual := getJSONPath(data, path)
	if actual != expected {
		t.Errorf("Expected %s to be %v, got %v", path, expected, actual)
	}
}

func mustHaveNumber(t *testing.T, data map[string]interface{}, path string, expected float64) {
	actual := getJSONPath(data, path)
	if num, ok := actual.(float64); !ok || num != expected {
		t.Errorf("Expected %s to be %v, got %v", path, expected, actual)
	}
}

func getJSONPath(data map[string]interface{}, path string) interface{} {
	parts := strings.Split(path, ".")
	current := data

	for i, part := range parts {
		if i == len(parts)-1 {
			return current[part]
		}

		if next, ok := current[part].(map[string]interface{}); ok {
			current = next
		} else {
			return nil
		}
	}

	return nil
}

// Thread-safe response writer for SSE testing
type threadSafeResponseWriter struct {
	events     chan string
	headers    http.Header
	statusCode int
}

func newThreadSafeResponseWriter() *threadSafeResponseWriter {
	return &threadSafeResponseWriter{
		events:     make(chan string, 100),
		headers:    make(http.Header),
		statusCode: 200,
	}
}

func (w *threadSafeResponseWriter) Header() http.Header {
	return w.headers
}

func (w *threadSafeResponseWriter) Write(data []byte) (int, error) {
	select {
	case w.events <- string(data):
		return len(data), nil
	default:
		return len(data), nil
	}
}

func (w *threadSafeResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

func (w *threadSafeResponseWriter) collectEvents(timeout time.Duration) []string {
	var events []string
	timeoutChan := time.After(timeout)

	for {
		select {
		case event := <-w.events:
			events = append(events, event)
		case <-timeoutChan:
			return events
		}
	}
}
