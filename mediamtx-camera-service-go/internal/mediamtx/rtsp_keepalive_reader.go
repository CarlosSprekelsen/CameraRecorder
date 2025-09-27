// rtsp_keepalive_reader.go
package mediamtx

import (
	"context"
	"fmt"
	"os/exec"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
)

// RTSPKeepaliveReader manages keepalive RTSP connections to trigger MediaMTX runOnDemand publishers.
//
// PURPOSE: Solves MediaMTX on-demand recording startup problem
// PROBLEM: MediaMTX waits for RTSP client connection before starting FFmpeg publisher
// SOLUTION: Create dummy RTSP connection to immediately trigger recording
//
// USER EXPERIENCE IMPACT:
// - WITHOUT keepalive: Recording starts only when first viewer connects (delayed/missed recording)
// - WITH keepalive: Recording starts immediately when user requests it (expected behavior)
//
// POWER USAGE ANALYSIS (48-hour battery target):
// - CPU impact: ~2-5% per active recording (acceptable for edge computing)
// - Scope: ONLY during active recording sessions (not continuous background)
// - Auto-cleanup: Stops when recording stops (no resource leak)
// - Alternative investigation: MediaMTX native auto-start options explored but not available
//
// ARCHITECTURE DECISION: Keep current implementation for optimal UX
// Battery impact is acceptable given limited scope (recording-only) and user expectation
// that recording should start immediately when requested.
type RTSPKeepaliveReader struct {
	config        *config.MediaMTXConfig
	logger        *logging.Logger
	activeReaders sync.Map // map[pathName]*keepaliveSession

	// Resource management
	running         int32 // Atomic flag for running state
	maxRestartCount int   // Maximum restart attempts per session
	resourceStats   *KeepaliveResourceStats
}

type keepaliveSession struct {
	pathName string
	rtspURL  string
	cmd      *exec.Cmd
	cancel   context.CancelFunc
	done     chan struct{}

	// Resource management
	restartCount int32 // Atomic counter for restart attempts
	startTime    time.Time
}

// KeepaliveResourceStats tracks resource usage for the keepalive reader
type KeepaliveResourceStats struct {
	ActiveSessions       int64 `json:"active_sessions"`
	TotalSessionsStarted int64 `json:"total_sessions_started"`
	TotalSessionsStopped int64 `json:"total_sessions_stopped"`
	ProcessRestarts      int64 `json:"process_restarts"`
	ProcessFailures      int64 `json:"process_failures"`
}

// NewRTSPKeepaliveReader creates a new keepalive reader manager
func NewRTSPKeepaliveReader(config *config.MediaMTXConfig, logger *logging.Logger) *RTSPKeepaliveReader {
	return &RTSPKeepaliveReader{
		config:          config,
		logger:          logger,
		running:         0, // Initially not running
		maxRestartCount: 3, // Default maximum restart attempts
		resourceStats:   &KeepaliveResourceStats{},
	}
}

// NewRTSPKeepaliveReaderWithConfig creates a new keepalive reader manager with recording config
func NewRTSPKeepaliveReaderWithConfig(config *config.MediaMTXConfig, recordingConfig *config.RecordingConfig, logger *logging.Logger) *RTSPKeepaliveReader {
	maxRestartCount := 3 // Default
	if recordingConfig != nil && recordingConfig.MaxRestartCount > 0 {
		maxRestartCount = recordingConfig.MaxRestartCount
	}

	return &RTSPKeepaliveReader{
		config:          config,
		logger:          logger,
		running:         0, // Initially not running
		maxRestartCount: maxRestartCount,
		resourceStats:   &KeepaliveResourceStats{},
	}
}

// StartKeepalive starts a keepalive reader for the given path
// This triggers runOnDemand and keeps the FFmpeg publisher alive
func (kr *RTSPKeepaliveReader) StartKeepalive(ctx context.Context, pathName string) error {
	// Check if already exists
	if _, exists := kr.activeReaders.Load(pathName); exists {
		kr.logger.WithField("path", pathName).Debug("Keepalive reader already exists")
		return nil
	}

	// Build RTSP URL
	rtspURL := fmt.Sprintf("rtsp://%s:%d/%s",
		kr.config.Host,
		kr.config.RTSPPort,
		pathName)

	// Create cancellable context
	ctx, cancel := context.WithCancel(ctx)

	session := &keepaliveSession{
		pathName:     pathName,
		rtspURL:      rtspURL,
		cancel:       cancel,
		done:         make(chan struct{}),
		restartCount: 0,
		startTime:    time.Now(),
	}

	// Start the keepalive reader
	if err := kr.startReader(ctx, session); err != nil {
		cancel()
		return fmt.Errorf("failed to start keepalive reader: %w", err)
	}

	// Store the session
	kr.activeReaders.Store(pathName, session)

	// Update statistics
	atomic.AddInt64(&kr.resourceStats.ActiveSessions, 1)
	atomic.AddInt64(&kr.resourceStats.TotalSessionsStarted, 1)

	kr.logger.WithFields(logging.Fields{
		"path":     pathName,
		"rtsp_url": rtspURL,
	}).Info("Keepalive reader started for recording")

	return nil
}

// StopKeepalive stops the keepalive reader for the given path (non-blocking)
func (kr *RTSPKeepaliveReader) StopKeepalive(pathName string) error {
	sessionI, exists := kr.activeReaders.LoadAndDelete(pathName)
	if !exists {
		kr.logger.WithField("path", pathName).Debug("No keepalive reader to stop")
		return nil
	}

	session := sessionI.(*keepaliveSession)

	// Cancel the context to signal stop
	session.cancel()

	// Start async cleanup goroutine to avoid blocking the caller
	go func() {
		// Wait for graceful shutdown with timeout
		select {
		case <-session.done:
			kr.logger.WithField("path", pathName).Info("Keepalive reader stopped gracefully")
		case <-time.After(time.Duration(kr.config.ProcessTerminationTimeout * float64(time.Second))):
			// Force kill process group if not stopped gracefully
			if session.cmd != nil && session.cmd.Process != nil {
				// Kill the entire process group to prevent orphaned processes
				syscall.Kill(-session.cmd.Process.Pid, syscall.SIGKILL)
			}
			kr.logger.WithField("path", pathName).Warn("Keepalive reader force stopped with process group kill")
		}

		// Update statistics
		atomic.AddInt64(&kr.resourceStats.ActiveSessions, -1)
		atomic.AddInt64(&kr.resourceStats.TotalSessionsStopped, 1)
	}()

	return nil
}

// StopKeepaliveSync stops the keepalive reader synchronously (for shutdown scenarios)
func (kr *RTSPKeepaliveReader) StopKeepaliveSync(pathName string) error {
	sessionI, exists := kr.activeReaders.LoadAndDelete(pathName)
	if !exists {
		kr.logger.WithField("path", pathName).Debug("No keepalive reader to stop")
		return nil
	}

	session := sessionI.(*keepaliveSession)

	// Cancel the context to stop the reader
	session.cancel()

	// Wait for graceful shutdown with timeout
	select {
	case <-session.done:
		kr.logger.WithField("path", pathName).Info("Keepalive reader stopped gracefully")
	case <-time.After(time.Duration(kr.config.ProcessTerminationTimeout * float64(time.Second))):
		// Force kill process group if not stopped gracefully
		if session.cmd != nil && session.cmd.Process != nil {
			syscall.Kill(-session.cmd.Process.Pid, syscall.SIGKILL)
		}
		kr.logger.WithField("path", pathName).Warn("Keepalive reader force stopped with process group kill")
	}

	// Update statistics
	atomic.AddInt64(&kr.resourceStats.ActiveSessions, -1)
	atomic.AddInt64(&kr.resourceStats.TotalSessionsStopped, 1)

	return nil
}

// startReader starts the actual RTSP reader process
func (kr *RTSPKeepaliveReader) startReader(ctx context.Context, session *keepaliveSession) error {
	// Use FFmpeg as a null sink reader - minimal resource usage
	// This connects to the RTSP stream and discards the data
	cmd := exec.CommandContext(ctx,
		"ffmpeg",
		"-rtsp_transport", "tcp", // Use TCP for reliability
		"-i", session.rtspURL, // Input RTSP URL
		"-f", "null", // Null output format
		"-", // Output to stdout (discarded)
	)

	// Set process group for proper cleanup - prevents orphaned processes
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true, // Create new process group
		Pgid:    0,    // Use process PID as group ID
	}

	// Suppress FFmpeg output unless debugging
	// Note: We'll suppress output by default for cleaner logs
	cmd.Stdout = nil
	cmd.Stderr = nil

	session.cmd = cmd

	// Start the reader process
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start FFmpeg reader: %w", err)
	}

	// Monitor the reader in a goroutine
	go kr.monitorReader(ctx, session)

	// Give it a moment to connect using context-aware timeout
	select {
	case <-time.After(time.Duration(kr.config.StreamReadiness.CheckInterval * float64(time.Second))):
		// Connection should be established now
	case <-ctx.Done():
		// Context cancelled, return early
		return fmt.Errorf("context cancelled while waiting for RTSP connection to establish")
	}

	return nil
}

// monitorReader monitors the reader process and restarts if needed
func (kr *RTSPKeepaliveReader) monitorReader(ctx context.Context, session *keepaliveSession) {
	defer func() {
		// Only close the channel if it hasn't been closed already
		select {
		case <-session.done:
			// Channel already closed, do nothing
		default:
			close(session.done)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			// Context cancelled, stop monitoring
			if session.cmd != nil && session.cmd.Process != nil {
				// Kill the entire process group
				syscall.Kill(-session.cmd.Process.Pid, syscall.SIGTERM)
			}
			return
		default:
			// Wait for the process to exit
			err := session.cmd.Wait()

			// Check if we should restart
			select {
			case <-ctx.Done():
				// Context cancelled during wait, don't restart
				return
			default:
				if err != nil {
					restartCount := atomic.AddInt32(&session.restartCount, 1)

					// Check restart limit
					if int(restartCount) > kr.maxRestartCount {
						atomic.AddInt64(&kr.resourceStats.ProcessFailures, 1)
						kr.logger.WithFields(logging.Fields{
							"path":          session.pathName,
							"restart_count": restartCount,
							"max_restarts":  kr.maxRestartCount,
							"error":         err.Error(),
						}).Error("Keepalive reader exceeded maximum restart attempts, stopping")
						return
					}

					atomic.AddInt64(&kr.resourceStats.ProcessRestarts, 1)
					kr.logger.WithFields(logging.Fields{
						"path":          session.pathName,
						"restart_count": restartCount,
						"error":         err.Error(),
					}).Warn("Keepalive reader exited, restarting...")

					// Exponential backoff based on restart count using configuration
					baseDelay := time.Duration(kr.config.StreamReadiness.RetryDelay * float64(time.Second))
					maxDelay := time.Duration(kr.config.HealthMonitorDefaults.MaxBackoffDelay * float64(time.Second))
					backoffDuration := time.Duration(restartCount) * baseDelay
					if backoffDuration > maxDelay {
						backoffDuration = maxDelay // Cap at configured maximum
					}

					// Wait before restart using context-aware timeout
					select {
					case <-time.After(backoffDuration):
						// Backoff period completed, proceed with restart
					case <-ctx.Done():
						// Context cancelled, exit early
						return
					}

					// Restart the reader
					if err := kr.startReader(ctx, session); err != nil {
						atomic.AddInt64(&kr.resourceStats.ProcessFailures, 1)
						kr.logger.WithError(err).WithFields(logging.Fields{
							"path":          session.pathName,
							"restart_count": restartCount,
						}).Error("Failed to restart keepalive reader")
						return
					}
				}
			}
		}
	}
}

// StopAll stops all active keepalive readers
func (kr *RTSPKeepaliveReader) StopAll() {
	kr.activeReaders.Range(func(key, value interface{}) bool {
		pathName := key.(string)
		kr.StopKeepaliveSync(pathName) // Use sync version for complete cleanup
		return true
	})
}

// GetActiveCount returns the number of active keepalive readers
func (kr *RTSPKeepaliveReader) GetActiveCount() int {
	count := 0
	kr.activeReaders.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	return count
}

// IsActive checks if a keepalive reader is active for the path
func (kr *RTSPKeepaliveReader) IsActive(pathName string) bool {
	_, exists := kr.activeReaders.Load(pathName)
	return exists
}

// Resource Management Methods - Implementation of camera.ResourceManager and camera.CleanupManager interfaces

// Start initializes the keepalive reader manager (implements camera.ResourceManager)
func (kr *RTSPKeepaliveReader) Start(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&kr.running, 0, 1) {
		return fmt.Errorf("keepalive reader manager is already running")
	}

	kr.logger.Info("RTSP keepalive reader manager started")
	return nil
}

// Stop gracefully shuts down the keepalive reader manager (implements camera.ResourceManager)
func (kr *RTSPKeepaliveReader) Stop(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&kr.running, 1, 0) {
		kr.logger.Debug("Keepalive reader manager is already stopped")
		return nil // Idempotent
	}

	kr.logger.Info("Stopping RTSP keepalive reader manager...")

	// Stop all active readers
	kr.StopAll()

	kr.logger.Info("RTSP keepalive reader manager stopped successfully")
	return nil
}

// IsRunning returns whether the keepalive reader manager is running (implements camera.ResourceManager)
func (kr *RTSPKeepaliveReader) IsRunning() bool {
	return atomic.LoadInt32(&kr.running) == 1
}

// Cleanup performs resource cleanup (implements camera.CleanupManager)
func (kr *RTSPKeepaliveReader) Cleanup(ctx context.Context) error {
	kr.logger.Info("Performing keepalive reader cleanup...")

	// Force stop all active sessions
	var sessionPaths []string
	kr.activeReaders.Range(func(key, value interface{}) bool {
		pathName := key.(string)
		sessionPaths = append(sessionPaths, pathName)
		return true
	})

	for _, pathName := range sessionPaths {
		if sessionI, exists := kr.activeReaders.LoadAndDelete(pathName); exists {
			session := sessionI.(*keepaliveSession)

			// Cancel context
			session.cancel()

			// Force kill process group immediately during cleanup
			if session.cmd != nil && session.cmd.Process != nil {
				syscall.Kill(-session.cmd.Process.Pid, syscall.SIGKILL)
			}

			kr.logger.WithField("path", pathName).Debug("Force stopped keepalive reader during cleanup")
		}
	}

	kr.logger.WithFields(logging.Fields{
		"stopped_sessions": len(sessionPaths),
	}).Info("Keepalive reader cleanup completed")
	return nil
}

// GetResourceStats returns current resource usage statistics (implements camera.CleanupManager)
func (kr *RTSPKeepaliveReader) GetResourceStats() map[string]interface{} {
	// Update active sessions count
	activeCount := int64(0)
	kr.activeReaders.Range(func(_, _ interface{}) bool {
		activeCount++
		return true
	})
	atomic.StoreInt64(&kr.resourceStats.ActiveSessions, activeCount)

	return map[string]interface{}{
		"running":                kr.IsRunning(),
		"active_sessions":        atomic.LoadInt64(&kr.resourceStats.ActiveSessions),
		"total_sessions_started": atomic.LoadInt64(&kr.resourceStats.TotalSessionsStarted),
		"total_sessions_stopped": atomic.LoadInt64(&kr.resourceStats.TotalSessionsStopped),
		"process_restarts":       atomic.LoadInt64(&kr.resourceStats.ProcessRestarts),
		"process_failures":       atomic.LoadInt64(&kr.resourceStats.ProcessFailures),
		"max_restart_count":      kr.maxRestartCount,
	}
}
