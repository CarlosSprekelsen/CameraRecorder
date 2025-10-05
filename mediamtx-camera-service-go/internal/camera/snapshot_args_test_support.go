package camera

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// SnapshotCase represents a snapshot test case
type SnapshotCase struct {
	Name   string
	Device string
	Output string
	Format string
	Width  int
	Height int
}

// MakeStandardCases creates standard test cases using t.TempDir()
func MakeStandardCases(t *testing.T) []SnapshotCase {
	outputDir := t.TempDir()
	return []SnapshotCase{
		{
			Name:   "default_params",
			Device: "/dev/video0",
			Output: filepath.Join(outputDir, "test1.jpg"),
			Format: "mjpeg",
			Width:  640,
			Height: 480,
		},
		{
			Name:   "high_res",
			Device: "/dev/video0",
			Output: filepath.Join(outputDir, "test2.jpg"),
			Format: "mjpeg",
			Width:  1920,
			Height: 1080,
		},
		{
			Name:   "low_res",
			Device: "/dev/video0",
			Output: filepath.Join(outputDir, "test3.jpg"),
			Format: "mjpeg",
			Width:  320,
			Height: 240,
		},
		{
			Name:   "different_format",
			Device: "/dev/video0",
			Output: filepath.Join(outputDir, "test4.jpg"),
			Format: "yuyv",
			Width:  640,
			Height: 480,
		},
	}
}

// MakeEdgeCases creates edge case test cases using t.TempDir()
func MakeEdgeCases(t *testing.T) []SnapshotCase {
	outputDir := t.TempDir()
	return []SnapshotCase{
		{
			Name:   "zero_dimensions",
			Device: "/dev/video0",
			Output: filepath.Join(outputDir, "test5.jpg"),
			Format: "mjpeg",
			Width:  0,
			Height: 0,
		},
		{
			Name:   "negative_dimensions",
			Device: "/dev/video0",
			Output: filepath.Join(outputDir, "test6.jpg"),
			Format: "mjpeg",
			Width:  -1,
			Height: -1,
		},
		{
			Name:   "empty_format",
			Device: "/dev/video0",
			Output: filepath.Join(outputDir, "test7.jpg"),
			Format: "",
			Width:  640,
			Height: 480,
		},
		{
			Name:   "empty_output",
			Device: "/dev/video0",
			Output: "",
			Format: "mjpeg",
			Width:  640,
			Height: 480,
		},
		{
			Name:   "empty_device",
			Device: "",
			Output: filepath.Join(outputDir, "test8.jpg"),
			Format: "mjpeg",
			Width:  640,
			Height: 480,
		},
	}
}

// MakeExtremeCases creates extreme value test cases using t.TempDir()
func MakeExtremeCases(t *testing.T) []SnapshotCase {
	outputDir := t.TempDir()
	return []SnapshotCase{
		{
			Name:   "very_large_dimensions",
			Device: "/dev/video0",
			Output: filepath.Join(outputDir, "test9.jpg"),
			Format: "mjpeg",
			Width:  10000,
			Height: 10000,
		},
		{
			Name:   "very_long_device_path",
			Device: strings.Repeat("/dev/video", 10) + "0", // Simulate a very long device path
			Output: filepath.Join(outputDir, "test10.jpg"),
			Format: "mjpeg",
			Width:  640,
			Height: 480,
		},
		{
			Name:   "very_long_output_path",
			Device: "/dev/video0",
			Output: filepath.Join(outputDir, strings.Repeat("a", 200), "test11.jpg"), // Simulate a very long output path
			Format: "mjpeg",
			Width:  640,
			Height: 480,
		},
	}
}

// AssertSnapshotArgs validates snapshot args with correct assertions
func AssertSnapshotArgs(t *testing.T, monitor *HybridCameraMonitor, c SnapshotCase) {
	args := monitor.buildV4L2SnapshotArgs(c.Device, c.Output, c.Format, c.Width, c.Height)

	// ✅ CORRECT: Test that args are not empty
	assert.NotEmpty(t, args, "Snapshot args should not be empty for %s", c.Name)

	// ✅ CORRECT: Test that args contain expected V4L2 arguments
	assert.Contains(t, args, "--stream-mmap", "Args should contain stream-mmap for %s", c.Name)
	assert.Contains(t, args, "--stream-to", "Args should contain stream-to for %s", c.Name)
	assert.Contains(t, args, "--stream-count", "Args should contain stream-count for %s", c.Name)

	// ✅ CORRECT: Test that args contain output path if provided
	if c.Output != "" {
		assert.Contains(t, args, c.Output, "Args should contain output path for %s", c.Name)
	}

	// ✅ CORRECT: Test resolution handling
	if c.Width > 0 && c.Height > 0 {
		expectedRes := fmt.Sprintf("width=%d,height=%d", c.Width, c.Height)
		assert.Contains(t, args, expectedRes, "Args should contain resolution for %s", c.Name)
	} else {
		assert.NotContains(t, args, "width=", "Args should NOT contain resolution for %s", c.Name)
		assert.NotContains(t, args, "height=", "Args should NOT contain resolution for %s", c.Name)
	}

	// ✅ CORRECT: Test that args do NOT contain device path (it's handled separately)
	// Only check this if device path is not empty
	if c.Device != "" {
		assert.NotContains(t, args, c.Device, "Args should NOT contain device path for %s", c.Name)
	}

	// ✅ CORRECT: Test that args do NOT contain format (it's handled separately)
	if c.Format != "" {
		assert.NotContains(t, args, c.Format, "Args should NOT contain format for %s", c.Name)
	}
}
