// Service Layer - Interface Contracts
// Implements the interfaces defined in architecture section 5.3

import { JsonRpcNotification } from '../../types/api';

// I.Discovery: list devices and stream links
export interface IDiscovery {
  getCameraList(): Promise<any>;
  getStreams(): Promise<any>;
  getStreamUrl(device: string): Promise<any>;
}

// I.Command: snapshot and recording operations
export interface ICommand {
  takeSnapshot(device: string, filename?: string): Promise<any>;
  startRecording(device: string, duration?: number, format?: string): Promise<any>;
  stopRecording(device: string): Promise<any>;
}

// I.FileCatalog: file listing and information
export interface IFileCatalog {
  listRecordings(limit?: number, offset?: number): Promise<any>;
  listSnapshots(limit?: number, offset?: number): Promise<any>;
  getRecordingInfo(filename: string): Promise<any>;
  getSnapshotInfo(filename: string): Promise<any>;
}

// I.FileActions: file operations
export interface IFileActions {
  downloadFile(url: string): void;
  deleteRecording(filename: string): Promise<any>;
  deleteSnapshot(filename: string): Promise<any>;
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
