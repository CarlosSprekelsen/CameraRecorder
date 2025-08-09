# Smoke Run Metrics (Final)

- Start: 2025-08-09T10:02:31+00:00
- End: 2025-08-09T10:02:39+00:00
- Wall duration: 8s

## Results
- Total tests (incl. skipped): 24
- Passed: 24
- Failed: 0
- Errors: 0
- Skipped: 0
- XFailed: 0
- XPassed: 0
- Pass rate: 100.00%
- Flaky: 0.00%

## Top slow tests
- 2.01s call     tests/ivv/test_real_system_validation.py::TestRealSystemValidation::test_websocket_server_real_validation
- 1.01s call     tests/ivv/test_integration_smoke.py::TestRealIntegration::test_real_websocket_server_integration
- 1.01s call     tests/ivv/test_integration_smoke.py::TestRealIntegration::test_real_error_handling_integration
- 1.01s call     tests/ivv/test_real_integration.py::TestRealIntegration::test_real_websocket_server_integration
- 1.01s call     tests/ivv/test_integration_smoke.py::TestRealIntegration::test_real_notification_system
- 1.01s call     tests/ivv/test_real_integration.py::TestRealIntegration::test_real_error_handling_integration
- 0.13s call     tests/ivv/test_camera_monitor_debug.py::TestCameraMonitorDebug::test_device_access_debug
- 0.08s teardown tests/ivv/test_integration_smoke.py::TestRealIntegration::test_real_performance_validation
- 0.08s teardown tests/ivv/test_camera_monitor_debug.py::TestCameraMonitorDebug::test_monitor_startup_debug
- 0.03s call     tests/ivv/test_integration_smoke.py::TestRealIntegration::test_real_service_manager_integration
