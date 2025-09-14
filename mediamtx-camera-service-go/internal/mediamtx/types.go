/*
MediaMTX Integration Types

Requirements Coverage:
- REQ-MTX-001: MediaMTX service integration
- REQ-MTX-002: Stream management capabilities
- REQ-MTX-003: Path creation and deletion
- REQ-MTX-004: Health monitoring

Test Categories: Unit/Integration
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package mediamtx

import (
	"context"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/camera"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
)

// EventNotifier interfaces for MediaMTX Controller event publishing
// These interfaces allow MediaMTX Controller to publish events without direct WebSocket dependencies

// DeviceToCameraIDMapper interface for converting device paths to camera IDs
// This ensures event payloads use proper abstraction (camera0 vs /dev/video0)
type DeviceToCameraIDMapper interface {
	GetCameraForDevicePath(devicePath string) (string, bool)
	GetDevicePathForCamera(cameraID string) (string, bool)
}

// MediaMTXEventNotifier interface for MediaMTX-specific events
type MediaMTXEventNotifier interface {
	NotifyRecordingStarted(device, sessionID, filename string)
	NotifyRecordingStopped(device, sessionID, filename string, duration time.Duration)
	NotifyStreamStarted(device, streamID, streamType string)
	NotifyStreamStopped(device, streamID, streamType string)
}

// FFmpegConfig represents FFmpeg-specific configuration settings
type FFmpegConfig struct {
	Snapshot  SnapshotConfig  `mapstructure:"snapshot"`
	Recording RecordingConfig `mapstructure:"recording"`
	// Fallback defaults for when configuration is missing
	FallbackDefaults FFmpegFallbackDefaults `mapstructure:"fallback_defaults"`
}

// FFmpegFallbackDefaults represents fallback defaults for FFmpeg operations
type FFmpegFallbackDefaults struct {
	RetryDelay             time.Duration `mapstructure:"retry_delay"`              // Default: 1.0 second
	ProcessCreationTimeout time.Duration `mapstructure:"process_creation_timeout"` // Default: 10.0 seconds
	ExecutionTimeout       time.Duration `mapstructure:"execution_timeout"`        // Default: 30.0 seconds
	MaxBackoffDelay        time.Duration `mapstructure:"max_backoff_delay"`        // Default: 30.0 seconds
}

// SnapshotConfig represents snapshot operation configuration
type SnapshotConfig struct {
	ProcessCreationTimeout time.Duration `mapstructure:"process_creation_timeout"` // Default: 10.0s
	ExecutionTimeout       time.Duration `mapstructure:"execution_timeout"`        // Default: 30.0s
	InternalTimeout        int64         `mapstructure:"internal_timeout"`         // Default: 5000000
	RetryAttempts          int           `mapstructure:"retry_attempts"`           // Default: 2
	RetryDelay             time.Duration `mapstructure:"retry_delay"`              // Default: 1.0s
}

// RecordingConfig represents recording operation configuration
type RecordingConfig struct {
	ProcessCreationTimeout time.Duration `mapstructure:"process_creation_timeout"` // Default: 15.0s
	ExecutionTimeout       time.Duration `mapstructure:"execution_timeout"`        // Default: 60.0s
	InternalTimeout        int64         `mapstructure:"internal_timeout"`         // Default: 10000000
	RetryAttempts          int           `mapstructure:"retry_attempts"`           // Default: 3
	RetryDelay             time.Duration `mapstructure:"retry_delay"`              // Default: 2.0s
}

// PerformanceConfig represents performance configuration settings
type PerformanceConfig struct {
	ResponseTimeTargets map[string]float64 `mapstructure:"response_time_targets"`
	SnapshotTiers       map[string]float64 `mapstructure:"snapshot_tiers"`
	Optimization        OptimizationConfig `mapstructure:"optimization"`
}

// OptimizationConfig represents optimization settings
type OptimizationConfig struct {
	EnableCaching           bool          `mapstructure:"enable_caching"`            // Default: true
	CacheTTL                time.Duration `mapstructure:"cache_ttl"`                 // Default: 300s
	MaxConcurrentOperations int           `mapstructure:"max_concurrent_operations"` // Default: 5
	ConnectionPoolSize      int           `mapstructure:"connection_pool_size"`      // Default: 10
}

// MediaMTXConfig is now imported from internal/config package
// This removes the duplicate struct definition

// HealthMonitorDefaults represents health monitoring default values
type HealthMonitorDefaults struct {
	CheckInterval   time.Duration `mapstructure:"check_interval"`    // Default: 5.0 seconds
	MaxBackoffDelay time.Duration `mapstructure:"max_backoff_delay"` // Default: 30.0 seconds
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`  // Default: 30.0 seconds
}

// CircuitBreakerConfig represents circuit breaker configuration
type CircuitBreakerConfig struct {
	FailureThreshold int           `mapstructure:"failure_threshold"`
	RecoveryTimeout  time.Duration `mapstructure:"recovery_timeout"`
	MaxFailures      int           `mapstructure:"max_failures"`
}

// ConnectionPoolConfig represents HTTP connection pool configuration
type ConnectionPoolConfig struct {
	MaxIdleConns        int           `mapstructure:"max_idle_conns"`
	MaxIdleConnsPerHost int           `mapstructure:"max_idle_conns_per_host"`
	IdleConnTimeout     time.Duration `mapstructure:"idle_conn_timeout"`
}

// Stream type removed - use Path from api_types.go instead
// This eliminates duplicate type definitions and schema drift issues

// Note: MediaMTX API types are now defined in api_types.go for single source of truth
// Legacy aliases are provided in api_types.go for backward compatibility

// HealthStatus represents MediaMTX service health status
type HealthStatus struct {
	Status              string            `json:"status"`
	Timestamp           time.Time         `json:"timestamp"`
	Details             string            `json:"details,omitempty"`
	Metrics             Metrics           `json:"metrics,omitempty"`
	ComponentStatus     map[string]string `json:"component_status,omitempty"`
	ErrorCount          int64             `json:"error_count"`
	LastCheck           time.Time         `json:"last_check"`
	CircuitBreakerState string            `json:"circuit_breaker_state"`
}

// SystemMetrics represents system performance metrics
type SystemMetrics struct {
	RequestCount        int64             `json:"request_count"`
	ResponseTime        float64           `json:"response_time"`
	ErrorCount          int64             `json:"error_count"`
	ActiveConnections   int64             `json:"active_connections"`
	MemoryUsage         float64           `json:"memory_usage"` // Memory usage in MB
	CPUUsage            float64           `json:"cpu_usage"`    // CPU usage percentage
	Goroutines          int               `json:"goroutines"`   // Number of goroutines
	HeapAlloc           int64             `json:"heap_alloc"`   // Heap allocation in bytes
	ComponentStatus     map[string]string `json:"component_status,omitempty"`
	ErrorCounts         map[string]int64  `json:"error_counts,omitempty"`
	LastCheck           time.Time         `json:"last_check"`
	CircuitBreakerState string            `json:"circuit_breaker_state"`
	// Camera metrics from camera monitor
	CameraMonitorMetrics *CameraMonitorMetrics `json:"camera_monitor_metrics,omitempty"`
}

// CameraMonitorMetrics represents camera monitor statistics
type CameraMonitorMetrics struct {
	DevicesConnected           int64   `json:"devices_connected"`
	DeviceEventsProcessed      int64   `json:"device_events_processed"`
	DeviceEventsDropped        int64   `json:"device_events_dropped"`
	UdevEventsProcessed        int64   `json:"udev_events_processed"`
	UdevEventsFiltered         int64   `json:"udev_events_filtered"`
	UdevEventsSkipped          int64   `json:"udev_events_skipped"`
	PollingCycles              int64   `json:"polling_cycles"`
	CapabilityProbesAttempted  int64   `json:"capability_probes_attempted"`
	CapabilityProbesSuccessful int64   `json:"capability_probes_successful"`
	CapabilityTimeouts         int64   `json:"capability_timeouts"`
	CapabilityParseErrors      int64   `json:"capability_parse_errors"`
	PollingFailureCount        int64   `json:"polling_failure_count"`
	CurrentPollInterval        float64 `json:"current_poll_interval"`
	KnownDevicesCount          int64   `json:"known_devices_count"`
	ActiveTasks                int64   `json:"active_tasks"`
	Running                    bool    `json:"running"`
}

// ServerInfo represents server information and capabilities
type ServerInfo struct {
	Name             string   `json:"name"`
	Version          string   `json:"version"`
	BuildDate        string   `json:"build_date"`
	GoVersion        string   `json:"go_version"`
	Architecture     string   `json:"architecture"`
	Capabilities     []string `json:"capabilities"`
	SupportedFormats []string `json:"supported_formats"`
	MaxCameras       int      `json:"max_cameras"`
}

// Metrics represents MediaMTX service metrics
type Metrics struct {
	ActiveStreams int     `json:"active_streams"`
	TotalStreams  int     `json:"total_streams"`
	CPUUsage      float64 `json:"cpu_usage"`
	MemoryUsage   float64 `json:"memory_usage"`
	Uptime        int64   `json:"uptime"`
}

// RecordingSession represents a recording session
type RecordingSession struct {
	ID             string        `json:"id"`
	Device         string        `json:"device"`
	DevicePath     string        `json:"device_path"`
	Path           string        `json:"path"`
	Status         string        `json:"status"`
	StartTime      time.Time     `json:"start_time"`
	EndTime        *time.Time    `json:"end_time,omitempty"`
	Duration       time.Duration `json:"duration"`
	FileSize       int64         `json:"file_size"`
	FilePath       string        `json:"file_path"`
	ContinuityID   string        `json:"continuity_id,omitempty"`
	State          SessionState  `json:"state"`
	Segments       []string      `json:"segments,omitempty"`
	CurrentSegment string        `json:"current_segment,omitempty"`
	PID            int           `json:"pid,omitempty"` // FFmpeg process ID for proper process management

	// Enhanced use case management (Phase 2 enhancement)
	UseCase       StreamUseCase `json:"use_case"`       // "recording", "viewing", "snapshot"
	Priority      int           `json:"priority"`       // Priority level (1=high, 2=medium, 3=low)
	AutoCleanup   bool          `json:"auto_cleanup"`   // Auto-cleanup when session ends
	RetentionDays int           `json:"retention_days"` // Days to retain files
	Quality       string        `json:"quality"`        // Recording quality (low, medium, high)
	MaxDuration   time.Duration `json:"max_duration"`   // Maximum recording duration
	AutoRotate    bool          `json:"auto_rotate"`    // Auto-rotate files
	RotationSize  int64         `json:"rotation_size"`  // Size threshold for rotation
}

// Snapshot represents a camera snapshot
type Snapshot struct {
	ID       string                 `json:"id"`
	Device   string                 `json:"device"`
	Path     string                 `json:"path"`
	FilePath string                 `json:"file_path"`
	Size     int64                  `json:"size"`
	Created  time.Time              `json:"created"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// SnapshotOptions represents snapshot capture options
type SnapshotOptions struct {
	Quality    int    `json:"quality"`    // Image quality (1-100)
	Format     string `json:"format"`     // Image format (jpg, png)
	Resolution string `json:"resolution"` // Resolution (e.g., "1920x1080")
	Timestamp  bool   `json:"timestamp"`  // Include timestamp in filename
}

// RecordingOptions represents recording options
type RecordingOptions struct {
	Duration     int    `json:"duration"`      // Recording duration in seconds (0 = unlimited)
	Quality      string `json:"quality"`       // Video quality (low, medium, high)
	FileRotation int    `json:"file_rotation"` // File rotation interval in minutes
	SegmentSize  int64  `json:"segment_size"`  // Segment size in bytes
}

// FileListResponse represents a paginated file list response
type FileListResponse struct {
	Files  []*FileMetadata `json:"files"`
	Total  int             `json:"total"`
	Limit  int             `json:"limit"`
	Offset int             `json:"offset"`
}

// FileMetadata represents file metadata for recordings and snapshots
type FileMetadata struct {
	FileName    string    `json:"filename"`
	FileSize    int64     `json:"file_size"`
	CreatedAt   time.Time `json:"created_at"`
	ModifiedAt  time.Time `json:"modified_at"`
	Duration    *int64    `json:"duration,omitempty"` // Duration in seconds for video files
	DownloadURL string    `json:"download_url"`
}

// ActiveRecording represents an active recording session (Phase 2 enhancement)
type ActiveRecording struct {
	SessionID  string    `json:"session_id"`
	DevicePath string    `json:"device_path"`
	StartTime  time.Time `json:"start_time"`
	StreamName string    `json:"stream_name"`
	Status     string    `json:"status"`
}

// CameraListResponse represents the response for camera list operations
// APICameraInfo represents camera information in API-ready format
type APICameraInfo struct {
	Device       string                 `json:"device"`       // API-ready camera ID (camera0)
	Status       string                 `json:"status"`       // Camera status
	Name         string                 `json:"name"`         // Camera name
	Resolution   string                 `json:"resolution"`   // Default resolution
	FPS          int                    `json:"fps"`          // Default FPS
	Streams      map[string]string      `json:"streams"`      // Stream URLs
	Capabilities map[string]interface{} `json:"capabilities"` // Camera capabilities
}

type CameraListResponse struct {
	Cameras   []*APICameraInfo `json:"cameras"`
	Total     int              `json:"total"`
	Connected int              `json:"connected"`
}

// CameraStatusResponse represents the response for camera status operations
type CameraStatusResponse struct {
	Device       string                    `json:"device"`
	Status       string                    `json:"status"`
	Name         string                    `json:"name"`
	Resolution   string                    `json:"resolution"`
	FPS          int                       `json:"fps"`
	Streams      map[string]string         `json:"streams"`
	Metrics      *CameraPerformanceMetrics `json:"metrics,omitempty"`
	Capabilities *camera.V4L2Capabilities  `json:"capabilities,omitempty"`
}

// CameraPerformanceMetrics represents camera performance metrics
type CameraPerformanceMetrics struct {
	BytesSent int64 `json:"bytes_sent"`
	Readers   int   `json:"readers"`
	Uptime    int64 `json:"uptime"`
}

// StorageInfo represents storage space and usage information
type StorageInfo struct {
	TotalSpace      uint64  `json:"total_space"`
	UsedSpace       uint64  `json:"used_space"`
	AvailableSpace  uint64  `json:"available_space"`
	UsagePercentage float64 `json:"usage_percentage"`
	RecordingsSize  int64   `json:"recordings_size"`
	SnapshotsSize   int64   `json:"snapshots_size"`
	LowSpaceWarning bool    `json:"low_space_warning"`
}

// GetUsagePercentage returns the usage percentage
func (s *StorageInfo) GetUsagePercentage() float64 {
	return s.UsagePercentage
}

// GetAvailableSpace returns the available space
func (s *StorageInfo) GetAvailableSpace() int64 {
	return int64(s.AvailableSpace)
}

// GetTotalSpace returns the total space
func (s *StorageInfo) GetTotalSpace() int64 {
	return int64(s.TotalSpace)
}

// IsLowSpaceWarning returns the low space warning status
func (s *StorageInfo) IsLowSpaceWarning() bool {
	return s.LowSpaceWarning
}

// MediaMTXController interface defines MediaMTX operations
type MediaMTXController interface {
	// Camera discovery operations
	GetCameraList(ctx context.Context) (*CameraListResponse, error)
	GetCameraStatus(ctx context.Context, device string) (*CameraStatusResponse, error)
	ValidateCameraDevice(ctx context.Context, device string) (bool, error)

	// Camera abstraction layer (delegate to PathManager)
	GetCameraForDevicePath(devicePath string) (string, bool) // /dev/video0 -> camera0
	GetDevicePathForCamera(cameraID string) (string, bool)   // camera0 -> /dev/video0

	// Health and status
	GetHealth(ctx context.Context) (*HealthStatus, error)
	GetMetrics(ctx context.Context) (*Metrics, error)
	GetSystemMetrics(ctx context.Context) (*SystemMetrics, error)
	GetStorageInfo(ctx context.Context) (*StorageInfo, error)
	GetServerInfo(ctx context.Context) (*ServerInfo, error)
	GetHealthMonitor() HealthMonitor

	// System readiness
	IsReady() bool
	GetReadinessState() map[string]interface{}
	SubscribeToReadiness() <-chan struct{}

	// Configuration management
	CleanupOldFiles(ctx context.Context) (map[string]interface{}, error)
	SetRetentionPolicy(ctx context.Context, enabled bool, policyType string, params map[string]interface{}) (map[string]interface{}, error)

	// Stream management (uses Path from api_types.go)
	GetStreams(ctx context.Context) ([]*Path, error)
	GetStream(ctx context.Context, id string) (*Path, error)
	CreateStream(ctx context.Context, name, source string) (*Path, error)
	DeleteStream(ctx context.Context, id string) error

	// Path management
	GetPaths(ctx context.Context) ([]*Path, error)
	GetPath(ctx context.Context, name string) (*Path, error)
	CreatePath(ctx context.Context, path *Path) error
	DeletePath(ctx context.Context, name string) error

	// External stream discovery
	DiscoverExternalStreams(ctx context.Context, options DiscoveryOptions) (*DiscoveryResult, error)
	AddExternalStream(ctx context.Context, stream *ExternalStream) error
	RemoveExternalStream(ctx context.Context, streamURL string) error
	GetExternalStreams(ctx context.Context) ([]*ExternalStream, error)

	// Recording operations
	StartRecording(ctx context.Context, device, path string) (*RecordingSession, error)
	StopRecording(ctx context.Context, sessionID string) error
	GetRecordingStatus(ctx context.Context, sessionID string) (*RecordingSession, error)

	// Streaming operations
	StartStreaming(ctx context.Context, device string) (*Path, error)
	StopStreaming(ctx context.Context, device string) error
	GetStreamURL(ctx context.Context, device string) (string, error)
	GetStreamStatus(ctx context.Context, device string) (*Path, error)

	// File listing operations
	ListRecordings(ctx context.Context, limit, offset int) (*FileListResponse, error)
	ListSnapshots(ctx context.Context, limit, offset int) (*FileListResponse, error)
	GetRecordingInfo(ctx context.Context, filename string) (*FileMetadata, error)
	GetSnapshotInfo(ctx context.Context, filename string) (*FileMetadata, error)
	DeleteRecording(ctx context.Context, filename string) error
	DeleteSnapshot(ctx context.Context, filename string) error

	// Advanced recording operations
	StartAdvancedRecording(ctx context.Context, device string, options map[string]interface{}) (*RecordingSession, error)
	StopAdvancedRecording(ctx context.Context, sessionID string) error
	GetAdvancedRecordingSession(sessionID string) (*RecordingSession, bool)
	ListAdvancedRecordingSessions() []*RecordingSession
	RotateRecordingFile(ctx context.Context, sessionID string) error
	GetSessionIDByDevice(device string) (string, bool)

	// RTSP Connection Management
	ListRTSPConnections(ctx context.Context, page, itemsPerPage int) (*RTSPConnectionList, error)
	GetRTSPConnection(ctx context.Context, id string) (*RTSPConnection, error)
	ListRTSPSessions(ctx context.Context, page, itemsPerPage int) (*RTSPConnectionSessionList, error)
	GetRTSPSession(ctx context.Context, id string) (*RTSPConnectionSession, error)
	KickRTSPSession(ctx context.Context, id string) error
	GetRTSPConnectionHealth(ctx context.Context) (*HealthStatus, error)
	GetRTSPConnectionMetrics(ctx context.Context) map[string]interface{}

	// Advanced snapshot operations
	TakeAdvancedSnapshot(ctx context.Context, device string, options map[string]interface{}) (*Snapshot, error)
	GetAdvancedSnapshot(snapshotID string) (*Snapshot, bool)
	ListAdvancedSnapshots() []*Snapshot
	DeleteAdvancedSnapshot(ctx context.Context, snapshotID string) error
	GetSnapshotSettings() *SnapshotSettings
	UpdateSnapshotSettings(settings *SnapshotSettings)

	// Configuration
	GetConfig(ctx context.Context) (*config.MediaMTXConfig, error)
	UpdateConfig(ctx context.Context, config *config.MediaMTXConfig) error

	// Manager access for cleanup operations
	GetRecordingManager() *RecordingManager
	GetSnapshotManager() *SnapshotManager

	// Lifecycle
	Start(ctx context.Context) error
	Stop(ctx context.Context) error

	// Active recording management (Phase 2 enhancement)
	IsDeviceRecording(devicePath string) bool
	StartActiveRecording(devicePath, sessionID, streamName string) error
	StopActiveRecording(devicePath string) error
	GetActiveRecordings() map[string]*ActiveRecording
	GetActiveRecording(devicePath string) *ActiveRecording
}

// MediaMTXControllerAPI is a restricted interface for higher layers (e.g., WebSocket)
// It exposes only camera-identifier-based methods and hides implementation/mapping details.
type MediaMTXControllerAPI interface {
	// Camera queries
	GetCameraList(ctx context.Context) (*CameraListResponse, error)
	GetCameraStatus(ctx context.Context, device string) (*CameraStatusResponse, error)
	ValidateCameraDevice(ctx context.Context, device string) (bool, error)

	// Health and metrics
	GetHealth(ctx context.Context) (*HealthStatus, error)
	GetMetrics(ctx context.Context) (*Metrics, error)
	GetSystemMetrics(ctx context.Context) (*SystemMetrics, error)
	GetStorageInfo(ctx context.Context) (*StorageInfo, error)
	GetServerInfo(ctx context.Context) (*ServerInfo, error)
	GetHealthMonitor() HealthMonitor

	// Streaming (uses Path from api_types.go)
	GetStreams(ctx context.Context) ([]*Path, error)
	StartStreaming(ctx context.Context, device string) (*Path, error)
	StopStreaming(ctx context.Context, device string) error
	GetStreamURL(ctx context.Context, device string) (string, error)
	GetStreamStatus(ctx context.Context, device string) (*Path, error)

	// Recording and snapshots (identifier based)
	TakeAdvancedSnapshot(ctx context.Context, device string, options map[string]interface{}) (*Snapshot, error)
	StartAdvancedRecording(ctx context.Context, device string, options map[string]interface{}) (*RecordingSession, error)
	StopAdvancedRecording(ctx context.Context, sessionID string) error
	GetSessionIDByDevice(device string) (string, bool)
	GetRecordingInfo(ctx context.Context, filename string) (*FileMetadata, error)
	GetSnapshotInfo(ctx context.Context, filename string) (*FileMetadata, error)
	ListRecordings(ctx context.Context, limit, offset int) (*FileListResponse, error)
	ListSnapshots(ctx context.Context, limit, offset int) (*FileListResponse, error)
	DeleteRecording(ctx context.Context, filename string) error
	DeleteSnapshot(ctx context.Context, filename string) error

	// Cleanup and manager access (for file retention operations)
	GetRecordingManager() *RecordingManager
	GetSnapshotManager() *SnapshotManager

	// External stream discovery
	DiscoverExternalStreams(ctx context.Context, options DiscoveryOptions) (*DiscoveryResult, error)
	AddExternalStream(ctx context.Context, stream *ExternalStream) error
	RemoveExternalStream(ctx context.Context, streamURL string) error
	GetExternalStreams(ctx context.Context) ([]*ExternalStream, error)
}

// Compile-time assertion: controller implements the restricted API
var _ MediaMTXControllerAPI = (*controller)(nil)

// MediaMTXClient interface defines HTTP client operations
type MediaMTXClient interface {
	// HTTP operations
	Get(ctx context.Context, path string) ([]byte, error)
	Post(ctx context.Context, path string, data []byte) ([]byte, error)
	Put(ctx context.Context, path string, data []byte) ([]byte, error)
	Patch(ctx context.Context, path string, data []byte) error
	Delete(ctx context.Context, path string) error

	// Health check
	HealthCheck(ctx context.Context) error

	// Connection management
	Close() error
}

// SystemEventNotifier interface for threshold-crossing notifications
type SystemEventNotifier interface {
	NotifySystemHealth(status string, metrics map[string]interface{})
}

// HealthMonitor interface defines health monitoring operations
type HealthMonitor interface {
	// Health monitoring
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	GetStatus() HealthStatus
	IsHealthy() bool
	GetMetrics() map[string]interface{}

	// Circuit breaker
	IsCircuitOpen() bool
	RecordSuccess()
	RecordFailure()

	// Threshold-crossing notifications
	SetSystemNotifier(notifier SystemEventNotifier)

	// Event-driven health monitoring
	SubscribeToHealthChanges() <-chan struct{}
}

// PathManager interface defines path management operations
type PathManager interface {
	// Path operations
	CreatePath(ctx context.Context, name, source string, options map[string]interface{}) error
	PatchPath(ctx context.Context, name string, config map[string]interface{}) error
	DeletePath(ctx context.Context, name string) error
	GetPath(ctx context.Context, name string) (*Path, error)
	ListPaths(ctx context.Context) ([]*PathConf, error)
	GetRuntimePaths(ctx context.Context) ([]*Path, error)

	// Path validation
	ValidatePath(ctx context.Context, name string) error
	PathExists(ctx context.Context, name string) bool

	// Path readiness
	WaitForPathReady(ctx context.Context, name string, timeout time.Duration) error
	ActivatePathPublisher(ctx context.Context, name string) error

	// Camera operations (PathManager handles camera-path integration)
	GetCameraList(ctx context.Context) (*CameraListResponse, error)
	GetCameraStatus(ctx context.Context, device string) (*CameraStatusResponse, error)
	ValidateCameraDevice(ctx context.Context, device string) (bool, error)

	// Camera-path mapping (abstraction layer)
	GetPathForCamera(cameraID string) (string, bool) // camera0 -> camera0 (MediaMTX path)
	GetCameraForPath(pathName string) (string, bool) // camera0 -> camera0 (camera ID)

	// Device-camera mapping (main abstraction layer)
	GetDevicePathForCamera(cameraID string) (string, bool)   // camera0 -> /dev/video0
	GetCameraForDevicePath(devicePath string) (string, bool) // /dev/video0 -> camera0
}

// StreamManager interface defines stream management operations
type StreamManager interface {
	// Stream operations (simplified - single path for all operations)
	StartStream(ctx context.Context, devicePath string) (*Path, error)

	// Stream lifecycle management
	StopStream(ctx context.Context, device string) error

	// Stream utilities
	GenerateStreamURL(streamName string) string
	GenerateStreamName(devicePath string, useCase StreamUseCase) string

	// Recording control (simplified - operates on stable paths)
	EnableRecording(ctx context.Context, devicePath string, outputPath string) error
	DisableRecording(ctx context.Context, devicePath string) error

	// Stream readiness management
	WaitForStreamReadiness(ctx context.Context, streamName string, timeout time.Duration) (bool, error)

	// Generic stream operations
	CreateStream(ctx context.Context, name, source string) (*Path, error)
	DeleteStream(ctx context.Context, id string) error
	GetStream(ctx context.Context, id string) (*Path, error)
	ListStreams(ctx context.Context) ([]*Path, error)

	// Stream monitoring
	MonitorStream(ctx context.Context, id string) error
	GetStreamStatus(ctx context.Context, id string) (string, error)
}

// StreamUseCase represents different stream use cases
type StreamUseCase string

const (
	UseCaseRecording StreamUseCase = "recording"
	// UseCaseViewing and UseCaseSnapshot removed - single path handles all operations
)

// UseCaseConfig represents configuration for different stream use cases
type UseCaseConfig struct {
	RunOnDemandCloseAfter   string `json:"run_on_demand_close_after"`
	RunOnDemandRestart      bool   `json:"run_on_demand_restart"`
	RunOnDemandStartTimeout string `json:"run_on_demand_start_timeout"`
	Suffix                  string `json:"suffix"`
}

// SessionState represents the state of a recording session
type SessionState string

const (
	SessionStateIdle      SessionState = "IDLE"
	SessionStateRecording SessionState = "RECORDING"
	SessionStateStopped   SessionState = "STOPPED"
	SessionStateError     SessionState = "ERROR"
)

// FFmpegManager interface defines FFmpeg process management for snapshots only
type FFmpegManager interface {
	// Process management
	StartProcess(ctx context.Context, command []string, outputPath string) (int, error)
	StopProcess(ctx context.Context, pid int) error
	IsProcessRunning(ctx context.Context, pid int) bool

	// Command building
	BuildCommand(args ...string) []string

	// Snapshot operations only
	TakeSnapshot(ctx context.Context, device, outputPath string) error

	// File management
	RotateFile(ctx context.Context, oldPath, newPath string) error
	GetFileInfo(ctx context.Context, path string) (int64, time.Time, error)
}

// RTSPConnection represents an RTSP connection from swagger.json
type RTSPConnection struct {
	ID            string    `json:"id"`
	Created       time.Time `json:"created"`
	RemoteAddr    string    `json:"remoteAddr"`
	BytesReceived int64     `json:"bytesReceived"`
	BytesSent     int64     `json:"bytesSent"`
	Session       *string   `json:"session,omitempty"`
}

// RTSPConnectionList represents a list of RTSP connections from swagger.json
type RTSPConnectionList struct {
	PageCount int64             `json:"pageCount"`
	ItemCount int64             `json:"itemCount"`
	Items     []*RTSPConnection `json:"items"`
}

// RTSPConnectionState represents RTSP session state from swagger.json
type RTSPConnectionState string

const (
	RTSPStateIdle    RTSPConnectionState = "idle"
	RTSPStateRead    RTSPConnectionState = "read"
	RTSPStatePublish RTSPConnectionState = "publish"
)

// RTSPConnectionSession represents an RTSP session from swagger.json
type RTSPConnectionSession struct {
	ID                  string              `json:"id"`
	Created             time.Time           `json:"created"`
	RemoteAddr          string              `json:"remoteAddr"`
	State               RTSPConnectionState `json:"state"`
	Path                string              `json:"path"`
	Query               string              `json:"query"`
	Transport           *string             `json:"transport,omitempty"`
	BytesReceived       int64               `json:"bytesReceived"`
	BytesSent           int64               `json:"bytesSent"`
	RTPPacketsReceived  int64               `json:"rtpPacketsReceived"`
	RTPPacketsSent      int64               `json:"rtpPacketsSent"`
	RTPPacketsLost      int64               `json:"rtpPacketsLost"`
	RTPPacketsInError   int64               `json:"rtpPacketsInError"`
	RTPPacketsJitter    float64             `json:"rtpPacketsJitter"`
	RTCPPacketsReceived int64               `json:"rtcpPacketsReceived"`
	RTCPPacketsSent     int64               `json:"rtcpPacketsSent"`
	RTCPPacketsInError  int64               `json:"rtcpPacketsInError"`
}

// RTSPConnectionSessionList represents a list of RTSP sessions from swagger.json
type RTSPConnectionSessionList struct {
	PageCount int64                    `json:"pageCount"`
	ItemCount int64                    `json:"itemCount"`
	Items     []*RTSPConnectionSession `json:"items"`
}

// RTSPConnectionManager interface defines RTSP connection management operations
type RTSPConnectionManager interface {
	// Connection monitoring
	ListConnections(ctx context.Context, page, itemsPerPage int) (*RTSPConnectionList, error)
	GetConnection(ctx context.Context, id string) (*RTSPConnection, error)

	// Session management
	ListSessions(ctx context.Context, page, itemsPerPage int) (*RTSPConnectionSessionList, error)
	GetSession(ctx context.Context, id string) (*RTSPConnectionSession, error)
	KickSession(ctx context.Context, id string) error

	// Health and monitoring
	GetConnectionHealth(ctx context.Context) (*HealthStatus, error)
	GetConnectionMetrics(ctx context.Context) map[string]interface{}
}
