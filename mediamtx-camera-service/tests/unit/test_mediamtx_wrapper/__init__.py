# tests/unit/test_mediamtx_wrapper/__init__.py
"""
MediaMTX wrapper test package for MediaMTX Camera Service.

Requirements Coverage:
- REQ-TECH-011: MediaMTX streaming server integration
- REQ-TECH-012: HTTP API integration with MediaMTX
- REQ-TECH-013: Stream management and camera stream discovery
- REQ-TECH-014: MediaMTX configuration and stream setup
- REQ-TECH-015: Real-time stream status monitoring
- REQ-TEST-001: Use single systemd-managed MediaMTX service instance
- REQ-TEST-002: No multiple MediaMTX instances or processes
- REQ-TEST-003: Validate against actual production MediaMTX service
- REQ-TEST-004: Use fixed systemd service ports (API: 9997, RTSP: 8554, WebRTC: 8889, HLS: 8888)
- REQ-TEST-005: Coordinate on shared service with proper test isolation
- REQ-TEST-006: Verify MediaMTX service is running via systemd before execution

Test Categories: Unit
"""
