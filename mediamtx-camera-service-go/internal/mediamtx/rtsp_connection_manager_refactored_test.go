/*
MediaMTX RTSP Connection Manager Tests - Refactored with Progressive Readiness

Requirements Coverage:
- REQ-MTX-001: MediaMTX service integration
- REQ-MTX-002: Stream management capabilities
- REQ-MTX-003: Path creation and deletion
- REQ-MTX-004: Health monitoring

Test Categories: Unit (using real MediaMTX server)
API Documentation Reference: docs/api/swagger.json

Refactored from rtsp_connection_manager_test.go (745 lines → ~200 lines)
Eliminates massive duplication using RTSPConnectionAsserter
*/

package mediamtx

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestNewRTSPConnectionManager_ReqMTX001_Refactored tests RTSP connection manager creation
func TestNewRTSPConnectionManager_ReqMTX001_Refactored(t *testing.T) {
	// REQ-MTX-001: MediaMTX service integration
	asserter := NewRTSPConnectionAsserter(t)
	defer asserter.Cleanup()

	// Test server health first
	err := asserter.GetHelper().TestMediaMTXHealth(t)
	require.NoError(t, err, "MediaMTX server should be healthy")

	// Validate RTSP manager creation
	asserter.AssertRTSPManagerCreation()
}

// TestRTSPConnectionManager_ListConnections_ReqMTX002_Refactored tests RTSP connection listing
func TestRTSPConnectionManager_ListConnections_ReqMTX002_Refactored(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	asserter := NewRTSPConnectionAsserter(t)
	defer asserter.Cleanup()

	// Test listing connections with different pagination
	testCases := []struct {
		name         string
		page         int
		itemsPerPage int
	}{
		{"first_page", 0, 10},
		{"second_page", 1, 5},
		{"large_page", 0, 100},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			asserter.AssertListConnections(tc.page, tc.itemsPerPage)
		})
	}
}

// TestRTSPConnectionManager_ListSessions_ReqMTX002_Refactored tests RTSP session listing
func TestRTSPConnectionManager_ListSessions_ReqMTX002_Refactored(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities
	asserter := NewRTSPConnectionAsserter(t)
	defer asserter.Cleanup()

	// Test listing sessions with different pagination
	testCases := []struct {
		name         string
		page         int
		itemsPerPage int
	}{
		{"first_page", 0, 10},
		{"second_page", 1, 5},
		{"large_page", 0, 100},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			asserter.AssertListSessions(tc.page, tc.itemsPerPage)
		})
	}
}

// TestRTSPConnectionManager_GetConnectionHealth_ReqMTX004_Refactored tests RTSP connection health
func TestRTSPConnectionManager_GetConnectionHealth_ReqMTX004_Refactored(t *testing.T) {
	// REQ-MTX-004: Health monitoring
	asserter := NewRTSPConnectionAsserter(t)
	defer asserter.Cleanup()

	// Test health checking
	asserter.AssertGetConnectionHealth()
}

// TestRTSPConnectionManager_GetConnectionMetrics_ReqMTX004_Refactored tests RTSP connection metrics
func TestRTSPConnectionManager_GetConnectionMetrics_ReqMTX004_Refactored(t *testing.T) {
	// REQ-MTX-004: Health monitoring
	asserter := NewRTSPConnectionAsserter(t)
	defer asserter.Cleanup()

	// Test metrics retrieval
	asserter.AssertGetConnectionMetrics()
}

// TestRTSPConnectionManager_Configuration_ReqMTX003_Refactored tests RTSP configuration management
func TestRTSPConnectionManager_Configuration_ReqMTX003_Refactored(t *testing.T) {
	// REQ-MTX-003: Path creation and deletion (configuration management)
	asserter := NewRTSPConnectionAsserter(t)
	defer asserter.Cleanup()

	asserter.AssertConfigurationManagement()
}

// TestRTSPConnectionManager_ErrorHandling_ReqMTX004_Refactored tests RTSP error handling
func TestRTSPConnectionManager_ErrorHandling_ReqMTX004_Refactored(t *testing.T) {
	// REQ-MTX-004: Health monitoring (error handling)
	asserter := NewRTSPConnectionAsserter(t)
	defer asserter.Cleanup()

	asserter.AssertErrorHandling()
}

// TestRTSPConnectionManager_Performance_ReqMTX002_Refactored tests RTSP performance
func TestRTSPConnectionManager_Performance_ReqMTX002_Refactored(t *testing.T) {
	// REQ-MTX-002: Stream management capabilities (performance)
	asserter := NewRTSPConnectionAsserter(t)
	defer asserter.Cleanup()

	asserter.AssertPerformanceTest()
}

// TestRTSPConnectionManager_RealMediaMTXServer_Refactored tests integration with real MediaMTX server
func TestRTSPConnectionManager_RealMediaMTXServer_Refactored(t *testing.T) {
	// Integration test with real MediaMTX server
	asserter := NewRTSPConnectionAsserter(t)
	defer asserter.Cleanup()

	// Test basic operations with real server
	asserter.AssertListConnections(0, 10)
	asserter.AssertListSessions(0, 10)
}

// TestRTSPConnectionManager_ConfigurationScenarios_Refactored tests various configuration scenarios
func TestRTSPConnectionManager_ConfigurationScenarios_Refactored(t *testing.T) {
	asserter := NewRTSPConnectionAsserter(t)
	defer asserter.Cleanup()

	// Test multiple configuration scenarios
	t.Run("basic_configuration", func(t *testing.T) {
		asserter.AssertConfigurationManagement()
	})

	t.Run("configuration_with_connections", func(t *testing.T) {
		asserter.AssertListConnections(0, 10)
		asserter.AssertConfigurationManagement()
	})
}

// TestRTSPConnectionManager_ErrorScenarios_Refactored tests various error scenarios
func TestRTSPConnectionManager_ErrorScenarios_Refactored(t *testing.T) {
	asserter := NewRTSPConnectionAsserter(t)
	defer asserter.Cleanup()

	// Test multiple error scenarios
	t.Run("invalid_connection_health", func(t *testing.T) {
		asserter.AssertErrorHandling()
	})

	t.Run("invalid_pagination", func(t *testing.T) {
		asserter.AssertInputValidation()
	})
}

// TestRTSPConnectionManager_ConcurrentAccess_Refactored tests concurrent access to RTSP manager
func TestRTSPConnectionManager_ConcurrentAccess_Refactored(t *testing.T) {
	asserter := NewRTSPConnectionAsserter(t)
	defer asserter.Cleanup()

	asserter.AssertConcurrentAccess()
}

// TestRTSPConnectionManager_StressTest_Refactored tests stress scenarios
func TestRTSPConnectionManager_StressTest_Refactored(t *testing.T) {
	asserter := NewRTSPConnectionAsserter(t)
	defer asserter.Cleanup()

	asserter.AssertStressTest()
}

// TestRTSPConnectionManager_IntegrationWithController_Refactored tests integration with controller
func TestRTSPConnectionManager_IntegrationWithController_Refactored(t *testing.T) {
	asserter := NewRTSPConnectionAsserter(t)
	defer asserter.Cleanup()

	asserter.AssertIntegrationWithController()
}

// TestRTSPConnectionManager_InputValidation_Refactored tests input validation
func TestRTSPConnectionManager_InputValidation_Refactored(t *testing.T) {
	// REQ-MTX-007: Error handling and recovery
	asserter := NewRTSPConnectionAsserter(t)
	defer asserter.Cleanup()

	asserter.AssertInputValidation()
}

// TestRTSPConnectionManager_ErrorScenarios_DangerousBugs_Refactored tests dangerous bug scenarios
func TestRTSPConnectionManager_ErrorScenarios_DangerousBugs_Refactored(t *testing.T) {
	asserter := NewRTSPConnectionAsserter(t)
	defer asserter.Cleanup()

	// Test dangerous scenarios that could cause system issues
	t.Run("malicious_input", func(t *testing.T) {
		asserter.AssertInputValidation()
	})

	t.Run("resource_exhaustion", func(t *testing.T) {
		asserter.AssertErrorHandling()
	})
}
