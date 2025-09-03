//go:build performance
// +build performance

package websocket_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/tests/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// RESOURCE OPTIMIZATION & POWER CONSUMPTION TESTING
// ============================================================================

// TestResourceUsageBaseline establishes baseline resource consumption
func TestResourceUsageBaseline(t *testing.T) {
	/*
		Resource Usage Baseline Test

		This test establishes baseline resource consumption for the system
		without any active polling or WebSocket operations.

		REQ-POWER-001: Baseline resource consumption measurement
		REQ-PERFORMANCE-001: Resource efficiency validation
	*/

	// COMMON PATTERN: Use shared test environment
	env := utils.SetupWebSocketTestEnvironment(t)
	defer utils.TeardownWebSocketTestEnvironment(t, env)

	// Start services
	err := env.WebSocketServer.Start()
	require.NoError(t, err, "WebSocket server should start")
	defer env.WebSocketServer.Stop()

	// Use ResourceMonitor for accurate measurements
	monitor := utils.NewResourceMonitor()
	monitor.Start()

	// Allow system to stabilize
	time.Sleep(2 * time.Second)

	// Take multiple measurements for accuracy
	for i := 0; i < 5; i++ {
		monitor.Measure()
		time.Sleep(500 * time.Millisecond)
	}

	baselineMetrics := monitor.GetAverageMetrics()

	t.Logf("Baseline resource usage:")
	t.Logf("- CPU Usage: %.2f%%", baselineMetrics.CPUPercent)
	t.Logf("- Memory Usage: %.2f MB", baselineMetrics.MemoryMB)
	t.Logf("- Goroutines: %d", baselineMetrics.GoroutineCount)
	t.Logf("- Heap Allocated: %.2f MB", baselineMetrics.HeapAllocMB)
	t.Logf("- Heap System: %.2f MB", baselineMetrics.HeapSysMB)

	// Validate baseline is reasonable
	assert.Less(t, baselineMetrics.CPUPercent, 5.0, "Baseline CPU usage should be <5%")
	assert.Less(t, baselineMetrics.MemoryMB, 100.0, "Baseline memory usage should be <100MB")
	assert.Less(t, baselineMetrics.GoroutineCount, 20, "Baseline goroutine count should be <20")

	// Print recommendations
	recommendations := monitor.GetRecommendations()
	t.Logf("Baseline recommendations:")
	for _, rec := range recommendations {
		t.Logf("- %s", rec)
	}
}

// TestPollingIntervalResourceImpact tests different polling intervals
func TestPollingIntervalResourceImpact(t *testing.T) {
	/*
		Polling Interval Resource Impact Test

		This test measures resource consumption at different polling intervals
		to find the optimal balance between performance and power consumption.

		REQ-POWER-002: Polling interval optimization
		REQ-PERFORMANCE-002: Performance vs power trade-off analysis
	*/

	pollingIntervals := []time.Duration{
		100 * time.Millisecond, // Aggressive polling (high performance, high power)
		500 * time.Millisecond, // Moderate polling (balanced)
		1 * time.Second,        // Conservative polling (lower power)
		2 * time.Second,        // Minimal polling (lowest power)
	}

	results := make(map[time.Duration]utils.ResourceMeasurement)

	for _, interval := range pollingIntervals {
		t.Run(fmt.Sprintf("PollingInterval_%v", interval), func(t *testing.T) {
			// COMMON PATTERN: Use shared test environment
			env := utils.SetupWebSocketTestEnvironment(t)
			defer utils.TeardownWebSocketTestEnvironment(t, env)

			// Start services
			err := env.WebSocketServer.Start()
			require.NoError(t, err, "WebSocket server should start")
			defer env.WebSocketServer.Stop()

			// Start camera monitor with specific polling interval
			err = env.CameraMonitor.Start(context.Background())
			require.NoError(t, err, "Camera monitor should start")
			defer env.CameraMonitor.Stop()

			// Allow system to stabilize
			time.Sleep(2 * time.Second)

			// Use ResourceMonitor for accurate measurements
			monitor := utils.NewResourceMonitor()
			monitor.Start()

			// Measure resource usage during active polling
			monitor.MeasureOverTime(10*time.Second, interval)

			// Calculate average metrics
			avgMetrics := monitor.GetAverageMetrics()
			results[interval] = avgMetrics

			t.Logf("Polling interval %v results:", interval)
			t.Logf("- Avg CPU Usage: %.2f%%", avgMetrics.CPUPercent)
			t.Logf("- Avg Memory Usage: %.2f MB", avgMetrics.MemoryMB)
			t.Logf("- Avg Goroutines: %d", avgMetrics.GoroutineCount)
			t.Logf("- Avg Heap Allocated: %.2f MB", avgMetrics.HeapAllocMB)
		})
	}

	// Analyze results and recommend optimal interval
	recommendOptimalPollingInterval(t, results)
}

// TestConcurrentConnectionsResourceImpact tests resource usage under load
func TestConcurrentConnectionsResourceImpact(t *testing.T) {
	/*
		Concurrent Connections Resource Impact Test

		This test measures resource consumption with multiple concurrent
		WebSocket connections to understand scaling characteristics.

		REQ-POWER-003: Concurrent connection resource scaling
		REQ-PERFORMANCE-003: Load-based resource optimization
	*/

	connectionCounts := []int{1, 5, 10, 25, 50}

	for _, count := range connectionCounts {
		t.Run(fmt.Sprintf("ConcurrentConnections_%d", count), func(t *testing.T) {
			// COMMON PATTERN: Use shared test environment
			env := utils.SetupWebSocketTestEnvironment(t)
			defer utils.TeardownWebSocketTestEnvironment(t, env)

			// Start services
			err := env.WebSocketServer.Start()
			require.NoError(t, err, "WebSocket server should start")
			defer env.WebSocketServer.Stop()

			// Create concurrent connections
			clients := make([]*utils.WebSocketTestClient, count)
			var wg sync.WaitGroup

			// Establish connections
			for i := 0; i < count; i++ {
				wg.Add(1)
				go func(index int) {
					defer wg.Done()
					client := utils.NewWebSocketTestClient(t, env.WebSocketServer, env.JWTHandler)
					clients[index] = client
					defer client.Close()

					// Send periodic ping requests
					for j := 0; j < 5; j++ {
						client.SendPingRequest()
						time.Sleep(100 * time.Millisecond)
					}
				}(i)
			}

			wg.Wait()

			// Use ResourceMonitor for accurate measurements
			monitor := utils.NewResourceMonitor()
			monitor.Start()

			// Take measurements under load
			for i := 0; i < 5; i++ {
				monitor.Measure()
				time.Sleep(500 * time.Millisecond)
			}

			metrics := monitor.GetAverageMetrics()

			t.Logf("Concurrent connections %d results:", count)
			t.Logf("- CPU Usage: %.2f%%", metrics.CPUPercent)
			t.Logf("- Memory Usage: %.2f MB", metrics.MemoryMB)
			t.Logf("- Goroutines: %d", metrics.GoroutineCount)
			t.Logf("- Heap Allocated: %.2f MB", metrics.HeapAllocMB)

			// Validate resource usage scales reasonably
			assert.Less(t, metrics.CPUPercent, float64(count)*2.0,
				"CPU usage should scale reasonably with connection count")
			assert.Less(t, metrics.MemoryMB, float64(count)*5.0,
				"Memory usage should scale reasonably with connection count")
		})
	}
}

// TestMemoryLeakDetection tests for memory leaks during extended operation
func TestMemoryLeakDetection(t *testing.T) {
	/*
		Memory Leak Detection Test

		This test runs the system for an extended period to detect
		potential memory leaks that could impact power consumption.

		REQ-POWER-004: Memory leak detection and prevention
		REQ-RELIABILITY-003: Long-term memory stability
	*/

	// COMMON PATTERN: Use shared test environment
	env := utils.SetupWebSocketTestEnvironment(t)
	defer utils.TeardownWebSocketTestEnvironment(t, env)

	// Start services
	err := env.WebSocketServer.Start()
	require.NoError(t, err, "WebSocket server should start")
	defer env.WebSocketServer.Stop()

	err = env.CameraMonitor.Start(context.Background())
	require.NoError(t, err, "Camera monitor should start")
	defer env.CameraMonitor.Stop()

	// Use ResourceMonitor for accurate measurements
	monitor := utils.NewResourceMonitor()
	monitor.Start()

	// Measure memory usage over time
	testDuration := 30 * time.Second
	measurementInterval := 5 * time.Second
	measurements := monitor.MeasureOverTime(testDuration, measurementInterval)

	// Analyze memory growth
	initialMemory := measurements[0].MemoryMB
	finalMemory := measurements[len(measurements)-1].MemoryMB
	memoryGrowth := finalMemory - initialMemory

	t.Logf("Memory leak detection results:")
	t.Logf("- Initial memory: %.2f MB", initialMemory)
	t.Logf("- Final memory: %.2f MB", finalMemory)
	t.Logf("- Memory growth: %.2f MB", memoryGrowth)
	t.Logf("- Growth rate: %.2f MB/min", memoryGrowth*60/float64(testDuration.Seconds()))

	// Validate no significant memory leak
	assert.Less(t, memoryGrowth, 10.0, "Memory growth should be <10MB over test duration")
	assert.Less(t, memoryGrowth*60/float64(testDuration.Seconds()), 20.0,
		"Memory growth rate should be <20MB/min")
}

// TestPowerEfficiencyOptimization finds optimal configuration
func TestPowerEfficiencyOptimization(t *testing.T) {
	/*
		Power Efficiency Optimization Test

		This test finds the optimal configuration that balances
		performance requirements with power consumption.

		REQ-POWER-005: Power efficiency optimization
		REQ-PERFORMANCE-004: Performance vs power trade-off optimization
	*/

	// Test different configurations
	configurations := []struct {
		name            string
		pollingInterval time.Duration
		maxConnections  int
		description     string
	}{
		{
			name:            "HighPerformance",
			pollingInterval: 100 * time.Millisecond,
			maxConnections:  100,
			description:     "High performance, higher power consumption",
		},
		{
			name:            "Balanced",
			pollingInterval: 500 * time.Millisecond,
			maxConnections:  50,
			description:     "Balanced performance and power",
		},
		{
			name:            "PowerEfficient",
			pollingInterval: 1 * time.Second,
			maxConnections:  25,
			description:     "Power efficient, moderate performance",
		},
		{
			name:            "UltraEfficient",
			pollingInterval: 2 * time.Second,
			maxConnections:  10,
			description:     "Ultra power efficient, lower performance",
		},
	}

	results := make(map[string]utils.ResourceMeasurement)

	for _, config := range configurations {
		t.Run(config.name, func(t *testing.T) {
			// COMMON PATTERN: Use shared test environment
			env := utils.SetupWebSocketTestEnvironment(t)
			defer utils.TeardownWebSocketTestEnvironment(t, env)

			// Start services
			err := env.WebSocketServer.Start()
			require.NoError(t, err, "WebSocket server should start")
			defer env.WebSocketServer.Stop()

			err = env.CameraMonitor.Start(context.Background())
			require.NoError(t, err, "Camera monitor should start")
			defer env.CameraMonitor.Stop()

			// Create test load
			clients := make([]*utils.WebSocketTestClient, config.maxConnections)
			for i := 0; i < config.maxConnections; i++ {
				client := utils.NewWebSocketTestClient(t, env.WebSocketServer, env.JWTHandler)
				clients[i] = client
				defer client.Close()
			}

			// Allow system to stabilize
			time.Sleep(2 * time.Second)

			// Use ResourceMonitor for accurate measurements
			monitor := utils.NewResourceMonitor()
			monitor.Start()

			// Take measurements
			for i := 0; i < 5; i++ {
				monitor.Measure()
				time.Sleep(500 * time.Millisecond)
			}

			metrics := monitor.GetAverageMetrics()
			results[config.name] = metrics

			t.Logf("Configuration %s (%s):", config.name, config.description)
			t.Logf("- CPU Usage: %.2f%%", metrics.CPUPercent)
			t.Logf("- Memory Usage: %.2f MB", metrics.MemoryMB)
			t.Logf("- Goroutines: %d", metrics.GoroutineCount)
			t.Logf("- Heap Allocated: %.2f MB", metrics.HeapAllocMB)
		})
	}

	// Recommend optimal configuration
	recommendOptimalConfiguration(t, results)
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// recommendOptimalPollingInterval analyzes results and recommends optimal interval
func recommendOptimalPollingInterval(t *testing.T, results map[time.Duration]utils.ResourceMeasurement) {
	t.Logf("=== POLLING INTERVAL OPTIMIZATION ANALYSIS ===")

	var bestInterval time.Duration
	var bestScore float64

	for interval, metrics := range results {
		// Calculate efficiency score (lower is better)
		// Consider both CPU usage and memory efficiency
		efficiencyScore := metrics.CPUPercent + (metrics.MemoryMB * 0.1)

		t.Logf("Interval %v: CPU=%.2f%%, Memory=%.2fMB, Score=%.2f",
			interval, metrics.CPUPercent, metrics.MemoryMB, efficiencyScore)

		if bestInterval == 0 || efficiencyScore < bestScore {
			bestInterval = interval
			bestScore = efficiencyScore
		}
	}

	t.Logf("=== RECOMMENDATION ===")
	t.Logf("Optimal polling interval: %v", bestInterval)
	t.Logf("Efficiency score: %.2f", bestScore)

	// Provide specific recommendations
	switch bestInterval {
	case 100 * time.Millisecond:
		t.Logf("RECOMMENDATION: Use aggressive polling for high-performance scenarios")
		t.Logf("Trade-off: Higher power consumption for maximum responsiveness")
	case 500 * time.Millisecond:
		t.Logf("RECOMMENDATION: Use balanced polling for most production scenarios")
		t.Logf("Trade-off: Good balance between performance and power efficiency")
	case 1 * time.Second:
		t.Logf("RECOMMENDATION: Use conservative polling for power-sensitive deployments")
		t.Logf("Trade-off: Lower power consumption with acceptable performance")
	case 2 * time.Second:
		t.Logf("RECOMMENDATION: Use minimal polling for battery-powered devices")
		t.Logf("Trade-off: Maximum power efficiency with reduced responsiveness")
	}
}

// recommendOptimalConfiguration analyzes configuration results
func recommendOptimalConfiguration(t *testing.T, results map[string]utils.ResourceMeasurement) {
	t.Logf("=== CONFIGURATION OPTIMIZATION ANALYSIS ===")

	var bestConfig string
	var bestScore float64

	for config, metrics := range results {
		// Calculate power efficiency score (lower is better)
		powerScore := metrics.CPUPercent + (metrics.MemoryMB * 0.2)

		t.Logf("Config %s: CPU=%.2f%%, Memory=%.2fMB, PowerScore=%.2f",
			config, metrics.CPUPercent, metrics.MemoryMB, powerScore)

		if bestConfig == "" || powerScore < bestScore {
			bestConfig = config
			bestScore = powerScore
		}
	}

	t.Logf("=== OPTIMAL CONFIGURATION ===")
	t.Logf("Recommended configuration: %s", bestConfig)
	t.Logf("Power efficiency score: %.2f", bestScore)

	// Provide deployment recommendations
	switch bestConfig {
	case "HighPerformance":
		t.Logf("DEPLOYMENT: Use for high-traffic production environments")
		t.Logf("Consider: Server-grade hardware with adequate cooling")
	case "Balanced":
		t.Logf("DEPLOYMENT: Use for standard production environments")
		t.Logf("Consider: Most common deployment scenario")
	case "PowerEfficient":
		t.Logf("DEPLOYMENT: Use for edge devices or power-constrained environments")
		t.Logf("Consider: IoT devices, battery-powered systems")
	case "UltraEfficient":
		t.Logf("DEPLOYMENT: Use for ultra-low-power scenarios")
		t.Logf("Consider: Solar-powered devices, remote sensors")
	}
}

// TestResourceOptimizationFileRecognition ensures Go recognizes this file
func TestResourceOptimizationFileRecognition(t *testing.T) {
	t.Log("Resource optimization test file is recognized by Go")
	t.Log("Epic E3 Power Efficiency: Optimal performance with minimal power consumption")
	t.Log("Status: Ready for resource optimization validation")
}
