/**
 * SINGLE mock implementation per API concern
 * Based on documented API responses only
 * 
 * Ground Truth References:
 * - API Documentation: ../mediamtx-camera-service-go/docs/api/mediamtx_camera_service_openrpc.json
 * 
 * Requirements Coverage:
 * - REQ-MOCK-001: Consistent mock patterns
 * - REQ-MOCK-002: API-compliant responses
 * - REQ-MOCK-003: No duplicate implementations
 * 
 * Test Categories: Unit/Mock
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
  AuthResult,
  RecordingFile
} from '../../../src/types/api';

export class APIMocks {
  /**
   * Get mock camera list result
   * MANDATORY: Use this mock for all camera list tests
   */
  static getCameraListResult(): CameraListResult {
    return {
      cameras: [
        {
          device: 'camera0',
          status: 'CONNECTED',
          name: 'Test Camera 0',
          resolution: '1920x1080',
          fps: 30,
          streams: {
            rtsp: 'rtsp://localhost:8554/camera0',
            hls: 'https://localhost/hls/camera0.m3u8'
          }
        },
        {
          device: 'camera1',
          status: 'DISCONNECTED',
          name: 'Test Camera 1',
          resolution: '1280x720',
          fps: 25,
          streams: {
            rtsp: 'rtsp://localhost:8554/camera1',
            hls: 'https://localhost/hls/camera1.m3u8'
          }
        }
      ],
      total: 2,
      connected: 1
    };
  }

  /**
   * Get mock camera object
   * MANDATORY: Use this mock for all camera tests
   */
  static getCamera(device: string = 'camera0'): Camera {
    return {
      device,
      status: 'CONNECTED',
      name: `Test Camera ${device}`,
      resolution: '1920x1080',
      fps: 30,
      streams: {
        rtsp: `rtsp://localhost:8554/${device}`,
        hls: `https://localhost/hls/${device}.m3u8`
      }
    };
  }

  /**
   * Get mock recording start result
   * MANDATORY: Use this mock for all recording start tests
   */
  static getRecordingStartResult(device: string = 'camera0'): RecordingStart {
    return {
      device,
      status: 'RECORDING',
      start_time: new Date().toISOString(),
      filename: `recording_${device}_${Date.now()}.mp4`,
      format: 'mp4'
    };
  }

  /**
   * Get mock recording stop result
   * MANDATORY: Use this mock for all recording stop tests
   */
  static getRecordingStopResult(device: string = 'camera0'): RecordingStop {
    return {
      device,
      status: 'STOPPED',
      start_time: new Date(Date.now() - 60000).toISOString(),
      end_time: new Date().toISOString(),
      duration: 60,
      file_size: 1024000,
      filename: `recording_${device}_${Date.now()}.mp4`,
      format: 'mp4'
    };
  }

  /**
   * Get mock snapshot info
   * MANDATORY: Use this mock for all snapshot tests
   */
  static getSnapshotInfo(device: string = 'camera0'): SnapshotInfo {
    return {
      device,
      filename: `snapshot_${device}_${Date.now()}.jpg`,
      status: 'SUCCESS',
      timestamp: new Date().toISOString(),
      file_size: 512000,
      file_path: `/snapshots/snapshot_${device}_${Date.now()}.jpg`
    };
  }

  /**
   * Get mock file list result
   * MANDATORY: Use this mock for all file list tests
   */
  static getListFilesResult(type: 'recordings' | 'snapshots' = 'recordings'): ListFilesResult {
    const files: RecordingFile[] = [
      {
        filename: `${type}_camera0_${Date.now()}.${type === 'recordings' ? 'mp4' : 'jpg'}`,
        file_size: type === 'recordings' ? 1024000 : 512000,
        modified_time: new Date().toISOString(),
        download_url: `https://localhost/downloads/${type}_camera0_${Date.now()}.${type === 'recordings' ? 'mp4' : 'jpg'}`
      },
      {
        filename: `${type}_camera1_${Date.now()}.${type === 'recordings' ? 'mp4' : 'jpg'}`,
        file_size: type === 'recordings' ? 2048000 : 768000,
        modified_time: new Date(Date.now() - 3600000).toISOString(),
        download_url: `https://localhost/downloads/${type}_camera1_${Date.now()}.${type === 'recordings' ? 'mp4' : 'jpg'}`
      }
    ];

    return {
      files,
      total: 2,
      limit: 50,
      offset: 0
    };
  }

  /**
   * Get mock stream status
   * MANDATORY: Use this mock for all stream tests
   */
  static getStreamStatus(device: string = 'camera0'): StreamStatus {
    return {
      device,
      status: 'ACTIVE',
      ready: true,
      ffmpeg_process: {
        running: true,
        pid: 12345,
        uptime: 300
      },
      mediamtx_path: {
        exists: true,
        ready: true,
        readers: 2
      },
      metrics: {
        bytes_sent: 1024000,
        frames_sent: 9000,
        bitrate: 2000000,
        fps: 30
      },
      start_time: new Date(Date.now() - 300000).toISOString()
    };
  }

  /**
   * Get mock metrics result
   * MANDATORY: Use this mock for all metrics tests
   */
  static getMetricsResult(): MetricsResult {
    return {
      timestamp: new Date().toISOString(),
      system_metrics: {
        cpu_usage: 45.5,
        memory_usage: 67.2,
        disk_usage: 23.8,
        goroutines: 150
      },
      camera_metrics: {
        connected_cameras: 2,
        cameras: {
          camera0: {
            status: 'CONNECTED',
            fps: 30,
            bitrate: 2000000
          },
          camera1: {
            status: 'DISCONNECTED',
            fps: 0,
            bitrate: 0
          }
        }
      },
      recording_metrics: {
        active_recordings: 1,
        total_recordings: 5,
        total_size: 1024000000
      },
      stream_metrics: {
        active_streams: 2,
        total_streams: 2,
        total_viewers: 3
      }
    };
  }

  /**
   * Get mock status result
   * MANDATORY: Use this mock for all status tests
   */
  static getStatusResult(): StatusResult {
    return {
      status: 'HEALTHY',
      uptime: 3600.5,
      version: '1.0.0',
      components: {
        websocket_server: 'HEALTHY',
        camera_monitor: 'HEALTHY',
        mediamtx: 'HEALTHY'
      }
    };
  }

  /**
   * Get mock server info
   * MANDATORY: Use this mock for all server info tests
   */
  static getServerInfo(): ServerInfo {
    return {
      name: 'MediaMTX Camera Service',
      version: '1.0.0',
      build_date: '2025-01-25T10:00:00Z',
      go_version: '1.21.0',
      architecture: 'linux/amd64',
      capabilities: ['recording', 'streaming', 'snapshots', 'file_management'],
      supported_formats: ['fmp4', 'mp4', 'mkv'],
      max_cameras: 10
    };
  }

  /**
   * Get mock storage info
   * MANDATORY: Use this mock for all storage tests
   */
  static getStorageInfo(): StorageInfo {
    return {
      total_space: 1000000000000, // 1TB
      used_space: 250000000000,  // 250GB
      available_space: 750000000000, // 750GB
      usage_percentage: 25.0,
      recordings_size: 200000000000, // 200GB
      snapshots_size: 50000000000,   // 50GB
      low_space_warning: false
    };
  }

  /**
   * Get mock authentication result
   * MANDATORY: Use this mock for all auth tests
   */
  static getAuthResult(role: 'admin' | 'operator' | 'viewer' = 'admin'): AuthResult {
    return {
      authenticated: true,
      role,
      permissions: this.getRolePermissions(role),
      expires_at: new Date(Date.now() + 3600000).toISOString(),
      session_id: `test-session-${Date.now()}`
    };
  }

  /**
   * Get mock error response
   * MANDATORY: Use this mock for all error tests
   */
  static getErrorResponse(code: number, message?: string): any {
    const errorMessages: { [key: number]: string } = {
      [-32600]: 'Invalid Request',
      [-32601]: 'Method Not Found',
      [-32602]: 'Invalid Params',
      [-32603]: 'Internal Error',
      [-32001]: 'Auth Failed',
      [-32002]: 'Permission Denied',
      [-32010]: 'Not Found',
      [-32020]: 'Invalid State',
      [-32030]: 'Unsupported',
      [-32040]: 'Rate Limited',
      [-32050]: 'Dependency Failed'
    };

    return {
      jsonrpc: '2.0',
      error: {
        code,
        message: message || errorMessages[code] || 'Unknown Error'
      },
      id: 1
    };
  }

  /**
   * Get role-specific permissions
   * MANDATORY: Use this method for role permission tests
   */
  private static getRolePermissions(role: 'admin' | 'operator' | 'viewer'): string[] {
    switch (role) {
      case 'admin':
        return ['read', 'write', 'delete', 'admin'];
      case 'operator':
        return ['read', 'write'];
      case 'viewer':
        return ['read'];
      default:
        return [];
    }
  }

  /**
   * Get mock WebSocket message
   * MANDATORY: Use this mock for all WebSocket tests
   */
  static getWebSocketMessage(method: string, params: any[] = []): any {
    return {
      jsonrpc: '2.0',
      method,
      params,
      id: Math.floor(Math.random() * 1000)
    };
  }

  /**
   * Get mock WebSocket response
   * MANDATORY: Use this mock for all WebSocket response tests
   */
  static getWebSocketResponse(result: any, id: number = 1): any {
    return {
      jsonrpc: '2.0',
      result,
      id
    };
  }

  /**
   * Get mock WebSocket notification
   * MANDATORY: Use this mock for all WebSocket notification tests
   */
  static getWebSocketNotification(method: string, params: any[] = []): any {
    return {
      jsonrpc: '2.0',
      method,
      params
    };
  }

  /**
   * Create mock device store
   * MANDATORY: Use this mock for all device store tests
   */
  static createMockDeviceStore() {
    return {
      cameras: [],
      streams: [],
      loading: false,
      error: null,
      lastUpdated: null,
      getCameraList: () => Promise.resolve(),
      getStreamUrl: () => Promise.resolve(null),
      getStreams: () => Promise.resolve(),
      setLoading: () => {},
      setError: () => {},
      updateCameraStatus: () => {},
      updateStreamStatus: () => {},
      handleCameraStatusUpdate: () => {},
      handleStreamUpdate: () => {},
      setDeviceService: () => {},
      reset: () => {}
    };
  }

  /**
   * Create mock file store
   * MANDATORY: Use this mock for all file store tests
   */
  static createMockFileStore() {
    return {
      recordings: [],
      snapshots: [],
      loading: false,
      error: null,
      pagination: { limit: 20, offset: 0, total: 0 },
      selectedFiles: [],
      currentTab: 'recordings' as const,
      loadRecordings: () => Promise.resolve(),
      loadSnapshots: () => Promise.resolve(),
      getRecordingInfo: () => Promise.resolve(null),
      getSnapshotInfo: () => Promise.resolve(null),
      downloadFile: () => Promise.resolve(),
      deleteRecording: () => Promise.resolve(false),
      deleteSnapshot: () => Promise.resolve(false),
      setLoading: () => {},
      setError: () => {},
      setCurrentTab: () => {},
      setSelectedFiles: () => {},
      toggleFileSelection: () => {},
      clearSelection: () => {},
      setPagination: () => {},
      nextPage: () => {},
      prevPage: () => {},
      goToPage: () => {},
      setFileService: () => {},
      reset: () => {}
    };
  }

  /**
   * Create mock recording store
   * MANDATORY: Use this mock for all recording store tests
   */
  static createMockRecordingStore() {
    return {
      activeRecordings: {},
      history: [],
      loading: false,
      error: null,
      setService: jest.fn(),
      takeSnapshot: jest.fn(),
      startRecording: jest.fn(),
      stopRecording: jest.fn(),
      handleRecordingStatusUpdate: jest.fn(),
      reset: jest.fn()
    };
  }

  /**
   * Create mock connection store
   * MANDATORY: Use this mock for all connection store tests
   */
  static createMockConnectionStore() {
    return {
      status: 'disconnected' as const,
      lastError: null,
      reconnectAttempts: 0,
      lastConnected: null,
      setStatus: jest.fn(),
      setError: jest.fn(),
      setReconnectAttempts: jest.fn(),
      setLastConnected: jest.fn(),
      reset: jest.fn()
    };
  }

  /**
   * Create mock auth store
   * MANDATORY: Use this mock for all auth store tests
   */
  static createMockAuthStore() {
    return {
      token: null,
      role: null,
      session_id: null,
      isAuthenticated: false,
      expires_at: null,
      permissions: [],
      setToken: jest.fn(),
      setRole: jest.fn(),
      setSessionId: jest.fn(),
      setExpiresAt: jest.fn(),
      setPermissions: jest.fn(),
      setAuthenticated: jest.fn(),
      login: jest.fn(),
      logout: jest.fn(),
      reset: jest.fn()
    };
  }

  /**
   * Create mock server store
   * MANDATORY: Use this mock for all server store tests
   */
  static createMockServerStore() {
    return {
      info: null,
      status: null,
      storage: null,
      loading: false,
      error: null,
      lastUpdated: null,
      setInfo: jest.fn(),
      setStatus: jest.fn(),
      setStorage: jest.fn(),
      setLoading: jest.fn(),
      setError: jest.fn(),
      setLastUpdated: jest.fn(),
      reset: jest.fn()
    };
  }

  /**
   * Create mock device service
   * MANDATORY: Use this mock for all device service tests
   */
  static createMockDeviceService() {
    return {
      getCameraList: jest.fn(),
      getStreamUrl: jest.fn(),
      getStreams: jest.fn(),
      getCameraStatus: jest.fn(),
      getCameraCapabilities: jest.fn()
    };
  }

  /**
   * Create mock file service
   * MANDATORY: Use this mock for all file service tests
   */
  static createMockFileService() {
    return {
      listRecordings: jest.fn(),
      listSnapshots: jest.fn(),
      getRecordingInfo: jest.fn(),
      getSnapshotInfo: jest.fn(),
      downloadFile: jest.fn(),
      deleteRecording: jest.fn(),
      deleteSnapshot: jest.fn()
    };
  }

  /**
   * Create mock recording service
   * MANDATORY: Use this mock for all recording service tests
   */
  static createMockRecordingService() {
    return {
      takeSnapshot: jest.fn(),
      startRecording: jest.fn(),
      stopRecording: jest.fn()
    };
  }
}
