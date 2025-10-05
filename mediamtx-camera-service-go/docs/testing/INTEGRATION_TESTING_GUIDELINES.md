# Integration Testing Guidelines - AI Instructions

**Version:** 1.0  
**Date:** 2025-01-15  
**Status:** Active - Enterprise Grade Standards  
**Location:** `tests/integration/`  
**Audience:** AI code generation systems

---

## Scope Definition

Integration tests validate component integration within application boundaries. Test that components work together correctly, not complete user workflows.

**Integration Test Scope:**
- Component interfaces communicate correctly
- Data flows between components as designed
- Dependencies wire together properly
- Error handling propagates through layers correctly
- Security integrations function at boundaries

**Explicitly Out of Scope:**
- Complete user workflows - belongs in tests/e2e
- Business logic internals - belongs in unit tests
- Performance benchmarking - belongs in performance tests
- Individual function behavior - belongs in unit tests

---

## File Organization Mandates

**Directory Structure Requirements:**
- Location: tests/integration/ directory only
- Maximum file size: 200 lines absolute limit, 150 lines preferred
- One functional area per file - no mixing concerns
- File naming: component_area_test.go pattern required

**File Naming Enforcement:**
- Pattern: {component}_{functional_area}_test.go
- Valid: websocket_protocol_test.go, mediamtx_client_test.go
- Invalid: test_websocket.go, websocket_integration.go, websocket_test_v2.go, websocket_real_test.go

**File Split Triggers:**
- Exceeds 200 lines - split immediately
- Covers multiple functional areas - split by area
- Different setup requirements - separate files
- Tests multiple components - split by component

**Splitting Strategy:**
- Extract table-driven tests to compress similar cases
- Create domain asserters for repeated validation patterns
- Group by component boundary not arbitrary logic
- One test group per functional integration point

---

## Mandatory Infrastructure

**Setup Infrastructure - No Exceptions:**
- Use testutils.SetupTest with config fixture - always
- Build component helpers on top of universal setup - never bypass
- Use defer for cleanup in every test function
- Never create custom setup outside testutils
- Never use global shared state between tests
- Never initialize components directly

**Timeout Management - No Exceptions:**
- Use testutils.DefaultTestTimeout for standard operations
- Use testutils.UniversalTimeoutVeryLong for extended operations
- Never hardcode timeout durations
- Never use time.Sleep for synchronization
- Always use context.WithTimeout from testutils constants
- Use testutils.WaitForCondition for async readiness

**Validation Infrastructure - No Exceptions:**
- Use testutils.DataValidationHelper for all file operations
- Use testutils.DataValidationHelper for state validation
- Never perform direct os.Stat or file system operations
- Never create custom validation functions
- Always validate through approved helper methods

---

## Test Organization Patterns

**Table-Driven Test Mandate:**
- Use for multiple input variations
- Use for different error conditions
- Use for role-based access scenarios
- Use for protocol compliance cases
- Structure: single test function with slice of test cases
- Execution: iterate with t.Run subtests
- Setup: shared setup for all cases in table
- Validation: consistent validation pattern across cases

**Subtest Organization:**
- Use for sequential operations with shared state
- Use for testing different aspects of same feature
- Use for clear logical grouping of related tests
- Never mix unrelated test scenarios in subtests
- Always use descriptive subtest names
- Maintain test isolation between subtests

**Test Function Structure Requirements:**
- Start with requirements comment: REQ-XXX-NNN
- Setup using testutils.SetupTest first
- Create component helpers second
- Register cleanup with defer third
- Execute test logic fourth
- Validate results explicitly fifth
- Never skip validation step

---

## Coverage Measurement

**Coverage Execution Command:**
- Command: go test ./tests/integration/... -coverpkg=./internal/... -coverprofile=coverage/integration/integration.out -v
- Coverage analysis: go tool cover -func=coverage/integration/integration.out
- HTML report: go tool cover -html=coverage/integration/integration.out
- Per-package view required for component boundary analysis

**Enterprise Coverage Targets:**
- Overall integration coverage: 85% minimum required
- Critical path coverage: 95% minimum required
- Component boundary coverage: 90% minimum required
- Error handling path coverage: 90% minimum required

**Coverage Validation Criteria:**
- Component integration points actually tested
- Data flow between components actually validated
- Error propagation actually verified
- Security boundaries actually enforced
- Tests with assertions count - tests without assertions do not count
- Real validation counts - checking only for no error does not count

**Coverage Exclusions:**
- Tests checking only error absence without result validation
- Tests without explicit assertions on outcomes
- Tests not validating integration actually occurred
- Tests passing without verifying component interaction

---

## Anti-Pattern Prevention

**Fake Passing Test Prevention:**
- Never test only for error absence without validating results
- Always validate result structure and content
- Always verify integration actually occurred
- Always check state changes happened correctly
- File creation tests must verify file exists with correct size and content
- Authentication tests must verify subsequent operations succeed or fail correctly
- Never consider test passing without explicit outcome validation

**Public API Pollution Prevention:**
- Test through public API only - never make internal functions public for testing
- If test requires internal access then test design is incorrect
- If test requires internal access then architecture needs refactoring
- Never add exported functions solely for test access
- Never add test-specific methods to production code

**Global State Prevention:**
- Never use package-level variables for test state
- Never share asserters between test functions
- Create fresh instances per test function with defer cleanup
- Ensure complete test isolation from other tests
- Each test must be runnable independently in any order

**Hardcoded Value Prevention:**
- Never hardcode timeout values - use testutils constants
- Never hardcode file paths - use testutils helpers
- Never hardcode error codes - reference from specification
- Never hardcode configuration - use fixtures

---

## Requirements Traceability

**File Header Mandate:**
- Every test file requires documentation header
- Header must list component name and purpose
- Header must enumerate requirements covered with REQ-XXX-NNN format
- Header must specify test category as Integration
- Header must reference API documentation source
- Header must list test organization with line references

**Test Function Requirements:**
- Every test function name must include requirement: TestComponent_Feature_ReqXXXNNN
- First line must be comment with requirement: REQ-XXX-NNN description
- Test must fail if requirement violated
- Test must validate requirement compliance explicitly
- Multiple requirements per test allowed with multiple comment lines

---

## Proven Patterns from WebSocket

**Domain Asserter Pattern - Mandatory Reuse:**
- Create ComponentIntegrationAsserter type per component
- Encapsulate complex integration validation logic in asserter
- Asserter built on testutils.SetupTest foundation
- Asserter provides Cleanup method called via defer
- Asserter methods return error for test to handle
- Reduces test code duplication significantly
- Makes tests focus on what to validate not how to validate

**Progressive Readiness Pattern - Mandatory for Async:**
- Never use arbitrary time delays for readiness
- Always use testutils.WaitForCondition for async operations
- Specify timeout from testutils constants
- Specify poll interval explicitly
- Provide condition function checking readiness
- Provide descriptive name for condition being awaited
- Handle timeout as test failure with clear message

**Asserter Benefits to Preserve:**
- Eliminates massive code duplication
- Encapsulates complex validation sequences
- Provides reusable integration patterns
- Enables consistent validation across tests
- Reduces test file line counts dramatically
- Makes tests readable and maintainable

---

## File Size Management

**Line Count Triggers:**
- 200 lines: immediate split required
- 150 lines: consider splitting for clarity
- Prefer smaller focused files over large comprehensive files

**Splitting Methodology:**
- Split by functional area not arbitrary line counts
- Extract tables to compress similar test cases
- Move complex validation to asserters
- Separate by component boundary
- Group by integration point being tested

**Organization After Split:**
- Original file name indicates primary function
- Split files indicate specific functional area
- Each file remains independently executable
- Shared setup moves to helper package
- Asserters shared across split files via test helpers

---

## Quality Gates

**Pre-Merge Validation Required:**
- All integration tests pass locally
- Coverage meets minimum thresholds per target
- No hardcoded timeouts anywhere in test code
- No global state between tests
- All tests use testutils.SetupTest for initialization
- All file operations use DataValidationHelper
- All tests have requirements traceability
- No fake passing tests present

**Continuous Validation:**
- Coverage trends upward toward enterprise targets
- Test execution time remains acceptable
- No test flakiness from timing issues
- Tests remain isolated and independent
- File sizes stay under limits

---

## Command Reference

**Execute Integration Tests:**
- Run all: go test ./tests/integration/... -v
- Run with coverage: go test ./tests/integration/... -coverpkg=./internal/... -coverprofile=coverage/integration/integration.out -v
- Run specific file: go test ./tests/integration/websocket_protocol_test.go -v
- Run specific test: go test ./tests/integration/... -run TestWebSocket_Protocol -v

**Analyze Coverage:**
- Function view: go tool cover -func=coverage/integration/integration.out
- HTML report: go tool cover -html=coverage/integration/integration.out
- Package view: go tool cover -func=coverage/integration/integration.out | grep "package-name"

**Common Issues Resolution:**
- Test timeout: increase timeout using testutils constant not hardcoded value
- Test flaky: replace time.Sleep with testutils.WaitForCondition
- Coverage low: add tests for uncovered integration points
- File too large: split by functional area following guidelines

---

## Success Criteria Summary

Integration tests are correct when:
- File sizes under 200 lines each
- Tests organized by functional area clearly
- All use testutils infrastructure exclusively
- Table-driven patterns compress similar cases
- Domain asserters reduce duplication
- Coverage meets 85% minimum overall
- No fake passing tests exist
- Requirements traceable in every test
- Tests isolated without shared state
- All validation explicit and meaningful
