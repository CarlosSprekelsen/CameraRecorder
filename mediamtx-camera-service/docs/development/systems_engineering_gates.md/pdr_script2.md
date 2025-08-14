# Technology-Agnostic PDR Execution Framework

## Systems Engineering Principles
**Primary Goal:** Validate system design through executable evidence
**Quality Gate:** Working system with meaningful validation coverage
**Failure Mode:** Process continues despite component failures - agents adapt and report

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
Create test_reality_assessment.md with:
- Total tests discovered: N
- Execution results: PASS/FAIL/TIMEOUT/ERROR counts
- Failure categorization by system impact
- Estimated fix effort per category (hours/days)

SUCCESS CRITERIA:
- Complete test inventory executed individually
- Real system issues vs test artifacts distinguished
- No process termination due to individual test failures

AGENT ADAPTATION:
If test framework unknown → investigate project structure and adapt
If tests fail to run → document tooling gaps and continue assessment
If timeout occurs → mark as infrastructure issue and continue
```

### Task 1.2: End-to-End System Validation
```
Your role: IV&V
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md
Task: Validate critical system workflows through black-box integration testing.

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
Create system_validation_report.md with:
- Use case execution results: WORKING/BROKEN/PARTIAL
- Performance characteristics: response times, resource usage
- Integration points: FUNCTIONAL/FAILED/UNTESTED
- Requirement coverage: VERIFIED/UNVERIFIED/CONTRADICTED

SUCCESS CRITERIA:
- Critical workflows tested end-to-end
- System behavior documented objectively
- Integration reality vs documentation gaps identified

AGENT ADAPTATION:
If external dependencies unavailable → document substitution strategy
If system fails to start → document startup issues and test what's possible
If requirements unclear → test observable system behavior
```

---

## Phase 2: System Correction (2-3 Days)

### Task 2.1: Critical System Repair
```
Your role: Developer
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md
Task: Repair system components to achieve working end-to-end functionality.

SYSTEMS ENGINEERING APPROACH:
1. Prioritize fixes by system impact (CRITICAL → HIGH → MEDIUM)
2. Focus on integration points and core workflows first
3. Implement minimal viable fixes that restore functionality
4. Validate each fix through system-level testing

REPAIR STRATEGY:
- Fix system functionality, not test compliance
- Address integration failures before unit test failures
- Implement error handling for known failure modes
- Document system behavior changes

QUALITY GATES:
- Each fix improves observable system behavior
- Integration points become functional
- Core workflows complete successfully
- System startup and shutdown work reliably

OUTPUT FORMAT:
Create system_repair_log.md with:
- Fix priority and system impact
- Before/after behavior comparison
- Integration point status changes
- Remaining known issues

SUCCESS CRITERIA:
- Critical workflows functional end-to-end
- System operates reliably in target environment
- Integration points work with real dependencies

AGENT ADAPTATION:
If dependencies missing → implement stubs with clear TODOs
If architecture issues found → document and implement minimal fixes
If requirements conflicts → document discrepancies and choose implementation
```

### Task 2.2: Test System Improvement
```
Your role: Developer
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md
Task: Improve test system to detect real regressions and validate requirements.

SYSTEMS ENGINEERING APPROACH:
1. Eliminate tests that don't detect real system problems
2. Add tests for critical failure modes
3. Ensure tests validate requirements, not implementation details
4. Focus on integration and boundary condition testing

TEST IMPROVEMENT STRATEGY:
- Remove over-mocked tests that hide integration issues
- Add tests that would catch real regressions
- Test error conditions and edge cases
- Validate requirement satisfaction through testing

QUALITY METRICS:
- Tests detect when requirements are violated
- Test failures indicate real system problems
- Integration tests exercise realistic scenarios
- Error condition coverage for critical paths

OUTPUT FORMAT:
Create test_improvement_report.md with:
- Tests removed/simplified with justification
- New tests added with coverage rationale
- Test execution time and reliability improvements
- Requirement coverage validation

SUCCESS CRITERIA:
- Test failures indicate real system issues
- Tests would catch regressions in critical functionality
- Test execution is reliable and informative

AGENT ADAPTATION:
If test framework limitations → work within constraints and document gaps
If requirements unclear → test observable system behavior
If test execution unreliable → simplify and improve reliability
```

---

## Phase 3: System Acceptance (1 Day)

### Task 3.1: System Acceptance Testing
```
Your role: IV&V
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md
Task: Validate system readiness through comprehensive acceptance testing.

SYSTEMS ENGINEERING APPROACH:
1. Deploy system in clean target environment
2. Execute complete acceptance test suite
3. Validate all critical requirements through system operation
4. Measure system performance under realistic conditions

ACCEPTANCE CRITERIA:
- System deploys successfully in target environment
- All critical use cases complete successfully
- Performance meets specified requirements
- Error handling works for known failure modes
- Integration points function with real dependencies

VALIDATION METHODOLOGY:
- Clean environment deployment testing
- Full workflow execution with realistic data
- Performance and reliability testing
- Boundary condition and error testing

OUTPUT FORMAT:
Create acceptance_test_results.md with:
- Deployment success/failure in clean environment
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
If requirements unclear → validate against observable system behavior
```

### Task 3.2: Authorization Decision
```
Your role: Project Manager
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md
Task: Make authorization decision based on system demonstration and risk assessment.

DECISION FRAMEWORK:
1. Review system acceptance test results
2. Assess risks of proceeding vs waiting
3. Evaluate system readiness for next phase
4. Make authorization decision with clear rationale

DECISION CRITERIA:
- AUTHORIZE: System demonstrates required functionality with acceptable risk
- CONDITIONAL: System functional but requires specific conditions/limitations
- DENY: Critical functionality missing or unacceptable risks identified

RISK ASSESSMENT:
- Technical risks from system gaps
- Schedule risks from remaining work
- Integration risks from external dependencies
- Operational risks from known issues

OUTPUT FORMAT:
Create authorization_decision.md with:
- Decision: AUTHORIZE/CONDITIONAL/DENY
- Risk assessment and mitigation strategy
- Conditions or limitations if applicable
- Next phase authorization and scope

SUCCESS CRITERIA:
- Decision based on demonstrated system capability
- Risks clearly identified and assessed
- Path forward clearly defined

AGENT ADAPTATION:
If results unclear → request clarification and additional testing
If risks high → define conditions and mitigations
If system inadequate → define specific requirements for retry
```

---

## Framework Characteristics

### Process Resilience
- **Failure Tolerance:** Individual component failures don't stop process
- **Agent Adaptation:** Clear guidance for handling unexpected conditions
- **Graceful Degradation:** Process continues with documented limitations

### Technology Independence
- **Language Agnostic:** Principles apply to any technology stack
- **Framework Neutral:** Adapts to any testing or build framework
- **Environment Flexible:** Works with any deployment target

### Systems Engineering Focus
- **Requirements Driven:** Validate requirement satisfaction through system behavior
- **Integration Emphasis:** Focus on component interaction over unit testing
- **Risk Based:** Prioritize by system impact and failure consequences
- **Evidence Based:** Decisions made on demonstrated system capability

### Quality Principles
- **Working System First:** Functionality over documentation compliance
- **Meaningful Testing:** Tests that detect real problems
- **Realistic Validation:** Testing with real dependencies and conditions
- **Continuous Adaptation:** Process adapts to project realities

### Timeline Optimization
- **Parallel Execution:** Tasks can overlap where dependencies allow
- **Individual Test Strategy:** Faster feedback and better isolation
- **Focused Effort:** Effort directed at system functionality
- **Decision Velocity:** Quick go/no-go based on demonstrated capability