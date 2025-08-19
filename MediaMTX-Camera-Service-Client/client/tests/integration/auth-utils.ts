import jwt from 'jsonwebtoken';

/**
 * Authentication utilities for integration tests
 * Uses environment variables set by set-test-env.sh
 */

// Get JWT secret from environment
const getJwtSecret = (): string => {
    const secret = process.env.CAMERA_SERVICE_JWT_SECRET;
    if (!secret) {
        throw new Error('CAMERA_SERVICE_JWT_SECRET environment variable not set. Run: ./set-test-env.sh');
    }
    return secret;
};

/**
 * Generate a valid JWT token for testing
 * @param userId - User ID for the token
 * @param role - User role (viewer, operator, admin)
 * @param expiresIn - Token expiration in seconds (default: 24 hours)
 * @returns JWT token
 */
export const generateValidToken = (userId = 'test_user', role = 'operator', expiresIn = 24 * 60 * 60): string => {
    const secret = getJwtSecret();
    
    const payload = {
        user_id: userId,
        role: role,
        iat: Math.floor(Date.now() / 1000),
        exp: Math.floor(Date.now() / 1000) + expiresIn
    };
    
    return jwt.sign(payload, secret, { algorithm: 'HS256' });
};

/**
 * Generate an invalid JWT token for testing authentication failures
 * @returns Invalid JWT token
 */
export const generateInvalidToken = (): string => {
    const payload = {
        user_id: 'test_user',
        role: 'operator',
        iat: Math.floor(Date.now() / 1000),
        exp: Math.floor(Date.now() / 1000) + (24 * 60 * 60)
    };
    
    // Use wrong secret to generate invalid token
    return jwt.sign(payload, 'wrong_secret', { algorithm: 'HS256' });
};

/**
 * Generate an expired JWT token for testing
 * @returns Expired JWT token
 */
export const generateExpiredToken = (): string => {
    const secret = getJwtSecret();
    
    const payload = {
        user_id: 'test_user',
        role: 'operator',
        iat: Math.floor(Date.now() / 1000) - (48 * 60 * 60), // 48 hours ago
        exp: Math.floor(Date.now() / 1000) - (24 * 60 * 60)  // 24 hours ago
    };
    
    return jwt.sign(payload, secret, { algorithm: 'HS256' });
};

/**
 * Validate that the test environment is properly set up
 * @returns True if environment is ready
 */
export const validateTestEnvironment = (): boolean => {
    try {
        getJwtSecret();
        console.log('✅ Test environment validated - JWT secret available');
        return true;
    } catch (error) {
        console.error('❌ Test environment validation failed:', (error as Error).message);
        console.error('💡 Run: ./set-test-env.sh to set up the test environment');
        return false;
    }
};

/**
 * Get different role tokens for authorization testing
 * @returns Tokens for different roles
 */
export const getRoleTokens = (): { viewer: string; operator: string; admin: string } => {
    return {
        viewer: generateValidToken('viewer_user', 'viewer'),
        operator: generateValidToken('operator_user', 'operator'),
        admin: generateValidToken('admin_user', 'admin')
    };
};
