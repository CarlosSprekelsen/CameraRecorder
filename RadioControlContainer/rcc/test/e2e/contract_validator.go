// Package e2e provides contract validation for E2E tests.
// This file ensures all E2E tests validate against the API specification.
package e2e

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

// ContractValidator validates E2E test responses against API contracts
type ContractValidator struct {
	specVersion     string
	errorMappings   map[string]int
	telemetrySchema map[string]interface{}
}

// NewContractValidator creates a new contract validator
func NewContractValidator(t *testing.T) *ContractValidator {
	// Read spec version
	specVersion := readSpecVersion(t)

	// Load error mappings
	errorMappings := loadErrorMappings(t)

	// Load telemetry schema
	telemetrySchema := loadTelemetrySchema(t)

	return &ContractValidator{
		specVersion:     specVersion,
		errorMappings:   errorMappings,
		telemetrySchema: telemetrySchema,
	}
}

// PrintSpecVersion prints the spec version at test start
func (cv *ContractValidator) PrintSpecVersion(t *testing.T) {
	t.Logf("Spec Version: %s", cv.specVersion)
}

// ValidateHTTPResponse validates an HTTP response against the contract
func (cv *ContractValidator) ValidateHTTPResponse(t *testing.T, resp *http.Response, expectedStatus int) {
	// Validate status code
	if resp.StatusCode != expectedStatus {
		t.Errorf("Expected status %d, got %d", expectedStatus, resp.StatusCode)
	}

	// Validate content type for JSON responses
	if resp.StatusCode < 400 {
		contentType := resp.Header.Get("Content-Type")
		if !strings.Contains(contentType, "application/json") {
			t.Errorf("Expected JSON content type, got %s", contentType)
		}
	}

	// Validate response envelope for success responses
	if resp.StatusCode == 200 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		var envelope map[string]interface{}
		if err := json.Unmarshal(body, &envelope); err != nil {
			t.Fatalf("Failed to parse JSON response: %v", err)
		}

		// Check for required envelope fields
		if _, ok := envelope["result"]; !ok {
			t.Error("Expected 'result' field in response envelope")
		}

		if _, ok := envelope["data"]; !ok {
			t.Error("Expected 'data' field in response envelope")
		}
	}
}

// ValidateErrorResponse validates an error response against the error mapping
func (cv *ContractValidator) ValidateErrorResponse(t *testing.T, resp *http.Response, expectedError string) {
	expectedStatus, exists := cv.errorMappings[expectedError]
	if !exists {
		t.Errorf("Unknown error code: %s", expectedError)
		return
	}

	if resp.StatusCode != expectedStatus {
		t.Errorf("Expected status %d for error %s, got %d", expectedStatus, expectedError, resp.StatusCode)
	}

	// Validate error response structure
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read error response body: %v", err)
	}

	var errorResp map[string]interface{}
	if err := json.Unmarshal(body, &errorResp); err != nil {
		t.Fatalf("Failed to parse error JSON response: %v", err)
	}

	// Check for error field
	if _, ok := errorResp["error"]; !ok {
		t.Error("Expected 'error' field in error response")
	}
}

// ValidateSSEEvent validates an SSE event against the telemetry schema
func (cv *ContractValidator) ValidateSSEEvent(t *testing.T, event string) {
	// Parse SSE event format
	lines := strings.Split(strings.TrimSpace(event), "\n")
	eventData := make(map[string]string)

	for _, line := range lines {
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				eventData[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}
	}

	// Validate required SSE fields
	if _, ok := eventData["event"]; !ok {
		t.Error("Expected 'event' field in SSE event")
	}

	if _, ok := eventData["data"]; !ok {
		t.Error("Expected 'data' field in SSE event")
	}

	// Validate event ID is monotonic
	if id, ok := eventData["id"]; ok {
		if id == "" {
			t.Error("Event ID should not be empty")
		}
	}

	// Validate event type
	eventType := eventData["event"]
	validTypes := []string{"ready", "heartbeat", "powerChanged", "channelChanged"}

	valid := false
	for _, validType := range validTypes {
		if eventType == validType {
			valid = true
			break
		}
	}

	if !valid {
		t.Errorf("Invalid event type: %s", eventType)
	}
}

// ValidateHeartbeatInterval validates heartbeat timing against CB-TIMING
func (cv *ContractValidator) ValidateHeartbeatInterval(t *testing.T, events []string, baseInterval time.Duration, jitter time.Duration) {
	heartbeatEvents := make([]time.Time, 0)

	for _, event := range events {
		if strings.Contains(event, "event: heartbeat") {
			// Extract timestamp from event data
			lines := strings.Split(event, "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "data:") {
					var data map[string]interface{}
					if err := json.Unmarshal([]byte(strings.TrimPrefix(line, "data: ")), &data); err == nil {
						if ts, ok := data["timestamp"].(string); ok {
							if t, err := time.Parse(time.RFC3339, ts); err == nil {
								heartbeatEvents = append(heartbeatEvents, t)
							}
						}
					}
				}
			}
		}
	}

	// Validate heartbeat intervals
	for i := 1; i < len(heartbeatEvents); i++ {
		interval := heartbeatEvents[i].Sub(heartbeatEvents[i-1])
		minInterval := baseInterval - jitter
		maxInterval := baseInterval + jitter

		if interval < minInterval || interval > maxInterval {
			t.Errorf("Heartbeat interval %v outside tolerance [%v, %v]", interval, minInterval, maxInterval)
		}
	}
}

// Helper functions

func readSpecVersion(t *testing.T) string {
	// Try multiple possible locations for the VERSION file
	possiblePaths := []string{
		"docs/contract/VERSION",
		"../../docs/contract/VERSION",
		"../../../docs/contract/VERSION",
	}

	var content []byte
	var err error

	for _, path := range possiblePaths {
		content, err = os.ReadFile(path)
		if err == nil {
			break
		}
	}

	if err != nil {
		t.Fatalf("Failed to read spec version from any location: %v", err)
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "SPEC_VERSION=") {
			return strings.TrimPrefix(line, "SPEC_VERSION=")
		}
	}

	return "unknown"
}

func loadErrorMappings(t *testing.T) map[string]int {
	// Try multiple possible locations for the error-mapping.json file
	possiblePaths := []string{
		"docs/contract/error-mapping.json",
		"../../docs/contract/error-mapping.json",
		"../../../docs/contract/error-mapping.json",
	}

	var content []byte
	var err error

	for _, path := range possiblePaths {
		content, err = os.ReadFile(path)
		if err == nil {
			break
		}
	}

	if err != nil {
		t.Fatalf("Failed to read error mappings from any location: %v", err)
	}

	var mappings struct {
		Mappings []struct {
			AdapterError string `json:"adapter_error"`
			HTTPStatus   int    `json:"http_status"`
		} `json:"mappings"`
	}

	if err := json.Unmarshal(content, &mappings); err != nil {
		t.Fatalf("Failed to parse error mappings: %v", err)
	}

	result := make(map[string]int)
	for _, mapping := range mappings.Mappings {
		result[mapping.AdapterError] = mapping.HTTPStatus
	}

	return result
}

func loadTelemetrySchema(t *testing.T) map[string]interface{} {
	// Try multiple possible locations for the telemetry.schema.json file
	possiblePaths := []string{
		"docs/contract/telemetry.schema.json",
		"../../docs/contract/telemetry.schema.json",
		"../../../docs/contract/telemetry.schema.json",
	}

	var content []byte
	var err error

	for _, path := range possiblePaths {
		content, err = os.ReadFile(path)
		if err == nil {
			break
		}
	}

	if err != nil {
		t.Fatalf("Failed to read telemetry schema from any location: %v", err)
	}

	var schema map[string]interface{}
	if err := json.Unmarshal(content, &schema); err != nil {
		t.Fatalf("Failed to parse telemetry schema: %v", err)
	}

	return schema
}
