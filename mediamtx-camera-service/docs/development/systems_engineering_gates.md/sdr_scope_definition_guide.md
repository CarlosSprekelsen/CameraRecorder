# SDR (System Design Review) - Design Feasibility Gate Framework

## Purpose
Validate system design feasibility through **minimal working demonstrations** and **requirements traceability**. Ensure design can satisfy requirements before detailed implementation (PDR/CDR).

## SDR Scope (NOT Production Readiness)
**SDR validates design feasibility:**
- Requirements have measurable acceptance criteria and are traceable to design
- Architecture can support requirements through MVP demonstration
- Critical interfaces work (2-3 key methods, not comprehensive testing)
- Basic security concepts proven (auth token rejection, not penetration testing)
- Sanity performance (startup works, not load/stability testing)

**Out of Scope (PDR/CDR Material):**
- End-to-end production scenarios
- Load testing and stability validation
- Comprehensive security testing
- Production deployment readiness
- Full integration testing

## Role-Based Gate Workflow
```
IV&V: Validate feasibility evidence → Report findings
Developer: Fix issues through merged PRs
PM: Gate decisions + waiver authority
```

## SDR-Appropriate Thresholds
```
ENTRY CRITERIA:
- Requirements catalog complete
- Acceptance criteria coverage ≥90%
- Assumptions/non-goals frozen

PASS CRITERIA:
- Requirements: ≥95% have measurable acceptance criteria and design traceability
- Architecture: MVP happy-path demo + requirement-to-component mapping complete
- Interfaces: 2-3 critical methods work (success + negative case)
- Security: Basic auth/token rejection demonstrated
- Performance: Sanity check (service starts, basic operations work)

EXIT CRITERIA:
- 0 Critical/High unresolved findings
- All fixes merged as PRs or documented waivers
- Baseline tag with change manifest
- Evidence pack complete
```

## Severity Taxonomy and Decision Matrix

### Severity Definitions
- **Critical**: Blocks feasibility demonstration (design cannot work)
- **High**: Significant feasibility concern (major design risk)
- **Medium**: Minor design issue (workable with mitigation)
- **Low**: Cosmetic or documentation issue

### Decision Matrix
- **Any Critical/High unresolved** → HALT
- **Only Medium with documented waivers** → CONDITIONAL
- **All issues resolved or waived** → PROCEED

---

## Phase 0: Requirements Feasibility Baseline

### 0. Requirements Traceability Validation (IV&V)

```
Your role: IV&V
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md

Task: Validate requirements have measurable acceptance criteria and design traceability

Validation approach:
- Verify each requirement has measurable acceptance criteria
- Check requirements trace to design components
- Identify untestable or unimplementable requirements
- Validate requirement priority and dependency clarity

Report format:
- Total requirements: X
- With measurable criteria: Y (target ≥95%)
- Traceable to design: Z (target ≥95%)
- Untestable/unimplementable: Count with examples

Create: evidence/sdr-actual/00_requirements_traceability_validation.md

PASS/FAIL CRITERIA:
- PASS: ≥95% have measurable criteria and design traceability
- FAIL: <95% or critical requirements untestable/unimplementable

Deliverable: Requirements feasibility assessment
```

### 0a. Ground Truth Consistency Validation (IV&V)

```
Your role: IV&V
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md

Task: Validate foundational documents support feasibility assessment

Validation scope:
- Requirements documents internally consistent
- Architecture aligns with requirements scope
- API specifications match architectural interfaces
- No contradictions that prevent feasibility demonstration

Report format:
- Documents reviewed: Count
- Inconsistencies found: Count with severity
- Feasibility blockers: Critical inconsistencies
- Resolution required: High/Critical issues only

Create: evidence/sdr-actual/00a_ground_truth_consistency.md

PASS/FAIL CRITERIA:
- PASS: No Critical/High inconsistencies blocking feasibility
- FAIL: Critical inconsistencies prevent design validation
```

### 0b. Requirements Feasibility Gate Review (Project Manager)

```
Your role: Project Manager
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md

Input: evidence/sdr-actual/00_requirements_traceability_validation.md, 00a_ground_truth_consistency.md

GATE REVIEW: Requirements baseline adequacy for design feasibility

Review criteria:
- Requirements have adequate acceptance criteria for design validation
- Design traceability sufficient for feasibility demonstration
- No critical inconsistencies blocking design validation

GATE DECISION:
- PROCEED: Requirements baseline adequate, authorize design feasibility validation
- REMEDIATE: Fix critical/high issues through remediation sprint
- HALT: Requirements fundamentally inadequate for design validation

Create: evidence/sdr-actual/00b_requirements_feasibility_gate_review.md

If REMEDIATE: Initiate 48h remediation sprint (see 0d)
If PROCEED: Authorize Phase 1 design feasibility validation
```

### 0c. Assumptions and Constraints Freeze (Project Manager)

```
Your role: Project Manager
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md

Task: Freeze assumptions and constraints to prevent scope drift during SDR

Document freeze scope:
- Design assumptions with expiry dates
- Technical constraints and limitations
- Explicit non-goals for SDR phase
- Change control requirements

Create: evidence/sdr-actual/00c_assumptions_constraints_freeze.md

Contents:
- Frozen assumptions (environment, dependencies, usage patterns)
- Design constraints (technology, interface, performance boundaries)
- SDR non-goals (what will not be validated at this gate)
- Change control: PM waiver required for assumption changes

Purpose: Prevent scope creep during feasibility validation
```

### 0d. Remediation Sprint (48h) (Developer + IV&V)

```
TRIGGERED: When gate review identifies Critical/High findings

Your role: Developer (with IV&V verification)
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md

Remediation protocol:
- Every Critical/High finding becomes:
  * Merged PR with fix, OR
  * Documented waiver (owner/date/reason/expiry)
- Time-boxed: 48 hours maximum
- IV&V verifies fix adequacy

Create: evidence/sdr-actual/00d_remediation_sprint_results.md

Exit criteria:
- All Critical/High findings resolved through merged PRs or documented waivers
- IV&V verification of fix adequacy
- No new Critical/High issues introduced

Deliverable: Clean findings ledger ready for baseline freeze
```

### 0e. Baseline Freeze and Change Manifest (Developer)

```
Your role: Developer
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md

Execute exactly:
1. git tag -a sdr-baseline-v1.0 -m "SDR baseline after remediation"
2. Create change_manifest.md documenting all changes since last baseline
3. Capture environment state for reproducibility

Create: evidence/sdr-actual/00e_baseline_freeze_manifest.md

EXIT GATE BLOCKER: Phase 1 cannot proceed without baseline tag

Contents:
- Baseline tag: sdr-baseline-v1.0
- Change manifest: All PRs merged during remediation
- Environment snapshot: Dependencies, versions, configuration
- Waiver register: All active waivers with expiry dates

Purpose: Prevent drift, enable reproducible validation
```

---

## Phase 1: Design Feasibility Validation

### 1. Architecture Feasibility Demonstration (Developer)

```
Your role: Developer
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md

Task: Demonstrate architecture can support requirements through MVP happy-path

Feasibility demonstration:
- MVP happy-path working (basic flow end-to-end)
- Requirements-to-component mapping complete
- Critical architectural decisions validated through minimal implementation
- Design adequacy proven, not production completeness

Report format:
- MVP demonstration: Happy path working evidence
- Component mapping: Requirements allocation to architecture
- Design decisions: Key choices validated through minimal proof
- Feasibility assessment: Architecture can support requirements

Create: evidence/sdr-actual/01_architecture_feasibility_demo.md

PASS/FAIL CRITERIA:
- PASS: MVP works, requirements map to components, design feasible
- FAIL: MVP fails, critical mapping gaps, or design infeasible

Deliverable: Architecture feasibility proof through minimal working demonstration
```

### 2. Interface Feasibility Validation (Developer)

```
Your role: Developer
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md

Task: Validate critical interfaces work through minimal exercise (not comprehensive testing)

Interface validation scope:
- 2-3 most critical API methods working
- Success case + one negative case per method
- Interface design feasible for requirements
- Not comprehensive integration testing (that's CDR scope)

Report format:
- Critical methods tested: List with results
- Success cases: Working proof for each method
- Negative cases: Error handling proof for each method
- Interface feasibility: Design can support requirements

Create: evidence/sdr-actual/02_interface_feasibility_validation.md

PASS/FAIL CRITERIA:
- PASS: Critical methods work, error handling demonstrated, design feasible
- FAIL: Methods fail, no error handling, or design infeasible

Deliverable: Interface feasibility proof through minimal working exercise
```

### 2a. Security Concept Validation (Developer)

```
Your role: Developer
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md

Task: Validate basic security concepts work (not comprehensive security testing)

Security validation scope:
- Authentication concept working (token validation)
- Authorization concept working (access rejection)
- Basic security design feasible
- Not penetration testing or comprehensive security validation (that's CDR scope)

Report format:
- Auth concept: Token validation working
- Access control: Unauthorized request rejection working
- Security design: Basic approach feasible for requirements
- Concept validation: Security design can be implemented

Create: evidence/sdr-actual/02a_security_concept_validation.md

PASS/FAIL CRITERIA:
- PASS: Auth/access control concepts work, design feasible
- FAIL: Security concepts fail or design infeasible

Deliverable: Security design feasibility proof through concept validation
```

### 2b. Performance Sanity Check (Developer)

```
Your role: Developer
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md

Task: Validate basic performance sanity (not load testing or stability validation)

Sanity check scope:
- Service starts successfully
- Basic operations complete within reasonable time
- No obvious performance blockers
- Not load testing, stress testing, or production performance validation (that's CDR scope)

Report format:
- Startup time: Service initialization duration
- Basic operations: Key operation timing
- Sanity assessment: No obvious performance blockers
- Feasibility: Performance design approach viable

Create: evidence/sdr-actual/02b_performance_sanity_check.md

PASS/FAIL CRITERIA:
- PASS: Service starts, operations work, no obvious blockers
- FAIL: Startup fails, operations timeout, or obvious blockers

Deliverable: Performance design feasibility through sanity validation
```

### 2c. Design Feasibility Gate Review (Project Manager)

```
Your role: Project Manager
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md

Input: evidence/sdr-actual/01_architecture_feasibility_demo.md, 02_interface_feasibility_validation.md, 02a_security_concept_validation.md, 02b_performance_sanity_check.md

GATE REVIEW: Design feasibility for requirements satisfaction

Review criteria:
- Architecture demonstrates feasibility through MVP
- Interfaces work sufficiently to prove design viability
- Security concepts adequate for design feasibility
- Performance sanity confirms design approach viable

GATE DECISION:
- PROCEED: Design feasibility demonstrated, authorize final assessment
- REMEDIATE: Fix feasibility blockers through remediation sprint
- HALT: Design infeasible, requires fundamental redesign

Create: evidence/sdr-actual/02c_design_feasibility_gate_review.md

If REMEDIATE: Trigger remediation sprint for Critical/High feasibility issues
If PROCEED: Authorize Phase 2 final assessment
```

---

## Phase 2: SDR Final Assessment and Authorization

### 3. SDR Feasibility Assessment (IV&V)

```
Your role: IV&V
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md

Task: Compile overall SDR feasibility assessment based on all validation evidence

Input: All validation evidence from Phases 0-1

Assessment scope:
- Requirements feasibility through acceptance criteria and traceability
- Architecture feasibility through MVP demonstration
- Interface feasibility through critical method validation
- Security/performance feasibility through concept validation

Report format:
- Overall feasibility: FEASIBLE/NOT FEASIBLE for detailed design
- Critical issues: Any remaining Critical/High unresolved
- Risk assessment: Design risks for PDR/CDR phases
- Recommendation: AUTHORIZE/DENY detailed design phase entry

Create: evidence/sdr-actual/03_sdr_feasibility_assessment.md

Deliverable: Design feasibility recommendation with comprehensive evidence
```

### 4. SDR Authorization Decision (Project Manager)

```
Your role: Project Manager
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md

Task: Make final SDR authorization decision based on IV&V feasibility assessment

Input: evidence/sdr-actual/03_sdr_feasibility_assessment.md + waiver log

Decision criteria:
- IV&V feasibility assessment and recommendation
- Waiver log review (no Critical/High unresolved)
- Business risk tolerance for design phase
- Resource availability for detailed design

AUTHORIZATION OPTIONS:
- AUTHORIZE: Design feasible, proceed to PDR phase
- DENY: Design infeasible, requires specific remediation

Create: evidence/sdr-actual/04_sdr_authorization_decision.md

Exit criteria validation:
- 0 Critical/High unresolved findings
- All fixes merged as PRs or documented waivers
- Baseline tag with change manifest exists
- Evidence pack complete

Deliverable: Final SDR authorization with detailed design phase entry approval

If DENY: Specify exactly what design issues must be resolved for reconsideration
If AUTHORIZE: Design approved for detailed implementation planning (PDR phase)
```

### 5. Waiver Log and Issue Ledger (Project Manager)

```
Your role: Project Manager
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md

Task: Maintain waiver log and issue resolution ledger throughout SDR

Create: evidence/sdr-actual/waiver_log.md

Waiver format:
- Issue ID: Reference to finding
- Severity: Critical/High/Medium/Low
- Waiver reason: Business/technical justification
- Owner: Responsible party
- Expiry date: When waiver expires
- Mitigation: Risk mitigation approach

Issue ledger format:
- Issue ID: Unique identifier
- Finding: Description of issue
- Resolution: Fixed via PR or waived
- PR/Waiver: Link to resolution
- Status: RESOLVED/WAIVED

Purpose: Track all issues to closure and maintain waiver accountability
```

---

## SDR Entry and Exit Gates

### Entry Gate Requirements
- Requirements catalog complete and consistent
- Acceptance criteria coverage ≥90%
- Assumptions and non-goals documented and frozen
- Ground truth documents available for validation

### Exit Gate Requirements  
- 0 Critical/High unresolved findings
- All Medium findings resolved or waived with expiry dates
- Baseline tag: sdr-baseline-vX.Y with change manifest
- Evidence pack complete with IV&V validation
- Waiver log up to date with owner accountability

## Authority and Edit Responsibilities

### Role Clarity
- **Project Manager**: Chairs gates, enforces process, authorizes waivers, makes final decisions
- **Developer**: Edits code/docs, merges fixes, implements remediation
- **IV&V**: Validates evidence, verifies fixes, reports findings

### Edit Authority
- **PM**: Process decisions, waiver approval, gate authorization
- **Developer**: Code changes, documentation fixes, baseline creation
- **IV&V**: Validation execution, evidence verification, finding reporting

## Timeline Framework
- **Phase 0**: 3-4 days (Requirements baseline + remediation sprint)
- **Phase 1**: 2-3 days (Design feasibility validation)
- **Phase 2**: 1-2 days (Final assessment and authorization)
- **Total**: 6-9 days

## Success Criteria
**SDR SUCCESS:** Design feasibility demonstrated through minimal working validation, enabling confident entry into detailed design phase (PDR) with clear requirements traceability and architectural foundation.