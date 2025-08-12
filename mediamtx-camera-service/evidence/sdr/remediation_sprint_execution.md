# SDR Remediation Sprint - Role-Based Execution Prompts

## Sprint Overview
**Duration**: 48 hours  
**Scope**: 1 High + 2 Medium priority issues from SDR Phase 1  
**Strategic Decision**: Replace fake MediaMTX with real MediaMTX v1.13.1 for authentic integration testing
**Team**: Developer (execution) + IV&V (validation) + PM (oversight)

---

## Strategic Decision: Real MediaMTX Integration

### **Rationale for Strategic Shift**

**Problem with Fake Servers:**
- Integration testing against fake MediaMTX builds assumption debt
- False confidence in design feasibility and integration approach
- Missed real API behavior, performance characteristics, and error conditions
- SDR/PDR gates compromised by testing against assumptions rather than reality

**Benefits of Real MediaMTX v1.13.1:**
- Authentic integration testing validates design against production-like dependencies
- Real API behavior eliminates assumption debt and integration surprises
- Performance characteristics and error handling proven with actual service
- SDR design feasibility validated with production-grade confidence

**Industry Standard Practice:**
- Integration tests should use real external dependencies when feasible
- Design reviews should validate against production-like environments
- Version freezing against specific dependency versions reduces compatibility risk

**Implementation Approach:**
- Install MediaMTX v1.13.1 from official release
- Configure minimal test environment for reproducible testing
- Remove all fake MediaMTX implementations from test suite
- Document MediaMTX requirements for team reproducibility

---

## Issue SDR-H-001: Security Middleware Permission Fix

### Developer Execution (Priority 1 - Day 1, Hours 1-3)

```
Your role: Developer
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md

Issue: SDR-H-001 - Security Middleware Permission Issue
Source: evidence/sdr-actual/01_architecture_feasibility_demo.md (Line 99)
Severity: HIGH - Security functionality affected

Problem: Permission denied for `/opt/camera-service/keys` directory, limiting security middleware functionality

Execute exactly:
1. Create remediation branch: git checkout -b sdr-remediation-h001
2. Investigate current permissions: ls -la /opt/camera-service/keys
3. Fix directory permissions: sudo chmod 755 /opt/camera-service/keys (or appropriate)
4. Fix file permissions: sudo chmod 644 /opt/camera-service/keys/* (if files exist)
5. Update deployment/config documentation with correct permissions
6. Test security middleware functionality: python3 -m pytest tests/security/ -v
7. Capture all commands and outputs

Create: evidence/sdr-actual/remediation/SDR-H-001_security_permission_fix.md

DELIVERABLE CRITERIA:
- Permission fix commands with before/after ls -la outputs
- Security middleware test results (all passing)
- Updated deployment configuration with correct permissions
- PR ready for merge with fix implementation

Validation: Security middleware can access keys, all security tests pass
Success confirmation: "SDR-H-001 fixed - security middleware fully functional"
```

### IV&V Validation for SDR-H-001

```
Your role: IV&V
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md

Input: evidence/sdr-actual/remediation/SDR-H-001_security_permission_fix.md

Task: Validate SDR-H-001 fix resolves original security middleware issue

Validation steps:
1. Review Developer's permission fix implementation
2. Verify security middleware can access key storage directory
3. Confirm all security tests pass after fix
4. Validate no new security issues introduced
5. Verify fix addresses original finding from 01_architecture_feasibility_demo.md

Create: evidence/sdr-actual/remediation/SDR-H-001_validation_results.md

PASS/FAIL CRITERIA:
- PASS: Security middleware functional, all tests pass, original issue resolved
- FAIL: Security middleware still has issues or new problems introduced

Deliverable: Clear pass/fail assessment with evidence for PM decision
```

---

## Issue SDR-M-001: Test Expectation Alignment Fix

### Developer Execution (Priority 2 - Day 1, Hours 4-8)

```
Your role: Developer
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md

Issue: SDR-M-001 - Test Expectation Mismatch
Source: evidence/sdr-actual/01_architecture_feasibility_demo.md (Line 94)
Severity: MEDIUM - Test reliability issue

Problem: API contract inconsistencies - test expects `result` to be list, but API returns object with `cameras`, `total`, `connected`

Execute exactly:
1. Create remediation branch: git checkout -b sdr-remediation-m001
2. Locate failing integration tests that expect list format
3. Update test expectations to match actual API contract (object format)
4. Ensure tests validate correct response structure: cameras, total, connected fields
5. Run integration test suite: python3 -m pytest tests/integration/ -v
6. Verify all 19 integration tests pass
7. Update API contract documentation if needed

Create: evidence/sdr-actual/remediation/SDR-M-001_test_alignment_fix.md

DELIVERABLE CRITERIA:
- Updated test code with correct API contract expectations
- All integration tests passing (19/19)
- API contract documentation alignment verified
- Before/after test output showing fix effectiveness

Test organization: Store all test artifacts in evidence/sdr-actual/remediation/test_outputs/
NOT in project root - keep project clean

Validation: All integration tests pass consistently
Success confirmation: "SDR-M-001 fixed - API contract tests aligned, 19/19 passing"
```

### IV&V Validation for SDR-M-001

```
Your role: IV&V
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md

Input: evidence/sdr-actual/remediation/SDR-M-001_test_alignment_fix.md

Task: Validate SDR-M-001 fix resolves test expectation mismatch

Validation steps:
1. Review Developer's test expectation updates
2. Verify all integration tests pass (19/19)
3. Confirm API contract consistency between tests and implementation
4. Validate no test coverage lost during alignment
5. Verify original issue from 01_architecture_feasibility_demo.md resolved

Create: evidence/sdr-actual/remediation/SDR-M-001_validation_results.md

PASS/FAIL CRITERIA:
- PASS: All integration tests pass, API contract aligned, no coverage lost
- FAIL: Tests still failing or API contract still inconsistent

Deliverable: Test execution verification with 100% pass rate evidence
```

---

## Issue SDR-M-002: Real MediaMTX Integration Implementation

### Developer Execution (Priority 3 - Day 2, Hours 1-8)

```
Your role: Developer
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md

Issue: SDR-M-002 - MediaMTX Health Degradation
Source: evidence/sdr-actual/01_architecture_feasibility_demo.md (Line 103)
Severity: MEDIUM - Integration health concern
STRATEGIC DECISION: Replace fake MediaMTX with real MediaMTX v1.13.1 for authentic integration testing

Problem: Integration testing against fake MediaMTX builds assumption debt and gives false confidence

Execute exactly:
1. Create remediation branch: git checkout -b sdr-remediation-m002-real-mediamtx

2. Install MediaMTX v1.13.1:
   wget https://github.com/bluenviron/mediamtx/releases/download/v1.13.1/mediamtx_v1.13.1_linux_amd64.tar.gz
   tar -xzf mediamtx_v1.13.1_linux_amd64.tar.gz
   sudo mv mediamtx /usr/local/bin/
   sudo chmod +x /usr/local/bin/mediamtx

3. Create MediaMTX test configuration:
   Create config/test-mediamtx.yml with minimal test configuration
   Document configuration parameters and rationale

4. Remove fake MediaMTX implementations:
   Remove fake MediaMTX server from tests/integration/test_service_manager_requirements.py
   Remove fake MediaMTX server from tests/integration/test_service_manager_e2e.py
   Update tests to connect to real MediaMTX instance

5. Test real MediaMTX integration:
   Start MediaMTX: mediamtx config/test-mediamtx.yml &
   Test health check: curl http://localhost:8554/v3/config/global/get
   Run integration tests: python3 -m pytest tests/integration/ -k mediamtx -v

6. Document MediaMTX v1.13.1 requirements:
   Installation instructions for team reproducibility
   Configuration management approach
   Version compatibility requirements

Create: evidence/sdr-actual/remediation/SDR-M-002_real_mediamtx_integration.md

DELIVERABLE CRITERIA:
- MediaMTX v1.13.1 installed and configured in test environment
- Fake MediaMTX implementations removed from test suite
- Integration tests updated to use real MediaMTX instance
- MediaMTX test configuration documented for reproducibility
- Health monitoring working with actual MediaMTX v1.13.1 API
- Installation and configuration guide for team

Investigation artifacts: Store in evidence/sdr-actual/remediation/mediamtx_real_integration/
- MediaMTX installation logs
- Test configuration files
- Before/after test comparison
- Real API response samples

Validation: Integration tests pass against real MediaMTX v1.13.1, health monitoring authentic
Success confirmation: "SDR-M-002 resolved - Real MediaMTX v1.13.1 integration implemented, fake servers eliminated"
```

### IV&V Validation for SDR-M-002

```
Your role: IV&V
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md

Input: evidence/sdr-actual/remediation/SDR-M-002_real_mediamtx_integration.md

Task: Validate SDR-M-002 strategic shift to real MediaMTX v1.13.1 integration

Validation steps:
1. Verify MediaMTX v1.13.1 correctly installed and configured
2. Confirm fake MediaMTX implementations completely removed from test suite
3. Validate integration tests work with real MediaMTX instance
4. Verify health monitoring uses authentic MediaMTX v1.13.1 API endpoints
5. Confirm MediaMTX configuration is reproducible and documented
6. Validate original health degradation issue resolved through real integration

Real integration validation:
- Test MediaMTX installation: mediamtx --version (should show v1.13.1)
- Test real API endpoints: curl http://localhost:8554/v3/config/global/get
- Verify no fake server code remains in integration tests
- Confirm integration tests pass with real MediaMTX running
- Validate health check uses actual MediaMTX API behavior

Create: evidence/sdr-actual/remediation/SDR-M-002_real_integration_validation.md

PASS/FAIL CRITERIA:
- PASS: Real MediaMTX v1.13.1 working, fake servers removed, integration tests authentic
- FAIL: Fake servers still present, MediaMTX not working, or integration tests failing

Strategic validation:
- Confirm integration testing now validates against production-like MediaMTX
- Verify assumption debt eliminated through real API testing
- Validate SDR design feasibility proven with authentic dependencies

Deliverable: Real MediaMTX integration verification with production-grade confidence
```

---

## Sprint Coordination and Final Validation

### PM Sprint Oversight (Throughout Sprint)

```
Your role: Project Manager
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md

Task: Coordinate remediation sprint execution and track progress

Sprint management:
1. Monitor Developer progress on each issue (4-hour check-ins)
2. Ensure IV&V validates each fix before proceeding to next issue
3. Maintain issue status tracking in evidence/sdr-actual/remediation/sprint_status.md
4. Make waiver decisions if any issue exceeds effort estimate
5. Ensure all artifacts stay in evidence folder (not project root)

Check-in schedule:
- Hour 4: SDR-H-001 status check
- Hour 8: SDR-M-001 status check  
- Hour 16: SDR-M-002 status check
- Hour 24: Sprint completion review

Issue escalation triggers:
- Any issue taking >150% of estimated effort → Consider waiver
- Any issue introducing new problems → Halt and reassess
- Any issue requiring scope expansion → Reject scope, find minimal fix

Create: evidence/sdr-actual/remediation/sprint_coordination.md

Deliverable: Sprint execution oversight with decision trail
```

### Final Sprint Validation (IV&V)

```
Your role: IV&V
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md

Input: All remediation validation results (SDR-H-001, SDR-M-001, SDR-M-002)

Task: Final validation that remediation sprint resolved all targeted issues AND implemented strategic real integration testing

Final validation:
1. Confirm all High/Medium issues resolved per individual validations
2. Validate strategic shift to real MediaMTX v1.13.1 integration implemented
3. Verify fake server implementations completely eliminated
4. Re-run complete validation suite with real MediaMTX integration
5. Verify no new Critical/High issues introduced
6. Validate overall design feasibility improved with authentic dependencies
7. Assess integration testing authenticity and production confidence

Strategic integration validation:
- Real MediaMTX v1.13.1 working in test environment
- Integration tests use authentic MediaMTX API endpoints
- Health monitoring validates against real MediaMTX behavior
- No assumption debt from fake server implementations
- Production-grade integration confidence achieved

Create: evidence/sdr-actual/remediation/final_sprint_validation.md

SPRINT SUCCESS CRITERIA:
- SDR-H-001: RESOLVED (security middleware functional)
- SDR-M-001: RESOLVED (all integration tests pass 19/19)
- SDR-M-002: RESOLVED (real MediaMTX v1.13.1 integration implemented)
- Strategic Goal: ACHIEVED (authentic integration testing established)
- No new Critical/High issues introduced
- Design feasibility demonstrated with real dependencies

Recommendation: PROCEED to final SDR assessment OR identify remaining blockers

Integration authenticity assessment:
- Confirm no fake servers remain in integration test suite
- Validate MediaMTX v1.13.1 API behavior matches production expectations
- Verify integration testing provides production-grade confidence

Deliverable: Go/no-go recommendation for final SDR assessment with real integration validation
```

---

## Evidence Organization Requirements

### Folder Structure (MANDATORY)
```
evidence/sdr-actual/remediation/
├── SDR-H-001_security_permission_fix.md
├── SDR-H-001_validation_results.md
├── SDR-M-001_test_alignment_fix.md
├── SDR-M-001_validation_results.md
├── SDR-M-002_real_mediamtx_integration.md
├── SDR-M-002_real_integration_validation.md
├── sprint_coordination.md
├── final_sprint_validation.md
├── test_outputs/
│   ├── integration_test_results.txt
│   ├── security_test_results.txt
│   └── before_after_comparisons.txt
├── mediamtx_real_integration/
│   ├── mediamtx_v1.13.1_installation.log
│   ├── test-mediamtx.yml (configuration file)
│   ├── real_api_responses.json
│   ├── fake_server_removal.diff
│   └── integration_test_real_outputs.txt
└── sprint_status.md
```

### Prohibited Locations
- ❌ Project root (keep clean)
- ❌ Random folders in src/
- ❌ Temporary files without cleanup
- ❌ Test artifacts outside evidence folder

### Required Cleanup
```
Before sprint completion:
1. Move any test artifacts from project root to evidence/sdr-actual/remediation/
2. Clean up temporary files and logs from project directories
3. Ensure all evidence is organized in evidence folder structure
4. Remove any debugging scripts or temporary configuration files
```

## Success Criteria Summary
**Sprint Passes When:**
- All 3 issues resolved (1 High + 2 Medium)
- Strategic goal achieved: Real MediaMTX v1.13.1 integration implemented
- Fake MediaMTX servers completely eliminated from test suite
- All validation confirms original findings resolved with authentic integration
- No new Critical/High issues introduced
- Evidence properly organized in evidence folder
- Project root cleaned of remediation artifacts

**Integration Testing Authenticity:**
- MediaMTX v1.13.1 installed and configured in test environment
- Integration tests use real MediaMTX API endpoints
- Health monitoring validates against authentic MediaMTX behavior
- Production-grade integration confidence achieved

**Gate Decision After Sprint:**
- **PROCEED**: All issues resolved, design feasibility confirmed with real dependencies
- **REMEDIATE**: Critical issues remain, additional sprint needed  
- **HALT**: Fundamental design problems discovered through real integration testing