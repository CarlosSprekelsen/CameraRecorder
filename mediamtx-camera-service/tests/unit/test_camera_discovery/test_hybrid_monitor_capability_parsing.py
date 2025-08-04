"""
Capability parsing and detection tests for hybrid camera monitor.

Test coverage:
- Frame rate extraction from various v4l2-ctl output formats
- Malformed output handling and error recovery
- Capability confirmation logic (provisional → confirmed)
- Frequency-weighted capability merging
- Timeout and subprocess failure handling

Created: 2025-08-04
Related: S3 Camera Discovery hardening, docs/roadmap.md
Evidence: src/camera_discovery/hybrid_monitor.py lines 500-700 (capability detection)
"""

import asyncio
import pytest
from unittest.mock import Mock, patch

# Test imports
from src.camera_discovery.hybrid_monitor import (
    HybridCameraMonitor,
    CapabilityDetectionResult,
    DeviceCapabilityState,
)


class TestCapabilityParsingVariations:
    """Test capability detection parsing with varied and malformed v4l2-ctl outputs."""

    @pytest.fixture
    def monitor(self):
        """Create monitor with capability detection enabled."""
        return HybridCameraMonitor(
            device_range=[0, 1, 2],
            enable_capability_detection=True,
            detection_timeout=2.0,
        )

    def test_frame_rate_extraction_comprehensive_patterns(self, monitor):
        """Test frame rate extraction from comprehensive v4l2-ctl output patterns."""
        test_cases = [
            # Standard patterns - Evidence: hybrid_monitor.py lines 520-540
            ("30.000 fps", {"30"}),
            ("25.000 FPS", {"25"}),
            ("Frame rate: 60.0", {"60"}),
            ("1920x1080@30", {"30"}),
            ("15 Hz", {"15"}),
            # Interval patterns - Evidence: hybrid_monitor.py lines 545-560
            ("Interval: [1/30]", {"30"}),
            ("[1/25]", {"25"}),
            ("1/15 s", {"15"}),
            # Complex patterns - Evidence: hybrid_monitor.py lines 565-580
            ("30 frames per second", {"30"}),
            ("rate: 25.5", {"25.5"}),
            ("fps: 60", {"60"}),
            # Multiple rates in one output - Real v4l2-ctl scenario
            ("30.000 fps, 25 FPS, [1/15], 60 Hz", {"30", "25", "15", "60"}),
            # Edge cases and bounds testing
            ("", set()),
            ("no frame rates here", set()),
            ("300 fps", {"300"}),  # High rate but valid
            ("1.5 fps", {"1.5"}),  # Low rate but valid
            ("0 fps", set()),  # Invalid rate (filtered out)
            ("500 fps", set()),  # Invalid rate (filtered out - over max)
            # Malformed patterns - Error recovery testing
            ("30.000.000 fps", set()),  # Double decimal
            ("abc fps", set()),  # Non-numeric
            ("fps without number", set()),  # No number
            ("30.000 f ps", set()),  # Broken pattern
            ("-30 fps", set()),  # Negative rate
        ]

        for output, expected in test_cases:
            result = monitor._extract_frame_rates_from_output(output)
            assert (
                result == expected
            ), f"Failed for output: '{output}' - expected {expected}, got {result}"

    @pytest.mark.asyncio
    async def test_capability_parsing_malformed_v4l2_outputs(self, monitor):
        """Test capability detection resilience against malformed v4l2-ctl outputs."""

        malformed_outputs = [
            # Truncated output
            '{"incomplete": "json"',
            # Non-JSON output when JSON expected
            "Error: device busy\nRetry later",
            # Empty output
            "",
            # Binary garbage
            b"\x00\x01\x02\x03invalid".decode("utf-8", errors="ignore"),
            # Partial v4l2-ctl output
            "Driver Info:\n  Driver name   : uvcvideo\n  Card type     :",  # Cut off
            # Mixed encoding issues
            "Camera: Café Français\nResolution: 1920×1080",  # Unicode chars
        ]

        device_path = "/dev/video0"

        for malformed_output in malformed_outputs:
            # Mock subprocess to return malformed output
            mock_process = Mock()
            mock_process.stdout = malformed_output
            mock_process.stderr = ""
            mock_process.returncode = 0

            with patch("subprocess.run", return_value=mock_process):
                result = await monitor._probe_device_capabilities(device_path)

            # Should handle malformation gracefully
            assert isinstance(result, CapabilityDetectionResult)
            assert result.detected == (
                len(result.formats) > 0 or len(result.resolutions) > 0
            )
            assert result.error is None or "parsing" in result.error.lower()

    @pytest.mark.asyncio
    async def test_capability_timeout_handling(self, monitor):
        """Test timeout handling during capability detection."""

        device_path = "/dev/video0"

        # Mock subprocess that times out
        async def mock_timeout_subprocess(*args, **kwargs):
            await asyncio.sleep(monitor._detection_timeout + 1.0)  # Exceed timeout
            raise asyncio.TimeoutError("Command timed out")

        with patch(
            "asyncio.create_subprocess_exec", side_effect=mock_timeout_subprocess
        ):
            result = await monitor._probe_device_capabilities(device_path)

        # Should handle timeout gracefully
        assert isinstance(result, CapabilityDetectionResult)
        assert result.detected is False
        assert result.accessible is False
        assert "timeout" in result.error.lower()
        assert result.timeout_context is not None

    @pytest.mark.asyncio
    async def test_subprocess_failure_error_handling(self, monitor):
        """Test subprocess failure handling during capability detection."""

        device_path = "/dev/video0"

        # Mock subprocess that fails with non-zero exit code
        mock_process = Mock()
        mock_process.stdout = ""
        mock_process.stderr = (
            "v4l2-ctl: failed to open /dev/video0: Device or resource busy"
        )
        mock_process.returncode = 1

        with patch("subprocess.run", return_value=mock_process):
            result = await monitor._probe_device_capabilities(device_path)

        # Should handle subprocess failure gracefully
        assert isinstance(result, CapabilityDetectionResult)
        assert result.detected is False
        assert result.accessible is False
        assert "busy" in result.error.lower()

    def test_resolution_extraction_patterns(self, monitor):
        """Test resolution extraction from various v4l2-ctl output formats."""

        test_cases = [
            # Standard resolution patterns
            ("Size: Discrete 1920x1080", ["1920x1080"]),
            ("Size: Discrete 1280x720", ["1280x720"]),
            ("Size: Discrete 640x480", ["640x480"]),
            # Multiple resolutions
            (
                "Size: Discrete 1920x1080\nSize: Discrete 1280x720\nSize: Discrete 640x480",
                ["1920x1080", "1280x720", "640x480"],
            ),
            # Alternative format patterns
            ("Resolution: 1920×1080", ["1920x1080"]),  # Unicode multiplication
            ("  1280 x 720  ", ["1280x720"]),  # Spaces
            ("1024*768", ["1024x768"]),  # Asterisk separator
            # Edge cases
            ("", []),
            ("No resolutions found", []),
            ("Size: Continuous", []),  # Continuous mode (not discrete)
            # Malformed resolutions
            ("Size: Discrete 1920", []),  # Incomplete
            ("Size: Discrete x1080", []),  # Missing width
            ("Size: Discrete abcxdef", []),  # Non-numeric
        ]

        for output, expected in test_cases:
            result = monitor._extract_resolutions_from_output(output)
            # Convert to list for easier comparison (order may vary)
            result_list = (
                sorted(list(result)) if isinstance(result, set) else sorted(result)
            )
            expected_list = sorted(expected)
            assert (
                result_list == expected_list
            ), f"Failed for output: '{output}' - expected {expected_list}, got {result_list}"

    def test_format_extraction_patterns(self, monitor):
        """Test pixel format extraction from v4l2-ctl output."""

        test_cases = [
            # Standard format patterns
            ("Pixel Format: 'YUYV'", [{"format": "YUYV", "description": "YUYV 4:2:2"}]),
            (
                "Pixel Format: 'MJPG'",
                [{"format": "MJPG", "description": "Motion-JPEG"}],
            ),
            # Multiple formats
            (
                "Pixel Format: 'YUYV'\nPixel Format: 'MJPG'",
                [
                    {"format": "YUYV", "description": "YUYV 4:2:2"},
                    {"format": "MJPG", "description": "Motion-JPEG"},
                ],
            ),
            # Format with description
            (
                "Pixel Format: 'YUYV' (YUYV 4:2:2)",
                [{"format": "YUYV", "description": "YUYV 4:2:2"}],
            ),
            # Edge cases
            ("", []),
            ("No pixel formats", []),
            # Malformed formats
            ("Pixel Format: ''", []),  # Empty format
            ("Pixel Format: YUYV", []),  # Missing quotes
            ("Format: 'UNKNOWN'", []),  # Different pattern
        ]

        for output, expected in test_cases:
            result = monitor._extract_formats_from_output(output)
            assert (
                result == expected
            ), f"Failed for output: '{output}' - expected {expected}, got {result}"


class TestCapabilityConfirmationLogic:
    """Test provisional → confirmed capability confirmation logic."""

    @pytest.fixture
    def monitor_with_confirmation(self):
        """Create monitor with capability confirmation enabled."""
        return HybridCameraMonitor(
            device_range=[0, 1], enable_capability_detection=True, detection_timeout=1.0
        )

    def test_capability_state_initialization(self, monitor_with_confirmation):
        """Test initial capability state setup for new devices."""

        device_path = "/dev/video0"

        # Get or create capability state
        state = monitor_with_confirmation._get_or_create_capability_state(device_path)

        assert isinstance(state, DeviceCapabilityState)
        assert state.device_path == device_path
        assert state.provisional_data is None
        assert state.confirmed_data is None
        assert state.consecutive_successes == 0
        assert state.consecutive_failures == 0
        assert state.confirmation_threshold >= 2  # Architecture requirement
        assert not state.is_confirmed()

    def test_provisional_capability_recording(self, monitor_with_confirmation):
        """Test recording of provisional capability data."""

        device_path = "/dev/video0"

        # Create mock capability result
        capability_result = CapabilityDetectionResult(
            device_path=device_path,
            detected=True,
            accessible=True,
            device_name="Test Camera",
            formats=[{"format": "YUYV", "description": "YUYV 4:2:2"}],
            resolutions=["1920x1080", "1280x720"],
            frame_rates=["30", "25"],
        )

        # Record provisional data
        monitor_with_confirmation._update_capability_state(
            device_path, capability_result
        )

        state = monitor_with_confirmation._capability_states[device_path]
        assert state.provisional_data is not None
        assert state.provisional_data == capability_result
        assert state.confirmed_data is None  # Not confirmed yet
        assert state.consecutive_successes == 1

    def test_capability_confirmation_threshold(self, monitor_with_confirmation):
        """Test capability confirmation after reaching threshold."""

        device_path = "/dev/video0"

        # Create consistent capability result
        capability_result = CapabilityDetectionResult(
            device_path=device_path,
            detected=True,
            accessible=True,
            device_name="Test Camera",
            formats=[{"format": "YUYV", "description": "YUYV 4:2:2"}],
            resolutions=["1920x1080"],
            frame_rates=["30"],
        )

        # Record multiple consistent results to reach confirmation threshold
        confirmation_threshold = 3
        for i in range(confirmation_threshold):
            monitor_with_confirmation._update_capability_state(
                device_path, capability_result
            )

        state = monitor_with_confirmation._capability_states[device_path]
        assert state.consecutive_successes == confirmation_threshold
        assert state.confirmed_data is not None
        assert state.is_confirmed()

        # Effective capability should return confirmed data
        effective = state.get_effective_capability()
        assert effective == state.confirmed_data

    def test_capability_frequency_weighted_merging(self, monitor_with_confirmation):
        """Test frequency-weighted capability merging logic."""

        device_path = "/dev/video0"
        state = monitor_with_confirmation._get_or_create_capability_state(device_path)

        # Simulate multiple detections with different frame rates
        frame_rate_detections = [
            "30",
            "30",
            "25",
            "30",
            "60",
            "30",
        ]  # 30 appears 4 times

        for frame_rate in frame_rate_detections:
            capability_result = CapabilityDetectionResult(
                device_path=device_path,
                detected=True,
                accessible=True,
                frame_rates=[frame_rate],
            )

            # Update frequency tracking
            monitor_with_confirmation._update_frequency_tracking(
                state, capability_result
            )

        # Verify frequency tracking
        assert state.frame_rate_frequency["30"] == 4  # Most frequent
        assert state.frame_rate_frequency["25"] == 1
        assert state.frame_rate_frequency["60"] == 1

        # Most frequent frame rate should be prioritized
        most_frequent = max(
            state.frame_rate_frequency.keys(),
            key=lambda x: state.frame_rate_frequency[x],
        )
        assert most_frequent == "30"

    def test_capability_inconsistency_handling(self, monitor_with_confirmation):
        """Test handling of inconsistent capability detections."""

        device_path = "/dev/video0"

        # First detection
        result1 = CapabilityDetectionResult(
            device_path=device_path,
            detected=True,
            accessible=True,
            resolutions=["1920x1080"],
            frame_rates=["30"],
        )

        # Inconsistent second detection
        result2 = CapabilityDetectionResult(
            device_path=device_path,
            detected=True,
            accessible=True,
            resolutions=["1280x720"],  # Different resolution
            frame_rates=["25"],  # Different frame rate
        )

        # Record both results
        monitor_with_confirmation._update_capability_state(device_path, result1)
        monitor_with_confirmation._update_capability_state(device_path, result2)

        state = monitor_with_confirmation._capability_states[device_path]

        # Should track both possibilities in frequency tables
        assert "1920x1080" in state.resolution_frequency
        assert "1280x720" in state.resolution_frequency
        assert "30" in state.frame_rate_frequency
        assert "25" in state.frame_rate_frequency

        # Confirmation should require consistency
        assert state.consecutive_successes < state.confirmation_threshold

    @pytest.mark.asyncio
    async def test_capability_detection_failure_recovery(
        self, monitor_with_confirmation
    ):
        """Test recovery from capability detection failures."""

        device_path = "/dev/video0"

        # Simulate detection failure
        failure_result = CapabilityDetectionResult(
            device_path=device_path,
            detected=False,
            accessible=False,
            error="Device busy",
        )

        monitor_with_confirmation._update_capability_state(device_path, failure_result)

        state = monitor_with_confirmation._capability_states[device_path]
        assert state.consecutive_failures == 1
        assert state.consecutive_successes == 0

        # Simulate successful detection after failure
        success_result = CapabilityDetectionResult(
            device_path=device_path,
            detected=True,
            accessible=True,
            resolutions=["1920x1080"],
            frame_rates=["30"],
        )

        monitor_with_confirmation._update_capability_state(device_path, success_result)

        # Should reset failure count and start success count
        assert state.consecutive_failures == 0
        assert state.consecutive_successes == 1
        assert state.provisional_data is not None


class TestCapabilityIntegration:
    """Test integration of capability detection with monitor lifecycle."""

    @pytest.fixture
    def monitor_integrated(self):
        """Create monitor for integration testing."""
        return HybridCameraMonitor(
            device_range=[0, 1], enable_capability_detection=True, detection_timeout=1.0
        )

    @pytest.mark.asyncio
    async def test_get_effective_capability_metadata_integration(
        self, monitor_integrated
    ):
        """Test the get_effective_capability_metadata method integration."""

        device_path = "/dev/video0"

        # Setup confirmed capability data
        confirmed_capability = CapabilityDetectionResult(
            device_path=device_path,
            detected=True,
            accessible=True,
            device_name="Confirmed Camera",
            resolutions=["1920x1080", "1280x720"],
            frame_rates=["30", "25"],
        )

        state = monitor_integrated._get_or_create_capability_state(device_path)
        state.confirmed_data = confirmed_capability
        state.consecutive_successes = 5

        # Get effective metadata
        metadata = monitor_integrated.get_effective_capability_metadata(device_path)

        # Should return metadata based on confirmed capability
        assert metadata is not None
        assert metadata["validation_status"] == "confirmed"
        assert metadata["consecutive_successes"] == 5
        assert metadata["resolution"] in ["1920x1080", "1280x720"]  # Highest priority
        assert metadata["fps"] in [30, 25]  # Highest priority frame rate
        assert "formats" in metadata
        assert "all_resolutions" in metadata

    @pytest.mark.asyncio
    async def test_provisional_metadata_fallback(self, monitor_integrated):
        """Test metadata fallback to provisional when confirmed unavailable."""

        device_path = "/dev/video0"

        # Setup only provisional capability data
        provisional_capability = CapabilityDetectionResult(
            device_path=device_path,
            detected=True,
            accessible=True,
            device_name="Provisional Camera",
            resolutions=["1280x720"],
            frame_rates=["30"],
        )

        state = monitor_integrated._get_or_create_capability_state(device_path)
        state.provisional_data = provisional_capability
        state.consecutive_successes = 1  # Below confirmation threshold

        # Get effective metadata
        metadata = monitor_integrated.get_effective_capability_metadata(device_path)

        # Should return metadata based on provisional capability
        assert metadata is not None
        assert metadata["validation_status"] == "provisional"
        assert metadata["consecutive_successes"] == 1
        assert metadata["resolution"] == "1280x720"
        assert metadata["fps"] == 30

    @pytest.mark.asyncio
    async def test_no_capability_data_fallback(self, monitor_integrated):
        """Test metadata fallback when no capability data available."""

        device_path = "/dev/video999"  # Non-existent device

        # Get effective metadata for device with no capability data
        metadata = monitor_integrated.get_effective_capability_metadata(device_path)

        # Should return None when no capability data available
        assert metadata is None
