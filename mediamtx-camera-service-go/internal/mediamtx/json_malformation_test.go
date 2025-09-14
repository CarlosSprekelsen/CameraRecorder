/*
JSON Malformation Tests - Dangerous Bug Detection

Requirements Coverage:
- REQ-MTX-001: MediaMTX service integration
- REQ-MTX-002: Stream management capabilities
- REQ-MTX-003: Path creation and deletion
- REQ-MTX-004: Health monitoring

Test Categories: Unit (focused on dangerous bug detection)
API Documentation Reference: docs/api/json_rpc_methods.md

Purpose: These tests are designed to catch dangerous bugs through systematic
JSON malformation testing, not just achieve coverage. FAIL is OK if it identifies
a real bug that needs to be fixed.
*/

package mediamtx

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestJSONParsingErrors_DangerousBugs tests JSON parsing functions with malformed data
// that can catch dangerous bugs in JSON parsing
func TestJSONParsingErrors_DangerousBugs(t *testing.T) {
	EnsureSequentialExecution(t)
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Test JSON parsing errors that can catch dangerous bugs
	helper.TestJSONParsingErrors(t)
}

// TestJSONParsingPanicProtection_DangerousBugs tests that JSON parsing functions
// don't panic with malformed data that could cause crashes
func TestJSONParsingPanicProtection_DangerousBugs(t *testing.T) {
	EnsureSequentialExecution(t)
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Test panic protection that can catch dangerous bugs
	helper.TestJSONParsingPanicProtection(t)
}

// TestJSONParsingFunctions_DangerousBugs tests individual JSON parsing functions
// that have 0% coverage and could hide dangerous bugs
func TestJSONParsingFunctions_DangerousBugs(t *testing.T) {
	t.Run("parseStreamResponse_EmptyData_Bug", func(t *testing.T) {
		// Test the specific function that has 0% coverage
		// This could catch dangerous bugs in empty response handling
		_, err := parseStreamResponse([]byte(""))

		// This should produce an error for empty data
		require.Error(t, err, "parseStreamResponse should error on empty data")
		t.Logf("parseStreamResponse correctly handled empty data: %v", err)
	})

	t.Run("parseStreamResponse_NullData_Bug", func(t *testing.T) {
		// Test null data handling
		_, err := parseStreamResponse([]byte("null"))

		// This should produce an error for null data
		require.Error(t, err, "parseStreamResponse should error on null data")
		t.Logf("parseStreamResponse correctly handled null data: %v", err)
	})

	t.Run("parseStreamResponse_MalformedJSON_Bug", func(t *testing.T) {
		// Test malformed JSON handling
		_, err := parseStreamResponse([]byte(`{"invalid": json}`))

		// This should produce an error for malformed JSON
		require.Error(t, err, "parseStreamResponse should error on malformed JSON")
		t.Logf("parseStreamResponse correctly handled malformed JSON: %v", err)
	})

	t.Run("determineStatus_Function_Bug", func(t *testing.T) {
		// Test the determineStatus function that has 0% coverage
		// This could catch dangerous bugs in status determination

		// Test with true
		status := determineStatus(true)
		require.Equal(t, "READY", status, "determineStatus(true) should return READY")

		// Test with false
		status = determineStatus(false)
		require.Equal(t, "PENDING", status, "determineStatus(false) should return PENDING")

		t.Logf("determineStatus function works correctly")
	})

	t.Run("parseMetricsResponse_Function_Bug", func(t *testing.T) {
		// Test the parseMetricsResponse function that has 0% coverage
		// This could catch dangerous bugs in metrics parsing

		// Test with empty data
		_, err := parseMetricsResponse([]byte(""))
		require.Error(t, err, "parseMetricsResponse should error on empty data")

		// Test with malformed JSON
		_, err = parseMetricsResponse([]byte(`{"invalid": json}`))
		require.Error(t, err, "parseMetricsResponse should error on malformed JSON")

		// Test with valid JSON structure
		validJSON := []byte(`{"connections": 5, "sessions": 3}`)
		metrics, err := parseMetricsResponse(validJSON)
		require.NoError(t, err, "parseMetricsResponse should handle valid JSON")
		require.NotNil(t, metrics, "parseMetricsResponse should return metrics")

		t.Logf("parseMetricsResponse function works correctly")
	})

	t.Run("marshalCreateStreamRequest_Function_Bug", func(t *testing.T) {
		// Test the marshalCreateStreamRequest function that has 0% coverage
		// This could catch dangerous bugs in request marshaling

		// Test with valid parameters
		data, err := marshalCreateStreamRequest("test_stream", "rtsp://localhost:8554/test")
		require.NoError(t, err, "marshalCreateStreamRequest should handle valid parameters")
		require.NotNil(t, data, "marshalCreateStreamRequest should return data")
		require.Contains(t, string(data), "test_stream", "marshaled data should contain stream name")
		require.Contains(t, string(data), "rtsp://localhost:8554/test", "marshaled data should contain source")

		t.Logf("marshalCreateStreamRequest function works correctly")
	})

	t.Run("marshalUpdatePathRequest_Function_Bug", func(t *testing.T) {
		// Test the marshalUpdatePathRequest function that has 0% coverage
		// This could catch dangerous bugs in path update marshaling

		// Test with valid parameters
		config := &PathConf{
			Source: "rtsp://localhost:8554/updated",
		}
		data, err := marshalUpdatePathRequest(config)
		require.NoError(t, err, "marshalUpdatePathRequest should handle valid parameters")
		require.NotNil(t, data, "marshalUpdatePathRequest should return data")
		require.Contains(t, string(data), "rtsp://localhost:8554/updated", "marshaled data should contain updated source")

		t.Logf("marshalUpdatePathRequest function works correctly")
	})
}

// TestJSONParsingEdgeCases_DangerousBugs tests edge cases that could cause
// dangerous bugs in JSON parsing
func TestJSONParsingEdgeCases_DangerousBugs(t *testing.T) {
	t.Run("JSON_Parsing_Edge_Cases", func(t *testing.T) {
		// Test edge cases that could cause dangerous bugs
		edgeCases := []struct {
			name     string
			data     []byte
			expected string
		}{
			{
				name:     "very_large_json",
				data:     []byte(`{"items": [], "large_field": "` + string(make([]byte, 1000000)) + `"}`),
				expected: "should handle gracefully",
			},
			{
				name:     "json_with_null_bytes",
				data:     []byte(`{"items": [], "null_field": "test\x00null\x00byte"}`),
				expected: "should handle gracefully",
			},
			{
				name:     "json_with_unicode_issues",
				data:     []byte(`{"items": [], "unicode": "test\u0000\u0001\u0002"}`),
				expected: "should handle gracefully",
			},
			{
				name:     "json_with_deep_nesting",
				data:     []byte(`{"items": [], "nested": {"a": {"b": {"c": {"d": {"e": {"f": {"g": {"h": {"i": {"j": "deep"}}}}}}}}}}`),
				expected: "should handle gracefully",
			},
		}

		for _, testCase := range edgeCases {
			t.Run(testCase.name, func(t *testing.T) {
				t.Logf("Testing edge case: %s - %s", testCase.name, testCase.expected)

				// Test that parsing doesn't panic or cause crashes
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("🚨 BUG DETECTED: JSON parsing caused panic with edge case %s: %v", testCase.name, r)
					}
				}()

				// Test all parsing functions with edge case data
				_, err1 := parsePathListResponse(testCase.data)
				_, err2 := parseStreamResponse(testCase.data)
				_, err3 := parseHealthResponse(testCase.data)
				_, err4 := parsePathConfListResponse(testCase.data)

				// We don't care about errors here, just that no panic occurred
				t.Logf("Edge case %s handled without panic (errors: %v, %v, %v, %v)",
					testCase.name, err1, err2, err3, err4)
			})
		}
	})
}
