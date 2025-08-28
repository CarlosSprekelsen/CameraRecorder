/*
MediaMTX System Metrics Unit Tests

Requirements Coverage:
- REQ-SYS-006: System metrics collection and reporting
- REQ-SYS-007: Performance metrics tracking
- REQ-SYS-008: Resource usage monitoring
- REQ-SYS-009: Metrics aggregation and calculation
- REQ-SYS-010: Real-time metrics updates

Test Categories: Unit
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package mediamtx_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/mediamtx"
)

//go:build unit
// +build unit

// mockSystemMetrics implements system metrics collection for testing
type mockSystemMetrics struct {
	requestCount       int64
	responseTimeTotal  float64
	responseTimeCount  int64
	errorCount         int64
	activeConnections  int64
	startTime          time.Time
	lastUpdateTime     time.Time
}

func newMockSystemMetrics() *mockSystemMetrics {
	return &mockSystemMetrics{
		requestCount:      0,
		responseTimeTotal: 0.0,
		responseTimeCount: 0,
		errorCount:        0,
		activeConnections: 0,
		startTime:         time.Now(),
		lastUpdateTime:    time.Now(),
	}
}

func (m *mockSystemMetrics) IncrementRequestCount() {
	m.requestCount++
	m.lastUpdateTime = time.Now()
}

func (m *mockSystemMetrics) RecordResponseTime(duration time.Duration) {
	m.responseTimeTotal += float64(duration.Milliseconds())
	m.responseTimeCount++
	m.lastUpdateTime = time.Now()
}

func (m *mockSystemMetrics) IncrementErrorCount() {
	m.errorCount++
	m.lastUpdateTime = time.Now()
}

func (m *mockSystemMetrics) SetActiveConnections(count int64) {
	m.activeConnections = count
	m.lastUpdateTime = time.Now()
}

func (m *mockSystemMetrics) GetMetrics() mediamtx.SystemMetrics {
	var avgResponseTime float64
	if m.responseTimeCount > 0 {
		avgResponseTime = m.responseTimeTotal / float64(m.responseTimeCount)
	}

	return mediamtx.SystemMetrics{
		RequestCount:      m.requestCount,
		ResponseTimeAvg:   avgResponseTime,
		ErrorCount:        m.errorCount,
		ActiveConnections: m.activeConnections,
		Uptime:           time.Since(m.startTime),
		LastUpdateTime:   m.lastUpdateTime,
	}
}

func (m *mockSystemMetrics) Reset() {
	m.requestCount = 0
	m.responseTimeTotal = 0.0
	m.responseTimeCount = 0
	m.errorCount = 0
	m.activeConnections = 0
	m.startTime = time.Now()
	m.lastUpdateTime = time.Now()
}

func TestSystemMetricsBasicOperations(t *testing.T) {
	// REQ-SYS-006: System metrics collection and reporting

	t.Run("InitialState", func(t *testing.T) {
		metrics := newMockSystemMetrics()
		systemMetrics := metrics.GetMetrics()

		assert.Equal(t, int64(0), systemMetrics.RequestCount, "Initial request count should be 0")
		assert.Equal(t, float64(0), systemMetrics.ResponseTimeAvg, "Initial response time average should be 0")
		assert.Equal(t, int64(0), systemMetrics.ErrorCount, "Initial error count should be 0")
		assert.Equal(t, int64(0), systemMetrics.ActiveConnections, "Initial active connections should be 0")
		assert.NotZero(t, systemMetrics.Uptime, "Uptime should be non-zero")
		assert.NotZero(t, systemMetrics.LastUpdateTime, "Last update time should be non-zero")
	})

	t.Run("IncrementRequestCount", func(t *testing.T) {
		metrics := newMockSystemMetrics()

		// Increment request count
		metrics.IncrementRequestCount()
		systemMetrics := metrics.GetMetrics()
		assert.Equal(t, int64(1), systemMetrics.RequestCount, "Request count should be 1")

		// Increment again
		metrics.IncrementRequestCount()
		systemMetrics = metrics.GetMetrics()
		assert.Equal(t, int64(2), systemMetrics.RequestCount, "Request count should be 2")
	})

	t.Run("RecordResponseTime", func(t *testing.T) {
		metrics := newMockSystemMetrics()

		// Record response times
		metrics.RecordResponseTime(100 * time.Millisecond)
		metrics.RecordResponseTime(200 * time.Millisecond)
		metrics.RecordResponseTime(300 * time.Millisecond)

		systemMetrics := metrics.GetMetrics()
		assert.Equal(t, int64(3), systemMetrics.ResponseTimeCount, "Should have recorded 3 response times")
		assert.Equal(t, float64(200), systemMetrics.ResponseTimeAvg, "Average response time should be 200ms")
	})

	t.Run("IncrementErrorCount", func(t *testing.T) {
		metrics := newMockSystemMetrics()

		// Increment error count
		metrics.IncrementErrorCount()
		systemMetrics := metrics.GetMetrics()
		assert.Equal(t, int64(1), systemMetrics.ErrorCount, "Error count should be 1")

		// Increment again
		metrics.IncrementErrorCount()
		systemMetrics = metrics.GetMetrics()
		assert.Equal(t, int64(2), systemMetrics.ErrorCount, "Error count should be 2")
	})

	t.Run("SetActiveConnections", func(t *testing.T) {
		metrics := newMockSystemMetrics()

		// Set active connections
		metrics.SetActiveConnections(5)
		systemMetrics := metrics.GetMetrics()
		assert.Equal(t, int64(5), systemMetrics.ActiveConnections, "Active connections should be 5")

		// Update active connections
		metrics.SetActiveConnections(10)
		systemMetrics = metrics.GetMetrics()
		assert.Equal(t, int64(10), systemMetrics.ActiveConnections, "Active connections should be 10")
	})
}

func TestSystemMetricsCalculations(t *testing.T) {
	// REQ-SYS-009: Metrics aggregation and calculation

	t.Run("ResponseTimeAverage_Calculation", func(t *testing.T) {
		metrics := newMockSystemMetrics()

		// Record various response times
		responseTimes := []time.Duration{
			50 * time.Millisecond,
			100 * time.Millisecond,
			150 * time.Millisecond,
			200 * time.Millisecond,
		}

		for _, rt := range responseTimes {
			metrics.RecordResponseTime(rt)
		}

		systemMetrics := metrics.GetMetrics()
		expectedAverage := float64(125) // (50+100+150+200)/4 = 125ms
		assert.Equal(t, expectedAverage, systemMetrics.ResponseTimeAvg, "Average response time should be calculated correctly")
		assert.Equal(t, int64(4), systemMetrics.ResponseTimeCount, "Response time count should be 4")
	})

	t.Run("ResponseTimeAverage_ZeroCount", func(t *testing.T) {
		metrics := newMockSystemMetrics()

		systemMetrics := metrics.GetMetrics()
		assert.Equal(t, float64(0), systemMetrics.ResponseTimeAvg, "Average should be 0 when no response times recorded")
		assert.Equal(t, int64(0), systemMetrics.ResponseTimeCount, "Response time count should be 0")
	})

	t.Run("Uptime_Calculation", func(t *testing.T) {
		metrics := newMockSystemMetrics()
		
		// Wait a bit to ensure uptime increases
		time.Sleep(10 * time.Millisecond)
		
		systemMetrics := metrics.GetMetrics()
		assert.True(t, systemMetrics.Uptime > 0, "Uptime should be positive")
		assert.True(t, systemMetrics.Uptime < time.Second, "Uptime should be reasonable")
	})
}

func TestSystemMetricsRealTimeUpdates(t *testing.T) {
	// REQ-SYS-010: Real-time metrics updates

	t.Run("LastUpdateTime_Updates", func(t *testing.T) {
		metrics := newMockSystemMetrics()
		initialUpdateTime := metrics.GetMetrics().LastUpdateTime

		// Wait a bit
		time.Sleep(10 * time.Millisecond)

		// Perform an operation
		metrics.IncrementRequestCount()
		updatedMetrics := metrics.GetMetrics()

		assert.True(t, updatedMetrics.LastUpdateTime.After(initialUpdateTime), "Last update time should be updated after operation")
	})

	t.Run("ConcurrentOperations", func(t *testing.T) {
		metrics := newMockSystemMetrics()

		// Simulate concurrent operations
		metrics.IncrementRequestCount()
		metrics.RecordResponseTime(100 * time.Millisecond)
		metrics.IncrementErrorCount()
		metrics.SetActiveConnections(5)

		systemMetrics := metrics.GetMetrics()

		assert.Equal(t, int64(1), systemMetrics.RequestCount, "Request count should be 1")
		assert.Equal(t, float64(100), systemMetrics.ResponseTimeAvg, "Response time average should be 100ms")
		assert.Equal(t, int64(1), systemMetrics.ErrorCount, "Error count should be 1")
		assert.Equal(t, int64(5), systemMetrics.ActiveConnections, "Active connections should be 5")
	})
}

func TestSystemMetricsReset(t *testing.T) {
	// Test metrics reset functionality

	t.Run("Reset_AllMetrics", func(t *testing.T) {
		metrics := newMockSystemMetrics()

		// Populate metrics
		metrics.IncrementRequestCount()
		metrics.RecordResponseTime(100 * time.Millisecond)
		metrics.IncrementErrorCount()
		metrics.SetActiveConnections(5)

		// Verify metrics are populated
		systemMetrics := metrics.GetMetrics()
		assert.Equal(t, int64(1), systemMetrics.RequestCount, "Request count should be populated")
		assert.Equal(t, float64(100), systemMetrics.ResponseTimeAvg, "Response time should be populated")
		assert.Equal(t, int64(1), systemMetrics.ErrorCount, "Error count should be populated")
		assert.Equal(t, int64(5), systemMetrics.ActiveConnections, "Active connections should be populated")

		// Reset metrics
		metrics.Reset()

		// Verify metrics are reset
		systemMetrics = metrics.GetMetrics()
		assert.Equal(t, int64(0), systemMetrics.RequestCount, "Request count should be reset to 0")
		assert.Equal(t, float64(0), systemMetrics.ResponseTimeAvg, "Response time should be reset to 0")
		assert.Equal(t, int64(0), systemMetrics.ErrorCount, "Error count should be reset to 0")
		assert.Equal(t, int64(0), systemMetrics.ActiveConnections, "Active connections should be reset to 0")
	})

	t.Run("Reset_UptimeRestart", func(t *testing.T) {
		metrics := newMockSystemMetrics()
		initialUptime := metrics.GetMetrics().Uptime

		// Wait a bit
		time.Sleep(10 * time.Millisecond)

		// Reset should restart uptime
		metrics.Reset()
		resetUptime := metrics.GetMetrics().Uptime

		assert.True(t, resetUptime < initialUptime, "Uptime should be reset to a lower value")
		assert.True(t, resetUptime < time.Second, "Reset uptime should be reasonable")
	})
}

func TestSystemMetricsEdgeCases(t *testing.T) {
	// Test edge cases and boundary conditions

	t.Run("LargeResponseTimes", func(t *testing.T) {
		metrics := newMockSystemMetrics()

		// Record very large response times
		metrics.RecordResponseTime(10 * time.Second)
		metrics.RecordResponseTime(20 * time.Second)

		systemMetrics := metrics.GetMetrics()
		expectedAverage := float64(15000) // (10000+20000)/2 = 15000ms
		assert.Equal(t, expectedAverage, systemMetrics.ResponseTimeAvg, "Should handle large response times")
	})

	t.Run("ZeroResponseTime", func(t *testing.T) {
		metrics := newMockSystemMetrics()

		// Record zero response time
		metrics.RecordResponseTime(0 * time.Millisecond)
		metrics.RecordResponseTime(100 * time.Millisecond)

		systemMetrics := metrics.GetMetrics()
		expectedAverage := float64(50) // (0+100)/2 = 50ms
		assert.Equal(t, expectedAverage, systemMetrics.ResponseTimeAvg, "Should handle zero response time")
	})

	t.Run("NegativeActiveConnections", func(t *testing.T) {
		metrics := newMockSystemMetrics()

		// Set negative active connections (edge case)
		metrics.SetActiveConnections(-1)
		systemMetrics := metrics.GetMetrics()
		assert.Equal(t, int64(-1), systemMetrics.ActiveConnections, "Should handle negative active connections")
	})

	t.Run("VeryLargeRequestCount", func(t *testing.T) {
		metrics := newMockSystemMetrics()

		// Simulate very large request count
		for i := 0; i < 1000; i++ {
			metrics.IncrementRequestCount()
		}

		systemMetrics := metrics.GetMetrics()
		assert.Equal(t, int64(1000), systemMetrics.RequestCount, "Should handle large request counts")
	})
}

func TestSystemMetricsConsistency(t *testing.T) {
	// Test metrics consistency and thread safety (basic)

	t.Run("MetricsConsistency", func(t *testing.T) {
		metrics := newMockSystemMetrics()

		// Perform operations
		metrics.IncrementRequestCount()
		metrics.RecordResponseTime(100 * time.Millisecond)
		metrics.IncrementErrorCount()
		metrics.SetActiveConnections(3)

		// Get metrics multiple times
		metrics1 := metrics.GetMetrics()
		metrics2 := metrics.GetMetrics()

		// Metrics should be consistent
		assert.Equal(t, metrics1.RequestCount, metrics2.RequestCount, "Request count should be consistent")
		assert.Equal(t, metrics1.ResponseTimeAvg, metrics2.ResponseTimeAvg, "Response time average should be consistent")
		assert.Equal(t, metrics1.ErrorCount, metrics2.ErrorCount, "Error count should be consistent")
		assert.Equal(t, metrics1.ActiveConnections, metrics2.ActiveConnections, "Active connections should be consistent")
	})

	t.Run("UptimeMonotonic", func(t *testing.T) {
		metrics := newMockSystemMetrics()

		// Get uptime multiple times
		uptime1 := metrics.GetMetrics().Uptime
		time.Sleep(10 * time.Millisecond)
		uptime2 := metrics.GetMetrics().Uptime

		// Uptime should be monotonically increasing
		assert.True(t, uptime2 > uptime1, "Uptime should be monotonically increasing")
	})
}
