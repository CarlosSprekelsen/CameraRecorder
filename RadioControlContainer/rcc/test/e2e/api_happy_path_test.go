// Package e2e provides end-to-end tests for the Radio Control Container API.
package e2e

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/radio-control/rcc/internal/adapter"
	"github.com/radio-control/rcc/internal/adapter/silvusmock"
	"github.com/radio-control/rcc/internal/api"
	"github.com/radio-control/rcc/internal/audit"
	"github.com/radio-control/rcc/internal/command"
	"github.com/radio-control/rcc/internal/config"
	"github.com/radio-control/rcc/internal/radio"
	"github.com/radio-control/rcc/internal/telemetry"
)

func TestE2E_HappyPath(t *testing.T) {
	ts := newServerForE2E(t)

	// 1) List radios
	body := httpGetJSON(t, ts.URL+"/api/v1/radios")
	mustHave(t, body, "result", "ok")

	// 2) Select radio
	httpPostJSON200(t, ts.URL+"/api/v1/radios/select", map[string]any{"radioId": "silvus-001"})

	// 3) Set power
	httpPostJSON200(t, ts.URL+"/api/v1/radios/silvus-001/power", map[string]any{"powerDbm": 10.0})

	// 4) Set channel (by index ⇒ frequency mapping)
	httpPostJSON200(t, ts.URL+"/api/v1/radios/silvus-001/channel", map[string]any{"channelIndex": 6})

	// 5) Read-back checks
	gotP := httpGetJSON(t, ts.URL+"/api/v1/radios/silvus-001/power")
	mustHaveNumber(t, gotP, "data.powerDbm", 10.0)

	gotC := httpGetJSON(t, ts.URL+"/api/v1/radios/silvus-001/channel")
	mustHaveNumber(t, gotC, "data.frequencyMhz", 2437.0)
}

func TestE2E_TelemetryIntegration(t *testing.T) {
	// Setup test environment
	cfg := config.LoadCBTimingBaseline()
	hub := telemetry.NewHub(cfg)
	defer hub.Stop()

	rm := radio.NewManager()
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

	aud := audit.NewInMemory()
	orch := command.NewOrchestratorWithRadioManager(hub, cfg, rm)
	s := api.NewServer(hub, orch, rm, 30*time.Second, 30*time.Second, 60*time.Second)
	ts := httptest.NewServer(s)
	defer ts.Close()

	// Subscribe to telemetry
	req, _ := http.NewRequest("GET", ts.URL+"/api/v1/telemetry", nil)
	req.Header.Set("Accept", "text/event-stream")

	// Create thread-safe response writer
	w := newThreadSafeResponseWriter()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Start telemetry subscription
	telemetryDone := make(chan error, 1)
	go func() {
		telemetryDone <- hub.Subscribe(ctx, w, req)
	}()

	// Wait for subscription to start
	time.Sleep(100 * time.Millisecond)

	// Trigger power change
	httpPostJSON200(t, ts.URL+"/api/v1/radios/silvus-001/power", map[string]any{"powerDbm": 25.0})

	// Wait for events
	time.Sleep(200 * time.Millisecond)

	// Collect telemetry events
	events := w.collectEvents(500 * time.Millisecond)
	response := strings.Join(events, "")

	// Verify telemetry events
	if !strings.Contains(response, "event: ready") {
		t.Error("Expected ready event in telemetry")
	}

	if !strings.Contains(response, "powerChanged") {
		t.Error("Expected powerChanged event in telemetry")
	}

	// Wait for telemetry to complete
	select {
	case err := <-telemetryDone:
		if err != nil && err != context.DeadlineExceeded {
			t.Errorf("Telemetry failed: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Telemetry did not complete")
	}

	t.Log("✅ Telemetry integration working correctly")
}
