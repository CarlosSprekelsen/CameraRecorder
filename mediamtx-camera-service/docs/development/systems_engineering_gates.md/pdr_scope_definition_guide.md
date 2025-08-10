PDR (Preliminary Design Review) – Scope Definition and Execution Guide (Revised, Action‑Oriented)

Purpose: Ensure the detailed design is implementable and validated through working code and measurable evidence, with time‑boxed remediation and a frozen baseline to prevent design drift prior to full implementation.

PDR Objective

Validate detailed system design completeness and implementability by executing critical prototypes, contract tests, and pipeline runs; convert findings into merged changes before advancing.

Non‑Goals and Scope Guardrails

No new features beyond SDR‑approved scope

No refactors except those required to meet PDR acceptance thresholds

No expansion of hardware/OS matrix beyond MVP targets

Global PDR Acceptance Thresholds

Design Completeness: 100% of SDR‑approved requirements mapped to design elements
Interface Compliance: 100% of external APIs have schemas + contract tests passing
Performance Budget: Prototype measurements meet or exceed PDR Budget Table
Security Design: Threat model complete; all High risks mitigated or waived with owner/date
Test Strategy: Working harnesses covering interfaces and critical flows
Build System: CI pipeline green on baseline; reproducible build with checksums
Evidence: All claims linked to artifacts (logs, test outputs, binaries)
Timebox: 7–12 working days; max 2 iterations

Phase 0: Design Baseline

0. Detailed Design Inventory and Validation (IV&V)

Role: IV&V
Task: Validate detailed design completeness and implementability.

Execute exactly:
1) Inventory all detailed design artifacts and specifications
2) Validate coverage of all SDR‑approved requirements (ID‑level)
3) Check consistency across components and interfaces
4) Verify implementation guidance is sufficient for development
5) Assess traceability to architecture and requirements

Create: evidence/pdr-actual/00_design_validation.md

Deliverable Criteria:
- Complete design catalog
- Coverage assessment: requirement → design element(s)
- Consistency report: contradictions and omissions
- Implementability assessment
- Traceability matrix

Success Criteria: Detailed design validated as complete and implementable.

0a. Design Baseline Gate Review (Project Manager)

Role: Project Manager
Input: 00_design_validation.md

Gate Review:
- Verify completeness/consistency
- Evaluate implementability
- Decide readiness for remediation

Decision: PROCEED | REMEDIATE | HALT

Create: evidence/pdr-actual/00a_design_gate_review.md

0d. Time‑Boxed Design Remediation Sprint (PM, Developer, IV&V)

Role: Project Manager (lead); Developer (edits); IV&V (verifies)
Objective: Resolve all High/Blocking design issues via merged changes.
Timebox: 48h (+ optional 24h mop‑up)

Execute exactly:
1) Developer opens focused PRs addressing IV&V findings (small diffs)
2) Each PR body includes: Finding ID(s), before/after, rationale
3) IV&V verifies each PR resolves the finding
4) PM maintains ledger mapping finding → PR/commit or waiver

Create: evidence/pdr-actual/00d_design_remediation_sprint.md

Exit Criteria:
- 100% High/Blocking findings are Merged or Waived (owner/date/reason)
- 0 unresolved cross‑document contradictions
- Implementability blockers removed

0e. Design Baseline Merge & Tag (Project Manager)

Role: Project Manager
Objective: Freeze the remediated design.

Execute exactly:
1) Merge remediation PRs into main
2) Generate evidence/pdr-actual/change_manifest.md (files changed, summaries, links)
3) Tag: git tag -a pdr-baseline-vX.Y -m "PDR baseline after remediation"
4) Archive design/ and api/ folders with checksums

Gate: Phase 1 cannot start without pdr-baseline-vX.Y and change_manifest.md.

Phase 1: Component and Interface Validation

1. Critical Component Prototyping (Developer)

Role: Developer
Task: Build and validate prototypes of highest‑risk components.

Execute exactly:
1) Select top risk components per 00_design_validation.md
2) Implement minimal, runnable prototypes per design specs
3) Measure behavior/perf under realistic conditions
4) Validate integration points and dependencies

Create: evidence/pdr-actual/01_component_prototyping.md

Deliverable Criteria:
- Prototype code references and commands
- Spec compliance evidence (logs, outputs)
- Performance snapshots (p95 latency, CPU/RSS)
- Integration points verified

Success Criteria: Critical components prototyped and validated against design specifications.

2. Interface Implementation and Testing (Developer)

Role: Developer
Task: Implement external and critical internal interfaces; validate with code.

Execute exactly:
1) Implement external APIs from design specs
2) Implement critical internal protocols
3) Create interface test suites
4) Validate with real data flows and error cases

Create: evidence/pdr-actual/02_interface_implementation.md

Deliverable Criteria:
- Working interface implementations
- Passing interface tests
- Data flow transcripts (success and error paths)
- Load sanity results

Success Criteria: All scoped interfaces implemented and validated with test suites.

2b. API/ICD Contract Validation & Freeze (Project Manager)

Role: Project Manager
Input: 02_interface_implementation.md
Task: Lock externally visible behavior via schemas and contract tests.

Execute exactly:
1) Developer publishes JSON Schemas (req/resp, error codes)
2) Contract tests run against service; all required methods covered
3) Versioning: declare API vMAJOR.MINOR and deprecation policy

Create: evidence/pdr-actual/02b_api_contract_validation.md

Exit Criteria:
- 100% contract tests PASS
- Schema + versioning recorded; policy published

3. Security Design Validation (Developer)

Role: Developer
Task: Validate security design through prototypes and tests.

Execute exactly:
1) Implement authentication/authorization prototypes
2) Validate threat model with targeted tests
3) Exercise common attack vectors; verify defenses
4) Validate security configuration and deployment procedures

Create: evidence/pdr-actual/03_security_validation.md

Deliverable Criteria:
- AuthZ/AuthN prototypes; logs of accepted/rejected flows
- Threat model coverage evidence
- Attack simulation results; remediations listed
- Secure deployment checklist validated

Success Criteria: Security design validated through working prototypes and attack testing.

3a. Component and Interface Gate Review (Project Manager)

Role: Project Manager
Inputs: 01_component_prototyping.md, 02_interface_implementation.md, 02b_api_contract_validation.md, 03_security_validation.md

Decision: PROCEED | REMEDIATE | CONDITIONAL | HALT

Create: evidence/pdr-actual/03a_component_interface_gate_review.md

Phase 2: System Integration and Performance Validation

4. Integration Planning and Validation (IV&V)

Role: IV&V
Task: Validate the integration approach via a working integration framework.

Execute exactly:
1) Create dependency‑ordered integration sequence
2) Build integration test harness/automation
3) Implement monitoring/validation hooks
4) Test sequence with prototypes; validate rollback/retry

Create: evidence/pdr-actual/04_integration_planning.md

Deliverable Criteria:
- Validated integration order
- Working integration test harness
- Monitoring hooks and health checks
- Recovery procedures validated

Success Criteria: Integration plan validated via working framework and prototype testing.

5. Performance Budget Validation (IV&V)

Role: IV&V
Task: Validate performance budgets through measurement and analysis.

Execute exactly:
1) Establish baseline measurements using prototypes
2) Create performance tests/benchmarks
3) Measure under realistic load
4) Validate budgets and scaling characteristics
5) Verify performance monitoring/alerts

Create: evidence/pdr-actual/05_performance_validation.md

Deliverable Criteria:
- Baseline data and benchmark definitions
- Load results and scaling analysis
- Monitoring/alert validation

Success Criteria: Performance budgets validated with realistic load tests; budgets deemed achievable.

5b. Performance Budget Sign‑Off (Project Manager)

Role: Project Manager
Input: 05_performance_validation.md
Task: Freeze PDR Budget Table (targets per API/path; CPU/RSS envelopes).

Create: evidence/pdr-actual/05b_performance_budget_signoff.md
Exit Criteria: Budget table approved; deviations carry waivers with owner/date.

6. Build and Deployment Pipeline Validation (Developer)

Role: Developer
Task: Implement and validate build, test, and deployment pipeline.

Execute exactly:
1) Create automated builds for all components
2) Configure CI with quality gates (lint/type, unit/integration)
3) Automate deployment/config management; verify rollback
4) Run pipeline end‑to‑end using prototypes

Create: evidence/pdr-actual/06_build_deployment_validation.md

Deliverable Criteria:
- Automated builds and CI runs (green)
- Deployment automation evidence
- Rollback validation steps/logs

Success Criteria: End‑to‑end pipeline validated; reproducible builds with checksums.

6b. Observability & Ops Readiness (Lite)

Role: Developer (with IV&V review)
Task: Establish minimal runbook and SLOs for PDR scope.

Create:
- ops/runbook.md (start/stop, health, logs, common errors, recovery)
- ops/slo.md (two SLOs: API p95 latency, recording success rate; basic alerts)

Exit Criteria: Documents present and referenced in 06_build_deployment_validation.md.

6a. System Integration Gate Review (Project Manager)

Role: Project Manager
Inputs: 04_integration_planning.md, 05_performance_validation.md, 05b_performance_budget_signoff.md, 06_build_deployment_validation.md, ops/runbook.md, ops/slo.md

Decision: PROCEED | REMEDIATE | CONDITIONAL | HALT

Create: evidence/pdr-actual/06a_system_integration_gate_review.md

Phase 3: Implementation Planning and PDR Decision

7. Implementation Strategy Validation (IV&V)

Role: IV&V
Task: Validate implementation strategy and development approach.

Create: evidence/pdr-actual/07_implementation_strategy.md
Deliverable Criteria: capability assessment, realistic timeline, QA strategy, risk mitigation, change control.
Success Criteria: Strategy executable with defined resources and controls.

8. PDR Technical Assessment (IV&V)

Role: IV&V
Task: Compile PDR assessment across design, components, interfaces, integration, performance, pipeline, and implementation strategy.

Create: evidence/pdr-actual/08_pdr_technical_assessment.md
Outcome: Recommendation = PROCEED | CONDITIONAL | DENY for full implementation.

9. PDR Authorization Decision (Project Manager)

Role: Project Manager
Task: Make authorization decision for full implementation.

Create: evidence/pdr-actual/09_pdr_authorization_decision.md
Decision: AUTHORIZE | CONDITIONAL | DENY with rationale, conditions, and risk acceptance.

Evidence Management

Document Template

# Document Title
**Version:** 1.0
**Date:** YYYY-MM-DD
**Role:** [Developer/IV&V/Project Manager]
**PDR Phase:** [Phase Number]
**Status:** [Draft/Review/Final]

## Purpose

## Implementation Results

## Validation Evidence

## Conclusion

Folders & Naming: evidence/pdr-actual/##_<descriptive>.md (00–09, with 0d, 0e, 2b, 5b, 6b as additions)

Evidence Integrity: Include command outputs and checksums. Preserve logs/artifacts under evidence/pdr-actual/artifacts/.

Gating: Subsequent phases cannot start without the specified tags, sign‑offs, and gate documents.

