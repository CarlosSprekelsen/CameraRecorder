/**
 * Camera-related type definitions
 * Aligned with MediaMTX Camera Service API specification
 * Server API Reference: ../mediamtx-camera-service/docs/api/json-rpc-methods.md
 */

/**
 * Camera connection status
 */
export type CameraStatus = 'CONNECTED' | 'DISCONNECTED' | 'ERROR';

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
 * Supported snapshot formats
 */
export type SnapshotFormat = 'jpg' | 'png';

/**
 * Recording status
 */
export type RecordingStatus = 'STARTED' | 'STOPPED' | 'ERROR';

/**
 * Camera device capabilities
 */
export interface CameraCapabilities {
  formats: VideoFormat[];
  resolutions: string[];
}

/**
 * Camera streaming endpoints
 */
export interface CameraStreams {
  rtsp: string;
  webrtc: string;
  hls: string;
}

/**
 * Camera device metrics
 */
export interface CameraMetrics {
  bytes_sent: number;
  readers: number;
  uptime: number;
}

/**
 * Core camera device information
 * Aligned with server get_camera_list and get_camera_status responses
 */
export interface CameraDevice {
  device: string;
  status: CameraStatus;
  name: string;
  resolution: string;
  fps: number;
  streams: CameraStreams;
  metrics?: CameraMetrics;
  capabilities?: CameraCapabilities;
}

/**
 * Camera list response from get_camera_list
 */
export interface CameraListResponse {
  cameras: CameraDevice[];
  total: number;
  connected: number;
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
 */
export interface StartRecordingParams {
  device: string;
  duration?: number; // Optional - omit for unlimited recording
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
  status: 'completed' | 'error';
  timestamp: string;
  file_size: number;
  file_path: string;
}

/**
 * Snapshot capture parameters
 */
export interface TakeSnapshotParams {
  device: string;
  filename?: string;
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

// Legacy types for backward compatibility
// These will be deprecated in favor of the aligned types above

/**
 * @deprecated Use RecordingSession instead
 */
export interface RecordingRequest {
  device: string;
  duration?: number;
  format?: RecordingFormat;
}

/**
 * @deprecated Use RecordingSession instead
 */
export interface RecordingResponse {
  success: boolean;
  session_id: string;
  file_path: string;
  duration?: number;
  format: RecordingFormat;
}

/**
 * @deprecated Use SnapshotResult instead
 */
export interface SnapshotRequest {
  device: string;
  format?: SnapshotFormat;
  quality?: number;
  filename?: string;
}

/**
 * @deprecated Use SnapshotResult instead
 */
export interface SnapshotResponse {
  success: boolean;
  file_path: string;
  format: SnapshotFormat;
  quality: number;
  size: number;
}

/**
 * @deprecated Server info not implemented in current API
 */
export interface ServerInfo {
  version: string;
  uptime: number;
  cameras_connected: number;
  total_recordings: number;
  total_snapshots: number;
}