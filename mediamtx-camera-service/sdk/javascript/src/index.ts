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

// CommonJS compatibility for Node.js v12
if (typeof module !== 'undefined' && module.exports) {
    module.exports = {
        CameraClient: require('./client').CameraClient,
        CameraServiceError: require('./exceptions').CameraServiceError,
        AuthenticationError: require('./exceptions').AuthenticationError,
        ConnectionError: require('./exceptions').ConnectionError,
        CameraNotFoundError: require('./exceptions').CameraNotFoundError,
        MediaMTXError: require('./exceptions').MediaMTXError,
        TimeoutError: require('./exceptions').TimeoutError,
        ValidationError: require('./exceptions').ValidationError
    };
}
