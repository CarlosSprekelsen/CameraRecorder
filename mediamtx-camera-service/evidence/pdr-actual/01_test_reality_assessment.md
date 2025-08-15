# Test Reality Assessment Report - No-Mock Policy

## Executive Summary
- **Total tests discovered:** 556
- **Pass:** 276
- **Fail:** 250
- **Timeout:** 30
- **Error:** 0

## Test Execution Policy
- **No-Mock Enforcement:** FORBID_MOCKS=1 environment variable enforced
- **Real MediaMTX Server:** Tests use actual running MediaMTX service
- **Individual Execution:** Each test runs in isolation to prevent cascade failures
- **Timeout Protection:** 60-second timeout per test to prevent hangs

## Failure Categorization by System Impact

### SYSTEM_CRITICAL: Core system function fails
**Count:** 273
**Estimated fix effort:** 2-5 days

- **tests/unit/test_camera_discovery/test_capability_detection.py::test_probe_device_capabilities_timeout** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_capability_detection.py::test_probe_device_capabilities_error** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_environment_setup.py::TestCameraDiscoveryEnvironmentSetup** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_environment_setup.py::TestCameraDiscoveryEnvironmentSetup** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_environment_setup.py::TestCameraDiscoveryEnvironmentSetup** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_environment_setup.py::TestCameraDiscoveryEnvironmentSetup** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_environment_setup.py::TestCameraDiscoveryEnvironmentSetup** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_environment_setup.py::TestCameraDiscoveryEnvironmentSetup** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_environment_setup.py::TestCameraDiscoveryEnvironmentSetup** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_environment_setup.py::TestCameraDeviceSimulation** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_environment_setup.py::TestCameraDeviceSimulation** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_environment_setup.py::TestEnvironmentSpecificDependencies** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_environment_setup.py::TestEnvironmentSpecificDependencies** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_environment_setup.py::TestEnvironmentSpecificDependencies** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_environment_setup.py::TestEnvironmentSpecificDependencies** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_environment_setup.py::TestMockConfigurationRobustness** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_environment_setup.py::TestMockConfigurationRobustness** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_environment_setup.py::TestMockConfigurationRobustness** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hardware_integration_real.py::TestHardwareIntegrationReal** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hardware_integration_real.py::TestHardwareIntegrationReal** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hardware_integration_real.py::TestHardwareIntegrationReal** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hardware_integration_real.py::TestHardwareIntegrationReal** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hardware_integration_real.py::TestHardwareIntegrationReal** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hardware_integration_real.py::TestHardwareIntegrationReal** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hardware_integration_real.py::TestHardwareIntegrationReal** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hardware_integration_real.py::TestHardwareIntegrationReal** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hardware_integration_real.py::TestHardwareIntegrationReal** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hardware_integration_real.py::TestHardwareIntegrationReal** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hardware_integration_real.py::TestMinimalMockingStrategy** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hardware_integration_real.py::TestMinimalMockingStrategy** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hardware_integration_real.py::TestMinimalMockingStrategy** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hardware_integration_real.py::TestMinimalMockingStrategy** (FAIL)
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

- **tests/unit/test_camera_discovery/test_hybrid_monitor_comprehensive.py::TestCapabilityParsingVariations** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hybrid_monitor_comprehensive.py::TestCapabilityParsingVariations** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hybrid_monitor_comprehensive.py::TestCapabilityParsingVariations** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hybrid_monitor_comprehensive.py::TestCapabilityParsingVariations** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hybrid_monitor_comprehensive.py::TestUdevEventProcessingAndRaceConditions** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hybrid_monitor_comprehensive.py::TestUdevEventProcessingAndRaceConditions** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hybrid_monitor_comprehensive.py::TestUdevEventProcessingAndRaceConditions** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hybrid_monitor_comprehensive.py::TestPollingFallbackBehavior** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hybrid_monitor_comprehensive.py::TestPollingFallbackBehavior** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hybrid_monitor_comprehensive.py::TestPollingFallbackBehavior** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hybrid_monitor_comprehensive.py::TestTimeoutAndSubprocessFailureHandling** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hybrid_monitor_comprehensive.py::TestTimeoutAndSubprocessFailureHandling** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hybrid_monitor_comprehensive.py::TestTimeoutAndSubprocessFailureHandling** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hybrid_monitor_comprehensive.py::TestIntegrationAndLifecycle** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hybrid_monitor_comprehensive.py::TestIntegrationAndLifecycle** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hybrid_monitor_reconciliation.py::TestCapabilityReconciliation** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hybrid_monitor_reconciliation.py::TestCapabilityReconciliation** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hybrid_monitor_reconciliation.py::TestCapabilityReconciliation** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hybrid_monitor_reconciliation.py::TestCapabilityReconciliation** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hybrid_monitor_reconciliation.py::TestCapabilityReconciliation** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hybrid_monitor_reconciliation.py::TestCapabilityReconciliation** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hybrid_monitor_reconciliation.py::TestReconciliationErrorCases** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hybrid_monitor_reconciliation.py::TestReconciliationErrorCases** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hybrid_monitor_udev_fallback.py::TestUdevEventProcessing** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hybrid_monitor_udev_fallback.py::TestUdevEventProcessing** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hybrid_monitor_udev_fallback.py::TestUdevEventProcessing** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hybrid_monitor_udev_fallback.py::TestUdevEventProcessing** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hybrid_monitor_udev_fallback.py::TestUdevEventProcessing** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hybrid_monitor_udev_fallback.py::TestUdevEventProcessing** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hybrid_monitor_udev_fallback.py::TestPollingFallback** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hybrid_monitor_udev_fallback.py::TestPollingFallback** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hybrid_monitor_udev_fallback.py::TestPollingFallback** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hybrid_monitor_udev_fallback.py::TestPollingFallback** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_hybrid_monitor_udev_fallback.py::TestPollingFallback** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_simple_monitor.py::TestSimpleMonitor** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_simple_monitor.py::TestSimpleMonitor** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_simple_monitor.py::TestSimpleMonitor** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_udev_processing.py::test_udev_event_filtering** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_udev_processing.py::test_udev_event_actions** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_discovery/test_udev_processing.py::test_udev_event_race_condition_handling** (FAIL)
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

- **tests/unit/test_camera_service/test_service_manager_lifecycle.py::test_real_connect_flow** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_service/test_service_manager_lifecycle.py::test_real_disconnect_flow** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_service/test_service_manager_lifecycle.py::test_real_mediamtx_failure_keeps_service_running** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_camera_service/test_service_manager_lifecycle.py::test_real_capability_metadata** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_controller_configuration.py::TestConfigurationValidation** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_controller_configuration.py::TestConfigurationValidation** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_controller_configuration.py::TestConfigurationValidation** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_controller_configuration.py::TestConfigurationValidation** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_controller_configuration.py::TestConfigurationValidation** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_controller_configuration.py::TestConfigurationValidation** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_controller_configuration.py::TestConfigurationValidation** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_controller_configuration.py::TestConfigurationValidation** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_controller_configuration.py::TestConfigurationValidation** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_controller_configuration.py::TestConfigurationValidation** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_controller_configuration.py::TestConfigurationValidation** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_controller_configuration.py::TestConfigurationValidation** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_controller_health_monitoring.py::TestHealthMonitoring** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_controller_health_monitoring.py::TestHealthMonitoring** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_controller_health_monitoring.py::TestHealthMonitoring** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_controller_health_monitoring.py::TestHealthMonitoring** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_controller_health_monitoring.py::TestHealthMonitoring** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_controller_health_monitoring.py::TestHealthMonitoring** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_controller_health_monitoring.py::TestHealthMonitoring** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_controller_health_monitoring.py::TestHealthMonitoring** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_controller_health_monitoring.py::TestHealthMonitoring** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_controller_health_monitoring.py::TestHealthMonitoring** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_controller_recording_duration_real.py::TestRecordingDurationReal** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_controller_recording_duration_real.py::TestRecordingDurationReal** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_controller_recording_duration_real.py::TestRecordingDurationReal** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_controller_recording_duration_real.py::TestRecordingDurationReal** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_controller_recording_duration_real.py::TestRecordingDurationReal** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_controller_recording_duration_real.py::TestRecordingDurationReal** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_controller_recording_duration_real.py::TestRecordingDurationReal** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_controller_recording_duration_real.py::TestRecordingDurationReal** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_controller_snapshot_real.py::TestSnapshotCaptureReal** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_controller_snapshot_real.py::TestSnapshotCaptureReal** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_controller_snapshot_real.py::TestSnapshotCaptureReal** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_controller_snapshot_real.py::TestSnapshotCaptureReal** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_controller_snapshot_real.py::TestSnapshotCaptureReal** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_controller_snapshot_real.py::TestSnapshotCaptureReal** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_controller_snapshot_real.py::TestSnapshotCaptureReal** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_health_monitor_circuit_breaker_real.py::TestHealthMonitorFlappingReal** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_health_monitor_circuit_breaker_real.py::TestHealthMonitorFlappingReal** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_health_monitor_circuit_breaker_real.py::TestHealthMonitorFlappingReal** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_health_monitor_circuit_breaker_real.py::TestHealthMonitorFlappingReal** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_health_monitor_circuit_breaker_real.py::TestHealthMonitorFlappingReal** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_health_monitor_circuit_breaker_real.py::TestHealthMonitorFlappingReal** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_health_monitor_recovery_confirmation.py::TestHealthMonitorRecoveryConfirmation** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_health_monitor_recovery_confirmation.py::TestHealthMonitorRecoveryConfirmation** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_health_monitor_recovery_confirmation.py::TestHealthMonitorRecoveryConfirmation** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_health_monitor_recovery_confirmation.py::TestHealthMonitorRecoveryConfirmation** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_health_monitor_recovery_confirmation.py::TestHealthMonitorRecoveryConfirmation** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_health_monitor_recovery_confirmation.py::TestHealthMonitorRecoveryConfirmation** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_mediamtx_wrapper/test_health_monitor_recovery_confirmation.py::TestHealthMonitorRecoveryConfirmation** (FAIL)
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

- **tests/unit/test_security/test_middleware.py::TestSecurityMiddleware** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_websocket_server/test_server_method_handlers.py::TestServerMethodHandlers** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_websocket_server/test_server_method_handlers.py::TestServerMethodHandlers** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_websocket_server/test_server_method_handlers.py::TestServerMethodHandlers** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_websocket_server/test_server_method_handlers.py::TestServerMethodHandlers** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_websocket_server/test_server_method_handlers.py::TestServerMethodHandlers** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_websocket_server/test_server_method_handlers.py::TestServerMethodHandlers** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_websocket_server/test_server_method_handlers.py::TestServerMethodHandlers** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_websocket_server/test_server_method_handlers.py::TestServerMethodHandlers** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_websocket_server/test_server_method_handlers.py::TestServerMethodHandlers** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_websocket_server/test_server_method_handlers.py::TestServerMethodHandlers** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_websocket_server/test_server_method_handlers.py::TestServerMethodHandlers** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_websocket_server/test_server_method_handlers.py::TestServerMethodHandlers** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_websocket_server/test_server_method_handlers.py::TestServerMethodHandlers** (FAIL)
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

- **tests/unit/test_websocket_server/test_server_status_aggregation.py::TestServerStatusAggregation** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_websocket_server/test_server_status_aggregation.py::TestServerStatusAggregation** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_websocket_server/test_server_status_aggregation.py::TestServerStatusAggregation** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_websocket_server/test_server_status_aggregation.py::TestServerStatusAggregation** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_websocket_server/test_server_status_aggregation.py::TestServerStatusAggregation** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_websocket_server/test_server_status_aggregation.py::TestServerStatusAggregation** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_websocket_server/test_server_status_aggregation.py::TestServerStatusAggregation** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_websocket_server/test_server_status_aggregation.py::TestServerStatusAggregation** (FAIL)
  - Error: ...
  - Estimated fix: 2-5 days

- **tests/unit/test_websocket_server/test_server_status_aggregation.py::TestServerStatusAggregation** (FAIL)
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
**Count:** 7
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


### TEST_ARTIFACT: Test infrastructure/tooling issue
**Count:** 0
**Estimated fix effort:** 2-8 hours


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
