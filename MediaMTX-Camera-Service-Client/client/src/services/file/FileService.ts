import { IAPIClient } from '../abstraction/IAPIClient';
import { LoggerService } from '../logger/LoggerService';
import { IFileCatalog, IFileActions } from '../interfaces/ServiceInterfaces';
import { FileListResult, RecordingInfo, SnapshotInfo, DeleteResult, RetentionPolicySetResult, CleanupResult } from '../../types/api';

/**
 * File Service - File management and operations
 *
 * Implements IFileCatalog and IFileActions interfaces for comprehensive file management.
 * Provides methods for listing, downloading, and deleting recordings and snapshots
 * with pagination support and server-provided download URLs.
 *
 * @class FileService
 * @implements {IFileCatalog} {@link IFileActions}
 *
 * @example
 * ```typescript
 * const fileService = new FileService(apiClient, logger);
 *
 * // List recordings with pagination
 * const recordings = await fileService.listRecordings(10, 0);
 *
 * // Get file information
 * const info = await fileService.getRecordingInfo('recording.mp4');
 *
 * // Download file
 * await fileService.downloadFile(info.download_url, 'recording.mp4');
 *
 * // Delete file
 * await fileService.deleteRecording('recording.mp4');
 * ```
 *
 * @see {@link ../interfaces/ServiceInterfaces#IFileCatalog} IFileCatalog interface
 * @see {@link ../interfaces/ServiceInterfaces#IFileActions} IFileActions interface
 * @see {@link ../../docs/architecture/client-architechture.md} Client Architecture
 */
export class FileService implements IFileCatalog, IFileActions {
  constructor(
    private apiClient: IAPIClient,
    private logger: LoggerService,
  ) {}

  /**
   * List recordings with pagination.
   * Implements list_recordings RPC method.
   */
  async listRecordings(
    limit: number,
    offset: number,
  ): Promise<{
    files: Array<{
      filename: string;
      file_size: number;
      modified_time: string;
      download_url: string;
    }>;
    total: number;
  }> {
    try {
      this.logger.info(`Listing recordings: limit=${limit}, offset=${offset}`);
      const response = await this.apiClient.call('list_recordings', { limit, offset }) as FileListResult;
      this.logger.info(`Found ${response.files?.length || 0} recordings`);
      return response;
    } catch (error) {
      this.logger.error(`Failed to list recordings`, error as Record<string, unknown>);
      throw error;
    }
  }

  /**
   * List snapshots with pagination.
   * Implements list_snapshots RPC method.
   */
  async listSnapshots(
    limit: number,
    offset: number,
  ): Promise<{
    files: Array<{
      filename: string;
      file_size: number;
      modified_time: string;
      download_url: string;
    }>;
    total: number;
  }> {
    try {
      this.logger.info(`Listing snapshots: limit=${limit}, offset=${offset}`);
      const response = await this.apiClient.call('list_snapshots', { limit, offset }) as FileListResult;
      this.logger.info(`Found ${response.files?.length || 0} snapshots`);
      return response;
    } catch (error) {
      this.logger.error(`Failed to list snapshots`, error as Record<string, unknown>);
      throw error;
    }
  }

  /**
   * Get recording file information.
   * Implements get_recording_info RPC method.
   */
  async getRecordingInfo(filename: string): Promise<{
    filename: string;
    file_size: number;
    created_time: string;
    download_url: string;
    duration?: number;
    format?: string;
    device?: string;
  }> {
    try {
      this.logger.info(`Getting recording info for: ${filename}`);
      const response = await this.apiClient.call('get_recording_info', { filename }) as RecordingInfo;
      this.logger.info(`Recording info retrieved for ${filename}`);
      // Return API response directly as per authoritative specification
      return {
        filename: response.filename,
        file_size: response.file_size,
        created_time: response.created_time,
        download_url: response.download_url,
        duration: response.duration
      };
    } catch (error) {
      this.logger.error(`Failed to get recording info for ${filename}`, error as Record<string, unknown>);
      throw error;
    }
  }

  /**
   * Get snapshot file information.
   * Implements get_snapshot_info RPC method.
   */
  async getSnapshotInfo(filename: string): Promise<{
    filename: string;
    file_size: number;
    created_time: string;
    download_url: string;
    format?: string;
    device?: string;
  }> {
    try {
      this.logger.info(`Getting snapshot info for: ${filename}`);
      const response = await this.apiClient.call('get_snapshot_info', { filename }) as SnapshotInfo;
      this.logger.info(`Snapshot info retrieved for ${filename}`);
      // Return API response directly as per authoritative specification
      return {
        filename: response.filename,
        file_size: response.file_size,
        created_time: response.created_time,
        download_url: response.download_url
      };
    } catch (error) {
      this.logger.error(`Failed to get snapshot info for ${filename}`, error as Record<string, unknown>);
      throw error;
    }
  }

  /**
   * Delete recording file.
   * Implements delete_recording RPC method.
   */
  async deleteRecording(filename: string): Promise<DeleteResult> {
    try {
      this.logger.info(`Deleting recording: ${filename}`);
      const response = await this.apiClient.call('delete_recording', { filename }) as DeleteResult;
      this.logger.info(`Recording deleted: ${filename}`);
      return response;
    } catch (error) {
      this.logger.error(`Failed to delete recording ${filename}`, error as Record<string, unknown>);
      throw error;
    }
  }

  /**
   * Delete snapshot file.
   * Implements delete_snapshot RPC method.
   */
  async deleteSnapshot(filename: string): Promise<DeleteResult> {
    try {
      this.logger.info(`Deleting snapshot: ${filename}`);
      const response = await this.apiClient.call('delete_snapshot', { filename }) as DeleteResult;
      this.logger.info(`Snapshot deleted: ${filename}`);
      return response;
    } catch (error) {
      this.logger.error(`Failed to delete snapshot ${filename}`, error as Record<string, unknown>);
      throw error;
    }
  }

  /**
   * Download file via server-provided URL.
   * Implements download hand-off per architecture section 13.4.
   */
  async downloadFile(downloadUrl: string, filename: string): Promise<void> {
    try {
      this.logger.info(`Downloading file: ${filename}`);
      // Create temporary link and trigger download
      const link = document.createElement('a');
      link.href = downloadUrl;
      link.download = filename;
      link.target = '_blank';
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      this.logger.info(`Download initiated for: ${filename}`);
    } catch (error) {
      this.logger.error(`Failed to download file ${filename}`, error as Record<string, unknown>);
      throw error;
    }
  }

  /**
   * Set file retention policy for automatic cleanup.
   * Implements set_retention_policy RPC method.
   */
  async setRetentionPolicy(
    policyType: 'age' | 'size' | 'manual',
    enabled: boolean,
    maxAgeDays?: number,
    maxSizeGb?: number
  ): Promise<RetentionPolicySetResult> {
    try {
      this.logger.info('Setting retention policy', { policyType, enabled, maxAgeDays, maxSizeGb });
      const params: Record<string, unknown> = {
        policy_type: policyType,
        enabled
      };
      if (maxAgeDays !== undefined) params.max_age_days = maxAgeDays;
      if (maxSizeGb !== undefined) params.max_size_gb = maxSizeGb;
      
      const response = await this.apiClient.call('set_retention_policy', params) as RetentionPolicySetResult;
      this.logger.info('Retention policy set successfully');
      return response;
    } catch (error) {
      this.logger.error('Failed to set retention policy', error as Record<string, unknown>);
      throw error;
    }
  }

  /**
   * Cleanup old files based on retention criteria.
   * Implements cleanup_old_files RPC method.
   */
  async cleanupOldFiles(): Promise<CleanupResult> {
    try {
      this.logger.info('Cleaning up old files');
      const response = await this.apiClient.call('cleanup_old_files') as CleanupResult;
      this.logger.info(`Cleanup completed: ${response.files_deleted} files deleted`);
      return response;
    } catch (error) {
      this.logger.error('Failed to cleanup old files', error as Record<string, unknown>);
      throw error;
    }
  }
}
