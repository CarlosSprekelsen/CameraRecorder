package testutils

import (
	"context"
	"testing"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/common"
)

// ServiceLifecycle provides common lifecycle patterns using common.Stoppable
type ServiceLifecycle struct {
	setup *UniversalTestSetup
}

// NewServiceLifecycle creates lifecycle helper
func NewServiceLifecycle(setup *UniversalTestSetup) *ServiceLifecycle {
	return &ServiceLifecycle{setup: setup}
}

// StartServiceWithCleanup starts service and registers cleanup
func (s *ServiceLifecycle) StartServiceWithCleanup(
	t *testing.T,
	service common.Stoppable,
	startFunc func(context.Context) error,
) error {
	ctx, cancel := s.setup.GetStandardContextWithTimeout(UniversalTimeoutLong)
	defer cancel()
	
	if err := startFunc(ctx); err != nil {
		return err
	}
	
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := s.setup.GetStandardContextWithTimeout(UniversalTimeoutShort)
		defer cleanupCancel()
		
		if err := service.Stop(cleanupCtx); err != nil {
			t.Logf("Warning: service cleanup failed: %v", err)
		}
	})
	
	return nil
}

// WaitForServiceReady uses WaitForCondition for readiness
func (s *ServiceLifecycle) WaitForServiceReady(
	ctx context.Context,
	checkFunc func() bool,
	description string,
) error {
	return WaitForCondition(ctx, checkFunc, UniversalTimeoutLong, description)
}
