// Package e2e provides negative test cases for the Radio Control Container API.
// This file implements black-box testing using only HTTP and contract validation.
package e2e

import (
	"testing"

	"github.com/radio-control/rcc/test/harness"
)

func TestE2E_ErrorHandling(t *testing.T) {
	// Initialize contract validator
	validator := NewContractValidator(t)
	validator.PrintSpecVersion(t)

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
	validator.ValidateErrorResponse(t, resp, "NOT_FOUND")

	// Test power out of range
	payload := `{"powerDbm": 100}`
	resp = httpPostWithStatus(t, server.URL+"/api/v1/radios/silvus-001/power", payload)
	validator.ValidateErrorResponse(t, resp, "INVALID_RANGE")

	// Test channel out of range
	payload = `{"frequencyMhz": 10000.0}`
	resp = httpPostWithStatus(t, server.URL+"/api/v1/radios/silvus-001/channel", payload)
	validator.ValidateErrorResponse(t, resp, "INVALID_RANGE")

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
	// Initialize contract validator
	validator := NewContractValidator(t)
	validator.PrintSpecVersion(t)

	opts := harness.DefaultOptions()
	server := harness.NewServer(t, opts)
	defer server.Shutdown()

	// Test malformed JSON
	payload := `{"powerDbm": invalid}`
	resp := httpPostWithStatus(t, server.URL+"/api/v1/radios/silvus-001/power", payload)
	validator.ValidateHTTPResponse(t, resp, 400)

	// Test missing required fields
	payload = `{}`
	resp = httpPostWithStatus(t, server.URL+"/api/v1/radios/silvus-001/power", payload)
	validator.ValidateHTTPResponse(t, resp, 400)

	t.Log("✅ Invalid JSON handling working correctly")
}

func TestE2E_RadioNotFound(t *testing.T) {
	// Initialize contract validator
	validator := NewContractValidator(t)
	validator.PrintSpecVersion(t)

	opts := harness.DefaultOptions()
	server := harness.NewServer(t, opts)
	defer server.Shutdown()

	// Test operations on non-existent radio
	payload := `{"powerDbm": 10.0}`
	resp := httpPostWithStatus(t, server.URL+"/api/v1/radios/non-existent/power", payload)
	validator.ValidateErrorResponse(t, resp, "NOT_FOUND")

	// Test getting state of non-existent radio
	resp = httpGetWithStatus(t, server.URL+"/api/v1/radios/non-existent/power")
	validator.ValidateErrorResponse(t, resp, "NOT_FOUND")

	t.Log("✅ Radio not found handling working correctly")
}

func TestE2E_AdapterBusy(t *testing.T) {
	// Initialize contract validator
	validator := NewContractValidator(t)
	validator.PrintSpecVersion(t)

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
	validator.ValidateErrorResponse(t, resp, "BUSY")

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
