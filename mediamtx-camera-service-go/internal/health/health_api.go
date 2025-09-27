/*
HealthAPI Interface Implementation

Requirements Coverage:
- REQ-HEALTH-001: Health Monitoring
- REQ-HEALTH-002: HTTP Health Endpoints

Test Categories: Unit/Integration
API Documentation Reference: docs/api/health-endpoints.md

Defines the HealthAPI interface for component integration.
Follows canonical interface patterns and thin delegation architecture.
*/

package health

import (
	"context"
	"time"
)

// HealthStatus represents the overall health status of the system
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusDegraded  HealthStatus = "degraded"
)

// ComponentStatus represents the status of individual components
type ComponentStatus struct {
	Name        string       `json:"name"`
	Status      HealthStatus `json:"status"`
	Message     string       `json:"message,omitempty"`
	LastChecked time.Time    `json:"last_checked"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

// HealthResponse represents the basic health response
type HealthResponse struct {
	Status    HealthStatus `json:"status"`
	Timestamp time.Time    `json:"timestamp"`
	Version   string       `json:"version,omitempty"`
	Uptime    string       `json:"uptime,omitempty"`
}

// DetailedHealthResponse represents the comprehensive health response
type DetailedHealthResponse struct {
	Status      HealthStatus       `json:"status"`
	Timestamp   time.Time          `json:"timestamp"`
	Version     string             `json:"version,omitempty"`
	Uptime      string             `json:"uptime,omitempty"`
	Components  []ComponentStatus  `json:"components,omitempty"`
	Metrics     map[string]interface{} `json:"metrics,omitempty"`
	Environment string             `json:"environment,omitempty"`
}

// ReadinessResponse represents the readiness probe response
type ReadinessResponse struct {
	Ready     bool      `json:"ready"`
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message,omitempty"`
}

// LivenessResponse represents the liveness probe response
type LivenessResponse struct {
	Alive     bool      `json:"alive"`
	Timestamp time.Time `json:"timestamp"`
	Message   string    `json:"message,omitempty"`
}

// HealthAPI defines the interface for health monitoring components
// This interface enables thin delegation pattern - HTTP server delegates all operations
type HealthAPI interface {
	// GetHealth returns basic health status
	GetHealth(ctx context.Context) (*HealthResponse, error)
	
	// GetDetailedHealth returns comprehensive health status
	GetDetailedHealth(ctx context.Context) (*DetailedHealthResponse, error)
	
	// IsReady checks if the system is ready to accept requests
	IsReady(ctx context.Context) (*ReadinessResponse, error)
	
	// IsAlive checks if the system is alive and responsive
	IsAlive(ctx context.Context) (*LivenessResponse, error)
}

// HealthMonitor implements the HealthAPI interface
// This component provides the actual health monitoring logic
type HealthMonitor struct {
	startTime time.Time
	version   string
	components map[string]ComponentStatus
}

// NewHealthMonitor creates a new health monitor instance
func NewHealthMonitor(version string) *HealthMonitor {
	return &HealthMonitor{
		startTime:  time.Now(),
		version:    version,
		components: make(map[string]ComponentStatus),
	}
}

// GetHealth returns basic health status
func (hm *HealthMonitor) GetHealth(ctx context.Context) (*HealthResponse, error) {
	// Determine overall health status
	status := hm.determineOverallStatus()
	
	response := &HealthResponse{
		Status:    status,
		Timestamp: time.Now(),
		Version:   hm.version,
		Uptime:    time.Since(hm.startTime).String(),
	}
	
	return response, nil
}

// GetDetailedHealth returns comprehensive health status
func (hm *HealthMonitor) GetDetailedHealth(ctx context.Context) (*DetailedHealthResponse, error) {
	// Determine overall health status
	status := hm.determineOverallStatus()
	
	// Collect component statuses
	components := make([]ComponentStatus, 0, len(hm.components))
	for _, component := range hm.components {
		components = append(components, component)
	}
	
	response := &DetailedHealthResponse{
		Status:     status,
		Timestamp:  time.Now(),
		Version:    hm.version,
		Uptime:     time.Since(hm.startTime).String(),
		Components: components,
		Metrics:    hm.collectMetrics(),
		Environment: "production", // Could be configurable
	}
	
	return response, nil
}

// IsReady checks if the system is ready to accept requests
func (hm *HealthMonitor) IsReady(ctx context.Context) (*ReadinessResponse, error) {
	// Check if all critical components are healthy
	ready := true
	message := "System is ready"
	
	for _, component := range hm.components {
		if component.Status == HealthStatusUnhealthy {
			ready = false
			message = "System not ready: " + component.Name + " is unhealthy"
			break
		}
	}
	
	response := &ReadinessResponse{
		Ready:     ready,
		Timestamp: time.Now(),
		Message:   message,
	}
	
	return response, nil
}

// IsAlive checks if the system is alive and responsive
func (hm *HealthMonitor) IsAlive(ctx context.Context) (*LivenessResponse, error) {
	// Basic liveness check - system is alive if it can respond
	alive := true
	message := "System is alive"
	
	response := &LivenessResponse{
		Alive:     alive,
		Timestamp: time.Now(),
		Message:   message,
	}
	
	return response, nil
}

// UpdateComponentStatus updates the status of a component
func (hm *HealthMonitor) UpdateComponentStatus(name string, status HealthStatus, message string, details map[string]interface{}) {
	hm.components[name] = ComponentStatus{
		Name:        name,
		Status:      status,
		Message:     message,
		LastChecked: time.Now(),
		Details:     details,
	}
}

// determineOverallStatus determines the overall system health status
func (hm *HealthMonitor) determineOverallStatus() HealthStatus {
	if len(hm.components) == 0 {
		return HealthStatusHealthy
	}
	
	hasUnhealthy := false
	hasDegraded := false
	
	for _, component := range hm.components {
		switch component.Status {
		case HealthStatusUnhealthy:
			hasUnhealthy = true
		case HealthStatusDegraded:
			hasDegraded = true
		}
	}
	
	if hasUnhealthy {
		return HealthStatusUnhealthy
	}
	if hasDegraded {
		return HealthStatusDegraded
	}
	
	return HealthStatusHealthy
}

// collectMetrics collects system metrics
func (hm *HealthMonitor) collectMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})
	
	// Basic metrics
	metrics["uptime_seconds"] = time.Since(hm.startTime).Seconds()
	metrics["component_count"] = len(hm.components)
	
	// Component health summary
	healthyCount := 0
	degradedCount := 0
	unhealthyCount := 0
	
	for _, component := range hm.components {
		switch component.Status {
		case HealthStatusHealthy:
			healthyCount++
		case HealthStatusDegraded:
			degradedCount++
		case HealthStatusUnhealthy:
			unhealthyCount++
		}
	}
	
	metrics["components_healthy"] = healthyCount
	metrics["components_degraded"] = degradedCount
	metrics["components_unhealthy"] = unhealthyCount
	
	return metrics
}
