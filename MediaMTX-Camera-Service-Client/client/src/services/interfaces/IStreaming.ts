/**
 * Streaming Interface
 * 
 * Architecture requirement: "Modular architecture enabling independent feature development" (Section 1.2)
 * Separates streaming concerns from discovery concerns
 */

import { StreamStartResult, StreamStopResult, StreamStatusResult } from '../../types/api';

export interface IStreaming {
  /**
   * Start streaming for a specific camera device
   * Implements start_streaming RPC method
   */
  startStreaming(device: string): Promise<StreamStartResult>;

  /**
   * Stop streaming for a specific camera device
   * Implements stop_streaming RPC method
   */
  stopStreaming(device: string): Promise<StreamStopResult>;

  /**
   * Get detailed status information for a specific camera stream
   * Implements get_stream_status RPC method
   */
  getStreamStatus(device: string): Promise<StreamStatusResult>;
}
