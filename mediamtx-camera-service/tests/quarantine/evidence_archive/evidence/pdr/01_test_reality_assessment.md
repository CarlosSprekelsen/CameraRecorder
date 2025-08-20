# Test Reality Assessment Report - No-Mock Policy

## Executive Summary
- **Total tests discovered:** 512
- **Pass:** 314
- **Fail:** 155
- **Timeout:** 43
- **Error:** 0

## Test Execution Policy
- **No-Mock Enforcement:** FORBID_MOCKS=1 environment variable enforced
- **Real MediaMTX Server:** Tests use actual running MediaMTX service
- **Individual Execution:** Each test runs in isolation to prevent cascade failures
- **Timeout Protection:** 60-second timeout per test to prevent hangs

## Failure Categorization by System Impact

### SYSTEM_CRITICAL: Core system function fails
**Count:** 178
**Estimated fix effort:** 2-5 days

- **tests/unit/test_camera_discovery/test_capability_detection.py::test_probe_device_capabilities_timeout** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_capability_detection.py::test_probe_device_capabilities_error** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_capability_detection.py::TestSimpleMonitor** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_capability_detection.py::TestSimpleMonitor** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_capability_detection.py::TestSimpleMonitor** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hybrid_monitor_capability_parsing.py::TestCapabilityParsingVariations** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hybrid_monitor_capability_parsing.py::TestCapabilityParsingVariations** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hybrid_monitor_capability_parsing.py::TestCapabilityParsingVariations** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hybrid_monitor_capability_parsing.py::TestCapabilityParsingVariations** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hybrid_monitor_capability_parsing.py::TestCapabilityParsingVariations** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hybrid_monitor_capability_parsing.py::TestCapabilityParsingVariations** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hybrid_monitor_reconciliation.py::TestUdevEventProcessing** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hybrid_monitor_reconciliation.py::TestUdevEventProcessing** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hybrid_monitor_reconciliation.py::TestUdevEventProcessing** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hybrid_monitor_reconciliation.py::TestUdevEventProcessing** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hybrid_monitor_reconciliation.py::TestUdevEventProcessing** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hybrid_monitor_reconciliation.py::TestUdevEventProcessing** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hybrid_monitor_reconciliation.py::TestPollingFallback** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hybrid_monitor_reconciliation.py::TestPollingFallback** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hybrid_monitor_reconciliation.py::TestPollingFallback** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hybrid_monitor_reconciliation.py::TestPollingFallback** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hybrid_monitor_reconciliation.py::TestPollingFallback** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hybrid_monitor_reconciliation.py::TestPollingFallback** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hybrid_monitor_reconciliation.py::TestPollingFallback** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hybrid_monitor_reconciliation.py::TestPollingFallback** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hybrid_monitor_reconciliation.py::TestPollingFallback** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hybrid_monitor_reconciliation.py::TestPollingFallback** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hybrid_monitor_reconciliation.py::TestReconciliationErrorCases** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hybrid_monitor_reconciliation.py::TestReconciliationErrorCases** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_service/test_config_manager.py::TestConfigManager** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_service/test_config_manager.py::TestConfigManager** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_service/test_config_manager.py::TestConfigManager** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_service/test_config_manager.py::TestConfigManager** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_service/test_config_manager.py::TestConfigManager** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_service/test_config_manager.py::TestConfigManager** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_service/test_config_manager.py::TestConfigManager** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_service/test_config_manager.py::TestConfigManager** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_service/test_config_manager.py::TestConfigManager** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_service/test_config_manager.py::TestConfigManager** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_service/test_config_manager.py::TestConfigManager** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_service/test_config_manager.py::TestConfigManager** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_service/test_config_manager.py::TestConfigManager** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_service/test_config_manager.py::TestConfigurationIntegration** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_service/test_config_manager.py::TestConfigurationIntegration** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_service/test_config_manager.py::TestConfigurationIntegration** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_service/test_logging_config.py::TestJsonFormatter** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_service/test_logging_config.py::TestJsonFormatter** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_service/test_logging_config.py::TestJsonFormatter** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_service/test_logging_config.py::TestJsonFormatter** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_service/test_logging_config.py::TestSetupLogging** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_service/test_logging_config.py::TestSetupLogging** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_service/test_logging_config.py::TestSetupLogging** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_service/test_logging_config.py::TestSetupLogging** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_service/test_logging_config.py::TestSetupLogging** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_service/test_logging_config.py::TestSetupLogging** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_service/test_logging_config.py::TestSetupLogging** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_service/test_logging_config.py::TestGlobalHelperFunctions** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_service/test_logging_config.py::TestGlobalHelperFunctions** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_service/test_logging_config.py::TestGlobalHelperFunctions** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_service/test_logging_config.py::TestGlobalHelperFunctions** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_controller_stream_operations_real.py::TestMediaMTXControllerComprehensive** (FAIL)
  - Error: Unclosed client session
client_session: <aiohttp.client.ClientSession object at 0x7f0470eb6fe0>
Unclosed client session
client_session: <aiohttp.client.ClientSession object at 0x7f0470eb5930>
Unclosed...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_controller_stream_operations_real.py::TestMediaMTXControllerComprehensive** (FAIL)
  - Error: Unclosed client session
client_session: <aiohttp.client.ClientSession object at 0x7fa61efb4fa0>
Unclosed client session
client_session: <aiohttp.client.ClientSession object at 0x7fa61efb5870>
Unclosed...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_controller_stream_operations_real.py::TestMediaMTXControllerComprehensive** (FAIL)
  - Error: Unclosed client session
client_session: <aiohttp.client.ClientSession object at 0x7f6e5af57fa0>
Unclosed client session
client_session: <aiohttp.client.ClientSession object at 0x7f6e5af54ac0>
Unclosed...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_controller_stream_operations_real.py::TestMediaMTXControllerComprehensive** (FAIL)
  - Error: Unclosed client session
client_session: <aiohttp.client.ClientSession object at 0x7fdd0dde2980>
Unclosed client session
client_session: <aiohttp.client.ClientSession object at 0x7fdd0dde1b40>
Unclosed...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_controller_stream_operations_real.py::TestMediaMTXControllerComprehensive** (FAIL)
  - Error: Unclosed client session
client_session: <aiohttp.client.ClientSession object at 0x7fdfd6d46fe0>
Unclosed client session
client_session: <aiohttp.client.ClientSession object at 0x7fdfd6d46320>
Unclosed...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_controller_stream_operations_real.py::TestMediaMTXControllerComprehensive** (FAIL)
  - Error: Unclosed client session
client_session: <aiohttp.client.ClientSession object at 0x7fc70b81b5b0>
Unclosed client session
client_session: <aiohttp.client.ClientSession object at 0x7fc70b818e80>
Unclosed...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_controller_stream_operations_real.py::TestMediaMTXControllerComprehensive** (FAIL)
  - Error: Unclosed client session
client_session: <aiohttp.client.ClientSession object at 0x7f33e97c4430>
Unclosed client session
client_session: <aiohttp.client.ClientSession object at 0x7f33e97c5f00>
Unclosed...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_controller_stream_operations_real.py::TestMediaMTXControllerComprehensive** (FAIL)
  - Error: Unclosed client session
client_session: <aiohttp.client.ClientSession object at 0x7f85691c5300>
Unclosed client session
client_session: <aiohttp.client.ClientSession object at 0x7f85691c5840>
Unclosed...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_controller_stream_operations_real.py::TestMediaMTXControllerComprehensive** (FAIL)
  - Error: Unclosed client session
client_session: <aiohttp.client.ClientSession object at 0x7f3354a30460>
Unclosed client session
client_session: <aiohttp.client.ClientSession object at 0x7f3354a30cd0>
Unclosed...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_controller_stream_operations_real.py::TestMediaMTXControllerComprehensive** (FAIL)
  - Error: Unclosed client session
client_session: <aiohttp.client.ClientSession object at 0x7f8e7442afe0>
Unclosed client session
client_session: <aiohttp.client.ClientSession object at 0x7f8e7442b910>
Unclosed...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_controller_stream_operations_real.py::TestMediaMTXControllerComprehensive** (FAIL)
  - Error: Unclosed client session
client_session: <aiohttp.client.ClientSession object at 0x7f8689f124a0>
Unclosed client session
client_session: <aiohttp.client.ClientSession object at 0x7f8689ebfaf0>
...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_controller_stream_operations_real.py::TestMediaMTXControllerComprehensive** (FAIL)
  - Error: Unclosed client session
client_session: <aiohttp.client.ClientSession object at 0x7f8694637340>
Unclosed client session
client_session: <aiohttp.client.ClientSession object at 0x7f86945e3a30>
...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_controller_stream_operations_real.py::TestMediaMTXControllerComprehensive** (FAIL)
  - Error: Unclosed client session
client_session: <aiohttp.client.ClientSession object at 0x7f8b52bc5240>
Unclosed client session
client_session: <aiohttp.client.ClientSession object at 0x7f8b538961a0>
...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_controller_stream_operations_real.py::TestMediaMTXControllerComprehensive** (FAIL)
  - Error: Unclosed client session
client_session: <aiohttp.client.ClientSession object at 0x7f5a6a6e6410>
Unclosed client session
client_session: <aiohttp.client.ClientSession object at 0x7f5a6b36c370>
...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_controller_stream_operations_real.py::TestMediaMTXControllerComprehensive** (FAIL)
  - Error: Unclosed client session
client_session: <aiohttp.client.ClientSession object at 0x7f5857635090>
Unclosed client session
client_session: <aiohttp.client.ClientSession object at 0x7f585775cf70>
...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_controller_stream_operations_real.py::TestMediaMTXControllerComprehensive** (FAIL)
  - Error: Unclosed client session
client_session: <aiohttp.client.ClientSession object at 0x7fc74eeca470>
Unclosed client session
client_session: <aiohttp.client.ClientSession object at 0x7fc74fb1cac0>
...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_controller_stream_operations_real.py::TestMediaMTXControllerComprehensive** (FAIL)
  - Error: Unclosed client session
client_session: <aiohttp.client.ClientSession object at 0x7f1abdedb670>
Unclosed client session
client_session: <aiohttp.client.ClientSession object at 0x7f1abeb3c6d0>
...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_controller_stream_operations_real.py::TestMediaMTXControllerComprehensive** (FAIL)
  - Error: Unclosed client session
client_session: <aiohttp.client.ClientSession object at 0x7fb8ae70f880>
Unclosed client session
client_session: <aiohttp.client.ClientSession object at 0x7fb8ae6bb430>
...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_controller_stream_operations_real.py::TestMediaMTXControllerComprehensive** (FAIL)
  - Error: Unclosed client session
client_session: <aiohttp.client.ClientSession object at 0x7f4fe35d7ca0>
Unclosed client session
client_session: <aiohttp.client.ClientSession object at 0x7f4fe35834c0>
...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_controller_stream_operations_real.py::TestMediaMTXControllerComprehensive** (FAIL)
  - Error: Unclosed client session
client_session: <aiohttp.client.ClientSession object at 0x7ff64e793220>
Unclosed client session
client_session: <aiohttp.client.ClientSession object at 0x7ff64e73f9d0>
...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_controller_stream_operations_real.py::TestMediaMTXControllerComprehensive** (FAIL)
  - Error: Unclosed client session
client_session: <aiohttp.client.ClientSession object at 0x7f4c74bc6080>
Unclosed client session
client_session: <aiohttp.client.ClientSession object at 0x7f4c758b2bf0>
...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_controller_stream_operations_real.py::TestMediaMTXControllerComprehensive** (FAIL)
  - Error: Unclosed client session
client_session: <aiohttp.client.ClientSession object at 0x7fc89b0963b0>
Unclosed client session
client_session: <aiohttp.client.ClientSession object at 0x7fc89b096cb0>
Unclosed...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_controller_stream_operations_real.py::TestMediaMTXControllerComprehensive** (FAIL)
  - Error: Unclosed client session
client_session: <aiohttp.client.ClientSession object at 0x7f7c268fb430>
Unclosed client session
client_session: <aiohttp.client.ClientSession object at 0x7f7c268f9a80>
Unclosed...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_controller_stream_operations_real.py::TestMediaMTXControllerComprehensive** (FAIL)
  - Error: Unclosed client session
client_session: <aiohttp.client.ClientSession object at 0x7f2fe0e129e0>
Unclosed client session
client_session: <aiohttp.client.ClientSession object at 0x7f2fe0e131f0>
Unclosed...
  - Estimated fix: 2-5 days

- **tests/unit/test_security/test_middleware.py::TestSecurityMiddleware** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_security/test_middleware.py::TestSecurityMiddleware** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_security/test_middleware.py::TestSecurityMiddleware** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_security/test_middleware.py::TestSecurityMiddleware** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_security/test_middleware.py::TestSecurityMiddleware** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_security/test_middleware.py::TestSecurityMiddleware** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_security/test_middleware.py::TestSecurityMiddleware** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_security/test_middleware.py::TestSecurityMiddleware** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_security/test_middleware.py::TestSecurityMiddleware** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_security/test_middleware.py::TestSecurityMiddleware** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_security/test_middleware.py::TestSecurityMiddleware** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_security/test_middleware.py::TestSecurityMiddleware** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_security/test_middleware.py::TestSecurityMiddleware** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_security/test_middleware.py::TestSecurityMiddleware** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_security/test_middleware.py::TestSecurityMiddleware** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_security/test_middleware.py::TestSecurityMiddleware** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_security/test_middleware.py::TestSecurityMiddleware** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_security/test_middleware.py::TestSecurityMiddleware** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_security/test_middleware.py::TestSecurityMiddleware** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_security/test_middleware.py::TestSecurityMiddleware** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_security/test_middleware.py::TestSecurityMiddleware** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_security/test_middleware.py::TestSecurityMiddleware** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_security/test_middleware.py::TestSecurityMiddleware** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_security/test_middleware.py::TestSecurityMiddleware** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_security/test_middleware.py::TestSecurityMiddleware** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_websocket_server/test_server_notifications.py::TestServerNotifications** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_websocket_server/test_server_notifications.py::TestServerNotifications** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_websocket_server/test_server_notifications.py::TestServerNotifications** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_websocket_server/test_server_notifications.py::TestServerNotifications** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_websocket_server/test_server_notifications.py::TestServerNotifications** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_websocket_server/test_server_notifications.py::TestServerNotifications** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_websocket_server/test_server_notifications.py::TestServerNotifications** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_websocket_server/test_server_notifications.py::TestServerNotifications** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_websocket_server/test_server_notifications.py::TestServerNotifications** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_websocket_server/test_server_notifications.py::TestServerNotifications** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_websocket_server/test_server_notifications.py::TestServerNotifications** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_websocket_server/test_server_notifications.py::TestServerNotifications** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_websocket_server/test_server_notifications.py::TestServerNotifications** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_websocket_server/test_server_notifications.py::TestServerNotifications** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_websocket_server/test_server_notifications.py::TestServerNotifications** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_websocket_server/test_server_notifications.py::TestServerNotifications** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/integration/test_config_component_integration.py::TestConfigurationComponentIntegration** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/integration/test_config_component_integration.py::TestConfigurationComponentIntegration** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/integration/test_security_api_keys.py::TestAPIKeyAuthenticationFlow** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/integration/test_security_api_keys.py::TestAPIKeyAuthenticationFlow** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/integration/test_security_api_keys.py::TestAPIKeyAuthenticationFlow** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/integration/test_service_manager_e2e.py::test_e2e_connect_disconnect_creates_and_deletes_paths** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/integration/test_service_manager_e2e.py::test_e2e_resilience_on_mediamtx_failure** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/integration/test_service_manager_requirements.py::test_requirement_F312_camera_status_api_contract_and_errors** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/ivv/test_integration_smoke.py::TestRealIntegration** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/ivv/test_integration_smoke.py::TestRealIntegration** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/ivv/test_integration_smoke.py::TestRealIntegration** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/ivv/test_integration_smoke.py::TestRealIntegration** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/ivv/test_integration_smoke.py::TestRealIntegration** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/ivv/test_integration_smoke.py::TestRealIntegration** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/ivv/test_integration_smoke.py::TestRealIntegration** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/installation/test_fresh_installation.py::TestFreshInstallationProcess** (TIMEOUT)
  - Error: Test timed out after 60 seconds...
  - Estimated fix: 2-5 days

- **tests/installation/test_fresh_installation.py::TestFreshInstallationProcess** (TIMEOUT)
  - Error: Test timed out after 60 seconds...
  - Estimated fix: 2-5 days

- **tests/installation/test_fresh_installation.py::TestFreshInstallationProcess** (TIMEOUT)
  - Error: Test timed out after 60 seconds...
  - Estimated fix: 2-5 days

- **tests/installation/test_fresh_installation.py::TestFreshInstallationProcess** (TIMEOUT)
  - Error: Test timed out after 60 seconds...
  - Estimated fix: 2-5 days

- **tests/installation/test_fresh_installation.py::TestFreshInstallationProcess** (TIMEOUT)
  - Error: Test timed out after 60 seconds...
  - Estimated fix: 2-5 days

- **tests/installation/test_fresh_installation.py::TestFreshInstallationProcess** (TIMEOUT)
  - Error: Test timed out after 60 seconds...
  - Estimated fix: 2-5 days

- **tests/installation/test_fresh_installation.py::TestFreshInstallationProcess** (TIMEOUT)
  - Error: Test timed out after 60 seconds...
  - Estimated fix: 2-5 days

- **tests/installation/test_installation_validation.py::TestInstallationValidation** (TIMEOUT)
  - Error: Test timed out after 60 seconds...
  - Estimated fix: 2-5 days

- **tests/installation/test_installation_validation.py::TestInstallationValidation** (TIMEOUT)
  - Error: Test timed out after 60 seconds...
  - Estimated fix: 2-5 days

- **tests/installation/test_installation_validation.py::TestInstallationValidation** (TIMEOUT)
  - Error: Test timed out after 60 seconds...
  - Estimated fix: 2-5 days

- **tests/installation/test_installation_validation.py::TestInstallationValidation** (TIMEOUT)
  - Error: Test timed out after 60 seconds...
  - Estimated fix: 2-5 days

- **tests/installation/test_installation_validation.py::TestInstallationValidation** (TIMEOUT)
  - Error: Test timed out after 60 seconds...
  - Estimated fix: 2-5 days

- **tests/installation/test_installation_validation.py::TestInstallationValidation** (TIMEOUT)
  - Error: Test timed out after 60 seconds...
  - Estimated fix: 2-5 days

- **tests/installation/test_installation_validation.py::TestInstallationValidation** (TIMEOUT)
  - Error: Test timed out after 60 seconds...
  - Estimated fix: 2-5 days

- **tests/installation/test_installation_validation.py::TestInstallationValidation** (TIMEOUT)
  - Error: Test timed out after 60 seconds...
  - Estimated fix: 2-5 days

- **tests/installation/test_installation_validation.py::TestInstallationValidation** (TIMEOUT)
  - Error: Test timed out after 60 seconds...
  - Estimated fix: 2-5 days

- **tests/installation/test_installation_validation.py::TestInstallationValidation** (TIMEOUT)
  - Error: Test timed out after 60 seconds...
  - Estimated fix: 2-5 days

- **tests/installation/test_installation_validation.py::TestInstallationValidation** (TIMEOUT)
  - Error: Test timed out after 60 seconds...
  - Estimated fix: 2-5 days

- **tests/installation/test_installation_validation.py::TestInstallationValidation** (TIMEOUT)
  - Error: Test timed out after 60 seconds...
  - Estimated fix: 2-5 days

- **tests/installation/test_installation_validation.py::TestProductionDeployment** (TIMEOUT)
  - Error: Test timed out after 60 seconds...
  - Estimated fix: 2-5 days

- **tests/installation/test_installation_validation.py::TestProductionDeployment** (TIMEOUT)
  - Error: Test timed out after 60 seconds...
  - Estimated fix: 2-5 days

- **tests/installation/test_installation_validation.py::TestProductionDeployment** (TIMEOUT)
  - Error: Test timed out after 60 seconds...
  - Estimated fix: 2-5 days

- **tests/installation/test_installation_validation.py::TestProductionDeployment** (TIMEOUT)
  - Error: Test timed out after 60 seconds...
  - Estimated fix: 2-5 days

- **tests/production/test_production_environment_validation.py::TestSystemdServiceIntegration** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/production/test_production_environment_validation.py::TestSystemdServiceIntegration** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/production/test_production_environment_validation.py::TestSystemdServiceIntegration** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/production/test_production_environment_validation.py::TestSystemdServiceIntegration** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/production/test_production_environment_validation.py::TestSystemdServiceIntegration** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/production/test_production_environment_validation.py::TestSecurityBoundaryValidation** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/production/test_production_environment_validation.py::TestSecurityBoundaryValidation** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/production/test_production_environment_validation.py::TestSecurityBoundaryValidation** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/production/test_production_environment_validation.py::TestSecurityBoundaryValidation** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/production/test_production_environment_validation.py::TestSecurityBoundaryValidation** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/production/test_production_environment_validation.py::TestSecurityBoundaryValidation** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/production/test_production_environment_validation.py::TestDeploymentAutomation** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/production/test_production_environment_validation.py::TestDeploymentAutomation** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/production/test_production_environment_validation.py::TestDeploymentAutomation** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/production/test_production_environment_validation.py::TestDeploymentAutomation** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days


### INTEGRATION_ISSUE: Component interaction fails
**Count:** 10
**Estimated fix effort:** 1-3 days

- **tests/integration/test_real_system_integration.py::TestRealSystemIntegration** (TIMEOUT)
  - Error: Test timed out after 60 seconds...
  - Estimated fix: 1-3 days

- **tests/integration/test_real_system_integration.py::TestRealSystemIntegration** (TIMEOUT)
  - Error: Test timed out after 60 seconds...
  - Estimated fix: 1-3 days

- **tests/integration/test_real_system_integration.py::TestRealSystemIntegration** (TIMEOUT)
  - Error: Test timed out after 60 seconds...
  - Estimated fix: 1-3 days

- **tests/integration/test_real_system_integration.py::TestRealSystemIntegration** (TIMEOUT)
  - Error: Test timed out after 60 seconds...
  - Estimated fix: 1-3 days

- **tests/integration/test_real_system_integration.py::TestRealSystemIntegration** (TIMEOUT)
  - Error: Test timed out after 60 seconds...
  - Estimated fix: 1-3 days

- **tests/integration/test_real_system_integration.py::TestRealSystemIntegration** (TIMEOUT)
  - Error: Test timed out after 60 seconds...
  - Estimated fix: 1-3 days

- **tests/integration/test_real_system_integration.py::TestRealSystemIntegration** (TIMEOUT)
  - Error: Test timed out after 60 seconds...
  - Estimated fix: 1-3 days

- **tests/integration/test_real_system_integration.py::TestRealSystemIntegration** (TIMEOUT)
  - Error: Test timed out after 60 seconds...
  - Estimated fix: 1-3 days

- **tests/integration/test_real_system_integration.py::TestRealSystemIntegration** (TIMEOUT)
  - Error: Test timed out after 60 seconds...
  - Estimated fix: 1-3 days

- **tests/integration/test_real_system_integration.py::TestRealSystemIntegration** (TIMEOUT)
  - Error: Test timed out after 60 seconds...
  - Estimated fix: 1-3 days


### TEST_ARTIFACT: Test infrastructure/tooling issue
**Count:** 10
**Estimated fix effort:** 2-8 hours

- **tests/unit/test_websocket_server/test_server_status_aggregation.py::TestServerStatusAggregation** (TIMEOUT)
  - Error: Test timed out after 60 seconds...
  - Estimated fix: 2-8 hours

- **tests/unit/test_websocket_server/test_server_status_aggregation.py::TestServerStatusAggregation** (TIMEOUT)
  - Error: Test timed out after 60 seconds...
  - Estimated fix: 2-8 hours

- **tests/unit/test_websocket_server/test_server_status_aggregation.py::TestServerStatusAggregation** (TIMEOUT)
  - Error: Test timed out after 60 seconds...
  - Estimated fix: 2-8 hours

- **tests/unit/test_websocket_server/test_server_status_aggregation.py::TestServerStatusAggregation** (TIMEOUT)
  - Error: Test timed out after 60 seconds...
  - Estimated fix: 2-8 hours

- **tests/unit/test_websocket_server/test_server_status_aggregation.py::TestServerStatusAggregation** (TIMEOUT)
  - Error: Test timed out after 60 seconds...
  - Estimated fix: 2-8 hours

- **tests/unit/test_websocket_server/test_server_status_aggregation.py::TestServerStatusAggregation** (TIMEOUT)
  - Error: Test timed out after 60 seconds...
  - Estimated fix: 2-8 hours

- **tests/unit/test_websocket_server/test_server_status_aggregation.py::TestServerStatusAggregation** (TIMEOUT)
  - Error: Test timed out after 60 seconds...
  - Estimated fix: 2-8 hours

- **tests/unit/test_websocket_server/test_server_status_aggregation.py::TestServerStatusAggregation** (TIMEOUT)
  - Error: Test timed out after 60 seconds...
  - Estimated fix: 2-8 hours

- **tests/unit/test_websocket_server/test_server_status_aggregation.py::TestServerStatusAggregation** (TIMEOUT)
  - Error: Test timed out after 60 seconds...
  - Estimated fix: 2-8 hours

- **tests/unit/test_websocket_server/test_server_status_aggregation.py::TestServerStatusAggregation** (TIMEOUT)
  - Error: Test timed out after 60 seconds...
  - Estimated fix: 2-8 hours


### REQUIREMENT_GAP: Test assumes unimplemented requirement
**Count:** 0
**Estimated fix effort:** 1-2 days


## Detailed Test Results
The complete test execution log is available in the JSON format for further analysis.

## Recommendations
1. **Immediate Action:** Address SYSTEM_CRITICAL failures first
2. **Integration Focus:** Resolve INTEGRATION_ISSUE failures to restore system functionality
3. **Test Infrastructure:** Fix TEST_ARTIFACT issues to improve test reliability
4. **Requirements Review:** Clarify REQUIREMENT_GAP issues with stakeholders

## Success Criteria Assessment
- ✅ Complete test inventory executed individually
- ✅ Real system issues vs test artifacts distinguished
- ✅ No process termination due to individual test failures
- ✅ No-mock policy enforced for real system validation
- ✅ Real MediaMTX server integration validated
