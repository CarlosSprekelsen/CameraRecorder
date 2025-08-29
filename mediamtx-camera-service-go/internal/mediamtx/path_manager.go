/*
MediaMTX Path Manager Implementation

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
	"time"

	"github.com/sirupsen/logrus"
)

// pathManager represents the MediaMTX path manager
type pathManager struct {
	client MediaMTXClient
	config *MediaMTXConfig
	logger *logrus.Logger
}

// NewPathManager creates a new MediaMTX path manager
func NewPathManager(client MediaMTXClient, config *MediaMTXConfig, logger *logrus.Logger) PathManager {
	return &pathManager{
		client: client,
		config: config,
		logger: logger,
	}
}

// CreatePath creates a new path
func (pm *pathManager) CreatePath(ctx context.Context, name, source string, options map[string]interface{}) error {
	pm.logger.WithFields(logrus.Fields{
		"name":    name,
		"source":  source,
		"options": options,
	}).Debug("Creating MediaMTX path")

	// Create path request
	path := &Path{
		Name:   name,
		Source: source,
	}

	// Apply options to path
	if sourceOnDemand, ok := options["sourceOnDemand"].(bool); ok {
		path.SourceOnDemand = sourceOnDemand
	}
	if startTimeout, ok := options["sourceOnDemandStartTimeout"].(string); ok {
		if duration, err := parseDuration(startTimeout); err == nil {
			path.SourceOnDemandStartTimeout = duration
		}
	}
	if closeAfter, ok := options["sourceOnDemandCloseAfter"].(string); ok {
		if duration, err := parseDuration(closeAfter); err == nil {
			path.SourceOnDemandCloseAfter = duration
		}
	}
	if publishUser, ok := options["publishUser"].(string); ok {
		path.PublishUser = publishUser
	}
	if publishPass, ok := options["publishPass"].(string); ok {
		path.PublishPass = publishPass
	}
	if readUser, ok := options["readUser"].(string); ok {
		path.ReadUser = readUser
	}
	if readPass, ok := options["readPass"].(string); ok {
		path.ReadPass = readPass
	}
	if runOnDemand, ok := options["runOnDemand"].(string); ok {
		path.RunOnDemand = runOnDemand
	}
	if runOnDemandRestart, ok := options["runOnDemandRestart"].(bool); ok {
		path.RunOnDemandRestart = runOnDemandRestart
	}
	if runOnDemandCloseAfter, ok := options["runOnDemandCloseAfter"].(string); ok {
		if duration, err := parseDuration(runOnDemandCloseAfter); err == nil {
			path.RunOnDemandCloseAfter = duration
		}
	}
	if runOnDemandStartTimeout, ok := options["runOnDemandStartTimeout"].(string); ok {
		if duration, err := parseDuration(runOnDemandStartTimeout); err == nil {
			path.RunOnDemandStartTimeout = duration
		}
	}

	// Marshal request
	data, err := marshalCreatePathRequest(path)
	if err != nil {
		return NewPathErrorWithErr(name, "create_path", "failed to marshal request", err)
	}

	// Send request
	_, err = pm.client.Post(ctx, "/v3/paths/add", data)
	if err != nil {
		return NewPathErrorWithErr(name, "create_path", "failed to create path", err)
	}

	pm.logger.WithField("name", name).Info("MediaMTX path created successfully")
	return nil
}

// DeletePath deletes a path
func (pm *pathManager) DeletePath(ctx context.Context, name string) error {
	pm.logger.WithField("name", name).Debug("Deleting MediaMTX path")

	err := pm.client.Delete(ctx, fmt.Sprintf("/v3/paths/delete/%s", name))
	if err != nil {
		return NewPathErrorWithErr(name, "delete_path", "failed to delete path", err)
	}

	pm.logger.WithField("name", name).Info("MediaMTX path deleted successfully")
	return nil
}

// GetPath gets a specific path
func (pm *pathManager) GetPath(ctx context.Context, name string) (*Path, error) {
	pm.logger.WithField("name", name).Debug("Getting MediaMTX path")

	data, err := pm.client.Get(ctx, fmt.Sprintf("/v3/paths/get/%s", name))
	if err != nil {
		return nil, NewPathErrorWithErr(name, "get_path", "failed to get path", err)
	}

	path, err := parsePathResponse(data)
	if err != nil {
		return nil, NewPathErrorWithErr(name, "get_path", "failed to parse path response", err)
	}

	return path, nil
}

// ListPaths lists all paths
func (pm *pathManager) ListPaths(ctx context.Context) ([]*Path, error) {
	pm.logger.Debug("Listing MediaMTX paths")

	data, err := pm.client.Get(ctx, "/v3/paths/list")
	if err != nil {
		return nil, fmt.Errorf("failed to list paths: %w", err)
	}

	paths, err := parsePathsResponse(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse paths response: %w", err)
	}

	pm.logger.WithField("count", len(paths)).Debug("MediaMTX paths listed successfully")
	return paths, nil
}

// ValidatePath validates a path
func (pm *pathManager) ValidatePath(ctx context.Context, name string) error {
	pm.logger.WithField("name", name).Debug("Validating MediaMTX path")

	// Check if path exists
	exists := pm.PathExists(ctx, name)
	if !exists {
		return NewPathError(name, "validate_path", "path does not exist")
	}

	// Get path details to validate configuration
	_, err := pm.GetPath(ctx, name)
	if err != nil {
		return NewPathErrorWithErr(name, "validate_path", "failed to get path details", err)
	}

	pm.logger.WithField("name", name).Debug("MediaMTX path validated successfully")
	return nil
}

// PathExists checks if a path exists
func (pm *pathManager) PathExists(ctx context.Context, name string) bool {
	pm.logger.WithField("name", name).Debug("Checking if MediaMTX path exists")

	_, err := pm.GetPath(ctx, name)
	return err == nil
}

// parseDuration parses a duration string
func parseDuration(durationStr string) (time.Duration, error) {
	return time.ParseDuration(durationStr)
}
