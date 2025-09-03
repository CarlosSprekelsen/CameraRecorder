/*
Test Configuration Constants - Build Tag Definitions

This file defines all build tag constants used across the test suite.
All test files must reference these constants for consistency and maintainability.

Requirements Coverage:
- REQ-TEST-001: Test environment setup
- REQ-TEST-002: Test configuration management
- REQ-TEST-003: Build tag standardization

Test Categories: Test Infrastructure
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package testutils

import (
	"path/filepath"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/mediamtx"
)

// ============================================================================
// PRIMARY CLASSIFICATION (Test Level)
// ============================================================================

// TestUnit represents unit test build tag
const TestUnit = "unit"

// TestIntegration represents integration test build tag
const TestIntegration = "integration"

// TestSecurity represents security test build tag
const TestSecurity = "security"

// TestPerformance represents performance test build tag
const TestPerformance = "performance"

// TestHealth represents health test build tag
const TestHealth = "health"

// ============================================================================
// SECONDARY CLASSIFICATION (Test Characteristics)
// ============================================================================

// TestRealMediaMTX represents real MediaMTX system test build tag
const TestRealMediaMTX = "real_mediamtx"

// TestRealWebSocket represents real WebSocket system test build tag
const TestRealWebSocket = "real_websocket"

// TestRealSystem represents real system test build tag
const TestRealSystem = "real_system"

// TestSudoRequired represents sudo-required test build tag
const TestSudoRequired = "sudo_required"

// ============================================================================
// TERTIARY CLASSIFICATION (Test Scope)
// ============================================================================

// TestEdgeCase represents edge case test build tag
const TestEdgeCase = "edge_case"

// TestSanity represents sanity test build tag
const TestSanity = "sanity"

// TestHardware represents hardware test build tag
const TestHardware = "hardware"

// TestNetwork represents network test build tag
const TestNetwork = "network"

// ============================================================================
// BUILD TAG COMBINATIONS
// ============================================================================

// BuildTagUnit returns unit test build tag
func BuildTagUnit() string {
	return TestUnit
}

// BuildTagIntegration returns integration test build tag
func BuildTagIntegration() string {
	return TestIntegration
}

// BuildTagPerformance returns performance test build tag
func BuildTagPerformance() string {
	return TestPerformance
}

// BuildTagSecurity returns security test build tag
func BuildTagSecurity() string {
	return TestSecurity
}

// BuildTagHealth returns health test build tag
func BuildTagHealth() string {
	return TestHealth
}

// BuildTagRealMediaMTX returns real MediaMTX test build tag
func BuildTagRealMediaMTX() string {
	return TestRealMediaMTX
}

// BuildTagRealWebSocket returns real WebSocket test build tag
func BuildTagRealWebSocket() string {
	return TestRealWebSocket
}

// BuildTagRealSystem returns real system test build tag
func BuildTagRealSystem() string {
	return TestRealSystem
}

// BuildTagSudoRequired returns sudo-required test build tag
func BuildTagSudoRequired() string {
	return TestSudoRequired
}

// ============================================================================
// COMMON BUILD TAG COMBINATIONS
// ============================================================================

// BuildTagUnitRealSystem returns unit test with real system build tag combination
func BuildTagUnitRealSystem() string {
	return TestUnit + "," + TestRealSystem
}

// BuildTagIntegrationRealMediaMTX returns integration test with real MediaMTX build tag combination
func BuildTagIntegrationRealMediaMTX() string {
	return TestIntegration + "," + TestRealMediaMTX
}

// BuildTagIntegrationRealWebSocket returns integration test with real WebSocket build tag combination
func BuildTagIntegrationRealWebSocket() string {
	return TestIntegration + "," + TestRealWebSocket
}

// BuildTagIntegrationRealSystem returns integration test with real system build tag combination
func BuildTagIntegrationRealSystem() string {
	return TestIntegration + "," + TestRealSystem
}

// BuildTagUnitRealMediaMTX returns unit test with real MediaMTX build tag combination
func BuildTagUnitRealMediaMTX() string {
	return TestUnit + "," + TestRealMediaMTX
}

// BuildTagUnitRealSystemRealMediaMTX returns unit test with real system and MediaMTX build tag combination
func BuildTagUnitRealSystemRealMediaMTX() string {
	return TestUnit + "," + TestRealSystem + "," + TestRealMediaMTX
}

// ============================================================================
// VALIDATION FUNCTIONS
// ============================================================================

// IsValidBuildTag checks if a build tag is valid
func IsValidBuildTag(tag string) bool {
	validTags := []string{
		TestUnit, TestIntegration, TestSecurity, TestPerformance, TestHealth,
		TestRealMediaMTX, TestRealWebSocket, TestRealSystem, TestSudoRequired,
		TestEdgeCase, TestSanity, TestHardware, TestNetwork,
	}

	for _, validTag := range validTags {
		if tag == validTag {
			return true
		}
	}
	return false
}

// GetValidBuildTags returns all valid build tags
func GetValidBuildTags() []string {
	return []string{
		TestUnit, TestIntegration, TestSecurity, TestPerformance, TestHealth,
		TestRealMediaMTX, TestRealWebSocket, TestRealSystem, TestSudoRequired,
		TestEdgeCase, TestSanity, TestHardware, TestNetwork,
	}
}

// ============================================================================
// TEST CONFIGURATION CONSTANTS
// ============================================================================

// TestServicePorts defines standard test service ports
const (
	TestMediaMTXAPIPort    = 9997
	TestMediaMTXRTSPPort   = 8554
	TestMediaMTXWebRTCPort = 8889
	TestMediaMTXHLSPort    = 8888
	TestWebSocketPort      = 8002
)

// TestServiceHosts defines standard test service hosts
const (
	TestMediaMTXHost  = "localhost"
	TestWebSocketHost = "localhost"
)

// TestServiceURLs defines standard test service URLs
const (
	TestMediaMTXAPIURL  = "http://localhost:9997"
	TestMediaMTXRTSPURL = "rtsp://localhost:8554"
	TestWebSocketURL    = "ws://localhost:8002/ws"
)

// TestServiceEndpoints defines standard test service endpoints
const (
	TestMediaMTXHealthEndpoint = "/v3/paths/list"
	TestWebSocketPath          = "/ws"
)

// ============================================================================
// TEST CONFIGURATION UTILITY FUNCTIONS
// ============================================================================

// GetTestMediaMTXConfig returns a standard test MediaMTX configuration
func GetTestMediaMTXConfig(tempDir string) *config.MediaMTXConfig {
	return &config.MediaMTXConfig{
		Host:               TestMediaMTXHost,
		APIPort:            TestMediaMTXAPIPort,
		HealthCheckTimeout: 5 * time.Second,
		RecordingsPath:     filepath.Join(tempDir, "recordings"),
		SnapshotsPath:      filepath.Join(tempDir, "snapshots"),
	}
}

// GetTestMediaMTXClientConfig returns a standard test MediaMTX client configuration
func GetTestMediaMTXClientConfig(timeout time.Duration) *mediamtx.MediaMTXConfig {
	return &mediamtx.MediaMTXConfig{
		BaseURL: TestMediaMTXAPIURL,
		Timeout: timeout,
		ConnectionPool: mediamtx.ConnectionPoolConfig{
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 2,
			IdleConnTimeout:     30 * time.Second,
		},
	}
}

// GetTestWebSocketURL returns the standard test WebSocket URL
func GetTestWebSocketURL() string {
	return TestWebSocketURL
}

// GetTestMediaMTXHealthURL returns the standard test MediaMTX health check URL
func GetTestMediaMTXHealthURL() string {
	return TestMediaMTXAPIURL + TestMediaMTXHealthEndpoint
}

// GetTestRTSPURL returns the standard test RTSP URL with path
func GetTestRTSPURL(path string) string {
	if path == "" {
		path = "test"
	}
	return TestMediaMTXRTSPURL + "/" + path
}
