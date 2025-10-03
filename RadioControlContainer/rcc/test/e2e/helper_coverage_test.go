// Package e2e provides helper function coverage tests.
// This file tests helper functions to improve E2E coverage.
package e2e

import (
	"testing"
	"time"
)

func TestHelperFunctions_mustHaveNumber(t *testing.T) {
	// Test positive case - valid number
	data := map[string]interface{}{
		"value": 42.5,
	}
	mustHaveNumber(t, data, "value", 42.5)

	// Test negative case - wrong number
	data = map[string]interface{}{
		"value": 42.5,
	}
	// This should fail but not crash
	t.Run("wrong_number", func(t *testing.T) {
		mustHaveNumber(t, data, "value", 10.0)
	})

	// Test negative case - not a number
	data = map[string]interface{}{
		"value": "not_a_number",
	}
	t.Run("not_number", func(t *testing.T) {
		mustHaveNumber(t, data, "value", 42.5)
	})

	// Test negative case - missing field
	data = map[string]interface{}{
		"other": 42.5,
	}
	t.Run("missing_field", func(t *testing.T) {
		mustHaveNumber(t, data, "value", 42.5)
	})
}

func TestHelperFunctions_getJSONPath(t *testing.T) {
	// Test nested path success
	data := map[string]interface{}{
		"level1": map[string]interface{}{
			"level2": map[string]interface{}{
				"value": "found",
			},
		},
	}
	result := getJSONPath(data, "level1.level2.value")
	if result != "found" {
		t.Errorf("Expected 'found', got %v", result)
	}

	// Test missing path failure
	result = getJSONPath(data, "level1.level2.missing")
	if result != nil {
		t.Errorf("Expected nil, got %v", result)
	}

	// Test single level path
	data = map[string]interface{}{
		"value": "direct",
	}
	result = getJSONPath(data, "value")
	if result != "direct" {
		t.Errorf("Expected 'direct', got %v", result)
	}

	// Test array access (if supported)
	data = map[string]interface{}{
		"items": []interface{}{
			map[string]interface{}{
				"name": "item1",
			},
		},
	}
	result = getJSONPath(data, "items.0.name")
	// This might not work depending on implementation
	t.Logf("Array access result: %v", result)
}

func TestHelperFunctions_threadSafeResponseWriter(t *testing.T) {
	// Test WriteHeader path
	w := newThreadSafeResponseWriter()
	
	// Test WriteHeader
	w.WriteHeader(404)
	if w.statusCode != 404 {
		t.Errorf("Expected status code 404, got %d", w.statusCode)
	}

	// Test Write
	data := []byte("test data")
	n, err := w.Write(data)
	if err != nil {
		t.Errorf("Write failed: %v", err)
	}
	if n != len(data) {
		t.Errorf("Expected to write %d bytes, wrote %d", len(data), n)
	}

	// Test Header
	headers := w.Header()
	if headers == nil {
		t.Error("Expected non-nil headers")
	}

	// Test collectEvents
	events := w.collectEvents(100 * time.Millisecond)
	if len(events) == 0 {
		t.Error("Expected at least one event")
	}
}

func TestHelperFunctions_mustHave(t *testing.T) {
	// Test positive case
	data := map[string]interface{}{
		"key": "value",
	}
	mustHave(t, data, "key", "value")

	// Test negative case - wrong value
	t.Run("wrong_value", func(t *testing.T) {
		mustHave(t, data, "key", "wrong")
	})

	// Test negative case - missing key
	t.Run("missing_key", func(t *testing.T) {
		mustHave(t, data, "missing", "value")
	})
}
