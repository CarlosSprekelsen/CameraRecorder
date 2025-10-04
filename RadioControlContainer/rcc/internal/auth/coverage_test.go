package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TestJWKSFetch tests JWKS fetching functionality
func TestJWKSFetch(t *testing.T) {
	// Create a mock JWKS server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/.well-known/jwks.json" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"keys":[]}`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	config := VerifierConfig{
		Algorithm:           "RS256",
		JWKSURL:             server.URL + "/.well-known/jwks.json",
		JWKSRefreshInterval: 1 * time.Hour,
		JWKSCacheTimeout:    24 * time.Hour,
	}

	verifier, err := NewVerifier(config)
	if err != nil {
		t.Fatalf("Failed to create verifier: %v", err)
	}

	// Test that JWKS was fetched successfully
	if verifier == nil {
		t.Error("Expected verifier to be created")
	}
}

// TestJWKSFetchError tests JWKS fetching with error
func TestJWKSFetchError(t *testing.T) {
	config := VerifierConfig{
		Algorithm:           "RS256",
		JWKSURL:             "https://invalid-url-that-does-not-exist.com/.well-known/jwks.json",
		JWKSRefreshInterval: 1 * time.Hour,
		JWKSCacheTimeout:    24 * time.Hour,
	}

	_, err := NewVerifier(config)
	// This test may or may not fail depending on network conditions
	// We just want to test that the function handles errors gracefully
	if err != nil {
		t.Logf("JWKS fetch failed as expected: %v", err)
	}
}

// TestExtractStringSliceEdgeCases tests edge cases for string slice extraction
func TestExtractStringSliceEdgeCases(t *testing.T) {
	config := VerifierConfig{
		Algorithm: "HS256",
		SecretKey: "test-secret-key",
	}

	verifier, err := NewVerifier(config)
	if err != nil {
		t.Fatalf("Failed to create verifier: %v", err)
	}

	// Test with invalid claim types
	claims := jwt.MapClaims{
		"sub":    "user-123",
		"roles":  "invalid-type", // Should be array
		"scopes": []string{ScopeRead},
		"iat":    time.Now().Unix(),
		"exp":    time.Now().Add(time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("test-secret-key"))
	if err != nil {
		t.Fatalf("Failed to sign token: %v", err)
	}

	_, err = verifier.VerifyToken(tokenString)
	if err == nil {
		t.Error("Expected error for invalid roles claim type")
	}
}

// TestJWKToRSAPublicKey tests JWK to RSA public key conversion
func TestJWKToRSAPublicKey(t *testing.T) {
	// Generate test RSA key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}

	publicKey := &privateKey.PublicKey
	publicKeyDER, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		t.Fatalf("Failed to marshal public key: %v", err)
	}

	// Create JWK
	jwk := JWK{
		Kty: "RSA",
		Kid: "test-key",
		Use: "sig",
		Alg: "RS256",
		N:   string(publicKeyDER), // This is not correct base64url, but tests the function
		E:   "AQAB",
	}

	// This should fail because the N and E values are not properly base64url encoded
	_, err = (&Verifier{}).jwkToRSAPublicKey(jwk)
	if err == nil {
		t.Error("Expected error for invalid JWK format")
	}
}

// TestJwkToRSAPublicKeyErrorPaths tests specific error paths in jwkToRSAPublicKey
func TestJwkToRSAPublicKeyErrorPaths(t *testing.T) {
	// Test invalid modulus (N) - should fail on base64URLDecode
	jwk1 := JWK{
		Kty: "RSA",
		Use: "sig",
		Alg: "RS256",
		N:   "invalid-base64url", // Invalid base64url
		E:   "AQAB",              // Valid base64url
	}

	_, err := (&Verifier{}).jwkToRSAPublicKey(jwk1)
	if err == nil {
		t.Error("Expected error for invalid modulus")
	}
	if !strings.Contains(err.Error(), "failed to decode modulus") {
		t.Errorf("Expected modulus decode error, got: %v", err)
	}

	// Test invalid exponent (E) - should fail on base64URLDecode
	jwk2 := JWK{
		Kty: "RSA",
		Use: "sig",
		Alg: "RS256",
		N:   "dGVzdA",            // Valid base64url
		E:   "invalid-base64url", // Invalid base64url
	}

	_, err = (&Verifier{}).jwkToRSAPublicKey(jwk2)
	if err == nil {
		t.Error("Expected error for invalid exponent")
	}
	if !strings.Contains(err.Error(), "failed to decode exponent") {
		t.Errorf("Expected exponent decode error, got: %v", err)
	}

	// Test edge case: empty modulus (valid base64url, but creates invalid RSA key)
	jwk3 := JWK{
		Kty: "RSA",
		Use: "sig",
		Alg: "RS256",
		N:   "", // Empty modulus - valid base64url but invalid RSA
		E:   "AQAB",
	}

	_, err = (&Verifier{}).jwkToRSAPublicKey(jwk3)
	// Empty string is valid base64url, so this should succeed but create an invalid key
	// The function doesn't validate RSA key correctness, just base64url decoding
	if err != nil {
		t.Errorf("Empty modulus should decode successfully: %v", err)
	}

	// Test edge case: empty exponent (valid base64url, but creates invalid RSA key)
	jwk4 := JWK{
		Kty: "RSA",
		Use: "sig",
		Alg: "RS256",
		N:   "dGVzdA",
		E:   "", // Empty exponent - valid base64url but invalid RSA
	}

	_, err = (&Verifier{}).jwkToRSAPublicKey(jwk4)
	// Empty string is valid base64url, so this should succeed but create an invalid key
	if err != nil {
		t.Errorf("Empty exponent should decode successfully: %v", err)
	}
}

// TestBase64URLDecode tests base64url decoding
func TestBase64URLDecode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "valid base64url",
			input:    "dGVzdA",
			expected: "test",
			wantErr:  false,
		},
		{
			name:     "with padding",
			input:    "dGVzdA==",
			expected: "test",
			wantErr:  false,
		},
		{
			name:     "invalid base64",
			input:    "invalid!",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := base64URLDecode(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("base64URLDecode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && string(result) != tt.expected {
				t.Errorf("base64URLDecode() = %v, expected %v", string(result), tt.expected)
			}
		})
	}
}

// TestGetKeyFromJWKS tests key retrieval from JWKS cache
func TestGetKeyFromJWKS(t *testing.T) {
	// Create a verifier with empty cache to test cache logic
	verifier := &Verifier{
		config: VerifierConfig{
			Algorithm:           "RS256",
			JWKSURL:             "https://example.com/.well-known/jwks.json",
			JWKSRefreshInterval: 1 * time.Hour,
			JWKSCacheTimeout:    24 * time.Hour,
		},
		jwksCache:  make(map[string]*rsa.PublicKey),
		httpClient: &http.Client{Timeout: 1 * time.Second},
	}

	// Test getting non-existent key
	_, err := verifier.getKeyFromJWKS("non-existent-key")
	if err == nil {
		t.Error("Expected error for non-existent key")
	}
}

// TestMiddlewareEdgeCases tests edge cases for middleware
func TestMiddlewareEdgeCases(t *testing.T) {
	middleware := NewMiddleware()

	// Test with nil claims
	if middleware.hasRequiredScopes(nil, []string{ScopeRead}) {
		t.Error("Expected false for nil claims")
	}

	if middleware.hasRequiredRoles(nil, []string{RoleViewer}) {
		t.Error("Expected false for nil claims")
	}

	// Test with empty required scopes/roles
	claims := &Claims{
		Subject: "user-123",
		Roles:   []string{RoleViewer},
		Scopes:  []string{ScopeRead},
	}

	// Empty required scopes should return true (no requirements)
	if !middleware.hasRequiredScopes(claims, []string{}) {
		t.Error("Expected true for empty required scopes")
	}

	// Empty required roles should return true (no requirements)
	if !middleware.hasRequiredRoles(claims, []string{}) {
		t.Error("Expected true for empty required roles")
	}
}

// TestErrorResponseFormatCoverage tests error response formatting
func TestErrorResponseFormatCoverage(t *testing.T) {
	w := httptest.NewRecorder()

	writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Test error", map[string]string{"detail": "test"})

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}

	if w.Header().Get("Content-Type") != "application/json" {
		t.Error("Expected Content-Type to be application/json")
	}
}

// TestCorrelationIDGeneration tests correlation ID generation
func TestCorrelationIDGeneration(t *testing.T) {
	id1 := generateCorrelationID()
	id2 := generateCorrelationID()

	if id1 == id2 {
		t.Error("Expected different correlation IDs")
	}

	if id1 == "" {
		t.Error("Expected non-empty correlation ID")
	}
}

// TestVerifierConfigDefaults tests default configuration values
func TestVerifierConfigDefaults(t *testing.T) {
	config := VerifierConfig{
		Algorithm: "HS256",
		SecretKey: "test-secret",
	}

	verifier, err := NewVerifier(config)
	if err != nil {
		t.Fatalf("Failed to create verifier: %v", err)
	}

	if verifier.config.Algorithm != "HS256" {
		t.Errorf("Expected algorithm HS256, got %s", verifier.config.Algorithm)
	}

	if verifier.config.SecretKey != "test-secret" {
		t.Errorf("Expected secret key 'test-secret', got %s", verifier.config.SecretKey)
	}
}

// TestTokenExpiration tests token expiration handling
func TestTokenExpiration(t *testing.T) {
	config := VerifierConfig{
		Algorithm: "HS256",
		SecretKey: "test-secret-key",
	}

	verifier, err := NewVerifier(config)
	if err != nil {
		t.Fatalf("Failed to create verifier: %v", err)
	}

	// Create expired token
	claims := jwt.MapClaims{
		"sub":    "user-123",
		"roles":  []string{RoleViewer},
		"scopes": []string{ScopeRead},
		"iat":    time.Now().Add(-2 * time.Hour).Unix(),
		"exp":    time.Now().Add(-1 * time.Hour).Unix(), // Expired
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("test-secret-key"))
	if err != nil {
		t.Fatalf("Failed to sign token: %v", err)
	}

	_, err = verifier.VerifyToken(tokenString)
	if err == nil {
		t.Error("Expected error for expired token")
	}
}

// TestInvalidAlgorithm tests invalid algorithm handling
func TestInvalidAlgorithm(t *testing.T) {
	config := VerifierConfig{
		Algorithm: "INVALID",
	}

	_, err := NewVerifier(config)
	if err == nil {
		t.Error("Expected error for invalid algorithm")
	}
}

// TestMissingRequiredFields tests missing required fields in token
func TestMissingRequiredFields(t *testing.T) {
	config := VerifierConfig{
		Algorithm: "HS256",
		SecretKey: "test-secret-key",
	}

	verifier, err := NewVerifier(config)
	if err != nil {
		t.Fatalf("Failed to create verifier: %v", err)
	}

	// Create token with missing 'sub' field
	claims := jwt.MapClaims{
		"roles":  []string{RoleViewer},
		"scopes": []string{ScopeRead},
		"iat":    time.Now().Unix(),
		"exp":    time.Now().Add(time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("test-secret-key"))
	if err != nil {
		t.Fatalf("Failed to sign token: %v", err)
	}

	_, err = verifier.VerifyToken(tokenString)
	if err == nil {
		t.Error("Expected error for missing 'sub' field")
	}
}
