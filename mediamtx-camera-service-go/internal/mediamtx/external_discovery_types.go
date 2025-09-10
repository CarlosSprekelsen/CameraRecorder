/*
External stream discovery types and structures.

Provides type definitions for external stream discovery including Skydio UAVs
and generic RTSP sources with STANAG 4609 compliance.

Requirements Coverage:
- REQ-MTX-001: External stream discovery and management
- REQ-MTX-002: STANAG 4609 compliance for UAV streams
- REQ-MTX-003: Configurable discovery parameters

Test Categories: Unit/Integration
API Documentation Reference: docs/api/external_discovery.md
*/

package mediamtx

import (
	"time"
)

// ExternalStream represents a discovered external stream
type ExternalStream struct {
	URL          string                 `json:"url"`
	Type         string                 `json:"type"` // "skydio_stanag4609", "generic_rtsp", etc.
	Name         string                 `json:"name"`
	Status       string                 `json:"status"` // "discovered", "connected", "error", "disconnected"
	DiscoveredAt time.Time              `json:"discovered_at"`
	LastSeen     time.Time              `json:"last_seen"`
	Capabilities map[string]interface{} `json:"capabilities"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// ExternalDiscoveryConfig represents discovery configuration
type ExternalDiscoveryConfig struct {
	Enabled            bool                  `json:"enabled"`
	ScanInterval       int                   `json:"scan_interval"`
	ScanTimeout        int                   `json:"scan_timeout"`
	MaxConcurrentScans int                   `json:"max_concurrent_scans"`
	EnableStartupScan  bool                  `json:"enable_startup_scan"`
	Skydio             SkydioDiscoveryConfig `json:"skydio"`
	GenericUAV         GenericUAVConfig      `json:"generic_uav"`
}

// SkydioDiscoveryConfig represents Skydio-specific configuration
type SkydioDiscoveryConfig struct {
	Enabled           bool     `json:"enabled"`
	NetworkRanges     []string `json:"network_ranges"`
	EOPort            int      `json:"eo_port"`
	IRPort            int      `json:"ir_port"`
	EOStreamPath      string   `json:"eo_stream_path"`
	IRStreamPath      string   `json:"ir_stream_path"`
	EnableBothStreams bool     `json:"enable_both_streams"`
	KnownIPs          []string `json:"known_ips"`
}

// GenericUAVConfig represents generic UAV configuration
type GenericUAVConfig struct {
	Enabled       bool     `json:"enabled"`
	NetworkRanges []string `json:"network_ranges"`
	CommonPorts   []int    `json:"common_ports"`
	StreamPaths   []string `json:"stream_paths"`
	KnownIPs      []string `json:"known_ips"`
}

// DiscoveryOptions represents options for discovery operations
type DiscoveryOptions struct {
	SkydioEnabled  bool `json:"skydio_enabled"`
	GenericEnabled bool `json:"generic_enabled"`
	ForceRescan    bool `json:"force_rescan"`
	IncludeOffline bool `json:"include_offline"`
}

// DiscoveryResult represents the result of a discovery operation
type DiscoveryResult struct {
	DiscoveredStreams []*ExternalStream `json:"discovered_streams"`
	SkydioStreams     []*ExternalStream `json:"skydio_streams"`
	GenericStreams    []*ExternalStream `json:"generic_streams"`
	ScanTimestamp     int64             `json:"scan_timestamp"`
	TotalFound        int               `json:"total_found"`
	DiscoveryOptions  DiscoveryOptions  `json:"discovery_options"`
	ScanDuration      time.Duration     `json:"scan_duration"`
	Errors            []string          `json:"errors,omitempty"`
}

// StreamHealth represents the health status of an external stream
type StreamHealth struct {
	URL          string    `json:"url"`
	Status       string    `json:"status"`
	LastChecked  time.Time `json:"last_checked"`
	ResponseTime int64     `json:"response_time_ms"`
	ErrorCount   int       `json:"error_count"`
	SuccessCount int       `json:"success_count"`
	LastError    string    `json:"last_error,omitempty"`
}

// NetworkRange represents a network range for scanning
type NetworkRange struct {
	CIDR    string   `json:"cidr"`
	IPs     []string `json:"ips,omitempty"`
	Enabled bool     `json:"enabled"`
}

// StreamValidationResult represents the result of stream validation
type StreamValidationResult struct {
	Valid        bool                   `json:"valid"`
	StreamType   string                 `json:"stream_type"`
	Capabilities map[string]interface{} `json:"capabilities"`
	Error        string                 `json:"error,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}
