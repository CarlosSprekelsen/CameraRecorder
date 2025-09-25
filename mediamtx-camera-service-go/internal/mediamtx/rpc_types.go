/*
JSON-RPC API Response Types - Single Source of Truth

This file contains ALL JSON-RPC API response type definitions to ensure
consistency with the API documentation and prevent schema drift.

Requirements Coverage:
- REQ-API-001: JSON-RPC 2.0 API compliance
- REQ-API-002: Response format consistency
- REQ-API-003: Type safety for API responses

API Documentation Reference: docs/api/json_rpc_methods.md
*/

package mediamtx

// ============================================================================
// JSON-RPC API RESPONSE TYPES - EXACT API SPECIFICATION MATCH
// ============================================================================

// TakeSnapshotResponse represents the response from take_snapshot method
type TakeSnapshotResponse struct {
	Device    string `json:"device"`    // Camera device identifier (e.g., "camera0")
	Filename  string `json:"filename"`  // Generated snapshot filename
	Status    string `json:"status"`    // Snapshot status ("SUCCESS", "FAILED")
	Timestamp string `json:"timestamp"` // Snapshot capture timestamp (ISO 8601)
	FileSize  int64  `json:"file_size"` // File size in bytes
	FilePath  string `json:"file_path"` // Full file path to saved snapshot
}

// StartRecordingResponse represents the response from start_recording method
type StartRecordingResponse struct {
	Device    string `json:"device"`     // Camera device identifier
	Filename  string `json:"filename"`   // Generated recording filename
	Status    string `json:"status"`     // Recording status ("RECORDING", "FAILED")
	StartTime string `json:"start_time"` // Recording start timestamp (ISO 8601)
	Format    string `json:"format"`     // Recording format ("fmp4", "mp4")
}

// StopRecordingResponse represents the response from stop_recording method
type StopRecordingResponse struct {
	Device    string  `json:"device"`     // Camera device identifier
	Filename  string  `json:"filename"`   // Recording filename
	Status    string  `json:"status"`     // Recording status ("STOPPED")
	StartTime string  `json:"start_time"` // Recording start timestamp (ISO 8601)
	EndTime   string  `json:"end_time"`   // Recording end timestamp (ISO 8601)
	Duration  float64 `json:"duration"`   // Recording duration in seconds
	FileSize  int64   `json:"file_size"`  // File size in bytes
	Format    string  `json:"format"`     // Recording format
}

// AuthenticateResponse represents the response from authenticate method
type AuthenticateResponse struct {
	Success   bool   `json:"success"`    // Authentication success status
	Token     string `json:"token"`      // JWT authentication token
	Role      string `json:"role"`       // User role (viewer, operator, admin)
	ExpiresAt int64  `json:"expires_at"` // Token expiration timestamp (Unix)
	Message   string `json:"message"`    // Success/error message
}

// CameraListResponse represents the response from get_camera_list method
type CameraListResponse struct {
	Cameras   []CameraInfo `json:"cameras"`   // List of discovered cameras
	Total     int          `json:"total"`     // Total number of cameras
	Connected int          `json:"connected"` // Number of connected cameras
}

// CameraInfo represents camera information for API responses
type CameraInfo struct {
	Device     string            `json:"device"`     // Camera device identifier
	Status     string            `json:"status"`     // Camera status (CONNECTED, DISCONNECTED, ERROR)
	Name       string            `json:"name"`       // Human-readable camera name
	Resolution string            `json:"resolution"` // Current resolution setting
	FPS        int               `json:"fps"`        // Frames per second
	Streams    map[string]string `json:"streams"`    // Available stream URLs
}

// GetCameraStatusResponse represents the response from get_camera_status method
type GetCameraStatusResponse struct {
	Device       string                 `json:"device"`                 // Camera device identifier
	Status       string                 `json:"status"`                 // Camera status
	Name         string                 `json:"name"`                   // Camera name
	Resolution   string                 `json:"resolution"`             // Current resolution
	FPS          int                    `json:"fps"`                    // Current FPS
	Streams      map[string]string      `json:"streams"`                // Available stream URLs
	Metrics      map[string]interface{} `json:"metrics,omitempty"`      // Performance metrics
	Capabilities map[string]interface{} `json:"capabilities,omitempty"` // Camera capabilities
}

// GetCameraCapabilitiesResponse represents the response from get_camera_capabilities method
type GetCameraCapabilitiesResponse struct {
	Device           string   `json:"device"`            // Camera device identifier
	Formats          []string `json:"formats"`           // Supported formats
	Resolutions      []string `json:"resolutions"`       // Supported resolutions
	FpsOptions       []int    `json:"fps_options"`       // Supported frame rates (API documentation field name)
	FrameRates       []int    `json:"frame_rates"`       // Supported frame rates (legacy field name)
	Capabilities     []string `json:"capabilities"`      // Camera capabilities
	ValidationStatus string   `json:"validation_status"` // Capability validation status ("none", "disconnected", "confirmed")
}

// StartStreamingResponse represents the response from start_streaming method
type StartStreamingResponse struct {
	Device         string `json:"device"`           // Camera device identifier
	StreamName     string `json:"stream_name"`      // Generated stream name
	StreamURL      string `json:"stream_url"`       // Generated stream URL
	Status         string `json:"status"`           // Stream status ("STARTED", "failed")
	StartTime      string `json:"start_time"`       // Streaming start timestamp (ISO 8601) - API compliant
	AutoCloseAfter string `json:"auto_close_after"` // Auto-close timeout setting
	FfmpegCommand  string `json:"ffmpeg_command"`   // FFmpeg command used
	Format         string `json:"format"`           // Stream format ("rtsp")
	Message        string `json:"message"`          // Success message
}

// StopStreamingResponse represents the response from stop_streaming method
type StopStreamingResponse struct {
	Device          string `json:"device"`           // Camera device identifier
	StreamName      string `json:"stream_name"`      // Stream name
	Status          string `json:"status"`           // Operation status ("STOPPED")
	StartTime       string `json:"start_time"`       // Streaming start timestamp (ISO 8601) - API compliant
	EndTime         string `json:"end_time"`         // Streaming end timestamp (ISO 8601) - API compliant
	Duration        int    `json:"duration"`         // Total streaming duration in seconds
	StreamContinues bool   `json:"stream_continues"` // Whether stream continues for other consumers
	Message         string `json:"message"`          // Success message
}

// GetStreamURLResponse represents the response from get_stream_url method
type GetStreamURLResponse struct {
	Device          string `json:"device"`           // Camera device identifier
	StreamName      string `json:"stream_name"`      // Generated stream name
	StreamURL       string `json:"stream_url"`       // Stream URL for consumption
	Available       bool   `json:"available"`        // Whether stream is available
	ActiveConsumers int    `json:"active_consumers"` // Number of active stream consumers
	StreamStatus    string `json:"stream_status"`    // Stream readiness status ("READY", "NOT_READY", "ERROR")
}

// GetStreamStatusResponse represents the response from get_stream_status method
type GetStreamStatusResponse struct {
	Device       string `json:"device"`        // Camera device identifier
	StreamName   string `json:"stream_name"`   // Generated stream name
	StreamURL    string `json:"stream_url"`    // Stream URL
	Status       string `json:"status"`        // Stream status ("active", "inactive")
	Ready        bool   `json:"ready"`         // Stream readiness status
	Viewers      int    `json:"viewers"`       // Current viewer count
	StartTime    string `json:"start_time"`    // Stream start time (ISO 8601)
	LastActivity string `json:"last_activity"` // Last activity time (ISO 8601)
	BytesSent    int64  `json:"bytes_sent"`    // Total bytes sent
}

// ListRecordingsResponse represents the response from list_recordings method
type ListRecordingsResponse struct {
	Files  []RecordingFileInfo `json:"files"`  // List of recording files
	Total  int                 `json:"total"`  // Total number of recordings
	Limit  int                 `json:"limit"`  // Requested limit
	Offset int                 `json:"offset"` // Requested offset
}

// RecordingFileInfo represents recording file information for API responses
type RecordingFileInfo struct {
	Device       string  `json:"device"`        // Camera device identifier
	Filename     string  `json:"filename"`      // Recording filename
	FileSize     int64   `json:"file_size"`     // File size in bytes
	Duration     float64 `json:"duration"`      // Recording duration in seconds
	ModifiedTime string  `json:"modified_time"` // File modification timestamp (ISO 8601) - API compliant
	Format       string  `json:"format"`        // Recording format
	DownloadURL  string  `json:"download_url"`  // Download URL for the file
}

// ListSnapshotsResponse represents the response from list_snapshots method
type ListSnapshotsResponse struct {
	Snapshots []SnapshotFileInfo `json:"snapshots"` // List of snapshot files
	Total     int                `json:"total"`     // Total number of snapshots
	Limit     int                `json:"limit"`     // Requested limit
	Offset    int                `json:"offset"`    // Requested offset
}

// SnapshotFileInfo represents snapshot file information for API responses
type SnapshotFileInfo struct {
	Device       string `json:"device"`        // Camera device identifier
	Filename     string `json:"filename"`      // Snapshot filename
	FileSize     int64  `json:"file_size"`     // File size in bytes
	ModifiedTime string `json:"modified_time"` // File modification timestamp (ISO 8601) - API compliant
	Format       string `json:"format"`        // Image format
	Resolution   string `json:"resolution"`    // Image resolution
	DownloadURL  string `json:"download_url"`  // Download URL for the file
}

// GetRecordingInfoResponse represents the response from get_recording_info method
type GetRecordingInfoResponse struct {
	Filename    string  `json:"filename"`     // Recording filename
	FileSize    int64   `json:"file_size"`    // File size in bytes
	Duration    float64 `json:"duration"`     // Recording duration in seconds
	CreatedTime string  `json:"created_time"` // Creation timestamp (ISO 8601) - API compliant
	Format      string  `json:"format"`       // Recording format
	Device      string  `json:"device"`       // Camera device identifier
}

// GetSnapshotInfoResponse represents the response from get_snapshot_info method
type GetSnapshotInfoResponse struct {
	Filename    string `json:"filename"`     // Snapshot filename
	FileSize    int64  `json:"file_size"`    // File size in bytes
	CreatedTime string `json:"created_time"` // Creation timestamp (ISO 8601) - API compliant
	Format      string `json:"format"`       // Image format
	Resolution  string `json:"resolution"`   // Image resolution
	Device      string `json:"device"`       // Camera device identifier
}

// DeleteRecordingResponse represents the response from delete_recording method
type DeleteRecordingResponse struct {
	Filename  string `json:"filename"`  // Deleted recording filename
	Status    string `json:"status"`    // Operation status ("deleted")
	Message   string `json:"message"`   // Success message
	Timestamp string `json:"timestamp"` // Deletion timestamp (ISO 8601)
}

// DeleteSnapshotResponse represents the response from delete_snapshot method
type DeleteSnapshotResponse struct {
	Filename  string `json:"filename"`  // Deleted snapshot filename
	Status    string `json:"status"`    // Operation status ("deleted")
	Message   string `json:"message"`   // Success message
	Timestamp string `json:"timestamp"` // Deletion timestamp (ISO 8601)
}

// GetStorageInfoResponse represents the response from get_storage_info method
type GetStorageInfoResponse struct {
	TotalSpace     int64   `json:"total_space"`     // Total storage space in bytes
	UsedSpace      int64   `json:"used_space"`      // Used storage space in bytes
	AvailableSpace int64   `json:"available_space"` // Available storage space in bytes
	UsagePercent   float64 `json:"usage_percent"`   // Usage percentage
	RecordingsSize int64   `json:"recordings_size"` // Size of recordings directory
	SnapshotsSize  int64   `json:"snapshots_size"`  // Size of snapshots directory
}

// CleanupOldFilesResponse represents the response from cleanup_old_files method
type CleanupOldFilesResponse struct {
	RecordingsRemoved int    `json:"recordings_removed"` // Number of recordings removed
	SnapshotsRemoved  int    `json:"snapshots_removed"`  // Number of snapshots removed
	SpaceFreed        int64  `json:"space_freed"`        // Space freed in bytes
	Status            string `json:"status"`             // Operation status
	Message           string `json:"message"`            // Success message
}

// GetStreamsResponse represents the response from get_streams method
type GetStreamsResponse struct {
	Streams   []StreamInfo `json:"streams"`   // List of active streams
	Total     int          `json:"total"`     // Total number of streams
	Active    int          `json:"active"`    // Number of active streams
	Inactive  int          `json:"inactive"`  // Number of inactive streams
	Timestamp string       `json:"timestamp"` // Response timestamp (ISO 8601)
}

// StreamInfo represents stream information for API responses
type StreamInfo struct {
	Name         string `json:"name"`          // Stream name/identifier
	Status       string `json:"status"`        // Stream status ("active", "inactive")
	Source       string `json:"source"`        // Stream source type
	Viewers      int    `json:"viewers"`       // Current viewer count
	StartTime    string `json:"start_time"`    // Stream start timestamp (ISO 8601)
	LastActivity string `json:"last_activity"` // Last activity timestamp (ISO 8601)
	BytesSent    int64  `json:"bytes_sent"`    // Total bytes sent
}

// GetHealthResponse represents the response from get_health method
type GetHealthResponse struct {
	Status       string                 `json:"status"`        // Overall health status ("healthy", "degraded", "unhealthy")
	Uptime       float64                `json:"uptime"`        // System uptime in seconds with sub-second precision
	Version      string                 `json:"version"`       // Service version
	Components   map[string]interface{} `json:"components"`    // Component health details
	Checks       []interface{}          `json:"checks"`        // Health check results
	Timestamp    string                 `json:"timestamp"`     // Health check timestamp (ISO 8601)
	ResponseTime float64                `json:"response_time"` // Health check response time in ms
}

// GetMetricsResponse represents the response from get_metrics method
type GetMetricsResponse struct {
	Timestamp        string                 `json:"timestamp"`         // Metrics collection timestamp (ISO 8601)
	SystemMetrics    map[string]interface{} `json:"system_metrics"`    // System-level metrics
	CameraMetrics    map[string]interface{} `json:"camera_metrics"`    // Camera-specific metrics
	RecordingMetrics map[string]interface{} `json:"recording_metrics"` // Recording performance metrics
	StreamMetrics    map[string]interface{} `json:"stream_metrics"`    // Streaming metrics
}

// GetSystemMetricsResponse represents the response from get_system_metrics method
type GetSystemMetricsResponse struct {
	CPUUsage    float64 `json:"cpu_usage"`    // CPU usage percentage
	MemoryUsage float64 `json:"memory_usage"` // Memory usage percentage
	DiskUsage   float64 `json:"disk_usage"`   // Disk usage percentage
	Goroutines  int     `json:"goroutines"`   // Number of active goroutines
	Timestamp   string  `json:"timestamp"`    // Metrics collection timestamp (ISO 8601)
}

// GetServerInfoResponse represents the response from get_server_info method
type GetServerInfoResponse struct {
	Name             string   `json:"name"`              // Service name
	Version          string   `json:"version"`           // Service version
	BuildDate        string   `json:"build_date"`        // Build date
	GoVersion        string   `json:"go_version"`        // Go version
	Architecture     string   `json:"architecture"`      // System architecture
	Capabilities     []string `json:"capabilities"`      // Service capabilities
	SupportedFormats []string `json:"supported_formats"` // Supported file formats
	MaxCameras       int      `json:"max_cameras"`       // Maximum supported cameras
}

// SetRetentionPolicyResponse represents the response from set_retention_policy method
type SetRetentionPolicyResponse struct {
	Success    bool   `json:"success"`     // Operation success status
	PolicyType string `json:"policy_type"` // Policy type applied
	MaxAge     string `json:"max_age"`     // Maximum age setting
	MaxSize    string `json:"max_size"`    // Maximum size setting
	Message    string `json:"message"`     // Success/status message
}

// DiscoverExternalStreamsResponse represents the response from discover_external_streams method
type DiscoverExternalStreamsResponse struct {
	DiscoveredStreams []ExternalStreamInfo `json:"discovered_streams"` // All discovered streams
	SkydioStreams     []ExternalStreamInfo `json:"skydio_streams"`     // Skydio-specific streams
	GenericStreams    []ExternalStreamInfo `json:"generic_streams"`    // Generic RTSP streams
	ScanTimestamp     int64                `json:"scan_timestamp"`     // Unix timestamp of scan
	TotalFound        int                  `json:"total_found"`        // Total streams found
	DiscoveryOptions  DiscoveryOptionsInfo `json:"discovery_options"`  // Options used for discovery
	ScanDuration      string               `json:"scan_duration"`      // Duration as string (e.g., "2.5s")
	Errors            []string             `json:"errors"`             // Discovery errors
}

// ExternalStreamInfo represents external stream information for API responses
type ExternalStreamInfo struct {
	URL          string                 `json:"url"`                // Stream URL
	Type         string                 `json:"type"`               // Stream type (skydio_stanag4609, generic_rtsp)
	Name         string                 `json:"name"`               // Stream name/identifier
	Status       string                 `json:"status"`             // Stream status (discovered, connected, error)
	DiscoveredAt string                 `json:"discovered_at"`      // Discovery timestamp (ISO 8601)
	LastSeen     string                 `json:"last_seen"`          // Last seen timestamp (ISO 8601)
	Capabilities map[string]interface{} `json:"capabilities"`       // Stream capabilities
	Metadata     map[string]interface{} `json:"metadata,omitempty"` // Additional metadata
}

// DiscoveryOptionsInfo represents discovery options used in API responses
type DiscoveryOptionsInfo struct {
	SkydioEnabled  bool `json:"skydio_enabled"`  // Skydio discovery enabled
	GenericEnabled bool `json:"generic_enabled"` // Generic discovery enabled
	ForceRescan    bool `json:"force_rescan"`    // Force rescan flag
	IncludeOffline bool `json:"include_offline"` // Include offline streams
}

// AddExternalStreamResponse represents the response from add_external_stream method
type AddExternalStreamResponse struct {
	StreamURL  string `json:"stream_url"`  // Added stream URL
	StreamName string `json:"stream_name"` // Stream name
	StreamType string `json:"stream_type"` // Stream type
	Status     string `json:"status"`      // Operation status ("added")
	Timestamp  int64  `json:"timestamp"`   // Unix timestamp when added
}

// RemoveExternalStreamResponse represents the response from remove_external_stream method
type RemoveExternalStreamResponse struct {
	StreamURL string `json:"stream_url"` // Removed stream URL
	Status    string `json:"status"`     // Operation status ("removed")
	Timestamp int64  `json:"timestamp"`  // Unix timestamp when removed
}

// GetExternalStreamsResponse represents the response from get_external_streams method
type GetExternalStreamsResponse struct {
	ExternalStreams []ExternalStreamInfo `json:"external_streams"` // All external streams
	SkydioStreams   []ExternalStreamInfo `json:"skydio_streams"`   // Skydio-specific streams
	GenericStreams  []ExternalStreamInfo `json:"generic_streams"`  // Generic RTSP streams
	TotalCount      int                  `json:"total_count"`      // Total stream count
	Timestamp       int64                `json:"timestamp"`        // Response timestamp (Unix)
}

// SetDiscoveryIntervalResponse represents the response from set_discovery_interval method
type SetDiscoveryIntervalResponse struct {
	ScanInterval int    `json:"scan_interval"` // Configured scan interval in seconds
	Status       string `json:"status"`        // Operation status
	Message      string `json:"message"`       // Success message
	Timestamp    int64  `json:"timestamp"`     // Unix timestamp when configured
}

// SubscribeResponse represents the response from subscribe_events method
type SubscribeResponse struct {
	Topic     string `json:"topic"`     // Subscribed topic
	Status    string `json:"status"`    // Subscription status ("subscribed")
	Message   string `json:"message"`   // Success message
	ClientID  string `json:"client_id"` // Client identifier
	Timestamp string `json:"timestamp"` // Subscription timestamp (ISO 8601)
}

// UnsubscribeResponse represents the response from unsubscribe_events method
type UnsubscribeResponse struct {
	Topic     string `json:"topic"`     // Unsubscribed topic
	Status    string `json:"status"`    // Unsubscription status ("unsubscribed")
	Message   string `json:"message"`   // Success message
	ClientID  string `json:"client_id"` // Client identifier
	Timestamp string `json:"timestamp"` // Unsubscription timestamp (ISO 8601)
}

// GetSubscriptionStatsResponse represents the response from get_subscription_stats method
type GetSubscriptionStatsResponse struct {
	GlobalStats  SubscriptionGlobalStats `json:"global_stats"`  // Global subscription statistics
	ClientTopics []string                `json:"client_topics"` // Topics subscribed by this client
	ClientID     string                  `json:"client_id"`     // Client identifier
}

// SubscriptionGlobalStats represents global subscription statistics
type SubscriptionGlobalStats struct {
	TotalSubscriptions int            `json:"total_subscriptions"` // Total number of subscriptions
	ActiveClients      int            `json:"active_clients"`      // Number of active clients
	TopicCounts        map[string]int `json:"topic_counts"`        // Subscription count per topic
}

// GetStatusResponse represents the response from get_status method
type GetStatusResponse struct {
	Status     string                 `json:"status"`     // Overall system status ("HEALTHY", "DEGRADED", "UNHEALTHY")
	Uptime     float64                `json:"uptime"`     // System uptime in seconds with sub-second precision
	Version    string                 `json:"version"`    // Service version
	Components map[string]interface{} `json:"components"` // Component operational states
}

// CameraStatusUpdateResponse represents the response from camera_status_update method
type CameraStatusUpdateResponse struct {
	Device    string `json:"device"`    // Camera device identifier
	Status    string `json:"status"`    // Updated camera status
	Message   string `json:"message"`   // Update confirmation message
	Timestamp string `json:"timestamp"` // Update timestamp (ISO 8601)
}

// RecordingStatusUpdateResponse represents the response from recording_status_update method
type RecordingStatusUpdateResponse struct {
	Device    string `json:"device"`    // Camera device identifier
	Status    string `json:"status"`    // Updated recording status
	Message   string `json:"message"`   // Update confirmation message
	Timestamp string `json:"timestamp"` // Update timestamp (ISO 8601)
}

// ============================================================================
// NOTE: No builder functions needed - use direct struct initialization
// This eliminates mismatch errors and makes the code more maintainable
// ============================================================================
