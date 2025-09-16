// rtsp_keepalive_reader.go
package mediamtx

import (
	"context"
	"fmt"
	"os/exec"
	"sync"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
)

// RTSPKeepaliveReader manages keepalive RTSP connections to trigger runOnDemand
type RTSPKeepaliveReader struct {
	config        *config.MediaMTXConfig
	logger        *logging.Logger
	activeReaders sync.Map // map[pathName]*keepaliveSession
}

type keepaliveSession struct {
	pathName string
	rtspURL  string
	cmd      *exec.Cmd
	cancel   context.CancelFunc
	done     chan struct{}
}

// NewRTSPKeepaliveReader creates a new keepalive reader manager
func NewRTSPKeepaliveReader(config *config.MediaMTXConfig, logger *logging.Logger) *RTSPKeepaliveReader {
	return &RTSPKeepaliveReader{
		config: config,
		logger: logger,
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
		pathName: pathName,
		rtspURL:  rtspURL,
		cancel:   cancel,
		done:     make(chan struct{}),
	}

	// Start the keepalive reader
	if err := kr.startReader(ctx, session); err != nil {
		cancel()
		return fmt.Errorf("failed to start keepalive reader: %w", err)
	}

	// Store the session
	kr.activeReaders.Store(pathName, session)

	kr.logger.WithFields(logging.Fields{
		"path":     pathName,
		"rtsp_url": rtspURL,
	}).Info("Keepalive reader started for recording")

	return nil
}

// StopKeepalive stops the keepalive reader for the given path
func (kr *RTSPKeepaliveReader) StopKeepalive(pathName string) error {
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
	case <-time.After(5 * time.Second):
		// Force kill if not stopped gracefully
		if session.cmd != nil && session.cmd.Process != nil {
			session.cmd.Process.Kill()
		}
		kr.logger.WithField("path", pathName).Warn("Keepalive reader force stopped")
	}

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
	case <-time.After(TestTimeoutLong):
		// Connection should be established now
	case <-ctx.Done():
		// Context cancelled, return early
		return fmt.Errorf("context cancelled while waiting for RTSP connection to establish")
	}

	return nil
}

// monitorReader monitors the reader process and restarts if needed
func (kr *RTSPKeepaliveReader) monitorReader(ctx context.Context, session *keepaliveSession) {
	defer close(session.done)

	for {
		select {
		case <-ctx.Done():
			// Context cancelled, stop monitoring
			if session.cmd != nil && session.cmd.Process != nil {
				session.cmd.Process.Kill()
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
					kr.logger.WithFields(logging.Fields{
						"path":  session.pathName,
						"error": err.Error(),
					}).Warn("Keepalive reader exited, restarting...")

					// Wait before restart to avoid rapid cycling using context-aware timeout
					select {
					case <-time.After(2 * time.Second):
						// Backoff period completed, proceed with restart
					case <-ctx.Done():
						// Context cancelled, exit early
						return
					}

					// Restart the reader
					if err := kr.startReader(ctx, session); err != nil {
						kr.logger.WithError(err).Error("Failed to restart keepalive reader")
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
		kr.StopKeepalive(pathName)
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
