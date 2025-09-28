import { APIClient } from '../abstraction/APIClient';
import { LoggerService } from '../logger/LoggerService';
import { SnapshotResult, RecordingStartResult, RecordingStopResult } from '../../types/api';
import { ICommand } from '../interfaces/ServiceInterfaces';

/**
 * Recording Service - Camera control operations
 *
 * Implements the ICommand interface for camera control operations including
 * snapshots and recording management. Provides methods for taking snapshots,
 * starting/stopping recordings with optional duration and format parameters.
 *
 * @class RecordingService
 * @implements {ICommand}
 *
 * @example
 * ```typescript
 * const recordingService = new RecordingService(wsService, logger);
 *
 * // Take a snapshot
 * const snapshot = await recordingService.takeSnapshot('camera-001', 'my-snapshot.jpg');
 *
 * // Start recording with duration
 * const recording = await recordingService.startRecording('camera-001', 60, 'mp4');
 *
 * // Stop recording
 * await recordingService.stopRecording('camera-001');
 * ```
 *
 * @see {@link ../interfaces/ServiceInterfaces#ICommand} ICommand interface
 * @see {@link ../../docs/architecture/client-architechture.md} Client Architecture
 */
export class RecordingService implements ICommand {
  constructor(
    private apiClient: APIClient,
    private logger: LoggerService,
  ) {}

  async takeSnapshot(device: string, filename?: string): Promise<SnapshotResult> {
    try {
      this.logger.info('take_snapshot request', { device, filename });
      return await this.apiClient.call('take_snapshot', { device, filename });
    } catch (error) {
      this.logger.error('take_snapshot failed', error as Record<string, unknown>);
      throw error;
    }
  }

  async startRecording(
    device: string,
    duration?: number,
    format?: string,
  ): Promise<RecordingStartResult> {
    try {
      this.logger.info('start_recording request', { device, duration, format });
      return await this.apiClient.call('start_recording', { device, duration, format });
    } catch (error) {
      this.logger.error('start_recording failed', error as Record<string, unknown>);
      throw error;
    }
  }

  async stopRecording(device: string): Promise<RecordingStopResult> {
    try {
      this.logger.info('stop_recording request', { device });
      return await this.apiClient.call('stop_recording', { device });
    } catch (error) {
      this.logger.error('stop_recording failed', error as Record<string, unknown>);
      throw error;
    }
  }
}
