package mediamtx

import (
	"os"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
)

func TestPathFallback(t *testing.T) {
	// Use fixture-based test helper following Path Management Solution
	helper, ctx := SetupMediaMTXTest(t)

	// Get configured paths from fixture
	recordingsPath := helper.GetConfiguredRecordingPath()

	// Create path validator using fixture configuration
	configManager := helper.GetConfigManager()
	cfg := configManager.GetConfig()
	logger := helper.GetLogger()
	validator := NewPathValidator(cfg, logger)

	// Test recording path validation - MINIMAL: Helper provides standard context
	result, err := validator.ValidateRecordingPath(ctx)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result, got nil")
	}

	if result.Path != recordingsPath {
		t.Errorf("Expected recordings path %s, got %s", recordingsPath, result.Path)
	}

	// With centralized path management, no fallback path is used
	if result.FallbackPath != "" {
		t.Error("Expected no fallback path with centralized path management")
	}
}

func TestPathValidatorCaching(t *testing.T) {
	// Use fixture-based test helper following Path Management Solution
	helper, ctx := SetupMediaMTXTest(t)

	// Create path validator using fixture configuration
	configManager := helper.GetConfigManager()
	cfg := configManager.GetConfig()
	logger := helper.GetLogger()

	// Create path validator with short validation period
	validator := &PathValidator{
		config:           cfg,
		logger:           logger,
		validationCache:  make(map[string]*PathValidationResult),
		validationPeriod: TestValidationPeriodShort, // Very short for testing
	}

	// First validation
	result1, err1 := validator.ValidateRecordingPath(ctx)
	if err1 != nil {
		t.Fatalf("First validation failed: %v", err1)
	}

	// Second validation (should use cache)
	result2, err2 := validator.ValidateRecordingPath(ctx)
	if err2 != nil {
		t.Fatalf("Second validation failed: %v", err2)
	}

	// Results should be the same (cached)
	if result1.ValidatedAt != result2.ValidatedAt {
		t.Error("Expected cached result, but got different validation times")
	}

	// Wait for cache to expire using proper synchronization
	select {
	case <-time.After(TestValidationPeriodLong):
		// Cache should be expired now
	case <-ctx.Done():
		// Context cancelled, exit early
		return
	}

	// Third validation (should re-validate)
	result3, err3 := validator.ValidateRecordingPath(ctx)
	if err3 != nil {
		t.Fatalf("Third validation failed: %v", err3)
	}

	// Results should be different (re-validated)
	if result1.ValidatedAt == result3.ValidatedAt {
		t.Error("Expected re-validated result, but got same validation time")
	}
}

func TestPathValidatorErrorHandling(t *testing.T) {
	_, ctx := SetupMediaMTXTest(t)

	// Create test config with non-existent paths
	cfg := &config.Config{
		MediaMTX: config.MediaMTXConfig{
			RecordingsPath: "/nonexistent/primary",
		},
		Storage: config.StorageConfig{
			FallbackPath: "/nonexistent/fallback",
		},
	}

	// Create logger
	logger := logging.GetLogger("mediamtx.path_validator") // Component-specific logger

	// Create path validator
	validator := NewPathValidator(cfg, logger)

	// Test recording path validation (should fail)
	_, err := validator.ValidateRecordingPath(ctx)
	if err == nil {
		t.Error("Expected error for non-existent paths, got none")
	}

	// Test snapshot path validation (should fail)
	_, err = validator.ValidateSnapshotPath(ctx)
	if err == nil {
		t.Error("Expected error for non-existent paths, got none")
	}
}

func TestPathValidatorSinglePathValidation(t *testing.T) {
	// Use centralized path management instead of hardcoded paths
	helper := SetupMediaMTXTestHelperOnly(t)

	// Create test directory using configured path
	testDir := helper.GetConfig().TestDataDir
	os.MkdirAll(testDir, 0755)
	defer os.RemoveAll(testDir)

	// Create logger
	logger := logging.GetLogger("mediamtx.path_validator") // Component-specific logger

	// Create path validator
	validator := &PathValidator{
		logger: logger,
	}

	// Test valid path
	result := validator.validateSinglePath(testDir)
	if !result.IsValid {
		t.Errorf("Expected valid path, got error: %v", result.Error)
	}

	if !result.IsWritable {
		t.Error("Expected writable path")
	}

	// Test invalid path
	result = validator.validateSinglePath("/nonexistent/path")
	if result.IsValid {
		t.Error("Expected invalid path")
	}

	if result.Error == nil {
		t.Error("Expected error for invalid path")
	}
}
