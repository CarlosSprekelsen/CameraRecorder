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

1. **No-mock guard (test runtime)** - Add to `tests/conftest.py`:
```python
# Enable by running: FORBID_MOCKS=1 pytest ...
import os, pytest, unittest.mock as um, sys

def _ban(*a, **k): 
    raise RuntimeError("Mocking forbidden in this lane (FORBID_MOCKS=1).")

def pytest_configure(config):
    if os.getenv("FORBID_MOCKS") == "1":
        # Complete mock blocking including mock_open
        forbidden_mock_module = type('MockModule', (), {
            'Mock': _ban,
            'MagicMock': _ban,
            'AsyncMock': _ban,
            'patch': _ban,
            'mock_open': _ban,  # Essential inclusion
            'create_autospec': _ban,
        })
        sys.modules['unittest.mock'] = forbidden_mock_module

def pytest_collection_modifyitems(config, items):
    """Precise directory-based marker assignment"""
    for item in items:
        file_path = str(item.fspath)
        # Precise path matching instead of string contains
        if "/prototypes/" in file_path:
            item.add_marker(pytest.mark.pdr)
        if "/contracts/" in file_path:
            item.add_marker(pytest.mark.integration)
        if "/ivv/" in file_path:
            item.add_marker(pytest.mark.ivv)

@pytest.fixture(autouse=True)
def _forbid_mocker(request):
    if os.getenv("FORBID_MOCKS") == "1" and "mocker" in request.fixturenames:
        raise RuntimeError("pytest-mock fixture forbidden.")
```

2. **Markers + CI gate (policy)** - Required implementation:
   - In `pytest.ini` define: `unit`, `integration`, `ivv`, `pdr`
   - Required CI job: `FORBID_MOCKS=1 pytest -m "integration or ivv or pdr" -q`
   - CI grep guard: `! grep -R "AsyncMock\|MagicMock\|[^A-Za-z]Mock\(|patch\(|aioresponses" tests/prototypes tests/contracts tests/ivv || (echo "mocks found"; exit 1)`

3. **Lint fence (static)** - Ruff/flake8 per-dir rule: disallow `unittest.mock` / `pytest-mock` imports under `tests/prototypes/**`, `tests/contracts/**`, `tests/ivv/**`

**Test Execution Commands:**
- Unit tests (mocks allowed): `python3 -m pytest tests/unit/ -v`
- PDR tests (no mocks): `FORBID_MOCKS=1 python3 -m pytest -m "pdr" -v`
- Integration tests (no mocks): `FORBID_MOCKS=1 python3 -m pytest -m "integration" -v`
- IV&V tests (no mocks): `FORBID_MOCKS=1 python3 -m pytest -m "ivv" -v`

**Waiver Rule:** Only external systems truly out of project control may be mocked, via documented allow-list and PR-level approval.

---

## Phase 0: Design Baseline

### 0-pre. PDR Entry Baseline (Project Manager)

```
Your role: Project Manager
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md
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

### 0. Critical Prototype Implementation (Developer)

```
Your role: Developer
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md
Task: Implement critical prototypes proving design implementability against real systems.

Execute exactly:
1. Implement critical design prototypes with real MediaMTX integration
2. Set up real RTSP stream handling with actual camera feeds or simulators
3. Implement core API endpoints with real aiohttp and actual request processing
4. Execute prototype validation: FORBID_MOCKS=1 pytest -m "pdr" tests/prototypes/ -v
5. Unit tests (informational only): pytest tests/unit/ -v
6. Capture prototype execution logs and real system interaction evidence

Create: evidence/pdr-actual/00_critical_prototype_implementation.md

Deliverable Criteria:
- Critical prototypes implemented with real system integration
- Real MediaMTX connection operational and tested
- Real RTSP stream processing functional
- Core API endpoints responding to real requests
- PDR prototype tests passing in no-mock environment
- Unit tests informational (not gating)

Success Criteria: Critical prototypes prove design implementability through real system execution.
```

### 0a. Prototype Implementation Validation (IV&V)

```
Your role: IV&V
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md
Input: 00_critical_prototype_implementation.md
Task: Validate prototype implementations through independent no-mock testing.

Execute exactly:
1. Review Developer's prototype implementations against design specifications
2. Execute independent prototype validation: FORBID_MOCKS=1 pytest -m "ivv" tests/ivv/ -v
3. Verify real system integrations operational (MediaMTX, RTSP streams)
4. Execute contract test validation: FORBID_MOCKS=1 pytest -m "integration" tests/contracts/ -v
5. Validate prototype meets basic implementability criteria
6. Identify implementation gaps requiring real system improvements

Create: evidence/pdr-actual/00a_prototype_implementation_review.md

Deliverable Criteria:
- Independent validation tests passing in no-mock environment
- Real system integrations verified operational
- Contract tests passing against real endpoints
- Implementation gap analysis with specific findings
- Evidence from real system execution

No-Mock Enforcement: All IV&V tests executed with FORBID_MOCKS=1.

Success Criteria: Prototype implementation validated through independent no-mock testing.
```

### 0d. Implementation Remediation Sprint (PM, Developer, IV&V)

```
Your role: Project Manager (lead); Developer (implements); IV&V (validates)
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md
Input: 00a_prototype_implementation_review.md with identified issues
Objective: Generate prompts to resolve implementation gaps via real system improvements.
Timebox: 48h (+ optional 24h mop-up)

Execute exactly:
1. Extract findings and assign GAP IDs prioritizing real implementation issues
2. Generate Developer prompts focusing on real system integration improvements
3. Generate IV&V prompts for no-mock validation of fixes
4. Require all test execution use FORBID_MOCKS=1 environment
5. Create remediation checklist tracking real implementation improvements

CRITICAL CONSTRAINTS:
- All fixes must improve real implementations, not mocks
- All test validation must use FORBID_MOCKS=1
- Mock fixes are PROHIBITED - address underlying implementation issues
- External system mocks require documented waiver and PM approval

Output Format:

PROMPT 1: Developer Real Implementation Fixes
Your role: Developer
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md
Task: [specific real implementation improvements]
Execute exactly: [real system improvement steps]
Validation: FORBID_MOCKS=1 pytest -m "pdr" [test area] -v
Create: [evidence file]
Success Criteria: [no-mock test validation]

PROMPT 2: IV&V No-Mock Validation
Your role: IV&V
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md
Input: [Developer evidence]
Task: [validate real implementation improvements]
Execute exactly: [no-mock validation steps]
Validation: FORBID_MOCKS=1 pytest -m "ivv" [test area] -v
Create: [validation evidence]
Success Criteria: [real system validation criteria]

Create: evidence/pdr-actual/00d_implementation_remediation_sprint.md

Success Criteria: Remediation prompts generated enforcing no-mock validation.
```

### 0e. Implementation Baseline (Project Manager)

```
Your role: Project Manager
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md
Objective: Freeze working implementation baseline with no-mock validation.

Execute exactly:
1. Verify all PDR tests passing: FORBID_MOCKS=1 pytest -m "pdr or integration or ivv" -v
2. Verify real system integrations operational
3. Commit implementation changes to pdr-working-vX.Y branch
4. Tag implementation baseline: git tag -a pdr-baseline-vX.Y -m "PDR implementation baseline"
5. Push baseline tag: git push origin pdr-baseline-vX.Y

Create: evidence/pdr-actual/00e_implementation_baseline.md

Gate: Phase 1 cannot start without pdr-baseline-vX.Y tag and 100% no-mock PDR test pass rate.

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
2. Execute API contract validation: FORBID_MOCKS=1 pytest -m "pdr" tests/contracts/ -v
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
- All tests executed with FORBID_MOCKS=1

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
2. Execute performance sanity tests: FORBID_MOCKS=1 pytest -m "pdr" tests/performance/ -v
3. Measure response times under light representative load (not stress testing)
4. Validate core operations meet PDR performance budget
5. Test basic resource usage (CPU/memory) under normal operation
6. Capture performance measurements from real system execution

Create: evidence/pdr-actual/02_performance_sanity_testing.md

Deliverable Criteria:
- Basic performance tests implemented for critical paths
- Performance measurements under light representative load
- PDR budget validation (not full performance compliance)
- Resource usage measurements under normal operation
- Performance evidence from real system execution

PDR Performance Scope:
- Light load testing (not stress/endurance)
- Basic response time validation
- Sanity check against PDR budget targets
- NO full performance compliance (CDR scope)

Success Criteria: Performance sanity validated through basic testing against real system.
```

### 3. Security Design Testing (Developer)

```
Your role: Developer
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md
Task: Validate security design through basic authentication flow testing.

Execute exactly:
1. Implement basic authentication/authorization flow tests
2. Execute security design validation: FORBID_MOCKS=1 pytest -m "pdr" tests/security/ -v
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
- Basic auth/authz flow validation
- Real token/credential testing
- NO penetration testing (CDR scope)
- NO attack simulation (CDR scope)
- NO full security lifecycle testing (CDR scope)

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
1. Execute full PDR test suite: FORBID_MOCKS=1 pytest -m "pdr or integration or ivv" -v
2. Validate all real system integrations operational
3. Verify contract, performance, and security tests passing without mocks
4. Validate system meets PDR acceptance criteria (not full CDR criteria)
5. Assess readiness for Phase 2 or need for additional remediation

Create: evidence/pdr-actual/03a_integration_validation_gate.md

Decision: PROCEED | REMEDIATE | CONDITIONAL | HALT

PDR Gate Criteria:
- 100% PDR-scope no-mock tests passing
- Real system integrations operational
- Basic performance sanity validated
- Security design functional
- NO requirement for full system compliance (CDR scope)

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
Task: Make PDR authorization decision based on no-mock validation results.

Execute exactly:
1. Review IV&V technical assessment results
2. Verify all PDR acceptance criteria met through no-mock testing
3. Assess design implementability evidence from real prototypes
4. Evaluate readiness for CDR phase based on PDR scope completion
5. Make authorization decision with supporting rationale

Create: evidence/pdr-actual/07_pdr_authorization_decision.md

Decision: AUTHORIZE | CONDITIONAL | DENY based on PDR-scope validation.

PDR Authorization Criteria:
- Critical prototypes demonstrate implementability
- Interface contracts validated against real systems
- Basic performance sanity confirmed
- Security design functional
- Build pipeline with no-mock CI integration operational

Success Criteria: PDR authorization decision based on validated design implementability.
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
- ✅ Unit tests: `python3 -m pytest tests/unit/ -v` (mocks allowed, informational only)
- ✅ PDR tests: `FORBID_MOCKS=1 python3 -m pytest -m "pdr" -v` (no mocks, gating)
- ✅ Integration tests: `FORBID_MOCKS=1 python3 -m pytest -m "integration" -v` (no mocks, gating)  
- ✅ IV&V tests: `FORBID_MOCKS=1 python3 -m pytest -m "ivv" -v` (no mocks, gating)
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