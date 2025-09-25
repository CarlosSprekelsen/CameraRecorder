// Centralized mocks for MediaMTX Camera Service Client tests
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
  // Client State Types
  ConnectionState,
  AuthState,
  ServerState
} from '../../src/types/api';

/**
 * Centralized mock utilities for MediaMTX Camera Service Client tests
 * All mocks comply with official RPC documentation schemas
 */

// ============================================================================
// CENTRALIZED MOCK UTILITIES (RULE COMPLIANCE)
// ============================================================================

/**
 * Centralized logger mock - prevents duplicate implementations
 */
export const mockLogger = {
  logger: {
    info: jest.fn(),
    warn: jest.fn(),
    error: jest.fn()
  }
};

/**
 * Centralized router mock - prevents duplicate implementations
 */
export const mockRouter = {
  useNavigate: jest.fn()
};

/**
 * Centralized auth store mock - prevents duplicate implementations
 */
export const mockAuthStore = jest.fn();

/**
 * Centralized mock data factory for MediaMTX Camera Service Client tests
 * All mocks comply with official RPC documentation schemas
 */
export class MockDataFactory {
  // ============================================================================
  // OFFICIAL RPC API MOCKS (GROUND TRUTH COMPLIANCE)
  // ============================================================================

  /**
   * Mock Camera List Result - matches official RPC spec exactly
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
          fps: 15,
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
   * Mock Camera Status Result - matches official RPC spec exactly
   */
  static getCameraStatusResult(): CameraStatusResult {
    return {
      device: 'camera0',
      status: 'CONNECTED',
      name: 'Test Camera 0',
      resolution: '1920x1080',
      fps: 30,
      streams: {
        rtsp: 'rtsp://localhost:8554/camera0',
        hls: 'https://localhost/hls/camera0.m3u8'
      },
      metrics: {
        bytes_sent: 12345678,
        readers: 2,
        uptime: 3600
      },
      capabilities: {
        formats: ['YUYV', 'MJPEG'],
        resolutions: ['1920x1080', '1280x720']
      }
    };
  }

  /**
   * Mock Camera Capabilities Result - matches official RPC spec exactly
   */
  static getCameraCapabilitiesResult(): CameraCapabilitiesResult {
    return {
      device: 'camera0',
      formats: ['YUYV', 'MJPEG', 'RGB24'],
      resolutions: ['1920x1080', '1280x720', '640x480'],
      fps_options: [15, 30, 60],
      validation_status: 'CONFIRMED'
    };
  }

  /**
   * Mock Recording Start Result - matches official RPC spec exactly
   */
  static getRecordingStartResult(): RecordingStartResult {
    return {
      device: 'camera0',
      filename: 'camera0_2025-01-15_14-30-00',
      status: 'RECORDING',
      start_time: '2025-01-15T14:30:00Z',
      format: 'fmp4'
    };
  }

  /**
   * Mock Recording Stop Result - matches official RPC spec exactly
   */
  static getRecordingStopResult(): RecordingStopResult {
    return {
      device: 'camera0',
      filename: 'camera0_2025-01-15_14-30-00',
      status: 'STOPPED',
      start_time: '2025-01-15T14:30:00Z',
      end_time: '2025-01-15T15:00:00Z',
      duration: 1800,
      file_size: 1073741824,
      format: 'fmp4'
    };
  }

  /**
   * Mock Snapshot Result - matches official RPC spec exactly
   */
  static getSnapshotResult(): SnapshotResult {
    return {
      device: 'camera0',
      filename: 'snapshot_2025-01-15_14-30-00.jpg',
      status: 'SUCCESS',
      timestamp: '2025-01-15T14:30:00Z',
      file_size: 204800,
      file_path: '/opt/camera-service/snapshots/snapshot_2025-01-15_14-30-00.jpg'
    };
  }

  /**
   * Mock Snapshot Info - matches official RPC spec exactly
   */
  static getSnapshotInfo(): SnapshotInfo {
    return {
      filename: 'snapshot_2025-01-15_14-30-00.jpg',
      file_size: 204800,
      created_time: '2025-01-15T14:30:00Z',
      download_url: '/files/snapshots/snapshot_2025-01-15_14-30-00.jpg'
    };
  }

  /**
   * Mock Recording Info - matches official RPC spec exactly
   */
  static getRecordingInfo(): RecordingInfo {
    return {
      filename: 'camera0_2025-01-15_14-30-00',
      file_size: 1073741824,
      duration: 3600,
      created_time: '2025-01-15T14:30:00Z',
      download_url: '/files/recordings/camera0_2025-01-15_14-30-00.fmp4'
    };
  }

  /**
   * Mock File List Result - matches official RPC spec exactly
   */
  static getFileListResult(): FileListResult {
    return {
      files: [
        {
          filename: 'camera0_2025-01-15_14-30-00',
          file_size: 1073741824,
          modified_time: '2025-01-15T14:30:00Z',
          download_url: '/files/recordings/camera0_2025-01-15_14-30-00.fmp4'
        },
        {
          filename: 'camera0_2025-01-15_15-00-00',
          file_size: 2147483648,
          modified_time: '2025-01-15T15:00:00Z',
          download_url: '/files/recordings/camera0_2025-01-15_15-00-00.fmp4'
        }
      ],
      total: 25,
      limit: 10,
      offset: 0
    };
  }

  /**
   * Mock Stream Start Result - matches official RPC spec exactly
   */
  static getStreamStartResult(): StreamStartResult {
    return {
      device: 'camera0',
      stream_name: 'camera_video0_viewing',
      stream_url: 'rtsp://localhost:8554/camera_video0_viewing',
      status: 'STARTED',
      start_time: '2025-01-15T14:30:00Z',
      auto_close_after: '300s',
      ffmpeg_command: 'ffmpeg -f v4l2 -i /dev/video0 -c:v libx264 -preset ultrafast -tune zerolatency -f rtsp rtsp://localhost:8554/camera_video0_viewing'
    };
  }

  /**
   * Mock Stream Stop Result - matches official RPC spec exactly
   */
  static getStreamStopResult(): StreamStopResult {
    return {
      device: 'camera0',
      stream_name: 'camera_video0_viewing',
      status: 'STOPPED',
      start_time: '2025-01-15T14:30:00Z',
      end_time: '2025-01-15T14:35:00Z',
      duration: 300,
      stream_continues: false,
      message: 'Stream stopped successfully'
    };
  }

  /**
   * Mock Stream URL Result - matches official RPC spec exactly
   */
  static getStreamUrlResult(): StreamUrlResult {
    return {
      device: 'camera0',
      stream_name: 'camera_video0_viewing',
      stream_url: 'rtsp://localhost:8554/camera_video0_viewing',
      available: true,
      active_consumers: 2,
      stream_status: 'READY'
    };
  }

  /**
   * Mock Stream Status Result - matches official RPC spec exactly
   */
  static getStreamStatusResult(): StreamStatusResult {
    return {
      device: 'camera0',
      stream_name: 'camera_video0_viewing',
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
        bytes_sent: 12345678,
        frames_sent: 9000,
        bitrate: 600000,
        fps: 30
      },
      start_time: '2025-01-15T14:30:00Z'
    };
  }

  /**
   * Mock Streams List Result - matches official RPC spec exactly
   */
  static getStreamsListResult(): StreamsListResult[] {
    return [
      {
        name: 'camera0',
        source: 'ffmpeg -f v4l2 -i /dev/video0 -c:v libx264 -profile:v baseline -level 3.0 -pix_fmt yuv420p -preset ultrafast -b:v 600k -f rtsp rtsp://127.0.0.1:8554/camera0',
        ready: true,
        readers: 2,
        bytes_sent: 12345678
      },
      {
        name: 'camera1',
        source: 'ffmpeg -f v4l2 -i /dev/video1 -c:v libx264 -profile:v baseline -level 3.0 -pix_fmt yuv420p -preset ultrafast -b:v 600k -f rtsp rtsp://127.0.0.1:8554/camera1',
        ready: false,
        readers: 0,
        bytes_sent: 0
      }
    ];
  }

  /**
   * Mock Authentication Result - matches official RPC spec exactly
   */
  static getAuthenticateResult(): AuthenticateResult {
    return {
      authenticated: true,
      role: 'operator',
      permissions: ['view', 'control'],
      expires_at: '2025-01-16T14:30:00.000Z',
      session_id: '550e8400-e29b-41d4-a716-446655440000'
    };
  }

  /**
   * Mock Server Info - matches official RPC spec exactly
   */
  static getServerInfo(): ServerInfo {
    return {
      name: 'MediaMTX Camera Service',
      version: '1.0.0',
      build_date: '2025-01-15',
      go_version: 'go1.24.6',
      architecture: 'amd64',
      capabilities: ['snapshots', 'recordings', 'streaming'],
      supported_formats: ['fmp4', 'mp4', 'mkv', 'jpg'],
      max_cameras: 10
    };
  }

  /**
   * Mock System Status - matches official RPC spec exactly
   */
  static getSystemStatus(): SystemStatus {
    return {
      status: 'HEALTHY',
      uptime: 86400.5,
      version: '1.0.0',
      components: {
        websocket_server: 'RUNNING',
        camera_monitor: 'RUNNING',
        mediamtx: 'RUNNING'
      }
    };
  }

  /**
   * Mock Storage Info - matches official RPC spec exactly
   */
  static getStorageInfo(): StorageInfo {
    return {
      total_space: 107374182400,
      used_space: 53687091200,
      available_space: 53687091200,
      usage_percentage: 50.0,
      recordings_size: 42949672960,
      snapshots_size: 10737418240,
      low_space_warning: false
    };
  }

  /**
   * Mock Metrics Result - matches official RPC spec exactly
   */
  static getMetricsResult(): MetricsResult {
    return {
      timestamp: '2025-01-15T14:30:00Z',
      system_metrics: {
        cpu_usage: 23.1,
        memory_usage: 85.5,
        disk_usage: 45.5,
        goroutines: 150
      },
      camera_metrics: {
        connected_cameras: 2,
        cameras: {
          'camera0': {
            path: 'camera0',
            name: 'USB 2.0 Camera: USB 2.0 Camera',
            status: 'CONNECTED',
            device_num: 0,
            last_seen: '2025-01-15T14:30:00Z',
            capabilities: {
              driver_name: 'uvcvideo',
              card_name: 'USB 2.0 Camera: USB 2.0 Camera',
              bus_info: 'usb-0000:00:1a.0-1.2',
              version: '6.14.8',
              capabilities: ['0x84a00001', 'Video Capture', 'Metadata Capture', 'Streaming', 'Extended Pix Format'],
              device_caps: ['0x04200001', 'Video Capture', 'Streaming', 'Extended Pix Format']
            },
            formats: [
              {
                pixel_format: 'YUYV',
                width: 640,
                height: 480,
                frame_rates: ['30.000', '20.000', '15.000', '10.000', '5.000']
              }
            ]
          }
        }
      },
      recording_metrics: {},
      stream_metrics: {
        active_streams: 0,
        total_streams: 4,
        total_viewers: 0
      }
    };
  }

  /**
   * Mock Subscription Result - matches official RPC spec exactly
   */
  static getSubscriptionResult(): SubscriptionResult {
    return {
      subscribed: true,
      topics: ['camera.connected', 'recording.start'],
      filters: {
        device: 'camera0'
      }
    };
  }

  /**
   * Mock Unsubscription Result - matches official RPC spec exactly
   */
  static getUnsubscriptionResult(): UnsubscriptionResult {
    return {
      unsubscribed: true,
      topics: ['camera.connected']
    };
  }

  /**
   * Mock Subscription Stats Result - matches official RPC spec exactly
   */
  static getSubscriptionStatsResult(): SubscriptionStatsResult {
    return {
      global_stats: {
        total_subscriptions: 15,
        active_clients: 3,
        topic_counts: {
          'camera.connected': 2,
          'recording.start': 1,
          'recording.stop': 1
        }
      },
      client_topics: ['camera.connected', 'recording.start'],
      client_id: 'client_123'
    };
  }

  /**
   * Mock External Stream Discovery Result - matches official RPC spec exactly
   */
  static getExternalStreamDiscoveryResult(): ExternalStreamDiscoveryResult {
    return {
      discovered_streams: [
        {
          url: 'rtsp://192.168.42.10:5554/subject',
          type: 'skydio_stanag4609',
          name: 'Skydio_EO_192.168.42.10_eo_/subject',
          status: 'DISCOVERED',
          discovered_at: '2025-01-15T14:30:00Z',
          last_seen: '2025-01-15T14:30:00Z',
          capabilities: {
            protocol: 'rtsp',
            format: 'stanag4609',
            source: 'skydio_uav',
            stream_type: 'eo',
            port: 5554,
            stream_path: '/subject',
            codec: 'h264',
            metadata: 'klv_mpegts'
          }
        }
      ],
      skydio_streams: [
        {
          url: 'rtsp://192.168.42.10:5554/subject',
          type: 'skydio_stanag4609',
          name: 'Skydio_EO_192.168.42.10_eo_/subject',
          status: 'DISCOVERED',
          discovered_at: '2025-01-15T14:30:00Z',
          last_seen: '2025-01-15T14:30:00Z',
          capabilities: {
            protocol: 'rtsp',
            format: 'stanag4609',
            source: 'skydio_uav',
            stream_type: 'eo',
            port: 5554,
            stream_path: '/subject',
            codec: 'h264',
            metadata: 'klv_mpegts'
          }
        }
      ],
      generic_streams: [],
      scan_timestamp: '2025-01-15T14:30:00Z',
      total_found: 1,
      discovery_options: {
        skydio_enabled: true,
        generic_enabled: false,
        force_rescan: false,
        include_offline: false
      },
      scan_duration: '2.5s',
      errors: []
    };
  }

  /**
   * Mock External Stream Add Result - matches official RPC spec exactly
   */
  static getExternalStreamAddResult(): ExternalStreamAddResult {
    return {
      stream_url: 'rtsp://192.168.42.15:5554/subject',
      stream_name: 'Skydio_UAV_15',
      stream_type: 'skydio_stanag4609',
      status: 'ADDED',
      timestamp: '2025-01-15T14:30:00Z'
    };
  }

  /**
   * Mock External Stream Remove Result - matches official RPC spec exactly
   */
  static getExternalStreamRemoveResult(): ExternalStreamRemoveResult {
    return {
      stream_url: 'rtsp://192.168.42.15:5554/subject',
      status: 'REMOVED',
      timestamp: '2025-01-15T14:30:00Z'
    };
  }

  /**
   * Mock External Streams List Result - matches official RPC spec exactly
   */
  static getExternalStreamsListResult(): ExternalStreamsListResult {
    return {
      external_streams: [
        {
          url: 'rtsp://192.168.42.10:5554/subject',
          type: 'skydio_stanag4609',
          name: 'Skydio_EO_192.168.42.10_eo_/subject',
          status: 'DISCOVERED',
          discovered_at: '2025-01-15T14:30:00Z',
          last_seen: '2025-01-15T14:30:00Z',
          capabilities: {
            protocol: 'rtsp',
            format: 'stanag4609',
            source: 'skydio_uav',
            stream_type: 'eo',
            port: 5554,
            stream_path: '/subject',
            codec: 'h264',
            metadata: 'klv_mpegts'
          }
        }
      ],
      skydio_streams: [
        {
          url: 'rtsp://192.168.42.10:5554/subject',
          type: 'skydio_stanag4609',
          name: 'Skydio_EO_192.168.42.10_eo_/subject',
          status: 'DISCOVERED',
          discovered_at: '2025-01-15T14:30:00Z',
          last_seen: '2025-01-15T14:30:00Z',
          capabilities: {
            protocol: 'rtsp',
            format: 'stanag4609',
            source: 'skydio_uav',
            stream_type: 'eo',
            port: 5554,
            stream_path: '/subject',
            codec: 'h264',
            metadata: 'klv_mpegts'
          }
        }
      ],
      generic_streams: [],
      total_count: 1,
      timestamp: '2025-01-15T14:30:00Z'
    };
  }

  /**
   * Mock Discovery Interval Set Result - matches official RPC spec exactly
   */
  static getDiscoveryIntervalSetResult(): DiscoveryIntervalSetResult {
    return {
      scan_interval: 300,
      status: 'UPDATED',
      message: 'Discovery interval updated (restart required for changes to take effect)',
      timestamp: '2025-01-15T14:30:00Z'
    };
  }

  /**
   * Mock Retention Policy Set Result - matches official RPC spec exactly
   */
  static getRetentionPolicySetResult(): RetentionPolicySetResult {
    return {
      policy_type: 'age',
      max_age_days: 30,
      enabled: true,
      message: 'Retention policy configured successfully'
    };
  }

  /**
   * Mock Cleanup Result - matches official RPC spec exactly
   */
  static getCleanupResult(): CleanupResult {
    return {
      cleanup_executed: true,
      files_deleted: 15,
      space_freed: 10737418240,
      message: 'Cleanup completed successfully'
    };
  }

  /**
   * Mock Delete Result - matches official RPC spec exactly
   */
  static getDeleteResult(): DeleteResult {
    return {
      filename: 'camera0_2025-01-15_14-30-00',
      deleted: true,
      message: 'Recording file deleted successfully'
    };
  }

  // ============================================================================
  // CLIENT STATE MOCKS (NOT FROM API)
  // ============================================================================

  /**
   * Mock Connection State
   */
  static getConnectionState(): ConnectionState {
    return {
      status: 'connected',
      lastError: null,
      reconnectAttempts: 0,
      lastConnected: '2025-01-15T14:30:00Z'
    };
  }

  /**
   * Mock Auth State
   */
  static getAuthState(): AuthState {
    return {
      token: 'mock-jwt-token',
      role: 'operator',
      session_id: '550e8400-e29b-41d4-a716-446655440000',
      isAuthenticated: true,
      expires_at: '2025-01-16T14:30:00.000Z',
      permissions: ['view', 'control']
    };
  }

  /**
   * Mock Server State
   */
  static getServerState(): ServerState {
    return {
      info: this.getServerInfo(),
      status: this.getSystemStatus(),
      storage: this.getStorageInfo(),
      loading: false,
      error: null,
      lastUpdated: '2025-01-15T14:30:00Z'
    };
  }

  // ============================================================================
  // SERVICE MOCKS (JEST MOCK FUNCTIONS)
  // ============================================================================

  /**
   * Mock Device Service
   */
  static createMockDeviceService() {
    return {
      getCameraList: () => Promise.resolve(this.getCameraListResult()),
      getCameraStatus: () => Promise.resolve(this.getCameraStatusResult()),
      getCameraCapabilities: () => Promise.resolve(this.getCameraCapabilitiesResult()),
      getStreamUrl: () => Promise.resolve(this.getStreamUrlResult()),
      getStreamStatus: () => Promise.resolve(this.getStreamStatusResult()),
      getStreams: () => Promise.resolve(this.getStreamsListResult())
    };
  }

  /**
   * Mock File Service
   */
  static createMockFileService() {
    return {
      listRecordings: () => Promise.resolve(this.getFileListResult()),
      listSnapshots: () => Promise.resolve(this.getFileListResult()),
      getRecordingInfo: () => Promise.resolve(this.getRecordingInfo()),
      getSnapshotInfo: () => Promise.resolve(this.getSnapshotInfo()),
      deleteRecording: () => Promise.resolve(this.getDeleteResult()),
      deleteSnapshot: () => Promise.resolve(this.getDeleteResult())
    };
  }

  /**
   * Mock Recording Service
   */
  static createMockRecordingService() {
    return {
      takeSnapshot: () => Promise.resolve(this.getSnapshotResult()),
      startRecording: () => Promise.resolve(this.getRecordingStartResult()),
      stopRecording: () => Promise.resolve(this.getRecordingStopResult())
    };
  }

  /**
   * Mock Connection Service
   */
  static createMockConnectionService() {
    return {
      connect: () => Promise.resolve(),
      disconnect: () => Promise.resolve(),
      isConnected: () => true,
      getConnectionState: () => this.getConnectionState()
    };
  }

  /**
   * Mock Auth Service
   */
  static createMockAuthService() {
    return {
      authenticate: () => Promise.resolve(this.getAuthenticateResult()),
      logout: () => Promise.resolve(),
      getAuthState: () => this.getAuthState(),
      isAuthenticated: () => true
    };
  }

  /**
   * Mock Server Service
   */
  static createMockServerService() {
    return {
      getServerInfo: () => Promise.resolve(this.getServerInfo()),
      getSystemStatus: () => Promise.resolve(this.getSystemStatus()),
      getStorageInfo: () => Promise.resolve(this.getStorageInfo()),
      getMetrics: () => Promise.resolve(this.getMetricsResult())
    };
  }

}