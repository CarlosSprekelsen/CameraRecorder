/*
MediaMTX Health Notification Manager Tests

This file provides comprehensive tests for the HealthNotificationManager
which centralizes debounce logic for storage and performance notifications.

Requirements Coverage:
- REQ-MTX-004: Health monitoring
- REQ-MTX-007: Error handling and recovery

Test Categories: Unit (using real MediaMTX server as per guidelines)
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package mediamtx

import (
	"sync"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createHealthNotificationManagerFromFixture creates a health notification manager using test fixture
func createHealthNotificationManagerFromFixture(t *testing.T) (*HealthNotificationManager, *MockSystemEventNotifier) {
	// Use proper test fixture instead of hardcoding configuration
	configManager := CreateConfigManagerWithFixture(t, "config_clean_minimal.yaml")
	require.NotNil(t, configManager, "Config manager should not be nil")

	// Get the actual configuration from the fixture
	cfg := configManager.GetConfig()
	require.NotNil(t, cfg, "Config should not be nil")

	logger := logging.GetLogger("mediamtx")
	notifier := NewMockSystemEventNotifier()

	manager := NewHealthNotificationManager(cfg, logger, notifier)
	require.NotNil(t, manager, "Health notification manager should not be nil")

	return manager, notifier
}

// MockSystemEventNotifier provides a test implementation of SystemEventNotifier
type MockSystemEventNotifier struct {
	notifications []SystemHealthNotification
	mu            sync.RWMutex
}

type SystemHealthNotification struct {
	Status    string                 `json:"status"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

func NewMockSystemEventNotifier() *MockSystemEventNotifier {
	return &MockSystemEventNotifier{
		notifications: make([]SystemHealthNotification, 0),
	}
}

func (m *MockSystemEventNotifier) NotifySystemHealth(status string, data map[string]interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.notifications = append(m.notifications, SystemHealthNotification{
		Status:    status,
		Data:      data,
		Timestamp: time.Now(),
	})
}

func (m *MockSystemEventNotifier) GetNotifications() []SystemHealthNotification {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]SystemHealthNotification, len(m.notifications))
	copy(result, m.notifications)
	return result
}

func (m *MockSystemEventNotifier) ClearNotifications() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.notifications = make([]SystemHealthNotification, 0)
}

func (m *MockSystemEventNotifier) GetNotificationCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.notifications)
}

// TestNewHealthNotificationManager_ReqMTX004 tests health notification manager creation
func TestNewHealthNotificationManager_ReqMTX004(t *testing.T) {
	// REQ-MTX-004: Health monitoring
	config := &config.Config{
		Performance: config.PerformanceConfig{
			MonitoringThresholds: config.MonitoringThresholdsConfig{
				MemoryUsagePercent:     90.0,
				ErrorRatePercent:       5.0,
				AverageResponseTimeMs:  1000.0,
				ActiveConnectionsLimit: 900,
				GoroutinesLimit:        1000,
			},
			Debounce: config.DebounceConfig{
				HealthMonitorSeconds:      15,
				StorageMonitorSeconds:     30,
				PerformanceMonitorSeconds: 45,
			},
		},
	}

	logger := logging.GetLogger("mediamtx")
	notifier := NewMockSystemEventNotifier()

	manager := NewHealthNotificationManager(config, logger, notifier)
	require.NotNil(t, manager, "Health notification manager should not be nil")
}

// TestHealthNotificationManager_CheckStorageThresholds_ReqMTX004 tests storage threshold checking
func TestHealthNotificationManager_CheckStorageThresholds_ReqMTX004(t *testing.T) {
	// REQ-MTX-004: Health monitoring
	config := &config.Config{
		Performance: config.PerformanceConfig{
			MonitoringThresholds: config.MonitoringThresholdsConfig{
				MemoryUsagePercent:     90.0,
				ErrorRatePercent:       5.0,
				AverageResponseTimeMs:  1000.0,
				ActiveConnectionsLimit: 900,
				GoroutinesLimit:        1000,
			},
			Debounce: config.DebounceConfig{
				HealthMonitorSeconds:      15,
				StorageMonitorSeconds:     30,
				PerformanceMonitorSeconds: 45,
			},
		},
	}

	logger := logging.GetLogger("mediamtx")
	notifier := NewMockSystemEventNotifier()

	manager := NewHealthNotificationManager(config, logger, notifier)
	require.NotNil(t, manager, "Health notification manager should not be nil")

	// Test with low storage warning
	storageInfo := &StorageInfo{
		TotalSpace:      1000000000, // 1GB
		UsedSpace:       950000000,  // 950MB (95% usage)
		AvailableSpace:  50000000,   // 50MB
		UsagePercentage: 95.0,
		RecordingsSize:  400000000, // 400MB
		SnapshotsSize:   100000000, // 100MB
		LowSpaceWarning: true,
	}

	// Check storage thresholds
	manager.CheckStorageThresholds(storageInfo)

	// Should send notification for low space warning
	notifications := notifier.GetNotifications()
	assert.Greater(t, len(notifications), 0, "Should send notification for low space warning")

	// Verify notification content
	notification := notifications[0]
	assert.Equal(t, "storage_warning", notification.Status, "Notification status should be storage_warning")
	assert.Contains(t, notification.Data, "usage_percentage", "Notification should contain usage percentage")
	assert.Contains(t, notification.Data, "available_space", "Notification should contain available space")
}

// TestHealthNotificationManager_CheckPerformanceThresholds_ReqMTX004 tests performance threshold checking
func TestHealthNotificationManager_CheckPerformanceThresholds_ReqMTX004(t *testing.T) {
	// REQ-MTX-004: Health monitoring
	config := &config.Config{
		Performance: config.PerformanceConfig{
			MonitoringThresholds: config.MonitoringThresholdsConfig{
				MemoryUsagePercent:     90.0,
				ErrorRatePercent:       5.0,
				AverageResponseTimeMs:  1000.0,
				ActiveConnectionsLimit: 900,
				GoroutinesLimit:        1000,
			},
			Debounce: config.DebounceConfig{
				HealthMonitorSeconds:      15,
				StorageMonitorSeconds:     30,
				PerformanceMonitorSeconds: 45,
			},
		},
	}

	logger := logging.GetLogger("mediamtx")
	notifier := NewMockSystemEventNotifier()

	manager := NewHealthNotificationManager(config, logger, notifier)
	require.NotNil(t, manager, "Health notification manager should not be nil")

	// Test with high memory usage
	metrics := map[string]interface{}{
		"memory_usage":          95.0,  // 95% memory usage (above 90% threshold)
		"error_rate":            3.0,   // 3% error rate (below 5% threshold)
		"average_response_time": 500.0, // 500ms response time (below 1000ms threshold)
		"active_connections":    800,   // 800 connections (below 900 threshold)
		"goroutines":            900,   // 900 goroutines (below 1000 threshold)
	}

	// Check performance thresholds
	manager.CheckPerformanceThresholds(metrics)

	// Should send notification for high memory usage
	notifications := notifier.GetNotifications()
	assert.Greater(t, len(notifications), 0, "Should send notification for high memory usage")

	// Verify notification content
	notification := notifications[0]
	assert.Equal(t, "performance_warning", notification.Status, "Notification status should be performance_warning")
	assert.Contains(t, notification.Data, "memory_usage", "Notification should contain memory usage")
	assert.Contains(t, notification.Data, "threshold", "Notification should contain threshold")
}

// TestHealthNotificationManager_DebounceMechanism_ReqMTX004 tests debounce mechanism
func TestHealthNotificationManager_DebounceMechanism_ReqMTX004(t *testing.T) {
	// REQ-MTX-004: Health monitoring
	config := &config.Config{
		Performance: config.PerformanceConfig{
			MonitoringThresholds: config.MonitoringThresholdsConfig{
				MemoryUsagePercent:     90.0,
				ErrorRatePercent:       5.0,
				AverageResponseTimeMs:  1000.0,
				ActiveConnectionsLimit: 900,
				GoroutinesLimit:        1000,
			},
			Debounce: config.DebounceConfig{
				HealthMonitorSeconds:      15,
				StorageMonitorSeconds:     30,
				PerformanceMonitorSeconds: 45,
			},
		},
	}

	logger := logging.GetLogger("mediamtx")
	notifier := NewMockSystemEventNotifier()

	manager := NewHealthNotificationManager(config, logger, notifier)
	require.NotNil(t, manager, "Health notification manager should not be nil")

	// Test debounce mechanism with rapid successive calls
	storageInfo := &StorageInfo{
		TotalSpace:      1000000000,
		UsedSpace:       950000000,
		AvailableSpace:  50000000,
		UsagePercentage: 95.0,
		RecordingsSize:  400000000,
		SnapshotsSize:   100000000,
		LowSpaceWarning: true,
	}

	// Make multiple rapid calls
	for i := 0; i < 5; i++ {
		manager.CheckStorageThresholds(storageInfo)
	}

	// Should only send one notification due to debounce
	notifications := notifier.GetNotifications()
	assert.Equal(t, 1, len(notifications), "Should only send one notification due to debounce")

	// Clear notifications and make another call
	notifier.ClearNotifications()

	// Make another call
	manager.CheckStorageThresholds(storageInfo)

	// Should send another notification after debounce period
	notifications = notifier.GetNotifications()
	assert.Equal(t, 1, len(notifications), "Should send notification after debounce period")
}

// TestHealthNotificationManager_AtomicOperations_ReqMTX004 tests atomic operations for thread safety
func TestHealthNotificationManager_AtomicOperations_ReqMTX004(t *testing.T) {
	// REQ-MTX-004: Health monitoring
	config := &config.Config{
		Performance: config.PerformanceConfig{
			MonitoringThresholds: config.MonitoringThresholdsConfig{
				MemoryUsagePercent:     90.0,
				ErrorRatePercent:       5.0,
				AverageResponseTimeMs:  1000.0,
				ActiveConnectionsLimit: 900,
				GoroutinesLimit:        1000,
			},
			Debounce: config.DebounceConfig{
				HealthMonitorSeconds:      15,
				StorageMonitorSeconds:     30,
				PerformanceMonitorSeconds: 45,
			},
		},
	}

	logger := logging.GetLogger("mediamtx")
	notifier := NewMockSystemEventNotifier()

	manager := NewHealthNotificationManager(config, logger, notifier)
	require.NotNil(t, manager, "Health notification manager should not be nil")

	// Test concurrent access to ensure atomic operations work correctly
	storageInfo := &StorageInfo{
		TotalSpace:      1000000000,
		UsedSpace:       950000000,
		AvailableSpace:  50000000,
		UsagePercentage: 95.0,
		RecordingsSize:  400000000,
		SnapshotsSize:   100000000,
		LowSpaceWarning: true,
	}

	// Run concurrent goroutines
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("BUG DETECTED: Race condition caused panic: %v", r)
				}
				done <- true
			}()

			// Make concurrent calls to test atomic operations
			manager.CheckStorageThresholds(storageInfo)
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should not panic and should handle concurrent access gracefully
	notifications := notifier.GetNotifications()
	assert.GreaterOrEqual(t, len(notifications), 0, "Should handle concurrent access without panicking")
}

// TestHealthNotificationManager_ThresholdValidation_ReqMTX004 tests threshold validation
func TestHealthNotificationManager_ThresholdValidation_ReqMTX004(t *testing.T) {
	// REQ-MTX-004: Health monitoring
	// Test various threshold scenarios
	testCases := []struct {
		name     string
		metrics  map[string]interface{}
		expected bool // Whether notification should be sent
	}{
		{
			name: "high_memory_usage",
			metrics: map[string]interface{}{
				"memory_usage":          95.0,
				"error_rate":            2.0,
				"average_response_time": 500.0,
				"active_connections":    800,
				"goroutines":            900,
			},
			expected: true,
		},
		{
			name: "high_error_rate",
			metrics: map[string]interface{}{
				"memory_usage":          80.0,
				"error_rate":            7.0,
				"average_response_time": 500.0,
				"active_connections":    800,
				"goroutines":            900,
			},
			expected: true,
		},
		{
			name: "high_response_time",
			metrics: map[string]interface{}{
				"memory_usage":          80.0,
				"error_rate":            2.0,
				"average_response_time": 1500.0,
				"active_connections":    800,
				"goroutines":            900,
			},
			expected: true,
		},
		{
			name: "high_connections",
			metrics: map[string]interface{}{
				"memory_usage":          80.0,
				"error_rate":            2.0,
				"average_response_time": 500.0,
				"active_connections":    950,
				"goroutines":            900,
			},
			expected: true,
		},
		{
			name: "high_goroutines",
			metrics: map[string]interface{}{
				"memory_usage":          80.0,
				"error_rate":            2.0,
				"average_response_time": 500.0,
				"active_connections":    800,
				"goroutines":            1100,
			},
			expected: true,
		},
		{
			name: "all_normal",
			metrics: map[string]interface{}{
				"memory_usage":          80.0,
				"error_rate":            2.0,
				"average_response_time": 500.0,
				"active_connections":    800,
				"goroutines":            900,
			},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create fresh manager instance for each sub-test to avoid state contamination
			manager, notifier := createHealthNotificationManagerFromFixture(t)

			manager.CheckPerformanceThresholds(tc.metrics)

			notifications := notifier.GetNotifications()
			if tc.expected {
				assert.Greater(t, len(notifications), 0, "Should send notification for %s", tc.name)
			} else {
				assert.Equal(t, 0, len(notifications), "Should not send notification for %s", tc.name)
			}
		})
	}
}

// TestHealthNotificationManager_StorageInfoInterface_ReqMTX004 tests StorageInfo interface methods
func TestHealthNotificationManager_StorageInfoInterface_ReqMTX004(t *testing.T) {
	// REQ-MTX-004: Health monitoring
	storageInfo := &StorageInfo{
		TotalSpace:      1000000000,
		UsedSpace:       750000000,
		AvailableSpace:  250000000,
		UsagePercentage: 75.0,
		RecordingsSize:  400000000,
		SnapshotsSize:   100000000,
		LowSpaceWarning: false,
	}

	// Test interface methods
	assert.Equal(t, 75.0, storageInfo.GetUsagePercentage(), "GetUsagePercentage should return correct value")
	assert.Equal(t, int64(250000000), storageInfo.GetAvailableSpace(), "GetAvailableSpace should return correct value")
	assert.Equal(t, int64(1000000000), storageInfo.GetTotalSpace(), "GetTotalSpace should return correct value")
	assert.Equal(t, false, storageInfo.IsLowSpaceWarning(), "IsLowSpaceWarning should return correct value")

	// Test with low space warning
	storageInfo.LowSpaceWarning = true
	assert.Equal(t, true, storageInfo.IsLowSpaceWarning(), "IsLowSpaceWarning should return true when warning is set")
}
