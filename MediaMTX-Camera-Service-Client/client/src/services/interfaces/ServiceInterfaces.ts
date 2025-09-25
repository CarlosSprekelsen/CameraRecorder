/**
 * Service Layer - Interface Contracts
 * Implements the interfaces defined in architecture section 5.3
 *
 * @fileoverview Defines the core service interfaces for the MediaMTX Camera Service Client
 * @author MediaMTX Development Team
 * @version 1.0.0
 */

import {
  JsonRpcNotification,
  SystemStatus,
  StorageInfo,
  ServerInfo,
  SnapshotResult,
  RecordingStartResult,
  RecordingStopResult,
  SubscriptionResult,
  UnsubscriptionResult,
  SubscriptionStatsResult,
  MetricsResult,
} from '../../types/api';
import { Camera, StreamInfo } from '../../stores/device/deviceStore';

/**
 * Device Discovery Interface
 * Provides methods for discovering cameras and retrieving stream information
 *
 * @interface IDiscovery
 */
export interface IDiscovery {
  /**
   * Retrieves the list of available cameras
   *
   * @returns {Promise<Camera[]>} Array of camera objects with device information
   * @throws {Error} When discovery fails or connection is lost
   *
   * @example
   * ```typescript
   * const cameras = await discoveryService.getCameraList();
   * console.log(`Found ${cameras.length} cameras`);
   * ```
   */
  getCameraList(): Promise<Camera[]>;

  /**
   * Retrieves the list of active streams
   *
   * @returns {Promise<StreamInfo[]>} Array of stream information objects
   * @throws {Error} When stream discovery fails
   *
   * @example
   * ```typescript
   * const streams = await discoveryService.getStreams();
   * streams.forEach(stream => console.log(stream.url));
   * ```
   */
  getStreams(): Promise<StreamInfo[]>;

  /**
   * Gets the stream URL for a specific device
   *
   * @param {string} device - The device identifier
   * @returns {Promise<string | null>} The stream URL or null if not available
   * @throws {Error} When device is not found or stream is unavailable
   *
   * @example
   * ```typescript
   * const url = await discoveryService.getStreamUrl('camera-001');
   * if (url) {
   *   console.log(`Stream URL: ${url}`);
   * }
   * ```
   */
  getStreamUrl(device: string): Promise<string | null>;
}

/**
 * Command Operations Interface
 * Provides methods for camera control operations (snapshot and recording)
 *
 * @interface ICommand
 */
export interface ICommand {
  /**
   * Takes a snapshot from the specified camera
   *
   * @param {string} device - The camera device identifier
   * @param {string} [filename] - Optional custom filename for the snapshot
   * @returns {Promise<SnapshotResult>} Snapshot operation result with download URL
   * @throws {Error} When device is not found or snapshot fails
   *
   * @example
   * ```typescript
   * const result = await commandService.takeSnapshot('camera-001', 'my-snapshot.jpg');
   * if (result.success) {
   *   console.log(`Snapshot saved: ${result.filename}`);
   * }
   * ```
   */
  takeSnapshot(device: string, filename?: string): Promise<SnapshotResult>;

  /**
   * Starts recording from the specified camera
   *
   * @param {string} device - The camera device identifier
   * @param {number} [duration] - Optional recording duration in seconds
   * @param {string} [format] - Optional recording format (mp4, avi, etc.)
   * @returns {Promise<RecordingStartResult>} Recording operation result with status
   * @throws {Error} When device is not found or recording fails to start
   *
   * @example
   * ```typescript
   * const result = await commandService.startRecording('camera-001', 60, 'mp4');
   * if (result.success) {
   *   console.log(`Recording started: ${result.recording_id}`);
   * }
   * ```
   */
  startRecording(device: string, duration?: number, format?: string): Promise<RecordingStartResult>;

  /**
   * Stops recording from the specified camera
   *
   * @param {string} device - The camera device identifier
   * @returns {Promise<RecordingStopResult>} Recording stop result with final status
   * @throws {Error} When device is not found or no active recording exists
   *
   * @example
   * ```typescript
   * const result = await commandService.stopRecording('camera-001');
   * if (result.success) {
   *   console.log(`Recording stopped: ${result.status}`);
   * }
   * ```
   */
  stopRecording(device: string): Promise<RecordingStopResult>;
}

// I.FileCatalog: file listing and information
export interface IFileCatalog {
  listRecordings(
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
  }>;
  listSnapshots(
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
  }>;
  getRecordingInfo(filename: string): Promise<{
    filename: string;
    file_size: number;
    modified_time: string;
    download_url: string;
    duration?: number;
    format?: string;
    device?: string;
  }>;
  getSnapshotInfo(filename: string): Promise<{
    filename: string;
    file_size: number;
    modified_time: string;
    download_url: string;
    format?: string;
    device?: string;
  }>;
}

// I.FileActions: file operations
export interface IFileActions {
  downloadFile(downloadUrl: string, filename: string): Promise<void>;
  deleteRecording(filename: string): Promise<{ success: boolean; message: string }>;
  deleteSnapshot(filename: string): Promise<{ success: boolean; message: string }>;
}

// I.Status: system status and events
export interface IStatus {
  getStatus(): Promise<SystemStatus>;
  getStorageInfo(): Promise<StorageInfo>;
  getServerInfo(): Promise<ServerInfo>;
  getMetrics(): Promise<MetricsResult>;
  subscribeEvents(topics: string[], filters?: Record<string, unknown>): Promise<SubscriptionResult>;
  unsubscribeEvents(topics?: string[]): Promise<UnsubscriptionResult>;
  getSubscriptionStats(): Promise<SubscriptionStatsResult>;
}

// Notification handler interface
export interface INotificationHandler {
  onCameraStatusUpdate(notification: JsonRpcNotification): void;
  onRecordingStatusUpdate(notification: JsonRpcNotification): void;
  onSystemHealthUpdate(notification: JsonRpcNotification): void;
}
