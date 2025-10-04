// Package testutils provides universal test utilities for the MediaMTX Camera Service.
//
// This package contains domain-agnostic test infrastructure that can be used
// across all modules, providing common setup patterns, assertion utilities,
// fixture loading, and configuration-driven testing support.
//
// Architecture Compliance:
//   - Domain-Agnostic Design: No module-specific dependencies
//   - Configuration-Driven: No hardcoded paths or values
//   - Fixture-Based Testing: Edit fixtures to affect all tests
//   - Progressive Migration: Support for gradual test infrastructure updates
//   - Universal Constants: Shared timeouts and test values across modules
//
// Key Components:
//   - UniversalTestSetup: Common test setup and teardown patterns
//   - Fixture Loading: Configuration-driven test fixture management
//   - Directory Management: Temporary directory creation and cleanup
//   - Assertion Helpers: Domain-agnostic validation utilities
//   - Progressive Readiness: Event-based test readiness coordination
//   - Universal Constants: Standardized timeouts and performance thresholds
//
// Test Utilities:
//   - SetupTest(): Universal test setup with configuration and logging
//   - TeardownTest(): Cleanup with directory and resource management
//   - LoadFixture(): Configuration-driven fixture loading
//   - CreateTempDir(): Temporary directory management
//   - Assert helpers for common validation patterns
//
// Design Principles:
//   - No module-specific dependencies for universal reusability
//   - Configuration-driven approach (no hardcoded paths)
//   - Fixture-based testing for maintainable test data
//   - Progressive migration support for gradual adoption
//   - Magic number elimination with universal constants
//
// Requirements Coverage:
//   - REQ-TEST-001: Universal test setup and teardown
//   - REQ-TEST-002: Configuration-driven directory management
//   - REQ-TEST-003: Standardized fixture loading
//   - REQ-TEST-004: Domain-agnostic assertion utilities
//   - REQ-TEST-005: Universal constants and timeouts
//
// Test Categories: Unit/Integration
// API Documentation Reference: docs/testing.md
package testutils
