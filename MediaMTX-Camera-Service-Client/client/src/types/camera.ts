/**
 * Camera-related type definitions
 * Aligned with MediaMTX Camera Service API specification
 * Server API Reference: ../mediamtx-camera-service/docs/api/json-rpc-methods.md
 */

/**
 * Camera device information
 * Aligned with server get_camera_list and get_camera_status responses
 */
export interface CameraDevice {
  device: string;
  status: CameraStatus;
  name: string;
  resolution: string;
  fps: number;
  streams: CameraStreams;
  metrics?: CameraPerformanceMetrics;
  capabilities?: CameraCapabilities;
  // Enhanced recording status (NEW)
  recording?: boolean;
  recording_session?: string;
  current_file?: string;
  elapsed_time?: number;
}

/**
 * Camera status values
 * Aligned with server API status values
 */
export type CameraStatus = 'CONNECTED' | 'DISCONNECTED' | 'ERROR';

/**
 * Camera streams configuration
 * Aligned with server API streams object
 */
export interface CameraStreams {
  rtsp: string;
  webrtc: string;
  hls: string;
}

/**
 * Camera metrics information
 * Aligned with server API metrics object
 */
export interface CameraPerformanceMetrics {
  bytes_sent: number;
  readers: number;
  uptime: number;
}

/**
 * Camera capabilities information
 * Aligned with server API capabilities object
 */
export interface CameraCapabilities {
  formats: string[];
  resolutions: string[];
}

/**
 * Stream information from MediaMTX
 * Aligned with server get_streams response
 */
export interface StreamInfo {
  name: string;
  source: string;
  ready: boolean;
  readers: number;
  bytes_sent: number;
}

/**
 * Stream list response
 * Aligned with server get_streams response
 */
export interface StreamListResponse {
  streams: StreamInfo[];
  total: number;
  active: number;
}

/**
 * Camera list response
 * Aligned with server get_camera_list response
 */
export interface CameraListResponse {
  cameras: CameraDevice[];
  total: number;
  connected: number;
}

/**
 * Camera capabilities validation status
 */
export type ValidationStatus = 'provisional' | 'confirmed';

/**
 * Supported video formats
 */
export type VideoFormat = 'YUYV' | 'MJPEG' | 'H264';

/**
 * Supported recording formats
 */
export type RecordingFormat = 'mp4' | 'mkv';

/**
 * Storage information (NEW)
 * Aligned with server get_storage_info response
 */
export interface StorageInfo {
  total_space: number;
  used_space: number;
  available_space: number;
  usage_percent: number;
  threshold_status: ThresholdStatus;
}

/**
 * Storage threshold status (NEW)
 * Aligned with server API threshold_status object
 */
export interface ThresholdStatus {
  warning_threshold: number;
  critical_threshold: number;
  current_status: 'normal' | 'warning' | 'critical';
}

/**
 * Storage usage information (NEW)
 */
export interface StorageUsage {
  total_space: number;
  used_space: number;
  available_space: number;
  usage_percent: number;
}

/**
 * Storage validation result (NEW)
 */
export interface StorageValidationResult {
  isValid: boolean;
  canRecord: boolean;
  canSnapshot: boolean;
  warnings: string[];
  errors: string[];
}

/**
 * Supported snapshot formats
 */
export type SnapshotFormat = 'jpg' | 'png';

/**
 * Recording status
 */
export type RecordingStatus = 'STARTED' | 'RECORDING' | 'STOPPED' | 'ERROR';

/**
 * Enhanced recording status (NEW)
 * Includes conflict and progress states
 */
export type EnhancedRecordingStatus = RecordingStatus | 'CONFLICT' | 'PAUSED' | 'ROTATING';



/**
 * Recording progress information (NEW)
 * Enhanced with comprehensive progress tracking (F1.4.3)
 */
export interface RecordingProgress {
  device: string;
  session_id: string;
  current_file: string;
  elapsed_time: number;
  file_size: number;
  rotation_count: number;
  is_continuous: boolean;
  // Enhanced progress information (F1.4.3)
  file_path?: string;
  recording_quality?: string;
  bitrate?: number;
  frame_rate?: number;
  resolution?: string;
  file_rotation_timestamp?: number;
}

/**
 * Recording session information
 * Aligned with server start_recording and stop_recording responses
 */
export interface RecordingSession {
  device: string;
  session_id: string;
  filename: string;
  status: RecordingStatus;
  start_time: string;
  end_time?: string;
  duration?: number;
  format: RecordingFormat;
  file_size?: number;
}

/**
 * Recording start parameters
 * Aligned with server start_recording method parameters
 */
export interface StartRecordingParams {
  device: string;
  duration_seconds?: number; // 1-3600 seconds
  duration_minutes?: number; // 1-1440 minutes  
  duration_hours?: number; // 1-24 hours
  format?: RecordingFormat;
}

/**
 * Recording stop parameters
 */
export interface StopRecordingParams {
  device: string;
}

/**
 * Snapshot capture result
 * Aligned with server take_snapshot response
 */
export interface SnapshotResult {
  device: string;
  filename: string;
  status: 'completed' | 'FAILED';
  timestamp: string;
  file_size: number;
  file_path: string;
  error?: string; // Present when status is 'FAILED'
}

/**
 * Snapshot capture parameters
 * Aligned with server take_snapshot method parameters
 */
export interface TakeSnapshotParams {
  device: string;
  format?: SnapshotFormat; // 'jpg' or 'png', defaults to 'jpg'
  quality?: number; // 1-100, defaults to 85
  filename?: string; // Optional custom filename
}

/**
 * File information for recordings and snapshots
 * Aligned with server list_recordings and list_snapshots responses
 */
export interface FileInfo {
  filename: string;
  file_size: number;
  modified_time: string;
  download_url: string;
}

/**
 * Configuration management types (NEW)
 */
export interface RecordingConfig {
  rotation_minutes: number;
  default_format: RecordingFormat;
  auto_rotation: boolean;
  maxFilesPerCamera: number;
  autoDelete: boolean;
}

export interface StorageConfig {
  warn_percent: number;
  block_percent: number;
  critical_percent: number;
  monitoring_enabled: boolean;
  maxUsagePercent: number;
}

export interface AppConfig {
  recording: RecordingConfig;
  storage: StorageConfig;
  environment: EnvironmentConfig;
  connection: ConnectionConfig;
  system: SystemConfig;
}

export interface EnvironmentConfig {
  RECORDING_ROTATION_MINUTES?: string;
  STORAGE_WARN_PERCENT?: string;
  STORAGE_BLOCK_PERCENT?: string;
}

export interface ConnectionConfig {
  websocketUrl: string;
  healthUrl: string;
  timeout: number;
}

export interface SystemConfig {
  logLevel: 'debug' | 'info' | 'warn' | 'error';
  autoRefresh: boolean;
  refreshInterval: number;
}

/**
 * Configuration validation result (NEW)
 */
export interface ConfigValidationResult {
  isValid: boolean;
  errors: string[];
  warnings: string[];
  config: AppConfig;
}

/**
 * File list response with pagination
 */
export interface FileListResponse {
  files: FileInfo[];
  total: number;
  limit: number;
  offset: number;
}

/**
 * File list parameters
 */
export interface FileListParams {
  limit?: number;
  offset?: number;
}

/**
 * Authentication parameters
 * Aligned with server authenticate method
 */
export interface AuthenticateParams {
  token: string;
}

/**
 * Authentication response
 * Aligned with server authenticate response
 */
export interface AuthenticateResponse {
  authenticated: boolean;
  role?: string;
  permissions?: string[];
  expires_at?: string;
  session_id?: string;
  user_id?: string;
  auth_method?: string;
}

/**
 * File type for file operations
 */
export type FileType = 'recordings' | 'snapshots';

/**
 * File item for file management
 * Aligned with server file list responses
 */
export interface FileItem {
  filename: string;
  file_size: number;
  created_time: string;
  modified_time: string;
  download_url: string;
  duration?: number; // Only for recordings
  format?: string; // File format extension
}

/**
 * Camera status update notification
 * Aligned with server camera_status_update notification
 */
export interface CameraStatusUpdateParams {
  device: string;
  status: CameraStatus;
  name: string;
  resolution: string;
  fps: number;
  streams: CameraStreams;
}

/**
 * Recording status update notification
 * Aligned with server recording_status_update notification
 */
export interface RecordingStatusUpdateParams {
  device: string;
  status: RecordingStatus;
  filename: string;
  duration: number;
}