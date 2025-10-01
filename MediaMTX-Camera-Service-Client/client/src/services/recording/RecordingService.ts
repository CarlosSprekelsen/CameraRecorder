import { BaseService } from '../base/BaseService';
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
export class RecordingService extends BaseService implements ICommand {
  constructor(
    apiClient: IAPIClient,
    logger: LoggerService,
  ) {
    super(apiClient, logger);
    this.logInitialization('RecordingService');
  }

  async takeSnapshot(device: string, filename?: string): Promise<SnapshotResult> {
    return this.callWithLogging('take_snapshot', { device, filename }) as Promise<SnapshotResult>;
  }

  async startRecording(
    device: string,
    duration?: number,
    format?: string,
  ): Promise<RecordingStartResult> {
    return this.callWithLogging('start_recording', { device, duration, format }) as Promise<RecordingStartResult>;
  }

  async stopRecording(device: string): Promise<RecordingStopResult> {
    return this.callWithLogging('stop_recording', { device }) as Promise<RecordingStopResult>;
  }
}
