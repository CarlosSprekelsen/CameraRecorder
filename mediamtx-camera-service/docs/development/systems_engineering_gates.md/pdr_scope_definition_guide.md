# PDR (Preliminary Design Review) – Scope Definition and Execution Guide (Enforced No-Mock)

**Purpose:** Ensure the detailed design is implementable by executing critical prototypes and contract tests against real systems, with technical enforcement preventing mock-based validation.

## PDR Objective

Validate detailed system design implementability through critical prototypes, interface contract tests, and basic performance sanity checks against real systems. Convert findings into merged changes addressing actual implementation issues.

## PDR Scope (NOT CDR/ORR)

**✅ PDR-appropriate:**
- Critical prototypes proving implementability (real MediaMTX, real RTSP streams)
- Interface contract tests against real endpoints (basic success/error paths)
- Initial performance sanity vs PDR budget (short representative load, not endurance)
- Security design completion + basic auth flow exercised (not full pen-test)
- CI green for build + no-mock integration lane
- Evidence package from real runs

**❌ CDR/ORR scope (NOT in PDR):**
- Full load/stress/endurance testing, sustained load, scaling
- Penetration testing, attack simulation, full security lifecycle testing
- Operational readiness (runbooks, SLOs, backup/restore, monitoring integration)
- Deployment automation on target infrastructure, rollback drills
- System-wide performance compliance with >95% coverage
- API version freezing based on exhaustive versioning tests

## No-Mock Enforcement

**Required Directory Structure:**
```
tests/
├── unit/          # Mocks allowed, informational
├── prototypes/    # No mocks, PDR gating
├── contracts/     # No mocks, integration gating  
└── ivv/          # No mocks, IV&V gating
```

**Technical Guardrails (Required Implementation):**

1. **No-mock guard (test runtime)** - Implement runtime mock blocking in `tests/conftest.py` with complete mock enumeration including mock_open and create_autospec

2. **Markers + CI gate (policy)** - Required implementation:
   - Configure pytest.ini with markers: `unit`, `integration`, `ivv`, `pdr`
   - Implement required CI job for no-mock validation
   - Add CI grep guard to prevent mock imports in restricted directories

3. **Lint fence (static)** - Configure static analysis to disallow unittest.mock imports in PDR test directories

**Test Execution Commands:**
- Unit tests (mocks allowed): Standard unit test execution
- PDR tests (no mocks): Execute with mock prohibition environment
- Integration tests (no mocks): Execute with mock prohibition environment  
- IV&V tests (no mocks): Execute with mock prohibition environment

**Waiver Rule:** Only external systems truly out of project control may be mocked, via documented allow-list and PR-level approval.

---

# Phase 0 — System Readiness for PDR (No‑Mock, Real Services)

## 0-pre. PDR Entry Baseline (Project Manager)

```
Your role: Project Manager
Task: Establish PDR entry baseline and no-mock enforcement.

Execute exactly:
1. Verify main branch is clean and up-to-date
2. Create PDR entry tag: git tag -a pdr-entry-vX.Y -m "PDR entry baseline"
3. Create PDR working branch: git checkout -b pdr-working-vX.Y
4. Add no-mock enforcement to tests/conftest.py per technical guardrails
5. Configure pytest.ini with markers: pdr, integration, ivv
6. Push entry tag: git push origin pdr-entry-vX.Y

Create: evidence/pdr-actual/00-pre_pdr_entry_baseline.md
Success Criteria: PDR baseline established with no-mock enforcement technically implemented.
```

---

## 0. System Readiness Validation (IV\&V)

```
Your role: IV&V
Task: Execute a system readiness gate to validate the current implementation status using no-mock tests.

Execute exactly:
1. Run preflight tests to confirm environment readiness (MediaMTX running, ffmpeg installed, /dev/video* available)
2. Execute no-mock PDR gating suite: FORBID_MOCKS=1 pytest -m "pdr or integration or ivv or unit" -v
3. Identify any blockers, failures, or flaky tests; confirm each has a requirement trace
4. Quarantine tests that are non-blocking, low-value, or unrelated to requirements
5. Classify failures: IMPLEMENTATION_GAP, DESIGN_DISCOVERY, TEST_ENVIRONMENT, VALIDATION_THEATER
6. Prepare remediation prompt set for each IMPLEMENTATION_GAP or DESIGN_DISCOVERY

Create: evidence/pdr-actual/00_system_readiness_validation.md
Success Criteria: Readiness confirmed OR blockers identified with clear remediation prompts.
```

---

## 0d. Implementation Remediation Sprint (PM, Developer, IV\&V)

```
Your role: Project Manager (lead); Developer (implements); IV&V (validates)
Input: 00_system_readiness_validation.md with identified issues
Objective: Resolve PDR blockers before baseline freeze.

Execute exactly:
1. Assign remediation actions per blocker classification
2. Implement fixes with no-mock validation after each resolution
3. Verify remediation after changes have been implemented (zero-trust policy).
4. Maintain a remediation checklist and update status until all blockers resolved


Create: evidence/pdr-actual/00d_implementation_remediation_sprint.md
Success Criteria: All gating issues resolved and validated in a no-mock environment.
```

---

## 0e. Implementation Baseline (Project Manager)

```
Your role: Project Manager
Task: Freeze working implementation baseline with no-mock validation.

Execute exactly:
1. Verify all PDR tests passing: FORBID_MOCKS=1 pytest -m "pdr or integration or ivv" -v
2. Verify real system integrations operational
3. Commit implementation changes to pdr-working-vX.Y branch
4. Tag implementation baseline: git tag -a pdr-baseline-vX.Y -m "PDR implementation baseline"
5. Push baseline tag: git push origin pdr-baseline-vX.Y

Create: evidence/pdr-actual/00e_implementation_baseline.md
Success Criteria: Implementation baseline established with no-mock test validation.
```


---

## Phase 1: Interface and Performance Validation

### 1. Interface Contract Testing (Developer)

```
Your role: Developer
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md
Task: Implement and execute interface contract tests against real endpoints.

Execute exactly:
1. Implement contract tests for external APIs using real MediaMTX endpoints
2. Execute API contract validation against real services
3. Test basic success and error paths against real API responses
4. Validate request/response schemas against actual API behavior
5. Test error handling using real error conditions from services
6. Capture contract test evidence from real endpoint interactions

Create: evidence/pdr-actual/01_interface_contract_testing.md

Deliverable Criteria:
- Contract tests implemented for all external interfaces
- Tests passing against real MediaMTX API endpoints
- Basic success/error path validation with real responses
- Schema validation against actual API behavior
- Error condition testing using real service errors

No-Mock Requirements:
- Real MediaMTX API accessible for contract testing
- Actual error conditions injectable from real services
- All tests executed with mock prohibition

Success Criteria: Interface contracts validated through testing against real endpoints.
```

### 2. Performance Sanity Testing (Developer)

```
Your role: Developer
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md
Task: Execute basic performance sanity tests against real system with PDR budget validation.

Execute exactly:
1. Implement basic performance tests for critical paths
2. Execute performance sanity tests with mock prohibition
3. Measure response times under light representative load
4. Validate core operations meet PDR performance budget
5. Test basic resource usage under normal operation
6. Capture performance measurements from real system execution

Create: evidence/pdr-actual/02_performance_sanity_testing.md

Deliverable Criteria:
- Basic performance tests implemented for critical paths
- Performance measurements under light representative load
- PDR budget validation against actual measurements
- Resource usage measurements under normal operation
- Performance evidence from real system execution

PDR Performance Scope:
- Light load testing, not stress or endurance testing
- Basic response time validation
- Sanity check against PDR budget targets
- Full performance compliance reserved for CDR scope

Success Criteria: Performance sanity validated through basic testing against real system.
```

### 3. Security Design Testing (Developer)

```
Your role: Developer
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md
Task: Validate security design through basic authentication flow testing.

Execute exactly:
1. Implement basic authentication and authorization flow tests
2. Execute security design validation with mock prohibition
3. Test basic auth token validation with real tokens
4. Validate basic error handling for invalid credentials
5. Test security configuration in real environment
6. Capture security validation evidence from real auth flows

Create: evidence/pdr-actual/03_security_design_testing.md

Deliverable Criteria:
- Basic auth flow tests implemented
- Authentication working with real tokens and credentials
- Basic authorization validation functional
- Security error handling tested with real invalid inputs
- Security configuration validated in real environment

PDR Security Scope:
- Basic authentication and authorization flow validation
- Real token and credential testing
- Penetration testing reserved for CDR scope
- Attack simulation reserved for CDR scope
- Full security lifecycle testing reserved for CDR scope

Success Criteria: Security design validated through basic auth flow testing against real mechanisms.
```

### 3a. Integration Validation Gate (Project Manager)

```
Your role: Project Manager
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md
Inputs: 01_interface_contract_testing.md, 02_performance_sanity_testing.md, 03_security_design_testing.md
Task: Execute comprehensive integration validation with no-mock enforcement.

Execute exactly:
1. Execute full PDR test suite with mock prohibition
2. Validate all real system integrations operational
3. Verify contract, performance, and security tests passing without mocks
4. Validate system meets PDR acceptance criteria
5. Assess readiness for Phase 2 or need for additional remediation

Create: evidence/pdr-actual/03a_integration_validation_gate.md

Decision: PROCEED | REMEDIATE | CONDITIONAL | HALT

PDR Gate Criteria:
- PDR-scope no-mock tests passing at acceptable rate
- Real system integrations operational
- Basic performance sanity validated
- Security design functional
- Full system compliance reserved for CDR scope

Success Criteria: PDR integration validated through no-mock testing against real systems.
```

---

## Phase 2: Build Integration and Evidence

### 4. Build Pipeline Integration (Developer)

```
Your role: Developer
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md
Task: Validate build pipeline with no-mock integration lane.

Execute exactly:
1. Execute automated build pipeline: make build && make test
2. Execute CI pipeline with no-mock gate: FORBID_MOCKS=1 pytest -m "integration or pdr" -v
3. Verify CI integration with no-mock enforcement operational
4. Test basic build reproducibility (single environment)
5. Capture build pipeline evidence

Create: evidence/pdr-actual/04_build_pipeline_integration.md

Deliverable Criteria:
- Build pipeline executing successfully
- CI no-mock gate passing consistently
- Basic build reproducibility in single environment
- CI integration with no-mock enforcement functional

PDR Build Scope:
- Basic build pipeline validation
- CI integration with no-mock testing
- Single environment reproducibility check

Success Criteria: Build pipeline validated with no-mock CI integration.
```

### 5. Evidence Package Validation (IV&V)

```
Your role: IV&V
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md
Task: Validate evidence package completeness from real system validation.

Execute exactly:
1. Execute comprehensive PDR validation: FORBID_MOCKS=1 pytest -m "ivv or pdr" -v
2. Verify all evidence from real system execution
3. Validate prototype implementability evidence
4. Verify contract test evidence against real endpoints
5. Validate performance evidence from real system measurements
6. Confirm security evidence from real authentication flows

Create: evidence/pdr-actual/05_evidence_package_validation.md

Deliverable Criteria:
- PDR-scope tests passing in no-mock environment
- Evidence package complete from real system execution
- Prototype implementability demonstrated
- Contract validation against real endpoints confirmed
- Performance evidence from real measurements
- Security validation from real authentication

Success Criteria: Evidence package validated through no-mock PDR-scope testing.
```

---

## Phase 3: PDR Decision

### 6. PDR Technical Assessment (IV&V)

```
Your role: IV&V
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md
Task: Execute final PDR technical assessment through no-mock validation.

Execute exactly:
1. Execute complete PDR validation suite: FORBID_MOCKS=1 pytest -m "pdr or integration or ivv" -v
2. Assess design implementability through prototype evidence
3. Validate interface contracts through real endpoint testing
4. Assess basic performance through sanity testing
5. Validate security design through basic auth testing
6. Compile technical assessment based on real system evidence

Create: evidence/pdr-actual/06_pdr_technical_assessment.md

Outcome: Recommendation = PROCEED | CONDITIONAL | DENY based on no-mock test results.

PDR Assessment Criteria:
- Design implementability demonstrated through prototypes
- Interface contracts validated against real endpoints
- Basic performance sanity confirmed
- Security design functional
- All validation through no-mock testing

Success Criteria: PDR technical assessment completed through no-mock validation.
```

### 7. PDR Authorization Decision (Project Manager)

```
Your role: Project Manager
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md
Input: 06_pdr_technical_assessment.md
Task: Make PDR authorization decision based on no-mock validation results with validation theater prevention.

Execute exactly:
1. Review IV&V technical assessment for actual execution evidence
2. Verify PDR acceptance criteria met through no-mock testing
3. Assess design implementability evidence from real prototypes
4. Validate MediaMTX FFmpeg integration with concrete stream evidence
5. Make authorization decision with supporting rationale

VALIDATION THEATER PREVENTION CONTROLS:
- Verify IV&V performed independent test execution with concrete results
- Reject test failure dismissals without technical evidence
- Verify real resource utilization in testing validation
- Demand root cause analysis for failures with available resources
- Require documented proof of working real system integration
- Validate RTSP stream accessibility as integration proof

VALIDATION THEATER RED FLAGS:
- Implementation claims without execution results
- Readiness assertions without actual test pass/fail counts
- Test skips when real resources available
- Integration claims without functional stream proof
- Normal failure excuses for obvious implementation issues

Authorization Checklist:
- ✅ IV&V independent test execution with concrete results
- ✅ Test results include actual pass/fail/skip counts
- ✅ MediaMTX FFmpeg integration proven through accessible streams
- ✅ Test failures have root cause analysis or technical waivers
- ✅ Normal failure claims supported by technical evidence
- ✅ Real system resources utilized in validation testing
- ✅ Working implementations verified through functional testing

MEDIAMTX INTEGRATION VALIDATION REQUIREMENTS:
- MediaMTX API path creation operational
- FFmpeg integration functional for camera streaming
- RTSP streams accessible for detected cameras
- Automatic discovery to streaming workflow proven

Create: evidence/pdr-actual/07_pdr_authorization_decision.md

Decision: AUTHORIZE | CONDITIONAL | DENY based on actual working system validation.

PDR AUTHORIZATION CRITERIA:
- Critical prototypes demonstrate implementability through real MediaMTX FFmpeg integration
- Interface contracts validated against real MediaMTX API endpoints
- Basic performance sanity confirmed through real measurements
- Security design functional through real authentication
- Build pipeline with no-mock CI integration operational
- MediaMTX FFmpeg integration working with accessible RTSP streams

Success Criteria: PDR authorization decision based on validated design implementability with zero-trust verification and working MediaMTX integration.
```

---

## Phase 4: PDR Completion

### 8. PDR Completion and Baseline (Project Manager)

```
Your role: Project Manager
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md
Task: Complete PDR with final no-mock validation and baseline creation.

Execute exactly:
1. Execute final PDR validation: FORBID_MOCKS=1 pytest -m "pdr or integration or ivv" -v
2. Organize evidence artifacts from real system execution
3. Create final pull request: pdr-working-vX.Y → main
4. Execute pre-merge validation with no-mock CI gate
5. Merge PDR branch after no-mock validation
6. Tag completion: git tag -a pdr-complete-vX.Y -m "PDR completed - no-mock validation"
7. Update roadmap.md with PDR completion status

Create: evidence/pdr-actual/08_pdr_completion_baseline.md
Update: docs/roadmap.md

Exit Criteria:
- 100% pass for PDR-scope contract & prototype tests in no-mock lane
- PDR working branch merged to main
- Completion tag created: pdr-complete-vX.Y
- Evidence package organized from real system execution
- Roadmap updated with CDR readiness status

Success Criteria: PDR completed with design implementability validated through no-mock PDR-scope testing.
```

---

## Evidence Management

### Document Template

```markdown
# Document Title
**Version:** 1.0
**Date:** YYYY-MM-DD
**Role:** [Developer/IV&V/Project Manager]
**PDR Phase:** [Phase Number]
**Status:** [Draft/Review/Final]

## No-Mock Test Execution Summary

## Real System Evidence

## Implementation Validation

## Conclusion
```

### No-Mock Enforcement Checklist

**Technical Implementation (Required):**
- ✅ `tests/conftest.py` contains enhanced no-mock runtime guard with complete mock blocking
- ✅ Directory structure: `tests/unit/`, `tests/prototypes/`, `tests/contracts/`, `tests/ivv/`
- ✅ Precise marker assignment using directory-based path matching (not string contains)
- ✅ Complete mock enumeration including: Mock, MagicMock, AsyncMock, patch, mock_open, create_autospec
- ✅ `pytest.ini` defines markers: `unit`, `integration`, `ivv`, `pdr`
- ✅ Required CI job runs: `FORBID_MOCKS=1 pytest -m "integration or ivv or pdr" -q`
- ✅ CI grep guard updated for correct directory paths: `tests/prototypes`, `tests/contracts`, `tests/ivv`
- ✅ Ruff/flake8 per-dir rule: disallow `unittest.mock` / `pytest-mock` imports under correct directories

**Test Execution (Gating):**
- ✅ Unit tests: Standard execution with mocks allowed (informational only)
- ✅ PDR tests: Execute with mock prohibition environment (no mocks, gating)
- ✅ Integration tests: Execute with mock prohibition environment (no mocks, gating)  
- ✅ IV&V tests: Execute with mock prohibition environment (no mocks, gating)
- ✅ Real system integrations operational for all PDR testing
- ✅ External system mocks documented in allow-list with PR-level approval

---

## Success Criteria Summary

**PDR Passes When:**
- 100% pass for PDR-scope contract & prototype tests in no-mock lane (FORBID_MOCKS=1)
- Critical prototypes demonstrate design implementability through real system execution
- Interface contracts validated through testing against real endpoints (basic success/error paths)
- Initial performance sanity confirmed through real system measurements (light load, not endurance)
- Security design validated through real authentication flow testing (basic auth, not pen-test)
- Build pipeline operational with no-mock CI integration
- Evidence package complete from real system execution

**PDR Scope Boundaries:**
- ✅ Critical prototypes proving implementability (real MediaMTX, real RTSP streams)
- ✅ Interface contract testing against real endpoints (basic success/error paths)
- ✅ Initial performance sanity vs PDR budget (short representative load)
- ✅ Security design completion + basic auth flow exercised
- ✅ CI green for build + no-mock integration lane
- ✅ Evidence package from real runs
- ❌ Full load/stress/endurance testing (CDR scope)
- ❌ Penetration testing, attack simulation (CDR scope)
- ❌ Operational readiness (runbooks, SLOs, backup/restore) (ORR scope)
- ❌ Deployment automation on target infrastructure (CDR scope)
- ❌ System-wide performance compliance with >95% coverage (CDR scope)
- ❌ API version freezing based on exhaustive versioning tests (CDR scope)

**No-Mock Enforcement:**
- ✅ Technical guardrails prevent mock usage in PDR lanes
- ✅ CI gates enforce no-mock testing for PDR validation
- ✅ Static analysis prevents mock imports in integration/ivv/pdr directories
- ✅ External system mocks require documented waivers with PR-level approval