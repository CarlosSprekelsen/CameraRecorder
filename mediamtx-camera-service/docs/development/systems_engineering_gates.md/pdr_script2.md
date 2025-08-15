# Technology-Agnostic PDR Execution Framework

## Systems Engineering Principles
**Primary Goal:** Validate detailed design through executable evidence and systematic remediation
**Quality Gate:** Working system with comprehensive validation coverage
**Key Innovation:** Every document output has follow-up prompts that use it to drive immediate action

## Universal Agent Framework
```
Your role: [Project Manager|Developer|IV&V]
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md
Task: [specific measurable outcome with clear success/failure criteria]
```

---

## Phase 1: System Reality Assessment (1 Day)

### Task 1.1: Individual Test Analysis
```
Your role: Developer
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md
Task: Execute test suite using individual test isolation strategy to identify real vs artifact failures.

SYSTEMS ENGINEERING APPROACH:
1. Discover all test units in system (use project's test discovery mechanism)
2. Execute each test individually with timeout protection
3. Classify failures by system impact:
   - SYSTEM_CRITICAL: Core system function fails
   - INTEGRATION_ISSUE: Component interaction fails  
   - TEST_ARTIFACT: Test infrastructure/tooling issue
   - REQUIREMENT_GAP: Test assumes unimplemented requirement

EXECUTION STRATEGY:
- Individual test execution prevents cascade failures
- Timeout each test to prevent process hangs
- Continue execution despite individual failures
- Categorize rather than fix during assessment

OUTPUT FORMAT:
Create evidence/pdr-actual/01_test_reality_assessment.md with:
- Total tests discovered: N
- Execution results: PASS/FAIL/TIMEOUT/ERROR counts
- Failure categorization by system impact
- Estimated fix effort per category (hours/days)
- Specific failing tests listed by category

SUCCESS CRITERIA:
- Complete test inventory executed individually
- Real system issues vs test artifacts distinguished
- No process termination due to individual test failures

AGENT ADAPTATION:
If test framework unknown → investigate project structure and adapt
If tests fail to run → document tooling gaps and continue assessment
If timeout occurs → mark as infrastructure issue and continue
```

### Task 1.2: System Critical Fix Execution
```
Your role: Developer
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md
Task: Use evidence/pdr-actual/01_test_reality_assessment.md to fix all SYSTEM_CRITICAL failures immediately.

INPUT DOCUMENT: evidence/pdr-actual/01_test_reality_assessment.md

EXECUTION APPROACH:
1. Extract all SYSTEM_CRITICAL failures from assessment document
2. For each failure: analyze root cause, implement fix, verify fix
3. Re-run each fixed test individually to confirm resolution
4. Update assessment document with fix results

FIX STRATEGY:
- Focus only on SYSTEM_CRITICAL issues that break core functionality
- Implement minimal viable fixes that restore functionality
- Verify each fix through test re-execution
- Document any fixes that create new issues

OUTPUT FORMAT:
Update evidence/pdr-actual/01_test_reality_assessment.md with:
- Fix implemented for each SYSTEM_CRITICAL issue
- Before/after test results
- Any new issues discovered during fixes
- Remaining SYSTEM_CRITICAL issues (if any) with blocking reasons

SUCCESS CRITERIA:
- All SYSTEM_CRITICAL test failures resolved or documented as blocked
- Core system functionality restored through test validation
- Assessment document updated with actual fix results

AGENT ADAPTATION:
If fix creates new issues → document in assessment and continue
If fix impossible → document blocking issue and rationale
If test passes but functionality still broken → investigate and document mismatch
```

### Task 1.3: End-to-End System Validation
```
Your role: IV&V
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md
Task: Validate critical system workflows through black-box integration testing.

INPUT CONTEXT: System with SYSTEM_CRITICAL issues resolved per Task 1.2

SYSTEMS ENGINEERING APPROACH:
1. Identify primary system use cases from requirements
2. Execute workflows without mocking external dependencies
3. Measure system behavior under realistic conditions
4. Document gaps between specified vs actual behavior

VALIDATION METHODOLOGY:
- Black-box testing of complete system
- Real external dependencies where possible
- Realistic data and load conditions
- Error condition and boundary testing

OUTPUT FORMAT:
Create evidence/pdr-actual/02_system_validation_report.md with:
- Use case execution results: WORKING/BROKEN/PARTIAL
- Performance characteristics: response times, resource usage
- Integration points: FUNCTIONAL/FAILED/UNTESTED
- Requirement coverage: VERIFIED/UNVERIFIED/CONTRADICTED
- Specific broken workflows with error details

SUCCESS CRITERIA:
- Critical workflows tested end-to-end
- System behavior documented objectively
- Integration reality vs documentation gaps identified

AGENT ADAPTATION:
If external dependencies unavailable → document substitution strategy
If system fails to start → document startup issues and test what's possible
If requirements unclear → test observable system behavior
```

### Task 1.4: Integration Issue Fix Execution
```
Your role: Developer
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md
Task: Use evidence/pdr-actual/02_system_validation_report.md to fix all BROKEN workflows.

INPUT DOCUMENT: evidence/pdr-actual/02_system_validation_report.md

EXECUTION APPROACH:
1. Extract all BROKEN workflows from validation report
2. For each broken workflow: analyze failure, implement fix, validate fix
3. Re-run workflow validation to confirm resolution
4. Update validation report with fix results

FIX STRATEGY:
- Prioritize workflows by business impact
- Fix integration points and external dependencies first
- Implement error handling for known failure modes
- Validate each fix through end-to-end workflow execution

OUTPUT FORMAT:
Update evidence/pdr-actual/02_system_validation_report.md with:
- Fix implemented for each BROKEN workflow
- Before/after workflow execution results
- Any new issues discovered during fixes
- Remaining BROKEN workflows with blocking reasons

SUCCESS CRITERIA:
- All critical workflows functional end-to-end
- System operates reliably in target environment
- Validation report updated with actual fix results

AGENT ADAPTATION:
If dependencies missing → implement stubs with clear TODOs and document
If architecture issues found → document and implement minimal fixes
If requirements conflicts → document discrepancies and choose implementation
```

---

## Phase 2: Test System Improvement (1-2 Days)

### Task 2.1: Test Artifact Cleanup
```
Your role: Developer
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md
Task: Use evidence/pdr-actual/01_test_reality_assessment.md to fix all TEST_ARTIFACT failures.

INPUT DOCUMENT: evidence/pdr-actual/01_test_reality_assessment.md

EXECUTION APPROACH:
1. Extract all TEST_ARTIFACT failures from assessment document
2. For each failure: fix test infrastructure, remove broken tests, or improve test quality
3. Re-run affected tests to confirm fixes
4. Document test improvements and removals

CLEANUP STRATEGY:
- Remove over-mocked tests that hide integration issues
- Fix test infrastructure and tooling problems
- Improve test reliability and execution speed
- Focus on tests that validate requirements, not implementation details

OUTPUT FORMAT:
Create evidence/pdr-actual/03_test_improvement_report.md with:
- Tests removed with justification for removal
- Test infrastructure fixes implemented
- New tests added with coverage rationale
- Test execution reliability improvements

SUCCESS CRITERIA:
- TEST_ARTIFACT failures resolved through infrastructure fixes or test removal
- Test execution reliable and informative
- Test suite focused on meaningful validation

AGENT ADAPTATION:
If test framework limitations → work within constraints and document gaps
If requirements unclear → test observable system behavior
If test execution unreliable → simplify and improve reliability
```

### Task 2.2: Integration Test Enhancement
```
Your role: Developer
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md
Task: Use evidence/pdr-actual/01_test_reality_assessment.md to address INTEGRATION_ISSUE failures.

INPUT DOCUMENT: evidence/pdr-actual/01_test_reality_assessment.md

EXECUTION APPROACH:
1. Extract all INTEGRATION_ISSUE failures from assessment document
2. For each issue: implement proper integration testing or fix component interactions
3. Add integration tests for workflows validated in Task 1.3
4. Focus on realistic integration scenarios, not mocked interactions

INTEGRATION STRATEGY:
- Test with real external dependencies where possible
- Add integration tests for critical component interactions
- Test error conditions and boundary cases
- Validate data flow and communication patterns

OUTPUT FORMAT:
Update evidence/pdr-actual/03_test_improvement_report.md with:
- Integration issues resolved through proper testing
- New integration tests added with coverage rationale
- Component interaction validation improvements
- External dependency testing enhancements

SUCCESS CRITERIA:
- INTEGRATION_ISSUE failures resolved through proper integration testing
- Integration tests exercise realistic scenarios
- Component interactions validated through testing

AGENT ADAPTATION:
If external systems unavailable → implement realistic stubs with behavior validation
If component interfaces unclear → test observable behavior and document requirements
If integration complex → focus on critical paths and document coverage gaps
```

---

## Phase 3: System Acceptance (1 Day)

### Task 3.1: Comprehensive System Validation
```
Your role: IV&V
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md
Task: Re-validate complete system using all previous fix results.

INPUT DOCUMENTS: 
- evidence/pdr-actual/01_test_reality_assessment.md (updated with fixes)
- evidence/pdr-actual/02_system_validation_report.md (updated with fixes)
- evidence/pdr-actual/03_test_improvement_report.md

VALIDATION APPROACH:
1. Deploy system in clean target environment
2. Execute complete test suite with individual test execution
3. Re-run all workflows from system validation report
4. Measure system performance under realistic conditions
5. Validate all fixes are working correctly

ACCEPTANCE CRITERIA:
- System deploys successfully in target environment
- All critical use cases complete successfully
- Performance meets specified requirements
- Error handling works for known failure modes
- Integration points function with real dependencies

OUTPUT FORMAT:
Create evidence/pdr-actual/04_acceptance_test_results.md with:
- Deployment success/failure in clean environment
- Complete test suite execution results
- Use case execution results with evidence
- Performance measurements vs requirements
- Integration point validation results
- Recommendation: ACCEPT/CONDITIONAL/REJECT

SUCCESS CRITERIA:
- System demonstrates specified functionality
- Performance meets requirements
- Reliability acceptable for intended use

AGENT ADAPTATION:
If deployment fails → document issues and test what's possible
If performance issues → measure and document actual performance
If new issues found → document for potential next iteration
```

### Task 3.2: Final Authorization Decision
```
Your role: Project Manager
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md
Task: Make authorization decision based on all validation evidence.

INPUT DOCUMENTS:
- evidence/pdr-actual/01_test_reality_assessment.md (with fixes)
- evidence/pdr-actual/02_system_validation_report.md (with fixes)  
- evidence/pdr-actual/03_test_improvement_report.md
- evidence/pdr-actual/04_acceptance_test_results.md

DECISION FRAMEWORK:
1. Review all validation evidence and fix results
2. Assess remaining risks and technical debt
3. Evaluate system readiness for production deployment
4. Make authorization decision with clear rationale

DECISION CRITERIA:
- AUTHORIZE: System demonstrates required functionality with acceptable risk
- CONDITIONAL: System functional but requires specific conditions/limitations
- DENY: Critical functionality missing or unacceptable risks identified

OUTPUT FORMAT:
Create evidence/pdr-actual/05_authorization_decision.md with:
- Decision: AUTHORIZE/CONDITIONAL/DENY
- Evidence summary: key validation results supporting decision
- Risk assessment: remaining technical risks and mitigation strategy
- Conditions: specific requirements if conditional authorization
- Next phase scope: clear direction for production deployment

SUCCESS CRITERIA:
- Decision based on demonstrated system capability
- Risks clearly identified and assessed
- Path forward clearly defined

AGENT ADAPTATION:
If evidence unclear → request specific additional validation
If risks high → define specific conditions and monitoring requirements
If system inadequate → specify requirements for next iteration
```

---

## Key Framework Improvements

### Document-Action Coupling
- Every document has immediate follow-up tasks that use it
- No document exists without a consumer task
- Assessment documents get updated with actual fix results

### Individual Task Focus
- Each task has single, clear objective
- Tasks build on previous results systematically
- No attempt to do everything in one prompt

### Technology Independence
- Framework adapts to any technology stack
- Uses project's existing tools and approaches
- Focuses on systems engineering principles, not implementation details

### Evidence-Based Progression
- Each phase validates that previous fixes actually worked
- System behavior drives decisions, not documentation compliance
- Real validation with working software throughout process