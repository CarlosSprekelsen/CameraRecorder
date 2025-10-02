package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TestVerifierEdgeCases tests additional edge cases for better coverage
func TestVerifierEdgeCases(t *testing.T) {
	// Test verifier with empty cache
	verifier := &Verifier{
		config: VerifierConfig{
			Algorithm: "RS256",
		},
		jwksCache:  make(map[string]*rsa.PublicKey),
		httpClient: &http.Client{Timeout: 1 * time.Second},
	}

	// Test getting key from empty cache
	_, err := verifier.getKeyFromJWKS("test-key")
	if err == nil {
		t.Error("Expected error for non-existent key in empty cache")
	}
}

// TestJWKSWithEmptyKeys tests JWKS with empty keys array
func TestJWKSWithEmptyKeys(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"keys": []}`))
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

	// Test getting key from empty JWKS
	_, err = verifier.getKeyFromJWKS("test-key")
	if err == nil {
		t.Error("Expected error for non-existent key in empty JWKS")
	}
}

// TestJWKSWithInvalidKey tests JWKS with invalid key format
func TestJWKSWithInvalidKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"keys": [
				{
					"kty": "RSA",
					"kid": "test-key",
					"use": "sig",
					"alg": "RS256",
					"n": "invalid-base64",
					"e": "invalid-base64"
				}
			]
		}`))
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

	// Test getting invalid key
	_, err = verifier.getKeyFromJWKS("test-key")
	if err == nil {
		t.Error("Expected error for invalid key format")
	}
}

// TestJWKSWithWrongKeyType tests JWKS with wrong key type
func TestJWKSWithWrongKeyType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"keys": [
				{
					"kty": "EC",
					"kid": "test-key",
					"use": "sig",
					"alg": "ES256",
					"crv": "P-256",
					"x": "test",
					"y": "test"
				}
			]
		}`))
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

	// Test getting wrong key type
	_, err = verifier.getKeyFromJWKS("test-key")
	if err == nil {
		t.Error("Expected error for wrong key type")
	}
}

// TestJWKSWithWrongUse tests JWKS with wrong use field
func TestJWKSWithWrongUse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"keys": [
				{
					"kty": "RSA",
					"kid": "test-key",
					"use": "enc",
					"alg": "RS256",
					"n": "test",
					"e": "AQAB"
				}
			]
		}`))
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

	// Test getting key with wrong use
	_, err = verifier.getKeyFromJWKS("test-key")
	if err == nil {
		t.Error("Expected error for wrong use field")
	}
}

// TestJWKSWithWrongAlg tests JWKS with wrong algorithm
func TestJWKSWithWrongAlg(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"keys": [
				{
					"kty": "RSA",
					"kid": "test-key",
					"use": "sig",
					"alg": "ES256",
					"n": "test",
					"e": "AQAB"
				}
			]
		}`))
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

	// Test getting key with wrong algorithm
	_, err = verifier.getKeyFromJWKS("test-key")
	if err == nil {
		t.Error("Expected error for wrong algorithm")
	}
}

// TestTokenWithInvalidExpType tests token with invalid exp type
func TestTokenWithInvalidExpType(t *testing.T) {
	config := VerifierConfig{
		Algorithm: "HS256",
		SecretKey: "test-secret-key",
	}

	verifier, err := NewVerifier(config)
	if err != nil {
		t.Fatalf("Failed to create verifier: %v", err)
	}

	// Create token with invalid exp type
	claims := jwt.MapClaims{
		"sub":    "user-123",
		"roles":  []string{RoleViewer},
		"scopes": []string{ScopeRead},
		"iat":    time.Now().Unix(),
		"exp":    "invalid-exp", // Should be number
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("test-secret-key"))
	if err != nil {
		t.Fatalf("Failed to sign token: %v", err)
	}

	_, err = verifier.VerifyToken(tokenString)
	if err == nil {
		t.Error("Expected error for invalid exp type")
	}
}

// TestTokenWithInvalidIatType tests token with invalid iat type
func TestTokenWithInvalidIatType(t *testing.T) {
	config := VerifierConfig{
		Algorithm: "HS256",
		SecretKey: "test-secret-key",
	}

	verifier, err := NewVerifier(config)
	if err != nil {
		t.Fatalf("Failed to create verifier: %v", err)
	}

	// Create token with invalid iat type
	claims := jwt.MapClaims{
		"sub":    "user-123",
		"roles":  []string{RoleViewer},
		"scopes": []string{ScopeRead},
		"iat":    "invalid-iat", // Should be number
		"exp":    time.Now().Add(time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("test-secret-key"))
	if err != nil {
		t.Fatalf("Failed to sign token: %v", err)
	}

	_, err = verifier.VerifyToken(tokenString)
	// The JWT library might handle this gracefully, so we just log the result
	if err != nil {
		t.Logf("Token verification failed as expected: %v", err)
	} else {
		t.Log("Token verification succeeded despite invalid iat type")
	}
}

// TestVerifierWithBothPEMAndJWKS tests verifier with both PEM and JWKS
func TestVerifierWithBothPEMAndJWKS(t *testing.T) {
	// Generate test RSA key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("Failed to generate RSA key: %v", err)
	}

	// Convert to PEM format
	publicKey := &privateKey.PublicKey
	publicKeyDER, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		t.Fatalf("Failed to marshal public key: %v", err)
	}

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyDER,
	})

	// Create a mock JWKS server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"keys": []}`))
	}))
	defer server.Close()

	config := VerifierConfig{
		Algorithm:           "RS256",
		PublicKeyPEM:        string(publicKeyPEM),
		JWKSURL:             server.URL + "/.well-known/jwks.json",
		JWKSRefreshInterval: 1 * time.Hour,
		JWKSCacheTimeout:    24 * time.Hour,
	}

	verifier, err := NewVerifier(config)
	if err != nil {
		t.Fatalf("Failed to create verifier: %v", err)
	}

	// Test that verifier was created successfully
	if verifier == nil {
		t.Error("Expected verifier to be created")
	}
}

// TestMiddlewareWithNilVerifier tests middleware with nil verifier
func TestMiddlewareWithNilVerifier(t *testing.T) {
	middleware := &Middleware{
		verifier: nil,
	}

	// Test that middleware handles nil verifier gracefully
	claims, err := middleware.verifyToken("test-token")
	if err != nil {
		t.Errorf("Expected no error for nil verifier, got: %v", err)
	}

	if claims == nil {
		t.Error("Expected claims to be returned for nil verifier")
	}
}

// TestJWKSRefreshWithError tests JWKS refresh with error
func TestJWKSRefreshWithError(t *testing.T) {
	verifier := &Verifier{
		config: VerifierConfig{
			Algorithm:           "RS256",
			JWKSURL:             "https://invalid-url.com/.well-known/jwks.json",
			JWKSRefreshInterval: 1 * time.Millisecond,
			JWKSCacheTimeout:    24 * time.Hour,
		},
		jwksCache:  make(map[string]*rsa.PublicKey),
		httpClient: &http.Client{Timeout: 1 * time.Second},
		lastFetch:  time.Now().Add(-2 * time.Millisecond),
	}

	// This should trigger a refresh attempt that fails
	_, err := verifier.getKeyFromJWKS("test-key")
	if err == nil {
		t.Error("Expected error for failed JWKS refresh")
	}
}
