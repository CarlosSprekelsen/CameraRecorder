package mediamtx

import "fmt"

// MediaMTX API Version - Change this single constant to upgrade API versions
const MediaMTXAPIVersion = "v3"

// MediaMTX API Endpoint Constants
// These constants provide a single source of truth for all MediaMTX API endpoints
// and enable easy migration to future API versions (v4, v5, etc.)

const (
	// Base paths
	mediamtxConfigBasePath     = "/" + MediaMTXAPIVersion + "/config"
	mediamtxPathsBasePath      = "/" + MediaMTXAPIVersion + "/paths"
	mediamtxRecordingsBasePath = "/" + MediaMTXAPIVersion + "/recordings"
	mediamtxAuthBasePath       = "/" + MediaMTXAPIVersion + "/auth"

	// Configuration endpoints
	MediaMTXConfigGlobalGet    = mediamtxConfigBasePath + "/global/get"
	MediaMTXConfigGlobalPatch  = mediamtxConfigBasePath + "/global/patch"
	MediaMTXConfigPathsList    = mediamtxConfigBasePath + "/paths/list"
	MediaMTXConfigPathsGet     = mediamtxConfigBasePath + "/paths/get/%s"     // %s = name
	MediaMTXConfigPathsAdd     = mediamtxConfigBasePath + "/paths/add/%s"     // %s = name
	MediaMTXConfigPathsPatch   = mediamtxConfigBasePath + "/paths/patch/%s"   // %s = name
	MediaMTXConfigPathsReplace = mediamtxConfigBasePath + "/paths/replace/%s" // %s = name
	MediaMTXConfigPathsDelete  = mediamtxConfigBasePath + "/paths/delete/%s"  // %s = name

	// Path defaults endpoints
	MediaMTXConfigPathDefaultsGet   = mediamtxConfigBasePath + "/pathdefaults/get"
	MediaMTXConfigPathDefaultsPatch = mediamtxConfigBasePath + "/pathdefaults/patch"

	// Runtime paths endpoints (actual path status, not configuration)
	MediaMTXPathsList = mediamtxPathsBasePath + "/list"
	MediaMTXPathsGet  = mediamtxPathsBasePath + "/get/%s" // %s = name

	// Recording endpoints
	MediaMTXRecordingsList          = mediamtxRecordingsBasePath + "/list"
	MediaMTXRecordingsGet           = mediamtxRecordingsBasePath + "/get/%s" // %s = name
	MediaMTXRecordingsDeleteSegment = mediamtxRecordingsBasePath + "/deletesegment"

	// Authentication endpoints
	MediaMTXAuthJWKSRefresh = mediamtxAuthBasePath + "/jwks/refresh"

	// HLS endpoints
	MediaMTXHLSMuxersList = "/" + MediaMTXAPIVersion + "/hlsmuxers/list"
	MediaMTXHLSMuxersGet  = "/" + MediaMTXAPIVersion + "/hlsmuxers/get/%s" // %s = name

	// RTSP connection endpoints
	MediaMTXRTSPConnsList = "/" + MediaMTXAPIVersion + "/rtspconns/list"
	MediaMTXRTSPConnsGet  = "/" + MediaMTXAPIVersion + "/rtspconns/get/%s" // %s = id

	// RTSP session endpoints
	MediaMTXRTSPSessionsList = "/" + MediaMTXAPIVersion + "/rtspsessions/list"
	MediaMTXRTSPSessionsGet  = "/" + MediaMTXAPIVersion + "/rtspsessions/get/%s"  // %s = id
	MediaMTXRTSPSessionsKick = "/" + MediaMTXAPIVersion + "/rtspsessions/kick/%s" // %s = id

	// RTSPS connection endpoints
	MediaMTXRTSPSConnsList = "/" + MediaMTXAPIVersion + "/rtspsconns/list"
	MediaMTXRTSPSConnsGet  = "/" + MediaMTXAPIVersion + "/rtspsconns/get/%s" // %s = id

	// RTSPS session endpoints
	MediaMTXRTSPSSessionsList = "/" + MediaMTXAPIVersion + "/rtspssessions/list"
	MediaMTXRTSPSSessionsGet  = "/" + MediaMTXAPIVersion + "/rtspssessions/get/%s"  // %s = id
	MediaMTXRTSPSSessionsKick = "/" + MediaMTXAPIVersion + "/rtspssessions/kick/%s" // %s = id

	// RTMP connection endpoints
	MediaMTXRTMPConnsList = "/" + MediaMTXAPIVersion + "/rtmpconns/list"
	MediaMTXRTMPConnsGet  = "/" + MediaMTXAPIVersion + "/rtmpconns/get/%s"  // %s = id
	MediaMTXRTMPConnsKick = "/" + MediaMTXAPIVersion + "/rtmpconns/kick/%s" // %s = id

	// RTMPS connection endpoints
	MediaMTXRTMPSConnsList = "/" + MediaMTXAPIVersion + "/rtmpsconns/list"
	MediaMTXRTMPSConnsGet  = "/" + MediaMTXAPIVersion + "/rtmpsconns/get/%s"  // %s = id
	MediaMTXRTMPSConnsKick = "/" + MediaMTXAPIVersion + "/rtmpsconns/kick/%s" // %s = id

	// SRT connection endpoints
	MediaMTXSRTConnsList = "/" + MediaMTXAPIVersion + "/srtconns/list"
	MediaMTXSRTConnsGet  = "/" + MediaMTXAPIVersion + "/srtconns/get/%s"  // %s = id
	MediaMTXSRTConnsKick = "/" + MediaMTXAPIVersion + "/srtconns/kick/%s" // %s = id

	// WebRTC session endpoints
	MediaMTXWebRTCSessionsList = "/" + MediaMTXAPIVersion + "/webrtcsessions/list"
	MediaMTXWebRTCSessionsGet  = "/" + MediaMTXAPIVersion + "/webrtcsessions/get/%s"  // %s = id
	MediaMTXWebRTCSessionsKick = "/" + MediaMTXAPIVersion + "/webrtcsessions/kick/%s" // %s = id
)

// Helper functions for formatted endpoints

// FormatConfigPathsGet returns the formatted endpoint for getting a specific path configuration
func FormatConfigPathsGet(name string) string {
	return fmt.Sprintf(MediaMTXConfigPathsGet, name)
}

// FormatConfigPathsAdd returns the formatted endpoint for adding a path configuration
func FormatConfigPathsAdd(name string) string {
	return fmt.Sprintf(MediaMTXConfigPathsAdd, name)
}

// FormatConfigPathsPatch returns the formatted endpoint for patching a path configuration
func FormatConfigPathsPatch(name string) string {
	return fmt.Sprintf(MediaMTXConfigPathsPatch, name)
}

// FormatConfigPathsReplace returns the formatted endpoint for replacing a path configuration
func FormatConfigPathsReplace(name string) string {
	return fmt.Sprintf(MediaMTXConfigPathsReplace, name)
}

// FormatConfigPathsDelete returns the formatted endpoint for deleting a path configuration
func FormatConfigPathsDelete(name string) string {
	return fmt.Sprintf(MediaMTXConfigPathsDelete, name)
}

// FormatPathsGet returns the formatted endpoint for getting a specific path status
func FormatPathsGet(name string) string {
	return fmt.Sprintf(MediaMTXPathsGet, name)
}

// FormatRecordingsGet returns the formatted endpoint for getting recordings for a path
func FormatRecordingsGet(name string) string {
	return fmt.Sprintf(MediaMTXRecordingsGet, name)
}

// FormatHLSMuxersGet returns the formatted endpoint for getting a specific HLS muxer
func FormatHLSMuxersGet(name string) string {
	return fmt.Sprintf(MediaMTXHLSMuxersGet, name)
}

// FormatRTSPConnsGet returns the formatted endpoint for getting a specific RTSP connection
func FormatRTSPConnsGet(id string) string {
	return fmt.Sprintf(MediaMTXRTSPConnsGet, id)
}

// FormatRTSPSessionsGet returns the formatted endpoint for getting a specific RTSP session
func FormatRTSPSessionsGet(id string) string {
	return fmt.Sprintf(MediaMTXRTSPSessionsGet, id)
}

// FormatRTSPSessionsKick returns the formatted endpoint for kicking a specific RTSP session
func FormatRTSPSessionsKick(id string) string {
	return fmt.Sprintf(MediaMTXRTSPSessionsKick, id)
}

// Additional helper functions can be added here as needed...

// GetAPIVersion returns the current MediaMTX API version
func GetAPIVersion() string {
	return MediaMTXAPIVersion
}

// GetHealthCheckEndpoint returns the standard health check endpoint
func GetHealthCheckEndpoint() string {
	return MediaMTXPathsList // Using paths list as health check
}
