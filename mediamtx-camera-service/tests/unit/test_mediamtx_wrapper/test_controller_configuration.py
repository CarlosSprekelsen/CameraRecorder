# tests/unit/test_mediamtx_wrapper/test_controller_configuration.py
"""
Test configuration validation, error accumulation, and safe fallback behavior.

Test policy: Verify input validation, error accumulation without system crashes,
and graceful degradation when invalid configurations are provided.
"""

import pytest
import asyncio
from unittest.mock import Mock, AsyncMock, patch
import aiohttp

from src.mediamtx_wrapper.controller import MediaMTXController


class TestConfigurationValidation:
    """Test configuration validation and error handling."""

    @pytest.fixture
    def controller(self):
        """Create MediaMTX controller with test configuration."""
        controller = MediaMTXController(
            host="localhost",
            api_port=9997,
            rtsp_port=8554,
            webrtc_port=8889,
            hls_port=8888,
            config_path="/tmp/test_config.yml",
            recordings_path="/tmp/recordings",
            snapshots_path="/tmp/snapshots",
        )
        
        # Simple mock session that raises validation errors before HTTP calls
        controller._session = Mock()
        return controller

    def _mock_response(self, status, json_data=None, text_data=""):
        """Helper to create mock HTTP response."""
        response = Mock()
        response.status = status
        response.json = AsyncMock(return_value=json_data or {})
        response.text = AsyncMock(return_value=text_data)
        return response

    @pytest.mark.asyncio
    async def test_configuration_validation_unknown_keys(self, controller):
        """Test validation rejects unknown configuration keys."""
        invalid_config = {
            "logLevel": "info",  # Valid
            "unknownKey": "value",  # Invalid
            "anotherUnknown": 123,  # Invalid
        }

        with pytest.raises(ValueError) as exc_info:
            await controller.update_configuration(invalid_config)

        # Verify error message lists unknown keys
        error_msg = str(exc_info.value)
        assert "Unknown configuration keys" in error_msg
        assert "unknownKey" in error_msg
        assert "anotherUnknown" in error_msg

    @pytest.mark.asyncio
    async def test_configuration_validation_type_errors(self, controller):
        """Test validation catches type mismatches."""
        invalid_configs = [
            {"logLevel": 123},  # Should be string
            {"api": "true"},  # Should be boolean
            {"readBufferCount": "invalid"},  # Should be int
        ]

        for invalid_config in invalid_configs:
            with pytest.raises(ValueError, match="Invalid type"):
                await controller.update_configuration(invalid_config)

    @pytest.mark.asyncio
    async def test_configuration_validation_value_constraints(self, controller):
        """Test validation enforces value constraints (min/max, allowed values)."""
        invalid_configs = [
            {"logLevel": "invalid_level"},  # Not in allowed values
            {"readTimeout": 0},  # Below minimum
            {"readTimeout": 500},  # Above maximum
            {"readBufferCount": 0},  # Below minimum
            {"readBufferCount": 10000},  # Above maximum
            {"udpMaxPayloadSize": 500},  # Below minimum
        ]

        for invalid_config in invalid_configs:
            with pytest.raises(ValueError) as exc_info:
                await controller.update_configuration(invalid_config)

            # Verify specific constraint violation is mentioned
            error_msg = str(exc_info.value)
            assert any(
                keyword in error_msg
                for keyword in [
                    "Invalid value",
                    "too small",
                    "too large",
                    "allowed values",
                ]
            )

    @pytest.mark.asyncio
    async def test_configuration_validation_pattern_matching(self, controller):
        """Test validation enforces string patterns (e.g., IP addresses)."""
        invalid_patterns = [
            {"apiAddress": "invalid_ip"},
            {"metricsAddress": "not.an.ip.address"},
            {"rtspAddress": "999.999.999.999"},  # Invalid IP
            {"webrtcAddress": "localhost:abc"},  # Invalid port format
        ]

        for invalid_config in invalid_patterns:
            with pytest.raises(ValueError, match="Invalid format"):
                await controller.update_configuration(invalid_config)

    @pytest.mark.asyncio
    async def test_configuration_validation_error_accumulation(self, controller):
        """Test validation accumulates multiple errors without crashing."""
        # Configuration with multiple validation errors
        invalid_config = {
            "logLevel": "invalid_level",  # Bad value
            "api": "not_boolean",  # Bad type
            "readTimeout": -1,  # Below minimum
            "unknownKey": "value",  # Unknown key
        }

        with pytest.raises(ValueError) as exc_info:
            await controller.update_configuration(invalid_config)

        # Verify multiple errors are reported
        error_msg = str(exc_info.value)
        # Should mention unknown keys error
        assert "Unknown configuration keys" in error_msg

        # Try validation-only errors (excluding unknown keys)
        validation_only_config = {
            "logLevel": "invalid_level",
            "api": "not_boolean",
            "readTimeout": -1,
        }

        with pytest.raises(ValueError) as exc_info:
            await controller.update_configuration(validation_only_config)

        error_msg = str(exc_info.value)
        # Should accumulate multiple validation errors
        error_count = error_msg.count(";")  # Errors separated by semicolons
        assert error_count >= 2  # At least 2 validation errors

    @pytest.mark.asyncio
    async def test_configuration_update_api_failure_safe_fallback(self, controller):
        """Test safe fallback behavior when MediaMTX API fails during update."""
        # Valid configuration
        valid_config = {"logLevel": "debug", "api": True}

        # Mock API failure
        error_response = self._mock_response(500, text_data="Internal Server Error")
        controller._session.post = AsyncMock(return_value=error_response)

        # Should raise ValueError (not crash system)
        with pytest.raises(ValueError, match="Failed to update configuration"):
            await controller.update_configuration(valid_config)

        # Controller should remain in usable state
        assert controller._session is not None

    @pytest.mark.asyncio
    async def test_configuration_update_network_error_handling(self, controller):
        """Test network error handling during configuration update."""
        valid_config = {"logLevel": "info"}

        # Mock network error
        controller._session.post = AsyncMock(
            side_effect=aiohttp.ClientError("Connection refused")
        )

        with pytest.raises(ConnectionError, match="MediaMTX unreachable"):
            await controller.update_configuration(valid_config)

    @pytest.mark.asyncio
    async def test_configuration_validation_empty_updates(self, controller):
        """Test validation handles empty or None configuration updates."""
        # Empty config
        with pytest.raises(ValueError, match="Configuration updates are required"):
            await controller.update_configuration({})

        # None config
        with pytest.raises(ValueError, match="Configuration updates are required"):
            await controller.update_configuration(None)

    @pytest.mark.asyncio
    async def test_configuration_validation_valid_config_success(self, controller):
        """Test successful configuration update with valid parameters."""
        valid_config = {
            "logLevel": "debug",
            "api": True,
            "readTimeout": 30,
            "readBufferCount": 512,
            "apiAddress": "127.0.0.1:9997",
        }

        # Mock successful API response
        success_response = self._mock_response(200)
        controller._session.post = AsyncMock(return_value=success_response)

        result = await controller.update_configuration(valid_config)

        assert result is True
        # Verify API was called with correct config
        controller._session.post.assert_called_once()
        call_args = controller._session.post.call_args
        assert call_args.kwargs["json"] == valid_config

    @pytest.mark.asyncio
    async def test_configuration_update_without_session(self, controller):
        """Test configuration update fails gracefully when controller not started."""
        controller._session = None

        with pytest.raises(ConnectionError, match="MediaMTX controller not started"):
            await controller.update_configuration({"logLevel": "info"})

    def test_configuration_validation_schema_completeness(self, controller):
        """Test validation schema covers all expected MediaMTX configuration options."""
        # Test that validation schema includes key MediaMTX settings
        valid_config_keys = [
            "logLevel",
            "logDestinations",
            "readTimeout",
            "writeTimeout",
            "readBufferCount",
            "udpMaxPayloadSize",
            "runOnConnect",
            "runOnConnectRestart",
            "api",
            "apiAddress",
            "metrics",
            "metricsAddress",
            "pprof",
            "pprofAddress",
            "rtsp",
            "rtspAddress",
            "rtspsAddress",
            "rtmp",
            "rtmpAddress",
            "rtmps",
            "rtmpsAddress",
            "hls",
            "hlsAddress",
            "hlsAllowOrigin",
            "webrtc",
            "webrtcAddress",
        ]

        # Each should be either accepted or rejected with specific error
        for key in valid_config_keys:
            test_config = {
                key: "test_value"
            }  # May be wrong type, but key should be recognized

            try:
                # This will likely fail due to wrong type/value, but should not fail due
                # to unknown key
                asyncio.run(controller.update_configuration(test_config))
            except ValueError as e:
                # Should be type/value error, not unknown key error
                assert "Unknown configuration keys" not in str(e)
            except ConnectionError:
                # Expected if no session
                pass

    @pytest.mark.asyncio
    async def test_configuration_validation_correlation_id_logging(self, controller):
        """Test configuration validation includes correlation IDs in logging."""
        correlation_ids = []

        def mock_set_correlation_id(cid):
            correlation_ids.append(cid)

        with patch(
            "src.mediamtx_wrapper.controller.set_correlation_id",
            side_effect=mock_set_correlation_id,
        ):
            # Mock successful response
            success_response = self._mock_response(200)
            controller._session.post = AsyncMock(return_value=success_response)

            await controller.update_configuration({"logLevel": "info"})

        # Verify correlation ID was set
        assert len(correlation_ids) > 0
        assert all(isinstance(cid, str) and len(cid) > 0 for cid in correlation_ids)


# Test configuration expectations:
# - Mock aiohttp ClientSession for configuration API calls
# - Test validation of different parameter types and constraints
# - Test error accumulation and reporting without system crashes
# - Verify correlation IDs are set for configuration operations
# - Test both successful and failed configuration updates
# - Verify safe fallback behavior on API/network failures
# - Test validation schema completeness for MediaMTX options
