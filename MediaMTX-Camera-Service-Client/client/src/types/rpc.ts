/**
 * JSON-RPC 2.0 type definitions
 * Aligned with MediaMTX Camera Service API specification
 * Server API Reference: ../mediamtx-camera-service/docs/api/json-rpc-methods.md
 */

/**
 * JSON-RPC 2.0 request
 */
export interface JSONRPCRequest {
  jsonrpc: '2.0';
  method: string;
  params?: Record<string, unknown>;
  id: number;
}

/**
 * JSON-RPC 2.0 error object
 */
export interface JSONRPCError {
  code: number;
  message: string;
  data?: unknown;
}

/**
 * JSON-RPC 2.0 response
 */
export interface JSONRPCResponse {
  jsonrpc: '2.0';
  result?: unknown;
  error?: JSONRPCError;
  id: number;
}

/**
 * JSON-RPC 2.0 notification (no id field)
 */
export interface JSONRPCNotification {
  jsonrpc: '2.0';
  method: string;
  params?: Record<string, unknown>;
}

/**
 * WebSocket message type (response or notification)
 */
export type WebSocketMessage = JSONRPCResponse | JSONRPCNotification;

/**
 * Common error codes used by the MediaMTX Camera Service
 * Aligned with Go server error codes exactly
 */
export const ERROR_CODES = {
  // Standard JSON-RPC 2.0 error codes
  PARSE_ERROR: -32700,
  INVALID_REQUEST: -32600,
  METHOD_NOT_FOUND: -32601,
  INVALID_PARAMS: -32602,
  INTERNAL_ERROR: -32603,
  
  // Go server error codes (aligned with server API documentation)
  AUTHENTICATION_FAILED: -32001,          // "Authentication failed or token expired"
  PERMISSION_DENIED: -32002,              // "Permission denied"
  NOT_FOUND: -32010,                      // "Not found"
  INVALID_STATE: -32020,                  // "Invalid state"
  UNSUPPORTED: -32030,                    // "Unsupported"
  RATE_LIMITED: -32040,                   // "Rate limited"
  DEPENDENCY_FAILED: -32050,              // "Dependency failed"
} as const;

/**
 * Type for error code values
 */
export type ErrorCode = typeof ERROR_CODES[keyof typeof ERROR_CODES];

/**
 * Available JSON-RPC methods
 * Aligned with Go server API methods
 */
export const RPC_METHODS = {
  // Authentication
  AUTHENTICATE: 'authenticate',
  
  // Core methods
  PING: 'ping',
  GET_CAMERA_LIST: 'get_camera_list',
  GET_CAMERA_STATUS: 'get_camera_status',
  GET_CAMERA_CAPABILITIES: 'get_camera_capabilities',
  
  // Camera control methods
  TAKE_SNAPSHOT: 'take_snapshot',
  START_RECORDING: 'start_recording',
  STOP_RECORDING: 'stop_recording',
  
  // Streaming methods
  START_STREAMING: 'start_streaming',
  STOP_STREAMING: 'stop_streaming',
  GET_STREAM_URL: 'get_stream_url',
  GET_STREAM_STATUS: 'get_stream_status',
  GET_STREAMS: 'get_streams',
  
  // File management methods
  LIST_RECORDINGS: 'list_recordings',
  LIST_SNAPSHOTS: 'list_snapshots',
  GET_RECORDING_INFO: 'get_recording_info',
  GET_SNAPSHOT_INFO: 'get_snapshot_info',
  DELETE_RECORDING: 'delete_recording',
  DELETE_SNAPSHOT: 'delete_snapshot',
  
  // Storage management methods
  GET_STORAGE_INFO: 'get_storage_info',
  SET_RETENTION_POLICY: 'set_retention_policy',
  CLEANUP_OLD_FILES: 'cleanup_old_files',
  
  // System information methods
  GET_STATUS: 'get_status',
  GET_SERVER_INFO: 'get_server_info',
  GET_METRICS: 'get_metrics',
  
  // Event subscription methods
  SUBSCRIBE_EVENTS: 'subscribe_events',
  UNSUBSCRIBE_EVENTS: 'unsubscribe_events',
  GET_SUBSCRIPTION_STATS: 'get_subscription_stats',
  
  // External stream discovery methods
  DISCOVER_EXTERNAL_STREAMS: 'discover_external_streams',
  ADD_EXTERNAL_STREAM: 'add_external_stream',
  REMOVE_EXTERNAL_STREAM: 'remove_external_stream',
  GET_EXTERNAL_STREAMS: 'get_external_streams',
  SET_DISCOVERY_INTERVAL: 'set_discovery_interval',
} as const;

/**
 * Type for RPC method values
 */
export type RPCMethod = typeof RPC_METHODS[keyof typeof RPC_METHODS];

/**
 * Available notification methods
 * Aligned with Go server notification methods
 */
export const NOTIFICATION_METHODS = {
  CAMERA_STATUS_UPDATE: 'camera_status_update',
  RECORDING_STATUS_UPDATE: 'recording_status_update',
  STORAGE_STATUS_UPDATE: 'storage_status_update',
} as const;

/**
 * Available event topics for subscription
 * Aligned with Go server event system
 */
export const EVENT_TOPICS = {
  CAMERA_CONNECTED: 'camera.connected',
  CAMERA_DISCONNECTED: 'camera.disconnected',
  CAMERA_STATUS_CHANGE: 'camera.status_change',
  RECORDING_START: 'recording.start',
  RECORDING_STOP: 'recording.stop',
  RECORDING_ERROR: 'recording.error',
  SNAPSHOT_TAKEN: 'snapshot.taken',
  SYSTEM_HEALTH: 'system.health',
  SYSTEM_STARTUP: 'system.startup',
  SYSTEM_SHUTDOWN: 'system.shutdown',
} as const;

/**
 * Type for notification method values
 */
export type NotificationMethod = typeof NOTIFICATION_METHODS[keyof typeof NOTIFICATION_METHODS];

/**
 * Type for event topic values
 */
export type EventTopic = typeof EVENT_TOPICS[keyof typeof EVENT_TOPICS];

/**
 * Authentication request parameters
 */
export interface AuthenticationParams {
  auth_token: string; // JWT token or API key
}

/**
 * Authentication response
 */
export interface AuthenticationResponse {
  authenticated: boolean;
  role: 'admin' | 'operator' | 'viewer';
  permissions: string[];
  expires_at: string; // ISO 8601 timestamp
  session_id: string;
}

/**
 * Event subscription parameters
 */
export interface EventSubscriptionParams {
  topics: string[];
  filters?: Record<string, unknown>;
}

/**
 * Event subscription response
 */
export interface EventSubscriptionResponse {
  subscribed: boolean;
  topics: string[];
  filters?: Record<string, unknown>;
}

/**
 * External stream discovery parameters
 */
export interface ExternalStreamDiscoveryParams {
  skydio_enabled?: boolean;
  generic_enabled?: boolean;
  force_rescan?: boolean;
  include_offline?: boolean;
}

/**
 * External stream information
 */
export interface ExternalStream {
  url: string;
  type: string;
  name: string;
  status: string;
  discovered_at: string;
  last_seen: string;
  capabilities: {
    protocol: string;
    format: string;
    source: string;
    stream_type: string;
    port: number;
    stream_path: string;
    codec: string;
    metadata: string;
  };
}

/**
 * External stream discovery response
 */
export interface ExternalStreamDiscoveryResponse {
  discovered_streams: ExternalStream[];
  skydio_streams: ExternalStream[];
  generic_streams: ExternalStream[];
  scan_timestamp: number;
  total_found: number;
  discovery_options: ExternalStreamDiscoveryParams;
  scan_duration: string;
  errors: string[];
}

/**
 * Camera status update notification
 */
export interface CameraStatusNotification {
  jsonrpc: '2.0';
  method: 'camera_status_update';
  params: {
    device: string;
    status: string;
    name: string;
    resolution: string;
    fps: number;
    streams: {
      rtsp: string;
      webrtc: string;
      hls: string;
    };
  };
}

/**
 * Recording status update notification
 */
export interface RecordingStatusNotification {
  jsonrpc: '2.0';
  method: 'recording_status_update';
  params: {
    device: string;
    status: string;
    filename: string;
    duration: number;
  };
}

/**
 * Storage status update notification (NEW)
 */
export interface StorageStatusNotification {
  jsonrpc: '2.0';
  method: 'storage_status_update';
  params: {
    total_space: number;
    used_space: number;
    available_space: number;
    usage_percent: number;
    threshold_status: 'normal' | 'warning' | 'critical';
  };
}

/**
 * Union type for all notification messages
 */
export type NotificationMessage = CameraStatusNotification | RecordingStatusNotification | StorageStatusNotification;

/**
 * WebSocket configuration
 */
export interface WebSocketConfig {
  url: string;
  reconnectInterval: number;
  maxReconnectAttempts: number;
  requestTimeout: number;
  heartbeatInterval: number;
  baseDelay: number;
  maxDelay: number;
}

/**
 * RPC call options
 */
export interface RPCCallOptions {
  timeout?: number;
  retries?: number;
}

/**
 * Type guard to check if a message is a notification
 */
export function isNotification(message: WebSocketMessage): message is JSONRPCNotification {
  return 'method' in message && !('id' in message);
}

/**
 * Type guard to check if a message is a response
 */
export function isResponse(message: WebSocketMessage): message is JSONRPCResponse {
  return 'id' in message;
}

/**
 * Type guard to check if a response is an error response
 */
export function isErrorResponse(response: JSONRPCResponse): response is JSONRPCResponse & { error: JSONRPCError } {
  return 'error' in response && response.error !== undefined;
}

/**
 * Performance targets (from server documentation)
 */
export const PERFORMANCE_TARGETS = {
  STATUS_METHODS: 50, // <50ms response time
  CONTROL_METHODS: 100, // <100ms response time
  WEBSOCKET_NOTIFICATIONS: 20, // <20ms delivery latency
  CLIENT_INITIAL_LOAD: 3000, // <3s initial load
  CLIENT_WEBSOCKET_CONNECTION: 1000, // <1s connection time
  CLIENT_BUNDLE_SIZE: 2 * 1024 * 1024, // <2MB bundle size
  CLIENT_MEMORY_USAGE: 50 * 1024 * 1024, // <50MB memory usage
} as const;