/*
MediaMTX FFmpeg Manager Implementation

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
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

// ffmpegManager represents the MediaMTX FFmpeg manager
type ffmpegManager struct {
	config *MediaMTXConfig
	logger *logrus.Logger

	// Process tracking
	processes map[int]*FFmpegProcess
	processMu sync.RWMutex

	// Performance tracking (Python parity)
	performanceMetrics map[string]*PerformanceMetrics
	metricsMu          sync.RWMutex

	// Retry and timeout management (Python parity)
	retryAttempts map[string]int
	retryMu       sync.RWMutex

	// Process cleanup tracking (Python parity)
	cleanupActions map[int][]string
	cleanupMu      sync.RWMutex
}

// PerformanceMetrics tracks performance metrics for operations
type PerformanceMetrics struct {
	OperationType    string
	TotalOperations  int64
	SuccessfulOps    int64
	FailedOps        int64
	AverageDuration  time.Duration
	LastOperation    time.Time
	ResponseTimeTargets map[string]float64
}

// FFmpegProcess represents an FFmpeg process
type FFmpegProcess struct {
	PID           int
	Command       []string
	OutputPath    string
	StartTime     time.Time
	Status        string
	cmd           *exec.Cmd
	SessionID     string
	Device        string
	FileSize      int64
	RotationCount int
	MaxFileSize   int64
	MaxDuration   time.Duration
	AutoRotate    bool
}

// NewFFmpegManager creates a new MediaMTX FFmpeg manager
func NewFFmpegManager(config *MediaMTXConfig, logger *logrus.Logger) FFmpegManager {
	// Set default FFmpeg configuration if not provided (Python parity)
	if config.FFmpeg.Snapshot.ProcessCreationTimeout == 0 {
		config.FFmpeg.Snapshot.ProcessCreationTimeout = 5 * time.Second
	}
	if config.FFmpeg.Snapshot.ExecutionTimeout == 0 {
		config.FFmpeg.Snapshot.ExecutionTimeout = 8 * time.Second
	}
	if config.FFmpeg.Snapshot.InternalTimeout == 0 {
		config.FFmpeg.Snapshot.InternalTimeout = 5000000
	}
	if config.FFmpeg.Snapshot.RetryAttempts == 0 {
		config.FFmpeg.Snapshot.RetryAttempts = 2
	}
	if config.FFmpeg.Snapshot.RetryDelay == 0 {
		config.FFmpeg.Snapshot.RetryDelay = 1 * time.Second
	}

	if config.FFmpeg.Recording.ProcessCreationTimeout == 0 {
		config.FFmpeg.Recording.ProcessCreationTimeout = 10 * time.Second
	}
	if config.FFmpeg.Recording.ExecutionTimeout == 0 {
		config.FFmpeg.Recording.ExecutionTimeout = 15 * time.Second
	}
	if config.FFmpeg.Recording.InternalTimeout == 0 {
		config.FFmpeg.Recording.InternalTimeout = 10000000
	}
	if config.FFmpeg.Recording.RetryAttempts == 0 {
		config.FFmpeg.Recording.RetryAttempts = 3
	}
	if config.FFmpeg.Recording.RetryDelay == 0 {
		config.FFmpeg.Recording.RetryDelay = 2 * time.Second
	}

	// Set default performance targets (Python parity)
	if config.Performance.ResponseTimeTargets == nil {
		config.Performance.ResponseTimeTargets = map[string]float64{
			"snapshot_capture": 2.0,
			"recording_start":  2.0,
			"recording_stop":   2.0,
			"file_listing":     1.0,
		}
	}

	if config.Performance.SnapshotTiers == nil {
		config.Performance.SnapshotTiers = map[string]float64{
			"tier1_rtsp_ready_check_timeout": 1.0,
			"tier2_activation_timeout":       3.0,
			"tier2_activation_trigger_timeout": 1.0,
			"tier3_direct_capture_timeout":   5.0,
			"total_operation_timeout":        10.0,
			"immediate_response_threshold":    0.5,
			"acceptable_response_threshold":  2.0,
			"slow_response_threshold":        5.0,
		}
	}

	if config.Performance.Optimization.EnableCaching == false {
		config.Performance.Optimization.EnableCaching = true
	}
	if config.Performance.Optimization.CacheTTL == 0 {
		config.Performance.Optimization.CacheTTL = 300 * time.Second
	}
	if config.Performance.Optimization.MaxConcurrentOperations == 0 {
		config.Performance.Optimization.MaxConcurrentOperations = 5
	}
	if config.Performance.Optimization.ConnectionPoolSize == 0 {
		config.Performance.Optimization.ConnectionPoolSize = 10
	}

	return &ffmpegManager{
		config:             config,
		logger:             logger,
		processes:          make(map[int]*FFmpegProcess),
		performanceMetrics: make(map[string]*PerformanceMetrics),
		retryAttempts:      make(map[string]int),
		cleanupActions:     make(map[int][]string),
	}
}

// StartProcess starts an FFmpeg process
func (fm *ffmpegManager) StartProcess(ctx context.Context, command []string, outputPath string) (int, error) {
	fm.logger.WithFields(logrus.Fields{
		"command":     command,
		"output_path": outputPath,
	}).Debug("Starting FFmpeg process")

	// Validate inputs (following Python implementation pattern)
	if len(command) == 0 {
		return 0, NewFFmpegError(0, "start_process", "start_process", "command cannot be empty")
	}

	if strings.TrimSpace(outputPath) == "" {
		return 0, NewFFmpegError(0, strings.Join(command, " "), "start_process", "output path cannot be empty")
	}

	// Create output directory if it doesn't exist
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return 0, NewFFmpegErrorWithErr(0, strings.Join(command, " "), "create_output_dir", "failed to create output directory", err)
	}

	// Create command
	cmd := exec.CommandContext(ctx, command[0], command[1:]...)

	// Set up process
	process := &FFmpegProcess{
		Command:    command,
		OutputPath: outputPath,
		StartTime:  time.Now(),
		Status:     "STARTING",
		cmd:        cmd,
	}

	// Start process
	if err := cmd.Start(); err != nil {
		return 0, NewFFmpegErrorWithErr(0, strings.Join(command, " "), "start_process", "failed to start FFmpeg process", err)
	}

	// Get PID
	process.PID = cmd.Process.Pid
	process.Status = "RUNNING"

	// Track process
	fm.processMu.Lock()
	fm.processes[process.PID] = process
	fm.processMu.Unlock()

	// Monitor process in background
	go fm.monitorProcess(process)

	fm.logger.WithFields(logrus.Fields{
		"pid":         process.PID,
		"command":     strings.Join(command, " "),
		"output_path": outputPath,
	}).Info("FFmpeg process started successfully")

	return process.PID, nil
}

// StopProcess stops an FFmpeg process with sophisticated cleanup (Python parity)
func (fm *ffmpegManager) StopProcess(ctx context.Context, pid int) error {
	fm.logger.WithField("pid", pid).Debug("Stopping FFmpeg process")

	fm.processMu.Lock()
	process, exists := fm.processes[pid]
	if !exists {
		fm.processMu.Unlock()
		return NewFFmpegError(pid, "stop_process", "stop_process", "process not found")
	}
	fm.processMu.Unlock()

	// Update status
	process.Status = "STOPPING"

	// Use sophisticated cleanup (Python parity)
	cleanupResult := fm.cleanupFFmpegProcess(process, pid, "stop_process")

	// Update status based on cleanup result
	if strings.Contains(cleanupResult, "graceful_exit") {
		process.Status = "STOPPED"
	} else if strings.Contains(cleanupResult, "force_exit") {
		process.Status = "FORCE_STOPPED"
	} else {
		process.Status = "CLEANUP_FAILED"
	}

	// Remove from tracking
	fm.processMu.Lock()
	delete(fm.processes, pid)
	fm.processMu.Unlock()

	fm.logger.WithFields(logrus.Fields{
		"pid":            pid,
		"cleanup_result": cleanupResult,
	}).Info("FFmpeg process stopped successfully")

	return nil
}

// cleanupFFmpegProcess implements sophisticated process cleanup (Python parity)
func (fm *ffmpegManager) cleanupFFmpegProcess(process *FFmpegProcess, pid int, operation string) string {
	correlationID := fmt.Sprintf("cleanup_%d_%s", pid, operation)
	cleanupActions := []string{}

	fm.logger.WithFields(logrus.Fields{
		"pid":            pid,
		"correlation_id": correlationID,
		"operation":      operation,
	}).Debug("Starting sophisticated FFmpeg process cleanup")

	// Check if process is already terminated
	if process.cmd.ProcessState != nil && process.cmd.ProcessState.Exited() {
		cleanupActions = append(cleanupActions, "already_exited")
		return strings.Join(cleanupActions, "_")
	}

	// Step 1: Graceful termination (SIGTERM)
	cleanupActions = append(cleanupActions, "terminate_attempt")
	if err := process.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		cleanupActions = append(cleanupActions, fmt.Sprintf("term_error_%s", err.Error()))
		fm.logger.WithError(err).WithFields(logrus.Fields{
			"pid":            pid,
			"correlation_id": correlationID,
		}).Warn("Failed to send SIGTERM to FFmpeg process")
	} else {
		cleanupActions = append(cleanupActions, "terminated")

		// Wait for graceful shutdown with configurable timeout
		terminationTimeout := time.Duration(fm.config.ProcessTerminationTimeout) * time.Second
		done := make(chan error, 1)
		go func() {
			done <- process.cmd.Wait()
		}()

		select {
		case err := <-done:
			if err != nil {
				fm.logger.WithError(err).WithFields(logrus.Fields{
					"pid":            pid,
					"correlation_id": correlationID,
				}).Warn("FFmpeg process exited with error during graceful shutdown")
			}
			cleanupActions = append(cleanupActions, "graceful_exit")
		case <-time.After(terminationTimeout):
			cleanupActions = append(cleanupActions, "term_timeout")
			fm.logger.WithFields(logrus.Fields{
				"pid":            pid,
				"correlation_id": correlationID,
				"timeout":        terminationTimeout,
			}).Warn("FFmpeg process did not terminate gracefully within timeout")
		}
	}

	// Step 2: Force kill (SIGKILL) if graceful shutdown failed
	if !strings.Contains(strings.Join(cleanupActions, "_"), "graceful_exit") {
		cleanupActions = append(cleanupActions, "kill_attempt")
		if err := process.cmd.Process.Kill(); err != nil {
			cleanupActions = append(cleanupActions, fmt.Sprintf("kill_error_%s", err.Error()))
			fm.logger.WithError(err).WithFields(logrus.Fields{
				"pid":            pid,
				"correlation_id": correlationID,
			}).Error("Failed to force kill FFmpeg process")
		} else {
			cleanupActions = append(cleanupActions, "killed")

			// Wait for force kill with configurable timeout
			killTimeout := time.Duration(fm.config.ProcessKillTimeout) * time.Second
			done := make(chan error, 1)
			go func() {
				done <- process.cmd.Wait()
			}()

			select {
			case err := <-done:
				if err != nil {
					fm.logger.WithError(err).WithFields(logrus.Fields{
						"pid":            pid,
						"correlation_id": correlationID,
					}).Warn("FFmpeg process exited with error during force kill")
				}
				cleanupActions = append(cleanupActions, "force_exit")
			case <-time.After(killTimeout):
				cleanupActions = append(cleanupActions, "kill_timeout")
				fm.logger.WithFields(logrus.Fields{
					"pid":            pid,
					"correlation_id": correlationID,
					"timeout":        killTimeout,
				}).Error("FFmpeg process did not respond to SIGKILL within timeout")
			}
		}
	}

	// Store cleanup actions for tracking
	fm.cleanupMu.Lock()
	fm.cleanupActions[pid] = cleanupActions
	fm.cleanupMu.Unlock()

	result := strings.Join(cleanupActions, "_")
	fm.logger.WithFields(logrus.Fields{
		"pid":            pid,
		"correlation_id": correlationID,
		"cleanup_result": result,
	}).Debug("FFmpeg process cleanup completed")

	return result
}

// IsProcessRunning checks if a process is running
func (fm *ffmpegManager) IsProcessRunning(ctx context.Context, pid int) bool {
	fm.processMu.RLock()
	process, exists := fm.processes[pid]
	fm.processMu.RUnlock()

	if !exists {
		return false
	}

	// Check if process is still running
	if process.cmd.Process == nil {
		return false
	}

	// Send signal 0 to check if process exists
	if err := process.cmd.Process.Signal(syscall.Signal(0)); err != nil {
		return false
	}

	return process.Status == "RUNNING"
}

// StartRecording starts an FFmpeg recording
func (fm *ffmpegManager) StartRecording(ctx context.Context, device, outputPath string, options map[string]string) (int, error) {
	fm.logger.WithFields(logrus.Fields{
		"device":      device,
		"output_path": outputPath,
		"options":     options,
	}).Debug("Starting FFmpeg recording")

	// Validate inputs (following Python implementation pattern)
	if strings.TrimSpace(device) == "" {
		return 0, NewFFmpegError(0, "start_recording", "start_recording", "device path cannot be empty")
	}

	if strings.TrimSpace(outputPath) == "" {
		return 0, NewFFmpegError(0, "start_recording", "start_recording", "output path cannot be empty")
	}

	// Build FFmpeg command
	command := fm.buildRecordingCommand(device, outputPath, options)

	// Start process
	pid, err := fm.StartProcess(ctx, command, outputPath)
	if err != nil {
		return 0, NewFFmpegErrorWithErr(0, strings.Join(command, " "), "start_recording", "failed to start recording", err)
	}

	fm.logger.WithFields(logrus.Fields{
		"pid":         pid,
		"device":      device,
		"output_path": outputPath,
	}).Info("FFmpeg recording started successfully")

	return pid, nil
}

// StopRecording stops an FFmpeg recording
func (fm *ffmpegManager) StopRecording(ctx context.Context, pid int) error {
	fm.logger.WithField("pid", pid).Debug("Stopping FFmpeg recording")

	return fm.StopProcess(ctx, pid)
}

// TakeSnapshot takes a snapshot using FFmpeg with retry logic (Python parity)
func (fm *ffmpegManager) TakeSnapshot(ctx context.Context, device, outputPath string) error {
	startTime := time.Now()
	operationType := "snapshot_capture"
	
	fm.logger.WithFields(logrus.Fields{
		"device":         device,
		"output_path":    outputPath,
		"operation_type": operationType,
	}).Debug("Taking FFmpeg snapshot with retry logic")

	// Track performance metrics
	defer func() {
		duration := time.Since(startTime)
		fm.recordPerformanceMetrics(operationType, duration, nil)
	}()

	// Build FFmpeg command for snapshot
	command := fm.buildSnapshotCommand(device, outputPath)

	// Create output directory if it doesn't exist
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return NewFFmpegErrorWithErr(0, strings.Join(command, " "), "create_output_dir", "failed to create output directory", err)
	}

	// Execute with retry logic (Python parity)
	return fm.executeWithRetry(ctx, command, fm.config.FFmpeg.Snapshot, "take_snapshot", "failed to take snapshot")
}

// executeWithRetry executes FFmpeg command with retry logic and exponential backoff (Python parity)
func (fm *ffmpegManager) executeWithRetry(ctx context.Context, command []string, config interface{}, operation, errorMsg string) error {
	var retryAttempts int
	var retryDelay time.Duration
	var processCreationTimeout time.Duration
	var executionTimeout time.Duration

	// Extract configuration based on type
	switch cfg := config.(type) {
	case SnapshotConfig:
		retryAttempts = cfg.RetryAttempts
		retryDelay = cfg.RetryDelay
		processCreationTimeout = cfg.ProcessCreationTimeout
		executionTimeout = cfg.ExecutionTimeout
	case RecordingConfig:
		retryAttempts = cfg.RetryAttempts
		retryDelay = cfg.RetryDelay
		processCreationTimeout = cfg.ProcessCreationTimeout
		executionTimeout = cfg.ExecutionTimeout
	default:
		// Default values if config is not provided
		retryAttempts = 2
		retryDelay = 1 * time.Second
		processCreationTimeout = 5 * time.Second
		executionTimeout = 8 * time.Second
	}

	correlationID := fmt.Sprintf("retry_%s_%d", operation, time.Now().Unix())
	var lastErr error

	for attempt := 0; attempt <= retryAttempts; attempt++ {
		fm.logger.WithFields(logrus.Fields{
			"attempt":        attempt,
			"max_attempts":   retryAttempts + 1,
			"correlation_id": correlationID,
			"operation":      operation,
		}).Debug("Executing FFmpeg command with retry")

		// Create context with timeout for process creation
		processCtx, cancel := context.WithTimeout(ctx, processCreationTimeout)
		
		// Create command
		cmd := exec.CommandContext(processCtx, command[0], command[1:]...)
		
		// Start process with timeout
		if err := cmd.Start(); err != nil {
			cancel()
			lastErr = NewFFmpegErrorWithErr(0, strings.Join(command, " "), operation, fmt.Sprintf("%s (attempt %d)", errorMsg, attempt+1), err)
			fm.logger.WithError(err).WithFields(logrus.Fields{
				"attempt":        attempt,
				"correlation_id": correlationID,
			}).Warn("Failed to start FFmpeg process")
			
			if attempt < retryAttempts {
				backoffDelay := fm.calculateBackoffDelay(retryDelay, attempt)
				fm.logger.WithFields(logrus.Fields{
					"attempt":        attempt,
					"backoff_delay":  backoffDelay,
					"correlation_id": correlationID,
				}).Debug("Retrying FFmpeg operation after backoff")
				time.Sleep(backoffDelay)
				continue
			}
			return lastErr
		}
		
		cancel() // Cancel process creation timeout

		// Execute with timeout
		execCtx, cancel := context.WithTimeout(ctx, executionTimeout)
		done := make(chan error, 1)
		go func() {
			done <- cmd.Wait()
		}()

		select {
		case err := <-done:
			cancel()
			if err == nil {
				fm.logger.WithFields(logrus.Fields{
					"attempt":        attempt,
					"correlation_id": correlationID,
				}).Info("FFmpeg operation completed successfully")
				return nil
			}
			lastErr = NewFFmpegErrorWithErr(0, strings.Join(command, " "), operation, fmt.Sprintf("%s (attempt %d)", errorMsg, attempt+1), err)
			fm.logger.WithError(err).WithFields(logrus.Fields{
				"attempt":        attempt,
				"correlation_id": correlationID,
			}).Warn("FFmpeg operation failed")
			
		case <-execCtx.Done():
			cancel()
			// Cleanup the process
			fm.cleanupFFmpegProcess(&FFmpegProcess{cmd: cmd}, 0, operation)
			lastErr = NewFFmpegError(0, operation, operation, fmt.Sprintf("execution timeout after %v", executionTimeout))
			fm.logger.WithFields(logrus.Fields{
				"attempt":        attempt,
				"timeout":        executionTimeout,
				"correlation_id": correlationID,
			}).Warn("FFmpeg operation timed out")
		}

		if attempt < retryAttempts {
			backoffDelay := fm.calculateBackoffDelay(retryDelay, attempt)
			fm.logger.WithFields(logrus.Fields{
				"attempt":       attempt,
				"backoff_delay": backoffDelay,
				"correlation_id": correlationID,
			}).Debug("Retrying FFmpeg operation after backoff")
			time.Sleep(backoffDelay)
		}
	}

	return lastErr
}

// calculateBackoffDelay calculates exponential backoff delay (Python parity)
func (fm *ffmpegManager) calculateBackoffDelay(baseDelay time.Duration, attempt int) time.Duration {
	// Exponential backoff: baseDelay * 2^attempt
	backoffMultiplier := 1 << uint(attempt)
	delay := time.Duration(float64(baseDelay) * float64(backoffMultiplier))
	
	// Add jitter to prevent thundering herd
	jitter := time.Duration(rand.Int63n(int64(delay) / 4))
	delay += jitter
	
	// Cap maximum delay at 30 seconds
	maxDelay := 30 * time.Second
	if delay > maxDelay {
		delay = maxDelay
	}
	
	return delay
}

// recordPerformanceMetrics records performance metrics for operations (Python parity)
func (fm *ffmpegManager) recordPerformanceMetrics(operationType string, duration time.Duration, err error) {
	fm.metricsMu.Lock()
	defer fm.metricsMu.Unlock()

	metrics, exists := fm.performanceMetrics[operationType]
	if !exists {
		metrics = &PerformanceMetrics{
			OperationType:       operationType,
			ResponseTimeTargets: fm.config.Performance.ResponseTimeTargets,
		}
		fm.performanceMetrics[operationType] = metrics
	}

	metrics.TotalOperations++
	metrics.LastOperation = time.Now()

	if err == nil {
		metrics.SuccessfulOps++
	} else {
		metrics.FailedOps++
	}

	// Update average duration
	if metrics.TotalOperations == 1 {
		metrics.AverageDuration = duration
	} else {
		// Calculate running average
		totalDuration := metrics.AverageDuration * time.Duration(metrics.TotalOperations-1)
		totalDuration += duration
		metrics.AverageDuration = totalDuration / time.Duration(metrics.TotalOperations)
	}

	fm.logger.WithFields(logrus.Fields{
		"operation_type":   operationType,
		"duration":         duration,
		"average_duration": metrics.AverageDuration,
		"total_ops":        metrics.TotalOperations,
		"successful_ops":   metrics.SuccessfulOps,
		"failed_ops":       metrics.FailedOps,
		"success_rate":     float64(metrics.SuccessfulOps) / float64(metrics.TotalOperations) * 100,
	}).Debug("Performance metrics recorded")
}

// RotateFile rotates a file
func (fm *ffmpegManager) RotateFile(ctx context.Context, oldPath, newPath string) error {
	fm.logger.WithFields(logrus.Fields{
		"old_path": oldPath,
		"new_path": newPath,
	}).Debug("Rotating FFmpeg file")

	// Create new directory if it doesn't exist
	newDir := filepath.Dir(newPath)
	if err := os.MkdirAll(newDir, 0755); err != nil {
		return NewFFmpegErrorWithErr(0, "rotate_file", "create_new_dir", "failed to create new directory", err)
	}

	// Rename file
	if err := os.Rename(oldPath, newPath); err != nil {
		return NewFFmpegErrorWithErr(0, "rotate_file", "rename_file", "failed to rename file", err)
	}

	fm.logger.WithFields(logrus.Fields{
		"old_path": oldPath,
		"new_path": newPath,
	}).Info("FFmpeg file rotated successfully")

	return nil
}

// GetFileInfo gets file information
func (fm *ffmpegManager) GetFileInfo(ctx context.Context, path string) (int64, time.Time, error) {
	fm.logger.WithField("path", path).Debug("Getting FFmpeg file info")

	// Get file info
	info, err := os.Stat(path)
	if err != nil {
		return 0, time.Time{}, NewFFmpegErrorWithErr(0, "get_file_info", "stat_file", "failed to get file info", err)
	}

	return info.Size(), info.ModTime(), nil
}

// buildRecordingCommand builds an FFmpeg command for recording
func (fm *ffmpegManager) buildRecordingCommand(device, outputPath string, options map[string]string) []string {
	command := []string{"ffmpeg"}

	// Input device
	command = append(command, "-f", "v4l2")
	command = append(command, "-i", device)

	// Video codec
	codec := options["codec"]
	if codec == "" {
		codec = "libx264"
	}
	command = append(command, "-c:v", codec)

	// Preset
	preset := options["preset"]
	if preset == "" {
		preset = "fast"
	}
	command = append(command, "-preset", preset)

	// CRF (quality)
	crf := options["crf"]
	if crf == "" {
		crf = "23"
	}
	command = append(command, "-crf", crf)

	// Format
	format := options["format"]
	if format == "" {
		format = "mp4"
	}
	command = append(command, "-f", format)

	// Output path
	command = append(command, outputPath)

	return command
}

// CreateSegmentedRecording creates a segmented recording with continuity support (Phase 3 enhancement)
func (fm *ffmpegManager) CreateSegmentedRecording(ctx context.Context, input, output string, settings *RotationSettings) error {
	fm.logger.WithFields(logrus.Fields{
		"input":    input,
		"output":   output,
		"settings": settings,
	}).Info("Creating segmented recording with continuity")

	// Build segment-based FFmpeg command
	args := fm.buildSegmentedRecordingCommand(input, output, settings)

	// Execute FFmpeg command
	return fm.executeFFmpeg(args)
}

// buildSegmentedRecordingCommand builds FFmpeg command for segmented recording (Phase 3 enhancement)
func (fm *ffmpegManager) buildSegmentedRecordingCommand(input, output string, settings *RotationSettings) []string {
	command := []string{"ffmpeg"}

	// Input
	command = append(command, "-i", input)

	// Segment-based output with continuity
	command = append(command, "-f", "segment")

	// Segment duration
	if settings.SegmentDuration > 0 {
		command = append(command, "-segment_time", settings.SegmentDuration.String())
	}

	// Reset timestamps for each segment
	if settings.ResetTimestamps {
		command = append(command, "-reset_timestamps", "1")
	}

	// Segment format with continuity support
	segmentFormat := fm.buildSegmentFormat(output, settings)
	command = append(command, segmentFormat)

	fm.logger.WithFields(logrus.Fields{
		"command":        command,
		"segment_format": segmentFormat,
	}).Debug("Built segmented recording command")

	return command
}

// buildSegmentFormat builds segment filename format with continuity support (Phase 3 enhancement)
func (fm *ffmpegManager) buildSegmentFormat(baseOutput string, settings *RotationSettings) string {
	// Base directory and filename
	baseDir := filepath.Dir(baseOutput)
	baseName := filepath.Base(baseOutput)
	nameWithoutExt := strings.TrimSuffix(baseName, filepath.Ext(baseName))

	// Build segment format
	var format string

	if settings.ContinuityMode {
		// Continuity mode: include continuity ID and segment index
		format = filepath.Join(baseDir, fmt.Sprintf("%s_%s_%%03d%s",
			nameWithoutExt,
			settings.ContinuityID,
			filepath.Ext(baseName)))
	} else {
		// Standard mode: just segment index
		format = filepath.Join(baseDir, fmt.Sprintf("%s_%%03d%s",
			nameWithoutExt,
			filepath.Ext(baseName)))
	}

	// Add strftime support if enabled
	if settings.StrftimeEnabled {
		format = strings.ReplaceAll(format, "%03d", "%Y%m%d_%H%M%S_%03d")
	}

	return format
}

// executeFFmpeg executes FFmpeg command with error handling (Phase 3 enhancement)
func (fm *ffmpegManager) executeFFmpeg(args []string) error {
	fm.logger.WithField("args", args).Debug("Executing FFmpeg command")

	// Create command
	cmd := exec.Command("ffmpeg", args...)

	// Capture output
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute command
	if err := cmd.Run(); err != nil {
		fm.logger.WithFields(logrus.Fields{
			"error":  err,
			"stdout": stdout.String(),
			"stderr": stderr.String(),
		}).Error("FFmpeg command failed")

		return NewFFmpegErrorWithErr(0, strings.Join(args, " "), "execute_ffmpeg", "FFmpeg command failed", err)
	}

	fm.logger.WithFields(logrus.Fields{
		"stdout": stdout.String(),
		"stderr": stderr.String(),
	}).Debug("FFmpeg command executed successfully")

	return nil
}

// buildSnapshotCommand builds an FFmpeg command for snapshot
func (fm *ffmpegManager) buildSnapshotCommand(device, outputPath string) []string {
	command := []string{"ffmpeg"}

	// Input device
	command = append(command, "-f", "v4l2")
	command = append(command, "-i", device)

	// Video frames
	command = append(command, "-vframes", "1")

	// Output path
	command = append(command, outputPath)

	return command
}

// monitorProcess monitors an FFmpeg process
func (fm *ffmpegManager) monitorProcess(process *FFmpegProcess) {
	// Wait for process to complete
	err := process.cmd.Wait()

	// Update status
	process.Status = "COMPLETED"
	if err != nil {
		process.Status = "FAILED"
		fm.logger.WithError(err).WithField("pid", process.PID).Error("FFmpeg process failed")
	} else {
		fm.logger.WithField("pid", process.PID).Info("FFmpeg process completed successfully")
	}

	// Remove from tracking
	fm.processMu.Lock()
	delete(fm.processes, process.PID)
	fm.processMu.Unlock()
}

// BuildCommand builds an FFmpeg command with the provided arguments
func (fm *ffmpegManager) BuildCommand(args ...string) []string {
	fm.logger.WithField("args", args).Debug("Building FFmpeg command")

	// Start with ffmpeg command
	command := []string{"ffmpeg"}

	// Add all provided arguments
	command = append(command, args...)

	fm.logger.WithField("command", command).Debug("FFmpeg command built successfully")
	return command
}
