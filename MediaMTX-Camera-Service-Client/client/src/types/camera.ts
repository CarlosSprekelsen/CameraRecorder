/**
 * Camera-related type definitions
 * Based on MediaMTX Camera Service API specification
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
export type VideoFormat = 'YUYV' | 'MJPEG';

/**
 * Supported recording formats
 */
export type RecordingFormat = 'mp4' | 'avi' | 'mkv';

/**
 * Supported snapshot formats
 */
export type SnapshotFormat = 'jpg' | 'png';

/**
 * Recording status
 */
export type RecordingStatus = 'RECORDING' | 'STOPPED' | 'ERROR';

/**
 * Camera device capabilities
 */
export interface CameraCapabilities {
  resolution: string;
  fps: number;
  validation_status: ValidationStatus;
  formats: VideoFormat[];
  all_resolutions?: string[];
}

/**
 * Camera streaming endpoints
 */
export interface CameraStreams {
  rtsp?: string;
  webrtc?: string;
  hls?: string;
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
 */
export interface CameraDevice {
  device: string;
  name: string;
  status: CameraStatus;
  capabilities?: CameraCapabilities;
  streams?: CameraStreams;
  metrics?: CameraMetrics;
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
 * Recording operation request
 */
export interface RecordingRequest {
  device: string;
  duration?: number;
  format?: RecordingFormat;
}

/**
 * Recording operation response
 */
export interface RecordingResponse {
  success: boolean;
  session_id: string;
  file_path: string;
  duration?: number;
  format: RecordingFormat;
}

/**
 * Snapshot operation request
 */
export interface SnapshotRequest {
  device: string;
  format?: SnapshotFormat;
  quality?: number;
  filename?: string;
}

/**
 * Snapshot operation response
 */
export interface SnapshotResponse {
  success: boolean;
  file_path: string;
  format: SnapshotFormat;
  quality: number;
  size: number;
}

/**
 * Server information response
 */
export interface ServerInfo {
  version: string;
  uptime: number;
  cameras_connected: number;
  total_recordings: number;
  total_snapshots: number;
}

/**
 * Camera status update notification parameters
 */
export interface CameraStatusUpdateParams {
  device: string;
  status: CameraStatus;
  capabilities?: CameraCapabilities;
  streams?: CameraStreams;
}

/**
 * Recording status update notification parameters
 */
export interface RecordingStatusUpdateParams {
  device: string;
  session_id: string;
  status: RecordingStatus;
  progress?: number;
  duration?: number;
}