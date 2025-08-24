import type { AppConfig, ConfigValidationResult as ValidationResult } from '../types/camera';

/**
 * Configuration Manager Service
 * 
 * Manages environment variables, configuration validation, and dynamic
 * configuration updates for the application.
 */
class ConfigurationManagerService {
  private currentConfig: AppConfig;
  private configCallbacks: Set<(config: AppConfig) => void> = new Set();
  private validationCallbacks: Set<(errors: string[]) => void> = new Set();

  constructor() {
    this.currentConfig = this.getDefaultConfiguration();
    this.loadConfiguration();
  }

  /**
   * Get recording rotation minutes from environment
   */
  getRecordingRotationMinutes(): number {
    const envValue = import.meta.env.VITE_RECORDING_ROTATION_MINUTES;
    if (envValue) {
      const parsed = parseInt(envValue, 10);
      if (!isNaN(parsed) && parsed > 0) {
        return parsed;
      }
    }
    return this.currentConfig.recording.rotationMinutes;
  }

  /**
   * Get storage warning percentage from environment
   */
  getStorageWarnPercent(): number {
    const envValue = import.meta.env.VITE_STORAGE_WARN_PERCENT;
    if (envValue) {
      const parsed = parseInt(envValue, 10);
      if (!isNaN(parsed) && parsed >= 0 && parsed <= 100) {
        return parsed;
      }
    }
    return this.currentConfig.storage.warnPercent;
  }

  /**
   * Get storage block percentage from environment
   */
  getStorageBlockPercent(): number {
    const envValue = import.meta.env.VITE_STORAGE_BLOCK_PERCENT;
    if (envValue) {
      const parsed = parseInt(envValue, 10);
      if (!isNaN(parsed) && parsed >= 0 && parsed <= 100) {
        return parsed;
      }
    }
    return this.currentConfig.storage.blockPercent;
  }

  /**
   * Get WebSocket URL from environment
   */
  getWebSocketUrl(): string {
    return import.meta.env.VITE_WEBSOCKET_URL || this.currentConfig.connection.websocketUrl;
  }

  /**
   * Get health endpoint URL from environment
   */
  getHealthUrl(): string {
    return import.meta.env.VITE_HEALTH_URL || this.currentConfig.connection.healthUrl;
  }

  /**
   * Get API timeout from environment
   */
  getApiTimeout(): number {
    const envValue = import.meta.env.VITE_API_TIMEOUT;
    if (envValue) {
      const parsed = parseInt(envValue, 10);
      if (!isNaN(parsed) && parsed > 0) {
        return parsed;
      }
    }
    return this.currentConfig.connection.timeout;
  }

  /**
   * Get log level from environment
   */
  getLogLevel(): string {
    return import.meta.env.VITE_LOG_LEVEL || this.currentConfig.system.logLevel;
  }

  /**
   * Validate configuration
   */
  validateConfiguration(): ValidationResult {
    const errors: string[] = [];

    // Validate recording configuration
    const rotationMinutes = this.getRecordingRotationMinutes();
    if (rotationMinutes <= 0) {
      errors.push('Recording rotation minutes must be greater than 0');
    }

    // Validate storage configuration
    const warnPercent = this.getStorageWarnPercent();
    const blockPercent = this.getStorageBlockPercent();
    
    if (warnPercent < 0 || warnPercent > 100) {
      errors.push('Storage warning percentage must be between 0 and 100');
    }
    
    if (blockPercent < 0 || blockPercent > 100) {
      errors.push('Storage block percentage must be between 0 and 100');
    }
    
    if (warnPercent >= blockPercent) {
      errors.push('Storage warning percentage must be less than block percentage');
    }

    // Validate connection configuration
    const websocketUrl = this.getWebSocketUrl();
    const healthUrl = this.getHealthUrl();
    const timeout = this.getApiTimeout();
    
    if (!websocketUrl) {
      errors.push('WebSocket URL is required');
    }
    
    if (!healthUrl) {
      errors.push('Health URL is required');
    }
    
    if (timeout <= 0) {
      errors.push('API timeout must be greater than 0');
    }

    // Validate system configuration
    const logLevel = this.getLogLevel();
    const validLogLevels = ['debug', 'info', 'warn', 'error'];
    if (!validLogLevels.includes(logLevel.toLowerCase())) {
      errors.push(`Log level must be one of: ${validLogLevels.join(', ')}`);
    }

    const result: ValidationResult = {
      isValid: errors.length === 0,
      errors: errors,
      warnings: [],
      config: this.currentConfig
    };

    // Notify validation callbacks
    this.notifyValidationCallbacks(errors);

    return result;
  }

  /**
   * Get configuration errors
   */
  getConfigurationErrors(): string[] {
    const validation = this.validateConfiguration();
    if (validation.isValid) {
      return [];
    }
    return validation.errors;
  }

  /**
   * Update configuration
   */
  updateConfiguration(config: Partial<AppConfig>): void {
    this.currentConfig = { ...this.currentConfig, ...config };
    this.notifyConfigCallbacks(this.currentConfig);
    
    // Validate after update
    this.validateConfiguration();
  }

  /**
   * Reload configuration from environment
   */
  async reloadConfiguration(): Promise<void> {
    this.loadConfiguration();
    this.notifyConfigCallbacks(this.currentConfig);
    
    // Validate after reload
    this.validateConfiguration();
  }

  /**
   * Get default configuration
   */
  getDefaultConfiguration(): AppConfig {
    return {
      recording: {
        rotation_minutes: 60,
        rotationMinutes: 60,
        default_format: 'mp4',
        auto_rotation: true,
        maxFilesPerCamera: 10,
        autoDelete: true
      },
      storage: {
        warn_percent: 80,
        warnPercent: 80,
        block_percent: 95,
        blockPercent: 95,
        critical_percent: 90,
        monitoring_enabled: true,
        maxUsagePercent: 90
      },
      connection: {
        websocketUrl: 'ws://localhost:8002/ws',
        healthUrl: 'http://localhost:8003',
        timeout: 30000
      },
      system: {
        logLevel: 'info',
        autoRefresh: true,
        refreshInterval: 5000
      },
      environment: {}
    };
  }

  /**
   * Get current configuration
   */
  getCurrentConfiguration(): AppConfig {
    return { ...this.currentConfig };
  }

  /**
   * Get environment variables summary
   */
  getEnvironmentVariables(): Record<string, string> {
    return {
      VITE_RECORDING_ROTATION_MINUTES: import.meta.env.VITE_RECORDING_ROTATION_MINUTES || 'not set',
      VITE_STORAGE_WARN_PERCENT: import.meta.env.VITE_STORAGE_WARN_PERCENT || 'not set',
      VITE_STORAGE_BLOCK_PERCENT: import.meta.env.VITE_STORAGE_BLOCK_PERCENT || 'not set',
      VITE_WEBSOCKET_URL: import.meta.env.VITE_WEBSOCKET_URL || 'not set',
      VITE_HEALTH_URL: import.meta.env.VITE_HEALTH_URL || 'not set',
      VITE_API_TIMEOUT: import.meta.env.VITE_API_TIMEOUT || 'not set',
      VITE_LOG_LEVEL: import.meta.env.VITE_LOG_LEVEL || 'not set'
    };
  }

  /**
   * Event handlers
   */
  onConfigurationChange(callback: (config: AppConfig) => void): void {
    this.configCallbacks.add(callback);
  }

  onValidationError(callback: (errors: string[]) => void): void {
    this.validationCallbacks.add(callback);
  }

  private notifyConfigCallbacks(config: AppConfig): void {
    this.configCallbacks.forEach(callback => callback(config));
  }

  private notifyValidationCallbacks(errors: string[]): void {
    this.validationCallbacks.forEach(callback => callback(errors));
  }

  /**
   * Load configuration from environment
   */
  private loadConfiguration(): void {
    this.currentConfig = {
      recording: {
        rotation_minutes: this.getRecordingRotationMinutes(),
        rotationMinutes: this.getRecordingRotationMinutes(),
        default_format: 'mp4',
        auto_rotation: true,
        maxFilesPerCamera: 10,
        autoDelete: true
      },
      storage: {
        warn_percent: this.getStorageWarnPercent(),
        warnPercent: this.getStorageWarnPercent(),
        block_percent: this.getStorageBlockPercent(),
        blockPercent: this.getStorageBlockPercent(),
        critical_percent: 90,
        monitoring_enabled: true,
        maxUsagePercent: 90
      },
      connection: {
        websocketUrl: this.getWebSocketUrl(),
        healthUrl: this.getHealthUrl(),
        timeout: this.getApiTimeout()
      },
      system: {
        logLevel: this.getLogLevel() as 'debug' | 'info' | 'warn' | 'error',
        autoRefresh: true,
        refreshInterval: 5000
      },
      environment: {}
    };
  }

  /**
   * Cleanup
   */
  cleanup(): void {
    this.configCallbacks.clear();
    this.validationCallbacks.clear();
  }
}

// Export singleton instance
export const configurationManagerService = new ConfigurationManagerService();
