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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestJSONParsingErrors_DangerousBugs tests JSON parsing functions with malformed data
// that can catch dangerous bugs in JSON parsing
func TestJSONParsingErrors_DangerousBugs(t *testing.T) {
	// PROGRESSIVE READINESS: No sequential execution - enables parallelism
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Use the scenario registry directly for comprehensive testing
	registry := NewJSONScenarioRegistry()

	// Test all response types with their scenarios
	responseTypes := []string{"path_list", "stream", "paths", "health"}

	for _, responseType := range responseTypes {
		t.Run(responseType+"_scenarios", func(t *testing.T) {
			scenarios := registry.GetScenarios(responseType)

			for _, scenario := range scenarios {
				t.Run(scenario.Name, func(t *testing.T) {
					t.Logf("Testing JSON scenario: %s - %s", scenario.Name, scenario.Description)

					// Test the appropriate parsing function based on response type
					var err error
					switch responseType {
					case "path_list":
						_, err = parsePathListResponse(scenario.JSONData)
					case "stream":
						_, err = parseStreamResponse(scenario.JSONData)
					case "paths":
						_, err = parsePathConfListResponse(scenario.JSONData)
					case "health":
						_, err = parseHealthResponse(scenario.JSONData)
					}

					// Verify expected behavior
					if scenario.ExpectError {
						require.Error(t, err, "Scenario %s should produce an error", scenario.Name)
						if scenario.ErrorMsg != "" {
							assert.Contains(t, err.Error(), scenario.ErrorMsg,
								"Error message should contain expected text for scenario %s", scenario.Name)
						}
						t.Logf("Scenario %s correctly produced expected error: %v", scenario.Name, err)
					} else {
						if err != nil {
							t.Errorf("BUG DETECTED: Scenario %s should be handled gracefully but got error: %v", scenario.Name, err)
						} else {
							t.Logf("Scenario %s handled gracefully (no error)", scenario.Name)
						}
					}
				})
			}
		})
	}
}

// TestJSONParsingPanicProtection_DangerousBugs tests that JSON parsing functions
// don't panic with malformed data that could cause crashes
func TestJSONParsingPanicProtection_DangerousBugs(t *testing.T) {
	// PROGRESSIVE READINESS: No sequential execution - enables parallelism
	helper := NewMediaMTXTestHelper(t, nil)
	defer helper.Cleanup(t)

	// Test panic protection with edge cases

	// Test panic protection with edge cases
	edgeCases := []struct {
		name     string
		data     []byte
		expected string
	}{
		{
			name:     "very_large_json",
			data:     []byte(`{"items": [], "large_field": "` + strings.Repeat("x", 1000000) + `"}`),
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
					t.Errorf("BUG DETECTED: JSON parsing caused panic with edge case %s: %v", testCase.name, r)
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
}

// TestJSONParsingFunctions_DangerousBugs tests individual JSON parsing functions
// that have 0% coverage and could hide dangerous bugs using scenario registry
func TestJSONParsingFunctions_DangerousBugs(t *testing.T) {
	registry := NewJSONScenarioRegistry()

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

	// Test all parsing functions using scenario registry for comprehensive coverage
	t.Run("comprehensive_parsing_coverage", func(t *testing.T) {
		responseTypes := []string{"path_list", "stream", "paths", "health"}

		for _, responseType := range responseTypes {
			t.Run(responseType+"_comprehensive", func(t *testing.T) {
				scenarios := registry.GetScenarios(responseType)

				for _, scenario := range scenarios {
					t.Run(scenario.Name, func(t *testing.T) {
						// Test the appropriate parsing function based on response type
						var err error
						switch responseType {
						case "path_list":
							_, err = parsePathListResponse(scenario.JSONData)
						case "stream":
							_, err = parseStreamResponse(scenario.JSONData)
						case "paths":
							_, err = parsePathConfListResponse(scenario.JSONData)
						case "health":
							_, err = parseHealthResponse(scenario.JSONData)
						}

						// Verify expected behavior matches scenario
						if scenario.ExpectError {
							require.Error(t, err, "Scenario %s should produce an error", scenario.Name)
						} else {
							// For scenarios that should handle gracefully, we don't require no error
							// but we do want to ensure no panic occurred
							t.Logf("Scenario %s handled (error: %v)", scenario.Name, err)
						}
					})
				}
			})
		}
	})
}

// TestJSONParsingEdgeCases_DangerousBugs tests edge cases that could cause
// dangerous bugs in JSON parsing using scenario registry
func TestJSONParsingEdgeCases_DangerousBugs(t *testing.T) {
	registry := NewJSONScenarioRegistry()

	t.Run("JSON_Parsing_Edge_Cases", func(t *testing.T) {
		// Test edge cases that could cause dangerous bugs
		edgeCases := []struct {
			name     string
			data     []byte
			expected string
		}{
			{
				name:     "very_large_json",
				data:     []byte(`{"items": [], "large_field": "` + strings.Repeat("x", 1000000) + `"}`),
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
						t.Errorf("BUG DETECTED: JSON parsing caused panic with edge case %s: %v", testCase.name, r)
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

	// Test edge cases from scenario registry
	t.Run("registry_edge_cases", func(t *testing.T) {
		responseTypes := []string{"path_list", "stream", "paths", "health"}

		for _, responseType := range responseTypes {
			t.Run(responseType+"_edge_cases", func(t *testing.T) {
				scenarios := registry.GetScenarios(responseType)

				// Focus on edge case scenarios
				edgeCaseNames := []string{
					"json_with_very_large_strings",
					"json_with_unicode_issues",
					"json_with_special_characters",
					"json_with_deep_nesting",
				}

				for _, scenario := range scenarios {
					for _, edgeCaseName := range edgeCaseNames {
						if scenario.Name == edgeCaseName {
							t.Run(scenario.Name, func(t *testing.T) {
								// Test that parsing doesn't panic with edge case data
								defer func() {
									if r := recover(); r != nil {
										t.Errorf("BUG DETECTED: JSON parsing caused panic with edge case %s: %v", scenario.Name, r)
									}
								}()

								// Test the appropriate parsing function based on response type
								var err error
								switch responseType {
								case "path_list":
									_, err = parsePathListResponse(scenario.JSONData)
								case "stream":
									_, err = parseStreamResponse(scenario.JSONData)
								case "paths":
									_, err = parsePathConfListResponse(scenario.JSONData)
								case "health":
									_, err = parseHealthResponse(scenario.JSONData)
								}

								// We don't care about errors here, just that no panic occurred
								t.Logf("Edge case %s handled without panic (error: %v)", scenario.Name, err)
							})
						}
					}
				}
			})
		}
	})
}
