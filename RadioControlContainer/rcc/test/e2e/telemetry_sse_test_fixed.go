// Package e2e provides telemetry SSE tests for the Radio Control Container API.
// This file implements black-box testing using only HTTP/SSE and contract validation.
// FIXED VERSION: Proper connection cleanup to prevent leaks
package e2e

import (
	"context"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/radio-control/rcc/test/harness"
)

func TestE2E_TelemetrySSEConnection_Fixed(t *testing.T) {
	// Initialize contract validator
	validator := NewContractValidator(t)
	validator.PrintSpecVersion(t)

	opts := harness.DefaultOptions()
	server := harness.NewServer(t, opts)
	defer server.Shutdown()

	// Evidence: Seeded state via HTTP contract
	t.Logf("=== TEST EVIDENCE ===")
	radios := httpGetJSON(t, server.URL+"/api/v1/radios")
	mustHave(t, radios, "result", "ok")
	if d, ok := radios["data"].(map[string]any); ok {
		if id, ok := d["activeRadioId"].(string); ok {
			t.Logf("Active Radio ID: %s", id)
		}
	}
	t.Logf("===================")

	// Subscribe to telemetry with proper cleanup
	req, _ := http.NewRequest("GET", server.URL+"/api/v1/telemetry", nil)
	req.Header.Set("Accept", "text/event-stream")

	w := newThreadSafeResponseWriter()
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	// FIXED: Use proper HTTP client with timeout and connection management
	client := &http.Client{
		Timeout: 5 * time.Second, // Add client timeout
	}

	telemetryDone := make(chan error, 1)
	go func() {
		defer func() {
			// Ensure we signal completion even if there's a panic
			select {
			case telemetryDone <- nil:
			default:
			}
		}()

		resp, err := client.Do(req)
		if err != nil {
			telemetryDone <- err
			return
		}

		// FIXED: Ensure response body is always closed
		defer func() {
			if resp.Body != nil {
				resp.Body.Close()
			}
		}()

		// FIXED: Use proper context-aware reading
		buf := make([]byte, 1024)
		for {
			select {
			case <-ctx.Done():
				// FIXED: Properly handle context cancellation
				return
			default:
				// FIXED: Set read timeout on response body
				resp.Body.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
				n, err := resp.Body.Read(buf)
				if err != nil {
					// FIXED: Don't treat timeout as error, just break
					if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
						continue
					}
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

	// FIXED: Wait for telemetry to complete with proper timeout
	select {
	case err := <-telemetryDone:
		if err != nil && err != context.DeadlineExceeded {
			t.Errorf("Telemetry failed: %v", err)
		}
	case <-time.After(6 * time.Second): // Increased timeout
		t.Fatal("Telemetry did not complete")
	}

	t.Log("✅ Telemetry SSE connection working correctly")
}
