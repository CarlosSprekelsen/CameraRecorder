/**
 * Streaming Service - Camera stream management
 * 
 * Architecture requirement: "Modular architecture enabling independent feature development" (Section 1.2)
 * Implements IStreaming interface for streaming operations
 */

import { BaseService } from '../base/BaseService';
import { IAPIClient } from '../abstraction/IAPIClient';
import { LoggerService } from '../logger/LoggerService';
import { IStreaming } from '../interfaces/IStreaming';
import { StreamStartResult, StreamStopResult, StreamStatusResult } from '../../types/api';
import { validateCameraDeviceId } from '../../utils/validation';

export class StreamingService extends BaseService implements IStreaming {
  constructor(
    apiClient: IAPIClient,
    logger: LoggerService,
  ) {
    super(apiClient, logger);
    this.logInitialization('StreamingService');
  }

  /**
   * Start streaming for a specific camera device
   * Implements start_streaming RPC method
   */
  async startStreaming(device: string): Promise<StreamStartResult> {
    if (!validateCameraDeviceId(device)) {
      throw new Error(`Invalid device ID format: ${device}. Expected format: camera[0-9]+`);
    }

    return this.callWithLogging<StreamStartResult>('start_streaming', { device });
  }

  /**
   * Stop streaming for a specific camera device
   * Implements stop_streaming RPC method
   */
  async stopStreaming(device: string): Promise<StreamStopResult> {
    if (!validateCameraDeviceId(device)) {
      throw new Error(`Invalid device ID format: ${device}. Expected format: camera[0-9]+`);
    }

    return this.callWithLogging<StreamStopResult>('stop_streaming', { device });
  }

  /**
   * Get detailed status information for a specific camera stream
   * Implements get_stream_status RPC method
   */
  async getStreamStatus(device: string): Promise<StreamStatusResult> {
    if (!validateCameraDeviceId(device)) {
      throw new Error(`Invalid device ID format: ${device}. Expected format: camera[0-9]+`);
    }

    return this.callWithLogging<StreamStatusResult>('get_stream_status', { device });
  }
}
