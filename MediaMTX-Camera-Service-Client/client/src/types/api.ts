// Generated types based on OpenRPC specification
// MediaMTX Camera Service API Types - OFFICIAL RPC DOCUMENTATION COMPLIANCE

// ============================================================================
// CORE TYPES FROM OFFICIAL RPC DOCUMENTATION
// ============================================================================

// Device ID pattern from official spec
export type DeviceId = string; // Pattern: ^camera[0-9]+$

// ISO Timestamp format
export type IsoTimestamp = string; // ISO 8601 format

// Pagination from official spec
export interface Pagination {
  limit?: number; // default: 50, max: 1000
  offset?: number; // default: 0
}

// Streams object from official spec
export interface Streams {
  rtsp: string; // e.g., "rtsp://<host>:8554/camera0"
  hls: string; // e.g., "https://<host>/hls/camera0.m3u8"
}

// Camera object from official spec
export interface Camera {
  device: DeviceId;
  status: 'CONNECTED' | 'DISCONNECTED' | 'ERROR';
  name?: string;
  resolution?: string;
  fps?: number;
  streams?: Streams;
}

// Camera List Result from official spec
export interface CameraListResult {
  cameras: Camera[];
  total: number;
  connected: number;
}

// Camera Status Result from official spec
export interface CameraStatusResult {
  device: DeviceId;
  status: 'CONNECTED' | 'DISCONNECTED' | 'ERROR';
  name?: string;
  resolution?: string;
  fps?: number;
  streams?: Streams;
  metrics?: {
    bytes_sent: number;
    readers: number;
    uptime: number;
  };
  capabilities?: {
    formats: string[];
    resolutions: string[];
  };
}

// Snapshot Result from official spec - FIXED: Use API spec values
export interface SnapshotResult {
  device: DeviceId;
  filename: string;
  file_size: number;
  status: 'SUCCESS' | 'FAILED';  // FIXED: API spec uses SUCCESS/FAILED
  timestamp: IsoTimestamp;
  file_path: string;
}

// Recording Start Result from official spec - FIXED: Use API spec values
export interface RecordingStartResult {
  device: DeviceId;
  status: 'SUCCESS' | 'FAILED';  // FIXED: API spec uses SUCCESS/FAILED for operations
  filename?: string;
  file_path?: string;
  duration?: number;
  format?: string;
}

// Recording Stop Result from official spec - FIXED: Use API spec values
export interface RecordingStopResult {
  device: DeviceId;
  status: 'SUCCESS' | 'FAILED';  // FIXED: API spec uses SUCCESS/FAILED for operations
  filename?: string;
  file_path?: string;
  duration?: number;
}

// Camera Capabilities Result from official spec
export interface CameraCapabilitiesResult {
  device: DeviceId;
  formats: string[];
  resolutions: string[];
  fps_options: number[];
  validation_status: 'NONE' | 'DISCONNECTED' | 'CONFIRMED';
}

// REMOVED: Duplicate RecordingStartResult and RecordingStopResult definitions

// REMOVED: Duplicate SnapshotResult definition

// Snapshot Info from official spec
export interface SnapshotInfo {
  filename: string;
  file_size: number;
  created_time: IsoTimestamp;
  download_url: string;
}

// Recording Info from official spec
export interface RecordingInfo {
  filename: string;
  file_size: number;
  duration: number;
  created_time: IsoTimestamp;
  download_url: string;
}

// File List Result from official spec
export interface FileListResult {
  files: Array<{
    filename: string;
    file_size: number;
    modified_time: IsoTimestamp;
    download_url: string;
  }>;
  total: number;
  limit: number;
  offset: number;
}

// Stream Start Result from official spec
export interface StreamStartResult {
  device: DeviceId;
  stream_name: string;
  stream_url: string;
  status: 'STARTED' | 'FAILED';
  start_time: IsoTimestamp;
  auto_close_after: string;
  ffmpeg_command: string;
}

// Stream Stop Result from official spec
export interface StreamStopResult {
  device: DeviceId;
  stream_name: string;
  status: 'STOPPED' | 'FAILED';
  start_time: IsoTimestamp;
  end_time: IsoTimestamp;
  duration: number;
  stream_continues: boolean;
  message?: string;
}

// Stream URL Result from official spec
export interface StreamUrlResult {
  device: DeviceId;
  stream_name: string;
  stream_url: string;
  available: boolean;
  active_consumers: number;
  stream_status: 'READY' | 'NOT_READY' | 'ERROR';
}

// Stream Status Result from official spec
export interface StreamStatusResult {
  device: DeviceId;
  stream_name: string;
  status: 'ACTIVE' | 'INACTIVE' | 'ERROR' | 'STARTING' | 'STOPPING';
  ready: boolean;
  ffmpeg_process: {
    running: boolean;
    pid: number;
    uptime: number;
  };
  mediamtx_path: {
    exists: boolean;
    ready: boolean;
    readers: number;
  };
  metrics: {
    bytes_sent: number;
    frames_sent: number;
    bitrate: number;
    fps: number;
  };
  start_time: IsoTimestamp;
}

// Streams List Result from official spec
export interface StreamsListResult {
  name: string;
  source: string;
  ready: boolean;
  readers: number;
  bytes_sent: number;
}

// Authentication Parameters
export interface AuthenticateParams {
  auth_token: string;
}

// Authentication Result from official spec
export interface AuthenticateResult {
  authenticated: boolean;
  role: 'admin' | 'operator' | 'viewer';
  permissions: string[];
  expires_at: IsoTimestamp;
  session_id: string;
}

// Server Info from official spec
export interface ServerInfo {
  name: string;
  version: string;
  build_date: string;
  go_version: string;
  architecture: string;
  capabilities: string[];
  supported_formats: string[];
  max_cameras: number;
}

// System Status from official spec (get_status - admin only)
export interface SystemStatus {
  status: 'HEALTHY' | 'DEGRADED' | 'UNHEALTHY';
  uptime: number;
  version: string;
  components: {
    websocket_server: 'RUNNING' | 'STOPPED' | 'ERROR' | 'STARTING' | 'STOPPING';
    camera_monitor: 'RUNNING' | 'STOPPED' | 'ERROR' | 'STARTING' | 'STOPPING';
    mediamtx: 'RUNNING' | 'STOPPED' | 'ERROR' | 'STARTING' | 'STOPPING';
  };
}

// System Readiness Status from official spec (get_system_status - viewer accessible)
export interface SystemReadinessStatus {
  status: 'ready' | 'partial' | 'starting';
  message: string;
  available_cameras: string[];
  discovery_active: boolean;
}

// Storage Info from official spec
export interface StorageInfo {
  total_space: number;
  used_space: number;
  available_space: number;
  usage_percentage: number;
  recordings_size: number;
  snapshots_size: number;
  low_space_warning: boolean;
}

// Metrics Result from official spec
export interface MetricsResult {
  timestamp: IsoTimestamp;
  system_metrics: {
    cpu_usage: number;
    memory_usage: number;
    disk_usage: number;
    goroutines: number;
  };
  camera_metrics: {
    connected_cameras: number;
    cameras: Record<
      string,
      {
        path: string;
        name: string;
        status: string;
        device_num: number;
        last_seen: IsoTimestamp;
        capabilities: Record<string, unknown>;
        formats: Array<{
          pixel_format: string;
          width: number;
          height: number;
          frame_rates: string[];
        }>;
      }
    >;
  };
  recording_metrics: Record<string, Record<string, unknown>>;
  stream_metrics: {
    active_streams: number;
    total_streams: number;
    total_viewers: number;
  };
}

// Subscription Result from official spec
export interface SubscriptionResult {
  subscribed: boolean;
  topics: string[];
  filters?: Record<string, unknown>;
}

// Unsubscription Result from official spec
export interface UnsubscriptionResult {
  unsubscribed: boolean;
  topics: string[];
}

// Subscription Stats Result from official spec
export interface SubscriptionStatsResult {
  global_stats: {
    total_subscriptions: number;
    active_clients: number;
    topic_counts: Record<string, number>;
  };
  client_topics: string[];
  client_id: string;
}

// External Stream Discovery Result from official spec
export interface ExternalStreamDiscoveryResult {
  discovered_streams: Array<{
    url: string;
    type: string;
    name: string;
    status: 'DISCOVERED' | 'ERROR';
    discovered_at: IsoTimestamp;
    last_seen: IsoTimestamp;
    capabilities: Record<string, unknown>;
  }>;
  skydio_streams: Array<{
    url: string;
    type: string;
    name: string;
    status: 'DISCOVERED' | 'ERROR';
    discovered_at: IsoTimestamp;
    last_seen: IsoTimestamp;
    capabilities: Record<string, unknown>;
  }>;
  generic_streams: Array<{
    url: string;
    type: string;
    name: string;
    status: 'DISCOVERED' | 'ERROR';
    discovered_at: IsoTimestamp;
    last_seen: IsoTimestamp;
    capabilities: Record<string, unknown>;
  }>;
  scan_timestamp: IsoTimestamp;
  total_found: number;
  discovery_options: {
    skydio_enabled: boolean;
    generic_enabled: boolean;
    force_rescan: boolean;
    include_offline: boolean;
  };
  scan_duration: string;
  errors: string[];
}

// External Stream Add Result from official spec
export interface ExternalStreamAddResult {
  stream_url: string;
  stream_name: string;
  stream_type: string;
  status: 'ADDED' | 'ERROR';
  timestamp: IsoTimestamp;
}

// External Stream Remove Result from official spec
export interface ExternalStreamRemoveResult {
  stream_url: string;
  status: 'REMOVED' | 'ERROR';
  timestamp: IsoTimestamp;
}

// External Streams List Result from official spec
export interface ExternalStreamsListResult {
  external_streams: Array<{
    url: string;
    type: string;
    name: string;
    status: 'DISCOVERED' | 'ERROR';
    discovered_at: IsoTimestamp;
    last_seen: IsoTimestamp;
    capabilities: Record<string, unknown>;
  }>;
  skydio_streams: Array<{
    url: string;
    type: string;
    name: string;
    status: 'DISCOVERED' | 'ERROR';
    discovered_at: IsoTimestamp;
    last_seen: IsoTimestamp;
    capabilities: Record<string, unknown>;
  }>;
  generic_streams: Array<{
    url: string;
    type: string;
    name: string;
    status: 'DISCOVERED' | 'ERROR';
    discovered_at: IsoTimestamp;
    last_seen: IsoTimestamp;
    capabilities: Record<string, unknown>;
  }>;
  total_count: number;
  timestamp: IsoTimestamp;
}

// Discovery Interval Set Result from official spec
export interface DiscoveryIntervalSetResult {
  scan_interval: number;
  status: 'UPDATED' | 'ERROR';
  message: string;
  timestamp: IsoTimestamp;
}

// Retention Policy Set Result from official spec
export interface RetentionPolicySetResult {
  policy_type: 'age' | 'size' | 'manual';
  max_age_days?: number;
  max_size_gb?: number;
  enabled: boolean;
  message: string;
}

// Cleanup Result from official spec
export interface CleanupResult {
  cleanup_executed: boolean;
  files_deleted: number;
  space_freed: number;
  message: string;
}

// Delete Result from official spec
export interface DeleteResult {
  filename: string;
  deleted: boolean;
  message: string;
}

// ============================================================================
// JSON-RPC 2.0 TYPES
// ============================================================================

export interface JsonRpcRequest {
  jsonrpc: '2.0';
  method: string;
  params?: Record<string, unknown>;
  id: string | number;
}

export interface JsonRpcResponse<T = unknown> {
  jsonrpc: '2.0';
  result?: T;
  error?: {
    code: number;
    message: string;
    data?: unknown;
  };
  id: string | number;
}

export interface JsonRpcNotification {
  jsonrpc: '2.0';
  method: string;
  params?: Record<string, unknown>;
}

// ============================================================================
// CLIENT STATE TYPES (NOT FROM API)
// ============================================================================

// WebSocket Connection Types
export type ConnectionStatus = 'connected' | 'connecting' | 'disconnected' | 'error';

export interface ConnectionState {
  status: ConnectionStatus;
  lastError: string | null;
  reconnectAttempts: number;
  lastConnected: string | null;
}

// Authentication State
export interface AuthState {
  token: string | null;
  role: 'admin' | 'operator' | 'viewer' | null;
  session_id: string | null;
  isAuthenticated: boolean;
  expires_at: string | null;
  permissions: string[];
}

// Server State
export interface ServerState {
  info: ServerInfo | null;
  status: SystemStatus | null;
  systemReadiness: SystemReadinessStatus | null;
  storage: StorageInfo | null;
  loading: boolean;
  error: string | null;
  lastUpdated: string | null;
}

// ============================================================================
// RPC METHOD NAMES (ALIGNED WITH OFFICIAL API)
// ============================================================================

export type RpcMethod =
  | 'ping'
  | 'authenticate'
  | 'get_server_info'
  | 'get_status'
  | 'get_system_status'  // FIXED: Added missing method
  | 'get_storage_info'
  | 'get_metrics'
  | 'get_camera_list'
  | 'get_camera_status'
  | 'get_camera_capabilities'
  | 'get_stream_url'
  | 'get_stream_status'
  | 'get_streams'
  | 'start_streaming'
  | 'stop_streaming'
  | 'take_snapshot'
  | 'start_recording'
  | 'stop_recording'
  | 'list_recordings'
  | 'list_snapshots'
  | 'get_recording_info'
  | 'get_snapshot_info'
  | 'delete_recording'
  | 'delete_snapshot'
  | 'subscribe_events'
  | 'unsubscribe_events'
  | 'get_subscription_stats'
  | 'discover_external_streams'
  | 'add_external_stream'
  | 'remove_external_stream'
  | 'get_external_streams'
  | 'set_discovery_interval'
  | 'set_retention_policy'
  | 'cleanup_old_files';

// ============================================================================
// ERROR CODES (FROM OFFICIAL API)
// ============================================================================

export const ERROR_CODES = {
  INVALID_REQUEST: -32600,
  METHOD_NOT_FOUND: -32601,
  INVALID_PARAMS: -32602,
  INTERNAL_ERROR: -32603,
  AUTH_FAILED: -32001,
  PERMISSION_DENIED: -32002,
  NOT_FOUND: -32010,
  INVALID_STATE: -32020,
  UNSUPPORTED: -32030,
  RATE_LIMITED: -32040,
  DEPENDENCY_FAILED: -32050,
} as const;
