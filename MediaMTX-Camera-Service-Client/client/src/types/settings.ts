/**
 * Application Settings Types
 * Defines all configurable settings for the camera management application
 */

/**
 * Connection settings
 */
export interface ConnectionSettings {
  websocketUrl: string;
  healthUrl: string;
  httpBaseUrl: string;
  connectionTimeout: number;
  requestTimeout: number;
  reconnectInterval: number;
  maxReconnectAttempts: number;
  pingInterval: number;
  heartbeatInterval: number;
  qualityThreshold: number;
  enableHttpFallback: boolean;
  pollingInterval: number;
  maxPollingDuration: number;
  enableMetrics: boolean;
  enableCircuitBreaker: boolean;
  circuitBreakerThreshold: number;
  circuitBreakerTimeout: number;
}

/**
 * Recording settings
 */
export interface RecordingSettings {
  defaultFormat: 'mp4' | 'mkv';
  defaultQuality: number;
  defaultDuration: number | null; // null for unlimited
  maxDuration: number; // maximum allowed duration in seconds
  defaultFrameRate: number; // fps
  defaultBitrate: number; // kbps
  autoStartOnConnect: boolean;
  maxFileSize: number; // in MB
  storagePath: string;
  storageDirectory: string; // storage directory path
  maxStorageSize: number; // maximum storage size in GB
  autoCleanupAge: number; // auto cleanup age in days
  enableAutoCleanup: boolean;
  enableAudio: boolean;
  enableWatermark: boolean;
  watermarkText: string;
  enableCompression: boolean;
}

/**
 * Snapshot settings
 */
export interface SnapshotSettings {
  defaultFormat: 'jpg' | 'png' | 'jpeg' | 'bmp' | 'webp';
  jpegQuality: number; // quality setting for JPEG images (1-100)
  defaultQuality: number; // general quality setting
  defaultWidth: number; // default image width in pixels
  defaultHeight: number; // default image height in pixels
  autoSave: boolean;
  storagePath: string;
  enableTimestamp: boolean; // add timestamp to images
  enableWatermark: boolean; // add watermark to images
}

/**
 * UI settings
 */
export interface UISettings {
  theme: 'light' | 'dark' | 'auto';
  language: string;
  autoRefresh: boolean;
  refreshInterval: number;
  showNotifications: boolean;
  notificationDuration: number;
  compactMode: boolean;
}

/**
 * Notification settings
 */
export interface NotificationSettings {
  enabled: boolean;
  soundEnabled: boolean;
  desktopNotifications: boolean;
  emailNotifications: boolean;
  emailAddress: string;
  notificationTypes: {
    cameraStatus: boolean;
    recordingEvents: boolean;
    systemAlerts: boolean;
    fileOperations: boolean;
  };
}

/**
 * Security settings
 */
export interface SecuritySettings {
  autoLogout: boolean;
  sessionTimeout: number;
  rememberCredentials: boolean;
  requireReauthForSensitive: boolean;
}

/**
 * Performance settings
 */
export interface PerformanceSettings {
  enableCaching: boolean;
  cacheSize: number;
  enableCompression: boolean;
  maxConcurrentDownloads: number;
  enableBackgroundSync: boolean;
}

/**
 * Complete application settings
 */
export interface AppSettings {
  connection: ConnectionSettings;
  recording: RecordingSettings;
  snapshot: SnapshotSettings;
  ui: UISettings;
  notifications: NotificationSettings;
  security: SecuritySettings;
  performance: PerformanceSettings;
  version: string;
  lastUpdated: Date;
}

/**
 * Settings validation result
 */
export interface SettingsValidation {
  isValid: boolean;
  errors: string[];
  warnings: string[];
}

/**
 * Settings change event
 */
export interface SettingsChangeEvent {
  category: keyof AppSettings;
  key: string;
  oldValue: unknown;
  newValue: unknown;
  timestamp: Date;
}

/**
 * Default settings values
 */
export const DEFAULT_SETTINGS: AppSettings = {
  connection: {
    websocketUrl: 'ws://localhost:8002/ws',
    healthUrl: 'http://localhost:8003',
    httpBaseUrl: 'http://localhost:8003',
    connectionTimeout: 10000,
    requestTimeout: 15000,
    reconnectInterval: 5000,
    maxReconnectAttempts: 5,
    pingInterval: 30000,
    heartbeatInterval: 30000,
    qualityThreshold: 80,
    enableHttpFallback: true,
    pollingInterval: 5000,
    maxPollingDuration: 30000,
    enableMetrics: true,
    enableCircuitBreaker: true,
    circuitBreakerThreshold: 5,
    circuitBreakerTimeout: 60000,
  },
  recording: {
    defaultFormat: 'mp4',
    defaultQuality: 85,
    defaultDuration: null, // unlimited
    maxDuration: 3600, // 1 hour max
    defaultFrameRate: 30,
    defaultBitrate: 2000, // 2 Mbps
    autoStartOnConnect: false,
    maxFileSize: 1024, // 1GB
    storagePath: '/downloads',
    storageDirectory: '/downloads/recordings',
    maxStorageSize: 10, // 10GB
    autoCleanupAge: 30, // 30 days
    enableAutoCleanup: true,
    enableAudio: true,
    enableWatermark: false,
    watermarkText: 'Camera Recording',
    enableCompression: true,
  },
  snapshot: {
    defaultFormat: 'jpg',
    jpegQuality: 85,
    defaultQuality: 80,
    defaultWidth: 1920,
    defaultHeight: 1080,
    autoSave: true,
    storagePath: '/downloads/snapshots',
    enableTimestamp: true,
    enableWatermark: false,
  },
  ui: {
    theme: 'auto',
    language: 'en',
    autoRefresh: true,
    refreshInterval: 10000,
    showNotifications: true,
    notificationDuration: 5000,
    compactMode: false,
  },
  notifications: {
    enabled: true,
    soundEnabled: true,
    desktopNotifications: true,
    emailNotifications: false,
    emailAddress: '',
    notificationTypes: {
      cameraStatus: true,
      recordingEvents: true,
      systemAlerts: true,
      fileOperations: true,
    },
  },
  security: {
    autoLogout: true,
    sessionTimeout: 3600000, // 1 hour
    rememberCredentials: false,
    requireReauthForSensitive: true,
  },
  performance: {
    enableCaching: true,
    cacheSize: 100, // MB
    enableCompression: true,
    maxConcurrentDownloads: 3,
    enableBackgroundSync: true,
  },
  version: '1.0.0',
  lastUpdated: new Date(),
};

/**
 * Settings categories for UI organization
 */
export const SETTINGS_CATEGORIES = {
  connection: {
    title: 'Connection',
    description: 'WebSocket and health endpoint configuration',
    icon: '🔌',
  },
  recording: {
    title: 'Recording',
    description: 'Video recording preferences and settings',
    icon: '🎬',
  },
  snapshot: {
    title: 'Snapshots',
    description: 'Photo capture settings and preferences',
    icon: '📸',
  },
  ui: {
    title: 'Interface',
    description: 'User interface and display preferences',
    icon: '🎨',
  },
  notifications: {
    title: 'Notifications',
    description: 'Notification and alert preferences',
    icon: '🔔',
  },
  security: {
    title: 'Security',
    description: 'Authentication and security settings',
    icon: '🔒',
  },
  performance: {
    title: 'Performance',
    description: 'Performance and optimization settings',
    icon: '⚡',
  },
} as const;

export type SettingsCategory = keyof typeof SETTINGS_CATEGORIES;
