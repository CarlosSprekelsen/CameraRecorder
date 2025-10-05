package e2e

import (
	"fmt"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/testutils"
)

type SnapshotCapture struct {
	fixture  *E2EFixture
	device   string
	basename string
}

func (f *E2EFixture) TakeSnapshot(device string) (*SnapshotCapture, error) {
	resp, err := f.client.TakeSnapshot(device)
	if err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("rpc error: %d %s", resp.Error.Code, resp.Error.Message)
	}
	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid result type")
	}
	bn, ok := result["filename"].(string)
	if !ok || bn == "" {
		return nil, fmt.Errorf("missing filename in result")
	}
	return &SnapshotCapture{fixture: f, device: device, basename: bn}, nil
}

func (s *SnapshotCapture) FilePath() string {
	return s.fixture.SnapshotPath(s.device, s.basename)
}

func (s *SnapshotCapture) AssertImageValid(dvh *testutils.DataValidationHelper) {
	path := s.FilePath()
	dvh.WaitForFileCreation(path, testutils.DefaultTestTimeout, "snapshot")
	dvh.AssertFileExists(path, 1024, "snapshot")
}
