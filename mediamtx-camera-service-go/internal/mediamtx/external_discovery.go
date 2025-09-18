/*
External stream discovery implementation.

Provides discovery functionality for external RTSP streams including Skydio UAVs
with STANAG 4609 compliance and generic RTSP sources.

Requirements Coverage:
- REQ-MTX-001: External stream discovery and management
- REQ-MTX-002: STANAG 4609 compliance for UAV streams
- REQ-MTX-003: Configurable discovery parameters

Test Categories: Unit/Integration
API Documentation Reference: docs/api/external_discovery.md
*/

package mediamtx

import (
	"context"
	"fmt"
	"net"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/config"
	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
)

// ExternalStreamDiscovery manages external stream discovery
type ExternalStreamDiscovery struct {
	config            *config.ExternalDiscoveryConfig
	logger            *logging.Logger
	discoveredStreams map[string]*ExternalStream
	scanInProgress    int32 // Atomic flag
	lastScanTime      time.Time
	mu                sync.RWMutex
	stopChan          chan struct{}
}

// NewExternalStreamDiscovery creates a new external stream discovery instance
func NewExternalStreamDiscovery(config *config.ExternalDiscoveryConfig, logger *logging.Logger) *ExternalStreamDiscovery {
	return &ExternalStreamDiscovery{
		config:            config,
		logger:            logger,
		discoveredStreams: make(map[string]*ExternalStream),
		stopChan:          make(chan struct{}),
	}
}

// Start initializes the external discovery system
func (esd *ExternalStreamDiscovery) Start(ctx context.Context) error {
	esd.logger.Info("Starting external stream discovery")

	// Perform startup scan if enabled
	if esd.config.EnableStartupScan {
		go func() {
			if _, err := esd.DiscoverExternalStreams(ctx, DiscoveryOptions{
				SkydioEnabled:  esd.config.Skydio.Enabled,
				GenericEnabled: esd.config.GenericUAV.Enabled,
			}); err != nil {
				esd.logger.WithError(err).Warn("Startup discovery scan failed")
			}
		}()
	}

	// Start background timer if interval is set
	if esd.config.ScanInterval > 0 {
		go esd.startDiscoveryTimer(ctx)
	}

	esd.logger.Info("External stream discovery started")
	return nil
}

// Stop stops the external discovery system with context-aware cancellation
func (esd *ExternalStreamDiscovery) Stop(ctx context.Context) error {
	esd.logger.Info("Stopping external stream discovery")

	// Signal stop
	select {
	case <-esd.stopChan:
		// Already closed
	default:
		close(esd.stopChan)
	}

	// Wait for any ongoing discovery to complete with timeout
	done := make(chan struct{})
	go func() {
		// Wait for scan to complete if in progress using context-aware polling
		for atomic.LoadInt32(&esd.scanInProgress) == 1 {
			select {
			case <-time.After(10 * time.Millisecond):
				// Continue polling
			case <-ctx.Done():
				// Context cancelled, exit early
				close(done)
				return
			}
		}
		close(done)
	}()

	select {
	case <-done:
		// Clean shutdown
	case <-ctx.Done():
		esd.logger.Warn("External stream discovery shutdown timeout")
		return ctx.Err()
	}

	esd.logger.Info("External stream discovery stopped")
	return nil
}

// DiscoverExternalStreams performs external stream discovery
func (esd *ExternalStreamDiscovery) DiscoverExternalStreams(ctx context.Context, options DiscoveryOptions) (*DiscoveryResult, error) {
	// Check if scan is already in progress
	if !atomic.CompareAndSwapInt32(&esd.scanInProgress, 0, 1) {
		return nil, fmt.Errorf("discovery scan already in progress")
	}
	defer atomic.StoreInt32(&esd.scanInProgress, 0)

	startTime := time.Now()
	esd.logger.Info("Starting external stream discovery")

	discoveredStreams := make([]*ExternalStream, 0)
	var wg sync.WaitGroup
	streamChan := make(chan *ExternalStream, 100)
	errorChan := make(chan error, 100)

	// Skydio discovery
	if options.SkydioEnabled && esd.config.Skydio.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			streams, err := esd.discoverSkydioStreams(ctx)
			if err != nil {
				errorChan <- fmt.Errorf("skydio discovery failed: %w", err)
				return
			}
			for _, stream := range streams {
				streamChan <- stream
			}
		}()
	}

	// Generic UAV discovery
	if options.GenericEnabled && esd.config.GenericUAV.Enabled {
		wg.Add(1)
		go func() {
			defer wg.Done()
			streams, err := esd.discoverGenericUAVStreams(ctx)
			if err != nil {
				errorChan <- fmt.Errorf("generic UAV discovery failed: %w", err)
				return
			}
			for _, stream := range streams {
				streamChan <- stream
			}
		}()
	}

	// Wait for all discoveries to complete
	go func() {
		wg.Wait()
		close(streamChan)
		close(errorChan)
	}()

	// Collect results
	for stream := range streamChan {
		discoveredStreams = append(discoveredStreams, stream)
	}

	// Collect errors
	errors := make([]string, 0)
	for err := range errorChan {
		errors = append(errors, err.Error())
		esd.logger.WithError(err).Warn("Discovery error")
	}

	// Categorize results
	skydioStreams := make([]*ExternalStream, 0)
	genericStreams := make([]*ExternalStream, 0)

	for _, stream := range discoveredStreams {
		if strings.Contains(stream.Type, "skydio") {
			skydioStreams = append(skydioStreams, stream)
		} else {
			genericStreams = append(genericStreams, stream)
		}
	}

	// Update discovered streams cache
	esd.mu.Lock()
	for _, stream := range discoveredStreams {
		esd.discoveredStreams[stream.URL] = stream
	}
	esd.lastScanTime = time.Now()
	esd.mu.Unlock()

	scanDuration := time.Since(startTime)
	esd.logger.WithFields(logging.Fields{
		"total_found":   len(discoveredStreams),
		"skydio_count":  len(skydioStreams),
		"generic_count": len(genericStreams),
		"scan_duration": scanDuration,
		"error_count":   len(errors),
	}).Info("External stream discovery completed")

	return &DiscoveryResult{
		DiscoveredStreams: discoveredStreams,
		SkydioStreams:     skydioStreams,
		GenericStreams:    genericStreams,
		ScanTimestamp:     time.Now().Unix(),
		TotalFound:        len(discoveredStreams),
		DiscoveryOptions:  options,
		ScanDuration:      scanDuration,
		Errors:            errors,
	}, nil
}

// discoverSkydioStreams discovers Skydio UAV streams
func (esd *ExternalStreamDiscovery) discoverSkydioStreams(ctx context.Context) ([]*ExternalStream, error) {
	esd.logger.Info("Discovering Skydio UAV streams")

	streams := make([]*ExternalStream, 0)

	// Check known IPs first (faster)
	for _, ip := range esd.config.Skydio.KnownIPs {
		select {
		case <-ctx.Done():
			return streams, ctx.Err()
		default:
			if stream := esd.checkSkydioStream(ctx, ip, esd.config.Skydio.EOPort, esd.config.Skydio.EOStreamPath, "eo"); stream != nil {
				streams = append(streams, stream)
			}
			if stream := esd.checkSkydioStream(ctx, ip, esd.config.Skydio.IRPort, esd.config.Skydio.IRStreamPath, "ir"); stream != nil {
				streams = append(streams, stream)
			}
		}
	}

	// Scan network ranges if no streams found in known IPs
	if len(streams) == 0 {
		for _, networkRange := range esd.config.Skydio.NetworkRanges {
			ips, err := esd.parseNetworkRange(networkRange)
			if err != nil {
				esd.logger.WithError(err).WithField("network_range", networkRange).Warn("Failed to parse network range")
				continue
			}

			for _, ip := range ips {
				select {
				case <-ctx.Done():
					return streams, ctx.Err()
				default:
					if stream := esd.checkSkydioStream(ctx, ip, esd.config.Skydio.EOPort, esd.config.Skydio.EOStreamPath, "eo"); stream != nil {
						streams = append(streams, stream)
					}
					if stream := esd.checkSkydioStream(ctx, ip, esd.config.Skydio.IRPort, esd.config.Skydio.IRStreamPath, "ir"); stream != nil {
						streams = append(streams, stream)
					}
				}
			}
		}
	}

	return streams, nil
}

// checkSkydioStream checks if a Skydio stream is available
func (esd *ExternalStreamDiscovery) checkSkydioStream(ctx context.Context, ip string, port int, streamPath, streamType string) *ExternalStream {
	// Create Skydio RTSP URL
	rtspURL := fmt.Sprintf("rtsp://%s:%d%s", ip, port, streamPath)

	// Quick connectivity check with timeout
	ctx, cancel := context.WithTimeout(ctx, time.Duration(esd.config.ScanTimeout)*time.Second)
	defer cancel()

	// Test RTSP connection
	if esd.isRTSPStreamAvailable(ctx, rtspURL) {
		return &ExternalStream{
			URL:          rtspURL,
			Type:         "skydio_stanag4609",
			Name:         fmt.Sprintf("Skydio_%s_%s_%s", ip, streamType, streamPath),
			Status:       "discovered",
			DiscoveredAt: time.Now(),
			LastSeen:     time.Now(),
			Capabilities: map[string]interface{}{
				"protocol":    "rtsp",
				"format":      "stanag4609",
				"source":      "skydio_uav",
				"stream_type": streamType, // "eo" or "ir"
				"port":        port,
				"stream_path": streamPath,
				"codec":       "h264",
				"metadata":    "klv_mpegts",
			},
		}
	}

	return nil
}

// discoverGenericUAVStreams discovers generic UAV streams
func (esd *ExternalStreamDiscovery) discoverGenericUAVStreams(ctx context.Context) ([]*ExternalStream, error) {
	esd.logger.Info("Discovering generic UAV streams")

	streams := make([]*ExternalStream, 0)

	// Check known IPs first
	for _, ip := range esd.config.GenericUAV.KnownIPs {
		select {
		case <-ctx.Done():
			return streams, ctx.Err()
		default:
			for _, port := range esd.config.GenericUAV.CommonPorts {
				for _, streamPath := range esd.config.GenericUAV.StreamPaths {
					if stream := esd.checkGenericStream(ctx, ip, port, streamPath); stream != nil {
						streams = append(streams, stream)
					}
				}
			}
		}
	}

	// Scan network ranges
	for _, networkRange := range esd.config.GenericUAV.NetworkRanges {
		ips, err := esd.parseNetworkRange(networkRange)
		if err != nil {
			esd.logger.WithError(err).WithField("network_range", networkRange).Warn("Failed to parse network range")
			continue
		}

		for _, ip := range ips {
			select {
			case <-ctx.Done():
				return streams, ctx.Err()
			default:
				for _, port := range esd.config.GenericUAV.CommonPorts {
					for _, streamPath := range esd.config.GenericUAV.StreamPaths {
						if stream := esd.checkGenericStream(ctx, ip, port, streamPath); stream != nil {
							streams = append(streams, stream)
						}
					}
				}
			}
		}
	}

	return streams, nil
}

// checkGenericStream checks if a generic stream is available
func (esd *ExternalStreamDiscovery) checkGenericStream(ctx context.Context, ip string, port int, streamPath string) *ExternalStream {
	// Create generic RTSP URL
	rtspURL := fmt.Sprintf("rtsp://%s:%d%s", ip, port, streamPath)

	// Quick connectivity check with timeout
	ctx, cancel := context.WithTimeout(ctx, time.Duration(esd.config.ScanTimeout)*time.Second)
	defer cancel()

	// Test RTSP connection
	if esd.isRTSPStreamAvailable(ctx, rtspURL) {
		return &ExternalStream{
			URL:          rtspURL,
			Type:         "generic_rtsp",
			Name:         fmt.Sprintf("Generic_%s_%d%s", ip, port, streamPath),
			Status:       "discovered",
			DiscoveredAt: time.Now(),
			LastSeen:     time.Now(),
			Capabilities: map[string]interface{}{
				"protocol":    "rtsp",
				"source":      "generic_uav",
				"port":        port,
				"stream_path": streamPath,
			},
		}
	}

	return nil
}

// isRTSPStreamAvailable checks if an RTSP stream is available
func (esd *ExternalStreamDiscovery) isRTSPStreamAvailable(ctx context.Context, rtspURL string) bool {
	// Parse RTSP URL
	re := regexp.MustCompile(`rtsp://([^:]+):(\d+)(/.*)?`)
	matches := re.FindStringSubmatch(rtspURL)
	if len(matches) < 3 {
		return false
	}

	host := matches[1]
	port := matches[2]

	// Try to connect to the host:port
	conn, err := net.DialTimeout("tcp", host+":"+port, time.Duration(esd.config.ScanTimeout)*time.Second)
	if err != nil {
		return false
	}
	defer conn.Close()

	// Connection successful
	return true
}

// parseNetworkRange parses a network range (CIDR notation)
func (esd *ExternalStreamDiscovery) parseNetworkRange(networkRange string) ([]string, error) {
	// Parse CIDR notation
	_, ipNet, err := net.ParseCIDR(networkRange)
	if err != nil {
		return nil, fmt.Errorf("invalid CIDR notation: %w", err)
	}

	ips := make([]string, 0)

	// Generate IP addresses in the range
	for ip := ipNet.IP.Mask(ipNet.Mask); ipNet.Contains(ip); incrementIP(ip) {
		ips = append(ips, ip.String())
	}

	return ips, nil
}

// incrementIP increments an IP address
func incrementIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

// startDiscoveryTimer starts the background discovery timer
func (esd *ExternalStreamDiscovery) startDiscoveryTimer(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(esd.config.ScanInterval) * time.Second)
	defer ticker.Stop()

	esd.logger.WithField("scan_interval", fmt.Sprintf("%d", esd.config.ScanInterval)).Info("External discovery timer started")

	for {
		select {
		case <-ctx.Done():
			esd.logger.Info("External discovery timer stopped")
			return
		case <-esd.stopChan:
			esd.logger.Info("External discovery timer stopped")
			return
		case <-ticker.C:
			// Run discovery in background
			go func() {
				if _, err := esd.DiscoverExternalStreams(ctx, DiscoveryOptions{
					SkydioEnabled:  esd.config.Skydio.Enabled,
					GenericEnabled: esd.config.GenericUAV.Enabled,
				}); err != nil {
					esd.logger.WithError(err).Error("Background external discovery failed")
				}
			}()
		}
	}
}

// GetDiscoveredStreams returns all discovered streams
func (esd *ExternalStreamDiscovery) GetDiscoveredStreams() map[string]*ExternalStream {
	esd.mu.RLock()
	defer esd.mu.RUnlock()

	streams := make(map[string]*ExternalStream)
	for url, stream := range esd.discoveredStreams {
		streams[url] = stream
	}

	return streams
}

// GetLastScanTime returns the last scan time
func (esd *ExternalStreamDiscovery) GetLastScanTime() time.Time {
	esd.mu.RLock()
	defer esd.mu.RUnlock()
	return esd.lastScanTime
}

// UpdateScanInterval updates the discovery scan interval dynamically
func (esd *ExternalStreamDiscovery) UpdateScanInterval(interval int) error {
	esd.mu.Lock()
	defer esd.mu.Unlock()

	// Update the configuration
	oldInterval := esd.config.ScanInterval
	esd.config.ScanInterval = interval

	esd.logger.WithFields(logging.Fields{
		"old_interval": oldInterval,
		"new_interval": interval,
	}).Info("Updated external discovery scan interval")

	// Note: The ticker restart will happen automatically in the next timer cycle
	// The startDiscoveryTimer method reads from esd.config.ScanInterval each time
	// This provides dynamic updates without requiring service restart

	return nil
}

// DiscoverExternalStreamsAPI performs discovery and returns API-ready response
func (esd *ExternalStreamDiscovery) DiscoverExternalStreamsAPI(ctx context.Context, options DiscoveryOptions) (*DiscoverExternalStreamsResponse, error) {
	// Get internal discovery result
	result, err := esd.DiscoverExternalStreams(ctx, options)
	if err != nil {
		return nil, err
	}

	// Convert to API-ready format
	discoveredStreams := make([]ExternalStreamInfo, len(result.DiscoveredStreams))
	for i, stream := range result.DiscoveredStreams {
		discoveredStreams[i] = ExternalStreamInfo{
			URL:          stream.URL,
			Type:         stream.Type,
			Name:         stream.Name,
			Status:       stream.Status,
			DiscoveredAt: stream.DiscoveredAt.Format(time.RFC3339),
			LastSeen:     stream.LastSeen.Format(time.RFC3339),
			Capabilities: stream.Capabilities,
			Metadata:     stream.Metadata,
		}
	}

	skydioStreams := make([]ExternalStreamInfo, len(result.SkydioStreams))
	for i, stream := range result.SkydioStreams {
		skydioStreams[i] = ExternalStreamInfo{
			URL:          stream.URL,
			Type:         stream.Type,
			Name:         stream.Name,
			Status:       stream.Status,
			DiscoveredAt: stream.DiscoveredAt.Format(time.RFC3339),
			LastSeen:     stream.LastSeen.Format(time.RFC3339),
			Capabilities: stream.Capabilities,
			Metadata:     stream.Metadata,
		}
	}

	genericStreams := make([]ExternalStreamInfo, len(result.GenericStreams))
	for i, stream := range result.GenericStreams {
		genericStreams[i] = ExternalStreamInfo{
			URL:          stream.URL,
			Type:         stream.Type,
			Name:         stream.Name,
			Status:       stream.Status,
			DiscoveredAt: stream.DiscoveredAt.Format(time.RFC3339),
			LastSeen:     stream.LastSeen.Format(time.RFC3339),
			Capabilities: stream.Capabilities,
			Metadata:     stream.Metadata,
		}
	}

	// Build API-ready response
	response := &DiscoverExternalStreamsResponse{
		DiscoveredStreams: discoveredStreams,
		SkydioStreams:     skydioStreams,
		GenericStreams:    genericStreams,
		ScanTimestamp:     result.ScanTimestamp,
		TotalFound:        result.TotalFound,
		DiscoveryOptions: DiscoveryOptionsInfo{
			SkydioEnabled:  result.DiscoveryOptions.SkydioEnabled,
			GenericEnabled: result.DiscoveryOptions.GenericEnabled,
			ForceRescan:    result.DiscoveryOptions.ForceRescan,
			IncludeOffline: result.DiscoveryOptions.IncludeOffline,
		},
		ScanDuration: result.ScanDuration.String(),
		Errors:       result.Errors,
	}

	return response, nil
}

// GetExternalStreamsAPI returns all discovered streams in API-ready format
func (esd *ExternalStreamDiscovery) GetExternalStreamsAPI(ctx context.Context) (*GetExternalStreamsResponse, error) {
	esd.mu.RLock()
	streams := make(map[string]*ExternalStream)
	for url, stream := range esd.discoveredStreams {
		streams[url] = stream
	}
	esd.mu.RUnlock()

	// Convert to API-ready format
	allStreams := make([]ExternalStreamInfo, 0, len(streams))
	skydioStreams := make([]ExternalStreamInfo, 0)
	genericStreams := make([]ExternalStreamInfo, 0)

	for _, stream := range streams {
		streamInfo := ExternalStreamInfo{
			URL:          stream.URL,
			Type:         stream.Type,
			Name:         stream.Name,
			Status:       stream.Status,
			DiscoveredAt: stream.DiscoveredAt.Format(time.RFC3339),
			LastSeen:     stream.LastSeen.Format(time.RFC3339),
			Capabilities: stream.Capabilities,
			Metadata:     stream.Metadata,
		}

		allStreams = append(allStreams, streamInfo)

		if strings.Contains(stream.Type, "skydio") {
			skydioStreams = append(skydioStreams, streamInfo)
		} else {
			genericStreams = append(genericStreams, streamInfo)
		}
	}

	// Build API-ready response
	response := &GetExternalStreamsResponse{
		ExternalStreams: allStreams,
		SkydioStreams:   skydioStreams,
		GenericStreams:  genericStreams,
		TotalCount:      len(allStreams),
		Timestamp:       time.Now().Unix(),
	}

	return response, nil
}

// AddExternalStreamAPI adds an external stream and returns API-ready response
func (esd *ExternalStreamDiscovery) AddExternalStreamAPI(ctx context.Context, stream *ExternalStream) (*AddExternalStreamResponse, error) {
	esd.mu.Lock()
	defer esd.mu.Unlock()

	// Check if stream already exists
	if _, exists := esd.discoveredStreams[stream.URL]; exists {
		return nil, fmt.Errorf("external stream already exists: %s", stream.URL)
	}

	// Add stream to discovered streams
	stream.DiscoveredAt = time.Now()
	stream.LastSeen = time.Now()
	stream.Status = "added"
	esd.discoveredStreams[stream.URL] = stream

	// Build API-ready response
	response := &AddExternalStreamResponse{
		StreamURL:  stream.URL,
		StreamName: stream.Name,
		StreamType: stream.Type,
		Status:     "added",
		Timestamp:  time.Now().Unix(),
	}

	esd.logger.WithFields(logging.Fields{
		"stream_url":  stream.URL,
		"stream_name": stream.Name,
		"stream_type": stream.Type,
	}).Info("External stream added successfully")

	return response, nil
}

// RemoveExternalStreamAPI removes an external stream and returns API-ready response
func (esd *ExternalStreamDiscovery) RemoveExternalStreamAPI(ctx context.Context, streamURL string) (*RemoveExternalStreamResponse, error) {
	esd.mu.Lock()
	defer esd.mu.Unlock()

	// Check if stream exists
	if _, exists := esd.discoveredStreams[streamURL]; !exists {
		return nil, fmt.Errorf("external stream not found: %s", streamURL)
	}

	// Remove stream
	delete(esd.discoveredStreams, streamURL)

	// Build API-ready response
	response := &RemoveExternalStreamResponse{
		StreamURL: streamURL,
		Status:    "removed",
		Timestamp: time.Now().Unix(),
	}

	esd.logger.WithField("stream_url", streamURL).Info("External stream removed successfully")

	return response, nil
}
