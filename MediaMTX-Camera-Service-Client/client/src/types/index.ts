/**
 * Central export point for all types
 * Provides unified access to all type definitions
 */

// Camera types
export type {
  CameraDevice,
  CameraStatus,
  CameraStreams,
  CameraMetrics,
  CameraCapabilities,
  StreamInfo,
  StreamListResponse,
} from './camera';

// Recording types
export type {
  RecordingSession,
  RecordingStatus,
  StartRecordingParams,
  StopRecordingParams,
  RecordingFormat,
} from './camera';

// Snapshot types
export type {
  SnapshotResult,
  TakeSnapshotParams,
  SnapshotFormat,
} from './camera';

// File types
export type {
  FileItem,
  FileType,
  FileListResponse,
  FileListParams,
} from './camera';

// Server types
export type {
  ServerInfo,
  CameraListResponse,
} from './camera';

// Settings types
export type {
  AppSettings,
  ConnectionSettings,
  RecordingSettings,
  SnapshotSettings,
  UISettings,
  NotificationSettings,
  SecuritySettings,
  PerformanceSettings,
  SettingsValidation,
  SettingsChangeEvent,
  SettingsCategory,
} from './settings';

export { DEFAULT_SETTINGS, SETTINGS_CATEGORIES } from './settings';

// RPC types
export type {
  JSONRPCRequest,
  JSONRPCResponse,
  JSONRPCNotification,
  JSONRPCError,
  WebSocketMessage,
  ErrorCode,
  RPCMethod,
} from './rpc';

export { RPC_METHODS, ERROR_CODES, NOTIFICATION_METHODS } from './rpc';

// UI types
export type {
  ViewMode,
  ConnectionStatus,
  RecordingSettings as UIRecordingSettings,
  SnapshotSettings as UISnapshotSettings,
  SettingsFormState,
} from './ui';