// Package auth implements Auth from Architecture §5.
//
// Requirements:
//   - Architecture §5: "Validate tokens; enforce role-based access control"
//
// Source: OpenAPI v1 §1.1 & §1.2
// Quote: "Send Authorization: Bearer <token> header on every request (except /health)"
// Quote: "viewer: read-only (list radios, get state, subscribe to telemetry)"
// Quote: "controller: all viewer privileges plus control actions (select radio, set power, set channel)"
package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/radio-control/rcc/internal/api"
)

// Claims represents the parsed token claims.
type Claims struct {
	Subject string   `json:"sub"`
	Roles   []string `json:"roles"`
	Scopes  []string `json:"scopes"`
}

// ContextKey is used for storing claims in request context.
type ContextKey string

const (
	ClaimsKey ContextKey = "claims"
)

// Role constants per OpenAPI v1 §1.2
const (
	RoleViewer     = "viewer"
	RoleController = "controller"
)

// Scope constants per OpenAPI v1 §1.2
const (
	ScopeRead      = "read"
	ScopeControl   = "control"
	ScopeTelemetry = "telemetry"
)

// Middleware handles authentication and authorization.
type Middleware struct {
	// TODO: Add real token verifier
	// For now, we'll use a mock verifier
}

// NewMiddleware creates a new auth middleware.
func NewMiddleware() *Middleware {
	return &Middleware{}
}

// RequireAuth creates middleware that requires authentication.
// Source: OpenAPI v1 §1.1
func (m *Middleware) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for health endpoint
		if r.URL.Path == "/api/v1/health" {
			next(w, r)
			return
		}

		// Extract bearer token
		token, err := m.extractBearerToken(r)
		if err != nil {
			api.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", 
				"Authentication required", nil)
			return
		}

		// Verify token and extract claims
		claims, err := m.verifyToken(token)
		if err != nil {
			api.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", 
				"Invalid token", nil)
			return
		}

		// Store claims in context
		ctx := context.WithValue(r.Context(), ClaimsKey, claims)
		next(w, r.WithContext(ctx))
	}
}

// RequireScope creates middleware that requires specific scopes.
// Source: OpenAPI v1 §1.2
func (m *Middleware) RequireScope(requiredScopes ...string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			claims := m.getClaimsFromContext(r.Context())
			if claims == nil {
				api.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", 
					"Authentication required", nil)
				return
			}

			// Check if user has required scopes
			if !m.hasRequiredScopes(claims, requiredScopes) {
				api.WriteError(w, http.StatusForbidden, "FORBIDDEN", 
					"Insufficient permissions", nil)
				return
			}

			next(w, r)
		}
	}
}

// RequireRole creates middleware that requires specific roles.
// Source: OpenAPI v1 §1.2
func (m *Middleware) RequireRole(requiredRoles ...string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			claims := m.getClaimsFromContext(r.Context())
			if claims == nil {
				api.WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED", 
					"Authentication required", nil)
				return
			}

			// Check if user has required roles
			if !m.hasRequiredRoles(claims, requiredRoles) {
				api.WriteError(w, http.StatusForbidden, "FORBIDDEN", 
					"Insufficient permissions", nil)
				return
			}

			next(w, r)
		}
	}
}

// extractBearerToken extracts the bearer token from the Authorization header.
func (m *Middleware) extractBearerToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("missing Authorization header")
	}

	// Check for Bearer prefix
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", fmt.Errorf("invalid Authorization header format")
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" {
		return "", fmt.Errorf("empty token")
	}

	return token, nil
}

// verifyToken verifies the token and returns claims.
// TODO: Implement real token verification
// For now, this is a mock implementation
func (m *Middleware) verifyToken(token string) (*Claims, error) {
	// Mock token verification
	// In production, this would verify JWT signature, check expiration, etc.
	
	// Simple mock tokens for testing
	switch token {
	case "viewer-token":
		return &Claims{
			Subject: "user-123",
			Roles:   []string{RoleViewer},
			Scopes:  []string{ScopeRead, ScopeTelemetry},
		}, nil
	case "controller-token":
		return &Claims{
			Subject: "admin-456",
			Roles:   []string{RoleController},
			Scopes:  []string{ScopeRead, ScopeControl, ScopeTelemetry},
		}, nil
	case "invalid-token":
		return nil, fmt.Errorf("token verification failed")
	default:
		// Default to viewer for unknown tokens
		return &Claims{
			Subject: "user-unknown",
			Roles:   []string{RoleViewer},
			Scopes:  []string{ScopeRead, ScopeTelemetry},
		}, nil
	}
}

// hasRequiredScopes checks if the user has all required scopes.
func (m *Middleware) hasRequiredScopes(claims *Claims, requiredScopes []string) bool {
	if claims == nil {
		return false
	}

	for _, required := range requiredScopes {
		found := false
		for _, scope := range claims.Scopes {
			if scope == required {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// hasRequiredRoles checks if the user has any of the required roles.
func (m *Middleware) hasRequiredRoles(claims *Claims, requiredRoles []string) bool {
	if claims == nil {
		return false
	}

	for _, required := range requiredRoles {
		for _, role := range claims.Roles {
			if role == required {
				return true
			}
		}
	}

	return false
}

// getClaimsFromContext extracts claims from the request context.
func (m *Middleware) getClaimsFromContext(ctx context.Context) *Claims {
	claims, ok := ctx.Value(ClaimsKey).(*Claims)
	if !ok {
		return nil
	}
	return claims
}

// GetClaimsFromRequest extracts claims from the request context.
// This is a helper function for use in handlers.
func GetClaimsFromRequest(r *http.Request) *Claims {
	claims, ok := r.Context().Value(ClaimsKey).(*Claims)
	if !ok {
		return nil
	}
	return claims
}

// IsViewer checks if the user has viewer role.
func (m *Middleware) IsViewer(claims *Claims) bool {
	return m.hasRequiredRoles(claims, []string{RoleViewer})
}

// IsController checks if the user has controller role.
func (m *Middleware) IsController(claims *Claims) bool {
	return m.hasRequiredRoles(claims, []string{RoleController})
}

// CanRead checks if the user can perform read operations.
func (m *Middleware) CanRead(claims *Claims) bool {
	return m.hasRequiredScopes(claims, []string{ScopeRead})
}

// CanControl checks if the user can perform control operations.
func (m *Middleware) CanControl(claims *Claims) bool {
	return m.hasRequiredScopes(claims, []string{ScopeControl})
}

// CanAccessTelemetry checks if the user can access telemetry.
func (m *Middleware) CanAccessTelemetry(claims *Claims) bool {
	return m.hasRequiredScopes(claims, []string{ScopeTelemetry})
}
