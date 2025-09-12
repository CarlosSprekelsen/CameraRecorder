/*
MediaMTX RTSP Connection Manager Implementation

Requirements Coverage:
- REQ-MTX-001: MediaMTX service integration
- REQ-MTX-002: Stream management capabilities
- REQ-MTX-003: Path creation and deletion
- REQ-MTX-004: Health monitoring

Test Categories: Unit/Integration
API Documentation Reference: docs/api/swagger.json

RTSP Connection Management for STANAG4606 streaming monitoring
*/

package mediamtx

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
)

// rtspConnectionManager implements RTSPConnectionManager interface
type rtspConnectionManager struct {
	client MediaMTXClient
	config *MediaMTXConfig
	logger *logging.Logger

	// Atomic state: optimized for high-frequency reads
	isHealthy int32 // 0 = false, 1 = true
	lastCheck int64 // Atomic timestamp (UnixNano)

	// Keep mutex only for complex data structures
	mu sync.RWMutex

	// Metrics cache with TTL
	lastConnections *RTSPConnectionList
	lastSessions    *RTSPConnectionSessionList
	metricsCache    map[string]interface{}
	cacheExpiry     int64 // TTL cache expiry timestamp (nanoseconds)
}

// NewRTSPConnectionManager creates a new RTSP connection manager
func NewRTSPConnectionManager(client MediaMTXClient, config *MediaMTXConfig, logger *logging.Logger) RTSPConnectionManager {
	return &rtspConnectionManager{
		client:       client,
		config:       config,
		logger:       logger,
		isHealthy:    1, // Assume healthy initially (1 = true)
		lastCheck:    time.Now().UnixNano(),
		metricsCache: make(map[string]interface{}),
	}
}

// ListConnections lists all RTSP connections
func (rcm *rtspConnectionManager) ListConnections(ctx context.Context, page, itemsPerPage int) (*RTSPConnectionList, error) {
	rcm.logger.WithFields(logging.Fields{
		"page":         strconv.Itoa(page),
		"itemsPerPage": strconv.Itoa(itemsPerPage),
	}).Debug("Listing RTSP connections")

	// Input validation - prevent API errors
	if page < 0 {
		return nil, fmt.Errorf("invalid page number: %d (must be >= 0)", page)
	}
	if itemsPerPage < 1 {
		return nil, fmt.Errorf("invalid items per page: %d (must be >= 1)", itemsPerPage)
	}
	if itemsPerPage > 1000 {
		return nil, fmt.Errorf("invalid items per page: %d (must be <= 1000)", itemsPerPage)
	}

	// Build query parameters
	params := fmt.Sprintf("?page=%d&itemsPerPage=%d", page, itemsPerPage)
	url := "/v3/rtspconns/list" + params

	// Make API call
	data, err := rcm.client.Get(ctx, url)
	if err != nil {
		rcm.logger.WithError(err).Error("Failed to list RTSP connections")
		return nil, fmt.Errorf("failed to list RTSP connections: %w", err)
	}

	// Parse response
	var connectionList RTSPConnectionList
	if err := json.Unmarshal(data, &connectionList); err != nil {
		rcm.logger.WithError(err).Error("Failed to parse RTSP connections response")
		return nil, fmt.Errorf("failed to parse RTSP connections response: %w", err)
	}

	// Cache for metrics
	rcm.mu.Lock()
	rcm.lastConnections = &connectionList
	rcm.mu.Unlock()

	rcm.logger.WithField("count", strconv.Itoa(len(connectionList.Items))).Info("RTSP connections listed successfully")
	return &connectionList, nil
}

// GetConnection gets a specific RTSP connection by ID
func (rcm *rtspConnectionManager) GetConnection(ctx context.Context, id string) (*RTSPConnection, error) {
	rcm.logger.WithField("id", id).Debug("Getting RTSP connection")

	url := fmt.Sprintf("/v3/rtspconns/get/%s", id)

	// Make API call
	data, err := rcm.client.Get(ctx, url)
	if err != nil {
		rcm.logger.WithError(err).WithField("id", id).Error("Failed to get RTSP connection")
		return nil, fmt.Errorf("failed to get RTSP connection %s: %w", id, err)
	}

	// Parse response
	var connection RTSPConnection
	if err := json.Unmarshal(data, &connection); err != nil {
		rcm.logger.WithError(err).WithField("id", id).Error("Failed to parse RTSP connection response")
		return nil, fmt.Errorf("failed to parse RTSP connection response: %w", err)
	}

	rcm.logger.WithField("id", id).Info("RTSP connection retrieved successfully")
	return &connection, nil
}

// ListSessions lists all RTSP sessions
func (rcm *rtspConnectionManager) ListSessions(ctx context.Context, page, itemsPerPage int) (*RTSPConnectionSessionList, error) {
	rcm.logger.WithFields(logging.Fields{
		"page":         strconv.Itoa(page),
		"itemsPerPage": strconv.Itoa(itemsPerPage),
	}).Debug("Listing RTSP sessions")

	// Build query parameters
	params := fmt.Sprintf("?page=%d&itemsPerPage=%d", page, itemsPerPage)
	url := "/v3/rtspsessions/list" + params

	// Make API call
	data, err := rcm.client.Get(ctx, url)
	if err != nil {
		rcm.logger.WithError(err).Error("Failed to list RTSP sessions")
		return nil, fmt.Errorf("failed to list RTSP sessions: %w", err)
	}

	// Parse response
	var sessionList RTSPConnectionSessionList
	if err := json.Unmarshal(data, &sessionList); err != nil {
		rcm.logger.WithError(err).Error("Failed to parse RTSP sessions response")
		return nil, fmt.Errorf("failed to parse RTSP sessions response: %w", err)
	}

	// Cache for metrics
	rcm.mu.Lock()
	rcm.lastSessions = &sessionList
	rcm.mu.Unlock()

	rcm.logger.WithField("count", strconv.Itoa(len(sessionList.Items))).Info("RTSP sessions listed successfully")
	return &sessionList, nil
}

// GetSession gets a specific RTSP session by ID
func (rcm *rtspConnectionManager) GetSession(ctx context.Context, id string) (*RTSPConnectionSession, error) {
	rcm.logger.WithField("id", id).Debug("Getting RTSP session")

	url := fmt.Sprintf("/v3/rtspsessions/get/%s", id)

	// Make API call
	data, err := rcm.client.Get(ctx, url)
	if err != nil {
		rcm.logger.WithError(err).WithField("id", id).Error("Failed to get RTSP session")
		return nil, fmt.Errorf("failed to get RTSP session %s: %w", id, err)
	}

	// Parse response
	var session RTSPConnectionSession
	if err := json.Unmarshal(data, &session); err != nil {
		rcm.logger.WithError(err).WithField("id", id).Error("Failed to parse RTSP session response")
		return nil, fmt.Errorf("failed to parse RTSP session response: %w", err)
	}

	rcm.logger.WithField("id", id).Info("RTSP session retrieved successfully")
	return &session, nil
}

// KickSession kicks out an RTSP session from the server
func (rcm *rtspConnectionManager) KickSession(ctx context.Context, id string) error {
	rcm.logger.WithField("id", id).Info("Kicking RTSP session")

	url := fmt.Sprintf("/v3/rtspsessions/kick/%s", id)

	// Make API call
	_, err := rcm.client.Post(ctx, url, nil)
	if err != nil {
		rcm.logger.WithError(err).WithField("id", id).Error("Failed to kick RTSP session")
		return fmt.Errorf("failed to kick RTSP session %s: %w", id, err)
	}

	rcm.logger.WithField("id", id).Info("RTSP session kicked successfully")
	return nil
}

// GetConnectionHealth returns the health status of RTSP connections
func (rcm *rtspConnectionManager) GetConnectionHealth(ctx context.Context) (*HealthStatus, error) {
	// Check if monitoring is enabled
	if !rcm.config.RTSPMonitoring.Enabled {
		return &HealthStatus{
			Status:    "disabled",
			Details:   "RTSP monitoring is disabled",
			Timestamp: time.Now(),
		}, nil
	}

	// Try to get current connections to check health
	_, err := rcm.ListConnections(ctx, 0, 10)
	if err != nil {
		atomic.StoreInt32(&rcm.isHealthy, 0)
		atomic.StoreInt64(&rcm.lastCheck, time.Now().UnixNano())
		return &HealthStatus{
			Status:    "unhealthy",
			Details:   fmt.Sprintf("Failed to list RTSP connections: %v", err),
			Timestamp: time.Now(),
		}, nil
	}

	// Check connection limits
	rcm.mu.RLock()
	connectionCount := 0
	if rcm.lastConnections != nil {
		connectionCount = len(rcm.lastConnections.Items)
	}
	rcm.mu.RUnlock()

	if connectionCount > rcm.config.RTSPMonitoring.MaxConnections {
		atomic.StoreInt32(&rcm.isHealthy, 0)
		atomic.StoreInt64(&rcm.lastCheck, time.Now().UnixNano())
		return &HealthStatus{
			Status:    "overloaded",
			Details:   fmt.Sprintf("Too many RTSP connections: %d > %d", connectionCount, rcm.config.RTSPMonitoring.MaxConnections),
			Timestamp: time.Now(),
		}, nil
	}

	atomic.StoreInt32(&rcm.isHealthy, 1)
	atomic.StoreInt64(&rcm.lastCheck, time.Now().UnixNano())
	return &HealthStatus{
		Status:    "healthy",
		Details:   fmt.Sprintf("RTSP connections healthy: %d connections", connectionCount),
		Timestamp: time.Now(),
	}, nil
}

// GetConnectionMetrics returns metrics about RTSP connections with TTL caching
func (rcm *rtspConnectionManager) GetConnectionMetrics(ctx context.Context) map[string]interface{} {
	now := time.Now().UnixNano()

	rcm.mu.RLock()
	// Check if cache is still valid (5 second TTL)
	if rcm.cacheExpiry > now && rcm.metricsCache != nil && len(rcm.metricsCache) > 0 {
		defer rcm.mu.RUnlock()
		rcm.logger.Debug("Returning cached RTSP connection metrics")
		return rcm.metricsCache
	}
	rcm.mu.RUnlock()

	// Cache expired, rebuild metrics
	rcm.mu.Lock()
	defer rcm.mu.Unlock()

	// Double-check after acquiring write lock
	if rcm.cacheExpiry > now && rcm.metricsCache != nil && len(rcm.metricsCache) > 0 {
		return rcm.metricsCache
	}

	// Read atomic values
	isHealthy := atomic.LoadInt32(&rcm.isHealthy) == 1
	lastCheckNano := atomic.LoadInt64(&rcm.lastCheck)
	lastCheckTime := time.Unix(0, lastCheckNano)

	metrics := make(map[string]interface{})
	metrics["is_healthy"] = isHealthy
	metrics["last_check"] = lastCheckTime
	metrics["monitoring_enabled"] = rcm.config.RTSPMonitoring.Enabled

	// Connection metrics
	if rcm.lastConnections != nil {
		metrics["total_connections"] = len(rcm.lastConnections.Items)
		metrics["connection_page_count"] = rcm.lastConnections.PageCount
		metrics["connection_item_count"] = rcm.lastConnections.ItemCount

		// Calculate bandwidth metrics
		totalBytesReceived := int64(0)
		totalBytesSent := int64(0)
		for _, conn := range rcm.lastConnections.Items {
			totalBytesReceived += conn.BytesReceived
			totalBytesSent += conn.BytesSent
		}
		metrics["total_bytes_received"] = totalBytesReceived
		metrics["total_bytes_sent"] = totalBytesSent
		metrics["total_bandwidth"] = totalBytesReceived + totalBytesSent
	}

	// Session metrics
	if rcm.lastSessions != nil {
		metrics["total_sessions"] = len(rcm.lastSessions.Items)
		metrics["session_page_count"] = rcm.lastSessions.PageCount
		metrics["session_item_count"] = rcm.lastSessions.ItemCount

		// Calculate session state distribution
		stateCounts := make(map[string]int)
		totalRTPPacketsReceived := int64(0)
		totalRTPPacketsSent := int64(0)
		totalRTPPacketsLost := int64(0)
		totalJitter := float64(0)
		sessionCount := 0

		for _, session := range rcm.lastSessions.Items {
			stateCounts[string(session.State)]++
			totalRTPPacketsReceived += session.RTPPacketsReceived
			totalRTPPacketsSent += session.RTPPacketsSent
			totalRTPPacketsLost += session.RTPPacketsLost
			totalJitter += session.RTPPacketsJitter
			sessionCount++
		}

		metrics["session_states"] = stateCounts
		metrics["total_rtp_packets_received"] = totalRTPPacketsReceived
		metrics["total_rtp_packets_sent"] = totalRTPPacketsSent
		metrics["total_rtp_packets_lost"] = totalRTPPacketsLost

		if sessionCount > 0 {
			metrics["average_jitter"] = totalJitter / float64(sessionCount)
			metrics["packet_loss_rate"] = float64(totalRTPPacketsLost) / float64(totalRTPPacketsReceived+totalRTPPacketsSent)
		}
	}

	// Configuration metrics
	metrics["max_connections"] = rcm.config.RTSPMonitoring.MaxConnections
	metrics["bandwidth_threshold"] = rcm.config.RTSPMonitoring.BandwidthThreshold
	metrics["packet_loss_threshold"] = rcm.config.RTSPMonitoring.PacketLossThreshold
	metrics["jitter_threshold"] = rcm.config.RTSPMonitoring.JitterThreshold

	// Cache the results with 5-second TTL
	rcm.metricsCache = metrics
	rcm.cacheExpiry = now + (5 * time.Second).Nanoseconds()

	return metrics
}
