/**
 * Streaming Interface
 * 
 * Architecture requirement: "Modular architecture enabling independent feature development" (Section 1.2)
 * Separates streaming concerns from discovery concerns
 */

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

export interface StreamStartResult {
  device: string;
  status: 'STARTING' | 'ACTIVE' | 'ERROR';
  stream_url?: string;
  message?: string;
}

export interface StreamStopResult {
  device: string;
  status: 'STOPPED' | 'ERROR';
  message?: string;
}

export interface StreamStatusResult {
  device: string;
  status: 'ACTIVE' | 'INACTIVE' | 'ERROR';
  stream_url?: string;
  viewers?: number;
  uptime?: number;
  message?: string;
}
