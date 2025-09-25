import { WebSocketService } from './websocket';
import { errorRecoveryService } from './errorRecoveryService';
import { logger, loggers } from './loggerService';
import type {
  StorageInfo,
  StorageUsage,
  ThresholdStatus
} from '../types/camera';
import { RPC_METHODS, ERROR_CODES } from '../types/rpc';

/**
 * Storage Monitor Service
 * 
 * Monitors storage usage, manages thresholds, and provides validation
 * for storage-dependent operations.
 */
class StorageMonitorService {
  private wsService: WebSocketService | null = null;
  private currentStorageInfo: StorageInfo | null = null;
  private monitoringInterval: NodeJS.Timeout | null = null;
  private thresholdCallbacks: Set<(threshold: ThresholdStatus) => void> = new Set();
  private criticalCallbacks: Set<(status: StorageInfo) => void> = new Set();
  private updateCallbacks: Set<(info: StorageInfo) => void> = new Set();

  // Default thresholds (can be overridden by configuration)
  private warnThreshold = 80; // 80% usage
  private criticalThreshold = 95; // 95% usage

  /**
   * Set WebSocket service instance
   */
  setWebSocketService(service: WebSocketService): void {
    this.wsService = service;
  }

  /**
   * Get current storage information
   */
  async getStorageInfo(): Promise<StorageInfo> {
    if (!this.wsService) {
      throw new Error('WebSocket service not initialized');
    }

    try {
      const info = await errorRecoveryService.executeWithRetry(
        async () => {
          const response = await this.wsService!.call(RPC_METHODS.GET_STORAGE_INFO, {});
          return response as StorageInfo;
        },
        'getStorageInfo'
      );

      this.currentStorageInfo = info;
      this.notifyUpdateCallbacks(info);
      return info;
    } catch (error) {
      throw new Error(`Failed to get storage info: ${error}`);
    }
  }

  /**
   * Get storage usage statistics
   */
  async getStorageUsage(): Promise<StorageUsage> {
    const info = await this.getStorageInfo();
    
    return {
      total_space: info.total_space,
      available_space: info.available_space,
      used_space: info.total_space - info.available_space,
      usage_percent: ((info.total_space - info.available_space) / info.total_space) * 100
    };
  }

  /**
   * Check storage thresholds
   */
  async checkStorageThresholds(): Promise<ThresholdStatus> {
    const usage = await this.getStorageUsage();
    
    let currentStatus: 'normal' | 'warning' | 'critical' = 'normal';
    if (usage.usage_percent >= this.criticalThreshold) {
      currentStatus = 'critical';
    } else if (usage.usage_percent >= this.warnThreshold) {
      currentStatus = 'warning';
    }
    
    const status: ThresholdStatus = {
      warning_threshold: this.warnThreshold,
      critical_threshold: this.criticalThreshold,
      current_status: currentStatus
    };

    // Notify callbacks if thresholds are exceeded
    if (currentStatus === 'critical') {
      this.notifyCriticalCallbacks(this.currentStorageInfo!);
    } else if (currentStatus === 'warning') {
      this.notifyThresholdCallbacks(status);
    }

    return status;
  }

  /**
   * Check if storage is available for operations
   */
  async isStorageAvailable(): Promise<boolean> {
    try {
      const thresholds = await this.checkStorageThresholds();
      return !thresholds.is_critical;
    } catch (error) {
      logger.error('Storage availability check failed', error as Error, 'storageMonitor');
      return false;
    }
  }

  /**
   * Validate storage for recording operation
   */
  async validateStorageForRecording(): Promise<ValidationResult> {
    try {
      const thresholds = await this.checkStorageThresholds();
      
      if (thresholds.is_critical) {
        return {
          valid: false,
          reason: `Storage space is critical (${thresholds.usage_percent.toFixed(1)}% used). Recording blocked.`
        };
      }

      if (thresholds.is_warning) {
        return {
          valid: true,
          reason: `Storage space is low (${thresholds.usage_percent.toFixed(1)}% used). Proceed with caution.`
        };
      }

      return { valid: true, reason: 'Storage space is adequate for recording.' };
    } catch (error) {
      return {
        valid: false,
        reason: `Storage validation failed: ${error}`
      };
    }
  }

  /**
   * Validate storage for any operation
   */
  async validateStorageForOperation(operation: string): Promise<ValidationResult> {
    try {
      const thresholds = await this.checkStorageThresholds();
      
      if (thresholds.is_critical) {
        return {
          valid: false,
          reason: `Storage space is critical (${thresholds.usage_percent.toFixed(1)}% used). ${operation} blocked.`
        };
      }

      return { valid: true, reason: `Storage space is adequate for ${operation}.` };
    } catch (error) {
      return {
        valid: false,
        reason: `Storage validation failed for ${operation}: ${error}`
      };
    }
  }

  /**
   * Start storage monitoring
   */
  startStorageMonitoring(interval: number = 30000): void {
    if (this.monitoringInterval) {
      this.stopStorageMonitoring();
    }

    this.monitoringInterval = setInterval(async () => {
      try {
        await this.checkStorageThresholds();
      } catch (error) {
        logger.error('Storage monitoring error', error as Error, 'storageMonitor');
      }
    }, interval);
  }

  /**
   * Stop storage monitoring
   */
  stopStorageMonitoring(): void {
    if (this.monitoringInterval) {
      clearInterval(this.monitoringInterval);
      this.monitoringInterval = null;
    }
  }

  /**
   * Set storage thresholds
   */
  setThresholds(warnPercent: number, criticalPercent: number): void {
    this.warnThreshold = Math.max(0, Math.min(100, warnPercent));
    this.criticalThreshold = Math.max(this.warnThreshold, Math.min(100, criticalPercent));
  }

  /**
   * Get current thresholds
   */
  getThresholds(): { warn: number; critical: number } {
    return {
      warn: this.warnThreshold,
      critical: this.criticalThreshold
    };
  }

  /**
   * Format bytes to human readable format
   */
  formatBytes(bytes: number): string {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  }

  /**
   * Get threshold message
   */
  private getThresholdMessage(usagePercent: number): string {
    if (usagePercent >= this.criticalThreshold) {
      return `Storage space is critical (${usagePercent.toFixed(1)}% used). Operations may be blocked.`;
    } else if (usagePercent >= this.warnThreshold) {
      return `Storage space is low (${usagePercent.toFixed(1)}% used). Consider cleanup.`;
    } else {
      return `Storage space is adequate (${usagePercent.toFixed(1)}% used).`;
    }
  }

  /**
   * Event handlers
   */
  onStorageThresholdExceeded(callback: (threshold: ThresholdStatus) => void): void {
    this.thresholdCallbacks.add(callback);
  }

  onStorageCritical(callback: (status: StorageInfo) => void): void {
    this.criticalCallbacks.add(callback);
  }

  onStorageUpdate(callback: (info: StorageInfo) => void): void {
    this.updateCallbacks.add(callback);
  }

  private notifyThresholdCallbacks(threshold: ThresholdStatus): void {
    this.thresholdCallbacks.forEach(callback => callback(threshold));
  }

  private notifyCriticalCallbacks(status: StorageInfo): void {
    this.criticalCallbacks.forEach(callback => callback(status));
  }

  private notifyUpdateCallbacks(info: StorageInfo): void {
    this.updateCallbacks.forEach(callback => callback(info));
  }

  /**
   * Handle storage status updates from WebSocket
   */
  handleStorageStatusUpdate(storageInfo: StorageInfo): void {
    this.currentStorageInfo = storageInfo;
    this.notifyUpdateCallbacks(storageInfo);
    
    // Check thresholds after update
    this.checkStorageThresholds().catch(error => {
      logger.error('Threshold check failed after update', error as Error, 'storageMonitor');
    });
  }

  /**
   * Cleanup
   */
  cleanup(): void {
    this.stopStorageMonitoring();
    this.thresholdCallbacks.clear();
    this.criticalCallbacks.clear();
    this.updateCallbacks.clear();
    this.currentStorageInfo = null;
  }
}

// Export singleton instance
export const storageMonitorService = new StorageMonitorService();
