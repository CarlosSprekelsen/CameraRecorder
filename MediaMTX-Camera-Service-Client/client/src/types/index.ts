/**
 * Central export point for all type definitions
 * Import types using: import { CameraDevice, JSONRPCRequest } from '@/types'
 */

// Camera types
export type {
  CameraStatus,
  ValidationStatus,
  VideoFormat,
  RecordingFormat,
  SnapshotFormat,
  RecordingStatus,
  CameraCapabilities,
  CameraStreams,
  CameraMetrics,
  CameraDevice,
  CameraListResponse,
  RecordingRequest,
  RecordingResponse,
  SnapshotRequest,
  SnapshotResponse,
  ServerInfo,
  CameraStatusUpdateParams,
  RecordingStatusUpdateParams,
} from './camera';

// RPC types
export type {
  JSONRPCRequest,
  JSONRPCError,
  JSONRPCResponse,
  JSONRPCNotification,
  WebSocketMessage,
  ErrorCode,
  RPCMethod,
  NotificationMethod,
  CameraStatusNotification,
  RecordingStatusNotification,
  NotificationMessage,
  WebSocketConfig,
  RPCCallOptions,
} from './rpc';

// Export RPC constants
export {
  ERROR_CODES,
  RPC_METHODS,
  NOTIFICATION_METHODS,
  isNotification,
  isResponse,
  isErrorResponse,
} from './rpc';

// UI types
export type {
  ViewMode,
  ThemeMode,
  LoadingState,
  ErrorState,
  ConnectionStatus,
  UIState,
  NotificationType,
  NotificationState,
  CameraCardProps,
  CameraDetailProps,
  CameraListProps,
  DashboardProps,
  AppShellProps,
  ConnectionStatusProps,
  ErrorBoundaryState,
  CameraOperations,
  FormField,
  RecordingSettings,
  SnapshotSettings,
  SettingsFormState,
  TableColumn,
  TableProps,
  ModalProps,
  ButtonProps,
  DialogResult,
  ConfirmDialogProps,
} from './ui';