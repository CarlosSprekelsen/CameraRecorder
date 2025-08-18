/**
 * UI-related type definitions
 * Component props, state interfaces, and UI-specific types
 */

import type { CameraDevice, RecordingResponse, SnapshotResponse } from './camera';
import type { JSONRPCError } from './rpc';

/**
 * Application view modes
 */
export type ViewMode = 'dashboard' | 'detail' | 'settings';

/**
 * Theme mode
 */
export type ThemeMode = 'light' | 'dark' | 'auto';

/**
 * Loading states for async operations
 */
export interface LoadingState {
  isLoading: boolean;
  message?: string;
}

/**
 * Error state for UI components
 */
export interface ErrorState {
  hasError: boolean;
  error?: JSONRPCError | Error | string;
  timestamp?: Date;
}

/**
 * Connection status for WebSocket
 */
export type ConnectionStatus = 'connected' | 'connecting' | 'disconnected' | 'error';

/**
 * Global UI state
 */
export interface UIState {
  selectedCamera: string | null;
  viewMode: ViewMode;
  theme: ThemeMode;
  sidebarOpen: boolean;
  notifications: NotificationState[];
  connection: ConnectionStatus;
  loading: LoadingState;
  error: ErrorState;
}

/**
 * Notification types
 */
export type NotificationType = 'info' | 'success' | 'warning' | 'error';

/**
 * Notification state
 */
export interface NotificationState {
  id: string;
  type: NotificationType;
  title: string;
  message?: string;
  timestamp: Date;
  autoClose?: boolean;
  duration?: number;
}

/**
 * Camera card component props
 */
export interface CameraCardProps {
  camera: CameraDevice;
  selected?: boolean;
  onSelect?: (device: string) => void;
  onSnapshot?: (device: string) => void;
  onRecord?: (device: string) => void;
  onViewStreams?: (device: string) => void;
}

/**
 * Camera detail view props
 */
export interface CameraDetailProps {
  device: string;
  camera?: CameraDevice;
  onBack?: () => void;
  onRefresh?: () => void;
}

/**
 * Camera list component props
 */
export interface CameraListProps {
  cameras: CameraDevice[];
  selectedDevice?: string;
  loading?: boolean;
  error?: string;
  onSelectCamera?: (device: string) => void;
  onRefresh?: () => void;
}

/**
 * Dashboard component props
 */
export interface DashboardProps {
  cameras: CameraDevice[];
  selectedDevice?: string;
  onSelectCamera?: (device: string) => void;
  onRefreshCameras?: () => void;
}

/**
 * App shell/layout props
 */
export interface AppShellProps {
  children: React.ReactNode;
  title?: string;
  showBackButton?: boolean;
  onBack?: () => void;
  actions?: React.ReactNode;
}

/**
 * Connection status indicator props
 */
export interface ConnectionStatusProps {
  status: ConnectionStatus;
  onReconnect?: () => void;
  compact?: boolean;
}

/**
 * Error boundary state
 */
export interface ErrorBoundaryState {
  hasError: boolean;
  error?: Error;
  errorInfo?: React.ErrorInfo;
}

/**
 * Camera operations (for hooks and state management)
 */
export interface CameraOperations {
  takeSnapshot: (device: string, options?: { format?: string; quality?: number }) => Promise<SnapshotResponse>;
  startRecording: (device: string, options?: { duration?: number; format?: string }) => Promise<RecordingResponse>;
  stopRecording: (device: string) => Promise<RecordingResponse>;
  refreshCamera: (device: string) => Promise<void>;
  refreshAllCameras: () => Promise<void>;
}

/**
 * Form field types
 */
export interface FormField<T = string> {
  value: T;
  error?: string;
  touched: boolean;
  valid: boolean;
}

/**
 * Recording settings form
 */
export interface RecordingSettings {
  duration: FormField<number>;
  format: FormField<string>;
  autoStop: FormField<boolean>;
}

/**
 * Snapshot settings form
 */
export interface SnapshotSettings {
  format: FormField<string>;
  quality: FormField<number>;
  filename: FormField<string>;
}

/**
 * Settings form state
 */
export interface SettingsFormState {
  recording: RecordingSettings;
  snapshot: SnapshotSettings;
  notifications: FormField<boolean>;
  autoRefresh: FormField<boolean>;
  refreshInterval: FormField<number>;
}

/**
 * Table column definition
 */
export interface TableColumn<T = unknown> {
  key: string;
  label: string;
  sortable?: boolean;
  render?: (value: unknown, row: T) => React.ReactNode;
  width?: string | number;
  align?: 'left' | 'center' | 'right';
}

/**
 * Table props
 */
export interface TableProps<T = unknown> {
  data: T[];
  columns: TableColumn<T>[];
  loading?: boolean;
  error?: string;
  emptyMessage?: string;
  onRowClick?: (row: T) => void;
  selectable?: boolean;
  selectedRows?: Set<string>;
  onSelectionChange?: (selectedRows: Set<string>) => void;
}

/**
 * Modal component props
 */
export interface ModalProps {
  open: boolean;
  onClose: () => void;
  title?: string;
  children: React.ReactNode;
  maxWidth?: 'xs' | 'sm' | 'md' | 'lg' | 'xl';
  fullWidth?: boolean;
  disableBackdropClick?: boolean;
}

/**
 * Button component props
 */
export interface ButtonProps {
  variant?: 'contained' | 'outlined' | 'text';
  color?: 'primary' | 'secondary' | 'error' | 'warning' | 'info' | 'success';
  size?: 'small' | 'medium' | 'large';
  disabled?: boolean;
  loading?: boolean;
  startIcon?: React.ReactNode;
  endIcon?: React.ReactNode;
  onClick?: () => void;
  children: React.ReactNode;
}

/**
 * Dialog action result
 */
export interface DialogResult<T = unknown> {
  action: 'confirm' | 'cancel' | 'close';
  data?: T;
}

/**
 * Confirmation dialog props
 */
export interface ConfirmDialogProps {
  open: boolean;
  title: string;
  message: string;
  onResult: (result: DialogResult) => void;
  confirmText?: string;
  cancelText?: string;
  severity?: 'info' | 'warning' | 'error';
}