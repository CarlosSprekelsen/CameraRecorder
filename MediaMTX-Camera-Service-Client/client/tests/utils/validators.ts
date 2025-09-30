// API Response Validators for MediaMTX Camera Service Client tests
// COMPLIANT WITH OFFICIAL RPC DOCUMENTATION

import {
  // Official RPC API Types
  CameraListResult,
  CameraStatusResult,
  CameraCapabilitiesResult,
  RecordingStartResult,
  RecordingStopResult,
  SnapshotResult,
  SnapshotInfo,
  RecordingInfo,
  FileListResult,
  StreamStartResult,
  StreamStopResult,
  StreamUrlResult,
  StreamStatusResult,
  StreamsListResult,
  AuthenticateResult,
  ServerInfo,
  SystemStatus,
  StorageInfo,
  MetricsResult,
  SubscriptionResult,
  UnsubscriptionResult,
  SubscriptionStatsResult,
  ExternalStreamDiscoveryResult,
  ExternalStreamAddResult,
  ExternalStreamRemoveResult,
  ExternalStreamsListResult,
  DiscoveryIntervalSetResult,
  RetentionPolicySetResult,
  CleanupResult,
  DeleteResult,
  DeviceId,
  IsoTimestamp,
  Streams
} from '../../src/types/api';

/**
 * API Response Validator - validates responses against official RPC documentation schemas
 * GROUND TRUTH COMPLIANCE: All validations match official RPC specification exactly
 */
export class APIResponseValidator {
  // ============================================================================
  // PARAMETER VALIDATION METHODS
  // ============================================================================

  /**
   * Validates camera device ID format according to server documentation
   * Pattern: camera[0-9]+ (e.g., camera0, camera1, camera10)
   * 
   * @deprecated Use validateCameraDeviceId from '../../src/utils/validation' instead
   */
  static validateCameraDeviceId(deviceId: string): boolean {
    // Import the centralized validation function
    const { validateCameraDeviceId } = require('../../src/utils/validation');
    return validateCameraDeviceId(deviceId);
  }

  /**
   * Validates parameter structure for API calls
   * Ensures parameters are objects, not arrays
   */
  static validateParameterStructure(params: any): boolean {
    if (params === null || params === undefined) {
      return true; // Optional parameters are valid
    }
    
    // Parameters must be objects, not arrays
    if (Array.isArray(params)) {
      return false;
    }
    
    return typeof params === 'object';
  }
  // ============================================================================
  // CORE VALIDATION METHODS
  // ============================================================================

  /**
   * Validate Device ID format - matches official RPC pattern: ^camera[0-9]+$
   */
  static validateDeviceId(deviceId: string): boolean {
    const pattern = /^camera[0-9]+$/;
    return pattern.test(deviceId);
  }

  /**
   * Validate ISO Timestamp format - matches official RPC ISO 8601 requirement
   */
  static validateIsoTimestamp(timestamp: string): boolean {
    try {
      const date = new Date(timestamp);
      if (isNaN(date.getTime())) {
        return false;
      }
      
      // Accept both UTC and timezone-aware formats
      const iso = date.toISOString();
      const original = timestamp;
      
      // Check if it's valid ISO format (with or without timezone)
      return iso.substring(0, 19) === original.substring(0, 19) ||
             /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}([.]\d{3})?([+-]\d{2}:\d{2}|Z)$/.test(original);
    } catch {
      return false;
    }
  }

  /**
   * Validate Camera List Result - matches official RPC spec exactly
   */
  static validateCameraListResult(result: unknown): result is CameraListResult {
    if (typeof result !== 'object' || result === null) return false;
    
    const obj = result as Record<string, unknown>;
    
    // Validate cameras array
    if (!Array.isArray(obj.cameras)) return false;
    for (const camera of obj.cameras) {
      if (!this.validateCamera(camera)) return false;
    }
    
    // Validate metadata
    return (
      typeof obj.total === 'number' &&
      typeof obj.connected === 'number' &&
      obj.total >= 0 &&
      obj.connected >= 0 &&
      obj.connected <= obj.total
    );
  }

  /**
   * Validate Camera object - matches official RPC spec exactly
   */
  static validateCamera(camera: unknown): boolean {
    if (typeof camera !== 'object' || camera === null) return false;
    
    const obj = camera as Record<string, unknown>;
    
    // Required fields
    if (!this.validateDeviceId(obj.device as string)) return false;
    if (!['CONNECTED', 'DISCONNECTED', 'ERROR'].includes(obj.status as string)) return false;
    
    // Optional fields
    if (obj.name !== undefined && typeof obj.name !== 'string') return false;
    if (obj.resolution !== undefined && typeof obj.resolution !== 'string') return false;
    if (obj.fps !== undefined && typeof obj.fps !== 'number') return false;
    if (obj.streams !== undefined && !this.validateStreams(obj.streams)) return false;
    
    return true;
  }

  /**
   * Validate Streams object - matches official RPC spec exactly
   * Accepts empty object {} when no streams are active
   */
  static validateStreams(streams: unknown): boolean {
    if (typeof streams !== 'object' || streams === null) return false;
    
    const obj = streams as Record<string, unknown>;
    
    // Empty streams object is valid (no active streams)
    if (Object.keys(obj).length === 0) return true;
    
    // If streams object has content, validate the structure
    return (
      typeof obj.rtsp === 'string' &&
      typeof obj.hls === 'string' &&
      obj.rtsp.startsWith('rtsp://') &&
      obj.hls.startsWith('http://')
    );
  }

  /**
   * Validate Camera Status Result - matches official RPC spec exactly
   */
  static validateCameraStatusResult(result: unknown): result is CameraStatusResult {
    if (typeof result !== 'object' || result === null) return false;
    
    const obj = result as Record<string, unknown>;
    
    // Validate base camera fields
    if (!this.validateCamera(obj)) return false;
    
    // Validate optional metrics
    if (obj.metrics !== undefined) {
      if (!this.validateMetrics(obj.metrics)) return false;
    }
    
    // Validate optional capabilities
    if (obj.capabilities !== undefined) {
      if (!this.validateCapabilities(obj.capabilities)) return false;
    }
    
    return true;
  }

  /**
   * Validate Metrics object - matches official RPC spec exactly
   */
  static validateMetrics(metrics: unknown): boolean {
    if (typeof metrics !== 'object' || metrics === null) return false;
    
    const obj = metrics as Record<string, unknown>;
    
    return (
      typeof obj.bytes_sent === 'number' &&
      typeof obj.readers === 'number' &&
      typeof obj.uptime === 'number' &&
      obj.bytes_sent >= 0 &&
      obj.readers >= 0 &&
      obj.uptime >= 0
    );
  }

  /**
   * Validate Capabilities object - matches official RPC spec exactly
   */
  static validateCapabilities(capabilities: unknown): boolean {
    if (typeof capabilities !== 'object' || capabilities === null) return false;
    
    const obj = capabilities as Record<string, unknown>;
    
    return (
      Array.isArray(obj.formats) &&
      Array.isArray(obj.resolutions) &&
      obj.formats.every((format: unknown) => typeof format === 'string') &&
      obj.resolutions.every((resolution: unknown) => typeof resolution === 'string')
    );
  }

  /**
   * Validate Camera Capabilities Result - matches official RPC spec exactly
   */
  static validateCameraCapabilitiesResult(result: unknown): result is CameraCapabilitiesResult {
    if (typeof result !== 'object' || result === null) return false;
    
    const obj = result as Record<string, unknown>;
    
    return (
      this.validateDeviceId(obj.device as string) &&
      Array.isArray(obj.formats) &&
      Array.isArray(obj.resolutions) &&
      Array.isArray(obj.fps_options) &&
      ['NONE', 'DISCONNECTED', 'CONFIRMED'].includes(obj.validation_status as string) &&
      obj.formats.every((format: unknown) => typeof format === 'string') &&
      obj.resolutions.every((resolution: unknown) => typeof resolution === 'string') &&
      obj.fps_options.every((fps: unknown) => typeof fps === 'number' && fps > 0)
    );
  }

  /**
   * Validate Recording Start Result - matches official RPC spec exactly
   */
  static validateRecordingStartResult(result: unknown): result is RecordingStartResult {
    if (typeof result !== 'object' || result === null) return false;
    
    const obj = result as Record<string, unknown>;
    
    return (
      this.validateDeviceId(obj.device as string) &&
      typeof obj.filename === 'string' &&
      ['RECORDING', 'STARTING', 'STOPPING', 'PAUSED', 'ERROR', 'FAILED'].includes(obj.status as string) &&  // FIXED: API spec for start_recording
      this.validateIsoTimestamp(obj.start_time as string) &&
      ['fmp4', 'mp4', 'mkv'].includes(obj.format as string)
    );
  }

  /**
   * Validate Recording Stop Result - matches official RPC spec exactly
   */
  static validateRecordingStopResult(result: unknown): result is RecordingStopResult {
    if (typeof result !== 'object' || result === null) return false;
    
    const obj = result as Record<string, unknown>;
    
    return (
      this.validateDeviceId(obj.device as string) &&
      typeof obj.filename === 'string' &&
      ['STOPPED', 'STARTING', 'STOPPING', 'PAUSED', 'ERROR', 'FAILED'].includes(obj.status as string) &&  // FIXED: API spec for stop_recording
      this.validateIsoTimestamp(obj.start_time as string) &&
      this.validateIsoTimestamp(obj.end_time as string) &&
      typeof obj.duration === 'number' &&
      typeof obj.file_size === 'number' &&
      ['fmp4', 'mp4', 'mkv'].includes(obj.format as string) &&
      obj.duration >= 0 &&
      obj.file_size >= 0
    );
  }

  /**
   * Validate Snapshot Result - matches official RPC spec exactly
   */
  static validateSnapshotResult(result: unknown): result is SnapshotResult {
    if (typeof result !== 'object' || result === null) return false;
    
    const obj = result as Record<string, unknown>;
    
    return (
      this.validateDeviceId(obj.device as string) &&
      typeof obj.filename === 'string' &&
      ['SUCCESS', 'FAILED'].includes(obj.status as string) &&
      this.validateIsoTimestamp(obj.timestamp as string) &&
      typeof obj.file_size === 'number' &&
      typeof obj.file_path === 'string' &&
      obj.file_size >= 0
    );
  }

  /**
   * Validate Snapshot Info - matches actual server response structure
   */
  static validateSnapshotInfo(result: unknown): result is SnapshotInfo {
    if (typeof result !== 'object' || result === null) return false;
    
    const obj = result as Record<string, unknown>;
    
    return (
      typeof obj.device === 'string' &&
      typeof obj.filename === 'string' &&
      typeof obj.file_size === 'number' &&
      typeof obj.status === 'string' &&
      typeof obj.timestamp === 'string' &&
      typeof obj.file_path === 'string' &&
      this.validateIsoTimestamp(obj.timestamp as string) &&
      obj.file_size >= 0 &&
      ['SUCCESS', 'FAILED'].includes(obj.status as string)
    );
  }

  /**
   * Validate Recording Info - matches official RPC spec exactly
   */
  static validateRecordingInfo(result: unknown): result is RecordingInfo {
    if (typeof result !== 'object' || result === null) return false;
    
    const obj = result as Record<string, unknown>;
    
    return (
      typeof obj.filename === 'string' &&
      typeof obj.file_size === 'number' &&
      typeof obj.duration === 'number' &&
      this.validateIsoTimestamp(obj.created_time as string) &&
      typeof obj.download_url === 'string' &&
      obj.file_size >= 0 &&
      obj.duration >= 0
    );
  }

  /**
   * Validate File List Result - matches official RPC spec exactly
   */
  static validateFileListResult(result: unknown): result is FileListResult {
    if (typeof result !== 'object' || result === null) return false;
    
    const obj = result as Record<string, unknown>;
    
    if (!Array.isArray(obj.files)) return false;
    
    // Validate each file entry
    for (const file of obj.files) {
      if (typeof file !== 'object' || file === null) return false;
      const fileObj = file as Record<string, unknown>;
      if (
        typeof fileObj.filename !== 'string' ||
        typeof fileObj.file_size !== 'number' ||
        !this.validateIsoTimestamp(fileObj.modified_time as string) ||
        typeof fileObj.download_url !== 'string' ||
        fileObj.file_size < 0
      ) {
        return false;
      }
    }
    
    return (
      typeof obj.total === 'number' &&
      typeof obj.limit === 'number' &&
      typeof obj.offset === 'number' &&
      obj.total >= 0 &&
      obj.limit > 0 &&
      obj.limit <= 1000 &&
      obj.offset >= 0
    );
  }

  /**
   * Validate Stream Start Result - matches official RPC spec exactly
   */
  static validateStreamStartResult(result: unknown): result is StreamStartResult {
    if (typeof result !== 'object' || result === null) return false;
    
    const obj = result as Record<string, unknown>;
    
    return (
      this.validateDeviceId(obj.device as string) &&
      typeof obj.stream_name === 'string' &&
      typeof obj.stream_url === 'string' &&
      ['STARTED', 'FAILED'].includes(obj.status as string) &&
      this.validateIsoTimestamp(obj.start_time as string) &&
      typeof obj.auto_close_after === 'string' &&
      typeof obj.ffmpeg_command === 'string'
    );
  }

  /**
   * Validate Stream Stop Result - matches official RPC spec exactly
   */
  static validateStreamStopResult(result: unknown): result is StreamStopResult {
    if (typeof result !== 'object' || result === null) return false;
    
    const obj = result as Record<string, unknown>;
    
    return (
      this.validateDeviceId(obj.device as string) &&
      typeof obj.stream_name === 'string' &&
      ['STOPPED', 'FAILED'].includes(obj.status as string) &&
      this.validateIsoTimestamp(obj.start_time as string) &&
      this.validateIsoTimestamp(obj.end_time as string) &&
      typeof obj.duration === 'number' &&
      typeof obj.stream_continues === 'boolean' &&
      obj.duration >= 0
    );
  }

  /**
   * Validate Stream URL Result - matches official RPC spec exactly
   */
  static validateStreamUrlResult(result: unknown): result is StreamUrlResult {
    if (typeof result !== 'object' || result === null) return false;
    
    const obj = result as Record<string, unknown>;
    
    return (
      this.validateDeviceId(obj.device as string) &&
      typeof obj.stream_name === 'string' &&
      typeof obj.stream_url === 'string' &&
      typeof obj.available === 'boolean' &&
      typeof obj.active_consumers === 'number' &&
      ['READY', 'NOT_READY', 'ERROR'].includes(obj.stream_status as string) &&
      obj.active_consumers >= 0
    );
  }

  /**
   * Validate Stream Status Result - matches official RPC spec exactly
   */
  static validateStreamStatusResult(result: unknown): result is StreamStatusResult {
    if (typeof result !== 'object' || result === null) return false;
    
    const obj = result as Record<string, unknown>;
    
    if (!this.validateDeviceId(obj.device as string)) return false;
    if (typeof obj.stream_name !== 'string') return false;
    if (!['ACTIVE', 'INACTIVE', 'ERROR', 'STARTING', 'STOPPING'].includes(obj.status as string)) return false;
    if (typeof obj.ready !== 'boolean') return false;
    if (!this.validateIsoTimestamp(obj.start_time as string)) return false;
    
    // Validate ffmpeg_process
    if (typeof obj.ffmpeg_process !== 'object' || obj.ffmpeg_process === null) return false;
    const ffmpeg = obj.ffmpeg_process as Record<string, unknown>;
    if (
      typeof ffmpeg.running !== 'boolean' ||
      typeof ffmpeg.pid !== 'number' ||
      typeof ffmpeg.uptime !== 'number' ||
      ffmpeg.pid < 0 ||
      ffmpeg.uptime < 0
    ) return false;
    
    // Validate mediamtx_path
    if (typeof obj.mediamtx_path !== 'object' || obj.mediamtx_path === null) return false;
    const mediamtx = obj.mediamtx_path as Record<string, unknown>;
    if (
      typeof mediamtx.exists !== 'boolean' ||
      typeof mediamtx.ready !== 'boolean' ||
      typeof mediamtx.readers !== 'number' ||
      mediamtx.readers < 0
    ) return false;
    
    // Validate metrics
    if (typeof obj.metrics !== 'object' || obj.metrics === null) return false;
    const metrics = obj.metrics as Record<string, unknown>;
    if (
      typeof metrics.bytes_sent !== 'number' ||
      typeof metrics.frames_sent !== 'number' ||
      typeof metrics.bitrate !== 'number' ||
      typeof metrics.fps !== 'number' ||
      metrics.bytes_sent < 0 ||
      metrics.frames_sent < 0 ||
      metrics.bitrate < 0 ||
      metrics.fps < 0
    ) return false;
    
    return true;
  }

  /**
   * Validate Streams List Result - matches official RPC spec exactly
   */
  static validateStreamsListResult(result: unknown): result is StreamsListResult[] {
    if (!Array.isArray(result)) return false;
    
    for (const stream of result) {
      if (typeof stream !== 'object' || stream === null) return false;
      const obj = stream as Record<string, unknown>;
      if (
        typeof obj.name !== 'string' ||
        typeof obj.source !== 'string' ||
        typeof obj.ready !== 'boolean' ||
        typeof obj.readers !== 'number' ||
        typeof obj.bytes_sent !== 'number' ||
        obj.readers < 0 ||
        obj.bytes_sent < 0
      ) {
        return false;
      }
    }
    
    return true;
  }

  /**
   * Validate Authentication Result - matches official RPC spec exactly
   * ALIGNED WITH REFACTORED AuthenticateResult type
   */
  static validateAuthenticateResult(result: unknown): result is AuthenticateResult {
    if (typeof result !== 'object' || result === null) return false;
    
    const obj = result as Record<string, unknown>;
    
    return (
      typeof obj.authenticated === 'boolean' &&
      ['admin', 'operator', 'viewer'].includes(obj.role as string) &&
      Array.isArray(obj.permissions) &&
      obj.permissions.every((permission: unknown) => typeof permission === 'string') &&
      this.validateIsoTimestamp(obj.expires_at as string) &&
      typeof obj.session_id === 'string'
    );
  }

  /**
   * Validate Server Info - matches official RPC spec exactly
   */
  static validateServerInfo(result: unknown): result is ServerInfo {
    if (typeof result !== 'object' || result === null) return false;
    
    const obj = result as Record<string, unknown>;
    
    return (
      typeof obj.name === 'string' &&
      typeof obj.version === 'string' &&
      typeof obj.build_date === 'string' &&
      typeof obj.go_version === 'string' &&
      typeof obj.architecture === 'string' &&
      Array.isArray(obj.capabilities) &&
      Array.isArray(obj.supported_formats) &&
      typeof obj.max_cameras === 'number' &&
      obj.capabilities.every((cap: unknown) => typeof cap === 'string') &&
      obj.supported_formats.every((format: unknown) => typeof format === 'string') &&
      obj.max_cameras > 0
    );
  }

  /**
   * Validate System Status - matches official RPC spec exactly
   */
  static validateSystemStatus(result: unknown): result is SystemStatus {
    if (typeof result !== 'object' || result === null) return false;
    
    const obj = result as Record<string, unknown>;
    
    if (!['HEALTHY', 'DEGRADED', 'UNHEALTHY'].includes(obj.status as string)) return false;
    if (typeof obj.uptime !== 'number' || obj.uptime < 0) return false;
    if (typeof obj.version !== 'string') return false;
    
    // Validate components
    if (typeof obj.components !== 'object' || obj.components === null) return false;
    const components = obj.components as Record<string, unknown>;
    const validStates = ['RUNNING', 'STOPPED', 'ERROR', 'STARTING', 'STOPPING'];
    
    return (
      validStates.includes(components.websocket_server as string) &&
      validStates.includes(components.camera_monitor as string) &&
      validStates.includes(components.mediamtx as string)
    );
  }

  /**
   * Validate Storage Info - matches official RPC spec exactly
   */
  static validateStorageInfo(result: unknown): result is StorageInfo {
    if (typeof result !== 'object' || result === null) return false;
    
    const obj = result as Record<string, unknown>;
    
    return (
      typeof obj.total_space === 'number' &&
      typeof obj.used_space === 'number' &&
      typeof obj.available_space === 'number' &&
      typeof obj.usage_percentage === 'number' &&
      typeof obj.recordings_size === 'number' &&
      typeof obj.snapshots_size === 'number' &&
      typeof obj.low_space_warning === 'boolean' &&
      obj.total_space >= 0 &&
      obj.used_space >= 0 &&
      obj.available_space >= 0 &&
      obj.usage_percentage >= 0 &&
      obj.usage_percentage <= 100 &&
      obj.recordings_size >= 0 &&
      obj.snapshots_size >= 0
    );
  }

  /**
   * Validate Metrics Result - matches official RPC spec exactly
   */
  static validateMetricsResult(result: unknown): result is MetricsResult {
    if (typeof result !== 'object' || result === null) return false;
    
    const obj = result as Record<string, unknown>;
    
    if (!this.validateIsoTimestamp(obj.timestamp as string)) return false;
    
    // Validate system_metrics
    if (typeof obj.system_metrics !== 'object' || obj.system_metrics === null) return false;
    const systemMetrics = obj.system_metrics as Record<string, unknown>;
    if (
      typeof systemMetrics.cpu_usage !== 'number' ||
      typeof systemMetrics.memory_usage !== 'number' ||
      typeof systemMetrics.disk_usage !== 'number' ||
      typeof systemMetrics.goroutines !== 'number' ||
      systemMetrics.cpu_usage < 0 || systemMetrics.cpu_usage > 100 ||
      systemMetrics.memory_usage < 0 || systemMetrics.memory_usage > 100 ||
      systemMetrics.disk_usage < 0 || systemMetrics.disk_usage > 100 ||
      systemMetrics.goroutines < 0
    ) return false;
    
    // Validate camera_metrics
    if (typeof obj.camera_metrics !== 'object' || obj.camera_metrics === null) return false;
    const cameraMetrics = obj.camera_metrics as Record<string, unknown>;
    if (
      typeof cameraMetrics.connected_cameras !== 'number' ||
      typeof cameraMetrics.cameras !== 'object' ||
      cameraMetrics.connected_cameras < 0
    ) return false;
    
    // Validate recording_metrics
    if (typeof obj.recording_metrics !== 'object' || obj.recording_metrics === null) return false;
    
    // Validate stream_metrics
    if (typeof obj.stream_metrics !== 'object' || obj.stream_metrics === null) return false;
    const streamMetrics = obj.stream_metrics as Record<string, unknown>;
    if (
      typeof streamMetrics.active_streams !== 'number' ||
      typeof streamMetrics.total_streams !== 'number' ||
      typeof streamMetrics.total_viewers !== 'number' ||
      streamMetrics.active_streams < 0 ||
      streamMetrics.total_streams < 0 ||
      streamMetrics.total_viewers < 0
    ) return false;
    
    return true;
  }

  /**
   * Validate Subscription Result - matches official RPC spec exactly
   */
  static validateSubscriptionResult(result: unknown): result is SubscriptionResult {
    if (typeof result !== 'object' || result === null) return false;
    
    const obj = result as Record<string, unknown>;
    
    return (
      typeof obj.subscribed === 'boolean' &&
      Array.isArray(obj.topics) &&
      obj.topics.every((topic: unknown) => typeof topic === 'string')
    );
  }

  /**
   * Validate Unsubscription Result - matches official RPC spec exactly
   */
  static validateUnsubscriptionResult(result: unknown): result is UnsubscriptionResult {
    if (typeof result !== 'object' || result === null) return false;
    
    const obj = result as Record<string, unknown>;
    
    return (
      typeof obj.unsubscribed === 'boolean' &&
      Array.isArray(obj.topics) &&
      obj.topics.every((topic: unknown) => typeof topic === 'string')
    );
  }

  /**
   * Validate Subscription Stats Result - matches official RPC spec exactly
   */
  static validateSubscriptionStatsResult(result: unknown): result is SubscriptionStatsResult {
    if (typeof result !== 'object' || result === null) return false;
    
    const obj = result as Record<string, unknown>;
    
    if (typeof obj.global_stats !== 'object' || obj.global_stats === null) return false;
    const globalStats = obj.global_stats as Record<string, unknown>;
    if (
      typeof globalStats.total_subscriptions !== 'number' ||
      typeof globalStats.active_clients !== 'number' ||
      typeof globalStats.topic_counts !== 'object' ||
      globalStats.total_subscriptions < 0 ||
      globalStats.active_clients < 0
    ) return false;
    
    return (
      Array.isArray(obj.client_topics) &&
      obj.client_topics.every((topic: unknown) => typeof topic === 'string') &&
      typeof obj.client_id === 'string'
    );
  }

  /**
   * Validate Delete Result - matches official RPC spec exactly
   */
  static validateDeleteResult(result: unknown): result is DeleteResult {
    if (typeof result !== 'object' || result === null) return false;
    
    const obj = result as Record<string, unknown>;
    
    return (
      typeof obj.filename === 'string' &&
      typeof obj.deleted === 'boolean' &&
      typeof obj.message === 'string'
    );
  }

  // ============================================================================
  // CONVENIENCE VALIDATION METHODS
  // ============================================================================

  /**
   * Validate any API response against its expected type
   */
  static validateApiResponse<T>(response: unknown, validator: (value: unknown) => value is T): response is T {
    return validator(response);
  }

  /**
   * Validate JSON-RPC 2.0 response envelope
   */
  static validateJsonRpcResponse(response: unknown): boolean {
    if (typeof response !== 'object' || response === null) return false;
    
    const obj = response as Record<string, unknown>;
    
    return (
      obj.jsonrpc === '2.0' &&
      (obj.result !== undefined || obj.error !== undefined) &&
      (obj.id === undefined || typeof obj.id === 'string' || typeof obj.id === 'number')
    );
  }

  /**
   * Validate JSON-RPC 2.0 error response
   */
  static validateJsonRpcError(response: unknown): boolean {
    if (typeof response !== 'object' || response === null) return false;
    
    const obj = response as Record<string, unknown>;
    
    return (
      obj.jsonrpc === '2.0' &&
      typeof obj.error === 'object' &&
      obj.error !== null &&
      typeof (obj.error as Record<string, unknown>).code === 'number' &&
      typeof (obj.error as Record<string, unknown>).message === 'string'
    );
  }

  // ============================================================================
  // MISSING VALIDATOR METHODS (CRITICAL FIX)
  // ============================================================================

  /**
   * Validate pagination parameters - matches official RPC spec
   */
  static validatePaginationParams(limit: number, offset: number): boolean {
    return (
      typeof limit === 'number' &&
      typeof offset === 'number' &&
      limit > 0 &&
      limit <= 1000 &&
      offset >= 0
    );
  }

  /**
   * Validate recording file object - matches official RPC spec
   */
  static validateRecordingFile(file: unknown): boolean {
    if (typeof file !== 'object' || file === null) return false;
    
    const obj = file as Record<string, unknown>;
    
    return (
      typeof obj.filename === 'string' &&
      typeof obj.file_size === 'number' &&
      typeof obj.created_time === 'string' &&
      typeof obj.download_url === 'string' &&
      obj.file_size >= 0 &&
      this.validateIsoTimestamp(obj.created_time)
    );
  }

  /**
   * Validate stream URL format - matches official RPC spec
   */
  static validateStreamUrl(url: string): boolean {
    try {
      const urlObj = new URL(url);
      return (
        urlObj.protocol === 'rtsp:' ||
        urlObj.protocol === 'https:' ||
        urlObj.protocol === 'http:'
      );
    } catch {
      return false;
    }
  }

  /**
   * Validate recording format - matches official RPC spec
   */
  static validateRecordingFormat(format: string): boolean {
    const validFormats = ['fmp4', 'mp4', 'mkv'];
    return validFormats.includes(format);
  }

  /**
   * Validate stream status - matches official RPC spec
   */
  static validateStreamStatus(status: unknown): boolean {
    if (typeof status !== 'object' || status === null) return false;
    
    const obj = status as Record<string, unknown>;
    
    return (
      typeof obj.device === 'string' &&
      typeof obj.status === 'string' &&
      typeof obj.ready === 'boolean' &&
      this.validateDeviceId(obj.device) &&
      ['ACTIVE', 'INACTIVE', 'ERROR', 'STARTING', 'STOPPING'].includes(obj.status)
    );
  }

  /**
   * Validate Ping Result - matches official RPC spec exactly
   */
  static validatePingResult(result: unknown): result is string {
    return typeof result === 'string' && result === 'pong';
  }

  /**
   * Validate Recording Info Result - matches official RPC spec exactly
   */
  static validateRecordingInfoResult(result: unknown): result is RecordingInfo {
    if (typeof result !== 'object' || result === null) return false;
    
    const obj = result as Record<string, unknown>;
    
    return (
      typeof obj.filename === 'string' &&
      typeof obj.file_size === 'number' &&
      typeof obj.duration === 'number' &&
      typeof obj.created_time === 'string' &&
      typeof obj.download_url === 'string' &&
      obj.file_size >= 0 &&
      obj.duration >= 0
    );
  }

  /**
   * Validate Snapshot Info Result - matches official RPC spec exactly
   */
  static validateSnapshotInfoResult(result: unknown): result is SnapshotInfo {
    if (typeof result !== 'object' || result === null) return false;
    
    const obj = result as Record<string, unknown>;
    
    return (
      typeof obj.filename === 'string' &&
      typeof obj.file_size === 'number' &&
      typeof obj.created_time === 'string' &&
      typeof obj.download_url === 'string' &&
      obj.file_size >= 0
    );
  }

  /**
   * Validate System Status Result - matches official RPC spec exactly
   */
  static validateSystemStatusResult(result: unknown): result is SystemStatus {
    if (typeof result !== 'object' || result === null) return false;
    
    const obj = result as Record<string, unknown>;
    
    return (
      typeof obj.status === 'string' &&
      typeof obj.uptime === 'number' &&
      typeof obj.version === 'string' &&
      typeof obj.components === 'object' &&
      obj.components !== null &&
      ['HEALTHY', 'DEGRADED', 'UNHEALTHY'].includes(obj.status) &&
      obj.uptime >= 0
    );
  }

  /**
   * Validate System Readiness Result - matches official RPC spec exactly
   */
  static validateSystemReadinessResult(result: unknown): boolean {
    if (typeof result !== 'object' || result === null) return false;
    
    const obj = result as Record<string, unknown>;
    
    return (
      typeof obj.status === 'string' &&
      typeof obj.message === 'string' &&
      Array.isArray(obj.available_cameras) &&
      typeof obj.discovery_active === 'boolean' &&
      ['starting', 'partial', 'ready'].includes(obj.status)
    );
  }

  /**
   * Validate Server Info Result - matches official RPC spec exactly
   */
  static validateServerInfoResult(result: unknown): result is ServerInfo {
    if (typeof result !== 'object' || result === null) return false;
    
    const obj = result as Record<string, unknown>;
    
    return (
      typeof obj.name === 'string' &&
      typeof obj.version === 'string' &&
      typeof obj.build_date === 'string' &&
      typeof obj.go_version === 'string' &&
      typeof obj.architecture === 'string' &&
      Array.isArray(obj.capabilities) &&
      Array.isArray(obj.supported_formats) &&
      typeof obj.max_cameras === 'number' &&
      obj.max_cameras >= 0
    );
  }

  /**
   * Validate Storage Info Result - matches official RPC spec exactly
   */
  static validateStorageInfoResult(result: unknown): result is StorageInfo {
    if (typeof result !== 'object' || result === null) return false;
    
    const obj = result as Record<string, unknown>;
    
    return (
      typeof obj.total_space === 'number' &&
      typeof obj.used_space === 'number' &&
      typeof obj.available_space === 'number' &&
      typeof obj.usage_percentage === 'number' &&
      typeof obj.recordings_size === 'number' &&
      typeof obj.snapshots_size === 'number' &&
      typeof obj.low_space_warning === 'boolean' &&
      obj.total_space >= 0 &&
      obj.used_space >= 0 &&
      obj.available_space >= 0 &&
      obj.usage_percentage >= 0 &&
      obj.usage_percentage <= 100 &&
      obj.recordings_size >= 0 &&
      obj.snapshots_size >= 0
    );
  }

  /**
   * Validate Retention Policy Result - matches official RPC spec exactly
   */
  static validateRetentionPolicyResult(result: unknown): boolean {
    if (typeof result !== 'object' || result === null) return false;
    
    const obj = result as Record<string, unknown>;
    
    return (
      typeof obj.policy_type === 'string' &&
      typeof obj.enabled === 'boolean' &&
      typeof obj.message === 'string' &&
      (obj.status === undefined || ['UPDATED', 'ERROR'].includes(obj.status as string))
    );
  }

  /**
   * Validate Cleanup Result - matches official RPC spec exactly
   */
  static validateCleanupResult(result: unknown): boolean {
    if (typeof result !== 'object' || result === null) return false;
    
    const obj = result as Record<string, unknown>;
    
    return (
      typeof obj.cleanup_executed === 'boolean' &&
      typeof obj.files_deleted === 'number' &&
      typeof obj.space_freed === 'number' &&
      typeof obj.message === 'string' &&
      obj.files_deleted >= 0 &&
      obj.space_freed >= 0
    );
  }

  /**
   * Validate External Stream Discovery Result - matches official RPC spec exactly
   */
  static validateExternalStreamDiscoveryResult(result: unknown): boolean {
    if (typeof result !== 'object' || result === null) return false;
    
    const obj = result as Record<string, unknown>;
    
    return (
      Array.isArray(obj.discovered_streams) &&
      Array.isArray(obj.skydio_streams) &&
      Array.isArray(obj.generic_streams) &&
      typeof obj.scan_timestamp === 'string' &&
      typeof obj.total_found === 'number' &&
      obj.total_found >= 0
    );
  }

  /**
   * Validate External Stream Add Result - matches official RPC spec exactly
   */
  static validateExternalStreamAddResult(result: unknown): boolean {
    if (typeof result !== 'object' || result === null) return false;
    
    const obj = result as Record<string, unknown>;
    
    return (
      typeof obj.stream_url === 'string' &&
      typeof obj.stream_name === 'string' &&
      typeof obj.stream_type === 'string' &&
      typeof obj.status === 'string' &&
      typeof obj.timestamp === 'string' &&
      ['ADDED', 'ERROR'].includes(obj.status)
    );
  }

  /**
   * Validate External Stream Remove Result - matches official RPC spec exactly
   */
  static validateExternalStreamRemoveResult(result: unknown): boolean {
    if (typeof result !== 'object' || result === null) return false;
    
    const obj = result as Record<string, unknown>;
    
    return (
      typeof obj.stream_url === 'string' &&
      typeof obj.status === 'string' &&
      typeof obj.timestamp === 'string' &&
      ['REMOVED', 'ERROR'].includes(obj.status)
    );
  }

  /**
   * Validate External Streams List Result - matches official RPC spec exactly
   */
  static validateExternalStreamsListResult(result: unknown): boolean {
    if (typeof result !== 'object' || result === null) return false;
    
    const obj = result as Record<string, unknown>;
    
    return (
      Array.isArray(obj.external_streams) &&
      Array.isArray(obj.skydio_streams) &&
      Array.isArray(obj.generic_streams) &&
      typeof obj.total_count === 'number' &&
      typeof obj.timestamp === 'string' &&
      obj.total_count >= 0
    );
  }

  /**
   * Validate Discovery Interval Set Result - matches official RPC spec exactly
   */
  static validateDiscoveryIntervalSetResult(result: unknown): boolean {
    if (typeof result !== 'object' || result === null) return false;
    
    const obj = result as Record<string, unknown>;
    
    return (
      typeof obj.scan_interval === 'number' &&
      typeof obj.status === 'string' &&
      typeof obj.message === 'string' &&
      typeof obj.timestamp === 'string' &&
      ['UPDATED', 'ERROR'].includes(obj.status) &&
      obj.scan_interval >= 0
    );
  }

  /**
   * Validate Camera Status Update Notification - matches official RPC spec exactly
   */
  static validateCameraStatusUpdateNotification(notification: unknown): boolean {
    if (typeof notification !== 'object' || notification === null) return false;
    
    const obj = notification as Record<string, unknown>;
    
    return (
      typeof obj.device === 'string' &&
      typeof obj.status === 'string' &&
      typeof obj.name === 'string' &&
      typeof obj.resolution === 'string' &&
      typeof obj.fps === 'number' &&
      typeof obj.streams === 'object' &&
      obj.streams !== null &&
      ['CONNECTED', 'DISCONNECTED', 'ERROR'].includes(obj.status) &&
      obj.fps >= 0
    );
  }

  /**
   * Validate Recording Status Update Notification - matches official RPC spec exactly
   */
  static validateRecordingStatusUpdateNotification(notification: unknown): boolean {
    if (typeof notification !== 'object' || notification === null) return false;
    
    const obj = notification as Record<string, unknown>;
    
    return (
      typeof obj.device === 'string' &&
      typeof obj.status === 'string' &&
      typeof obj.filename === 'string' &&
      typeof obj.duration === 'number' &&
      ['STARTED', 'STOPPED', 'ERROR'].includes(obj.status) &&
      obj.duration >= 0
    );
  }
}