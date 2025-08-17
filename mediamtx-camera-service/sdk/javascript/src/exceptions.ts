/**
 * Custom exceptions for MediaMTX Camera Service SDK.
 */

export class CameraServiceError extends Error {
    constructor(message: string) {
        super(message);
        this.name = 'CameraServiceError';
    }
}

export class AuthenticationError extends CameraServiceError {
    constructor(message: string) {
        super(message);
        this.name = 'AuthenticationError';
    }
}

export class ConnectionError extends CameraServiceError {
    constructor(message: string) {
        super(message);
        this.name = 'ConnectionError';
    }
}

export class CameraNotFoundError extends CameraServiceError {
    constructor(message: string) {
        super(message);
        this.name = 'CameraNotFoundError';
    }
}

export class MediaMTXError extends CameraServiceError {
    constructor(message: string) {
        super(message);
        this.name = 'MediaMTXError';
    }
}

export class TimeoutError extends CameraServiceError {
    constructor(message: string) {
        super(message);
        this.name = 'TimeoutError';
    }
}

export class ValidationError extends CameraServiceError {
    constructor(message: string) {
        super(message);
        this.name = 'ValidationError';
    }
}
