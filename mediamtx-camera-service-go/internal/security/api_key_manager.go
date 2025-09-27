/*
API Key Manager Implementation

Requirements Coverage:
- REQ-SEC-014: Key Management
- REQ-SEC-015: Production API Key Management
- REQ-SEC-016: Key Rotation and Expiration

Test Categories: Unit/Integration
API Documentation Reference: docs/api/json_rpc_methods.md

Provides production-grade API key management with secure generation, storage, and lifecycle management.
Follows canonical configuration patterns and event-based progressive readiness architecture.
*/

package security

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/google/uuid"
)

// APIKey represents a single API key with metadata
type APIKey struct {
	ID          string    `json:"id"`
	Key         string    `json:"key"`
	Role        Role      `json:"role"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
	Description string    `json:"description"`
	LastUsed    time.Time `json:"last_used"`
	UsageCount  int64     `json:"usage_count"`
	Status      string    `json:"status"` // active, revoked, expired
}

// APIKeyStorage represents the storage structure for API keys
type APIKeyStorage struct {
	Keys map[string]*APIKey `json:"keys"`
	mu   sync.RWMutex
}

// APIKeyManager manages API key lifecycle and operations
type APIKeyManager struct {
	config      *config.APIKeyManagementConfig
	logger      *logging.Logger
	storage     *APIKeyStorage
	storagePath string
	mu          sync.RWMutex
}

// NewAPIKeyManager creates a new API key manager instance
func NewAPIKeyManager(config *config.APIKeyManagementConfig, logger *logging.Logger) (*APIKeyManager, error) {
	if config == nil {
		return nil, fmt.Errorf("configuration cannot be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}

	manager := &APIKeyManager{
		config:      config,
		logger:      logger,
		storage:     &APIKeyStorage{Keys: make(map[string]*APIKey)},
		storagePath: config.StoragePath,
	}

	// Load existing keys from storage
	if err := manager.loadKeys(); err != nil {
		manager.logger.WithError(err).Warn("Failed to load existing API keys, starting with empty storage")
	}

	manager.logger.WithFields(logging.Fields{
		"storage_path": config.StoragePath,
		"key_count":    len(manager.storage.Keys),
	}).Info("API Key Manager initialized")

	return manager, nil
}

// GenerateKey generates a new API key with the specified role and expiry
func (km *APIKeyManager) GenerateKey(role Role, expiry time.Duration, description string) (*APIKey, error) {
	km.mu.Lock()
	defer km.mu.Unlock()

	// Validate role
	if role == 0 {
		return nil, fmt.Errorf("role cannot be empty")
	}

	// Check maximum keys per role
	if km.config.MaxKeysPerRole > 0 {
		count := km.countKeysByRole(role)
		if count >= km.config.MaxKeysPerRole {
			return nil, fmt.Errorf("maximum keys per role exceeded for role %s", role)
		}
	}

	// Generate secure key
	keyBytes := make([]byte, km.config.KeyLength)
	if _, err := rand.Read(keyBytes); err != nil {
		return nil, fmt.Errorf("failed to generate secure key: %w", err)
	}

	// Format key based on configuration
	var keyString string
	switch km.config.KeyFormat {
	case "hex":
		keyString = hex.EncodeToString(keyBytes)
	case "base64":
		keyString = base64.StdEncoding.EncodeToString(keyBytes)
	case "base64url":
		keyString = base64.URLEncoding.EncodeToString(keyBytes)
	default:
		keyString = base64.URLEncoding.EncodeToString(keyBytes)
	}

	// Add prefix if configured
	if km.config.KeyPrefix != "" {
		keyString = km.config.KeyPrefix + keyString
	}

	// Create API key
	now := time.Now()
	apiKey := &APIKey{
		ID:          uuid.New().String(),
		Key:         keyString,
		Role:        role,
		CreatedAt:   now,
		ExpiresAt:   now.Add(expiry),
		Description: description,
		LastUsed:    time.Time{}, // Zero time indicates never used
		UsageCount:  0,
		Status:      "active",
	}

	// Store key
	km.storage.Keys[apiKey.ID] = apiKey

	// Save to storage
	if err := km.saveKeys(); err != nil {
		// Remove from memory if save failed
		delete(km.storage.Keys, apiKey.ID)
		return nil, fmt.Errorf("failed to save API key: %w", err)
	}

	// Log key generation
	km.logger.WithFields(logging.Fields{
		"key_id":      apiKey.ID,
		"role":        role,
		"expires_at":  apiKey.ExpiresAt,
		"description": description,
	}).Info("API key generated successfully")

	return apiKey, nil
}

// ValidateKey validates an API key and returns its metadata
func (km *APIKeyManager) ValidateKey(key string) (*APIKey, error) {
	km.mu.RLock()
	defer km.mu.RUnlock()

	// Find key by value
	for _, apiKey := range km.storage.Keys {
		if apiKey.Key == key {
			// Check if key is active
			if apiKey.Status != "active" {
				return nil, fmt.Errorf("key is not active (status: %s)", apiKey.Status)
			}

			// Check if key is expired
			if time.Now().After(apiKey.ExpiresAt) {
				// Mark as expired
				apiKey.Status = "expired"
				km.saveKeys() // Save status change
				return nil, fmt.Errorf("key has expired")
			}

			// Update usage tracking
			if km.config.UsageTracking {
				apiKey.LastUsed = time.Now()
				apiKey.UsageCount++
				km.saveKeys() // Save usage update
			}

			return apiKey, nil
		}
	}

	return nil, fmt.Errorf("invalid API key")
}

// RevokeKey revokes an API key by ID
func (km *APIKeyManager) RevokeKey(keyID string) error {
	km.mu.Lock()
	defer km.mu.Unlock()

	apiKey, exists := km.storage.Keys[keyID]
	if !exists {
		return fmt.Errorf("key not found: %s", keyID)
	}

	if apiKey.Status == "revoked" {
		return fmt.Errorf("key is already revoked")
	}

	// Revoke key
	apiKey.Status = "revoked"

	// Save changes
	if err := km.saveKeys(); err != nil {
		return fmt.Errorf("failed to save key revocation: %w", err)
	}

	// Log revocation
	km.logger.WithFields(logging.Fields{
		"key_id": keyID,
		"role":   apiKey.Role,
	}).Info("API key revoked successfully")

	return nil
}

// ListKeys returns all API keys, optionally filtered by role
func (km *APIKeyManager) ListKeys(role Role) ([]*APIKey, error) {
	km.mu.RLock()
	defer km.mu.RUnlock()

	var keys []*APIKey
	for _, apiKey := range km.storage.Keys {
		if role == 0 || apiKey.Role == role {
			// Create a copy to avoid exposing internal state
			keyCopy := *apiKey
			keys = append(keys, &keyCopy)
		}
	}

	return keys, nil
}

// RotateKeys rotates all keys for a specific role
func (km *APIKeyManager) RotateKeys(role Role, force bool) error {
	km.mu.Lock()
	defer km.mu.Unlock()

	// Find keys to rotate
	var keysToRotate []*APIKey
	for _, apiKey := range km.storage.Keys {
		if apiKey.Role == role && apiKey.Status == "active" {
			keysToRotate = append(keysToRotate, apiKey)
		}
	}

	if len(keysToRotate) == 0 {
		return fmt.Errorf("no active keys found for role %s", role)
	}

	// Rotate keys
	for _, apiKey := range keysToRotate {
		// Generate new key
		newKey, err := km.GenerateKey(apiKey.Role, time.Until(apiKey.ExpiresAt), apiKey.Description+" (rotated)")
		if err != nil {
			km.logger.WithError(err).WithField("key_id", apiKey.ID).Error("Failed to generate replacement key during rotation")
			continue
		}

		// Revoke old key
		apiKey.Status = "revoked"

		km.logger.WithFields(logging.Fields{
			"old_key_id": apiKey.ID,
			"new_key_id": newKey.ID,
			"role":       role,
		}).Info("API key rotated successfully")
	}

	// Save changes
	if err := km.saveKeys(); err != nil {
		return fmt.Errorf("failed to save key rotation: %w", err)
	}

	return nil
}

// CleanupExpiredKeys removes expired keys from storage
func (km *APIKeyManager) CleanupExpiredKeys() error {
	km.mu.Lock()
	defer km.mu.Unlock()

	now := time.Now()
	var expiredKeys []string

	for id, apiKey := range km.storage.Keys {
		if now.After(apiKey.ExpiresAt) && apiKey.Status == "active" {
			apiKey.Status = "expired"
			expiredKeys = append(expiredKeys, id)
		}
	}

	if len(expiredKeys) > 0 {
		// Save changes
		if err := km.saveKeys(); err != nil {
			return fmt.Errorf("failed to save expired key cleanup: %w", err)
		}

		km.logger.WithField("expired_count", fmt.Sprintf("%d", len(expiredKeys))).Info("Expired API keys cleaned up")
	}

	return nil
}

// loadKeys loads API keys from storage file
func (km *APIKeyManager) loadKeys() error {
	// Check if storage file exists
	if _, err := os.Stat(km.storagePath); os.IsNotExist(err) {
		// Create directory if it doesn't exist
		dir := filepath.Dir(km.storagePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create storage directory: %w", err)
		}
		return nil // No existing keys to load
	}

	// Read storage file
	data, err := os.ReadFile(km.storagePath)
	if err != nil {
		return fmt.Errorf("failed to read storage file: %w", err)
	}

	// Decrypt if encryption is enabled
	if km.config.EncryptionKey != "" {
		// TODO: Implement encryption/decryption
		km.logger.Warn("Key encryption not yet implemented, storing keys in plain text")
	}

	// Unmarshal JSON
	var storage APIKeyStorage
	if err := json.Unmarshal(data, &storage); err != nil {
		return fmt.Errorf("failed to unmarshal storage data: %w", err)
	}

	// Validate loaded keys
	for id, apiKey := range storage.Keys {
		if apiKey.ID == "" {
			apiKey.ID = id // Ensure ID is set
		}
		if apiKey.Status == "" {
			apiKey.Status = "active" // Default status
		}
	}

	km.storage = &storage
	return nil
}

// saveKeys saves API keys to storage file
func (km *APIKeyManager) saveKeys() error {
	// Marshal to JSON
	data, err := json.MarshalIndent(km.storage, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal storage data: %w", err)
	}

	// Encrypt if encryption is enabled
	if km.config.EncryptionKey != "" {
		// TODO: Implement encryption/decryption
		km.logger.Warn("Key encryption not yet implemented, storing keys in plain text")
	}

	// Write to file
	if err := os.WriteFile(km.storagePath, data, 0600); err != nil {
		return fmt.Errorf("failed to write storage file: %w", err)
	}

	return nil
}

// countKeysByRole counts active keys for a specific role
func (km *APIKeyManager) countKeysByRole(role Role) int {
	count := 0
	for _, apiKey := range km.storage.Keys {
		if apiKey.Role == role && apiKey.Status == "active" {
			count++
		}
	}
	return count
}

// GetStats returns statistics about API key usage
func (km *APIKeyManager) GetStats() map[string]interface{} {
	km.mu.RLock()
	defer km.mu.RUnlock()

	stats := map[string]interface{}{
		"total_keys":   len(km.storage.Keys),
		"active_keys":  0,
		"revoked_keys": 0,
		"expired_keys": 0,
		"keys_by_role": make(map[string]int),
	}

	for _, apiKey := range km.storage.Keys {
		switch apiKey.Status {
		case "active":
			stats["active_keys"] = stats["active_keys"].(int) + 1
		case "revoked":
			stats["revoked_keys"] = stats["revoked_keys"].(int) + 1
		case "expired":
			stats["expired_keys"] = stats["expired_keys"].(int) + 1
		}

		// Count by role
		roleCount := stats["keys_by_role"].(map[string]int)
		roleCount[string(apiKey.Role)]++
	}

	return stats
}
