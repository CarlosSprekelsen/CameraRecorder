Test Validation Audit (Emergency)

- Context: IV&V assessing whether tests discover real bugs vs accommodate broken code.

1) Mutation testing
- Mutated `_method_ping` to return "PONG!".
- Failing tests: `tests/unit/test_websocket_bind.py::test_websocket_server_binds_and_ping`, `tests/unit/test_websocket_server/test_server_method_handlers.py::TestServerMethodHandlers::test_ping_method`.
- Reverted; tests passed again. These detect regressions as intended.

2) Unit test run findings (why they fail)
- LoggingConfig: TypeError from iterating `root_logger.filters` because tests replace list with `Mock`. Brittle mocking.
- MediaMTX controller (config/update): aiohttp client mocked without async context (`__aenter__/__aexit__`) → `AttributeError: __aenter__`.
- Health/circuit breaker: expected activations not observed; simulated failures not driven long enough; thresholds/intervals not tuned for tests.
- Snapshot (real): failures due to logging misuse in code (`extra` uses reserved key `filename`) causing logging exceptions; status becomes failed.
- WebSocket API drift: tests expect `status == "completed"`, code returns `"SUCCESS"`; tests expect graceful defaults when `camera_monitor` missing, code raises `CameraNotFoundError` and prepopulates default stream URLs.
- Notifications: client considered disconnected because websocket mock lacks realistic `open/closed` attributes/AsyncMock send.
- Tooling: `mediamtx --version` assertion expects the word "mediamtx" in stdout; actual output is version only.

3) Mocking vs real systems
- Many unit tests rely on mocks (camera monitor, MediaMTX controller, aiohttp). Several mocks are incorrect for async usage.
- Real FFmpeg/MediaMTX are used in integration/IV&V suites; some tests allow failure paths (skip/partial accept), which can mask real defects.
- Recommendation: keep real components in integration gates; fix unit tests to use correct AsyncMock patterns; add a mandatory smoke test that fails hard on WebSocket startup/connectivity.

4) Coverage vs functionality gap
- WebSocket startup issues not strongly asserted at unit level; integration/IV&V sometimes report-only instead of gating failures → gaps where system can be broken yet tests still "pass".

IV&V decision
- Conditional rejection of unit-test gate as a release-quality signal until:
  - Code: rename logging extras (avoid `filename`), ensure aiohttp session lifecycle and context managers, decide canonical API contract (status strings, degradation behavior, default streams) and align code.
  - Tests: correct async mocks (`AsyncMock` + async CM), adjust brittle assertions (version output), ensure websocket mock attributes.
  - Gates: enforce real startup/connectivity smoke as a blocking check.

Artifacts
- Unit run: `pytest tests/unit` (see console excerpt in session).
- Mutation: `_method_ping` regression detected and reverted.
