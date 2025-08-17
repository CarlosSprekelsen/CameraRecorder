/**
 * Type definitions for MediaMTX Camera Service SDK.
 */

export type AuthType = 'jwt' | 'api_key';

export interface ClientConfig {
    host?: string;
    port?: number;
    useSsl?: boolean;
    authType?: AuthType;
    authToken?: string;
    apiKey?: string;
    maxRetries?: number;
    retryDelay?: number;
}

export interface CameraInfo {
    devicePath: string;
    name: string;
    capabilities: string[];
    status: string;
    streamUrl?: string;
}

export interface RecordingInfo {
    devicePath: string;
    recordingId: string;
    filename: string;
    startTime: number;
    duration?: number;
    status: string;
}

export interface SnapshotInfo {
    devicePath: string;
    filename: string;
    timestamp: number;
    sizeBytes?: number;
}

export interface JsonRpcRequest {
    jsonrpc: '2.0';
    id: number;
    method: string;
    params?: Record<string, any>;
}

export interface JsonRpcResponse {
    jsonrpc: '2.0';
    id: number;
    result?: any;
    error?: {
        code: number;
        message: string;
        data?: any;
    };
}

export interface JsonRpcNotification {
    jsonrpc: '2.0';
    method: string;
    params?: Record<string, any>;
}

export type JsonRpcMessage = JsonRpcRequest | JsonRpcResponse | JsonRpcNotification;
