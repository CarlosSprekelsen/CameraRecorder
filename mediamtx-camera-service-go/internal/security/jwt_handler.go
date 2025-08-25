package security

import (
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/sirupsen/logrus"
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

// JWTHandler manages JWT token generation and validation.
// Implements JWT authentication with HS256 algorithm, configurable expiry,
// and role-based access control as specified in Architecture Decision AD-7.
type JWTHandler struct {
	secretKey string
	algorithm string
	logger    *logrus.Logger
}

// NewJWTHandler creates a new JWT handler instance.
// Returns an error if the secret key is empty or invalid.
func NewJWTHandler(secretKey string) (*JWTHandler, error) {
	if strings.TrimSpace(secretKey) == "" {
		return nil, fmt.Errorf("secret key must be provided")
	}

	handler := &JWTHandler{
		secretKey: secretKey,
		algorithm: "HS256",
		logger:    logrus.New(),
	}

	handler.logger.WithField("algorithm", handler.algorithm).Info("JWT handler initialized")
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

	h.logger.WithFields(logrus.Fields{
		"user_id": userID,
		"role":    role,
		"expires": time.Unix(claims.EXP, 0).Format(time.RFC3339),
	}).Debug("JWT token generated successfully")

	return tokenString, nil
}

// ValidateToken validates a JWT token and extracts claims.
// Returns the claims if valid, nil if invalid or expired.
func (h *JWTHandler) ValidateToken(tokenString string) (*JWTClaims, error) {
	if strings.TrimSpace(tokenString) == "" {
		return nil, fmt.Errorf("token cannot be empty")
	}

	// Parse and validate the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(h.secretKey), nil
	})

	if err != nil {
		h.logger.WithError(err).Warn("JWT token validation failed")
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		h.logger.Warn("JWT token claims are invalid")
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

	h.logger.WithFields(logrus.Fields{
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
