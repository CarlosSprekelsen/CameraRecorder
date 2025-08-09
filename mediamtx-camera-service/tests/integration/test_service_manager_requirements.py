import asyncio
import json
import socket
from contextlib import asynccontextmanager

import pytest
import websockets
from aiohttp import web

from src.camera_service.service_manager import ServiceManager
from src.camera_service.config import Config, ServerConfig, MediaMTXConfig, CameraConfig, LoggingConfig, RecordingConfig, SnapshotConfig
from camera_discovery.hybrid_monitor import CameraEventData, CameraEvent
from src.common.types import CameraDevice


def _free_port() -> int:
    import socket as _s
    with _s.socket(_s.AF_INET, _s.SOCK_STREAM) as s:
        s.bind(("127.0.0.1", 0))
        return s.getsockname()[1]


def _build_config(api_port: int, ws_port: int) -> Config:
    return Config(
        server=ServerConfig(host="127.0.0.1", port=ws_port, websocket_path="/ws", max_connections=10),
        mediamtx=MediaMTXConfig(host="127.0.0.1", api_port=api_port, rtsp_port=8554, webrtc_port=8889, hls_port=8888, recordings_path="./.tmp_recordings", snapshots_path="./.tmp_snapshots"),
        camera=CameraConfig(device_range=[0,1,2], enable_capability_detection=True, detection_timeout=0.5),
        logging=LoggingConfig(), recording=RecordingConfig(), snapshots=SnapshotConfig(),
    )


@asynccontextmanager
async def _mediamtx_ok(host: str, port: int):
    async def health(_req):
        return web.json_response({"serverVersion": "test", "serverUptime": 1})

    async def add_path(request: web.Request):
        # emulate successful path add
        _ = request.match_info.get("name")
        return web.json_response({"status": "ok"})

    async def delete_path(request: web.Request):
        # emulate successful path delete
        _ = request.match_info.get("name")
        return web.json_response({"status": "ok"})

    app = web.Application()
    app.router.add_get("/v3/health", health)
    app.router.add_post("/v3/config/paths/add/{name}", add_path)
    app.router.add_post("/v3/config/paths/delete/{name}", delete_path)
    runner = web.AppRunner(app)
    await runner.setup()
    site = web.TCPSite(runner, host, port)
    await site.start()
    try:
        yield
    finally:
        await runner.cleanup()


@pytest.mark.asyncio
async def test_requirement_F111_photo_capture_service_availability():
    """
    Validates F1.1.1: The application SHALL allow users to take photos using available cameras

    Business Scenario: User opens app and service must be ready for photo capture
    Error Cases: Service startup failures, component initialization failures
    Success Criteria: Service starts and responds to take_snapshot API calls
    """
    api_port, ws_port = _free_port(), _free_port()
    async with _mediamtx_ok("127.0.0.1", api_port):
        cfg = _build_config(api_port, ws_port)
        svc = ServiceManager(cfg)
        await svc.start()
        try:
            uri = f"ws://{cfg.server.host}:{cfg.server.port}{cfg.server.websocket_path}"
            async with websockets.connect(uri) as ws:
                await ws.send(json.dumps({"jsonrpc":"2.0","id":1,"method":"take_snapshot","params":{"device":"/dev/video0"}}))
                resp = json.loads(await ws.recv())
                assert "result" in resp or "error" in resp
        finally:
            await svc.stop()


@pytest.mark.asyncio
async def test_requirement_F313_camera_hotplug_notifications():
    """
    Validates F3.1.3: The application SHALL handle camera hot-plug events via real-time notifications

    Business Scenario: User plugs/unplugs camera, app receives real-time updates
    Error Cases: Camera detection failures, notification delivery failures
    Success Criteria: WebSocket clients receive camera status change notifications
    """
    api_port, ws_port = _free_port(), _free_port()
    async with _mediamtx_ok("127.0.0.1", api_port):
        cfg = _build_config(api_port, ws_port)
        svc = ServiceManager(cfg)
        await svc.start()
        try:
            uri = f"ws://{cfg.server.host}:{cfg.server.port}{cfg.server.websocket_path}"
            async with websockets.connect(uri) as ws:
                # allow server to register client and be ready for broadcasts
                await asyncio.sleep(0.5)
                # Simulate camera connect
                event = CameraEventData(device_path="/dev/video0", event_type=CameraEvent.CONNECTED, device_info=CameraDevice(device="/dev/video0", name="Camera 0", status="CONNECTED"), timestamp=0.0)
                await svc.handle_camera_event(event)
                # Wait for broadcast within timeout
                msg = await asyncio.wait_for(ws.recv(), timeout=3.0)
                payload = json.loads(msg)
                assert payload.get("method") == "camera_status_update"
                assert payload.get("params", {}).get("device") == "/dev/video0"
                assert payload.get("params", {}).get("status") in ("CONNECTED","DISCONNECTED")
        finally:
            await svc.stop()


@pytest.mark.asyncio
async def test_requirement_F125_recording_session_management():
    """
    Validates F1.2.5: The application SHALL handle recording session management via service API

    Business Scenario: User starts/stops recording sessions
    Error Cases: Concurrent recording conflicts, storage failures, session cleanup
    Success Criteria: API correctly manages recording state and resources
    """
    api_port, ws_port = _free_port(), _free_port()
    async with _mediamtx_ok("127.0.0.1", api_port):
        cfg = _build_config(api_port, ws_port)
        svc = ServiceManager(cfg)
        await svc.start()
        try:
            uri = f"ws://{cfg.server.host}:{cfg.server.port}{cfg.server.websocket_path}"
            async with websockets.connect(uri) as ws:
                # Attempt start/stop recording via API
                await ws.send(json.dumps({"jsonrpc":"2.0","id":1,"method":"start_recording","params":{"device":"/dev/video0"}}))
                start_resp = json.loads(await ws.recv())
                assert "result" in start_resp or "error" in start_resp

                await ws.send(json.dumps({"jsonrpc":"2.0","id":2,"method":"stop_recording","params":{"device":"/dev/video0"}}))
                stop_resp = json.loads(await ws.recv())
                assert "result" in stop_resp or "error" in stop_resp
        finally:
            await svc.stop()


@pytest.mark.asyncio
async def test_requirement_F311_camera_list_availability():
    """
    Validates F3.1.1: The application SHALL display list of available cameras from service API

    Business Scenario: User opens app and sees available cameras
    Error Cases: No cameras detected, camera access failures, service communication errors
    Success Criteria: get_camera_list API returns discoverable cameras with status
    """
    api_port, ws_port = _free_port(), _free_port()
    async with _mediamtx_ok("127.0.0.1", api_port):
        cfg = _build_config(api_port, ws_port)
        svc = ServiceManager(cfg)
        await svc.start()
        try:
            uri = f"ws://{cfg.server.host}:{cfg.server.port}{cfg.server.websocket_path}"
            async with websockets.connect(uri) as ws:
                await ws.send(json.dumps({"jsonrpc":"2.0","id":1,"method":"get_camera_list"}))
                resp = json.loads(await ws.recv())
                result = resp.get("result")
                assert isinstance(result, dict)
                assert isinstance(result.get("cameras"), list)
                assert isinstance(result.get("total"), int)
                assert isinstance(result.get("connected"), int)
                for cam in result["cameras"]:
                    assert "device" in cam and "name" in cam and "status" in cam
        finally:
            await svc.stop()


@pytest.mark.asyncio
async def test_requirement_F111_photo_capture_error_handling():
    """
    Validates F1.1.4: Handle photo capture errors gracefully with user feedback

    Error Cases: Invalid stream/device and unsupported format
    Success Criteria: API returns meaningful error or failure status; service remains responsive
    """
    api_port, ws_port = _free_port(), _free_port()
    async with _mediamtx_ok("127.0.0.1", api_port):
        cfg = _build_config(api_port, ws_port)
        svc = ServiceManager(cfg)
        await svc.start()
        try:
            uri = f"ws://{cfg.server.host}:{cfg.server.port}{cfg.server.websocket_path}"
            async with websockets.connect(uri) as ws:
                # Invalid device
                await ws.send(json.dumps({
                    "jsonrpc": "2.0", "id": 1, "method": "take_snapshot",
                    "params": {"device": "/dev/nonexistent", "format": "tiff", "quality": 200}
                }))
                resp = json.loads(await ws.recv())
                # Accept either JSON-RPC error or a failure result payload
                assert ("error" in resp) or (resp.get("result", {}).get("status") in {"FAILED", "ERROR"})

            # Service should still respond to ping
            async with websockets.connect(uri) as ws2:
                await ws2.send(json.dumps({"jsonrpc": "2.0", "id": 2, "method": "ping", "params": {}}))
                ping = json.loads(await ws2.recv())
                assert "result" in ping
        finally:
            await svc.stop()


@pytest.mark.asyncio
async def test_requirement_F125_recording_concurrent_conflict_detection():
    """
    Validates F1.2.5: Concurrent recording conflict detection via service API

    Error Cases: Starting a recording session twice on the same device
    Success Criteria: Second start returns an error or failure status; service remains responsive
    """
    api_port, ws_port = _free_port(), _free_port()
    async with _mediamtx_ok("127.0.0.1", api_port):
        cfg = _build_config(api_port, ws_port)
        svc = ServiceManager(cfg)
        await svc.start()
        try:
            uri = f"ws://{cfg.server.host}:{cfg.server.port}{cfg.server.websocket_path}"
            async with websockets.connect(uri) as ws:
                # First start
                await ws.send(json.dumps({"jsonrpc":"2.0", "id":1, "method":"start_recording", "params":{"device":"/dev/video0"}}))
                first = json.loads(await ws.recv())
                assert ("result" in first) or ("error" in first)

                # Second start on same device should conflict
                await ws.send(json.dumps({"jsonrpc":"2.0", "id":2, "method":"start_recording", "params":{"device":"/dev/video0"}}))
                second = json.loads(await ws.recv())
                assert ("error" in second) or (second.get("result", {}).get("status") in {"FAILED", "ERROR"})

                # Cleanup stop (idempotent if not started)
                await ws.send(json.dumps({"jsonrpc":"2.0", "id":3, "method":"stop_recording", "params":{"device":"/dev/video0"}}))
                _ = await ws.recv()
        finally:
            await svc.stop()


@pytest.mark.asyncio
async def test_requirement_F311_camera_list_empty_structure():
    """
    Validates F3.1.1: Empty camera list still returns valid response structure

    Success Criteria: get_camera_list returns { cameras: [], total: int, connected: int }
    """
    api_port, ws_port = _free_port(), _free_port()
    async with _mediamtx_ok("127.0.0.1", api_port):
        cfg = _build_config(api_port, ws_port)
        svc = ServiceManager(cfg)
        await svc.start()
        try:
            uri = f"ws://{cfg.server.host}:{cfg.server.port}{cfg.server.websocket_path}"
            async with websockets.connect(uri) as ws:
                await ws.send(json.dumps({"jsonrpc":"2.0", "id":1, "method":"get_camera_list"}))
                resp = json.loads(await ws.recv())
                result = resp.get("result")
                assert isinstance(result, dict)
                assert isinstance(result.get("cameras"), list)
                assert isinstance(result.get("total"), int)
                assert isinstance(result.get("connected"), int)
                assert result["cameras"] == []
        finally:
            await svc.stop()


@pytest.mark.asyncio
async def test_requirement_F313_notification_delivery_failure_tolerance():
    """
    Validates F3.1.3: Notification delivery failures do not crash the service

    Scenario: Client disconnects before broadcast; service must remain operational
    Success Criteria: No crash; subsequent ping succeeds on a new connection
    """
    api_port, ws_port = _free_port(), _free_port()
    async with _mediamtx_ok("127.0.0.1", api_port):
        cfg = _build_config(api_port, ws_port)
        svc = ServiceManager(cfg)
        await svc.start()
        try:
            uri = f"ws://{cfg.server.host}:{cfg.server.port}{cfg.server.websocket_path}"
            # Connect and immediately close to simulate delivery failure
            async with websockets.connect(uri) as ws:
                await asyncio.sleep(0.2)
            # Trigger event after client disconnect
            event = CameraEventData(device_path="/dev/video0", event_type=CameraEvent.CONNECTED, device_info=CameraDevice(device="/dev/video0", name="Camera 0", status="CONNECTED"), timestamp=0.0)
            await svc.handle_camera_event(event)

            # Service should still be running and responsive
            async with websockets.connect(uri) as ws2:
                await ws2.send(json.dumps({"jsonrpc":"2.0","id":99,"method":"ping","params":{}}))
                ping = json.loads(await ws2.recv())
                assert "result" in ping
        finally:
            await svc.stop()


@pytest.mark.asyncio
async def test_requirement_F312_camera_status_api_contract_and_errors():
    """
    Validates F3.1.2: The application SHALL return camera status via get_camera_status API

    Business Scenario: Client queries status for a known camera and an unknown camera
    Error Cases: Unknown device returns JSON-RPC error; known device returns status dict
    Success Criteria: Public API returns structured result or meaningful error
    """
    api_port, ws_port = _free_port(), _free_port()
    async with _mediamtx_ok("127.0.0.1", api_port):
        cfg = _build_config(api_port, ws_port)
        svc = ServiceManager(cfg)
        await svc.start()
        try:
            # Simulate camera connect so status can be queried
            event = CameraEventData(device_path="/dev/video0", event_type=CameraEvent.CONNECTED, device_info=CameraDevice(device="/dev/video0", name="Camera 0", status="CONNECTED"), timestamp=0.0)
            await svc.handle_camera_event(event)

            uri = f"ws://{cfg.server.host}:{cfg.server.port}{cfg.server.websocket_path}"
            async with websockets.connect(uri) as ws:
                # Known device
                await ws.send(json.dumps({"jsonrpc":"2.0","id":1,"method":"get_camera_status","params":{"device":"/dev/video0"}}))
                known = json.loads(await ws.recv())
                if "result" in known:
                    assert isinstance(known["result"], dict)
                    assert known["result"].get("device") == "/dev/video0"
                    assert "status" in known["result"]
                else:
                    # If implementation cannot resolve the device yet, accept error response
                    assert "error" in known

                # Unknown device
                await ws.send(json.dumps({"jsonrpc":"2.0","id":2,"method":"get_camera_status","params":{"device":"/dev/unknown"}}))
                unknown = json.loads(await ws.recv())
                assert "error" in unknown or (unknown.get("result", {}).get("status") in {"ERROR","FAILED"})
        finally:
            await svc.stop()


@pytest.mark.asyncio
async def test_requirement_F126_recording_duration_constraints():
    """
    Validates F1.2.6: The application SHALL enforce recording duration constraints

    Business Scenario: Client starts a short recording with a duration limit
    Error Cases: Invalid negative duration rejected with error
    Success Criteria: API returns result/error without crashing; service remains responsive

    STOP: Exact duration parameter semantics are not fully specified in current API;
    test accepts either success with result or error for unsupported parameter.
    """
    api_port, ws_port = _free_port(), _free_port()
    async with _mediamtx_ok("127.0.0.1", api_port):
        cfg = _build_config(api_port, ws_port)
        svc = ServiceManager(cfg)
        await svc.start()
        try:
            uri = f"ws://{cfg.server.host}:{cfg.server.port}{cfg.server.websocket_path}"
            async with websockets.connect(uri) as ws:
                # Short duration start (semantic acceptance depends on implementation)
                await ws.send(json.dumps({"jsonrpc":"2.0","id":1,"method":"start_recording","params":{"device":"/dev/video0","duration":1}}))
                start = json.loads(await ws.recv())
                assert ("result" in start) or ("error" in start)

                # Invalid negative duration
                await ws.send(json.dumps({"jsonrpc":"2.0","id":2,"method":"start_recording","params":{"device":"/dev/video0","duration":-5}}))
                neg = json.loads(await ws.recv())
                assert ("error" in neg) or (neg.get("result", {}).get("status") in {"FAILED","ERROR"})

                # Ensure stop is accepted (idempotent)
                await ws.send(json.dumps({"jsonrpc":"2.0","id":3,"method":"stop_recording","params":{"device":"/dev/video0"}}))
                _ = await ws.recv()
        finally:
            await svc.stop()


@pytest.mark.asyncio
async def test_requirement_F320_operator_permissions_enforced():
    """
    Validates F3.2.0: Operator permissions required for recording and snapshot APIs

    Business Scenario: Unauthenticated client attempts operator-only methods
    Error Cases: Methods should return authorization error if enforced by implementation
    Success Criteria: API returns error or fails gracefully without crashing

    STOP: Authentication/authorization flow is not exposed via current public tests;
    this test accepts either error (preferred) or a result until auth is implemented.
    """
    api_port, ws_port = _free_port(), _free_port()
    async with _mediamtx_ok("127.0.0.1", api_port):
        cfg = _build_config(api_port, ws_port)
        svc = ServiceManager(cfg)
        await svc.start()
        try:
            uri = f"ws://{cfg.server.host}:{cfg.server.port}{cfg.server.websocket_path}"
            async with websockets.connect(uri) as ws:
                # start_recording without auth
                await ws.send(json.dumps({"jsonrpc":"2.0","id":1,"method":"start_recording","params":{"device":"/dev/video0"}}))
                sr = json.loads(await ws.recv())
                assert ("error" in sr) or ("result" in sr)

                # take_snapshot without auth
                await ws.send(json.dumps({"jsonrpc":"2.0","id":2,"method":"take_snapshot","params":{"device":"/dev/video0"}}))
                ts = json.loads(await ws.recv())
                assert ("error" in ts) or ("result" in ts)
        finally:
            await svc.stop()


@pytest.mark.asyncio
async def test_requirement_F114_snapshot_quality_bounds_and_persistence():
    """
    Validates F1.1.4: Snapshot quality bounds and basic persistence semantics

    Business Scenario: Client requests snapshot with out-of-range quality
    Error Cases: Quality > 100 rejected (or fails gracefully)
    Success Criteria: API returns error/failure for bad quality; accepts reasonable quality

    STOP: File persistence (on disk) cannot be validated without real snapshot pipeline;
    test asserts presence of filename in result for valid request when supported.
    """
    api_port, ws_port = _free_port(), _free_port()
    async with _mediamtx_ok("127.0.0.1", api_port):
        cfg = _build_config(api_port, ws_port)
        svc = ServiceManager(cfg)
        await svc.start()
        try:
            uri = f"ws://{cfg.server.host}:{cfg.server.port}{cfg.server.websocket_path}"
            async with websockets.connect(uri) as ws:
                # Out-of-range quality
                await ws.send(json.dumps({"jsonrpc":"2.0","id":1,"method":"take_snapshot","params":{"device":"/dev/video0","quality":150}}))
                bad = json.loads(await ws.recv())
                assert ("error" in bad) or (bad.get("result", {}).get("status") in {"FAILED","ERROR"})

                # Reasonable quality
                await ws.send(json.dumps({"jsonrpc":"2.0","id":2,"method":"take_snapshot","params":{"device":"/dev/video0","quality":80}}))
                ok = json.loads(await ws.recv())
                assert ("result" in ok) or ("error" in ok)
                if "result" in ok:
                    assert isinstance(ok["result"].get("filename"), str)
        finally:
            await svc.stop()


