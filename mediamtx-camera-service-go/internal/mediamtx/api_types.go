/*
MediaMTX API Types - Single Source of Truth

This file contains ALL MediaMTX API-related type definitions to ensure
consistency across the codebase and prevent schema drift.

Requirements Coverage:
- REQ-MTX-001: MediaMTX service integration
- REQ-MTX-002: Stream management capabilities
- REQ-MTX-003: Path creation and deletion
- REQ-MTX-004: Health monitoring

API Documentation Reference: docs/api/json_rpc_methods.md
*/

package mediamtx

import "time"

// ============================================================================
// MEDIAMTX API RESPONSE TYPES (Runtime Status)
// ============================================================================

// MediaMTXPathResponse represents the actual response from MediaMTX /v3/paths/get/{name} endpoint
// This is the SINGLE SOURCE OF TRUTH for MediaMTX runtime path responses
type MediaMTXPathResponse struct {
	Name          string        `json:"name"`
	ConfName      string        `json:"confName"`
	Source        interface{}   `json:"source"` // Can be null, string, or object
	Ready         bool          `json:"ready"`
	ReadyTime     interface{}   `json:"readyTime"` // Can be null or timestamp
	Tracks        []interface{} `json:"tracks"`
	BytesReceived int64         `json:"bytesReceived"`
	BytesSent     int64         `json:"bytesSent"`
	Readers       []interface{} `json:"readers"`
}

// MediaMTXPathsListResponse represents the response from MediaMTX /v3/paths/list endpoint
// This is the SINGLE SOURCE OF TRUTH for MediaMTX paths list responses
type MediaMTXPathsListResponse struct {
	ItemCount int                    `json:"itemCount"`
	PageCount int                    `json:"pageCount"`
	Items     []MediaMTXPathResponse `json:"items"`
}

// MediaMTXHealthResponse represents the response from MediaMTX health endpoints
// This is the SINGLE SOURCE OF TRUTH for MediaMTX health responses
type MediaMTXHealthResponse struct {
	Status    string  `json:"status"`
	Timestamp string  `json:"timestamp"`
	Metrics   Metrics `json:"metrics"`
}

// ============================================================================
// MEDIAMTX API REQUEST TYPES (Configuration)
// ============================================================================

// MediaMTXPathConfig represents MediaMTX path configuration for API requests
// This is the SINGLE SOURCE OF TRUTH for MediaMTX path configuration
type MediaMTXPathConfig struct {
	ID                         string        `json:"id"`
	Name                       string        `json:"name"`
	Source                     string        `json:"source"`
	SourceOnDemand             bool          `json:"source_on_demand"`
	SourceOnDemandStartTimeout time.Duration `json:"source_on_demand_start_timeout"`
	SourceOnDemandCloseAfter   time.Duration `json:"source_on_demand_close_after"`
	PublishUser                string        `json:"publish_user"`
	PublishPass                string        `json:"publish_pass"`
	ReadUser                   string        `json:"read_user"`
	ReadPass                   string        `json:"read_pass"`
	RunOnDemand                string        `json:"run_on_demand"`
	RunOnDemandRestart         bool          `json:"run_on_demand_restart"`
	RunOnDemandCloseAfter      time.Duration `json:"run_on_demand_close_after"`
	RunOnDemandStartTimeout    time.Duration `json:"run_on_demand_start_timeout"`
}

// MediaMTXCreatePathRequest represents a path creation request to MediaMTX API
type MediaMTXCreatePathRequest struct {
	Name   string                 `json:"name"`
	Source string                 `json:"source"`
	Config map[string]interface{} `json:"config,omitempty"`
}

// MediaMTXUpdatePathRequest represents a path update request to MediaMTX API
type MediaMTXUpdatePathRequest struct {
	Config map[string]interface{} `json:"config"`
}

// ============================================================================
// MEDIAMTX API COMPONENT TYPES
// ============================================================================

// MediaMTXPathSource represents the source configuration in MediaMTX responses
type MediaMTXPathSource struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// MediaMTXPathReader represents a reader connected to a MediaMTX path
type MediaMTXPathReader struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// ============================================================================
// LEGACY ALIASES (for backward compatibility)
// ============================================================================

// Path is an alias for MediaMTXPathConfig to maintain backward compatibility
// DEPRECATED: Use MediaMTXPathConfig directly for new code
type Path = MediaMTXPathConfig

// PathSource is an alias for MediaMTXPathSource to maintain backward compatibility
// DEPRECATED: Use MediaMTXPathSource directly for new code
type PathSource = MediaMTXPathSource

// PathReader is an alias for MediaMTXPathReader to maintain backward compatibility
// DEPRECATED: Use MediaMTXPathReader directly for new code
type PathReader = MediaMTXPathReader
