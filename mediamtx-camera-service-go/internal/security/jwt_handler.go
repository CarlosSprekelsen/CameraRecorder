package security

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/golang-jwt/jwt/v4"
)

// JWTClaims represents the claims structure for JWT tokens.
// Mirrors the Python JWTClaims dataclass structure for compatibility.
type JWTClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	IAT    int64  `json:"iat"`
	EXP    int64  `json:"exp"`
}

// ValidRoles defines the valid user roles in the system.
// Follows the Python VALID_ROLES set for compatibility.
var ValidRoles = map[string]bool{
	"viewer":   true,
	"operator": true,
	"admin":    true,
}

// ClientRateInfo represents rate limiting information for a client
type ClientRateInfo struct {
	ClientID     string
	RequestCount int64
	LastRequest  time.Time
	WindowStart  time.Time
}

// JWTHandler manages JWT token generation and validation.
// Implements JWT authentication with HS256 algorithm, configurable expiry,
// role-based access control, and rate limiting as specified in Architecture Decision AD-7.
type JWTHandler struct {
	secretKey string
	algorithm string
	logger    *logging.Logger

	// Rate limiting extensions (Phase 1 enhancement)
	clientRates map[string]*ClientRateInfo
	rateMutex   sync.RWMutex
	rateLimit   int64         // Requests per window
	rateWindow  time.Duration // Time window for rate limiting
}

// NewJWTHandler creates a new JWT handler instance with dependency injection.
// This constructor accepts a logger to ensure consistent logging across the security module
// and eliminates the creation of separate logger instances, following centralized logging principles.
// Returns an error if the secret key is empty or invalid.
func NewJWTHandler(secretKey string, logger *logging.Logger) (*JWTHandler, error) {
	if strings.TrimSpace(secretKey) == "" {
		return nil, fmt.Errorf("secret key must be provided")
	}

	// Use provided logger or create a default one if none provided
	if logger == nil {
		logger = logging.GetLogger("jwt-handler")
	}

	handler := &JWTHandler{
		secretKey: secretKey,
		algorithm: "HS256",
		logger:    logger,

		// Rate limiting initialization (Phase 1 enhancement)
		clientRates: make(map[string]*ClientRateInfo),
		rateLimit:   100,         // Default: 100 requests per window
		rateWindow:  time.Minute, // Default: 1 minute window
	}

	handler.logger.WithFields(logging.Fields{
		"algorithm":   handler.algorithm,
		"rate_limit":  handler.rateLimit,
		"rate_window": handler.rateWindow,
	}).Info("JWT handler initialized with rate limiting")
	return handler, nil
}

// GenerateToken creates a new JWT token with the specified claims.
// Returns the token string and any error encountered during generation.
func (h *JWTHandler) GenerateToken(userID, role string, expiryHours int) (string, error) {
	// Validate input parameters
	if strings.TrimSpace(userID) == "" {
		return "", fmt.Errorf("user ID cannot be empty")
	}

	if !ValidRoles[role] {
		return "", fmt.Errorf("invalid role: %s", role)
	}

	if expiryHours <= 0 {
		expiryHours = 24 // Default to 24 hours
	}

	// Create claims with current timestamp
	now := time.Now().Unix()
	claims := JWTClaims{
		UserID: userID,
		Role:   role,
		IAT:    now,
		EXP:    now + int64(expiryHours*3600),
	}

	// Create JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": claims.UserID,
		"role":    claims.Role,
		"iat":     claims.IAT,
		"exp":     claims.EXP,
	})

	// Sign the token
	tokenString, err := token.SignedString([]byte(h.secretKey))
	if err != nil {
		h.logger.Errorf("Failed to sign JWT token: %v", err)
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	h.logger.WithFields(logging.Fields{
		"user_id": userID,
		"role":    role,
		"expires": time.Unix(claims.EXP, 0).Format(time.RFC3339),
	}).Debug("JWT token generated successfully")

	return tokenString, nil
}

// Rate limiting methods (Phase 1 enhancement)

// CheckRateLimit checks if a client has exceeded the rate limit.
// This method implements a sliding window rate limiter that tracks requests per client
// and prevents abuse by blocking clients that exceed the configured rate limit.
// Returns true if the request is allowed, false if rate limit is exceeded.
func (h *JWTHandler) CheckRateLimit(clientID string) bool {
	h.rateMutex.Lock()
	defer h.rateMutex.Unlock()

	now := time.Now()
	clientRate, exists := h.clientRates[clientID]

	if !exists {
		// First request for this client
		h.clientRates[clientID] = &ClientRateInfo{
			ClientID:     clientID,
			RequestCount: 1,
			LastRequest:  now,
			WindowStart:  now,
		}
		return true
	}

	// Check if we're in a new time window
	if now.Sub(clientRate.WindowStart) >= h.rateWindow {
		// Reset for new window
		clientRate.RequestCount = 1
		clientRate.WindowStart = now
		clientRate.LastRequest = now
		return true
	}

	// Check if rate limit is exceeded before incrementing
	if clientRate.RequestCount >= h.rateLimit {
		h.logger.WithFields(logging.Fields{
			"client_id":     clientID,
			"request_count": clientRate.RequestCount,
			"rate_limit":    h.rateLimit,
			"window_start":  clientRate.WindowStart,
		}).Warn("Rate limit exceeded for client")
		return false
	}

	// Increment request count after checking
	clientRate.RequestCount++
	clientRate.LastRequest = now

	return true
}

// RecordRequest records a request for rate limiting (alternative to CheckRateLimit)
func (h *JWTHandler) RecordRequest(clientID string) {
	h.rateMutex.Lock()
	defer h.rateMutex.Unlock()

	now := time.Now()
	clientRate, exists := h.clientRates[clientID]

	if !exists {
		h.clientRates[clientID] = &ClientRateInfo{
			ClientID:     clientID,
			RequestCount: 1,
			LastRequest:  now,
			WindowStart:  now,
		}
		return
	}

	// Check if we're in a new time window
	if now.Sub(clientRate.WindowStart) >= h.rateWindow {
		clientRate.RequestCount = 1
		clientRate.WindowStart = now
		clientRate.LastRequest = now
		return
	}

	clientRate.RequestCount++
	clientRate.LastRequest = now
}

// GetClientRateInfo returns rate limiting information for a client
func (h *JWTHandler) GetClientRateInfo(clientID string) *ClientRateInfo {
	h.rateMutex.RLock()
	defer h.rateMutex.RUnlock()

	if clientRate, exists := h.clientRates[clientID]; exists {
		// Return a copy to avoid race conditions
		return &ClientRateInfo{
			ClientID:     clientRate.ClientID,
			RequestCount: clientRate.RequestCount,
			LastRequest:  clientRate.LastRequest,
			WindowStart:  clientRate.WindowStart,
		}
	}

	return nil
}

// SetRateLimit configures the rate limiting parameters
func (h *JWTHandler) SetRateLimit(limit int64, window time.Duration) {
	h.rateMutex.Lock()
	defer h.rateMutex.Unlock()

	h.rateLimit = limit
	h.rateWindow = window

	h.logger.WithFields(logging.Fields{
		"rate_limit":  limit,
		"rate_window": window,
	}).Info("Rate limiting configuration updated")
}

// CleanupExpiredClients removes rate limiting data for inactive clients
func (h *JWTHandler) CleanupExpiredClients(maxInactive time.Duration) {
	h.rateMutex.Lock()
	defer h.rateMutex.Unlock()

	now := time.Now()
	expiredClients := []string{}

	for clientID, clientRate := range h.clientRates {
		if now.Sub(clientRate.LastRequest) > maxInactive {
			expiredClients = append(expiredClients, clientID)
		}
	}

	for _, clientID := range expiredClients {
		delete(h.clientRates, clientID)
	}

	if len(expiredClients) > 0 {
		h.logger.WithField("expired_clients", fmt.Sprintf("%d", len(expiredClients))).Debug("Cleaned up expired client rate limiting data")
	}
}

// ValidateToken validates a JWT token and extracts claims with security best practices.
// This method implements algorithm restriction to prevent algorithm confusion attacks,
// validates all required claims, and performs comprehensive security checks.
// Returns the claims if valid, error if invalid or expired.
// Matches Python implementation security model: uses JWT library validation with explicit algorithm restriction.
func (h *JWTHandler) ValidateToken(tokenString string) (*JWTClaims, error) {
	if strings.TrimSpace(tokenString) == "" {
		return nil, fmt.Errorf("token cannot be empty")
	}

	// Use JWT library validation with explicit algorithm restriction (like Python)
	// This prevents algorithm confusion attacks and follows security best practices
	token, err := jwt.ParseWithClaims(tokenString, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate algorithm explicitly (like Python's algorithms=[self.algorithm])
		if token.Method.Alg() != "HS256" {
			h.logger.WithField("algorithm", token.Method.Alg()).Warn("Unsupported signing method detected")
			return nil, fmt.Errorf("unsupported signing method: %v", token.Method.Alg())
		}

		h.logger.WithField("signing_method", token.Method.Alg()).Debug("JWT signing method validated")
		return []byte(h.secretKey), nil
	})

	if err != nil {
		// Log the specific error for security auditing (like Python)
		h.logger.WithError(err).Warn("JWT token validation failed")

		// Return the original error for proper error handling (like Python)
		// Don't mask specific error types that could indicate security issues
		return nil, fmt.Errorf("failed to validate JWT token: %w", err)
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		h.logger.Warn("JWT token claims are not MapClaims")
		return nil, fmt.Errorf("invalid token claims")
	}

	if !token.Valid {
		h.logger.Warn("JWT token is not valid")
		return nil, fmt.Errorf("invalid token claims")
	}

	// Validate required fields
	requiredFields := []string{"user_id", "role", "iat", "exp"}
	for _, field := range requiredFields {
		if _, exists := claims[field]; !exists {
			h.logger.Warnf("JWT token missing required field: %s", field)
			return nil, fmt.Errorf("missing required field: %s", field)
		}
	}

	// Validate role
	role, ok := claims["role"].(string)
	if !ok || !ValidRoles[role] {
		h.logger.Warnf("JWT token has invalid role: %v", claims["role"])
		return nil, fmt.Errorf("invalid role: %v", claims["role"])
	}

	// Extract and validate timestamps
	iat, ok := claims["iat"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid issued at timestamp")
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid expiration timestamp")
	}

	// Check if token is expired
	if time.Now().Unix() > int64(exp) {
		h.logger.Warn("JWT token has expired")
		return nil, fmt.Errorf("token has expired")
	}

	// Create JWTClaims structure
	jwtClaims := &JWTClaims{
		UserID: claims["user_id"].(string),
		Role:   role,
		IAT:    int64(iat),
		EXP:    int64(exp),
	}

	h.logger.WithFields(logging.Fields{
		"user_id": jwtClaims.UserID,
		"role":    jwtClaims.Role,
		"expires": time.Unix(jwtClaims.EXP, 0).Format(time.RFC3339),
	}).Debug("JWT token validated successfully")

	return jwtClaims, nil
}

// IsTokenExpired checks if a JWT token is expired without full validation.
// Returns true if the token is expired, false otherwise.
func (h *JWTHandler) IsTokenExpired(tokenString string) bool {
	if strings.TrimSpace(tokenString) == "" {
		return true
	}

	// Parse token without validation to extract claims
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		h.logger.WithError(err).Debug("Failed to parse token for expiry check")
		return true
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return true
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return true
	}

	return time.Now().Unix() > int64(exp)
}

// GetSecretKey returns the secret key used for JWT signing.
// This method is primarily used for testing purposes.
func (h *JWTHandler) GetSecretKey() string {
	return h.secretKey
}

// GetAlgorithm returns the algorithm used for JWT signing.
func (h *JWTHandler) GetAlgorithm() string {
	return h.algorithm
}
