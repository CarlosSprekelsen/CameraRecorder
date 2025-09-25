import { WebSocketService } from '../websocket/WebSocketService';
import { LoggerService } from '../logger/LoggerService';
import { IFileCatalog, IFileActions } from '../interfaces/ServiceInterfaces';

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
 * const fileService = new FileService(wsService, logger);
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
    private wsService: WebSocketService,
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
      const response = await this.wsService.sendRPC('list_recordings', { limit, offset });
      this.logger.info(`Found ${response.files?.length || 0} recordings`);
      return response;
    } catch (error) {
      this.logger.error(`Failed to list recordings`, error as Error);
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
      const response = await this.wsService.sendRPC('list_snapshots', { limit, offset });
      this.logger.info(`Found ${response.files?.length || 0} snapshots`);
      return response;
    } catch (error) {
      this.logger.error(`Failed to list snapshots`, error as Error);
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
    modified_time: string;
    download_url: string;
    duration?: number;
    format?: string;
    device?: string;
  }> {
    try {
      this.logger.info(`Getting recording info for: ${filename}`);
      const response = await this.wsService.sendRPC('get_recording_info', { filename });
      this.logger.info(`Recording info retrieved for ${filename}`);
      return response;
    } catch (error) {
      this.logger.error(`Failed to get recording info for ${filename}`, error as Error);
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
    modified_time: string;
    download_url: string;
    format?: string;
    device?: string;
  }> {
    try {
      this.logger.info(`Getting snapshot info for: ${filename}`);
      const response = await this.wsService.sendRPC('get_snapshot_info', { filename });
      this.logger.info(`Snapshot info retrieved for ${filename}`);
      return response;
    } catch (error) {
      this.logger.error(`Failed to get snapshot info for ${filename}`, error as Error);
      throw error;
    }
  }

  /**
   * Delete recording file.
   * Implements delete_recording RPC method.
   */
  async deleteRecording(filename: string): Promise<{ success: boolean; message: string }> {
    try {
      this.logger.info(`Deleting recording: ${filename}`);
      const response = await this.wsService.sendRPC('delete_recording', { filename });
      this.logger.info(`Recording deleted: ${filename}`);
      return response;
    } catch (error) {
      this.logger.error(`Failed to delete recording ${filename}`, error as Error);
      throw error;
    }
  }

  /**
   * Delete snapshot file.
   * Implements delete_snapshot RPC method.
   */
  async deleteSnapshot(filename: string): Promise<{ success: boolean; message: string }> {
    try {
      this.logger.info(`Deleting snapshot: ${filename}`);
      const response = await this.wsService.sendRPC('delete_snapshot', { filename });
      this.logger.info(`Snapshot deleted: ${filename}`);
      return response;
    } catch (error) {
      this.logger.error(`Failed to delete snapshot ${filename}`, error as Error);
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
      this.logger.error(`Failed to download file ${filename}`, error as Error);
      throw error;
    }
  }
}
