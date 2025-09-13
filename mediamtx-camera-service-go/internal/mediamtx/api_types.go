/*
MediaMTX API Types - Single Source of Truth

This file contains ALL MediaMTX API-related type definitions to ensure
consistency across the codebase and prevent schema drift.

Requirements Coverage:
- REQ-MTX-001: MediaMTX service integration
- REQ-MTX-002: Stream management capabilities
- REQ-MTX-003: Path creation and deletion
- REQ-MTX-004: Health monitoring

API Documentation Reference: docs/api/swagger.json
*/

package mediamtx

// ============================================================================
// MEDIAMTX API RESPONSE TYPES (Runtime Status) - EXACT SWAGGER MATCH
// ============================================================================

// Path represents the actual response from MediaMTX /v3/paths/get/{name} endpoint
// This matches the Swagger "Path" schema exactly
type Path struct {
	Name          string       `json:"name"`
	ConfName      string       `json:"confName"`
	Source        *PathSource  `json:"source"` // Nullable PathSource
	Ready         bool         `json:"ready"`
	ReadyTime     *string      `json:"readyTime"` // Nullable string timestamp
	Tracks        []string     `json:"tracks"`    // Array of strings
	BytesReceived int64        `json:"bytesReceived"`
	BytesSent     int64        `json:"bytesSent"`
	Readers       []PathReader `json:"readers"` // Array of PathReader
}

// PathList represents the response from MediaMTX /v3/paths/list endpoint
// This matches the Swagger "PathList" schema exactly
type PathList struct {
	PageCount int    `json:"pageCount"`
	ItemCount int    `json:"itemCount"`
	Items     []Path `json:"items"`
}

// PathSource represents the source configuration in MediaMTX responses
// This matches the Swagger "PathSource" schema exactly with enum validation
type PathSource struct {
	Type string `json:"type"` // enum: hlsSource, redirect, rpiCameraSource, rtmpConn, rtmpSource, rtspSession, rtspSource, rtspsSession, srtConn, srtSource, mpegtsSource, rtpSource, webRTCSession, webRTCSource
	ID   string `json:"id"`
}

// PathReader represents a reader connected to a MediaMTX path
// This matches the Swagger "PathReader" schema exactly with enum validation
type PathReader struct {
	Type string `json:"type"` // enum: hlsMuxer, rtmpConn, rtspSession, rtspsSession, srtConn, webRTCSession
	ID   string `json:"id"`
}

// ============================================================================
// MEDIAMTX API CONFIGURATION TYPES - EXACT SWAGGER MATCH
// ============================================================================

// PathConf represents MediaMTX path configuration for API requests
// This matches the Swagger "PathConf" schema exactly
type PathConf struct {
	Name                       string `json:"name,omitempty"`
	Source                     string `json:"source,omitempty"`
	SourceFingerprint          string `json:"sourceFingerprint,omitempty"`
	SourceOnDemand             bool   `json:"sourceOnDemand,omitempty"`
	SourceOnDemandStartTimeout string `json:"sourceOnDemandStartTimeout,omitempty"`
	SourceOnDemandCloseAfter   string `json:"sourceOnDemandCloseAfter,omitempty"`
	MaxReaders                 int64  `json:"maxReaders,omitempty"`
	SrtReadPassphrase          string `json:"srtReadPassphrase,omitempty"`
	Fallback                   string `json:"fallback,omitempty"`
	UseAbsoluteTimestamp       bool   `json:"useAbsoluteTimestamp,omitempty"`
	Record                     bool   `json:"record,omitempty"`
	RecordPath                 string `json:"recordPath,omitempty"`
	RecordFormat               string `json:"recordFormat,omitempty"`
	RecordPartDuration         string `json:"recordPartDuration,omitempty"`
	RecordMaxPartSize          string `json:"recordMaxPartSize,omitempty"`
	RecordSegmentDuration      string `json:"recordSegmentDuration,omitempty"`
	RecordDeleteAfter          string `json:"recordDeleteAfter,omitempty"`
	OverridePublisher          bool   `json:"overridePublisher,omitempty"`
	SrtPublishPassphrase       string `json:"srtPublishPassphrase,omitempty"`
	RtspTransport              string `json:"rtspTransport,omitempty"`
	RtspAnyPort                bool   `json:"rtspAnyPort,omitempty"`
	RtspRangeType              string `json:"rtspRangeType,omitempty"`
	RtspRangeStart             string `json:"rtspRangeStart,omitempty"`
	RtspUDPReadBufferSize      int64  `json:"rtspUDPReadBufferSize,omitempty"`
	MpegtsUDPReadBufferSize    int64  `json:"mpegtsUDPReadBufferSize,omitempty"`
	RtpSDP                     string `json:"rtpSDP,omitempty"`
	RtpUDPReadBufferSize       int64  `json:"rtpUDPReadBufferSize,omitempty"`
	SourceRedirect             string `json:"sourceRedirect,omitempty"`
	// Raspberry Pi Camera specific fields
	RpiCameraCamID               int64     `json:"rpiCameraCamID,omitempty"`
	RpiCameraSecondary           bool      `json:"rpiCameraSecondary,omitempty"`
	RpiCameraWidth               int64     `json:"rpiCameraWidth,omitempty"`
	RpiCameraHeight              int64     `json:"rpiCameraHeight,omitempty"`
	RpiCameraHFlip               bool      `json:"rpiCameraHFlip,omitempty"`
	RpiCameraVFlip               bool      `json:"rpiCameraVFlip,omitempty"`
	RpiCameraBrightness          float64   `json:"rpiCameraBrightness,omitempty"`
	RpiCameraContrast            float64   `json:"rpiCameraContrast,omitempty"`
	RpiCameraSaturation          float64   `json:"rpiCameraSaturation,omitempty"`
	RpiCameraSharpness           float64   `json:"rpiCameraSharpness,omitempty"`
	RpiCameraExposure            string    `json:"rpiCameraExposure,omitempty"`
	RpiCameraAWB                 string    `json:"rpiCameraAWB,omitempty"`
	RpiCameraAWBGains            []float64 `json:"rpiCameraAWBGains,omitempty"`
	RpiCameraDenoise             string    `json:"rpiCameraDenoise,omitempty"`
	RpiCameraShutter             int64     `json:"rpiCameraShutter,omitempty"`
	RpiCameraMetering            string    `json:"rpiCameraMetering,omitempty"`
	RpiCameraGain                float64   `json:"rpiCameraGain,omitempty"`
	RpiCameraEV                  float64   `json:"rpiCameraEV,omitempty"`
	RpiCameraROI                 string    `json:"rpiCameraROI,omitempty"`
	RpiCameraHDR                 bool      `json:"rpiCameraHDR,omitempty"`
	RpiCameraTuningFile          string    `json:"rpiCameraTuningFile,omitempty"`
	RpiCameraMode                string    `json:"rpiCameraMode,omitempty"`
	RpiCameraFPS                 float64   `json:"rpiCameraFPS,omitempty"`
	RpiCameraAfMode              string    `json:"rpiCameraAfMode,omitempty"`
	RpiCameraAfRange             string    `json:"rpiCameraAfRange,omitempty"`
	RpiCameraAfSpeed             string    `json:"rpiCameraAfSpeed,omitempty"`
	RpiCameraLensPosition        float64   `json:"rpiCameraLensPosition,omitempty"`
	RpiCameraAfWindow            string    `json:"rpiCameraAfWindow,omitempty"`
	RpiCameraFlickerPeriod       int64     `json:"rpiCameraFlickerPeriod,omitempty"`
	RpiCameraTextOverlayEnable   bool      `json:"rpiCameraTextOverlayEnable,omitempty"`
	RpiCameraTextOverlay         string    `json:"rpiCameraTextOverlay,omitempty"`
	RpiCameraCodec               string    `json:"rpiCameraCodec,omitempty"`
	RpiCameraIDRPeriod           int64     `json:"rpiCameraIDRPeriod,omitempty"`
	RpiCameraBitrate             int64     `json:"rpiCameraBitrate,omitempty"`
	RpiCameraHardwareH264Profile string    `json:"rpiCameraHardwareH264Profile,omitempty"`
	RpiCameraHardwareH264Level   string    `json:"rpiCameraHardwareH264Level,omitempty"`
	RpiCameraSoftwareH264Profile string    `json:"rpiCameraSoftwareH264Profile,omitempty"`
	RpiCameraSoftwareH264Level   string    `json:"rpiCameraSoftwareH264Level,omitempty"`
	RpiCameraMJPEGQuality        int64     `json:"rpiCameraMJPEGQuality,omitempty"`
	// Script execution fields
	RunOnInit                  string `json:"runOnInit,omitempty"`
	RunOnInitRestart           bool   `json:"runOnInitRestart,omitempty"`
	RunOnDemand                string `json:"runOnDemand,omitempty"`
	RunOnDemandRestart         bool   `json:"runOnDemandRestart,omitempty"`
	RunOnDemandStartTimeout    string `json:"runOnDemandStartTimeout,omitempty"`
	RunOnDemandCloseAfter      string `json:"runOnDemandCloseAfter,omitempty"`
	RunOnUnDemand              string `json:"runOnUnDemand,omitempty"`
	RunOnReady                 string `json:"runOnReady,omitempty"`
	RunOnReadyRestart          bool   `json:"runOnReadyRestart,omitempty"`
	RunOnNotReady              string `json:"runOnNotReady,omitempty"`
	RunOnRead                  string `json:"runOnRead,omitempty"`
	RunOnReadRestart           bool   `json:"runOnReadRestart,omitempty"`
	RunOnUnread                string `json:"runOnUnread,omitempty"`
	RunOnRecordSegmentCreate   string `json:"runOnRecordSegmentCreate,omitempty"`
	RunOnRecordSegmentComplete string `json:"runOnRecordSegmentComplete,omitempty"`
}

// PathConfList represents the response from MediaMTX /v3/config/paths/list endpoint
// This matches the Swagger "PathConfList" schema exactly
type PathConfList struct {
	PageCount int        `json:"pageCount"`
	ItemCount int        `json:"itemCount"`
	Items     []PathConf `json:"items"`
}

// ============================================================================
// MEDIAMTX API ERROR TYPES - EXACT SWAGGER MATCH
// ============================================================================

// Error represents MediaMTX API error responses
// This matches the Swagger "Error" schema exactly
type Error struct {
	Error string `json:"error"`
}

// ============================================================================
// MEDIAMTX API GLOBAL CONFIGURATION TYPES - EXACT SWAGGER MATCH
// ============================================================================

// GlobalConf represents MediaMTX global configuration
// This matches the Swagger "GlobalConf" schema exactly
type GlobalConf struct {
	LogLevel                    string                       `json:"logLevel,omitempty"`
	LogDestinations             []string                     `json:"logDestinations,omitempty"`
	LogFile                     string                       `json:"logFile,omitempty"`
	SysLogPrefix                string                       `json:"sysLogPrefix,omitempty"`
	ReadTimeout                 string                       `json:"readTimeout,omitempty"`
	WriteTimeout                string                       `json:"writeTimeout,omitempty"`
	WriteQueueSize              int64                        `json:"writeQueueSize,omitempty"`
	UdpMaxPayloadSize           int64                        `json:"udpMaxPayloadSize,omitempty"`
	RunOnConnect                string                       `json:"runOnConnect,omitempty"`
	RunOnConnectRestart         bool                         `json:"runOnConnectRestart,omitempty"`
	RunOnDisconnect             string                       `json:"runOnDisconnect,omitempty"`
	AuthMethod                  string                       `json:"authMethod,omitempty"`
	AuthInternalUsers           []AuthInternalUser           `json:"authInternalUsers,omitempty"`
	AuthHTTPAddress             string                       `json:"authHTTPAddress,omitempty"`
	AuthHTTPExclude             []AuthInternalUserPermission `json:"authHTTPExclude,omitempty"`
	AuthJWTJWKS                 string                       `json:"authJWTJWKS,omitempty"`
	AuthJWTJWKSFingerprint      string                       `json:"authJWTJWKSFingerprint,omitempty"`
	AuthJWTClaimKey             string                       `json:"authJWTClaimKey,omitempty"`
	AuthJWTExclude              []AuthInternalUserPermission `json:"authJWTExclude,omitempty"`
	AuthJWTInHTTPQuery          bool                         `json:"authJWTInHTTPQuery,omitempty"`
	API                         bool                         `json:"api,omitempty"`
	APIAddress                  string                       `json:"apiAddress,omitempty"`
	APIEncryption               bool                         `json:"apiEncryption,omitempty"`
	APIServerKey                string                       `json:"apiServerKey,omitempty"`
	APIServerCert               string                       `json:"apiServerCert,omitempty"`
	APIAllowOrigin              string                       `json:"apiAllowOrigin,omitempty"`
	APITrustedProxies           []string                     `json:"apiTrustedProxies,omitempty"`
	Metrics                     bool                         `json:"metrics,omitempty"`
	MetricsAddress              string                       `json:"metricsAddress,omitempty"`
	MetricsEncryption           bool                         `json:"metricsEncryption,omitempty"`
	MetricsServerKey            string                       `json:"metricsServerKey,omitempty"`
	MetricsServerCert           string                       `json:"metricsServerCert,omitempty"`
	MetricsAllowOrigin          string                       `json:"metricsAllowOrigin,omitempty"`
	MetricsTrustedProxies       []string                     `json:"metricsTrustedProxies,omitempty"`
	Pprof                       bool                         `json:"pprof,omitempty"`
	PprofAddress                string                       `json:"pprofAddress,omitempty"`
	PprofEncryption             bool                         `json:"pprofEncryption,omitempty"`
	PprofServerKey              string                       `json:"pprofServerKey,omitempty"`
	PprofServerCert             string                       `json:"pprofServerCert,omitempty"`
	PprofAllowOrigin            string                       `json:"pprofAllowOrigin,omitempty"`
	PprofTrustedProxies         []string                     `json:"pprofTrustedProxies,omitempty"`
	Playback                    bool                         `json:"playback,omitempty"`
	PlaybackAddress             string                       `json:"playbackAddress,omitempty"`
	PlaybackEncryption          bool                         `json:"playbackEncryption,omitempty"`
	PlaybackServerKey           string                       `json:"playbackServerKey,omitempty"`
	PlaybackServerCert          string                       `json:"playbackServerCert,omitempty"`
	PlaybackAllowOrigin         string                       `json:"playbackAllowOrigin,omitempty"`
	PlaybackTrustedProxies      []string                     `json:"playbackTrustedProxies,omitempty"`
	Rtsp                        bool                         `json:"rtsp,omitempty"`
	RtspTransports              []string                     `json:"rtspTransports,omitempty"`
	RtspEncryption              string                       `json:"rtspEncryption,omitempty"`
	RtspAddress                 string                       `json:"rtspAddress,omitempty"`
	RtspsAddress                string                       `json:"rtspsAddress,omitempty"`
	RtpAddress                  string                       `json:"rtpAddress,omitempty"`
	RtcpAddress                 string                       `json:"rtcpAddress,omitempty"`
	MulticastIPRange            string                       `json:"multicastIPRange,omitempty"`
	MulticastRTPPort            int64                        `json:"multicastRTPPort,omitempty"`
	MulticastRTCPPort           int64                        `json:"multicastRTCPPort,omitempty"`
	SrtpAddress                 string                       `json:"srtpAddress,omitempty"`
	SrtcpAddress                string                       `json:"srtcpAddress,omitempty"`
	MulticastSRTPPort           int64                        `json:"multicastSRTPPort,omitempty"`
	MulticastSRTCPPort          int64                        `json:"multicastSRTCPPort,omitempty"`
	RtspServerKey               string                       `json:"rtspServerKey,omitempty"`
	RtspServerCert              string                       `json:"rtspServerCert,omitempty"`
	RtspAuthMethods             []string                     `json:"rtspAuthMethods,omitempty"`
	RtspUDPReadBufferSize       int64                        `json:"rtspUDPReadBufferSize,omitempty"`
	Rtmp                        bool                         `json:"rtmp,omitempty"`
	RtmpAddress                 string                       `json:"rtmpAddress,omitempty"`
	RtmpEncryption              string                       `json:"rtmpEncryption,omitempty"`
	RtmpsAddress                string                       `json:"rtmpsAddress,omitempty"`
	RtmpServerKey               string                       `json:"rtmpServerKey,omitempty"`
	RtmpServerCert              string                       `json:"rtmpServerCert,omitempty"`
	Hls                         bool                         `json:"hls,omitempty"`
	HlsAddress                  string                       `json:"hlsAddress,omitempty"`
	HlsEncryption               bool                         `json:"hlsEncryption,omitempty"`
	HlsServerKey                string                       `json:"hlsServerKey,omitempty"`
	HlsServerCert               string                       `json:"hlsServerCert,omitempty"`
	HlsAllowOrigin              string                       `json:"hlsAllowOrigin,omitempty"`
	HlsTrustedProxies           []string                     `json:"hlsTrustedProxies,omitempty"`
	HlsAlwaysRemux              bool                         `json:"hlsAlwaysRemux,omitempty"`
	HlsVariant                  string                       `json:"hlsVariant,omitempty"`
	HlsSegmentCount             int64                        `json:"hlsSegmentCount,omitempty"`
	HlsSegmentDuration          string                       `json:"hlsSegmentDuration,omitempty"`
	HlsPartDuration             string                       `json:"hlsPartDuration,omitempty"`
	HlsSegmentMaxSize           string                       `json:"hlsSegmentMaxSize,omitempty"`
	HlsDirectory                string                       `json:"hlsDirectory,omitempty"`
	HlsMuxerCloseAfter          string                       `json:"hlsMuxerCloseAfter,omitempty"`
	Webrtc                      bool                         `json:"webrtc,omitempty"`
	WebrtcAddress               string                       `json:"webrtcAddress,omitempty"`
	WebrtcEncryption            bool                         `json:"webrtcEncryption,omitempty"`
	WebrtcServerKey             string                       `json:"webrtcServerKey,omitempty"`
	WebrtcServerCert            string                       `json:"webrtcServerCert,omitempty"`
	WebrtcAllowOrigin           string                       `json:"webrtcAllowOrigin,omitempty"`
	WebrtcTrustedProxies        []string                     `json:"webrtcTrustedProxies,omitempty"`
	WebrtcLocalUDPAddress       string                       `json:"webrtcLocalUDPAddress,omitempty"`
	WebrtcLocalTCPAddress       string                       `json:"webrtcLocalTCPAddress,omitempty"`
	WebrtcIPsFromInterfaces     bool                         `json:"webrtcIPsFromInterfaces,omitempty"`
	WebrtcIPsFromInterfacesList []string                     `json:"webrtcIPsFromInterfacesList,omitempty"`
	WebrtcAdditionalHosts       []string                     `json:"webrtcAdditionalHosts,omitempty"`
	WebrtcICEServers2           []WebrtcICEServer            `json:"webrtcICEServers2,omitempty"`
	WebrtcHandshakeTimeout      string                       `json:"webrtcHandshakeTimeout,omitempty"`
	WebrtcTrackGatherTimeout    string                       `json:"webrtcTrackGatherTimeout,omitempty"`
	WebrtcSTUNGatherTimeout     string                       `json:"webrtcSTUNGatherTimeout,omitempty"`
	Srt                         bool                         `json:"srt,omitempty"`
	SrtAddress                  string                       `json:"srtAddress,omitempty"`
}

// AuthInternalUser represents internal authentication user configuration
type AuthInternalUser struct {
	User        string                       `json:"user"`
	Pass        string                       `json:"pass"`
	IPs         []string                     `json:"ips,omitempty"`
	Permissions []AuthInternalUserPermission `json:"permissions,omitempty"`
}

// AuthInternalUserPermission represents user permission configuration
type AuthInternalUserPermission struct {
	Action string `json:"action"`
	Path   string `json:"path"`
}

// WebrtcICEServer represents WebRTC ICE server configuration
type WebrtcICEServer struct {
	URL        string `json:"url"`
	Username   string `json:"username,omitempty"`
	Password   string `json:"password,omitempty"`
	ClientOnly bool   `json:"clientOnly,omitempty"`
}

// ============================================================================
// MEDIAMTX API REQUEST TYPES (Simplified for internal use)
// ============================================================================

// MediaMTXCreatePathRequest represents a path creation request to MediaMTX API
type MediaMTXCreatePathRequest struct {
	Name   string   `json:"name"`
	Source string   `json:"source"`
	Config PathConf `json:"config,omitempty"`
}

// MediaMTXUpdatePathRequest represents a path update request to MediaMTX API
type MediaMTXUpdatePathRequest struct {
	Config PathConf `json:"config"`
}

// ============================================================================
// MEDIAMTX HEALTH RESPONSE TYPES
// ============================================================================

// MediaMTXHealthResponse represents the response from MediaMTX health endpoints
type MediaMTXHealthResponse struct {
	Status    string  `json:"status"`
	Timestamp string  `json:"timestamp"`
	Metrics   Metrics `json:"metrics"`
}

// ============================================================================
// LEGACY ALIASES (for backward compatibility)
// ============================================================================

// MediaMTXPathResponse is an alias for Path to maintain backward compatibility
// DEPRECATED: Use Path directly for new code
type MediaMTXPathResponse = Path

// MediaMTXPathsListResponse is an alias for PathList to maintain backward compatibility
// DEPRECATED: Use PathList directly for new code
type MediaMTXPathsListResponse = PathList

// MediaMTXPathConfig is an alias for PathConf to maintain backward compatibility
// DEPRECATED: Use PathConf directly for new code
type MediaMTXPathConfig = PathConf

// MediaMTXPathSource is an alias for PathSource to maintain backward compatibility
// DEPRECATED: Use PathSource directly for new code
type MediaMTXPathSource = PathSource

// MediaMTXPathReader is an alias for PathReader to maintain backward compatibility
// DEPRECATED: Use PathReader directly for new code
type MediaMTXPathReader = PathReader
