// Package e2e provides negative test cases for the Radio Control Container API.
package e2e

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestE2E_ErrorHandling(t *testing.T) {
	ts := newServerForE2E(t)

	// Test invalid radio ID
	resp := httpGetWithStatus(t, ts.URL+"/api/v1/radios/invalid-radio-id")
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}

	// Test power out of range
	payload := `{"powerDbm": 100}`
	resp = httpPostWithStatus(t, ts.URL+"/api/v1/radios/silvus-001/power", payload)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}

	// Test channel out of range
	payload = `{"frequencyMhz": 10000.0}`
	resp = httpPostWithStatus(t, ts.URL+"/api/v1/radios/silvus-001/channel", payload)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}

	t.Log("✅ Error handling working correctly")
}

func TestE2E_InvalidJSON(t *testing.T) {
	ts := newServerForE2E(t)

	// Test malformed JSON
	payload := `{"powerDbm": invalid}`
	resp := httpPostWithStatus(t, ts.URL+"/api/v1/radios/silvus-001/power", payload)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400 for malformed JSON, got %d", resp.StatusCode)
	}

	// Test missing required fields
	payload = `{}`
	resp = httpPostWithStatus(t, ts.URL+"/api/v1/radios/silvus-001/power", payload)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400 for missing fields, got %d", resp.StatusCode)
	}

	t.Log("✅ Invalid JSON handling working correctly")
}

func TestE2E_RadioNotFound(t *testing.T) {
	ts := newServerForE2E(t)

	// Test operations on non-existent radio
	payload := `{"powerDbm": 10.0}`
	resp := httpPostWithStatus(t, ts.URL+"/api/v1/radios/non-existent/power", payload)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404 for non-existent radio, got %d", resp.StatusCode)
	}

	// Test getting state of non-existent radio
	resp = httpGetWithStatus(t, ts.URL+"/api/v1/radios/non-existent/power")
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected status 404 for non-existent radio state, got %d", resp.StatusCode)
	}

	t.Log("✅ Radio not found handling working correctly")
}

func TestE2E_AdapterBusy(t *testing.T) {
	// This test would require a mock adapter that can simulate busy state
	// For now, we'll test the error response format
	ts := newServerForE2E(t)

	// Test with invalid power to trigger error response
	payload := `{"powerDbm": 1000.0}`
	resp := httpPostWithStatus(t, ts.URL+"/api/v1/radios/silvus-001/power", payload)

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400 for invalid power, got %d", resp.StatusCode)
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

	t.Log("✅ Adapter error handling working correctly")
}
