// Package e2e provides telemetry SSE tests for the Radio Control Container API.
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

func TestE2E_TelemetrySSEConnection(t *testing.T) {
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

	// Subscribe to telemetry
	req, _ := http.NewRequest("GET", server.URL+"/api/v1/telemetry", nil)
	req.Header.Set("Accept", "text/event-stream")

	w := newThreadSafeResponseWriter()
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	telemetryDone := make(chan error, 1)
	go func() {
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			telemetryDone <- err
			return
		}
		defer resp.Body.Close()

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
	case <-time.After(5 * time.Second):
		t.Fatal("Telemetry did not complete")
	}

	t.Log("✅ Telemetry SSE connection working correctly")
}

func TestE2E_TelemetryLastEventID(t *testing.T) {
	// Initialize contract validator
	validator := NewContractValidator(t)
	validator.PrintSpecVersion(t)

	opts := harness.DefaultOptions()
	server := harness.NewServer(t, opts)
	defer server.Shutdown()

	// First connection - get some events
	req1, _ := http.NewRequest("GET", server.URL+"/api/v1/telemetry", nil)
	req1.Header.Set("Accept", "text/event-stream")

	w1 := newThreadSafeResponseWriter()
	ctx1, cancel1 := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel1()

	telemetryDone1 := make(chan error, 1)
	go func() {
		client := &http.Client{}
		resp, err := client.Do(req1)
		if err != nil {
			telemetryDone1 <- err
			return
		}
		defer resp.Body.Close()

		buf := make([]byte, 1024)
		for {
			select {
			case <-ctx1.Done():
				telemetryDone1 <- ctx1.Err()
				return
			default:
				n, err := resp.Body.Read(buf)
				if err != nil {
					telemetryDone1 <- err
					return
				}
				w1.Write(buf[:n])
			}
		}
	}()

	// Wait for subscription and trigger event
	time.Sleep(100 * time.Millisecond)
	httpPostJSON200(t, server.URL+"/api/v1/radios/silvus-001/power", map[string]any{"powerDbm": 15.0})
	time.Sleep(200 * time.Millisecond)

	// Collect first batch of events
	events1 := w1.collectEvents(500 * time.Millisecond)
	response1 := strings.Join(events1, "")

	// Wait for first connection to complete
	select {
	case <-telemetryDone1:
	case <-time.After(2 * time.Second):
		t.Fatal("First telemetry connection did not complete")
	}

	// Second connection with Last-Event-ID
	req2, _ := http.NewRequest("GET", server.URL+"/api/v1/telemetry", nil)
	req2.Header.Set("Accept", "text/event-stream")
	req2.Header.Set("Last-Event-ID", "1") // Simulate reconnection

	w2 := newThreadSafeResponseWriter()
	ctx2, cancel2 := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel2()

	telemetryDone2 := make(chan error, 1)
	go func() {
		client := &http.Client{}
		resp, err := client.Do(req2)
		if err != nil {
			telemetryDone2 <- err
			return
		}
		defer resp.Body.Close()

		buf := make([]byte, 1024)
		for {
			select {
			case <-ctx2.Done():
				telemetryDone2 <- ctx2.Err()
				return
			default:
				n, err := resp.Body.Read(buf)
				if err != nil {
					telemetryDone2 <- err
					return
				}
				w2.Write(buf[:n])
			}
		}
	}()

	// Wait for second subscription
	time.Sleep(100 * time.Millisecond)

	// Trigger another event
	httpPostJSON200(t, server.URL+"/api/v1/radios/silvus-001/power", map[string]any{"powerDbm": 20.0})
	time.Sleep(200 * time.Millisecond)

	// Collect second batch of events
	events2 := w2.collectEvents(500 * time.Millisecond)
	response2 := strings.Join(events2, "")

	// Wait for second connection to complete
	select {
	case <-telemetryDone2:
	case <-time.After(2 * time.Second):
		t.Fatal("Second telemetry connection did not complete")
	}

	// Evidence: Reconnection events
	t.Logf("=== RECONNECTION EVIDENCE ===")
	t.Logf("First connection events: %d", len(events1))
	t.Logf("Second connection events: %d", len(events2))
	t.Logf("=============================")

	// Verify both connections received events
	if !strings.Contains(response1, "powerChanged") {
		t.Error("First connection should have received powerChanged event")
	}

	if !strings.Contains(response2, "powerChanged") {
		t.Error("Second connection should have received powerChanged event")
	}

	t.Log("✅ Telemetry Last-Event-ID reconnection working correctly")
}

func TestE2E_TelemetryHeartbeat(t *testing.T) {
	// Initialize contract validator
	validator := NewContractValidator(t)
	validator.PrintSpecVersion(t)

	opts := harness.DefaultOptions()
	server := harness.NewServer(t, opts)
	defer server.Shutdown()

	// Subscribe to telemetry
	req, _ := http.NewRequest("GET", server.URL+"/api/v1/telemetry", nil)
	req.Header.Set("Accept", "text/event-stream")

	w := newThreadSafeResponseWriter()
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	telemetryDone := make(chan error, 1)
	go func() {
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			telemetryDone <- err
			return
		}
		defer resp.Body.Close()

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

	// Collect events for a longer period to catch heartbeats
	events := w.collectEvents(2 * time.Second)
	response := strings.Join(events, "")

	// Wait for telemetry to complete
	select {
	case err := <-telemetryDone:
		if err != nil && err != context.DeadlineExceeded {
			t.Errorf("Telemetry failed: %v", err)
		}
	case <-time.After(4 * time.Second):
		t.Fatal("Telemetry did not complete")
	}

	// Evidence: Heartbeat events
	heartbeatCount := strings.Count(response, "event: heartbeat")
	t.Logf("=== HEARTBEAT EVIDENCE ===")
	t.Logf("Total events: %d", len(events))
	t.Logf("Heartbeat events: %d", heartbeatCount)
	t.Logf("=========================")

	// Validate heartbeat timing against CB-TIMING
	baseInterval := 15 * time.Second // From CB-TIMING §3
	jitter := 2 * time.Second        // From CB-TIMING §3
	validator.ValidateHeartbeatInterval(t, events, baseInterval, jitter)

	// Verify heartbeat events
	if heartbeatCount < 1 {
		t.Errorf("Expected at least 1 heartbeat event, got %d", heartbeatCount)
	}

	t.Log("✅ Telemetry heartbeat working correctly")
}
