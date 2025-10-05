package testutils

import (
	"testing"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/security"
	"github.com/stretchr/testify/require"
)

// SecurityHelper provides JWT utilities for tests
type SecurityHelper struct {
	setup      *UniversalTestSetup
	jwtHandler *security.JWTHandler
}

// NewSecurityHelper creates helper with real JWT handler
// Takes testing.T to fail test immediately if JWT handler creation fails
func NewSecurityHelper(t *testing.T, setup *UniversalTestSetup) *SecurityHelper {
	config := setup.GetConfigManager().GetConfig()
	logger := setup.GetLogger()
	
	jwtHandler, err := security.NewJWTHandler(config.Security.JWTSecretKey, logger)
	require.NoError(t, err, "Failed to create JWT handler - check config fixture")
	
	return &SecurityHelper{
		setup:      setup,
		jwtHandler: jwtHandler,
	}
}

// GenerateTestToken creates JWT with specified role
func (s *SecurityHelper) GenerateTestToken(userID, role string, duration time.Duration) (string, error) {
	hours := int(duration.Hours())
	return s.jwtHandler.GenerateToken(userID, role, hours)
}

// Convenience methods using UniversalTestUserRole constant
func (s *SecurityHelper) GenerateAdminToken() (string, error) {
	return s.GenerateTestToken(UniversalTestUserID, UniversalTestUserRole, 24*time.Hour)
}

func (s *SecurityHelper) GenerateOperatorToken() (string, error) {
	return s.GenerateTestToken(UniversalTestUserID, "operator", 24*time.Hour)
}

func (s *SecurityHelper) GenerateViewerToken() (string, error) {
	return s.GenerateTestToken(UniversalTestUserID, "viewer", 24*time.Hour)
}
