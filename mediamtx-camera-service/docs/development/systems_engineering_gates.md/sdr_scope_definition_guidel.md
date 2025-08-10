SDR (System Design Review) – Scope Definition and Execution Guide

Purpose: Establish a disciplined, evidence-based process to resolve foundational issues prior to PDR/CDR while avoiding scope expansion and focusing on MVP deliverables.

SDR Objective

Validate that requirements are complete/testable and that the proposed architecture/technology stack is feasible through lightweight, working proofs before detailed design begins.

Non‑Goals and Scope Guardrails

No new features or scope expansion

No refactors beyond what is necessary for prototype validation

No additional hardware models beyond the MVP target list

Global SDR Acceptance Thresholds

Requirements: 100% traceable, testable, and validated
Architecture: Feasible via working proof‑of‑concept (PoC)
Interfaces: Defined and exercised via runnable prototypes
Technology: Critical selections proven via focused spikes
Risk: All High risks have concrete mitigation plans with evidence
Evidence: All claims backed by working demos and captured outputs
Timebox: 5–10 working days; 2 iterations max

Phase 0: Requirements Baseline

0. Requirements Validation and Baseline (IV&V)

Your role: IV&V
Ground rules: docs/development/project-ground-rules.md
Role reference: docs/development/roles-responsibilities.md

Task: Validate ALL requirements are complete, testable, unambiguous.

Execute exactly:
1) Inventory all requirements from source specs
2) Ensure each requirement has measurable acceptance criteria
3) Assess testability, measurability, and priority (criticality)
4) Identify dependencies and conflicts
5) Cross‑check requirements vs architecture vs API docs for consistency
6) Record explicit discrepancies with source references

VALIDATION LOOP:
- Do not stop on first issue; enumerate all gaps/conflicts
- Mark untestable items and missing criteria precisely

Create: evidence/sdr-actual/00_requirements_validation.md

DELIVERABLE CRITERIA:
- Complete catalog with IDs/descriptions
- Testability assessment per requirement
- Conflict and dependency analysis
- Cross‑document consistency audit
- Priority matrix (customer‑critical, system‑critical, etc.)
- Issue register with exact citations

Success Criteria: All requirements validated and discrepancies documented.

0a. Ground Truth Document Remediation (Project Manager)

Your role: Project Manager
Input: evidence/sdr-actual/00_requirements_validation.md

Task: Edit ground‑truth docs to fix ONLY the discrepancies identified by IV&V.

Scope control constraints:
- No new features or architectural scope
- Only clarify, add acceptance criteria, and align cross‑docs

Execute exactly:
1) Apply minimal edits to docs/requirements to add missing criteria
2) Align docs/architecture to requirements (consistency only)
3) Align docs/api to architecture interfaces
4) Validate each edit traces to an IV&V finding

Create: evidence/sdr-actual/00a_ground_truth_remediation.md

DELIVERABLE CRITERIA:
- Mapping table: finding → edit (with links/diffs)
- Before/after snippets for each change
- Cross‑doc consistency check report
- Scope‑creep check (no new scope)

Success Criteria: All IV&V findings resolved via targeted document edits; no scope creep.

0b. Requirements Baseline Gate Review (Project Manager)

Your role: Project Manager
Inputs: 00_requirements_validation.md, 00a_ground_truth_remediation.md

Gate review:
- Verify all findings are resolved
- Confirm cross‑doc consistency and clarity
- Decide readiness to proceed

Decision options: PROCEED | REMEDIATE | HALT

Create: evidence/sdr-actual/00b_requirements_baseline_gate_review.md

DELIVERABLE CRITERIA:
- Decision with rationale and evidence references
- Scope control confirmation
- Next steps authorization

Success Criteria: Requirements baseline gate review complete with authorized next steps.

0c. Assumptions & Non‑Goals Baseline (Project Manager)

Your role: Project Manager
Task: Freeze assumptions and non‑goals to prevent scope creep during SDR.

Create: evidence/sdr-actual/00c_assumptions_and_non_goals.md
Contents: explicit assumptions (with expiry/owner), non‑goals list, constraints (e.g., OS/camera list), and change‑control rule: PM waiver required for any deviation.

Phase 1: Architecture Feasibility

1. High‑Level Architecture Design (Developer)

Your role: Developer
Task: Design component architecture and prove feasibility via PoC.

Execute exactly:
1) Map components to validated requirements (ID‑level)
2) Define key interfaces and data flows
3) Select core technologies (justify briefly)
4) Implement a runnable PoC skeleton (service starts; minimal happy path works)
5) Capture outputs (logs, sample media) as evidence

Create: evidence/sdr-actual/01_architecture_design.md

DELIVERABLE CRITERIA:
- Diagram(s), interface/data‑flow descriptions
- Component↔requirement allocation table (high‑level)
- PoC run log + artifact samples proving feasibility

Success Criteria: Architecture designed; PoC demonstrates feasibility on MVP path.

2. Interface Definition & Validation (Developer)

Your role: Developer
Task: Define and exercise interfaces via executable prototypes.

Execute exactly:
1) Define external API/protocols (request/response schemas; error codes)
2) Define critical internal interfaces
3) Provide tiny harnesses (e.g., JSON‑RPC sample client) and run them
4) Log actual request/response exchanges; validate error handling

Create: evidence/sdr-actual/02_interface_validation.md

DELIVERABLE CRITERIA:
- Interface specs (schemas/examples)
- Prototype transcripts (real exchanges)
- Error‑path evidence (negative tests)

Success Criteria: Interfaces defined and validated through working prototypes.

2a. Minimal Security & Time Integrity Baseline (Developer)

Your role: Developer
Task: Establish baseline safeguards relevant to camera evidence.

Execute exactly:
1) Time source policy: document clock source and timestamping approach
2) Logging policy: avoid PII; include request IDs and monotonic timestamps
3) Auth baseline: ensure token validation path exists and rejects invalid/expired

Create: evidence/sdr-actual/02a_security_time_baseline.md

Success Criteria: Time, logging, and minimal auth behaviors demonstrated in prototype logs.

3. Technology Spike Validation (Developer)

Your role: Developer
Task: Validate critical technology selections via focused spikes.

Execute exactly:
1) Identify top 2–3 high‑risk tech choices (e.g., MediaMTX integration, WebSocket handling)
2) Implement minimal spikes proving integration/perf viability
3) Capture load sanity (e.g., one recording for 60–120 s; CPU/RSS snapshots)

Create: evidence/sdr-actual/03_technology_validation.md

DELIVERABLE CRITERIA:
- Spike code refs and commands used
- Screenshots/logs showing behavior under light load
- Brief risk notes per tech

Success Criteria: Critical technologies validated with working spikes.

3a. Architecture Feasibility Gate Review (Project Manager)

Your role: Project Manager
Inputs: 01_architecture_design.md, 02_interface_validation.md, 02a_security_time_baseline.md, 03_technology_validation.md

Gate review: evaluate PoC evidence, interface validation, minimal security/time integrity, and spikes.
Decision: PROCEED | REMEDIATE | CONDITIONAL | HALT

Create: evidence/sdr-actual/03a_architecture_feasibility_gate_review.md

Phase 2: Risk & Integration Assessment

4. Risk Assessment & Mitigation Planning (IV&V)

Your role: IV&V
Task: Identify risks; propose evidence‑based mitigations and monitors.

Create: evidence/sdr-actual/04_risk_assessment.md
Contents:
- Risk register with probability/impact
- Mitigation plans with success criteria
- Early‑warning indicators/telemetry
- Validation evidence for critical mitigations (tiny demos acceptable)

5. Integration Strategy Validation (IV&V)

Your role: IV&V
Task: Define and validate the integration approach through simulation/mocks.

Create: evidence/sdr-actual/05_integration_strategy.md
Contents:
- Dependency‑ordered integration sequence
- Integration test harness outline and a working mock sim
- Failure‑mode/recovery checks (e.g., camera flap, restart)

5a. Risk & Integration Gate Review (Project Manager)

Your role: Project Manager
Inputs: 04_risk_assessment.md, 05_integration_strategy.md
Decision: PROCEED | REMEDIATE | CONDITIONAL | HALT

Create: evidence/sdr-actual/05a_risk_integration_gate_review.md

Phase 3: SDR Decision

6. SDR Technical Assessment (IV&V)

Your role: IV&V
Task: Compile SDR assessment across requirements, architecture, interfaces, technology, risks, and integration.

Create: evidence/sdr-actual/06_sdr_technical_assessment.md
Outcome: Recommendation = PROCEED/CONDITIONAL/DENY for detailed design (PDR entry)

7. SDR Authorization Decision (Project Manager)

Your role: Project Manager
Task: Make SDR authorization decision and conditions for PDR entry.

Create: evidence/sdr-actual/07_sdr_authorization_decision.md

Evidence Management & Structure

Document Template:

# Document Title
**Version:** 1.0
**Date:** YYYY-MM-DD
**Role:** [Developer/IV&V/Project Manager]
**SDR Phase:** [Phase Number]
**Status:** [Draft/Review/Final]

## Purpose

## Execution Results

## Validation Evidence

## Conclusion

Folders & Naming: evidence/sdr-actual/##_<descriptive>.md (00–07, with 0c, 2a as additions)

Evidence Integrity: Include command outputs and hashes where relevant. Keep all PoC logs and media under evidence/sdr-actual/artifacts/.

Change Control: Any deviation from assumptions/non‑goals requires a one‑line PM waiver appended to 00c_assumptions_and_non_goals.md.