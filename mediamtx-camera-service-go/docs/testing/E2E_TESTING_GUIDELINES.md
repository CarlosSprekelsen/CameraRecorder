# E2E Testing Guidelines - AI Instructions

**Version:** 1.0  
**Date:** 2025-01-15  
**Status:** Active - Enterprise Grade Standards  
**Location:** `tests/e2e/`  
**Audience:** AI code generation systems

---

## Scope Definition

E2E tests validate complete user workflows from start to finish across all system boundaries. Test real user scenarios that cross multiple components and verify actual business outcomes.

**E2E Test Scope:**
- Complete user workflows from beginning to end
- Real file creation and verification on disk
- Actual state changes in system verified
- Cross-component interactions validated
- Business outcomes achieved and verified
- Real external service interactions where applicable

**Explicitly Out of Scope:**
- Component integration details - belongs in tests/integration
- Internal function behavior - belongs in unit tests
- Performance benchmarking - belongs in performance tests
- Component boundary testing - belongs in integration tests

---

## Critical Distinction from Integration Tests

**Integration Tests Answer:** Do our components wire together correctly?
- Focus: Component interfaces and boundaries
- Scope: Within application code boundaries
- Speed: Faster execution
- Validation: Interface contracts and data flow

**E2E Tests Answer:** Can a user accomplish their complete goal?
- Focus: Complete user journeys and outcomes
- Scope: Full system including external interactions
- Speed: Slower due to comprehensive validation
- Validation: Actual files on disk, real state changes, business outcomes

**Example Contrast:**
- Integration: Does authenticate method correctly call JWTHandler and store session?
- E2E: Can user connect, authenticate, start recording, stop recording, and find video file on disk?

---

## File Organization Mandates

**Directory Structure Requirements:**
- Location: tests/e2e/ directory only
- Maximum file size: 300 lines absolute limit per workflow file
- One complete workflow category per file
- File naming: {feature}_workflows_test.go pattern required

**File Naming Enforcement:**
- Pattern: {feature}_workflows_test.go
- Valid: camera_workflows_test.go, recording_workflows_test.go, snapshot_workflows_test.go
- Invalid: test_camera_e2e.go, e2e_camera.go, camera_test_v2.go, camera_e2e_real.go

**Workflow Categories - One File Each:**
- camera_workflows_test.go - Camera discovery, status, management workflows
- recording_workflows_test.go - Start, stop, list, verify recording file workflows
- snapshot_workflows_test.go - Capture, list, verify snapshot file workflows  
- health_workflows_test.go - Health monitoring, metrics collection workflows
- security_workflows_test.go - Authentication, authorization, session workflows

**File Split Triggers:**
- Exceeds 300 lines - split workflow categories
- Mixes unrelated workflows - separate by user goal
- Different user personas - split by role workflows
- Different system states - split by initial conditions

**Splitting Strategy:**
- Each file represents cohesive set of user goals
- Workflows within file share common setup requirements
- Split by user role if workflows differ significantly by permission level
- Group workflows that validate same business capability

---

## Mandatory Infrastructure

**Setup Infrastructure - No Exceptions:**
- Use testutils.SetupTest with e2e-specific config fixture
- Build E2E helpers on top of universal setup
- Use defer for cleanup ensuring files removed after test
- Never create custom setup outside testutils
- Never use global shared state between workflow tests
- Never initialize system components directly

**Timeout Management - No Exceptions:**
- Use testutils.UniversalTimeoutVeryLong for workflow operations
- Use testutils.UniversalTimeoutExtreme for multi-step workflows
- Never hardcode timeout durations anywhere
- Never use time.Sleep for workflow synchronization
- Always use context.WithTimeout from testutils constants
- Use testutils.WaitForCondition for async operations in workflows

**Validation Infrastructure - No Exceptions:**
- Use testutils.DataValidationHelper for all file verification
- Verify files exist with correct size and accessible permissions
- Verify state changes through system query not assumptions
- Never assume success without explicit verification
- Never skip outcome validation steps
- Always validate business outcome achieved

---

## Workflow Test Structure

**Complete Workflow Pattern:**
Each E2E test validates one complete user workflow from start to finish with explicit outcome verification.

**Workflow Template Structure:**
- Workflow name clearly describes user goal
- Setup creates clean initial state using testutils
- Execute workflow steps sequentially
- Verify outcome at each critical step
- Validate final business outcome achieved
- Cleanup removes all test artifacts

**Workflow Step Requirements:**
- Each step represents user action or system response
- Validate step succeeded before proceeding to next
- Use progressive readiness for async steps
- Log workflow progress for debugging
- Handle errors explicitly at each step
- Never continue workflow after step failure

**Outcome Validation Requirements:**
- File workflows must verify file exists on disk with correct properties
- State workflows must query system state and verify change occurred
- Authentication workflows must verify subsequent operations succeed or fail correctly
- Multi-component workflows must verify data propagated correctly across boundaries

---

## Workflow Categories

**Camera Workflow Requirements:**
- Discovery workflow: Connect, discover cameras, verify list contents
- Status workflow: Connect, authenticate, get camera status, verify status fields
- Management workflow: Connect, authenticate, list cameras, get specific camera, verify data consistency

**Recording Workflow Requirements:**
- Basic recording: Connect, auth, start recording, wait duration, stop recording, verify file exists with minimum size
- Long recording: Connect, auth, start recording, wait extended duration, verify file growing, stop recording, verify final file
- List recordings: Connect, auth, create recordings, list recordings, verify all recordings present in list
- Multiple cameras: Connect, auth, start recording on multiple cameras simultaneously, verify all files created

**Snapshot Workflow Requirements:**
- Single snapshot: Connect, auth, take snapshot, verify file exists with image format
- Multiple snapshots: Connect, auth, take multiple snapshots, verify all files exist with unique names
- List snapshots: Connect, auth, create snapshots, list snapshots, verify all present
- Snapshot metadata: Connect, auth, take snapshot, verify metadata includes camera info and timestamp

**Health Workflow Requirements:**
- System health: Connect, auth, get system health, verify all components reporting
- Metrics collection: Connect, auth, get metrics, verify metric structure and values
- Health monitoring: Connect, auth, monitor health over time, verify health changes detected

**Security Workflow Requirements:**
- Authentication: Connect, authenticate with valid token, verify authenticated operations succeed
- Authorization: Connect, authenticate with different roles, verify role permissions enforced
- Session management: Connect, authenticate, perform operations, verify session maintained, logout, verify session ended
- Token expiry: Connect, authenticate, wait for expiry, verify operations fail after expiry

---

## Coverage Measurement

**Coverage Execution Command:**
- Command: go test ./tests/e2e/... -coverpkg=./internal/... -coverprofile=coverage/e2e/e2e.out -v -timeout=30m
- Extended timeout required for complete workflows
- Coverage analysis: go tool cover -func=coverage/e2e/e2e.out
- HTML report: go tool cover -html=coverage/e2e/e2e.out

**Enterprise Coverage Targets:**
- Overall E2E coverage: 75% minimum required
- Critical user workflows: 100% coverage required
- Happy path workflows: 100% coverage required
- Error recovery workflows: 90% coverage required

**Coverage Validation Criteria:**
- Complete user workflows tested end-to-end
- Actual file creation verified on disk
- Real state changes verified in system
- Business outcomes validated explicitly
- Cross-component data flow verified
- Error recovery paths validated

**Coverage Exclusions:**
- Workflows checking only operation completion without outcome verification
- Workflows not verifying files actually created
- Workflows not validating state actually changed
- Workflows assuming success without explicit validation

---

## Anti-Pattern Prevention

**Incomplete Workflow Prevention:**
- Never test only part of user workflow
- Always complete full workflow from user perspective
- Never skip outcome verification steps
- Always verify business goal achieved
- File creation workflows must verify file exists with correct content
- State change workflows must query state and verify change occurred

**Fake Outcome Prevention:**
- Never assume file created without verifying file exists on disk
- Never assume state changed without querying state after operation
- Never assume operation succeeded without validating outcome
- Never trust return value without verifying actual result
- Recording workflows must verify video file exists with size greater than threshold
- Snapshot workflows must verify image file exists with valid image format

**Incomplete Cleanup Prevention:**
- Always remove test files created during workflow
- Always reset system state after workflow completes
- Use defer to ensure cleanup happens even on test failure
- Verify cleanup succeeded before test ends
