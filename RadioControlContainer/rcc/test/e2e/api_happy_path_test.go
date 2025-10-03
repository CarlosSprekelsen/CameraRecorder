// Package e2e provides end-to-end tests for the Radio Control Container API.
// This file implements black-box testing using only HTTP/SSE and contract validation.
package e2e

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/radio-control/rcc/test/harness"
)

func TestE2E_HappyPath(t *testing.T) {
	// Initialize contract validator
	validator := NewContractValidator(t)
	validator.PrintSpecVersion(t)

	// Create test harness with seeded state
	opts := harness.DefaultOptions()
	opts.ActiveRadioID = "silvus-001"
	opts.CorrelationID = "test-001"

	server := harness.NewServer(t, opts)
	defer server.Shutdown()

	// Evidence: Route table and seeded IDs
	t.Logf("=== TEST EVIDENCE ===")
	t.Logf("Server URL: %s", server.URL)
	t.Logf("===================")

	// 1) List radios - should return seeded radio
	resp := httpGetWithStatus(t, server.URL+"/api/v1/radios")
	validator.ValidateHTTPResponse(t, resp, 200)

	body := httpGetJSON(t, server.URL+"/api/v1/radios")
	mustHave(t, body, "result", "ok")

	// Normalize dynamic fields in assertions
	// Check if data is a slice or map
	data := body["data"]
	if data == nil {
		t.Fatal("Expected 'data' field in response")
	}

	// Handle different response structures
	switch v := data.(type) {
	case []interface{}:
		if len(v) == 0 {
			t.Fatal("Expected at least one radio")
		}
		radio := v[0].(map[string]interface{})
		mustHave(t, radio, "id", "silvus-001")
		mustHave(t, radio, "type", "Silvus")
	case map[string]interface{}:
		// Single radio response
		mustHave(t, v, "id", "silvus-001")
		mustHave(t, v, "type", "Silvus")
	default:
		t.Fatalf("Unexpected data type: %T", v)
	}

	// 2) Select radio (should already be active)
	httpPostJSON200(t, server.URL+"/api/v1/radios/select", map[string]any{"radioId": "silvus-001"})

	// 3) Set power
	httpPostJSON200(t, server.URL+"/api/v1/radios/silvus-001/power", map[string]any{"powerDbm": 10.0})

	// 4) Set channel (by index ⇒ frequency mapping)
	httpPostJSON200(t, server.URL+"/api/v1/radios/silvus-001/channel", map[string]any{"channelIndex": 6})

	// 5) Read-back checks
	gotP := httpGetJSON(t, server.URL+"/api/v1/radios/silvus-001/power")
	mustHaveNumber(t, gotP, "data.powerDbm", 10.0)

	gotC := httpGetJSON(t, server.URL+"/api/v1/radios/silvus-001/channel")
	mustHaveNumber(t, gotC, "data.frequencyMhz", 2437.0)

	// Audit logs are server-side only per Architecture §8.6; no E2E access

	t.Log("✅ Happy path working correctly")
}

func TestE2E_TelemetryIntegration(t *testing.T) {
	// Initialize contract validator
	validator := NewContractValidator(t)
	validator.PrintSpecVersion(t)

	// Create test harness
	opts := harness.DefaultOptions()
	server := harness.NewServer(t, opts)
	defer server.Shutdown()

	// Evidence: Seeded state
	t.Logf("=== TEST EVIDENCE ===")
	t.Logf("Server URL: %s", server.URL)
	t.Logf("===================")

	// Subscribe to telemetry using HTTP SSE endpoint
	req, _ := http.NewRequest("GET", server.URL+"/api/v1/telemetry", nil)
	req.Header.Set("Accept", "text/event-stream")

	// Create thread-safe response writer
	w := newThreadSafeResponseWriter()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Start telemetry subscription via HTTP
	telemetryDone := make(chan error, 1)
	go func() {
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			telemetryDone <- err
			return
		}
		defer resp.Body.Close()

		// Read SSE stream
		buf := make([]byte, 1024)
		for {
			select {
			case <-ctx.Done():
				telemetryDone <- ctx.Err()
				return
			default:
				n, err := resp.Body.Read(buf)
				if err != nil {
					telemetryDone <- err
					return
				}
				w.Write(buf[:n])
			}
		}
	}()

	// Wait for subscription to start
	time.Sleep(100 * time.Millisecond)

	// Trigger power change
	httpPostJSON200(t, server.URL+"/api/v1/radios/silvus-001/power", map[string]any{"powerDbm": 25.0})

	// Wait for events
	time.Sleep(200 * time.Millisecond)

	// Collect telemetry events
	events := w.collectEvents(500 * time.Millisecond)
	response := strings.Join(events, "")

	// Evidence: SSE events
	t.Logf("=== SSE EVIDENCE ===")
	t.Logf("Received %d events", len(events))
	for i, event := range events {
		t.Logf("Event %d: %s", i+1, strings.TrimSpace(event))
		// Validate each event against contract
		validator.ValidateSSEEvent(t, event)
	}
	t.Logf("===================")

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
