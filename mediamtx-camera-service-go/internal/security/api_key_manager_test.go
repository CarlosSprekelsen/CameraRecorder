/*
API Key Manager Unit Tests

Requirements Coverage:
- REQ-SEC-014: Key Management
- REQ-SEC-015: Production API Key Management

Test Categories: Unit
API Documentation Reference: docs/api/json_rpc_methods.md

Unit tests for API Key Manager following existing testing patterns.
Tests key generation, validation, revocation, and lifecycle management.
*/

package security

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAPIKeyManager(t *testing.T) {
	tests := []struct {
		name        string
		config      *config.APIKeyManagementConfig
		logger      *logging.Logger
		expectError bool
	}{
		{
			name:        "valid configuration",
			config:      &config.APIKeyManagementConfig{StoragePath: "/tmp/test-keys.json"},
			logger:      logging.GetLogger("test"),
			expectError: false,
		},
		{
			name:        "nil configuration",
			config:      nil,
			logger:      logging.GetLogger("test"),
			expectError: true,
		},
		{
			name:        "nil logger",
			config:      &config.APIKeyManagementConfig{StoragePath: "/tmp/test-keys.json"},
			logger:      nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager, err := NewAPIKeyManager(tt.config, tt.logger)
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, manager)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, manager)
				assert.Equal(t, tt.config, manager.config)
				assert.Equal(t, tt.logger, manager.logger)
			}
		})
	}
}

func TestAPIKeyManager_GenerateKey(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	storagePath := filepath.Join(tempDir, "test-keys.json")
	
	config := &config.APIKeyManagementConfig{
		StoragePath:     storagePath,
		KeyLength:       16, // Smaller for testing
		KeyPrefix:       "test_",
		KeyFormat:       "base64url",
		MaxKeysPerRole:  5,
		UsageTracking:   true,
		AuditLogging:    true,
	}
	
	logger := logging.GetLogger("test")
	manager, err := NewAPIKeyManager(config, logger)
	require.NoError(t, err)

	tests := []struct {
		name        string
		role        Role
		expiry      time.Duration
		description string
		expectError bool
	}{
		{
			name:        "valid admin key",
			role:        RoleAdmin,
			expiry:      90 * 24 * time.Hour,
			description: "Test admin key",
			expectError: false,
		},
		{
			name:        "valid operator key",
			role:        RoleOperator,
			expiry:      30 * 24 * time.Hour,
			description: "Test operator key",
			expectError: false,
		},
		{
			name:        "valid viewer key",
			role:        RoleViewer,
			expiry:      7 * 24 * time.Hour,
			description: "Test viewer key",
			expectError: false,
		},
		{
			name:        "empty role",
			role:        "",
			expiry:      90 * 24 * time.Hour,
			description: "Test key",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiKey, err := manager.GenerateKey(tt.role, tt.expiry, tt.description)
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, apiKey)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, apiKey)
				assert.NotEmpty(t, apiKey.ID)
				assert.NotEmpty(t, apiKey.Key)
				assert.Equal(t, tt.role, apiKey.Role)
				assert.Equal(t, tt.description, apiKey.Description)
				assert.Equal(t, "active", apiKey.Status)
				assert.True(t, apiKey.KeyHasPrefix("test_"))
				assert.True(t, apiKey.ExpiresAt.After(time.Now()))
			}
		})
	}
}

func TestAPIKeyManager_ValidateKey(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	storagePath := filepath.Join(tempDir, "test-keys.json")
	
	config := &config.APIKeyManagementConfig{
		StoragePath:   storagePath,
		KeyLength:     16,
		KeyPrefix:     "test_",
		KeyFormat:     "base64url",
		UsageTracking: true,
	}
	
	logger := logging.GetLogger("test")
	manager, err := NewAPIKeyManager(config, logger)
	require.NoError(t, err)

	// Generate a test key
	apiKey, err := manager.GenerateKey(RoleAdmin, 90*24*time.Hour, "Test key")
	require.NoError(t, err)

	tests := []struct {
		name        string
		key         string
		expectError bool
	}{
		{
			name:        "valid key",
			key:         apiKey.Key,
			expectError: false,
		},
		{
			name:        "invalid key",
			key:         "invalid_key",
			expectError: true,
		},
		{
			name:        "empty key",
			key:         "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validatedKey, err := manager.ValidateKey(tt.key)
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, validatedKey)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, validatedKey)
				assert.Equal(t, apiKey.ID, validatedKey.ID)
				assert.Equal(t, apiKey.Key, validatedKey.Key)
				assert.Equal(t, apiKey.Role, validatedKey.Role)
			}
		})
	}
}

func TestAPIKeyManager_RevokeKey(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	storagePath := filepath.Join(tempDir, "test-keys.json")
	
	config := &config.APIKeyManagementConfig{
		StoragePath: storagePath,
		KeyLength:   16,
		KeyPrefix:   "test_",
		KeyFormat:   "base64url",
	}
	
	logger := logging.GetLogger("test")
	manager, err := NewAPIKeyManager(config, logger)
	require.NoError(t, err)

	// Generate a test key
	apiKey, err := manager.GenerateKey(RoleAdmin, 90*24*time.Hour, "Test key")
	require.NoError(t, err)

	tests := []struct {
		name        string
		keyID       string
		expectError bool
	}{
		{
			name:        "valid key ID",
			keyID:       apiKey.ID,
			expectError: false,
		},
		{
			name:        "invalid key ID",
			keyID:       "invalid-id",
			expectError: true,
		},
		{
			name:        "empty key ID",
			keyID:       "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.RevokeKey(tt.keyID)
			
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				
				// Verify key is revoked
				keys, err := manager.ListKeys("")
				require.NoError(t, err)
				
				var revokedKey *APIKey
				for _, key := range keys {
					if key.ID == tt.keyID {
						revokedKey = key
						break
					}
				}
				
				assert.NotNil(t, revokedKey)
				assert.Equal(t, "revoked", revokedKey.Status)
			}
		})
	}
}

func TestAPIKeyManager_ListKeys(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	storagePath := filepath.Join(tempDir, "test-keys.json")
	
	config := &config.APIKeyManagementConfig{
		StoragePath: storagePath,
		KeyLength:   16,
		KeyPrefix:   "test_",
		KeyFormat:   "base64url",
	}
	
	logger := logging.GetLogger("test")
	manager, err := NewAPIKeyManager(config, logger)
	require.NoError(t, err)

	// Generate test keys
	adminKey, err := manager.GenerateKey(RoleAdmin, 90*24*time.Hour, "Admin key")
	require.NoError(t, err)
	
	operatorKey, err := manager.GenerateKey(RoleOperator, 30*24*time.Hour, "Operator key")
	require.NoError(t, err)
	
	viewerKey, err := manager.GenerateKey(RoleViewer, 7*24*time.Hour, "Viewer key")
	require.NoError(t, err)

	tests := []struct {
		name     string
		role     Role
		expected int
	}{
		{
			name:     "all keys",
			role:     "",
			expected: 3,
		},
		{
			name:     "admin keys only",
			role:     RoleAdmin,
			expected: 1,
		},
		{
			name:     "operator keys only",
			role:     RoleOperator,
			expected: 1,
		},
		{
			name:     "viewer keys only",
			role:     RoleViewer,
			expected: 1,
		},
		{
			name:     "non-existent role",
			role:     "nonexistent",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keys, err := manager.ListKeys(tt.role)
			assert.NoError(t, err)
			assert.Len(t, keys, tt.expected)
			
			// Verify all returned keys have the correct role
			for _, key := range keys {
				if tt.role != "" {
					assert.Equal(t, tt.role, key.Role)
				}
			}
		})
	}
}

func TestAPIKeyManager_RotateKeys(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	storagePath := filepath.Join(tempDir, "test-keys.json")
	
	config := &config.APIKeyManagementConfig{
		StoragePath: storagePath,
		KeyLength:   16,
		KeyPrefix:   "test_",
		KeyFormat:   "base64url",
	}
	
	logger := logging.GetLogger("test")
	manager, err := NewAPIKeyManager(config, logger)
	require.NoError(t, err)

	// Generate test keys
	adminKey1, err := manager.GenerateKey(RoleAdmin, 90*24*time.Hour, "Admin key 1")
	require.NoError(t, err)
	
	adminKey2, err := manager.GenerateKey(RoleAdmin, 90*24*time.Hour, "Admin key 2")
	require.NoError(t, err)

	tests := []struct {
		name        string
		role        Role
		force       bool
		expectError bool
	}{
		{
			name:        "rotate admin keys",
			role:        RoleAdmin,
			force:       false,
			expectError: false,
		},
		{
			name:        "rotate non-existent role",
			role:        "nonexistent",
			force:       false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.RotateKeys(tt.role, tt.force)
			
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				
				// Verify old keys are revoked
				keys, err := manager.ListKeys(tt.role)
				require.NoError(t, err)
				
				// Should have new keys (rotated) and old keys should be revoked
				var activeKeys, revokedKeys int
				for _, key := range keys {
					if key.Status == "active" {
						activeKeys++
					} else if key.Status == "revoked" {
						revokedKeys++
					}
				}
				
				assert.Greater(t, activeKeys, 0, "Should have active keys after rotation")
				assert.Greater(t, revokedKeys, 0, "Should have revoked keys after rotation")
			}
		})
	}
}

func TestAPIKeyManager_CleanupExpiredKeys(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	storagePath := filepath.Join(tempDir, "test-keys.json")
	
	config := &config.APIKeyManagementConfig{
		StoragePath: storagePath,
		KeyLength:   16,
		KeyPrefix:   "test_",
		KeyFormat:   "base64url",
	}
	
	logger := logging.GetLogger("test")
	manager, err := NewAPIKeyManager(config, logger)
	require.NoError(t, err)

	// Generate a key that's already expired
	expiredKey, err := manager.GenerateKey(RoleAdmin, -1*time.Hour, "Expired key")
	require.NoError(t, err)

	// Manually set expiry to past
	expiredKey.ExpiresAt = time.Now().Add(-1 * time.Hour)
	manager.storage.Keys[expiredKey.ID] = expiredKey

	// Generate a valid key
	validKey, err := manager.GenerateKey(RoleAdmin, 90*24*time.Hour, "Valid key")
	require.NoError(t, err)

	// Cleanup expired keys
	err = manager.CleanupExpiredKeys()
	assert.NoError(t, err)

	// Verify expired key is marked as expired
	keys, err := manager.ListKeys("")
	require.NoError(t, err)

	var expiredKeyFound, validKeyFound bool
	for _, key := range keys {
		if key.ID == expiredKey.ID {
			expiredKeyFound = true
			assert.Equal(t, "expired", key.Status)
		}
		if key.ID == validKey.ID {
			validKeyFound = true
			assert.Equal(t, "active", key.Status)
		}
	}

	assert.True(t, expiredKeyFound, "Expired key should be found")
	assert.True(t, validKeyFound, "Valid key should be found")
}

func TestAPIKeyManager_GetStats(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	storagePath := filepath.Join(tempDir, "test-keys.json")
	
	config := &config.APIKeyManagementConfig{
		StoragePath: storagePath,
		KeyLength:   16,
		KeyPrefix:   "test_",
		KeyFormat:   "base64url",
	}
	
	logger := logging.GetLogger("test")
	manager, err := NewAPIKeyManager(config, logger)
	require.NoError(t, err)

	// Generate test keys
	adminKey, err := manager.GenerateKey(RoleAdmin, 90*24*time.Hour, "Admin key")
	require.NoError(t, err)
	
	operatorKey, err := manager.GenerateKey(RoleOperator, 30*24*time.Hour, "Operator key")
	require.NoError(t, err)

	// Revoke one key
	err = manager.RevokeKey(operatorKey.ID)
	require.NoError(t, err)

	// Get statistics
	stats := manager.GetStats()

	// Verify statistics
	assert.Equal(t, 2, stats["total_keys"])
	assert.Equal(t, 1, stats["active_keys"])
	assert.Equal(t, 1, stats["revoked_keys"])
	assert.Equal(t, 0, stats["expired_keys"])

	keysByRole := stats["keys_by_role"].(map[string]int)
	assert.Equal(t, 1, keysByRole["admin"])
	assert.Equal(t, 1, keysByRole["operator"])
}

func TestAPIKeyManager_MaxKeysPerRole(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	storagePath := filepath.Join(tempDir, "test-keys.json")
	
	config := &config.APIKeyManagementConfig{
		StoragePath:    storagePath,
		KeyLength:      16,
		KeyPrefix:      "test_",
		KeyFormat:      "base64url",
		MaxKeysPerRole: 2,
	}
	
	logger := logging.GetLogger("test")
	manager, err := NewAPIKeyManager(config, logger)
	require.NoError(t, err)

	// Generate keys up to the limit
	key1, err := manager.GenerateKey(RoleAdmin, 90*24*time.Hour, "Admin key 1")
	require.NoError(t, err)
	
	key2, err := manager.GenerateKey(RoleAdmin, 90*24*time.Hour, "Admin key 2")
	require.NoError(t, err)

	// Try to generate one more key (should fail)
	_, err = manager.GenerateKey(RoleAdmin, 90*24*time.Hour, "Admin key 3")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "maximum keys per role exceeded")

	// Revoke one key
	err = manager.RevokeKey(key1.ID)
	require.NoError(t, err)

	// Now should be able to generate a new key
	key3, err := manager.GenerateKey(RoleAdmin, 90*24*time.Hour, "Admin key 3")
	assert.NoError(t, err)
	assert.NotNil(t, key3)
}

func TestAPIKeyManager_UsageTracking(t *testing.T) {
	// Setup
	tempDir := t.TempDir()
	storagePath := filepath.Join(tempDir, "test-keys.json")
	
	config := &config.APIKeyManagementConfig{
		StoragePath:   storagePath,
		KeyLength:     16,
		KeyPrefix:     "test_",
		KeyFormat:     "base64url",
		UsageTracking: true,
	}
	
	logger := logging.GetLogger("test")
	manager, err := NewAPIKeyManager(config, logger)
	require.NoError(t, err)

	// Generate a test key
	apiKey, err := manager.GenerateKey(RoleAdmin, 90*24*time.Hour, "Test key")
	require.NoError(t, err)

	// Validate key multiple times
	for i := 0; i < 3; i++ {
		validatedKey, err := manager.ValidateKey(apiKey.Key)
		require.NoError(t, err)
		assert.Equal(t, int64(i+1), validatedKey.UsageCount)
		assert.True(t, validatedKey.LastUsed.After(apiKey.CreatedAt))
	}
}

// Helper methods for testing
func (ak *APIKey) KeyHasPrefix(prefix string) bool {
	return len(ak.Key) > len(prefix) && ak.Key[:len(prefix)] == prefix
}
