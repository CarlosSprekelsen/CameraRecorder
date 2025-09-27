/*
MediaMTX Metadata Manager Implementation

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
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/logging"
)

// MetadataManager manages unified metadata extraction for video and image files.
//
// RESPONSIBILITIES:
// - FFprobe integration with JSON parsing for video and image metadata
// - Unified metadata structure for consistency across recording and snapshot operations
// - Performance optimization with timeout management and error handling
// - Integration with existing ConfigIntegration pattern
//
// ARCHITECTURE:
// - Follows existing ConfigIntegration pattern for centralized configuration
// - Uses structured logging consistent with other managers
// - Provides graceful degradation when metadata extraction fails
// - Implements timeout-based execution to prevent hanging operations
type MetadataManager struct {
	configIntegration *ConfigIntegration
	logger            *logging.Logger
	ffmpegManager     FFmpegManager
}

// MediaMetadata represents comprehensive metadata for video and image files
type MediaMetadata struct {
	// Common fields
	FilePath   string    `json:"file_path"`
	FileSize   int64     `json:"file_size"`
	Duration   *float64  `json:"duration,omitempty"` // Video duration in seconds
	Format     string    `json:"format"`             // File format (mp4, jpg, etc.)
	CreatedAt  time.Time `json:"created_at"`
	ModifiedAt time.Time `json:"modified_at"`

	// Video-specific fields
	VideoCodec *string `json:"video_codec,omitempty"` // Video codec name
	Bitrate    *int64  `json:"bitrate,omitempty"`     // Bitrate in bits per second

	// Image-specific fields
	Width      *int    `json:"width,omitempty"`      // Image width in pixels
	Height     *int    `json:"height,omitempty"`     // Image height in pixels
	Resolution *string `json:"resolution,omitempty"` // Formatted resolution (e.g., "1920x1080")

	// Extraction metadata
	ExtractionMethod string                 `json:"extraction_method"`      // "ffprobe", "filesystem", etc.
	ExtractionTime   time.Time              `json:"extraction_time"`        // When metadata was extracted
	Success          bool                   `json:"success"`                // Whether extraction succeeded
	Error            string                 `json:"error,omitempty"`        // Error message if extraction failed
	RawMetadata      map[string]interface{} `json:"raw_metadata,omitempty"` // Raw FFprobe output
}

// FFprobeResult represents the JSON structure returned by ffprobe
type FFprobeResult struct {
	Streams []FFprobeStream `json:"streams"`
	Format  FFprobeFormat   `json:"format"`
}

// FFprobeStream represents a single stream in ffprobe output
type FFprobeStream struct {
	Index         int     `json:"index"`
	CodecName     string  `json:"codec_name,omitempty"`
	CodecLongName string  `json:"codec_long_name,omitempty"`
	CodecType     string  `json:"codec_type,omitempty"`
	Width         *int    `json:"width,omitempty"`
	Height        *int    `json:"height,omitempty"`
	BitRate       *string `json:"bit_rate,omitempty"`
	Duration      *string `json:"duration,omitempty"`
	PixFmt        string  `json:"pix_fmt,omitempty"`
}

// FFprobeFormat represents format information in ffprobe output
type FFprobeFormat struct {
	Filename   string  `json:"filename"`
	NbStreams  int     `json:"nb_streams"`
	FormatName string  `json:"format_name,omitempty"`
	Duration   *string `json:"duration,omitempty"`
	Size       *string `json:"size,omitempty"`
	BitRate    *string `json:"bit_rate,omitempty"`
}

// NewMetadataManager creates a new metadata manager with ConfigIntegration
func NewMetadataManager(configIntegration *ConfigIntegration, ffmpegManager FFmpegManager, logger *logging.Logger) *MetadataManager {
	return &MetadataManager{
		configIntegration: configIntegration,
		logger:            logger,
		ffmpegManager:     ffmpegManager,
	}
}

// ExtractVideoMetadata extracts metadata from video files using ffprobe
func (mm *MetadataManager) ExtractVideoMetadata(ctx context.Context, filePath string) (*MediaMetadata, error) {

	// Get file stats first
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file stats: %w", err)
	}

	// Initialize metadata with basic information
	metadata := &MediaMetadata{
		FilePath:         filePath,
		FileSize:         fileInfo.Size(),
		CreatedAt:        fileInfo.ModTime(), // Use ModTime as CreatedAt
		ModifiedAt:       fileInfo.ModTime(),
		ExtractionMethod: "ffprobe",
		ExtractionTime:   time.Now(),
		Success:          false,
	}

	// Execute ffprobe with timeout
	ffprobeResult, err := mm.executeFFprobe(ctx, filePath)
	if err != nil {
		metadata.Error = err.Error()
		mm.logger.WithError(err).WithField("file_path", filePath).Warn("FFprobe extraction failed, returning basic metadata")
		return metadata, nil // Return partial metadata instead of error
	}

	// Parse ffprobe results
	err = mm.parseVideoMetadata(ffprobeResult, metadata)
	if err != nil {
		metadata.Error = err.Error()
		mm.logger.WithError(err).WithField("file_path", filePath).Warn("FFprobe parsing failed, returning basic metadata")
		return metadata, nil // Return partial metadata instead of error
	}

	metadata.Success = true
	mm.logger.WithFields(logging.Fields{
		"file_path": filePath,
		"duration":  metadata.Duration,
		"codec":     metadata.VideoCodec,
		"bitrate":   metadata.Bitrate,
	}).Info("Video metadata extracted successfully")

	return metadata, nil
}

// ExtractImageMetadata extracts metadata from image files using ffprobe
func (mm *MetadataManager) ExtractImageMetadata(ctx context.Context, filePath string) (*MediaMetadata, error) {

	// Get file stats first
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file stats: %w", err)
	}

	// Initialize metadata with basic information
	metadata := &MediaMetadata{
		FilePath:         filePath,
		FileSize:         fileInfo.Size(),
		CreatedAt:        fileInfo.ModTime(),
		ModifiedAt:       fileInfo.ModTime(),
		ExtractionMethod: "ffprobe",
		ExtractionTime:   time.Now(),
		Success:          false,
	}

	// Execute ffprobe with timeout
	ffprobeResult, err := mm.executeFFprobe(ctx, filePath)
	if err != nil {
		metadata.Error = err.Error()
		mm.logger.WithError(err).WithField("file_path", filePath).Warn("FFprobe extraction failed, returning basic metadata")
		return metadata, nil // Return partial metadata instead of error
	}

	// Parse ffprobe results for image
	err = mm.parseImageMetadata(ffprobeResult, metadata)
	if err != nil {
		metadata.Error = err.Error()
		mm.logger.WithError(err).WithField("file_path", filePath).Warn("FFprobe parsing failed, returning basic metadata")
		return metadata, nil // Return partial metadata instead of error
	}

	metadata.Success = true
	mm.logger.WithFields(logging.Fields{
		"file_path":  filePath,
		"width":      metadata.Width,
		"height":     metadata.Height,
		"resolution": metadata.Resolution,
	}).Info("Image metadata extracted successfully")

	return metadata, nil
}

// executeFFprobe executes ffprobe command with timeout and returns parsed JSON
func (mm *MetadataManager) executeFFprobe(ctx context.Context, filePath string) (*FFprobeResult, error) {
	// Get timeout from configuration
	timeout := 10 * time.Second // Default fallback
	if mm.configIntegration != nil {
		if cfg, err := mm.configIntegration.GetConfig(); err == nil && cfg != nil {
			// Use FFmpeg snapshot configuration for timeout
			timeout = time.Duration(cfg.MediaMTX.FFmpeg.Snapshot.ExecutionTimeout * float64(time.Second))
		}
	}

	// Create timeout context
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Build ffprobe command
	command := []string{
		"ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		filePath,
	}

	mm.logger.WithFields(logging.Fields{
		"command": strings.Join(command, " "),
		"timeout": timeout,
	}).Info("Executing ffprobe command")

	// Execute command with timeout
	cmd := exec.CommandContext(timeoutCtx, command[0], command[1:]...)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ffprobe command failed: %w", err)
	}

	// Parse JSON output
	var result FFprobeResult
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse ffprobe JSON output: %w", err)
	}

	mm.logger.WithFields(logging.Fields{
		"file_path":     filePath,
		"streams_count": len(result.Streams),
		"format":        result.Format.FormatName,
	}).Info("FFprobe command executed successfully")

	return &result, nil
}

// parseVideoMetadata parses ffprobe results for video files
func (mm *MetadataManager) parseVideoMetadata(ffprobeResult *FFprobeResult, metadata *MediaMetadata) error {
	// Extract format information
	if ffprobeResult.Format.Duration != nil {
		if duration, err := strconv.ParseFloat(*ffprobeResult.Format.Duration, 64); err == nil {
			metadata.Duration = &duration
		}
	}

	if ffprobeResult.Format.BitRate != nil {
		if bitrate, err := strconv.ParseInt(*ffprobeResult.Format.BitRate, 10, 64); err == nil {
			metadata.Bitrate = &bitrate
		}
	}

	// Extract video stream information
	for _, stream := range ffprobeResult.Streams {
		if stream.CodecType == "video" {
			if stream.CodecName != "" {
				metadata.VideoCodec = &stream.CodecName
			}

			// If format duration is not available, try stream duration
			if metadata.Duration == nil && stream.Duration != nil {
				if duration, err := strconv.ParseFloat(*stream.Duration, 64); err == nil {
					metadata.Duration = &duration
				}
			}

			// If format bitrate is not available, try stream bitrate
			if metadata.Bitrate == nil && stream.BitRate != nil {
				if bitrate, err := strconv.ParseInt(*stream.BitRate, 10, 64); err == nil {
					metadata.Bitrate = &bitrate
				}
			}

			break // Use first video stream
		}
	}

	// Set format from ffprobe
	if ffprobeResult.Format.FormatName != "" {
		metadata.Format = ffprobeResult.Format.FormatName
	}

	// Store raw metadata for debugging
	metadata.RawMetadata = map[string]interface{}{
		"format":  ffprobeResult.Format,
		"streams": ffprobeResult.Streams,
	}

	return nil
}

// parseImageMetadata parses ffprobe results for image files
func (mm *MetadataManager) parseImageMetadata(ffprobeResult *FFprobeResult, metadata *MediaMetadata) error {
	// Extract image stream information
	for _, stream := range ffprobeResult.Streams {
		if stream.CodecType == "video" || stream.Width != nil || stream.Height != nil {
			if stream.Width != nil {
				metadata.Width = stream.Width
			}

			if stream.Height != nil {
				metadata.Height = stream.Height
			}

			// Generate resolution string
			if metadata.Width != nil && metadata.Height != nil {
				resolution := fmt.Sprintf("%dx%d", *metadata.Width, *metadata.Height)
				metadata.Resolution = &resolution
			}

			if stream.CodecName != "" {
				metadata.VideoCodec = &stream.CodecName // For images, this is the image codec
			}

			break // Use first video/image stream
		}
	}

	// Set format from ffprobe
	if ffprobeResult.Format.FormatName != "" {
		metadata.Format = ffprobeResult.Format.FormatName
	}

	// Store raw metadata for debugging
	metadata.RawMetadata = map[string]interface{}{
		"format":  ffprobeResult.Format,
		"streams": ffprobeResult.Streams,
	}

	return nil
}

// GetFileSize returns file size using filesystem stat (utility method)
func (mm *MetadataManager) GetFileSize(filePath string) (int64, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to get file size: %w", err)
	}
	return fileInfo.Size(), nil
}

// GetBasicFileInfo returns basic file information without ffprobe
func (mm *MetadataManager) GetBasicFileInfo(filePath string) (*MediaMetadata, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// Extract format from file extension
	format := "unknown"
	if parts := strings.Split(filePath, "."); len(parts) > 1 {
		format = strings.ToLower(parts[len(parts)-1])
	}

	return &MediaMetadata{
		FilePath:         filePath,
		FileSize:         fileInfo.Size(),
		Format:           format,
		CreatedAt:        fileInfo.ModTime(),
		ModifiedAt:       fileInfo.ModTime(),
		ExtractionMethod: "filesystem",
		ExtractionTime:   time.Now(),
		Success:          true,
	}, nil
}
