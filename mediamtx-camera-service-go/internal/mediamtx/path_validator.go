/*
MediaMTX Path Validator Implementation

Requirements Coverage:
- REQ-MTX-001: MediaMTX service integration
- REQ-MTX-002: Stream management capabilities
- REQ-MTX-003: Path creation and deletion
- REQ-MTX-004: Health monitoring

Test Categories: Unit/Integration
API Documentation Reference: docs/api/json_rpc_methods.md
*/

package mediamtx

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
)

// PathValidator handles runtime path validation and fallback
type PathValidator struct {
	config           *config.Config
	logger           *logging.Logger
	validationCache  map[string]*PathValidationResult
	cacheMutex       sync.RWMutex
	validationPeriod time.Duration
}

// PathValidationResult represents the result of path validation
type PathValidationResult struct {
	Path         string
	IsValid      bool
	IsWritable   bool
	Error        error
	ValidatedAt  time.Time
	FallbackPath string
}

// NewPathValidator creates a new path validator instance
func NewPathValidator(config *config.Config, logger *logging.Logger) *PathValidator {
	return &PathValidator{
		config:           config,
		logger:           logger,
		validationCache:  make(map[string]*PathValidationResult),
		validationPeriod: 5 * time.Minute, // Re-validate every 5 minutes
	}
}

// ValidateRecordingPath validates and returns a usable recording path
func (pv *PathValidator) ValidateRecordingPath(ctx context.Context) (*PathValidationResult, error) {
	primaryPath := pv.config.MediaMTX.RecordingsPath
	// Use centralized path management - no fallback needed as paths are validated during config loading
	return pv.validatePathWithFallback(ctx, "recordings", primaryPath, "")
}

// ValidateSnapshotPath validates and returns a usable snapshot path
func (pv *PathValidator) ValidateSnapshotPath(ctx context.Context) (*PathValidationResult, error) {
	primaryPath := pv.config.MediaMTX.SnapshotsPath
	// With centralized path management, no fallback needed - paths are validated during config loading
	return pv.validatePathWithFallback(ctx, "snapshots", primaryPath, "")
}

// validatePathWithFallback attempts primary path, falls back if needed
func (pv *PathValidator) validatePathWithFallback(ctx context.Context, name, primaryPath, fallbackPath string) (*PathValidationResult, error) {
	// Check cache first
	pv.cacheMutex.RLock()
	cached, exists := pv.validationCache[primaryPath]
	pv.cacheMutex.RUnlock()

	if exists && time.Since(cached.ValidatedAt) < pv.validationPeriod {
		if cached.IsValid {
			return cached, nil
		}
	}

	// Validate primary path
	result := pv.validateSinglePath(primaryPath)

	if !result.IsValid && fallbackPath != "" {
		pv.logger.WithFields(logging.Fields{
			"primary_path":  primaryPath,
			"fallback_path": fallbackPath,
			"error":         result.Error,
		}).Warn("Primary path validation failed, trying fallback")

		// Try fallback
		fallbackResult := pv.validateSinglePath(fallbackPath)
		if fallbackResult.IsValid {
			result.FallbackPath = fallbackPath
			result.Path = fallbackPath
			result.IsValid = true
			result.IsWritable = fallbackResult.IsWritable
			result.Error = fmt.Errorf("using fallback path: %s (primary failed: %w)", fallbackPath, result.Error)
		} else {
			result.Error = fmt.Errorf("both primary and fallback paths failed: primary=%w, fallback=%w",
				result.Error, fallbackResult.Error)
		}
	}

	// Update cache
	pv.cacheMutex.Lock()
	pv.validationCache[primaryPath] = result
	pv.cacheMutex.Unlock()

	if !result.IsValid {
		return nil, result.Error
	}

	return result, nil
}

// validateSinglePath validates a single path
func (pv *PathValidator) validateSinglePath(path string) *PathValidationResult {
	result := &PathValidationResult{
		Path:        path,
		ValidatedAt: time.Now(),
	}

	// Check if path exists
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		// Try to create it
		if err := os.MkdirAll(path, 0755); err != nil {
			result.Error = fmt.Errorf("path does not exist and cannot be created: %w", err)
			return result
		}
		info, _ = os.Stat(path)
	} else if err != nil {
		result.Error = fmt.Errorf("cannot access path: %w", err)
		return result
	}

	// Check if it's a directory
	if info != nil && !info.IsDir() {
		result.Error = fmt.Errorf("path is not a directory")
		return result
	}

	// Check write permission
	testFile := filepath.Join(path, fmt.Sprintf(".write_test_%d_%d", os.Getpid(), time.Now().UnixNano()))
	file, err := os.Create(testFile)
	if err != nil {
		result.Error = fmt.Errorf("path is not writable: %w", err)
		result.IsValid = false
		result.IsWritable = false
		return result
	}
	file.Close()
	os.Remove(testFile)

	result.IsValid = true
	result.IsWritable = true
	return result
}

// StartPeriodicValidation starts background validation
func (pv *PathValidator) StartPeriodicValidation(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(pv.validationPeriod)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				pv.revalidateAll(ctx)
			}
		}
	}()
}

// revalidateAll clears cache and revalidates all paths
func (pv *PathValidator) revalidateAll(ctx context.Context) {
	pv.cacheMutex.Lock()
	pv.validationCache = make(map[string]*PathValidationResult)
	pv.cacheMutex.Unlock()

	// Trigger revalidation
	pv.ValidateRecordingPath(ctx)
	pv.ValidateSnapshotPath(ctx)
}
