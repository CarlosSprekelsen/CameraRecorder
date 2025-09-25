/**
 * SINGLE validation utility for all API responses
 * Validates against documented schemas only
 * 
 * Ground Truth References:
 * - API Documentation: ../mediamtx-camera-service-go/docs/api/mediamtx_camera_service_openrpc.json
 * 
 * Requirements Coverage:
 * - REQ-VAL-001: API response validation
 * - REQ-VAL-002: Error code validation
 * - REQ-VAL-003: Schema compliance
 * 
 * Test Categories: Unit/Integration/API-Compliance
 * API Documentation Reference: mediamtx_camera_service_openrpc.json
 */

import { 
  Camera, 
  CameraListResult, 
  RecordingStart, 
  RecordingStop, 
  SnapshotInfo,
  ListFilesResult,
  StreamStatus,
  MetricsResult,
  StatusResult,
  ServerInfo,
  StorageInfo,
  AuthResult
} from '@/types/api';

export class APIResponseValidator {
  /**
   * Validate camera list result against documented schema
   * MANDATORY: Use this validation for all camera list tests
   */
  static validateCameraListResult(result: any): result is CameraListResult {
    return (
      typeof result === 'object' &&
      result !== null &&
      Array.isArray(result.cameras) &&
      typeof result.total === 'number' &&
      typeof result.connected === 'number' &&
      result.cameras.every((camera: any) => this.validateCamera(camera))
    );
  }

  /**
   * Validate camera object against documented schema
   * MANDATORY: Use this validation for all camera tests
   */
  static validateCamera(camera: any): camera is Camera {
    return (
      typeof camera === 'object' &&
      camera !== null &&
      typeof camera.device === 'string' &&
      /^camera[0-9]+$/.test(camera.device) &&
      typeof camera.status === 'string' &&
      ['CONNECTED', 'DISCONNECTED', 'ERROR'].includes(camera.status)
    );
  }

  /**
   * Validate recording start result against documented schema
   * MANDATORY: Use this validation for all recording start tests
   */
  static validateRecordingStartResult(result: any): result is RecordingStart {
    return (
      typeof result === 'object' &&
      result !== null &&
      typeof result.device === 'string' &&
      /^camera[0-9]+$/.test(result.device) &&
      typeof result.status === 'string' &&
      ['RECORDING', 'STARTING', 'STOPPING', 'PAUSED', 'ERROR', 'FAILED'].includes(result.status) &&
      typeof result.start_time === 'string' &&
      this.validateIsoTimestamp(result.start_time)
    );
  }

  /**
   * Validate recording stop result against documented schema
   * MANDATORY: Use this validation for all recording stop tests
   */
  static validateRecordingStopResult(result: any): result is RecordingStop {
    return (
      typeof result === 'object' &&
      result !== null &&
      typeof result.device === 'string' &&
      /^camera[0-9]+$/.test(result.device) &&
      typeof result.status === 'string' &&
      ['STOPPED', 'FAILED'].includes(result.status) &&
      typeof result.end_time === 'string' &&
      this.validateIsoTimestamp(result.end_time)
    );
  }

  /**
   * Validate snapshot info against documented schema
   * MANDATORY: Use this validation for all snapshot tests
   */
  static validateSnapshotInfo(result: any): result is SnapshotInfo {
    return (
      typeof result === 'object' &&
      result !== null &&
      typeof result.device === 'string' &&
      /^camera[0-9]+$/.test(result.device) &&
      typeof result.filename === 'string' &&
      typeof result.status === 'string' &&
      ['SUCCESS', 'FAILED'].includes(result.status) &&
      typeof result.timestamp === 'string' &&
      this.validateIsoTimestamp(result.timestamp)
    );
  }

  /**
   * Validate file list result against documented schema
   * MANDATORY: Use this validation for all file list tests
   */
  static validateListFilesResult(result: any): result is ListFilesResult {
    return (
      typeof result === 'object' &&
      result !== null &&
      Array.isArray(result.files) &&
      typeof result.total === 'number' &&
      typeof result.limit === 'number' &&
      typeof result.offset === 'number' &&
      result.files.every((file: any) => this.validateRecordingFile(file))
    );
  }

  /**
   * Validate recording file against documented schema
   * MANDATORY: Use this validation for all file tests
   */
  static validateRecordingFile(file: any): boolean {
    return (
      typeof file === 'object' &&
      file !== null &&
      typeof file.filename === 'string'
    );
  }

  /**
   * Validate stream status against documented schema
   * MANDATORY: Use this validation for all stream tests
   */
  static validateStreamStatus(result: any): result is StreamStatus {
    return (
      typeof result === 'object' &&
      result !== null &&
      typeof result.device === 'string' &&
      /^camera[0-9]+$/.test(result.device) &&
      typeof result.status === 'string' &&
      ['ACTIVE', 'INACTIVE', 'ERROR', 'STARTING', 'STOPPING'].includes(result.status)
    );
  }

  /**
   * Validate metrics result against documented schema
   * MANDATORY: Use this validation for all metrics tests
   */
  static validateMetricsResult(result: any): result is MetricsResult {
    return (
      typeof result === 'object' &&
      result !== null &&
      typeof result.timestamp === 'string' &&
      this.validateIsoTimestamp(result.timestamp) &&
      typeof result.system_metrics === 'object' &&
      typeof result.camera_metrics === 'object' &&
      typeof result.recording_metrics === 'object' &&
      typeof result.stream_metrics === 'object'
    );
  }

  /**
   * Validate status result against documented schema
   * MANDATORY: Use this validation for all status tests
   */
  static validateStatusResult(result: any): result is StatusResult {
    return (
      typeof result === 'object' &&
      result !== null &&
      typeof result.status === 'string' &&
      ['HEALTHY', 'DEGRADED', 'UNHEALTHY'].includes(result.status)
    );
  }

  /**
   * Validate server info against documented schema
   * MANDATORY: Use this validation for all server info tests
   */
  static validateServerInfo(result: any): result is ServerInfo {
    return (
      typeof result === 'object' &&
      result !== null &&
      typeof result.name === 'string' &&
      typeof result.version === 'string' &&
      typeof result.build_date === 'string' &&
      typeof result.go_version === 'string' &&
      typeof result.architecture === 'string' &&
      Array.isArray(result.capabilities) &&
      Array.isArray(result.supported_formats) &&
      typeof result.max_cameras === 'number'
    );
  }

  /**
   * Validate storage info against documented schema
   * MANDATORY: Use this validation for all storage tests
   */
  static validateStorageInfo(result: any): result is StorageInfo {
    return (
      typeof result === 'object' &&
      result !== null &&
      typeof result.total_space === 'number' &&
      typeof result.used_space === 'number' &&
      typeof result.available_space === 'number' &&
      typeof result.usage_percentage === 'number' &&
      typeof result.recordings_size === 'number' &&
      typeof result.snapshots_size === 'number' &&
      typeof result.low_space_warning === 'boolean'
    );
  }

  /**
   * Validate authentication result against documented schema
   * MANDATORY: Use this validation for all auth tests
   */
  static validateAuthResult(result: any): result is AuthResult {
    return (
      typeof result === 'object' &&
      result !== null &&
      typeof result.authenticated === 'boolean' &&
      typeof result.role === 'string' &&
      ['admin', 'operator', 'viewer'].includes(result.role) &&
      Array.isArray(result.permissions) &&
      typeof result.session_id === 'string'
    );
  }

  /**
   * Validate error response against documented error codes
   * MANDATORY: Use this validation for all error tests
   */
  static validateErrorResponse(error: any, expectedCode: number): boolean {
    return (
      typeof error === 'object' &&
      error !== null &&
      typeof error.code === 'number' &&
      error.code === expectedCode &&
      typeof error.message === 'string'
    );
  }

  /**
   * Validate ISO timestamp format
   * MANDATORY: Use this validation for all timestamp tests
   */
  static validateIsoTimestamp(timestamp: string): boolean {
    try {
      const date = new Date(timestamp);
      return !isNaN(date.getTime()) && timestamp.includes('T') && timestamp.includes('Z');
    } catch {
      return false;
    }
  }

  /**
   * Validate device ID format
   * MANDATORY: Use this validation for all device ID tests
   */
  static validateDeviceId(deviceId: string): boolean {
    return /^camera[0-9]+$/.test(deviceId);
  }

  /**
   * Validate pagination parameters
   * MANDATORY: Use this validation for all pagination tests
   */
  static validatePaginationParams(limit: number, offset: number): boolean {
    return (
      typeof limit === 'number' &&
      limit >= 0 &&
      limit <= 1000 &&
      typeof offset === 'number' &&
      offset >= 0
    );
  }

  /**
   * Validate recording format
   * MANDATORY: Use this validation for all format tests
   */
  static validateRecordingFormat(format: string): boolean {
    return ['fmp4', 'mp4', 'mkv'].includes(format);
  }

  /**
   * Validate stream URL format
   * MANDATORY: Use this validation for all stream URL tests
   */
  static validateStreamUrl(url: string): boolean {
    try {
      const parsedUrl = new URL(url);
      return ['http:', 'https:', 'rtsp:'].includes(parsedUrl.protocol);
    } catch {
      return false;
    }
  }
}
