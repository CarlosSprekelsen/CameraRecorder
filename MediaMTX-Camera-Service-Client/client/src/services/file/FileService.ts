import { BaseService } from '../base/BaseService';
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
export class FileService extends BaseService implements IFileCatalog, IFileActions {
  constructor(
    apiClient: IAPIClient,
    logger: LoggerService,
  ) {
    super(apiClient, logger);
    this.logInitialization('FileService');
  }

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
    return this.callWithLogging<FileListResult>('list_recordings', { limit, offset });
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
    return this.callWithLogging<FileListResult>('list_snapshots', { limit, offset });
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
    const response = await this.callWithLogging('get_recording_info', { filename }, `getRecordingInfo(${filename})`) as RecordingInfo;
    // Return API response directly as per authoritative specification
    return {
      filename: response.filename,
      file_size: response.file_size,
      created_time: response.created_time,
      download_url: response.download_url,
      duration: response.duration
    };
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
    const response = await this.callWithLogging('get_snapshot_info', { filename }, `getSnapshotInfo(${filename})`) as SnapshotInfo;
    // Return API response directly as per authoritative specification
    return {
      filename: response.filename,
      file_size: response.file_size,
      created_time: response.created_time,
      download_url: response.download_url
    };
  }

  /**
   * Delete recording file.
   * Implements delete_recording RPC method.
   */
  async deleteRecording(filename: string): Promise<DeleteResult> {
    return this.callWithLogging('delete_recording', { filename }, `deleteRecording(${filename})`) as Promise<DeleteResult>;
  }

  /**
   * Delete snapshot file.
   * Implements delete_snapshot RPC method.
   */
  async deleteSnapshot(filename: string): Promise<DeleteResult> {
    return this.callWithLogging('delete_snapshot', { filename }, `deleteSnapshot(${filename})`) as Promise<DeleteResult>;
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
    const params: Record<string, unknown> = {
      policy_type: policyType,
      enabled
    };
    if (maxAgeDays !== undefined) params.max_age_days = maxAgeDays;
    if (maxSizeGb !== undefined) params.max_size_gb = maxSizeGb;
    
    return this.callWithLogging('set_retention_policy', params, 'setRetentionPolicy') as Promise<RetentionPolicySetResult>;
  }

  /**
   * Cleanup old files based on retention criteria.
   * Implements cleanup_old_files RPC method.
   */
  async cleanupOldFiles(): Promise<CleanupResult> {
    return this.callWithLogging('cleanup_old_files', {}, 'cleanupOldFiles') as Promise<CleanupResult>;
  }
}
