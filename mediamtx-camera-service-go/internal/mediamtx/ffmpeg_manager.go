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
	return &ffmpegManager{
		config:    config,
		logger:    logger,
		processes: make(map[int]*FFmpegProcess),
	}
}

// StartProcess starts an FFmpeg process
func (fm *ffmpegManager) StartProcess(ctx context.Context, command []string, outputPath string) (int, error) {
	fm.logger.WithFields(logrus.Fields{
		"command":     command,
		"output_path": outputPath,
	}).Debug("Starting FFmpeg process")

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

// StopProcess stops an FFmpeg process
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

	// Send SIGTERM first
	if err := process.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		fm.logger.WithError(err).WithField("pid", pid).Warn("Failed to send SIGTERM to FFmpeg process")
	}

	// Wait for graceful shutdown
	done := make(chan error, 1)
	go func() {
		done <- process.cmd.Wait()
	}()

	select {
	case err := <-done:
		if err != nil {
			fm.logger.WithError(err).WithField("pid", pid).Warn("FFmpeg process exited with error")
		}
	case <-time.After(5 * time.Second):
		// Force kill if graceful shutdown fails
		fm.logger.WithField("pid", pid).Warn("FFmpeg process did not stop gracefully, force killing")
		if err := process.cmd.Process.Kill(); err != nil {
			fm.logger.WithError(err).WithField("pid", pid).Error("Failed to force kill FFmpeg process")
		}
		<-done // Wait for process to exit
	}

	// Update status
	process.Status = "STOPPED"

	// Remove from tracking
	fm.processMu.Lock()
	delete(fm.processes, pid)
	fm.processMu.Unlock()

	fm.logger.WithField("pid", pid).Info("FFmpeg process stopped successfully")
	return nil
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

// TakeSnapshot takes a snapshot using FFmpeg
func (fm *ffmpegManager) TakeSnapshot(ctx context.Context, device, outputPath string) error {
	fm.logger.WithFields(logrus.Fields{
		"device":      device,
		"output_path": outputPath,
	}).Debug("Taking FFmpeg snapshot")

	// Build FFmpeg command for snapshot
	command := fm.buildSnapshotCommand(device, outputPath)

	// Create output directory if it doesn't exist
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return NewFFmpegErrorWithErr(0, strings.Join(command, " "), "create_output_dir", "failed to create output directory", err)
	}

	// Create command
	cmd := exec.CommandContext(ctx, command[0], command[1:]...)

	// Execute command
	if err := cmd.Run(); err != nil {
		return NewFFmpegErrorWithErr(0, strings.Join(command, " "), "take_snapshot", "failed to take snapshot", err)
	}

	fm.logger.WithFields(logrus.Fields{
		"device":      device,
		"output_path": outputPath,
	}).Info("FFmpeg snapshot taken successfully")

	return nil
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
		"input":   input,
		"output":  output,
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
		"command": command,
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
