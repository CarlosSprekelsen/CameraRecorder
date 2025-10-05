package e2e

import (
	"fmt"
	"os"
	"time"

	"github.com/camerarecorder/mediamtx-camera-service-go/internal/testutils"
	"github.com/stretchr/testify/require"
)

type RecordingSession struct {
	fixture  *E2EFixture
	client   *testutils.WebSocketTestClient
	device   string
	basename string
}

func (f *E2EFixture) StartRecording(device string) (*RecordingSession, error) {
	resp, err := f.client.StartRecording(device)
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
	return &RecordingSession{fixture: f, client: f.client, device: device, basename: bn}, nil
}

func (r *RecordingSession) FilePath() string {
	return r.fixture.RecordingPath(r.device, r.basename)
}

func (r *RecordingSession) Stop() error {
	resp, err := r.client.StopRecording(r.device)
	if err != nil {
		return err
	}
	if resp.Error != nil {
		return fmt.Errorf("rpc error: %d %s", resp.Error.Code, resp.Error.Message)
	}
	return nil
}

func (r *RecordingSession) AssertFileExists(dvh *testutils.DataValidationHelper) {
	path := r.FilePath()
	dvh.WaitForFileCreation(path, testutils.DefaultTestTimeout, "recording")
	dvh.AssertFileExists(path, testutils.UniversalMinRecordingFileSize, "recording")
}

func (r *RecordingSession) AssertGrowing() {
	path := r.FilePath()

	info1, err := os.Stat(path)
	require.NoError(r.fixture.t, err, "File must exist before checking growth")

	time.Sleep(testutils.UniversalTimeoutShort)

	info2, err := os.Stat(path)
	require.NoError(r.fixture.t, err, "File must exist after delay")

	require.Greater(r.fixture.t, info2.Size(), info1.Size(), "File should grow during recording")
}
