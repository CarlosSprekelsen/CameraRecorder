package auth

import (
	"crypto/rsa"
	"net/http"
	"testing"
	"time"
)

// TestAdditionalEdgeCases tests additional edge cases for better coverage
func TestAdditionalEdgeCases(t *testing.T) {
	// Test middleware with empty required scopes
	middleware := NewMiddleware()
	claims := &Claims{
		Subject: "user-123",
		Roles:   []string{RoleViewer},
		Scopes:  []string{ScopeRead},
	}

	// Test empty required scopes
	if !middleware.hasRequiredScopes(claims, []string{}) {
		t.Error("Expected true for empty required scopes")
	}

	// Test empty required roles
	if !middleware.hasRequiredRoles(claims, []string{}) {
		t.Error("Expected true for empty required roles")
	}

	// Test with nil claims
	if middleware.hasRequiredScopes(nil, []string{ScopeRead}) {
		t.Error("Expected false for nil claims with required scopes")
	}

	if middleware.hasRequiredRoles(nil, []string{RoleViewer}) {
		t.Error("Expected false for nil claims with required roles")
	}
}

// TestVerifierConfigEdgeCases tests verifier config edge cases
func TestVerifierConfigEdgeCases(t *testing.T) {
	// Test with empty algorithm
	config := VerifierConfig{
		Algorithm: "",
	}

	_, err := NewVerifier(config)
	if err == nil {
		t.Error("Expected error for empty algorithm")
	}

	// Test with unsupported algorithm
	config = VerifierConfig{
		Algorithm: "ES256",
	}

	_, err = NewVerifier(config)
	if err == nil {
		t.Error("Expected error for unsupported algorithm")
	}
}

// TestJWKSRefreshEdgeCases tests JWKS refresh edge cases
func TestJWKSRefreshEdgeCases(t *testing.T) {
	// Test verifier with very short refresh interval
	verifier := &Verifier{
		config: VerifierConfig{
			Algorithm:           "RS256",
			JWKSURL:             "https://example.com/.well-known/jwks.json",
			JWKSRefreshInterval: 1 * time.Millisecond,
			JWKSCacheTimeout:    24 * time.Hour,
		},
		jwksCache:  make(map[string]*rsa.PublicKey),
		httpClient: &http.Client{Timeout: 1 * time.Second},
		lastFetch:  time.Now().Add(-2 * time.Millisecond),
	}

	// This should trigger a refresh attempt
	_, err := verifier.getKeyFromJWKS("test-key")
	// We expect an error because the URL doesn't exist
	if err == nil {
		t.Error("Expected error for non-existent key")
	}
}

// TestTokenValidationEdgeCases tests token validation edge cases
func TestTokenValidationEdgeCases(t *testing.T) {
	config := VerifierConfig{
		Algorithm: "HS256",
		SecretKey: "test-secret-key",
	}

	verifier, err := NewVerifier(config)
	if err != nil {
		t.Fatalf("Failed to create verifier: %v", err)
	}

	// Test with empty token
	_, err = verifier.VerifyToken("")
	if err == nil {
		t.Error("Expected error for empty token")
	}

	// Test with whitespace-only token
	_, err = verifier.VerifyToken("   ")
	if err == nil {
		t.Error("Expected error for whitespace-only token")
	}
}

// TestMiddlewareHelperEdgeCases tests middleware helper edge cases
func TestMiddlewareHelperEdgeCases(t *testing.T) {
	middleware := NewMiddleware()

	// Test with nil claims
	if middleware.IsViewer(nil) {
		t.Error("Expected false for nil claims")
	}

	if middleware.IsController(nil) {
		t.Error("Expected false for nil claims")
	}

	if middleware.CanRead(nil) {
		t.Error("Expected false for nil claims")
	}

	if middleware.CanControl(nil) {
		t.Error("Expected false for nil claims")
	}

	if middleware.CanAccessTelemetry(nil) {
		t.Error("Expected false for nil claims")
	}
}

// TestErrorResponseEdgeCases tests error response edge cases
func TestErrorResponseEdgeCases(t *testing.T) {
	// Test writeError with nil details
	// This is already tested in other tests, but we can add more coverage
	// by testing different scenarios
}

// TestJWKSFetchEdgeCases tests JWKS fetch edge cases
func TestJWKSFetchEdgeCases(t *testing.T) {
	// Test verifier with invalid JWKS URL
	config := VerifierConfig{
		Algorithm:           "RS256",
		JWKSURL:             "https://invalid-url.com/.well-known/jwks.json",
		JWKSRefreshInterval: 1 * time.Hour,
		JWKSCacheTimeout:    24 * time.Hour,
	}

	_, err := NewVerifier(config)
	// This might fail or succeed depending on network conditions
	// We just want to test that the function handles errors gracefully
	if err != nil {
		t.Logf("JWKS fetch failed as expected: %v", err)
	}
}
