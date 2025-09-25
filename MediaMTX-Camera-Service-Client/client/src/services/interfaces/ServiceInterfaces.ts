// Service Layer - Interface Contracts
// Implements the interfaces defined in architecture section 5.3

import { JsonRpcNotification } from '../../types/api';
import { Camera, StreamInfo } from '../../stores/device/deviceStore';

// I.Discovery: list devices and stream links
export interface IDiscovery {
  getCameraList(): Promise<Camera[]>;
  getStreams(): Promise<StreamInfo[]>;
  getStreamUrl(device: string): Promise<string | null>;
}

// I.Command: snapshot and recording operations
export interface ICommand {
  takeSnapshot(device: string, filename?: string): Promise<any>;
  startRecording(device: string, duration?: number, format?: string): Promise<any>;
  stopRecording(device: string): Promise<any>;
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
  getStatus(): Promise<any>;
  getStorageInfo(): Promise<any>;
  getServerInfo(): Promise<any>;
  getMetrics(): Promise<any>;
  subscribeEvents(topics: string[], filters?: any): Promise<any>;
  unsubscribeEvents(topics?: string[]): Promise<any>;
  getSubscriptionStats(): Promise<any>;
}

// Notification handler interface
export interface INotificationHandler {
  onCameraStatusUpdate(notification: JsonRpcNotification): void;
  onRecordingStatusUpdate(notification: JsonRpcNotification): void;
  onSystemHealthUpdate(notification: JsonRpcNotification): void;
}
