/**
 * JSON-RPC 2.0 type definitions
 * Based on JSON-RPC 2.0 specification and MediaMTX Camera Service API
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
 */
export const ERROR_CODES = {
  // Camera-specific errors
  CAMERA_NOT_FOUND: -1000,
  CAMERA_ALREADY_CONNECTED: -1001,
  CAMERA_CONNECTION_FAILED: -1002,
  MEDIAMTX_ERROR: -1003,
  
  // JSON-RPC standard errors
  PARSE_ERROR: -32700,
  INVALID_REQUEST: -32600,
  METHOD_NOT_FOUND: -32601,
  INVALID_PARAMS: -32602,
  INTERNAL_ERROR: -32603,
} as const;

/**
 * Type for error code values
 */
export type ErrorCode = typeof ERROR_CODES[keyof typeof ERROR_CODES];

/**
 * Available JSON-RPC methods
 */
export const RPC_METHODS = {
  // Camera operations
  GET_CAMERA_LIST: 'get_camera_list',
  GET_CAMERA_STATUS: 'get_camera_status',
  
  // Recording operations
  START_RECORDING: 'start_recording',
  STOP_RECORDING: 'stop_recording',
  
  // Snapshot operations
  TAKE_SNAPSHOT: 'take_snapshot',
  
  // Utility operations
  PING: 'ping',
  GET_SERVER_INFO: 'get_server_info',
} as const;

/**
 * Type for RPC method values
 */
export type RPCMethod = typeof RPC_METHODS[keyof typeof RPC_METHODS];

/**
 * Available notification methods
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
export interface CameraStatusNotification extends JSONRPCNotification {
  method: typeof NOTIFICATION_METHODS.CAMERA_STATUS_UPDATE;
  params: {
    device: string;
    status: string;
    capabilities?: Record<string, unknown>;
    streams?: Record<string, unknown>;
  };
}

/**
 * Recording status update notification
 */
export interface RecordingStatusNotification extends JSONRPCNotification {
  method: typeof NOTIFICATION_METHODS.RECORDING_STATUS_UPDATE;
  params: {
    device: string;
    session_id: string;
    status: string;
    progress?: number;
    duration?: number;
  };
}

/**
 * Union type for all notification types
 */
export type NotificationMessage = CameraStatusNotification | RecordingStatusNotification;

/**
 * WebSocket configuration for RPC connection
 */
export interface WebSocketConfig {
  url: string;
  maxReconnectAttempts: number;
  baseDelay: number;
  maxDelay: number;
  requestTimeout: number;
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
  return !('id' in message);
}

/**
 * Type guard to check if a message is a response
 */
export function isResponse(message: WebSocketMessage): message is JSONRPCResponse {
  return 'id' in message;
}

/**
 * Type guard to check if a response contains an error
 */
export function isErrorResponse(response: JSONRPCResponse): boolean {
  return 'error' in response && response.error !== undefined;
}