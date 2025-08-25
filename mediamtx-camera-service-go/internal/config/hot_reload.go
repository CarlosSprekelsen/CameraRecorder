package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
)

// ConfigWatcher handles hot reload functionality for configuration files.
type ConfigWatcher struct {
	watcher        *fsnotify.Watcher
	configPath     string
	reloadCallback func(*Config) error
	logger         *logrus.Logger
	mu             sync.RWMutex
	isRunning      bool
	ctx            context.Context
	cancel         context.CancelFunc
}

// NewConfigWatcher creates a new configuration watcher.
func NewConfigWatcher(configPath string, reloadCallback func(*Config) error) (*ConfigWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &ConfigWatcher{
		watcher:        watcher,
		configPath:     configPath,
		reloadCallback: reloadCallback,
		logger:         logrus.New(),
		ctx:            ctx,
		cancel:         cancel,
	}, nil
}

// Start begins watching the configuration file for changes.
func (cw *ConfigWatcher) Start() error {
	cw.mu.Lock()
	defer cw.mu.Unlock()

	if cw.isRunning {
		return fmt.Errorf("config watcher is already running")
	}

	// Verify the configuration file exists
	if _, err := os.Stat(cw.configPath); os.IsNotExist(err) {
		return fmt.Errorf("configuration file does not exist: %s", cw.configPath)
	}

	// Watch the directory containing the configuration file
	configDir := filepath.Dir(cw.configPath)
	if err := cw.watcher.Add(configDir); err != nil {
		return fmt.Errorf("failed to watch directory %s: %w", configDir, err)
	}

	cw.isRunning = true
	cw.logger.Info("Configuration hot reload started")

	// Start the file watching goroutine
	go cw.watchLoop()

	return nil
}

// Stop stops watching the configuration file.
func (cw *ConfigWatcher) Stop() error {
	cw.mu.Lock()
	defer cw.mu.Unlock()

	if !cw.isRunning {
		return nil
	}

	cw.cancel()
	cw.isRunning = false

	if err := cw.watcher.Close(); err != nil {
		return fmt.Errorf("failed to close file watcher: %w", err)
	}

	cw.logger.Info("Configuration hot reload stopped")
	return nil
}

// IsRunning returns whether the watcher is currently running.
func (cw *ConfigWatcher) IsRunning() bool {
	cw.mu.RLock()
	defer cw.mu.RUnlock()
	return cw.isRunning
}

// watchLoop handles the file system events.
func (cw *ConfigWatcher) watchLoop() {
	var lastReloadTime time.Time
	debounceInterval := 500 * time.Millisecond

	for {
		select {
		case <-cw.ctx.Done():
			return

		case event, ok := <-cw.watcher.Events:
			if !ok {
				return
			}

			// Check if the event is for our configuration file
			if filepath.Clean(event.Name) == filepath.Clean(cw.configPath) {
				// Debounce rapid file changes
				if time.Since(lastReloadTime) < debounceInterval {
					cw.logger.Debug("Ignoring rapid configuration file change (debounced)")
					continue
				}

				// Handle different event types
				switch event.Op {
				case fsnotify.Write:
					cw.logger.Info("Configuration file modified, reloading...")
					if err := cw.reloadConfig(); err != nil {
						cw.logger.Errorf("Failed to reload configuration: %v", err)
					} else {
						lastReloadTime = time.Now()
					}

				case fsnotify.Remove:
					cw.logger.Warn("Configuration file removed")
					// Continue watching in case the file is recreated

				case fsnotify.Create:
					cw.logger.Info("Configuration file created")
					if err := cw.reloadConfig(); err != nil {
						cw.logger.Errorf("Failed to reload configuration: %v", err)
					} else {
						lastReloadTime = time.Now()
					}

				case fsnotify.Rename:
					cw.logger.Info("Configuration file renamed")
					// Continue watching in case a new file is created with the same name
				}
			}

		case err, ok := <-cw.watcher.Errors:
			if !ok {
				return
			}
			cw.logger.Errorf("File watcher error: %v", err)
		}
	}
}

// reloadConfig reloads the configuration file and calls the reload callback.
func (cw *ConfigWatcher) reloadConfig() error {
	// Wait for file to be stable (no size changes)
	if err := cw.waitForFileStable(); err != nil {
		return fmt.Errorf("failed to wait for file stability: %w", err)
	}

	// Create a new config loader and load the configuration
	loader := NewConfigLoader()
	config, err := loader.LoadConfig(cw.configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Call the reload callback
	if cw.reloadCallback != nil {
		if err := cw.reloadCallback(config); err != nil {
			return fmt.Errorf("reload callback failed: %w", err)
		}
	}

	cw.logger.Info("Configuration reloaded successfully")
	return nil
}

// waitForFileStable waits for the configuration file to be stable (no size changes).
func (cw *ConfigWatcher) waitForFileStable() error {
	const (
		maxWaitTime    = 5 * time.Second
		checkInterval  = 100 * time.Millisecond
		stabilityCount = 3
	)

	startTime := time.Now()
	lastSize := int64(-1)
	stableChecks := 0

	for time.Since(startTime) < maxWaitTime {
		stat, err := os.Stat(cw.configPath)
		if err != nil {
			if os.IsNotExist(err) {
				// File might be temporarily unavailable during write
				time.Sleep(checkInterval)
				continue
			}
			return fmt.Errorf("failed to stat configuration file: %w", err)
		}

		currentSize := stat.Size()
		if currentSize == lastSize {
			stableChecks++
			if stableChecks >= stabilityCount {
				return nil // File is stable
			}
		} else {
			stableChecks = 0
			lastSize = currentSize
		}

		time.Sleep(checkInterval)
	}

	return fmt.Errorf("configuration file did not stabilize within %v", maxWaitTime)
}

// GetViper returns the underlying Viper instance for advanced usage.
// This method is included for backward compatibility and advanced use cases.
func (cw *ConfigWatcher) GetWatcher() *fsnotify.Watcher {
	return cw.watcher
}
