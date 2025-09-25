// Service Layer - Interface Contracts
// Implements the interfaces defined in architecture section 5.3

import {
  JsonRpcNotification,
  SystemStatus,
  StorageInfo,
  ServerInfo,
  SnapshotResult,
  RecordingResult,
  SubscriptionResult,
  UnsubscriptionResult,
  SubscriptionStatsResult,
  MetricsResult,
} from '../../types/api';
import { Camera, StreamInfo } from '../../stores/device/deviceStore';

// I.Discovery: list devices and stream links
export interface IDiscovery {
  getCameraList(): Promise<Camera[]>;
  getStreams(): Promise<StreamInfo[]>;
  getStreamUrl(device: string): Promise<string | null>;
}

// I.Command: snapshot and recording operations
export interface ICommand {
  takeSnapshot(device: string, filename?: string): Promise<SnapshotResult>;
  startRecording(device: string, duration?: number, format?: string): Promise<RecordingResult>;
  stopRecording(device: string): Promise<RecordingResult>;
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
