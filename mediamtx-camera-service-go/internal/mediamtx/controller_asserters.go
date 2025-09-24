/*
Controller Test Asserters - Eliminate Massive Duplication

This file provides domain-specific asserters for controller tests that eliminate
the massive duplication found in controller_test.go (1,557 lines).

Duplication Patterns Eliminated:
- SetupMediaMTXTest + GetReadyController (41 times)
- Camera ID retrieval (22+ times)
- Progressive Readiness pattern (15+ times)
- File validation (10+ times)
- Health checking (8+ times)

Usage:
    asserter := NewControllerAsserter(t)
    defer asserter.Cleanup()
    // Test-specific logic only
    asserter.AssertHealthResponse(controller.GetHealth(ctx))
*/

package mediamtx

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ControllerAsserter encapsulates all controller test patterns
type ControllerAsserter struct {
	t          *testing.T
	helper     *MediaMTXTestHelper
	ctx        context.Context
	cancel     context.CancelFunc
	controller MediaMTXController
}

// NewControllerAsserter creates a new controller asserter with full setup
// Eliminates: helper, _ := SetupMediaMTXTest(t) + GetReadyController pattern
func NewControllerAsserter(t *testing.T) *ControllerAsserter {
	helper, _ := SetupMediaMTXTest(t)
	controller, ctx, cancel := helper.GetReadyController(t)

	return &ControllerAsserter{
		t:          t,
		helper:     helper,
		ctx:        ctx,
		cancel:     cancel,
		controller: controller,
	}
}

// Cleanup must be called in test cleanup (defer asserter.Cleanup())
func (ca *ControllerAsserter) Cleanup() {
	ca.cancel()
	ca.controller.Stop(ca.ctx)
}

// GetReadyController returns the ready controller (eliminates setup duplication)
func (ca *ControllerAsserter) GetReadyController() MediaMTXController {
	return ca.controller
}

// GetContext returns the test context
func (ca *ControllerAsserter) GetContext() context.Context {
	return ca.ctx
}

// GetHelper returns the test helper
func (ca *ControllerAsserter) GetHelper() *MediaMTXTestHelper {
	return ca.helper
}

// MustGetCameraID gets camera ID with error handling (eliminates 22+ duplications)
func (ca *ControllerAsserter) MustGetCameraID() string {
	cameraID, err := ca.helper.GetAvailableCameraIdentifierFromController(ca.ctx, ca.controller)
	require.NoError(ca.t, err, "Must have available camera for test")
	return cameraID
}

// AssertHealthResponse validates health response (eliminates 8+ duplications)
func (ca *ControllerAsserter) AssertHealthResponse(health *GetHealthResponse, err error) {
	ca.helper.AssertHealthResponse(ca.t, health, err, "Controller health check")
}

// AssertRecordingResponse validates recording response
func (ca *ControllerAsserter) AssertRecordingResponse(response *StartRecordingResponse, err error) {
	ca.helper.AssertRecordingResponse(ca.t, response, err)
}

// AssertSnapshotResponse validates snapshot response
func (ca *ControllerAsserter) AssertSnapshotResponse(response *TakeSnapshotResponse, err error) {
	ca.helper.AssertSnapshotResponse(ca.t, response, err)
}

// AssertFileExists validates file exists with size check (eliminates 10+ duplications)
func (ca *ControllerAsserter) AssertFileExists(filePath string, minSize int64, description string) {
	require.FileExists(ca.t, filePath, "%s must exist: %s", description, filePath)

	fileInfo, err := os.Stat(filePath)
	require.NoError(ca.t, err, "Should be able to stat %s", description)
	assert.Greater(ca.t, fileInfo.Size(), minSize, "%s must have meaningful content (>%d bytes)", description, minSize)

	ca.t.Logf("✅ %s validated: %s (%d bytes)", description, filePath, fileInfo.Size())
}

// AssertRecordingFileExists validates recording file (uses universal constants)
func (ca *ControllerAsserter) AssertRecordingFileExists(filePath string) {
	ca.AssertFileExists(filePath, int64(testutils.UniversalMinRecordingFileSize), "Recording file")
}

// AssertSnapshotFileExists validates snapshot file (uses universal constants)
func (ca *ControllerAsserter) AssertSnapshotFileExists(filePath string) {
	ca.AssertFileExists(filePath, int64(testutils.UniversalMinSnapshotFileSize), "Snapshot file")
}

// ============================================================================
// RECORDING LIFECYCLE ASSERTERS
// ============================================================================

// RecordingLifecycleAsserter handles complete recording lifecycle testing
type RecordingLifecycleAsserter struct {
	*ControllerAsserter
}

// NewRecordingLifecycleAsserter creates a recording-focused asserter
func NewRecordingLifecycleAsserter(t *testing.T) *RecordingLifecycleAsserter {
	return &RecordingLifecycleAsserter{
		ControllerAsserter: NewControllerAsserter(t),
	}
}

// AssertCompleteRecordingLifecycle performs start → record → stop → validate file
// Eliminates 50+ lines of recording test duplication
func (rla *RecordingLifecycleAsserter) AssertCompleteRecordingLifecycle(cameraID string, duration time.Duration) *StartRecordingResponse {
	// Use the existing helper method we created earlier
	return rla.helper.MustStartAndStopRecording(rla.t, rla.ctx, rla.controller, cameraID, duration)
}

// AssertRecordingStart validates recording start only
func (rla *RecordingLifecycleAsserter) AssertRecordingStart(cameraID string, options *PathConf) *StartRecordingResponse {
	// Use Progressive Readiness pattern
	result := testutils.TestProgressiveReadiness(rla.t, func() (*StartRecordingResponse, error) {
		return rla.controller.StartRecording(rla.ctx, cameraID, options)
	}, rla.controller, "StartRecording")

	require.NoError(rla.t, result.Error, "Recording must start")
	require.NotNil(rla.t, result.Result, "Recording response must not be nil")

	if result.UsedFallback {
		rla.t.Log("⚠️  PROGRESSIVE READINESS FALLBACK: Start needed readiness event")
	} else {
		rla.t.Log("✅ PROGRESSIVE READINESS: Start succeeded immediately")
	}

	return result.Result
}

// AssertRecordingStop validates recording stop only
func (rla *RecordingLifecycleAsserter) AssertRecordingStop(cameraID string) *StopRecordingResponse {
	// Use Progressive Readiness pattern
	result := testutils.TestProgressiveReadiness(rla.t, func() (*StopRecordingResponse, error) {
		return rla.controller.StopRecording(rla.ctx, cameraID)
	}, rla.controller, "StopRecording")

	require.NoError(rla.t, result.Error, "Recording must stop")
	require.NotNil(rla.t, result.Result, "Stop response must not be nil")

	if result.UsedFallback {
		rla.t.Log("⚠️  PROGRESSIVE READINESS FALLBACK: Stop needed readiness event")
	} else {
		rla.t.Log("✅ PROGRESSIVE READINESS: Stop succeeded immediately")
	}

	return result.Result
}

// ============================================================================
// SNAPSHOT ASSERTERS
// ============================================================================

// SnapshotAsserter handles snapshot testing
type SnapshotAsserter struct {
	*ControllerAsserter
}

// NewSnapshotAsserter creates a snapshot-focused asserter
func NewSnapshotAsserter(t *testing.T) *SnapshotAsserter {
	return &SnapshotAsserter{
		ControllerAsserter: NewControllerAsserter(t),
	}
}

// AssertSnapshotCapture performs complete snapshot capture with validation
// Eliminates 30+ lines of snapshot test duplication
func (sa *SnapshotAsserter) AssertSnapshotCapture(cameraID string, options *SnapshotOptions) *TakeSnapshotResponse {
	// Use Progressive Readiness pattern
	result := testutils.TestProgressiveReadiness(sa.t, func() (*TakeSnapshotResponse, error) {
		return sa.controller.TakeAdvancedSnapshot(sa.ctx, cameraID, options)
	}, sa.controller, "TakeSnapshot")

	require.NoError(sa.t, result.Error, "Snapshot must succeed")
	require.NotNil(sa.t, result.Result, "Snapshot response must not be nil")

	if result.UsedFallback {
		sa.t.Log("⚠️  PROGRESSIVE READINESS FALLBACK: Snapshot needed readiness event")
	} else {
		sa.t.Log("✅ PROGRESSIVE READINESS: Snapshot succeeded immediately")
	}

	// Validate file creation
	sa.AssertSnapshotFileExists(result.Result.FilePath)

	return result.Result
}

// ============================================================================
// HEALTH ASSERTERS
// ============================================================================

// HealthAsserter handles health testing
type HealthAsserter struct {
	*ControllerAsserter
}

// NewHealthAsserter creates a health-focused asserter
func NewHealthAsserter(t *testing.T) *HealthAsserter {
	return &HealthAsserter{
		ControllerAsserter: NewControllerAsserter(t),
	}
}

// AssertHealthCheck validates health endpoint
func (ha *HealthAsserter) AssertHealthCheck() *GetHealthResponse {
	health, err := ha.controller.GetHealth(ha.ctx)
	ha.AssertHealthResponse(health, err)
	return health
}

// AssertMetricsCheck validates metrics endpoint
func (ha *HealthAsserter) AssertMetricsCheck() *GetMetricsResponse {
	metrics, err := ha.controller.GetMetrics(ha.ctx)
	require.NoError(ha.t, err, "Metrics should be available")
	require.NotNil(ha.t, metrics, "Metrics response should not be nil")
	return metrics
}

// ============================================================================
// STREAM ASSERTER - Stream Management Operations
// ============================================================================

// StreamAsserter encapsulates stream management test patterns
type StreamAsserter struct {
	t          *testing.T
	helper     *MediaMTXTestHelper
	controller MediaMTXController
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewStreamAsserter creates a new stream asserter with a ready controller
func NewStreamAsserter(t *testing.T) *StreamAsserter {
	helper, _ := SetupMediaMTXTest(t)
	controller, ctx, cancel := helper.GetReadyController(t)
	return &StreamAsserter{
		t:          t,
		helper:     helper,
		controller: controller,
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Cleanup stops the controller and cancels the context
func (sa *StreamAsserter) Cleanup() {
	sa.controller.Stop(sa.ctx)
	sa.cancel()
}

// GetHelper returns the underlying MediaMTXTestHelper
func (sa *StreamAsserter) GetHelper() *MediaMTXTestHelper {
	return sa.helper
}

// GetReadyController returns the ready controller
func (sa *StreamAsserter) GetReadyController() MediaMTXController {
	return sa.controller
}

// GetContext returns the test context
func (sa *StreamAsserter) GetContext() context.Context {
	return sa.ctx
}

// MustGetCameraID retrieves an available camera ID or fails the test
func (sa *StreamAsserter) MustGetCameraID() string {
	cameraID, err := sa.helper.GetAvailableCameraIdentifierFromController(sa.ctx, sa.controller)
	require.NoError(sa.t, err, "Must have available camera for test")
	return cameraID
}

// AssertGetPaths gets paths and validates the response
func (sa *StreamAsserter) AssertGetPaths() []*Path {
	result := testutils.TestProgressiveReadiness(sa.t, func() ([]*Path, error) {
		return sa.controller.GetPaths(sa.ctx)
	}, sa.controller, "GetPaths")

	require.NoError(sa.t, result.Error, "GetPaths must succeed with Progressive Readiness")
	require.NotNil(sa.t, result.Result, "Paths response must not be nil")

	sa.t.Logf("✅ PROGRESSIVE READINESS: GetPaths succeeded immediately")
	return result.Result
}

// AssertGetStreams gets streams and validates the response
func (sa *StreamAsserter) AssertGetStreams() *GetStreamsResponse {
	result := testutils.TestProgressiveReadiness(sa.t, func() (*GetStreamsResponse, error) {
		return sa.controller.GetStreams(sa.ctx)
	}, sa.controller, "GetStreams")

	require.NoError(sa.t, result.Error, "GetStreams must succeed with Progressive Readiness")
	require.NotNil(sa.t, result.Result, "Streams response must not be nil")

	sa.t.Logf("✅ PROGRESSIVE READINESS: GetStreams succeeded immediately")
	return result.Result
}

// AssertGetStream gets a specific stream and validates the response
func (sa *StreamAsserter) AssertGetStream(cameraID string) *GetStreamStatusResponse {
	result := testutils.TestProgressiveReadiness(sa.t, func() (*GetStreamStatusResponse, error) {
		return sa.controller.GetStreamStatus(sa.ctx, cameraID)
	}, sa.controller, "GetStreamStatus")

	require.NoError(sa.t, result.Error, "GetStreamStatus must succeed with Progressive Readiness")
	require.NotNil(sa.t, result.Result, "Stream response must not be nil")

	sa.t.Logf("✅ PROGRESSIVE READINESS: GetStreamStatus succeeded immediately")
	return result.Result
}

// AssertCreateStream creates a stream and validates the response
func (sa *StreamAsserter) AssertCreateStream(cameraID string) *StartStreamingResponse {
	result := testutils.TestProgressiveReadiness(sa.t, func() (*StartStreamingResponse, error) {
		return sa.controller.StartStreaming(sa.ctx, cameraID)
	}, sa.controller, "StartStreaming")

	require.NoError(sa.t, result.Error, "StartStreaming must succeed with Progressive Readiness")
	require.NotNil(sa.t, result.Result, "Start streaming response must not be nil")

	sa.t.Logf("✅ PROGRESSIVE READINESS: StartStreaming succeeded immediately")
	return result.Result
}

// AssertDeleteStream deletes a stream and validates the response
func (sa *StreamAsserter) AssertDeleteStream(cameraID string) error {
	err := sa.controller.DeleteStream(sa.ctx, cameraID)
	require.NoError(sa.t, err, "DeleteStream must succeed")

	sa.t.Logf("✅ DeleteStream succeeded")
	return err
}

// ============================================================================
// CONFIG ASSERTER - Configuration Management Operations
// ============================================================================

// ConfigAsserter encapsulates configuration test patterns
type ConfigAsserter struct {
	t          *testing.T
	helper     *MediaMTXTestHelper
	controller MediaMTXController
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewConfigAsserter creates a new config asserter with a ready controller
func NewConfigAsserter(t *testing.T) *ConfigAsserter {
	helper, _ := SetupMediaMTXTest(t)
	controller, ctx, cancel := helper.GetReadyController(t)
	return &ConfigAsserter{
		t:          t,
		helper:     helper,
		controller: controller,
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Cleanup stops the controller and cancels the context
func (ca *ConfigAsserter) Cleanup() {
	ca.controller.Stop(ca.ctx)
	ca.cancel()
}

// GetHelper returns the underlying MediaMTXTestHelper
func (ca *ConfigAsserter) GetHelper() *MediaMTXTestHelper {
	return ca.helper
}

// GetReadyController returns the ready controller
func (ca *ConfigAsserter) GetReadyController() MediaMTXController {
	return ca.controller
}

// GetContext returns the test context
func (ca *ConfigAsserter) GetContext() context.Context {
	return ca.ctx
}

// AssertGetRecordingConfig gets recording config and validates it
func (ca *ConfigAsserter) AssertGetRecordingConfig() *config.RecordingConfig {
	configIntegration := ca.helper.GetConfigIntegration()
	config, err := configIntegration.GetRecordingConfig()
	require.NoError(ca.t, err, "GetRecordingConfig should succeed")
	require.NotNil(ca.t, config, "Recording config should not be nil")
	
	ca.t.Logf("✅ Recording config retrieved successfully")
	return config
}

// AssertGetSnapshotConfig gets snapshot config and validates it
func (ca *ConfigAsserter) AssertGetSnapshotConfig() *config.SnapshotConfig {
	configIntegration := ca.helper.GetConfigIntegration()
	config, err := configIntegration.GetSnapshotConfig()
	require.NoError(ca.t, err, "GetSnapshotConfig should succeed")
	require.NotNil(ca.t, config, "Snapshot config should not be nil")
	
	ca.t.Logf("✅ Snapshot config retrieved successfully")
	return config
}

// AssertGetFFmpegConfig gets FFmpeg config and validates it
func (ca *ConfigAsserter) AssertGetFFmpegConfig() *config.FFmpegConfig {
	configIntegration := ca.helper.GetConfigIntegration()
	config, err := configIntegration.GetFFmpegConfig()
	require.NoError(ca.t, err, "GetFFmpegConfig should succeed")
	require.NotNil(ca.t, config, "FFmpeg config should not be nil")
	
	ca.t.Logf("✅ FFmpeg config retrieved successfully")
	return config
}

// AssertGetCameraConfig gets camera config and validates it
func (ca *ConfigAsserter) AssertGetCameraConfig() *config.CameraConfig {
	configIntegration := ca.helper.GetConfigIntegration()
	config, err := configIntegration.GetCameraConfig()
	require.NoError(ca.t, err, "GetCameraConfig should succeed")
	require.NotNil(ca.t, config, "Camera config should not be nil")
	
	ca.t.Logf("✅ Camera config retrieved successfully")
	return config
}

// AssertGetPerformanceConfig gets performance config and validates it
func (ca *ConfigAsserter) AssertGetPerformanceConfig() *config.PerformanceConfig {
	configIntegration := ca.helper.GetConfigIntegration()
	config, err := configIntegration.GetPerformanceConfig()
	require.NoError(ca.t, err, "GetPerformanceConfig should succeed")
	require.NotNil(ca.t, config, "Performance config should not be nil")
	
	ca.t.Logf("✅ Performance config retrieved successfully")
	return config
}

// AssertGetMainConfig gets main config and validates it
func (ca *ConfigAsserter) AssertGetMainConfig() *config.MediaMTXConfig {
	config, err := ca.controller.GetConfig(ca.ctx)
	require.NoError(ca.t, err, "GetConfig should succeed")
	require.NotNil(ca.t, config, "Main config should not be nil")
	
	ca.t.Logf("✅ Main config retrieved successfully")
	return config
}
