# Streamlined CDR Execution Plan

## CDR Objective
Prove MediaMTX Camera Service production readiness through actual execution, testing, and validation. Working software over documentation.

## Global Acceptance Thresholds
```
Coverage: ≥70% overall, ≥80% critical paths
Security: 0 Critical/High vulnerabilities, no secrets
Performance: JSON-RPC p95 ≤200ms, start-record ≤2s  
Resilience: Recovery <30s, RSS drift <5%/4h, error rate <0.5%
Evidence: Command outputs and verification data required
```

## Phase Gate Review Protocol

**At the end of each phase, PM conducts gate review:**

### Phase Gate Review (Project Manager)
```
Your role: Project Manager
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md

Input: All evidence files from completed phase

Task: Assess phase completion and make continuation decision

GATE REVIEW PROCESS:
1. Review all evidence files for completeness and quality
2. Identify issues by severity: Critical, Major, Minor
3. Assess risk of proceeding with known issues
4. Make informed business decision

DECISION OPTIONS:
- PROCEED: Continue to next phase (acceptable risk level)
- REMEDIATE: Fix critical/major issues before continuing  
- CONDITIONAL: Proceed with specific risk mitigations
- HALT: Stop CDR due to unacceptable issues

Create: evidence/sprint-3-actual/[Phase]_gate_review.md
Include: Issue assessment, risk analysis, decision rationale, next steps

If REMEDIATE: Generate specific copy-paste ready prompts for required remediation work

Deliverable: Gate decision with justification and ready-to-execute remediation prompts (if needed).
```

---

## Phase 0: Baseline

### 0. CDR Baseline (Developer)
```
Your role: Developer
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md

Execute exactly:
1. git tag -a v1.0.0-cdr -m "CDR baseline"
2. git rev-parse HEAD > baseline-sha.txt
3. pip freeze > requirements-cdr.txt  
4. uname -a > environment.txt
5. python --version >> environment.txt
6. sha256sum requirements-cdr.txt > checksums.txt

VALIDATION LOOP:
- If any command fails, troubleshoot and re-execute until successful
- Verify each output file contains actual data (not empty)
- Check file sizes: requirements-cdr.txt >100 bytes, environment.txt >50 bytes, checksums.txt >30 bytes
- If validation fails, iterate until all outputs are complete

Create: evidence/sprint-3-actual/00_cdr_baseline_and_build.md

DELIVERABLE CRITERIA:
- Requirements section: Complete pip freeze output (not empty)
- Environment section: uname -a + python version (not empty)  
- Checksums section: sha256sum results (not empty)
- Task incomplete until ALL criteria met

Success confirmation: "All commands executed successfully, evidence file complete with populated sections"
```

### 0a. Baseline Gate Review (Project Manager)
```
Your role: Project Manager
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md

Input: evidence/sprint-3-actual/00_cdr_baseline_and_build.md

GATE REVIEW: Assess baseline establishment completion
- Verify baseline tag created and environment documented
- Identify any baseline establishment issues
- Decide if baseline is sufficient for CDR execution

DECISION: PROCEED/REMEDIATE/HALT

Create: evidence/sprint-3-actual/00a_baseline_gate_review.md
Include: Baseline assessment, any issues identified, gate decision

If REMEDIATE: Generate copy-paste ready Developer prompt with specific baseline fixes required
If PROCEED: Authorize Phase 1 continuation
```

---

## Phase 1: Foundation

### 1. CDR Scope Definition (Project Manager)
```
Your role: Project Manager
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md

Input: evidence/sprint-3-actual/00_cdr_baseline_and_build.md

Validate input exists and contains: tag details, SHA, requirements, environment, checksums
If missing: Document gaps and return to Developer

Task: Define CDR scope covering all requirements, architecture, testing, security, performance, deployment readiness.

Create: evidence/sprint-3-actual/01_cdr_scope_definition.md
Include: Objectives, scope, success criteria, baseline approval

Deliverable: Complete scope definition with baseline approval.
```

### 2. Requirements Inventory (IV&V)
```
Your role: IV&V
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md

Input: evidence/sprint-3-actual/01_cdr_scope_definition.md

Task: Inventory ALL requirements from docs/requirements/client-requirements.md
Categorize: customer-critical, system-critical, security-critical, performance-critical

Create: evidence/sprint-3-actual/02_requirements_inventory.md
Include: Complete catalog with priority matrix and testability assessment

Deliverable: Requirements register with categories and priorities.
```

### 3. Code Quality Gate (Developer)
```
Your role: Developer
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md

Execute exactly:
1. ruff check . > lint-results.txt
2. bandit -r src/ > security-scan.txt
3. pip-audit --format=json > audit-results.json
4. cyclonedx-py -o sbom.json

VALIDATION LOOP:
- If any command fails, troubleshoot and re-execute until successful
- Verify each output file contains actual results (not empty)
- Check file sizes: lint-results.txt >10 bytes, security-scan.txt >50 bytes, audit-results.json >20 bytes
- Apply thresholds: 0 Critical/High vulnerabilities, no secrets, lint clean
- If validation fails, iterate until all outputs are complete

Create: evidence/sprint-3-actual/03_code_quality_gate.md

DELIVERABLE CRITERIA:
- Lint results section: Complete ruff output (not empty)
- Security scan section: Complete bandit output (not empty)
- Audit results section: pip-audit JSON results (not empty)
- SBOM file: Generated and referenced
- Pass/fail assessment: Clear threshold evaluation
- Task incomplete until ALL criteria met

Success confirmation: "All quality tools executed successfully, evidence file complete with threshold assessment"
```

### 4. Architecture Traceability (IV&V)
```
Your role: IV&V
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md

Input: evidence/sprint-3-actual/02_requirements_inventory.md

Task: Map every requirement to architecture component from docs/architecture/overview.md
Create Requirements Verification Traceability Matrix (RVTM)

Create: evidence/sprint-3-actual/04_architecture_rvtm.md
Include: Complete mapping, gap analysis, adequacy assessment

Deliverable: RVTM with 100% requirement allocation or gap identification.
```

### 4a. Foundation Phase Gate Review (Project Manager)
```
Your role: Project Manager
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md

Input: evidence/sprint-3-actual/01_cdr_scope_definition.md, 02_requirements_inventory.md, 03_code_quality_gate.md, 04_architecture_rvtm.md

GATE REVIEW: Assess foundation phase completion and readiness for system testing
- Evaluate code quality gate results vs global thresholds
- Review architecture traceability completeness
- Assess cumulative risk of identified issues
- Decide if foundation is solid enough for system validation

DECISION OPTIONS:
- PROCEED: Foundation acceptable, authorize system testing
- REMEDIATE: Fix critical issues before system testing (specify scope)
- CONDITIONAL: Proceed with specific mitigations
- HALT: Foundation issues too severe for continuation

Create: evidence/sprint-3-actual/04a_foundation_gate_review.md
Include: Foundation assessment, risk analysis, issue prioritization, gate decision with rationale

If REMEDIATE: Generate copy-paste ready Developer prompts with specific issue resolution requirements (e.g., security vulnerability fixes, code quality improvements)
If PROCEED: Authorize Phase 2 system validation
```

---

## Phase 2: System Validation

### 5. System Startup Test (IV&V)
```
Your role: IV&V
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md

Execute exactly:
1. git checkout v1.0.0-cdr
2. python3 -m camera_service &
3. sleep 5 && curl http://localhost:8002/health > health-check.txt
4. Test API: get_camera_list, take_snapshot, start_recording, stop_recording
5. ls -la recordings/ > file-evidence.txt
6. pkill -f camera_service

VALIDATION LOOP:
- If service fails to start, check logs and troubleshoot until running
- If API calls fail, verify service status and retry until successful
- Verify health-check.txt contains valid response (not error)
- Verify file-evidence.txt shows created files (not empty directory)
- If validation fails, iterate until all operations succeed

Create: evidence/sprint-3-actual/05_system_startup_test.md

DELIVERABLE CRITERIA:
- Service startup: Successful launch confirmation
- Health check: Valid HTTP response received
- API operations: All methods executed successfully with responses
- File outputs: Evidence of created photos/videos
- Service shutdown: Clean termination
- Task incomplete until ALL criteria met

Success confirmation: "Service started successfully, all API operations completed, file outputs verified"
```

### 6. Test Suite Execution (IV&V)
```
Your role: IV&V
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md

Execute exactly:
1. export PYTHONPATH=$PWD/src:$PYTHONPATH
2. python3 -m pytest tests/unit/ -v --cov --tb=short > unit-results.txt
3. python3 -m pytest tests/integration/ -v --tb=short > integration-results.txt  
4. python3 -m pytest tests/security/ -v --tb=short > security-results.txt
5. python3 run_all_tests.py > full-results.txt

VALIDATION LOOP:
- If any test suite fails to run, troubleshoot environment and retry
- If tests hang, timeout after 10 minutes and document hanging tests
- Verify each results file contains actual test output (not empty)
- Check for test counts: collected X items, Y passed, Z failed
- If validation fails, iterate until all test suites execute completely

Create: evidence/sprint-3-actual/06_test_execution_results.md

DELIVERABLE CRITERIA:
- Unit test results: Complete pytest output with pass/fail counts and coverage %
- Integration test results: Complete pytest output with pass/fail counts
- Security test results: Complete pytest output with pass/fail counts  
- Full results: Complete test automation output
- Summary analysis: Total tests, failures, coverage metrics
- Task incomplete until ALL criteria met

Success confirmation: "All test suites executed completely, results captured with pass/fail analysis"
```

### 7. Requirements Coverage (IV&V)
```
Your role: IV&V
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md

Input: evidence/sprint-3-actual/06_test_execution_results.md, 02_requirements_inventory.md

Task: Map requirements to tests, identify coverage gaps
Validate tests use public APIs and business outcomes

Create: evidence/sprint-3-actual/07_requirements_coverage.md
Include: Requirement-to-test matrix, gap analysis

Deliverable: Coverage assessment with gap identification.
```

### 7a. System Validation Phase Gate Review (Project Manager)
```
Your role: Project Manager
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md

Input: evidence/sprint-3-actual/05_system_startup_test.md, 06_test_execution_results.md, 07_requirements_coverage.md

GATE REVIEW: Assess system validation completion and readiness for production testing
- Evaluate system startup and basic functionality results
- Review test execution results and coverage metrics vs global thresholds
- Assess requirements coverage completeness
- Decide if system validation sufficient for production readiness testing

DECISION OPTIONS:
- PROCEED: System validation acceptable, authorize production readiness testing
- REMEDIATE: Fix critical system issues before production testing (specify scope)
- CONDITIONAL: Proceed with specific system limitations documented
- HALT: System validation issues too severe for production consideration

Create: evidence/sprint-3-actual/07a_system_validation_gate_review.md
Include: System validation assessment, coverage analysis, issue evaluation, gate decision with rationale

If REMEDIATE: Generate copy-paste ready Developer/IV&V prompts with specific system issue resolution requirements (e.g., test fixes, coverage improvements, functionality corrections)
If PROCEED: Authorize Phase 3 production readiness testing
```

---

## Phase 3: Production Readiness

### 8. Security Testing (IV&V)
```
Your role: IV&V
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md

Execute exactly:
1. Test invalid JWT tokens
2. Attempt unauthorized API access
3. nmap localhost -p 8002
4. Test rate limiting with request flooding
5. Scan for exposed endpoints

Create: evidence/sprint-3-actual/08_security_testing.md
Include: Attack logs, scan results, vulnerability findings

Deliverable: Security validation with vulnerability assessment.
```

### 9. Performance Testing (IV&V)
```
Your role: IV&V
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md

Execute exactly:
1. Baseline: measure single user response times
2. Load test: 10, 50, 100 concurrent WebSocket connections
3. Monitor: CPU, memory, response times
4. Apply thresholds: JSON-RPC p95 ≤200ms

Create: evidence/sprint-3-actual/09_performance_testing.md
Include: Metrics, graphs, threshold compliance

Deliverable: Performance validation against thresholds.
```

### 10. Resilience Testing (IV&V)
```
Your role: IV&V
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md

Execute exactly:
1. python3 -m camera_service &
2. Start recording and monitor for 4 hours: while true; do ps aux | grep camera_service >> resource-log.txt; sleep 300; done &
3. After 2 hours: simulate camera disconnect/reconnect, measure recovery time
4. After 3 hours: pkill -9 camera_service, restart, verify clean recovery
5. Kill monitoring processes

VALIDATION LOOP:
- If service fails during soak, restart and continue from failure point
- If monitoring fails, restart monitoring process
- Verify resource-log.txt shows continuous monitoring (not empty, >48 entries)
- Measure actual recovery times with timestamps
- Verify service restart succeeds without configuration corruption
- If validation fails, repeat test cycle until successful completion

Create: evidence/sprint-3-actual/10_resilience_testing.md

DELIVERABLE CRITERIA:
- Soak test: 4-hour monitoring log with resource data
- Recovery test: Measured recovery times <30s with evidence
- Restart test: Clean service restart verification
- Threshold assessment: RSS drift <5%, error rate <0.5%
- Monitoring data: Actual CPU/memory measurements over time
- Task incomplete until ALL criteria met

Success confirmation: "4-hour soak completed, recovery times measured, thresholds validated"
```

### 11. Deployment Testing (IV&V)
```
Your role: IV&V
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md

Execute exactly:
1. Fresh environment deployment
2. Test start/stop/restart operations
3. Configuration management validation
4. Backup/restore testing
5. Log management verification

Create: evidence/sprint-3-actual/11_deployment_testing.md
Include: Installation logs, operation results

Deliverable: Deployment readiness with operational validation.
```

### 12. API Contract Validation (Developer)
```
Your role: Developer
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md

Execute exactly:
1. Create JSON schemas for all API methods
2. Implement contract tests validating schemas
3. Run contract test suite
4. Document API versioning policy

Create: evidence/sprint-3-actual/12_api_contract_validation.md
Include: Schema definitions, test results, versioning policy

Deliverable: API contracts with validation proof.
```

### 12a. Production Readiness Phase Gate Review (Project Manager)
```
Your role: Project Manager
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md

Input: evidence/sprint-3-actual/08_security_testing.md, 09_performance_testing.md, 10_resilience_testing.md, 11_deployment_testing.md, 12_api_contract_validation.md

GATE REVIEW: Assess production readiness completion and authorization readiness
- Evaluate security testing results vs global thresholds
- Review performance and resilience testing vs global thresholds  
- Assess deployment and operational readiness
- Evaluate API contract stability and client protection
- Decide if system ready for production authorization consideration

DECISION OPTIONS:
- PROCEED: Production readiness acceptable, authorize final decision phase
- REMEDIATE: Fix critical production issues before authorization (specify scope)
- CONDITIONAL: Proceed with specific production limitations/mitigations
- HALT: Production readiness issues too severe for authorization

Create: evidence/sprint-3-actual/12a_production_readiness_gate_review.md
Include: Production readiness assessment, threshold compliance analysis, risk evaluation, gate decision with rationale

If REMEDIATE: Generate copy-paste ready Developer/IV&V prompts with specific production issue resolution requirements (e.g., security fixes, performance optimizations, deployment corrections)
If PROCEED: Authorize Phase 4 final decision process
```

---

## Phase 4: Decision

### 13. Issue Compilation (IV&V)
```
Your role: IV&V
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md

Input: All evidence files 05-12

Task: Compile issues from all testing phases
Categorize: Critical (blocks production), Major (reduces capability), Minor

Create: evidence/sprint-3-actual/13_issue_compilation.md
Include: Issue register, severity classification, remediation priorities

Deliverable: Complete issue assessment with production blocker identification.
```

### 14. Issue Remediation (Developer)
```
Your role: Developer
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md

Input: evidence/sprint-3-actual/13_issue_compilation.md

Execute for each Critical/Major issue:
1. Implement fix for identified issue
2. Re-run specific tests that failed: pytest [specific_test] -v > fix-validation.txt
3. Verify fix resolves issue without introducing new failures
4. Update affected documentation if needed
5. Repeat for next issue until all Critical/Major issues resolved

VALIDATION LOOP:
- If fix doesn't resolve issue, iterate until successful
- If fix introduces new failures, rollback and retry different approach
- If tests still fail, continue fixing until tests pass
- Verify fix-validation.txt shows successful test execution for each fix
- If validation fails, continue iteration until all issues properly resolved

Create: evidence/sprint-3-actual/14_issue_remediation.md

DELIVERABLE CRITERIA:
- Fix implementation: Code/config changes documented for each issue
- Validation evidence: Re-test results showing issue resolution
- Regression check: Confirmation no new issues introduced
- Documentation updates: Changes to affected docs
- Complete resolution: All Critical/Major issues marked resolved with evidence
- Task incomplete until ALL criteria met

Success confirmation: "All Critical/Major issues resolved with validation evidence, no new failures introduced"
```

### 15. Final Assessment (IV&V)
```
Your role: IV&V
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md

Input: All evidence files 01-14

Task: Compile comprehensive technical assessment
Recommend: AUTHORIZE/CONDITIONAL/DENY for production

Create: evidence/sprint-3-actual/15_final_assessment.md
Include: Complete evaluation, production readiness recommendation

Deliverable: Production readiness assessment with recommendation.
```

### 16. Production Decision (Project Manager)
```
Your role: Project Manager
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md

Input: evidence/sprint-3-actual/15_final_assessment.md

Task: Make production authorization decision based on evidence
Decision: AUTHORIZE/CONDITIONAL/DENY with justification

Create: evidence/sprint-3-actual/16_production_decision.md
Include: Decision, rationale, conditions (if any)

Deliverable: Final production authorization decision.
```

---

## Evidence Management

**Document Structure:**
```markdown
# Document Title
**Version:** 1.0
**Date:** YYYY-MM-DD  
**Role:** [Developer/IV&V/Project Manager]
**CDR Phase:** [Phase Number]

## Purpose
[Brief task description]

## Execution Results  
[Command outputs and evidence]

## Conclusion
[Pass/fail assessment]
```

**File Naming:** ##_descriptive_name.md (00-16)
**Location:** evidence/sprint-3-actual/
**Requirements:** Include actual command outputs, not summaries