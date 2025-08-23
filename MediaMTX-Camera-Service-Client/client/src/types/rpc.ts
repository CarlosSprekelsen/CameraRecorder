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
 * Aligned with server error codes exactly
 */
export const ERROR_CODES = {
  // Standard JSON-RPC 2.0 error codes
  PARSE_ERROR: -32700,
  INVALID_REQUEST: -32600,
  METHOD_NOT_FOUND: -32601,
  INVALID_PARAMS: -32602,
  INTERNAL_ERROR: -32603,
  
  // Service-specific error codes (aligned with server)
  CAMERA_NOT_FOUND_OR_DISCONNECTED: -32001,
  RECORDING_ALREADY_IN_PROGRESS: -32002,
  MEDIAMTX_SERVICE_UNAVAILABLE: -32003,
  AUTHENTICATION_REQUIRED: -32004,
  INSUFFICIENT_STORAGE_SPACE: -32005,
  CAMERA_CAPABILITY_NOT_SUPPORTED: -32006,
} as const;

/**
 * Type for error code values
 */
export type ErrorCode = typeof ERROR_CODES[keyof typeof ERROR_CODES];

/**
 * Available JSON-RPC methods
 * Aligned with server API methods
 */
export const RPC_METHODS = {
  // Core methods
  PING: 'ping',
  GET_CAMERA_LIST: 'get_camera_list',
  GET_CAMERA_STATUS: 'get_camera_status',
  
  // Camera control methods
  TAKE_SNAPSHOT: 'take_snapshot',
  START_RECORDING: 'start_recording',
  STOP_RECORDING: 'stop_recording',
  
  // File management methods
  LIST_RECORDINGS: 'list_recordings',
  LIST_SNAPSHOTS: 'list_snapshots',
  
  // Stream management methods
  GET_STREAMS: 'get_streams',
} as const;

/**
 * Type for RPC method values
 */
export type RPCMethod = typeof RPC_METHODS[keyof typeof RPC_METHODS];

/**
 * Available notification methods
 * Aligned with server notification methods
 */
export const NOTIFICATION_METHODS = {
  CAMERA_STATUS_UPDATE: 'camera_status_update',
  RECORDING_STATUS_UPDATE: 'recording_status_update',
} as const;

/**
 * Type for notification method values
 */
export type NotificationMethod = typeof NOTIFICATION_METHODS[keyof typeof NOTIFICATION_METHODS];

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
 * Union type for all notification messages
 */
export type NotificationMessage = CameraStatusNotification | RecordingStatusNotification;

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