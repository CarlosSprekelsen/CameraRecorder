/**
 * Centralized Validation Utilities
 * 
 * Follows architecture principle: Single Responsibility
 * All validation logic consolidated in one place
 */

/**
 * Validates camera device ID format according to server documentation
 * Pattern: camera[0-9]+ (e.g., camera0, camera1, camera10)
 */
export function validateCameraDeviceId(deviceId: string): boolean {
  if (!deviceId || typeof deviceId !== 'string') {
    return false;
  }
  
  // Server documentation pattern: camera[0-9]+
  const cameraIdPattern = /^camera\d+$/;
  return cameraIdPattern.test(deviceId);
}

/**
 * Validates parameter structure for API calls
 * Ensures parameters are objects, not arrays
 */
export function validateParameterStructure(params: any): boolean {
  return typeof params === 'object' && params !== null && !Array.isArray(params);
}

/**
 * Validates JWT token format
 */
export function validateJWTToken(token: string): boolean {
  if (!token || typeof token !== 'string') {
    return false;
  }
  
  // Basic JWT format validation (3 parts separated by dots)
  const jwtPattern = /^[A-Za-z0-9-_]+\.[A-Za-z0-9-_]+\.[A-Za-z0-9-_]+$/;
  return jwtPattern.test(token);
}

/**
 * Validates ISO timestamp format (timezone-aware)
 */
export function validateIsoTimestamp(timestamp: string): boolean {
  if (!timestamp || typeof timestamp !== 'string') {
    return false;
  }
  
  try {
    const date = new Date(timestamp);
    return !isNaN(date.getTime()) && timestamp.includes('T') && timestamp.includes(':');
  } catch {
    return false;
  }
}
