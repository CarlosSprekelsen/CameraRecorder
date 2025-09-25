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
  ServerInfo,
  SystemStatus,
  StorageInfo,
  MetricsResult,
  CameraStatusResult,
  StreamUrlResult,
  SnapshotResult,
  RecordingResult
} from '../../src/types/api';

export class APIMocks {
  /**
   * Get mock camera list result
   * MANDATORY: Use this mock for all camera list tests
   */
  static getCameraListResult(): CameraStatusResult[] {
    return [
      {
        device: 'camera0',
        status: 'CONNECTED',
        last_seen: new Date().toISOString(),
        capabilities: ['recording', 'streaming', 'snapshots']
      },
      {
        device: 'camera1',
        status: 'DISCONNECTED',
        last_seen: new Date(Date.now() - 3600000).toISOString(),
        capabilities: ['recording', 'streaming']
      }
    ];
  }

  /**
   * Get mock camera object
   * MANDATORY: Use this mock for all camera tests
   */
  static getCamera(device: string = 'camera0'): CameraStatusResult {
    return {
      device,
      status: 'CONNECTED',
      last_seen: new Date().toISOString(),
      capabilities: ['recording', 'streaming', 'snapshots']
    };
  }

  /**
   * Get mock recording start result
   * MANDATORY: Use this mock for all recording start tests
   */
  static getRecordingStartResult(device: string = 'camera0'): RecordingResult {
    return {
      success: true,
      recording_id: `recording_${device}_${Date.now()}`,
      status: 'started'
    };
  }

  /**
   * Get mock recording stop result
   * MANDATORY: Use this mock for all recording stop tests
   */
  static getRecordingStopResult(device: string = 'camera0'): RecordingResult {
    return {
      success: true,
      recording_id: `recording_${device}_${Date.now()}`,
      status: 'stopped'
    };
  }

  /**
   * Get mock snapshot info
   * MANDATORY: Use this mock for all snapshot tests
   */
  static getSnapshotInfo(device: string = 'camera0'): SnapshotResult {
    return {
      success: true,
      filename: `snapshot_${device}_${Date.now()}.jpg`,
      download_url: `https://localhost/downloads/snapshot_${device}_${Date.now()}.jpg`
    };
  }

  /**
   * Get mock file list result
   * MANDATORY: Use this mock for all file list tests
   */
  static getListFilesResult(type: 'recordings' | 'snapshots' = 'recordings'): any {
    const files: any[] = [
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
  static getStreamStatus(device: string = 'camera0'): StreamUrlResult {
    return {
      device,
      hls_url: `https://localhost/hls/${device}.m3u8`,
      webrtc_url: `https://localhost/webrtc/${device}`,
      status: 'ACTIVE'
    };
  }

  /**
   * Get mock metrics result
   * MANDATORY: Use this mock for all metrics tests
   */
  static getMetricsResult(): MetricsResult {
    return {
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
        active_recordings: { count: 1 },
        total_recordings: { count: 5 },
        total_size: { bytes: 1024000000 }
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
  static getStatusResult(): SystemStatus {
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
  static getAuthResult(role: 'admin' | 'operator' | 'viewer' = 'admin'): any {
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
      setService: () => {},
      takeSnapshot: () => Promise.resolve(),
      startRecording: () => Promise.resolve(),
      stopRecording: () => Promise.resolve(),
      handleRecordingStatusUpdate: () => {},
      reset: () => {}
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
      setStatus: () => {},
      setError: () => {},
      setReconnectAttempts: () => {},
      setLastConnected: () => {},
      reset: () => {}
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
      setToken: () => {},
      setRole: () => {},
      setSessionId: () => {},
      setExpiresAt: () => {},
      setPermissions: () => {},
      setAuthenticated: () => {},
      login: () => {},
      logout: () => {},
      reset: () => {}
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
      setInfo: () => {},
      setStatus: () => {},
      setStorage: () => {},
      setLoading: () => {},
      setError: () => {},
      setLastUpdated: () => {},
      reset: () => {}
    };
  }

  /**
   * Create mock device service
   * MANDATORY: Use this mock for all device service tests
   */
  static createMockDeviceService() {
    return {
      getCameraList: () => Promise.resolve([]),
      getStreamUrl: () => Promise.resolve(null),
      getStreams: () => Promise.resolve([]),
      getCameraStatus: () => Promise.resolve(null),
      getCameraCapabilities: () => Promise.resolve(null)
    };
  }

  /**
   * Create mock file service
   * MANDATORY: Use this mock for all file service tests
   */
  static createMockFileService() {
    return {
      listRecordings: () => Promise.resolve({ files: [], total: 0, limit: 20, offset: 0 }),
      listSnapshots: () => Promise.resolve({ files: [], total: 0, limit: 20, offset: 0 }),
      getRecordingInfo: () => Promise.resolve(null),
      getSnapshotInfo: () => Promise.resolve(null),
      downloadFile: () => Promise.resolve(),
      deleteRecording: () => Promise.resolve({ success: false, message: 'Not implemented' }),
      deleteSnapshot: () => Promise.resolve({ success: false, message: 'Not implemented' })
    };
  }

  /**
   * Create mock recording service
   * MANDATORY: Use this mock for all recording service tests
   */
  static createMockRecordingService() {
    return {
      takeSnapshot: () => Promise.resolve(),
      startRecording: () => Promise.resolve(),
      stopRecording: () => Promise.resolve()
    };
  }

  /**
   * Get centralized logger mock
   * SINGLE mock implementation for logger service
   * MANDATORY: Use this mock for all logger tests
   */
  static getMockLogger() {
    return {
      info: jest.fn(),
      warn: jest.fn(),
      error: jest.fn()
    };
  }
}
