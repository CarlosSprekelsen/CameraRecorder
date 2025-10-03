// Package e2e provides negative test cases for the Radio Control Container API.
package e2e

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/radio-control/rcc/test/harness"
)

func TestE2E_ErrorHandling(t *testing.T) {
	// Create test harness with seeded state
	opts := harness.DefaultOptions()
	server := harness.NewServer(t, opts)
	defer server.Shutdown()

	// Evidence: Seeded state
	t.Logf("=== TEST EVIDENCE ===")
	t.Logf("Active Radio ID: %s", server.RadioManager.GetActive())
	t.Logf("Available Radios: %+v", server.RadioManager.List())
	t.Logf("===================")

	// Test invalid radio ID
	resp := httpGetWithStatus(t, server.URL+"/api/v1/radios/invalid-radio-id")
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}

	// Test power out of range
	payload := `{"powerDbm": 100}`
	resp = httpPostWithStatus(t, server.URL+"/api/v1/radios/silvus-001/power", payload)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}

	// Test channel out of range
	payload = `{"frequencyMhz": 10000.0}`
	resp = httpPostWithStatus(t, server.URL+"/api/v1/radios/silvus-001/channel", payload)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}

	// Evidence: Audit log for error cases
	auditLines, err := server.GetAuditLogs(3)
	if err != nil {
		t.Logf("Could not read audit logs: %v", err)
	} else {
		t.Logf("=== AUDIT EVIDENCE (Error Cases) ===")
		for i, line := range auditLines {
			t.Logf("Audit Line %d: %s", i+1, line)
		}
		t.Logf("===================================")
	}

	t.Log("✅ Error handling working correctly")
}

func TestE2E_InvalidJSON(t *testing.T) {
	opts := harness.DefaultOptions()
	server := harness.NewServer(t, opts)
	defer server.Shutdown()

	// Test malformed JSON
	payload := `{"powerDbm": invalid}`
	resp := httpPostWithStatus(t, server.URL+"/api/v1/radios/silvus-001/power", payload)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400 for malformed JSON, got %d", resp.StatusCode)
	}

	// Test missing required fields
	payload = `{}`
	resp = httpPostWithStatus(t, server.URL+"/api/v1/radios/silvus-001/power", payload)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400 for missing fields, got %d", resp.StatusCode)
	}

	t.Log("✅ Invalid JSON handling working correctly")
}

func TestE2E_RadioNotFound(t *testing.T) {
	opts := harness.DefaultOptions()
	server := harness.NewServer(t, opts)
	defer server.Shutdown()

	// Test operations on non-existent radio
	payload := `{"powerDbm": 10.0}`
	resp := httpPostWithStatus(t, server.URL+"/api/v1/radios/non-existent/power", payload)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404 for non-existent radio, got %d", resp.StatusCode)
	}

	// Test getting state of non-existent radio
	resp = httpGetWithStatus(t, server.URL+"/api/v1/radios/non-existent/power")
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404 for non-existent radio state, got %d", resp.StatusCode)
	}

	t.Log("✅ Radio not found handling working correctly")
}

func TestE2E_AdapterBusy(t *testing.T) {
	opts := harness.DefaultOptions()
	server := harness.NewServer(t, opts)
	defer server.Shutdown()

	// Configure SilvusMock to simulate busy state
	server.SetSilvusFaultMode("busy")

	// Evidence: Fault mode configuration
	t.Logf("=== TEST EVIDENCE ===")
	t.Logf("SilvusMock fault mode: busy")
	t.Logf("===================")

	// Test with valid power to trigger busy response
	payload := `{"powerDbm": 10.0}`
	resp := httpPostWithStatus(t, server.URL+"/api/v1/radios/silvus-001/power", payload)

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503 for busy adapter, got %d", resp.StatusCode)
	}

	// Verify error response format
	var errorResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
		t.Fatalf("Failed to decode error response: %v", err)
	}

	// Check that error response has expected structure
	if _, ok := errorResp["error"]; !ok {
		t.Error("Expected error response to have 'error' field")
	}

	// Evidence: Audit log for busy case
	auditLines, err := server.GetAuditLogs(1)
	if err != nil {
		t.Logf("Could not read audit logs: %v", err)
	} else {
		t.Logf("=== AUDIT EVIDENCE (Busy Case) ===")
		for i, line := range auditLines {
			t.Logf("Audit Line %d: %s", i+1, line)
		}
		t.Logf("==================================")
	}

	t.Log("✅ Adapter busy handling working correctly")
}
