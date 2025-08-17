/**
 * MediaMTX Camera Service JavaScript SDK
 * 
 * A JavaScript SDK for interacting with the MediaMTX Camera Service via WebSocket JSON-RPC.
 */

export { CameraClient } from './client';
export {
    CameraServiceError,
    AuthenticationError,
    ConnectionError,
    CameraNotFoundError,
    MediaMTXError,
    TimeoutError,
    ValidationError
} from './exceptions';
export type {
    CameraInfo,
    RecordingInfo,
    SnapshotInfo,
    ClientConfig,
    AuthType
} from './types';
