// Generated types based on OpenRPC specification
// MediaMTX Camera Service API Types

export interface AuthenticateParams {
  auth_token: string;
}

export interface AuthenticateResult {
  authenticated: boolean;
  role: 'admin' | 'operator' | 'viewer';
  permissions: string[];
  expires_at: string;
  session_id: string;
}

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

export interface SystemStatus {
  status: 'HEALTHY' | 'DEGRADED' | 'UNHEALTHY';
  uptime: number;
  version: string;
  components: {
    websocket_server: string;
    camera_monitor: string;
    mediamtx: string;
  };
}

export interface StorageInfo {
  total_space: number;
  used_space: number;
  available_space: number;
  usage_percentage: number;
  recordings_size: number;
  snapshots_size: number;
  low_space_warning: boolean;
}

// JSON-RPC 2.0 Types
export interface JsonRpcRequest {
  jsonrpc: '2.0';
  method: string;
  params?: any;
  id: string | number;
}

export interface JsonRpcResponse<T = any> {
  jsonrpc: '2.0';
  result?: T;
  error?: {
    code: number;
    message: string;
    data?: any;
  };
  id: string | number;
}

export interface JsonRpcNotification {
  jsonrpc: '2.0';
  method: string;
  params?: any;
}

// WebSocket Connection Types
export type ConnectionStatus = 'connected' | 'connecting' | 'disconnected' | 'error';

export interface ConnectionState {
  status: ConnectionStatus;
  lastError: string | null;
  reconnectAttempts: number;
  lastConnected: string | null;
}

// Authentication Types
export interface AuthState {
  token: string | null;
  role: 'admin' | 'operator' | 'viewer' | null;
  session_id: string | null;
  isAuthenticated: boolean;
  expires_at: string | null;
  permissions: string[];
}

// Server State Types
export interface ServerState {
  info: ServerInfo | null;
  status: SystemStatus | null;
  storage: StorageInfo | null;
  loading: boolean;
  error: string | null;
  lastUpdated: string | null;
}

// RPC Method Names (for type safety) - ALIGNED WITH SERVER API
export type RpcMethod = 
  | 'ping'
  | 'authenticate'
  | 'get_server_info'
  | 'get_status'
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

// Error Codes (from OpenRPC spec)
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
