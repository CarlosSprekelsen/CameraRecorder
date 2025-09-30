/**
 * Streaming Service - Camera stream management
 * 
 * Architecture requirement: "Modular architecture enabling independent feature development" (Section 1.2)
 * Implements IStreaming interface for streaming operations
 */

import { APIClient } from '../abstraction/APIClient';
import { LoggerService } from '../logger/LoggerService';
import { IStreaming } from '../interfaces/IStreaming';
import { StreamStartResult, StreamStopResult, StreamStatusResult } from '../../types/api';
import { validateCameraDeviceId } from '../../utils/validation';

export class StreamingService implements IStreaming {
  constructor(
    private apiClient: APIClient,
    private logger: LoggerService,
  ) {}

  /**
   * Start streaming for a specific camera device
   * Implements start_streaming RPC method
   */
  async startStreaming(device: string): Promise<StreamStartResult> {
    if (!validateCameraDeviceId(device)) {
      throw new Error(`Invalid device ID format: ${device}. Expected format: camera[0-9]+`);
    }

    try {
      this.logger.info(`Starting streaming for device: ${device}`);

      const response = await this.apiClient.call<StreamStartResult>('start_streaming', { device });

      this.logger.info(`Started streaming for ${device}`);
      return response;
    } catch (error) {
      this.logger.error(`Failed to start streaming for device: ${device}`, error as Record<string, unknown>);
      throw error;
    }
  }

  /**
   * Stop streaming for a specific camera device
   * Implements stop_streaming RPC method
   */
  async stopStreaming(device: string): Promise<StreamStopResult> {
    if (!validateCameraDeviceId(device)) {
      throw new Error(`Invalid device ID format: ${device}. Expected format: camera[0-9]+`);
    }

    try {
      this.logger.info(`Stopping streaming for device: ${device}`);

      const response = await this.apiClient.call<StreamStopResult>('stop_streaming', { device });

      this.logger.info(`Stopped streaming for ${device}`);
      return response;
    } catch (error) {
      this.logger.error(`Failed to stop streaming for device: ${device}`, error as Record<string, unknown>);
      throw error;
    }
  }

  /**
   * Get detailed status information for a specific camera stream
   * Implements get_stream_status RPC method
   */
  async getStreamStatus(device: string): Promise<StreamStatusResult> {
    if (!validateCameraDeviceId(device)) {
      throw new Error(`Invalid device ID format: ${device}. Expected format: camera[0-9]+`);
    }

    try {
      this.logger.info(`Getting stream status for device: ${device}`);

      const response = await this.apiClient.call<StreamStatusResult>('get_stream_status', { device });

      this.logger.info(`Retrieved stream status for ${device}`);
      return response;
    } catch (error) {
      this.logger.error(`Failed to get stream status for device: ${device}`, error as Record<string, unknown>);
      throw error;
    }
  }
}
